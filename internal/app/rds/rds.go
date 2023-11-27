// Package rds implements methods to retrieve RDS information
package rds

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	aws_rds "github.com/aws/aws-sdk-go-v2/service/rds"
	aws_rds_types "github.com/aws/aws-sdk-go-v2/service/rds/types"
	converter "github.com/qonto/prometheus-rds-exporter/internal/app/unit"
)

type Configuration struct {
	CollectLogsSize     bool
	CollectMaintenances bool
}

type Metrics struct {
	Instances map[string]RdsInstanceMetrics
}

type Statistics struct {
	RdsAPICall float64
}

type RdsInstanceMetrics struct {
	Engine                           string
	EngineVersion                    string
	DBInstanceClass                  string
	DbiResourceID                    string
	StorageType                      string
	AllocatedStorage                 int64
	StorageThroughput                int64
	MaxAllocatedStorage              int64
	MaxIops                          int64
	LogFilesSize                     *int64
	PendingMaintenanceAction         string
	PendingModifiedValues            bool
	BackupRetentionPeriod            int32
	Status                           int
	DeletionProtection               bool
	PubliclyAccessible               bool
	PerformanceInsightsEnabled       bool
	MultiAZ                          bool
	IAMDatabaseAuthenticationEnabled bool
	Role                             string
	SourceDBInstanceIdentifier       string
	CACertificateIdentifier          string
	CertificateValidTill             time.Time
	Age                              *float64
}

const (
	InstanceStatusAvailable                int    = 1
	InstanceStatusBackingUp                int    = 2
	InstanceStatusStarting                 int    = 3
	InstanceStatusModifying                int    = 4
	InstanceStatusStopped                  int    = 0
	InstanceStatusUnknown                  int    = -1
	InstanceStatusStopping                 int    = -2
	InstanceStatusCreating                 int    = -3
	InstanceStatusDeleting                 int    = -4
	NoPendingMaintenanceOperation          string = "no"
	UnknownMaintenanceOperation            string = "unknown"
	UnscheduledPendingMaintenanceOperation string = "pending"
	AutoAppliedPendingMaintenanceOperation string = "auto-applied"
	ForcedPendingMaintenanceOperation      string = "forced"
	gp2IOPSMin                             int64  = 100
	gp2IOPSMax                             int64  = 16000
	gp2IOPSPerGB                           int64  = 3
	gp2StorageThroughputVolumeThreshold    int64  = 334
	gp2StorageThroughputSmallVolume        int64  = 128
	gp2StorageThroughputLargeVolume        int64  = 250
	io1HighIOPSThroughputThreshold         int64  = 64000
	io1HighIOPSThroughputValue             int64  = 1000
	io1LargeIOPSThroughputThreshold        int64  = 32000
	io1LargeIOPSThroughputValue            int64  = 16
	io1MediumIOPSThroughputThreshold       int64  = 2000
	io1MediumIOPSThroughputValue           int64  = 500
	io1DefaultIOPSThroughputValue          int64  = 256
	primaryRole                            string = "primary"
	replicaRole                            string = "replica"
)

var instanceStatuses = map[string]int{
	"available":  InstanceStatusAvailable,
	"backing-up": InstanceStatusBackingUp,
	"creating":   InstanceStatusCreating,
	"deleting":   InstanceStatusDeleting,
	"modifying":  InstanceStatusModifying,
	"starting":   InstanceStatusStarting,
	"stopped":    InstanceStatusStopped,
	"stopping":   InstanceStatusStopping,
	"unknown":    InstanceStatusUnknown,
}

type RDSClient interface {
	DescribeDBInstances(ctx context.Context, params *aws_rds.DescribeDBInstancesInput, optFns ...func(*aws_rds.Options)) (*aws_rds.DescribeDBInstancesOutput, error)
	DescribePendingMaintenanceActions(context.Context, *aws_rds.DescribePendingMaintenanceActionsInput, ...func(*aws_rds.Options)) (*aws_rds.DescribePendingMaintenanceActionsOutput, error)
	DescribeDBLogFiles(context.Context, *aws_rds.DescribeDBLogFilesInput, ...func(*aws_rds.Options)) (*aws_rds.DescribeDBLogFilesOutput, error)
}

