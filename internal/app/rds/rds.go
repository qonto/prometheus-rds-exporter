// Package rds implements methods to retrieve RDS information
package rds

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	aws_rds "github.com/aws/aws-sdk-go-v2/service/rds"
	aws_rds_types "github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi"
	tag_types "github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi/types"
	converter "github.com/qonto/prometheus-rds-exporter/internal/app/unit"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

type Configuration struct {
	CollectLogsSize     bool
	CollectMaintenances bool
	TagSelections       map[string][]string
}

type Metrics struct {
	Instances map[string]RdsInstanceMetrics
	Clusters  map[string]ClusterMetrics
}

type Statistics struct {
	RdsAPICall float64
	TagAPICall float64
}

type ClusterMetrics struct {
	// Seconds since cluster creation date.
	Age float64

	// The Amazon Resource Name (ARN) for the DB cluster.
	Arn string

	// The database engine used for this DB cluster.
	Engine string

	EngineVersion string

	// For all database engines except Amazon Aurora, AllocatedStorage specifies the
	// allocated storage size in gibibytes (GiB). For Aurora, AllocatedStorage always
	// returns 1, because Aurora DB cluster storage size isn't fixed, but instead
	// automatically adjusts as needed.
	AllocatedStorage int64

	// The user-supplied identifier for the DB cluster. This identifier is the unique
	// key that identifies a DB cluster.
	DBClusterIdentifier string

	// The Amazon Web Services Region-unique, immutable identifier for the DB cluster.
	// This identifier is found in Amazon Web Services CloudTrail log entries whenever
	// the KMS key for the DB cluster is accessed.
	DbClusterResourceId string

	// Members
	Members map[string]DBRole

	// dbidentifier of the write node
	WriterDBInstanceIdentifier string

	// AWS tags on the cluster.
	Tags map[string]string
}

type RdsInstanceMetrics struct {
	// Seconds since instance creation date.
	Age *float64

	// The Amazon Resource Name (ARN) for the DB instance.
	Arn string

	// The version of the database engine.
	AllocatedStorage int64

	// The number of days for which automatic DB snapshots are retained.
	BackupRetentionPeriod int32

	// The identifier of the CA certificate for this DB instance.
	CACertificateIdentifier string

	// Certificate expiration date
	CertificateValidTill *time.Time

	// The name of the compute and memory capacity class of the DB instance.
	DBInstanceClass string

	// The Amazon Web Services Region-unique, immutable identifier for the DB
	DbiResourceID string

	// Indicates whether the DB instance has deletion protection enabled. The database
	DeletionProtection bool

	// The database engine used for this instance.
	Engine string

	// The version of the database engine.
	EngineVersion string

	// Indicates whether mapping of Amazon Web Services Identity and Access Management
	// (IAM) accounts to database accounts is enabled for the DB instance.
	IAMDatabaseAuthenticationEnabled bool

	// Total amount of log files (GiB)
	LogFilesSize *int64

	// The upper limit in gibibytes (GiB) to which Amazon RDS can automatically scale
	MaxAllocatedStorage int64

	// Maximum provisioned IOPS per GiB for a DB instance.
	MaxIops int64

	// Indicates whether the Single-AZ DB instance will change to a Multi-AZ deployment.
	MultiAZ bool

	// Pending maintenance action
	PendingMaintenanceAction string

	// Define if instance is pending for modification
	PendingModifiedValues bool

	// Indicates whether Performance Insights is enabled for the DB cluster.
	PerformanceInsightsEnabled bool

	// Indicates whether the DB instance is publicly accessible.
	PubliclyAccessible bool

	// Role of Instance primary or replica
	Role DBRole

	// If db instance is a replica, specify the identifier of the source
	SourceDBInstanceIdentifier string

	// Code representing instance status
	Status int

	// The storage throughput for the DB instance.
	StorageThroughput int64

	// The storage type associated with the DB instance.
	StorageType string

	// AWS tags on the cluster.
	Tags map[string]string
}

// DBRole defines the type for database instance roles such as primary, replica, writer or reader.
type DBRole string

func (r DBRole) String() string {
	return string(r)
}

