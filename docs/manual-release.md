# Manual release

This document explains how to manually release Prometheus RDS version.

/!\ It's here only for troubleshooting since release process is fully automated using Github workflow.

## Authenticate on AWS

1. Grant `qonto-team-devops` in ECR policy (managed by Terraform in `opensource` stack)

1. Use `opensource-production` AWS PROFILE

    ```bash
    AWS_PROFILE=opensource-production
    ```

1. Login to AWS ECR

    ```bash
    aws ecr-public get-login-password --region us-east-1 | helm registry login --username AWS --password-stdin public.ecr.aws
    ```

## Github release and ECR images

1. Create git tag

    ```bash
    git tag <version>
    ```

1. Release new version

    ```bash
    goreleaser release --clean
    ```

## Helm chart

1. Build and publish the new release

    ```bash
    make helm-publish
    ```

1. Perform dry-run install

    ```bash
    helm install prometheus-rds-exporter oci://public.ecr.aws/qonto/test1-chart --dry-run
    ```
