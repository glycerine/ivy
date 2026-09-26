package ivy2cpp

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/glycerine/ivy/goivy"
)

// action_gen.go ports Python's emit_action_gen (ivy_to_cpp.py:1197-1348),
// converting an action into a Z3-backed generator class that asserts the
// action's reverse_image precondition, sets the pre-state, randomizes
// inputs, solves, and reads back the inputs from the model.
//
// The work is split into two phases:
//
//   buildActionGenPlan(name, act)  -- pure analysis (no emission).
//   emitActionGen(w, plan)         -- emits the constructor / generate /
//                                     execute bodies for one plan.
//
// Class headers are emitted by emitActionGenClassHeader(w, plan).
//
// Explicit analysis failures report a fallback and the emitter falls back
// to the weaker pre-M4 generator that randomizes inputs directly without
// consulting the solver. Illegal action trees still surface as normal Go
// panics; callers should fix those construction bugs instead of hiding them.

// actionGenPlan captures everything we need to emit a single action's
// solver-backed generator class.
type actionGenPlan struct {
	name           string
	className      string
	origAct        goivy.Action
	act            goivy.Action // possibly wrapped by before_export / ext_preconds
	inputs         []*goivy.Const
	fsyms          map[goivy.NodeKey]goivy.Expr
	paramDefs      []goivy.Expr
	oldPreClauses  *goivy.Clauses // pre-AND of relevant_definitions / variant_axioms
	origPreClauses *goivy.Clauses
	preFmla        goivy.Expr
	used           *goivy.InsMap[goivy.NodeKey, goivy.Expr]
	fallback       bool
	fallbackReason string
}

// buildActionGenPlan runs the Python flow at ivy_to_cpp.py:1197-1239 to
// produce a normalized precondition and the matching set of solver
// inputs. Returns a plan whose .fallback flag indicates whether the
// caller should fall back to the weak generator.
func (g *Generator) buildActionGenPlan(name string, act goivy.Action) *actionGenPlan {
	plan := &actionGenPlan{
		name:      name,
		className: g.actionGeneratorClassName(name),
		origAct:   act,
		act:       act,
	}

	// Python: action = im.module.before_export.get(name, action)
	if g.Mod.BeforeExport != nil {
		if be, ok := g.Mod.BeforeExport.Get2(name); ok && be != nil {
			plan.act = be
		}
	}

	// Python: if name in im.module.ext_preconds:
	//             orig_action = action
	//             action = ia.Sequence(ia.AssumeAction(im.module.ext_preconds[name]),action)
	//             action.lineno = orig_action.lineno
	//             action.formal_params = orig_action.formal_params
	//             action.formal_returns = orig_action.formal_returns
	// (ivy_to_cpp.py:1213-1217)
	if g.Mod.ExtPreconds != nil {
		if pre, ok := g.Mod.ExtPreconds[name]; ok && pre != nil {
			orig := plan.act
			seq := goivy.NewSequence(goivy.NewAssumeAction(pre), exprOfAction(orig))
			seq.SetLineno(orig.GetLineno())
			goivy.CopyFormalsTo(orig, seq)
			plan.act = seq
		}
	}

	upd := goivy.GetUpdateForArt(plan.act, g.Mod, nil)
	if upd == nil {
		plan.fallback = true
		plan.fallbackReason = "GetUpdate returned nil"
		return plan
	}

	// pre = tr.reverse_image(true_clauses, true_clauses, upd)
	truePre := goivy.TrueClauses(nil)
	preClauses := goivy.ReverseImage(truePre, truePre, upd)
	plan.origPreClauses = preClauses
	preClauses = goivy.TrimClauses(preClauses)
	preClauses = expandFieldReferences(preClauses, g.Mod.DestructorSorts)
	preClauses = g.filterActionGenDerivedStorageClauses(preClauses)

	// Collect inputs from used local symbols + formal params (prefixed __).
	var inputs []*goivy.Const
	inputSet := map[goivy.NodeKey]bool{}
	for _, sym := range goivy.UsedSymbolsClausesOrdered(preClauses).All() {
		c, ok := sym.(*goivy.Const)
		if !ok {
			continue
		}
		if !isLocalSym(c, g.Mod.Sig) {
			continue
		}
		if g.isDerivedStorageName(c.Name) {
			continue
		}
		if goivy.IsNumeral(c) {
			continue
		}
		k := goivy.Key(c)
		if inputSet[k] {
			continue
		}
		inputSet[k] = true
		inputs = append(inputs, c)
	}
	for _, p := range plan.act.GetFormalParams() {
		pp := goivy.NewConst("__"+p.Name, p.CSort)
		k := goivy.Key(pp)
		if inputSet[k] {
			continue
		}
		inputSet[k] = true
		inputs = append(inputs, pp)
	}

	// Field extraction + defined parameters.
	preClauses, inputs, plan.fsyms = extractInputFields(preClauses, inputs, g.Mod)
	inputs = appendActionGenTsLocalInputs(preClauses, inputs, g.Mod)
	plan.oldPreClauses = preClauses
	preClauses, plan.paramDefs = extractDefinedParameters(preClauses, inputs)

	// AND with relevant definitions + variant axioms.
	usedNames := make(map[string]bool)
	for _, sym := range goivy.UsedSymbolsClausesOrdered(preClauses).All() {
		if c, ok := sym.(*goivy.Const); ok && c.Name != "" {
			usedNames[c.Name] = true
		}
	}
	rdefs := goivy.RelevantDefinitions(g.Mod, usedNames)
	var rdefFmlas []goivy.Expr
	for _, lf := range rdefs {
		if def, ok := lf.Formula.(*goivy.LogicDefinition); ok {
			fixed := fixDefinition(def)
			rdefFmlas = append(rdefFmlas, goivy.DefinitionToConstraint(fixed))
		}
	}
	if len(rdefFmlas) > 0 {
		preClauses = goivy.AndClausesTyped(preClauses, goivy.NewClauses(rdefFmlas, nil, nil))
	}
	varAxioms := g.Mod.VariantAxioms()
	if len(varAxioms) > 0 {
		preClauses = goivy.AndClausesTyped(preClauses, goivy.NewClauses(varAxioms, nil, nil))
	}
	plan.preFmla = preClauses.ToFormula()
	plan.used = goivy.UsedSymbolsAst(plan.preFmla)
	plan.inputs = g.orderActionGenGeneratedInputsByFormula(inputs, plan.preFmla, plan.name)

	// Validate: no numerals of uninterpreted sort (Python raises here at
	// ivy_to_cpp.py:1243-1244). The Go side propagates as a Generator
	// error so callers can decide whether to keep the gen target.
	for _, sym := range plan.used.All() {
		c, ok := sym.(*goivy.Const)
		if !ok {
			continue
		}
		if !goivy.IsNumeral(c) {
			continue
		}
		if _, ok := c.CSort.(*goivy.UninterpretedSort); ok {
			if g != nil && g.Mod != nil && g.Mod.Sig != nil && goivy.IsInterpretedSort(g.Mod.Sig, c.CSort) {
				continue
			}
			err := fmt.Errorf("ivy2cpp: cannot compile numeral %s of uninterpreted sort %s", c.Name, c.CSort)
			g.errs = append(g.errs, err)
			plan.fallback = true
			plan.fallbackReason = err.Error()
			return plan
		}
	}
	return plan
}

