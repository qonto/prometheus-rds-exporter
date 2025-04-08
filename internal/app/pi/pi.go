package pi

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	aws_performanceinsights "github.com/aws/aws-sdk-go-v2/service/pi"
	"github.com/aws/aws-sdk-go-v2/service/pi/types"
	"github.com/qonto/prometheus-rds-exporter/internal/app/rds"
	"log/slog"
	"time"
)

// There are only few values which I can use
const periodInSeconds = 60
const maxMetricPerQuery = 15

type DbMetrics struct {
	Instances map[string]PerformanceInsightsMetrics
}

type PerformanceInsightsStatistics struct {
	UsageAPICall float64
}

type PerformanceInsightsClient interface {
	GetResourceMetrics(ctx context.Context, params *aws_performanceinsights.GetResourceMetricsInput, optFns ...func(*aws_performanceinsights.Options)) (*aws_performanceinsights.GetResourceMetricsOutput, error)
}

type PerformanceInsightsFetcher struct {
	ctx        context.Context
	logger     *slog.Logger
	client     PerformanceInsightsClient
	statistics PerformanceInsightsStatistics
}

func NewFetcher(ctx context.Context, client PerformanceInsightsClient, logger slog.Logger) *PerformanceInsightsFetcher {
	return &PerformanceInsightsFetcher{
		ctx:    ctx,
		client: client,
		logger: &logger,
	}
}

func (p *PerformanceInsightsFetcher) GetStatistics() PerformanceInsightsStatistics {
	return p.statistics
}

func (p *PerformanceInsightsFetcher) GetDBInstanceMetrics(instances map[string]rds.RdsInstanceMetrics) (DbMetrics, error) {
	output := &DbMetrics{
		Instances: make(map[string]PerformanceInsightsMetrics),
	}

	for instanceName, dbInstance := range instances {
		if dbInstance.PerformanceInsightsEnabled == false {
			continue
		}
		result, err := p.makeRDSQuery(dbInstance.DbiResourceID)
		if err != nil {
			p.logger.Error("Failed to get performance insights metrics ", slog.Any("err", err))
			return *output, err
		}
		output.Instances[instanceName] = result
	}
	return *output, nil
}

func (p *PerformanceInsightsFetcher) makeRDSQuery(instanceID string) (PerformanceInsightsMetrics, error) {
	// We have to divide query into multiple queries because we can only get 15 metrics per query
	var queries [][]types.MetricQuery
	var metrics []types.MetricKeyDataPoints
	var output PerformanceInsightsMetrics

	// Got only 1 minutes of data
	startTime := time.Now().Add(-time.Minute)
	endTime := time.Now()

	// Allocate memory for queries
	queries = make([][]types.MetricQuery, len(DbMetricNames)/maxMetricPerQuery+1)

	chunkIdx := 0
	for i := 0; i < len(DbMetricNames); i += maxMetricPerQuery {
		end := i + maxMetricPerQuery
		if end > len(DbMetricNames) {
			end = len(DbMetricNames)
		}
		for index, _ := range DbMetricNames[i:end] {
			query := types.MetricQuery{
				Metric: &DbMetricNames[index+i],
			}
			queries[chunkIdx] = append(queries[chunkIdx], query)
		}
		chunkIdx++
	}
	for i := 0; i < len(queries); i++ {
		params := &aws_performanceinsights.GetResourceMetricsInput{
			StartTime: &startTime,
			EndTime:   &endTime,
			// We're using 60 second as a period because we want to get the metrics by at least 15 seconds
			PeriodInSeconds: aws.Int32(periodInSeconds),
			Identifier:      &instanceID,
			ServiceType:     types.ServiceTypeRds,
			MetricQueries:   queries[i],
		}
		result, err := p.client.GetResourceMetrics(p.ctx, params)
		p.statistics.UsageAPICall++
		if err != nil {
			p.logger.Error("Failed to get performance insights metrics", slog.Any("err", err))
			return output, err
		}
		metrics = append(metrics, result.MetricList...)
	}
	output = fillMetricsData(metrics)
	return output, nil
}
