package startingpoint

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
)

var ppInteger = regexp.MustCompile(`\b(?:0[xX][0-9a-fA-F]+|[0-9]+)[uUlL]*\b`)

// WriteConfigTemplate writes an editable, inactive template without overwriting
// an existing template. Directives are comments; every discovered symbol is 0.
func WriteConfigTemplate(path string, sources []string) (bool, error) {
	symbols := map[string]bool{}
	var choices []string
	for _, source := range sources {
		data, err := os.ReadFile(source)
		if err != nil {
			return false, err
		}
		ts, err := Lex(string(data))
		if err != nil {
			continue
		} // fallback files may not lex
		for _, t := range ts {
			name := directiveName(t)
			if name != "if" && name != "ifdef" && name != "ifndef" && name != "elif" && name != "elseif" {
				continue
			}
			line := strings.Count(string(data[:t.Start]), "\n") + 1
			choices = append(choices, fmt.Sprintf("// %s:%d: %s\n", source, line, strings.Join(strings.Fields(t.Text), " ")))
			expr := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(strings.TrimPrefix(t.Text, "#")), name))
			expr = ppInteger.ReplaceAllString(expr, " ")
			tokens, err := Lex(expr)
			if err != nil {
				continue
			}
			for _, token := range tokens {
				if token.Kind == "word" && token.Text != "defined" {
					symbols[token.Text] = true
				}
			}
		}
	}
	var names []string
	for name := range symbols {
		names = append(names, name)
	}
	sort.Strings(names)
	var output strings.Builder
	output.WriteString("// Suggested platform choices, not active until renamed to starting.cfg.\n// All values are placeholders: review before enabling. 0 means disabled.\n// Unknown symbols remain unresolved when omitted from the active config.\n\n")
	for _, name := range names {
		fmt.Fprintf(&output, "%s = 0\n", name)
	}
	output.WriteString("\n// Observed conditional directives (including header guards):\n")
	for _, choice := range choices {
		output.WriteString(choice)
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if os.IsExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	_, err = file.WriteString(output.String())
	closeErr := file.Close()
	if err != nil {
		return false, err
	}
	if closeErr != nil {
		return false, closeErr
	}
	return true, nil
}
