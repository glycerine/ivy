package ivy2go

import (
	"fmt"
	"sort"
	"strings"

	"github.com/glycerine/ivy/goivy"
)

// action_gen.go mirrors ivy2cpp/action_gen.go. For target=test/gen,
// the C++ generator emits one action_gen_<name> class per action that
// uses Z3 to synthesise inputs satisfying the action's reverse-image
// precondition. ivy2go's equivalent emits one actionGen<Name> struct
// per action, holding a *goivy.Solver wired through the goivy facade.
//
// M9 ships a structural skeleton: each generator constructs a Solver
// (via newIvySolver from z3.go), calls pushStateIntoSolver (from
// solver_emit.go) to assert the pre-state, then chooses inputs via
// ivyChoose as the M9.0 fallback. M9.1+ will replace the input
// selection with Solver.GetSmallModel.

// emitActionGenStructs walks the module's actions and emits one
// actionGen<Name> struct per public action. Only runs for target=test
// or target=gen.
//
// OPEN 055.3: for each action we also emit a buildPrecondition_<Name>
// helper that conjoins the state facts with a runtime-reified copy
// of action.GetUpdate(ctx).Pre.Fmlas — so the solver's input
// synthesis is informed by the action's actual reverse image.
func (g *Generator) emitActionGenStructs(w *goWriter) {
	if !g.usesZ3() {
		return
	}
	if g.Ctx != nil {
		g.Ctx.AddImport("actions", g.Config.GoivyImportPath, "")
	}
	for _, name := range g.actionGenNames() {
		g.emitPreconditionForAction(w, name)
		g.emitOneActionGenStruct(w, name)
	}
}

// emitPreconditionForAction emits a `buildPrecondition_<Name>` helper
// per action. The helper conjoins stateFactsAsClauses(state) with
// the runtime-reified Pre.Fmlas (with __fml:<param> references
// rewritten to the input symbols __in<i>_<param>).
//
// When the Pre can't be reified (unsupported Expr type), the helper
// just returns the state facts so the round-trip still works.
func (g *Generator) emitPreconditionForAction(w *goWriter, name string) {
	act, _ := g.Mod.Actions.Get2(name)
	if act == nil {
		return
	}
	params := act.GetFormalParams()

	// Compute the action's update at emit time.
	ctx := &goivy.UpdateContext{
		Domain: g.Mod,
		PVars:  map[string]bool{},
		ActCfg: g.Mod.Cfg.ActCfg,
	}
	var update *goivy.Update
	func() {
		// goivy.GetUpdate may panic on actions outside its supported
		// shape (e.g. some thunk forms). Recover and let `update`
		// stay nil so the rest of emission proceeds.
		defer func() { _ = recover() }()
		update = goivy.GetUpdate(act, ctx)
	}()

	helperName := "buildPrecondition_" + goExportedName(name)
	// Build the input-symbol substitution: __fml:<param.Name> →
	// inputs[i]. We collect a Go-source map literal so the
	// reified Pre can pick the right param by name.
	paramArgs := make([]string, len(params))
	for i, p := range params {
		if p == nil {
			continue
		}
		paramArgs[i] = fmt.Sprintf("__in%d *goivy.Const", i)
	}
	w.linef("// %s reifies action %q's Pre clauses at runtime.", helperName, name)
	w.linef("// Generated from goivy.GetUpdate(action, ctx).Pre at emit time.")
	w.linef("func %s(state *%s%s) *goivy.Clauses {", helperName, g.StateTypeName, optComma(len(params))+strings.Join(paramArgs, ", "))
	w.linef("\tbase := stateFactsAsClauses(state)")
	if update == nil || update.Pre == nil || update.Pre.IsFalse() {
		// Pre is trivially-false → action always safe → no extra
		// constraints needed.
		w.linef("\treturn base")
		w.line("}")
		w.blank()
		return
	}
	// Build the param-name → input-symbol substitution map.
	if len(params) > 0 {
		w.line("\tparamSub := map[string]*goivy.Const{")
		for i, p := range params {
			if p == nil {
				continue
			}
			w.linef("\t\t%q: __in%d,", "__fml:"+p.Name, i)
			w.linef("\t\t%q: __in%d,", p.Name, i)
		}
		w.line("\t}")
		w.line("\t_ = paramSub")
	}
	// Walk each fmla in Pre and reify it.
	w.line("\tvar extra []goivy.Expr")
	for _, fmla := range update.Pre.Fmlas {
		code, ok := g.reifyExprAsGoCode(fmla, params)
		if !ok {
			// Unsupported shape — record and skip this fmla.
			w.linef("\t// OPEN 055.3 unreifiable Pre fmla: %s", strings.ReplaceAll(fmla.String(), "\n", " "))
			continue
		}
		w.linef("\textra = append(extra, %s)", code)
	}
	w.line("\treturn conjClauses(base, extra...)")
	w.line("}")
	w.blank()
}

