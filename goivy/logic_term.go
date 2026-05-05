package goivy

import (
	"fmt"
	"strings"
)

// ReprExpr returns the Repr() of an expression if available, else String().
// This matches Python's __repr__ behavior for sort-qualified display.
func ReprExpr(e Expr) string {
	type reprer interface{ Repr() string }
	if r, ok := e.(reprer); ok {
		return r.Repr()
	}
	if e == nil {
		return "<nil>"
	}
	return e.String()
}

// ReprNode returns the sort-qualified repr of any ast.Node.
// Calls Repr() if available (on lg.Variable, lg.Apply, ast.Atom, etc.),
// else falls back to String(). This is the public entry point for traces.
func ReprNode(n Node) string {
	if n == nil {
		return ""
	}
	type reprer interface{ Repr() string }
	if r, ok := n.(reprer); ok {
		return r.Repr()
	}
	return fmt.Sprint(n)
}

// Variable represents a variable. Name must start with uppercase.
type LogicVariable struct {
	Base
	Name  string
	VSort Sort
}

func NewVariable(name string, sort Sort) (*LogicVariable, error) {
	if len(name) == 0 || !isUpper(name[0]) {
		return nil, &IvyError{Msg: fmt.Sprintf("Bad variable name: %q", name)}
	}
	return &LogicVariable{Name: name, VSort: sort}, nil
}

func (v *LogicVariable) NodeSort() Sort   { return v.VSort }
func (v *LogicVariable) Children() []Expr { return nil }

// String matches Python ivy_logic.py:1444-1446 which monkey-patches
// lg.Var.__str__ = pretty_fmla. PrettyFmla calls
// drop_annotations(False, set()).ugly(0).
func (v *LogicVariable) String() string { return PrettyFmla(v) }

// Repr returns a sort-qualified string, matching Python ast.Variable.__repr__
// which always includes ':sort'. String() matches Python logic.Var.__str__
// which returns just the name.
func (v *LogicVariable) Repr() string {
	if v.VSort != nil && !IsTopSort(v.VSort) {
		return v.Name + ":" + v.VSort.String()
	}
	return v.Name
}

func (v *LogicVariable) Equal(n Expr) bool {
	if o, ok := n.(*LogicVariable); ok {
		return v.Name == o.Name && v.VSort.Equal(o.VSort)
	}
	return false
}

// Call applies the variable as a function. Returns self if no args.
func (v *LogicVariable) Call(terms ...Expr) (Expr, error) {
	if len(terms) == 0 {
		return v, nil
	}
	return NewApply(v, terms...)
}

// Symbol represents a constant symbol.
type Const struct {
	Base
	Name  string
	CSort Sort
	sexp  NodeKey
}

func NewConst(name string, sort Sort) *Const {
	c := &Const{Name: name, CSort: sort}
	var sortSexp NodeKey
	if sort != nil {
		sortSexp = sort.Sexp()
	} else {
		sortSexp = "nil"
	}
	c.sexp = NodeKey(fmt.Sprintf("(Symbol name:%v sort:%v)", name, sortSexp))
	return c
}

func (c *Const) NodeSort() Sort   { return c.CSort }
func (c *Const) Children() []Expr { return nil }

// String matches Python ivy_logic.py:1444-1446 which monkey-patches
// lg.Const.__str__ = pretty_fmla. PrettyFmla calls
// drop_annotations(False, set()).ugly(0); for a Const this returns
// "name:sortName" for numerals with non-TopSort, else just "name".
func (c *Const) String() string { return PrettyFmla(c) }

// Repr returns the sort-qualified representation. For Const, same as String().
func (c *Const) Repr() string { return c.Name }

func (c *Const) Equal(n Expr) bool {
	if o, ok := n.(*Const); ok {
		return c.Name == o.Name && c.CSort.Equal(o.CSort)
	}
	return false
}

// Call applies the constant as a function.
// Matches Python: Symbol.__call__ = lambda self,*args: App(self,*args)
//
//	if len(args) > 0 or isinstance(self.sort, FunctionSort) else self
//
// If zero args and CSort is FunctionSort, creates Apply(c) (nullary application).
// If zero args and CSort is NOT FunctionSort, returns self.
func (c *Const) Call(terms ...Expr) (Expr, error) {
	if len(terms) == 0 {
		if _, isFS := c.CSort.(*LogicFunctionSort); isFS {
			return NewApply(c) // nullary application
		}
		return c, nil
	}
	return NewApply(c, terms...)
}

// Apply represents function application.
type Apply struct {
	Base
	Func  Expr
	Terms []Expr
	aSort Sort // cached sort
}

