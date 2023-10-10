# Prometheus RDS exporter

Are you ready to take your AWS RDS monitoring to the next level? Say hello to prometheus-rds-exporter, your ultimate solution for comprehensive, real-time insights into your Amazon RDS instances!

Built by SRE Engineers, designed for production: Meticulously crafted by a team of Site Reliability Engineers with years of hands-on experience in managing RDS production systems. Trust in their expertise to supercharge your monitoring.

It collect key metrics about:
- Hardware resource usage
- Underlying EC2 instances hard limits
- Pending AWS RDS maintenances
- Pending modifications
- Logs size
- RDS quota usage information

### Key metrics

📊 Advanced Metrics: Gain deep visibility with advanced metrics for AWS RDS. Monitor performance, query efficiency, and resource utilization like never before.

🧩 AWS Quotas Insights: Stay in control with real-time information about AWS quotas. Ensure you never hit limits unexpectedly.

💡 Hard Limits visibility: Know the hard limits of EC2 instance used by RDS and manage your resources effectively.

🔔 Alerting at Your Fingertips: Easily set up Prometheus alerting rules to stay informed of critical events, ensuring you're always ahead of issues.

🛠️ Simple Setup: Getting started is a breeze! Our clear documentation and examples will have you up and running in no time.

📈 Scalable and Reliable: Prometheus-RDS Exporter scales with your AWS infrastructure, providing reliable monitoring even as you grow.

🌐 Community-Driven: Join a vibrant community of users and contributors. Collaborate, share knowledge, and shape the future of AWS RDS monitoring together.

