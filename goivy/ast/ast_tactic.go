package ast

import (
	"fmt"
	"iter"
	"strings"

	iu "github.com/glycerine/ivy/goivy/ivyutils"
	"github.com/glycerine/ivy/goivy/xtracer"
)

// --- Tactic types ---
// Tactics are used in proof scripts.

// Tactic is the base for all tactic nodes.
type AstTactic struct {
	Base
	Elems []Node
}

func (t *AstTactic) Args() []Node           { return t.Elems }
func (t *AstTactic) Clone(args []Node) Node { return &AstTactic{Base: t.Base, Elems: args} }
func (t *AstTactic) String() string         { return "tactic" }
func (t *AstTactic) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(tactic%v elems:%v)", t.Base.canonFields(), SliceCanon(t.Elems)))
}

// SchemaInstantiation applies a schema.
type SchemaInstantiation struct {
	Base
	SchemaName Node
	Ren        Node   // renaming
	Matches    []Node // match args
}

func (s *SchemaInstantiation) Args() []Node {
	args := []Node{s.SchemaName, s.Ren}
	args = append(args, s.Matches...)
	return args
}
func (s *SchemaInstantiation) Clone(args []Node) Node {
	return &SchemaInstantiation{Base: s.Base, SchemaName: args[0], Ren: args[1], Matches: args[2:]}
}
func (s *SchemaInstantiation) String() string {
	res := "apply " + fmt.Sprint(s.SchemaName)
	if s.Ren != nil {
		res += " " + fmt.Sprint(s.Ren)
	}
	if len(s.Matches) > 0 {
		parts := make([]string, len(s.Matches))
		for i, m := range s.Matches {
			parts[i] = fmt.Sprint(m)
		}
		res += " with " + strings.Join(parts, ",")
	}
	return res
}
func (s *SchemaInstantiation) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(schemaInstantiation%v schemaName:%v ren:%v matches:%v)", s.Base.canonFields(), nodeCanon(s.SchemaName), nodeCanon(s.Ren), SliceCanon(s.Matches)))
}

// AssumeTactic assumes a schema.
type AssumeTactic struct {
	Base
	TLabel     Node // label for the assumption
	SchemaName Node
	Ren        Node
	Matches    []Node
}

func (a *AssumeTactic) Args() []Node {
	args := []Node{a.SchemaName, a.Ren}
	args = append(args, a.Matches...)
	return args
}
func (a *AssumeTactic) Clone(args []Node) Node {
	c := &AssumeTactic{Base: a.Base, TLabel: a.TLabel, SchemaName: args[0], Ren: args[1], Matches: args[2:]}
	return c
}
func (a *AssumeTactic) String() string { return "assume " + fmt.Sprint(a.SchemaName) }
func (a *AssumeTactic) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(assumeTactic%v tLabel:%v schemaName:%v ren:%v matches:%v)", a.Base.canonFields(), nodeCanon(a.TLabel), nodeCanon(a.SchemaName), nodeCanon(a.Ren), SliceCanon(a.Matches)))
}

// AssumeGlobalTactic is a global assume tactic.
type AssumeGlobalTactic struct {
	AssumeTactic
}

func (a *AssumeGlobalTactic) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(assumeGlobalTactic assumeTactic:%v)", a.AssumeTactic.Canon()))
}

// UnfoldSpec specifies what to unfold.
type UnfoldSpec struct {
	Base
	DefName   Node
	Renamings []Node
}

func (cfg *AstConfig) NewUnfoldSpec(defName Node, renamings []Node) *UnfoldSpec {
	u := &UnfoldSpec{DefName: defName, Renamings: renamings}
	u.Cfg = cfg
	return u
}

func (u *UnfoldSpec) Args() []Node {
	return append([]Node{u.DefName}, u.Renamings...)
}
func (u *UnfoldSpec) Clone(args []Node) Node {
	return &UnfoldSpec{Base: u.Base, DefName: args[0], Renamings: args[1:]}
}
func (u *UnfoldSpec) String() string { return fmt.Sprint(u.DefName) }
func (u *UnfoldSpec) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(unfoldSpec%v defName:%v renamings:%v)", u.Base.canonFields(), nodeCanon(u.DefName), SliceCanon(u.Renamings)))
}

