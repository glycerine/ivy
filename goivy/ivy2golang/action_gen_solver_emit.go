package ivy2golang

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/glycerine/ivy/goivy"
)

type runtimeActionSolverPlan struct {
	plan       *actionGenPlan
	assigns    []runtimeActionSolverAssign
	stateSyms  []stateSymbol
	needsState map[string]bool
}

type runtimeActionSolverAssign struct {
	input  *goivy.Const
	formal *goivy.Const
}

func (g *Generator) actionGeneratorUsesRuntimeSolver() bool {
	if g == nil || (g.Config.Target != "test" && g.Config.Target != "gen") {
		return false
	}
	for _, name := range g.runnableActionNamesFrom(g.publicActionNamesSorted()) {
		act, _ := g.Mod.Actions.Get2(name)
		if _, ok := g.runtimeActionSolverPlan(name, g.actionGeneratorAnalysisAction(name, act)); ok {
			return true
		}
	}
	return false
}

func (g *Generator) runtimeActionSolverPlan(name string, act goivy.Action) (_ *runtimeActionSolverPlan, ok bool) {
	defer func() {
		if recover() != nil {
			ok = false
		}
	}()
	if g == nil || g.Mod == nil || act == nil {
		return nil, false
	}
	if len(g.actionGeneratorChoiceOverridePlans(act)) > 0 {
		return nil, false
	}
	oldErrs := len(g.errs)
	plan := g.buildActionGenPlan(name, act)
	if plan == nil || plan.fallback || plan.preFmla == nil || (goivy.IsTrue(plan.preFmla) && len(plan.paramDefs) == 0) {
		g.errs = g.errs[:oldErrs]
		return nil, false
	}
	formals := map[string]*goivy.Const{}
	for _, p := range act.GetFormalParams() {
		if p == nil {
			continue
		}
		if _, isFn := p.CSort.(*goivy.LogicFunctionSort); isFn {
			g.errs = g.errs[:oldErrs]
			return nil, false
		}
		for _, alias := range formalExprOverrideNames(p.Name) {
			formals[alias] = p
		}
	}
	var assigns []runtimeActionSolverAssign
	for _, in := range plan.inputs {
		if in == nil || strings.HasPrefix(in.Name, "__ts") || in.Name == "*>" || g.isDerivedStorageName(in.Name) {
			continue
		}
		if actionGenStrictDefIdx(in, plan) {
			continue
		}
		p, found := formals[in.Name]
		if !found {
			g.errs = g.errs[:oldErrs]
			return nil, false
		}
		assigns = append(assigns, runtimeActionSolverAssign{input: in, formal: p})
	}
	if len(assigns) == 0 {
		g.errs = g.errs[:oldErrs]
		return nil, false
	}
	preUsed := goivy.UsedSymbolsAst(plan.preFmla)
	defNames := preDefinedNames(plan.oldPreClauses)
	needsState := map[string]bool{}
	var stateSyms []stateSymbol
	for _, sym := range g.stateSymbols() {
		if g.isParamName(sym.Name) || defNames[sym.Name] || !preUsedContains(preUsed, sym.Name) {
			continue
		}
		if !g.runtimeSolverSupportsStateSymbol(sym) {
			g.errs = g.errs[:oldErrs]
			return nil, false
		}
		needsState[sym.Name] = true
		stateSyms = append(stateSyms, sym)
	}
	g.errs = g.errs[:oldErrs]
	return &runtimeActionSolverPlan{plan: plan, assigns: assigns, stateSyms: stateSyms, needsState: needsState}, true
}

func (g *Generator) runtimeSolverSupportsScalarSort(s goivy.Sort) bool {
	if s == nil {
		return false
	}
	if _, isFn := s.(*goivy.LogicFunctionSort); isFn {
		return false
	}
	switch s.(type) {
	case *goivy.BooleanSort, *goivy.LogicEnumeratedSort, *goivy.RangeSort:
		return true
	case *goivy.UninterpretedSort:
		if g.hasIntOrNatInterp(s) {
			return true
		}
		if _, ok := g.rangeSortFor(s); ok {
			return true
		}
		if g.sortCard(s) > 0 {
			return true
		}
		return false
	default:
		if _, ok := g.rangeSortFor(s); ok {
			return true
		}
		if g.sortCard(s) > 0 {
			return true
		}
		return false
	}
}

