package ivy2golang

import (
	"fmt"

	"github.com/glycerine/ivy/goivy"
)

// extensionalRels returns the cached set of extensional relation names.
// It mirrors ivy2cpp's per-generator cache for Python's
// `the_extensional_relations`.
func (g *Generator) extensionalRels() map[string]bool {
	if g == nil {
		return nil
	}
	if g.extRel != nil {
		return g.extRel
	}
	g.extRel = g.extensionalRelations()
	if g.extRel == nil {
		g.extRel = map[string]bool{}
	}
	return g.extRel
}

type sparseSupportSummary struct {
	PreserveOld bool
	Points      [][]goivy.Expr
}

func (g *Generator) quantifierSupportRels() map[string]bool {
	res := map[string]bool{}
	for name := range g.extensionalRels() {
		res[name] = true
	}
	for name := range g.sparseSupportRels() {
		res[name] = true
	}
	return res
}

func (g *Generator) actionQuantifierSupportRels() map[string]bool {
	return g.quantifierSupportRels()
}

func (g *Generator) localWitnessSupportRelation(name string, fs *goivy.LogicFunctionSort) bool {
	if name == "" {
		return false
	}
	if g.quantifierSupportRels()[name] {
		return true
	}
	sort, ok := g.isStateSymbolName(name)
	if !ok {
		return false
	}
	stateFS, ok := sort.(*goivy.LogicFunctionSort)
	if !ok || len(stateFS.Domain()) == 0 || !isBooleanSort(stateFS.Range()) {
		return false
	}
	if fs != nil && (len(fs.Domain()) != len(stateFS.Domain()) || !isBooleanSort(fs.Range())) {
		return false
	}
	return g.goFunctionStorageFor(stateFS.Domain(), stateFS.Range()).Large
}

func (g *Generator) sparseSupportRels() map[string]bool {
	if g == nil {
		return nil
	}
	if g.supportRel != nil {
		return g.supportRel
	}
	g.supportRel = g.sparseSupportRelations()
	if g.supportRel == nil {
		g.supportRel = map[string]bool{}
	}
	return g.supportRel
}

func (g *Generator) sparseSupportRelations() map[string]bool {
	if g == nil || g.Mod == nil {
		return nil
	}
	candidates := map[string]bool{}
	for _, sym := range g.stateSymbols() {
		fs, ok := sym.Sort.(*goivy.LogicFunctionSort)
		if !ok || len(fs.Domain()) == 0 || !isBooleanSort(fs.Range()) {
			continue
		}
		if g.goFunctionStorageFor(fs.Domain(), fs.Range()).Large {
			candidates[sym.Name] = true
		}
	}
	established := map[string]bool{}
	bad := map[string]bool{}

	checkAssign := func(a *goivy.LogicAssignAction, init bool) {
		name, summary, ok, touched := g.classifySparseSupportAssign(a, candidates)
		if !touched {
			return
		}
		if !ok {
			bad[name] = true
			return
		}
		if init && !summary.PreserveOld {
			established[name] = true
		}
	}
	for _, ini := range g.Mod.Initializers {
		if ini.Action == nil {
			continue
		}
		for _, sub := range ini.Action.IterSubactions() {
			if a, ok := sub.(*goivy.LogicAssignAction); ok {
				checkAssign(a, true)
			}
		}
	}
	if len(g.Mod.Initializers) == 0 && g.Mod.Actions != nil && g.Mod.Mixins != nil {
		for _, mixin := range g.Mod.Mixins.Get("init") {
			if act, ok := g.Mod.Actions.Get2(mixin.Mixer()); ok {
				if action, ok := act.(goivy.Action); ok {
					for _, sub := range action.IterSubactions() {
						if a, ok := sub.(*goivy.LogicAssignAction); ok {
							checkAssign(a, true)
						}
					}
				}
			}
		}
	}
	if g.Mod.Actions != nil {
		for _, action := range g.Mod.Actions.All() {
			if action == nil {
				continue
			}
			for _, sub := range action.IterSubactions() {
				if a, ok := sub.(*goivy.LogicAssignAction); ok {
					checkAssign(a, false)
				}
			}
		}
	}

	res := map[string]bool{}
	for name := range candidates {
		if established[name] && !bad[name] {
			res[name] = true
		}
	}
	return res
}

