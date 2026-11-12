package gia

import "testing"

func TestGlitchPropagationReconvergent(t *testing.T) {
	p := Gli_ManAlloc(5, 0, 5)
	a := Gli_ManCreateCi(p, 2)
	ins := Vec_IntAlloc(1)
	Vec_IntPush(ins, a)
	buffer := Gli_ManCreateNode(p, ins, 1, []word{0xaaaaaaaaaaaaaaaa})
	invIns := Vec_IntAlloc(1)
	Vec_IntPush(invIns, buffer)
	inv := Gli_ManCreateNode(p, invIns, 1, []word{0x5555555555555555})
	both := Vec_IntAlloc(2)
	Vec_IntPush(both, a)
	Vec_IntPush(both, inv)
	out := Gli_ManCreateNode(p, both, 1, []word{0x8888888888888888})
	Gli_ManCreateCo(p, out)
	Gli_ManFinalize(p)
	// a=0->1: the direct edge changes before its inverted reconvergent edge.
	gliFlip(p, Gli_ManObj(p, a))
	Gli_ManSwitching(p)
	Gli_ManGlitching(p)
	node := Gli_ManObj(p, out)
	if node.nSwitches != 0 || node.nGlitches != 2 || node.fPhase != 0 || node.fPhase2 != 0 {
		t.Fatalf("switches=%d glitches=%d phases=%d,%d", node.nSwitches, node.nGlitches, node.fPhase, node.fPhase2)
	}
	Gli_ManVerify(p)
}
func TestGlitchSequentialTruth(t *testing.T) {
	p := Gli_ManAlloc(4, 0, 3)
	a, b := Gli_ManCreateCi(p, 1), Gli_ManCreateCi(p, 1)
	fanins := Vec_IntAlloc(2)
	Vec_IntPush(fanins, a)
	Vec_IntPush(fanins, b)
	id := Gli_ManCreateNode(p, fanins, 1, []word{0x8888888888888888})
	Gli_ManCreateCo(p, id)
	Gli_ManObj(p, a).uSimInfo = 0xaaaaaaaa
	Gli_ManObj(p, b).uSimInfo = 0xcccccccc
	if got := Gli_ManSimulateSeqNode(p, Gli_ManObj(p, id)); got != 0x88888888 {
		t.Fatalf("got %08x", got)
	}
}
