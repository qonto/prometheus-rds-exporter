#!/bin/bash

set -xe

GO_RUNTIME_DASHBOARD_ID=14061

DASHBOARD_IDS="${GO_RUNTIME_DASHBOARD_ID}"
DASHBOARD_PROVISIONING_FOLDER=/var/tmp/public-grafana-dashboards

download_dashboards() {
    if [ ! -d $DASHBOARD_PROVISIONING_FOLDER ]; then
        mkdir ${DASHBOARD_PROVISIONING_FOLDER}
    fi

    for DASHBOARD_ID in $DASHBOARD_IDS;
    do
        curl --silent \
            --fail-with-body \
            --connect-timeout 60 \
            --max-time 60 \
            --header "Accept: application/json" \
            --header "Content-Type: application/json;charset=UTF-8" \
            https://grafana.com/api/dashboards/${DASHBOARD_ID}/revisions/latest/download | sed -e 's/DS_PROMETHEUS/datasource/' > ${DASHBOARD_PROVISIONING_FOLDER}/${DASHBOARD_ID}.json
    done
}

download_dashboards

/run.sh # Launch Grafana
