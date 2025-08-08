package ec2

import "strings"

func chunkBy[T any](items []T, chunkSize int) (chunks [][]T) {
	for chunkSize < len(items) {
		items, chunks = items[chunkSize:], append(chunks, items[0:chunkSize:chunkSize])
	}

	return append(chunks, items)
}

// addDBPrefix adds "db." prefix for RDS instance type
func addDBPrefix(instance string) string {
	return "db." + instance
}

// removeDBPrefix removes "db." prefix for RDS instance type
func removeDBPrefix(instance string) string {
	return strings.Trim(instance, "db.")
}

// overrideInvalidInstanceTypes normalizes EC2 instance type names to handle
// inconsistencies between RDS and EC2 services.
// x2g RDS instances which are memory-optimized instance classes with AWS Graviton2 processors
// are referenced as x2gd in EC2 API
// See: https://github.com/qonto/prometheus-rds-exporter/issues/258
func overrideInvalidInstanceTypes(instanceType string) string {
	if strings.HasPrefix(instanceType, "x2g.") {
		return strings.Replace(instanceType, "x2g.", "x2gd.", 1)
	}
	return instanceType
}
