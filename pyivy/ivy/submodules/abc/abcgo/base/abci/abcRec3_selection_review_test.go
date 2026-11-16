package abci

import (
	"fmt"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	ifpkg "github.com/glycerine/abcgo/map/if"
	"github.com/glycerine/abcgo/misc/vec"
)

func recordedReviewIF(t *testing.T, arrivals []float32) (*ifpkg.If_Man_t, []*ifpkg.If_Cut_t) {
	t.Helper()
	p := ifpkg.If_ManStart(&ifpkg.If_Par_t{NLutSize: 6, NCutsMax: 8, FTruth: 1, FUsePerm: 1, FCutMin: 1})
	t.Cleanup(func() { ifpkg.If_ManStop(p) })
	cuts := make([]*ifpkg.If_Cut_t, len(arrivals))
	for i, arrival := range arrivals {
		obj := ifpkg.If_ManCreateCi(p)
		cuts[i] = ifpkg.If_ObjCutBest(obj)
		ifpkg.If_ManSetupCutTriv(p, cuts[i], ifpkg.If_ObjId(obj))
		cuts[i].Delay = arrival
	}
	return p, cuts
}

func recordedReviewMerge(t *testing.T, p *ifpkg.If_Man_t, a, b *ifpkg.If_Cut_t, ca, cb int32) *ifpkg.If_Cut_t {
	t.Helper()
	cut := &ifpkg.If_Cut_t{}
	ifpkg.If_CutSetup(p, cut)
	if ifpkg.If_CutMergeOrdered(p, a, b, cut) != 1 {
		t.Fatal("fixture cut merge failed")
	}
	ifpkg.If_CutComputeTruth(p, cut, a, b, ca, cb)
	return cut
}

func recordedReviewCompetingLibrary(t *testing.T) {
	t.Helper()
	previous := s_pMan3
	s_pMan3 = nil
	lib := gia.Gia_ManStart(256)
	inputs := make([]int32, 6)
	for i := range inputs {
		inputs[i] = gia.Gia_ManAppendCi(lib)
	}
	var outputs []int32
	for phase := int32(0); phase < 8; phase++ {
		a, b, c := inputs[0]^(phase&1), inputs[1]^((phase>>1)&1), inputs[2]^((phase>>2)&1)
		ab, ac, bc := gia.Gia_ManAppendAnd(lib, a, b), gia.Gia_ManAppendAnd(lib, a, c), gia.Gia_ManAppendAnd(lib, b, c)
		// Same function, different profiles/areas. Put the larger cone first
		// and duplicate the last profile to test both source tie breakers.
		roots := []int32{gia.Gia_ManAppendAnd(lib, ab, ac), gia.Gia_ManAppendAnd(lib, bc, a), gia.Gia_ManAppendAnd(lib, ac, b), gia.Gia_ManAppendAnd(lib, ab, c), gia.Gia_ManAppendAnd(lib, ab, c)}
		for complement := int32(0); complement < 2; complement++ {
			for _, root := range roots {
				outputs = append(outputs, root^complement)
			}
		}
	}
	for _, lit := range outputs {
		gia.Gia_ManAppendCo(lib, lit)
	}
	Abc_NtkRecStart3(lib, 6, 8, 0, 0)
	t.Cleanup(func() { Abc_NtkRecStop3(); s_pMan3 = previous })
	Lms_ManPrepare(s_pMan3)
	// Sixteen truth classes, five adjacent implementations per class.
	var wantIndex []int32
	var wantDelays []uint64
	var wantAreas []byte
	for i := 0; i < 16; i++ {
		wantIndex = append(wantIndex, int32(5*i))
		wantDelays = append(wantDelays, 0x222, 0x221, 0x212, 0x122, 0x122)
		wantAreas = append(wantAreas, 3, 2, 2, 2, 2)
	}
	wantIndex = append(wantIndex, 80)
	if !slices.Equal(vec.Vec_IntArray(s_pMan3.vTruthPo), wantIndex) || !slices.Equal(vec.Vec_WrdArray(s_pMan3.vDelays), wantDelays) || !slices.Equal(vec.Vec_StrArray(s_pMan3.vAreas), wantAreas) {
		t.Fatal("class grouping, physical pin depths or cone areas differ from fixture")
	}
}

