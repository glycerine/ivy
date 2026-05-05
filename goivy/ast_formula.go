package goivy

import (
	"fmt"
	"strings"
)

// --- Formula node types ---
// These represent the parsed syntax tree formulas, distinct from logic/ IR.

// And is a conjunction of formulas. Empty And = true.
type AstAnd struct {
	Base
	Terms []Node
}

func (cfg *AstConfig) NewAnd(terms ...Node) *AstAnd {
	a := &AstAnd{Terms: terms}
	a.Cfg = cfg
	return a
}

func (a *AstAnd) Args() []Node           { return a.Terms }
func (a *AstAnd) Clone(args []Node) Node { return &AstAnd{Base: a.Base, Terms: args} }
func (a *AstAnd) String() string {
	if len(a.Terms) == 0 {
		return "true"
	}
	return NaryRepr("&", a.Terms)
}
func (a *AstAnd) Canon() Canonical {
	return Canonical(fmt.Sprintf("(and%v terms:%v)", a.Base.canonFields(), SliceCanon(a.Terms)))
}

// Or is a disjunction of formulas. Empty Or = false.
type AstOr struct {
	Base
	Terms []Node
}

func (cfg *AstConfig) NewOr(terms ...Node) *AstOr {
	a := &AstOr{Terms: terms}
	a.Cfg = cfg
	return a
}

func (o *AstOr) Args() []Node           { return o.Terms }
func (o *AstOr) Clone(args []Node) Node { return &AstOr{Base: o.Base, Terms: args} }
func (o *AstOr) String() string {
	if len(o.Terms) == 0 {
		return "false"
	}
	return NaryRepr("|", o.Terms)
}
func (o *AstOr) Canon() Canonical {
	return Canonical(fmt.Sprintf("(or%v terms:%v)", o.Base.canonFields(), SliceCanon(o.Terms)))
}

// Not is the negation of a formula.
type AstNot struct {
	Base
	Body Node
}

func (cfg *AstConfig) NewNot(body Node) *AstNot {
	a := &AstNot{Body: body}
	a.Cfg = cfg
	return a
}

func (n *AstNot) Args() []Node           { return []Node{n.Body} }
func (n *AstNot) Clone(args []Node) Node { return &AstNot{Base: n.Base, Body: args[0]} }
func (n *AstNot) String() string {
	if a, ok := n.Body.(*Atom); ok && IsEquals(a.Rep) {
		parts := make([]string, len(a.Terms))
		for i, t := range a.Terms {
			parts[i] = fmt.Sprint(t)
		}
		return strings.Join(parts, " ~= ")
	}
	return "~" + fmt.Sprint(n.Body)
}
func (n *AstNot) Canon() Canonical {
	return Canonical(fmt.Sprintf("(not%v body:%v)", n.Base.canonFields(), nodeCanon(n.Body)))
}

// Implies is a logical implication.
type AstImplies struct {
	Base
	T1, T2 Node
}

func (cfg *AstConfig) NewImplies(t1, t2 Node) *AstImplies {
	a := &AstImplies{T1: t1, T2: t2}
	a.Cfg = cfg
	return a
}

func (i *AstImplies) Args() []Node { return []Node{i.T1, i.T2} }
func (i *AstImplies) Clone(args []Node) Node {
	return &AstImplies{Base: i.Base, T1: args[0], T2: args[1]}
}
func (i *AstImplies) String() string { return fmt.Sprint(i.T1) + " -> " + fmt.Sprint(i.T2) }
func (i *AstImplies) Canon() Canonical {
	return Canonical(fmt.Sprintf("(implies%v t1:%v t2:%v)", i.Base.canonFields(), nodeCanon(i.T1), nodeCanon(i.T2)))
}

// Iff is a biconditional (if and only if).
type AstIff struct {
	Base
	T1, T2 Node
}

func (cfg *AstConfig) NewIff(t1, t2 Node) *AstIff {
	a := &AstIff{T1: t1, T2: t2}
	a.Cfg = cfg
	return a
}

func (f *AstIff) Args() []Node { return []Node{f.T1, f.T2} }
func (f *AstIff) Clone(args []Node) Node {
	return &AstIff{Base: f.Base, T1: args[0], T2: args[1]}
}
func (f *AstIff) String() string { return fmt.Sprint(f.T1) + " <-> " + fmt.Sprint(f.T2) }
func (f *AstIff) Canon() Canonical {
	return Canonical(fmt.Sprintf("(iff%v t1:%v t2:%v)", f.Base.canonFields(), nodeCanon(f.T1), nodeCanon(f.T2)))
}

