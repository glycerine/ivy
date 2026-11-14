package gia

import "testing"

func TestGiaManagerTruthTableNodeSnapshotAndHashState(t *testing.T) {
	p := Gia_ManStart(16)
	defer Gia_ManStop(p)
	if Gia_ManHashStarted(p) != 0 {
		t.Fatal("new manager unexpectedly has a structural hash table")
	}
	Gia_ManHashStart(p)
	if Gia_ManHashStarted(p) == 0 {
		t.Fatal("hash table state was not reported after initialization")
	}
	Gia_ObjComputeTruthTableStart(p, 3)
	defer Gia_ObjComputeTruthTableStop(p)

	a := Gia_ManAppendCi(p)
	b := Gia_ManAppendCi(p)
	c := Gia_ManAppendCi(p)
	ab := Gia_ManHashAnd(p, a, b)
	root := Gia_ManHashAnd(p, ab, c)
	Gia_ObjCollectInternal(p, Gia_ManObj(p, root>>1))
	ids := Gia_ManTruthTableNodeIDs(p)
	want := []int32{ab >> 1, root >> 1}
	if len(ids) != len(want) {
		t.Fatalf("node IDs = %v, want %v", ids, want)
	}
	for i := range want {
		if ids[i] != want[i] {
			t.Fatalf("node IDs = %v, want postorder %v", ids, want)
		}
	}
	ids[0] = -1
	if got := Gia_ManTruthTableNodeIDs(p)[0]; got != want[0] {
		t.Fatalf("returned node IDs alias manager storage: first ID = %d", got)
	}
}
