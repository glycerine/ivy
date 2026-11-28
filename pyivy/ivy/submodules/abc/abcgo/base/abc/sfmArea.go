package abc

// Line-by-line port of src/opt/sfm/sfmArea.c (Abc_NtkChangePerform, the
// polarity optimization run after "mfs3 -a"), working on the mapped
// Abc_Ntk_t as in C.

import (
	"fmt"

	"github.com/glycerine/abcgo/map/mio"
	"github.com/glycerine/abcgo/map/scl"
	"github.com/glycerine/abcgo/misc/extra"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/opt/dau"
)

func Abc_NtkPrecomputeCellPairs(pCells []mio.Mio_Cell2_t, nCells int32) []int32 {
	vInfo := make([]int32, 0, 1000)
	var iBestArea, tCur, iThis uint64
	var pPerm [7][]int32
	var nPerms [7]int32
	var Perm [7]int32
	var iBestCell, iBestPerm, iBestDiff int32
	Count := 0
	iGate1, iGate2 := int32(-1), int32(-1)

	for i := int32(1); i <= 6; i++ {
		pPerm[i] = extra.Extra_PermSchedule(i)
	}
	for i := int32(1); i <= 6; i++ {
		nPerms[i] = extra.Extra_Factorial(i)
	}

	for i := int32(2); i < nCells; i++ {
		nFanins := pCells[i].Fanins()
		for n := int32(0); n <= nFanins; n++ {
			// get the truth table
			if n == nFanins {
				iThis = ^pCells[i].Truth()
			} else {
				iThis = util.Abc_Tt6Flip(pCells[i].Truth(), n)
			}
			// init the comparison
			iBestArea = ^uint64(0)
			iBestCell, iBestPerm, iBestDiff = -1, -1, -1
			// iterate through cells
			for k := int32(2); k < nCells; k++ {
				if nFanins != pCells[k].Fanins() {
					continue
				}
				if i != k && pCells[i].Truth() == pCells[k].Truth() {
					iGate1 = i
					iGate2 = k
					Count++
					continue
				}
				// set unit permutation
				for v := int32(0); v < nFanins; v++ {
					Perm[v] = v
				}
				// go through all permutation of Cell[k]
				tCur = pCells[k].Truth()
				for p := int32(0); p < nPerms[nFanins]; p++ {
					if iThis == tCur && iBestArea > pCells[k].AreaW {
						iBestArea = pCells[k].AreaW
						iBestCell = k
						iBestPerm = 0
						for v := int32(0); v < nFanins; v++ {
							iBestPerm |= v << (Perm[v] << 2)
						}
						if pCells[i].AreaW >= pCells[k].AreaW {
							iBestDiff = int32(pCells[i].AreaW - pCells[k].AreaW)
						} else {
							iBestDiff = -int32(pCells[k].AreaW - pCells[i].AreaW)
						}
					}
					if nPerms[nFanins] == 1 {
						continue
					}
					// update
					tCur = util.Abc_Tt6SwapAdjacent(tCur, pPerm[nFanins][p])
					s := pPerm[nFanins][p]
					Perm[s], Perm[s+1] = Perm[s+1], Perm[s]
				}
				assert(tCur == pCells[k].Truth(), "tCur == pCells[k].uTruth")
			}
			vInfo = append(vInfo, iBestCell, iBestPerm, iBestDiff)
		}
	}

	if Count != 0 {
		fmt.Printf("In this library, %d cell pairs have equal functions (for example, %s and %s).\n", Count/2, pCells[iGate1].Name(), pCells[iGate2].Name())
	}
	return vInfo
}
func Abc_NtkPrecomputeFirsts(pCells []mio.Mio_Cell2_t, nCells int32) []int32 {
	Index := int32(0)
	vFirst := []int32{-1, -1}
	for i := int32(2); i < nCells; i++ {
		vFirst = append(vFirst, Index)
		Index += 3 * (pCells[i].Fanins() + 1)
	}
	assert(nCells == int32(len(vFirst)), "nCells == Vec_IntSize(vFirst)")
	return vFirst
}
func Abc_NtkPrecomputePrint(pCells []mio.Mio_Cell2_t, nCells int32, vInfo []int32) int32 {
	Index, nRecUsed := int32(0), int32(0)
	for i := int32(2); i < nCells; i++ {
		nFanins := pCells[i].Fanins()
		fmt.Printf("%3d : %8s   Fanins = %d   ", i, pCells[i].Name(), nFanins)
		dau.Dau_DsdPrintFromTruth([]uint64{pCells[i].Truth()}, nFanins)
		for n := int32(0); n <= nFanins; n, Index = n+1, Index+3 {
			iCellA := vInfo[Index+0]
			iPerm := vInfo[Index+1]
			Diff := vInfo[Index+2]
			if iCellA == -1 {
				continue
			}
			fmt.Printf("%d : {", n)
			for v := int32(0); v < nFanins; v++ {
				fmt.Printf(" %d ", (iPerm>>(v<<2))&15)
			}
			fmt.Printf("}  Index = %d  ", Index)

			fmt.Printf("Gain = %6.2f  ", scl.Scl_Int2Flt(Diff))
			dau.Dau_DsdPrintFromTruth([]uint64{pCells[iCellA].Truth()}, pCells[iCellA].Fanins())
			nRecUsed++
		}
	}
	return nRecUsed
}