const (
	InstanceStatusAvailable                     int     = 1
	InstanceStatusBackingUp                     int     = 2
	InstanceStatusStarting                      int     = 3
	InstanceStatusModifying                     int     = 4
	InstanceStatusConfiguringEnhancedMonitoring int     = 5
	InstanceStatusStorageInitialization         int     = 10
	InstanceStatusStorageOptimization           int     = 11
	InstanceStatusRenaming                      int     = 20
	InstanceStatusStopped                       int     = 0
	InstanceStatusUnknown                       int     = -1
	InstanceStatusStopping                      int     = -2
	InstanceStatusCreating                      int     = -3
	InstanceStatusDeleting                      int     = -4
	InstanceStatusRebooting                     int     = -5
	InstanceStatusFailed                        int     = -6
	InstanceStatusStorageFull                   int     = -7
	InstanceStatusUpgrading                     int     = -8
	InstanceStatusMaintenance                   int     = -9
	InstanceStatusRestoreError                  int     = -10
	NoPendingMaintenanceOperation               string  = "no"
	UnknownMaintenanceOperation                 string  = "unknown"
	UnscheduledPendingMaintenanceOperation      string  = "pending"
	AutoAppliedPendingMaintenanceOperation      string  = "auto-applied"
	ForcedPendingMaintenanceOperation           string  = "forced"
	gp2IOPSMin                                  int64   = 100
	gp2IOPSMax                                  int64   = 16000
	gp2IOPSPerGB                                int64   = 3
	gp2StorageThroughputVolumeThreshold         int64   = 334
	gp2StorageThroughputSmallVolume             int64   = 128
	gp2StorageThroughputLargeVolume             int64   = 250
	io1HighIOPSThroughputThreshold              int64   = 64000
	io1HighIOPSThroughputValue                  int64   = 1000
	io1LargeIOPSThroughputThreshold             int64   = 32000
	io1LargeIOPSThroughputValue                 int64   = 16
	io1MediumIOPSThroughputThreshold            int64   = 2000
	io1MediumIOPSThroughputValue                int64   = 500
	io1DefaultIOPSThroughputValue               int64   = 256
	io2StorageMinThroughput                     int64   = 256  // 1000 IOPS * 0.256 MiB/s per provisioned IOPS
	io2StorageMaxThroughput                     int64   = 4000 // AWS EBS limit
	io2StorageThroughputPerIOPS                 float64 = 0.256
	RolePrimary                                 DBRole  = "primary"
	RoleReplica                                 DBRole  = "replica"
	RoleWriter                                  DBRole  = "writer"
	RoleReader                                  DBRole  = "reader"
)

var tracer = otel.Tracer("github/qonto/prometheus-rds-exporter/internal/app/rds")

var instanceStatuses = map[string]int{ // retrieved from https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/accessing-monitoring.html
	"available":                       InstanceStatusAvailable,
	"backing-up":                      InstanceStatusBackingUp,
	"configuring-enhanced-monitoring": InstanceStatusConfiguringEnhancedMonitoring,
	"creating":                        InstanceStatusCreating,
	"deleting":                        InstanceStatusDeleting,
	"failed":                          InstanceStatusFailed,
	"maintenance":                     InstanceStatusMaintenance,
	"modifying":                       InstanceStatusModifying,
	"rebooting":                       InstanceStatusRebooting,
	"renaming":                        InstanceStatusRenaming,
	"restore-error":                   InstanceStatusRestoreError,
	"starting":                        InstanceStatusStarting,
	"stopped":                         InstanceStatusStopped,
	"storage-full":                    InstanceStatusStorageFull,
	"storage-initialization":          InstanceStatusStorageInitialization,
	"storage-optimization":            InstanceStatusStorageOptimization,
	"stopping":                        InstanceStatusStopping,
	"unknown":                         InstanceStatusUnknown,
	"upgrading":                       InstanceStatusUpgrading,
}

type RDSClient interface {
	DescribeDBInstances(ctx context.Context, params *aws_rds.DescribeDBInstancesInput, optFns ...func(*aws_rds.Options)) (*aws_rds.DescribeDBInstancesOutput, error)
	DescribeDBClusters(ctx context.Context, params *aws_rds.DescribeDBClustersInput, optFns ...func(*aws_rds.Options)) (*aws_rds.DescribeDBClustersOutput, error)
	DescribePendingMaintenanceActions(context.Context, *aws_rds.DescribePendingMaintenanceActionsInput, ...func(*aws_rds.Options)) (*aws_rds.DescribePendingMaintenanceActionsOutput, error)
	DescribeDBLogFiles(context.Context, *aws_rds.DescribeDBLogFilesInput, ...func(*aws_rds.Options)) (*aws_rds.DescribeDBLogFilesOutput, error)
}

