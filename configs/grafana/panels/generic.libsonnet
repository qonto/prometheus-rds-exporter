local common = import '../common.libsonnet';
local g = import '../g.libsonnet';
local colors = common.colors;

{
  stat: {
    local stat = g.panel.stat,
    local options = stat.options,
    local standardOptions = stat.standardOptions,
    local thresholds = stat.standardOptions.thresholds,
    local step = stat.standardOptions.threshold.step,

    __stat(title, description, targets):
      stat.new(title)
      + stat.panelOptions.withDescription(description)
      + stat.queryOptions.withTargets(targets)
      + stat.options.withGraphMode('none'),

    reset:
      thresholds.withMode('absolute')
      + thresholds.withSteps([
        step.withValue(0) + step.withColor('white'),
      ]),

    base(title, description, targets):
      self.__stat(title, description, targets)
      + self.reset
      + standardOptions.withNoValue('n/a'),

    graph(title, description, targets):
      self.__stat(title, description, targets)
      + options.withGraphMode('area'),

    alert(title, description, targets):
      self.graph(title, description, targets)
      + standardOptions.withNoValue('0')
      + options.withColorMode('background')
      + thresholds.withSteps([
        step.withValue(0) + step.withColor(colors.ok),
        step.withValue(1) + step.withColor(colors.warning),
      ]),

    errors(title, description, targets):
      self.graph(title, description, targets)
      + thresholds.withMode('absolute')
      + thresholds.withSteps([
        step.withValue(0) + step.withColor(colors.ok),
        step.withValue(1) + step.withColor(colors.warning),
      ]),

    field(title, description, targets, field):
      self.base(title, description, targets)
      + options.reduceOptions.withFields(field),

    bytes(title, description, targets):
      self.base(title, description, targets)
      + standardOptions.withUnit('bytes')
      + standardOptions.withDecimals(2),

    lag(title, description, targets):
      self.__stat(title, description, targets)
      + options.withGraphMode('area')
      + standardOptions.withMin(0)
      + standardOptions.withUnit('s')
      + standardOptions.thresholds.withMode('absolute')
      + thresholds.withSteps([
        step.withValue(0) + step.withColor(colors.ok),
        step.withValue(10) + step.withColor(colors.warning),
        step.withValue(30) + step.withColor(colors.danger),
      ])
      + standardOptions.withNoValue('n/a'),

    enabledOrDisabled(title, description, targets):
      self.base(title, description, targets)
      + standardOptions.withMappings([
        standardOptions.mapping.ValueMap.withType('value')
        + standardOptions.mapping.ValueMap.withOptions({
          '0': { index: 0, color: colors.warning, text: 'Disabled' },
          '1': { index: 1, color: colors.ok, text: 'Enabled' },
        }),
      ]),
  },
  gauge: {
    local gauge = g.panel.gauge,
    local standardOptions = gauge.standardOptions,
    local thresholds = gauge.standardOptions.thresholds,
    local step = gauge.standardOptions.threshold.step,

    base(title, description, targets):
      gauge.new(title)
      + gauge.panelOptions.withDescription(description)
      + gauge.queryOptions.withTargets(targets),

    percent(title, description, targets):
      self.base(title, description, targets)
      + standardOptions.withUnit('percent')
      + standardOptions.withMin(0)
      + standardOptions.withMax(100)
      + standardOptions.withDecimals(0)
      + thresholds.withMode('percentage')
      + thresholds.withSteps([
        step.withValue(0) + step.withColor(colors.ok),
        step.withValue(80) + step.withColor(colors.warning),
        step.withValue(90) + step.withColor(colors.danger),
      ]),
  },
  timeSeries: {
    local timeSeries = g.panel.timeSeries,
    local options = timeSeries.options,
    local legend = timeSeries.options.legend,
    local panelOptions = timeSeries.panelOptions,
    local standardOptions = timeSeries.standardOptions,
    local thresholds = standardOptions.thresholds,
    local step = standardOptions.threshold.step,
    local fieldOverride = g.panel.timeSeries.fieldOverride,

    base(title, description, targets):
      timeSeries.new(title)
      + panelOptions.withDescription(description)
      + timeSeries.queryOptions.withTargets(targets)
      + legend.withDisplayMode('table')
      + legend.withPlacement('right')
      + legend.withCalcs([
        'min',
        'mean',
        'max',
      ])
      + timeSeries.fieldConfig.defaults.custom.withFillOpacity(10)
      + standardOptions.withMin(0)
      + standardOptions.withDecimals(0),

    single(title, description, targets):
      self.base(title, description, targets)
      + legend.withDisplayMode('list')
      + legend.withPlacement('bottom'),

    errors(title, description, targets):
      self.base(title, description, targets)
      + standardOptions.color.withMode('thresholds')
      + thresholds.withMode('absolute')
      + thresholds.withSteps([
        step.withValue(0) + step.withColor(colors.ok),
        step.withValue(1) + step.withColor(colors.warning),
      ]),

    max:
      timeSeries.standardOptions.override.byType.withPropertiesFromOptions(
        standardOptions.color.withMode('fixed')
        + standardOptions.color.withFixedColor(colors.danger)
        + timeSeries.fieldConfig.defaults.custom.stacking.withMode('off')
        + timeSeries.fieldConfig.defaults.custom.hideFrom.withLegend(value=true)
        + timeSeries.fieldConfig.defaults.custom.withFillOpacity(0)
      ),

    maxBytes:
      standardOptions.withOverrides([
        fieldOverride.byRegexp.new('Max.*')
        + standardOptions.override.byType.withPropertiesFromOptions(
          standardOptions.withUnit('bytes')
          + standardOptions.color.withMode('fixed')
          + standardOptions.color.withFixedColor(colors.danger)
          + timeSeries.fieldConfig.defaults.custom.withFillOpacity(0)
        ),
      ]),

    singleMetric:
      options.legend.withDisplayMode('list')
      + options.legend.withPlacement('bottom'),

    percent(title, description, targets):
      self.base(title, description, targets)
      + standardOptions.withMin(0)
      + standardOptions.withMax(100)
      + standardOptions.withDecimals(0)
      + standardOptions.withUnit('percent')
      + timeSeries.standardOptions.color.withMode('thresholds')
      + timeSeries.fieldConfig.defaults.custom.lineStyle.withDash(true)
      + timeSeries.fieldConfig.defaults.custom.lineStyle.withFill('solid')
      + timeSeries.fieldConfig.defaults.custom.withThresholdsStyle({
        mode: 'line',
      })
      + thresholds.withMode('percentage')
      + thresholds.withSteps([
        step.withValue(0) + step.withColor(colors.ok),
        step.withValue(100) + step.withColor(colors.danger),
      ]),
  },
  table: {
    local table = g.panel.table,

    base(title, description, targets):
      table.new(title)
      + table.panelOptions.withDescription(description)
      + table.queryOptions.withTargets(targets),
  },
  pie: {
    local pie = g.panel.pieChart,
    local standardOptions = pie.standardOptions,
    local legend = pie.options.legend,

    base(title, description, targets):
      pie.new(title)
      + pie.panelOptions.withDescription(description)
      + pie.queryOptions.withTargets(targets)
      + standardOptions.withDecimals(0)
      + legend.withValues(['percent', 'value'])
      + legend.withDisplayMode('table')
      + legend.withPlacement('right')
      + legend.withCalcs([
        'Percent',
        'Value',
      ]),

  },
}
