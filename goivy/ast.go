// Package ast defines the abstract syntax tree for Ivy programs.
// This represents the parsed syntax tree before compilation to the logic IR.
// It is distinct from the logic/ package which represents the semantic IR.
package goivy

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/glycerine/ivy/goivy/xtracer"
)

// Location represents a source code position.
// Python: LocationTuple([filename, line]) or LocationTuple([filename, line, original_lineno])
// The Reference field chains to the original location before module instantiation.
type Location struct {
	Filename  string
	Line      int
	Reference *Location // Python: LocationTuple[2] — original lineno before instantiation
}

func (s *Location) Canon() Canonical {
	return Canonical(fmt.Sprintf("(location filename:%v line:%v)", s.Filename, s.Line))
}

// String returns a human-readable location string matching
// Python's LocationTuple.__str__ on Unix:
//
//	filename: line N:
//
// When Reference is set (from lineno_add_ref during module instantiation),
// delegates to the reference's String().
func (l Location) String() string {
	if l.Reference != nil {
		return l.Reference.String()
	}
	res := ""
	if l.Filename != "" {
		res += l.Filename + ": "
	}
	if l.Line > 0 {
		res += "line " + strconv.Itoa(l.Line) + ": "
	}
	return res
}

// FileLineKey returns a compact "file:line" string for use as a comparison
// key. Matches the format used by actions/update.go for CheckedAssert.
func (l Location) FileLineKey() string {
	return l.Filename + ":" + strconv.Itoa(l.Line)
}

// safeLinenoAddRef extracts cfg from a node and applies LinenoAddRef.
// Returns loc unchanged if the node has no AstConfig (nil-safe).
func safeLinenoAddRef(n Node, loc Location) Location {
	if cfg := n.GetAstConfig(); cfg != nil {
		return cfg.LinenoAddRef(loc)
	}
	return loc
}

// Node is the interface implemented by all AST nodes.
//
// Note that the logic.Expr interface embeds the
// ast.Node interface, so any lg.Expr is also an ast.Node.
// See ~/goivy/logic/node.go for all details.
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
	Canon() Canonical

	// GetAstConfig returns the AstConfig stored on the node's Base.
	// This enables Clone methods and other per-node operations to
	// access session state (counters, referenceLineno, etc.) without globals.
	GetAstConfig() *AstConfig
}

// reprNode returns Repr() if available on the node, else String().
// Used by Atom.Repr() to get sort-qualified display of children.
// The reprer interface is checked at runtime, catching lg.Variable.Repr(),
// lg.Apply.Repr(), etc. without ast importing logic.
func reprNode(n Node) string {
	type reprer interface{ Repr() string }
	if r, ok := n.(reprer); ok {
		return r.Repr()
	}
	return fmt.Sprint(n)
}

// GetFormalSortAnnotation extracts the sort-annotation string from an AST
// node that carries one. It returns "" for nodes that have no sort annotation.
//
// This is one of three sort-name extraction helpers in the codebase.
// All three are needed; they operate on different types and answer
// different questions:
//
//  1. ivylogic.IvySortName(s logic.Sort) string      [ivylogic/ivylogic.go:11]
//     Input:  a compiled logic.Sort (e.g. *UninterpretedSort, *LogicFunctionSort).
//     Returns: the name of that compiled sort ("bool", "int", "node -> data").
//     Used:   53 call sites across compiler/, module/, solver/, actions/.
//
//  2. astExtractSortRep(n ast.Node) string           [compiler/compiler.go:1435]
//     Input:  an AST node that IS a sort reference (e.g. an Atom whose Rep
//     ;       is the sort name, or a Symbol whose Rep is the sort name).
//     Returns: the node's Rep — the name the node represents.
//     Example: astExtractSortRep(Atom{Rep:"int"}) → "int"
//     Used:   20+ call sites in compiler/decl.go, compiler/compiler.go,
//     ;       compiler/phase6.go, compiler/helpers.go.
//
//  3. GetFormalSortAnnotation(n ast.Node) string    [ast/ast.go — this function]
//     Input:  an AST node that HAS a sort annotation (e.g. a Variable with
//     VSort "int", or an App with ASort pointing to a sort node).
//     Returns: the sort annotation on the node — what type the node is declared as.
//     Example: GetFormalSortAnnotation(Variable{Rep:"x", VSort:"int"}) → "int"
//     Used:   in compiler/helpers.go CompileInlineCall to get the sort of formal
//     parameters and return values from ActionInfo.FormalAST/FormalRetAST,
//     matching Python's p.sort access in ivy_compiler.py compile_inline_call.
//
// The 8 AST types that carry sort annotations (and return non-empty from this function):
//
//	ast/ast.go:
//	  Variable     (line 373)  — VSort string        e.g. "tar_clock"
//	  Atom         (line 218)  — ASort Node          e.g. a Symbol node for the sort
//	  App          (line 286)  — ASort Node          e.g. a sort annotation on func application
//	  Symbol       (line 197)  — Sort  Node          e.g. a sort annotation on an identifier
//	  ThunkAction  (line 762)  — Sort  Node          e.g. the type of the thunk
//
//	ast/decl_ast.go:
//	  VariantDef   (line 669)  — VSort Node          e.g. the supertype sort
//	  NativeExpr   (line 1621) — ASort Node          e.g. sort on a native expression
//	  Instantiation(line 1713) — Sort  Node          e.g. the sort in "name : sort"
//
// All other ast.Node types (~40+ of them) return "" from this function.
func GetFormalSortAnnotation(n Node) string {
	if n == nil {
		return ""
	}
	switch v := n.(type) {
	case *Variable:
		return v.VSort
	case *Atom:
		if v.ASort != nil {
			return fmt.Sprint(v.ASort)
		}
		return ""
	case *App:
		if v.ASort != nil {
			return fmt.Sprint(v.ASort)
		}
		return ""
	case *Symbol:
		if v.Sort != nil {
			return fmt.Sprint(v.Sort)
		}
		return ""
	case *ThunkAction:
		if v.Sort != nil {
			return fmt.Sprint(v.Sort)
		}
		return ""
	case *VariantDef:
		if v.VSort != nil {
			return fmt.Sprint(v.VSort)
		}
		return ""
	case *NativeExpr:
		if v.ASort != nil {
			return fmt.Sprint(v.ASort)
		}
		return ""
	case *Instantiation:
		if v.Sort != nil {
			return fmt.Sprint(v.Sort)
		}
		return ""
	default:
		return ""
	}
}

// Base provides common fields for all AST nodes.
type Base struct {
	Loc    Location
	HasLoc bool
	Cfg    *AstConfig `json:"-"` // per-session config; set by constructors, propagated by Clone
}

