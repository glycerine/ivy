// Copyright (c) Microsoft Corporation. All Rights Reserved.
// Ported to Go from ivy_to_md.py.

// Package mdgen converts Ivy specification files (.ivy) to Markdown
// documentation. Lines beginning with '#' (but not '#-') are treated as
// prose comments; all other non-blank lines are wrapped in fenced code
// blocks. The first line of the input is always skipped (it is
// typically the Ivy header).
package mdgen

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// ConvertLines converts Ivy source lines to Markdown, writing the
// result to w.  The first line (index 0) is skipped per the Python
// convention.  Returns the number of bytes written and any error.
func ConvertLines(w io.Writer, lines []string) (int, error) {
	total := 0
	lastWasComment := true

	// Skip the first line (lines[0]), process lines[1:]
	for i := 1; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			n, err := fmt.Fprint(w, "\n")
			total += n
			if err != nil {
				return total, err
			}
			continue
		}

		if strings.HasPrefix(trimmed, "#") && !strings.HasPrefix(trimmed, "#-") {
			// Comment line -> prose
			if !lastWasComment {
				n, err := fmt.Fprint(w, "```\n")
				total += n
				if err != nil {
					return total, err
				}
			}
			// Extract comment text after the first '#'
			idx := strings.Index(line, "#")
			comment := line[idx+1:]
			// Split on '#' and rejoin everything after the first
			parts := strings.SplitN(line, "#", 2)
			if len(parts) > 1 {
				comment = parts[1]
			}
			if strings.HasPrefix(comment, " ") {
				comment = comment[1:]
			}
			n, err := fmt.Fprint(w, comment)
			total += n
			if err != nil {
				return total, err
			}
			// Add newline if comment doesn't end with one
			if !strings.HasSuffix(comment, "\n") {
				n2, err2 := fmt.Fprint(w, "\n")
				total += n2
				if err2 != nil {
					return total, err2
				}
			}
			lastWasComment = true
		} else {
			// Code line
			if lastWasComment {
				n, err := fmt.Fprint(w, "```\n")
				total += n
				if err != nil {
					return total, err
				}
			}
			n, err := fmt.Fprint(w, line)
			total += n
			if err != nil {
				return total, err
			}
			if !strings.HasSuffix(line, "\n") {
				n2, err2 := fmt.Fprint(w, "\n")
				total += n2
				if err2 != nil {
					return total, err2
				}
			}
			lastWasComment = false
		}
	}

	// Close any open code block
	if !lastWasComment {
		n, err := fmt.Fprint(w, "```\n")
		total += n
		if err != nil {
			return total, err
		}
	}

	return total, nil
}

// ConvertString converts an Ivy source string to Markdown.
func ConvertString(input string) string {
	lines := strings.Split(input, "\n")
	// Preserve trailing newlines by keeping split results as-is
	// but add newlines back since Split removes them
	restored := make([]string, len(lines))
	for i, l := range lines {
		if i < len(lines)-1 {
			restored[i] = l + "\n"
		} else {
			restored[i] = l
		}
	}
	var buf strings.Builder
	ConvertLines(&buf, restored)
	return buf.String()
}

// ConvertFile reads an .ivy file and writes the corresponding .md file.
// If outPath is empty, it is derived by replacing the .ivy extension.
func ConvertFile(inPath, outPath string) error {
	if !strings.HasSuffix(inPath, ".ivy") {
		return fmt.Errorf("mdgen: input file must have .ivy extension: %s", inPath)
	}

	data, err := os.ReadFile(inPath)
	if err != nil {
		return fmt.Errorf("mdgen: %w", err)
	}

	if outPath == "" {
		outPath = strings.TrimSuffix(inPath, ".ivy") + ".md"
	}

	lines := strings.Split(string(data), "\n")
	restored := make([]string, len(lines))
	for i, l := range lines {
		if i < len(lines)-1 {
			restored[i] = l + "\n"
		} else {
			restored[i] = l
		}
	}

	f, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("mdgen: %w", err)
	}
	defer f.Close()

	_, err = ConvertLines(f, restored)
	return err
}
