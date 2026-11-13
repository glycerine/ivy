package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"time"
)

// cec4SimulationRound propagates input patterns, creating candidate classes on
// the first round and splitting existing candidates on subsequent rounds.
func cec4SimulationRound(sim *cec3Simulation, graph *gia.Gia_Man_t, maxLevel int32, initialized bool) {
	cec4SimulationRoundTimed(sim, graph, maxLevel, initialized)
}

func cec4SimulationRoundTimed(sim *cec3Simulation, graph *gia.Gia_Man_t, maxLevel int32, initialized bool) (int64, int64) {
	start := time.Now()
	cec4SimulateGraph(sim, graph)
	simTime := time.Since(start).Nanoseconds()
	start = time.Now()
	if !initialized {
		cec4CreateClasses(sim, graph, maxLevel)
		return simTime, time.Since(start).Nanoseconds()
	}
	// The original only schedules a class when one of its AND nodes differs
	// from its representative; standalone CI classes are left untouched.
	var heads []int32
	seen := make(map[int32]bool)
	for id := int32(1); id < gia.Gia_ManObjNum(graph); id++ {
		if gia.Gia_ObjIsAnd(gia.Gia_ManObj(graph, id)) == 0 {
			continue
		}
		repr := gia.Gia_ObjRepr(graph, id)
		if repr != gia.GIA_VOID && !seen[repr] && !sim.equivalent(repr, id) {
			seen[repr] = true
			heads = append(heads, repr)
		}
	}
	for _, head := range heads {
		cec4RefineClass(sim, graph, head)
	}
	return simTime, time.Since(start).Nanoseconds()
}
