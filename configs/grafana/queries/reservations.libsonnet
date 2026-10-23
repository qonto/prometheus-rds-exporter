local g = import '../g.libsonnet';
local prometheusQuery = g.query.prometheus;
local variables = import '../variables.libsonnet';

{
  __table:
    prometheusQuery.withFormat('table'),

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
