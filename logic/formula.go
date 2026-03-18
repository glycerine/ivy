package logic

import (
	"fmt"
	"sort"
	"strings"
)

// True and False are the logical constants (empty And / empty Or).
var (
	True  Node = &And{}
	False Node = &Or{}
)

// --- Eq ---

type Eq struct {
	T1, T2 Node
}

func NewEq(t1, t2 Node) (*Eq, error) {
	s1, s2 := t1.NodeSort(), t2.NodeSort()
	_, t1Top := s1.(*TopSort)
	_, t2Top := s2.(*TopSort)
	if t1Top || t2Top {
		// pass
	} else if !s1.Equal(s2) {
		return nil, &SortError{Msg: fmt.Sprintf("Cannot compare different sorts: %s:%s == %s:%s", t1, s1, t2, s2)}
	} else if !FirstOrderSort(s1) {
		return nil, &SortError{Msg: fmt.Sprintf("Cannot compare high order sorts: %s == %s", t1, t2)}
	}
	return &Eq{T1: t1, T2: t2}, nil
}

func (e *Eq) NodeSort() Sort   { return Boolean }
func (e *Eq) Children() []Node { return []Node{e.T1, e.T2} }
func (e *Eq) String() string   { return fmt.Sprintf("(%s == %s)", e.T1, e.T2) }
func (e *Eq) Equal(n Node) bool {
	if o, ok := n.(*Eq); ok {
		return e.T1.Equal(o.T1) && e.T2.Equal(o.T2)
	}
	return false
}

// --- Ite ---

type Ite struct {
	ISort Sort
	Cond  Node
	Then  Node
	Else  Node
}

func NewIte(cond, then_, else_ Node) (*Ite, error) {
	if !IsBooleanOrTop(cond.NodeSort()) {
		return nil, &SortError{Msg: fmt.Sprintf("Ite condition must be Boolean: %s", cond)}
	}
	s1, s2 := then_.NodeSort(), else_.NodeSort()
	_, t1Top := s1.(*TopSort)
	_, t2Top := s2.(*TopSort)
	if !t1Top && !t2Top && !s1.Equal(s2) {
		return nil, &SortError{Msg: fmt.Sprintf("Ite then and else terms must have same sort: %s, %s", then_, else_)}
	}
	return &Ite{ISort: then_.NodeSort(), Cond: cond, Then: then_, Else: else_}, nil
}

func (t *Ite) NodeSort() Sort   { return t.ISort }
func (t *Ite) Children() []Node { return []Node{t.Cond, t.Then, t.Else} }
func (t *Ite) String() string   { return fmt.Sprintf("Ite(%s, %s, %s)", t.Cond, t.Then, t.Else) }
func (t *Ite) Equal(n Node) bool {
	if o, ok := n.(*Ite); ok {
		return t.Cond.Equal(o.Cond) && t.Then.Equal(o.Then) && t.Else.Equal(o.Else)
	}
	return false
}

// --- Not ---

type Not struct {
	Body Node
}

func NewNot(body Node) (*Not, error) {
	if !IsBooleanOrTop(body.NodeSort()) {
		return nil, &SortError{Msg: fmt.Sprintf("Negation body must be Boolean: %s", body)}
	}
	return &Not{Body: body}, nil
}

func (n *Not) NodeSort() Sort   { return Boolean }
func (n *Not) Children() []Node { return []Node{n.Body} }
func (n *Not) String() string {
	if eq, ok := n.Body.(*Eq); ok {
		return fmt.Sprintf("(%s != %s)", eq.T1, eq.T2)
	}
	return fmt.Sprintf("~%s", n.Body)
}
func (n *Not) Equal(nd Node) bool {
	if o, ok := nd.(*Not); ok {
		return n.Body.Equal(o.Body)
	}
	return false
}

// --- Globally ---

type Globally struct {
	Environ *string
	Body    Node
}

func NewGlobally(environ *string, body Node) (*Globally, error) {
	if !IsBooleanOrTop(body.NodeSort()) {
		return nil, &SortError{Msg: fmt.Sprintf("Globally body must be Boolean: %s", body)}
	}
	return &Globally{Environ: environ, Body: body}, nil
}

func (g *Globally) NodeSort() Sort   { return Boolean }
func (g *Globally) Children() []Node { return []Node{g.Body} }
func (g *Globally) String() string {
	env := ""
	if g.Environ != nil {
		env = "[" + *g.Environ + "]"
	}
	return fmt.Sprintf("globally%s(%s)", env, g.Body)
}
func (g *Globally) Equal(n Node) bool {
	if o, ok := n.(*Globally); ok {
		return ptrStrEqual(g.Environ, o.Environ) && g.Body.Equal(o.Body)
	}
	return false
}

