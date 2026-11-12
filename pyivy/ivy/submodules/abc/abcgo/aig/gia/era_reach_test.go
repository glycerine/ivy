package gia

import "testing"

func TestExplicitReachability(t *testing.T) {
	p := Gia_ManStart(8)
	input, reg := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	Gia_ManAppendCo(p, reg)
	Gia_ManAppendCo(p, input)
	Gia_ManSetRegNum(p, 1)
	if got := Gia_ManCollectReachable(p, 100, 1, 0, 0); got != 0 {
		t.Fatalf("expected reachable failure, got %d", got)
	}
	if got := Gia_ManCollectReachable(p, 100, 0, 0, 0); got != 1 {
		t.Fatalf("expected exhaustive traversal, got %d", got)
	}
	if got := Gia_ManCollectReachable(p, 2, 0, 0, 0); got != -1 {
		t.Fatalf("expected state limit, got %d", got)
	}
}
func TestExplicitReachabilityStateAndTrace(t *testing.T) {
	p := Gia_ManStart(8)
	input, reg := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	Gia_ManAppendCo(p, reg)
	Gia_ManAppendCo(p, input)
	Gia_ManSetRegNum(p, 1)
	m := Gia_ManEraCreate(p)
	first := Gia_ManEraCreateState(m)
	m.pBins[Gia_ManEraStateHash(first.pData, m.nWordsDat, m.nBins)] = first.Num
	Gia_ManInsertState(m, first)
	Gia_ManPerformOneIter(m)
	if Gia_ManAnalyzeResult(m, first, 1, 1) != 0 {
		t.Fatal("initial state should be safe")
	}
	if len(m.vStates) != 3 || Vec_IntSize(m.vStgDump) != 8 {
		t.Fatalf("states=%d transitions=%v", len(m.vStates), Vec_IntArray(m.vStgDump))
	}
	next := m.vStates[2]
	Gia_ManInsertState(m, next)
	Gia_ManPerformOneIter(m)
	if Gia_ManAnalyzeResult(m, next, 1, 0) != 1 {
		t.Fatal("expected failure")
	}
	trace := Vec_IntArray(m.vBugTrace)
	if len(trace) != 3 || trace[1] != 1 {
		t.Fatalf("unexpected bug trace %v", trace)
	}
}
