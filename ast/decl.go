package ast

import (
	"fmt"
	"strings"
)

// --- Declaration types ---

// LabeledFormula associates a label with a formula (used in axioms, properties, etc.).
type LabeledFormula struct {
	Base
	Label      Node // label (may be nil)
	Formula    Node
	ID         int64
	Temporal   Node // optional temporal annotation
	Explicit   bool
	IsDefinition bool
	Assumed    bool
	Unprovable bool
}

func NewLabeledFormula(label, formula Node) *LabeledFormula {
	return &LabeledFormula{
		Label:   label,
		Formula: formula,
		ID:      nextLFID(),
	}
}

func (lf *LabeledFormula) Args() []Node { return []Node{lf.Label, lf.Formula} }
func (lf *LabeledFormula) Clone(args []Node) Node {
	c := &LabeledFormula{
		Base:         lf.Base,
		Label:        args[0],
		Formula:      args[1],
		ID:           lf.ID, // preserve ID by default
		Temporal:     lf.Temporal,
		Explicit:     lf.Explicit,
		IsDefinition: lf.IsDefinition,
		Assumed:      lf.Assumed,
		Unprovable:   lf.Unprovable,
	}
	return c
}
func (lf *LabeledFormula) CloneWithFreshID(args []Node) *LabeledFormula {
	c := lf.Clone(args).(*LabeledFormula)
	c.ID = nextLFID()
	return c
}
func (lf *LabeledFormula) String() string {
	if lf.Label != nil {
		return "[" + fmt.Sprint(lf.Label) + "] " + fmt.Sprint(lf.Formula)
	}
	return fmt.Sprint(lf.Formula)
}
func (lf *LabeledFormula) LabelName() string {
	if lf.Label == nil {
		return ""
	}
	if a, ok := lf.Label.(*Atom); ok {
		return a.Relname()
	}
	return fmt.Sprint(lf.Label)
}
func (lf *LabeledFormula) Rename(s string) *LabeledFormula {
	newLabel := lf.Label
	if a, ok := lf.Label.(*Atom); ok {
		newLabel = a.Rename(s)
	}
	return lf.Clone([]Node{newLabel, lf.Formula}).(*LabeledFormula)
}

// Decl is the base for all declaration nodes.
// In Python, Decl has args and attributes. In Go, each specific Decl type
// has its own fields. We provide a common DeclBase.
type DeclBase struct {
	Base
	DeclArgs   []Node
	Attributes []Node
	Common     Node // optional common block
}

func (d *DeclBase) Args() []Node            { return d.DeclArgs }
func (d *DeclBase) GetAttributes() []Node   { return d.Attributes }
func (d *DeclBase) SetAttributes(a []Node)  { d.Attributes = a }
func (d *DeclBase) Defines() []string       { return nil }

// ModuleDecl declares a module.
type ModuleDecl struct {
	DeclBase
	FormalParams []Node // formal parameters for instantiation
	BodyDecls    []Node // parsed body declarations for instantiation
}

func NewModuleDecl(args ...Node) *ModuleDecl {
	return &ModuleDecl{DeclBase: DeclBase{DeclArgs: args}}
}

func (d *ModuleDecl) Clone(args []Node) Node {
	return &ModuleDecl{DeclBase: DeclBase{Base: d.Base, DeclArgs: args, Attributes: d.Attributes, Common: d.Common}}
}
func (d *ModuleDecl) String() string {
	parts := make([]string, len(d.DeclArgs))
	for i, a := range d.DeclArgs {
		parts[i] = fmt.Sprint(a)
	}
	return "module " + strings.Join(parts, ", ")
}

// MacroDecl declares a macro.
type MacroDecl struct {
	DeclBase
}

func NewMacroDecl(args ...Node) *MacroDecl {
	return &MacroDecl{DeclBase: DeclBase{DeclArgs: args}}
}

func (d *MacroDecl) Clone(args []Node) Node {
	return &MacroDecl{DeclBase: DeclBase{Base: d.Base, DeclArgs: args, Attributes: d.Attributes, Common: d.Common}}
}
func (d *MacroDecl) String() string { return "macro" }

// ObjectDecl declares an object.
type ObjectDecl struct {
	DeclBase
}

func NewObjectDecl(args ...Node) *ObjectDecl {
	return &ObjectDecl{DeclBase: DeclBase{DeclArgs: args}}
}

func (d *ObjectDecl) Clone(args []Node) Node {
	return &ObjectDecl{DeclBase: DeclBase{Base: d.Base, DeclArgs: args, Attributes: d.Attributes, Common: d.Common}}
}
func (d *ObjectDecl) String() string { return "object" }

// ActionDecl declares one or more actions.
type ActionDecl struct {
	DeclBase
}

func NewActionDecl(args ...Node) *ActionDecl {
	return &ActionDecl{DeclBase: DeclBase{DeclArgs: args}}
}

func (d *ActionDecl) Clone(args []Node) Node {
	return &ActionDecl{DeclBase: DeclBase{Base: d.Base, DeclArgs: args, Attributes: d.Attributes, Common: d.Common}}
}
func (d *ActionDecl) String() string { return "action" }

// ActionDef defines an action with formal params and returns.
type ActionDef struct {
	Base
	Name          Node
	Body          Node
	FormalParams  []Node
	FormalReturns []Node
}

func NewActionDef(name, body Node, params, returns []Node) *ActionDef {
	return &ActionDef{Name: name, Body: body, FormalParams: params, FormalReturns: returns}
}

