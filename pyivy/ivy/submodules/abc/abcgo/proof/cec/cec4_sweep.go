package cec

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/sat/glucose2"
)

type cec4SweepStats struct{ sat, unsat, unknown, recycles int }

// cec4Sweep performs Cec4's simulation-guided, single-pass SAT sweep.
// Witnesses are returned to the caller for storage in the source manager.
func cec4Sweep(graph *gia.Gia_Man_t, pars *Cec_ParFra_t, simOnly, finishGraph bool, bindSolver ...func(*glucose2.SimpSolver)) (*gia.Gia_Man_t, *util.Abc_Cex_t, cec4SweepStats) {
	var stats cec4SweepStats
	pack, witness := cec4PrepareSimulation(graph, pars, simOnly)
	if witness != nil || simOnly {
		return nil, witness, stats
	}
	sim := pack.values
	reduced := gia.Gia_ManStart(gia.Gia_ManObjNum(graph))
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
	if len(bindSolver) > 0 {
		bindSolver[0](cnf.solver)
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
		if pattern > 0 {
			// Refresh all class members before representative selection.
			cec4SimulateGraph(sim, graph)
		}
		repr = cec4FindRepr(sim, graph, id)
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
		if pars.nSatVarMax != 0 && calls > int(pars.nCallsRecycle) && len(cnf.ids) > int(pars.nSatVarMax) {
			cnf.recycle()
			if len(bindSolver) > 0 {
				bindSolver[0](cnf.solver)
			}
			calls = 0
			stats.recycles++
		}
		limit := pars.nBTLimit
		if failed[copies[id]>>1] || failed[copies[repr]>>1] {
			limit = max(int32(1), limit/10)
		}
		status := cec4SweepNode(graph, reduced, sim, cnf, copies, id, repr, limit, &pattern)
		switch status {
		case glucose2.GLUCOSE_SAT:
			stats.sat++
			if pattern == 64*sim.words-2 {
				cec4SimulationRound(sim, graph, pars.nLevelMax, true)
				pattern = 0
				clear(pack.reserved.data)
			}
		case glucose2.GLUCOSE_UNSAT:
			stats.unsat++
		default:
			stats.unknown++
			failed[copies[id]>>1] = true
		}
	}
	if pattern > 0 {
		cec4SimulationRound(sim, graph, pars.nLevelMax, true)
	}
	if pars.fVerbose != 0 {
		fmt.Printf("SAT calls = %d: P = %d D = %d F = %d Recyc = %d\n", stats.sat+stats.unsat+stats.unknown, stats.unsat, stats.sat, stats.unknown, stats.recycles)
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
