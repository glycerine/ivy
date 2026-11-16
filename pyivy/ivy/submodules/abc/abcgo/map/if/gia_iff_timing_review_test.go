package if_

import (
	"fmt"
	"math"
	"slices"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/tim"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestIffStrictTiesRetainEarlierSmallerClusters(t *testing.T) {
	for _, tc := range []struct {
		name                   string
		degree                 int32
		pairDelay, tripleDelay float32
		wantPair               bool
	}{
		{"single", 3, 1, 2, false},
		{"firstPair", 2, .5, 2, true},
		{"pairBeforeEqualTriple", 3, .5, 1.5, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			graph, x, y, root := iffReviewTree(t)
			lib := &If_LibLut_t{LutMax: 10}
			lib.pLutDelays[2][0], lib.pLutDelays[3][0], lib.pLutDelays[4][0] = 1, tc.pairDelay, tc.tripleDelay
			timing := tim.Tim_ManStart(4, 1)
			defer tim.Tim_ManStop(timing)
			dsdReportCapture(t, func() {
				p := Gia_ManIffPerform(graph, lib, timing, 4, tc.degree)
				defer Gia_ManIffStop(p)
				wantMatch, wantDelay := int32(-1), float32(2)
				wantPacking := []int32{3, 1, x, 1, y, 1, root}
				if tc.wantPair {
					wantMatch, wantDelay = x, 1.5
					wantPacking = []int32{2, 1, y, 2, x, root}
				}
				if p.matches[2][root] != wantMatch || p.matches[3][root] != -1 || p.times[root] != wantDelay {
					t.Fatalf("root matches%d/%d time%v", p.matches[2][root], p.matches[3][root], p.times[root])
				}
				if got := Vec_IntArray(Gia_ManIffSelect(p)); !slices.Equal(got, wantPacking) {
					t.Fatalf("packing%v want%v", got, wantPacking)
				}
			})
		})
	}
}

func TestIffTripleTieFollowsFaninOrder(t *testing.T) {
	graph := Gia_ManStart(24)
	inputs := make([]int32, 6)
	for i := range inputs {
		inputs[i] = Gia_ManAppendCi(graph)
	}
	x := gia.Gia_ManAppendAnd(graph, inputs[0], inputs[1])
	y := gia.Gia_ManAppendAnd(graph, inputs[2], inputs[3])
	z := gia.Gia_ManAppendAnd(graph, inputs[4], inputs[5])
	root := gia.Gia_ManAppendAnd(graph, gia.Gia_ManAppendAnd(graph, x, y), z)
	Gia_ManAppendCo(graph, root)
	x, y, z, root = x>>1, y>>1, z>>1, root>>1
	iffReviewMapping(t, graph, map[int32][]int32{
		x: {1, 2}, y: {3, 4}, z: {5, 6}, root: {z, x, y},
	})
	lib := &If_LibLut_t{LutMax: 10}
	lib.pLutDelays[2][0], lib.pLutDelays[3][0] = 1, 1
	lib.pLutDelays[4][0], lib.pLutDelays[5][0] = .5, .25
	timing := tim.Tim_ManStart(6, 1)
	defer tim.Tim_ManStop(timing)
	dsdReportCapture(t, func() {
		p := Gia_ManIffPerform(graph, lib, timing, 4, 3)
		defer Gia_ManIffStop(p)
		if p.matches[2][root] != z || p.matches[3][root] != x || p.times[root] != 1.25 {
			t.Fatalf("first ordered triple lost: %d/%d time%v", p.matches[2][root], p.matches[3][root], p.times[root])
		}
		want := []int32{2, 1, y, 3, z, x, root}
		if got := Vec_IntArray(Gia_ManIffSelect(p)); !slices.Equal(got, want) {
			t.Fatalf("triple postorder%v want%v", got, want)
		}
	})
}