func (a *ActionDef) Args() []Node { return []Node{a.Name, a.Body} }
func (a *ActionDef) Clone(args []Node) Node {
	return &ActionDef{Base: a.Base, Name: args[0], Body: args[1],
		FormalParams: a.FormalParams, FormalReturns: a.FormalReturns}
}
func (a *ActionDef) String() string {
	parts := make([]string, len(a.FormalParams))
	for i, p := range a.FormalParams {
		parts[i] = fmt.Sprint(p)
	}
	return fmt.Sprint(a.Name) + "(" + strings.Join(parts, ",") + ") = " + fmt.Sprint(a.Body)
}
func (a *ActionDef) Defines() string {
	if atom, ok := a.Name.(*Atom); ok {
		return atom.Relname()
	}
	return fmt.Sprint(a.Name)
}

// RelationDecl declares a relation.
type RelationDecl struct {
	DeclBase
}

func NewRelationDecl(args ...Node) *RelationDecl {
	return &RelationDecl{DeclBase: DeclBase{DeclArgs: args}}
}

func (d *RelationDecl) Clone(args []Node) Node {
	return &RelationDecl{DeclBase: DeclBase{Base: d.Base, DeclArgs: args, Attributes: d.Attributes, Common: d.Common}}
}
func (d *RelationDecl) String() string { return "relation" }

// ConstantDecl declares a constant (individual).
type ConstantDecl struct {
	DeclBase
}

func NewConstantDecl(args ...Node) *ConstantDecl {
	return &ConstantDecl{DeclBase: DeclBase{DeclArgs: args}}
}

func (d *ConstantDecl) Clone(args []Node) Node {
	return &ConstantDecl{DeclBase: DeclBase{Base: d.Base, DeclArgs: args, Attributes: d.Attributes, Common: d.Common}}
}
func (d *ConstantDecl) String() string { return "individual" }

// ParameterDecl declares a parameter.
type ParameterDecl struct {
	DeclBase
}

func NewParameterDecl(args ...Node) *ParameterDecl {
	return &ParameterDecl{DeclBase: DeclBase{DeclArgs: args}}
}

func (d *ParameterDecl) Clone(args []Node) Node {
	return &ParameterDecl{DeclBase: DeclBase{Base: d.Base, DeclArgs: args, Attributes: d.Attributes, Common: d.Common}}
}
func (d *ParameterDecl) String() string { return "parameter" }

// FreshConstantDecl declares a fresh constant.
type FreshConstantDecl struct {
	ConstantDecl
}

// DestructorDecl declares a destructor.
type DestructorDecl struct {
	DeclBase
}

func NewDestructorDecl(args ...Node) *DestructorDecl {
	return &DestructorDecl{DeclBase: DeclBase{DeclArgs: args}}
}

func (d *DestructorDecl) Clone(args []Node) Node {
	return &DestructorDecl{DeclBase: DeclBase{Base: d.Base, DeclArgs: args, Attributes: d.Attributes, Common: d.Common}}
}
func (d *DestructorDecl) String() string { return "destructor" }

// ConstructorDecl declares a constructor.
type ConstructorDecl struct {
	DeclBase
}

func NewConstructorDecl(args ...Node) *ConstructorDecl {
	return &ConstructorDecl{DeclBase: DeclBase{DeclArgs: args}}
}

func (d *ConstructorDecl) Clone(args []Node) Node {
	return &ConstructorDecl{DeclBase: DeclBase{Base: d.Base, DeclArgs: args, Attributes: d.Attributes, Common: d.Common}}
}
func (d *ConstructorDecl) String() string { return "constructor" }

// TypeDecl declares a type.
type TypeDecl struct {
	DeclBase
}

func NewTypeDecl(args ...Node) *TypeDecl {
	return &TypeDecl{DeclBase: DeclBase{DeclArgs: args}}
}

func (d *TypeDecl) Clone(args []Node) Node {
	return &TypeDecl{DeclBase: DeclBase{Base: d.Base, DeclArgs: args, Attributes: d.Attributes, Common: d.Common}}
}
func (d *TypeDecl) String() string { return "type" }

// TypeDef defines a named type.
type TypeDef struct {
	Base
	Name   Node
	Value  Node
	Finite bool
}

func NewTypeDef(name, value Node) *TypeDef {
	return &TypeDef{Name: name, Value: value}
}

func (t *TypeDef) Args() []Node { return []Node{t.Name, t.Value} }
func (t *TypeDef) Clone(args []Node) Node {
	return &TypeDef{Base: t.Base, Name: args[0], Value: args[1], Finite: t.Finite}
}
func (t *TypeDef) String() string {
	prefix := ""
	if t.Finite {
		prefix = "finite "
	}
	return prefix + fmt.Sprint(t.Name) + " = " + fmt.Sprint(t.Value)
}
func (t *TypeDef) Defines() []string {
	var syms []string
	if sym, ok := t.Name.(*Symbol); ok {
		syms = append(syms, sym.Rep)
	} else if a, ok := t.Name.(*Atom); ok {
		syms = append(syms, a.Rep)
	}
	// Add names defined by the value (e.g., enum elements)
	type definer interface{ Defines() []string }
	if d, ok := t.Value.(definer); ok {
		syms = append(syms, d.Defines()...)
	}
	return syms
}

// GhostTypeDef is a ghost (specification-only) type definition.
type GhostTypeDef struct {
	TypeDef
}

func (g *GhostTypeDef) Clone(args []Node) Node {
	return &GhostTypeDef{TypeDef: *g.TypeDef.Clone(args).(*TypeDef)}
}

