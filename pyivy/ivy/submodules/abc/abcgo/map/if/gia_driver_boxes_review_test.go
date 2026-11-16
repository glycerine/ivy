package if_

import (
	"bytes"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	abcmain "github.com/glycerine/abcgo/base/main"
	"github.com/glycerine/abcgo/misc/tim"
)

func driverBoxesReviewGraph(t *testing.T, signs int, normalized, future bool) (*Gia_Man_t, *abcmain.Frame) {
	t.Helper()
	src := Gia_ManStart(32)
	src.SetIdentity("outer-box-review", "outer-box-spec")
	a, b, c, d := Gia_ManAppendCi(src), Gia_ManAppendCi(src), Gia_ManAppendCi(src), Gia_ManAppendCi(src)
	first := b ^ int32(signs&1)
	if future {
		first = d
	}
	x := gia.Gia_ManAppendAnd(src, a, first)
	y := gia.Gia_ManAppendAnd(src, c, b^int32(signs>>1&1))
	z := gia.Gia_ManAppendAnd(src, d^int32(signs>>2&1), c)
	Gia_ManAppendCo(src, x)
	Gia_ManAppendCo(src, y)
	Gia_ManAppendCo(src, z^int32(signs>>3&1))
	timing := tim.Tim_ManStart(4, 3)
	tim.Tim_ManSetDelayTables(timing, [][]float32{{0, 1, 1, 1.5}, {1, 1, 1, 2.25}})
	tim.Tim_ManCreateBox(timing, 0, 1, 2, 1, 0, 0)
	tim.Tim_ManCreateBox(timing, 1, 1, 3, 1, 1, 0)
	tim.Tim_ManInitPiArrival(timing, 0, 1.25)
	tim.Tim_ManInitPiArrival(timing, 1, 4.5)
	tim.Tim_ManInitPoRequiredAll(timing, 50)
	extra := Gia_ManStart(4)
	input := Gia_ManAppendCi(extra)
	Gia_ManAppendCo(extra, input)
	Gia_ManAppendCo(extra, input^1)
	src.SetBoxTiming(timing, extra)
	if !normalized {
		work := gia.Gia_ManDupUnnormalize(src)
		if work == nil {
			t.Fatal("cannot construct interleaved box graph")
		}
		gia.Gia_ManTransferTiming(work, src)
		Gia_ManStop(src)
		src = work
	}
	frame := abcmain.NewFrame()
	frame.SetResource("pGia", src)
	t.Cleanup(func() { frame.Close() })
	gia.Abc_FrameMiniAigSetCiArrivals(frame, []int32{30, 31, 32, 33})
	gia.Abc_FrameMiniAigSetCoRequireds(frame, []int32{60, 61, 62})
	src.SetMappingOutputArrivals([]int32{90, 91, 92})
	src.VNamesIn, src.VNamesOut = gia.Vec_PtrAlloc(4), gia.Vec_PtrAlloc(3)
	for _, name := range []string{"a", "b", "box1", "box2"} {
		gia.Vec_PtrPush(src.VNamesIn, name)
	}
	for _, name := range []string{"box1-input", "box2-input", "out"} {
		gia.Vec_PtrPush(src.VNamesOut, name)
	}
	src.PCellStr = "obsolete-cell"
	src.VConfigs = gia.Vec_IntAllocArrayCopy([]int32{0, 1}, 2)
	return src, frame
}

