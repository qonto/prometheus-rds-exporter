package cloudwatch_test

import (
	"log/slog"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	aws_cloudwatch_types "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/qonto/prometheus-rds-exporter/internal/app/cloudwatch"
	cloudwatch_mock "github.com/qonto/prometheus-rds-exporter/internal/app/cloudwatch/mock"
	converter "github.com/qonto/prometheus-rds-exporter/internal/app/unit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetUsageMetrics(t *testing.T) {
	expected := cloudwatch.UsageMetrics{
		AllocatedStorage:    100,
		DBInstances:         42,
		ManualSnapshots:     10,
		ReservedDBInstances: 3,
	}

	client := cloudwatch_mock.CloudwatchClient{
		Metrics: []aws_cloudwatch_types.MetricDataResult{
			{
				Label:  aws.String("AllocatedStorage"),
				Values: []float64{expected.AllocatedStorage},
			},
			{
				Label:  aws.String("DBInstances"),
				Values: []float64{expected.DBInstances},
			},
			{
				Label:  aws.String("ManualSnapshots"),
				Values: []float64{expected.ManualSnapshots},
			},
			{
				Label:  aws.String("ReservedDBInstances"),
				Values: []float64{expected.ReservedDBInstances},
			},
		},
	}

	fetcher := cloudwatch.NewUsageFetcher(client, slog.Logger{})
	result, err := fetcher.GetUsageMetrics()

	require.NoError(t, err, "GetUsageMetrics must succeed")
	assert.Equal(t, converter.GigaBytesToBytes(expected.AllocatedStorage), result.AllocatedStorage, "Allocated storage mismatch")
	assert.Equal(t, expected.DBInstances, result.DBInstances, "DB instances count mismatch")
	assert.Equal(t, expected.ManualSnapshots, result.ManualSnapshots, "Manual snapshots mismatch")
	assert.Equal(t, expected.ReservedDBInstances, result.ReservedDBInstances, "Reserved DB instances mismatch")

	assert.Equal(t, float64(1), fetcher.GetStatistics().CloudWatchAPICall, "One call to Cloudwatch API")
}
