package startingpoint

import "strconv"

// lowerPostfixIndexAssignments handles dst = src[i++] (and i--) when both
// sides are simple access paths. Calls, lazy expressions, reused index names,
// and control-header assignments need more evaluation-order analysis.
func lowerPostfixIndexAssignments(source string) (string, error) {
	ts, pairs, err := significantTokens(source)
	if err != nil {
		return "", err
	}
	labels := statementLabelEnds(ts)
	// Never turn one control-header clause into multiple statements.
	headers := map[int]bool{}
	for i, t := range ts {
		if t.Text != "for" && t.Text != "if" && t.Text != "switch" && t.Text != "while" {
			continue
		}
		for j := i + 1; j < len(ts) && ts[j].Text != "{"; j++ {
			headers[j] = true
			if end, ok := pairs[j]; ok && end > j {
				for k := j + 1; k <= end; k++ {
					headers[k] = true
				}
				j = end
			}
		}
	}
	var edits []Edit
	for i, t := range ts {
		if t.Kind != "word" || headers[i] {
			continue
		}
		if i > 0 && ts[i-1].Text != "{" && ts[i-1].Text != "}" && ts[i-1].Text != ";" && !labels[ts[i-1].End] {
			continue
		}
		lhs := postfixAccessEnd(ts, pairs, i)
		if lhs >= len(ts) || ts[lhs].Text != "=" {
			continue
		}
		rhs := lhs + 1
		if rhs >= len(ts) || ts[rhs].Kind != "word" {
			continue
		}
		end := postfixAccessEnd(ts, pairs, rhs)
		// The final indexed operand has exactly identifier ++/-- as its index.
		if end+5 >= len(ts) || ts[end].Text != "[" || ts[end+1].Kind != "word" {
			continue
		}
		a, b := ts[end+2], ts[end+3]
		if (a.Text != "+" && a.Text != "-") || b.Text != a.Text || a.End != b.Start || ts[end+4].Text != "]" || ts[end+5].Text != ";" {
			continue
		}
		name := ts[end+1].Text
		reused := false
		for j := i; j < end; j++ {
			if ts[j].Kind == "word" && ts[j].Text == name {
				reused = true
			}
		}
		if reused {
			continue
		}
		edits = append(edits, Edit{a.Start, b.End, ""}, Edit{ts[end+5].End, ts[end+5].End, " " + name + a.Text + b.Text + ";"})
	}
	return ApplyEdits(source, edits)
}

// Consume a simple identifier/member/index path, stopping before a postfix
// index. Restrict ordinary indices to identifiers or integer literals.
func postfixAccessEnd(ts []Token, pairs map[int]int, i int) int {
	if i >= len(ts) || ts[i].Kind != "word" {
		return i
	}
	i++
	for i < len(ts) {
		if (ts[i].Text == "." || ts[i].Text == "->") && i+1 < len(ts) && ts[i+1].Kind == "word" {
			i += 2
			continue
		}
		if ts[i].Text != "[" {
			break
		}
		end, ok := pairs[i]
		if !ok || end <= i+1 {
			break
		}
		if end == i+2 && ts[i+1].Kind == "word" {
			i = end + 1
			continue
		}
		literal := ""
		for j := i + 1; j < end; j++ {
			literal += ts[j].Text
		}
		if _, err := strconv.ParseUint(literal, 0, 64); err == nil {
			i = end + 1
			continue
		}
		break
	}
	return i
}