func TestRecordedLibraryCompetingProfilesAndAreaTies(t *testing.T) {
	recordedReviewCompetingLibrary(t)
	profiles := [][3]int32{{2, 2, 2}, {1, 2, 2}, {2, 1, 2}, {2, 2, 1}, {2, 2, 1}}
	areas := []uint32{3, 2, 2, 2, 2}
	timings := [][]float32{{0, 0, 0}, {12, 0, 0}, {0, 12, 0}, {0, 0, 12}, {4, 4, 0}, {4, 0, 4}, {0, 4, 4}, {1.5, 2.75, 3.25}}
	for phase := int32(0); phase < 8; phase++ {
		for index, arrivals := range timings {
			t.Run(fmt.Sprintf("phase%d/times%d", phase, index), func(t *testing.T) {
				p, ci := recordedReviewIF(t, arrivals)
				ab := recordedReviewMerge(t, p, ci[0], ci[1], phase&1, (phase>>1)&1)
				cut := recordedReviewMerge(t, p, ab, ci[2], 0, (phase>>2)&1)
				var canonical [16]byte
				var polarity uint32
				best := int32(-1)
				gotDelay := If_CutFindBestStruct(p, cut, canonical[:], &polarity, &best)
				if best < 0 || best >= 80 {
					t.Fatal("no recorded implementation selected")
				}
				seen := uint8(0)
				for _, pin := range canonical[:3] {
					if pin >= 3 {
						t.Fatalf("invalid canonical permutation%v", canonical[:3])
					}
					seen |= 1 << pin
				}
				if seen != 7 {
					t.Fatal("canonical inputs are not a permutation")
				}
				// Rank the hand-written physical profiles with a stable sort,
				// independently of the backend's scan-and-replace algorithm.
				delays := make([]int32, 5)
				order := []int{0, 1, 2, 3, 4}
				for i, profile := range profiles {
					for pin, depth := range profile {
						delays[i] = max(delays[i], int32(arrivals[canonical[pin]]+float32(depth)))
					}
				}
				sort.SliceStable(order, func(i, j int) bool {
					a, b := order[i], order[j]
					if delays[a] != delays[b] {
						return delays[a] < delays[b]
					}
					return areas[a] < areas[b]
				})
				winner := order[0]
				if int(best)%5 != winner || gotDelay != delays[winner] || ifpkg.If_CutCost(cut) != areas[winner] {
					t.Fatalf("PO%d delay%d area%d: want candidate%d delay%d area%d", best, gotDelay, ifpkg.If_CutCost(cut), winner, delays[winner], areas[winner])
				}
				wantPins := make([]byte, 3)
				for pin, leaf := range canonical[:3] {
					wantPins[leaf] = byte(profiles[winner][pin])
				}
				if !slices.Equal(ifpkg.If_CutPerm(cut), wantPins) {
					t.Fatalf("pin remapping%v want%v", ifpkg.If_CutPerm(cut), wantPins)
				}
				if index != len(timings)-1 {
					// Integer arrivals are the source GIA driver's domain.
					// Read the selected profile through IF's actual timing code.
					cut.Delay = float32(gotDelay)
					if timed := ifpkg.If_CutDelay(p, nil, cut); timed != float32(gotDelay) {
						t.Fatalf("selected profile delay%f want%d", timed, gotDelay)
					}
					ifpkg.If_CutPropagateRequired(p, nil, cut, 20)
					for pin, depth := range wantPins {
						leaf := ifpkg.If_CutLeaf(p, cut, int32(pin))
						if leaf.Required != 20-float32(depth) {
							t.Fatalf("pin%d required%f want%f", pin, leaf.Required, 20-float32(depth))
						}
					}
				}
				for hash := int32(0); hash < 2; hash++ {
					dst := gia.Gia_ManStart(16)
					gia.Gia_ManHashAlloc(dst)
					leaves := vec.Vec_IntAlloc(3)
					for i := 0; i < 3; i++ {
						vec.Vec_IntPush(leaves, gia.Gia_ManAppendCi(dst))
					}
					literal := Abc_RecToGia3(dst, p, cut, leaves, hash)
					gia.Gia_ManAppendCo(dst, literal)
					gia.Gia_ManAppendCo(dst, literal^1)
					for assignment := 0; assignment < 8; assignment++ {
						inputs := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0}
						value := (inputs[0] != (phase&1 != 0)) && (inputs[1] != (phase&2 != 0)) && (inputs[2] != (phase&4 != 0))
						got, err := dst.Evaluate(inputs)
						if err != nil || !slices.Equal(got, []bool{value, !value}) {
							t.Fatalf("hash%d input%d: %v (%v)", hash, assignment, got, err)
						}
					}
					if gia.Gia_ManAndNum(dst) != 2 {
						t.Fatal("reconstruction lost the smaller selected cone")
					}
					vec.Vec_IntFree(leaves)
					gia.Gia_ManStop(dst)
					for id := int32(1); id < gia.Gia_ManObjNum(s_pMan3.pGia); id++ {
						if gia.Gia_ObjMark0(gia.Gia_ManObj(s_pMan3.pGia, id)) != 0 {
							t.Fatal("reconstruction left library marks")
						}
					}
				}
				for i, leaf := range ci {
					if leaf.Delay != arrivals[i] {
						t.Fatal("selection changed leaf timing")
					}
				}
			})
		}
	}
}

