package gia

import (
	"fmt"
	"time"
)

func Sbl_ManPrintRuntime(p *Sbl_Man_t) {
	p.timeOther = p.timeTotal - p.timeWin - p.timeCut - p.timeSat - p.timeTime
	fmt.Println("Runtime breakdown:")
	for _, entry := range []struct {
		name    string
		elapsed time.Duration
	}{{"Win", p.timeWin}, {"Cut", p.timeCut}, {"Sat", p.timeSat}, {"Sat SAT", p.timeSatSat}, {"Sat UNSAT", p.timeSatUns}, {"Sat unknown", p.timeSatUnd}, {"Timing", p.timeTime}, {"Other", p.timeOther}, {"All", p.timeTotal}} {
		percent := float64(0)
		if p.timeTotal != 0 {
			percent = 100 * float64(entry.elapsed) / float64(p.timeTotal)
		}
		fmt.Printf("%-11s %8.3f sec %6.2f %%\n", entry.name, entry.elapsed.Seconds(), percent)
	}
}
func Gia_ManLutSat(g *Gia_Man_t, lutSize, number, improves, conflictLimit, delayMax, edges, delay, reverse, verbose, veryVerbose int32) {
	p := Sbl_ManAlloc(g, number)
	defer Sbl_ManStop(p)
	p.LutSize, p.nBTLimit, p.DelayMax, p.nEdges = lutSize, conflictLimit, delayMax, edges
	p.fDelay, p.fReverse, p.fVerbose, p.fVeryVerbose = delay, reverse, verbose|veryVerbose, veryVerbose
	if p.fVerbose != 0 {
		fmt.Printf("Parameters: WinSize = %d AIG nodes. Conf = %d. DelayMax = %d.\n", p.nVars, p.nBTLimit, p.DelayMax)
	}
	if delay != 0 && g.VEdge1 != nil && p.DelayMax == 0 {
		p.DelayMax = Gia_ManEvalEdgeDelay(g)
	}
	Gia_ManComputeOneWinStart(g, number, reverse)
	defer func() { Gia_ManComputeOneWin(g, -1, nil, nil, nil, nil); g.VPacking = nil }()
	var improved int32
	for id := int32(1); id < g.nObjs; id++ {
		if Gia_ObjIsLut2(g, id) == 0 {
			continue
		}
		if Sbl_ManTestSat(p, id) != 2 {
			continue
		}
		improved++
		if improved == improves {
			break
		}
	}
	if p.fVerbose != 0 {
		fmt.Printf("Tried = %d. Used = %d. HashWin = %d. SmallWin = %d. LargeWin = %d. IterOut = %d. SAT runs = %d.\n", p.nTried, p.nImproved, p.nHashWins, p.nSmallWins, p.nLargeWins, p.nIterOuts, p.nRuns)
		Sbl_ManPrintRuntime(p)
	}
}
