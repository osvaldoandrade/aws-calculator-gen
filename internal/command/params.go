package command

import (
	"fmt"
	"strings"
)

// ParseParams converts a slice of key=value strings into a map.
// It returns an error if any item does not follow the expected format.
func ParseParams(items []string) (map[string]string, error) {
	params := make(map[string]string)
	for _, item := range items {
		if item == "" {
			continue
		}
		parts := strings.SplitN(item, "=", 2)
		if len(parts) != 2 || parts[0] == "" {
			return nil, fmt.Errorf("invalid parameter: %s", item)
		}
		params[parts[0]] = parts[1]
	}
	return params, nil
}
