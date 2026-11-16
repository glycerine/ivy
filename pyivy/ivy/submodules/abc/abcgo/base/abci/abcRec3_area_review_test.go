package abci

import (
	"fmt"
	"slices"
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	ifpkg "github.com/glycerine/abcgo/map/if"
	"github.com/glycerine/abcgo/misc/vec"
)

// Each root is a separate tree of exactly area ANDs. Repeated CI occurrences
// keep its truth a two-input minterm; its depth is eight for every tested area.
// Keeping depths below the nibble cap isolates area-byte interpretation.
func recordedAreaReviewLibrary(t *testing.T, areas []int) {
	t.Helper()
	lib := gia.Gia_ManStart(9000)
	inputs := make([]int32, 6)
	for i := range inputs {
		inputs[i] = gia.Gia_ManAppendCi(lib)
	}
	var outputs []int32
	for phase := int32(0); phase < 4; phase++ {
		a, b := inputs[0]^(phase&1), inputs[1]^((phase>>1)&1)
		var roots []int32
		for _, area := range areas {
			leaf := 0
			var build func(int, bool) int32
			build = func(count int, root bool) int32 {
				if count == 0 {
					leaf++
					if leaf&1 != 0 {
						return a
					}
					return b
				}
				left := (count - 1) / 2
				if root && count < 128 {
					left = (count - 1) / 4
				}
				return gia.Gia_ManAppendAnd(lib, build(left, false), build(count-1-left, false))
			}
			roots = append(roots, build(area, true))
		}
		for complement := int32(0); complement < 2; complement++ {
			for _, root := range roots {
				outputs = append(outputs, root^complement)
			}
		}
	}
	for _, literal := range outputs {
		gia.Gia_ManAppendCo(lib, literal)
	}
	previous := s_pMan3
	s_pMan3 = nil
	Abc_NtkRecStart3(lib, 6, 8, 0, 0)
	t.Cleanup(func() { Abc_NtkRecStop3(); s_pMan3 = previous })
	Lms_ManPrepare(s_pMan3)
	for i := 0; i < 8*len(areas); i++ {
		if got := vec.Vec_StrEntry(s_pMan3.vAreas, i); int(got) != areas[i%len(areas)] {
			t.Fatalf("area[%d]=%d want%d", i, got, areas[i%len(areas)])
		}
		if got := vec.Vec_WrdEntry(s_pMan3.vDelays, i); got != 0x88 {
			t.Fatalf("profile[%d]=%x want88", i, got)
		}
	}
}

func TestRecordedAreaBytesRetainGateCountSemantics(t *testing.T) {
	for _, areas := range [][]int{{250, 129, 128, 251, 127, 126}, {128}, {251}} {
		t.Run(fmt.Sprint(areas), func(t *testing.T) {
			recordedAreaReviewLibrary(t, areas)
			wantArea := slices.Min(areas)
			for phase := int32(0); phase < 4; phase++ {
				for _, arrivals := range [][]float32{{0, 0}, {3, 7}, {9, 1}} {
					p, ci := recordedReviewIF(t, arrivals)
					cut := recordedReviewMerge(t, p, ci[0], ci[1], phase&1, (phase>>1)&1)
					delay := If_CutDelayRecCost3(p, cut, nil)
					if cost := ifpkg.If_CutCost(cut); cost != uint32(wantArea) || delay != int32(max(arrivals[0], arrivals[1]))+8 {
						t.Fatalf("phase%d arrival%v: cost%d want%d delay%d", phase, arrivals, cost, wantArea, delay)
					}
					if !slices.Equal(ifpkg.If_CutPerm(cut), []byte{8, 8}) {
						t.Fatal("area conversion affected pin delays")
					}
					for hash := int32(0); hash < 2; hash++ {
						dst := gia.Gia_ManStart(int32(2*wantArea + 8))
						gia.Gia_ManHashAlloc(dst)
						a, b := gia.Gia_ManAppendCi(dst), gia.Gia_ManAppendCi(dst)
						leaves := vec.Vec_IntAllocArrayCopy([]int32{a, b}, 2)
						literal := Abc_RecToGia3(dst, p, cut, leaves, hash)
						gia.Gia_ManAppendCo(dst, literal)
						for assignment := 0; assignment < 4; assignment++ {
							inputs := []bool{assignment&1 != 0, assignment&2 != 0}
							want := (inputs[0] != (phase&1 != 0)) && (inputs[1] != (phase&2 != 0))
							got, err := dst.Evaluate(inputs)
							if err != nil || !slices.Equal(got, []bool{want}) {
								t.Fatalf("phase%d hash%d input%d: %v (%v)", phase, hash, assignment, got, err)
							}
						}
						nodes := gia.Gia_ManAndNum(dst)
						if hash == 0 && nodes != int32(wantArea) || hash != 0 && (nodes <= 0 || nodes >= int32(wantArea)) {
							t.Fatalf("hash%d nodes%d raw area%d", hash, nodes, wantArea)
						}
						vec.Vec_IntFree(leaves)
						gia.Gia_ManStop(dst)
					}
				}
			}
		})
	}
}
