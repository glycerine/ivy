package gia

import (
	"fmt"
	"slices"
	"testing"

	"github.com/glycerine/abcgo/misc/tim"
)

func TestMappingOverlapRejectsAbsentMappingBeforeTraversal(t *testing.T) {
	for shape := 0; shape < 3; shape++ {
		for _, method := range []struct {
			name string
			run  func(*Gia_Man_t) int32
		}{{"epoch", Gia_ManComputeOverlap}, {"labels", Gia_ManComputeOverlap2}} {
			t.Run(fmt.Sprintf("shape%d/%s", shape, method.name), func(t *testing.T) {
				p := Gia_ManStart(8)
				defer Gia_ManStop(p)
				if shape != 0 {
					a := Gia_ManAppendCi(p)
					if shape == 2 {
						a = Gia_ManAppendAnd(p, a, Gia_ManAppendCi(p))
					}
					Gia_ManAppendCo(p, a)
				}
				var caught any
				func() { defer func() { caught = recover() }(); method.run(p) }()
				if caught != "gia: mapping is absent" || p.nTravIds != 0 {
					t.Fatalf("precondition panic%v epoch%d", caught, p.nTravIds)
				}
			})
		}
	}
}

func TestMappingMetricsEmptyAndTerminalBoundaries(t *testing.T) {
	for _, terminals := range []bool{false, true} {
		p := Gia_ManStart(8)
		if terminals {
			a := Gia_ManAppendCi(p)
			for _, lit := range []int32{0, 1, a, a ^ 1} {
				Gia_ManAppendCo(p, lit)
			}
			Gia_ManSetRegNum(p, 1)
		}
		p.VMapping = Vec_IntStart(int(p.nObjs))
		var levels []int32
		if Gia_ManLutNum(p) != 0 || Gia_ManLutFaninCount(p) != 0 || Gia_ManLutSizeMax(p) != -1 || Gia_ManLutLevel(p, &levels) != 0 || Gia_ManComputeOverlap(p) != 0 || Gia_ManComputeOverlap2(p) != 0 {
			t.Fatal("empty mapped metrics disagree with source")
		}
		if !slices.Equal(levels, make([]int32, p.nObjs)) {
			t.Fatal("terminal mapped levels are nonzero")
		}
		var luts, edges, depth int32 = -1, -1, -1
		Gia_ManLutParams(p, &luts, &edges, &depth)
		if luts != 0 || edges != 0 || depth != 0 {
			t.Fatal("empty mapped parameters are nonzero")
		}
		Gia_ManSetRefsMapped(p)
		Gia_ManSetLutRefs(p)
		want := make([]int32, p.nObjs)
		if terminals {
			want[0], want[1] = 2, 2
		}
		if !slices.Equal(p.Refs, want) || !slices.Equal(p.LUTReferenceCounts(), want) {
			t.Fatalf("terminal references%v/%v want%v", p.Refs, p.LUTReferenceCounts(), want)
		}
		Gia_ManStop(p)
	}
}

func TestMappingLutParamsTimingManagerBranches(t *testing.T) {
	for seed := uint32(1); seed <= 32; seed++ {
		p, boxedLevels, _ := mappingReviewBoxGraph(seed, false)
		originalTiming := p.PManTime
		var cached []int32
		flat := make([]int32, p.nObjs)
		var count, flatMax, boxedMax int32
		for id := int32(1); id < p.nObjs; id++ {
			a, b, _, kind := p.Gate(id)
			if kind == 2 {
				count++ // This fixture maps every raw AND to its two stored fanins.
				flat[id] = max(flat[a>>1], flat[b>>1]) + 1
			} else if kind == 6 {
				flat[id] = flat[a>>1]
				flatMax = max(flatMax, flat[id])
				boxedMax = max(boxedMax, boxedLevels[id])
			}
		}
		// Populate the box-level cache first, then ensure flat calls ignore it.
		for _, mode := range []int{2, 1, 0} {
			t.Run(fmt.Sprintf("seed%d/timing%d", seed, mode), func(t *testing.T) {
				switch mode {
				case 2:
					p.PManTime = originalTiming
				case 1:
					p.PManTime = tim.Tim_ManStart(int(Gia_ManCiNum(p)), int(Gia_ManCoNum(p)))
				case 0:
					p.PManTime = nil
				}
				var luts, edges, depth int32
				Gia_ManLutParams(p, &luts, &edges, &depth)
				wantDepth := flatMax
				if mode == 2 {
					wantDepth = boxedMax
				}
				if luts != count || edges != 2*count || depth != wantDepth {
					t.Fatalf("parameters %d/%d/%d want %d/%d/%d", luts, edges, depth, count, 2*count, wantDepth)
				}
				if mode == 2 {
					cached = mappingReviewStoredLevels(p)
				} else if !slices.Equal(mappingReviewStoredLevels(p), cached) || p.nLevels != boxedMax {
					t.Fatal("flat parameters overwrote the cached box levels")
				}
			})
		}
		Gia_ManStop(p)
	}
}

func TestMappingMetricsDanglingCutsAndPhysicalNodes(t *testing.T) {
	p := Gia_ManStart(16)
	defer Gia_ManStop(p)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	y := Gia_ManAppendAnd(p, x, c)
	Gia_ManAppendAnd(p, b, c) // Physical but unindexed: source overlap subtracts it.
	Gia_ManAppendCo(p, a)
	p.VMapping = Vec_IntStart(int(p.nObjs))
	for _, row := range [][3]int32{{x, a, b}, {y, x, c}} {
		id := row[0] >> 1
		Vec_IntWriteEntry(p.VMapping, int(id), int32(Vec_IntSize(p.VMapping)))
		for _, value := range []int32{2, row[1] >> 1, row[2] >> 1, id} {
			Vec_IntPush(p.VMapping, value)
		}
	}
	var levels []int32
	if Gia_ManLutLevel(p, &levels) != 0 || !slices.Equal(levels, []int32{0, 0, 0, 0, 1, 2, 0, 0}) {
		t.Fatalf("output depth includes dangling LUTs: %v", levels)
	}
	var luts, edges, depth int32
	Gia_ManLutParams(p, &luts, &edges, &depth)
	if luts != 2 || edges != 4 || depth != 0 || Gia_ManLutSizeMax(p) != 2 {
		t.Fatal("dangling cuts were excluded from size/count totals")
	}
	if Gia_ManComputeOverlap(p) != -1 || Gia_ManComputeOverlap2(p) != -1 {
		t.Fatal("source physical-node overlap subtraction changed")
	}
	Gia_ManSetRefsMapped(p)
	Gia_ManSetLutRefs(p)
	want := []int32{0, 2, 1, 1, 1, 0, 0, 0}
	if !slices.Equal(p.Refs, want) || !slices.Equal(p.LUTReferenceCounts(), want) {
		t.Fatalf("dangling-cut references %v/%v want%v", p.Refs, p.LUTReferenceCounts(), want)
	}
}
