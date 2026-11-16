package if_

import (
	"math"
	"slices"
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/tim"
	"github.com/glycerine/abcgo/misc/vec"
)

func iffReviewMapping(t *testing.T, graph *Gia_Man_t, cuts map[int32][]int32) {
	t.Helper()
	mapping := vec.Vec_IntStart(int(Gia_ManObjNum(graph)))
	for id := int32(0); id < Gia_ManObjNum(graph); id++ {
		leaves, ok := cuts[id]
		if !ok {
			continue
		}
		vec.Vec_IntWriteEntry(mapping, int(id), int32(Vec_IntSize(mapping)))
		Vec_IntPush(mapping, int32(len(leaves)))
		for _, leaf := range leaves {
			Vec_IntPush(mapping, leaf)
		}
		Vec_IntPush(mapping, id)
	}
	if err := graph.SetMappedData(mapping, nil, nil, ""); err != nil {
		t.Fatal(err)
	}
}

func iffReviewTree(t *testing.T) (graph *Gia_Man_t, x, y, root int32) {
	t.Helper()
	graph = Gia_ManStart(16)
	a, b, c, d := Gia_ManAppendCi(graph), Gia_ManAppendCi(graph), Gia_ManAppendCi(graph), Gia_ManAppendCi(graph)
	x, y = gia.Gia_ManAppendAnd(graph, a, b), gia.Gia_ManAppendAnd(graph, c, d)
	root = gia.Gia_ManAppendAnd(graph, x, y)
	Gia_ManAppendCo(graph, root)
	x, y, root = x>>1, y>>1, root>>1
	iffReviewMapping(t, graph, map[int32][]int32{x: {a >> 1, b >> 1}, y: {c >> 1, d >> 1}, root: {x, y}})
	return
}

func iffReviewWide(t *testing.T, inputs int) (graph *Gia_Man_t, root int32) {
	t.Helper()
	graph = Gia_ManStart(int32(2*inputs + 4))
	leaves := make([]int32, inputs)
	for i := range leaves {
		leaves[i] = Gia_ManAppendCi(graph) >> 1
	}
	root = leaves[0] * 2
	for _, leaf := range leaves[1:] {
		root = gia.Gia_ManAppendAnd(graph, root, leaf*2)
	}
	root >>= 1
	iffReviewMapping(t, graph, map[int32][]int32{root: leaves})
	return
}

func TestIffSourceInitialFloatBits(t *testing.T) {
	graph, _, _, _ := iffReviewTree(t)
	p := Gia_ManIffStart(graph)
	defer Gia_ManIffStop(p)
	for id, arrival := range p.times {
		// Vec_FltStartFull fills bytes with FF; it does not assign float(-1).
		if bits := math.Float32bits(arrival); bits != 0xffffffff {
			t.Fatalf("arrival%d bits%08x want ffffffff", id, bits)
		}
		if p.matches[2][id] != -1 || p.matches[3][id] != -1 {
			t.Fatal("unselected match is not -1")
		}
	}
}

func TestIffTimeIgnoresUnsetArrivalLikeSource(t *testing.T) {
	graph, x, y, root := iffReviewTree(t)
	p := Gia_ManIffStart(graph)
	defer Gia_ManIffStop(p)
	p.times[x], p.times[y] = math.Float32frombits(0xffffffff), 2
	if got := Gia_IffObjTimeOne(p, root, -1, -1); got != 2 {
		t.Fatalf("explicit source comparison should ignore NaN: got%v want2", got)
	}
}

func TestIffTripleRejectsUnsetBoundary(t *testing.T) {
	graph, x, y, root := iffReviewTree(t)
	p := Gia_ManIffStart(graph)
	defer Gia_ManIffStop(p)
	defer func() {
		if recover() == nil {
			t.Fatal("source requires a nonnegative triple boundary arrival")
		}
	}()
	Gia_IffObjTimeOne(p, root, x, y)
}

func TestIffEmptyOutputMaximum(t *testing.T) {
	graph := Gia_ManStart(1)
	timing := tim.Tim_ManStart(0, 0)
	defer tim.Tim_ManStop(timing)
	got := dsdReportCapture(t, func() {
		p := Gia_ManIffPerform(graph, &If_LibLut_t{}, timing, 4, 2)
		Gia_ManIffStop(p)
	})
	want := "Max delay = -1000000000.00.  Count1 = 0.  Count2 = 0.  Count3 = 0.\n"
	if got != want {
		t.Fatalf("empty maximum: got%q want%q", got, want)
	}
}

func TestIffCountUsesTraversalEpoch(t *testing.T) {
	graph, x, y, root := iffReviewTree(t)
	gia.Gia_ManIncrementTravId(graph)
	gia.Gia_ObjSetTravIdCurrentId(graph, root)
	if count := Gia_IffObjCount(graph, root, x, y); count != 4 {
		t.Fatalf("expanded boundary count%d want4", count)
	}
	for id := int32(0); id < Gia_ManObjNum(graph); id++ {
		want := id >= 1 && id <= 4
		if marked := gia.Gia_ObjIsTravIdCurrentId(graph, id) != 0; marked != want {
			t.Fatalf("boundary traversal mark%d got%t want%t", id, marked, want)
		}
	}
}

