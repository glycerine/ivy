package mfs

import (
	"fmt"
	"os"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/hop"
	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/bool/bdc"
	"github.com/glycerine/abcgo/misc/extra"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/cnf"
)

// Port of mfsCore.c.

func Abc_NtkMfsParsDefault(pPars *Mfs_Par_t) {
	*pPars = Mfs_Par_t{}
	pPars.NWinTfoLevs = 2
	pPars.NFanoutsMax = 30
	pPars.NDepthMax = 20
	pPars.NWinMax = 300
	pPars.NGrowthLevel = 0
	pPars.NBTLimit = 5000
	pPars.FRrOnly = 0
	pPars.FResub = 1
	pPars.FArea = 0
	pPars.FMoreEffort = 0
	pPars.FSwapEdge = 0
	pPars.FOneHotness = 0
	pPars.FVerbose = 0
	pPars.FVeryVerbose = 0
}

// abcNtkForEachNode iterates like Abc_NtkForEachNode: the object vector is
// re-read on every step, so nodes created during the loop are visited too.
func abcNtkForEachNode(pNtk *abc.Network, f func(pNode *abc.Object, i int32)) {
	for i := int32(0); i < abc.Abc_NtkObjNumMax(pNtk); i++ {
		pNode := abc.Abc_NtkObj(pNtk, i)
		if pNode == nil || abc.Abc_ObjIsNode(pNode) == 0 {
			continue
		}
		f(pNode, i)
	}
}

// Abc_WinNode prepares the window of the node; returns 1 if something is wrong.
func Abc_WinNode(p *Mfs_Man_t, pNode *abc.Object) int32 {
	p.nNodesTried++
	// prepare data structure for this node
	Mfs_ManClean(p)
	// compute window roots, window support, and window nodes
	p.vRoots = Abc_MfsComputeRoots(pNode, p.pPars.NWinTfoLevs, p.pPars.NFanoutsMax)
	p.vSupp = abc.Abc_NtkNodeSupport(p.pNtk, p.vRoots, int32(len(p.vRoots)))
	p.vNodes = abc.Abc_NtkDfsNodes(p.pNtk, p.vRoots, int32(len(p.vRoots)))
	if p.pPars.NWinMax != 0 && int32(len(p.vNodes)) > p.pPars.NWinMax {
		return 1
	}
	// compute the divisors of the window
	p.vDivs = Abc_MfsComputeDivisors(p, pNode, abc.Abc_ObjRequiredLevel(pNode)-1)
	p.nTotalDivs += int32(len(p.vDivs)) - abc.Abc_ObjFaninNum(pNode)
	// construct AIG for the window
	p.pAigWin = Abc_NtkConstructAig(p, pNode)
	// translate it into CNF
	p.pCnf = cnf.Cnf_DeriveSimple(p.pAigWin, 1+int32(len(p.vDivs)))
	// create the SAT problem
	p.pSat = Abc_MfsCreateSolverResub(p, nil, 0)
	if p.pSat == nil {
		p.nNodesBad++
		return 1
	}
	return 0
}

func Abc_NtkMfsPowerResub(p *Mfs_Man_t, pPars *Mfs_Par_t) {
	pNtk := p.pNtk
	nFaninMax := abc.Abc_NtkGetFaninMax(p.pNtk)

	abcNtkForEachNode(pNtk, func(pNode *abc.Object, k int32) {
		if p.pPars.NDepthMax != 0 && int32(pNode.Level) > p.pPars.NDepthMax {
			return
		}
		if abc.Abc_ObjFaninNum(pNode) < 2 || abc.Abc_ObjFaninNum(pNode) > nFaninMax {
			return
		}
		if Abc_WinNode(p, pNode) != 0 { // something wrong
			return
		}
		// solve the SAT problem
		// try replacing area critical fanins
		for i := int32(0); i < abc.Abc_ObjFaninNum(pNode); i++ {
			pFanin := abc.Abc_ObjFanin(pNode, i)
			if Abc_MfsObjProb(p, pFanin) >= 0.35 && Abc_NtkMfsSolveSatResub(p, pNode, i, 0, 0) != 0 {
				continue
			}
		}
	})

	abcNtkForEachNode(pNtk, func(pNode *abc.Object, k int32) {
		if p.pPars.NDepthMax != 0 && int32(pNode.Level) > p.pPars.NDepthMax {
			return
		}
		if abc.Abc_ObjFaninNum(pNode) < 2 || abc.Abc_ObjFaninNum(pNode) > nFaninMax {
			return
		}
		if Abc_WinNode(p, pNode) != 0 { // something wrong
			return
		}
		for i := int32(0); i < abc.Abc_ObjFaninNum(pNode); i++ {
			pFanin := abc.Abc_ObjFanin(pNode, i)
			if Abc_MfsObjProb(p, pFanin) >= 0.35 && Abc_NtkMfsSolveSatResub(p, pNode, i, 0, 0) != 0 {
				continue
			}
		}
	})

	abcNtkForEachNode(pNtk, func(pNode *abc.Object, k int32) {
		if p.pPars.NDepthMax != 0 && int32(pNode.Level) > p.pPars.NDepthMax {
			return
		}
		if abc.Abc_ObjFaninNum(pNode) < 2 || abc.Abc_ObjFaninNum(pNode) > nFaninMax {
			return
		}
		if Abc_WinNode(p, pNode) != 0 { // something wrong
			return
		}
		for i := int32(0); i < abc.Abc_ObjFaninNum(pNode); i++ {
			pFanin := abc.Abc_ObjFanin(pNode, i)
			if Abc_MfsObjProb(p, pFanin) >= 0.2 && Abc_NtkMfsSolveSatResub(p, pNode, i, 1, 0) != 0 {
				continue
			}
		}
	})
}

