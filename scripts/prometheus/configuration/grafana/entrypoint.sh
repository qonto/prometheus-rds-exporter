#!/bin/bash

set -xe

DASHBOARD_IDS="19646 19647 19679 14061"
DASHBOARD_PROVISIONNING=/var/tmp/dashboards

download_dashboards() {
    if [ ! -d $DASHBOARD_PROVISIONNING ]; then
        mkdir ${DASHBOARD_PROVISIONNING}
    fi

    for DASHBOARD_ID in $DASHBOARD_IDS;
    do
        curl --silent \
            --fail-with-body \
            --connect-timeout 60 \
            --max-time 60 \
            --header "Accept: application/json" \
            --header "Content-Type: application/json;charset=UTF-8" \
            https://grafana.com/api/dashboards/${DASHBOARD_ID}/revisions/latest/download | sed -e 's/DS_PROMETHEUS/datasource/' > ${DASHBOARD_PROVISIONNING}/${DASHBOARD_ID}.json
    done
}

download_dashboards

/run.sh # Launch Grafana
