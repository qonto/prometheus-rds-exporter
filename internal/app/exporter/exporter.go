// Package exporter implements Prometheus exporter
package exporter

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"sync"

	"github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/qonto/prometheus-rds-exporter/internal/app/cloudwatch"
	"github.com/qonto/prometheus-rds-exporter/internal/app/ec2"
	"github.com/qonto/prometheus-rds-exporter/internal/app/rds"
	"github.com/qonto/prometheus-rds-exporter/internal/app/servicequotas"
	"github.com/qonto/prometheus-rds-exporter/internal/infra/build"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const (
	exporterUpStatusCode   float64 = 1
	exporterDownStatusCode float64 = 0
)

var tracer = otel.Tracer("github/qonto/prometheus-rds-exporter/internal/app/exporter")

type Configuration struct {
	CollectInstanceMetrics bool
	CollectInstanceTags    bool
	CollectInstanceTypes   bool
	CollectLogsSize        bool
	CollectMaintenances    bool
	CollectQuotas          bool
	CollectUsages          bool
	IncludeTagsInMetrics   bool
	TagSelections          map[string][]string
}

type counters struct {
	CloudwatchAPICalls    float64
	EC2APIcalls           float64
	Errors                float64
	RDSAPIcalls           float64
	ServiceQuotasAPICalls float64
	UsageAPIcalls         float64
	TagAPICalls           float64
}

type metrics struct {
	ServiceQuota        servicequotas.Metrics
	RDS                 rds.Metrics
	EC2                 ec2.Metrics
	CloudwatchInstances cloudwatch.CloudWatchMetrics
	CloudWatchUsage     cloudwatch.UsageMetrics
}

type rdsCollector struct {
	ctx           context.Context
	wg            sync.WaitGroup
	logger        slog.Logger
	counters      counters
	metrics       metrics
	awsAccountID  string
	awsRegion     string
	configuration Configuration

	rdsClient           rdsClient
	EC2Client           EC2Client
	servicequotasClient servicequotasClient
	cloudWatchClient    cloudWatchClient
	tagClient           resourcegroupstaggingapi.GetResourcesAPIClient

	errors                      *prometheus.Desc
	DBLoad                      *prometheus.Desc
	dBLoadCPU                   *prometheus.Desc
	dBLoadNonCPU                *prometheus.Desc
	allocatedStorage            *prometheus.Desc
	allocatedDiskIOPS           *prometheus.Desc
	allocatedDiskThroughput     *prometheus.Desc
	information                 *prometheus.Desc
	instanceBaselineIops        *prometheus.Desc
	instanceMaximumIops         *prometheus.Desc
	instanceBaselineThroughput  *prometheus.Desc
	instanceMaximumThroughput   *prometheus.Desc
	instanceMemory              *prometheus.Desc
	instanceVCPU                *prometheus.Desc
	instanceTags                *prometheus.Desc
	logFilesSize                *prometheus.Desc
	maxAllocatedStorage         *prometheus.Desc
	maxIops                     *prometheus.Desc
	status                      *prometheus.Desc
	storageThroughput           *prometheus.Desc
	up                          *prometheus.Desc
	cpuUtilisation              *prometheus.Desc
	freeStorageSpace            *prometheus.Desc
	databaseConnections         *prometheus.Desc
	freeableMemory              *prometheus.Desc
	swapUsage                   *prometheus.Desc
	writeIOPS                   *prometheus.Desc
	readIOPS                    *prometheus.Desc
	replicaLag                  *prometheus.Desc
	replicationSlotDiskUsage    *prometheus.Desc
	maximumUsedTransactionIDs   *prometheus.Desc
	apiCall                     *prometheus.Desc
	readThroughput              *prometheus.Desc
	writeThroughput             *prometheus.Desc
	backupRetentionPeriod       *prometheus.Desc
	quotaDBInstances            *prometheus.Desc
	quotaTotalStorage           *prometheus.Desc
	quotaMaxDBInstanceSnapshots *prometheus.Desc
	usageAllocatedStorage       *prometheus.Desc
	usageDBInstances            *prometheus.Desc
	usageManualSnapshots        *prometheus.Desc
	exporterBuildInformation    *prometheus.Desc
	transactionLogsDiskUsage    *prometheus.Desc
	certificateValidTill        *prometheus.Desc
	age                         *prometheus.Desc
}

