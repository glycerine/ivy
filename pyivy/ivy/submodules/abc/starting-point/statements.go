package startingpoint

import "strings"

// braceStatements recognizes statement boundaries using tokens, so nested
// if/else, for bodies, comments and semicolons in expressions preserve their scope.
func braceStatements(source string) (string, error) {
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
	pairs, err := MatchDelimiters(ts)
	if err != nil {
		return "", err
	}
	labelEnds := statementLabelEnds(ts)
	var edits []Edit
	var statement func(int) (int, bool)
	statement = func(i int) (int, bool) {
		if i >= len(ts) {
			return i, false
		}
		// A label prefixes a statement; it must not hide that statement's
		// control flow from the walker. Recursion also handles stacked labels.
		if i+1 < len(ts) && labelEnds[ts[i+1].End] {
			return statement(i + 2)
		}
		switch ts[i].Text {
		case "{":
			end, ok := pairs[i]
			if !ok {
				return i, false
			}
			for j := i + 1; j < end; {
				checkpoint := len(edits)
				next, valid := statement(j)
				if !valid || next <= j {
					edits = edits[:checkpoint]
					j++
				} else {
					j = next
				}
			}
			return end + 1, true
		case "if":
			if i+1 >= len(ts) || ts[i+1].Text != "(" {
				return i, false
			}
			close, ok := pairs[i+1]
			if !ok {
				return i, false
			}
			begin := close + 1
			end, ok := statement(begin)
			if !ok {
				return i, false
			}
			hasElse := end < len(ts) && ts[end].Text == "else"
			if ts[begin].Text != "{" {
				closing, text := ts[end-1].End, " }"
				// Keep else on the same Go line as its closing brace, without
				// deleting comments between the source body and else.
				if hasElse && strings.ContainsAny(source[closing:ts[end].Start], "\r\n") {
					closing, text = ts[end].Start, "} "
				}
				edits = append(edits, Edit{ts[close].End, ts[close].End, " {"}, Edit{closing, closing, text})
			} else if hasElse && strings.ContainsAny(source[ts[end-1].End:ts[end].Start], "\r\n") {
				edits = append(edits, Edit{ts[end-1].Start, ts[end-1].End, ""}, Edit{ts[end].Start, ts[end].Start, "} "})
			}
			if end < len(ts) && ts[end].Text == "else" {
				branch := end + 1
				after, ok := statement(branch)
				if !ok {
					return i, false
				}
				// Retain idiomatic else-if chains; their if bodies are handled recursively.
				if ts[branch].Text != "{" && ts[branch].Text != "if" {
					edits = append(edits, Edit{ts[end].End, ts[end].End, " {"}, Edit{ts[after-1].End, ts[after-1].End, " }"})
				}
				return after, true
			}
			return end, true
		case "for", "while", "switch":
			if i+1 >= len(ts) || ts[i+1].Text != "(" {
				return i, false
			}
			close, ok := pairs[i+1]
			if !ok {
				return i, false
			}
			begin := close + 1
			end, valid := statement(begin)
			if !valid {
				return i, false
			}
			if ts[i].Text == "while" {
				edits = append(edits, Edit{ts[i].Start, ts[i].End, "for"})
			}
			if (ts[i].Text == "for" || ts[i].Text == "while") && ts[begin].Text != "{" {
				closing := " }"
				if strings.Contains(source[ts[close].End:ts[end-1].End], "\n") {
					lineStart := strings.LastIndex(source[:ts[i].Start], "\n") + 1
					j := lineStart
					for j < ts[i].Start && (source[j] == ' ' || source[j] == '\t') {
						j++
					}
					closing = "\n" + source[lineStart:j] + "}"
				}
				edits = append(edits, Edit{ts[close].End, ts[close].End, " {"}, Edit{ts[end-1].End, ts[end-1].End, closing})
			}
			return end, true
		case "do":
			end, ok := statement(i + 1)
			if !ok {
				return i, false
			}
			if end+1 >= len(ts) || ts[end].Text != "while" || ts[end+1].Text != "(" {
				return i, false
			}
			close, ok := pairs[end+1]
			if !ok || close+1 >= len(ts) || ts[close+1].Text != ";" {
				return i, false
			}
			return close + 2, true
		}
		// Ordinary expression/declaration/return statement. Skip matched groups;
		// a directive or an unmatched body boundary is not a statement to guess at.
		for j := i; j < len(ts); j++ {
			if ts[j].Kind == "directive" || ts[j].Text == "}" || ts[j].Text == "{" {
				return i, false
			}
			if ts[j].Text == ";" {
				return j + 1, true
			}
			if end, ok := pairs[j]; ok {
				j = end
			}
		}
		return i, false
	}
	for i := 0; i < len(ts); {
		// Roll back edits if recognizing a complete statement fails.
		checkpoint := len(edits)
		next, ok := statement(i)
		if !ok || next <= i {
			edits = edits[:checkpoint]
			i++
		} else {
			i = next
		}
	}
	converted, err := ApplyEdits(source, edits)
	if err != nil {
		return "", err
	}
	return joinControlBraces(converted)
}

// joinControlBraces moves existing braces ahead of intervening comments so
// line comments cannot swallow the opening brace. Literal contents are untouched.
func joinControlBraces(source string) (string, error) {
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
	pairs, err := MatchDelimiters(ts)
	if err != nil {
		return "", err
	}
	var edits []Edit
	for i, t := range ts {
		if t.Text != "if" && t.Text != "for" && t.Text != "while" && t.Text != "switch" && t.Text != "else" {
			continue
		}
		end := i
		if t.Text != "else" {
			if i+1 >= len(ts) || ts[i+1].Text != "(" {
				continue
			}
			var ok bool
			end, ok = pairs[i+1]
			if !ok {
				continue
			}
		}
		if end+1 >= len(ts) || ts[end+1].Text != "{" {
			continue
		}
		gap := source[ts[end].End:ts[end+1].Start]
		if !strings.ContainsAny(gap, "\r\n") || strings.Contains(gap, "SP_OPAQUE_") {
			continue
		}
		if strings.TrimSpace(gap) == "" {
			edits = append(edits, Edit{ts[end].End, ts[end+1].End, " {"})
		} else {
			edits = append(edits, Edit{ts[end].End, ts[end].End, " {"}, Edit{ts[end+1].Start, ts[end+1].End, ""})
		}
	}
	return ApplyEdits(source, edits)
}
