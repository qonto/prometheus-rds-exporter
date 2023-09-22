package cloudwatch

import (
	"context"

	aws_cloudwatch "github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	aws_cloudwatch_types "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
)

type Statistics struct {
	CloudWatchAPICall float64
}

type CloudWatchMetricRequest struct {
	Query        aws_cloudwatch_types.MetricDataQuery
	Dbidentifier string
	MetricName   string
}

type CloudWatchClient interface {
	GetMetricData(context.Context, *aws_cloudwatch.GetMetricDataInput, ...func(*aws_cloudwatch.Options)) (*aws_cloudwatch.GetMetricDataOutput, error)
}
