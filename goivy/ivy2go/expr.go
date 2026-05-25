package ivy2go

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/glycerine/ivy/goivy"
)

// emitExpr is the top-level dispatch for translating an Ivy logic
// expression into a Go source expression. Mirrors ivy2cpp/expr.go
// emitExpr case-by-case; each case substitutes Go syntax for C++.
//
// Heavy subsystems that ivy2cpp emitExpr delegates to (quantifier
// loop generation, `some` expressions, native blocks, destructor
// access, variant relations, range arithmetic clamping) are kept in
// their own per-file emitters; M3 lands the core dispatch and the
// simple operators. Advanced lowering paths are stubs that return
// errors with a "M{N}: …" marker so they fail loudly when exercised
// prematurely.
func (g *Generator) emitExpr(e goivy.Expr) (string, error) {
	if e == nil {
		return "", fmt.Errorf("ivy2go: nil expression")
	}
	if repl, ok := g.aliasForSymbol(e); ok {
		return g.emitExpr(repl)
	}
	switch n := e.(type) {
	case *goivy.Const:
		if n.CSort == goivy.Boolean {
			switch n.Name {
			case "true":
				return "true", nil
			case "false":
				return "false", nil
			}
		}
		if code, ok := g.emitBVNumeral(n); ok {
			return code, nil
		}
		if code, ok := g.emitRangeNumeral(n); ok {
			return code, nil
		}
		if code, ok, err := g.emitStringInterpConst(n); ok || err != nil {
			return code, err
		}
		if goivy.IsNumeral(n) && !goivy.IsLiteralString(n) {
			return n.Name, nil
		}
		if g.isDefinitionName(n.Name) {
			fn, err := funName(n.Name)
			if err != nil {
				return "", err
			}
			return fn + "()", nil
		}
		// Enum constants are namespaced under their typed iota block
		// in Go, so a bare reference to `red` should emit `Red`.
		if g.isEnumConstantName(n.Name) {
			return goExportedName(n.Name), nil
		}
		// State symbols (relations / functions / signature entries)
		// live as exported fields on *State; emit `s.<Exported>`.
		if g.isStateSymbolName(n.Name) {
			return "s." + goExportedName(n.Name), nil
		}
		return goIdent(n.Name), nil
	case *goivy.LogicVariable:
		return goIdent(n.Name), nil
	case *goivy.Apply:
		return g.emitApply(n)
	case *goivy.Eq:
		l, err := g.emitExpr(n.T1)
		if err != nil {
			return "", err
		}
		r, err := g.emitExpr(n.T2)
		if err != nil {
			return "", err
		}
		return "(" + l + " == " + r + ")", nil
	case *goivy.LogicNot:
		body, err := g.emitExpr(n.Body)
		if err != nil {
			return "", err
		}
		return "!(" + body + ")", nil
	case *goivy.LogicLiteral:
		body, err := g.emitExpr(n.Atom)
		if err != nil {
			return "", err
		}
		if n.Polarity == 0 {
			return "!(" + body + ")", nil
		}
		return body, nil
	case *goivy.LogicAnd:
		return g.emitNary(n.Terms, "&&", "true")
	case *goivy.LogicOr:
		return g.emitNary(n.Terms, "||", "false")
	case *goivy.LogicImplies:
		l, err := g.emitExpr(n.T1)
		if err != nil {
			return "", err
		}
		r, err := g.emitExpr(n.T2)
		if err != nil {
			return "", err
		}
		return "(!(" + l + ") || (" + r + "))", nil
	case *goivy.LogicIff:
		l, err := g.emitExpr(n.T1)
		if err != nil {
			return "", err
		}
		r, err := g.emitExpr(n.T2)
		if err != nil {
			return "", err
		}
		return "((" + l + ") == (" + r + "))", nil
	case *goivy.LogicIte:
		// Go has no ternary; emit a call to the `ite` runtime helper.
		// One specialisation per result type lands in runtime.go via
		// requestIteHelper (M3+).
		c, err := g.emitExpr(n.Cond)
		if err != nil {
			return "", err
		}
		t, err := g.emitExpr(n.Then)
		if err != nil {
			return "", err
		}
		f, err := g.emitExpr(n.Else)
		if err != nil {
			return "", err
		}
		helperName := g.requestIteHelper(n.NodeSort())
		return fmt.Sprintf("%s(%s, %s, %s)", helperName, c, t, f), nil
	case *goivy.ForAll:
		return g.emitQuant(n.Variables, n.Body, true)
	case *goivy.LogicExists:
		return g.emitQuant(n.Variables, n.Body, false)
	case *goivy.LogicNativeExpr:
		return g.emitNativeExpr(n)
	case *goivy.LogicSome:
		return g.emitSome(n)
	case *goivy.LogicLet:
		return g.emitLetExpr(n)
	case *goivy.LogicNamedBinder:
		return "", fmt.Errorf("ivy2go: named binder %q is not supported as a Go expression: %s", n.Name, n.String())
	case *goivy.LogicGlobally, *goivy.LogicEventually, *goivy.LogicWhenOperator:
		return "", fmt.Errorf("ivy2go: temporal expression %T is not supported in Go generation: %s", e, e.String())
	default:
		return "", fmt.Errorf("ivy2go: unsupported expression %T: %s", e, e.String())
	}
}