func NewCollector(logger slog.Logger, collectorConfiguration Configuration, awsAccountID string, awsRegion string, rdsClient rdsClient, ec2Client EC2Client, cloudWatchClient cloudWatchClient, servicequotasClient servicequotasClient, tagClient resourcegroupstaggingapi.GetResourcesAPIClient) *rdsCollector {
	return &rdsCollector{
		logger:              logger,
		awsAccountID:        awsAccountID,
		awsRegion:           awsRegion,
		rdsClient:           rdsClient,
		servicequotasClient: servicequotasClient,
		EC2Client:           ec2Client,
		cloudWatchClient:    cloudWatchClient,
		tagClient:           tagClient,

		configuration: collectorConfiguration,

		exporterBuildInformation: prometheus.NewDesc("rds_exporter_build_info",
			"A metric with constant '1' value labeled by version from which exporter was built",
			[]string{"version", "commit_sha", "build_date"}, nil,
		),
		errors: prometheus.NewDesc("rds_exporter_errors_total",
			"Total number of errors encountered by the exporter",
			[]string{}, nil,
		),
		allocatedStorage: prometheus.NewDesc("rds_allocated_storage_bytes",
			"Allocated storage",
			[]string{"aws_account_id", "aws_region", "dbidentifier"}, nil,
		),
		allocatedDiskIOPS: prometheus.NewDesc("rds_allocated_disk_iops_average",
			"Allocated disk IOPS",
			[]string{"aws_account_id", "aws_region", "dbidentifier"}, nil,
		),
		allocatedDiskThroughput: prometheus.NewDesc("rds_allocated_disk_throughput_bytes",
			"Allocated disk throughput",
			[]string{"aws_account_id", "aws_region", "dbidentifier"}, nil,
		),
		information: prometheus.NewDesc("rds_instance_info",
			"RDS instance information",
			[]string{"aws_account_id", "aws_region", "dbidentifier", "dbi_resource_id", "instance_class", "engine", "engine_version", "storage_type", "multi_az", "deletion_protection", "role", "source_dbidentifier", "pending_modified_values", "pending_maintenance", "performance_insights_enabled", "ca_certificate_identifier", "arn"}, nil,
		),
		age: prometheus.NewDesc("rds_instance_age_seconds",
			"Time since instance creation",
			[]string{"aws_account_id", "aws_region", "dbidentifier"}, nil,
		),
		maxAllocatedStorage: prometheus.NewDesc("rds_max_allocated_storage_bytes",
			"Upper limit in gibibytes to which Amazon RDS can automatically scale the storage of the DB instance",
			[]string{"aws_account_id", "aws_region", "dbidentifier"}, nil,
		),
		maxIops: prometheus.NewDesc("rds_max_disk_iops_average",
			"Max disk IOPS evaluated with disk IOPS and EC2 capacity",
			[]string{"aws_account_id", "aws_region", "dbidentifier"}, nil,
		),
		storageThroughput: prometheus.NewDesc("rds_max_storage_throughput_bytes",
			"Max disk throughput evaluated with disk throughput and EC2 capacity",
			[]string{"aws_account_id", "aws_region", "dbidentifier"}, nil,
		),
		readThroughput: prometheus.NewDesc("rds_read_throughput_bytes",
			"Average number of bytes read from disk per second",
			[]string{"aws_account_id", "aws_region", "dbidentifier"}, nil,
		),
		writeThroughput: prometheus.NewDesc("rds_write_throughput_bytes",
			"Average number of bytes written to disk per second",
			[]string{"aws_account_id", "aws_region", "dbidentifier"}, nil,
		),
		status: prometheus.NewDesc("rds_instance_status",
			"Instance status (0 stopped or can't scrape) (1 ok | 2 backup | 3 startup | 4 modify | 5 monitoring config | 1X storage | 20 renaming) (-1 unknown | -2 stopping | -3 creating | -4 deleting | -5 rebooting | -6 failed | -7 full storage | -8 upgrading | -9 maintenance | -10 restore error)",
			[]string{"aws_account_id", "aws_region", "dbidentifier"}, nil,
		),
		logFilesSize: prometheus.NewDesc("rds_instance_log_files_size_bytes",
			"Total of log files on the instance",
			[]string{"aws_account_id", "aws_region", "dbidentifier"}, nil,
		),
		instanceVCPU: prometheus.NewDesc("rds_instance_vcpu_average",
			"Total vCPU for this instance class",
			[]string{"aws_account_id", "aws_region", "instance_class"}, nil,
		),
		instanceMemory: prometheus.NewDesc("rds_instance_memory_bytes",
			"Instance class memory",
			[]string{"aws_account_id", "aws_region", "instance_class"}, nil,
		),
		instanceTags: prometheus.NewDesc("rds_instance_tags",
			"AWS tags attached to the instance",
			[]string{"aws_account_id", "aws_region", "dbidentifier"}, nil,
		),
		cpuUtilisation: prometheus.NewDesc("rds_cpu_usage_percent_average",
			"Instance CPU used",
			[]string{"aws_account_id", "aws_region", "dbidentifier"}, nil,
		),
		instanceMaximumThroughput: prometheus.NewDesc("rds_instance_max_throughput_bytes",
			"Maximum throughput of underlying EC2 instance class",
			[]string{"aws_account_id", "aws_region", "instance_class"}, nil,
		),
		instanceBaselineThroughput: prometheus.NewDesc("rds_instance_baseline_throughput_bytes",
			"Baseline throughput of underlying EC2 instance class",
			[]string{"aws_account_id", "aws_region", "instance_class"}, nil,
		),
		instanceMaximumIops: prometheus.NewDesc("rds_instance_max_iops_average",
			"Maximum IOPS of underlying EC2 instance class",
			[]string{"aws_account_id", "aws_region", "instance_class"}, nil,
		),
		instanceBaselineIops: prometheus.NewDesc("rds_instance_baseline_iops_average",
			"Baseline IOPS of underlying EC2 instance class",
			[]string{"aws_account_id", "aws_region", "instance_class"}, nil,
		),
		freeStorageSpace: prometheus.NewDesc("rds_free_storage_bytes",
			"Free storage on the instance",
			[]string{"aws_account_id", "aws_region", "dbidentifier"}, nil,
		),
		databaseConnections: prometheus.NewDesc("rds_database_connections_average",
			"The number of client network connections to the database instance",
			[]string{"aws_account_id", "aws_region", "dbidentifier"}, nil,
		),
		up: prometheus.NewDesc("up",
			"Was the last scrape of RDS successful",
			nil, nil,
		),
		swapUsage: prometheus.NewDesc("rds_swap_usage_bytes",
			"Amount of swap space used on the DB instance. This metric is not available for SQL Server",
			[]string{"aws_account_id", "aws_region", "dbidentifier"}, nil,
		),
		writeIOPS: prometheus.NewDesc("rds_write_iops_average",
			"Average number of disk write I/O operations per second",
			[]string{"aws_account_id", "aws_region", "dbidentifier"}, nil,
		),
		readIOPS: prometheus.NewDesc("rds_read_iops_average",
			"Average number of disk read I/O operations per second",
			[]string{"aws_account_id", "aws_region", "dbidentifier"}, nil,
		),
		replicaLag: prometheus.NewDesc("rds_replica_lag_seconds",
			"For read replica configurations, the amount of time a read replica DB instance lags behind the source DB instance. Applies to MariaDB, Microsoft SQL Server, MySQL, Oracle, and PostgreSQL read replicas",
			[]string{"aws_account_id", "aws_region", "dbidentifier"}, nil,
		),
		replicationSlotDiskUsage: prometheus.NewDesc("rds_replication_slot_disk_usage_bytes",
			"Disk space used by replication slot files. Applies to PostgreSQL",
			[]string{"aws_account_id", "aws_region", "dbidentifier"}, nil,
		),
		maximumUsedTransactionIDs: prometheus.NewDesc("rds_maximum_used_transaction_ids_average",
			"Maximum transaction IDs that have been used. Applies to only PostgreSQL",
			[]string{"aws_account_id", "aws_region", "dbidentifier"}, nil,
		),
		freeableMemory: prometheus.NewDesc("rds_freeable_memory_bytes",
			"Amount of available random access memory. For MariaDB, MySQL, Oracle, and PostgreSQL DB instances, this metric reports the value of the MemAvailable field of /proc/meminfo",
			[]string{"aws_account_id", "aws_region", "dbidentifier"}, nil,
		),
		apiCall: prometheus.NewDesc("rds_api_call_total",
			"Number of call to AWS API",
			[]string{"aws_account_id", "aws_region", "api"}, nil,
		),
		backupRetentionPeriod: prometheus.NewDesc("rds_backup_retention_period_seconds",
			"Automatic DB snapshots retention period",
			[]string{"aws_account_id", "aws_region", "dbidentifier"}, nil,
		),
		DBLoad: prometheus.NewDesc("rds_dbload_average",
			"Number of active sessions for the DB engine",
			[]string{"aws_account_id", "aws_region", "dbidentifier"}, nil,
		),
		dBLoadCPU: prometheus.NewDesc("rds_dbload_cpu_average",
			"Number of active sessions where the wait event type is CPU",
			[]string{"aws_account_id", "aws_region", "dbidentifier"}, nil,
		),
		dBLoadNonCPU: prometheus.NewDesc("rds_dbload_noncpu_average",
			"Number of active sessions where the wait event type is not CPU",
			[]string{"aws_account_id", "aws_region", "dbidentifier"}, nil,
		),
		transactionLogsDiskUsage: prometheus.NewDesc("rds_transaction_logs_disk_usage_bytes",
			"Disk space used by transaction logs (only on PostgreSQL)",
			[]string{"aws_account_id", "aws_region", "dbidentifier"}, nil,
		),
		certificateValidTill: prometheus.NewDesc("rds_certificate_expiry_timestamp_seconds",
			"Timestamp of the expiration of the Instance certificate",
			[]string{"aws_account_id", "aws_region", "dbidentifier"}, nil,
		),
		quotaDBInstances: prometheus.NewDesc("rds_quota_max_dbinstances_average",
			"Maximum number of RDS instances allowed in the AWS account",
			[]string{"aws_account_id", "aws_region"}, nil,
		),
		quotaTotalStorage: prometheus.NewDesc("rds_quota_total_storage_bytes",
			"Maximum total storage for all DB instances",
			[]string{"aws_account_id", "aws_region"}, nil,
		),
		quotaMaxDBInstanceSnapshots: prometheus.NewDesc("rds_quota_maximum_db_instance_snapshots_average",
			"Maximum number of manual DB instance snapshots",
			[]string{"aws_account_id", "aws_region"}, nil,
		),
		usageAllocatedStorage: prometheus.NewDesc("rds_usage_allocated_storage_bytes",
			"Total storage used by AWS RDS instances",
			[]string{"aws_account_id", "aws_region"}, nil,
		),
		usageDBInstances: prometheus.NewDesc("rds_usage_db_instances_average",
			"AWS RDS instance count",
			[]string{"aws_account_id", "aws_region"}, nil,
		),
		usageManualSnapshots: prometheus.NewDesc("rds_usage_manual_snapshots_average",
			"Manual snapshots count",
			[]string{"aws_account_id", "aws_region"}, nil,
		),
	}
}

