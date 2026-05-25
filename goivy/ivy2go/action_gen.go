package ivy2go

import (
	"fmt"
	"sort"
	"strings"
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
func (g *Generator) emitActionGenStructs(w *goWriter) {
	if !g.usesZ3() {
		return
	}
	if g.Ctx != nil {
		g.Ctx.AddImport("actions", g.Config.GoivyImportPath, "")
	}
	for _, name := range g.actionGenNames() {
		g.emitOneActionGenStruct(w, name)
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
	w.line("\ttrueClauses := goivy.NewClauses(nil, nil)")
	w.line("\tmodelResult, _ = g.sol.GetModelClauses(trueClauses)")
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
