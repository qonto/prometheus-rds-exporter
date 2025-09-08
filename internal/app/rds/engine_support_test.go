package rds_test

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	aws_rds "github.com/aws/aws-sdk-go-v2/service/rds"
	aws_rds_types "github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/qonto/prometheus-rds-exporter/internal/app/rds"
	mock "github.com/qonto/prometheus-rds-exporter/internal/app/rds/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFetchEngineVersionLifecycles(t *testing.T) {
	ctx := context.Background()
	logger := slog.Default()

	// Create mock client with engine version data
	client := mock.NewRDSClient()
	
	// Mock response with PostgreSQL engine versions
	standardEndDate := time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC)
	extendedEndDate := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	
	mockResponse := &aws_rds.DescribeDBMajorEngineVersionsOutput{
		DBMajorEngineVersions: []aws_rds_types.DBMajorEngineVersion{
			{
				MajorEngineVersion: aws.String("13"),
				SupportedEngineLifecycles: []aws_rds_types.SupportedEngineLifecycle{
					{
						LifecycleSupportName:    aws_rds_types.LifecycleSupportNameOpenSourceRdsStandardSupport,
						LifecycleSupportEndDate: &standardEndDate,
					},
					{
						LifecycleSupportName:    aws_rds_types.LifecycleSupportNameOpenSourceRdsExtendedSupport,
						LifecycleSupportEndDate: &extendedEndDate,
					},
				},
			},
			{
				MajorEngineVersion: aws.String("14"),
				SupportedEngineLifecycles: []aws_rds_types.SupportedEngineLifecycle{
					{
						LifecycleSupportName:    aws_rds_types.LifecycleSupportNameOpenSourceRdsStandardSupport,
						LifecycleSupportEndDate: &standardEndDate,
					},
				},
			},
		},
	}
	
	client.WithDescribeDBMajorEngineVersionsOutput(mockResponse)
	
	// Create engine support service
	service := rds.NewEngineSupportService(client, logger)
	
	// Test fetching engine version lifecycles
	lifecycles, err := service.FetchEngineVersionLifecycles(ctx, "postgres")
	
	require.NoError(t, err, "FetchEngineVersionLifecycles should succeed")
	require.Len(t, lifecycles, 2, "Should return 2 engine versions")
	
	// Verify first engine version (13)
	assert.Equal(t, "postgres", lifecycles[0].Engine)
	assert.Equal(t, "13", lifecycles[0].MajorEngineVersion)
	assert.NotNil(t, lifecycles[0].StandardSupportEndDate)
	assert.NotNil(t, lifecycles[0].ExtendedSupportEndDate)
	assert.Equal(t, standardEndDate, *lifecycles[0].StandardSupportEndDate)
	assert.Equal(t, extendedEndDate, *lifecycles[0].ExtendedSupportEndDate)
	
	// Verify second engine version (14) - only standard support
	assert.Equal(t, "postgres", lifecycles[1].Engine)
	assert.Equal(t, "14", lifecycles[1].MajorEngineVersion)
	assert.NotNil(t, lifecycles[1].StandardSupportEndDate)
	assert.Nil(t, lifecycles[1].ExtendedSupportEndDate)
	assert.Equal(t, standardEndDate, *lifecycles[1].StandardSupportEndDate)
}

func TestFetchEngineVersionLifecyclesCache(t *testing.T) {
	ctx := context.Background()
	logger := slog.Default()

	// Create mock client
	client := mock.NewRDSClient()
	
	// Mock response
	standardEndDate := time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC)
	mockResponse := &aws_rds.DescribeDBMajorEngineVersionsOutput{
		DBMajorEngineVersions: []aws_rds_types.DBMajorEngineVersion{
			{
				MajorEngineVersion: aws.String("13"),
				SupportedEngineLifecycles: []aws_rds_types.SupportedEngineLifecycle{
					{
						LifecycleSupportName:    aws_rds_types.LifecycleSupportNameOpenSourceRdsStandardSupport,
						LifecycleSupportEndDate: &standardEndDate,
					},
				},
			},
		},
	}
	
	client.WithDescribeDBMajorEngineVersionsOutput(mockResponse)
	
	// Create engine support service
	service := rds.NewEngineSupportService(client, logger)
	
	// First call should hit the API
	lifecycles1, err1 := service.FetchEngineVersionLifecycles(ctx, "postgres")
	require.NoError(t, err1)
	require.Len(t, lifecycles1, 1)
	
	// Second call should use cache (verify by checking call count)
	lifecycles2, err2 := service.FetchEngineVersionLifecycles(ctx, "postgres")
	require.NoError(t, err2)
	require.Len(t, lifecycles2, 1)
	
	// Results should be identical
	assert.Equal(t, lifecycles1, lifecycles2)
	
	// Verify only one API call was made (cache working)
	assert.Equal(t, 1, client.GetDescribeDBMajorEngineVersionsCallCount())
}

