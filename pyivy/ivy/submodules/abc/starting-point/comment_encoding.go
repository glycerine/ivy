package startingpoint

import "unicode/utf8"

// SanitizeCommentUTF8 replaces each invalid UTF-8 byte in a comment with '_'.
// Valid Unicode (including a literal replacement character) and non-comment
// bytes are unchanged. One-byte substitutions preserve byte offsets.
func SanitizeCommentUTF8(source string) (string, error) {
	ts, err := Lex(source)
	if err != nil {
		return "", err
	}
	var edits []Edit
	for _, t := range ts {
		if t.Kind != "comment" {
			continue
		}
		for i := 0; i < len(t.Text); {
			r, n := utf8.DecodeRuneInString(t.Text[i:])
			if r == utf8.RuneError && n == 1 {
				edits = append(edits, Edit{t.Start + i, t.Start + i + 1, "_"})
			}
			i += n
		}
	}
	return ApplyEdits(source, edits)
}