// optComma returns ", " when n > 0, else "". Used to splice the
// state arg comfortably with the inputs in a function signature.
func optComma(n int) string {
	if n > 0 {
		return ", "
	}
	return ""
}

// reifyExprAsGoCode walks an Expr and returns a Go source string that,
// when evaluated at runtime, constructs the same Expr. The params
// slice lets us rewrite formal-param references (`__fml:<name>` or
// bare param name) to the input symbol __in<i> in the emitted
// helper's scope.
//
// Returns (code, true) on success; (_, false) when the Expr shape
// isn't supported (caller skips the fmla and emits a comment).
func (g *Generator) reifyExprAsGoCode(e goivy.Expr, params []*goivy.Const) (string, bool) {
	if e == nil {
		return "nil", true
	}
	switch n := e.(type) {
	case *goivy.Const:
		// Param reference: rewrite to the runtime input symbol.
		// goivy uses several name conventions for the same formal
		// parameter ("b", "fml:b", "__fml:b"). Match any of them.
		for i, p := range params {
			if p == nil {
				continue
			}
			short := strings.TrimPrefix(strings.TrimPrefix(p.Name, "__fml:"), "fml:")
			if n.Name == p.Name ||
				n.Name == short ||
				n.Name == "fml:"+short ||
				n.Name == "__fml:"+short {
				return fmt.Sprintf("__in%d", i), true
			}
		}
		// Plain constant: build a fresh goivy.Const with the
		// sort reified.
		sortCode, ok := g.reifySortAsGoCode(n.CSort)
		if !ok {
			return "", false
		}
		return fmt.Sprintf("goivy.NewConst(%q, %s)", n.Name, sortCode), true
	case *goivy.LogicNot:
		body, ok := g.reifyExprAsGoCode(n.Body, params)
		if !ok {
			return "", false
		}
		return fmt.Sprintf("&goivy.LogicNot{Body: %s}", body), true
	case *goivy.LogicAnd:
		parts := make([]string, 0, len(n.Terms))
		for _, t := range n.Terms {
			code, ok := g.reifyExprAsGoCode(t, params)
			if !ok {
				return "", false
			}
			parts = append(parts, code)
		}
		return fmt.Sprintf("&goivy.LogicAnd{Terms: []goivy.Expr{%s}}", strings.Join(parts, ", ")), true
	case *goivy.LogicOr:
		parts := make([]string, 0, len(n.Terms))
		for _, t := range n.Terms {
			code, ok := g.reifyExprAsGoCode(t, params)
			if !ok {
				return "", false
			}
			parts = append(parts, code)
		}
		return fmt.Sprintf("&goivy.LogicOr{Terms: []goivy.Expr{%s}}", strings.Join(parts, ", ")), true
	case *goivy.LogicImplies:
		l, ok := g.reifyExprAsGoCode(n.T1, params)
		if !ok {
			return "", false
		}
		r, ok := g.reifyExprAsGoCode(n.T2, params)
		if !ok {
			return "", false
		}
		return fmt.Sprintf("&goivy.LogicImplies{T1: %s, T2: %s}", l, r), true
	case *goivy.LogicIff:
		l, ok := g.reifyExprAsGoCode(n.T1, params)
		if !ok {
			return "", false
		}
		r, ok := g.reifyExprAsGoCode(n.T2, params)
		if !ok {
			return "", false
		}
		return fmt.Sprintf("&goivy.LogicIff{T1: %s, T2: %s}", l, r), true
	case *goivy.Eq:
		l, ok := g.reifyExprAsGoCode(n.T1, params)
		if !ok {
			return "", false
		}
		r, ok := g.reifyExprAsGoCode(n.T2, params)
		if !ok {
			return "", false
		}
		return fmt.Sprintf("&goivy.Eq{T1: %s, T2: %s}", l, r), true

	case *goivy.LogicVariable:
		sortCode, ok := g.reifySortAsGoCode(n.VSort)
		if !ok {
			return "", false
		}
		// LogicVariable.NewVariable validates uppercase first
		// letter; rely on the mustNewVariable runtime helper for a
		// panic-on-failure facade.
		g.requireMustHelpers()
		return fmt.Sprintf("mustNewVariable(%q, %s)", n.Name, sortCode), true

	case *goivy.Apply:
		fnCode, ok := g.reifyExprAsGoCode(n.Func, params)
		if !ok {
			return "", false
		}
		argCodes := make([]string, 0, len(n.Terms))
		for _, t := range n.Terms {
			c, ok := g.reifyExprAsGoCode(t, params)
			if !ok {
				return "", false
			}
			argCodes = append(argCodes, c)
		}
		g.requireMustHelpers()
		if len(argCodes) == 0 {
			return fmt.Sprintf("mustApply(%s)", fnCode), true
		}
		return fmt.Sprintf("mustApply(%s, %s)", fnCode, strings.Join(argCodes, ", ")), true

	case *goivy.ForAll:
		varCodes, sortOK := g.reifyVariableSlice(n.Variables)
		if !sortOK {
			return "", false
		}
		body, ok := g.reifyExprAsGoCode(n.Body, params)
		if !ok {
			return "", false
		}
		return fmt.Sprintf("&goivy.ForAll{Variables: %s, Body: %s}", varCodes, body), true

	case *goivy.LogicExists:
		varCodes, sortOK := g.reifyVariableSlice(n.Variables)
		if !sortOK {
			return "", false
		}
		body, ok := g.reifyExprAsGoCode(n.Body, params)
		if !ok {
			return "", false
		}
		return fmt.Sprintf("&goivy.LogicExists{Variables: %s, Body: %s}", varCodes, body), true

	case *goivy.LogicIte:
		c, ok := g.reifyExprAsGoCode(n.Cond, params)
		if !ok {
			return "", false
		}
		t, ok := g.reifyExprAsGoCode(n.Then, params)
		if !ok {
			return "", false
		}
		e, ok := g.reifyExprAsGoCode(n.Else, params)
		if !ok {
			return "", false
		}
		g.requireMustHelpers()
		return fmt.Sprintf("mustNewIte(%s, %s, %s)", c, t, e), true

	default:
		return "", false
	}
}

