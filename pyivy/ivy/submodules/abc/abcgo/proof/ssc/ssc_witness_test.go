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
	p := &Ssc_Man_t{Aig: g, Classes: Ssc_GiaClassesInit(g, []bool{false, false, false}), Pattern: []bool{false, false, true}}
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
	p := &Ssc_Man_t{Aig: g, Classes: Ssc_GiaClassesInit(g, []bool{false, false}), Pattern: []bool{true, false}}
	if got := Ssc_GiaResimulateOneClass(p, 0, 1); got != 1 || p.Classes.Repr[1] == 0 {
		t.Fatalf("split=%d, representative=%d", got, p.Classes.Repr[1])
	}
}