// emitLetExpr mirrors ivy2cpp/expr.go emitLetExpr. Expands `let d in
// body` by substituting each definition's LHS for its RHS, then
// recursing.
func (g *Generator) emitLetExpr(l *goivy.LogicLet) (string, error) {
	if l == nil {
		return "", fmt.Errorf("ivy2go: nil let expression")
	}
	subs := map[goivy.NodeKey]goivy.Expr{}
	for _, d := range l.Defs {
		def, ok := d.(*goivy.LogicDefinition)
		if !ok {
			return "", fmt.Errorf("ivy2go: let definition has unsupported shape %T: %s", d, d.String())
		}
		switch sym := def.Lhs.(type) {
		case *goivy.Const:
			subs[goivy.Key(sym)] = def.Rhs
		case *goivy.LogicVariable:
			subs[goivy.Key(sym)] = def.Rhs
		default:
			return "", fmt.Errorf("ivy2go: parametric let definition not yet supported: %s", def.Lhs.String())
		}
	}
	body, err := goivy.Substitute(l.Body, subs)
	if err != nil {
		return "", fmt.Errorf("ivy2go: let substitution: %w", err)
	}
	return g.emitExpr(body)
}

// emitNary joins terms with a binary operator, using ident as the
// neutral value when terms is empty. Mirrors ivy2cpp/expr.go emitNary.
func (g *Generator) emitNary(terms []goivy.Expr, op, ident string) (string, error) {
	if len(terms) == 0 {
		return ident, nil
	}
	parts := make([]string, len(terms))
	for i, t := range terms {
		s, err := g.emitExpr(t)
		if err != nil {
			return "", err
		}
		parts[i] = "(" + s + ")"
	}
	return strings.Join(parts, " "+op+" "), nil
}

// emitApply lowers an Apply node. Mirrors ivy2cpp/expr.go emitApply.
func (g *Generator) emitApply(a *goivy.Apply) (string, error) {
	if g != nil && g.Mod != nil && g.Mod.Cfg != nil && g.Mod.Cfg.IuCfg != nil &&
		goivy.IsMacro(a, g.Mod.Cfg.IuCfg) {
		return g.emitExpr(goivy.ExpandMacro(a))
	}
	fnExpr := a.Func
	if repl, ok := g.aliasForSymbol(fnExpr); ok {
		fnExpr = repl
	}
	name := goivy.ExprName(fnExpr)
	if code, ok, err := g.emitCastApply(name, a); ok || err != nil {
		return code, err
	}
	if code, ok, err := g.emitBVApply(name, a); ok || err != nil {
		return code, err
	}
	if code, ok, err := g.emitNatMinusApply(name, a); ok || err != nil {
		return code, err
	}
	if code, ok, err := g.emitRangeArithApply(name, a); ok || err != nil {
		return code, err
	}
	if len(a.Terms) == 2 && isInfix(name) {
		l, err := g.emitExpr(a.Terms[0])
		if err != nil {
			return "", err
		}
		r, err := g.emitExpr(a.Terms[1])
		if err != nil {
			return "", err
		}
		return "(" + l + " " + name + " " + r + ")", nil
	}
	if code, ok, err := g.emitVariantRelation(name, a.Terms); ok || err != nil {
		return code, err
	}
	if code, ok, err := g.emitDestructorApply(name, a.Terms); ok || err != nil {
		return code, err
	}
	fn, err := funName(name)
	if err != nil {
		return "", err
	}
	if len(a.Terms) == 0 {
		return fn, nil
	}
	args := make([]string, len(a.Terms))
	for i, t := range a.Terms {
		s, err := g.emitExpr(t)
		if err != nil {
			return "", err
		}
		args[i] = s
	}
	if g.isDefinitionName(name) {
		return fn + "(" + strings.Join(args, ", ") + ")", nil
	}
	return g.goStorageAccess(name, fnExpr.NodeSort(), args), nil
}

