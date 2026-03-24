// Package ast defines the abstract syntax tree for Ivy programs.
// This represents the parsed syntax tree before compilation to the logic IR.
// It is distinct from the logic/ package which represents the semantic IR.
package ast

import (
	"fmt"
	"strings"
	"sync/atomic"

	iu "github.com/glycerine/goivy/ivyutils"
)

// Location represents a source code position.
type Location struct {
	Filename string
	Line     int
}

func (s *Location) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(location filename:%v line:%v)", s.Filename, s.Line))
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

	// Canon returns a canonical (reproducible)
	// compact s-expression string, safe for hashing.
	// It must capture/represent all of the
	// ast.Node internal state.
	Canon() iu.Canonical
}

// Base provides common fields for all AST nodes.
type Base struct {
	Loc    Location
	HasLoc bool
}

func (b *Base) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(base hasLoc:%v loc:%v)", b.HasLoc, b.Loc))
}

// canonFields returns the flattened lineno fields from Base for inclusion
// in parent Canon() output. This avoids nesting (base:(base ...)) which
// Python cannot reproduce due to flat class inheritance.
func (b *Base) canonFields() string {
	// python's line numbers are off, omit for now.
	return "" // or fake: "lineno:0"

	if b.Loc.Filename != "" {
		return fmt.Sprintf("filename:%q lineno:%d", b.Loc.Filename, b.Loc.Line)
	}
	return fmt.Sprintf("lineno:%d", b.Loc.Line)
}

// nodeCanon returns the Canon() of a Node, or "nil" if the node is nil.
func nodeCanon(n Node) iu.Canonical {
	if n == nil {
		return "nil"
	}
	return n.Canon()
}

// sortCanon returns the canonical form of a sort Node.
// In Python, sorts on Atom/App/Variable are plain strings, so node_canon
// returns just the bare string. In Go, sorts are wrapped in *Symbol.
// This extracts the bare string to match Python.
func sortCanon(n Node) iu.Canonical {
	if n == nil {
		return "nil"
	}
	if sym, ok := n.(*Symbol); ok && sym.Sort == nil {
		return iu.Canonical(sym.Rep)
	}
	return n.Canon()
}

// sliceCanon returns the canonical form of a []Node slice.
func sliceCanon(nodes []Node) string {
	if len(nodes) == 0 {
		return "[]"
	}
	var sb strings.Builder
	sb.WriteByte('[')
	for i, n := range nodes {
		if i > 0 {
			sb.WriteByte(' ')
		}
		sb.WriteString(string(nodeCanon(n)))
	}
	sb.WriteByte(']')
	return sb.String()
}

// boolPtrCanon returns "nil", "true", or "false" for a *bool.
func boolPtrCanon(b *bool) string {
	if b == nil {
		return "nil"
	}
	if *b {
		return "true"
	}
	return "false"
}

// stringSliceCanon returns canonical form for []string.
func stringSliceCanon(ss []string) string {
	if len(ss) == 0 {
		return "[]"
	}
	var sb strings.Builder
	sb.WriteByte('[')
	for i, s := range ss {
		if i > 0 {
			sb.WriteByte(' ')
		}
		sb.WriteString(fmt.Sprintf("%q", s))
	}
	sb.WriteByte(']')
	return sb.String()
}

func (b *Base) GetLineno() Location  { return b.Loc }
func (b *Base) SetLineno(l Location) { b.Loc = l; b.HasLoc = true }
func (b *Base) HasLocSet() bool      { return b.HasLoc }

// SLN sets the line number and returns the base (for chaining).
func SLN(n Node, loc Location) {
	n.SetLineno(loc)
}

// --- Core node types ---

// NoneAST is a placeholder for absent optional nodes.
type NoneAST struct {
	Base
}

func (n *NoneAST) Args() []Node           { return nil }
func (n *NoneAST) Clone(args []Node) Node { return &NoneAST{Base: n.Base} }
func (n *NoneAST) String() string         { return "" }
func (n *NoneAST) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(noneAST %v)", n.Base.canonFields()))
}

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
func (s *Symbol) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(symbol %v rep:%q sort:%v)", s.Base.canonFields(), s.Rep, nodeCanon(s.Sort)))
}

