package sfm

// Port of src/opt/sfm/sfmWin.c.

import (
	"fmt"
	"time"
)

// Returns the MFFC size.
func Sfm_ObjRef_rec(p *Sfm_Ntk_t, iObj int) int {
	if Sfm_ObjIsPi(p, iObj) {
		return 0
	}
	assert(Sfm_ObjIsNode(p, iObj), "Sfm_ObjIsNode(p, iObj)")
	Value := Sfm_ObjRefIncrement(p, iObj)
	if Value > 1 {
		return 0
	}
	assert(Value == 1, "Value == 1")
	Count := 1
	for _, iFanin := range p.Fanins[iObj] {
		Count += Sfm_ObjRef_rec(p, iFanin)
	}
	return Count
}
func Sfm_ObjRef(p *Sfm_Ntk_t, iObj int) int {
	Count := 1
	for _, iFanin := range p.Fanins[iObj] {
		Count += Sfm_ObjRef_rec(p, iFanin)
	}
	return Count
}
func Sfm_ObjDeref_rec(p *Sfm_Ntk_t, iObj int) int {
	if Sfm_ObjIsPi(p, iObj) {
		return 0
	}
	assert(Sfm_ObjIsNode(p, iObj), "Sfm_ObjIsNode(p, iObj)")
	Value := Sfm_ObjRefDecrement(p, iObj)
	if Value > 0 {
		return 0
	}
	assert(Value == 0, "Value == 0")
	Count := 1
	for _, iFanin := range p.Fanins[iObj] {
		Count += Sfm_ObjDeref_rec(p, iFanin)
	}
	return Count
}
func Sfm_ObjDeref(p *Sfm_Ntk_t, iObj int) int {
	Count := 1
	for _, iFanin := range p.Fanins[iObj] {
		Count += Sfm_ObjDeref_rec(p, iFanin)
	}
	return Count
}
func Sfm_ObjMffcSize(p *Sfm_Ntk_t, iObj int) int {
	if Sfm_ObjIsPi(p, iObj) {
		return 0
	}
	if Sfm_ObjFanoutNum(p, iObj) != 1 {
		return 0
	}
	assert(Sfm_ObjIsNode(p, iObj), "Sfm_ObjIsNode( p, iObj )")
	Count1 := Sfm_ObjDeref(p, iObj)
	Count2 := Sfm_ObjRef(p, iObj)
	assert(Count1 == Count2, "Count1 == Count2")
	return Count1
}

// Working with traversal IDs.
func Sfm_NtkIncrementTravId(p *Sfm_Ntk_t)              { p.nTravIds++ }
func Sfm_ObjSetTravIdCurrent(p *Sfm_Ntk_t, Id int)     { p.vTravIds[Id] = p.nTravIds }
func Sfm_ObjIsTravIdCurrent(p *Sfm_Ntk_t, Id int) bool { return p.vTravIds[Id] == p.nTravIds }
func Sfm_ObjIsTravIdPrevious(p *Sfm_Ntk_t, Id int) bool {
	return p.vTravIds[Id] == p.nTravIds-1
}
func Sfm_NtkIncrementTravId2(p *Sfm_Ntk_t)          { p.nTravIds2++ }
func Sfm_ObjSetTravIdCurrent2(p *Sfm_Ntk_t, Id int) { p.vTravIds2[Id] = p.nTravIds2 }
func Sfm_ObjIsTravIdCurrent2(p *Sfm_Ntk_t, Id int) bool {
	return p.vTravIds2[Id] == p.nTravIds2
}

// Collects used internal nodes in a topological order. Additionally considers
// objects in groups as a single object and collects them in a topological
// order together as single entity.
func Sfm_NtkDfs_rec(p *Sfm_Ntk_t, iNode int, vNodes *[]int, vGroups [][]int, vGroupMap []int, vBoxesLeft *[]int) {
	if Sfm_ObjIsPi(p, iNode) {
		return
	}
	if Sfm_ObjIsTravIdCurrent(p, iNode) {
		return
	}
	if vGroupMap[iNode] >= 0 {
		iGroup := vGroupMap[iNode] >> 1
		vGroup := vGroups[iGroup]
		for _, iNode := range vGroup {
			assert(Sfm_ObjIsNode(p, iNode), "Sfm_ObjIsNode(p, iNode)")
		}
		for _, iNode := range vGroup {
			Sfm_ObjSetTravIdCurrent(p, iNode)
		}
		for _, iNode := range vGroup {
			for _, iFanin := range p.Fanins[iNode] {
				Sfm_NtkDfs_rec(p, iFanin, vNodes, vGroups, vGroupMap, vBoxesLeft)
			}
		}
		*vNodes = append(*vNodes, vGroup...)
		*vBoxesLeft = append(*vBoxesLeft, iGroup)
	} else {
		Sfm_ObjSetTravIdCurrent(p, iNode)
		for _, iFanin := range p.Fanins[iNode] {
			Sfm_NtkDfs_rec(p, iFanin, vNodes, vGroups, vGroupMap, vBoxesLeft)
		}
		*vNodes = append(*vNodes, iNode)
	}
}

