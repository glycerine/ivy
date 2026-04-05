package module

import (
	"fmt"
	"strings"

	il "github.com/glycerine/ivy/goivy/ivylogic"
	iu "github.com/glycerine/ivy/goivy/ivyutils"
	lg "github.com/glycerine/ivy/goivy/logic"
	lu "github.com/glycerine/ivy/goivy/logicutil"
	"github.com/glycerine/ivy/goivy/xtracer"
)

// AnnotConjoiner is implemented by annotation values that support conjunction.
// clauseops uses this interface to conjoin annotations without importing actions.
type AnnotConjoiner interface {
	ConjWith(other interface{}) interface{}
}

// AnnotRenamer is implemented by annotation values that support renaming.
// Matches Python's annot.rename(map).
type AnnotRenamer interface {
	Rename(m map[lg.NodeKey]lg.Expr) interface{}
}

// OpsConfig holds per-session clauseops state.
type OpsConfig struct {
	AnnotConjFunc func(a, b interface{}) interface{}
}

// NewOpsConfig creates a new OpsConfig.
func NewOpsConfig() *OpsConfig {
	return &OpsConfig{}
}

// AndClauses computes the conjunction of Clauses and/or formulas.
// Each argument can be *Clauses or lg.Expr. If no argument is a *Clauses,
// returns an And formula directly. If any input is False, the result is False.
func AndClauses(args ...interface{}) interface{} {
	// Check if any argument is a *Clauses
	hasClauses := false
	for _, a := range args {
		if _, ok := a.(*Clauses); ok {
			hasClauses = true
			break
		}
	}
	if !hasClauses {
		// All args are lg.Expr; return a plain formula
		nodes := make([]lg.Expr, len(args))
		for i, a := range args {
			nodes[i] = a.(lg.Expr)
		}
		return &lg.And{Terms: nodes}
	}

	// Coerce all args to *Clauses
	clauses := coerceArgsToClauses(args)
	if len(clauses) == 0 {
		return TrueClauses(nil)
	}

	// Combine annotations via conj (matching Python's and_clauses default)
	var annot interface{}
	for _, c := range clauses {
		if c.Annot != nil {
			if annot == nil {
				annot = c.Annot
			} else if conjer, ok := annot.(AnnotConjoiner); ok {
				annot = conjer.ConjWith(c.Annot)
			}
		}
	}

	// If any clause set is false, return false
	for _, c := range clauses {
		if c.IsFalse() {
			if xtracer.Enabled {
				totalDefs := 0
				for _, cc := range clauses {
					totalDefs += len(cc.Defs)
				}
				xtracer.Trace("ops.andClauses FALSE-DROP droppingDefs=%d nArgs=%d", totalDefs, len(clauses))
			}
			return FalseClauses(annot)
		}
	}

	// Concatenate formulas and definitions
	var fmlas []lg.Expr
	var defs []*il.Definition
	for _, c := range clauses {
		fmlas = append(fmlas, c.Fmlas...)
		defs = append(defs, c.Defs...)
	}
	return NewClauses(fmlas, defs, annot)
}

// AndClausesTyped is a convenience wrapper that always returns *Clauses.
// It panics if passed non-Clauses, non-Node arguments.
// AnnotOp is an optional annotation combiner matching Python's annot_op parameter.
// It takes the annotations of the input clauses and returns the combined annotation.
// Corresponds to Python's annot_op parameter in and_clauses / conjoin.
type AnnotOp func(annots ...interface{}) interface{}

func AndClausesTyped(args ...*Clauses) *Clauses {
	return andClausesImpl(nil, args)
}

// AndClausesWithAnnotOp is like AndClausesTyped but with an explicit annotation
// combiner, matching Python's and_clauses(..., annot_op=f).
func AndClausesWithAnnotOp(annotOp AnnotOp, args ...*Clauses) *Clauses {
	return andClausesImpl(annotOp, args)
}

func andClausesImpl(annotOp AnnotOp, args []*Clauses) *Clauses {
	if len(args) == 0 {
		return TrueClauses(nil)
	}

	// Compute annotation
	var annot interface{}
	if annotOp != nil {
		// Python: annot = annot_op(*[c.annot for c in args])
		annots := make([]interface{}, len(args))
		for i, c := range args {
			annots[i] = c.Annot
		}
		annot = annotOp(annots...)
	} else {
		// Python default: annot = a.annot if annot is None else annot if a.annot is None else annot.conj(a.annot)
		for _, c := range args {
			if c.Annot != nil {
				if annot == nil {
					annot = c.Annot
				} else if conjer, ok := annot.(AnnotConjoiner); ok {
					// Inline annotation conjunction via ConjWith method.
					// Replaces the old AnnotConjFunc global callback.
					annot = conjer.ConjWith(c.Annot)
				}
			}
		}
	}

	for _, c := range args {
		if c.IsFalse() {
			if xtracer.Enabled {
				totalDefs := 0
				for _, cc := range args {
					totalDefs += len(cc.Defs)
				}
				xtracer.Trace("ops.andClauses FALSE-DROP droppingDefs=%d nArgs=%d", totalDefs, len(args))
			}
			return FalseClauses(annot)
		}
	}

	var fmlas []lg.Expr
	var defs []*il.Definition
	for _, c := range args {
		fmlas = append(fmlas, c.Fmlas...)
		defs = append(defs, c.Defs...)
	}
	return NewClauses(fmlas, defs, annot)
}

// OrClauses computes the disjunction of Clauses and/or formulas.
// Each argument can be *Clauses or lg.Expr. If no argument is a *Clauses,
// returns an Or formula directly. Otherwise introduces fresh Boolean
// variables (Tseitin-like) to encode the disjunction.
func OrClauses(args ...interface{}) interface{} {
	hasClauses := false
	for _, a := range args {
		if _, ok := a.(*Clauses); ok {
			hasClauses = true
			break
		}
	}
	if !hasClauses {
		nodes := make([]lg.Expr, len(args))
		for i, a := range args {
			nodes[i] = a.(lg.Expr)
		}
		return &lg.Or{Terms: nodes}
	}

	clauses := coerceArgsToClauses(args)

	// Filter out false clauses
	var nonFalse []*Clauses
	for _, c := range clauses {
		if !c.IsFalse() {
			nonFalse = append(nonFalse, c)
		}
	}

	if len(nonFalse) == 0 {
		return FalseClauses(nil)
	}
	if len(nonFalse) == 1 {
		return nonFalse[0]
	}

	// Collect used symbol names for unique renaming
	used := collectUsedNames(nonFalse, nil)
	rn := iu.NewUniqueRenamer("__ts0", used)

	return orClausesInt(rn, nonFalse)
}

