package bmc

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

// The C target-state builder appends an AND for a one-literal cube, even when
// one fanin is constant one. AIG structural hashing collapses that gate, so
// the simple exporter must preserve the actual edge mapping rather than
// assuming every source object retains its numeric ID.
func TestAigTargetStatesHandlesSourceUnsimplifiedCubeGate(t *testing.T) {
	g := aig.NewGraph()
	pi, reg := g.NewInput(), g.NewInput()
	g.AddOutput(reg)
	g.AddOutput(pi)
	aig.Aig_ManSetRegNum(g, 1)
	c := util.Abc_CexAlloc(1, 1, 2)
	c.IFrame = 1
	util.Abc_InfoSetBit(c.PData, 1)

	giaGraph := gia.Gia_ManFromAigSimple(g)
	var initial *vec.Vec_Bit_t
	generated := Bmc_GiaGenerateGiaOne(giaGraph, c, &initial, 0, 1)
	if gia.Gia_ManObjNum(generated) != 4 || gia.Gia_ObjIsAnd(gia.Gia_ManObj(generated, 2)) == 0 {
		t.Fatal("fixture no longer produces the source's unsimplified cube gate")
	}
	simple := gia.Gia_ManToAigSimple(generated)
	if mapped := gia.Gia_ObjValue(gia.Gia_ManObj(generated, 2)); mapped != int32(aig.Aig_ManCi(simple, 0)) || mapped == 2*2 {
		t.Fatalf("collapsed gate mapping = %d, want the AIG input edge", mapped)
	}
	result := Bmc_AigTargetStates(g, c, 0, 1, 1, 1, 0, 0)
	if result == nil {
		t.Fatal("target-state conversion failed")
	}
	for _, value := range []bool{false, true} {
		outputs, err := result.Evaluate([]bool{value})
		if err != nil || len(outputs) != 1 || outputs[0] != value {
			t.Fatalf("input %v got %v err %v", value, outputs, err)
		}
	}
}

func TestGiaTargetStatesDriver(t *testing.T) {
	g := gia.Gia_ManStart(16)
	pi, lo := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, lo)
	gia.Gia_ManAppendCo(g, pi)
	gia.Gia_ManSetRegNum(g, 1)
	c := util.Abc_CexAlloc(1, 1, 2)
	c.IFrame = 1
	c.PData[0] = 3
	for _, mode := range []struct{ one, all int32 }{{1, 0}, {0, 0}, {0, 1}} {
		comb := Bmc_GiaTargetStates(g, c, 0, 1, 1, mode.one, mode.all, 0)
		if comb == nil {
			t.Fatal("missing combinational target")
		}
		for _, x := range []bool{false, true} {
			out, err := comb.Evaluate([]bool{x})
			if err != nil || out[0] != x {
				t.Fatalf("comb %v -> %v (%v)", x, out, err)
			}
		}
		seq := Bmc_GiaTargetStates(g, c, 0, 1, 0, mode.one, mode.all, 0)
		if seq == nil || gia.Gia_ManRegNum(seq) != 1 {
			t.Fatal("missing sequential target")
		}
		for bits := 0; bits < 4; bits++ {
			in := []bool{bits&1 != 0, bits&2 != 0}
			out, err := seq.Evaluate(in)
			if err != nil || out[0] != !in[1] || out[1] != !in[0] {
				t.Fatalf("seq %v -> %v (%v)", in, out, err)
			}
		}
	}
	if Bmc_GiaTargetStates(g, c, 2, 1, 1, 1, 0, 0) != nil {
		t.Fatal("invalid range accepted")
	}
}
