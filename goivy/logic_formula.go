package goivy

import (
	"fmt"
	"sort"
	"strings"
)

// True and False are the logical constants (empty And / empty Or).
var (
	True  Expr = &LogicAnd{}
	False Expr = &LogicOr{}
)

// IsTrue returns true if the node is logical true (empty And).
// Handles both the singleton pointer and any structurally-equivalent &LogicAnd{}.
func IsTrue(n Expr) bool {
	a, ok := n.(*LogicAnd)
	return ok && len(a.Terms) == 0
}

// IsFalse returns true if the node is logical false (empty Or).
// Handles both the singleton pointer and any structurally-equivalent &LogicOr{}.
func IsFalse(n Expr) bool {
	o, ok := n.(*LogicOr)
	return ok && len(o.Terms) == 0
}

// --- Eq ---

type Eq struct {
	Base
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

type LogicIte struct {
	Base
	ISort Sort
	Cond  Expr
	Then  Expr
	Else  Expr
}

func NewIte(cond, then_, else_ Expr) (*LogicIte, error) {
	if !IsBooleanOrTop(cond.NodeSort()) {
		return nil, &SortError{Msg: fmt.Sprintf("Ite condition must be Boolean: %s", cond)}
	}
	s1, s2 := then_.NodeSort(), else_.NodeSort()
	_, t1Top := s1.(*TopSort)
	_, t2Top := s2.(*TopSort)
	if !t1Top && !t2Top && !s1.Equal(s2) {
		return nil, &SortError{Msg: fmt.Sprintf("Ite then and else terms must have same sort: %s, %s", then_, else_)}
	}
	return &LogicIte{ISort: then_.NodeSort(), Cond: cond, Then: then_, Else: else_}, nil
}

func (t *LogicIte) NodeSort() Sort   { return t.ISort }
func (t *LogicIte) Children() []Expr { return []Expr{t.Cond, t.Then, t.Else} }

// String matches Python ivy_logic.py:1444-1446 which monkey-patches
// lg.Ite.__str__ = pretty_fmla.
func (t *LogicIte) String() string { return PrettyFmla(t) }
func (t *LogicIte) Equal(n Expr) bool {
	if o, ok := n.(*LogicIte); ok {
		return t.Cond.Equal(o.Cond) && t.Then.Equal(o.Then) && t.Else.Equal(o.Else)
	}
	return false
}

// --- Not ---

type LogicNot struct {
	Base
	Body Expr
}

func NewNot(body Expr) (*LogicNot, error) {
	if !IsBooleanOrTop(body.NodeSort()) {
		return nil, &SortError{Msg: fmt.Sprintf("Negation body must be Boolean: %s", body)}
	}
	return &LogicNot{Body: body}, nil
}

func (n *LogicNot) NodeSort() Sort   { return Boolean }
func (n *LogicNot) Children() []Expr { return []Expr{n.Body} }

// String matches Python ivy_logic.py:1444-1446 which monkey-patches
// lg.Not.__str__ = pretty_fmla. Note Not.ugly handles Not(Eq(a,b)) → "a ~= b"
// and Not(other) → "~other".
func (n *LogicNot) String() string { return PrettyFmla(n) }
func (n *LogicNot) Equal(nd Expr) bool {
	if o, ok := nd.(*LogicNot); ok {
		return n.Body.Equal(o.Body)
	}
	return false
}

// --- Globally ---

type LogicGlobally struct {
	Base
	Environ *string
	Body    Expr
}

func NewGlobally(environ *string, body Expr) (*LogicGlobally, error) {
	if !IsBooleanOrTop(body.NodeSort()) {
		return nil, &SortError{Msg: fmt.Sprintf("Globally body must be Boolean: %s", body)}
	}
	return &LogicGlobally{Environ: environ, Body: body}, nil
}

func (g *LogicGlobally) NodeSort() Sort   { return Boolean }
func (g *LogicGlobally) Children() []Expr { return []Expr{g.Body} }
func (g *LogicGlobally) String() string {
	env := ""
	if g.Environ != nil {
		env = "[" + *g.Environ + "]"
	}
	return fmt.Sprintf("globally%s(%s)", env, g.Body)
}
func (g *LogicGlobally) Equal(n Expr) bool {
	if o, ok := n.(*LogicGlobally); ok {
		return ptrStrEqual(g.Environ, o.Environ) && g.Body.Equal(o.Body)
	}
	return false
}

// --- Eventually ---

type LogicEventually struct {
	Base
	Environ *string
	Body    Expr
}

func NewEventually(environ *string, body Expr) (*LogicEventually, error) {
	if !IsBooleanOrTop(body.NodeSort()) {
		return nil, &SortError{Msg: fmt.Sprintf("Eventually body must be Boolean: %s", body)}
	}
	return &LogicEventually{Environ: environ, Body: body}, nil
}

func (e *LogicEventually) NodeSort() Sort   { return Boolean }
func (e *LogicEventually) Children() []Expr { return []Expr{e.Body} }
func (e *LogicEventually) String() string {
	env := ""
	if e.Environ != nil {
		env = "[" + *e.Environ + "]"
	}
	return fmt.Sprintf("eventually%s(%s)", env, e.Body)
}
func (e *LogicEventually) Equal(n Expr) bool {
	if o, ok := n.(*LogicEventually); ok {
		return ptrStrEqual(e.Environ, o.Environ) && e.Body.Equal(o.Body)
	}
	return false
}

// --- WhenOperator ---

type LogicWhenOperator struct {
	Base
	WSort Sort
	Name  string
	T1    Expr
	T2    Expr
}

func NewWhenOperator(name string, t1, t2 Expr) (*LogicWhenOperator, error) {
	if !IsBooleanOrTop(t2.NodeSort()) {
		return nil, &SortError{Msg: fmt.Sprintf("WhenOperator second argument must be Boolean: %s", t2)}
	}
	return &LogicWhenOperator{WSort: t1.NodeSort(), Name: name, T1: t1, T2: t2}, nil
}

func (w *LogicWhenOperator) NodeSort() Sort   { return w.WSort }
func (w *LogicWhenOperator) Children() []Expr { return []Expr{w.T1, w.T2} }
func (w *LogicWhenOperator) String() string {
	return fmt.Sprintf("WhenOperator(%s,%s,%s)", w.Name, w.T1, w.T2)
}
func (w *LogicWhenOperator) Equal(n Expr) bool {
	if o, ok := n.(*LogicWhenOperator); ok {
		return w.Name == o.Name && w.T1.Equal(o.T1) && w.T2.Equal(o.T2)
	}
	return false
}

// --- Cond ---

type Cond struct {
	Base
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

type LogicAnd struct {
	Base
	Terms []Expr
}

func NewAnd(terms ...Expr) (*LogicAnd, error) {
	for i, t := range terms {
		if !IsBooleanOrTop(t.NodeSort()) {
			return nil, &SortError{Msg: fmt.Sprintf("Bad sorts in: And(%s) (positions: [%d])",
				nodeSliceStr(terms), i)}
		}
	}
	cp := make([]Expr, len(terms))
	copy(cp, terms)
	return &LogicAnd{Terms: cp}, nil
}

func (a *LogicAnd) NodeSort() Sort   { return Boolean }
func (a *LogicAnd) Children() []Expr { return a.Terms }

// String matches Python ivy_logic.py:1444-1446 which monkey-patches
// lg.And.__str__ = pretty_fmla. Empty And renders as "true".
func (a *LogicAnd) String() string { return PrettyFmla(a) }
func (a *LogicAnd) Equal(n Expr) bool {
	if o, ok := n.(*LogicAnd); ok {
		return nodeSliceEqual(a.Terms, o.Terms)
	}
	return false
}

// --- Or ---

type LogicOr struct {
	Base
	Terms []Expr
}

func NewOr(terms ...Expr) (*LogicOr, error) {
	for i, t := range terms {
		if !IsBooleanOrTop(t.NodeSort()) {
			return nil, &SortError{Msg: fmt.Sprintf("Bad sorts in: Or(%s) (positions: [%d])",
				nodeSliceStr(terms), i)}
		}
	}
	cp := make([]Expr, len(terms))
	copy(cp, terms)
	return &LogicOr{Terms: cp}, nil
}

func (o *LogicOr) NodeSort() Sort   { return Boolean }
func (o *LogicOr) Children() []Expr { return o.Terms }

// String matches Python ivy_logic.py:1444-1446 which monkey-patches
// lg.Or.__str__ = pretty_fmla. Empty Or renders as "false".
func (o *LogicOr) String() string { return PrettyFmla(o) }
func (o *LogicOr) Equal(n Expr) bool {
	if oo, ok := n.(*LogicOr); ok {
		return nodeSliceEqual(o.Terms, oo.Terms)
	}
	return false
}

// --- Implies ---

type LogicImplies struct {
	Base
	T1, T2 Expr
}

func NewImplies(t1, t2 Expr) (*LogicImplies, error) {
	if !IsBooleanOrTop(t1.NodeSort()) || !IsBooleanOrTop(t2.NodeSort()) {
		return nil, &SortError{Msg: fmt.Sprintf("Bad sorts in: Implies(%s, %s)", t1, t2)}
	}
	return &LogicImplies{T1: t1, T2: t2}, nil
}

func (i *LogicImplies) NodeSort() Sort   { return Boolean }
func (i *LogicImplies) Children() []Expr { return []Expr{i.T1, i.T2} }

// String matches Python ivy_logic.py:1444-1446 which monkey-patches
// lg.Implies.__str__ = pretty_fmla.
func (i *LogicImplies) String() string { return PrettyFmla(i) }
func (i *LogicImplies) Equal(n Expr) bool {
	if o, ok := n.(*LogicImplies); ok {
		return i.T1.Equal(o.T1) && i.T2.Equal(o.T2)
	}
	return false
}

// --- Iff ---

type LogicIff struct {
	Base
	T1, T2 Expr
}

func NewIff(t1, t2 Expr) (*LogicIff, error) {
	if !IsBooleanOrTop(t1.NodeSort()) || !IsBooleanOrTop(t2.NodeSort()) {
		return nil, &SortError{Msg: fmt.Sprintf("Bad sorts in: Iff(%s, %s)", t1, t2)}
	}
	return &LogicIff{T1: t1, T2: t2}, nil
}

func (i *LogicIff) NodeSort() Sort   { return Boolean }
func (i *LogicIff) Children() []Expr { return []Expr{i.T1, i.T2} }

// String matches Python ivy_logic.py:1444-1446 which monkey-patches
// lg.Iff.__str__ = pretty_fmla.
func (i *LogicIff) String() string { return PrettyFmla(i) }
func (i *LogicIff) Equal(n Expr) bool {
	if o, ok := n.(*LogicIff); ok {
		return i.T1.Equal(o.T1) && i.T2.Equal(o.T2)
	}
	return false
}

// --- ForAll ---

type ForAll struct {
	Base
	Variables []*LogicVariable
	Body      Expr
}

func NewForAll(variables []*LogicVariable, body Expr) (*ForAll, error) {
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

// RawForAll is an internal solver helper quantifier. Unlike ordinary Ivy
// ForAll, the Z3 translator must not add interpreted-sort domain constraints.
type RawForAll struct {
	Base
	Variables []*LogicVariable
	Body      Expr
	RawNames  bool
}

func (f *RawForAll) NodeSort() Sort   { return Boolean }
func (f *RawForAll) Children() []Expr { return []Expr{f.Body} }
func (f *RawForAll) String() string   { return PrettyFmla(f) }
func (f *RawForAll) Equal(n Expr) bool {
	if o, ok := n.(*RawForAll); ok {
		return f.RawNames == o.RawNames && varSliceEqual(f.Variables, o.Variables) && f.Body.Equal(o.Body)
	}
	return false
}

// --- Exists ---

type LogicExists struct {
	Base
	Variables []*LogicVariable
	Body      Expr
}

func NewExists(variables []*LogicVariable, body Expr) (*LogicExists, error) {
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
	return &LogicExists{Variables: cp, Body: body}, nil
}

func (e *LogicExists) NodeSort() Sort   { return Boolean }
func (e *LogicExists) Children() []Expr { return []Expr{e.Body} }

// String matches Python ivy_logic.py:1444-1446 which monkey-patches
// lg.Exists.__str__ = pretty_fmla.
func (e *LogicExists) String() string { return PrettyFmla(e) }
func (e *LogicExists) Equal(n Expr) bool {
	if o, ok := n.(*LogicExists); ok {
		return varSliceEqual(e.Variables, o.Variables) && e.Body.Equal(o.Body)
	}
	return false
}

// --- Lambda ---

type Lambda struct {
	Base
	Variables []*LogicVariable
	Body      Expr
}

func NewLambda(variables []*LogicVariable, body Expr) (*Lambda, error) {
	for _, v := range variables {
		if v == nil {
			return nil, &IvyError{Msg: "Can only abstract over variables"}
		}
	}
	cp := make([]*LogicVariable, len(variables))
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

type LogicNamedBinder struct {
	Base
	Name      string
	Variables []*LogicVariable
	Environ   *string
	Body      Expr
}

func NewNamedBinder(name string, variables []*LogicVariable, environ *string, body Expr) (*LogicNamedBinder, error) {
	for _, v := range variables {
		if v == nil {
			return nil, &IvyError{Msg: "Can only abstract over variables"}
		}
	}
	cp := make([]*LogicVariable, len(variables))
	copy(cp, variables)
	return &LogicNamedBinder{Name: name, Variables: cp, Environ: environ, Body: body}, nil
}

func (nb *LogicNamedBinder) NodeSort() Sort {
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

func (nb *LogicNamedBinder) Children() []Expr { return []Expr{nb.Body} }

// String matches Python ivy_logic.py:1444-1446 which monkey-patches
// lg.NamedBinder.__str__ = pretty_fmla.
func (nb *LogicNamedBinder) String() string { return PrettyFmla(nb) }

func (nb *LogicNamedBinder) Equal(n Expr) bool {
	if o, ok := n.(*LogicNamedBinder); ok {
		return nb.Name == o.Name &&
			ptrStrEqual(nb.Environ, o.Environ) &&
			varSliceEqual(nb.Variables, o.Variables) &&
			nb.Body.Equal(o.Body)
	}
	return false
}

// Call applies the binder as a function. Returns self if no args.
func (nb *LogicNamedBinder) Call(terms ...Expr) (Expr, error) {
	if len(terms) == 0 {
		return nb, nil
	}
	return NewApply(nb, terms...)
}

// --- helpers ---

// deduplicateAndSortVars deduplicates variables by name and sorts by name.
// This matches Python's frozenset(variables) behavior for ForAll/Exists:
// unordered, deduplicated. We sort by name to produce a canonical order.
func deduplicateAndSortVars(vars []*LogicVariable) []*LogicVariable {
	// Uses Sexp-based structural identity to match Python's frozenset
	// which deduplicates by structural equality (name + sort).
	seen := make(map[NodeKey]bool, len(vars))
	var result []*LogicVariable
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

func varSliceEqual(a, b []*LogicVariable) bool {
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
func varSortList(vars []*LogicVariable) string {
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
