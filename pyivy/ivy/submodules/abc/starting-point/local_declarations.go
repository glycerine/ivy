package startingpoint

import "strings"

// lowerLocalDeclarations groups simple declarators whose zero values suffice,
// or emits separate declarations to retain initializer order and pointer types.
func lowerLocalDeclarations(source string, types map[string]bool) (string, error) {
	return lowerDeclarations(source, types, false)
}

func lowerDeclarations(source string, types map[string]bool, allInitializers bool) (string, error) {
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
	for i, t := range ts {
		if t.Kind != "word" || !isDeclarationType(t.Text, types) {
			continue
		}
		if i == 0 || (ts[i-1].Text != "{" && ts[i-1].Text != "}" && ts[i-1].Text != ";" && !labelEnds[ts[i-1].End]) {
			continue
		}
		// Only statement declarations; for-header declarations need separate lowering.
		end, start := i+1, i+1
		var parts [][2]int
		for ; end < len(ts); end++ {
			if ts[end].Kind == "directive" || ts[end].Text == "}" {
				break
			}
			if ts[end].Text == ";" {
				parts = append(parts, [2]int{start, end})
				break
			}
			if ts[end].Text == "," {
				parts = append(parts, [2]int{start, end})
				start = end + 1
				continue
			}
			if close, ok := pairs[end]; ok && close > end {
				end = close
			}
		}
		if end >= len(ts) || ts[end].Text != ";" {
			continue
		}
		type decl struct{ name, typ, value string }
		var declarations []decl
		valid, allZero := true, true
		hasArray := false
		for _, part := range parts {
			j, stars := part[0], ""
			for j < part[1] && ts[j].Text == "*" {
				stars += "*"
				j++
			}
			if j >= part[1] || ts[j].Kind != "word" {
				valid = false
				break
			}
			d := decl{name: ts[j].Text, typ: goType(t.Text + stars)}
			j++
			suffix := ""
			for j < part[1] && ts[j].Text == "[" {
				close, ok := pairs[j]
				if !ok || close >= part[1] {
					valid = false
					break
				}
				suffix += source[ts[j].Start:ts[close].End]
				j = close + 1
			}
			if !valid {
				break
			}
			if suffix != "" {
				hasArray = true
				d.typ = arrayType(d.typ, suffix, false)
			}
			if j < part[1] {
				if ts[j].Text != "=" || j+1 >= part[1] {
					valid = false
					break
				}
				d.value = strings.TrimSpace(source[ts[j+1].Start:ts[part[1]-1].End])
				if d.value == "0" {
					d.value = ""
				}
			}
			allZero = allZero && d.value == ""
			declarations = append(declarations, d)
		}
		if !valid || len(declarations) == 0 || (len(declarations) == 1 && !allZero && !allInitializers && !hasArray) {
			continue
		}
		// Preserve comments exactly by leaving commented declarations to the repair pass.
		commented := false
		for _, tok := range all {
			if tok.Start >= t.Start && tok.End <= ts[end].End && tok.Kind == "comment" {
				commented = true
				break
			}
		}
		if commented {
			continue
		}
		sameType := true
		var names []string
		for _, d := range declarations {
			sameType = sameType && d.typ == declarations[0].typ
			names = append(names, d.name)
		}
		text := ""
		if allZero && sameType {
			text = "var " + strings.Join(names, ", ") + " " + declarations[0].typ + ";"
		} else {
			line := strings.LastIndex(source[:t.Start], "\n") + 1
			indent := source[line:t.Start]
			if strings.TrimSpace(indent) != "" {
				indent = ""
			}
			var lines []string
			for _, d := range declarations {
				s := "var " + d.name + " " + d.typ
				if d.value != "" {
					s += " = " + d.value
				}
				lines = append(lines, s+";")
			}
			text = strings.Join(lines, "\n"+indent)
		}
		edits = append(edits, Edit{t.Start, ts[end].End, text})
	}
	return ApplyEdits(source, edits)
}
