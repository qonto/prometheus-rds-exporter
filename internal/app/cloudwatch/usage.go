package cloudwatch

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	aws_cloudwatch "github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	aws_cloudwatch_types "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	converter "github.com/qonto/prometheus-rds-exporter/internal/app/unit"
)

type UsageMetrics struct {
	AllocatedStorage    float64
	DBInstances         float64
	ManualSnapshots     float64
	ReservedDBInstances float64
}

func (u *UsageMetrics) Update(field string, value float64) error {
	switch field {
	case "AllocatedStorage":
		u.AllocatedStorage = converter.GigaBytesToBytes(value)
	case "DBInstances":
		u.DBInstances = value
	case "ManualSnapshots":
		u.ManualSnapshots = value
	case "ReservedDBInstances":
		u.ReservedDBInstances = value
	default:
		return fmt.Errorf("can't process %s metrics: %w", field, errUnknownMetric)
	}

	return nil
}

func generateCloudWatchDataQueriesForUsage(service string, metricsName []string) map[string]CloudWatchMetricRequest {
	requests := make(map[string]CloudWatchMetricRequest)

	for i, metricName := range metricsName {
		id := aws.String(fmt.Sprintf("%s_%d", strings.ToLower(metricName), i))
		query := &aws_cloudwatch_types.MetricDataQuery{
			Id: id,
			MetricStat: &aws_cloudwatch_types.MetricStat{
				Metric: &aws_cloudwatch_types.Metric{
					Namespace:  aws.String("AWS/Usage"),
					MetricName: aws.String("ResourceCount"),
					Dimensions: []aws_cloudwatch_types.Dimension{
						{
							Name:  aws.String("Service"),
							Value: aws.String(service),
						},
						{
							Name:  aws.String("Type"),
							Value: aws.String("Resource"),
						},
						{
							Name:  aws.String("Resource"),
							Value: aws.String(metricName),
						},
						{
							Name:  aws.String("Class"),
							Value: aws.String("None"),
						},
					},
				},
				Stat:   aws.String("Average"), // Specify the statistic to retrieve
				Period: aws.Int32(Minute),     // Specify the period of the metric. Granularity - 1 minute
			},
		}

		requests[*id] = CloudWatchMetricRequest{
			Dbidentifier: "",
			MetricName:   metricName,
			Query:        *query,
		}
	}

	return requests
}

func generateCloudWatchQueriesForUsage() *aws_cloudwatch.GetMetricDataInput {
	metricsName := []string{
		"AllocatedStorage",
		"DBInstances",
		"ManualSnapshots",
	}

	cloudwatchDataQueries := []aws_cloudwatch_types.MetricDataQuery{}
	queries := generateCloudWatchDataQueriesForUsage("RDS", metricsName)

	for _, usageQuery := range queries {
		query := aws_cloudwatch_types.MetricDataQuery{
			Id: usageQuery.Query.Id,
			MetricStat: &aws_cloudwatch_types.MetricStat{
				Metric: usageQuery.Query.MetricStat.Metric,
				Stat:   aws.String("Average"),
				Period: aws.Int32(CloudwatchUsagePeriod * Minute),
			},
		}
		cloudwatchDataQueries = append(cloudwatchDataQueries, query)
	}

	return &aws_cloudwatch.GetMetricDataInput{
		StartTime:         aws.Time(time.Now().Add(-5 * time.Hour)),
		EndTime:           aws.Time(time.Now()),
		ScanBy:            "TimestampDescending",
		MetricDataQueries: cloudwatchDataQueries,
	}
}

func NewUsageFetcher(client CloudWatchClient, logger slog.Logger) *usageFetcher {
	return &usageFetcher{
		client: client,
		logger: &logger,
	}
}

type usageFetcher struct {
	client     CloudWatchClient
	statistics Statistics
	logger     *slog.Logger
}

func (u *usageFetcher) GetStatistics() Statistics {
	return u.statistics
}

// GetUsageMetrics returns RDS service usages metrics
func (u *usageFetcher) GetUsageMetrics() (UsageMetrics, error) {
	metrics := UsageMetrics{}

	query := generateCloudWatchQueriesForUsage()

	resp, err := u.client.GetMetricData(context.TODO(), query)
	u.statistics.CloudWatchAPICall++

	if err != nil {
		return metrics, fmt.Errorf("error calling GetMetricData: %w", err)
	}

	for _, m := range resp.MetricDataResults {
		if m.Values == nil {
			u.logger.Warn("cloudwatch value is empty", "metric", *m.Label)

			continue
		}

		if len(m.Values) > 0 {
			err = metrics.Update(*m.Label, m.Values[0])
			if err != nil {
				return metrics, fmt.Errorf("can't update internal values: %w", err)
			}
		}
	}

	return metrics, nil
}
