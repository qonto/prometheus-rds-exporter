local common = import '../common.libsonnet';
local g = import '../g.libsonnet';
local queries = import '../queries/reservations.libsonnet';
local variables = import '../variables.libsonnet';
local generic = import 'generic.libsonnet';
local colors = common.colors;

{
  // Capacity reservation states, with the same colors as the state column of the capacity reservations list
  local states = {
    active: { color: colors.ok, text: 'Active' },
    pending: { color: colors.notice, text: 'Pending' },
    assessing: { color: colors.info, text: 'Assessing' },
    scheduled: { color: colors.info, text: 'Scheduled' },
    delayed: { color: colors.warning, text: 'Delayed' },
    expired: { color: colors.default, text: 'Expired' },
    cancelled: { color: colors.default, text: 'Cancelled' },
    failed: { color: colors.danger, text: 'Failed' },
    unsupported: { color: colors.danger, text: 'Unsupported' },
  },

  stat: {
    local stat = generic.stat,
    local standardOptions = g.panel.stat.standardOptions,
    local thresholds = standardOptions.thresholds,
    local step = standardOptions.threshold.step,

    __count(query):
      stat.graph('Capacity reservations', 'Number of active On-Demand Capacity Reservations in selected AWS account(s) and region(s)', [query])
      + stat.reset,

    __reserved(query):
      stat.graph('Total reserved capacity', 'Number of DB instances reserved by active On-Demand Capacity Reservations', [query])
      + stat.reset
      + standardOptions.withDecimals(0),

    __unused(query):
      stat.alert('Unused capacity reservations', 'Number of DB instances reserved by active On-Demand Capacity Reservations and not used by any RDS instance', [query])
      + standardOptions.withDecimals(0)
      + thresholds.withSteps([
        step.withValue(0) + step.withColor(colors.ok),
        step.withValue(2) + step.withColor(colors.warning),
      ]),

    __ending(query):
      stat.alert('Capacity reservation ending soon', 'Total number of active On-Demand Capacity Reservations ending in less than 30 days', [query])
      + standardOptions.withDecimals(0),

    byState:
      stat.base('Reservations by state', 'Current number of On-Demand Capacity Reservations per state', [queries.countByState])
      + g.panel.stat.options.withTextMode('value_and_name')
      + g.panel.stat.options.withColorMode('background')
      + g.panel.stat.options.withOrientation('horizontal')
      + standardOptions.withDecimals(0)
      + standardOptions.withNoValue('No capacity reservation')
      + standardOptions.withOverrides([
        g.panel.stat.fieldOverride.byName.new(state)
        + standardOptions.override.byType.withPropertiesFromOptions(
          standardOptions.withDisplayName(states[state].text)
          + standardOptions.color.withMode('fixed')
          + standardOptions.color.withFixedColor(states[state].color)
        )
        for state in std.objectFields(states)
      ]),

    count: self.__count(queries.count),
    unused: self.__unused(queries.unused),
    ending: self.__ending(queries.ending),

    // Stats filtered by the capacity_reservation_id variable
    byReservation: {
      count: $.stat.__count(queries.byReservation.count),
      reserved: $.stat.__reserved(queries.byReservation.reserved),
      unused: $.stat.__unused(queries.byReservation.unused),
      ending: $.stat.__ending(queries.byReservation.ending),
    },
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

  barGauge: {
    local barGauge = g.panel.barGauge,
    local standardOptions = barGauge.standardOptions,

    usage:
      barGauge.new('Current capacity reservations usage')
      + barGauge.panelOptions.withDescription('Current percentage of reserved capacity used by RDS instances, per capacity reservation (all availability zones). Reserved capacity not used by any RDS instance is billed at the On-Demand rate.')
      + barGauge.queryOptions.withTargets([queries.usage.current])
      + barGauge.options.withOrientation('horizontal')
      + barGauge.options.withDisplayMode('lcd')
      + barGauge.options.withShowUnfilled(true)
      + barGauge.options.reduceOptions.withCalcs(['lastNotNull'])
      + barGauge.options.reduceOptions.withValues(true)
      + $.__usageThresholds(standardOptions)
      // Sort bars by usage, then by name: series are reduced to rows, then sorted by name and by usage (stable sort)
      + barGauge.queryOptions.withTransformations([
        {
          id: 'reduce',
          options: {
            mode: 'seriesToRows',
            reducers: ['lastNotNull'],
          },
        },
        {
          id: 'sortBy',
          options: {
            sort: [{ field: 'Field' }],
          },
        },
        {
          id: 'sortBy',
          options: {
            sort: [{ field: 'Last *' }],
          },
        },
      ]),
  },

  timeSeries: {
    local ts = generic.timeSeries,
    local timeSeries = g.panel.timeSeries,
    local standardOptions = timeSeries.standardOptions,
    local mapping = standardOptions.mapping,
    local custom = timeSeries.fieldConfig.defaults.custom,
    local codes = queries.stateCodes,

    // One line per reservation, states are displayed on the Y axis like the RDS instance status
    state:
      ts.base('Capacity reservations state', 'State of each On-Demand Capacity Reservation over time (e.g. pending to active, scheduled to active, active to expired)', [queries.state])
      + standardOptions.withMappings([
        mapping.ValueMap.withType('value')
        + mapping.ValueMap.withOptions(
          {
            [std.toString(codes[state])]: states[state] { index: codes[state] }
            for state in std.objectFields(codes)
          }
          + { '-1': { index: 0, color: 'purple', text: 'Unknown' } }
        ),
      ])
      + standardOptions.withMin(null)
      + standardOptions.withMax(null)
      + standardOptions.withDecimals(0)
      + custom.withFillOpacity(0)
      + timeSeries.options.legend.withDisplayMode('list')
      + timeSeries.options.legend.withPlacement('bottom')
      + timeSeries.options.legend.withCalcs([]),

    unused:
      ts.base('Unused reserved capacity', 'Reserved capacity (number of DB instances) not used by any RDS instance, per capacity reservation and availability zone', [queries.usage.unused])
      + g.panel.timeSeries.fieldConfig.defaults.custom.stacking.withMode('normal'),
  },

  table: {
    local t = generic.table,
    local table = g.panel.table,
    local fieldOverride = g.panel.table.fieldOverride,

    // Open the capacity reservation in the AWS console
    __consoleLink:
      fieldOverride.byName.new('capacity_reservation_name')
      + table.standardOptions.override.byType.withPropertiesFromOptions(
        table.standardOptions.withLinks([
          {
            title: 'Open in AWS console',
            url: 'https://${__data.fields.aws_region}.console.aws.amazon.com/rds/home?region=${__data.fields.aws_region}#capacity-reservation:id=${__data.fields.capacity_reservation_id}',
            targetBlank: true,
          },
        ])
      ),
    local thresholds = table.standardOptions.thresholds,
    local step = table.standardOptions.threshold.step,
    local mapping = table.standardOptions.mapping,

    ending:
      t.base('Capacity reservations ending soon', 'Active On-Demand Capacity Reservations ending in less than 30 days', [queries.endingTable])
      + table.fieldConfig.defaults.custom.withDisplayMode('color-text')
      + table.standardOptions.withNoValue('No capacity reservation')
      + table.standardOptions.withDecimals(0)
      + table.standardOptions.withOverrides([
        self.__consoleLink,
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
      t.base('Capacity reservations', 'List of On-Demand Capacity Reservations in selected AWS account(s) and region(s)', [queries.all, queries.remainingDays])
      + table.fieldConfig.defaults.custom.withDisplayMode('color-text')
      + table.standardOptions.withNoValue('No capacity reservation')
      + thresholds.withMode('absolute')
      + thresholds.withSteps([
        step.withValue(0) + step.withColor(colors.default),
      ])
      + table.standardOptions.withOverrides([
        self.__consoleLink,
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
        // Value #A is the end date in milliseconds, 0 when the reservation doesn't expire
        fieldOverride.byName.new('Value #A')
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
        // Value #B is the number of days before the end date (NaN when there is no end date), red when less than 30 days.
        // A range mapping is used instead of thresholds to keep the default color when there is no end date.
        fieldOverride.byName.new('Value #B')
        + table.standardOptions.override.byType.withPropertiesFromOptions(
          table.standardOptions.withUnit('d')
          + table.standardOptions.withDecimals(0)
          + table.standardOptions.withDisplayName('Days remaining')
          + table.standardOptions.withMappings([
            mapping.RangeMap.withType('range')
            + mapping.RangeMap.options.withTo(30)
            + mapping.RangeMap.options.result.withColor(colors.danger),
            mapping.SpecialValueMap.withType('special')
            + mapping.SpecialValueMap.options.withMatch('null+nan')
            + mapping.SpecialValueMap.options.result.withText('No end date'),
          ])
        ),
      ])
      + table.queryOptions.withTransformations([
        // Merge end date (A) and days remaining (B) of each reservation in a single row
        {
          id: 'merge',
          options: {},
        },
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
              engine: 2,
              instance_class: 3,
              state: 4,
              capacity_reservation_name: 5,
              'Value #A': 6,
              'Value #B': 7,
              capacity_reservation_id: 8,
              arn: 9,
            },
          },
        },
      ]),

    instances:
      t.base('RDS instances using a capacity reservation', 'RDS instances using the reserved capacity of the selected On-Demand Capacity Reservations', [queries.instances])
      + table.fieldConfig.defaults.custom.withDisplayMode('color-text')
      + table.standardOptions.withNoValue('No instance')
      + thresholds.withMode('absolute')
      + thresholds.withSteps([
        step.withValue(0) + step.withColor(colors.default),
      ])
      + table.standardOptions.withOverrides([
        fieldOverride.byName.new('dbidentifier')
        + table.standardOptions.override.byType.withPropertiesFromOptions(
          table.standardOptions.withLinks([
            {
              title: '',
              url: '/d/' + common.uuids.rdsInstance + '/?${datasource:queryparam}&${__url_time_range}&var-' + variables.aws_account_id.name + '=${__data.fields.aws_account_id}&var-' + variables.aws_region.name + '=${__data.fields.aws_region}&var-' + variables.dbidentifier.name + '=${__data.fields.dbidentifier}',
            },
          ])
        ),
      ])
      + table.queryOptions.withTransformations([
        {
          id: 'organize',
          options: {
            excludeByName: {
              Time: true,
              Value: true,
              __name__: true,
              arn: true,
              ca_certificate_identifier: true,
              container: true,
              context: true,
              dbi_resource_id: true,
              deletion_protection: true,
              endpoint: true,
              environment: true,
              instance: true,
              job: true,
              kubernetes_cluster: true,
              namespace: true,
              pending_maintenance: true,
              pending_modified_values: true,
              performance_insights_enabled: true,
              pod: true,
              prometheus: true,
              service: true,
              source_dbidentifier: true,
              storage_type: true,
            },
            indexByName: {
              aws_account_id: 0,
              aws_region: 1,
              availability_zone: 2,
              capacity_reservation_name: 3,
              dbidentifier: 4,
              cluster_identifier: 5,
              role: 6,
              engine: 7,
              engine_version: 8,
              instance_class: 9,
              multi_az: 10,
              capacity_reservation_id: 11,
            },
          },
        },
      ]),
  },
}
