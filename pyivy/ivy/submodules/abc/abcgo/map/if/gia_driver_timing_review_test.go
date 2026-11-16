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

// Build the box inputs before its output CI, as required by MappingInt.
func driverTimingReviewGraph(t *testing.T, boxes, gates bool) (*Gia_Man_t, *tim.Tim_Man_t, *abcmain.Frame) {
	t.Helper()
	src := Gia_ManStart(16)
	src.SetIdentity("timing-review", "timing-spec")
	a, b := Gia_ManAppendCi(src), Gia_ManAppendCi(src)
	var c int32
	if !boxes {
		c = Gia_ManAppendCi(src)
	}
	first := a
	if gates {
		first = gia.Gia_ManAppendAnd(src, a, b^1)
	}
	Gia_ManAppendCo(src, first)
	if boxes {
		c = Gia_ManAppendCi(src)
	}
	last := c
	if gates {
		last = gia.Gia_ManAppendAnd(src, c, a)
	}
	Gia_ManAppendCo(src, last^1)
	timing := tim.Tim_ManStart(3, 2)
	tim.Tim_ManInitPiArrival(timing, 0, 1.5)
	tim.Tim_ManInitPiArrival(timing, 1, 7.25)
	tim.Tim_ManInitPiArrival(timing, 2, 4.75)
	tim.Tim_ManInitPoRequiredAll(timing, 40)
	if boxes {
		tim.Tim_ManSetDelayTables(timing, [][]float32{{0, 1, 1, 4.25}})
		tim.Tim_ManCreateBox(timing, 0, 1, 2, 1, 0, 0)
	}
	src.SetBoxTiming(timing, nil)
	frame := abcmain.NewFrame()
	frame.SetResource("pGia", src)
	t.Cleanup(func() { frame.Close() })
	return src, timing, frame
}

func TestMappingDriverTimingDuplication(t *testing.T) {
	for _, boxes := range []bool{false, true} {
		for mode := 0; mode < 4; mode++ {
			for constraints := 0; constraints < 4; constraints++ {
				t.Run(fmt.Sprintf("boxes%t/mode%d/constraints%d", boxes, mode, constraints), func(t *testing.T) {
					src, timing, frame := driverTimingReviewGraph(t, boxes, true)
					if constraints&1 != 0 {
						gia.Abc_FrameMiniAigSetCiArrivals(frame, []int32{20, 21, 22})
					}
					if constraints&2 != 0 {
						gia.Abc_FrameMiniAigSetCoRequireds(frame, []int32{50, 51})
					}
					src.SetMappingOutputArrivals([]int32{90, 91})
					before := tim.Tim_ManSave(timing, 0)
					var pars If_Par_t
					Gia_ManSetIfParsDefault(&pars)
					pars.NLutSize, pars.FTruth, pars.FDeriveLuts = 4, 1, 1
					switch mode {
					case 1:
						pars.FDelayOpt = 1
					case 2:
						pars.FDelayOptLut = 1
						// LUT balancing requires a physical structure and its DSD manager.
						pars.FUseDsd, pars.PLutStruct = 1, "44"
						dsd := If_DsdManAlloc(4, 4)
						defer If_DsdManFree(dsd, 0)
						pars.SetDsdManager(dsd)
					case 3:
						pars.FDsdBalance = 1
						pars.FUseDsd = 1
						dsd := If_DsdManAlloc(4, 0)
						defer If_DsdManFree(dsd, 0)
						pars.SetDsdManager(dsd)
					}
					var cloned *tim.Tim_Man_t
					writes := 0
					var want []float32
					if mode == 0 {
						want = []float32{8.25, 5.75}
						if boxes {
							want[1] = 13.5
						}
					} else {
						want = []float32{1, 1}
						if boxes {
							want[1] = 3
						}
					}
					pars.pFuncWrite = func(p *If_Man_t) int32 {
						writes++
						cloned = p.pManTim
						if cloned == nil || cloned == timing {
							t.Fatal("mapping borrowed source timing instead of duplicating")
						}
						if boxes {
							delay := float32(4.25)
							if mode != 0 {
								delay = 1
							}
							if got := tim.Tim_ManBoxDelayTable(cloned, 0)[3]; got != delay {
								t.Errorf("box delay %g, want %g", got, delay)
							}
						}
						for i, value := range want {
							if got := If_ObjArrTime(If_ObjFanin0(If_ManCo(p, int32(i)))); got != value {
								t.Errorf("CO%d arrival %g, want %g", i, got, value)
							}
						}
						return 0
					}
					dsdReportCapture(t, func() {
						result := Gia_ManPerformMappingInt(src, &pars)
						if result == nil {
							t.Fatal("timed mapping failed")
						}
						defer Gia_ManStop(result)
						for x := 0; x < 8; x++ {
							got, err := result.Evaluate([]bool{x&1 != 0, x&2 != 0, x&4 != 0})
							want := []bool{x&1 != 0 && x&2 == 0, !(x&4 != 0 && x&1 != 0)}
							if err != nil || !slices.Equal(got, want) {
								t.Fatalf("input%d: got%v want%v err%v", x, got, want, err)
							}
						}
					})
					if writes != 1 || tim.Tim_ManCiNum(cloned) != 0 || pars.pTimesArr != nil || pars.pTimesReq != nil {
						t.Fatal("temporary timing was not cleaned up")
					}
					if !bytes.Equal(before, tim.Tim_ManSave(timing, 0)) {
						t.Fatal("source timing changed")
					}
					wantInts := []int32{90, 91}
					if constraints != 0 {
						wantInts = []int32{int32(want[0]), int32(want[1])}
					}
					if got := gia.Abc_FrameMiniAigReadCoArrivals(frame); !slices.Equal(got, wantInts) {
						t.Fatalf("output arrivals %v, want %v", got, wantInts)
					}
				})
			}
		}
	}
}

func TestMappingDriverTimingSuppressesPowerSimulation(t *testing.T) {
	for _, boxes := range []bool{false, true} {
		t.Run(fmt.Sprint(boxes), func(t *testing.T) {
			// No AND cuts: isolate the driver's source warning/suppression rule
			// from the source power-cost routines' nonnil-vector precondition.
			src, timing, _ := driverTimingReviewGraph(t, boxes, false)
			before := tim.Tim_ManSave(timing, 0)
			var pars If_Par_t
			Gia_ManSetIfParsDefault(&pars)
			pars.NLutSize, pars.FPower = 4, 1
			pars.SetSwitchingProvider(func(*If_Man_t) { t.Fatal("timing manager did not suppress simulation") })
			writes := 0
			pars.pFuncWrite = func(p *If_Man_t) int32 {
				writes++
				if p.vSwitching != nil || p.dPower != 0 {
					t.Fatal("unexpected switching vector or power")
				}
				return 0
			}
			output := dsdReportCapture(t, func() {
				result := Gia_ManPerformMappingInt(src, &pars)
				if result == nil {
					t.Fatal("terminal mapping failed")
				}
				Gia_ManStop(result)
			})
			if writes != 1 || strings.Count(output, "Switching activity computation for designs with boxes is disabled.") != 1 {
				t.Fatalf("missing suppression warning: %q", output)
			}
			if !bytes.Equal(before, tim.Tim_ManSave(timing, 0)) {
				t.Fatal("source timing changed")
			}
		})
	}
}
