local g = import '../g.libsonnet';
local prometheusQuery = g.query.prometheus;
local variables = import '../variables.libsonnet';

{
  __table:
    prometheusQuery.withFormat('table'),

  count:
    prometheusQuery.new(
      '$' + variables.datasource.name,
      |||
        count(rds_capacity_reservation_info{aws_account_id=~"$aws_account_id",aws_region=~"$aws_region",state="active"})
      |||
    ),

  all:
    prometheusQuery.new(
      '$' + variables.datasource.name,
      |||
        rds_capacity_reservation_info{aws_account_id=~"$aws_account_id",aws_region=~"$aws_region",capacity_reservation_id=~"$capacity_reservation_id"}
        * on (aws_account_id, aws_region, capacity_reservation_id) group_left ()
        max by (aws_account_id, aws_region, capacity_reservation_id) (
          rds_capacity_reservation_end_timestamp_seconds{aws_account_id=~"$aws_account_id",aws_region=~"$aws_region"} * 1000
          or on (aws_account_id, aws_region, capacity_reservation_id)
          0 * rds_capacity_reservation_info{aws_account_id=~"$aws_account_id",aws_region=~"$aws_region"}
        )
      |||
    )
    + prometheusQuery.withInstant(true)
    + self.__table,

  usage: {
    ratio:
      prometheusQuery.new(
        '$' + variables.datasource.name,
        |||
          100 * (
            sum by (aws_account_id, aws_region, engine, instance_class, availability_zone) (rds_capacity_reservation_reserved_average{aws_account_id=~"$aws_account_id",aws_region=~"$aws_region"})
            - sum by (aws_account_id, aws_region, engine, instance_class, availability_zone) (rds_capacity_reservation_available_average{aws_account_id=~"$aws_account_id",aws_region=~"$aws_region"})
          )
          / sum by (aws_account_id, aws_region, engine, instance_class, availability_zone) (rds_capacity_reservation_reserved_average{aws_account_id=~"$aws_account_id",aws_region=~"$aws_region"})
        |||
      )
      + prometheusQuery.withLegendFormat('{{ aws_account_id }}:{{ aws_region }} {{ engine }} {{ instance_class }} {{ availability_zone }}'),

    unused:
      prometheusQuery.new(
        '$' + variables.datasource.name,
        |||
          sum by (aws_account_id, aws_region, engine, instance_class, availability_zone) (rds_capacity_reservation_available_average{aws_account_id=~"$aws_account_id",aws_region=~"$aws_region"})
        |||
      )
      + prometheusQuery.withLegendFormat('{{ aws_account_id }}:{{ aws_region }} {{ engine }} {{ instance_class }} {{ availability_zone }}'),
  },

  unused:
    prometheusQuery.new(
      '$' + variables.datasource.name,
      |||
        sum(rds_capacity_reservation_available_average{aws_account_id=~"$aws_account_id",aws_region=~"$aws_region"})
      |||
    ),

  ending:
    prometheusQuery.new(
      '$' + variables.datasource.name,
      |||
        count(
          ((rds_capacity_reservation_end_timestamp_seconds{aws_account_id=~"$aws_account_id",aws_region=~"$aws_region"} - time()) / 86400 < 30 > 0)
          and on (aws_account_id, aws_region, capacity_reservation_id)
          rds_capacity_reservation_info{aws_account_id=~"$aws_account_id",aws_region=~"$aws_region",state="active"}
        )
      |||
    ),

  endingTable:
    prometheusQuery.new(
      '$' + variables.datasource.name,
      |||
        ((rds_capacity_reservation_end_timestamp_seconds{aws_account_id=~"$aws_account_id",aws_region=~"$aws_region"} - time()) / 86400 < 30 > 0)
        * on (aws_account_id, aws_region, capacity_reservation_id) group_left (capacity_reservation_name, engine, instance_class)
        rds_capacity_reservation_info{aws_account_id=~"$aws_account_id",aws_region=~"$aws_region",state="active"}
      |||
    )
    + prometheusQuery.withInstant(true)
    + self.__table,
}
