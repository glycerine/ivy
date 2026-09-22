package ivy2golang

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/glycerine/ivy/goivy"
	"github.com/glycerine/ivy/goivy/smt"
)

type initialLocatedError struct {
	loc goivy.Location
	err error
}

type initialStateConstraints struct {
	Formulas []goivy.Expr
	Used     map[string]bool
}

type initialDomainValue struct {
	Expr goivy.Expr
	Go   string
}

func (e *initialLocatedError) Error() string { return e.err.Error() }
func (e *initialLocatedError) Unwrap() error { return e.err }

func withInitialErrorLocation(expr goivy.Expr, err error) error {
	if err == nil || expr == nil {
		return err
	}
	var located *initialLocatedError
	if errors.As(err, &located) {
		return err
	}
	loc := expr.GetLineno()
	if loc == (goivy.Location{}) {
		return err
	}
	return &initialLocatedError{loc: loc, err: err}
}

func initialErrorLocation(err error) goivy.Location {
	var located *initialLocatedError
	if errors.As(err, &located) {
		return located.loc
	}
	return goivy.Location{}
}

func (g *Generator) initialStateConstraints() (*initialStateConstraints, error) {
	if g == nil || g.Mod == nil {
		return &initialStateConstraints{Used: map[string]bool{}}, nil
	}
	if err := g.checkInitialStateParameters("initial condition", g.Mod.LabeledInits); err != nil {
		return nil, err
	}
	if err := g.checkInitialStateParameters("axiom", g.Mod.LabeledAxioms); err != nil {
		return nil, err
	}
	var formulas []goivy.Expr
	if g.Mod.InitCond != nil && !g.Mod.InitCond.IsTrue() {
		formulas = append(formulas, g.Mod.InitCond.Fmlas...)
	}
	formulas = append(formulas, g.Mod.Axioms()...)
	if err := g.checkInitialStateParameterFormulas("initial condition", formulas); err != nil {
		return nil, err
	}
	for _, lf := range goivy.RelevantDefinitions(g.Mod, initialUsedSymbolNames(formulas)) {
		def, ok := lf.Formula.(*goivy.LogicDefinition)
		if !ok || def == nil {
			continue
		}
		formulas = append(formulas, goivy.DefinitionToConstraint(def))
	}
	return &initialStateConstraints{
		Formulas: formulas,
		Used:     initialUsedSymbolNames(formulas),
	}, nil
}

func (g *Generator) emitSolvedInitialState(w *goWriter) error {
	constraints, err := g.initialStateConstraints()
	if err != nil {
		return err
	}
	var slv *goivy.Solver
	var model *smt.Model
	if len(constraints.Formulas) > 0 {
		slv = goivy.NewSolver(g.Mod, nil)
		mr, err := slv.GetModelClauses(goivy.NewClauses(constraints.Formulas, nil, nil))
		if err != nil {
			return fmt.Errorf("ivy2golang: initial state solver failed: %w", err)
		}
		if mr == nil || mr.Model == nil {
			return fmt.Errorf("ivy2golang: axioms and/or initial condition are inconsistent")
		}
		model = mr.Model
	}
	for _, sym := range g.stateSymbols() {
		if g.isParamName(sym.Name) {
			continue
		}
		if !constraints.Used[sym.Name] || model == nil {
			g.emitRandomizeSymbol(w, sym, "init")
			continue
		}
		if err := g.emitSolvedInitialStateSymbol(w, slv, model, sym); err != nil {
			return err
		}
	}
	return nil
}

func (g *Generator) emitSolvedInitialStateSymbol(w *goWriter, slv *goivy.Solver, model *smt.Model, sym stateSymbol) error {
	if slv == nil || model == nil {
		g.emitRandomizeSymbol(w, sym, "init")
		return nil
	}
	fs, ok := sym.Sort.(*goivy.LogicFunctionSort)
	if !ok || len(fs.Domain()) == 0 {
		term := initialStateSymbolTerm(sym, nil)
		value, err := g.initialModelValue(slv, model, term, initialStateRange(sym.Sort))
		if err != nil {
			return fmt.Errorf("ivy2golang: initial value for %s: %w", sym.Name, err)
		}
		w.linef("ivy.%s = %s", goName(sym.Name), value)
		return nil
	}
	tuples, ok := g.initialDomainTuples(fs.Domain())
	if !ok {
		return fmt.Errorf("ivy2golang: cannot enumerate initial-state domain of %s", sym.Name)
	}
	for _, tuple := range tuples {
		args := make([]goivy.Expr, len(tuple))
		goArgs := make([]string, len(tuple))
		for i, v := range tuple {
			args[i] = v.Expr
			goArgs[i] = v.Go
		}
		term := initialStateSymbolTerm(sym, args)
		value, err := g.initialModelValue(slv, model, term, fs.Range())
		if err != nil {
			return fmt.Errorf("ivy2golang: initial value for %s: %w", sym.Name, err)
		}
		if call, ok, err := g.goStorageSet(term, value); ok || err != nil {
			if err != nil {
				return fmt.Errorf("ivy2golang: initial value for %s: %w", sym.Name, err)
			}
			w.line(call)
			continue
		}
		w.linef("%s = %s", g.goStorageAccess(sym.Name, sym.Sort, goArgs, "ivy"), value)
	}
	return nil
}

