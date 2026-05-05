package logic

import (
	"fmt"
	"sort"
	"strings"

	"github.com/glycerine/ivy/goivy/ast"
)

// True and False are the logical constants (empty And / empty Or).
var (
	True  Expr = &And{}
	False Expr = &Or{}
)

// IsTrue returns true if the node is logical true (empty And).
// Handles both the singleton pointer and any structurally-equivalent &And{}.
func IsTrue(n Expr) bool {
	a, ok := n.(*And)
	return ok && len(a.Terms) == 0
}

// IsFalse returns true if the node is logical false (empty Or).
// Handles both the singleton pointer and any structurally-equivalent &Or{}.
func IsFalse(n Expr) bool {
	o, ok := n.(*Or)
	return ok && len(o.Terms) == 0
}

// --- Eq ---

type Eq struct {
	ast.Base
	T1, T2 Expr
}

func NewEq(t1, t2 Expr) (*Eq, error) {
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
func (e *Eq) Children() []Expr { return []Expr{e.T1, e.T2} }

// String matches Python ivy_logic.py:1444-1446 which monkey-patches
// lg.Eq.__str__ = pretty_fmla.
func (e *Eq) String() string { return PrettyFmla(e) }
func (e *Eq) Equal(n Expr) bool {
	if o, ok := n.(*Eq); ok {
		return e.T1.Equal(o.T1) && e.T2.Equal(o.T2)
	}
	return false
}

// --- Ite ---

type Ite struct {
	ast.Base
	ISort Sort
	Cond  Expr
	Then  Expr
	Else  Expr
}

func NewIte(cond, then_, else_ Expr) (*Ite, error) {
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
func (t *Ite) Children() []Expr { return []Expr{t.Cond, t.Then, t.Else} }

// String matches Python ivy_logic.py:1444-1446 which monkey-patches
// lg.Ite.__str__ = pretty_fmla.
func (t *Ite) String() string { return PrettyFmla(t) }
func (t *Ite) Equal(n Expr) bool {
	if o, ok := n.(*Ite); ok {
		return t.Cond.Equal(o.Cond) && t.Then.Equal(o.Then) && t.Else.Equal(o.Else)
	}
	return false
}

// --- Not ---

type Not struct {
	ast.Base
	Body Expr
}

func NewNot(body Expr) (*Not, error) {
	if !IsBooleanOrTop(body.NodeSort()) {
		return nil, &SortError{Msg: fmt.Sprintf("Negation body must be Boolean: %s", body)}
	}
	return &Not{Body: body}, nil
}

func (n *Not) NodeSort() Sort   { return Boolean }
func (n *Not) Children() []Expr { return []Expr{n.Body} }

// String matches Python ivy_logic.py:1444-1446 which monkey-patches
// lg.Not.__str__ = pretty_fmla. Note Not.ugly handles Not(Eq(a,b)) → "a ~= b"
// and Not(other) → "~other".
func (n *Not) String() string { return PrettyFmla(n) }
func (n *Not) Equal(nd Expr) bool {
	if o, ok := nd.(*Not); ok {
		return n.Body.Equal(o.Body)
	}
	return false
}

// --- Globally ---

type Globally struct {
	ast.Base
	Environ *string
	Body    Expr
}

func NewGlobally(environ *string, body Expr) (*Globally, error) {
	if !IsBooleanOrTop(body.NodeSort()) {
		return nil, &SortError{Msg: fmt.Sprintf("Globally body must be Boolean: %s", body)}
	}
	return &Globally{Environ: environ, Body: body}, nil
}

func (g *Globally) NodeSort() Sort   { return Boolean }
func (g *Globally) Children() []Expr { return []Expr{g.Body} }
func (g *Globally) String() string {
	env := ""
	if g.Environ != nil {
		env = "[" + *g.Environ + "]"
	}
	return fmt.Sprintf("globally%s(%s)", env, g.Body)
}
func (g *Globally) Equal(n Expr) bool {
	if o, ok := n.(*Globally); ok {
		return ptrStrEqual(g.Environ, o.Environ) && g.Body.Equal(o.Body)
	}
	return false
}

// --- Eventually ---

type Eventually struct {
	ast.Base
	Environ *string
	Body    Expr
}

func NewEventually(environ *string, body Expr) (*Eventually, error) {
	if !IsBooleanOrTop(body.NodeSort()) {
		return nil, &SortError{Msg: fmt.Sprintf("Eventually body must be Boolean: %s", body)}
	}
	return &Eventually{Environ: environ, Body: body}, nil
}

func (e *Eventually) NodeSort() Sort   { return Boolean }
func (e *Eventually) Children() []Expr { return []Expr{e.Body} }
func (e *Eventually) String() string {
	env := ""
	if e.Environ != nil {
		env = "[" + *e.Environ + "]"
	}
	return fmt.Sprintf("eventually%s(%s)", env, e.Body)
}
func (e *Eventually) Equal(n Expr) bool {
	if o, ok := n.(*Eventually); ok {
		return ptrStrEqual(e.Environ, o.Environ) && e.Body.Equal(o.Body)
	}
	return false
}

// --- WhenOperator ---

type WhenOperator struct {
	ast.Base
	WSort Sort
	Name  string
	T1    Expr
	T2    Expr
}

func NewWhenOperator(name string, t1, t2 Expr) (*WhenOperator, error) {
	if !IsBooleanOrTop(t2.NodeSort()) {
		return nil, &SortError{Msg: fmt.Sprintf("WhenOperator second argument must be Boolean: %s", t2)}
	}
	return &WhenOperator{WSort: t1.NodeSort(), Name: name, T1: t1, T2: t2}, nil
}

func (w *WhenOperator) NodeSort() Sort   { return w.WSort }
func (w *WhenOperator) Children() []Expr { return []Expr{w.T1, w.T2} }
func (w *WhenOperator) String() string {
	return fmt.Sprintf("WhenOperator(%s,%s,%s)", w.Name, w.T1, w.T2)
}
func (w *WhenOperator) Equal(n Expr) bool {
	if o, ok := n.(*WhenOperator); ok {
		return w.Name == o.Name && w.T1.Equal(o.T1) && w.T2.Equal(o.T2)
	}
	return false
}

// --- Cond ---

type Cond struct {
	ast.Base
	CSort Sort
	T1    Expr
	T2    Expr
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
func NewCond(t1, t2 Expr) (*Cond, error) {
	return &Cond{CSort: t2.NodeSort(), T1: t1, T2: t2}, nil
}

func (c *Cond) NodeSort() Sort   { return c.CSort }
func (c *Cond) Children() []Expr { return []Expr{c.T1, c.T2} }
func (c *Cond) String() string   { return fmt.Sprintf("Cond(%s, %s)", c.T1, c.T2) }
func (c *Cond) Equal(n Expr) bool {
	if o, ok := n.(*Cond); ok {
		return c.T1.Equal(o.T1) && c.T2.Equal(o.T2)
	}
	return false
}

// --- And ---

type And struct {
	ast.Base
	Terms []Expr
}

func NewAnd(terms ...Expr) (*And, error) {
	for i, t := range terms {
		if !IsBooleanOrTop(t.NodeSort()) {
			return nil, &SortError{Msg: fmt.Sprintf("Bad sorts in: And(%s) (positions: [%d])",
				nodeSliceStr(terms), i)}
		}
	}
	cp := make([]Expr, len(terms))
	copy(cp, terms)
	return &And{Terms: cp}, nil
}

func (a *And) NodeSort() Sort   { return Boolean }
func (a *And) Children() []Expr { return a.Terms }

// String matches Python ivy_logic.py:1444-1446 which monkey-patches
// lg.And.__str__ = pretty_fmla. Empty And renders as "true".
func (a *And) String() string { return PrettyFmla(a) }
func (a *And) Equal(n Expr) bool {
	if o, ok := n.(*And); ok {
		return nodeSliceEqual(a.Terms, o.Terms)
	}
	return false
}

// --- Or ---

type Or struct {
	ast.Base
	Terms []Expr
}

func NewOr(terms ...Expr) (*Or, error) {
	for i, t := range terms {
		if !IsBooleanOrTop(t.NodeSort()) {
			return nil, &SortError{Msg: fmt.Sprintf("Bad sorts in: Or(%s) (positions: [%d])",
				nodeSliceStr(terms), i)}
		}
	}
	cp := make([]Expr, len(terms))
	copy(cp, terms)
	return &Or{Terms: cp}, nil
}

func (o *Or) NodeSort() Sort   { return Boolean }
func (o *Or) Children() []Expr { return o.Terms }

// String matches Python ivy_logic.py:1444-1446 which monkey-patches
// lg.Or.__str__ = pretty_fmla. Empty Or renders as "false".
func (o *Or) String() string { return PrettyFmla(o) }
func (o *Or) Equal(n Expr) bool {
	if oo, ok := n.(*Or); ok {
		return nodeSliceEqual(o.Terms, oo.Terms)
	}
	return false
}

// --- Implies ---

type Implies struct {
	ast.Base
	T1, T2 Expr
}

func NewImplies(t1, t2 Expr) (*Implies, error) {
	if !IsBooleanOrTop(t1.NodeSort()) || !IsBooleanOrTop(t2.NodeSort()) {
		return nil, &SortError{Msg: fmt.Sprintf("Bad sorts in: Implies(%s, %s)", t1, t2)}
	}
	return &Implies{T1: t1, T2: t2}, nil
}

func (i *Implies) NodeSort() Sort   { return Boolean }
func (i *Implies) Children() []Expr { return []Expr{i.T1, i.T2} }

// String matches Python ivy_logic.py:1444-1446 which monkey-patches
// lg.Implies.__str__ = pretty_fmla.
func (i *Implies) String() string { return PrettyFmla(i) }
func (i *Implies) Equal(n Expr) bool {
	if o, ok := n.(*Implies); ok {
		return i.T1.Equal(o.T1) && i.T2.Equal(o.T2)
	}
	return false
}

// --- Iff ---

type Iff struct {
	ast.Base
	T1, T2 Expr
}

func NewIff(t1, t2 Expr) (*Iff, error) {
	if !IsBooleanOrTop(t1.NodeSort()) || !IsBooleanOrTop(t2.NodeSort()) {
		return nil, &SortError{Msg: fmt.Sprintf("Bad sorts in: Iff(%s, %s)", t1, t2)}
	}
	return &Iff{T1: t1, T2: t2}, nil
}

func (i *Iff) NodeSort() Sort   { return Boolean }
func (i *Iff) Children() []Expr { return []Expr{i.T1, i.T2} }

// String matches Python ivy_logic.py:1444-1446 which monkey-patches
// lg.Iff.__str__ = pretty_fmla.
func (i *Iff) String() string { return PrettyFmla(i) }
func (i *Iff) Equal(n Expr) bool {
	if o, ok := n.(*Iff); ok {
		return i.T1.Equal(o.T1) && i.T2.Equal(o.T2)
	}
	return false
}

// --- ForAll ---

type ForAll struct {
	ast.Base
	Variables []*Variable
	Body      Expr
}

func NewForAll(variables []*Variable, body Expr) (*ForAll, error) {
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
func (f *ForAll) Children() []Expr { return []Expr{f.Body} }

// String matches Python ivy_logic.py:1444-1446 which monkey-patches
// lg.ForAll.__str__ = pretty_fmla.
func (f *ForAll) String() string { return PrettyFmla(f) }
func (f *ForAll) Equal(n Expr) bool {
	if o, ok := n.(*ForAll); ok {
		return varSliceEqual(f.Variables, o.Variables) && f.Body.Equal(o.Body)
	}
	return false
}

// --- Exists ---

type Exists struct {
	ast.Base
	Variables []*Variable
	Body      Expr
}

func NewExists(variables []*Variable, body Expr) (*Exists, error) {
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
func (e *Exists) Children() []Expr { return []Expr{e.Body} }

// String matches Python ivy_logic.py:1444-1446 which monkey-patches
// lg.Exists.__str__ = pretty_fmla.
func (e *Exists) String() string { return PrettyFmla(e) }
func (e *Exists) Equal(n Expr) bool {
	if o, ok := n.(*Exists); ok {
		return varSliceEqual(e.Variables, o.Variables) && e.Body.Equal(o.Body)
	}
	return false
}

// --- Lambda ---

type Lambda struct {
	ast.Base
	Variables []*Variable
	Body      Expr
}

func NewLambda(variables []*Variable, body Expr) (*Lambda, error) {
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
func (l *Lambda) Children() []Expr { return []Expr{l.Body} }

// String matches Python ivy_logic.py:1444-1446 which monkey-patches
// lg.Lambda.__str__ = pretty_fmla.
func (l *Lambda) String() string { return PrettyFmla(l) }
func (l *Lambda) Equal(n Expr) bool {
	if o, ok := n.(*Lambda); ok {
		return varSliceEqual(l.Variables, o.Variables) && l.Body.Equal(o.Body)
	}
	return false
}

// --- NamedBinder ---

type NamedBinder struct {
	ast.Base
	Name      string
	Variables []*Variable
	Environ   *string
	Body      Expr
}

func NewNamedBinder(name string, variables []*Variable, environ *string, body Expr) (*NamedBinder, error) {
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

func (nb *NamedBinder) Children() []Expr { return []Expr{nb.Body} }

// String matches Python ivy_logic.py:1444-1446 which monkey-patches
// lg.NamedBinder.__str__ = pretty_fmla.
func (nb *NamedBinder) String() string { return PrettyFmla(nb) }

func (nb *NamedBinder) Equal(n Expr) bool {
	if o, ok := n.(*NamedBinder); ok {
		return nb.Name == o.Name &&
			ptrStrEqual(nb.Environ, o.Environ) &&
			varSliceEqual(nb.Variables, o.Variables) &&
			nb.Body.Equal(o.Body)
	}
	return false
}

// Call applies the binder as a function. Returns self if no args.
func (nb *NamedBinder) Call(terms ...Expr) (Expr, error) {
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

func nodeSliceStr(nodes []Expr) string {
	parts := make([]string, len(nodes))
	for i, n := range nodes {
		parts[i] = n.String()
	}
	return strings.Join(parts, ", ")
}

func nodeSliceEqual(a, b []Expr) bool {
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
