package ssc

import (
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
)

func TestResimulationRejectsNonSeparatingWitness(t *testing.T) {
	g := gia.Gia_ManStart(8)
	gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCi(g)
	// The third input splits away, but the alleged witness does not distinguish
	// the actual pair 1 and 2; merely observing any split would miss the error.
	p := &Ssc_Man_t{Aig: g, Original: g, Classes: Ssc_GiaClassesInit(g, []bool{false, false, false}), Pattern: []bool{false, false, true}}
	defer func() {
		if recover() == nil {
			t.Fatal("non-separating witness accepted")
		}
	}()
	Ssc_GiaResimulateOneClass(p, 0, 2)
}

func TestResimulationAcceptsSeparatingWitness(t *testing.T) {
	g := gia.Gia_ManStart(8)
	gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCi(g)
	p := &Ssc_Man_t{Aig: g, Original: g, Classes: Ssc_GiaClassesInit(g, []bool{false, false}), Pattern: []bool{true, false}}
	if got := Ssc_GiaResimulateOneClass(p, 0, 1); got != 1 || p.Classes.Repr[1] == 0 {
		t.Fatalf("split=%d, representative=%d", got, p.Classes.Repr[1])
	}
}

func TestResimulationUsesMappedFraigValues(t *testing.T) {
	g := gia.Gia_ManStart(8)
	for i := 0; i < 3; i++ {
		gia.Gia_ManAppendCi(g)
	}
	fraig := newBuilder(3, 8)
	// The first original input was substituted with the third FRAIG input.
	// The witness distinguishes the mapped pair but not the original pair.
	p := &Ssc_Man_t{Original: g, Aig: fraig.graph, mapping: []int32{0, 6, 4, 6},
		Classes: &Classes{Groups: [][]int32{{0}, {1, 2, 3}}, Repr: []int32{-1, -1, 1, 1}, Phase: make([]bool, 4)}, Pattern: []bool{false, false, true}}
	if got := Ssc_GiaResimulateOneClass(p, 1, 2); got != 1 || p.Classes.Repr[2] == 1 {
		t.Fatalf("split=%d, representative=%d", got, p.Classes.Repr[2])
	}
}
