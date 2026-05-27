package ivy2go

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/glycerine/ivy/goivy"
)

// Mirrors ivy2cpp/bv_expr.go. Lowers Ivy bitvector operations into Go
// expressions. Width-banded lowering (per ARCHITECTURE_TODO.md §3.5.7):
//
//   - bits ≤ 32  → uint32 arithmetic, mask with 1<<bits - 1
//   - bits ≤ 64  → uint64 arithmetic, mask with 1<<bits - 1
//   - bits > 64 → *big.Int via big.Int operations and masks

// emitBVNumeral mirrors ivy2cpp/bv_expr.go emitBVNumeral.
func (g *Generator) emitBVNumeral(c *goivy.Const) (string, bool) {
	if c == nil || !goivy.IsNumeral(c) || goivy.IsLiteralString(c) {
		return "", false
	}
	it, ok := g.goInterpType(c.CSort)
	if !ok || it.Kind != goInterpBV {
		return "", false
	}
	switch {
	case it.Bits > 64:
		return g.wideBVLiteralExpr(c.Name, it.Bits), true
	}
	value := c.Name
	mask := bvMask(it.Bits)
	primitive := it.primitiveType()
	if primitive == "" {
		primitive = "uint64"
	}
	return fmt.Sprintf("(%s(%s) & %s)", primitive, value, mask), true
}

func (g *Generator) wideBVLiteralExpr(value string, bits int) string {
	g.requireBigInt()
	return fmt.Sprintf("func() *big.Int { v, ok := new(big.Int).SetString(%s, 0); if !ok { v = new(big.Int) }; return v.And(v, bigIntMask(%d)) }()",
		strconv.Quote(value), bits)
}

