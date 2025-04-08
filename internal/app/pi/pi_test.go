package pi

import (
	mock "github.com/qonto/prometheus-rds-exporter/internal/app/pi/mock"
	"github.com/qonto/prometheus-rds-exporter/internal/app/rds"
	"github.com/stretchr/testify/require"
	"log/slog"

	"context"
	"testing"
)

func TestPerformanceInsightsFetcher_GetDBInstanceMetrics(t *testing.T) {

	testInstances := map[string]rds.RdsInstanceMetrics{
		"test": {
			PerformanceInsightsEnabled: true,
			DbiResourceID:              "DB-123",
		},
		"test2": {
			PerformanceInsightsEnabled: false,
			DbiResourceID:              "DB-456",
		},
	}

	ctx := context.TODO()
	client := mock.PerformanceInsightsClient{}
	fetcher := NewFetcher(ctx, client, slog.Logger{})
	metrics, err := fetcher.GetDBInstanceMetrics(testInstances)
	require.NoError(t, err, "GetInstancesMetrics must succeed")

	require.Len(t, metrics.Instances, 1, "GetInstancesMetrics must return 1 host")
	require.Contains(t, metrics.Instances, "test", "GetInstancesMetrics must return metrics for test")
	require.NotContains(t, metrics.Instances, "test2", "GetInstancesMetrics must not return metrics for test2")

	require.Equal(t, metrics.Instances["test"].DbCacheBlksHit, float64(2), "GetInstancesMetrics must return correct DbiResourceID")
	require.Equal(t, metrics.Instances["test"].DbCacheBuffersAlloc, 0.0006, "GetInstancesMetrics must return correct DbiResourceID")
}
