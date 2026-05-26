package ivy2go

import (
	"fmt"
	"sort"
	"strings"

	"github.com/glycerine/ivy/goivy"
)

// action_gen.go is a literal port of ivy2cpp/action_gen.go. Per
// CLAUDE.md B.6–B.9 the function set, names, and call graph mirror
// ivy2cpp exactly; only the leaf-level emission swaps Go for C++.
//
// Each public action gets one generator: a struct holding a
// *goivy.Solver and one field per formal-param input, plus three
// methods:
//
//   - constructor (new<Name>) — creates the solver wrapper.
//   - generate(state) bool   — assembles state facts + the action's
//                              reverse-image precondition into a
//                              *goivy.Clauses, hands it to
//                              Solver.GetModelClauses, and returns
//                              false on UNSAT (action skipped) /
//                              true on SAT (input fields populated).
//   - execute(state)         — traces `> name(args)` and invokes
//                              state.<Name>(g.in0, g.in1, …).
//
// The fire/skip decision lives in one place: the solver. There is
// no static-boolean `wouldFail_<Name>` shortcut — that earlier
// invention drifted away from ivy2cpp's design and caused the
// double-`> intf.pong` bug, where assume-style preconditions
// surfaced by `ReverseImage(true, true, upd).Fmlas` but absent from
// `Update.Pre.Fmlas` were lost.

// actionGenPlan captures everything we need to emit one action
// generator. Mirrors ivy2cpp/action_gen.go:31-45.
type actionGenPlan struct {
	name           string
	structName     string
	act            goivy.Action // possibly wrapped by before_export / ext_preconds
	origAct        goivy.Action
	inputs         []*goivy.Const
	oldPreClauses  *goivy.Clauses
	paramDefs      []goivy.Expr
	preFmla        goivy.Expr
	used           *goivy.InsMap[goivy.NodeKey, goivy.Expr]
	fallback       bool
	fallbackReason string
}

// preDefinedNames mirrors ivy2cpp/action_gen.go:709. Returns the set
// of defining-symbol names for each Def in `clauses`.
func preDefinedNames(clauses *goivy.Clauses) map[string]bool {
	out := make(map[string]bool)
	if clauses == nil {
		return out
	}
	for _, d := range clauses.Defs {
		def := d.Defines()
		if c, ok := def.(*goivy.Const); ok {
			out[c.Name] = true
		}
	}
	return out
}

// preUsedContains mirrors ivy2cpp/action_gen.go:725.
func preUsedContains(used *goivy.InsMap[goivy.NodeKey, goivy.Expr], name string) bool {
	if used == nil {
		return false
	}
	for _, sym := range used.All() {
		if c, ok := sym.(*goivy.Const); ok && c.Name == name {
			return true
		}
	}
	return false
}

// defedParamSet mirrors ivy2cpp/action_gen.go:738. Returns the
// NodeKey set of LHS Consts of each paramDef.
func (p *actionGenPlan) defedParamSet() map[goivy.NodeKey]bool {
	out := make(map[goivy.NodeKey]bool, len(p.paramDefs))
	for _, pd := range p.paramDefs {
		eq, ok := pd.(*goivy.Eq)
		if !ok {
			continue
		}
		if c, ok := eq.T1.(*goivy.Const); ok {
			out[goivy.Key(c)] = true
		}
	}
	return out
}

// exprRoot mirrors ivy2cpp/action_gen.go:754. Peels single-arg
// destructor applications to reach the receiver leaf.
func exprRoot(f goivy.Expr) goivy.Expr {
	for {
		ap, ok := f.(*goivy.Apply)
		if !ok || len(ap.Terms) != 1 {
			return f
		}
		f = ap.Terms[0]
	}
}

// exprAsConst mirrors ivy2cpp/action_gen.go:767.
func exprAsConst(f goivy.Expr) (*goivy.Const, bool) {
	switch n := f.(type) {
	case *goivy.Const:
		return n, true
	case *goivy.Apply:
		if c, ok := n.Func.(*goivy.Const); ok && len(n.Terms) == 0 {
			return c, true
		}
	}
	return nil, false
}

