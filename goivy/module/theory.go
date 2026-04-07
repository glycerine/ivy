// Theory generation methods for Module.
//
// This corresponds to the theory-related methods in Python's ivy_module.py:
// background_theory, update_theory, get_axioms, theory_context, variant_axioms,
// axioms (property), conjs (property).
package module

import (
	"sort"

	"github.com/glycerine/ivy/goivy/ast"
	il "github.com/glycerine/ivy/goivy/ivylogic"
	lg "github.com/glycerine/ivy/goivy/logic"
)

// BackgroundTheory returns the cached background theory clauses.
// If UpdateTheory has not been called, returns an empty Clauses.
//
// The inScope parameter is currently unused but mirrors the Python API
// where symbols can filter the returned theory.
//
// Corresponds to Python's Module.background_theory (ivy_module.py:116-119):
//
//	if hasattr(self,"theory"): return self.theory
//	return lu.Clauses([])
func (m *Module) BackgroundTheory(inScope map[string]bool) *Clauses {
	if m.Theory != nil {
		return m.Theory
	}
	return NewClauses(nil, nil, nil)
}

// UpdateTheory rebuilds the background theory from axioms and definitions,
// including extensionality axioms for structs and exclusivity axioms for
// variants. The result is cached for subsequent BackgroundTheory calls.
//
// Corresponds to Python's Module.update_theory.
func (m *Module) UpdateTheory() {
	theory := m.GetAxioms()

	var defs []*il.Definition

	// Axioms of derived relations from definitions.
	for _, ldf := range m.Definitions {
		fmla := ldf.Formula

		// Check if it is a Definition node.
		def, isDef := fmla.(*il.Definition)
		if !isDef {
			continue
		}

		// Check that all LHS args are variables.
		lhsArgs := getLhsArgs(def)
		if !allVariables(lhsArgs) {
			continue
		}

		// Skip DefinitionSchema.
		if _, isSchema := fmla.(*il.DefinitionSchema); isSchema {
			continue
		}

		// If RHS is a Some, convert to constraint and add as axiom.
		if _, isSome := def.Rhs.(*il.Some); isSome {
			ax := defToConstraint(def)
			if len(lhsArgs) > 0 {
				vars := nodesToVars(lhsArgs)
				if len(vars) > 0 {
					ax = il.ForAll(vars, ax)
				}
			}
			theory = append(theory, ax)
		} else {
			defs = append(defs, def)
		}
	}

	// Extensionality axioms for structs.
	sortNames := make([]string, 0, len(m.SortDestructors))
	for s := range m.SortDestructors {
		sortNames = append(sortNames, s)
	}
	sort.Strings(sortNames)

	for _, sname := range sortNames {
		destrs := m.SortDestructors[sname]
		// Check if any destructor is in the signature.
		anyInSig := false
		for _, d := range destrs {
			if _, ok := m.Sig.Symbols[d.Name]; ok {
				anyInSig = true
				break
			}
		}
		if anyInSig {
			ea := il.Extensionality(destrs)
			if isEPR(ea) {
				theory = append(theory, ea)
			}
		}
	}

	// Exclusivity axioms for variants.
	theory = append(theory, m.VariantAxioms()...)

	cls := NewClauses(theory, defs, nil)
	m.Theory = cls
}

// GetAxioms retrieves all axioms including schema instances.
// Corresponds to Python's Module.get_axioms.
func (m *Module) GetAxioms() []lg.Expr {
	res := m.Axioms()
	// In Python, schema instances are accumulated on sch.formula.instances.
	// Here we iterate over schemata and collect any instances if they
	// implement an interface that provides them.
	for _, sch := range m.Schemata {
		if si, ok := sch.(SchemaWithInstances); ok {
			res = append(res, si.Instances()...)
		}
	}
	return res
}

// SchemaWithInstances is an interface for schema objects that can provide
// their instantiated formulas.
type SchemaWithInstances interface {
	Instances() []lg.Expr
}

// Axioms returns the non-temporal axiom formulas (without labels) from
// LabeledAxioms. Corresponds to Python's Module.axioms property.
func (m *Module) Axioms() []lg.Expr {
	var result []lg.Expr
	for _, lf := range m.LabeledAxioms {
		if !lf.IsTemporal() {
			result = append(result, DropLabel(lf))
		}
	}
	return result
}

// Conjs returns the conjectures as a slice of Clauses (without labels).
// Each conjecture formula is converted to a Clauses.
// Corresponds to Python's Module.conjs property.
func (m *Module) Conjs() []*Clauses {
	var result []*Clauses
	for _, c := range m.LabeledConjs {
		cls := FormulaToClauses(c.Formula.(lg.Expr), nil)
		// Attach line number info as annotation if needed.
		// The Python code sets clauses.lineno = c.lineno.
		result = append(result, cls)
	}
	return result
}

