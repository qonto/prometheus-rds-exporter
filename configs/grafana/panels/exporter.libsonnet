local common = import '../common.libsonnet';
local g = import '../g.libsonnet';
local queries = import '../queries/exporter.libsonnet';
local generic = import 'generic.libsonnet';
local colors = common.colors;

local t = generic.table;
local table = g.panel.table;
local stat = generic.stat;

{
  stat: {
    APICalls:
      stat.base('AWS API calls', 'Total calls to AWS APIs to monitor RDS instances', [queries.awsAPICalls]),

    AWSAccounts:
      stat.graph('AWS accounts', 'Number of AWS accounts monitored by RDS exporters', [queries.awsAccounts]),

    exporters:
      stat.graph('Exporters', 'Number of Prometheus RDS exporter', [queries.count]),

    errors:
      stat.errors('Errors', 'Number of errors reported by RDS exporters', [queries.errors]),

    RDSinstances:
      stat.graph('RDS instances', 'Number of AWS RDS instances monitored by RDS exporters', [queries.rdsInstances]),

    status:
      stat.graph('Exporter status', 'Status of RDS exporters', [queries.down])
      + g.panel.stat.standardOptions.withMappings([
        g.panel.stat.standardOptions.mapping.ValueMap.withType('value')
        + g.panel.stat.standardOptions.mapping.ValueMap.withOptions({
          '0': { index: 0, color: colors.danger, text: 'Down' },
          '1': { index: 1, color: colors.ok, text: 'Up' },
        }),
      ]),
  },

  table: {
    list:
      t.base('RDS exporters', 'List of RDS exporters', [queries.all])
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
              kubernetes_cluster: true,
              namespace: true,
              pod: true,
              prometheus: true,
              service: true,
            },
            indexByName: {
              Time: 2,
              Value: 16,
              __name__: 3,
              build_date: 4,
              commit_sha: 5,
              container: 6,
              context: 7,
              endpoint: 8,
              environment: 9,
              instance: 0,
              job: 10,
              kubernetes_cluster: 11,
              namespace: 12,
              pod: 13,
              prometheus: 14,
              service: 15,
              version: 1,
            },
          },
        },
      ]),
  },
  timeSeries: {
    local ts = generic.timeSeries,
    local timeSeries = g.panel.timeSeries,

    APICalls:
      ts.base('AWS API calls per minute', 'Number of HTTP calls to AWS APIs', [queries.awsAPICallsPerMinute])
      + timeSeries.standardOptions.withDecimals(1)
      + timeSeries.fieldConfig.defaults.custom.stacking.withMode('normal'),

    errorsPerMinute:
      ts.errors('Errors per minute', 'Number of RDS exporter errors per minute', [queries.errorsPerMinute]),
  },
}
