package goivy

import (
	"fmt"
	"strings"

	"github.com/glycerine/ivy/goivy/xtracer"
)

// DomainSetup processes top-level declarations, replacing Python's
// IvyDomainSetup class.
type DomainSetup struct {
	Compiler *Compiler

	// LastFact holds the last compiled property/axiom LabeledFormula, used by
	// proof and named declarations. Matches Python's self.last_fact.
	LastFact *LabeledFormula
}

// NewDomainSetup creates a new declaration interpreter.
func NewDomainSetup(c *Compiler) *DomainSetup {
	return &DomainSetup{Compiler: c}
}

// ProcessDecls processes all declarations in a declaration list.
// Each declaration is dispatched to the appropriate handler based on its type.
func (d *DomainSetup) ProcessDecls(decls []Node) error {
	for _, decl := range decls {
		if err := d.ProcessDecl(decl); err != nil {
			return fmt.Errorf("at %s: %w", decl.GetLineno(), err)
		}
	}
	return nil
}

// ProcessDecl dispatches a single declaration to its handler.
func (d *DomainSetup) ProcessDecl(decl Node) error {
	name := DeclName(decl)
	if name == "definition" {
		xtracer.Trace("compiler.IvyDomainSetup.dispatch name=%s", name)
		//pp("sn=%d lineno=%v", decl.(*ast.DefinitionDecl).Sn, decl.GetLineno())
	} else {
		xtracer.Trace("compiler.IvyDomainSetup.dispatch name=%s", name)
		//pp("goType=%T", decl)
	}
	switch n := decl.(type) {
	case *TypeDecl:
		for _, arg := range n.DeclArgs {
			if err := d.TypeDecl(arg); err != nil {
				return err
			}
		}
	case *AxiomDecl:
		for _, arg := range n.DeclArgs {
			if err := d.Axiom(arg); err != nil {
				return err
			}
		}
	case *PropertyDecl:
		for _, arg := range n.DeclArgs {
			if err := d.Property(arg); err != nil {
				return err
			}
		}
	case *ConjectureDecl:
		for _, arg := range n.DeclArgs {
			if err := d.Conjecture(arg); err != nil {
				return err
			}
		}
	case *RelationDecl:
		for _, arg := range n.DeclArgs {
			if err := d.Relation(arg); err != nil {
				return err
			}
		}
	case *ConstantDecl:
		for _, arg := range n.DeclArgs {
			if err := d.Individual(arg); err != nil {
				return err
			}
		}
	case *DerivedDecl:
		for _, arg := range n.DeclArgs {
			if err := d.Derived(arg); err != nil {
				return err
			}
		}
	case *DefinitionDecl:
		for _, arg := range n.DeclArgs {
			if err := d.DefinitionDecl(arg); err != nil {
				return err
			}
		}
	case *ActionDecl:
		for _, arg := range n.DeclArgs {
			if err := d.Action(arg); err != nil {
				return err
			}
		}
	// NOTE: InitDecl is NOT handled in pass 1 (DomainSetup).
	// Python's IvyDomainSetup does not have an 'init' method.
	// Init is handled in pass 3 (ARGSetup), matching Python's IvyARGSetup.init.
	// Do the Same for:
	// ast.ExportDecl, ast.ImportDecl, ast.IsolateDecl,
	// ast.DelegateDecl, ast.NativeDecl, ast.AttributeDecl,
	// ast.PrivateDecl, ast.AssertDecl
	case *InitDecl: // skip — handled in ARGSetup (pass 3)
	case *ExportDecl: // skip — handled in ARGSetup (pass 3)
	case *ImportDecl: // skip — handled in ARGSetup (pass 3)
	case *IsolateDecl: // skip — handled in ARGSetup (pass 3)
	case *DelegateDecl: // skip — handled in ARGSetup (pass 3)
	case *NativeDecl: // skip — handled in ARGSetup (pass 3)
	case *AttributeDecl: // skip — handled in ARGSetup (pass 3)
	case *PrivateDecl: // skip — handled in ARGSetup (pass 3)
	case *AssertDecl: // skip — handled in ARGSetup (pass 3)

	case *ObjectDecl:
		for _, arg := range n.DeclArgs {
			if err := d.Object(arg); err != nil {
				return err
			}
		}
	case *VariantDecl:
		for _, arg := range n.DeclArgs {
			if err := d.Variant(arg); err != nil {
				return err
			}
		}
	case *InterpretDecl:
		for _, arg := range n.DeclArgs {
			if err := d.Interpret(arg); err != nil {
				return err
			}
		}
	case *MixinDecl:
		for _, arg := range n.DeclArgs {
			if err := d.Mixin(arg); err != nil {
				return err
			}
		}
	case *AliasDecl:
		for _, arg := range n.DeclArgs {
			if err := d.Alias(arg); err != nil {
				return err
			}
		}
	case *ProgressDecl:
		for _, arg := range n.DeclArgs {
			if err := d.Progress(arg); err != nil {
				return err
			}
		}
	case *SchemaDecl:
		for _, arg := range n.DeclArgs {
			if err := d.Schema(arg); err != nil {
				return err
			}
		}
	case *InstantiateDecl:
		for _, arg := range n.DeclArgs {
			if err := d.Instantiate(arg); err != nil {
				return err
			}
		}
	case *ProofDecl:
		for _, arg := range n.DeclArgs {
			if err := d.Proof(arg); err != nil {
				return err
			}
		}
	case *NamedDecl:
		for _, arg := range n.DeclArgs {
			if err := d.Named(arg); err != nil {
				return err
			}
		}
	case *TheoremDecl:
		for _, arg := range n.DeclArgs {
			if err := d.Theorem(arg); err != nil {
				return err
			}
		}
	case *ParameterDecl:
		for _, arg := range n.DeclArgs {
			if err := d.Parameter(arg); err != nil {
				return err
			}
		}
	case *DestructorDecl:
		for _, arg := range n.DeclArgs {
			if err := d.Destructor(arg); err != nil {
				return err
			}
		}
	case *ConstructorDecl:
		for _, arg := range n.DeclArgs {
			if err := d.Constructor(arg); err != nil {
				return err
			}
		}
	case *ConceptDecl:
		for _, arg := range n.DeclArgs {
			if err := d.Concept(arg); err != nil {
				return err
			}
		}
	case *RelyDecl:
		for _, arg := range n.DeclArgs {
			if err := d.Rely(arg); err != nil {
				return err
			}
		}
	case *MixOrdDecl:
		for _, arg := range n.DeclArgs {
			if err := d.Mixord(arg); err != nil {
				return err
			}
		}
	case *UpdateDecl:
		for _, arg := range n.DeclArgs {
			if err := d.Update(arg); err != nil {
				return err
			}
		}
	case *ScenarioDecl:
		for _, arg := range n.DeclArgs {
			if err := d.Scenario(arg); err != nil {
				return err
			}
		}
	case *ImplementTypeDecl:
		for _, arg := range n.DeclArgs {
			if err := d.Implementtype(arg); err != nil {
				return err
			}
		}
	default:
		// Unknown declaration type: skip with no error
	}
	return nil
}

// collectASTVariables recursively collects free Variable nodes from an AST tree.
// Variables bound by quantifiers (ForAll, Exists, Some, NamedBinder) are excluded.
// Corresponds to Python's variables_ast (ivy_logic_utils.py:461-473).
func collectASTVariables(node Node) []*Variable {
	if node == nil {
		return nil
	}
	if v, ok := node.(*Variable); ok {
		return []*Variable{v}
	}
	// Handle binder nodes: exclude bound variables from results
	bounds, bodyArgs := astBinderInfo(node)
	if bounds != nil {
		var result []*Variable
		for _, arg := range bodyArgs {
			for _, v := range collectASTVariables(arg) {
				if !bounds[v.Rep] {
					result = append(result, v)
				}
			}
		}
		return result
	}
	// Non-binder: recurse into all args
	var result []*Variable
	for _, arg := range node.Args() {
		result = append(result, collectASTVariables(arg)...)
	}
	return result
}