// Sfm_NtkDfs returns the used internal nodes in topological order and the
// indexes of the groups (boxes) that were collected. vGroupMap entries
// encode group<<1|phase, or -1 for objects outside any group.
func Sfm_NtkDfs(p *Sfm_Ntk_t, vGroups [][]int, vGroupMap []int, fAllBoxes bool) ([]int, []int) {
	var vBoxesLeft []int
	vNodes := make([]int, 0, p.nObjs)
	Sfm_NtkIncrementTravId(p)
	if fAllBoxes {
		for _, vGroup := range vGroups {
			Sfm_NtkDfs_rec(p, vGroup[0], &vNodes, vGroups, vGroupMap, &vBoxesLeft)
		}
	}
	for i := len(p.Fanins) - p.Pos; i < len(p.Fanins); i++ {
		Sfm_NtkDfs_rec(p, Sfm_ObjFanin(p, i, 0), &vNodes, vGroups, vGroupMap, &vBoxesLeft)
	}
	return vNodes, vBoxesLeft
}

// Check if this fanout overlaps with TFI cone of the node.
func Sfm_NtkCheckOverlap_rec(p *Sfm_Ntk_t, iThis, iNode int) bool {
	if Sfm_ObjIsTravIdCurrent2(p, iThis) || iThis == iNode {
		return false
	}
	//    if ( Sfm_ObjIsTravIdCurrent(p, iThis) )
	if Sfm_ObjIsTravIdPrevious(p, iThis) {
		return true
	}
	Sfm_ObjSetTravIdCurrent2(p, iThis)
	for _, iFanin := range p.Fanins[iThis] {
		if Sfm_NtkCheckOverlap_rec(p, iFanin, iNode) {
			return true
		}
	}
	return false
}
func Sfm_NtkCheckOverlap(p *Sfm_Ntk_t, iFan, iNode int) bool {
	Sfm_NtkIncrementTravId2(p)
	return Sfm_NtkCheckOverlap_rec(p, iFan, iNode)
}

// Recursively collects roots of the window.
func Sfm_NtkCheckRoot(p *Sfm_Ntk_t, iNode int, nLevelMax int) bool {
	// the node is the root if one of the following is true:
	// (1) the node has more than fanouts than the limit or has no fanouts (should not happen in general)
	if Sfm_ObjFanoutNum(p, iNode) == 0 || Sfm_ObjFanoutNum(p, iNode) > p.pPars.FanoutMax {
		return true
	}
	// (2) the node has CO fanouts
	// (3) the node has fanouts above the cutoff level
	for _, iFanout := range p.Fanouts[iNode] {
		if Sfm_ObjIsPo(p, iFanout) || Sfm_ObjLevel(p, iFanout) > nLevelMax { //|| !Sfm_NtkCheckOverlap(p, iFanout, iNode) )
			return true
		}
	}
	return false
}
func Sfm_NtkComputeRoots_rec(p *Sfm_Ntk_t, iNode int, nLevelMax int, vRoots, vTfo *[]int) {
	assert(Sfm_ObjIsNode(p, iNode), "Sfm_ObjIsNode(p, iNode)")
	if Sfm_ObjIsTravIdCurrent(p, iNode) {
		return
	}
	Sfm_ObjSetTravIdCurrent(p, iNode)
	if iNode != p.iPivotNode {
		*vTfo = append(*vTfo, iNode)
	}
	// check if the node should be the root
	if Sfm_NtkCheckRoot(p, iNode, nLevelMax) {
		*vRoots = append(*vRoots, iNode)
	} else { // if not, explore its fanouts
		for _, iFanout := range p.Fanouts[iNode] {
			Sfm_NtkComputeRoots_rec(p, iFanout, nLevelMax, vRoots, vTfo)
		}
	}
}