func NewFetcher(client RDSClient, configuration Configuration) RDSFetcher {
	return RDSFetcher{
		client:        client,
		configuration: configuration,
	}
}

type RDSFetcher struct {
	client        RDSClient
	statistics    Statistics
	configuration Configuration
}

func (r *RDSFetcher) GetStatistics() Statistics {
	return r.statistics
}

func (r *RDSFetcher) getPendingMaintenances() (map[string]string, error) {
	instances := make(map[string]string)

	inputMaintenance := &aws_rds.DescribePendingMaintenanceActionsInput{}

	maintenances, err := r.client.DescribePendingMaintenanceActions(context.TODO(), inputMaintenance)
	if err != nil {
		return nil, fmt.Errorf("can't describe pending maintenance actions: %w", err)
	}
	r.statistics.RdsAPICall++

	if maintenances == nil {
		return nil, nil
	}

	for _, maintenance := range maintenances.PendingMaintenanceActions {
		var maintenanceMode string

		dbIdentifier := GetDBIdentifierFromARN(*maintenance.ResourceIdentifier)

		for _, action := range maintenance.PendingMaintenanceActionDetails {
			switch {
			case action.ForcedApplyDate != nil:
				maintenanceMode = ForcedPendingMaintenanceOperation
			case action.AutoAppliedAfterDate != nil && maintenanceMode != ForcedPendingMaintenanceOperation:
				maintenanceMode = AutoAppliedPendingMaintenanceOperation
			default:
				maintenanceMode = UnscheduledPendingMaintenanceOperation
			}
		}

		instances[dbIdentifier] = maintenanceMode
	}

	return instances, nil
}

func (r *RDSFetcher) GetInstancesMetrics() (Metrics, error) {
	metrics := make(map[string]RdsInstanceMetrics)

	var err error

	var instanceMaintenances map[string]string

	if r.configuration.CollectMaintenances {
		instanceMaintenances, err = r.getPendingMaintenances()
		if err != nil {
			return Metrics{}, fmt.Errorf("can't get RDS maintenances: %w", err)
		}
	}

	input := &aws_rds.DescribeDBInstancesInput{}

	paginator := aws_rds.NewDescribeDBInstancesPaginator(r.client, input)
	for paginator.HasMorePages() {
		r.statistics.RdsAPICall++

		output, err := paginator.NextPage(context.TODO())
		if err != nil {
			return Metrics{}, fmt.Errorf("can't get instances: %w", err)
		}

		for _, dbInstance := range output.DBInstances {
			dbIdentifier := dbInstance.DBInstanceIdentifier

			instanceMetrics, err := r.computeInstanceMetrics(dbInstance, instanceMaintenances)
			if err != nil {
				return Metrics{}, fmt.Errorf("can't compute instance metrics for %s: %w", *dbIdentifier, err)
			}

			metrics[*dbIdentifier] = instanceMetrics
		}
	}

	return Metrics{Instances: metrics}, nil
}

