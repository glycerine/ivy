package gia

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
)

// Gia_ManOrderPios follows the GIA object order for corresponding AIG ports.
func Gia_ManOrderPios(graph *aig.Graph, order *Gia_Man_t) *Vec_Ptr_t {
	if aig.Aig_ManCiNum(graph) != Gia_ManCiNum(order) || aig.Aig_ManCoNum(graph) != Gia_ManCoNum(order) {
		panic("gia: incompatible AIG port counts")
	}
	result := vec.Vec_PtrAlloc(int(Gia_ManCiNum(order) + Gia_ManCoNum(order)))
	for id := int32(0); id < order.nObjs; id++ {
		obj := Gia_ManObj(order, id)
		if Gia_ObjIsCi(obj) != 0 {
			vec.Vec_PtrPush(result, aig.Aig_ManCi(graph, Gia_ObjCioId(obj)))
		} else if Gia_ObjIsCo(obj) != 0 {
			vec.Vec_PtrPush(result, aig.Aig_ManCo(graph, Gia_ObjCioId(obj)))
		}
	}
	return result
}