// astBinderInfo returns the set of bound variable names and the body args
// for binder nodes (ForAll, Exists, Some, NamedBinder).
// Returns nil, nil for non-binder nodes.
// Corresponds to Python's binder_vars and binder_args (ivy_logic.py:640-649).
func astBinderInfo(node Node) (bounds map[string]bool, bodyArgs []Node) {
	switch n := node.(type) {
	case *Forall:
		bounds = astBoundNames(n.Bounds)
		return bounds, []Node{n.Body}
	case *Exists:
		bounds = astBoundNames(n.Bounds)
		return bounds, []Node{n.Body}
	case *Some:
		bounds = astBoundNames(n.Params)
		// Python: binder_args for Some returns args[1:] (the formula, not the params)
		return bounds, []Node{n.Fmla}
	case *NamedBinder:
		bounds = astBoundNames(n.Bounds)
		return bounds, []Node{n.Body}
	}
	return nil, nil
}

// astBoundNames extracts variable names from a list of bound variable nodes.
func astBoundNames(nodes []Node) map[string]bool {
	names := make(map[string]bool)
	for _, n := range nodes {
		if v, ok := n.(*Variable); ok {
			names[v.Rep] = true
		}
	}
	return names
}

// addDefinitionChecks validates that a compiled definition's LHS has no
// duplicate variables and that all RHS variables appear on the LHS.
// Corresponds to Python IvyDomainSetup.add_definition, which runs after
// compile_defn and checks lu.variables_ast/lu.used_variables_ast on the
// compiled ivy_logic.Definition.
func addDefinitionChecks(defn Expr) error {
	var lhs, rhs Expr
	switch d := defn.(type) {
	case *LogicDefinitionSchema:
		lhs = d.Lhs
		rhs = d.Rhs
	case *LogicDefinition:
		lhs = d.Lhs
		rhs = d.Rhs
	default:
		return nil
	}

	var lhsVars []*LogicVariable
	variablesAstOccurrencesRec(lhs, &lhsVars, nil)
	seen := make(map[NodeKey]bool)
	for _, v := range lhsVars {
		k := Key(v)
		if seen[k] {
			return NewIvyError(defn, fmt.Sprintf(
				"Variable %s occurs twice on left-hand side of definition", v))
		}
		seen[k] = true
	}

	rhsVars := VariablesAstList(rhs)
	for _, v := range rhsVars {
		if !seen[Key(v)] {
			return NewIvyError(defn, fmt.Sprintf(
				"Variable %s occurs free on right-hand side of definition", v))
		}
	}
	return nil
}

// AddDefinition routes a labeled definition to the appropriate module list and
// updates LastFact. Corresponds to the routing+last_fact portion of Python
// IvyDomainSetup.add_definition (ivy_compiler.py:1317-1327):
//
//	def add_definition(self, ldf):
//	    defs = self.domain.native_definitions
//	           if isinstance(ldf.formula.args[1], ivy_ast.NativeExpr)
//	           else self.domain.labeled_props
//	    # ... variable validation ...
//	    defs.append(ldf)
//	    self.last_fact = ldf
//
// Routing: if astDefNode.Rhs is a *ast.NativeExpr, the definition is appended
// to NativeDefinitions; otherwise to LabeledProps.
//
// The astDefNode parameter is the pre-compile AST Definition. It is passed
// explicitly because ldf.Formula by this point holds the compiled
// *lg.Definition (logic form), and the NativeExpr routing check needs the AST
// form. Callers (Derived, DefinitionDecl) already have the AST defNode in
// scope from earlier in their flow.
func (d *DomainSetup) AddDefinition(ldf *LabeledFormula, astDefNode *Definition) error {
	if compiled, ok := ldf.Formula.(Expr); ok {
		if err := addDefinitionChecks(compiled); err != nil {
			return err
		}
	}

	// Python (ivy_compiler.py:1318):
	//   defs = self.domain.native_definitions
	//          if isinstance(ldf.formula.args[1], ivy_ast.NativeExpr)
	//          else self.domain.labeled_props
	if _, isNative := astDefNode.Rhs.(*NativeExpr); isNative {
		d.Compiler.Module.NativeDefinitions = append(
			d.Compiler.Module.NativeDefinitions, ldf)
	} else {
		d.Compiler.Module.LabeledProps = append(
			d.Compiler.Module.LabeledProps, ldf)
	}
	d.LastFact = ldf
	return nil
}

// --- Individual declaration handlers ---

// TypeDecl processes a type declaration.
// Corresponds to Python IvyDomainSetup.typedef (ivy_compiler.py:1213-1247).
func (d *DomainSetup) TypeDecl(node Node) error {
	//xtracer.Trace("compiler.DomainSetup.type ENTER\n goType=%T", node)
	xtracer.Trace("compiler.DomainSetup.type ENTER")
	// Check for GhostTypeDef first — it embeds TypeDef, so *ast.TypeDef
	// assertion won't match it. Extract the inner TypeDef and mark as ghost.
	var td *TypeDef
	if gtd, ok := node.(*GhostTypeDef); ok {
		td = &gtd.TypeDef
		ghostName := compilerExtractSortRep(td.Name)
		if ghostName != "" {
			// Python: self.domain.ghost_sorts.add(typedef.name)
			d.Compiler.Module.GhostSorts[ghostName] = true
		}
	} else {
		var ok bool
		td, ok = node.(*TypeDef)
		if !ok {
			// Plain type declaration (no definition)
			if sym, ok := node.(*Symbol); ok {
				sort := &UninterpretedSort{Name: sym.Rep}
				xtracer.Trace("compiler.DomainSetup.type sort=UninterpretedSort name=%s ext=[]", sym.Rep)
				if err := d.Compiler.Sig.AddSort(sort); err != nil {
					// Sort already exists - not fatal
					return nil
				}
				d.Compiler.Module.SortOrder = append(d.Compiler.Module.SortOrder, sym.Rep)
				return nil
			}
			if atom, ok := node.(*Atom); ok {
				sort := &UninterpretedSort{Name: atom.Rep}
				xtracer.Trace("compiler.DomainSetup.type sort=UninterpretedSort name=%s ext=[]", atom.Rep)
				if err := d.Compiler.Sig.AddSort(sort); err != nil {
					return nil
				}
				d.Compiler.Module.SortOrder = append(d.Compiler.Module.SortOrder, atom.Rep)
				return nil
			}
			return nil
		}
	}

	// Type definition
	name := compilerExtractSortRep(td.Name)
	if name == "" {
		return NewIvyError(td, "type definition has no name")
	}

	switch v := td.Value.(type) {
	case *ConstantSort, *UninterpretedSortAST:
		sort := &UninterpretedSort{Name: name}
		xtracer.Trace("compiler.DomainSetup.type sort=UninterpretedSort name=%s ext=[]", name)
		if err := d.Compiler.Sig.AddSort(sort); err != nil {
			return nil
		}
		if td.Finite {
			d.Compiler.Module.FiniteSorts[name] = true
		}
	case *EnumeratedSort:
		ext := v.Extension()
		sort := &LogicEnumeratedSort{Name: name, Extension: ext}
		xtracer.Trace("compiler.DomainSetup.type sort=EnumeratedSort name=%s ext=%v", name, matchPythonStringSlice(ext))
		if err := d.Compiler.Sig.AddSort(sort); err != nil {
			return nil
		}
		// Add each enum value as a symbol
		// Python (ivy_compiler.py:1243-1247):
		//   self.domain.functions[sym] = 0
		//   self.domain.sig.constructors.add(sym)
		mod := d.Compiler.Module
		sig := d.Compiler.Sig
		for _, elemName := range ext {
			sym, err := d.Compiler.AddSymbol(elemName, sort, sig)
			_ = sym
			if err != nil {
				return err
			}
			xtracer.Trace("compiler.DomainSetup.type enum_constructor name=%s sort=%v", elemName, sort)
			//pp("sym=%v",  sym)
			mod.Functions.Set(Key(sym), sym.CSort)
			sig.Constructors[elemName] = true
		}
		if td.Finite {
			mod.FiniteSorts[name] = true
		}
	case *Range:
		lo := NumeralBound{Value: fmt.Sprint(v.Lo)}
		hi := NumeralBound{Value: fmt.Sprint(v.Hi)}
		sort := &RangeSort{Name: name, Lb: lo, Ub: hi}
		if err := d.Compiler.Sig.AddSort(sort); err != nil {
			return nil
		}
		if td.Finite {
			d.Compiler.Module.FiniteSorts[name] = true
		}
	case *StructSort:
		// Add the sort and its destructors
		// Corresponds to Python ivy_compiler.py:1225-1239
		sort := &UninterpretedSort{Name: name}
		xtracer.Trace("compiler.DomainSetup.type sort=struct name=%s", name)
		if err := d.Compiler.Sig.AddSort(sort); err != nil {
			return nil
		}
		if td.Finite {
			d.Compiler.Module.FiniteSorts[name] = true
		}
		// Python line 1229-1230: initialize empty destructor list for empty structs
		if _, exists := d.Compiler.Module.SortDestructors.Get2(name); !exists {
			d.Compiler.Module.SortDestructors.Set(name, []*Const{})
		}
		for _, field := range v.Fields {
			fieldName := NodeRep(field)
			var fieldSortNode Node
			switch f := field.(type) {
			case *Atom:
				fieldSortNode = f.ASort
			case *App:
				fieldSortNode = f.ASort
			default:
				continue
			}
			// Python line 1233-1234: validate field has a sort
			if fieldSortNode == nil {
				return NewIvyError(field, fmt.Sprintf("no sort provided for field %s", fieldName))
			}

			// Python: p = a.clone([Variable('V:dstr', sort.name)] + a.args);
			// p.sort = a.sort; self.destructor(p)
			args := append([]Node{&Variable{Rep: "V:dstr", VSort: name}}, field.Args()...)
			destrNode := field.Clone(args)
			switch destr := destrNode.(type) {
			case *Atom:
				destr.Rep = fieldName
				destr.ASort = fieldSortNode
			case *App:
				destr.Rep = &Symbol{Rep: fieldName}
				destr.ASort = fieldSortNode
			}
			if err := d.Destructor(destrNode); err != nil {
				return err
			}
		}
	default:
		// Uninterpreted sort
		sort := &UninterpretedSort{Name: name}
		if err := d.Compiler.Sig.AddSort(sort); err != nil {
			return nil
		}
	}

	d.Compiler.Module.SortOrder = append(d.Compiler.Module.SortOrder, name)
	return nil
}

