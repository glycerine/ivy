package ivy2cpp

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/glycerine/ivy/goivy"
	"github.com/glycerine/ivy/goivy/smt"
)

type initialStateConstraints struct {
	Formulas []goivy.Expr
	Used     map[string]bool
}

type initialDomainValue struct {
	Expr  goivy.Expr
	Cpp   string
	Z3Int string
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
	for _, lf := range g.Mod.LabeledAxioms {
		if f, ok := lf.Formula.(goivy.Expr); ok && f != nil {
			formulas = append(formulas, f)
		}
	}
	used := usedSymbolNames(formulas)
	for _, lf := range goivy.RelevantDefinitions(g.Mod, used) {
		def, ok := lf.Formula.(*goivy.LogicDefinition)
		if !ok || def == nil {
			continue
		}
		formulas = append(formulas, goivy.DefinitionToConstraint(def))
	}
	return &initialStateConstraints{
		Formulas: formulas,
		Used:     usedSymbolNames(formulas),
	}, nil
}

func (g *Generator) checkInitialStateParameters(kind string, lfs []*goivy.LabeledFormula) error {
	if g == nil || g.Mod == nil || len(g.Mod.Params) == 0 {
		return nil
	}
	params := make(map[string]bool, len(g.Mod.Params))
	for _, p := range g.Mod.Params {
		if p != nil && p.Name != "" {
			params[p.Name] = true
		}
	}
	for _, lf := range lfs {
		f, ok := lf.Formula.(goivy.Expr)
		if !ok || f == nil {
			continue
		}
		for name := range usedSymbolNames([]goivy.Expr{f}) {
			if !params[name] {
				continue
			}
			line := ""
			if lf != nil && lf.Lineno() > 0 {
				line = fmt.Sprintf(" at line %d", lf.Lineno())
			}
			return fmt.Errorf("ivy2cpp: %s%s depends on stripped parameter %q", kind, line, name)
		}
	}
	return nil
}

func usedSymbolNames(formulas []goivy.Expr) map[string]bool {
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

func (g *Generator) emitOneInitialState(w *cppWriter) error {
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
			return fmt.Errorf("ivy2cpp: initial state solver failed: %w", err)
		}
		if mr == nil || mr.Model == nil {
			return fmt.Errorf("ivy2cpp: axioms and/or initial condition are inconsistent")
		}
		model = mr.Model
	}
	for _, sym := range g.stateSymbols() {
		if g.isParamName(sym.Name) {
			continue
		}
		if !constraints.Used[sym.Name] || model == nil {
			g.emitDefaultInitialState(w, sym)
			continue
		}
		if err := g.emitSolvedInitialState(w, slv, model, sym); err != nil {
			return err
		}
	}
	return nil
}

func (g *Generator) emitDefaultInitialState(w *cppWriter, sym stateSymbol) {
	if fs, ok := sym.Sort.(*goivy.LogicFunctionSort); ok && len(fs.Domain()) > 0 {
		st := cppFunctionStorageFor(g, fs.Domain(), fs.Range(), "")
		if st.Kind == cppStorageArray {
			tuples, ok := g.initialDomainTuples(fs.Domain())
			if ok {
				for _, tuple := range tuples {
					args := make([]string, len(tuple))
					for i, v := range tuple {
						args[i] = v.Cpp
					}
					w.linef("%s = %s;", g.cppStorageAccess(sym.Name, sym.Sort, args, ""), g.cppZeroValue(fs.Range()))
				}
			}
			return
		}
		w.linef("%s = %s();", varName(sym.Name), st.Type)
		return
	}
	rng := initialStateRange(sym.Sort)
	w.linef("%s = %s;", varName(sym.Name), g.cppZeroValue(rng))
}

