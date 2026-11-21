package gia

import (
	"fmt"
	"os/exec"
	"slices"
	"strings"
	"testing"
)

func windowReviewGraph(signs, typed, layout int) *Gia_Man_t {
	g := Gia_ManStart(32)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	var c, d int32
	if layout&1 == 0 {
		c, d = Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	}
	x := Gia_ManAppendAnd(g, a^int32(signs&1), b^int32(signs>>1&1))
	if layout&1 != 0 {
		Gia_ManAppendCo(g, x^int32(signs&1))
		c = Gia_ManAppendCi(g)
	}
	y := Gia_ManAppendAnd(g, x, c^int32(signs>>2&1))
	if layout&1 != 0 {
		Gia_ManAppendCo(g, y^int32(signs>>1&1))
		d = Gia_ManAppendCi(g)
	}
	var z, m int32
	if typed != 0 {
		g.EnableMuxes()
		z = Gia_ManAppendXorReal(g, b^1, d)
		m = Gia_ManAppendMuxReal(g, a, y, z)
	} else {
		z = Gia_ManAppendAnd(g, b^1, d)
		m = Gia_ManAppendAnd(g, y^1, z)
	}
	if layout&1 == 0 {
		Gia_ManAppendCo(g, x^int32(signs&1))
		Gia_ManAppendCo(g, y^int32(signs>>1&1))
	}
	Gia_ManAppendCo(g, z^int32(signs>>2&1))
	Gia_ManAppendCo(g, m^1)
	if layout&2 != 0 {
		for _, ports := range []*Vec_Int_t{g.vCis, g.vCos} {
			slices.Reverse(Vec_IntArray(ports))
			for i, id := range Vec_IntArray(ports) {
				Gia_ObjSetCioId(g.pObjs[id], int32(i))
			}
		}
	}
	return g
}

