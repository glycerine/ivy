package cec

import "github.com/glycerine/abcgo/aig/gia"

// cec4SimulationRound propagates input patterns, creating candidate classes on
// the first round and splitting existing candidates on subsequent rounds.
func cec4SimulationRound(sim *cec3Simulation, graph *gia.Gia_Man_t, maxLevel int32, initialized bool) {
	cec4SimulateGraph(sim, graph)
	if !initialized {
		cec4CreateClasses(sim, graph, maxLevel)
		return
	}
	// Save the original heads: refinement may create new heads during the pass.
	var heads []int32
	for id := int32(0); id < gia.Gia_ManObjNum(graph); id++ {
		if gia.Gia_ObjIsHead(graph, id) != 0 {
			heads = append(heads, id)
		}
	}
	for _, head := range heads {
		cec4RefineClass(sim, graph, head)
	}
}
