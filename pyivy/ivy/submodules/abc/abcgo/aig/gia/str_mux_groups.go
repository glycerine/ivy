package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"strings"
)

func Str_MuxInputsCollect_rec(p *Gia_Man_t, obj *Gia_Obj_t, nodes *Vec_Int_t) {
	if obj.fMark0 == 0 {
		vec.Vec_IntPush(nodes, obj.id)
		return
	}
	vec.Vec_IntPush(nodes, Gia_ObjFaninId2(p, obj.id))
	Str_MuxInputsCollect_rec(p, Gia_ObjFanin0(obj), nodes)
	Str_MuxInputsCollect_rec(p, Gia_ObjFanin1(obj), nodes)
}
func Str_MuxInputsCollect(p *Gia_Man_t, obj *Gia_Obj_t, nodes *Vec_Int_t) {
	if obj.fMark0 != 0 {
		panic("gia: mux root already marked")
	}
	obj.fMark0 = 1
	defer func() { obj.fMark0 = 0 }()
	vec.Vec_IntClear(nodes)
	Str_MuxInputsCollect_rec(p, obj, nodes)
}
func Str_MuxStructCollect_rec(p *Gia_Man_t, obj *Gia_Obj_t, nodes *Vec_Int_t) {
	if obj.fMark0 == 0 {
		return
	}
	Str_MuxStructCollect_rec(p, Gia_ObjFanin0(obj), nodes)
	Str_MuxStructCollect_rec(p, Gia_ObjFanin1(obj), nodes)
	vec.Vec_IntPush(nodes, obj.id)
}
func Str_MuxStructCollect(p *Gia_Man_t, obj *Gia_Obj_t, nodes *Vec_Int_t) {
	if obj.fMark0 != 0 {
		panic("gia: mux root already marked")
	}
	obj.fMark0 = 1
	defer func() { obj.fMark0 = 0 }()
	vec.Vec_IntClear(nodes)
	Str_MuxStructCollect_rec(p, obj, nodes)
}
func Str_MuxStructDump_rec(p *Gia_Man_t, obj *Gia_Obj_t, out *Vec_Str_t) {
	if obj.fMark0 == 0 {
		return
	}
	vec.Vec_StrPush(out, '[')
	vec.Vec_StrPush(out, '(')
	vec.Vec_StrPrintNum(out, Gia_ObjFaninId2(p, obj.id))
	vec.Vec_StrPush(out, ')')
	left, right := Gia_ObjFanin1(obj), Gia_ObjFanin0(obj)
	if Gia_ObjFaninC2(p, obj) != 0 {
		left, right = right, left
	}
	Str_MuxStructDump_rec(p, left, out)
	vec.Vec_StrPush(out, '|')
	Str_MuxStructDump_rec(p, right, out)
	vec.Vec_StrPush(out, ']')
}
func Str_MuxStructDump(p *Gia_Man_t, obj *Gia_Obj_t, out *Vec_Str_t) {
	if obj.fMark0 != 0 {
		panic("gia: mux root already marked")
	}
	obj.fMark0 = 1
	defer func() { obj.fMark0 = 0 }()
	vec.Vec_StrClear(out)
	Str_MuxStructDump_rec(p, obj, out)
	vec.Vec_StrPush(out, 0)
}
func Str_ManMuxCountOne(s string) int32 { return int32(strings.Count(s, "[")) }
func Str_ManDeriveTrees(p *Gia_Man_t) *Vec_Wec_t {
	if p.PMuxes == nil {
		panic("gia: mux representation required")
	}
	Gia_ManCreateRefs(p)
	for _, obj := range p.pObjs {
		if Gia_ObjIsMux(p, obj) == 0 {
			continue
		}
		for _, fan := range []*Gia_Obj_t{Gia_ObjFanin0(obj), Gia_ObjFanin1(obj)} {
			if Gia_ObjIsMux(p, fan) != 0 && Gia_ObjRefNum(p, fan) == 1 {
				fan.fMark0 = 1
			}
		}
	}
	groups := vec.Vec_WecAlloc(1000)
	vec.Vec_WecPushLevel(groups)
	names := make(map[string]int)
	text := vec.Vec_StrAlloc(1000)
	for _, obj := range p.pObjs {
		if Gia_ObjIsMux(p, obj) == 0 || obj.fMark0 != 0 || Gia_ObjFanin0(obj).fMark0 == 0 && Gia_ObjFanin1(obj).fMark0 == 0 {
			continue
		}
		Str_MuxStructDump(p, obj, text)
		key := string(vec.Vec_StrArray(text))
		index, ok := names[key]
		if !ok {
			index = vec.Vec_WecSize(groups)
			names[key] = index
			vec.Vec_WecPushLevel(groups)
		}
		vec.Vec_IntPush(vec.Vec_WecEntry(groups, index), obj.id)
	}
	return groups
}
func Str_ManCreateRoots(groups *Vec_Wec_t, objects int32) *Vec_Int_t {
	roots := vec.Vec_IntStartFull(int(objects))
	for i := 0; i < vec.Vec_WecSize(groups); i++ {
		for _, id := range vec.Vec_IntArray(vec.Vec_WecEntry(groups, i)) {
			vec.Vec_IntWriteEntry(roots, int(id), int32(i))
		}
	}
	return roots
}
func Str_MuxTraverse_rec(p *Gia_Man_t, id int32) {
	if Gia_ObjIsTravIdCurrentId(p, id) != 0 {
		return
	}
	Gia_ObjSetTravIdCurrentId(p, id)
	obj := Gia_ManObj(p, id)
	if Gia_ObjIsAnd(obj) == 0 {
		return
	}
	Str_MuxTraverse_rec(p, Gia_ObjFaninId0(obj, id))
	Str_MuxTraverse_rec(p, Gia_ObjFaninId1(obj, id))
	if Gia_ObjIsMux(p, obj) != 0 {
		Str_MuxTraverse_rec(p, Gia_ObjFaninId2(p, id))
	}
}
func Str_ManCheckOverlap(p *Gia_Man_t, groups *Vec_Wec_t) {
	for i := 0; i < vec.Vec_WecSize(groups); i++ {
		for k := 0; k < vec.Vec_IntSize(vec.Vec_WecEntry(groups, i)); k++ {
			group := vec.Vec_WecEntry(groups, i)
			if vec.Vec_IntSize(group) == 1 {
				continue
			}
			id := vec.Vec_IntEntry(group, k)
			Gia_ManIncrementTravId(p)
			Str_MuxTraverse_rec(p, id)
			overlap := false
			for _, other := range vec.Vec_IntArray(group) {
				if other != id && Gia_ObjIsTravIdCurrentId(p, other) != 0 {
					overlap = true
					break
				}
			}
			if !overlap {
				continue
			}
			rest := append([]int32(nil), vec.Vec_IntArray(group)[1:]...)
			vec.Vec_IntShrink(group, 1)
			for _, other := range rest {
				vec.Vec_IntPush(vec.Vec_WecPushLevel(groups), other)
			}
		}
	}
}
