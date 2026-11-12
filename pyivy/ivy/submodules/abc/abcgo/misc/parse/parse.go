// Package parse parses ABC Boolean formulas independently of their graph backend.
package parse

import "fmt"

type Operator uint8

const (
	Constant Operator = iota
	Variable
	Not
	And
	Xor
	Or
	Equivalent
	Implies
	ImpliedBy
	Shift
)

// Expression is an immutable formula tree. Shift lowers all BDD variable indices
// by one rank (the rank width is supplied when compiling).
type Expression struct {
	Op          Operator
	Value       bool
	Index       int
	Left, Right *Expression
}
type Operations[T any] struct {
	Constant     func(bool) T
	Variable     func(int) T
	Not          func(T) T
	And, Or, Xor func(T, T) T
	Shift        func(T, int) (T, error)
}

func Compile[T any](e *Expression, ops Operations[T], rankWidth int) (T, error) {
	var zero T
	if e == nil {
		return zero, fmt.Errorf("parse: nil expression")
	}
	switch e.Op {
	case Constant:
		if ops.Constant == nil {
			return zero, fmt.Errorf("parse: missing constant backend")
		}
		return ops.Constant(e.Value), nil
	case Variable:
		if ops.Variable == nil {
			return zero, fmt.Errorf("parse: missing variable backend")
		}
		return ops.Variable(e.Index), nil
	}
	a, err := Compile(e.Left, ops, rankWidth)
	if err != nil {
		return zero, err
	}
	if e.Op == Not {
		if ops.Not == nil {
			return zero, fmt.Errorf("parse: missing negation backend")
		}
		return ops.Not(a), nil
	}
	if e.Op == Shift {
		if ops.Shift == nil {
			return zero, fmt.Errorf("parse: backend does not support rank shifts")
		}
		return ops.Shift(a, -rankWidth)
	}
	b, err := Compile(e.Right, ops, rankWidth)
	if err != nil {
		return zero, err
	}
	switch e.Op {
	case And:
		if ops.And != nil {
			return ops.And(a, b), nil
		}
	case Or:
		if ops.Or != nil {
			return ops.Or(a, b), nil
		}
	case Xor:
		if ops.Xor != nil {
			return ops.Xor(a, b), nil
		}
	case Equivalent:
		if ops.Xor != nil && ops.Not != nil {
			return ops.Not(ops.Xor(a, b)), nil
		}
	case Implies:
		if ops.Or != nil && ops.Not != nil {
			return ops.Or(ops.Not(a), b), nil
		}
	case ImpliedBy:
		if ops.Or != nil && ops.Not != nil {
			return ops.Or(a, ops.Not(b)), nil
		}
	}
	return zero, fmt.Errorf("parse: missing backend for operator %d", e.Op)
}

// Evaluate computes an ordinary Boolean assignment; ranked expressions require
// a graph backend capable of permuting variable indices.
func (e *Expression) Evaluate(values []bool) (bool, error) {
	var check func(*Expression) error
	check = func(n *Expression) error {
		if n == nil {
			return nil
		}
		if n.Op == Variable && (n.Index < 0 || n.Index >= len(values)) {
			return fmt.Errorf("parse: assignment lacks variable %d", n.Index)
		}
		if err := check(n.Left); err != nil {
			return err
		}
		return check(n.Right)
	}
	if err := check(e); err != nil {
		return false, err
	}
	return Compile(e, Operations[bool]{Constant: func(v bool) bool { return v }, Variable: func(i int) bool { return values[i] }, Not: func(v bool) bool { return !v }, And: func(a, b bool) bool { return a && b }, Or: func(a, b bool) bool { return a || b }, Xor: func(a, b bool) bool { return a != b }}, 0)
}
