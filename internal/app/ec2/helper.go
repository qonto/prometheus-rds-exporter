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
