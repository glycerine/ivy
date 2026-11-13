package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/glucose2"
	"time"
)

type cec4SweepStats struct{ sat, unsat, unknown, recycles int }

// cec4Sweep performs Cec4's simulation-guided, single-pass SAT sweep.
// Witnesses are returned to the caller for storage in the source manager.
func cec4Sweep(graph *gia.Gia_Man_t, pars *Cec_ParFra_t, simOnly, finishGraph bool, managers ...*Cec4_Man_t) (*gia.Gia_Man_t, *util.Abc_Cex_t, cec4SweepStats) {
	var stats cec4SweepStats
	var manager *Cec4_Man_t
	if len(managers) > 0 {
		manager = managers[0]
	}
	pack, witness := cec4PrepareSimulation(graph, pars, simOnly, manager)
	if witness != nil || simOnly {
		return nil, witness, stats
	}
	if manager != nil {
		manager.nSatSat = 0
	}
	sim := pack.values
	clear(pack.reserved.data)
	if manager != nil {
		manager.packer = pack
		manager.simState = &cec4SimulationState{sim: sim, initialized: true}
	}
	reduced := gia.Gia_ManStart(gia.Gia_ManObjNum(graph))
	if manager != nil {
		manager.pNew = reduced
	}
	reduced.CopyIdentityFrom(graph)
	gia.Gia_ManFillValue(graph)
	gia.Gia_ObjSetValue(gia.Gia_ManConst0(graph), 0)
	copies := make([]int32, gia.Gia_ManObjNum(graph))
	for i := range copies {
		copies[i] = -1
	}
	copies[0] = 0
	for i := int32(0); i < gia.Gia_ManCiNum(graph); i++ {
		obj := gia.Gia_ManCi(graph, i)
		id := gia.Gia_ObjId(graph, obj)
		copies[id] = gia.Gia_ManAppendCi(reduced)
		gia.Gia_ObjSetValue(obj, copies[id])
	}
	gia.Gia_ManSetRegNum(reduced, gia.Gia_ManRegNum(graph))
	gia.Gia_ManHashAlloc(reduced)
	cnf := newCec4CNF(reduced, pars.jType)
	if manager != nil {
		cnf.solver = manager.pSat
		manager.cnf = cnf
	}
	var stamps []int32
	if manager != nil {
		vec.Vec_IntFill(manager.vCexStamps, graph.ObjectCount(), 0)
		stamps = vec.Vec_IntArray(manager.vCexStamps)
	} else {
		stamps = make([]int32, graph.ObjectCount())
	}
	pattern, calls := 0, 0
	failed := make(map[int32]bool)
	for id := int32(1); id < gia.Gia_ManObjNum(graph); id++ {
		a, b, c, kind := graph.Gate(id)
		if kind == 5 {
			copies[id] = copies[a>>1] ^ (a & 1)
			gia.Gia_ObjSetValue(gia.Gia_ManObj(graph, id), copies[id])
			continue
		}
		if kind != 2 && kind != 3 && kind != 4 {
			continue
		}
		if manager != nil {
			manager.nAndNodes++
		}
		litA, litB := copies[a>>1]^(a&1), copies[b>>1]^(b&1)
		var copyLit int32
		switch kind {
		case 2:
			copyLit = gia.Gia_ManHashAnd(reduced, litA, litB)
		case 3:
			copyLit = gia.Gia_ManHashXorReal(reduced, litA, litB)
		case 4:
			copyLit = gia.Gia_ManHashMux(reduced, litA, litB, copies[c>>1]^(c&1))
		}
		copies[id] = copyLit
		gia.Gia_ObjSetValue(gia.Gia_ManObj(graph, id), copyLit)
		if pars.nLevelMax != 0 && gia.Gia_ObjLevelId(graph, id) > pars.nLevelMax {
			continue
		}
		repr := gia.Gia_ObjRepr(graph, id)
		if repr == gia.GIA_VOID {
			continue
		}
		localStart := time.Now()
		repr = cec4FindReprAtPattern(sim, graph, id, int32(pattern), stamps)
		if manager != nil {
			manager.timeResimLoc += time.Since(localStart).Nanoseconds()
		}
		if repr < 0 {
			continue
		}
		rel := gia.Gia_ObjPhase(gia.Gia_ManObj(graph, id)) ^ gia.Gia_ObjPhase(gia.Gia_ManObj(graph, repr))
		if copies[id]>>1 == copies[repr]>>1 {
			if (copies[id] ^ copies[repr]) != rel {
				panic("cec4: structurally equal nodes have inconsistent phase")
			}
			gia.Gia_ObjSetProved(graph, id)
			continue
		}
		calls++
		if manager != nil {
			manager.nCallsSince = int32(calls)
		}
		if pars.nSatVarMax != 0 && calls > int(pars.nCallsRecycle) && len(cnf.ids) > int(pars.nSatVarMax) {
			cnf.recycle()
			if manager != nil {
				manager.pSat = cnf.solver
				manager.nRecycles++
				manager.nCallsSince = 0
			}
			calls = 0
			stats.recycles++
		}
		limit := pars.nBTLimit
		if failed[copies[id]>>1] || failed[copies[repr]>>1] {
			limit = max(int32(1), limit/10)
		}
		satStart := time.Now()
		cnfBefore := cnf.cnfTime
		gatesBefore := cnf.gates
		status := cec4SweepNode(graph, reduced, sim, cnf, copies, id, repr, limit, &pattern)
		elapsed := time.Since(satStart).Nanoseconds()
		if manager != nil {
			manager.pattern = pattern
			manager.timeCnf += cnf.cnfTime - cnfBefore
			manager.nGates[0] += cnf.gates[0] - gatesBefore[0]
			manager.nGates[1] += cnf.gates[1] - gatesBefore[1]
			if pars.fVerbose != 0 && status != glucose2.GLUCOSE_UNDEC {
				bucket := 0
				if status == glucose2.GLUCOSE_UNSAT {
					bucket = 1
				}
				if cnf.lastConflicts == 0 {
					manager.nConflicts[bucket][0]++
				}
				manager.nConflicts[bucket][1] += int32(cnf.lastConflicts)
				manager.nConflicts[bucket][2] = max(manager.nConflicts[bucket][2], int32(cnf.lastMaxConflicts))
			}
		}
		switch status {
		case glucose2.GLUCOSE_SAT:
			stats.sat++
			if manager != nil {
				manager.nSatSat++
				manager.nPatterns++
				if pars.fVerbose != 0 && cnf.lastConflicts == 0 {
					manager.timeSatSat0 += elapsed
				} else {
					manager.timeSatSat += elapsed
				}
			}
			if pattern == 64*sim.words-2 {
				resimStart := time.Now()
				simTime, refTime := cec4SimulationRoundTimed(sim, graph, pars.nLevelMax, true)
				if manager != nil {
					manager.nSimulates++
					manager.timeSim += simTime
					manager.timeRefine += refTime
					manager.timeResimGlo += time.Since(resimStart).Nanoseconds()
					Cec4_ManPrintStats(graph, pars, manager, 0)
					manager.pattern = 0
				}
				pattern = 0
				clear(stamps)
				clear(pack.reserved.data)
			}
		case glucose2.GLUCOSE_UNSAT:
			stats.unsat++
			if manager != nil {
				manager.nSatUnsat++
				if pars.fVerbose != 0 && cnf.lastConflicts == 0 {
					manager.timeSatUnsat0 += elapsed
				} else {
					manager.timeSatUnsat += elapsed
				}
			}
		default:
			stats.unknown++
			if manager != nil {
				manager.nSatUndec++
				manager.timeSatUndec += elapsed
			}
			failed[copies[id]>>1] = true
		}
	}
	if pattern > 0 {
		resimStart := time.Now()
		simTime, refTime := cec4SimulationRoundTimed(sim, graph, pars.nLevelMax, true)
		if manager != nil {
			manager.nSimulates++
			manager.timeSim += simTime
			manager.timeRefine += refTime
			manager.timeResimGlo += time.Since(resimStart).Nanoseconds()
			manager.pattern = 0
		}
		clear(stamps)
	}
	if manager != nil {
		Cec4_ManPrintStats(graph, pars, manager, 0)
	}
	if !finishGraph {
		return nil, nil, stats
	}
	for i := int32(0); i < gia.Gia_ManCoNum(graph); i++ {
		obj := gia.Gia_ManCo(graph, i)
		gia.Gia_ManAppendCo(reduced, gia.Gia_ObjFanin0Copy(obj))
	}
	return gia.Gia_ManCleanup(reduced), nil, stats
}
