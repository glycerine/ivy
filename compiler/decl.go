package compiler

import (
	"fmt"

	"github.com/glycerine/goivy/ast"
	il "github.com/glycerine/goivy/ivylogic"
	lg "github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/module"
)

// DeclInterp processes top-level declarations, replacing Python's
// IvyDomainSetup class.
type DeclInterp struct {
	Compiler *Compiler

	// LastFact holds the last compiled property/axiom, used by
	// named declarations.
	LastFact lg.Node
}

// NewDeclInterp creates a new declaration interpreter.
func NewDeclInterp(c *Compiler) *DeclInterp {
	return &DeclInterp{Compiler: c}
}

// ProcessDecls processes all declarations in a declaration list.
// Each declaration is dispatched to the appropriate handler based on its type.
func (d *DeclInterp) ProcessDecls(decls []ast.Node) error {
	for _, decl := range decls {
		if err := d.ProcessDecl(decl); err != nil {
			return fmt.Errorf("at %s: %w", decl.GetLineno(), err)
		}
	}
	return nil
}

// ProcessDecl dispatches a single declaration to its handler.
func (d *DeclInterp) ProcessDecl(decl ast.Node) error {
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
	case *ast.InitDecl:
		for _, arg := range n.DeclArgs {
			if err := d.Init(arg); err != nil {
				return err
			}
		}
	case *ast.ObjectDecl:
		for _, arg := range n.DeclArgs {
			if err := d.Object(arg); err != nil {
				return err
			}
		}
	case *ast.ModuleDecl:
		for _, arg := range n.DeclArgs {
			if err := d.ModuleD(arg); err != nil {
				return err
			}
		}
	case *ast.VariantDecl:
		for _, arg := range n.DeclArgs {
			if err := d.Variant(arg); err != nil {
				return err
			}
		}
	case *ast.ExportDecl:
		for _, arg := range n.DeclArgs {
			if err := d.Export(arg); err != nil {
				return err
			}
		}
	case *ast.ImportDecl:
		for _, arg := range n.DeclArgs {
			if err := d.Import(arg); err != nil {
				return err
			}
		}
	case *ast.IsolateDecl:
		for _, arg := range n.DeclArgs {
			if err := d.Isolate(arg); err != nil {
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
	case *ast.DelegateDecl:
		for _, arg := range n.DeclArgs {
			if err := d.Delegate(arg); err != nil {
				return err
			}
		}
	case *ast.NativeDecl:
		for _, arg := range n.DeclArgs {
			if err := d.Native(arg); err != nil {
				return err
			}
		}
	case *ast.AliasDecl:
		for _, arg := range n.DeclArgs {
			if err := d.Alias(arg); err != nil {
				return err
			}
		}
	case *ast.AttributeDecl:
		for _, arg := range n.DeclArgs {
			if err := d.Attribute(arg); err != nil {
				return err
			}
		}
	case *ast.ProgressDecl:
		for _, arg := range n.DeclArgs {
			if err := d.Progress(arg); err != nil {
				return err
			}
		}
	case *ast.PrivateDecl:
		for _, arg := range n.DeclArgs {
			if err := d.Private(arg); err != nil {
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
	case *ast.AssertDecl:
		for _, arg := range n.DeclArgs {
			if err := d.Assert(arg); err != nil {
				return err
			}
		}
	default:
		// Unknown declaration type: skip with no error
	}
	return nil
}

// --- Individual declaration handlers ---

// TypeDecl processes a type declaration.
func (d *DeclInterp) TypeDecl(node ast.Node) error {
	td, ok := node.(*ast.TypeDef)
	if !ok {
		// Plain type declaration (no definition)
		if sym, ok := node.(*ast.Symbol); ok {
			sort := &lg.UninterpretedSort{Name: sym.Rep}
			if err := d.Compiler.Sig.AddSort(sort); err != nil {
				// Sort already exists - not fatal
				return nil
			}
			d.Compiler.Module.SortOrder = append(d.Compiler.Module.SortOrder, sym.Rep)
			return nil
		}
		if atom, ok := node.(*ast.Atom); ok {
			sort := &lg.UninterpretedSort{Name: atom.Rep}
			if err := d.Compiler.Sig.AddSort(sort); err != nil {
				return nil
			}
			d.Compiler.Module.SortOrder = append(d.Compiler.Module.SortOrder, atom.Rep)
			return nil
		}
		return nil
	}

	// Type definition
	name := extractSortName(td.Name)
	if name == "" {
		return &lg.IvyError{Msg: "type definition has no name"}
	}

	switch v := td.Value.(type) {
	case *ast.ConstantSort:
		sort := &lg.UninterpretedSort{Name: name}
		if err := d.Compiler.Sig.AddSort(sort); err != nil {
			return nil
		}
	case *ast.EnumeratedSort:
		ext := v.Extension()
		sort := &lg.EnumeratedSort{Name: name, Extension: ext}
		if err := d.Compiler.Sig.AddSort(sort); err != nil {
			return nil
		}
		// Add each enum value as a symbol
		for _, elemName := range ext {
			_, err := d.Compiler.AddSymbol(elemName, sort, d.Compiler.Sig)
			if err != nil {
				return err
			}
		}
		if td.Finite {
			d.Compiler.Module.FiniteSorts[name] = true
		}
	case *ast.Range:
		lo := fmt.Sprint(v.Lo)
		hi := fmt.Sprint(v.Hi)
		sort := &lg.RangeSort{Name: name, Lb: lo, Ub: hi}
		if err := d.Compiler.Sig.AddSort(sort); err != nil {
			return nil
		}
	case *ast.StructSort:
		// Add the sort and its destructors
		sort := &lg.UninterpretedSort{Name: name}
		if err := d.Compiler.Sig.AddSort(sort); err != nil {
			return nil
		}
		for _, field := range v.Fields {
			if atom, ok := field.(*ast.Atom); ok {
				fieldName := atom.Rep
				qualName := name + "." + fieldName

				// Get the field's sort
				var fieldSort lg.Sort = lg.TopS
				if atom.ASort != nil {
					sn := extractSortName(atom.ASort)
					if sn != "" {
						if s, err := d.Compiler.CmplSort(sn); err == nil {
							fieldSort = s
						}
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

// Axiom processes an axiom declaration.
func (d *DeclInterp) Axiom(node ast.Node) error {
	lf, ok := node.(*ast.LabeledFormula)
	if !ok {
		return nil
	}
	compiled, err := d.Compiler.CompileNode(lf)
	if err != nil {
		return err
	}

	// Check if it's a schema body
	if _, ok := lf.Formula.(*ast.SchemaBody); ok {
		labelName := lf.LabelName()
		if labelName != "" {
			d.Compiler.Module.Schemata[labelName] = compiled
		}
		return nil
	}

	mlf := &module.LabeledFormula{
		Formula: compiled,
		Lineno:  lf.GetLineno().Line,
	}
	d.Compiler.Module.LabeledAxioms = append(d.Compiler.Module.LabeledAxioms, mlf)
	return nil
}

// Property processes a property declaration.
func (d *DeclInterp) Property(node ast.Node) error {
	lf, ok := node.(*ast.LabeledFormula)
	if !ok {
		return nil
	}
	compiled, err := d.Compiler.CompileNode(lf)
	if err != nil {
		return err
	}

	mlf := &module.LabeledFormula{
		Formula: compiled,
		Lineno:  lf.GetLineno().Line,
	}
	d.Compiler.Module.LabeledProps = append(d.Compiler.Module.LabeledProps, mlf)
	d.LastFact = compiled
	return nil
}

// Conjecture processes a conjecture declaration.
func (d *DeclInterp) Conjecture(node ast.Node) error {
	lf, ok := node.(*ast.LabeledFormula)
	if !ok {
		return nil
	}
	compiled, err := d.Compiler.CompileNode(lf)
	if err != nil {
		return err
	}

	mlf := &module.LabeledFormula{
		Formula: compiled,
		Lineno:  lf.GetLineno().Line,
	}
	d.Compiler.Module.LabeledConjs = append(d.Compiler.Module.LabeledConjs, mlf)
	d.LastFact = nil
	return nil
}

// Relation processes a relation declaration.
func (d *DeclInterp) Relation(node ast.Node) error {
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
	_, err := d.Compiler.AddSymbol(atom.Rep, sort, d.Compiler.Sig)
	if err != nil {
		return err
	}
	d.Compiler.Module.Relations[atom.Rep] = sort
	return nil
}

// Individual processes a constant (individual) declaration.
func (d *DeclInterp) Individual(node ast.Node) error {
	_, err := d.Compiler.CompileConst(node, d.Compiler.Sig)
	return err
}

// Derived processes a derived relation/function declaration.
func (d *DeclInterp) Derived(node ast.Node) error {
	// TODO: implement derived declaration compilation
	return nil
}

// DefinitionDecl processes a definition declaration.
func (d *DeclInterp) DefinitionDecl(node ast.Node) error {
	// TODO: implement definition compilation (compile_defn)
	return nil
}

// Action processes an action declaration.
func (d *DeclInterp) Action(node ast.Node) error {
	// TODO: implement action declaration compilation
	// This will call compile_action_def from Python
	return nil
}

// Init processes an init declaration.
func (d *DeclInterp) Init(node ast.Node) error {
	lf, ok := node.(*ast.LabeledFormula)
	if !ok {
		return nil
	}
	compiled, err := d.Compiler.CompileNode(lf)
	if err != nil {
		return err
	}

	mlf := &module.LabeledFormula{
		Formula: compiled,
		Lineno:  lf.GetLineno().Line,
	}
	d.Compiler.Module.LabeledInits = append(d.Compiler.Module.LabeledInits, mlf)
	return nil
}

// Object processes an object declaration.
func (d *DeclInterp) Object(node ast.Node) error {
	if atom, ok := node.(*ast.Atom); ok {
		d.Compiler.Module.AddObject(atom.Rep)
	}
	return nil
}

// ModuleD processes a module declaration.
func (d *DeclInterp) ModuleD(node ast.Node) error {
	// TODO: implement module instantiation (complex, stubbed)
	return nil
}

// Variant processes a variant declaration.
func (d *DeclInterp) Variant(node ast.Node) error {
	vd, ok := node.(*ast.VariantDef)
	if !ok {
		return nil
	}
	sortName := extractSortName(vd.Name)
	variantName := extractSortName(vd.VSort)
	if sortName == "" || variantName == "" {
		return nil
	}
	variantSort := &lg.UninterpretedSort{Name: variantName}
	d.Compiler.Module.Variants[sortName] = append(
		d.Compiler.Module.Variants[sortName], variantSort)
	return nil
}

// Export processes an export declaration.
func (d *DeclInterp) Export(node ast.Node) error {
	// TODO: implement export handling
	return nil
}

// Import processes an import declaration.
func (d *DeclInterp) Import(node ast.Node) error {
	// TODO: implement import handling
	return nil
}

// Isolate processes an isolate declaration.
func (d *DeclInterp) Isolate(node ast.Node) error {
	// TODO: implement isolate handling
	return nil
}

// Interpret processes a type interpretation.
func (d *DeclInterp) Interpret(node ast.Node) error {
	// TODO: implement interpretation handling
	return nil
}

// Mixin processes a mixin declaration.
func (d *DeclInterp) Mixin(node ast.Node) error {
	// TODO: implement mixin handling
	return nil
}

// Delegate processes a delegate declaration.
func (d *DeclInterp) Delegate(node ast.Node) error {
	// TODO: implement delegate handling
	return nil
}

// Native processes a native code declaration.
func (d *DeclInterp) Native(node ast.Node) error {
	// TODO: implement native handling
	return nil
}

// Alias processes an alias declaration.
func (d *DeclInterp) Alias(node ast.Node) error {
	if def, ok := node.(*ast.Definition); ok {
		aliasName := extractSortName(def.Lhs)
		targetName := extractSortName(def.Rhs)
		if aliasName != "" && targetName != "" {
			resolved := ResolveAlias(targetName, d.Compiler.Module)
			d.Compiler.Module.Aliases[aliasName] = resolved
		}
	}
	return nil
}

// Attribute processes an attribute declaration.
func (d *DeclInterp) Attribute(node ast.Node) error {
	if attr, ok := node.(*ast.AttributeDef); ok {
		name := extractSortName(attr.Name)
		if name != "" {
			d.Compiler.Module.Attributes[name] = attr.Value
		}
	}
	return nil
}

// Progress processes a progress declaration.
func (d *DeclInterp) Progress(node ast.Node) error {
	// TODO: implement progress handling
	return nil
}

// Private processes a private declaration.
func (d *DeclInterp) Private(node ast.Node) error {
	if atom, ok := node.(*ast.Atom); ok {
		d.Compiler.Module.Privates[atom.Rep] = true
	}
	return nil
}

// Schema processes a schema declaration.
func (d *DeclInterp) Schema(node ast.Node) error {
	// TODO: implement schema handling
	return nil
}

// Instantiate processes an instantiation declaration.
func (d *DeclInterp) Instantiate(node ast.Node) error {
	// TODO: implement instantiation handling (complex, stubbed)
	return nil
}

// Proof processes a proof declaration.
func (d *DeclInterp) Proof(node ast.Node) error {
	// TODO: implement proof handling (already in proof/ package)
	return nil
}

// Named processes a named declaration.
func (d *DeclInterp) Named(node ast.Node) error {
	// TODO: implement named declaration handling
	return nil
}

// Theorem processes a theorem declaration.
func (d *DeclInterp) Theorem(node ast.Node) error {
	// TODO: implement theorem handling
	return nil
}

// Assert processes an assert declaration.
func (d *DeclInterp) Assert(node ast.Node) error {
	// TODO: implement assert declaration handling
	return nil
}
