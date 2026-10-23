local g = import '../g.libsonnet';
local prometheusQuery = g.query.prometheus;
local variables = import '../variables.libsonnet';

// Selectors of all reservations, and of reservations selected by the capacity_reservation_id variable
local selector = 'aws_account_id=~"$aws_account_id",aws_region=~"$aws_region"';
local reservationSelector = selector + ',capacity_reservation_id=~"$capacity_reservation_id"';

{
  __table:
    prometheusQuery.withFormat('table'),

  __count(selector):
    prometheusQuery.new(
      '$' + variables.datasource.name,
      |||
        count(rds_capacity_reservation_info{%(selector)s,state="active"})
      ||| % { selector: selector }
    ),

  __unused(selector):
    prometheusQuery.new(
      '$' + variables.datasource.name,
      |||
        sum(rds_capacity_reservation_available_average{%(selector)s})
      ||| % { selector: selector }
    ),

  __ending(selector):
    prometheusQuery.new(
      '$' + variables.datasource.name,
      |||
        count(
          ((rds_capacity_reservation_end_timestamp_seconds{%(selector)s} - time()) / 86400 < 30 > 0)
          and on (aws_account_id, aws_region, capacity_reservation_id)
          rds_capacity_reservation_info{%(selector)s,state="active"}
        )
      ||| % { selector: selector }
    ),

  count: self.__count(selector),
  unused: self.__unused(selector),
  ending: self.__ending(selector),

  // Queries filtered by the capacity_reservation_id variable
  byReservation: {
    count: $.__count(reservationSelector),
    unused: $.__unused(reservationSelector),
    ending: $.__ending(reservationSelector),
  },

  all:
    prometheusQuery.new(
      '$' + variables.datasource.name,
      |||
        rds_capacity_reservation_info{%(reservationSelector)s}
        * on (aws_account_id, aws_region, capacity_reservation_id) group_left ()
        max by (aws_account_id, aws_region, capacity_reservation_id) (
          rds_capacity_reservation_end_timestamp_seconds{%(selector)s} * 1000
          or on (aws_account_id, aws_region, capacity_reservation_id)
          0 * rds_capacity_reservation_info{%(selector)s}
        )
      ||| % { selector: selector, reservationSelector: reservationSelector }
    )
    + prometheusQuery.withInstant(true)
    + self.__table,

  // Usage per capacity reservation and availability zone, with reservation name, engine and instance class
  usage: {
    local legend = '{{ aws_account_id }}:{{ aws_region }} {{ capacity_reservation_name }} {{ availability_zone }} ({{ engine }} {{ instance_class }})',

    ratio:
      prometheusQuery.new(
        '$' + variables.datasource.name,
        |||
          100 * (
            sum by (aws_account_id, aws_region, capacity_reservation_id, availability_zone) (rds_capacity_reservation_reserved_average{%(reservationSelector)s})
            - sum by (aws_account_id, aws_region, capacity_reservation_id, availability_zone) (rds_capacity_reservation_available_average{%(reservationSelector)s})
          )
          / sum by (aws_account_id, aws_region, capacity_reservation_id, availability_zone) (rds_capacity_reservation_reserved_average{%(reservationSelector)s})
          * on (aws_account_id, aws_region, capacity_reservation_id) group_left (capacity_reservation_name, engine, instance_class)
          rds_capacity_reservation_info{%(reservationSelector)s}
        ||| % { reservationSelector: reservationSelector }
      )
      + prometheusQuery.withLegendFormat(legend),

    unused:
      prometheusQuery.new(
        '$' + variables.datasource.name,
        |||
          sum by (aws_account_id, aws_region, capacity_reservation_id, availability_zone) (rds_capacity_reservation_available_average{%(reservationSelector)s})
          * on (aws_account_id, aws_region, capacity_reservation_id) group_left (capacity_reservation_name, engine, instance_class)
          rds_capacity_reservation_info{%(reservationSelector)s}
        ||| % { reservationSelector: reservationSelector }
      )
      + prometheusQuery.withLegendFormat(legend),
  },

  endingTable:
    prometheusQuery.new(
      '$' + variables.datasource.name,
      |||
        ((rds_capacity_reservation_end_timestamp_seconds{%(selector)s} - time()) / 86400 < 30 > 0)
        * on (aws_account_id, aws_region, capacity_reservation_id) group_left (capacity_reservation_name, engine, instance_class)
        rds_capacity_reservation_info{%(selector)s,state="active"}
      ||| % { selector: selector }
    )
    + prometheusQuery.withInstant(true)
    + self.__table,
}
