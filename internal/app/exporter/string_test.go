package exporter_test

import (
	"fmt"
	"slices"
	"testing"

	"github.com/qonto/prometheus-rds-exporter/internal/app/exporter"
)

func TestRemoveElementsByValue(t *testing.T) {
	testCases := []struct {
		input     []string
		removable []string
		want      []string
	}{
		{[]string{"a", "b", "c"}, []string{"c"}, []string{"a", "b"}},
		{[]string{"a", "b", "c"}, []string{"b"}, []string{"a", "c"}},
		{[]string{"a", "b", "c"}, []string{"a"}, []string{"b", "c"}},
	}

	for _, tc := range testCases {
		testName := fmt.Sprintf("Label %s", tc.input)

		t.Run(testName, func(t *testing.T) {
			got := exporter.RemoveElementsByValue(tc.input, tc.removable)
			if !slices.Equal(got, tc.want) {
				t.Errorf("got %q; want %q", got, tc.want)
			}
		})
	}
}
