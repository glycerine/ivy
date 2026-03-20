// Package ast defines the abstract syntax tree for Ivy programs.
// This represents the parsed syntax tree before compilation to the logic IR.
// It is distinct from the logic/ package which represents the semantic IR.
package ast

import (
	"fmt"
	"strings"
	"sync/atomic"
)

// Location represents a source code position.
type Location struct {
	Filename string
	Line     int
}

func (l Location) String() string {
	if l.Filename == "" {
		return fmt.Sprintf("line %d", l.Line)
	}
	return fmt.Sprintf("%s:%d", l.Filename, l.Line)
}

// Node is the interface implemented by all AST nodes.
type Node interface {
	// Args returns the child nodes for generic traversal.
	Args() []Node
	// Clone creates a copy of this node with different child args.
	Clone(args []Node) Node
	// Lineno returns the source location, if set.
	GetLineno() Location
	// SetLineno sets the source location.
	SetLineno(Location)
	// String returns a human-readable representation.
	String() string
}

// Base provides common fields for all AST nodes.
type Base struct {
	Loc      Location
	HasLoc   bool
}

func (b *Base) GetLineno() Location  { return b.Loc }
func (b *Base) SetLineno(l Location) { b.Loc = l; b.HasLoc = true }

// SLN sets the line number and returns the base (for chaining).
func SLN(n Node, loc Location) {
	n.SetLineno(loc)
}

// --- Core node types ---

// NoneAST is a placeholder for absent optional nodes.
type NoneAST struct {
	Base
}

func (n *NoneAST) Args() []Node              { return nil }
func (n *NoneAST) Clone(args []Node) Node    { return &NoneAST{Base: n.Base} }
func (n *NoneAST) String() string             { return "" }

// Symbol is a named identifier with an optional sort annotation.
type Symbol struct {
	Base
	Rep  string
	Sort Node // sort annotation, may be nil
}

func NewSymbol(rep string, sort Node) *Symbol {
	return &Symbol{Rep: rep, Sort: sort}
}

func (s *Symbol) Args() []Node           { return nil }
func (s *Symbol) Clone(args []Node) Node { return &Symbol{Base: s.Base, Rep: s.Rep, Sort: s.Sort} }
func (s *Symbol) String() string         { return s.Rep }
func (s *Symbol) Relname() string        { return s.Rep }

// Atom is an n-ary relation/predicate applied to terms.
type Atom struct {
	Base
	Rep     string // relation name
	Terms   []Node // arguments
	ASort   Node   // optional sort
}

func NewAtom(rep string, terms ...Node) *Atom {
	return &Atom{Rep: rep, Terms: terms}
}

func (a *Atom) Args() []Node { return a.Terms }
func (a *Atom) Clone(args []Node) Node {
	c := &Atom{Base: a.Base, Rep: a.Rep, Terms: args}
	c.ASort = a.ASort
	return c
}
func (a *Atom) String() string {
	if IsEquals(a.Rep) && len(a.Terms) == 2 {
		return fmt.Sprintf("%s = %s", a.Terms[0], a.Terms[1])
	}
	if len(a.Terms) == 0 {
		return a.Rep
	}
	parts := make([]string, len(a.Terms))
	for i, t := range a.Terms {
		parts[i] = fmt.Sprint(t)
	}
	return a.Rep + "(" + strings.Join(parts, ", ") + ")"
}
func (a *Atom) Relname() string { return a.Rep }

func (a *Atom) Prefix(s string) *Atom {
	c := a.Clone(a.Terms).(*Atom)
	c.Rep = s + c.Rep
	return c
}

func (a *Atom) Suffix(s string) *Atom {
	c := a.Clone(a.Terms).(*Atom)
	c.Rep = c.Rep + s
	return c
}

func (a *Atom) Rename(s string) *Atom {
	c := a.Clone(a.Terms).(*Atom)
	c.Rep = s
	c.ASort = a.ASort
	return c
}

// App is a function application (term level).
type App struct {
	Base
	Rep   Node   // function symbol (usually *Symbol)
	Terms []Node // arguments
	ASort Node   // optional sort annotation
}

func NewApp(rep Node, terms ...Node) *App {
	return &App{Rep: rep, Terms: terms}
}