// formulaToSmtlib mirrors ivy2cpp/action_gen.go:694. Translates an
// Ivy formula into its SMT-LIB textual form via goivy.Solver +
// Translator. Useful for cross-referencing the precondition shape
// against ivy2cpp's emitted assertion strings.
//
// Returns ("true", true) for nil input, and ("", false) when the
// translator errors out.
func (g *Generator) formulaToSmtlib(fmla goivy.Expr) (string, bool) {
	if fmla == nil {
		return "true", true
	}
	solver := goivy.NewSolver(g.Mod, nil)
	defer solver.Close()
	z3expr, err := solver.FormulaToZ3(fmla)
	if err != nil {
		return "", false
	}
	return cleanSmtlib(z3expr.String()), true
}

// cleanSmtlib mirrors ivy2cpp/solver_emit.go:30. Applies Python's
// two SMT-LIB sanitization replacements so emitted text matches
// ivy2cpp's reference output.
func cleanSmtlib(s string) string {
	s = strings.ReplaceAll(s, "|!1", "!1|")
	s = strings.ReplaceAll(s, `\|`, "")
	return s
}

// stripZ3Bars mirrors ivy2cpp/initial_state.go:250. Removes the
// surrounding `|…|` quoting Z3 applies to identifiers with special
// characters.
func stripZ3Bars(s string) string {
	if len(s) >= 2 && s[0] == '|' && s[len(s)-1] == '|' {
		return s[1 : len(s)-1]
	}
	return s
}

// buildActionGenPlan mirrors ivy2cpp/action_gen.go:51 buildActionGenPlan.
// It runs the same flow: pick up any BeforeExport/ExtPreconds wrappers,
// call GetUpdateForArt, compute the reverse image, and stash the
// precondition formula.
//
// The clauses_helpers.go normalisations ivy2cpp applies
// (TrimClauses / expandFieldReferences / extractInputFields /
// extractDefinedParameters / RelevantDefinitions / VariantAxioms) are
// not yet ported on the ivy2go side — they are queued in the
// divergence audit. Without them the precondition is still the raw
// reverse image, which is sufficient for the fixtures exercised today.
func (g *Generator) buildActionGenPlan(name string, act goivy.Action) *actionGenPlan {
	plan := &actionGenPlan{
		name:       name,
		structName: "actionGen_" + goExportedName(name),
		origAct:    act,
		act:        act,
	}

	if g.Mod != nil && g.Mod.BeforeExport != nil {
		if be, ok := g.Mod.BeforeExport.Get2(name); ok && be != nil {
			plan.act = be
		}
	}
	if g.Mod != nil && g.Mod.ExtPreconds != nil {
		if pre, ok := g.Mod.ExtPreconds[name]; ok && pre != nil {
			orig := plan.act
			seq := goivy.NewSequence(goivy.NewAssumeAction(pre), exprOfAction(orig))
			seq.SetLineno(orig.GetLineno())
			goivy.CopyFormalsTo(orig, seq)
			plan.act = seq
		}
	}

	var upd *goivy.Update
	func() {
		defer func() { _ = recover() }()
		upd = goivy.GetUpdateForArt(plan.act, g.Mod, nil)
	}()
	if upd == nil {
		plan.fallback = true
		plan.fallbackReason = "GetUpdate returned nil"
		plan.inputs = plan.act.GetFormalParams()
		return plan
	}

	// pre = tr.reverse_image(true_clauses, true_clauses, upd) gives the
	// action's transition relation + entry-precondition (assume side),
	// but does NOT include internal asserts. ivy2cpp inherits the same
	// gap from ivy_to_cpp.py; we widen the precondition here so the
	// solver gates internal asserts too: conjoin NOT(AND Update.Pre.Fmlas).
	//
	// Update.Pre.Fmlas semantics: action fails iff every fmla holds.
	// Negating the conjunction gives "action does not fail" — exactly
	// the assert-side gate the test driver needs to avoid synthesising
	// inputs that crash the body.
	truePre := goivy.TrueClauses(nil)
	preClauses := goivy.ReverseImage(truePre, truePre, upd)
	revFmla := preClauses.ToFormula()

	var assertOK goivy.Expr
	switch len(upd.Pre.Fmlas) {
	case 0:
		assertOK = nil
	case 1:
		assertOK = &goivy.LogicNot{Body: upd.Pre.Fmlas[0]}
	default:
		assertOK = &goivy.LogicNot{Body: &goivy.LogicAnd{Terms: upd.Pre.Fmlas}}
	}

	switch {
	case assertOK == nil:
		plan.preFmla = revFmla
	case revFmla == nil:
		plan.preFmla = assertOK
	default:
		plan.preFmla = &goivy.LogicAnd{Terms: []goivy.Expr{revFmla, assertOK}}
	}
	plan.inputs = plan.act.GetFormalParams()
	return plan
}

