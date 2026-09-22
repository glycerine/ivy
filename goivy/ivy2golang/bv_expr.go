package ivy2golang

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/glycerine/ivy/goivy"
)

type goInterpKind string

const (
	goInterpBV    goInterpKind = "bv"
	goInterpStrBV goInterpKind = "strbv"
	goInterpIntBV goInterpKind = "intbv"
)

type goInterpType struct {
	Kind goInterpKind
	Bits int
	Lo   int
	Hi   int
}

func parseGoInterpType(text string) (goInterpType, bool) {
	base, params, ok := parseBracketInts(strings.TrimSpace(text))
	if !ok {
		return goInterpType{}, false
	}
	switch base {
	case "bv":
		if len(params) != 1 {
			return goInterpType{}, false
		}
		return goInterpType{Kind: goInterpBV, Bits: params[0]}, true
	case "strbv":
		if len(params) != 1 {
			return goInterpType{}, false
		}
		return goInterpType{Kind: goInterpStrBV, Bits: params[0]}, true
	case "intbv":
		if len(params) != 3 {
			return goInterpType{}, false
		}
		return goInterpType{Kind: goInterpIntBV, Lo: params[0], Hi: params[1], Bits: params[2]}, true
	default:
		return goInterpType{}, false
	}
}

func parseBracketInts(text string) (string, []int, bool) {
	idx := strings.IndexByte(text, '[')
	if idx < 0 {
		return text, nil, true
	}
	base := text[:idx]
	rest := text[idx:]
	var params []int
	for rest != "" {
		if rest[0] != '[' {
			return "", nil, false
		}
		end := strings.IndexByte(rest, ']')
		if end < 0 {
			return "", nil, false
		}
		value, err := strconv.ParseInt(rest[1:end], 0, 0)
		if err != nil {
			return "", nil, false
		}
		params = append(params, int(value))
		rest = rest[end+1:]
	}
	return base, params, true
}

func (g *Generator) goInterpType(s goivy.Sort) (goInterpType, bool) {
	text, ok := g.sortInterpString(s)
	if !ok {
		return goInterpType{}, false
	}
	return parseGoInterpType(text)
}

func (it goInterpType) card() int {
	switch it.Kind {
	case goInterpBV, goInterpStrBV:
		if it.Bits < 0 || it.Bits >= strconv.IntSize {
			return -1
		}
		return 1 << it.Bits
	case goInterpIntBV:
		if it.Hi < it.Lo {
			return -1
		}
		return it.Hi - it.Lo + 1
	default:
		return -1
	}
}

func (g *Generator) emitBVNumeral(c *goivy.Const) (string, bool, error) {
	if c == nil || !goivy.IsNumeral(c) || goivy.IsLiteralString(c) {
		return "", false, nil
	}
	it, ok := g.goInterpType(c.CSort)
	if !ok || it.Kind != goInterpBV {
		return "", false, nil
	}
	value, err := strconv.ParseInt(c.Name, 0, 0)
	if err != nil {
		return "", true, fmt.Errorf("ivy2golang: cannot parse bitvector numeral %q", c.Name)
	}
	mask, err := goBVMaskValue(it.Bits)
	if err != nil {
		return "", true, err
	}
	return strconv.FormatInt(value&mask, 10), true, nil
}

