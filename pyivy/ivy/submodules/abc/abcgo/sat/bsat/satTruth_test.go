package bsat

import (
	"math/rand"
	"reflect"
	"testing"
)

func TestTruthVariablePatterns(t *testing.T) {
	p := Tru_ManAlloc(8)
	for v := int32(0); v < 8; v++ {
		table := Tru_ManVar(p, v)
		for point := 0; point < 256; point++ {
			got := (table[point/64] >> uint(point%64)) & 1
			want := uint64((point >> uint(v)) & 1)
			if got != want {
				t.Fatal(v, point, got, want)
			}
		}
	}
	if Tru_ManInsert(p, make([]word, p.nWords)) != 0 {
		t.Fatal("zero")
	}
	ones := make([]word, p.nWords)
	Tru_ManFill(ones, p.nWords)
	if Tru_ManInsert(p, ones) != 1 {
		t.Fatal("one")
	}
}
func TestTruthInterningComplementsAndResize(t *testing.T) {
	p := Tru_ManAlloc(6)
	rng := rand.New(rand.NewSource(912))
	var handles []int32
	var tables [][]word
	for i := 0; i < 300; i++ {
		table := []word{rng.Uint64() &^ 1}
		original := append([]word(nil), table...)
		h := Tru_ManInsert(p, table)
		if h&1 != 0 || !reflect.DeepEqual(table, original) || Tru_ManInsert(p, table) != h {
			t.Fatal("interning")
		}
		not := []word{^table[0]}
		if Tru_ManInsert(p, not) != h^1 || not[0] != ^table[0] {
			t.Fatal("complement")
		}
		handles = append(handles, h)
		tables = append(tables, table)
	}
	Tru_ManResize(p)
	for i, h := range handles {
		if !reflect.DeepEqual(Tru_ManFunc(p, h), tables[i]) || Tru_ManInsert(p, tables[i]) != h {
			t.Fatal("rehash lost handle")
		}
	}
}
func TestTruthBitOperations(t *testing.T) {
	a := []word{0x55, 0x0f}
	b := []word{0xaa, 0x33}
	out := make([]word, 2)
	Tru_ManCopy(out, a, 2)
	Tru_ManAndNotCond(out, b, 2, 1)
	if !reflect.DeepEqual(out, []word{0x55, 0x0c}) {
		t.Fatal(out)
	}
	Tru_ManOrNotCond(out, b, 2, 0)
	if !reflect.DeepEqual(out, []word{0xff, 0x3f}) {
		t.Fatal(out)
	}
	Tru_ManCopyNotCond(out, a, 2, 1)
	if Tru_ManEqual(out, []word{^word(0x55), ^word(0x0f)}, 2) == 0 {
		t.Fatal(out)
	}
	Tru_ManClear(out, 2)
	if Tru_ManEqual0(out, 2) == 0 {
		t.Fatal(out)
	}
	Tru_ManFill(out, 2)
	if Tru_ManEqual1(out, 2) == 0 {
		t.Fatal(out)
	}
}
