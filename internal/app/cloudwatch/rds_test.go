package cloudwatch_test

import (
	"fmt"
	"log/slog"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	aws_cloudwatch_types "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/qonto/prometheus-rds-exporter/internal/app/cloudwatch"
	cloudwatch_mock "github.com/qonto/prometheus-rds-exporter/internal/app/cloudwatch/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var db1ExpecteRdsMetrics = cloudwatch.RdsMetrics{
	BurstBalance:              aws.Float64(100),
	CheckpointLag:             aws.Float64(0),
	CPUCreditBalance:          aws.Float64(100),
	CPUCreditUsage:            aws.Float64(10),
	CPUSurplusCreditBalance:   aws.Float64(0),
	CPUSurplusCreditsCharged:  aws.Float64(0),
	CPUUtilization:            aws.Float64(10),
	DBLoad:                    aws.Float64(1),
	DBLoadCPU:                 aws.Float64(2),
	DBLoadNonCPU:              aws.Float64(4),
	DatabaseConnections:       aws.Float64(42),
	DiskQueueDepth:            aws.Float64(0),
	EBSByteBalance:            aws.Float64(100),
	EBSIOBalance:              aws.Float64(100),
	FreeStorageSpace:          aws.Float64(5),
	FreeableMemory:            aws.Float64(10),
	MaximumUsedTransactionIDs: aws.Float64(1000000),
	NetworkReceiveThroughput:  aws.Float64(1000),
	NetworkTransmitThroughput: aws.Float64(1000),
	OldestReplicationSlotLag:  aws.Float64(0),
	ReadLatency:               aws.Float64(0.1),
	ReadIOPS:                  aws.Float64(100),
	ReadThroughput:            aws.Float64(101),
	ReplicaLag:                aws.Float64(42),
	ReplicationSlotDiskUsage:  aws.Float64(100),
	SwapUsage:                 aws.Float64(10),
	TransactionLogsDiskUsage:  aws.Float64(24),
	TransactionLogsGeneration: aws.Float64(100),
	WriteLatency:              aws.Float64(0.1),
	WriteIOPS:                 aws.Float64(11),
	WriteThroughput:           aws.Float64(12),
}

var db2ExpecteRdsMetrics = cloudwatch.RdsMetrics{
	BurstBalance:              aws.Float64(100),
	CheckpointLag:             aws.Float64(0),
	CPUCreditBalance:          aws.Float64(100),
	CPUCreditUsage:            aws.Float64(10),
	CPUSurplusCreditBalance:   aws.Float64(0),
	CPUSurplusCreditsCharged:  aws.Float64(0),
	CPUUtilization:            aws.Float64(40),
	DBLoad:                    aws.Float64(2),
	DBLoadCPU:                 aws.Float64(8),
	DBLoadNonCPU:              aws.Float64(1),
	DatabaseConnections:       aws.Float64(1000),
	DiskQueueDepth:            aws.Float64(0),
	EBSByteBalance:            aws.Float64(100),
	EBSIOBalance:              aws.Float64(100),
	FreeStorageSpace:          aws.Float64(10),
	FreeableMemory:            aws.Float64(10),
	MaximumUsedTransactionIDs: aws.Float64(1000000),
	NetworkReceiveThroughput:  aws.Float64(1000),
	NetworkTransmitThroughput: aws.Float64(1000),
	OldestReplicationSlotLag:  aws.Float64(0),
	ReadLatency:               aws.Float64(0.1),
	ReadIOPS:                  aws.Float64(100),
	ReadThroughput:            aws.Float64(101),
	ReplicaLag:                aws.Float64(42),
	ReplicationSlotDiskUsage:  aws.Float64(100),
	SwapUsage:                 aws.Float64(10),
	TransactionLogsDiskUsage:  aws.Float64(24),
	TransactionLogsGeneration: aws.Float64(100),
	WriteLatency:              aws.Float64(0.1),
	WriteIOPS:                 aws.Float64(11),
	WriteThroughput:           aws.Float64(12),
}