// UnfoldTactic unfolds definitions.
type UnfoldTactic struct {
	Base
	TLabel   Node
	Premise  Node   // first arg, may be NoneAST
	UnfSpecs []Node // unfold specifications
}

func (u *UnfoldTactic) Args() []Node {
	return append([]Node{u.Premise}, u.UnfSpecs...)
}
func (u *UnfoldTactic) Clone(args []Node) Node {
	c := &UnfoldTactic{Base: u.Base, TLabel: u.TLabel, Premise: args[0], UnfSpecs: args[1:]}
	return c
}
func (u *UnfoldTactic) String() string { return "unfold" }
func (u *UnfoldTactic) HasPremise() bool {
	_, isNone := u.Premise.(*NoneAST)
	return !isNone
}
func (u *UnfoldTactic) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(unfoldTactic%v tLabel:%v premise:%v unfSpecs:%v)", u.Base.canonFields(), nodeCanon(u.TLabel), nodeCanon(u.Premise), SliceCanon(u.UnfSpecs)))
}

// ForgetTactic forgets named facts.
type ForgetTactic struct {
	Base
	Names []Node
}

func (f *ForgetTactic) Args() []Node           { return f.Names }
func (f *ForgetTactic) Clone(args []Node) Node { return &ForgetTactic{Base: f.Base, Names: args} }
func (f *ForgetTactic) String() string         { return "forget" }
func (f *ForgetTactic) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(forgetTactic%v names:%v)", f.Base.canonFields(), SliceCanon(f.Names)))
}

// ShowGoalsTactic displays current goals.
type ShowGoalsTactic struct {
	Base
}

func (s *ShowGoalsTactic) Args() []Node           { return nil }
func (s *ShowGoalsTactic) Clone(args []Node) Node { return &ShowGoalsTactic{Base: s.Base} }
func (s *ShowGoalsTactic) String() string         { return "showgoals" }
func (s *ShowGoalsTactic) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(showGoalsTactic%v)", s.Base.canonFields()))
}

// DeferGoalTactic defers a goal.
type DeferGoalTactic struct {
	Base
	Elems []Node
}

func (d *DeferGoalTactic) Args() []Node           { return d.Elems }
func (d *DeferGoalTactic) Clone(args []Node) Node { return &DeferGoalTactic{Base: d.Base, Elems: args} }
func (d *DeferGoalTactic) String() string         { return "defergoal" }
func (d *DeferGoalTactic) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(deferGoalTactic%v elems:%v)", d.Base.canonFields(), SliceCanon(d.Elems)))
}

// NullTactic is an empty tactic.
type NullTactic struct {
	Base
}

func (n *NullTactic) Args() []Node           { return nil }
func (n *NullTactic) Clone(args []Node) Node { return &NullTactic{Base: n.Base} }
func (n *NullTactic) String() string         { return "{}" }
func (n *NullTactic) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(nullTactic%v)", n.Base.canonFields()))
}

// LetTactic introduces local definitions in a proof.
type LetTactic struct {
	Base
	Defs []Node
}

func (l *LetTactic) Args() []Node           { return l.Defs }
func (l *LetTactic) Clone(args []Node) Node { return &LetTactic{Base: l.Base, Defs: args} }
func (l *LetTactic) String() string {
	parts := make([]string, len(l.Defs))
	for i, d := range l.Defs {
		parts[i] = fmt.Sprint(d)
	}
	return "let " + strings.Join(parts, ",")
}
func (l *LetTactic) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(letTactic%v defs:%v)", l.Base.canonFields(), SliceCanon(l.Defs)))
}

// WitnessTactic provides witnesses.
type WitnessTactic struct {
	Base
	Witnesses []Node
}

func (w *WitnessTactic) Args() []Node           { return w.Witnesses }
func (w *WitnessTactic) Clone(args []Node) Node { return &WitnessTactic{Base: w.Base, Witnesses: args} }
func (w *WitnessTactic) String() string {
	parts := make([]string, len(w.Witnesses))
	for i, w := range w.Witnesses {
		parts[i] = fmt.Sprint(w)
	}
	return "witness " + strings.Join(parts, ",")
}
func (w *WitnessTactic) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(witnessTactic%v witnesses:%v)", w.Base.canonFields(), SliceCanon(w.Witnesses)))
}