// --- Eventually ---

type Eventually struct {
	Environ *string
	Body    Node
}

func NewEventually(environ *string, body Node) (*Eventually, error) {
	if !IsBooleanOrTop(body.NodeSort()) {
		return nil, &SortError{Msg: fmt.Sprintf("Eventually body must be Boolean: %s", body)}
	}
	return &Eventually{Environ: environ, Body: body}, nil
}

func (e *Eventually) NodeSort() Sort   { return Boolean }
func (e *Eventually) Children() []Node { return []Node{e.Body} }
func (e *Eventually) String() string {
	env := ""
	if e.Environ != nil {
		env = "[" + *e.Environ + "]"
	}
	return fmt.Sprintf("eventually%s(%s)", env, e.Body)
}
func (e *Eventually) Equal(n Node) bool {
	if o, ok := n.(*Eventually); ok {
		return ptrStrEqual(e.Environ, o.Environ) && e.Body.Equal(o.Body)
	}
	return false
}

// --- WhenOperator ---

type WhenOperator struct {
	WSort Sort
	Name  string
	T1    Node
	T2    Node
}

func NewWhenOperator(name string, t1, t2 Node) (*WhenOperator, error) {
	if !IsBooleanOrTop(t2.NodeSort()) {
		return nil, &SortError{Msg: fmt.Sprintf("WhenOperator second argument must be Boolean: %s", t2)}
	}
	return &WhenOperator{WSort: t1.NodeSort(), Name: name, T1: t1, T2: t2}, nil
}

func (w *WhenOperator) NodeSort() Sort   { return w.WSort }
func (w *WhenOperator) Children() []Node { return []Node{w.T1, w.T2} }
func (w *WhenOperator) String() string {
	return fmt.Sprintf("WhenOperator(%s,%s,%s)", w.Name, w.T1, w.T2)
}
func (w *WhenOperator) Equal(n Node) bool {
	if o, ok := n.(*WhenOperator); ok {
		return w.Name == o.Name && w.T1.Equal(o.T1) && w.T2.Equal(o.T2)
	}
	return false
}

// --- Cond ---

type Cond struct {
	CSort Sort
	T1    Node
	T2    Node
}

// ### 1.12 `Cond` sort validation: Python has dead-code validation
//
// Python: (logic.py:287-288):
// bad_sorts = [i for i, t in enumerate([t1]) if i == 1 and t.sort not in (Boolean, TopS)]
// --  This iterates over `[t1]` (single element) with `i == 1` which
// is never true for a single-element list. So the validation is
// effectively dead code.
//
// Go: (formula.go:200-201): `NewCond` has no sort validation at all.
//
// **Status**: Both effectively skip validation. CONFORMANT (both have the same bug/non-behavior).
func NewCond(t1, t2 Node) (*Cond, error) {
	return &Cond{CSort: t2.NodeSort(), T1: t1, T2: t2}, nil
}

func (c *Cond) NodeSort() Sort   { return c.CSort }
func (c *Cond) Children() []Node { return []Node{c.T1, c.T2} }
func (c *Cond) String() string   { return fmt.Sprintf("Cond(%s, %s)", c.T1, c.T2) }
func (c *Cond) Equal(n Node) bool {
	if o, ok := n.(*Cond); ok {
		return c.T1.Equal(o.T1) && c.T2.Equal(o.T2)
	}
	return false
}

// --- And ---

type And struct {
	Terms []Node
}

func NewAnd(terms ...Node) (*And, error) {
	for i, t := range terms {
		if !IsBooleanOrTop(t.NodeSort()) {
			return nil, &SortError{Msg: fmt.Sprintf("Bad sorts in: And(%s) (positions: [%d])",
				nodeSliceStr(terms), i)}
		}
	}
	cp := make([]Node, len(terms))
	copy(cp, terms)
	return &And{Terms: cp}, nil
}

func (a *And) NodeSort() Sort   { return Boolean }
func (a *And) Children() []Node { return a.Terms }
func (a *And) String() string {
	return fmt.Sprintf("And(%s)", nodeSliceStr(a.Terms))
}
func (a *And) Equal(n Node) bool {
	if o, ok := n.(*And); ok {
		return nodeSliceEqual(a.Terms, o.Terms)
	}
	return false
}

// --- Or ---

type Or struct {
	Terms []Node
}

func NewOr(terms ...Node) (*Or, error) {
	for i, t := range terms {
		if !IsBooleanOrTop(t.NodeSort()) {
			return nil, &SortError{Msg: fmt.Sprintf("Bad sorts in: Or(%s) (positions: [%d])",
				nodeSliceStr(terms), i)}
		}
	}
	cp := make([]Node, len(terms))
	copy(cp, terms)
	return &Or{Terms: cp}, nil
}

