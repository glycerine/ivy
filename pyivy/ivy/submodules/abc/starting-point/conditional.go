package startingpoint

import (
	"regexp"
	"strings"
)

// conditional splits only an entire conditional expression, respecting nesting,
// strings and comments. It never hoists a branch out of its lazy evaluation site.
func conditional(expr string) (condition, yes, no string, ok bool) {
	expr = strings.TrimSpace(expr)
	ts, err := Lex(expr)
	if err != nil {
		return
	}
	mask := Mask(ts)
	// Peel grouping only when it encloses the entire expression. Keep comments
	// inside those parentheses in the selected condition/branches.
	for strings.HasPrefix(mask, "(") && strings.HasSuffix(mask, ")") {
		depth, end := 0, -1
		for i, c := range mask {
			if c == '(' {
				depth++
			}
			if c == ')' {
				depth--
				if depth == 0 {
					end = i
					break
				}
			}
		}
		if end != len(mask)-1 {
			break
		}
		expr = strings.TrimSpace(expr[1 : len(expr)-1])
		ts, err = Lex(expr)
		if err != nil {
			return
		}
		mask = Mask(ts)
	}
	depth, q, nested := 0, -1, 0
	for i, c := range mask {
		switch c {
		case '(', '[':
			depth++
		case ')', ']':
			depth--
		case '?':
			if depth == 0 {
				if q < 0 {
					q = i
				} else {
					nested++
				}
			}
		case ':':
			if depth == 0 && q >= 0 {
				if nested > 0 {
					nested--
				} else {
					condition = strings.TrimSpace(expr[:q])
					yes = strings.TrimSpace(expr[q+1 : i])
					no = strings.TrimSpace(expr[i+1:])
					ok = condition != "" && yes != "" && no != ""
					return
				}
			}
		}
	}
	return
}

func emitConditional(expr, action, indent, step string) string {
	c, y, n, ok := conditional(expr)
	if !ok {
		return indent + action + strings.TrimSpace(expr) + ";\n"
	}
	return indent + "if (" + c + ") {\n" + emitConditional(y, action, indent+step, step) + indent + "} else {\n" + emitConditional(n, action, indent+step, step) + indent + "}\n"
}

var initializer = regexp.MustCompile(`^\s*(var\s+[A-Za-z_]\w*\s+\*?[A-Za-z_]\w*)\s*=\s*`)
var simpleAssignment = regexp.MustCompile(`^\s*([A-Za-z_]\w*(?:\.[A-Za-z_]\w*)*)\s*=\s*([^=])`)

// lowerConditionals handles whole return expressions, variable initializers and
// assignments to simple lvalues. Embedded expressions remain explicit review
// work until typed expression lowering can supply safe temporaries.
func lowerConditionals(body string) (string, error) {
	ts, err := Lex(body)
	if err != nil {
		return "", err
	}
	mask := Mask(ts)
	var edits []Edit
	labelEnds := statementLabelEnds(ts)
	start, depth := 0, 0
	for i, c := range mask {
		switch c {
		case ':':
			if depth == 0 && labelEnds[i+1] {
				start = i + 1
			}
		case '(', '[':
			depth++
		case ')', ']':
			depth--
		case '{', '}':
			if depth == 0 {
				start = i + 1
			}
		case ';':
			if depth != 0 {
				continue
			}
			segment := body[start:i]
			m := mask[start:i]
			// Preserve leading trivia separately, including comments.
			offset := 0
			st, _ := Lex(segment)
			for _, t := range st {
				if t.Kind == "space" || t.Kind == "comment" {
					offset = t.End
				} else {
					break
				}
			}
			s := segment[offset:]
			plain := m[offset:]
			prefix, expr, action := "", "", ""
			if strings.HasPrefix(plain, "return ") || strings.HasPrefix(plain, "return\n") || strings.HasPrefix(plain, "return\t") || strings.HasPrefix(plain, "return(") {
				expr = s[len("return"):]
				action = "return "
			} else if loc := initializer.FindStringSubmatchIndex(plain); loc != nil {
				prefix = s[loc[2]:loc[3]] + ";\n" + commentsIn(s[loc[3]:loc[1]])
				fields := strings.Fields(plain[loc[2]:loc[3]])
				action = fields[1] + " = "
				expr = s[loc[1]:]
			} else if loc := simpleAssignment.FindStringSubmatchIndex(plain); loc != nil {
				prefix = commentsIn(s[loc[3]:loc[4]])
				action = s[loc[2]:loc[3]] + " = "
				expr = s[loc[4]:]
			}
			if _, _, _, ok := conditional(expr); ok {
				position := start + offset
				lineStart := strings.LastIndex(body[:position], "\n") + 1
				j := lineStart
				for j < position && (body[j] == ' ' || body[j] == '\t') {
					j++
				}
				indent := body[lineStart:j]
				step := "    "
				if strings.Contains(indent, "\t") {
					step = "\t"
				}
				generated := emitConditional(expr, action, indent, step)
				// Original leading trivia already supplies the first line's
				// indentation. An initializer prefix needs subsequent lines
				// (including the generated if) indented independently.
				if prefix == "" {
					generated = strings.TrimPrefix(generated, indent)
				}
				generated = prefix + strings.TrimSuffix(generated, "\n")
				edits = append(edits, Edit{position, i + 1, generated})
			}
			start = i + 1
		}
	}
	return ApplyEdits(body, edits)
}

// Keep comments between a declaration/lvalue and its initializer when splitting
// one statement into a declaration followed by conditional assignments.
func commentsIn(s string) string {
	ts, _ := Lex(s)
	var b strings.Builder
	for _, t := range ts {
		if t.Kind == "comment" {
			b.WriteString(t.Text)
			b.WriteByte('\n')
		}
	}
	return b.String()
}
