package exporter_test

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	aws_rds "github.com/aws/aws-sdk-go-v2/service/rds"
	aws_rds_types "github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/qonto/prometheus-rds-exporter/internal/app/exporter"
	"github.com/qonto/prometheus-rds-exporter/internal/infra/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	cloudwatch_mock "github.com/qonto/prometheus-rds-exporter/internal/app/cloudwatch/mock"
	ec2_mock "github.com/qonto/prometheus-rds-exporter/internal/app/ec2/mock"
	rds_mock "github.com/qonto/prometheus-rds-exporter/internal/app/rds/mock"
	servicequotas_mock "github.com/qonto/prometheus-rds-exporter/internal/app/servicequotas/mock"
)

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			findSubstring(s, substr))))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestEngineSupport_MetricDescriptors(t *testing.T) {
	awsAccountID := "123456789012"
	awsRegion := "eu-west-3"

	logger, _ := logger.New(true, "text")
	rdsClient := rds_mock.NewRDSClient()
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
	}

	collector := exporter.NewCollector(*logger, configuration, awsAccountID, awsRegion, rdsClient, ec2Client, cloudWatchClient, servicequotasClient, nil)

	// Test that metric descriptors are properly registered
	ch := make(chan *prometheus.Desc, 100)
	collector.Describe(ch)
	close(ch)

	var standardSupportDesc, extendedSupportDesc *prometheus.Desc
	for desc := range ch {
		descStr := desc.String()
		if contains(descStr, "rds_standard_support_engine_remaining_days") {
			standardSupportDesc = desc
		}
		if contains(descStr, "rds_extended_support_engine_remaining_days") {
			extendedSupportDesc = desc
		}
	}

	assert.NotNil(t, standardSupportDesc, "Standard support metric descriptor should be registered")
	assert.NotNil(t, extendedSupportDesc, "Extended support metric descriptor should be registered")
}

func TestEngineSupport_PostgreSQLEngineFiltering(t *testing.T) {
	awsAccountID := "123456789012"
	awsRegion := "eu-west-3"

	// Create test instances with different engines
	postgresInstance := rds_mock.NewRdsInstance()
	postgresInstance.Engine = aws.String("postgres")
	postgresInstance.EngineVersion = aws.String("14.9")
	postgresInstance.DBInstanceIdentifier = aws.String("postgres-instance")

	mysqlInstance := rds_mock.NewRdsInstance()
	mysqlInstance.Engine = aws.String("mysql")
	mysqlInstance.EngineVersion = aws.String("8.0.35")
	mysqlInstance.DBInstanceIdentifier = aws.String("mysql-instance")

	// Create mock engine version data with lifecycle support
	standardEndDate := time.Now().AddDate(0, 6, 0) // 6 months from now
	extendedEndDate := time.Now().AddDate(1, 0, 0) // 1 year from now

	engineVersionsOutput := &aws_rds.DescribeDBMajorEngineVersionsOutput{
		DBMajorEngineVersions: []aws_rds_types.DBMajorEngineVersion{
			{
				Engine:             aws.String("postgres"),
				MajorEngineVersion: aws.String("14"), // Major version only
				SupportedEngineLifecycles: []aws_rds_types.SupportedEngineLifecycle{
					{
						LifecycleSupportName:      aws_rds_types.LifecycleSupportNameOpenSourceRdsStandardSupport,
						LifecycleSupportEndDate:   &standardEndDate,
						LifecycleSupportStartDate: aws.Time(time.Now().AddDate(-1, 0, 0)),
					},
					{
						LifecycleSupportName:      aws_rds_types.LifecycleSupportNameOpenSourceRdsExtendedSupport,
						LifecycleSupportEndDate:   &extendedEndDate,
						LifecycleSupportStartDate: aws.Time(time.Now().AddDate(-1, 0, 0)),
					},
				},
			},
		},
	}

	logger, _ := logger.New(true, "text")
	rdsClient := rds_mock.NewRDSClient().
		WithDBInstances(*postgresInstance, *mysqlInstance).
		WithDescribeDBMajorEngineVersionsOutput(engineVersionsOutput)

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

	// Get metrics
	metricFamilies, err := registry.Gather()
	require.NoError(t, err)

	// Find engine support metrics
	var standardSupportMetrics, extendedSupportMetrics *dto.MetricFamily
	for _, mf := range metricFamilies {
		switch mf.GetName() {
		case "rds_standard_support_engine_remaining_days":
			standardSupportMetrics = mf
		case "rds_extended_support_engine_remaining_days":
			extendedSupportMetrics = mf
		}
	}

	// Verify that only PostgreSQL instances have engine support metrics
	if standardSupportMetrics != nil {
		for _, metric := range standardSupportMetrics.GetMetric() {
			var dbidentifier, engine string
			for _, label := range metric.GetLabel() {
				switch label.GetName() {
				case "dbidentifier":
					dbidentifier = label.GetValue()
				case "engine":
					engine = label.GetValue()
				}
			}
			assert.Equal(t, "postgres", engine, "Only PostgreSQL instances should have engine support metrics")
			assert.Equal(t, "postgres-instance", dbidentifier, "Should be the PostgreSQL instance")
		}
	}

	if extendedSupportMetrics != nil {
		for _, metric := range extendedSupportMetrics.GetMetric() {
			var dbidentifier, engine string
			for _, label := range metric.GetLabel() {
				switch label.GetName() {
				case "dbidentifier":
					dbidentifier = label.GetValue()
				case "engine":
					engine = label.GetValue()
				}
			}
			assert.Equal(t, "postgres", engine, "Only PostgreSQL instances should have engine support metrics")
			assert.Equal(t, "postgres-instance", dbidentifier, "Should be the PostgreSQL instance")
		}
	}

	// Verify API was called for PostgreSQL engine versions
	assert.Greater(t, rdsClient.GetDescribeDBMajorEngineVersionsCallCount(), 0, "Should call DescribeDBMajorEngineVersions API")
}

