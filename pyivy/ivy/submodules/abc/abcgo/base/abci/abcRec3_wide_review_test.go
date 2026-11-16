package abci

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	ifpkg "github.com/glycerine/abcgo/map/if"
	"github.com/glycerine/abcgo/misc/vec"
)

// Build the canonical minterm as an actual GIA, using only the canonicalizer
// to obtain the library key. Truth and physical depths are checked separately.
func recordedWideReviewLibrary(t *testing.T, width, support int32, balanced bool) []int32 {
	t.Helper()
	mask := (1 << support) - 1
	truth := make([]uint64, max(1, 1<<max(support-6, 0)))
	for assignment := 0; assignment < max(64, 1<<support); assignment++ {
		if assignment&mask == mask {
			truth[assignment>>6] |= uint64(1) << (assignment & 63)
		}
	}
	Abc_TtCanonicize(truth, support, make([]byte, support))
	ones, minority := 0, -1
	for assignment := 0; assignment < 1<<support; assignment++ {
		if truth[assignment>>6]>>(assignment&63)&1 != 0 {
			ones++
		}
	}
	complemented := ones != 1
	if ones != 1 && ones != (1<<support)-1 {
		t.Fatal("canonical AND is neither a minterm nor its complement")
	}
	for assignment := 0; assignment < 1<<support; assignment++ {
		if (truth[assignment>>6]>>(assignment&63)&1 != 0) != complemented {
			minority = assignment
			break
		}
	}
	lib := gia.Gia_ManStart(2*width + 8)
	inputs := make([]int32, width)
	for i := range inputs {
		inputs[i] = gia.Gia_ManAppendCi(lib)
	}
	depths := make([]int32, support)
	var build func(int32, int32, int32) int32
	build = func(lo, hi, depth int32) int32 {
		if hi-lo == 1 {
			depths[lo] = depth
			literal := inputs[lo]
			if minority&(1<<lo) == 0 {
				literal ^= 1
			}
			return literal
		}
		mid := hi - 1
		if balanced {
			mid = (lo + hi) / 2
		}
		return gia.Gia_ManAppendAnd(lib, build(lo, mid, depth+1), build(mid, hi, depth+1))
	}
	root := build(0, support, 0)
	if complemented {
		root ^= 1
	}
	gia.Gia_ManAppendCo(lib, root)
	previous := s_pMan3
	s_pMan3 = nil
	Abc_NtkRecStart3(lib, width, 8, 0, 0)
	t.Cleanup(func() { Abc_NtkRecStop3(); s_pMan3 = previous })
	Lms_ManPrepare(s_pMan3)
	wantTruth := make([]uint64, 1<<(width-6))
	for assignment := 0; assignment < 1<<width; assignment++ {
		if (assignment&mask == minority) != complemented {
			wantTruth[assignment>>6] |= uint64(1) << (assignment & 63)
		}
	}
	if !slices.Equal(vec.Vec_MemReadEntry(s_pMan3.vTtMem, 0), wantTruth) || !slices.Equal(vec.Vec_IntArray(s_pMan3.vTruthPo), []int32{0, 1}) {
		t.Fatal("wide library truth or class bounds changed")
	}
	var profile word
	for i, depth := range depths {
		// C stores depths with a four-bit field and saturates before removing
		// the CI offset at a CO. Thus the sixteen-input chain clips15 to14.
		profile |= word(min(depth, 14)) << (4 * i)
	}
	if vec.Vec_WrdEntry(s_pMan3.vDelays, 0) != profile || vec.Vec_StrEntry(s_pMan3.vAreas, 0) != byte(support-1) {
		t.Fatalf("wide physical metadata: delay%x want%x area%d want%d", vec.Vec_WrdEntry(s_pMan3.vDelays, 0), profile, vec.Vec_StrEntry(s_pMan3.vAreas, 0), support-1)
	}
	return depths
}

