# Grafana dashboard update

This document explains how to update Grafana dashboards.

Requirements:

- Member of Qonto's organization on Grafana.com (to publish dashboard)

Steps:

1. Reset the local development environment to use the **latest dashboard versions**

    > [!IMPORTANT]
    > This step is required to avoid dashboard regression

    ```bash
    cd scripts/prometheus/
    docker compose rm grafana --stop --force # Stop container
    rm -rf .grafana_data # Remove local storage
    ```

1. Start local development environment

    ```bash
    docker compose up
    ```

1. Update the dashboards

    First time with Grafonnet? Look at [Grafonnet documentation](https://grafana.github.io/grafonnet/index.html).

    Edit Grafonnet resources (eg. Panel, dashoards) located in `configs/grafana/`.

    During development, you can build and publish dashboards on local Grafana with following commands:

    ```bash
    cd configs/grafana/
    ./build.sh && ./publish.sh
    ```

    And connect to Grafana on <http://localhost:3000>.

    Tips: Grafana automatically reload dashboards on change.

1. Commit changes within a Merge Request

1. Once merge request is merged, upload dashboard on Grafana.com

    1. Go to the Grafana.com dashboard public page (see [list](https://github.com/qonto/prometheus-rds-exporter#dashboards)), select revisions and click on `Upload revision`

    1. Select `export.json` and click on `Upload new revision`

    1. Check new version is displayed in the revisions