func (g *Generator) runtimeSolverSupportsStateSymbol(sym stateSymbol) bool {
	if fs, ok := sym.Sort.(*goivy.LogicFunctionSort); ok {
		if !g.runtimeSolverSupportsScalarSort(fs.Range()) {
			return false
		}
		product := 1
		for _, d := range fs.Domain() {
			values, ok := g.finiteValueExprs(d)
			if !ok || len(values) == 0 {
				return false
			}
			product *= len(values)
			if product > goLargeThresh {
				return false
			}
		}
		return true
	}
	return g.runtimeSolverSupportsScalarSort(sym.Sort)
}

func (g *Generator) emitRuntimeSolverSupport(w *goWriter) {
	if !g.actionGeneratorUsesRuntimeSolver() {
		return
	}
	w.open("func __ivy_solver_func_sort(sorts ...goivy.Sort) goivy.Sort {")
	w.line("s, err := goivy.NewFunctionSort(sorts...)")
	w.open("if err != nil {")
	w.line("panic(err)")
	w.close("")
	w.line("return s")
	w.close("")
	w.blank()
	w.open("func __ivy_solver_module() *goivy.Module {")
	w.line("mod := goivy.New()")
	w.line("mod.Cfg = goivy.NewConfig()")
	w.line("mod.Sig = goivy.NewSigOn(mod.Cfg.IuCfg)")
	for _, s := range g.runtimeSolverSorts() {
		if sortName(s) == "bool" {
			continue
		}
		w.linef("_ = mod.Sig.AddSort(%s)", g.goIvySortExpr(s))
	}
	for _, c := range g.runtimeSolverSymbols() {
		w.linef("_, _ = mod.Sig.AddSymbol(%q, %s)", c.Name, g.goIvySortExpr(c.CSort))
		if es, ok := c.CSort.(*goivy.LogicEnumeratedSort); ok {
			for _, val := range es.Extension {
				if val == c.Name {
					w.linef("mod.Sig.Constructors[%q] = true", c.Name)
					break
				}
			}
		}
	}
	w.line("return mod")
	w.close("")
	w.blank()
}

func (g *Generator) runtimeSolverSorts() []goivy.Sort {
	seen := map[string]bool{}
	var out []goivy.Sort
	var add func(goivy.Sort)
	add = func(s goivy.Sort) {
		if s == nil {
			return
		}
		if fs, ok := s.(*goivy.LogicFunctionSort); ok {
			for _, sub := range fs.Sorts {
				add(sub)
			}
			return
		}
		name := sortName(s)
		if name == "" || seen[name] {
			return
		}
		seen[name] = true
		out = append(out, s)
	}
	for _, s := range g.Mod.Sig.Sorts.All() {
		add(s)
	}
	sort.SliceStable(out, func(i, j int) bool { return sortName(out[i]) < sortName(out[j]) })
	return out
}