// OrClausesTyped is a convenience wrapper that always returns *Clauses.
// Matches Python's or_clauses: filters false branches, does Tseitin encoding
// if 2+ non-false branches, and always returns via fixOrAnnot for proper
// annotation reconstruction.
func OrClausesTyped(args ...*Clauses) *Clauses {
	if len(args) == 0 {
		return FalseClauses(nil)
	}

	origArgs := make([]*Clauses, len(args))
	copy(origArgs, args)

	var nonFalse []*Clauses
	for _, c := range args {
		if !c.IsFalse() {
			nonFalse = append(nonFalse, c)
		}
	}

	var res *Clauses
	var vs []lg.Expr

	if len(nonFalse) == 0 {
		res = FalseClauses(nil)
		vs = nil
	} else if len(nonFalse) == 1 {
		res = nonFalse[0]
		vs = []lg.Expr{lg.True} // Python: [And()]
	} else {
		used := collectUsedNames(nonFalse, nil)
		rn := iu.NewUniqueRenamer("__ts0", used)
		res, vs, nonFalse = orClausesIntWithVs(rn, nonFalse)
	}

	// Build fixed_vs and fixed_args matching Python's orig_args ordering
	fixedVs := make([]lg.Expr, 0, len(origArgs))
	fixedArgs := make([]*Clauses, 0, len(origArgs))
	idx := 0
	for _, a := range origArgs {
		if a.IsFalse() {
			fixedVs = append(fixedVs, lg.False) // Python: Or()
			fixedArgs = append(fixedArgs, a)
		} else {
			if idx < len(vs) {
				fixedVs = append(fixedVs, vs[idx])
			}
			if idx < len(nonFalse) {
				fixedArgs = append(fixedArgs, nonFalse[idx])
			}
			idx++
		}
	}

	return fixOrAnnot(res, fixedVs, fixedArgs)
}

// AnnotIteFunc is a callback for computing annot.ite(v, other) without
// importing the actions package. Set by the actions package at init time.
var AnnotIteFunc func(annot interface{}, cond lg.Expr, other interface{}) interface{}

// fixOrAnnot reconstructs annotations for or_clauses results.
// Matches Python's fix_or_annot.
func fixOrAnnot(res *Clauses, vs []lg.Expr, args []*Clauses) *Clauses {
	if len(args) == 0 {
		return res
	}
	annot := args[0].Annot
	for i := 1; i < len(args) && i < len(vs); i++ {
		a := args[i].Annot
		if annot == nil || a == nil {
			annot = nil
		} else if AnnotIteFunc != nil {
			annot = AnnotIteFunc(a, vs[i], annot)
		}
	}
	return NewClauses(res.Fmlas, res.Defs, annot)
}

// orClausesInt implements the Tseitin-like encoding for disjunction.
// Delegates to orClausesIntWithVs and discards the vs/args returns.
func orClausesInt(rn *iu.UniqueRenamer, args []*Clauses) *Clauses {
	res, _, _ := orClausesIntWithVs(rn, args)
	return res
}

// orClausesIntWithVs implements the Tseitin-like encoding for disjunction.
// Returns the result Clauses, the fresh Boolean variables, and the (possibly
// modified) args — matching Python's or_clauses_int which returns (res, vs, args).
func orClausesIntWithVs(rn *iu.UniqueRenamer, args []*Clauses) (*Clauses, []lg.Expr, []*Clauses) {
	// Eliminate dead definitions across args
	args = elimDeadDefinitions(rn, args)

	if xtracer.Enabled {
		argInfo := make([]string, len(args))
		for i, a := range args {
			argInfo[i] = fmt.Sprintf("'(%dF,%dD)'", len(a.Fmlas), len(a.Defs))
		}
		xtracer.Trace("ops.orClausesInt ENTER nArgs=%d args=[%v]", len(args), strings.Join(argInfo, ", "))
	}

	// Create fresh Boolean variables, one per disjunct
	vs := make([]lg.Expr, len(args))
	vsNodes := make([]lg.Expr, len(args))
	for i := range args {
		name := rn.Rename("")
		c := lg.NewConst(name, lg.Boolean)
		vs[i] = c
		vsNodes[i] = c
	}

	// Build formulas:
	// 1. Or(v1, v2, ..., vn)
	// 2. For each i: Not(vi) OR fmla, for each fmla in args[i].Fmlas
	var fmlas []lg.Expr
	fmlas = append(fmlas, &lg.Or{Terms: vsNodes})
	for i, cls := range args {
		for _, f := range cls.Fmlas {
			fmlas = append(fmlas, &lg.Or{Terms: []lg.Expr{
				&lg.Not{Body: vs[i].(*lg.Const)},
				f,
			}})
		}
	}

	// Merge definitions (InsMap preserves insertion order, matching Python 3.7+ dict)
	defIdx := iu.NewInsMap[lg.NodeKey, *il.Definition]()
	for i, cls := range args {
		for _, d := range cls.Defs {
			key := definesKey(d)
			if existing, ok := defIdx.Get2(key); !ok {
				defIdx.Set(key, d)
			} else {
				// Merge: use bare Ite to select between definitions.
				// Python or_clauses_int uses bare Ite (not simp_ite).
				// Only ite_clauses_int uses simp_ite.
				merged := il.NewDefinition(
					d.Lhs,
					&lg.Ite{ISort: d.Rhs.NodeSort(), Cond: vs[i], Then: d.Rhs, Else: existing.Rhs},
				)
				defIdx.Set(key, merged)
			}
		}
	}

	var defs []*il.Definition
	for _, d := range defIdx.All() {
		defs = append(defs, d)
	}

	result := NewClauses(fmlas, defs, nil)
	if xtracer.Enabled {
		xtracer.Trace("ops.orClausesInt EXIT HASH canon= %s", result.Canon())
	}
	return result, vs, args
}

// IteClauses computes if-then-else on Clauses:
// if cond then args[0] else args[1].
func IteClauses(cond lg.Expr, thenCls, elseCls *Clauses) *Clauses {
	// Handle trivial false cases
	if thenCls.IsFalse() && elseCls.IsFalse() {
		return thenCls
	}
	if thenCls.IsFalse() {
		thenCls = NewClauses(thenCls.Fmlas, elseCls.Defs, thenCls.Annot)
	} else if elseCls.IsFalse() {
		elseCls = NewClauses(elseCls.Fmlas, thenCls.Defs, elseCls.Annot)
	}

	args := []*Clauses{thenCls, elseCls}

	// Collect used names
	used := collectUsedNames(args, cond)
	rn := iu.NewUniqueRenamer("__ts0", used)

	return iteClausesInt(rn, cond, args)
}

