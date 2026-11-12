package startingpoint

import "strings"

// Carry block-comment markers through expression rewriting, then move the
// explanation above the containing code line. A // inside a return or argument
// would change semicolon insertion or swallow the rest of the expression.
func macroInliningComments(source string) (string, error) {
	ts, err := Lex(source)
	if err != nil {
		return "", err
	}
	const prefix = "/*__starting_point_inline_macro:"
	const suffix = "__*/"
	lines := map[int][]string{}
	var removals []Edit
	for _, t := range ts {
		if t.Kind != "comment" || !strings.HasPrefix(t.Text, prefix) || !strings.HasSuffix(t.Text, suffix) {
			continue
		}
		name := strings.TrimSuffix(strings.TrimPrefix(t.Text, prefix), suffix)
		if !identifier.MatchString(name) {
			continue
		}
		start := strings.LastIndex(source[:t.Start], "\n") + 1
		duplicate := false
		for _, old := range lines[start] {
			if old == name {
				duplicate = true
			}
		}
		if !duplicate {
			lines[start] = append(lines[start], name)
		}
		removals = append(removals, Edit{t.Start, t.End, ""})
	}
	var edits []Edit
	for start, names := range lines {
		end := start
		for end < len(source) && (source[end] == ' ' || source[end] == '\t') {
			end++
		}
		indent := source[start:end]
		var comment strings.Builder
		for _, name := range names {
			comment.WriteString(indent + "// inlining of macro " + name + "\n")
		}
		edits = append(edits, Edit{start, start, comment.String()})
	}
	edits = append(edits, removals...)
	return ApplyEdits(source, edits)
}
