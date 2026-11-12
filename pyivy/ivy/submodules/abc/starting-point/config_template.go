package startingpoint

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var ppInteger = regexp.MustCompile(`\b(?:0[xX][0-9a-fA-F]+|[0-9]+)[uUlL]*\b`)

// WriteConfigTemplate writes an editable, inactive template without overwriting
// an existing template. Directives are comments; every discovered symbol is 0.
func WriteConfigTemplate(path string, sources []string) (bool, error) {
	symbols := map[string]bool{}
	guardNames := map[string]bool{}
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
		guards := map[int]bool{}
		if strings.EqualFold(filepath.Ext(source), ".h") {
			for name, offset := range configHeaderGuards(ts, source) {
				guardNames[name] = true
				guards[offset] = true
			}
		}
		for _, t := range ts {
			if guards[t.Start] {
				continue
			}
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
	for name := range guardNames {
		delete(symbols, name)
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
	output.WriteString("\n// Observed conditional directives (header guards omitted):\n")
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

// Match a near-top #ifndef NAME / #define NAME pair in a header. Comments and
// whitespace do not interrupt the pair. Preamble directives are allowed; stop
// at ordinary code, and do not mistake nested platform choices for file guards.
func configHeaderGuards(all []Token, source string) map[string]int {
	var ts []Token
	for _, t := range all {
		if t.Kind != "space" && t.Kind != "comment" {
			ts = append(ts, t)
		}
	}
	guards := map[string]int{}
	argument := func(t Token) string {
		parts := directiveParts.FindStringSubmatch(t.Text)
		if len(parts) < 3 {
			return ""
		}
		tokens, err := Lex(parts[2])
		if err != nil {
			return ""
		}
		return strings.TrimSpace(Mask(tokens))
	}
	guardCondition := func(t Token) string {
		name := argument(t)
		if directiveName(t) == "ifndef" && identifier.MatchString(name) {
			return name
		}
		if directiveName(t) == "if" {
			if match := negativeDefinedGuard.FindStringSubmatch(name); match != nil {
				if match[1] != "" {
					return match[1]
				}
				return match[2]
			}
		}
		return ""
	}
	depth := 0
	for i, t := range ts {
		if i >= 32 || t.Kind != "directive" {
			break
		}
		kind := directiveName(t)
		// Some ABC headers define their internal marker before testing it, or
		// define it under a differently named outer guard. Require header identity
		// and an empty definition; never treat a valued feature macro as a guard.
		if kind == "define" && depth <= 1 {
			name := argument(t)
			if identifier.MatchString(name) && headerSpecificGuard(name, source) {
				guards[name] = -1
			}
		}
		if depth == 0 && i+1 < len(ts) && directiveName(ts[i+1]) == "define" {
			name, defined := guardCondition(t), argument(ts[i+1])
			if name != "" && (defined == name || (headerSpecificGuard(name, source) && headerSpecificGuard(defined, source))) {
				guards[name] = t.Start
				if defined != name {
					guards[defined] = t.Start
				}
			}
		}
		switch kind {
		case "if", "ifdef", "ifndef":
			depth++
		case "endif":
			depth--
		}
	}
	// A reversed guard test may occur after declarations or namespace markers.
	for _, t := range ts {
		name := guardCondition(t)
		if _, ok := guards[name]; ok {
			guards[name] = t.Start
		}
	}
	return guards
}

var negativeDefinedGuard = regexp.MustCompile(`^!\s*defined\s*(?:\(\s*([A-Za-z_]\w*)\s*\)|([A-Za-z_]\w*))$`)

func headerSpecificGuard(name, source string) bool {
	stem := strings.ReplaceAll(filepath.Base(source), ".", "_")
	if strings.EqualFold(strings.Trim(name, "_"), stem) {
		return true
	}
	return strings.HasPrefix(name, "ABC__") && strings.HasSuffix(strings.ToLower(name), "__"+strings.ToLower(stem))
}
