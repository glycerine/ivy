package mfs

import (
	"fmt"

	"github.com/glycerine/abcgo/aig/hop"
	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/misc/util"
)

// Port of mfsResub.c.

// Abc_NtkMfsUpdateNetwork updates the network after resubstitution.
func Abc_NtkMfsUpdateNetwork(p *Mfs_Man_t, pObj *abc.Object, vMfsFanins []*abc.Object, pFunc *hop.Hop_Obj_t) {
	// create the new node
	pObjNew := abc.Abc_NtkCreateNode(abc.Abc_ObjNtk(pObj))
	abc.Abc_ObjSetData(pObjNew, pFunc)
	for _, pFanin := range vMfsFanins {
		abc.Abc_ObjAddFanin(pObjNew, pFanin)
	}
	// replace the old node by the new node
	// update the level of the node
	abc.Abc_NtkUpdate(pObj, pObjNew, p.vLevels)
}

// Abc_NtkMfsPrintResubStats prints resubstitution stats (computes only, as in C).
func Abc_NtkMfsPrintResubStats(p *Mfs_Man_t) {
	nAreaCrits, nAreaExpanse := 0, 0
	nFaninMax := abc.Abc_NtkGetFaninMax(p.pNtk)
	for i := int32(0); i < abc.Abc_NtkObjNumMax(p.pNtk); i++ {
		pNode := abc.Abc_NtkObj(p.pNtk, i)
		if pNode == nil || abc.Abc_ObjIsNode(pNode) == 0 {
			continue
		}
		for k := int32(0); k < abc.Abc_ObjFaninNum(pNode); k++ {
			pFanin := abc.Abc_ObjFanin(pNode, k)
			if abc.Abc_ObjIsCi(pFanin) == 0 && abc.Abc_ObjFanoutNum(pFanin) == 1 {
				nAreaCrits++
				if abc.Abc_ObjFaninNum(pNode) < nFaninMax {
					nAreaExpanse++
				}
			}
		}
	}
	_, _ = nAreaCrits, nAreaExpanse
}

// Abc_NtkMfsTryResubOnce tries resubstitution.
// Returns -1 on timeout, 1 if UNSAT (resub possible), 0 with a stored cex.
func Abc_NtkMfsTryResubOnce(p *Mfs_Man_t, pCands []int32) int32 {
	fVeryVerbose := false
	p.nSatCalls++
	RetValue := sat_solver_solve(p.pSat, pCands, int64(p.pPars.NBTLimit), 0, 0, 0)
	if RetValue == l_False {
		if fVeryVerbose {
			fmt.Printf("U ")
		}
		return 1
	}
	if RetValue != l_True {
		if fVeryVerbose {
			fmt.Printf("T ")
		}
		p.nTimeOuts++
		return -1
	}
	if fVeryVerbose {
		fmt.Printf("S ")
	}
	p.nSatCexes++
	// store the counter-example
	for i, iVar := range p.vProjVarsSat {
		pData := p.vDivCexes[i]
		if sat_solver_var_value(p.pSat, iVar) == 0 { // remove 0s!!!
			assert(util.Abc_InfoHasBit(pData, p.nCexes), "Abc_InfoHasBit(pData, p->nCexes)")
			util.Abc_InfoXorBit(pData, p.nCexes)
		}
	}
	p.nCexes++
	return 0
}

// fillSimInfo is Vec_PtrFillSimInfo( p->vDivCexes, 0, p->nDivWords ).
func fillSimInfo(p *Mfs_Man_t) {
	for _, d := range p.vDivCexes {
		for w := range d[:p.nDivWords] {
			d[w] = ^uint32(0)
		}
	}
}