func initialStateRange(s goivy.Sort) goivy.Sort {
	if fs, ok := s.(*goivy.LogicFunctionSort); ok {
		return fs.Range()
	}
	return s
}

func initialStateSymbolTerm(sym stateSymbol, args []goivy.Expr) goivy.Expr {
	c := goivy.NewConst(sym.Name, sym.Sort)
	fs, ok := sym.Sort.(*goivy.LogicFunctionSort)
	if !ok {
		return c
	}
	if len(args) == 0 && len(fs.Domain()) == 0 {
		app, err := goivy.NewApply(c)
		if err == nil {
			return app
		}
		return c
	}
	app, err := goivy.NewApply(c, args...)
	if err != nil {
		return c
	}
	return app
}

func (g *Generator) initialModelValue(slv *goivy.Solver, model *smt.Model, term goivy.Expr, s goivy.Sort) (string, error) {
	zterm, err := slv.FormulaToZ3(term)
	if err != nil {
		return "", err
	}
	value, ok := model.Eval(zterm, true)
	if !ok {
		return "", fmt.Errorf("model did not evaluate %s", term.String())
	}
	return g.modelValueToGo(value, s)
}

func (g *Generator) modelValueToGo(value smt.Z3Expr, s goivy.Sort) (string, error) {
	text := stripZ3Bars(strings.TrimSpace(value.String()))
	switch st := s.(type) {
	case *goivy.BooleanSort:
		switch text {
		case "true", "1":
			return "true", nil
		case "false", "0":
			return "false", nil
		default:
			return "", fmt.Errorf("expected bool model value, got %q", text)
		}
	case *goivy.LogicEnumeratedSort:
		for i, name := range st.Extension {
			if text == name || text == fmt.Sprintf("%s_%d", initialZ3SortName(st), i) {
				return goName(name), nil
			}
		}
		if idx, ok := parseTrailingModelIndex(text); ok && idx >= 0 && idx < len(st.Extension) {
			return goName(st.Extension[idx]), nil
		}
		return "", fmt.Errorf("expected %s model value, got %q", sortName(s), text)
	default:
		if n, err := strconv.ParseInt(text, 10, 64); err == nil {
			return strconv.FormatInt(n, 10), nil
		}
		if idx, ok := parseTrailingModelIndex(text); ok {
			return strconv.Itoa(idx), nil
		}
		return "", fmt.Errorf("unsupported model value %q for sort %s", text, sortName(s))
	}
}

func stripZ3Bars(s string) string {
	if len(s) >= 2 && s[0] == '|' && s[len(s)-1] == '|' {
		return s[1 : len(s)-1]
	}
	return s
}

func parseTrailingModelIndex(s string) (int, bool) {
	pos := strings.LastIndexByte(s, '_')
	if pos < 0 || pos+1 >= len(s) {
		return 0, false
	}
	n, err := strconv.Atoi(s[pos+1:])
	if err != nil {
		return 0, false
	}
	return n, true
}

func (g *Generator) initialDomainTuples(domain []goivy.Sort) ([][]initialDomainValue, bool) {
	tuples := [][]initialDomainValue{{}}
	for _, s := range domain {
		vals, ok := g.initialDomainValues(s)
		if !ok {
			return nil, false
		}
		var next [][]initialDomainValue
		for _, tuple := range tuples {
			for _, v := range vals {
				cp := append([]initialDomainValue{}, tuple...)
				cp = append(cp, v)
				next = append(next, cp)
			}
		}
		tuples = next
	}
	return tuples, true
}

func (g *Generator) initialDomainValues(s goivy.Sort) ([]initialDomainValue, bool) {
	switch st := s.(type) {
	case *goivy.BooleanSort:
		return []initialDomainValue{
			{Expr: goivy.NewConst("false", goivy.Boolean), Go: "false"},
			{Expr: goivy.NewConst("true", goivy.Boolean), Go: "true"},
		}, true
	case *goivy.LogicEnumeratedSort:
		vals := make([]initialDomainValue, 0, len(st.Extension))
		for _, name := range st.Extension {
			vals = append(vals, initialDomainValue{
				Expr: goivy.NewConst(name, st),
				Go:   goName(name),
			})
		}
		return vals, true
	default:
		if rs, ok := g.rangeSortFor(s); ok {
			lo, hi, ok := numericRangeBounds(rs)
			if !ok || hi < lo {
				return nil, false
			}
			vals := make([]initialDomainValue, 0, hi-lo+1)
			for i := lo; i <= hi; i++ {
				text := strconv.Itoa(i)
				vals = append(vals, initialDomainValue{
					Expr: goivy.NewConst(text, s),
					Go:   text,
				})
			}
			return vals, true
		}
		if card := g.sortCard(s); card > 0 {
			vals := make([]initialDomainValue, 0, card)
			for i := 0; i < card; i++ {
				text := strconv.Itoa(i)
				vals = append(vals, initialDomainValue{
					Expr: goivy.NewConst(text, s),
					Go:   text,
				})
			}
			return vals, true
		}
		return nil, false
	}
}

