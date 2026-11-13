package cec

import "github.com/glycerine/abcgo/aig/gia"

// cec4ResimulateNode updates the transitive fanin of one candidate at most
// once per packed-pattern generation, following Cec4_ManSimulate_rec.
func cec4ResimulateNode(sim *cec3Simulation, graph *gia.Gia_Man_t, id, stamp int32, stamps []int32) {
	if id == 0 || stamps[id] == stamp {
		return
	}
	stamps[id] = stamp
	a, b, c, kind := graph.Gate(id)
	switch kind {
	case 1:
		return
	case 2, 3:
		cec4ResimulateNode(sim, graph, a>>1, stamp, stamps)
		cec4ResimulateNode(sim, graph, b>>1, stamp, stamps)
	case 4:
		cec4ResimulateNode(sim, graph, a>>1, stamp, stamps)
		cec4ResimulateNode(sim, graph, b>>1, stamp, stamps)
		cec4ResimulateNode(sim, graph, c>>1, stamp, stamps)
	default:
		panic("cec4: unsupported local resimulation node")
	}
	cec4SimulateNode(sim, graph, id)
}
