package startingpoint

// emptyForBodies fills null statements and, for macro fragments, a trailing
// bodyless for-header. Existing nonempty loop bodies remain untouched.
func emptyForBodies(source string, macroFragment bool) (string, error) {
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
		close, ok := pairs[i+1]
		if !ok {
			continue
		}
		next := close + 1
		if next < len(ts) && ts[next].Text == ";" {
			text := "{}"
			if ts[next].Start == ts[close].End {
				text = " {}"
			}
			edits = append(edits, Edit{ts[next].Start, ts[next].End, text})
		} else if macroFragment && next == len(ts) {
			edits = append(edits, Edit{ts[close].End, ts[close].End, " {}"})
		}
	}
	return ApplyEdits(source, edits)
}
