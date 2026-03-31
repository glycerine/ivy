package logic

import (
	"fmt"

	"github.com/glycerine/ivy/goivy/ast"
)

// Definition represents a formula of the form lhs = rhs (or lhs <-> rhs).
// In Python this is ivy_logic.Definition.
type Definition struct {
	ast.Base
	Lhs Expr
	Rhs Expr
}

func NewDefinition(lhs, rhs Expr) *Definition {
	return &Definition{Lhs: lhs, Rhs: rhs}
}

func (d *Definition) NodeSort() Sort { return Boolean }

func (d *Definition) Children() []Expr {
	return []Expr{d.Lhs, d.Rhs}
}

func (d *Definition) String() string {
	return fmt.Sprintf("%s = %s", d.Lhs, d.Rhs)
}

func (d *Definition) Equal(n Expr) bool {
	o, ok := n.(*Definition)
	if !ok {
		return false
	}
	return d.Lhs.Equal(o.Lhs) && d.Rhs.Equal(o.Rhs)
}

// Defines returns the defining symbol (the Func of the LHS if it's an Apply,
// or the LHS itself if it's a Symbol).
func (d *Definition) Defines() Expr {
	if app, ok := d.Lhs.(*Apply); ok {
		return app.Func
	}
	return d.Lhs
}

// DefinitionSchema is a parametrized definition.
type DefinitionSchema struct {
	Definition
}

func NewDefinitionSchema(lhs, rhs Expr) *DefinitionSchema {
	return &DefinitionSchema{Definition{Lhs: lhs, Rhs: rhs}}
}
