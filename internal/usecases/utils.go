package usecases

import (
	"bufio"
	"fmt"
	"regexp"
	"strings"
)

func extractTagAndText(message string) (string, string, error) {
	re := regexp.MustCompile(`(?m)^#\s*(\w+)\s*\n([\s\S]*)`)

	match := re.FindStringSubmatch(message)

	if len(match) > 2 {
		return match[1], match[2], nil
	}

	return "", "", fmt.Errorf("no tag found in the message")
}

func isSingleLine(input string) bool {
	re := regexp.MustCompile(`^[^\n\r]*$`)
	return re.MatchString(input)
}

func transformPlaceholders(input string) string {
	// Regular expression to match {{key}} placeholders
	re := regexp.MustCompile(`{{\s*([a-zA-Z0-9_]+)\s*}}`)
	return re.ReplaceAllStringFunc(input, func(match string) string {
		// Extract the placeholder key
		key := strings.TrimSpace(match[2 : len(match)-2])
		// Convert to {{.Key}} format
		return fmt.Sprintf("{{.%s}}", strings.Title(key))
	})
}

func extractItems(text string) ([]string, error) {
	var result []string

	// Create a new scanner to read the text line by line
	scanner := bufio.NewScanner(strings.NewReader(text))
	for scanner.Scan() {
		line := scanner.Text()

		// Trim leading and trailing whitespace
		trimmedLine := strings.TrimSpace(line)

		// Skip empty lines
		if len(trimmedLine) == 0 {
			continue
		}

		// Apply rules to process the line
		if strings.HasPrefix(trimmedLine, "- ") {
			if len(trimmedLine) == 2 {
				continue
			}

			result = append(result, trimmedLine)
		} else if strings.HasPrefix(trimmedLine, "-") {
			if len(trimmedLine) == 1 {
				continue
			}

			result = append(result, "- "+strings.TrimSpace(trimmedLine[1:]))
		} else {
			result = append(result, "- "+trimmedLine)
		}
	}

	// Check for errors during scanning
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan items: %w", err)
	}

	return result, nil
}
