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
	if !g.isStateSymbolName(name) {
		return g.goLocalFunctionAccess(name, fnExpr.NodeSort(), args), nil
	}
	return g.goStorageAccess(name, fnExpr.NodeSort(), args), nil
}

// goStorageAccess returns the Go expression that reads or writes
// function / relation symbol `name` with the given args. Mirrors
// ivy2cpp/expr.go cppStorageAccess but lowers to Go array / map
// indexing.
//
//   - Scalar (no args):              s.<Name>
//   - Array storage (small fns):     s.<Name>[arg0][arg1]…
//   - Hash-thunk read (lhsContext=false): s.get<Name>(<key>)
//   - Hash-thunk write (lhsContext=true): s.<Name>[<key>]
//
// The lhsContext gate (OPEN 061.1) lets emitAssign emit an
// assignable map expression while non-assignment readers transparently
// fall through to the thunk slot.
func (g *Generator) goStorageAccess(name string, fnSort goivy.Sort, args []string) string {
	return g.goStorageAccessBase("s."+goExportedName(name), name, fnSort, args, true)
}

// goLocalFunctionAccess is the non-state sibling of goStorageAccess.
// Function-valued action parameters and locals have the same physical
// storage shape as state functions, but their base expression is the
// local Go identifier rather than a *State field. Map-backed locals
// also read through the raw map index because they do not have a
// thunk/getter slot.
func (g *Generator) goLocalFunctionAccess(name string, fnSort goivy.Sort, args []string) string {
	return g.goStorageAccessBase(goIdent(name), name, fnSort, args, false)
}

