// Theory generation methods for Module.
//
// This corresponds to the theory-related methods in Python's ivy_module.py:
// background_theory, update_theory, get_axioms, theory_context, variant_axioms,
// axioms (property), conjs (property).
package module

import (
	"sort"
	"sync"

	co "github.com/glycerine/goivy/clauseops"
	il "github.com/glycerine/goivy/ivylogic"
	lg "github.com/glycerine/goivy/logic"
)

// theory is the cached background theory, set by UpdateTheory.
// Access via BackgroundTheory.
// We store it directly on the Module struct.

// Add a cached theory field. Since we can't modify the struct in module.go
// from here, we use a package-level map keyed by module pointer.
// This avoids modifying the existing struct definition.
var (
	theoryMu    sync.RWMutex
	theoryCache = make(map[*Module]*co.Clauses)
)

func getTheory(m *Module) (*co.Clauses, bool) {
	theoryMu.RLock()
	defer theoryMu.RUnlock()
	t, ok := theoryCache[m]
	return t, ok
}

func setTheory(m *Module, t *co.Clauses) {
	theoryMu.Lock()
	defer theoryMu.Unlock()
	theoryCache[m] = t
}

// BackgroundTheory returns the cached background theory clauses.
// If UpdateTheory has not been called, returns an empty Clauses.
//
// The inScope parameter is currently unused but mirrors the Python API
// where symbols can filter the returned theory.
func (m *Module) BackgroundTheory(inScope map[string]bool) *co.Clauses {
	if t, ok := getTheory(m); ok {
		return t
	}
	return co.NewClauses(nil, nil, nil)
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

	cls := co.NewClauses(theory, defs, nil)
	setTheory(m, cls)
}

// GetAxioms retrieves all axioms including schema instances.
// Corresponds to Python's Module.get_axioms.
func (m *Module) GetAxioms() []lg.Node {
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
	Instances() []lg.Node
}

// Axioms returns the non-temporal axiom formulas (without labels) from
// LabeledAxioms. Corresponds to Python's Module.axioms property.
func (m *Module) Axioms() []lg.Node {
	var result []lg.Node
	for _, lf := range m.LabeledAxioms {
		if !lf.Temporal {
			result = append(result, DropLabel(lf))
		}
	}
	return result
}

