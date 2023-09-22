package ec2_test

import (
	"context"
	"testing"

	aws_ec2 "github.com/aws/aws-sdk-go-v2/service/ec2"
	aws_ec2_types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/qonto/prometheus-rds-exporter/internal/app/ec2"
	converter "github.com/qonto/prometheus-rds-exporter/internal/app/unit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var t3Large = ec2.EC2InstanceMetrics{
	MaximumIops:       15700,
	MaximumThroughput: 347.5,
	Memory:            8,
	Vcpu:              2,
}

var t3Small = ec2.EC2InstanceMetrics{
	MaximumIops:       11800,
	MaximumThroughput: 260.62,
	Memory:            2,
	Vcpu:              2,
}

type mockEC2Client struct{}

func (m mockEC2Client) DescribeInstanceTypes(ctx context.Context, input *aws_ec2.DescribeInstanceTypesInput, optFns ...func(*aws_ec2.Options)) (*aws_ec2.DescribeInstanceTypesOutput, error) {
	var instances []aws_ec2_types.InstanceTypeInfo

	for _, instanceType := range input.InstanceTypes {
		//nolint // Hide "missing cases in switch" alert because instanceType has many values. Mock with return empty result for unknown instances
		switch instanceType {
		case "t3.large":
			instances = append(instances, aws_ec2_types.InstanceTypeInfo{
				InstanceType: instanceType,
				VCpuInfo:     &aws_ec2_types.VCpuInfo{DefaultVCpus: &t3Large.Vcpu},
				MemoryInfo:   &aws_ec2_types.MemoryInfo{SizeInMiB: &t3Large.Memory},
				EbsInfo: &aws_ec2_types.EbsInfo{EbsOptimizedInfo: &aws_ec2_types.EbsOptimizedInfo{
					MaximumIops:             &t3Large.MaximumIops,
					MaximumThroughputInMBps: &t3Large.MaximumThroughput,
				}},
			})
		case "t3.small":
			instances = append(instances, aws_ec2_types.InstanceTypeInfo{
				InstanceType: instanceType,
				VCpuInfo:     &aws_ec2_types.VCpuInfo{DefaultVCpus: &t3Small.Vcpu},
				MemoryInfo:   &aws_ec2_types.MemoryInfo{SizeInMiB: &t3Small.Memory},
				EbsInfo: &aws_ec2_types.EbsInfo{EbsOptimizedInfo: &aws_ec2_types.EbsOptimizedInfo{
					MaximumIops:             &t3Small.MaximumIops,
					MaximumThroughputInMBps: &t3Small.MaximumThroughput,
				}},
			})
		}
	}

	return &aws_ec2.DescribeInstanceTypesOutput{InstanceTypes: instances}, nil
}

func TestGetDBInstanceTypeInformation(t *testing.T) {
	mock := mockEC2Client{}

	instanceTypes := []string{"db.t3.large", "db.t3.small"}
	client := ec2.NewFetcher(mock)
	result, err := client.GetDBInstanceTypeInformation(instanceTypes)

	require.NoError(t, err, "GetDBInstanceTypeInformation must succeed")
	assert.Equal(t, t3Large.Vcpu, result.Instances["db.t3.large"].Vcpu, "vCPU don't match")
	assert.Equal(t, converter.MegaBytesToBytes(t3Large.Memory), result.Instances["db.t3.large"].Memory, "Memory don't match")
	assert.Equal(t, t3Large.MaximumIops, result.Instances["db.t3.large"].MaximumIops, "MaximumThroughput don't match")
	assert.Equal(t, converter.MegaBytesToBytes(t3Large.MaximumThroughput), result.Instances["db.t3.large"].MaximumThroughput, "MaximumThroughput don't match")

	assert.Equal(t, t3Small.Vcpu, result.Instances["db.t3.small"].Vcpu, "vCPU don't match")
	assert.Equal(t, converter.MegaBytesToBytes(t3Small.Memory), result.Instances["db.t3.small"].Memory, "Memory don't match")

	assert.Equal(t, float64(1), client.GetStatistics().EC2ApiCall, "EC2 API call don't match")
}
