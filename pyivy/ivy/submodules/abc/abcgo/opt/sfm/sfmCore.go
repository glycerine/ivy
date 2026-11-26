package sfm

// Port of src/opt/sfm/sfmCore.c.

import (
	"fmt"
	"time"

	"github.com/glycerine/abcgo/bool/kit"
)

// Prints statistics.
func Sfm_NtkPrintStats(p *Sfm_Ntk_t) {
	s := &p.Stats
	p.timeOther = p.timeTotal - p.timeWin - p.timeDiv - p.timeCnf - p.timeSat
	fmt.Printf("Nodes = %d. Try = %d. Resub = %d. Div = %d (ave = %d). SAT calls = %d. Timeouts = %d. MaxDivs = %d.\n",
		Sfm_NtkNodeNum(p), s.NodesTried, s.Removes+s.Resubs, s.TotalDivs, s.TotalDivs/max(1, s.NodesTried), s.SATCalls, s.Timeouts, s.MaxDivs)

	fmt.Printf("Attempts :   ")
	fmt.Printf("Remove %6d out of %6d (%6.2f %%)   ", s.Removes, s.TryRemoves, 100.0*float64(s.Removes)/float64(max(1, s.TryRemoves)))
	fmt.Printf("Resub  %6d out of %6d (%6.2f %%)   ", s.Resubs, s.TryResubs, 100.0*float64(s.Resubs)/float64(max(1, s.TryResubs)))
	if p.pPars.UseDcs {
		fmt.Printf("Improves %6d out of %6d (%6.2f %%)   ", s.Improves, s.TryImproves, 100.0*float64(s.Improves)/float64(max(1, s.TryImproves)))
	}
	fmt.Printf("\n")

	fmt.Printf("Reduction:   ")
	fmt.Printf("Nodes  %6d out of %6d (%6.2f %%)   ", s.NodesBefore-s.NodesAfter, s.NodesBefore, 100.0*float64(s.NodesBefore-s.NodesAfter)/float64(max(1, s.NodesBefore)))
	fmt.Printf("Edges  %6d out of %6d (%6.2f %%)   ", s.EdgesBefore-s.EdgesAfter, s.EdgesBefore, 100.0*float64(s.EdgesBefore-s.EdgesAfter)/float64(max(1, s.EdgesBefore)))
	fmt.Printf("\n")

	abcPrtp("Win", p.timeWin, p.timeTotal)
	abcPrtp("Div", p.timeDiv, p.timeTotal)
	abcPrtp("Cnf", p.timeCnf, p.timeTotal)
	abcPrtp("Sat", p.timeSat, p.timeTotal)
	abcPrtp("Oth", p.timeOther, p.timeTotal)
	abcPrtp("ALL", p.timeTotal, p.timeTotal)
}

// abcPrtp mirrors ABC_PRTP.
func abcPrtp(a string, t, T time.Duration) {
	pct := 0.0
	if T != 0 {
		pct = 100.0 * float64(t) / float64(T)
	}
	fmt.Printf("%s =", a)
	fmt.Printf("%9.2f sec (%6.2f %%)\n", t.Seconds(), pct)
}