// TheoryContext sets up to instantiate the non-EPR axioms, then returns
// a cleanup function. This is the Go equivalent of Python's context manager
// Module.theory_context().
//
// Usage:
//
//	cleanup := m.TheoryContext()
//	defer cleanup()
//	// ... use the theory ...
func (m *Module) TheoryContext() func() {
	m.UpdateTheory()

	// Collect non-EPR definitions (definitions with non-variable parameters).
	nonEPR := make(map[lg.NodeKey]nonEPREntry)
	for _, ldf := range m.Definitions {
		def, isDef := ldf.Formula.(*il.Definition)
		if !isDef {
			continue
		}
		lhsArgs := getLhsArgs(def)
		if allVariables(lhsArgs) {
			continue
		}
		// This definition has non-variable parameters.
		cnst := defToConstraint(def)
		defines := def.Defines()
		if defines != nil {
			nonEPR[lg.Key(defines)] = nonEPREntry{ldf: ldf, constraint: cnst}
		}
	}

	// Set the instantiator and return a cleanup function that restores it.
	// Python: lu.instantiator = ModuleTheoryContext(non_epr)
	oldInstantiator := m.Instantiator
	m.Instantiator = func(groundTerms []lg.Expr) *Clauses {
		return instantiateNonEPREntries(nonEPR, groundTerms)
	}

	return func() {
		m.Instantiator = oldInstantiator
	}
}

// ModuleTheoryContext is the Go struct equivalent of Python's ModuleTheoryContext class.
// It holds non-EPR definitions and can instantiate them with ground terms.
// It also supports renaming (Python's ModuleTheoryContext.rename).
//
// Corresponds to Python's ivy_module.ModuleTheoryContext.
type ModuleTheoryContext struct {
	NonEPR map[lg.NodeKey]nonEPREntry
	// OldInstantiator stores the previous instantiator to restore on Exit.
	OldInstantiator func([]lg.Expr) *Clauses
}

// NewModuleTheoryContext creates a new ModuleTheoryContext from non-EPR entries.
func NewModuleTheoryContext(nonEPR map[lg.NodeKey]nonEPREntry) *ModuleTheoryContext {
	return &ModuleTheoryContext{NonEPR: nonEPR}
}

// Enter installs this context as the module's instantiator.
// Corresponds to Python's ModuleTheoryContext.__enter__.
func (tc *ModuleTheoryContext) Enter(m *Module) {
	tc.OldInstantiator = m.Instantiator
	m.Instantiator = tc.Call
}

// Exit restores the previous instantiator.
// Corresponds to Python's ModuleTheoryContext.__exit__.
func (tc *ModuleTheoryContext) Exit(m *Module) {
	m.Instantiator = tc.OldInstantiator
}

// Call instantiates non-EPR definitions with the given ground terms.
// Corresponds to Python's ModuleTheoryContext.__call__.
func (tc *ModuleTheoryContext) Call(groundTerms []lg.Expr) *Clauses {
	return instantiateNonEPREntries(tc.NonEPR, groundTerms)
}

// Rename renames non-EPR entries according to a substitution map,
// extending the non-EPR set with renamed versions.
// Corresponds to Python's ModuleTheoryContext.rename.
func (tc *ModuleTheoryContext) Rename(subst map[lg.NodeKey]*lg.Const) {
	var newEntries []nonEPREntry
	for _, entry := range tc.NonEPR {
		def, isDef := entry.ldf.Formula.(*il.Definition)
		if !isDef {
			continue
		}
		defines := def.Defines()
		if defines == nil {
			continue
		}
		defSym, isSym := defines.(*lg.Const)
		if !isSym {
			continue
		}
		if _, ok := subst[lg.Key(defSym)]; ok {
			renamedLdf := entry.ldf.Cfg.NewLabeledFormulaFrom(entry.ldf, RenameAST(entry.ldf.Formula.(lg.Expr), subst))
			renamedLdf.ID = entry.ldf.ID
			renamedConstraint := RenameAST(entry.constraint, subst)
			newEntries = append(newEntries, nonEPREntry{
				ldf:        renamedLdf,
				constraint: renamedConstraint,
			})
		}
	}
	// Extend the map with new entries.
	for _, ne := range newEntries {
		def, isDef := ne.ldf.Formula.(*il.Definition)
		if !isDef {
			continue
		}
		defines := def.Defines()
		if defines != nil {
			tc.NonEPR[lg.Key(defines)] = ne
		}
	}
}

