package ast

import (
	"fmt"
	"strings"

	iu "github.com/glycerine/goivy/ivyutils"
)

// --- Formula node types ---
// These represent the parsed syntax tree formulas, distinct from logic/ IR.

// And is a conjunction of formulas. Empty And = true.
type And struct {
	Base
	Terms []Node
}

func NewAnd(terms ...Node) *And {
	a := &And{Terms: terms}
	a.Cfg = DefaultAstConfig
	return a
}

func (cfg *AstConfig) NewAnd(terms ...Node) *And {
	a := &And{Terms: terms}
	a.Cfg = cfg
	return a
}

func (a *And) Args() []Node           { return a.Terms }
func (a *And) Clone(args []Node) Node { return &And{Base: a.Base, Terms: args} }
func (a *And) String() string {
	if len(a.Terms) == 0 {
		return "true"
	}
	return NaryRepr("&", a.Terms)
}
func (a *And) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(and %v terms:%v)", a.Base.canonFields(), SliceCanon(a.Terms)))
}

// Or is a disjunction of formulas. Empty Or = false.
type Or struct {
	Base
	Terms []Node
}

func NewOr(terms ...Node) *Or {
	a := &Or{Terms: terms}
	a.Cfg = DefaultAstConfig
	return a
}

func (cfg *AstConfig) NewOr(terms ...Node) *Or {
	a := &Or{Terms: terms}
	a.Cfg = cfg
	return a
}

func (o *Or) Args() []Node           { return o.Terms }
func (o *Or) Clone(args []Node) Node { return &Or{Base: o.Base, Terms: args} }
func (o *Or) String() string {
	if len(o.Terms) == 0 {
		return "false"
	}
	return NaryRepr("|", o.Terms)
}
func (o *Or) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(or %v terms:%v)", o.Base.canonFields(), SliceCanon(o.Terms)))
}

// Not is the negation of a formula.
type Not struct {
	Base
	Body Node
}

func NewNot(body Node) *Not {
	a := &Not{Body: body}
	a.Cfg = DefaultAstConfig
	return a
}

func (cfg *AstConfig) NewNot(body Node) *Not {
	a := &Not{Body: body}
	a.Cfg = cfg
	return a
}

func (n *Not) Args() []Node           { return []Node{n.Body} }
func (n *Not) Clone(args []Node) Node { return &Not{Base: n.Base, Body: args[0]} }
func (n *Not) String() string {
	if a, ok := n.Body.(*Atom); ok && IsEquals(a.Rep) {
		parts := make([]string, len(a.Terms))
		for i, t := range a.Terms {
			parts[i] = fmt.Sprint(t)
		}
		return strings.Join(parts, " ~= ")
	}
	return "~" + fmt.Sprint(n.Body)
}
func (n *Not) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(not %v body:%v)", n.Base.canonFields(), nodeCanon(n.Body)))
}

// Implies is a logical implication.
type Implies struct {
	Base
	T1, T2 Node
}

func NewImplies(t1, t2 Node) *Implies {
	a := &Implies{T1: t1, T2: t2}
	a.Cfg = DefaultAstConfig
	return a
}

func (cfg *AstConfig) NewImplies(t1, t2 Node) *Implies {
	a := &Implies{T1: t1, T2: t2}
	a.Cfg = cfg
	return a
}

func (i *Implies) Args() []Node { return []Node{i.T1, i.T2} }
func (i *Implies) Clone(args []Node) Node {
	return &Implies{Base: i.Base, T1: args[0], T2: args[1]}
}
func (i *Implies) String() string { return fmt.Sprint(i.T1) + " -> " + fmt.Sprint(i.T2) }
func (i *Implies) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(implies %v t1:%v t2:%v)", i.Base.canonFields(), nodeCanon(i.T1), nodeCanon(i.T2)))
}

// Iff is a biconditional (if and only if).
type Iff struct {
	Base
	T1, T2 Node
}

func NewIff(t1, t2 Node) *Iff {
	a := &Iff{T1: t1, T2: t2}
	a.Cfg = DefaultAstConfig
	return a
}

func (cfg *AstConfig) NewIff(t1, t2 Node) *Iff {
	a := &Iff{T1: t1, T2: t2}
	a.Cfg = cfg
	return a
}

