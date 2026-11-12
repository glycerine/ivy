package startingpoint

// lowerEnums handles standalone file-scope enum definitions. It preserves
// comments with token edits and leaves trailing variable declarators untouched.
func lowerEnums(source string) (string, error) {
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
	var structStack []bool
	for i := 0; i < len(ts); i++ {
		if ts[i].Text == "{" {
			nested := len(structStack) > 0 && structStack[len(structStack)-1]
			if i > 0 && (ts[i-1].Text == "struct" || ts[i-1].Text == "union") {
				nested = true
			}
			if i > 1 && (ts[i-2].Text == "struct" || ts[i-2].Text == "union") {
				nested = true
			}
			structStack = append(structStack, nested)
			continue
		}
		if ts[i].Text == "}" {
			if len(structStack) > 0 {
				structStack = structStack[:len(structStack)-1]
			}
			continue
		}
		if (len(structStack) > 0 && structStack[len(structStack)-1]) || ts[i].Text != "enum" {
			continue
		}
		typedef := i > 0 && ts[i-1].Text == "typedef"
		open, tag := i+1, ""
		if open < len(ts) && ts[open].Kind == "word" {
			tag = ts[open].Text
			open++
		}
		if typedef && tag != "" && open+1 < len(ts) && ts[open].Kind == "word" && ts[open+1].Text == ";" {
			edits = append(edits, Edit{ts[i-1].Start, ts[i-1].End, "type"}, Edit{ts[i].Start, ts[i].End, ts[open].Text + " ="}, Edit{ts[open].Start, ts[open].End, ""})
			i = open + 1
			continue
		}
		if open >= len(ts) || ts[open].Text != "{" {
			continue
		}
		close, ok := pairs[open]
		if !ok {
			continue
		}
		tail := close + 1
		name := tag
		if typedef {
			if tail >= len(ts) || ts[tail].Kind != "word" {
				continue
			}
			name = ts[tail].Text
			tail++
		}
		if tail >= len(ts) || ts[tail].Text != ";" {
			continue
		}
		type member struct {
			first, last, comma int
			explicit           bool
		}
		var members []member
		valid, hasExplicit := true, false
		for at := open + 1; at < close; {
			if ts[at].Kind != "word" {
				valid = false
				break
			}
			end := at + 1
			for end < close && ts[end].Text != "," {
				if ts[end].Kind == "directive" {
					valid = false
					break
				}
				if next, ok := pairs[end]; ok && next > end {
					end = next
				}
				end++
			}
			if !valid {
				break
			}
			explicit := end > at+1
			if explicit && (ts[at+1].Text != "=" || end == at+2) {
				valid = false
				break
			}
			hasExplicit = hasExplicit || explicit
			comma := -1
			if end < close {
				comma = end
			}
			members = append(members, member{at, end, comma, explicit})
			at = end + 1
		}
		if !valid || len(members) == 0 {
			continue
		}
		header := "const"
		if name != "" {
			header = "type " + name + " int32\n"
			if tag != "" && tag != name {
				header += "type " + tag + " = " + name + "\n"
			}
			header += "const"
		}
		if typedef {
			edits = append(edits, Edit{ts[i-1].Start, ts[i-1].End, ""}, Edit{ts[close+1].Start, ts[close+1].End, ""})
		}
		edits = append(edits, Edit{ts[i].Start, ts[i].End, header}, Edit{ts[open].Start, ts[open].End, "("}, Edit{ts[close].Start, ts[close].End, "\n)\n"}, Edit{ts[tail].Start, ts[tail].End, ""})
		if tag != "" {
			edits = append(edits, Edit{ts[i+1].Start, ts[i+1].End, ""})
		}
		for n, m := range members {
			suffix := ""
			if name != "" && (hasExplicit || n == 0) {
				suffix = " " + name
			}
			if !m.explicit {
				switch {
				case !hasExplicit && n == 0:
					suffix += " = iota"
				case hasExplicit && n == 0:
					suffix += " = 0"
				case hasExplicit:
					suffix += " = " + ts[members[n-1].first].Text + " + 1"
				}
			}
			if suffix != "" {
				edits = append(edits, Edit{ts[m.first].End, ts[m.first].End, suffix})
			}
			if m.comma >= 0 {
				edits = append(edits, Edit{ts[m.comma].Start, ts[m.comma].End, "\n"})
			}
		}
		i = tail
	}
	return ApplyEdits(source, edits)
}