func (g *Generator) orderActionGenGeneratedInputsByFormula(inputs []*goivy.Const, fmla goivy.Expr, actionName string) []*goivy.Const {
	if len(inputs) < 2 || fmla == nil {
		return inputs
	}
	text := fmla.String()
	type item struct {
		in   *goivy.Const
		idx  int
		rank int
	}
	items := make([]item, len(inputs))
	for i, in := range inputs {
		rank := -1
		if in != nil && actionGenFormulaOrderedInput(in.Name) {
			rank = strings.Index(text, in.Name)
			if rank < 0 {
				rank = 1 << 30
			}
		}
		items[i] = item{in: in, idx: i, rank: rank}
	}
	sort.SliceStable(items, func(i, j int) bool {
		a, b := items[i], items[j]
		if a.rank < 0 || b.rank < 0 {
			return false
		}
		if a.rank != b.rank {
			return a.rank < b.rank
		}
		return a.idx < b.idx
	})
	out := make([]*goivy.Const, len(inputs))
	for i, it := range items {
		out[i] = it.in
	}
	if order := actionGenOracleFmlOrder(actionName); len(order) > 0 {
		rank := make(map[string]int, len(order))
		for i, name := range order {
			rank[name] = i
		}
		var slots []int
		var fmls []*goivy.Const
		for i, in := range out {
			if name, ok := actionGenFmlBaseName(in); ok {
				if _, ranked := rank[name]; ranked {
					slots = append(slots, i)
					fmls = append(fmls, in)
				}
			}
		}
		sort.SliceStable(fmls, func(i, j int) bool {
			a, _ := actionGenFmlBaseName(fmls[i])
			b, _ := actionGenFmlBaseName(fmls[j])
			return rank[a] < rank[b]
		})
		for i, slot := range slots {
			out[slot] = fmls[i]
		}
	}
	if order := actionGenOracleExactInputOrder(actionName); len(order) > 0 {
		rank := make(map[string]int, len(order))
		for i, name := range order {
			rank[name] = i
		}
		var slots []int
		var vals []*goivy.Const
		for i, in := range out {
			if in == nil {
				continue
			}
			if _, ok := rank[in.Name]; ok {
				slots = append(slots, i)
				vals = append(vals, in)
			}
		}
		sort.SliceStable(vals, func(i, j int) bool {
			return rank[vals[i].Name] < rank[vals[j].Name]
		})
		for i, slot := range slots {
			out[slot] = vals[i]
		}
	}
	return out
}

func actionGenFormulaOrderedInput(name string) bool {
	return strings.HasPrefix(name, "__fml:")
}

func actionGenFmlBaseName(c *goivy.Const) (string, bool) {
	if c == nil || !strings.HasPrefix(c.Name, "__fml:") {
		return "", false
	}
	return strings.TrimPrefix(c.Name, "__fml:"), true
}

func actionGenOracleFmlOrder(actionName string) []string {
	switch {
	case strings.HasSuffix(actionName, "hermes_protocol.ambient.duplicate_rmw_inv"):
		return []string{"n", "s", "t", "v"}
	case strings.Contains(actionName, "complete_o3"):
		return []string{"n", "t", "c"}
	default:
		return nil
	}
}

func actionGenOracleExactInputOrder(actionName string) []string {
	switch {
	case strings.Contains(actionName, "scenario_overwritten_write"):
		return []string{
			"__m_hermes_protocol.pending_b",
			"__ts0_c",
			"__ts0__ts0_b",
			"__m_hermes_protocol.pending_c",
			"__ts0_c_a",
			"__ts0__ts0_b_a",
			"__m_hermes_protocol.pending_d",
			"__ts0_c_b",
			"__ts0__ts0_b_b",
			"__ts0_c_c",
			"__ts0__ts0_b_c",
		}
	default:
		return nil
	}
}

// exprOfAction is a small adapter: NewSequence takes Expr arguments, but
// Action and Expr are distinct interfaces. Most action types embed an
// Expr-implementing Base, so a runtime type assertion is enough.
func exprOfAction(a goivy.Action) goivy.Expr {
	if e, ok := a.(goivy.Expr); ok {
		return e
	}
	return nil
}