func (a *App) Args() []Node { return a.Terms }
func (a *App) Clone(args []Node) Node {
	c := &App{Base: a.Base, Rep: a.Rep, Terms: args}
	c.ASort = a.ASort
	return c
}
func (a *App) String() string {
	if len(a.Terms) > 0 {
		parts := make([]string, len(a.Terms))
		for i, t := range a.Terms {
			parts[i] = fmt.Sprint(t)
		}
		res := fmt.Sprint(a.Rep) + "(" + strings.Join(parts, ", ") + ")"
		if a.ASort != nil {
			res += ":" + fmt.Sprint(a.ASort)
		}
		return res
	}
	res := fmt.Sprint(a.Rep)
	if a.ASort != nil {
		res += ":" + fmt.Sprint(a.ASort)
	}
	return res
}
func (a *App) Relname() string {
	if s, ok := a.Rep.(*Symbol); ok {
		return s.Rep
	}
	return fmt.Sprint(a.Rep)
}

func (a *App) Prefix(s string) *App {
	c := a.Clone(a.Terms).(*App)
	if sym, ok := c.Rep.(*Symbol); ok {
		c.Rep = NewSymbol(s+sym.Rep, sym.Sort)
	}
	return c
}

func (a *App) DropPrefix(s string) *App {
	c := a.Clone(a.Terms).(*App)
	if sym, ok := c.Rep.(*Symbol); ok && strings.HasPrefix(sym.Rep, s) {
		c.Rep = NewSymbol(sym.Rep[len(s):], sym.Sort)
	}
	c.ASort = a.ASort
	return c
}

func (a *App) Rename(s string) *App {
	c := a.Clone(a.Terms).(*App)
	if sym, ok := c.Rep.(*Symbol); ok {
		c.Rep = NewSymbol(s, sym.Sort)
	}
	c.ASort = a.ASort
	return c
}

// Variable represents a sorted variable in the AST.
type Variable struct {
	Base
	Rep   string
	VSort Node // sort annotation
}

func NewVariable(rep string, sort Node) *Variable {
	return &Variable{Rep: rep, VSort: sort}
}

func (v *Variable) Args() []Node           { return nil }
func (v *Variable) Clone(args []Node) Node { return v }
func (v *Variable) String() string {
	if v.VSort != nil {
		return v.Rep + ":" + fmt.Sprint(v.VSort)
	}
	return v.Rep
}
func (v *Variable) Relname() string { return v.Rep }

// ToConst creates an Atom with the given prefix prepended to the variable name,
// copying the sort. Matches Python ivy_ast.py Variable.to_const().
func (v *Variable) ToConst(prefix string) *Atom {
	a := NewAtom(prefix + v.Rep)
	a.ASort = v.VSort
	return a
}

// ToConstAtom creates an Atom with the given prefix prepended to the atom name,
// copying the sort. Used for prm: prefix substitution.
func ToConstAtom(a *Atom, prefix string) *Atom {
	res := NewAtom(prefix+a.Rep, a.Terms...)
	res.ASort = a.ASort
	return res
}

func (v *Variable) Resort(sort Node) *Variable {
	nv := &Variable{Base: v.Base, Rep: v.Rep, VSort: sort}
	return nv
}

// Old wraps a term with the temporal "old" operator.
type Old struct {
	Base
	Term Node
}

func NewOld(term Node) *Old { return &Old{Term: term} }

func (o *Old) Args() []Node           { return []Node{o.Term} }
func (o *Old) Clone(args []Node) Node { return &Old{Base: o.Base, Term: args[0]} }
func (o *Old) String() string         { return "old " + fmt.Sprint(o.Term) }

// This represents a self-reference.
type This struct {
	Base
}

func (t *This) Args() []Node           { return nil }
func (t *This) Clone(args []Node) Node { return &This{Base: t.Base} }
func (t *This) String() string         { return "this" }
func (t *This) Relname() string        { return "this" }

// MethodCall represents obj.method style calls.
type MethodCall struct {
	Base
	Obj    Node
	Method Node
}

func (m *MethodCall) Args() []Node { return []Node{m.Obj, m.Method} }
func (m *MethodCall) Clone(args []Node) Node {
	return &MethodCall{Base: m.Base, Obj: args[0], Method: args[1]}
}
func (m *MethodCall) String() string {
	return fmt.Sprint(m.Obj) + "." + fmt.Sprint(m.Method)
}

// Literal is a positive or negative atomic formula.
type Literal struct {
	Base
	Polarity int // 1 = positive, 0 = negative
	Atom     Node
}

func NewLiteral(polarity int, atom Node) *Literal {
	return &Literal{Polarity: polarity, Atom: atom}
}