func NewFetcher(ctx context.Context, client RDSClient, tagClient resourcegroupstaggingapi.GetResourcesAPIClient, logger slog.Logger, configuration Configuration) RDSFetcher {
	return RDSFetcher{
		ctx:           ctx,
		client:        client,
		tagClient:     tagClient,
		logger:        logger,
		configuration: configuration,
	}
}

type RDSFetcher struct {
	ctx           context.Context
	client        RDSClient
	statistics    Statistics
	configuration Configuration
	tagClient     resourcegroupstaggingapi.GetResourcesAPIClient
	logger        slog.Logger
}

func (r *RDSFetcher) GetStatistics() Statistics {
	return r.statistics
}

func (r *RDSFetcher) getPendingMaintenances(ctx context.Context) (map[string]string, error) {
	_, span := tracer.Start(ctx, "collect-pending-maintenances")
	defer span.End()

	instances := make(map[string]string)

	inputMaintenance := &aws_rds.DescribePendingMaintenanceActionsInput{}

	maintenances, err := r.client.DescribePendingMaintenanceActions(context.TODO(), inputMaintenance)
	if err != nil {
		span.SetStatus(codes.Error, "failed to get maintenances")
		span.RecordError(err)

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

	span.SetStatus(codes.Ok, "maintenances fetched")

	return instances, nil
}

func (r *RDSFetcher) getClusters(ctx context.Context, filters []aws_rds_types.Filter) (map[string]ClusterMetrics, error) {
	clusterMetrics := make(map[string]ClusterMetrics)

	inputCluster := &aws_rds.DescribeDBClustersInput{Filters: filters}

	paginatorCluster := aws_rds.NewDescribeDBClustersPaginator(r.client, inputCluster)
	for paginatorCluster.HasMorePages() {
		_, span := tracer.Start(ctx, "describe-rds-clusters")
		defer span.End()

		r.statistics.RdsAPICall++

		output, err := paginatorCluster.NextPage(context.TODO())
		if err != nil {
			span.SetStatus(codes.Error, "can't describe RDS clusters")
			span.RecordError(err)

			return clusterMetrics, fmt.Errorf("can't describe RDS clusters: %w", err)
		}

		span.SetStatus(codes.Ok, "metrics fetched")
		span.SetAttributes(attribute.Int("qonto.prometheus_rds_exporter.cluster_count", len(output.DBClusters)))

		for _, dbCluster := range output.DBClusters {

			var writerDBInstanceIdentifier string
			members := make(map[string]DBRole)

			for _, member := range dbCluster.DBClusterMembers {
				instanceID := aws.ToString(member.DBInstanceIdentifier)

				if aws.ToBool(member.IsClusterWriter) {
					members[instanceID] = RoleWriter
					writerDBInstanceIdentifier = instanceID
				} else {
					members[instanceID] = RoleReader
				}
			}

			clusterMetrics[*dbCluster.DBClusterIdentifier] = ClusterMetrics{
				Arn:                        *dbCluster.DBClusterArn,
				Engine:                     *dbCluster.Engine,
				EngineVersion:              *dbCluster.EngineVersion,
				AllocatedStorage:           converter.GigaBytesToBytes(int64(*dbCluster.AllocatedStorage)),
				DBClusterIdentifier:        *dbCluster.DBClusterIdentifier,
				Members:                    members,
				WriterDBInstanceIdentifier: writerDBInstanceIdentifier,
				DbClusterResourceId:        *dbCluster.DbClusterResourceId,
				Age:                        time.Since(*dbCluster.ClusterCreateTime).Seconds(),
				Tags:                       ConvertRDSTagsToMap(dbCluster.TagList),
			}
		}
	}

	return clusterMetrics, nil
}

func (r *RDSFetcher) GetInstancesMetrics() (Metrics, error) {
	ctx, span := tracer.Start(r.ctx, "collect-instance-metrics")
	defer span.End()

	metrics := make(map[string]RdsInstanceMetrics)

	var err error

	var instanceMaintenances map[string]string

	if r.configuration.CollectMaintenances {
		instanceMaintenances, err = r.getPendingMaintenances(ctx)
		if err != nil {
			span.SetStatus(codes.Error, "can't get RDS maintenances")
			span.RecordError(err)

			return Metrics{}, fmt.Errorf("can't get RDS maintenances: %w", err)
		}
	}

	filters, err := r.getDBInstanceFilters(ctx)
	if err != nil {
		return Metrics{}, err
	}

	clusterMetrics, err := r.getClusters(ctx, filters)
	if err != nil {
		return Metrics{}, fmt.Errorf("can't get cluster metrics: %w", err)
	}

	input := &aws_rds.DescribeDBInstancesInput{Filters: filters}
	paginator := aws_rds.NewDescribeDBInstancesPaginator(r.client, input)
	for paginator.HasMorePages() {
		instanceCtx, instanceSpan := tracer.Start(ctx, "collect-rds-instances")
		defer instanceSpan.End()

		r.statistics.RdsAPICall++

		output, err := paginator.NextPage(context.TODO())
		if err != nil {
			span.SetStatus(codes.Error, "can't get RDS instances")
			span.RecordError(err)

			return Metrics{}, fmt.Errorf("can't get RDS instances: %w", err)
		}

		for _, dbInstance := range output.DBInstances {
			dbIdentifier := dbInstance.DBInstanceIdentifier

			instanceMetrics, err := r.computeInstanceMetrics(instanceCtx, dbInstance, instanceMaintenances, &clusterMetrics)
			if err != nil {
				span.SetStatus(codes.Error, "can't compute instance metrics")
				span.RecordError(err)

				return Metrics{}, fmt.Errorf("can't compute instance metrics for %s: %w", *dbIdentifier, err)
			}

			metrics[*dbIdentifier] = instanceMetrics
		}

		instanceSpan.SetStatus(codes.Ok, "instance metrics fetch")
	}

	span.SetStatus(codes.Ok, "metrics fetched")

	return Metrics{Instances: metrics, Clusters: clusterMetrics}, nil
}

func (r *RDSFetcher) getDBInstanceFilters(ctx context.Context) ([]aws_rds_types.Filter, error) {
	var filters []aws_rds_types.Filter
	if r.configuration.TagSelections == nil {
		return filters, nil
	}

	var tagFilters []tag_types.TagFilter

	for k, v := range r.configuration.TagSelections {
		keyCopy := k
		tagFilters = append(tagFilters, tag_types.TagFilter{
			Key:    &keyCopy,
			Values: v,
		})
	}

	_, resourcesSpan := tracer.Start(ctx, "find-rds-instances")
	resourcesInput := &resourcegroupstaggingapi.GetResourcesInput{
		ResourceTypeFilters: []string{"rds:db"},
		TagFilters:          tagFilters,
	}

	var arns []string

	resourcesPaginator := resourcegroupstaggingapi.NewGetResourcesPaginator(r.tagClient, resourcesInput)

	for resourcesPaginator.HasMorePages() {
		r.statistics.TagAPICall++

		resources, err := resourcesPaginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("can't find instances for tags %v: %w", r.configuration.TagSelections, err)
		}

		for _, res := range resources.ResourceTagMappingList {
			arns = append(arns, *res.ResourceARN)
		}
	}

	if len(arns) == 0 {
		r.logger.Warn(fmt.Sprintf("no resources any matching tag selection (won't limit which dbs to get metrics for): %v", r.configuration.TagSelections))
		resourcesSpan.SetStatus(codes.Error, "did not find any RDS instances matching tag selection")
	} else {
		id := "db-instance-id"
		filters = append(filters, aws_rds_types.Filter{
			Name:   &id,
			Values: arns,
		})

		resourcesSpan.SetStatus(codes.Ok, "found RDS instances matching tag selection")
	}

	return filters, nil
}

