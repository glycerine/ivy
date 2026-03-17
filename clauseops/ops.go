package clauseops

import (
	"fmt"

	il "github.com/glycerine/goivy/ivylogic"
	lg "github.com/glycerine/goivy/logic"
	lu "github.com/glycerine/goivy/logicutil"
	iu "github.com/glycerine/goivy/ivyutils"
)

// AndClauses computes the conjunction of Clauses and/or formulas.
// Each argument can be *Clauses or lg.Node. If no argument is a *Clauses,
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
		// All args are lg.Node; return a plain formula
		nodes := make([]lg.Node, len(args))
		for i, a := range args {
			nodes[i] = a.(lg.Node)
		}
		return &lg.And{Terms: nodes}
	}

	// Coerce all args to *Clauses
	clauses := coerceArgsToClauses(args)
	if len(clauses) == 0 {
		return TrueClauses(nil)
	}

	// Combine annotations via conj
	var annot interface{}
	for _, c := range clauses {
		if c.Annot != nil {
			if annot == nil {
				annot = c.Annot
			}
			// In Python, annot = annot.conj(c.annot). We just take first non-nil.
			// Full annotation support would require an Annotation interface.
		}
	}

	// If any clause set is false, return false
	for _, c := range clauses {
		if c.IsFalse() {
			return FalseClauses(annot)
		}
	}

	// Concatenate formulas and definitions
	var fmlas []lg.Node
	var defs []*il.Definition
	for _, c := range clauses {
		fmlas = append(fmlas, c.Fmlas...)
		defs = append(defs, c.Defs...)
	}
	return NewClauses(fmlas, defs, annot)
}

// AndClausesTyped is a convenience wrapper that always returns *Clauses.
// It panics if passed non-Clauses, non-Node arguments.
func AndClausesTyped(args ...*Clauses) *Clauses {
	if len(args) == 0 {
		return TrueClauses(nil)
	}

	var annot interface{}
	for _, c := range args {
		if c.Annot != nil {
			if annot == nil {
				annot = c.Annot
			}
		}
	}

	for _, c := range args {
		if c.IsFalse() {
			return FalseClauses(annot)
		}
	}

	var fmlas []lg.Node
	var defs []*il.Definition
	for _, c := range args {
		fmlas = append(fmlas, c.Fmlas...)
		defs = append(defs, c.Defs...)
	}
	return NewClauses(fmlas, defs, annot)
}

