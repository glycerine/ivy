package startingpoint

import "strings"

// promoteSprintfDestinations changes the declaration bound to an emitted
// Sprintf assignment. Resolve names lexically so another function or a nested
// shadowing variable with the same name is not changed by association.
func promoteSprintfDestinations(source, qualifier string) (string, error) {
	ts, pairs, err := significantTokens(source)
	if err != nil {
		return "", err
	}
	type binding struct {
		start, end int
		typ        string
	}
	scopes := []map[string]binding{{}}
	pendingParams := map[int][]string{}
	for i, t := range ts {
		if t.Text != "func" || i+2 >= len(ts) || ts[i+2].Text != "(" {
			continue
		}
		close := pairs[i+2]
		open := close + 1
		for open < len(ts) && ts[open].Text != "{" {
			open++
		}
		// Parameter identifiers block outer declarations; including type names is
		// harmless here and avoids guessing parameter types in imperfect drafts.
		for j := i + 3; j < close; j++ {
			if ts[j].Kind == "word" {
				pendingParams[open] = append(pendingParams[open], ts[j].Text)
			}
		}
	}
	chosen := map[int]binding{}
	for i, t := range ts {
		switch t.Text {
		case "{":
			scope := map[string]binding{}
			for _, name := range pendingParams[i] {
				scope[name] = binding{}
			}
			scopes = append(scopes, scope)
		case "}":
			if len(scopes) > 1 {
				scopes = scopes[:len(scopes)-1]
			}
		case "var":
			if i+2 >= len(ts) {
				continue
			}
			name := ts[i+1].Text
			b := binding{}
			start := i + 2
			if ts[start].Text == "[" {
				close, ok := pairs[start]
				if ok && close > start+1 && close+1 < len(ts) && ts[close+1].Text == "byte" {
					end := close + 2
					// Zero initializers have already been removed by aggregate lowering.
					// Only annotate a complete, uninitialized declaration here.
					if end < len(ts) && ts[end].Text == ";" && strings.TrimSpace(source[ts[close+1].End:ts[end].Start]) == "" {
						b = binding{ts[start].Start, ts[end].End, source[ts[start].Start:ts[close+1].End]}
					}
				}
			}
			scopes[len(scopes)-1][name] = b
		}
		// Block a same-name local introduced with := as well.
		if t.Kind == "word" && i+1 < len(ts) && ts[i+1].Text == ":=" {
			scopes[len(scopes)-1][t.Text] = binding{}
		}
		if t.Kind != "word" || i+2 >= len(ts) || ts[i+1].Text != "=" {
			continue
		}
		if i > 0 && (ts[i-1].Text == "." || ts[i-1].Text == "->") {
			continue
		}
		call := i + 2
		if qualifier != "" {
			if call+2 >= len(ts) || ts[call].Text != qualifier || ts[call+1].Text != "." {
				continue
			}
			call += 2
		}
		if call+1 >= len(ts) || ts[call].Text != "Sprintf" || ts[call+1].Text != "(" {
			continue
		}
		for j := len(scopes) - 1; j >= 0; j-- {
			if b, ok := scopes[j][t.Text]; ok {
				if b.typ != "" && !strings.ContainsAny(b.typ, "\r\n") && !strings.Contains(b.typ, "/*") {
					chosen[b.start] = b
				}
				break
			}
		}
	}
	var edits []Edit
	for _, b := range chosen {
		// Always terminate the annotation: later code may share the C source line.
		edits = append(edits, Edit{b.start, b.end, "string // was " + b.typ + " in C\n"})
	}
	return ApplyEdits(source, edits)
}
