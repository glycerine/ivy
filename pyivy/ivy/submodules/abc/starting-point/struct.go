package startingpoint

import (
	"regexp"
	"strings"
)

var structAlias = regexp.MustCompile(`\btypedef\s+struct\s+([A-Za-z_]\w*)\s+([A-Za-z_]\w*)\s*;`)
var structHeader = regexp.MustCompile(`^\s*struct\s+([A-Za-z_]\w*)\s*\{`)

// structEdits joins adjacent typedef/struct pairs (comments may separate them).
// Edits use original offsets; comments and field line breaks remain in place.
func structEdits(source string, ts []Token, ends map[int]int, path string) ([]Edit, []Issue) {
	mask := Mask(ts)
	var edits []Edit
	var issues []Issue
	for _, a := range structAlias.FindAllStringSubmatchIndex(mask, -1) {
		// Nested pairs are handled with their enclosing declaration in a later pass.
		level := 0
		for _, t := range ts {
			if t.Start >= a[0] {
				break
			}
			if t.Kind == "punct" {
				if t.Text == "{" {
					level++
				}
				if t.Text == "}" {
					level--
				}
			}
		}
		if level != 0 {
			continue
		}
		tag, alias := mask[a[2]:a[3]], mask[a[4]:a[5]]
		h := structHeader.FindStringSubmatchIndex(mask[a[1]:])
		if h == nil || mask[a[1]+h[2]:a[1]+h[3]] != tag {
			continue
		}
		open := a[1] + h[1] - 1
		end, ok := ends[open]
		if !ok {
			continue
		}
		// Do not cross preprocessor directives or erase trailing object declarators.
		blocked := false
		for _, t := range ts {
			if t.Start >= a[0] && t.End <= end && t.Kind == "directive" {
				blocked = true
			}
		}
		tail := end
		for tail < len(mask) && strings.ContainsRune(" \t\r\n", rune(mask[tail])) {
			tail++
		}
		if blocked || tail == len(mask) || mask[tail] != ';' {
			continue
		}
		// Remove typedef code tokens only, retaining its comments and whitespace.
		for _, t := range ts {
			if t.Start >= a[0] && t.End <= a[1] && t.Kind != "space" && t.Kind != "comment" {
				edits = append(edits, Edit{t.Start, t.End, ""})
			}
		}
		// Change struct tag into type alias struct without consuming header comments.
		var header []Token
		for _, t := range ts {
			if t.Start >= a[1] && t.End <= open && t.Kind != "space" && t.Kind != "comment" {
				header = append(header, t)
			}
		}
		if len(header) != 2 {
			continue
		}
		edits = append(edits, Edit{header[0].Start, header[0].End, "type"}, Edit{header[1].Start, header[1].End, alias + " struct"}, Edit{tail, tail + 1, ""})
		edits = append(edits, structBraceEdits(source, header[1].End, open)...)
		fe, fi := structFieldEdits(source, ts, open, end, tag, alias, path)
		edits = append(edits, fe...)
		issues = append(issues, fi...)
		issues = append(issues, Issue{path + "::struct::" + alias, "struct-review", alias, "open", "Typedef/struct pair combined; review pointer versus slice representation and any retained complex fields.", a[0], tail + 1})
	}
	se, si := standaloneStructEdits(source, ts, ends, path, edits)
	edits = append(edits, se...)
	issues = append(issues, si...)
	return edits, issues
}

func fieldEdits(f []Token, tag, alias string) ([]Edit, bool) {
	if len(f) < 2 || f[0].Kind != "word" || f[len(f)-1].Kind != "word" {
		return nil, false
	}
	base, n := f[0].Text, 1
	if base == "struct" {
		if len(f) < 3 || f[1].Text != tag {
			return nil, false
		}
		base = alias
		n = 2
	}
	stars := ""
	for _, t := range f[n : len(f)-1] {
		if t.Text != "*" {
			return nil, false
		}
		stars += "*"
	}
	name := f[len(f)-1]
	edits := []Edit{{f[0].Start, f[0].End, name.Text}}
	for _, t := range f[1 : len(f)-1] {
		edits = append(edits, Edit{t.Start, t.End, ""})
	}
	// The name's original position follows whitespace or a removed star. Ensure
	// a separator for spellings such as unsigned*data with no spaces.
	sep := ""
	adjacent := true
	for i := 1; i < len(f); i++ {
		if f[i-1].End != f[i].Start {
			adjacent = false
		}
	}
	if adjacent {
		sep = " "
	}
	edits = append(edits, Edit{name.Start, name.End, sep + goType(base+stars)})
	return edits, true
}

func structFieldEdits(source string, ts []Token, open, end int, tag, alias, path string) ([]Edit, []Issue) {
	var edits []Edit
	var issues []Issue
	// Only ordinary scalar/pointer fields are recognized. Complex declarations
	// stay visible in the otherwise transformed struct for second-pass repair.
	start, depth := open+1, 0
	for _, t := range ts {
		if t.Start < start || t.Start >= end-1 {
			continue
		}
		if t.Kind == "punct" {
			if strings.Contains("([{", t.Text) {
				depth++
			}
			if strings.Contains(")]}", t.Text) {
				depth--
			}
		}
		if t.Text != ";" || t.Kind != "punct" || depth != 0 {
			continue
		}
		var field []Token
		for _, f := range ts {
			if f.Start >= start && f.End <= t.Start && f.Kind != "space" && f.Kind != "comment" {
				field = append(field, f)
			}
		}
		if len(field) > 0 {
			fe, valid := fieldEdits(field, tag, alias)
			if valid {
				edits = append(edits, fe...)
			} else {
				issues = append(issues, Issue{path + "::struct::" + alias + "::" + strings.TrimSpace(source[start:t.End]), "struct-field-review", alias, "open", "Complex field declaration retained verbatim; repair its Go type and representation.", start, t.End})
			}
		}
		start = t.End
	}
	return edits, issues
}

// Keep the opening brace on the type header line without discarding comments.
func structBraceEdits(source string, headerEnd, open int) []Edit {
	gap := source[headerEnd:open]
	if strings.TrimSpace(gap) == "" {
		return []Edit{{headerEnd, open, " "}}
	}
	if strings.ContainsAny(gap, "\r\n") {
		return []Edit{{headerEnd, headerEnd, " {"}, {open, open + 1, ""}}
	}
	return nil
}