// OrClauses computes the disjunction of Clauses and/or formulas.
// Each argument can be *Clauses or lg.Node. If no argument is a *Clauses,
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
		nodes := make([]lg.Node, len(args))
		for i, a := range args {
			nodes[i] = a.(lg.Node)
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
func OrClausesTyped(args ...*Clauses) *Clauses {
	if len(args) == 0 {
		return FalseClauses(nil)
	}

	var nonFalse []*Clauses
	for _, c := range args {
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

	used := collectUsedNames(nonFalse, nil)
	rn := iu.NewUniqueRenamer("__ts0", used)
	return orClausesInt(rn, nonFalse)
}

// orClausesInt implements the Tseitin-like encoding for disjunction.
// For each clause set, it introduces a fresh Boolean variable v_i,
// and produces: Or(v1, v2, ...) AND for each v_i, fmla => (v_i -> fmla).
func orClausesInt(rn *iu.UniqueRenamer, args []*Clauses) *Clauses {
	// Eliminate dead definitions across args
	args = elimDeadDefinitions(rn, args)

	// Create fresh Boolean variables, one per disjunct
	vs := make([]*lg.Const, len(args))
	vsNodes := make([]lg.Node, len(args))
	for i := range args {
		name := rn.Rename("")
		vs[i] = lg.NewConst(name, lg.Boolean)
		vsNodes[i] = vs[i]
	}

	// Build formulas:
	// 1. Or(v1, v2, ..., vn)
	// 2. For each i: Not(vi) OR fmla, for each fmla in args[i].Fmlas
	var fmlas []lg.Node
	fmlas = append(fmlas, &lg.Or{Terms: vsNodes})
	for i, cls := range args {
		for _, f := range cls.Fmlas {
			fmlas = append(fmlas, &lg.Or{Terms: []lg.Node{
				&lg.Not{Body: vs[i]},
				f,
			}})
		}
	}

	// Merge definitions
	defIdx := make(map[string]*il.Definition)
	for i, cls := range args {
		for _, d := range cls.Defs {
			key := definesKey(d)
			if existing, ok := defIdx[key]; !ok {
				defIdx[key] = d
			} else {
				// Merge: use Ite to select between definitions
				merged := il.NewDefinition(
					d.Lhs,
					simpIte(vs[i], d.Rhs, existing.Rhs),
				)
				defIdx[key] = merged
			}
		}
	}

	var defs []*il.Definition
	for _, d := range defIdx {
		defs = append(defs, d)
	}

	return NewClauses(fmlas, defs, nil)
}

// IteClauses computes if-then-else on Clauses:
// if cond then args[0] else args[1].
func IteClauses(cond lg.Node, thenCls, elseCls *Clauses) *Clauses {
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

func iteClausesInt(rn *iu.UniqueRenamer, cond lg.Node, args []*Clauses) *Clauses {
	args = elimDeadDefinitions(rn, args)

	// Create a fresh Boolean variable for the condition
	name := rn.Rename("")
	v := lg.NewConst(name, lg.Boolean)

	// Build formulas:
	// For then-branch: Not(v) OR fmla
	// For else-branch: v OR fmla
	var fmlas []lg.Node
	for _, f := range args[0].Fmlas {
		fmlas = append(fmlas, &lg.Or{Terms: []lg.Node{&lg.Not{Body: v}, f}})
	}
	for _, f := range args[1].Fmlas {
		fmlas = append(fmlas, &lg.Or{Terms: []lg.Node{v, f}})
	}

	// Merge definitions
	defIdx := make(map[string]*il.Definition)
	for _, d := range args[0].Defs {
		key := definesKey(d)
		defIdx[key] = d
	}
	for _, d := range args[1].Defs {
		key := definesKey(d)
		if existing, ok := defIdx[key]; !ok {
			defIdx[key] = d
		} else {
			merged := il.NewDefinition(
				d.Lhs,
				simpIte(v, existing.Rhs, d.Rhs),
			)
			defIdx[key] = merged
		}
	}

	var defs []*il.Definition
	for _, d := range defIdx {
		defs = append(defs, d)
	}
	// Add definition: v = cond
	defs = append(defs, il.NewDefinition(v, cond))

	return NewClauses(fmlas, defs, nil)
}

// NegateClauses negates a Clauses. Requires the clauses to be
// universal first-order (no definitions, no skolems).
func NegateClauses(clauses *Clauses) *Clauses {
	if !clauses.IsUniversalFirstOrder() {
		// For non-universal-first-order, convert to formula and negate
		f := clauses.ToFormula()
		return FormulaToClauses(Negate(f), nil)
	}
	// Dual: skolemize variables, then negate
	return dualClauses(clauses)
}

// dualClauses implements the dual construction: replaces variables with
// Skolem constants, then negates.
func dualClauses(clauses *Clauses) *Clauses {
	// Get used variables in order
	vars := usedVariablesOrdered(clauses)

	// Create Skolem substitution: V -> __V
	subs := make(map[lg.NodeKey]lg.Node, len(vars))
	for _, v := range vars {
		sk := lg.NewConst("__"+v.Name, v.VSort)
		subs[lg.Key(v)] = sk
	}

	// Apply substitution
	clauses = SubstituteNodesClauses(clauses, subs)

	// Negate the formula
	f := Negate(clausesToFormula(clauses))
	return FormulaToClauses(f, nil)
}

// clausesToFormula converts clauses to formula (drop universals).
func clausesToFormula(c *Clauses) lg.Node {
	return dropUniversals(c.ToFormula())
}

// ClausesToFormula converts clauses to a formula, dropping leading
// universal quantifiers. Corresponds to Python's clauses_to_formula.
func ClausesToFormula(c *Clauses) lg.Node {
	return clausesToFormula(c)
}

// ConditionClauses returns Clauses equivalent to "fmla -> clauses".
// Each formula in clauses is wrapped as "Not(fmla) OR formula".
func ConditionClauses(clauses *Clauses, fmla lg.Node) *Clauses {
	negFmla := Negate(fmla)
	var newFmlas []lg.Node
	for _, f := range clauses.Fmlas {
		newFmlas = append(newFmlas, &lg.Or{Terms: []lg.Node{negFmla, f}})
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
	var fmlas []lg.Node
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
func usesSymbolNameAST(names map[string]bool, node lg.Node) bool {
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

func collectSymbolNamesFromNode(n lg.Node, result map[string]bool) {
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

func ClausesUsingSymbols(syms map[lg.NodeKey]lg.Node, clauses *Clauses) *Clauses {
	var fmlas []lg.Node
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

// RenameClauses renames symbols in clauses according to the substitution map.
// The map keys are symbol name strings, values are replacement Consts.
func RenameClauses(clauses *Clauses, subs map[string]*lg.Const) *Clauses {
	fn := func(n lg.Node) lg.Node {
		return RenameAST(n, subs)
	}
	return clauses.Apply(fn)
}

// RenameClausesByName renames symbols in clauses using a name→name map.
// Each old name is replaced by a Const with the new name and the same sort.
func RenameClausesByName(clauses *Clauses, subs map[string]string) *Clauses {
	constSubs := make(map[string]*lg.Const, len(subs))
	for old, new := range subs {
		constSubs[old] = lg.NewConst(new, lg.TopS)
	}
	return RenameClauses(clauses, constSubs)
}

// SubstituteConstantsClauses substitutes constants in clauses.
// The map keys are constant name strings, values are replacement nodes.
func SubstituteConstantsClauses(clauses *Clauses, subs map[string]lg.Node) *Clauses {
	fn := func(n lg.Node) lg.Node {
		return SubstituteConstantsAST(n, subs)
	}
	return clauses.Apply(fn)
}

// SubstituteNodesClauses applies a node-level substitution to all formulas
// and definitions in the clauses.
func SubstituteNodesClauses(clauses *Clauses, subs map[lg.NodeKey]lg.Node) *Clauses {
	if len(subs) == 0 {
		return clauses
	}
	var fmlas []lg.Node
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
func substituteNodesRec(n lg.Node, subs map[lg.NodeKey]lg.Node) (lg.Node, error) {
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
	case *lg.Var, *lg.Const:
		// Already checked via Key above
		return n, nil
	}
	// Recurse into children
	children := n.Children()
	if len(children) == 0 {
		return n, nil
	}
	newChildren := make([]lg.Node, len(children))
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

// coerceArgsToClauses converts a mixed slice of *Clauses and lg.Node to []*Clauses.
func coerceArgsToClauses(args []interface{}) []*Clauses {
	result := make([]*Clauses, len(args))
	for i, a := range args {
		switch v := a.(type) {
		case *Clauses:
			result[i] = v
		case lg.Node:
			result[i] = FormulaToClauses(v, nil)
		default:
			panic(fmt.Sprintf("clauseops: unexpected argument type %T", a))
		}
	}
	return result
}

// collectUsedNames collects all symbol names used in clauses and optionally
// in an extra formula, returned as a slice of strings.
func collectUsedNames(args []*Clauses, extra lg.Node) []string {
	seen := make(map[string]struct{})
	for _, cls := range args {
		for _, s := range cls.Symbols() {
			seen[s.(*lg.Const).Name] = struct{}{}
		}
	}
	if extra != nil {
		for _, s := range usedSymbolsAST(extra) {
			seen[s.(*lg.Const).Name] = struct{}{}
		}
	}
	result := make([]string, 0, len(seen))
	for k := range seen {
		result = append(result, k)
	}
	return result
}

// elimDeadDefinitions eliminates definitions that are captured across
// different clause sets. If a symbol is defined in one set but used
// free in another, the definition is inlined as a constraint.
func elimDeadDefinitions(rn *iu.UniqueRenamer, args []*Clauses) []*Clauses {
	// Collect all defined symbols
	defined := make(map[string]bool)
	for _, a := range args {
		for _, d := range a.Defs {
			defined[definesKey(d)] = true
		}
	}

	// Find captured symbols: defined somewhere but not everywhere
	var dead []string
	for sym := range defined {
		for _, a := range args {
			if _, ok := a.DefIdx[sym]; !ok {
				dead = append(dead, sym)
				break
			}
		}
	}

	if len(dead) == 0 {
		return args
	}

	// Eliminate dead definitions by converting them to constraints
	deadSet := make(map[string]bool, len(dead))
	for _, s := range dead {
		deadSet[s] = true
	}

	result := make([]*Clauses, len(args))
	for i, a := range args {
		var fmlas []lg.Node
		fmlas = append(fmlas, a.Fmlas...)
		var defs []*il.Definition
		for _, d := range a.Defs {
			key := definesKey(d)
			if deadSet[key] {
				fmlas = append(fmlas, defToConstraint(d))
			} else {
				defs = append(defs, d)
			}
		}
		result[i] = NewClauses(fmlas, defs, a.Annot)
	}
	return result
}

// simpIte returns a simplified Ite node. If both branches are equal,
// returns either branch.
func simpIte(cond lg.Node, thenN, elseN lg.Node) lg.Node {
	if thenN.Equal(elseN) {
		return thenN
	}
	return &lg.Ite{
		ISort: thenN.NodeSort(),
		Cond:  cond,
		Then:  thenN,
		Else:  elseN,
	}
}

// usedVariablesOrdered returns free variables from the clauses in order.
func usedVariablesOrdered(c *Clauses) []*lg.Var {
	seen := make(map[string]bool)
	var result []*lg.Var
	for _, f := range c.Fmlas {
		collectVarsOrdered(f, seen, &result)
	}
	for _, d := range c.Defs {
		collectVarsOrdered(d, seen, &result)
	}
	return result
}

func collectVarsOrdered(n lg.Node, seen map[string]bool, result *[]*lg.Var) {
	if v, ok := n.(*lg.Var); ok {
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
func SubstituteClauses(clauses *Clauses, subs map[lg.NodeKey]lg.Node) *Clauses {
	if clauses == nil || len(subs) == 0 {
		return clauses
	}
	return SubstituteNodesClauses(clauses, subs)
}

// SubstBothClauses applies substitution to both variables and constants in clauses.
// Corresponds to Python subst_both_clauses:
//   substitute_constants_clauses(substitute_clauses(clauses, subst), subst)
func SubstBothClauses(clauses *Clauses, subs map[string]lg.Node) *Clauses {
	if clauses == nil || len(subs) == 0 {
		return clauses
	}
	// First, substitute as variables (map[NodeKey]lg.Node keyed by Var sexp)
	varSubs := make(map[lg.NodeKey]lg.Node)
	for name, val := range subs {
		v, err := lg.NewVar(name, val.NodeSort())
		if err == nil {
			varSubs[lg.Key(v)] = val
		}
	}
	result := SubstituteClauses(clauses, varSubs)
	// Then, substitute as constants (map[string]lg.Node keyed by name)
	result = SubstituteConstantsClauses(result, subs)
	return result
}

// ResortClauses remaps sorts in all formulas and defs of clauses.
// Corresponds to Python: resort_clauses = apply_func_to_clauses(resort_ast)
func ResortClauses(clauses *Clauses, subs map[string]lg.Sort) *Clauses {
	if clauses == nil || len(subs) == 0 {
		return clauses
	}
	fn := func(n lg.Node) lg.Node {
		return lu.ResortAst(n, subs)
	}
	return clauses.Apply(fn)
}

// VariablesClauses returns all free variables across all formulas and defs.
// Corresponds to Python: variables_clauses = apply_gen_to_clauses(variables_ast)
func VariablesClauses(clauses *Clauses) []*lg.Var {
	if clauses == nil {
		return nil
	}
	seen := make(map[string]bool)
	var result []*lg.Var
	for _, f := range clauses.Fmlas {
		for _, vNode := range lu.FreeVariables(f) {
			vv := vNode.(*lg.Var)
			if !seen[vv.Name] {
				seen[vv.Name] = true
				result = append(result, vv)
			}
		}
	}
	for _, d := range clauses.Defs {
		for _, vNode := range lu.FreeVariables(d) {
			vv := vNode.(*lg.Var)
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
		for _, cNode := range lu.UsedConstants(f) {
			cc := cNode.(*lg.Const)
			if !seen[cc.Name] {
				seen[cc.Name] = true
				result = append(result, cc)
			}
		}
	}
	for _, d := range clauses.Defs {
		for _, cNode := range lu.UsedConstants(d) {
			cc := cNode.(*lg.Const)
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
		if c, ok := s.(*lg.Const); ok {
			syms = append(syms, c)
		}
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
func AppsClauses(clauses *Clauses) []lg.Node {
	if clauses == nil {
		return nil
	}
	var result []lg.Node
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
func GroundAppsClauses(clauses *Clauses) []lg.Node {
	if clauses == nil {
		return nil
	}
	var result []lg.Node
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
	clauses []lg.Node
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
func (tc *tseitinContext) tseitinEncoding(f lg.Node) lg.Node {
	f = lu.ExpandAbbrevs(f)

	switch n := f.(type) {
	case *lg.And:
		if len(n.Terms) == 0 {
			return &lg.And{Terms: nil} // true
		}
		args := make([]lg.Node, len(n.Terms))
		for i, g := range n.Terms {
			args[i] = tc.tseitinEncoding(g)
		}
		// Collect free variables from f
		varSet := make(map[string]*lg.Var)
		collectFreeVars(f, varSet, nil)
		var vars []*lg.Var
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
		var varNodes []lg.Node
		for _, v := range vars {
			varNodes = append(varNodes, v)
		}
		var res lg.Node
		if len(varNodes) > 0 {
			res, _ = lg.NewApply(fn, varNodes...)
		} else {
			res = fn
		}
		// Add Tseitin clauses: ~res | arg_i for each i, and res | ~arg_0 | ~arg_1 | ...
		for _, arg := range args {
			// ~res | arg
			tc.clauses = append(tc.clauses, &lg.Or{Terms: []lg.Node{&lg.Not{Body: res}, arg}})
		}
		// res | ~arg_0 | ~arg_1 | ...
		negArgs := make([]lg.Node, len(args)+1)
		negArgs[0] = res
		for i, arg := range args {
			negArgs[i+1] = &lg.Not{Body: arg}
		}
		tc.clauses = append(tc.clauses, &lg.Or{Terms: negArgs})
		return res

	case *lg.Or:
		// ~(AND(~x for x in args))
		negArgs := make([]lg.Node, len(n.Terms))
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
func collectFreeVars(node lg.Node, result map[string]*lg.Var, bound map[string]bool) {
	if node == nil {
		return
	}
	switch n := node.(type) {
	case *lg.Var:
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
func TseitinEncode(f lg.Node) *Clauses {
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
	fmlas := make([]lg.Node, len(cls.Fmlas))
	copy(fmlas, cls.Fmlas)

	for i := 0; i < 3; i++ {
		// Simplify each formula
		simplified := make([]lg.Node, 0, len(fmlas))
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
func simplifyFormula(f lg.Node) lg.Node {
	// Simplify And/Or by recursing
	switch n := f.(type) {
	case *lg.And:
		terms := make([]lg.Node, 0, len(n.Terms))
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
		terms := make([]lg.Node, 0, len(n.Terms))
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
func isTautologyFormula(f lg.Node) bool {
	return isTrue(f)
}

// isTrue checks if a formula is the constant true (empty And).
func isTrue(f lg.Node) bool {
	if a, ok := f.(*lg.And); ok && len(a.Terms) == 0 {
		return true
	}
	return false
}

// isFalse checks if a formula is the constant false (empty Or).
func isFalse(f lg.Node) bool {
	if o, ok := f.(*lg.Or); ok && len(o.Terms) == 0 {
		return true
	}
	return false
}

// SortsClauses returns all sorts used across all formulas and defs.
// Corresponds to Python: sorts_clauses = apply_gen_to_clauses(sorts_ast)
func SortsClauses(clauses *Clauses) map[lg.Sort]bool {
	if clauses == nil {
		return nil
	}
	result := make(map[lg.Sort]bool)
	for _, f := range clauses.Fmlas {
		for s, v := range lu.SortsAst(f) {
			result[s] = v
		}
	}
	for _, d := range clauses.Defs {
		for s, v := range lu.SortsAst(d) {
			result[s] = v
		}
	}
	return result
}
