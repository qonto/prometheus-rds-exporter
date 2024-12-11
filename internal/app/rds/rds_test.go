package rds_test

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	aws_rds "github.com/aws/aws-sdk-go-v2/service/rds"
	aws_rds_types "github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/qonto/prometheus-rds-exporter/internal/app/rds"
	mock "github.com/qonto/prometheus-rds-exporter/internal/app/rds/mock"
	converter "github.com/qonto/prometheus-rds-exporter/internal/app/unit"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetMetrics(t *testing.T) {
	rdsInstance := mock.NewRdsInstance()
	mockDescribeDBInstancesOutput := &aws_rds.DescribeDBInstancesOutput{DBInstances: []aws_rds_types.DBInstance{*rdsInstance}}

	ctx := context.TODO()
	client := mock.RDSClient{DescribeDBInstancesOutput: mockDescribeDBInstancesOutput}
	configuration := rds.Configuration{CollectLogsSize: true}
	fetcher := rds.NewFetcher(ctx, client, nil, slog.Logger{}, configuration)
	metrics, err := fetcher.GetInstancesMetrics()

	require.NoError(t, err, "GetInstancesMetrics must succeed")

	var emptyInt64 *int64

	m := metrics.Instances[*rdsInstance.DBInstanceIdentifier]
	assert.Equal(t, rds.InstanceStatusAvailable, m.Status, "Instance is available")
	assert.Equal(t, "primary", m.Role, "Should be primary node")
	assert.Equal(t, emptyInt64, m.LogFilesSize, "Log file size mismatch")
	assert.Equal(t, fmt.Sprintf("arn:aws:rds:eu-west-3:123456789012:db:%s", *rdsInstance.DBInstanceIdentifier), m.Arn, "ARN mismatch")

	assert.Equal(t, converter.GigaBytesToBytes(int64(*rdsInstance.AllocatedStorage)), m.AllocatedStorage, "Allocated storage mismatch")
	assert.Equal(t, converter.GigaBytesToBytes(int64(*rdsInstance.MaxAllocatedStorage)), m.MaxAllocatedStorage, "Max allocated storage (aka autoscaling) mismatch")
	assert.Equal(t, int64(*rdsInstance.Iops), m.MaxIops, "Max IOPS mismatch")
	assert.Equal(t, converter.DaystoSeconds(*rdsInstance.BackupRetentionPeriod), m.BackupRetentionPeriod, "Backup retention mismatch")
	assert.Equal(t, *rdsInstance.DeletionProtection, m.DeletionProtection, "Deletion protection mismatch")
	assert.Equal(t, *rdsInstance.MultiAZ, m.MultiAZ, "MultiAZ mismatch")
	assert.Equal(t, *rdsInstance.Engine, m.Engine, "Engine mismatch")
	assert.Equal(t, *rdsInstance.EngineVersion, m.EngineVersion, "Engine version mismatch")
	assert.Equal(t, *rdsInstance.PerformanceInsightsEnabled, m.PerformanceInsightsEnabled, "PerformanceInsights enabled mismatch")
	assert.Equal(t, *rdsInstance.PubliclyAccessible, m.PubliclyAccessible, "PubliclyAccessible mismatch")
	assert.Equal(t, *rdsInstance.DbiResourceId, m.DbiResourceID, "DbiResourceId mismatch")
	assert.Equal(t, *rdsInstance.DBInstanceClass, m.DBInstanceClass, "DBInstanceIdentifier mismatch")
	assert.Equal(t, *rdsInstance.CACertificateIdentifier, m.CACertificateIdentifier, "CACertificateIdentifier mismatch")
	assert.Equal(t, *rdsInstance.CertificateDetails.ValidTill, *m.CertificateValidTill, "CertificateValidTill mismatch")
	assert.Equal(t, "unittest", m.Tags["Environment"], "Environment tag mismatch")
	assert.Equal(t, "sre", m.Tags["Team"], "Team tag mismatch")
}

