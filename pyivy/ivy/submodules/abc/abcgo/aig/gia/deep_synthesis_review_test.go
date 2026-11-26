package gia

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"

	abcmain "github.com/glycerine/abcgo/base/main"
)

func deepReviewFrame(t *testing.T) *abcmain.Frame {
	t.Helper()
	old := abcmain.Abc_FrameReadGlobalFrame()
	f := abcmain.NewFrame()
	abcmain.Abc_FrameSetGlobalFrame(f)
	t.Cleanup(func() { f.Close(); abcmain.Abc_FrameSetGlobalFrame(old) })
	f.Err = &bytes.Buffer{}
	return f
}

func TestDeepSynthesisRestoresBatchAfterPanic(t *testing.T) {
	for initial := int32(0); initial < 2; initial++ {
		t.Run(fmt.Sprint(initial), func(t *testing.T) {
			f := deepReviewFrame(t)
			g := deepReviewGraph(6)
			f.SetResource("pGia", g)
			abcmain.Abc_FrameSetBatchMode(initial)
			f.Dispatcher = &deepReviewDispatcher{last: g, panicAt: 1}
			func() {
				defer func() {
					if got := recover(); got != "scripted synthesis panic" {
						t.Fatalf("unexpected panic: %v", got)
					}
				}()
				Gia_ManDeepSynOne(3, 0, 0, 0, 0, 0)
			}()
			if abcmain.Abc_FrameIsBatchMode() != initial {
				t.Fatal("command panic leaked batch mode")
			}
		})
	}
}

func TestDeepSynthesisTimeoutIncludesInitialCopy(t *testing.T) {
	f := deepReviewFrame(t)
	g := deepReviewGraph(6)
	f.SetResource("pGia", g)
	constant := g.pObjs[0]
	constant.Value = 1234
	base := time.Unix(100, 0)
	clock := base
	first := true
	now := func() time.Time {
		if first {
			first = false
			// Dup writes the source literal map. Model its two-second cost when the
			// clock is first read after that observable copy side effect.
			if constant.Value != 1234 {
				clock = base.Add(2 * time.Second)
			}
		}
		return clock
	}
	d := &deepReviewDispatcher{last: g, after: func() { clock = base.Add(2 * time.Second) }}
	f.Dispatcher = d
	var result *Gia_Man_t
	report := hashReviewStdout(t, func() { result = giaManDeepSynOne(3, 1, 0, 0, 0, 0, now) })
	if result == nil {
		t.Fatal("timeout discarded the best graph")
	}
	defer Gia_ManStop(result)
	if d.calls != 1 || !strings.Contains(report, "Runtime limit (1 sec)") {
		t.Fatalf("initial copy was excluded from timeout: calls%d report%s", d.calls, report)
	}
}

func TestDeepSynthesisDocumentedQualityGoal(t *testing.T) {
	for _, outer := range []bool{false, true} {
		t.Run(fmt.Sprint(outer), func(t *testing.T) {
			f := deepReviewFrame(t)
			g := deepReviewGraph(6)
			f.SetResource("pGia", g)
			d := &deepReviewDispatcher{scenario: 2, last: g}
			f.Dispatcher = d
			var result *Gia_Man_t
			report := hashReviewStdout(t, func() {
				if outer {
					result = Gia_ManDeepSyn(g, 3, 3, 0, 4, 0, 0, 0)
				} else {
					result = Gia_ManDeepSynOne(3, 0, 4, 0, 0, 0)
				}
			})
			if result == nil {
				t.Fatal("quality goal failed")
			}
			defer Gia_ManStop(result)
			calls := 6
			rounds := 1
			if outer {
				calls = 18
				rounds = 3
			}
			if Gia_ManAndNum(result) != 4 || d.calls != calls || strings.Count(report, "Quality goal (4 nodes <= 4 nodes)") != rounds {
				t.Fatalf("quality goal result%d calls%d report%s", Gia_ManAndNum(result), d.calls, report)
			}
		})
	}
}

