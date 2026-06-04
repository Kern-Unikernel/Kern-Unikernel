package common

import "strings"

var SupportedRuntimes = []string{"node", "bun", "go", "python", "rust"}

// Returns all the supported runtimes separated by spaces
func GetStringSupportedRuntimes() string {
	return strings.Join(SupportedRuntimes, " ")
}