// Abc_NtkMfsSolveSatResub performs resubstitution for the node.
func Abc_NtkMfsSolveSatResub(p *Mfs_Man_t, pNode *abc.Object, iFanin int32, fOnlyRemove int32, fSkipUpdate int32) int32 {
	fVeryVerbose := false
	var pCandsArr [MFS_FANIN_MAX]int32
	assert(iFanin >= 0, "iFanin >= 0")
	p.nTryRemoves++

	// clean simulation info
	fillSimInfo(p)
	p.nCexes = 0
	nDivs := int32(len(p.vDivs))
	nFanins := abc.Abc_ObjFaninNum(pNode)
	if p.pPars.FVeryVerbose != 0 {
		mffc := int32(0)
		if abc.Abc_ObjFanoutNum(abc.Abc_ObjFanin(pNode, iFanin)) == 1 {
			mffc = abc.Abc_NodeMffcLabel(abc.Abc_ObjFanin(pNode, iFanin), nil)
		}
		fmt.Printf("%5d : Lev =%3d. Leaf =%3d. Node =%3d. Divs =%3d.  Fanin = %4d (%d/%d), MFFC = %d\n",
			pNode.ID, pNode.Level, len(p.vSupp), len(p.vNodes), nDivs-nFanins,
			abc.Abc_ObjFanin(pNode, iFanin).ID, iFanin, nFanins, mffc)
	}

	// try fanins without the critical fanin
	nCands := int32(0)
	p.vMfsFanins = p.vMfsFanins[:0]
	for i := int32(0); i < nFanins; i++ {
		pFanin := abc.Abc_ObjFanin(pNode, i)
		if i == iFanin {
			continue
		}
		p.vMfsFanins = append(p.vMfsFanins, pFanin)
		iVar := nDivs - nFanins + i
		pCandsArr[nCands] = toLitCond(p.vProjVarsSat[iVar], 1)
		nCands++
	}
	pCands := pCandsArr[:]
	RetValue := Abc_NtkMfsTryResubOnce(p, pCands[:nCands])
	if RetValue == -1 {
		return 0
	}
	if RetValue == 1 {
		if p.pPars.FVeryVerbose != 0 {
			fmt.Printf("Node %d: Fanin %d can be removed.\n", pNode.ID, iFanin)
		}
		p.nNodesResub++
		p.nNodesGainedLevel++
		if fSkipUpdate != 0 {
			return 1
		}
		clk := util.Abc_Clock()
		// derive the function
		pFunc := Abc_NtkMfsInterplate(p, pCands[:nCands])
		if pFunc == nil {
			return 0
		}
		// update the network
		Abc_NtkMfsUpdateNetwork(p, pNode, p.vMfsFanins, pFunc)
		p.timeInt += util.Abc_Clock() - clk
		p.nRemoves++
		return 1
	}

	if fOnlyRemove != 0 || p.pPars.FRrOnly != 0 {
		return 0
	}

	p.nTryResubs++
	if fVeryVerbose {
		for i := 0; i < 9; i++ {
			fmt.Printf(" ")
		}
		for i := int32(0); i < nDivs-nFanins; i++ {
			fmt.Printf("%d", i%10)
		}
		for i := int32(0); i < nFanins; i++ {
			if i == iFanin {
				fmt.Printf("*")
			} else {
				fmt.Printf("%c", 'a'+i)
			}
		}
		fmt.Printf("\n")
	}
	iVar := int32(-1)
	for {
		if fVeryVerbose {
			fmt.Printf("%3d: %3d ", p.nCexes, iVar)
			for i := int32(0); i < nDivs; i++ {
				pData := p.vDivCexes[i]
				fmt.Printf("%d", b2i(util.Abc_InfoHasBit(pData, p.nCexes-1)))
			}
			fmt.Printf("\n")
		}

		// find the next divisor to try
		nWords := util.Abc_BitWordNum(p.nCexes)
		assert(nWords <= p.nDivWords, "nWords <= p->nDivWords")
		for iVar = 0; iVar < nDivs-nFanins; iVar++ {
			if p.pPars.FPower != 0 {
				pDiv := p.vDivs[iVar]
				// only accept the divisor if it is "cool"
				if Abc_MfsObjProb(p, pDiv) >= 0.15 {
					continue
				}
			}
			pData := p.vDivCexes[iVar]
			w := int32(0)
			for ; w < nWords; w++ {
				if pData[w] != ^uint32(0) {
					break
				}
			}
			if w == nWords {
				break
			}
		}
		if iVar == nDivs-nFanins {
			return 0
		}

		pCands[nCands] = toLitCond(p.vProjVarsSat[iVar], 1)
		RetValue = Abc_NtkMfsTryResubOnce(p, pCands[:nCands+1])
		if RetValue == -1 {
			return 0
		}
		if RetValue == 1 {
			if p.pPars.FVeryVerbose != 0 {
				fmt.Printf("Node %d: Fanin %d can be replaced by divisor %d.\n", pNode.ID, iFanin, iVar)
			}
			p.nNodesResub++
			p.nNodesGainedLevel++
			if fSkipUpdate != 0 {
				return 1
			}
			clk := util.Abc_Clock()
			// derive the function
			pFunc := Abc_NtkMfsInterplate(p, pCands[:nCands+1])
			if pFunc == nil {
				return 0
			}
			// update the network
			p.vMfsFanins = append(p.vMfsFanins, p.vDivs[iVar])
			Abc_NtkMfsUpdateNetwork(p, pNode, p.vMfsFanins, pFunc)
			p.timeInt += util.Abc_Clock() - clk
			p.nResubs++
			return 1
		}
		if p.nCexes >= p.pPars.NWinMax {
			break
		}
	}
	if p.pPars.FVeryVerbose != 0 {
		fmt.Printf("Node %d: Cannot find replacement for fanin %d.\n", pNode.ID, iFanin)
	}
	return 0
}