func (l *Literal) Args() []Node { return []Node{l.Atom} }
func (l *Literal) Clone(args []Node) Node {
	return &Literal{Base: l.Base, Polarity: l.Polarity, Atom: args[0]}
}
func (l *Literal) String() string {
	if l.Polarity == 0 {
		return "~" + fmt.Sprint(l.Atom)
	}
	return fmt.Sprint(l.Atom)
}
func (l *Literal) Invert() *Literal {
	return &Literal{Base: l.Base, Polarity: 1 - l.Polarity, Atom: l.Atom}
}

// Dot represents field access (a.b).
type Dot struct {
	Base
	Left  Node
	Right Node
}

func NewDot(left, right Node) *Dot { return &Dot{Left: left, Right: right} }

func (d *Dot) Args() []Node { return []Node{d.Left, d.Right} }
func (d *Dot) Clone(args []Node) Node {
	return &Dot{Base: d.Base, Left: args[0], Right: args[1]}
}
func (d *Dot) String() string { return fmt.Sprint(d.Left) + "." + fmt.Sprint(d.Right) }

// Bracket represents subscript access (a[b]).
type Bracket struct {
	Base
	Left  Node
	Right Node
}

func NewBracket(left, right Node) *Bracket { return &Bracket{Left: left, Right: right} }

func (b *Bracket) Args() []Node { return []Node{b.Left, b.Right} }
func (b *Bracket) Clone(args []Node) Node {
	return &Bracket{Base: b.Base, Left: args[0], Right: args[1]}
}
func (b *Bracket) String() string { return fmt.Sprint(b.Left) + "[" + fmt.Sprint(b.Right) + "]" }

// Tuple wraps a list of nodes in parentheses.
type Tuple struct {
	Base
	Elems []Node
}

func NewTuple(elems ...Node) *Tuple { return &Tuple{Elems: elems} }

func (t *Tuple) Args() []Node           { return t.Elems }
func (t *Tuple) Clone(args []Node) Node { return &Tuple{Base: t.Base, Elems: args} }
func (t *Tuple) String() string {
	parts := make([]string, len(t.Elems))
	for i, e := range t.Elems {
		parts[i] = fmt.Sprint(e)
	}
	return "(" + strings.Join(parts, ",") + ")"
}

// Some represents "some X. phi" existential choice.
type Some struct {
	Base
	Params []Node // bound variables (all but last)
	Fmla   Node   // formula (last arg)
}

func NewSome(params []Node, fmla Node) *Some {
	return &Some{Params: params, Fmla: fmla}
}

func (s *Some) Args() []Node {
	return append(append([]Node{}, s.Params...), s.Fmla)
}
func (s *Some) Clone(args []Node) Node {
	return &Some{Base: s.Base, Params: args[:len(args)-1], Fmla: args[len(args)-1]}
}
func (s *Some) String() string {
	parts := make([]string, len(s.Params))
	for i, p := range s.Params {
		parts[i] = fmt.Sprint(p)
	}
	return "some " + strings.Join(parts, ",") + ". " + fmt.Sprint(s.Fmla)
}

// SomeMin represents "some X. phi minimizing idx".
type SomeMin struct {
	Base
	Params []Node
	Fmla   Node
	Index  Node
}

func (s *SomeMin) Args() []Node {
	args := append([]Node{}, s.Params...)
	args = append(args, s.Fmla, s.Index)
	return args
}
func (s *SomeMin) Clone(args []Node) Node {
	n := len(args)
	return &SomeMin{Base: s.Base, Params: args[:n-2], Fmla: args[n-2], Index: args[n-1]}
}
func (s *SomeMin) String() string {
	parts := make([]string, len(s.Params))
	for i, p := range s.Params {
		parts[i] = fmt.Sprint(p)
	}
	return "some " + strings.Join(parts, ",") + ". " + fmt.Sprint(s.Fmla) + " minimizing " + fmt.Sprint(s.Index)
}

// SomeMax represents "some X. phi maximizing idx".
type SomeMax struct {
	Base
	Params []Node
	Fmla   Node
	Index  Node
}

func (s *SomeMax) Args() []Node {
	args := append([]Node{}, s.Params...)
	args = append(args, s.Fmla, s.Index)
	return args
}
func (s *SomeMax) Clone(args []Node) Node {
	n := len(args)
	return &SomeMax{Base: s.Base, Params: args[:n-2], Fmla: args[n-2], Index: args[n-1]}
}
func (s *SomeMax) String() string {
	parts := make([]string, len(s.Params))
	for i, p := range s.Params {
		parts[i] = fmt.Sprint(p)
	}
	return "some " + strings.Join(parts, ",") + ". " + fmt.Sprint(s.Fmla) + " maximizing " + fmt.Sprint(s.Index)
}