func (g *Generator) classifySparseSupportAssign(a *goivy.LogicAssignAction, candidates map[string]bool) (string, sparseSupportSummary, bool, bool) {
	if a == nil {
		return "", sparseSupportSummary{}, false, false
	}
	app, ok := a.LHS.(*goivy.Apply)
	if !ok {
		return "", sparseSupportSummary{}, false, false
	}
	name := goivy.ExprName(app.Func)
	if name == "" || !candidates[name] {
		return "", sparseSupportSummary{}, false, false
	}
	if allArgsNonVariable(app.Terms) {
		return name, sparseSupportSummary{}, true, true
	}
	summary, ok := g.sparseSupportAssignment(name, app.Terms, a.RHS)
	return name, summary, ok, true
}

func (g *Generator) sparseSupportAssignment(name string, lhsTerms []goivy.Expr, rhs goivy.Expr) (sparseSupportSummary, bool) {
	vars, ok := supportVarsForTerms(lhsTerms)
	if !ok {
		return sparseSupportSummary{}, false
	}
	return sparseSupportExpr(name, vars, rhs)
}

func sparseSupportExpr(name string, vars []*goivy.LogicVariable, expr goivy.Expr) (sparseSupportSummary, bool) {
	if goivy.IsFalse(expr) {
		return sparseSupportSummary{}, true
	}
	if relationAppMatchesVars(expr, name, vars) {
		return sparseSupportSummary{PreserveOld: true}, true
	}
	if point, ok := pointForVars(vars, expr); ok {
		return sparseSupportSummary{Points: [][]goivy.Expr{point}}, true
	}
	if or, ok := expr.(*goivy.LogicOr); ok {
		var summary sparseSupportSummary
		for _, term := range or.Terms {
			part, ok := sparseSupportExpr(name, vars, term)
			if !ok {
				return sparseSupportSummary{}, false
			}
			if part.PreserveOld {
				summary.PreserveOld = true
			}
			summary.Points = append(summary.Points, part.Points...)
		}
		return summary, true
	}
	return sparseSupportSummary{}, false
}

func supportVarsForTerms(terms []goivy.Expr) ([]*goivy.LogicVariable, bool) {
	vars := make([]*goivy.LogicVariable, len(terms))
	seen := map[string]bool{}
	for i, term := range terms {
		v, ok := term.(*goivy.LogicVariable)
		if !ok || v == nil || seen[v.Name] {
			return nil, false
		}
		seen[v.Name] = true
		vars[i] = v
	}
	return vars, true
}

func relationAppMatchesVars(expr goivy.Expr, name string, vars []*goivy.LogicVariable) bool {
	app, ok := expr.(*goivy.Apply)
	if !ok || goivy.ExprName(app.Func) != name || len(app.Terms) != len(vars) {
		return false
	}
	for i, term := range app.Terms {
		v, ok := term.(*goivy.LogicVariable)
		if !ok || v == nil || vars[i] == nil || v.Name != vars[i].Name {
			return false
		}
	}
	return true
}

func pointForVars(vars []*goivy.LogicVariable, expr goivy.Expr) ([]goivy.Expr, bool) {
	terms := []goivy.Expr{expr}
	if and, ok := expr.(*goivy.LogicAnd); ok {
		terms = and.Terms
	}
	values := make([]goivy.Expr, len(vars))
	seen := make([]bool, len(vars))
	for _, term := range terms {
		idx, value, ok := pointEqForVars(vars, term)
		if !ok || seen[idx] {
			return nil, false
		}
		values[idx] = value
		seen[idx] = true
	}
	for _, ok := range seen {
		if !ok {
			return nil, false
		}
	}
	return values, true
}

func pointEqForVars(vars []*goivy.LogicVariable, expr goivy.Expr) (int, goivy.Expr, bool) {
	eq, ok := expr.(*goivy.Eq)
	if !ok || eq == nil {
		return 0, nil, false
	}
	if idx, ok := varIndex(vars, eq.T1); ok && !mentionsVars(eq.T2, vars) {
		return idx, eq.T2, true
	}
	if idx, ok := varIndex(vars, eq.T2); ok && !mentionsVars(eq.T1, vars) {
		return idx, eq.T1, true
	}
	return 0, nil, false
}

func varIndex(vars []*goivy.LogicVariable, expr goivy.Expr) (int, bool) {
	v, ok := expr.(*goivy.LogicVariable)
	if !ok || v == nil {
		return 0, false
	}
	for i, candidate := range vars {
		if candidate != nil && candidate.Name == v.Name {
			return i, true
		}
	}
	return 0, false
}

func mentionsVars(expr goivy.Expr, vars []*goivy.LogicVariable) bool {
	for _, found := range goivy.VariablesAstList(expr) {
		for _, v := range vars {
			if found != nil && v != nil && found.Name == v.Name {
				return true
			}
		}
	}
	return false
}

