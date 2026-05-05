package goivy

import (
	"strings"
)

// Some represents "some X:t. phi" — an indefinite description.
type Some struct {
	Base
	Params  []Expr // bound variables
	Fmla    Expr   // formula/constraint
	IfVal   Expr   // optional: value if exists (may be nil)
	ElseVal Expr   // optional: value if not exists (may be nil)
}

func NewSome(params []Expr, fmla Expr) *Some {
	return &Some{Params: params, Fmla: fmla}
}

func NewSomeWithElse(params []Expr, fmla, ifVal, elseVal Expr) *Some {
	return &Some{Params: params, Fmla: fmla, IfVal: ifVal, ElseVal: elseVal}
}

func (s *Some) NodeSort() Sort {
	if len(s.Params) > 0 {
		return s.Params[0].NodeSort()
	}
	return TopS
}

func (s *Some) Children() []Expr {
	result := make([]Expr, 0, len(s.Params)+3)
	result = append(result, s.Params...)
	result = append(result, s.Fmla)
	if s.IfVal != nil {
		result = append(result, s.IfVal)
	}
	if s.ElseVal != nil {
		result = append(result, s.ElseVal)
	}
	return result
}

func (s *Some) String() string {
	var b strings.Builder
	b.WriteString("some ")
	if len(s.Params) > 0 {
		b.WriteString(s.Params[0].String())
	}
	b.WriteString(". ")
	b.WriteString(s.Fmla.String())
	if s.IfVal != nil {
		b.WriteString(" in ")
		b.WriteString(s.IfVal.String())
	}
	if s.ElseVal != nil {
		b.WriteString(" else ")
		b.WriteString(s.ElseVal.String())
	}
	return b.String()
}

func (s *Some) Equal(n Expr) bool {
	o, ok := n.(*Some)
	if !ok {
		return false
	}
	if len(s.Params) != len(o.Params) {
		return false
	}
	for i := range s.Params {
		if !s.Params[i].Equal(o.Params[i]) {
			return false
		}
	}
	if !s.Fmla.Equal(o.Fmla) {
		return false
	}
	if (s.IfVal == nil) != (o.IfVal == nil) {
		return false
	}
	if s.IfVal != nil && !s.IfVal.Equal(o.IfVal) {
		return false
	}
	if (s.ElseVal == nil) != (o.ElseVal == nil) {
		return false
	}
	if s.ElseVal != nil && !s.ElseVal.Equal(o.ElseVal) {
		return false
	}
	return true
}

func (s *Some) BinderVars() []*Variable {
	vars := make([]*Variable, 0, len(s.Params))
	for _, p := range s.Params {
		if v, ok := p.(*Variable); ok {
			vars = append(vars, v)
		}
	}
	return vars
}

func (s *Some) BinderBody() Expr {
	return s.Fmla
}

// CloneBinder clones the Some with new variables and body.
func (s *Some) CloneBinder(vs []Expr, body Expr) *Some {
	result := &Some{Params: vs, Fmla: body}
	if s.IfVal != nil {
		result.IfVal = s.IfVal
	}
	if s.ElseVal != nil {
		result.ElseVal = s.ElseVal
	}
	return result
}

// IvyDefinition is now in the logic package. Re-exported here for backward compatibility.
type IvyDefinition = Definition

func NewIvyDefinition(lhs, rhs Expr) *IvyDefinition {
	return NewDefinition(lhs, rhs)
}

// IvyDefinitionSchema is a parametrized definition.
type IvyDefinitionSchema = DefinitionSchema

func NewIvyDefinitionSchema(lhs, rhs Expr) *IvyDefinitionSchema {
	return NewDefinitionSchema(lhs, rhs)
}

// Let represents "let defs in body".
type Let struct {
	Base
	Defs []Expr
	Body Expr
}

func NewLet(defs []Expr, body Expr) *Let {
	return &Let{Defs: defs, Body: body}
}

func (l *Let) NodeSort() Sort { return l.Body.NodeSort() }

func (l *Let) Children() []Expr {
	result := make([]Expr, 0, len(l.Defs)+1)
	result = append(result, l.Defs...)
	result = append(result, l.Body)
	return result
}

func (l *Let) String() string {
	if len(l.Defs) == 0 {
		return l.Body.String()
	}
	parts := make([]string, len(l.Defs))
	for i, d := range l.Defs {
		parts[i] = d.String()
	}
	return "let " + strings.Join(parts, ", ") + " in " + l.Body.String()
}

func (l *Let) Equal(n Expr) bool {
	o, ok := n.(*Let)
	if !ok {
		return false
	}
	if len(l.Defs) != len(o.Defs) {
		return false
	}
	for i := range l.Defs {
		if !l.Defs[i].Equal(o.Defs[i]) {
			return false
		}
	}
	return l.Body.Equal(o.Body)
}

// Literal represents a positive or negative atomic formula.
// Literals are not formulas — use Not(IvyAtom(...)) for a negated formula.
type Literal struct {
	Base
	Polarity int // 1 = positive, 0 = negative
	Atom     Expr
}

func NewLiteral(polarity int, atom Expr) *Literal {
	return &Literal{Polarity: polarity, Atom: atom}
}

func (l *Literal) NodeSort() Sort { return Boolean }

func (l *Literal) Children() []Expr {
	return []Expr{l.Atom}
}

func (l *Literal) String() string {
	if l.Polarity == 0 {
		return "~" + l.Atom.String()
	}
	return l.Atom.String()
}

func (l *Literal) Equal(n Expr) bool {
	o, ok := n.(*Literal)
	if !ok {
		return false
	}
	return l.Polarity == o.Polarity && l.Atom.Equal(o.Atom)
}

// Invert returns the negation of this literal.
func (l *Literal) Invert() *Literal {
	return &Literal{Polarity: 1 - l.Polarity, Atom: l.Atom}
}

// Predicate is a literal factory (not an AST node).
type Predicate struct {
	Name  string
	Arity int
}

// Call creates a positive literal from the predicate applied to terms.
func (p *Predicate) Call(terms ...Expr) *Literal {
	app := MustApply(NewConst(p.Name, TopS), terms...)
	return NewLiteral(1, app)
}
