package cudd

import (
	"github.com/glycerine/abcgo/bdd/cudd/core"
	"github.com/glycerine/abcgo/bdd/mtr"
)

// Cudd_MakeTreeNode adds a variable group in level coordinates, matching the
// CUDD group-tree contract.
func Cudd_MakeTreeNode(dd *DdManager, low, size, flags uint32) *mtr.MtrNode {
	if dd == nil || size == 0 {
		return nil
	}
	level := int(low)
	if level < len(dd.Vars) {
		level = int(Cudd_ReadPerm(dd, int32(low)))
	}
	if level < 0 || uint64(level)+uint64(size)-1 > uint64(mtr.MTR_MAXHIGH) {
		return nil
	}
	tree := core.Cudd_ReadTree(dd)
	if tree == nil {
		treeSize := len(dd.Vars)
		if level+int(size) > treeSize {
			treeSize = level + int(size)
		}
		if treeSize == 0 {
			treeSize = 1
		}
		tree = mtr.Mtr_InitGroupTree(0, treeSize)
		if tree == nil {
			return nil
		}
		if len(dd.InvPerm) != 0 {
			tree.Index = uint32(dd.InvPerm[0])
		}
		core.Cudd_SetTree(dd, tree)
	} else if uint32(level)+size > tree.Size {
		tree.Size = uint32(level) + size
	}
	group := mtr.Mtr_MakeGroup(tree, uint32(level), size, flags)
	if group != nil {
		group.Index = low
	}
	return group
}

func cuddTreeSifting(dd *DdManager, method Cudd_ReorderingType) int32 {
	if dd == nil {
		return 0
	}
	if core.Cudd_ReduceHeap(dd, method, 0) != 0 {
		return 1
	}
	return 0
}
