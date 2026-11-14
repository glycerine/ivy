package core

// Cudd_bddLinearTransform returns f with variable x replaced by XNOR(x,y).
// The result is rebuilt through the manager's canonical ITE constructor, so
// arbitrary variable orders and complemented edges are handled exactly.
// x and y are variable indices (not levels). The input root is not modified;
// callers that own roots must transfer references in the usual CUDD manner.
func Cudd_bddLinearTransform(dd *DdManager, f *DdNode, x, y int) *DdNode {
	if !dd.valid(f) || x < 0 || y < 0 || x == y || x >= len(dd.Vars) || y >= len(dd.Vars) {
		return nil
	}
	xnor := Cudd_bddXnor(dd, dd.Vars[x], dd.Vars[y])
	if xnor == nil {
		return nil
	}
	return Cudd_bddCompose(dd, f, xnor, x)
}

// Cudd_bddLinearTransformInPlace applies the adjacent linear transform to
// every canonical BDD node. x and y are adjacent levels, with x < y. Existing
// regular node objects remain the roots for their transformed functions; new
// nodes are allocated for subfunctions introduced by the new decomposition.
func Cudd_bddLinearTransformInPlace(dd *DdManager, x, y int) bool {
	if dd == nil || dd.closed || x < 0 || y != x+1 || y >= len(dd.InvPerm) || len(dd.Perm) != len(dd.Vars) {
		return false
	}
	xi, yi := dd.InvPerm[x], dd.InvPerm[y]
	if xi < 0 || yi < 0 || xi >= len(dd.Vars) || yi >= len(dd.Vars) || dd.Perm[xi] != x || dd.Perm[yi] != y {
		return false
	}
	oldNodes := make([]*DdNode, 0, len(dd.unique))
	for _, n := range dd.unique {
		oldNodes = append(oldNodes, n)
	}
	// Build the transformed graph in a scratch manager first. No live state is
	// changed until every old canonical function has a transformed image.
	scratch := NewBDDManager(len(dd.Vars))
	copy(scratch.InvPerm, dd.InvPerm)
	copy(scratch.Perm, dd.Perm)
	transferMemo := map[*DdNode]*DdNode{}
	var transfer func(*DdNode) *DdNode
	transfer = func(edge *DdNode) *DdNode {
		if Cudd_IsConstant(edge) {
			return Cudd_NotCond(scratch.One, Cudd_IsComplement(edge))
		}
		regular := Cudd_Regular(edge)
		fresh := transferMemo[regular]
		if fresh == nil {
			fresh = scratch.node(regular.Index, transfer(regular.T), transfer(regular.E))
			transferMemo[regular] = fresh
		}
		return Cudd_NotCond(fresh, Cudd_IsComplement(edge))
	}
	xnor := Cudd_bddXnor(scratch, scratch.Vars[xi], scratch.Vars[yi])
	byImage := make(map[*DdNode]*DdNode, len(oldNodes))
	for _, old := range oldNodes {
		image := Cudd_bddCompose(scratch, transfer(old), xnor, xi)
		if image == nil || Cudd_IsComplement(image) || Cudd_IsConstant(image) {
			return false
		}
		if prior := byImage[image]; prior != nil && prior != old {
			return false
		}
		byImage[image] = old
	}
	// Allocate handles for scratch nodes that are needed as children but are
	// not themselves the image of an old canonical node.
	remap := make(map[*DdNode]*DdNode, len(scratch.unique))
	for fresh, old := range byImage {
		remap[fresh] = old
	}
	for key, fresh := range scratch.unique {
		if remap[fresh] == nil {
			n := &DdNode{Index: key.index, ID: dd.nextID, manager: dd}
			n.inverse = &DdNode{Index: key.index, ID: dd.nextID + 1, manager: dd, regular: n, inverse: n}
			dd.nextID += 2
			remap[fresh] = n
		}
	}
	remapEdge := func(edge *DdNode) *DdNode {
		if Cudd_IsConstant(edge) {
			return Cudd_NotCond(dd.One, Cudd_IsComplement(edge))
		}
		return Cudd_NotCond(remap[Cudd_Regular(edge)], Cudd_IsComplement(edge))
	}
	for fresh, node := range remap {
		node.Index = fresh.Index
		node.T = remapEdge(fresh.T)
		node.E = remapEdge(fresh.E)
		node.Value = fresh.Value
		if node.inverse != nil {
			node.inverse.Index = node.Index
		}
	}
	nextUnique := make(map[bddUniqueKey]*DdNode, len(remap))
	for _, node := range remap {
		if Cudd_IsConstant(node) {
			continue
		}
		key := bddUniqueKey{node.Index, node.T, node.E}
		if prior := nextUnique[key]; prior != nil && prior != node {
			return false
		}
		nextUnique[key] = node
	}
	dd.unique = nextUnique
	dd.cache = map[bddITEKey]*DdNode{}
	dd.computed = map[computedKey]*DdNode{}
	return true
}

