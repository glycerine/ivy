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
		act, err := g.initialConditionAction(f)
		if err != nil {
			return nil, err
		}
		actions = append(actions, act)
	}
	return actions, nil
}

func (g *Generator) initialConditionAction(f goivy.Expr) (goivy.Action, error) {
	switch n := f.(type) {
	case *goivy.Eq:
		if !g.isStateTarget(n.T1) {
			return nil, fmt.Errorf("left side of initial equality is not mutable state: %s", n.T1)
		}
		return goivy.NewAssignAction(n.T1, n.T2), nil
	case *goivy.LogicIff:
		if !g.isStateTarget(n.T1) {
			return nil, fmt.Errorf("left side of initial equivalence is not mutable state: %s", n.T1)
		}
		return goivy.NewAssignAction(n.T1, n.T2), nil
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
	default:
		return goivy.ExprName(e)
	}
}
