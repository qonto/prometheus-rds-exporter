local g = import '../g.libsonnet';
local dashboard = g.dashboard;
local p = import '../panels/instance.libsonnet';
local common = import '../common.libsonnet';
local variables = import '../variables.libsonnet';

local title = 'RDS instance';
local description = 'AWS RDS instance details';
local tags = ['dmf', 'server'];

dashboard.new(title)
+ dashboard.withUid(common.uuids.rdsInstance)
+ dashboard.withDescription(description)
+ dashboard.withTags(tags)
+ dashboard.withEditable(value=false)
+ dashboard.graphTooltip.withSharedCrosshair()
+ dashboard.withLinks(common.links)
+ dashboard.withVariables([
  variables.datasource,
  variables.aws_account_id,
  variables.aws_region,
  variables.dbidentifier,
])
+ dashboard.withPanels(
  g.util.grid.wrapPanels(
    [
      g.panel.row.new('Information'),
      p.stat.engine { gridPos+: { w: 2, h: 2 } },
      p.stat.engineVersion { gridPos+: { w: 2, h: 2 } },
      p.stat.age { gridPos+: { w: 3, h: 2 } },
      p.stat.instanceClass { gridPos+: { w: 3, h: 2 } },
      p.stat.vCPU { gridPos+: { w: 2, h: 2 } },
      p.stat.memory { gridPos+: { w: 3, h: 2 } },
      p.stat.storage { gridPos+: { w: 3, h: 2 } },
      p.stat.storageType { gridPos+: { w: 3, h: 2 } },
      p.stat.multiAZ { gridPos+: { w: 3, h: 2 } },
      p.stat.role { gridPos+: { w: 2, h: 2 } },
      p.stat.replicas { gridPos+: { w: 2, h: 2 } },
      p.stat.source { gridPos+: { w: 8, h: 2 } },
      p.stat.dbi { gridPos+: { w: 6, h: 2 } },
      p.stat.ca { gridPos+: { w: 3, h: 2 } },
      p.stat.snapshotRetention { gridPos+: { w: 3, h: 2 } }
      + g.panel.row.withCollapsed(true),

      g.panel.row.new('Main'),
      p.gauge.cpu { gridPos+: { w: 3, h: 3 } },
      p.gauge.memory { gridPos+: { w: 3, h: 3 } },
      p.gauge.diskIOPS { gridPos+: { w: 3, h: 3 } },
      p.gauge.storage { gridPos+: { w: 3, h: 3 } },
      p.stat.lag { gridPos+: { w: 3, h: 3 } },
      p.stat.status { gridPos+: { w: 3, h: 3 } },
      p.stat.pendingModification { gridPos+: { w: 3, h: 3 } },
      p.stat.pendingMaintenance { gridPos+: { w: 3, h: 3 } },
      p.timeSeries.activeQueries { gridPos+: { w: 12 } },
      p.timeSeries.cpu { gridPos+: { w: 12 } },
      p.timeSeries.databaseConnections { gridPos+: { w: 12 } },
      p.timeSeries.diskIOPS { gridPos+: { w: 12 } },
      p.timeSeries.memory { gridPos+: { w: 12 } },
      p.timeSeries.diskThroughput { gridPos+: { w: 12 } },
      p.timeSeries.swap { gridPos+: { w: 12 } },
      p.timeSeries.storage { gridPos+: { w: 12 } },
      p.timeSeries.status { gridPos+: { w: 12 } },
      p.timeSeries.storagePercent { gridPos+: { w: 12 } },

      g.panel.row.new('PostgreSQL'),
      p.timeSeries.postgresqlMaxUsedTransaction { gridPos+: { w: 12 } },
      p.timeSeries.replicasLag { gridPos+: { w: 12 } },

      g.panel.row.new('Scaling capacity'),
      p.timeSeries.diskIOPSScaling { gridPos+: { w: 12 } },
      p.timeSeries.autoscalingUsage { gridPos+: { w: 12 } },
      p.timeSeries.diskThroughputScaling { gridPos+: { w: 12 } },
    ]
  )
)