func TestEngineSupport_MetricEmission(t *testing.T) {
	awsAccountID := "123456789012"
	awsRegion := "eu-west-3"

	// Create PostgreSQL instance
	postgresInstance := rds_mock.NewRdsInstance()
	postgresInstance.Engine = aws.String("postgres")
	postgresInstance.EngineVersion = aws.String("13.7")
	postgresInstance.DBInstanceIdentifier = aws.String("test-postgres")

	// Create mock engine version data with lifecycle support
	standardEndDate := time.Now().AddDate(0, 3, 15) // ~105 days from now
	extendedEndDate := time.Now().AddDate(0, 8, 20) // ~260 days from now

	engineVersionsOutput := &aws_rds.DescribeDBMajorEngineVersionsOutput{
		DBMajorEngineVersions: []aws_rds_types.DBMajorEngineVersion{
			{
				Engine:             aws.String("postgres"),
				MajorEngineVersion: aws.String("13"), // Major version only
				SupportedEngineLifecycles: []aws_rds_types.SupportedEngineLifecycle{
					{
						LifecycleSupportName:      aws_rds_types.LifecycleSupportNameOpenSourceRdsStandardSupport,
						LifecycleSupportEndDate:   &standardEndDate,
						LifecycleSupportStartDate: aws.Time(time.Now().AddDate(-2, 0, 0)),
					},
					{
						LifecycleSupportName:      aws_rds_types.LifecycleSupportNameOpenSourceRdsExtendedSupport,
						LifecycleSupportEndDate:   &extendedEndDate,
						LifecycleSupportStartDate: aws.Time(time.Now().AddDate(-2, 0, 0)),
					},
				},
			},
		},
	}

	logger, _ := logger.New(true, "text")
	rdsClient := rds_mock.NewRDSClient().
		WithDBInstances(*postgresInstance).
		WithDescribeDBMajorEngineVersionsOutput(engineVersionsOutput)

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

	// Test metric emission
	registry := prometheus.NewRegistry()
	registry.MustRegister(collector)

	metricFamilies, err := registry.Gather()
	require.NoError(t, err)

	// Find and verify standard support metric
	var standardSupportMetric *dto.MetricFamily
	for _, mf := range metricFamilies {
		if mf.GetName() == "rds_standard_support_engine_remaining_days" {
			standardSupportMetric = mf
			break
		}
	}

	require.NotNil(t, standardSupportMetric, "Standard support metric should be present")
	require.Len(t, standardSupportMetric.GetMetric(), 1, "Should have one metric for the PostgreSQL instance")

	metric := standardSupportMetric.GetMetric()[0]

	// Verify labels
	labelMap := make(map[string]string)
	for _, label := range metric.GetLabel() {
		labelMap[label.GetName()] = label.GetValue()
	}

	assert.Equal(t, awsAccountID, labelMap["aws_account_id"])
	assert.Equal(t, awsRegion, labelMap["aws_region"])
	assert.Equal(t, "test-postgres", labelMap["dbidentifier"])
	assert.Equal(t, "postgres", labelMap["engine"])
	assert.Equal(t, "13.7", labelMap["engine_version"])

	// Verify metric value is reasonable (should be around 105 days)
	value := metric.GetGauge().GetValue()
	assert.Greater(t, value, 100.0, "Standard support should have more than 100 days remaining")
	assert.Less(t, value, 110.0, "Standard support should have less than 110 days remaining")

	// Find and verify extended support metric
	var extendedSupportMetric *dto.MetricFamily
	for _, mf := range metricFamilies {
		if mf.GetName() == "rds_extended_support_engine_remaining_days" {
			extendedSupportMetric = mf
			break
		}
	}

	require.NotNil(t, extendedSupportMetric, "Extended support metric should be present")
	require.Len(t, extendedSupportMetric.GetMetric(), 1, "Should have one metric for the PostgreSQL instance")

	extMetric := extendedSupportMetric.GetMetric()[0]

	// Verify extended support metric value is reasonable (should be around 260 days)
	extValue := extMetric.GetGauge().GetValue()
	assert.Greater(t, extValue, 250.0, "Extended support should have more than 250 days remaining")
	assert.Less(t, extValue, 270.0, "Extended support should have less than 270 days remaining")
}

