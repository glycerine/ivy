package cec

import "github.com/glycerine/abcgo/aig/gia"

// Cec4_ManStartNew initializes the reduced graph with corresponding inputs.
func Cec4_ManStartNew(graph *Gia_Man_t) *Gia_Man_t {
	out := gia.Gia_ManStart(gia.Gia_ManObjNum(graph))
	out.CopyIdentityFrom(graph)
	if gia.Gia_ManMuxNum(graph) > 0 {
		out.EnableMuxes()
	}
	gia.Gia_ManFillValue(graph)
	gia.Gia_ObjSetValue(gia.Gia_ManConst0(graph), 0)
	for i := int32(0); i < gia.Gia_ManCiNum(graph); i++ {
		obj := gia.Gia_ManCi(graph, i)
		gia.Gia_ObjSetValue(obj, gia.Gia_ManAppendCi(out))
	}
	gia.Gia_ManHashAlloc(out)
	gia.Gia_ManSetRegNum(out, gia.Gia_ManRegNum(graph))
	return out
}