func (c *rdsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.DBLoad
	ch <- c.age
	ch <- c.allocatedStorage
	ch <- c.allocatedDiskIOPS
	ch <- c.allocatedDiskThroughput
	ch <- c.apiCall
	ch <- c.apiCall
	ch <- c.backupRetentionPeriod
	ch <- c.certificateValidTill
	ch <- c.cpuUtilisation
	ch <- c.dBLoadCPU
	ch <- c.dBLoadNonCPU
	ch <- c.databaseConnections
	ch <- c.errors
	ch <- c.exporterBuildInformation
	ch <- c.freeStorageSpace
	ch <- c.freeableMemory
	ch <- c.information
	ch <- c.instanceBaselineIops
	ch <- c.instanceMaximumIops
	ch <- c.instanceBaselineThroughput
	ch <- c.instanceMaximumThroughput
	ch <- c.instanceMemory
	ch <- c.instanceVCPU
	ch <- c.logFilesSize
	ch <- c.maxAllocatedStorage
	ch <- c.maxIops
	ch <- c.maximumUsedTransactionIDs
	ch <- c.quotaDBInstances
	ch <- c.quotaMaxDBInstanceSnapshots
	ch <- c.quotaTotalStorage
	ch <- c.readIOPS
	ch <- c.readThroughput
	ch <- c.replicaLag
	ch <- c.replicationSlotDiskUsage
	ch <- c.status
	ch <- c.storageThroughput
	ch <- c.swapUsage
	ch <- c.transactionLogsDiskUsage
	ch <- c.up
	ch <- c.usageAllocatedStorage
	ch <- c.usageDBInstances
	ch <- c.usageManualSnapshots
	ch <- c.writeIOPS
	ch <- c.writeThroughput
}

