package project

import "strings"

func RemovePrefix(commands string, prefix string) (string, error) {
	if prefix == "" {
		return commands, nil
	}

	// If prefix doesn't contain "/", return original command
	if !strings.Contains(prefix, "/") {
		return commands, nil
	}

	// If the prefix ends with "/", use it as is
	// Otherwise, find the last "/" in the prefix and use everything up to and including it
	prefixToUse := prefix
	if !strings.HasSuffix(prefix, "/") {
		lastSlash := strings.LastIndex(prefix[:strings.LastIndex(prefix, "/")+1], "/")
		if lastSlash != -1 {
			prefixToUse = prefix[:lastSlash+1]
		}
	}

	// If the command starts with the prefix we want to use, remove it
	if strings.HasPrefix(commands, prefixToUse) {
		return strings.TrimPrefix(commands, prefixToUse), nil
	}

	return commands, nil
}
