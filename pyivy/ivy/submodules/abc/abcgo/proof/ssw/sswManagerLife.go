package ssw

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
	"strconv"
)

func sswReduction(before, after int32) float64 {
	denominator := before
	if denominator == 0 {
		denominator = 1
	}
	return 100 * float64(before-after) / float64(denominator)
}
func Ssw_ManPrintStats(p *Ssw_Man_t) {
	memory := float64(len(p.pAig.Nodes())) * float64(p.nFrames) * (2*4 + 2*strconv.IntSize/8) / (1 << 20)
	fmt.Printf("Parameters: F = %d. AddF = %d. C-lim = %d. Constr = %d. MaxLev = %d. Mem = %0.2f MB.\n", p.pPars.NFramesK, p.pPars.NFramesAddSim, p.pPars.NBTLimit, aig.Aig_ManConstrNum(p.pAig), p.pPars.NMaxLevs, memory)
	fmt.Printf("AIG       : PI = %d. PO = %d. Latch = %d. Node = %d.  Ave SAT vars = %d.\n", saig.Saig_ManPiNum(p.pAig), saig.Saig_ManPoNum(p.pAig), p.pAig.RegisterCount(), aig.Aig_ManNodeNum(p.pAig), 0)
	fmt.Printf("SAT calls : Proof = %d. Cex = %d. Fail = %d. Lits proved = %d.\n", p.nSatProof, p.nSatCallsSat, p.nSatFailsReal, Ssw_ManCountEquivs(p))
	fmt.Printf("SAT solver: Vars max = %d. Calls max = %d. Recycles = %d. Sim rounds = %d.\n", p.nVarsMax, p.nCallsMax, p.nRecyclesTotal, p.nSimRounds)
	fmt.Printf("NBeg = %d. NEnd = %d. (Gain = %6.2f %%).  RBeg = %d. REnd = %d. (Gain = %6.2f %%).\n", p.nNodesBeg, p.nNodesEnd, sswReduction(p.nNodesBeg, p.nNodesEnd), p.nRegsBeg, p.nRegsEnd, sswReduction(p.nRegsBeg, p.nRegsEnd))
	p.timeOther = p.timeTotal - p.timeBmc - p.timeReduce - p.timeMarkCones - p.timeSimSat - p.timeSat
	fmt.Printf("BMC: %s. Spec reduce: %s. Mark cones: %s. Sim SAT: %s. SAT: %s. Other: %s. Total: %s.\n", p.timeBmc, p.timeReduce, p.timeMarkCones, p.timeSimSat, p.timeSat, p.timeOther, p.timeTotal)
	if aig.Aig_ManConstrNum(p.pAig) != 0 {
		fmt.Printf("Total cones = %d. Constraint cones = %d. Total equivs = %d. Removed equivs = %d.\n", p.nConesTotal, p.nConesConstr, p.nEquivsTotal, p.nEquivsConstr)
		fmt.Printf("NBeg = %d. NEnd = %d. (Gain = %6.2f %%).  RBeg = %d. REnd = %d. (Gain = %6.2f %%).\n", p.nNodesBegC, p.nNodesEndC, sswReduction(p.nNodesBegC, p.nNodesEndC), p.nRegsBegC, p.nRegsEndC, sswReduction(p.nRegsBegC, p.nRegsEndC))
	}
}
func Ssw_ManCleanup(p *Ssw_Man_t) {
	if p.pMSat != nil {
		panic("ssw: SAT manager still active")
	}
	if p.pFrames != nil {
		aig.Aig_ManCleanMarkAB(p.pFrames)
		aig.Aig_ManStop(p.pFrames)
		p.pFrames = nil
		for i := range p.pNodeToFrames {
			p.pNodeToFrames[i] = aig.InvalidEdge
		}
	}
	p.vSimInfo = nil
	p.nConstrTotal = 0
	p.nConstrReduced = 0
}
func Ssw_ManStop(p *Ssw_Man_t) {
	if p == nil {
		return
	}
	if p.pPars.FVerbose != 0 {
		Ssw_ManPrintStats(p)
	}
	// ABC leaves the input graph's fanout index active after manager stop.
	if p.ppClasses != nil {
		Ssw_ClassesStop(p.ppClasses)
	}
	if p.pSml != nil {
		Ssw_SmlStop(p.pSml)
	}
	*p = Ssw_Man_t_{}
}
