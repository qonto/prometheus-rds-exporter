local common = import '../common.libsonnet';
local g = import '../g.libsonnet';
local queries = import '../queries/reservations.libsonnet';
local generic = import 'generic.libsonnet';
local colors = common.colors;

{
  stat: {
    local stat = generic.stat,
    local standardOptions = g.panel.stat.standardOptions,
    local thresholds = standardOptions.thresholds,
    local step = standardOptions.threshold.step,

    unused:
      stat.alert('Unused capacity reservations', 'Number of DB instances reserved by active On-Demand Capacity Reservations and not used by any RDS instance', [queries.unused])
      + standardOptions.withDecimals(0)
      + thresholds.withSteps([
        step.withValue(0) + step.withColor(colors.ok),
        step.withValue(2) + step.withColor(colors.warning),
      ]),

    ending:
      stat.alert('Capacity reservation ending soon', 'Total number of active On-Demand Capacity Reservations ending in less than 30 days', [queries.ending])
      + standardOptions.withDecimals(0),
  },
  table: {
    local t = generic.table,
    local table = g.panel.table,
    local fieldOverride = g.panel.table.fieldOverride,
    local thresholds = table.standardOptions.thresholds,
    local step = table.standardOptions.threshold.step,

    ending:
      t.base('Capacity reservations ending soon', 'Active On-Demand Capacity Reservations ending in less than 30 days', [queries.endingTable])
      + table.fieldConfig.defaults.custom.withDisplayMode('color-text')
      + table.standardOptions.withNoValue('No capacity reservation')
      + table.standardOptions.withDecimals(0)
      + table.standardOptions.withOverrides([
        fieldOverride.byName.new('Value')
        + table.standardOptions.override.byType.withPropertiesFromOptions(
          table.standardOptions.withUnit('d')
          + table.standardOptions.withDisplayName('Days remaining')
          + thresholds.withMode('absolute')
          + thresholds.withSteps([
            step.withValue(0) + step.withColor(colors.danger),
            step.withValue(7) + step.withColor(colors.warning),
            step.withValue(30) + step.withColor(colors.ok),
          ])
        ),
      ])
      + table.queryOptions.withTransformations([
        {
          id: 'organize',
          options: {
            excludeByName: {
              Time: true,
              __name__: true,
              container: true,
              context: true,
              endpoint: true,
              environment: true,
              instance: true,
              job: true,
              kubernetes_cluster: true,
              namespace: true,
              pod: true,
              prometheus: true,
              service: true,
            },
            indexByName: {
              aws_account_id: 0,
              aws_region: 1,
              capacity_reservation_id: 2,
              capacity_reservation_name: 3,
              Value: 4,
              engine: 5,
              instance_class: 6,
            },
          },
        },
      ]),
  },
}
