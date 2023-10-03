package servicequotas_test

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	aws_servicequotas "github.com/aws/aws-sdk-go-v2/service/servicequotas"
	aws_servicequotas_types "github.com/aws/aws-sdk-go-v2/service/servicequotas/types"
	"github.com/qonto/prometheus-rds-exporter/internal/app/servicequotas"
	"github.com/qonto/prometheus-rds-exporter/internal/app/servicequotas/mocks"
	converter "github.com/qonto/prometheus-rds-exporter/internal/app/unit"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Defines expected values for the mock and tests
const (
	UnknownServiceQuota       = float64(42)
	DBinstancesQuota          = float64(10)
	totalStorage              = float64(24)
	manualDBInstanceSnapshots = float64(100)
)

func TestGetRDSQuotas(t *testing.T) {
	mockClient := mocks.NewServiceQuotasClient(t)

	dbInstanceResponse := aws_servicequotas.GetServiceQuotaOutput{Quota: &aws_servicequotas_types.ServiceQuota{Value: aws.Float64(DBinstancesQuota)}}
	mockClient.On("GetServiceQuota", context.TODO(), &aws_servicequotas.GetServiceQuotaInput{QuotaCode: aws.String(servicequotas.DBinstancesQuotacode), ServiceCode: aws.String("rds")}).Return(&dbInstanceResponse, nil).Once()

	totalStorageResponse := aws_servicequotas.GetServiceQuotaOutput{Quota: &aws_servicequotas_types.ServiceQuota{Value: aws.Float64(totalStorage)}}
	mockClient.On("GetServiceQuota", context.TODO(), &aws_servicequotas.GetServiceQuotaInput{QuotaCode: aws.String(servicequotas.TotalStorageQuotaCode), ServiceCode: aws.String("rds")}).Return(&totalStorageResponse, nil).Once()

	manualSnapshotResponse := aws_servicequotas.GetServiceQuotaOutput{Quota: &aws_servicequotas_types.ServiceQuota{Value: aws.Float64(manualDBInstanceSnapshots)}}
	mockClient.On("GetServiceQuota", context.TODO(), &aws_servicequotas.GetServiceQuotaInput{QuotaCode: aws.String(servicequotas.ManualDBInstanceSnapshotsQuotaCode), ServiceCode: aws.String("rds")}).Return(&manualSnapshotResponse, nil).Once()

	result, err := servicequotas.NewFetcher(mockClient).GetRDSQuotas()
	require.NoError(t, err, "GetRDSQuotas must succeed")
	assert.Equal(t, DBinstancesQuota, result.DBinstances, "DbInstance quota is incorrect")
	assert.Equal(t, float64(converter.GigaBytesToBytes(totalStorage)), result.TotalStorage, "Total storage quota is incorrect")
	assert.Equal(t, manualDBInstanceSnapshots, result.ManualDBInstanceSnapshots, "Manual db instance snapshot quota is incorrect")
}
