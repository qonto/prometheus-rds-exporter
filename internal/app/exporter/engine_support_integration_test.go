package exporter_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	aws_rds "github.com/aws/aws-sdk-go-v2/service/rds"
	aws_rds_types "github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/patrickmn/go-cache"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/qonto/prometheus-rds-exporter/internal/app/exporter"
	"github.com/qonto/prometheus-rds-exporter/internal/app/rds"
	"github.com/qonto/prometheus-rds-exporter/internal/infra/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	cloudwatch_mock "github.com/qonto/prometheus-rds-exporter/internal/app/cloudwatch/mock"
	ec2_mock "github.com/qonto/prometheus-rds-exporter/internal/app/ec2/mock"
	rds_mock "github.com/qonto/prometheus-rds-exporter/internal/app/rds/mock"
	servicequotas_mock "github.com/qonto/prometheus-rds-exporter/internal/app/servicequotas/mock"
)

// TestEngineSupport_EndToEndMetricCollection tests the complete flow from RDS instances
// to metric emission, including API calls, caching, and Prometheus metric generation
func TestEngineSupport_EndToEndMetricCollection(t *testing.T) {
	tests := []struct {
		name                    string
		instances               []aws_rds_types.DBInstance
		engineVersionsResponse  *aws_rds.DescribeDBMajorEngineVersionsOutput
		expectedStandardMetrics int
		expectedExtendedMetrics int
		expectedAPICallCount    int
		validateMetricValues    func(t *testing.T, standardMetrics, extendedMetrics []*dto.Metric)
	}{
		{
			name: "single PostgreSQL instance with both support types",
			instances: []aws_rds_types.DBInstance{
				createPostgreSQLInstance("test-postgres-1", "13.7"),
			},
			engineVersionsResponse: createEngineVersionsResponse("postgres", []engineVersionData{
				{
					version:            "13",
					standardEndDate:    time.Now().AddDate(0, 6, 0), // 6 months from now
					extendedEndDate:    time.Now().AddDate(1, 0, 0), // 1 year from now
					hasStandardSupport: true,
					hasExtendedSupport: true,
				},
			}),
			expectedStandardMetrics: 1,
			expectedExtendedMetrics: 1,
			expectedAPICallCount:    1,
			validateMetricValues: func(t *testing.T, standardMetrics, extendedMetrics []*dto.Metric) {
				// Standard support should be around 180 days
				standardValue := standardMetrics[0].GetGauge().GetValue()
				assert.Greater(t, standardValue, 170.0)
				assert.Less(t, standardValue, 190.0)

				// Extended support should be around 365 days
				extendedValue := extendedMetrics[0].GetGauge().GetValue()
				assert.Greater(t, extendedValue, 355.0)
				assert.Less(t, extendedValue, 375.0)
			},
		},
		{
			name: "multiple PostgreSQL instances with different versions",
			instances: []aws_rds_types.DBInstance{
				createPostgreSQLInstance("postgres-13", "13.7"),
				createPostgreSQLInstance("postgres-14", "14.9"),
				createPostgreSQLInstance("postgres-15", "15.3"),
			},
			engineVersionsResponse: createEngineVersionsResponse("postgres", []engineVersionData{
				{
					version:            "13",
					standardEndDate:    time.Now().AddDate(0, 3, 0), // 3 months from now
					extendedEndDate:    time.Now().AddDate(0, 9, 0), // 9 months from now
					hasStandardSupport: true,
					hasExtendedSupport: true,
				},
				{
					version:            "14",
					standardEndDate:    time.Now().AddDate(0, 12, 0), // 12 months from now
					hasStandardSupport: true,
					hasExtendedSupport: false,
				},
				{
					version:            "15",
					standardEndDate:    time.Now().AddDate(1, 6, 0), // 18 months from now
					extendedEndDate:    time.Now().AddDate(2, 0, 0), // 2 years from now
					hasStandardSupport: true,
					hasExtendedSupport: true,
				},
			}),
			expectedStandardMetrics: 3, // All three instances should have standard support metrics
			expectedExtendedMetrics: 2, // Only versions 13 and 15 have extended support
			expectedAPICallCount:    1, // Single API call for postgres engine
			validateMetricValues: func(t *testing.T, standardMetrics, extendedMetrics []*dto.Metric) {
				// Verify we have the expected number of metrics
				assert.Len(t, standardMetrics, 3)
				assert.Len(t, extendedMetrics, 2)

				// All standard support values should be positive (future dates)
				for _, metric := range standardMetrics {
					value := metric.GetGauge().GetValue()
					assert.Greater(t, value, 0.0, "Standard support should be positive for future dates")
				}

				// All extended support values should be positive (future dates)
				for _, metric := range extendedMetrics {
					value := metric.GetGauge().GetValue()
					assert.Greater(t, value, 0.0, "Extended support should be positive for future dates")
				}
			},
		},
		{
			name: "mixed engine types - only PostgreSQL should have metrics",
			instances: []aws_rds_types.DBInstance{
				createPostgreSQLInstance("postgres-instance", "14.9"),
				createMySQLInstance("mysql-instance", "8.0.35"),
				createMariaDBInstance("mariadb-instance", "10.6.14"),
			},
			engineVersionsResponse: createEngineVersionsResponse("postgres", []engineVersionData{
				{
					version:            "14",
					standardEndDate:    time.Now().AddDate(0, 8, 0),
					hasStandardSupport: true,
					hasExtendedSupport: false,
				},
			}),
			expectedStandardMetrics: 1, // Only PostgreSQL instance
			expectedExtendedMetrics: 0, // No extended support for this version
			expectedAPICallCount:    3, // Called once per unique engine type (postgres, mysql, mariadb)
			validateMetricValues: func(t *testing.T, standardMetrics, extendedMetrics []*dto.Metric) {
				// Verify only PostgreSQL instance has metrics
				assert.Len(t, standardMetrics, 1)
				assert.Len(t, extendedMetrics, 0)

				// Verify the metric is for the PostgreSQL instance
				labels := standardMetrics[0].GetLabel()
				var engine, dbidentifier string
				for _, label := range labels {
					switch label.GetName() {
					case "engine":
						engine = label.GetValue()
					case "dbidentifier":
						dbidentifier = label.GetValue()
					}
				}
				assert.Equal(t, "postgres", engine)
				assert.Equal(t, "postgres-instance", dbidentifier)
			},
		},
		{
			name: "PostgreSQL instances with past support dates (negative values)",
			instances: []aws_rds_types.DBInstance{
				createPostgreSQLInstance("old-postgres", "11.20"),
			},
			engineVersionsResponse: createEngineVersionsResponse("postgres", []engineVersionData{
				{
					version:            "11",
					standardEndDate:    time.Now().AddDate(0, 0, -60), // 60 days ago
					extendedEndDate:    time.Now().AddDate(0, 0, -30), // 30 days ago
					hasStandardSupport: true,
					hasExtendedSupport: true,
				},
			}),
			expectedStandardMetrics: 1,
			expectedExtendedMetrics: 1,
			expectedAPICallCount:    1,
			validateMetricValues: func(t *testing.T, standardMetrics, extendedMetrics []*dto.Metric) {
				// Standard support should be negative (around -60 days)
				standardValue := standardMetrics[0].GetGauge().GetValue()
				assert.Less(t, standardValue, 0.0)
				assert.Greater(t, standardValue, -65.0)
				assert.Less(t, standardValue, -55.0)

				// Extended support should be negative (around -30 days)
				extendedValue := extendedMetrics[0].GetGauge().GetValue()
				assert.Less(t, extendedValue, 0.0)
				assert.Greater(t, extendedValue, -35.0)
				assert.Less(t, extendedValue, -25.0)
			},
		},
		{
			name: "PostgreSQL instance with no lifecycle data",
			instances: []aws_rds_types.DBInstance{
				createPostgreSQLInstance("postgres-no-lifecycle", "16.0"),
			},
			engineVersionsResponse: &aws_rds.DescribeDBMajorEngineVersionsOutput{
				DBMajorEngineVersions: []aws_rds_types.DBMajorEngineVersion{
					{
						Engine:                    aws.String("postgres"),
						MajorEngineVersion:        aws.String("16"),
						SupportedEngineLifecycles: []aws_rds_types.SupportedEngineLifecycle{}, // No lifecycle data
					},
				},
			},
			expectedStandardMetrics: 0, // No metrics when no lifecycle data
			expectedExtendedMetrics: 0,
			expectedAPICallCount:    1,
			validateMetricValues: func(t *testing.T, standardMetrics, extendedMetrics []*dto.Metric) {
				// Should have no metrics when no lifecycle data is available
				assert.Empty(t, standardMetrics)
				assert.Empty(t, extendedMetrics)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			awsAccountID := "123456789012"
			awsRegion := "eu-west-3"

			logger, _ := logger.New(true, "text")
			rdsClient := rds_mock.NewRDSClient().
				WithDBInstances(tt.instances...).
				WithDescribeDBMajorEngineVersionsOutput(tt.engineVersionsResponse)

			ec2Client := ec2_mock.EC2Client{}
			cloudWatchClient := cloudwatch_mock.CloudwatchClient{}
			servicequotasClient := servicequotas_mock.ServiceQuotasClient{}

			configuration := exporter.Configuration{
				CollectInstanceMetrics: true,
				CollectInstanceTypes:   false,
				CollectInstanceTags:    false,
				CollectLogsSize:        false,
				CollectMaintenances:    false,
				CollectQuotas:          false,
				CollectUsages:          false,
				CollectEngineSupport:   true,
			}

			collector := exporter.NewCollector(*logger, configuration, awsAccountID, awsRegion, rdsClient, ec2Client, cloudWatchClient, servicequotasClient, nil)

			// Collect metrics
			registry := prometheus.NewRegistry()
			registry.MustRegister(collector)

			metricFamilies, err := registry.Gather()
			require.NoError(t, err)

			// Extract engine support metrics
			var standardSupportMetrics, extendedSupportMetrics []*dto.Metric
			for _, mf := range metricFamilies {
				switch mf.GetName() {
				case "rds_standard_support_engine_remaining_days":
					standardSupportMetrics = mf.GetMetric()
				case "rds_extended_support_engine_remaining_days":
					extendedSupportMetrics = mf.GetMetric()
				}
			}

			// Verify metric counts
			assert.Len(t, standardSupportMetrics, tt.expectedStandardMetrics, "Standard support metric count mismatch")
			assert.Len(t, extendedSupportMetrics, tt.expectedExtendedMetrics, "Extended support metric count mismatch")

			// Verify API call count
			assert.Equal(t, tt.expectedAPICallCount, rdsClient.GetDescribeDBMajorEngineVersionsCallCount(), "API call count mismatch")

			// Run custom validation
			if tt.validateMetricValues != nil {
				tt.validateMetricValues(t, standardSupportMetrics, extendedSupportMetrics)
			}

			// Verify all metrics have proper labels
			allMetrics := append(standardSupportMetrics, extendedSupportMetrics...)
			for _, metric := range allMetrics {
				verifyMetricLabels(t, metric, awsAccountID, awsRegion)
			}
		})
	}
}

