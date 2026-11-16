package if_

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	abcmain "github.com/glycerine/abcgo/base/main"
)

func driverCallbackReviewGraph(t *testing.T, arrivals, required bool) *Gia_Man_t {
	t.Helper()
	src := Gia_ManStart(16)
	src.SetIdentity("callback-review", "callback-spec")
	a, b, c := Gia_ManAppendCi(src), Gia_ManAppendCi(src), Gia_ManAppendCi(src)
	root := gia.Gia_ManAppendAnd(src, gia.Gia_ManAppendAnd(src, a, b^1), c)
	Gia_ManAppendCo(src, root)
	Gia_ManAppendCo(src, root^1)
	frame := abcmain.NewFrame()
	t.Cleanup(func() { frame.Close() })
	frame.SetResource("pGia", src)
	if arrivals {
		gia.Abc_FrameMiniAigSetCiArrivals(frame, []int32{2, 7, 1})
	}
	if required {
		gia.Abc_FrameMiniAigSetCoRequireds(frame, []int32{30, 31})
	}
	return src
}

func TestMappingDelayTraceSourceContext(t *testing.T) {
	for _, withLibrary := range []bool{false, true} {
		for mode := 0; mode < 4; mode++ {
			t.Run(fmt.Sprintf("library%t/timing%d", withLibrary, mode), func(t *testing.T) {
				src := driverCallbackReviewGraph(t, mode&1 != 0, mode&2 != 0)
				var pars If_Par_t
				Gia_ManSetIfParsDefault(&pars)
				pars.NLutSize, pars.FTruth, pars.FDeriveLuts, pars.FVerboseTrace = 4, 1, 1, 1
				lib := &If_LibLut_t{LutMax: 4}
				for size := 1; size <= 4; size++ {
					lib.pLutAreas[size] = 1
					lib.pLutDelays[size][0] = 1
				}
				if withLibrary {
					pars.PLutLib = lib
				}
				beforeLib := *lib
				var mappedManager *If_Man_t
				writes, traces := 0, 0
				pars.pFuncWrite = func(manager *If_Man_t) int32 {
					writes++
					mappedManager = manager
					if manager.stopped || manager.pPars != &pars {
						t.Error("write callback lost its live mapper")
					}
					if (pars.pTimesArr != nil) != (mode&1 != 0) || (pars.pTimesReq != nil) != (mode&2 != 0) {
						t.Error("write callback missing mapping constraints")
					}
					return 0
				}
				pars.SetDelayTrace(func(result *Gia_Man_t) {
					traces++
					if mappedManager == nil || !mappedManager.stopped {
						t.Error("delay trace ran before mapper cleanup")
					}
					if pars.pTimesArr != nil || pars.pTimesReq != nil {
						t.Error("delay trace retained mapper-owned timing arrays")
					}
					if withLibrary {
						if result.PLutLib != lib {
							t.Error("delay trace missing selected LUT library")
						}
					} else if result.PLutLib != nil {
						t.Error("delay trace acquired an unrelated LUT library")
					}
					if name, spec := result.Identity(); name != "callback-review" || spec != "callback-spec" {
						t.Error("delay trace ran before identity transfer")
					}
				})
				var result *Gia_Man_t
				output := dsdReportCapture(t, func() { result = Gia_ManPerformMappingInt(src, &pars) })
				if result == nil {
					t.Fatalf("mapping failed: %s", output)
				}
				defer Gia_ManStop(result)
				if writes != 1 || traces != 1 || pars.pTimesArr != nil || pars.pTimesReq != nil || result.PLutLib != nil || *lib != beforeLib {
					t.Fatal("callback count or final ownership/context mismatch")
				}
				for x := 0; x < 8; x++ {
					want := x&1 != 0 && x&2 == 0 && x&4 != 0
					got, err := result.Evaluate([]bool{x&1 != 0, x&2 != 0, x&4 != 0})
					if err != nil || !slices.Equal(got, []bool{want, !want}) {
						t.Fatalf("assignment%d got%v err%v", x, got, err)
					}
				}
			})
		}
	}
}

func TestMappingDelayTraceCanReuseParameters(t *testing.T) {
	src := driverCallbackReviewGraph(t, true, true)
	var pars If_Par_t
	Gia_ManSetIfParsDefault(&pars)
	pars.NLutSize, pars.FTruth, pars.FDeriveLuts, pars.FVerboseTrace = 4, 1, 1, 1
	var outer, nested *Gia_Man_t
	t.Cleanup(func() {
		if outer != nil {
			Gia_ManStop(outer)
		}
		if nested != nil {
			Gia_ManStop(nested)
		}
	})
	calls := 0
	pars.SetDelayTrace(func(result *Gia_Man_t) {
		calls++
		outer = result
		// Reporting is after mapper teardown in C. Reuse its now-empty timing
		// slots for a second mapping, without recursively enabling its reporter.
		pars.FVerboseTrace = 0
		nested = Gia_ManPerformMappingInt(src, &pars)
		pars.FVerboseTrace = 1
	})
	var failure any
	func() {
		defer func() { failure = recover() }()
		result := Gia_ManPerformMappingInt(src, &pars)
		if result != outer {
			t.Error("reporter received another result")
		}
	}()
	if failure != nil {
		t.Fatalf("post-mapping parameter reuse failed: %v", failure)
	}
	if calls != 1 || outer == nil || nested == nil || pars.pTimesArr != nil || pars.pTimesReq != nil {
		t.Fatal("nested reporting lifecycle mismatch")
	}
	for x := 0; x < 8; x++ {
		input := []bool{x&1 != 0, x&2 != 0, x&4 != 0}
		want := x&1 != 0 && x&2 == 0 && x&4 != 0
		for _, result := range []*Gia_Man_t{outer, nested} {
			got, err := result.Evaluate(input)
			if err != nil || !slices.Equal(got, []bool{want, !want}) {
				t.Fatalf("assignment%d got%v err%v", x, got, err)
			}
		}
	}
}