// Atom is an n-ary relation/predicate applied to terms.
type Atom struct {
	Base
	Rep   string // relation name
	Terms []Node // arguments
	ASort Node   // optional sort
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

// DropPrefix removes a prefix from the atom's name if present.
// Matches Python: Atom.drop_prefix / Symbol.drop_prefix.
func (a *Atom) DropPrefix(s string) *Atom {
	if !strings.HasPrefix(a.Rep, s) {
		return a
	}
	c := a.Clone(a.Terms).(*Atom)
	c.Rep = a.Rep[len(s):]
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
func (a *Atom) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(atom %v rep:%q terms:%v aSort:%v)", a.Base.canonFields(), a.Rep, sliceCanon(a.Terms), sortCanon(a.ASort)))
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
func (a *App) Canon() iu.Canonical {
	// Python's App.rep is a string, so node_canon(self.rep) gives just
	// the bare string. Go's App.Rep is a Node (*Symbol). Extract the
	// bare string to match Python's canon output.
	repCanon := nodeCanon(a.Rep)
	if sym, ok := a.Rep.(*Symbol); ok {
		repCanon = iu.Canonical(sym.Rep)
	}
	return iu.Canonical(fmt.Sprintf("(app %v rep:%v terms:%v aSort:%v)", a.Base.canonFields(), repCanon, sliceCanon(a.Terms), sortCanon(a.ASort)))
}

// Variable represents a sorted variable in the AST.
type Variable struct {
	Base
	Rep   string
	VSort string // the sort
}

func NewVariable(rep string, sort string) *Variable {
	return &Variable{Rep: rep, VSort: sort}
}

func (v *Variable) Args() []Node           { return nil }
func (v *Variable) Clone(args []Node) Node { return v }
func (v *Variable) String() string {
	if v.VSort != "" {
		return v.Rep + ":" + v.VSort
	}
	return v.Rep
}
func (v *Variable) Relname() string { return v.Rep }

// ToConst creates an App with the given prefix prepended to the variable name,
// copying the sort. Matches Python ivy_ast.py Variable.to_const() which returns App.
func (v *Variable) ToConst(prefix string) *App {
	a := NewApp(&Symbol{Rep: prefix + v.Rep})
	a.ASort = &Symbol{Rep: v.VSort}
	return a
}

// ToConstAtom creates an Atom with the given prefix
// prepended to the atom name,
// copying the sort. Used for prm: prefix substitution.
func ToConstAtom(a *Atom, prefix string) *Atom {
	res := NewAtom(prefix+a.Rep, a.Terms...)
	res.Base = a.Base
	res.ASort = a.ASort
	return res
}

// ToConstApp creates an App with the given prefix
// prepended to the atom name,
// copying the sort. Used for prm: prefix substitution.
func ToConstApp(a *App, prefix string) (res *App, rep1 string) {
	var newRep *Symbol
	switch x := a.Rep.(type) {
	case *Symbol:
		rep1 = prefix + x.Rep
		newRep = &Symbol{Rep: rep1}
	default:
		panicf("how to handle %T ?", a.Rep)
	}
	res = NewApp(newRep, a.Terms...)
	res.Base = a.Base
	res.ASort = a.ASort
	return
}

func (v *Variable) Resort(sort string) *Variable {
	nv := &Variable{Base: v.Base, Rep: v.Rep, VSort: sort}
	return nv
}
func (v *Variable) Canon() iu.Canonical {
	var vsort string
	switch v.VSort {
	case "":
		vsort = "nil"
	case "this":
		vsort = string((&This{}).Canon())
	default:
		vsort = v.VSort
	}
	return iu.Canonical(fmt.Sprintf("(variable %v rep:%q vSort:%v)", v.Base.canonFields(), v.Rep, vsort))
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
func (o *Old) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(old %v term:%v)", o.Base.canonFields(), nodeCanon(o.Term)))
}

// This represents a self-reference.
type This struct {
	Base
}

func (t *This) Args() []Node           { return nil }
func (t *This) Clone(args []Node) Node { return &This{Base: t.Base} }
func (t *This) String() string         { return "this" }
func (t *This) Relname() string        { return "this" }
func (t *This) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(this %v)", t.Base.canonFields()))
}

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
func (m *MethodCall) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(methodCall %v obj:%v method:%v)", m.Base.canonFields(), nodeCanon(m.Obj), nodeCanon(m.Method)))
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
func (l *Literal) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(literal %v polarity:%d atom:%v)", l.Base.canonFields(), l.Polarity, nodeCanon(l.Atom)))
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
func (d *Dot) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(dot %v left:%v right:%v)", d.Base.canonFields(), nodeCanon(d.Left), nodeCanon(d.Right)))
}

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
func (b *Bracket) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(bracket %v left:%v right:%v)", b.Base.canonFields(), nodeCanon(b.Left), nodeCanon(b.Right)))
}

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
func (t *Tuple) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(tuple %v elems:%v)", t.Base.canonFields(), sliceCanon(t.Elems)))
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
func (s *Some) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(some %v params:%v fmla:%v)", s.Base.canonFields(), sliceCanon(s.Params), nodeCanon(s.Fmla)))
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
func (s *SomeMin) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(someMin %v params:%v fmla:%v index:%v)", s.Base.canonFields(), sliceCanon(s.Params), nodeCanon(s.Fmla), nodeCanon(s.Index)))
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
func (s *SomeMax) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(someMax %v params:%v fmla:%v index:%v)", s.Base.canonFields(), sliceCanon(s.Params), nodeCanon(s.Fmla), nodeCanon(s.Index)))
}