func (b *Base) Canon() Canonical {
	return Canonical(fmt.Sprintf("(base hasLoc:%v loc:%v)", b.HasLoc, b.Loc))
}

// GetAstConfig returns the AstConfig stored on this node.
// Satisfies the Node interface. Returns nil if not set.
func (b *Base) GetAstConfig() *AstConfig { return b.Cfg }

// canonFields returns the flattened lineno fields from Base for inclusion
// in parent Canon() output. This avoids nesting (base:(base ...)) which
// Python cannot reproduce due to flat class inheritance.
func (b *Base) canonFields() string {
	// python's line numbers are off, omit for now.
	return "" // or fake: " lineno:0"

	// When enabled, return with leading space so callers can use
	// "(typeName%v field:..." with no double-space when empty.
	if b.Loc.Filename != "" {
		return fmt.Sprintf(" filename:%q lineno:%d", b.Loc.Filename, b.Loc.Line)
	}
	return fmt.Sprintf(" lineno:%d", b.Loc.Line)
}

// nodeCanon returns the Canon() of a Node, or "nil" if the node is nil.
func nodeCanon(n Node) Canonical {
	if n == nil {
		return "nil"
	}
	return n.Canon()
}

// sortCanon returns the canonical form of a sort Node.
// In Python, sorts on Atom/App/Variable are plain strings, so node_canon
// returns just the bare string. In Go, sorts are wrapped in *Symbol.
// This extracts the bare string to match Python.
func sortCanon(n Node) Canonical {
	if n == nil {
		return "nil"
	}
	if sym, ok := n.(*Symbol); ok && sym.Sort == nil {
		return Canonical(sym.Rep)
	}
	return n.Canon()
}

