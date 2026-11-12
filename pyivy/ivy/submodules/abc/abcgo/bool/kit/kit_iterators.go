package kit

import "iter"

func Kit_DsdNtkForEachObj(ntk *Kit_DsdNtk_t) iter.Seq2[int32, *Kit_DsdObj_t] {
	return func(yield func(int32, *Kit_DsdObj_t) bool) {
		for i, obj := range ntk.pNodes {
			if !yield(int32(i), obj) {
				return
			}
		}
	}
}
func Kit_DsdObjForEachFanin(obj *Kit_DsdObj_t) iter.Seq2[int32, uint16] {
	return func(yield func(int32, uint16) bool) {
		for i, lit := range obj.pFans[:obj.nFans] {
			if !yield(int32(i), lit) {
				return
			}
		}
	}
}
func Kit_GraphForEachNode(g *Kit_Graph_t) iter.Seq2[int32, *Kit_Node_t] {
	return func(yield func(int32, *Kit_Node_t) bool) {
		for i := g.nLeaves; i < int32(len(g.pNodes)); i++ {
			if !yield(i, g.pNodes[i]) {
				return
			}
		}
	}
}
func Kit_GraphForEachLeaf(g *Kit_Graph_t) iter.Seq2[int32, *Kit_Node_t] {
	return func(yield func(int32, *Kit_Node_t) bool) {
		for i := int32(0); i < g.nLeaves; i++ {
			if !yield(i, g.pNodes[i]) {
				return
			}
		}
	}
}
func Kit_SopForEachCube(sop *Kit_Sop_t) iter.Seq2[int32, uint32] {
	return func(yield func(int32, uint32) bool) {
		for i, cube := range sop.pCubes[:sop.cubeNum()] {
			if !yield(int32(i), cube) {
				return
			}
		}
	}
}