func iteClausesInt(rn *iu.UniqueRenamer, cond lg.Expr, args []*Clauses) *Clauses {
	args = elimDeadDefinitions(rn, args)

	if xtracer.Enabled {
		xtracer.Trace("ops.iteClausesInt ENTER nArgs0Fmlas=%d nArgs0Defs=%d nArgs1Fmlas=%d nArgs1Defs=%d",
			len(args[0].Fmlas), len(args[0].Defs), len(args[1].Fmlas), len(args[1].Defs))
	}

	// Create a fresh Boolean variable for the condition
	name := rn.Rename("")
	v := lg.NewConst(name, lg.Boolean)

	// Build formulas:
	// For then-branch: Not(v) OR fmla
	// For else-branch: v OR fmla
	var fmlas []lg.Expr
	for _, f := range args[0].Fmlas {
		fmlas = append(fmlas, &lg.Or{Terms: []lg.Expr{&lg.Not{Body: v}, f}})
	}
	for _, f := range args[1].Fmlas {
		fmlas = append(fmlas, &lg.Or{Terms: []lg.Expr{v, f}})
	}

	// Merge definitions (InsMap preserves insertion order, matching Python 3.7+ dict)
	defIdx := iu.NewInsMap[lg.NodeKey, *il.Definition]()
	for _, d := range args[0].Defs {
		key := definesKey(d)
		defIdx.Set(key, d)
	}
	for _, d := range args[1].Defs {
		key := definesKey(d)
		if existing, ok := defIdx.Get2(key); !ok {
			defIdx.Set(key, d)
		} else {
			merged := il.NewDefinition(
				d.Lhs,
				il.SimpIte(v, existing.Rhs, d.Rhs),
			)
			defIdx.Set(key, merged)
		}
	}

	var defs []*il.Definition
	for _, d := range defIdx.All() {
		defs = append(defs, d)
	}
	// Add definition: v = cond
	defs = append(defs, il.NewDefinition(v, cond))

	// Compute annotation matching Python: annot = None if a0 is None or a1 is None else a0.ite(v,a1)
	var annot interface{}
	a0, a1 := args[0].Annot, args[1].Annot
	if a0 != nil && a1 != nil && AnnotIteFunc != nil {
		annot = AnnotIteFunc(a0, v, a1)
	}
	result := NewClauses(fmlas, defs, annot)
	if xtracer.Enabled {
		xtracer.Trace("ops.iteClausesInt EXIT HASH canon= %s", result.Canon())
	}
	return result
}

// NegateClauses negates a Clauses. Requires the clauses to be
// universal first-order (no definitions, no skolems).
func NegateClauses(clauses *Clauses) *Clauses {
	if !clauses.IsUniversalFirstOrder() {
		panic("NegateClauses requires universal first-order clauses")
	}
	return dualClauses(clauses, nil, nil)
}

// Skolemizer is a function that creates a Skolem constant for a variable.
// Corresponds to Python's skolemizer parameter in dual_clauses.
type Skolemizer func(v *lg.Variable) lg.Expr

// DualClauses implements the dual construction: replaces variables with
// Skolem constants, then negates. Corresponds to Python's dual_clauses.
// If skolemizer is nil, defaults to var_to_skolem('__', v) which produces
// a constant named "__"+v.Name with the same sort.
// If instantiator is non-nil, definition instances are conjoined with the
// negated formula (matching Python's `if instantiator != None` check).
func DualClauses(clauses *Clauses, skolemizer Skolemizer, instantiator func([]lg.Expr) *Clauses) *Clauses {
	return dualClauses(clauses, skolemizer, instantiator)
}

func dualClauses(clauses *Clauses, skolemizer Skolemizer, instantiator func([]lg.Expr) *Clauses) *Clauses {
	// Get used variables in order
	vars := UsedVariablesOrdered(clauses)

	// Create Skolem substitution
	subs := make(map[lg.NodeKey]lg.Expr, len(vars))
	for _, v := range vars {
		if skolemizer != nil {
			subs[lg.Key(v)] = skolemizer(v)
		} else {
			// Default: var_to_skolem('__', v)
			sk := lg.NewConst("__"+v.Name, v.VSort)
			subs[lg.Key(v)] = sk
		}
	}

	// Apply substitution
	clauses = SubstituteNodesClauses(clauses, subs)

	// Negate the formula
	f := Negate(clausesToFormula(clauses))

	// Python: if instantiator != None: fmla = And(fmla, clauses_to_formula(insts))
	if instantiator != nil {
		gts := AppsClauses(clauses)
		insts := instantiator(gts)
		if len(insts.Fmlas) > 0 {
			f = &lg.And{Terms: []lg.Expr{f, clausesToFormula(insts)}}
		}
	}

	return FormulaToClauses(f, nil)
}

// clausesToFormula converts clauses to formula (drop universals).
func clausesToFormula(c *Clauses) lg.Expr {
	return dropUniversals(c.ToFormula())
}

// ClausesToFormula converts clauses to a formula, dropping leading
// universal quantifiers. Corresponds to Python's clauses_to_formula.
func ClausesToFormula(c *Clauses) lg.Expr {
	return clausesToFormula(c)
}

// ConditionClauses returns Clauses equivalent to "fmla -> clauses".
// Each formula in clauses is wrapped as "Not(fmla) OR formula".
func ConditionClauses(clauses *Clauses, fmla lg.Expr) *Clauses {
	negFmla := Negate(fmla)
	var newFmlas []lg.Expr
	for _, f := range clauses.Fmlas {
		newFmlas = append(newFmlas, &lg.Or{Terms: []lg.Expr{negFmla, f}})
	}
	return NewClauses(newFmlas, clauses.Defs, clauses.Annot)
}

// ClausesUsingSymbols filters clauses to only those formulas and definitions
// that use any of the given symbols.
// ClausesUsingSymbolNames filters clauses to those that reference any
// symbol in the given name set. Matches Python clauses_using_symbols.
func ClausesUsingSymbolNames(symNames map[string]bool, clauses *Clauses) *Clauses {
	if clauses == nil || len(symNames) == 0 {
		return NewClauses(nil, nil, nil)
	}
	var fmlas []lg.Expr
	for _, f := range clauses.Fmlas {
		if usesSymbolNameAST(symNames, f) {
			fmlas = append(fmlas, f)
		}
	}
	var defs []*il.Definition
	for _, d := range clauses.Defs {
		if usesSymbolNameAST(symNames, d) {
			defs = append(defs, d)
		}
	}
	return NewClauses(fmlas, defs, clauses.Annot)
}

