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
		instanceType      string
		vCPU              int32
		memory            int64
		maximumIops       int32
		maximumThroughput float64
	}{
		{
			instanceType:      "t3.large",
			vCPU:              mock.InstanceT3Large.Vcpu,
			memory:            converter.MegaBytesToBytes(mock.InstanceT3Large.Memory),
			maximumIops:       mock.InstanceT3Large.MaximumIops,
			maximumThroughput: converter.MegaBytesToBytes(mock.InstanceT3Large.MaximumThroughput),
		},
		{
			instanceType:      "t3.small",
			vCPU:              mock.InstanceT3Small.Vcpu,
			memory:            converter.MegaBytesToBytes(mock.InstanceT3Small.Memory),
			maximumIops:       mock.InstanceT3Small.MaximumIops,
			maximumThroughput: converter.MegaBytesToBytes(mock.InstanceT3Small.MaximumThroughput),
		},
	}
	expectedAPICalls := float64(1)

	instanceTypes := []string{"db.t3.large", "db.t3.small"}
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
			assert.Equal(t, tc.maximumIops, instance.MaximumIops, "Maximum IOPS don't match")
			assert.Equal(t, tc.maximumThroughput, instance.MaximumThroughput, "Maximum throughput don't match")
		})
	}
}
