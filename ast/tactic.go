package ast

import (
	"fmt"
	"strings"

	iu "github.com/glycerine/goivy/ivyutils"
)

// --- Tactic types ---
// Tactics are used in proof scripts.

// Tactic is the base for all tactic nodes.
type Tactic struct {
	Base
	Elems []Node
}

func (t *Tactic) Args() []Node           { return t.Elems }
func (t *Tactic) Clone(args []Node) Node { return &Tactic{Base: t.Base, Elems: args} }
func (t *Tactic) String() string          { return "tactic" }
func (t *Tactic) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(tactic %v elems:%v)", t.Base.canonFields(), SliceCanon(t.Elems)))
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
	return iu.Canonical(fmt.Sprintf("(schemaInstantiation %v schemaName:%v ren:%v matches:%v)", s.Base.canonFields(), nodeCanon(s.SchemaName), nodeCanon(s.Ren), SliceCanon(s.Matches)))
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
	return iu.Canonical(fmt.Sprintf("(assumeTactic %v tLabel:%v schemaName:%v ren:%v matches:%v)", a.Base.canonFields(), nodeCanon(a.TLabel), nodeCanon(a.SchemaName), nodeCanon(a.Ren), SliceCanon(a.Matches)))
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

func (u *UnfoldSpec) Args() []Node {
	return append([]Node{u.DefName}, u.Renamings...)
}
func (u *UnfoldSpec) Clone(args []Node) Node {
	return &UnfoldSpec{Base: u.Base, DefName: args[0], Renamings: args[1:]}
}
func (u *UnfoldSpec) String() string { return fmt.Sprint(u.DefName) }
func (u *UnfoldSpec) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(unfoldSpec %v defName:%v renamings:%v)", u.Base.canonFields(), nodeCanon(u.DefName), SliceCanon(u.Renamings)))
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
	return iu.Canonical(fmt.Sprintf("(unfoldTactic %v tLabel:%v premise:%v unfSpecs:%v)", u.Base.canonFields(), nodeCanon(u.TLabel), nodeCanon(u.Premise), SliceCanon(u.UnfSpecs)))
}

// ForgetTactic forgets named facts.
type ForgetTactic struct {
	Base
	Names []Node
}

func (f *ForgetTactic) Args() []Node           { return f.Names }
func (f *ForgetTactic) Clone(args []Node) Node { return &ForgetTactic{Base: f.Base, Names: args} }
func (f *ForgetTactic) String() string          { return "forget" }
func (f *ForgetTactic) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(forgetTactic %v names:%v)", f.Base.canonFields(), SliceCanon(f.Names)))
}

// ShowGoalsTactic displays current goals.
type ShowGoalsTactic struct {
	Base
}

func (s *ShowGoalsTactic) Args() []Node           { return nil }
func (s *ShowGoalsTactic) Clone(args []Node) Node { return &ShowGoalsTactic{Base: s.Base} }
func (s *ShowGoalsTactic) String() string          { return "showgoals" }
func (s *ShowGoalsTactic) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(showGoalsTactic %v)", s.Base.canonFields()))
}

// DeferGoalTactic defers a goal.
type DeferGoalTactic struct {
	Base
	Elems []Node
}

func (d *DeferGoalTactic) Args() []Node           { return d.Elems }
func (d *DeferGoalTactic) Clone(args []Node) Node { return &DeferGoalTactic{Base: d.Base, Elems: args} }
func (d *DeferGoalTactic) String() string          { return "defergoal" }
func (d *DeferGoalTactic) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(deferGoalTactic %v elems:%v)", d.Base.canonFields(), SliceCanon(d.Elems)))
}

// NullTactic is an empty tactic.
type NullTactic struct {
	Base
}

func (n *NullTactic) Args() []Node           { return nil }
func (n *NullTactic) Clone(args []Node) Node { return &NullTactic{Base: n.Base} }
func (n *NullTactic) String() string          { return "{}" }
func (n *NullTactic) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(nullTactic %v)", n.Base.canonFields()))
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
	return iu.Canonical(fmt.Sprintf("(letTactic %v defs:%v)", l.Base.canonFields(), SliceCanon(l.Defs)))
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
	return iu.Canonical(fmt.Sprintf("(witnessTactic %v witnesses:%v)", w.Base.canonFields(), SliceCanon(w.Witnesses)))
}

// SpoilTactic spoils a proof state.
type SpoilTactic struct {
	Base
	Target Node
}

func (s *SpoilTactic) Args() []Node           { return []Node{s.Target} }
func (s *SpoilTactic) Clone(args []Node) Node { return &SpoilTactic{Base: s.Base, Target: args[0]} }
func (s *SpoilTactic) String() string          { return "spoil " + fmt.Sprint(s.Target) }
func (s *SpoilTactic) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(spoilTactic %v target:%v)", s.Base.canonFields(), nodeCanon(s.Target)))
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
	return iu.Canonical(fmt.Sprintf("(ifTactic %v cond:%v then:%v else:%v)", i.Base.canonFields(), nodeCanon(i.Cond), nodeCanon(i.Then), nodeCanon(i.Else)))
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
	return iu.Canonical(fmt.Sprintf("(propertyTactic %v prop:%v pName:%v proof:%v)", p.Base.canonFields(), nodeCanon(p.Prop), nodeCanon(p.PName), nodeCanon(p.Proof)))
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
	return iu.Canonical(fmt.Sprintf("(functionTactic %v elems:%v)", f.Base.canonFields(), SliceCanon(f.Elems)))
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
	return iu.Canonical(fmt.Sprintf("(tacticWith %v elems:%v)", tw.Base.canonFields(), SliceCanon(tw.Elems)))
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
	return iu.Canonical(fmt.Sprintf("(tacticLets %v lets:%v)", tl.Base.canonFields(), SliceCanon(tl.Lets)))
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
	return iu.Canonical(fmt.Sprintf("(tacticTactic %v tName:%v body:%v proof:%v labels:%v)", t.Base.canonFields(), nodeCanon(t.TName), nodeCanon(t.Body), nodeCanon(t.Proof), stringSliceCanon(t.Labels)))
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
type ProofTactic struct {
	Base
	TLabel Node
	Proof  Node
}

func (p *ProofTactic) Args() []Node { return []Node{p.TLabel, p.Proof} }
func (p *ProofTactic) Clone(args []Node) Node {
	return &ProofTactic{Base: p.Base, TLabel: args[0], Proof: args[1]}
}
func (p *ProofTactic) String() string {
	return "proof [" + fmt.Sprint(p.TLabel) + "] {" + fmt.Sprint(p.Proof) + "}"
}
func (p *ProofTactic) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(proofTactic %v tLabel:%v proof:%v)", p.Base.canonFields(), nodeCanon(p.TLabel), nodeCanon(p.Proof)))
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
	return iu.Canonical(fmt.Sprintf("(composeTactics %v tactics:%v)", c.Base.canonFields(), SliceCanon(c.Tactics)))
}