func (g *Generator) emitBVApply(name string, a *goivy.Apply) (string, bool, error) {
	if a == nil || name == "" {
		return "", false, nil
	}
	if strings.HasPrefix(name, "bfe[") {
		return g.emitBFEApply(name, a)
	}
	if isNamedBVOperator(name) && g.signatureHasSymbol(name) {
		return "", false, nil
	}
	result, ok := g.goInterpType(a.NodeSort())
	if !ok || result.Kind != goInterpBV {
		if looksLikeBVOperator(name) {
			return "", true, fmt.Errorf("ivy2golang: unknown BV operator %s/%d", name, len(a.Terms))
		}
		return "", false, nil
	}
	switch name {
	case "concat":
		if len(a.Terms) != 2 {
			return "", true, fmt.Errorf("ivy2golang: concat expected 2 arguments, got %d", len(a.Terms))
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
			return "", true, fmt.Errorf("ivy2golang: concat rhs is not a bitvector: %s", a.Terms[1].String())
		}
		code, err := g.maskBVExpr(fmt.Sprintf("((%s) << %d | (%s))", lhs, rhsWidth, rhs), result)
		return code, true, err
	case "bvand", "bvor", "bvxor", "+", "-", "*", "/", "%", "bvadd", "bvsub", "bvmul", "bvudiv", "bvurem":
		if len(a.Terms) != 2 {
			return "", true, fmt.Errorf("ivy2golang: %s expected 2 arguments, got %d", name, len(a.Terms))
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
		code, err := g.maskBVExpr(fmt.Sprintf("(%s %s %s)", lhs, op, rhs), result)
		return code, true, err
	case "bvnot":
		if len(a.Terms) != 1 {
			return "", true, fmt.Errorf("ivy2golang: bvnot expected 1 argument, got %d", len(a.Terms))
		}
		body, err := g.emitExpr(a.Terms[0])
		if err != nil {
			return "", true, err
		}
		code, err := g.maskBVExpr("(^"+body+")", result)
		return code, true, err
	case "bvneg":
		if len(a.Terms) != 1 {
			return "", true, fmt.Errorf("ivy2golang: bvneg expected 1 argument, got %d", len(a.Terms))
		}
		body, err := g.emitExpr(a.Terms[0])
		if err != nil {
			return "", true, err
		}
		code, err := g.maskBVExpr("(-("+body+"))", result)
		return code, true, err
	case "bvshl", "<<":
		return g.emitBVShiftApply(name, a, result, bvShiftLeft)
	case "bvlshr", ">>":
		return g.emitBVShiftApply(name, a, result, bvShiftLogicalRight)
	case "bvashr":
		return g.emitBVShiftApply(name, a, result, bvShiftArithmeticRight)
	case "cast":
		if len(a.Terms) != 1 {
			return "", true, fmt.Errorf("ivy2golang: cast expected 1 argument, got %d", len(a.Terms))
		}
		body, err := g.emitExpr(a.Terms[0])
		if err != nil {
			return "", true, err
		}
		code, err := g.maskBVExpr(body, result)
		return code, true, err
	default:
		if looksLikeBVOperator(name) {
			return "", true, fmt.Errorf("ivy2golang: unknown BV operator %s/%d", name, len(a.Terms))
		}
		return "", false, nil
	}
}

func (g *Generator) signatureHasSymbol(name string) bool {
	if g == nil || g.Mod == nil || g.Mod.Sig == nil || name == "" {
		return false
	}
	_, ok := g.Mod.Sig.Symbols.Get2(name)
	return ok
}

func isNamedBVOperator(name string) bool {
	switch name {
	case "bvand", "bvor", "bvxor", "bvnot", "bvneg", "bvshl", "bvlshr", "bvashr", "bvadd", "bvsub", "bvmul", "bvudiv", "bvurem":
		return true
	default:
		return false
	}
}

type bvShiftKind int

const (
	bvShiftLeft bvShiftKind = iota
	bvShiftLogicalRight
	bvShiftArithmeticRight
)

func (g *Generator) emitBVShiftApply(name string, a *goivy.Apply, result goInterpType, kind bvShiftKind) (string, bool, error) {
	if len(a.Terms) != 2 {
		return "", true, fmt.Errorf("ivy2golang: %s expected 2 arguments, got %d", name, len(a.Terms))
	}
	lhs, err := g.emitExpr(a.Terms[0])
	if err != nil {
		return "", true, err
	}
	rhs, err := g.emitExpr(a.Terms[1])
	if err != nil {
		return "", true, err
	}
	mask, err := goBVMaskLiteral(result.Bits)
	if err != nil {
		return "", true, err
	}
	signMask, err := goBVSignMaskLiteral(result.Bits)
	if err != nil {
		return "", true, err
	}
	switch kind {
	case bvShiftLeft:
		return fmt.Sprintf("func() int {\n__x := (%s) & %s\n__s := ivyBVShiftAmount(%s)\nif __s >= %d {\nreturn 0\n}\nreturn ((__x << __s) & %s)\n}()", lhs, mask, rhs, result.Bits, mask), true, nil
	case bvShiftLogicalRight:
		return fmt.Sprintf("func() int {\n__x := (%s) & %s\n__s := ivyBVShiftAmount(%s)\nif __s >= %d {\nreturn 0\n}\nreturn ((__x >> __s) & %s)\n}()", lhs, mask, rhs, result.Bits, mask), true, nil
	case bvShiftArithmeticRight:
		return fmt.Sprintf("func() int {\n__x := (%s) & %s\n__s := ivyBVShiftAmount(%s)\n__sign := (__x & %s) != 0\nif __s >= %d {\nif __sign {\nreturn %s\n}\nreturn 0\n}\n__res := ((__x >> __s) & %s)\nif __sign && __s != 0 {\n__res = (__res | ((%s << (%d - __s)) & %s))\n}\nreturn (__res & %s)\n}()", lhs, mask, rhs, signMask, result.Bits, mask, mask, mask, result.Bits, mask, mask), true, nil
	default:
		return "", true, fmt.Errorf("ivy2golang: internal unknown BV shift kind %d", kind)
	}
}

func looksLikeBVOperator(name string) bool {
	if strings.HasPrefix(name, "bv") || strings.HasPrefix(name, "bfe[") {
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
		return "", true, fmt.Errorf("ivy2golang: malformed bit extraction operator %q", name)
	}
	if len(a.Terms) != 1 {
		return "", true, fmt.Errorf("ivy2golang: %s expected 1 argument, got %d", name, len(a.Terms))
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
	if result, ok := g.goInterpType(a.NodeSort()); ok && result.Kind == goInterpBV {
		code, err := g.maskBVExpr(fmt.Sprintf("(%s >> %d)", body, shift), result)
		return code, true, err
	}
	mask, err := goBVMaskLiteral(width)
	if err != nil {
		return "", true, err
	}
	return fmt.Sprintf("((%s >> %d) & %s)", body, shift, mask), true, nil
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
	it, ok := g.goInterpType(s)
	if !ok || it.Kind != goInterpBV {
		return 0, false
	}
	return it.Bits, true
}

func (g *Generator) maskBVExpr(expr string, it goInterpType) (string, error) {
	mask, err := goBVMaskLiteral(it.Bits)
	if err != nil {
		return "", err
	}
	if inner, ok := stripOuterParens(expr); ok {
		expr = inner
	}
	return fmt.Sprintf("((%s) & %s)", expr, mask), nil
}

func goBVMaskValue(bits int) (int64, error) {
	if bits <= 0 {
		return 0, nil
	}
	if bits >= strconv.IntSize {
		return 0, fmt.Errorf("ivy2golang: bitvector width %d exceeds Go int-backed support", bits)
	}
	return int64((1 << bits) - 1), nil
}

func goBVMaskLiteral(bits int) (string, error) {
	mask, err := goBVMaskValue(bits)
	if err != nil {
		return "", err
	}
	return strconv.FormatInt(mask, 10), nil
}

func goBVSignMaskLiteral(bits int) (string, error) {
	if bits <= 0 {
		return "0", nil
	}
	if bits >= strconv.IntSize {
		return "", fmt.Errorf("ivy2golang: bitvector width %d exceeds Go int-backed support", bits)
	}
	return strconv.FormatInt(int64(1<<(bits-1)), 10), nil
}

func stripOuterParens(expr string) (string, bool) {
	if len(expr) < 2 || expr[0] != '(' || expr[len(expr)-1] != ')' {
		return expr, false
	}
	depth := 0
	for i, r := range expr {
		switch r {
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 && i != len(expr)-1 {
				return expr, false
			}
		}
		if depth < 0 {
			return expr, false
		}
	}
	return expr[1 : len(expr)-1], depth == 0
}