// Ite is an if-then-else expression.
type AstIte struct {
	Base
	Cond, Then, Else Node
}

func (cfg *AstConfig) NewIte(cond, then_, else_ Node) *AstIte {
	a := &AstIte{Cond: cond, Then: then_, Else: else_}
	a.Cfg = cfg
	return a
}

func (i *AstIte) Args() []Node { return []Node{i.Cond, i.Then, i.Else} }
func (i *AstIte) Clone(args []Node) Node {
	return &AstIte{Base: i.Base, Cond: args[0], Then: args[1], Else: args[2]}
}
func (i *AstIte) String() string {
	return "(" + fmt.Sprint(i.Then) + " if " + fmt.Sprint(i.Cond) + " else " + fmt.Sprint(i.Else) + ")"
}
func (i *AstIte) Canon() Canonical {
	return Canonical(fmt.Sprintf("(ite%v cond:%v then:%v else:%v)", i.Base.canonFields(), nodeCanon(i.Cond), nodeCanon(i.Then), nodeCanon(i.Else)))
}

// Forall is a universal quantifier with bounds.
type AstForall struct {
	Base
	Bounds []Node // bound variables
	Body   Node
}

func (cfg *AstConfig) NewForall(bounds []Node, body Node) *AstForall {
	a := &AstForall{Bounds: bounds, Body: body}
	a.Cfg = cfg
	return a
}

func (f *AstForall) Args() []Node { return []Node{f.Body} }
func (f *AstForall) Clone(args []Node) Node {
	return &AstForall{Base: f.Base, Bounds: f.Bounds, Body: args[0]}
}
func (f *AstForall) String() string {
	parts := make([]string, len(f.Bounds))
	for i, b := range f.Bounds {
		parts[i] = fmt.Sprint(b)
	}
	return "forall " + strings.Join(parts, ",") + ". " + fmt.Sprint(f.Body)
}
func (f *AstForall) Canon() Canonical {
	return Canonical(fmt.Sprintf("(forall%v bounds:%v body:%v)", f.Base.canonFields(), SliceCanon(f.Bounds), nodeCanon(f.Body)))
}

// Exists is an existential quantifier with bounds.
type AstExists struct {
	Base
	Bounds []Node
	Body   Node
}

func (cfg *AstConfig) NewExists(bounds []Node, body Node) *AstExists {
	a := &AstExists{Bounds: bounds, Body: body}
	a.Cfg = cfg
	return a
}

func (e *AstExists) Args() []Node { return []Node{e.Body} }
func (e *AstExists) Clone(args []Node) Node {
	return &AstExists{Base: e.Base, Bounds: e.Bounds, Body: args[0]}
}
func (e *AstExists) String() string {
	parts := make([]string, len(e.Bounds))
	for i, b := range e.Bounds {
		parts[i] = fmt.Sprint(b)
	}
	return "exists " + strings.Join(parts, ",") + ". " + fmt.Sprint(e.Body)
}
func (e *AstExists) Canon() Canonical {
	return Canonical(fmt.Sprintf("(exists%v bounds:%v body:%v)", e.Base.canonFields(), SliceCanon(e.Bounds), nodeCanon(e.Body)))
}

// Isa is a type test formula.
type AstIsa struct {
	Base
	Terms []Node
}

func (cfg *AstConfig) NewIsa(terms ...Node) *AstIsa {
	i := &AstIsa{Terms: terms}
	i.Cfg = cfg
	return i
}

func (i *AstIsa) Args() []Node           { return i.Terms }
func (i *AstIsa) Clone(args []Node) Node { return &AstIsa{Base: i.Base, Terms: args} }
func (i *AstIsa) String() string         { return NaryRepr("isa", i.Terms) }
func (i *AstIsa) Canon() Canonical {
	return Canonical(fmt.Sprintf("(isa%v terms:%v)", i.Base.canonFields(), SliceCanon(i.Terms)))
}

// Globally is the temporal "globally" operator.
type AstGlobally struct {
	Base
	Body Node
}

func (cfg *AstConfig) NewGlobally(body Node) *AstGlobally {
	a := &AstGlobally{Body: body}
	a.Cfg = cfg
	return a
}

