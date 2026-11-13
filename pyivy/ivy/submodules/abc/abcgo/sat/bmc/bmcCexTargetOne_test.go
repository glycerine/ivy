package bmc

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestTargetCubeAndInitialState(t *testing.T) {
	g := gia.Gia_ManStart(16)
	pi, ro := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, ro)
	gia.Gia_ManAppendCo(g, pi)
	gia.Gia_ManSetRegNum(g, 1)
	c := util.Abc_CexAlloc(1, 1, 2)
	c.IFrame = 1
	c.PData[0] = 1 << 1
	var initial *vec.Vec_Bit_t
	q := Bmc_GiaGenerateGiaOne(g, c, &initial, 0, 1)
	if vec.Vec_BitEntry(initial, 0) != 0 {
		t.Fatal("wrong initial state")
	}
	for _, x := range []bool{false, true} {
		out, err := q.Evaluate([]bool{x})
		if err != nil || out[0] != x {
			t.Fatalf("%v -> %v (%v)", x, out, err)
		}
	}
}

func TestTargetAllFrameJustifications(t *testing.T) {
	g := gia.Gia_ManStart(16)
	pi, ro := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, ro)
	gia.Gia_ManAppendCo(g, pi)
	gia.Gia_ManSetRegNum(g, 1)
	c := util.Abc_CexAlloc(1, 1, 2)
	c.IFrame = 1
	c.PData[0] = 1 << 1
	var initial *vec.Vec_Bit_t
	q := Bmc_GiaGenerateGiaAllFrames(g, c, &initial, 0, 0)
	if gia.Gia_ManCiNum(q) != 1 || gia.Gia_ManCoNum(q) != 2 {
		t.Fatalf("inputs %d outputs %d", gia.Gia_ManCiNum(q), gia.Gia_ManCoNum(q))
	}
	for _, x := range []bool{false, true} {
		out, err := q.Evaluate([]bool{x})
		if err != nil || !out[0] || out[1] != x {
			t.Fatalf("%v -> %v (%v)", x, out, err)
		}
	}
	q = Bmc_GiaGenerateGiaAllOne(g, c, &initial, 0, 1)
	for _, x := range []bool{false, true} {
		out, err := q.Evaluate([]bool{x})
		if err != nil || out[0] != x {
			t.Fatalf("%v -> %v (%v)", x, out, err)
		}
	}
}
