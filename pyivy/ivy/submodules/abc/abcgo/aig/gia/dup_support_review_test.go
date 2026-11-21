package gia

import (
	"fmt"
	"os/exec"
	"slices"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/misc/vec"
)

func supportReviewGraph(seed uint32, typed, layout bool, regs int32) *Gia_Man_t {
	g := Gia_ManStart(1)
	g.fGiaSimple = 1
	if typed {
		g.EnableMuxes()
	}
	lits := []int32{0, 1}
	for i := 0; i < 6; i++ {
		lits = append(lits, Gia_ManAppendCi(g))
	}
	next := func() int32 { seed = seed*1664525 + 1013904223; return int32(seed % uint32(len(lits))) }
	for i := 0; i < 48; i++ {
		a, b, c := lits[next()], lits[next()], lits[next()]
		var lit int32
		switch {
		case typed && i%4 == 1:
			if a>>1 == b>>1 {
				b = Gia_ManAppendBuf(g, b)
			}
			lit = Gia_ManAppendXorReal(g, a, b^1)
		case typed && i%4 == 2:
			// Raw MUX storage requires three distinct fanin objects.
			a, b, c = Gia_ManAppendBuf(g, a), Gia_ManAppendBuf(g, b), Gia_ManAppendBuf(g, c)
			lit = Gia_ManAppendMuxReal(g, a, b, c^1)
		case typed && i%4 == 3:
			lit = Gia_ManAppendBuf(g, a^1)
		default:
			if a>>1 == b>>1 {
				b = lits[2]
				if a>>1 == b>>1 {
					b = lits[3]
				}
			}
			lit = Gia_ManAppendAnd(g, a, b^1)
		}
		lits = append(lits, lit)
	}
	for _, lit := range []int32{0, 1, lits[2], lits[2] ^ 1, lits[3], lits[8], lits[14], lits[23], lits[37], lits[49], lits[55], lits[55] ^ 1} {
		Gia_ManAppendCo(g, lit)
	}
	Gia_ManSetRegNum(g, regs)
	if layout {
		g = Gia_ManDupOrderDfsReverse(g)
	}
	return g
}

// Discover structural CI reachability by a fresh walk from each root; neither
// propagation, sorted merges nor semantic cancellation determine this oracle.
func supportReviewReachable(g *Gia_Man_t, id int32) []int32 {
	seen := make(map[int32]bool)
	found := make(map[int32]bool)
	var walk func(int32)
	walk = func(id int32) {
		if id == 0 || seen[id] {
			return
		}
		seen[id] = true
		obj := g.pObjs[id]
		if Gia_ObjIsCi(obj) != 0 {
			found[Gia_ObjCioId(obj)] = true
			return
		}
		walk(Gia_ObjFaninId0(obj, id))
		if Gia_ObjIsCo(obj) != 0 || Gia_ObjIsBuf(obj) != 0 {
			return
		}
		walk(Gia_ObjFaninId1(obj, id))
		if Gia_ObjIsMux(g, obj) != 0 {
			walk(Gia_ObjFanin2(g, obj).id)
		}
	}
	walk(id)
	var result []int32
	for i := int32(0); i < Gia_ManCiNum(g); i++ {
		if found[i] {
			result = append(result, i)
		}
	}
	return result
}

