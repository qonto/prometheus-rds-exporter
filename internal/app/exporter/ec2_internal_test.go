package exporter

import (
	"slices"
	"testing"

	"github.com/qonto/prometheus-rds-exporter/internal/app/rds"
)

func TestGetUniqTypeAndIdentifiers(t *testing.T) {
	smallInstance := rds.RdsInstanceMetrics{DBInstanceClass: "db.t4.small"}
	largeInstance := rds.RdsInstanceMetrics{DBInstanceClass: "db.t4.large"}
	serverLessInstance := rds.RdsInstanceMetrics{DBInstanceClass: "db.serverless"}

	testCases := []struct {
		testName            string
		instances           map[string]rds.RdsInstanceMetrics
		instanceIdentifiers []string
		instanceTypes       []string
	}{
		{
			"Test instance types aggregation",
			map[string]rds.RdsInstanceMetrics{"db1": smallInstance, "db2": smallInstance, "db3": largeInstance},
			[]string{"db1", "db2", "db3"},
			[]string{largeInstance.DBInstanceClass, smallInstance.DBInstanceClass},
		},
		{
			"Test serverless instances are not returned",
			map[string]rds.RdsInstanceMetrics{"db1": smallInstance, "db2": serverLessInstance},
			[]string{"db1", "db2"},
			[]string{smallInstance.DBInstanceClass},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			instanceIdentifiers, instanceTypes := getUniqTypeAndIdentifiers(tc.instances)
			if !slices.Equal(instanceIdentifiers, tc.instanceIdentifiers) {
				t.Errorf("instance identifiers mismatch. got %q; want %q", instanceIdentifiers, tc.instanceIdentifiers)
			}

			if !slices.Equal(instanceTypes, tc.instanceTypes) {
				t.Errorf("instance types mismatch. got %q; want %q", instanceTypes, tc.instanceTypes)
			}
		})
	}
}
