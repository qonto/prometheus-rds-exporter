// Package mocks contains mock for Cloudwatch client
package mocks

import (
	"context"

	aws_cloudwatch "github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	aws_cloudwatch_types "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
)

type CloudwatchClient struct {
	Metrics []aws_cloudwatch_types.MetricDataResult
}

// GetMetricData returns custom metrics
func (m CloudwatchClient) GetMetricData(ctx context.Context, input *aws_cloudwatch.GetMetricDataInput, fn ...func(*aws_cloudwatch.Options)) (*aws_cloudwatch.GetMetricDataOutput, error) {
	response := &aws_cloudwatch.GetMetricDataOutput{}
	response.MetricDataResults = m.Metrics

	return response, nil
}