// goStorageAccess returns the Go expression that reads function /
// relation symbol `name` with the given args. Mirrors ivy2cpp/expr.go
// cppStorageAccess but lowers to Go array / map indexing.
//
//   - Scalar (no args):              s.<Name>
//   - Array storage (small fns):     s.<Name>[arg0][arg1]…
//   - Hash-thunk storage (large fns):s.<Name>[<keyExpr>]
//
// The leading `s.` receiver is appropriate when this expression lives
// inside an action method on *State; we always emit it for now and
// rely on M4/M5 to refine the addressing model if needed.
func (g *Generator) goStorageAccess(name string, fnSort goivy.Sort, args []string) string {
	field := "s." + goExportedName(name)
	if len(args) == 0 {
		return field
	}
	fs, ok := fnSort.(*goivy.LogicFunctionSort)
	if !ok {
		return field + "(" + strings.Join(args, ", ") + ")"
	}
	st := goFunctionStorageFor(g, fs.Domain(), fs.Range())
	switch st.Kind {
	case goStorageArray:
		var b strings.Builder
		b.WriteString(field)
		for _, a := range args {
			b.WriteByte('[')
			b.WriteString(a)
			b.WriteByte(']')
		}
		return b.String()
	case goStorageHashThunk:
		if len(args) == 1 {
			return field + "[" + args[0] + "]"
		}
		keyType := goCTupleNameWith(g, fs.Domain())
		var b strings.Builder
		b.WriteString(field)
		b.WriteByte('[')
		b.WriteString(keyType)
		b.WriteByte('{')
		for i, a := range args {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(a)
		}
		b.WriteByte('}')
		b.WriteByte(']')
		return b.String()
	default:
		return field
	}
}

// isInfix mirrors ivy2cpp/expr.go isInfix.
func isInfix(name string) bool {
	switch name {
	case "+", "-", "*", "/", "%", "<", "<=", ">", ">=":
		return true
	default:
		return false
	}
}

// aliasForSymbol mirrors ivy2cpp/expr.go aliasForSymbol. Resolves
// symbol-renaming aliases the generator may install during action
// emission. M3 holds the alias map empty; later milestones populate.
func (g *Generator) aliasForSymbol(e goivy.Expr) (goivy.Expr, bool) {
	if g == nil || len(g.exprAliases) == 0 || e == nil {
		return nil, false
	}
	switch e.(type) {
	case *goivy.Const, *goivy.LogicVariable:
	default:
		return nil, false
	}
	name := goivy.ExprName(e)
	if name == "" {
		return nil, false
	}
	repl, ok := g.exprAliases[name]
	if !ok || repl == nil || goivy.ExprName(repl) == name {
		return nil, false
	}
	return repl, true
}

// isDefinitionName mirrors ivy2cpp's same-named helper: returns true
// when `name` is a definitional axiom (renders as `name()`).
func (g *Generator) isDefinitionName(name string) bool {
	if g == nil || g.Mod == nil || name == "" {
		return false
	}
	for _, d := range g.Mod.Definitions {
		if d == nil {
			continue
		}
		if def, ok := d.Formula.(*goivy.LogicDefinition); ok {
			if c, ok := def.Lhs.(*goivy.Const); ok && c.Name == name {
				return true
			}
			if a, ok := def.Lhs.(*goivy.Apply); ok {
				if goivy.ExprName(a.Func) == name {
					return true
				}
			}
		}
	}
	return false
}

// isStateSymbolName returns true when `name` is a signature symbol
// that lives as a field on *State (i.e., a relation, function, or
// individual constant declared in the module). Routed through
// `s.<exported>` by emitExpr.
func (g *Generator) isStateSymbolName(name string) bool {
	if g == nil || g.Mod == nil || g.Mod.Sig == nil || name == "" {
		return false
	}
	if _, ok := g.Mod.Sig.Symbols.Get2(name); ok {
		return true
	}
	return false
}

