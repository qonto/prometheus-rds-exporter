package exporter

import (
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

	return instanceIdentifiers, instanceTypes
}