func windowReviewSelection(mask int) *Vec_Int_t {
	v := Vec_IntAlloc(4)
	for i := 0; i < 4; i++ {
		if mask>>i&1 != 0 {
			Vec_IntPush(v, int32(i))
		}
	}
	return v
}
func windowReviewPrintList(label string, v *Vec_Int_t) {
	fmt.Print(label)
	for _, id := range Vec_IntArray(v) {
		fmt.Printf(" %d", id)
	}
	fmt.Println()
}
func windowReviewPrintGraph(t *testing.T, g *Gia_Man_t) {
	t.Helper()
	fmt.Printf("graph %d %d %d nodes", g.nObjs, Gia_ManCiNum(g), Gia_ManCoNum(g))
	for _, obj := range g.pObjs {
		a, b, c, kind := g.Gate(obj.id)
		fmt.Printf(" %d", kind)
		switch kind {
		case 2, 3:
			fmt.Printf(":%d:%d", a, b)
		case 4:
			fmt.Printf(":%d:%d:%d", a, b, c)
		case 5, 6:
			fmt.Printf(":%d", a)
		}
	}
	fmt.Println()
	windowReviewPrintList("cis", g.vCis)
	windowReviewPrintList("cos", g.vCos)
	fmt.Print("truth")
	for mask := 0; mask < 1<<Gia_ManCiNum(g); mask++ {
		inputs := make([]bool, Gia_ManCiNum(g))
		for i := range inputs {
			inputs[i] = mask>>i&1 != 0
		}
		values, err := g.Evaluate(inputs)
		if err != nil {
			t.Fatal(err)
		}
		bits := 0
		for i, value := range values {
			if value {
				bits |= 1 << i
			}
		}
		fmt.Printf(" %d", bits)
	}
	fmt.Println()
}
func windowReviewPrintSource(g *Gia_Man_t) {
	fmt.Print("source")
	for _, obj := range g.pObjs {
		fmt.Printf(" %d:%d:%d:%d", obj.Value, obj.fMark0, obj.fMark1, Gia_ObjIsTravIdCurrentId(g, obj.id))
	}
	fmt.Println()
}
func windowReviewMapping(g *Gia_Man_t, mode int) {
	if mode == 0 {
		return
	}
	areaReviewInstallMapping(g)
	if mode == 2 {
		// Collapse each LUT to its CI support to distinguish mapped from AIG depth.
		for _, obj := range g.pObjs {
			if Gia_ObjIsAnd(obj) == 0 {
				continue
			}
			var leaves []int32
			seen := make(map[int32]bool)
			var collect func(int32)
			collect = func(id int32) {
				if seen[id] {
					return
				}
				seen[id] = true
				o := g.pObjs[id]
				if Gia_ObjIsCi(o) != 0 {
					leaves = append(leaves, id)
					return
				}
				if Gia_ObjIsAnd(o) != 0 {
					collect(Gia_ObjFaninId0(o, id))
					collect(Gia_ObjFaninId1(o, id))
				}
			}
			collect(obj.id)
			slices.Sort(leaves)
			Vec_IntWriteEntry(g.VMapping, int(obj.id), int32(Vec_IntSize(g.VMapping)))
			Vec_IntPush(g.VMapping, int32(len(leaves)))
			for _, leaf := range leaves {
				Vec_IntPush(g.VMapping, leaf)
			}
			Vec_IntPush(g.VMapping, obj.id)
		}
	}
}
func TestWindowOperationsMatchOriginalABC(t *testing.T) {
	exe := compileGiaSourceOracle(t, "balance_windows.c")
	native, err := exec.Command(exe).CombinedOutput()
	if err != nil {
		t.Fatalf("native window: %v\n%s", err, native)
	}
	got := hashReviewStdout(t, func() {
		for signs := 0; signs < 8; signs++ {
			for layout := 0; layout < 4; layout++ {
				for mask := 0; mask < 16; mask++ {
					for po := int32(0); po < 2; po++ {
						fmt.Printf("case %d %d %d %d\n", signs, layout, mask, po)
						g := windowReviewGraph(signs, 0, layout)
						Gia_ManFillValue(g)
						for _, obj := range g.pObjs {
							obj.fMark0 = 1
							obj.fMark1 = uint8(obj.id & 1)
						}
						selected := windowReviewSelection(mask)
						var inputs, outputs, gates *Vec_Int_t
						Gia_ManPrepareWin(g, selected, &inputs, &outputs, &gates, po)
						windowReviewPrintList("inputs", inputs)
						windowReviewPrintList("boundary", outputs)
						windowReviewPrintList("gates", gates)
						w := Gia_ManExtractWin(g, selected, po)
						windowReviewPrintGraph(t, w)
						windowReviewPrintSource(g)
						if po == 0 {
							r := Gia_ManInsertWin(g, selected, w)
							windowReviewPrintGraph(t, r)
							windowReviewPrintSource(g)
						}
					}
				}
			}
		}
		for layout := 0; layout < 4; layout++ {
			for mapping := 0; mapping < 3; mapping++ {
				for _, choice := range [][2]int32{{-1, 0}, {0, 0}, {1, 0}, {2, 0}, {3, 0}, {4, 0}, {100, 0}, {1, -10}, {1, 1}, {1, 25}, {1, 50}, {1, 75}, {1, 100}, {1, 150}} {
					g := windowReviewGraph(5, 0, layout)
					windowReviewMapping(g, mapping)
					fmt.Printf("latest %d %d %d %d\n", layout, mapping, choice[0], choice[1])
					windowReviewPrintList("outputs", Gia_ManFindLatest(g, choice[0], choice[1]))
				}
			}
		}
	})
	if got != string(native) {
		gotLines, wantLines := strings.Split(got, "\n"), strings.Split(string(native), "\n")
		for i := 0; i < min(len(gotLines), len(wantLines)); i++ {
			if gotLines[i] != wantLines[i] {
				t.Fatalf("native window line%d differs\nGo: %s\nC:  %s", i+1, gotLines[i], wantLines[i])
			}
		}
		t.Fatalf("native window output length differs: Go%d C%d", len(got), len(native))
	}
	t.Log("1024 native boundary/extraction cases, 512 insertions and 168 mapped/unmapped selection cases match")
}
