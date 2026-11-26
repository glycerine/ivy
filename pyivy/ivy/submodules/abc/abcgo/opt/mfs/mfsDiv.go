package mfs

import (
	"slices"

	"github.com/glycerine/abcgo/base/abc"
)

// Port of mfsDiv.c.

// Abc_MfsWinMarkTfi_rec marks and collects the TFI cone of the node.
func Abc_MfsWinMarkTfi_rec(pObj *abc.Object, vCone *[]*abc.Object) {
	if abc.Abc_NodeIsTravIdCurrent(pObj) != 0 {
		return
	}
	abc.Abc_NodeSetTravIdCurrent(pObj)
	if abc.Abc_ObjIsCi(pObj) != 0 {
		*vCone = append(*vCone, pObj)
		return
	}
	assert(abc.Abc_ObjIsNode(pObj) != 0, "Abc_ObjIsNode(pObj)")
	// visit the fanins of the node
	for i := int32(0); i < abc.Abc_ObjFaninNum(pObj); i++ {
		Abc_MfsWinMarkTfi_rec(abc.Abc_ObjFanin(pObj, i), vCone)
	}
	*vCone = append(*vCone, pObj)
}

// Abc_MfsWinMarkTfi marks and collects the TFI cone of the node.
func Abc_MfsWinMarkTfi(pNode *abc.Object) []*abc.Object {
	vCone := make([]*abc.Object, 0, 100)
	Abc_MfsWinMarkTfi_rec(pNode, &vCone)
	return vCone
}

// Abc_MfsWinSweepLeafTfo_rec marks the TFO of the collected nodes.
func Abc_MfsWinSweepLeafTfo_rec(pObj *abc.Object, nLevelLimit int32) {
	if abc.Abc_ObjIsCo(pObj) != 0 || int32(pObj.Level) > nLevelLimit {
		return
	}
	if abc.Abc_NodeIsTravIdCurrent(pObj) != 0 {
		return
	}
	abc.Abc_NodeSetTravIdCurrent(pObj)
	for i := int32(0); i < abc.Abc_ObjFanoutNum(pObj); i++ {
		Abc_MfsWinSweepLeafTfo_rec(abc.Abc_ObjFanout(pObj, i), nLevelLimit)
	}
}

// Abc_MfsComputeDivisors computes divisors and window for the node.
// (Abc_MfsNodeDeref_rec / Abc_MfsNodeRef_rec / Abc_MfsWinVisitMffc are only
// called from commented-out code in C and are not ported.)
func Abc_MfsComputeDivisors(p *Mfs_Man_t, pNode *abc.Object, nLevDivMax int32) []*abc.Object {
	assert(p.vDivs == nil, "p->vDivs == NULL")

	// mark the TFI with the current trav ID
	abc.Abc_NtkIncrementTravId(abc.Abc_ObjNtk(pNode))
	vCone := Abc_MfsWinMarkTfi(pNode)

	// mark with the current trav ID those nodes that should not be divisors:
	// (1) the node and its TFO
	// (2) the MFFC of the node
	// (3) the node's fanins (these are treated as a special case)
	abc.Abc_NtkIncrementTravId(abc.Abc_ObjNtk(pNode))
	Abc_MfsWinSweepLeafTfo_rec(pNode, nLevDivMax)
	for k := int32(0); k < abc.Abc_ObjFaninNum(pNode); k++ {
		abc.Abc_NodeSetTravIdCurrent(abc.Abc_ObjFanin(pNode, k))
	}
	// at this point the nodes are marked with two trav IDs:
	// nodes to be collected as divisors are marked with previous trav ID
	// nodes to be avoided as divisors are marked with current trav ID

	// start collecting the divisors
	nWinMax := int(p.pPars.NWinMax)
	vDivs := make([]*abc.Object, 0, nWinMax+MFS_FANIN_MAX)
	for _, pObj := range vCone {
		if abc.Abc_NodeIsTravIdPrevious(pObj) == 0 {
			continue
		}
		if int32(pObj.Level) > nLevDivMax {
			continue
		}
		vDivs = append(vDivs, pObj)
		if len(vDivs) >= nWinMax {
			break
		}
	}

	// explore the fanouts of already collected divisors
	if len(vDivs) < nWinMax {
		for k := 0; k < len(vDivs); k++ {
			pObj := vDivs[k]
			// consider fanouts of this node
			for f := int32(0); f < abc.Abc_ObjFanoutNum(pObj); f++ {
				pFanout := abc.Abc_ObjFanout(pObj, f)
				// stop if there are too many fanouts
				if p.pPars.NFanoutsMax != 0 && f > p.pPars.NFanoutsMax {
					break
				}
				// skip nodes that are already added
				if abc.Abc_NodeIsTravIdPrevious(pFanout) != 0 {
					continue
				}
				// skip nodes in the TFO or in the MFFC of node
				if abc.Abc_NodeIsTravIdCurrent(pFanout) != 0 {
					continue
				}
				// skip COs
				if abc.Abc_ObjIsNode(pFanout) == 0 {
					continue
				}
				// skip nodes with large level
				if int32(pFanout.Level) > nLevDivMax {
					continue
				}
				// skip nodes whose fanins are not divisors  -- here we skip more than we need to skip!!! (revise later)  August 7, 2009
				m := int32(0)
				for ; m < abc.Abc_ObjFaninNum(pFanout); m++ {
					if abc.Abc_NodeIsTravIdPrevious(abc.Abc_ObjFanin(pFanout, m)) == 0 {
						break
					}
				}
				if m < abc.Abc_ObjFaninNum(pFanout) {
					continue
				}
				// add the node to the divisors
				vDivs = append(vDivs, pFanout)
				if !slices.Contains(p.vNodes, pFanout) {
					p.vNodes = append(p.vNodes, pFanout)
				}
				abc.Abc_NodeSetTravIdPrevious(pFanout)
				if len(vDivs) >= nWinMax {
					break
				}
			}
			if len(vDivs) >= nWinMax {
				break
			}
		}
	}
	if len(vDivs) >= nWinMax {
		p.nMaxDivs++
	}

	// sort the divisors by level in the increasing order
	slices.SortFunc(vDivs, func(a, b *abc.Object) int {
		return int(abc.Abc_NodeCompareLevelsIncrease(&a, &b))
	})

	// add the fanins of the node
	for k := int32(0); k < abc.Abc_ObjFaninNum(pNode); k++ {
		vDivs = append(vDivs, abc.Abc_ObjFanin(pNode, k))
	}
	return vDivs
}
