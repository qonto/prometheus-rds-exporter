// Package mocks contains mock for servicequotas client
package mocks

import (
	"context"

	aws_servicequotas "github.com/aws/aws-sdk-go-v2/service/servicequotas"
	aws_servicequotas_types "github.com/aws/aws-sdk-go-v2/service/servicequotas/types"
	"github.com/qonto/prometheus-rds-exporter/internal/app/servicequotas"
)

// Defines expected values for the mock and tests
const (
	UnknownServiceQuota       = float64(42)
	DBinstancesQuota          = float64(10)
	TotalStorage              = float64(10)
	ManualDBInstanceSnapshots = float64(42)
)

type ServiceQuotasClient struct{}

func (m ServiceQuotasClient) GetServiceQuota(context context.Context, input *aws_servicequotas.GetServiceQuotaInput, optFns ...func(*aws_servicequotas.Options)) (*aws_servicequotas.GetServiceQuotaOutput, error) {
	value := UnknownServiceQuota

	if *input.ServiceCode == servicequotas.RDSServiceCode {
		switch *input.QuotaCode {
		case servicequotas.DBinstancesQuotacode:
			value = DBinstancesQuota
		case servicequotas.TotalStorageQuotaCode:
			value = TotalStorage
		case servicequotas.ManualDBInstanceSnapshotsQuotaCode:
			value = ManualDBInstanceSnapshots
		}
	}

	quota := &aws_servicequotas_types.ServiceQuota{Value: &value}

	return &aws_servicequotas.GetServiceQuotaOutput{Quota: quota}, nil
}
