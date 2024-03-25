package exporter

import (
	"fmt"
	"testing"
)

func TestClearPrometheusLabel(t *testing.T) {
	testCases := []struct {
		label string
		want  string
	}{
		{"aws:cloudformation:logical_id", "aws_cloudformation_logical_id"},
		{"services.k8s.aws/controller-version", "services_k8s_aws_controller_version"},
		{"1InvalidFirstCharacter", "InvalidFirstCharacter"},
		{":InvalidFirstCharacter", "InvalidFirstCharacter"},
		{"_InvalidFirstCharacter", "InvalidFirstCharacter"},
	}

	for _, tc := range testCases {
		testName := fmt.Sprintf("Label %s", tc.label)

		t.Run(testName, func(t *testing.T) {
			got := clearPrometheusLabel(tc.label)
			if got != tc.want {
				t.Errorf("got %q; want %q", got, tc.want)
			}
		})
	}
}
