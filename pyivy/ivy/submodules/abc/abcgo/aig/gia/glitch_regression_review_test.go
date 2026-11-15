package gia

import "testing"

func glitchReviewBuffer() *Gli_Man_t {
	p := Gli_ManAlloc(3, 0, 2)
	input := Gli_ManCreateCi(p, 1)
	fanins := Vec_IntAlloc(1)
	Vec_IntPush(fanins, input)
	node := Gli_ManCreateNode(p, fanins, 1, []word{0xaaaaaaaaaaaaaaaa})
	Gli_ManCreateCo(p, node)
	return p
}

func TestGlitchFinalizeResetsActivity(t *testing.T) {
	p := glitchReviewBuffer()
	defer Gli_ManStop(p)
	Gli_ManFinalize(p)
	for _, o := range p.objects {
		o.nSwitches, o.nGlitches = 17, 23
	}
	Gli_ManFinalize(p)
	for _, o := range p.objects {
		if o.nSwitches != 0 || o.nGlitches != 0 {
			t.Fatalf("object%d retains switches=%d glitches=%d after finalize", o.Handle, o.nSwitches, o.nGlitches)
		}
	}
}

func TestGlitchRepeatedDriverStartsNewActivityCounts(t *testing.T) {
	p := glitchReviewBuffer()
	defer Gli_ManStop(p)
	Gli_ManSwitchesAndGlitches(p, 64, 0.5, 0)
	if Gli_ObjNumSwitches(p, 0) == 0 {
		t.Fatal("first simulation did not exercise input transitions")
	}
	Gli_ManSwitchesAndGlitches(p, 0, 0.5, 0)
	for _, o := range p.objects {
		if o.nSwitches != 0 || o.nGlitches != 0 {
			t.Fatalf("object%d retains prior driver activity: %d/%d", o.Handle, o.nSwitches, o.nGlitches)
		}
	}
}
