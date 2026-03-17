package logic

import "fmt"

// Definition represents a formula of the form lhs = rhs (or lhs <-> rhs).
// In Python this is ivy_logic.Definition.
type Definition struct {
	Lhs Node
	Rhs Node
}

func NewDefinition(lhs, rhs Node) *Definition {
	return &Definition{Lhs: lhs, Rhs: rhs}
}

func (d *Definition) NodeSort() Sort { return Boolean }

func (d *Definition) Children() []Node {
	return []Node{d.Lhs, d.Rhs}
}

func (d *Definition) String() string {
	return fmt.Sprintf("%s = %s", d.Lhs, d.Rhs)
}

func (d *Definition) Equal(n Node) bool {
	o, ok := n.(*Definition)
	if !ok {
		return false
	}
	return d.Lhs.Equal(o.Lhs) && d.Rhs.Equal(o.Rhs)
}

// Defines returns the defining symbol (the Func of the LHS if it's an Apply,
// or the LHS itself if it's a Const).
func (d *Definition) Defines() Node {
	if app, ok := d.Lhs.(*Apply); ok {
		return app.Func
	}
	return d.Lhs
}

// DefinitionSchema is a parametrized definition.
type DefinitionSchema struct {
	Definition
}

func NewDefinitionSchema(lhs, rhs Node) *DefinitionSchema {
	return &DefinitionSchema{Definition{Lhs: lhs, Rhs: rhs}}
}
