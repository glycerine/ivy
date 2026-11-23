package goivy

import (
	"strings"
	"testing"
)

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

// Some sorts supply their own structural spelling. An embedded Sort keeps this
// test independent of the rest of the expression interface.
type observedKeySort struct {
	Sort
	calls int
}

func (s *observedKeySort) Sexp() NodeKey {
	s.calls++
	return "(custom sort)"
}

func TestSymbolKeysWithNestedAndWideSorts(t *testing.T) {
	leaf := &UninterpretedSort{Name: "a)\x00"}
	custom := &observedKeySort{Sort: Boolean}
	inner := &LogicFunctionSort{Sorts: []Sort{leaf, custom, Boolean}}
	outer := &LogicFunctionSort{Sorts: []Sort{inner, &TopSort{Name: "target"}}}
	const innerKey = "(FunctionSort sorts:[(UninterpretedSort name:a)\x00) (custom sort) (BooleanSort)])"
	const outerKey = "(FunctionSort sorts:[" + innerKey + " (TopSort name:target)])"
	symbol := NewConst("p", outer)
	variable := &LogicVariable{Name: "V", VSort: outer}
	if symbol.Sexp() != "(Symbol name:p sort:"+outerKey+")" || variable.Sexp() != "(Variable name:V sort:"+outerKey+")" || outer.Sexp() != outerKey {
		t.Fatal("nested sort spelling changed")
	}
	if custom.calls != 3 {
		t.Fatalf("custom sort called %d times, want 3", custom.calls)
	}
	leaf.Name = "changed"
	if got := symbol.Sexp(); got != "(Symbol name:p sort:"+outerKey+")" {
		t.Fatalf("mutation changed the constant's captured key: %q", got)
	}
	if got := variable.Sexp(); !strings.Contains(string(got), "(UninterpretedSort name:changed)") {
		t.Fatalf("variable key missed the new sort name: %q", got)
	}
	wide := &LogicFunctionSort{}
	var parts []string
	for i := 0; i < 40; i++ {
		wide.Sorts = append(wide.Sorts, Boolean)
		parts = append(parts, "(BooleanSort)")
	}
	want := "(FunctionSort sorts:[" + strings.Join(parts, " ") + "])"
	if got := wide.Sexp(); string(got) != want {
		t.Fatalf("wide sort key = %q", got)
	}
	if got := (&LogicFunctionSort{}).Sexp(); got != "(FunctionSort sorts:[])" {
		t.Fatalf("empty function sort = %q", got)
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
		{"Nested", &LogicFunctionSort{Sorts: []Sort{&LogicFunctionSort{Sorts: []Sort{individual, Boolean}}, individual}}},
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
