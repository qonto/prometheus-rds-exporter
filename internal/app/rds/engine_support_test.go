package rds_test

import (
	"context"
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	aws_rds "github.com/aws/aws-sdk-go-v2/service/rds"
	aws_rds_types "github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/patrickmn/go-cache"
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

func TestFetchEngineVersionLifecyclesWithoutSupport(t *testing.T) {
	ctx := context.Background()
	logger := slog.Default()

	// Create mock client with engine version data
	client := mock.NewRDSClient()

	mockResponse := &aws_rds.DescribeDBMajorEngineVersionsOutput{
		DBMajorEngineVersions: []aws_rds_types.DBMajorEngineVersion{
			{
				MajorEngineVersion: aws.String("13.00"),
			},
			{
				MajorEngineVersion: aws.String("14.00"),
			},
			{
				MajorEngineVersion: aws.String("15.00"),
			},
			{
				MajorEngineVersion: aws.String("16.00"),
			},
		},
	}

	client.WithDescribeDBMajorEngineVersionsOutput(mockResponse)

	// Create engine support service
	service := rds.NewEngineSupportService(client, logger)

	// Test fetching engine version lifecycles
	lifecycles, err := service.FetchEngineVersionLifecycles(ctx, "sqlserver-ee")

	require.NoError(t, err, "FetchEngineVersionLifecycles should succeed")
	require.Len(t, lifecycles, 0, "Should return 0 engine versions")
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

func TestCalculateRemainingDaysEdgeCases(t *testing.T) {
	logger := slog.Default()
	client := mock.NewRDSClient()
	service := rds.NewEngineSupportService(client, logger)

	now := time.Now()

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
			name:     "exact same time returns 0",
			endDate:  &now,
			expected: int64Ptr(0),
		},
		{
			name:     "1 second in future returns 1 day",
			endDate:  timePtr(now.Add(1 * time.Second)),
			expected: int64Ptr(1),
		},
		{
			name:     "1 second in past returns 0 days",
			endDate:  timePtr(now.Add(-1 * time.Second)),
			expected: int64Ptr(0),
		},
		{
			name:     "exactly 24 hours future returns 1 day",
			endDate:  timePtr(now.Add(24 * time.Hour)),
			expected: int64Ptr(1),
		},
		{
			name:     "exactly 24 hours past returns -1 day",
			endDate:  timePtr(now.Add(-24 * time.Hour)),
			expected: int64Ptr(-1),
		},
		{
			name:     "25 hours future returns 2 days (ceil)",
			endDate:  timePtr(now.Add(25 * time.Hour)),
			expected: int64Ptr(2),
		},
		{
			name:     "23 hours future returns 1 day (ceil)",
			endDate:  timePtr(now.Add(23 * time.Hour)),
			expected: int64Ptr(1),
		},
		{
			name:     "far future date",
			endDate:  timePtr(time.Date(2030, 12, 31, 0, 0, 0, 0, time.UTC)),
			expected: nil, // We'll check it's positive
		},
		{
			name:     "far past date",
			endDate:  timePtr(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)),
			expected: nil, // We'll check it's negative
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.CalculateRemainingDays(tt.endDate)

			if tt.expected == nil && tt.name == "nil date returns nil" {
				assert.Nil(t, result)
			} else if tt.expected == nil && tt.name == "far future date" {
				require.NotNil(t, result)
				assert.Greater(t, *result, int64(1000)) // Should be many days in future
			} else if tt.expected == nil && tt.name == "far past date" {
				require.NotNil(t, result)
				assert.Less(t, *result, int64(-1000)) // Should be many days in past
			} else {
				require.NotNil(t, result)
				// Allow for small differences due to test execution time
				assert.InDelta(t, *tt.expected, *result, 1)
			}
		})
	}
}

