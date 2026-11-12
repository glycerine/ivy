package startingpoint

import "strings"

// Comment only using-namespace directives, retaining aliases and using
// declarations for separate translation. Strings and existing comments survive.
func commentUsingNamespaces(source string) (string, error) {
	all, err := Lex(source)
	if err != nil {
		return "", err
	}
	var ts []Token
	for _, t := range all {
		if t.Kind != "space" && t.Kind != "comment" {
			ts = append(ts, t)
		}
	}
	var edits []Edit
	for i := 0; i+2 < len(ts); i++ {
		if ts[i].Text != "using" || ts[i+1].Text != "namespace" {
			continue
		}
		j := i + 2
		if ts[j].Text == "::" {
			j++
		}
		if j >= len(ts) || ts[j].Kind != "word" {
			continue
		}
		j++
		for j+1 < len(ts) && ts[j].Text == "::" && ts[j+1].Kind == "word" {
			j += 2
		}
		if j >= len(ts) || ts[j].Text != ";" {
			continue
		}
		start, end := ts[i].Start, ts[j].End
		text := "//" + strings.ReplaceAll(source[start:end], "\n", "\n//") + "\n"
		line := strings.LastIndex(source[:start], "\n") + 1
		if strings.TrimSpace(source[line:start]) != "" {
			text = "\n" + text
		}
		edits = append(edits, Edit{start, end, text})
		i = j
	}
	return ApplyEdits(source, edits)
}