// reifyVariableSlice helps the ForAll/LogicExists cases by turning a
// []*goivy.LogicVariable into a Go composite-literal string.
// Returns (code, true) when every variable's sort is reifiable.
func (g *Generator) reifyVariableSlice(vars []*goivy.LogicVariable) (string, bool) {
	if len(vars) == 0 {
		return "nil", true
	}
	parts := make([]string, 0, len(vars))
	for _, v := range vars {
		if v == nil {
			parts = append(parts, "nil")
			continue
		}
		sortCode, ok := g.reifySortAsGoCode(v.VSort)
		if !ok {
			return "", false
		}
		g.requireMustHelpers()
		parts = append(parts, fmt.Sprintf("mustNewVariable(%q, %s)", v.Name, sortCode))
	}
	return fmt.Sprintf("[]*goivy.LogicVariable{%s}", strings.Join(parts, ", ")), true
}

// reifySortAsGoCode returns a Go source expression that constructs s.
// Supports BooleanSort, UninterpretedSort, LogicEnumeratedSort,
// RangeSort (numeral bounds only), and LogicFunctionSort. Anything
// else returns false so the caller falls back.
func (g *Generator) reifySortAsGoCode(s goivy.Sort) (string, bool) {
	switch t := s.(type) {
	case *goivy.BooleanSort:
		return "goivy.Boolean", true
	case *goivy.UninterpretedSort:
		return fmt.Sprintf("&goivy.UninterpretedSort{Name: %q}", t.Name), true
	case *goivy.LogicEnumeratedSort:
		ext := make([]string, len(t.Extension))
		for i, v := range t.Extension {
			ext[i] = fmt.Sprintf("%q", v)
		}
		return fmt.Sprintf("&goivy.LogicEnumeratedSort{Name: %q, Extension: []string{%s}}",
			t.Name, strings.Join(ext, ", ")), true
	case *goivy.RangeSort:
		// Only numeral bounds are reifiable today; compiled bounds
		// (parameter references) require module context at runtime.
		lb, lbOK := t.Lb.(goivy.NumeralBound)
		ub, ubOK := t.Ub.(goivy.NumeralBound)
		if !lbOK || !ubOK {
			return "", false
		}
		lbSort, ok := g.reifySortAsGoCode(lb.Sort)
		if !ok {
			lbSort = "nil"
		}
		ubSort, ok := g.reifySortAsGoCode(ub.Sort)
		if !ok {
			ubSort = "nil"
		}
		return fmt.Sprintf(
			"&goivy.RangeSort{Name: %q, Lb: goivy.NumeralBound{Value: %q, Sort: %s}, Ub: goivy.NumeralBound{Value: %q, Sort: %s}}",
			t.Name, lb.Value, lbSort, ub.Value, ubSort,
		), true
	case *goivy.LogicFunctionSort:
		sorts := t.Sorts
		parts := make([]string, 0, len(sorts))
		for _, sub := range sorts {
			code, ok := g.reifySortAsGoCode(sub)
			if !ok {
				return "", false
			}
			parts = append(parts, code)
		}
		g.requireMustHelpers()
		return fmt.Sprintf("mustNewFunctionSort(%s)", strings.Join(parts, ", ")), true
	}
	return "", false
}

