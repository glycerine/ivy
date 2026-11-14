package gia

import "github.com/glycerine/abcgo/misc/vec"

// Dynamic fanouts retain the source's five-word circular-list record layout.
func Gia_ObjAddFanout(p *Gia_Man_t, obj, fanout *Gia_Obj_t) {
	if p.pFanData == nil {
		panic("gia: fanout storage is disabled")
	}
	source, target := Gia_ObjId(p, obj), Gia_ObjId(p, fanout)
	if target <= 0 {
		panic("gia: invalid fanout target")
	}
	if lastID := max(source, target); lastID >= p.nFansAlloc {
		count := max(16, lastID*2)
		p.pFanData = append(p.pFanData, make([]int32, 5*int(count)-len(p.pFanData))...)
		p.nFansAlloc = count
	}
	edge := giaFanoutEdge(p, obj, fanout, false)
	prev, next := Gia_FanoutPrev(p.pFanData, edge), Gia_FanoutNext(p.pFanData, edge)
	first := Gia_FanoutObj(p.pFanData, source)
	if *first == 0 {
		*first, *prev, *next = edge, edge, edge
		return
	}
	last := *Gia_FanoutPrev(p.pFanData, *first)
	if *Gia_FanoutNext(p.pFanData, last) != *first {
		panic("gia: corrupt fanout ring")
	}
	*prev, *next = last, *first
	*Gia_FanoutPrev(p.pFanData, *first) = edge
	*Gia_FanoutNext(p.pFanData, last) = edge
}

func Gia_ManBuiltInSimPerform(p *Gia_Man_t, id int32) {
	for w := int32(0); w < p.NSimWords; w++ {
		vec.Vec_WrdPush(p.vSims, 0)
	}
	Gia_ManBuiltInSimPerformInt(p, id)
}
func Gia_ManBuiltInSimPerformInt(p *Gia_Man_t, id int32) {
	if p.fBuiltInSim == 0 && p.fIncrSim == 0 {
		panic("gia: built-in simulation disabled")
	}
	data := vec.Vec_WrdArray(p.vSims)
	if len(data) != int(p.nObjs*p.NSimWords) {
		panic("gia: invalid simulation storage size")
	}
	obj := Gia_ManObj(p, id)
	a, b := Gia_ObjFaninId0(obj, id), Gia_ObjFaninId1(obj, id)
	for w := int32(0); w < p.NSimWords; w++ {
		x, y := data[a*p.NSimWords+w], data[b*p.NSimWords+w]
		if obj.fCompl0 != 0 {
			x = ^x
		}
		if obj.fCompl1 != 0 {
			y = ^y
		}
		data[id*p.NSimWords+w] = x & y
	}
}
func Gia_ManQuantSetSuppAnd(p *Gia_Man_t, obj *Gia_Obj_t) {
	if Gia_ObjIsAnd(obj) == 0 {
		panic("gia: AND support requested for non-AND")
	}
	for w := int32(0); w < p.nSuppWords; w++ {
		vec.Vec_WrdPush(p.vSuppWords, 0)
	}
	data := vec.Vec_WrdArray(p.vSuppWords)
	if len(data) != int(p.nObjs*p.nSuppWords) {
		panic("gia: invalid support storage size")
	}
	id := obj.id
	a, b := Gia_ObjFaninId0(obj, id), Gia_ObjFaninId1(obj, id)
	for w := int32(0); w < p.nSuppWords; w++ {
		data[id*p.nSuppWords+w] = data[a*p.nSuppWords+w] | data[b*p.nSuppWords+w]
	}
}