// getMetrics collects and return all RDS metrics
func (c *rdsCollector) fetchMetrics() error {
	c.logger.Debug("received query")

	// Fetch serviceQuotas metrics
	if c.configuration.CollectQuotas {
		go c.getQuotasMetrics(c.servicequotasClient)
		c.wg.Add(1)
	}

	// Fetch usages metrics
	if c.configuration.CollectUsages {
		go c.getUsagesMetrics(c.cloudWatchClient)
		c.wg.Add(1)
	}

	// Fetch RDS instances metrics
	c.logger.Debug("get RDS metrics")

	rdsFetcher := rds.NewFetcher(c.ctx, c.rdsClient, c.tagClient, c.logger, rds.Configuration{
		CollectLogsSize:     c.configuration.CollectLogsSize,
		CollectMaintenances: c.configuration.CollectMaintenances,
		TagSelections:       c.configuration.TagSelections,
	})

	rdsMetrics, err := rdsFetcher.GetInstancesMetrics()
	if err != nil {
		return fmt.Errorf("can't fetch RDS metrics: %w", err)
	}

	c.metrics.RDS = rdsMetrics
	c.counters.RDSAPIcalls += rdsFetcher.GetStatistics().RdsAPICall
	c.counters.TagAPICalls += rdsFetcher.GetStatistics().TagAPICall
	c.logger.Debug("RDS metrics fetched")

	// Compute uniq instances identifiers and instance types
	instanceIdentifiers, instanceTypes := getUniqTypeAndIdentifiers(rdsMetrics.Instances)

	// Fetch EC2 Metrics for instance types
	if c.configuration.CollectInstanceTypes && len(instanceTypes) > 0 {
		go c.getEC2Metrics(c.EC2Client, instanceTypes)
		c.wg.Add(1)
	}

	// Fetch Cloudwatch metrics for instances
	if c.configuration.CollectInstanceMetrics {
		go c.getCloudwatchMetrics(c.cloudWatchClient, instanceIdentifiers)
		c.wg.Add(1)
	}

	// Wait for all go routines to finish
	c.wg.Wait()

	return nil
}

func (c *rdsCollector) getCloudwatchMetrics(client cloudwatch.CloudWatchClient, instanceIdentifiers []string) {
	defer c.wg.Done()
	c.logger.Debug("fetch cloudwatch metrics")

	_, span := tracer.Start(c.ctx, "collect-cloudwatch-metrics")
	defer span.End()

	fetcher := cloudwatch.NewRDSFetcher(client, c.logger)

	metrics, err := fetcher.GetRDSInstanceMetrics(instanceIdentifiers)
	if err != nil {
		c.counters.Errors++
	}

	c.counters.CloudwatchAPICalls += fetcher.GetStatistics().CloudWatchAPICall
	c.metrics.CloudwatchInstances = metrics

	c.logger.Debug("cloudwatch metrics fetched", "metrics", metrics)
}

func (c *rdsCollector) getUsagesMetrics(client cloudwatch.CloudWatchClient) {
	defer c.wg.Done()
	c.logger.Debug("fetch usage metrics")

	fetcher := cloudwatch.NewUsageFetcher(c.ctx, client, c.logger)

	metrics, err := fetcher.GetUsageMetrics()
	if err != nil {
		c.counters.Errors++
		c.logger.Error(fmt.Sprintf("can't fetch usage metrics: %s", err))
	}

	c.counters.UsageAPIcalls += fetcher.GetStatistics().CloudWatchAPICall
	c.metrics.CloudWatchUsage = metrics

	c.logger.Debug("usage metrics fetched", "metrics", metrics)
}

func (c *rdsCollector) getEC2Metrics(client ec2.EC2Client, instanceTypes []string) {
	defer c.wg.Done()
	c.logger.Debug("fetch EC2 metrics")

	fetcher := ec2.NewFetcher(c.ctx, client)

	metrics, err := fetcher.GetDBInstanceTypeInformation(instanceTypes)
	if err != nil {
		c.counters.Errors++
		c.logger.Error(fmt.Sprintf("can't fetch EC2 metrics: %s", err))
	}

	c.counters.EC2APIcalls += fetcher.GetStatistics().EC2ApiCall
	c.metrics.EC2 = metrics

	c.logger.Debug("EC2 metrics fetched", "metrics", metrics)
}

func (c *rdsCollector) getQuotasMetrics(client servicequotas.ServiceQuotasClient) {
	defer c.wg.Done()

	ctx, span := tracer.Start(c.ctx, "collect-quota-metrics")
	defer span.End()

	c.logger.Debug("fetch quotas")

	fetcher := servicequotas.NewFetcher(ctx, client, c.logger)

	metrics, err := fetcher.GetRDSQuotas()
	if err != nil {
		c.counters.Errors++
		c.logger.Error(fmt.Sprintf("can't fetch service quota metrics: %s", err))
		span.SetStatus(codes.Error, "can't fetch service quota metrics")
		span.RecordError(err)
	}

	c.counters.ServiceQuotasAPICalls += fetcher.GetStatistics().UsageAPICall
	c.metrics.ServiceQuota = metrics

	span.SetStatus(codes.Ok, "quota fetched")
}

