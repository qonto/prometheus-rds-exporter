package trace_test

import (
	"reflect"
	"testing"

	"github.com/qonto/prometheus-rds-exporter/internal/app/trace"
	"go.opentelemetry.io/otel/attribute"
)

func TestOtelKeys(t *testing.T) {
	type test struct {
		name  string
		want  attribute.KeyValue
		key   string
		value any
	}

	tests := []test{
		{"QuotaServiceCode", trace.AWSQuotaServiceCode("unittest"), "qonto.prometheus_rds_exporter.aws.quota.service_code", "unittest"},
		{"QuotaCode", trace.AWSQuotaCode("unittest"), "qonto.prometheus_rds_exporter.aws.quota.code", "unittest"},
		{"InstanceTypesCount", trace.AWSInstanceTypesCount(42), "qonto.prometheus_rds_exporter.aws.instance-types-count", int64(42)},
	}

	for _, tc := range tests {
		if !reflect.DeepEqual(string(tc.want.Key), tc.key) {
			t.Fatalf("%s: expected key: %v, got: %v", tc.name, tc.want.Key, tc.key)
		}

		switch tc.value.(type) {
		case string:
			if !reflect.DeepEqual(tc.want.Value.AsString(), tc.value) {
				t.Fatalf("%s: expected value: %v, got: %v", tc.name, tc.want.Value.AsString(), tc.value)
			}
		case int64:
			if !reflect.DeepEqual(tc.want.Value.AsInt64(), tc.value) {
				t.Fatalf("%s: expected value: %v, got: %v", tc.name, tc.want.Value.AsInt64(), tc.value)
			}
		default:
			t.Fatalf("%s: %s type is not implemented. Add it to the test suite", tc.name, reflect.TypeOf(tc.value))
		}
	}
}