func matchPythonStringSlice(slice []string) string {
	s := "["
	last := len(slice) - 1
	for i, e := range slice {
		if i != last {
			s += fmt.Sprintf("'%v', ", e)
		} else {
			s += fmt.Sprintf("'%v'", e)
		}
	}
	return s + "]"
}

// Axiom processes an axiom declaration.
func (d *DomainSetup) Axiom(node Node) error {
	xtracer.Trace("compiler.DomainSetup.axiom ENTER")
	lf, ok := node.(*LabeledFormula)
	if !ok {
		return nil
	}
	// Python: cax = ax.compile() — goes through thing() → LF.cmpl.
	cax, err := d.Compiler.ThingLF(lf)
	if err != nil {
		return err
	}

	// Python: if isinstance(cax.formula, SchemaBody): self.domain.schemata[cax.label.relname] = cax
	if _, ok := cax.Formula.(*SchemaBody); ok {
		labelName := cax.LabelName()
		if labelName != "" {
			xtracer.Trace("compiler.DomainSetup.axiom schemata.insert key='%s' value=%s", labelName, cax.Canon())
			d.Compiler.Module.Schemata.Set(labelName, cax)
		}
	} else {
		d.Compiler.Module.LabeledAxioms = append(d.Compiler.Module.LabeledAxioms, cax)
	}
	return nil
}

// Property processes a property declaration.
// Matches Python IvyDomainSetup.property (ivy_compiler.py:1038-1041):
//
//	def property(self, ax):
//	    lf = ax.compile()
//	    self.domain.labeled_props.append(lf)
//	    self.last_fact = lf
func (d *DomainSetup) Property(node Node) error {
	xtracer.Trace("compiler.DomainSetup.property ENTER")
	lf, ok := node.(*LabeledFormula)
	if !ok {
		return nil
	}
	// Python: lf = ax.compile() — goes through thing() → LF.cmpl.
	clf, err := d.Compiler.ThingLF(lf)
	if err != nil {
		return err
	}
	d.Compiler.Module.LabeledProps = append(d.Compiler.Module.LabeledProps, clf)
	// Python: self.last_fact = lf — stores the compiled LabeledFormula.
	d.LastFact = clf
	return nil
}

// Conjecture processes a conjecture declaration.
// Conjecture in DomainSetup (Pass 1) just resets LastFact.
// Actual conjecture compilation happens in ConjSetup (Pass 2).
// Matches Python IvyDomainSetup.conjecture (ivy_compiler.py:1041-1042):
//
//	def conjecture(self, c):
//	    self.last_fact = None
func (d *DomainSetup) Conjecture(node Node) error {
	xtracer.Trace("compiler.DomainSetup.conjecture ENTER")
	d.LastFact = nil
	return nil
}

// Relation processes a relation declaration.
func (d *DomainSetup) Relation(node Node) error {
	xtracer.Trace("compiler.DomainSetup.relation ENTER")
	atom, ok := node.(*Atom)
	if !ok {
		return nil
	}

	var domSorts []Sort
	for _, arg := range atom.Terms {
		if v, ok := arg.(*Variable); ok {
			s, err := d.Compiler.variableSort(v)
			if err != nil {
				return err
			}
			domSorts = append(domSorts, s)
		}
	}

	sort := LogicRelationSort(domSorts)
	sym, err := d.Compiler.AddSymbol(atom.Rep, sort, d.Compiler.Sig)
	if err != nil {
		return err
	}
	// Python: self.domain.all_relations.append((sym, len(rel.args)))
	mod := d.Compiler.Module
	mod.AllRelations = append(mod.AllRelations, sym)
	mod.Relations.Set(Key(sym), sym.CSort)
	return nil
}

// Individual processes a constant (individual) declaration.
// Corresponds to Python IvyDomainSetup.individual (ivy_compiler.py:1103-1106).
func (d *DomainSetup) Individual(node Node) error {
	_, err := d.individual(node)
	return err
}

func (d *DomainSetup) individual(node Node) (*Const, error) {
	xtracer.Trace("compiler.DomainSetup.individual ENTER")
	sym, err := d.Compiler.CompileConst(node, d.Compiler.Sig)
	if err != nil {
		return nil, err
	}
	// Python: self.domain.functions[sym] = len(v.args)
	if sym != nil {
		d.Compiler.Module.Functions.Set(Key(sym), sym.CSort)
	}
	return sym, nil
}

// Derived processes a derived relation/function declaration.
// Corresponds to Python IvyDomainSetup.derived.
func (d *DomainSetup) Derived(node Node) error {
	xtracer.Trace("compiler.DomainSetup.derived ENTER")
	if xtracer.Enabled {
		d.Compiler.SigCheck("DomainSetup.derived")
	}
	lf, ok := node.(*LabeledFormula)
	if !ok {
		return nil
	}
	df := lf.Formula
	// Check for DefinitionSchema first (embeds Definition)
	var defNode *Definition
	var isSchema bool
	if ds, ok := df.(*DefinitionSchema); ok {
		defNode = &ds.Definition
		isSchema = true
	} else if dn, ok := df.(*Definition); ok {
		defNode = dn
	} else {
		return nil
	}
	lhs := defNode.Lhs
	lhsAtom, ok := lhs.(*Atom)
	if !ok {
		return nil
	}
	// Add a temporary symbol with top function sort
	sym, err := d.Compiler.AddSymbol(lhsAtom.Rep, TopFunctionSort(len(lhsAtom.Terms)), d.Compiler.Sig)
	if err != nil {
		return err
	}

	// Compile the definition
	var compiled Expr
	if isSchema {
		compiled, err = d.Compiler.CompileDefnSchema(d.Compiler.Module.Cfg.AstCfg.NewDefinitionSchema(*defNode))
	} else {
		compiled, err = d.Compiler.CompileDefn(defNode)
	}
	if err != nil {
		return err
	}

	// Remove the temporary symbol and re-add with inferred sort
	d.Compiler.Sig.Symbols.Delkey(sym.Name)
	if def, ok := compiled.(*IvyDefinition); ok {
		definesNode := def.Defines()
		if cnst, ok := definesNode.(*Const); ok {
			d.Compiler.AddSymbol(cnst.Name, cnst.CSort, d.Compiler.Sig)
		}
	}

	// Extract concretely-sorted symbol from compiled definition.
	// Python's DerivedUpdate(df) uses defn.args[0].rep which has concrete sorts
	// from compilation. The original `sym` still has TopFunctionSort.
	derivedSym := sym // fallback
	if def, ok := compiled.(*IvyDefinition); ok {
		if cnst, ok := def.Defines().(*Const); ok {
			derivedSym = cnst
		}
	}

	// Python: self.add_definition(ldf.clone([label, df]))
	// Clone the LabeledFormula with the compiled definition, preserving metadata.
	// AddDefinition validates the compiled definition variables and routes to
	// either NativeDefinitions or LabeledProps based on whether the AST RHS is a
	// NativeExpr.
	mlf := lf.Clone([]Node{lf.Label, compiled}).(*LabeledFormula)
	if err := d.AddDefinition(mlf, defNode); err != nil {
		return err
	}
	mod := d.Compiler.Module
	mod.SymbolOrder = append(mod.SymbolOrder, sym)

	// Python: self.domain.all_relations.append((sym, len(lhs.args)))
	// Python: self.domain.relations[sym] = len(lhs.args)
	mod.AllRelations = append(mod.AllRelations, sym)
	mod.Relations.Set(Key(sym), sym.CSort)

	// Python: self.domain.updates.append(DerivedUpdate(df))
	mod.Updates = append(mod.Updates,
		NewDerivedUpdate(derivedSym, compiled))

	return nil
}

