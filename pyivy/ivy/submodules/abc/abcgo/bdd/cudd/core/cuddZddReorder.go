package core

type zddReorderKey struct {
	node  *DdNode
	level int
}
type zddCofactorKey struct {
	node     *DdNode
	variable int
	include  bool
}

// reorderZdd rebuilds the zero-suppressed graph under a new level order while
// preserving all existing node handles. The source table is left untouched
// until the complete replacement graph has been constructed.
func reorderZdd(dd *DdManager, size int, permutation, inverse []int) bool {
	if dd == nil || size < dd.ZddSize || len(permutation) != size || len(inverse) != size {
		return false
	}
	scratch := NewBDDManager(0)
	scratch.ZddSize = size
	scratch.PermZ = append([]int(nil), permutation...)
	scratch.InvPermZ = append([]int(nil), inverse...)
	if !validZddOrder(scratch) {
		return false
	}
	cofactorMemo := make(map[zddCofactorKey]*DdNode)
	var cofactor func(*DdNode, int, bool) *DdNode
	cofactor = func(node *DdNode, variable int, include bool) *DdNode {
		if node == dd.Zero {
			return dd.Zero
		}
		if node == dd.One {
			if include {
				return dd.Zero
			}
			return dd.One
		}
		if variable >= len(dd.PermZ) {
			if include {
				return dd.Zero
			}
			return node
		}
		key := zddCofactorKey{node: node, variable: variable, include: include}
		if result, ok := cofactorMemo[key]; ok {
			return result
		}
		oldLevel, variableLevel := dd.PermZ[node.Index], dd.PermZ[variable]
		var result *DdNode
		switch {
		case oldLevel > variableLevel:
			if include {
				result = dd.Zero
			} else {
				result = node
			}
		case node.Index == uint32(variable):
			if include {
				result = node.T
			} else {
				result = node.E
			}
		default:
			t := cofactor(node.T, variable, include)
			e := cofactor(node.E, variable, include)
			result = dd.zddNode(node.Index, t, e)
		}
		cofactorMemo[key] = result
		return result
	}
	reorderMemo := make(map[zddReorderKey]*DdNode)
	var visit func(*DdNode, int) *DdNode
	visit = func(node *DdNode, level int) *DdNode {
		if node == dd.Zero {
			return scratch.Zero
		}
		if node == dd.One {
			return scratch.One
		}
		key := zddReorderKey{node: node, level: level}
		if result := reorderMemo[key]; result != nil {
			return result
		}
		if level >= size {
			return nil
		}
		variable := inverse[level]
		include := cofactor(node, variable, true)
		exclude := cofactor(node, variable, false)
		t := visit(include, level+1)
		e := visit(exclude, level+1)
		if t == nil || e == nil {
			return nil
		}
		result := scratch.zddNode(uint32(variable), t, e)
		reorderMemo[key] = result
		return result
	}
	oldNodes := make([]*DdNode, 0, len(dd.zddUnique)+len(dd.Univ))
	seen := map[*DdNode]bool{}
	for _, node := range dd.zddUnique {
		if !seen[node] {
			seen[node] = true
			oldNodes = append(oldNodes, node)
		}
	}
	for _, node := range dd.Univ {
		if node != nil && !seen[node] {
			seen[node] = true
			oldNodes = append(oldNodes, node)
		}
	}
	remap := make(map[*DdNode]*DdNode, len(oldNodes))
	preserve := make(map[*DdNode]*DdNode, len(oldNodes))
	for _, node := range oldNodes {
		fresh := visit(node, 0)
		if fresh == nil {
			return false
		}
		remap[node] = fresh
		if !Cudd_IsConstant(fresh) {
			if previous := preserve[fresh]; previous == nil || node.Ref > previous.Ref {
				preserve[fresh] = node
			}
		}
	}
	newNodes := make([]*DdNode, 0, len(scratch.zddUnique))
	for _, node := range scratch.zddUnique {
		newNodes = append(newNodes, node)
	}
	mapped := map[*DdNode]*DdNode{scratch.One: dd.One, scratch.Zero: dd.Zero}
	for _, fresh := range newNodes {
		if old := preserve[fresh]; old != nil {
			mapped[fresh] = old
		} else {
			mapped[fresh] = &DdNode{ID: dd.nextID, manager: dd, family: 1}
			dd.nextID++
		}
	}
	for _, fresh := range newNodes {
		node := mapped[fresh]
		node.Index = fresh.Index
		node.T = mapped[fresh.T]
		node.E = mapped[fresh.E]
		node.manager = dd
		node.family = 1
	}
	nextTable := make(map[bddUniqueKey]*DdNode, len(newNodes))
	for _, fresh := range newNodes {
		node := mapped[fresh]
		key := bddUniqueKey{index: node.Index, t: node.T, e: node.E}
		if other := nextTable[key]; other != nil && other != node {
			return false
		}
		nextTable[key] = node
	}
	newUniverse := make([]*DdNode, len(dd.Univ))
	for i, node := range dd.Univ {
		if node != nil {
			fresh := remap[node]
			if fresh == nil {
				return false
			}
			newUniverse[i] = mapped[Cudd_Regular(fresh)]
		}
	}
	dd.zddUnique = nextTable
	dd.ZddSize = size
	dd.PermZ = append(dd.PermZ[:0], permutation...)
	dd.InvPermZ = append(dd.InvPermZ[:0], inverse...)
	dd.Univ = newUniverse
	return cuddZddInitUniv(dd) == 1
}

func validZddOrder(dd *DdManager) bool {
	if dd == nil || dd.ZddSize < 0 || len(dd.PermZ) != dd.ZddSize || len(dd.InvPermZ) != dd.ZddSize {
		return false
	}
	seen := make([]bool, dd.ZddSize)
	for level, index := range dd.InvPermZ {
		if index < 0 || index >= dd.ZddSize || seen[index] || dd.PermZ[index] != level {
			return false
		}
		seen[index] = true
	}
	return true
}
