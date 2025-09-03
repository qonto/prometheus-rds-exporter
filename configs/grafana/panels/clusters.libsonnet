local common = import '../common.libsonnet';
local g = import '../g.libsonnet';
local queries = import '../queries/clusters.libsonnet';
local variables = import '../variables.libsonnet';
local generic = import 'generic.libsonnet';
local colors = common.colors;

{
  stat: {
    local stat = generic.stat,

    instances:
      stat.graph('Clusters', 'Number of AWS RDS clusters (Amazon Aurora or RDS Multi-AZ DB clusters) in selected AWS account(s) and region(s)', [queries.count])
      + stat.reset,
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
      + table.standardOptions.withOverrides([
        fieldOverride.byName.new('cluster_identifier')
        + table.standardOptions.override.byType.withPropertiesFromOptions(
          table.standardOptions.withLinks([
            {
              title: '',
              url: '/d/' + common.uuids.rdsCluster + '/?${datasource:queryparam}&${__url_time_range}&var-' + variables.aws_account_id.name + '=${__data.fields.aws_account_id}&var-' + variables.aws_region.name + '=${__data.fields.aws_region}&var-' + variables.cluster_identifier.name + '=${__data.fields.cluster_identifier}',
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
              Time: 2,
              Value: 17,
              aws_account_id: 0,
              aws_region: 1,
              ca_certificate_identifier: 15,
              dbi_resource_id: 18,
              cluster_identifier: 3,
              deletion_protection: 11,
              engine: 4,
              engine_version: 6,
              instance: 7,
              instance_class: 8,
              job: 9,
            },
          },
        },
      ]),

    clusters:
      self.__table('RDS clusters', 'List of RDS clusters (Amazon Aurora or RDS Multi-AZ DB clusters) in selected AWS account(s) and region(s)', [queries.all]),
  },
  pie: {
    local pie = generic.pie,

    byEngine:
      pie.base('Clusters by engine', 'Total number of cluster per engines in selected AWS account(s) and region(s)', [queries.byEngine]),

  },
}
