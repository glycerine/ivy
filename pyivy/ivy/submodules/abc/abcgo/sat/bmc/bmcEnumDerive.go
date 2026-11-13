package bmc

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

// Gia_ManDeriveOne builds the condition preserving each output's value for the
// given input pattern, choosing controlling fanins as in the source routine.
func Gia_ManDeriveOne(p *gia.Gia_Man_t, pattern *vec.Vec_Int_t) *gia.Gia_Man_t {
	if p == nil || pattern == nil || vec.Vec_IntSize(pattern) != int(gia.Gia_ManCiNum(p)) {
		return nil
	}
	q := gia.Gia_ManStart(max(1, gia.Gia_ManObjNum(p)))
	values := make([]justifiedBit, gia.Gia_ManObjNum(p))
	values[0] = justifiedBit{false, 0}
	for i, bit := range vec.Vec_IntArray(pattern) {
		if bit != 0 && bit != 1 {
			panic("bmc: invalid input pattern")
		}
		lit := gia.Gia_ManAppendCi(q)
		if bit == 0 {
			lit ^= 1
		}
		values[gia.Gia_ObjId(p, gia.Gia_ManCi(p, int32(i)))] = justifiedBit{bit != 0, lit}
	}
	gia.Gia_ManHashStart(q)
	read := func(lit int32) justifiedBit {
		v := values[lit>>1]
		if lit&1 != 0 {
			v = justifiedNot(v)
		}
		return v
	}
	for id := int32(1); id < gia.Gia_ManObjNum(p); id++ {
		a, b, c, kind := p.Gate(id)
		switch kind {
		case 2:
			values[id] = justifiedAnd(q, read(a), read(b))
		case 3:
			x, y := read(a), read(b)
			values[id] = justifiedBit{x.value != y.value, gia.Gia_ManHashAnd(q, x.dep, y.dep)}
		case 4:
			ctl, yes, no := read(a), read(b), read(c)
			x := justifiedAnd(q, ctl, yes)
			y := justifiedAnd(q, justifiedNot(ctl), no)
			values[id] = justifiedNot(justifiedAnd(q, justifiedNot(x), justifiedNot(y)))
		case 5, 6:
			values[id] = read(a)
		}
	}
	for i := int32(0); i < gia.Gia_ManCoNum(p); i++ {
		gia.Gia_ManAppendCo(q, values[gia.Gia_ObjId(p, gia.Gia_ManCo(p, i))].dep)
	}
	gia.Gia_ManSetRegNum(q, gia.Gia_ManRegNum(p))
	gia.Gia_ManHashStop(q)
	return q
}
func Gia_ManDeriveOneTest2(p *gia.Gia_Man_t) *gia.Gia_Man_t {
	if p == nil {
		return nil
	}
	return Gia_ManDeriveOne(p, vec.Vec_IntStart(int(gia.Gia_ManCiNum(p))))
}