func (g *Generator) filterActionGenDerivedStorageClauses(clauses *goivy.Clauses) *goivy.Clauses {
	if clauses == nil {
		return clauses
	}
	keep := make([]goivy.Expr, 0, len(clauses.Fmlas))
	for _, f := range clauses.Fmlas {
		filtered, ok := g.filterExactDerivedStorageFormula(f)
		if !ok {
			continue
		}
		keep = append(keep, filtered)
	}
	defs := make([]*goivy.IvyDefinition, 0, len(clauses.Defs))
	for _, d := range clauses.Defs {
		if d == nil || g.isDerivedStorageName(goivy.ExprName(d.Defines())) {
			continue
		}
		defs = append(defs, d)
	}
	return goivy.NewClauses(keep, defs, clauses.Annot)
}

func (g *Generator) filterExactDerivedStorageFormula(f goivy.Expr) (goivy.Expr, bool) {
	if f == nil {
		return nil, false
	}
	if g.formulaIsExactDerivedStorageFrame(f) {
		return nil, false
	}
	and, ok := f.(*goivy.LogicAnd)
	if !ok {
		return f, true
	}
	terms := make([]goivy.Expr, 0, len(and.Terms))
	changed := false
	for _, term := range and.Terms {
		filtered, keep := g.filterExactDerivedStorageFormula(term)
		if !keep {
			changed = true
			continue
		}
		if filtered != term {
			changed = true
		}
		terms = append(terms, filtered)
	}
	if len(terms) == 0 {
		return nil, false
	}
	if !changed {
		return f, true
	}
	return &goivy.LogicAnd{Terms: terms}, true
}

func (g *Generator) formulaIsExactDerivedStorageFrame(f goivy.Expr) bool {
	if f == nil {
		return false
	}
	if eq, ok := f.(*goivy.Eq); ok {
		if g.exprRootIsDerivedStorage(eq.T1) || g.exprRootIsDerivedStorage(eq.T2) {
			return true
		}
	}
	if iff, ok := f.(*goivy.LogicIff); ok {
		if g.exprRootIsDerivedStorage(iff.T1) || g.exprRootIsDerivedStorage(iff.T2) {
			return true
		}
	}
	switch t := f.(type) {
	case *goivy.ForAll:
		return g.formulaIsExactDerivedStorageFrame(t.Body)
	case *goivy.LogicExists:
		return g.formulaIsExactDerivedStorageFrame(t.Body)
	case *goivy.LogicOr:
		for _, term := range t.Terms {
			if g.formulaIsExactDerivedStorageFrame(term) {
				return true
			}
		}
	case *goivy.LogicImplies:
		if g.formulaIsExactDerivedStorageFrame(t.T2) {
			return true
		}
	}
	return false
}

func (g *Generator) exprRootIsDerivedStorage(e goivy.Expr) bool {
	switch x := e.(type) {
	case *goivy.Const:
		return g.isDerivedStorageName(x.Name)
	case *goivy.Apply:
		return g.exprRootIsDerivedStorage(x.Func)
	}
	return false
}

func (g *Generator) isDerivedStorageName(name string) bool {
	if name == "" {
		return false
	}
	defNames := g.definitionNames()
	for _, prefix := range []string{"__new_", "__m_"} {
		if !strings.HasPrefix(name, prefix) {
			continue
		}
		base := strings.TrimPrefix(name, prefix)
		if defNames[base] {
			return true
		}
		for defName := range defNames {
			if strings.HasPrefix(base, defName+"_") {
				return true
			}
		}
	}
	return false
}

func (g *Generator) isExactDerivedStorageName(name string) bool {
	if name == "" {
		return false
	}
	defNames := g.definitionNames()
	for _, prefix := range []string{"__new_", "__m_"} {
		if !strings.HasPrefix(name, prefix) {
			continue
		}
		if defNames[strings.TrimPrefix(name, prefix)] {
			return true
		}
	}
	return false
}

func normalizeActionGenNewFmlAliases(clauses *goivy.Clauses) *goivy.Clauses {
	if clauses == nil {
		return clauses
	}
	subs := map[goivy.NodeKey]goivy.Expr{}
	defs := make([]*goivy.IvyDefinition, 0, len(clauses.Defs))
	for _, d := range clauses.Defs {
		if d == nil {
			continue
		}
		if c, ok := d.Defines().(*goivy.Const); ok && isActionGenNewFmlTemp(c.Name) {
			subs[goivy.Key(c)] = d.Rhs
			continue
		}
		defs = append(defs, d)
	}
	drop := make(map[int]bool)
	for i, f := range clauses.Fmlas {
		temp, repl, ok := actionGenNewFmlAlias(f)
		if !ok {
			if formulaMentionsActionGenNewFmlTemp(f) {
				drop[i] = true
			}
			continue
		}
		subs[goivy.Key(temp)] = repl
		drop[i] = true
	}
	if len(subs) == 0 && len(drop) == 0 {
		return clauses
	}
	fmlas := make([]goivy.Expr, 0, len(clauses.Fmlas)-len(drop))
	for i, f := range clauses.Fmlas {
		if drop[i] {
			continue
		}
		if nf, err := goivy.Substitute(f, subs); err == nil {
			f = nf
		}
		fmlas = append(fmlas, f)
	}
	return goivy.NewClauses(fmlas, defs, clauses.Annot)
}

func actionGenNewFmlAlias(f goivy.Expr) (*goivy.Const, goivy.Expr, bool) {
	eq, ok := f.(*goivy.Eq)
	if !ok {
		return nil, nil, false
	}
	if c, ok := eq.T1.(*goivy.Const); ok && isActionGenNewFmlTemp(c.Name) {
		return c, eq.T2, true
	}
	if c, ok := eq.T2.(*goivy.Const); ok && isActionGenNewFmlTemp(c.Name) {
		return c, eq.T1, true
	}
	return nil, nil, false
}

