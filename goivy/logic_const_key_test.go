package goivy

import "testing"

func TestConstantKeyIsCapturedAtConstruction(t *testing.T) {
	sort := &UninterpretedSort{Name: "a"}
	term := NewConst("x", sort)
	sort.Name = "b"
	term.Name = "y"
	if got := term.Sexp(); got != "(Symbol name:x sort:(UninterpretedSort name:a))" {
		t.Fatalf("captured key = %q", got)
	}
	if got := NewConst("%v\x00)", nil).Sexp(); got != "(Symbol name:%v\x00) sort:nil)" {
		t.Fatalf("key with nil sort = %q", got)
	}
}

func BenchmarkConstantStructuralKey(b *testing.B) {
	individual := &UninterpretedSort{Name: "individual"}
	cases := []struct {
		name string
		sort Sort
	}{
		{"Boolean", Boolean},
		{"Uninterpreted", individual},
		{"BinaryPredicate", &LogicFunctionSort{Sorts: []Sort{individual, individual, Boolean}}},
	}
	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				_ = NewConst("new_trans", tc.sort)
			}
		})
	}
}