// getBaseLabels returns the standard labels without tags as a map
func (c *rdsCollector) getBaseLabels(dbidentifier string) map[string]string {
	return map[string]string{
		"aws_account_id": c.awsAccountID,
		"aws_region":     c.awsRegion,
		"dbidentifier":   dbidentifier,
	}
}

// getTagsLabels returns the instance tags as labels if a collection is enabled
func (c *rdsCollector) getTagLabels(instance rds.RdsInstanceMetrics) (keys []string, values []string) {
	if !c.configuration.CollectInstanceTags {
		return nil, nil
	}

	// Create a map for deduplication
	tagLabels := make(map[string]string)

	for k, v := range instance.Tags {
		// Skip empty keys or values
		if k == "" || v == "" {
			continue
		}

		// Normalize the key name for Prometheus
		normalizedKey := clearPrometheusLabel(k)
		if normalizedKey == "" {
			continue
		}

		// Prefix tag keys with tag_ to avoid conflicts with base labels
		labelName := fmt.Sprintf("tag_%s", normalizedKey)
		tagLabels[labelName] = v
	}

	// Convert to slices
	for k, v := range tagLabels {
		keys = append(keys, k)
		values = append(values, v)
	}

	return keys, values
}

// getInstanceTagLabels returns the base labels and optionally includes tags as labels
func (c *rdsCollector) getInstanceTagLabels(dbidentifier string, instance rds.RdsInstanceMetrics) (keys []string, values []string) {
	// Start with base labels
	labels := c.getBaseLabels(dbidentifier)

	// Add instance tags to labels only if collection is enabled and they should be included in metrics
	if c.configuration.CollectInstanceTags && c.configuration.IncludeTagsInMetrics {
		tagKeys, tagValues := c.getTagLabels(instance)
		for i, key := range tagKeys {
			labels[key] = tagValues[i]
		}
	}

	// Convert map to slices
	for k, v := range labels {
		keys = append(keys, k)
		values = append(values, v)
	}

	return keys, values
}

// createDynamicMetric creates a new metric with tags included as labels
func (c *rdsCollector) createDynamicMetric(name, help string, instance rds.RdsInstanceMetrics, dbidentifier string) *prometheus.Desc {
	keys, _ := c.getInstanceTagLabels(dbidentifier, instance)
	return prometheus.NewDesc(name, help, keys, nil)
}