func Abc_NodeCheckFanoutHasFanin(pNode, pFanin *Object) int32 {
	for i, pThis := range pNode.fanouts {
		if Abc_NodeFindFanin(pThis, pFanin) >= 0 {
			return int32(i)
		}
	}
	return -1
}

// Abc_ObjHasDupFanins evaluates changes.
func Abc_ObjHasDupFanins(pObj *Object) int32 {
	Limit := len(pObj.fanins)
	for i := 0; i < Limit; i++ {
		for k := i + 1; k < Limit; k++ {
			if pObj.fanins[i].Object == pObj.fanins[k].Object {
				return 1
			}
		}
	}
	return 0
}
func Abc_ObjHasDupFanouts(pObj *Object) int32 {
	Limit := len(pObj.fanouts)
	for i := 0; i < Limit; i++ {
		for k := i + 1; k < Limit; k++ {
			if pObj.fanouts[i] == pObj.fanouts[k] {
				return 1
			}
		}
	}
	return 0
}

// sfmAreaNodeIsInv and sfmAreaNodeIsBuf are Abc_NodeIsInv and
// Abc_NodeIsBuf on a mapped network: the node's gate is the current
// library's inverter (buffer).
func sfmAreaNodeIsInv(pNode *Object) int32 {
	if Abc_ObjFaninNum(pNode) != 1 {
		return 0
	}
	pLib, _ := mioCurrentLibGen()
	g, _ := pNode.Data.(*mio.Mio_Gate_t)
	return sfmB2I(g == mio.Mio_LibraryReadInv(pLib))
}
func sfmAreaNodeIsBuf(pNode *Object) int32 {
	if Abc_ObjFaninNum(pNode) != 1 {
		return 0
	}
	pLib, _ := mioCurrentLibGen()
	g, _ := pNode.Data.(*mio.Mio_Gate_t)
	return sfmB2I(g == mio.Mio_LibraryReadBuf(pLib))
}
func mioCurrentLibGen() (*mio.Mio_Library_t, bool) {
	pLib, ok := sfmLibGen().(*mio.Mio_Library_t)
	return pLib, ok
}

