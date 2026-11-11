package startingpoint

// rewriteLibraryCalls handles direct C formatting and strlen calls in executable blocks.
// sprintf is lowered only as a standalone write into a simple named buffer;
// uses of its C character-count result need separate repair.
func rewriteLibraryCalls(source string, addImport bool) (string, error) {
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
	qualifier, imported := "fmt.", false
	// Recognize ordinary and grouped imports without matching strings/comments
	// elsewhere in the file. Reuse an explicitly aliased fmt import as well.
	for i, t := range ts {
		if t.Text != "import" {
			continue
		}
		end := i + 2
		if i+1 < len(ts) && ts[i+1].Text == "(" {
			end = pairs[i+1]
		} else if i+2 < len(ts) && ts[i+1].Kind != "string" {
			end = i + 3
		}
		if end > len(ts) {
			end = len(ts)
		}
		for j := i + 1; j < end; j++ {
			if ts[j].Kind != "string" || (ts[j].Text != `"fmt"` && ts[j].Text != "`fmt`") {
				continue
			}
			imported = true
			if j > i+1 && (ts[j-1].Kind == "word" || ts[j-1].Text == ".") {
				alias := ts[j-1].Text
				if alias == "_" {
					imported = false
				} else if alias == "." {
					qualifier = ""
				} else {
					qualifier = alias + "."
				}
			}
		}
	}
	var edits []Edit
	depth, changed := 0, false
	for i, t := range ts {
		if t.Text == "{" {
			depth++
			continue
		}
		if t.Text == "}" {
			depth--
			continue
		}
		if depth == 0 || t.Kind != "word" || (t.Text != "printf" && t.Text != "fprintf" && t.Text != "sprintf" && t.Text != "strlen") {
			continue
		}
		if i > 0 && (ts[i-1].Text == "." || ts[i-1].Text == "->" || ts[i-1].Text == "::") {
			continue
		}
		if i+1 >= len(ts) || ts[i+1].Text != "(" {
			continue
		}
		end, ok := pairs[i+1]
		if !ok || end+1 < len(ts) && ts[end+1].Text == "{" {
			continue
		}
		replacement := qualifier + "Printf"
		if t.Text == "strlen" {
			replacement = "len"
		}
		if t.Text == "fprintf" {
			replacement = qualifier + "Fprintf"
		}
		if t.Text == "sprintf" {
			if i == 0 || (ts[i-1].Text != "{" && ts[i-1].Text != "}" && ts[i-1].Text != ";") {
				continue
			}
			if i+4 >= end || ts[i+2].Kind != "word" || ts[i+3].Text != "," || end+1 >= len(ts) || ts[end+1].Text != ";" {
				continue
			}
			replacement = ts[i+2].Text + " = " + qualifier + "Sprintf"
			edits = append(edits, Edit{ts[i+2].Start, ts[i+2].End, ""}, Edit{ts[i+3].Start, ts[i+3].End, ""})
		}
		edits = append(edits, Edit{t.Start, t.End, replacement})
		changed = changed || t.Text != "strlen"
	}
	if changed && addImport && !imported {
		// Upgrade a blank fmt import rather than introducing a duplicate path.
		upgraded := false
		for i, t := range ts {
			if i > 0 && t.Kind == "string" && (t.Text == `"fmt"` || t.Text == "`fmt`") && ts[i-1].Text == "_" {
				edits = append(edits, Edit{ts[i-1].Start, ts[i-1].End, ""})
				upgraded = true
				break
			}
		}
		if !upgraded {
			for i, t := range ts {
				if t.Text == "package" && i+1 < len(ts) {
					edits = append(edits, Edit{ts[i+1].End, ts[i+1].End, "\n\nimport \"fmt\""})
					break
				}
			}
		}
	}
	result, err := ApplyEdits(source, edits)
	return result, err
}
