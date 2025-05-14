package utils

import "strings"

func IsQuestionSelection(text string) bool {
	if strings.Contains(text, ".") {
		parts := strings.SplitN(text, ".", 2)
		return len(parts) > 0 && len(parts[0]) > 0
	}
	return false
}
