package ivy2cpp

import (
	"fmt"
	"strings"
)

type CPPToken struct {
	Text   string
	Offset int
}

type CPPTokenDiff struct {
	Equal        bool
	Index        int
	LeftName     string
	RightName    string
	LeftToken    string
	RightToken   string
	LeftContext  string
	RightContext string
	Message      string
}

func TokenizeCPP(text string) []CPPToken {
	var toks []CPPToken
	for i := 0; i < len(text); {
		c := text[i]
		if isCPPWhitespace(c) {
			i++
			continue
		}
		if c == '/' && i+1 < len(text) {
			switch text[i+1] {
			case '/':
				i += 2
				for i < len(text) && text[i] != '\n' {
					i++
				}
				continue
			case '*':
				i += 2
				for i+1 < len(text) && !(text[i] == '*' && text[i+1] == '/') {
					i++
				}
				if i+1 < len(text) {
					i += 2
				}
				continue
			}
		}
		if end, ok := scanCPPRawString(text, i); ok {
			toks = append(toks, CPPToken{Text: text[i:end], Offset: i})
			i = end
			continue
		}
		if c == '"' || c == '\'' {
			end := scanCPPQuoted(text, i)
			toks = append(toks, CPPToken{Text: text[i:end], Offset: i})
			i = end
			continue
		}
		if isCPPIdentStart(c) {
			j := i + 1
			for j < len(text) && isCPPIdentPart(text[j]) {
				j++
			}
			toks = append(toks, CPPToken{Text: text[i:j], Offset: i})
			i = j
			continue
		}
		if isCPPDigit(c) || (c == '.' && i+1 < len(text) && isCPPDigit(text[i+1])) {
			j := scanCPPNumber(text, i)
			toks = append(toks, CPPToken{Text: text[i:j], Offset: i})
			i = j
			continue
		}
		if op, ok := scanCPPOperator(text, i); ok {
			toks = append(toks, CPPToken{Text: op, Offset: i})
			i += len(op)
			continue
		}
		toks = append(toks, CPPToken{Text: text[i : i+1], Offset: i})
		i++
	}
	return toks
}

func CompareCPPTokens(leftName, leftText, rightName, rightText string) CPPTokenDiff {
	left := TokenizeCPP(leftText)
	right := TokenizeCPP(rightText)
	limit := len(left)
	if len(right) < limit {
		limit = len(right)
	}
	for i := 0; i < limit; i++ {
		if left[i].Text == right[i].Text {
			continue
		}
		return cppTokenDiff(leftName, leftText, rightName, rightText, left, right, i)
	}
	if len(left) != len(right) {
		return cppTokenDiff(leftName, leftText, rightName, rightText, left, right, limit)
	}
	return CPPTokenDiff{Equal: true, Index: -1, LeftName: leftName, RightName: rightName}
}

func (d CPPTokenDiff) Error() string {
	if d.Equal {
		return ""
	}
	if d.Message != "" {
		return d.Message
	}
	return fmt.Sprintf("C++ token divergence at token %d: %s has %q, %s has %q\n%s context: %s\n%s context: %s",
		d.Index, d.LeftName, d.LeftToken, d.RightName, d.RightToken, d.LeftName, d.LeftContext, d.RightName, d.RightContext)
}

func cppTokenDiff(leftName, leftText, rightName, rightText string, left, right []CPPToken, index int) CPPTokenDiff {
	leftTok, rightTok := "<eof>", "<eof>"
	leftOff, rightOff := len(leftText), len(rightText)
	if index < len(left) {
		leftTok = left[index].Text
		leftOff = left[index].Offset
	}
	if index < len(right) {
		rightTok = right[index].Text
		rightOff = right[index].Offset
	}
	diff := CPPTokenDiff{
		Equal:        false,
		Index:        index,
		LeftName:     leftName,
		RightName:    rightName,
		LeftToken:    leftTok,
		RightToken:   rightTok,
		LeftContext:  cppContext(leftText, leftOff, 80),
		RightContext: cppContext(rightText, rightOff, 80),
	}
	diff.Message = diff.Error()
	return diff
}

func scanCPPQuoted(text string, start int) int {
	quote := text[start]
	i := start + 1
	for i < len(text) {
		if text[i] == '\\' {
			i += 2
			continue
		}
		if text[i] == quote {
			return i + 1
		}
		i++
	}
	return len(text)
}

func scanCPPRawString(text string, start int) (int, bool) {
	prefixes := []string{`u8R"`, `uR"`, `UR"`, `LR"`, `R"`}
	for _, prefix := range prefixes {
		if !strings.HasPrefix(text[start:], prefix) {
			continue
		}
		openQuote := start + len(prefix) - 1
		openParen := openQuote + 1
		for openParen < len(text) && text[openParen] != '(' {
			openParen++
		}
		if openParen >= len(text) {
			return len(text), true
		}
		delim := text[openQuote+1 : openParen]
		closeSeq := ")" + delim + `"`
		search := openParen + 1
		if idx := strings.Index(text[search:], closeSeq); idx >= 0 {
			return search + idx + len(closeSeq), true
		}
		return len(text), true
	}
	return 0, false
}

func scanCPPNumber(text string, start int) int {
	i := start
	if text[i] == '.' {
		i++
	}
	for i < len(text) {
		c := text[i]
		if isCPPIdentPart(c) || c == '.' || c == '\'' {
			i++
			continue
		}
		if (c == '+' || c == '-') && i > start {
			prev := text[i-1]
			if prev == 'e' || prev == 'E' || prev == 'p' || prev == 'P' {
				i++
				continue
			}
		}
		break
	}
	return i
}

func scanCPPOperator(text string, start int) (string, bool) {
	for _, ops := range [][]string{
		{"<<=", ">>=", "...", "->*"},
		{"::", "->", "++", "--", "==", "!=", "<=", ">=", "&&", "||", "<<", ">>", "+=", "-=", "*=", "/=", "%=", "&=", "|=", "^=", "##", ".*"},
	} {
		for _, op := range ops {
			if strings.HasPrefix(text[start:], op) {
				return op, true
			}
		}
	}
	return "", false
}

func cppContext(text string, offset, width int) string {
	if offset < 0 {
		offset = 0
	}
	if offset > len(text) {
		offset = len(text)
	}
	half := width / 2
	start := offset - half
	if start < 0 {
		start = 0
	}
	end := start + width
	if end > len(text) {
		end = len(text)
		start = end - width
		if start < 0 {
			start = 0
		}
	}
	ctx := text[start:end]
	ctx = strings.ReplaceAll(ctx, "\n", `\n`)
	ctx = strings.ReplaceAll(ctx, "\t", `\t`)
	return ctx
}

func isCPPWhitespace(c byte) bool {
	return c == ' ' || c == '\n' || c == '\r' || c == '\t' || c == '\f' || c == '\v'
}

func isCPPDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func isCPPAlpha(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func isCPPIdentStart(c byte) bool {
	return isCPPAlpha(c) || c == '_' || c == '$'
}

func isCPPIdentPart(c byte) bool {
	return isCPPIdentStart(c) || isCPPDigit(c)
}
