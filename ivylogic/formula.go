package ivylogic

import (
	"strings"

	lg "github.com/glycerine/goivy/logic"
)

// Some represents "some X:t. phi" — an indefinite description.
type Some struct {
	Params []lg.Node // bound variables
	Fmla   lg.Node   // formula/constraint
	IfVal  lg.Node   // optional: value if exists (may be nil)
	ElseVal lg.Node  // optional: value if not exists (may be nil)
}

func NewSome(params []lg.Node, fmla lg.Node) *Some {
	return &Some{Params: params, Fmla: fmla}
}

func NewSomeWithElse(params []lg.Node, fmla, ifVal, elseVal lg.Node) *Some {
	return &Some{Params: params, Fmla: fmla, IfVal: ifVal, ElseVal: elseVal}
}

func (s *Some) NodeSort() lg.Sort {
	if len(s.Params) > 0 {
		return s.Params[0].NodeSort()
	}
	return lg.TopS
}

func (s *Some) Children() []lg.Node {
	result := make([]lg.Node, 0, len(s.Params)+3)
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

func (s *Some) Equal(n lg.Node) bool {
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

// CloneBinder clones the Some with new variables and body.
func (s *Some) CloneBinder(vs []lg.Node, body lg.Node) *Some {
	result := &Some{Params: vs, Fmla: body}
	if s.IfVal != nil {
		result.IfVal = s.IfVal
	}
	if s.ElseVal != nil {
		result.ElseVal = s.ElseVal
	}
	return result
}

// Definition is now in the logic package. Re-exported here for backward compatibility.
type Definition = lg.Definition

func NewDefinition(lhs, rhs lg.Node) *Definition {
	return lg.NewDefinition(lhs, rhs)
}

// DefinitionSchema is a parametrized definition.
type DefinitionSchema = lg.DefinitionSchema

func NewDefinitionSchema(lhs, rhs lg.Node) *DefinitionSchema {
	return lg.NewDefinitionSchema(lhs, rhs)
}

// Let represents "let defs in body".
type Let struct {
	Defs []lg.Node
	Body lg.Node
}

func NewLet(defs []lg.Node, body lg.Node) *Let {
	return &Let{Defs: defs, Body: body}
}

func (l *Let) NodeSort() lg.Sort { return l.Body.NodeSort() }

func (l *Let) Children() []lg.Node {
	result := make([]lg.Node, 0, len(l.Defs)+1)
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

func (l *Let) Equal(n lg.Node) bool {
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
// Literals are not formulas — use Not(Atom(...)) for a negated formula.
type Literal struct {
	Polarity int // 1 = positive, 0 = negative
	Atom     lg.Node
}

func NewLiteral(polarity int, atom lg.Node) *Literal {
	return &Literal{Polarity: polarity, Atom: atom}
}

func (l *Literal) NodeSort() lg.Sort { return lg.Boolean }

func (l *Literal) Children() []lg.Node {
	return []lg.Node{l.Atom}
}

func (l *Literal) String() string {
	if l.Polarity == 0 {
		return "~" + l.Atom.String()
	}
	return l.Atom.String()
}

func (l *Literal) Equal(n lg.Node) bool {
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
func (p *Predicate) Call(terms ...lg.Node) *Literal {
	app := &lg.Apply{Func: lg.NewSymbol(p.Name, lg.TopS), Terms: terms}
	return NewLiteral(1, app)
}
