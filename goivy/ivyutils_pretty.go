package goivy

import "strings"

// Pretty formats a string by splitting on semicolons and braces,
// then indenting based on brace nesting. Truncates to maxLines if > 0.
// Corresponds to Python's pretty(s, max_lines=None) in ivy_utils.py lines 680-694.
func Pretty(s string, maxLines int) string {
	s = strings.ReplaceAll(s, ";", ";\n")
	s = strings.ReplaceAll(s, "{", "{\n")
	s = strings.ReplaceAll(s, "}", "\n}")
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimSpace(line)
	}
	if maxLines > 0 && len(lines) > maxLines {
		lines = lines[:maxLines-1]
		lines = append(lines, "...")
	}
	indent := 0
	var res []string
	for _, line := range lines {
		if strings.Contains(line, "}") {
			indent--
		}
		if indent < 0 {
			indent = 0
		}
		res = append(res, strings.Repeat("    ", indent)+line)
		if strings.Contains(line, "{") {
			indent++
		}
	}
	return strings.Join(res, "\n") + strings.Repeat("}", indent)
}