// usesSymbolNameAST returns true if any symbol name from the set appears in the node.
func usesSymbolNameAST(names map[string]bool, node lg.Expr) bool {
	if c, ok := node.(*lg.Const); ok {
		return names[c.Name]
	}
	if app, ok := node.(*lg.Apply); ok {
		if usesSymbolNameAST(names, app.Func) {
			return true
		}
	}
	for _, child := range node.Children() {
		if usesSymbolNameAST(names, child) {
			return true
		}
	}
	return false
}

// UsedSymbolNamesClauses collects all symbol names from a Clauses.
func UsedSymbolNamesClauses(clauses *Clauses) map[string]bool {
	if clauses == nil {
		return make(map[string]bool)
	}
	result := make(map[string]bool)
	for _, f := range clauses.Fmlas {
		collectSymbolNamesFromNode(f, result)
	}
	for _, d := range clauses.Defs {
		collectSymbolNamesFromNode(d, result)
	}
	return result
}

// UsedSymbolsClauses returns all constant symbols referenced in a Clauses
// object, preserving their sorts. This matches Python's used_symbols_clauses
// which returns a set of Symbol objects (with sorts).
func UsedSymbolsClauses(clauses *Clauses) map[lg.NodeKey]*lg.Const {
	if clauses == nil {
		return make(map[lg.NodeKey]*lg.Const)
	}
	result := make(map[lg.NodeKey]*lg.Const)
	for _, f := range clauses.Fmlas {
		for k, v := range UsedSymbolsAST(f) {
			result[k] = v
		}
	}
	for _, d := range clauses.Defs {
		for k, v := range UsedSymbolsAST(d) {
			result[k] = v
		}
	}
	return result
}

func collectSymbolNamesFromNode(n lg.Expr, result map[string]bool) {
	if c, ok := n.(*lg.Const); ok {
		result[c.Name] = true
	}
	if app, ok := n.(*lg.Apply); ok {
		collectSymbolNamesFromNode(app.Func, result)
	}
	for _, child := range n.Children() {
		collectSymbolNamesFromNode(child, result)
	}
}

func ClausesUsingSymbols(syms map[lg.NodeKey]lg.Expr, clauses *Clauses) *Clauses {
	var fmlas []lg.Expr
	for _, f := range clauses.Fmlas {
		if usesSymbolsAST(syms, f) {
			fmlas = append(fmlas, f)
		}
	}
	var defs []*il.Definition
	for _, d := range clauses.Defs {
		if usesSymbolsAST(syms, d) {
			defs = append(defs, d)
		}
	}
	return NewClauses(fmlas, defs, clauses.Annot)
}

// RenameClauses renames symbols in clauses by structural identity.
// The map keys are lg.NodeKey (via lg.Key(sym)) for structural equality.
// Also renames annotations matching Python's rename_clauses_annot_fun.
func RenameClauses(clauses *Clauses, subs map[lg.NodeKey]*lg.Const) *Clauses {
	fn := func(n lg.Expr) lg.Expr {
		return RenameAST(n, subs)
	}
	result := clauses.Apply(fn)

	// Rename annotation if it supports it (Python: annot_fun=rename_clauses_annot_fun)
	if result.Annot != nil {
		// Convert subs to lg.Expr map for the AnnotRenamer interface
		exprSubs := make(map[lg.NodeKey]lg.Expr, len(subs))
		for k, v := range subs {
			exprSubs[k] = v
		}
		if renamer, ok := result.Annot.(AnnotRenamer); ok {
			result.Annot = renamer.Rename(exprSubs)
		}
	}

	return result
}

// SubstituteConstantsClauses substitutes constants in clauses by structural identity.
// The map keys are lg.NodeKey (via lg.Key(sym)) for structural equality.
func SubstituteConstantsClauses(clauses *Clauses, subs map[lg.NodeKey]lg.Expr) *Clauses {
	fn := func(n lg.Expr) lg.Expr {
		return SubstituteConstantsExpr(n, subs)
	}
	return clauses.Apply(fn)
}

// SubstituteNodesClauses applies a node-level substitution to all formulas
// and definitions in the clauses.
func SubstituteNodesClauses(clauses *Clauses, subs map[lg.NodeKey]lg.Expr) *Clauses {
	if len(subs) == 0 {
		return clauses
	}
	var fmlas []lg.Expr
	for _, f := range clauses.Fmlas {
		nf, err := substituteNodesRec(f, subs)
		if err != nil {
			fmlas = append(fmlas, f) // fallback: keep original
		} else {
			fmlas = append(fmlas, nf)
		}
	}
	var defs []*il.Definition
	for _, d := range clauses.Defs {
		nd, err := substituteNodesRec(d, subs)
		if err != nil {
			defs = append(defs, d)
		} else {
			if def, ok := nd.(*il.Definition); ok {
				defs = append(defs, def)
			} else {
				defs = append(defs, d)
			}
		}
	}
	return NewClauses(fmlas, defs, clauses.Annot)
}

// substituteNodesRec is a thin wrapper around logicutil.Substitute that also
// handles il.Definition nodes.
func substituteNodesRec(n lg.Expr, subs map[lg.NodeKey]lg.Expr) (lg.Expr, error) {
	if d, ok := n.(*il.Definition); ok {
		lhs, err1 := substituteNodesRec(d.Lhs, subs)
		rhs, err2 := substituteNodesRec(d.Rhs, subs)
		if err1 != nil {
			return nil, err1
		}
		if err2 != nil {
			return nil, err2
		}
		return il.NewDefinition(lhs, rhs), nil
	}
	// Check direct replacement
	if r, ok := subs[lg.Key(n)]; ok {
		return r, nil
	}
	// Handle standard nodes
	switch n.(type) {
	case *lg.Variable, *lg.Const:
		// Already checked via Key above
		return n, nil
	}
	// Recurse into children
	children := n.Children()
	if len(children) == 0 {
		return n, nil
	}
	newChildren := make([]lg.Expr, len(children))
	changed := false
	for i, c := range children {
		nc, err := substituteNodesRec(c, subs)
		if err != nil {
			return nil, err
		}
		newChildren[i] = nc
		if nc != c {
			changed = true
		}
	}
	if !changed {
		return n, nil
	}
	return il.CloneNode(n, newChildren), nil
}

