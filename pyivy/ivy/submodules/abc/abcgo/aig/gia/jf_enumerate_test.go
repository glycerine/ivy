package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestJfEnumerateAndCutModes(t *testing.T) {
	for _, mode := range []struct {
		name     string
		min, dsd int32
	}{{"structural", 0, 0}, {"truth", 1, 0}, {"dsd", 1, 1}} {
		t.Run(mode.name, func(t *testing.T) {
			g := Gia_ManStart(8)
			a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
			x := Gia_ManAppendAnd(g, a, b)
			Gia_ManAppendCo(g, x)
			p := Jf_ManAlloc(g, &Jf_Par_t{NLutSize: 4, NCutNum: 4, FCutMin: mode.min, FFuncDsd: mode.dsd})
			defer Jf_ManFree(p)
			p.pCutCmp = Jf_CutCompareDelay
			Jf_ObjAssignCut(p, Gia_ManObj(g, a>>1))
			Jf_ObjAssignCut(p, Gia_ManObj(g, b>>1))
			Jf_ObjComputeCuts(p, Gia_ManObj(g, x>>1), 0)
			cuts := jfCutList(Jf_ObjCuts(p, x>>1))
			if len(cuts) != 2 || Jf_CutSize(cuts[0]) != 2 || cuts[0][1] != a || cuts[0][2] != b || Jf_CutSize(cuts[1]) != 1 {
				t.Fatalf("cuts: %v", cuts)
			}
			if Jf_ObjArr(p, x>>1) != 1 || Jf_ObjFlow(p, x>>1) != 1 {
				t.Fatal("arrival/flow")
			}
			if mode.min != 0 && Jf_CutFunc(cuts[0]) != 4 {
				t.Fatalf("function %d", Jf_CutFunc(cuts[0]))
			}
		})
	}
}

func TestJfTruthCutPhasesAndSupport(t *testing.T) {
	g := Gia_ManStart(8)
	Gia_ManAppendCi(g)
	Gia_ManAppendCi(g)
	p := Jf_ManAlloc(g, &Jf_Par_t{NLutSize: 4, NCutNum: 4, FCutMin: 1})
	defer Jf_ManFree(p)
	a, b := []int32{1, 2}, []int32{1, 4}
	for phase := int32(0); phase < 4; phase++ {
		out := []int32{2, 2, 4, 0, 0}
		f := Jf_TtComputeForCut(p, 2|(phase&1), 2|(phase>>1), a, b, out)
		truth := vec.Vec_MemReadEntry(p.vTtMem, int(f>>1))[0]
		if f&1 != 0 {
			truth = ^truth
		}
		for assignment := 0; assignment < 4; assignment++ {
			want := ((assignment & 1) != int(phase&1)) && (((assignment >> 1) & 1) != int(phase>>1))
			if (truth>>uint(assignment)&1 != 0) != want {
				t.Fatalf("phase=%d assignment=%d truth=%x", phase, assignment, truth)
			}
		}
	}
	out := []int32{1, 2, 0, 0}
	f := Jf_TtComputeForCut(p, 2, 3, a, a, out)
	if out[0] != 0 || f != 0 {
		t.Fatalf("contradiction not reduced: cut=%v function=%d", out, f)
	}
}
