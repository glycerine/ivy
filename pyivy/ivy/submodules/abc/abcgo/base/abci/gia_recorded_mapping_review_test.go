package abci

import (
	"fmt"
	"slices"
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	abcmain "github.com/glycerine/abcgo/base/main"
	ifpkg "github.com/glycerine/abcgo/map/if"
)

func recordedMappingReviewLibrary(t *testing.T) {
	t.Helper()
	previous := s_pMan3
	s_pMan3 = nil
	lib := gia.Gia_ManStart(32)
	inputs := make([]int32, 6)
	for i := range inputs {
		inputs[i] = gia.Gia_ManAppendCi(lib)
	}
	var roots []int32
	for phase := int32(0); phase < 4; phase++ {
		roots = append(roots, gia.Gia_ManAppendAnd(lib, inputs[0]^(phase&1), inputs[1]^((phase>>1)&1)))
	}
	// Include every phase so this source-shaped library contains the canonical
	// representative selected for any two-input AND cut, without a C oracle.
	for _, root := range roots {
		gia.Gia_ManAppendCo(lib, root)
		gia.Gia_ManAppendCo(lib, root^1)
	}
	Abc_NtkRecStart3(lib, 6, 8, 0, 0)
	t.Cleanup(func() { Abc_NtkRecStop3(); s_pMan3 = previous })
}

func TestGiaRecordedLibraryMappingUsesSourceBackend(t *testing.T) {
	recordedMappingReviewLibrary(t)
	p := gia.Gia_ManStart(8)
	defer gia.Gia_ManStop(p)
	a, b := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
	root := gia.Gia_ManAppendAnd(p, a, b)
	gia.Gia_ManAppendCo(p, root)
	var pars ifpkg.If_Par_t
	ifpkg.Gia_ManSetIfParsDefault(&pars)
	pars.NLutSize = 6
	pars.FUserRecLib, pars.FTruth, pars.FUsePerm, pars.FCutMin = 1, 1, 1, 1
	pars.FExpRed = 0
	result := ifpkg.Gia_ManPerformMapping(p, &pars)
	if result == nil {
		t.Fatal("recorded mapping failed")
	}
	defer gia.Gia_ManStop(result)
	if pars.FinalDelay != 1 || pars.FinalArea != 1 || result.VMapping != nil {
		t.Fatalf("delay%f area%f map%v", pars.FinalDelay, pars.FinalArea, result.VMapping)
	}
	for x := 0; x < 4; x++ {
		inputs := []bool{x&1 != 0, x&2 != 0}
		got, err := result.Evaluate(inputs)
		if err != nil || !slices.Equal(got, []bool{inputs[0] && inputs[1]}) {
			t.Fatalf("input%d: %v (%v)", x, got, err)
		}
	}
}

func TestGiaRecordedLibraryMappingPhaseTimingAndRegisters(t *testing.T) {
	recordedMappingReviewLibrary(t)
	// Plain AIG results have no LUT mapping, so the source fHashMapping
	// post-pass is inapplicable. Exercise both source delay/area modes instead.
	for phase := int32(0); phase < 64; phase++ {
		for arrivalCase, arrivals := range [][]int32{{0, 0, 0}, {1, 4, 2}, {0, 0, 8}} {
			for regs := int32(0); regs < 2; regs++ {
				for area := int32(0); area < 2; area++ {
					t.Run(fmt.Sprintf("phase%d/times%d/regs%d/area%d", phase, arrivalCase, regs, area), func(t *testing.T) {
						p := gia.Gia_ManStart(16)
						frame := abcmain.NewFrame()
						frame.SetResource("pGia", p)
						defer frame.Close()
						a, b, c := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
						x := gia.Gia_ManAppendAnd(p, a^(phase&1), b^((phase>>1)&1))
						y := gia.Gia_ManAppendAnd(p, x^((phase>>2)&1), c^((phase>>3)&1))
						for _, lit := range []int32{x ^ ((phase >> 4) & 1), y ^ ((phase >> 5) & 1), 1, a ^ 1} {
							gia.Gia_ManAppendCo(p, lit)
						}
						gia.Gia_ManSetRegNum(p, regs)
						p.SetIdentity("recorded-map", "recorded-map.aig")
						gia.Abc_FrameMiniAigSetCiArrivals(frame, arrivals)
						var pars ifpkg.If_Par_t
						ifpkg.Gia_ManSetIfParsDefault(&pars)
						pars.NLutSize = 6
						pars.FUserRecLib, pars.FTruth, pars.FUsePerm, pars.FCutMin = 1, 1, 1, 1
						pars.FExpRed, pars.FArea = 0, area
						result := ifpkg.Gia_ManPerformMapping(p, &pars)
						if result == nil {
							t.Fatal("recorded mapping failed")
						}
						defer gia.Gia_ManStop(result)
						delay := max(max(arrivals[0], arrivals[1])+1, arrivals[2]) + 1
						if pars.FinalDelay != float32(delay) || pars.FinalArea != 2 || result.VMapping != nil || gia.Gia_ManRegNum(result) != regs {
							t.Fatalf("delay%f want%d area%f regs%d", pars.FinalDelay, delay, pars.FinalArea, gia.Gia_ManRegNum(result))
						}
						if name, spec := result.Identity(); name != "recorded-map" || spec != "recorded-map.aig" {
							t.Fatal("identity lost")
						}
						if got, _, _, _ := result.MappingInputs(); !slices.Equal(got, arrivals) {
							t.Fatal("arrival metadata not transferred")
						}
						for assignment := 0; assignment < 8; assignment++ {
							inputs := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0}
							xv := (inputs[0] != (phase&1 != 0)) && (inputs[1] != (phase&2 != 0))
							yv := (xv != (phase&4 != 0)) && (inputs[2] != (phase&8 != 0))
							want := []bool{xv != (phase&16 != 0), yv != (phase&32 != 0), true, !inputs[0]}
							got, err := result.Evaluate(inputs)
							if err != nil || !slices.Equal(got, want) {
								t.Fatalf("input%d: got%v want%v err%v", assignment, got, want, err)
							}
						}
					})
				}
			}
		}
	}
}
