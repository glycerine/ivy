package sim

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
)

func Sim_ManStart(g *aig.Graph, lightweight bool) *Sim_Man_t {
	order := graphOrder(g)
	p := &Sim_Man_t{Graph: g, Inputs: g.Inputs(), Outputs: g.Outputs(), Order: order, Words: 64, rng: newRandom()}
	p.Sim0 = Sim_UtilInfoAlloc(int(aig.Aig_ManObjNumMax(g)), p.Words, true)
	Sim_UtilSetConst(p.Sim0[0], p.Words, true)
	if !lightweight {
		p.Sim1 = Sim_UtilInfoAlloc(int(aig.Aig_ManObjNumMax(g)), p.Words, true)
		p.Structural = Sim_ComputeStrSupp(g)
		p.Functional = Sim_UtilInfoAlloc(len(p.Outputs), SIM_NUM_WORDS(len(p.Inputs)), true)
		p.Targets = make([][]int, len(p.Inputs))
	}
	return p
}
func Sim_ManStop(p *Sim_Man_t) { *p = Sim_Man_t{} }
func Sim_ManPrintStats(p *Sim_Man_t) {
	fmt.Printf("Functional supports = %d. Structural supports = %d. SAT = %d. UNSAT = %d.\n", Sim_UtilCountSuppSizes(p, false), Sim_UtilCountSuppSizes(p, true), p.SAT, p.UNSAT)
}
func Sim_ManPatAlloc(p *Sim_Man_t) *Sim_Pat_t {
	return &Sim_Pat_t{Output: -1, Data: make([]uint32, SIM_NUM_WORDS(len(p.Inputs)))}
}
func Sim_ManPatFree(_ *Sim_Man_t, p *Sim_Pat_t) { p.Data = nil }
func Sym_ManStart(g *aig.Graph, verbose bool) *Sym_Man_t {
	p := &Sym_Man_t{Graph: g, Inputs: g.Inputs(), Outputs: g.Outputs(), Order: graphOrder(g), rng: newRandom()}
	p.Words = SIM_NUM_WORDS(len(p.Inputs))
	p.Sim = Sim_UtilInfoAlloc(int(aig.Aig_ManObjNumMax(g)), p.Words, true)
	Sim_UtilSetConst(p.Sim[0], p.Words, true)
	p.Functional = Sim_ComputeFunSupp(g, verbose)
	n := len(p.Outputs)
	p.Supports = make([][]int, n)
	p.Symmetric = make([]Matrix, n)
	p.NonSymmetric = make([]Matrix, n)
	p.PairsTotal = make([]int, n)
	p.PairsSym = make([]int, n)
	p.PairsNonSym = make([]int, n)
	for o := range p.Outputs {
		p.Symmetric[o] = NewMatrix(len(p.Inputs))
		p.NonSymmetric[o] = NewMatrix(len(p.Inputs))
		for i := range p.Inputs {
			p.Symmetric[o].Set(i, i)
			if Sim_HasBit(p.Functional[o], i) {
				p.Supports[o] = append(p.Supports[o], i)
			}
		}
		n := len(p.Supports[o])
		p.PairsTotal[o] = n * (n - 1) / 2
	}
	return p
}
func Sym_ManStop(p *Sym_Man_t) { *p = Sym_Man_t{} }
func Sym_ManPrintStats(p *Sym_Man_t) {
	Sim_UtilCountPairsAll(p)
	fmt.Printf("Symmetric pairs = %d. Non-symmetric = %d. SAT = %d. UNSAT = %d.\n", sum(p.PairsSym), sum(p.PairsNonSym), p.SAT, p.UNSAT)
}
func sum(a []int) int {
	n := 0
	for _, x := range a {
		n += x
	}
	return n
}