// VariantDecl declares a variant type.
type VariantDecl struct {
	DeclBase
}

func NewVariantDecl(args ...Node) *VariantDecl {
	return &VariantDecl{DeclBase: DeclBase{DeclArgs: args}}
}

func (d *VariantDecl) Clone(args []Node) Node {
	return &VariantDecl{DeclBase: DeclBase{Base: d.Base, DeclArgs: args, Attributes: d.Attributes, Common: d.Common}}
}
func (d *VariantDecl) String() string { return "variant" }

// VariantDef defines a variant of a type.
type VariantDef struct {
	Base
	Name  Node
	VSort Node
}

func NewVariantDef(name, sort Node) *VariantDef {
	return &VariantDef{Name: name, VSort: sort}
}

func (v *VariantDef) Args() []Node { return []Node{v.Name, v.VSort} }
func (v *VariantDef) Clone(args []Node) Node {
	return &VariantDef{Base: v.Base, Name: args[0], VSort: args[1]}
}
func (v *VariantDef) String() string { return fmt.Sprint(v.Name) + " of " + fmt.Sprint(v.VSort) }

// AxiomDecl declares an axiom.
type AxiomDecl struct {
	DeclBase
}

func NewAxiomDecl(args ...Node) *AxiomDecl {
	return &AxiomDecl{DeclBase: DeclBase{DeclArgs: args}}
}

func (d *AxiomDecl) Clone(args []Node) Node {
	return &AxiomDecl{DeclBase: DeclBase{Base: d.Base, DeclArgs: args, Attributes: d.Attributes, Common: d.Common}}
}
func (d *AxiomDecl) String() string { return "axiom" }

// PropertyDecl declares a property (provable axiom).
type PropertyDecl struct {
	AxiomDecl
}

func NewPropertyDecl(args ...Node) *PropertyDecl {
	return &PropertyDecl{AxiomDecl: AxiomDecl{DeclBase: DeclBase{DeclArgs: args}}}
}

func (d *PropertyDecl) Clone(args []Node) Node {
	return &PropertyDecl{AxiomDecl: *d.AxiomDecl.Clone(args).(*AxiomDecl)}
}
func (d *PropertyDecl) String() string { return "property" }

// ConjectureDecl declares a conjecture.
type ConjectureDecl struct {
	DeclBase
}

func NewConjectureDecl(args ...Node) *ConjectureDecl {
	return &ConjectureDecl{DeclBase: DeclBase{DeclArgs: args}}
}

func (d *ConjectureDecl) Clone(args []Node) Node {
	return &ConjectureDecl{DeclBase: DeclBase{Base: d.Base, DeclArgs: args, Attributes: d.Attributes, Common: d.Common}}
}
func (d *ConjectureDecl) String() string { return "conjecture" }

// ProofDecl declares a proof.
type ProofDecl struct {
	DeclBase
}

func NewProofDecl(args ...Node) *ProofDecl {
	return &ProofDecl{DeclBase: DeclBase{DeclArgs: args}}
}

func (d *ProofDecl) Clone(args []Node) Node {
	return &ProofDecl{DeclBase: DeclBase{Base: d.Base, DeclArgs: args, Attributes: d.Attributes, Common: d.Common}}
}
func (d *ProofDecl) String() string { return "proof" }

// NamedDecl declares a named entity.
type NamedDecl struct {
	DeclBase
}

func NewNamedDecl(args ...Node) *NamedDecl {
	return &NamedDecl{DeclBase: DeclBase{DeclArgs: args}}
}

func (d *NamedDecl) Clone(args []Node) Node {
	return &NamedDecl{DeclBase: DeclBase{Base: d.Base, DeclArgs: args, Attributes: d.Attributes, Common: d.Common}}
}
func (d *NamedDecl) String() string { return "named" }

// SchemaDecl declares a proof schema.
type SchemaDecl struct {
	DeclBase
}

func NewSchemaDecl(args ...Node) *SchemaDecl {
	return &SchemaDecl{DeclBase: DeclBase{DeclArgs: args}}
}

func (d *SchemaDecl) Clone(args []Node) Node {
	return &SchemaDecl{DeclBase: DeclBase{Base: d.Base, DeclArgs: args, Attributes: d.Attributes, Common: d.Common}}
}
func (d *SchemaDecl) String() string { return "schema" }

// SchemaBody holds premises and a conclusion.
type SchemaBody struct {
	Base
	Elems []Node // premises + conclusion (last elem)
}

func NewSchemaBody(elems ...Node) *SchemaBody { return &SchemaBody{Elems: elems} }

func (s *SchemaBody) Args() []Node           { return s.Elems }
func (s *SchemaBody) Clone(args []Node) Node { return &SchemaBody{Base: s.Base, Elems: args} }
func (s *SchemaBody) String() string          { return "{...}" }
func (s *SchemaBody) Prems() []Node {
	if len(s.Elems) == 0 {
		return nil
	}
	return s.Elems[:len(s.Elems)-1]
}
func (s *SchemaBody) Conc() Node {
	if len(s.Elems) == 0 {
		return nil
	}
	return s.Elems[len(s.Elems)-1]
}

// Schema wraps a Definition for schema declarations.
// Matches Python's Schema(AST) from ivy_actions.py.
type Schema struct {
	Base
	Defn      Node   // the Definition
	Fresh     []Node // fresh variables
	Instances []Node // instantiation records
}

func NewSchema(defn Node) *Schema {
	return &Schema{Defn: defn}
}

