#!/bin/bash

# This script tests that all Helm chart test values and default chart values return a valid Kubernetes manifest

set -e
set -o pipefail

HELM_CHART_DIRECTORY=$1
KUBERNETES_VERSION=${KUBERNETES_VERSION-1.25.0}
KUBECONFORM_CACHE_DIRECTORY=${KUBECONFORM_CACHE_DIRECTORY-/tmp}

HELM_DEFAULT_VALUE_FILE=${HELM_CHART_DIRECTORY}/values.yaml
HELM_TEST_VALUES_DIRECTORY=${HELM_CHART_DIRECTORY}/tests/values

check_parameters() {
	if [[ -z $HELM_CHART_DIRECTORY ]]; then
		echo "ERRROR: You must specify the helm chart directory"
		usage
	fi
	if [[ ! -f $HELM_DEFAULT_VALUE_FILE ]]; then
		echo "ERRROR: Default Helm values ${HELM_DEFAULT_VALUE_FILE} does not exists"
		usage
	fi
	if [[ ! -d $HELM_TEST_VALUES_DIRECTORY ]]; then
		echo "ERRROR: Helm test values directory ${HELM_TEST_VALUES_DIRECTORY} does not exists"
		usage
	fi
}

usage() {
	echo ""
	echo "Usage: $0 <helm_chart_directory>"
	exit 1
}

check_parameters

HELM_VALUE_FILES=$(find ${HELM_DEFAULT_VALUE_FILE} ${HELM_TEST_VALUES_DIRECTORY}/*.yaml)

for FILE in $HELM_VALUE_FILES;
do
    printf "\033[32mTest chart with ${FILE}\033[0m\n"
	echo ""

	# Redirect Helm error output to /dev/null to remove symlink warning, see https://github.com/helm/helm/issues/7019
	helm template configs/helm \
		-f $FILE \
		2> /dev/null \
	| kubeconform \
		--strict \
		-exit-on-error \
		-kubernetes-version ${KUBERNETES_VERSION} \
		-cache ${KUBECONFORM_CACHE_DIRECTORY} \
		-schema-location default \
		-schema-location 'scripts/kubeconform/{{.Group}}/{{ .ResourceKind }}_{{.ResourceAPIVersion}}.json' \
		-schema-location 'https://raw.githubusercontent.com/datreeio/CRDs-catalog/main/{{.Group}}/{{.ResourceKind}}_{{.ResourceAPIVersion}}.json' \
		-summary

	echo ""
done
