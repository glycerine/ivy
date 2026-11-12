package startingpoint

import "strings"

// unwrapControlHeaders removes the single outer C header delimiter pair after
// bodies have been braced. Requiring the matching ')' to precede the body avoids
// stripping a parenthesized subexpression from an already-Go header.
func unwrapControlHeaders(source string) (string, error) {
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
		if t.Kind != "word" || (t.Text != "for" && t.Text != "if" && t.Text != "switch") || i+1 >= len(ts) || ts[i+1].Text != "(" {
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
		leftEnd, rightStart := ts[i+1].End, ts[end].Start
		if t.Text == "if" && end > i+2 {
			if strings.TrimSpace(source[leftEnd:ts[i+2].Start]) == "" {
				leftEnd = ts[i+2].Start
			}
			if strings.TrimSpace(source[ts[end-1].End:rightStart]) == "" {
				rightStart = ts[end-1].End
			}
		}
		edits = append(edits, Edit{ts[i+1].Start, leftEnd, space}, Edit{rightStart, ts[end].End, ""})
	}
	return ApplyEdits(source, edits)
}