func (f *Iff) Args() []Node { return []Node{f.T1, f.T2} }
func (f *Iff) Clone(args []Node) Node {
	return &Iff{Base: f.Base, T1: args[0], T2: args[1]}
}
func (f *Iff) String() string { return fmt.Sprint(f.T1) + " <-> " + fmt.Sprint(f.T2) }
func (f *Iff) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(iff %v t1:%v t2:%v)", f.Base.canonFields(), nodeCanon(f.T1), nodeCanon(f.T2)))
}

// Ite is an if-then-else expression.
type Ite struct {
	Base
	Cond, Then, Else Node
}

func NewIte(cond, then_, else_ Node) *Ite {
	a := &Ite{Cond: cond, Then: then_, Else: else_}
	a.Cfg = DefaultAstConfig
	return a
}

func (cfg *AstConfig) NewIte(cond, then_, else_ Node) *Ite {
	a := &Ite{Cond: cond, Then: then_, Else: else_}
	a.Cfg = cfg
	return a
}

func (i *Ite) Args() []Node { return []Node{i.Cond, i.Then, i.Else} }
func (i *Ite) Clone(args []Node) Node {
	return &Ite{Base: i.Base, Cond: args[0], Then: args[1], Else: args[2]}
}
func (i *Ite) String() string {
	return "(" + fmt.Sprint(i.Then) + " if " + fmt.Sprint(i.Cond) + " else " + fmt.Sprint(i.Else) + ")"
}
func (i *Ite) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(ite %v cond:%v then:%v else:%v)", i.Base.canonFields(), nodeCanon(i.Cond), nodeCanon(i.Then), nodeCanon(i.Else)))
}

// Forall is a universal quantifier with bounds.
type Forall struct {
	Base
	Bounds []Node // bound variables
	Body   Node
}

func NewForall(bounds []Node, body Node) *Forall {
	a := &Forall{Bounds: bounds, Body: body}
	a.Cfg = DefaultAstConfig
	return a
}

func (cfg *AstConfig) NewForall(bounds []Node, body Node) *Forall {
	a := &Forall{Bounds: bounds, Body: body}
	a.Cfg = cfg
	return a
}

func (f *Forall) Args() []Node           { return []Node{f.Body} }
func (f *Forall) Clone(args []Node) Node {
	return &Forall{Base: f.Base, Bounds: f.Bounds, Body: args[0]}
}
func (f *Forall) String() string {
	parts := make([]string, len(f.Bounds))
	for i, b := range f.Bounds {
		parts[i] = fmt.Sprint(b)
	}
	return "forall " + strings.Join(parts, ",") + ". " + fmt.Sprint(f.Body)
}
func (f *Forall) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(forall %v bounds:%v body:%v)", f.Base.canonFields(), SliceCanon(f.Bounds), nodeCanon(f.Body)))
}

// Exists is an existential quantifier with bounds.
type Exists struct {
	Base
	Bounds []Node
	Body   Node
}

func NewExists(bounds []Node, body Node) *Exists {
	a := &Exists{Bounds: bounds, Body: body}
	a.Cfg = DefaultAstConfig
	return a
}

func (cfg *AstConfig) NewExists(bounds []Node, body Node) *Exists {
	a := &Exists{Bounds: bounds, Body: body}
	a.Cfg = cfg
	return a
}

func (e *Exists) Args() []Node           { return []Node{e.Body} }
func (e *Exists) Clone(args []Node) Node {
	return &Exists{Base: e.Base, Bounds: e.Bounds, Body: args[0]}
}
func (e *Exists) String() string {
	parts := make([]string, len(e.Bounds))
	for i, b := range e.Bounds {
		parts[i] = fmt.Sprint(b)
	}
	return "exists " + strings.Join(parts, ",") + ". " + fmt.Sprint(e.Body)
}
func (e *Exists) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(exists %v bounds:%v body:%v)", e.Base.canonFields(), SliceCanon(e.Bounds), nodeCanon(e.Body)))
}

// Isa is a type test formula.
type Isa struct {
	Base
	Terms []Node
}

func (i *Isa) Args() []Node           { return i.Terms }
func (i *Isa) Clone(args []Node) Node { return &Isa{Base: i.Base, Terms: args} }
func (i *Isa) String() string         { return NaryRepr("isa", i.Terms) }
func (i *Isa) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(isa %v terms:%v)", i.Base.canonFields(), SliceCanon(i.Terms)))
}

