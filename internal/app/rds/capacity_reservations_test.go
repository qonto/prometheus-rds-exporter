package rds_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	aws_rds_types "github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/qonto/prometheus-rds-exporter/internal/app/rds"
	mock "github.com/qonto/prometheus-rds-exporter/internal/app/rds/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func capacityKey(reservation aws_rds_types.DBCapacityReservation, az string) rds.CapacityReservationKey {
	return rds.CapacityReservationKey{
		CapacityReservationID: *reservation.DBCapacityReservationIdentifier,
		Engine:                *reservation.Engine,
		DBInstanceClass:       *reservation.DBInstanceClass,
		AvailabilityZone:      az,
	}
}

func TestGetCapacityReservations(t *testing.T) {
	unusedPostgres := mock.NewDBCapacityReservation("postgres", "db.t4g.micro", 1, map[string]int32{"us-east-1a": 1, "us-east-1b": 1})
	usedPostgres := mock.NewDBCapacityReservation("postgres", "db.t4g.micro", 1, map[string]int32{"us-east-1a": 0, "us-east-1b": 0})
	unusedAurora := mock.NewDBCapacityReservation("aurora-postgresql", "db.t4g.medium", 1, map[string]int32{"us-east-1a": 1, "us-east-1b": 1})
	partiallyUsed := mock.NewDBCapacityReservation("postgres", "db.t4g.micro", 3, map[string]int32{"us-east-1a": 3, "us-east-1b": 2})

	testCases := []struct {
		name         string
		reservations []aws_rds_types.DBCapacityReservation
		expected     map[rds.CapacityReservationKey]rds.CapacityUsage
	}{
		{
			name:         "unused reservations",
			reservations: []aws_rds_types.DBCapacityReservation{unusedPostgres, unusedAurora},
			expected: map[rds.CapacityReservationKey]rds.CapacityUsage{
				capacityKey(unusedPostgres, "us-east-1a"): {Reserved: 1, Available: 1},
				capacityKey(unusedPostgres, "us-east-1b"): {Reserved: 1, Available: 1},
				capacityKey(unusedAurora, "us-east-1a"):   {Reserved: 1, Available: 1},
				capacityKey(unusedAurora, "us-east-1b"):   {Reserved: 1, Available: 1},
			},
		},
		{
			name:         "postgres reservation used by 2 instances",
			reservations: []aws_rds_types.DBCapacityReservation{usedPostgres, unusedAurora},
			expected: map[rds.CapacityReservationKey]rds.CapacityUsage{
				capacityKey(usedPostgres, "us-east-1a"): {Reserved: 1, Available: 0},
				capacityKey(usedPostgres, "us-east-1b"): {Reserved: 1, Available: 0},
				capacityKey(unusedAurora, "us-east-1a"): {Reserved: 1, Available: 1},
				capacityKey(unusedAurora, "us-east-1b"): {Reserved: 1, Available: 1},
			},
		},
		{
			name:         "reservations with same engine, instance class and availability zone are kept separate",
			reservations: []aws_rds_types.DBCapacityReservation{usedPostgres, partiallyUsed},
			expected: map[rds.CapacityReservationKey]rds.CapacityUsage{
				capacityKey(usedPostgres, "us-east-1a"):  {Reserved: 1, Available: 0},
				capacityKey(usedPostgres, "us-east-1b"):  {Reserved: 1, Available: 0},
				capacityKey(partiallyUsed, "us-east-1a"): {Reserved: 3, Available: 3},
				capacityKey(partiallyUsed, "us-east-1b"): {Reserved: 3, Available: 2},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := mock.NewRDSClient().WithDBCapacityReservations(tc.reservations...)

			fetcher := rds.NewCapacityReservationFetcher(context.TODO(), client, *slog.New(slog.DiscardHandler))
			metrics, err := fetcher.GetCapacityReservations()

			require.NoError(t, err, "GetCapacityReservations must succeed")
			assert.Equal(t, tc.expected, metrics.Capacity, "Capacity mismatch")
			assert.Len(t, metrics.Reservations, len(tc.reservations), "All reservations must be returned")
			assert.InDelta(t, float64(1), fetcher.GetStatistics().RdsAPICall, 0, "API call count mismatch")
		})
	}
}