func TestGP2StorageType(t *testing.T) {
	rdsInstanceWithSmallDisk := mock.NewRdsInstance()
	rdsInstanceWithSmallDisk.StorageType = aws.String("gp2")
	rdsInstanceWithSmallDisk.AllocatedStorage = aws.Int32(10)

	rdsInstanceWithMediumDisk := mock.NewRdsInstance()
	rdsInstanceWithMediumDisk.StorageType = aws.String("gp2")
	rdsInstanceWithMediumDisk.AllocatedStorage = aws.Int32(400)

	rdsInstanceWithLargeDisk := mock.NewRdsInstance()
	rdsInstanceWithLargeDisk.StorageType = aws.String("gp2")
	rdsInstanceWithLargeDisk.AllocatedStorage = aws.Int32(20000)

	ctx := context.TODO()
	mockDescribeDBInstancesOutput := &aws_rds.DescribeDBInstancesOutput{DBInstances: []aws_rds_types.DBInstance{*rdsInstanceWithSmallDisk, *rdsInstanceWithMediumDisk, *rdsInstanceWithLargeDisk}}
	client := mock.RDSClient{DescribeDBInstancesOutput: mockDescribeDBInstancesOutput}
	configuration := rds.Configuration{}
	fetcher := rds.NewFetcher(ctx, client, nil, slog.Logger{}, configuration)
	metrics, err := fetcher.GetInstancesMetrics()

	require.NoError(t, err, "GetInstancesMetrics must succeed")
	assert.Equal(t, int64(100), metrics.Instances[*rdsInstanceWithSmallDisk.DBInstanceIdentifier].MaxIops, "Minimum is 100 IOPS")
	assert.Equal(t, converter.MegaBytesToBytes(int64(128)), metrics.Instances[*rdsInstanceWithSmallDisk.DBInstanceIdentifier].StorageThroughput, "Minimum is 128 MiB/s")

	assert.Equal(t, int64(1200), metrics.Instances[*rdsInstanceWithMediumDisk.DBInstanceIdentifier].MaxIops, "Should be 3 * disk size")
	assert.Equal(t, converter.MegaBytesToBytes(int64(250)), metrics.Instances[*rdsInstanceWithMediumDisk.DBInstanceIdentifier].StorageThroughput, "Max 250 MiB/s")

	assert.Equal(t, int64(16000), metrics.Instances[*rdsInstanceWithLargeDisk.DBInstanceIdentifier].MaxIops, "Should be limited to 16K")
	assert.Equal(t, converter.MegaBytesToBytes(int64(250)), metrics.Instances[*rdsInstanceWithLargeDisk.DBInstanceIdentifier].StorageThroughput, "Large volume are limited to 250 MiB/s")
}

func TestGP3StorageType(t *testing.T) {
	rdsInstanceWithSmallDisk := mock.NewRdsInstance()
	rdsInstanceWithSmallDisk.StorageType = aws.String("gp3")
	rdsInstanceWithSmallDisk.AllocatedStorage = aws.Int32(10)
	rdsInstanceWithSmallDisk.Iops = aws.Int32(3000)

	rdsInstanceWithLargeDisk := mock.NewRdsInstance()
	rdsInstanceWithLargeDisk.StorageType = aws.String("gp3")
	rdsInstanceWithLargeDisk.AllocatedStorage = aws.Int32(500)
	rdsInstanceWithLargeDisk.Iops = aws.Int32(12000)

	ctx := context.TODO()
	mockDescribeDBInstancesOutput := &aws_rds.DescribeDBInstancesOutput{DBInstances: []aws_rds_types.DBInstance{*rdsInstanceWithSmallDisk, *rdsInstanceWithLargeDisk}}
	client := mock.RDSClient{DescribeDBInstancesOutput: mockDescribeDBInstancesOutput}
	configuration := rds.Configuration{}
	fetcher := rds.NewFetcher(ctx, client, nil, slog.Logger{}, configuration)
	metrics, err := fetcher.GetInstancesMetrics()

	require.NoError(t, err, "GetInstancesMetrics must succeed")
	assert.Equal(t, int64(3000), metrics.Instances[*rdsInstanceWithSmallDisk.DBInstanceIdentifier].MaxIops, "IOPS should the same than RDS instance information")
	assert.Equal(t, int64(12000), metrics.Instances[*rdsInstanceWithLargeDisk.DBInstanceIdentifier].MaxIops, "IOPS should the same than RDS instance information")
}