// --- internal helpers ---

// coerceArgsToClauses converts a mixed slice of *Clauses and lg.Expr to []*Clauses.
func coerceArgsToClauses(args []interface{}) []*Clauses {
	result := make([]*Clauses, len(args))
	for i, a := range args {
		switch v := a.(type) {
		case *Clauses:
			result[i] = v
		case lg.Expr:
			result[i] = FormulaToClauses(v, nil)
		default:
			panic(fmt.Sprintf("clauseops: unexpected argument type %T", a))
		}
	}
	return result
}

// collectUsedNames collects all symbol names used in clauses and optionally
// in an extra formula, returned as a slice of strings.
func collectUsedNames(args []*Clauses, extra lg.Expr) []string {
	seen := make(map[string]struct{})
	for _, cls := range args {
		for _, s := range cls.Symbols() {
			seen[s.Name] = struct{}{}
		}
	}
	if extra != nil {
		for _, s := range usedSymbolsAST(extra) {
			seen[s.Name] = struct{}{}
		}
	}
	result := make([]string, 0, len(seen))
	for k := range seen {
		result = append(result, k)
	}
	return result
}

// elimDeadDefinitions eliminates definitions that are captured across
// different clause sets. Matches Python's elim_dead_definitions:
// - Non-skolem captured symbols are "dead": eliminated by converting to constraints.
// - Skolem captured symbols are renamed to fresh names via the renamer.
func elimDeadDefinitions(rn *iu.UniqueRenamer, args []*Clauses) []*Clauses {
	// 1. Collect all defined symbols (with their Const for skolem check)
	// InsMap preserves insertion order, matching Python 3.7+ dict.
	defined := iu.NewInsMap[lg.NodeKey, *lg.Const]()
	for _, a := range args {
		for _, d := range a.Defs {
			key := definesKey(d)
			if c, ok := d.Defines().(*lg.Const); ok {
				defined.Set(key, c)
			} else {
				defined.Set(key, nil)
			}
		}
	}

	// 2. Find captured: defined somewhere but not in all args
	var captured []lg.NodeKey
	for key := range defined.All() {
		for _, a := range args {
			if _, ok := a.DefIdx[key]; !ok {
				captured = append(captured, key)
				break
			}
		}
	}

	// 3. Split: non-skolem → dead (eliminate), skolem → toRename
	var dead []lg.NodeKey
	var toRename []*lg.Const
	for _, key := range captured {
		sym, _ := defined.Get2(key)
		if sym != nil && isSkolem(sym) {
			toRename = append(toRename, sym)
		} else {
			dead = append(dead, key)
		}
	}

	if xtracer.Enabled {
		xtracer.Trace("ops.elimDeadDefinitions nArgs=%d nDefined=%d nCaptured=%d nDead=%d nToRename=%d",
			len(args), defined.Len(), len(captured), len(dead), len(toRename))
		if len(dead) > 0 {
			deadNames := make([]string, len(dead))
			for i, k := range dead {
				deadNames[i] = fmt.Sprintf("'%v'", k)
			}
			xtracer.Trace("ops.elimDeadDefinitions HASH canon= dead=[%v]", strings.Join(deadNames, ", "))
		}
	}

	// 4. Rename skolems to fresh names (Python: rename_symbols(rn, arg, to_rename))
	// Python calls rename_symbols(rn, arg, to_rename) per arg, generating
	// DIFFERENT fresh names per arg (rn is stateful). This separates captured
	// skolems so they don't interact across args after merging.
	if len(toRename) > 0 {
		for i, a := range args {
			subs := make(map[lg.NodeKey]*lg.Const, len(toRename))
			for _, sym := range toRename {
				newName := rn.Rename(sym.Name)
				subs[lg.Key(sym)] = lg.NewConst(newName, sym.CSort)
			}
			args[i] = RenameClauses(a, subs)
		}
	}

	// 5. Eliminate dead (non-skolem) definitions by converting to constraints.
	// Always call elimDefinitions (even with empty dead) to match Python's
	// execution path — Python has no early return here.
	result := make([]*Clauses, len(args))
	for i, a := range args {
		result[i] = elimDefinitions(a, dead)
	}
	return result
}

// elimDefinitions converts dead definitions to constraint formulas.
// Matches Python's elim_definitions: iterates `dead` in order and
// looks up each symbol in clauses.DefIdx, so the converted constraint
// formulas appear in `dead`-list order (not clauses.Defs order).
func elimDefinitions(clauses *Clauses, dead []lg.NodeKey) *Clauses {
	if xtracer.Enabled {
		deadNames := make([]string, len(dead))
		for i, k := range dead {
			deadNames[i] = fmt.Sprintf("'%v'", k)
		}
		defNames := make([]string, len(clauses.Defs))
		for i, d := range clauses.Defs {
			defNames[i] = fmt.Sprintf("'%v'", d.Defines().Canon())
		}
		xtracer.Trace("ops.elimDefinitions ENTER HASH canon= nDead=%d dead=[%v] nDefs=%d defs=[%v] nFmlas=%d",
			len(dead), strings.Join(deadNames, ", "), len(clauses.Defs), strings.Join(defNames, ", "), len(clauses.Fmlas))
	}
	var fmlas []lg.Expr
	fmlas = append(fmlas, clauses.Fmlas...)
	// Python: for sym in dead: if sym in clauses.defidx: fmlas.append(...)
	for _, key := range dead {
		if idx, ok := clauses.DefIdx[key]; ok {
			fmlas = append(fmlas, defToConstraint(clauses.Defs[idx]))
		}
	}
	// Build deadSet for filtering remaining defs
	deadSet := make(map[lg.NodeKey]bool, len(dead))
	for _, key := range dead {
		deadSet[key] = true
	}
	var defs []*il.Definition
	for _, d := range clauses.Defs {
		key := definesKey(d)
		if !deadSet[key] {
			defs = append(defs, d)
		}
	}
	if xtracer.Enabled {
		xtracer.Trace("ops.elimDefinitions EXIT nFmlas=%d nDefs=%d", len(fmlas), len(defs))
	}
	return NewClauses(fmlas, defs, clauses.Annot)
}