// Cudd_bddLinearNodeCount returns the BDD unique-table key count excluding
// isolated projection nodes, matching table->keys - table->isolated.
func Cudd_bddLinearNodeCount(dd *DdManager) int {
	if dd == nil || dd.closed {
		return 0
	}
	return len(bddLinearCountedNodes(dd))
}

// Cudd_bddLinearPreprocessSize includes isolated projection nodes, matching
// Cudd_ReduceHeap's pre-reordering minsize check.
func Cudd_bddLinearPreprocessSize(dd *DdManager) int {
	if dd == nil || dd.closed {
		return 0
	}
	live := bddLinearLiveNodes(dd)
	for _, variable := range dd.Vars {
		if !live[Cudd_Regular(variable)] {
			live[Cudd_Regular(variable)] = true
		}
	}
	return len(live)
}

// Cudd_bddLinearLevelNodeCount returns the BDD unique-table population at a
// variable's current level, excluding its isolated projection node.
func Cudd_bddLinearLevelNodeCount(dd *DdManager, level int) int {
	if dd == nil || dd.closed || level < 0 || level >= len(dd.InvPerm) {
		return 0
	}
	index := uint32(dd.InvPerm[level])
	count := 0
	for node := range bddLinearCountedNodes(dd) {
		if node.Index == index {
			count++
		}
	}
	return count
}

func bddLinearLiveNodes(dd *DdManager) map[*DdNode]bool {
	live := make(map[*DdNode]bool)
	var visit func(*DdNode)
	visit = func(edge *DdNode) {
		if edge == nil || Cudd_IsConstant(edge) {
			return
		}
		node := Cudd_Regular(edge)
		if live[node] {
			return
		}
		live[node] = true
		visit(node.T)
		visit(node.E)
	}
	for _, node := range dd.unique {
		if node.Ref != 0 {
			visit(node)
		}
	}
	// Projection handles are manager roots even though Go Ref does not encode
	// their implicit CUDD manager pin. They are added here and filtered below
	// when their Go Ref count is zero.
	for _, variable := range dd.Vars {
		visit(variable)
	}
	return live
}

func bddLinearCountedNodes(dd *DdManager) map[*DdNode]bool {
	live := bddLinearLiveNodes(dd)
	projections := make(map[*DdNode]bool, len(dd.Vars))
	for _, variable := range dd.Vars {
		if variable != nil && !Cudd_IsConstant(variable) {
			projections[Cudd_Regular(variable)] = true
		}
	}
	for node := range live {
		// In this Go core, Ref counts all incoming child arcs and explicit root
		// references, but does not store the manager's implicit projection pin.
		// Thus Ref==0 corresponds to CUDD's isolated projection ref==1.
		if projections[node] && node.Ref == 0 {
			delete(live, node)
		}
	}
	return live
}
