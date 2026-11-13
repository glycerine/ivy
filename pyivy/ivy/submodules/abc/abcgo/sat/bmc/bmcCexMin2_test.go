package bmc

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestAnnotateUnrollingAndUnate(t *testing.T) {
	p := gia.Gia_ManStart(12)
	x, y := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
	gate := gia.Gia_ManAppendAnd(p, x, y)
	gia.Gia_ManAppendCo(p, gate)
	c := util.Abc_CexAlloc(0, 2, 1)
	util.Abc_InfoSetBit(c.PData, 0)
	util.Abc_InfoSetBit(c.PData, 1)
	if Gia_ManAnnotateUnrolling(p, c, 0) != 1 {
		t.Fatal("valid cex rejected")
	}
	defer unrollingAnnotations.Delete(p)
	if got := Gia_ManGetTwo(p, 0, gia.Gia_ManObj(p, gate>>1)); got != 3 {
		t.Fatalf("gate annotation %d", got)
	}
	q := Gia_ManCreateUnate(p, c, 0, 0, 1)
	if len(q.InputLiterals()) != 2 {
		t.Fatalf("inputs: %v", q.InputLiterals())
	}
	for bits := 0; bits < 4; bits++ {
		in := []bool{bits&1 != 0, bits&2 != 0}
		out, err := q.Evaluate(in)
		if err != nil || out[0] != (in[0] && in[1]) {
			t.Fatalf("in %v out %v error %v", in, out, err)
		}
	}
}

func TestPackedAnnotationAcrossWordBoundaries(t *testing.T) {
	v := vec.Vec_IntStart(4)
	Abc_InfoSet2Bits(v, 2, 0, 15, 3)
	Abc_InfoAdd2Bits(v, 2, 0, 16, 2)
	Abc_InfoAdd2Bits(v, 2, 1, 0, 1)
	if Abc_InfoGet2Bits(v, 2, 0, 15) != 3 || Abc_InfoGet2Bits(v, 2, 0, 16) != 2 || Abc_InfoGet2Bits(v, 2, 1, 0) != 1 {
		t.Fatal("bad packed annotation")
	}
}

func TestAnnotateSequentialBoundary(t *testing.T) {
	p := gia.Gia_ManStart(12)
	pi, ro := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
	gia.Gia_ManAppendCo(p, ro)
	gia.Gia_ManAppendCo(p, pi)
	gia.Gia_ManSetRegNum(p, 1)
	c := util.Abc_CexAlloc(1, 1, 2)
	c.IFrame = 1
	util.Abc_InfoSetBit(c.PData, 1)
	if Gia_ManAnnotateUnrolling(p, c, 0) != 1 {
		t.Fatal("valid sequential cex rejected")
	}
	defer unrollingAnnotations.Delete(p)
	for _, check := range []struct {
		frame  int32
		object *gia.Gia_Obj_t
	}{
		{1, gia.Gia_ManRo(p, 0)}, {0, gia.Gia_ManRi(p, 0)}, {0, gia.Gia_ManCi(p, 0)},
	} {
		if value := Gia_ManGetTwo(p, check.frame, check.object); value != 3 {
			t.Fatalf("frame %d annotation %d", check.frame, value)
		}
	}
	q := Gia_ManCreateUnate(p, c, 1, -1, 0)
	for _, input := range []bool{false, true} {
		out, err := q.Evaluate([]bool{input})
		if err != nil || out[0] != input {
			t.Fatalf("input %v out %v err %v", input, out, err)
		}
	}
}