// DefinitionDecl processes a definition declaration.
// Corresponds to Python IvyDomainSetup.definition.
func (d *DomainSetup) DefinitionDecl(node Node) error {
	xtracer.Trace("compiler.DomainSetup.definition ENTER")
	if xtracer.Enabled {
		d.Compiler.SigCheck("DomainSetup.definition")
	}
	lf, ok := node.(*LabeledFormula)
	if !ok {
		return nil
	}
	df := lf.Formula
	// Check for DefinitionSchema first
	var defNode *Definition
	var isSchemaD bool
	if ds, ok := df.(*DefinitionSchema); ok {
		defNode = &ds.Definition
		isSchemaD = true
	} else if dn, ok := df.(*Definition); ok {
		defNode = dn
	} else {
		return nil
	}
	var compiled Expr
	var err error
	if isSchemaD {
		compiled, err = d.Compiler.CompileDefnSchema(d.Compiler.Module.Cfg.AstCfg.NewDefinitionSchema(*defNode))
	} else {
		compiled, err = d.Compiler.CompileDefn(defNode)
	}
	if err != nil {
		return err
	}

	// Python: self.add_definition(ldf.clone([label, df]))
	// Clone the LabeledFormula with the compiled definition, preserving metadata.
	// AddDefinition validates the compiled definition variables and routes to
	// either NativeDefinitions or LabeledProps based on whether the AST RHS is a
	// NativeExpr.
	mlf := lf.Clone([]Node{lf.Label, compiled}).(*LabeledFormula)
	if err := d.AddDefinition(mlf, defNode); err != nil {
		return err
	}

	// Add the defined symbol if not already in the signature
	if def, ok := compiled.(*IvyDefinition); ok {
		definesNode := def.Defines()
		if cnst, ok := definesNode.(*Const); ok {
			if !d.Compiler.Sig.ContainsSymbol(cnst.Name, cnst.CSort) {
				d.Compiler.AddSymbol(cnst.Name, cnst.CSort, d.Compiler.Sig)
			}
			d.Compiler.Module.SymbolOrder = append(d.Compiler.Module.SymbolOrder, cnst)
		}
		// Python: self.domain.updates.append(DerivedUpdate(df))
		d.Compiler.Module.Updates = append(d.Compiler.Module.Updates,
			NewDerivedUpdate(def.Defines(), compiled))
	}
	return nil
}

// Action processes an action declaration in pass 1 (DomainSetup).
// Corresponds to Python IvyDomainSetup.action (ivy_compiler.py:1344-1370).
// In pass 1, Python only scans for ThunkAction instances to declare thunk types.
// Actual action compilation happens in pass 3 (ARGSetup).
func (d *DomainSetup) Action(node Node) error {
	xtracer.Trace("compiler.DomainSetup.action ENTER")
	actDef, ok := node.(*ActionDef)
	if !ok {
		return nil
	}
	// Python: for action in a.args[1].iter_subactions():
	//             if isinstance(action, ThunkAction): ...
	return iterASTSubactions(actDef.Body, func(sub Node) error {
		thunk, ok := sub.(*ThunkAction)
		if !ok {
			return nil
		}
		subtype := thunk.Label // Python: action.args[0]
		suptype := thunk.Sort  // Python: action.args[2]

		// Python: tdef = TypeDef(subtype, UninterpretedSort()); self.type(tdef)
		cfg := d.Compiler.Module.Cfg.AstCfg
		tdef := cfg.NewTypeDef(subtype, cfg.NewConstantSort())
		if err := d.TypeDecl(tdef); err != nil {
			return err
		}

		// Python: vdef = VariantDef(subtype, suptype); self.variant(vdef)
		vdef := cfg.NewVariantDef(subtype, suptype)
		if err := d.Variant(vdef); err != nil {
			return err
		}

		// Python: actname = compose_names(subtype.relname, 'run')
		subtypeAtom, ok := subtype.(*Atom)
		if !ok {
			return nil
		}
		actname := d.Compiler.Module.Cfg.IuCfg.ComposeNames(subtypeAtom.Relname(), "run")

		// Python: selfparam = Atom('fml:$self', []); selfparam.sort = subtype.relname
		selfparam := cfg.NewAtom("fml:$self")
		selfparam.ASort = cfg.NewAtom(subtypeAtom.Relname())

		// Python: orig_args = action.args[0].args + action.args[1].args
		var origArgs []Node
		if thunk.Label != nil {
			origArgs = append(origArgs, thunk.Label.Args()...)
		}
		if thunk.Action != nil {
			origArgs = append(origArgs, thunk.Action.Args()...)
		}

		// Python: top_context.actions[actname] = (orig_args + [selfparam], [], len(orig_args))
		allFormals := make([]Node, len(origArgs)+1)
		copy(allFormals, origArgs)
		allFormals[len(origArgs)] = selfparam
		if d.Compiler.TopCtx != nil {
			d.Compiler.TopCtx.Actions[actname] = &ActionInfo{
				FormalAST: allFormals,
				KeyPos:    len(origArgs),
			}
		}
		return nil
	})
}

// iterASTSubactions walks an AST action tree, calling fn for each node.
// This is the AST-level equivalent of Python's Action.iter_subactions().
func iterASTSubactions(node Node, fn func(Node) error) error {
	if node == nil {
		return nil
	}
	if err := fn(node); err != nil {
		return err
	}
	for _, child := range node.Args() {
		if err := iterASTSubactions(child, fn); err != nil {
			return err
		}
	}
	return nil
}

// Init in pass 1 (DomainSetup) is a no-op.
// Python's IvyDomainSetup does NOT have an init method — init is
// only handled in pass 3 (ARGSetup). ProcessDecl already skips
// InitDecl via the "case *ast.InitDecl: // skip" branch.
// See ivy_compiler.py:1404-1413 (ARGSetup only).
func (d *DomainSetup) Init(node Node) error {
	return nil
}

// Object processes an object declaration.
func (d *DomainSetup) Object(node Node) error {
	xtracer.Trace("compiler.DomainSetup.object ENTER")
	if atom, ok := node.(*Atom); ok {
		d.Compiler.Module.AddObject(atom.Rep)
	}
	return nil
}

// Variant processes a variant declaration.
// Corresponds to Python IvyDomainSetup.variant (ivy_compiler.py:1248-1253).
// Python: variants[v.args[1].rep].append(sig.sorts[v.args[0].rep])
//
//	supertypes[v.args[0].rep] = sig.sorts[v.args[1].rep]
func (d *DomainSetup) Variant(node Node) error {
	xtracer.Trace("compiler.DomainSetup.variant ENTER")
	vd, ok := node.(*VariantDef)
	if !ok {
		return nil
	}
	sortName := compilerExtractSortRep(vd.Name)     // subtype (args[0] in Python)
	variantName := compilerExtractSortRep(vd.VSort) // supertype (args[1] in Python)
	if sortName == "" || variantName == "" {
		return nil
	}

	// Validate both sorts exist (Python: if r.rep not in self.domain.sig.sorts)
	subtypeSort, err := d.Compiler.Sig.FindSort(sortName, false)
	if err != nil {
		return NewIvyError(vd, fmt.Sprintf("undefined sort: %s", sortName))
	}
	supertypeSort, err := d.Compiler.Sig.FindSort(variantName, false)
	if err != nil {
		return NewIvyError(vd, fmt.Sprintf("undefined sort: %s", variantName))
	}

	// variants[supertype] ← append subtype sort
	// Python: self.domain.variants[v.args[1].rep].append(self.domain.sig.sorts[v.args[0].rep])
	d.Compiler.Module.Variants[variantName] = append(
		d.Compiler.Module.Variants[variantName], subtypeSort)

	d.Compiler.Module.Supertypes.Set(sortName, supertypeSort)

	return nil
}