🚀 When combined with [prometheus-community/postgres_exporter](https://github.com/prometheus-community/postgres_exporter), it provides a production ready monitoring framework for RDS PostgreSQL.

## Metrics

| Name | Labels | Description |
| ---- | ------ | ----------- |
| rds_allocated_storage_bytes | `aws_account_id`, `aws_region`, `dbidentifier` | Allocated storage |
| rds_api_call_total | `api`, `aws_account_id`, `aws_region` | Number of call to AWS API |
| rds_backup_retention_period_seconds | `aws_account_id`, `aws_region`, `dbidentifier` | Automatic DB snapshots retention period |
| rds_cpu_usage_percent_average | `aws_account_id`, `aws_region`, `dbidentifier` | Instance CPU used |
| rds_database_connections_average | `aws_account_id`, `aws_region`, `dbidentifier` | The number of client network connections to the database instance |
| rds_dbload_average | `aws_account_id`, `aws_region`, `dbidentifier` | Number of active sessions for the DB engine |
| rds_dbload_cpu_average | `aws_account_id`, `aws_region`, `dbidentifier` | Number of active sessions where the wait event type is CPU |
| rds_dbload_noncpu_average | `aws_account_id`, `aws_region`, `dbidentifier` | Number of active sessions where the wait event type is not CPU |
| rds_exporter_build_info | `build_date`, `commit_sha`, `version` | A metric with constant '1' value labeled by version from which exporter was built |
| rds_exporter_errors_total | | Total number of errors encountered by the exporter |
| rds_free_storage_bytes | `aws_account_id`, `aws_region`, `dbidentifier` | Free storage on the instance |
| rds_freeable_memory_bytes | `aws_account_id`, `aws_region`, `dbidentifier` | Amount of available random access memory. For MariaDB, MySQL, Oracle, and PostgreSQL DB instances, this metric reports the value of the MemAvailable field of /proc/meminfo |
| rds_instance_info | `aws_account_id`, `aws_region`, `dbi_resource_id`, `dbidentifier`, `deletion_protection`, `engine`, `engine_version`, `instance_class`, `multi_az`, `pending_maintenance`, `pending_modified_values`, `role`, `source_dbidentifier`, `storage_type` | RDS instance information |
| rds_instance_log_files_size_bytes | `aws_account_id`, `aws_region`, `dbidentifier` | Total of log files on the instance |
| rds_instance_max_iops_average | `aws_account_id`, `aws_region`, `dbidentifier` | Maximum IOPS of underlying EC2 instance |
| rds_instance_max_throughput_bytes | `aws_account_id`, `aws_region`, `dbidentifier` | Maximum throughput of underlying EC2 instance |
| rds_instance_memory_bytes | `aws_account_id`, `aws_region`, `dbidentifier` | Instance memory |
| rds_instance_status | `aws_account_id`, `aws_region`, `dbidentifier` | Instance status (1: ok, 0: can't scrap metrics) |
| rds_instance_vcpu_average | `aws_account_id`, `aws_region`, `dbidentifier` | Total vCPU for this isntance class |
| rds_max_allocated_storage_bytes | `aws_account_id`, `aws_region`, `dbidentifier` | Upper limit in gibibytes to which Amazon RDS can automatically scale the storage of the DB instance |
| rds_max_disk_iops_average | `aws_account_id`, `aws_region`, `dbidentifier` | Max IOPS for the instance |
| rds_max_storage_throughput_bytes | `aws_account_id`, `aws_region`, `dbidentifier` | Max storage throughput |
| rds_maximum_used_transaction_ids_average | `aws_account_id`, `aws_region`, `dbidentifier` | Maximum transaction IDs that have been used. Applies to only PostgreSQL |
| rds_quota_max_dbinstances_average | `aws_account_id`, `aws_region` | Maximum number of RDS instances allowed in the AWS account |
| rds_quota_maximum_db_instance_snapshots_average | `aws_account_id`, `aws_region` | Maximum number of manual DB instance snapshots |
| rds_quota_total_storage_bytes | `aws_account_id`, `aws_region` | Maximum total storage for all DB instances |
| rds_read_iops_average | `aws_account_id`, `aws_region`, `dbidentifier` | Average number of disk read I/O operations per second |
| rds_read_throughput_bytes | `aws_account_id`, `aws_region`, `dbidentifier` | Average number of bytes read from disk per second |
| rds_replica_lag_seconds | `aws_account_id`, `aws_region`, `dbidentifier` | For read replica configurations, the amount of time a read replica DB instance lags behind the source DB instance. Applies to MariaDB, Microsoft SQL Server, MySQL, Oracle, and PostgreSQL read replicas |
| rds_replication_slot_disk_usage_average | `aws_account_id`, `aws_region`, `dbidentifier` | Disk space used by replication slot files. Applies to PostgreSQL |
| rds_swap_usage_bytes | `aws_account_id`, `aws_region`, `dbidentifier` | Amount of swap space used on the DB instance. This metric is not available for SQL Server |
| rds_usage_allocated_storage_average | `aws_account_id`, `aws_region` | Total storage used by AWS RDS instances |
| rds_usage_db_instances_average | `aws_account_id`, `aws_region` | AWS RDS instance count |
| rds_usage_manual_snapshots_average | `aws_account_id`, `aws_region` | Manual snapshots count |
| rds_write_iops_average | `aws_account_id`, `aws_region`, `dbidentifier` | Average number of disk write I/O operations per second |
| rds_write_throughput_bytes | `aws_account_id`, `aws_region`, `dbidentifier` | Average number of bytes written to disk per second |
| up | | Was the last scrape of RDS successful |

<details>
  <summary>Standard Go and Prometheus metrics are also available</summary>

| Name   | Labels | Description |
| ------ | -------| ----------- |
| go_gc_duration_seconds | `quantile` | A summary of the pause duration of garbage collection cycles. |
| go_goroutines | | Number of goroutines that currently exist. |
| go_info | `version` | Information about the Go environment. |
| go_memstats_alloc_bytes | | Number of bytes allocated and still in use. |
| go_memstats_alloc_bytes_total | | Total number of bytes allocated, even if freed. |
| go_memstats_buck_hash_sys_bytes | | Number of bytes used by the profiling bucket hash table. |
| go_memstats_frees_total | | Total number of frees. |
| go_memstats_gc_sys_bytes | | Number of bytes used for garbage collection system metadata. |
| go_memstats_heap_alloc_bytes | | Number of heap bytes allocated and still in use. |
| go_memstats_heap_idle_bytes | | Number of heap bytes waiting to be used. |
| go_memstats_heap_inuse_bytes | | Number of heap bytes that are in use. |
| go_memstats_heap_objects | | Number of allocated objects. |
| go_memstats_heap_released_bytes | | Number of heap bytes released to OS. |
| go_memstats_heap_sys_bytes | | Number of heap bytes obtained from system. |
| go_memstats_last_gc_time_seconds | | Number of seconds since 1970 of last garbage collection. |
| go_memstats_lookups_total | | Total number of pointer lookups. |
| go_memstats_mallocs_total | | Total number of mallocs. |
| go_memstats_mcache_inuse_bytes | | Number of bytes in use by mcache structures. |
| go_memstats_mcache_sys_bytes | | Number of bytes used for mcache structures obtained from system. |
| go_memstats_mspan_inuse_bytes | | Number of bytes in use by mspan structures. |
| go_memstats_mspan_sys_bytes | | Number of bytes used for mspan structures obtained from system. |
| go_memstats_next_gc_bytes | | Number of heap bytes when next garbage collection will take place. |
| go_memstats_other_sys_bytes | | Number of bytes used for other system allocations. |
| go_memstats_stack_inuse_bytes | | Number of bytes in use by the stack allocator. |
| go_memstats_stack_sys_bytes | | Number of bytes obtained from system for stack allocator. |
| go_memstats_sys_bytes | | Number of bytes obtained from system. |
| go_threads | | Number of OS threads created. |
| promhttp_metric_handler_requests_in_flight | | Current number of scrapes being served. |
| promhttp_metric_handler_requests_total | `code` | Total number of scrapes by HTTP status code. |

</details>

## Dashboard

Grafana dashoards are available on Grafana labs:

<table>
<tr>
<td>

![Instances overview](docs/screenshots/instances-overview.png)

<a href="https://grafana.com/grafana/dashboards/19647-rds-instances-overview/">RDS instances overview</a> (ID `19647`)
</td>
<td>

![Instance details](docs/screenshots/instance-details.png)

<a href="https://grafana.com/grafana/dashboards/19646-rds-instance-details/">RDS instance details</a> (ID: `19646`)
</td>
<td>

![RDS exporters](docs/screenshots/rds-exporter.png)

<a href="https://grafana.com/grafana/dashboards/19679-rds-exporter/">RDS exporters</a> (ID: `19679`)
</td>
</tr>
</table>

## Configuration

Configuration could be defined in `.prometheus-rds-exporter.yaml` or environment variables (format `PROMETHEUS_RDS_EXPORTER_<PARAMETER_NAME>`).

| Parameter | Description | Default |
| --- | --- | --- |
| aws-assume-role-arn | AWS IAM ARN role to assume to fetch metrics | |
| aws-assume-role-session | AWS assume role session name | prometheus-rds-exporter |
| debug | Enable debug mode | |
| listen-address | Address to listen on for web interface | :9043 |
| log-format | Log format (`text` or `json`) | json |
| metrics-path | Path under which to expose metrics | /metrics |

Configuration parameters priorities:

1. `$HOME/.prometheus-rds-exporter.yaml` file
2. `.prometheus-rds-exporter.yaml` file
3. Environment variables
4. Command line flags

### AWS authentication

Prometheus RDS exporter needs read only AWS IAM permissions to fetch metrics from AWS RDS, CloudWatch, EC2 and ServiceQuota AWS APIs.

Standard AWS authentication methods (AWS credentials, SSO and assume role), see https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-envvars.html.

If you are running on [AWS EKS](), we strongly recommend to use [IRSA](https://docs.aws.amazon.com/eks/latest/userguide/iam-roles-for-service-accounts.html)

Minimal required IAM permissions:

```yaml
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "AllowFetchingRDSMetrics",
            "Effect": "Allow",
            "Action": [
                "cloudwatch:GetMetricData",
                "ec2:DescribeInstanceTypes",
                "rds:DescribeAccountAttributes",
                "rds:DescribeDBInstances",
                "rds:DescribeDBLogFiles",
                "rds:DescribePendingMaintenanceActions",
                "servicequotas:GetServiceQuota"
            ],
            "Resource": "*"
        }
    ]
}
```

## Installation

### Locally

1. Connect on AWS with any method

    ```bash
    aws configure
    ```

2. Start application

    ```bash
    prometheus-rds-exporter
    ```

### Docker

1. Connect on AWS with any method

2. Start application

    ```bash
    docker run -p 9043:9043 -e AWS_PROFILE=${AWS_PROFILE} -v $HOME/.aws:/app/.aws public.ecr.aws/qonto/prometheus-rds-exporter:latest
    ```

### EKS (using IRSA and Helm)

1. Create an IAM policy

    ```bash
    IAM_POLICY_NAME=prometheus-rds-exporter
    aws iam create-policy --policy-name ${IAM_POLICY_NAME} --policy-document file://configs/aws/policy.json
    ```

2. Create IAM role and EKS service account

    ```bash
    IAM_ROLE_NAME=prometheus-rds-exporter
    EKS_CLUSTER_NAME=default # Replace with your EKS cluster name
    KUBERNETES_NAMESPACE=default # Replace with the namespace that you want to use
    KUBERNETES_SERVICE_ACCOUNT_NAME=prometheus-rds-exporter
    AWS_ACCOUNT_ID=$(aws sts get-caller-identity --query "Account" --output text) # Replace with your AWS ACCOUNT ID
    eksctl create iamserviceaccount --cluster ${EKS_CLUSTER_NAME} --namespace ${KUBERNETES_NAMESPACE} --name ${KUBERNETES_SERVICE_ACCOUNT_NAME} --role-name ${IAM_ROLE_NAME} --attach-policy-arn arn:aws:iam::${AWS_ACCOUNT_ID}:policy/${IAM_POLICY_NAME} --approve
    ```

3. Deploy chart with service account annotation

    ```bash
    helm install prometheus-rds-exporter oci://public.ecr.aws/qonto/prometheus-rds-exporter-chart --namespace ${KUBERNETES_NAMESPACE} --set serviceAccount.annotations."eks\.amazonaws\.com\/role-arn"="arn:aws:iam::${AWS_ACCOUNT_ID}:role/${IAM_ROLE_NAME}"
    ```

### Terraform

You can take example on Terraform code in `configs/terraform/`.

## Alternative

[percona/rds_exporter](https://github.com/percona/rds_exporter) and [mtanda/rds_enhanced_monitoring_exporter](https://github.com/mtanda/rds_enhanced_monitoring_exporter) provides are great alternatives.

[prometheus/cloudwatch_exporter](https://github.com/prometheus/cloudwatch_exporter) could be used to collect additional CloudWatch metrics.

## Contribute

See CONTRIBUTING.md

## Development

### Running the tests

Execute golang tests:

```bash
make test
```

Execute Helm chart tests:

```bash
make helm-test # Helm unit test
make kubeconform # Kubernetes manifest validation
```

### Development environment

You can start a simple development environment using the Docker compose configuration in `/scripts/prometheus`.

It will start and configure Grafana, Prometheus, and the RDS exporter:

1. Connect on AWS using the AWS CLI

1. Launch development stack

    ```bash
    cd scripts/prometheus
    docker compose up --build
    ```

1. Connect on the services

    - Grafana: http://localhost:3000 (credential: admin/hackme)
    - Prometheus: http://localhost:9090
    - Prometheus RDS exporter: http://localhost:9043
