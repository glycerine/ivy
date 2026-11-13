package cec

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/glucose"
	"time"
)

// cec3SweepSettings holds the controls used by Cec2_ManPerformSweeping.
type cec3SweepSettings struct {
	words, rounds, iterations int
	conflictLimit             int64
	useCones                  bool
	isMiter                   bool
	finishGraph               bool
	verbose, veryVerbose      bool
}

// cec3SweepEquivalences performs Cec3's simulation-guided SAT sweeping and
// constructs the reduced graph. The returned simulation is also used for
// source-style miter counterexample handling by the caller.
func cec3SweepEquivalences(g *gia.Gia_Man_t, opts cec3SweepSettings) (*gia.Gia_Man_t, *cec3Simulation, *util.Abc_Cex_t) {
	if opts.words < 1 || opts.rounds < 0 || opts.iterations < 0 {
		panic("cec: invalid Cec3 settings")
	}
	pars := &Cec3_Par_t{nSimWords: int32(opts.words), nSimRounds: int32(opts.rounds), nItersMax: int32(opts.iterations), nConfLimit: int32(opts.conflictLimit)}
	if opts.useCones {
		pars.fUseCones = 1
	}
	if opts.isMiter {
		pars.fIsMiter = 1
	}
	man := Cec3_ManCreate(g, pars)
	defer Cec3_ManDestroy(man)
	n := g.ObjectCount()
	copies := make([]int32, n)
	for i := range copies {
		copies[i] = -1
	}
	copies[0] = 0
	reduced := man.pNew
	for i := int32(0); i < gia.Gia_ManCiNum(g); i++ {
		id := gia.Gia_ManCiIdToId(g, i)
		copies[id] = gia.Gia_ObjValue(gia.Gia_ManObj(g, id))
	}
	gia.Gia_ManRandomW(1)
	gia.Gia_ManSetPhase(g)
	if opts.isMiter {
		if cex := cec2ZeroPhaseWitness(g); cex != nil {
			return nil, nil, cex
		}
	}
	sim := newCec3Simulation(g.ObjectCount(), opts.words)
	telemetry := &cecSweepTiming{start: time.Now()}
	defer telemetry.print(opts.verbose)
	simulate := func() { a, b := cec2SimulateAndsTimed(sim, g); telemetry.simTime += a; telemetry.refineTime += b }
	cec2RandomizeCIs(sim, g)
	simulate()
	if opts.isMiter {
		if cex := cec2CaptureMiterFailure(sim, g); cex != nil {
			return nil, sim, cex
		}
	}
	classStart := time.Now()
	sim.createClasses(g)
	telemetry.refineTime += time.Since(classStart).Nanoseconds()
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
		man.nPatterns = 0
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
			gia.Gia_ManExtendCopy2Array(reduced)
			satStart := time.Now()
			status := Cec3_ManSolveTwo(man, copies[repr]>>1, lit>>1, int32(phase))
			switch status {
			case glucose.GLUCOSE_SAT:
				telemetry.sat++
				telemetry.satTime += time.Since(satStart).Nanoseconds()
				pattern++
				if pattern == 64*opts.words {
					pattern = 1
				}
				pairs := vec.Vec_IntArray(man.vObjSatPairs)
				for i := 0; i+1 < len(pairs); i += 2 {
					sim.setInputBit(pairs[i], pattern, glucose.Bmcg_sat_solver_read_cex_varvalue(man.pSat, pairs[i+1]) != 0)
				}
				triples = append(triples, cecReplayTriple{repr: repr, node: id, pattern: pattern, phase: gia.Gia_ObjPhase(gia.Gia_ManObj(g, id)) ^ gia.Gia_ObjPhase(gia.Gia_ManObj(g, repr))})
				man.nSatSat++
				man.nPatterns++
				copies[id] = -1
				gia.Gia_ObjSetValue(gia.Gia_ManObj(g, id), -1)
				disproved = true
			case glucose.GLUCOSE_UNSAT:
				telemetry.unsat++
				telemetry.unsatTime += time.Since(satStart).Nanoseconds()
				man.nSatUnsat++
				copies[id] = copies[repr] ^ (gia.Gia_ObjPhase(gia.Gia_ManObj(g, id)) ^ gia.Gia_ObjPhase(gia.Gia_ManObj(g, repr)))
				gia.Gia_ObjSetValue(gia.Gia_ManObj(g, id), copies[id])
				gia.Gia_ObjSetProved(g, id)
			default:
				telemetry.undec++
				telemetry.undecTime += time.Since(satStart).Nanoseconds()
				man.nSatUndec++
				gia.Gia_ObjSetFailed(g, id)
			}
			if !opts.useCones {
				extraStart := time.Now()
				glucose.Bmcg_sat_solver_reset(man.pSat)
				telemetry.extraTime += time.Since(extraStart).Nanoseconds()
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
