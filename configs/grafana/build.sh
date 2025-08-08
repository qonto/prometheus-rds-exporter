#!/bin/bash

set -e

install_dependencies() {
    jb install
}

build_dashboard() {
    DASHBOARD_NAME=$1

    INPUT_FILE=dashboards/${DASHBOARD_NAME}.jsonnet
    OUTPUT_FILE=public/${DASHBOARD_NAME}.json

    if ! jsonnet -J vendor ${INPUT_FILE} > ${OUTPUT_FILE}; then
        echo "Error: Failed to build dashboard ${DASHBOARD_NAME}" >&2
        exit 1
    fi

    echo ${OUTPUT_FILE}
}

install_dependencies

# Build all dashboards if no specified dashboard
if [ -z $1 ]; then
    DASHBOARDS=`ls dashboards/`
else
    DASHBOARDS=$1
fi

# Build and publish dashboards
for DASHBOARD in ${DASHBOARDS};
do
    DASHBOARD_NAME=${DASHBOARD%%.*}

    echo "=> Build ${DASHBOARD_NAME}"
    DASHBOARD_FILENAME=$(build_dashboard ${DASHBOARD_NAME})
done