func (s *Schema) Args() []Node           { return []Node{s.Defn} }
func (s *Schema) Clone(args []Node) Node {
	ns := &Schema{Base: s.Base, Fresh: s.Fresh, Instances: s.Instances}
	if len(args) > 0 {
		ns.Defn = args[0]
	}
	return ns
}
func (s *Schema) String() string {
	res := fmt.Sprintf("%v", s.Defn)
	if len(s.Fresh) > 0 {
		res += " fresh "
		for i, f := range s.Fresh {
			if i > 0 {
				res += ","
			}
			res += fmt.Sprint(f)
		}
	}
	return res
}
func (s *Schema) Defines() string {
	if d, ok := s.Defn.(*Definition); ok {
		return d.Defines()
	}
	return ""
}

// SchemaCompiler is the interface needed by Schema.GetInstance to compile AST to logic.
// Python: compile_with_sort_inference is monkey-patched onto AST; we use an explicit interface.
type SchemaCompiler interface {
	CompileWithSortInference(node Node) (Node, error)
}

// SchemaClauseConverter converts logic formulas to clauses.
// Python: formula_to_clauses(fmla)
type SchemaClauseConverter interface {
	FormulaToClauses(fmla Node) (Node, error)
}

// GetInstance creates an instance of this schema with the given parameters.
// Python: Schema.get_instance(self, params, to_clauses=True)
func (s *Schema) GetInstance(params []Node, compiler SchemaCompiler, clauseConverter SchemaClauseConverter, toClauses bool) (Node, error) {
	defn, ok := s.Defn.(*Definition)
	if !ok {
		return nil, fmt.Errorf("schema defn is not a Definition")
	}
	// defn.Lhs is the atom with formal parameters
	var lhsArgs []Node
	if atom, ok := defn.Lhs.(*Atom); ok {
		lhsArgs = atom.Terms
	}
	if len(params) != len(lhsArgs) {
		return nil, fmt.Errorf("schema parameter count mismatch: expected %d, got %d", len(lhsArgs), len(params))
	}
	// Build substitution: formal param name → actual param name
	subst := make(map[string]string)
	for i, formal := range lhsArgs {
		var formalName string
		switch f := formal.(type) {
		case *Atom:
			formalName = f.Rep
		case *Symbol:
			formalName = f.Rep
		case *Variable:
			formalName = f.Rep
		default:
			formalName = fmt.Sprint(formal)
		}
		var actualName string
		switch a := params[i].(type) {
		case *Atom:
			actualName = a.Rep
		case *Symbol:
			actualName = a.Rep
		case *Variable:
			actualName = a.Rep
		default:
			actualName = fmt.Sprint(params[i])
		}
		subst[formalName] = actualName
	}
	// Rewrite the body with the substitution
	rewriter := NewAstRewriteSubstPrefix(subst, nil)
	rewrittenBody := AstRewrite(defn.Rhs, rewriter)

	// Compile with sort inference
	fmla, err := compiler.CompileWithSortInference(rewrittenBody)
	if err != nil {
		return nil, err
	}

	if toClauses && clauseConverter != nil {
		return clauseConverter.FormulaToClauses(fmla)
	}
	return fmla, nil
}

// Instantiate adds an instance to this schema's instance list.
// Python: Schema.instantiate(self, params)
func (s *Schema) Instantiate(params []Node, compiler SchemaCompiler) {
	inst, err := s.GetInstance(params, compiler, nil, false)
	if err == nil {
		s.Instances = append(s.Instances, inst)
	}
}

// TheoremDecl declares a theorem.
type TheoremDecl struct {
	DeclBase
}

func NewTheoremDecl(args ...Node) *TheoremDecl {
	return &TheoremDecl{DeclBase: DeclBase{DeclArgs: args}}
}

func (d *TheoremDecl) Clone(args []Node) Node {
	return &TheoremDecl{DeclBase: DeclBase{Base: d.Base, DeclArgs: args, Attributes: d.Attributes, Common: d.Common}}
}
func (d *TheoremDecl) String() string { return "theorem" }

// DerivedDecl declares a derived relation/function.
type DerivedDecl struct {
	DeclBase
}

func NewDerivedDecl(args ...Node) *DerivedDecl {
	return &DerivedDecl{DeclBase: DeclBase{DeclArgs: args}}
}

func (d *DerivedDecl) Clone(args []Node) Node {
	return &DerivedDecl{DeclBase: DeclBase{Base: d.Base, DeclArgs: args, Attributes: d.Attributes, Common: d.Common}}
}
func (d *DerivedDecl) String() string { return "derived" }

// DefinitionDecl declares a definition.
type DefinitionDecl struct {
	DeclBase
}

func NewDefinitionDecl(args ...Node) *DefinitionDecl {
	return &DefinitionDecl{DeclBase: DeclBase{DeclArgs: args}}
}

func (d *DefinitionDecl) Clone(args []Node) Node {
	return &DefinitionDecl{DeclBase: DeclBase{Base: d.Base, DeclArgs: args, Attributes: d.Attributes, Common: d.Common}}
}
func (d *DefinitionDecl) String() string { return "definition" }

// ProgressDecl declares a progress property.
type ProgressDecl struct {
	DeclBase
}

func NewProgressDecl(args ...Node) *ProgressDecl {
	return &ProgressDecl{DeclBase: DeclBase{DeclArgs: args}}
}

func (d *ProgressDecl) Clone(args []Node) Node {
	return &ProgressDecl{DeclBase: DeclBase{Base: d.Base, DeclArgs: args, Attributes: d.Attributes, Common: d.Common}}
}
func (d *ProgressDecl) String() string { return "progress" }

