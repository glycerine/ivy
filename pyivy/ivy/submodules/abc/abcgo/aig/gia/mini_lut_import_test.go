package gia

import (
	"testing"

	"github.com/glycerine/abcgo/aig/miniaig"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestMiniLutImport(t *testing.T) {
	p := miniaig.Mini_LutStart(2)
	a, b := miniaig.Mini_LutCreatePi(p), miniaig.Mini_LutCreatePi(p)
	x := miniaig.Mini_LutCreateNode(p, 2, []int{a, b}, []uint32{0x66666666})
	miniaig.Mini_LutCreatePo(p, x)
	var copies *Vec_Int_t
	g := Gia_ManFromMiniLut(p, &copies)
	if vec.Vec_IntSize(copies) != miniaig.Mini_LutNodeNum(p) {
		t.Fatal("incorrect copy map")
	}
	for i := 0; i < 4; i++ {
		out, err := g.Evaluate([]bool{i&1 != 0, i&2 != 0})
		if err != nil || len(out) != 1 || out[0] != (i == 1 || i == 2) {
			t.Fatalf("input %d: %v %v", i, out, err)
		}
	}
}
