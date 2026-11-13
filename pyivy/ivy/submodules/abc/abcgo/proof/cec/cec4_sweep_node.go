package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/sat/glucose2"
)

// cec4SweepNode compares a candidate with its representative on the current
// reduced graph. A SAT result replays the CI model into the next simulation
// pattern; an UNSAT result marks the candidate proved and substitutes its copy.
func cec4SweepNode(graph, reduced *gia.Gia_Man_t, sim *cec3Simulation, cnf *cec4CNF, copies []int32, id, repr, conflictLimit int32, pattern *int) int32 {
	obj, representative := gia.Gia_ManObj(graph, id), gia.Gia_ManObj(graph, repr)
	phase := gia.Gia_ObjPhase(obj) ^ gia.Gia_ObjPhase(representative)
	relation := (copies[id]^copies[repr])&1 ^ phase
	same := copies[repr]>>1 == copies[id]>>1
	var status int32
	if same {
		status = glucose2.GLUCOSE_UNSAT
		if relation != 0 {
			panic("cec4: inconsistent phase for identical reduced objects")
		}
	} else {
		status = cnf.solveTwo(copies[repr]>>1, copies[id]>>1, relation, conflictLimit)
	}
	cec4ApplySweepResult(graph, reduced, sim, cnf, copies, id, repr, pattern, status)
	return status
}

func cec4ApplySweepResult(graph, reduced *gia.Gia_Man_t, sim *cec3Simulation, cnf *cec4CNF, copies []int32, id, repr int32, pattern *int, status int32) {
	obj, representative := gia.Gia_ManObj(graph, id), gia.Gia_ManObj(graph, repr)
	phase := gia.Gia_ObjPhase(obj) ^ gia.Gia_ObjPhase(representative)
	switch status {
	case glucose2.GLUCOSE_SAT:
		*pattern++
		if *pattern >= 64*sim.words {
			panic("cec4: simulation pattern storage is full")
		}
		if cnf.jType == 0 {
			// ABC replays reduced-graph CI IDs directly on the original graph.
			// With interleaved CIs this can write a different original object.
			for variable, object := range cnf.objects {
				if mapped, ok := cnf.inputs[object]; ok && int(mapped) == variable {
					sim.setInputBit(object, *pattern, cnf.solver.Model[variable] == glucose2.LTrue)
				}
			}
		} else {
			// The source replays only the complemented justification trail.
			for _, lit := range cnf.solver.JustModel[1:] {
				v := int(lit >> 1)
				if v < len(cnf.objects) && cnf.objects[v] >= 0 {
					sim.setInputBit(cnf.objects[v], *pattern, lit&1 == 0)
				}
			}
		}
	case glucose2.GLUCOSE_UNSAT:
		gia.Gia_ObjSetProved(graph, id)
		copies[id] = copies[repr] ^ phase
		gia.Gia_ObjSetValue(obj, copies[id])
	case glucose2.GLUCOSE_UNDEC:
		gia.Gia_ObjSetFailed(graph, id)
	}
}