// Globally is the temporal "globally" operator.
type Globally struct {
	Base
	Body Node
}

func NewGlobally(body Node) *Globally {
	a := &Globally{Body: body}
	a.Cfg = DefaultAstConfig
	return a
}

func (cfg *AstConfig) NewGlobally(body Node) *Globally {
	a := &Globally{Body: body}
	a.Cfg = cfg
	return a
}

func (g *Globally) Args() []Node           { return []Node{g.Body} }
func (g *Globally) Clone(args []Node) Node { return &Globally{Base: g.Base, Body: args[0]} }
func (g *Globally) String() string         { return "(globally " + fmt.Sprint(g.Body) + ")" }
func (g *Globally) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(globally %v body:%v)", g.Base.canonFields(), nodeCanon(g.Body)))
}

// Eventually is the temporal "eventually" operator.
type Eventually struct {
	Base
	Body Node
}

func NewEventually(body Node) *Eventually {
	a := &Eventually{Body: body}
	a.Cfg = DefaultAstConfig
	return a
}

func (cfg *AstConfig) NewEventually(body Node) *Eventually {
	a := &Eventually{Body: body}
	a.Cfg = cfg
	return a
}

func (e *Eventually) Args() []Node           { return []Node{e.Body} }
func (e *Eventually) Clone(args []Node) Node { return &Eventually{Base: e.Base, Body: args[0]} }
func (e *Eventually) String() string         { return "(eventually " + fmt.Sprint(e.Body) + ")" }
func (e *Eventually) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(eventually %v body:%v)", e.Base.canonFields(), nodeCanon(e.Body)))
}

// WhenOperator is the temporal "whennext" operator.
type WhenOperator struct {
	Base
	Name string
	T1   Node
	T2   Node
}

func NewWhenOperator(name string, t1, t2 Node) *WhenOperator {
	a := &WhenOperator{Name: name, T1: t1, T2: t2}
	a.Cfg = DefaultAstConfig
	return a
}

func (cfg *AstConfig) NewWhenOperator(name string, t1, t2 Node) *WhenOperator {
	a := &WhenOperator{Name: name, T1: t1, T2: t2}
	a.Cfg = cfg
	return a
}

func (w *WhenOperator) Args() []Node { return []Node{w.T1, w.T2} }
// Clone matches Python WhenOperator.clone (ivy_ast.py:144-148) which calls lineno_add_ref.
func (w *WhenOperator) Clone(args []Node) Node {
	c := &WhenOperator{Base: w.Base, Name: w.Name, T1: args[0], T2: args[1]}
	c.SetLineno(safeLinenoAddRef(w, w.GetLineno()))
	return c
}
func (w *WhenOperator) String() string {
	return "(" + fmt.Sprint(w.T1) + " " + w.Name + "whennext " + fmt.Sprint(w.T2) + ")"
}
func (w *WhenOperator) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(whenOperator %v name:%q t1:%v t2:%v)", w.Base.canonFields(), w.Name, nodeCanon(w.T1), nodeCanon(w.T2)))
}

// Let is "let p(X,...) <-> fmla, ... in fmla".
type Let struct {
	Base
	Defs []Node // definitions (all but last)
	Body Node   // body formula (last)
}

func NewLet(defs []Node, body Node) *Let {
	a := &Let{Defs: defs, Body: body}
	a.Cfg = DefaultAstConfig
	return a
}

func (cfg *AstConfig) NewLet(defs []Node, body Node) *Let {
	a := &Let{Defs: defs, Body: body}
	a.Cfg = cfg
	return a
}

func (l *Let) Args() []Node {
	return append(append([]Node{}, l.Defs...), l.Body)
}
func (l *Let) Clone(args []Node) Node {
	return &Let{Base: l.Base, Defs: args[:len(args)-1], Body: args[len(args)-1]}
}
func (l *Let) String() string {
	if len(l.Defs) == 0 {
		return fmt.Sprint(l.Body)
	}
	parts := make([]string, len(l.Defs))
	for i, d := range l.Defs {
		parts[i] = fmt.Sprint(d)
	}
	return "let " + strings.Join(parts, ", ") + " in " + fmt.Sprint(l.Body)
}
func (l *Let) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(let %v defs:%v body:%v)", l.Base.canonFields(), SliceCanon(l.Defs), nodeCanon(l.Body)))
}

