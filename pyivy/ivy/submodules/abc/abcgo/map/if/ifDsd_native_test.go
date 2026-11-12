package if_

import (
	"testing"

	"github.com/glycerine/abcgo/opt/dau"
)

func TestIfDsdComputeKeepsTruthAndLiteralPermutation(t *testing.T) {
	truth := dau.Dau_DsdToTruth("(abc)", 3)
	p := If_DsdManAlloc(3, 2)
	defer If_DsdManFree(p, 0)
	permutation := make([]byte, 3)
	lit := If_DsdManCompute(p, truth, 3, permutation, "2")
	if lit&1 != 0 {
		t.Fatalf("unexpected complemented DSD literal %d", lit)
	}
	if got, want := permutation, []byte{0, 2, 4}; string(got) != string(want) {
		t.Fatalf("pin selectors = %v, want %v", got, want)
	}
	obj := dsdObject(p, lit)
	if got := dau.Dau_DsdToTruth(obj.text, 3); got[0] != truth[0] {
		t.Fatalf("DSD text %q evaluates to %#x, want %#x", obj.text, got[0], truth[0])
	}
	if obj.nSupp != 3 {
		t.Fatalf("support = %d, want 3", obj.nSupp)
	}
}

func TestIfDsdCheckXYFindsAndMasksDecomposition(t *testing.T) {
	truth := dau.Dau_DsdToTruth("((ab)c)", 3)
	p := If_DsdManAlloc(3, 2)
	defer If_DsdManFree(p, 0)
	perm := make([]byte, 3)
	lit := If_DsdManCompute(p, truth, 3, perm, "2")
	set := If_DsdManCheckXY(p, lit, 2, 0, 0, 0, 0)
	if set == 0 {
		t.Fatal("expected a two-LUT decomposition")
	}
	if got := If_DsdManCheckXY(p, lit, 2, 0, ^uint32(0), 0, 0); got != 0 {
		t.Fatalf("masked decomposition %#x remained available", got)
	}
}

func TestIfDsdDelayRecUsesAssociativeCounter(t *testing.T) {
	text, _ := dau.DecomposeDSD(dau.Dau_DsdToTruth("(ab)", 2), 2)
	network := dau.Dss_NtkCreate(text, 2, nil)
	var support int32
	delay, profile := dsdDelayRec(network, network.Root, []byte{0, 2}, []int32{2, 5}, &support, 2)
	if support != 2 {
		t.Fatalf("visited support %d, want 2", support)
	}
	if delay != 6 {
		t.Fatalf("AND delay %d, want 6", delay)
	}
	if profile == 0 {
		t.Fatal("expected a non-empty per-pin delay profile")
	}
}