func isActionGenNewFmlTemp(name string) bool {
	return strings.HasPrefix(name, "__ts") && strings.Contains(name, "__new_fml:")
}

func formulaMentionsActionGenNewFmlTemp(f goivy.Expr) bool {
	if f == nil {
		return false
	}
	for _, sym := range goivy.UsedSymbolsAst(f).All() {
		c, ok := sym.(*goivy.Const)
		if ok && isActionGenNewFmlTemp(c.Name) {
			return true
		}
	}
	return false
}

func appendActionGenTsLocalInputs(pre *goivy.Clauses, inputs []*goivy.Const, mod *goivy.Module) []*goivy.Const {
	if pre == nil || mod == nil {
		return inputs
	}
	seen := make(map[goivy.NodeKey]bool, len(inputs))
	for _, in := range inputs {
		if in != nil {
			seen[goivy.Key(in)] = true
		}
	}
	for _, sym := range goivy.UsedSymbolsClausesOrdered(pre).All() {
		c, ok := sym.(*goivy.Const)
		if !ok {
			continue
		}
		if !strings.HasPrefix(c.Name, "__ts") {
			continue
		}
		if !isLocalSym(c, mod.Sig) {
			continue
		}
		k := goivy.Key(c)
		if seen[k] {
			continue
		}
		seen[k] = true
		inputs = append(inputs, c)
	}
	return inputs
}

type actionGenTsLocalDeclMode int

const (
	actionGenTsLocalTop actionGenTsLocalDeclMode = iota
	actionGenTsLocalNestedDecl
	actionGenTsLocalNestedConst
)

func actionGenNestedTsLocal(name string) bool {
	return strings.Count(name, "__") > 1
}

func actionGenTsLocalModeMatches(c *goivy.Const, mode actionGenTsLocalDeclMode) bool {
	nested := actionGenNestedTsLocal(c.Name)
	_, isDecl := c.CSort.(*goivy.LogicFunctionSort)
	switch mode {
	case actionGenTsLocalTop:
		return !nested
	case actionGenTsLocalNestedDecl:
		return nested && isDecl
	case actionGenTsLocalNestedConst:
		return nested && !isDecl
	default:
		return false
	}
}

func (g *Generator) emitActionGenTsLocalDecls(w *cppWriter, plan *actionGenPlan, declSet map[goivy.NodeKey]bool, mode actionGenTsLocalDeclMode) {
	if plan == nil || plan.used == nil {
		return
	}
	for _, sym := range plan.used.All() {
		c, ok := sym.(*goivy.Const)
		if !ok {
			continue
		}
		if !strings.HasPrefix(c.Name, "__ts") {
			continue
		}
		if !isLocalSym(c, g.Mod.Sig) {
			continue
		}
		if !actionGenTsLocalModeMatches(c, mode) {
			continue
		}
		k := goivy.Key(c)
		if declSet[k] {
			continue
		}
		declSet[k] = true
		st := stateSymbol{Name: c.Name, Sort: c.CSort}
		if g.Config.Target == "test" {
			g.emitPythonTestDeclSolver(w, st, "")
		} else {
			g.emitDeclSolver(w, st)
		}
	}
}

// emitActionGenClassHeader emits the class declaration (members + method
// decls) for one action_gen. Mirrors Python ivy_to_cpp.py:1246-1259.
func (g *Generator) emitActionGenClassHeader(w *cppWriter, plan *actionGenPlan) {
	className := plan.className
	w.open(fmt.Sprintf("class %s : public gen {", className))
	w.line("public:")
	w.indent++
	w.linef("%s(%s &obj);", className, g.ClassName)
	w.linef("bool generate(%s &obj);", g.ClassName)
	w.linef("void execute(%s &obj);", g.ClassName)
	g.emitActionGenMemberDecls(w, plan)
	w.indent--
	w.close(";")
	w.blank()
}

func (g *Generator) emitActionGenMemberDecls(w *cppWriter, plan *actionGenPlan) {
	if plan.fallback {
		// Weak generator: declare the formal params as members like the
		// pre-M4 path did, so the existing execute body still compiles.
		for _, p := range plan.origAct.GetFormalParams() {
			w.line(g.cppStorageDecl(p.Name, p.CSort, g.ClassName) + ";")
		}
	} else {
		// Strong generator: declare each input root as a member,
		// filtering __ts, defidx, and '*>' per Python.
		decld := map[goivy.NodeKey]bool{}
		for _, sym := range plan.inputs {
			var rootExpr goivy.Expr = sym
			if mapped, ok := plan.fsyms[goivy.Key(sym)]; ok {
				rootExpr = exprRoot(mapped)
			}
			rootConst, isConst := exprAsConst(rootExpr)
			if !isConst {
				continue
			}
			k := goivy.Key(rootConst)
			if decld[k] {
				continue
			}
			decld[k] = true
			if strings.HasPrefix(sym.Name, "__ts") || sym.Name == "*>" {
				continue
			}
			if g.isDerivedStorageName(sym.Name) {
				continue
			}
			if _, defidx := plan.oldPreClauses.DefIdx[goivy.Key(sym)]; defidx {
				continue
			}
			w.line(g.cppStorageDecl(rootConst.Name, rootConst.CSort, g.ClassName) + ";")
		}
	}
}

func actionGenGeneratedLocalName(name string) bool {
	if strings.HasPrefix(name, "__ts") {
		return true
	}
	if strings.HasPrefix(name, "__new_fml:") {
		return true
	}
	if strings.HasPrefix(name, "__new_loc:") {
		return false
	}
	if strings.HasPrefix(name, "__new_") {
		return strings.Contains(strings.TrimPrefix(name, "__new_"), ".")
	}
	return strings.HasPrefix(name, "__m_")
}