func initialZ3SortName(s goivy.Sort) string {
	switch st := s.(type) {
	case *goivy.BooleanSort:
		return "bool"
	case *goivy.LogicEnumeratedSort:
		if st.Name == "" {
			return "int"
		}
		return st.Name
	case *goivy.RangeSort:
		if st.Name == "" {
			return "int"
		}
		return st.Name
	case *goivy.UninterpretedSort:
		if st.Name == "" {
			return "int"
		}
		return st.Name
	default:
		return sortName(s)
	}
}

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
	var formulas []goivy.Expr
	if g.Mod.InitCond != nil && !g.Mod.InitCond.IsTrue() {
		formulas = append(formulas, g.Mod.InitCond.Fmlas...)
	}
	axioms := g.Mod.Axioms()
	allFormulas := append(append([]goivy.Expr{}, formulas...), axioms...)
	if g.Config.Target == "gen" {
		if err := g.checkGenInitialQuantifierDomains(g.genInitialConstraintFormulas(allFormulas)); err != nil {
			return nil, err
		}
	}
	if err := g.checkInitialStateParameterFormulas("initial condition", formulas); err != nil {
		return nil, err
	}
	if err := g.checkSimpleInitialStateConflicts(allFormulas); err != nil {
		return nil, err
	}
	var actions []goivy.Action
	for _, f := range formulas {
		acts, err := g.initialConditionActionsFor(f)
		if err != nil {
			return nil, withInitialErrorLocation(f, err)
		}
		actions = append(actions, acts...)
	}
	for _, f := range axioms {
		actions = append(actions, g.initialAxiomActionsFor(f)...)
	}
	return actions, nil
}

func (g *Generator) genInitialConstraintFormulas(formulas []goivy.Expr) []goivy.Expr {
	res := append([]goivy.Expr{}, formulas...)
	if g == nil || g.Mod == nil {
		return res
	}
	for _, lf := range goivy.RelevantDefinitions(g.Mod, initialUsedSymbolNames(res)) {
		def, ok := lf.Formula.(*goivy.LogicDefinition)
		if !ok || def == nil {
			continue
		}
		res = append(res, goivy.DefinitionToConstraint(def))
	}
	return res
}

func (g *Generator) checkGenInitialQuantifierDomains(formulas []goivy.Expr) error {
	for _, f := range formulas {
		if err := g.checkGenInitialQuantifierDomainsInExpr(f); err != nil {
			return withInitialErrorLocation(f, err)
		}
	}
	return nil
}

func (g *Generator) checkGenInitialQuantifierDomainsInExpr(f goivy.Expr) error {
	if f == nil {
		return nil
	}
	if expanded, ok, err := g.expandDefinitionExprOnce(f); ok || err != nil {
		if err != nil {
			return err
		}
		return g.checkGenInitialQuantifierDomainsInExpr(expanded)
	}
	switch n := f.(type) {
	case *goivy.ForAll:
		for _, v := range n.Variables {
			if _, ok := g.initialFiniteValueTerms(v.VSort); !ok {
				return fmt.Errorf("ivy2golang: cannot enumerate quantified initial-state variable %s:%s", v.Name, sortName(v.VSort))
			}
		}
		return g.checkGenInitialQuantifierDomainsInExpr(n.Body)
	case *goivy.LogicAnd:
		for _, term := range n.Terms {
			if err := g.checkGenInitialQuantifierDomainsInExpr(term); err != nil {
				return err
			}
		}
	case *goivy.LogicOr:
		for _, term := range n.Terms {
			if err := g.checkGenInitialQuantifierDomainsInExpr(term); err != nil {
				return err
			}
		}
	case *goivy.LogicImplies:
		if err := g.checkGenInitialQuantifierDomainsInExpr(n.T1); err != nil {
			return err
		}
		return g.checkGenInitialQuantifierDomainsInExpr(n.T2)
	case *goivy.LogicIff:
		if err := g.checkGenInitialQuantifierDomainsInExpr(n.T1); err != nil {
			return err
		}
		return g.checkGenInitialQuantifierDomainsInExpr(n.T2)
	case *goivy.LogicNot:
		return g.checkGenInitialQuantifierDomainsInExpr(n.Body)
	case *goivy.LogicLiteral:
		return g.checkGenInitialQuantifierDomainsInExpr(n.Atom)
	}
	return nil
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
			return withInitialErrorLocation(f, err)
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
			if target, value, ok := g.initialNegatedEqualityAssignment(n.Atom); ok {
				return g.recordSimpleInitialAssignment(target, value, seen)
			}
		}
		if n.Polarity == 0 {
			return g.recordSimpleInitialAssignment(n.Atom, goivy.NewConst("false", goivy.Boolean), seen)
		}
		return g.recordSimpleInitialAssignment(n.Atom, goivy.NewConst("true", goivy.Boolean), seen)
	case *goivy.LogicNot:
		if target, value, ok := g.initialNegatedEqualityAssignment(n.Body); ok {
			return g.recordSimpleInitialAssignment(target, value, seen)
		}
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
	normalized, err := g.expandInitialSimpleValue(value)
	if err != nil {
		return err
	}
	value = normalized
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

