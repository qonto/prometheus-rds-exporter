#!/bin/bash

set -e

install_dependencies() {
    jb install
}

build_dashboard() {
    DASHBOARD_NAME=$1

    INPUT_FILE=dashboards/${DASHBOARD_NAME}.jsonnet
    OUTPUT_FILE=public/${DASHBOARD_NAME}.json

    jsonnet -J vendor ${INPUT_FILE} > ${OUTPUT_FILE}

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
