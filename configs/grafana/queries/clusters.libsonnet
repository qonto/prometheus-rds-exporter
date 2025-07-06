local g = import '../g.libsonnet';
local prometheusQuery = g.query.prometheus;
local variables = import '../variables.libsonnet';

{
  all:
    prometheusQuery.new(
      '$' + variables.datasource.name,
      |||
        rds_cluster_info{aws_account_id=~"$aws_account_id",aws_region=~"$aws_region"}
      |||
    )
    + prometheusQuery.withInstant(true)
    + prometheusQuery.withFormat('table'),

  count:
    prometheusQuery.new(
      '$' + variables.datasource.name,
      |||
        count(rds_cluster_info{aws_account_id=~"$aws_account_id",aws_region=~"$aws_region"})
      |||
    ),

  byEngine:
    prometheusQuery.new(
      '$' + variables.datasource.name,
      |||
        sum(rds_cluster_info{aws_account_id=~"$aws_account_id",aws_region=~"$aws_region"}) by (engine)
      |||
    )
    + prometheusQuery.withInstant(true)
    + prometheusQuery.withLegendFormat('{{ engine }}'),
}
