package startingpoint

import (
	"regexp"
	"strings"
)

var directiveParts = regexp.MustCompile(`^#\s*([A-Za-z_]\w*)\b([\s\S]*)$`)

// headerGuardEdits recognizes an outer #ifndef NAME / #define NAME pair
// enclosing the whole file. Nested conditionals are counted, not removed.
func headerGuardEdits(ts []Token) []Edit {
	var significant []Token
	for _, t := range ts {
		if t.Kind != "space" && t.Kind != "comment" {
			significant = append(significant, t)
		}
	}
	if len(significant) < 3 {
		return nil
	}
	parse := func(t Token) (string, string) {
		if t.Kind != "directive" {
			return "", ""
		}
		// Strip comments from directives for matching, retaining the original token
		// span so only the recognized wrapper is removed.
		content := strings.TrimPrefix(t.Text, "#")
		parts, err := Lex(content)
		if err != nil {
			return "", ""
		}
		m := directiveParts.FindStringSubmatch("#" + Mask(parts))
		if m == nil {
			return "", ""
		}
		return m[1], strings.TrimSpace(m[2])
	}
	kind, name := parse(significant[0])
	if kind != "ifndef" || !identifier.MatchString(name) {
		return nil
	}
	kind, defined := parse(significant[1])
	if kind != "define" || defined != name {
		return nil
	}
	depth := 1
	end := -1
	for i := 2; i < len(significant); i++ {
		k, _ := parse(significant[i])
		switch k {
		case "if", "ifdef", "ifndef":
			depth++
		case "else", "elif":
			if depth == 1 {
				return nil
			}
		case "endif":
			depth--
			if depth == 0 {
				end = i
			}
		}
		if end >= 0 {
			break
		}
	}
	if end != len(significant)-1 {
		return nil
	}
	_, tail := parse(significant[end])
	if tail != "" {
		return nil
	}
	return []Edit{{significant[0].Start, significant[0].End, ""}, {significant[1].Start, significant[1].End, ""}, {significant[end].Start, significant[end].End, ""}}
}