func Abc_NtkMfsResub(p *Mfs_Man_t, pNode *abc.Object) int32 {
	p.nNodesTried++
	// prepare data structure for this node
	Mfs_ManClean(p)
	// compute window roots, window support, and window nodes
	clk := util.Abc_Clock()
	p.vRoots = Abc_MfsComputeRoots(pNode, p.pPars.NWinTfoLevs, p.pPars.NFanoutsMax)
	p.vSupp = abc.Abc_NtkNodeSupport(p.pNtk, p.vRoots, int32(len(p.vRoots)))
	p.vNodes = abc.Abc_NtkDfsNodes(p.pNtk, p.vRoots, int32(len(p.vRoots)))
	p.timeWin += util.Abc_Clock() - clk
	if p.pPars.NWinMax != 0 && int32(len(p.vNodes)) > p.pPars.NWinMax {
		p.nMaxDivs++
		return 1
	}
	// compute the divisors of the window
	clk = util.Abc_Clock()
	p.vDivs = Abc_MfsComputeDivisors(p, pNode, abc.Abc_ObjRequiredLevel(pNode)-1)
	p.nTotalDivs += int32(len(p.vDivs)) - abc.Abc_ObjFaninNum(pNode)
	p.timeDiv += util.Abc_Clock() - clk
	// construct AIG for the window
	clk = util.Abc_Clock()
	p.pAigWin = Abc_NtkConstructAig(p, pNode)
	p.timeAig += util.Abc_Clock() - clk
	// translate it into CNF
	clk = util.Abc_Clock()
	p.pCnf = cnf.Cnf_DeriveSimple(p.pAigWin, 1+int32(len(p.vDivs)))
	p.timeCnf += util.Abc_Clock() - clk
	// create the SAT problem
	clk = util.Abc_Clock()
	p.pSat = Abc_MfsCreateSolverResub(p, nil, 0)
	if p.pSat == nil {
		p.nNodesBad++
		return 1
	}
	// solve the SAT problem
	if p.pPars.FPower != 0 {
		Abc_NtkMfsEdgePower(p, pNode)
	} else if p.pPars.FSwapEdge != 0 {
		Abc_NtkMfsEdgeSwapEval(p, pNode)
	} else {
		Abc_NtkMfsResubNode(p, pNode)
		if p.pPars.FMoreEffort != 0 {
			Abc_NtkMfsResubNode2(p, pNode)
		}
	}
	p.timeSat += util.Abc_Clock() - clk
	return 1
}

