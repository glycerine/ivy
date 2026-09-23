package ivy2golang

import (
	"fmt"
	"sort"
	"strings"

	"github.com/glycerine/ivy/goivy"
)

// actionGenPlan is the Go-output counterpart of ivy2cpp.actionGenPlan. It is
// pure analysis: it records the reverse-image formula and solver inputs needed
// by Python/C++ style action generators.
type actionGenPlan struct {
	name           string
	className      string
	origAct        goivy.Action
	act            goivy.Action
	inputs         []*goivy.Const
	fsyms          map[goivy.NodeKey]goivy.Expr
	paramDefs      []goivy.Expr
	oldPreClauses  *goivy.Clauses
	origPreClauses *goivy.Clauses
	preFmla        goivy.Expr
	used           *goivy.InsMap[goivy.NodeKey, goivy.Expr]
	fallback       bool
	fallbackReason string
}

func (g *Generator) buildActionGenPlan(name string, act goivy.Action) *actionGenPlan {
	plan := &actionGenPlan{
		name:      name,
		className: g.goActionGeneratorTypeName(name),
		origAct:   act,
		act:       act,
	}
	if g == nil || g.Mod == nil || act == nil {
		plan.fallback = true
		plan.fallbackReason = "nil generator, module, or action"
		return plan
	}
	if g.Mod.BeforeExport != nil {
		if be, ok := g.Mod.BeforeExport.Get2(name); ok && be != nil {
			plan.act = be
		}
	}
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
	truePre := goivy.TrueClauses(nil)
	preClauses := goivy.ReverseImage(truePre, truePre, upd)
	plan.origPreClauses = preClauses
	preClauses = goivy.TrimClauses(preClauses)
	preClauses = expandFieldReferences(preClauses, g.Mod.DestructorSorts)
	preClauses = g.filterActionGenDerivedStorageClauses(preClauses)

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
	preClauses, inputs, plan.fsyms = extractInputFields(preClauses, inputs, g.Mod)
	inputs = appendActionGenTsLocalInputs(preClauses, inputs, g.Mod)
	plan.oldPreClauses = preClauses
	preClauses, plan.paramDefs = extractDefinedParameters(preClauses, inputs)

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
			rdefFmlas = append(rdefFmlas, goivy.DefinitionToConstraint(fixDefinition(def)))
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

	for _, sym := range plan.used.All() {
		c, ok := sym.(*goivy.Const)
		if !ok || !goivy.IsNumeral(c) {
			continue
		}
		if _, ok := c.CSort.(*goivy.UninterpretedSort); ok {
			if g.Mod.Sig != nil && goivy.IsInterpretedSort(g.Mod.Sig, c.CSort) {
				continue
			}
			err := fmt.Errorf("ivy2golang: cannot compile numeral %s of uninterpreted sort %s", c.Name, c.CSort)
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
	for _, prefix := range []string{"__new_", "__m_"} {
		if !strings.HasPrefix(name, prefix) {
			continue
		}
		base := strings.TrimPrefix(name, prefix)
		if g.isDefinitionName(base) {
			return true
		}
		for _, d := range g.allDefinitions() {
			if strings.HasPrefix(base, d.Name+"_") {
				return true
			}
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
		if !ok || !strings.HasPrefix(c.Name, "__ts") {
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

func actionGenGeneratedLocalName(name string) bool {
	if strings.HasPrefix(name, "__ts") || strings.HasPrefix(name, "__new_fml:") {
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
			if d == nil || !reachable[goivy.Key(d.Defines())] {
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

func preDefinedKeys(clauses *goivy.Clauses) map[goivy.NodeKey]bool {
	out := make(map[goivy.NodeKey]bool)
	if clauses == nil {
		return out
	}
	for key := range clauses.DefIdx {
		out[key] = true
	}
	return out
}

func preUsedContainsStateSymbol(used *goivy.InsMap[goivy.NodeKey, goivy.Expr], sym stateSymbol) bool {
	if used == nil {
		return false
	}
	key := stateSymbolNodeKey(sym)
	for _, usedSym := range used.All() {
		if c, ok := usedSym.(*goivy.Const); ok && goivy.Key(c) == key {
			return true
		}
	}
	return false
}

func stateSymbolNodeKey(sym stateSymbol) goivy.NodeKey {
	return goivy.Key(goivy.NewConst(sym.Name, sym.Sort))
}

func (p *actionGenPlan) defedParamSet() map[goivy.NodeKey]bool {
	out := make(map[goivy.NodeKey]bool, len(p.paramDefs))
	for _, pd := range p.paramDefs {
		lhs, _, ok := runtimeActionSolverDefinedInputTerms(pd)
		if !ok {
			continue
		}
		if c, ok := exprAsConst(lhs); ok {
			out[goivy.Key(c)] = true
		}
	}
	return out
}

func exprRoot(f goivy.Expr) goivy.Expr {
	for {
		ap, ok := f.(*goivy.Apply)
		if !ok || len(ap.Terms) != 1 {
			return f
		}
		f = ap.Terms[0]
	}
}

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