// RelyDecl declares a rely condition.
type RelyDecl struct {
	DeclBase
}

func NewRelyDecl(args ...Node) *RelyDecl {
	return &RelyDecl{DeclBase: DeclBase{DeclArgs: args}}
}

func (d *RelyDecl) Clone(args []Node) Node {
	return &RelyDecl{DeclBase: DeclBase{Base: d.Base, DeclArgs: args, Attributes: d.Attributes, Common: d.Common}}
}
func (d *RelyDecl) String() string { return "rely" }

// MixOrdDecl declares a mixin ordering.
type MixOrdDecl struct {
	DeclBase
}

func NewMixOrdDecl(args ...Node) *MixOrdDecl {
	return &MixOrdDecl{DeclBase: DeclBase{DeclArgs: args}}
}

func (d *MixOrdDecl) Clone(args []Node) Node {
	return &MixOrdDecl{DeclBase: DeclBase{Base: d.Base, DeclArgs: args, Attributes: d.Attributes, Common: d.Common}}
}
func (d *MixOrdDecl) String() string { return "mixord" }

// ConceptDecl declares a concept.
type ConceptDecl struct {
	DeclBase
}

func NewConceptDecl(args ...Node) *ConceptDecl {
	return &ConceptDecl{DeclBase: DeclBase{DeclArgs: args}}
}

func (d *ConceptDecl) Clone(args []Node) Node {
	return &ConceptDecl{DeclBase: DeclBase{Base: d.Base, DeclArgs: args, Attributes: d.Attributes, Common: d.Common}}
}
func (d *ConceptDecl) String() string { return "concept" }

// InitDecl declares initialization code.
type InitDecl struct {
	DeclBase
}

func NewInitDecl(args ...Node) *InitDecl {
	return &InitDecl{DeclBase: DeclBase{DeclArgs: args}}
}

func (d *InitDecl) Clone(args []Node) Node {
	return &InitDecl{DeclBase: DeclBase{Base: d.Base, DeclArgs: args, Attributes: d.Attributes, Common: d.Common}}
}
func (d *InitDecl) String() string { return "init" }

// StateDecl declares state variables.
type StateDecl struct {
	DeclBase
}

func NewStateDecl(args ...Node) *StateDecl {
	return &StateDecl{DeclBase: DeclBase{DeclArgs: args}}
}

func (d *StateDecl) Clone(args []Node) Node {
	return &StateDecl{DeclBase: DeclBase{Base: d.Base, DeclArgs: args, Attributes: d.Attributes, Common: d.Common}}
}
func (d *StateDecl) String() string { return "state" }

// UpdateDecl declares an update.
type UpdateDecl struct {
	DeclBase
}

func NewUpdateDecl(args ...Node) *UpdateDecl {
	return &UpdateDecl{DeclBase: DeclBase{DeclArgs: args}}
}

func (d *UpdateDecl) Clone(args []Node) Node {
	return &UpdateDecl{DeclBase: DeclBase{Base: d.Base, DeclArgs: args, Attributes: d.Attributes, Common: d.Common}}
}
func (d *UpdateDecl) String() string { return "update" }

// AssertDecl declares an assertion.
type AssertDecl struct {
	DeclBase
}

func NewAssertDecl(args ...Node) *AssertDecl {
	return &AssertDecl{DeclBase: DeclBase{DeclArgs: args}}
}

func (d *AssertDecl) Clone(args []Node) Node {
	return &AssertDecl{DeclBase: DeclBase{Base: d.Base, DeclArgs: args, Attributes: d.Attributes, Common: d.Common}}
}
func (d *AssertDecl) String() string { return "assert" }

// InterpretDecl interprets a type.
type InterpretDecl struct {
	DeclBase
}

func NewInterpretDecl(args ...Node) *InterpretDecl {
	return &InterpretDecl{DeclBase: DeclBase{DeclArgs: args}}
}

func (d *InterpretDecl) Clone(args []Node) Node {
	return &InterpretDecl{DeclBase: DeclBase{Base: d.Base, DeclArgs: args, Attributes: d.Attributes, Common: d.Common}}
}
func (d *InterpretDecl) String() string { return "interpret" }

// --- Mixin declarations ---

// MixinDecl declares mixin relationships.
type MixinDecl struct {
	DeclBase
}

func NewMixinDecl(args ...Node) *MixinDecl {
	return &MixinDecl{DeclBase: DeclBase{DeclArgs: args}}
}

func (d *MixinDecl) Clone(args []Node) Node {
	return &MixinDecl{DeclBase: DeclBase{Base: d.Base, DeclArgs: args, Attributes: d.Attributes, Common: d.Common}}
}
func (d *MixinDecl) String() string { return "mixin" }

// MixinBeforeDef defines "X before Y".
type MixinBeforeDef struct {
	Base
	Mixer Node
	Mixee Node
}

func (m *MixinBeforeDef) Args() []Node { return []Node{m.Mixer, m.Mixee} }
func (m *MixinBeforeDef) Clone(args []Node) Node {
	return &MixinBeforeDef{Base: m.Base, Mixer: args[0], Mixee: args[1]}
}
func (m *MixinBeforeDef) String() string {
	return fmt.Sprint(m.Mixer) + " before " + fmt.Sprint(m.Mixee)
}

// MixinImplementDef defines "X implement Y".
type MixinImplementDef struct {
	Base
	Mixer Node
	Mixee Node
}