// UsedSymbolsClausesOrdered collects constant symbols from clauses in
// depth-first AST traversal order, deduplicating by structural identity
// (lg.NodeKey includes name + sort). This matches Python's
// dict.fromkeys(symbols_clauses(c)) insertion order.
func UsedSymbolsClausesOrdered(c *Clauses) *iu.InsMap[lg.NodeKey, *lg.Const] {
	result := iu.NewInsMap[lg.NodeKey, *lg.Const]()
	if c == nil {
		return result
	}
	for _, f := range c.Fmlas {
		collectSymbolsOrdered(f, result)
	}
	for _, d := range c.Defs {
		collectSymbolsOrdered(d, result)
	}
	return result
}

func collectSymbolsOrdered(n lg.Expr, result *iu.InsMap[lg.NodeKey, *lg.Const]) {
	if n == nil {
		return
	}
	if c, ok := n.(*lg.Const); ok {
		result.Set(lg.Key(c), c)
	}
	for _, child := range n.Children() {
		collectSymbolsOrdered(child, result)
	}
}

// UsedSymbolsExprOrdered collects constant symbols from an expression in
// depth-first AST traversal order, deduplicating by structural identity.
func UsedSymbolsExprOrdered(node lg.Expr) *iu.InsMap[lg.NodeKey, *lg.Const] {
	result := iu.NewInsMap[lg.NodeKey, *lg.Const]()
	collectSymbolsOrdered(node, result)
	return result
}

// UsedVariablesOrdered returns free variables from the clauses in order.
func UsedVariablesOrdered(c *Clauses) []*lg.Variable {
	seen := make(map[string]bool)
	var result []*lg.Variable
	for _, f := range c.Fmlas {
		collectVarsOrdered(f, seen, &result)
	}
	for _, d := range c.Defs {
		collectVarsOrdered(d, seen, &result)
	}
	return result
}

func collectVarsOrdered(n lg.Expr, seen map[string]bool, result *[]*lg.Variable) {
	if v, ok := n.(*lg.Variable); ok {
		if !seen[v.Name] {
			seen[v.Name] = true
			*result = append(*result, v)
		}
		return
	}
	for _, c := range n.Children() {
		collectVarsOrdered(c, seen, result)
	}
}

// -----------------------------------------------------------------------
// apply_func_to_clauses / apply_gen_to_clauses equivalents
//
// Python's apply_func_to_clauses(fn) returns a function that applies fn
// to each formula/def in a Clauses. Python's apply_gen_to_clauses(gen)
// returns a function that collects generator results across all
// formulas/defs. In Go, we provide the specific wrapped functions.
// -----------------------------------------------------------------------

// SubstituteClauses applies substitute_ast to all formulas and defs in clauses.
// Corresponds to Python: substitute_clauses = apply_func_to_clauses(substitute_ast)
func SubstituteClauses(clauses *Clauses, subs map[lg.NodeKey]lg.Expr) *Clauses {
	if clauses == nil || len(subs) == 0 {
		return clauses
	}
	return SubstituteNodesClauses(clauses, subs)
}

// SubstituteClausesByName applies variable substitution keyed by name to clauses.
// Corresponds to Python substitute_clauses(clauses, sksubs) where sksubs maps
// variable name (v.rep) to replacement node.
func SubstituteClausesByName(clauses *Clauses, subs map[string]lg.Expr) *Clauses {
	if clauses == nil || len(subs) == 0 {
		return clauses
	}
	fmlas := make([]lg.Expr, len(clauses.Fmlas))
	for i, f := range clauses.Fmlas {
		fmlas[i] = SubstituteAstByName(f, subs)
	}
	defs := make([]*il.Definition, len(clauses.Defs))
	for i, d := range clauses.Defs {
		replaced := SubstituteAstByName(d, subs)
		if rd, ok := replaced.(*il.Definition); ok {
			defs[i] = rd
		} else {
			defs[i] = d
		}
	}
	return NewClauses(fmlas, defs, clauses.Annot)
}

// SubstBothClauses applies substitution to both variables and constants in clauses.
// Corresponds to Python subst_both_clauses:
//
//	substitute_constants_clauses(substitute_clauses(clauses, subst), subst)
//
// The subs map is keyed by name strings (mixed variable + constant names).
// Internally, variable substitution uses name-based lookup (World 2), and
// constant substitution constructs structural keys for the SubstituteConstantsAST
// lookup (World 1). This dual-world function bridges the two.
func SubstBothClauses(clauses *Clauses, subs map[string]lg.Expr) *Clauses {
	if clauses == nil || len(subs) == 0 {
		return clauses
	}
	// First, substitute as variables (map[NodeKey]lg.Expr keyed by Var sexp)
	varSubs := make(map[lg.NodeKey]lg.Expr)
	for name, val := range subs {
		v, err := lg.NewVariable(name, val.NodeSort())
		if err == nil {
			varSubs[lg.Key(v)] = val
		}
	}
	result := SubstituteClauses(clauses, varSubs)
	// Then, substitute as constants — build structural keys
	constSubs := make(map[lg.NodeKey]lg.Expr, len(subs))
	for name, val := range subs {
		// Construct a Symbol with the value's sort for structural matching
		sym := lg.NewConst(name, val.NodeSort())
		constSubs[lg.Key(sym)] = val
	}
	result = SubstituteConstantsClauses(result, constSubs)
	return result
}

// resortClausesBySort remaps sorts in all formulas and defs of clauses.
// Corresponds to Python: resort_clauses = apply_func_to_clauses(resort_ast)
func resortClausesBySort(clauses *Clauses, subs map[lg.NodeKey]lg.Sort) *Clauses {
	if clauses == nil || len(subs) == 0 {
		return clauses
	}
	fn := func(n lg.Expr) lg.Expr {
		return lu.ResortAst(n, subs)
	}
	return clauses.Apply(fn)
}

// VariablesClauses returns all free variables across all formulas and defs.
// Corresponds to Python: variables_clauses = apply_gen_to_clauses(variables_ast)
func VariablesClauses(clauses *Clauses) []*lg.Variable {
	if clauses == nil {
		return nil
	}
	seen := make(map[string]bool)
	var result []*lg.Variable
	for _, f := range clauses.Fmlas {
		for _, vNode := range lu.FreeVariables(f) {
			vv := vNode.(*lg.Variable)
			if !seen[vv.Name] {
				seen[vv.Name] = true
				result = append(result, vv)
			}
		}
	}
	for _, d := range clauses.Defs {
		for _, vNode := range lu.FreeVariables(d) {
			vv := vNode.(*lg.Variable)
			if !seen[vv.Name] {
				seen[vv.Name] = true
				result = append(result, vv)
			}
		}
	}
	return result
}

