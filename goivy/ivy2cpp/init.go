package ivy2cpp

import (
	"fmt"

	"github.com/glycerine/ivy/goivy"
)

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

func (g *Generator) isStateTarget(e goivy.Expr) bool {
	name := stateTargetName(e)
	if name == "" || g == nil || g.Mod == nil {
		return false
	}
	if g.Mod.Sig != nil && g.Mod.Sig.Constructors[name] {
		return false
	}
	if g.Mod.DestructorSorts != nil {
		if _, ok := g.Mod.DestructorSorts[name]; ok {
			return false
		}
	}
	if g.Mod.Relations != nil {
		if _, ok := g.Mod.Relations.Get2(name); ok {
			return true
		}
	}
	if g.Mod.Functions != nil {
		if _, ok := g.Mod.Functions.Get2(name); ok {
			return true
		}
	}
	return false
}

func stateTargetName(e goivy.Expr) string {
	switch n := e.(type) {
	case *goivy.Apply:
		return goivy.ExprName(n.Func)
	case *goivy.LogicLiteral:
		return stateTargetName(n.Atom)
	case *goivy.LogicNot:
		return stateTargetName(n.Body)
	case *goivy.Const, *goivy.LogicVariable:
		return goivy.ExprName(e)
	default:
		return ""
	}
}
