// Package exporter implements Prometheus exporter
package exporter

import (
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
)

const (
	exporterUpStatusCode   float64 = 1
	exporterDownStatusCode float64 = 0
)

type counters struct {
	cloudwatchAPICalls    float64
	ec2APIcalls           float64
	errors                float64
	rdsAPIcalls           float64
	serviceQuotasAPICalls float64
	usageAPIcalls         float64
}

type metrics struct {
	serviceQuota        servicequotas.Metrics
	rds                 rds.Metrics
	ec2                 ec2.Metrics
	cloudwatchInstances cloudwatch.CloudWatchMetrics
	cloudWatchUsage     cloudwatch.UsageMetrics
}

type rdsCollector struct {
	wg       sync.WaitGroup
	logger   slog.Logger
	counters counters
	metrics  metrics

	rdsClient           rdsClient
	EC2Client           EC2Client
	servicequotasClient servicequotasClient
	cloudWatchClient    cloudWatchClient

	errors                      *prometheus.Desc
	DBLoad                      *prometheus.Desc
	dBLoadCPU                   *prometheus.Desc
	dBLoadNonCPU                *prometheus.Desc
	allocatedStorage            *prometheus.Desc
	information                 *prometheus.Desc
	instanceMaximumIops         *prometheus.Desc
	instanceMaximumThroughput   *prometheus.Desc
	instanceMemory              *prometheus.Desc
	instanceVCPU                *prometheus.Desc
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
}

