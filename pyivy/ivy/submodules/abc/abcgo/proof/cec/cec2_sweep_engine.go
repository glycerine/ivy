package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/sat/satoko"
)

// cec2SweepSettings holds the controls used by Cec2_ManPerformSweeping.
type cec2SweepSettings struct {
	words, rounds, iterations int
	conflictLimit             int64
	useCones                  bool
	isMiter                   bool
	finishGraph               bool
}

// cec2SweepEquivalences performs Cec2's simulation-guided SAT sweeping and
// constructs the reduced graph. The returned simulation is also used for
// source-style miter counterexample handling by the caller.
func cec2SweepEquivalences(g *gia.Gia_Man_t, opts cec2SweepSettings) (*gia.Gia_Man_t, *cec3Simulation, *util.Abc_Cex_t) {
	if opts.words < 1 || opts.rounds < 0 || opts.iterations < 0 {
		panic("cec: invalid Cec2 settings")
	}
	defer gia.Gia_ManCleanMark01(g)
	n := g.ObjectCount()
	copies := make([]int32, n)
	for i := range copies {
		copies[i] = -1
	}
	copies[0] = 0
	gia.Gia_ManFillValue(g)
	gia.Gia_ObjSetValue(gia.Gia_ManObj(g, 0), 0)
	reduced := gia.Gia_ManStart(int32(n))
	for i := int32(0); i < gia.Gia_ManCiNum(g); i++ {
		id := gia.Gia_ManCiIdToId(g, i)
		copies[id] = gia.Gia_ManAppendCi(reduced)
		gia.Gia_ObjSetValue(gia.Gia_ManObj(g, id), copies[id])
	}
	gia.Gia_ManHashAlloc(reduced)
	solver := satoko.New()
	if !opts.useCones {
		solver.Bookmark()
	}
	cnf := newCec2CNF(reduced, solver)
	gia.Gia_ManRandomW(1)
	gia.Gia_ManSetPhase(g)
	sim := newCec3Simulation(g.ObjectCount(), opts.words)
	cec2RandomizeCIs(sim, g)
	cec2SimulateAnds(sim, g)
	if opts.isMiter {
		if cex := cec2CaptureMiterFailure(sim, g); cex != nil {
			return nil, sim, cex
		}
	}
	sim.createClasses(g)
	for i := 0; i < opts.rounds; i++ {
		cec2RandomizeCIs(sim, g)
		cec2SimulateAnds(sim, g)
		if opts.isMiter {
			if cex := cec2CaptureMiterFailure(sim, g); cex != nil {
				return nil, sim, cex
			}
		}
	}
	pattern := 0
	var failure *util.Abc_Cex_t
	for iter := 0; iter < opts.iterations; iter++ {
		disproved := false
		cec2RandomizeCIs(sim, g)
		for id := int32(1); int(id) < n; id++ {
			a, b, _, kind := g.Gate(id)
			if kind != 2 || copies[id] >= 0 || gia.Gia_ObjFailed(g, id) != 0 {
				continue
			}
			if copies[a>>1] < 0 || copies[b>>1] < 0 {
				continue
			}
			oldN := reduced.ObjectCount()
			lit := gia.Gia_ManHashAnd(reduced, copies[a>>1]^(a&1), copies[b>>1]^(b&1))
			copies[id] = lit
			gia.Gia_ObjSetValue(gia.Gia_ManObj(g, id), lit)
			if reduced.ObjectCount() != oldN {
				obj := gia.Gia_ManObj(reduced, lit>>1)
				gia.Gia_ObjSetMark0(obj, gia.Gia_ObjIsMuxType(obj))
				gia.Gia_ObjSetPhase(reduced, obj)
			}
			repr := gia.Gia_ObjRepr(g, id)
			if repr == gia.GIA_VOID || copies[repr] < 0 {
				continue
			}
			phase := int((lit^copies[repr])&1) ^ int(gia.Gia_ObjPhase(gia.Gia_ManObj(g, id))^gia.Gia_ObjPhase(gia.Gia_ManObj(g, repr)))
			if lit>>1 == copies[repr]>>1 {
				gia.Gia_ObjSetProved(g, id)
				continue
			}
			status, inputs := cnf.solveTwoWithCone(copies[repr]>>1, lit>>1, phase, opts.conflictLimit, opts.useCones)
			switch status {
			case satoko.SATOKO_SAT:
				pattern++
				if pattern == 64*opts.words {
					pattern = 1
				}
				cec2ReplayModel(sim, g, solver, inputs, pattern)
				copies[id] = -1
				gia.Gia_ObjSetValue(gia.Gia_ManObj(g, id), -1)
				disproved = true
			case satoko.SATOKO_UNSAT:
				copies[id] = copies[repr] ^ (gia.Gia_ObjPhase(gia.Gia_ManObj(g, id)) ^ gia.Gia_ObjPhase(gia.Gia_ManObj(g, repr)))
				gia.Gia_ObjSetValue(gia.Gia_ManObj(g, id), copies[id])
				gia.Gia_ObjSetProved(g, id)
			default:
				gia.Gia_ObjSetFailed(g, id)
			}
			if !opts.useCones {
				cnf.rollback()
			}
		}
		if !disproved {
			break
		}
		cec2SimulateAnds(sim, g)
		if opts.isMiter {
			failure = cec2CaptureMiterFailure(sim, g)
			if failure != nil {
				break
			}
		}
	}
	if !opts.finishGraph {
		return nil, sim, failure
	}
	for id := int32(1); int(id) < n; id++ {
		a, b, _, kind := g.Gate(id)
		if kind == 2 && copies[id] < 0 {
			copies[id] = gia.Gia_ManHashAnd(reduced, copies[a>>1]^(a&1), copies[b>>1]^(b&1))
			gia.Gia_ObjSetValue(gia.Gia_ManObj(g, id), copies[id])
		}
		if kind == 6 {
			copies[id] = gia.Gia_ManAppendCo(reduced, copies[a>>1]^(a&1))
			gia.Gia_ObjSetValue(gia.Gia_ManObj(g, id), copies[id])
		}
	}
	gia.Gia_ManHashStop(reduced)
	cleaned := gia.Gia_ManCleanup(reduced)
	gia.Gia_ManCopyNameSpec(cleaned, g)
	return cleaned, sim, failure
}

func cec2CaptureMiterFailure(sim *cec3Simulation, g *gia.Gia_Man_t) *util.Abc_Cex_t {
	output, pattern := cec2FailingOutput(sim, g)
	if output < 0 {
		return nil
	}
	cex := util.Abc_CexAlloc(0, gia.Gia_ManCiNum(g), 1)
	cex.IPo = output
	for i := int32(0); i < gia.Gia_ManCiNum(g); i++ {
		id := gia.Gia_ManCiIdToId(g, i)
		if sim.object(id)[pattern/64]&(uint64(1)<<uint(pattern%64)) != 0 {
			util.Abc_InfoSetBit(cex.PData, i)
		}
	}
	if err := g.SetCecCombinationalWitness(cex); err != nil {
		panic(err)
	}
	return cex
}
