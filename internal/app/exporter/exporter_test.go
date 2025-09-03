package exporter_test

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/qonto/prometheus-rds-exporter/internal/app/exporter"
	"github.com/qonto/prometheus-rds-exporter/internal/infra/logger"
	"github.com/stretchr/testify/assert"

	cloudwatch_mock "github.com/qonto/prometheus-rds-exporter/internal/app/cloudwatch/mock"
	ec2_mock "github.com/qonto/prometheus-rds-exporter/internal/app/ec2/mock"
	rds_mock "github.com/qonto/prometheus-rds-exporter/internal/app/rds/mock"
	servicequotas_mock "github.com/qonto/prometheus-rds-exporter/internal/app/servicequotas/mock"
	converter "github.com/qonto/prometheus-rds-exporter/internal/app/unit"
)

func TestWithAllDisabledCollectors(t *testing.T) {
	awsAccountID := "123456789012"
	awsRegion := "eu-west-3"

	rdsInstance := rds_mock.NewRdsInstance()

	logger, _ := logger.New(true, "text")
	rdsClient := rds_mock.NewRDSClient().WithDBInstances(*rdsInstance)
	ec2Client := ec2_mock.EC2Client{}
	cloudWatchClient := cloudwatch_mock.CloudwatchClient{}
	servicequotasClient := servicequotas_mock.ServiceQuotasClient{}

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

	testutil.CollectAndCount(collector)

	counter := collector.GetStatistics()
	assert.Equal(t, float64(0), counter.Errors, "should not have any error")
	assert.Equal(t, float64(2), counter.RDSAPIcalls, "should have 2 call to RDS API (1 for clusters and another for instances)")
	assert.Equal(t, float64(0), counter.EC2APIcalls, "should not have any call")
	assert.Equal(t, float64(0), counter.ServiceQuotasAPICalls, "should not have any call")
	assert.Equal(t, float64(0), counter.UsageAPIcalls, "should not have any call")
	assert.Equal(t, float64(0), counter.CloudwatchAPICalls, "should not have any call")
}

func TestCollector(t *testing.T) {
	awsAccountID := "123456789012"
	awsRegion := "eu-west-3"

	instance := rds_mock.NewRdsInstance()

	logger, _ := logger.New(true, "text")
	rdsClient := rds_mock.NewRDSClient().WithDBInstances(*instance)
	ec2Client := ec2_mock.EC2Client{}
	cloudWatchClient := cloudwatch_mock.CloudwatchClient{}
	servicequotasClient := servicequotas_mock.ServiceQuotasClient{}

	configuration := exporter.Configuration{
		CollectInstanceMetrics: true,
		CollectInstanceTypes:   true,
		CollectInstanceTags:    false,
		CollectLogsSize:        true,
		CollectMaintenances:    true,
		CollectQuotas:          true,
		CollectUsages:          true,
	}

	collector := exporter.NewCollector(*logger, configuration, awsAccountID, awsRegion, rdsClient, ec2Client, cloudWatchClient, servicequotasClient, nil)

	testutil.CollectAndCount(collector)

	// Check API calls
	counter := collector.GetStatistics()
	assert.Equal(t, float64(0), counter.Errors, "should not have any error")
	assert.Equal(t, float64(4), counter.RDSAPIcalls, "should have 4 call to RDS API")
	assert.Equal(t, float64(1), counter.EC2APIcalls, "should have 1 call to EC2 API")
	assert.Equal(t, float64(3), counter.ServiceQuotasAPICalls, "should have 1 call to ServiceQuota API")
	assert.Equal(t, float64(1), counter.UsageAPIcalls, "should have 1 call to UsageAPIcalls API")
	assert.Equal(t, float64(1), counter.CloudwatchAPICalls, "should have 1 call to CloudWatch API")

	// Get internal metrics
	metrics := collector.GetMetrics()

	// Check instance details
	instanceName := instance.DBInstanceIdentifier
	assert.Equal(t, "postgres", metrics.RDS.Instances[*instanceName].Engine, "Engine should match")
	assert.Equal(t, "14.9", metrics.RDS.Instances[*instanceName].EngineVersion, "Version should match")

	// Check serviceQuota information
	assert.Equal(t, servicequotas_mock.DBinstancesQuota, metrics.ServiceQuota.DBinstances, "DBinstance quota should match")
	assert.Equal(t, servicequotas_mock.ManualDBInstanceSnapshots, metrics.ServiceQuota.ManualDBInstanceSnapshots, "Manual instance snapshot quota should match")
	assert.Equal(t, converter.GigaBytesToBytes(servicequotas_mock.TotalStorage), metrics.ServiceQuota.TotalStorage, "TotalStorage quota should match")
}
