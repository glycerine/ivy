package startingpoint

import (
	"fmt"
	"strings"
)

// Token includes trivia: concatenating Text reconstructs the input exactly.
// Offsets are byte offsets into the original, unnormalized source.
type Token struct {
	Kind       string
	Start, End int
	Text       string
}

func Lex(s string) ([]Token, error) {
	var out []Token
	for i := 0; i < len(s); {
		start, kind := i, "punct"
		switch {
		case strings.ContainsRune(" \t\r\n\v\f", rune(s[i])):
			kind = "space"
			for i < len(s) && strings.ContainsRune(" \t\r\n\v\f", rune(s[i])) {
				i++
			}
		case strings.HasPrefix(s[i:], "//"):
			kind = "comment"
			i += 2
			for i < len(s) {
				if s[i] == '\\' && i+1 < len(s) && s[i+1] == '\n' {
					i += 2
					continue
				}
				if s[i] == '\\' && i+2 < len(s) && s[i+1:i+3] == "\r\n" {
					i += 3
					continue
				}
				if s[i] == '\n' {
					break
				}
				i++
			}
		case strings.HasPrefix(s[i:], "/*"):
			kind = "comment"
			end := strings.Index(s[i+2:], "*/")
			if end < 0 {
				return nil, &syntaxError{"unterminated comment", i}
			}
			i += end + 4
		case rawPrefix(s[i:]) != "":
			kind = "string"
			p := rawPrefix(s[i:])
			d := i + len(p)
			k := strings.IndexByte(s[d:], '(')
			if k < 0 || k > 16 || strings.ContainsAny(s[d:d+k], " \t\r\n\\)") {
				return nil, &syntaxError{"invalid raw string", i}
			}
			end := strings.Index(s[d+k+1:], ")"+s[d:d+k]+"\"")
			if end < 0 {
				return nil, &syntaxError{"unterminated raw string", i}
			}
			i = d + k + 1 + end + 1 + k + 1
		case s[i] == '"' || s[i] == '\'':
			kind = "string"
			q := s[i]
			i++
			closed := false
			for i < len(s) {
				if s[i] == '\\' {
					i += 2
					continue
				}
				if s[i] == q {
					i++
					closed = true
					break
				}
				i++
			}
			if !closed {
				return nil, &syntaxError{"unterminated literal", start}
			}
		case s[i] == '#' && directiveStart(out):
			kind = "directive"
			end, err := directiveEnd(s, i)
			if err != nil {
				return nil, err
			}
			i = end
		case ident(s[i]):
			kind = "word"
			i++
			for i < len(s) && (ident(s[i]) || s[i] >= '0' && s[i] <= '9') {
				i++
			}
		default:
			i++
			if i < len(s) && (s[start:i+1] == "->" || s[start:i+1] == "::") {
				i++
			}
		}
		out = append(out, Token{kind, start, i, s[start:i]})
	}
	return out, nil
}

func ident(b byte) bool { return b == '_' || b >= 'a' && b <= 'z' || b >= 'A' && b <= 'Z' }
func rawPrefix(s string) string {
	for _, p := range []string{"u8R\"", "uR\"", "UR\"", "LR\"", "R\""} {
		if strings.HasPrefix(s, p) {
			return p
		}
	}
	return ""
}

// Mask retains offsets and newlines while hiding protected text from regex rules.
func Mask(ts []Token) string {
	var b strings.Builder
	for _, t := range ts {
		if t.Kind != "comment" && t.Kind != "string" && t.Kind != "directive" {
			b.WriteString(t.Text)
			continue
		}
		for _, c := range []byte(t.Text) {
			if c == '\n' || c == '\r' {
				b.WriteByte(c)
			} else {
				b.WriteByte(' ')
			}
		}
	}
	return b.String()
}

// MatchDelimiters ignores trivia and literals and rejects malformed nesting.
func MatchDelimiters(ts []Token) (map[int]int, error) {
	m := map[int]int{}
	var stack []int
	for i, t := range ts {
		if t.Kind != "punct" {
			continue
		}
		if strings.Contains("([{", t.Text) {
			stack = append(stack, i)
		}
		if strings.Contains(")]}", t.Text) {
			if len(stack) == 0 {
				return nil, &syntaxError{fmt.Sprintf("unmatched %s", t.Text), t.Start}
			}
			j := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if strings.Index("([{", ts[j].Text) != strings.Index(")]}", t.Text) {
				return nil, &syntaxError{"mismatched delimiter", t.Start}
			}
			m[j] = i
		}
	}
	if len(stack) > 0 {
		return nil, &syntaxError{"unclosed delimiter", ts[stack[0]].Start}
	}
	return m, nil
}

// directiveEnd keeps multiline comments inside the directive token. Literal
// and line-comment state prevents comment-like text from opening a block.
func directiveEnd(s string, start int) (int, error) {
	quote := byte(0)
	lineComment := false
	for i := start + 1; i < len(s); {
		if s[i] == '\\' && i+1 < len(s) && s[i+1] == '\n' {
			i += 2
			continue
		}
		if s[i] == '\\' && i+2 < len(s) && s[i+1:i+3] == "\r\n" {
			i += 3
			continue
		}
		if s[i] == '\n' {
			return i, nil
		}
		if lineComment {
			i++
			continue
		}
		if quote != 0 {
			if s[i] == '\\' && i+1 < len(s) {
				i += 2
				continue
			}
			if s[i] == quote {
				quote = 0
			}
			i++
			continue
		}
		if strings.HasPrefix(s[i:], "//") {
			lineComment = true
			i += 2
			continue
		}
		if strings.HasPrefix(s[i:], "/*") {
			end := strings.Index(s[i+2:], "*/")
			if end < 0 {
				return 0, &syntaxError{"unterminated comment", i}
			}
			i += end + 4
			continue
		}
		if s[i] == '"' || s[i] == '\'' {
			quote = s[i]
		}
		i++
	}
	return len(s), nil
}

// Directives start at the first non-trivia token on a logical line. Hashes
// within macro replacement drafts are stringification/token-pasting operators.
func directiveStart(tokens []Token) bool {
	for i := len(tokens) - 1; i >= 0; i-- {
		t := tokens[i]
		if t.Kind != "space" && t.Kind != "comment" {
			return false
		}
		if strings.Contains(t.Text, "\n") {
			if t.Kind == "space" && i > 0 && tokens[i-1].Text == "\\" && strings.Count(t.Text, "\n") == 1 {
				i--
				continue
			}
			return true
		}
	}
	return true
}
