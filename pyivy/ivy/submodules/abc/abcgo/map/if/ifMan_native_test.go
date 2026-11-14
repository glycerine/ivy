package if_

import "testing"

func testManagerPars() *If_Par_t {
	return &If_Par_t_{nLutSize: 4, nCutsMax: 8, nFlowIters: 1, nAreaIters: 2}
}

func TestIfManagerBuildsAndFoldsGraph(t *testing.T) {
	p := If_ManStart(testManagerPars())
	defer If_ManStop(p)
	if got := If_ManObjNum(p); got != 1 {
		t.Fatalf("initial objects = %d, want constant node only", got)
	}
	ci0, ci1 := If_ManCreateCi(p), If_ManCreateCi(p)
	and := If_ManCreateAnd(p, ci0, ci1)
	if got := If_ManCreateAnd(p, ci0, ci0); got != ci0 {
		t.Fatal("x & x was not folded to x")
	}
	if got := If_ManCreateAnd(p, ci0, If_Not(ci0)); got != If_Not(p.pConst1) {
		t.Fatal("x & !x was not folded to false")
	}
	out := If_ManCreateCo(p, and)
	if out.pFanin0 != and || If_ManCiNum(p) != 2 || If_ManCoNum(p) != 1 || If_ManAndNum(p) != 1 {
		t.Fatalf("unexpected graph counts or output driver: CIs=%d COs=%d ANDs=%d", If_ManCiNum(p), If_ManCoNum(p), If_ManAndNum(p))
	}
	if and.nRefs != 1 || ci0.nRefs != 1 || ci1.nRefs != 1 {
		t.Fatalf("unexpected fanin reference counts: and=%d ci0=%d ci1=%d", and.nRefs, ci0.nRefs, ci1.nRefs)
	}
}

func TestIfManagerMetadataAccessors(t *testing.T) {
	p := If_ManStart(testManagerPars())
	defer If_ManStop(p)
	If_ManSetName(p, "mapper-input")
	if p.pName != "mapper-input" {
		t.Fatalf("manager name = %q, want mapper-input", p.pName)
	}
	if If_ManObjBytes(p) <= 0 {
		t.Fatalf("manager object size = %d, want positive size", If_ManObjBytes(p))
	}
}

func TestIfManagerCutSetPoolReuseAndRestart(t *testing.T) {
	p := If_ManStart(testManagerPars())
	defer If_ManStop(p)
	ci0, ci1 := If_ManCreateCi(p), If_ManCreateCi(p)
	and := If_ManCreateAnd(p, ci0, ci1)
	If_ManSetupCiCutSets(p)
	If_ManSetupSetAll(p, 0)
	set := If_ManSetupNodeCutSet(p, and)
	If_ManCutSetRecycle(p, set)
	if got := If_ManCutSetFetch(p); got != set {
		t.Fatal("recycled cut set was not reused")
	}

	If_ManRestart(p)
	if If_ManCiNum(p) != 0 || If_ManCoNum(p) != 0 || If_ManAndNum(p) != 0 || If_ManObjNum(p) != 1 {
		t.Fatalf("restart retained graph state: ci=%d co=%d and=%d obj=%d", If_ManCiNum(p), If_ManCoNum(p), If_ManAndNum(p), If_ManObjNum(p))
	}
}

func TestIfManagerStopIsIdempotent(t *testing.T) {
	p := If_ManStart(testManagerPars())
	If_ManStop(p)
	If_ManStop(p)
	if !p.stopped {
		t.Fatal("manager stop flag was not set")
	}
}
