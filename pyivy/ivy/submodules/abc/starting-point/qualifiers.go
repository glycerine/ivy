package startingpoint

// stripConstQualifiers removes C/C++ const tokens without touching comments,
// literals or preprocessor text. This pass processes C-derived drafts, not
// arbitrary Go files (where const declarations have their own meaning).
func stripConstQualifiers(source string) (string, error) {
	ts, err := Lex(source)
	if err != nil {
		return "", err
	}
	var edits []Edit
	for _, t := range ts {
		if t.Kind != "word" || t.Text != "const" {
			continue
		}
		end := t.End
		for end < len(source) && (source[end] == ' ' || source[end] == '\t') {
			end++
		}
		edits = append(edits, Edit{t.Start, end, ""})
	}
	return ApplyEdits(source, edits)
}