func (o *Or) NodeSort() Sort   { return Boolean }
func (o *Or) Children() []Node { return o.Terms }
func (o *Or) String() string {
	return fmt.Sprintf("Or(%s)", nodeSliceStr(o.Terms))
}
func (o *Or) Equal(n Node) bool {
	if oo, ok := n.(*Or); ok {
		return nodeSliceEqual(o.Terms, oo.Terms)
	}
	return false
}

// --- Implies ---

type Implies struct {
	T1, T2 Node
}

func NewImplies(t1, t2 Node) (*Implies, error) {
	if !IsBooleanOrTop(t1.NodeSort()) || !IsBooleanOrTop(t2.NodeSort()) {
		return nil, &SortError{Msg: fmt.Sprintf("Bad sorts in: Implies(%s, %s)", t1, t2)}
	}
	return &Implies{T1: t1, T2: t2}, nil
}

func (i *Implies) NodeSort() Sort   { return Boolean }
func (i *Implies) Children() []Node { return []Node{i.T1, i.T2} }
func (i *Implies) String() string   { return fmt.Sprintf("(%s -> %s)", i.T1, i.T2) }
func (i *Implies) Equal(n Node) bool {
	if o, ok := n.(*Implies); ok {
		return i.T1.Equal(o.T1) && i.T2.Equal(o.T2)
	}
	return false
}

// --- Iff ---

type Iff struct {
	T1, T2 Node
}

func NewIff(t1, t2 Node) (*Iff, error) {
	if !IsBooleanOrTop(t1.NodeSort()) || !IsBooleanOrTop(t2.NodeSort()) {
		return nil, &SortError{Msg: fmt.Sprintf("Bad sorts in: Iff(%s, %s)", t1, t2)}
	}
	return &Iff{T1: t1, T2: t2}, nil
}

func (i *Iff) NodeSort() Sort   { return Boolean }
func (i *Iff) Children() []Node { return []Node{i.T1, i.T2} }
func (i *Iff) String() string   { return fmt.Sprintf("Iff(%s, %s)", i.T1, i.T2) }
func (i *Iff) Equal(n Node) bool {
	if o, ok := n.(*Iff); ok {
		return i.T1.Equal(o.T1) && i.T2.Equal(o.T2)
	}
	return false
}

// --- ForAll ---

type ForAll struct {
	Variables []*Variable
	Body      Node
}

func NewForAll(variables []*Variable, body Node) (*ForAll, error) {
	if len(variables) == 0 {
		return nil, &IvyError{Msg: "Must quantify over at least one variable"}
	}
	for _, v := range variables {
		if v == nil {
			return nil, &IvyError{Msg: "Can only quantify over variables"}
		}
	}
	if !IsBooleanOrTop(body.NodeSort()) {
		return nil, &SortError{Msg: fmt.Sprintf("Quantified body must be Boolean: %s", body)}
	}
	// Python stores variables as frozenset (unordered, deduplicated).
	// Sort and deduplicate by name to match Python's set semantics.
	cp := deduplicateAndSortVars(variables)
	return &ForAll{Variables: cp, Body: body}, nil
}

func (f *ForAll) NodeSort() Sort   { return Boolean }
func (f *ForAll) Children() []Node { return []Node{f.Body} }
func (f *ForAll) String() string {
	return fmt.Sprintf("(ForAll %s. %s)", varSortList(f.Variables), f.Body)
}
func (f *ForAll) Equal(n Node) bool {
	if o, ok := n.(*ForAll); ok {
		return varSliceEqual(f.Variables, o.Variables) && f.Body.Equal(o.Body)
	}
	return false
}

// --- Exists ---

type Exists struct {
	Variables []*Variable
	Body      Node
}

func NewExists(variables []*Variable, body Node) (*Exists, error) {
	if len(variables) == 0 {
		return nil, &IvyError{Msg: "Must quantify over at least one variable"}
	}
	for _, v := range variables {
		if v == nil {
			return nil, &IvyError{Msg: "Can only quantify over variables"}
		}
	}
	if !IsBooleanOrTop(body.NodeSort()) {
		return nil, &SortError{Msg: fmt.Sprintf("Quantified body must be Boolean: %s", body)}
	}
	// Python stores variables as frozenset (unordered, deduplicated).
	cp := deduplicateAndSortVars(variables)
	return &Exists{Variables: cp, Body: body}, nil
}

func (e *Exists) NodeSort() Sort   { return Boolean }
func (e *Exists) Children() []Node { return []Node{e.Body} }
func (e *Exists) String() string {
	return fmt.Sprintf("(Exists %s. %s)", varSortList(e.Variables), e.Body)
}
func (e *Exists) Equal(n Node) bool {
	if o, ok := n.(*Exists); ok {
		return varSliceEqual(e.Variables, o.Variables) && e.Body.Equal(o.Body)
	}
	return false
}

