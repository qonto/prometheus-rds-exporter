local common = import '../common.libsonnet';
local g = import '../g.libsonnet';
local queries = import '../queries/cluster.libsonnet';
local variables = import '../variables.libsonnet';
local generic = import 'generic.libsonnet';
local colors = common.colors;

{
  stat: {
    local stat = generic.stat,
    local standardOptions = g.panel.stat.standardOptions,
    local thresholds = g.panel.stat.standardOptions.thresholds,
    local step = standardOptions.threshold.step,

    instances:
      stat.graph('Instances', 'Number of instances in the cluster', [queries.instances.count])
      + stat.reset,

    engine:
      stat.field('Engine', 'Engine', queries.cluster.info, 'engine'),

    engineVersion:
      stat.field('Engine Version', 'Cluster engine version', queries.cluster.info, 'engine_version'),

    byClassType:
      stat.graph('Instances class', 'Number of instances by instance class type', [queries.instances.byClassType])
      + thresholds.withMode('absolute')
      + thresholds.withSteps([
        step.withValue(0) + step.withColor('white'),
      ]),

    serverLessConfiguration:
      stat.graph('Serverless configuration', 'Minimum/Maximum number of Aurora capacity units (ACUs) for a DB instance in an Aurora Serverless v2 cluster.', [queries.cluster.serverless.minACU, queries.cluster.serverless.maxACU])
      + standardOptions.withDecimals(1)
      + thresholds.withMode('absolute')
      + thresholds.withSteps([
        step.withValue(0) + step.withColor('white'),
      ]),

  },
  table: {
    local t = generic.table,
    local table = g.panel.table,
    local fieldOverride = g.panel.timeSeries.fieldOverride,
    local thresholds = table.standardOptions.thresholds,
    local step = table.standardOptions.threshold.step,
    local mapping = table.standardOptions.mapping,

    __booleanField(fieldName):
      fieldOverride.byName.new(fieldName)
      + self.__trueOrFalse,

    __invertedBooleanField(fieldName):
      fieldOverride.byName.new(fieldName)
      + self.__falseOrTrue,

    __trueOrFalse:
      table.standardOptions.override.byType.withPropertiesFromOptions(
        table.standardOptions.withMappings([
          mapping.ValueMap.withType('value')
          + mapping.ValueMap.withOptions({
            'true': { index: 0, color: colors.ok },
            'false': { index: 1, color: colors.notice },
          }),
        ]),
      ),

    __falseOrTrue:
      table.standardOptions.override.byType.withPropertiesFromOptions(
        table.standardOptions.withMappings([
          mapping.ValueMap.withType('value')
          + mapping.ValueMap.withOptions({
            'false': { index: 0, color: colors.ok },
            'true': { index: 1, color: colors.notice },
          }),
        ]),
      ),

    __table(title, description, targets):
      t.base(title, description, targets)
      + table.fieldConfig.defaults.custom.withDisplayMode('color-text')
      + thresholds.withMode('absolute')
      + thresholds.withSteps([
        step.withValue(0) + step.withColor('white'),
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
        self.__booleanField('multi_az'),
        self.__booleanField('deletion_protection'),
        self.__invertedBooleanField('pending_modified_values'),
        self.__booleanField('performance_insights_enabled'),
        fieldOverride.byName.new('pending_maintenance')
        + table.standardOptions.override.byType.withPropertiesFromOptions(
          table.standardOptions.withMappings([
            mapping.ValueMap.withType('value')
            + mapping.ValueMap.withOptions({
              no: { index: 0, color: colors.ok },
              pending: { index: 0, color: colors.notice },
              'auto-applied': { index: 0, color: colors.warning },
              forced: { index: 1, color: colors.danger },
            }),
          ]),
        ),
      ])
      + table.queryOptions.withTransformations([
        {
          id: 'organize',
          options: {
            excludeByName: {
              __name__: true,
              Time: true,
              Value: true,
              cluster_identifier: true,
              container: true,
              context: true,
              endpoint: true,
              engine: true,
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
              dbidentifier: 3,
              role: 4,
              engine: 5,
              engine_version: 6,
              instance: 7,
              instance_class: 8,
              multi_az: 9,
              pending_modified_values: 10,
              pending_maintenance: 11,
              performance_insights_enabled: 12,
              deletion_protection: 13,
              storage_type: 14,
              ca_certificate_identifier: 15,
              dbi_resource_id: 16,
              arn: 17,
            },
          },
        },
      ]),

    instances:
      self.__table('Instances', 'List of RDS instances in the clusters', [queries.instances.all]),
  },
}