// Performs resubstitution for the node.
func Sfm_NodeResubSolve(p *Sfm_Ntk_t, iNode, f int, fRemoveOnly bool) bool {
	fSkipUpdate := false
	fVeryVerbose := false //p->pPars->fVeryVerbose && Vec_IntSize(p->vDivs) < 200;// || pNode->Id == 556;
	iVar := -1
	iFaninRem, iFaninSkip := -1, -1
	var uTruth, uMask uint64
	assert(Sfm_ObjIsNode(p, iNode), "Sfm_ObjIsNode(p, iNode)")
	assert(f >= 0 && f < Sfm_ObjFaninNum(p, iNode), "f >= 0 && f < Sfm_ObjFaninNum(p, iNode)")
	p.Stats.TryRemoves++
	// report init stats
	if p.pPars.VeryVerbose {
		fmt.Printf("%5d : Lev =%3d. Leaf =%3d.  Node =%3d.  Div=%3d.  Fanin =%4d (%d/%d). MFFC = %d\n",
			iNode, Sfm_ObjLevel(p, iNode), 0, len(p.vNodes), len(p.vDivs),
			Sfm_ObjFanin(p, iNode, f), f, Sfm_ObjFaninNum(p, iNode), Sfm_ObjMffcSize(p, Sfm_ObjFanin(p, iNode, f)))
	}
	// clean simulation info
	p.nCexes = 0
	p.vDivCexes = wrdVecFill(p.vDivCexes, len(p.vDivs), 0)
	// try removing the critical fanin
	p.vDivIds = p.vDivIds[:0]
	for i, iFanin := range p.Fanins[iNode] {
		if i != f {
			p.vDivIds = append(p.vDivIds, Sfm_ObjSatVar(p, iFanin))
		} else {
			iFaninRem = iFanin
		}
	}
	assert(iFaninRem != -1, "iFaninRem != -1")
	// find fanin to skip
	if Sfm_ObjIsFixed(p, iFaninRem) && Sfm_ObjFaninNum(p, iFaninRem) == 1 {
		iFaninSkip = Sfm_ObjFanin(p, iFaninRem, 0)
	}
	clk := time.Now()
	uTruth = Sfm_ComputeInterpolant(p)
	p.timeSat += time.Since(clk)
	// analyze outcomes
	if uTruth == SFM_SAT_UNDEC {
		p.Stats.Timeouts++
		return false
	}
	if uTruth != SFM_SAT_SAT {
		goto finish
	}
	if fRemoveOnly || p.pPars.RrOnly || len(p.vDivs) == 0 {
		return false
	}

	p.Stats.TryResubs++
	if fVeryVerbose {
		for i := 0; i < 9; i++ {
			fmt.Printf(" ")
		}
		for i := 0; i < len(p.vDivs); i++ {
			fmt.Printf("%d", i%10)
		}
		fmt.Printf("\n")
	}
	for {
		if fVeryVerbose {
			fmt.Printf("%3d: %3d ", p.nCexes, iVar)
			for _, uSign := range p.vDivCexes {
				fmt.Printf("%d", (uSign>>uint(p.nCexes-1))&1)
			}
			fmt.Printf("\n")
		}
		// find the next divisor to try
		uMask = (^uint64(0)) >> uint(64-p.nCexes)
		for iVar = 0; iVar < len(p.vDivCexes); iVar++ {
			if p.vDivCexes[iVar] == uMask && p.vDivs[iVar] != iFaninSkip {
				break
			}
		}
		if iVar == len(p.vDivs) {
			return false
		}
		assert(p.vDivs[iVar] != iFaninSkip, "Vec_IntEntry(p->vDivs, iVar) != iFaninSkip")
		// try replacing the critical fanin
		p.vDivIds = append(p.vDivIds, Sfm_ObjSatVar(p, p.vDivs[iVar]))
		clk = time.Now()
		uTruth = Sfm_ComputeInterpolant(p)
		p.timeSat += time.Since(clk)
		// analyze outcomes
		if uTruth == SFM_SAT_UNDEC {
			p.Stats.Timeouts++
			return false
		}
		if uTruth != SFM_SAT_SAT {
			goto finish
		}
		if p.nCexes == 64 {
			return false
		}
		// remove the last variable
		p.vDivIds = p.vDivIds[:len(p.vDivIds)-1]
	}
finish:
	if p.pPars.VeryVerbose {
		if iVar == -1 {
			fmt.Printf("Node %d: Fanin %d (%d) can be removed.  ", iNode, f, Sfm_ObjFanin(p, iNode, f))
		} else {
			fmt.Printf("Node %d: Fanin %d (%d) can be replaced by divisor %d (%d).   ",
				iNode, f, Sfm_ObjFanin(p, iNode, f), iVar, p.vDivs[iVar])
		}
		kitDsdPrintFromTruth(uTruth, len(p.vDivIds))
		fmt.Printf("\n")
	}
	if iVar == -1 {
		p.Stats.Removes++
	} else {
		p.Stats.Resubs++
	}
	if fSkipUpdate {
		return false
	}
	// update the network
	iFaninNew := iVar
	if iVar != -1 {
		iFaninNew = p.vDivs[iVar]
	}
	Sfm_NtkUpdate(p, iNode, f, iFaninNew, uTruth)
	return true
}