// exprOfAction mirrors ivy2cpp/action_gen.go:184 exprOfAction.
func exprOfAction(a goivy.Action) goivy.Expr {
	if e, ok := a.(goivy.Expr); ok {
		return e
	}
	return nil
}

// emitActionGenStructs is the top-level entry: it walks every public
// action, builds a plan, and emits one generator per plan. Mirrors
// the per-action loop in ivy2cpp/generator.go's emitClasses /
// emitActionGenerators.
func (g *Generator) emitActionGenStructs(w *goWriter) {
	if !g.usesZ3() {
		return
	}
	if g.Ctx != nil {
		g.Ctx.AddImport("actions", g.Config.GoivyImportPath, "")
		g.Ctx.AddImport("actions", "fmt", "")
	}
	for _, name := range g.actionGenNames() {
		act, _ := g.Mod.Actions.Get2(name)
		if act == nil {
			continue
		}
		plan := g.buildActionGenPlan(name, act)
		g.emitActionGenStructDecl(w, plan)
		g.emitActionGen(w, plan)
	}
}

// emitActionGenStructDecl emits the struct declaration with one
// solver field plus one field per formal-param input. Mirrors
// ivy2cpp/action_gen.go:193 emitActionGenClassHeader +
// emitActionGenMemberDecls.
func (g *Generator) emitActionGenStructDecl(w *goWriter, plan *actionGenPlan) {
	w.linef("// %s drives solver-backed input synthesis for the %s action.", plan.structName, plan.name)
	w.linef("type %s struct {", plan.structName)
	w.line("\tsol *goivy.Solver")
	for _, p := range plan.inputs {
		if p == nil {
			continue
		}
		w.linef("\t%s %s", goActionGenFieldName(p.Name), g.goType(p.CSort))
	}
	w.line("}")
	w.blank()
}

// emitActionGen emits the constructor + generate() + execute() +
// Close() methods for one plan. Mirrors ivy2cpp/action_gen.go:245
// emitActionGen.
func (g *Generator) emitActionGen(w *goWriter, plan *actionGenPlan) {
	g.emitActionGenConstructor(w, plan)
	g.emitActionGenGenerate(w, plan)
	g.emitActionGenExecute(w, plan)
	g.emitActionGenClose(w, plan)
}

// emitActionGenConstructor emits new<structName>(state) → just
// creates the Solver. Mirrors the C++ constructor at
// ivy2cpp/action_gen.go:257 — minus the static SMT-LIB precondition
// add(), because goivy.Solver lacks push/pop semantics so we
// re-build the Clauses per generate() call instead.
func (g *Generator) emitActionGenConstructor(w *goWriter, plan *actionGenPlan) {
	w.linef("func new%s(state *%s) *%s {", plan.structName, g.StateTypeName, plan.structName)
	w.line("\t_ = state")
	w.linef("\treturn &%s{sol: newIvySolver()}", plan.structName)
	w.line("}")
	w.blank()
}

