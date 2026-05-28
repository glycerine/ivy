// Theory generation methods for Module.
//
// This corresponds to the theory-related methods in Python's ivy_module.py:
// background_theory, update_theory, get_axioms, theory_context, variant_axioms,
// axioms (property), conjs (property).
package goivy

import (
	"sort"
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

	var defs []*IvyDefinition

	// Axioms of derived relations from definitions.
	for _, ldf := range m.Definitions {
		fmla := ldf.Formula

		// Extract Definition — handle both *LogicDefinition and *LogicDefinitionSchema.
		// In Go, *LogicDefinitionSchema embeds Definition but doesn't satisfy
		// the *LogicDefinition type assertion. Python isinstance() handles both.
		var def *IvyDefinition
		var isSchema bool
		if d, ok := fmla.(*IvyDefinition); ok {
			def = d
		} else if ds, ok := fmla.(*IvyDefinitionSchema); ok {
			def = &ds.LogicDefinition
			isSchema = true
		} else {
			continue
		}

		// Python: cnst = ldf.formula.to_constraint()
		// Called unconditionally for ALL definitions to match Python trace output.
		// The result is unused here but the call produces an xtracer trace.
		moduleDefToConstraint(def)

		// Check that all LHS args are variables.
		lhsArgs := getLhsArgs(def)
		if !allVariables(lhsArgs) {
			continue
		}

		// Skip DefinitionSchema.
		if isSchema {
			continue
		}

		// If RHS is a Some, convert to constraint and add as axiom.
		if _, isSome := def.Rhs.(*LogicSome); isSome {
			ax := moduleDefToConstraint(def)
			if len(lhsArgs) > 0 {
				vars := nodesToVars(lhsArgs)
				if len(vars) > 0 {
					ax = IvyForAll(vars, ax)
				}
			}
			theory = append(theory, ax)
		} else {
			defs = append(defs, def)
		}
	}

	// Extensionality axioms for structs.
	sortNames := make([]string, 0, m.SortDestructors.Len())
	for s := range m.SortDestructors.All() {
		sortNames = append(sortNames, s)
	}
	sort.Strings(sortNames)

	for _, sname := range sortNames {
		destrs := m.SortDestructors.Get(sname)
		// Check if any destructor is in the signature.
		anyInSig := false
		for _, d := range destrs {
			if _, ok := m.Sig.Symbols.Get2(d.Name); ok {
				anyInSig = true
				break
			}
		}
		if anyInSig {
			ea := Extensionality(destrs)
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
func (m *Module) GetAxioms() []Expr {
	res := m.Axioms()
	// In Python, schema instances are accumulated on sch.formula.instances.
	// Here we iterate over schemata and collect any instances if they
	// implement an interface that provides them.
	for _, sch := range m.Schemata.All() {
		if astSchema, ok := sch.(*Schema); ok {
			for _, inst := range astSchema.Instances {
				if expr, ok := inst.(Expr); ok {
					res = append(res, expr)
				}
			}
		} else if si, ok := sch.(SchemaWithInstances); ok {
			res = append(res, si.Instances()...)
		}
	}
	return res
}

// SchemaWithInstances is an interface for schema objects that can provide
// their instantiated formulas.
type SchemaWithInstances interface {
	Instances() []Expr
}

// Axioms returns the non-temporal axiom formulas (without labels) from
// LabeledAxioms. Corresponds to Python's Module.axioms property.
func (m *Module) Axioms() []Expr {
	var result []Expr
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
		cls := FormulaToClauses(c.Formula.(Expr), nil)
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
	nonEPR := make(map[NodeKey]nonEPREntry)
	for _, ldf := range m.Definitions {
		// Extract Definition — handle both *LogicDefinition and *LogicDefinitionSchema.
		var def *IvyDefinition
		if d, ok := ldf.Formula.(*IvyDefinition); ok {
			def = d
		} else if ds, ok := ldf.Formula.(*IvyDefinitionSchema); ok {
			def = &ds.LogicDefinition
		} else {
			continue
		}

		// Python: cnst = ldf.formula.to_constraint()
		// Called for ALL definitions (result used only for non-EPR).
		cnst := moduleDefToConstraint(def)

		lhsArgs := getLhsArgs(def)
		if allVariables(lhsArgs) {
			continue
		}
		// This definition has non-variable parameters.
		defines := def.Defines()
		if defines != nil {
			nonEPR[Key(defines)] = nonEPREntry{ldf: ldf, constraint: cnst}
		}
	}

	// Set the instantiator and return a cleanup function that restores it.
	// Python: lu.instantiator = ModuleTheoryContext(non_epr)
	oldInstantiator := m.Instantiator
	m.Instantiator = func(groundTerms []Expr) *Clauses {
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
	NonEPR map[NodeKey]nonEPREntry
	// OldInstantiator stores the previous instantiator to restore on Exit.
	OldInstantiator func([]Expr) *Clauses
}

// NewModuleTheoryContext creates a new ModuleTheoryContext from non-EPR entries.
func NewModuleTheoryContext(nonEPR map[NodeKey]nonEPREntry) *ModuleTheoryContext {
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
func (tc *ModuleTheoryContext) Call(groundTerms []Expr) *Clauses {
	return instantiateNonEPREntries(tc.NonEPR, groundTerms)
}

// Rename renames non-EPR entries according to a substitution map,
// extending the non-EPR set with renamed versions.
// Corresponds to Python's ModuleTheoryContext.rename.
func (tc *ModuleTheoryContext) Rename(subst map[NodeKey]*Const) {
	var newEntries []nonEPREntry
	for _, entry := range tc.NonEPR {
		def, isDef := entry.ldf.Formula.(*IvyDefinition)
		if !isDef {
			continue
		}
		defines := def.Defines()
		if defines == nil {
			continue
		}
		defSym, isSym := defines.(*Const)
		if !isSym {
			continue
		}
		if _, ok := subst[Key(defSym)]; ok {
			// Python ivy_module.py:1026 calls lu.rename_ast(df, subst), which
			// (via ivy_logic_utils.py:207 `ast.clone(args)`) dispatches to
			// LF.clone for an LF — preserving id and metadata.
			renamedFormula := RenameAST(entry.ldf.Formula.(Expr), subst)
			renamedLdf := entry.ldf.Clone([]Node{entry.ldf.Label, renamedFormula}).(*LabeledFormula)
			renamedConstraint := RenameAST(entry.constraint, subst)
			newEntries = append(newEntries, nonEPREntry{
				ldf:        renamedLdf,
				constraint: renamedConstraint,
			})
		}
	}
	// Extend the map with new entries.
	for _, ne := range newEntries {
		def, isDef := ne.ldf.Formula.(*IvyDefinition)
		if !isDef {
			continue
		}
		defines := def.Defines()
		if defines != nil {
			tc.NonEPR[Key(defines)] = ne
		}
	}
}

// instantiateNonEPREntries instantiates non-EPR definitions using ground terms.
// For each ground term whose head symbol has a non-EPR definition,
// the definition is instantiated by substituting the non-variable
// parameters with the term's arguments.
//
// Corresponds to Python instantiate_non_epr (lines 329-343).
func instantiateNonEPREntries(nonEPR map[NodeKey]nonEPREntry, groundTerms []Expr) *Clauses {
	var theory []Expr
	if groundTerms == nil {
		return NewClauses(theory, nil, nil)
	}

	matched := make(map[NodeKey]bool)
	for _, term := range groundTerms {
		// Get the head symbol (for structural key lookup into nonEPR)
		var headSym *Const
		var termArgs []Expr
		switch t := term.(type) {
		case *Const:
			headSym = t
		case *Apply:
			if c, ok := t.Func.(*Const); ok {
				headSym = c
				termArgs = t.Terms
			}
		}
		if headSym == nil {
			continue
		}

		entry, ok := nonEPR[Key(headSym)]
		if !ok || matched[Key(term)] {
			continue
		}

		// Build substitution: non-variable params → term args
		def, isDef := entry.ldf.Formula.(*IvyDefinition)
		if !isDef {
			continue
		}
		lhsArgs := getLhsArgs(def)
		subst := make(map[NodeKey]Expr)
		for i, v := range lhsArgs {
			if i >= len(termArgs) {
				break
			}
			if _, isVar := v.(*LogicVariable); !isVar {
				if c, ok := v.(*Const); ok {
					subst[Key(c)] = termArgs[i]
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
		matched[Key(term)] = true
	}

	return NewClauses(theory, nil, nil)
}

// isGroundNode returns true if a logic node contains no free variables.
func isGroundNode(n Expr) bool {
	if n == nil {
		return true
	}
	if _, isVar := n.(*LogicVariable); isVar {
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
	constraint Expr
}

// VariantAxioms generates exclusivity axioms for variant types.
// For each sort with variants, if those variant sorts exist in the
// signature, an exclusivity axiom is generated.
//
// Corresponds to Python's Module.variant_axioms.
func (m *Module) VariantAxioms() []Expr {
	var theory []Expr

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
			vname := IvySortName(v)
			if _, ok := m.Sig.Sorts.Get2(vname); ok {
				anyInSig = true
				break
			}
		}
		if !anyInSig {
			continue
		}
		parentSort, parentOk := m.Sig.Sorts.Get2(sname)
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
//	    def pto(s): return Symbol('*>', LogicRelationSort([sort, s]))
//	    excs = [partial_function(pto(s)) for s in variants]
//	    for s in variants:
//	        x,y,z = [Variable(n,s) for n,s in [('X',sort),('Y',sort),('Z',s)]]
//	        excs.append(Implies(And(pto(s)(x,z),pto(s)(y,z)),Equals(x,y)))
//	    for i1,s1 in enumerate(variants):
//	        for s2 in variants[:i1]:
//	            x,y,z = [Variable(n,s) for n,s in [('X',sort),('Y',s1),('Z',s2)]]
//	            excs.append(Not(And(pto(s1)(x,y),pto(s2)(x,z))))
//	    return LogicAnd(*excs)
//
// Generates three categories of axioms:
//  1. Partial function axioms: ∀X,Y,Z. (pto(s)(X,Y) ∧ pto(s)(X,Z)) → Y=Z
//  2. Injectivity axioms: ∀X,Y,Z. (pto(s)(X,Z) ∧ pto(s)(Y,Z)) → X=Y
//  3. Pairwise exclusion: ∀X,Y,Z. ¬(pto(s1)(X,Y) ∧ pto(s2)(X,Z))
func Exclusivity(parentSort Sort, variants []Sort) Expr {
	if len(variants) == 0 {
		return &LogicAnd{} // true
	}

	// pto(s) = Symbol("*>", LogicRelationSort([parentSort, s]))
	pto := func(s Sort) *Const {
		rsort := LogicRelationSort([]Sort{parentSort, s})
		return NewConst("*>", rsort)
	}

	var excs []Expr

	// 1. Partial function axioms (Python: [partial_function(pto(s)) for s in variants])
	for _, s := range variants {
		excs = append(excs, PartialFunction(pto(s)))
	}

	// 2. Injectivity axioms
	// Python: Implies(And(pto(s)(x,z), pto(s)(y,z)), Equals(x,y))
	for _, s := range variants {
		rel := pto(s)
		x, _ := NewVariable("X", parentSort)
		y, _ := NewVariable("Y", parentSort)
		z, _ := NewVariable("Z", s)
		relXZ := MustApply(rel, x, z)
		relYZ := MustApply(rel, y, z)
		body := &LogicImplies{
			T1: &LogicAnd{Terms: []Expr{relXZ, relYZ}},
			T2: &Eq{T1: x, T2: y},
		}
		excs = append(excs, &ForAll{Variables: []*LogicVariable{x, y, z}, Body: body})
	}

	// 3. Pairwise exclusion
	// Python: for i1,s1 in enumerate(variants): for s2 in variants[:i1]:
	for i1, s1 := range variants {
		for _, s2 := range variants[:i1] {
			x, _ := NewVariable("X", parentSort)
			y, _ := NewVariable("Y", s1)
			z, _ := NewVariable("Z", s2)
			body := &LogicNot{Body: &LogicAnd{Terms: []Expr{
				MustApply(pto(s1), x, y),
				MustApply(pto(s2), x, z),
			}}}
			excs = append(excs, &ForAll{Variables: []*LogicVariable{x, y, z}, Body: body})
		}
	}

	return &LogicAnd{Terms: excs}
}

// DropLabel removes the label from a LabeledFormula, returning just
// the formula. If the argument is not a *ast.LabeledFormula, it is returned
// as-is (for interface{} compatibility).
//
// Corresponds to Python's drop_label function.
func DropLabel(lf interface{}) Expr {
	switch v := lf.(type) {
	case *LabeledFormula:
		return v.Formula.(Expr)
	case Expr:
		return v
	default:
		return nil
	}
}

// --- helper functions ---

// getLhsArgs returns the arguments of a definition's LHS.
// If LHS is an Apply, returns its Terms. Otherwise returns nil.
func getLhsArgs(def *IvyDefinition) []Expr {
	if app, ok := def.Lhs.(*Apply); ok {
		return app.Terms
	}
	return nil
}

// allVariables returns true if all nodes are *lg.Variable.
func allVariables(nodes []Expr) bool {
	for _, n := range nodes {
		if _, ok := n.(*LogicVariable); !ok {
			return false
		}
	}
	return true
}

// nodesToVars converts a slice of lg.Expr to a slice of *lg.Variable,
// skipping any non-variable nodes.
func nodesToVars(nodes []Expr) []*LogicVariable {
	var vars []*LogicVariable
	for _, n := range nodes {
		if v, ok := n.(*LogicVariable); ok {
			vars = append(vars, v)
		}
	}
	return vars
}

// defToConstraint is now in clauses.go (merged from clauseops)

// isEPR matches Python ivy_logic.is_epr: reject only existential quantifiers
// that depend on currently universal variables. In particular, a universal
// quantifier nested under a connective is still EPR, which matters for
// higher-arity struct-field extensionality axioms.
func isEPR(n Expr) bool {
	return IsEPR(n)
}