// computeInstanceMetrics returns metrics about the specified instance
func (r *RDSFetcher) computeInstanceMetrics(ctx context.Context, dbInstance aws_rds_types.DBInstance, instanceMaintenances map[string]string, clusterMetrics *map[string]ClusterMetrics) (RdsInstanceMetrics, error) {
	dbIdentifier := dbInstance.DBInstanceIdentifier

	var iops int64
	if dbInstance.Iops != nil {
		iops = int64(*dbInstance.Iops)
	}

	var throughput int64
	if dbInstance.StorageThroughput != nil {
		throughput = int64(*dbInstance.StorageThroughput)
	}

	iops, storageThroughput := getStorageMetrics(*dbInstance.StorageType, int64(*dbInstance.AllocatedStorage), iops, throughput)

	var maxAllocatedStorage int64 = 0
	if dbInstance.MaxAllocatedStorage != nil {
		maxAllocatedStorage = int64(*dbInstance.MaxAllocatedStorage)
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

		logFilesSize, err = r.getLogFilesSize(ctx, *dbIdentifier)
		if err != nil {
			return RdsInstanceMetrics{}, fmt.Errorf("can't get log files size for %d: %w", dbIdentifier, err)
		}
	}

	var clusterDetails ClusterMetrics

	if dbInstance.DBClusterIdentifier != nil {
		if details, exists := (*clusterMetrics)[*dbInstance.DBClusterIdentifier]; exists {
			clusterDetails = details
		}
	}
	role, sourceDBInstanceIdentifier := GetInstanceRole(&dbInstance, clusterDetails)

	var age *float64

	if dbInstance.InstanceCreateTime != nil {
		diff := time.Since(*dbInstance.InstanceCreateTime).Seconds()
		age = &diff
	}

	var certificateValidTill *time.Time

	if dbInstance.CertificateDetails != nil && dbInstance.CertificateDetails.ValidTill != nil {
		certificateValidTill = dbInstance.CertificateDetails.ValidTill
	}

	metrics := RdsInstanceMetrics{
		Arn:                        *dbInstance.DBInstanceArn,
		AllocatedStorage:           converter.GigaBytesToBytes(int64(*dbInstance.AllocatedStorage)),
		BackupRetentionPeriod:      converter.DaystoSeconds(*dbInstance.BackupRetentionPeriod),
		DBInstanceClass:            *dbInstance.DBInstanceClass,
		DbiResourceID:              *dbInstance.DbiResourceId,
		DeletionProtection:         aws.ToBool(dbInstance.DeletionProtection),
		Engine:                     *dbInstance.Engine,
		EngineVersion:              *dbInstance.EngineVersion,
		LogFilesSize:               logFilesSize,
		MaxAllocatedStorage:        converter.GigaBytesToBytes(maxAllocatedStorage),
		MaxIops:                    iops,
		MultiAZ:                    aws.ToBool(dbInstance.MultiAZ),
		PendingMaintenanceAction:   pendingMaintenanceAction,
		PendingModifiedValues:      pendingModifiedValues,
		PerformanceInsightsEnabled: aws.ToBool(dbInstance.PerformanceInsightsEnabled),
		PubliclyAccessible:         aws.ToBool(dbInstance.PubliclyAccessible),
		Role:                       role,
		SourceDBInstanceIdentifier: sourceDBInstanceIdentifier,
		Status:                     GetDBInstanceStatusCode(*dbInstance.DBInstanceStatus),
		StorageThroughput:          converter.MegaBytesToBytes(storageThroughput),
		StorageType:                aws.ToString(dbInstance.StorageType),
		CACertificateIdentifier:    aws.ToString(dbInstance.CACertificateIdentifier),
		CertificateValidTill:       certificateValidTill,
		Age:                        age,
		Tags:                       ConvertRDSTagsToMap(dbInstance.TagList),
	}

	return metrics, nil
}