// emitActionGenGenerate emits the generate(state) bool method.
// Returns false when the precondition is UNSAT in the current state
// (action skipped); true when SAT and per-input fields are
// populated from the model. Mirrors the C++ generate() at
// ivy2cpp/action_gen.go:299.
func (g *Generator) emitActionGenGenerate(w *goWriter, plan *actionGenPlan) {
	w.linef("func (g *%s) generate(state *%s) bool {", plan.structName, g.StateTypeName)
	w.line("\t_ = state")

	if plan.fallback {
		w.linef("\t// fallback: %s", plan.fallbackReason)
		g.emitFallbackInputAssignments(w, plan)
		w.line("\treturn true")
		w.line("}")
		w.blank()
		return
	}

	// Declare per-input *goivy.Const symbols (named to match the
	// formal-param names the precondition formula references). For
	// destructor/variant records we also declare per-field symbols
	// so the solver can synthesise each field independently.
	for i, p := range plan.inputs {
		if p == nil {
			continue
		}
		sortCode, ok := g.reifySortAsGoCode(p.CSort)
		if !ok {
			sortCode = "goivy.Boolean"
		}
		w.linef("\t__in%d := goivy.NewConst(%q, %s)", i, p.Name, sortCode)
		w.linef("\t_ = __in%d", i)
		if recName, ok := g.destructorStructName(p.CSort); ok {
			for _, f := range g.destructorScalarFields(recName) {
				fSortCode, ok := g.reifySortAsGoCode(f.Sort)
				if !ok {
					fSortCode = "goivy.Boolean"
				}
				w.linef("\t__in%d_%s := goivy.NewConst(%q, %s)",
					i, goIdent(f.Name),
					fmt.Sprintf("__in%d_%s_%s", i, goIdent(p.Name), goIdent(f.Name)),
					fSortCode)
				w.linef("\t_ = __in%d_%s", i, goIdent(f.Name))
			}
		}
		if superName, ok := g.variantSuperName(p.CSort); ok {
			for _, leaf := range g.variantLeaves(superName) {
				for _, f := range leaf.Fields {
					fSortCode, ok := g.reifySortAsGoCode(f.Sort)
					if !ok {
						fSortCode = "goivy.Boolean"
					}
					w.linef("\t__in%d_%s_%s := goivy.NewConst(%q, %s)",
						i, goIdent(leaf.Name), goIdent(f.Name),
						fmt.Sprintf("__in%d_%s_%s_%s",
							i, goIdent(p.Name), goIdent(leaf.Name), goIdent(f.Name)),
						fSortCode)
					w.linef("\t_ = __in%d_%s_%s", i, goIdent(leaf.Name), goIdent(f.Name))
				}
			}
		}
	}

	// Reify the precondition formula as a runtime *goivy.Expr.
	code, ok := g.reifyExprAsGoCode(plan.preFmla, plan.inputs)
	if !ok {
		w.linef("\t// precondition unreifiable: %s", strings.ReplaceAll(plan.preFmla.String(), "\n", " "))
		g.emitFallbackInputAssignments(w, plan)
		w.line("\treturn true")
		w.line("}")
		w.blank()
		return
	}

	w.line("\tfacts := stateFactsAsClauses(state)")
	w.linef("\tpreFmla := %s", code)
	w.line("\tpreClauses := conjClauses(facts, preFmla)")

	// Per-input randomization preferences. Mirrors cpp's per-input
	// `g.randomize(__in, "<sort>")` calls in
	// ivy2cpp/action_gen.go:326 + emitPythonTestRandomizeSolver,
	// which consume one chacha8 word per input via random_range and
	// assert `__in == picked` as an alit. We assert the equalities
	// as hard constraints and retry without them on UNSAT (cpp does
	// alit-shrinking; the simpler all-or-nothing fallback still
	// preserves byte-equivalent PRNG consumption order for the common
	// case where preferences are satisfiable).
	g.emitInputPreferences(w, plan)

	w.line("\tvar modelResult *goivy.ModelResult")
	w.line("\tif g.sol != nil {")
	w.line("\t\tif len(__prefs) > 0 {")
	w.line("\t\t\tmodelResult, _ = g.sol.GetModelClauses(conjClauses(preClauses, __prefs...))")
	w.line("\t\t\tif modelResult == nil {")
	w.line("\t\t\t\tmodelResult, _ = g.sol.GetModelClauses(preClauses)")
	w.line("\t\t\t}")
	w.line("\t\t} else {")
	w.line("\t\t\tmodelResult, _ = g.sol.GetModelClauses(preClauses)")
	w.line("\t\t}")
	w.line("\t}")
	w.line("\tif modelResult == nil {")
	w.line("\t\t// UNSAT: action's precondition cannot be satisfied in the current state — skip.")
	w.line("\t\treturn false")
	w.line("\t}")

	// Extract per-input values from the model into struct fields.
	for i, p := range plan.inputs {
		if p == nil {
			continue
		}
		g.emitInputExtraction(w, i, p)
	}

	w.line("\treturn true")
	w.line("}")
	w.blank()
	g.Ctx.OnceGlobals["__need_pickinput"] = true
}