func (g *Generator) expandInitialSimpleValue(e goivy.Expr) (goivy.Expr, error) {
	if e == nil {
		return nil, nil
	}
	for depth := 0; depth < 64; depth++ {
		expanded, ok, err := g.expandDefinitionExprOnce(e)
		if err != nil {
			return nil, err
		}
		if !ok {
			return e, nil
		}
		e = expanded
	}
	return nil, fmt.Errorf("ivy2golang: recursive derived definition in initial value %s", e.String())
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
				return nil, withInitialErrorLocation(term, err)
			}
			actions = append(actions, acts...)
		}
		return actions, nil
	default:
		act, err := g.initialConditionAction(f)
		if err != nil {
			return nil, withInitialErrorLocation(f, err)
		}
		return []goivy.Action{act}, nil
	}
}

func (g *Generator) initialAxiomActionsFor(f goivy.Expr) []goivy.Action {
	if expanded, ok, err := g.expandDefinitionExprOnce(f); ok || err != nil {
		if err != nil {
			return nil
		}
		return g.initialAxiomActionsFor(expanded)
	}
	switch n := f.(type) {
	case *goivy.ForAll:
		actions := g.initialMinimumElementActionsFor(n)
		actions = append(actions, g.initialAxiomActionsFor(n.Body)...)
		return actions
	case *goivy.LogicExists:
		return g.initialExistentialAxiomActionsFor(n)
	case *goivy.LogicAnd:
		var actions []goivy.Action
		for _, term := range n.Terms {
			actions = append(actions, g.initialAxiomActionsFor(term)...)
		}
		return actions
	default:
		act, err := g.initialConditionAction(f)
		if err != nil {
			return nil
		}
		return []goivy.Action{act}
	}
}

func (g *Generator) initialMinimumElementActionsFor(fa *goivy.ForAll) []goivy.Action {
	if fa == nil || fa.Body == nil {
		return nil
	}
	app, ok := initialPositiveApply(fa.Body)
	if !ok || len(app.Terms) != 2 {
		return nil
	}
	minTerm := app.Terms[0]
	if minTerm == nil || !g.isStateTarget(minTerm) || len(goivy.VariablesAstList(minTerm)) != 0 {
		return nil
	}
	bound, ok := app.Terms[1].(*goivy.LogicVariable)
	if !ok || bound == nil || !initialForAllBindsVariable(fa, bound) {
		return nil
	}
	if !goivy.SortEqual(minTerm.NodeSort(), bound.VSort) {
		return nil
	}
	values, ok := g.initialFiniteValueTerms(minTerm.NodeSort())
	if !ok || len(values) == 0 {
		return nil
	}
	return []goivy.Action{goivy.NewAssignAction(minTerm, values[0])}
}

func initialForAllBindsVariable(fa *goivy.ForAll, v *goivy.LogicVariable) bool {
	if fa == nil || v == nil {
		return false
	}
	for _, candidate := range fa.Variables {
		if candidate != nil && candidate.Equal(v) {
			return true
		}
	}
	return false
}

func (g *Generator) initialExistentialAxiomActionsFor(ex *goivy.LogicExists) []goivy.Action {
	if ex == nil || ex.Body == nil || len(ex.Variables) == 0 {
		return nil
	}
	subs := make(map[goivy.NodeKey]goivy.Expr, len(ex.Variables))
	for _, v := range ex.Variables {
		if v == nil {
			return nil
		}
		values, ok := g.initialFiniteValueTerms(v.VSort)
		if !ok || len(values) == 0 {
			return nil
		}
		subs[goivy.Key(v)] = values[0]
	}
	body, err := goivy.Substitute(ex.Body, subs)
	if err != nil {
		return nil
	}
	return g.initialAxiomActionsFor(body)
}