func Abc_ObjChangeEval(pObj *Object, vInfo []int32, vFirst []int32, InvArea int32, pfUseInv *int32) int32 {
	iNodeCell := mio.Mio_GateReadCell(sfmMioGate(pObj))
	pNodeInfo := vInfo[vFirst[iNodeCell]:]
	fNeedInv, Gain := int32(0), int32(0)
	iFanin := Abc_ObjFaninNum(pObj)
	fUseInv := sfmAreaNodeIsInv(pObj)
	assert(iFanin > 0, "iFanin > 0")
	*pfUseInv = 0
	if pNodeInfo[3*iFanin] == -1 {
		return 0
	}
	if fUseInv != 0 {
		Gain = InvArea
	} else {
		Gain = pNodeInfo[3*iFanin+2]
	}
	for _, pNext := range pObj.fanouts {
		if fUseInv != 0 && Abc_NodeFindFanin(pNext, Abc_ObjFanin0(pObj)) >= 0 {
			return 0
		}
		if Abc_ObjHasDupFanins(pNext) != 0 {
			return 0
		}
		if Abc_ObjIsNode(pNext) == 0 || sfmAreaNodeIsBuf(pNext) != 0 {
			fNeedInv = 1
			continue
		}
		if sfmAreaNodeIsInv(pNext) != 0 {
			if Abc_NodeCheckFanoutHasFanin(pNext, pObj) >= 0 {
				return 0
			}
			Gain += InvArea
			continue
		}
		iFanCell := mio.Mio_GateReadCell(sfmMioGate(pNext))
		pFanInfo := vInfo[vFirst[iFanCell]:]
		iFanin = Abc_NodeFindFanin(pNext, pObj)
		if pFanInfo[3*iFanin] == -1 {
			fNeedInv = 1
			continue
		}
		Gain += pFanInfo[3*iFanin+2]
	}
	if fNeedInv != 0 {
		Gain -= InvArea
	}
	*pfUseInv = fNeedInv
	return Gain
}
func Abc_ObjChangeUpdate(pObj *Object, iFanin int32, pCells []mio.Mio_Cell2_t, pNodeInfo []int32, vTemp *[]Edge) {
	iNodeCell := pNodeInfo[3*iFanin]
	pObj.Data = pCells[iNodeCell].Gate()
	Perm := pNodeInfo[3*iFanin+1]
	*vTemp = (*vTemp)[:0]
	for v := int32(0); v < Abc_ObjFaninNum(pObj); v++ {
		*vTemp = append(*vTemp, pObj.fanins[(Perm>>(v<<2))&15])
	}
	copy(pObj.fanins, *vTemp)
}
func Abc_ObjChangePerform(pObj *Object, vInfo []int32, vFirst []int32, fUseInv int32, vTemp *[]Edge, vFanout, vFanout2 *vec.Vec_Ptr_t, pCells []mio.Mio_Cell2_t) {
	var pNodeInv *Object
	iNodeCell := mio.Mio_GateReadCell(sfmMioGate(pObj))
	pNodeInfo := vInfo[vFirst[iNodeCell]:]
	iFanin := Abc_ObjFaninNum(pObj)
	assert(iFanin > 0 && pNodeInfo[3*iFanin] != -1, "iFanin > 0 && pNodeInfo[3*iFanin] != -1")
	// update the node
	Abc_NodeCollectFanouts(pObj, vFanout)
	if sfmAreaNodeIsInv(pObj) != 0 {
		pFanin := Abc_ObjFanin0(pObj)
		for k := 0; k < vec.Vec_PtrSize(vFanout); k++ {
			pNext := vec.Vec_PtrEntry(vFanout, k).(*Object)
			Abc_ObjPatchFanin(pNext, pObj, pFanin)
		}
		assert(Abc_ObjFanoutNum(pObj) == 0, "Abc_ObjFanoutNum(pObj) == 0")
		Abc_NtkDeleteObj(pObj)
		pObj = pFanin
		//        assert( fUseInv == 0 );
	} else {
		Abc_ObjChangeUpdate(pObj, iFanin, pCells, pNodeInfo, vTemp)
	}
	// add inverter if needed
	if fUseInv != 0 {
		pNodeInv = Abc_NtkCreateNodeInv(pObj.network, pObj)
	}
	// update the fanouts
	for i := 0; i < vec.Vec_PtrSize(vFanout); i++ {
		pNext := vec.Vec_PtrEntry(vFanout, i).(*Object)
		if Abc_ObjIsNode(pNext) == 0 || sfmAreaNodeIsBuf(pNext) != 0 {
			Abc_ObjPatchFanin(pNext, pObj, pNodeInv)
			continue
		}
		if sfmAreaNodeIsInv(pNext) != 0 {
			Abc_NodeCollectFanouts(pNext, vFanout2)
			for k := 0; k < vec.Vec_PtrSize(vFanout2); k++ {
				pNext2 := vec.Vec_PtrEntry(vFanout2, k).(*Object)
				Abc_ObjPatchFanin(pNext2, pNext, pObj)
			}
			assert(Abc_ObjFanoutNum(pNext) == 0, "Abc_ObjFanoutNum(pNext) == 0")
			Abc_NtkDeleteObj(pNext)
			continue
		}
		iFanin = Abc_NodeFindFanin(pNext, pObj)
		iFanCell := mio.Mio_GateReadCell(sfmMioGate(pNext))
		pFanInfo := vInfo[vFirst[iFanCell]:]
		if pFanInfo[3*iFanin] == -1 {
			Abc_ObjPatchFanin(pNext, pObj, pNodeInv)
			continue
		}
		Abc_ObjChangeUpdate(pNext, iFanin, pCells, pFanInfo, vTemp)
	}
}
func Abc_NtkChangePerform(pNtk *Network, fVerbose int32) {
	clk := util.Abc_Clock()
	var fNeedInv, nCells int32
	GainAll, Count, CountInv := int32(0), int32(0), int32(0)
	pCells := mio.Mio_CollectRootsNewDefault2(6, &nCells, 0)
	if pCells == nil { // a constant, buffer or inverter is missing (reported)
		return
	}
	vInfo := Abc_NtkPrecomputeCellPairs(pCells, nCells)
	vFirst := Abc_NtkPrecomputeFirsts(pCells, nCells)

	vFanout := vec.Vec_PtrAlloc(100)
	vFanout2 := vec.Vec_PtrAlloc(100)
	var vTemp []Edge
	for i := int32(0); i < Abc_NtkObjNumMax(pNtk); i++ {
		pObj := Abc_NtkObj(pNtk, i)
		if pObj == nil || Abc_ObjIsNode(pObj) == 0 {
			continue
		}
		if Abc_ObjFaninNum(pObj) < 2 && sfmAreaNodeIsInv(pObj) == 0 {
			continue
		}
		if Abc_ObjHasDupFanouts(pObj) != 0 {
			continue
		}
		Gain := Abc_ObjChangeEval(pObj, vInfo, vFirst, int32(pCells[3].AreaW), &fNeedInv)
		if Gain <= 0 {
			continue
		}
		//printf( "Obj %d\n", Abc_ObjId(pObj) );
		Count++
		CountInv += sfmAreaNodeIsInv(pObj)
		GainAll += Gain
		Abc_ObjChangePerform(pObj, vInfo, vFirst, fNeedInv, &vTemp, vFanout, vFanout2, pCells)
	}

	if fVerbose != 0 {
		fmt.Printf("Total gain in area = %6.2f after %d changes (including %d inverters). ", scl.Scl_Int2Flt(GainAll), Count, CountInv)
	}
	if fVerbose != 0 {
		util.Abc_PrintTime(1, "Time", util.Abc_Clock()-clk)
	}
}
