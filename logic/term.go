package logic

import (
	"fmt"
	"strings"
)

// Variable represents a variable. Name must start with uppercase.
type Variable struct {
	Name  string
	VSort Sort
}

func NewVariable(name string, sort Sort) (*Variable, error) {
	if len(name) == 0 || !isUpper(name[0]) {
		return nil, &IvyError{Msg: fmt.Sprintf("Bad variable name: %q", name)}
	}
	return &Variable{Name: name, VSort: sort}, nil
}

func (v *Variable) NodeSort() Sort    { return v.VSort }
func (v *Variable) Children() []Node  { return nil }
func (v *Variable) String() string    { return v.Name }

func (v *Variable) Equal(n Node) bool {
	if o, ok := n.(*Variable); ok {
		return v.Name == o.Name && v.VSort.Equal(o.VSort)
	}
	return false
}

// Call applies the variable as a function. Returns self if no args.
func (v *Variable) Call(terms ...Node) (Node, error) {
	if len(terms) == 0 {
		return v, nil
	}
	return NewApply(v, terms...)
}

// Symbol represents a constant symbol.
type Symbol struct {
	Name  string
	CSort Sort
}

func NewSymbol(name string, sort Sort) *Symbol {
	return &Symbol{Name: name, CSort: sort}
}

func (c *Symbol) NodeSort() Sort    { return c.CSort }
func (c *Symbol) Children() []Node  { return nil }
func (c *Symbol) String() string    { return c.Name }

func (c *Symbol) Equal(n Node) bool {
	if o, ok := n.(*Symbol); ok {
		return c.Name == o.Name && c.CSort.Equal(o.CSort)
	}
	return false
}

// Call applies the constant as a function.
// Matches Python: Symbol.__call__ = lambda self,*args: App(self,*args)
//   if len(args) > 0 or isinstance(self.sort, FunctionSort) else self
// If zero args and CSort is FunctionSort, creates Apply(c) (nullary application).
// If zero args and CSort is NOT FunctionSort, returns self.
func (c *Symbol) Call(terms ...Node) (Node, error) {
	if len(terms) == 0 {
		if _, isFS := c.CSort.(*FunctionSort); isFS {
			return NewApply(c) // nullary application
		}
		return c, nil
	}
	return NewApply(c, terms...)
}

// Apply represents function application.
type Apply struct {
	Func  Node
	Terms []Node
	aSort Sort // cached sort
}

func NewApply(fn Node, terms ...Node) (*Apply, error) {
	fnSort := fn.NodeSort()

	switch fs := fnSort.(type) {
	case *TopSort:
		// TopSort: accept anything, result is TopS
		cp := make([]Node, len(terms))
		copy(cp, terms)
		return &Apply{Func: fn, Terms: cp, aSort: TopS}, nil

	case *FunctionSort:
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
			if tSort.Equal(dSort) {
				continue
			}
			_, tIsTop := tSort.(*TopSort)
			_, dIsTop := dSort.(*TopSort)
			if tIsTop || dIsTop {
				continue
			}
			return nil, reportBadSort(fn, i, dSort, tSort)
		}
		cp := make([]Node, len(terms))
		copy(cp, terms)
		return &Apply{Func: fn, Terms: cp, aSort: fs.Range()}, nil

	default:
		return nil, &SortError{
			Msg: fmt.Sprintf("Tried to apply a non-function: %s", fn.String()),
		}
	}
}

func (a *Apply) NodeSort() Sort { return a.aSort }

func (a *Apply) Children() []Node {
	// Returns Terms only — matches Python's Apply.args property
	// (ivy_logic.py:281: Apply.args = property(lambda self: self.terms)).
	// Code that needs to walk the Func must access a.Func explicitly.
	cp := make([]Node, len(a.Terms))
	copy(cp, a.Terms)
	return cp
}

func (a *Apply) String() string {
	if len(a.Terms) == 0 {
		return a.Func.String()
	}
	parts := make([]string, len(a.Terms))
	for i, t := range a.Terms {
		parts[i] = t.String()
	}
	return fmt.Sprintf("%s(%s)", a.Func.String(), strings.Join(parts, ","))
}

func (a *Apply) Equal(n Node) bool {
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