// SomeExpr represents "some X. phi in expr else expr".
type SomeExpr struct {
	Base
	Param    Node
	Fmla     Node
	IfValue  Node // optional
	ElseVal  Node // optional
}

func (s *SomeExpr) Args() []Node {
	args := []Node{s.Param, s.Fmla}
	if s.IfValue != nil {
		args = append(args, s.IfValue)
	}
	if s.ElseVal != nil {
		args = append(args, s.ElseVal)
	}
	return args
}
func (s *SomeExpr) Clone(args []Node) Node {
	c := &SomeExpr{Base: s.Base, Param: args[0], Fmla: args[1]}
	if len(args) >= 3 {
		c.IfValue = args[2]
	}
	if len(args) >= 4 {
		c.ElseVal = args[3]
	}
	return c
}
func (s *SomeExpr) String() string {
	res := "some " + fmt.Sprint(s.Param) + ". " + fmt.Sprint(s.Fmla)
	if s.IfValue != nil {
		res += " in " + fmt.Sprint(s.IfValue)
	}
	if s.ElseVal != nil {
		res += " else " + fmt.Sprint(s.ElseVal)
	}
	return res
}

// KeyArg wraps an App with a ^ prefix (key argument).
type KeyArg struct {
	*App
}

func (k *KeyArg) String() string { return "^" + k.App.String() }
func (k *KeyArg) Clone(args []Node) Node {
	return &KeyArg{App: k.App.Clone(args).(*App)}
}

// DebugItem holds a name=value pair for debugging.
type DebugItem struct {
	Base
	Name  Node
	Value Node
}

func (d *DebugItem) Args() []Node { return []Node{d.Name, d.Value} }
func (d *DebugItem) Clone(args []Node) Node {
	return &DebugItem{Base: d.Base, Name: args[0], Value: args[1]}
}
func (d *DebugItem) String() string { return fmt.Sprint(d.Name) + "=" + fmt.Sprint(d.Value) }

// ThunkAction represents "thunk [label] name(args) : type := body".
// Python: ThunkAction(Action) from ivy_actions.py
// args = [label, action_atom, type, body]
type ThunkAction struct {
	Base
	Label  Node // the label (Atom)
	Action Node // the action name with args (Atom)
	Sort   Node // the sort/type
	Body   Node // the body sequence
}

func NewThunkAction(label, action, sort, body Node) *ThunkAction {
	return &ThunkAction{Label: label, Action: action, Sort: sort, Body: body}
}

func (t *ThunkAction) Args() []Node { return []Node{t.Label, t.Action, t.Sort, t.Body} }
func (t *ThunkAction) Clone(args []Node) Node {
	return &ThunkAction{Base: t.Base, Label: args[0], Action: args[1], Sort: args[2], Body: args[3]}
}
func (t *ThunkAction) String() string {
	return "thunk [" + fmt.Sprint(t.Label) + "] " + fmt.Sprint(t.Action) + " : " + fmt.Sprint(t.Sort) + " := " + fmt.Sprint(t.Body)
}

// TemporalModels represents M |= phi.
// CrashAction represents "action name = *" (havoc/crash).
// Python: CrashAction(Action) from ivy_actions.py
type CrashAction struct {
	Base
	DeclArgs []Node
}

func NewCrashAction(args ...Node) *CrashAction {
	return &CrashAction{DeclArgs: args}
}

func (c *CrashAction) Args() []Node           { return c.DeclArgs }
func (c *CrashAction) Clone(args []Node) Node  { return &CrashAction{Base: c.Base, DeclArgs: args} }
func (c *CrashAction) String() string {
	if len(c.DeclArgs) > 0 {
		return "crash " + fmt.Sprint(c.DeclArgs[0])
	}
	return "crash"
}

type TemporalModels struct {
	Base
	Model Node
	Fmla  Node
}

func (t *TemporalModels) Args() []Node { return []Node{t.Fmla} }
func (t *TemporalModels) Clone(args []Node) Node {
	return &TemporalModels{Base: t.Base, Model: t.Model, Fmla: args[0]}
}
func (t *TemporalModels) String() string {
	return fmt.Sprint(t.Model) + " |= " + fmt.Sprint(t.Fmla)
}

// --- Predefined constants ---

// Equals is the predefined equality symbol.
var Equals = NewSymbol("=", &RelationSort{Dom: []Node{nil, nil}})

// IsEquals checks if a name is the equality symbol.
func IsEquals(name string) bool {
	return name == "="
}

// --- Labeled formula counter ---

var lfCounter int64

