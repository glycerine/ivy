package startingpoint

// statementLabelEnds identifies named statement labels, not ternary colons,
// bitfields or C++ scope operators. Keys are byte offsets just after the colon.
func statementLabelEnds(all []Token) map[int]bool {
	var ts []Token
	for _, t := range all {
		if t.Kind != "space" && t.Kind != "comment" {
			ts = append(ts, t)
		}
	}
	ends := map[int]bool{}
	pairs, _ := MatchDelimiters(ts)
	opens := map[int]int{}
	for open, close := range pairs {
		if open < close {
			opens[close] = open
		}
	}
	for i := 0; i+1 < len(ts); i++ {
		if ts[i].Kind != "word" || ts[i+1].Text != ":" {
			continue
		}
		boundary := i == 0
		if i > 0 {
			prev := ts[i-1]
			boundary = prev.Text == "{" || prev.Text == "}" || prev.Text == ";" || prev.Text == "else" || prev.Text == "do" || ends[prev.End]
			if prev.Text == ")" {
				if open, ok := opens[i-1]; ok && open > 0 {
					switch ts[open-1].Text {
					case "if", "for", "while", "switch":
						boundary = true
					}
				}
			}
		}
		if boundary {
			ends[ts[i+1].End] = true
		}
	}
	return ends
}