func (g *Generator) initialStateRetryFormulas() []goivy.Expr {
	if g == nil || g.Mod == nil || g.Config.Target != "test" {
		return nil
	}
	constructedOrders := g.initialConstructedOrderRelations()
	var formulas []goivy.Expr
	for _, f := range g.Mod.Axioms() {
		formulas = append(formulas, g.initialAxiomRetryFormulasFor(f, constructedOrders)...)
	}
	return formulas
}

func (g *Generator) initialConstructedOrderRelations() map[string]bool {
	out := map[string]bool{}
	if g == nil || g.Mod == nil {
		return out
	}
	for _, f := range g.Mod.Axioms() {
		g.collectInitialConstructedOrderRelations(f, out)
	}
	return out
}

func (g *Generator) collectInitialConstructedOrderRelations(f goivy.Expr, out map[string]bool) {
	if f == nil {
		return
	}
	if expanded, ok, err := g.expandDefinitionExprOnce(f); ok || err != nil {
		if err != nil {
			return
		}
		g.collectInitialConstructedOrderRelations(expanded, out)
		return
	}
	switch n := f.(type) {
	case *goivy.ForAll:
		g.collectInitialConstructedOrderRelations(n.Body, out)
	case *goivy.LogicAnd:
		for _, term := range n.Terms {
			g.collectInitialConstructedOrderRelations(term, out)
		}
	case *goivy.LogicOr:
		if name := initialTotalityOrderRelationName(n.Terms); name != "" {
			out[name] = true
		}
	}
}

func (g *Generator) initialAxiomRetryFormulasFor(f goivy.Expr, constructedOrders map[string]bool) []goivy.Expr {
	if f == nil {
		return nil
	}
	if expanded, ok, err := g.expandDefinitionExprOnce(f); ok || err != nil {
		if err != nil {
			return nil
		}
		return g.initialAxiomRetryFormulasFor(expanded, constructedOrders)
	}
	if and, ok := f.(*goivy.LogicAnd); ok {
		var formulas []goivy.Expr
		for _, term := range and.Terms {
			formulas = append(formulas, g.initialAxiomRetryFormulasFor(term, constructedOrders)...)
		}
		return formulas
	}
	if name := initialConstructedOrderCheckRelationName(f); name != "" && constructedOrders[name] {
		return nil
	}
	if len(g.initialAxiomActionsFor(f)) > 0 {
		return nil
	}
	return []goivy.Expr{f}
}

func (g *Generator) initialConditionAction(f goivy.Expr) (goivy.Action, error) {
	switch n := f.(type) {
	case *goivy.LogicImplies:
		return g.initialGuardedConditionAction(n.T1, n.T2)
	case *goivy.LogicOr:
		if act, ok := g.initialTotalityOrderAction(n.Terms); ok {
			return act, nil
		}
		return g.initialDisjunctiveConditionAction(n.Terms)
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
		if n.Polarity == 0 {
			if target, value, ok := g.initialNegatedEqualityAssignment(n.Atom); ok {
				return goivy.NewAssignAction(target, value), nil
			}
		}
		if !g.isStateTarget(n.Atom) {
			return nil, fmt.Errorf("initial literal is not mutable state: %s", n.Atom)
		}
		return g.initialSetActionWithMinimum(n.Atom, goivy.NewSetAction(n)), nil
	case *goivy.LogicNot:
		if target, value, ok := g.initialNegatedEqualityAssignment(n.Body); ok {
			return goivy.NewAssignAction(target, value), nil
		}
		if !g.isStateTarget(n.Body) {
			return nil, fmt.Errorf("initial negation is not mutable state: %s", n.Body)
		}
		return goivy.NewSetAction(n), nil
	case *goivy.Const, *goivy.Apply:
		if !g.isStateTarget(f) {
			return nil, fmt.Errorf("initial literal is not mutable state: %s", f)
		}
		return g.initialSetActionWithMinimum(f, goivy.NewSetAction(f)), nil
	default:
		return nil, fmt.Errorf("unsupported initial formula %T: %s", f, f.String())
	}
}

func (g *Generator) initialSetActionWithMinimum(target goivy.Expr, set goivy.Action) goivy.Action {
	app, ok := target.(*goivy.Apply)
	if !ok {
		return set
	}
	min := g.initialMinimumElementActionForApply(app)
	if min == nil {
		return set
	}
	minExpr, minOK := min.(goivy.Expr)
	setExpr, setOK := set.(goivy.Expr)
	if !minOK || !setOK {
		return set
	}
	return goivy.NewSequence(minExpr, setExpr)
}

func (g *Generator) initialMinimumElementActionForApply(app *goivy.Apply) goivy.Action {
	if app == nil || len(app.Terms) != 2 {
		return nil
	}
	minTerm := app.Terms[0]
	if minTerm == nil || !g.isStateTarget(minTerm) || len(goivy.VariablesAstList(minTerm)) != 0 {
		return nil
	}
	bound, ok := app.Terms[1].(*goivy.LogicVariable)
	if !ok || bound == nil || !goivy.SortEqual(minTerm.NodeSort(), bound.VSort) {
		return nil
	}
	values, ok := g.initialFiniteValueTerms(minTerm.NodeSort())
	if !ok || len(values) == 0 {
		return nil
	}
	return goivy.NewAssignAction(minTerm, values[0])
}

