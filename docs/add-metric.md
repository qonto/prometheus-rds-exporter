# Add a metric

This document explains how to add a new metric in the exporter:

Requirements:

- Metric must be relevant to create alert or visualization (🚩 Don't collect metrics just for convenient visualization)
- Metric names and labels must follow the [Prometheus best practices](https://prometheus.io/docs/practices/naming/)

Steps:

1. Define metric names and labels

1. Implement it

    1. Add a new field in `rdsCollector` structure in `internal/app/exporter/exporter.go`
    1. Add metrics description in `NewCollector()` function in `internal/app/exporter/exporter.go`
    1. Collect the metrics
    1. Export the result in `Collect()` in `internal/app/exporter/exporter.go`

1. Add tests

1. Test metrics in an AWS environment

1. Add metric in the `README.md`

1. Commit with `feat(metric): Add <metric> <short description>` message to mark it a new feature in the release's changelog.

1. Optional. Update Grafana dashboards to display the metric

1. Optional. Add alert using the new metric in [DMF](https://github.com/qonto/database-monitoring-framework)
