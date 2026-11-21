package goivy

import "testing"

// SymbolsIluAst visits shared subexpressions once for each occurrence. This
// fixture exercises fixed-arity children and the stored argument slice of Apply.
func BenchmarkSymbolsIluAstTraversal(b *testing.B) {
	leaf := &Const{Name: "x", CSort: Boolean}
	other := &Const{Name: "y", CSort: Boolean}
	predicateSort, err := NewFunctionSort(Boolean, Boolean, Boolean)
	if err != nil {
		b.Fatal(err)
	}
	predicate := &Const{Name: "f", CSort: predicateSort}
	var tree Expr = &Eq{T1: leaf, T2: other}
	for range 6 {
		tree = &LogicImplies{T1: tree, T2: &LogicNot{Body: tree}}
	}
	for _, tc := range []struct {
		name string
		root Expr
		want int
	}{
		{"FixedArity", tree, 128},
		{"StoredArguments", &Apply{Func: predicate, Terms: []Expr{tree, tree}}, 257},
	} {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				n := 0
				for range SymbolsIluAst(tc.root) {
					n++
				}
				if n != tc.want {
					b.Fatalf("visited %d symbols, want %d", n, tc.want)
				}
			}
		})
	}
}