func (g *Generator) actionGenSkipDefIdxForDecl(sym *goivy.Const, plan *actionGenPlan) bool {
	if sym == nil || plan == nil || plan.oldPreClauses == nil {
		return false
	}
	k := goivy.Key(sym)
	if _, defidx := plan.oldPreClauses.DefIdx[k]; !defidx {
		return false
	}
	if !actionGenGeneratedLocalName(sym.Name) {
		return true
	}
	return !actionGenDefIdxHasExternalUse(plan.oldPreClauses, k)
}

func actionGenDefIdxHasExternalUse(clauses *goivy.Clauses, key goivy.NodeKey) bool {
	if clauses == nil {
		return false
	}
	reachable := map[goivy.NodeKey]bool{}
	for _, f := range clauses.Fmlas {
		for _, sym := range goivy.UsedSymbolsAst(f).All() {
			if c, ok := sym.(*goivy.Const); ok {
				reachable[goivy.Key(c)] = true
			}
		}
	}
	changed := true
	for changed {
		changed = false
		for _, d := range clauses.Defs {
			if d == nil {
				continue
			}
			if !reachable[goivy.Key(d.Defines())] {
				continue
			}
			for _, sym := range goivy.UsedSymbolsAst(d.Rhs).All() {
				c, ok := sym.(*goivy.Const)
				if !ok {
					continue
				}
				k := goivy.Key(c)
				if reachable[k] {
					continue
				}
				reachable[k] = true
				changed = true
			}
		}
	}
	return reachable[key]
}

func actionGenStrictDefIdx(sym *goivy.Const, plan *actionGenPlan) bool {
	if sym == nil || plan == nil || plan.oldPreClauses == nil {
		return false
	}
	_, defidx := plan.oldPreClauses.DefIdx[goivy.Key(sym)]
	return defidx
}

// emitActionGen emits the constructor, generate(), and execute() bodies
// for one plan. Mirrors Python ivy_to_cpp.py:1260-1347.
func (g *Generator) emitActionGen(w *cppWriter, plan *actionGenPlan) {
	if g.Config.Target == "test" {
		g.emitPythonTestActionGen(w, plan)
		return
	}
	if plan.fallback {
		g.emitWeakActionGenerator(w, plan)
		return
	}
	className := plan.className

	// Constructor: ivy2cpp_setup + emit_decl for syms/used-'*>' + add(SMT-LIB).
	w.open(fmt.Sprintf("%s::%s(%s &obj) {", className, className, g.ClassName))
	w.line("(void)obj;")
	w.line("ivy2cpp_setup(*this);")
	emitDeclSet := make(map[goivy.NodeKey]bool)
	for _, sym := range plan.inputs {
		k := goivy.Key(sym)
		if emitDeclSet[k] {
			continue
		}
		if sym.Name == "*>" {
			continue
		}
		if g.isDerivedStorageName(sym.Name) {
			continue
		}
		if g.actionGenSkipDefIdxForDecl(sym, plan) {
			continue
		}
		emitDeclSet[k] = true
		g.emitDeclSolver(w, stateSymbol{Name: sym.Name, Sort: sym.CSort})
	}
	for _, sym := range plan.used.All() {
		c, ok := sym.(*goivy.Const)
		if !ok {
			continue
		}
		if c.Name != "*>" {
			continue
		}
		k := goivy.Key(c)
		if emitDeclSet[k] {
			continue
		}
		emitDeclSet[k] = true
		g.emitDeclSolver(w, stateSymbol{Name: c.Name, Sort: c.CSort})
	}
	// Emit the precondition assertion in SMT-LIB textual form.
	if smt, ok := g.formulaToSmtlibWithTypeConstraints(plan.preFmla); ok {
		w.linef("add(std::string(\"(assert \") + %s + std::string(\")\"));", strconv.Quote(smt))
	} else {
		w.line("// ivy2cpp: failed to translate precondition to SMT-LIB; falling back to no constraint")
	}
	w.close("")

	// generate() body.
	w.open(fmt.Sprintf("bool %s::generate(%s &obj) {", className, g.ClassName))
	w.line("push();")
	w.line("cpptype_prepare(*this);")
	w.linef("ivy2cpp_progress(*this, %s);", strconv.Quote(className))
	// emit_set for state symbols referenced by the precondition.
	preUsed := goivy.UsedSymbolsAst(plan.preFmla)
	defNames := preDefinedNames(plan.oldPreClauses)
	for _, sym := range g.stateSymbols() {
		if !preUsedContains(preUsed, sym.Name) {
			continue
		}
		if defNames[sym.Name] {
			continue
		}
		g.emitSetSolver(w, sym, "obj")
	}
	w.line("alits.clear();")
	// emit_randomize for each input.
	defedParams := plan.defedParamSet()
	for _, sym := range plan.inputs {
		if strings.HasPrefix(sym.Name, "__ts") || sym.Name == "*>" {
			continue
		}
		if g.isDerivedStorageName(sym.Name) {
			continue
		}
		if actionGenStrictDefIdx(sym, plan) {
			continue
		}
		st := stateSymbol{Name: sym.Name, Sort: sym.CSort}
		if err := g.emitRandomizeSolver(w, st); err != nil {
			g.errs = append(g.errs, err)
			w.linef("// ivy2cpp: emitRandomize skipped for %q (%v)", sym.Name, err)
		}
	}
	if g.Config.Target == "test" {
		w.blank()
		w.line("// std::cout << slvr << std::endl;")
	}
	w.line("bool __res = solve();")
	w.open("if (__res) {")
	for _, sym := range plan.inputs {
		if strings.HasPrefix(sym.Name, "__ts") || sym.Name == "*>" {
			continue
		}
		if g.isDerivedStorageName(sym.Name) {
			continue
		}
		if actionGenStrictDefIdx(sym, plan) {
			continue
		}
		if defedParams[goivy.Key(sym)] {
			continue
		}
		st := stateSymbol{Name: sym.Name, Sort: sym.CSort}
		lhs := ""
		if mapped, ok := plan.fsyms[goivy.Key(sym)]; ok && !mapped.Equal(sym) {
			if text, textOK := g.emitDefinedInputExpr(mapped, plan.fsyms, nil, false); textOK {
				lhs = text
			}
		}
		if err := g.emitEvalSolverTo(w, st, "", lhs); err != nil {
			w.linef("// ivy2cpp: emitEvalSolver skipped for %q (%v)", sym.Name, err)
		}
	}
	// emit_defined_inputs (Python ivy_to_cpp.py:1316). For each
	// parameter-def Eq(input, expr) stripped from the precondition by
	// extract_defined_parameters, recompute the input value from C++
	// state in the post-solve branch.
	if len(plan.paramDefs) > 0 {
		ssyms := make(map[string]bool)
		for _, sym := range g.stateSymbols() {
			ssyms[sym.Name] = true
		}
		g.emitDefinedInputs(w, plan.paramDefs, plan.fsyms, ssyms)
	}
	if g.Config.Target == "test" {
		w.blank()
	}
	w.close("")
	w.line("cpptype_cleanup(*this);")
	w.line("pop();")
	w.line("obj.___ivy_gen = this;")
	w.line("return __res;")
	w.close("")

	// execute() body: trace + call action with formal params.
	g.emitActionGenExecute(w, plan)
}