func (g *Generator) runtimeSolverSymbols() []*goivy.Const {
	seen := map[string]bool{}
	var out []*goivy.Const
	for name, entry := range g.Mod.Sig.Symbols.All() {
		if entry == nil || seen[name] {
			continue
		}
		seen[name] = true
		out = append(out, goivy.NewConst(name, entry.Sort))
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func (g *Generator) goIvySortExpr(s goivy.Sort) string {
	switch st := s.(type) {
	case *goivy.BooleanSort:
		return "goivy.Boolean"
	case *goivy.LogicEnumeratedSort:
		ext := make([]string, len(st.Extension))
		for i, v := range st.Extension {
			ext[i] = strconv.Quote(v)
		}
		return fmt.Sprintf("&goivy.LogicEnumeratedSort{Name: %q, Extension: []string{%s}}", st.Name, strings.Join(ext, ", "))
	case *goivy.RangeSort:
		return fmt.Sprintf("&goivy.RangeSort{Name: %q, Lb: goivy.NumeralBound{Value: %q}, Ub: goivy.NumeralBound{Value: %q}}", st.Name, st.LbString(), st.UbString())
	case *goivy.UninterpretedSort:
		return fmt.Sprintf("&goivy.UninterpretedSort{Name: %q}", st.Name)
	case *goivy.LogicFunctionSort:
		parts := make([]string, len(st.Sorts))
		for i, sub := range st.Sorts {
			parts[i] = g.goIvySortExpr(sub)
		}
		return fmt.Sprintf("__ivy_solver_func_sort(%s)", strings.Join(parts, ", "))
	default:
		return fmt.Sprintf("&goivy.UninterpretedSort{Name: %q}", sortName(s))
	}
}

func (g *Generator) goIvyExprExpr(e goivy.Expr) (string, bool) {
	switch n := e.(type) {
	case *goivy.Const:
		return fmt.Sprintf("goivy.NewConst(%q, %s)", n.Name, g.goIvySortExpr(n.CSort)), true
	case *goivy.LogicVariable:
		return fmt.Sprintf("&goivy.LogicVariable{Name: %q, VSort: %s}", n.Name, g.goIvySortExpr(n.VSort)), true
	case *goivy.Apply:
		fn, ok := g.goIvyExprExpr(n.Func)
		if !ok {
			return "", false
		}
		args := make([]string, len(n.Terms))
		for i, term := range n.Terms {
			arg, ok := g.goIvyExprExpr(term)
			if !ok {
				return "", false
			}
			args[i] = arg
		}
		if len(args) == 0 {
			return fn, true
		}
		return fmt.Sprintf("goivy.MustApply(%s, %s)", fn, strings.Join(args, ", ")), true
	case *goivy.Eq:
		t1, ok := g.goIvyExprExpr(n.T1)
		if !ok {
			return "", false
		}
		t2, ok := g.goIvyExprExpr(n.T2)
		if !ok {
			return "", false
		}
		return fmt.Sprintf("&goivy.Eq{T1: %s, T2: %s}", t1, t2), true
	case *goivy.LogicAnd:
		return g.goIvyNaryExpr(n.Terms, "&goivy.LogicAnd{Terms: []goivy.Expr{%s}}")
	case *goivy.LogicOr:
		return g.goIvyNaryExpr(n.Terms, "&goivy.LogicOr{Terms: []goivy.Expr{%s}}")
	case *goivy.LogicNot:
		body, ok := g.goIvyExprExpr(n.Body)
		if !ok {
			return "", false
		}
		return fmt.Sprintf("&goivy.LogicNot{Body: %s}", body), true
	case *goivy.LogicLiteral:
		atom, ok := g.goIvyExprExpr(n.Atom)
		if !ok {
			return "", false
		}
		return fmt.Sprintf("&goivy.LogicLiteral{Atom: %s, Polarity: %d}", atom, n.Polarity), true
	case *goivy.LogicImplies:
		t1, ok := g.goIvyExprExpr(n.T1)
		if !ok {
			return "", false
		}
		t2, ok := g.goIvyExprExpr(n.T2)
		if !ok {
			return "", false
		}
		return fmt.Sprintf("&goivy.LogicImplies{T1: %s, T2: %s}", t1, t2), true
	case *goivy.LogicIff:
		t1, ok := g.goIvyExprExpr(n.T1)
		if !ok {
			return "", false
		}
		t2, ok := g.goIvyExprExpr(n.T2)
		if !ok {
			return "", false
		}
		return fmt.Sprintf("&goivy.LogicIff{T1: %s, T2: %s}", t1, t2), true
	case *goivy.LogicIte:
		cond, ok := g.goIvyExprExpr(n.Cond)
		if !ok {
			return "", false
		}
		thenExpr, ok := g.goIvyExprExpr(n.Then)
		if !ok {
			return "", false
		}
		elseExpr, ok := g.goIvyExprExpr(n.Else)
		if !ok {
			return "", false
		}
		return fmt.Sprintf("&goivy.LogicIte{ISort: %s, Cond: %s, Then: %s, Else: %s}", g.goIvySortExpr(n.NodeSort()), cond, thenExpr, elseExpr), true
	case *goivy.ForAll:
		vars, ok := g.goIvyVariableSliceExpr(n.Variables)
		if !ok {
			return "", false
		}
		body, ok := g.goIvyExprExpr(n.Body)
		if !ok {
			return "", false
		}
		return fmt.Sprintf("&goivy.ForAll{Variables: %s, Body: %s}", vars, body), true
	case *goivy.LogicExists:
		vars, ok := g.goIvyVariableSliceExpr(n.Variables)
		if !ok {
			return "", false
		}
		body, ok := g.goIvyExprExpr(n.Body)
		if !ok {
			return "", false
		}
		return fmt.Sprintf("&goivy.LogicExists{Variables: %s, Body: %s}", vars, body), true
	default:
		return "", false
	}
}

func (g *Generator) goIvyNaryExpr(terms []goivy.Expr, format string) (string, bool) {
	parts := make([]string, len(terms))
	for i, term := range terms {
		part, ok := g.goIvyExprExpr(term)
		if !ok {
			return "", false
		}
		parts[i] = part
	}
	return fmt.Sprintf(format, strings.Join(parts, ", ")), true
}

func (g *Generator) goIvyVariableSliceExpr(vars []*goivy.LogicVariable) (string, bool) {
	parts := make([]string, len(vars))
	for i, v := range vars {
		if v == nil {
			return "", false
		}
		parts[i] = fmt.Sprintf("&goivy.LogicVariable{Name: %q, VSort: %s}", v.Name, g.goIvySortExpr(v.VSort))
	}
	return "[]*goivy.LogicVariable{" + strings.Join(parts, ", ") + "}", true
}

func (g *Generator) emitRuntimeActionSolverMethod(w *goWriter, typeName string, rsp *runtimeActionSolverPlan) {
	if rsp == nil || rsp.plan == nil {
		return
	}
	preExpr, ok := g.goIvyExprExpr(rsp.plan.preFmla)
	if !ok {
		return
	}
	w.open(fmt.Sprintf("func (gen *%s) __ivy_generate_with_solver() bool {", typeName))
	w.line("ivy := gen.ivy")
	w.open("if ivy == nil {")
	w.line("return false")
	w.close("")
	w.line("mod := __ivy_solver_module()")
	w.line("solver := goivy.NewSolver(mod, nil)")
	w.linef("clauses := goivy.NewClauses([]goivy.Expr{%s}, nil, nil)", preExpr)
	for _, def := range rsp.plan.paramDefs {
		defExpr, ok := g.goIvyExprExpr(def)
		if !ok {
			w.line("return false")
			continue
		}
		w.linef("clauses.Fmlas = append(clauses.Fmlas, %s)", defExpr)
	}
	for _, sym := range rsp.stateSyms {
		g.emitRuntimeActionSolverStateEquality(w, sym)
	}
	w.line("model, err := solver.GetModelClauses(clauses)")
	w.open("if err != nil || model == nil {")
	w.line("return false")
	w.close("")
	w.line("hm := goivy.NewHerbrandModel(solver, model.Solver, model.Model, model.Vocab)")
	for _, assign := range rsp.assigns {
		inputExpr, ok := g.goIvyExprExpr(assign.input)
		if !ok {
			w.line("return false")
			continue
		}
		g.emitRuntimeActionSolverAssignModelValue(w, "gen."+goName(assign.formal.Name), assign.formal.CSort, fmt.Sprintf("hm.EvalToConstant(%s)", inputExpr))
	}
	w.line("return true")
	w.close("")
	w.blank()
}

func (g *Generator) emitRuntimeActionSolverStateEquality(w *goWriter, sym stateSymbol) {
	if fs, ok := sym.Sort.(*goivy.LogicFunctionSort); ok {
		g.emitRuntimeActionSolverFunctionStateEqualities(w, sym, fs, nil, nil)
		return
	}
	lhs := fmt.Sprintf("goivy.NewConst(%q, %s)", sym.Name, g.goIvySortExpr(sym.Sort))
	rhs, ok := g.runtimeActionSolverValueExpr("ivy."+goName(sym.Name), sym.Sort)
	if !ok {
		w.line("return false")
		return
	}
	w.linef("clauses.Fmlas = append(clauses.Fmlas, &goivy.Eq{T1: %s, T2: %s})", lhs, rhs)
}

func (g *Generator) emitRuntimeActionSolverFunctionStateEqualities(w *goWriter, sym stateSymbol, fs *goivy.LogicFunctionSort, goArgs []string, ivyArgs []string) {
	depth := len(goArgs)
	domain := fs.Domain()
	if depth < len(domain) {
		values, ok := g.finiteValueExprs(domain[depth])
		if !ok {
			w.line("return false")
			return
		}
		for _, value := range values {
			argExpr, ok := g.runtimeActionSolverValueExpr(value, domain[depth])
			if !ok {
				w.line("return false")
				return
			}
			g.emitRuntimeActionSolverFunctionStateEqualities(
				w,
				sym,
				fs,
				append(append([]string(nil), goArgs...), value),
				append(append([]string(nil), ivyArgs...), argExpr),
			)
		}
		return
	}
	lhs := fmt.Sprintf("goivy.MustApply(goivy.NewConst(%q, %s), %s)", sym.Name, g.goIvySortExpr(sym.Sort), strings.Join(ivyArgs, ", "))
	rhsValue := g.goStorageAccess(sym.Name, sym.Sort, goArgs, "ivy")
	rhs, ok := g.runtimeActionSolverValueExpr(rhsValue, fs.Range())
	if !ok {
		w.line("return false")
		return
	}
	w.linef("clauses.Fmlas = append(clauses.Fmlas, &goivy.Eq{T1: %s, T2: %s})", lhs, rhs)
}

func (g *Generator) runtimeActionSolverValueExpr(value string, s goivy.Sort) (string, bool) {
	switch st := s.(type) {
	case *goivy.BooleanSort:
		_ = st
		return fmt.Sprintf("goivy.NewConst(strconv.FormatBool(%s), goivy.Boolean)", value), true
	case *goivy.LogicEnumeratedSort:
		sortExpr := g.goIvySortExpr(s)
		var b strings.Builder
		b.WriteString("func() goivy.Expr { switch " + value + " {")
		for _, val := range st.Extension {
			b.WriteString("case " + goName(val) + ": return goivy.NewConst(" + strconv.Quote(val) + ", " + sortExpr + ");")
		}
		b.WriteString("}; return goivy.NewConst(strconv.Itoa(int(" + value + ")), " + sortExpr + ") }()")
		return b.String(), true
	default:
		if g.runtimeSolverSupportsScalarSort(s) {
			return fmt.Sprintf("goivy.NewConst(strconv.Itoa(int(%s)), %s)", value, g.goIvySortExpr(s)), true
		}
		return "", false
	}
}

func (g *Generator) emitRuntimeActionSolverAssignModelValue(w *goWriter, target string, s goivy.Sort, modelExpr string) {
	w.open("{")
	w.linef("__ivy_value, ok := %s.(*goivy.Const)", modelExpr)
	w.open("if !ok || __ivy_value == nil {")
	w.line("return false")
	w.close("")
	switch st := s.(type) {
	case *goivy.BooleanSort:
		_ = st
		w.open(`switch __ivy_value.Name {`)
		w.linef(`case "true":`)
		w.indent++
		w.linef("%s = true", target)
		w.indent--
		w.linef(`case "false":`)
		w.indent++
		w.linef("%s = false", target)
		w.indent--
		w.line("default:")
		w.indent++
		w.line("return false")
		w.indent--
		w.close("")
	case *goivy.LogicEnumeratedSort:
		w.open(`switch __ivy_value.Name {`)
		for _, val := range st.Extension {
			w.linef("case %q:", val)
			w.indent++
			w.linef("%s = %s", target, goName(val))
			w.indent--
		}
		w.line("default:")
		w.indent++
		w.line("return false")
		w.indent--
		w.close("")
	default:
		w.line("__ivy_num, err := strconv.Atoi(__ivy_value.Name)")
		w.open("if err != nil {")
		w.line("return false")
		w.close("")
		w.linef("%s = __ivy_num", target)
	}
	w.close("")
}
