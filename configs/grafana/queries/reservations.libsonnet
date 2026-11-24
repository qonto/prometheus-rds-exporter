local g = import '../g.libsonnet';
local prometheusQuery = g.query.prometheus;
local variables = import '../variables.libsonnet';

// Selectors of all reservations, and of reservations selected by the capacity_reservation_id variable
local selector = 'aws_account_id=~"$aws_account_id",aws_region=~"$aws_region"';
local reservationSelector = selector + ',capacity_reservation_id=~"$capacity_reservation_id"';

{
  // Numeric code of each capacity reservation state, to display states over time.
  // See https://docs.aws.amazon.com/AmazonRDS/latest/AuroraUserGuide/capacity-reservations.getting-started.html#capacity-reservations.states
  stateCodes: {
    active: 1,
    pending: 2,
    assessing: 3,
    scheduled: 4,
    delayed: 5,
    expired: 6,
    cancelled: 7,
    failed: 8,
    unsupported: 9,
  },

  __table:
    prometheusQuery.withFormat('table'),

  __count(selector):
    prometheusQuery.new(
      '$' + variables.datasource.name,
      |||
        count(rds_capacity_reservation_info{%(selector)s,state="active"})
      ||| % { selector: selector }
    ),

  __reserved(selector):
    prometheusQuery.new(
      '$' + variables.datasource.name,
      |||
        sum(rds_capacity_reservation_reserved_average{%(selector)s})
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
    reserved: $.__reserved(reservationSelector),
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
    + prometheusQuery.withRefId('A')
    + self.__table,

  // Days remaining before the end date, NaN when the reservation doesn't expire
  remainingDays:
    prometheusQuery.new(
      '$' + variables.datasource.name,
      |||
        rds_capacity_reservation_info{%(reservationSelector)s}
        * on (aws_account_id, aws_region, capacity_reservation_id) group_left ()
        max by (aws_account_id, aws_region, capacity_reservation_id) (
          (rds_capacity_reservation_end_timestamp_seconds{%(selector)s} - time()) / 86400
          or on (aws_account_id, aws_region, capacity_reservation_id)
          (0 * rds_capacity_reservation_info{%(selector)s} + NaN)
        )
      ||| % { selector: selector, reservationSelector: reservationSelector }
    )
    + prometheusQuery.withInstant(true)
    + prometheusQuery.withRefId('B')
    + self.__table,

  // Usage per capacity reservation and availability zone, with reservation name, engine and instance class
  usage: {

    // Last usage per reservation (all availability zones), named after the reservation, engine and instance class
    current:
      prometheusQuery.new(
        '$' + variables.datasource.name,
        |||
          100 * (
            sum by (aws_account_id, aws_region, capacity_reservation_id) (rds_capacity_reservation_reserved_average{%(reservationSelector)s})
            - sum by (aws_account_id, aws_region, capacity_reservation_id) (rds_capacity_reservation_available_average{%(reservationSelector)s})
          )
          / sum by (aws_account_id, aws_region, capacity_reservation_id) (rds_capacity_reservation_reserved_average{%(reservationSelector)s})
          * on (aws_account_id, aws_region, capacity_reservation_id) group_left (capacity_reservation_name, engine, instance_class)
          rds_capacity_reservation_info{%(reservationSelector)s}
        ||| % { reservationSelector: reservationSelector }
      )
      + prometheusQuery.withInstant(true)
      + prometheusQuery.withLegendFormat('{{ capacity_reservation_name }} ({{ engine }} - {{ instance_class }})'),

    unused:
      prometheusQuery.new(
        '$' + variables.datasource.name,
        |||
          sum by (aws_account_id, aws_region, capacity_reservation_id, availability_zone) (rds_capacity_reservation_available_average{%(reservationSelector)s})
          * on (aws_account_id, aws_region, capacity_reservation_id) group_left (capacity_reservation_name, engine, instance_class)
          rds_capacity_reservation_info{%(reservationSelector)s}
        ||| % { reservationSelector: reservationSelector }
      )
      + prometheusQuery.withLegendFormat('{{ capacity_reservation_name }} in {{ aws_account_id }}.{{ availability_zone }} ({{ engine }} - {{ instance_class }})'),
  },

  // State code of each reservation over time, -1 for an unknown state
  state:
    local states = std.objectFields($.stateCodes);
    local byState = [
      'rds_capacity_reservation_info{%s,state="%s"} * %d' % [reservationSelector, state, $.stateCodes[state]]
      for state in states
    ];
    local unknown = 'rds_capacity_reservation_info{%s,state!~"%s"} * -1' % [reservationSelector, std.join('|', states)];

    prometheusQuery.new(
      '$' + variables.datasource.name,
      |||
        max by (aws_account_id, aws_region, capacity_reservation_id, capacity_reservation_name) (
          %s
        )
      ||| % std.join('\n  or ', byState + [unknown])
    )
    + prometheusQuery.withLegendFormat('{{ capacity_reservation_name }}'),

  // Current number of reservations per state
  countByState:
    prometheusQuery.new(
      '$' + variables.datasource.name,
      |||
        count by (state) (rds_capacity_reservation_info{%(reservationSelector)s})
      ||| % { reservationSelector: reservationSelector }
    )
    + prometheusQuery.withInstant(true)
    + prometheusQuery.withLegendFormat('{{ state }}'),

  // RDS instances using the selected capacity reservations, with the reservation name
  instances:
    prometheusQuery.new(
      '$' + variables.datasource.name,
      |||
        rds_instance_info{%(reservationSelector)s,capacity_reservation_id!=""}
        * on (aws_account_id, aws_region, capacity_reservation_id) group_left (capacity_reservation_name)
        (
          group by (aws_account_id, aws_region, capacity_reservation_id, capacity_reservation_name) (rds_capacity_reservation_info{%(selector)s})
          # Keep instances whose reservation is not collected
          or on (aws_account_id, aws_region, capacity_reservation_id)
          group by (aws_account_id, aws_region, capacity_reservation_id) (rds_instance_info{%(selector)s})
        )
      ||| % { selector: selector, reservationSelector: reservationSelector }
    )
    + prometheusQuery.withInstant(true)
    + self.__table,

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