// extensionalRelations mirrors Python ivy_to_cpp.py:44-76, via the trusted
// ivy2cpp port. A relation is extensional when it is state, initialized to
// all-false, and only updated to false or at concrete points.
func (g *Generator) extensionalRelations() map[string]bool {
	if g == nil || g.Mod == nil {
		return nil
	}

	bad := map[string]bool{}
	if g.Mod.Actions != nil {
		for _, action := range g.Mod.Actions.All() {
			if action == nil {
				continue
			}
			for _, sub := range action.IterSubactions() {
				g.markBadExtensional(sub, bad)
			}
		}
	}

	inited := map[string]bool{}
	for _, ini := range g.Mod.Initializers {
		g.collectInitedExtensional(ini.Action, inited)
	}
	if len(g.Mod.Initializers) == 0 && g.Mod.Actions != nil && g.Mod.Mixins != nil {
		for _, mixin := range g.Mod.Mixins.Get("init") {
			if act, ok := g.Mod.Actions.Get2(mixin.Mixer()); ok {
				if a, ok := act.(goivy.Action); ok {
					g.collectInitedExtensional(a, inited)
				}
			}
		}
	}

	res := map[string]bool{}
	for _, sym := range g.allStateSymbols() {
		if inited[sym.Name] && !bad[sym.Name] {
			res[sym.Name] = true
		}
	}
	return res
}

func (g *Generator) markBadExtensional(sub goivy.Action, bad map[string]bool) {
	switch a := sub.(type) {
	case *goivy.LogicAssignAction:
		name := repName(a.LHS)
		if name == "" || g.isDestructorSort(name) {
			return
		}
		if !goivy.IsFalse(a.RHS) && !allArgsNonVariable(argsOf(a.LHS)) {
			bad[name] = true
		}
	case *goivy.LogicHavocAction:
		name := repName(a.Target)
		if name == "" || g.isDestructorSort(name) {
			return
		}
		if !allArgsNonVariable(argsOf(a.Target)) {
			bad[name] = true
		}
	case *goivy.LogicCallAction:
		for _, ret := range a.ActualReturns {
			name := repName(ret)
			if name == "" || g.isDestructorSort(name) {
				continue
			}
			if !allArgsNonVariable(argsOf(ret)) {
				bad[name] = true
			}
		}
	}
}

func (g *Generator) collectInitedExtensional(action goivy.Action, inited map[string]bool) {
	if action == nil {
		return
	}
	switch a := action.(type) {
	case *goivy.LogicSequence:
		for _, elem := range a.Elems {
			if act, ok := elem.(goivy.Action); ok {
				g.collectInitedExtensional(act, inited)
			}
		}
	case *goivy.LogicAssignAction:
		name := repName(a.LHS)
		if name == "" || g.isDestructorSort(name) {
			return
		}
		if goivy.IsFalse(a.RHS) && allArgsVariable(argsOf(a.LHS)) {
			inited[name] = true
		}
	}
}

func (g *Generator) isDestructorSort(name string) bool {
	if g == nil || g.Mod == nil || g.Mod.DestructorSorts == nil {
		return false
	}
	_, ok := g.Mod.DestructorSorts[name]
	return ok
}

func repName(e goivy.Expr) string {
	switch n := e.(type) {
	case *goivy.Const:
		return n.Name
	case *goivy.Apply:
		return goivy.ExprName(n.Func)
	default:
		return ""
	}
}

func argsOf(e goivy.Expr) []goivy.Expr {
	if app, ok := e.(*goivy.Apply); ok {
		return app.Terms
	}
	return nil
}

func allArgsNonVariable(args []goivy.Expr) bool {
	for _, a := range args {
		if _, isVar := a.(*goivy.LogicVariable); isVar {
			return false
		}
	}
	return true
}

// emitExtensionalRelationClear handles the Go map equivalent of the C++
// hash_thunk all-false reset path: `r(X, ...) := false` becomes a fresh empty
// map because absent bool relation entries read back as false.
func (g *Generator) emitExtensionalRelationClear(w *goWriter, a *goivy.LogicAssignAction) bool {
	if a == nil {
		return false
	}
	app, ok := a.LHS.(*goivy.Apply)
	if !ok {
		return false
	}
	if !allArgsVariable(app.Terms) || !goivy.IsFalse(a.RHS) {
		return false
	}
	name := goivy.ExprName(app.Func)
	if name == "" {
		return false
	}
	sort, ok := g.isStateSymbolName(name)
	if !ok {
		return false
	}
	fs, ok := sort.(*goivy.LogicFunctionSort)
	if !ok || len(fs.Domain()) == 0 || !isBooleanSort(fs.Range()) {
		return false
	}
	st := g.goFunctionStorageFor(fs.Domain(), fs.Range())
	if st.Large {
		w.linef("ivy.%s = %s", goName(name), g.goFunctionStorageInit(fs.Domain(), fs.Range()))
	} else {
		w.linef("ivy.%s = %s{}", goName(name), st.Type)
	}
	return true
}