// requireMustHelpers flags the runtime to emit the must* facades
// (mustApply / mustNewVariable / mustNewIte / mustNewFunctionSort).
// All four share a single emission gate; needing any one pulls them
// all in to keep the dependency graph trivial.
func (g *Generator) requireMustHelpers() {
	if g != nil && g.Ctx != nil {
		g.Ctx.OnceGlobals["__need_musthelpers"] = true
	}
}

// actionGenNames returns the actions eligible for action-gen
// emission, in deterministic order. Mirrors replActionNames'
// filtering rule (no ext:/imp:/ivy:/__ prefixed names).
func (g *Generator) actionGenNames() []string {
	if g == nil || g.Mod == nil || g.Mod.Actions == nil {
		return nil
	}
	names := make([]string, 0)
	for name := range g.Mod.Actions.All() {
		if hasPrefixAny(name, "ext:", "imp:", "ivy:", "__") {
			continue
		}
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// emitOneActionGenStruct emits a single actionGen<Name> struct + its
// constructor and Generate method.
//
// OPEN 055.1: the Generate method now runs a real solver round-trip:
//
//  1. Build a fresh `*goivy.Const` for each formal param.
//  2. Build a trivial `*goivy.Clauses` (currently asserting `true`
//     since the reverse-image precondition derivation is the next
//     phase of work).
//  3. Call `sol.GetModelClauses(clauses)` to obtain a Z3 model.
//  4. For each input, extract the model value via
//     `ModelResult.Eval` + per-sort parsing (parseZ3Bool /
//     parseZ3Uint64). Fall back to `ivyChoose` if extraction fails.
//  5. Apply the action.
//
// This is genuine solver-driven gen even though the precondition is
// still trivial — adding real preconditions only requires changing
// step 2, with the rest of the wiring already in place.
func (g *Generator) emitOneActionGenStruct(w *goWriter, name string) {
	structName := "actionGen_" + goExportedName(name)
	w.linef("// %s drives solver-backed input synthesis for the %s action.", structName, name)
	w.open(fmt.Sprintf("type %s struct {", structName))
	w.line("sol *goivy.Solver")
	w.close("")
	w.blank()

	w.open(fmt.Sprintf("func new%s(state *%s) *%s {", structName, g.StateTypeName, structName))
	w.linef("g := &%s{sol: newIvySolver()}", structName)
	w.line("_ = pushStateIntoSolver(state, g.sol)")
	w.line("return g")
	w.close("")
	w.blank()

	act, _ := g.Mod.Actions.Get2(name)
	if act == nil {
		return
	}
	params := act.GetFormalParams()

	w.open(fmt.Sprintf("func (g *%s) Generate(state *%s) {", structName, g.StateTypeName))
	if len(params) == 0 {
		// Zero-arg action: nothing to solve for; just call.
		w.linef("state.%s()", goExportedName(name))
		w.close("")
		w.blank()
		return
	}
	// 1. Build per-input symbols + a trivial Clauses.
	w.line("// OPEN 055.1: build a Clauses for the action's reverse")
	w.line("// image and extract input values from the solved model.")
	w.line("// Trivial-true precondition for now (OPEN 055.2 will derive)")
	w.line("// the real reverse image via goivy.ActionUpdate analysis.")
	for i, p := range params {
		w.linef("__in%d := goivy.NewConst(\"__in%d_%s\", goivy.Boolean)", i, i, goIdent(p.Name))
		w.linef("_ = __in%d", i)
	}
	w.line("var modelResult *goivy.ModelResult")
	w.line("if g.sol != nil {")
	w.line("\t// OPEN 055.3: seed with state facts AND the action's")
	w.line("\t// own reified Pre clauses (action-specific).")
	preInputs := make([]string, 0, len(params))
	for i, p := range params {
		if p == nil {
			continue
		}
		_ = p
		preInputs = append(preInputs, fmt.Sprintf("__in%d", i))
	}
	preCall := "buildPrecondition_" + goExportedName(name) + "(state"
	if len(preInputs) > 0 {
		preCall += ", " + strings.Join(preInputs, ", ")
	}
	preCall += ")"
	w.linef("\tpreclauses := %s", preCall)
	w.line("\tmodelResult, _ = g.sol.GetModelClauses(preclauses)")
	w.line("}")
	w.line("_ = modelResult")
	// 2. Pick each input. Try the model first; fall back to
	// ivyChoose. Per D6 we don't use generics — the per-card path
	// returns uint64 and the caller inlines the typed cast.
	callArgs := make([]string, len(params))
	for i, p := range params {
		if p == nil {
			callArgs[i] = ""
			continue
		}
		typeName := g.goType(p.CSort)
		card := goSortCard(g, p.CSort)
		switch {
		case typeName == "bool":
			w.linef("v%d := pickBoolOrChoose(modelResult, __in%d)", i, i)
		case card > 0:
			w.linef("v%d := %s(pickUintOrChoose(modelResult, __in%d, %d))", i, typeName, i, card)
		default:
			w.linef("v%d := %s(ivyChoose(2))", i, typeName)
		}
		callArgs[i] = fmt.Sprintf("v%d", i)
	}
	w.linef("state.%s(%s)", goExportedName(name), strings.Join(callArgs, ", "))
	w.close("")
	w.blank()
	// Mark that the runtime needs the pickInput helpers.
	g.Ctx.OnceGlobals["__need_pickinput"] = true
}

// emitCloseSolver writes a Close method on each action generator so
// callers can free the solver. Aggregated emission keeps action_gen.go
// from sprouting per-struct close methods.
func (g *Generator) emitCloseSolver(w *goWriter) {
	if !g.usesZ3() {
		return
	}
	for _, name := range g.actionGenNames() {
		structName := "actionGen_" + goExportedName(name)
		w.open(fmt.Sprintf("func (g *%s) Close() error {", structName))
		w.open("if g.sol != nil {")
		w.line("err := g.sol.Close()")
		w.line("g.sol = nil")
		w.line("return err")
		w.close("")
		w.line("return nil")
		w.close("")
		w.blank()
	}
}
