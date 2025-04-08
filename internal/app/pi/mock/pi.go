package mocks

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	aws_performanceinsights "github.com/aws/aws-sdk-go-v2/service/pi"
	"github.com/aws/aws-sdk-go-v2/service/pi/types"
	"github.com/aws/smithy-go/middleware"
	"time"
)

type PerformanceInsightsClient struct{}

// GetResourceMetrics calls GetResourceMetricsFunc
func (m PerformanceInsightsClient) GetResourceMetrics(ctx context.Context, params *aws_performanceinsights.GetResourceMetricsInput, optFns ...func(*aws_performanceinsights.Options)) (*aws_performanceinsights.GetResourceMetricsOutput, error) {
	return &aws_performanceinsights.GetResourceMetricsOutput{
		AlignedStartTime: aws.Time(time.Date(2021, time.January, 1, 0, 0, 0, 0, time.UTC)),
		AlignedEndTime:   aws.Time(time.Date(2021, time.January, 1, 0, 0, 0, 0, time.UTC)),
		Identifier:       aws.String("DB-123"),
		MetricList: []types.MetricKeyDataPoints{
			{
				Key: &types.ResponseResourceMetricKey{
					Metric:     aws.String("db.Cache.blks_hit.avg"),
					Dimensions: nil,
				},
				DataPoints: []types.DataPoint{
					{
						Timestamp: aws.Time(time.Date(2021, time.January, 1, 0, 0, 0, 0, time.UTC)),
						Value:     aws.Float64(2),
					},
				},
			},
			{
				Key: &types.ResponseResourceMetricKey{
					Metric:     aws.String("db.Cache.buffers_alloc.avg"),
					Dimensions: nil,
				},
				DataPoints: []types.DataPoint{
					{
						Timestamp: aws.Time(time.Date(2021, time.January, 1, 0, 0, 0, 0, time.UTC)),
						Value:     aws.Float64(0.0005),
					},
					{
						Timestamp: aws.Time(time.Date(2021, time.January, 1, 0, 0, 0, 0, time.UTC)),
						Value:     aws.Float64(0.0006),
					},
				},
			},
		},
		NextToken:      nil,
		ResultMetadata: middleware.Metadata{},
	}, nil
}