// Abc_NtkMfsSolveSatResub2 performs resubstitution for the node
// (two fanins removed, two divisors added).
func Abc_NtkMfsSolveSatResub2(p *Mfs_Man_t, pNode *abc.Object, iFanin, iFanin2 int32) int32 {
	fVeryVerbose := p.pPars.FVeryVerbose != 0 && len(p.vDivs) < 80
	var pCandsArr [MFS_FANIN_MAX]int32
	assert(iFanin >= 0, "iFanin >= 0")
	assert(iFanin2 >= 0 || iFanin2 == -1, "iFanin2 >= 0 || iFanin2 == -1")

	// clean simulation info
	fillSimInfo(p)
	p.nCexes = 0
	nDivs := int32(len(p.vDivs))
	nFanins := abc.Abc_ObjFaninNum(pNode)
	if fVeryVerbose {
		mffc := int32(0)
		if abc.Abc_ObjFanoutNum(abc.Abc_ObjFanin(pNode, iFanin)) == 1 {
			mffc = abc.Abc_NodeMffcLabel(abc.Abc_ObjFanin(pNode, iFanin), nil)
		}
		fmt.Printf("\n")
		fmt.Printf("Node %5d : Level = %2d. Divs = %3d.  Fanins = %d/%d (out of %d). MFFC = %d\n",
			pNode.ID, pNode.Level, nDivs-nFanins, iFanin, iFanin2, nFanins, mffc)
	}

	// try fanins without the critical fanin
	nCands := int32(0)
	p.vMfsFanins = p.vMfsFanins[:0]
	for i := int32(0); i < nFanins; i++ {
		pFanin := abc.Abc_ObjFanin(pNode, i)
		if i == iFanin || i == iFanin2 {
			continue
		}
		p.vMfsFanins = append(p.vMfsFanins, pFanin)
		iVar := nDivs - nFanins + i
		pCandsArr[nCands] = toLitCond(p.vProjVarsSat[iVar], 1)
		nCands++
	}
	pCands := pCandsArr[:]
	RetValue := Abc_NtkMfsTryResubOnce(p, pCands[:nCands])
	if RetValue == -1 {
		return 0
	}
	if RetValue == 1 {
		if fVeryVerbose {
			fmt.Printf("Node %d: Fanins %d/%d can be removed.\n", pNode.ID, iFanin, iFanin2)
		}
		p.nNodesResub++
		p.nNodesGainedLevel++
		clk := util.Abc_Clock()
		// derive the function
		pFunc := Abc_NtkMfsInterplate(p, pCands[:nCands])
		if pFunc == nil {
			return 0
		}
		// update the network
		Abc_NtkMfsUpdateNetwork(p, pNode, p.vMfsFanins, pFunc)
		p.timeInt += util.Abc_Clock() - clk
		return 1
	}

	if fVeryVerbose {
		for i := 0; i < 11; i++ {
			fmt.Printf(" ")
		}
		for i := int32(0); i < nDivs-nFanins; i++ {
			fmt.Printf("%d", i%10)
		}
		for i := int32(0); i < nFanins; i++ {
			if i == iFanin || i == iFanin2 {
				fmt.Printf("*")
			} else {
				fmt.Printf("%c", 'a'+i)
			}
		}
		fmt.Printf("\n")
	}
	iVar, iVar2 := int32(-1), int32(-1)
	for {
		if fVeryVerbose {
			fmt.Printf("%3d: %2d %2d ", p.nCexes, iVar, iVar2)
			for i := int32(0); i < nDivs; i++ {
				pData := p.vDivCexes[i]
				fmt.Printf("%d", b2i(util.Abc_InfoHasBit(pData, p.nCexes-1)))
			}
			fmt.Printf("\n")
		}

		// find the next divisor to try
		nWords := util.Abc_BitWordNum(p.nCexes)
		assert(nWords <= p.nDivWords, "nWords <= p->nDivWords")
		fBreak := false
		for iVar = 1; iVar < nDivs-nFanins; iVar++ {
			pData := p.vDivCexes[iVar]
			if p.pPars.FPower != 0 {
				pDiv := p.vDivs[iVar]
				// only accept the divisor if it is "cool"
				if Abc_MfsObjProb(p, pDiv) >= 0.12 {
					continue
				}
			}
			for iVar2 = 0; iVar2 < iVar; iVar2++ {
				pData2 := p.vDivCexes[iVar2]
				if p.pPars.FPower != 0 {
					pDiv := p.vDivs[iVar2]
					// only accept the divisor if it is "cool"
					if Abc_MfsObjProb(p, pDiv) >= 0.12 {
						continue
					}
				}
				w := int32(0)
				for ; w < nWords; w++ {
					if (pData[w] | pData2[w]) != ^uint32(0) {
						break
					}
				}
				if w == nWords {
					fBreak = true
					break
				}
			}
			if fBreak {
				break
			}
		}
		if iVar == nDivs-nFanins {
			return 0
		}

		pCands[nCands] = toLitCond(p.vProjVarsSat[iVar2], 1)
		pCands[nCands+1] = toLitCond(p.vProjVarsSat[iVar], 1)
		RetValue = Abc_NtkMfsTryResubOnce(p, pCands[:nCands+2])
		if RetValue == -1 {
			return 0
		}
		if RetValue == 1 {
			if fVeryVerbose {
				fmt.Printf("Node %d: Fanins %d/%d can be replaced by divisors %d/%d.\n", pNode.ID, iFanin, iFanin2, iVar, iVar2)
			}
			p.nNodesResub++
			p.nNodesGainedLevel++
			clk := util.Abc_Clock()
			// derive the function
			pFunc := Abc_NtkMfsInterplate(p, pCands[:nCands+2])
			if pFunc == nil {
				return 0
			}
			// update the network
			p.vMfsFanins = append(p.vMfsFanins, p.vDivs[iVar2])
			p.vMfsFanins = append(p.vMfsFanins, p.vDivs[iVar])
			assert(int32(len(p.vMfsFanins)) == nCands+2, "Vec_PtrSize(p->vMfsFanins) == nCands + 2")
			Abc_NtkMfsUpdateNetwork(p, pNode, p.vMfsFanins, pFunc)
			p.timeInt += util.Abc_Clock() - clk
			return 1
		}
		if p.nCexes >= p.pPars.NWinMax {
			break
		}
	}
	return 0
}