// SpoilTactic spoils a proof state.
type SpoilTactic struct {
	Base
	Target Node
}

func (s *SpoilTactic) Args() []Node           { return []Node{s.Target} }
func (s *SpoilTactic) Clone(args []Node) Node { return &SpoilTactic{Base: s.Base, Target: args[0]} }
func (s *SpoilTactic) String() string         { return "spoil " + fmt.Sprint(s.Target) }
func (s *SpoilTactic) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(spoilTactic%v target:%v)", s.Base.canonFields(), nodeCanon(s.Target)))
}

// IfTactic is a conditional tactic.
type IfTactic struct {
	Base
	Cond Node
	Then Node
	Else Node
}

func (i *IfTactic) Args() []Node { return []Node{i.Cond, i.Then, i.Else} }
func (i *IfTactic) Clone(args []Node) Node {
	return &IfTactic{Base: i.Base, Cond: args[0], Then: args[1], Else: args[2]}
}
func (i *IfTactic) String() string {
	return "if " + fmt.Sprint(i.Cond) + " { " + fmt.Sprint(i.Then) + " } else { " + fmt.Sprint(i.Else) + " }"
}
func (i *IfTactic) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(ifTactic%v cond:%v then:%v else:%v)", i.Base.canonFields(), nodeCanon(i.Cond), nodeCanon(i.Then), nodeCanon(i.Else)))
}

// PropertyTactic introduces a property in a proof.
type PropertyTactic struct {
	Base
	Prop  Node
	PName Node
	Proof Node
}

func (p *PropertyTactic) Args() []Node { return []Node{p.Prop, p.PName, p.Proof} }
func (p *PropertyTactic) Clone(args []Node) Node {
	return &PropertyTactic{Base: p.Base, Prop: args[0], PName: args[1], Proof: args[2]}
}
func (p *PropertyTactic) String() string { return "property " + fmt.Sprint(p.Prop) }
func (p *PropertyTactic) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(propertyTactic%v prop:%v pName:%v proof:%v)", p.Base.canonFields(), nodeCanon(p.Prop), nodeCanon(p.PName), nodeCanon(p.Proof)))
}

// FunctionTactic introduces a function in a proof.
type FunctionTactic struct {
	Base
	Elems []Node
}

func (f *FunctionTactic) Args() []Node           { return f.Elems }
func (f *FunctionTactic) Clone(args []Node) Node { return &FunctionTactic{Base: f.Base, Elems: args} }
func (f *FunctionTactic) String() string {
	parts := make([]string, len(f.Elems))
	for i, e := range f.Elems {
		parts[i] = fmt.Sprint(e)
	}
	return "function " + strings.Join(parts, ",")
}
func (f *FunctionTactic) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(functionTactic%v elems:%v)", f.Base.canonFields(), SliceCanon(f.Elems)))
}

// TacticWith holds the "with" clause elements for a tactic invocation.
// Corresponds to Python's TacticWith class (ivy_ast.py line 928).
type TacticWith struct {
	Base
	Elems []Node
}

func (tw *TacticWith) Args() []Node           { return tw.Elems }
func (tw *TacticWith) Clone(args []Node) Node { return &TacticWith{Base: tw.Base, Elems: args} }
func (tw *TacticWith) String() string {
	if len(tw.Elems) == 0 {
		return ""
	}
	parts := make([]string, len(tw.Elems))
	for i, e := range tw.Elems {
		parts[i] = fmt.Sprint(e)
	}
	return " with " + strings.Join(parts, " ")
}
func (tw *TacticWith) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(tacticWith%v elems:%v)", tw.Base.canonFields(), SliceCanon(tw.Elems)))
}

// TacticLets holds the let-bindings for a tactic invocation.
// Corresponds to Python's TacticLets class (ivy_ast.py line 932).
type TacticLets struct {
	Base
	Lets []Node
}

func (tl *TacticLets) Args() []Node           { return tl.Lets }
func (tl *TacticLets) Clone(args []Node) Node { return &TacticLets{Base: tl.Base, Lets: args} }
func (tl *TacticLets) String() string {
	if len(tl.Lets) == 0 {
		return ""
	}
	parts := make([]string, len(tl.Lets))
	for i, l := range tl.Lets {
		parts[i] = fmt.Sprint(l)
	}
	return " with " + strings.Join(parts, " ")
}
func (tl *TacticLets) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(tacticLets%v lets:%v)", tl.Base.canonFields(), SliceCanon(tl.Lets)))
}