func (g *Generator) emitSolvedInitialState(w *cppWriter, slv *goivy.Solver, model *smt.Model, sym stateSymbol) error {
	if slv == nil || model == nil {
		g.emitDefaultInitialState(w, sym)
		return nil
	}
	fs, ok := sym.Sort.(*goivy.LogicFunctionSort)
	if !ok || len(fs.Domain()) == 0 {
		term := initialStateSymbolTerm(sym, nil)
		value, err := g.initialModelValue(slv, model, term, initialStateRange(sym.Sort))
		if err != nil {
			return fmt.Errorf("initial value for %s: %w", sym.Name, err)
		}
		w.linef("%s = %s;", varName(sym.Name), value)
		return nil
	}
	tuples, ok := g.initialDomainTuples(fs.Domain())
	if !ok {
		return fmt.Errorf("ivy2cpp: cannot enumerate initial-state domain of %s", sym.Name)
	}
	for _, tuple := range tuples {
		args := make([]goivy.Expr, len(tuple))
		cppArgs := make([]string, len(tuple))
		for i, v := range tuple {
			args[i] = v.Expr
			cppArgs[i] = v.Cpp
		}
		term := initialStateSymbolTerm(sym, args)
		value, err := g.initialModelValue(slv, model, term, fs.Range())
		if err != nil {
			return fmt.Errorf("initial value for %s: %w", sym.Name, err)
		}
		w.linef("%s = %s;", g.cppStorageAccess(sym.Name, sym.Sort, cppArgs, ""), value)
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
	return g.modelValueToCpp(value, s)
}

func (g *Generator) modelValueToCpp(value smt.Z3Expr, s goivy.Sort) (string, error) {
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
			if text == name || text == fmt.Sprintf("%s_%d", z3SortName(st), i) {
				return varName(name), nil
			}
		}
		if idx, ok := parseTrailingModelIndex(text); ok && idx >= 0 && idx < len(st.Extension) {
			return varName(st.Extension[idx]), nil
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
			{Expr: goivy.NewConst("false", goivy.Boolean), Cpp: "false", Z3Int: "0"},
			{Expr: goivy.NewConst("true", goivy.Boolean), Cpp: "true", Z3Int: "1"},
		}, true
	case *goivy.LogicEnumeratedSort:
		vals := make([]initialDomainValue, 0, len(st.Extension))
		for i, name := range st.Extension {
			vals = append(vals, initialDomainValue{
				Expr:  goivy.NewConst(name, st),
				Cpp:   varName(name),
				Z3Int: strconv.Itoa(i),
			})
		}
		return vals, true
	default:
		if rs, ok := g.rangeSortFor(s); ok {
			loText, hiText, ok := numericRangeBounds(rs)
			if !ok {
				return nil, false
			}
			lo, _ := strconv.Atoi(loText)
			hi, _ := strconv.Atoi(hiText)
			vals := make([]initialDomainValue, 0, hi-lo+1)
			for i := lo; i <= hi; i++ {
				text := strconv.Itoa(i)
				vals = append(vals, initialDomainValue{
					Expr:  goivy.NewConst(text, s),
					Cpp:   text,
					Z3Int: text,
				})
			}
			return vals, true
		}
		return nil, false
	}
}

func (g *Generator) emitProgressCounterResets(w *cppWriter, obj string) {
	for _, p := range g.progressDecls() {
		g.emitProgressCounterReset(w, p, obj)
	}
}

func (g *Generator) emitProgressCounterReset(w *cppWriter, p progressDecl, obj string) {
	// Python ivy_to_cpp.py:884-895 emit_clear_progress: opens the loop
	// over free variables of the progress LHS and emits a per-iteration
	// `lhs = 0;` assignment. No hash_thunk shortcut.
	prefix := ""
	if obj != "" {
		prefix = obj + "."
	}
	if len(p.Vars) == 0 {
		w.linef("%s%s = 0;", prefix, varName(p.Name))
		return
	}
	opened := g.openProgressLoops(w, p)
	lhs := g.progressCounterLValueWithObj(p, obj)
	w.linef("%s = 0;", lhs)
	g.closeAssignmentLoops(w, opened)
}

func (g *Generator) emitZ3InitialConstraints(w *cppWriter) error {
	constraints, err := g.initialStateConstraints()
	if err != nil {
		return err
	}
	for _, f := range constraints.Formulas {
		if err := g.emitZ3AddInitialFormula(w, f, map[string]goivy.Sort{}); err != nil {
			return err
		}
	}
	return nil
}

func (g *Generator) emitZ3AddInitialFormula(w *cppWriter, f goivy.Expr, env map[string]goivy.Sort) error {
	switch n := f.(type) {
	case *goivy.ForAll:
		return g.emitZ3ForAllInitialFormula(w, n, env)
	case *goivy.LogicAnd:
		for _, term := range n.Terms {
			if err := g.emitZ3AddInitialFormula(w, term, env); err != nil {
				return err
			}
		}
		return nil
	default:
		expr, err := g.z3InitialExpr(f, env)
		if err != nil {
			return err
		}
		w.linef("add(%s);", expr)
		return nil
	}
}