// Collects divisors of the node.
func Sfm_NtkAddDivisors(p *Sfm_Ntk_t, iNode int, nLevelMax int) {
	for i, iFanout := range p.Fanouts[iNode] {
		// skip some of the fanouts if the number is large
		if p.pPars.FanoutMax != 0 && i > p.pPars.FanoutMax {
			return
		}
		// skip TFI nodes, PO nodes, or nodes with high logic level
		if Sfm_ObjIsTravIdCurrent(p, iFanout) || Sfm_ObjIsPo(p, iFanout) || Sfm_ObjLevel(p, iFanout) > nLevelMax {
			continue
		}
		// handle single-input nodes
		if Sfm_ObjFaninNum(p, iFanout) == 1 {
			p.vDivs = append(p.vDivs, iFanout)
		} else if !Sfm_ObjIsTravIdCurrent2(p, iFanout) {
			// visit node for the first time
			assert(Sfm_ObjFaninNum(p, iFanout) > 1, "Sfm_ObjFaninNum(p, iFanout) > 1")
			Sfm_ObjSetTravIdCurrent2(p, iFanout)
			Sfm_ObjResetFaninCount(p, iFanout)
		} else if Sfm_ObjUpdateFaninCount(p, iFanout) == 0 {
			// visit node again
			p.vDivs = append(p.vDivs, iFanout)
		}
	}
}

// Fixed object is useful when it has a non-fixed fanout.
func Sfm_ObjIsUseful(p *Sfm_Ntk_t, iNode int) bool {
	if !Sfm_ObjIsFixed(p, iNode) {
		return true
	}
	for _, iFanout := range p.Fanouts[iNode] {
		if !Sfm_ObjIsFixed(p, iFanout) {
			return true
		}
	}
	return false
}

