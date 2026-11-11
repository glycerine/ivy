package startingpoint

// unwrapForHeaders removes the single outer C header delimiter pair after
// bodies have been braced. Requiring the matching ')' to precede the body avoids
// stripping a parenthesized subexpression from an already-Go header.
func unwrapForHeaders(source string) (string, error) {
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
		if t.Kind != "word" || t.Text != "for" || i+1 >= len(ts) || ts[i+1].Text != "(" {
			continue
		}
		end, ok := pairs[i+1]
		if !ok || end+1 >= len(ts) || ts[end+1].Text != "{" {
			continue
		}
		space := ""
		if t.End == ts[i+1].Start {
			space = " "
		}
		edits = append(edits, Edit{ts[i+1].Start, ts[i+1].End, space}, Edit{ts[end].Start, ts[end].End, ""})
	}
	return ApplyEdits(source, edits)
}
