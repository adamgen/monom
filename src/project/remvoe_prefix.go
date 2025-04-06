package project

import "strings"

func RemovePrefix(commands string, prefix string) (string, error) {
	if prefix == "" {
		return commands, nil
	}

	// If prefix doesn't contain "/", return original command
	if !strings.Contains(prefix, "/") && prefix != commands {
		return commands, nil
	}

	// If the command starts with the prefix
	if strings.HasPrefix(commands, prefix) {
		// If prefix ends with "/" or is the exact command, return everything after the last "/"
		if strings.HasSuffix(prefix, "/") || prefix == commands {
			parts := strings.Split(commands, "/")
			return parts[len(parts)-1], nil
		}
		// If prefix is followed by "/", return everything after the prefix
		if len(commands) > len(prefix) && commands[len(prefix)] == '/' {
			return commands[len(prefix)+1:], nil
		}
	}

	// If prefix contains "/" and is a prefix of a path component, return the last component
	if strings.Contains(prefix, "/") && strings.HasPrefix(commands, prefix[:strings.LastIndex(prefix, "/")+1]) {
		parts := strings.Split(commands, "/")
		return parts[len(parts)-1], nil
	}

	return commands, nil
}