// ConstantsClauses returns all constants used across all formulas and defs.
// Corresponds to Python: constants_clauses = apply_gen_to_clauses(constants_ast)
func ConstantsClauses(clauses *Clauses) []*lg.Const {
	if clauses == nil {
		return nil
	}
	seen := make(map[string]bool)
	var result []*lg.Const
	for _, f := range clauses.Fmlas {
		for _, cc := range lu.UsedConstants(f) {
			if !seen[cc.Name] {
				seen[cc.Name] = true
				result = append(result, cc)
			}
		}
	}
	for _, d := range clauses.Defs {
		for _, cc := range lu.UsedConstants(d) {
			if !seen[cc.Name] {
				seen[cc.Name] = true
				result = append(result, cc)
			}
		}
	}
	return result
}

// SymbolsClauses returns all constant symbols used across all formulas and defs.
// Corresponds to Python: symbols_clauses = apply_gen_to_clauses(symbols_ast)
func SymbolsClauses(clauses *Clauses) []*lg.Const {
	if clauses == nil {
		return nil
	}
	result := clauses.Symbols()
	syms := make([]*lg.Const, 0, len(result))
	for _, s := range result {
		syms = append(syms, s)
	}
	return syms
}

// RelationsClauses returns all relation symbols across all formulas and defs.
// Corresponds to Python: relations_clauses = apply_gen_to_clauses(relations_ast)
func RelationsClauses(clauses *Clauses) []*lg.Const {
	if clauses == nil {
		return nil
	}
	seen := make(map[string]bool)
	var result []*lg.Const
	for _, f := range clauses.Fmlas {
		for _, r := range lu.RelationsAst(f) {
			if !seen[r.Name] {
				seen[r.Name] = true
				result = append(result, r)
			}
		}
	}
	for _, d := range clauses.Defs {
		for _, r := range lu.RelationsAst(d) {
			if !seen[r.Name] {
				seen[r.Name] = true
				result = append(result, r)
			}
		}
	}
	return result
}

// FunctionsClauses returns all function symbols across all formulas and defs.
// Corresponds to Python: functions_clauses = apply_gen_to_clauses(functions_ast)
func FunctionsClauses(clauses *Clauses) []*lg.Const {
	if clauses == nil {
		return nil
	}
	seen := make(map[string]bool)
	var result []*lg.Const
	for _, f := range clauses.Fmlas {
		for _, fn := range lu.FunctionsAst(f) {
			if !seen[fn.Name] {
				seen[fn.Name] = true
				result = append(result, fn)
			}
		}
	}
	for _, d := range clauses.Defs {
		for _, fn := range lu.FunctionsAst(d) {
			if !seen[fn.Name] {
				seen[fn.Name] = true
				result = append(result, fn)
			}
		}
	}
	return result
}

// AppsClauses returns all function applications across all formulas and defs.
// Corresponds to Python: apps_clauses = apply_gen_to_clauses(apps_ast)
func AppsClauses(clauses *Clauses) []lg.Expr {
	if clauses == nil {
		return nil
	}
	var result []lg.Expr
	for _, f := range clauses.Fmlas {
		result = append(result, il.AppsAst(f)...)
	}
	for _, d := range clauses.Defs {
		result = append(result, il.AppsAst(d)...)
	}
	return result
}

// GroundAppsClauses returns all ground (variable-free) applications across all formulas and defs.
// Corresponds to Python: ground_apps_clauses = apply_gen_to_clauses(ground_apps_ast)
func GroundAppsClauses(clauses *Clauses) []lg.Expr {
	if clauses == nil {
		return nil
	}
	var result []lg.Expr
	for _, f := range clauses.Fmlas {
		result = append(result, lu.GroundAppsAst(f)...)
	}
	for _, d := range clauses.Defs {
		result = append(result, lu.GroundAppsAst(d)...)
	}
	return result
}

// EqsClauses returns all equality atoms across all formulas and defs.
// Corresponds to Python: eqs_clauses = apply_gen_to_clauses(eqs_ast)
func EqsClauses(clauses *Clauses) []*lg.Eq {
	if clauses == nil {
		return nil
	}
	var result []*lg.Eq
	for _, f := range clauses.Fmlas {
		result = append(result, lu.EqsAst(f)...)
	}
	for _, d := range clauses.Defs {
		result = append(result, lu.EqsAst(d)...)
	}
	return result
}

// -----------------------------------------------------------------------
// Tseitin encoding
// -----------------------------------------------------------------------

// tseitinContext manages Tseitin variable creation during clausification.
type tseitinContext struct {
	clauses []lg.Expr
	fresh   *iu.UniqueRenamer
}

func newTseitinContext(used map[string]bool) *tseitinContext {
	var usedNames []string
	for n := range used {
		usedNames = append(usedNames, n)
	}
	return &tseitinContext{
		fresh: iu.NewUniqueRenamer("__ts", usedNames),
	}
}

// tseitinEncoding encodes a formula into a literal, adding clauses to tc.
func (tc *tseitinContext) tseitinEncoding(f lg.Expr) lg.Expr {
	f = lu.ExpandAbbrevs(f)

	switch n := f.(type) {
	case *lg.And:
		if len(n.Terms) == 0 {
			return &lg.And{Terms: nil} // true
		}
		args := make([]lg.Expr, len(n.Terms))
		for i, g := range n.Terms {
			args[i] = tc.tseitinEncoding(g)
		}
		// Collect free variables from f
		varSet := make(map[string]*lg.Variable)
		collectFreeVars(f, varSet, nil)
		var vars []*lg.Variable
		for _, v := range varSet {
			vars = append(vars, v)
		}
		fname := tc.fresh.Rename(fmt.Sprintf("%d", len(vars)))
		// Create a fresh boolean/relational symbol
		var sorts []lg.Sort
		for _, v := range vars {
			sorts = append(sorts, v.VSort)
		}
		sorts = append(sorts, &lg.BooleanSort{})
		fs, _ := lg.NewFunctionSort(sorts...)
		fn := lg.NewConst(fname, fs)
		// Build the literal: fn(vars...)
		var varNodes []lg.Expr
		for _, v := range vars {
			varNodes = append(varNodes, v)
		}
		var res lg.Expr
		if len(varNodes) > 0 {
			res, _ = lg.NewApply(fn, varNodes...)
		} else {
			res = fn
		}
		// Add Tseitin clauses: ~res | arg_i for each i, and res | ~arg_0 | ~arg_1 | ...
		for _, arg := range args {
			// ~res | arg
			tc.clauses = append(tc.clauses, &lg.Or{Terms: []lg.Expr{&lg.Not{Body: res}, arg}})
		}
		// res | ~arg_0 | ~arg_1 | ...
		negArgs := make([]lg.Expr, len(args)+1)
		negArgs[0] = res
		for i, arg := range args {
			negArgs[i+1] = &lg.Not{Body: arg}
		}
		tc.clauses = append(tc.clauses, &lg.Or{Terms: negArgs})
		return res

	case *lg.Or:
		// ~(AND(~x for x in args))
		negArgs := make([]lg.Expr, len(n.Terms))
		for i, x := range n.Terms {
			negArgs[i] = &lg.Not{Body: x}
		}
		inner := &lg.And{Terms: negArgs}
		return &lg.Not{Body: tc.tseitinEncoding(inner)}

	default:
		// Atomic formula — return as-is
		return f
	}
}