func TestMappingDriverBoxedNormalizationAndOwnership(t *testing.T) {
	for signs := 0; signs < 16; signs++ {
		for _, normalized := range []bool{false, true} {
			for hash := int32(0); hash < 2; hash++ {
				for mode := 0; mode < 5; mode++ {
					// The source non-boxed hash branch requires LUT mapping; AIG
					// balancing produces no LUT records for that combination.
					if mode >= 3 && !normalized && hash != 0 {
						continue
					}
					t.Run(fmt.Sprintf("signs%d/normalized%t/hash%d/mode%d", signs, normalized, hash, mode), func(t *testing.T) {
						src, frame := driverBoxesReviewGraph(t, signs, normalized, false)
						timing, extra := src.BoxTiming()
						before := tim.Tim_ManSave(timing, 0)
						namesIn, namesOut := src.VNamesIn, src.VNamesOut
						var pars If_Par_t
						Gia_ManSetIfParsDefault(&pars)
						pars.NLutSize, pars.FHashMapping = 4, hash
						if mode != 0 {
							pars.FTruth, pars.FDeriveLuts = 1, 1
						}
						if mode == 2 {
							pars.FEnableCheck07 = 1
						}
						wantTimes := []float32{5.5, 8, 11.25}
						if mode >= 3 {
							pars.FExpRed = 0
							wantTimes = []float32{1, 3, 5}
							if mode == 3 {
								pars.FDelayOpt = 1
							} else {
								pars.FDsdBalance, pars.FUseDsd = 1, 1
								dsd := If_DsdManAlloc(4, 0)
								defer If_DsdManFree(dsd, 0)
								pars.SetDsdManager(dsd)
							}
						}
						writes := 0
						pars.pFuncWrite = func(p *If_Man_t) int32 {
							writes++
							for i, want := range wantTimes {
								if got := If_ObjArrTime(If_ObjFanin0(If_ManCo(p, int32(i)))); got != want {
									t.Errorf("CO%d arrival %g, want %g", i, got, want)
								}
							}
							return 0
						}
						var result *Gia_Man_t
						output := dsdReportCapture(t, func() { result = Gia_ManPerformMapping(src, &pars) })
						if result == nil {
							t.Fatalf("boxed mapping failed: %s", output)
						}
						defer Gia_ManStop(result)
						if writes != 1 || strings.Contains(output, "Gia_ManMappingVerify:") {
							t.Fatalf("mapping/verification failed: %q", output)
						}
						if (gia.Gia_ManIsNormalized(result) != 0) != (normalized || hash != 0 || mode >= 3) {
							t.Fatal("unexpected final object ordering")
						}
						gotTiming, gotExtra := result.BoxTiming()
						if gotTiming != timing || gotExtra != extra || !bytes.Equal(before, tim.Tim_ManSave(timing, 0)) {
							t.Fatal("timing/model ownership or source timing values changed")
						}
						if oldTiming, oldExtra := src.BoxTiming(); oldTiming != nil || oldExtra != nil {
							t.Fatal("source retained box metadata")
						}
						if result.VNamesIn != namesIn || result.VNamesOut != namesOut || src.VNamesIn != nil || src.VNamesOut != nil {
							t.Fatal("name ownership did not move")
						}
						arrivals, required, _, _ := result.MappingInputs()
						if !slices.Equal(arrivals, []int32{30, 31, 32, 33}) || !slices.Equal(required, []int32{60, 61, 62}) {
							t.Fatal("plain timing vectors lost")
						}
						oldArr, oldReq, _, _ := src.MappingInputs()
						if oldArr != nil || oldReq != nil || result.VConfigs != nil || result.PCellStr != "" || src.VConfigs != nil || src.PCellStr != "" {
							t.Fatal("timing ownership or configuration replacement failed")
						}
						// Replacing the frame slot closes the old source, which no
						// longer owns the transferred metadata.
						frame.SetResource("pGia", result)
						wantArrivals := []int32{int32(wantTimes[0]), int32(wantTimes[1]), int32(wantTimes[2])}
						if got := gia.Abc_FrameMiniAigReadCoArrivals(frame); !slices.Equal(got, wantArrivals) {
							t.Fatalf("CO arrivals %v", got)
						}
						if result.MappedDelay != int32(pars.FinalDelay) || result.MappedArea != int32(pars.FinalArea) {
							t.Fatal("mapping summary not copied")
						}
						if (result.VMapping == nil) != (mode >= 3) {
							t.Fatal("unexpected mapping representation")
						}
						if mode == 2 {
							data := gia.Vec_IntArray(result.VPacking)
							if !slices.Equal(data[:1], []int32{3}) || len(data) != 7 {
								t.Fatalf("packing %v", data)
							}
							for i := 1; i < len(data); i += 2 {
								if data[i] != 1 || gia.Gia_ObjIsLut(result, data[i+1]) == 0 {
									t.Fatal("packing lost LUT identity")
								}
							}
						} else if result.VPacking != nil {
							t.Fatal("unexpected packing")
						}
						if name, spec := result.Identity(); name != "outer-box-review" || spec != "outer-box-spec" {
							t.Fatal("identity lost")
						}
						for x := 0; x < 16; x++ {
							a, b, c, d := x&1 != 0, x&2 != 0, x&4 != 0, x&8 != 0
							want := []bool{a && (b != (signs&1 != 0)), c && (b != (signs&2 != 0)), ((d != (signs&4 != 0)) && c) != (signs&8 != 0)}
							got, err := result.Evaluate([]bool{a, b, c, d})
							if err != nil || !slices.Equal(got, want) {
								t.Fatalf("input%d: got%v want%v err%v", x, got, want, err)
							}
						}
					})
				}
			}
		}
	}
}

