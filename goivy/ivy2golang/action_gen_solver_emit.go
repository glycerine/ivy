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
	baseSMT    string
	assigns    []runtimeActionSolverAssign
	stateSyms  []stateSymbol
	needsState map[string]bool
}

type runtimeActionSolverPlanCacheEntry struct {
	plan *runtimeActionSolverPlan
	ok   bool
}

type runtimeActionSolverAssign struct {
	input      *goivy.Const
	target     string
	sort       goivy.Sort
	declare    bool
	mapped     goivy.Expr
	defined    bool
	noReadback bool
}

type runtimeActionSolverMember struct {
	name   string
	target string
	sort   goivy.Sort
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

func (g *Generator) initGeneratorUsesRuntimeSolver() bool {
	if g == nil || (g.Config.Target != "test" && g.Config.Target != "gen") {
		return false
	}
	constraints, err := g.initialStateConstraints()
	return err == nil && constraints != nil && len(constraints.Formulas) > 0
}

func (g *Generator) assignmentUsesRuntimeSolverSupport() bool {
	if g == nil {
		return false
	}
	for _, sym := range g.stateSymbols() {
		fs, ok := sym.Sort.(*goivy.LogicFunctionSort)
		if ok && len(fs.Domain()) > 0 && g.goFunctionStorageFor(fs.Domain(), fs.Range()).Large {
			return true
		}
	}
	return false
}

func (g *Generator) runtimeActionSolverPlan(name string, act goivy.Action) (res *runtimeActionSolverPlan, ok bool) {
	cacheKey := ""
	oldErrs := 0
	if g != nil {
		oldErrs = len(g.errs)
		cacheKey = fmt.Sprintf("%s\x00%p", name, act)
		if g.runtimeActionSolverPlanCache != nil {
			if entry, found := g.runtimeActionSolverPlanCache[cacheKey]; found {
				return entry.plan, entry.ok
			}
		}
	}
	defer func() {
		if recover() != nil {
			res = nil
			ok = false
		}
		if g != nil {
			if !ok && oldErrs <= len(g.errs) {
				g.errs = g.errs[:oldErrs]
			}
			if g.runtimeActionSolverPlanCache == nil {
				g.runtimeActionSolverPlanCache = make(map[string]runtimeActionSolverPlanCacheEntry)
			}
			g.runtimeActionSolverPlanCache[cacheKey] = runtimeActionSolverPlanCacheEntry{plan: res, ok: ok}
		}
	}()
	if g == nil || g.Mod == nil || act == nil {
		return nil, false
	}
	if len(act.GetFormalReturns()) > 0 && len(act.GetFormalParams()) == 0 {
		return nil, false
	}
	plan := g.buildActionGenPlan(name, act)
	if plan == nil || plan.fallback || plan.preFmla == nil {
		g.errs = g.errs[:oldErrs]
		return nil, false
	}
	baseSMT, ok := g.runtimeActionSolverSMTLIBAssertion(plan.preFmla)
	if !ok {
		g.errs = g.errs[:oldErrs]
		return nil, false
	}
	formals := map[string]*goivy.Const{}
	for _, p := range act.GetFormalParams() {
		if p == nil {
			continue
		}
		for _, alias := range formalExprOverrideNames(p.Name) {
			formals[alias] = p
		}
	}
	declaredTargets := map[string]bool{}
	for _, p := range act.GetFormalParams() {
		if p != nil {
			declaredTargets[goName(p.Name)] = true
		}
	}
	definedInputs := plan.defedParamSet()
	var assigns []runtimeActionSolverAssign
	hasNoReadback := false
	for _, in := range plan.inputs {
		if in == nil || strings.HasPrefix(in.Name, "__ts") || in.Name == "*>" || g.isDerivedStorageName(in.Name) {
			continue
		}
		if actionGenStrictDefIdx(in, plan) {
			continue
		}
		if mapped, found := plan.fsyms[goivy.Key(in)]; found && mapped != nil && !mapped.Equal(in) {
			if !g.runtimeSolverSupportsInputSort(in.CSort) {
				g.errs = g.errs[:oldErrs]
				return nil, false
			}
			noReadback := g.runtimeSolverInputNoReadback(in.CSort)
			hasNoReadback = hasNoReadback || noReadback
			assigns = append(assigns, runtimeActionSolverAssign{input: in, target: goName(in.Name), sort: in.CSort, mapped: mapped, defined: definedInputs[goivy.Key(in)], noReadback: noReadback})
			continue
		}
		if p, found := formals[in.Name]; found {
			if !g.runtimeSolverSupportsInputSort(p.CSort) {
				g.errs = g.errs[:oldErrs]
				return nil, false
			}
			noReadback := g.runtimeSolverInputNoReadback(p.CSort)
			hasNoReadback = hasNoReadback || noReadback
			assigns = append(assigns, runtimeActionSolverAssign{input: in, target: goName(p.Name), sort: p.CSort, defined: definedInputs[goivy.Key(in)], noReadback: noReadback})
			continue
		}
		if !g.runtimeSolverSupportsInputSort(in.CSort) {
			g.errs = g.errs[:oldErrs]
			return nil, false
		}
		target := goName(in.Name)
		declare := !declaredTargets[target]
		declaredTargets[target] = true
		noReadback := g.runtimeSolverInputNoReadback(in.CSort)
		hasNoReadback = hasNoReadback || noReadback
		assigns = append(assigns, runtimeActionSolverAssign{input: in, target: target, sort: in.CSort, declare: declare, defined: definedInputs[goivy.Key(in)], noReadback: noReadback})
	}
	if hasNoReadback && g.testActionNeedsTrial(act) {
		g.errs = g.errs[:oldErrs]
		return nil, false
	}
	preUsed := goivy.UsedSymbolsAst(plan.preFmla)
	defKeys := preDefinedKeys(plan.oldPreClauses)
	needsState := map[string]bool{}
	var stateSyms []stateSymbol
	for _, sym := range g.stateSymbols() {
		key := stateSymbolNodeKey(sym)
		if defKeys[key] || (!preUsedContainsStateSymbol(preUsed, sym) && !g.runtimeActionSolverNeedsRangeBoundParam(sym)) {
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
	return &runtimeActionSolverPlan{plan: plan, baseSMT: baseSMT, assigns: assigns, stateSyms: stateSyms, needsState: needsState}, true
}

func (g *Generator) runtimeActionSolverNeedsRangeBoundParam(sym stateSymbol) bool {
	if g == nil || g.Mod == nil || g.Mod.Sig == nil || sym.Name == "" || !g.isParamName(sym.Name) {
		return false
	}
	rs, ok := g.Mod.Sig.Interp[sortName(sym.Sort)].(*goivy.RangeSort)
	if !ok || rs == nil {
		return false
	}
	return rangeBoundUsesSymbol(rs.Lb, sym.Name) || rangeBoundUsesSymbol(rs.Ub, sym.Name)
}

func rangeBoundUsesSymbol(bound goivy.NumeralOrCompiledBound, name string) bool {
	cb, ok := bound.(goivy.CompiledBound)
	if !ok || cb.Expr == nil || name == "" {
		return false
	}
	for _, sym := range goivy.UsedSymbolsAst(cb.Expr).All() {
		if c, ok := sym.(*goivy.Const); ok && c.Name == name {
			return true
		}
	}
	return false
}

func (g *Generator) runtimeSolverSupportsInputSort(s goivy.Sort) bool {
	if fs, ok := s.(*goivy.LogicFunctionSort); ok {
		if len(fs.Domain()) == 0 {
			return g.runtimeSolverSupportsScalarSort(fs.Range())
		}
		if !g.runtimeSolverCanLoopDomain(fs.Domain()) {
			return g.runtimeSolverInputNoReadback(s)
		}
		return g.runtimeSolverSupportsScalarSort(fs.Range())
	}
	return g.runtimeSolverSupportsScalarSort(s)
}

func (g *Generator) runtimeSolverInputNoReadback(s goivy.Sort) bool {
	fs, ok := s.(*goivy.LogicFunctionSort)
	if !ok || len(fs.Domain()) == 0 || g.runtimeSolverCanLoopDomain(fs.Domain()) {
		return false
	}
	return g.goFunctionStorageFor(fs.Domain(), fs.Range()).Large && g.runtimeSolverSupportsScalarSort(fs.Range())
}

func (g *Generator) runtimeSolverSupportsScalarSort(s goivy.Sort) bool {
	return g.runtimeSolverSupportsScalarSortSeen(s, map[string]bool{})
}

func (g *Generator) runtimeSolverSupportsScalarSortSeen(s goivy.Sort, seen map[string]bool) bool {
	if s == nil {
		return false
	}
	if _, isFn := s.(*goivy.LogicFunctionSort); isFn {
		return false
	}
	if g.runtimeSolverSupportsRecordSortSeen(s, seen) {
		return true
	}
	if g.runtimeSolverSupportsVariantSortSeen(s, seen) {
		return true
	}
	if g.isVariantSuperName(sortName(s)) {
		return false
	}
	switch s.(type) {
	case *goivy.BooleanSort, *goivy.LogicEnumeratedSort, *goivy.RangeSort:
		return true
	case *goivy.UninterpretedSort:
		if g.hasStrBVInterp(s) {
			return true
		}
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

func (g *Generator) runtimeSolverSupportsRecordSortSeen(s goivy.Sort, seen map[string]bool) bool {
	name := sortName(s)
	if name == "" {
		return false
	}
	fields := g.destructorStructFieldInfos(name)
	if len(fields) == 0 {
		return false
	}
	if seen[name] {
		return false
	}
	seen[name] = true
	defer delete(seen, name)
	for _, field := range fields {
		if field.Sort == nil {
			return false
		}
		domain := field.Sort.Domain()
		if len(domain) == 0 || !g.runtimeSolverCanLoopDomain(domain[1:]) {
			return false
		}
		if !g.runtimeSolverSupportsScalarSortSeen(field.Sort.Range(), seen) {
			return false
		}
	}
	return true
}

func (g *Generator) runtimeSolverSupportsVariantSortSeen(s goivy.Sort, seen map[string]bool) bool {
	name := sortName(s)
	if name == "" || !g.isVariantSuperName(name) || g.Mod == nil {
		return false
	}
	if seen[name] {
		return false
	}
	seen[name] = true
	defer delete(seen, name)
	variants := g.Mod.Variants[name]
	if len(variants) == 0 {
		return false
	}
	for _, sub := range variants {
		if !g.runtimeSolverSupportsScalarSortSeen(sub, seen) {
			return false
		}
	}
	return true
}

func (g *Generator) runtimeSolverSupportsStateSymbol(sym stateSymbol) bool {
	if fs, ok := sym.Sort.(*goivy.LogicFunctionSort); ok {
		if g.runtimeSolverSupportsSparseThunkStateSymbol(sym, fs) {
			return true
		}
		if !g.runtimeSolverSupportsScalarSort(fs.Range()) {
			return false
		}
		return g.runtimeSolverCanLoopDomain(fs.Domain())
	}
	return g.runtimeSolverSupportsScalarSort(sym.Sort)
}

func (g *Generator) runtimeSolverCanLoopDomain(domain []goivy.Sort) bool {
	for _, d := range domain {
		if _, ok := g.runtimeSolverLoopHeaderForSort(d, "__ivy_probe"); ok {
			continue
		}
		return false
	}
	return true
}

func (g *Generator) runtimeSolverLoopHeaderForSort(s goivy.Sort, name string) (string, bool) {
	if g.hasStrBVInterp(s) {
		card := g.sortCard(s)
		if card <= 0 {
			return "", false
		}
		idx := name + "_idx"
		return fmt.Sprintf("for %s := 0; %s < %d; %s++ {", idx, idx, card, idx), true
	}
	header, ok, err := g.goLoopHeaderForSort(s, name)
	if err != nil || !ok {
		return "", false
	}
	return header, true
}

func (g *Generator) runtimeSolverSupportsSparseThunkStateSymbol(sym stateSymbol, fs *goivy.LogicFunctionSort) bool {
	if fs == nil || len(fs.Domain()) == 0 {
		return false
	}
	if !g.goFunctionStorageFor(fs.Domain(), fs.Range()).Large {
		return false
	}
	for _, d := range fs.Domain() {
		if !g.runtimeSolverSupportsScalarSort(d) {
			return false
		}
	}
	if !g.runtimeSolverSupportsScalarSort(fs.Range()) {
		return false
	}
	return true
}

func (g *Generator) emitRuntimeSolverSupport(w *goWriter) {
	if !g.actionGeneratorUsesRuntimeSolver() && !g.initGeneratorUsesRuntimeSolver() && !g.assignmentUsesRuntimeSolverSupport() {
		return
	}
	if g.actionGeneratorUsesRuntimeSolver() {
		w.open("func __ivy_solver_choice_symbol(name string) string {")
		w.line("normalize := func(leaf string) (string, bool) {")
		w.indent++
		w.line(`parts := strings.Split(leaf, ":")`)
		w.open("if len(parts) >= 2 {")
		w.line(`payload := strings.ReplaceAll(parts[1], ".", "__")`)
		w.open("switch parts[0] {")
		w.line(`case "__fml", "__loc", "__ret":`)
		w.indent++
		w.line(`return parts[0] + ":" + payload, true`)
		w.indent--
		w.line(`case "fml", "loc", "ret":`)
		w.indent++
		w.line(`return "__" + parts[0] + ":" + payload, true`)
		w.indent--
		w.close("")
		w.close("")
		w.line(`return leaf, false`)
		w.indent--
		w.line("}")
		w.open("if sym, ok := normalize(name); ok {")
		w.line("return sym")
		w.close("")
		w.line("leaf := name")
		w.open(`if idx := strings.LastIndex(leaf, "."); idx >= 0 {`)
		w.line("leaf = leaf[idx+1:]")
		w.close("")
		w.open("if sym, ok := normalize(leaf); ok {")
		w.line("return sym")
		w.close("")
		w.line("return leaf")
		w.close("")
		w.blank()
	}
	w.open("func __ivy_solver_bool_expr(v bool) goivy.Expr {")
	w.open("if v {")
	w.line("return goivy.True")
	w.close("")
	w.line("return goivy.False")
	w.close("")
	w.blank()
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
	g.emitRuntimeSolverModuleMetadata(w)
	for _, s := range g.runtimeSolverSorts() {
		if sortName(s) == "bool" {
			continue
		}
		w.linef("_ = mod.Sig.AddSort(%s)", g.goIvySortExpr(s))
	}
	for _, interp := range g.runtimeSolverInterps() {
		w.linef("mod.Sig.Interp[%q] = %s", interp.name, interp.valueExpr)
	}
	for _, ctor := range g.runtimeSolverConstructors() {
		w.linef("mod.Sig.Constructors[%q] = true", ctor)
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

func (g *Generator) emitRuntimeSolverModuleMetadata(w *goWriter) {
	if g == nil || g.Mod == nil {
		return
	}
	if g.Mod.Sig != nil {
		if g.Mod.Sig.DefaultSort != nil {
			w.linef("mod.Sig.DefaultSort = %s", g.goIvySortExpr(g.Mod.Sig.DefaultSort))
		}
		if g.Mod.Sig.DefaultNumericSort != nil {
			w.linef("mod.Sig.DefaultNumericSort = %s", g.goIvySortExpr(g.Mod.Sig.DefaultNumericSort))
		}
		if g.Mod.Sig.AllowUnsorted {
			w.line("mod.Sig.AllowUnsorted = true")
		}
	}
	if len(g.Mod.SortOrder) > 0 {
		parts := make([]string, len(g.Mod.SortOrder))
		for i, name := range g.Mod.SortOrder {
			parts[i] = strconv.Quote(name)
		}
		w.linef("mod.SortOrder = []string{%s}", strings.Join(parts, ", "))
	}
	if len(g.Mod.Variants) > 0 {
		names := make([]string, 0, len(g.Mod.Variants))
		for name := range g.Mod.Variants {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			variants := g.Mod.Variants[name]
			parts := make([]string, 0, len(variants))
			for _, sub := range variants {
				parts = append(parts, g.goIvySortExpr(sub))
			}
			w.linef("mod.Variants[%q] = []goivy.Sort{%s}", name, strings.Join(parts, ", "))
		}
	}
	if g.Mod.Supertypes != nil {
		supertypes := map[string]goivy.Sort{}
		for name, super := range g.Mod.Supertypes.All() {
			supertypes[name] = super
		}
		names := make([]string, 0, len(supertypes))
		for name := range supertypes {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			w.linef("mod.Supertypes.Set(%q, %s)", name, g.goIvySortExpr(supertypes[name]))
		}
	}
	if g.Mod.SortConstructors != nil {
		sortConstructors := map[string][]*goivy.Const{}
		for name, ctors := range g.Mod.SortConstructors.All() {
			sortConstructors[name] = append([]*goivy.Const(nil), ctors...)
		}
		names := make([]string, 0, len(sortConstructors))
		for name := range sortConstructors {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			ctors := sortConstructors[name]
			sort.SliceStable(ctors, func(i, j int) bool {
				if ctors[i] == nil {
					return ctors[j] != nil
				}
				if ctors[j] == nil {
					return false
				}
				return ctors[i].Name < ctors[j].Name
			})
			parts := make([]string, 0, len(ctors))
			for _, ctor := range ctors {
				if ctor == nil {
					continue
				}
				parts = append(parts, fmt.Sprintf("goivy.NewConst(%q, %s)", ctor.Name, g.goIvySortExpr(ctor.CSort)))
			}
			w.linef("mod.SortConstructors.Set(%q, []*goivy.Const{%s})", name, strings.Join(parts, ", "))
		}
	}
}

func (g *Generator) emitRuntimeInitGeneratorGenerate(w *goWriter) bool {
	constraints, err := g.initialStateConstraints()
	if err != nil || constraints == nil || len(constraints.Formulas) == 0 {
		return false
	}
	clausesExpr, useClauses := g.runtimeInitialClausesExpr(constraints)
	baseSMT := ""
	if !useClauses {
		baseSMT, err = g.runtimeInitialSMTLIBAssertion(constraints)
	}
	if err != nil {
		return false
	}
	w.line("solver := goivy.NewSolver(__ivy_solver_module(), nil)")
	if useClauses {
		w.linef("clauses := %s", clausesExpr)
	} else {
		w.linef("baseSMT := %s", strconv.Quote(baseSMT))
	}
	w.line("soft := []goivy.Expr{}")
	used := constraints.Used
	for _, sym := range g.stateSymbols() {
		if g.isParamName(sym.Name) {
			if useClauses && used[sym.Name] {
				g.emitRuntimeActionSolverStateEquality(w, sym)
			}
			continue
		}
		if used[sym.Name] {
			g.emitRuntimeInitSoftRandomizeSymbol(w, sym)
			continue
		}
		g.emitRandomizeSymbolWithChooser(w, sym, "init", "___ivy_rand")
	}
	if useClauses {
		w.line("model, err := solver.GetModelClausesWithSoftAssumptionsLogged(clauses, soft, func(n int) int { if n <= 0 { return 0 }; return ivyRand31() % n }, ivySoftAssumptionModelLog)")
	} else {
		w.line("model, err := solver.GetModelSMTLIBWithSoftAssumptionsLogged(baseSMT, soft, func(n int) int { if n <= 0 { return 0 }; return ivyRand31() % n }, ivySoftAssumptionModelLog)")
	}
	w.open("if err != nil || model == nil {")
	w.line("ivy.___ivy_gen = gen")
	w.line("ivy.__init()")
	w.line("return false")
	w.close("")
	g.emitRuntimeSolverSatModelLog(w)
	w.line("hm := goivy.NewHerbrandModel(solver, model.Solver, model.Model, model.Vocab)")
	w.line("_ = hm")
	for _, sym := range g.stateSymbols() {
		if g.isParamName(sym.Name) || !used[sym.Name] {
			continue
		}
		g.emitRuntimeInitEvalStateSymbol(w, sym)
	}
	w.line("ivy.___ivy_gen = gen")
	w.line("ivy.__init()")
	w.line("return true")
	return true
}

func (g *Generator) runtimeInitialClausesExpr(constraints *initialStateConstraints) (string, bool) {
	if constraints == nil {
		return "goivy.NewClauses(nil, nil, nil)", true
	}
	parts := make([]string, 0, len(constraints.Formulas))
	for _, f := range constraints.Formulas {
		expr, ok := g.goIvyExprExpr(closeFormulaForSMTLIB(f))
		if !ok {
			return "", false
		}
		parts = append(parts, expr)
	}
	if len(parts) == 0 {
		return "goivy.NewClauses(nil, nil, nil)", true
	}
	return fmt.Sprintf("goivy.NewClauses([]goivy.Expr{%s}, nil, nil)", strings.Join(parts, ", ")), true
}

func (g *Generator) runtimeInitialSMTLIBAssertion(constraints *initialStateConstraints) (string, error) {
	if constraints == nil {
		return "(assert true)", nil
	}
	var smts []string
	if g.Mod == nil || g.Mod.InitCond == nil || g.Mod.InitCond.IsTrue() {
		smts = append(smts, "true")
	}
	for _, f := range constraints.Formulas {
		smt, err := g.formulaToSmtlibErr(closeFormulaForSMTLIB(f))
		if err != nil {
			return "", err
		}
		smts = append(smts, smt)
	}
	return "(assert (and\n  " + strings.Join(smts, "\n  ") + "\n))", nil
}

func (g *Generator) formulaToSmtlibErr(fmla goivy.Expr) (string, error) {
	if fmla == nil {
		return "true", nil
	}
	solver := goivy.NewSolver(g.Mod, nil)
	z3expr, err := solver.FormulaToZ3(fmla)
	if err != nil {
		return "", err
	}
	return cleanSmtlib(z3expr.String()), nil
}

func (g *Generator) formulaToSmtlibWithTypeConstraints(fmla goivy.Expr) (string, bool) {
	if fmla == nil || goivy.IsTrue(fmla) {
		return "true", true
	}
	solver := goivy.NewSolver(g.Mod, nil)
	z3expr, err := solver.ClausesToZ3(goivy.NewClauses([]goivy.Expr{fmla}, nil, nil))
	if err != nil {
		return "", false
	}
	return cleanSmtlib(z3expr.String()), true
}

func (g *Generator) runtimeActionSolverSMTLIBAssertion(fmla goivy.Expr) (string, bool) {
	smt, ok := g.formulaToSmtlibWithTypeConstraints(fmla)
	if !ok {
		return "", false
	}
	return "(assert " + smt + ")", true
}

func (g *Generator) emitRuntimeSolverCheckSatLog(w *goWriter) {
	w.open("if __ivy_modelfile != nil {")
	w.line(`ivyModelLogf("(check-sat")`)
	w.open("for __ivy_alit := range soft {")
	w.line(`ivyModelLogf(" alit:%d", __ivy_alit)`)
	w.close("")
	w.line(`ivyModelLogf(")\n")`)
	w.close("")
}

func (g *Generator) emitRuntimeSolverSatModelLog(w *goWriter) {
	if g == nil || g.Config.Target != "test" {
		return
	}
	w.open("if model != nil && model.Solver != nil {")
	w.line(`ivyModelLogf("%s", model.String())`)
	w.close("")
}

func cleanSmtlib(s string) string {
	s = strings.ReplaceAll(s, "|!1", "!1|")
	s = strings.ReplaceAll(s, `\|`, "")
	return s
}

func closeFormulaForSMTLIB(f goivy.Expr) goivy.Expr {
	if f == nil {
		return nil
	}
	return goivy.IvyForAll(goivy.FreeVariablesList(f), f)
}

func (g *Generator) emitRuntimeInitSoftRandomizeSymbol(w *goWriter, sym stateSymbol) {
	if fs, ok := sym.Sort.(*goivy.LogicFunctionSort); ok && len(fs.Domain()) > 0 {
		g.emitRuntimeInitSoftRandomizeFunction(w, sym, fs, nil, nil)
		return
	}
	lhs := fmt.Sprintf("goivy.NewConst(%q, %s)", sym.Name, g.goIvySortExpr(sym.Sort))
	g.emitRuntimeInitSoftRandomizeTerm(w, lhs, sym.Sort, sym.Name, 0)
}

func (g *Generator) emitRuntimeInitSoftRandomizeFunction(w *goWriter, sym stateSymbol, fs *goivy.LogicFunctionSort, goArgs []string, ivyArgs []string) {
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
			g.emitRuntimeInitSoftRandomizeFunction(
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
	label := sym.Name
	if len(goArgs) > 0 {
		label += "." + strings.Join(goArgs, ".")
	}
	g.emitRuntimeInitSoftRandomizeTerm(w, lhs, fs.Range(), label, int64(len(goArgs)))
}

func (g *Generator) emitRuntimeInitSoftRandomizeTerm(w *goWriter, lhs string, sort goivy.Sort, label string, id int64) {
	randExpr, err := g.goActionParamRandomValueExpr(sort, label, id)
	if err != nil {
		g.unsupported(w, "unsupported runtime init random input %s: %s", label, strings.TrimPrefix(err.Error(), "ivy2golang: "))
		w.line("return false")
		return
	}
	tmp := g.nextTemp("__ivy_init_rand")
	w.linef("%s := %s", tmp, randExpr)
	if !g.emitRuntimeActionSolverAppendValueEquality(w, "soft", lhs, tmp, sort) {
		w.line("return false")
		return
	}
}

func (g *Generator) emitRuntimeInitEvalStateSymbol(w *goWriter, sym stateSymbol) {
	if fs, ok := sym.Sort.(*goivy.LogicFunctionSort); ok && len(fs.Domain()) > 0 {
		g.emitRuntimeInitEvalFunctionStateSymbol(w, sym, fs, nil, nil)
		return
	}
	term := fmt.Sprintf("goivy.NewConst(%q, %s)", sym.Name, g.goIvySortExpr(sym.Sort))
	g.emitRuntimeActionSolverAssignModelTermValue(w, "ivy."+goName(sym.Name), sym.Sort, term)
}

func (g *Generator) emitRuntimeInitEvalFunctionStateSymbol(w *goWriter, sym stateSymbol, fs *goivy.LogicFunctionSort, goArgs []string, ivyArgs []string) {
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
			g.emitRuntimeInitEvalFunctionStateSymbol(
				w,
				sym,
				fs,
				append(append([]string(nil), goArgs...), value),
				append(append([]string(nil), ivyArgs...), argExpr),
			)
		}
		return
	}
	term := fmt.Sprintf("goivy.MustApply(goivy.NewConst(%q, %s), %s)", sym.Name, g.goIvySortExpr(sym.Sort), strings.Join(ivyArgs, ", "))
	target := g.goStorageAccess(sym.Name, sym.Sort, goArgs, "ivy")
	g.emitRuntimeActionSolverAssignModelTermValue(w, target, fs.Range(), term)
}

type runtimeSolverInterp struct {
	name      string
	valueExpr string
}

func (g *Generator) runtimeSolverInterps() []runtimeSolverInterp {
	if g == nil || g.Mod == nil || g.Mod.Sig == nil {
		return nil
	}
	seen := map[string]bool{}
	out := make([]runtimeSolverInterp, 0, len(g.Mod.Sig.Interp)+len(g.Mod.NativeTypes))
	for name, value := range g.Mod.Sig.Interp {
		expr, ok := g.goIvyInterpValueExpr(value)
		if !ok {
			continue
		}
		out = append(out, runtimeSolverInterp{name: name, valueExpr: expr})
		seen[name] = true
	}
	for name, nt := range g.Mod.NativeTypes {
		if seen[name] || !g.nativeTypeNameTranslatesAsGoInt(name, nt) {
			continue
		}
		out = append(out, runtimeSolverInterp{name: name, valueExpr: strconv.Quote("int")})
		seen[name] = true
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].name < out[j].name })
	return out
}

func (g *Generator) runtimeSolverConstructors() []string {
	if g == nil || g.Mod == nil || g.Mod.Sig == nil {
		return nil
	}
	out := make([]string, 0, len(g.Mod.Sig.Constructors))
	for name, ok := range g.Mod.Sig.Constructors {
		if ok {
			out = append(out, name)
		}
	}
	sort.Strings(out)
	return out
}

func (g *Generator) goIvyInterpValueExpr(value any) (string, bool) {
	switch v := value.(type) {
	case string:
		return strconv.Quote(v), true
	case goivy.Sort:
		return g.goIvySortExpr(v), true
	default:
		return "", false
	}
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
	for _, name := range g.runnableActionNamesFrom(g.publicActionNamesSorted()) {
		act, _ := g.Mod.Actions.Get2(name)
		rsp, ok := g.runtimeActionSolverPlan(name, g.actionGeneratorAnalysisAction(name, act))
		if !ok || rsp == nil || rsp.plan == nil {
			continue
		}
		for _, input := range rsp.plan.inputs {
			if input != nil {
				add(input.CSort)
			}
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return sortName(out[i]) < sortName(out[j]) })
	return out
}

func (g *Generator) runtimeSolverSymbols() []*goivy.Const {
	seen := map[string]bool{}
	var out []*goivy.Const
	add := func(c *goivy.Const) {
		if c == nil || c.Name == "" || c.CSort == nil {
			return
		}
		key := c.Name + "\x00" + string(c.CSort.Sexp())
		if seen[key] {
			return
		}
		seen[key] = true
		out = append(out, c)
	}
	for _, c := range g.Mod.Sig.AllSymbols() {
		add(c)
	}
	for _, name := range g.runnableActionNamesFrom(g.publicActionNamesSorted()) {
		act, _ := g.Mod.Actions.Get2(name)
		rsp, ok := g.runtimeActionSolverPlan(name, g.actionGeneratorAnalysisAction(name, act))
		if !ok || rsp == nil || rsp.plan == nil {
			continue
		}
		for _, input := range rsp.plan.inputs {
			add(input)
		}
	}
	if g.Mod.SortDestructors != nil {
		for _, destrs := range g.Mod.SortDestructors.All() {
			for _, d := range destrs {
				add(goivy.NewConst(d.Name, d.CSort))
			}
		}
	}
	if g.Mod.Variants != nil {
		names := make([]string, 0, len(g.Mod.Variants))
		for name := range g.Mod.Variants {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			super, ok := g.Mod.Sig.Sorts.Get2(name)
			if !ok || super == nil {
				super = &goivy.UninterpretedSort{Name: name}
			}
			for _, sub := range g.Mod.Variants[name] {
				relSort, err := goivy.NewFunctionSort(super, sub, goivy.Boolean)
				if err != nil {
					continue
				}
				add(goivy.NewConst("*>", relSort))
			}
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Name != out[j].Name {
			return out[i].Name < out[j].Name
		}
		return string(out[i].CSort.Sexp()) < string(out[j].CSort.Sexp())
	})
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
		return fmt.Sprintf("&goivy.RangeSort{Name: %q, Lb: %s, Ub: %s}", st.Name, g.goIvyRangeBoundExpr(st.Lb, st.Name), g.goIvyRangeBoundExpr(st.Ub, st.Name))
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

func (g *Generator) goIvyRangeBoundExpr(b goivy.NumeralOrCompiledBound, parentName string) string {
	switch bound := b.(type) {
	case goivy.NumeralBound:
		if bound.Sort != nil {
			return fmt.Sprintf("goivy.NumeralBound{Value: %q, Sort: %s}", bound.Value, g.goIvyRangeBoundSortExpr(bound.Sort, parentName))
		}
		return fmt.Sprintf("goivy.NumeralBound{Value: %q}", bound.Value)
	case goivy.CompiledBound:
		return fmt.Sprintf("goivy.CompiledBound{Expr: %s}", g.goIvyRangeBoundCompiledExpr(bound.Expr, parentName))
	default:
		if b == nil {
			return `goivy.NumeralBound{Value: "0"}`
		}
		return fmt.Sprintf("goivy.NumeralBound{Value: %q}", b.BoundString())
	}
}

func (g *Generator) goIvyRangeBoundSortExpr(s goivy.Sort, parentName string) string {
	if s == nil {
		return "nil"
	}
	if sortName(s) == parentName {
		return fmt.Sprintf("&goivy.UninterpretedSort{Name: %q}", parentName)
	}
	return g.goIvySortExpr(s)
}

func (g *Generator) goIvyRangeBoundCompiledExpr(e goivy.Expr, parentName string) string {
	switch n := e.(type) {
	case *goivy.Const:
		return fmt.Sprintf("goivy.NewConst(%q, %s)", n.Name, g.goIvyRangeBoundSortExpr(n.CSort, parentName))
	case *goivy.LogicVariable:
		return fmt.Sprintf("&goivy.LogicVariable{Name: %q, VSort: %s}", n.Name, g.goIvyRangeBoundSortExpr(n.VSort, parentName))
	case *goivy.Apply:
		fn := g.goIvyRangeBoundCompiledExpr(n.Func, parentName)
		args := make([]string, len(n.Terms))
		for i, term := range n.Terms {
			args[i] = g.goIvyRangeBoundCompiledExpr(term, parentName)
		}
		if len(args) == 0 {
			return fn
		}
		return fmt.Sprintf("goivy.MustApply(%s, %s)", fn, strings.Join(args, ", "))
	default:
		if expr, ok := g.goIvyExprExpr(e); ok {
			return expr
		}
		return fmt.Sprintf("goivy.NewConst(%q, &goivy.UninterpretedSort{Name: %q})", fmt.Sprint(e), parentName)
	}
}

func (g *Generator) goIvyExprExpr(e goivy.Expr) (string, bool) {
	if repl, ok := g.aliasForSymbol(e); ok {
		return g.goIvyExprExpr(repl)
	}
	switch n := e.(type) {
	case *goivy.Const:
		if code, ok := g.exprOverride(n.Name); ok {
			return code, true
		}
		return fmt.Sprintf("goivy.NewConst(%q, %s)", n.Name, g.goIvySortExpr(n.CSort)), true
	case *goivy.LogicVariable:
		if code, ok := g.exprOverride(n.Name); ok {
			return code, true
		}
		return fmt.Sprintf("&goivy.LogicVariable{Name: %q, VSort: %s}", n.Name, g.goIvySortExpr(n.VSort)), true
	case *goivy.LogicLet:
		if expanded, ok := actionGeneratorExpandLetExpr(n); ok {
			return g.goIvyExprExpr(expanded)
		}
		return "", false
	case *goivy.LogicDefinition:
		lhs, ok := g.goIvyExprExpr(n.Lhs)
		if !ok {
			return "", false
		}
		rhs, ok := g.goIvyExprExpr(n.Rhs)
		if !ok {
			return "", false
		}
		return fmt.Sprintf("goivy.NewDefinition(%s, %s)", lhs, rhs), true
	case *goivy.SomeCondition:
		if lowered, ok := preimageSomeConditionAsExists(n); ok {
			return g.goIvyExprExpr(lowered)
		}
		return "", false
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
	baseSMT := rsp.baseSMT
	if baseSMT == "" {
		var ok bool
		baseSMT, ok = g.runtimeActionSolverSMTLIBAssertion(rsp.plan.preFmla)
		if !ok {
			return
		}
	}
	w.open(fmt.Sprintf("func (gen *%s) __ivy_generate_with_solver() bool {", typeName))
	w.line("ivy := gen.ivy")
	w.open("if ivy == nil {")
	w.line("return false")
	w.close("")
	w.line("defer func() { ivy.___ivy_gen = gen }()")
	w.open("if gen.__ivy_solver == nil {")
	w.line("gen.__ivy_solver = goivy.NewSolver(__ivy_solver_module(), nil)")
	w.close("")
	w.linef("baseSMT := %s", strconv.Quote(baseSMT))
	w.line("solver := gen.__ivy_solver")
	w.open("if gen.__ivy_solver_base == nil {")
	w.line("var err error")
	w.line("gen.__ivy_solver_base, err = solver.NewSMTLIBBaseSolver(baseSMT)")
	w.open("if err != nil {")
	w.line("return false")
	w.close("")
	w.close("")
	w.open("if gen.__ivy_solver_pre == nil {")
	w.line("gen.__ivy_solver_pre = goivy.NewClauses(nil, nil, nil)")
	w.close("")
	w.line("clauses := gen.__ivy_solver_pre.Copy()")
	for _, sym := range rsp.stateSyms {
		g.emitRuntimeActionSolverStateEquality(w, sym)
	}
	w.line("soft := []goivy.Expr{}")
	for i, assign := range rsp.assigns {
		g.emitRuntimeActionSolverRandomInputEquality(w, rsp, assign, i)
	}
	w.line("model, err := solver.GetModelSMTLIBBaseClausesWithSoftAssumptionsLogged(gen.__ivy_solver_base, clauses, soft, func(n int) int { if n <= 0 { return 0 }; return ivyRand31() % n }, ivySoftAssumptionModelLog)")
	w.open("if err != nil || model == nil {")
	w.line("return false")
	w.close("")
	g.emitRuntimeSolverSatModelLog(w)
	if len(rsp.assigns) > 0 {
		w.line("hm := goivy.NewHerbrandModel(solver, model.Solver, model.Model, model.Vocab)")
		w.line("_ = hm")
		for _, assign := range rsp.assigns {
			inputExpr, ok := g.goIvyExprExpr(assign.input)
			if !ok {
				w.line("return false")
				continue
			}
			g.emitRuntimeActionSolverAssignModelToTarget(w, rsp, assign, inputExpr)
		}
	}
	g.emitRuntimeActionSolverDefinedInputs(w, rsp)
	w.line("return true")
	w.close("")
	w.blank()
	g.emitRuntimeActionSolverChooseMethod(w, typeName, rsp)
}

func (g *Generator) emitRuntimeActionSolverDefinedInputs(w *goWriter, rsp *runtimeActionSolverPlan) {
	if rsp == nil || rsp.plan == nil || len(rsp.plan.paramDefs) == 0 {
		return
	}
	overrides := g.runtimeActionSolverExprOverrides(rsp)
	g.pushExprOverrides(overrides)
	defer g.popExprOverrides()
	declaredGeneratedLocals := map[string]bool{}
	neededGeneratedLocals := g.runtimeActionSolverNeededGeneratedLocalNames(rsp)
	for _, def := range rsp.plan.paramDefs {
		lhs, rhs, ok := runtimeActionSolverDefinedInputTerms(def)
		if !ok {
			continue
		}
		lhs = g.runtimeActionSolverMappedDefinedInputLHS(rsp, lhs)
		if name, ok := runtimeActionSolverGeneratedLocalDefinitionName(lhs); ok && !neededGeneratedLocals[name] {
			continue
		}
		lhsExpr, err := g.emitExpr(lhs)
		if err != nil {
			w.line("return false")
			continue
		}
		rhsExpr, err := g.emitExpr(rhs)
		if err != nil {
			w.line("return false")
			continue
		}
		g.emitRuntimeActionSolverDefinedLocalDecl(w, declaredGeneratedLocals, lhs, lhsExpr)
		if call, ok, err := g.goStorageSet(lhs, rhsExpr); ok || err != nil {
			if err != nil {
				w.line("return false")
				continue
			}
			w.line(call)
			continue
		}
		w.linef("%s = %s", lhsExpr, rhsExpr)
	}
}

func (g *Generator) runtimeActionSolverMappedDefinedInputLHS(rsp *runtimeActionSolverPlan, lhs goivy.Expr) goivy.Expr {
	if rsp == nil || rsp.plan == nil || lhs == nil {
		return lhs
	}
	if c, ok := exprAsConst(lhs); ok {
		if mapped, found := rsp.plan.fsyms[goivy.Key(c)]; found {
			return mapped
		}
	}
	return lhs
}

func (g *Generator) runtimeActionSolverNeededGeneratedLocalNames(rsp *runtimeActionSolverPlan) map[string]bool {
	out := map[string]bool{}
	if rsp == nil || rsp.plan == nil {
		return out
	}
	type defInfo struct {
		lhs  goivy.Expr
		rhs  goivy.Expr
		name string
		gen  bool
	}
	var defs []defInfo
	for _, def := range rsp.plan.paramDefs {
		lhs, rhs, ok := runtimeActionSolverDefinedInputTerms(def)
		if !ok {
			continue
		}
		lhs = g.runtimeActionSolverMappedDefinedInputLHS(rsp, lhs)
		name, gen := runtimeActionSolverGeneratedLocalDefinitionName(lhs)
		defs = append(defs, defInfo{lhs: lhs, rhs: rhs, name: name, gen: gen})
		if gen {
			continue
		}
		runtimeActionSolverAddUsedGeneratedLocalNames(out, lhs, "")
		runtimeActionSolverAddUsedGeneratedLocalNames(out, rhs, "")
	}
	changed := true
	for changed {
		changed = false
		for _, def := range defs {
			if !def.gen || !out[def.name] {
				continue
			}
			before := len(out)
			runtimeActionSolverAddUsedGeneratedLocalNames(out, def.lhs, def.name)
			runtimeActionSolverAddUsedGeneratedLocalNames(out, def.rhs, "")
			changed = changed || len(out) != before
		}
	}
	return out
}

func runtimeActionSolverGeneratedLocalDefinitionName(lhs goivy.Expr) (string, bool) {
	c, ok := runtimeActionSolverDefinitionHeadConst(lhs)
	if !ok || c == nil || !runtimeActionSolverGeneratedLocalName(c.Name) {
		return "", false
	}
	return c.Name, true
}

func runtimeActionSolverDefinitionHeadConst(lhs goivy.Expr) (*goivy.Const, bool) {
	if c, ok := exprAsConst(lhs); ok {
		return c, true
	}
	if app, ok := lhs.(*goivy.Apply); ok && app != nil {
		return exprAsConst(app.Func)
	}
	return nil, false
}

func runtimeActionSolverAddUsedGeneratedLocalNames(out map[string]bool, expr goivy.Expr, exclude string) {
	if expr == nil {
		return
	}
	for _, sym := range goivy.UsedSymbolsAst(expr).All() {
		c, ok := sym.(*goivy.Const)
		if !ok || c == nil || c.Name == exclude || !runtimeActionSolverGeneratedLocalName(c.Name) {
			continue
		}
		out[c.Name] = true
	}
}

func runtimeActionSolverGeneratedLocalName(name string) bool {
	return strings.HasPrefix(name, "__new_loc:") || actionGenGeneratedLocalName(name)
}

func (g *Generator) emitRuntimeActionSolverDefinedLocalDecl(w *goWriter, declared map[string]bool, lhs goivy.Expr, lhsExpr string) {
	if lhsExpr == "" || declared[lhsExpr] {
		return
	}
	c, ok := exprAsConst(lhs)
	if !ok || c == nil || !runtimeActionSolverGeneratedLocalName(c.Name) {
		return
	}
	sort := c.CSort
	if sort == nil {
		sort = lhs.NodeSort()
	}
	w.linef("var %s %s", lhsExpr, g.goType(sort))
	declared[lhsExpr] = true
}

func (g *Generator) runtimeActionSolverExprOverrides(rsp *runtimeActionSolverPlan) map[string]string {
	overrides := g.runtimeActionSolverBaseExprOverrides(rsp)
	if rsp == nil {
		return overrides
	}
	for _, assign := range rsp.assigns {
		if assign.input == nil || assign.target == "" {
			continue
		}
		target, ok := g.runtimeActionSolverAssignTargetExpr(rsp, assign)
		if !ok {
			target = "gen." + assign.target
		}
		for _, name := range formalExprOverrideNames(assign.input.Name) {
			overrides[name] = target
		}
	}
	return overrides
}

func (g *Generator) runtimeActionSolverBaseExprOverrides(rsp *runtimeActionSolverPlan) map[string]string {
	overrides := g.runtimeActionSolverFormalExprOverrides(rsp)
	for _, member := range g.runtimeActionSolverGeneratedMembers(rsp) {
		if member.target == "" {
			continue
		}
		target := "gen." + member.target
		overrides[member.target] = target
		for _, name := range formalExprOverrideNames(member.name) {
			overrides[name] = target
		}
	}
	if rsp != nil && rsp.plan != nil {
		for _, input := range rsp.plan.inputs {
			if input == nil {
				continue
			}
			mapped, found := rsp.plan.fsyms[goivy.Key(input)]
			if !found || mapped == nil || mapped.Equal(input) {
				continue
			}
			g.pushExprOverrides(overrides)
			target, err := g.emitExpr(mapped)
			g.popExprOverrides()
			if err != nil || target == "" {
				continue
			}
			for _, name := range formalExprOverrideNames(input.Name) {
				overrides[name] = target
			}
		}
	}
	return overrides
}

func (g *Generator) runtimeActionSolverFormalExprOverrides(rsp *runtimeActionSolverPlan) map[string]string {
	overrides := map[string]string{}
	if rsp == nil || rsp.plan == nil || rsp.plan.act == nil {
		return overrides
	}
	for _, p := range rsp.plan.act.GetFormalParams() {
		if p == nil {
			continue
		}
		target := "gen." + goName(p.Name)
		for _, name := range formalExprOverrideNames(p.Name) {
			overrides[name] = target
		}
	}
	return overrides
}

func (g *Generator) runtimeActionSolverAssignTargetExpr(rsp *runtimeActionSolverPlan, assign runtimeActionSolverAssign) (string, bool) {
	if assign.mapped == nil {
		return "gen." + assign.target, assign.target != ""
	}
	g.pushExprOverrides(g.runtimeActionSolverBaseExprOverrides(rsp))
	defer g.popExprOverrides()
	expr, err := g.emitExpr(assign.mapped)
	if err != nil || expr == "" {
		return "", false
	}
	return expr, true
}

func (g *Generator) emitRuntimeActionSolverSetMappedInput(w *goWriter, rsp *runtimeActionSolverPlan, assign runtimeActionSolverAssign, rhs string) bool {
	if assign.mapped == nil {
		return false
	}
	g.pushExprOverrides(g.runtimeActionSolverBaseExprOverrides(rsp))
	defer g.popExprOverrides()
	if call, ok, err := g.goStorageSet(assign.mapped, rhs); ok || err != nil {
		if err != nil {
			w.line("return false")
			return true
		}
		w.line(call)
		return true
	}
	lhs, err := g.emitExpr(assign.mapped)
	if err != nil || lhs == "" {
		w.line("return false")
		return true
	}
	w.linef("%s = %s", lhs, rhs)
	return true
}

func runtimeActionSolverDefinedInputTerms(def goivy.Expr) (goivy.Expr, goivy.Expr, bool) {
	switch n := def.(type) {
	case *goivy.Eq:
		if n == nil {
			return nil, nil, false
		}
		return n.T1, n.T2, true
	case *goivy.LogicIff:
		if n == nil {
			return nil, nil, false
		}
		return n.T1, n.T2, true
	case *goivy.LogicDefinition:
		if n == nil {
			return nil, nil, false
		}
		return n.Lhs, n.Rhs, true
	default:
		return nil, nil, false
	}
}

func (g *Generator) emitRuntimeActionSolverGenerateReturn(w *goWriter) {
	w.open("if gen.__ivy_generate_with_solver() {")
	w.line("return true")
	w.close("")
	w.line("return false")
}

func (g *Generator) emitRuntimeActionSolverChooseMethod(w *goWriter, typeName string, rsp *runtimeActionSolverPlan) {
	if rsp == nil {
		return
	}
	w.open(fmt.Sprintf("func (gen *%s) choose(rng int, name string) int {", typeName))
	w.line("_ = rng")
	w.open("switch __ivy_solver_choice_symbol(name) {")
	seen := map[string]bool{}
	for _, assign := range rsp.assigns {
		if assign.input == nil || assign.target == "" || seen[assign.input.Name] {
			continue
		}
		if _, isFn := assign.sort.(*goivy.LogicFunctionSort); isFn {
			continue
		}
		if !g.runtimeActionSolverCanChooseSort(assign.sort) {
			continue
		}
		seen[assign.input.Name] = true
		w.linef("case %q:", assign.input.Name)
		w.indent++
		value, ok := g.runtimeActionSolverAssignTargetExpr(rsp, assign)
		if !ok {
			w.line("return 0")
		} else {
			g.emitRuntimeActionSolverChooseReturn(w, value, assign.sort)
		}
		w.indent--
	}
	w.close("")
	w.line("return 0")
	w.close("")
	w.blank()
}

func (g *Generator) runtimeActionSolverCanChooseSort(s goivy.Sort) bool {
	if s == nil {
		return false
	}
	if len(g.destructorStructFieldInfos(sortName(s))) > 0 {
		return false
	}
	if g.isVariantSuperName(sortName(s)) {
		return false
	}
	switch s.(type) {
	case *goivy.BooleanSort, *goivy.LogicEnumeratedSort, *goivy.RangeSort:
		return true
	case *goivy.UninterpretedSort:
		if g.hasStringValuedInterp(s) || g.hasIntOrNatInterp(s) {
			return true
		}
		if _, ok := g.rangeSortFor(s); ok {
			return true
		}
		return g.sortCard(s) > 0
	default:
		if _, ok := g.rangeSortFor(s); ok {
			return true
		}
		return g.sortCard(s) > 0
	}
}

func (g *Generator) emitRuntimeActionSolverChooseReturn(w *goWriter, value string, sort goivy.Sort) {
	if isBooleanSort(sort) {
		w.open(fmt.Sprintf("if %s {", value))
		w.line("return 1")
		w.close("")
		w.line("return 0")
		return
	}
	if g.hasStringValuedInterp(sort) {
		tmp := g.nextTemp("__ivy_choice")
		w.open(fmt.Sprintf("if %s, err := strconv.Atoi(%s); err == nil {", tmp, value))
		w.linef("return %s", tmp)
		w.close("")
		w.line("return 0")
		return
	}
	w.linef("return int(%s)", value)
}

func (g *Generator) emitRuntimeActionSolverAssignModelToTarget(w *goWriter, rsp *runtimeActionSolverPlan, assign runtimeActionSolverAssign, termExpr string) {
	if assign.noReadback {
		return
	}
	if assign.defined {
		return
	}
	if fs, ok := assign.sort.(*goivy.LogicFunctionSort); ok && len(fs.Domain()) > 0 {
		g.emitRuntimeActionSolverAssignFunctionModelToTarget(w, rsp, assign, fs)
		return
	}
	if assign.mapped == nil {
		g.emitRuntimeActionSolverAssignModelTermValue(w, "gen."+assign.target, assign.sort, termExpr)
		return
	}
	w.open("{")
	tmp := g.nextTemp("__ivy_solver_model")
	w.linef("var %s %s", tmp, g.goType(assign.sort))
	g.emitRuntimeActionSolverAssignModelTermValue(w, tmp, assign.sort, termExpr)
	g.emitRuntimeActionSolverSetMappedInput(w, rsp, assign, tmp)
	w.close("")
}

func (g *Generator) emitRuntimeActionSolverRandomInputEquality(w *goWriter, rsp *runtimeActionSolverPlan, assign runtimeActionSolverAssign, idx int) {
	if assign.input == nil || assign.target == "" || assign.sort == nil {
		w.line("return false")
		return
	}
	if assign.noReadback {
		return
	}
	if fs, ok := assign.sort.(*goivy.LogicFunctionSort); ok && len(fs.Domain()) > 0 {
		g.emitRuntimeActionSolverRandomFunctionInputEqualities(w, rsp, assign, idx, fs)
		return
	}
	randExpr := ""
	if it, ok := g.goInterpType(assign.sort); ok && it.Kind == goInterpIntBV {
		randExpr = fmt.Sprintf("ivyIntBVRandomSoft(%q, %d, %d, %d)", sortName(assign.sort), it.Lo, it.Hi, it.Bits)
	} else if g.Config.Target == "gen" && g.hasStrBVInterp(assign.sort) {
		if it, ok := g.goInterpType(assign.sort); ok {
			randExpr = fmt.Sprintf("ivyStrBVRandomSoft(%q, %d)", sortName(assign.sort), it.Bits)
		}
	}
	if randExpr == "" {
		var err error
		randExpr, err = g.goActionParamRandomValueExpr(assign.sort, assign.input.Name, int64(idx))
		if err != nil {
			g.unsupported(w, "unsupported runtime solver random input %s: %s", assign.input.Name, strings.TrimPrefix(err.Error(), "ivy2golang: "))
			w.line("return false")
			return
		}
	}
	inputExpr, ok := g.goIvyExprExpr(assign.input)
	if !ok {
		w.line("return false")
		return
	}
	valueExpr := "gen." + assign.target
	if assign.mapped != nil {
		tmp := g.nextTemp("__ivy_solver_rand")
		w.linef("%s := %s", tmp, randExpr)
		g.emitRuntimeActionSolverSetMappedInput(w, rsp, assign, tmp)
		valueExpr = tmp
	} else {
		w.linef("gen.%s = %s", assign.target, randExpr)
	}
	if !g.emitRuntimeActionSolverAppendValueEquality(w, "soft", inputExpr, valueExpr, assign.sort) {
		w.line("return false")
		return
	}
}

func (g *Generator) emitRuntimeActionSolverRandomFunctionInputEqualities(w *goWriter, rsp *runtimeActionSolverPlan, assign runtimeActionSolverAssign, idx int, fs *goivy.LogicFunctionSort) {
	if fs == nil || !g.runtimeSolverCanLoopDomain(fs.Domain()) {
		w.line("return false")
		return
	}
	randExpr, err := g.goActionParamRandomValueExpr(assign.sort, assign.input.Name, int64(idx))
	if err != nil {
		g.unsupported(w, "unsupported runtime solver random function input %s: %s", assign.input.Name, strings.TrimPrefix(err.Error(), "ivy2golang: "))
		w.line("return false")
		return
	}
	valueBase := "gen." + assign.target
	if assign.mapped != nil {
		valueBase = g.nextTemp("__ivy_solver_rand_fn")
		w.linef("%s := %s", valueBase, randExpr)
		g.emitRuntimeActionSolverSetMappedInput(w, rsp, assign, valueBase)
	} else {
		w.linef("gen.%s = %s", assign.target, randExpr)
	}
	inputExpr, ok := g.goIvyExprExpr(assign.input)
	if !ok {
		w.line("return false")
		return
	}
	g.emitRuntimeActionSolverFunctionInputCells(w, fs, nil, nil, func(goArgs []string, ivyArgs []string) {
		lhs := fmt.Sprintf("goivy.MustApply(%s, %s)", inputExpr, strings.Join(ivyArgs, ", "))
		valueExpr := g.runtimeActionSolverFunctionStorageAccess(valueBase, fs, goArgs)
		if !g.emitRuntimeActionSolverAppendValueEquality(w, "soft", lhs, valueExpr, fs.Range()) {
			w.line("return false")
			return
		}
	})
}

func (g *Generator) emitRuntimeActionSolverAssignFunctionModelToTarget(w *goWriter, rsp *runtimeActionSolverPlan, assign runtimeActionSolverAssign, fs *goivy.LogicFunctionSort) {
	if fs == nil || !g.runtimeSolverCanLoopDomain(fs.Domain()) {
		w.line("return false")
		return
	}
	inputExpr, ok := g.goIvyExprExpr(assign.input)
	if !ok {
		w.line("return false")
		return
	}
	targetBase := "gen." + assign.target
	if assign.mapped != nil {
		targetBase = g.nextTemp("__ivy_solver_model_fn")
		w.linef("%s := %s", targetBase, g.goFunctionStorageInit(fs.Domain(), fs.Range()))
	} else {
		w.linef("gen.%s = %s", assign.target, g.goFunctionStorageInit(fs.Domain(), fs.Range()))
	}
	g.emitRuntimeActionSolverFunctionInputCells(w, fs, nil, nil, func(goArgs []string, ivyArgs []string) {
		term := fmt.Sprintf("goivy.MustApply(%s, %s)", inputExpr, strings.Join(ivyArgs, ", "))
		if g.goFunctionStorageFor(fs.Domain(), fs.Range()).Large {
			w.open("{")
			tmp := g.nextTemp("__ivy_solver_cell")
			w.linef("var %s %s", tmp, g.goScalarType(fs.Range()))
			g.emitRuntimeActionSolverAssignModelTermValue(w, tmp, fs.Range(), term)
			g.emitRuntimeActionSolverSetFunctionCell(w, targetBase, fs, goArgs, tmp)
			w.close("")
			return
		}
		target := g.runtimeActionSolverFunctionStorageAccess(targetBase, fs, goArgs)
		g.emitRuntimeActionSolverAssignModelTermValue(w, target, fs.Range(), term)
	})
	if assign.mapped != nil {
		g.emitRuntimeActionSolverSetMappedInput(w, rsp, assign, targetBase)
	}
}

func (g *Generator) emitRuntimeActionSolverFunctionInputCells(w *goWriter, fs *goivy.LogicFunctionSort, goArgs []string, ivyArgs []string, body func([]string, []string)) {
	depth := len(goArgs)
	domain := fs.Domain()
	if depth < len(domain) {
		sort := domain[depth]
		name := fmt.Sprintf("__ivy_solver_arg%d", depth)
		header, ok := g.runtimeSolverLoopHeaderForSort(sort, name)
		if !ok {
			w.line("return false")
			return
		}
		w.open(header)
		if g.hasStrBVInterp(sort) {
			w.linef("%s := strconv.Itoa(%s_idx)", name, name)
		}
		argExpr, ok := g.runtimeActionSolverValueExpr(name, sort)
		if !ok {
			w.line("return false")
			w.close("")
			return
		}
		g.emitRuntimeActionSolverFunctionInputCells(
			w,
			fs,
			append(append([]string(nil), goArgs...), name),
			append(append([]string(nil), ivyArgs...), argExpr),
			body,
		)
		w.close("")
		return
	}
	body(goArgs, ivyArgs)
}

func (g *Generator) emitRuntimeActionSolverSetFunctionCell(w *goWriter, base string, fs *goivy.LogicFunctionSort, goArgs []string, value string) {
	key := g.goMapKeyValue(fs.Domain(), goArgs)
	value = g.emitClonedValueExpr(w, value, fs.Range())
	if g.goFunctionStorageFor(fs.Domain(), fs.Range()).Large {
		w.linef("%s.Set(%s, %s)", base, key, value)
		return
	}
	w.linef("%s[%s] = %s", base, key, value)
}

func (g *Generator) runtimeActionSolverFunctionStorageAccess(base string, fs *goivy.LogicFunctionSort, goArgs []string) string {
	key := g.goMapKeyValue(fs.Domain(), goArgs)
	if g.goFunctionStorageFor(fs.Domain(), fs.Range()).Large {
		return fmt.Sprintf("%s.Get(%s)", base, key)
	}
	return fmt.Sprintf("%s[%s]", base, key)
}

func (g *Generator) emitRuntimeActionSolverExtraFields(w *goWriter, rsp *runtimeActionSolverPlan) {
	if rsp == nil {
		return
	}
	w.line("__ivy_solver *goivy.Solver")
	w.line("__ivy_solver_base *goivy.SMTLIBBaseSolver")
	w.line("__ivy_solver_pre *goivy.Clauses")
	for _, member := range g.runtimeActionSolverGeneratedMembers(rsp) {
		w.linef("%s %s", member.target, g.goType(member.sort))
	}
}

func (g *Generator) runtimeActionSolverGeneratedMembers(rsp *runtimeActionSolverPlan) []runtimeActionSolverMember {
	if rsp == nil {
		return nil
	}
	seen := map[string]bool{}
	if rsp.plan != nil && rsp.plan.act != nil {
		for _, p := range rsp.plan.act.GetFormalParams() {
			if p == nil {
				continue
			}
			seen[goName(p.Name)] = true
			for _, alias := range formalExprOverrideNames(p.Name) {
				seen[goName(alias)] = true
			}
		}
	}
	var out []runtimeActionSolverMember
	add := func(name string, sort goivy.Sort) {
		if name == "" || sort == nil {
			return
		}
		target := goName(name)
		if seen[target] {
			return
		}
		seen[target] = true
		out = append(out, runtimeActionSolverMember{name: name, target: target, sort: sort})
	}
	for _, assign := range rsp.assigns {
		if assign.declare {
			add(assign.target, assign.sort)
		}
	}
	for _, assign := range rsp.assigns {
		if assign.input == nil || assign.mapped == nil || assign.mapped.Equal(assign.input) {
			continue
		}
		if strings.HasPrefix(assign.input.Name, "__ts") || assign.input.Name == "*>" || g.isDerivedStorageName(assign.input.Name) {
			continue
		}
		if actionGenStrictDefIdx(assign.input, rsp.plan) {
			continue
		}
		rootConst, ok := exprAsConst(exprRoot(assign.mapped))
		if !ok || rootConst == nil {
			continue
		}
		if _, isState := g.isStateSymbolName(rootConst.Name); isState {
			continue
		}
		add(rootConst.Name, rootConst.CSort)
	}
	return out
}

func (g *Generator) emitRuntimeActionSolverStateEquality(w *goWriter, sym stateSymbol) {
	if fs, ok := sym.Sort.(*goivy.LogicFunctionSort); ok {
		fnExpr := fmt.Sprintf("goivy.NewConst(%q, %s)", sym.Name, g.goIvySortExpr(sym.Sort))
		valueBase := "ivy." + goName(sym.Name)
		if g.runtimeSolverSupportsSparseThunkStateSymbol(sym, fs) {
			g.emitRuntimeActionSolverSparseThunkValue(w, "clauses.Fmlas", fnExpr, valueBase, fs)
			return
		}
		g.emitRuntimeActionSolverFunctionValueEqualities(w, "clauses.Fmlas", fnExpr, valueBase, fs, nil, nil)
		return
	}
	lhs := fmt.Sprintf("goivy.NewConst(%q, %s)", sym.Name, g.goIvySortExpr(sym.Sort))
	if !g.emitRuntimeActionSolverAppendValueEquality(w, "clauses.Fmlas", lhs, "ivy."+goName(sym.Name), sym.Sort) {
		w.line("return false")
		return
	}
}

func (g *Generator) emitRuntimeActionSolverSparseThunkValue(w *goWriter, dst string, fnExpr string, valueBase string, fs *goivy.LogicFunctionSort) {
	g.emitRuntimeActionSolverSparseThunkValueWithReturn(w, dst, fnExpr, valueBase, fs, "return false")
}

func (g *Generator) emitRuntimeActionSolverSparseThunkValueWithReturn(w *goWriter, dst string, fnExpr string, valueBase string, fs *goivy.LogicFunctionSort, failStmt string) {
	domain := fs.Domain()
	if len(domain) == 0 {
		w.line(failStmt)
		return
	}
	w.open("{")
	w.line("__ivy_sparse_vars := []*goivy.LogicVariable{")
	w.indent++
	for i, d := range domain {
		w.linef("&goivy.LogicVariable{Name: %q, VSort: %s},", fmt.Sprintf("__ivy_sparse_%d", i), g.goIvySortExpr(d))
	}
	w.indent--
	w.line("}")
	w.line("__ivy_sparse_args := make([]goivy.Expr, len(__ivy_sparse_vars))")
	w.open("for __ivy_sparse_idx, __ivy_sparse_var := range __ivy_sparse_vars {")
	w.line("__ivy_sparse_args[__ivy_sparse_idx] = __ivy_sparse_var")
	w.close("")
	w.linef("__ivy_sparse_app := goivy.MustApply(%s, __ivy_sparse_args...)", fnExpr)
	w.line("__ivy_sparse_terms := []goivy.Expr{}")
	w.line("__ivy_sparse_disj := []goivy.Expr{}")
	w.line("__ivy_sparse_value_index := 0")
	w.open(fmt.Sprintf("for __ivy_sparse_key, __ivy_sparse_val := range %s.overrides {", valueBase))
	w.line("__ivy_sparse_value_index++")
	if isBooleanSort(fs.Range()) {
		w.open(fmt.Sprintf("if %s.base != nil && %s.base(__ivy_sparse_key) == __ivy_sparse_val {", valueBase, valueBase))
		w.line("continue")
		w.close("")
	}
	w.line("__ivy_sparse_eqs := []goivy.Expr{}")
	for i, d := range domain {
		keyExpr := "__ivy_sparse_key"
		if len(domain) != 1 {
			keyExpr = fmt.Sprintf("__ivy_sparse_key.A%d", i)
		}
		if !g.emitRuntimeActionSolverAppendStorageValueEquality(w, "__ivy_sparse_eqs", fmt.Sprintf("__ivy_sparse_vars[%d]", i), keyExpr, d) {
			w.line("_ = __ivy_sparse_key")
			w.line(failStmt)
			continue
		}
	}
	w.line("__ivy_sparse_cond := goivy.Expr(&goivy.LogicAnd{Terms: __ivy_sparse_eqs})")
	w.line("__ivy_sparse_disj = append(__ivy_sparse_disj, __ivy_sparse_cond)")
	valueExpr, ok := g.emitRuntimeActionSolverValueTerm(w, "clauses.Fmlas", "__ivy_sparse_val", fs.Range(), `"__ivy_sparse_val_" + strconv.Itoa(__ivy_sparse_value_index)`)
	if !ok {
		w.line(failStmt)
		w.close("")
		w.close("")
		return
	}
	w.linef("__ivy_sparse_terms = append(__ivy_sparse_terms, &goivy.LogicImplies{T1: __ivy_sparse_cond, T2: &goivy.Eq{T1: __ivy_sparse_app, T2: %s}})", valueExpr)
	w.close("")
	w.open(fmt.Sprintf("if %s.solverBase != nil {", valueBase))
	w.linef("__ivy_sparse_base_terms := %s.solverBase(__ivy_sparse_args, __ivy_sparse_app)", valueBase)
	w.open("if len(__ivy_sparse_base_terms) > 0 {")
	w.line("__ivy_sparse_quant_terms := []goivy.Expr{}")
	w.open("for _, __ivy_sparse_base_term := range __ivy_sparse_base_terms {")
	w.open("if len(goivy.FreeVariablesList(__ivy_sparse_base_term)) == 0 {")
	w.linef("%s = append(%s, __ivy_sparse_base_term)", dst, dst)
	w.nextBranch(" else {")
	w.line("__ivy_sparse_quant_terms = append(__ivy_sparse_quant_terms, __ivy_sparse_base_term)")
	w.close("")
	w.close("")
	w.open("if len(__ivy_sparse_quant_terms) > 0 {")
	w.line("__ivy_sparse_base := goivy.Expr(&goivy.LogicAnd{Terms: __ivy_sparse_quant_terms})")
	w.open("if len(__ivy_sparse_quant_terms) == 1 {")
	w.line("__ivy_sparse_base = __ivy_sparse_quant_terms[0]")
	w.close("")
	w.open("if len(__ivy_sparse_disj) == 0 {")
	w.line("__ivy_sparse_terms = append(__ivy_sparse_terms, __ivy_sparse_base)")
	w.nextBranch(" else {")
	w.line("__ivy_sparse_terms = append(__ivy_sparse_terms, &goivy.LogicOr{Terms: []goivy.Expr{&goivy.LogicOr{Terms: __ivy_sparse_disj}, __ivy_sparse_base}})")
	w.close("")
	w.close("")
	w.close("")
	w.close("")
	w.open("if len(__ivy_sparse_terms) > 0 {")
	w.line("__ivy_sparse_body := goivy.Expr(&goivy.LogicAnd{Terms: __ivy_sparse_terms})")
	w.open("if len(__ivy_sparse_terms) == 1 {")
	w.line("__ivy_sparse_body = __ivy_sparse_terms[0]")
	w.close("")
	w.linef("%s = append(%s, &goivy.RawForAll{Variables: __ivy_sparse_vars, Body: __ivy_sparse_body})", dst, dst)
	w.close("")
	w.close("")
}

func (g *Generator) emitRuntimeActionSolverFunctionValueEqualities(w *goWriter, dst string, fnExpr string, valueBase string, fs *goivy.LogicFunctionSort, goArgs []string, ivyArgs []string) {
	depth := len(goArgs)
	domain := fs.Domain()
	if depth < len(domain) {
		sort := domain[depth]
		name := fmt.Sprintf("__ivy_solver_state_arg%d", depth)
		header, ok := g.runtimeSolverLoopHeaderForSort(sort, name)
		if !ok {
			w.line("return false")
			return
		}
		w.open(header)
		if g.hasStrBVInterp(sort) {
			w.linef("%s := strconv.Itoa(%s_idx)", name, name)
		}
		argExpr, ok := g.runtimeActionSolverValueExpr(name, sort)
		if !ok {
			w.line("return false")
			w.close("")
			return
		}
		g.emitRuntimeActionSolverFunctionValueEqualities(
			w,
			dst,
			fnExpr,
			valueBase,
			fs,
			append(append([]string(nil), goArgs...), name),
			append(append([]string(nil), ivyArgs...), argExpr),
		)
		w.close("")
		return
	}
	lhs := fmt.Sprintf("goivy.MustApply(%s, %s)", fnExpr, strings.Join(ivyArgs, ", "))
	rhsValue := g.runtimeActionSolverFunctionStorageAccess(valueBase, fs, goArgs)
	if !g.emitRuntimeActionSolverAppendValueEquality(w, dst, lhs, rhsValue, fs.Range()) {
		w.line("return false")
		return
	}
}

func (g *Generator) runtimeActionSolverValueExpr(value string, s goivy.Sort) (string, bool) {
	switch st := s.(type) {
	case *goivy.BooleanSort:
		_ = st
		return fmt.Sprintf("__ivy_solver_bool_expr(%s)", value), true
	case *goivy.LogicEnumeratedSort:
		if isNumericEnum(st) {
			return fmt.Sprintf("goivy.NewConst(strconv.Itoa(int(%s)), %s)", value, g.goIvySortExpr(s)), true
		}
		sortExpr := g.goIvySortExpr(s)
		var b strings.Builder
		b.WriteString("func() goivy.Expr { switch " + value + " {")
		for _, val := range st.Extension {
			b.WriteString("case " + goName(val) + ": return goivy.NewConst(" + strconv.Quote(val) + ", " + sortExpr + ");")
		}
		b.WriteString("}; return goivy.NewConst(strconv.Itoa(int(" + value + ")), " + sortExpr + ") }()")
		return b.String(), true
	default:
		if len(g.destructorStructFieldInfos(sortName(s))) > 0 {
			return "", false
		}
		if g.isVariantSuperName(sortName(s)) {
			return "", false
		}
		if it, ok := g.goInterpType(s); ok && it.Kind == goInterpIntBV {
			return fmt.Sprintf("goivy.NewConst(strconv.Itoa(ivyIntBVXToBV(%q, %d, %d, %d, int(%s))), %s)", sortName(s), it.Lo, it.Hi, it.Bits, value, g.goIvySortExpr(s)), true
		}
		if g.Config.Target == "gen" && g.hasStrBVInterp(s) {
			if it, ok := g.goInterpType(s); ok {
				return fmt.Sprintf("goivy.NewConst(strconv.Itoa(ivyStrBVXToBV(%q, %d, %s)), %s)", sortName(s), it.Bits, value, g.goIvySortExpr(s)), true
			}
		}
		if g.hasStringValuedInterp(s) {
			return fmt.Sprintf("goivy.NewConst(strconv.Quote(%s), %s)", value, g.goIvySortExpr(s)), true
		}
		if g.runtimeSolverSupportsScalarSort(s) {
			return fmt.Sprintf("goivy.NewConst(strconv.Itoa(int(%s)), %s)", value, g.goIvySortExpr(s)), true
		}
		return "", false
	}
}

func (g *Generator) emitRuntimeActionSolverAppendValueEquality(w *goWriter, dst string, lhs string, value string, s goivy.Sort) bool {
	if g.isVariantSuperName(sortName(s)) {
		return g.emitRuntimeActionSolverAppendVariantEquality(w, dst, lhs, value, s)
	}
	if fields := g.destructorStructFieldInfos(sortName(s)); len(fields) > 0 {
		g.emitRuntimeActionSolverAppendRecordValueEquality(w, dst, lhs, value, fields)
		return true
	}
	expr, ok := g.emitRuntimeActionSolverValueConstraintExpr(w, dst, lhs, value, s)
	if !ok {
		return false
	}
	w.linef("%s = append(%s, %s)", dst, dst, expr)
	return true
}

func (g *Generator) emitRuntimeActionSolverAppendStorageValueEquality(w *goWriter, dst string, lhs string, value string, s goivy.Sort) bool {
	expr, ok := g.emitRuntimeActionSolverStorageValueConstraintExpr(lhs, value, s)
	if !ok {
		return false
	}
	w.linef("%s = append(%s, %s)", dst, dst, expr)
	return true
}

func (g *Generator) emitRuntimeActionSolverStorageValueConstraintExpr(lhs string, value string, s goivy.Sort) (string, bool) {
	rhs, ok := g.runtimeActionSolverStorageValueExpr(value, s)
	if !ok {
		return "", false
	}
	return fmt.Sprintf("&goivy.Eq{T1: %s, T2: %s}", lhs, rhs), true
}

func (g *Generator) emitRuntimeActionSolverValueConstraintExpr(w *goWriter, dst string, lhs string, value string, s goivy.Sort) (string, bool) {
	if fields := g.destructorStructFieldInfos(sortName(s)); len(fields) > 0 {
		return g.emitRuntimeActionSolverRecordConstraintExpr(w, dst, lhs, value, s, fields), true
	}
	rhs, ok := g.runtimeActionSolverValueExpr(value, s)
	if !ok {
		return "", false
	}
	return fmt.Sprintf("&goivy.Eq{T1: %s, T2: %s}", lhs, rhs), true
}

func (g *Generator) runtimeActionSolverStorageValueExpr(value string, s goivy.Sort) (string, bool) {
	if expr, ok := g.runtimeActionSolverValueExpr(value, s); ok {
		return expr, true
	}
	if _, ok := s.(*goivy.UninterpretedSort); ok &&
		!g.hasStringValuedInterp(s) &&
		len(g.destructorStructFieldInfos(sortName(s))) == 0 &&
		!g.isVariantSuperName(sortName(s)) {
		return fmt.Sprintf("goivy.NewConst(strconv.Itoa(int(%s)), %s)", value, g.goIvySortExpr(s)), true
	}
	return "", false
}

func (g *Generator) emitRuntimeActionSolverValueTerm(w *goWriter, dst string, value string, s goivy.Sort, nameExpr string) (string, bool) {
	if expr, ok := g.runtimeActionSolverValueExpr(value, s); ok {
		return expr, true
	}
	if !g.runtimeSolverSupportsScalarSort(s) {
		return "", false
	}
	tmp := g.nextTemp("__ivy_solver_value")
	name := strconv.Quote(tmp)
	if nameExpr != "" {
		name = nameExpr
	}
	w.linef("%s := goivy.NewConst(%s, %s)", tmp, name, g.goIvySortExpr(s))
	if !g.emitRuntimeActionSolverConstrainValueTerm(w, dst, tmp, value, s) {
		return "", false
	}
	return tmp, true
}

func (g *Generator) emitRuntimeActionSolverConstrainValueTerm(w *goWriter, dst string, term string, value string, s goivy.Sort) bool {
	if g.isVariantSuperName(sortName(s)) {
		return g.emitRuntimeActionSolverAppendVariantEquality(w, dst, term, value, s)
	}
	if fields := g.destructorStructFieldInfos(sortName(s)); len(fields) > 0 {
		g.emitRuntimeActionSolverAppendRecordValueEquality(w, dst, term, value, fields)
		return true
	}
	rhs, ok := g.runtimeActionSolverValueExpr(value, s)
	if !ok {
		return false
	}
	w.linef("%s = append(%s, &goivy.Eq{T1: %s, T2: %s})", dst, dst, term, rhs)
	return true
}

func (g *Generator) emitRuntimeActionSolverAppendRecordValueEquality(w *goWriter, dst string, lhs string, value string, fields []goDestructorField) {
	for _, field := range fields {
		g.emitRuntimeActionSolverAppendRecordFieldEquality(w, dst, lhs, value, field, nil, nil)
	}
}

func (g *Generator) emitRuntimeActionSolverRecordConstraintExpr(w *goWriter, dst string, lhs string, value string, s goivy.Sort, fields []goDestructorField) string {
	tmp := g.nextTemp("__ivy_solver_record")
	w.linef("%s := goivy.NewConst(%q, %s)", tmp, tmp, g.goIvySortExpr(s))
	for _, field := range fields {
		g.emitRuntimeActionSolverAppendRecordFieldEquality(w, dst, tmp, value, field, nil, nil)
	}
	return fmt.Sprintf("&goivy.Eq{T1: %s, T2: %s}", lhs, tmp)
}

func (g *Generator) emitRuntimeActionSolverAppendVariantEquality(w *goWriter, dst string, lhs string, value string, s goivy.Sort) bool {
	if g == nil || g.Mod == nil {
		return false
	}
	variants := g.Mod.Variants[sortName(s)]
	if len(variants) == 0 {
		return false
	}
	matched := g.nextTemp("__ivy_solver_variant_matched")
	w.open("{")
	w.linef("%s := false", matched)
	for i, sub := range variants {
		w.open(fmt.Sprintf("if %s.valid && %s.tag == %d {", value, value, i))
		w.linef("%s = true", matched)
		varName := g.nextTemp("__ivy_solver_variant")
		w.linef("%s := &goivy.LogicVariable{Name: %q, VSort: %s}", varName, varName, g.goIvySortExpr(sub))
		relSort, err := goivy.NewFunctionSort(s, sub, goivy.Boolean)
		if err != nil {
			w.line("return false")
			w.close("")
			continue
		}
		rel := fmt.Sprintf("goivy.MustApply(goivy.NewConst(%q, %s), %s, %s)", "*>", g.goIvySortExpr(relSort), lhs, varName)
		payload := fmt.Sprintf("%s.value.(%s)", value, g.goScalarType(sub))
		terms := g.nextTemp("__ivy_solver_variant_terms")
		w.linef("%s := []goivy.Expr{%s}", terms, rel)
		if !g.emitRuntimeActionSolverAppendValueEquality(w, terms, varName, payload, sub) {
			w.line("return false")
			w.close("")
			continue
		}
		body := fmt.Sprintf("&goivy.LogicAnd{Terms: %s}", terms)
		w.linef("%s = append(%s, &goivy.LogicExists{Variables: []*goivy.LogicVariable{%s}, Body: %s})", dst, dst, varName, body)
		w.close("")
	}
	w.open(fmt.Sprintf("if !%s {", matched))
	w.linef("%s = append(%s, goivy.False)", dst, dst)
	w.close("")
	w.close("")
	return true
}

func (g *Generator) emitRuntimeActionSolverAppendRecordFieldEquality(w *goWriter, dst string, receiverTerm string, receiverValue string, field goDestructorField, goArgs []string, ivyArgs []string) {
	if field.Sort == nil || field.Const == nil {
		w.line("return false")
		return
	}
	domain := field.Sort.Domain()
	if len(domain) == 0 {
		w.line("return false")
		return
	}
	extra := domain[1:]
	depth := len(goArgs)
	if depth < len(extra) {
		sort := extra[depth]
		name := fmt.Sprintf("__ivy_solver_record_arg%d", depth)
		header, ok := g.runtimeSolverLoopHeaderForSort(sort, name)
		if !ok {
			w.line("return false")
			return
		}
		w.open(header)
		if g.hasStrBVInterp(sort) {
			w.linef("%s := strconv.Itoa(%s_idx)", name, name)
		}
		argExpr, ok := g.runtimeActionSolverValueExpr(name, sort)
		if !ok {
			w.line("return false")
			w.close("")
			return
		}
		g.emitRuntimeActionSolverAppendRecordFieldEquality(
			w,
			dst,
			receiverTerm,
			receiverValue,
			field,
			append(append([]string(nil), goArgs...), name),
			append(append([]string(nil), ivyArgs...), argExpr),
		)
		w.close("")
		return
	}
	args := append([]string{receiverTerm}, ivyArgs...)
	lhs := fmt.Sprintf("goivy.MustApply(goivy.NewConst(%q, %s), %s)", field.Const.Name, g.goIvySortExpr(field.Sort), strings.Join(args, ", "))
	fieldValue := g.runtimeActionSolverRecordFieldValueExpr(receiverValue, field, goArgs)
	if !g.emitRuntimeActionSolverAppendValueEquality(w, dst, lhs, fieldValue, field.Sort.Range()) {
		w.line("return false")
		return
	}
}

func (g *Generator) runtimeActionSolverRecordFieldValueExpr(receiverValue string, field goDestructorField, goArgs []string) string {
	base := receiverValue + "." + field.FieldName
	if g.destructorFieldIsLarge(field) {
		key := g.goMapKeyValue(field.Sort.Domain()[1:], goArgs)
		return fmt.Sprintf("ivyThunkGet(%s, %s, %s)", base, key, g.goZeroValue(field.Sort.Range()))
	}
	if len(goArgs) > 0 {
		base += goIndexSuffix(goArgs)
	}
	return base
}

func (g *Generator) emitRuntimeActionSolverAssignModelTermValue(w *goWriter, target string, s goivy.Sort, termExpr string) {
	if g.isVariantSuperName(sortName(s)) {
		g.emitRuntimeActionSolverAssignVariantModelTermValue(w, target, s, termExpr)
		return
	}
	if fields := g.destructorStructFieldInfos(sortName(s)); len(fields) > 0 {
		g.emitRuntimeActionSolverAssignRecordModelTermValue(w, target, s, termExpr, fields)
		return
	}
	g.emitRuntimeActionSolverAssignModelValue(w, target, s, fmt.Sprintf("hm.EvalToConstant(%s)", termExpr))
}

func (g *Generator) emitRuntimeActionSolverAssignVariantModelTermValue(w *goWriter, target string, s goivy.Sort, termExpr string) {
	if g == nil || g.Mod == nil {
		w.line("return false")
		return
	}
	variants := g.Mod.Variants[sortName(s)]
	if len(variants) == 0 {
		w.line("return false")
		return
	}
	assigned := g.nextTemp("__ivy_solver_variant_assigned")
	w.open("{")
	w.linef("%s := false", assigned)
	for _, sub := range variants {
		relSort, err := goivy.NewFunctionSort(s, sub, goivy.Boolean)
		if err != nil {
			w.line("return false")
			continue
		}
		univ := g.nextTemp("__ivy_solver_variant_univ")
		w.open(fmt.Sprintf("for _, %s := range hm.SortUniverse(%s) {", univ, g.goIvySortExpr(sub)))
		w.linef("if %s {", assigned)
		w.indent++
		w.line("break")
		w.indent--
		w.line("}")
		rel := fmt.Sprintf("goivy.MustApply(goivy.NewConst(%q, %s), %s, %s)", "*>", g.goIvySortExpr(relSort), termExpr, univ)
		w.open(fmt.Sprintf("if goivy.IsTrue(hm.EvalToConstant(%s)) {", rel))
		tmp := g.nextTemp("__ivy_solver_variant_payload")
		w.linef("var %s %s", tmp, g.goScalarType(sub))
		g.emitRuntimeActionSolverAssignModelTermValue(w, tmp, sub, univ)
		w.linef("%s = %s", target, g.variantUpcastExpr(s, sub, tmp))
		w.linef("%s = true", assigned)
		w.line("break")
		w.close("")
		w.close("")
	}
	w.open(fmt.Sprintf("if !%s {", assigned))
	w.linef("%s = %s{}", target, g.goScalarType(s))
	w.close("")
	w.close("")
}

func (g *Generator) emitRuntimeActionSolverAssignRecordModelTermValue(w *goWriter, target string, s goivy.Sort, termExpr string, fields []goDestructorField) {
	_ = s
	w.open("{")
	for _, field := range fields {
		g.emitRuntimeActionSolverAssignRecordFieldModelTermValue(w, target, termExpr, field, nil, nil)
	}
	w.close("")
}

func (g *Generator) emitRuntimeActionSolverAssignRecordFieldModelTermValue(w *goWriter, target string, receiverTerm string, field goDestructorField, goArgs []string, ivyArgs []string) {
	if field.Sort == nil || field.Const == nil {
		w.line("return false")
		return
	}
	domain := field.Sort.Domain()
	if len(domain) == 0 {
		w.line("return false")
		return
	}
	extra := domain[1:]
	depth := len(goArgs)
	if depth < len(extra) {
		sort := extra[depth]
		name := fmt.Sprintf("__ivy_solver_record_eval_arg%d", depth)
		header, ok := g.runtimeSolverLoopHeaderForSort(sort, name)
		if !ok {
			w.line("return false")
			return
		}
		w.open(header)
		if g.hasStrBVInterp(sort) {
			w.linef("%s := strconv.Itoa(%s_idx)", name, name)
		}
		argExpr, ok := g.runtimeActionSolverValueExpr(name, sort)
		if !ok {
			w.line("return false")
			w.close("")
			return
		}
		g.emitRuntimeActionSolverAssignRecordFieldModelTermValue(
			w,
			target,
			receiverTerm,
			field,
			append(append([]string(nil), goArgs...), name),
			append(append([]string(nil), ivyArgs...), argExpr),
		)
		w.close("")
		return
	}
	args := append([]string{receiverTerm}, ivyArgs...)
	term := fmt.Sprintf("goivy.MustApply(goivy.NewConst(%q, %s), %s)", field.Const.Name, g.goIvySortExpr(field.Sort), strings.Join(args, ", "))
	if g.destructorFieldIsLarge(field) {
		w.open("{")
		tmp := g.nextTemp("__ivy_solver_record_field")
		w.linef("var %s %s", tmp, g.goScalarType(field.Sort.Range()))
		g.emitRuntimeActionSolverAssignModelTermValue(w, tmp, field.Sort.Range(), term)
		tmp = g.emitClonedValueExpr(w, tmp, field.Sort.Range())
		key := g.goMapKeyValue(field.Sort.Domain()[1:], goArgs)
		w.linef("ivyThunkSet(&%s.%s, %s, %s, %s)", target, field.FieldName, key, tmp, g.goZeroValue(field.Sort.Range()))
		w.close("")
		return
	}
	fieldTarget := target + "." + field.FieldName
	if len(goArgs) > 0 {
		fieldTarget += goIndexSuffix(goArgs)
	}
	g.emitRuntimeActionSolverAssignModelTermValue(w, fieldTarget, field.Sort.Range(), term)
}

func (g *Generator) emitRuntimeActionSolverAssignModelValue(w *goWriter, target string, s goivy.Sort, modelExpr string) {
	w.open("{")
	switch st := s.(type) {
	case *goivy.BooleanSort:
		_ = st
		w.linef("__ivy_value := %s", modelExpr)
		w.open("if goivy.IsTrue(__ivy_value) {")
		w.linef("%s = true", target)
		w.nextBranch(" else if goivy.IsFalse(__ivy_value) {")
		w.linef("%s = false", target)
		w.nextBranch(" else if __ivy_const, ok := __ivy_value.(*goivy.Const); ok && __ivy_const != nil {")
		w.open(`switch __ivy_const.Name {`)
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
		w.nextBranch(" else {")
		w.line("return false")
		w.close("")
	case *goivy.LogicEnumeratedSort:
		if isNumericEnum(st) {
			g.emitRuntimeActionSolverAssignNumericModelValue(w, target, s, modelExpr)
			break
		}
		if g.Config.Target == "gen" {
			if len(st.Extension) == 0 {
				w.line("return false")
			} else {
				w.linef("%s = %s", target, goName(st.Extension[0]))
			}
			break
		}
		w.linef("__ivy_value, ok := %s.(*goivy.Const)", modelExpr)
		w.open("if !ok || __ivy_value == nil {")
		w.line("return false")
		w.close("")
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
		if g.hasStringValuedInterp(s) {
			g.emitRuntimeActionSolverAssignStringModelValue(w, target, s, modelExpr)
		} else {
			g.emitRuntimeActionSolverAssignNumericModelValue(w, target, s, modelExpr)
		}
	}
	w.close("")
}

func (g *Generator) emitRuntimeActionSolverAssignStringModelValue(w *goWriter, target string, s goivy.Sort, modelExpr string) {
	w.linef("__ivy_value, ok := %s.(*goivy.Const)", modelExpr)
	w.open("if !ok || __ivy_value == nil {")
	w.line("return false")
	w.close("")
	w.line("__ivy_name := __ivy_value.Name")
	w.open(`if len(__ivy_name) >= 2 && __ivy_name[0] == '"' && __ivy_name[len(__ivy_name)-1] == '"' {`)
	w.line("__ivy_unquoted, err := strconv.Unquote(__ivy_name)")
	w.open("if err != nil {")
	w.line("return false")
	w.close("")
	w.line("__ivy_name = __ivy_unquoted")
	w.close("")
	if g.hasStrBVInterp(s) {
		w.line("__ivy_num, err := strconv.Atoi(__ivy_name)")
		w.open("if err != nil {")
		w.open(`if len(__ivy_name) > 2 && __ivy_name[0] == '#' && (__ivy_name[1] == 'x' || __ivy_name[1] == 'X') {`)
		w.line("__ivy_parsed, err2 := strconv.ParseInt(__ivy_name[2:], 16, 64)")
		w.open("if err2 != nil {")
		w.line("return false")
		w.close("")
		w.line("__ivy_num = int(__ivy_parsed)")
		w.nextBranch(` else if len(__ivy_name) > 2 && __ivy_name[0] == '#' && (__ivy_name[1] == 'b' || __ivy_name[1] == 'B') {`)
		w.line("__ivy_parsed, err2 := strconv.ParseInt(__ivy_name[2:], 2, 64)")
		w.open("if err2 != nil {")
		w.line("return false")
		w.close("")
		w.line("__ivy_num = int(__ivy_parsed)")
		w.nextBranch(" else {")
		w.line("return false")
		w.close("")
		w.close("")
		if g.Config.Target == "gen" {
			if it, ok := g.goInterpType(s); ok {
				w.linef("%s = ivyStrBVBVToX(%q, %d, __ivy_num)", target, sortName(s), it.Bits)
				return
			}
		}
		w.linef("%s = strconv.Itoa(__ivy_num)", target)
		return
	}
	w.linef("%s = __ivy_name", target)
}

func (g *Generator) emitRuntimeActionSolverAssignNumericModelValue(w *goWriter, target string, s goivy.Sort, modelExpr string) {
	w.linef("__ivy_value, ok := %s.(*goivy.Const)", modelExpr)
	w.open("if !ok || __ivy_value == nil {")
	w.line("return false")
	w.close("")
	w.line("__ivy_num, err := strconv.Atoi(__ivy_value.Name)")
	w.open("if err != nil {")
	w.open(`if __ivy_sep := strings.IndexByte(__ivy_value.Name, ':'); __ivy_sep > 0 {`)
	w.line("__ivy_num, err = strconv.Atoi(__ivy_value.Name[:__ivy_sep])")
	w.close("")
	w.close("")
	w.open("if err != nil {")
	w.open(`if len(__ivy_value.Name) > 2 && __ivy_value.Name[0] == '#' && (__ivy_value.Name[1] == 'x' || __ivy_value.Name[1] == 'X') {`)
	w.line("__ivy_parsed, err2 := strconv.ParseInt(__ivy_value.Name[2:], 16, 64)")
	w.open("if err2 != nil {")
	w.line("return false")
	w.close("")
	w.line("__ivy_num = int(__ivy_parsed)")
	w.nextBranch(` else if len(__ivy_value.Name) > 2 && __ivy_value.Name[0] == '#' && (__ivy_value.Name[1] == 'b' || __ivy_value.Name[1] == 'B') {`)
	w.line("__ivy_parsed, err2 := strconv.ParseInt(__ivy_value.Name[2:], 2, 64)")
	w.open("if err2 != nil {")
	w.line("return false")
	w.close("")
	w.line("__ivy_num = int(__ivy_parsed)")
	w.nextBranch(" else {")
	w.line("__ivy_found := false")
	w.open(fmt.Sprintf("for __ivy_idx, __ivy_candidate := range hm.SortUniverse(%s) {", g.goIvySortExpr(s)))
	w.open("if __ivy_value.Equal(__ivy_candidate) {")
	w.line("__ivy_num = __ivy_idx")
	w.line("__ivy_found = true")
	w.line("break")
	w.close("")
	w.close("")
	w.open("if !__ivy_found {")
	w.line("return false")
	w.close("")
	w.close("")
	w.close("")
	if it, ok := g.goInterpType(s); ok && it.Kind == goInterpIntBV {
		w.linef("%s = ivyIntBVBVToX(%q, %d, %d, %d, __ivy_num)", target, sortName(s), it.Lo, it.Hi, it.Bits)
		return
	}
	w.linef("%s = __ivy_num", target)
}
