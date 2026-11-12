package gia

import "github.com/glycerine/abcgo/misc/vec"

// Gia_ManAutomTranspose64 transposes the bits of a 64 by 64 matrix in place.
func Gia_ManAutomTranspose64(a []word) {
	if len(a) < 64 {
		panic("gia: transpose needs 64 words")
	}
	mask := word(0x00000000ffffffff)
	for shift := uint(32); shift != 0; {
		for k := 0; k < 64; k = (k + int(shift) + 1) &^ int(shift) {
			t := (a[k] ^ (a[k+int(shift)] >> shift)) & mask
			a[k] ^= t
			a[k+int(shift)] ^= t << shift
		}
		shift >>= 1
		mask ^= mask << shift
	}
}
func Gia_ManAutomSim0(p *Gia_Man_t, obj *Gia_Obj_t, temp *Vec_Wrd_t) word {
	lit := Gia_ObjFaninLit0(obj, Gia_ObjId(p, obj))
	value := vec.Vec_WrdEntry(temp, int(lit>>1))
	if lit&1 != 0 {
		value = ^value
	}
	return value
}
func Gia_ManAutomSim1(p *Gia_Man_t, obj *Gia_Obj_t, temp *Vec_Wrd_t) word {
	lit := Gia_ObjFaninLit1(obj, Gia_ObjId(p, obj))
	value := vec.Vec_WrdEntry(temp, int(lit>>1))
	if lit&1 != 0 {
		value = ^value
	}
	return value
}

// Gia_ManAutomStep evaluates all one-hot inputs simultaneously.
func Gia_ManAutomStep(p *Gia_Man_t, current word, next []word, temp *Vec_Wrd_t) word {
	if Gia_ManPoNum(p) != 1 || Gia_ManPiNum(p) > 64 || Gia_ManRegNum(p) > 64 || len(next) < 64 || vec.Vec_WrdSize(temp) < int(p.nObjs) {
		panic("gia: invalid automaton simulation dimensions")
	}
	vec.Vec_WrdWriteEntry(temp, 0, 0)
	for i := int32(0); i < Gia_ManPiNum(p); i++ {
		vec.Vec_WrdWriteEntry(temp, int(Gia_ObjId(p, Gia_ManPi(p, i))), word(1)<<(63-i))
	}
	for i := int32(0); i < Gia_ManRegNum(p); i++ {
		var v word
		if current&(word(1)<<(63-i)) != 0 {
			v = ^word(0)
		}
		vec.Vec_WrdWriteEntry(temp, int(Gia_ObjId(p, Gia_ManRo(p, i))), v)
	}
	read := func(lit int32) word {
		v := vec.Vec_WrdEntry(temp, int(lit>>1))
		if lit&1 != 0 {
			return ^v
		}
		return v
	}
	for id := int32(1); id < p.nObjs; id++ {
		if Gia_ObjIsAnd(Gia_ManObj(p, id)) == 0 {
			continue
		}
		a, b, c, kind := p.Gate(id)
		var v word
		switch kind {
		case 2:
			v = read(a) & read(b)
		case 3:
			v = read(a) ^ read(b)
		case 4:
			v = (read(a) & read(b)) | (^read(a) & read(c))
		case 5:
			v = read(a)
		default:
			panic("gia: unsupported automaton gate")
		}
		vec.Vec_WrdWriteEntry(temp, int(id), v)
	}
	for i := range next[:64] {
		next[i] = 0
	}
	for i := int32(0); i < Gia_ManRegNum(p); i++ {
		next[i] = Gia_ManAutomSim0(p, Gia_ManRi(p, i), temp)
	}
	Gia_ManAutomTranspose64(next)
	return Gia_ManAutomSim0(p, Gia_ManPo(p, 0), temp)
}
