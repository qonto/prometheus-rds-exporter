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

    count:
      stat.graph('Capacity reservations', 'Number of active On-Demand Capacity Reservations in selected AWS account(s) and region(s)', [queries.count])
      + stat.reset,

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
  // Usage of reserved capacity: low usage means paying for unused capacity
  __usageThresholds(standardOptions):
    standardOptions.withUnit('percent')
    + standardOptions.withMin(0)
    + standardOptions.withMax(100)
    + standardOptions.withDecimals(0)
    + standardOptions.thresholds.withMode('absolute')
    + standardOptions.thresholds.withSteps([
      standardOptions.threshold.step.withValue(0) + standardOptions.threshold.step.withColor(colors.danger),
      standardOptions.threshold.step.withValue(50) + standardOptions.threshold.step.withColor(colors.warning),
      standardOptions.threshold.step.withValue(90) + standardOptions.threshold.step.withColor(colors.ok),
    ]),

  statusHistory: {
    local statusHistory = g.panel.statusHistory,
    local standardOptions = statusHistory.standardOptions,

    usage:
      statusHistory.new('Capacity reservations usage')
      + statusHistory.panelOptions.withDescription('Percentage of reserved DB instances used by RDS instances, per AWS account, region, engine, instance class and availability zone')
      + statusHistory.queryOptions.withTargets([queries.usage.ratio])
      + statusHistory.options.withShowValue('auto')
      + statusHistory.options.withRowHeight(0.9)
      + standardOptions.color.withMode('thresholds')
      + $.__usageThresholds(standardOptions),
  },

  barGauge: {
    local barGauge = g.panel.barGauge,
    local standardOptions = barGauge.standardOptions,

    usage:
      barGauge.new('Current capacity reservations usage')
      + barGauge.panelOptions.withDescription('Current percentage of reserved DB instances used by RDS instances, per AWS account, region, engine, instance class and availability zone')
      + barGauge.queryOptions.withTargets([queries.usage.ratio])
      + barGauge.options.withOrientation('horizontal')
      + barGauge.options.withDisplayMode('basic')
      + barGauge.options.withShowUnfilled(true)
      + barGauge.options.reduceOptions.withCalcs(['lastNotNull'])
      + $.__usageThresholds(standardOptions),
  },

  timeSeries: {
    local ts = generic.timeSeries,

    unused:
      ts.base('Unused reserved DB instances', 'Number of reserved DB instances not used by any RDS instance, per AWS account, region, engine, instance class and availability zone', [queries.usage.unused]),
  },

  table: {
    local t = generic.table,
    local table = g.panel.table,
    local fieldOverride = g.panel.table.fieldOverride,
    local thresholds = table.standardOptions.thresholds,
    local step = table.standardOptions.threshold.step,
    local mapping = table.standardOptions.mapping,

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

    reservations:
      t.base('Capacity reservations', 'List of On-Demand Capacity Reservations in selected AWS account(s) and region(s)', [queries.all])
      + table.fieldConfig.defaults.custom.withDisplayMode('color-text')
      + table.standardOptions.withNoValue('No capacity reservation')
      + thresholds.withMode('absolute')
      + thresholds.withSteps([
        step.withValue(0) + step.withColor(colors.default),
      ])
      + table.standardOptions.withOverrides([
        fieldOverride.byName.new('state')
        + table.standardOptions.override.byType.withPropertiesFromOptions(
          table.standardOptions.withMappings([
            mapping.ValueMap.withType('value')
            + mapping.ValueMap.withOptions({
              active: { index: 0, color: colors.ok },
              assessing: { index: 1, color: colors.info },
              scheduled: { index: 2, color: colors.info },
              pending: { index: 3, color: colors.notice },
              delayed: { index: 4, color: colors.warning },
              failed: { index: 5, color: colors.danger },
              unsupported: { index: 6, color: colors.danger },
              expired: { index: 7, color: colors.default },
              cancelled: { index: 8, color: colors.default },
            }),
          ]),
        ),
        // Value is the end date in milliseconds, 0 when the reservation doesn't expire
        fieldOverride.byName.new('Value')
        + table.standardOptions.override.byType.withPropertiesFromOptions(
          table.standardOptions.withUnit('dateTimeAsIso')
          + table.standardOptions.withDisplayName('End date')
          + table.standardOptions.withMappings([
            mapping.ValueMap.withType('value')
            + mapping.ValueMap.withOptions({
              '0': { index: 0, text: 'No end date' },
            }),
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
              engine: 4,
              instance_class: 5,
              state: 6,
              Value: 7,
              arn: 8,
            },
          },
        },
      ]),
  },
}