// SliceCanon returns the canonical form of a []Node slice.
func SliceCanon(nodes []Node) string {
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
func astStringSliceCanon(ss []string) string {
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
func (n *NoneAST) Canon() Canonical {
	return Canonical(fmt.Sprintf("(noneAST%v)", n.Base.canonFields()))
}

func (cfg *AstConfig) NewNoneAST() *NoneAST {
	return &NoneAST{Base: Base{Cfg: cfg}}
}

// Symbol is a named identifier with an optional sort annotation.
type Symbol struct {
	Base
	Rep  string
	Sort Node // sort annotation, may be nil
}

func (cfg *AstConfig) NewSymbol(rep string, sort Node) *Symbol {
	s := &Symbol{Rep: rep, Sort: sort}
	s.Cfg = cfg
	return s
}

func (s *Symbol) Args() []Node           { return nil }
func (s *Symbol) Clone(args []Node) Node { return &Symbol{Base: s.Base, Rep: s.Rep, Sort: s.Sort} }
func (s *Symbol) String() string         { return s.Rep }
func (s *Symbol) Relname() string        { return s.Rep }
func (s *Symbol) Canon() Canonical {
	return Canonical(fmt.Sprintf("(symbol%v rep:%q sort:%v)", s.Base.canonFields(), s.Rep, nodeCanon(s.Sort)))
}

// Atom is an n-ary relation/predicate applied to terms.
type Atom struct {
	Base
	Rep   string // relation name
	Terms []Node // arguments
	ASort Node   // optional sort
}

func (cfg *AstConfig) NewAtom(rep string, terms ...Node) *Atom {
	a := &Atom{Rep: rep, Terms: terms}
	a.Cfg = cfg
	return a
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

// Repr returns a sort-qualified string, using Repr() on children when available.
// Matches Python ast.Atom.__repr__ which calls str() on args — and for
// ast.Variable, str() falls to __repr__ which includes sort qualifiers.
func (a *Atom) Repr() string {
	if IsEquals(a.Rep) && len(a.Terms) == 2 {
		return reprNode(a.Terms[0]) + " = " + reprNode(a.Terms[1])
	}
	if len(a.Terms) == 0 {
		return a.Rep
	}
	parts := make([]string, len(a.Terms))
	for i, t := range a.Terms {
		parts[i] = reprNode(t)
	}
	return a.Rep + "(" + strings.Join(parts, ", ") + ")"
}

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
func (a *Atom) Canon() Canonical {
	if a == nil {
		return Canonical("nil")
	}
	return Canonical(fmt.Sprintf("(atom%v rep:%q terms:%v aSort:%v)", a.Base.canonFields(), a.Rep, SliceCanon(a.Terms), sortCanon(a.ASort)))
}

// App is a function application (term level).
type App struct {
	Base
	Rep   Node   // function symbol (usually *Symbol)
	Terms []Node // arguments
	ASort Node   // optional sort annotation
}

func (cfg *AstConfig) NewApp(rep Node, terms ...Node) *App {
	a := &App{Rep: rep, Terms: terms}
	a.Cfg = cfg
	return a
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
		ns := &Symbol{Rep: s + sym.Rep, Sort: sym.Sort}
		ns.Cfg = a.Cfg
		c.Rep = ns
	}
	return c
}

func (a *App) DropPrefix(s string) *App {
	c := a.Clone(a.Terms).(*App)
	if sym, ok := c.Rep.(*Symbol); ok && strings.HasPrefix(sym.Rep, s) {
		ns := &Symbol{Rep: sym.Rep[len(s):], Sort: sym.Sort}
		ns.Cfg = a.Cfg
		c.Rep = ns
	}
	c.ASort = a.ASort
	return c
}

func (a *App) Rename(s string) *App {
	c := a.Clone(a.Terms).(*App)
	if sym, ok := c.Rep.(*Symbol); ok {
		ns := &Symbol{Rep: s, Sort: sym.Sort}
		ns.Cfg = a.Cfg
		c.Rep = ns
	}
	c.ASort = a.ASort
	return c
}
func (a *App) Canon() Canonical {
	// Python's App.rep is a string, so node_canon(self.rep) gives just
	// the bare string. Go's App.Rep is a Node (*Symbol). Extract the
	// bare string to match Python's canon output.
	repCanon := nodeCanon(a.Rep)
	if sym, ok := a.Rep.(*Symbol); ok {
		repCanon = Canonical(sym.Rep)
	}
	return Canonical(fmt.Sprintf("(app%v rep:%v terms:%v aSort:%v)", a.Base.canonFields(), repCanon, SliceCanon(a.Terms), sortCanon(a.ASort)))
}

// Variable represents a sorted variable in the AST.
type Variable struct {
	Base
	Rep   string
	VSort string // the sort
}

func (cfg *AstConfig) NewVariable(rep string, sort string) *Variable {
	v := &Variable{Rep: rep, VSort: sort}
	v.Cfg = cfg
	return v
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
	a := &App{Rep: &Symbol{Rep: prefix + v.Rep}}
	a.Cfg = v.Cfg
	a.ASort = &Symbol{Rep: v.VSort}
	return a
}

// ToConstAtom creates an Atom with the given prefix
// prepended to the atom name,
// copying the sort. Used for prm: prefix substitution.
func ToConstAtom(a *Atom, prefix string) *Atom {
	res := &Atom{Rep: prefix + a.Rep, Terms: a.Terms}
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
		newRep.Cfg = a.Cfg
	default:
		panicf("how to handle %T ?", a.Rep)
	}
	res = &App{Rep: newRep, Terms: a.Terms}
	res.Base = a.Base
	res.ASort = a.ASort
	return
}

// Resort creates a new Variable with a different sort.
// Matches Python Variable.resort (ivy_ast.py:402-408) which calls lineno_add_ref.
func (v *Variable) Resort(sort string) *Variable {
	nv := &Variable{Base: v.Base, Rep: v.Rep, VSort: sort}
	nv.SetLineno(safeLinenoAddRef(v, v.GetLineno()))
	return nv
}
func (v *Variable) Canon() Canonical {
	var vsort string
	switch v.VSort {
	case "":
		vsort = "nil"
	case "this":
		vsort = string((&This{}).Canon())
	default:
		vsort = v.VSort
	}
	return Canonical(fmt.Sprintf("(variable%v rep:%q vSort:%v)", v.Base.canonFields(), v.Rep, vsort))
}

// Old wraps a term with the temporal "old" operator.
type Old struct {
	Base
	Term Node
}

func (cfg *AstConfig) NewOld(term Node) *Old {
	o := &Old{Term: term}
	o.Cfg = cfg
	return o
}

func (o *Old) Args() []Node           { return []Node{o.Term} }
func (o *Old) Clone(args []Node) Node { return &Old{Base: o.Base, Term: args[0]} }
func (o *Old) String() string         { return "old " + fmt.Sprint(o.Term) }
func (o *Old) Canon() Canonical {
	return Canonical(fmt.Sprintf("(old%v term:%v)", o.Base.canonFields(), nodeCanon(o.Term)))
}

// This represents a self-reference.
type This struct {
	Base
}

func (cfg *AstConfig) NewThis() *This {
	t := &This{}
	t.Cfg = cfg
	return t
}

func (t *This) Args() []Node           { return nil }
func (t *This) Clone(args []Node) Node { return &This{Base: t.Base} }
func (t *This) String() string         { return "this" }
func (t *This) Relname() string        { return "this" }
func (t *This) Canon() Canonical {
	return Canonical(fmt.Sprintf("(this%v)", t.Base.canonFields()))
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
func (m *MethodCall) Canon() Canonical {
	return Canonical(fmt.Sprintf("(methodCall%v obj:%v method:%v)", m.Base.canonFields(), nodeCanon(m.Obj), nodeCanon(m.Method)))
}

func (cfg *AstConfig) NewMethodCall(obj, method Node) *MethodCall {
	m := &MethodCall{Obj: obj, Method: method}
	m.Cfg = cfg
	return m
}

// Literal is a positive or negative atomic formula.
type Literal struct {
	Base
	Polarity int // 1 = positive, 0 = negative
	Atom     Node
}

func (cfg *AstConfig) NewLiteral(polarity int, atom Node) *Literal {
	l := &Literal{Polarity: polarity, Atom: atom}
	l.Cfg = cfg
	return l
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
func (l *Literal) Canon() Canonical {
	return Canonical(fmt.Sprintf("(literal%v polarity:%d atom:%v)", l.Base.canonFields(), l.Polarity, nodeCanon(l.Atom)))
}

// Dot represents field access (a.b).
type Dot struct {
	Base
	Left  Node
	Right Node
}

func (cfg *AstConfig) NewDot(left, right Node) *Dot {
	d := &Dot{Left: left, Right: right}
	d.Cfg = cfg
	return d
}

func (d *Dot) Args() []Node { return []Node{d.Left, d.Right} }
func (d *Dot) Clone(args []Node) Node {
	return &Dot{Base: d.Base, Left: args[0], Right: args[1]}
}
func (d *Dot) String() string { return fmt.Sprint(d.Left) + "." + fmt.Sprint(d.Right) }
func (d *Dot) Canon() Canonical {
	return Canonical(fmt.Sprintf("(dot%v left:%v right:%v)", d.Base.canonFields(), nodeCanon(d.Left), nodeCanon(d.Right)))
}

// Bracket represents subscript access (a[b]).
type Bracket struct {
	Base
	Left  Node
	Right Node
}

func (cfg *AstConfig) NewBracket(left, right Node) *Bracket {
	b := &Bracket{Left: left, Right: right}
	b.Cfg = cfg
	return b
}

func (b *Bracket) Args() []Node { return []Node{b.Left, b.Right} }
func (b *Bracket) Clone(args []Node) Node {
	return &Bracket{Base: b.Base, Left: args[0], Right: args[1]}
}
func (b *Bracket) String() string { return fmt.Sprint(b.Left) + "[" + fmt.Sprint(b.Right) + "]" }
func (b *Bracket) Canon() Canonical {
	return Canonical(fmt.Sprintf("(bracket%v left:%v right:%v)", b.Base.canonFields(), nodeCanon(b.Left), nodeCanon(b.Right)))
}

// Tuple wraps a list of nodes in parentheses.
type Tuple struct {
	Base
	Elems []Node
}

func (cfg *AstConfig) NewTuple(elems ...Node) *Tuple {
	t := &Tuple{Elems: elems}
	t.Cfg = cfg
	return t
}

func (t *Tuple) Args() []Node           { return t.Elems }
func (t *Tuple) Clone(args []Node) Node { return &Tuple{Base: t.Base, Elems: args} }
func (t *Tuple) String() string {
	parts := make([]string, len(t.Elems))
	for i, e := range t.Elems {
		parts[i] = fmt.Sprint(e)
	}
	return "(" + strings.Join(parts, ",") + ")"
}
func (t *Tuple) Canon() Canonical {
	return Canonical(fmt.Sprintf("(tuple%v elems:%v)", t.Base.canonFields(), SliceCanon(t.Elems)))
}

// Some represents "some X. phi" existential choice.
type Some struct {
	Base
	Params []Node // bound variables (all but last)
	Fmla   Node   // formula (last arg)
}

func (cfg *AstConfig) NewSome(params []Node, fmla Node) *Some {
	s := &Some{Params: params, Fmla: fmla}
	s.Cfg = cfg
	return s
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
func (s *Some) Canon() Canonical {
	return Canonical(fmt.Sprintf("(some%v params:%v fmla:%v)", s.Base.canonFields(), SliceCanon(s.Params), nodeCanon(s.Fmla)))
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
func (s *SomeMin) Canon() Canonical {
	return Canonical(fmt.Sprintf("(someMin%v params:%v fmla:%v index:%v)", s.Base.canonFields(), SliceCanon(s.Params), nodeCanon(s.Fmla), nodeCanon(s.Index)))
}

func (cfg *AstConfig) NewSomeMin(params []Node, fmla, index Node) *SomeMin {
	s := &SomeMin{Params: params, Fmla: fmla, Index: index}
	s.Cfg = cfg
	return s
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
func (s *SomeMax) Canon() Canonical {
	return Canonical(fmt.Sprintf("(someMax%v params:%v fmla:%v index:%v)", s.Base.canonFields(), SliceCanon(s.Params), nodeCanon(s.Fmla), nodeCanon(s.Index)))
}

func (cfg *AstConfig) NewSomeMax(params []Node, fmla, index Node) *SomeMax {
	s := &SomeMax{Params: params, Fmla: fmla, Index: index}
	s.Cfg = cfg
	return s
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
func (s *SomeExpr) Canon() Canonical {
	return Canonical(fmt.Sprintf("(someExpr%v param:%v fmla:%v ifValue:%v elseVal:%v)", s.Base.canonFields(), nodeCanon(s.Param), nodeCanon(s.Fmla), nodeCanon(s.IfValue), nodeCanon(s.ElseVal)))
}

func (cfg *AstConfig) NewSomeExpr(param, fmla Node) *SomeExpr {
	s := &SomeExpr{Param: param, Fmla: fmla}
	s.Cfg = cfg
	return s
}

// KeyArg wraps an App with a ^ prefix (key argument).
type KeyArg struct {
	*App
}

func (cfg *AstConfig) NewKeyArg(app *App) *KeyArg {
	return &KeyArg{App: app}
}

func (k *KeyArg) String() string { return "^" + k.App.String() }
func (k *KeyArg) Clone(args []Node) Node {
	return &KeyArg{App: k.App.Clone(args).(*App)}
}
func (k *KeyArg) Canon() Canonical {
	return Canonical(fmt.Sprintf("(keyArg app:%v)", k.App.Canon()))
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
func (d *DebugItem) Canon() Canonical {
	return Canonical(fmt.Sprintf("(debugItem%v name:%v value:%v)", d.Base.canonFields(), nodeCanon(d.Name), nodeCanon(d.Value)))
}

func (cfg *AstConfig) NewDebugItem(name, value Node) *DebugItem {
	d := &DebugItem{Name: name, Value: value}
	d.Cfg = cfg
	return d
}

// ThunkAction represents "thunk [label] name(args) : type := body".
// Python: ThunkAction(Action) from ivy_actions.py
// args = [label, action_atom, type, body]
type ThunkAction struct {
	Base
	Label        Node // the label (Atom)
	Action       Node // the action name with args (Atom)
	Sort         Node // the sort/type
	Body         Node // the body sequence
	Continuation Node // optional: set by lower_var_stmts when appending scoped continuation
}

func (cfg *AstConfig) NewThunkAction(label, action, sort, body Node) *ThunkAction {
	t := &ThunkAction{Label: label, Action: action, Sort: sort, Body: body}
	t.Cfg = cfg
	return t
}

func (t *ThunkAction) Args() []Node {
	args := []Node{t.Label, t.Action, t.Sort, t.Body}
	if t.Continuation != nil {
		args = append(args, t.Continuation)
	}
	return args
}
func (t *ThunkAction) Clone(args []Node) Node {
	c := &ThunkAction{Base: t.Base, Label: args[0], Action: args[1], Sort: args[2], Body: args[3]}
	if len(args) > 4 {
		c.Continuation = args[4]
	}
	return c
}
func (t *ThunkAction) String() string {
	s := "thunk [" + fmt.Sprint(t.Label) + "] " + fmt.Sprint(t.Action) + " : " + fmt.Sprint(t.Sort) + " := " + fmt.Sprint(t.Body)
	if t.Continuation != nil {
		s += " ; " + fmt.Sprint(t.Continuation)
	}
	return s
}
func (t *ThunkAction) Canon() Canonical {
	cont := "nil"
	if t.Continuation != nil {
		cont = string(t.Continuation.Canon())
	}
	return Canonical(fmt.Sprintf("(thunkAction%v label:%v action:%v sort:%v body:%v continuation:%v)", t.Base.canonFields(), nodeCanon(t.Label), nodeCanon(t.Action), nodeCanon(t.Sort), nodeCanon(t.Body), cont))
}

// TemporalModels represents M |= phi.
// CrashAction represents "action name = *" (havoc/crash).
// Python: CrashAction(Action) from ivy_actions.py
type CrashAction struct {
	Base
	DeclArgs []Node
}

func (cfg *AstConfig) NewCrashAction(args ...Node) *CrashAction {
	c := &CrashAction{DeclArgs: args}
	c.Cfg = cfg
	return c
}

func (c *CrashAction) Args() []Node { return c.DeclArgs }
func (c *CrashAction) Clone(args []Node) Node {
	return &CrashAction{Base: c.Base, DeclArgs: args}
}
func (c *CrashAction) String() string {
	if len(c.DeclArgs) > 0 {
		return "crash " + fmt.Sprint(c.DeclArgs[0])
	}
	return "crash"
}
func (c *CrashAction) Canon() Canonical {
	return Canonical(fmt.Sprintf("(crashAction%v declArgs:%v)", c.Base.canonFields(), SliceCanon(c.DeclArgs)))
}

// ChoiceAction represents "if * { ... } else { ... }" non-deterministic choice.
// Python: class ChoiceAction(Action) from ivy_actions.py:805.
// Each instance gets a unique_id for determinization.
type ChoiceAction struct {
	Base
	Branches []Node
	UniqueID int64
}

func (cfg *AstConfig) NewChoiceAction(branches ...Node) *ChoiceAction {
	id := cfg.IuCfg.ChoiceActionCtr
	cfg.IuCfg.ChoiceActionCtr++
	xtracer.Trace("ChoiceAction.__init__ uniqueID=%d counter=%d caller=ChoiceAction", id, cfg.IuCfg.ChoiceActionCtr)
	ca := &ChoiceAction{Branches: branches, UniqueID: id}
	ca.Cfg = cfg
	return ca
}

func (c *ChoiceAction) Args() []Node { return c.Branches }
func (c *ChoiceAction) Clone(args []Node) Node {
	cfg := c.Cfg
	if cfg == nil {
		panic("ast: Clone called on node with nil AstConfig — node was not created via cfg.NewFoo()")
	}
	id := cfg.IuCfg.ChoiceActionCtr
	cfg.IuCfg.ChoiceActionCtr++
	xtracer.Trace("ChoiceAction.__init__ uniqueID=%d counter=%d caller=ChoiceAction", id, cfg.IuCfg.ChoiceActionCtr)
	return &ChoiceAction{Base: c.Base, Branches: args, UniqueID: id}
}
func (c *ChoiceAction) String() string { return "choice" }
func (c *ChoiceAction) Canon() Canonical {
	return Canonical(fmt.Sprintf("(choiceAction%v branches:%v uniqueID:%d)", c.Base.canonFields(), SliceCanon(c.Branches), c.UniqueID))
}

// EnvAction represents an environment action (non-deterministic choice of public actions).
// Python: class EnvAction(ChoiceAction) from ivy_actions.py:864.
// Compiler-created: ivy_isolate.py:1735 during isolate extraction (all Ivy versions).
// Like ChoiceAction but hides child formal params/returns.
type EnvAction struct {
	Base
	Branches []Node
	UniqueID int64
}

func (cfg *AstConfig) NewEnvAction(branches ...Node) *EnvAction {
	id := cfg.IuCfg.ChoiceActionCtr
	cfg.IuCfg.ChoiceActionCtr++
	xtracer.Trace("ChoiceAction.__init__ uniqueID=%d counter=%d caller=EnvAction", id, cfg.IuCfg.ChoiceActionCtr)
	ea := &EnvAction{Branches: branches, UniqueID: id}
	ea.Cfg = cfg
	return ea
}

func (a *EnvAction) Args() []Node { return a.Branches }
func (a *EnvAction) Clone(args []Node) Node {
	cfg := a.Cfg
	if cfg == nil {
		panic("ast: Clone called on EnvAction with nil AstConfig")
	}
	id := cfg.IuCfg.ChoiceActionCtr
	cfg.IuCfg.ChoiceActionCtr++
	xtracer.Trace("ChoiceAction.__init__ uniqueID=%d counter=%d caller=EnvAction", id, cfg.IuCfg.ChoiceActionCtr)
	return &EnvAction{Base: a.Base, Branches: args, UniqueID: id}
}
func (a *EnvAction) String() string { return "env" }
func (a *EnvAction) Canon() Canonical {
	return Canonical(fmt.Sprintf("(envAction%v branches:%v uniqueID:%d)", a.Base.canonFields(), SliceCanon(a.Branches), a.UniqueID))
}

// LetAction represents "let x = y, ... { body }".
// Python: class LetAction(Action) from ivy_actions.py:1081.
// Args are all-but-last = bindings, last = body.
type LetAction struct {
	Base
	Bindings []Node // equation bindings (Atom("=", lhs, rhs) nodes)
	Body     Node   // the body action (last arg)
}

func (cfg *AstConfig) NewLetAction(args ...Node) *LetAction {
	var l *LetAction
	if len(args) == 0 {
		l = &LetAction{}
	} else {
		l = &LetAction{Bindings: args[:len(args)-1], Body: args[len(args)-1]}
	}
	l.Cfg = cfg
	return l
}

func (l *LetAction) Args() []Node {
	if l.Body == nil {
		return l.Bindings
	}
	return append(append([]Node{}, l.Bindings...), l.Body)
}
func (l *LetAction) Clone(args []Node) Node {
	if len(args) == 0 {
		return &LetAction{Base: l.Base}
	}
	return &LetAction{Base: l.Base, Bindings: args[:len(args)-1], Body: args[len(args)-1]}
}
func (l *LetAction) String() string { return "let" }
func (l *LetAction) Canon() Canonical {
	return Canonical(fmt.Sprintf("(letAction%v bindings:%v body:%v)", l.Base.canonFields(), SliceCanon(l.Bindings), nodeCanon(l.Body)))
}

// Ranking wraps a formula for DECREASES clauses.
// Python: class Ranking(Action) from ivy_actions.py:953.
type Ranking struct {
	Base
	Fmla Node
}

func (cfg *AstConfig) NewRanking(fmla Node) *Ranking {
	r := &Ranking{Fmla: fmla}
	r.Cfg = cfg
	return r
}

func (r *Ranking) Args() []Node           { return []Node{r.Fmla} }
func (r *Ranking) Clone(args []Node) Node { return &Ranking{Base: r.Base, Fmla: args[0]} }
func (r *Ranking) String() string         { return "decreases" }
func (r *Ranking) Canon() Canonical {
	return Canonical(fmt.Sprintf("(ranking%v fmla:%v)", r.Base.canonFields(), nodeCanon(r.Fmla)))
}

// AssertAction asserts a formula (can fail verification).
// Python: class AssertAction(Action) from ivy_actions.py:332.
type AssertAction struct {
	Base
	Elems []Node
}

func (cfg *AstConfig) NewAssertAction(args ...Node) *AssertAction {
	a := &AssertAction{Elems: args}
	a.Cfg = cfg
	return a
}

func (a *AssertAction) Args() []Node           { return a.Elems }
func (a *AssertAction) Clone(args []Node) Node { return &AssertAction{Base: a.Base, Elems: args} }
func (a *AssertAction) String() string         { return "assert" }
func (a *AssertAction) Canon() Canonical {
	return Canonical(fmt.Sprintf("(assertAction%v elems:%v)", a.Base.canonFields(), SliceCanon(a.Elems)))
}

// AssumeAction assumes a formula holds.
// Python: class AssumeAction(Action) from ivy_actions.py:309.
type AssumeAction struct {
	Base
	Elems []Node
}

func (cfg *AstConfig) NewAssumeAction(args ...Node) *AssumeAction {
	a := &AssumeAction{Elems: args}
	a.Cfg = cfg
	return a
}

func (a *AssumeAction) Args() []Node           { return a.Elems }
func (a *AssumeAction) Clone(args []Node) Node { return &AssumeAction{Base: a.Base, Elems: args} }
func (a *AssumeAction) String() string         { return "assume" }
func (a *AssumeAction) Canon() Canonical {
	return Canonical(fmt.Sprintf("(assumeAction%v elems:%v)", a.Base.canonFields(), SliceCanon(a.Elems)))
}

// EnsuresAction is like assert but for postconditions.
// Python: class EnsuresAction(Action) from ivy_actions.py.
type EnsuresAction struct {
	Base
	Elems []Node
}

func (cfg *AstConfig) NewEnsuresAction(args ...Node) *EnsuresAction {
	a := &EnsuresAction{Elems: args}
	a.Cfg = cfg
	return a
}

func (a *EnsuresAction) Args() []Node { return a.Elems }
func (a *EnsuresAction) Clone(args []Node) Node {
	return &EnsuresAction{Base: a.Base, Elems: args}
}
func (a *EnsuresAction) String() string { return "ensures" }
func (a *EnsuresAction) Canon() Canonical {
	return Canonical(fmt.Sprintf("(ensuresAction%v elems:%v)", a.Base.canonFields(), SliceCanon(a.Elems)))
}

// RequiresAction is like assert but for preconditions.
// Python: class RequiresAction(Action) from ivy_actions.py.
type RequiresAction struct {
	Base
	Elems []Node
}

func (cfg *AstConfig) NewRequiresAction(args ...Node) *RequiresAction {
	a := &RequiresAction{Elems: args}
	a.Cfg = cfg
	return a
}

func (a *RequiresAction) Args() []Node { return a.Elems }
func (a *RequiresAction) Clone(args []Node) Node {
	return &RequiresAction{Base: a.Base, Elems: args}
}
func (a *RequiresAction) String() string { return "requires" }
func (a *RequiresAction) Canon() Canonical {
	return Canonical(fmt.Sprintf("(requiresAction%v elems:%v)", a.Base.canonFields(), SliceCanon(a.Elems)))
}

// SubgoalAction represents a proof subgoal assertion.
// Python: class SubgoalAction(AssertAction) from ivy_actions.py:394.
// Has a custom clone that preserves the Kind field (mirrors Python's self.kind).
type SubgoalAction struct {
	Base
	Elems []Node
	Kind  string // preserved across clone, mirrors Python's self.kind
}

func (cfg *AstConfig) NewSubgoalAction(args ...Node) *SubgoalAction {
	a := &SubgoalAction{Elems: args}
	a.Cfg = cfg
	return a
}

func (a *SubgoalAction) Args() []Node { return a.Elems }
func (a *SubgoalAction) Clone(args []Node) Node {
	return &SubgoalAction{Base: a.Base, Elems: args, Kind: a.Kind}
}
func (a *SubgoalAction) String() string { return "subgoal" }
func (a *SubgoalAction) Canon() Canonical {
	return Canonical(fmt.Sprintf("(subgoalAction%v elems:%v)", a.Base.canonFields(), SliceCanon(a.Elems)))
}

// AssignFieldAction assigns to a destructor field: obj.field := value.
// Python: class AssignFieldAction(Action) from ivy_actions.py:710.
// Parser-created: Ivy version <= 1.2 only (p_action_field_assign_term).
// Has sort_infer_root = True in Python.
type AssignFieldAction struct {
	Base
	Elems []Node // [obj, field_name, value] — 3 args
}

func (cfg *AstConfig) NewAssignFieldAction(args ...Node) *AssignFieldAction {
	a := &AssignFieldAction{Elems: args}
	a.Cfg = cfg
	return a
}

func (a *AssignFieldAction) Args() []Node { return a.Elems }
func (a *AssignFieldAction) Clone(args []Node) Node {
	return &AssignFieldAction{Base: a.Base, Elems: args}
}
func (a *AssignFieldAction) String() string { return "assign_field" }
func (a *AssignFieldAction) Canon() Canonical {
	return Canonical(fmt.Sprintf("(assignFieldAction%v elems:%v)", a.Base.canonFields(), SliceCanon(a.Elems)))
}

// NullFieldAction sets a destructor field to null/default: obj.field := null.
// Python: class NullFieldAction(Action) from ivy_actions.py:723.
// Parser-created: Ivy version <= 1.2 only (p_action_field_assign_null, p_action_field_assign_false).
// Has sort_infer_root = True in Python.
type NullFieldAction struct {
	Base
	Elems []Node // [obj, field_name] — 2 args
}

func (cfg *AstConfig) NewNullFieldAction(args ...Node) *NullFieldAction {
	a := &NullFieldAction{Elems: args}
	a.Cfg = cfg
	return a
}

func (a *NullFieldAction) Args() []Node { return a.Elems }
func (a *NullFieldAction) Clone(args []Node) Node {
	return &NullFieldAction{Base: a.Base, Elems: args}
}
func (a *NullFieldAction) String() string { return "null_field" }
func (a *NullFieldAction) Canon() Canonical {
	return Canonical(fmt.Sprintf("(nullFieldAction%v elems:%v)", a.Base.canonFields(), SliceCanon(a.Elems)))
}

// CopyFieldAction copies a destructor field between objects: dst.field := src.field.
// Python: class CopyFieldAction(Action) from ivy_actions.py:736.
// Parser-created: Ivy version <= 1.2 only (p_action_field_assign_field).
// Has sort_infer_root = True in Python.
type CopyFieldAction struct {
	Base
	Elems []Node // [dst_obj, dst_field, src_obj, src_field] — 4 args
}

func (cfg *AstConfig) NewCopyFieldAction(args ...Node) *CopyFieldAction {
	a := &CopyFieldAction{Elems: args}
	a.Cfg = cfg
	return a
}

func (a *CopyFieldAction) Args() []Node { return a.Elems }
func (a *CopyFieldAction) Clone(args []Node) Node {
	return &CopyFieldAction{Base: a.Base, Elems: args}
}
func (a *CopyFieldAction) String() string { return "assign_field" } // Python CopyFieldAction.name() returns "assign_field"
func (a *CopyFieldAction) Canon() Canonical {
	return Canonical(fmt.Sprintf("(copyFieldAction%v elems:%v)", a.Base.canonFields(), SliceCanon(a.Elems)))
}

// BindOldsAction binds old values before executing an inner action body.
// Python: class BindOldsAction(Action) from ivy_actions.py:1231.
// Compiler-created: ivy_actions.py:1309 in CallAction.int_update() (all Ivy versions).
type BindOldsAction struct {
	Base
	Elems []Node // [inner_action] — 1 arg
}

func (cfg *AstConfig) NewBindOldsAction(args ...Node) *BindOldsAction {
	a := &BindOldsAction{Elems: args}
	a.Cfg = cfg
	return a
}

func (a *BindOldsAction) Args() []Node { return a.Elems }
func (a *BindOldsAction) Clone(args []Node) Node {
	return &BindOldsAction{Base: a.Base, Elems: args}
}
func (a *BindOldsAction) String() string { return "bindolds" }
func (a *BindOldsAction) Canon() Canonical {
	return Canonical(fmt.Sprintf("(bindOldsAction%v elems:%v)", a.Base.canonFields(), SliceCanon(a.Elems)))
}

// AssignAction represents "lhs := rhs".
// Python: class AssignAction(Action) from ivy_actions.py:469.
type AssignAction struct {
	Base
	Elems []Node
}

func (cfg *AstConfig) NewAssignAction(args ...Node) *AssignAction {
	a := &AssignAction{Elems: args}
	a.Cfg = cfg
	return a
}
func (a *AssignAction) Args() []Node           { return a.Elems }
func (a *AssignAction) Clone(args []Node) Node { return &AssignAction{Base: a.Base, Elems: args} }
func (a *AssignAction) String() string         { return "assign" }
func (a *AssignAction) Canon() Canonical {
	return Canonical(fmt.Sprintf("(assignAction%v elems:%v)", a.Base.canonFields(), SliceCanon(a.Elems)))
}

// HavocAction represents "x := *" (nondeterministic assignment).
// Python: class HavocAction(Action) from ivy_actions.py.
type HavocAction struct {
	Base
	Elems []Node
}

func (cfg *AstConfig) NewHavocAction(args ...Node) *HavocAction {
	a := &HavocAction{Elems: args}
	a.Cfg = cfg
	return a
}
func (a *HavocAction) Args() []Node           { return a.Elems }
func (a *HavocAction) Clone(args []Node) Node { return &HavocAction{Base: a.Base, Elems: args} }
func (a *HavocAction) String() string         { return "havoc" }
func (a *HavocAction) Canon() Canonical {
	return Canonical(fmt.Sprintf("(havocAction%v elems:%v)", a.Base.canonFields(), SliceCanon(a.Elems)))
}

// VarAction represents local variable declarations.
// Python: class VarAction(Action) from ivy_actions.py.
type VarAction struct {
	Base
	Elems []Node
}

func (cfg *AstConfig) NewVarAction(args ...Node) *VarAction {
	a := &VarAction{Elems: args}
	a.Cfg = cfg
	return a
}
func (a *VarAction) Args() []Node { return a.Elems }
func (a *VarAction) Clone(args []Node) Node {
	return &VarAction{Base: a.Base, Elems: args}
}
func (a *VarAction) String() string { return "var" }
func (a *VarAction) Canon() Canonical {
	return Canonical(fmt.Sprintf("(varAction%v elems:%v)", a.Base.canonFields(), SliceCanon(a.Elems)))
}

// SetAction represents "set" commands.
// Python: class SetAction(Action) from ivy_actions.py.
type SetAction struct {
	Base
	Elems []Node
}

func (cfg *AstConfig) NewSetAction(args ...Node) *SetAction {
	a := &SetAction{Elems: args}
	a.Cfg = cfg
	return a
}
func (a *SetAction) Args() []Node { return a.Elems }
func (a *SetAction) Clone(args []Node) Node {
	return &SetAction{Base: a.Base, Elems: args}
}
func (a *SetAction) String() string { return "set" }
func (a *SetAction) Canon() Canonical {
	return Canonical(fmt.Sprintf("(setAction%v elems:%v)", a.Base.canonFields(), SliceCanon(a.Elems)))
}

// InstantiateAction represents "instantiate" commands.
// Python: class InstantiateAction(Action) from ivy_actions.py.
type InstantiateAction struct {
	Base
	Elems []Node
}

func (cfg *AstConfig) NewInstantiateAction(args ...Node) *InstantiateAction {
	a := &InstantiateAction{Elems: args}
	a.Cfg = cfg
	return a
}
func (a *InstantiateAction) Args() []Node { return a.Elems }
func (a *InstantiateAction) Clone(args []Node) Node {
	return &InstantiateAction{Base: a.Base, Elems: args}
}
func (a *InstantiateAction) String() string { return "instantiate" }
func (a *InstantiateAction) Canon() Canonical {
	return Canonical(fmt.Sprintf("(instantiateAction%v elems:%v)", a.Base.canonFields(), SliceCanon(a.Elems)))
}

// DebugAction represents "debug" commands.
// Python: class DebugAction(Action) from ivy_actions.py.
type DebugAction struct {
	Base
	Elems []Node
}

func (cfg *AstConfig) NewDebugAction(args ...Node) *DebugAction {
	a := &DebugAction{Elems: args}
	a.Cfg = cfg
	return a
}
func (a *DebugAction) Args() []Node           { return a.Elems }
func (a *DebugAction) Clone(args []Node) Node { return &DebugAction{Base: a.Base, Elems: args} }
func (a *DebugAction) String() string         { return "debug" }
func (a *DebugAction) Canon() Canonical {
	return Canonical(fmt.Sprintf("(debugAction%v elems:%v)", a.Base.canonFields(), SliceCanon(a.Elems)))
}

// NativeAction represents native code blocks.
// Python: class NativeAction(Action) from ivy_actions.py.
type NativeAction struct {
	Base
	Elems []Node
}

func (cfg *AstConfig) NewNativeAction(args ...Node) *NativeAction {
	a := &NativeAction{Elems: args}
	a.Cfg = cfg
	return a
}
func (a *NativeAction) Args() []Node           { return a.Elems }
func (a *NativeAction) Clone(args []Node) Node { return &NativeAction{Base: a.Base, Elems: args} }
func (a *NativeAction) String() string         { return "native" }
func (a *NativeAction) Canon() Canonical {
	return Canonical(fmt.Sprintf("(nativeAction%v elems:%v)", a.Base.canonFields(), SliceCanon(a.Elems)))
}

// WhileAction represents while loops.
// Python: class WhileAction(Action) from ivy_actions.py.
type WhileAction struct {
	Base
	Elems []Node
}

func (cfg *AstConfig) NewWhileAction(args ...Node) *WhileAction {
	a := &WhileAction{Elems: args}
	a.Cfg = cfg
	return a
}
func (a *WhileAction) Args() []Node           { return a.Elems }
func (a *WhileAction) Clone(args []Node) Node { return &WhileAction{Base: a.Base, Elems: args} }
func (a *WhileAction) String() string         { return "while" }
func (a *WhileAction) Canon() Canonical {
	return Canonical(fmt.Sprintf("(whileAction%v elems:%v)", a.Base.canonFields(), SliceCanon(a.Elems)))
}

// IfAction represents if-then-else.
// Python: class IfAction(Action) from ivy_actions.py.
type IfAction struct {
	Base
	Cond Node
	Then Node
	Else Node
}

func (cfg *AstConfig) NewIfAction(cond, then, els Node) *IfAction {
	a := &IfAction{Cond: cond, Then: then, Else: els}
	a.Cfg = cfg
	return a
}
func (a *IfAction) Args() []Node {
	if a.Else != nil {
		return []Node{a.Cond, a.Then, a.Else}
	}
	return []Node{a.Cond, a.Then}
}
func (a *IfAction) Clone(args []Node) Node {
	var els Node
	if len(args) > 2 {
		els = args[2]
	}
	return &IfAction{Base: a.Base, Cond: args[0], Then: args[1], Else: els}
}
func (a *IfAction) String() string { return "if" }
func (a *IfAction) Canon() Canonical {
	return Canonical(fmt.Sprintf("(ifAction%v cond:%v then:%v else:%v)",
		a.Base.canonFields(), nodeCanon(a.Cond), nodeCanon(a.Then), nodeCanon(a.Else)))
}

// LocalAction represents local scoping of actions.
// Python: class LocalAction(Action) from ivy_actions.py.
type LocalAction struct {
	Base
	Elems    []Node
	UniqueID int64
}

func (cfg *AstConfig) NewLocalAction(caller string, args ...Node) *LocalAction {
	id := cfg.IuCfg.LocalActionCtr
	cfg.IuCfg.LocalActionCtr++
	la := &LocalAction{Elems: args, UniqueID: id}
	la.Cfg = cfg
	xtracer.Trace("LocalAction.__init__ uniqueID=%d caller=%s", id, caller)
	return la
}
func (a *LocalAction) Args() []Node { return a.Elems }
func (a *LocalAction) Clone(args []Node) Node {
	// Python's clone calls __init__ which allocates a new unique_id.
	cfg := a.Cfg
	if cfg == nil {
		panic("ast: Clone called on node with nil AstConfig — node was not created via cfg.NewFoo()")
	}
	la := cfg.NewLocalAction("ast.LocalAction.clone", args...)
	la.Base = a.Base
	return la
}
func (a *LocalAction) String() string { return "local" }
func (a *LocalAction) Canon() Canonical {
	return Canonical(fmt.Sprintf("(localAction%v elems:%v uniqueID:%d)",
		a.Base.canonFields(), SliceCanon(a.Elems), a.UniqueID))
}

// SomeAssignAction represents "if some x. P { body }".
// Python: SomeMinEqualAction or similar from ivy_actions.py.
type SomeAssignAction struct {
	Base
	Elems []Node
}

func (cfg *AstConfig) NewSomeAssignAction(args ...Node) *SomeAssignAction {
	a := &SomeAssignAction{Elems: args}
	a.Cfg = cfg
	return a
}
func (a *SomeAssignAction) Args() []Node { return a.Elems }
func (a *SomeAssignAction) Clone(args []Node) Node {
	return &SomeAssignAction{Base: a.Base, Elems: args}
}
func (a *SomeAssignAction) String() string { return "some_assign" }
func (a *SomeAssignAction) Canon() Canonical {
	return Canonical(fmt.Sprintf("(someAssignAction%v elems:%v)", a.Base.canonFields(), SliceCanon(a.Elems)))
}

// CallAction inlines a named state or action.
// Python: class CallAction(Action) from ivy_actions.py:1182.
// args[0] is the callee atom; args[1:] are actual returns.
type CallAction struct {
	Base
	Elems    []Node
	UniqueID int64
}

func (cfg *AstConfig) NewCallAction(args ...Node) *CallAction {
	id := cfg.IuCfg.CallActionCtr
	cfg.IuCfg.CallActionCtr++
	ca := &CallAction{Elems: args, UniqueID: id}
	ca.Cfg = cfg
	xtracer.Trace("CallAction.__init__ uniqueID=%d counter=%d", id, cfg.IuCfg.CallActionCtr) //seen
	return ca
}

func (c *CallAction) Args() []Node { return c.Elems }
func (c *CallAction) Clone(args []Node) Node {
	// Python's clone calls __init__ which allocates a new unique_id.
	cfg := c.Cfg
	if cfg == nil {
		panic("ast: Clone called on node with nil AstConfig — node was not created via cfg.NewFoo()")
	}
	ca := cfg.NewCallAction(args...)
	ca.Base = c.Base
	return ca
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
func (c *CallAction) Canon() Canonical {
	return Canonical(fmt.Sprintf("(callAction%v elems:%v uniqueID:%d)", c.Base.canonFields(), SliceCanon(c.Elems), c.UniqueID))
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
// The new LALR parser (in parser/) uses ast.Sequence to be faithful to the
// original Python grammar, which is important because downstream code
// (e.g. int_update, compose_updates) uses isinstance(x, Sequence) type
// checks that distinguish action sequences from logical conjunctions.
type Sequence struct {
	Base
	Stmts []Node
}

func (cfg *AstConfig) NewSequence(stmts ...Node) *Sequence {
	s := &Sequence{Stmts: stmts}
	s.Cfg = cfg
	return s
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
func (s *Sequence) Canon() Canonical {
	return Canonical(fmt.Sprintf("(sequence%v stmts:%v)", s.Base.canonFields(), SliceCanon(s.Stmts)))
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
func (t *TemporalModels) Canon() Canonical {
	var b strings.Builder
	b.WriteString("(temporalModels")
	b.WriteString(t.Base.canonFields())
	b.WriteString(" model:")
	b.WriteString(string(nodeCanon(t.Model)))
	b.WriteString(" fmla:")
	b.WriteString(string(nodeCanon(t.Fmla)))
	b.WriteByte(')')
	return Canonical(b.String())
}

func (cfg *AstConfig) NewTemporalModels(model, fmla Node) *TemporalModels {
	t := &TemporalModels{Model: model, Fmla: fmla}
	t.Cfg = cfg
	return t
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

// Equal returns true if other is a *LogicLiteral with the same Polarity and Atom.
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
	res := &Atom{Rep: fmt.Sprint(a.Rep), Terms: a.Terms}
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
var Equals = &Symbol{Rep: "=", Sort: &RelationSort{Dom: []Node{nil, nil}}}

// IsEquals checks if a name is the equality symbol.
func IsEquals(name string) bool {
	return name == "="
}

// --- Labeled formula counter ---

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
func AstIsTrue(n Node) bool {
	if a, ok := n.(*And); ok {
		return len(a.Terms) == 0
	}
	return false
}

// IsFalse checks if an AST node represents false (empty Or).
func AstIsFalse(n Node) bool {
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

func (cfg *AstConfig) NewCompiledNode(node interface{}) *CompiledNode {
	c := &CompiledNode{Node: node}
	c.Cfg = cfg
	return c
}

func (c *CompiledNode) Args() []Node           { return nil }
func (c *CompiledNode) Clone(args []Node) Node { return &CompiledNode{Base: c.Base, Node: c.Node} }
func (c *CompiledNode) String() string         { return fmt.Sprint(c.Node) }
func (c *CompiledNode) Canon() Canonical {
	if cz, ok := c.Node.(Canonizer); ok {
		return cz.Canon()
	}
	return Canonical(fmt.Sprintf("(compiledNode%v)", c.Base.canonFields()))
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

func astExtractSortRep(n Node) string {
	if s, ok := n.(*Symbol); ok {
		return s.Rep
	}
	return fmt.Sprint(n)
}