// SomeExpr represents "some X. phi in expr else expr".
type SomeExpr struct {
	Base
	Param   Node
	Fmla    Node
	IfValue Node // optional
	ElseVal Node // optional
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
func (s *SomeExpr) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(someExpr %v param:%v fmla:%v ifValue:%v elseVal:%v)", s.Base.canonFields(), nodeCanon(s.Param), nodeCanon(s.Fmla), nodeCanon(s.IfValue), nodeCanon(s.ElseVal)))
}

// KeyArg wraps an App with a ^ prefix (key argument).
type KeyArg struct {
	*App
}

func (k *KeyArg) String() string { return "^" + k.App.String() }
func (k *KeyArg) Clone(args []Node) Node {
	return &KeyArg{App: k.App.Clone(args).(*App)}
}
func (k *KeyArg) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(keyArg app:%v)", k.App.Canon()))
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
func (d *DebugItem) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(debugItem %v name:%v value:%v)", d.Base.canonFields(), nodeCanon(d.Name), nodeCanon(d.Value)))
}

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
func (t *ThunkAction) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(thunkAction %v label:%v action:%v sort:%v body:%v)", t.Base.canonFields(), nodeCanon(t.Label), nodeCanon(t.Action), nodeCanon(t.Sort), nodeCanon(t.Body)))
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
func (c *CrashAction) Clone(args []Node) Node { return &CrashAction{Base: c.Base, DeclArgs: args} }
func (c *CrashAction) String() string {
	if len(c.DeclArgs) > 0 {
		return "crash " + fmt.Sprint(c.DeclArgs[0])
	}
	return "crash"
}
func (c *CrashAction) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(crashAction %v declArgs:%v)", c.Base.canonFields(), sliceCanon(c.DeclArgs)))
}

// CallAction inlines a named state or action.
// Python: class CallAction(Action) from ivy_actions.py:1182.
// args[0] is the callee atom; args[1:] are actual returns.
var callActionCtr int

type CallAction struct {
	Base
	Elems    []Node
	UniqueID int
}

func NewCallAction(args ...Node) *CallAction {
	ca := &CallAction{Elems: args, UniqueID: callActionCtr}
	callActionCtr++
	return ca
}

func (c *CallAction) Args() []Node { return c.Elems }
func (c *CallAction) Clone(args []Node) Node {
	return &CallAction{Base: c.Base, Elems: args, UniqueID: c.UniqueID}
}
func (c *CallAction) String() string {
	if len(c.Elems) == 0 {
		return "call"
	}
	if len(c.Elems) == 1 {
		return "call " + fmt.Sprint(c.Elems[0])
	}
	// Python: 'call ' + (','.join(str(a) for a in actual_returns) + ' := ' if actual_returns else '') + str(self.args[0])
	returns := make([]string, len(c.Elems)-1)
	for i, a := range c.Elems[1:] {
		returns[i] = fmt.Sprint(a)
	}
	return "call " + strings.Join(returns, ",") + " := " + fmt.Sprint(c.Elems[0])
}
func (c *CallAction) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(callAction %v elems:%v uniqueID:%d)", c.Base.canonFields(), sliceCanon(c.Elems), c.UniqueID))
}

// Sequence represents an action sequence: { stmt; stmt; ... }.
// Python: Sequence(Action) from ivy_actions.py:773.
//
// In Python Ivy, Sequence and And are completely distinct types in different
// class hierarchies:
//   - Sequence extends Action — represents action blocks { stmt; stmt }
//   - And extends Formula — represents logical conjunction p & q
//
// The hand-rolled parser in parser/ conflates these, using ast.And for both.
// The new lalr_full LALR parser uses ast.Sequence to be faithful to the
// original Python grammar, which is important because downstream code
// (e.g. int_update, compose_updates) uses isinstance(x, Sequence) type
// checks that distinguish action sequences from logical conjunctions.
type Sequence struct {
	Base
	Stmts []Node
}

func NewSequence(stmts ...Node) *Sequence {
	return &Sequence{Stmts: stmts}
}