func TestGetEngineSupportMetricsErrorHandling(t *testing.T) {
	ctx := context.Background()
	logger := slog.Default()

	t.Run("empty engine parameter returns error", func(t *testing.T) {
		client := mock.NewRDSClient()
		service := rds.NewEngineSupportService(client, logger)

		metrics, err := service.GetEngineSupportMetrics(ctx, "", "13")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "engine parameter cannot be empty")
		assert.Equal(t, rds.EngineSupportMetrics{}, metrics)
	})

	t.Run("empty version parameter returns error", func(t *testing.T) {
		client := mock.NewRDSClient()
		service := rds.NewEngineSupportService(client, logger)

		metrics, err := service.GetEngineSupportMetrics(ctx, "postgres", "")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "version parameter cannot be empty")
		assert.Equal(t, rds.EngineSupportMetrics{}, metrics)
	})

	t.Run("API error propagates correctly", func(t *testing.T) {
		client := mock.NewRDSClient()
		client.WithDescribeDBMajorEngineVersionsError(fmt.Errorf("AccessDenied: insufficient permissions"))
		service := rds.NewEngineSupportService(client, logger)

		metrics, err := service.GetEngineSupportMetrics(ctx, "postgres", "13")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to fetch engine lifecycle data")
		assert.Contains(t, err.Error(), "AccessDenied")
		assert.Equal(t, rds.EngineSupportMetrics{}, metrics)
	})
}

func TestFetchEngineVersionLifecyclesErrorHandling(t *testing.T) {
	ctx := context.Background()
	logger := slog.Default()

	t.Run("empty engine parameter returns error", func(t *testing.T) {
		client := mock.NewRDSClient()
		service := rds.NewEngineSupportService(client, logger)

		lifecycles, err := service.FetchEngineVersionLifecycles(ctx, "")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "engine parameter cannot be empty")
		assert.Nil(t, lifecycles)
	})

	t.Run("API access denied error", func(t *testing.T) {
		client := mock.NewRDSClient()
		client.WithDescribeDBMajorEngineVersionsError(fmt.Errorf("AccessDenied: User is not authorized to perform: rds:DescribeDBMajorEngineVersions"))
		service := rds.NewEngineSupportService(client, logger)

		lifecycles, err := service.FetchEngineVersionLifecycles(ctx, "postgres")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to describe DB major engine versions")
		assert.Contains(t, err.Error(), "AccessDenied")
		assert.Nil(t, lifecycles)
	})

	t.Run("API invalid parameter error", func(t *testing.T) {
		client := mock.NewRDSClient()
		client.WithDescribeDBMajorEngineVersionsError(fmt.Errorf("InvalidParameterValue: Invalid engine name"))
		service := rds.NewEngineSupportService(client, logger)

		lifecycles, err := service.FetchEngineVersionLifecycles(ctx, "invalid-engine")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to describe DB major engine versions")
		assert.Contains(t, err.Error(), "InvalidParameterValue")
		assert.Nil(t, lifecycles)
	})

	t.Run("API throttling error", func(t *testing.T) {
		client := mock.NewRDSClient()
		client.WithDescribeDBMajorEngineVersionsError(fmt.Errorf("RequestLimitExceeded: Request rate exceeded"))
		service := rds.NewEngineSupportService(client, logger)

		lifecycles, err := service.FetchEngineVersionLifecycles(ctx, "postgres")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to describe DB major engine versions")
		assert.Contains(t, err.Error(), "RequestLimitExceeded")
		assert.Nil(t, lifecycles)
	})

	t.Run("nil API response handled gracefully", func(t *testing.T) {
		client := mock.NewRDSClient()
		client.WithDescribeDBMajorEngineVersionsOutput(nil)
		service := rds.NewEngineSupportService(client, logger)

		lifecycles, err := service.FetchEngineVersionLifecycles(ctx, "postgres")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "received nil response from AWS API")
		assert.Nil(t, lifecycles)
	})
}