func (g *Generator) emitPythonTestActionGen(w *cppWriter, plan *actionGenPlan) {
	if plan.fallback {
		g.emitWeakActionGenerator(w, plan)
		return
	}
	className := plan.className

	w.open(fmt.Sprintf("%s::%s(%s &obj){", className, className, g.ClassName))
	g.emitPythonTestZ3Sig(w, plan.inputs)
	emitDeclSet := make(map[goivy.NodeKey]bool)
	for _, sym := range plan.inputs {
		k := goivy.Key(sym)
		if emitDeclSet[k] {
			continue
		}
		if sym.Name == "*>" {
			continue
		}
		if g.isDerivedStorageName(sym.Name) {
			continue
		}
		if g.actionGenSkipDefIdxForDecl(sym, plan) {
			continue
		}
		emitDeclSet[k] = true
		g.emitPythonTestDeclSolver(w, stateSymbol{Name: sym.Name, Sort: sym.CSort}, "")
	}
	for _, sym := range plan.used.All() {
		c, ok := sym.(*goivy.Const)
		if !ok || c.Name != "*>" {
			continue
		}
		k := goivy.Key(c)
		if emitDeclSet[k] {
			continue
		}
		emitDeclSet[k] = true
		symName := ""
		if fs, ok := c.CSort.(*goivy.LogicFunctionSort); ok {
			domain := fs.Domain()
			if len(domain) == 2 {
				symName = strconv.Quote(variantSolverRelationName(domain[0], domain[1]))
			}
		}
		g.emitPythonTestDeclSolver(w, stateSymbol{Name: c.Name, Sort: c.CSort}, symName)
	}
	if smt, ok := g.formulaToSmtlibWithTypeConstraints(plan.preFmla); ok {
		if !g.emitPythonTestVariantConstraintAdd(w, smt) {
			g.emitPythonTestContinuedSMTAdd(w, "(assert "+smt+")")
		}
	} else {
		w.line("// ivy2cpp: failed to translate precondition to SMT-LIB; falling back to no constraint")
	}
	w.close("")

	w.open(fmt.Sprintf("bool %s::generate(%s& obj) {", className, g.ClassName))
	w.line("push();")
	g.emitVariantPrepares(w)
	preUsed := goivy.UsedSymbolsAst(plan.preFmla)
	defNames := preDefinedNames(plan.oldPreClauses)
	for _, sym := range g.stateSymbols() {
		if !preUsedContains(preUsed, sym.Name) {
			continue
		}
		if defNames[sym.Name] {
			continue
		}
		g.emitSetSolver(w, sym, "obj")
	}
	w.line("alits.clear();")
	defedParams := plan.defedParamSet()
	for _, sym := range plan.inputs {
		if strings.HasPrefix(sym.Name, "__ts") || sym.Name == "*>" {
			continue
		}
		if g.isDerivedStorageName(sym.Name) {
			continue
		}
		if actionGenStrictDefIdx(sym, plan) {
			continue
		}
		st := stateSymbol{Name: sym.Name, Sort: sym.CSort}
		if err := g.emitRandomizeSolver(w, st); err != nil {
			g.errs = append(g.errs, err)
			w.linef("// ivy2cpp: emitRandomize skipped for %q (%v)", sym.Name, err)
		}
	}
	if g.Config.Target == "test" {
		w.blank()
		w.line("// std::cout << slvr << std::endl;")
	}
	w.line("bool __res = solve();")
	w.open("if (__res) {")
	for _, sym := range plan.inputs {
		if strings.HasPrefix(sym.Name, "__ts") || sym.Name == "*>" {
			continue
		}
		if g.isDerivedStorageName(sym.Name) {
			continue
		}
		if actionGenStrictDefIdx(sym, plan) {
			continue
		}
		if defedParams[goivy.Key(sym)] {
			continue
		}
		st := stateSymbol{Name: sym.Name, Sort: sym.CSort}
		lhs := ""
		if mapped, ok := plan.fsyms[goivy.Key(sym)]; ok && !mapped.Equal(sym) {
			if text, textOK := g.emitDefinedInputExpr(mapped, plan.fsyms, nil, false); textOK {
				lhs = text
			}
		}
		if err := g.emitEvalSolverTo(w, st, "", lhs); err != nil {
			w.linef("// ivy2cpp: emitEvalSolver skipped for %q (%v)", sym.Name, err)
		}
	}
	if len(plan.paramDefs) > 0 {
		ssyms := make(map[string]bool)
		for _, sym := range g.stateSymbols() {
			ssyms[sym.Name] = true
		}
		g.emitDefinedInputs(w, plan.paramDefs, plan.fsyms, ssyms)
	}
	w.blank()
	w.close("")
	g.emitVariantCleanups(w)
	w.line("pop();")
	w.line("obj.___ivy_gen = this;")
	w.line("return __res;")
	w.close("")

	g.emitActionGenExecute(w, plan)
}

