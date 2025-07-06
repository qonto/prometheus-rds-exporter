local g = import '../g.libsonnet';
local prometheusQuery = g.query.prometheus;
local variables = import '../variables.libsonnet';

{
  cluster: {
    info:
      prometheusQuery.new(
        '$' + variables.datasource.name,
        |||
          rds_cluster_info{aws_account_id="$aws_account_id",aws_region="$aws_region",cluster_identifier="$cluster_identifier"}
        |||
      )
      + prometheusQuery.withFormat('table')
      + prometheusQuery.withInstant(true),
  },

  instances: {
    count:
      prometheusQuery.new(
        '$' + variables.datasource.name,
        |||
          count(rds_instance_info{aws_account_id=~"$aws_account_id",aws_region=~"$aws_region",cluster_identifier=~"$cluster_identifier"})
        |||
      ),

    all:
      prometheusQuery.new(
        '$' + variables.datasource.name,
        |||
          rds_instance_info{aws_account_id=~"$aws_account_id",aws_region=~"$aws_region",cluster_identifier=~"$cluster_identifier"}
        |||
      )
      + prometheusQuery.withInstant(true)
      + prometheusQuery.withFormat('table'),

    byClassType:
      prometheusQuery.new(
        '$' + variables.datasource.name,
        |||
          sum(rds_instance_info{aws_account_id=~"$aws_account_id",aws_region=~"$aws_region",cluster_identifier=~"$cluster_identifier"}) by (instance_class)
        |||
      )
      + prometheusQuery.withInstant(true)
      + prometheusQuery.withLegendFormat('{{ instance_class }}'),
  },
}