func TestDeepSynthesisFailureAndMissingGraphs(t *testing.T) {
	for initial := int32(0); initial < 2; initial++ {
		for _, kind := range []string{"first", "later", "outer", "removed", "absent", "wrong-type", "typed-nil"} {
			t.Run(fmt.Sprintf("%d/%s", initial, kind), func(t *testing.T) {
				f := deepReviewFrame(t)
				g := deepReviewGraph(6)
				f.SetResource("pGia", g)
				abcmain.Abc_FrameSetBatchMode(initial)
				d := &deepReviewDispatcher{scenario: 2, last: g}
				switch kind {
				case "first":
					d.failAt = 1
				case "later":
					d.failAt = 5
				case "outer":
					d.failAt = 20
				case "removed":
					d.remove = true
				case "absent":
					f.SetResource("pGia", nil)
				case "wrong-type":
					f.SetResource("pGia", "wrong type")
				case "typed-nil":
					var missing *Gia_Man_t
					f.SetResource("pGia", missing)
				}
				f.Dispatcher = d
				var result *Gia_Man_t
				hashReviewStdout(t, func() {
					if kind == "outer" {
						result = Gia_ManDeepSyn(g, 3, 3, 0, 0, 0, 0, 0)
					} else {
						result = Gia_ManDeepSynOne(3, 0, 0, 0, 0, 0)
					}
				})
				// Gia_ManDeepSyn keeps the best graph so far (CBUGS_FIXED.md #13).
				if (kind == "outer") != (result != nil) || abcmain.Abc_FrameIsBatchMode() != initial {
					t.Fatal("wrong failure result or leaked batch mode")
				}
				if kind == "outer" && Gia_ManAndNum(result) != Gia_ManAndNum(g) {
					t.Fatal("failure did not keep the best graph")
				}
				if f.Err.(*bytes.Buffer).Len() == 0 {
					t.Fatal("missing error report")
				}
				switch kind {
				case "first", "later", "outer":
					if d.calls != d.failAt {
						t.Fatal("failure did not stop command execution")
					}
				case "removed":
					if d.calls != 1 {
						t.Fatal("missing output graph was not detected")
					}
				default:
					if d.calls != 0 {
						t.Fatal("invalid initial graph executed a command")
					}
				}
				if current, ok := abcmain.Abc_FrameReadGia(f).(*Gia_Man_t); ok && current != nil && current.nObjs == 0 {
					t.Fatal("failure destroyed frame-owned graph")
				}
			})
		}
	}
}

func TestDeepSynthesisZeroOuterIterationsOwnsResult(t *testing.T) {
	f := deepReviewFrame(t)
	current := deepReviewGraph(3)
	f.SetResource("pGia", current)
	input := deepReviewGraph(6)
	defer Gia_ManStop(input)
	d := &deepReviewDispatcher{last: current}
	f.Dispatcher = d
	result := Gia_ManDeepSyn(input, 0, 3, 0, 0, 0, 0, 0)
	if result == nil || result == input || result == current || Gia_ManAndNum(result) != 6 || d.calls != 0 || abcmain.Abc_FrameReadGia(f) != current {
		t.Fatal("zero trials changed frame or borrowed input")
	}
	Gia_ManStop(result)
	if input.nObjs == 0 || current.nObjs == 0 {
		t.Fatal("disposing owned result invalidated borrowed graph")
	}
}

