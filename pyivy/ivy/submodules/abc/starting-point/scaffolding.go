package startingpoint

import (
	"regexp"
	"strings"
)

var includeDirective = regexp.MustCompile(`^#\s*include\b`)
var declaredType = regexp.MustCompile(`\b(?:type|struct|class|enum)\s+([A-Za-z_]\w*)`)

// commentScaffolding runs on the assembled draft, so recognized declarations
// have already been rewritten. Standalone top-level marker words are retained
// as comments; words within expressions, functions or declarations are not.
func commentScaffolding(source string) (string, error) {
	ts, err := Lex(source)
	if err != nil {
		return "", err
	}
	known := map[string]bool{}
	for _, word := range strings.Fields("auto bool char const double enum extern float inline int long register restrict short signed static struct typedef union unsigned void volatile class namespace template typename public private protected virtual explicit friend constexpr using func type var const package import return break continue else case default") {
		known[word] = true
	}
	for _, m := range declaredType.FindAllStringSubmatch(Mask(ts), -1) {
		known[m[1]] = true
	}
	var edits []Edit
	depth := 0
	for i, t := range ts {
		if t.Kind == "directive" {
			if includeDirective.MatchString(t.Text) {
				edits = append(edits, Edit{t.Start, t.End, "//" + strings.ReplaceAll(t.Text, "\n", "\n//")})
			}
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
		if depth != 0 || t.Kind != "word" || known[t.Text] {
			continue
		}
		start := strings.LastIndex(source[:t.Start], "\n") + 1
		end := strings.IndexByte(source[t.End:], '\n')
		if end < 0 {
			end = len(source)
		} else {
			end += t.End
		}
		// ABC sometimes places its namespace terminator directly after a closing
		// brace. Split it off before applying the standalone-word heuristic.
		if t.Text == "ABC_NAMESPACE_IMPL_END" || t.Text == "ABC_NAMESPACE_HEADER_END" {
			previous := i - 1
			for previous >= 0 && (ts[previous].Kind == "space" || ts[previous].Kind == "comment") {
				previous--
			}
			if previous >= 0 && ts[previous].Text == "}" && strings.TrimSpace(source[start:t.Start]) != "" {
				text := "\n//" + t.Text
				// Keep any following code out of the new line comment as well.
				if strings.TrimSpace(source[t.End:end]) != "" {
					text += "\n"
				}
				edits = append(edits, Edit{t.Start, t.End, text})
				continue
			}
		}
		if strings.TrimSpace(source[start:t.Start]) != "" || strings.TrimSpace(source[t.End:end]) != "" {
			continue
		}
		// A separate-line return type or declarator is not a marker. Preserve
		// tokens followed by a declarator punctuation or a function name + '('.
		var next []Token
		for _, n := range ts[i+1:] {
			if n.Kind == "space" || n.Kind == "comment" {
				continue
			}
			next = append(next, n)
			if len(next) == 2 {
				break
			}
		}
		if len(next) > 0 && (next[0].Text == "(" || next[0].Text == "*" || next[0].Text == ";" || next[0].Text == "{") {
			continue
		}
		if len(next) == 2 && next[0].Kind == "word" && next[1].Text == "(" {
			continue
		}
		edits = append(edits, Edit{t.Start, t.Start, "//"})
	}
	return ApplyEdits(source, edits)
}

var pthreadInclude = regexp.MustCompile(`^#\s*include\s*[<"]pthread\.h[>"]`)

// Unlike general includes, pthread.h is commented even in restored opaque regions.
func commentPthreadIncludes(source string) (string, error) {
	ts, err := Lex(source)
	if err != nil {
		return "", err
	}
	var edits []Edit
	for _, t := range ts {
		if t.Kind == "directive" && pthreadInclude.MatchString(t.Text) {
			edits = append(edits, Edit{t.Start, t.End, "//" + strings.ReplaceAll(t.Text, "\n", "\n//")})
		}
	}
	return ApplyEdits(source, edits)
}
