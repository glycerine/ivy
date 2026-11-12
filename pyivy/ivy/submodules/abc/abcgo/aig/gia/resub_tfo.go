package gia

import "github.com/glycerine/abcgo/misc/vec"

func Gia_ManUnivTfo_rec(p *Gia_Man_t, id int32, nodes, outputs *Vec_Int_t) int32 {
	if Gia_ObjIsTravIdCurrentId(p, id) != 0 {
		return 0
	}
	Gia_ObjSetTravIdCurrentId(p, id)
	if Gia_ObjIsCo(Gia_ManObj(p, id)) != 0 {
		if nodes != nil {
			vec.Vec_IntPush(nodes, id)
		}
		if outputs != nil {
			vec.Vec_IntPush(outputs, id)
		}
	}
	count := int32(1)
	for i := int32(0); i < Gia_ObjFanoutNumId(p, id); i++ {
		count += Gia_ManUnivTfo_rec(p, Gia_ObjFanoutId(p, id, i), nodes, outputs)
	}
	return count
}
func Gia_ManUnivTfo(p *Gia_Man_t, objects []int32, nodes, outputs **Vec_Int_t) int32 {
	if nodes != nil {
		if *nodes == nil {
			*nodes = vec.Vec_IntAlloc(100)
		} else {
			vec.Vec_IntClear(*nodes)
		}
	}
	if outputs != nil {
		if *outputs == nil {
			*outputs = vec.Vec_IntAlloc(100)
		} else {
			vec.Vec_IntClear(*outputs)
		}
	}
	var ns, os *Vec_Int_t
	if nodes != nil {
		ns = *nodes
	}
	if outputs != nil {
		os = *outputs
	}
	Gia_ManIncrementTravId(p)
	count := int32(0)
	for _, id := range objects {
		if id < 0 || id >= p.nObjs {
			panic("gia: invalid TFO root")
		}
		count += Gia_ManUnivTfo_rec(p, id, ns, os)
	}
	if ns != nil {
		vec.Vec_IntSort(ns, 0)
	}
	if os != nil {
		vec.Vec_IntSort(os, 0)
	}
	return count
}
