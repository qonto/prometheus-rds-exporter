terraform {
  // TODO Replace with your Terraform configuration (eg. S3 backend)
}

provider "aws" {
  // TODO  Replace with your AWS provider configuration
}

resource "aws_iam_role" "default" {
  name               = "prometheus-rds-exporter"
  assume_role_policy = data.aws_iam_policy_document.prometheus-rds-exporter-relationship.json
}

data "aws_iam_policy_document" "prometheus-rds-exporter-relationship" {
  // TODO Replace with asumme role policy
}

resource "aws_iam_role_policy" "prometheus-rds-exporter" {
  name   = "prometheus-rds-exporter"
  role   = aws_iam_role.default.name
  policy = data.aws_iam_policy_document.prometheus-rds-exporter.json
}

data "aws_iam_policy_document" "prometheus-rds-exporter" {
  statement {
    sid    = "AllowInstanceAndLogDescriptions"
    effect = "Allow"
    actions = [
      "rds:DescribeDBInstances",
      "rds:DescribeDBLogFiles",
    ]
    resources = [
      "arn:aws:rds:*:*:db:*",
    ]
  }

  statement {
    sid    = "AllowMaintenanceDescriptions"
    effect = "Allow"
    actions = [
      "rds:DescribePendingMaintenanceActions",
    ]
    resources = ["*"]
  }

  statement {
    sid    = "AllowGettingCloudWatchMetrics"
    effect = "Allow"
    actions = [
      "cloudwatch:GetMetricData",
    ]
    resources = ["*"]
  }

  statement {
    sid    = "AllowRDSUsageDescriptions"
    effect = "Allow"
    actions = [
      "rds:DescribeAccountAttributes",
    ]
    resources = ["*"]
  }

  statement {
    sid    = "AllowQuotaDescriptions"
    effect = "Allow"
    actions = [
      "servicequotas:GetServiceQuota",
    ]
    resources = ["*"]
  }

  statement {
    sid    = "AllowInstanceTypesDescriptions"
    effect = "Allow"
    actions = [
      "ec2:DescribeInstanceTypes",
    ]
    resources = ["*"]
  }
}