func TestGetCapacityReservationsDetails(t *testing.T) {
	startDate := time.Date(2026, time.November, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, time.November, 15, 0, 0, 0, 0, time.UTC)

	active := mock.NewDBCapacityReservation("aurora-postgresql", "db.r7g.large", 2, map[string]int32{"us-east-1a": 1})
	active.EndDate = aws.Time(endDate)

	scheduled := mock.NewDBCapacityReservation("aurora-postgresql", "db.r7g.large", 16, map[string]int32{"us-east-1a": 16})
	scheduled.State = aws_rds_types.DBCapacityReservationStateScheduled
	scheduled.StartDate = aws.Time(startDate)

	client := mock.NewRDSClient().WithDBCapacityReservations(active, scheduled)

	fetcher := rds.NewCapacityReservationFetcher(context.TODO(), client, *slog.New(slog.DiscardHandler))
	metrics, err := fetcher.GetCapacityReservations()

	require.NoError(t, err, "GetCapacityReservations must succeed")

	activeMetrics := metrics.Reservations[*active.DBCapacityReservationIdentifier]
	assert.Equal(t, *active.DBCapacityReservationName, activeMetrics.Name, "Name mismatch")
	assert.Equal(t, *active.DBCapacityReservationArn, activeMetrics.Arn, "ARN mismatch")
	assert.Equal(t, "aurora-postgresql", activeMetrics.Engine, "Engine mismatch")
	assert.Equal(t, "db.r7g.large", activeMetrics.DBInstanceClass, "Instance class mismatch")
	assert.Equal(t, "active", activeMetrics.State, "State mismatch")
	assert.Nil(t, activeMetrics.StartDate, "Immediate reservation has no start date")
	assert.Equal(t, endDate, *activeMetrics.EndDate, "End date mismatch")

	scheduledMetrics := metrics.Reservations[*scheduled.DBCapacityReservationIdentifier]
	assert.Equal(t, "scheduled", scheduledMetrics.State, "State mismatch")
	assert.Equal(t, startDate, *scheduledMetrics.StartDate, "Start date mismatch")
	assert.Nil(t, scheduledMetrics.EndDate, "Reservation without end date")

	// Non-active reservations don't hold capacity
	assert.Equal(t, map[rds.CapacityReservationKey]rds.CapacityUsage{capacityKey(active, "us-east-1a"): {Reserved: 2, Available: 1}}, metrics.Capacity, "Only active reservations must hold capacity")
}

func TestGetCapacityReservationsPagination(t *testing.T) {
	first := mock.NewDBCapacityReservation("postgres", "db.t4g.micro", 1, map[string]int32{"us-east-1a": 1})
	second := mock.NewDBCapacityReservation("postgres", "db.t4g.micro", 1, map[string]int32{"us-east-1a": 0})

	client := mock.NewRDSClient().WithDBCapacityReservationsPages(
		[]aws_rds_types.DBCapacityReservation{first},
		[]aws_rds_types.DBCapacityReservation{second},
	)

	fetcher := rds.NewCapacityReservationFetcher(context.TODO(), client, *slog.New(slog.DiscardHandler))
	metrics, err := fetcher.GetCapacityReservations()

	require.NoError(t, err, "GetCapacityReservations must succeed")
	assert.Len(t, metrics.Reservations, 2, "Reservations of all pages must be returned")

	expected := map[rds.CapacityReservationKey]rds.CapacityUsage{
		capacityKey(first, "us-east-1a"):  {Reserved: 1, Available: 1},
		capacityKey(second, "us-east-1a"): {Reserved: 1, Available: 0},
	}
	assert.Equal(t, expected, metrics.Capacity, "Capacity of all pages must be returned")
	assert.InDelta(t, float64(2), fetcher.GetStatistics().RdsAPICall, 0, "One API call per page")
}

func TestGetCapacityReservationsError(t *testing.T) {
	client := mock.NewRDSClient().WithDBCapacityReservationsError(errors.New("AccessDenied"))

	fetcher := rds.NewCapacityReservationFetcher(context.TODO(), client, *slog.New(slog.DiscardHandler))
	_, err := fetcher.GetCapacityReservations()

	require.Error(t, err, "GetCapacityReservations must return API errors")
	assert.ErrorContains(t, err, "AccessDenied")
}
