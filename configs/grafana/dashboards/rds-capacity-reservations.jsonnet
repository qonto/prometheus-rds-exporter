local g = import '../g.libsonnet';
local dashboard = g.dashboard;
local common = import '../common.libsonnet';
local variables = import '../variables.libsonnet';
local r = import '../panels/reservations.libsonnet';

local title = 'RDS capacity reservations';
local description = 'RDS On-Demand Capacity Reservations usage and inventory';
local tags = ['dmf', 'misc'];

dashboard.new(title)
+ dashboard.withUid(common.uuids.rdsCapacityReservations)
+ dashboard.withDescription(description)
+ dashboard.withTags(tags)
+ dashboard.withEditable(value=false)
+ dashboard.graphTooltip.withSharedCrosshair()
+ dashboard.withVariables([
  variables.datasource,
  variables.aws_account_ids,
  variables.aws_regions,
  variables.capacity_reservation_id,
])
+ dashboard.withLinks(common.links)
+ dashboard.withPanels(
  g.util.grid.wrapPanels(
    [
      g.panel.row.new('Overview'),
      r.stat.byReservation.count { gridPos+: { w: 3, h: 3 } },
      r.stat.byReservation.reserved { gridPos+: { w: 3, h: 3 } },
      r.stat.byReservation.unused { gridPos+: { w: 3, h: 3 } },
      r.stat.byReservation.ending { gridPos+: { w: 3, h: 3 } },

      g.panel.row.new('Usage'),
      r.barGauge.usage { gridPos+: { w: 12, h: 8 } },
      r.timeSeries.unused { gridPos+: { w: 12, h: 8 } },

      g.panel.row.new('Inventory'),
      r.table.reservations { gridPos+: { w: 24, h: 7 } },
      r.timeSeries.state { gridPos+: { w: 18, h: 8 } },
      r.stat.byState { gridPos+: { w: 6, h: 8 } },

      g.panel.row.new('Instances'),
      r.table.instances { gridPos+: { w: 24, h: 7 } },
    ]
  )
)
