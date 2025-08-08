// Package ec2 implements methods to retrieve EC2 instance capabilities
package ec2

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	aws_ec2 "github.com/aws/aws-sdk-go-v2/service/ec2"
	aws_ec2_types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/qonto/prometheus-rds-exporter/internal/app/trace"
	converter "github.com/qonto/prometheus-rds-exporter/internal/app/unit"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
)

const (
	maxInstanceTypesPerEC2APIRequest int = 100 // Limit the number of instance types per request due to AWS API limits
)

var tracer = otel.Tracer("github/qonto/prometheus-rds-exporter/internal/app/ec2")

type EC2InstanceMetrics struct {
	BaselineIOPS       int32
	BaselineThroughput float64
	MaximumIops        int32
	MaximumThroughput  float64
	Memory             int64
	Vcpu               int32
}

type Metrics struct {
	Instances map[string]EC2InstanceMetrics
}

type Statistics struct {
	EC2ApiCall float64
}

type EC2Client interface {
	DescribeInstanceTypes(ctx context.Context, input *aws_ec2.DescribeInstanceTypesInput, fn ...func(*aws_ec2.Options)) (*aws_ec2.DescribeInstanceTypesOutput, error)
}

func NewFetcher(context context.Context, client EC2Client) *EC2Fetcher {
	return &EC2Fetcher{
		ctx:    context,
		client: client,
	}
}

type EC2Fetcher struct {
	ctx        context.Context
	client     EC2Client
	statistics Statistics
}

func (e *EC2Fetcher) GetStatistics() Statistics {
	return e.statistics
}

// GetDBInstanceTypeInformation returns information about specified AWS EC2 instance types
// AWS RDS API use "db." prefix while AWS EC2 API don't so we must remove it to obtains instance type information
func (e *EC2Fetcher) GetDBInstanceTypeInformation(instanceTypes []string) (Metrics, error) {
	ctx, span := tracer.Start(e.ctx, "collect-ec2-metrics")
	defer span.End()

	metrics := make(map[string]EC2InstanceMetrics)

	for _, instances := range chunkBy(instanceTypes, maxInstanceTypesPerEC2APIRequest) {
		_, instanceTypeSpan := tracer.Start(ctx, "collect-ec2-instance-types-metrics")
		defer instanceTypeSpan.End()

		instanceTypeSpan.SetAttributes(trace.AWSInstanceTypesCount(int64(len(instances))))

		// Remove "db." prefix from instance types
		instanceTypesToFetch := make([]aws_ec2_types.InstanceType, len(instances))
		for i, instance := range instances {
			instanceTypesToFetch[i] = (aws_ec2_types.InstanceType)(overrideInvalidInstanceTypes(removeDBPrefix(instance)))
		}

		input := &aws_ec2.DescribeInstanceTypesInput{InstanceTypes: instanceTypesToFetch}

		resp, err := e.client.DescribeInstanceTypes(context.TODO(), input)
		if err != nil {
			instanceTypeSpan.SetStatus(codes.Error, "can't fetch describe instance types")
			instanceTypeSpan.RecordError(err)

			return Metrics{}, fmt.Errorf("can't fetch describe instance types: %w", err)
		}

		e.statistics.EC2ApiCall++

		for _, i := range resp.InstanceTypes {
			instanceMetrics := EC2InstanceMetrics{}

			if i.VCpuInfo != nil {
				instanceMetrics.Vcpu = aws.ToInt32(i.VCpuInfo.DefaultVCpus)
			}

			if i.MemoryInfo != nil {
				instanceMetrics.Memory = converter.MegaBytesToBytes(aws.ToInt64(i.MemoryInfo.SizeInMiB))
			}

			if i.EbsInfo != nil && i.EbsInfo.EbsOptimizedInfo != nil {
				instanceMetrics.BaselineIOPS = aws.ToInt32(i.EbsInfo.EbsOptimizedInfo.BaselineIops)
				instanceMetrics.BaselineThroughput = converter.MegaBytesToBytes(aws.ToFloat64(i.EbsInfo.EbsOptimizedInfo.BaselineThroughputInMBps))

				instanceMetrics.MaximumIops = aws.ToInt32(i.EbsInfo.EbsOptimizedInfo.MaximumIops)
				instanceMetrics.MaximumThroughput = converter.MegaBytesToBytes(aws.ToFloat64(i.EbsInfo.EbsOptimizedInfo.MaximumThroughputInMBps))
			}

			instanceName := addDBPrefix(string(i.InstanceType))
			metrics[instanceName] = instanceMetrics
		}

		instanceTypeSpan.SetStatus(codes.Ok, "metrics fetched")
	}

	return Metrics{
		Instances: metrics,
	}, nil
}
