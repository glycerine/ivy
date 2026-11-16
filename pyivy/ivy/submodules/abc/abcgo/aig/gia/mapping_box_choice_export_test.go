package gia

import "github.com/glycerine/abcgo/misc/tim"

// NewMappingBoxChoiceGraphForTest builds two choice classes, one feeding each
// of two boxes. Every AND is reachable structurally or through a choice edge.
func NewMappingBoxChoiceGraphForTest(signs int, regs int32) *Gia_Man_t {
	p := Gia_ManStart(32)
	a, b, u, v, w := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	phase := func(bit uint) int32 { return int32(signs>>bit) & 1 }
	preAlt := Gia_ManAppendAnd(p, a^phase(0), b^phase(1))
	pre := Gia_ManAppendAnd(p, a^phase(0), b^phase(1))
	other := Gia_ManAppendAnd(p, a, b^phase(2))
	midAlt := Gia_ManAppendAnd(p, u^phase(3), v^phase(4))
	mid := Gia_ManAppendAnd(p, u^phase(3), v^phase(4))
	last0 := Gia_ManAppendAnd(p, w^phase(5), mid^phase(6))
	last1 := Gia_ManAppendAnd(p, w^phase(7), a)
	for _, lit := range []int32{pre, other, mid, last0, last1, b ^ 1} {
		Gia_ManAppendCo(p, lit)
	}
	p.pSibls = make([]int32, p.nObjs)
	p.pSibls[pre>>1], p.pSibls[mid>>1] = preAlt>>1, midAlt>>1
	Gia_ManSetRegNum(p, regs)
	timing := tim.Tim_ManStart(5, 6)
	tim.Tim_ManSetDelayTables(timing, [][]float32{{0, 2, 2, 1, 2, 3, 4}, {1, 1, 1, 5}})
	tim.Tim_ManCreateBox(timing, 0, 2, 2, 2, 0, 0)
	tim.Tim_ManCreateBox(timing, 2, 1, 4, 1, 1, 0)
	p.SetBoxTiming(timing, nil)
	p.SetIdentity("boxed-choices", "boxed-choices.aig")
	return p
}
