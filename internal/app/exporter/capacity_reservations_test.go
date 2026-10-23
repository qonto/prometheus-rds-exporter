package exporter_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	aws_rds_types "github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/qonto/prometheus-rds-exporter/internal/app/exporter"
	"github.com/qonto/prometheus-rds-exporter/internal/infra/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	cloudwatch_mock "github.com/qonto/prometheus-rds-exporter/internal/app/cloudwatch/mock"
	ec2_mock "github.com/qonto/prometheus-rds-exporter/internal/app/ec2/mock"
	rds_mock "github.com/qonto/prometheus-rds-exporter/internal/app/rds/mock"
	servicequotas_mock "github.com/qonto/prometheus-rds-exporter/internal/app/servicequotas/mock"
)

const (
	testAWSAccountID = "123456789012"
	testAWSRegion    = "eu-west-3"
)

func newCapacityReservation(identifier string, state aws_rds_types.DBCapacityReservationState, engine string, instanceClass string, reservedPerAZ int32, availableByAZ map[string]int32) aws_rds_types.DBCapacityReservation {
	reservation := rds_mock.NewDBCapacityReservation(engine, instanceClass, reservedPerAZ, availableByAZ)
	reservation.DBCapacityReservationIdentifier = aws.String(identifier)
	reservation.DBCapacityReservationName = aws.String(identifier + "-name")
	reservation.DBCapacityReservationArn = aws.String("arn:aws:rds:eu-west-3:123456789012:db-capacity-reservation:" + identifier)
	reservation.State = state

	return reservation
}

func newCapacityReservationConfiguration(enabled bool) exporter.Configuration {
	return exporter.Configuration{
		CollectClusterMetrics:       true,
		CollectCapacityReservations: enabled,
	}
}