// TacticTactic invokes a named tactic.
type TacticTactic struct {
	Base
	TName  Node
	Body   Node // TacticWith or TacticLets
	Proof  Node // optional
	Labels []string
}

func (t *TacticTactic) Args() []Node {
	args := []Node{t.TName, t.Body}
	if t.Proof != nil {
		args = append(args, t.Proof)
	}
	return args
}
func (t *TacticTactic) Clone(args []Node) Node {
	c := &TacticTactic{Base: t.Base, TName: args[0], Body: args[1]}
	if len(args) > 2 {
		c.Proof = args[2]
	}
	return c
}
func (t *TacticTactic) String() string { return "tactic " + fmt.Sprint(t.TName) }
func (t *TacticTactic) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(tacticTactic%v tName:%v body:%v proof:%v labels:%v)", t.Base.canonFields(), nodeCanon(t.TName), nodeCanon(t.Body), nodeCanon(t.Proof), stringSliceCanon(t.Labels)))
}

// TacticDeclsList returns the tactic declarations (from TacticWith body).
// Corresponds to Python's TacticTactic.tactic_decls property.
func (t *TacticTactic) TacticDeclsList() []Node {
	if tw, ok := t.Body.(*TacticWith); ok {
		return tw.Elems
	}
	return nil
}

// TacticLetsList returns the tactic let-bindings (from TacticLets body).
// Corresponds to Python's TacticTactic.tactic_lets property.
func (t *TacticTactic) TacticLetsList() []Node {
	if tl, ok := t.Body.(*TacticLets); ok {
		return tl.Lets
	}
	return nil
}

// TacticProofNode returns the tactic's optional proof, or nil.
// Corresponds to Python's TacticTactic.tactic_proof property.
func (t *TacticTactic) TacticProofNode() Node {
	if t.Proof != nil {
		if _, isNone := t.Proof.(*NoneAST); !isNone {
			return t.Proof
		}
	}
	return nil
}

// ProofTactic wraps a labeled proof.
type AstProofTactic struct {
	Base
	TLabel Node
	Proof  Node
}

func (p *AstProofTactic) Args() []Node { return []Node{p.TLabel, p.Proof} }
func (p *AstProofTactic) Clone(args []Node) Node {
	return &AstProofTactic{Base: p.Base, TLabel: args[0], Proof: args[1]}
}
func (p *AstProofTactic) String() string {
	return "proof [" + fmt.Sprint(p.TLabel) + "] {" + fmt.Sprint(p.Proof) + "}"
}
func (p *AstProofTactic) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(proofTactic%v tLabel:%v proof:%v)", p.Base.canonFields(), nodeCanon(p.TLabel), nodeCanon(p.Proof)))
}

// ComposeTactics is a sequence of tactics.
type ComposeTactics struct {
	Base
	Tactics []Node
}

func (c *ComposeTactics) Args() []Node           { return c.Tactics }
func (c *ComposeTactics) Clone(args []Node) Node { return &ComposeTactics{Base: c.Base, Tactics: args} }
func (c *ComposeTactics) String() string {
	parts := make([]string, len(c.Tactics))
	for i, t := range c.Tactics {
		parts[i] = fmt.Sprint(t)
	}
	return strings.Join(parts, "; ")
}
func (c *ComposeTactics) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(composeTactics%v tactics:%v)", c.Base.canonFields(), SliceCanon(c.Tactics)))
}

// --- Vocab: extract symbol names from proof tactic trees ---
//
// Port of Python Tactic.vocab(self, names) method hierarchy (ivy_ast.py:760-950).
// Python's vocab methods call
// names.update(symbols_ivy_ast(m.args[1])) where
// symbols_ivy_ast is the AST-level generator at ivy_ast.py:1879.
// (we renamed from symbols_ast to symbols_ivy_ast to avoid
// confusion with the ~/pyivy/ivy/ivy/ivy_logic_utils.py:537
// function of the same name.

