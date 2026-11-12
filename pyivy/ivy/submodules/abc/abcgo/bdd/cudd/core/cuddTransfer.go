package core

// Transfer rebuilds an owned BDD or ADD in destination. A nil permutation keeps
// variable indices; otherwise permutation[source index] is the destination index.
// The result is canonical in destination and is not externally referenced.
func (dd *DdManager) Transfer(root *DdNode, destination *DdManager, permutation []int) *DdNode {
	if dd == nil || destination == nil || destination.closed {
		return nil
	}
	add := Cudd_IsAdd(root)
	if add {
		if !dd.validAdd(root) {
			return nil
		}
	} else if !dd.valid(root) {
		return nil
	}
	support := dd.Support(root)
	for _, index := range support {
		target := index
		if permutation != nil {
			if index >= len(permutation) {
				return nil
			}
			target = permutation[index]
		}
		if target < 0 || uint64(target) >= uint64(bddConstantIndex) {
			return nil
		}
	}
	memo := map[*DdNode]*DdNode{}
	var transfer func(*DdNode) *DdNode
	transfer = func(edge *DdNode) *DdNode {
		node := Cudd_Regular(edge)
		result := memo[node]
		if result == nil {
			if Cudd_IsConstant(node) {
				result = destination.constant(node.Value)
			} else {
				target := int(node.Index)
				if permutation != nil {
					target = permutation[target]
				}
				t, e := transfer(node.T), transfer(node.E)
				if add {
					result = Cudd_addIte(destination, Cudd_addIthVar(destination, target), t, e)
				} else {
					result = destination.ite(Cudd_bddIthVar(destination, target), t, e)
				}
			}
			memo[node] = result
		}
		return Cudd_NotCond(result, Cudd_IsComplement(edge))
	}
	return transfer(root)
}
func Cudd_bddTransfer(source, destination *DdManager, root *DdNode) *DdNode {
	if source == nil || !source.valid(root) {
		return nil
	}
	return source.Transfer(root, destination, nil)
}
