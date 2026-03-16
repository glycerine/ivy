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
// Corresponds to Python IvyDomainSetup.derived.
func (d *DeclInterp) Derived(node ast.Node) error {
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
	var compiled lg.Node
	if isSchema {
		compiled, err = d.Compiler.CompileDefnSchema(&ast.DefinitionSchema{Definition: *defNode})
	} else {
		compiled, err = d.Compiler.CompileDefn(defNode)
	}
	if err != nil {
		return err
	}

	// Remove the temporary symbol and re-add with inferred sort
	delete(d.Compiler.Sig.Symbols, sym.Name)
	if def, ok := compiled.(*il.Definition); ok {
		definesNode := def.Defines()
		if cnst, ok := definesNode.(*lg.Const); ok {
			d.Compiler.AddSymbol(cnst.Name, cnst.CSort, d.Compiler.Sig)
		}
	}

	// Add to module
	mlf := &module.LabeledFormula{
		Formula: compiled,
		Lineno:  lf.GetLineno().Line,
	}
	d.Compiler.Module.LabeledProps = append(d.Compiler.Module.LabeledProps, mlf)
	d.LastFact = compiled
	d.Compiler.Module.SymbolOrder = append(d.Compiler.Module.SymbolOrder, sym)
	return nil
}

// DefinitionDecl processes a definition declaration.
// Corresponds to Python IvyDomainSetup.definition.
func (d *DeclInterp) DefinitionDecl(node ast.Node) error {
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
	var compiled lg.Node
	var err error
	if isSchemaD {
		compiled, err = d.Compiler.CompileDefnSchema(&ast.DefinitionSchema{Definition: *defNode})
	} else {
		compiled, err = d.Compiler.CompileDefn(defNode)
	}
	if err != nil {
		return err
	}
	mlf := &module.LabeledFormula{
		Formula: compiled,
		Lineno:  lf.GetLineno().Line,
	}
	d.Compiler.Module.LabeledProps = append(d.Compiler.Module.LabeledProps, mlf)
	d.LastFact = compiled

	// Add the defined symbol if not already in the signature
	if def, ok := compiled.(*il.Definition); ok {
		definesNode := def.Defines()
		if cnst, ok := definesNode.(*lg.Const); ok {
			if _, exists := d.Compiler.Sig.Symbols[cnst.Name]; !exists {
				d.Compiler.AddSymbol(cnst.Name, cnst.CSort, d.Compiler.Sig)
			}
			d.Compiler.Module.SymbolOrder = append(d.Compiler.Module.SymbolOrder, cnst)
		}
	}
	return nil
}

