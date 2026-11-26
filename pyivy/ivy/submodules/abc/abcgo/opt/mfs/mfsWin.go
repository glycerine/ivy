package mfs

import "github.com/glycerine/abcgo/base/abc"

// Port of mfsWin.c.

// Abc_MfsComputeRootsCheck returns 1 if the node should be a window root.
func Abc_MfsComputeRootsCheck(pNode *abc.Object, nLevelMax, nFanoutLimit int32) int32 {
	// the node is the root if one of the following is true:
	// (1) the node has more than fanouts than the limit
	if abc.Abc_ObjFanoutNum(pNode) > nFanoutLimit {
		return 1
	}
	// (2) the node has CO fanouts
	// (3) the node has fanouts above the cutoff level
	for i := int32(0); i < abc.Abc_ObjFanoutNum(pNode); i++ {
		pFanout := abc.Abc_ObjFanout(pNode, i)
		if abc.Abc_ObjIsCo(pFanout) != 0 || int32(pFanout.Level) > nLevelMax {
			return 1
		}
	}
	return 0
}

func Abc_MfsComputeRoots_rec(pNode *abc.Object, nLevelMax, nFanoutLimit int32, vRoots *[]*abc.Object) {
	assert(abc.Abc_ObjIsNode(pNode) != 0, "Abc_ObjIsNode(pNode)")
	if abc.Abc_NodeIsTravIdCurrent(pNode) != 0 {
		return
	}
	abc.Abc_NodeSetTravIdCurrent(pNode)
	// check if the node should be the root
	if Abc_MfsComputeRootsCheck(pNode, nLevelMax, nFanoutLimit) != 0 {
		*vRoots = append(*vRoots, pNode)
	} else { // if not, explore its fanouts
		for i := int32(0); i < abc.Abc_ObjFanoutNum(pNode); i++ {
			Abc_MfsComputeRoots_rec(abc.Abc_ObjFanout(pNode, i), nLevelMax, nFanoutLimit, vRoots)
		}
	}
}

// Abc_MfsComputeRoots computes the roots of the window.
func Abc_MfsComputeRoots(pNode *abc.Object, nWinTfoMax, nFanoutLimit int32) []*abc.Object {
	vRoots := make([]*abc.Object, 0, 10)
	abc.Abc_NtkIncrementTravId(abc.Abc_ObjNtk(pNode))
	Abc_MfsComputeRoots_rec(pNode, int32(pNode.Level)+nWinTfoMax, nFanoutLimit, &vRoots)
	assert(len(vRoots) > 0, "Vec_PtrSize(vRoots) > 0")
	return vRoots
}