func TestIO1StorageType(t *testing.T) {
	rdsInstanceWithSmallIOPS := mock.NewRdsInstance()
	rdsInstanceWithSmallIOPS.StorageType = aws.String("io1")
	rdsInstanceWithSmallIOPS.Iops = aws.Int32(1000)

	rdsInstanceWithMediumIOPS := mock.NewRdsInstance()
	rdsInstanceWithMediumIOPS.StorageType = aws.String("io1")
	rdsInstanceWithMediumIOPS.Iops = aws.Int32(4000)

	rdsInstanceWithLargeIOPS := mock.NewRdsInstance()
	rdsInstanceWithLargeIOPS.StorageType = aws.String("io1")
	rdsInstanceWithLargeIOPS.Iops = aws.Int32(48000)

	rdsInstanceWithHighIOPS := mock.NewRdsInstance()
	rdsInstanceWithHighIOPS.StorageType = aws.String("io1")
	rdsInstanceWithHighIOPS.Iops = aws.Int32(64000)

	ctx := context.TODO()
	mockDescribeDBInstancesOutput := &aws_rds.DescribeDBInstancesOutput{DBInstances: []aws_rds_types.DBInstance{*rdsInstanceWithSmallIOPS, *rdsInstanceWithMediumIOPS, *rdsInstanceWithLargeIOPS, *rdsInstanceWithHighIOPS}}
	client := mock.RDSClient{DescribeDBInstancesOutput: mockDescribeDBInstancesOutput}
	configuration := rds.Configuration{}
	fetcher := rds.NewFetcher(ctx, client, nil, slog.Logger{}, configuration)
	metrics, err := fetcher.GetInstancesMetrics()

	require.NoError(t, err, "GetInstancesMetrics must succeed")
	assert.Equal(t, converter.MegaBytesToBytes(int64(250)), metrics.Instances[*rdsInstanceWithSmallIOPS.DBInstanceIdentifier].StorageThroughput, "Minimum is 256 MiB/s")
	assert.Equal(t, converter.MegaBytesToBytes(int64(500)), metrics.Instances[*rdsInstanceWithMediumIOPS.DBInstanceIdentifier].StorageThroughput, "500 MiB/s for more than 2K IOPS")
	assert.Equal(t, converter.MegaBytesToBytes(int64(750)), metrics.Instances[*rdsInstanceWithLargeIOPS.DBInstanceIdentifier].StorageThroughput, "16 * IOPS")
	assert.Equal(t, converter.MegaBytesToBytes(int64(1000)), metrics.Instances[*rdsInstanceWithHighIOPS.DBInstanceIdentifier].StorageThroughput, "Max is 1 GiB/s")
}

