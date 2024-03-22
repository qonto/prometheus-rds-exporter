local g = import '../g.libsonnet';
local prometheusQuery = g.query.prometheus;
local variables = import '../variables.libsonnet';

{
  __table:
    prometheusQuery.withFormat('table'),

  all:
    prometheusQuery.new(
      '$' + variables.datasource.name,
      |||
        rds_instance_info{aws_account_id=~"$aws_account_id",aws_region=~"$aws_region"} * on (dbidentifier) group_left max(rds_allocated_storage_bytes{}) by (dbidentifier)
      |||
    )
    + prometheusQuery.withInstant(true)
    + prometheusQuery.withFormat('table'),

  count:
    prometheusQuery.new(
      '$' + variables.datasource.name,
      |||
        count(rds_instance_info{aws_account_id=~"$aws_account_id",aws_region=~"$aws_region"})
      |||
    ),

  instancesWithPendingMaintenance:
    prometheusQuery.new(
      '$' + variables.datasource.name,
      |||
        count(rds_instance_info{aws_account_id=~"$aws_account_id",aws_region=~"$aws_region",pending_maintenance!="no"} > 0)
      |||
    ),

  instancesWithPendingMaintenanceTable:
    prometheusQuery.new(
      '$' + variables.datasource.name,
      |||
        rds_instance_info{aws_account_id=~"$aws_account_id",aws_region=~"$aws_region",pending_maintenance!="no"}
      |||
    )
    + prometheusQuery.withInstant(true)
    + self.__table,

  instancesWithPendingModification:
    prometheusQuery.new(
      '$' + variables.datasource.name,
      |||
        count(rds_instance_info{aws_account_id=~"$aws_account_id",aws_region=~"$aws_region",pending_modified_values!="false"})
      |||
    ),

  instancesWithPendingModificationTable:
    prometheusQuery.new(
      '$' + variables.datasource.name,
      |||
        rds_instance_info{aws_account_id=~"$aws_account_id",aws_region=~"$aws_region",pending_modified_values!="false"}
      |||
    )
    + prometheusQuery.withInstant(true)
    + self.__table,

  instancesWithDeprecatedCertificate:
    prometheusQuery.new(
      '$' + variables.datasource.name,
      |||
        count(rds_instance_info{aws_account_id=~"$aws_account_id",aws_region=~"$aws_region", ca_certificate_identifier!="rds-ca-rsa2048-g1"})
      |||
    ),

  instances: {
    total:
      prometheusQuery.new(
        '$' + variables.datasource.name,
        |||
          count(
              count(rds_instance_info{aws_account_id=~"$aws_account_id",aws_region=~"$aws_region"}) by (aws_account_id, aws_region, dbidentifier)
          ) by (aws_account_id, aws_region)
        |||
      )
      + prometheusQuery.withLegendFormat('Total {{ aws_account_id }}:{{ aws_region }}'),

    max:
      prometheusQuery.new(
        '$' + variables.datasource.name,
        |||
          sum(
              max(rds_quota_max_dbinstances_average{aws_account_id=~"$aws_account_id",aws_region=~"$aws_region"}) by (aws_account_id, aws_region)
          ) by (aws_account_id, aws_region)
        |||
      )
      + prometheusQuery.withLegendFormat('Max {{ aws_account_id }}:{{ aws_region }}'),

    ratio:
      prometheusQuery.new(
        '$' + variables.datasource.name,
        |||
          count(
              count(rds_instance_info{aws_account_id=~"$aws_account_id",aws_region=~"$aws_region"}) by (aws_account_id, aws_region, dbidentifier)
          ) by (aws_account_id, aws_region)
          * 100
          /
          sum(
              max(rds_quota_max_dbinstances_average{aws_account_id=~"$aws_account_id",aws_region=~"$aws_region"}) by (aws_account_id, aws_region)
          ) by (aws_account_id, aws_region)
        |||
      )
      + prometheusQuery.withLegendFormat('{{ aws_account_id }}:{{ aws_region }}'),
  },

  storage: {
    total:
      prometheusQuery.new(
        '$' + variables.datasource.name,
        |||
          sum(
              max(rds_allocated_storage_bytes{aws_account_id=~"$aws_account_id",aws_region=~"$aws_region"}) by (aws_account_id, aws_region, dbidentifier)
          ) by (aws_account_id, aws_region)
        |||
      )
      + prometheusQuery.withLegendFormat('Total {{ aws_account_id }}:{{ aws_region }}'),

    max:
      prometheusQuery.new(
        '$' + variables.datasource.name,
        |||
          sum(
              max(rds_quota_total_storage_bytes{aws_account_id=~"$aws_account_id",aws_region=~"$aws_region"}) by (aws_account_id, aws_region)
          ) by (aws_account_id, aws_region)
        |||
      )
      + prometheusQuery.withLegendFormat('Max {{ aws_account_id }}:{{ aws_region }}'),

    ratio:
      prometheusQuery.new(
        '$' + variables.datasource.name,
        |||
          sum(
              max(rds_allocated_storage_bytes{aws_account_id=~"$aws_account_id",aws_region=~"$aws_region"}) by (aws_account_id, aws_region, dbidentifier)
          ) by (aws_account_id, aws_region)
          * 100
          /
          sum(
              max(rds_quota_total_storage_bytes{aws_account_id=~"$aws_account_id",aws_region=~"$aws_region"}) by (aws_account_id, aws_region)
          ) by (aws_account_id, aws_region)
        |||
      )
      + prometheusQuery.withLegendFormat('{{ aws_account_id }}:{{ aws_region }}'),
  },
}
