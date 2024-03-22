package exporter

import (
	"regexp"

	"github.com/qonto/prometheus-rds-exporter/internal/app/rds"
	"golang.org/x/exp/slices"
)

func getUniqTypeAndIdentifiers(instances map[string]rds.RdsInstanceMetrics) ([]string, []string) {
	var (
		instanceTypes       []string
		instanceIdentifiers []string
	)

	for dbinstanceName := range instances {
		instanceClass := instances[dbinstanceName].DBInstanceClass
		if !slices.Contains(instanceTypes, instanceClass) {
			instanceTypes = append(instanceTypes, instanceClass)
		}

		if !slices.Contains(instanceIdentifiers, dbinstanceName) {
			instanceIdentifiers = append(instanceIdentifiers, dbinstanceName)
		}
	}

	// Remove incompatible instance types
	instanceTypes = RemoveElementsByValue(instanceTypes, []string{"db.serverless"})

	return instanceIdentifiers, instanceTypes
}

func ClearPrometheusLabel(str string) string {
	// Prometheus metric names may contain ASCII letters, digits, underscores, and colons.
	// https://prometheus.io/docs/concepts/data_model/#metric-names-and-labels
	InvalidFirstLetterCharacters := regexp.MustCompile(`[^a-zA-Z]+`)
	InvalidCharacters := regexp.MustCompile(`[^a-zA-Z0-9_]+`)

	return InvalidFirstLetterCharacters.ReplaceAllString(string(str[0]), "") + InvalidCharacters.ReplaceAllString(str[1:], "_")
}
