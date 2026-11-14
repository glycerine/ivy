package gia

import "github.com/glycerine/abcgo/misc/vec"

func Emb_ManPerformBfs(p *Emb_Man_t, current, next *Vec_Int_t, distances []Emb_Dat_t) *Emb_Obj_t {
	if vec.Vec_IntSize(current) == 0 {
		panic("gia: empty embedding BFS frontier")
	}
	p.nDistMax = 0
	for vec.Vec_IntSize(current) > 0 {
		p.nReached += int32(vec.Vec_IntSize(current))
		vec.Vec_IntClear(next)
		for _, id := range vec.Vec_IntArray(current) {
			o := Emb_ManObj(p, id)
			if distances != nil {
				distances[o.Value] = float32(p.nDistMax)
			}
			for _, neighbors := range [][]*Emb_Obj_t{o.fanins, o.fanouts} {
				for _, neighbor := range neighbors {
					if Emb_ObjIsTravIdCurrent(p, neighbor) != 0 {
						continue
					}
					Emb_ObjSetTravIdCurrent(p, neighbor)
					vec.Vec_IntPush(next, neighbor.hHandle)
				}
			}
		}
		current, next = next, current
		p.nDistMax++
	}
	return Emb_ManObj(p, vec.Vec_IntEntry(next, 0))
}
func Emb_ManConnectedComponents(p *Emb_Man_t) *Vec_Int_t {
	result, current, next := vec.Vec_IntAlloc(16), vec.Vec_IntAlloc(16), vec.Vec_IntAlloc(16)
	p.nReached = 0
	Emb_ManIncrementTravId(p)
	for i := int32(0); i < Emb_ManCoNum(p); i++ {
		o := Emb_ManCo(p, i)
		if Emb_ObjIsTravIdCurrent(p, o) != 0 {
			continue
		}
		Emb_ObjSetTravIdCurrent(p, o)
		vec.Vec_IntPush(result, o.hHandle)
		vec.Vec_IntClear(current)
		vec.Vec_IntPush(current, o.hHandle)
		Emb_ManPerformBfs(p, current, next, nil)
	}
	return result
}
func Emb_ManFindDistances(p *Emb_Man_t, starts *Vec_Int_t, distances []Emb_Dat_t) *Emb_Obj_t {
	p.nReached = 0
	p.nDistMax = 0
	Emb_ManIncrementTravId(p)
	current, next := vec.Vec_IntAlloc(vec.Vec_IntSize(starts)), vec.Vec_IntAlloc(16)
	for _, id := range vec.Vec_IntArray(starts) {
		o := Emb_ManObj(p, id)
		if Emb_ObjIsTravIdCurrent(p, o) != 0 {
			continue
		}
		Emb_ObjSetTravIdCurrent(p, o)
		vec.Vec_IntPush(current, id)
	}
	return Emb_ManPerformBfs(p, current, next, distances)
}
func Emb_ManRandomVertex(p *Emb_Man_t) *Emb_Obj_t {
	for {
		id := int32((911 * Gia_ManRandom(0)) % uint32(Gia_ManObjNum(p.pGia)))
		handle := Gia_ObjValue(Gia_ManObj(p.pGia, id))
		if handle >= 0 {
			o := Emb_ManObj(p, handle)
			if Emb_ObjIsNode(o) != 0 {
				return o
			}
		}
	}
}