func (g *Generator) initialTotalityOrderAction(terms []goivy.Expr) (goivy.Action, bool) {
	if initialTotalityOrderRelationName(terms) == "" {
		return nil, false
	}
	if len(terms) != 2 {
		return nil, false
	}
	left, ok := initialPositiveApply(terms[0])
	if !ok {
		return nil, false
	}
	right, ok := initialPositiveApply(terms[1])
	if !ok {
		return nil, false
	}
	if !g.isStateTarget(left) || len(left.Terms) != 2 || len(right.Terms) != 2 {
		return nil, false
	}
	if !left.Func.Equal(right.Func) {
		return nil, false
	}
	if !left.Terms[0].Equal(right.Terms[1]) || !left.Terms[1].Equal(right.Terms[0]) {
		return nil, false
	}
	if !goivy.SortEqual(left.Terms[0].NodeSort(), left.Terms[1].NodeSort()) {
		return nil, false
	}
	fs, err := goivy.NewFunctionSort(left.Terms[0].NodeSort(), left.Terms[1].NodeSort(), goivy.Boolean)
	if err != nil {
		return nil, false
	}
	order, err := goivy.NewApply(goivy.NewConst("<=", fs), left.Terms[0], left.Terms[1])
	if err != nil {
		return nil, false
	}
	return goivy.NewAssignAction(left, order), true
}

func initialTotalityOrderRelationName(terms []goivy.Expr) string {
	if len(terms) != 2 {
		return ""
	}
	left, ok := initialPositiveApply(terms[0])
	if !ok {
		return ""
	}
	right, ok := initialPositiveApply(terms[1])
	if !ok || len(left.Terms) != 2 || len(right.Terms) != 2 {
		return ""
	}
	if !left.Func.Equal(right.Func) {
		return ""
	}
	if !left.Terms[0].Equal(right.Terms[1]) || !left.Terms[1].Equal(right.Terms[0]) {
		return ""
	}
	if !goivy.SortEqual(left.Terms[0].NodeSort(), left.Terms[1].NodeSort()) {
		return ""
	}
	return goivy.ExprName(left.Func)
}

func initialConstructedOrderCheckRelationName(f goivy.Expr) string {
	if fa, ok := f.(*goivy.ForAll); ok {
		return initialConstructedOrderCheckRelationName(fa.Body)
	}
	imp, ok := f.(*goivy.LogicImplies)
	if !ok {
		return ""
	}
	terms := initialAndTerms(imp.T1)
	if len(terms) != 2 {
		return ""
	}
	if name := initialTransitiveOrderCheckRelationName(terms[0], terms[1], imp.T2); name != "" {
		return name
	}
	if name := initialTransitiveOrderCheckRelationName(terms[1], terms[0], imp.T2); name != "" {
		return name
	}
	if name := initialAntisymmetricOrderCheckRelationName(terms[0], terms[1], imp.T2); name != "" {
		return name
	}
	return initialAntisymmetricOrderCheckRelationName(terms[1], terms[0], imp.T2)
}

func initialAndTerms(e goivy.Expr) []goivy.Expr {
	if and, ok := e.(*goivy.LogicAnd); ok {
		return and.Terms
	}
	return nil
}

func initialTransitiveOrderCheckRelationName(aExpr, bExpr, cExpr goivy.Expr) string {
	a, ok := initialPositiveApply(aExpr)
	if !ok || len(a.Terms) != 2 {
		return ""
	}
	b, ok := initialPositiveApply(bExpr)
	if !ok || len(b.Terms) != 2 {
		return ""
	}
	c, ok := initialPositiveApply(cExpr)
	if !ok || len(c.Terms) != 2 {
		return ""
	}
	if !a.Func.Equal(b.Func) || !a.Func.Equal(c.Func) {
		return ""
	}
	if a.Terms[1].Equal(b.Terms[0]) && a.Terms[0].Equal(c.Terms[0]) && b.Terms[1].Equal(c.Terms[1]) {
		return goivy.ExprName(a.Func)
	}
	return ""
}

func initialAntisymmetricOrderCheckRelationName(aExpr, bExpr, eqExpr goivy.Expr) string {
	a, ok := initialPositiveApply(aExpr)
	if !ok || len(a.Terms) != 2 {
		return ""
	}
	b, ok := initialPositiveApply(bExpr)
	if !ok || len(b.Terms) != 2 || !a.Func.Equal(b.Func) {
		return ""
	}
	if !a.Terms[0].Equal(b.Terms[1]) || !a.Terms[1].Equal(b.Terms[0]) {
		return ""
	}
	eq, ok := eqExpr.(*goivy.Eq)
	if !ok {
		return ""
	}
	if (eq.T1.Equal(a.Terms[0]) && eq.T2.Equal(a.Terms[1])) || (eq.T1.Equal(a.Terms[1]) && eq.T2.Equal(a.Terms[0])) {
		return goivy.ExprName(a.Func)
	}
	return ""
}

