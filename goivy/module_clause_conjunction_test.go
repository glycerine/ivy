package goivy

import (
	"slices"
	"testing"
)

func TestClauseConjunctionBoundaries(t *testing.T) {
	a, b := moduleMkConst("a"), moduleMkConst("b")
	v := mkBoolVar("V")
	nested := &ForAll{Variables: []*LogicVariable{v}, Body: a}
	not := &LogicNot{Body: &LogicExists{Variables: []*LogicVariable{v}, Body: b}}
	input := []Expr{&ForAll{Variables: []*LogicVariable{v}, Body: &LogicAnd{Terms: []Expr{a, nested, &LogicAnd{Terms: []Expr{b, True}}}}}, not}
	cl := NewClauses(input, nil, nil)
	if len(cl.Fmlas) != 4 || !slices.Equal(cl.Fmlas[:3], []Expr{a, nested, b}) {
		t.Fatalf("flattened formulas %v", cl.Fmlas)
	}
	neg, ok := cl.Fmlas[3].(*LogicNot)
	if !ok || neg.Body != b {
		t.Fatalf("negated existential %v", cl.Fmlas[3])
	}
	input[0] = b
	if cl.Fmlas[0] != a {
		t.Fatal("NewClauses borrowed the caller's formula storage")
	}
	empty := NewClauses([]Expr{True, &LogicAnd{Terms: []Expr{True, True}}}, nil, nil)
	if empty.Fmlas != nil || empty.Defs != nil {
		t.Fatal("empty clause set", empty)
	}
	empty.DefIdx[Key(a)] = 1
	first := &IvyDefinition{Lhs: a, Rhs: b}
	second := &IvyDefinition{Lhs: a, Rhs: a}
	for _, join := range []func(*Clauses, *Clauses) *Clauses{
		func(x, y *Clauses) *Clauses { return AndClauses(x, y).(*Clauses) },
		func(x, y *Clauses) *Clauses { return AndClausesTyped(x, y) },
	} {
		annotation := &EmptyAnnotation{}
		left := NewClauses([]Expr{a}, []*IvyDefinition{first}, annotation)
		right := NewClauses([]Expr{b}, []*IvyDefinition{second}, nil)
		got := join(left, right)
		if got.Annot != annotation || !slices.Equal(got.Fmlas, []Expr{a, b}) || !slices.Equal(got.Defs, []*IvyDefinition{first, second}) || got.DefIdx[Key(a)] != 1 {
			t.Fatal("combined clauses", got)
		}
		left.Fmlas[0] = b
		right.Fmlas[0] = a
		left.Defs[0] = second
		right.Defs[0] = first
		if !slices.Equal(got.Fmlas, []Expr{a, b}) || !slices.Equal(got.Defs, []*IvyDefinition{first, second}) {
			t.Fatal("joined clauses shared mutable input slices")
		}
	}
}
func BenchmarkClauseConjunction(b *testing.B) {
	a := moduleMkConst("a")
	flat := make([]Expr, 128)
	for i := range flat {
		flat[i] = a
	}
	var nested Expr = a
	for i := 0; i < 64; i++ {
		nested = &LogicAnd{Terms: []Expr{a, nested}}
	}
	groups := make([]*Clauses, 8)
	for i := range groups {
		groups[i] = NewClauses(flat[:16], nil, nil)
	}
	for _, tc := range []struct {
		name  string
		count int
		make  func() *Clauses
	}{
		{"Flat128", 128, func() *Clauses { return NewClauses(flat, nil, nil) }},
		{"Nested65", 65, func() *Clauses { return NewClauses([]Expr{nested}, nil, nil) }},
		{"Join8x16", 128, func() *Clauses { return AndClausesTyped(groups...) }},
	} {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				if c := tc.make(); len(c.Fmlas) != tc.count {
					b.Fatal(len(c.Fmlas))
				}
			}
		})
	}
}