func Abc_NtkMfsNode(p *Mfs_Man_t, pNode *abc.Object) int32 {
	p.nNodesTried++
	// prepare data structure for this node
	Mfs_ManClean(p)
	// compute window roots, window support, and window nodes
	clk := util.Abc_Clock()
	p.vRoots = Abc_MfsComputeRoots(pNode, p.pPars.NWinTfoLevs, p.pPars.NFanoutsMax)
	p.vSupp = abc.Abc_NtkNodeSupport(p.pNtk, p.vRoots, int32(len(p.vRoots)))
	p.vNodes = abc.Abc_NtkDfsNodes(p.pNtk, p.vRoots, int32(len(p.vRoots)))
	p.timeWin += util.Abc_Clock() - clk
	// construct AIG for the window
	clk = util.Abc_Clock()
	p.pAigWin = Abc_NtkConstructAig(p, pNode)
	p.timeAig += util.Abc_Clock() - clk
	// translate it into CNF
	clk = util.Abc_Clock()
	p.pCnf = cnf.Cnf_DeriveSimple(p.pAigWin, abc.Abc_ObjFaninNum(pNode))
	p.timeCnf += util.Abc_Clock() - clk
	// create the SAT problem
	clk = util.Abc_Clock()
	p.pSat = cnf.Cnf_DataWriteIntoSolver(p.pCnf, 1, 0)
	if p.pSat != nil && p.pPars.FOneHotness != 0 {
		Abc_NtkAddOneHotness(p)
	}
	if p.pSat == nil {
		return 0
	}
	// solve the SAT problem
	RetValue := Abc_NtkMfsSolveSat(p, pNode)
	p.nTotConfLevel += int32(satConflicts(p.pSat))
	p.timeSat += util.Abc_Clock() - clk
	if RetValue == 0 {
		p.nTimeOutsLevel++
		p.nTimeOuts++
		return 0
	}
	// minimize the local function of the node using bi-decomposition
	assert(p.nFanins == abc.Abc_ObjFaninNum(pNode), "p->nFanins == Abc_ObjFaninNum(pNode)")
	dProb := float32(-1.0)
	if p.pPars.FPower != 0 {
		dProb = util.Abc_Int2Float(vec.Vec_IntEntry(p.vProbs, pNode.ID))
	}
	pHop := abc.Abc_ObjNtk(pNode).FunctionManager.(*hop.Hop_Man_t)
	pObj := NodeIfNodeResyn(p.pManDec, pHop, pNode.Data.(*hop.Hop_Obj_t), p.nFanins, p.vTruth, p.uCare[:], dProb)
	nGain := hop.Hop_DagSize(pNode.Data.(*hop.Hop_Obj_t)) - hop.Hop_DagSize(pObj)
	if nGain >= 0 {
		p.nNodesDec++
		p.nNodesGained += nGain
		p.nNodesGainedLevel += nGain
		abc.Abc_ObjSetData(pNode, pObj)
	}
	return 1
}

