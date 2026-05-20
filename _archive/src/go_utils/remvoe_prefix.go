package go_utils

import (
	"strings"
)

func RemovePrefix(command string, prefix string) (string, error) {
	if prefix == "" {
		return command, nil
	}

	// Special case: if the prefix is just a single character and matches any part of the command
	if !strings.HasPrefix(command, prefix) {
		return "", nil
	}

	// If the prefix ends with "/", use it as is
	// Otherwise, find the last "/" in the prefix and use everything up to and including it
	prefixToUse := prefix
	if !strings.HasSuffix(prefix, "/") {
		lastSlash := strings.LastIndex(prefix, "/")
		if lastSlash != -1 {
			prefixToUse = prefix[:lastSlash+1]
		}
	}

	// If the command starts with the prefix we want to use, remove it
	if strings.HasPrefix(command, prefixToUse) && strings.HasSuffix(prefixToUse, "/") {
		result := strings.TrimPrefix(command, prefixToUse)
		return result, nil
	}

	return command, nil
}
