package exporter

import "golang.org/x/exp/slices"

func removeElementsByValue(originalList, stringsToRemove []string) []string {
	resultList := slices.DeleteFunc(originalList, func(n string) bool {
		return slices.Contains(stringsToRemove, n)
	})

	return resultList
}
