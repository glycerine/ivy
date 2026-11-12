package io

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

type edifExpr struct {
	atom string
	list []*edifExpr
	line int
}

func (e *edifExpr) head() string {
	if e == nil || len(e.list) == 0 {
		return ""
	}
	return strings.ToLower(e.list[0].atom)
}
func (e *edifExpr) child(key string) *edifExpr {
	if e == nil {
		return nil
	}
	for _, child := range e.list {
		if child.head() == key {
			return child
		}
	}
	return nil
}
func (e *edifExpr) children(key string) []*edifExpr {
	var result []*edifExpr
	if e != nil {
		for _, child := range e.list {
			if child.head() == key {
				result = append(result, child)
			}
		}
	}
	return result
}
func edifIdentifier(e *edifExpr) (string, error) {
	if e == nil {
		return "", fmt.Errorf("missing EDIF identifier")
	}
	if e.list == nil {
		if e.atom == "" {
			return "", fmt.Errorf("empty EDIF identifier")
		}
		return e.atom, nil
	}
	if (e.head() == "rename" || e.head() == "array") && len(e.list) >= 2 {
		return edifIdentifier(e.list[1])
	}
	if e.head() == "member" && len(e.list) == 3 {
		name, err := edifIdentifier(e.list[1])
		if err != nil {
			return "", err
		}
		index, err := strconv.Atoi(e.list[2].atom)
		if err != nil || index < 0 {
			return "", fmt.Errorf("invalid EDIF member index")
		}
		return fmt.Sprintf("%s[%d]", name, index), nil
	}
	return "", fmt.Errorf("invalid EDIF identifier")
}
func parseEdifSyntax(name, text string) (*edifExpr, error) {
	position, line := 0, 1
	var parse func(int) (*edifExpr, error)
	space := func() {
		for position < len(text) && unicode.IsSpace(rune(text[position])) {
			if text[position] == '\n' {
				line++
			}
			position++
		}
	}
	parse = func(depth int) (*edifExpr, error) {
		space()
		if depth > 2048 {
			return nil, fmt.Errorf("%s:%d: EDIF nesting exceeds 2048", name, line)
		}
		if position >= len(text) {
			return nil, fmt.Errorf("%s:%d: truncated EDIF expression", name, line)
		}
		node := &edifExpr{line: line}
		if text[position] == '(' {
			position++
			node.list = []*edifExpr{}
			for {
				space()
				if position >= len(text) {
					return nil, fmt.Errorf("%s:%d: unclosed EDIF expression", name, node.line)
				}
				if text[position] == ')' {
					position++
					break
				}
				child, err := parse(depth + 1)
				if err != nil {
					return nil, err
				}
				node.list = append(node.list, child)
			}
			return node, nil
		}
		if text[position] == ')' {
			return nil, fmt.Errorf("%s:%d: unmatched EDIF parenthesis", name, line)
		}
		start := position
		if text[position] == '"' {
			position++
			for position < len(text) && text[position] != '"' {
				if text[position] == '\\' {
					position++
					if position >= len(text) {
						break
					}
				}
				if text[position] == '\n' {
					line++
				}
				position++
			}
			if position >= len(text) {
				return nil, fmt.Errorf("%s:%d: unclosed EDIF string", name, node.line)
			}
			position++
			value, err := strconv.Unquote(text[start:position])
			if err != nil {
				return nil, fmt.Errorf("%s:%d: %w", name, node.line, err)
			}
			node.atom = value
			return node, nil
		}
		for position < len(text) && !strings.ContainsRune("() \t\r\n", rune(text[position])) {
			position++
		}
		node.atom = text[start:position]
		return node, nil
	}
	root, err := parse(0)
	if err != nil {
		return nil, err
	}
	space()
	if position != len(text) {
		return nil, fmt.Errorf("%s:%d: trailing EDIF expression", name, line)
	}
	if root.head() != "edif" {
		return nil, fmt.Errorf("%s: expected edif root", name)
	}
	return root, nil
}