// TestEngineSupport_CacheExpirationAndRefresh tests cache behavior with time manipulation
func TestEngineSupport_CacheExpirationAndRefresh(t *testing.T) {
	awsAccountID := "123456789012"
	awsRegion := "eu-west-3"

	// Create PostgreSQL instance
	postgresInstance := createPostgreSQLInstance("test-postgres", "14.9")

	// Create mock engine version data
	engineVersionsResponse := createEngineVersionsResponse("postgres", []engineVersionData{
		{
			version:            "14",
			standardEndDate:    time.Now().AddDate(0, 6, 0),
			hasStandardSupport: true,
			hasExtendedSupport: false,
		},
	})

	logger, _ := logger.New(true, "text")
	rdsClient := rds_mock.NewRDSClient().
		WithDBInstances(postgresInstance).
		WithDescribeDBMajorEngineVersionsOutput(engineVersionsResponse)

	ec2Client := ec2_mock.EC2Client{}
	cloudWatchClient := cloudwatch_mock.CloudwatchClient{}
	servicequotasClient := servicequotas_mock.ServiceQuotasClient{}

	configuration := exporter.Configuration{
		CollectInstanceMetrics: true,
		CollectInstanceTypes:   false,
		CollectInstanceTags:    false,
		CollectLogsSize:        false,
		CollectMaintenances:    false,
		CollectQuotas:          false,
		CollectUsages:          false,
		CollectEngineSupport:   true,
	}

	collector := exporter.NewCollector(*logger, configuration, awsAccountID, awsRegion, rdsClient, ec2Client, cloudWatchClient, servicequotasClient, nil)

	// First collection - should call API
	registry1 := prometheus.NewRegistry()
	registry1.MustRegister(collector)
	_, err := registry1.Gather()
	require.NoError(t, err)

	firstCallCount := rdsClient.GetDescribeDBMajorEngineVersionsCallCount()
	assert.Equal(t, 1, firstCallCount, "First collection should call API")

	// Second collection immediately - should use cache
	registry2 := prometheus.NewRegistry()
	registry2.MustRegister(collector)
	_, err = registry2.Gather()
	require.NoError(t, err)

	secondCallCount := rdsClient.GetDescribeDBMajorEngineVersionsCallCount()
	assert.Equal(t, 1, secondCallCount, "Second collection should use cache")

	// Test cache expiration by creating a new service with short TTL
	// This simulates cache expiration
	ctx := context.Background()
	shortTTLService := rds.NewEngineSupportServiceWithCache(rdsClient, *logger, cache.New(1*time.Millisecond, 1*time.Millisecond))

	// Wait for cache to expire
	time.Sleep(10 * time.Millisecond)

	// This should trigger a new API call due to expired cache
	_, err = shortTTLService.GetEngineSupportMetrics(ctx, "postgres", "14")
	require.NoError(t, err)

	// Verify API was called again after cache expiration
	finalCallCount := rdsClient.GetDescribeDBMajorEngineVersionsCallCount()
	assert.Equal(t, 2, finalCallCount, "Should call API again after cache expiration")
}

