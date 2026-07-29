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
		CollectClusterMetrics:  true,
		CollectQuotas:          false,
		CollectUsages:          false,
	}

	collector := exporter.NewCollector(*logger, configuration, awsAccountID, awsRegion, rdsClient, ec2Client, cloudWatchClient, servicequotasClient, nil, make(map[string]interface{}), make(map[string]interface{}), make(map[string]interface{}), make(map[string]interface{}))

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
		CollectClusterMetrics:  true,
		CollectQuotas:          true,
		CollectUsages:          true,
	}

	collector := exporter.NewCollector(*logger, configuration, awsAccountID, awsRegion, rdsClient, ec2Client, cloudWatchClient, servicequotasClient, nil, make(map[string]interface{}), make(map[string]interface{}), make(map[string]interface{}), make(map[string]interface{}))

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
	quota := metrics.ServiceQuota[awsRegion]
	assert.Equal(t, servicequotas_mock.DBinstancesQuota, quota.DBinstances, "DBinstance quota should match")
	assert.Equal(t, servicequotas_mock.ManualDBInstanceSnapshots, quota.ManualDBInstanceSnapshots, "Manual instance snapshot quota should match")
	assert.Equal(t, converter.GigaBytesToBytes(servicequotas_mock.TotalStorage), quota.TotalStorage, "TotalStorage quota should match")
}

func TestCollectorMultiRegion(t *testing.T) {
	awsAccountID := "123456789012"
	awsRegion := "eu-west-3"
	secondaryRegion := "us-east-1"

	primaryInstance := rds_mock.NewRdsInstance()
	secondaryInstance := rds_mock.NewRdsInstance()

	logger, _ := logger.New(true, "text")

	rdsClient := rds_mock.NewRDSClient().WithDBInstances(*primaryInstance)
	secondaryRDSClient := rds_mock.NewRDSClient().WithDBInstances(*secondaryInstance)

	ec2Client := ec2_mock.EC2Client{}
	cloudWatchClient := cloudwatch_mock.CloudwatchClient{}
	servicequotasClient := servicequotas_mock.ServiceQuotasClient{}

	configuration := exporter.Configuration{
		Regions:                []string{awsRegion, secondaryRegion},
		CollectInstanceMetrics: true,
		CollectInstanceTypes:   true,
		CollectInstanceTags:    false,
		CollectLogsSize:        true,
		CollectMaintenances:    true,
		CollectClusterMetrics:  true,
		CollectQuotas:          true,
		CollectUsages:          true,
	}

	rdsRegionClients := map[string]interface{}{
		awsRegion:       rdsClient,
		secondaryRegion: secondaryRDSClient,
	}
	cloudWatchRegionClients := map[string]interface{}{
		awsRegion:       cloudWatchClient,
		secondaryRegion: cloudWatchClient,
	}
	ec2RegionClients := map[string]interface{}{
		awsRegion:       ec2Client,
		secondaryRegion: ec2Client,
	}
	servicequotasRegionClients := map[string]interface{}{
		awsRegion:       servicequotasClient,
		secondaryRegion: servicequotasClient,
	}

	collector := exporter.NewCollector(*logger, configuration, awsAccountID, awsRegion, rdsClient, ec2Client, cloudWatchClient, servicequotasClient, nil, rdsRegionClients, cloudWatchRegionClients, ec2RegionClients, servicequotasRegionClients)

	testutil.CollectAndCount(collector)

	counter := collector.GetStatistics()
	assert.Equal(t, float64(0), counter.Errors, "should not have any error")

	metrics := collector.GetMetrics()

	// Each region's instance should be tagged with its own region
	assert.Equal(t, awsRegion, metrics.RDS.Instances[*primaryInstance.DBInstanceIdentifier].Region, "primary region instance should be tagged with its own region")
	assert.Equal(t, secondaryRegion, metrics.RDS.Instances[*secondaryInstance.DBInstanceIdentifier].Region, "secondary region instance should be tagged with its own region")

	// Quotas should be aggregated per region
	assert.Contains(t, metrics.ServiceQuota, awsRegion, "quotas should include the primary region")
	assert.Contains(t, metrics.ServiceQuota, secondaryRegion, "quotas should include the secondary region")
}