func allArgsVariable(args []goivy.Expr) bool {
	for _, a := range args {
		if _, isVar := a.(*goivy.LogicVariable); !isVar {
			return false
		}
	}
	return true
}

func isBooleanSort(s goivy.Sort) bool {
	if _, ok := s.(*goivy.BooleanSort); ok {
		return true
	}
	return goivy.SortEqual(s, goivy.Boolean)
}

type derivedDefinition struct {
	Name   string
	Head   goivy.Expr
	Params []goivy.Expr
	RHS    goivy.Expr
	Sort   goivy.Sort
}

func (g *Generator) derivedDefinitions() []derivedDefinition {
	if g == nil || g.Mod == nil {
		return nil
	}
	defs := make([]derivedDefinition, 0, len(g.Mod.Definitions))
	for _, lf := range g.Mod.Definitions {
		def, ok := lf.Formula.(*goivy.LogicDefinition)
		if !ok || def == nil {
			continue
		}
		dd, ok := newDerivedDefinition(def)
		if ok {
			defs = append(defs, dd)
		}
	}
	return defs
}

func (g *Generator) nativeDefinitions() []derivedDefinition {
	if g == nil || g.Mod == nil {
		return nil
	}
	defs := make([]derivedDefinition, 0, len(g.Mod.NativeDefinitions))
	for _, lf := range g.Mod.NativeDefinitions {
		def, ok := lf.Formula.(*goivy.LogicDefinition)
		if !ok || def == nil {
			continue
		}
		dd, ok := newDerivedDefinition(def)
		if ok {
			defs = append(defs, dd)
		}
	}
	return defs
}

func (g *Generator) allDefinitions() []derivedDefinition {
	defs := g.derivedDefinitions()
	return append(defs, g.nativeDefinitions()...)
}

func newDerivedDefinition(def *goivy.LogicDefinition) (derivedDefinition, bool) {
	if def == nil || def.Lhs == nil || def.Rhs == nil {
		return derivedDefinition{}, false
	}
	name := goivy.ExprName(def.Defines())
	if name == "" {
		return derivedDefinition{}, false
	}
	var params []goivy.Expr
	if app, ok := def.Lhs.(*goivy.Apply); ok {
		params = append(params, app.Terms...)
	}
	return derivedDefinition{
		Name:   name,
		Head:   def.Defines(),
		Params: params,
		RHS:    def.Rhs,
		Sort:   def.Lhs.NodeSort(),
	}, true
}

func (g *Generator) definitionByName(name string) (derivedDefinition, bool) {
	if name == "" {
		return derivedDefinition{}, false
	}
	for _, d := range g.allDefinitions() {
		if d.Name == name {
			return d, true
		}
	}
	return derivedDefinition{}, false
}

func (g *Generator) derivedDefinitionByName(name string) (derivedDefinition, bool) {
	if name == "" {
		return derivedDefinition{}, false
	}
	for _, d := range g.derivedDefinitions() {
		if d.Name == name {
			return d, true
		}
	}
	return derivedDefinition{}, false
}

func (g *Generator) nativeDefinitionByName(name string) (derivedDefinition, bool) {
	if name == "" {
		return derivedDefinition{}, false
	}
	for _, d := range g.nativeDefinitions() {
		if d.Name == name {
			return d, true
		}
	}
	return derivedDefinition{}, false
}

func (g *Generator) isDefinitionName(name string) bool {
	_, ok := g.definitionByName(name)
	return ok
}

func (g *Generator) isNativeDefinitionName(name string) bool {
	_, ok := g.nativeDefinitionByName(name)
	return ok
}