// emitInputPreferences emits the per-input random-preference loop:
// for each scalar input (bool, enum, integer-range), pick a value via
// ivyRandomRange (one chacha8 word per input — matching cpp's
// ivy_z3_gen.hpp `random_range` after the chacha8 migration) and
// append `__in_i == picked` as a preference Expr. The caller
// AND-ANDs them into the solver query; on UNSAT it retries without
// the preference list. Record / variant inputs would need per-field
// preferences; we currently emit no preferences for them (the
// solver-picked model still satisfies the precondition, just without
// preferring a chacha8-picked value).
func (g *Generator) emitInputPreferences(w *goWriter, plan *actionGenPlan) {
	w.line("\t__prefs := []goivy.Expr{}")
	for i, p := range plan.inputs {
		if p == nil {
			continue
		}
		if plan.oldPreClauses != nil {
			if _, defidx := plan.oldPreClauses.DefIdx[goivy.Key(p)]; defidx {
				// Input pinned by a def — cpp skips randomize for it
				// (action_gen.go:322 `if defidx { continue }`).
				continue
			}
		}
		g.emitOneInputPreference(w, i, p)
	}
}

// emitOneInputPreference emits the per-input preference body.
// Currently handles bool, enum, and range/integer scalars. Other
// shapes (records, variants, function-sorted) are skipped so the
// solver picks freely.
func (g *Generator) emitOneInputPreference(w *goWriter, i int, p *goivy.Const) {
	sortCode, ok := g.reifySortAsGoCode(p.CSort)
	if !ok {
		return
	}
	switch s := p.CSort.(type) {
	case *goivy.BooleanSort:
		_ = s
		w.linef("\t{")
		w.linef("\t\t__picked := ivyRandomRange(0, 1)")
		w.linef("\t\t__rhs := goivy.False")
		w.linef("\t\tif __picked == 1 { __rhs = goivy.True }")
		w.linef("\t\t_ = __in%d", i)
		w.linef("\t\t__prefs = append(__prefs, &goivy.Eq{T1: __in%d, T2: __rhs})", i)
		w.linef("\t}")
	case *goivy.LogicEnumeratedSort:
		if len(s.Extension) == 0 {
			return
		}
		extLits := make([]string, len(s.Extension))
		for k, v := range s.Extension {
			extLits[k] = fmt.Sprintf("%q", v)
		}
		w.linef("\t{")
		w.linef("\t\t__sort := %s", sortCode)
		w.linef("\t\t__names := []string{%s}", strings.Join(extLits, ", "))
		w.linef("\t\t__picked := ivyRandomRange(0, %d)", len(s.Extension)-1)
		w.linef("\t\t__prefs = append(__prefs, &goivy.Eq{T1: __in%d, T2: goivy.NewConst(__names[__picked], __sort)})", i)
		w.linef("\t}")
	default:
		card := goSortCard(g, p.CSort)
		if card <= 0 {
			return
		}
		w.linef("\t{")
		w.linef("\t\t__sort := %s", sortCode)
		w.linef("\t\t__picked := ivyRandomRange(0, %d)", card-1)
		w.linef("\t\t__prefs = append(__prefs, &goivy.Eq{T1: __in%d, T2: goivy.NewConst(strconv.FormatUint(__picked, 10), __sort)})", i)
		w.linef("\t}")
		g.Ctx.AddImport("runtime", "strconv", "")
	}
}

