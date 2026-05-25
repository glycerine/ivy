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
func (g *Generator) emitOneActionGenStruct(w *goWriter, name string) {
	structName := "actionGen_" + goExportedName(name)
	w.linef("// %s drives solver-backed input synthesis for the %s action.", structName, name)
	w.linef("// M9 skeleton: solver is constructed but the Generate method")
	w.linef("// falls back to ivyChoose for input selection. Real solver-")
	w.linef("// driven input selection lands in M9.1+.")
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

	// Generate method: random fallback for input picking, then
	// dispatches the action. Per-sort parameter selection uses the
	// same parseArg / random helpers the REPL would use.
	act, _ := g.Mod.Actions.Get2(name)
	if act == nil {
		return
	}
	params := act.GetFormalParams()
	method := "Generate"
	w.open(fmt.Sprintf("func (g *%s) %s(state *%s) {", structName, method, g.StateTypeName))
	// Pick each input via the per-sort random helper.
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
			w.linef("v%d := ivyChoose(2) == 1", i)
		case card > 0:
			w.linef("v%d := %s(ivyChoose(%d))", i, typeName, card)
		default:
			w.linef("v%d := %s(ivyChoose(2))", i, typeName)
		}
		callArgs[i] = fmt.Sprintf("v%d", i)
	}
	w.linef("state.%s(%s)", goExportedName(name), strings.Join(callArgs, ", "))
	w.close("")
	w.blank()
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