func (s *Sequence) Args() []Node           { return s.Stmts }
func (s *Sequence) Clone(args []Node) Node { return &Sequence{Base: s.Base, Stmts: args} }
func (s *Sequence) String() string {
	if len(s.Stmts) == 0 {
		return "{}"
	}
	parts := make([]string, len(s.Stmts))
	for i, st := range s.Stmts {
		parts[i] = fmt.Sprint(st)
	}
	return "{" + joinSemi(parts) + "}"
}
func (s *Sequence) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(sequence %v stmts:%v)", s.Base.canonFields(), sliceCanon(s.Stmts)))
}

// joinSemi joins strings with "; ".
func joinSemi(parts []string) string {
	result := ""
	for i, p := range parts {
		if i > 0 {
			result += "; "
		}
		result += p
	}
	return result
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
func (t *TemporalModels) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(temporalModels %v model:%v fmla:%v)", t.Base.canonFields(), nodeCanon(t.Model), nodeCanon(t.Fmla)))
}

// --- Equality methods ---
// Match Python's __eq__ semantics for each AST type.

// Equal returns true if other is a *Symbol with the same Rep.
func (s *Symbol) Equal(other Node) bool {
	o, ok := other.(*Symbol)
	return ok && s.Rep == o.Rep
}

// Equal returns true if other is an *Atom with the same Rep and Terms.
func (a *Atom) Equal(other Node) bool {
	o, ok := other.(*Atom)
	if !ok || a.Rep != o.Rep || len(a.Terms) != len(o.Terms) {
		return false
	}
	for i, t := range a.Terms {
		if eq, ok2 := t.(interface{ Equal(Node) bool }); ok2 {
			if !eq.Equal(o.Terms[i]) {
				return false
			}
		} else if t != o.Terms[i] {
			return false
		}
	}
	return true
}

// Equal returns true if other is an *App with the same Rep and Terms.
func (a *App) Equal(other Node) bool {
	o, ok := other.(*App)
	if !ok || len(a.Terms) != len(o.Terms) {
		return false
	}
	if repEq, ok2 := a.Rep.(interface{ Equal(Node) bool }); ok2 {
		if !repEq.Equal(o.Rep) {
			return false
		}
	} else if a.Rep != o.Rep {
		return false
	}
	for i, t := range a.Terms {
		if eq, ok2 := t.(interface{ Equal(Node) bool }); ok2 {
			if !eq.Equal(o.Terms[i]) {
				return false
			}
		} else if t != o.Terms[i] {
			return false
		}
	}
	return true
}

// Equal returns true if other is a *Variable with the same Rep.
func (v *Variable) Equal(other Node) bool {
	o, ok := other.(*Variable)
	return ok && v.Rep == o.Rep
}

// Equal returns true if other is a *Literal with the same Polarity and Atom.
func (l *Literal) Equal(other Node) bool {
	o, ok := other.(*Literal)
	if !ok || l.Polarity != o.Polarity {
		return false
	}
	if eq, ok2 := l.Atom.(interface{ Equal(Node) bool }); ok2 {
		return eq.Equal(o.Atom)
	}
	return l.Atom == o.Atom
}

// --- AppToAtom / AppsToAtoms ---

// AppToAtom converts an App to an Atom, preserving attributes.
// Does not convert Old, Some, SomeMin, SomeMax, Variable, or Ite nodes.
// Matches Python ivy_ast.py:1496-1506 app_to_atom.
func AppToAtom(app Node) Node {
	a, ok := app.(*App)
	if !ok {
		return app
	}
	// Don't convert special types that inherit from App in Python
	switch app.(type) {
	case *Variable:
		return app
	}
	res := NewAtom(fmt.Sprint(a.Rep), a.Terms...)
	res.Base = a.Base
	res.ASort = a.ASort
	return res
}

// AppsToAtoms converts a slice of Apps to Atoms.
// Matches Python ivy_ast.py:1508-1509 apps_to_atoms.
func AppsToAtoms(apps []Node) []Node {
	result := make([]Node, len(apps))
	for i, a := range apps {
		result[i] = AppToAtom(a)
	}
	return result
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
func (c *CompiledNode) String() string         { return fmt.Sprint(c.Node) }
func (c *CompiledNode) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(compiledNode %v)", c.Base.canonFields()))
}

// SetVariableSorts adds sorts to unsorted free variables in an AST node.
// subs maps variable names to sort AST nodes.
// Corresponds to Python ivy_ast.set_variable_sorts.
func SetVariableSorts(node Node, subs map[string]string) Node {
	if node == nil {
		return nil
	}
	if v, ok := node.(*Variable); ok {
		if sortNode, found := subs[v.Rep]; found {
			if v.VSort == "" || v.VSort == "S" {
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

func copySubst(m map[string]string) map[string]string {
	r := make(map[string]string, len(m))
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
