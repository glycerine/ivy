package compiler

import (
	"fmt"
	"strings"

	"github.com/glycerine/ivy/goivy/actions"
	"github.com/glycerine/ivy/goivy/ast"
	il "github.com/glycerine/ivy/goivy/ivylogic"
	iu "github.com/glycerine/ivy/goivy/ivyutils"
	lg "github.com/glycerine/ivy/goivy/logic"
	"github.com/glycerine/ivy/goivy/module"
	"github.com/glycerine/ivy/goivy/xtracer"
	solver "github.com/glycerine/ivy/goivy/z3bridge"
)

// DomainSetup processes top-level declarations, replacing Python's
// IvyDomainSetup class.
type DomainSetup struct {
	Compiler *Compiler

	// LastFact holds the last compiled property/axiom LabeledFormula, used by
	// proof and named declarations. Matches Python's self.last_fact.
	LastFact *ast.LabeledFormula
}

// NewDomainSetup creates a new declaration interpreter.
func NewDomainSetup(c *Compiler) *DomainSetup {
	return &DomainSetup{Compiler: c}
}

// ProcessDecls processes all declarations in a declaration list.
// Each declaration is dispatched to the appropriate handler based on its type.
func (d *DomainSetup) ProcessDecls(decls []ast.Node) error {
	for _, decl := range decls {
		if err := d.ProcessDecl(decl); err != nil {
			return fmt.Errorf("at %s: %w", decl.GetLineno(), err)
		}
	}
	return nil
}

