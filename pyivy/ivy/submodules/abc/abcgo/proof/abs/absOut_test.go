package abs

import (
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestGlaRefineRealAndSpuriousCounterexamples(t *testing.T) {
	for _, included := range []bool{true, false} {
		g := gia.Gia_ManStart(12)
		a, b, c := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
		gate := gia.Gia_ManAppendAnd(g, a, b)
		root := gia.Gia_ManAppendAnd(g, gate, c)
		gia.Gia_ManAppendCo(g, root)
		classes := vec.Vec_IntStart(int(gia.Gia_ManObjNum(g)))
		vec.Vec_IntWriteEntry(classes, 0, 1)
		vec.Vec_IntWriteEntry(classes, int(root>>1), 1)
		pis := int32(2)
		if included {
			vec.Vec_IntWriteEntry(classes, int(gate>>1), 1)
			pis = 3
		}
		gia.Gia_ManSetGateClasses(g, classes)
		cex := util.Abc_CexAlloc(0, pis, 1)
		for i := int32(0); i < pis; i++ {
			util.Abc_InfoSetBit(cex.PData, i)
		}
		got := Gia_ManGlaRefine(g, cex, 0, 0)
		if included {
			if got != 0 || g.SequentialCounterexample() == nil || gia.Gia_ManVerifyCex(g, g.SequentialCounterexample(), 0) == 0 {
				t.Fatalf("real CEX result=%d", got)
			}
		} else {
			if got != -1 || vec.Vec_IntEntry(classes, int(gate>>1)) != 1 || g.SequentialCounterexample() != nil {
				t.Fatalf("spurious CEX result=%d classes=%v", got, vec.Vec_IntArray(classes))
			}
		}
	}
}
func TestNewRefineRequiresGateClasses(t *testing.T) {
	g := gia.Gia_ManStart(4)
	a := gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, a)
	cex := util.Abc_CexAlloc(0, 1, 1)
	if got := Gia_ManNewRefine(g, cex, 0, 0, 0); got != -1 {
		t.Fatalf("missing map result=%d", got)
	}
}

func TestNewRefineRejectsInvalidCounterexample(t *testing.T) {
	for _, badCount := range []bool{false, true} {
		g := gia.Gia_ManStart(8)
		a := gia.Gia_ManAppendCi(g)
		b := gia.Gia_ManAppendCi(g)
		gate := gia.Gia_ManAppendAnd(g, a, b)
		gia.Gia_ManAppendCo(g, gate)
		classes := vec.Vec_IntStart(int(gia.Gia_ManObjNum(g)))
		vec.Vec_IntWriteEntry(classes, 0, 1)
		vec.Vec_IntWriteEntry(classes, int(gate>>1), 1)
		gia.Gia_ManSetGateClasses(g, classes)
		n := int32(2)
		if badCount {
			n = 3
		}
		cex := util.Abc_CexAlloc(0, n, 1)
		if got := Gia_ManNewRefine(g, cex, 0, 0, 0); got != -1 {
			t.Fatalf("badCount=%v result=%d", badCount, got)
		}
		if gia.Gia_ManGateClasses(g) != classes {
			t.Fatal("invalid counterexample replaced abstraction")
		}
	}
}
