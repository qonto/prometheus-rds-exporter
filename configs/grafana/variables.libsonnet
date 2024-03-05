local g = import './g.libsonnet';
local var = g.dashboard.variable;
local RefreshOnTime = 2;

{
  __multi:
    var.custom.selectionOptions.withIncludeAll(true)
    + var.custom.selectionOptions.withMulti(true),

  datasource:
    var.datasource.new('datasource', 'prometheus')
    + var.datasource.generalOptions.withLabel('Data source')
    + var.custom.generalOptions.withDescription('Prometheus data source'),

  aws_account_id:
    var.query.new('aws_account_id')
    + var.datasource.generalOptions.withLabel('Account')
    + var.custom.generalOptions.withDescription('AWS Account ID')
    + var.query.withDatasourceFromVariable(self.datasource)
    + var.query.queryTypes.withLabelValues(
      'aws_account_id',
      'rds_instance_info',
    )
    + var.query.withRefresh(RefreshOnTime),

  aws_account_ids:
    self.aws_account_id
    + self.__multi,

  aws_region:
    var.query.new('aws_region')
    + var.datasource.generalOptions.withLabel('Region')
    + var.custom.generalOptions.withDescription('AWS region')
    + var.query.withDatasourceFromVariable(self.datasource)
    + var.query.queryTypes.withLabelValues(
      'aws_region',
      'rds_instance_info{aws_account_id="$aws_account_id"}',
    )
    + var.query.withRefresh(RefreshOnTime),

  aws_regions:
    self.aws_region
    + self.__multi
    + var.query.queryTypes.withLabelValues(
      'aws_region',
      'rds_instance_info{aws_account_id=~"$aws_account_id"}',
    ),

  dbidentifier:
    var.query.new('dbidentifier')
    + var.datasource.generalOptions.withLabel('Instance')
    + var.custom.generalOptions.withDescription('AWS RDS instance')
    + var.query.withDatasourceFromVariable(self.datasource)
    + var.query.queryTypes.withLabelValues(
      'dbidentifier',
      'rds_instance_info{aws_account_id=~"$aws_account_id", aws_region=~"$aws_region"}',
    )
    + var.query.withRefresh(RefreshOnTime),

  exporter:
    var.query.new('instance')
    + var.datasource.generalOptions.withLabel('Instance')
    + var.custom.generalOptions.withDescription('Prometheus RDS exporter')
    + var.custom.selectionOptions.withIncludeAll(true)
    + var.query.withDatasourceFromVariable(self.datasource)
    + var.query.queryTypes.withLabelValues(
      'instance',
      'rds_exporter_build_info{}',
    )
    + var.query.withRefresh(RefreshOnTime),
}