// VocabNames is a sorted set of symbol name strings.
// Uses Omap (red-black tree) for deterministic sorted iteration.
// Python's all_names is set() with mixed str/Symbol, but App.rep is
// often a plain str in Python (due to prefix/rename/drop_prefix ops),
// so everything deduplicates as strings. Go matches by keying on string.
type VocabNames = iu.Omap[string, bool]

// NewVocabNames creates an empty VocabNames set.
func NewVocabNames() *VocabNames {
	return iu.NewOmap[string, bool]()
}

// VocabNamesUpdate consumes an iter.Seq[any] and adds to the set.
// Mirrors Python: names.update(symbols_ast(...))
// Extracts the string name from any yielded value (string, *Symbol, *This).
func VocabNamesUpdate(vn *VocabNames, seq iter.Seq[any]) {
	for val := range seq {
		switch v := val.(type) {
		case string:
			vn.Set(v, true)
			if xtracer.Enabled {
				//xtracer.Trace("vocab.add src=Atom name=%s val_type=str\n string", v)
				xtracer.Trace("vocab.add src=Atom name=%s val_type=str", v)
			}
		case *Symbol:
			vn.Set(v.Rep, true)
			if xtracer.Enabled {
				//xtracer.Trace("vocab.add src=App name=%s val_type=str\n Symbol='%#v'", v.Rep, v)
				xtracer.Trace("vocab.add src=App name=%s val_type=str", v.Rep)
			}
		case *This:
			vn.Set("this", true)
			if xtracer.Enabled {
				//xtracer.Trace("vocab.add src=App name=this val_type=str\n *This='%#v'", v)
				xtracer.Trace("vocab.add src=App name=this val_type=str")
			}
		case *AstNamedBinder:
			vn.Set(fmt.Sprintf("ptr=%p; v=%#v", v, v), true)
			//vn.Set(string(v.Canon()), true)
			//vn.Set(v.Name, true)
			// Python adds the NamedBinder object to the set, but it never matches
			// any string lookup (it's inert). Skipping xtrace on both sides.
			//if xtracer.Enabled {
			//	xtracer.Trace("vocab.add src=App name=%v val_type=NamedBinder", v.Name)
			//}
		default:
			panicf("unhandled type=%T/val=%v", val, val)
		}
	}
}

// Vocaber is implemented by AST nodes that extract symbol names from proof trees.
// Port of Python Tactic.vocab(self, names) method hierarchy (ivy_ast.py).
type Vocaber interface {
	Vocab(names *VocabNames)
}

// VocabNode calls Vocab on the node if it implements Vocaber, otherwise no-op.
// Matches Python's base Tactic.vocab which is pass.
func VocabNode(node Node, names *VocabNames) {
	if v, ok := node.(Vocaber); ok {
		v.Vocab(names)
	}
}

// IterSymbolsASTNode yields values from an AST node tree.
// Port of Python ivy_ast.symbols_ast (ivy_ast.py:1879) as iter.Seq[any].
//
// Yields string for *Atom (matching Python str from Atom.rep)
// and Node for *App (matching Python Symbol object from App.rep).
// Both Atom and App (and all other nodes) recurse on Args() children.
func IterSymbolsASTNode(node Node) iter.Seq[any] {
	return func(yield func(any) bool) {
		iterSymbolsASTNodeRec(node, yield)
	}
}

func iterSymbolsASTNodeRec(node Node, yield func(any) bool) bool {
	if node == nil {
		return true
	}
	switch v := node.(type) {
	case *Atom:
		if v.Rep != "" {
			if !yield(v.Rep) { // yields string — matches Python str
				return false
			}
		}
	case *App:
		if v.Rep != nil {
			if !yield(v.Rep) { // yields Node (usually *Symbol) — matches Python Symbol
				return false
			}
		}
	}
	for _, child := range node.Args() {
		if !iterSymbolsASTNodeRec(child, yield) {
			return false
		}
	}
	return true
}

// --- Vocab methods on Tactic types ---

// SchemaInstantiation.Vocab — Python TacticWithMatch.vocab (ivy_ast.py:780)
// Python: for m in self.match(): names.update(symbols_ast(m.args[1]))
func (s *SchemaInstantiation) Vocab(names *VocabNames) {
	for _, m := range s.Matches {
		args := m.Args()
		if len(args) >= 2 {
			VocabNamesUpdate(names, IterSymbolsASTNode(args[1]))
		}
	}
}

