package rds

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	aws_rds "github.com/aws/aws-sdk-go-v2/service/rds"
	aws_rds_types "github.com/aws/aws-sdk-go-v2/service/rds/types"
	"go.opentelemetry.io/otel/codes"
)

// capacityReservationsMaxRecords is the maximum page size allowed by DescribeDBCapacityReservations
const capacityReservationsMaxRecords int32 = 100

type CapacityReservationClient interface {
	DescribeDBCapacityReservations(ctx context.Context, params *aws_rds.DescribeDBCapacityReservationsInput, optFns ...func(*aws_rds.Options)) (*aws_rds.DescribeDBCapacityReservationsOutput, error)
}

type CapacityReservationMetrics struct {
	// Reservations indexed by DBCapacityReservationIdentifier, in any state
	Reservations map[string]CapacityReservation

	// Capacity of active reservations, per reservation and availability zone
	Capacity map[CapacityReservationKey]CapacityUsage
}

type CapacityReservation struct {
	Name            string
	Arn             string
	Engine          string
	DBInstanceClass string
	State           string

	// Date when the reservation becomes active (future-dated reservations only)
	StartDate *time.Time

	// Date when the reservation expires (nil when the reservation doesn't expire)
	EndDate *time.Time
}

type CapacityReservationKey struct {
	CapacityReservationID string
	Engine                string
	DBInstanceClass       string
	AvailabilityZone      string
}

type CapacityUsage struct {
	// Number of DB instances reserved
	Reserved float64

	// Number of reserved DB instances not used by any DB instance
	Available float64
}

type CapacityReservationStatistics struct {
	RdsAPICall float64
}

type CapacityReservationFetcher struct {
	ctx        context.Context
	client     CapacityReservationClient
	statistics CapacityReservationStatistics
	logger     slog.Logger
}

func NewCapacityReservationFetcher(ctx context.Context, client CapacityReservationClient, logger slog.Logger) *CapacityReservationFetcher {
	return &CapacityReservationFetcher{
		ctx:    ctx,
		client: client,
		logger: logger,
	}
}

func (c *CapacityReservationFetcher) GetStatistics() CapacityReservationStatistics {
	return c.statistics
}

// GetCapacityReservations returns all On-Demand Capacity Reservations and the aggregated capacity of the active ones
func (c *CapacityReservationFetcher) GetCapacityReservations() (CapacityReservationMetrics, error) {
	ctx, span := tracer.Start(c.ctx, "collect-capacity-reservations")
	defer span.End()

	metrics := CapacityReservationMetrics{
		Reservations: make(map[string]CapacityReservation),
		Capacity:     make(map[CapacityReservationKey]CapacityUsage),
	}

	input := &aws_rds.DescribeDBCapacityReservationsInput{MaxRecords: aws.Int32(capacityReservationsMaxRecords)}

	paginator := aws_rds.NewDescribeDBCapacityReservationsPaginator(c.client, input)
	for paginator.HasMorePages() {
		c.statistics.RdsAPICall++

		output, err := paginator.NextPage(ctx)
		if err != nil {
			span.SetStatus(codes.Error, "can't describe capacity reservations")
			span.RecordError(err)

			return CapacityReservationMetrics{}, fmt.Errorf("can't describe capacity reservations: %w", err)
		}

		for _, reservation := range output.DBCapacityReservations {
			addCapacityReservation(&metrics, reservation)
		}
	}

	c.logger.Debug("capacity reservations fetched", "count", len(metrics.Reservations))

	span.SetStatus(codes.Ok, "capacity reservations fetched")

	return metrics, nil
}

func addCapacityReservation(metrics *CapacityReservationMetrics, reservation aws_rds_types.DBCapacityReservation) {
	identifier := aws.ToString(reservation.DBCapacityReservationIdentifier)
	engine := aws.ToString(reservation.Engine)
	instanceClass := aws.ToString(reservation.DBInstanceClass)

	metrics.Reservations[identifier] = CapacityReservation{
		Name:            aws.ToString(reservation.DBCapacityReservationName),
		Arn:             aws.ToString(reservation.DBCapacityReservationArn),
		Engine:          engine,
		DBInstanceClass: instanceClass,
		State:           string(reservation.State),
		StartDate:       reservation.StartDate,
		EndDate:         reservation.EndDate,
	}

	// Only active reservations hold capacity and are billed
	if reservation.State != aws_rds_types.DBCapacityReservationStateActive {
		return
	}

	reservedPerAZ := float64(aws.ToInt32(reservation.ReservedCapacityPerAvailabilityZone))

	for _, az := range reservation.AvailableCapacityPerAvailabilityZoneList {
		key := CapacityReservationKey{
			CapacityReservationID: identifier,
			Engine:                engine,
			DBInstanceClass:       instanceClass,
			AvailabilityZone:      aws.ToString(az.AvailabilityZone),
		}

		metrics.Capacity[key] = CapacityUsage{
			Reserved:  reservedPerAZ,
			Available: float64(aws.ToInt32(az.AvailableCapacity)),
		}
	}
}
