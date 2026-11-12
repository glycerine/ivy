package gia

import "github.com/glycerine/abcgo/misc/vec"

func Gia_ObjFaninId(obj *Gia_Obj_t, id, index int32) int32 {
	if index == 0 {
		return Gia_ObjFaninId0(obj, id)
	}
	return Gia_ObjFaninId1(obj, id)
}
func Gia_ObjTravIsTopTwo(p *Gia_Man_t, id int32) int32 {
	return giaBool(p.pTravIds[id] >= p.nTravIds-1)
}
func Gia_ObjTravIsSame(p *Gia_Man_t, a, b int32) int32 {
	return giaBool(p.pTravIds[a] == p.pTravIds[b])
}
func Gia_ObjTravSetSame(p *Gia_Man_t, a, b int32) { p.pTravIds[a] = p.pTravIds[b] }
func Gia_RsbWindowGather(p *Gia_Man_t, paths *Vec_Int_t, node int32, visited *Vec_Int_t) {
	for node != 0 {
		vec.Vec_IntPush(visited, node)
		prev := vec.Vec_IntEntry(paths, int(node))
		if prev != 0 && Gia_ObjTravIsSame(p, prev, node) == 0 {
			panic("gia: inconsistent Rsb2 path")
		}
		node = prev
	}
}
func Gia_RsbWindowExplore(p *Gia_Man_t, visited *Vec_Int_t, start int32, paths *Vec_Int_t, meet, node *int32) int32 {
	*meet, *node = 0, 0
	end := int32(vec.Vec_IntSize(visited))
	for i := start; i < end; i++ {
		id := vec.Vec_IntEntry(visited, int(i))
		obj := Gia_ManObj(p, id)
		if Gia_ObjIsAnd(obj) == 0 {
			continue
		}
		for n := int32(0); n < 2; n++ {
			fan := Gia_ObjFaninId(obj, id, n)
			if Gia_ObjTravIsTopTwo(p, id) != 0 && Gia_ObjTravIsTopTwo(p, fan) != 0 && Gia_ObjTravIsSame(p, id, fan) == 0 {
				*meet, *node = fan, id
				return 1
			}
			if Gia_ObjTravIsTopTwo(p, fan) != 0 {
				if Gia_ObjTravIsSame(p, id, fan) == 0 {
					panic("gia: inconsistent Rsb2 traversal")
				}
				continue
			}
			Gia_ObjTravSetSame(p, fan, id)
			vec.Vec_IntWriteEntry(paths, int(fan), id)
			vec.Vec_IntPush(visited, fan)
		}
	}
	return 0
}
func Gia_RsbWindowInit(p *Gia_Man_t, paths *Vec_Int_t, pivot, iterations int32) *Vec_Int_t {
	obj := Gia_ManObj(p, pivot)
	if Gia_ObjIsAnd(obj) == 0 || vec.Vec_IntSize(paths) < int(p.nObjs) {
		panic("gia: invalid Rsb2 pivot")
	}
	visited := vec.Vec_IntAlloc(100)
	for n := int32(0); n < 2; n++ {
		fan := Gia_ObjFaninId(obj, pivot, n)
		Gia_ManIncrementTravId(p)
		vec.Vec_IntPush(visited, fan)
		vec.Vec_IntWriteEntry(paths, int(fan), 0)
		Gia_ObjSetTravIdCurrentId(p, fan)
	}
	start := int32(0)
	for i := int32(0); i < iterations; i++ {
		next := int32(vec.Vec_IntSize(visited))
		var meet, node int32
		if Gia_RsbWindowExplore(p, visited, start, paths, &meet, &node) != 0 {
			vec.Vec_IntClear(visited)
			Gia_RsbWindowGather(p, paths, vec.Vec_IntEntry(paths, int(meet)), visited)
			Gia_RsbWindowGather(p, paths, node, visited)
			vec.Vec_IntPush(visited, pivot)
			return visited
		}
		start = next
	}
	return nil
}
func Gia_RsbCreateWindowInputs(p *Gia_Man_t, window *Vec_Int_t) *Vec_Int_t {
	inputs := vec.Vec_IntAlloc(10)
	Gia_ManIncrementTravId(p)
	for _, id := range vec.Vec_IntArray(window) {
		Gia_ObjSetTravIdCurrentId(p, id)
	}
	initial := vec.Vec_IntSize(window)
	for i := 0; i < initial; i++ {
		id := vec.Vec_IntEntry(window, i)
		obj := Gia_ManObj(p, id)
		if Gia_ObjIsAnd(obj) == 0 {
			panic("gia: invalid Rsb2 initial window")
		}
		for n := int32(0); n < 2; n++ {
			fan := Gia_ObjFaninId(obj, id, n)
			if Gia_ObjIsTravIdCurrentId(p, fan) == 0 {
				vec.Vec_IntPushUnique(inputs, fan)
			}
		}
	}
	for _, id := range vec.Vec_IntArray(inputs) {
		Gia_ObjSetTravIdCurrentId(p, id)
		vec.Vec_IntPush(window, id)
	}
	return inputs
}