func TestGenerateCacheKey(t *testing.T) {
	logger := slog.Default()
	client := mock.NewRDSClient()
	service := rds.NewEngineSupportService(client, logger)
	
	key := service.GenerateCacheKey("postgres")
	assert.Equal(t, "engine_lifecycle_postgres", key)
	
	key2 := service.GenerateCacheKey("mysql")
	assert.Equal(t, "engine_lifecycle_mysql", key2)
}

func TestCalculateRemainingDays(t *testing.T) {
	logger := slog.Default()
	client := mock.NewRDSClient()
	service := rds.NewEngineSupportService(client, logger)
	
	tests := []struct {
		name     string
		endDate  *time.Time
		expected *int64
	}{
		{
			name:     "nil date returns nil",
			endDate:  nil,
			expected: nil,
		},
		{
			name:     "future date returns positive days",
			endDate:  timePtr(time.Now().Add(30 * 24 * time.Hour)),
			expected: int64Ptr(30),
		},
		{
			name:     "past date returns negative days",
			endDate:  timePtr(time.Now().Add(-10 * 24 * time.Hour)),
			expected: int64Ptr(-10),
		},
		{
			name:     "fractional future days rounds up",
			endDate:  timePtr(time.Now().Add(25*time.Hour + 30*time.Minute)),
			expected: int64Ptr(2), // 1.02 days rounds up to 2
		},
		{
			name:     "fractional past days rounds up (towards zero)",
			endDate:  timePtr(time.Now().Add(-25*time.Hour - 30*time.Minute)),
			expected: int64Ptr(-1), // -1.02 days rounds up to -1
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.CalculateRemainingDays(tt.endDate)
			
			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				require.NotNil(t, result)
				// Allow for small differences due to test execution time
				assert.InDelta(t, *tt.expected, *result, 1)
			}
		})
	}
}

func TestGetEngineSupportMetrics(t *testing.T) {
	ctx := context.Background()
	logger := slog.Default()
	
	// Create mock client
	client := mock.NewRDSClient()
	
	// Mock response with PostgreSQL engine versions
	standardEndDate := time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC)
	extendedEndDate := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	
	mockResponse := &aws_rds.DescribeDBMajorEngineVersionsOutput{
		DBMajorEngineVersions: []aws_rds_types.DBMajorEngineVersion{
			{
				MajorEngineVersion: aws.String("13"),
				SupportedEngineLifecycles: []aws_rds_types.SupportedEngineLifecycle{
					{
						LifecycleSupportName:    aws_rds_types.LifecycleSupportNameOpenSourceRdsStandardSupport,
						LifecycleSupportEndDate: &standardEndDate,
					},
					{
						LifecycleSupportName:    aws_rds_types.LifecycleSupportNameOpenSourceRdsExtendedSupport,
						LifecycleSupportEndDate: &extendedEndDate,
					},
				},
			},
			{
				MajorEngineVersion: aws.String("14"),
				SupportedEngineLifecycles: []aws_rds_types.SupportedEngineLifecycle{
					{
						LifecycleSupportName:    aws_rds_types.LifecycleSupportNameOpenSourceRdsStandardSupport,
						LifecycleSupportEndDate: &standardEndDate,
					},
				},
			},
		},
	}
	
	client.WithDescribeDBMajorEngineVersionsOutput(mockResponse)
	
	// Create engine support service
	service := rds.NewEngineSupportService(client, logger)
	
	t.Run("returns metrics for existing version with both support types", func(t *testing.T) {
		metrics, err := service.GetEngineSupportMetrics(ctx, "postgres", "13")
		
		require.NoError(t, err)
		assert.NotNil(t, metrics.StandardSupportRemainingDays)
		assert.NotNil(t, metrics.ExtendedSupportRemainingDays)
		
		// Should be positive since dates are in the future
		assert.Greater(t, *metrics.StandardSupportRemainingDays, int64(0))
		assert.Greater(t, *metrics.ExtendedSupportRemainingDays, int64(0))
	})
	
	t.Run("returns metrics for existing version with only standard support", func(t *testing.T) {
		metrics, err := service.GetEngineSupportMetrics(ctx, "postgres", "14")
		
		require.NoError(t, err)
		assert.NotNil(t, metrics.StandardSupportRemainingDays)
		assert.Nil(t, metrics.ExtendedSupportRemainingDays)
		
		// Should be positive since date is in the future
		assert.Greater(t, *metrics.StandardSupportRemainingDays, int64(0))
	})
	
	t.Run("returns empty metrics for non-existing version", func(t *testing.T) {
		metrics, err := service.GetEngineSupportMetrics(ctx, "postgres", "99")
		
		require.NoError(t, err)
		assert.Nil(t, metrics.StandardSupportRemainingDays)
		assert.Nil(t, metrics.ExtendedSupportRemainingDays)
	})
}

// Helper functions for tests
func timePtr(t time.Time) *time.Time {
	return &t
}

func int64Ptr(i int64) *int64 {
	return &i
}