func NewCollector(logger slog.Logger, rdsClient rdsClient, ec2Client EC2Client, cloudWatchClient cloudWatchClient, servicequotasClient servicequotasClient) *rdsCollector {
	return &rdsCollector{
		logger:              logger,
		rdsClient:           rdsClient,
		servicequotasClient: servicequotasClient,
		EC2Client:           ec2Client,
		cloudWatchClient:    cloudWatchClient,

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
			[]string{"dbidentifier"}, nil,
		),
		information: prometheus.NewDesc("rds_instance_info",
			"RDS instance information",
			[]string{"dbidentifier", "dbi_resource_id", "instance_class", "engine", "engine_version", "storage_type", "multi_az", "deletion_protection", "role", "source_dbidentifier", "pending_modified_values", "pending_maintenance"}, nil,
		),
		maxAllocatedStorage: prometheus.NewDesc("rds_max_allocated_storage_bytes",
			"Upper limit in gibibytes to which Amazon RDS can automatically scale the storage of the DB instance",
			[]string{"dbidentifier"}, nil,
		),
		maxIops: prometheus.NewDesc("rds_max_disk_iops_average",
			"Max IOPS for the instance",
			[]string{"dbidentifier"}, nil,
		),
		storageThroughput: prometheus.NewDesc("rds_max_storage_throughput_bytes",
			"Max storage throughput",
			[]string{"dbidentifier"}, nil,
		),
		readThroughput: prometheus.NewDesc("rds_read_throughput_bytes",
			"Average number of bytes read from disk per second",
			[]string{"dbidentifier"}, nil,
		),
		writeThroughput: prometheus.NewDesc("rds_write_throughput_bytes",
			"Average number of bytes written to disk per second",
			[]string{"dbidentifier"}, nil,
		),
		status: prometheus.NewDesc("rds_instance_status",
			fmt.Sprintf("Instance status (%d: ok, %d: can't scrap metrics)", int(exporterUpStatusCode), int(exporterDownStatusCode)),
			[]string{"dbidentifier"}, nil,
		),
		logFilesSize: prometheus.NewDesc("rds_instance_log_files_size_bytes",
			"Total of log files on the instance",
			[]string{"dbidentifier"}, nil,
		),
		instanceVCPU: prometheus.NewDesc("rds_instance_vcpu_average",
			"Total vCPU for this isntance class",
			[]string{"instance_class"}, nil,
		),
		instanceMemory: prometheus.NewDesc("rds_instance_memory_bytes",
			"Instance memory",
			[]string{"instance_class"}, nil,
		),
		cpuUtilisation: prometheus.NewDesc("rds_cpu_usage_percent_average",
			"Instance CPU used",
			[]string{"dbidentifier"}, nil,
		),
		instanceMaximumThroughput: prometheus.NewDesc("rds_instance_max_throughput_bytes",
			"Maximum throughput of underlying EC2 instance",
			[]string{"instance_class"}, nil,
		),
		instanceMaximumIops: prometheus.NewDesc("rds_instance_maxIops_average",
			"Maximum IOPS of underlying EC2 instance",
			[]string{"instance_class"}, nil,
		),
		freeStorageSpace: prometheus.NewDesc("rds_free_storage_bytes",
			"Free storage on the instance",
			[]string{"dbidentifier"}, nil,
		),
		databaseConnections: prometheus.NewDesc("rds_database_connections_average",
			"The number of client network connections to the database instance",
			[]string{"dbidentifier"}, nil,
		),
		up: prometheus.NewDesc("up",
			"Was the last scrape of RDS successful",
			nil, nil,
		),
		swapUsage: prometheus.NewDesc("rds_swap_usage_bytes",
			"Amount of swap space used on the DB instance. This metric is not available for SQL Server",
			[]string{"dbidentifier"}, nil,
		),
		writeIOPS: prometheus.NewDesc("rds_write_iops_average",
			"Average number of disk write I/O operations per second",
			[]string{"dbidentifier"}, nil,
		),
		readIOPS: prometheus.NewDesc("rds_read_iops_average",
			"Average number of disk read I/O operations per second",
			[]string{"dbidentifier"}, nil,
		),
		replicaLag: prometheus.NewDesc("rds_replica_lag_seconds",
			"For read replica configurations, the amount of time a read replica DB instance lags behind the source DB instance. Applies to MariaDB, Microsoft SQL Server, MySQL, Oracle, and PostgreSQL read replicas",
			[]string{"dbidentifier"}, nil,
		),
		replicationSlotDiskUsage: prometheus.NewDesc("rds_replication_slot_disk_usage_average",
			"Disk space used by replication slot files. Applies to PostgreSQL",
			[]string{"dbidentifier"}, nil,
		),
		maximumUsedTransactionIDs: prometheus.NewDesc("rds_maximum_used_transaction_ids_average",
			"Maximum transaction IDs that have been used. Applies to only PostgreSQL",
			[]string{"dbidentifier"}, nil,
		),
		freeableMemory: prometheus.NewDesc("rds_freeable_memory_bytes",
			"Amount of available random access memory. For MariaDB, MySQL, Oracle, and PostgreSQL DB instances, this metric reports the value of the MemAvailable field of /proc/meminfo",
			[]string{"dbidentifier"}, nil,
		),
		apiCall: prometheus.NewDesc("rds_api_call_total",
			"Number of call to AWS API",
			[]string{"api"}, nil,
		),
		backupRetentionPeriod: prometheus.NewDesc("rds_backup_retention_period_seconds",
			"Automatic DB snapshots retention period",
			[]string{"dbidentifier"}, nil,
		),
		DBLoad: prometheus.NewDesc("rds_dbload_average",
			"Number of active sessions for the DB engine",
			[]string{"dbidentifier"}, nil,
		),
		dBLoadCPU: prometheus.NewDesc("rds_dbload_cpu_average",
			"Number of active sessions where the wait event type is CPU",
			[]string{"dbidentifier"}, nil,
		),
		dBLoadNonCPU: prometheus.NewDesc("rds_dbload_noncpu_average",
			"Number of active sessions where the wait event type is not CPU",
			[]string{"dbidentifier"}, nil,
		),
		quotaDBInstances: prometheus.NewDesc("rds_quota_max_dbinstances_average",
			"Maximum number of RDS instances allowed in the AWS account",
			nil, nil,
		),
		quotaTotalStorage: prometheus.NewDesc("rds_quota_total_storage_bytes",
			"Maximum total storage for all DB instances",
			nil, nil,
		),
		quotaMaxDBInstanceSnapshots: prometheus.NewDesc("rds_quota_maximum_db_instance_snapshots_average",
			"Maximum number of manual DB instance snapshots",
			nil, nil,
		),
		usageAllocatedStorage: prometheus.NewDesc("rds_usage_allocated_storage_average",
			"Total storage used by AWS RDS instances",
			nil, nil,
		),
		usageDBInstances: prometheus.NewDesc("rds_usage_db_instances_average",
			"AWS RDS instance count",
			nil, nil,
		),
		usageManualSnapshots: prometheus.NewDesc("rds_usage_manual_snapshots_average",
			"Manual snapshots count",
			nil, nil,
		),
	}
}

