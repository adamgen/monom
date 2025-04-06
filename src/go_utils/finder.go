package go_utils

func FindCommands(commands []string, prefix string) ([]string, error) {
	var result []string
	for _, command := range commands {
		command, err := RemovePrefix(command, prefix)
		if err != nil {
			return nil, err
		}
		if command != "" {
			result = append(result, command)

		}
	}

	return result, nil
}