// Computes structural window.
func Sfm_NtkCollectTfi_rec(p *Sfm_Ntk_t, iNode int, vNodes *[]int) bool {
	if Sfm_ObjIsTravIdCurrent(p, iNode) {
		return false
	}
	Sfm_ObjSetTravIdCurrent(p, iNode)
	for _, iFanin := range p.Fanins[iNode] {
		if Sfm_NtkCollectTfi_rec(p, iFanin, vNodes) {
			return true
		}
	}
	*vNodes = append(*vNodes, iNode)
	return p.pPars.WinSizeMax != 0 && (len(*vNodes) > p.pPars.WinSizeMax)
}
func Sfm_NtkCreateWindow(p *Sfm_Ntk_t, iNode int, fVerbose bool) bool {
	clkWin := time.Now()
	assert(Sfm_ObjIsNode(p, iNode), "Sfm_ObjIsNode( p, iNode )")
	p.iPivotNode = iNode
	p.vNodes = p.vNodes[:0] // internal
	p.vDivs = p.vDivs[:0]   // divisors
	p.vRoots = p.vRoots[:0] // roots
	p.vTfo = p.vTfo[:0]     // roots
	p.vOrder = p.vOrder[:0] // variable order
	// collect transitive fanin
	Sfm_NtkIncrementTravId(p)
	if Sfm_NtkCollectTfi_rec(p, iNode, &p.vNodes) {
		p.Stats.MaxDivs++
		p.timeWin += time.Since(clkWin)
		return false
	}
	// create divisors
	clkDiv := time.Now()
	p.vDivs = p.vDivs[:0]
	p.vDivs = append(p.vDivs, p.vNodes...)
	p.vDivs = p.vDivs[:len(p.vDivs)-1]
	// add non-topological divisors
	if len(p.vDivs) < p.pPars.WinSizeMax+0 {
		Sfm_NtkIncrementTravId2(p)
		// Vec_IntForEachEntry re-reads the size, so divisors added by
		// Sfm_NtkAddDivisors are visited too.
		for i := 0; i < len(p.vDivs); i++ {
			iTemp := p.vDivs[i]
			if len(p.vDivs) < p.pPars.WinSizeMax+0 {
				//                Sfm_NtkAddDivisors( p, iTemp, Sfm_ObjLevel(p, iNode) - 1 );
				Sfm_NtkAddDivisors(p, iTemp, p.LevelMax-Sfm_ObjLevelR(p, iNode))
			}
		}
	}
	if len(p.vDivs) > p.pPars.WinSizeMax {
		p.vDivs = p.vDivs[:p.pPars.WinSizeMax]
	}
	assert(len(p.vDivs) <= p.pPars.WinSizeMax, "Vec_IntSize(p->vDivs) <= p->pPars->nWinSizeMax")
	if len(p.vDivs) == p.pPars.WinSizeMax {
		p.Stats.MaxDivs++
	}
	// remove node/fanins from divisors
	// mark fanins
	Sfm_NtkIncrementTravId2(p)
	Sfm_ObjSetTravIdCurrent2(p, iNode)
	for _, iTemp := range p.Fanins[iNode] {
		Sfm_ObjSetTravIdCurrent2(p, iTemp)
	}
	// compact divisors
	k := 0
	for _, iTemp := range p.vDivs {
		if !Sfm_ObjIsTravIdCurrent2(p, iTemp) && Sfm_ObjIsUseful(p, iTemp) {
			p.vDivs[k] = iTemp
			k++
		}
	}
	p.vDivs = p.vDivs[:k]
	assert(len(p.vDivs) <= p.pPars.WinSizeMax, "Vec_IntSize(p->vDivs) <= p->pPars->nWinSizeMax")
	elapsedDiv := time.Since(clkDiv)
	p.timeDiv += elapsedDiv
	p.Stats.TotalDivs += len(p.vDivs)
	// collect TFO and window roots
	if p.pPars.TfoLevMax > 0 && !Sfm_NtkCheckRoot(p, iNode, Sfm_ObjLevel(p, iNode)+p.pPars.TfoLevMax) {
		// explore transitive fanout
		Sfm_NtkIncrementTravId(p)
		Sfm_NtkComputeRoots_rec(p, iNode, Sfm_ObjLevel(p, iNode)+p.pPars.TfoLevMax, &p.vRoots, &p.vTfo)
		assert(len(p.vRoots) > 0, "Vec_IntSize(p->vRoots) > 0")
		assert(len(p.vTfo) > 0, "Vec_IntSize(p->vTfo) > 0")
		// compute new leaves and nodes
		Sfm_NtkIncrementTravId(p)
		for _, iTemp := range p.vRoots {
			if Sfm_NtkCollectTfi_rec(p, iTemp, &p.vOrder) {
				p.vRoots = p.vRoots[:0]
				p.vTfo = p.vTfo[:0]
				p.vOrder = p.vOrder[:0]
				break
			}
		}
		if len(p.vRoots) > 0 {
			for _, iTemp := range p.vTfo {
				if Sfm_NtkCollectTfi_rec(p, iTemp, &p.vOrder) {
					p.vRoots = p.vRoots[:0]
					p.vTfo = p.vTfo[:0]
					p.vOrder = p.vOrder[:0]
					break
				}
			}
		}
		if len(p.vRoots) > 0 {
			for _, iTemp := range p.vDivs {
				if Sfm_NtkCollectTfi_rec(p, iTemp, &p.vOrder) {
					p.vRoots = p.vRoots[:0]
					p.vTfo = p.vTfo[:0]
					p.vOrder = p.vOrder[:0]
					break
				}
			}
		}
	}
	if len(p.vOrder) == 0 {
		Temp := p.pPars.WinSizeMax
		p.pPars.WinSizeMax = 0
		Sfm_NtkIncrementTravId(p)
		Sfm_NtkCollectTfi_rec(p, iNode, &p.vOrder)
		for _, iTemp := range p.vDivs {
			Sfm_NtkCollectTfi_rec(p, iTemp, &p.vOrder)
		}
		p.pPars.WinSizeMax = Temp
	}
	// statistics
	p.timeWin += time.Since(clkWin) - elapsedDiv
	if !fVerbose {
		return true
	}
	// print stats about the window
	fmt.Printf("%6d : ", iNode)
	fmt.Printf("Leaves = %5d. ", 0)
	fmt.Printf("Nodes = %5d. ", len(p.vNodes))
	fmt.Printf("Roots = %5d. ", len(p.vRoots))
	fmt.Printf("Divs = %5d. ", len(p.vDivs))
	fmt.Printf("\n")
	return true
}
func Sfm_NtkWindowTest(p *Sfm_Ntk_t, iNode int) {
	for i := p.Pis; i+p.Pos < len(p.Fanins); i++ {
		Sfm_NtkCreateWindow(p, i, true)
	}
}