func TestEngineSupport_ErrorScenarios(t *testing.T) {
	tests := []struct {
		name               string
		setupMock          func(*rds_mock.RDSClient)
		expectMetrics      bool
		expectAPICall      bool
		expectedErrorCount float64
	}{
		{
			name: "API call failure",
			setupMock: func(client *rds_mock.RDSClient) {
				postgresInstance := rds_mock.NewRdsInstance()
				postgresInstance.Engine = aws.String("postgres")
				postgresInstance.EngineVersion = aws.String("14.9")

				client.WithDBInstances(*postgresInstance).
					WithDescribeDBMajorEngineVersionsError(assert.AnError)
			},
			expectMetrics:      false,
			expectAPICall:      true,
			expectedErrorCount: 1, // API errors are counted as exporter errors
		},
		{
			name: "No lifecycle data available",
			setupMock: func(client *rds_mock.RDSClient) {
				postgresInstance := rds_mock.NewRdsInstance()
				postgresInstance.Engine = aws.String("postgres")
				postgresInstance.EngineVersion = aws.String("14.9")

				// Empty response - no lifecycle data
				engineVersionsOutput := &aws_rds.DescribeDBMajorEngineVersionsOutput{
					DBMajorEngineVersions: []aws_rds_types.DBMajorEngineVersion{
						{
							Engine:                    aws.String("postgres"),
							MajorEngineVersion:        aws.String("14"), // Major version only
							SupportedEngineLifecycles: []aws_rds_types.SupportedEngineLifecycle{},
						},
					},
				}

				client.WithDBInstances(*postgresInstance).
					WithDescribeDBMajorEngineVersionsOutput(engineVersionsOutput)
			},
			expectMetrics:      false,
			expectAPICall:      true,
			expectedErrorCount: 0,
		},
		{
			name: "Missing LifecycleSupportStartDate",
			setupMock: func(client *rds_mock.RDSClient) {
				postgresInstance := rds_mock.NewRdsInstance()
				postgresInstance.Engine = aws.String("postgres")
				postgresInstance.EngineVersion = aws.String("14.9")

				standardEndDate := time.Now().AddDate(0, 6, 0)

				engineVersionsOutput := &aws_rds.DescribeDBMajorEngineVersionsOutput{
					DBMajorEngineVersions: []aws_rds_types.DBMajorEngineVersion{
						{
							Engine:             aws.String("postgres"),
							MajorEngineVersion: aws.String("14"), // Major version only
							SupportedEngineLifecycles: []aws_rds_types.SupportedEngineLifecycle{
								{
									LifecycleSupportName:    aws_rds_types.LifecycleSupportNameOpenSourceRdsStandardSupport,
									LifecycleSupportEndDate: &standardEndDate,
									// Missing LifecycleSupportStartDate - but metrics should still be emitted
								},
							},
						},
					},
				}

				client.WithDBInstances(*postgresInstance).
					WithDescribeDBMajorEngineVersionsOutput(engineVersionsOutput)
			},
			expectMetrics:      true, // Metrics should be emitted even without LifecycleSupportStartDate
			expectAPICall:      true,
			expectedErrorCount: 0,
		},
		{
			name: "Non-PostgreSQL engine ignored",
			setupMock: func(client *rds_mock.RDSClient) {
				mysqlInstance := rds_mock.NewRdsInstance()
				mysqlInstance.Engine = aws.String("mysql")
				mysqlInstance.EngineVersion = aws.String("8.0.35")

				client.WithDBInstances(*mysqlInstance)
			},
			expectMetrics:      false,
			expectAPICall:      true, // API is called for all engines, but no matching lifecycle data
			expectedErrorCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			awsAccountID := "123456789012"
			awsRegion := "eu-west-3"

			logger, _ := logger.New(true, "text")
			rdsClient := rds_mock.NewRDSClient()
			tt.setupMock(rdsClient)

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

			// Check if engine support metrics are present
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

			// Check API call count
			if tt.expectAPICall {
				assert.Greater(t, rdsClient.GetDescribeDBMajorEngineVersionsCallCount(), 0, "Should call DescribeDBMajorEngineVersions API")
			} else {
				assert.Equal(t, 0, rdsClient.GetDescribeDBMajorEngineVersionsCallCount(), "Should not call DescribeDBMajorEngineVersions API")
			}

			// Check error count
			counter := collector.GetStatistics()
			assert.Equal(t, tt.expectedErrorCount, counter.Errors, "Error count should match expected")
		})
	}
}