func TestIffWhiteBoxTimingAndWrapperOwnership(t *testing.T) {
	for _, limit := range []int32{7, 10} {
		for _, borrowed := range []bool{false, true} {
			t.Run(fmt.Sprintf("library%d/borrowed%t", limit, borrowed), func(t *testing.T) {
				graph := Gia_ManStart(16)
				a, b := Gia_ManAppendCi(graph), Gia_ManAppendCi(graph)
				x := gia.Gia_ManAppendAnd(graph, a, b)
				Gia_ManAppendCo(graph, x) // Box input must precede its output CI.
				boxOutput := Gia_ManAppendCi(graph)
				root := gia.Gia_ManAppendAnd(graph, boxOutput, a^1)
				Gia_ManAppendCo(graph, root^1)
				x, root = x>>1, root>>1
				iffReviewMapping(t, graph, map[int32][]int32{x: {a >> 1, b >> 1}, root: {boxOutput >> 1, a >> 1}})
				lib := &If_LibLut_t{LutMax: limit}
				for i := range lib.pLutDelays {
					lib.pLutDelays[i][0] = 1
				}
				var timing *Tim_Man_t
				if borrowed {
					timing = tim.Tim_ManStart(3, 2)
					defer tim.Tim_ManStop(timing)
					tim.Tim_ManSetDelayTables(timing, [][]float32{{0, 1, 1, 2.5}})
					tim.Tim_ManCreateBox(timing, 0, 1, 2, 1, 0, 0)
					tim.Tim_ManInitPiArrival(timing, 0, 1)
					tim.Tim_ManInitPiArrival(timing, 1, 3)
					graph.PManTime = timing
				}
				mapping := graph.VMapping
				for range 2 {
					previous := vec.Vec_IntAllocArrayCopy([]int32{0}, 1)
					if err := graph.ReplacePacking(previous); err != nil {
						t.Fatal(err)
					}
					got := dsdReportCapture(t, func() { Gia_ManIffTest(graph, lib, 0) })
					delay := "1.00"
					if borrowed {
						delay = "7.50"
						if tim.Tim_ManCiNum(timing) != 3 || tim.Tim_ManCoNum(timing) != 2 || tim.Tim_ManGetCoArrival(timing, 0) != 4 || tim.Tim_ManGetCoArrival(timing, 1) != 7.5 || tim.Tim_ManGetCiArrival(timing, 2) != 6.5 {
							t.Fatal("borrowed timing was destroyed or box propagation changed")
						}
					}
					wantReport := "Max delay = " + delay + ".  Count1 = 2.  Count2 = 0.  Count3 = 0.\n"
					if got != wantReport {
						t.Fatalf("white-box summary%q want%q", got, wantReport)
					}
					_, _, _, attached := graph.MappingInputs()
					wantPacking := []int32{2, 1, x, 1, root}
					if attached != timing || graph.VMapping != mapping || graph.VPacking == previous || !slices.Equal(Vec_IntArray(graph.VPacking), wantPacking) {
						t.Fatal("wrapper changed timing/mapping ownership or failed to replace packing")
					}
				}
			})
		}
	}
}

func TestIffWrapperValidLibrarySizes(t *testing.T) {
	for _, tc := range []struct {
		inputs               int
		limit, degree, width int32
	}{
		{2, 7, 2, 4}, {4, 10, 3, 4}, {5, 11, 2, 6}, {6, 16, 3, 6},
	} {
		t.Run(fmt.Sprintf("inputs%d/library%d", tc.inputs, tc.limit), func(t *testing.T) {
			graph, root := iffReviewWide(t, tc.inputs)
			lib := &If_LibLut_t{LutMax: tc.limit}
			for i := range lib.pLutDelays {
				lib.pLutDelays[i][0] = 1
			}
			got := dsdReportCapture(t, func() { Gia_ManIffTest(graph, lib, 1) })
			prefix := fmt.Sprintf("Performing %d-clustering with %d-LUTs:\nMax delay = 1.00.  Count1 = 1.  Count2 = 0.  Count3 = 0.\n", tc.degree, tc.width)
			if !strings.HasPrefix(got, prefix) || !slices.Equal(Vec_IntArray(graph.VPacking), []int32{1, 1, root}) {
				t.Fatalf("wrapper size/packing: %q %v", got, Vec_IntArray(graph.VPacking))
			}
		})
	}
}

func TestIffAccessorAndManagerIsolation(t *testing.T) {
	graph, _, _, root := iffReviewTree(t)
	p, other := Gia_ManIffStart(graph), Gia_ManIffStart(graph)
	defer Gia_ManIffStop(p)
	defer Gia_ManIffStop(other)
	obj := Gia_ManObj(graph, root)
	Iff_ObjSetTime(p, obj, 3.5)
	if Iff_ObjTimeId(p, root) != 3.5 {
		t.Fatal("object setter does not agree with ID getter")
	}
	Iff_ObjSetTimeId(p, root, 7.25)
	if Iff_ObjTime(p, obj) != 7.25 || !math.IsNaN(float64(Iff_ObjTime(other, obj))) {
		t.Fatal("ID setter disagrees or manager arrival storage aliases")
	}
	for _, kind := range []int32{2, 3} {
		Iff_ObjSetMatch(p, obj, kind, 1)
		if Iff_ObjMatchId(p, root, kind) != 1 || Iff_ObjMatch(other, obj, kind) != -1 {
			t.Fatal("match object/ID access or manager isolation changed")
		}
		if kind == 3 && Iff_ObjMatch(p, obj, 2) != 2 {
			t.Fatal("pair and triple match storage aliases")
		}
		Iff_ObjSetMatchId(p, root, kind, 2)
		if Iff_ObjMatch(p, obj, kind) != 2 {
			t.Fatal("match ID setter does not agree with object getter")
		}
	}
}
