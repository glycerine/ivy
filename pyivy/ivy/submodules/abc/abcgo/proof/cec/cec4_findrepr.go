package cec

import "github.com/glycerine/abcgo/aig/gia"

// cec4FindRepr chooses the earliest live class member whose current simulation
// matches the candidate up to phase, as in Cec4_ManFindRepr. The caller must
// propagate the current input patterns before invoking it.
func cec4FindRepr(sim *cec3Simulation, graph *gia.Gia_Man_t, id int32) int32 {
	repr := gia.Gia_ObjRepr(graph, id)
	if repr == gia.GIA_VOID || gia.Gia_ObjProved(graph, id) != 0 || gia.Gia_ObjProved(graph, repr) != 0 {
		panic("cec4: invalid representative lookup")
	}
	if sim.equivalent(id, repr) {
		return repr
	}
	for member := gia.Gia_ObjNext(graph, repr); member > 0 && member != id; member = gia.Gia_ObjNext(graph, member) {
		if gia.Gia_ObjProved(graph, member) == 0 && gia.Gia_ObjFailed(graph, member) == 0 && sim.equivalent(id, member) {
			return member
		}
	}
	return -1
}
