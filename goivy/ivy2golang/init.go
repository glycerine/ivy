package ivy2golang

import (
	"fmt"

	"github.com/glycerine/ivy/goivy"
)

func (g *Generator) initialConditionActions() ([]goivy.Action, error) {
	if g == nil || g.Mod == nil {
		return nil, nil
	}
	if err := g.checkInitialStateParameters("initial condition", g.Mod.LabeledInits); err != nil {
		return nil, err
	}
	if err := g.checkInitialStateParameters("axiom", g.Mod.LabeledAxioms); err != nil {
		return nil, err
	}
	if g.Mod.InitCond == nil || g.Mod.InitCond.IsTrue() {
		return nil, nil
	}
	if err := g.checkInitialStateParameterFormulas("initial condition", g.Mod.InitCond.Fmlas); err != nil {
		return nil, err
	}
	if err := g.checkSimpleInitialStateConflicts(g.Mod.InitCond.Fmlas); err != nil {
		return nil, err
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

func (g *Generator) checkInitialStateParameters(kind string, lfs []*goivy.LabeledFormula) error {
	if g == nil || g.Mod == nil || len(g.Mod.Params) == 0 {
		return nil
	}
	params := g.initialParameterNames()
	for _, lf := range lfs {
		if lf == nil {
			continue
		}
		f, ok := lf.Formula.(goivy.Expr)
		if !ok || f == nil {
			continue
		}
		for name := range initialUsedSymbolNames([]goivy.Expr{f}) {
			if !params[name] {
				continue
			}
			line := ""
			if lf.Lineno() > 0 {
				line = fmt.Sprintf(" at line %d", lf.Lineno())
			}
			return fmt.Errorf("ivy2golang: %s%s depends on stripped parameter %q", kind, line, name)
		}
	}
	return nil
}

func (g *Generator) checkInitialStateParameterFormulas(kind string, formulas []goivy.Expr) error {
	if g == nil || g.Mod == nil || len(g.Mod.Params) == 0 {
		return nil
	}
	params := g.initialParameterNames()
	for name := range initialUsedSymbolNames(formulas) {
		if params[name] {
			return fmt.Errorf("ivy2golang: %s depends on stripped parameter %q", kind, name)
		}
	}
	return nil
}

func (g *Generator) initialParameterNames() map[string]bool {
	params := map[string]bool{}
	if g == nil || g.Mod == nil {
		return params
	}
	for _, p := range g.Mod.Params {
		if p != nil && p.Name != "" {
			params[p.Name] = true
		}
	}
	return params
}

func initialUsedSymbolNames(formulas []goivy.Expr) map[string]bool {
	used := map[string]bool{}
	for _, f := range formulas {
		if f == nil {
			continue
		}
		for _, sym := range goivy.UsedSymbolsAst(f).All() {
			name := goivy.ExprName(sym)
			if name != "" {
				used[name] = true
			}
		}
	}
	return used
}

func (g *Generator) checkSimpleInitialStateConflicts(formulas []goivy.Expr) error {
	seen := map[string]string{}
	for _, f := range formulas {
		if err := g.collectSimpleInitialAssignments(f, seen); err != nil {
			return err
		}
	}
	return nil
}

func (g *Generator) collectSimpleInitialAssignments(f goivy.Expr, seen map[string]string) error {
	if f == nil {
		return nil
	}
	if expanded, ok, err := g.expandDefinitionExprOnce(f); ok || err != nil {
		if err != nil {
			return err
		}
		return g.collectSimpleInitialAssignments(expanded, seen)
	}
	switch n := f.(type) {
	case *goivy.ForAll:
		return g.collectSimpleInitialAssignments(n.Body, seen)
	case *goivy.LogicAnd:
		for _, term := range n.Terms {
			if err := g.collectSimpleInitialAssignments(term, seen); err != nil {
				return err
			}
		}
	case *goivy.Eq:
		if g.isStateTarget(n.T1) {
			return g.recordSimpleInitialAssignment(n.T1, n.T2, seen)
		}
		if g.isStateTarget(n.T2) {
			return g.recordSimpleInitialAssignment(n.T2, n.T1, seen)
		}
	case *goivy.LogicIff:
		if g.isStateTarget(n.T1) {
			return g.recordSimpleInitialAssignment(n.T1, n.T2, seen)
		}
		if g.isStateTarget(n.T2) {
			return g.recordSimpleInitialAssignment(n.T2, n.T1, seen)
		}
	case *goivy.LogicLiteral:
		if n.Polarity == 0 {
			return g.recordSimpleInitialAssignment(n.Atom, goivy.NewConst("false", goivy.Boolean), seen)
		}
		return g.recordSimpleInitialAssignment(n.Atom, goivy.NewConst("true", goivy.Boolean), seen)
	case *goivy.LogicNot:
		return g.recordSimpleInitialAssignment(n.Body, goivy.NewConst("false", goivy.Boolean), seen)
	case *goivy.Const, *goivy.Apply:
		return g.recordSimpleInitialAssignment(f, goivy.NewConst("true", goivy.Boolean), seen)
	}
	return nil
}

func (g *Generator) recordSimpleInitialAssignment(target, value goivy.Expr, seen map[string]string) error {
	if target == nil || value == nil || !g.isStateTarget(target) {
		return nil
	}
	if len(goivy.VariablesAstList(target)) > 0 || len(goivy.VariablesAstList(value)) > 0 {
		return nil
	}
	if g.exprReferencesState(value) {
		return nil
	}
	key := target.String()
	val := value.String()
	if prev, ok := seen[key]; ok && prev != val {
		return fmt.Errorf("ivy2golang: axioms and/or initial condition are inconsistent: %s initialized to both %s and %s", key, prev, val)
	}
	seen[key] = val
	return nil
}

func (g *Generator) exprReferencesState(e goivy.Expr) bool {
	if e == nil {
		return false
	}
	for _, sym := range goivy.UsedSymbolsAst(e).All() {
		if name := goivy.ExprName(sym); name != "" {
			if _, ok := g.isStateSymbolName(name); ok {
				return true
			}
		}
	}
	return false
}

func (g *Generator) initialConditionActionsFor(f goivy.Expr) ([]goivy.Action, error) {
	if expanded, ok, err := g.expandDefinitionExprOnce(f); ok || err != nil {
		if err != nil {
			return nil, err
		}
		return g.initialConditionActionsFor(expanded)
	}
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
