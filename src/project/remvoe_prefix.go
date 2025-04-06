package project

import "strings"

func RemovePrefix(commands string, prefix string) (string, error) {
	if prefix == "" {
		return commands, nil
	}

	// Special case: if the prefix is just a single character and matches any part of the command
	if len(prefix) == 1 && strings.Contains(commands, prefix) {
		return "", nil
	}

	// If prefix doesn't contain "/", return original command
	if !strings.Contains(prefix, "/") {
		return commands, nil
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
	if strings.HasPrefix(commands, prefixToUse) {
		result := strings.TrimPrefix(commands, prefixToUse)
		return result, nil
	}

	return commands, nil
}