func TestIO2StorageType(t *testing.T) {
	rdsInstanceWithSmallIOPS := mock.NewRdsInstance()
	rdsInstanceWithSmallIOPS.StorageType = aws.String("io2")
	rdsInstanceWithSmallIOPS.Iops = aws.Int32(1000)

	rdsInstanceWithMediumIOPS := mock.NewRdsInstance()
	rdsInstanceWithMediumIOPS.StorageType = aws.String("io2")
	rdsInstanceWithMediumIOPS.Iops = aws.Int32(4000)

	rdsInstanceWithLargeIOPS := mock.NewRdsInstance()
	rdsInstanceWithLargeIOPS.StorageType = aws.String("io2")
	rdsInstanceWithLargeIOPS.Iops = aws.Int32(48000)

	rdsInstanceWithHighIOPS := mock.NewRdsInstance()
	rdsInstanceWithHighIOPS.StorageType = aws.String("io2")
	rdsInstanceWithHighIOPS.Iops = aws.Int32(64000)

	mockDescribeDBInstancesOutput := &aws_rds.DescribeDBInstancesOutput{DBInstances: []aws_rds_types.DBInstance{*rdsInstanceWithSmallIOPS, *rdsInstanceWithMediumIOPS, *rdsInstanceWithLargeIOPS, *rdsInstanceWithHighIOPS}}
	client := mock.RDSClient{DescribeDBInstancesOutput: mockDescribeDBInstancesOutput}
	configuration := rds.Configuration{}
	ctx := context.TODO()
	fetcher := rds.NewFetcher(ctx, client, nil, slog.Logger{}, configuration)
	metrics, err := fetcher.GetInstancesMetrics()

	require.NoError(t, err, "GetInstancesMetrics must succeed")
	assert.Equal(t, converter.MegaBytesToBytes(int64(256)), metrics.Instances[*rdsInstanceWithSmallIOPS.DBInstanceIdentifier].StorageThroughput, "Minimum is 256 MiB/s")
	assert.Equal(t, converter.MegaBytesToBytes(int64(1024)), metrics.Instances[*rdsInstanceWithMediumIOPS.DBInstanceIdentifier].StorageThroughput, "500 MiB/s for more than 2K IOPS")
	assert.Equal(t, converter.MegaBytesToBytes(int64(4000)), metrics.Instances[*rdsInstanceWithLargeIOPS.DBInstanceIdentifier].StorageThroughput, "16 * IOPS")
	assert.Equal(t, converter.MegaBytesToBytes(int64(4000)), metrics.Instances[*rdsInstanceWithHighIOPS.DBInstanceIdentifier].StorageThroughput, "Max is 4 GiB/s")
}

func TestLogSize(t *testing.T) {
	// Mock RDS instance
	rdsInstance := mock.NewRdsInstance()
	mockDescribeDBInstancesOutput := &aws_rds.DescribeDBInstancesOutput{DBInstances: []aws_rds_types.DBInstance{*rdsInstance}}

	// Mock log files
	logFileCount := int64(3)
	logFileSize := int64(1024)
	expectedLogFilesSize := logFileSize * logFileCount

	rdsLogFiles := []aws_rds_types.DescribeDBLogFilesDetails{}
	for i := int64(0); i < logFileCount; i++ {
		rdsLogFiles = append(rdsLogFiles, aws_rds_types.DescribeDBLogFilesDetails{Size: aws.Int64(logFileSize)})
	}

	mockDescribeDBLogFilesOutput := &aws_rds.DescribeDBLogFilesOutput{DescribeDBLogFiles: rdsLogFiles}

	client := mock.RDSClient{
		DescribeDBInstancesOutput: mockDescribeDBInstancesOutput,
		DescribeDBLogFilesOutput:  mockDescribeDBLogFilesOutput,
	}
	configuration := rds.Configuration{CollectLogsSize: true}
	ctx := context.TODO()
	fetcher := rds.NewFetcher(ctx, client, nil, slog.Logger{}, configuration)
	metrics, err := fetcher.GetInstancesMetrics()

	require.NoError(t, err, "GetInstancesMetrics must succeed")
	assert.Equal(t, aws.Int64(expectedLogFilesSize), metrics.Instances[*rdsInstance.DBInstanceIdentifier].LogFilesSize, "Log files size mismatch")
}