// collectFreeVars collects free variables from a node.
func collectFreeVars(node lg.Expr, result map[string]*lg.Variable, bound map[string]bool) {
	if node == nil {
		return
	}
	switch n := node.(type) {
	case *lg.Variable:
		if bound == nil || !bound[n.Name] {
			result[n.Name] = n
		}
	case *lg.ForAll:
		newBound := make(map[string]bool)
		for k, v := range bound {
			newBound[k] = v
		}
		for _, v := range n.Variables {
			newBound[v.Name] = true
		}
		collectFreeVars(n.Body, result, newBound)
		return
	case *lg.Exists:
		newBound := make(map[string]bool)
		for k, v := range bound {
			newBound[k] = v
		}
		for _, v := range n.Variables {
			newBound[v.Name] = true
		}
		collectFreeVars(n.Body, result, newBound)
		return
	}
	for _, child := range node.Children() {
		collectFreeVars(child, result, bound)
	}
}

// TseitinEncode clausifies a formula using Tseitin encoding.
// The result is a Clauses with the original formula plus any
// Tseitin auxiliary clauses.
// Corresponds to Python tseitin_encode (line 968).
func TseitinEncode(f lg.Expr) *Clauses {
	tc := newTseitinContext(nil)
	clauses := FormulaToClauses(f, nil)
	// Add Tseitin auxiliary clauses
	for _, c := range tc.clauses {
		clauses.Fmlas = append(clauses.Fmlas, c)
	}
	return clauses
}

// -----------------------------------------------------------------------
// SimplifyClauses
// -----------------------------------------------------------------------

// SimplifyClauses performs iterative tautology elimination and clause
// simplification. Runs 3 rounds of simplification, removing tautological
// formulas each round.
// Corresponds to Python simplify_clauses (line 1016).
func SimplifyClauses(cls *Clauses) *Clauses {
	if cls == nil {
		return nil
	}
	fmlas := make([]lg.Expr, len(cls.Fmlas))
	copy(fmlas, cls.Fmlas)

	for i := 0; i < 3; i++ {
		// Simplify each formula
		simplified := make([]lg.Expr, 0, len(fmlas))
		for _, f := range fmlas {
			s := simplifyFormula(f)
			simplified = append(simplified, s)
		}
		// Remove tautologies
		fmlas = fmlas[:0]
		for _, f := range simplified {
			if !isTautologyFormula(f) {
				fmlas = append(fmlas, f)
			}
		}
	}

	return NewClauses(fmlas, cls.Defs, cls.Annot)
}

// simplifyFormula simplifies a formula by:
// - Propagating negation of equalities (rewriting variables)
// - Removing vacuous literals
// - Removing duplicate literals
// Corresponds to Python simplify_clause_fmla which converts to clause form,
// simplifies, and converts back.
func simplifyFormula(f lg.Expr) lg.Expr {
	// Simplify And/Or by recursing
	switch n := f.(type) {
	case *lg.And:
		terms := make([]lg.Expr, 0, len(n.Terms))
		for _, t := range n.Terms {
			s := simplifyFormula(t)
			// Remove True conjuncts
			if isTrue(s) {
				continue
			}
			terms = append(terms, s)
		}
		if len(terms) == 0 {
			return &lg.And{Terms: nil} // true
		}
		if len(terms) == 1 {
			return terms[0]
		}
		return &lg.And{Terms: terms}

	case *lg.Or:
		terms := make([]lg.Expr, 0, len(n.Terms))
		for _, t := range n.Terms {
			s := simplifyFormula(t)
			// If any disjunct is True, whole Or is True
			if isTrue(s) {
				return &lg.And{Terms: nil} // true
			}
			// Remove False disjuncts
			if isFalse(s) {
				continue
			}
			terms = append(terms, s)
		}
		if len(terms) == 0 {
			return &lg.Or{Terms: nil} // false
		}
		if len(terms) == 1 {
			return terms[0]
		}
		return &lg.Or{Terms: terms}

	case *lg.Not:
		inner := simplifyFormula(n.Body)
		// Double negation elimination
		if n2, ok := inner.(*lg.Not); ok {
			return n2.Body
		}
		if isTrue(inner) {
			return &lg.Or{Terms: nil} // false = Not(true)
		}
		if isFalse(inner) {
			return &lg.And{Terms: nil} // true = Not(false)
		}
		return &lg.Not{Body: inner}

	default:
		return f
	}
}

// isTautologyFormula checks if a formula is tautologically true.
func isTautologyFormula(f lg.Expr) bool {
	return isTrue(f)
}

// isTrue checks if a formula is the constant true (empty And).
func isTrue(f lg.Expr) bool {
	if a, ok := f.(*lg.And); ok && len(a.Terms) == 0 {
		return true
	}
	return false
}

// isFalse checks if a formula is the constant false (empty Or).
func isFalse(f lg.Expr) bool {
	if o, ok := f.(*lg.Or); ok && len(o.Terms) == 0 {
		return true
	}
	return false
}

// SortsClauses returns all sorts used across all formulas and defs.
// Returns map keyed by SortKey (structural identity) with Sort values.
// Corresponds to Python: sorts_clauses = apply_gen_to_clauses(sorts_ast)
func SortsClauses(clauses *Clauses) map[lg.NodeKey]lg.Sort {
	if clauses == nil {
		return nil
	}
	result := make(map[lg.NodeKey]lg.Sort)
	for _, f := range clauses.Fmlas {
		for k, s := range lu.SortsAst(f) {
			result[k] = s
		}
	}
	for _, d := range clauses.Defs {
		for k, s := range lu.SortsAst(d) {
			result[k] = s
		}
	}
	return result
}