func TestEngineSupport_NegativeValues(t *testing.T) {
	awsAccountID := "123456789012"
	awsRegion := "eu-west-3"

	// Create PostgreSQL instance
	postgresInstance := rds_mock.NewRdsInstance()
	postgresInstance.Engine = aws.String("postgres")
	postgresInstance.EngineVersion = aws.String("11.20")
	postgresInstance.DBInstanceIdentifier = aws.String("old-postgres")

	// Create mock engine version data with past end dates
	standardEndDate := time.Now().AddDate(0, 0, -30) // 30 days ago
	extendedEndDate := time.Now().AddDate(0, 0, -10) // 10 days ago

	engineVersionsOutput := &aws_rds.DescribeDBMajorEngineVersionsOutput{
		DBMajorEngineVersions: []aws_rds_types.DBMajorEngineVersion{
			{
				Engine:             aws.String("postgres"),
				MajorEngineVersion: aws.String("11"), // Major version only
				SupportedEngineLifecycles: []aws_rds_types.SupportedEngineLifecycle{
					{
						LifecycleSupportName:      aws_rds_types.LifecycleSupportNameOpenSourceRdsStandardSupport,
						LifecycleSupportEndDate:   &standardEndDate,
						LifecycleSupportStartDate: aws.Time(time.Now().AddDate(-3, 0, 0)),
					},
					{
						LifecycleSupportName:      aws_rds_types.LifecycleSupportNameOpenSourceRdsExtendedSupport,
						LifecycleSupportEndDate:   &extendedEndDate,
						LifecycleSupportStartDate: aws.Time(time.Now().AddDate(-3, 0, 0)),
					},
				},
			},
		},
	}

	logger, _ := logger.New(true, "text")
	rdsClient := rds_mock.NewRDSClient().
		WithDBInstances(*postgresInstance).
		WithDescribeDBMajorEngineVersionsOutput(engineVersionsOutput)

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

	// Test metric emission with negative values
	registry := prometheus.NewRegistry()
	registry.MustRegister(collector)

	metricFamilies, err := registry.Gather()
	require.NoError(t, err)

	// Find standard support metric
	var standardSupportMetric *dto.MetricFamily
	for _, mf := range metricFamilies {
		if mf.GetName() == "rds_standard_support_engine_remaining_days" {
			standardSupportMetric = mf
			break
		}
	}

	require.NotNil(t, standardSupportMetric, "Standard support metric should be present even with negative values")
	require.Len(t, standardSupportMetric.GetMetric(), 1, "Should have one metric for the PostgreSQL instance")

	metric := standardSupportMetric.GetMetric()[0]
	value := metric.GetGauge().GetValue()

	// Verify negative value is exposed (past end date)
	assert.Less(t, value, 0.0, "Should expose negative values for past end dates")
	assert.Greater(t, value, -35.0, "Should be around -30 days")
	assert.Less(t, value, -25.0, "Should be around -30 days")

	// Find extended support metric
	var extendedSupportMetric *dto.MetricFamily
	for _, mf := range metricFamilies {
		if mf.GetName() == "rds_extended_support_engine_remaining_days" {
			extendedSupportMetric = mf
			break
		}
	}

	require.NotNil(t, extendedSupportMetric, "Extended support metric should be present even with negative values")
	require.Len(t, extendedSupportMetric.GetMetric(), 1, "Should have one metric for the PostgreSQL instance")

	extMetric := extendedSupportMetric.GetMetric()[0]
	extValue := extMetric.GetGauge().GetValue()

	// Verify negative value is exposed (past end date)
	assert.Less(t, extValue, 0.0, "Should expose negative values for past end dates")
	assert.Greater(t, extValue, -15.0, "Should be around -10 days")
	assert.Less(t, extValue, -5.0, "Should be around -10 days")
}