func TestMappingDriverBoxOrderFailureRetainsInput(t *testing.T) {
	src, frame := driverBoxesReviewGraph(t, 0, true, true)
	timing, extra := src.BoxTiming()
	before := tim.Tim_ManSave(timing, 0)
	configs, namesIn, namesOut := src.VConfigs, src.VNamesIn, src.VNamesOut
	for hash := int32(0); hash < 2; hash++ {
		var pars If_Par_t
		Gia_ManSetIfParsDefault(&pars)
		pars.NLutSize, pars.FHashMapping = 4, hash
		pars.pFuncWrite = func(*If_Man_t) int32 { t.Fatal("invalid box order reached mapping"); return 0 }
		output := dsdReportCapture(t, func() {
			if result := Gia_ManPerformMapping(src, &pars); result != nil {
				Gia_ManStop(result)
				t.Fatal("accepted a future-box dependency")
			}
		})
		if output != "Boxes are not in topological order.\n" {
			t.Fatalf("unexpected failure diagnostic %q", output)
		}
		gotTiming, gotExtra := src.BoxTiming()
		if gotTiming != timing || gotExtra != extra || !bytes.Equal(before, tim.Tim_ManSave(timing, 0)) {
			t.Fatal("failed ordering changed box metadata")
		}
		if src.VConfigs != configs || src.PCellStr != "obsolete-cell" || src.VNamesIn != namesIn || src.VNamesOut != namesOut {
			t.Fatal("failed ordering discarded source configuration/names")
		}
		arrivals, required, _, _ := src.MappingInputs()
		if !slices.Equal(arrivals, []int32{30, 31, 32, 33}) || !slices.Equal(required, []int32{60, 61, 62}) || !slices.Equal(gia.Abc_FrameMiniAigReadCoArrivals(frame), []int32{90, 91, 92}) {
			t.Fatal("failed ordering changed source timing vectors")
		}
		if pars.pTimesArr != nil || pars.pTimesReq != nil {
			t.Fatal("failed ordering installed mapping constraints")
		}
		for x := 0; x < 16; x++ {
			a, b, c, d := x&1 != 0, x&2 != 0, x&4 != 0, x&8 != 0
			want := []bool{a && d, c && b, d && c}
			got, err := src.Evaluate([]bool{a, b, c, d})
			if err != nil || !slices.Equal(got, want) {
				t.Fatalf("input%d after failed ordering: got%v want%v err%v", x, got, want, err)
			}
		}
	}
}

func TestMappingDriverSequentialBoxShuffle(t *testing.T) {
	for signs := 0; signs < 16; signs++ {
		for _, normalized := range []bool{false, true} {
			for hash := int32(0); hash < 2; hash++ {
				t.Run(fmt.Sprintf("signs%d/normalized%t/hash%d", signs, normalized, hash), func(t *testing.T) {
					// Source normalization expects PI + FO + box-CI order on entry,
					// then moves FOs last. Here b is the FO, before the two box CIs.
					src, _ := driverBoxesReviewGraph(t, signs, normalized, false)
					Gia_ManSetRegNum(src, 1)
					var pars If_Par_t
					Gia_ManSetIfParsDefault(&pars)
					pars.NLutSize, pars.FHashMapping, pars.FTruth, pars.FDeriveLuts = 4, hash, 1, 1
					var result *Gia_Man_t
					output := dsdReportCapture(t, func() { result = Gia_ManPerformMapping(src, &pars) })
					if result == nil {
						t.Fatalf("sequential boxed mapping failed: %s", output)
					}
					defer Gia_ManStop(result)
					if Gia_ManRegNum(result) != 1 || Gia_ManCiNum(result) != 4 || Gia_ManCoNum(result) != 3 {
						t.Fatal("sequential interface changed")
					}
					warning := "Warning: Shuffled CI order to be correct sequential AIG.\n"
					wantWarnings := 0
					if normalized {
						wantWarnings = 1
					}
					if strings.Count(output, warning) != wantWarnings {
						t.Errorf("shuffle warnings %q, want %d", output, wantWarnings)
					}
					if strings.Contains(output, "Gia_ManMappingVerify:") {
						t.Errorf("invalid mapping: %q", output)
					}
					for x := 0; x < 16; x++ {
						values := []bool{x&1 != 0, x&2 != 0, x&4 != 0, x&8 != 0}
						a, b, c, d := values[0], values[1], values[2], values[3]
						want := []bool{a && (b != (signs&1 != 0)), c && (b != (signs&2 != 0)), ((d != (signs&4 != 0)) && c) != (signs&8 != 0)}
						if normalized {
							values = []bool{a, c, d, b}
						}
						got, err := result.Evaluate(values)
						if err != nil || !slices.Equal(got, want) {
							t.Fatalf("input%d: got%v want%v err%v", x, got, want, err)
						}
					}
				})
			}
		}
	}
}