// Action processes an action declaration.
// Corresponds to Python IvyARGSetup.action.
func (d *DeclInterp) Action(node ast.Node) error {
	actDef, ok := node.(*ast.ActionDef)
	if !ok {
		return nil
	}
	name := actDef.Defines()
	compiled, err := d.Compiler.CompileAction(actDef)
	if err != nil {
		return err
	}
	d.Compiler.Module.Actions[name] = compiled
	d.Compiler.Module.PublicActions[name] = true
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
// Module instantiation is complex and deferred; we store the raw node.
func (d *DeclInterp) ModuleD(node ast.Node) error {
	// Module declarations define parameterized modules. They are not
	// compiled eagerly; they are stored and instantiated later when
	// an "instantiate" declaration references them.
	// Store as-is in the module's schemata keyed by name.
	if atom, ok := node.(*ast.Atom); ok {
		d.Compiler.Module.Schemata[atom.Rep] = node
	}
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
// Corresponds to Python IvyARGSetup.export.
func (d *DeclInterp) Export(node ast.Node) error {
	expDef, ok := node.(*ast.ExportDef)
	if !ok {
		return nil
	}
	d.Compiler.Module.Exports = append(d.Compiler.Module.Exports, expDef)
	return nil
}

// Import processes an import declaration.
// Corresponds to Python IvyARGSetup.import_.
func (d *DeclInterp) Import(node ast.Node) error {
	impDef, ok := node.(*ast.ImportDef)
	if !ok {
		return nil
	}
	d.Compiler.Module.Imports = append(d.Compiler.Module.Imports, impDef)
	return nil
}

// Isolate processes an isolate declaration.
// Corresponds to Python IvyARGSetup.isolate.
func (d *DeclInterp) Isolate(node ast.Node) error {
	isoDef, ok := node.(*ast.IsolateDef)
	if !ok {
		return nil
	}
	isoName := isoDef.IsoName()
	d.Compiler.Module.Isolates[isoName] = isoDef
	return nil
}

// Interpret processes a type interpretation.
// Corresponds to Python IvyDomainSetup.interpret.
// The full interpret logic is complex (ranges, enums, solver sorts);
// here we handle the common cases.
func (d *DeclInterp) Interpret(node ast.Node) error {
	lf, ok := node.(*ast.LabeledFormula)
	if !ok {
		return nil
	}
	fmla := lf.Formula
	defNode, ok := fmla.(*ast.Definition)
	if !ok {
		return nil
	}
	lhs := ResolveAlias(extractSortName(defNode.Lhs), d.Compiler.Module)
	rhs := defNode.Rhs

	// Store interpretation in the module
	d.Compiler.Module.Interps[lhs] = append(d.Compiler.Module.Interps[lhs], node)

	// Handle native type interpretation
	if _, ok := rhs.(*ast.NativeType); ok {
		d.Compiler.Module.NativeTypes[lhs] = rhs
		return nil
	}

	// Handle range interpretation
	if rng, ok := rhs.(*ast.Range); ok {
		lo := fmt.Sprint(rng.Lo)
		hi := fmt.Sprint(rng.Hi)
		sort := &lg.RangeSort{Name: lhs, Lb: lo, Ub: hi}
		d.Compiler.Sig.Interp[lhs] = sort
		return nil
	}

	// Handle enumerated sort interpretation
	if enumSort, ok := rhs.(*ast.EnumeratedSort); ok {
		ext := enumSort.Extension()
		sort := &lg.EnumeratedSort{Name: lhs, Extension: ext}
		d.Compiler.Sig.Interp[lhs] = sort
		// Register constructors
		for _, c := range ext {
			if existingSort, hasSig := d.Compiler.Sig.Sorts[lhs]; hasSig {
				d.Compiler.Sig.Symbols[c] = &il.SymbolEntry{Sort: existingSort}
			}
		}
		return nil
	}

	// For simple symbol/sort interpretations, store the name
	rhsName := extractSortName(rhs)
	if rhsName != "" {
		d.Compiler.Sig.Interp[lhs] = rhsName
	}
	return nil
}

// Mixin processes a mixin declaration.
// Corresponds to Python IvyARGSetup.mixin.
func (d *DeclInterp) Mixin(node ast.Node) error {
	// Mixins define before/after/implement hooks on actions.
	// Extract the mixee name and register the mixin.
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
	d.Compiler.Module.Mixins[mixeeName] = append(d.Compiler.Module.Mixins[mixeeName], node)
	return nil
}

// Delegate processes a delegate declaration.
// Corresponds to Python IvyARGSetup.delegate.
func (d *DeclInterp) Delegate(node ast.Node) error {
	d.Compiler.Module.Delegates = append(d.Compiler.Module.Delegates, node)
	return nil
}

// Native processes a native code declaration.
// Corresponds to Python IvyARGSetup.native.
func (d *DeclInterp) Native(node ast.Node) error {
	// Native declarations embed target-language code. We store them as-is;
	// the code generation backend will process them later.
	d.Compiler.Module.Natives = append(d.Compiler.Module.Natives, node)
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
// Corresponds to Python IvyDomainSetup.progress.
func (d *DeclInterp) Progress(node ast.Node) error {
	// Progress properties relate a relation to a temporal progress condition.
	// Compile with sort inference and store.
	compiled, err := d.Compiler.SortifyWithInference(node)
	if err != nil {
		return err
	}
	d.Compiler.Module.Progress = append(d.Compiler.Module.Progress, compiled)
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
// Corresponds to Python IvyDomainSetup.schema.
func (d *DeclInterp) Schema(node ast.Node) error {
	// A schema has a defn with args[0]=name, args[1]=body.
	// If the body is a SchemaBody, compile it and store as a labeled formula.
	// Otherwise store the raw schema.
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
		compiled, err := d.Compiler.CompileNode(df.Rhs)
		if err != nil {
			return err
		}
		label := ast.NewAtom(defName)
		clf := &module.LabeledFormula{
			Formula: compiled,
			Lineno:  lf.GetLineno().Line,
		}
		_ = label
		d.Compiler.Module.Schemata[defName] = clf
	} else {
		d.Compiler.Module.Schemata[defName] = node
	}
	return nil
}

// Instantiate processes an instantiation declaration.
// Corresponds to Python IvyDomainSetup.instantiate.
func (d *DeclInterp) Instantiate(node ast.Node) error {
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

	schema, ok := d.Compiler.Module.Schemata[instName]
	if !ok {
		return &lg.IvyError{Msg: fmt.Sprintf("%s undefined in instantiation", instName)}
	}

	// Store the instantiation for later processing
	d.Compiler.Module.Instantiations = append(d.Compiler.Module.Instantiations,
		struct {
			Schema interface{}
			Inst   ast.Node
		}{schema, node})
	return nil
}

// Proof processes a proof declaration.
// Corresponds to Python IvyDomainSetup.proof.
func (d *DeclInterp) Proof(node ast.Node) error {
	// If the proof is a labeled formula, it has its own label.
	if lf, ok := node.(*ast.LabeledFormula); ok {
		// Compile the proof body as a tactic (not as a logic node).
		compiledProof, err := d.Compiler.CompileTactic(lf.Formula)
		if err != nil {
			return err
		}
		proofLF := &module.LabeledFormula{
			Label:   nil,
			Formula: nil,
		}
		if lf.Label != nil {
			label, _ := d.Compiler.CompileNode(lf.Label)
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

	lastLF := &module.LabeledFormula{
		Formula: d.LastFact,
	}
	d.Compiler.Module.Proofs = append(d.Compiler.Module.Proofs, module.ProofEntry{
		Formula: lastLF,
		Proof:   compiled,
	})
	return nil
}

// Named processes a named declaration.
// Corresponds to Python IvyDomainSetup.named.
// A named declaration gives a name to an existential property.
func (d *DeclInterp) Named(node ast.Node) error {
	lhs, ok := node.(*ast.Atom)
	if !ok {
		return nil
	}
	if d.LastFact == nil {
		return &lg.IvyError{Msg: "named declaration without preceding property"}
	}

	// The last fact should be an existential formula.
	// We extract the existential's range sort and create a function symbol.
	cond := il.DropUniversals(d.LastFact)
	if !il.IsExists(cond) {
		return &lg.IvyError{Msg: "property is not existential"}
	}

	// Get the existential variable's sort as the range
	ex, ok := cond.(*lg.Exists)
	if !ok || len(ex.Variables) != 1 {
		return &lg.IvyError{Msg: "property is not existential"}
	}
	rng := ex.Variables[0].VSort

	// Build domain sorts from the lhs parameters
	var domSorts []lg.Sort
	for _, arg := range lhs.Terms {
		compiled, err := d.Compiler.CompileNode(arg)
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

	lastLF := &module.LabeledFormula{Formula: d.LastFact}
	d.Compiler.Module.Named = append(d.Compiler.Module.Named, module.NamedEntry{
		Formula: lastLF,
		Name:    sym,
	})
	return nil
}

// Theorem processes a theorem declaration.
// Corresponds to Python IvyDomainSetup.theorem.
func (d *DeclInterp) Theorem(node ast.Node) error {
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
		compiled, err := d.Compiler.CompileNode(df.Rhs)
		if err != nil {
			return err
		}
		mlf := &module.LabeledFormula{
			Formula: compiled,
			Lineno:  lf.GetLineno().Line,
		}
		d.Compiler.Module.LabeledProps = append(d.Compiler.Module.LabeledProps, mlf)
		d.Compiler.Module.Theorems[defName] = compiled
		d.LastFact = compiled
	}
	return nil
}

// Assert processes an assert declaration.
// Corresponds to Python IvyARGSetup._assert.
func (d *DeclInterp) Assert(node ast.Node) error {
	lf, ok := node.(*ast.LabeledFormula)
	if !ok {
		return nil
	}
	compiled, err := d.Compiler.SortifyWithInference(lf.Formula)
	if err != nil {
		return err
	}
	mlf := &module.LabeledFormula{
		Formula: compiled,
		Lineno:  lf.GetLineno().Line,
	}
	d.Compiler.Module.Assertions = append(d.Compiler.Module.Assertions, mlf)
	return nil
}