// instantiateNonEPREntries instantiates non-EPR definitions using ground terms.
// For each ground term whose head symbol has a non-EPR definition,
// the definition is instantiated by substituting the non-variable
// parameters with the term's arguments.
//
// Corresponds to Python instantiate_non_epr (lines 329-343).
func instantiateNonEPREntries(nonEPR map[lg.NodeKey]nonEPREntry, groundTerms []lg.Expr) *Clauses {
	var theory []lg.Expr
	if groundTerms == nil {
		return NewClauses(theory, nil, nil)
	}

	matched := make(map[lg.NodeKey]bool)
	for _, term := range groundTerms {
		// Get the head symbol (for structural key lookup into nonEPR)
		var headSym *lg.Const
		var termArgs []lg.Expr
		switch t := term.(type) {
		case *lg.Const:
			headSym = t
		case *lg.Apply:
			if c, ok := t.Func.(*lg.Const); ok {
				headSym = c
				termArgs = t.Terms
			}
		}
		if headSym == nil {
			continue
		}

		entry, ok := nonEPR[lg.Key(headSym)]
		if !ok || matched[lg.Key(term)] {
			continue
		}

		// Build substitution: non-variable params → term args
		def, isDef := entry.ldf.Formula.(*il.Definition)
		if !isDef {
			continue
		}
		lhsArgs := getLhsArgs(def)
		subst := make(map[lg.NodeKey]lg.Expr)
		for i, v := range lhsArgs {
			if i >= len(termArgs) {
				break
			}
			if _, isVar := v.(*lg.Variable); !isVar {
				if c, ok := v.(*lg.Const); ok {
					subst[lg.Key(c)] = termArgs[i]
				}
			}
		}

		// Check all substituted values are ground
		allGround := true
		for _, val := range subst {
			if !isGroundNode(val) {
				allGround = false
				break
			}
		}

		if allGround && len(subst) > 0 {
			inst := SubstituteConstantsExpr(entry.constraint, subst)
			theory = append(theory, inst)
		}
		matched[lg.Key(term)] = true
	}

	return NewClauses(theory, nil, nil)
}

// isGroundNode returns true if a logic node contains no free variables.
func isGroundNode(n lg.Expr) bool {
	if n == nil {
		return true
	}
	if _, isVar := n.(*lg.Variable); isVar {
		return false
	}
	for _, c := range n.Children() {
		if !isGroundNode(c) {
			return false
		}
	}
	return true
}

// nonEPREntry holds a labeled formula and its constraint form for non-EPR
// instantiation.
type nonEPREntry struct {
	ldf        *ast.LabeledFormula
	constraint lg.Expr
}

// VariantAxioms generates exclusivity axioms for variant types.
// For each sort with variants, if those variant sorts exist in the
// signature, an exclusivity axiom is generated.
//
// Corresponds to Python's Module.variant_axioms.
func (m *Module) VariantAxioms() []lg.Expr {
	var theory []lg.Expr

	sortNames := make([]string, 0, len(m.Variants))
	for s := range m.Variants {
		sortNames = append(sortNames, s)
	}
	sort.Strings(sortNames)

	for _, sname := range sortNames {
		sortVariants := m.Variants[sname]
		// Check if any variant sort is in sig.Sorts and the parent sort
		// is also in sig.Sorts.
		anyInSig := false
		for _, v := range sortVariants {
			vname := il.SortName(v)
			if _, ok := m.Sig.Sorts[vname]; ok {
				anyInSig = true
				break
			}
		}
		if !anyInSig {
			continue
		}
		parentSort, parentOk := m.Sig.Sorts[sname]
		if !parentOk {
			continue
		}
		ea := Exclusivity(parentSort, sortVariants)
		theory = append(theory, ea)
	}
	return theory
}