func TestIffSelectionUsesTraversalEpoch(t *testing.T) {
	graph, x, y, root := iffReviewTree(t)
	p := Gia_ManIffStart(graph)
	defer Gia_ManIffStop(p)
	p.matches[2][root] = x
	gia.Gia_ManIncrementTravId(graph)
	co := gia.Gia_ObjId(graph, Gia_ManCo(graph, 0))
	gia.Gia_ObjSetTravIdCurrentId(graph, co)
	for range 2 {
		packing := Gia_ManIffSelect(p)
		if !slices.Equal(Vec_IntArray(packing), []int32{2, 1, y, 2, x, root}) {
			t.Fatalf("packed postorder: %v", Vec_IntArray(packing))
		}
		for id := int32(0); id < Gia_ManObjNum(graph); id++ {
			want := id != x && id != co
			if marked := gia.Gia_ObjIsTravIdCurrentId(graph, id) != 0; marked != want {
				t.Fatalf("selected traversal mark%d got%t want%t", id, marked, want)
			}
		}
	}
}

func TestIffCountReusesTraversalStorage(t *testing.T) {
	graph, root := iffReviewWide(t, 16)
	Gia_IffObjCount(graph, root, -1, -1)
	allocs := testing.AllocsPerRun(100, func() {
		if count := Gia_IffObjCount(graph, root, -1, -1); count != 16 {
			t.Fatalf("boundary count%d want16", count)
		}
	})
	if allocs != 0 {
		t.Fatalf("warmed boundary count allocates%g times, want0", allocs)
	}
}

func TestIffTimeWithOnlyThirdSkipUsesSourceEarlyReturn(t *testing.T) {
	graph, x, y, root := iffReviewTree(t)
	p := Gia_ManIffStart(graph)
	defer Gia_ManIffStop(p)
	for i := range p.times {
		p.times[i] = 9
	}
	p.times[x] = 2
	if got := Gia_IffObjTimeOne(p, root, -1, y); got != 2 {
		t.Fatalf("source returns before expanding a third skip without a second: got%v want2", got)
	}
}

func TestIffEmptyBoundaryMaximum(t *testing.T) {
	graph := Gia_ManStart(1)
	iffReviewMapping(t, graph, nil)
	p := Gia_ManIffStart(graph)
	defer Gia_ManIffStop(p)
	if got := Gia_IffObjTimeOne(p, 0, -1, -1); got != -1000000000 {
		t.Fatalf("source empty boundary maximum: got%v want-1000000000", got)
	}
}

func TestIffTripleValidatesTimingBeforeCounting(t *testing.T) {
	graph, _, _, root := iffReviewTree(t)
	p := Gia_ManIffStart(graph)
	defer Gia_ManIffStop(p)
	p.library = &If_LibLut_t{LutMax: 10}
	gia.Gia_ManIncrementTravId(graph)
	gia.Gia_ObjSetTravIdCurrentId(graph, root)
	defer func() {
		if recover() == nil {
			t.Error("source rejects an unset triple boundary")
		}
		if gia.Gia_ObjIsTravIdCurrentId(graph, root) == 0 {
			t.Error("support counting changed traversal state before timing validation")
		}
	}()
	var a, b int32
	Gia_IffObjTimeThree(p, root, &a, &b, 10)
}

func TestIffUnsupportedLibraryDiagnostic(t *testing.T) {
	for _, tc := range []struct {
		inputs int
		limit  int32
		want   string
	}{
		{2, 6, "LUT library for packing 4-LUTs should have 7 or 10 inputs.\n"},
		{5, 10, "LUT library for packing 6-LUTs should have 11 or 16 inputs.\n"},
		{7, 16, "The LUT size is more than 6.\n"},
	} {
		graph, _ := iffReviewWide(t, tc.inputs)
		got := dsdReportCapture(t, func() { Gia_ManIffTest(graph, &If_LibLut_t{LutMax: tc.limit}, 0) })
		if got != tc.want {
			t.Errorf("unsupported library for%d inputs: got%q want%q", tc.inputs, got, tc.want)
		}
		_, _, _, timing := graph.MappingInputs()
		if graph.VPacking != nil || timing != nil {
			t.Fatal("rejected wrapper changed packing or timing ownership")
		}
	}
}

func TestIffReportUsesSourceMaximumComparison(t *testing.T) {
	graph := Gia_ManStart(8)
	a, b := Gia_ManAppendCi(graph), Gia_ManAppendCi(graph)
	Gia_ManAppendCo(graph, a)
	Gia_ManAppendCo(graph, b)
	timing := tim.Tim_ManStart(2, 2)
	defer tim.Tim_ManStop(timing)
	tim.Tim_ManInitPiArrival(timing, 0, math.Float32frombits(0xffffffff))
	tim.Tim_ManInitPiArrival(timing, 1, 2)
	got := dsdReportCapture(t, func() {
		p := Gia_ManIffPerform(graph, &If_LibLut_t{}, timing, 4, 2)
		Gia_ManIffStop(p)
	})
	want := "Max delay = 2.00.  Count1 = 0.  Count2 = 0.  Count3 = 0.\n"
	if got != want {
		t.Fatalf("source comparison should recover at the next finite output: got%q want%q", got, want)
	}
}
