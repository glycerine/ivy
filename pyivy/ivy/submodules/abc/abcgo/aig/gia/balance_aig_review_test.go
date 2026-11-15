package gia

import (
	"slices"
	"strconv"
	"testing"
)

func balanceSimpleOppositeReviewGraph() *Gia_Man_t {
	g := Gia_ManStart(1)
	g.fGiaSimple = 1
	a, r := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	zero := Gia_ManAppendAnd(g, a, a^1)
	Gia_ManAppendCo(g, zero)
	Gia_ManAppendCo(g, r)
	Gia_ManSetRegNum(g, 1)
	return g
}

func TestOppositeSimpleAndIsOrdinaryAnd(t *testing.T) {
	g := balanceSimpleOppositeReviewGraph()
	obj := Gia_ObjFanin0(Gia_ManPo(g, 0))
	if Gia_ObjIsBuf(obj) != 0 || Gia_ObjIsAndReal(g, obj) == 0 {
		t.Fatal("simple-mode x & !x must be an ordinary AND, not a buffer or unknown gate")
	}
}

func TestBalancingRetainsOppositeSimpleAndTruth(t *testing.T) {
	for _, area := range []bool{false, true} {
		t.Run(map[bool]string{false: "depth", true: "area"}[area], func(t *testing.T) {
			defer func() {
				if err := recover(); err != nil {
					t.Fatalf("valid simple-mode graph rejected: %v", err)
				}
			}()
			src := balanceSimpleOppositeReviewGraph()
			var out *Gia_Man_t
			if area {
				out = Gia_ManAreaBalance(src, 1, 20, 0, 0)
			} else {
				out = Gia_ManBalance(src, 1, 1, 0)
			}
			for mask := 0; mask < 4; mask++ {
				checkQuantificationOutputs(t, out, []bool{mask&1 != 0, mask&2 != 0}, []bool{false, mask&2 != 0})
			}
		})
	}
}

func TestOrdinaryAndClassifierMatchesLogicalGateKinds(t *testing.T) {
	for phase := int32(0); phase < 8; phase++ {
		for _, simple := range []bool{false, true} {
			g := Gia_ManStart(1)
			g.fGiaSimple = giaBool(simple)
			a, b, c := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
			g.EnableMuxes()
			nodes := []int32{Gia_ManAppendAnd(g, a^(phase&1), b^((phase>>1)&1)), Gia_ManAppendBuf(g, a^(phase&1)), Gia_ManAppendXorReal(g, a^(phase&1), b^((phase>>1)&1)), Gia_ManAppendMuxReal(g, c^((phase>>2)&1), a^(phase&1), b^((phase>>1)&1))}
			if simple {
				nodes = append(nodes, Gia_ManAppendAnd(g, a^(phase&1), a^((phase>>1)&1)))
			}
			for _, lit := range nodes {
				Gia_ManAppendCo(g, lit)
			}
			for _, obj := range g.pObjs {
				_, _, _, kind := g.Gate(obj.id)
				if (Gia_ObjIsAndReal(g, obj) != 0) != (kind == 2) {
					t.Fatalf("phase%d simple%v node%d kind%d", phase, simple, obj.id, kind)
				}
			}
		}
	}
}

func TestSimpleGateBalancingMatrix(t *testing.T) {
	cases := 0
	for signs := 0; signs < 32; signs++ {
		for _, layout := range []bool{false, true} {
			for simple := int32(0); simple < 2; simple++ {
				for mode := 0; mode < 4; mode++ {
					src := Gia_ManStart(1)
					src.fGiaSimple = 1
					a, b, r := Gia_ManAppendCi(src), Gia_ManAppendCi(src), Gia_ManAppendCi(src)
					phase := func(bit uint) int32 { return int32(signs>>bit) & 1 }
					x := Gia_ManAppendAnd(src, a^phase(0), a^phase(1))
					y := Gia_ManAppendAnd(src, x, b)
					z := Gia_ManAppendAnd(src, r^phase(2), r^phase(3))
					next := Gia_ManAppendAnd(src, y, r)
					for _, lit := range []int32{x, y, z ^ phase(4), next} {
						Gia_ManAppendCo(src, lit)
					}
					Gia_ManSetRegNum(src, 1)
					if layout {
						src = Gia_ManDupOrderDfsReverse(src)
					}
					if err := src.SetAigerProperties(1, 1); err != nil {
						t.Fatal(err)
					}
					var out *Gia_Man_t
					if mode < 2 {
						out = Gia_ManBalance(src, simple, int32(mode), 0)
					} else {
						limit := int32(0)
						if mode == 3 {
							limit = 20
						}
						out = Gia_ManAreaBalance(src, simple, limit, 0, 0)
					}
					restored := partitionReviewRoundTrip(t, out, 1, 1)
					for mask := 0; mask < 8; mask++ {
						inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
						want, err := src.Evaluate(inputs)
						if err != nil {
							t.Fatal(err)
						}
						checkQuantificationOutputs(t, out, inputs, want)
						checkQuantificationOutputs(t, restored, inputs, want)
					}
					cases++
				}
			}
		}
	}
	t.Logf("%d simple-mode depth/area copies and %d truth rows including serialization", cases, cases*8)
}