// emitActionGenExecute emits the execute(state) method: trace +
// invoke. Mirrors ivy2cpp/action_gen.go:552 emitActionGenExecute.
func (g *Generator) emitActionGenExecute(w *goWriter, plan *actionGenPlan) {
	w.linef("func (g *%s) execute(state *%s) {", plan.structName, g.StateTypeName)
	w.line("\t_ = g")
	w.line("\t_ = state")

	// `> name(args)` trace line.
	display := plan.name
	if idx := strings.LastIndex(display, ":"); idx >= 0 {
		display = display[idx+1:]
	}
	if len(plan.inputs) == 0 {
		w.linef("\tfmt.Fprintln(ivyTraceOut, %q)", "> "+display)
	} else {
		var fmtStr strings.Builder
		fmtStr.WriteString("> ")
		fmtStr.WriteString(display)
		fmtStr.WriteByte('(')
		for i := range plan.inputs {
			if i > 0 {
				fmtStr.WriteByte(',')
			}
			fmtStr.WriteString("%v")
		}
		fmtStr.WriteString(")\n")
		args := make([]string, 0, len(plan.inputs))
		for _, p := range plan.inputs {
			if p == nil {
				continue
			}
			args = append(args, "g."+goActionGenFieldName(p.Name))
		}
		w.linef(`	fmt.Fprintf(ivyTraceOut, %q, %s)`, fmtStr.String(), strings.Join(args, ", "))
	}

	// Invoke state.<Name>(g.in0, g.in1, …). If the action has a
	// single return value, capture it and echo `= <value>` to the
	// trace stream — mirrors ivy2cpp/action_gen.go:617
	// (`__ivy_out << "= " << callExpr << std::endl;`).
	args := make([]string, 0, len(plan.inputs))
	for _, p := range plan.inputs {
		if p == nil {
			continue
		}
		args = append(args, "g."+goActionGenFieldName(p.Name))
	}
	var returns []*goivy.Const
	if plan.origAct != nil {
		returns = plan.origAct.GetFormalReturns()
	}
	callExpr := fmt.Sprintf("state.%s(%s)", goExportedName(plan.name), strings.Join(args, ", "))
	switch len(returns) {
	case 0:
		w.linef("\t%s", callExpr)
	case 1:
		w.linef("\t__res := %s", callExpr)
		w.line(`	fmt.Fprintf(ivyTraceOut, "= %v\n", __res)`)
	default:
		// Multi-return: declare extras as zero values and pass by
		// reference. Multi-return is rarely exercised by oracle
		// fixtures; mirror cpp's emitPythonTestActionGenExecute
		// shape but skip the result echo (cpp also skips it).
		for _, r := range returns {
			if r == nil {
				continue
			}
			zero := g.goZeroValue(r.CSort)
			w.linef("\tvar %s = %s", goActionGenFieldName(r.Name), zero)
			args = append(args, goActionGenFieldName(r.Name))
		}
		w.linef("\tstate.%s(%s)", goExportedName(plan.name), strings.Join(args, ", "))
	}
	w.line("}")
	w.blank()
}

// emitActionGenClose emits the Close() method that frees the
// underlying solver. ivy2cpp's class destructor does the same via
// RAII; in Go we expose it explicitly so the test driver can defer
// it.
func (g *Generator) emitActionGenClose(w *goWriter, plan *actionGenPlan) {
	w.linef("func (g *%s) Close() error {", plan.structName)
	w.line("\tif g.sol != nil {")
	w.line("\t\terr := g.sol.Close()")
	w.line("\t\tg.sol = nil")
	w.line("\t\treturn err")
	w.line("\t}")
	w.line("\treturn nil")
	w.line("}")
	w.blank()
}

// emitInputExtraction emits code to pull the i-th input's value
// from modelResult into the corresponding struct field.
func (g *Generator) emitInputExtraction(w *goWriter, i int, p *goivy.Const) {
	field := "g." + goActionGenFieldName(p.Name)
	typeName := g.goType(p.CSort)
	card := goSortCard(g, p.CSort)
	switch {
	case typeName == "bool":
		w.linef("\t%s = pickBoolOrChoose(g.sol, modelResult, __in%d)", field, i)
	case card > 0 && goIsAnyIntegerType(g, p.CSort):
		w.linef("\t%s = %s(pickUintOrChoose(g.sol, modelResult, __in%d, %d))", field, typeName, i, card)
	default:
		if recName, ok := g.destructorStructName(p.CSort); ok {
			g.emitStructInputAssembly(w, i, typeName, recName)
			w.linef("\t%s = v%d", field, i)
			return
		}
		if superName, ok := g.variantSuperName(p.CSort); ok {
			g.emitVariantInputAssembly(w, i, typeName, superName)
			w.linef("\t%s = v%d", field, i)
			return
		}
		w.linef("\tvar v%d %s; _ = v%d", i, typeName, i)
		w.linef("\t%s = v%d", field, i)
	}
}