// ProcessDecl dispatches a single declaration to its handler.
func (d *DomainSetup) ProcessDecl(decl ast.Node) error {
	name := ast.DeclName(decl)
	if name == "definition" {
		xtracer.Trace("compiler.IvyDomainSetup.dispatch name=%s", name)
		//pp("sn=%d lineno=%v", decl.(*ast.DefinitionDecl).Sn, decl.GetLineno())
	} else {
		xtracer.Trace("compiler.IvyDomainSetup.dispatch name=%s", name)
		//pp("goType=%T", decl)
	}
	switch n := decl.(type) {
	case *ast.TypeDecl:
		for _, arg := range n.DeclArgs {
			if err := d.TypeDecl(arg); err != nil {
				return err
			}
		}
	case *ast.AxiomDecl:
		for _, arg := range n.DeclArgs {
			if err := d.Axiom(arg); err != nil {
				return err
			}
		}
	case *ast.PropertyDecl:
		for _, arg := range n.DeclArgs {
			if err := d.Property(arg); err != nil {
				return err
			}
		}
	case *ast.ConjectureDecl:
		for _, arg := range n.DeclArgs {
			if err := d.Conjecture(arg); err != nil {
				return err
			}
		}
	case *ast.RelationDecl:
		for _, arg := range n.DeclArgs {
			if err := d.Relation(arg); err != nil {
				return err
			}
		}
	case *ast.ConstantDecl:
		for _, arg := range n.DeclArgs {
			if err := d.Individual(arg); err != nil {
				return err
			}
		}
	case *ast.DerivedDecl:
		for _, arg := range n.DeclArgs {
			if err := d.Derived(arg); err != nil {
				return err
			}
		}
	case *ast.DefinitionDecl:
		for _, arg := range n.DeclArgs {
			if err := d.DefinitionDecl(arg); err != nil {
				return err
			}
		}
	case *ast.ActionDecl:
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
	case *ast.InitDecl: // skip — handled in ARGSetup (pass 3)
	case *ast.ExportDecl: // skip — handled in ARGSetup (pass 3)
	case *ast.ImportDecl: // skip — handled in ARGSetup (pass 3)
	case *ast.IsolateDecl: // skip — handled in ARGSetup (pass 3)
	case *ast.DelegateDecl: // skip — handled in ARGSetup (pass 3)
	case *ast.NativeDecl: // skip — handled in ARGSetup (pass 3)
	case *ast.AttributeDecl: // skip — handled in ARGSetup (pass 3)
	case *ast.PrivateDecl: // skip — handled in ARGSetup (pass 3)
	case *ast.AssertDecl: // skip — handled in ARGSetup (pass 3)

	case *ast.ObjectDecl:
		for _, arg := range n.DeclArgs {
			if err := d.Object(arg); err != nil {
				return err
			}
		}
	case *ast.VariantDecl:
		for _, arg := range n.DeclArgs {
			if err := d.Variant(arg); err != nil {
				return err
			}
		}
	case *ast.InterpretDecl:
		for _, arg := range n.DeclArgs {
			if err := d.Interpret(arg); err != nil {
				return err
			}
		}
	case *ast.MixinDecl:
		for _, arg := range n.DeclArgs {
			if err := d.Mixin(arg); err != nil {
				return err
			}
		}
	case *ast.AliasDecl:
		for _, arg := range n.DeclArgs {
			if err := d.Alias(arg); err != nil {
				return err
			}
		}
	case *ast.ProgressDecl:
		for _, arg := range n.DeclArgs {
			if err := d.Progress(arg); err != nil {
				return err
			}
		}
	case *ast.SchemaDecl:
		for _, arg := range n.DeclArgs {
			if err := d.Schema(arg); err != nil {
				return err
			}
		}
	case *ast.InstantiateDecl:
		for _, arg := range n.DeclArgs {
			if err := d.Instantiate(arg); err != nil {
				return err
			}
		}
	case *ast.ProofDecl:
		for _, arg := range n.DeclArgs {
			if err := d.Proof(arg); err != nil {
				return err
			}
		}
	case *ast.NamedDecl:
		for _, arg := range n.DeclArgs {
			if err := d.Named(arg); err != nil {
				return err
			}
		}
	case *ast.TheoremDecl:
		for _, arg := range n.DeclArgs {
			if err := d.Theorem(arg); err != nil {
				return err
			}
		}
	case *ast.ParameterDecl:
		for _, arg := range n.DeclArgs {
			if err := d.Parameter(arg); err != nil {
				return err
			}
		}
	case *ast.DestructorDecl:
		for _, arg := range n.DeclArgs {
			if err := d.Destructor(arg); err != nil {
				return err
			}
		}
	case *ast.ConstructorDecl:
		for _, arg := range n.DeclArgs {
			if err := d.Constructor(arg); err != nil {
				return err
			}
		}
	case *ast.ConceptDecl:
		for _, arg := range n.DeclArgs {
			if err := d.Concept(arg); err != nil {
				return err
			}
		}
	case *ast.RelyDecl:
		for _, arg := range n.DeclArgs {
			if err := d.Rely(arg); err != nil {
				return err
			}
		}
	case *ast.MixOrdDecl:
		for _, arg := range n.DeclArgs {
			if err := d.Mixord(arg); err != nil {
				return err
			}
		}
	case *ast.UpdateDecl:
		for _, arg := range n.DeclArgs {
			if err := d.Update(arg); err != nil {
				return err
			}
		}
	case *ast.ScenarioDecl:
		for _, arg := range n.DeclArgs {
			if err := d.Scenario(arg); err != nil {
				return err
			}
		}
	case *ast.ImplementTypeDecl:
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
func collectASTVariables(node ast.Node) []*ast.Variable {
	if node == nil {
		return nil
	}
	if v, ok := node.(*ast.Variable); ok {
		return []*ast.Variable{v}
	}
	// Handle binder nodes: exclude bound variables from results
	bounds, bodyArgs := astBinderInfo(node)
	if bounds != nil {
		var result []*ast.Variable
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
	var result []*ast.Variable
	for _, arg := range node.Args() {
		result = append(result, collectASTVariables(arg)...)
	}
	return result
}

// astBinderInfo returns the set of bound variable names and the body args
// for binder nodes (ForAll, Exists, Some, NamedBinder).
// Returns nil, nil for non-binder nodes.
// Corresponds to Python's binder_vars and binder_args (ivy_logic.py:640-649).
func astBinderInfo(node ast.Node) (bounds map[string]bool, bodyArgs []ast.Node) {
	switch n := node.(type) {
	case *ast.Forall:
		bounds = astBoundNames(n.Bounds)
		return bounds, []ast.Node{n.Body}
	case *ast.Exists:
		bounds = astBoundNames(n.Bounds)
		return bounds, []ast.Node{n.Body}
	case *ast.Some:
		bounds = astBoundNames(n.Params)
		// Python: binder_args for Some returns args[1:] (the formula, not the params)
		return bounds, []ast.Node{n.Fmla}
	case *ast.NamedBinder:
		bounds = astBoundNames(n.Bounds)
		return bounds, []ast.Node{n.Body}
	}
	return nil, nil
}

// astBoundNames extracts variable names from a list of bound variable nodes.
func astBoundNames(nodes []ast.Node) map[string]bool {
	names := make(map[string]bool)
	for _, n := range nodes {
		if v, ok := n.(*ast.Variable); ok {
			names[v.Rep] = true
		}
	}
	return names
}

// addDefinitionChecks validates that a definition's LHS has no duplicate
// variables and that all RHS variables appear on the LHS.
// Corresponds to Python's add_definition checks (ivy_compiler.py:1143-1152).
func addDefinitionChecks(defNode *ast.Definition) error {
	lhsVars := collectASTVariables(defNode.Lhs)
	seen := make(map[string]bool)
	for _, v := range lhsVars {
		if seen[v.Rep] {
			return lg.NewIvyError(defNode, fmt.Sprintf(
				"Variable %s occurs twice on left-hand side of definition", v.Rep))
		}
		seen[v.Rep] = true
	}
	rhsVars := collectASTVariables(defNode.Rhs)
	for _, v := range rhsVars {
		if !seen[v.Rep] {
			return lg.NewIvyError(defNode, fmt.Sprintf(
				"Variable %s occurs free on right-hand side of definition", v.Rep))
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
//
// NOTE: Variable validation (addDefinitionChecks) is intentionally NOT done
// here. Validation must run BEFORE compile so a free RHS variable is reported
// as "occurs free on right-hand side of definition" rather than the
// compile-side "unknown symbol" error. Callers do the validation early,
// before invoking CompileDefn.
func (d *DomainSetup) AddDefinition(ldf *ast.LabeledFormula, astDefNode *ast.Definition) error {
	// Python (ivy_compiler.py:1318):
	//   defs = self.domain.native_definitions
	//          if isinstance(ldf.formula.args[1], ivy_ast.NativeExpr)
	//          else self.domain.labeled_props
	if _, isNative := astDefNode.Rhs.(*ast.NativeExpr); isNative {
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
func (d *DomainSetup) TypeDecl(node ast.Node) error {
	//xtracer.Trace("compiler.DomainSetup.type ENTER\n goType=%T", node)
	xtracer.Trace("compiler.DomainSetup.type ENTER")
	// Check for GhostTypeDef first — it embeds TypeDef, so *ast.TypeDef
	// assertion won't match it. Extract the inner TypeDef and mark as ghost.
	var td *ast.TypeDef
	if gtd, ok := node.(*ast.GhostTypeDef); ok {
		td = &gtd.TypeDef
		ghostName := extractSortRep(td.Name)
		if ghostName != "" {
			// Python: self.domain.ghost_sorts.add(typedef.name)
			d.Compiler.Module.GhostSorts[ghostName] = true
		}
	} else {
		var ok bool
		td, ok = node.(*ast.TypeDef)
		if !ok {
			// Plain type declaration (no definition)
			if sym, ok := node.(*ast.Symbol); ok {
				sort := &lg.UninterpretedSort{Name: sym.Rep}
				xtracer.Trace("compiler.DomainSetup.type sort=UninterpretedSort name=%s ext=[]", sym.Rep)
				if err := d.Compiler.Sig.AddSort(sort); err != nil {
					// Sort already exists - not fatal
					return nil
				}
				d.Compiler.Module.SortOrder = append(d.Compiler.Module.SortOrder, sym.Rep)
				return nil
			}
			if atom, ok := node.(*ast.Atom); ok {
				sort := &lg.UninterpretedSort{Name: atom.Rep}
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
	name := extractSortRep(td.Name)
	if name == "" {
		return lg.NewIvyError(td, "type definition has no name")
	}

	switch v := td.Value.(type) {
	case *ast.ConstantSort, *ast.UninterpretedSortAST:
		sort := &lg.UninterpretedSort{Name: name}
		xtracer.Trace("compiler.DomainSetup.type sort=UninterpretedSort name=%s ext=[]", name)
		if err := d.Compiler.Sig.AddSort(sort); err != nil {
			return nil
		}
	case *ast.EnumeratedSort:
		ext := v.Extension()
		sort := &lg.EnumeratedSort{Name: name, Extension: ext}
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
			mod.Functions.Set(elemName, sort)
			sig.Constructors[elemName] = true
		}
		if td.Finite {
			mod.FiniteSorts[name] = true
		}
	case *ast.Range:
		lo := lg.NumeralBound{Value: fmt.Sprint(v.Lo)}
		hi := lg.NumeralBound{Value: fmt.Sprint(v.Hi)}
		sort := &lg.RangeSort{Name: name, Lb: lo, Ub: hi}
		if err := d.Compiler.Sig.AddSort(sort); err != nil {
			return nil
		}
	case *ast.StructSort:
		// Add the sort and its destructors
		// Corresponds to Python ivy_compiler.py:1225-1239
		sort := &lg.UninterpretedSort{Name: name}
		if err := d.Compiler.Sig.AddSort(sort); err != nil {
			return nil
		}
		// Python line 1229-1230: initialize empty destructor list for empty structs
		if _, exists := d.Compiler.Module.SortDestructors[name]; !exists {
			d.Compiler.Module.SortDestructors[name] = []*lg.Const{}
		}
		for _, field := range v.Fields {
			if atom, ok := field.(*ast.Atom); ok {
				fieldName := atom.Rep
				qualName := name + "." + fieldName

				// Python line 1233-1234: validate field has a sort
				if atom.ASort == nil {
					return lg.NewIvyError(atom, fmt.Sprintf("no sort provided for field %s", fieldName))
				}

				// Get the field's sort
				var fieldSort lg.Sort = lg.TopS
				sn := extractSortRep(atom.ASort)
				if sn != "" {
					if s, err := d.Compiler.CmplSort(sn); err == nil {
						fieldSort = s
					}
				}

				// Create destructor: sort -> fieldSort
				destrSort := il.FuncConstSort(sort, fieldSort)
				destr, err := d.Compiler.AddSymbol(qualName, destrSort, d.Compiler.Sig)
				if err != nil {
					return err
				}
				d.Compiler.Module.DestructorSorts[qualName] = sort
				d.Compiler.Module.SortDestructors[name] = append(
					d.Compiler.Module.SortDestructors[name], destr)
			}
		}
	default:
		// Uninterpreted sort
		sort := &lg.UninterpretedSort{Name: name}
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
func (d *DomainSetup) Axiom(node ast.Node) error {
	xtracer.Trace("compiler.DomainSetup.axiom ENTER")
	lf, ok := node.(*ast.LabeledFormula)
	if !ok {
		return nil
	}
	// Python: cax = ax.compile() — goes through thing() → LF.cmpl.
	cax, err := d.Compiler.ThingLF(lf)
	if err != nil {
		return err
	}

	// Python: if isinstance(cax.formula, SchemaBody): self.domain.schemata[cax.label.relname] = cax
	if _, ok := cax.Formula.(*ast.SchemaBody); ok {
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
func (d *DomainSetup) Property(node ast.Node) error {
	xtracer.Trace("compiler.DomainSetup.property ENTER")
	lf, ok := node.(*ast.LabeledFormula)
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
func (d *DomainSetup) Conjecture(node ast.Node) error {
	xtracer.Trace("compiler.DomainSetup.conjecture ENTER")
	d.LastFact = nil
	return nil
}

// Relation processes a relation declaration.
func (d *DomainSetup) Relation(node ast.Node) error {
	xtracer.Trace("compiler.DomainSetup.relation ENTER")
	atom, ok := node.(*ast.Atom)
	if !ok {
		return nil
	}

	var domSorts []lg.Sort
	for _, arg := range atom.Terms {
		if v, ok := arg.(*ast.Variable); ok {
			s, err := d.Compiler.variableSort(v)
			if err != nil {
				return err
			}
			domSorts = append(domSorts, s)
		}
	}

	sort := il.RelationSort(domSorts)
	sym, err := d.Compiler.AddSymbol(atom.Rep, sort, d.Compiler.Sig)
	if err != nil {
		return err
	}
	// Python: self.domain.all_relations.append((sym, len(rel.args)))
	mod := d.Compiler.Module
	mod.AllRelations = append(mod.AllRelations, sym)
	mod.Relations.Set(atom.Rep, sort)
	return nil
}

// Individual processes a constant (individual) declaration.
// Corresponds to Python IvyDomainSetup.individual (ivy_compiler.py:1103-1106).
func (d *DomainSetup) Individual(node ast.Node) error {
	xtracer.Trace("compiler.DomainSetup.individual ENTER")
	sym, err := d.Compiler.CompileConst(node, d.Compiler.Sig)
	if err != nil {
		return err
	}
	// Python: self.domain.functions[sym] = len(v.args)
	if sym != nil {
		d.Compiler.Module.Functions.Set(sym.Name, sym.CSort)
	}
	return nil
}

// Derived processes a derived relation/function declaration.
// Corresponds to Python IvyDomainSetup.derived.
func (d *DomainSetup) Derived(node ast.Node) error {
	xtracer.Trace("compiler.DomainSetup.derived ENTER")
	if xtracer.Enabled {
		d.Compiler.SigCheck("DomainSetup.derived")
	}
	lf, ok := node.(*ast.LabeledFormula)
	if !ok {
		return nil
	}
	df := lf.Formula
	// Check for DefinitionSchema first (embeds Definition)
	var defNode *ast.Definition
	var isSchema bool
	if ds, ok := df.(*ast.DefinitionSchema); ok {
		defNode = &ds.Definition
		isSchema = true
	} else if dn, ok := df.(*ast.Definition); ok {
		defNode = dn
	} else {
		return nil
	}
	// Validate definition variables BEFORE compile so a free RHS variable is
	// reported as "occurs free on right-hand side of definition" rather than
	// the compile-side "unknown symbol" error. Python (ivy_compiler.py:1319-1325)
	// validates inside add_definition (post-compile) but the error semantics are
	// equivalent because variable names are invariant under compile.
	if err := addDefinitionChecks(defNode); err != nil {
		return err
	}

	lhs := defNode.Lhs
	lhsAtom, ok := lhs.(*ast.Atom)
	if !ok {
		return nil
	}
	// Add a temporary symbol with top function sort
	sym, err := d.Compiler.AddSymbol(lhsAtom.Rep, il.TopFunctionSort(len(lhsAtom.Terms)), d.Compiler.Sig)
	if err != nil {
		return err
	}

	// Compile the definition
	var compiled lg.Expr
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
	if def, ok := compiled.(*il.Definition); ok {
		definesNode := def.Defines()
		if cnst, ok := definesNode.(*lg.Const); ok {
			d.Compiler.AddSymbol(cnst.Name, cnst.CSort, d.Compiler.Sig)
		}
	}

	// Extract concretely-sorted symbol from compiled definition.
	// Python's DerivedUpdate(df) uses defn.args[0].rep which has concrete sorts
	// from compilation. The original `sym` still has TopFunctionSort.
	derivedSym := sym // fallback
	if def, ok := compiled.(*il.Definition); ok {
		if cnst, ok := def.Defines().(*lg.Const); ok {
			derivedSym = cnst
		}
	}

	// Python: self.add_definition(ldf.clone([label, df]))
	// Clone the LabeledFormula with the compiled definition, preserving metadata.
	// AddDefinition validates the definition variables (using astDefNode, the
	// pre-compile AST form) and routes to either NativeDefinitions or
	// LabeledProps based on whether the AST RHS is a NativeExpr.
	mlf := lf.Clone([]ast.Node{lf.Label, compiled}).(*ast.LabeledFormula)
	if err := d.AddDefinition(mlf, defNode); err != nil {
		return err
	}
	mod := d.Compiler.Module
	mod.SymbolOrder = append(mod.SymbolOrder, sym)

	// Python: self.domain.all_relations.append((sym, len(lhs.args)))
	// Python: self.domain.relations[sym] = len(lhs.args)
	mod.AllRelations = append(mod.AllRelations, sym)
	mod.Relations.Set(sym.Name, sym.CSort)

	// Python: self.domain.updates.append(DerivedUpdate(df))
	mod.Updates = append(mod.Updates,
		module.NewDerivedUpdate(derivedSym, compiled))

	return nil
}

// DefinitionDecl processes a definition declaration.
// Corresponds to Python IvyDomainSetup.definition.
func (d *DomainSetup) DefinitionDecl(node ast.Node) error {
	xtracer.Trace("compiler.DomainSetup.definition ENTER")
	if xtracer.Enabled {
		d.Compiler.SigCheck("DomainSetup.definition")
	}
	lf, ok := node.(*ast.LabeledFormula)
	if !ok {
		return nil
	}
	df := lf.Formula
	// Check for DefinitionSchema first
	var defNode *ast.Definition
	var isSchemaD bool
	if ds, ok := df.(*ast.DefinitionSchema); ok {
		defNode = &ds.Definition
		isSchemaD = true
	} else if dn, ok := df.(*ast.Definition); ok {
		defNode = dn
	} else {
		return nil
	}
	// Validate definition variables BEFORE compile so a free RHS variable is
	// reported as "occurs free on right-hand side of definition" rather than
	// the compile-side "unknown symbol" error. Python (ivy_compiler.py:1319-1325)
	// validates inside add_definition (post-compile) but the error semantics are
	// equivalent because variable names are invariant under compile.
	if err := addDefinitionChecks(defNode); err != nil {
		return err
	}

	// Add a temporary symbol so compilation can resolve the defined name
	var tempSym *lg.Const
	if lhsAtom, ok := defNode.Lhs.(*ast.Atom); ok {
		if _, exists := d.Compiler.Sig.Symbols.Get2(lhsAtom.Rep); !exists {
			var err error
			tempSym, err = d.Compiler.AddSymbol(lhsAtom.Rep, il.TopFunctionSort(len(lhsAtom.Terms)), d.Compiler.Sig)
			if err != nil {
				return err
			}
		}
	}

	var compiled lg.Expr
	var err error
	if isSchemaD {
		compiled, err = d.Compiler.CompileDefnSchema(d.Compiler.Module.Cfg.AstCfg.NewDefinitionSchema(*defNode))
	} else {
		compiled, err = d.Compiler.CompileDefn(defNode)
	}
	if err != nil {
		return err
	}

	// Remove temporary symbol and re-add with inferred sort
	if tempSym != nil {
		d.Compiler.Sig.Symbols.Delkey(tempSym.Name)
	}

	// Python: self.add_definition(ldf.clone([label, df]))
	// Clone the LabeledFormula with the compiled definition, preserving metadata.
	// AddDefinition validates the definition variables (using astDefNode, the
	// pre-compile AST form) and routes to either NativeDefinitions or
	// LabeledProps based on whether the AST RHS is a NativeExpr.
	mlf := lf.Clone([]ast.Node{lf.Label, compiled}).(*ast.LabeledFormula)
	if err := d.AddDefinition(mlf, defNode); err != nil {
		return err
	}

	// Add the defined symbol if not already in the signature
	if def, ok := compiled.(*il.Definition); ok {
		definesNode := def.Defines()
		if cnst, ok := definesNode.(*lg.Const); ok {
			if _, exists := d.Compiler.Sig.Symbols.Get2(cnst.Name); !exists {
				d.Compiler.AddSymbol(cnst.Name, cnst.CSort, d.Compiler.Sig)
			}
			d.Compiler.Module.SymbolOrder = append(d.Compiler.Module.SymbolOrder, cnst)
		}
		// Python: self.domain.updates.append(DerivedUpdate(df))
		d.Compiler.Module.Updates = append(d.Compiler.Module.Updates,
			module.NewDerivedUpdate(def.Defines(), compiled))
	}
	return nil
}

// Action processes an action declaration in pass 1 (DomainSetup).
// Corresponds to Python IvyDomainSetup.action (ivy_compiler.py:1344-1370).
// In pass 1, Python only scans for ThunkAction instances to declare thunk types.
// Actual action compilation happens in pass 3 (ARGSetup).
func (d *DomainSetup) Action(node ast.Node) error {
	xtracer.Trace("compiler.DomainSetup.action ENTER")
	actDef, ok := node.(*ast.ActionDef)
	if !ok {
		return nil
	}
	// Python: for action in a.args[1].iter_subactions():
	//             if isinstance(action, ThunkAction): ...
	return iterASTSubactions(actDef.Body, func(sub ast.Node) error {
		thunk, ok := sub.(*ast.ThunkAction)
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
		subtypeAtom, ok := subtype.(*ast.Atom)
		if !ok {
			return nil
		}
		actname := d.Compiler.Module.Cfg.IuCfg.ComposeNames(subtypeAtom.Relname(), "run")

		// Python: selfparam = Atom('fml:$self', []); selfparam.sort = subtype.relname
		selfparam := cfg.NewAtom("fml:$self")
		selfparam.ASort = cfg.NewAtom(subtypeAtom.Relname())

		// Python: orig_args = action.args[0].args + action.args[1].args
		var origArgs []ast.Node
		if thunk.Label != nil {
			origArgs = append(origArgs, thunk.Label.Args()...)
		}
		if thunk.Action != nil {
			origArgs = append(origArgs, thunk.Action.Args()...)
		}

		// Python: top_context.actions[actname] = (orig_args + [selfparam], [], len(orig_args))
		allFormals := make([]ast.Node, len(origArgs)+1)
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
func iterASTSubactions(node ast.Node, fn func(ast.Node) error) error {
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
func (d *DomainSetup) Init(node ast.Node) error {
	return nil
}

// Object processes an object declaration.
func (d *DomainSetup) Object(node ast.Node) error {
	xtracer.Trace("compiler.DomainSetup.object ENTER")
	if atom, ok := node.(*ast.Atom); ok {
		d.Compiler.Module.AddObject(atom.Rep)
	}
	return nil
}

// Variant processes a variant declaration.
// Corresponds to Python IvyDomainSetup.variant (ivy_compiler.py:1248-1253).
// Python: variants[v.args[1].rep].append(sig.sorts[v.args[0].rep])
//
//	supertypes[v.args[0].rep] = sig.sorts[v.args[1].rep]
func (d *DomainSetup) Variant(node ast.Node) error {
	xtracer.Trace("compiler.DomainSetup.variant ENTER")
	vd, ok := node.(*ast.VariantDef)
	if !ok {
		return nil
	}
	sortName := extractSortRep(vd.Name)     // subtype (args[0] in Python)
	variantName := extractSortRep(vd.VSort) // supertype (args[1] in Python)
	if sortName == "" || variantName == "" {
		return nil
	}

	// Validate both sorts exist (Python: if r.rep not in self.domain.sig.sorts)
	subtypeSort, err := d.Compiler.Sig.FindSort(sortName, false)
	if err != nil {
		return lg.NewIvyError(vd, fmt.Sprintf("undefined sort: %s", sortName))
	}
	supertypeSort, err := d.Compiler.Sig.FindSort(variantName, false)
	if err != nil {
		return lg.NewIvyError(vd, fmt.Sprintf("undefined sort: %s", variantName))
	}

	// variants[supertype] ← append subtype sort
	// Python: self.domain.variants[v.args[1].rep].append(self.domain.sig.sorts[v.args[0].rep])
	d.Compiler.Module.Variants[variantName] = append(
		d.Compiler.Module.Variants[variantName], subtypeSort)

	// supertypes[subtype] = supertype sort
	// Python: self.domain.supertypes[v.args[0].rep] = self.domain.sig.sorts[v.args[1].rep]
	d.Compiler.Module.Supertypes[sortName] = []lg.Sort{supertypeSort}

	return nil
}

// Export processes an export declaration in pass 1 (DomainSetup).
// Python's IvyDomainSetup does NOT have an export method — exports
// are only handled in pass 3 (ARGSetup). This is a no-op.
func (d *DomainSetup) Export(node ast.Node) error {
	xtracer.Trace("compiler.DomainSetup.export ENTER")
	return nil
}

// Import processes an import declaration in pass 1 (DomainSetup).
// Python's IvyDomainSetup does NOT have an import method — imports
// are only handled in pass 3 (ARGSetup). This is a no-op.
func (d *DomainSetup) Import(node ast.Node) error {
	xtracer.Trace("compiler.DomainSetup.import ENTER")
	return nil
}

// Isolate in pass 1 (DomainSetup) is a no-op.
// Python's IvyDomainSetup does NOT have an isolate method — isolates
// are only handled in pass 3 (ARGSetup). See ivy_compiler.py:1429-1434.
func (d *DomainSetup) Isolate(node ast.Node) error {
	xtracer.Trace("compiler.DomainSetup.isolate ENTER")
	return nil
}

// Interpret processes a type interpretation.
// Faithful port of Python IvyDomainSetup.interpret (ivy_compiler.py:1333-1399).
func (d *DomainSetup) Interpret(node ast.Node) error {
	xtracer.Trace("compiler.DomainSetup.interpret ENTER")

	// BB0: Extract lhs and rhs from the Implies formula inside the LabeledFormula.
	// Python: thing.formula is an Implies; .args[0] = lhs, .args[1] = rhs
	lf, ok := node.(*ast.LabeledFormula)
	if !ok {
		return nil
	}
	impl, ok := lf.Formula.(*ast.Implies)
	if !ok {
		return nil
	}
	sig := d.Compiler.Sig
	mod := d.Compiler.Module
	interp := sig.Interp

	// Python: lhs = resolve_alias(thing.formula.args[0].rep)
	lhs := ResolveAlias(extractSortRep(impl.T1), mod)
	// Python: rhs = thing.formula.args[1]  (the AST node)
	rhs := impl.T2

	// Python: xtracer.trace("compiler.DomainSetup.interpret lhs=%s rhs=%s\n rhsType=%s" % (lhs, type(rhs).__name__, type(rhs).__name__))
	rhsTypeName := astTypeName(rhs)
	xtracer.Trace("compiler.DomainSetup.interpret lhs=%s rhs=%s", lhs, rhsTypeName)
	//pp("rhsType=%s", rhsTypeName)

	// BB1: Handle native type interpretation
	// Python: if isinstance(thing.formula.args[1], ivy_ast.NativeType):
	if nt, ok := rhs.(*ast.NativeType); ok {
		xtracer.Trace("compiler.DomainSetup.interpret branch=nativeType")
		// Python: if lhs in interp or lhs in self.domain.native_types:
		if _, exists := interp[lhs]; exists {
			return lg.NewIvyError(node, fmt.Sprintf("%s is already interpreted", lhs))
		}
		if _, exists := mod.NativeTypes[lhs]; exists {
			return lg.NewIvyError(node, fmt.Sprintf("%s is already interpreted", lhs))
		}
		// Python: self.domain.native_types[lhs] = compile_native_type(thing.formula.args[1])
		mod.NativeTypes[lhs] = compileNativeType(nt, mod)
		// Python: if thing.formula.args[1].args[0].code.strip() == 'int':
		//             compile_theory(self.domain, lhs, 'int')
		if len(nt.Elems) > 0 {
			isInt := false
			if atom, ok := nt.Elems[0].(*ast.Atom); ok && atom.Rep == "int" {
				isInt = true
			}
			if nc, ok := nt.Elems[0].(*ast.NativeCode); ok && strings.TrimSpace(nc.Code) == "int" {
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
	case *ast.Range:
		// rhsName stays empty; handled in BB4 below
	case *ast.EnumeratedSort:
		// rhsName stays empty; handled in BB5 below
	default:
		rhsName = extractSortRep(rhs)
	}
	xtracer.Trace("compiler.DomainSetup.interpret branch=non-native rhsName=%s", rhsName)

	// Python: self.domain.interps[lhs].append(thing)
	mod.Interps[lhs] = append(mod.Interps[lhs], node)

	// Python: if lhs in self.domain.native_types: raise IvyError(...)
	if _, exists := mod.NativeTypes[lhs]; exists {
		return lg.NewIvyError(node, fmt.Sprintf("%s is already interpreted", lhs))
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
		return lg.NewIvyError(node, fmt.Sprintf("%s is already interpreted", lhs))
	}

	// BB4: Range interpretation
	// Python: if isinstance(rhs, ivy_ast.Range):
	if rng, ok := rhs.(*ast.Range); ok {
		xtracer.Trace("compiler.DomainSetup.interpret branch=range")
		// Python: if lhs not in sig.sorts: raise IvyError(...)
		if _, exists := sig.Sorts.Get2(lhs); !exists {
			return lg.NewIvyError(node, fmt.Sprintf("%s is not a sort", lhs))
		}
		sort := sig.Sorts.Get(lhs)
		// Python: if not isinstance(sort, ivy_logic.UninterpretedSort): raise IvyError(...)
		if _, isUn := sort.(*lg.UninterpretedSort); !isUn {
			return lg.NewIvyError(node, fmt.Sprintf("%s is already interpreted", lhs))
		}
		// Python: compile_bound for lo and hi
		lo := d.compileBound(rng.Lo, lhs, sort, node)
		hi := d.compileBound(rng.Hi, lhs, sort, node)
		rangeSort := &lg.RangeSort{Name: lhs, Lb: lo, Ub: hi}
		interp[lhs] = rangeSort
		// Python: compile_theory(self.domain, lhs, interp[lhs])
		// get_theory_schemata maps RangeSort → "int"
		if err := d.Compiler.CompileTheory(lhs, "int"); err != nil {
			return err
		}
		xtracer.Trace("compiler.DomainSetup.interpret return=range")
		return nil
	}

	// BB5: Enumerated sort interpretation
	// Python: if isinstance(rhs, ivy_ast.EnumeratedSort):
	if enumSort, ok := rhs.(*ast.EnumeratedSort); ok {
		xtracer.Trace("compiler.DomainSetup.interpret branch=enum")
		// Python: if lhs not in self.domain.sig.sorts: raise IvyError(...)
		if _, exists := sig.Sorts.Get2(lhs); !exists {
			return lg.NewIvyError(node, fmt.Sprintf("%s is not a type", lhs))
		}
		ext := enumSort.Extension()
		sort := &lg.EnumeratedSort{Name: lhs, Extension: ext}
		interp[lhs] = sort
		// Python: for c in sort.defines(): register constructors
		for _, c := range ext {
			if existingSort, hasSig := sig.Sorts.Get2(lhs); hasSig {
				sym := lg.NewConst(c, existingSort)
				sig.Symbols.Set(c, &il.SymbolEntry{Sort: existingSort})
				mod.Functions.Set(c, existingSort)
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
		if !solver.IsSolverSort(rhsName) {
			return lg.NewIvyError(node, fmt.Sprintf("%s not a native sort", rhsName))
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
		if !solver.IsSolverOp(rhsName) {
			return lg.NewIvyError(node, fmt.Sprintf("%s not a native symbol", rhsName))
		}
		interp[lhs] = rhsName
		xtracer.Trace("compiler.DomainSetup.interpret return=solver-symbol")
		return nil
	}

	// BB7: Python: raise IvyUndefined(thing, lhs)
	xtracer.Trace("compiler.DomainSetup.interpret branch=undefined")
	return lg.NewIvyError(node, fmt.Sprintf("%s undefined", lhs))
}

// compileBound compiles a range bound, returning either a NumeralBound
// or a CompiledBound. Matches Python ivy_compiler.py:1295-1306 compile_bound.
func (d *DomainSetup) compileBound(b ast.Node, lhsName string, sort lg.Sort, context ast.Node) lg.NumeralOrCompiledBound {
	if b == nil {
		return lg.NumeralBound{Value: "0"}
	}
	rep := fmt.Sprint(b)
	if il.IsNumeralName(rep) {
		return lg.NumeralBound{Value: rep}
	}
	// Non-numeral bound: compile as parameter
	// Python: b.sort = lhs; self.parameter(b); res = b.compile()
	if atom, ok := b.(*ast.Atom); ok {
		cfg := d.Compiler.Module.Cfg.AstCfg
		atom.ASort = cfg.NewAtom(lhsName)
		_ = d.Parameter(b) // register as parameter
	}
	compiled, err := d.Compiler.Thing(b)
	if err != nil {
		// Fall back to string representation
		return lg.NumeralBound{Value: rep}
	}
	return lg.CompiledBound{Expr: compiled}
}

// Mixin processes a mixin declaration in pass 1 (DomainSetup).
// Corresponds to Python IvyDomainSetup.mixin (ivy_compiler.py:1340-1342).
// Pass 1 only validates the mixee exists; it does NOT store the mixin.
// Mixin storage happens in pass 3 (ARGSetup).
func (d *DomainSetup) Mixin(node ast.Node) error {
	xtracer.Trace("compiler.DomainSetup.mixin ENTER")
	args := node.Args()
	if len(args) < 2 {
		return nil
	}
	var mixeeName string
	if atom, ok := args[1].(*ast.Atom); ok {
		mixeeName = atom.Relname()
	} else {
		return nil
	}
	// Validate mixee: must be 'init' or a known action
	// Python: if m.args[1].relname != 'init' and m.args[1].relname not in top_context.actions:
	if mixeeName != "init" && d.Compiler.TopCtx != nil {
		if _, ok := d.Compiler.TopCtx.Actions[mixeeName]; !ok {
			return lg.NewIvyError(node, fmt.Sprintf("unknown action: %s", mixeeName))
		}
	}
	// Do NOT store mixin here — that's ARGSetup's job (pass 3)
	return nil
}

// Delegate processes a delegate declaration in pass 1 (DomainSetup).
// Python's IvyDomainSetup does NOT have a delegate method — delegates
// are only handled in pass 3 (ARGSetup). This is a no-op.
func (d *DomainSetup) Delegate(node ast.Node) error {
	xtracer.Trace("compiler.DomainSetup.delegate ENTER")
	return nil
}

// Native processes a native code declaration.
// Python only handles native in IvyARGSetup (pass 3), not IvyDomainSetup (pass 1).
// ARGSetup.ProcessDecls already compiles via CompileNativeDef and appends.
func (d *DomainSetup) Native(node ast.Node) error {
	xtracer.Trace("compiler.DomainSetup.native ENTER")
	return nil
}

// Alias processes an alias declaration.
func (d *DomainSetup) Alias(node ast.Node) error {
	xtracer.Trace("compiler.DomainSetup.alias ENTER")
	if def, ok := node.(*ast.Definition); ok {
		aliasName := extractSortRep(def.Lhs)
		targetName := extractSortRep(def.Rhs)
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
func (d *DomainSetup) Attribute(node ast.Node) error {
	xtracer.Trace("compiler.DomainSetup.attribute ENTER")
	return nil
}

// Progress processes a progress declaration.
// Corresponds to Python IvyDomainSetup.progress (ivy_compiler.py:1197-1202).
func (d *DomainSetup) Progress(node ast.Node) error {
	xtracer.Trace("compiler.DomainSetup.progress ENTER")
	args := node.Args()
	if len(args) >= 2 {
		rel := args[0]
		body := args[1]
		if atom, ok := rel.(*ast.Atom); ok {
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
func (d *DomainSetup) Private(node ast.Node) error {
	xtracer.Trace("compiler.DomainSetup.private ENTER")
	return nil
}

// Schema processes a schema declaration.
// Corresponds to Python IvyDomainSetup.schema.
func (d *DomainSetup) Schema(node ast.Node) error {
	xtracer.Trace("compiler.DomainSetup.schema ENTER")
	if xtracer.Enabled {
		d.Compiler.SigCheck("DomainSetup.schema")
	}
	// Handle *ast.Schema directly (e.g. from theory compilation).
	// Python: schema(self, sch) accesses sch.defn.args[1] and compiles
	// it if it's a SchemaBody. We must do the same — not just store raw.
	if schema, ok := node.(*ast.Schema); ok {
		defn := schema.Defn.(*ast.Definition)
		// Check if RHS is SchemaBody — if so, compile it
		// Python: if isinstance(sch.defn.args[1], ivy_ast.SchemaBody):
		//   ldf = ivy_ast.LabeledFormula(label, sch.defn.args[1].compile())
		// Note: Python's SchemaBody.compile = compile_schema_body (not thing())
		if sb, ok := defn.Rhs.(*ast.SchemaBody); ok {
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
func (d *DomainSetup) Instantiate(node ast.Node) error {
	xtracer.Trace("compiler.DomainSetup.instantiate ENTER")
	// Instantiation applies a schema. Extract the prefix and inst name,
	// look up the schema, and apply it.
	inst, ok := node.(*ast.Instantiation)
	if !ok {
		return nil
	}
	var instName string
	if inst.Sort != nil {
		if atom, ok := inst.Sort.(*ast.Atom); ok {
			instName = atom.Relname()
		}
	}
	if instName == "" {
		return nil
	}

	schema, ok := d.Compiler.Module.Schemata.Get2(instName)
	if ok {
		if c, canOk := schema.(iu.Canonizer); canOk {
			xtracer.Trace("compiler.DomainSetup.instantiate schemata.lookup key='%s' found=true value=%s", instName, c.Canon())
		} else {
			xtracer.Trace("compiler.DomainSetup.instantiate schemata.lookup key='%s' found=true value=%v", instName, schema)
		}
	} else {
		xtracer.Trace("compiler.DomainSetup.instantiate schemata.lookup key='%s' found=false", instName)
		return lg.NewIvyError(inst, fmt.Sprintf("%s undefined in instantiation", instName))
	}

	// Store the instantiation for later processing
	d.Compiler.Module.Instantiations = append(d.Compiler.Module.Instantiations,
		module.Instantiation{Schema: schema, Inst: node})
	return nil
}

// Proof processes a proof declaration.
// Corresponds to Python IvyDomainSetup.proof.
func (d *DomainSetup) Proof(node ast.Node) error {
	xtracer.Trace("compiler.DomainSetup.proof ENTER")
	// If the proof is a labeled formula, it has its own label.
	if lf, ok := node.(*ast.LabeledFormula); ok {
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
		d.Compiler.Module.Proofs = append(d.Compiler.Module.Proofs, module.ProofEntry{
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
	d.Compiler.Module.Proofs = append(d.Compiler.Module.Proofs, module.ProofEntry{
		Formula: d.LastFact,
		Proof:   compiled,
	})
	return nil
}

// Named processes a named declaration.
// Corresponds to Python IvyDomainSetup.named.
// A named declaration gives a name to an existential property.
func (d *DomainSetup) Named(node ast.Node) error {
	xtracer.Trace("compiler.DomainSetup.named ENTER")
	lhs, ok := node.(*ast.Atom)
	if !ok {
		return nil
	}
	if d.LastFact == nil {
		return lg.NewIvyError(node, "named declaration without preceding property")
	}

	// The last fact should be an existential formula.
	// We extract the existential's range sort and create a function symbol.
	// Python: cond = ivy_logic.drop_universals(self.last_fact.formula)
	lastFormula, ok := d.LastFact.Formula.(lg.Expr)
	if !ok {
		return lg.NewIvyError(node, "named declaration without preceding property")
	}
	cond := il.DropUniversals(lastFormula)
	if !il.IsExists(cond) {
		return lg.NewIvyError(node, "property is not existential")
	}

	// Get the existential variable's sort as the range
	ex, ok := cond.(*lg.Exists)
	if !ok || len(ex.Variables) != 1 {
		return lg.NewIvyError(node, "property is not existential")
	}
	rng := ex.Variables[0].VSort

	// Build domain sorts from the lhs parameters
	var domSorts []lg.Sort
	for _, arg := range lhs.Terms {
		compiled, err := d.Compiler.Thing(arg)
		if err != nil {
			return err
		}
		domSorts = append(domSorts, compiled.NodeSort())
	}

	sort := il.FuncConstSort(append(domSorts, rng)...)
	sym, err := d.Compiler.AddSymbol(lhs.Rep, sort, d.Compiler.Sig)
	if err != nil {
		return err
	}

	// Python: self.domain.named.append((self.last_fact, sym(...)))
	d.Compiler.Module.Named = append(d.Compiler.Module.Named, module.NamedEntry{
		Formula: d.LastFact,
		Name:    sym,
	})

	// Python (ivy_compiler.py:1237): self.domain.updates.append(NamedUpdate(sym, cond))
	d.Compiler.Module.Updates = append(d.Compiler.Module.Updates,
		actions.NewNamedUpdate(sym, cond))
	return nil
}

// Theorem processes a theorem declaration.
// Corresponds to Python IvyDomainSetup.theorem.
func (d *DomainSetup) Theorem(node ast.Node) error {
	xtracer.Trace("compiler.DomainSetup.theorem ENTER")
	// A theorem is like a schema but added as a labeled property.
	lf, ok := node.(*ast.LabeledFormula)
	if !ok {
		return nil
	}
	if lf.Formula == nil {
		return nil
	}

	df, ok := lf.Formula.(*ast.Definition)
	if !ok {
		return nil
	}

	defName := df.Defines()

	if _, ok := df.Rhs.(*ast.SchemaBody); ok {
		compiled, err := d.Compiler.Thing(df.Rhs)
		if err != nil {
			return err
		}
		acfg := d.Compiler.Module.Cfg.AstCfg
		mlf := acfg.NewLabeledFormula(nil, compiled)
		mlf.SetLineno(lf.GetLineno())
		d.Compiler.Module.LabeledProps = append(d.Compiler.Module.LabeledProps, mlf)
		d.Compiler.Module.Theorems[defName] = compiled
		d.LastFact = mlf
	}
	return nil
}

// Assert in pass 1 (DomainSetup) is a no-op.
// Python's IvyDomainSetup does NOT have an _assert method — asserts
// are only handled in pass 3 (ARGSetup). See ivy_compiler.py:1426-1428.
func (d *DomainSetup) Assert(node ast.Node) error {
	return nil
}

// Parameter processes a parameter declaration.
// Corresponds to Python IvyDomainSetup.parameter (ivy_compiler.py:1108).
func (d *DomainSetup) Parameter(node ast.Node) error {
	xtracer.Trace("compiler.DomainSetup.parameter ENTER")
	mod := d.Compiler.Module
	sig := d.Compiler.Sig
	var sym *lg.Const
	var dflt ast.Node // raw AST node, matching Python
	if def, ok := node.(*ast.Definition); ok {
		var err error
		sym, err = d.Compiler.CompileConst(def.Lhs, sig)
		if err != nil {
			return err
		}
		dflt = def.Rhs // Python stores raw AST node, not string
	} else {
		var err error
		sym, err = d.Compiler.CompileConst(node, sig)
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
func (d *DomainSetup) Destructor(node ast.Node) error {
	xtracer.Trace("compiler.DomainSetup.destructor ENTER")
	mod := d.Compiler.Module
	sym, err := d.Compiler.CompileConst(node, d.Compiler.Sig)
	if err != nil {
		return err
	}
	dom := il.SortDomain(sym.CSort)
	if len(dom) == 0 {
		return lg.NewIvyError(node, "A destructor must have at least one parameter")
	}
	mod.DestructorSorts[sym.Name] = dom[0]
	mod.SortDestructors[il.SortName(dom[0])] = append(
		mod.SortDestructors[il.SortName(dom[0])], sym)
	return nil
}

// Constructor processes a constructor declaration.
// Corresponds to Python IvyDomainSetup.constructor (ivy_compiler.py:1125).
func (d *DomainSetup) Constructor(node ast.Node) error {
	xtracer.Trace("compiler.DomainSetup.constructor ENTER")
	mod := d.Compiler.Module
	sym, err := d.Compiler.CompileConst(node, d.Compiler.Sig)
	if err != nil {
		return err
	}
	rng := il.SortRange(sym.CSort)
	mod.ConstructorSorts[sym.Name] = rng
	sortName := il.SortName(rng)
	mod.SortConstructors[sortName] = append(mod.SortConstructors[sortName], sym)
	return nil
}

// Concept processes a concept declaration.
// Corresponds to Python IvyDomainSetup.concept (ivy_compiler.py:1208).
func (d *DomainSetup) Concept(node ast.Node) error {
	xtracer.Trace("compiler.DomainSetup.concept ENTER")
	mod := d.Compiler.Module
	lf, ok := node.(*ast.LabeledFormula)
	if !ok {
		return nil
	}
	// lf.Label is the relation atom, lf.Formula is the body
	// Python: rel = c.args[0]; add_symbol(rel.relname, get_relation_sort(sig, rel.args, c.args[1]))
	body := lf.Formula
	if atom, ok := lf.Label.(*ast.Atom); ok {
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
	var compiledLabel lg.Expr
	if lf.Label != nil {
		var err error
		compiledLabel, err = d.Compiler.SortifyWithInference(lf.Label)
		if err != nil {
			return err
		}
	}
	var compiledBody lg.Expr
	if lf.Formula != nil {
		var err error
		compiledBody, err = d.Compiler.SortifyWithInference(lf.Formula)
		if err != nil {
			return err
		}
	}
	mod.ConceptSpaces = append(mod.ConceptSpaces, module.ConceptSpace{Label: compiledLabel, Body: compiledBody})
	return nil
}

// Rely processes a rely declaration.
// Corresponds to Python IvyDomainSetup.rely (ivy_compiler.py:1203).
func (d *DomainSetup) Rely(node ast.Node) error {
	xtracer.Trace("compiler.DomainSetup.rely ENTER")
	mod := d.Compiler.Module
	lf, ok := node.(*ast.LabeledFormula)
	if !ok {
		return nil
	}
	compiled, err := d.Compiler.SortifyWithInference(lf.Formula)
	if err != nil {
		return err
	}
	mod.Rely = append(mod.Rely, compiled)
	return nil
}

// Mixord processes a mixord declaration.
// Corresponds to Python IvyDomainSetup.mixord (ivy_compiler.py:1206).
func (d *DomainSetup) Mixord(node ast.Node) error {
	xtracer.Trace("compiler.DomainSetup.mixord ENTER")
	d.Compiler.Module.MixOrd = append(d.Compiler.Module.MixOrd, node)
	return nil
}

// Update processes an update declaration.
// Corresponds to Python IvyDomainSetup.update (ivy_compiler.py:1214).
// Python: self.domain.updates.append(upd.compile())
func (d *DomainSetup) Update(node ast.Node) error {
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
func (d *DomainSetup) Scenario(node ast.Node) error {
	xtracer.Trace("compiler.DomainSetup.scenario ENTER")
	mod := d.Compiler.Module
	sig := d.Compiler.Sig
	scenDef, ok := node.(*ast.ScenarioDef)
	if !ok {
		return nil
	}
	relSort := il.RelationSort([]lg.Sort{})
	for _, pi := range scenDef.Places() {
		sym, err := d.Compiler.AddSymbol(pi.Name, relSort, sig)
		if err != nil {
			return err
		}
		mod.AllRelations = append(mod.AllRelations, sym)
		mod.Relations.Set(pi.Name, relSort)
	}
	return nil
}

// Implementtype processes an implement type declaration.
// Corresponds to Python IvyDomainSetup.implementtype (ivy_compiler.py:1254).
func (d *DomainSetup) Implementtype(node ast.Node) error {
	xtracer.Trace("compiler.DomainSetup.implementtype ENTER")
	mod := d.Compiler.Module
	sig := d.Compiler.Sig
	lf, ok := node.(*ast.LabeledFormula)
	if !ok {
		return nil
	}
	def, ok := lf.Formula.(*ast.Definition)
	if !ok {
		return nil
	}
	impd := extractSortRep(def.Lhs)
	impr := extractSortRep(def.Rhs)
	// Validate both sorts exist
	if _, ok := sig.Sorts.Get2(impd); !ok {
		return lg.NewIvyError(lf, fmt.Sprintf("undefined sort: %s", impd))
	}
	if _, ok := sig.Sorts.Get2(impr); !ok {
		return lg.NewIvyError(lf, fmt.Sprintf("undefined sort: %s", impr))
	}
	// Check not already interpreted
	if _, ok := mod.NativeTypes[impd]; ok {
		return lg.NewIvyError(lf, fmt.Sprintf("%s is already interpreted", impd))
	}
	if _, ok := sig.Interp[impd]; ok {
		return lg.NewIvyError(lf, fmt.Sprintf("%s is already interpreted", impd))
	}
	impdSort := sig.Sorts.Get(impd)
	imprSort := sig.Sorts.Get(impr)
	il.ImplementType(sig, impdSort, imprSort)
	mod.Interps[impd] = append(mod.Interps[impd], node)
	return nil
}