// TestEngineSupport_APIErrorHandling tests various API error scenarios
func TestEngineSupport_APIErrorHandling(t *testing.T) {
	tests := []struct {
		name                string
		apiError            error
		expectMetrics       bool
		expectErrorIncrease bool
	}{
		{
			name:                "AccessDenied error",
			apiError:            fmt.Errorf("AccessDenied: User is not authorized"),
			expectMetrics:       false,
			expectErrorIncrease: true,
		},
		{
			name:                "InvalidParameterValue error",
			apiError:            fmt.Errorf("InvalidParameterValue: Invalid engine name"),
			expectMetrics:       false,
			expectErrorIncrease: true,
		},
		{
			name:                "RequestLimitExceeded error",
			apiError:            fmt.Errorf("RequestLimitExceeded: Request rate exceeded"),
			expectMetrics:       false,
			expectErrorIncrease: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			awsAccountID := "123456789012"
			awsRegion := "eu-west-3"

			// Create PostgreSQL instance
			postgresInstance := createPostgreSQLInstance("test-postgres", "14.9")

			logger, _ := logger.New(true, "text")
			rdsClient := rds_mock.NewRDSClient().
				WithDBInstances(postgresInstance).
				WithDescribeDBMajorEngineVersionsError(tt.apiError)

			ec2Client := ec2_mock.EC2Client{}
			cloudWatchClient := cloudwatch_mock.CloudwatchClient{}
			servicequotasClient := servicequotas_mock.ServiceQuotasClient{}

			configuration := exporter.Configuration{
				CollectInstanceMetrics: true,
				CollectInstanceTypes:   false,
				CollectInstanceTags:    false,
				CollectLogsSize:        false,
				CollectMaintenances:    false,
				CollectQuotas:          false,
				CollectUsages:          false,
				CollectEngineSupport:   true,
			}

			collector := exporter.NewCollector(*logger, configuration, awsAccountID, awsRegion, rdsClient, ec2Client, cloudWatchClient, servicequotasClient, nil)

			// Get initial error count
			initialStats := collector.GetStatistics()
			initialErrors := initialStats.Errors

			// Collect metrics
			registry := prometheus.NewRegistry()
			registry.MustRegister(collector)

			metricFamilies, err := registry.Gather()
			require.NoError(t, err)

			// Check for engine support metrics
			var hasStandardSupport, hasExtendedSupport bool
			for _, mf := range metricFamilies {
				switch mf.GetName() {
				case "rds_standard_support_engine_remaining_days":
					hasStandardSupport = len(mf.GetMetric()) > 0
				case "rds_extended_support_engine_remaining_days":
					hasExtendedSupport = len(mf.GetMetric()) > 0
				}
			}

			if tt.expectMetrics {
				assert.True(t, hasStandardSupport || hasExtendedSupport, "Should have engine support metrics")
			} else {
				assert.False(t, hasStandardSupport, "Should not have standard support metrics")
				assert.False(t, hasExtendedSupport, "Should not have extended support metrics")
			}

			// Check error count increase
			finalStats := collector.GetStatistics()
			finalErrors := finalStats.Errors

			if tt.expectErrorIncrease {
				assert.Greater(t, finalErrors, initialErrors, "Error count should increase")
			} else {
				assert.Equal(t, initialErrors, finalErrors, "Error count should not change")
			}

			// Verify API was called
			assert.Greater(t, rdsClient.GetDescribeDBMajorEngineVersionsCallCount(), 0, "Should attempt API call")
		})
	}
}

