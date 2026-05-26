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
	// back to the safer clamp form. Use the named type for the
	// result so comparisons / arithmetic against typed variables
	// (e.g. `i < 4` where i:Idx) don't trip Go's type checker.
	x := c.Name
	typ := g.goType(c.CSort)
	if typ == "" {
		typ = "int"
	}
	return fmt.Sprintf("func() %s { v := %s(%s); lo, hi := %s(%s), %s(%s); %s }()",
		typ, typ, x, typ, lo, typ, hi, rangeClampExpr("v", "lo", "hi")), true
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
		expr := obj + "." + field
		for _, a := range args {
			expr += "[" + a + "]"
		}
		return expr, true, nil
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
