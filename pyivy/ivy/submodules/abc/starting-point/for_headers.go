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

// simplifyConditionOnlyFor drops empty init/post clauses, retaining trivia.
func simplifyConditionOnlyFor(source string) (string, error) {
	ts, pairs, err := significantTokens(source)
	if err != nil {
		return "", err
	}
	var edits []Edit
	for i, t := range ts {
		if t.Text != "for" {
			continue
		}
		start, end := i+1, i+1
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
		if end >= len(ts) || len(semis) != 2 || semis[0] != start || semis[1] != end-1 {
			continue
		}
		first, last := semis[0], semis[1]
		if first+1 == last {
			// for (;;) is simply for {}, but comments in the empty clauses survive.
			gap := source[t.End:ts[end].Start]
			if !strings.Contains(gap, "/*") && !strings.Contains(gap, "//") {
				edits = append(edits, Edit{t.End, ts[end].Start, " "})
				continue
			}
		}
		leftStart, leftEnd := ts[first].Start, ts[first].End
		rightStart, rightEnd := ts[last].Start, ts[last].End
		if strings.TrimSpace(source[t.End:leftStart]) == "" && first+1 < last && strings.TrimSpace(source[leftEnd:ts[first+1].Start]) == "" {
			leftStart = t.End
			leftEnd = ts[first+1].Start
		}
		if strings.TrimSpace(source[rightEnd:ts[end].Start]) == "" && first+1 < last && strings.TrimSpace(source[ts[last-1].End:rightStart]) == "" {
			rightStart = ts[last-1].End
			rightEnd = ts[end].Start
		}
		leftText, rightText := "", ""
		if leftStart == t.End {
			leftText = " "
		}
		if rightEnd == ts[end].Start {
			rightText = " "
		}
		edits = append(edits, Edit{leftStart, leftEnd, leftText}, Edit{rightStart, rightEnd, rightText})
	}
	return ApplyEdits(source, edits)
}