// emitBVApply mirrors ivy2cpp/bv_expr.go emitBVApply. Returns
// (code, handled, err) — handled=false means the operator isn't a BV
// operator and emitApply should keep dispatching.
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
		return "", false, nil
	}
	primitive := result.primitiveType()
	if primitive == "" || result.Bits > 64 {
		// Wide BV (>64 bits) lowering via *big.Int. Mirrors
		// ivy2cpp's ivy_uint<N> template path but uses math/big
		// for arithmetic; mask after every op.
		return g.emitWideBVApply(name, a, result)
	}
	switch name {
	case "concat":
		if len(a.Terms) != 2 {
			return "", true, fmt.Errorf("ivy2go: concat expected 2 arguments, got %d", len(a.Terms))
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
			return "", true, fmt.Errorf("ivy2go: concat rhs is not a bitvector: %s", a.Terms[1].String())
		}
		expr := fmt.Sprintf("((%s(%s) << %d) | %s(%s))", primitive, lhs, rhsWidth, primitive, rhs)
		return g.maskBVExpr(expr, result), true, nil
	case "bvand", "bvor", "bvxor", "+", "-", "*", "/", "%", "bvadd", "bvsub", "bvmul", "bvudiv", "bvurem":
		if len(a.Terms) != 2 {
			return "", true, fmt.Errorf("ivy2go: %s expected 2 arguments, got %d", name, len(a.Terms))
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
		expr := fmt.Sprintf("(%s(%s) %s %s(%s))", primitive, lhs, op, primitive, rhs)
		return g.maskBVExpr(expr, result), true, nil
	case "bvnot":
		if len(a.Terms) != 1 {
			return "", true, fmt.Errorf("ivy2go: bvnot expected 1 argument, got %d", len(a.Terms))
		}
		body, err := g.emitExpr(a.Terms[0])
		if err != nil {
			return "", true, err
		}
		return g.maskBVExpr(fmt.Sprintf("(^%s(%s))", primitive, body), result), true, nil
	case "bvneg":
		if len(a.Terms) != 1 {
			return "", true, fmt.Errorf("ivy2go: bvneg expected 1 argument, got %d", len(a.Terms))
		}
		body, err := g.emitExpr(a.Terms[0])
		if err != nil {
			return "", true, err
		}
		// Go has no `-` on unsigned; do `(^x + 1)` modulo mask.
		return g.maskBVExpr(fmt.Sprintf("(^%s(%s) + 1)", primitive, body), result), true, nil
	case "bvshl", "<<":
		return g.emitBVShiftApply(name, a, result, bvShiftLeft)
	case "bvlshr", ">>":
		return g.emitBVShiftApply(name, a, result, bvShiftLogicalRight)
	case "bvashr":
		return g.emitBVShiftApply(name, a, result, bvShiftArithmeticRight)
	case "cast":
		if len(a.Terms) != 1 {
			return "", true, fmt.Errorf("ivy2go: cast expected 1 argument, got %d", len(a.Terms))
		}
		body, err := g.emitExpr(a.Terms[0])
		if err != nil {
			return "", true, err
		}
		return g.maskBVExpr(fmt.Sprintf("%s(%s)", primitive, body), result), true, nil
	default:
		if looksLikeBVOperator(name) {
			return "", true, fmt.Errorf("ivy2go: unknown BV operator %s/%d", name, len(a.Terms))
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

// emitBVShiftApply mirrors ivy2cpp/bv_expr.go emitBVShiftApply, lowering
// Ivy BV shift operators with the same out-of-range saturation rules.
func (g *Generator) emitBVShiftApply(name string, a *goivy.Apply, result goInterpType, kind bvShiftKind) (string, bool, error) {
	if len(a.Terms) != 2 {
		return "", true, fmt.Errorf("ivy2go: %s expected 2 arguments, got %d", name, len(a.Terms))
	}
	primitive := result.primitiveType()
	if primitive == "" || result.Bits > 64 {
		return "", true, fmt.Errorf("ivy2go: BV shift %q on bv[%d] needs wide-BV lowering (deferred)", name, result.Bits)
	}
	lhs, err := g.emitExpr(a.Terms[0])
	if err != nil {
		return "", true, err
	}
	rhs, err := g.emitExpr(a.Terms[1])
	if err != nil {
		return "", true, err
	}
	mask := bvMask(result.Bits)
	switch kind {
	case bvShiftLeft:
		return fmt.Sprintf("func() %s { x := %s(%s); y := uint(%s); if y >= %d { return 0 }; return (x << y) & %s }()",
			primitive, primitive, lhs, rhs, result.Bits, mask), true, nil
	case bvShiftLogicalRight:
		return fmt.Sprintf("func() %s { x := %s(%s); y := uint(%s); if y >= %d { return 0 }; return (x >> y) & %s }()",
			primitive, primitive, lhs, rhs, result.Bits, mask), true, nil
	case bvShiftArithmeticRight:
		signMask := bvSignMask(result.Bits)
		return fmt.Sprintf(`func() %s {
	x := %s(%s); y := uint(%s); mask := %s(%s); signBit := %s(%s)
	sign := (x & signBit) != 0
	if y >= %d { if sign { return mask } else { return 0 } }
	res := (x >> y) & mask
	if sign && y != 0 {
		res |= (mask << (%d - y)) & mask
	}
	return res & mask
}()`, primitive, primitive, lhs, rhs, primitive, mask, primitive, signMask, result.Bits, result.Bits), true, nil
	default:
		return "", true, fmt.Errorf("ivy2go: internal unknown BV shift kind %d", kind)
	}
}

// emitBFEApply mirrors ivy2cpp/bv_expr.go emitBFEApply: `bfe[lo,hi]`
// extracts a bit range from a BV.
func (g *Generator) emitBFEApply(name string, a *goivy.Apply) (string, bool, error) {
	lo, hi, ok := parseBFEParams(name)
	if !ok {
		return "", true, fmt.Errorf("ivy2go: malformed bit extraction operator %q", name)
	}
	if len(a.Terms) != 1 {
		return "", true, fmt.Errorf("ivy2go: %s expected 1 argument, got %d", name, len(a.Terms))
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
		primitive := result.primitiveType()
		if primitive == "" || result.Bits > 64 {
			return "", true, fmt.Errorf("ivy2go: bfe target bv[%d] needs wide-BV lowering (deferred)", result.Bits)
		}
		code := fmt.Sprintf("(%s(%s) >> %d)", primitive, body, shift)
		return g.maskBVExpr(code, result), true, nil
	}
	return fmt.Sprintf("((%s >> %d) & %s)", body, shift, bvMask(width)), true, nil
}

// parseBFEParams mirrors ivy2cpp/bv_expr.go parseBFEParams.
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

// bvWidthForSort mirrors ivy2cpp/bv_expr.go bvWidthForSort.
func (g *Generator) bvWidthForSort(s goivy.Sort) (int, bool) {
	it, ok := g.goInterpType(s)
	if !ok || it.Kind != goInterpBV {
		return 0, false
	}
	return it.Bits, true
}

// maskBVExpr mirrors ivy2cpp/bv_expr.go maskBVExpr.
func (g *Generator) maskBVExpr(expr string, it goInterpType) string {
	return fmt.Sprintf("((%s) & %s)", expr, bvMask(it.Bits))
}

// bvMask returns the all-ones mask for `bits`. Mirrors ivy2cpp's same
// helper with Go literal syntax.
//   - bits ≤ 64 → untyped integer literal
//   - bits > 64 → "bigIntMask(bits)" runtime call
func bvMask(bits int) string {
	switch {
	case bits <= 0:
		return "0"
	case bits == 64:
		return "0xFFFFFFFFFFFFFFFF"
	case bits > 64:
		return fmt.Sprintf("bigIntMask(%d)", bits)
	}
	return strconv.FormatUint(uint64(1)<<uint(bits)-1, 10)
}

// bvSignMask returns the high-bit mask for `bits`. Mirrors ivy2cpp.
func bvSignMask(bits int) string {
	switch {
	case bits <= 0:
		return "0"
	case bits <= 64:
		return fmt.Sprintf("(uint64(1) << %d)", bits-1)
	default:
		return fmt.Sprintf("bigIntSignMask(%d)", bits)
	}
}

// signatureHasSymbol mirrors ivy2cpp/bv_expr.go signatureHasSymbol.
func (g *Generator) signatureHasSymbol(name string) bool {
	if g == nil || g.Mod == nil || g.Mod.Sig == nil || name == "" {
		return false
	}
	_, ok := g.Mod.Sig.Symbols.Get2(name)
	return ok
}

// isNamedBVOperator mirrors ivy2cpp/bv_expr.go isNamedBVOperator.
func isNamedBVOperator(name string) bool {
	switch name {
	case "bvand", "bvor", "bvxor", "bvnot", "bvneg",
		"bvshl", "bvlshr", "bvashr",
		"bvadd", "bvsub", "bvmul", "bvudiv", "bvurem":
		return true
	}
	return false
}

// looksLikeBVOperator mirrors ivy2cpp/bv_expr.go looksLikeBVOperator.
func looksLikeBVOperator(name string) bool {
	if strings.HasPrefix(name, "bv") {
		return true
	}
	switch name {
	case "<<", ">>", "concat":
		return true
	}
	return false
}

// requireBigInt records that the emitted runtime needs math/big.
// The per-stream math/big imports are added lazily by finalize when it
// detects "big.Int" in a stream's body.
func (g *Generator) requireBigInt() {
	if g == nil || g.Ctx == nil {
		return
	}
	g.Ctx.OnceGlobals["__need_bigint"] = true
}

// --- OPEN 054: wide BV operator lowering via *big.Int ----------------

// emitWideBVApply lowers BV operators whose result sort has bits > 64
// by routing through math/big. Mirrors ivy2cpp's ivy_uint<N> template
// path. Each operand is wrapped in a big.Int via the wideBVAs helper
// (emitted into the runtime), the operation is applied, and the
// result is masked.
func (g *Generator) emitWideBVApply(name string, a *goivy.Apply, result goInterpType) (string, bool, error) {
	g.requireBigInt()
	switch name {
	case "bvand", "bvor", "bvxor",
		"bvadd", "bvsub", "bvmul", "bvudiv", "bvurem",
		"+", "-", "*", "/", "%":
		if len(a.Terms) != 2 {
			return "", true, fmt.Errorf("ivy2go: %s expected 2 arguments, got %d", name, len(a.Terms))
		}
		lhs, err := g.emitExpr(a.Terms[0])
		if err != nil {
			return "", true, err
		}
		rhs, err := g.emitExpr(a.Terms[1])
		if err != nil {
			return "", true, err
		}
		op := wideBVMethodFor(name)
		return fmt.Sprintf("wideBV%s(%s, %s, %d)", op, wideBVAsCall(lhs), wideBVAsCall(rhs), result.Bits), true, nil
	case "bvnot":
		if len(a.Terms) != 1 {
			return "", true, fmt.Errorf("ivy2go: bvnot expected 1 argument, got %d", len(a.Terms))
		}
		body, err := g.emitExpr(a.Terms[0])
		if err != nil {
			return "", true, err
		}
		return fmt.Sprintf("wideBVNot(%s, %d)", wideBVAsCall(body), result.Bits), true, nil
	case "bvneg":
		if len(a.Terms) != 1 {
			return "", true, fmt.Errorf("ivy2go: bvneg expected 1 argument, got %d", len(a.Terms))
		}
		body, err := g.emitExpr(a.Terms[0])
		if err != nil {
			return "", true, err
		}
		return fmt.Sprintf("wideBVNeg(%s, %d)", wideBVAsCall(body), result.Bits), true, nil
	case "bvshl", "<<":
		return g.emitWideBVShift(name, a, result, "Shl")
	case "bvlshr", ">>":
		return g.emitWideBVShift(name, a, result, "ShrLogical")
	case "bvashr":
		return g.emitWideBVShift(name, a, result, "ShrArith")
	case "cast":
		if len(a.Terms) != 1 {
			return "", true, fmt.Errorf("ivy2go: cast expected 1 argument, got %d", len(a.Terms))
		}
		body, err := g.emitExpr(a.Terms[0])
		if err != nil {
			return "", true, err
		}
		return fmt.Sprintf("wideBVMask(%s, %d)", wideBVAsCall(body), result.Bits), true, nil
	case "concat":
		if len(a.Terms) != 2 {
			return "", true, fmt.Errorf("ivy2go: concat expected 2 arguments, got %d", len(a.Terms))
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
			return "", true, fmt.Errorf("ivy2go: concat rhs is not a bitvector: %s", a.Terms[1].String())
		}
		return fmt.Sprintf("wideBVConcat(%s, %s, %d, %d)", wideBVAsCall(lhs), wideBVAsCall(rhs), rhsWidth, result.Bits), true, nil
	default:
		return "", true, fmt.Errorf("ivy2go: unsupported wide BV operator %q", name)
	}
}

func (g *Generator) emitWideBVShift(name string, a *goivy.Apply, result goInterpType, method string) (string, bool, error) {
	if len(a.Terms) != 2 {
		return "", true, fmt.Errorf("ivy2go: %s expected 2 arguments, got %d", name, len(a.Terms))
	}
	lhs, err := g.emitExpr(a.Terms[0])
	if err != nil {
		return "", true, err
	}
	rhs, err := g.emitExpr(a.Terms[1])
	if err != nil {
		return "", true, err
	}
	return fmt.Sprintf("wideBV%s(%s, %s, %d)", method, wideBVAsCall(lhs), wideBVAsCall(rhs), result.Bits), true, nil
}

// wideBVMethodFor maps a BV operator name to the helper method name
// emitted in runtime.go (e.g. "bvadd" → "Add").
func wideBVMethodFor(name string) string {
	switch name {
	case "bvand":
		return "And"
	case "bvor":
		return "Or"
	case "bvxor":
		return "Xor"
	case "+", "bvadd":
		return "Add"
	case "-", "bvsub":
		return "Sub"
	case "*", "bvmul":
		return "Mul"
	case "/", "bvudiv":
		return "Div"
	case "%", "bvurem":
		return "Mod"
	}
	return "Add"
}

// wideBVAsCall wraps an expression in toBigInt(...) so all operands
// land in *big.Int form regardless of their Go primitive type.
func wideBVAsCall(expr string) string {
	return "toBigInt(" + expr + ")"
}