func TestDeepSynthesisDeterministicTimeoutAndVerbose(t *testing.T) {
	for _, limit := range []int32{0, 1, 3} {
		for initial := int32(0); initial < 2; initial++ {
			t.Run(fmt.Sprintf("limit%d/batch%d", limit, initial), func(t *testing.T) {
				f := deepReviewFrame(t)
				g := deepReviewGraph(6)
				f.SetResource("pGia", g)
				abcmain.Abc_FrameSetBatchMode(initial)
				clock := time.Unix(100, 0)
				d := &deepReviewDispatcher{scenario: 1, last: g, after: func() { clock = clock.Add(time.Second) }}
				f.Dispatcher = d
				var result *Gia_Man_t
				report := hashReviewStdout(t, func() { result = giaManDeepSynOne(2, limit, 0, 3, 1, 1, func() time.Time { return clock }) })
				if result == nil {
					t.Fatal("timed synthesis failed")
				}
				defer Gia_ManStop(result)
				expected := int(limit)
				if limit == 0 {
					expected = 4
				}
				if d.calls != expected || Gia_ManAndNum(result) != 5 || abcmain.Abc_FrameIsBatchMode() != initial {
					t.Fatalf("timeout result%d calls%d expected%d", Gia_ManAndNum(result), d.calls, expected)
				}
				if strings.Count(report, "<== best :") != 1 || !strings.Contains(report, "Time     1.00 sec") {
					t.Fatalf("incorrect improvement report: %s", report)
				}
				if limit != 0 && !strings.Contains(report, "Runtime limit") {
					t.Fatal("timeout did not report its reason")
				}
			})
		}
	}
}

func TestDeepSynthesisIterationLimitKeepsFirstTie(t *testing.T) {
	f := deepReviewFrame(t)
	g := deepReviewGraph(6)
	g.pName = "original best"
	f.SetResource("pGia", g)
	d := &deepReviewDispatcher{last: g, keep: true}
	d.after = func() {
		if d.calls == 1 {
			g.pName = "later equal-cost candidate"
		}
	}
	f.Dispatcher = d
	clock := time.Unix(100, 0)
	var result *Gia_Man_t
	report := hashReviewStdout(t, func() { result = giaManDeepSynOne(100001, 0, 0, 0, 0, 0, func() time.Time { return clock }) })
	if result == nil {
		t.Fatal("iteration limit lost best graph")
	}
	defer Gia_ManStop(result)
	if d.calls != 100000 || !strings.Contains(report, "Iteration limit (100000 iters)") {
		t.Fatalf("iteration bound: calls%d report%s", d.calls, report)
	}
	if result.pName != "original best" || g.pName != "later equal-cost candidate" || result == g {
		t.Fatal("equal-cost candidate replaced first owned best")
	}
}

func TestDeepSynthesisZeroTrialsDoesNotCreateFrame(t *testing.T) {
	old := abcmain.Abc_FrameReadGlobalFrame()
	abcmain.Abc_FrameSetGlobalFrame(nil)
	defer func() {
		if created := abcmain.Abc_FrameReadGlobalFrame(); created != nil {
			created.Close()
		}
		abcmain.Abc_FrameSetGlobalFrame(old)
	}()
	input := deepReviewGraph(6)
	defer Gia_ManStop(input)
	result := Gia_ManDeepSyn(input, 0, 3, 0, 0, 0, 0, 0)
	if result == nil || result == input || Gia_ManAndNum(result) != 6 {
		t.Fatal("zero trials failed to copy input")
	}
	defer Gia_ManStop(result)
	if abcmain.Abc_FrameReadGlobalFrame() != nil {
		t.Fatal("zero trials created a global frame")
	}
}

func TestDeepSynthesisNoImprovementBoundaries(t *testing.T) {
	for _, limit := range []int32{0, 1, 3} {
		for scenario := 0; scenario < 2; scenario++ {
			t.Run(fmt.Sprintf("limit%d/scenario%d", limit, scenario), func(t *testing.T) {
				f := deepReviewFrame(t)
				g := deepReviewGraph(6)
				f.SetResource("pGia", g)
				d := &deepReviewDispatcher{scenario: scenario, last: g}
				f.Dispatcher = d
				var result *Gia_Man_t
				hashReviewStdout(t, func() { result = Gia_ManDeepSynOne(limit, 0, 0, 0, 0, 0) })
				if result == nil {
					t.Fatal("stagnation discarded best graph")
				}
				defer Gia_ManStop(result)
				expected := int(limit) + 1
				if scenario == 1 {
					expected++
				}
				if d.calls != expected {
					t.Fatalf("stagnation calls%d expected%d", d.calls, expected)
				}
			})
		}
	}
}