// Export processes an export declaration in pass 1 (DomainSetup).
// Python's IvyDomainSetup does NOT have an export method — exports
// are only handled in pass 3 (ARGSetup). This is a no-op.
func (d *DomainSetup) Export(node Node) error {
	xtracer.Trace("compiler.DomainSetup.export ENTER")
	return nil
}

// Import processes an import declaration in pass 1 (DomainSetup).
// Python's IvyDomainSetup does NOT have an import method — imports
// are only handled in pass 3 (ARGSetup). This is a no-op.
func (d *DomainSetup) Import(node Node) error {
	xtracer.Trace("compiler.DomainSetup.import ENTER")
	return nil
}

// Isolate in pass 1 (DomainSetup) is a no-op.
// Python's IvyDomainSetup does NOT have an isolate method — isolates
// are only handled in pass 3 (ARGSetup). See ivy_compiler.py:1429-1434.
func (d *DomainSetup) Isolate(node Node) error {
	xtracer.Trace("compiler.DomainSetup.isolate ENTER")
	return nil
}

// Interpret processes a type interpretation.
// Faithful port of Python IvyDomainSetup.interpret (ivy_compiler.py:1333-1399).
func (d *DomainSetup) Interpret(node Node) error {
	xtracer.Trace("compiler.DomainSetup.interpret ENTER")

	// BB0: Extract lhs and rhs from the Implies formula inside the LabeledFormula.
	// Python: thing.formula is an Implies; .args[0] = lhs, .args[1] = rhs
	lf, ok := node.(*LabeledFormula)
	if !ok {
		return nil
	}
	impl, ok := lf.Formula.(*Implies)
	if !ok {
		return nil
	}
	sig := d.Compiler.Sig
	mod := d.Compiler.Module
	interp := sig.Interp

	// Python: lhs = resolve_alias(thing.formula.args[0].rep)
	lhs := ResolveAlias(compilerExtractSortRep(impl.T1), mod)
	// Python: rhs = thing.formula.args[1]  (the AST node)
	rhs := impl.T2

	// Python: xtracer.trace("compiler.DomainSetup.interpret lhs=%s rhs=%s\n rhsType=%s" % (lhs, type(rhs).__name__, type(rhs).__name__))
	rhsTypeName := astTypeName(rhs)
	xtracer.Trace("compiler.DomainSetup.interpret lhs=%s rhs=%s", lhs, rhsTypeName)
	//pp("rhsType=%s", rhsTypeName)

	// BB1: Handle native type interpretation
	// Python: if isinstance(thing.formula.args[1], ivy_ast.NativeType):
	if nt, ok := rhs.(*NativeType); ok {
		xtracer.Trace("compiler.DomainSetup.interpret branch=nativeType")
		// Python: if lhs in interp or lhs in self.domain.native_types:
		if _, exists := interp[lhs]; exists {
			return NewIvyError(node, fmt.Sprintf("%s is already interpreted", lhs))
		}
		if _, exists := mod.NativeTypes[lhs]; exists {
			return NewIvyError(node, fmt.Sprintf("%s is already interpreted", lhs))
		}
		// Python: self.domain.native_types[lhs] = compile_native_type(thing.formula.args[1])
		mod.NativeTypes[lhs] = compileNativeType(nt, mod)
		// Python: if thing.formula.args[1].args[0].code.strip() == 'int':
		//             compile_theory(self.domain, lhs, 'int')
		if len(nt.Elems) > 0 {
			isInt := false
			if atom, ok := nt.Elems[0].(*Atom); ok && atom.Rep == "int" {
				isInt = true
			}
			if nc, ok := nt.Elems[0].(*NativeCode); ok && strings.TrimSpace(nc.Code) == "int" {
				isInt = true
			}
			if isInt {
				xtracer.Trace("compiler.DomainSetup.interpret branch=nativeType-int")
				if err := d.Compiler.CompileTheory(lhs, "int"); err != nil {
					return err
				}
			}
		}
		xtracer.Trace("compiler.DomainSetup.interpret return=nativeType")
		return nil
	}

	// BB2: Non-native path
	// Python: rhs = thing.formula.args[1].rep
	// For Atom/Symbol, .rep is a string. For Range/EnumeratedSort, .rep returns self.
	// In Go, we keep rhs as ast.Node and extract the string name when needed.
	var rhsName string
	switch rhs.(type) {
	case *Range:
		// rhsName stays empty; handled in BB4 below
	case *EnumeratedSort:
		// rhsName stays empty; handled in BB5 below
	default:
		rhsName = compilerExtractSortRep(rhs)
	}
	xtracer.Trace("compiler.DomainSetup.interpret branch=non-native rhsName=%s", rhsName)

	// Python: self.domain.interps[lhs].append(thing)
	mod.Interps[lhs] = append(mod.Interps[lhs], node)

	// Python: if lhs in self.domain.native_types: raise IvyError(...)
	if _, exists := mod.NativeTypes[lhs]; exists {
		return NewIvyError(node, fmt.Sprintf("%s is already interpreted", lhs))
	}

	// BB3: Already interpreted check
	// Python: if lhs in interp:
	if existing, exists := interp[lhs]; exists {
		xtracer.Trace("compiler.DomainSetup.interpret branch=already-interpreted")
		// Python: if interp[lhs] != rhs: raise IvyError(...)
		if existingStr, ok := existing.(string); ok && existingStr == rhsName {
			xtracer.Trace("compiler.DomainSetup.interpret return=already-interpreted")
			return nil
		}
		return NewIvyError(node, fmt.Sprintf("%s is already interpreted", lhs))
	}

	// BB4: Range interpretation
	// Python: if isinstance(rhs, ivy_ast.Range):
	if rng, ok := rhs.(*Range); ok {
		xtracer.Trace("compiler.DomainSetup.interpret branch=range")
		// Python: if lhs not in sig.sorts: raise IvyError(...)
		if _, exists := sig.Sorts.Get2(lhs); !exists {
			return NewIvyError(node, fmt.Sprintf("%s is not a sort", lhs))
		}
		sort := sig.Sorts.Get(lhs)
		// Python: if not isinstance(sort, ivy_logic.UninterpretedSort): raise IvyError(...)
		if _, isUn := sort.(*UninterpretedSort); !isUn {
			return NewIvyError(node, fmt.Sprintf("%s is already interpreted", lhs))
		}
		// Python: compile_bound for lo and hi
		lo := d.compileBound(rng.Lo, lhs, sort, node)
		hi := d.compileBound(rng.Hi, lhs, sort, node)
		rangeSort := &RangeSort{Name: lhs, Lb: lo, Ub: hi}
		interp[lhs] = rangeSort
		// Python: compile_theory(self.domain, lhs, interp[lhs])
		// get_theory_schemata maps RangeSort → "int"
		if err := d.Compiler.CompileTheory(lhs, rangeSort); err != nil {
			return err
		}
		xtracer.Trace("compiler.DomainSetup.interpret return=range")
		return nil
	}

	// BB5: Enumerated sort interpretation
	// Python: if isinstance(rhs, ivy_ast.EnumeratedSort):
	if enumSort, ok := rhs.(*EnumeratedSort); ok {
		xtracer.Trace("compiler.DomainSetup.interpret branch=enum")
		// Python: if lhs not in self.domain.sig.sorts: raise IvyError(...)
		if _, exists := sig.Sorts.Get2(lhs); !exists {
			return NewIvyError(node, fmt.Sprintf("%s is not a type", lhs))
		}
		ext := enumSort.Extension()
		sort := &LogicEnumeratedSort{Name: lhs, Extension: ext}
		interp[lhs] = sort
		// Python: for c in sort.defines(): register constructors
		for _, c := range ext {
			if existingSort, hasSig := sig.Sorts.Get2(lhs); hasSig {
				sym := NewConst(c, existingSort)
				sig.Symbols.Set(c, &SymbolEntry{Sort: existingSort})
				mod.Functions.Set(Key(sym), sym.CSort)
				sig.Constructors[sym.Name] = true
			}
		}
		xtracer.Trace("compiler.DomainSetup.interpret return=enum")
		return nil
	}

	// BB6: Solver sort/symbol interpretation
	// Python: for x,y,z in zip([sig.sorts,sig.symbols], [is_solver_sort,is_solver_op], ['sort','symbol']):
	_, inSorts := sig.Sorts.Get2(lhs)
	_, inSymbols := sig.Symbols.Get2(lhs)

	// BB6a: Check sorts first
	if inSorts {
		xtracer.Trace("compiler.DomainSetup.interpret branch=solver-sort")
		// Python: if not solver.is_solver_sort(rhs): raise IvyError(...)
		if !IsSolverSort(rhsName) {
			return NewIvyError(node, fmt.Sprintf("%s not a native sort", rhsName))
		}
		interp[lhs] = rhsName
		// Python: if z == 'sort' and isinstance(rhs, str): compile_theory(...)
		if err := d.Compiler.CompileTheory(lhs, rhsName); err != nil {
			return err
		}
		xtracer.Trace("compiler.DomainSetup.interpret return=solver-sort")
		return nil
	}

	// BB6b: Check symbols
	if inSymbols {
		xtracer.Trace("compiler.DomainSetup.interpret branch=solver-symbol")
		// Python: if not solver.is_solver_op(rhs): raise IvyError(...)
		if !IsSolverOp(rhsName) {
			return NewIvyError(node, fmt.Sprintf("%s not a native symbol", rhsName))
		}
		interp[lhs] = rhsName
		xtracer.Trace("compiler.DomainSetup.interpret return=solver-symbol")
		return nil
	}

	// BB7: Python: raise IvyUndefined(thing, lhs)
	xtracer.Trace("compiler.DomainSetup.interpret branch=undefined")
	return NewIvyError(node, fmt.Sprintf("%s undefined", lhs))
}

