package goivy

import (
	"fmt"
)

// Definition represents a formula of the form lhs = rhs (or lhs <-> rhs).
// In Python this is ivy_logic.Definition.
type LogicDefinition struct {
	Base
	Lhs Expr
	Rhs Expr
}

func NewDefinition(lhs, rhs Expr) *LogicDefinition {
	return &LogicDefinition{Lhs: lhs, Rhs: rhs}
}

func (d *LogicDefinition) NodeSort() Sort { return Boolean }

func (d *LogicDefinition) Children() []Expr {
	return []Expr{d.Lhs, d.Rhs}
}

func (d *LogicDefinition) String() string {
	return fmt.Sprintf("%s = %s", d.Lhs, d.Rhs)
}

func (d *LogicDefinition) Equal(n Expr) bool {
	o, ok := n.(*LogicDefinition)
	if !ok {
		return false
	}
	return d.Lhs.Equal(o.Lhs) && d.Rhs.Equal(o.Rhs)
}

// Defines returns the defining symbol (the Func of the LHS if it's an Apply,
// or the LHS itself if it's a Symbol).
func (d *LogicDefinition) Defines() Expr {
	if app, ok := d.Lhs.(*Apply); ok {
		return app.Func
	}
	return d.Lhs
}

// DefinitionSchema is a parametrized definition.
type LogicDefinitionSchema struct {
	LogicDefinition
}

func NewDefinitionSchema(lhs, rhs Expr) *LogicDefinitionSchema {
	return &LogicDefinitionSchema{LogicDefinition{Lhs: lhs, Rhs: rhs}}
}
