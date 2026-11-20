package cnf

import (
	"slices"
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

func TestPostprocessFoldsOneReferencedCutByCombinedCost(t *testing.T) {
	for _, tc := range []struct {
		name                                string
		leftCost, rightCost, rootCost       byte
		shared, unused, foldRight, foldLeft bool
	}{
		{"cheaper second leaf", 5, 3, 3, false, false, true, false},
		{"ties keep leaf order", 3, 3, 3, false, false, false, true},
		{"equal combined cost", 2, 2, 2, false, false, false, true},
		{"exceeds combined cost", 1, 1, 2, false, false, false, false},
		{"shared cheapest leaf", 2, 3, 3, true, false, true, false},
		{"unreferenced root", 3, 3, 9, false, true, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := aig.NewGraph()
			a, b, c, d := g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput()
			left, right := g.AND(a, b), g.AND(c, d)
			root := g.AND(left, right)
			if !tc.unused {
				g.AddOutput(root)
			}
			if tc.shared {
				g.AddOutput(left)
			}
			m := Cnf_ManStart()
			defer Cnf_ManStop(m)
			m.pManAig = g
			assign := func(e, x, y aig.Edge, cost byte) *Cnf_Cut_t {
				cut := Cnf_CutFromTruth([]int32{int32(x.NodeID()), int32(y.NodeID())}, []uint32{0x88888888})
				cut.Cost = cost
				Cnf_ObjSetBestCut(m, &e, cut)
				return cut
			}
			origLeft := assign(left, a, b, tc.leftCost)
			origRight := assign(right, c, d, tc.rightCost)
			origRoot := assign(root, left, right, tc.rootCost)
			Cnf_ManScanMapping(m, 0, 0)
			Cnf_ManPostprocess(m)
			cut := Cnf_ObjBestCut(m, &root)
			if !tc.foldLeft && !tc.foldRight {
				if cut != origRoot || Cnf_ObjBestCut(m, &left) != origLeft || Cnf_ObjBestCut(m, &right) != origRight {
					t.Fatal("changed excluded or uneconomical mapping")
				}
				return
			}
			removed, kept := left, right
			want := []int32{int32(a.NodeID()), int32(b.NodeID()), int32(right.NodeID())}
			if tc.foldRight {
				removed, kept = right, left
				want = []int32{int32(c.NodeID()), int32(d.NodeID()), int32(left.NodeID())}
			}
			if cut == nil || !slices.Equal(cut.pFanins, want) || cut.Cost != 4 || cut.truth[0]&255 != 0x80 {
				t.Fatalf("postprocessed root %+v, want leaves %v at cost 4", cut, want)
			}
			if Cnf_ObjBestCut(m, &removed) != nil || g.Node(removed).Refs != 0 || Cnf_ObjBestCut(m, &kept) == nil {
				t.Fatal("absorbed gate retained or other fanin lost")
			}
			before := g.Nodes()
			Cnf_ManScanMapping(m, 0, 0)
			for _, e := range g.LiveEdges() {
				if before[e.NodeID()].Refs != g.Node(e).Refs {
					t.Errorf("reference mismatch at %d", e.NodeID())
				}
			}
		})
	}
}

func TestSmallComposedCutUsesNativeMinimumSOP(t *testing.T) {
	m := Cnf_ManStart()
	defer Cnf_ManStop(m)
	fan := Cnf_CutFromTruth([]int32{4}, []uint32{0xaaaaaaaa})
	top := Cnf_CutFromTruth([]int32{1, 2, 3, 5}, []uint32{0})
	for f := 0; f <= 0xffff; f++ {
		top.truth[0] = uint32(f) * 0x10001
		c := Cnf_CutCompose(m, top, fan, 5)
		want := int(m.pSopSizes[f]) + int(m.pSopSizes[^uint16(f)])
		if c == nil || c.truth[0]&0xffff != uint32(f) || int(c.Cost) != want {
			t.Fatalf("truth %04x composed cut %+v, want cost %d", f, c, want)
		}
	}
}

func TestComposedAndsRequireEveryLeaf(t *testing.T) {
	for _, tc := range []struct {
		n     int
		truth uint32
	}{{2, 0x88888888}, {4, 0x80008000}} {
		top := Cnf_CutFromTruth([]int32{5, 6, 7, 8}[:tc.n], []uint32{tc.truth})
		fan := Cnf_CutFromTruth([]int32{1, 2, 3, 4}[:tc.n], []uint32{tc.truth})
		r := Cnf_CutCompose(nil, top, fan, 5)
		rows := 1 << (2*tc.n - 1)
		for row := range rows {
			if got := r.truth[row>>5] >> uint(row&31) & 1; (got != 0) != (row == rows-1) {
				t.Fatalf("%d leaves, row %d, got %d", r.nFanins, row, got)
			}
		}
	}
}

