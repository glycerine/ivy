package gia

import (
	abcmain "github.com/glycerine/abcgo/base/main"
	"reflect"
	"testing"
)

type groupProofDispatcher struct{ calls int }

func (d *groupProofDispatcher) Execute(frame *abcmain.Frame, command string) int32 {
	p := abcmain.Abc_FrameReadGia(frame).(*Gia_Man_t)
	n := int(Gia_ManPoNum(p))
	statuses := make([]int32, n)
	cexes := make([]any, n)
	for i := range statuses {
		statuses[i] = int32((d.calls+i)%3 - 1)
		if statuses[i] == 0 {
			cexes[i] = "counterexample"
		}
	}
	frame.SetIntValues("vStatuses", statuses)
	frame.SetListValues("vCexVec", cexes)
	d.calls++
	return 0
}
func TestGroupProvePreservesStatuses(t *testing.T) {
	prior := abcmain.Abc_FrameReadGlobalFrame()
	defer abcmain.Abc_FrameSetGlobalFrame(prior)
	frame := abcmain.NewFrame()
	engine := new(groupProofDispatcher)
	frame.SetDispatcher(engine)
	abcmain.Abc_FrameSetGlobalFrame(frame)
	p := Gia_ManStart(10)
	a := Gia_ManAppendCi(p)
	for i := 0; i < 5; i++ {
		Gia_ManAppendCo(p, a^int32(i&1))
	}
	out := Gia_ManGroupProve(p, "prove", 2, 0)
	if got := Vec_IntArray(out); !reflect.DeepEqual(got, []int32{-1, 0, 0, 1, 1}) {
		t.Fatalf("statuses %v", got)
	}
	if engine.calls != 3 || len(abcmain.Abc_FrameReadCexVec(frame)) != 5 {
		t.Fatal("group counterexamples or call count")
	}
}

func TestLegacyChoiceDetection(t *testing.T) {
	p := Gia_ManStart(10)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x, y := Gia_ManAppendAnd(p, a, b), Gia_ManAppendAnd(p, a, b)
	Gia_ManAppendCo(p, x)
	if Gia_ManHasChoices_very_old(p) != 0 {
		t.Fatal("choices without representatives")
	}
	p.pReprs = make([]Gia_Rpr_t, p.nObjs)
	for i := range p.pReprs {
		p.pReprs[i].iRepr = GIA_VOID
	}
	p.pNexts = make([]int32, p.nObjs)
	Gia_ObjSetRepr(p, y>>1, x>>1)
	p.pNexts[x>>1] = y >> 1
	if Gia_ManHasChoices_very_old(p) != 1 {
		t.Fatal("choice not detected")
	}
}
