package startingpoint

import "strings"

// rewriteMainBody handles the conventional zero- or two-parameter C entry point.
func rewriteMainBody(body string, params []string) (string, bool, error) {
	argc, argv := "", ""
	if len(params) == 2 {
		argc = strings.Fields(params[0])[0]
		argv = strings.Fields(params[1])[0]
	}
	all, err := Lex(body)
	if err != nil {
		return "", false, err
	}
	var ts []Token
	for _, t := range all {
		if t.Kind != "space" && t.Kind != "comment" {
			ts = append(ts, t)
		}
	}
	pairs, err := MatchDelimiters(ts)
	if err != nil {
		return "", false, err
	}
	var edits []Edit
	usesOS := false
	for i := 0; i < len(ts); i++ {
		t := ts[i]
		// Macro functions inside a body have their own parameters and returns.
		if t.Text == "func" {
			for j := i + 1; j < len(ts); j++ {
				if ts[j].Text == "{" {
					i = pairs[j]
					break
				}
			}
			continue
		}
		if t.Kind != "word" {
			continue
		}
		member := i > 0 && (ts[i-1].Text == "." || ts[i-1].Text == "->" || ts[i-1].Text == "::")
		if !member && ((argc != "" && t.Text == argc) || (argv != "" && t.Text == argv)) {
			replacement := "os.Args"
			if t.Text == argc {
				replacement = "len(os.Args)"
			}
			edits = append(edits, Edit{t.Start, t.End, replacement})
			usesOS = true
		}
		if t.Text != "return" {
			continue
		}
		end := i + 1
		for ; end < len(ts) && ts[end].Text != ";"; end++ {
			if close, ok := pairs[end]; ok && close > end {
				end = close
			}
		}
		if end >= len(ts) || end == i+1 {
			continue
		}
		if end == i+2 && ts[i+1].Text == "0" {
			edits = append(edits, Edit{ts[i+1].Start, ts[i+1].End, ""})
		} else {
			edits = append(edits, Edit{t.Start, t.End, "os.Exit(int("}, Edit{ts[end].Start, ts[end].Start, "))"})
			usesOS = true
		}
	}
	result, err := ApplyEdits(body, edits)
	return result, usesOS, err
}

func ensureOSImport(source string) (string, error) {
	ts, err := Lex(source)
	if err != nil {
		return "", err
	}
	var significant []Token
	for _, t := range ts {
		if t.Kind != "space" && t.Kind != "comment" {
			significant = append(significant, t)
		}
	}
	pairs, err := MatchDelimiters(significant)
	if err != nil {
		return "", err
	}
	for i, t := range significant {
		if t.Kind != "word" || t.Text != "import" || i+1 >= len(significant) {
			continue
		}
		end := i + 2
		if significant[i+1].Text == "(" {
			end = pairs[i+1]
		} else if significant[i+1].Kind != "string" {
			end = i + 3
		}
		if end > len(significant) {
			end = len(significant)
		}
		for _, spec := range significant[i+1 : end] {
			if spec.Kind == "string" && (spec.Text == `"os"` || spec.Text == "`os`") {
				return source, nil
			}
		}
	}
	for i, t := range ts {
		if t.Text != "package" || t.Kind != "word" {
			continue
		}
		for _, name := range ts[i+1:] {
			if name.Kind == "word" {
				return ApplyEdits(source, []Edit{{name.End, name.End, "\n\nimport \"os\""}})
			}
		}
	}
	return source, nil
}