func (g *Generator) emitZ3ForAllInitialFormula(w *cppWriter, f *goivy.ForAll, env map[string]goivy.Sort) error {
	nextEnv := cloneSortEnv(env)
	opened := 0
	for _, v := range f.Variables {
		header, ok := g.z3LoopHeaderForSort(v.VSort, varName(v.Name))
		if !ok {
			return fmt.Errorf("ivy2cpp: cannot enumerate quantified initial-state variable %s:%s", v.Name, sortName(v.VSort))
		}
		w.line(header)
		w.indent++
		opened++
		nextEnv[v.Name] = v.VSort
	}
	err := g.emitZ3AddInitialFormula(w, f.Body, nextEnv)
	for i := 0; i < opened; i++ {
		w.indent--
		w.line("}")
	}
	return err
}

func cloneSortEnv(env map[string]goivy.Sort) map[string]goivy.Sort {
	out := make(map[string]goivy.Sort, len(env))
	for k, v := range env {
		out[k] = v
	}
	return out
}

func (g *Generator) z3InitialExpr(e goivy.Expr, env map[string]goivy.Sort) (string, error) {
	switch n := e.(type) {
	case *goivy.Const:
		return g.z3InitialConst(n, env)
	case *goivy.LogicVariable:
		return g.z3InitialVariable(n.Name, n.VSort, env)
	case *goivy.Apply:
		return g.z3InitialApply(n, env)
	case *goivy.Eq:
		l, err := g.z3InitialExpr(n.T1, env)
		if err != nil {
			return "", err
		}
		r, err := g.z3InitialExpr(n.T2, env)
		if err != nil {
			return "", err
		}
		return "(" + l + " == " + r + ")", nil
	case *goivy.LogicIff:
		l, err := g.z3InitialExpr(n.T1, env)
		if err != nil {
			return "", err
		}
		r, err := g.z3InitialExpr(n.T2, env)
		if err != nil {
			return "", err
		}
		return "(" + l + " == " + r + ")", nil
	case *goivy.LogicNot:
		body, err := g.z3InitialExpr(n.Body, env)
		if err != nil {
			return "", err
		}
		return "!(" + body + ")", nil
	case *goivy.LogicLiteral:
		body, err := g.z3InitialExpr(n.Atom, env)
		if err != nil {
			return "", err
		}
		if n.Polarity == 0 {
			return "!(" + body + ")", nil
		}
		return body, nil
	case *goivy.LogicAnd:
		return g.z3InitialNary(n.Terms, "&&", "ctx.bool_val(true)", env)
	case *goivy.LogicOr:
		return g.z3InitialNary(n.Terms, "||", "ctx.bool_val(false)", env)
	case *goivy.LogicImplies:
		l, err := g.z3InitialExpr(n.T1, env)
		if err != nil {
			return "", err
		}
		r, err := g.z3InitialExpr(n.T2, env)
		if err != nil {
			return "", err
		}
		return "z3::implies(" + l + ", " + r + ")", nil
	default:
		return "", fmt.Errorf("ivy2cpp: unsupported initial-state Z3 expression %T: %s", e, e.String())
	}
}

func (g *Generator) z3InitialNary(terms []goivy.Expr, op, ident string, env map[string]goivy.Sort) (string, error) {
	if len(terms) == 0 {
		return ident, nil
	}
	parts := make([]string, len(terms))
	for i, term := range terms {
		part, err := g.z3InitialExpr(term, env)
		if err != nil {
			return "", err
		}
		parts[i] = "(" + part + ")"
	}
	return strings.Join(parts, " "+op+" "), nil
}

func (g *Generator) z3InitialConst(c *goivy.Const, env map[string]goivy.Sort) (string, error) {
	if c == nil {
		return "", fmt.Errorf("ivy2cpp: nil initial-state const")
	}
	if s, ok := env[c.Name]; ok {
		return g.z3InitialVariable(c.Name, s, env)
	}
	if c.CSort == goivy.Boolean {
		switch c.Name {
		case "true":
			return "ctx.bool_val(true)", nil
		case "false":
			return "ctx.bool_val(false)", nil
		}
	}
	if goivy.IsNumeral(c) && !goivy.IsLiteralString(c) {
		return fmt.Sprintf("int_to_z3(%q, %s)", z3SortName(c.CSort), c.Name), nil
	}
	if enum, ok := c.CSort.(*goivy.LogicEnumeratedSort); ok {
		for i, name := range enum.Extension {
			if name == c.Name {
				return fmt.Sprintf("int_to_z3(%q, %d)", z3SortName(enum), i), nil
			}
		}
	}
	name := goivy.ExprName(c)
	if g.z3HasDecl(name) {
		return fmt.Sprintf("mk_apply_expr(%q, {})", name), nil
	}
	return "", fmt.Errorf("ivy2cpp: unsupported initial-state constant %s", c.String())
}