// Conjs returns the conjectures as a slice of Clauses (without labels).
// Each conjecture formula is converted to a Clauses.
// Corresponds to Python's Module.conjs property.
func (m *Module) Conjs() []*co.Clauses {
	var result []*co.Clauses
	for _, c := range m.LabeledConjs {
		cls := co.FormulaToClauses(c.Formula, nil)
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
	m.Instantiator = func(groundTerms []lg.Node) *co.Clauses {
		return instantiateNonEPREntries(nonEPR, groundTerms)
	}

	return func() {
		m.Instantiator = oldInstantiator
	}
}

// instantiateNonEPREntries instantiates non-EPR definitions using ground terms.
// For each ground term whose head symbol has a non-EPR definition,
// the definition is instantiated by substituting the non-variable
// parameters with the term's arguments.
//
// Corresponds to Python instantiate_non_epr (lines 329-343).
func instantiateNonEPREntries(nonEPR map[lg.NodeKey]nonEPREntry, groundTerms []lg.Node) *co.Clauses {
	var theory []lg.Node
	if groundTerms == nil {
		return co.NewClauses(theory, nil, nil)
	}

	matched := make(map[lg.NodeKey]bool)
	for _, term := range groundTerms {
		// Get the head symbol (for structural key lookup into nonEPR)
		var headSym *lg.Const
		var termArgs []lg.Node
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
		subst := make(map[string]lg.Node)
		for i, v := range lhsArgs {
			if i >= len(termArgs) {
				break
			}
			if _, isVar := v.(*lg.Var); !isVar {
				if c, ok := v.(*lg.Const); ok {
					subst[c.Name] = termArgs[i]
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
			inst := co.SubstituteConstantsAST(entry.constraint, subst)
			theory = append(theory, inst)
		}
		matched[lg.Key(term)] = true
	}

	return co.NewClauses(theory, nil, nil)
}

// isGroundNode returns true if a logic node contains no free variables.
func isGroundNode(n lg.Node) bool {
	if n == nil {
		return true
	}
	if _, isVar := n.(*lg.Var); isVar {
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
	ldf        *LabeledFormula
	constraint lg.Node
}

// VariantAxioms generates exclusivity axioms for variant types.
// For each sort with variants, if those variant sorts exist in the
// signature, an exclusivity axiom is generated.
//
// Corresponds to Python's Module.variant_axioms.
func (m *Module) VariantAxioms() []lg.Node {
	var theory []lg.Node

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

// Exclusivity generates an exclusivity axiom for a sort with variants.
// For a sort S with variants V1, V2, ..., Vn, generates:
//
//	forall X:S. (is_V1(X) | is_V2(X) | ... | is_Vn(X))
//	& forall X:S. ~(is_Vi(X) & is_Vj(X)) for all i != j
//
// This is a simplified version; the full version would match Python's
// il.exclusivity exactly.
func Exclusivity(parentSort lg.Sort, variants []lg.Sort) lg.Node {
	if len(variants) == 0 {
		return &lg.And{} // true
	}

	x, _ := lg.NewVar("X", parentSort)

	// Build "is_variant" predicates for each variant.
	// In Ivy, variant membership is tested via type predicates.
	// Here we generate Eq(cast(X), X) style or use variant sort checks.
	// Simplified: generate pairwise inequality for different variants.
	var conjuncts []lg.Node

	// For each pair of distinct variants, X cannot be both.
	for i := 0; i < len(variants); i++ {
		for j := i + 1; j < len(variants); j++ {
			vi := variants[i]
			vj := variants[j]
			// Create sort-check predicates.
			// This is a placeholder — the actual Ivy exclusivity axiom
			// depends on the sort-checking mechanism.
			isVi := makeVariantCheck(x, vi)
			isVj := makeVariantCheck(x, vj)
			if isVi != nil && isVj != nil {
				// ~(is_Vi(X) & is_Vj(X))
				pairConflict := &lg.Not{Body: &lg.And{Terms: []lg.Node{isVi, isVj}}}
				conjuncts = append(conjuncts, pairConflict)
			}
		}
	}

	if len(conjuncts) == 0 {
		return &lg.And{} // true
	}

	body := &lg.And{Terms: conjuncts}
	return &lg.ForAll{Variables: []*lg.Var{x}, Body: body}
}

// makeVariantCheck creates a formula that tests whether x belongs to variant
// sort vs. Returns an application of "is[vsName]" to x.
func makeVariantCheck(x *lg.Var, vs lg.Sort) lg.Node {
	vsName := il.SortName(vs)
	predName := "is." + vsName
	predSort := il.RelationSort([]lg.Sort{x.VSort})
	pred := lg.NewConst(predName, predSort)
	app, err := lg.NewApply(pred, x)
	if err != nil {
		return nil
	}
	return app
}

// DropLabel removes the label from a LabeledFormula, returning just
// the formula. If the argument is not a *LabeledFormula, it is returned
// as-is (for interface{} compatibility).
//
// Corresponds to Python's drop_label function.
func DropLabel(lf interface{}) lg.Node {
	switch v := lf.(type) {
	case *LabeledFormula:
		return v.Formula
	case lg.Node:
		return v
	default:
		return nil
	}
}

// --- helper functions ---

// getLhsArgs returns the arguments of a definition's LHS.
// If LHS is an Apply, returns its Terms. Otherwise returns nil.
func getLhsArgs(def *il.Definition) []lg.Node {
	if app, ok := def.Lhs.(*lg.Apply); ok {
		return app.Terms
	}
	return nil
}

// allVariables returns true if all nodes are *lg.Var.
func allVariables(nodes []lg.Node) bool {
	for _, n := range nodes {
		if _, ok := n.(*lg.Var); !ok {
			return false
		}
	}
	return true
}

// nodesToVars converts a slice of lg.Node to a slice of *lg.Var,
// skipping any non-variable nodes.
func nodesToVars(nodes []lg.Node) []*lg.Var {
	var vars []*lg.Var
	for _, n := range nodes {
		if v, ok := n.(*lg.Var); ok {
			vars = append(vars, v)
		}
	}
	return vars
}

// defToConstraint converts a Definition to a constraint formula.
// For Boolean-sorted RHS, produces Iff(lhs, rhs).
// For non-Boolean, produces Eq(lhs, rhs).
func defToConstraint(d *il.Definition) lg.Node {
	lhs := d.Lhs
	rhs := d.Rhs
	var constraint lg.Node
	if lg.SortEqual(rhs.NodeSort(), lg.Boolean) {
		constraint = &lg.Iff{T1: lhs, T2: rhs}
	} else {
		constraint = &lg.Eq{T1: lhs, T2: rhs}
	}
	return constraint
}

// isEPR checks if a formula is in the EPR fragment (effectively
// propositional after grounding — no function symbols in quantified
// positions). This is a simplified check.
func isEPR(n lg.Node) bool {
	return il.IsPrenexUniversal(n)
}
