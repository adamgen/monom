package go_utils

import (
	"slices"
	"strings"
)

func FindCommands(commands []string, prefix string) ([]string, error) {
	var result []string
	for _, command := range commands {
		command, err := RemovePrefix(command, prefix)
		if err != nil {
			return nil, err
		}
		if command != "" {
			parts := strings.SplitN(command, "/", 2)
			if !slices.Contains(result, parts[0]) {
				result = append(result, parts[0])
			}
		}
	}

	return result, nil
}