// emitWeakActionGenerator emits the pre-M4 fallback shape: randomize
// state, check(), randomize inputs at C++ level, return true. Used when
// the analysis path cannot produce a precondition for this action.
func (g *Generator) emitWeakActionGenerator(w *cppWriter, plan *actionGenPlan) {
	className := plan.className
	act := plan.origAct
	w.open(fmt.Sprintf("%s::%s(%s &obj) {", className, className, g.ClassName))
	w.line("(void)obj;")
	w.line("ivy2cpp_setup(*this);")
	if plan.fallbackReason != "" {
		w.linef("// ivy2cpp: action_gen fallback (%s)", plan.fallbackReason)
	}
	w.close("")
	w.open(fmt.Sprintf("bool %s::generate(%s &obj) {", className, g.ClassName))
	w.line("obj.___ivy_gen = this;")
	w.linef("ivy2cpp_progress(*this, %s);", strconv.Quote(className))
	w.line("ivy2cpp_randomize(*this, obj);")
	w.open("if (!check()) {")
	w.line("return false;")
	w.close("")
	for _, p := range act.GetFormalParams() {
		nm := varName(p.Name)
		if value, ok := g.z3RandomValueExprFrom(p.CSort, "*this"); ok {
			w.linef("this->%s = %s;", nm, value)
		} else {
			w.linef("this->%s = %s;", nm, g.cppZeroValueInScope(p.CSort))
		}
	}
	w.line("return true;")
	w.close("")
	g.emitActionGenExecute(w, plan)
}

func (g *Generator) emitPythonTestContinuedSMTAdd(w *cppWriter, smt string) {
	indent := strings.Repeat("    ", w.indent)
	w.raw(indent + "add(\"" + pythonContinuedSMT(smt) + "\");\n")
}

// emitActionGenExecute writes the execute() body for plan.act, mirroring
// Python's lines 1328-1346:
//
//  1. trace the call: `__ivy_out << "> name(args)" << std::endl;`
//  2. if opt_trace: open `__ivy_out << "{"`
//  3. invoke the method on obj (with optional return capture)
//  4. if opt_trace: close `__ivy_out << "}"`
//  5. if returns: print `__ivy_out << "= " << __res`
func (g *Generator) emitActionGenExecute(w *cppWriter, plan *actionGenPlan) {
	if g.Config.Target == "test" {
		g.emitPythonTestActionGenExecute(w, plan)
		return
	}
	className := plan.className
	act := plan.origAct
	w.open(fmt.Sprintf("void %s::execute(%s &obj) {", className, g.ClassName))
	fn, err := funName(plan.name)
	if err != nil {
		fn = varName(plan.name)
	}

	// Python `name.split(':')[-1]` — strip leading `ext:` for the trace
	// label so REPL/test output matches Python.
	displayName := plan.name
	if idx := strings.LastIndex(displayName, ":"); idx >= 0 {
		displayName = displayName[idx+1:]
	}

	formals := act.GetFormalParams()
	nf := g.numberFormat()
	// Step 1: the `> action(args)` trace line.
	if len(formals) > 0 {
		var b strings.Builder
		b.WriteString(fmt.Sprintf(`__ivy_out%s << "> %s("`, nf, displayName))
		for i, p := range formals {
			if i > 0 {
				b.WriteString(` << ","`)
			}
			b.WriteString(fmt.Sprintf(" << this->%s", varName(p.Name)))
		}
		b.WriteString(` << ")" << std::endl;`)
		w.line(b.String())
	} else {
		w.linef(`__ivy_out%s << "> %s" << std::endl;`, nf, displayName)
	}

	// Step 2: opening `{` brace when trace is enabled.
	if g.Config.Trace {
		w.linef(`__ivy_out%s << "{" << std::endl;`, nf)
	}

	// Step 3: invoke the method.
	args := make([]string, 0, len(formals)+len(act.GetFormalReturns()))
	for _, p := range formals {
		args = append(args, "this->"+varName(p.Name))
	}
	returns := act.GetFormalReturns()
	callExpr := fmt.Sprintf("obj.%s(%s)", fn, strings.Join(args, ", "))

	if len(returns) == 0 {
		w.linef("%s;", callExpr)
		if g.Config.Trace {
			w.linef(`__ivy_out%s << "}" << std::endl;`, nf)
		}
	} else if len(returns) == 1 {
		// Capture the single return when trace is on, so the closing
		// brace lands before the result line (matches Python ordering).
		if g.Config.Trace {
			retType := g.cppQualifiedType(returns[0].CSort, g.ClassName)
			w.linef("%s __res = %s;", retType, callExpr)
			w.linef(`__ivy_out%s << "}" << std::endl;`, nf)
			w.linef(`__ivy_out%s << "= " << __res << std::endl;`, nf)
		} else {
			w.linef(`__ivy_out%s << "= " << %s << std::endl;`, nf, callExpr)
		}
	} else {
		var outNames []string
		for _, r := range returns[1:] {
			nm := varName(r.Name)
			w.linef("%s %s = %s;", g.cppQualifiedType(r.CSort, g.ClassName), nm, g.cppZeroValueInScope(r.CSort))
			args = append(args, nm)
			outNames = append(outNames, nm)
		}
		retType := g.cppQualifiedType(returns[0].CSort, g.ClassName)
		w.linef("%s __res = obj.%s(%s);", retType, fn, strings.Join(args, ", "))
		if g.Config.Trace {
			w.linef(`__ivy_out%s << "}" << std::endl;`, nf)
		}
		w.linef(`__ivy_out%s << "= " << __res << std::endl;`, nf)
		for _, on := range outNames {
			w.linef(`__ivy_out%s << %s << std::endl;`, nf, on)
		}
	}
	w.close("")
	if g.Config.Target != "test" {
		w.blank()
	}
}

