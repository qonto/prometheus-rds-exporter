package servicequotas_test

import (
	"context"
	"log/slog"
	"testing"

	aws_servicequotas "github.com/aws/aws-sdk-go-v2/service/servicequotas"
	"github.com/aws/aws-sdk-go-v2/service/servicequotas/types"
	"github.com/qonto/prometheus-rds-exporter/internal/app/servicequotas"
	mock "github.com/qonto/prometheus-rds-exporter/internal/app/servicequotas/mock"
	converter "github.com/qonto/prometheus-rds-exporter/internal/app/unit"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetRDSQuotas(t *testing.T) {
	logger := slog.Default()
	t.Run("GetRDSQuotasHappyPath", func(t *testing.T) {
		context := context.TODO()
		client := mock.ServiceQuotasClient{}

		result, err := servicequotas.NewFetcher(context, client, *logger).GetRDSQuotas()
		require.NoError(t, err, "GetRDSQuotas must succeed")
		assert.Equal(t, mock.DBinstancesQuota, result.DBinstances, "DbInstance quota is incorrect")
		assert.Equal(t, converter.GigaBytesToBytes(mock.TotalStorage), result.TotalStorage, "Total storage quota is incorrect")
		assert.Equal(t, mock.ManualDBInstanceSnapshots, result.ManualDBInstanceSnapshots, "Manual db instance snapshot quota is incorrect")
	})

	t.Run("GetRDSQuotasErrorFetchingQuotaNilErrorMessage", func(t *testing.T) {
		context := context.TODO()
		client := mock.ServiceQuotasClientQuotaError{
			ExpectedErrorQotaCode: servicequotas.DBinstancesQuotacode,
			ExpectedErrorQuotaOutput: &aws_servicequotas.GetServiceQuotaOutput{
				Quota: &types.ServiceQuota{
					ErrorReason: &types.ErrorReason{
						ErrorCode: types.ErrorCodeServiceQuotaNotAvailableError,
					},
				},
			},
		}

		_, err := servicequotas.NewFetcher(context, client, *logger).GetRDSQuotas()
		require.EqualError(t, err, "can't fetch DBinstance quota: AWS return error for this quota")
	})
}