// Definition is "p(X,...) = fmla".
type Definition struct {
	Base
	Lhs Node
	Rhs Node
}

func NewDefinition(lhs, rhs Node) *Definition {
	a := &Definition{Lhs: lhs, Rhs: rhs}
	a.Cfg = DefaultAstConfig
	return a
}

func (cfg *AstConfig) NewDefinition(lhs, rhs Node) *Definition {
	a := &Definition{Lhs: lhs, Rhs: rhs}
	a.Cfg = cfg
	return a
}

// ToConstraint converts a Definition to a constraint formula.
// If the LHS is an App (function definition), returns Atom("=", lhs, rhs).
// Otherwise returns Iff(lhs, rhs).
// Matches Python ivy_ast.py:188-191 Definition.to_constraint().
func (d *Definition) ToConstraint() Node {
	if _, ok := d.Lhs.(*App); ok {
		return NewAtom("=", d.Lhs, d.Rhs)
	}
	return NewIff(d.Lhs, d.Rhs)
}

func (d *Definition) Args() []Node { return []Node{d.Lhs, d.Rhs} }
func (d *Definition) Clone(args []Node) Node {
	return &Definition{Base: d.Base, Lhs: args[0], Rhs: args[1]}
}
func (d *Definition) String() string { return fmt.Sprint(d.Lhs) + " = " + fmt.Sprint(d.Rhs) }
func (d *Definition) Defines() string {
	return NodeRep(d.Lhs)
}
func (d *Definition) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(definition %v lhs:%v rhs:%v)", d.Base.canonFields(), nodeCanon(d.Lhs), nodeCanon(d.Rhs)))
}

// DefinitionSchema is a definition used as a schema.
type DefinitionSchema struct {
	Definition
}

func (ds *DefinitionSchema) Clone(args []Node) Node {
	return &DefinitionSchema{Definition: *ds.Definition.Clone(args).(*Definition)}
}
func (d *DefinitionSchema) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(definitionSchema definition:%v)", d.Definition.Canon()))
}

// NamedBinder is a binder with a name, bounds, and body.
type NamedBinder struct {
	Base
	Name   string
	Bounds []Node
	Body   Node
}

func NewNamedBinder(name string, bounds []Node, body Node) *NamedBinder {
	a := &NamedBinder{Name: name, Bounds: bounds, Body: body}
	a.Cfg = DefaultAstConfig
	return a
}

func (cfg *AstConfig) NewNamedBinder(name string, bounds []Node, body Node) *NamedBinder {
	a := &NamedBinder{Name: name, Bounds: bounds, Body: body}
	a.Cfg = cfg
	return a
}

func (n *NamedBinder) Args() []Node           { return []Node{n.Body} }
func (n *NamedBinder) Clone(args []Node) Node {
	return &NamedBinder{Base: n.Base, Name: n.Name, Bounds: n.Bounds, Body: args[0]}
}
func (n *NamedBinder) String() string {
	parts := make([]string, len(n.Bounds))
	for i, b := range n.Bounds {
		parts[i] = fmt.Sprint(b)
	}
	return n.Name + "(" + strings.Join(parts, ",") + "). " + fmt.Sprint(n.Body)
}
func (n *NamedBinder) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(namedBinder %v name:%q bounds:%v body:%v)", n.Base.canonFields(), n.Name, SliceCanon(n.Bounds), nodeCanon(n.Body)))
}

// Trigger is a quantifier trigger/pattern hint.
type Trigger struct {
	Base
	Pattern Node
	Terms   []Node
}

func (t *Trigger) Args() []Node {
	return append([]Node{t.Pattern}, t.Terms...)
}
func (t *Trigger) Clone(args []Node) Node {
	return &Trigger{Base: t.Base, Pattern: args[0], Terms: args[1:]}
}
func (t *Trigger) String() string {
	parts := make([]string, len(t.Terms))
	for i, s := range t.Terms {
		parts[i] = fmt.Sprint(s)
	}
	return "trigger" + fmt.Sprint(t.Pattern) + " with " + strings.Join(parts, ",")
}
func (t *Trigger) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(trigger %v pattern:%v terms:%v)", t.Base.canonFields(), nodeCanon(t.Pattern), SliceCanon(t.Terms)))
}
