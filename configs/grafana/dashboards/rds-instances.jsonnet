local g = import '../g.libsonnet';
local dashboard = g.dashboard;
local common = import '../common.libsonnet';
local variables = import '../variables.libsonnet';
local p = import '../panels/instances.libsonnet';
local quota = import '../panels/quota.libsonnet';

local uid = 'qonto-prometheus-rds-exporter-instances';
local title = 'RDS instances';
local description = 'RDS instances inventory';
local tags = ['dmf', 'server'];

dashboard.new(title)
+ dashboard.withUid(common.uuids.prometheusRDSExporter)
+ dashboard.withDescription(description)
+ dashboard.withTags(tags)
+ dashboard.withEditable(value=false)
+ dashboard.graphTooltip.withSharedCrosshair()
+ dashboard.withVariables([
  variables.datasource,
  variables.aws_account_ids,
  variables.aws_regions,
])
+ dashboard.withLinks(common.links)
+ dashboard.withPanels(
  g.util.grid.wrapPanels(
    [
      g.panel.row.new('Overview'),
      p.stat.instances { gridPos+: { w: 3, h: 3 } },
      p.stat.pendingMaintenances { gridPos+: { w: 3, h: 3 } },
      p.stat.pendingModification { gridPos+: { w: 3, h: 3 } },
      p.stat.deprecatedCertificates { gridPos+: { w: 3, h: 3 } },

      g.panel.row.new('Pending operations'),
      p.table.pendingMaintenances { gridPos+: { w: 12, h: 5 } },
      p.table.pendingModifications { gridPos+: { w: 12, h: 5 } },

      g.panel.row.new('Inventory'),
      p.table.instances { gridPos+: { w: 24, h: 7 } },

      g.panel.row.new('AWS Quota'),
      quota.timeSeries.instancesRatio { gridPos+: { w: 12, h: 8 } },
      quota.timeSeries.storageRatio { gridPos+: { w: 12, h: 8 } },
      quota.timeSeries.instances { gridPos+: { w: 12, h: 8 } },
      quota.timeSeries.storage { gridPos+: { w: 12, h: 8 } },
    ]
  )
)