// generateMockedMetricsForInstance returns cloudwatch API output for the instance
func generateMockedMetricsForInstance(id int, m cloudwatch.RdsMetrics) []aws_cloudwatch_types.MetricDataResult {
	metrics := []aws_cloudwatch_types.MetricDataResult{
		{
			Id:     aws.String(fmt.Sprintf("burstbalance_%d", id)),
			Label:  aws.String("BurstBalance"),
			Values: []float64{*m.BurstBalance},
		},
		{
			Id:     aws.String(fmt.Sprintf("checkpointlag_%d", id)),
			Label:  aws.String("CheckpointLag"),
			Values: []float64{*m.CheckpointLag},
		},
		{
			Id:     aws.String(fmt.Sprintf("cpucreditbalance_%d", id)),
			Label:  aws.String("CPUCreditBalance"),
			Values: []float64{*m.CPUCreditBalance},
		},
		{
			Id:     aws.String(fmt.Sprintf("cpucreditusage_%d", id)),
			Label:  aws.String("CPUCreditUsage"),
			Values: []float64{*m.CPUCreditUsage},
		},
		{
			Id:     aws.String(fmt.Sprintf("cpusurpluscreditbalance_%d", id)),
			Label:  aws.String("CPUSurplusCreditBalance"),
			Values: []float64{*m.CPUSurplusCreditBalance},
		},
		{
			Id:     aws.String(fmt.Sprintf("cpusurpluscreditscharged_%d", id)),
			Label:  aws.String("CPUSurplusCreditsCharged"),
			Values: []float64{*m.CPUSurplusCreditsCharged},
		},
		{
			Id:     aws.String(fmt.Sprintf("cpuutilization_%d", id)),
			Label:  aws.String("CPUUtilization"),
			Values: []float64{*m.CPUUtilization},
		},
		{
			Id:     aws.String(fmt.Sprintf("dbload_%d", id)),
			Label:  aws.String("DBLoad"),
			Values: []float64{*m.DBLoad},
		},
		{
			Id:     aws.String(fmt.Sprintf("dbloadcpu_%d", id)),
			Label:  aws.String("DBLoadCPU"),
			Values: []float64{*m.DBLoadCPU},
		},
		{
			Id:     aws.String(fmt.Sprintf("dbloadnoncpu_%d", id)),
			Label:  aws.String("DBLoadNonCPU"),
			Values: []float64{*m.DBLoadNonCPU},
		},
		{
			Id:     aws.String(fmt.Sprintf("databaseconnections_%d", id)),
			Label:  aws.String("DatabaseConnections"),
			Values: []float64{*m.DatabaseConnections},
		},
		{
			Id:     aws.String(fmt.Sprintf("diskqueuedepth_%d", id)),
			Label:  aws.String("DiskQueueDepth"),
			Values: []float64{*m.DiskQueueDepth},
		},
		{
			Id:     aws.String(fmt.Sprintf("ebsbytebalance_%d", id)),
			Label:  aws.String("EBSByteBalance%"),
			Values: []float64{*m.EBSByteBalance},
		},
		{
			Id:     aws.String(fmt.Sprintf("ebsiobalance_%d", id)),
			Label:  aws.String("EBSIOBalance%"),
			Values: []float64{*m.EBSIOBalance},
		},
		{
			Id:     aws.String(fmt.Sprintf("freestoragespace_%d", id)),
			Label:  aws.String("FreeStorageSpace"),
			Values: []float64{*m.FreeStorageSpace},
		},
		{
			Id:     aws.String(fmt.Sprintf("freeablememory_%d", id)),
			Label:  aws.String("FreeableMemory"),
			Values: []float64{*m.FreeableMemory},
		},
		{
			Id:     aws.String(fmt.Sprintf("maximumusedtransactionids_%d", id)),
			Label:  aws.String("MaximumUsedTransactionIDs"),
			Values: []float64{*m.MaximumUsedTransactionIDs},
		},
		{
			Id:     aws.String(fmt.Sprintf("networkreceivethroughput_%d", id)),
			Label:  aws.String("NetworkReceiveThroughput"),
			Values: []float64{*m.NetworkReceiveThroughput},
		},
		{
			Id:     aws.String(fmt.Sprintf("networktransmitthroughput_%d", id)),
			Label:  aws.String("NetworkTransmitThroughput"),
			Values: []float64{*m.NetworkTransmitThroughput},
		},
		{
			Id:     aws.String(fmt.Sprintf("oldestreplicationslotlag_%d", id)),
			Label:  aws.String("OldestReplicationSlotLag"),
			Values: []float64{*m.OldestReplicationSlotLag},
		},
		{
			Id:     aws.String(fmt.Sprintf("readlatency_%d", id)),
			Label:  aws.String("ReadLatency"),
			Values: []float64{*m.ReadLatency},
		},
		{
			Id:     aws.String(fmt.Sprintf("readiops_%d", id)),
			Label:  aws.String("ReadIOPS"),
			Values: []float64{*m.ReadIOPS},
		},
		{
			Id:     aws.String(fmt.Sprintf("readthroughput_%d", id)),
			Label:  aws.String("ReadThroughput"),
			Values: []float64{*m.ReadThroughput},
		},
		{
			Id:     aws.String(fmt.Sprintf("replicalag_%d", id)),
			Label:  aws.String("ReplicaLag"),
			Values: []float64{*m.ReplicaLag},
		},
		{
			Id:     aws.String(fmt.Sprintf("replicationslotdiskusage_%d", id)),
			Label:  aws.String("ReplicationSlotDiskUsage"),
			Values: []float64{*m.ReplicationSlotDiskUsage},
		},
		{
			Id:     aws.String(fmt.Sprintf("swapusage_%d", id)),
			Label:  aws.String("SwapUsage"),
			Values: []float64{*m.SwapUsage},
		},
		{
			Id:     aws.String(fmt.Sprintf("transactionlogsdiskusage_%d", id)),
			Label:  aws.String("TransactionLogsDiskUsage"),
			Values: []float64{*m.TransactionLogsDiskUsage},
		},
		{
			Id:     aws.String(fmt.Sprintf("transactionlogsgeneration_%d", id)),
			Label:  aws.String("TransactionLogsGeneration"),
			Values: []float64{*m.TransactionLogsGeneration},
		},
		{
			Id:     aws.String(fmt.Sprintf("writelatency_%d", id)),
			Label:  aws.String("WriteLatency"),
			Values: []float64{*m.WriteLatency},
		},
		{
			Id:     aws.String(fmt.Sprintf("writeiops_%d", id)),
			Label:  aws.String("WriteIOPS"),
			Values: []float64{*m.WriteIOPS},
		},
		{
			Id:     aws.String(fmt.Sprintf("writethroughput_%d", id)),
			Label:  aws.String("WriteThroughput"),
			Values: []float64{*m.WriteThroughput},
		},
	}

	return metrics
}