// wrdVecFill mirrors Vec_WrdFill: it resizes the vector to nSize entries,
// all equal to Fill.
func wrdVecFill(a []uint64, nSize int, Fill uint64) []uint64 {
	if cap(a) < nSize {
		a = make([]uint64, nSize)
	}
	a = a[:nSize]
	for i := range a {
		a[i] = Fill
	}
	return a
}

// Performs resubstitution for the node.
func Sfm_NodeResubOne(p *Sfm_Ntk_t, iNode int) bool {
	fSkipUpdate := false
	assert(Sfm_ObjIsNode(p, iNode), "Sfm_ObjIsNode(p, iNode)")
	p.Stats.TryImproves++
	// report init stats
	if p.pPars.VeryVerbose {
		fmt.Printf("%5d : Lev =%3d. Leaf =%3d.  Node =%3d.  Div=%3d.  Fanins = %d.  MFFC = %d\n",
			iNode, Sfm_ObjLevel(p, iNode), 0, len(p.vNodes), len(p.vDivs),
			Sfm_ObjFaninNum(p, iNode), Sfm_ObjMffcSize(p, iNode))
	}
	// collect fanins
	p.vDivIds = p.vDivIds[:0]
	for _, iFanin := range p.Fanins[iNode] {
		p.vDivIds = append(p.vDivIds, Sfm_ObjSatVar(p, iFanin))
	}
	clk := time.Now()
	uTruth := Sfm_ComputeInterpolant2(p)
	p.timeSat += time.Since(clk)
	// analyze outcomes
	if uTruth == SFM_SAT_UNDEC {
		p.Stats.Timeouts++
		return false
	}
	assert(uTruth != SFM_SAT_SAT, "uTruth != SFM_SAT_SAT")
	if uTruth == p.vTruths[iNode] {
		return false
	} else {
		uTruth0 := p.vTruths[iNode]
		//word uTruth0N = ~uTruth0;
		//word uTruthN  = ~uTruth;
		Old := kitTruthLitNum(uTruth0, Sfm_ObjFaninNum(p, iNode), &p.vCover)
		//int OldN = Kit_TruthLitNum((unsigned*)&uTruth0N, Sfm_ObjFaninNum(p, iNode), p->vCover);
		New := kitTruthLitNum(uTruth, Sfm_ObjFaninNum(p, iNode), &p.vCover)
		//int NewN = Kit_TruthLitNum((unsigned*)&uTruthN,  Sfm_ObjFaninNum(p, iNode), p->vCover);
		//if ( Abc_MinInt(New, NewN) > Abc_MinInt(Old, OldN) )
		if New > Old {
			return false
		}
	}
	p.Stats.Improves++
	if fSkipUpdate {
		return false
	}
	// update truth table
	p.vTruths[iNode] = uTruth
	Sfm_TruthToCnf(uTruth, nil, Sfm_ObjFaninNum(p, iNode), &p.vCover, &p.vCnfs[iNode])
	return true
}

func kitTruthLitNum(uTruth uint64, nVars int, vCover *[]uint32) int {
	t := [2]uint32{uint32(uTruth), uint32(uTruth >> 32)}
	return int(kit.Kit_TruthLitNum(t[:], int32(nVars), vCover))
}

func Sfm_NodeResub(p *Sfm_Ntk_t, iNode int) bool {
	p.Stats.NodesTried++
	// prepare SAT solver
	if !Sfm_NtkCreateWindow(p, iNode, p.pPars.VeryVerbose) {
		return false
	}
	if !Sfm_NtkWindowToSolver(p) {
		return false
	}
	// try replacing area critical fanins
	for i := 0; i < Sfm_ObjFaninNum(p, iNode); i++ {
		iFanin := Sfm_ObjFanin(p, iNode, i)
		if Sfm_ObjIsNode(p, iFanin) && Sfm_ObjFanoutNum(p, iFanin) == 1 {
			if Sfm_NodeResubSolve(p, iNode, i, false) {
				return true
			}
		}
	}
	// try removing redundant edges
	if !p.pPars.Area {
		for i := 0; i < Sfm_ObjFaninNum(p, iNode); i++ {
			iFanin := Sfm_ObjFanin(p, iNode, i)
			if !(Sfm_ObjIsNode(p, iFanin) && Sfm_ObjFanoutNum(p, iFanin) == 1) {
				if Sfm_NodeResubSolve(p, iNode, i, true) {
					return true
				}
			}
		}
	}
	// try simplifying local functions
	if p.pPars.UseDcs {
		if Sfm_NodeResubOne(p, iNode) {
			return true
		}
	}
	/*
	   // try replacing area critical fanins while adding two new fanins
	   if ( Sfm_ObjFaninNum(p, iNode) < p->nFaninMax )
	       Abc_ObjForEachFanin( pNode, pFanin, i )
	           if ( !Abc_ObjIsCi(pFanin) && Abc_ObjFanoutNum(pFanin) == 1 )
	           {
	               if ( Abc_NtkMfsSolveSatResub2( p, pNode, i, -1 ) )
	                   return 1;
	           }
	*/
	return false
}

