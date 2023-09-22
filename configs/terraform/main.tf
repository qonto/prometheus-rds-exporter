terraform {
  // Replace with your Terraform configuration (eg. S3 backend)
}

provider "aws" {
  // Replace with your AWS provider configuration
}

resource "aws_iam_role" "default" {
  name               = "prometheus-rds-exporter"
  assume_role_policy = data.aws_iam_policy_document.prometheus-rds-exporter-relationship.json
}

data "aws_iam_policy_document" "prometheus-rds-exporter-relationship" {
  // Replace with asumme role policy
}

resource "aws_iam_role_policy" "prometheus-rds-exporter" {
  name   = "prometheus-rds-exporter"
  role   = aws_iam_role.default.name
  policy = data.aws_iam_policy_document.prometheus-rds-exporter.json
}

data "aws_iam_policy_document" "prometheus-rds-exporter" {
  statement {
    sid    = "AllowFetchingRDSMetrics"
    effect = "Allow"
    actions = [
      "cloudwatch:GetMetricData",
      "ec2:DescribeInstanceTypes",
      "rds:DescribeAccountAttributes",
      "rds:DescribeDBInstances",
      "rds:DescribeDBLogFiles",
      "rds:DescribePendingMaintenanceActions",
      "servicequotas:GetServiceQuota",
    ]
    resources = ["*"]
  }
}
