# Grafana dashboard update

This document explains how to update Grafana dashboards on Grafana.com.

Requirements:

- Member of Qonto's organization on Grafana.com

Steps:

1. Reset the local development environment to use the **latest dashboard versions**

    > [!IMPORTANT]
    > This step is required to avoid dashboard regression

    ```bash
    cd scripts/prometheus/
    docker compose rm grafana --stop --force # Stop container
    rm -rf scripts/prometheus/.grafana_data # Remove local storage
    ```

1. Start local development environment

1. Update the dashboard with your changes

1. Export the dashboard

    Click on `Share dashboard > Export`, then select `Export for sharing externally` and click on `Save to file`.

1. Remove inputs and fix the source name

    We must edit the exported dashboard because we're using a dynamic data source (usually Grafana dashboards have a hard-coded data source).

    ```bash
    FILENAME=RDS-instances-1701236044179.json # Replace with your exported dashboard

    cat ${FILENAME} | jq 'del( .__inputs )' | sed -e "s/DS_PROMETHEUS/datasource/" > export.json
    ```

1. Upload dashboard on Grafana.com

    1. Go to the Grafana.com dashboard public page (see [list](https://github.com/qonto/prometheus-rds-exporter#dashboards)), select revisions and click on `Upload revision`

    1. Select `export.json` and click on `Upload new revision`

    1. Check new version is displayed in the revisions
