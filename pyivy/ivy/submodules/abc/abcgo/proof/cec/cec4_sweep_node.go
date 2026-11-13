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
			status = glucose2.GLUCOSE_SAT
		}
	} else {
		status = cnf.solveTwo(copies[repr]>>1, copies[id]>>1, relation, conflictLimit)
	}
	switch status {
	case glucose2.GLUCOSE_SAT:
		*pattern++
		if *pattern >= 64*sim.words {
			panic("cec4: simulation pattern storage is full")
		}
		for i := int32(0); !same && i < gia.Gia_ManCiNum(graph); i++ {
			oldID := gia.Gia_ObjId(graph, gia.Gia_ManCi(graph, i))
			newID := gia.Gia_ObjId(reduced, gia.Gia_ManCi(reduced, i))
			if v, ok := cnf.inputs[newID]; ok {
				sim.setInputBit(oldID, *pattern, cnf.solver.Model[v] == glucose2.LTrue)
			}
		}
	case glucose2.GLUCOSE_UNSAT:
		gia.Gia_ObjSetProved(graph, id)
		copies[id] = copies[repr] ^ phase
		gia.Gia_ObjSetValue(obj, copies[id])
	case glucose2.GLUCOSE_UNDEC:
		gia.Gia_ObjSetFailed(graph, id)
	}
	return status
}
