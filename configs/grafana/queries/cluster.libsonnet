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

    serverless: {
      maxACU:
        prometheusQuery.new(
          '$' + variables.datasource.name,
          |||
            rds_cluster_acu_max_average{aws_account_id="$aws_account_id",aws_region="$aws_region",cluster_identifier="$cluster_identifier"}
          |||
        )
        //        + prometheusQuery.withFormat('table')
        + prometheusQuery.withLegendFormat('Max ACU/instance'),
      minACU:
        prometheusQuery.new(
          '$' + variables.datasource.name,
          |||
            rds_cluster_acu_min_average{aws_account_id="$aws_account_id",aws_region="$aws_region",cluster_identifier="$cluster_identifier"}
          |||
        )
        //        + prometheusQuery.withFormat('table')
        + prometheusQuery.withLegendFormat('Min ACU/instance'),

      currentACU:
        prometheusQuery.new(
          '$' + variables.datasource.name,
          |||
            sum by (cluster_identifier) (
              rds_serverless_instance_acu_average
                * on(dbidentifier)
                  group_left(cluster_identifier)
                  (rds_instance_info{aws_account_id="$aws_account_id",aws_region="$aws_region",cluster_identifier="$cluster_identifier"})
            )
          |||
        )
        + prometheusQuery.withLegendFormat('Current ACU'),

      ACUperInstance:
        prometheusQuery.new(
          '$' + variables.datasource.name,
          |||
            rds_serverless_instance_acu_average
                * on(dbidentifier)
                  group_left(cluster_identifier)
                  (rds_instance_info{aws_account_id="$aws_account_id",aws_region="$aws_region",cluster_identifier="$cluster_identifier"})
          |||
        )
        + prometheusQuery.withLegendFormat('{{dbidentifier}}'),

      ACUUsedPercentage:
        prometheusQuery.new(
          '$' + variables.datasource.name,
          |||
            max by (cluster_identifier) (
              rds_serverless_instance_acu_average
                * on(dbidentifier)
                  group_left(cluster_identifier)
                  (rds_instance_info{aws_account_id="$aws_account_id", aws_region="$aws_region", cluster_identifier="$cluster_identifier"})
            )
            * 100
            / 
            max by (cluster_identifier) (rds_cluster_acu_max_average{aws_account_id="$aws_account_id", aws_region="$aws_region", cluster_identifier="$cluster_identifier"})
          |||
        )
        + prometheusQuery.withLegendFormat('{{cluster_identifier}}'),

    },
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
