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
	switch {
	case it.Bits > 128:
		return fmt.Sprintf("(ivy_uint<%d>(%s) & %s)", it.Bits, strconv.Quote(c.Name), bvMask(it.Bits)), true
	case it.Bits > 64:
		return fmt.Sprintf("(ivy_uint128_from_string(%s) & %s)", strconv.Quote(c.Name), bvMask(it.Bits)), true
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
		lhs = g.bvCastExpr(result, lhs)
		rhs = g.bvCastExpr(result, rhs)
		return g.maskBVExpr(fmt.Sprintf("((%s) << %d | (%s))", lhs, rhsWidth, rhs), result), true, nil
	case "bvand", "bvor", "bvxor", "+", "-", "*", "/", "%", "bvadd", "bvsub", "bvmul", "bvudiv", "bvurem":
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
		case "bvxor":
			op = "^"
		case "bvadd":
			op = "+"
		case "bvsub":
			op = "-"
		case "bvmul":
			op = "*"
		case "bvudiv":
			op = "/"
		case "bvurem":
			op = "%"
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
	case "bvneg":
		if len(a.Terms) != 1 {
			return "", true, fmt.Errorf("ivy2cpp: bvneg expected 1 argument, got %d", len(a.Terms))
		}
		body, err := g.emitExpr(a.Terms[0])
		if err != nil {
			return "", true, err
		}
		return g.maskBVExpr("(-("+body+"))", result), true, nil
	case "bvshl", "<<":
		return g.emitBVShiftApply(name, a, result, bvShiftLeft)
	case "bvlshr", ">>":
		return g.emitBVShiftApply(name, a, result, bvShiftLogicalRight)
	case "bvashr":
		return g.emitBVShiftApply(name, a, result, bvShiftArithmeticRight)
	case "cast":
		if len(a.Terms) != 1 {
			return "", true, fmt.Errorf("ivy2cpp: cast expected 1 argument, got %d", len(a.Terms))
		}
		body, err := g.emitExpr(a.Terms[0])
		if err != nil {
			return "", true, err
		}
		return g.maskBVExpr(g.bvCastExpr(result, body), result), true, nil
	default:
		if looksLikeBVOperator(name) {
			return "", true, fmt.Errorf("ivy2cpp: unknown BV operator %s/%d", name, len(a.Terms))
		}
		return "", false, nil
	}
}

type bvShiftKind int

const (
	bvShiftLeft bvShiftKind = iota
	bvShiftLogicalRight
	bvShiftArithmeticRight
)

func (g *Generator) emitBVShiftApply(name string, a *goivy.Apply, result cppInterpType, kind bvShiftKind) (string, bool, error) {
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
	raw := bvRawType(result)
	zero := bvZeroExpr(result)
	mask := bvMask(result.Bits)
	lhsInit := g.bvCastExpr(result, lhs)
	shiftAmount := g.bvShiftAmountExpr("__y", a.Terms[1].NodeSort())
	switch kind {
	case bvShiftLeft:
		return fmt.Sprintf("([&]() -> %s { %s __x = %s; auto __y = (%s); unsigned __s = %s; if (__s >= %d) return %s; return ((__x << __s) & %s); })()",
			raw, raw, lhsInit, rhs, shiftAmount, result.Bits, zero, mask), true, nil
	case bvShiftLogicalRight:
		return fmt.Sprintf("([&]() -> %s { %s __x = %s; auto __y = (%s); unsigned __s = %s; if (__s >= %d) return %s; return ((__x >> __s) & %s); })()",
			raw, raw, lhsInit, rhs, shiftAmount, result.Bits, zero, mask), true, nil
	case bvShiftArithmeticRight:
		if result.Bits <= 0 {
			return zero, true, nil
		}
		signMask := bvSignMask(result.Bits)
		return fmt.Sprintf("([&]() -> %s { %s __x = %s; auto __y = (%s); unsigned __s = %s; %s __mask = %s; %s __zero = %s; bool __sign = ((__x & %s) != __zero); if (__s >= %d) return __sign ? __mask : __zero; %s __res = ((__x >> __s) & __mask); if (__sign && __s != 0) __res = (__res | ((__mask << (%d - __s)) & __mask)); return (__res & __mask); })()",
			raw, raw, lhsInit, rhs, shiftAmount, raw, mask, raw, zero, signMask, result.Bits, raw, result.Bits), true, nil
	default:
		return "", true, fmt.Errorf("ivy2cpp: internal unknown BV shift kind %d", kind)
	}
}

func bvRawType(it cppInterpType) string {
	if typ := it.primitiveType(); typ != "" {
		return typ
	}
	return fmt.Sprintf("ivy_uint<%d>", it.Bits)
}

func (g *Generator) bvCastExpr(it cppInterpType, expr string) string {
	return fmt.Sprintf("static_cast<%s>(%s)", bvRawType(it), expr)
}

func bvZeroExpr(it cppInterpType) string {
	if it.Bits > 128 {
		return fmt.Sprintf("ivy_uint<%d>(0)", it.Bits)
	}
	return fmt.Sprintf("static_cast<%s>(0)", bvRawType(it))
}

func (g *Generator) bvShiftAmountExpr(value string, sort goivy.Sort) string {
	if it, ok := g.cppInterpType(sort); ok && it.Kind == cppInterpBV {
		switch {
		case it.Bits <= 32:
			return fmt.Sprintf("static_cast<unsigned>(%s)", value)
		case it.Bits <= 64:
			return fmt.Sprintf("((%s) > 4294967295ULL ? 4294967295U : static_cast<unsigned>(%s))", value, value)
		default:
			return fmt.Sprintf("ivy_bv_shift_amount(%s)", value)
		}
	}
	return fmt.Sprintf("static_cast<unsigned>(%s)", value)
}

func looksLikeBVOperator(name string) bool {
	if strings.HasPrefix(name, "bv") {
		return true
	}
	switch name {
	case "<<", ">>", "concat":
		return true
	default:
		return false
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
	if result, ok := g.cppInterpType(a.NodeSort()); ok && result.Kind == cppInterpBV {
		code := g.bvCastExpr(result, fmt.Sprintf("(%s >> %d)", body, shift))
		return g.maskBVExpr(code, result), true, nil
	}
	code := fmt.Sprintf("((%s >> %d) & %s)", body, shift, bvMask(width))
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

func bvSignMask(bits int) string {
	switch {
	case bits <= 0:
		return "0"
	case bits <= 32:
		return fmt.Sprintf("(1U << %d)", bits-1)
	case bits <= 64:
		return fmt.Sprintf("(1ULL << %d)", bits-1)
	case bits <= 128:
		return fmt.Sprintf("(((unsigned __int128)1) << %d)", bits-1)
	default:
		return fmt.Sprintf("(ivy_uint<%d>(1) << %d)", bits, bits-1)
	}
}
