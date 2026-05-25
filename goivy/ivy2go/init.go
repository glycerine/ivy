package ivy2go

import "fmt"

// init.go: emits the State.Init() method body. Mirrors the role of
// ivy2cpp/init.go but the C++ "after-init action list" is collapsed
// into a single Go method here.

// emitInitMethod writes:
//
//	func (s *State) Init() {
//	    // nondet initial state (per emitInitialState)
//	    // after-init actions, if any (M7 lands them)
//	}
func (g *Generator) emitInitMethod(w *goWriter) {
	w.open(fmt.Sprintf("func (s *%s) Init() {", g.StateTypeName))
	g.emitInitialState(w)
	g.emitAfterInitActions(w)
	w.close("")
	w.blank()
}

// emitAfterInitActions emits the imperative `after init { … }` blocks
// the Ivy module declared. Mirrors ivy2cpp/generator.go emitInit lines
// 1224-1227: walk g.Mod.InitialActions and feed each through the
// regular action emitter. Without this, state assignments like
// `side := left; ball := true` never run, leaving every field at its
// Go zero value — and downstream guards like `if s.LeftPlayerBall`
// silently skip the body.
func (g *Generator) emitAfterInitActions(w *goWriter) {
	if g == nil || g.Mod == nil {
		return
	}
	for _, act := range g.Mod.InitialActions {
		g.emitAction(w, act)
	}
}