func TestEngineSupport_MetricsDisabled(t *testing.T) {
	awsAccountID := "123456789012"
	awsRegion := "eu-west-3"

	// Create PostgreSQL instance
	postgresInstance := rds_mock.NewRdsInstance()
	postgresInstance.Engine = aws.String("postgres")
	postgresInstance.EngineVersion = aws.String("14.9")

	logger, _ := logger.New(true, "text")
	rdsClient := rds_mock.NewRDSClient().WithDBInstances(*postgresInstance)
	ec2Client := ec2_mock.EC2Client{}
	cloudWatchClient := cloudwatch_mock.CloudwatchClient{}
	servicequotasClient := servicequotas_mock.ServiceQuotasClient{}

	// Disable instance metrics collection
	configuration := exporter.Configuration{
		CollectInstanceMetrics: false,
		CollectInstanceTypes:   false,
		CollectInstanceTags:    false,
		CollectLogsSize:        false,
		CollectMaintenances:    false,
		CollectQuotas:          false,
		CollectUsages:          false,
	}

	collector := exporter.NewCollector(*logger, configuration, awsAccountID, awsRegion, rdsClient, ec2Client, cloudWatchClient, servicequotasClient, nil)

	// Collect metrics
	registry := prometheus.NewRegistry()
	registry.MustRegister(collector)

	metricFamilies, err := registry.Gather()
	require.NoError(t, err)

	// Verify no engine support metrics are present when instance metrics are disabled
	for _, mf := range metricFamilies {
		assert.NotEqual(t, "rds_standard_support_engine_remaining_days", mf.GetName(), "Should not have standard support metrics when instance metrics disabled")
		assert.NotEqual(t, "rds_extended_support_engine_remaining_days", mf.GetName(), "Should not have extended support metrics when instance metrics disabled")
	}

	// Verify API was NOT called (engine support metrics are not collected when CollectEngineSupport is disabled)
	assert.Equal(t, 0, rdsClient.GetDescribeDBMajorEngineVersionsCallCount(), "Should not call DescribeDBMajorEngineVersions API when engine support collection is disabled")
}