// compileBound compiles a range bound, returning either a NumeralBound
// or a CompiledBound. Matches Python ivy_compiler.py:1515-1523 compile_bound.
func (d *DomainSetup) compileBound(b Node, lhsName string, sort Sort, context Node) NumeralOrCompiledBound {
	if b == nil {
		return NumeralBound{Value: "0", Sort: sort}
	}
	rep := fmt.Sprint(b)
	// Python: if not ivy_logic.is_numeral_name(b.rep): b.sort = lhs; self.parameter(b)
	if !IsNumeralName(rep) {
		cfg := d.Compiler.Module.Cfg.AstCfg
		sortNode := cfg.NewAtom(lhsName)
		switch x := b.(type) {
		case *Atom:
			x.ASort = sortNode
		case *App:
			x.ASort = sortNode
		case *Symbol:
			x.Sort = sortNode
		}
		_ = d.Parameter(b)
	}
	// Python: with top_sort_as_default(): res = b.compile()
	tsDefault := TopSortAsDefault(d.Compiler.Sig)
	tsDefault.Enter()
	res, err := d.Compiler.Thing(b)
	tsDefault.Exit()
	if err != nil {
		return NumeralBound{Value: rep, Sort: sort}
	}
	// Python: with ASTContext(thing): res = sort_infer(res, sort)
	inferred, err := SortInfer(res, sort)
	if err != nil {
		return NumeralBound{Value: rep, Sort: sort}
	}
	// For numerals, preserve the literal representation downstream
	// (RangeSort.LbString/UbString relies on this).
	if IsNumeralName(rep) {
		return NumeralBound{Value: rep, Sort: sort}
	}
	return CompiledBound{Expr: inferred}
}

// Mixin processes a mixin declaration in pass 1 (DomainSetup).
// Corresponds to Python IvyDomainSetup.mixin (ivy_compiler.py:1340-1342).
// Pass 1 only validates the mixee exists; it does NOT store the mixin.
// Mixin storage happens in pass 3 (ARGSetup).
func (d *DomainSetup) Mixin(node Node) error {
	xtracer.Trace("compiler.DomainSetup.mixin ENTER")
	args := node.Args()
	if len(args) < 2 {
		return nil
	}
	var mixeeName string
	if atom, ok := args[1].(*Atom); ok {
		mixeeName = atom.Relname()
	} else {
		return nil
	}
	// Validate mixee: must be 'init' or a known action
	// Python: if m.args[1].relname != 'init' and m.args[1].relname not in top_context.actions:
	if mixeeName != "init" && d.Compiler.TopCtx != nil {
		if _, ok := d.Compiler.TopCtx.Actions[mixeeName]; !ok {
			return NewIvyError(node, fmt.Sprintf("unknown action: %s", mixeeName))
		}
	}
	// Do NOT store mixin here — that's ARGSetup's job (pass 3)
	return nil
}

// Delegate processes a delegate declaration in pass 1 (DomainSetup).
// Python's IvyDomainSetup does NOT have a delegate method — delegates
// are only handled in pass 3 (ARGSetup). This is a no-op.
func (d *DomainSetup) Delegate(node Node) error {
	xtracer.Trace("compiler.DomainSetup.delegate ENTER")
	return nil
}

// Native processes a native code declaration.
// Python only handles native in IvyARGSetup (pass 3), not IvyDomainSetup (pass 1).
// ARGSetup.ProcessDecls already compiles via CompileNativeDef and appends.
func (d *DomainSetup) Native(node Node) error {
	xtracer.Trace("compiler.DomainSetup.native ENTER")
	return nil
}

// Alias processes an alias declaration.
func (d *DomainSetup) Alias(node Node) error {
	xtracer.Trace("compiler.DomainSetup.alias ENTER")
	if def, ok := node.(*Definition); ok {
		aliasName := compilerExtractSortRep(def.Lhs)
		targetName := compilerExtractSortRep(def.Rhs)
		if aliasName != "" && targetName != "" {
			resolved := ResolveAlias(targetName, d.Compiler.Module)
			d.Compiler.Module.Aliases[aliasName] = resolved
		}
	}
	return nil
}

// DefinedAttributes is the set of valid attribute names.
// Python: defined_attributes (ivy_compiler.py:1022)
var DefinedAttributes = map[string]bool{
	"weight": true, "test": true, "check": true, "mc": true, "bmc": true,
	"method": true, "separate": true, "iterable": true, "cardinality": true,
	"radix": true, "override": true, "cppstd": true, "libspec": true,
	"macro_finder": true, "global_parameter": true, "complete": true,
}

// KnownLogics matches Python ivy_logic.logics.
var KnownLogics = map[string]bool{"epr": true, "qf": true, "fo": true}

// Attribute in pass 1 (DomainSetup) is a no-op.
// Python's IvyDomainSetup does NOT have an attribute method — attributes
// are only handled in pass 3 (ARGSetup). See ivy_compiler.py:1447-1461.
func (d *DomainSetup) Attribute(node Node) error {
	xtracer.Trace("compiler.DomainSetup.attribute ENTER")
	return nil
}

// Progress processes a progress declaration.
// Corresponds to Python IvyDomainSetup.progress (ivy_compiler.py:1197-1202).
func (d *DomainSetup) Progress(node Node) error {
	xtracer.Trace("compiler.DomainSetup.progress ENTER")
	args := node.Args()
	if len(args) >= 2 {
		rel := args[0]
		body := args[1]
		if atom, ok := rel.(*Atom); ok {
			// Python: add_symbol(rel.relname, get_relation_sort(sig, rel.args, df.args[1]))
			relArgs := atom.Args()
			relSort, err := d.Compiler.GetRelationSortWithTerm(relArgs, body)
			if err == nil {
				d.Compiler.Sig.AddSymbol(atom.Relname(), relSort)
			}
		}
	}
	compiled, err := d.Compiler.SortifyWithInference(node)
	if err != nil {
		return err
	}
	d.Compiler.Module.Progress = append(d.Compiler.Module.Progress, compiled)
	return nil
}

// Private in pass 1 (DomainSetup) is a no-op.
// Python's IvyDomainSetup does NOT have a private method — privates
// are only handled in pass 3 (ARGSetup). See ivy_compiler.py:1441-1442.
func (d *DomainSetup) Private(node Node) error {
	xtracer.Trace("compiler.DomainSetup.private ENTER")
	return nil
}

