package logseq

import (
	"regexp"
	"strings"
)

var propertyLineRe = regexp.MustCompile(`^[a-zA-Z_-]+:: .*$`)

// StripProperties removes Logseq property lines (key:: value) from block content.
// Returns cleaned content and extracted property lines for later restoration.
func StripProperties(content string) (string, []string) {
	lines := strings.Split(content, "\n")
	var cleaned []string
	var props []string
	for _, line := range lines {
		if propertyLineRe.MatchString(strings.TrimSpace(line)) {
			props = append(props, strings.TrimSpace(line))
		} else {
			cleaned = append(cleaned, line)
		}
	}
	return strings.TrimSpace(strings.Join(cleaned, "\n")), props
}

// RestoreProperties re-attaches property lines to content.
func RestoreProperties(content string, propLines []string) string {
	if len(propLines) == 0 {
		return content
	}
	return content + "\n" + strings.Join(propLines, "\n")
}
