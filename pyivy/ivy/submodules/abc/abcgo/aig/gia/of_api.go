package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/satoko"
)

// Of_Man_t holds the native LUT-structure mapping state for diagnostics.
type Of_Man_t struct{ mapper *ofMapper }

func Of_ManSetDefaultPars(pars *Jf_Par_t)                          { ofDefaultParameters(pars) }
func Of_ManPerformMapping(g *Gia_Man_t, pars *Jf_Par_t) *Gia_Man_t { return ofMap(g, pars) }
func Of_StoCreate(g *Gia_Man_t, pars *Jf_Par_t) *Of_Man_t {
	cuts := newOfCutManager(g, pars)
	return &Of_Man_t{mapper: newOfMapper(cuts)}
}
func Of_StoDelete(p *Of_Man_t) {
	if p == nil || p.mapper == nil {
		return
	}
	if truth := p.mapper.cuts.truth; truth != nil {
		vec.Vec_MemHashFree(truth.memory)
		vec.Vec_MemFree(truth.memory)
	}
	p.mapper = nil
}
func Of_ManComputeCuts(p *Of_Man_t)              { p.mapper.cuts.computeCuts() }
func Of_ManComputeForward1(p *Of_Man_t)          { p.mapper.forward() }
func Of_ManComputeBackward1(p *Of_Man_t)         { p.mapper.backward() }
func Of_ManComputeForward2(p *Of_Man_t)          { p.mapper.forwardExact() }
func Of_ManComputeBackward3(p *Of_Man_t)         { p.mapper.backwardExact() }
func Of_ManComputeForwardDircon1(p *Of_Man_t)    { p.mapper.forwardDirect() }
func Of_ManComputeBackwardDircon1(p *Of_Man_t)   { p.mapper.backwardDirect() }
func Of_ManDeriveMapping(p *Of_Man_t) *Gia_Man_t { return p.mapper.deriveMapping() }

// Of_ManPrintCuts prints the SAT diagnostic used by the original mapper.
func Of_ManPrintCuts(p *Of_Man_t) {
	m := p.mapper
	nodes, cuts, best := 0, 0, 0
	for _, obj := range m.cuts.graph.pObjs {
		if Gia_ObjIsAnd(obj) == 0 {
			continue
		}
		nodes++
		if m.objects[obj.id].refs > 0 {
			best++
		}
		if m.cuts.sets[obj.id] != 0 {
			cuts += int(m.cuts.storage.at(m.cuts.sets[obj.id])[0])
		}
	}
	fmt.Printf("Total:   Ands = %d.  Luts = %d.  Cuts = %d.\n", nodes, best, cuts)
	ofPrintSatResult(m.satProblem(best - 2))
}

func ofPrintSatResult(solver *satoko.Solver) {
	status := solver.SolveAssumptionsLimit(nil, 1000000)
	switch status {
	case satoko.SATOKO_SAT:
		fmt.Println("Satisfiable.")
	case satoko.SATOKO_UNSAT:
		fmt.Println("Unsatisfiable.")
	default:
		fmt.Println("Undecided.")
	}
}

func Of_ManComputeBackward2(p *Of_Man_t) { p.mapper.backwardRequired() }

// Of_ManCreateSat runs the LUT-count diagnostic on the enumerated cuts.
func Of_ManCreateSat(p *Of_Man_t, nCutsAll int32, first, cutNum, bestNode, bestCut *Vec_Int_t) {
	if p == nil || p.mapper == nil || first == nil || cutNum == nil || bestNode == nil || bestCut == nil {
		panic("gia: invalid Of SAT diagnostic")
	}
	count := int32(0)
	for _, obj := range p.mapper.cuts.graph.pObjs {
		if Gia_ObjIsAnd(obj) == 0 {
			continue
		}
		if p.mapper.cuts.sets[obj.id] == 0 {
			panic("gia: missing Of SAT cuts")
		}
		count += p.mapper.cuts.storage.at(p.mapper.cuts.sets[obj.id])[0]
	}
	if count != nCutsAll {
		panic("gia: inconsistent Of SAT cut count")
	}
	ofPrintSatResult(p.mapper.satProblem(vec.Vec_IntSize(bestNode) - 2))
}
