package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
)

func Gia_RsbFindOutputs(p *Gia_Man_t, window, inputs, refs *Vec_Int_t) *Vec_Int_t {
	if vec.Vec_IntSize(refs) < int(p.nObjs) {
		panic("gia: short Rsb2 reference vector")
	}
	isInput := make([]bool, p.nObjs)
	for _, id := range vec.Vec_IntArray(inputs) {
		isInput[id] = true
	}
	for _, id := range vec.Vec_IntArray(window) {
		obj := Gia_ManObj(p, id)
		if !isInput[id] && Gia_ObjIsAnd(obj) != 0 {
			for _, f := range giaObjectFanins(p, obj) {
				vec.Vec_IntAddToEntry(refs, int(f), 1)
			}
		}
	}
	result := vec.Vec_IntAlloc(100)
	for _, id := range vec.Vec_IntArray(window) {
		if !isInput[id] && Gia_ObjFanoutNumId(p, id) != vec.Vec_IntEntry(refs, int(id)) {
			vec.Vec_IntPush(result, id)
		}
	}
	for _, id := range vec.Vec_IntArray(window) {
		obj := Gia_ManObj(p, id)
		if !isInput[id] && Gia_ObjIsAnd(obj) != 0 {
			for _, f := range giaObjectFanins(p, obj) {
				vec.Vec_IntAddToEntry(refs, int(f), -1)
			}
		}
	}
	return result
}
func Gia_RsbDeriveGiaFromWindows(p *Gia_Man_t, window, inputs, outputs *Vec_Int_t) *Gia_Man_t {
	q := Gia_ManStart(p.nObjs)
	q.pName, q.PSpec = p.pName, p.PSpec
	Gia_ManHashAlloc(q)
	copies := make([]int32, p.nObjs)
	for i := range copies {
		copies[i] = -1
	}
	copies[0] = 0
	for _, id := range vec.Vec_IntArray(inputs) {
		copies[id] = Gia_ManAppendCi(q)
	}
	read := func(lit int32) int32 {
		value := copies[lit>>1]
		if value < 0 {
			panic("gia: Rsb2 window is not topologically ordered")
		}
		return value ^ (lit & 1)
	}
	for _, id := range vec.Vec_IntArray(window) {
		if copies[id] >= 0 {
			continue
		}
		a, b, c, kind := p.Gate(id)
		switch kind {
		case 2:
			copies[id] = Gia_ManHashAnd(q, read(a), read(b))
		case 3:
			copies[id] = Gia_ManHashXor(q, read(a), read(b))
		case 4:
			copies[id] = Gia_ManHashMux(q, read(a), read(b), read(c))
		case 5, 6:
			copies[id] = read(a)
		default:
			panic("gia: invalid Rsb2 window object")
		}
	}
	for _, id := range vec.Vec_IntArray(outputs) {
		if copies[id] < 0 {
			panic("gia: missing Rsb2 window output")
		}
		Gia_ManAppendCo(q, copies[id])
	}
	Gia_ManHashStop(q)
	return q
}
func Gia_LutComputeTruth66_rec(p *Gia_Man_t, obj *Gia_Obj_t) word {
	if Gia_ManCiNum(p) > 6 {
		panic("gia: truth table needs at most six inputs")
	}
	patterns := [6]word{0xaaaaaaaaaaaaaaaa, 0xcccccccccccccccc, 0xf0f0f0f0f0f0f0f0, 0xff00ff00ff00ff00, 0xffff0000ffff0000, 0xffffffff00000000}
	memo := make([]word, p.nObjs)
	done := make([]bool, p.nObjs)
	var compute func(int32) word
	compute = func(id int32) word {
		if done[id] {
			return memo[id]
		}
		done[id] = true
		if id == 0 {
			return 0
		}
		o := Gia_ManObj(p, id)
		if Gia_ObjIsCi(o) != 0 {
			memo[id] = patterns[Gia_ObjCioId(o)]
			return memo[id]
		}
		a, b, c, kind := p.Gate(id)
		read := func(lit int32) word {
			v := compute(lit >> 1)
			if lit&1 != 0 {
				v = ^v
			}
			return v
		}
		switch kind {
		case 2:
			memo[id] = read(a) & read(b)
		case 3:
			memo[id] = read(a) ^ read(b)
		case 4:
			x := read(a)
			memo[id] = (x & read(b)) | (^x & read(c))
		case 5, 6:
			memo[id] = read(a)
		default:
			panic("gia: invalid truth-table gate")
		}
		return memo[id]
	}
	return compute(Gia_ObjId(p, obj))
}
func Gia_ManVerifyTwoTruths(a, b *Gia_Man_t) int32 {
	if Gia_ManCoNum(a) != Gia_ManCoNum(b) || Gia_ManCiNum(a) != Gia_ManCiNum(b) {
		panic("gia: inconsistent truth-table interfaces")
	}
	ok := int32(1)
	for i := int32(0); i < Gia_ManCoNum(a); i++ {
		if Gia_LutComputeTruth66_rec(a, Gia_ManCo(a, i)) != Gia_LutComputeTruth66_rec(b, Gia_ManCo(b, i)) {
			fmt.Printf("Verification failed for output %d (out of %d).\n", i, Gia_ManCoNum(a))
			ok = 0
		}
	}
	return ok
}