func TestRecordedLibraryMissingClassLeavesBestIndexUntouched(t *testing.T) {
	recordedMappingReviewLibrary(t) // This library has only two-input classes.
	p, ci := recordedReviewIF(t, []float32{0, 0, 0})
	cut := recordedReviewMerge(t, p, recordedReviewMerge(t, p, ci[0], ci[1], 0, 0), ci[2], 0, 0)
	Lms_ManPrepare(s_pMan3)
	var canonical [16]byte
	var phase uint32
	best := int32(-37)
	pins := ifpkg.If_CutPerm(cut)
	copy(pins, []byte{3, 2, 1})
	before := slices.Clone(pins)
	if delay := If_CutFindBestStruct(p, cut, canonical[:], &phase, &best); delay != ABC_INFINITY || ifpkg.If_CutCost(cut) != uint32(IF_COST_MAX) || best != -37 {
		t.Fatalf("missing class: delay%d cost%d best%d", delay, ifpkg.If_CutCost(cut), best)
	}
	if !slices.Equal(pins, before) {
		t.Fatal("missing class replaced pin delays")
	}
}

func TestRecordedLibrarySmallSupportCostsAndReconstruction(t *testing.T) {
	recordedMappingReviewLibrary(t)
	for _, minimize := range []int32{0, 1} {
		for _, kind := range []string{"zero", "one", "variable", "negative"} {
			t.Run(fmt.Sprintf("min%d/%s", minimize, kind), func(t *testing.T) {
				p, ci := recordedReviewIF(t, []float32{2, 7})
				p.Parameters().FCutMin = minimize
				zero := recordedReviewMerge(t, p, ci[0], ci[0], 0, 1)
				cut, wantDelay := zero, int32(0)
				switch kind {
				case "one":
					cut = recordedReviewMerge(t, p, zero, zero, 1, 1)
				case "variable", "negative":
					complement := int32(0)
					if kind == "negative" {
						complement = 1
					}
					cut = recordedReviewMerge(t, p, zero, ci[1], 1, complement)
					wantDelay = 7
				}
				p.Parameters().FCutMin = 1 // Source reconstruction requires this.
				delay := If_CutDelayRecCost3(p, cut, nil)
				canonical := [16]byte{13}
				polarity, best := uint32(9), int32(77)
				if again := If_CutFindBestStruct(p, cut, canonical[:], &polarity, &best); again != delay || canonical[0] != 13 || polarity != 9 || best != 77 {
					t.Fatal("trivial cut changed unused canonical outputs")
				}
				if delay != wantDelay || ifpkg.If_CutCost(cut) != 1 {
					t.Fatalf("delay%d area%d", delay, ifpkg.If_CutCost(cut))
				}
				pins := ifpkg.If_CutPerm(cut)
				for i, id := range ifpkg.If_CutLeaves(cut)[:ifpkg.If_CutLeaveNum(cut)] {
					want := byte(IF_BIG_CHAR)
					if (kind == "variable" || kind == "negative") && id == ifpkg.If_CutLeaves(ci[1])[0] {
						want = 0
					}
					if pins[i] != want {
						t.Fatalf("pin%d=%d want%d", i, pins[i], want)
					}
				}
				dst := gia.Gia_ManStart(8)
				a, b := gia.Gia_ManAppendCi(dst), gia.Gia_ManAppendCi(dst)
				leaves := vec.Vec_IntAlloc(2)
				for _, id := range ifpkg.If_CutLeaves(cut)[:ifpkg.If_CutLeaveNum(cut)] {
					lit := a
					if id == ifpkg.If_CutLeaves(ci[1])[0] {
						lit = b
					}
					vec.Vec_IntPush(leaves, lit)
				}
				literal := Abc_RecToGia3(dst, p, cut, leaves, 0)
				gia.Gia_ManAppendCo(dst, literal)
				for assignment := 0; assignment < 4; assignment++ {
					inputs := []bool{assignment&1 != 0, assignment&2 != 0}
					want := kind == "one" || kind == "variable" && inputs[1] || kind == "negative" && !inputs[1]
					got, err := dst.Evaluate(inputs)
					if err != nil || !slices.Equal(got, []bool{want}) {
						t.Fatalf("input%d: got%v want%v err%v", assignment, got, want, err)
					}
				}
				if gia.Gia_ManAndNum(dst) != 0 {
					t.Fatal("trivial recorded cut allocated a gate")
				}
				vec.Vec_IntFree(leaves)
				gia.Gia_ManStop(dst)
			})
		}
	}
}