func NewApply(fn Expr, terms ...Expr) (*Apply, error) {
	fnSort := fn.NodeSort()

	switch fs := fnSort.(type) {
	case *TopSort:
		// TopSort: accept anything, result is TopS
		cp := make([]Expr, len(terms))
		copy(cp, terms)
		return &Apply{Func: fn, Terms: cp, aSort: TopS}, nil

	case *LogicFunctionSort:
		if fs == nil {
			// Nil typed pointer — treat like TopSort
			cp := make([]Expr, len(terms))
			copy(cp, terms)
			return &Apply{Func: fn, Terms: cp, aSort: TopS}, nil
		}
		if fs.Arity() != len(terms) {
			termStrs := make([]string, len(terms))
			for i, t := range terms {
				termStrs[i] = t.String()
			}
			return nil, &SortError{
				Msg: fmt.Sprintf("Bad arity in: %s(%s) : function has arity %d (sort %s), but %d terms provided",
					fn.String(), strings.Join(termStrs, ", "), fs.Arity(), fs.String(), len(terms)),
			}
		}
		domain := fs.Domain()
		for i := range terms {
			tSort := terms[i].NodeSort()
			dSort := domain[i]
			if tSort == nil || dSort == nil || tSort.Equal(dSort) {
				continue
			}
			_, tIsTop := tSort.(*TopSort)
			_, dIsTop := dSort.(*TopSort)
			if tIsTop || dIsTop {
				continue
			}
			return nil, reportBadSort(fn, i, dSort, tSort)
		}
		cp := make([]Expr, len(terms))
		copy(cp, terms)
		return &Apply{Func: fn, Terms: cp, aSort: fs.Range()}, nil

	default:
		return nil, &SortError{
			Msg: fmt.Sprintf("Tried to apply a non-function: %s", fn.String()),
		}
	}
}

// MustApply is like NewApply but panics on error.
// Matches Python's Apply constructor which raises SortError on bad input.
func MustApply(fn Expr, terms ...Expr) *Apply {
	a, err := NewApply(fn, terms...)
	if err != nil {
		panic(fmt.Sprintf("MustApply: %v", err))
	}
	return a
}

// TryApply is like NewApply but falls back to NewApplyUnchecked on error
// instead of returning an error. Use this in contexts where sort mismatches
// are expected (e.g., formulas containing schema parameter sorts).
func TryApply(fn Expr, terms ...Expr) *Apply {
	a, err := NewApply(fn, terms...)
	if err != nil {
		return NewApplyUnchecked(fn, terms...)
	}
	return a
}

// CloneApplyTerms creates a new Apply with the same Func and aSort but new terms.
// Skips sort validation, matching Python's Apply.clone(args) behavior.
func CloneApplyTerms(orig *Apply, newTerms []Expr) *Apply {
	cp := make([]Expr, len(newTerms))
	copy(cp, newTerms)
	return &Apply{Func: orig.Func, Terms: cp, aSort: orig.aSort}
}

// NewApplyUnchecked creates an Apply without sort validation, computing
// aSort from the function's sort. Matches Python's Apply construction
// in contexts where sort mismatches are expected (e.g., schema expansion).
func NewApplyUnchecked(fn Expr, terms ...Expr) *Apply {
	cp := make([]Expr, len(terms))
	copy(cp, terms)
	var resultSort Sort
	if fnSort := fn.NodeSort(); fnSort != nil {
		if fs, ok := fnSort.(*LogicFunctionSort); ok {
			resultSort = fs.Range()
		} else {
			resultSort = TopS
		}
	} else {
		resultSort = TopS
	}
	return &Apply{Func: fn, Terms: cp, aSort: resultSort}
}

func (a *Apply) NodeSort() Sort { return a.aSort }

func (a *Apply) Children() []Expr {
	// Returns Terms only — matches Python's Apply.args property
	// (ivy_logic.py:281: Apply.args = property(lambda self: self.terms)).
	// Code that needs to walk the Func must access a.Func explicitly.
	cp := make([]Expr, len(a.Terms))
	copy(cp, a.Terms)
	return cp
}

// String matches Python ivy_logic.py:1444-1446 which monkey-patches
// lg.Apply.__str__ = pretty_fmla. PrettyFmla calls
// drop_annotations(False, set()).ugly(0), which uses app_ugly for Apply
// (handles infix operators and precedence).
func (a *Apply) String() string { return PrettyFmla(a) }

// Repr returns the sort-qualified representation, using Repr() on children.
// Matches Python ast.Atom.__repr__ which calls str() on args (which for
// ast.Variable invokes __repr__ including sort qualifiers).
func (a *Apply) Repr() string {
	if len(a.Terms) == 0 {
		return ReprExpr(a.Func)
	}
	parts := make([]string, len(a.Terms))
	for i, t := range a.Terms {
		parts[i] = ReprExpr(t)
	}
	return fmt.Sprintf("%s(%s)", ReprExpr(a.Func), strings.Join(parts, ", "))
}

func (a *Apply) Equal(n Expr) bool {
	o, ok := n.(*Apply)
	if !ok {
		return false
	}
	if !a.Func.Equal(o.Func) || len(a.Terms) != len(o.Terms) {
		return false
	}
	for i := range a.Terms {
		if !a.Terms[i].Equal(o.Terms[i]) {
			return false
		}
	}
	return true
}

func isUpper(b byte) bool {
	return b >= 'A' && b <= 'Z'
}
