local g = import '../g.libsonnet';
local dashboard = g.dashboard;
local common = import '../common.libsonnet';
local variables = import '../variables.libsonnet';
local p = import '../panels/clusters.libsonnet';
local quota = import '../panels/quota.libsonnet';

local uid = 'qonto-prometheus-rds-exporter-clusters';
local title = 'RDS clusters';
local description = 'RDS clusters inventory';
local tags = ['dmf', 'server'];

dashboard.new(title)
+ dashboard.withUid(common.uuids.rdsClusters)
+ dashboard.withDescription(description)
+ dashboard.withTags(tags)
+ dashboard.withEditable(value=true)  // TODO Fix this
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
      p.stat.instances { gridPos+: { w: 3, h: 5 } },
      p.pie.byEngine { gridPos+: { w: 8, h: 5 } },

      g.panel.row.new('Inventory'),
      p.table.clusters { gridPos+: { w: 24, h: 7 } },
    ]
  )
)