// Exclusivity generates exclusivity axioms for a sort with variants.
// Corresponds to Python's il.exclusivity (ivy_logic.py:694-706):
//
//	def exclusivity(sort, variants):
//	    def pto(s): return Symbol('*>', RelationSort([sort, s]))
//	    excs = [partial_function(pto(s)) for s in variants]
//	    for s in variants:
//	        x,y,z = [Variable(n,s) for n,s in [('X',sort),('Y',sort),('Z',s)]]
//	        excs.append(Implies(And(pto(s)(x,z),pto(s)(y,z)),Equals(x,y)))
//	    for i1,s1 in enumerate(variants):
//	        for s2 in variants[:i1]:
//	            x,y,z = [Variable(n,s) for n,s in [('X',sort),('Y',s1),('Z',s2)]]
//	            excs.append(Not(And(pto(s1)(x,y),pto(s2)(x,z))))
//	    return And(*excs)
//
// Generates three categories of axioms:
//  1. Partial function axioms: ∀X,Y,Z. (pto(s)(X,Y) ∧ pto(s)(X,Z)) → Y=Z
//  2. Injectivity axioms: ∀X,Y,Z. (pto(s)(X,Z) ∧ pto(s)(Y,Z)) → X=Y
//  3. Pairwise exclusion: ∀X,Y,Z. ¬(pto(s1)(X,Y) ∧ pto(s2)(X,Z))
func Exclusivity(parentSort lg.Sort, variants []lg.Sort) lg.Expr {
	if len(variants) == 0 {
		return &lg.And{} // true
	}

	// pto(s) = Symbol("*>", RelationSort([parentSort, s]))
	pto := func(s lg.Sort) *lg.Const {
		rsort := il.RelationSort([]lg.Sort{parentSort, s})
		return lg.NewConst("*>", rsort)
	}

	var excs []lg.Expr

	// 1. Partial function axioms (Python: [partial_function(pto(s)) for s in variants])
	for _, s := range variants {
		excs = append(excs, il.PartialFunction(pto(s)))
	}

	// 2. Injectivity axioms
	// Python: Implies(And(pto(s)(x,z), pto(s)(y,z)), Equals(x,y))
	for _, s := range variants {
		rel := pto(s)
		x, _ := lg.NewVariable("X", parentSort)
		y, _ := lg.NewVariable("Y", parentSort)
		z, _ := lg.NewVariable("Z", s)
		relXZ := lg.MustApply(rel, x, z)
		relYZ := lg.MustApply(rel, y, z)
		body := &lg.Implies{
			T1: &lg.And{Terms: []lg.Expr{relXZ, relYZ}},
			T2: &lg.Eq{T1: x, T2: y},
		}
		excs = append(excs, &lg.ForAll{Variables: []*lg.Variable{x, y, z}, Body: body})
	}

	// 3. Pairwise exclusion
	// Python: for i1,s1 in enumerate(variants): for s2 in variants[:i1]:
	for i1, s1 := range variants {
		for _, s2 := range variants[:i1] {
			x, _ := lg.NewVariable("X", parentSort)
			y, _ := lg.NewVariable("Y", s1)
			z, _ := lg.NewVariable("Z", s2)
			body := &lg.Not{Body: &lg.And{Terms: []lg.Expr{
				lg.MustApply(pto(s1), x, y),
				lg.MustApply(pto(s2), x, z),
			}}}
			excs = append(excs, &lg.ForAll{Variables: []*lg.Variable{x, y, z}, Body: body})
		}
	}

	return &lg.And{Terms: excs}
}

// DropLabel removes the label from a LabeledFormula, returning just
// the formula. If the argument is not a *ast.LabeledFormula, it is returned
// as-is (for interface{} compatibility).
//
// Corresponds to Python's drop_label function.
func DropLabel(lf interface{}) lg.Expr {
	switch v := lf.(type) {
	case *ast.LabeledFormula:
		return v.Formula.(lg.Expr)
	case lg.Expr:
		return v
	default:
		return nil
	}
}

// --- helper functions ---

// getLhsArgs returns the arguments of a definition's LHS.
// If LHS is an Apply, returns its Terms. Otherwise returns nil.
func getLhsArgs(def *il.Definition) []lg.Expr {
	if app, ok := def.Lhs.(*lg.Apply); ok {
		return app.Terms
	}
	return nil
}

// allVariables returns true if all nodes are *lg.Variable.
func allVariables(nodes []lg.Expr) bool {
	for _, n := range nodes {
		if _, ok := n.(*lg.Variable); !ok {
			return false
		}
	}
	return true
}

// nodesToVars converts a slice of lg.Expr to a slice of *lg.Variable,
// skipping any non-variable nodes.
func nodesToVars(nodes []lg.Expr) []*lg.Variable {
	var vars []*lg.Variable
	for _, n := range nodes {
		if v, ok := n.(*lg.Variable); ok {
			vars = append(vars, v)
		}
	}
	return vars
}

// defToConstraint is now in clauses.go (merged from clauseops)

// isEPR checks if a formula is in the EPR fragment (effectively
// propositional after grounding — no function symbols in quantified
// positions). This is a simplified check.
func isEPR(n lg.Expr) bool {
	return il.IsPrenexUniversal(n)
}
