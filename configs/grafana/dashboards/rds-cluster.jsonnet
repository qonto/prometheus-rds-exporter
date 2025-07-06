local g = import '../g.libsonnet';
local dashboard = g.dashboard;
local common = import '../common.libsonnet';
local variables = import '../variables.libsonnet';
local p = import '../panels/cluster.libsonnet';
local quota = import '../panels/quota.libsonnet';

local title = 'RDS cluster details';
local description = 'RDS cluster details';
local tags = ['dmf', 'server'];

dashboard.new(title)
+ dashboard.withUid(common.uuids.rdsCluster)
+ dashboard.withDescription(description)
+ dashboard.withTags(tags)
+ dashboard.withEditable(value=true)  // TODO Fix this
+ dashboard.graphTooltip.withSharedCrosshair()
+ dashboard.withVariables([
  variables.datasource,
  variables.aws_account_ids,
  variables.aws_regions,
  variables.cluster_identifier,
])
+ dashboard.withLinks(common.links)
+ dashboard.withPanels(
  g.util.grid.wrapPanels(
    [
      g.panel.row.new('Overview'),
      p.stat.engine { gridPos+: { w: 4, h: 3 } },
      p.stat.engineVersion { gridPos+: { w: 3, h: 3 } },
      p.stat.instances { gridPos+: { w: 3, h: 3 } },
      p.stat.byClassType { gridPos+: { w: 14, h: 3 } },

      g.panel.row.new('Inventory'),
      p.table.instances { gridPos+: { w: 24, h: 7 } },
    ]
  )
)