func TestRecordedLibraryRejectsHoleySupportBeforeChangingCost(t *testing.T) {
	recordedMappingReviewLibrary(t)
	p, ci := recordedReviewIF(t, []float32{0, 0, 0})
	p.Parameters().FCutMin = 0
	zeroB := recordedReviewMerge(t, p, ci[1], ci[1], 0, 1)
	a := recordedReviewMerge(t, p, ci[0], zeroB, 0, 1)
	cut := recordedReviewMerge(t, p, a, ci[2], 0, 0)
	if support := Abc_TtSupport(If_CutTruthWordSlice(p, cut), 3); support != 5 {
		t.Fatalf("fixture support%d want5", support)
	}
	ifpkg.If_CutSetCost(cut, 17)
	pins := ifpkg.If_CutPerm(cut)
	copy(pins, []byte{3, 2, 1})
	before := slices.Clone(pins)
	defer func() {
		failure := recover()
		if failure == nil || !strings.Contains(fmt.Sprint(failure), "Abc_TtSuppOnlyOne(uSupport)") {
			t.Fatalf("source support assertion: %v", failure)
		}
		if ifpkg.If_CutCost(cut) != 17 || !slices.Equal(pins, before) {
			t.Fatalf("rejected cut changed cost%d or pins%v", ifpkg.If_CutCost(cut), pins)
		}
	}()
	var canonical [16]byte
	var phase uint32
	If_CutFindBestStruct(p, cut, canonical[:], &phase, nil)
}
