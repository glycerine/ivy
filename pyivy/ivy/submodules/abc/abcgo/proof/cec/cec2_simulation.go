package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/sat/satoko"
	"math/bits"
	"time"
)

type cecReplayTriple struct {
	repr, node int32
	pattern    int
	phase      int32
}

// cec2CountReplayFailures mirrors the optional counterexample replay check
// performed by Cec2_ManSimulate and Cec3_ManSimulate.
func cec2CountReplayFailures(sim *cec3Simulation, triples []cecReplayTriple) int {
	failures := 0
	for _, triple := range triples {
		mask := uint64(1) << uint(triple.pattern&63)
		a := sim.object(triple.repr)[triple.pattern>>6]&mask != 0
		b := sim.object(triple.node)[triple.pattern>>6]&mask != 0
		if (a != (triple.phase != 0)) == b {
			failures++
		}
	}
	return failures
}

// cec2BeginSweepIteration matches Cec2/3_ManSimulateCis: each newly
// randomized pattern batch starts writing counterexamples at pattern one.
func cec2BeginSweepIteration(sim *cec3Simulation, graph *gia.Gia_Man_t, pattern *int) {
	cec2RandomizeCIs(sim, graph)
	*pattern = 0
}

// cec2RandomizeCIs preserves pattern zero as the all-zero assignment, as in
// Cec2_ObjSimCi. Pattern zero is used to normalize class phases.
func cec2RandomizeCIs(sim *cec3Simulation, graph *gia.Gia_Man_t) {
	for ci := int32(0); ci < gia.Gia_ManCiNum(graph); ci++ {
		id := gia.Gia_ManCiIdToId(graph, ci)
		words := sim.object(id)
		for i := range words {
			words[i] = gia.Gia_ManRandomW(0)
		}
		words[0] <<= 1
	}
}

// cec2SimulateAnds replays all patterns and refines each existing candidate
// class exactly once, matching Cec2_ManSimulate.
func cec2SimulateAnds(sim *cec3Simulation, graph *gia.Gia_Man_t) {
	cec2SimulateAndsTimed(sim, graph)
}

func cec2SimulateAndsTimed(sim *cec3Simulation, graph *gia.Gia_Man_t) (int64, int64) {
	start := time.Now()
	for id := int32(1); int(id) < graph.ObjectCount(); id++ {
		_, _, _, kind := graph.Gate(id)
		switch kind {
		case 1, 6:
		case 2:
			sim.simulateAnd(graph, id)
		default:
			panic("cec: Cec2 simulation requires an AND-inverter graph")
		}
	}
	simTime := time.Since(start).Nanoseconds()
	start = time.Now()
	if graph.HasRepresentatives() {
		for id := int32(0); int(id) < graph.ObjectCount(); id++ {
			if gia.Gia_ObjIsHead(graph, id) != 0 {
				sim.refineClass(graph, id)
			}
		}
	}
	return simTime, time.Since(start).Nanoseconds()
}

// cec2FailingOutput returns the first output and simulation pattern that is
// high; -1, -1 means no simulated pattern falsifies the miter.
func cec2FailingOutput(sim *cec3Simulation, graph *gia.Gia_Man_t) (int32, int) {
	for i := int32(0); i < gia.Gia_ManCoNum(graph); i++ {
		id := gia.Gia_ObjId(graph, gia.Gia_ManCo(graph, i))
		sim.simulateOutput(graph, id)
		if sim.equivalent(id, 0) {
			continue
		}
		for w, word := range sim.object(id) {
			if word != 0 {
				return i, w*64 + bits.TrailingZeros64(word)
			}
		}
	}
	return -1, -1
}

// cec2ReplayModel writes the SAT witness into a simulation pattern on the
// original graph. Input map keys are corresponding CI indices in both graphs.
func cec2ReplayModel(sim *cec3Simulation, graph *gia.Gia_Man_t, solver *satoko.Solver, inputs map[int32]int, pattern int) {
	for index, v := range inputs {
		sim.setInputBit(gia.Gia_ManCiIdToId(graph, index), pattern, solver.ModelValue(v) == 1)
	}
}

// cec2ZeroPhaseWitness implements the all-zero miter check before simulation.
func cec2ZeroPhaseWitness(g *gia.Gia_Man_t) *util.Abc_Cex_t {
	for i := int32(0); i < gia.Gia_ManCoNum(g); i++ {
		if gia.Gia_ObjPhase(gia.Gia_ManCo(g, i)) == 0 {
			continue
		}
		cex := util.Abc_CexAlloc(0, gia.Gia_ManCiNum(g), 1)
		cex.IPo = i
		if err := g.SetCecCombinationalWitness(cex); err != nil {
			panic(err)
		}
		return cex
	}
	return nil
}
