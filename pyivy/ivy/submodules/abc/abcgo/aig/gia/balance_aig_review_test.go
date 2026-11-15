package gia

import "testing"

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