// Schema processes a schema declaration.
// Corresponds to Python IvyDomainSetup.schema.
func (d *DomainSetup) Schema(node Node) error {
	xtracer.Trace("compiler.DomainSetup.schema ENTER")
	if xtracer.Enabled {
		d.Compiler.SigCheck("DomainSetup.schema")
	}
	// Handle *ast.Schema directly (e.g. from theory compilation).
	// Python: schema(self, sch) accesses sch.defn.args[1] and compiles
	// it if it's a SchemaBody. We must do the same — not just store raw.
	if schema, ok := node.(*Schema); ok {
		defn := schema.Defn.(*Definition)
		// Check if RHS is SchemaBody — if so, compile it
		// Python: if isinstance(sch.defn.args[1], ivy_ast.SchemaBody):
		//   ldf = ivy_ast.LabeledFormula(label, sch.defn.args[1].compile())
		// Note: Python's SchemaBody.compile = compile_schema_body (not thing())
		if sb, ok := defn.Rhs.(*SchemaBody); ok {
			compiled, err := d.Compiler.CompileSchemaBody(sb)
			if err != nil {
				return err
			}
			defName := defn.Defines()
			cfg := d.Compiler.Module.Cfg.AstCfg
			label := cfg.NewAtom(defName)
			clf := cfg.NewLabeledFormula(label, compiled)
			clf.SetLineno(schema.GetLineno())
			xtracer.Trace("compiler.DomainSetup.schema.body schemata.insert key='%s' value=%s", label.Rep, clf.Canon())
			d.Compiler.Module.Schemata.Set(label.Rep, clf)
		} else {
			xtracer.Trace("compiler.DomainSetup.schema.nonBody schemata.insert key='%s' value=%s", schema.Defines(), schema.Canon())
			d.Compiler.Module.Schemata.Set(schema.Defines(), schema)
		}
		return nil
	}
	return nil
}

// Instantiate processes an instantiation declaration.
// Corresponds to Python IvyDomainSetup.instantiate.
func (d *DomainSetup) Instantiate(node Node) error {
	xtracer.Trace("compiler.DomainSetup.instantiate ENTER")
	// Instantiation applies a schema. Extract the prefix and inst name,
	// look up the schema, and apply it.
	inst, ok := node.(*Instantiation)
	if !ok {
		return nil
	}
	var instName string
	if inst.Sort != nil {
		if atom, ok := inst.Sort.(*Atom); ok {
			instName = atom.Relname()
		}
	}
	if instName == "" {
		return nil
	}

	schema, ok := d.Compiler.Module.Schemata.Get2(instName) // Schemata *iu.InsMap[string, ast.Node]
	if ok {
		if c, canOk := schema.(Canonizer); canOk {
			xtracer.Trace("compiler.DomainSetup.instantiate schemata.lookup key='%s' found=true value=%s", instName, c.Canon())
		} else {
			xtracer.Trace("compiler.DomainSetup.instantiate schemata.lookup key='%s' found=true value=%v", instName, schema) // fallback. should not need?
		}
	} else {
		xtracer.Trace("compiler.DomainSetup.instantiate schemata.lookup key='%s' found=false", instName)
		return NewIvyError(inst, fmt.Sprintf("%s undefined in instantiation", instName))
	}

	// Python applies top-level schema instantiations immediately:
	// self.domain.schemata[inst.relname].instantiate(inst.args)
	sch, ok := schema.(*Schema)
	if !ok {
		return NewIvyError(inst, fmt.Sprintf("%s is not an instantiable schema", instName))
	}
	instAtom, ok := inst.Sort.(*Atom)
	if !ok {
		return NewIvyError(inst, fmt.Sprintf("%s undefined in instantiation", instName))
	}
	fmla, err := sch.GetInstance(instAtom.Terms, d.Compiler, nil, false)
	if err != nil {
		return NewIvyError(inst, "wrong number of parameters in instantiation")
	}
	sch.Instances = append(sch.Instances, fmla)
	return nil
}

// Proof processes a proof declaration.
// Corresponds to Python IvyDomainSetup.proof.
func (d *DomainSetup) Proof(node Node) error {
	xtracer.Trace("compiler.DomainSetup.proof ENTER")
	// If the proof is a labeled formula, it has its own label.
	if lf, ok := node.(*LabeledFormula); ok {
		// Compile the proof body as a tactic (not as a logic node).
		compiledProof, err := d.Compiler.CompileTactic(lf.Formula)
		if err != nil {
			return err
		}
		acfg := d.Compiler.Module.Cfg.AstCfg
		proofLF := acfg.NewLabeledFormula(nil, nil)
		if lf.Label != nil {
			label, _ := d.Compiler.Thing(lf.Label)
			proofLF.Label = label
		}
		d.Compiler.Module.Proofs = append(d.Compiler.Module.Proofs, ProofEntry{
			Formula: proofLF,
			Proof:   compiledProof,
		})
		return nil
	}

	// Otherwise attach to the last fact.
	if d.LastFact == nil {
		return nil // this is a conjecture, skip
	}

	// Set last_fmla for schema instantiation context (Python: last_fmla = self.last_fact.formula)
	// Compile as a tactic rather than a logic node.
	compiled, err := d.Compiler.CompileTactic(node)
	if err != nil {
		return err
	}

	// Python: self.domain.proofs.append((self.last_fact, pf.compile()))
	d.Compiler.Module.Proofs = append(d.Compiler.Module.Proofs, ProofEntry{
		Formula: d.LastFact,
		Proof:   compiled,
	})
	return nil
}

// Named processes a named declaration.
// Corresponds to Python IvyDomainSetup.named.
// A named declaration gives a name to an existential property.
func (d *DomainSetup) Named(node Node) error {
	xtracer.Trace("compiler.DomainSetup.named ENTER")
	lhs, ok := node.(*Atom)
	if !ok {
		return nil
	}
	if d.LastFact == nil {
		return NewIvyError(node, "named declaration without preceding property")
	}

	// The last fact should be an existential formula.
	// We extract the existential's range sort and create a function symbol.
	// Python: cond = ivy_logic.drop_universals(self.last_fact.formula)
	lastFormula, ok := d.LastFact.Formula.(Expr)
	if !ok {
		return NewIvyError(node, "named declaration without preceding property")
	}
	cond := IvyDropUniversals(lastFormula)
	if !IsExists(cond) {
		return NewIvyError(node, "property is not existential")
	}

	// Get the existential variable's sort as the range
	ex, ok := cond.(*LogicExists)
	if !ok || len(ex.Variables) != 1 {
		return NewIvyError(node, "property is not existential")
	}
	rng := ex.Variables[0].VSort

	// Build domain sorts from the lhs parameters
	var domSorts []Sort
	for _, arg := range lhs.Terms {
		compiled, err := d.Compiler.Thing(arg)
		if err != nil {
			return err
		}
		domSorts = append(domSorts, compiled.NodeSort())
	}

	sort := FuncConstSort(append(domSorts, rng)...)
	sym, err := d.Compiler.AddSymbol(lhs.Rep, sort, d.Compiler.Sig)
	if err != nil {
		return err
	}

	// Python: self.domain.named.append((self.last_fact, sym(...)))
	d.Compiler.Module.Named = append(d.Compiler.Module.Named, NamedEntry{
		Formula: d.LastFact,
		Name:    sym,
	})

	// Python (ivy_compiler.py:1237): self.domain.updates.append(NamedUpdate(sym, cond))
	d.Compiler.Module.Updates = append(d.Compiler.Module.Updates,
		NewNamedUpdate(sym, cond))
	return nil
}

// Theorem processes a theorem declaration.
// Corresponds to Python IvyDomainSetup.theorem.
func (d *DomainSetup) Theorem(node Node) error {
	xtracer.Trace("compiler.DomainSetup.theorem ENTER")
	schema, ok := node.(*Schema)
	if !ok || schema == nil {
		return nil
	}
	df, ok := schema.Defn.(*Definition)
	if !ok || df == nil {
		return nil
	}
	if sb, ok := df.Rhs.(*SchemaBody); ok {
		compiled, err := d.Compiler.CompileSchemaBody(sb)
		if err != nil {
			return err
		}
		acfg := d.Compiler.Module.Cfg.AstCfg
		defName := df.Defines()
		label := acfg.NewAtom(defName)
		mlf := acfg.NewLabeledFormula(label, compiled)
		mlf.SetLineno(sb.GetLineno())
		d.Compiler.Module.LabeledProps = append(d.Compiler.Module.LabeledProps, mlf)
		d.Compiler.Module.Theorems[defName] = compiled
		d.LastFact = mlf
	}
	return nil
}