func TestGetDBInstanceTypeInformation(t *testing.T) {
	instancesName := []string{}
	data := []aws_cloudwatch_types.MetricDataResult{}

	// Generate instances metrics
	instances := make(map[string]cloudwatch.RdsMetrics)
	instances["db1"] = db1ExpecteRdsMetrics
	instances["db2"] = db2ExpecteRdsMetrics

	// Generate Cloudwatch API output metrics
	i := 0

	for id := range instances {
		instancesName = append(instancesName, id)
		instancesMetrics := generateMockedMetricsForInstance(i, instances[id])

		data = append(data, instancesMetrics...)
		i++
	}

	client := cloudwatch_mock.CloudwatchClient{Metrics: data}
	fetcher := cloudwatch.NewRDSFetcher(client, slog.Logger{})
	result, err := fetcher.GetRDSInstanceMetrics(instancesName)

	require.NoError(t, err, "GetRDSInstanceMetrics must succeed")
	assert.Equal(t, float64(1), fetcher.GetStatistics().CloudWatchAPICall, "One call to Cloudwatch API")

	for id, value := range instances {
		assert.Equal(t, value.BurstBalance, result.Instances[id].BurstBalance, "BurstBalance mismatch")
		assert.Equal(t, value.CheckpointLag, result.Instances[id].CheckpointLag, "CheckpointLag mismatch")
		assert.Equal(t, value.CPUCreditBalance, result.Instances[id].CPUCreditBalance, "CPUCreditBalance mismatch")
		assert.Equal(t, value.CPUCreditUsage, result.Instances[id].CPUCreditUsage, "CPUCreditUsage mismatch")
		assert.Equal(t, value.CPUSurplusCreditBalance, result.Instances[id].CPUSurplusCreditBalance, "CPUSurplusCreditBalance mismatch")
		assert.Equal(t, value.CPUSurplusCreditsCharged, result.Instances[id].CPUSurplusCreditsCharged, "CPUSurplusCreditsCharged mismatch")
		assert.Equal(t, value.CPUUtilization, result.Instances[id].CPUUtilization, "CPU utilization mismatch")
		assert.Equal(t, value.DBLoad, result.Instances[id].DBLoad, "DBLoad mismatch")
		assert.Equal(t, value.DBLoadCPU, result.Instances[id].DBLoadCPU, "DBLoadCPU mismatch")
		assert.Equal(t, value.DBLoadNonCPU, result.Instances[id].DBLoadNonCPU, "DBLoadNonCPU mismatch")
		assert.Equal(t, value.DatabaseConnections, result.Instances[id].DatabaseConnections, "DatabaseConnections mismatch")
		assert.Equal(t, value.DiskQueueDepth, result.Instances[id].DiskQueueDepth, "DiskQueueDepth mismatch")
		assert.Equal(t, value.EBSByteBalance, result.Instances[id].EBSByteBalance, "EBSByteBalance mismatch")
		assert.Equal(t, value.EBSIOBalance, result.Instances[id].EBSIOBalance, "EBSIOBalance mismatch")
		assert.Equal(t, value.FreeStorageSpace, result.Instances[id].FreeStorageSpace, "FreeStorageSpace mismatch")
		assert.Equal(t, value.FreeableMemory, result.Instances[id].FreeableMemory, "FreeableMemory mismatch")
		assert.Equal(t, value.MaximumUsedTransactionIDs, result.Instances[id].MaximumUsedTransactionIDs, "MaximumUsedTransactionIDs mismatch")
		assert.Equal(t, value.NetworkReceiveThroughput, result.Instances[id].NetworkReceiveThroughput, "NetworkReceiveThroughput mismatch")
		assert.Equal(t, value.NetworkTransmitThroughput, result.Instances[id].NetworkTransmitThroughput, "NetworkTransmitThroughput mismatch")
		assert.Equal(t, value.OldestReplicationSlotLag, result.Instances[id].OldestReplicationSlotLag, "OldestReplicationSlotLag mismatch")
		assert.Equal(t, value.ReadLatency, result.Instances[id].ReadLatency, "ReadLatency mismatch")
		assert.Equal(t, value.ReadIOPS, result.Instances[id].ReadIOPS, "ReadIOPS mismatch")
		assert.Equal(t, value.ReadThroughput, result.Instances[id].ReadThroughput, "ReadThroughput mismatch")
		assert.Equal(t, value.ReplicaLag, result.Instances[id].ReplicaLag, "ReplicaLag mismatch")
		assert.Equal(t, value.ReplicationSlotDiskUsage, result.Instances[id].ReplicationSlotDiskUsage, "ReplicationSlotDiskUsage mismatch")
		assert.Equal(t, value.SwapUsage, result.Instances[id].SwapUsage, "SwapUsage mismatch")
		assert.Equal(t, value.TransactionLogsDiskUsage, result.Instances[id].TransactionLogsDiskUsage, "TransactionLogsDiskUsage mismatch")
		assert.Equal(t, value.TransactionLogsGeneration, result.Instances[id].TransactionLogsGeneration, "TransactionLogsGeneration mismatch")
		assert.Equal(t, value.WriteLatency, result.Instances[id].WriteLatency, "WriteLatency mismatch")
		assert.Equal(t, value.WriteIOPS, result.Instances[id].WriteIOPS, "WriteIOPS mismatch")
		assert.Equal(t, value.WriteThroughput, result.Instances[id].WriteThroughput, "WriteThroughput mismatch")
	}
}
