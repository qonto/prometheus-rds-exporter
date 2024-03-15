// Package trace provides OTEL tracing resources
package trace

import "go.opentelemetry.io/otel/attribute"

// Attribute Naming convention
//
// Use namespacing to avoid name clashes. Delimit the namespaces using a dot character. For example service.version denotes the service version where service is the namespace and version is an attribute in that namespace.
// For each multi-word dot-delimited component of the attribute name separate the words by underscores (i.e. use snake_case).
// See https://opentelemetry.io/docs/specs/semconv/general/attribute-naming/

const (
	AWSServiceCodeOtelKey    = attribute.Key("qonto.prometheus_rds_exporter.aws.quota.service_code")
	AWSQuotaCodeOtelKey      = attribute.Key("qonto.prometheus_rds_exporter.aws.quota.code")
	AWSInstanceTypesCountKey = attribute.Key("qonto.prometheus_rds_exporter.aws.instance-types-count")
)

func AWSQuotaServiceCode(val string) attribute.KeyValue {
	return AWSServiceCodeOtelKey.String(val)
}

func AWSQuotaCode(val string) attribute.KeyValue {
	return AWSQuotaCodeOtelKey.String(val)
}

func AWSInstanceTypesCount(val int64) attribute.KeyValue {
	return AWSInstanceTypesCountKey.Int64(val)
}