// isEnumConstantName returns true when `name` is a member of a
// non-numeric enumerated sort. Used so bare references emit the
// PascalCase Go constant defined in types.go.
func (g *Generator) isEnumConstantName(name string) bool {
	if g == nil || g.Mod == nil || g.Mod.Sig == nil || name == "" {
		return false
	}
	for _, s := range g.Mod.Sig.Sorts.All() {
		es, ok := s.(*goivy.LogicEnumeratedSort)
		if !ok || isNumericEnum(es) {
			continue
		}
		for _, v := range es.Extension {
			if v == name {
				return true
			}
		}
	}
	return false
}

// requestIteHelper records that the runtime needs an ite(cond,t,f)
// helper specialised for the Go type of s. Returns the name to call.
// One helper is emitted per distinct Go type (no generics per D6).
func (g *Generator) requestIteHelper(s goivy.Sort) string {
	typ := g.goType(s)
	suffix := iteHelperSuffix(typ)
	name := "ite_" + suffix
	if g.Ctx != nil {
		g.Ctx.OnceGlobals[name] = true // marker; real emission lands in runtime helper writer
	}
	return name
}

// iteHelperSuffix returns a safe Go-identifier suffix for a type
// expression so we can append it to "ite_". E.g. "uint32" → "uint32",
// "[3]uint32" → "arr3_uint32", "Color" → "Color".
func iteHelperSuffix(typ string) string {
	r := strings.NewReplacer(
		"[", "arr",
		"]", "_",
		"*", "ptr_",
		" ", "_",
		".", "_",
		"{", "_",
		"}", "_",
		",", "_",
	)
	return r.Replace(typ)
}

// emitStringInterpConst mirrors ivy2cpp/expr.go emitStringInterpConst.
func (g *Generator) emitStringInterpConst(c *goivy.Const) (string, bool, error) {
	if c == nil || !goivy.IsNumeral(c) || goivy.IsLiteralString(c) {
		return "", false, nil
	}
	if !g.hasStringInterp(c.CSort) {
		return "", false, nil
	}
	if c.Name == "0" {
		return `""`, true, nil
	}
	return "", true, fmt.Errorf("ivy2go: cannot compile numeral %s of string sort %s", c.Name, sortName(c.CSort))
}

// emitCastApply ports ivy2cpp/expr.go emitCastApply. The BV branch
// delegates to emitBVApply; range / nat targets clamp the operand.
func (g *Generator) emitCastApply(name string, a *goivy.Apply) (string, bool, error) {
	if name != "cast" || len(a.Terms) != 1 {
		return "", false, nil
	}
	rng := a.NodeSort()
	if _, ok := g.bvWidthForSort(rng); ok {
		return "", false, nil
	}
	operand, err := g.emitExpr(a.Terms[0])
	if err != nil {
		return "", true, err
	}
	if rs, ok := g.rangeSortFor(rng); ok {
		lo, hi, ok := numericRangeBounds(rs)
		if !ok {
			return "", true, fmt.Errorf("ivy2go: cannot emit cast to non-numeric range %s", sortName(rng))
		}
		return rangeClampExpr(operand, lo, hi), true, nil
	}
	if g.hasNatInterp(rng) {
		return natSaturateExpr(operand), true, nil
	}
	if interp, ok := g.sortInterpString(rng); ok && interp == "int" {
		return operand, true, nil
	}
	return operand, true, nil
}

// emitNatMinusApply ports ivy2cpp/expr.go emitNatMinusApply: `x - y`
// at result sort `nat` lowers to a saturating subtraction.
func (g *Generator) emitNatMinusApply(name string, a *goivy.Apply) (string, bool, error) {
	if name != "-" || len(a.Terms) != 2 {
		return "", false, nil
	}
	if !g.hasNatInterp(a.NodeSort()) {
		return "", false, nil
	}
	l, err := g.emitExpr(a.Terms[0])
	if err != nil {
		return "", true, err
	}
	r, err := g.emitExpr(a.Terms[1])
	if err != nil {
		return "", true, err
	}
	// Inline closure: like ivy2cpp's `([&](){…})()` but in Go.
	return fmt.Sprintf("func() uint64 { a, b := uint64(%s), uint64(%s); if a < b { return 0 }; return a - b }()", l, r), true, nil
}

