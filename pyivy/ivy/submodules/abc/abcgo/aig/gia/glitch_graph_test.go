package gia

import "testing"

func TestGlitchGraphConstruction(t *testing.T) {
	p := Gli_ManAlloc(4, 0, 3)
	a, b := Gli_ManCreateCi(p, 1), Gli_ManCreateCi(p, 1)
	fanins := Vec_IntAlloc(2)
	Vec_IntPush(fanins, a)
	Vec_IntPush(fanins, b)
	n := Gli_ManCreateNode(p, fanins, 1, []word{0x8888888888888888})
	out := Gli_ManCreateCo(p, n)
	Gli_ManFinalize(p)
	if Gli_ManObjNum(p) != 4 || Gli_ManNodeNum(p) != 1 || Gli_ObjFanin(Gli_ManObj(p, n), 0) != Gli_ManObj(p, a) || Gli_ObjFanout(Gli_ManObj(p, n), 0) != Gli_ManObj(p, out) {
		t.Fatal("invalid graph")
	}
	Gli_ManObj(p, a).fPhase = 1
	Gli_ManObj(p, b).fPhase = 1
	if Gli_NodeComputeValue(Gli_ManObj(p, n)) != 1 {
		t.Fatal("AND truth evaluation")
	}
}
