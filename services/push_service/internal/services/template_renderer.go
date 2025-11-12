package services

import (
	"fmt"
	"regexp"
)

var placeholderRegex = regexp.MustCompile(`\{\{\s*([a-zA-Z0-9_]+)\s*\}\}`)

// RenderTemplate performs naive moustache-style replacement for {{key}} placeholders.
func RenderTemplate(template string, variables map[string]interface{}) string {
	if template == "" || len(variables) == 0 {
		return template
	}

	return placeholderRegex.ReplaceAllStringFunc(template, func(match string) string {
		submatch := placeholderRegex.FindStringSubmatch(match)
		if len(submatch) != 2 {
			return match
		}
		key := submatch[1]
		if value, ok := variables[key]; ok {
			return fmt.Sprint(value)
		}
		return match
	})
}
