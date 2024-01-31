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

	instanceTypes := []string{"db.t3.large", "db.t3.small"}
	fetcher := ec2.NewFetcher(context, client)
	result, err := fetcher.GetDBInstanceTypeInformation(instanceTypes)

	require.NoError(t, err, "GetDBInstanceTypeInformation must succeed")
	assert.Equal(t, mock.InstanceT3Large.Vcpu, result.Instances["db.t3.large"].Vcpu, "vCPU don't match")
	assert.Equal(t, converter.MegaBytesToBytes(mock.InstanceT3Large.Memory), result.Instances["db.t3.large"].Memory, "Memory don't match")
	assert.Equal(t, mock.InstanceT3Large.MaximumIops, result.Instances["db.t3.large"].MaximumIops, "MaximumThroughput don't match")
	assert.Equal(t, converter.MegaBytesToBytes(mock.InstanceT3Large.MaximumThroughput), result.Instances["db.t3.large"].MaximumThroughput, "MaximumThroughput don't match")

	assert.Equal(t, mock.InstanceT3Small.Vcpu, result.Instances["db.t3.small"].Vcpu, "vCPU don't match")
	assert.Equal(t, converter.MegaBytesToBytes(mock.InstanceT3Small.Memory), result.Instances["db.t3.small"].Memory, "Memory don't match")

	assert.Equal(t, float64(1), fetcher.GetStatistics().EC2ApiCall, "EC2 API call don't match")
}
