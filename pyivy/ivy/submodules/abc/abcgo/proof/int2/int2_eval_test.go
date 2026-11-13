package int2

import "github.com/glycerine/abcgo/aig/gia"

func int2EvalGraph(p *Gia_Man_t, inputs, state, values, next []bool) bool {
	bad, _ := int2EvalFrame(p, inputs, state, values, next)
	return bad
}

// int2EvalFrame reports whether a property fails and whether the input is
// legal. Constraint outputs are violation signals, following ABC's convention.
func int2EvalFrame(p *Gia_Man_t, inputs, state, values, next []bool) (bool, bool) {
	clear(values)
	for i, bit := range inputs {
		obj := gia.Gia_ManPi(p, int32(i))
		values[gia.Gia_ObjId(p, obj)] = bit
	}
	for i, bit := range state {
		obj := gia.Gia_ManRo(p, int32(i))
		values[gia.Gia_ObjId(p, obj)] = bit
	}
	for id := 1; id < len(values); id++ {
		a, b, c, kind := p.Gate(int32(id))
		read := func(lit int32) bool { return values[lit>>1] != (lit&1 != 0) }
		switch kind {
		case 2:
			values[id] = read(a) && read(b)
		case 3:
			values[id] = read(a) != read(b)
		case 4:
			if read(a) {
				values[id] = read(b)
			} else {
				values[id] = read(c)
			}
		case 5, 6:
			values[id] = read(a)
		}
	}
	nPos := gia.Gia_ManPoNum(p)
	nConstraints := gia.Gia_ManConstrNum(p)
	for i := nPos - nConstraints; i < nPos; i++ {
		co := gia.Gia_ManPo(p, i)
		if values[gia.Gia_ObjId(p, gia.Gia_ObjFanin0(co))] != (gia.Gia_ObjFaninC0(co) != 0) {
			return false, false
		}
	}
	for i := int32(0); i < nPos-nConstraints; i++ {
		co := gia.Gia_ManPo(p, i)
		bad := values[gia.Gia_ObjId(p, gia.Gia_ObjFanin0(co))] != (gia.Gia_ObjFaninC0(co) != 0)
		if bad {
			return true, true
		}
	}
	for i := range next {
		co := gia.Gia_ManRi(p, int32(i))
		next[i] = values[gia.Gia_ObjId(p, gia.Gia_ObjFanin0(co))] != (gia.Gia_ObjFaninC0(co) != 0)
	}
	return false, true
}