func (m *MixinImplementDef) Args() []Node { return []Node{m.Mixer, m.Mixee} }
func (m *MixinImplementDef) Clone(args []Node) Node {
	return &MixinImplementDef{Base: m.Base, Mixer: args[0], Mixee: args[1]}
}
func (m *MixinImplementDef) String() string {
	return fmt.Sprint(m.Mixer) + " implement " + fmt.Sprint(m.Mixee)
}

// MixinAfterDef defines "X after Y".
type MixinAfterDef struct {
	Base
	Mixer Node
	Mixee Node
}

func (m *MixinAfterDef) Args() []Node { return []Node{m.Mixer, m.Mixee} }
func (m *MixinAfterDef) Clone(args []Node) Node {
	return &MixinAfterDef{Base: m.Base, Mixer: args[0], Mixee: args[1]}
}
func (m *MixinAfterDef) String() string {
	return fmt.Sprint(m.Mixer) + " after " + fmt.Sprint(m.Mixee)
}

// --- Isolate declarations ---

// IsolateDecl declares an isolate.
type IsolateDecl struct {
	DeclBase
}

func NewIsolateDecl(args ...Node) *IsolateDecl {
	return &IsolateDecl{DeclBase: DeclBase{DeclArgs: args}}
}

func (d *IsolateDecl) Clone(args []Node) Node {
	return &IsolateDecl{DeclBase: DeclBase{Base: d.Base, DeclArgs: args, Attributes: d.Attributes, Common: d.Common}}
}
func (d *IsolateDecl) String() string { return "isolate" }

// IsolateDef defines an isolate with verified and present components.
type IsolateDef struct {
	Base
	Elems    []Node // args[0] = name, rest = verified + present
	WithArgs int    // number of "with" args at end
}

func (i *IsolateDef) Args() []Node           { return i.Elems }
func (i *IsolateDef) Clone(args []Node) Node {
	return &IsolateDef{Base: i.Base, Elems: args, WithArgs: i.WithArgs}
}
func (i *IsolateDef) IsoName() string {
	if len(i.Elems) > 0 {
		if a, ok := i.Elems[0].(*Atom); ok {
			return a.Relname()
		}
	}
	return ""
}
func (i *IsolateDef) Verified() []Node {
	end := len(i.Elems) - i.WithArgs
	if end > 1 {
		return i.Elems[1:end]
	}
	return nil
}
func (i *IsolateDef) Present() []Node {
	start := len(i.Elems) - i.WithArgs
	if start >= 0 && start < len(i.Elems) {
		return i.Elems[start:]
	}
	return nil
}
func (i *IsolateDef) String() string {
	parts := make([]string, len(i.Elems))
	for j, e := range i.Elems {
		parts[j] = fmt.Sprint(e)
	}
	return strings.Join(parts, ", ")
}

// TrustedIsolateDef is a trusted (unverified) isolate.
type TrustedIsolateDef struct{ IsolateDef }

func (t *TrustedIsolateDef) Clone(args []Node) Node {
	return &TrustedIsolateDef{IsolateDef: *t.IsolateDef.Clone(args).(*IsolateDef)}
}

// ExtractDef is an extraction target.
type ExtractDef struct{ IsolateDef }

func (e *ExtractDef) Clone(args []Node) Node {
	return &ExtractDef{IsolateDef: *e.IsolateDef.Clone(args).(*IsolateDef)}
}

// ProcessDef is a process definition.
type ProcessDef struct{ ExtractDef }

func (p *ProcessDef) Clone(args []Node) Node {
	return &ProcessDef{ExtractDef: *p.ExtractDef.Clone(args).(*ExtractDef)}
}

// --- Export/Import declarations ---

// ExportDecl declares exports.
type ExportDecl struct {
	DeclBase
}

func NewExportDecl(args ...Node) *ExportDecl {
	return &ExportDecl{DeclBase: DeclBase{DeclArgs: args}}
}

func (d *ExportDecl) Clone(args []Node) Node {
	return &ExportDecl{DeclBase: DeclBase{Base: d.Base, DeclArgs: args, Attributes: d.Attributes, Common: d.Common}}
}
func (d *ExportDecl) String() string { return "export" }

// ExportDef defines what is exported.
type ExportDef struct {
	Base
	Exported Node
	Scope    Node
}

func (e *ExportDef) Args() []Node { return []Node{e.Exported, e.Scope} }
func (e *ExportDef) Clone(args []Node) Node {
	return &ExportDef{Base: e.Base, Exported: args[0], Scope: args[1]}
}
func (e *ExportDef) String() string { return fmt.Sprint(e.Exported) }

// ImportDecl declares imports.
type ImportDecl struct {
	DeclBase
}

func NewImportDecl(args ...Node) *ImportDecl {
	return &ImportDecl{DeclBase: DeclBase{DeclArgs: args}}
}

func (d *ImportDecl) Clone(args []Node) Node {
	return &ImportDecl{DeclBase: DeclBase{Base: d.Base, DeclArgs: args, Attributes: d.Attributes, Common: d.Common}}
}
func (d *ImportDecl) String() string { return "import" }

// ImportDef defines what is imported.
type ImportDef struct {
	Base
	Imported Node
	Scope    Node
}

func (i *ImportDef) Args() []Node { return []Node{i.Imported, i.Scope} }
func (i *ImportDef) Clone(args []Node) Node {
	return &ImportDef{Base: i.Base, Imported: args[0], Scope: args[1]}
}
func (i *ImportDef) String() string { return fmt.Sprint(i.Imported) }

// PrivateDecl declares private members.
type PrivateDecl struct {
	DeclBase
}

func NewPrivateDecl(args ...Node) *PrivateDecl {
	return &PrivateDecl{DeclBase: DeclBase{DeclArgs: args}}
}