// emitFallbackInputAssignments emits randomized assignments for
// every input, used when the precondition isn't reifiable (we fall
// back to plain ivyChoose-style randomization, matching ivy2cpp's
// emitWeakActionGenerator at action_gen.go:514).
func (g *Generator) emitFallbackInputAssignments(w *goWriter, plan *actionGenPlan) {
	for _, p := range plan.inputs {
		if p == nil {
			continue
		}
		field := "g." + goActionGenFieldName(p.Name)
		typeName := g.goType(p.CSort)
		card := goSortCard(g, p.CSort)
		switch {
		case typeName == "bool":
			w.linef("\t%s = ivyChoose(2) == 1", field)
		case card > 0:
			w.linef("\t%s = %s(ivyChoose(%d))", field, typeName, card)
		default:
			w.linef("\tvar __fb %s; %s = __fb", typeName, field)
		}
	}
}

// goActionGenFieldName returns the struct field name for an input.
// Different from goExportedName so we don't collide with the action
// method name (`s.Hit(...)` invoked on the State while the gen
// struct also has fields named after formals — we want stable
// `In<N>` style names instead).
func goActionGenFieldName(paramName string) string {
	// Strip goivy's `__fml:` / `fml:` prefixes for readability.
	n := strings.TrimPrefix(paramName, "__fml:")
	n = strings.TrimPrefix(n, "fml:")
	return "In_" + goExportedName(n)
}