func nextLFID() int64 {
	return atomic.AddInt64(&lfCounter, 1) - 1
}

// --- Helpers ---

// NaryRepr formats args joined by op.
func NaryRepr(op string, args []Node) string {
	parts := make([]string, len(args))
	for i, a := range args {
		parts[i] = fmt.Sprint(a)
	}
	res := strings.Join(parts, " "+op+" ")
	if len(args) > 1 {
		return "(" + res + ")"
	}
	return res
}

// CopyLineno copies source location from src to dst if set.
func CopyLineno(src, dst Node) {
	if b, ok := src.(interface{ GetLineno() Location }); ok {
		loc := b.GetLineno()
		if loc.Line != 0 || loc.Filename != "" {
			dst.SetLineno(loc)
		}
	}
}

// IsTrue checks if an AST node represents true (empty And).
func IsTrue(n Node) bool {
	if a, ok := n.(*And); ok {
		return len(a.Terms) == 0
	}
	return false
}

// IsFalse checks if an AST node represents false (empty Or).
func IsFalse(n Node) bool {
	if o, ok := n.(*Or); ok {
		return len(o.Terms) == 0
	}
	return false
}

// HasTemporal checks if a formula contains temporal operators.
func HasTemporal(f Node) bool {
	if f == nil {
		return false
	}
	switch f.(type) {
	case *Globally, *Eventually, *WhenOperator:
		return true
	}
	for _, arg := range f.Args() {
		if HasTemporal(arg) {
			return true
		}
	}
	return false
}

// CompiledNode wraps a compiled logic node (interface{}) as an AST node.
// This is used when tactic compilation produces a compiled formula that
// needs to be stored back into an AST-level tactic node (e.g., the
// condition of an IfTactic after sort inference).
type CompiledNode struct {
	Base
	Node interface{} // holds a lg.Expr or similar compiled result
}

func (c *CompiledNode) Args() []Node           { return nil }
func (c *CompiledNode) Clone(args []Node) Node { return &CompiledNode{Base: c.Base, Node: c.Node} }
func (c *CompiledNode) String() string          { return fmt.Sprint(c.Node) }

// SetVariableSorts adds sorts to unsorted free variables in an AST node.
// subs maps variable names to sort AST nodes.
// Corresponds to Python ivy_ast.set_variable_sorts.
func SetVariableSorts(node Node, subs map[string]Node) Node {
	if node == nil {
		return nil
	}
	if v, ok := node.(*Variable); ok {
		if sortNode, found := subs[v.Rep]; found {
			if v.VSort == nil || extractSortRep(v.VSort) == "S" {
				return &Variable{Base: v.Base, Rep: v.Rep, VSort: sortNode}
			}
		}
		return v
	}
	// For quantifiers and named binders, remove bound variables from subs
	switch n := node.(type) {
	case *Forall:
		newSubs := copySubst(subs)
		for _, b := range n.Bounds {
			if v, ok := b.(*Variable); ok {
				delete(newSubs, v.Rep)
			}
		}
		args := n.Args()
		newArgs := make([]Node, len(args))
		for i, a := range args {
			newArgs[i] = SetVariableSorts(a, newSubs)
		}
		return n.Clone(newArgs)
	case *Exists:
		newSubs := copySubst(subs)
		for _, b := range n.Bounds {
			if v, ok := b.(*Variable); ok {
				delete(newSubs, v.Rep)
			}
		}
		args := n.Args()
		newArgs := make([]Node, len(args))
		for i, a := range args {
			newArgs[i] = SetVariableSorts(a, newSubs)
		}
		return n.Clone(newArgs)
	case *NamedBinder:
		newSubs := copySubst(subs)
		for _, b := range n.Bounds {
			if v, ok := b.(*Variable); ok {
				delete(newSubs, v.Rep)
			}
		}
		args := n.Args()
		newArgs := make([]Node, len(args))
		for i, a := range args {
			newArgs[i] = SetVariableSorts(a, newSubs)
		}
		return n.Clone(newArgs)
	}

	args := node.Args()
	if len(args) == 0 {
		return node
	}
	newArgs := make([]Node, len(args))
	for i, a := range args {
		newArgs[i] = SetVariableSorts(a, subs)
	}
	return node.Clone(newArgs)
}

func copySubst(m map[string]Node) map[string]Node {
	r := make(map[string]Node, len(m))
	for k, v := range m {
		r[k] = v
	}
	return r
}

func extractSortRep(n Node) string {
	if s, ok := n.(*Symbol); ok {
		return s.Rep
	}
	return fmt.Sprint(n)
}
