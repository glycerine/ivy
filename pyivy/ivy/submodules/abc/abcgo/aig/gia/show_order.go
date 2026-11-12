package gia

import "github.com/glycerine/abcgo/misc/vec"

func Gia_ShowAddOut(adds, mapping *Vec_Int_t, node int32) int32 {
	box := vec.Vec_IntEntry(mapping, int(node))
	if box >= 0 {
		return vec.Vec_IntEntry(adds, int(6*box+4))
	}
	return node
}

func Gia_ShowMapAdds(p *Gia_Man_t, adds *Vec_Int_t, fullOnly int32, bold *Vec_Int_t) *Vec_Int_t {
	marked := make([]bool, p.nObjs)
	if bold != nil {
		for _, id := range vec.Vec_IntArray(bold) {
			marked[id] = true
		}
	}
	mapping := vec.Vec_IntStartFull(int(p.nObjs))
	data := vec.Vec_IntArray(adds)
	for i := 0; 6*i < len(data); i++ {
		box := data[6*i : 6*i+6]
		if fullOnly != 0 && box[2] == 0 || marked[box[3]] || marked[box[4]] {
			continue
		}
		vec.Vec_IntWriteEntry(mapping, int(box[3]), int32(i))
		vec.Vec_IntWriteEntry(mapping, int(box[4]), int32(i))
	}
	return mapping
}

func Gia_ShowMapXors(p *Gia_Man_t, xors *Vec_Int_t) *Vec_Int_t {
	mapping := vec.Vec_IntStartFull(int(p.nObjs))
	data := vec.Vec_IntArray(xors)
	for i := 0; 4*i < len(data); i++ {
		vec.Vec_IntWriteEntry(mapping, int(data[4*i]), int32(i))
	}
	return mapping
}

func Gia_ShowCollectObjs_rec(p *Gia_Man_t, obj *Gia_Obj_t, adds, xors, mapAdds, mapXors, order *Vec_Int_t) int32 {
	if Gia_ObjIsTravIdCurrent(p, obj) != 0 {
		return Gia_ObjLevel(p, obj)
	}
	Gia_ObjSetTravIdCurrent(p, obj)
	if Gia_ObjIsCi(obj) != 0 {
		return 0
	}
	id := Gia_ObjId(p, obj)
	level := int32(0)
	visit := func(id int32) {
		level = max(level, Gia_ShowCollectObjs_rec(p, Gia_ManObj(p, id), adds, xors, mapAdds, mapXors, order))
	}
	if box := vec.Vec_IntEntry(mapAdds, int(id)); box >= 0 {
		data := vec.Vec_IntArray(adds)[6*int(box):]
		Gia_ObjSetTravIdCurrentId(p, data[3])
		Gia_ObjSetTravIdCurrentId(p, data[4])
		visit(data[0])
		visit(data[1])
		if data[2] != 0 {
			visit(data[2])
		}
		level++
		Gia_ObjSetLevelId(p, data[3], level)
		Gia_ObjSetLevelId(p, data[4], level)
		obj = Gia_ManObj(p, data[4])
	} else if box := vec.Vec_IntEntry(mapXors, int(id)); box >= 0 {
		data := vec.Vec_IntArray(xors)[4*int(box):]
		visit(data[1])
		visit(data[2])
		if data[3] != 0 {
			visit(data[3])
		}
		level++
		Gia_ObjSetLevel(p, obj, level)
	} else {
		if Gia_ObjIsMux(p, obj) != 0 {
			panic("gia: cannot show unexpanded mux")
		}
		level = 1 + max(Gia_ShowCollectObjs_rec(p, Gia_ObjFanin0(obj), adds, xors, mapAdds, mapXors, order), Gia_ShowCollectObjs_rec(p, Gia_ObjFanin1(obj), adds, xors, mapAdds, mapXors, order))
		Gia_ObjSetLevel(p, obj, level)
	}
	vec.Vec_IntPush(order, Gia_ObjId(p, obj))
	p.nLevels = max(p.nLevels, level)
	return level
}

func Gia_ShowCollectObjs(p *Gia_Man_t, adds, xors, mapAdds, mapXors *Vec_Int_t) *Vec_Int_t {
	order := vec.Vec_IntAlloc(100)
	Gia_ManCleanLevels(p, p.nObjs)
	p.nLevels = 0
	Gia_ManIncrementTravId(p)
	Gia_ObjSetTravIdCurrent(p, Gia_ManConst0(p))
	for _, id := range vec.Vec_IntArray(p.vCis) {
		Gia_ObjSetTravIdCurrentId(p, id)
	}
	for _, id := range vec.Vec_IntArray(p.vCos) {
		Gia_ShowCollectObjs_rec(p, Gia_ObjFanin0(Gia_ManObj(p, id)), adds, xors, mapAdds, mapXors, order)
	}
	return order
}