func (g *AstGlobally) Args() []Node           { return []Node{g.Body} }
func (g *AstGlobally) Clone(args []Node) Node { return &AstGlobally{Base: g.Base, Body: args[0]} }
func (g *AstGlobally) String() string         { return "(globally " + fmt.Sprint(g.Body) + ")" }
func (g *AstGlobally) Canon() Canonical {
	return Canonical(fmt.Sprintf("(globally%v body:%v)", g.Base.canonFields(), nodeCanon(g.Body)))
}

// Eventually is the temporal "eventually" operator.
type AstEventually struct {
	Base
	Body Node
}

func (cfg *AstConfig) NewEventually(body Node) *AstEventually {
	a := &AstEventually{Body: body}
	a.Cfg = cfg
	return a
}

func (e *AstEventually) Args() []Node           { return []Node{e.Body} }
func (e *AstEventually) Clone(args []Node) Node { return &AstEventually{Base: e.Base, Body: args[0]} }
func (e *AstEventually) String() string         { return "(eventually " + fmt.Sprint(e.Body) + ")" }
func (e *AstEventually) Canon() Canonical {
	return Canonical(fmt.Sprintf("(eventually%v body:%v)", e.Base.canonFields(), nodeCanon(e.Body)))
}

// WhenOperator is the temporal "whennext" operator.
type AstWhenOperator struct {
	Base
	Name string
	T1   Node
	T2   Node
}

func (cfg *AstConfig) NewWhenOperator(name string, t1, t2 Node) *AstWhenOperator {
	a := &AstWhenOperator{Name: name, T1: t1, T2: t2}
	a.Cfg = cfg
	return a
}

func (w *AstWhenOperator) Args() []Node { return []Node{w.T1, w.T2} }

// Clone matches Python WhenOperator.clone (ivy_ast.py:144-148) which calls lineno_add_ref.
func (w *AstWhenOperator) Clone(args []Node) Node {
	c := &AstWhenOperator{Base: w.Base, Name: w.Name, T1: args[0], T2: args[1]}
	c.SetLineno(safeLinenoAddRef(w, w.GetLineno()))
	return c
}
func (w *AstWhenOperator) String() string {
	return "(" + fmt.Sprint(w.T1) + " " + w.Name + "whennext " + fmt.Sprint(w.T2) + ")"
}
func (w *AstWhenOperator) Canon() Canonical {
	return Canonical(fmt.Sprintf("(whenOperator%v name:%q t1:%v t2:%v)", w.Base.canonFields(), w.Name, nodeCanon(w.T1), nodeCanon(w.T2)))
}

// Let is "let p(X,...) <-> fmla, ... in fmla".
type AstLet struct {
	Base
	Defs []Node // definitions (all but last)
	Body Node   // body formula (last)
}

func (cfg *AstConfig) NewLet(defs []Node, body Node) *AstLet {
	a := &AstLet{Defs: defs, Body: body}
	a.Cfg = cfg
	return a
}

func (l *AstLet) Args() []Node {
	return append(append([]Node{}, l.Defs...), l.Body)
}
func (l *AstLet) Clone(args []Node) Node {
	return &AstLet{Base: l.Base, Defs: args[:len(args)-1], Body: args[len(args)-1]}
}
func (l *AstLet) String() string {
	if len(l.Defs) == 0 {
		return fmt.Sprint(l.Body)
	}
	parts := make([]string, len(l.Defs))
	for i, d := range l.Defs {
		parts[i] = fmt.Sprint(d)
	}
	return "let " + strings.Join(parts, ", ") + " in " + fmt.Sprint(l.Body)
}
func (l *AstLet) Canon() Canonical {
	return Canonical(fmt.Sprintf("(let%v defs:%v body:%v)", l.Base.canonFields(), SliceCanon(l.Defs), nodeCanon(l.Body)))
}

// Definition is "p(X,...) = fmla".
type AstDefinition struct {
	Base
	Lhs Node
	Rhs Node
}

func (cfg *AstConfig) NewDefinition(lhs, rhs Node) *AstDefinition {
	a := &AstDefinition{Lhs: lhs, Rhs: rhs}
	a.Cfg = cfg
	return a
}