func TestLogSizeInCreation(t *testing.T) {
	// Mock RDS instance
	rdsInstance := mock.NewRdsInstance()
	mockDescribeDBInstancesOutput := &aws_rds.DescribeDBInstancesOutput{DBInstances: []aws_rds_types.DBInstance{*rdsInstance}}

	client := mock.RDSClient{
		DescribeDBInstancesOutput:     mockDescribeDBInstancesOutput,
		DescribeDBLogFilesOutputError: &aws_rds_types.DBInstanceNotFoundFault{},
	}
	configuration := rds.Configuration{CollectLogsSize: true}
	ctx := context.TODO()
	fetcher := rds.NewFetcher(ctx, client, nil, slog.Logger{}, configuration)
	metrics, err := fetcher.GetInstancesMetrics()

	var emptyInt64 *int64

	require.NoError(t, err, "GetInstancesMetrics must succeed")
	assert.Equal(t, emptyInt64, metrics.Instances[*rdsInstance.DBInstanceIdentifier].LogFilesSize, "Log files size mismatch")
}

func TestReplicaNode(t *testing.T) {
	primaryInstance := "primary-instance"

	// Mock RDS instance
	rdsInstance := mock.NewRdsInstance()
	rdsInstance.ReadReplicaSourceDBInstanceIdentifier = aws.String(primaryInstance)
	mockDescribeDBInstancesOutput := &aws_rds.DescribeDBInstancesOutput{DBInstances: []aws_rds_types.DBInstance{*rdsInstance}}

	client := mock.RDSClient{DescribeDBInstancesOutput: mockDescribeDBInstancesOutput}
	configuration := rds.Configuration{CollectLogsSize: true}
	ctx := context.TODO()
	fetcher := rds.NewFetcher(ctx, client, nil, slog.Logger{}, configuration)
	metrics, err := fetcher.GetInstancesMetrics()

	require.NoError(t, err, "GetInstancesMetrics must succeed")
	assert.Equal(t, "replica", metrics.Instances[*rdsInstance.DBInstanceIdentifier].Role, "Should be replica")
	assert.Equal(t, primaryInstance, metrics.Instances[*rdsInstance.DBInstanceIdentifier].SourceDBInstanceIdentifier, "Should be replica")
}

func TestThresholdValue(t *testing.T) {
	assert.Equal(t, int64(100), rds.ThresholdValue(100, 42, 1000), "Should return minimum")
	assert.Equal(t, int64(500), rds.ThresholdValue(100, 500, 1000), "Should return the value")
	assert.Equal(t, int64(1000), rds.ThresholdValue(100, 999999, 1000), "Should return the maximum")
}

func TestGetDBIdentifierFromARN(t *testing.T) {
	assert.Equal(t, "pg1", rds.GetDBIdentifierFromARN("arn:aws:rds:eu-west-3:123456789012:db:pg1"), "Should return only the dbidentifier")
}

func TestGetDBInstanceStatusCode(t *testing.T) {
	type test struct {
		input string
		want  int
	}

	tests := []test{
		{input: "available", want: rds.InstanceStatusAvailable},
		{input: "backing-up", want: rds.InstanceStatusBackingUp},
		{input: "creating", want: rds.InstanceStatusCreating},
		{input: "deleting", want: rds.InstanceStatusDeleting},
		{input: "future", want: rds.InstanceStatusUnknown},
		{input: "modifying", want: rds.InstanceStatusModifying},
		{input: "stopped", want: rds.InstanceStatusStopped},
		{input: "stopping", want: rds.InstanceStatusStopping},
		{input: "unknown", want: rds.InstanceStatusUnknown},
	}

	for _, tc := range tests {
		got := rds.GetDBInstanceStatusCode(tc.input)
		if !reflect.DeepEqual(tc.want, got) {
			t.Fatalf("expected: %v, got: %v", tc.want, got)
		}
	}
}

