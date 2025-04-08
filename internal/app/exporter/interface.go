package exporter

import (
	"context"
	aws_cloudwatch "github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	aws_ec2 "github.com/aws/aws-sdk-go-v2/service/ec2"
	aws_performanceinsights "github.com/aws/aws-sdk-go-v2/service/pi"
	aws_rds "github.com/aws/aws-sdk-go-v2/service/rds"
	aws_servicequotas "github.com/aws/aws-sdk-go-v2/service/servicequotas"
)

type rdsClient interface {
	DescribeDBInstances(context.Context, *aws_rds.DescribeDBInstancesInput, ...func(*aws_rds.Options)) (*aws_rds.DescribeDBInstancesOutput, error)
	DescribePendingMaintenanceActions(context.Context, *aws_rds.DescribePendingMaintenanceActionsInput, ...func(*aws_rds.Options)) (*aws_rds.DescribePendingMaintenanceActionsOutput, error)
	DescribeDBLogFiles(context.Context, *aws_rds.DescribeDBLogFilesInput, ...func(*aws_rds.Options)) (*aws_rds.DescribeDBLogFilesOutput, error)
}

type EC2Client interface {
	DescribeInstanceTypes(context.Context, *aws_ec2.DescribeInstanceTypesInput, ...func(*aws_ec2.Options)) (*aws_ec2.DescribeInstanceTypesOutput, error)
}

type cloudWatchClient interface {
	GetMetricData(context.Context, *aws_cloudwatch.GetMetricDataInput, ...func(*aws_cloudwatch.Options)) (*aws_cloudwatch.GetMetricDataOutput, error)
}

type performanceInsightsClient interface {
	GetResourceMetrics(ctx context.Context, params *aws_performanceinsights.GetResourceMetricsInput, optFns ...func(*aws_performanceinsights.Options)) (*aws_performanceinsights.GetResourceMetricsOutput, error)
}

type servicequotasClient interface {
	GetServiceQuota(context.Context, *aws_servicequotas.GetServiceQuotaInput, ...func(*aws_servicequotas.Options)) (*aws_servicequotas.GetServiceQuotaOutput, error)
}
