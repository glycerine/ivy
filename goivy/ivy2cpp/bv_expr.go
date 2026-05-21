package ivy2cpp

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/glycerine/ivy/goivy"
)

func (g *Generator) emitBVNumeral(c *goivy.Const) (string, bool) {
	if c == nil || !goivy.IsNumeral(c) || goivy.IsLiteralString(c) {
		return "", false
	}
	it, ok := g.cppInterpType(c.CSort)
	if !ok || it.Kind != cppInterpBV {
		return "", false
	}
	return fmt.Sprintf("(%s & %s)", c.Name, bvMask(it.Bits)), true
}

func (g *Generator) emitBVApply(name string, a *goivy.Apply) (string, bool, error) {
	if a == nil || name == "" {
		return "", false, nil
	}
	if strings.HasPrefix(name, "bfe[") {
		return g.emitBFEApply(name, a)
	}
	result, ok := g.cppInterpType(a.NodeSort())
	if !ok || result.Kind != cppInterpBV {
		return "", false, nil
	}
	switch name {
	case "concat":
		if len(a.Terms) != 2 {
			return "", true, fmt.Errorf("ivy2cpp: concat expected 2 arguments, got %d", len(a.Terms))
		}
		lhs, err := g.emitExpr(a.Terms[0])
		if err != nil {
			return "", true, err
		}
		rhs, err := g.emitExpr(a.Terms[1])
		if err != nil {
			return "", true, err
		}
		rhsWidth, ok := g.bvWidthForSort(a.Terms[1].NodeSort())
		if !ok {
			return "", true, fmt.Errorf("ivy2cpp: concat rhs is not a bitvector: %s", a.Terms[1].String())
		}
		return fmt.Sprintf("((%s) << %d | (%s))", lhs, rhsWidth, rhs), true, nil
	case "bvand", "bvor", "+", "-", "*", "/", "%":
		if len(a.Terms) != 2 {
			return "", true, fmt.Errorf("ivy2cpp: %s expected 2 arguments, got %d", name, len(a.Terms))
		}
		lhs, err := g.emitExpr(a.Terms[0])
		if err != nil {
			return "", true, err
		}
		rhs, err := g.emitExpr(a.Terms[1])
		if err != nil {
			return "", true, err
		}
		op := name
		switch name {
		case "bvand":
			op = "&"
		case "bvor":
			op = "|"
		}
		return g.maskBVExpr(fmt.Sprintf("(%s %s %s)", lhs, op, rhs), result), true, nil
	case "bvnot":
		if len(a.Terms) != 1 {
			return "", true, fmt.Errorf("ivy2cpp: bvnot expected 1 argument, got %d", len(a.Terms))
		}
		body, err := g.emitExpr(a.Terms[0])
		if err != nil {
			return "", true, err
		}
		return g.maskBVExpr("(~"+body+")", result), true, nil
	case "cast":
		if len(a.Terms) != 1 {
			return "", true, fmt.Errorf("ivy2cpp: cast expected 1 argument, got %d", len(a.Terms))
		}
		body, err := g.emitExpr(a.Terms[0])
		if err != nil {
			return "", true, err
		}
		return g.maskBVExpr(body, result), true, nil
	default:
		return "", false, nil
	}
}

func (g *Generator) emitBFEApply(name string, a *goivy.Apply) (string, bool, error) {
	lo, hi, ok := parseBFEParams(name)
	if !ok {
		return "", true, fmt.Errorf("ivy2cpp: malformed bit extraction operator %q", name)
	}
	if len(a.Terms) != 1 {
		return "", true, fmt.Errorf("ivy2cpp: %s expected 1 argument, got %d", name, len(a.Terms))
	}
	body, err := g.emitExpr(a.Terms[0])
	if err != nil {
		return "", true, err
	}
	shift := lo
	top := hi
	if top < shift {
		shift, top = top, shift
	}
	width := top - shift + 1
	code := fmt.Sprintf("((%s >> %d) & %s)", body, shift, bvMask(width))
	if result, ok := g.cppInterpType(a.NodeSort()); ok && result.Kind == cppInterpBV {
		return g.maskBVExpr(code, result), true, nil
	}
	return code, true, nil
}

func parseBFEParams(name string) (int, int, bool) {
	if !strings.HasPrefix(name, "bfe[") {
		return 0, 0, false
	}
	_, params, ok := parseBracketInts(name)
	if ok && len(params) == 2 {
		return params[0], params[1], true
	}
	var buf strings.Builder
	for _, r := range strings.TrimPrefix(name, "bfe") {
		switch r {
		case '[', ']', ',', ':':
			buf.WriteByte(' ')
		default:
			buf.WriteRune(r)
		}
	}
	fields := strings.Fields(buf.String())
	if len(fields) != 2 {
		return 0, 0, false
	}
	lo, err := strconv.Atoi(fields[0])
	if err != nil {
		return 0, 0, false
	}
	hi, err := strconv.Atoi(fields[1])
	if err != nil {
		return 0, 0, false
	}
	return lo, hi, true
}

func (g *Generator) bvWidthForSort(s goivy.Sort) (int, bool) {
	it, ok := g.cppInterpType(s)
	if !ok || it.Kind != cppInterpBV {
		return 0, false
	}
	return it.Bits, true
}

func (g *Generator) maskBVExpr(expr string, it cppInterpType) string {
	return fmt.Sprintf("((%s) & %s)", expr, bvMask(it.Bits))
}
