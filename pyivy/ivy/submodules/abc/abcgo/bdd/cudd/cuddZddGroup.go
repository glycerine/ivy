package cudd

import (
	"github.com/glycerine/abcgo/bdd/cudd/core"
	"github.com/glycerine/abcgo/bdd/mtr"
)

func Cudd_MakeZddTreeNode(dd *DdManager, low, size, flags uint32) *mtr.MtrNode {
	if dd == nil || size == 0 {
		return nil
	}
	level := int(low)
	if level < len(dd.PermZ) {
		level = dd.PermZ[level]
	}
	if level < 0 || uint64(level)+uint64(size)-1 > uint64(mtr.MTR_MAXHIGH) {
		return nil
	}
	tree := core.Cudd_ReadZddTree(dd)
	if tree == nil {
		treeSize := dd.ZddSize
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
		if len(dd.InvPermZ) != 0 {
			tree.Index = uint32(dd.InvPermZ[0])
		}
		core.Cudd_SetZddTree(dd, tree)
	} else if uint32(level)+size > tree.Size {
		tree.Size = uint32(level) + size
	}
	group := mtr.Mtr_MakeGroup(tree, uint32(level), size, flags)
	if group != nil {
		group.Index = low
	}
	return group
}
