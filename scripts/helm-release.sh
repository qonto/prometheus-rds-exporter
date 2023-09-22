#!/bin/bash

# Build and release helm chart if the version does not already exists in the specified AWS ECR public repository

CHART_NAME=$1
CHART_DIRECTORY=$2
RELEASE_VERSION=$3
REPOSITORY=$4

usage() {
    echo "Usage: $0 <chart_name> <chart_directory> <release_version> <AWS_ECR_public_repository>"
    exit 1
}

check_parameters() {
    if [ -z $CHART_NAME ];
    then
        echo "ERROR: Chart name must be specified"
        usage
    fi

    if [ -z $CHART_DIRECTORY ];
    then
        echo "ERROR: Chart directory must be specified"
        usage
    fi

    if [ -z $RELEASE_VERSION ];
    then
        echo "ERROR: Release version must be specified"
        usage
    fi

    if [ -z $REPOSITORY ];
    then
        echo "ERROR: Repository must be specified"
        usage
    fi
}

check_version_exists() {
    AWS_ERROR=$(aws ecr-public describe-images --region us-east-1 --repository-name ${CHART_NAME} --image-ids imageTag=${RELEASE_VERSION} --output json 2>&1 > /dev/null)
    AWS_EXIT_CODE=$?
    if [ $AWS_EXIT_CODE -eq 0 ];
    then
        echo "Release ${RELEASE_VERSION} already exists in AWS ECR"
        exit 0
    elif [ ! $AWS_EXIT_CODE -eq 254 ];
    then
        echo "Unexpected error while checking if ${RELEASE_VERSION} version exists: exit code ${AWS_EXIT_CODE}"
        echo ${AWS_ERROR}
        exit 1
    fi
}

build() {
    helm package ${CHART_DIRECTORY} --app-version ${RELEASE_VERSION} --version ${RELEASE_VERSION}
}

publish() {
    helm push ${CHART_NAME}-${RELEASE_VERSION}.tgz oci://public.ecr.aws/${REPOSITORY}
}

check_parameters
check_version_exists

set -x

build
publish
