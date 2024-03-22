local g = import '../g.libsonnet';
local dashboard = g.dashboard;
local p = import '../panels/exporter.libsonnet';
local common = import '../common.libsonnet';
local variables = import '../variables.libsonnet';

local uid = 'qonto-prometheus-rds-exporter-exporter';
local title = 'Prometheus RDS exporter';
local description = 'Prometheus RDS exporter internal metrics';
local tags = ['dmf', 'misc'];

dashboard.new(title)
+ dashboard.withUid(common.uuids.rdsInstances)
+ dashboard.withDescription(description)
+ dashboard.withTags(tags)
+ dashboard.withEditable(value=false)
+ dashboard.graphTooltip.withSharedCrosshair()
+ dashboard.withVariables([
  variables.datasource,
  variables.exporter,
])
+ dashboard.withLinks(common.links)
+ dashboard.withPanels(
  g.util.grid.wrapPanels(
    [
      g.panel.row.new('Exporters'),
      p.stat.exporters { gridPos+: { w: 3, h: 3 } },
      p.stat.status { gridPos+: { w: 3, h: 3 } },
      p.stat.errors { gridPos+: { w: 3, h: 3 } },
      p.stat.AWSAccounts { gridPos+: { w: 3, h: 3 } },
      p.stat.RDSinstances { gridPos+: { w: 3, h: 3 } },
      p.table.list { gridPos+: { w: 24, h: 5 } },

      g.panel.row.new('Errors'),
      p.timeSeries.errorsPerMinute { gridPos+: { w: 24, h: 7 } },

      g.panel.row.new('API calls'),
      p.timeSeries.APICalls { gridPos+: { w: 19, h: 6 } },
      p.stat.APICalls { gridPos+: { w: 5, h: 6 } },
    ]
  )
)