func (g *Generator) z3InitialVariable(name string, s goivy.Sort, env map[string]goivy.Sort) (string, error) {
	if s == nil {
		if envSort, ok := env[name]; ok {
			s = envSort
		}
	}
	if s == nil {
		return "", fmt.Errorf("ivy2cpp: initial-state variable %s has no sort", name)
	}
	return fmt.Sprintf("__to_solver(*this, %q, %s)", z3SortName(s), varName(name)), nil
}

func (g *Generator) z3InitialApply(a *goivy.Apply, env map[string]goivy.Sort) (string, error) {
	name := goivy.ExprName(a.Func)
	if len(a.Terms) == 2 && isInfix(name) {
		l, err := g.z3InitialExpr(a.Terms[0], env)
		if err != nil {
			return "", err
		}
		r, err := g.z3InitialExpr(a.Terms[1], env)
		if err != nil {
			return "", err
		}
		return "(" + l + " " + name + " " + r + ")", nil
	}
	if !g.z3HasDecl(name) {
		return "", fmt.Errorf("ivy2cpp: unsupported initial-state apply %s", a.String())
	}
	args := make([]string, len(a.Terms))
	for i, term := range a.Terms {
		arg, err := g.z3InitialApplyArg(term, env)
		if err != nil {
			return "", err
		}
		args[i] = arg
	}
	return fmt.Sprintf("mk_apply_expr(%q, {%s})", name, strings.Join(args, ", ")), nil
}

func (g *Generator) z3InitialApplyArg(e goivy.Expr, env map[string]goivy.Sort) (string, error) {
	switch n := e.(type) {
	case *goivy.LogicVariable:
		return "static_cast<int>(" + varName(n.Name) + ")", nil
	case *goivy.Const:
		if s, ok := env[n.Name]; ok {
			_ = s
			return "static_cast<int>(" + varName(n.Name) + ")", nil
		}
		if n.CSort == goivy.Boolean {
			if n.Name == "true" {
				return "1", nil
			}
			if n.Name == "false" {
				return "0", nil
			}
		}
		if goivy.IsNumeral(n) && !goivy.IsLiteralString(n) {
			return n.Name, nil
		}
		if enum, ok := n.CSort.(*goivy.LogicEnumeratedSort); ok {
			for i, name := range enum.Extension {
				if name == n.Name {
					return strconv.Itoa(i), nil
				}
			}
		}
	}
	return "", fmt.Errorf("ivy2cpp: unsupported initial-state application argument %s", e.String())
}

func (g *Generator) z3HasDecl(name string) bool {
	if name == "" {
		return false
	}
	for _, sym := range g.z3DeclSymbols() {
		if sym.Name == name {
			return true
		}
	}
	return false
}

func (g *Generator) emitZ3InitialStateEvaluation(w *cppWriter, obj string) error {
	constraints, err := g.initialStateConstraints()
	if err != nil {
		return err
	}
	for _, sym := range g.stateSymbols() {
		if g.isParamName(sym.Name) {
			continue
		}
		if !constraints.Used[sym.Name] {
			continue
		}
		if err := g.emitZ3EvaluateStateSymbol(w, obj, sym); err != nil {
			return err
		}
	}
	return nil
}

func (g *Generator) emitZ3EvaluateStateSymbol(w *cppWriter, obj string, sym stateSymbol) error {
	// Use the shared emit_eval body so init_gen and action_gen produce
	// identical evaluation code (Python ivy_to_cpp.py:772-799).
	return g.emitFromSolverLoop(w, obj, sym)
}

func (g *Generator) isParamName(name string) bool {
	if g == nil || g.Mod == nil || name == "" {
		return false
	}
	for _, p := range g.Mod.Params {
		if p != nil && p.Name == name {
			return true
		}
	}
	return false
}