func (g *Generator) expandDefinitionApplication(name string, terms []goivy.Expr) (goivy.Expr, bool, error) {
	def, ok := g.derivedDefinitionByName(name)
	if !ok {
		return nil, false, nil
	}
	if len(def.Params) != len(terms) {
		return nil, false, nil
	}
	if g.defStack == nil {
		g.defStack = map[string]bool{}
	}
	if g.defStack[name] {
		return nil, true, fmt.Errorf("ivy2golang: recursive derived definition %s", name)
	}
	g.defStack[name] = true
	defer delete(g.defStack, name)

	subs := map[goivy.NodeKey]goivy.Expr{}
	for i, param := range def.Params {
		switch p := param.(type) {
		case *goivy.LogicVariable:
			subs[goivy.Key(p)] = terms[i]
		case *goivy.Const:
			subs[goivy.Key(p)] = terms[i]
		default:
			return nil, true, fmt.Errorf("ivy2golang: unsupported definition parameter %T in %s", param, name)
		}
	}
	if len(subs) == 0 {
		return def.RHS, true, nil
	}
	expr, err := goivy.Substitute(def.RHS, subs)
	if err != nil {
		return nil, true, fmt.Errorf("ivy2golang: expand definition %s: %w", name, err)
	}
	return expr, true, nil
}

func (g *Generator) expandDefinitionExprOnce(e goivy.Expr) (goivy.Expr, bool, error) {
	switch n := e.(type) {
	case *goivy.Apply:
		return g.expandDefinitionApplication(goivy.ExprName(n.Func), n.Terms)
	case *goivy.Const:
		return g.expandDefinitionApplication(n.Name, nil)
	case *goivy.LogicLiteral:
		expr, ok, err := g.expandDefinitionExprOnce(n.Atom)
		if !ok || err != nil {
			return nil, ok, err
		}
		if n.Polarity == 0 {
			return &goivy.LogicNot{Body: expr}, true, nil
		}
		return expr, true, nil
	case *goivy.LogicNot:
		expr, ok, err := g.expandDefinitionExprOnce(n.Body)
		if !ok || err != nil {
			return nil, ok, err
		}
		return &goivy.LogicNot{Body: expr}, true, nil
	default:
		return nil, false, nil
	}
}

// matchExtensionalBoundExprs mirrors ivy2cpp's port of Python
// `get_extensional_bound_exprs`, collecting extensional-relation applications
// that bind v0 in an existential polarity.
func (g *Generator) matchExtensionalBoundExprs(v0 *goivy.LogicVariable, body goivy.Expr, exists bool, res *[]*goivy.Apply) {
	if v0 == nil || body == nil {
		return
	}
	if not, ok := body.(*goivy.LogicNot); ok {
		g.matchExtensionalBoundExprs(v0, not.Body, !exists, res)
		return
	}
	if lit, ok := body.(*goivy.LogicLiteral); ok {
		nextExists := exists
		if lit.Polarity == 0 {
			nextExists = !exists
		}
		g.matchExtensionalBoundExprs(v0, lit.Atom, nextExists, res)
		return
	}
	app, isApp := body.(*goivy.Apply)
	if isApp {
		name := goivy.ExprName(app.Func)
		if name != "" && g.actionQuantifierSupportRels()[name] && exists && containsVariableByName(app.Terms, v0.Name) {
			*res = append(*res, app)
		}
	}
	if imp, ok := body.(*goivy.LogicImplies); ok {
		if !exists {
			g.matchExtensionalBoundExprs(v0, imp.T1, !exists, res)
			g.matchExtensionalBoundExprs(v0, imp.T2, exists, res)
		}
		return
	}
	if or, ok := body.(*goivy.LogicOr); ok {
		if !exists {
			for _, t := range or.Terms {
				g.matchExtensionalBoundExprs(v0, t, exists, res)
			}
		}
		return
	}
	if and, ok := body.(*goivy.LogicAnd); ok {
		if exists {
			for _, t := range and.Terms {
				g.matchExtensionalBoundExprs(v0, t, exists, res)
			}
		}
		return
	}
	if !isApp || !containsVariableByName(app.Terms, v0.Name) {
		return
	}
	def, ok := g.definitionByName(goivy.ExprName(app.Func))
	if !ok || def.RHS == nil {
		return
	}
	if !allArgsVariable(def.Params) || len(def.Params) != len(app.Terms) {
		return
	}
	subs := map[goivy.NodeKey]goivy.Expr{}
	for i, p := range def.Params {
		pv, isVar := p.(*goivy.LogicVariable)
		if !isVar {
			return
		}
		subs[goivy.Key(pv)] = app.Terms[i]
	}
	substituted, err := goivy.Substitute(def.RHS, subs)
	if err != nil {
		return
	}
	g.matchExtensionalBoundExprs(v0, substituted, exists, res)
}

func containsVariableByName(terms []goivy.Expr, name string) bool {
	for _, t := range terms {
		if v, ok := t.(*goivy.LogicVariable); ok && v.Name == name {
			return true
		}
	}
	return false
}
