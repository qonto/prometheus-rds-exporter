local g = import '../g.libsonnet';
local prometheusQuery = g.query.prometheus;
local variables = import '../variables.libsonnet';

{
  instance: {
    info:
      prometheusQuery.new(
        '$' + variables.datasource.name,
        |||
          rds_instance_info{aws_account_id="$aws_account_id",aws_region="$aws_region",dbidentifier="$dbidentifier"}
        |||
      )
      + prometheusQuery.withFormat('table')
      + prometheusQuery.withInstant(true),

    age:
      prometheusQuery.new(
        '$' + variables.datasource.name,
        |||
          max(rds_instance_age_seconds{aws_account_id="$aws_account_id",aws_region="$aws_region",dbidentifier="$dbidentifier"})
        |||
      )
      + prometheusQuery.withInstant(true),

    status:
      prometheusQuery.new(
        '$' + variables.datasource.name,
        |||
          max(rds_instance_status{aws_account_id="$aws_account_id",aws_region="$aws_region",dbidentifier="$dbidentifier"})
        |||
      )
      + prometheusQuery.withLegendFormat('Status'),

    replicasLag:
      prometheusQuery.new(
        '$' + variables.datasource.name,
        |||
          max(rds_replica_lag_seconds{}) by (dbidentifier)
          * on (dbidentifier)
          group_left(source_dbidentifier)
          max(rds_instance_info{aws_account_id="$aws_account_id",aws_region="$aws_region",source_dbidentifier="$dbidentifier"}) by (dbidentifier)
        |||
      )
      + prometheusQuery.withLegendFormat('{{dbidentifier}}'),
    autoscalingUsage:
      prometheusQuery.new(
        '$' + variables.datasource.name,
        |||
          max(rds_allocated_storage_bytes{aws_account_id="$aws_account_id",aws_region="$aws_region",dbidentifier="$dbidentifier"} * 100 / rds_max_allocated_storage_bytes{aws_account_id="$aws_account_id",aws_region="$aws_region",dbidentifier="$dbidentifier"})
        |||
      )
      + prometheusQuery.withLegendFormat('Usage'),

    cpu: {
      count:
        prometheusQuery.new(
          '$' + variables.datasource.name,
          |||
            max(
                max(rds_instance_vcpu_average{}) by (instance_class)
                * on (instance_class)
                group_left(aws_account_id,aws_region,dbidentifier)
                max(rds_instance_info{aws_account_id="$aws_account_id",aws_region="$aws_region",dbidentifier="$dbidentifier"}) by (aws_account_id, aws_region, dbidentifier, instance_class)
            )
          |||
        )
        + prometheusQuery.withLegendFormat('Number of vCPU'),
      usagePercent:
        prometheusQuery.new(
          '$' + variables.datasource.name,
          |||
            max(rds_cpu_usage_percent_average{aws_account_id="$aws_account_id",aws_region="$aws_region",dbidentifier="$dbidentifier"}) by (dbidentifier)
          |||
        )
        + prometheusQuery.withLegendFormat('{{dbidentifier}}'),
      wait:
        prometheusQuery.new(
          '$' + variables.datasource.name,
          |||
            max(rds_dbload_cpu_average{aws_account_id="$aws_account_id",aws_region="$aws_region",dbidentifier="$dbidentifier"})
          |||
        )
        + prometheusQuery.withLegendFormat('CPU wait'),
      nonWait:
        prometheusQuery.new(
          '$' + variables.datasource.name,
          |||
            max(rds_dbload_noncpu_average{aws_account_id="$aws_account_id",aws_region="$aws_region",dbidentifier="$dbidentifier"})
          |||
        )
        + prometheusQuery.withLegendFormat('Non CPU wait'),
    },
    storage: {
      allocated:
        prometheusQuery.new(
          '$' + variables.datasource.name,
          |||
            max(rds_allocated_storage_bytes{aws_account_id="$aws_account_id",aws_region="$aws_region",dbidentifier="$dbidentifier"})
          |||
        )
        + prometheusQuery.withLegendFormat('Allocated'),

      free:
        prometheusQuery.new(
          '$' + variables.datasource.name,
          |||
            max(rds_free_storage_bytes{aws_account_id="$aws_account_id",aws_region="$aws_region",dbidentifier="$dbidentifier"})
          |||
        )
        + prometheusQuery.withLegendFormat('Free'),

      logs:
        prometheusQuery.new(
          '$' + variables.datasource.name,
          |||
            max(rds_instance_log_files_size_bytes{aws_account_id="$aws_account_id",aws_region="$aws_region",dbidentifier="$dbidentifier"})
          |||
        )
        + prometheusQuery.withLegendFormat('Logs'),

      replicationSlots:
        prometheusQuery.new(
          '$' + variables.datasource.name,
          |||
            max(rds_replication_slot_disk_usage_bytes{aws_account_id="$aws_account_id",aws_region="$aws_region",dbidentifier="$dbidentifier"})
          |||
        )
        + prometheusQuery.withLegendFormat('Replication slots'),

      wal:
        prometheusQuery.new(
          '$' + variables.datasource.name,
          |||
            max(rds_transaction_logs_disk_usage_bytes{aws_account_id="$aws_account_id",aws_region="$aws_region",dbidentifier="$dbidentifier"})
          |||
        )
        + prometheusQuery.withLegendFormat('WAL'),
      usagePercent:
        prometheusQuery.new(
          '$' + variables.datasource.name,
          |||
            100 - max(
              rds_free_storage_bytes{aws_account_id="$aws_account_id",aws_region="$aws_region",dbidentifier="$dbidentifier"}
              * 100
              / rds_allocated_storage_bytes{aws_account_id="$aws_account_id",aws_region="$aws_region",dbidentifier="$dbidentifier"}
            )
          |||
        )
        + prometheusQuery.withLegendFormat('{{dbidentifier}}'),
    },
    replicas: {
      lag:
        prometheusQuery.new(
          '$' + variables.datasource.name,
          |||
            max(rds_replica_lag_seconds{aws_account_id="$aws_account_id",aws_region="$aws_region",dbidentifier="$dbidentifier"})
          |||
        ),

      count:
        prometheusQuery.new(
          '$' + variables.datasource.name,
          |||
            count(rds_instance_info{aws_account_id="$aws_account_id",aws_region="$aws_region",source_dbidentifier="$dbidentifier"})
          |||
        ),
    },
    disk: {
      iops: {
        read:
          prometheusQuery.new(
            '$' + variables.datasource.name,
            |||
              max(rds_read_iops_average{aws_account_id="$aws_account_id",aws_region="$aws_region",dbidentifier="$dbidentifier"})
            |||
          )
          + prometheusQuery.withLegendFormat('Read'),

        write:
          prometheusQuery.new(
            '$' + variables.datasource.name,
            |||
              max(rds_write_iops_average{aws_account_id="$aws_account_id",aws_region="$aws_region",dbidentifier="$dbidentifier"})
            |||
          )
          + prometheusQuery.withLegendFormat('Write'),

        max:
          prometheusQuery.new(
            '$' + variables.datasource.name,
            |||
              max(rds_max_disk_iops_average{aws_account_id="$aws_account_id",aws_region="$aws_region",dbidentifier="$dbidentifier"})
            |||
          )
          + prometheusQuery.withLegendFormat('Max'),

        usage:
          prometheusQuery.new(
            '$' + variables.datasource.name,
            |||
              max(rds_read_iops_average{aws_account_id="$aws_account_id",aws_region="$aws_region",dbidentifier="$dbidentifier"})
              + max(rds_write_iops_average{aws_account_id="$aws_account_id",aws_region="$aws_region",dbidentifier="$dbidentifier"})
            |||
          )
          + prometheusQuery.withLegendFormat('Usage'),

        usagePercent:
          prometheusQuery.new(
            '$' + variables.datasource.name,
            |||
              max(
                (
                  rds_read_iops_average{aws_account_id="$aws_account_id",aws_region="$aws_region",dbidentifier="$dbidentifier"}
                  + rds_write_iops_average{aws_account_id="$aws_account_id",aws_region="$aws_region",dbidentifier="$dbidentifier"}
                )
                * 100
                / rds_max_disk_iops_average{aws_account_id="$aws_account_id",aws_region="$aws_region",dbidentifier="$dbidentifier"}
              )
            |||
          )
          + prometheusQuery.withLegendFormat('Usage'),

        allocated:
          prometheusQuery.new(
            '$' + variables.datasource.name,
            |||
              max(rds_allocated_disk_iops_average{dbidentifier="$dbidentifier"})
            |||
          )
          + prometheusQuery.withLegendFormat('Allocated'),

        instanceTypeBurst:
          prometheusQuery.new(
            '$' + variables.datasource.name,
            |||
              max(
                  max(rds_instance_max_iops_average{}) by (instance_class)
                  * on (instance_class)
                  group_left(aws_account_id,aws_region,dbidentifier)
                  max(rds_instance_info{aws_account_id="$aws_account_id",aws_region="$aws_region",dbidentifier="$dbidentifier"}) by (aws_account_id, aws_region, dbidentifier, instance_class)
              ) by (instance_class)
            |||
          )
          + prometheusQuery.withLegendFormat('{{instance_class}} burst'),

        instanceTypeBaseline:
          prometheusQuery.new(
            '$' + variables.datasource.name,
            |||
              max(
                  max(rds_instance_baseline_iops_average{}) by (instance_class)
                  * on (instance_class)
                  group_left(aws_account_id,aws_region,dbidentifier)
                  max(rds_instance_info{aws_account_id="$aws_account_id",aws_region="$aws_region",dbidentifier="$dbidentifier"}) by (aws_account_id, aws_region, dbidentifier, instance_class)
              ) by (instance_class)
            |||
          )
          + prometheusQuery.withLegendFormat('{{instance_class}} baseline'),
      },
      throughput: {
        read:
          prometheusQuery.new(
            '$' + variables.datasource.name,
            |||
              max(rds_read_throughput_bytes{dbidentifier="$dbidentifier"})
            |||
          )
          + prometheusQuery.withLegendFormat('Read'),

        write:
          prometheusQuery.new(
            '$' + variables.datasource.name,
            |||
              max(rds_write_throughput_bytes{dbidentifier="$dbidentifier"})
            |||
          )
          + prometheusQuery.withLegendFormat('Write'),

        usage:
          prometheusQuery.new(
            '$' + variables.datasource.name,
            |||
              max(rds_read_throughput_bytes{dbidentifier="$dbidentifier"})
              + max(rds_write_throughput_bytes{dbidentifier="$dbidentifier"})
            |||
          )
          + prometheusQuery.withLegendFormat('Usage'),

        max:
          prometheusQuery.new(
            '$' + variables.datasource.name,
            |||
              max(rds_max_storage_throughput_bytes{dbidentifier="$dbidentifier"})
            |||
          )
          + prometheusQuery.withLegendFormat('Max'),

        allocated:
          prometheusQuery.new(
            '$' + variables.datasource.name,
            |||
              max(rds_allocated_disk_throughput_bytes{dbidentifier="$dbidentifier"})
            |||
          )
          + prometheusQuery.withLegendFormat('Allocated'),

        instanceTypeBurst:
          prometheusQuery.new(
            '$' + variables.datasource.name,
            |||
              max(
              max(rds_instance_max_throughput_bytes{}) by (instance_class)
              * on (instance_class)
              group_left(aws_account_id,aws_region,dbidentifier)
              max(rds_instance_info{aws_account_id="$aws_account_id",aws_region="$aws_region",dbidentifier="$dbidentifier"}) by (aws_account_id, aws_region, dbidentifier, instance_class)
              ) by (instance_class)
            |||
          )
          + prometheusQuery.withLegendFormat('{{instance_class}} burst'),

        instanceTypeBaseline:
          prometheusQuery.new(
            '$' + variables.datasource.name,
            |||
              max(
              max(rds_instance_baseline_throughput_bytes{}) by (instance_class)
              * on (instance_class)
              group_left(aws_account_id,aws_region,dbidentifier)
              max(rds_instance_info{aws_account_id="$aws_account_id",aws_region="$aws_region",dbidentifier="$dbidentifier"}) by (aws_account_id, aws_region, dbidentifier, instance_class)
              ) by (instance_class)
            |||
          )
          + prometheusQuery.withLegendFormat('{{instance_class}} baseline'),
      },
    },
    memory: {
      max:
        prometheusQuery.new(
          '$' + variables.datasource.name,
          |||
            max(
                max(rds_instance_memory_bytes{}) by (instance_class)
                * on (instance_class)
                group_left(dbidentifier) max(rds_instance_info{aws_account_id="$aws_account_id",aws_region="$aws_region",dbidentifier="$dbidentifier"}) by (dbidentifier, instance_class)
            )
          |||
        )
        + prometheusQuery.withLegendFormat('Max'),

      freeable:
        prometheusQuery.new(
          '$' + variables.datasource.name,
          |||
            max(rds_freeable_memory_bytes{aws_account_id="$aws_account_id",aws_region="$aws_region",dbidentifier="$dbidentifier"})
          |||
        )
        + prometheusQuery.withLegendFormat('Freeable'),

      usagePercent:
        prometheusQuery.new(
          '$' + variables.datasource.name,
          |||
            100
            - sum(rds_freeable_memory_bytes{aws_account_id="$aws_account_id",aws_region="$aws_region",dbidentifier="$dbidentifier"}) by (dbidentifier)
              * 100
              / sum(rds_instance_memory_bytes{}
              * on (instance_class) group_left(dbidentifier) rds_instance_info{aws_account_id="$aws_account_id",aws_region="$aws_region",dbidentifier="$dbidentifier"}
            )
            by (dbidentifier)
          |||
        )
        + prometheusQuery.withLegendFormat('{{dbidentifier}}'),

      usage:
        prometheusQuery.new(
          '$' + variables.datasource.name,
          |||
            max(
              max(rds_instance_memory_bytes{}) by (instance_class)
              * on (instance_class)
              group_left(aws_account_id,aws_region,dbidentifier)
              max(rds_instance_info{aws_account_id="$aws_account_id",aws_region="$aws_region",dbidentifier="$dbidentifier"}) by (aws_account_id, aws_region, dbidentifier, instance_class)
            )
          |||
        )
        + prometheusQuery.withLegendFormat('{{dbidentifier}}'),
      swap:
        prometheusQuery.new(
          '$' + variables.datasource.name,
          |||
            max(rds_swap_usage_bytes{dbidentifier="$dbidentifier"}) by (dbidentifier)
          |||
        )
        + prometheusQuery.withLegendFormat('{{dbidentifier}}'),

    },
    backup: {
      retention:
        prometheusQuery.new(
          '$' + variables.datasource.name,
          |||
            max(rds_backup_retention_period_seconds{aws_account_id="$aws_account_id",aws_region="$aws_region",dbidentifier="$dbidentifier"})
          |||
        )
        + prometheusQuery.withLegendFormat('{{dbidentifier}}'),
    },
    databaseConnections:
      prometheusQuery.new(
        '$' + variables.datasource.name,
        |||
          max(rds_database_connections_average{aws_account_id="$aws_account_id",aws_region="$aws_region",dbidentifier="$dbidentifier"}) by (dbidentifier)
        |||
      )
      + prometheusQuery.withLegendFormat('{{dbidentifier}}'),
  },
  postgresql: {
    maxUsedTransaction:
      prometheusQuery.new(
        '$' + variables.datasource.name,
        |||
          max(rds_maximum_used_transaction_ids_average{aws_account_id="$aws_account_id",aws_region="$aws_region",dbidentifier="$dbidentifier"}) by (dbidentifier)
        |||
      )
      + prometheusQuery.withLegendFormat('{{dbidentifier}}'),
  },
}
