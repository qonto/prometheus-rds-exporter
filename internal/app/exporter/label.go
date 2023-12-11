package exporter

import "regexp"

type Label string

func (l Label) Sanitize() string {
	// Prometheus metric names may contain ASCII letters, digits, underscores, and colons.
	// It must match the regex [a-zA-Z_:][a-zA-Z0-9_:]*.
	// https://prometheus.io/docs/concepts/data_model/#metric-names-and-labels
	invalidFirstLetterCharacters := regexp.MustCompile(`[^a-zA-Z_:]+`)
	invalidCharacters := regexp.MustCompile(`[^a-zA-Z0-9_:]+`)

	return invalidFirstLetterCharacters.ReplaceAllString(string(l[0]), "_") + invalidCharacters.ReplaceAllString(string(l[1:]), "_")
}
