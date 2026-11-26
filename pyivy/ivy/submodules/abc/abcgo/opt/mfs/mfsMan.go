package mfs

import (
	"fmt"

	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/bool/bdc"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
)

// Port of mfsMan.c.

func Mfs_ManAlloc(pPars *Mfs_Par_t) *Mfs_Man_t {
	p := &Mfs_Man_t{}
	p.pPars = pPars
	p.vProjVarsCnf = make([]int32, 0, 100)
	p.vProjVarsSat = make([]int32, 0, 100)
	p.nDivWords = util.Abc_BitWordNum(p.pPars.NWinMax + MFS_FANIN_MAX)
	nEntries := int(p.pPars.NWinMax + MFS_FANIN_MAX + 1)
	p.vDivCexes = make([][]uint32, nEntries)
	storage := make([]uint32, nEntries*int(p.nDivWords))
	for i := range p.vDivCexes {
		p.vDivCexes[i] = storage[i*int(p.nDivWords) : (i+1)*int(p.nDivWords) : (i+1)*int(p.nDivWords)]
	}
	p.pMan = bsat.Int_ManAlloc()
	p.vLevels = vec.Vec_VecStart(32)
	p.vMfsFanins = make([]*abc.Object, 0, 32)
	return p
}

func Mfs_ManClean(p *Mfs_Man_t) {
	if p.pSat != nil {
		p.pSat.Delete()
	}
	p.pAigWin = nil
	p.pCnf = nil
	p.pSat = nil
	p.vRoots = nil
	p.vSupp = nil
	p.vNodes = nil
	p.vDivs = nil
}

func Mfs_ManPrint(p *Mfs_Man_t) {
	if p.pPars.FResub != 0 {
		fmt.Printf("Nodes = %d. Try = %d. Resub = %d. Div = %d. SAT calls = %d. Timeouts = %d. MaxDivs = %d.\n",
			p.nTotalNodesBeg, p.nNodesTried, p.nNodesResub, p.nTotalDivs, p.nSatCalls, p.nTimeOuts, p.nMaxDivs)
		fmt.Printf("Attempts :   ")
		fmt.Printf("Remove %6d out of %6d (%6.2f %%)   ", p.nRemoves, p.nTryRemoves, 100.0*float64(p.nRemoves)/float64(max(1, p.nTryRemoves)))
		fmt.Printf("Resub  %6d out of %6d (%6.2f %%)   ", p.nResubs, p.nTryResubs, 100.0*float64(p.nResubs)/float64(max(1, p.nTryResubs)))
		fmt.Printf("\n")
		fmt.Printf("Reduction:   ")
		fmt.Printf("Nodes  %6d out of %6d (%6.2f %%)   ", p.nTotalNodesBeg-p.nTotalNodesEnd, p.nTotalNodesBeg, 100.0*float64(p.nTotalNodesBeg-p.nTotalNodesEnd)/float64(max(1, p.nTotalNodesBeg)))
		fmt.Printf("Edges  %6d out of %6d (%6.2f %%)   ", p.nTotalEdgesBeg-p.nTotalEdgesEnd, p.nTotalEdgesBeg, 100.0*float64(p.nTotalEdgesBeg-p.nTotalEdgesEnd)/float64(max(1, p.nTotalEdgesBeg)))
		fmt.Printf("\n")
		if p.pPars.FPower != 0 {
			fmt.Printf("Power( %5.2f, %4.2f%%) \n",
				p.TotalSwitchingBeg-p.TotalSwitchingEnd,
				100.0*(p.TotalSwitchingBeg-p.TotalSwitchingEnd)/p.TotalSwitchingBeg)
		}
		if p.pPars.FSwapEdge != 0 {
			fmt.Printf("Swappable edges = %d. Total edges = %d. Ratio = %5.2f.\n",
				p.nNodesResub, abc.Abc_NtkGetTotalFanins(p.pNtk), 1.00*float64(p.nNodesResub)/float64(abc.Abc_NtkGetTotalFanins(p.pNtk)))
		}
	} else {
		fmt.Printf("Nodes = %d. Try = %d. Total mints = %d. Local DC mints = %d. Ratio = %5.2f.\n",
			p.nTotalNodesBeg, p.nNodesTried, p.nMintsTotal, p.nMintsTotal-p.nMintsCare,
			1.0*float64(p.nMintsTotal-p.nMintsCare)/float64(p.nMintsTotal))
		fmt.Printf("Nodes resyn = %d. Ratio = %5.2f.  Total AIG node gain = %d. Timeouts = %d.\n",
			p.nNodesDec, 1.0*float64(p.nNodesDec)/float64(p.nNodesTried), p.nNodesGained, p.nTimeOuts)
	}
	prtp := func(name string, t int64) {
		total := p.timeTotal
		pct := 0.0
		if total != 0 {
			pct = 100.0 * float64(t) / float64(total)
		}
		fmt.Printf("%s =%9.2f sec (%6.2f %%)\n", name, float64(t)/1e9, pct)
	}
	prtp("Win", p.timeWin)
	prtp("Div", p.timeDiv)
	prtp("Aig", p.timeAig)
	prtp("Gia", p.timeGia)
	prtp("Cnf", p.timeCnf)
	prtp("Sat", p.timeSat-p.timeInt)
	prtp("Int", p.timeInt)
	prtp("ALL", p.timeTotal)
}

func Mfs_ManStop(p *Mfs_Man_t) {
	if p.pPars.FVerbose != 0 {
		Mfs_ManPrint(p)
	}
	if p.pManDec != nil {
		bdc.Bdc_ManFree(p.pManDec)
	}
	Mfs_ManClean(p)
	bsat.Int_ManFree(p.pMan)
	vec.Vec_VecFree(p.vLevels)
}
