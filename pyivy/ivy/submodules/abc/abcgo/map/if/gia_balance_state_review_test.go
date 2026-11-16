package if_

import (
	"slices"
	"sync"
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	abcmain "github.com/glycerine/abcgo/base/main"
)

func balanceStateReviewGraph() *Gia_Man_t {
	src := Gia_ManStart(16)
	a, b, c := Gia_ManAppendCi(src), Gia_ManAppendCi(src), Gia_ManAppendCi(src)
	root := gia.Gia_ManAppendAnd(src, gia.Gia_ManAppendAnd(src, a, b^1), c)
	Gia_ManAppendCo(src, root)
	Gia_ManAppendCo(src, root^1)
	return src
}

func balanceReviewIsolateFrame(t *testing.T) *abcmain.Frame {
	t.Helper()
	previous := abcmain.Abc_FrameReadGlobalFrame()
	frame := abcmain.NewFrame()
	abcmain.Abc_FrameSetGlobalFrame(frame)
	t.Cleanup(func() {
		abcmain.Abc_FrameSetGlobalFrame(previous)
		if manager, ok := frame.Resource("pManDsd2").(*If_DsdMan_t); ok {
			If_DsdManFree(manager, 0)
		}
		frame.Close()
	})
	return frame
}

func TestDsdBalanceRetainsFrameManager(t *testing.T) {
	frame := balanceReviewIsolateFrame(t)
	var retained *If_DsdMan_t
	var count int32
	for pass := 0; pass < 3; pass++ {
		src := balanceStateReviewGraph()
		var result *Gia_Man_t
		dsdReportCapture(t, func() { result = Gia_ManPerformDsdBalance(src, 6, 8, 0, 0) })
		if result == nil {
			t.Fatal("DSD balancing failed")
		}
		manager, ok := frame.Resource("pManDsd2").(*If_DsdMan_t)
		if !ok || manager == nil {
			t.Fatal("balancing did not retain the source frame-owned DSD manager")
		}
		if pass == 0 {
			retained, count = manager, If_DsdManObjNum(manager)
			if count <= 2 || manager.nVars != 6 || manager.vIsops[3] == nil {
				t.Fatal("retained manager lost decomposition/ISOP data")
			}
		} else if manager != retained || If_DsdManObjNum(manager) != count {
			t.Fatal("repeated balancing did not reuse canonical DSD records")
		}
		for x := 0; x < 8; x++ {
			want := x&1 != 0 && x&2 == 0 && x&4 != 0
			got, err := result.Evaluate([]bool{x&1 != 0, x&2 != 0, x&4 != 0})
			if err != nil || !slices.Equal(got, []bool{want, !want}) {
				t.Fatalf("input%d: got%v err%v", x, got, err)
			}
		}
		Gia_ManStop(result)
		Gia_ManStop(src)
	}
}

func TestDsdBalanceUsesExistingFrameManager(t *testing.T) {
	frame := balanceReviewIsolateFrame(t)
	manager := If_DsdManAlloc(8, 0)
	frame.SetResource("pManDsd2", manager)
	perm := make([]byte, 3)
	lit := If_DsdManCompute(manager, []word{0x9696969696969696}, 3, perm, "")
	before := slices.Clone(If_DsdManComputeTruth(manager, lit, perm))
	src := balanceStateReviewGraph()
	defer Gia_ManStop(src)
	var result *Gia_Man_t
	dsdReportCapture(t, func() { result = Gia_ManPerformDsdBalance(src, 6, 8, 0, 0) })
	if result == nil {
		t.Fatal("balancing failed")
	}
	defer Gia_ManStop(result)
	if frame.Resource("pManDsd2") != manager || manager.nVars != 8 || !slices.Equal(If_DsdManComputeTruth(manager, lit, perm), before) {
		t.Fatal("balancing replaced or damaged the existing manager")
	}
	if manager.vIsops[3] == nil {
		t.Fatal("borrowed manager did not receive ISOPs")
	}
}

func TestDsdBalanceSerializesSharedInterning(t *testing.T) {
	frame := balanceReviewIsolateFrame(t)
	const workers = 8
	results := make(chan *Gia_Man_t, workers)
	var wg sync.WaitGroup
	dsdReportCapture(t, func() {
		for i := 0; i < workers; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				src := balanceStateReviewGraph()
				defer Gia_ManStop(src)
				results <- Gia_ManPerformDsdBalance(src, 6, 8, 0, 0)
			}()
		}
		wg.Wait()
	})
	close(results)
	for result := range results {
		if result == nil {
			t.Fatal("concurrent balancing failed")
		}
		for x := 0; x < 8; x++ {
			want := x&1 != 0 && x&2 == 0 && x&4 != 0
			got, err := result.Evaluate([]bool{x&1 != 0, x&2 != 0, x&4 != 0})
			if err != nil || !slices.Equal(got, []bool{want, !want}) {
				t.Errorf("input%d: got%v err%v", x, got, err)
			}
		}
		Gia_ManStop(result)
	}
	manager, ok := frame.Resource("pManDsd2").(*If_DsdMan_t)
	if !ok || manager == nil || If_DsdManObjNum(manager) <= 2 {
		t.Fatal("shared manager lost its canonical records")
	}
}