func (d *PrivateDecl) Clone(args []Node) Node {
	return &PrivateDecl{DeclBase: DeclBase{Base: d.Base, DeclArgs: args, Attributes: d.Attributes, Common: d.Common}}
}
func (d *PrivateDecl) String() string { return "private" }

// AliasDecl declares a type alias.
type AliasDecl struct {
	DeclBase
}

func NewAliasDecl(args ...Node) *AliasDecl {
	return &AliasDecl{DeclBase: DeclBase{DeclArgs: args}}
}

func (d *AliasDecl) Clone(args []Node) Node {
	return &AliasDecl{DeclBase: DeclBase{Base: d.Base, DeclArgs: args, Attributes: d.Attributes, Common: d.Common}}
}
func (d *AliasDecl) String() string { return "alias" }

// DelegateDecl declares delegation.
type DelegateDecl struct {
	DeclBase
}

func NewDelegateDecl(args ...Node) *DelegateDecl {
	return &DelegateDecl{DeclBase: DeclBase{DeclArgs: args}}
}

func (d *DelegateDecl) Clone(args []Node) Node {
	return &DelegateDecl{DeclBase: DeclBase{Base: d.Base, DeclArgs: args, Attributes: d.Attributes, Common: d.Common}}
}
func (d *DelegateDecl) String() string { return "delegate" }

// DelegateDef defines a delegation.
type DelegateDef struct {
	Base
	Elems []Node
}

func (d *DelegateDef) Args() []Node           { return d.Elems }
func (d *DelegateDef) Clone(args []Node) Node { return &DelegateDef{Base: d.Base, Elems: args} }
func (d *DelegateDef) String() string          { return "delegate" }

// ImplementTypeDecl declares a type implementation.
type ImplementTypeDecl struct {
	DeclBase
}

func NewImplementTypeDecl(args ...Node) *ImplementTypeDecl {
	return &ImplementTypeDecl{DeclBase: DeclBase{DeclArgs: args}}
}

func (d *ImplementTypeDecl) Clone(args []Node) Node {
	return &ImplementTypeDecl{DeclBase: DeclBase{Base: d.Base, DeclArgs: args, Attributes: d.Attributes, Common: d.Common}}
}
func (d *ImplementTypeDecl) String() string { return "implementtype" }

// --- Native code ---

// NativeCode holds embedded native code strings.
type NativeCode struct {
	Base
	Code string
}

func NewNativeCode(code string) *NativeCode { return &NativeCode{Code: code} }

func (n *NativeCode) Args() []Node           { return nil }
func (n *NativeCode) Clone(args []Node) Node { return &NativeCode{Base: n.Base, Code: n.Code} }
func (n *NativeCode) String() string          { return n.Code }

// NativeType wraps a native type expression.
type NativeType struct {
	Base
	Elems []Node
}

func (n *NativeType) Args() []Node           { return n.Elems }
func (n *NativeType) Clone(args []Node) Node { return &NativeType{Base: n.Base, Elems: args} }
func (n *NativeType) String() string          { return "<<<...>>>" }

// NativeExpr wraps a native expression.
type NativeExpr struct {
	Base
	Elems []Node
	ASort Node
}

func (n *NativeExpr) Args() []Node           { return n.Elems }
func (n *NativeExpr) Clone(args []Node) Node {
	return &NativeExpr{Base: n.Base, Elems: args, ASort: n.ASort}
}
func (n *NativeExpr) String() string { return "<<<...>>>" }

// NativeDef defines a native block.
type NativeDef struct {
	Base
	Elems []Node
}

func (n *NativeDef) Args() []Node           { return n.Elems }
func (n *NativeDef) Clone(args []Node) Node { return &NativeDef{Base: n.Base, Elems: args} }
func (n *NativeDef) String() string {
	if len(n.Elems) > 0 {
		res := ""
		if n.Elems[0] != nil {
			res = "[" + fmt.Sprint(n.Elems[0]) + "] "
		}
		for _, s := range n.Elems[1:] {
			res += " " + fmt.Sprint(s)
		}
		return res
	}
	return "native"
}

// NativeDecl declares native code.
type NativeDecl struct {
	DeclBase
}

func NewNativeDecl(args ...Node) *NativeDecl {
	return &NativeDecl{DeclBase: DeclBase{DeclArgs: args}}
}

func (d *NativeDecl) Clone(args []Node) Node {
	return &NativeDecl{DeclBase: DeclBase{Base: d.Base, DeclArgs: args, Attributes: d.Attributes, Common: d.Common}}
}
func (d *NativeDecl) String() string { return "native" }

// --- Attribute declarations ---

// AttributeDef defines an attribute.
type AttributeDef struct {
	Base
	Name  Node
	Value Node
}

func NewAttributeDef(name, value Node) *AttributeDef {
	return &AttributeDef{Name: name, Value: value}
}

func (a *AttributeDef) Args() []Node { return []Node{a.Name, a.Value} }
func (a *AttributeDef) Clone(args []Node) Node {
	return &AttributeDef{Base: a.Base, Name: args[0], Value: args[1]}
}
func (a *AttributeDef) String() string {
	return "attribute " + fmt.Sprint(a.Name) + " = " + fmt.Sprint(a.Value)
}

// AttributeDecl declares attributes.
type AttributeDecl struct {
	DeclBase
}

func NewAttributeDecl(args ...Node) *AttributeDecl {
	return &AttributeDecl{DeclBase: DeclBase{DeclArgs: args}}
}