func TestCapacityReservationMetrics(t *testing.T) {
	usedPostgres := newCapacityReservation("dbcr-postgres", aws_rds_types.DBCapacityReservationStateActive, "postgres", "db.t4g.micro", 1, map[string]int32{"us-east-1a": 0, "us-east-1b": 0})
	usedPostgres.EndDate = aws.Time(time.Date(2026, time.November, 15, 0, 0, 0, 0, time.UTC))

	unusedAurora := newCapacityReservation("dbcr-aurora", aws_rds_types.DBCapacityReservationStateActive, "aurora-postgresql", "db.t4g.medium", 1, map[string]int32{"us-east-1a": 1, "us-east-1b": 1})

	scheduled := newCapacityReservation("dbcr-scheduled", aws_rds_types.DBCapacityReservationStateScheduled, "aurora-postgresql", "db.r7g.large", 16, map[string]int32{"us-east-1a": 16})
	scheduled.StartDate = aws.Time(time.Date(2026, time.December, 1, 0, 0, 0, 0, time.UTC))

	rdsClient := rds_mock.NewRDSClient().WithDBCapacityReservations(usedPostgres, unusedAurora, scheduled)
	logger, _ := logger.New(true, "text")
	collector := exporter.NewCollector(*logger, newCapacityReservationConfiguration(true), testAWSAccountID, testAWSRegion, rdsClient, ec2_mock.EC2Client{}, cloudwatch_mock.CloudwatchClient{}, servicequotas_mock.ServiceQuotasClient{}, nil)

	expected := `
# HELP rds_capacity_reservation_available_average Number of DB instances reserved by the active On-Demand Capacity Reservation in the availability zone and not used by any DB instance
# TYPE rds_capacity_reservation_available_average gauge
rds_capacity_reservation_available_average{availability_zone="us-east-1a",aws_account_id="123456789012",aws_region="eu-west-3",capacity_reservation_id="dbcr-aurora",engine="aurora-postgresql",instance_class="db.t4g.medium"} 1
rds_capacity_reservation_available_average{availability_zone="us-east-1a",aws_account_id="123456789012",aws_region="eu-west-3",capacity_reservation_id="dbcr-postgres",engine="postgres",instance_class="db.t4g.micro"} 0
rds_capacity_reservation_available_average{availability_zone="us-east-1b",aws_account_id="123456789012",aws_region="eu-west-3",capacity_reservation_id="dbcr-aurora",engine="aurora-postgresql",instance_class="db.t4g.medium"} 1
rds_capacity_reservation_available_average{availability_zone="us-east-1b",aws_account_id="123456789012",aws_region="eu-west-3",capacity_reservation_id="dbcr-postgres",engine="postgres",instance_class="db.t4g.micro"} 0
# HELP rds_capacity_reservation_end_timestamp_seconds Timestamp when the On-Demand Capacity Reservation expires
# TYPE rds_capacity_reservation_end_timestamp_seconds gauge
rds_capacity_reservation_end_timestamp_seconds{aws_account_id="123456789012",aws_region="eu-west-3",capacity_reservation_id="dbcr-postgres"} 1.7947008e+09
# HELP rds_capacity_reservation_info On-Demand Capacity Reservation information
# TYPE rds_capacity_reservation_info gauge
rds_capacity_reservation_info{arn="arn:aws:rds:eu-west-3:123456789012:db-capacity-reservation:dbcr-aurora",aws_account_id="123456789012",aws_region="eu-west-3",capacity_reservation_id="dbcr-aurora",capacity_reservation_name="dbcr-aurora-name",engine="aurora-postgresql",instance_class="db.t4g.medium",state="active"} 1
rds_capacity_reservation_info{arn="arn:aws:rds:eu-west-3:123456789012:db-capacity-reservation:dbcr-postgres",aws_account_id="123456789012",aws_region="eu-west-3",capacity_reservation_id="dbcr-postgres",capacity_reservation_name="dbcr-postgres-name",engine="postgres",instance_class="db.t4g.micro",state="active"} 1
rds_capacity_reservation_info{arn="arn:aws:rds:eu-west-3:123456789012:db-capacity-reservation:dbcr-scheduled",aws_account_id="123456789012",aws_region="eu-west-3",capacity_reservation_id="dbcr-scheduled",capacity_reservation_name="dbcr-scheduled-name",engine="aurora-postgresql",instance_class="db.r7g.large",state="scheduled"} 1
# HELP rds_capacity_reservation_reserved_average Number of DB instances reserved by the active On-Demand Capacity Reservation in the availability zone
# TYPE rds_capacity_reservation_reserved_average gauge
rds_capacity_reservation_reserved_average{availability_zone="us-east-1a",aws_account_id="123456789012",aws_region="eu-west-3",capacity_reservation_id="dbcr-aurora",engine="aurora-postgresql",instance_class="db.t4g.medium"} 1
rds_capacity_reservation_reserved_average{availability_zone="us-east-1a",aws_account_id="123456789012",aws_region="eu-west-3",capacity_reservation_id="dbcr-postgres",engine="postgres",instance_class="db.t4g.micro"} 1
rds_capacity_reservation_reserved_average{availability_zone="us-east-1b",aws_account_id="123456789012",aws_region="eu-west-3",capacity_reservation_id="dbcr-aurora",engine="aurora-postgresql",instance_class="db.t4g.medium"} 1
rds_capacity_reservation_reserved_average{availability_zone="us-east-1b",aws_account_id="123456789012",aws_region="eu-west-3",capacity_reservation_id="dbcr-postgres",engine="postgres",instance_class="db.t4g.micro"} 1
# HELP rds_capacity_reservation_start_timestamp_seconds Timestamp when the future-dated On-Demand Capacity Reservation becomes active
# TYPE rds_capacity_reservation_start_timestamp_seconds gauge
rds_capacity_reservation_start_timestamp_seconds{aws_account_id="123456789012",aws_region="eu-west-3",capacity_reservation_id="dbcr-scheduled"} 1.7960832e+09
`

	err := testutil.CollectAndCompare(collector, strings.NewReader(expected),
		"rds_capacity_reservation_available_average",
		"rds_capacity_reservation_end_timestamp_seconds",
		"rds_capacity_reservation_info",
		"rds_capacity_reservation_reserved_average",
		"rds_capacity_reservation_start_timestamp_seconds",
	)
	require.NoError(t, err, "capacity reservation metrics mismatch")

	counter := collector.GetStatistics()
	assert.InDelta(t, float64(0), counter.Errors, 0, "should not have any error")
	assert.InDelta(t, float64(1), counter.CapacityReservationAPICalls, 0, "should have 1 call to DescribeDBCapacityReservations")
}