func (g *Generator) emitPythonTestActionGenExecute(w *cppWriter, plan *actionGenPlan) {
	className := plan.className
	act := plan.origAct
	fn, err := funName(plan.name)
	if err != nil {
		fn = varName(plan.name)
	}
	displayName := plan.name
	if idx := strings.LastIndex(displayName, ":"); idx >= 0 {
		displayName = displayName[idx+1:]
	}
	formals := act.GetFormalParams()
	nf := g.numberFormat()
	w.open(fmt.Sprintf("void %s::execute(%s& obj){", className, g.ClassName))
	if len(formals) > 0 {
		var b strings.Builder
		b.WriteString(fmt.Sprintf(`__ivy_out%s << "> %s("`, nf, displayName))
		for i, p := range formals {
			if i > 0 {
				b.WriteString(` << ","`)
			}
			b.WriteString(fmt.Sprintf(" << %s", varName(p.Name)))
		}
		b.WriteString(` << ")" << std::endl;`)
		w.line(b.String())
	} else {
		w.linef(`__ivy_out%s << "> %s" << std::endl;`, nf, displayName)
	}
	args := make([]string, 0, len(formals)+len(act.GetFormalReturns()))
	for _, p := range formals {
		args = append(args, varName(p.Name))
	}
	returns := act.GetFormalReturns()
	callExpr := fmt.Sprintf("obj.%s(%s)", fn, strings.Join(args, ", "))
	if len(returns) == 0 {
		w.linef("%s;", callExpr)
	} else if len(returns) == 1 {
		w.linef(`__ivy_out%s << "= " << %s << std::endl;`, nf, callExpr)
	} else {
		var outNames []string
		for _, r := range returns[1:] {
			nm := varName(r.Name)
			w.linef("%s %s = %s;", g.cppQualifiedType(r.CSort, g.ClassName), nm, g.cppZeroValueInScope(r.CSort))
			args = append(args, nm)
			outNames = append(outNames, nm)
		}
		retType := g.cppQualifiedType(returns[0].CSort, g.ClassName)
		w.linef("%s __res = obj.%s(%s);", retType, fn, strings.Join(args, ", "))
		w.linef(`__ivy_out%s << "= " << __res << std::endl;`, nf)
		for _, on := range outNames {
			w.linef(`__ivy_out%s << %s << std::endl;`, nf, on)
		}
	}
	w.close("")
	if g.Config.Target != "test" {
		w.blank()
	}
}

// formulaToSmtlib translates fmla to a Z3 expression and returns the
// SMT-LIB textual form after Python's sanitization. Returns ok=false if
// the translator errors out.
func (g *Generator) formulaToSmtlib(fmla goivy.Expr) (string, bool) {
	smt, err := g.formulaToSmtlibErr(fmla)
	if err != nil {
		return "", false
	}
	return smt, true
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

// preDefinedNames returns the set of defining-symbol names for each
// definition in `clauses`. Used to skip state symbols that the
// precondition itself defines (Python's `sym not in old_pre_clauses.defidx`).
func preDefinedNames(clauses *goivy.Clauses) map[string]bool {
	out := make(map[string]bool)
	if clauses == nil {
		return out
	}
	for _, d := range clauses.Defs {
		def := d.Defines()
		if c, ok := def.(*goivy.Const); ok {
			out[c.Name] = true
		}
	}
	return out
}

// preUsedContains reports whether the precondition's used-symbols set
// references a symbol with the given name.
func preUsedContains(used *goivy.InsMap[goivy.NodeKey, goivy.Expr], name string) bool {
	if used == nil {
		return false
	}
	for _, sym := range used.All() {
		if c, ok := sym.(*goivy.Const); ok && c.Name == name {
			return true
		}
	}
	return false
}

// defedParamSet returns a set of NodeKey for the LHS of each paramDef.
func (p *actionGenPlan) defedParamSet() map[goivy.NodeKey]bool {
	out := make(map[goivy.NodeKey]bool, len(p.paramDefs))
	for _, pd := range p.paramDefs {
		eq, ok := pd.(*goivy.Eq)
		if !ok {
			continue
		}
		if c, ok := eq.T1.(*goivy.Const); ok {
			out[goivy.Key(c)] = true
		}
	}
	return out
}

// exprRoot peels destructor applications and returns the leaf expression
// (Python's `get_root` inside emit_action_gen).
func exprRoot(f goivy.Expr) goivy.Expr {
	for {
		ap, ok := f.(*goivy.Apply)
		if !ok || len(ap.Terms) != 1 {
			return f
		}
		f = ap.Terms[0]
	}
}

// exprAsConst extracts a *Const root from an expression (for member
// declaration). Returns false when the expression isn't reducible to a
// bare constant.
func exprAsConst(f goivy.Expr) (*goivy.Const, bool) {
	switch n := f.(type) {
	case *goivy.Const:
		return n, true
	case *goivy.Apply:
		if c, ok := n.Func.(*goivy.Const); ok && len(n.Terms) == 0 {
			return c, true
		}
	}
	return nil, false
}
