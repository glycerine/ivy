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
	edge := int32(0)
	if Gia_ObjFanin0(fanout) == obj {
		edge = target << 1
	} else if Gia_ObjIsAnd(fanout) != 0 && Gia_ObjFanin1(fanout) == obj {
		edge = target<<1 | 1
	} else {
		panic("gia: object is not a fanin")
	}
	max := source
	if target > max {
		max = target
	}
	if max >= p.nFansAlloc {
		count := max * 2
		if count < 16 {
			count = 16
		}
		p.pFanData = append(p.pFanData, make([]int32, 5*int(count)-len(p.pFanData))...)
		p.nFansAlloc = count
	}
	prev := func(e int32) int { return int(5*(e>>1) + 1 + (e & 1)) }
	next := func(e int32) int { return int(5*(e>>1) + 3 + (e & 1)) }
	first := p.pFanData[5*source]
	if first == 0 {
		p.pFanData[5*source] = edge
		p.pFanData[prev(edge)] = edge
		p.pFanData[next(edge)] = edge
		return
	}
	last := p.pFanData[prev(first)]
	if p.pFanData[next(last)] != first {
		panic("gia: corrupt fanout ring")
	}
	p.pFanData[prev(edge)] = last
	p.pFanData[next(edge)] = first
	p.pFanData[prev(first)] = edge
	p.pFanData[next(last)] = edge
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
