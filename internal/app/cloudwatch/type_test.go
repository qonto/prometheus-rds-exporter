package cloudwatch_test

import (
	"context"

	aws_cloudwatch "github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	aws_cloudwatch_types "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
)

type mockCloudwatchClient struct {
	metrics []aws_cloudwatch_types.MetricDataResult
}

// GetMetricData returns custom metrics
func (m mockCloudwatchClient) GetMetricData(ctx context.Context, input *aws_cloudwatch.GetMetricDataInput, fn ...func(*aws_cloudwatch.Options)) (*aws_cloudwatch.GetMetricDataOutput, error) {
	response := &aws_cloudwatch.GetMetricDataOutput{}
	response.MetricDataResults = m.metrics

	return response, nil
}