func TestAPIResponseParsing(t *testing.T) {
	ctx := context.Background()
	logger := slog.Default()

	t.Run("empty API response returns empty lifecycles", func(t *testing.T) {
		client := mock.NewRDSClient()
		client.WithDescribeDBMajorEngineVersionsOutput(&aws_rds.DescribeDBMajorEngineVersionsOutput{
			DBMajorEngineVersions: []aws_rds_types.DBMajorEngineVersion{},
		})
		service := rds.NewEngineSupportService(client, logger)

		lifecycles, err := service.FetchEngineVersionLifecycles(ctx, "postgres")

		require.NoError(t, err)
		assert.Empty(t, lifecycles)
	})

	t.Run("version with nil MajorEngineVersion is skipped", func(t *testing.T) {
		client := mock.NewRDSClient()
		client.WithDescribeDBMajorEngineVersionsOutput(&aws_rds.DescribeDBMajorEngineVersionsOutput{
			DBMajorEngineVersions: []aws_rds_types.DBMajorEngineVersion{
				{
					MajorEngineVersion: nil, // This should be skipped
				},
				{
					MajorEngineVersion: aws.String("13"),
					SupportedEngineLifecycles: []aws_rds_types.SupportedEngineLifecycle{
						{
							LifecycleSupportName:    aws_rds_types.LifecycleSupportNameOpenSourceRdsStandardSupport,
							LifecycleSupportEndDate: timePtr(time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC)),
						},
					},
				},
			},
		})
		service := rds.NewEngineSupportService(client, logger)

		lifecycles, err := service.FetchEngineVersionLifecycles(ctx, "postgres")

		require.NoError(t, err)
		require.Len(t, lifecycles, 1)
		assert.Equal(t, "13", lifecycles[0].MajorEngineVersion)
	})

	t.Run("version with nil SupportedEngineLifecycles is skipped", func(t *testing.T) {
		client := mock.NewRDSClient()
		client.WithDescribeDBMajorEngineVersionsOutput(&aws_rds.DescribeDBMajorEngineVersionsOutput{
			DBMajorEngineVersions: []aws_rds_types.DBMajorEngineVersion{
				{
					MajorEngineVersion:        aws.String("12"),
					SupportedEngineLifecycles: nil, // This should be skipped
				},
				{
					MajorEngineVersion: aws.String("13"),
					SupportedEngineLifecycles: []aws_rds_types.SupportedEngineLifecycle{
						{
							LifecycleSupportName:    aws_rds_types.LifecycleSupportNameOpenSourceRdsStandardSupport,
							LifecycleSupportEndDate: timePtr(time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC)),
						},
					},
				},
			},
		})
		service := rds.NewEngineSupportService(client, logger)

		lifecycles, err := service.FetchEngineVersionLifecycles(ctx, "postgres")

		require.NoError(t, err)
		require.Len(t, lifecycles, 1)
		assert.Equal(t, "13", lifecycles[0].MajorEngineVersion)
	})

	t.Run("lifecycle with nil LifecycleSupportEndDate is skipped", func(t *testing.T) {
		client := mock.NewRDSClient()
		client.WithDescribeDBMajorEngineVersionsOutput(&aws_rds.DescribeDBMajorEngineVersionsOutput{
			DBMajorEngineVersions: []aws_rds_types.DBMajorEngineVersion{
				{
					MajorEngineVersion: aws.String("13"),
					SupportedEngineLifecycles: []aws_rds_types.SupportedEngineLifecycle{
						{
							LifecycleSupportName:    aws_rds_types.LifecycleSupportNameOpenSourceRdsStandardSupport,
							LifecycleSupportEndDate: nil, // This should be skipped
						},
						{
							LifecycleSupportName:    aws_rds_types.LifecycleSupportNameOpenSourceRdsExtendedSupport,
							LifecycleSupportEndDate: timePtr(time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)),
						},
					},
				},
			},
		})
		service := rds.NewEngineSupportService(client, logger)

		lifecycles, err := service.FetchEngineVersionLifecycles(ctx, "postgres")

		require.NoError(t, err)
		require.Len(t, lifecycles, 1)
		assert.Equal(t, "13", lifecycles[0].MajorEngineVersion)
		assert.Nil(t, lifecycles[0].StandardSupportEndDate)
		assert.NotNil(t, lifecycles[0].ExtendedSupportEndDate)
	})

	t.Run("unknown support type is ignored", func(t *testing.T) {
		client := mock.NewRDSClient()
		client.WithDescribeDBMajorEngineVersionsOutput(&aws_rds.DescribeDBMajorEngineVersionsOutput{
			DBMajorEngineVersions: []aws_rds_types.DBMajorEngineVersion{
				{
					MajorEngineVersion: aws.String("13"),
					SupportedEngineLifecycles: []aws_rds_types.SupportedEngineLifecycle{
						{
							LifecycleSupportName:    "unknown-support-type", // This should be ignored
							LifecycleSupportEndDate: timePtr(time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC)),
						},
						{
							LifecycleSupportName:    aws_rds_types.LifecycleSupportNameOpenSourceRdsStandardSupport,
							LifecycleSupportEndDate: timePtr(time.Date(2025, 6, 30, 0, 0, 0, 0, time.UTC)),
						},
					},
				},
			},
		})
		service := rds.NewEngineSupportService(client, logger)

		lifecycles, err := service.FetchEngineVersionLifecycles(ctx, "postgres")

		require.NoError(t, err)
		require.Len(t, lifecycles, 1)
		assert.Equal(t, "13", lifecycles[0].MajorEngineVersion)
		assert.NotNil(t, lifecycles[0].StandardSupportEndDate)
		assert.Nil(t, lifecycles[0].ExtendedSupportEndDate)
	})

	t.Run("version with no valid support dates is skipped", func(t *testing.T) {
		client := mock.NewRDSClient()
		client.WithDescribeDBMajorEngineVersionsOutput(&aws_rds.DescribeDBMajorEngineVersionsOutput{
			DBMajorEngineVersions: []aws_rds_types.DBMajorEngineVersion{
				{
					MajorEngineVersion: aws.String("12"),
					SupportedEngineLifecycles: []aws_rds_types.SupportedEngineLifecycle{
						{
							LifecycleSupportName:    "unknown-support-type",
							LifecycleSupportEndDate: timePtr(time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC)),
						},
					},
				},
				{
					MajorEngineVersion: aws.String("13"),
					SupportedEngineLifecycles: []aws_rds_types.SupportedEngineLifecycle{
						{
							LifecycleSupportName:    aws_rds_types.LifecycleSupportNameOpenSourceRdsStandardSupport,
							LifecycleSupportEndDate: timePtr(time.Date(2025, 6, 30, 0, 0, 0, 0, time.UTC)),
						},
					},
				},
			},
		})
		service := rds.NewEngineSupportService(client, logger)

		lifecycles, err := service.FetchEngineVersionLifecycles(ctx, "postgres")

		require.NoError(t, err)
		require.Len(t, lifecycles, 1) // Only version 13 should be included
		assert.Equal(t, "13", lifecycles[0].MajorEngineVersion)
	})
}