// actionGenNames returns the actions eligible for action-gen
// emission, in deterministic order. Mirrors ivy2cpp's per-isolate
// public-action selection.
func (g *Generator) actionGenNames() []string {
	if g == nil || g.Mod == nil || g.Mod.Actions == nil {
		return nil
	}
	names := make([]string, 0)
	if g.Mod.PublicActions != nil && g.Mod.PublicActions.Len() > 0 {
		for name := range g.Mod.PublicActions.All() {
			if hasPrefixAny(name, "imp__", "__") {
				continue
			}
			if _, ok := g.Mod.Actions.Get2(name); !ok {
				continue
			}
			names = append(names, name)
		}
	} else {
		for name := range g.Mod.Actions.All() {
			if hasPrefixAny(name, "ext:", "imp:", "ivy:", "imp__", "__") {
				continue
			}
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}

// ---------------------------------------------------------------------------
// reify*: emit-time helpers that translate a goivy.Expr / goivy.Sort
// into a Go source string that, at runtime, reconstructs the same
// value. ivy2cpp accomplishes the same effect by translating the
// formula to an SMT-LIB string and passing it to Z3's add(); the Go
// runtime uses goivy.Solver which takes goivy.Expr trees directly,
// so we serialise the tree as Go AST-construction code instead.
//
// These helpers are legitimately Go-specific scaffolding (no C++
// counterpart) and are flagged as such in the divergence audit.
// ---------------------------------------------------------------------------

func (g *Generator) reifyExprAsGoCode(e goivy.Expr, params []*goivy.Const) (string, bool) {
	if e == nil {
		return "nil", true
	}
	switch n := e.(type) {
	case *goivy.Const:
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
		lbCode, ok := g.reifyBoundAsGoCode(t.Lb)
		if !ok {
			return "", false
		}
		ubCode, ok := g.reifyBoundAsGoCode(t.Ub)
		if !ok {
			return "", false
		}
		return fmt.Sprintf("&goivy.RangeSort{Name: %q, Lb: %s, Ub: %s}", t.Name, lbCode, ubCode), true
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

func (g *Generator) reifyBoundAsGoCode(b goivy.NumeralOrCompiledBound) (string, bool) {
	switch v := b.(type) {
	case goivy.NumeralBound:
		sortCode := "nil"
		if v.Sort != nil {
			if c, ok := g.reifySortAsGoCode(v.Sort); ok {
				sortCode = c
			}
		}
		return fmt.Sprintf("goivy.NumeralBound{Value: %q, Sort: %s}", v.Value, sortCode), true
	case goivy.CompiledBound:
		exprCode, ok := g.reifyExprAsGoCode(v.Expr, nil)
		if !ok {
			return "", false
		}
		return fmt.Sprintf("goivy.CompiledBound{Expr: %s}", exprCode), true
	}
	return "", false
}

// requireMustHelpers flags the runtime to emit the must* facades
// (mustApply / mustNewVariable / mustNewIte / mustNewFunctionSort).
func (g *Generator) requireMustHelpers() {
	if g != nil && g.Ctx != nil {
		g.Ctx.OnceGlobals["__need_musthelpers"] = true
	}
}

// emitStructInputAssembly emits per-field pick + struct-literal
// assembly for a destructor-record param. Kept intact from the
// previous implementation pending the broader clauses_helpers port.
func (g *Generator) emitStructInputAssembly(w *goWriter, i int, typeName, recName string) {
	fields := g.destructorScalarFields(recName)
	for _, f := range fields {
		fcard := goSortCard(g, f.Sort)
		switch ft := g.goType(f.Sort); {
		case ft == "bool":
			w.linef("v%d_%s := pickBoolOrChoose(g.sol, modelResult, __in%d_%s)",
				i, goIdent(f.Name), i, goIdent(f.Name))
		case fcard > 0 && goIsAnyIntegerType(g, f.Sort):
			w.linef("v%d_%s := %s(pickUintOrChoose(g.sol, modelResult, __in%d_%s, %d))",
				i, goIdent(f.Name), ft, i, goIdent(f.Name), fcard)
		default:
			w.linef("var v%d_%s %s", i, goIdent(f.Name), ft)
			w.linef("_ = v%d_%s", i, goIdent(f.Name))
		}
	}
	assign := make([]string, 0, len(fields))
	for _, f := range fields {
		assign = append(assign, fmt.Sprintf("%s: v%d_%s", goExportedName(f.Name), i, goIdent(f.Name)))
	}
	w.linef("v%d := %s{%s}", i, typeName, strings.Join(assign, ", "))
}

// emitVariantInputAssembly emits the tag-pick + switch-by-tag + per-
// leaf constructor invocation for a variant-super-typed param.
func (g *Generator) emitVariantInputAssembly(w *goWriter, i int, typeName, superName string) {
	leaves := g.variantLeaves(superName)
	if len(leaves) == 0 {
		w.linef("var v%d %s", i, typeName)
		w.linef("_ = v%d", i)
		return
	}
	w.linef("v%d_tag := ivyChoose(%d)", i, len(leaves))
	w.linef("var v%d %s", i, typeName)
	w.linef("switch v%d_tag {", i)
	for tag, leaf := range leaves {
		w.linef("case %d:", tag)
		ctor := "New" + goExportedName(superName) + goExportedName(leaf.Name)
		if leaf.IsPlain {
			w.linef("\tv%d = %s()", i, ctor)
			continue
		}
		leafType := goExportedName(leaf.Name)
		for _, f := range leaf.Fields {
			fcard := goSortCard(g, f.Sort)
			switch ft := g.goType(f.Sort); {
			case ft == "bool":
				w.linef("\tv%d_%s_%s := pickBoolOrChoose(g.sol, modelResult, __in%d_%s_%s)",
					i, goIdent(leaf.Name), goIdent(f.Name),
					i, goIdent(leaf.Name), goIdent(f.Name))
			case fcard > 0 && goIsAnyIntegerType(g, f.Sort):
				w.linef("\tv%d_%s_%s := %s(pickUintOrChoose(g.sol, modelResult, __in%d_%s_%s, %d))",
					i, goIdent(leaf.Name), goIdent(f.Name),
					ft, i, goIdent(leaf.Name), goIdent(f.Name), fcard)
			default:
				w.linef("\tvar v%d_%s_%s %s", i, goIdent(leaf.Name), goIdent(f.Name), ft)
				w.linef("\t_ = v%d_%s_%s", i, goIdent(leaf.Name), goIdent(f.Name))
			}
		}
		assign := make([]string, 0, len(leaf.Fields))
		for _, f := range leaf.Fields {
			assign = append(assign, fmt.Sprintf("%s: v%d_%s_%s",
				goExportedName(f.Name), i, goIdent(leaf.Name), goIdent(f.Name)))
		}
		w.linef("\tv%d = %s(%s{%s})", i, ctor, leafType, strings.Join(assign, ", "))
	}
	w.line("}")
}
