local common = import '../common.libsonnet';
local g = import '../g.libsonnet';
local queries = import '../queries/instance.libsonnet';
local generic = import 'generic.libsonnet';
local colors = common.colors;

{
  gauge: {
    local g = generic.gauge,

    cpu:
      g.percent('CPU', 'Current CPU usage', queries.instance.cpu.usagePercent),

    diskIOPS:
      g.percent('Disk IOPS', 'Current disk IOPS usage', queries.instance.disk.iops.usagePercent),

    memory:
      g.percent('Memory', 'Curent memory usage', queries.instance.memory.usagePercent),

    storage:
      g.percent('Storage', 'Current disk storage usage', queries.instance.storage.usagePercent),
  },
  stat: {
    local stat = g.panel.stat,
    local s = generic.stat,
    local standardOptions = stat.standardOptions,
    local step = standardOptions.threshold.step,

    age:
      s.base('Age', 'Time since RDS instance creation', queries.instance.age)
      + standardOptions.withUnit('s')
      + standardOptions.withDecimals(0),

    ca:
      s.field('CA certificate identifier', 'Certificate Authority bundle used by this RDS instance for TLS connections', queries.instance.info, 'ca_certificate_identifier'),

    dbi:
      s.field('DBI resource ID', 'AWS RDS instance internal reference', queries.instance.info, 'dbi_resource_id'),

    engine:
      s.field('Engine', 'RDS Engine', queries.instance.info, 'engine'),

    engineVersion:
      s.field('Engine Version', 'RDS Engine version', queries.instance.info, 'engine_version'),

    instanceClass:
      s.field('Instance class', 'RDS instance class (db.<family>.<instance size>)', queries.instance.info, 'instance_class'),

    memory:
      s.base('Memory', 'Total memory', queries.instance.memory.max)
      + standardOptions.withUnit('bytes')
      + standardOptions.withDecimals(0),

    multiAZ:
      s.enabledOrDisabled('Multi AZ', 'Multi Availability Zone deployment', [queries.instance.info]),

    pendingMaintenance:
      s.__stat('Pending maintenance', 'AWS may required hardward or software maintenance. Maintenance could be scheduled during maintenance windows or triggered manually.', queries.instance.info)
      + stat.options.reduceOptions.withFields('pending_maintenance')
      + standardOptions.withMappings([
        standardOptions.mapping.ValueMap.withType('value')
        + standardOptions.mapping.ValueMap.withOptions({
          no: { index: 0, color: colors.ok, text: 'No' },
          pending: { index: 0, color: colors.notice, text: 'Pending' },
          'auto-applied': { index: 0, color: colors.warning, text: 'Auto-applied' },
          forced: { index: 1, color: colors.danger, text: 'Forced' },
        }),
      ]),

    pendingModification:
      s.__stat('Pending modification', 'Yes when RDS instance need a modification (e.g. Apply new parameter group settings). Modification could be scheduled during maintenance windows or triggered manually.', queries.instance.info)
      + stat.options.reduceOptions.withFields('pending_modified_values')
      + standardOptions.withMappings([
        standardOptions.mapping.ValueMap.withType('value')
        + standardOptions.mapping.ValueMap.withOptions({
          'true': { index: 0, color: colors.notice, text: 'Yes' },
          'false': { index: 1, color: colors.ok, text: 'No' },
        }),
      ]),

    replicas:
      s.base('Replicas', 'Number of RDS instances replicating this instance', [queries.instance.replicas.count]),

    role:
      s.field('Role', 'If instance is part of a DB cluster (Aurora or RDS Multi-AZ DB cluster): `writer` or `reader`. Otherwise role is `primary` for standalone DB instance and `replica` for an instance configured as replica.', queries.instance.info, 'role'),

    lag:
      s.lag('Replication lag', 'Replication lag for replica RDS instance', queries.instance.replicas.lag),

    snapshotRetention:
      s.base('Snapshot retention', 'AWS RDS snapshot retention. This is not related to AWS Backup service', [queries.instance.backup.retention])
      + standardOptions.withUnit('s')
      + standardOptions.withDecimals(0)
      + standardOptions.thresholds.withSteps([
        step.withValue(0) + step.withColor(colors.notice),
        step.withValue(1) + step.withColor(colors.ok),
      ])
      + standardOptions.withMappings([
        standardOptions.mapping.ValueMap.withType('value')
        + standardOptions.mapping.ValueMap.withOptions({
          '0': { index: 0, color: colors.notice, text: 'Disabled' },
        }),
      ]),

    engineStandardSupportRemainingDays:
      s.base('Standard support remaining days', 'Days before end of standard support. After the standard support period ends, AWS will continue providing security patches and critical fixes for that engine version during the Extended Support window to give you more time to upgrade. But an additional Extended Support cost will be applied until you upgrade. See https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/extended-support.html', [queries.instance.engineSupport.standardRemainingDays])
      + standardOptions.withUnit('d')
      + standardOptions.withDecimals(0)
      + standardOptions.thresholds.withSteps([
        step.withValue(0) + step.withColor(colors.danger),
        step.withValue(30) + step.withColor(colors.warning),
        step.withValue(90) + step.withColor(colors.ok),
      ])
      + standardOptions.withMappings([
        standardOptions.mapping.RangeMap.withType('range')
        + standardOptions.mapping.RangeMap.withOptions({
          from: -1825, /* 5 years */
          to: 0,
          result: { index: 0, color: colors.danger, text: 'Expired - instance in extended support' },
        }),
      ]),

    engineExtendedSupportRemainingDays:
      s.base('Extended support remaining days', 'Days before extended support end. RDS Extended Support is available for up to 3 years past the RDS end of standard support date for a major engine version. After this time, if instance has not been upgraded to major engine version to a supported version, then Amazon RDS will automatically upgrade your major engine version. We recommend that you upgrade to a supported major engine version as soon as possible. See https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/extended-support.html', [queries.instance.engineSupport.extendedRemainingDays])
      + standardOptions.withUnit('d')
      + standardOptions.withDecimals(0)
      + standardOptions.thresholds.withSteps([
        step.withValue(0) + step.withColor(colors.danger),
        step.withValue(30) + step.withColor(colors.warning),
        step.withValue(90) + step.withColor(colors.ok),
      ]),

    source:
      s.field('Source database', 'Primary instance name for a replica)', queries.instance.info, 'source_dbidentifier'),

    storage:
      s.bytes('Allocated storage', 'Total disk storage', queries.instance.storage.allocated),

    storageType:
      s.field('Storage type', 'Storage class type)', queries.instance.info, 'storage_type'),

    vCPU:
      s.base('vCPU', 'Total number of vCPU', queries.instance.cpu.count),

    status:
      s.__stat('Instance status', 'Current RDS instance status', queries.instance.status)
      + standardOptions.withMappings([
        standardOptions.mapping.ValueMap.withType('value')
        + standardOptions.mapping.ValueMap.withOptions({
          '-10': { index: 0, color: colors.danger, text: 'Restore error' },
          '-9': { index: 1, color: colors.warning, text: 'Maintenance' },
          '-8': { index: 2, color: colors.info, text: 'Upgrading' },
          '-7': { index: 3, color: colors.danger, text: 'Storage-full' },
          '-6': { index: 4, color: colors.danger, text: 'Failed' },
          '-5': { index: 5, color: colors.warning, text: 'Rebooting' },
          '-4': { index: 6, color: colors.danger, text: 'Deleting' },
          '-3': { index: 7, color: colors.notice, text: 'Creating' },
          '-2': { index: 8, color: colors.notice, text: 'Stopping' },
          '-1': { index: 9, color: 'purple', text: 'Unknown' },
          '0': { index: 10, color: colors.danger, text: 'Stopped' },
          '1': { index: 11, color: colors.ok, text: 'Available' },
          '2': { index: 12, color: colors.notice, text: 'Backing-up' },
          '3': { index: 13, color: colors.ok, text: 'Starting' },
          '4': { index: 14, color: colors.info, text: 'Modifying' },
          '5': { index: 15, color: colors.info, text: 'Configuring enhanced monitoring' },
          '10': { index: 16, colors: colors.info, text: 'Storage initialization' },
          '11': { index: 17, colors: colors.notice, text: 'Storage optimization' },
          '20': { index: 18, colors: colors.notice, text: 'Renaming' },
        }),
      ]),
  },
  timeSeries: {
    local ts = generic.timeSeries,
    local timeSeries = g.panel.timeSeries,
    local fieldOverride = g.panel.timeSeries.fieldOverride,
    local options = timeSeries.options,
    local custom = timeSeries.fieldConfig.defaults.custom,
    local standardOptions = timeSeries.standardOptions,
    local color = standardOptions.color,

    autoscalingUsage:
      ts.percent('RDS storage autoscaling usage', 'If RDS storage autoscaling is enabled, display the current usage', [queries.instance.autoscalingUsage])
      + ts.singleMetric
      + standardOptions.withNoValue('Disabled'),

    activeQueries:
      ts.base('Average Active Sessions', "A session is active when it's either running on CPU or waiting for a resource to become available so that it can proceed (e.g. IOPS or CPU). For optimal performances, you should not have more AAS than the total number of vCPU. Investigate AAS in RDS performance insights. See also https://www.kylehailey.com/post/setting-the-record-straight-a-comprehensive-guide-to-understanding-the-aas-metric-in-databases and https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_PerfInsights.Overview.ActiveSessions.html", [queries.instance.cpu.wait, queries.instance.cpu.nonWait, queries.instance.cpu.count])
      + standardOptions.withDecimals(1)
      + custom.stacking.withMode('normal')
      + standardOptions.withOverrides([
        fieldOverride.byName.new('Number of vCPU')
        + ts.max,
        fieldOverride.byName.new('CPU execution')
        + standardOptions.override.byType.withPropertiesFromOptions(
          color.withMode('fixed')
          + color.withFixedColor(colors.ok)
        ),
        fieldOverride.byName.new('Non CPU execution')
        + standardOptions.override.byType.withPropertiesFromOptions(
          color.withMode('fixed')
          + color.withFixedColor(colors.warning)
        ),
      ]),

    diskIOPSScaling:
      ts.base('Disk IOPS', 'The RDS instance cannot use more disk IOPS than supported by the EC2 instance baseline, but it can burst 30 minutes at least once every 24 hours. See https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ebs-optimized.html', [queries.instance.disk.iops.usage, queries.instance.disk.iops.allocated, queries.instance.disk.iops.instanceTypeBaseline, queries.instance.disk.iops.instanceTypeBurst])
      + options.legend.withSortBy('Max')
      + options.legend.withSortDesc(true)
      + standardOptions.withUnit('locale')
      + standardOptions.withOverrides([
        fieldOverride.byName.new('Allocated')
        + standardOptions.override.byType.withPropertiesFromOptions(
          color.withMode('fixed')
          + color.withFixedColor(colors.warning)
          + custom.withFillOpacity(0)
        ),
        fieldOverride.byRegexp.new('.* burst')
        + standardOptions.override.byType.withPropertiesFromOptions(
          timeSeries.fieldConfig.defaults.custom.lineStyle.withDash([10, 10])
          + timeSeries.fieldConfig.defaults.custom.lineStyle.withFill('dash')
          + color.withMode('fixed')
          + color.withFixedColor(colors.notice)
          + custom.withFillOpacity(0)
        ),
        fieldOverride.byRegexp.new('.* baseline')
        + standardOptions.override.byType.withPropertiesFromOptions(
          color.withMode('fixed')
          + color.withFixedColor(colors.limit)
          + custom.withFillOpacity(0)
        ),
      ]),

    diskThroughputScaling:
      ts.base('Disk throughput', 'The RDS instance cannot use more disk throughput than supported by the EC2 instance baseline, but it can burst 30 minutes at least once every 24 hours. See https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ebs-optimized.html', [queries.instance.disk.throughput.usage, queries.instance.disk.throughput.allocated, queries.instance.disk.throughput.instanceTypeBaseline, queries.instance.disk.throughput.instanceTypeBurst])
      + options.legend.withSortBy('Max')
      + options.legend.withSortDesc(true)
      + standardOptions.withUnit('bytes')
      + standardOptions.withOverrides([
        fieldOverride.byName.new('Allocated')
        + standardOptions.override.byType.withPropertiesFromOptions(
          color.withMode('fixed')
          + color.withFixedColor(colors.warning)
          + custom.withFillOpacity(0)
        ),
        fieldOverride.byRegexp.new('.* burst')
        + standardOptions.override.byType.withPropertiesFromOptions(
          timeSeries.fieldConfig.defaults.custom.lineStyle.withDash([10, 10])
          + timeSeries.fieldConfig.defaults.custom.lineStyle.withFill('dash')
          + color.withMode('fixed')
          + color.withFixedColor(colors.limit)
          + custom.withFillOpacity(0)
        ),
        fieldOverride.byRegexp.new('.* baseline')
        + standardOptions.override.byType.withPropertiesFromOptions(
          color.withMode('fixed')
          + color.withFixedColor(colors.notice)
          + custom.withFillOpacity(0)
        ),
      ]),

    postgresqlMaxUsedTransaction:
      ts.single('Used Transactions', 'The number of transactions currently used. Hard limit to 2 billion transactions (See Transaction ID Wraparound)', [queries.postgresql.maxUsedTransaction])
      + standardOptions.withUnit('locale')
      + standardOptions.thresholds.withMode('absolute')
      + standardOptions.thresholds.withSteps([
        standardOptions.threshold.step.withColor(colors.transparent)
        + standardOptions.threshold.step.withValue(0),
        standardOptions.threshold.step.withColor(colors.warning)
        + standardOptions.threshold.step.withValue(1600000000),
        standardOptions.threshold.step.withColor(colors.danger)
        + standardOptions.threshold.step.withValue(2000000000),
      ])
      + custom.withThresholdsStyle({
        mode: 'line+area',
      })
      + standardOptions.withMax(2000000000),

    replicasLag:
      ts.base('Replication lag', 'Lag of PostgreSQL replica instances', [queries.instance.replicasLag])
      + standardOptions.withUnit('s'),

    cpu:
      ts.percent('CPU usage', 'Ratio of CPU usage', [queries.instance.cpu.usagePercent])
      + ts.singleMetric,

    databaseConnections:
      ts.base('Database connections', 'The number of client network connections to the database instance', [queries.instance.databaseConnections])
      + ts.singleMetric,

    diskIOPS:
      ts.base('Disk IOPS usage', 'Total of read and write disk IOPS regarding RDS instance IOPS limits. For optimal performances, you should not reach IOPS limits', [queries.instance.disk.iops.max, queries.instance.disk.iops.read, queries.instance.disk.iops.write])
      + standardOptions.withOverrides([
        fieldOverride.byName.new('Max')
        + standardOptions.override.byType.withPropertiesFromOptions(
          standardOptions.withUnit('locale')
          + color.withMode('fixed')
          + color.withFixedColor(colors.danger)
          + custom.withFillOpacity(0)
        ),
        fieldOverride.byName.new('Read')
        + standardOptions.override.byType.withPropertiesFromOptions(
          standardOptions.withUnit('locale')
          + color.withMode('fixed')
          + color.withFixedColor(colors.ok)
          + custom.stacking.withMode('normal')
        ),
        fieldOverride.byName.new('Write')
        + standardOptions.override.byType.withPropertiesFromOptions(
          standardOptions.withUnit('locale')
          + color.withMode('fixed')
          + color.withFixedColor(colors.notice)
          + custom.stacking.withMode('normal')
        ),
      ]),

    diskThroughput:
      ts.base('Disk throughput', 'The average number of bytes read/write from disk per second regarding RDS instance disk throughput limits. For optimal performances, you should not reach disk throughput', [queries.instance.disk.throughput.read, queries.instance.disk.throughput.write, queries.instance.disk.throughput.max])
      + standardOptions.withDecimals(0)
      + standardOptions.withUnit('bytes')
      + standardOptions.withOverrides([
        fieldOverride.byName.new('Max')
        + standardOptions.override.byType.withPropertiesFromOptions(
          standardOptions.withUnit('bytes')
          + custom.withLineWidth('2')
          + color.withMode('fixed')
          + color.withFixedColor(colors.danger)
          + custom.withFillOpacity(0)
        ),
        fieldOverride.byName.new('Read')
        + standardOptions.override.byType.withPropertiesFromOptions(
          standardOptions.withUnit('bytes')
          + color.withMode('fixed')
          + color.withFixedColor(colors.ok)
          + custom.stacking.withMode('normal')
        ),
        fieldOverride.byName.new('Write')
        + standardOptions.override.byType.withPropertiesFromOptions(
          standardOptions.withUnit('bytes')
          + color.withMode('fixed')
          + color.withFixedColor(colors.notice)
          + custom.stacking.withMode('normal')
        ),
      ]),

    networkThroughput:
      ts.base('Network throughput', 'The average number of bytes received/transmitted per second from/to the network. Storage network metrics are specific to Aurora storage subsystem.', [queries.instance.network.throughput.receive, queries.instance.network.throughput.transmit, queries.instance.network.throughput.max, queries.instance.network.throughput.storageReceive, queries.instance.network.throughput.storageTransmit])
      + standardOptions.withDecimals(0)
      + standardOptions.withUnit('decbytes')
      + standardOptions.withOverrides([
        fieldOverride.byName.new('Max')
        + standardOptions.override.byType.withPropertiesFromOptions(
          standardOptions.withUnit('decbytes')
          + custom.withLineWidth('2')
          + color.withMode('fixed')
          + color.withFixedColor(colors.danger)
          + custom.withFillOpacity(0)
        ),
        fieldOverride.byName.new('Receive')
        + standardOptions.override.byType.withPropertiesFromOptions(
          standardOptions.withUnit('decbytes')
          + color.withMode('fixed')
          + color.withFixedColor(colors.ok)
          + custom.stacking.withMode('normal')
        ),
        fieldOverride.byName.new('Transmit')
        + standardOptions.override.byType.withPropertiesFromOptions(
          standardOptions.withUnit('decbytes')
          + color.withMode('fixed')
          + color.withFixedColor(colors.notice)
          + custom.stacking.withMode('normal')
        ),
        fieldOverride.byName.new('Storage: Receive')
        + standardOptions.override.byType.withPropertiesFromOptions(
          standardOptions.withUnit('decbytes')
          + color.withMode('fixed')
          + color.withFixedColor(colors.info)
          + custom.stacking.withMode('normal')
        ),
        fieldOverride.byName.new('Storage: Transmit')
        + standardOptions.override.byType.withPropertiesFromOptions(
          standardOptions.withUnit('decbytes')
          + color.withMode('fixed')
          + color.withFixedColor('purple')
          + custom.stacking.withMode('normal')
        ),
      ]),

    memory:
      ts.base('Memory usage', 'The amount of available Random Access Memory', [queries.instance.memory.max, queries.instance.memory.freeable])
      + ts.maxBytes
      + standardOptions.withDecimals(0)
      + standardOptions.withUnit('bytes')
      + standardOptions.withOverrides([
        fieldOverride.byName.new('Max')
        + standardOptions.override.byType.withPropertiesFromOptions(
          standardOptions.withUnit('bytes')
          + custom.withLineWidth('2')
          + color.withMode('fixed')
          + color.withFixedColor(colors.danger)
          + custom.withFillOpacity(0)
        ),
        fieldOverride.byName.new('Freeable')
        + standardOptions.override.byType.withPropertiesFromOptions(
          standardOptions.withUnit('bytes')
          + color.withMode('fixed')
          + color.withFixedColor(colors.ok)
          + custom.stacking.withMode('normal')
        ),
        fieldOverride.byName.new('Used')
        + standardOptions.override.byType.withPropertiesFromOptions(
          standardOptions.withUnit('bytes')
          + color.withMode('fixed')
          + color.withFixedColor(colors.notice)
          + custom.stacking.withMode('normal')
        ),
      ])
      + timeSeries.queryOptions.withTransformations([
        {
          id: 'calculateField',
          options: {
            alias: 'Used',
            binary: {
              left: 'Max',
              operator: '-',
              reducer: 'sum',
              right: 'Freeable',
            },
            mode: 'binary',
            reduce: {
              reducer: 'sum',
            },
          },
        },
      ])
    ,

    swap:
      ts.base('Swap usage', "Amount of swap space used on the DB instance. There's nothing wrong with a used SWAP, but for optimial perfomance you should avoid frequent changes", [queries.instance.memory.swap])
      + ts.singleMetric
      + standardOptions.withUnit('bytes'),

    storage:
      ts.base('Storage usage', 'Storage size per type. WAL only applied to PostgreSQL', [queries.instance.storage.allocated, queries.instance.storage.replicationSlots, queries.instance.storage.wal, queries.instance.storage.logs, queries.instance.storage.free])
      + options.legend.withSortBy('Mean')
      + options.legend.withSortDesc(true)
      + standardOptions.withUnit('bytes')
      + options.legend.withCalcs([
        'min',
        'mean',
        'diff',
      ])
      + standardOptions.withOverrides([
        fieldOverride.byName.new('Allocated')
        + standardOptions.override.byType.withPropertiesFromOptions(
          custom.withLineWidth('1')
          + color.withMode('fixed')
          + color.withFixedColor(colors.danger)
          + custom.withFillOpacity(0)
        ),
        fieldOverride.byName.new('Logs')
        + standardOptions.override.byType.withPropertiesFromOptions(
          color.withMode('fixed')
          + color.withFixedColor('#808080')
          + custom.stacking.withMode('normal')
        ),
        fieldOverride.byName.new('WAL')
        + standardOptions.override.byType.withPropertiesFromOptions(
          color.withMode('fixed')
          + color.withFixedColor('purple')
          + custom.stacking.withMode('normal')
        ),
        fieldOverride.byName.new('Replication slots')
        + standardOptions.override.byType.withPropertiesFromOptions(
          color.withMode('fixed')
          + color.withFixedColor(colors.info)
          + custom.stacking.withMode('normal')
        ),
        fieldOverride.byName.new('Other')
        + standardOptions.override.byType.withPropertiesFromOptions(
          color.withMode('fixed')
          + color.withFixedColor(colors.warning)
          + custom.stacking.withMode('normal')
        ),
        fieldOverride.byName.new('Used')
        + standardOptions.override.byType.withPropertiesFromOptions(
          custom.hideFrom.withLegend(value=true)
          + custom.hideFrom.withTooltip(value=true)
          + custom.hideFrom.withViz(value=true)
        ),
        fieldOverride.byName.new('Free')
        + standardOptions.override.byType.withPropertiesFromOptions(
          color.withMode('fixed')
          + color.withFixedColor(colors.ok)
          + custom.stacking.withMode('normal')
        ),
      ])
      + timeSeries.queryOptions.withTransformations([
        {
          id: 'calculateField',
          options: {
            alias: 'Used',
            mode: 'reduceRow',
            reduce: {
              include: [
                'Free',
                'Logs',
                'Replication slots',
                'WAL',
              ],
              reducer: 'sum',
            },
          },
        },
        {
          id: 'calculateField',
          options: {
            alias: 'Other',
            binary: {
              left: 'Allocated',
              operator: '-',
              reducer: 'sum',
              right: 'Used',
            },
            mode: 'binary',
            reduce: {
              include: [
                'Used',
              ],
              reducer: 'sum',
            },
          },
        },
      ]),

    storagePercent:
      ts.percent('Used storage', 'Ratio of free disk space', [queries.instance.storage.usagePercent])
      + standardOptions.withDecimals(null)
      + ts.singleMetric,

    status:
      ts.base('Status', "RDS instance status. Some AWS operations could not be performed when instance is not in 'available' status", [queries.instance.status])
      + standardOptions.withMappings([
        standardOptions.mapping.ValueMap.withType('value')
        + standardOptions.mapping.ValueMap.withOptions({
          '-10': { index: 0, color: colors.danger, text: 'Restore error' },
          '-9': { index: 1, color: colors.warning, text: 'Maintenance' },
          '-8': { index: 2, color: colors.info, text: 'Upgrading' },
          '-7': { index: 3, color: colors.danger, text: 'Storage-full' },
          '-6': { index: 4, color: colors.danger, text: 'Failed' },
          '-5': { index: 5, color: colors.warning, text: 'Rebooting' },
          '-4': { index: 6, color: colors.danger, text: 'Deleting' },
          '-3': { index: 7, color: colors.notice, text: 'Creating' },
          '-2': { index: 8, color: colors.notice, text: 'Stopping' },
          '-1': { index: 9, color: 'purple', text: 'Unknown' },
          '0': { index: 10, color: colors.danger, text: 'Stopped' },
          '1': { index: 11, color: colors.ok, text: 'Available' },
          '2': { index: 12, color: colors.notice, text: 'Backing-up' },
          '3': { index: 13, color: colors.ok, text: 'Starting' },
          '4': { index: 14, color: colors.info, text: 'Modifying' },
          '5': { index: 15, color: colors.info, text: 'Configuring enhanced monitoring' },
          '6': { index: 16, color: colors.info, text: 'Configuring log exports' },
          '10': { index: 17, colors: colors.info, text: 'Storage initialization' },
          '11': { index: 18, colors: colors.notice, text: 'Storage optimization' },
          '20': { index: 19, colors: colors.notice, text: 'Renaming' },
        }),
      ])
      + standardOptions.withMin(null)
      + standardOptions.withMax(null)
      + standardOptions.withDecimals(0)
      + custom.withFillOpacity(0)
      + options.legend.withDisplayMode('hidden'),
  },
}