// TestEngineSupport_ComplexAPIResponses tests parsing of complex API responses
func TestEngineSupport_ComplexAPIResponses(t *testing.T) {
	tests := []struct {
		name                    string
		apiResponse             *aws_rds.DescribeDBMajorEngineVersionsOutput
		instanceVersion         string
		expectedStandardMetrics int
		expectedExtendedMetrics int
		description             string
	}{
		{
			name: "multiple versions with mixed support types",
			apiResponse: &aws_rds.DescribeDBMajorEngineVersionsOutput{
				DBMajorEngineVersions: []aws_rds_types.DBMajorEngineVersion{
					{
						Engine:             aws.String("postgres"),
						MajorEngineVersion: aws.String("12"),
						SupportedEngineLifecycles: []aws_rds_types.SupportedEngineLifecycle{
							{
								LifecycleSupportName:    aws_rds_types.LifecycleSupportNameOpenSourceRdsStandardSupport,
								LifecycleSupportEndDate: timePtr(time.Now().AddDate(0, 0, -30)), // Past
							},
							{
								LifecycleSupportName:    aws_rds_types.LifecycleSupportNameOpenSourceRdsExtendedSupport,
								LifecycleSupportEndDate: timePtr(time.Now().AddDate(0, 3, 0)), // Future
							},
						},
					},
					{
						Engine:             aws.String("postgres"),
						MajorEngineVersion: aws.String("13"),
						SupportedEngineLifecycles: []aws_rds_types.SupportedEngineLifecycle{
							{
								LifecycleSupportName:    aws_rds_types.LifecycleSupportNameOpenSourceRdsStandardSupport,
								LifecycleSupportEndDate: timePtr(time.Now().AddDate(0, 6, 0)), // Future
							},
						},
					},
					{
						Engine:             aws.String("postgres"),
						MajorEngineVersion: aws.String("14"),
						SupportedEngineLifecycles: []aws_rds_types.SupportedEngineLifecycle{
							{
								LifecycleSupportName:    aws_rds_types.LifecycleSupportNameOpenSourceRdsStandardSupport,
								LifecycleSupportEndDate: timePtr(time.Now().AddDate(1, 0, 0)), // Future
							},
							{
								LifecycleSupportName:    aws_rds_types.LifecycleSupportNameOpenSourceRdsExtendedSupport,
								LifecycleSupportEndDate: timePtr(time.Now().AddDate(2, 0, 0)), // Future
							},
						},
					},
				},
			},
			instanceVersion:         "13.7",
			expectedStandardMetrics: 1, // Version 13 has standard support
			expectedExtendedMetrics: 0, // Version 13 has no extended support
			description:             "Instance version 13 should only get standard support metric",
		},
		{
			name: "version with nil end dates",
			apiResponse: &aws_rds.DescribeDBMajorEngineVersionsOutput{
				DBMajorEngineVersions: []aws_rds_types.DBMajorEngineVersion{
					{
						Engine:             aws.String("postgres"),
						MajorEngineVersion: aws.String("15"),
						SupportedEngineLifecycles: []aws_rds_types.SupportedEngineLifecycle{
							{
								LifecycleSupportName:    aws_rds_types.LifecycleSupportNameOpenSourceRdsStandardSupport,
								LifecycleSupportEndDate: nil, // Nil end date
							},
							{
								LifecycleSupportName:    aws_rds_types.LifecycleSupportNameOpenSourceRdsExtendedSupport,
								LifecycleSupportEndDate: timePtr(time.Now().AddDate(1, 0, 0)),
							},
						},
					},
				},
			},
			instanceVersion:         "15.3",
			expectedStandardMetrics: 0, // Nil end date should not generate metric
			expectedExtendedMetrics: 1, // Valid extended support
			description:             "Nil end dates should not generate metrics",
		},
		{
			name: "unknown support type names",
			apiResponse: &aws_rds.DescribeDBMajorEngineVersionsOutput{
				DBMajorEngineVersions: []aws_rds_types.DBMajorEngineVersion{
					{
						Engine:             aws.String("postgres"),
						MajorEngineVersion: aws.String("16"),
						SupportedEngineLifecycles: []aws_rds_types.SupportedEngineLifecycle{
							{
								LifecycleSupportName:    "unknown-support-type",
								LifecycleSupportEndDate: timePtr(time.Now().AddDate(1, 0, 0)),
							},
							{
								LifecycleSupportName:    aws_rds_types.LifecycleSupportNameOpenSourceRdsStandardSupport,
								LifecycleSupportEndDate: timePtr(time.Now().AddDate(0, 6, 0)),
							},
						},
					},
				},
			},
			instanceVersion:         "16.1",
			expectedStandardMetrics: 1, // Known standard support type
			expectedExtendedMetrics: 0, // Unknown support type ignored
			description:             "Unknown support types should be ignored",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			awsAccountID := "123456789012"
			awsRegion := "eu-west-3"

			// Create PostgreSQL instance with specified version
			postgresInstance := createPostgreSQLInstance("test-postgres", tt.instanceVersion)

			logger, _ := logger.New(true, "text")
			rdsClient := rds_mock.NewRDSClient().
				WithDBInstances(postgresInstance).
				WithDescribeDBMajorEngineVersionsOutput(tt.apiResponse)

			ec2Client := ec2_mock.EC2Client{}
			cloudWatchClient := cloudwatch_mock.CloudwatchClient{}
			servicequotasClient := servicequotas_mock.ServiceQuotasClient{}

			configuration := exporter.Configuration{
				CollectInstanceMetrics: true,
				CollectInstanceTypes:   false,
				CollectInstanceTags:    false,
				CollectLogsSize:        false,
				CollectMaintenances:    false,
				CollectQuotas:          false,
				CollectUsages:          false,
				CollectEngineSupport:   true,
			}

			collector := exporter.NewCollector(*logger, configuration, awsAccountID, awsRegion, rdsClient, ec2Client, cloudWatchClient, servicequotasClient, nil)

			// Collect metrics
			registry := prometheus.NewRegistry()
			registry.MustRegister(collector)

			metricFamilies, err := registry.Gather()
			require.NoError(t, err)

			// Extract engine support metrics
			var standardSupportMetrics, extendedSupportMetrics []*dto.Metric
			for _, mf := range metricFamilies {
				switch mf.GetName() {
				case "rds_standard_support_engine_remaining_days":
					standardSupportMetrics = mf.GetMetric()
				case "rds_extended_support_engine_remaining_days":
					extendedSupportMetrics = mf.GetMetric()
				}
			}

			// Verify metric counts
			assert.Len(t, standardSupportMetrics, tt.expectedStandardMetrics, "Standard support metric count mismatch: %s", tt.description)
			assert.Len(t, extendedSupportMetrics, tt.expectedExtendedMetrics, "Extended support metric count mismatch: %s", tt.description)

			// Verify API was called
			assert.Greater(t, rdsClient.GetDescribeDBMajorEngineVersionsCallCount(), 0, "Should call API")
		})
	}
}