// --- Lambda ---

type Lambda struct {
	Variables []*Variable
	Body      Node
}

func NewLambda(variables []*Variable, body Node) (*Lambda, error) {
	for _, v := range variables {
		if v == nil {
			return nil, &IvyError{Msg: "Can only abstract over variables"}
		}
	}
	cp := make([]*Variable, len(variables))
	copy(cp, variables)
	return &Lambda{Variables: cp, Body: body}, nil
}

func (l *Lambda) NodeSort() Sort   { return Boolean }
func (l *Lambda) Children() []Node { return []Node{l.Body} }
func (l *Lambda) String() string {
	return fmt.Sprintf("(Lambda %s. %s)", varSortList(l.Variables), l.Body)
}
func (l *Lambda) Equal(n Node) bool {
	if o, ok := n.(*Lambda); ok {
		return varSliceEqual(l.Variables, o.Variables) && l.Body.Equal(o.Body)
	}
	return false
}

// --- NamedBinder ---

type NamedBinder struct {
	Name      string
	Variables []*Variable
	Environ   *string
	Body      Node
}

func NewNamedBinder(name string, variables []*Variable, environ *string, body Node) (*NamedBinder, error) {
	for _, v := range variables {
		if v == nil {
			return nil, &IvyError{Msg: "Can only abstract over variables"}
		}
	}
	cp := make([]*Variable, len(variables))
	copy(cp, variables)
	return &NamedBinder{Name: name, Variables: cp, Environ: environ, Body: body}, nil
}

func (nb *NamedBinder) NodeSort() Sort {
	if len(nb.Variables) > 0 {
		sorts := make([]Sort, len(nb.Variables)+1)
		for i, v := range nb.Variables {
			sorts[i] = v.VSort
		}
		sorts[len(nb.Variables)] = nb.Body.NodeSort()
		fs, err := NewFunctionSort(sorts...)
		if err != nil {
			// Fallback — shouldn't happen with well-formed binders
			return nb.Body.NodeSort()
		}
		return fs
	}
	return nb.Body.NodeSort()
}

func (nb *NamedBinder) Children() []Node { return []Node{nb.Body} }

func (nb *NamedBinder) String() string {
	env := ""
	if nb.Environ != nil {
		env = "[" + *nb.Environ + "]"
	}
	return fmt.Sprintf("($%s%s %s. %s)", nb.Name, env, varSortList(nb.Variables), nb.Body)
}

func (nb *NamedBinder) Equal(n Node) bool {
	if o, ok := n.(*NamedBinder); ok {
		return nb.Name == o.Name &&
			ptrStrEqual(nb.Environ, o.Environ) &&
			varSliceEqual(nb.Variables, o.Variables) &&
			nb.Body.Equal(o.Body)
	}
	return false
}

// Call applies the binder as a function. Returns self if no args.
func (nb *NamedBinder) Call(terms ...Node) (Node, error) {
	if len(terms) == 0 {
		return nb, nil
	}
	return NewApply(nb, terms...)
}

// --- helpers ---

// deduplicateAndSortVars deduplicates variables by name and sorts by name.
// This matches Python's frozenset(variables) behavior for ForAll/Exists:
// unordered, deduplicated. We sort by name to produce a canonical order.
func deduplicateAndSortVars(vars []*Variable) []*Variable {
	// Uses Sexp-based structural identity to match Python's frozenset
	// which deduplicates by structural equality (name + sort).
	seen := make(map[NodeKey]bool, len(vars))
	var result []*Variable
	for _, v := range vars {
		k := Key(v)
		if !seen[k] {
			seen[k] = true
			result = append(result, v)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

func ptrStrEqual(a, b *string) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func nodeSliceStr(nodes []Node) string {
	parts := make([]string, len(nodes))
	for i, n := range nodes {
		parts[i] = n.String()
	}
	return strings.Join(parts, ", ")
}

func nodeSliceEqual(a, b []Node) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !a[i].Equal(b[i]) {
			return false
		}
	}
	return true
}

func varSliceEqual(a, b []*Variable) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !a[i].Equal(b[i]) {
			return false
		}
	}
	return true
}

// varSortList returns sorted "V:Sort, W:Sort" string for variables.
func varSortList(vars []*Variable) string {
	type vs struct {
		name string
		sort string
	}
	entries := make([]vs, len(vars))
	for i, v := range vars {
		entries[i] = vs{v.Name, v.VSort.String()}
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].name < entries[j].name
	})
	parts := make([]string, len(entries))
	for i, e := range entries {
		parts[i] = e.name + ":" + e.sort
	}
	return strings.Join(parts, ", ")
}
