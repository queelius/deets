package store

import (
	"fmt"
	"os"
	"strings"
)

// ValidateName checks that a name is a valid TOML bare key (a-z, A-Z, 0-9, -, _).
func ValidateName(name string) error {
	if name == "" {
		return fmt.Errorf("name must not be empty")
	}
	for _, r := range name {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') || r == '-' || r == '_') {
			return fmt.Errorf("invalid character %q in name %q (allowed: a-z, A-Z, 0-9, -, _)", r, name)
		}
	}
	return nil
}

// ValidateCategoryName checks that a category name is valid.
// Category names may contain dots (for nested TOML tables like "profiles.github"),
// but each dot-separated segment must be a valid bare key.
func ValidateCategoryName(name string) error {
	if name == "" {
		return fmt.Errorf("name must not be empty")
	}
	parts := strings.Split(name, ".")
	for _, part := range parts {
		if err := ValidateName(part); err != nil {
			return err
		}
	}
	return nil
}

// HasCategory reports whether the TOML file at filePath contains a [category] section.
func HasCategory(filePath, category string) bool {
	lines, err := readLines(filePath)
	if err != nil {
		return false
	}
	return findSection(lines, category) != -1
}

// SetValue sets a value for the given key within the specified category in the
// TOML file at filePath. If the file does not exist it is created. If the
// category or key does not exist it is appended. Existing lines, comments, and
// formatting are preserved.
func SetValue(filePath, category, key, value string) error {
	if err := ValidateCategoryName(category); err != nil {
		return fmt.Errorf("invalid category: %w", err)
	}
	if err := ValidateName(key); err != nil {
		return fmt.Errorf("invalid key: %w", err)
	}

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
	if err := ValidateCategoryName(category); err != nil {
		return fmt.Errorf("invalid category: %w", err)
	}
	if err := ValidateName(key); err != nil {
		return fmt.Errorf("invalid key: %w", err)
	}

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
	if err := ValidateCategoryName(category); err != nil {
		return fmt.Errorf("invalid category: %w", err)
	}

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
		if !strings.HasPrefix(trimmed, key) {
			continue
		}
		// After stripping whitespace, the remainder must start with '='.
		// This prevents "name" from matching "name_desc".
		rest := strings.TrimLeft(trimmed[len(key):], " \t")
		if len(rest) > 0 && rest[0] == '=' {
			return i
		}
	}
	return -1
}

// formatValue formats a value for TOML output. If the value is a TOML array
// literal it is written as-is. If it starts with a double quote, it is assumed
// to be already quoted. Otherwise, the value is wrapped in double quotes.
func formatValue(value string) string {
	if isArrayLiteral(value) {
		return value
	}
	if strings.HasPrefix(value, "\"") {
		return value
	}
	return fmt.Sprintf("%q", value)
}

// isArrayLiteral reports whether value looks like a TOML array literal
// (e.g. `[]`, `["a", "b"]`, `[1, 2]`). Plain strings that happen to start
// with '[' (like "[hello]") are NOT array literals.
func isArrayLiteral(value string) bool {
	if len(value) < 2 || value[0] != '[' || value[len(value)-1] != ']' {
		return false
	}
	if value == "[]" {
		return true
	}
	// Skip any whitespace after '[', then check the first content character.
	// Array elements start with '"' (strings), a digit or '-' (numbers),
	// or 't'/'f' (booleans). Bare words like "[hello]" don't match.
	for i := 1; i < len(value); i++ {
		switch value[i] {
		case ' ', '\t':
			continue
		case '"', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '-', 't', 'f':
			return true
		default:
			return false
		}
	}
	return false
}