// Helper functions for creating test data

type engineVersionData struct {
	version            string
	standardEndDate    time.Time
	extendedEndDate    time.Time
	hasStandardSupport bool
	hasExtendedSupport bool
}

func createPostgreSQLInstance(identifier, version string) aws_rds_types.DBInstance {
	instance := *rds_mock.NewRdsInstance()
	instance.DBInstanceIdentifier = aws.String(identifier)
	instance.Engine = aws.String("postgres")
	instance.EngineVersion = aws.String(version)
	return instance
}

func createMySQLInstance(identifier, version string) aws_rds_types.DBInstance {
	instance := *rds_mock.NewRdsInstance()
	instance.DBInstanceIdentifier = aws.String(identifier)
	instance.Engine = aws.String("mysql")
	instance.EngineVersion = aws.String(version)
	return instance
}

func createMariaDBInstance(identifier, version string) aws_rds_types.DBInstance {
	instance := *rds_mock.NewRdsInstance()
	instance.DBInstanceIdentifier = aws.String(identifier)
	instance.Engine = aws.String("mariadb")
	instance.EngineVersion = aws.String(version)
	return instance
}

func createEngineVersionsResponse(engine string, versions []engineVersionData) *aws_rds.DescribeDBMajorEngineVersionsOutput {
	var dbMajorEngineVersions []aws_rds_types.DBMajorEngineVersion

	for _, v := range versions {
		var lifecycles []aws_rds_types.SupportedEngineLifecycle

		if v.hasStandardSupport {
			lifecycles = append(lifecycles, aws_rds_types.SupportedEngineLifecycle{
				LifecycleSupportName:    aws_rds_types.LifecycleSupportNameOpenSourceRdsStandardSupport,
				LifecycleSupportEndDate: &v.standardEndDate,
			})
		}

		if v.hasExtendedSupport {
			lifecycles = append(lifecycles, aws_rds_types.SupportedEngineLifecycle{
				LifecycleSupportName:    aws_rds_types.LifecycleSupportNameOpenSourceRdsExtendedSupport,
				LifecycleSupportEndDate: &v.extendedEndDate,
			})
		}

		dbMajorEngineVersions = append(dbMajorEngineVersions, aws_rds_types.DBMajorEngineVersion{
			Engine:                    aws.String(engine),
			MajorEngineVersion:        aws.String(v.version),
			SupportedEngineLifecycles: lifecycles,
		})
	}

	return &aws_rds.DescribeDBMajorEngineVersionsOutput{
		DBMajorEngineVersions: dbMajorEngineVersions,
	}
}

func verifyMetricLabels(t *testing.T, metric *dto.Metric, expectedAccountID, expectedRegion string) {
	labels := make(map[string]string)
	for _, label := range metric.GetLabel() {
		labels[label.GetName()] = label.GetValue()
	}

	// Verify required labels are present
	assert.Contains(t, labels, "aws_account_id", "Metric should have aws_account_id label")
	assert.Contains(t, labels, "aws_region", "Metric should have aws_region label")
	assert.Contains(t, labels, "dbidentifier", "Metric should have dbidentifier label")
	assert.Contains(t, labels, "engine", "Metric should have engine label")
	assert.Contains(t, labels, "engine_version", "Metric should have engine_version label")

	// Verify label values
	assert.Equal(t, expectedAccountID, labels["aws_account_id"], "aws_account_id label mismatch")
	assert.Equal(t, expectedRegion, labels["aws_region"], "aws_region label mismatch")
	assert.Equal(t, "postgres", labels["engine"], "engine label should be postgres")
	assert.NotEmpty(t, labels["dbidentifier"], "dbidentifier should not be empty")
	assert.NotEmpty(t, labels["engine_version"], "engine_version should not be empty")
}

func timePtr(t time.Time) *time.Time {
	return &t
}
