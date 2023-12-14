package exporter_test

import (
	"fmt"
	"testing"

	"github.com/qonto/prometheus-rds-exporter/internal/app/exporter"
)

func TestClearPrometheusLabel(t *testing.T) {
	testCases := []struct {
		label string
		want  string
	}{
		{"tag_services.k8s.aws/controller-version", "tag_services_k8s_aws_controller_version"},
		{"Unreplaced_LabeL:1", "Unreplaced_LabeL:1"},
		{"1stInvalidCharacter", "_stInvalidCharacter"},
		{"_IsValidAValidLabel", "_IsValidAValidLabel"},
		{":IsValidAValidLabel", ":IsValidAValidLabel"},
	}

	for _, tc := range testCases {
		testName := fmt.Sprintf("Label %s", tc.label)

		t.Run(testName, func(t *testing.T) {
			got := exporter.ClearPrometheusLabel(tc.label)
			if got != tc.want {
				t.Errorf("got %q; want %q", got, tc.want)
			}
		})
	}
}
