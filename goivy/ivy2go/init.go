package ivy2go

import (
	"fmt"

	"github.com/glycerine/ivy/goivy"
)

// init.go: emits the State.Init() method body. Mirrors the role of
// ivy2cpp/init.go but the C++ "after-init action list" is collapsed
// into a single Go method here.

// emitInit writes:
//
//	func (s *State) Init() {
//	    // nondet initial state (per emitOneInitialState)
//	    // after-init actions, if any (M7 lands them)
//	}
func (g *Generator) emitInit(w *goWriter) {
	w.open(fmt.Sprintf("func (s *%s) Init() {", g.StateTypeName))
	g.emitOneInitialState(w)
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

// initialConditionActions mirrors ivy2cpp/init.go initialConditionActions
// (line 9). Walks the module's InitCond formulas and lifts each into
// an Action that sets the mutable state symbol(s) the formula
// constrains. Returns an empty list when there is no InitCond.
func (g *Generator) initialConditionActions() ([]goivy.Action, error) {
	if g == nil || g.Mod == nil || g.Mod.InitCond == nil || g.Mod.InitCond.IsTrue() {
		return nil, nil
	}
	var actions []goivy.Action
	for _, f := range g.Mod.InitCond.Fmlas {
		acts, err := g.initialConditionActionsFor(f)
		if err != nil {
			return nil, err
		}
		actions = append(actions, acts...)
	}
	return actions, nil
}

// initialConditionActionsFor mirrors ivy2cpp/init.go:24. Recursively
// flattens ForAll / LogicAnd wrappers and lifts each leaf formula
// into its initial-state assignment action.
func (g *Generator) initialConditionActionsFor(f goivy.Expr) ([]goivy.Action, error) {
	switch n := f.(type) {
	case *goivy.ForAll:
		return g.initialConditionActionsFor(n.Body)
	case *goivy.LogicAnd:
		var actions []goivy.Action
		for _, term := range n.Terms {
			acts, err := g.initialConditionActionsFor(term)
			if err != nil {
				return nil, err
			}
			actions = append(actions, acts...)
		}
		return actions, nil
	default:
		act, err := g.initialConditionAction(f)
		if err != nil {
			return nil, err
		}
		return []goivy.Action{act}, nil
	}
}

// initialConditionAction mirrors ivy2cpp/init.go:47. Lifts a single
// initial-state formula leaf into the AssignAction (or SetAction)
// that sets the mutable state symbol to the formula's value.
func (g *Generator) initialConditionAction(f goivy.Expr) (goivy.Action, error) {
	switch n := f.(type) {
	case *goivy.Eq:
		if g.isStateTarget(n.T1) {
			return goivy.NewAssignAction(n.T1, n.T2), nil
		}
		if g.isStateTarget(n.T2) {
			return goivy.NewAssignAction(n.T2, n.T1), nil
		}
		return nil, fmt.Errorf("neither side of initial equality is mutable state: %s", f)
	case *goivy.LogicIff:
		if g.isStateTarget(n.T1) {
			return goivy.NewAssignAction(n.T1, n.T2), nil
		}
		if g.isStateTarget(n.T2) {
			return goivy.NewAssignAction(n.T2, n.T1), nil
		}
		return nil, fmt.Errorf("neither side of initial equivalence is mutable state: %s", f)
	case *goivy.LogicLiteral:
		if !g.isStateTarget(n.Atom) {
			return nil, fmt.Errorf("initial literal is not mutable state: %s", n.Atom)
		}
		return goivy.NewSetAction(n), nil
	case *goivy.LogicNot:
		if !g.isStateTarget(n.Body) {
			return nil, fmt.Errorf("initial negation is not mutable state: %s", n.Body)
		}
		return goivy.NewSetAction(n), nil
	case *goivy.Const, *goivy.Apply:
		if !g.isStateTarget(f) {
			return nil, fmt.Errorf("initial literal is not mutable state: %s", f)
		}
		return goivy.NewSetAction(f), nil
	default:
		return nil, fmt.Errorf("unsupported initial formula %T: %s", f, f.String())
	}
}

// isStateTarget mirrors ivy2cpp/init.go:85. Returns true when e
// references a mutable state symbol (relation or function), not a
// constructor / destructor sort name / pure definition.
func (g *Generator) isStateTarget(e goivy.Expr) bool {
	sym := stateTargetSymbol(e)
	if sym == nil || g == nil || g.Mod == nil {
		return false
	}
	name := sym.Name
	if g.Mod.Sig != nil && g.Mod.Sig.Constructors[name] {
		return false
	}
	if g.isSortConstructorName(name) {
		return false
	}
	if g.Mod.DestructorSorts != nil {
		if _, ok := g.Mod.DestructorSorts[name]; ok {
			return false
		}
	}
	if g.Mod.Relations != nil {
		if _, ok := g.Mod.Relations.Get2(goivy.Key(sym)); ok {
			return true
		}
	}
	if g.Mod.Functions != nil {
		if _, ok := g.Mod.Functions.Get2(goivy.Key(sym)); ok {
			return true
		}
	}
	return false
}

// stateTargetSymbol mirrors ivy2cpp/init.go:115. Walks past
// LogicLiteral / LogicNot / Apply wrappers and returns the
// underlying *Const symbol that the expression references.
func stateTargetSymbol(e goivy.Expr) *goivy.Const {
	switch n := e.(type) {
	case *goivy.Apply:
		if c, ok := n.Func.(*goivy.Const); ok {
			return c
		}
		return nil
	case *goivy.LogicLiteral:
		return stateTargetSymbol(n.Atom)
	case *goivy.LogicNot:
		return stateTargetSymbol(n.Body)
	case *goivy.Const:
		return n
	default:
		return nil
	}
}