// emitRangeArithApply ports ivy2cpp/expr.go emitRangeArithApply: binary
// arith on a range result sort clamps to [lo, hi].
func (g *Generator) emitRangeArithApply(name string, a *goivy.Apply) (string, bool, error) {
	if !isInfix(name) || len(a.Terms) != 2 {
		return "", false, nil
	}
	switch name {
	case "+", "-", "*", "/", "%":
	default:
		return "", false, nil
	}
	rs, ok := g.rangeSortFor(a.NodeSort())
	if !ok {
		return "", false, nil
	}
	lo, hi, ok := numericRangeBounds(rs)
	if !ok {
		return "", false, nil
	}
	l, err := g.emitExpr(a.Terms[0])
	if err != nil {
		return "", true, err
	}
	r, err := g.emitExpr(a.Terms[1])
	if err != nil {
		return "", true, err
	}
	body := fmt.Sprintf("x := (%s) %s (%s); %s", l, name, r, "return "+rangeClampExpr("x", lo, hi))
	return fmt.Sprintf("func() int { %s }()", body), true, nil
}

// rangeClampExpr returns the Go ternary-equivalent: nested if. Used as
// a return-value expression body in arithmetic clamping. Mirrors
// ivy2cpp/expr.go rangeClampExpr but emits an `if` chain instead of
// `?:` since Go lacks the ternary.
func rangeClampExpr(x, lo, hi string) string {
	return fmt.Sprintf("if %s < %s { return %s } else if %s < %s { return %s } else { return %s }", x, lo, lo, hi, x, hi, x)
}

// natSaturateExpr ports ivy2cpp/expr.go natSaturateExpr.
func natSaturateExpr(x string) string {
	return fmt.Sprintf("func() uint64 { v := int64(%s); if v < 0 { return 0 }; return uint64(v) }()", x)
}

// emitRangeNumeral ports ivy2cpp/expr.go emitRangeNumeral. Wraps a
// numeral on a range sort in a clamp.
func (g *Generator) emitRangeNumeral(c *goivy.Const) (string, bool) {
	if c == nil || !goivy.IsNumeral(c) || goivy.IsLiteralString(c) {
		return "", false
	}
	rs, ok := g.rangeSortFor(c.CSort)
	if !ok {
		return "", false
	}
	lo, hi, ok := numericRangeBounds(rs)
	if !ok {
		return "", false
	}
	// Literal numerals are already in-bounds at parse time; emit raw
	// digits unless the literal is out of range, in which case fall
	// back to the safer clamp form. For now we always clamp to keep
	// behaviour identical to ivy2cpp/expr.go.
	x := c.Name
	return fmt.Sprintf("func() int { v := %s; %s }()", x, rangeClampExpr("v", lo, hi)), true
}

// --- M3 stubs for advanced expression emitters -----------------------
// All of the following will be fleshed out in later milestones. Each
// returns a (true, error) so any test that exercises the path fails
// loudly rather than emitting wrong Go.

func (g *Generator) emitQuant(vars []*goivy.LogicVariable, body goivy.Expr, forall bool) (string, error) {
	_ = vars
	_ = body
	_ = forall
	return "", fmt.Errorf("ivy2go: quantifier emission deferred (M4/M5)")
}

func (g *Generator) emitNativeExpr(n *goivy.LogicNativeExpr) (string, error) {
	_ = n
	return "", fmt.Errorf("ivy2go: native expression emission deferred (M10)")
}

func (g *Generator) emitSome(s *goivy.LogicSome) (string, error) {
	_ = s
	return "", fmt.Errorf("ivy2go: `some` expression emission deferred (M4)")
}

func (g *Generator) emitVariantRelation(name string, terms []goivy.Expr) (string, bool, error) {
	if name != "*>" {
		return "", false, nil
	}
	_ = terms
	return "", true, fmt.Errorf("ivy2go: variant relation (*>) emission deferred (M8)")
}

func (g *Generator) emitDestructorApply(name string, terms []goivy.Expr) (string, bool, error) {
	if _, ok := g.destructorFieldName(name); !ok {
		return "", false, nil
	}
	_ = terms
	return "", true, fmt.Errorf("ivy2go: destructor apply emission deferred (M8)")
}

// destructorFieldName mirrors ivy2cpp/expr.go same-named helper.
// Stubbed to return false (no destructor symbols) until M8.
func (g *Generator) destructorFieldName(name string) (string, bool) {
	if g == nil || g.Mod == nil || g.Mod.DestructorSorts == nil || name == "" {
		return "", false
	}
	if _, ok := g.Mod.DestructorSorts[name]; !ok {
		return "", false
	}
	return varName(memName(name)), true
}

// Keep strconv imported until M4 actions/numerals use it.
var _ = strconv.Itoa