func (c *rdsCollector) Collect(ch chan<- prometheus.Metric) {
	ch <- prometheus.MustNewConstMetric(c.exporterBuildInformation, prometheus.GaugeValue, 1, build.Version, build.CommitSHA, build.Date)
	ch <- prometheus.MustNewConstMetric(c.errors, prometheus.CounterValue, c.counters.Errors)

	var span trace.Span

	c.ctx, span = tracer.Start(context.TODO(), "collect-metrics")
	defer span.End()

	// Get all metrics
	err := c.fetchMetrics()
	if err != nil {
		c.logger.Error(fmt.Sprintf("can't scrape metrics: %s", err))
		// Mark exporter as down
		ch <- prometheus.MustNewConstMetric(c.up, prometheus.CounterValue, exporterDownStatusCode)

		span.SetStatus(codes.Error, "failed to get metrics")
		span.RecordError(err)

		return
	}

	span.End()

	ch <- prometheus.MustNewConstMetric(c.up, prometheus.CounterValue, exporterUpStatusCode)

	// RDS metrics
	ch <- prometheus.MustNewConstMetric(c.apiCall, prometheus.CounterValue, c.counters.RDSAPIcalls, c.awsAccountID, c.awsRegion, "rds")
	ch <- prometheus.MustNewConstMetric(c.apiCall, prometheus.CounterValue, c.counters.TagAPICalls, c.awsAccountID, c.awsRegion, "tag")

	for dbidentifier, instance := range c.metrics.RDS.Instances {
		// For each instance, create metrics with tags included
		keys, values := c.getInstanceTagLabels(dbidentifier, instance)

		// Create dynamic metric descriptors with tags as labels
		allocatedStorageDesc := prometheus.NewDesc("rds_allocated_storage_bytes", "Allocated storage", keys, nil)
		ch <- prometheus.MustNewConstMetric(
			allocatedStorageDesc,
			prometheus.GaugeValue,
			float64(instance.AllocatedStorage),
			values...,
		)
		// Combine the standard info labels with tag labels
		infoKeys := append([]string{}, keys...)
		infoValues := append([]string{}, values...)

		// Add the remaining information fields that are specific to the info metric
		infoKeys = append(infoKeys, "dbi_resource_id", "instance_class", "engine", "engine_version", "storage_type", "multi_az",
			"deletion_protection", "role", "source_dbidentifier", "pending_modified_values", "pending_maintenance",
			"performance_insights_enabled", "ca_certificate_identifier", "arn")

		infoValues = append(infoValues, instance.DbiResourceID, instance.DBInstanceClass, instance.Engine, instance.EngineVersion,
			instance.StorageType, strconv.FormatBool(instance.MultiAZ), strconv.FormatBool(instance.DeletionProtection),
			instance.Role, instance.SourceDBInstanceIdentifier, strconv.FormatBool(instance.PendingModifiedValues),
			instance.PendingMaintenanceAction, strconv.FormatBool(instance.PerformanceInsightsEnabled),
			instance.CACertificateIdentifier, instance.Arn)

		informationDesc := prometheus.NewDesc("rds_instance_info", "RDS instance information", infoKeys, nil)
		ch <- prometheus.MustNewConstMetric(
			informationDesc,
			prometheus.GaugeValue,
			1,
			infoValues...,
		)
		if instance.MaxAllocatedStorage > 0 {
			maxAllocatedStorageDesc := prometheus.NewDesc("rds_max_allocated_storage_bytes", "Upper limit in gibibytes to which Amazon RDS can automatically scale the storage of the DB instance", keys, nil)
			ch <- prometheus.MustNewConstMetric(maxAllocatedStorageDesc, prometheus.GaugeValue, float64(instance.MaxAllocatedStorage), values...)
		}
		if instance.MaxIops > 0 {
			allocatedDiskIOPSDesc := prometheus.NewDesc("rds_allocated_disk_iops_average", "Allocated disk IOPS", keys, nil)
			ch <- prometheus.MustNewConstMetric(allocatedDiskIOPSDesc, prometheus.GaugeValue, float64(instance.MaxIops), values...)
		}
		if instance.StorageThroughput > 0 {
			allocatedDiskThroughputDesc := prometheus.NewDesc("rds_allocated_disk_throughput_bytes", "Allocated disk throughput", keys, nil)
			ch <- prometheus.MustNewConstMetric(allocatedDiskThroughputDesc, prometheus.GaugeValue, float64(instance.StorageThroughput), values...)
		}

		statusDesc := prometheus.NewDesc("rds_instance_status", "Instance status (0 stopped or can't scrape) (1 ok | 2 backup | 3 startup | 4 modify | 5 monitoring config | 1X storage | 20 renaming) (-1 unknown | -2 stopping | -3 creating | -4 deleting | -5 rebooting | -6 failed | -7 full storage | -8 upgrading | -9 maintenance | -10 restore error)", keys, nil)
		ch <- prometheus.MustNewConstMetric(statusDesc, prometheus.GaugeValue, float64(instance.Status), values...)

		backupRetentionPeriodDesc := prometheus.NewDesc("rds_backup_retention_period_seconds", "Automatic DB snapshots retention period", keys, nil)
		ch <- prometheus.MustNewConstMetric(backupRetentionPeriodDesc, prometheus.GaugeValue, float64(instance.BackupRetentionPeriod), values...)

		maxIops := instance.MaxIops
		storageThroughput := float64(instance.StorageThroughput)

		// RDS disk performance are limited by the EBS volume attached the RDS instance
		if ec2Metrics, ok := c.metrics.EC2.Instances[instance.DBInstanceClass]; ok {
			maxIops = min(instance.MaxIops, int64(ec2Metrics.BaselineIOPS))
			storageThroughput = min(float64(instance.StorageThroughput), ec2Metrics.BaselineThroughput)
		}

		if maxIops > 0 {
			maxIopsDesc := prometheus.NewDesc("rds_max_disk_iops_average", "Max disk IOPS evaluated with disk IOPS and EC2 capacity", keys, nil)
			ch <- prometheus.MustNewConstMetric(maxIopsDesc, prometheus.GaugeValue, float64(maxIops), values...)
		}
		if storageThroughput > 0 {
			storageThroughputDesc := prometheus.NewDesc("rds_max_storage_throughput_bytes", "Max disk throughput evaluated with disk throughput and EC2 capacity", keys, nil)
			ch <- prometheus.MustNewConstMetric(storageThroughputDesc, prometheus.GaugeValue, storageThroughput, values...)
		}

		// We still keep the instanceTags metric for backward compatibility
		if c.configuration.CollectInstanceTags {
			// For rds_instance_tags, always include tags regardless of IncludeTagsInMetrics setting
			labels := c.getBaseLabels(dbidentifier)
			tagKeys, tagValues := c.getTagLabels(instance)

			// Add tags to the labels map
			for i, key := range tagKeys {
				labels[key] = tagValues[i]
			}

			// Convert map to slices
			var allKeys []string
			var allValues []string
			for k, v := range labels {
				allKeys = append(allKeys, k)
				allValues = append(allValues, v)
			}

			instanceTagsDesc := prometheus.NewDesc("rds_instance_tags", "AWS tags attached to the instance", allKeys, nil)
			ch <- prometheus.MustNewConstMetric(instanceTagsDesc, prometheus.GaugeValue, 0, allValues...)
		}

		if instance.CertificateValidTill != nil {
			certificateValidTillDesc := prometheus.NewDesc("rds_certificate_expiry_timestamp_seconds", "Timestamp of the expiration of the Instance certificate", keys, nil)
			ch <- prometheus.MustNewConstMetric(certificateValidTillDesc, prometheus.GaugeValue, float64(instance.CertificateValidTill.Unix()), values...)
		}

		if instance.Age != nil {
			ageDesc := prometheus.NewDesc("rds_instance_age_seconds", "Time since instance creation", keys, nil)
			ch <- prometheus.MustNewConstMetric(ageDesc, prometheus.GaugeValue, *instance.Age, values...)
		}

		if instance.LogFilesSize != nil {
			logFilesSizeDesc := prometheus.NewDesc("rds_instance_log_files_size_bytes", "Total of log files on the instance", keys, nil)
			ch <- prometheus.MustNewConstMetric(logFilesSizeDesc, prometheus.GaugeValue, float64(*instance.LogFilesSize), values...)
		}
	}

	// Cloudwatch metrics
	ch <- prometheus.MustNewConstMetric(c.apiCall, prometheus.CounterValue, c.counters.CloudwatchAPICalls, c.awsAccountID, c.awsRegion, "cloudwatch")

	for dbidentifier, cloudwatchInstance := range c.metrics.CloudwatchInstances.Instances {
		// For cloudwatch metrics, we need to get the instance tags from the RDS instances
		rdsInstance, ok := c.metrics.RDS.Instances[dbidentifier]
		if !ok {
			// If we can't find the RDS instance, use the standard labels without tags
			if cloudwatchInstance.DatabaseConnections != nil {
				ch <- prometheus.MustNewConstMetric(c.databaseConnections, prometheus.GaugeValue, *cloudwatchInstance.DatabaseConnections, c.awsAccountID, c.awsRegion, dbidentifier)
			}
			continue
		}

		// Get tags from the RDS instance
		keys, values := c.getInstanceTagLabels(dbidentifier, rdsInstance)

		if cloudwatchInstance.DatabaseConnections != nil {
			databaseConnectionsDesc := prometheus.NewDesc("rds_database_connections_average", "The number of client network connections to the database instance", keys, nil)
			ch <- prometheus.MustNewConstMetric(databaseConnectionsDesc, prometheus.GaugeValue, *cloudwatchInstance.DatabaseConnections, values...)
		}

		if cloudwatchInstance.FreeStorageSpace != nil {
			freeStorageSpaceDesc := prometheus.NewDesc("rds_free_storage_bytes", "Free storage on the instance", keys, nil)
			ch <- prometheus.MustNewConstMetric(freeStorageSpaceDesc, prometheus.GaugeValue, *cloudwatchInstance.FreeStorageSpace, values...)
		}

		if cloudwatchInstance.FreeableMemory != nil {
			freeableMemoryDesc := prometheus.NewDesc("rds_freeable_memory_bytes", "Amount of available random access memory", keys, nil)
			ch <- prometheus.MustNewConstMetric(freeableMemoryDesc, prometheus.GaugeValue, *cloudwatchInstance.FreeableMemory, values...)
		}

		if cloudwatchInstance.MaximumUsedTransactionIDs != nil {
			maximumUsedTransactionIDsDesc := prometheus.NewDesc("rds_maximum_used_transaction_ids_average", "Maximum transaction IDs that have been used. Applies to only PostgreSQL", keys, nil)
			ch <- prometheus.MustNewConstMetric(maximumUsedTransactionIDsDesc, prometheus.GaugeValue, *cloudwatchInstance.MaximumUsedTransactionIDs, values...)
		}

		if cloudwatchInstance.ReadThroughput != nil {
			readThroughputDesc := prometheus.NewDesc("rds_read_throughput_bytes", "Average number of bytes read from disk per second", keys, nil)
			ch <- prometheus.MustNewConstMetric(readThroughputDesc, prometheus.GaugeValue, *cloudwatchInstance.ReadThroughput, values...)
		}

		if cloudwatchInstance.ReplicaLag != nil {
			replicaLagDesc := prometheus.NewDesc("rds_replica_lag_seconds", "For read replica configurations, the amount of time a read replica DB instance lags behind the source DB instance", keys, nil)
			ch <- prometheus.MustNewConstMetric(replicaLagDesc, prometheus.GaugeValue, *cloudwatchInstance.ReplicaLag, values...)
		}

		if cloudwatchInstance.ReplicationSlotDiskUsage != nil {
			replicationSlotDiskUsageDesc := prometheus.NewDesc("rds_replication_slot_disk_usage_bytes", "Disk space used by replication slot files. Applies to PostgreSQL", keys, nil)
			ch <- prometheus.MustNewConstMetric(replicationSlotDiskUsageDesc, prometheus.GaugeValue, *cloudwatchInstance.ReplicationSlotDiskUsage, values...)
		}

		if cloudwatchInstance.SwapUsage != nil {
			swapUsageDesc := prometheus.NewDesc("rds_swap_usage_bytes", "Amount of swap space used on the DB instance. This metric is not available for SQL Server", keys, nil)
			ch <- prometheus.MustNewConstMetric(swapUsageDesc, prometheus.GaugeValue, *cloudwatchInstance.SwapUsage, values...)
		}

		if cloudwatchInstance.ReadIOPS != nil {
			readIOPSDesc := prometheus.NewDesc("rds_read_iops_average", "Average number of disk read I/O operations per second", keys, nil)
			ch <- prometheus.MustNewConstMetric(readIOPSDesc, prometheus.GaugeValue, *cloudwatchInstance.ReadIOPS, values...)
		}

		if cloudwatchInstance.WriteIOPS != nil {
			writeIOPSDesc := prometheus.NewDesc("rds_write_iops_average", "Average number of disk write I/O operations per second", keys, nil)
			ch <- prometheus.MustNewConstMetric(writeIOPSDesc, prometheus.GaugeValue, *cloudwatchInstance.WriteIOPS, values...)
		}

		if cloudwatchInstance.WriteThroughput != nil {
			writeThroughputDesc := prometheus.NewDesc("rds_write_throughput_bytes", "Average number of bytes written to disk per second", keys, nil)
			ch <- prometheus.MustNewConstMetric(writeThroughputDesc, prometheus.GaugeValue, *cloudwatchInstance.WriteThroughput, values...)
		}

		if cloudwatchInstance.TransactionLogsDiskUsage != nil {
			transactionLogsDiskUsageDesc := prometheus.NewDesc("rds_transaction_logs_disk_usage_bytes", "Disk space used by transaction logs (only on PostgreSQL)", keys, nil)
			ch <- prometheus.MustNewConstMetric(transactionLogsDiskUsageDesc, prometheus.GaugeValue, *cloudwatchInstance.TransactionLogsDiskUsage, values...)
		}

		if cloudwatchInstance.DBLoad != nil {
			dbLoadDesc := prometheus.NewDesc("rds_dbload_average", "Number of active sessions for the DB engine", keys, nil)
			ch <- prometheus.MustNewConstMetric(dbLoadDesc, prometheus.GaugeValue, *cloudwatchInstance.DBLoad, values...)
		}

		if cloudwatchInstance.CPUUtilization != nil {
			cpuUtilisationDesc := prometheus.NewDesc("rds_cpu_usage_percent_average", "Instance CPU used", keys, nil)
			ch <- prometheus.MustNewConstMetric(cpuUtilisationDesc, prometheus.GaugeValue, *cloudwatchInstance.CPUUtilization, values...)
		}

		if cloudwatchInstance.DBLoadCPU != nil {
			dbLoadCPUDesc := prometheus.NewDesc("rds_dbload_cpu_average", "Number of active sessions where the wait event type is CPU", keys, nil)
			ch <- prometheus.MustNewConstMetric(dbLoadCPUDesc, prometheus.GaugeValue, *cloudwatchInstance.DBLoadCPU, values...)
		}

		if cloudwatchInstance.DBLoadNonCPU != nil {
			dbLoadNonCPUDesc := prometheus.NewDesc("rds_dbload_noncpu_average", "Number of active sessions where the wait event type is not CPU", keys, nil)
			ch <- prometheus.MustNewConstMetric(dbLoadNonCPUDesc, prometheus.GaugeValue, *cloudwatchInstance.DBLoadNonCPU, values...)
		}
	}

	// usage metrics
	if c.configuration.CollectUsages {
		ch <- prometheus.MustNewConstMetric(c.apiCall, prometheus.CounterValue, c.counters.UsageAPIcalls, c.awsAccountID, c.awsRegion, "usage")
		ch <- prometheus.MustNewConstMetric(c.usageAllocatedStorage, prometheus.GaugeValue, c.metrics.CloudWatchUsage.AllocatedStorage, c.awsAccountID, c.awsRegion)
		ch <- prometheus.MustNewConstMetric(c.usageDBInstances, prometheus.GaugeValue, c.metrics.CloudWatchUsage.DBInstances, c.awsAccountID, c.awsRegion)
		ch <- prometheus.MustNewConstMetric(c.usageManualSnapshots, prometheus.GaugeValue, c.metrics.CloudWatchUsage.ManualSnapshots, c.awsAccountID, c.awsRegion)
	}

	// EC2 metrics
	ch <- prometheus.MustNewConstMetric(c.apiCall, prometheus.CounterValue, c.counters.EC2APIcalls, c.awsAccountID, c.awsRegion, "ec2")
	for instanceType, instance := range c.metrics.EC2.Instances {
		// For EC2 metrics, we don't have tags directly associated, but we can add standard labels
		baseLabels := []string{"aws_account_id", "aws_region", "instance_class"}
		baseValues := []string{c.awsAccountID, c.awsRegion, instanceType}

		instanceBaselineIopsDesc := prometheus.NewDesc("rds_instance_baseline_iops_average", "Baseline IOPS of underlying EC2 instance class", baseLabels, nil)
		ch <- prometheus.MustNewConstMetric(instanceBaselineIopsDesc, prometheus.GaugeValue, float64(instance.BaselineIOPS), baseValues...)

		instanceBaselineThroughputDesc := prometheus.NewDesc("rds_instance_baseline_throughput_bytes", "Baseline throughput of underlying EC2 instance class", baseLabels, nil)
		ch <- prometheus.MustNewConstMetric(instanceBaselineThroughputDesc, prometheus.GaugeValue, instance.BaselineThroughput, baseValues...)

		instanceMaximumIopsDesc := prometheus.NewDesc("rds_instance_max_iops_average", "Maximum IOPS of underlying EC2 instance class", baseLabels, nil)
		ch <- prometheus.MustNewConstMetric(instanceMaximumIopsDesc, prometheus.GaugeValue, float64(instance.MaximumIops), baseValues...)

		instanceMaximumThroughputDesc := prometheus.NewDesc("rds_instance_max_throughput_bytes", "Maximum throughput of underlying EC2 instance class", baseLabels, nil)
		ch <- prometheus.MustNewConstMetric(instanceMaximumThroughputDesc, prometheus.GaugeValue, instance.MaximumThroughput, baseValues...)

		instanceMemoryDesc := prometheus.NewDesc("rds_instance_memory_bytes", "Instance class memory", baseLabels, nil)
		ch <- prometheus.MustNewConstMetric(instanceMemoryDesc, prometheus.GaugeValue, float64(instance.Memory), baseValues...)

		instanceVCPUDesc := prometheus.NewDesc("rds_instance_vcpu_average", "Total vCPU for this instance class", baseLabels, nil)
		ch <- prometheus.MustNewConstMetric(instanceVCPUDesc, prometheus.GaugeValue, float64(instance.Vcpu), baseValues...)
	}

	// serviceQuotas metrics
	if c.configuration.CollectQuotas {
		ch <- prometheus.MustNewConstMetric(c.apiCall, prometheus.CounterValue, c.counters.ServiceQuotasAPICalls, c.awsAccountID, c.awsRegion, "servicequotas")
		ch <- prometheus.MustNewConstMetric(c.quotaDBInstances, prometheus.GaugeValue, c.metrics.ServiceQuota.DBinstances, c.awsAccountID, c.awsRegion)
		ch <- prometheus.MustNewConstMetric(c.quotaTotalStorage, prometheus.GaugeValue, c.metrics.ServiceQuota.TotalStorage, c.awsAccountID, c.awsRegion)
		ch <- prometheus.MustNewConstMetric(c.quotaMaxDBInstanceSnapshots, prometheus.GaugeValue, c.metrics.ServiceQuota.ManualDBInstanceSnapshots, c.awsAccountID, c.awsRegion)
	}
}

func (c *rdsCollector) GetStatistics() counters {
	return c.counters
}

func (c *rdsCollector) GetMetrics() metrics {
	return c.metrics
}