// computeInstanceMetrics returns metrics about the specified instance
func (r *RDSFetcher) computeInstanceMetrics(dbInstance aws_rds_types.DBInstance, instanceMaintenances map[string]string) (RdsInstanceMetrics, error) {
	dbIdentifier := dbInstance.DBInstanceIdentifier

	var iops int64
	if dbInstance.Iops != nil {
		iops = int64(*dbInstance.Iops)
	}

	var throughput int64
	if dbInstance.StorageThroughput != nil {
		throughput = int64(*dbInstance.StorageThroughput)
	}

	iops, storageThroughput := getStorageMetrics(*dbInstance.StorageType, int64(dbInstance.AllocatedStorage), iops, throughput)

	var maxAllocatedStorage int32 = 0
	if dbInstance.MaxAllocatedStorage != nil {
		maxAllocatedStorage = *dbInstance.MaxAllocatedStorage
	}

	pendingModifiedValues := false

	// PendingModifiedValues reports only instance changes
	if dbInstance.PendingModifiedValues != nil && !reflect.DeepEqual(dbInstance.PendingModifiedValues, &aws_rds_types.PendingModifiedValues{}) {
		pendingModifiedValues = true
	}

	// Report pending modified values if at lease one parameter group is not applied
	for _, parameterGroup := range dbInstance.DBParameterGroups {
		if *parameterGroup.ParameterApplyStatus != "in-sync" {
			pendingModifiedValues = true
		}
	}

	pendingMaintenanceAction := NoPendingMaintenanceOperation
	if !r.configuration.CollectMaintenances {
		pendingMaintenanceAction = UnknownMaintenanceOperation
	} else {
		if maintenanceMode, isFound := instanceMaintenances[*dbIdentifier]; isFound {
			pendingMaintenanceAction = maintenanceMode
		}
	}

	var logFilesSize *int64

	if r.configuration.CollectLogsSize {
		var err error
		logFilesSize, err = r.getLogFilesSize(*dbIdentifier)

		if err != nil {
			return RdsInstanceMetrics{}, fmt.Errorf("can't get log files size for %d: %w", dbIdentifier, err)
		}
	}

	role, sourceDBInstanceIdentifier := getRoleInCluster(&dbInstance)

	var age *float64

	if dbInstance.InstanceCreateTime != nil {
		diff := time.Since(*dbInstance.InstanceCreateTime).Seconds()
		age = &diff
	}

	metrics := RdsInstanceMetrics{
		AllocatedStorage:           converter.GigaBytesToBytes(dbInstance.AllocatedStorage),
		BackupRetentionPeriod:      converter.DaystoSeconds(dbInstance.BackupRetentionPeriod),
		DBInstanceClass:            *dbInstance.DBInstanceClass,
		DbiResourceID:              *dbInstance.DbiResourceId,
		DeletionProtection:         dbInstance.DeletionProtection,
		Engine:                     *dbInstance.Engine,
		EngineVersion:              *dbInstance.EngineVersion,
		LogFilesSize:               logFilesSize,
		MaxAllocatedStorage:        converter.GigaBytesToBytes(maxAllocatedStorage),
		MaxIops:                    iops,
		MultiAZ:                    dbInstance.MultiAZ,
		PendingMaintenanceAction:   pendingMaintenanceAction,
		PendingModifiedValues:      pendingModifiedValues,
		PerformanceInsightsEnabled: aws.ToBool(dbInstance.PerformanceInsightsEnabled),
		PubliclyAccessible:         dbInstance.PubliclyAccessible,
		Role:                       role,
		SourceDBInstanceIdentifier: sourceDBInstanceIdentifier,
		Status:                     GetDBInstanceStatusCode(*dbInstance.DBInstanceStatus),
		StorageThroughput:          converter.MegaBytesToBytes(storageThroughput),
		StorageType:                *dbInstance.StorageType,
		CACertificateIdentifier:    *dbInstance.CACertificateIdentifier,
		CertificateValidTill:       *dbInstance.CertificateDetails.ValidTill,
		Age:                        age,
	}

	return metrics, nil
}

// getLogFilesSize returns the size of all logs on the specified instance
func (r *RDSFetcher) getLogFilesSize(dbidentifier string) (*int64, error) {
	var filesSize *int64

	input := &aws_rds.DescribeDBLogFilesInput{DBInstanceIdentifier: &dbidentifier}

	r.statistics.RdsAPICall++

	result, err := r.client.DescribeDBLogFiles(context.TODO(), input)
	if err != nil {
		var notFoundError *aws_rds_types.DBInstanceNotFoundFault
		if errors.As(err, &notFoundError) { // Replica in "creating" status may return notFoundError exception
			return filesSize, nil
		}

		return filesSize, fmt.Errorf("can't describe db logs files for %s: %w", dbidentifier, err)
	}

	if result != nil && len(result.DescribeDBLogFiles) > 0 {
		if filesSize == nil {
			filesSize = new(int64)
		}

		for _, file := range result.DescribeDBLogFiles {
			*filesSize += file.Size
		}
	}

	return filesSize, nil
}
