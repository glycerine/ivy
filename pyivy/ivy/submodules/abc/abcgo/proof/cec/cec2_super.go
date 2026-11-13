package cec

import "github.com/glycerine/abcgo/aig/gia"

// cec2CollectSuper translates Cec2_CollectSuper using GIA literal IDs instead
// of complemented object pointers. The result has unique literals in DFS
// order; complemented edges, shared nodes, CIs and marked muxes stop descent.
func cec2CollectSuper(graph *gia.Gia_Man_t, root int32, useMuxes bool) []int32 {
	if graph == nil || root&1 != 0 || root <= 0 {
		panic("cec: invalid supergate root")
	}
	seen := make(map[int32]bool)
	var leaves []int32
	var visit func(int32, bool)
	visit = func(lit int32, first bool) {
		id := lit >> 1
		obj := gia.Gia_ManObj(graph, id)
		if obj == nil {
			panic("cec: invalid supergate object")
		}
		refs := gia.Gia_ObjValue(obj)
		if graph.Refs != nil {
			refs = gia.Gia_ObjRefNumId(graph, id)
		}
		if lit&1 != 0 || gia.Gia_ObjIsCi(obj) != 0 || !first && refs > 1 || useMuxes && gia.Gia_ObjMark0(obj) != 0 {
			if !seen[lit] {
				seen[lit] = true
				leaves = append(leaves, lit)
			}
			return
		}
		if gia.Gia_ObjIsAnd(obj) == 0 {
			panic("cec: supergate contains a non-AND node")
		}
		visit(gia.Gia_ObjFaninLit0(obj, id), false)
		visit(gia.Gia_ObjFaninLit1(obj, id), false)
	}
	visit(root, true)
	return leaves
}