func (d *AttributeDecl) Clone(args []Node) Node {
	return &AttributeDecl{DeclBase: DeclBase{Base: d.Base, DeclArgs: args, Attributes: d.Attributes, Common: d.Common}}
}
func (d *AttributeDecl) String() string { return "attribute" }

// --- Instantiation ---

// Instantiation represents "name : sort" instantiation.
type Instantiation struct {
	Base
	Name Node // may be nil
	Sort Node
}

func NewInstantiation(name, sort Node) *Instantiation {
	return &Instantiation{Name: name, Sort: sort}
}

func (i *Instantiation) Args() []Node { return []Node{i.Name, i.Sort} }
func (i *Instantiation) Clone(args []Node) Node {
	return &Instantiation{Base: i.Base, Name: args[0], Sort: args[1]}
}
func (i *Instantiation) String() string {
	if i.Name != nil {
		return fmt.Sprint(i.Name) + " : " + fmt.Sprint(i.Sort)
	}
	return fmt.Sprint(i.Sort)
}

// InstantiateDecl declares instantiation.
type InstantiateDecl struct {
	DeclBase
}

func NewInstantiateDecl(args ...Node) *InstantiateDecl {
	return &InstantiateDecl{DeclBase: DeclBase{DeclArgs: args}}
}

func (d *InstantiateDecl) Clone(args []Node) Node {
	return &InstantiateDecl{DeclBase: DeclBase{Base: d.Base, DeclArgs: args, Attributes: d.Attributes, Common: d.Common}}
}
func (d *InstantiateDecl) String() string { return "instantiate" }

// AutoInstanceDecl declares automatic instantiation.
type AutoInstanceDecl struct {
	DeclBase
}

func NewAutoInstanceDecl(args ...Node) *AutoInstanceDecl {
	return &AutoInstanceDecl{DeclBase: DeclBase{DeclArgs: args}}
}

func (d *AutoInstanceDecl) Clone(args []Node) Node {
	return &AutoInstanceDecl{DeclBase: DeclBase{Base: d.Base, DeclArgs: args, Attributes: d.Attributes, Common: d.Common}}
}
func (d *AutoInstanceDecl) String() string { return "autoinstance" }

// StateDef defines a state.
type StateDef struct {
	Base
	Name  string
	State Node
}

func NewStateDef(name string, state Node) *StateDef {
	return &StateDef{Name: name, State: state}
}

func (s *StateDef) Args() []Node { return []Node{NewAtom(s.Name), s.State} }
func (s *StateDef) Clone(args []Node) Node {
	return &StateDef{Base: s.Base, Name: s.Name, State: args[1]}
}
func (s *StateDef) String() string { return s.Name + " = " + fmt.Sprint(s.State) }

// Renaming holds rename mappings.
type Renaming struct {
	Base
	Elems []Node
}

func (r *Renaming) Args() []Node           { return r.Elems }
func (r *Renaming) Clone(args []Node) Node { return &Renaming{Base: r.Base, Elems: args} }
func (r *Renaming) String() string {
	parts := make([]string, len(r.Elems))
	for i, e := range r.Elems {
		parts[i] = fmt.Sprint(e)
	}
	return "<" + strings.Join(parts, ",") + ">"
}

// --- Scenario types ---

// ScenarioDecl declares a scenario.
type ScenarioDecl struct {
	DeclBase
}

func NewScenarioDecl(args ...Node) *ScenarioDecl {
	return &ScenarioDecl{DeclBase: DeclBase{DeclArgs: args}}
}

func (d *ScenarioDecl) Clone(args []Node) Node {
	return &ScenarioDecl{DeclBase: DeclBase{Base: d.Base, DeclArgs: args, Attributes: d.Attributes, Common: d.Common}}
}
func (d *ScenarioDecl) String() string { return "scenario" }

// PlaceList is a comma-separated list of places.
type PlaceList struct {
	Base
	Elems []Node
}

func (p *PlaceList) Args() []Node           { return p.Elems }
func (p *PlaceList) Clone(args []Node) Node { return &PlaceList{Base: p.Base, Elems: args} }
func (p *PlaceList) String() string {
	parts := make([]string, len(p.Elems))
	for i, e := range p.Elems {
		parts[i] = fmt.Sprint(e)
	}
	return strings.Join(parts, ",")
}

// ScenarioTransition represents a state transition.
type ScenarioTransition struct {
	Base
	From   Node
	To     Node
	Action Node
}

func (s *ScenarioTransition) Args() []Node { return []Node{s.From, s.To, s.Action} }
func (s *ScenarioTransition) Clone(args []Node) Node {
	return &ScenarioTransition{Base: s.Base, From: args[0], To: args[1], Action: args[2]}
}
func (s *ScenarioTransition) String() string {
	return fmt.Sprint(s.From) + "->" + fmt.Sprint(s.To) + " : " + fmt.Sprint(s.Action)
}

// ScenarioDef defines a scenario.
type ScenarioDef struct {
	Base
	Elems []Node
}

func (s *ScenarioDef) Args() []Node           { return s.Elems }
func (s *ScenarioDef) Clone(args []Node) Node { return &ScenarioDef{Base: s.Base, Elems: args} }
func (s *ScenarioDef) String() string          { return "scenario{...}" }

// IsolateObjectDecl is an isolate for an object (no defines).
type IsolateObjectDecl struct {
	IsolateDecl
}

func (d *IsolateObjectDecl) Clone(args []Node) Node {
	return &IsolateObjectDecl{IsolateDecl: *d.IsolateDecl.Clone(args).(*IsolateDecl)}
}