// getLogFilesSize returns the size of all logs on the specified instance
func (r *RDSFetcher) getLogFilesSize(ctx context.Context, dbidentifier string) (*int64, error) {
	_, span := tracer.Start(ctx, "collect-instance-log")
	defer span.End()

	span.SetAttributes(semconv.DBInstanceID(dbidentifier))

	var filesSize *int64

	input := &aws_rds.DescribeDBLogFilesInput{DBInstanceIdentifier: &dbidentifier}

	r.statistics.RdsAPICall++

	result, err := r.client.DescribeDBLogFiles(context.TODO(), input)
	if err != nil {
		var notFoundError *aws_rds_types.DBInstanceNotFoundFault
		if errors.As(err, &notFoundError) { // Replica in "creating" status may return notFoundError exception
			return filesSize, nil
		}

		span.SetStatus(codes.Error, "can't describe db logs files")
		span.RecordError(err)

		return filesSize, fmt.Errorf("can't describe db logs files for %s: %w", dbidentifier, err)
	}

	if result != nil && len(result.DescribeDBLogFiles) > 0 {
		if filesSize == nil {
			filesSize = new(int64)
		}

		for _, file := range result.DescribeDBLogFiles {
			*filesSize += *file.Size
		}
	}

	return filesSize, nil
}
