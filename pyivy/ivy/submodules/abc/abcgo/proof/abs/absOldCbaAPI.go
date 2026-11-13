package abs

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"time"
)

func printCbaStats(p *Saig_ManCba_t, reasons, inputs *vec.Vec_Int_t, started time.Time) {
	fmt.Printf("Frame PIs = %4d (essential = %4d)   AIG PIs = %4d (essential = %4d)   Time = %s\n", len(p.pFrames.Inputs()), vec.Vec_IntSize(reasons), saig.Saig_ManPiNum(p.pAig)-p.nInputs, vec.Vec_IntSize(inputs), time.Since(started))
}
func Saig_ManCbaFindCexCareBits(g *aig.Graph, cex *util.Abc_Cex_t, nInputs, verbose int32) *util.Abc_Cex_t {
	started := time.Now()
	p := Saig_ManCbaStart(g, cex, nInputs, verbose)
	p.pFrames = Saig_ManCbaUnrollWithCex(g, cex, nInputs, &p.vMapPiF2A, &p.vReg2Frame)
	reasons := Saig_ManCbaFindReason(p)
	if p.vReg2Frame != nil {
		Saig_ManCbaShrink(p)
	}
	if verbose != 0 {
		inputs := Saig_ManCbaReason2Inputs(p, reasons)
		printCbaStats(p, reasons, inputs, started)
	}
	care := Saig_ManCbaReason2Cex(p, reasons)
	Saig_ManCbaStop(p)
	return care
}
func Saig_ManCbaFilterInputs(g *aig.Graph, firstFlopPI int32, cex *util.Abc_Cex_t, verbose int32) *vec.Vec_Int_t {
	if saig.Saig_ManPiNum(g) != cex.NPis {
		fmt.Printf("Saig_ManCbaFilterInputs(): The PI count of AIG (%d) does not match that of cex (%d).\n", len(g.Inputs()), cex.NPis)
		return nil
	}
	started := time.Now()
	p := Saig_ManCbaStart(g, cex, firstFlopPI, verbose)
	p.pFrames = Saig_ManCbaUnrollWithCex(g, cex, firstFlopPI, &p.vMapPiF2A, &p.vReg2Frame)
	reasons := Saig_ManCbaFindReason(p)
	inputs := Saig_ManCbaReason2Inputs(p, reasons)
	if verbose != 0 {
		printCbaStats(p, reasons, inputs, started)
	}
	Saig_ManCbaStop(p)
	return inputs
}
