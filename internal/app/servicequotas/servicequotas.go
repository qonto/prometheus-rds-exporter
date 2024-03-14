// Package servicequotas implements methods to retrieve AWS Service Quotas information
package servicequotas

import (
	"context"
	"errors"
	"fmt"

	aws_servicequotas "github.com/aws/aws-sdk-go-v2/service/servicequotas"
	"github.com/qonto/prometheus-rds-exporter/internal/app/trace"
	converter "github.com/qonto/prometheus-rds-exporter/internal/app/unit"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"golang.org/x/exp/slog"
)

var (
	errNoQuota    = errors.New("no AWS quota with this code")
	errQuotaError = errors.New("AWS return error for this quota")
	tracer        = otel.Tracer("github/qonto/prometheus-rds-exporter/internal/app/servicequotas")
)

const (
	RDSServiceCode = "rds" // AWS RDS service code in AWS quotas API

	// AWS RDS service quotas codes can be listed with "aws service-quotas list-service-quotas --service-code rds"
	DBinstancesQuotacode               = "L-7B6409FD" // DB instances
	TotalStorageQuotaCode              = "L-7ADDB58A" // Total storage for all DB instances
	ManualDBInstanceSnapshotsQuotaCode = "L-272F1212" // Manual DB instance snapshots
)

// Metrics contains the quotas to be monitored for the AWS RDS service
type Metrics struct {
	DBinstances               float64
	TotalStorage              float64
	ManualDBInstanceSnapshots float64
}

type Statistics struct {
	UsageAPICall float64
}

type ServiceQuotasClient interface {
	GetServiceQuota(ctx context.Context, input *aws_servicequotas.GetServiceQuotaInput, optFns ...func(*aws_servicequotas.Options)) (*aws_servicequotas.GetServiceQuotaOutput, error)
}

func NewFetcher(ctx context.Context, client ServiceQuotasClient) *serviceQuotaFetcher {
	return &serviceQuotaFetcher{
		ctx:    ctx,
		client: client,
	}
}

type serviceQuotaFetcher struct {
	ctx        context.Context
	logger     *slog.Logger
	client     ServiceQuotasClient
	statistics Statistics
}

func (s *serviceQuotaFetcher) GetStatistics() Statistics {
	return s.statistics
}

// GetQuota retrieves and returns the AWS quota value for the specified serviceCode and quotaCode
func (s *serviceQuotaFetcher) getQuota(serviceCode string, quotaCode string) (float64, error) {
	_, span := tracer.Start(s.ctx, "get-quota")
	defer span.End()

	span.SetAttributes(trace.AWSQuotaServiceCode(serviceCode), trace.AWSQuotaCode(quotaCode))

	params := &aws_servicequotas.GetServiceQuotaInput{
		ServiceCode: &serviceCode,
		QuotaCode:   &quotaCode,
	}

	s.statistics.UsageAPICall++

	result, err := s.client.GetServiceQuota(s.ctx, params)
	if err != nil {
		span.SetStatus(codes.Error, "failed to get quota")
		span.RecordError(err)

		return 0, fmt.Errorf("can't get %s/%s service quota: %w", serviceCode, quotaCode, err)
	}

	// AWS response payload could contains errors (eg. missing permission)
	if result.Quota.ErrorReason != nil {
		s.logger.Error("AWS quota error: ", "errorCode", result.Quota.ErrorReason.ErrorCode, "message", *result.Quota.ErrorReason.ErrorMessage)
		span.SetStatus(codes.Error, "failed to fetch quota")
		span.RecordError(errQuotaError)

		return 0, errQuotaError
	}

	if result.Quota == nil {
		span.SetStatus(codes.Error, "no quota")
		span.RecordError(errQuotaError)

		return 0, fmt.Errorf("no quota for %s/%s: %w", serviceCode, quotaCode, errNoQuota)
	}

	span.SetStatus(codes.Ok, "quota fetched")

	return *result.Quota.Value, nil
}

// GetRDSQuotas retrieves quotas for the AWS RDS service
func (s *serviceQuotaFetcher) GetRDSQuotas() (Metrics, error) {
	DBinstances, err := s.getQuota(RDSServiceCode, DBinstancesQuotacode)
	if err != nil {
		return Metrics{}, fmt.Errorf("can't fetch DBinstance quota: %w", err)
	}

	totalStorage, err := s.getQuota(RDSServiceCode, TotalStorageQuotaCode)
	if err != nil {
		return Metrics{}, fmt.Errorf("can't fetch total storage quota: %w", err)
	}

	manualDBInstanceSnapshots, err := s.getQuota(RDSServiceCode, ManualDBInstanceSnapshotsQuotaCode)
	if err != nil {
		return Metrics{}, fmt.Errorf("can't fetch manual db instance snapshots quota: %w", err)
	}

	return Metrics{
		DBinstances:               DBinstances,
		TotalStorage:              converter.GigaBytesToBytes(totalStorage),
		ManualDBInstanceSnapshots: manualDBInstanceSnapshots,
	}, nil
}