func initialPositiveApply(e goivy.Expr) (*goivy.Apply, bool) {
	switch n := e.(type) {
	case *goivy.Apply:
		return n, true
	case *goivy.LogicLiteral:
		if n.Polarity != 0 {
			if app, ok := n.Atom.(*goivy.Apply); ok {
				return app, true
			}
		}
	}
	return nil, false
}

func (g *Generator) initialDisjunctiveConditionAction(terms []goivy.Expr) (goivy.Action, error) {
	if len(terms) < 2 {
		return nil, fmt.Errorf("unsupported initial disjunction with %d terms", len(terms))
	}
	if len(terms) == 2 {
		if guard, ok := initialNegatedGuard(terms[0]); ok {
			return g.initialGuardedConditionAction(guard, terms[1])
		}
		if guard, ok := initialNegatedGuard(terms[1]); ok {
			return g.initialGuardedConditionAction(guard, terms[0])
		}
	}
	for i := range terms {
		guard, err := initialDisjunctionGuard(terms, i)
		if err != nil {
			return nil, err
		}
		if act, err := g.initialGuardedConditionAction(guard, terms[i]); err == nil {
			return act, nil
		}
	}
	return nil, fmt.Errorf("unsupported initial disjunction with %d terms", len(terms))
}

func initialDisjunctionGuard(terms []goivy.Expr, consequent int) (goivy.Expr, error) {
	guards := make([]goivy.Expr, 0, len(terms)-1)
	for i, term := range terms {
		if i == consequent {
			continue
		}
		if guard, ok := initialNegatedGuard(term); ok {
			guards = append(guards, guard)
		} else {
			guards = append(guards, &goivy.LogicNot{Body: term})
		}
	}
	if len(guards) == 1 {
		return guards[0], nil
	}
	return goivy.NewAnd(guards...)
}

func initialNegatedGuard(e goivy.Expr) (goivy.Expr, bool) {
	switch n := e.(type) {
	case *goivy.LogicNot:
		return n.Body, true
	case *goivy.LogicLiteral:
		if n.Polarity == 0 {
			return n.Atom, true
		}
	}
	return nil, false
}

func (g *Generator) initialGuardedConditionAction(cond, consequent goivy.Expr) (goivy.Action, error) {
	if expanded, ok, err := g.expandDefinitionExprOnce(consequent); ok || err != nil {
		if err != nil {
			return nil, err
		}
		return g.initialGuardedConditionAction(cond, expanded)
	}
	if and, ok := consequent.(*goivy.LogicAnd); ok {
		elems := make([]goivy.Expr, 0, len(and.Terms))
		for _, term := range and.Terms {
			act, err := g.initialGuardedConditionAction(cond, term)
			if err != nil {
				return nil, err
			}
			elems = append(elems, act)
		}
		return goivy.NewSequence(elems...), nil
	}
	target, value, err := g.initialAssignmentTargetValue(consequent)
	if err != nil {
		return nil, fmt.Errorf("unsupported initial implication consequent: %w", err)
	}
	if err := initialGuardVariablesBoundByTarget(target, cond, value); err != nil {
		return nil, err
	}
	rhs, err := goivy.NewIte(cond, value, target)
	if err != nil {
		return nil, err
	}
	return goivy.NewAssignAction(target, rhs), nil
}

