local g = import '../g.libsonnet';
local prometheusQuery = g.query.prometheus;
local variables = import '../variables.libsonnet';

{
  all:
    prometheusQuery.new(
      '$' + variables.datasource.name,
      |||
        rds_exporter_build_info{instance=~"$instance"}
      |||
    )
    + prometheusQuery.withInstant(true)
    + prometheusQuery.withFormat('table'),

  count:
    prometheusQuery.new(
      '$' + variables.datasource.name,
      |||
        count(rds_exporter_build_info{instance=~"$instance"})
      |||
    ),

  down:
    prometheusQuery.new(
      '$' + variables.datasource.name,
      |||
        avg(up{} * on (instance) rds_exporter_build_info{instance=~"$instance"})
      |||
    ),

  errors:
    prometheusQuery.new(
      '$' + variables.datasource.name,
      |||
        sum(rds_exporter_errors_total{instance=~"$instance"})
      |||
    ),

  errorsPerMinute:
    prometheusQuery.new(
      '$' + variables.datasource.name,
      |||
        sum(rate(rds_exporter_errors_total{instance=~"$instance"}[1m]) * 60) by (instance)
      |||
    )
    + prometheusQuery.withLegendFormat('{{instance}}'),

  awsAccounts:
    prometheusQuery.new(
      '$' + variables.datasource.name,
      |||
        count(
            sum by (aws_account_id) (rds_instance_info{instance=~"$instance"})
        )
      |||
    )
    + prometheusQuery.withLegendFormat('Total'),

  rdsInstances:
    prometheusQuery.new(
      '$' + variables.datasource.name,
      |||
        sum(rds_instance_info{instance=~"$instance"})
      |||
    )
    + prometheusQuery.withLegendFormat('Total'),

  awsAPICalls:
    prometheusQuery.new(
      '$' + variables.datasource.name,
      |||
        sum(rds_api_call_total{instance=~"$instance"}) by (api)
      |||
    )
    + prometheusQuery.withLegendFormat('{{api}}'),

  awsAPICallsPerMinute:
    prometheusQuery.new(
      '$' + variables.datasource.name,
      |||
        sum(rate(rds_api_call_total{instance=~"$instance"}[5m]) * 60) by (api)
      |||
    )
    + prometheusQuery.withLegendFormat('{{api}}'),
}
