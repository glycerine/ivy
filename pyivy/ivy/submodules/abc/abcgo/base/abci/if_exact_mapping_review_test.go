package abci

import (
	"fmt"
	"slices"
	"testing"

	ifpkg "github.com/glycerine/abcgo/map/if"
)

func exactMappingReviewStore(t *testing.T) {
	t.Helper()
	exactMappingReviewStoreMode(t, 1)
}

func exactMappingReviewStoreMode(t *testing.T, aig int32) {
	t.Helper()
	previous := s_pSesStore
	s_pSesStore = nil
	Abc_ExactStart(10000, aig, 0, 0, "")
	t.Cleanup(func() { Abc_ExactStop(""); s_pSesStore = previous })
}

func TestExactIfMappingUsesSourceBackend(t *testing.T) {
	exactMappingReviewStore(t)
	var pars ifpkg.If_Par_t
	ifpkg.If_ManSetDefaultPars(&pars)
	pars.NLutSize, pars.FTruth, pars.FUsePerm, pars.FUserSesLib = 6, 1, 1, 1
	pars.FExpRed = 0
	p := ifpkg.If_ManStart(&pars)
	defer ifpkg.If_ManStop(p)
	a, b := ifpkg.If_ManCreateCi(p), ifpkg.If_ManCreateCi(p)
	root := ifpkg.If_ManCreateAnd(p, a, b)
	ifpkg.If_ManCreateCo(p, root)
	if ifpkg.If_ManPerformMapping(p) != 1 {
		t.Fatal("exact-delay IF mapping failed")
	}
	cut := ifpkg.If_ObjCutBest(root)
	if pars.FinalDelay != 1 || pars.FinalArea != 0 || ifpkg.If_CutCost(cut) != 0 || !slices.Equal(ifpkg.If_CutPerm(cut), []byte{1, 1}) {
		t.Fatalf("delay%f area%f cost%d pins%v", pars.FinalDelay, pars.FinalArea, ifpkg.If_CutCost(cut), ifpkg.If_CutPerm(cut))
	}
	if s_pSesStore.nCutCount < 2 || s_pSesStore.nSatCalls == 0 || s_pSesStore.nCacheHits == 0 {
		t.Fatal("mapping did not synthesize and reuse the real SAT cache")
	}
}

func TestExactIfMappingPhasesTimingAndGateModes(t *testing.T) {
	for _, aig := range []int32{0, 1} {
		t.Run(fmt.Sprintf("aig%d", aig), func(t *testing.T) {
			exactMappingReviewStoreMode(t, aig)
			for _, xor := range []bool{false, true} {
				for phase := int32(0); phase < 4; phase++ {
					for times, arrivals := range [][]float32{{0, 0}, {2, 5}, {5, 1}} {
						t.Run(fmt.Sprintf("xor%t/phase%d/times%d", xor, phase, times), func(t *testing.T) {
							var pars ifpkg.If_Par_t
							ifpkg.If_ManSetDefaultPars(&pars)
							pars.NLutSize, pars.FTruth, pars.FUsePerm, pars.FUserSesLib = 6, 1, 1, 1
							pars.FExpRed = 0
							pars.SetArrivalTimes(arrivals)
							p := ifpkg.If_ManStart(&pars)
							defer ifpkg.If_ManStop(p)
							a, b := ifpkg.If_ManCreateCi(p), ifpkg.If_ManCreateCi(p)
							a, b = ifpkg.If_NotCond(a, phase&1), ifpkg.If_NotCond(b, (phase>>1)&1)
							var driver *ifpkg.If_Obj_t
							if xor {
								driver = ifpkg.If_ManCreateXor(p, a, b)
							} else {
								driver = ifpkg.If_ManCreateAnd(p, a, b)
							}
							ifpkg.If_ManCreateCo(p, driver)
							if ifpkg.If_ManPerformMapping(p) != 1 {
								t.Fatal("exact IF mapping failed")
							}
							root := ifpkg.If_Regular(driver)
							cut := ifpkg.If_ObjCutBest(root)
							pin := byte(1)
							if xor && aig != 0 {
								pin = 2
							}
							wantDelay := max(arrivals[0], arrivals[1]) + float32(pin)
							if pars.FinalDelay != wantDelay || pars.FinalArea != 0 || ifpkg.If_CutCost(cut) != 0 || !slices.Equal(ifpkg.If_CutPerm(cut), []byte{pin, pin}) {
								t.Fatalf("delay%f want%f area%f cost%d pins%v", pars.FinalDelay, wantDelay, pars.FinalArea, ifpkg.If_CutCost(cut), ifpkg.If_CutPerm(cut))
							}
							truth := ifpkg.If_CutTruthWordSlice(p, cut)
							for assignment := 0; assignment < 4; assignment++ {
								av, bv := (assignment&1 != 0) != (phase&1 != 0), (assignment&2 != 0) != (phase&2 != 0)
								want := av && bv
								if xor {
									want = av != bv
								}
								got := (truth[0]>>assignment&1 != 0) != (ifpkg.If_IsComplement(driver) != 0)
								if got != want {
									t.Fatalf("selected function input%d: got%t want%t", assignment, got, want)
								}
							}
						})
					}
				}
			}
			if s_pSesStore.nSatCalls == 0 || s_pSesStore.nCacheHits == 0 {
				t.Fatal("exact mapper did not use synthesis/cache")
			}
		})
	}
}

func TestExactIfPinAdapterPreservesInactiveBytes(t *testing.T) {
	exactMappingReviewStore(t)
	truth, arrival, cost := uint64(0xaaaaaaaaaaaaaaaa), int32(7), int32(-1)
	pins := []byte{255, 128, 200, 1, 2, 3, 4, 5, 254, 129}
	before := slices.Clone(pins)
	if delay := ifExactDelayCost(&truth, 1, &arrival, pins, &cost, 9); delay != 7 || cost != 0 || pins[0] != 0 || !slices.Equal(pins[1:], before[1:]) {
		t.Fatalf("adapter delay%d cost%d pins%v", delay, cost, pins)
	}
	if delay := ifExactDelayCost(&truth, 0, nil, nil, &cost, 0); delay != 0 || cost != 0 {
		t.Fatal("zero-support adapter accessed nonexistent pins")
	}
}

func TestExactDelayCostAcceptsCompactTruthWords(t *testing.T) {
	for _, inverted := range []bool{false, true} {
		t.Run(fmt.Sprintf("inverted%t", inverted), func(t *testing.T) {
			exactMappingReviewStore(t)
			truth := new(word) // Exactly one caller-owned word for two variables.
			*truth = 0x8888888888888888
			if inverted {
				*truth = ^*truth
			}
			original := *truth
			arrivals := []int32{0, 0}
			pins := make([]int8, 2)
			var cost int32
			if delay := Abc_ExactDelayCost(truth, 2, &arrivals[0], pins, &cost, 1); delay != 1 || cost != 1 || *truth != original || !slices.Equal(pins, []int8{1, 1}) {
				t.Fatalf("compact truth: delay%d cost%d truth%x pins%v", delay, cost, *truth, pins)
			}
		})
	}
}
