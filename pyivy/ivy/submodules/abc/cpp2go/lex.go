package cpp2go

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
				return nil, fmt.Errorf("unterminated comment at byte %d", i)
			}
			i += end + 4
		case rawPrefix(s[i:]) != "":
			kind = "string"
			p := rawPrefix(s[i:])
			d := i + len(p)
			k := strings.IndexByte(s[d:], '(')
			if k < 0 || k > 16 || strings.ContainsAny(s[d:d+k], " \t\r\n\\)") {
				return nil, fmt.Errorf("invalid raw string at byte %d", i)
			}
			end := strings.Index(s[d+k+1:], ")"+s[d:d+k]+"\"")
			if end < 0 {
				return nil, fmt.Errorf("unterminated raw string at byte %d", i)
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
				return nil, fmt.Errorf("unterminated literal at byte %d", start)
			}
		case s[i] == '#':
			kind = "directive"
			i++
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
				return nil, fmt.Errorf("unmatched %s at %d", t.Text, t.Start)
			}
			j := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if strings.Index("([{", ts[j].Text) != strings.Index(")]}", t.Text) {
				return nil, fmt.Errorf("mismatched delimiter at %d", t.Start)
			}
			m[j] = i
		}
	}
	if len(stack) > 0 {
		return nil, fmt.Errorf("unclosed delimiter at %d", ts[stack[0]].Start)
	}
	return m, nil
}
