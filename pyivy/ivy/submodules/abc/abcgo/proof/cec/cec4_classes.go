package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
)

// cec4CreateClasses implements the initial hash grouping and exact simulation
// refinement performed by Cec4_RefineInit and Cec4_RefineClasses.
func cec4CreateClasses(s *cec3Simulation, graph *gia.Gia_Man_t, maxLevel int32) {
	count := gia.Gia_ManObjNum(graph)
	graph.StartRepresentatives()
	for id := int32(0); id < count; id++ {
		gia.Gia_ObjSetNext(graph, id, -1)
	}
	bins := make([]int32, util.Abc_PrimeCudd(uint32(count)))
	for i := range bins {
		bins[i] = -1
	}
	for id := count - 1; id >= 0; id-- {
		obj := gia.Gia_ManObj(graph, id)
		if gia.Gia_ObjIsCo(obj) != 0 || maxLevel != 0 && gia.Gia_ObjLevel(graph, obj) > maxLevel {
			continue
		}
		key := Cec4_ManSimHashKey(s.object(id), int32(s.words), int32(len(bins)))
		gia.Gia_ObjSetNext(graph, id, bins[key])
		bins[key] = id
	}
	for _, head := range bins {
		if head < 0 {
			continue
		}
		for id := gia.Gia_ObjNext(graph, head); id > 0; id = gia.Gia_ObjNext(graph, id) {
			gia.Gia_ObjSetRepr(graph, id, head)
		}
		if gia.Gia_ObjNext(graph, head) > 0 {
			cec4RefineClass(s, graph, head)
		}
	}
}
