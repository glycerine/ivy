package gia

import (
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

type Gia_SimRsbMan_t struct {
	pGia                            *Gia_Man_t
	vTfo, vCands, vFanins, vFanins2 *Vec_Int_t
	vSimsObj, vSimsObj2             *Vec_Wrd_t
	nWords                          int32
	pFunc                           [3][]uint64
}

func Gia_SimRsbAlloc(g *Gia_Man_t) *Gia_SimRsbMan_t {
	p := &Gia_SimRsbMan_t{pGia: g, nWords: giaPatternWords(g), vTfo: vec.Vec_IntAlloc(1000), vCands: vec.Vec_IntAlloc(1000), vFanins: vec.Vec_IntAlloc(10), vFanins2: vec.Vec_IntAlloc(10)}
	for i := range p.pFunc {
		p.pFunc[i] = make([]uint64, p.nWords)
	}
	p.vSimsObj = Gia_ManSimPatSim(g)
	p.vSimsObj2 = vec.Vec_WrdStart(vec.Vec_WrdSize(p.vSimsObj))
	Gia_ManStaticFanoutStart(g)
	return p
}
func Gia_SimRsbFree(p *Gia_SimRsbMan_t) {
	Gia_ManStaticFanoutStop(p.pGia)
	p.vTfo = nil
	p.vCands = nil
	p.vFanins = nil
	p.vFanins2 = nil
	p.vSimsObj = nil
	p.vSimsObj2 = nil
	p.pFunc = [3][]uint64{}
}
func Gia_SimRsbTfo_rec(g *Gia_Man_t, id, fanout int32, tfo *Vec_Int_t) {
	if Gia_ObjIsTravIdCurrentId(g, id) != 0 {
		return
	}
	Gia_ObjSetTravIdCurrentId(g, id)
	for i := int32(0); i < Gia_ObjFanoutNumId(g, id); i++ {
		next := Gia_ObjFanoutId(g, id, i)
		if fanout == -1 || fanout == next {
			Gia_SimRsbTfo_rec(g, next, -1, tfo)
		}
	}
	vec.Vec_IntPush(tfo, id)
}
func Gia_SimRsbTfo(p *Gia_SimRsbMan_t, id, fanout int32) *Vec_Int_t {
	if id <= 0 {
		panic("gia: invalid resubstitution object")
	}
	vec.Vec_IntClear(p.vTfo)
	Gia_ManIncrementTravId(p.pGia)
	Gia_SimRsbTfo_rec(p.pGia, id, fanout, p.vTfo)
	if vec.Vec_IntPop(p.vTfo) != id {
		panic("gia: invalid fanout cone")
	}
	vec.Vec_IntSort(p.vTfo, 0)
	return p.vTfo
}
func Gia_SimRsbFunc(p *Gia_SimRsbMan_t, id int32, fanins *Vec_Int_t, on int32) []uint64 {
	ids := vec.Vec_IntArray(fanins)
	if len(ids) > 16 {
		panic("gia: too many resubstitution fanins")
	}
	truth := make([]uint64, util.Abc_Truth6WordNum(int32(len(ids))))
	data := vec.Vec_WrdArray(p.vSimsObj)
	target := data[id*p.nWords : (id+1)*p.nWords]
	for bit := int32(0); bit < 64*p.nWords; bit++ {
		if giaSimBit(p.pFunc[2], bit) == 0 || giaSimBit(target, bit) != uint64(on) {
			continue
		}
		var mint int32
		for i, fanin := range ids {
			words := data[fanin*p.nWords : (fanin+1)*p.nWords]
			if giaSimBit(words, bit) != 0 {
				mint |= 1 << uint(i)
			}
		}
		truth[mint>>6] |= uint64(1) << uint(mint&63)
	}
	return truth
}
func Gia_SimRsbResubVerify(p *Gia_SimRsbMan_t, id int32, fanins *Vec_Int_t) int32 {
	off, on := Gia_SimRsbFunc(p, id, fanins, 0), Gia_SimRsbFunc(p, id, fanins, 1)
	for i := range off {
		if off[i]&on[i] != 0 {
			return 0
		}
	}
	return 1
}
func Gia_SimRsbSimAndCareSet(g *Gia_Man_t, id int32, obj *Gia_Obj_t, n int32, sims, changed *Vec_Wrd_t) {
	original, modified := vec.Vec_WrdArray(sims), vec.Vec_WrdArray(changed)
	a, b := Gia_ObjFaninId0(obj, id), Gia_ObjFaninId1(obj, id)
	for w := int32(0); w < n; w++ {
		x, y := original[a*n+w], original[b*n+w]
		if Gia_ObjIsTravIdCurrentId(g, a) != 0 {
			x = modified[a*n+w]
		}
		if Gia_ObjIsTravIdCurrentId(g, b) != 0 {
			y = modified[b*n+w]
		}
		if obj.fCompl0 != 0 {
			x = ^x
		}
		if obj.fCompl1 != 0 {
			y = ^y
		}
		if Gia_ObjIsXor(obj) != 0 {
			modified[id*n+w] = x ^ y
		} else {
			modified[id*n+w] = x & y
		}
	}
}
func Gia_SimRsbCareSet(p *Gia_SimRsbMan_t, id int32, tfo *Vec_Int_t) []uint64 {
	original, changed := vec.Vec_WrdArray(p.vSimsObj), vec.Vec_WrdArray(p.vSimsObj2)
	for w := int32(0); w < p.nWords; w++ {
		changed[id*p.nWords+w] = ^original[id*p.nWords+w]
	}
	clear(p.pFunc[2])
	for _, node := range vec.Vec_IntArray(tfo) {
		obj := Gia_ManObj(p.pGia, node)
		if Gia_ObjIsAnd(obj) != 0 {
			Gia_SimRsbSimAndCareSet(p.pGia, node, obj, p.nWords, p.vSimsObj, p.vSimsObj2)
		} else if Gia_ObjIsCo(obj) != 0 {
			fan := Gia_ObjFaninId0(obj, node)
			for w := int32(0); w < p.nWords; w++ {
				p.pFunc[2][w] |= original[fan*p.nWords+w] ^ changed[fan*p.nWords+w]
			}
		} else {
			panic("gia: invalid fanout cone")
		}
	}
	return p.pFunc[2]
}