// ToConstraint converts a Definition to a constraint formula.
// If the LHS is an App (function definition), returns Atom("=", lhs, rhs).
// Otherwise returns Iff(lhs, rhs).
// Matches Python ivy_ast.py:188-191 Definition.to_constraint().
func (d *AstDefinition) ToConstraint() Node {
	if _, ok := d.Lhs.(*App); ok {
		a := &Atom{Rep: "=", Terms: []Node{d.Lhs, d.Rhs}}
		a.Cfg = d.Cfg
		return a
	}
	iff := &AstIff{T1: d.Lhs, T2: d.Rhs}
	iff.Cfg = d.Cfg
	return iff
}

func (d *AstDefinition) Args() []Node { return []Node{d.Lhs, d.Rhs} }
func (d *AstDefinition) Clone(args []Node) Node {
	return &AstDefinition{Base: d.Base, Lhs: args[0], Rhs: args[1]}
}
func (d *AstDefinition) String() string { return fmt.Sprint(d.Lhs) + " = " + fmt.Sprint(d.Rhs) }
func (d *AstDefinition) Defines() string {
	return NodeRep(d.Lhs)
}
func (d *AstDefinition) Canon() Canonical {
	return Canonical(fmt.Sprintf("(definition%v lhs:%v rhs:%v)", d.Base.canonFields(), nodeCanon(d.Lhs), nodeCanon(d.Rhs)))
}

// DefinitionSchema is a definition used as a schema.
type AstDefinitionSchema struct {
	AstDefinition
}

func (cfg *AstConfig) NewDefinitionSchema(def AstDefinition) *AstDefinitionSchema {
	ds := &AstDefinitionSchema{AstDefinition: def}
	ds.Cfg = cfg
	return ds
}

func (ds *AstDefinitionSchema) Clone(args []Node) Node {
	return &AstDefinitionSchema{AstDefinition: *ds.AstDefinition.Clone(args).(*AstDefinition)}
}
func (d *AstDefinitionSchema) Canon() Canonical {
	return Canonical(fmt.Sprintf("(definitionSchema definition:%v)", d.AstDefinition.Canon()))
}

// NamedBinder is a binder with a name, bounds, and body.
type AstNamedBinder struct {
	Base
	Name   string
	Bounds []Node
	Body   Node
}

func (cfg *AstConfig) NewNamedBinder(name string, bounds []Node, body Node) *AstNamedBinder {
	a := &AstNamedBinder{Name: name, Bounds: bounds, Body: body}
	a.Cfg = cfg
	return a
}

func (n *AstNamedBinder) Args() []Node { return []Node{n.Body} }
func (n *AstNamedBinder) Clone(args []Node) Node {
	return &AstNamedBinder{Base: n.Base, Name: n.Name, Bounds: n.Bounds, Body: args[0]}
}
func (n *AstNamedBinder) String() string {
	parts := make([]string, len(n.Bounds))
	for i, b := range n.Bounds {
		parts[i] = fmt.Sprint(b)
	}
	return n.Name + "(" + strings.Join(parts, ",") + "). " + fmt.Sprint(n.Body)
}
func (n *AstNamedBinder) Canon() Canonical {
	return Canonical(fmt.Sprintf("(namedBinder%v name:%q bounds:%v body:%v)", n.Base.canonFields(), n.Name, SliceCanon(n.Bounds), nodeCanon(n.Body)))
}

// Trigger is a quantifier trigger/pattern hint.
type AstTrigger struct {
	Base
	Pattern Node
	Terms   []Node
}

func (cfg *AstConfig) NewTrigger(pattern Node, terms ...Node) *AstTrigger {
	t := &AstTrigger{Pattern: pattern, Terms: terms}
	t.Cfg = cfg
	return t
}

func (t *AstTrigger) Args() []Node {
	return append([]Node{t.Pattern}, t.Terms...)
}
func (t *AstTrigger) Clone(args []Node) Node {
	return &AstTrigger{Base: t.Base, Pattern: args[0], Terms: args[1:]}
}
func (t *AstTrigger) String() string {
	parts := make([]string, len(t.Terms))
	for i, s := range t.Terms {
		parts[i] = fmt.Sprint(s)
	}
	return "trigger" + fmt.Sprint(t.Pattern) + " with " + strings.Join(parts, ",")
}
func (t *AstTrigger) Canon() Canonical {
	return Canonical(fmt.Sprintf("(trigger%v pattern:%v terms:%v)", t.Base.canonFields(), nodeCanon(t.Pattern), SliceCanon(t.Terms)))
}