func TestRecordedWideLibraryWidths(t *testing.T) {
	for width := int32(6); width <= ifpkg.IF_MAX_FUNC_LUTSIZE; width++ {
		supports := []int32{2, 3, 5, 6}
		if width != 6 {
			supports = append(supports, width)
		}
		for _, support := range supports {
			for _, balanced := range []bool{false, true} {
				t.Run(fmt.Sprintf("n%d/support%d/balanced%t", width, support, balanced), func(t *testing.T) {
					depths := recordedWideReviewLibrary(t, width, support, balanced)
					mask := uint32(1<<support) - 1
					phases := []uint32{0, mask, mask & 0xaaaa, 1 << (support - 1)}
					used := make(map[uint32]bool)
					for _, phase := range phases {
						if used[phase] {
							continue
						}
						used[phase] = true
						t.Run(fmt.Sprintf("phase%x", phase), func(t *testing.T) {
							pars := &ifpkg.If_Par_t{NLutSize: width, NCutsMax: 8, FTruth: 1, FUsePerm: 1, FCutMin: 1}
							p := ifpkg.If_ManStart(pars)
							defer ifpkg.If_ManStop(p)
							ci := make([]*ifpkg.If_Cut_t, support)
							arrivals := make([]int32, support)
							for i := range ci {
								obj := ifpkg.If_ManCreateCi(p)
								ci[i] = ifpkg.If_ObjCutBest(obj)
								ifpkg.If_ManSetupCutTriv(p, ci[i], ifpkg.If_ObjId(obj))
								arrivals[i] = int32((3*i + 1) % 7)
								ci[i].Delay = float32(arrivals[i])
							}
							cut := ci[0]
							for i := int32(1); i < support; i++ {
								complement := int32(0)
								if i == 1 {
									complement = int32(phase & 1)
								}
								cut = recordedReviewMerge(t, p, cut, ci[i], complement, int32(phase>>i&1))
							}
							delay := If_CutDelayRecCost3(p, cut, nil)
							var canonical [LMS_VAR_MAX]byte
							var polarity uint32
							best := int32(-1)
							if again := If_CutFindBestStruct(p, cut, canonical[:], &polarity, &best); again != delay || best != 0 || ifpkg.If_CutCost(cut) != uint32(support-1) {
								t.Fatal("wide lookup selected different cost/record")
							}
							wantDelay := int32(0)
							wantPins := make([]byte, support)
							seen := uint32(0)
							for i, depth := range depths {
								pin := canonical[i]
								if int32(pin) >= support || seen&(1<<pin) != 0 {
									t.Fatal("wide canonical permutation lost an input")
								}
								seen |= 1 << pin
								wantPins[pin] = byte(min(depth, 14))
								wantDelay = max(wantDelay, arrivals[pin]+min(depth, 14))
							}
							if delay != wantDelay || !slices.Equal(ifpkg.If_CutPerm(cut), wantPins) {
								t.Fatalf("wide cost%d want%d pins%v want%v", delay, wantDelay, ifpkg.If_CutPerm(cut), wantPins)
							}
							cut.Delay = float32(delay)
							if timed := ifpkg.If_CutDelay(p, nil, cut); timed != float32(delay) {
								t.Fatal("wide selected pin timing changed")
							}
							for hash := int32(0); hash < 2; hash++ {
								dst := gia.Gia_ManStart(2*support + 8)
								gia.Gia_ManHashAlloc(dst)
								leaves := vec.Vec_IntAlloc(int(support))
								for i := int32(0); i < support; i++ {
									vec.Vec_IntPush(leaves, gia.Gia_ManAppendCi(dst))
								}
								literal := Abc_RecToGia3(dst, p, cut, leaves, hash)
								gia.Gia_ManAppendCo(dst, literal)
								gia.Gia_ManAppendCo(dst, literal^1)
								inputs := make([]bool, support)
								for assignment := 0; assignment < 1<<support; assignment++ {
									for i := range inputs {
										inputs[i] = assignment&(1<<i) != 0
									}
									value := uint32(assignment) == mask^phase
									got, err := dst.Evaluate(inputs)
									if err != nil || !slices.Equal(got, []bool{value, !value}) {
										t.Fatalf("hash%d input%x: %v (%v)", hash, assignment, got, err)
									}
								}
								if gia.Gia_ManAndNum(dst) != support-1 {
									t.Fatal("wide reconstruction changed physical area")
								}
								vec.Vec_IntFree(leaves)
								gia.Gia_ManStop(dst)
							}
						})
					}
				})
			}
		}
	}
}

// LMS stores sixteen variables, but source IF truth arrays have indices0..15.
// Exercise the larger storage/profile boundary without an invalid IF setup.
func TestRecordedWideLibraryStorageSixteen(t *testing.T) {
	for _, support := range []int32{2, 3, 5, 6, 15, 16} {
		for _, balanced := range []bool{false, true} {
			t.Run(fmt.Sprintf("support%d/balanced%t", support, balanced), func(t *testing.T) { recordedWideReviewLibrary(t, 16, support, balanced) })
		}
	}
}

func TestRecordedWideLibraryRetainsIfTruthCacheBound(t *testing.T) {
	defer func() {
		failure := recover()
		if failure == nil || !strings.Contains(fmt.Sprint(failure), "truth LUT size is within the supported truth-cache bound") {
			t.Fatalf("unsupported IF truth width: %v", failure)
		}
	}()
	ifpkg.If_ManStart(&ifpkg.If_Par_t{NLutSize: 16, NCutsMax: 8, FTruth: 1})
}

func TestRecordedLibraryCostPreconditionsBeforePreparation(t *testing.T) {
	for _, constructing := range []bool{false, true} {
		t.Run(fmt.Sprintf("constructing%t", constructing), func(t *testing.T) {
			recordedMappingReviewLibrary(t)
			width := int32(7) // Deliberately differs from the six-variable library.
			want := "p.nVars == int32(If_CutLimit(pCut))"
			if constructing {
				width, s_pMan3.fLibConstr, want = 6, 1, "!p.fLibConstr"
			}
			p := ifpkg.If_ManStart(&ifpkg.If_Par_t{NLutSize: width, NCutsMax: 8, FTruth: 1})
			defer ifpkg.If_ManStop(p)
			a, b := ifpkg.If_ManCreateCi(p), ifpkg.If_ManCreateCi(p)
			for _, obj := range []*ifpkg.If_Obj_t{a, b} {
				ifpkg.If_ManSetupCutTriv(p, ifpkg.If_ObjCutBest(obj), ifpkg.If_ObjId(obj))
			}
			cut := recordedReviewMerge(t, p, ifpkg.If_ObjCutBest(a), ifpkg.If_ObjCutBest(b), 0, 0)
			ifpkg.If_CutSetCost(cut, 17)
			defer func() {
				failure := recover()
				if failure == nil || !strings.Contains(fmt.Sprint(failure), want) {
					t.Fatalf("source precondition: %v", failure)
				}
				if s_pMan3.vTruthPo != nil || ifpkg.If_CutCost(cut) != 17 {
					t.Fatal("invalid library request prepared state or changed cost")
				}
			}()
			If_CutDelayRecCost3(p, cut, nil)
		})
	}
}