func balanceGateBorrowReviewSetup(xor bool) (*Gia_Man_t, *Gia_Obj_t) {
	g := Gia_ManStart(64)
	inputs := make([]int32, 16)
	for i := range inputs {
		inputs[i] = Gia_ManAppendCi(g)
	}
	g.EnableMuxes()
	Gia_ManCleanLevels(g, g.nObjs)
	for i, lit := range inputs {
		Gia_ObjSetLevelId(g, lit>>1, int32(i%4))
	}
	Gia_ManHashStart(g)
	template := Gia_ManHashAnd(g, inputs[0], inputs[15])
	if xor {
		template = Gia_ManHashXorReal(g, inputs[0], inputs[15])
	}
	return g, Gia_ManObj(g, template>>1)
}

func TestBalanceGateBorrowedInputSpans(t *testing.T) {
	for _, xor := range []bool{false, true} {
		for n := 1; n <= 12; n++ {
			for offset := 0; offset < 3; offset++ {
				for phase := int32(0); phase < 4; phase++ {
					owned, ownedTemplate := balanceGateBorrowReviewSetup(xor)
					borrowed, borrowedTemplate := balanceGateBorrowReviewSetup(xor)
					input := make([]int32, offset+n)
					for i := range input {
						input[i] = int32((i%8+1)<<1) ^ ((phase >> uint(i%2)) & 1)
					}
					first, second := permutation(input...), permutation(input...)
					literals := append([]int32(nil), input[offset:]...)
					want := Gia_ManBalanceGate(owned, ownedTemplate, first, literals, int32(n))
					got := Gia_ManBalanceGate(borrowed, borrowedTemplate, second, Vec_IntArray(second)[offset:], int32(n))
					if got != want || borrowed.nObjs != owned.nObjs || !slices.Equal(Vec_IntArray(first), Vec_IntArray(second)) {
						t.Fatalf("xor%v n%d offset%d phase%d: borrowed input changed balancing", xor, n, offset, phase)
					}
					for mask := 0; mask < 256; mask++ {
						inputs := make([]bool, 16)
						for i := range inputs {
							inputs[i] = mask>>uint(i%8)&1 != 0
						}
						values, _, err := borrowed.evaluateObjects(inputs)
						if err != nil {
							t.Fatal(err)
						}
						expected := !xor
						for _, lit := range literals {
							value := values[lit>>1] != (lit&1 != 0)
							if xor {
								expected = expected != value
							} else {
								expected = expected && value
							}
						}
						if (values[got>>1] != (got&1 != 0)) != expected {
							t.Fatal("balanced gate changed Boolean reduction")
						}
					}
				}
			}
		}
	}
}

func BenchmarkBalanceGateReuse(b *testing.B) {
	for _, n := range []int{2, 8, 12} {
		b.Run(strconv.Itoa(n), func(b *testing.B) {
			g, template := balanceGateBorrowReviewSetup(false)
			literals := make([]int32, n)
			for i := range literals {
				literals[i] = int32((i + 1) << 1)
			}
			super := Vec_IntAlloc(32)
			Gia_ManBalanceGate(g, template, super, literals, int32(n))
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				Gia_ManBalanceGate(g, template, super, literals, int32(n))
			}
		})
	}
}
