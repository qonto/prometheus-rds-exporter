// Package exporter implements Prometheus exporter
package exporter

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"sync"

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
	FilterInstances        []string
	CollectInstanceTags    bool
	CollectInstanceTypes   bool
	CollectLogsSize        bool
	CollectMaintenances    bool
	CollectQuotas          bool
	CollectUsages          bool
}

type counters struct {
	CloudwatchAPICalls    float64
	EC2APIcalls           float64
	Errors                float64
	RDSAPIcalls           float64
	ServiceQuotasAPICalls float64
	UsageAPIcalls         float64
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

func NewCollector(logger slog.Logger, collectorConfiguration Configuration, awsAccountID string, awsRegion string, rdsClient rdsClient, ec2Client EC2Client, cloudWatchClient cloudWatchClient, servicequotasClient servicequotasClient) *rdsCollector {
	return &rdsCollector{
		logger:              logger,
		awsAccountID:        awsAccountID,
		awsRegion:           awsRegion,
		rdsClient:           rdsClient,
		servicequotasClient: servicequotasClient,
		EC2Client:           ec2Client,
		cloudWatchClient:    cloudWatchClient,

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
			fmt.Sprintf("Instance status (%d: ok, %d: can't scrap metrics)", int(exporterUpStatusCode), int(exporterDownStatusCode)),
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
	c.logger.Info("get RDS metrics")

	rdsFetcher := rds.NewFetcher(c.ctx, c.rdsClient, rds.Configuration{
		CollectLogsSize:     c.configuration.CollectLogsSize,
		CollectMaintenances: c.configuration.CollectMaintenances,
		FilterInstances:     c.configuration.FilterInstances,
	})

	rdsMetrics, err := rdsFetcher.GetInstancesMetrics()
	if err != nil {
		return fmt.Errorf("can't fetch RDS metrics: %w", err)
	}

	c.metrics.RDS = rdsMetrics
	c.counters.RDSAPIcalls += rdsFetcher.GetStatistics().RdsAPICall
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

	fetcher := servicequotas.NewFetcher(ctx, client)

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

func (c *rdsCollector) getInstanceTagLabels(dbidentifier string, instance rds.RdsInstanceMetrics) (keys []string, values []string) {
	labels := map[string]string{
		"aws_account_id": c.awsAccountID,
		"aws_region":     c.awsRegion,
		"dbidentifier":   dbidentifier,
	}

	// Add instance tags to labels
	// Prefix label containing instance's tags with "tag_" prefix to avoid conflict with other labels
	for k, v := range instance.Tags {
		labelName := fmt.Sprintf("tag_%s", clearPrometheusLabel(k))
		labels[labelName] = v
	}

	for k, v := range labels {
		keys = append(keys, k)
		values = append(values, v)
	}

	return keys, values
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
	for dbidentifier, instance := range c.metrics.RDS.Instances {
		ch <- prometheus.MustNewConstMetric(
			c.allocatedStorage,
			prometheus.GaugeValue,
			float64(instance.AllocatedStorage),
			c.awsAccountID, c.awsRegion, dbidentifier,
		)
		ch <- prometheus.MustNewConstMetric(
			c.information,
			prometheus.GaugeValue,
			1,
			c.awsAccountID,
			c.awsRegion,
			dbidentifier,
			instance.DbiResourceID,
			instance.DBInstanceClass,
			instance.Engine,
			instance.EngineVersion,
			instance.StorageType,
			strconv.FormatBool(instance.MultiAZ),
			strconv.FormatBool(instance.DeletionProtection),
			instance.Role,
			instance.SourceDBInstanceIdentifier,
			strconv.FormatBool(instance.PendingModifiedValues),
			instance.PendingMaintenanceAction,
			strconv.FormatBool(instance.PerformanceInsightsEnabled),
			instance.CACertificateIdentifier,
			instance.Arn,
		)
		ch <- prometheus.MustNewConstMetric(c.maxAllocatedStorage, prometheus.GaugeValue, float64(instance.MaxAllocatedStorage), c.awsAccountID, c.awsRegion, dbidentifier)
		ch <- prometheus.MustNewConstMetric(c.allocatedDiskIOPS, prometheus.GaugeValue, float64(instance.MaxIops), c.awsAccountID, c.awsRegion, dbidentifier)
		ch <- prometheus.MustNewConstMetric(c.allocatedDiskThroughput, prometheus.GaugeValue, float64(instance.StorageThroughput), c.awsAccountID, c.awsRegion, dbidentifier)
		ch <- prometheus.MustNewConstMetric(c.status, prometheus.GaugeValue, float64(instance.Status), c.awsAccountID, c.awsRegion, dbidentifier)
		ch <- prometheus.MustNewConstMetric(c.backupRetentionPeriod, prometheus.GaugeValue, float64(instance.BackupRetentionPeriod), c.awsAccountID, c.awsRegion, dbidentifier)

		maxIops := instance.MaxIops
		storageThroughput := float64(instance.StorageThroughput)

		// RDS disk performance are limited by the EBS volume attached the RDS instance
		if ec2Metrics, ok := c.metrics.EC2.Instances[instance.DBInstanceClass]; ok {
			maxIops = min(instance.MaxIops, int64(ec2Metrics.BaselineIOPS))
			storageThroughput = min(float64(instance.StorageThroughput), ec2Metrics.BaselineThroughput)
		}

		ch <- prometheus.MustNewConstMetric(c.maxIops, prometheus.GaugeValue, float64(maxIops), c.awsAccountID, c.awsRegion, dbidentifier)
		ch <- prometheus.MustNewConstMetric(c.storageThroughput, prometheus.GaugeValue, storageThroughput, c.awsAccountID, c.awsRegion, dbidentifier)

		if c.configuration.CollectInstanceTags {
			names, values := c.getInstanceTagLabels(dbidentifier, instance)

			c.instanceTags = prometheus.NewDesc("rds_instance_tags", "AWS tags attached to the instance", names, nil)
			ch <- prometheus.MustNewConstMetric(c.instanceTags, prometheus.GaugeValue, 0, values...)
		}

		if instance.CertificateValidTill != nil {
			ch <- prometheus.MustNewConstMetric(c.certificateValidTill, prometheus.GaugeValue, float64(instance.CertificateValidTill.Unix()), c.awsAccountID, c.awsRegion, dbidentifier)
		}

		if instance.Age != nil {
			ch <- prometheus.MustNewConstMetric(c.age, prometheus.GaugeValue, *instance.Age, c.awsAccountID, c.awsRegion, dbidentifier)
		}

		if instance.LogFilesSize != nil {
			ch <- prometheus.MustNewConstMetric(c.logFilesSize, prometheus.GaugeValue, float64(*instance.LogFilesSize), c.awsAccountID, c.awsRegion, dbidentifier)
		}
	}

	// Cloudwatch metrics
	ch <- prometheus.MustNewConstMetric(c.apiCall, prometheus.CounterValue, c.counters.CloudwatchAPICalls, c.awsAccountID, c.awsRegion, "cloudwatch")

	for dbidentifier, instance := range c.metrics.CloudwatchInstances.Instances {
		if instance.DatabaseConnections != nil {
			ch <- prometheus.MustNewConstMetric(c.databaseConnections, prometheus.GaugeValue, *instance.DatabaseConnections, c.awsAccountID, c.awsRegion, dbidentifier)
		}

		if instance.FreeStorageSpace != nil {
			ch <- prometheus.MustNewConstMetric(c.freeStorageSpace, prometheus.GaugeValue, *instance.FreeStorageSpace, c.awsAccountID, c.awsRegion, dbidentifier)
		}

		if instance.FreeableMemory != nil {
			ch <- prometheus.MustNewConstMetric(c.freeableMemory, prometheus.GaugeValue, *instance.FreeableMemory, c.awsAccountID, c.awsRegion, dbidentifier)
		}

		if instance.MaximumUsedTransactionIDs != nil {
			ch <- prometheus.MustNewConstMetric(c.maximumUsedTransactionIDs, prometheus.GaugeValue, *instance.MaximumUsedTransactionIDs, c.awsAccountID, c.awsRegion, dbidentifier)
		}

		if instance.ReadThroughput != nil {
			ch <- prometheus.MustNewConstMetric(c.readThroughput, prometheus.GaugeValue, *instance.ReadThroughput, c.awsAccountID, c.awsRegion, dbidentifier)
		}

		if instance.ReplicaLag != nil {
			ch <- prometheus.MustNewConstMetric(c.replicaLag, prometheus.GaugeValue, *instance.ReplicaLag, c.awsAccountID, c.awsRegion, dbidentifier)
		}

		if instance.ReplicationSlotDiskUsage != nil {
			ch <- prometheus.MustNewConstMetric(c.replicationSlotDiskUsage, prometheus.GaugeValue, *instance.ReplicationSlotDiskUsage, c.awsAccountID, c.awsRegion, dbidentifier)
		}

		if instance.SwapUsage != nil {
			ch <- prometheus.MustNewConstMetric(c.swapUsage, prometheus.GaugeValue, *instance.SwapUsage, c.awsAccountID, c.awsRegion, dbidentifier)
		}

		if instance.ReadIOPS != nil {
			ch <- prometheus.MustNewConstMetric(c.readIOPS, prometheus.GaugeValue, *instance.ReadIOPS, c.awsAccountID, c.awsRegion, dbidentifier)
		}

		if instance.WriteIOPS != nil {
			ch <- prometheus.MustNewConstMetric(c.writeIOPS, prometheus.GaugeValue, *instance.WriteIOPS, c.awsAccountID, c.awsRegion, dbidentifier)
		}

		if instance.WriteThroughput != nil {
			ch <- prometheus.MustNewConstMetric(c.writeThroughput, prometheus.GaugeValue, *instance.WriteThroughput, c.awsAccountID, c.awsRegion, dbidentifier)
		}

		if instance.TransactionLogsDiskUsage != nil {
			ch <- prometheus.MustNewConstMetric(c.transactionLogsDiskUsage, prometheus.GaugeValue, *instance.TransactionLogsDiskUsage, c.awsAccountID, c.awsRegion, dbidentifier)
		}

		if instance.DBLoad != nil {
			ch <- prometheus.MustNewConstMetric(c.DBLoad, prometheus.GaugeValue, *instance.DBLoad, c.awsAccountID, c.awsRegion, dbidentifier)
		}

		if instance.CPUUtilization != nil {
			ch <- prometheus.MustNewConstMetric(c.cpuUtilisation, prometheus.GaugeValue, *instance.CPUUtilization, c.awsAccountID, c.awsRegion, dbidentifier)
		}

		if instance.DBLoadCPU != nil {
			ch <- prometheus.MustNewConstMetric(c.dBLoadCPU, prometheus.GaugeValue, *instance.DBLoadCPU, c.awsAccountID, c.awsRegion, dbidentifier)
		}

		if instance.DBLoadNonCPU != nil {
			ch <- prometheus.MustNewConstMetric(c.dBLoadNonCPU, prometheus.GaugeValue, *instance.DBLoadNonCPU, c.awsAccountID, c.awsRegion, dbidentifier)
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
		ch <- prometheus.MustNewConstMetric(c.instanceBaselineIops, prometheus.GaugeValue, float64(instance.BaselineIOPS), c.awsAccountID, c.awsRegion, instanceType)
		ch <- prometheus.MustNewConstMetric(c.instanceBaselineThroughput, prometheus.GaugeValue, instance.BaselineThroughput, c.awsAccountID, c.awsRegion, instanceType)
		ch <- prometheus.MustNewConstMetric(c.instanceMaximumIops, prometheus.GaugeValue, float64(instance.MaximumIops), c.awsAccountID, c.awsRegion, instanceType)
		ch <- prometheus.MustNewConstMetric(c.instanceMaximumThroughput, prometheus.GaugeValue, instance.MaximumThroughput, c.awsAccountID, c.awsRegion, instanceType)
		ch <- prometheus.MustNewConstMetric(c.instanceMemory, prometheus.GaugeValue, float64(instance.Memory), c.awsAccountID, c.awsRegion, instanceType)
		ch <- prometheus.MustNewConstMetric(c.instanceVCPU, prometheus.GaugeValue, float64(instance.Vcpu), c.awsAccountID, c.awsRegion, instanceType)
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