func Abc_NtkMfs(pNtk *abc.Network, pPars *Mfs_Par_t) int32 {
	var pDecPars bdc.Bdc_Par_t
	clk := util.Abc_Clock()
	nTotalNodesBeg := abc.Abc_NtkNodeNum(pNtk)
	nTotalEdgesBeg := abc.Abc_NtkGetTotalFanins(pNtk)

	assert(abc.Abc_NtkIsLogic(pNtk) != 0, "Abc_NtkIsLogic(pNtk)")
	nFaninMax := abc.Abc_NtkGetFaninMax(pNtk)
	if pPars.FResub != 0 {
		if nFaninMax > 8 {
			fmt.Printf("Nodes with more than %d fanins will not be processed.\n", 8)
			nFaninMax = 8
		}
	} else {
		if nFaninMax > MFS_FANIN_MAX {
			fmt.Printf("Nodes with more than %d fanins will not be processed.\n", MFS_FANIN_MAX)
			nFaninMax = MFS_FANIN_MAX
		}
	}
	// convert into the AIG
	if abc.Abc_NtkToAig(pNtk) == 0 {
		fmt.Fprintf(os.Stdout, "Converting to AIGs has failed.\n")
		return 0
	}
	assert(abc.Abc_NtkHasAig(pNtk) != 0, "Abc_NtkHasAig(pNtk)")

	// start the manager
	p := Mfs_ManAlloc(pPars)
	p.pNtk = pNtk
	p.nFaninMax = nFaninMax

	// precomputer power-aware metrics
	if pPars.FPower != 0 {
		if pPars.FResub != 0 {
			p.vProbs = NtkPowerEstimate(pNtk, 0)
		} else {
			p.vProbs = NtkPowerEstimate(pNtk, 1)
		}
		p.TotalSwitchingBeg = NtkMfsTotalSwitching(pNtk)
	}

	if pNtk.Excare != nil {
		if abc.Abc_NtkPiNum(pNtk.Excare) != abc.Abc_NtkCiNum(pNtk) {
			fmt.Printf("The PI count of careset (%d) and logic network (%d) differ. Careset is not used.\n",
				abc.Abc_NtkPiNum(pNtk.Excare), abc.Abc_NtkCiNum(pNtk))
		} else {
			pTemp := NtkStrash(pNtk.Excare, 0, 0, 0)
			p.pCare = abc.Abc_NtkToDar(pTemp, 0, 0)
			abc.Abc_NtkDelete(pTemp)
			p.vSuppsInv = aig.Aig_ManSupportsInverse(p.pCare)
		}
	}
	if p.pCare != nil {
		fmt.Printf("Performing optimization with %d external care clauses.\n", aig.Aig_ManCoNum(p.pCare))
	}
	// prepare the BDC manager
	if pPars.FResub == 0 {
		pDecPars.NVarsMax = max(3, nFaninMax)
		pDecPars.FVerbose = pPars.FVerbose
		p.vTruth = vec.Vec_IntAlloc(0)
		p.pManDec = bdc.Bdc_ManAlloc(&pDecPars)
	}

	// label the register outputs
	if p.pCare != nil {
		p.ciIndex = make([]int32, abc.Abc_NtkObjNumMax(pNtk))
		for i := int32(0); i < abc.Abc_NtkCiNum(pNtk); i++ {
			p.ciIndex[abc.Abc_NtkCi(pNtk, i).ID] = i
		}
	}

	// compute levels
	abc.Abc_NtkLevel(pNtk)
	abc.Abc_NtkStartReverseLevels(pNtk, pPars.NGrowthLevel)

	// compute don't-cares for each node
	nNodes := 0
	p.nTotalNodesBeg = nTotalNodesBeg
	p.nTotalEdgesBeg = nTotalEdgesBeg
	if pPars.FResub != 0 {
		if pPars.FPower != 0 {
			Abc_NtkMfsPowerResub(p, pPars)
		} else {
			pProgress := extra.Extra_ProgressBarStart(os.Stdout, abc.Abc_NtkObjNumMax(pNtk))
			abcNtkForEachNode(pNtk, func(pObj *abc.Object, i int32) {
				if p.pPars.NDepthMax != 0 && int32(pObj.Level) > p.pPars.NDepthMax {
					return
				}
				if abc.Abc_ObjFaninNum(pObj) < 2 || abc.Abc_ObjFaninNum(pObj) > nFaninMax {
					return
				}
				if p.pPars.FVeryVerbose == 0 {
					extra.Extra_ProgressBarUpdate(pProgress, i, "")
				}
				if pPars.FResub != 0 {
					Abc_NtkMfsResub(p, pObj)
				} else {
					Abc_NtkMfsNode(p, pObj)
				}
			})
			extra.Extra_ProgressBarStop(pProgress)
		}
	} else {
		pProgress := extra.Extra_ProgressBarStart(os.Stdout, abc.Abc_NtkNodeNum(pNtk))
		vLevels := abc.Abc_NtkLevelize(pNtk)
		for k := 1; k < len(vLevels); k++ {
			vNodes := vLevels[k]
			if p.pPars.FVeryVerbose == 0 {
				extra.Extra_ProgressBarUpdate(pProgress, int32(nNodes), "")
			}
			p.nNodesGainedLevel = 0
			p.nTotConfLevel = 0
			p.nTimeOutsLevel = 0
			for _, pObj := range vNodes {
				if p.pPars.NDepthMax != 0 && int32(pObj.Level) > p.pPars.NDepthMax {
					break
				}
				if abc.Abc_ObjFaninNum(pObj) < 2 || abc.Abc_ObjFaninNum(pObj) > nFaninMax {
					continue
				}
				if pPars.FResub != 0 {
					Abc_NtkMfsResub(p, pObj)
				} else {
					Abc_NtkMfsNode(p, pObj)
				}
			}
			nNodes += len(vNodes)
		}
		extra.Extra_ProgressBarStop(pProgress)
	}
	abc.Abc_NtkStopReverseLevels(pNtk)

	p.nTotalNodesEnd = abc.Abc_NtkNodeNum(pNtk)
	p.nTotalEdgesEnd = abc.Abc_NtkGetTotalFanins(pNtk)

	// undo labesl
	if p.pCare != nil {
		p.ciIndex = nil
	}

	if pPars.FPower != 0 {
		p.TotalSwitchingEnd = NtkMfsTotalSwitching(pNtk)
	}

	// free the manager
	p.timeTotal = util.Abc_Clock() - clk
	Mfs_ManStop(p)
	return 1
}
