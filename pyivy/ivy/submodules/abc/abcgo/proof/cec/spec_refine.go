package cec

import (
	"fmt"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
)

// SpecBMC runs bounded model checking on the speculative-reduction graph.
// The callback avoids a GIA/CEC/BMC import cycle and returns an owned witness.
type SpecBMC func(graph *aig.Graph, start, frames, conflicts, verbose int32) *util.Abc_Cex_t

// Gia_CommandSpecI iteratively refines speculative equivalences using BMC.
// It writes the source's diagnostic gore.aig and gsrm.aig files after each
// successful resimulation, and returns zero when classes are absent.
func Gia_CommandSpecI(p *Gia_Man_t, frames, conflicts, useStart, checkMiter, verbose int32, run SpecBMC) int32 {
	if p == nil || !p.HasRepresentatives() {
		fmt.Println("Gia_CommandSpecI(): Equivalence classes are not defined.")
		return 0
	}
	if run == nil {
		panic("cec: speculative refinement requires a BMC callback")
	}
	gia.Gia_ManCleanMark0(p)
	var start int32
	for iteration := 0; ; iteration++ {
		if gia.Gia_ManHasNoEquivs(p) != 0 {
			fmt.Println("Gia_CommandSpecI: No equivalences left.")
			break
		}
		fmt.Printf("ITER %3d : ", iteration)
		gia.Gia_ManPrintStatsClasses(p)
		if Cec_ManCheckNonTrivialCands(p) == 0 {
			fmt.Println("Gia_CommandSpecI: There are only trivial equiv candidates left (PO drivers). Quitting.")
			break
		}
		reduced := gia.Gia_ManSpecReduce(p, 0, 0, 1, 0, 0)
		if reduced == nil {
			panic("cec: speculative reduction failed")
		}
		graph := gia.Gia_ManToAig(reduced, 0)
		gia.Gia_ManStop(reduced)
		cex := run(graph, start, frames, conflicts, verbose)
		aig.Aig_ManStop(graph)
		if cex == nil {
			fmt.Println("Gia_CommandSpecI(): Internal BMC could not find a counter-example.")
			break
		}
		if useStart != 0 {
			start = cex.IFrame
		}
		var pars Cec_ParSim_t
		Cec_ManSimSetDefaultParams(&pars)
		pars.FCheckMiter = checkMiter
		if Cec_ManSeqResimulateCounter(p, &pars, cex) != 0 {
			break
		}
		if err := gia.Gia_AigerWrite(p, "gore.aig", 0, 0, 0); err != nil {
			panic(fmt.Sprintf("cec: write equivalence diagnostic: %v", err))
		}
		reduced = gia.Gia_ManSpecReduce(p, 0, 0, 1, 0, 0)
		if reduced != nil {
			swept := gia.Gia_ManSeqStructSweep(reduced, 1, 1, 0)
			gia.Gia_ManStop(reduced)
			if err := gia.Gia_AigerWrite(swept, "gsrm.aig", 0, 0, 0); err != nil {
				panic(fmt.Sprintf("cec: write reduction diagnostic: %v", err))
			}
			gia.Gia_ManStop(swept)
		}
	}
	return 1
}