func supportReviewRecords(t *testing.T, typed bool) string {
	t.Helper()
	var records strings.Builder
	for seed := uint32(0); seed < 32; seed++ {
		for _, layout := range []bool{false, true} {
			for _, regs := range []int32{0, 2} {
				g := supportReviewGraph(seed, typed, layout, regs)
				nodes := Vec_IntAlloc(int(g.nObjs) + 2)
				// Reverse order and repeat roots: queries must be independent and sorted.
				for id := g.nObjs - 1; id >= 0; id-- {
					Vec_IntPush(nodes, id)
				}
				Vec_IntPush(nodes, 0)
				Vec_IntPush(nodes, g.nObjs-1)
				for _, obj := range g.pObjs {
					obj.Value = uint32(obj.id + 42)
					obj.fMark0 = uint8(obj.id & 1)
					obj.fMark1 = uint8((obj.id >> 1) & 1)
				}
				result := Gia_ManCreateNodeSupps(g, nodes, 0)
				cos := Gia_ManCreateCoSupps(g, 0)
				fmt.Fprintf(&records, "%d %t %d", seed, layout, regs)
				for i, id := range Vec_IntArray(nodes) {
					got := vec.Vec_IntArray(vec.Vec_WecEntry(result, i))
					want := supportReviewReachable(g, id)
					if !slices.Equal(got, want) {
						t.Fatalf("seed%d typed%v layout%v regs%d root%d: support %v want %v", seed, typed, layout, regs, id, got, want)
					}
				}
				// The original C node routine does not populate CO entries. Compare its
				// CI/internal requests and the separate CO routine, rather than that gap.
				for id, obj := range g.pObjs {
					if Gia_ObjIsCo(obj) == 0 {
						supp := supportReviewReachable(g, int32(id))
						fmt.Fprintf(&records, " n%d:%d", id, len(supp))
						for _, ci := range supp {
							fmt.Fprintf(&records, ",%d", ci)
						}
					}
					if obj.Value != uint32(obj.id+42) || obj.fMark0 != uint8(obj.id&1) || obj.fMark1 != uint8((obj.id>>1)&1) {
						t.Fatal("support query changed source scratch")
					}
				}
				maxSize, largest := int32(1), int32(-1)
				sizes := make([]int, Gia_ManCoNum(g))
				for i := int32(0); i < Gia_ManCoNum(g); i++ {
					got := vec.Vec_IntArray(vec.Vec_WecEntry(cos, int(i)))
					want := supportReviewReachable(g, Gia_ManCo(g, i).id)
					if !slices.Equal(got, want) {
						t.Fatalf("CO%d support %v want %v", i, got, want)
					}
					sizes[i] = len(want)
					maxSize = max(maxSize, int32(len(want)))
					if largest < 0 || len(want) > sizes[largest] {
						largest = i
					}
					fmt.Fprintf(&records, " c%d:%d", i, len(got))
					for _, ci := range got {
						fmt.Fprintf(&records, ",%d", ci)
					}
				}
				if Gia_ManCoSuppSizeMax(g, cos) != maxSize || Gia_ManCoLargestSupp(g, cos) != largest {
					t.Fatal("wrong largest/max support")
				}
				order := Vec_IntArray(Gia_ManSortCoBySuppSize(g, cos))
				seen := make(map[int32]bool)
				prev := int(^uint(0) >> 1)
				for _, id := range order {
					if id < 0 || int(id) >= len(sizes) || seen[id] || sizes[id] > prev {
						t.Fatal("support sort is not a descending permutation")
					}
					seen[id] = true
					prev = sizes[id]
				}
				if len(order) != len(sizes) {
					t.Fatal("sort dropped output")
				}
				fmt.Fprintf(&records, " max%d largest%d\n", maxSize, largest)
				// Public vectors own their storage even if internal supports share lists.
				if vec.Vec_IntSize(vec.Vec_WecEntry(result, 0)) > 0 {
					vec.Vec_IntWriteEntry(vec.Vec_WecEntry(result, 0), 0, 999)
					repeated := Gia_ManCreateNodeSupps(g, nodes, 0)
					if vec.Vec_IntEntry(vec.Vec_WecEntry(repeated, 0), 0) == 999 {
						t.Fatal("support results alias another query")
					}
					if vec.Vec_IntSize(vec.Vec_WecEntry(result, int(g.nObjs)+1)) > 0 && vec.Vec_IntEntry(vec.Vec_WecEntry(result, int(g.nObjs)+1), 0) == 999 {
						t.Fatal("duplicate roots alias result storage")
					}
				}
			}
		}
	}
	return records.String()
}

func TestStructuralSupportsReview(t *testing.T) {
	supportReviewRecords(t, false)
	supportReviewRecords(t, true)
}

func TestSupportsMatchOriginalC(t *testing.T) {
	exe := compileGiaSourceOracle(t, "supports.c")
	output, err := exec.Command(exe).CombinedOutput()
	if err != nil {
		t.Fatalf("C supports: %v: %s", err, output)
	}
	got, want := strings.Split(strings.TrimSpace(supportReviewRecords(t, false)), "\n"), strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(got) != len(want) {
		t.Fatalf("Go cases%d C cases%d", len(got), len(want))
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("case%d\nGo %s\nC %s", i, got[i], want[i])
		}
	}
}

func TestSupportBoundaryCases(t *testing.T) {
	g := Gia_ManStart(1)
	check := func(maxSize, largest int32, order []int32) {
		t.Helper()
		supp := Gia_ManCreateCoSupps(g, 0)
		if Gia_ManCoSuppSizeMax(g, supp) != maxSize || Gia_ManCoLargestSupp(g, supp) != largest || !slices.Equal(Vec_IntArray(Gia_ManSortCoBySuppSize(g, supp)), order) {
			t.Fatal("empty/constant support boundary")
		}
	}
	check(1, -1, nil)
	Gia_ManAppendCo(g, 0)
	Gia_ManAppendCo(g, 1)
	check(1, 0, []int32{0, 1})
	a := Gia_ManAppendCi(g)
	Gia_ManAppendCo(g, a)
	check(1, 2, []int32{2, 0, 1})
	bad := Vec_IntAlloc(1)
	Vec_IntPush(bad, g.nObjs)
	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("invalid node accepted")
			}
		}()
		Gia_ManCreateNodeSupps(g, bad, 0)
	}()
	for _, f := range []func(*Gia_Man_t, *Vec_Wec_t) int32{Gia_ManCoSuppSizeMax, Gia_ManCoLargestSupp} {
		func() {
			defer func() {
				if recover() == nil {
					t.Fatal("invalid CO count accepted")
				}
			}()
			f(g, vec.Vec_WecStart(0))
		}()
	}
	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("invalid sort CO count accepted")
			}
		}()
		Gia_ManSortCoBySuppSize(g, vec.Vec_WecStart(0))
	}()
}

func BenchmarkCreateCoSupports(b *testing.B) {
	g := Gia_ManStart(1)
	a, c := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	x := Gia_ManAppendAnd(g, a, c)
	for i := 0; i < 256; i++ {
		x = Gia_ManAppendBuf(g, x)
		Gia_ManAppendCo(g, x)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Gia_ManCreateCoSupps(g, 0)
	}
}
