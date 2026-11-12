package startingpoint

import "strings"

func splitExpressions(source string) ([]string, bool) {
	ts, pairs, err := significantTokens(source)
	if err != nil {
		return nil, false
	}
	start := 0
	var parts []string
	for i := 0; i < len(ts); i++ {
		if close, ok := pairs[i]; ok && close > i {
			i = close
			continue
		}
		if ts[i].Text == "," {
			parts = append(parts, strings.TrimSpace(source[start:ts[i].Start]))
			start = ts[i].End
		}
	}
	parts = append(parts, strings.TrimSpace(source[start:]))
	return parts, len(parts) > 1
}
func independentAssignments(source string) (string, bool) {
	parts, ok := splitExpressions(source)
	if !ok {
		return source, false
	}
	var left, right []string
	for _, part := range parts {
		if strings.HasSuffix(part, "++") || strings.HasSuffix(part, "--") {
			name := strings.TrimSpace(part[:len(part)-2])
			if !identifier.MatchString(name) {
				return source, false
			}
			left = append(left, name)
			right = append(right, name+string(part[len(part)-1])+"1")
			continue
		}
		split := strings.SplitN(part, "=", 2)
		if len(split) != 2 || !identifier.MatchString(strings.TrimSpace(split[0])) || strings.HasPrefix(strings.TrimSpace(split[1]), "=") {
			return source, false
		}
		left = append(left, strings.TrimSpace(split[0]))
		right = append(right, strings.TrimSpace(split[1]))
	}
	seen := map[string]bool{}
	for _, name := range left {
		if seen[name] {
			return source, false
		}
		seen[name] = true
	}
	for i, expr := range right {
		tokens, err := Lex(expr)
		if err != nil {
			return source, false
		}
		for _, t := range tokens {
			if t.Kind == "word" && seen[t.Text] && t.Text != left[i] {
				return source, false
			}
			// Calls and indexing can hide dependencies and side effects.
			if t.Text == "(" || t.Text == "[" {
				return source, false
			}
		}
	}
	return strings.Join(left, ", ") + " = " + strings.Join(right, ", "), true
}

func lowerCommaExpressions(source string) (string, error) {
	ts, pairs, err := significantTokens(source)
	if err != nil {
		return "", err
	}
	labelEnds := statementLabelEnds(ts)
	var edits []Edit
	for i, t := range ts {
		if t.Text == "for" {
			end := i + 1
			var semis []int
			for end < len(ts) && ts[end].Text != "{" {
				if ts[end].Text == ";" {
					semis = append(semis, end)
				}
				if close, ok := pairs[end]; ok && close > end {
					end = close
				}
				end++
			}
			if len(semis) == 2 && end < len(ts) {
				for _, span := range [][2]int{{i + 1, semis[0]}, {semis[1] + 1, end}} {
					if span[0] >= span[1] {
						continue
					}
					start, finish := ts[span[0]].Start, ts[span[1]-1].End
					if value, ok := independentAssignments(source[start:finish]); ok {
						edits = append(edits, Edit{start, finish, value})
					}
				}
			}
		}
		if i > 0 && t.Kind == "word" && i+1 < len(ts) && (ts[i-1].Text == "{" || ts[i-1].Text == ";" || ts[i-1].Text == "}" || labelEnds[ts[i-1].End]) && (ts[i+1].Text == "=" || ts[i+1].Text == "(") && t.Text != "if" && t.Text != "for" && t.Text != "switch" && t.Text != "while" {
			end := i + 1
			for end < len(ts) && ts[end].Text != ";" && ts[end].Text != "{" && ts[end].Text != "}" {
				if close, ok := pairs[end]; ok && close > end {
					end = close
				}
				end++
			}
			if end < len(ts) && ts[end].Text == ";" {
				if parts, ok := splitExpressions(source[t.Start:ts[end].Start]); ok {
					edits = append(edits, Edit{t.Start, ts[end].Start, strings.Join(parts, ";\n")})
				}
			}
		}
		// Parenthesized, discarded comma expressions at statement boundaries.
		if t.Text != "(" || i == 0 || (ts[i-1].Text != "{" && ts[i-1].Text != ";" && ts[i-1].Text != "}" && !labelEnds[ts[i-1].End]) {
			continue
		}
		end := pairs[i]
		if end+1 >= len(ts) || (ts[end+1].Text != ";" && ts[end+1].Text != "}") {
			continue
		}
		inner := source[t.End:ts[end].Start]
		if parts, ok := splitExpressions(inner); ok {
			edits = append(edits, Edit{t.Start, ts[end].End, strings.Join(parts, ";\n") + ";"})
		}
	}
	return ApplyEdits(source, edits)
}