// AssumeTactic.Vocab — inherits TacticWithMatch.vocab pattern (ivy_ast.py:780)
func (a *AssumeTactic) Vocab(names *VocabNames) {
	for _, m := range a.Matches {
		args := m.Args()
		if len(args) >= 2 {
			VocabNamesUpdate(names, IterSymbolsASTNode(args[1]))
		}
	}
}

// LetTactic.Vocab — Python LetTactic.vocab (ivy_ast.py:852)
// Python: for m in self.args: names.update(symbols_ast(m.args[1]))
func (l *LetTactic) Vocab(names *VocabNames) {
	for _, d := range l.Defs {
		args := d.Args()
		if len(args) >= 2 {
			VocabNamesUpdate(names, IterSymbolsASTNode(args[1]))
		}
	}
}

// WitnessTactic.Vocab — Python WitnessTactic.vocab (ivy_ast.py:861)
// Python: for m in self.args: names.update(symbols_ast(m.args[1]))
func (w *WitnessTactic) Vocab(names *VocabNames) {
	for _, m := range w.Witnesses {
		args := m.Args()
		if len(args) >= 2 {
			VocabNamesUpdate(names, IterSymbolsASTNode(args[1]))
		}
	}
}

// IfTactic.Vocab — Python IfTactic.vocab (ivy_ast.py:876)
// Python: names.update(symbols_ast(self.args[0]))
//
//	for arg in self.args[1:]: arg.vocab(names)
func (i *IfTactic) Vocab(names *VocabNames) {
	VocabNamesUpdate(names, IterSymbolsASTNode(i.Cond))
	VocabNode(i.Then, names)
	VocabNode(i.Else, names)
}

// PropertyTactic.Vocab — Python PropertyTactic.vocab (ivy_ast.py:895)
// Python: if not isinstance(self.args[2], NoneAST): self.args[2].vocab(names)
func (p *PropertyTactic) Vocab(names *VocabNames) {
	if p.Proof != nil {
		if _, isNone := p.Proof.(*NoneAST); !isNone {
			VocabNode(p.Proof, names)
		}
	}
}

// TacticTactic.Vocab — Python TacticTactic.vocab (ivy_ast.py:920)
// Python: names.update(symbols_ast(self.args[1]))
//
//	if not isinstance(self.args[2], NoneAST): self.args[2].vocab(names)
func (t *TacticTactic) Vocab(names *VocabNames) {
	VocabNamesUpdate(names, IterSymbolsASTNode(t.Body))
	if t.Proof != nil {
		if _, isNone := t.Proof.(*NoneAST); !isNone {
			VocabNode(t.Proof, names)
		}
	}
}

// ProofTactic.Vocab — Python ProofTactic.vocab (ivy_ast.py:934)
// Python: self.args[1].vocab(names)
func (p *AstProofTactic) Vocab(names *VocabNames) {
	VocabNode(p.Proof, names)
}

// ComposeTactics.Vocab — Python ComposeTactics.vocab (ivy_ast.py:948)
// Python: for arg in self.args: arg.vocab(names)
func (c *ComposeTactics) Vocab(names *VocabNames) {
	for _, t := range c.Tactics {
		VocabNode(t, names)
	}
}

// --- Constructors (methods on *AstConfig) ---

func (cfg *AstConfig) NewSchemaInstantiation(schemaName, ren Node) *SchemaInstantiation {
	s := &SchemaInstantiation{SchemaName: schemaName, Ren: ren}
	s.Cfg = cfg
	return s
}

func (cfg *AstConfig) NewSchemaInstantiationWithMatches(schemaName, ren Node, matches []Node) *SchemaInstantiation {
	s := &SchemaInstantiation{SchemaName: schemaName, Ren: ren, Matches: matches}
	s.Cfg = cfg
	return s
}

func (cfg *AstConfig) NewAssumeTactic(schemaName, ren Node) *AssumeTactic {
	a := &AssumeTactic{SchemaName: schemaName, Ren: ren}
	a.Cfg = cfg
	return a
}