func TestBalanceWrappersBoxedMetadataPolicy(t *testing.T) {
	balanceReviewIsolateFrame(t)
	for signs := 0; signs < 16; signs++ {
		for _, normalized := range []bool{false, true} {
			for kind := 0; kind < 4; kind++ {
				for _, relax := range []int32{0, 20} {
					t.Run(fmt.Sprintf("signs%d/normalized%t/kind%d/relax%d", signs, normalized, kind, relax), func(t *testing.T) {
						src, frame := driverBoxesReviewGraph(t, signs, normalized, false)
						timing, extra := src.BoxTiming()
						before := tim.Tim_ManSave(timing, 0)
						configs, namesIn, namesOut := src.VConfigs, src.VNamesIn, src.VNamesOut
						var result *Gia_Man_t
						dsdReportCapture(t, func() {
							switch kind {
							case 0:
								result = Gia_ManPerformSopBalance(src, 8, relax, 0)
							case 1:
								result = Gia_ManPerformDsdBalance(src, 4, 8, relax, 0)
							case 2:
								result = gia.Gia_ManPerformSopBalance(src, 8, relax, 0)
							case 3:
								result = gia.Gia_ManPerformDsdBalance(src, 4, 8, relax, 0)
							}
						})
						if result == nil {
							t.Fatal("balancing failed")
						}
						defer Gia_ManStop(result)
						gotTiming, gotExtra := result.BoxTiming()
						if gotTiming != timing || gotExtra != extra || !bytes.Equal(before, tim.Tim_ManSave(timing, 0)) {
							t.Fatal("balancing changed box timing/model ownership or values")
						}
						if result.VConfigs != configs || result.PCellStr != "obsolete-cell" || result.VNamesIn != namesIn || result.VNamesOut != namesOut {
							t.Fatal("balancing discarded configuration/name metadata")
						}
						if oldTiming, oldExtra := src.BoxTiming(); oldTiming != nil || oldExtra != nil || src.VConfigs != nil || src.PCellStr != "" || src.VNamesIn != nil || src.VNamesOut != nil {
							t.Fatal("source retained moved metadata")
						}
						arrivals, required, _, _ := result.MappingInputs()
						oldArrivals, oldRequired, _, _ := src.MappingInputs()
						if !slices.Equal(arrivals, []int32{30, 31, 32, 33}) || !slices.Equal(required, []int32{60, 61, 62}) || oldArrivals != nil || oldRequired != nil {
							t.Fatal("constraint metadata ownership changed")
						}
						if result.VMapping != nil || result.VPacking != nil || gia.Gia_ManIsNormalized(result) == 0 {
							t.Fatal("balancing returned an unexpected representation")
						}
						if name, spec := result.Identity(); name != "outer-box-review" || spec != "outer-box-spec" {
							t.Fatal("balancing lost identity")
						}
						frame.SetResource("pGia", result)
						if got := gia.Abc_FrameMiniAigReadCoArrivals(frame); !slices.Equal(got, []int32{90, 91, 92}) {
							t.Fatalf("balancing recomputed output arrival metadata: %v", got)
						}
						for x := 0; x < 16; x++ {
							a, b, c, d := x&1 != 0, x&2 != 0, x&4 != 0, x&8 != 0
							want := []bool{a && (b != (signs&1 != 0)), c && (b != (signs&2 != 0)), ((d != (signs&4 != 0)) && c) != (signs&8 != 0)}
							got, err := result.Evaluate([]bool{a, b, c, d})
							if err != nil || !slices.Equal(got, want) {
								t.Fatalf("input%d: got%v want%v err%v", x, got, want, err)
							}
						}
					})
				}
			}
		}
	}
}
