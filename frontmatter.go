// ABOUTME: Markdown frontmatter parsing and rendering utilities.
// ABOUTME: Splits/joins YAML frontmatter (between --- delimiters) and markdown body text.
package mdstore

import (
	"strings"

	"gopkg.in/yaml.v3"
)

// ParseFrontmatter splits YAML frontmatter from markdown body.
// Returns the raw YAML string (between --- delimiters) and the body text.
// If no frontmatter found, returns empty yaml and full content as body.
func ParseFrontmatter(content string) (yamlStr string, body string) {
	trimmed := strings.TrimSpace(content)

	if !strings.HasPrefix(trimmed, "---") {
		return "", content
	}

	// Find the closing ---
	rest := trimmed[3:]
	// Skip the newline after opening ---
	if len(rest) > 0 && rest[0] == '\n' {
		rest = rest[1:]
	} else if len(rest) > 1 && rest[0] == '\r' && rest[1] == '\n' {
		rest = rest[2:]
	}

	closingIdx := strings.Index(rest, "\n---")
	if closingIdx < 0 {
		// No closing delimiter found
		return "", content
	}

	yamlStr = rest[:closingIdx]

	// Body is everything after the closing ---
	afterClose := rest[closingIdx+4:] // len("\n---") == 4
	// Trim a single leading newline from the body
	if len(afterClose) > 0 && afterClose[0] == '\n' {
		afterClose = afterClose[1:]
	} else if len(afterClose) > 1 && afterClose[0] == '\r' && afterClose[1] == '\n' {
		afterClose = afterClose[2:]
	}

	return yamlStr, afterClose
}

// RenderFrontmatter renders YAML frontmatter + body into a complete markdown string.
// metadata is marshaled to YAML between --- delimiters.
func RenderFrontmatter(metadata interface{}, body string) (string, error) {
	yamlBytes, err := yaml.Marshal(metadata)
	if err != nil {
		return "", err
	}

	var b strings.Builder
	b.WriteString("---\n")
	b.Write(yamlBytes)
	b.WriteString("---\n")
	if body != "" {
		b.WriteString(body)
	}

	return b.String(), nil
}