// Abc_NtkMfsEdgeSwapEval evaluates the possibility of replacing each fanin.
func Abc_NtkMfsEdgeSwapEval(p *Mfs_Man_t, pNode *abc.Object) int32 {
	for i := int32(0); i < abc.Abc_ObjFaninNum(pNode); i++ {
		Abc_NtkMfsSolveSatResub(p, pNode, i, 0, 1)
	}
	return 0
}

// Abc_NtkMfsEdgePower performs power-aware resubstitution of the fanins.
func Abc_NtkMfsEdgePower(p *Mfs_Man_t, pNode *abc.Object) int32 {
	// try replacing area critical fanins
	for i := int32(0); i < abc.Abc_ObjFaninNum(pNode); i++ {
		pFanin := abc.Abc_ObjFanin(pNode, i)
		if Abc_MfsObjProb(p, pFanin) >= 0.35 {
			if Abc_NtkMfsSolveSatResub(p, pNode, i, 0, 0) != 0 {
				return 1
			}
		} else if Abc_MfsObjProb(p, pFanin) >= 0.25 { // sjang
			if Abc_NtkMfsSolveSatResub(p, pNode, i, 1, 0) != 0 {
				return 1
			}
		}
	}
	return 0
}

// Abc_NtkMfsResubNode performs resubstitution for the node.
func Abc_NtkMfsResubNode(p *Mfs_Man_t, pNode *abc.Object) int32 {
	// try replacing area critical fanins
	for i := int32(0); i < abc.Abc_ObjFaninNum(pNode); i++ {
		pFanin := abc.Abc_ObjFanin(pNode, i)
		if abc.Abc_ObjIsCi(pFanin) == 0 && abc.Abc_ObjFanoutNum(pFanin) == 1 {
			if Abc_NtkMfsSolveSatResub(p, pNode, i, 0, 0) != 0 {
				return 1
			}
		}
	}
	// try removing redundant edges
	if p.pPars.FArea == 0 {
		for i := int32(0); i < abc.Abc_ObjFaninNum(pNode); i++ {
			pFanin := abc.Abc_ObjFanin(pNode, i)
			if abc.Abc_ObjIsCi(pFanin) != 0 || abc.Abc_ObjFanoutNum(pFanin) != 1 {
				if Abc_NtkMfsSolveSatResub(p, pNode, i, 1, 0) != 0 {
					return 1
				}
			}
		}
	}
	if abc.Abc_ObjFaninNum(pNode) == p.nFaninMax {
		return 0
	}
	return 0
}

// Abc_NtkMfsResubNode2 performs resubstitution for the node.
func Abc_NtkMfsResubNode2(p *Mfs_Man_t, pNode *abc.Object) int32 {
	if abc.Abc_ObjFaninNum(pNode) < 2 {
		return 0
	}
	// try replacing one area critical fanin and one other fanin while adding two new fanins
	for i := int32(0); i < abc.Abc_ObjFaninNum(pNode); i++ {
		pFanin := abc.Abc_ObjFanin(pNode, i)
		if abc.Abc_ObjIsCi(pFanin) == 0 && abc.Abc_ObjFanoutNum(pFanin) == 1 {
			// consider second fanin to remove at the same time
			for k := int32(0); k < abc.Abc_ObjFaninNum(pNode); k++ {
				if i != k && Abc_NtkMfsSolveSatResub2(p, pNode, i, k) != 0 {
					return 1
				}
			}
		}
	}
	return 0
}
