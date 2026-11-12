package startingpoint

// combineAllocationDeclarations moves an uninitialized local pointer's binding
// to its first use, if that use is a same-type allocation in the same straight-
// line block. Do not change package variables, control flow or static types.
func combineAllocationDeclarations(source string) (string, error) {
	ts, pairs, err := significantTokens(source)
	if err != nil {
		return "", err
	}
	var edits []Edit
	for f, t := range ts {
		if t.Text != "func" || f+2 >= len(ts) || ts[f+1].Kind != "word" || ts[f+2].Text != "(" {
			continue
		}
		argsEnd, ok := pairs[f+2]
		if !ok {
			continue
		}
		body := argsEnd + 1
		for body < len(ts) && ts[body].Text != "{" {
			body++
		}
		end, ok := pairs[body]
		if !ok {
			continue
		}
		for i := body + 1; i+5 < end; i++ {
			if ts[i].Text != "var" || ts[i+1].Kind != "word" || ts[i+2].Text != "*" || ts[i+3].Kind != "word" || ts[i+4].Text != ";" {
				continue
			}
			// The declaration itself must be a standalone statement.
			if ts[i-1].Text != "{" && ts[i-1].Text != "}" && ts[i-1].Text != ";" {
				continue
			}
			name, typ := ts[i+1].Text, ts[i+3].Text
			for j := i + 5; j < end; j++ {
				tok := ts[j]
				// Empty composite literals in intervening assignments do not
				// introduce a lexical block or refer to the candidate variable.
				if tok.Text == "{" && pairs[j] == j+1 && j > 0 && ts[j-1].Kind == "word" {
					j++
					continue
				}
				// Crossing blocks, labels, unknown directives or control headers would
				// require a control-flow analysis before shortening the variable's scope.
				if tok.Kind == "directive" || tok.Text == "{" || tok.Text == "}" || tok.Text == ":" || tok.Text == "if" || tok.Text == "for" || tok.Text == "switch" || tok.Text == "goto" || tok.Text == "return" || tok.Text == "case" || tok.Text == "default" {
					break
				}
				if tok.Kind != "word" || tok.Text != name {
					continue
				}
				if j+6 < end && ts[j-1].Text == ";" && ts[j+1].Text == "=" {
					literal := ts[j+2].Text == "&" && ts[j+3].Text == typ && ts[j+4].Text == "{" && ts[j+5].Text == "}" && ts[j+6].Text == ";"
					allocated := ts[j+2].Text == "new" && ts[j+3].Text == "(" && ts[j+4].Text == typ && ts[j+5].Text == ")" && ts[j+6].Text == ";"
					if literal || allocated {
						// Keep declaration comments, including comments between its tokens.
						original := source[ts[i].Start:ts[i+4].End]
						edits = append(edits, Edit{ts[i].Start, ts[i+4].End, commentsIn(original)}, Edit{ts[j+1].Start, ts[j+1].Start, ":"})
					}
				}
				// Any other first use (including an earlier assignment or address-taking)
				// rules out this declaration. Never hunt for a later convenient write.
				break
			}
		}
	}
	return ApplyEdits(source, edits)
}