func TestPendingModification(t *testing.T) {
	// Mock RDS instance
	rdsInstance := mock.NewRdsInstance()
	mockDescribeDBInstancesOutput := &aws_rds.DescribeDBInstancesOutput{DBInstances: []aws_rds_types.DBInstance{*rdsInstance}}

	client := mock.RDSClient{DescribeDBInstancesOutput: mockDescribeDBInstancesOutput}
	configuration := rds.Configuration{}
	ctx := context.TODO()
	fetcher := rds.NewFetcher(ctx, client, nil, slog.Logger{}, configuration)
	metrics, err := fetcher.GetInstancesMetrics()

	require.NoError(t, err, "GetInstancesMetrics must succeed")
	assert.Equal(t, false, metrics.Instances[*rdsInstance.DBInstanceIdentifier].PendingModifiedValues, "Should not have any pending modification")
}

func TestPendingModificationDueToInstanceModification(t *testing.T) {
	// Mock RDS instance
	rdsInstance := mock.NewRdsInstance()
	pendingModifications := aws_rds_types.PendingModifiedValues{AllocatedStorage: aws.Int32(int32(42))}
	rdsInstance.PendingModifiedValues = &pendingModifications
	mockDescribeDBInstancesOutput := &aws_rds.DescribeDBInstancesOutput{DBInstances: []aws_rds_types.DBInstance{*rdsInstance}}

	ctx := context.TODO()
	client := mock.RDSClient{DescribeDBInstancesOutput: mockDescribeDBInstancesOutput}
	configuration := rds.Configuration{}
	fetcher := rds.NewFetcher(ctx, client, nil, slog.Logger{}, configuration)
	metrics, err := fetcher.GetInstancesMetrics()

	require.NoError(t, err, "GetInstancesMetrics must succeed")
	assert.Equal(t, true, metrics.Instances[*rdsInstance.DBInstanceIdentifier].PendingModifiedValues, "Should have allocated storage pending modification")
}

func TestPendingModificationDueToUnappliedParameterGroup(t *testing.T) {
	// Mock RDS instance
	rdsInstance := mock.NewRdsInstance()
	rdsInstance.DBParameterGroups = []aws_rds_types.DBParameterGroupStatus{{DBParameterGroupName: aws.String("my_parameter_group"), ParameterApplyStatus: aws.String("pending-reboot")}}
	mockDescribeDBInstancesOutput := &aws_rds.DescribeDBInstancesOutput{DBInstances: []aws_rds_types.DBInstance{*rdsInstance}}

	ctx := context.TODO()
	client := mock.RDSClient{DescribeDBInstancesOutput: mockDescribeDBInstancesOutput}
	configuration := rds.Configuration{}
	fetcher := rds.NewFetcher(ctx, client, nil, slog.Logger{}, configuration)
	metrics, err := fetcher.GetInstancesMetrics()

	require.NoError(t, err, "GetInstancesMetrics must succeed")
	assert.Equal(t, true, metrics.Instances[*rdsInstance.DBInstanceIdentifier].PendingModifiedValues, "Should have pending modification")
}

func TestInstanceAge(t *testing.T) {
	// Mock RDS instance
	rdsInstance := mock.NewRdsInstance()
	creationDate := time.Date(2023, 9, 25, 12, 25, 0, 0, time.UTC) // Date of our first release
	rdsInstance.InstanceCreateTime = &creationDate
	mockDescribeDBInstancesOutput := &aws_rds.DescribeDBInstancesOutput{DBInstances: []aws_rds_types.DBInstance{*rdsInstance}}

	ctx := context.TODO()
	client := mock.RDSClient{DescribeDBInstancesOutput: mockDescribeDBInstancesOutput}
	configuration := rds.Configuration{}
	fetcher := rds.NewFetcher(ctx, client, nil, slog.Logger{}, configuration)
	metrics, err := fetcher.GetInstancesMetrics()

	expectedAge := time.Since(*rdsInstance.InstanceCreateTime)

	require.NoError(t, err, "GetInstancesMetrics must succeed")
	assert.Equal(t, int(expectedAge.Seconds()), int(*metrics.Instances[*rdsInstance.DBInstanceIdentifier].Age), "Age should match expected age")
}
