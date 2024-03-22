#!/bin/bash

set -e

upload_dashboard() {
    DASHBOARD_FILENAME=$1

    GRAFANA_URL=${GRAFANA_URL-http://localhost:3000}
    GRAFANA_USERNAME=${GRAFANA_USERNAME-admin}
    GRAFANA_PASSWORD=${GRAFANA_PASSWORD-hackme}

    GRAFANA_API_DASHBOARD_UPDATE_PAYLOAD="grafana-api.update-dashboard.json"
    GRAFANA_API_REQUEST_FILE=/tmp/dashboard.json
    GRAFANA_OUTPUT_FILE=/tmp/upload_output.json

    # Build API request
    jq --argjson dashboard "$(<${DASHBOARD_FILENAME})" '.dashboard += $dashboard' ${GRAFANA_API_DASHBOARD_UPDATE_PAYLOAD} > ${GRAFANA_API_REQUEST_FILE}

    # Update dashboard on Grafana
    curl \
        ${GRAFANA_URL}/api/dashboards/db \
        --user "${GRAFANA_USERNAME}:${GRAFANA_PASSWORD}" \
        --header 'Content-Type: application/json' \
        --silent \
        --fail \
        -d @${GRAFANA_API_REQUEST_FILE} \
        > ${GRAFANA_OUTPUT_FILE}

    DASHBOARD_URL=`jq -r '.url' ${GRAFANA_OUTPUT_FILE}`
    DASHBOARD_SLUG=`jq -r '.slug' ${GRAFANA_OUTPUT_FILE}`
    DASHBOARD_VERSION=`jq -r '.version' ${GRAFANA_OUTPUT_FILE}`

    echo "Version '${DASHBOARD_VERSION}' of '${DASHBOARD_SLUG}' dashboard uploaded on ${GRAFANA_URL}${DASHBOARD_URL}"
}

# Publish all dashboards if no specified dashboard
if [ -z $1 ]; then
    DASHBOARDS=`ls public/`
else
    DASHBOARDS=$1.json
fi

# Build and publish dashboards
for DASHBOARD_FILENAME in ${DASHBOARDS};
do
    echo "=> Publish ${DASHBOARD_FILENAME%%.*}"
    upload_dashboard public/${DASHBOARD_FILENAME}
done
