package abc

import "testing"

func TestBACBridgeAccessors(t *testing.T) {
	n := NewNetwork("bridge")
	pi := Abc_NtkCreatePi(n)
	bar := Abc_NtkCreateNode(n)
	bar.Barrier = true
	po := Abc_NtkCreatePo(n)
	Abc_ObjAddFanin(bar, pi)
	Abc_ObjAddFanin(po, bar)

	if Abc_NtkPiNum(n) != 1 || Abc_NtkPoNum(n) != 1 || Abc_NtkBarBufNum(n) != 0 {
		t.Fatalf("unexpected interface/barrier counts")
	}
	if Abc_ObjIsBarBuf(bar) != 1 || len(Abc_NtkNodes(n)) != 1 || Abc_NtkNodes(n)[0] != bar {
		t.Fatalf("node/barrier accessors did not expose the source objects")
	}
	Abc_ObjSetTemp(bar, 17)
	if Abc_ObjTemp(bar) != 17 {
		t.Fatalf("temporary mark = %d, want 17", Abc_ObjTemp(bar))
	}
	if Abc_NtkSpec(n) != "" {
		t.Fatalf("unexpected spec %q", Abc_NtkSpec(n))
	}
}
