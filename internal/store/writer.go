package store

import (
	"fmt"
	"os"
	"strings"
)

// SetValue sets a value for the given key within the specified category in the
// TOML file at filePath. If the file does not exist it is created. If the
// category or key does not exist it is appended. Existing lines, comments, and
// formatting are preserved.
func SetValue(filePath, category, key, value string) error {
	lines, err := readLines(filePath)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		// File does not exist — create with section and key.
		lines = []string{
			fmt.Sprintf("[%s]", category),
			fmt.Sprintf("%s = %s", key, formatValue(value)),
		}
		return writeLines(filePath, lines)
	}

	formatted := formatValue(value)
	sectionIdx := findSection(lines, category)

	if sectionIdx == -1 {
		// Category does not exist — append it.
		if len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) != "" {
			lines = append(lines, "")
		}
		lines = append(lines, fmt.Sprintf("[%s]", category))
		lines = append(lines, fmt.Sprintf("%s = %s", key, formatted))
		return writeLines(filePath, lines)
	}

	// Category exists — look for the key within it.
	nextSection := findNextSection(lines, sectionIdx)
	keyIdx := findKey(lines, sectionIdx+1, nextSection, key)

	if keyIdx != -1 {
		// Key exists — replace the line.
		lines[keyIdx] = fmt.Sprintf("%s = %s", key, formatted)
	} else {
		// Key does not exist — insert before the next section (or EOF).
		insertAt := nextSection
		newLine := fmt.Sprintf("%s = %s", key, formatted)
		lines = append(lines[:insertAt], append([]string{newLine}, lines[insertAt:]...)...)
	}

	return writeLines(filePath, lines)
}

// RemoveValue removes a key from the specified category in the TOML file at
// filePath. If the category becomes empty (no keys left), the section header
// is also removed. Returns an error if the key is not found.
func RemoveValue(filePath, category, key string) error {
	lines, err := readLines(filePath)
	if err != nil {
		return err
	}

	sectionIdx := findSection(lines, category)
	if sectionIdx == -1 {
		return fmt.Errorf("category %q not found in %s", category, filePath)
	}

	nextSection := findNextSection(lines, sectionIdx)
	keyIdx := findKey(lines, sectionIdx+1, nextSection, key)
	if keyIdx == -1 {
		return fmt.Errorf("key %q not found in category %q in %s", key, category, filePath)
	}

	// Remove the key line.
	lines = append(lines[:keyIdx], lines[keyIdx+1:]...)

	// Check if the category is now empty (no non-blank, non-comment, non-section lines).
	nextSection = findNextSection(lines, sectionIdx)
	empty := true
	for i := sectionIdx + 1; i < nextSection; i++ {
		trimmed := strings.TrimSpace(lines[i])
		if trimmed != "" && !strings.HasPrefix(trimmed, "#") {
			empty = false
			break
		}
	}

	if empty {
		// Remove the section header and any blank/comment lines that belong to it.
		lines = append(lines[:sectionIdx], lines[nextSection:]...)
	}

	return writeLines(filePath, lines)
}

// RemoveCategory removes an entire category (header and all lines until the
// next section or EOF) from the TOML file at filePath. Returns an error if
// the category is not found.
func RemoveCategory(filePath, category string) error {
	lines, err := readLines(filePath)
	if err != nil {
		return err
	}

	sectionIdx := findSection(lines, category)
	if sectionIdx == -1 {
		return fmt.Errorf("category %q not found in %s", category, filePath)
	}

	nextSection := findNextSection(lines, sectionIdx)
	lines = append(lines[:sectionIdx], lines[nextSection:]...)

	return writeLines(filePath, lines)
}

// readLines reads the file at path and returns its content split into lines.
func readLines(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	content := string(data)
	if content == "" {
		return []string{}, nil
	}
	// Remove trailing newline to avoid a phantom empty line at the end.
	content = strings.TrimRight(content, "\n")
	return strings.Split(content, "\n"), nil
}

// writeLines writes the given lines back to the file at path with 0644
// permissions. A trailing newline is appended.
func writeLines(path string, lines []string) error {
	content := strings.Join(lines, "\n") + "\n"
	return os.WriteFile(path, []byte(content), 0644)
}

// findSection returns the line index of the [category] header in lines,
// or -1 if the section is not found.
func findSection(lines []string, category string) int {
	target := fmt.Sprintf("[%s]", category)
	for i, line := range lines {
		if strings.TrimSpace(line) == target {
			return i
		}
	}
	return -1
}

// findNextSection returns the line index of the next [section] header after
// afterLine, or len(lines) if no subsequent section is found.
func findNextSection(lines []string, afterLine int) int {
	for i := afterLine + 1; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])
		if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
			return i
		}
	}
	return len(lines)
}

// findKey searches for a line matching "key = " (with optional whitespace)
// between indices start (inclusive) and end (exclusive). Returns the line
// index or -1 if not found.
func findKey(lines []string, start, end int, key string) int {
	for i := start; i < end; i++ {
		trimmed := strings.TrimSpace(lines[i])
		// Match "key = ..." or "key=..."
		if strings.HasPrefix(trimmed, key) {
			rest := trimmed[len(key):]
			rest = strings.TrimLeft(rest, " \t")
			if strings.HasPrefix(rest, "=") {
				return i
			}
		}
	}
	return -1
}

// formatValue formats a value for TOML output. If the value starts with "[",
// it is treated as an array literal and written as-is. If it starts with a
// double quote, it is assumed to be already quoted. Otherwise, the value is
// wrapped in double quotes.
func formatValue(value string) string {
	if strings.HasPrefix(value, "[") {
		return value
	}
	if strings.HasPrefix(value, "\"") {
		return value
	}
	return fmt.Sprintf("%q", value)
}
