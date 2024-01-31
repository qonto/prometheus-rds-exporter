package servicequotas_test

import (
	"context"
	"testing"

	"github.com/qonto/prometheus-rds-exporter/internal/app/servicequotas"
	mock "github.com/qonto/prometheus-rds-exporter/internal/app/servicequotas/mock"
	converter "github.com/qonto/prometheus-rds-exporter/internal/app/unit"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetRDSQuotas(t *testing.T) {
	context := context.TODO()
	client := mock.ServiceQuotasClient{}

	result, err := servicequotas.NewFetcher(context, client).GetRDSQuotas()
	require.NoError(t, err, "GetRDSQuotas must succeed")
	assert.Equal(t, mock.DBinstancesQuota, result.DBinstances, "DbInstance quota is incorrect")
	assert.Equal(t, converter.GigaBytesToBytes(mock.TotalStorage), result.TotalStorage, "Total storage quota is incorrect")
	assert.Equal(t, mock.ManualDBInstanceSnapshots, result.ManualDBInstanceSnapshots, "Manual db instance snapshot quota is incorrect")
}
