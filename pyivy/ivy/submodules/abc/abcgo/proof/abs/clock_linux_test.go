//go:build linux

package abs

import (
	"bytes"
	"runtime"
	"testing"
	"time"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/sat/bsat"
)

func TestCPUDeadlineSleepAndSolverRetry(t *testing.T) {
	defer absPinClockThread()()
	s := bsat.NewLabeledSolver2()
	s.NewVariable()
	deadline := absCPUClock() + int64(30*time.Millisecond)
	s.SetRuntimeDeadline(absCPUClock, deadline)
	time.Sleep(80 * time.Millisecond)
	var status int32
	if core := Vta_ManUnsatCore(int32(bsat.Positive(0)), s, 0, 0, &status, nil); status != 0 || core != nil {
		t.Fatalf("sleep consumed CPU budget: status=%d core=%v", status, core)
	}
	for absCPUClock() <= deadline {
	}
	if core := Vta_ManUnsatCore(int32(bsat.Positive(0)), s, 0, 0, &status, nil); status != -1 || core != nil {
		t.Fatalf("CPU deadline ignored: status=%d core=%v", status, core)
	}
	s.SetDeadline(time.Time{})
	if core := Vta_ManUnsatCore(int32(bsat.Positive(0)), s, 0, 0, &status, nil); status != 0 || core != nil {
		t.Fatalf("wall deadline reset failed: status=%d core=%v", status, core)
	}
}

func TestCPUDeadlineUnlocksOnEarlyReturn(t *testing.T) {
	locked := func() bool {
		var buf [1024]byte
		n := runtime.Stack(buf[:], false)
		return bytes.Contains(bytes.SplitN(buf[:n], []byte("\n"), 2)[0], []byte("locked to thread"))
	}
	if locked() {
		t.Skip("test goroutine is already thread-locked")
	}
	g := gia.Gia_ManStart(2)
	gia.Gia_ManAppendCo(g, 0)
	p := NewParameters()
	o := p.Options()
	o.TimeoutSeconds = 1
	p.SetOptions(o)
	if got := Gia_ManPerformGla(g, p); got != 1 {
		t.Fatalf("constant output returned %d", got)
	}
	if locked() {
		t.Fatal("GLA left the calling goroutine locked to its OS thread")
	}
	o.FramesMax, o.FramesStart = 1, 2
	p.SetOptions(o)
	panicked := false
	func() {
		defer func() { panicked = recover() != nil }()
		Gia_VtaPerformInt(g, p)
	}()
	if !panicked {
		t.Fatal("expected VTA frame-bound assertion")
	}
	if locked() {
		t.Fatal("VTA panic left the calling goroutine locked to its OS thread")
	}
}
