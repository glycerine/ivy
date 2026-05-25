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

// emitAfterInitActions emits calls to any `after init {…}` blocks the
// Ivy module declared. M5 leaves this empty; M7 fills it in alongside
// the REPL setup.
func (g *Generator) emitAfterInitActions(w *goWriter) {
	_ = w
}