func TestCapacityReservationMetricsDisabled(t *testing.T) {
	reservation := newCapacityReservation("dbcr-postgres", aws_rds_types.DBCapacityReservationStateActive, "postgres", "db.t4g.micro", 1, map[string]int32{"us-east-1a": 1})

	rdsClient := rds_mock.NewRDSClient().WithDBCapacityReservations(reservation)
	logger, _ := logger.New(true, "text")
	collector := exporter.NewCollector(*logger, newCapacityReservationConfiguration(false), testAWSAccountID, testAWSRegion, rdsClient, ec2_mock.EC2Client{}, cloudwatch_mock.CloudwatchClient{}, servicequotas_mock.ServiceQuotasClient{}, nil)

	count := testutil.CollectAndCount(collector,
		"rds_capacity_reservation_available_average",
		"rds_capacity_reservation_info",
		"rds_capacity_reservation_reserved_average",
	)

	assert.Equal(t, 0, count, "should not export capacity reservation metrics when collector is disabled")
	assert.InDelta(t, float64(0), collector.GetStatistics().CapacityReservationAPICalls, 0, "should not call DescribeDBCapacityReservations")
}

func TestCapacityReservationMetricsError(t *testing.T) {
	rdsClient := rds_mock.NewRDSClient().WithDBCapacityReservationsError(errors.New("AccessDenied"))
	logger, _ := logger.New(true, "text")
	collector := exporter.NewCollector(*logger, newCapacityReservationConfiguration(true), testAWSAccountID, testAWSRegion, rdsClient, ec2_mock.EC2Client{}, cloudwatch_mock.CloudwatchClient{}, servicequotas_mock.ServiceQuotasClient{}, nil)

	expected := `
# HELP up Was the last scrape of RDS successful
# TYPE up counter
up 1
`
	err := testutil.CollectAndCompare(collector, strings.NewReader(expected), "up")
	require.NoError(t, err, "exporter must stay up when capacity reservations can't be fetched")

	assert.InDelta(t, float64(1), collector.GetStatistics().Errors, 0, "should count the error")
	assert.Equal(t, 0, testutil.CollectAndCount(collector, "rds_capacity_reservation_info"), "should not export capacity reservation metrics on error")
}

func TestInstanceCapacityReservationLabel(t *testing.T) {
	instance := rds_mock.NewRdsInstance()
	instance.DBCapacityReservationIdentifier = aws.String("dbcr-postgres")

	rdsClient := rds_mock.NewRDSClient().WithDBInstances(*instance)
	logger, _ := logger.New(true, "text")
	collector := exporter.NewCollector(*logger, newCapacityReservationConfiguration(false), testAWSAccountID, testAWSRegion, rdsClient, ec2_mock.EC2Client{}, cloudwatch_mock.CloudwatchClient{}, servicequotas_mock.ServiceQuotasClient{}, nil)

	registry := prometheus.NewPedanticRegistry()
	require.NoError(t, registry.Register(collector))

	families, err := registry.Gather()
	require.NoError(t, err)

	var labels map[string]string

	for _, family := range families {
		if family.GetName() != "rds_instance_info" {
			continue
		}

		labels = make(map[string]string)
		for _, label := range family.GetMetric()[0].GetLabel() {
			labels[label.GetName()] = label.GetValue()
		}
	}

	require.NotNil(t, labels, "rds_instance_info must be exported")
	assert.Equal(t, "dbcr-postgres", labels["capacity_reservation_id"], "capacity_reservation_id label mismatch")
}