func (c *rdsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.status
	ch <- c.up
}

// getMetrics collects and return all RDS metrics
func (c *rdsCollector) fetchMetrics() error {
	c.logger.Debug("received query")

	// Fetch serviceQuotas metrics
	go c.getQuotasMetrics(c.servicequotasClient)
	c.wg.Add(1)

	// Fetch usages metrics
	go c.getUsagesMetrics(c.cloudWatchClient)
	c.wg.Add(1)

	// Fetch RDS instances metrics
	c.logger.Info("get RDS metrics")

	rdsFetcher := rds.NewFetcher(c.rdsClient)

	rdsMetrics, err := rdsFetcher.GetInstancesMetrics()
	if err != nil {
		return fmt.Errorf("can't fetch RDS metrics: %w", err)
	}

	c.metrics.rds = rdsMetrics
	c.counters.rdsAPIcalls += rdsFetcher.GetStatistics().RdsAPICall
	c.logger.Debug("RDS metrics fetched")

	// Compute uniq instances identifiers and instance types
	instanceIdentifiers, instanceTypes := getUniqTypeAndIdentifiers(rdsMetrics.Instances)

	// Fetch EC2 Metrics for instance types
	if len(instanceTypes) > 0 {
		go c.getEC2Metrics(c.EC2Client, instanceTypes)
		c.wg.Add(1)
	}

	// Fetch Cloudwatch metrics for instances
	go c.getCloudwatchMetrics(c.cloudWatchClient, instanceIdentifiers)
	c.wg.Add(1)

	// Wait for all go routines to finish
	c.wg.Wait()

	return nil
}

func (c *rdsCollector) getCloudwatchMetrics(client cloudwatch.CloudWatchClient, instanceIdentifiers []string) {
	defer c.wg.Done()
	c.logger.Debug("fetch cloudwatch metrics")

	fetcher := cloudwatch.NewRDSFetcher(client, c.logger)

	metrics, err := fetcher.GetRDSInstanceMetrics(instanceIdentifiers)
	if err != nil {
		c.counters.errors++
	}

	c.counters.cloudwatchAPICalls += fetcher.GetStatistics().CloudWatchAPICall
	c.metrics.cloudwatchInstances = metrics

	c.logger.Debug("cloudwatch metrics fetched", "metrics", metrics)
}

func (c *rdsCollector) getUsagesMetrics(client cloudwatch.CloudWatchClient) {
	defer c.wg.Done()
	c.logger.Debug("fetch usage metrics")

	fetcher := cloudwatch.NewUsageFetcher(client, c.logger)

	metrics, err := fetcher.GetUsageMetrics()
	if err != nil {
		c.counters.errors++
		c.logger.Error(fmt.Sprintf("can't fetch usage metrics: %s", err))
	}

	c.counters.usageAPIcalls += fetcher.GetStatistics().CloudWatchAPICall
	c.metrics.cloudWatchUsage = metrics

	c.logger.Debug("usage metrics fetched", "metrics", metrics)
}

func (c *rdsCollector) getEC2Metrics(client ec2.EC2Client, instanceTypes []string) {
	defer c.wg.Done()
	c.logger.Debug("fetch EC2 metrics")

	fetcher := ec2.NewFetcher(client)

	metrics, err := fetcher.GetDBInstanceTypeInformation(instanceTypes)
	if err != nil {
		c.counters.errors++
		c.logger.Error(fmt.Sprintf("can't fetch EC2 metrics: %s", err))
	}

	c.counters.ec2APIcalls += fetcher.GetStatistics().EC2ApiCall
	c.metrics.ec2 = metrics

	c.logger.Debug("EC2 metrics fetched", "metrics", metrics)
}

