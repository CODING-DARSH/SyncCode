package parser

import (
	"strings"
)

func SplitArguments(input string) []string {
	result := []string{}
	current := ""
	brackets := 0
	inString := false

	for _, ch := range input {
		switch ch {
		case '[':
			brackets++
			current += string(ch)
		case ']':
			brackets--
			current += string(ch)
		case '"':
			inString = !inString
			current += string(ch)
		case ',':
			if brackets == 0 && !inString {
				result = append(result, strings.TrimSpace(current))
				current = ""
			} else {
				current += string(ch)
			}
		default:
			current += string(ch)
		}
	}

	if current != "" {
		result = append(result, strings.TrimSpace(current))
	}

	return result
}