func TestMappingDelayTracePanicClearsBorrowedContext(t *testing.T) {
	src := driverCallbackReviewGraph(t, true, true)
	var pars If_Par_t
	Gia_ManSetIfParsDefault(&pars)
	pars.NLutSize, pars.FTruth, pars.FDeriveLuts, pars.FVerboseTrace = 4, 1, 1, 1
	lib := &If_LibLut_t{LutMax: 4}
	for size := 1; size <= 4; size++ {
		lib.pLutAreas[size] = 1
		lib.pLutDelays[size][0] = 1
	}
	pars.PLutLib = lib
	var escaped *Gia_Man_t
	var manager *If_Man_t
	pars.pFuncWrite = func(p *If_Man_t) int32 { manager = p; return 0 }
	pars.SetDelayTrace(func(result *Gia_Man_t) {
		escaped = result
		if result.PLutLib != lib {
			t.Error("reporter lost its borrowed library")
		}
		panic("report failed")
	})
	var failure any
	dsdReportCapture(t, func() { func() { defer func() { failure = recover() }(); Gia_ManPerformMappingInt(src, &pars) }() })
	if escaped != nil {
		defer Gia_ManStop(escaped)
	}
	if failure != "report failed" {
		t.Fatalf("panic%v want reporter failure", failure)
	}
	if escaped == nil || escaped.PLutLib != nil || manager == nil || !manager.stopped || pars.pTimesArr != nil || pars.pTimesReq != nil {
		t.Fatal("panic retained borrowed context or mapper state")
	}
	// The caller's library and timing inputs remain reusable after reporting.
	pars.FVerboseTrace = 0
	var result *Gia_Man_t
	dsdReportCapture(t, func() { result = Gia_ManPerformMappingInt(src, &pars) })
	if result == nil {
		t.Fatal("mapping could not retry after reporter panic")
	}
	defer Gia_ManStop(result)
	for x := 0; x < 8; x++ {
		input := []bool{x&1 != 0, x&2 != 0, x&4 != 0}
		want := input[0] && !input[1] && input[2]
		got, err := result.Evaluate(input)
		if err != nil || !slices.Equal(got, []bool{want, !want}) {
			t.Fatalf("assignment%d got%v err%v", x, got, err)
		}
	}
}

func TestMappingDelayTraceDefaultAndLibraryTiming(t *testing.T) {
	for _, custom := range []bool{false, true} {
		for _, withLibrary := range []bool{false, true} {
			t.Run(fmt.Sprintf("custom%t/library%t", custom, withLibrary), func(t *testing.T) {
				src := driverCallbackReviewGraph(t, false, false)
				var pars If_Par_t
				Gia_ManSetIfParsDefault(&pars)
				pars.NLutSize, pars.FTruth, pars.FDeriveLuts, pars.FVerboseTrace = 4, 1, 1, 1
				if withLibrary {
					lib := &If_LibLut_t{LutMax: 4}
					for size := 1; size <= 4; size++ {
						lib.pLutAreas[size] = 1
						lib.pLutDelays[size][0] = 7
					}
					pars.PLutLib = lib
				}
				calls := 0
				var measured float32
				if custom {
					pars.SetDelayTrace(func(result *Gia_Man_t) { calls++; measured = gia.Gia_ManDelayTraceLutPrint(result, 1) })
				}
				var result *Gia_Man_t
				var failure any
				output := dsdReportCapture(t, func() {
					func() { defer func() { failure = recover() }(); result = Gia_ManPerformMappingInt(src, &pars) }()
				})
				if failure != nil {
					t.Fatalf("verbose mapping failed: %v", failure)
				}
				if result == nil {
					t.Fatal("mapping returned no graph")
				}
				defer Gia_ManStop(result)
				want, model := float32(1), "unit-delay"
				if withLibrary {
					want, model = 7, "LUT library"
				}
				expected := fmt.Sprintf("Max delay = %6.2f. Delay trace using %s model:\n", want, model)
				if strings.Count(output, expected) != 1 {
					t.Fatalf("trace lacks%q: %q", expected, output)
				}
				if custom && (calls != 1 || measured != want) {
					t.Fatalf("reporter calls%d delay%g want%g", calls, measured, want)
				}
				if result.PLutLib != nil || pars.pTimesArr != nil || pars.pTimesReq != nil {
					t.Fatal("reporting retained borrowed state")
				}
				for x := 0; x < 8; x++ {
					input := []bool{x&1 != 0, x&2 != 0, x&4 != 0}
					expected := input[0] && !input[1] && input[2]
					got, err := result.Evaluate(input)
					if err != nil || !slices.Equal(got, []bool{expected, !expected}) {
						t.Fatalf("assignment%d got%v err%v", x, got, err)
					}
				}
			})
		}
	}
}
