package cec

import (
	"fmt"
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
	verbose, veryVerbose      bool
}

// cec2SweepEquivalences performs Cec2's simulation-guided SAT sweeping and
// constructs the reduced graph. The returned simulation is also used for
// source-style miter counterexample handling by the caller.
func cec2SweepEquivalences(g *gia.Gia_Man_t, opts cec2SweepSettings) (*gia.Gia_Man_t, *cec3Simulation, *util.Abc_Cex_t) {
	unpin := cecPinClockThread()
	defer unpin()
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
	if opts.isMiter {
		if cex := cec2ZeroPhaseWitness(g); cex != nil {
			return nil, nil, cex
		}
	}
	sim := newCec3Simulation(g.ObjectCount(), opts.words)
	telemetry := &cecSweepTiming{start: cecCPUClock()}
	defer telemetry.print(opts.verbose)
	simulate := func() { a, b := cec2SimulateAndsTimed(sim, g); telemetry.simTime += a; telemetry.refineTime += b }
	cec2RandomizeCIs(sim, g)
	simulate()
	if opts.isMiter {
		if cex := cec2CaptureMiterFailure(sim, g); cex != nil {
			return nil, sim, cex
		}
	}
	classStart := cecCPUClock()
	sim.createClasses(g)
	telemetry.refineTime += (cecCPUClock() - classStart)
	telemetry.classes(g, opts.verbose, opts.veryVerbose)
	for i := 0; i < opts.rounds; i++ {
		cec2RandomizeCIs(sim, g)
		simulate()
		if opts.isMiter {
			if cex := cec2CaptureMiterFailure(sim, g); cex != nil {
				return nil, sim, cex
			}
		}
		telemetry.classes(g, opts.verbose, opts.veryVerbose)
	}
	pattern := 0
	var failure *util.Abc_Cex_t
	for iter := 0; iter < opts.iterations; iter++ {
		disproved := false
		cec2BeginSweepIteration(sim, g, &pattern)
		var triples []cecReplayTriple
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
			satStart := cecCPUClock()
			status, inputs := cnf.solveTwoWithCone(copies[repr]>>1, lit>>1, phase, opts.conflictLimit, opts.useCones)
			switch status {
			case satoko.SATOKO_SAT:
				telemetry.sat++
				telemetry.satTime += (cecCPUClock() - satStart)
				pattern++
				if pattern == 64*opts.words {
					pattern = 1
				}
				cec2ReplayModel(sim, g, solver, inputs, pattern)
				triples = append(triples, cecReplayTriple{repr: repr, node: id, pattern: pattern, phase: gia.Gia_ObjPhase(gia.Gia_ManObj(g, id)) ^ gia.Gia_ObjPhase(gia.Gia_ManObj(g, repr))})
				copies[id] = -1
				gia.Gia_ObjSetValue(gia.Gia_ManObj(g, id), -1)
				disproved = true
			case satoko.SATOKO_UNSAT:
				telemetry.unsat++
				telemetry.unsatTime += (cecCPUClock() - satStart)
				copies[id] = copies[repr] ^ (gia.Gia_ObjPhase(gia.Gia_ManObj(g, id)) ^ gia.Gia_ObjPhase(gia.Gia_ManObj(g, repr)))
				gia.Gia_ObjSetValue(gia.Gia_ManObj(g, id), copies[id])
				gia.Gia_ObjSetProved(g, id)
			default:
				telemetry.undec++
				telemetry.undecTime += (cecCPUClock() - satStart)
				gia.Gia_ObjSetFailed(g, id)
			}
			if !opts.useCones {
				extraStart := cecCPUClock()
				cnf.rollback()
				telemetry.extraTime += (cecCPUClock() - extraStart)
			}
		}
		if !disproved {
			telemetry.classes(g, opts.verbose, opts.veryVerbose)
			break
		}
		simulate()
		if fails := cec2CountReplayFailures(sim, triples); fails > 0 && opts.verbose {
			fmt.Printf("Failed to resimulate %d times with pattern = %d  (total = %d).\n", fails, len(triples), 64*opts.words)
		}
		if opts.isMiter {
			failure = cec2CaptureMiterFailure(sim, g)
			if failure != nil {
				break
			}
		}
		telemetry.classes(g, opts.verbose, opts.veryVerbose)
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