func (c *rdsCollector) getQuotasMetrics(client servicequotas.ServiceQuotasClient) {
	defer c.wg.Done()
	c.logger.Debug("fetch quotas")

	fetcher := servicequotas.NewFetcher(client)

	metrics, err := fetcher.GetRDSQuotas()
	if err != nil {
		c.counters.errors++
		c.logger.Error(fmt.Sprintf("can't fetch service quota metrics: %s", err))
	}

	c.counters.serviceQuotasAPICalls += fetcher.GetStatistics().UsageAPICall
	c.metrics.serviceQuota = metrics

	c.logger.Debug("quota metrics fetched", "metrics", metrics)
}

func (c *rdsCollector) Collect(ch chan<- prometheus.Metric) {
	ch <- prometheus.MustNewConstMetric(c.exporterBuildInformation, prometheus.GaugeValue, 1, build.Version, build.CommitSHA, build.Date)
	ch <- prometheus.MustNewConstMetric(c.errors, prometheus.CounterValue, c.counters.errors)

	// Get all metrics
	err := c.fetchMetrics()
	if err != nil {
		c.logger.Error(fmt.Sprintf("can't scrape metrics: %s", err))
		// Mark exporter as down
		ch <- prometheus.MustNewConstMetric(c.up, prometheus.CounterValue, exporterDownStatusCode)

		return
	}
	ch <- prometheus.MustNewConstMetric(c.up, prometheus.CounterValue, exporterUpStatusCode)

	// RDS metrics
	ch <- prometheus.MustNewConstMetric(c.apiCall, prometheus.CounterValue, c.counters.rdsAPIcalls, "rds")
	for dbidentifier, instance := range c.metrics.rds.Instances {
		ch <- prometheus.MustNewConstMetric(c.allocatedStorage, prometheus.GaugeValue, float64(instance.AllocatedStorage), dbidentifier)
		ch <- prometheus.MustNewConstMetric(c.information, prometheus.GaugeValue, 1, dbidentifier, instance.DbiResourceID, instance.DBInstanceClass, instance.Engine, instance.EngineVersion, instance.StorageType, strconv.FormatBool(instance.MultiAZ), strconv.FormatBool(instance.DeletionProtection), instance.Role, instance.SourceDBInstanceIdentifier, strconv.FormatBool(instance.PendingModifiedValues), instance.PendingMaintenanceAction)
		ch <- prometheus.MustNewConstMetric(c.logFilesSize, prometheus.GaugeValue, float64(instance.LogFilesSize), dbidentifier)
		ch <- prometheus.MustNewConstMetric(c.maxAllocatedStorage, prometheus.GaugeValue, float64(instance.MaxAllocatedStorage), dbidentifier)
		ch <- prometheus.MustNewConstMetric(c.maxIops, prometheus.GaugeValue, float64(instance.MaxIops), dbidentifier)
		ch <- prometheus.MustNewConstMetric(c.status, prometheus.GaugeValue, float64(instance.Status), dbidentifier)
		ch <- prometheus.MustNewConstMetric(c.storageThroughput, prometheus.GaugeValue, float64(instance.StorageThroughput), dbidentifier)
		ch <- prometheus.MustNewConstMetric(c.backupRetentionPeriod, prometheus.GaugeValue, float64(instance.BackupRetentionPeriod), dbidentifier)
	}

	// Cloudwatch metrics
	ch <- prometheus.MustNewConstMetric(c.apiCall, prometheus.CounterValue, c.counters.cloudwatchAPICalls, "cloudwatch")
	for dbidentifier, instance := range c.metrics.cloudwatchInstances.Instances {
		ch <- prometheus.MustNewConstMetric(c.cpuUtilisation, prometheus.GaugeValue, instance.CPUUtilization, dbidentifier)
		ch <- prometheus.MustNewConstMetric(c.databaseConnections, prometheus.GaugeValue, instance.DatabaseConnections, dbidentifier)
		ch <- prometheus.MustNewConstMetric(c.freeStorageSpace, prometheus.GaugeValue, instance.FreeStorageSpace, dbidentifier)
		ch <- prometheus.MustNewConstMetric(c.freeableMemory, prometheus.GaugeValue, instance.FreeableMemory, dbidentifier)
		ch <- prometheus.MustNewConstMetric(c.maximumUsedTransactionIDs, prometheus.GaugeValue, instance.MaximumUsedTransactionIDs, dbidentifier)
		ch <- prometheus.MustNewConstMetric(c.readIOPS, prometheus.GaugeValue, instance.ReadIOPS, dbidentifier)
		ch <- prometheus.MustNewConstMetric(c.readThroughput, prometheus.GaugeValue, instance.ReadThroughput, dbidentifier)
		ch <- prometheus.MustNewConstMetric(c.replicaLag, prometheus.GaugeValue, instance.ReplicaLag, dbidentifier)
		ch <- prometheus.MustNewConstMetric(c.replicationSlotDiskUsage, prometheus.GaugeValue, instance.ReplicationSlotDiskUsage, dbidentifier)
		ch <- prometheus.MustNewConstMetric(c.swapUsage, prometheus.GaugeValue, instance.SwapUsage, dbidentifier)
		ch <- prometheus.MustNewConstMetric(c.writeIOPS, prometheus.GaugeValue, instance.WriteIOPS, dbidentifier)
		ch <- prometheus.MustNewConstMetric(c.writeThroughput, prometheus.GaugeValue, instance.WriteThroughput, dbidentifier)
		ch <- prometheus.MustNewConstMetric(c.DBLoad, prometheus.GaugeValue, instance.DBLoad, dbidentifier)
		ch <- prometheus.MustNewConstMetric(c.dBLoadCPU, prometheus.GaugeValue, instance.DBLoadCPU, dbidentifier)
		ch <- prometheus.MustNewConstMetric(c.dBLoadNonCPU, prometheus.GaugeValue, instance.DBLoadNonCPU, dbidentifier)
	}

	// usage metrics
	ch <- prometheus.MustNewConstMetric(c.apiCall, prometheus.CounterValue, c.counters.usageAPIcalls, "usage")
	ch <- prometheus.MustNewConstMetric(c.usageAllocatedStorage, prometheus.GaugeValue, c.metrics.cloudWatchUsage.AllocatedStorage)
	ch <- prometheus.MustNewConstMetric(c.usageDBInstances, prometheus.GaugeValue, c.metrics.cloudWatchUsage.DBInstances)
	ch <- prometheus.MustNewConstMetric(c.usageManualSnapshots, prometheus.GaugeValue, c.metrics.cloudWatchUsage.ManualSnapshots)

	// EC2 metrics
	ch <- prometheus.MustNewConstMetric(c.apiCall, prometheus.CounterValue, c.counters.ec2APIcalls, "ec2")
	for instanceType, instance := range c.metrics.ec2.Instances {
		ch <- prometheus.MustNewConstMetric(c.instanceMaximumIops, prometheus.GaugeValue, float64(instance.MaximumIops), instanceType)
		ch <- prometheus.MustNewConstMetric(c.instanceMaximumThroughput, prometheus.GaugeValue, instance.MaximumThroughput, instanceType)
		ch <- prometheus.MustNewConstMetric(c.instanceMemory, prometheus.GaugeValue, float64(instance.Memory), instanceType)
		ch <- prometheus.MustNewConstMetric(c.instanceVCPU, prometheus.GaugeValue, float64(instance.Vcpu), instanceType)
	}

	// serviceQuotas metrics
	ch <- prometheus.MustNewConstMetric(c.apiCall, prometheus.CounterValue, c.counters.serviceQuotasAPICalls, "servicequotas")
	ch <- prometheus.MustNewConstMetric(c.quotaDBInstances, prometheus.GaugeValue, c.metrics.serviceQuota.DBinstances)
	ch <- prometheus.MustNewConstMetric(c.quotaTotalStorage, prometheus.GaugeValue, c.metrics.serviceQuota.TotalStorage)
	ch <- prometheus.MustNewConstMetric(c.quotaMaxDBInstanceSnapshots, prometheus.GaugeValue, c.metrics.serviceQuota.ManualDBInstanceSnapshots)
}
