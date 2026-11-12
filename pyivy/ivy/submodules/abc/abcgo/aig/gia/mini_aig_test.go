package gia

import (
	"testing"

	"github.com/glycerine/abcgo/aig/miniaig"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestMiniAigRoundTripAndCopies(t *testing.T) {
	m := miniaig.Mini_AigStart()
	a, b := miniaig.Mini_AigCreatePi(m), miniaig.Mini_AigCreatePi(m)
	gate := miniaig.Mini_AigAnd(m, a, b^1)
	miniaig.Mini_AigCreatePo(m, gate)
	var copies *Vec_Int_t
	g := Gia_ManFromMiniAig(m, &copies)
	if vec.Vec_IntSize(copies) != miniaig.Mini_AigNodeNum(m) {
		t.Fatal("missing copy mappings")
	}
	round := Gia_ManFromMiniAig(Gia_ManToMiniAig(g), nil)
	for i := 0; i < 4; i++ {
		got, err := round.Evaluate([]bool{i&1 != 0, i&2 != 0})
		if err != nil || len(got) != 1 || got[0] != (i == 1) {
			t.Fatalf("input %d: %v %v", i, got, err)
		}
	}
}
