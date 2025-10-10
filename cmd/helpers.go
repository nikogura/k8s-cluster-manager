package cmd

import "strings"

// stripDomainSuffix removes domain suffix from node names for comparison.
// E.g., "charlie-cp-1.terrace.fi" -> "charlie-cp-1".
func stripDomainSuffix(name string) (shortName string) {
	parts := strings.Split(name, ".")
	shortName = parts[0]
	return shortName
}