func BenchmarkComposeMappedCuts(b *testing.B) {
	for _, tc := range []struct {
		name  string
		n     int
		truth uint32
		cost  byte
	}{{"ThreeLeaves", 2, 0x88888888, 4}, {"SevenLeaves", 4, 0x80008000, 8}} {
		b.Run(tc.name, func(b *testing.B) {
			m := Cnf_ManStart()
			defer Cnf_ManStop(m)
			top := Cnf_CutFromTruth([]int32{5, 6, 7, 8}[:tc.n], []uint32{tc.truth})
			fan := Cnf_CutFromTruth([]int32{1, 2, 3, 4}[:tc.n], []uint32{tc.truth})
			b.ReportAllocs()
			for b.Loop() {
				r := Cnf_CutCompose(m, top, fan, 5)
				if r == nil || r.Cost != tc.cost || int(r.nFanins) != 2*tc.n-1 {
					b.Fatal(r)
				}
				Cnf_CutFree(r)
			}
		})
	}
}

func TestTruthPhaseUsesPositionsInMergedCut(t *testing.T) {
	merged := Cnf_CutFromTruth([]int32{2, 5, 7, 11, 13}, nil)
	subset := Cnf_CutFromTruth([]int32{5, 11}, nil)
	if got := Cnf_TruthPhase(merged, subset); got != 0b01010 {
		t.Fatalf("phase %05b, want bits 1 and 3", got)
	}
}

func TestComposeMixedSupportsAcrossWords(t *testing.T) {
	top := Cnf_CutFromTruth([]int32{1, 3, 5, 7, 9}, []uint32{0xc8674e23})
	fan := Cnf_CutFromTruth([]int32{2, 3, 4, 6, 8}, []uint32{0x394f952a})
	full := Cnf_CutCompose(nil, top, fan, 5)
	if full == nil || !slices.Equal(full.pFanins, []int32{1, 2, 3, 4, 6, 7, 8, 9}) {
		t.Fatalf("merged leaves %+v", full)
	}
	for row := range 256 {
		pick := func(bit int) int { return row >> bit & 1 }
		fanRow := pick(1) | pick(2)<<1 | pick(3)<<2 | pick(4)<<3 | pick(6)<<4
		fanBit := int(fan.truth[0] >> fanRow & 1)
		topRow := pick(0) | pick(2)<<1 | fanBit<<2 | pick(5)<<3 | pick(7)<<4
		if (full.truth[row>>5]>>uint(row&31))&1 != top.truth[0]>>topRow&1 {
			t.Fatalf("truth row %d", row)
		}
	}
}

func TestPostprocessRejectsSaturatedCoverCost(t *testing.T) {
	g := aig.NewGraph()
	in := make([]aig.Edge, 7)
	for i := range in {
		in[i] = g.NewInput()
	}
	fan := g.NewXor(in[0], in[1])
	root := fan
	for _, e := range in[2:] {
		root = g.NewXor(root, e)
	}
	g.AddOutput(root)
	m := Cnf_ManStart()
	defer Cnf_ManStop(m)
	m.pManAig = g
	inner := Cnf_CutFromTruth([]int32{int32(in[0].NodeID()), int32(in[1].NodeID())}, []uint32{0x66666666})
	inner.Cost = 4
	leaves := []int32{}
	for _, e := range in[2:] {
		leaves = append(leaves, int32(e.NodeID()))
	}
	leaves = append(leaves, int32(fan.NodeID()))
	outer := Cnf_CutFromTruth(leaves, []uint32{0x96696996, 0x69969669})
	outer.Cost = 126
	Cnf_ObjSetBestCut(m, &fan, inner)
	Cnf_ObjSetBestCut(m, &root, outer)
	Cnf_ManScanMapping(m, 0, 0)
	candidate := Cnf_CutCompose(m, outer, inner, int32(fan.NodeID()))
	if candidate.Cost != 127 {
		t.Fatalf("expected saturated parity cost, got %d", candidate.Cost)
	}
	Cnf_ManPostprocess(m)
	if Cnf_ObjBestCut(m, &fan) != inner || Cnf_ObjBestCut(m, &root) != outer || g.Node(fan).Refs != 1 {
		t.Fatal("absorbed a cut with saturated cover cost")
	}
}
