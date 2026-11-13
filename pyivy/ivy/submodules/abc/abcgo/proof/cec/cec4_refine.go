package cec

import "github.com/glycerine/abcgo/aig/gia"

// cec4RefineClass applies Cec4_RefineOneClassIter, repeatedly splitting a
// linked candidate class until each class has one phase-normalized signature.
func cec4RefineClass(sim *cec3Simulation, graph *gia.Gia_Man_t, head int32) {
	for {
		second := gia.Gia_ObjNext(graph, head)
		for second > 0 && sim.equivalent(head, second) {
			second = gia.Gia_ObjNext(graph, second)
		}
		if second <= 0 {
			return
		}
		sim.refineClass(graph, head)
		if gia.Gia_ObjNext(graph, second) <= 0 {
			return
		}
		head = second
	}
}