func Sfm_NtkPrint(p *Sfm_Ntk_t) {
	for i := 0; i < p.nObjs; i++ {
		vArray := p.Fanins[i]
		fmt.Printf("Obj %3d : ", i)
		fmt.Printf("Fixed %d  ", boolInt8(p.Fixed[i]))
		fmt.Printf("Empty %d  ", boolInt8(p.Empty != nil && p.Empty[i]))
		fmt.Printf("Truth  ")
		fmt.Printf("%016X", p.vTruths[i])
		fmt.Printf("  ")
		fmt.Println(vArray)
	}
}

// wecSizeUsedLimits mirrors Vec_WecSizeUsedLimits.
func wecSizeUsedLimits(p [][]int, iStart, iStop int) int {
	Counter := 0
	for i := iStart; i < iStop; i++ {
		if len(p[i]) > 0 {
			Counter++
		}
	}
	return Counter
}

// wecSizeSize mirrors Vec_WecSizeSize.
func wecSizeSize(p [][]int) int {
	Counter := 0
	for _, a := range p {
		Counter += len(a)
	}
	return Counter
}

// Sfm_NtkPerform optimizes the network and returns the number of changed
// nodes. A nil pPars uses the parameters stored in the network (defaults).
func Sfm_NtkPerform(p *Sfm_Ntk_t, pPars *Sfm_Par_t) int {
	Counter := 0
	if pPars == nil {
		pPars = &p.Params
	}
	//Sfm_NtkPrint( p );
	start := time.Now()
	if pPars.Verbose {
		nFixed, nEmpty := 0, 0
		for i := range p.Fixed {
			if p.Fixed[i] {
				nFixed++
			}
		}
		for i := range p.Empty {
			if p.Empty[i] {
				nEmpty++
			}
		}
		fmt.Printf("Performing MFS with %d PIs, %d POs, %d nodes (%d flexible, %d fixed, %d empty).\n",
			p.Pis, p.Pos, p.nNodes, p.nNodes-nFixed, nFixed, nEmpty)
	}
	p.pPars = pPars
	Sfm_NtkPrepare(p)
	//    Sfm_ComputeInterpolantCheck( p );
	//    return 0;
	p.Stats.NodesBefore = wecSizeUsedLimits(p.Fanins, Sfm_NtkPiNum(p), len(p.Fanins)-Sfm_NtkPoNum(p))
	p.Stats.EdgesBefore = wecSizeSize(p.Fanins) - Sfm_NtkPoNum(p)
	for i := p.Pis; i+p.Pos < p.nObjs; i++ {
		if Sfm_ObjIsFixed(p, i) {
			continue
		}
		if p.pPars.DepthMax != 0 && Sfm_ObjLevel(p, i) > p.pPars.DepthMax {
			continue
		}
		if Sfm_ObjFaninNum(p, i) < 2 || Sfm_ObjFaninNum(p, i) > 6 {
			continue
		}
		k := 0
		for ; Sfm_NodeResub(p, i); k++ {
			//            Counter++;
			//            break;
		}
		if k > 0 {
			Counter++
		}
		if pPars.NodesMax != 0 && Counter >= pPars.NodesMax {
			break
		}
	}
	p.Stats.NodesAfter = wecSizeUsedLimits(p.Fanins, Sfm_NtkPiNum(p), len(p.Fanins)-Sfm_NtkPoNum(p))
	p.Stats.EdgesAfter = wecSizeSize(p.Fanins) - Sfm_NtkPoNum(p)
	p.timeTotal = time.Since(start)
	if pPars.Verbose {
		Sfm_NtkPrintStats(p)
	}
	//Sfm_NtkPrint( p );
	return Counter
}
