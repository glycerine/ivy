package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"runtime"
	"sync/atomic"
	"testing"
	"time"
)

// The C coordinator scans completed worker slots in increasing index order.
// Both completed SAT answers are ready at one scan boundary; slot zero must
// win even though the other worker may have finished first.
func TestSplitParallelSimultaneousReadySlots(t *testing.T) {
	g := gia.Gia_ManStart(12)
	a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, gia.Gia_ManAppendAnd(g, a, b))
	solve := func(p splitPart, _ int32) splitResult {
		if len(p.fixes) == 0 {
			return splitResult{status: -1}
		}
		values, err := p.graph.Evaluate([]bool{true, true})
		if err != nil {
			panic(err)
		}
		return splitResult{status: 0, model: []bool{values[0]}}
	}
	synchronized := false
	hook := func(slots []splitSlot) {
		if synchronized || slots[0].part == nil || slots[1].part == nil {
			return
		}
		deadline := time.Now().Add(5 * time.Second)
		for len(slots[0].done) == 0 || len(slots[1].done) == 0 {
			if time.Now().After(deadline) {
				t.Fatal("workers did not publish results")
			}
			runtime.Gosched()
		}
		synchronized = true
	}
	status, model := splitParallelWithScanHook(g, 2, 0, 0, 1, 0, 0, 1, time.Now(), solve, hook)
	if !synchronized || status != 0 || len(model) != 1 || !model[0] {
		t.Fatalf("synchronized=%v status=%d model=%v", synchronized, status, model)
	}
}

// The source checks nIterMax after completing a pass through all ready slots.
func TestSplitParallelMaxIterChecksAfterReadyPass(t *testing.T) {
	g := gia.Gia_ManStart(16)
	a, b, c := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, gia.Gia_ManAppendAnd(g, gia.Gia_ManAppendAnd(g, a, b), c))
	release := make(chan struct{})
	var deeper atomic.Int32
	solve := func(p splitPart, _ int32) splitResult {
		if len(p.fixes) == 1 {
			<-release
		}
		if len(p.fixes) == 2 {
			deeper.Add(1)
		}
		return splitResult{status: -1}
	}
	synchronized := false
	hook := func(slots []splitSlot) {
		if synchronized || slots[0].part == nil || slots[1].part == nil {
			return
		}
		close(release)
		deadline := time.Now().Add(5 * time.Second)
		for len(slots[0].done) == 0 || len(slots[1].done) == 0 {
			if time.Now().After(deadline) {
				t.Fatal("workers did not publish results")
			}
			runtime.Gosched()
		}
		synchronized = true
	}
	status, _ := splitParallelWithScanHook(g, 2, 0, 2, 1, 0, 0, 1, time.Now(), solve, hook)
	if !synchronized || status != -1 || deeper.Load() != 2 {
		t.Fatalf("synchronized=%v status=%d deeper=%d", synchronized, status, deeper.Load())
	}
}
