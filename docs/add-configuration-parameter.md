# Add a configuration parameter

This document explains how to add a configuration parameter in the exporter.

Steps:

1. Define parameter name

1. Implement it

    1. Add a new field in `exporterConfig` structure in `cmd/root.go`
    1. Add the parameter in flag `cmd.Flags()`

1. Add tests

1. Add parameter in the `README.md`

1. Add parameter in YAML default configuration file in `configs/prometheus-rds-exporter/prometheus-rds-exporter.yaml`
