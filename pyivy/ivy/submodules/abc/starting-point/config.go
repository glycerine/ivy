package startingpoint

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// ParseConfig reads explicit platform choices. Zero means disabled, including
// for ifdef/defined; other integers mean enabled. Unlisted symbols stay unknown.
func ParseConfig(text string) (map[string]int64, error) {
	values := map[string]int64{}
	for i, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(strings.SplitN(line, "//", 2)[0])
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			continue
		}
		parts := strings.Split(line, "=")
		if len(parts) != 2 || !identifier.MatchString(strings.TrimSpace(parts[0])) {
			return nil, fmt.Errorf("line %d: expected NAME = INTEGER", i+1)
		}
		name := strings.TrimSpace(parts[0])
		value, err := strconv.ParseInt(strings.TrimSpace(parts[1]), 0, 64)
		if err != nil {
			return nil, fmt.Errorf("line %d: invalid value for %s: %w", i+1, name, err)
		}
		if _, exists := values[name]; exists {
			return nil, fmt.Errorf("line %d: duplicate setting %s", i+1, name)
		}
		values[name] = value
	}
	return values, nil
}

func ReadConfig(path string) (map[string]int64, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	values, err := ParseConfig(string(data))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return values, nil
}

var definedWithoutParens = regexp.MustCompile(`\bdefined\s+([A-Za-z_]\w*)`)

func configCondition(d Token, values map[string]int64) (bool, bool) {
	parts := directiveParts.FindStringSubmatch(d.Text)
	if len(parts) < 2 {
		return false, false
	}
	name := parts[1]
	expr := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(d.Text), "#"))
	expr = strings.TrimSpace(strings.TrimPrefix(expr, name))
	expr = strings.ReplaceAll(strings.ReplaceAll(expr, "\\\r\n", ""), "\\\n", "")
	ts, err := Lex(expr)
	if err != nil {
		return false, false
	}
	var clean strings.Builder
	for _, t := range ts {
		if t.Kind == "comment" {
			clean.WriteByte(' ')
		} else {
			clean.WriteString(t.Text)
		}
	}
	expr = strings.TrimSpace(clean.String())
	if name == "ifdef" || name == "ifndef" {
		value, ok := values[expr]
		if !ok {
			return false, false
		}
		if name == "ifndef" {
			return value == 0, true
		}
		return value != 0, true
	}
	expr = definedWithoutParens.ReplaceAllString(expr, "defined($1)")
	node, err := parser.ParseExpr(expr)
	if err != nil {
		return false, false
	}
	value, ok := evalConfigExpr(node, values)
	return value != 0, ok
}
func boolInt(b bool) int64 {
	if b {
		return 1
	}
	return 0
}
func evalConfigExpr(expr ast.Expr, values map[string]int64) (int64, bool) {
	switch n := expr.(type) {
	case *ast.ParenExpr:
		return evalConfigExpr(n.X, values)
	case *ast.Ident:
		v, ok := values[n.Name]
		return v, ok
	case *ast.BasicLit:
		if n.Kind != token.INT {
			return 0, false
		}
		v, err := strconv.ParseInt(n.Value, 0, 64)
		return v, err == nil
	case *ast.CallExpr:
		f, ok := n.Fun.(*ast.Ident)
		if !ok || f.Name != "defined" || len(n.Args) != 1 {
			return 0, false
		}
		id, ok := n.Args[0].(*ast.Ident)
		if !ok {
			return 0, false
		}
		v, ok := values[id.Name]
		return boolInt(v != 0), ok
	case *ast.UnaryExpr:
		v, ok := evalConfigExpr(n.X, values)
		if !ok {
			return 0, false
		}
		switch n.Op {
		case token.NOT:
			return boolInt(v == 0), true
		case token.ADD:
			return v, true
		case token.SUB:
			return -v, true
		}
	case *ast.BinaryExpr:
		a, ak := evalConfigExpr(n.X, values)
		b, bk := evalConfigExpr(n.Y, values)
		if n.Op == token.LAND && ((ak && a == 0) || (bk && b == 0)) {
			return 0, true
		}
		if n.Op == token.LOR && ((ak && a != 0) || (bk && b != 0)) {
			return 1, true
		}
		if !ak || !bk {
			return 0, false
		}
		switch n.Op {
		case token.LAND:
			return boolInt(a != 0 && b != 0), true
		case token.LOR:
			return boolInt(a != 0 || b != 0), true
		case token.EQL:
			return boolInt(a == b), true
		case token.NEQ:
			return boolInt(a != b), true
		case token.LSS:
			return boolInt(a < b), true
		case token.LEQ:
			return boolInt(a <= b), true
		case token.GTR:
			return boolInt(a > b), true
		case token.GEQ:
			return boolInt(a >= b), true
		}
	}
	return 0, false
}

// selectConfiguredBranches blanks removed bytes rather than deleting them, so
// original line/column/byte coordinates remain valid throughout analysis.
func selectConfiguredBranches(source string, values map[string]int64) (string, error) {
	ts, err := Lex(source)
	if err != nil {
		return "", err
	}
	guards := map[int]bool{}
	for _, edit := range headerGuardEdits(ts) {
		guards[edit.Start] = true
	}
	var directives []Token
	for _, t := range ts {
		if t.Kind == "directive" && !guards[t.Start] {
			directives = append(directives, t)
		}
	}
	result := []byte(source)
	blank := func(start, end int) {
		for i := start; i < end; i++ {
			if result[i] != '\n' && result[i] != '\r' {
				result[i] = ' '
			}
		}
	}
	local := newLocalPreprocessor()
	var walk func(int, int) error
	walk = func(lo, hi int) error {
		for i := lo; i < hi; i++ {
			kind := directiveName(directives[i])
			if kind != "if" && kind != "ifdef" && kind != "ifndef" {
				local.observe(directives[i])
				continue
			}
			depth, end := 1, -1
			branches := []int{i}
			for j := i + 1; j < hi; j++ {
				switch directiveName(directives[j]) {
				case "if", "ifdef", "ifndef":
					depth++
				case "endif":
					depth--
					if depth == 0 {
						end = j
					}
				case "elif", "elseif", "else":
					if depth == 1 {
						branches = append(branches, j)
					}
				}
				if end >= 0 {
					break
				}
			}
			if end < 0 {
				// With no explicit choices, preserve incomplete scaffolding for
				// the existing opaque-region recovery instead of rejecting it here.
				if len(values) == 0 {
					return nil
				}
				return &syntaxError{"unclosed conditional", directives[i].Start}
			}
			selected, known := -1, true
			for b, at := range branches {
				if directiveName(directives[at]) == "else" {
					selected = b
					break
				}
				take, ok := local.condition(directives[at], values)
				if !ok {
					known = false
					break
				}
				if take {
					selected = b
					break
				}
			}
			if known {
				for b, at := range branches {
					next := end
					if b+1 < len(branches) {
						next = branches[b+1]
					}
					blank(directives[at].Start, directives[at].End)
					if b != selected {
						blank(directives[at].End, directives[next].Start)
					} else if err := walk(at+1, next); err != nil {
						return err
					}
				}
				blank(directives[end].Start, directives[end].End)
			}
			if !known {
				// A branch may redefine or undefine a previously known symbol.
				for j := i + 1; j < end; j++ {
					local.forget(directives[j])
				}
			}
			// Unknown enclosing choices remain intact, including their nested branches.
			i = end
		}
		return nil
	}
	if err := walk(0, len(directives)); err != nil {
		return "", err
	}
	return string(result), nil
}
