package startingpoint

import (
	"regexp"
	"strings"
)

// freeAndNull matches the common ABC_FREE expansion. Its restricted lvalue
// grammar deliberately excludes arbitrary expressions so the lowering cannot
// change evaluation count or ownership behavior accidentally.
var freeAndNull = regexp.MustCompile(`\(\s*\(\s*([A-Za-z_]\w*(?:\s*(?:\.\s*[A-Za-z_]\w*|\[\s*[A-Za-z_]\w*\s*\]))*)\s*\)\s*\?\s*\(\s*free\s*\(\s*\(\s*(?:char|void)\s*\*\s*\)\s*\(\s*([A-Za-z_]\w*(?:\s*(?:\.\s*[A-Za-z_]\w*|\[\s*[A-Za-z_]\w*\s*\]))*)\s*\)\s*\)\s*,\s*\(\s*([A-Za-z_]\w*(?:\s*(?:\.\s*[A-Za-z_]\w*|\[\s*[A-Za-z_]\w*\s*]))*)\s*\)\s*=\s*(?:0|NULL)\s*\)\s*:\s*(?:0|NULL)\s*\)`)

// lowerFreeAndNull replaces the null-safe free-and-clear idiom with a Go nil
// assignment. Comments and literals are excluded so examples remain intact.
func lowerFreeAndNull(source string) (string, error) {
	ts, err := Lex(source)
	if err != nil {
		return "", err
	}
	matches := freeAndNull.FindAllStringSubmatchIndex(source, -1)
	var edits []Edit
	for _, m := range matches {
		start, end := m[0], m[1]
		inert := false
		for _, t := range ts {
			if t.Start >= end {
				break
			}
			if t.End > start && (t.Kind == "comment" || t.Kind == "string") {
				inert = true
				break
			}
		}
		if inert || compactLvalue(source[m[2]:m[3]]) != compactLvalue(source[m[4]:m[5]]) || compactLvalue(source[m[2]:m[3]]) != compactLvalue(source[m[6]:m[7]]) {
			continue
		}
		edits = append(edits, Edit{Start: start, End: end, Text: source[m[2]:m[3]] + " = nil"})
	}
	return ApplyEdits(source, edits)
}

func compactLvalue(s string) string {
	return strings.Join(strings.Fields(s), "")
}
