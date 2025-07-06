local common = import '../common.libsonnet';
local g = import '../g.libsonnet';
local queries = import '../queries/instances.libsonnet';
local variables = import '../variables.libsonnet';
local generic = import 'generic.libsonnet';
local colors = common.colors;

{
  stat: {
    local stat = generic.stat,

    instances:
      stat.graph('Instances', 'Number of AWS RDS instances in selected AWS account(s) and region(s)', [queries.count])
      + stat.reset,

    pendingMaintenances:
      stat.alert('Pending maintenances', 'Total number of RDS instances with a pending maintenance', [queries.instancesWithPendingMaintenance]),

    pendingModification:
      stat.alert('Pending modifications', 'Total number of RDS instances with a pending modification', [queries.instancesWithPendingModification]),

    deprecatedCertificates:
      stat.alert('Deprecated certificates', 'Total number of RDS instances with deprecated certificate', [queries.instancesWithDeprecatedCertificate]),
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
        fieldOverride.byName.new('Value')
        + table.standardOptions.override.byType.withPropertiesFromOptions(
          table.standardOptions.withUnit('bytes')
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
              Time: true,
              Value: false,
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
              Value: 19,
              aws_account_id: 0,
              aws_region: 1,
              ca_certificate_identifier: 16,
              dbi_resource_id: 18,
              dbidentifier: 4,
              cluster_identifier: 3,
              deletion_protection: 12,
              engine: 5,
              engine_version: 7,
              instance: 8,
              instance_class: 9,
              job: 10,
              multi_az: 11,
              pending_maintenance: 13,
              pending_modified_values: 14,
              performance_insights_enabled: 15,
              role: 6,
              storage_type: 17,
            },
          },
        },
      ]),

    instances:
      self.__table('RDS instances', 'List of RDS instances in selected AWS account(s) and region(s)', [queries.all]),

    pendingMaintenances:
      self.__table('Instances with pending maintenance', 'RDS instances with pending maintenance', [queries.instancesWithPendingMaintenanceTable])
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
              ca_certificate_identifier: true,
              dbi_resource_id: true,
              deletion_protection: true,
              instance_class: true,
              pending_modified_values: true,
              performance_insights_enabled: true,
              storage_type: true,
              source_dbidentifier: true,
              arn: true,
              role: true,
              engine: true,
              engine_version: true,
              multi_az: true,
            },
            indexByName: {
              aws_account_id: 1,
              aws_region: 2,
              dbidentifier: 3,
              pending_maintenance: 4,
            },
          },
        },
      ]),

    pendingModifications:
      self.__table('Instances with pending modification', 'RDS instances with pending modification', [queries.instancesWithPendingModificationTable])
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
              ca_certificate_identifier: true,
              dbi_resource_id: true,
              deletion_protection: true,
              instance_class: true,
              pending_maintenance: true,
              performance_insights_enabled: true,
              storage_type: true,
              source_dbidentifier: true,
              arn: true,
              role: true,
              engine: true,
              engine_version: true,
              multi_az: true,
            },
            indexByName: {
              aws_account_id: 1,
              aws_region: 2,
              dbidentifier: 3,
              pending_modified_values: 4,
            },
          },
        },
      ]),
  },
}