func TestCacheCorruptionHandling(t *testing.T) {
	ctx := context.Background()
	logger := slog.Default()

	t.Run("corrupted cache data triggers API call", func(t *testing.T) {
		client := mock.NewRDSClient()

		// Set up valid API response
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

		service := rds.NewEngineSupportService(client, logger)

		// Manually corrupt the cache by inserting invalid data
		cacheKey := service.GenerateCacheKey("postgres")
		service.GetCache().Set(cacheKey, "invalid-data-type", cache.DefaultExpiration)

		// This should detect corruption, remove it, and make API call
		lifecycles, err := service.FetchEngineVersionLifecycles(ctx, "postgres")

		require.NoError(t, err)
		require.Len(t, lifecycles, 1)
		assert.Equal(t, "13", lifecycles[0].MajorEngineVersion)

		// Verify API was called (corruption triggered fallback)
		assert.Equal(t, 1, client.GetDescribeDBMajorEngineVersionsCallCount())
	})
}

func TestCacheHitMissScenarios(t *testing.T) {
	ctx := context.Background()
	logger := slog.Default()

	t.Run("cache miss then cache hit", func(t *testing.T) {
		client := mock.NewRDSClient()

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

		service := rds.NewEngineSupportService(client, logger)

		// First call - cache miss, should call API
		lifecycles1, err1 := service.FetchEngineVersionLifecycles(ctx, "postgres")
		require.NoError(t, err1)
		require.Len(t, lifecycles1, 1)
		assert.Equal(t, 1, client.GetDescribeDBMajorEngineVersionsCallCount())

		// Second call - cache hit, should not call API
		lifecycles2, err2 := service.FetchEngineVersionLifecycles(ctx, "postgres")
		require.NoError(t, err2)
		require.Len(t, lifecycles2, 1)
		assert.Equal(t, 1, client.GetDescribeDBMajorEngineVersionsCallCount()) // Still 1

		// Results should be identical
		assert.Equal(t, lifecycles1, lifecycles2)
	})

	t.Run("different engines have separate cache entries", func(t *testing.T) {
		client := mock.NewRDSClient()

		standardEndDate := time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC)
		mockResponse := &aws_rds.DescribeDBMajorEngineVersionsOutput{
			DBMajorEngineVersions: []aws_rds_types.DBMajorEngineVersion{
				{
					MajorEngineVersion: aws.String("8.0"),
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

		service := rds.NewEngineSupportService(client, logger)

		// Call for postgres - cache miss
		_, err1 := service.FetchEngineVersionLifecycles(ctx, "postgres")
		require.NoError(t, err1)
		assert.Equal(t, 1, client.GetDescribeDBMajorEngineVersionsCallCount())

		// Call for mysql - cache miss (different engine)
		_, err2 := service.FetchEngineVersionLifecycles(ctx, "mysql")
		require.NoError(t, err2)
		assert.Equal(t, 2, client.GetDescribeDBMajorEngineVersionsCallCount())

		// Call for postgres again - cache hit
		_, err3 := service.FetchEngineVersionLifecycles(ctx, "postgres")
		require.NoError(t, err3)
		assert.Equal(t, 2, client.GetDescribeDBMajorEngineVersionsCallCount()) // Still 2
	})
}

func TestGetEngineSupportMetricsIntegration(t *testing.T) {
	ctx := context.Background()
	logger := slog.Default()

	t.Run("version with no support dates returns empty metrics", func(t *testing.T) {
		client := mock.NewRDSClient()
		client.WithDescribeDBMajorEngineVersionsOutput(&aws_rds.DescribeDBMajorEngineVersionsOutput{
			DBMajorEngineVersions: []aws_rds_types.DBMajorEngineVersion{
				{
					MajorEngineVersion:        aws.String("13"),
					SupportedEngineLifecycles: []aws_rds_types.SupportedEngineLifecycle{}, // Empty
				},
			},
		})
		service := rds.NewEngineSupportService(client, logger)

		metrics, err := service.GetEngineSupportMetrics(ctx, "postgres", "13")

		require.NoError(t, err)
		assert.Nil(t, metrics.StandardSupportRemainingDays)
		assert.Nil(t, metrics.ExtendedSupportRemainingDays)
	})

	t.Run("version with past support dates returns negative values", func(t *testing.T) {
		client := mock.NewRDSClient()

		pastStandardDate := time.Date(2020, 12, 31, 0, 0, 0, 0, time.UTC)
		pastExtendedDate := time.Date(2021, 12, 31, 0, 0, 0, 0, time.UTC)

		client.WithDescribeDBMajorEngineVersionsOutput(&aws_rds.DescribeDBMajorEngineVersionsOutput{
			DBMajorEngineVersions: []aws_rds_types.DBMajorEngineVersion{
				{
					MajorEngineVersion: aws.String("11"),
					SupportedEngineLifecycles: []aws_rds_types.SupportedEngineLifecycle{
						{
							LifecycleSupportName:    aws_rds_types.LifecycleSupportNameOpenSourceRdsStandardSupport,
							LifecycleSupportEndDate: &pastStandardDate,
						},
						{
							LifecycleSupportName:    aws_rds_types.LifecycleSupportNameOpenSourceRdsExtendedSupport,
							LifecycleSupportEndDate: &pastExtendedDate,
						},
					},
				},
			},
		})
		service := rds.NewEngineSupportService(client, logger)

		metrics, err := service.GetEngineSupportMetrics(ctx, "postgres", "11")

		require.NoError(t, err)
		require.NotNil(t, metrics.StandardSupportRemainingDays)
		require.NotNil(t, metrics.ExtendedSupportRemainingDays)

		// Should be negative since dates are in the past
		assert.Less(t, *metrics.StandardSupportRemainingDays, int64(0))
		assert.Less(t, *metrics.ExtendedSupportRemainingDays, int64(0))
	})
}

// Helper functions for tests
func timePtr(t time.Time) *time.Time {
	return &t
}

func int64Ptr(i int64) *int64 {
	return &i
}
