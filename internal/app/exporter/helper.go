package exporter

import (
	"regexp"
)

func ClearPrometheusLabel(str string) string {
	// Prometheus metric names may contain ASCII letters, digits, underscores, and colons.
	// https://prometheus.io/docs/concepts/data_model/#metric-names-and-labels
	InvalidFirstLetterCharacters := regexp.MustCompile(`[^a-zA-Z]+`)
	InvalidCharacters := regexp.MustCompile(`[^a-zA-Z0-9_]+`)

	return InvalidFirstLetterCharacters.ReplaceAllString(string(str[0]), "") + InvalidCharacters.ReplaceAllString(str[1:], "_")
}
