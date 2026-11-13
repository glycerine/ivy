package abs

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"time"
)

func printOldRefinementStats(p *Saig_RefMan_t, reasons, inputs *vec.Vec_Int_t, started time.Time) {
	fmt.Printf("Frame PIs = %4d (essential = %4d)   AIG PIs = %4d (essential = %4d)   Time = %s\n", len(p.pFrames.Inputs()), vec.Vec_IntSize(reasons), saig.Saig_ManPiNum(p.pAig)-p.nInputs, vec.Vec_IntSize(inputs), time.Since(started))
}

// Saig_ManFindCexCareBits computes a sufficient set of care bits by
// tracing controlling values in the unrolled graph, as in ABC's old flow.
func Saig_ManFindCexCareBits(g *aig.Graph, cex *util.Abc_Cex_t, nInputs, newOrder, verbose int32) *util.Abc_Cex_t {
	started := time.Now()
	p := Saig_RefManStart(g, cex, nInputs, verbose)
	reasons := Saig_RefManFindReason(p)
	inputs := Saig_RefManReason2Inputs(p, reasons)
	printOldRefinementStats(p, reasons, inputs, started)
	care := Saig_RefManReason2Cex(p, reasons)
	Saig_RefManStop(p)
	return care
}

func Saig_ManExtendCounterExampleTest3(g *aig.Graph, firstFlopPI int32, cex *util.Abc_Cex_t, verbose int32) *vec.Vec_Int_t {
	if saig.Saig_ManPiNum(g) != cex.NPis {
		fmt.Printf("Saig_ManExtendCounterExampleTest3(): The PI count of AIG (%d) does not match that of cex (%d).\n", len(g.Inputs()), cex.NPis)
		return nil
	}
	started := time.Now()
	p := Saig_RefManStart(g, cex, firstFlopPI, verbose)
	reasons := Saig_RefManFindReason(p)
	inputs := Saig_RefManReason2Inputs(p, reasons)
	printOldRefinementStats(p, reasons, inputs, started)
	Saig_RefManStop(p)
	return inputs
}