func (cfg *AstConfig) NewAssumeGlobalTactic(schemaName, ren Node) *AssumeGlobalTactic {
	at := &AssumeGlobalTactic{AssumeTactic: AssumeTactic{SchemaName: schemaName, Ren: ren}}
	at.Cfg = cfg
	return at
}

func (cfg *AstConfig) NewAssumeGlobalTacticWithMatches(schemaName, ren Node, matches []Node) *AssumeGlobalTactic {
	at := &AssumeGlobalTactic{AssumeTactic: AssumeTactic{SchemaName: schemaName, Ren: ren, Matches: matches}}
	at.Cfg = cfg
	return at
}

func (cfg *AstConfig) NewAssumeTacticWithMatches(schemaName, ren Node, matches []Node) *AssumeTactic {
	a := &AssumeTactic{SchemaName: schemaName, Ren: ren, Matches: matches}
	a.Cfg = cfg
	return a
}

func (cfg *AstConfig) NewRenaming(elems []Node) *AstRenaming {
	r := &AstRenaming{Elems: elems}
	r.Cfg = cfg
	return r
}

func (cfg *AstConfig) NewUnfoldTactic(premise Node, unfSpecs []Node) *UnfoldTactic {
	u := &UnfoldTactic{Premise: premise, UnfSpecs: unfSpecs}
	u.Cfg = cfg
	return u
}

func (cfg *AstConfig) NewForgetTactic(names []Node) *ForgetTactic {
	f := &ForgetTactic{Names: names}
	f.Cfg = cfg
	return f
}

func (cfg *AstConfig) NewShowGoalsTactic() *ShowGoalsTactic {
	s := &ShowGoalsTactic{}
	s.Cfg = cfg
	return s
}

func (cfg *AstConfig) NewDeferGoalTactic() *DeferGoalTactic {
	d := &DeferGoalTactic{}
	d.Cfg = cfg
	return d
}

func (cfg *AstConfig) NewNullTactic() *NullTactic {
	n := &NullTactic{}
	n.Cfg = cfg
	return n
}

func (cfg *AstConfig) NewLetTactic(defs []Node) *LetTactic {
	l := &LetTactic{Defs: defs}
	l.Cfg = cfg
	return l
}

func (cfg *AstConfig) NewWitnessTactic(witnesses []Node) *WitnessTactic {
	w := &WitnessTactic{Witnesses: witnesses}
	w.Cfg = cfg
	return w
}

func (cfg *AstConfig) NewSpoilTactic(target Node) *SpoilTactic {
	s := &SpoilTactic{Target: target}
	s.Cfg = cfg
	return s
}

func (cfg *AstConfig) NewIfTactic(cond, then, els Node) *IfTactic {
	i := &IfTactic{Cond: cond, Then: then, Else: els}
	i.Cfg = cfg
	return i
}

func (cfg *AstConfig) NewPropertyTactic(prop, pName, proof Node) *PropertyTactic {
	p := &PropertyTactic{Prop: prop, PName: pName, Proof: proof}
	p.Cfg = cfg
	return p
}

func (cfg *AstConfig) NewFunctionTactic(elems []Node) *FunctionTactic {
	f := &FunctionTactic{Elems: elems}
	f.Cfg = cfg
	return f
}

func (cfg *AstConfig) NewTacticWith(elems []Node) *TacticWith {
	tw := &TacticWith{Elems: elems}
	tw.Cfg = cfg
	return tw
}

func (cfg *AstConfig) NewTacticLets(lets []Node) *TacticLets {
	tl := &TacticLets{Lets: lets}
	tl.Cfg = cfg
	return tl
}

func (cfg *AstConfig) NewTacticTactic(tName, body, proof Node) *TacticTactic {
	t := &TacticTactic{TName: tName, Body: body, Proof: proof}
	t.Cfg = cfg
	return t
}

func (cfg *AstConfig) NewProofTactic(tLabel, proof Node) *AstProofTactic {
	p := &AstProofTactic{TLabel: tLabel, Proof: proof}
	p.Cfg = cfg
	return p
}

func (cfg *AstConfig) NewComposeTactics(tactics []Node) *ComposeTactics {
	c := &ComposeTactics{Tactics: tactics}
	c.Cfg = cfg
	return c
}