func (g *Generator) initialAssignmentTargetValue(f goivy.Expr) (goivy.Expr, goivy.Expr, error) {
	if expanded, ok, err := g.expandDefinitionExprOnce(f); ok || err != nil {
		if err != nil {
			return nil, nil, err
		}
		return g.initialAssignmentTargetValue(expanded)
	}
	switch n := f.(type) {
	case *goivy.Eq:
		if g.isStateTarget(n.T1) {
			return n.T1, n.T2, nil
		}
		if g.isStateTarget(n.T2) {
			return n.T2, n.T1, nil
		}
		return nil, nil, fmt.Errorf("neither side of initial equality is mutable state: %s", f)
	case *goivy.LogicIff:
		if g.isStateTarget(n.T1) {
			return n.T1, n.T2, nil
		}
		if g.isStateTarget(n.T2) {
			return n.T2, n.T1, nil
		}
		return nil, nil, fmt.Errorf("neither side of initial equivalence is mutable state: %s", f)
	case *goivy.LogicLiteral:
		if n.Polarity == 0 {
			if target, value, ok := g.initialNegatedEqualityAssignment(n.Atom); ok {
				return target, value, nil
			}
			if !g.isStateTarget(n.Atom) {
				return nil, nil, fmt.Errorf("initial literal is not mutable state: %s", n.Atom)
			}
			return n.Atom, goivy.NewConst("false", goivy.Boolean), nil
		}
		if !g.isStateTarget(n.Atom) {
			return nil, nil, fmt.Errorf("initial literal is not mutable state: %s", n.Atom)
		}
		return n.Atom, goivy.NewConst("true", goivy.Boolean), nil
	case *goivy.LogicNot:
		if target, value, ok := g.initialNegatedEqualityAssignment(n.Body); ok {
			return target, value, nil
		}
		if !g.isStateTarget(n.Body) {
			return nil, nil, fmt.Errorf("initial negation is not mutable state: %s", n.Body)
		}
		return n.Body, goivy.NewConst("false", goivy.Boolean), nil
	case *goivy.Const, *goivy.Apply:
		if !g.isStateTarget(f) {
			return nil, nil, fmt.Errorf("initial literal is not mutable state: %s", f)
		}
		return f, goivy.NewConst("true", goivy.Boolean), nil
	default:
		return nil, nil, fmt.Errorf("unsupported initial formula %T: %s", f, f.String())
	}
}

func initialGuardVariablesBoundByTarget(target goivy.Expr, exprs ...goivy.Expr) error {
	bound := map[goivy.NodeKey]bool{}
	for _, v := range goivy.VariablesAstList(target) {
		bound[goivy.Key(v)] = true
	}
	for _, expr := range exprs {
		for _, v := range goivy.VariablesAstList(expr) {
			if !bound[goivy.Key(v)] {
				return fmt.Errorf("guarded initial assignment for %s uses unbound variable %s", target.String(), v.Repr())
			}
		}
	}
	return nil
}

func (g *Generator) initialNegatedEqualityAssignment(e goivy.Expr) (goivy.Expr, goivy.Expr, bool) {
	eq, ok := e.(*goivy.Eq)
	if !ok {
		return nil, nil, false
	}
	if g.isStateTarget(eq.T1) {
		if value, ok := g.initialUniqueDifferentValue(eq.T1.NodeSort(), eq.T2); ok {
			return eq.T1, value, true
		}
	}
	if g.isStateTarget(eq.T2) {
		if value, ok := g.initialUniqueDifferentValue(eq.T2.NodeSort(), eq.T1); ok {
			return eq.T2, value, true
		}
	}
	return nil, nil, false
}

func (g *Generator) initialUniqueDifferentValue(s goivy.Sort, forbidden goivy.Expr) (goivy.Expr, bool) {
	if s == nil || forbidden == nil {
		return nil, false
	}
	if len(goivy.VariablesAstList(forbidden)) > 0 || g.exprReferencesState(forbidden) {
		return nil, false
	}
	values, ok := g.initialFiniteValueTerms(s)
	if !ok || len(values) != 2 {
		return nil, false
	}
	forbiddenKey := initialFiniteValueKey(forbidden)
	if forbiddenKey == "" {
		return nil, false
	}
	var forced goivy.Expr
	matchedForbidden := false
	for _, value := range values {
		if initialFiniteValueKey(value) == forbiddenKey {
			matchedForbidden = true
			continue
		}
		if forced != nil {
			return nil, false
		}
		forced = value
	}
	if !matchedForbidden || forced == nil {
		return nil, false
	}
	return forced, true
}

func (g *Generator) initialFiniteValueTerms(s goivy.Sort) ([]goivy.Expr, bool) {
	switch st := s.(type) {
	case *goivy.BooleanSort:
		return []goivy.Expr{
			goivy.NewConst("false", goivy.Boolean),
			goivy.NewConst("true", goivy.Boolean),
		}, true
	case *goivy.LogicEnumeratedSort:
		values := make([]goivy.Expr, 0, len(st.Extension))
		for _, name := range st.Extension {
			values = append(values, goivy.NewConst(name, st))
		}
		return values, true
	default:
		if rs, ok := g.rangeSortFor(s); ok {
			lo, hi, ok := numericRangeBounds(rs)
			if !ok || hi < lo {
				return nil, false
			}
			values := make([]goivy.Expr, 0, hi-lo+1)
			for i := lo; i <= hi; i++ {
				values = append(values, goivy.NewConst(strconv.Itoa(i), s))
			}
			return values, true
		}
		if card := g.sortCard(s); card > 0 {
			values := make([]goivy.Expr, 0, card)
			for i := 0; i < card; i++ {
				values = append(values, goivy.NewConst(strconv.Itoa(i), s))
			}
			return values, true
		}
		return nil, false
	}
}

func initialFiniteValueKey(e goivy.Expr) string {
	if e == nil {
		return ""
	}
	if name := goivy.ExprName(e); name != "" {
		return name
	}
	return e.String()
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