// Assert in pass 1 (DomainSetup) is a no-op.
// Python's IvyDomainSetup does NOT have an _assert method — asserts
// are only handled in pass 3 (ARGSetup). See ivy_compiler.py:1426-1428.
func (d *DomainSetup) Assert(node Node) error {
	return nil
}

// Parameter processes a parameter declaration.
// Corresponds to Python IvyDomainSetup.parameter (ivy_compiler.py:1108).
func (d *DomainSetup) Parameter(node Node) error {
	xtracer.Trace("compiler.DomainSetup.parameter ENTER")
	mod := d.Compiler.Module
	var sym *Const
	var dflt Node // raw AST node, matching Python
	if def, ok := node.(*Definition); ok {
		var err error
		sym, err = d.individual(def.Lhs)
		if err != nil {
			return err
		}
		dflt = def.Rhs // Python stores raw AST node, not string
	} else {
		var err error
		sym, err = d.individual(node)
		if err != nil {
			return err
		}
		dflt = nil
	}
	mod.Params = append(mod.Params, sym)
	mod.ParamDefaults = append(mod.ParamDefaults, dflt)
	return nil
}

// Destructor processes a destructor declaration.
// Corresponds to Python IvyDomainSetup.destructor (ivy_compiler.py:1118).
func (d *DomainSetup) Destructor(node Node) error {
	xtracer.Trace("compiler.DomainSetup.destructor ENTER")
	mod := d.Compiler.Module
	sym, err := d.individual(node)
	if err != nil {
		return err
	}
	dom := SortDomain(sym.CSort)
	if len(dom) == 0 {
		return NewIvyError(node, "A destructor must have at least one parameter")
	}
	mod.DestructorSorts[sym.Name] = dom[0]
	sortName := IvySortName(dom[0])
	destrs, _ := mod.SortDestructors.Get2(sortName)
	mod.SortDestructors.Set(sortName, append(destrs, sym))
	return nil
}

// Constructor processes a constructor declaration.
// Corresponds to Python IvyDomainSetup.constructor (ivy_compiler.py:1125).
func (d *DomainSetup) Constructor(node Node) error {
	xtracer.Trace("compiler.DomainSetup.constructor ENTER")
	mod := d.Compiler.Module
	sym, err := d.Compiler.CompileConst(node, d.Compiler.Sig)
	if err != nil {
		return err
	}
	rng := SortRange(sym.CSort)
	mod.ConstructorSorts[sym.Name] = rng
	sortName := IvySortName(rng)
	conss, _ := mod.SortConstructors.Get2(sortName)
	mod.SortConstructors.Set(sortName, append(conss, sym))
	return nil
}

// Concept processes a concept declaration.
// Corresponds to Python IvyDomainSetup.concept (ivy_compiler.py:1208).
func (d *DomainSetup) Concept(node Node) error {
	xtracer.Trace("compiler.DomainSetup.concept ENTER")
	mod := d.Compiler.Module
	lf, ok := node.(*LabeledFormula)
	if !ok {
		return nil
	}
	// lf.Label is the relation atom, lf.Formula is the body
	// Python: rel = c.args[0]; add_symbol(rel.relname, get_relation_sort(sig, rel.args, c.args[1]))
	body := lf.Formula
	if atom, ok := lf.Label.(*Atom); ok {
		relSort, err := d.Compiler.GetRelationSortWithTerm(atom.Terms, body)
		if err != nil {
			return err
		}
		_, err = d.Compiler.AddSymbol(atom.Rep, relSort, d.Compiler.Sig)
		if err != nil {
			return err
		}
	}
	// Python: c = sortify_with_inference(c)
	// Python: self.domain.concept_spaces.append((c.args[0], c.args[1]))
	// Sortify the label and formula separately, then store the pair.
	var compiledLabel Expr
	if lf.Label != nil {
		var err error
		compiledLabel, err = d.Compiler.SortifyWithInference(lf.Label)
		if err != nil {
			return err
		}
	}
	var compiledBody Expr
	if lf.Formula != nil {
		var err error
		compiledBody, err = d.Compiler.SortifyWithInference(lf.Formula)
		if err != nil {
			return err
		}
	}
	mod.ConceptSpaces = append(mod.ConceptSpaces, ConceptSpace{Label: compiledLabel, Body: compiledBody})
	return nil
}

// Rely processes a rely declaration.
// Corresponds to Python IvyDomainSetup.rely (ivy_compiler.py:1203).
func (d *DomainSetup) Rely(node Node) error {
	xtracer.Trace("compiler.DomainSetup.rely ENTER")
	mod := d.Compiler.Module
	lf, ok := node.(*LabeledFormula)
	formula := node
	if ok {
		formula = lf.Formula
	}
	compiled, err := d.Compiler.SortifyWithInference(formula)
	if err != nil {
		return err
	}
	mod.Rely = append(mod.Rely, compiled)
	return nil
}

// Mixord processes a mixord declaration.
// Corresponds to Python IvyDomainSetup.mixord (ivy_compiler.py:1206).
func (d *DomainSetup) Mixord(node Node) error {
	xtracer.Trace("compiler.DomainSetup.mixord ENTER")
	d.Compiler.Module.MixOrd = append(d.Compiler.Module.MixOrd, node)
	return nil
}

// Update processes an update declaration.
// Corresponds to Python IvyDomainSetup.update (ivy_compiler.py:1214).
// Python: self.domain.updates.append(upd.compile())
func (d *DomainSetup) Update(node Node) error {
	xtracer.Trace("compiler.DomainSetup.update ENTER")
	mod := d.Compiler.Module
	compiled, err := d.Compiler.Thing(node)
	if err != nil {
		// If the node can't be compiled (e.g. unknown symbol), store raw node.
		// Python's upd.compile() delegates to the AST node's own compile method.
		mod.Updates = append(mod.Updates, node)
		return nil
	}
	mod.Updates = append(mod.Updates, compiled)
	return nil
}

// Scenario processes a scenario declaration.
// Corresponds to Python IvyDomainSetup.scenario (ivy_compiler.py:1333).
func (d *DomainSetup) Scenario(node Node) error {
	xtracer.Trace("compiler.DomainSetup.scenario ENTER")
	mod := d.Compiler.Module
	sig := d.Compiler.Sig
	scenDef, ok := node.(*ScenarioDef)
	if !ok {
		return nil
	}
	relSort := LogicRelationSort([]Sort{})
	for _, pi := range scenDef.Places() {
		sym, err := d.Compiler.AddSymbol(pi.Name, relSort, sig)
		if err != nil {
			return err
		}
		mod.AllRelations = append(mod.AllRelations, sym)
		mod.Relations.Set(Key(sym), sym.CSort)
	}
	return nil
}

// Implementtype processes an implement type declaration.
// Corresponds to Python IvyDomainSetup.implementtype (ivy_compiler.py:1254).
func (d *DomainSetup) Implementtype(node Node) error {
	xtracer.Trace("compiler.DomainSetup.implementtype ENTER")
	mod := d.Compiler.Module
	sig := d.Compiler.Sig
	lf, ok := node.(*LabeledFormula)
	if !ok {
		return nil
	}
	def, ok := lf.Formula.(*Definition)
	if !ok {
		return nil
	}
	impd := compilerExtractSortRep(def.Lhs)
	impr := compilerExtractSortRep(def.Rhs)
	// Validate both sorts exist
	if _, ok := sig.Sorts.Get2(impd); !ok {
		return NewIvyError(lf, fmt.Sprintf("undefined sort: %s", impd))
	}
	if _, ok := sig.Sorts.Get2(impr); !ok {
		return NewIvyError(lf, fmt.Sprintf("undefined sort: %s", impr))
	}
	// Check not already interpreted
	if _, ok := mod.NativeTypes[impd]; ok {
		return NewIvyError(lf, fmt.Sprintf("%s is already interpreted", impd))
	}
	if _, ok := sig.Interp[impd]; ok {
		return NewIvyError(lf, fmt.Sprintf("%s is already interpreted", impd))
	}
	impdSort := sig.Sorts.Get(impd)
	imprSort := sig.Sorts.Get(impr)
	ImplementType(sig, impdSort, imprSort)
	mod.Interps[impd] = append(mod.Interps[impd], node)
	return nil
}
