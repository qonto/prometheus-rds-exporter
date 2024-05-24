package ec2_test

import (
	"context"
	"testing"

	"github.com/qonto/prometheus-rds-exporter/internal/app/ec2"
	mock "github.com/qonto/prometheus-rds-exporter/internal/app/ec2/mock"
	converter "github.com/qonto/prometheus-rds-exporter/internal/app/unit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetDBInstanceTypeInformation(t *testing.T) {
	context := context.TODO()
	client := mock.EC2Client{}

	testCases := []struct {
		instanceType       string
		vCPU               int32
		memory             int64
		baselineIops       int32
		maximumIops        int32
		baselineThroughput float64
		maximumThroughput  float64
	}{
		{
			baselineIops:       mock.InstanceT3Large.BaselineIOPS,
			baselineThroughput: converter.MegaBytesToBytes(mock.InstanceT3Large.BaselineThroughput),
			instanceType:       "t3.large",
			vCPU:               mock.InstanceT3Large.Vcpu,
			memory:             converter.MegaBytesToBytes(mock.InstanceT3Large.Memory),
			maximumIops:        mock.InstanceT3Large.MaximumIops,
			maximumThroughput:  converter.MegaBytesToBytes(mock.InstanceT3Large.MaximumThroughput),
		},
		{
			baselineIops:       mock.InstanceT3Small.BaselineIOPS,
			baselineThroughput: converter.MegaBytesToBytes(mock.InstanceT3Small.BaselineThroughput),
			instanceType:       "t3.small",
			vCPU:               mock.InstanceT3Small.Vcpu,
			memory:             converter.MegaBytesToBytes(mock.InstanceT3Small.Memory),
			maximumIops:        mock.InstanceT3Small.MaximumIops,
			maximumThroughput:  converter.MegaBytesToBytes(mock.InstanceT3Small.MaximumThroughput),
		},
		{
			baselineIops:       0, // Don't have Maximum IOPS for non EBS optimized instances
			baselineThroughput: 0, // Don't have Maximum IOPS for non EBS optimized instances
			instanceType:       "t2.small",
			vCPU:               mock.InstanceT2Small.Vcpu,
			memory:             converter.MegaBytesToBytes(mock.InstanceT2Small.Memory),
			maximumIops:        0, // Don't have Maximum IOPS for non EBS optimized instances
			maximumThroughput:  0, // Don't have Maximum throughput for non EBS optimized instances
		},
	}
	expectedAPICalls := float64(1)

	instanceTypes := make([]string, len(testCases))
	for i, instance := range testCases {
		instanceTypes[i] = instance.instanceType
	}

	fetcher := ec2.NewFetcher(context, client)
	result, err := fetcher.GetDBInstanceTypeInformation(instanceTypes)

	require.NoError(t, err, "GetDBInstanceTypeInformation must succeed")
	assert.Equal(t, expectedAPICalls, fetcher.GetStatistics().EC2ApiCall, "EC2 API calls don't match")

	for _, tc := range testCases {
		testName := "Test " + tc.instanceType
		t.Run(testName, func(t *testing.T) {
			instance := result.Instances["db."+tc.instanceType]

			assert.Equal(t, tc.vCPU, instance.Vcpu, "vCPU don't match")
			assert.Equal(t, tc.memory, instance.Memory, "Memory don't match")
			assert.Equal(t, tc.baselineIops, instance.BaselineIOPS, "Baseline IOPS don't match")
			assert.Equal(t, tc.baselineThroughput, instance.BaselineThroughput, "Baseline throughput don't match")
			assert.Equal(t, tc.maximumIops, instance.MaximumIops, "Maximum IOPS don't match")
			assert.Equal(t, tc.maximumThroughput, instance.MaximumThroughput, "Maximum throughput don't match")
		})
	}
}
