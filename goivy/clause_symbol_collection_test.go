package goivy

import (
	"fmt"
	"slices"
	"testing"
)

func TestClauseSymbolOrderAndLastRepresentative(t *testing.T) {
	first := NewConst("x", Boolean)
	second := NewConst("x", Boolean)
	y := NewConst("y", Boolean)
	z := NewConst("z", Boolean)
	shared := &Eq{T1: first, T2: y}
	input := &Clauses{Fmlas: []Expr{shared, second, shared}, Defs: []*IvyDefinition{{Lhs: z, Rhs: second}}}
	for _, set := range []*InsMap[NodeKey, Expr]{input.Symbols(), UsedSymbolsClauses(input), UsedSymbolsClausesOrdered(input)} {
		var values []Expr
		for _, sym := range set.All() {
			values = append(values, sym)
		}
		if !slices.Equal(values, []Expr{second, y, z}) {
			t.Fatalf("wrong symbol order or representative: %v", values)
		}
	}
	if UsedSymbolsClauses(nil).Len() != 0 || UsedSymbolsClausesOrdered(nil).Len() != 0 {
		t.Fatal("nil clauses contained symbols")
	}
	input.Defs = nil
	if input.Symbols().Get(Key(first)) != first {
		t.Fatal("repeated subtree must supply its final symbol representative")
	}
}

func BenchmarkClauseSymbolCollection(b *testing.B) {
	for _, count := range []int{32, 128} {
		common := NewConst("common", Boolean)
		c := &Clauses{}
		for i := 0; i < count; i++ {
			unique := NewConst(fmt.Sprintf("p%d", i), Boolean)
			c.Fmlas = append(c.Fmlas, &Eq{T1: unique, T2: common})
		}
		b.Run(fmt.Sprintf("Distinct%d", count), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				if UsedSymbolsClauses(c).Len() != count+1 {
					b.Fatal("lost a symbol")
				}
			}
		})
	}
	for _, depth := range []int{0, 8, 14} {
		x := NewConst("x", Boolean)
		var tree Expr = &Eq{T1: x, T2: NewConst("y", Boolean)}
		for i := 0; i < depth; i++ {
			tree = &LogicImplies{T1: tree, T2: &LogicNot{Body: tree}}
		}
		c := &Clauses{}
		for i := 0; i < 32; i++ {
			c.Fmlas = append(c.Fmlas, tree)
		}
		b.Run(fmt.Sprintf("Depth%d", depth), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				if UsedSymbolsClauses(c).Len() != 2 {
					b.Fatal("expected two symbols")
				}
			}
		})
	}
}
