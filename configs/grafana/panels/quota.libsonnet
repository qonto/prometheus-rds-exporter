local common = import '../common.libsonnet';
local g = import '../g.libsonnet';
local queries = import '../queries/instances.libsonnet';
local generic = import 'generic.libsonnet';
local colors = common.colors;

{
  timeSeries: {
    local ts = generic.timeSeries,
    local timeSeries = g.panel.timeSeries,
    local standardOptions = timeSeries.standardOptions,
    local fieldOverride = g.panel.timeSeries.fieldOverride,

    withTotal:
      standardOptions.withOverrides([
        fieldOverride.byRegexp.new('Max.*')
        + standardOptions.override.byType.withPropertiesFromOptions(
          standardOptions.color.withMode('fixed')
          + standardOptions.color.withFixedColor(colors.danger)
          + timeSeries.fieldConfig.defaults.custom.withFillOpacity(0)
        ),
        fieldOverride.byRegexp.new('Total.*')
        + standardOptions.override.byType.withPropertiesFromOptions(
          standardOptions.color.withMode('fixed')
          + standardOptions.color.withFixedColor(colors.ok)
        ),
      ]),

    instances:
      ts.base('RDS instance quota', 'The number of instances regarding your RDS instances quota. You can request a quota increase to AWS', [queries.instances.max, queries.instances.total])
      + self.withTotal,

    instancesRatio:
      ts.percent('RDS instance quota usage (ratio)', 'Ratio of used RDS instances regarding AWS quota', [queries.instances.ratio]),

    storage:
      ts.base('RDS storage quota', 'The total storage used regarding your RDS total storage quota. You can request a quota increase to AWS', [queries.storage.max, queries.storage.total])
      + self.withTotal
      + timeSeries.standardOptions.withUnit('bytes'),

    storageRatio:
      ts.percent('RDS storage quota usage (ratio)', 'Ratio of used RDS storage regarding AWS quota', [queries.storage.ratio]),
  },
}