func (g *Generator) goStorageAccessBase(base, name string, fnSort goivy.Sort, args []string, stateSymbol bool) string {
	field := base
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
		return field + g.compactArrayIndexSuffix(fs.Domain(), args)
	case goStorageHashThunk:
		// Read context routes through the per-symbol getter so the
		// thunk slot is honoured. Write context (LHS of assignment)
		// must produce an assignable expression — raw map index.
		key := args[0]
		if len(args) > 1 {
			keyType := goCTupleNameWith(g, fs.Domain())
			var b strings.Builder
			b.WriteString(keyType)
			b.WriteByte('{')
			for i, a := range args {
				if i > 0 {
					b.WriteString(", ")
				}
				b.WriteString(a)
			}
			b.WriteByte('}')
			key = b.String()
		}
		if g.lhsContext {
			return field + "[" + key + "]"
		}
		if !stateSymbol {
			return field + "[" + key + "]"
		}
		return "s.get" + goExportedName(name) + "(" + key + ")"
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

// isDefinitionName mirrors ivy2cpp/definitions.go:103. Routes through
// the literal-port allDefinitions catalog so derived AND native
// definitions both register.
func (g *Generator) isDefinitionName(name string) bool {
	if g == nil || name == "" {
		return false
	}
	for _, d := range g.allDefinitions() {
		if d.Name == name {
			return true
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
		typ := g.goType(rng)
		return rangeClampExpr(fmt.Sprintf("%s(%s)", typ, operand), lo, hi, typ), true, nil
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
	typ := g.goType(a.NodeSort())
	return rangeClampExpr(fmt.Sprintf("(%s) %s (%s)", l, name, r), lo, hi, typ), true, nil
}

// rangeClampExpr returns an expression-shaped Go IIFE that evaluates x
// once and clamps it to [lo, hi]. This is the Go counterpart of
// ivy_to_cpp.py's ternary clamp expression.
func rangeClampExpr(x, lo, hi, typ string) string {
	return fmt.Sprintf("func() %s { x := %s; lo, hi := %s(%s), %s(%s); if x < lo { return lo }; if hi < x { return hi }; return x }()", typ, x, typ, lo, typ, hi)
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
	// back to the safer clamp form. Use the named type for the
	// result so comparisons / arithmetic against typed variables
	// (e.g. `i < 4` where i:Idx) don't trip Go's type checker.
	x := c.Name
	typ := g.goType(c.CSort)
	if typ == "" {
		typ = "int"
	}
	return rangeClampExpr(fmt.Sprintf("%s(%s)", typ, x), lo, hi, typ), true
}

// emitQuant lowers `forall x:T . body(x)` or `exists x:T . body(x)`
// into a Go IIFE (immediately-invoked function expression) that
// scans the sort with an early-exit loop. Mirrors ivy2cpp/expr.go
// emitQuant: nested loops for multi-variable quantifiers, return
// false on the first violation (forall) or true on the first hit
// (exists).
//
// Per-sort iteration shape:
//   - bool → range over [false, true]
//   - enum / range / uninterpreted with known finite card → integer
//     for-loop cast to the sort's Go type
//   - anything else → error (deferred to a later milestone alongside
//     extensional-relation iteration)
func (g *Generator) emitQuant(vars []*goivy.LogicVariable, body goivy.Expr, forall bool) (string, error) {
	if len(vars) == 0 {
		return g.emitExpr(body)
	}
	// Fast-path: when the body is a single ∃ on an extensional
	// relation atom (e.g. `exists X. link(X, Y)`), iterate the
	// relation's actual map entries instead of the full Cartesian
	// product. Mirrors cpp's emitExtensionalQuant call.
	if code, ok, err := g.emitExtensionalQuant(vars, body, forall); ok || err != nil {
		return code, err
	}
	// Fast-path: when the first variable has an `iterable` attribute,
	// use the user-supplied iter sort instead of the full enum domain.
	if h, ok, err := g.quantIterableHeader(vars[0]); ok || err != nil {
		if err != nil {
			return "", err
		}
		return g.emitQuantWithHeaders(vars, body, forall, []string{h}, true)
	}
	headers := make([]string, len(vars))
	closers := make([]string, len(vars))
	for i, v := range vars {
		h, c, err := g.loopHeaderForVar(v)
		if err != nil {
			return "", err
		}
		headers[i] = h
		closers[i] = c
	}
	inner, err := g.emitExpr(body)
	if err != nil {
		return "", err
	}

	var b strings.Builder
	b.WriteString("func() bool {\n")
	for _, h := range headers {
		b.WriteString("\t")
		b.WriteString(h)
		b.WriteString("\n")
	}
	if forall {
		b.WriteString("\t\tif !(")
		b.WriteString(inner)
		b.WriteString(") { return false }\n")
	} else {
		b.WriteString("\t\tif ")
		b.WriteString(inner)
		b.WriteString(" { return true }\n")
	}
	// Close loops in reverse order.
	for i := len(closers) - 1; i >= 0; i-- {
		b.WriteString("\t")
		b.WriteString(closers[i])
		b.WriteString("\n")
	}
	if forall {
		b.WriteString("\treturn true\n")
	} else {
		b.WriteString("\treturn false\n")
	}
	b.WriteString("}()")
	return b.String(), nil
}

// loopHeaderForVar returns the Go loop header for iterating over a
// variable's sort, plus the matching close line. Mirrors
// ivy2cpp/expr.go loopHeaderForVar / loopHeaderForSort.
//
// The returned header/closer pair brackets the loop body. The body
// must use the variable's name (passed via v.Name) — the loop
// variable is named accordingly.
func (g *Generator) loopHeaderForVar(v *goivy.LogicVariable) (string, string, error) {
	if v == nil {
		return "", "", fmt.Errorf("ivy2go: nil loop variable")
	}
	return g.loopHeaderForSort(v.VSort, goIdent(v.Name))
}

func (g *Generator) loopHeaderForSort(s goivy.Sort, name string) (string, string, error) {
	switch sv := s.(type) {
	case *goivy.BooleanSort:
		return fmt.Sprintf("for _, %s := range [2]bool{false, true} {", name), "}", nil
	case *goivy.LogicEnumeratedSort:
		if isNumericEnum(sv) {
			card := len(sv.Extension)
			return fmt.Sprintf("for %s := 0; %s < %d; %s++ {", name, name, card, name), "}", nil
		}
		typeName := goExportedName(sv.Name)
		card := len(sv.Extension)
		return fmt.Sprintf("for __i := 0; __i < %d; __i++ { %s := %s(__i)", card, name, typeName), "}", nil
	case *goivy.RangeSort:
		lo, hi, ok := numericRangeBounds(sv)
		if !ok {
			return "", "", fmt.Errorf("ivy2go: cannot emit bounded loop over non-numeric range %s", sortName(s))
		}
		typ := g.goType(s)
		return fmt.Sprintf("for %s := %s(%s); %s <= %s(%s); %s++ {", name, typ, lo, name, typ, hi, name), "}", nil
	}
	if rs, ok := g.rangeSortFor(s); ok {
		lo, hi, ok := numericRangeBounds(rs)
		if !ok {
			return "", "", fmt.Errorf("ivy2go: cannot emit bounded loop over non-numeric range %s", sortName(s))
		}
		typ := g.goType(s)
		return fmt.Sprintf("for %s := %s(%s); %s <= %s(%s); %s++ {", name, typ, lo, name, typ, hi, name), "}", nil
	}
	if goIsAnyIntegerType(g, s) {
		card := goSortCard(g, s)
		if card > 0 {
			typ := g.goType(s)
			return fmt.Sprintf("for __i := 0; __i < %d; __i++ { %s := %s(__i)", card, name, typ), "}", nil
		}
	}
	return "", "", fmt.Errorf("ivy2go: cannot emit bounded loop over %s", sortName(s))
}

func (g *Generator) emitNativeExpr(n *goivy.LogicNativeExpr) (string, error) {
	_ = n
	return "", fmt.Errorf("ivy2go: native expression emission deferred (M10)")
}

func (g *Generator) emitSome(s *goivy.LogicSome) (string, error) {
	_ = s
	return "", fmt.Errorf("ivy2go: `some` expression emission deferred (M4)")
}

// emitVariantRelation lowers Ivy's `super *> sub` downcast operator
// into a tag-check + payload extraction. Mirrors ivy2cpp/expr.go
// emitVariantRelation.
//
// Result expression shape:
//
//	(super.Tag == <idx> && (*super.<Sub>) == sub)
//
// where <idx> is the variant index and <Sub> is the payload field
// name. For plain-leaf subtypes the body collapses to just the tag
// check (no payload to compare).
func (g *Generator) emitVariantRelation(name string, terms []goivy.Expr) (string, bool, error) {
	if name != "*>" {
		return "", false, nil
	}
	if len(terms) != 2 {
		return "", true, fmt.Errorf("ivy2go: variant relation *> expected 2 arguments, got %d", len(terms))
	}
	if g == nil || g.Mod == nil || !g.Mod.IsVariant(terms[0].NodeSort(), terms[1].NodeSort()) {
		return "", true, fmt.Errorf("ivy2go: %s *> %s is not a known variant relation", terms[0].String(), terms[1].String())
	}
	lhs, err := g.emitExpr(terms[0])
	if err != nil {
		return "", true, err
	}
	idx := g.Mod.VariantIndex(terms[0].NodeSort(), terms[1].NodeSort())
	if idx < 0 {
		return "", true, fmt.Errorf("ivy2go: no variant index for %s in %s",
			sortName(terms[1].NodeSort()), sortName(terms[0].NodeSort()))
	}
	subName := sortName(terms[1].NodeSort())
	if g.isPlainVariantSubtypeName(subName) {
		// Plain leaves carry no payload; the tag check is the full
		// downcast predicate.
		return fmt.Sprintf("(%s.Tag == %d)", lhs, idx), true, nil
	}
	rhs, err := g.emitExpr(terms[1])
	if err != nil {
		return "", true, err
	}
	field := goExportedName(subName)
	return fmt.Sprintf("(%s.Tag == %d && (*%s.%s) == %s)", lhs, idx, lhs, field, rhs), true, nil
}

// emitDestructorApply lowers destructor reads. Given
// `field(record, args...)`, where `field` is a registered destructor,
// emit `record.Field` (for scalar fields) or `record.Field[args...]`
// (for indexed fields). Mirrors ivy2cpp/expr.go emitDestructorApply.
func (g *Generator) emitDestructorApply(name string, terms []goivy.Expr) (string, bool, error) {
	if _, ok := g.destructorFieldName(name); !ok {
		return "", false, nil
	}
	if len(terms) < 1 {
		return "", true, fmt.Errorf("ivy2go: destructor %s expected at least 1 argument", name)
	}
	obj, err := g.emitExpr(terms[0])
	if err != nil {
		return "", true, err
	}
	field := goExportedName(memName(name))
	if len(terms) == 1 {
		return obj + "." + field, true, nil
	}
	args := make([]string, len(terms)-1)
	for i, t := range terms[1:] {
		code, err := g.emitExpr(t)
		if err != nil {
			return "", true, err
		}
		args[i] = code
	}
	// Pick array vs map indexing based on destructor sig.
	domain, rng := g.destructorFieldSig(name)
	st := goFunctionStorageFor(g, domain, rng)
	switch st.Kind {
	case goStorageArray:
		return obj + "." + field + g.compactArrayIndexSuffix(domain, args), true, nil
	case goStorageHashThunk:
		if len(args) == 1 {
			return obj + "." + field + "[" + args[0] + "]", true, nil
		}
		keyType := goCTupleNameWith(g, domain)
		var b strings.Builder
		b.WriteString(obj)
		b.WriteString(".")
		b.WriteString(field)
		b.WriteString("[")
		b.WriteString(keyType)
		b.WriteString("{")
		for i, a := range args {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(a)
		}
		b.WriteString("}]")
		return b.String(), true, nil
	default:
		return obj + "." + field, true, nil
	}
}

// destructorFieldSig returns the destructor's remaining-domain and
// range sorts. Ported from ivy2cpp/expr.go destructorFieldSig.
func (g *Generator) destructorFieldSig(name string) ([]goivy.Sort, goivy.Sort) {
	if g == nil || g.Mod == nil || g.Mod.DestructorSorts == nil {
		return nil, nil
	}
	owner, ok := g.Mod.DestructorSorts[name]
	if !ok {
		return nil, nil
	}
	ownerName := sortName(owner)
	destrs := g.Mod.SortDestructors.Get(ownerName)
	for _, d := range destrs {
		if d != nil && d.Name == name {
			if fs, ok := d.CSort.(*goivy.LogicFunctionSort); ok {
				dom := fs.Domain()
				if len(dom) > 0 {
					dom = dom[1:]
				}
				return dom, fs.Range()
			}
		}
	}
	return nil, nil
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

// nameOfTerm mirrors ivy2cpp/expr.go nameOfTerm. Returns the
// variable / const name of `t`, or "" for any other shape.
func nameOfTerm(t goivy.Expr) string {
	switch x := t.(type) {
	case *goivy.LogicVariable:
		return x.Name
	case *goivy.Const:
		return x.Name
	}
	return ""
}

// nameIn mirrors ivy2cpp/expr.go nameIn. Reports whether `n`
// matches any variable's name in `vars`.
func nameIn(vars []*goivy.LogicVariable, n string) bool {
	for _, v := range vars {
		if v != nil && v.Name == n {
			return true
		}
	}
	return false
}

// containsVariableByName mirrors ivy2cpp/expr.go
// containsVariableByName. Reports whether any term in `terms` is a
// LogicVariable whose name matches `name`.
func containsVariableByName(terms []goivy.Expr, name string) bool {
	for _, t := range terms {
		if v, ok := t.(*goivy.LogicVariable); ok && v.Name == name {
			return true
		}
	}
	return false
}

// variantPayloadField mirrors ivy2cpp/expr.go variantPayloadField.
// Returns the synthetic payload-field name for a variant sort —
// `__<sort>` (Go: kept as bare-identifier form for parity).
func variantPayloadField(s goivy.Sort) string {
	return "__" + goIdent(sortName(s))
}

// sortCardinalityAttr mirrors ivy2cpp/expr.go:1078. Returns the
// user-declared cardinality attribute for an UninterpretedSort
// (e.g. `attribute foo.cardinality = 8`). Used by callers that need
// to size iteration ranges over an uninterpreted sort with a
// declared bound.
func (g *Generator) sortCardinalityAttr(s goivy.Sort) (string, bool) {
	if g == nil || g.Mod == nil || g.Mod.Cfg == nil || g.Mod.Cfg.IuCfg == nil {
		return "", false
	}
	if _, ok := s.(*goivy.UninterpretedSort); !ok {
		return "", false
	}
	name := sortName(s)
	if name == "" {
		return "", false
	}
	attrKey := g.Mod.Cfg.IuCfg.ComposeNames(name, "cardinality")
	val, ok := g.Mod.Attributes.Get2(attrKey)
	if !ok {
		return "", false
	}
	var rep string
	switch v := val.(type) {
	case goivy.Expr:
		rep = goivy.ExprName(v)
		if rep == "" {
			rep = string(v.Sexp())
		}
	case interface{ Relname() string }:
		rep = v.Relname()
	case string:
		rep = v
	default:
		return "", false
	}
	if rep == "" {
		return "", false
	}
	return goIdent(g.Mod.Cfg.IuCfg.ComposeNames(name, rep)), true
}

// sortHasNegativeValues mirrors ivy2cpp/expr.go:965. Returns true
// when the sort is interpreted as signed `int`. Range / nat / bool /
// enum sorts have non-negative values so callers can safely lower-
// bound them by 0.
func (g *Generator) sortHasNegativeValues(s goivy.Sort) bool {
	text, ok := g.sortInterpString(s)
	return ok && text == "int"
}

// boundExpr mirrors ivy2cpp/expr.go boundExpr — one comparison
// application carrying its polarity inside the bound-derivation walk.
type boundExpr struct {
	app *goivy.Apply
	neg bool
}

// matchBoundExprs mirrors ivy2cpp/expr.go:879. Walks `body`
// collecting `<` / `<=` / `>` / `>=` applications that constrain
// `v0`. `exists` tracks the polarity (universal vs existential
// context). Recurses through Not / Implies / Or / And per Python
// semantics, and unfolds derived definitions when the call site
// references v0 directly.
func (g *Generator) matchBoundExprs(v0 *goivy.LogicVariable, body goivy.Expr, exists bool, res *[]boundExpr) {
	if v0 == nil || body == nil {
		return
	}
	if not, ok := body.(*goivy.LogicNot); ok {
		g.matchBoundExprs(v0, not.Body, !exists, res)
		return
	}
	if lit, ok := body.(*goivy.LogicLiteral); ok {
		nextExists := exists
		if lit.Polarity == 0 {
			nextExists = !exists
		}
		g.matchBoundExprs(v0, lit.Atom, nextExists, res)
		return
	}
	app, isApp := body.(*goivy.Apply)
	if isApp {
		switch goivy.ExprName(app.Func) {
		case "<", "<=", ">", ">=":
			*res = append(*res, boundExpr{app: app, neg: !exists})
		}
	}
	if imp, ok := body.(*goivy.LogicImplies); ok {
		if !exists {
			g.matchBoundExprs(v0, imp.T1, !exists, res)
			g.matchBoundExprs(v0, imp.T2, exists, res)
		}
		return
	}
	if or, ok := body.(*goivy.LogicOr); ok {
		if !exists {
			for _, t := range or.Terms {
				g.matchBoundExprs(v0, t, exists, res)
			}
		}
		return
	}
	if and, ok := body.(*goivy.LogicAnd); ok {
		if exists {
			for _, t := range and.Terms {
				g.matchBoundExprs(v0, t, exists, res)
			}
		}
		return
	}
	if !isApp {
		return
	}
	if !containsVariableByName(app.Terms, v0.Name) {
		return
	}
	def, ok := g.definitionByName(goivy.ExprName(app.Func))
	if !ok || def.RHS == nil {
		return
	}
	if !allArgsVariable(def.Params) || len(def.Params) != len(app.Terms) {
		return
	}
	subs := map[goivy.NodeKey]goivy.Expr{}
	for i, p := range def.Params {
		pv, isVar := p.(*goivy.LogicVariable)
		if !isVar {
			return
		}
		subs[goivy.Key(pv)] = app.Terms[i]
	}
	substituted, err := goivy.Substitute(def.RHS, subs)
	if err != nil {
		return
	}
	g.matchBoundExprs(v0, substituted, exists, res)
}

// getBounds mirrors ivy2cpp/expr.go:999. Returns (lo, hi) as Go
// expression strings for a quantified variable, or an error if no
// bound could be derived.
func (g *Generator) getBounds(v0 *goivy.LogicVariable, others []*goivy.LogicVariable, body goivy.Expr, exists bool) (string, string, error) {
	var bes []boundExpr
	g.matchBoundExprs(v0, body, exists, &bes)
	var los, his []string
	for _, be := range bes {
		op := goivy.ExprName(be.app.Func)
		strict := op == "<" || op == ">"
		args := be.app.Terms
		if op == ">" || op == ">=" {
			args = []goivy.Expr{args[1], args[0]}
		}
		if be.neg {
			strict = !strict
			args = []goivy.Expr{args[1], args[0]}
		}
		if len(args) != 2 {
			continue
		}
		ln := nameOfTerm(args[0])
		rn := nameOfTerm(args[1])
		if ln == v0.Name && rn != v0.Name && !nameIn(others, rn) {
			e, err := g.emitExpr(args[1])
			if err != nil {
				return "", "", err
			}
			if strict {
				his = append(his, e)
			} else {
				his = append(his, "("+e+")+1")
			}
		}
		if rn == v0.Name && ln != v0.Name && !nameIn(others, ln) {
			e, err := g.emitExpr(args[0])
			if err != nil {
				return "", "", err
			}
			if strict {
				los = append(los, "("+e+")+1")
			} else {
				los = append(los, e)
			}
		}
	}
	if !g.sortHasNegativeValues(v0.VSort) {
		los = append(los, "0")
	}
	if card := goSortCard(g, v0.VSort); card > 0 {
		his = append(his, strconv.Itoa(card))
	}
	if rs, ok := g.rangeSortFor(v0.VSort); ok {
		lo, hi, ok2 := numericRangeBounds(rs)
		if ok2 {
			los = append(los, lo)
			his = append(his, "("+hi+")+1")
		}
	}
	if len(los) == 0 {
		return "", "", fmt.Errorf("ivy2go: cannot find a lower bound for %s", v0.Name)
	}
	if len(his) == 0 {
		if hi, ok := g.sortCardinalityAttr(v0.VSort); ok {
			his = append(his, hi)
		} else {
			return "", "", fmt.Errorf("ivy2go: cannot find an upper bound for %s", v0.Name)
		}
	}
	return los[0], his[0], nil
}

// getAllBounds mirrors ivy2cpp/expr.go:1118. Derives bounds for each
// variable in `vars`, using the *remaining* variables (slice tail) as
// the others-set so siblings can't appear on the constant side.
func (g *Generator) getAllBounds(vars []*goivy.LogicVariable, body goivy.Expr, exists bool) ([][2]string, error) {
	bounds := make([][2]string, len(vars))
	for i, v := range vars {
		lo, hi, err := g.getBounds(v, vars[i+1:], body, exists)
		if err != nil {
			return nil, err
		}
		bounds[i] = [2]string{lo, hi}
	}
	return bounds, nil
}

// emitQuantWithHeaders mirrors ivy2cpp/expr.go:554. Renders a
// quantifier expression as a Go `func() bool { … }()` closure
// using caller-provided loop headers. When `iterFirstOnly` is set,
// only the first header is opened here; the remaining vars are
// quantified recursively through emitQuant so the inner closure
// uses whatever bound-derivation / iterable / generic header the
// sub-call picks.
func (g *Generator) emitQuantWithHeaders(vars []*goivy.LogicVariable, body goivy.Expr, forall bool, headers []string, iterFirstOnly bool) (string, error) {
	var b strings.Builder
	b.WriteString("func() bool {\n")
	emitted := headers
	if iterFirstOnly {
		emitted = headers[:1]
	}
	for _, h := range emitted {
		b.WriteString("\t")
		b.WriteString(h)
		b.WriteString("\n")
	}
	if iterFirstOnly && len(vars) > 1 {
		inner, err := g.emitQuant(vars[1:], body, forall)
		if err != nil {
			return "", err
		}
		if forall {
			b.WriteString("\t\tif !(")
			b.WriteString(inner)
			b.WriteString(") { return false }\n")
		} else {
			b.WriteString("\t\tif ")
			b.WriteString(inner)
			b.WriteString(" { return true }\n")
		}
	} else {
		expr, err := g.emitExpr(body)
		if err != nil {
			return "", err
		}
		if forall {
			b.WriteString("\t\tif !(")
			b.WriteString(expr)
			b.WriteString(") { return false }\n")
		} else {
			b.WriteString("\t\tif ")
			b.WriteString(expr)
			b.WriteString(" { return true }\n")
		}
	}
	for range emitted {
		b.WriteString("\t}\n")
	}
	if forall {
		b.WriteString("\treturn true\n")
	} else {
		b.WriteString("\treturn false\n")
	}
	b.WriteString("}()")
	return b.String(), nil
}

// quantIterableHeader mirrors ivy2cpp/expr.go:604. Emits an
// iter-based Go `for` header when the variable's sort declares an
// `iterable` attribute with companion `create` / `is_end` / `next`
// helpers. Used by quantifier emission to iterate over a
// user-supplied finite proxy instead of the full uninterpreted sort.
func (g *Generator) quantIterableHeader(v *goivy.LogicVariable) (string, bool, error) {
	if v == nil {
		return "", false, nil
	}
	iterPrefix, iterSort, ok := g.iterableSortFor(v.VSort)
	if !ok {
		return "", false, nil
	}
	createName := goIdent(g.Mod.Cfg.IuCfg.ComposeNames(iterPrefix, "create"))
	isEndName := goIdent(g.Mod.Cfg.IuCfg.ComposeNames(iterPrefix, "is_end"))
	nextName := goIdent(g.Mod.Cfg.IuCfg.ComposeNames(iterPrefix, "next"))
	idx := goIdent(v.Name)
	return fmt.Sprintf("for %s := %s(%s(0)); !%s(%s); %s = %s(%s) {",
		idx, g.goType(iterSort), createName, isEndName, idx, idx, nextName, idx), true, nil
}

// emitExistsVariantRelation mirrors ivy2cpp/expr.go:620. Fast-path
// for `exists X. <super> *> X` (one-variable existential check on a
// variant downcast): emits `(<super>.tag == <leaf-index>)` instead
// of looping. Returns ("", false, nil) when the formula isn't of
// that shape.
func (g *Generator) emitExistsVariantRelation(vars []*goivy.LogicVariable, body goivy.Expr) (string, bool, error) {
	if len(vars) != 1 || g == nil || g.Mod == nil {
		return "", false, nil
	}
	app, ok := body.(*goivy.Apply)
	if !ok || goivy.ExprName(app.Func) != "*>" || len(app.Terms) != 2 {
		return "", false, nil
	}
	bound := vars[0]
	if bound == nil || goivy.ExprName(app.Terms[1]) != bound.Name {
		return "", false, nil
	}
	if !g.Mod.IsVariant(app.Terms[0].NodeSort(), bound.VSort) {
		return "", true, fmt.Errorf("ivy2go: %s *> %s is not a known variant relation", app.Terms[0].String(), bound.String())
	}
	lhs, err := g.emitExpr(app.Terms[0])
	if err != nil {
		return "", true, err
	}
	idx := g.Mod.VariantIndex(app.Terms[0].NodeSort(), bound.VSort)
	if idx < 0 {
		return "", true, fmt.Errorf("ivy2go: no variant index for %s in %s", sortName(bound.VSort), sortName(app.Terms[0].NodeSort()))
	}
	return fmt.Sprintf("(%s.Tag == %d)", lhs, idx), true, nil
}

// emitExtensionalQuant mirrors ivy2cpp/expr.go:654. Iterates the
// extensional relation's actual entries (its backing map) instead
// of the full Cartesian product of the variable sorts. Returns
// ("", false, nil) when the formula isn't of the extensional shape.
func (g *Generator) emitExtensionalQuant(vars []*goivy.LogicVariable, body goivy.Expr, forall bool) (string, bool, error) {
	if len(vars) == 0 || g == nil || g.Mod == nil {
		return "", false, nil
	}
	v0 := vars[0]
	if v0 == nil {
		return "", false, nil
	}
	exists := !forall
	var ebnds []*goivy.Apply
	g.matchExtensionalBoundExprs(v0, body, exists, &ebnds)
	if len(ebnds) == 0 {
		return "", false, nil
	}
	ebnd := ebnds[0]
	fs, ok := ebnd.Func.NodeSort().(*goivy.LogicFunctionSort)
	if !ok {
		return "", false, nil
	}
	st := goFunctionStorageFor(g, fs.Domain(), fs.Range())
	if st.Kind != goStorageHashThunk {
		return "", false, nil
	}
	relName := goivy.ExprName(ebnd.Func)
	rel := "s." + goExportedName(relName)

	var b strings.Builder
	b.WriteString("func() bool {\n")
	b.WriteString(fmt.Sprintf("\tfor __k, __v := range %s {\n", rel))
	b.WriteString("\t\tif !__v { continue }\n")

	// Bind each quantified variable that appears as an argument of
	// ebnd to the corresponding field of the map key.
	boundNames := map[string]bool{}
	multiArg := len(ebnd.Terms) > 1
	for pos, term := range ebnd.Terms {
		tv, isVar := term.(*goivy.LogicVariable)
		if !isVar {
			continue
		}
		var matched *goivy.LogicVariable
		for _, v := range vars {
			if v != nil && v.Name == tv.Name {
				matched = v
				break
			}
		}
		if matched == nil || boundNames[matched.Name] {
			continue
		}
		boundNames[matched.Name] = true
		if multiArg {
			b.WriteString(fmt.Sprintf("\t\t%s := __k.Arg%d\n", goIdent(matched.Name), pos))
		} else {
			b.WriteString(fmt.Sprintf("\t\t%s := __k\n", goIdent(matched.Name)))
		}
		b.WriteString(fmt.Sprintf("\t\t_ = %s\n", goIdent(matched.Name)))
	}

	// Open generic loops over any quantified variable not bound by
	// the extensional atom.
	var remaining []*goivy.LogicVariable
	for _, v := range vars {
		if v != nil && !boundNames[v.Name] {
			remaining = append(remaining, v)
		}
	}
	nestedOpened := 0
	for _, v := range remaining {
		header, _, err := g.loopHeaderForVar(v)
		if err != nil {
			return "", true, err
		}
		b.WriteString("\t\t")
		b.WriteString(header)
		b.WriteString("\n")
		nestedOpened++
	}

	expr, err := g.emitExpr(body)
	if err != nil {
		return "", true, err
	}
	if forall {
		b.WriteString(fmt.Sprintf("\t\tif !(%s) { return false }\n", expr))
	} else {
		b.WriteString(fmt.Sprintf("\t\tif %s { return true }\n", expr))
	}

	for i := 0; i < nestedOpened; i++ {
		b.WriteString("\t\t}\n")
	}
	b.WriteString("\t}\n")
	if forall {
		b.WriteString("\treturn true\n")
	} else {
		b.WriteString("\treturn false\n")
	}
	b.WriteString("}()")
	return b.String(), true, nil
}

// matchExtensionalBoundExprs mirrors ivy2cpp/expr.go:765. Walks
// `body` collecting extensional-relation applications that
// constrain `v0`. Tracks polarity through Not / Implies / Or / And,
// and unfolds derived definitions when the call references v0.
func (g *Generator) matchExtensionalBoundExprs(v0 *goivy.LogicVariable, body goivy.Expr, exists bool, res *[]*goivy.Apply) {
	if v0 == nil || body == nil {
		return
	}
	if not, ok := body.(*goivy.LogicNot); ok {
		g.matchExtensionalBoundExprs(v0, not.Body, !exists, res)
		return
	}
	if lit, ok := body.(*goivy.LogicLiteral); ok {
		nextExists := exists
		if lit.Polarity == 0 {
			nextExists = !exists
		}
		g.matchExtensionalBoundExprs(v0, lit.Atom, nextExists, res)
		return
	}
	app, isApp := body.(*goivy.Apply)
	if isApp {
		name := goivy.ExprName(app.Func)
		if name != "" && g.extensionalRels()[name] && exists && containsVariableByName(app.Terms, v0.Name) {
			*res = append(*res, app)
		}
	}
	if imp, ok := body.(*goivy.LogicImplies); ok {
		if !exists {
			g.matchExtensionalBoundExprs(v0, imp.T1, !exists, res)
			g.matchExtensionalBoundExprs(v0, imp.T2, exists, res)
		}
		return
	}
	if or, ok := body.(*goivy.LogicOr); ok {
		if !exists {
			for _, t := range or.Terms {
				g.matchExtensionalBoundExprs(v0, t, exists, res)
			}
		}
		return
	}
	if and, ok := body.(*goivy.LogicAnd); ok {
		if exists {
			for _, t := range and.Terms {
				g.matchExtensionalBoundExprs(v0, t, exists, res)
			}
		}
		return
	}
	if !isApp {
		return
	}
	if !containsVariableByName(app.Terms, v0.Name) {
		return
	}
	def, ok := g.definitionByName(goivy.ExprName(app.Func))
	if !ok || def.RHS == nil {
		return
	}
	if !allArgsVariable(def.Params) || len(def.Params) != len(app.Terms) {
		return
	}
	subs := map[goivy.NodeKey]goivy.Expr{}
	for i, p := range def.Params {
		pv, isVar := p.(*goivy.LogicVariable)
		if !isVar {
			return
		}
		subs[goivy.Key(pv)] = app.Terms[i]
	}
	substituted, err := goivy.Substitute(def.RHS, subs)
	if err != nil {
		return
	}
	g.matchExtensionalBoundExprs(v0, substituted, exists, res)
}

// finiteValues mirrors ivy2cpp/expr.go:1347. Returns the finite
// extension of a small enumerable sort, or (nil, false) for sorts
// whose values can't be enumerated cheaply.
func finiteValues(s goivy.Sort) ([]string, bool) {
	switch st := s.(type) {
	case *goivy.BooleanSort:
		return []string{"false", "true"}, true
	case *goivy.LogicEnumeratedSort:
		vals := make([]string, len(st.Extension))
		for i, v := range st.Extension {
			vals[i] = goExportedName(v)
		}
		return vals, true
	default:
		return nil, false
	}
}

// iterableSortFor mirrors ivy2cpp/expr.go:1138. Resolves the
// companion `<sort>.iter` (or `<sort>.iter.t`) sort when an
// UninterpretedSort carries an `iterable` attribute. Callers use the
// iter sort as a finite proxy for the larger enumerated type.
func (g *Generator) iterableSortFor(s goivy.Sort) (string, goivy.Sort, bool) {
	if g == nil || g.Mod == nil || g.Mod.Cfg == nil || g.Mod.Cfg.IuCfg == nil || g.Mod.Sig == nil {
		return "", nil, false
	}
	us, ok := s.(*goivy.UninterpretedSort)
	if !ok {
		return "", nil, false
	}
	attrKey := g.Mod.Cfg.IuCfg.ComposeNames(us.Name, "iterable")
	if _, ok := g.Mod.Attributes.Get2(attrKey); !ok {
		return "", nil, false
	}
	iterName := g.Mod.Cfg.IuCfg.ComposeNames(us.Name, "iter")
	if iterSort, ok := g.Mod.Sig.Sorts.Get2(iterName); ok {
		return iterName, iterSort, true
	}
	iterT := g.Mod.Cfg.IuCfg.ComposeNames(iterName, "t")
	if iterSort, ok := g.Mod.Sig.Sorts.Get2(iterT); ok {
		return iterName, iterSort, true
	}
	return "", nil, false
}

// someLoopHeaders mirrors ivy2cpp/expr.go:1213. Picks per-variable
// loop headers for an existential `some` expression. Uses
// `getAllBounds` to derive tighter bounds when the variables are
// integer-typed; otherwise falls back to `loopHeaderForVar`.
func (g *Generator) someLoopHeaders(vars []*goivy.LogicVariable, body goivy.Expr) ([]string, error) {
	headers := make([]string, len(vars))
	useBounds := false
	if len(vars) > 0 && goIsAnyIntegerType(g, vars[0].VSort) {
		if bounds, err := g.getAllBounds(vars, body, true); err == nil {
			useBounds = true
			for i, v := range vars {
				h, herr := g.loopHeaderForSortBounds(v.VSort, goIdent(v.Name), bounds[i][0], bounds[i][1])
				if herr != nil {
					useBounds = false
					break
				}
				headers[i] = h
			}
		}
	}
	if useBounds {
		return headers, nil
	}
	for i, v := range vars {
		h, _, err := g.loopHeaderForVar(v)
		if err != nil {
			return nil, err
		}
		headers[i] = h
	}
	return headers, nil
}

// loopIntCType mirrors ivy2cpp/expr.go:1413. Picks the Go integer
// type used as the loop counter for `s`. Enum and BV sorts use the
// sort's natural Go type; bool widens to `int`; everything else
// defaults to the sort's Go type or `int` as a fallback.
func loopIntCType(g *Generator, s goivy.Sort) string {
	if _, ok := s.(*goivy.LogicEnumeratedSort); ok {
		return g.goType(s)
	}
	if g != nil {
		if it, ok := g.goInterpType(s); ok && it.Kind == goInterpBV {
			return g.goType(s)
		}
	}
	ct := g.goType(s)
	switch ct {
	case "bool":
		return "int"
	}
	return ct
}

// loopHeaderForSortBounds mirrors ivy2cpp/expr.go:1395. Emits a Go
// `for` loop header using explicit `lo` / `hi` integer-string
// bounds. Enum-sorted variables loop through `int` indices and cast
// the loop variable back to the named enum type on each iteration.
func (g *Generator) loopHeaderForSortBounds(s goivy.Sort, name, lo, hi string) (string, error) {
	if lo == "" || hi == "" {
		return "", fmt.Errorf("ivy2go: empty bounds for %s", name)
	}
	if es, ok := s.(*goivy.LogicEnumeratedSort); ok {
		if isNumericEnum(es) {
			return fmt.Sprintf("for %s := (%s); %s < (%s); %s++ {", name, lo, name, hi, name), nil
		}
		typeName := goExportedName(es.Name)
		return fmt.Sprintf("for __i := (%s); __i < (%s); __i++ { %s := %s(__i)", lo, hi, name, typeName), nil
	}
	typ := g.goType(s)
	return fmt.Sprintf("for %s := %s(%s); %s < %s(%s); %s++ {", name, typ, lo, name, typ, hi, name), nil
}
