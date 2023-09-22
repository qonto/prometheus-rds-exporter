package servicequotas_test

import (
	"context"
	"testing"

	aws_servicequotas "github.com/aws/aws-sdk-go-v2/service/servicequotas"
	aws_servicequotas_types "github.com/aws/aws-sdk-go-v2/service/servicequotas/types"
	"github.com/qonto/prometheus-rds-exporter/internal/app/servicequotas"
	converter "github.com/qonto/prometheus-rds-exporter/internal/app/unit"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Defines expected values for the mock and tests
const (
	UnknownServiceQuota       = float64(42)
	DBinstancesQuota          = float64(10)
	totalStorage              = float64(10)
	manualDBInstanceSnapshots = float64(42)
)

// mockServiceQuotasClient defines a mock for the AWS service quotas
type mockServiceQuotasClient struct{}

func (m mockServiceQuotasClient) GetServiceQuota(context context.Context, input *aws_servicequotas.GetServiceQuotaInput, optFns ...func(*aws_servicequotas.Options)) (*aws_servicequotas.GetServiceQuotaOutput, error) {
	value := UnknownServiceQuota

	if *input.ServiceCode == servicequotas.RDSServiceCode {
		switch *input.QuotaCode {
		case servicequotas.DBinstancesQuotacode:
			value = DBinstancesQuota
		case servicequotas.TotalStorageQuotaCode:
			value = totalStorage
		case servicequotas.ManualDBInstanceSnapshotsQuotaCode:
			value = manualDBInstanceSnapshots
		}
	}

	quota := &aws_servicequotas_types.ServiceQuota{Value: &value}

	return &aws_servicequotas.GetServiceQuotaOutput{Quota: quota}, nil
}

func TestGetRDSQuotas(t *testing.T) {
	mockClient := mockServiceQuotasClient{}

	result, err := servicequotas.NewFetcher(mockClient).GetRDSQuotas()
	require.NoError(t, err, "GetRDSQuotas must succeed")
	assert.Equal(t, DBinstancesQuota, result.DBinstances, "DbInstance quota is incorrect")
	assert.Equal(t, float64(converter.GigaBytesToBytes(totalStorage)), result.TotalStorage, "Total storage quota is incorrect")
	assert.Equal(t, manualDBInstanceSnapshots, result.ManualDBInstanceSnapshots, "Manual db instance snapshot quota is incorrect")
}
