// DFS and cone manipulation, ported from ABC hopDfs.c (Alan Mishchenko).
package hop

import "fmt"

// Postorder visits each internal node once. Mark A is the original shared
// traversal mark; a separate active stack diagnoses cycles before stack overflow.
func hopDfsVisit(root *Hop_Obj_t, active map[*Hop_Obj_t]bool, visit func(*Hop_Obj_t)) {
	if root == nil || Hop_IsComplement(root) != 0 {
		panic("hop: DFS requires a regular node")
	}
	if Hop_ObjIsNode(root) == 0 || root.fMarkA != 0 {
		return
	}
	if active[root] {
		panic("hop: cycle in AIG")
	}
	active[root] = true
	hopDfsVisit(Hop_ObjFanin0(root), active, visit)
	hopDfsVisit(Hop_ObjFanin1(root), active, visit)
	delete(active, root)
	visit(root)
	root.fMarkA = 1
}
func Hop_ManDfs_rec(o *Hop_Obj_t, nodes *Vec_Ptr_t) {
	hopDfsVisit(o, map[*Hop_Obj_t]bool{}, func(o *Hop_Obj_t) { Vec_PtrPush(nodes, o) })
}
func Hop_ManDfs(p *Hop_Man_t) *Vec_Ptr_t {
	nodes := Vec_PtrAlloc(Hop_ManNodeNum(p))
	for _, head := range p.pTable {
		for o := head; o != nil; o = o.pNext {
			Hop_ManDfs_rec(o, nodes)
		}
	}
	for _, head := range p.pTable {
		for o := head; o != nil; o = o.pNext {
			o.fMarkA = 0
		}
	}
	return nodes
}
func Hop_ManDfsNode(p *Hop_Man_t, o *Hop_Obj_t) *Vec_Ptr_t {
	nodes := Vec_PtrAlloc(16)
	Hop_ManDfs_rec(o, nodes)
	for _, entry := range nodes.pArray[:nodes.nSize] {
		entry.(*Hop_Obj_t).fMarkA = 0
	}
	return nodes
}
func Hop_ManCountLevels(p *Hop_Man_t) int32 {
	p.pConst1.pData = nil
	for _, entry := range p.vPis.pArray[:p.vPis.nSize] {
		entry.(*Hop_Obj_t).pData = nil
	}
	nodes := Hop_ManDfs(p)
	level := func(o *Hop_Obj_t) int32 {
		if o.pData == nil {
			return 0
		}
		return o.pData.(int32)
	}
	for _, entry := range nodes.pArray[:nodes.nSize] {
		o := entry.(*Hop_Obj_t)
		a, b := level(Hop_ObjFanin0(o)), level(Hop_ObjFanin1(o))
		if b > a {
			a = b
		}
		o.pData = 1 + Hop_ObjIsExor(o) + a
	}
	var max int32
	for _, entry := range p.vPos.pArray[:p.vPos.nSize] {
		if v := level(Hop_ObjFanin0(entry.(*Hop_Obj_t))); v > max {
			max = v
		}
	}
	return max
}
func Hop_ManCreateRefs(p *Hop_Man_t) {
	if p.fRefCount != 0 {
		return
	}
	p.fRefCount = 1
	p.pConst1.nRefs = 0
	for _, entry := range p.vPis.pArray[:p.vPis.nSize] {
		entry.(*Hop_Obj_t).nRefs = 0
	}
	for _, entry := range p.vPos.pArray[:p.vPos.nSize] {
		entry.(*Hop_Obj_t).nRefs = 0
	}
	for _, head := range p.pTable {
		for o := head; o != nil; o = o.pNext {
			o.nRefs = 0
		}
	}
	for _, head := range p.pTable {
		for o := head; o != nil; o = o.pNext {
			Hop_ObjFanin0(o).nRefs++
			Hop_ObjFanin1(o).nRefs++
		}
	}
	for _, entry := range p.vPos.pArray[:p.vPos.nSize] {
		Hop_ObjFanin0(entry.(*Hop_Obj_t)).nRefs++
	}
}
func Hop_ConeMark_rec(o *Hop_Obj_t) { hopDfsVisit(o, map[*Hop_Obj_t]bool{}, func(*Hop_Obj_t) {}) }
func Hop_ConeCleanAndMark_rec(o *Hop_Obj_t) {
	hopDfsVisit(o, map[*Hop_Obj_t]bool{}, func(o *Hop_Obj_t) { o.pData = nil })
}
func Hop_ConeCountAndMark_rec(o *Hop_Obj_t) int32 {
	var count int32
	hopDfsVisit(o, map[*Hop_Obj_t]bool{}, func(*Hop_Obj_t) { count++ })
	return count
}
func Hop_ConeUnmark_rec(o *Hop_Obj_t) {
	if o == nil || Hop_IsComplement(o) != 0 {
		panic("hop: unmark requires a regular node")
	}
	if Hop_ObjIsNode(o) == 0 || o.fMarkA == 0 {
		return
	}
	// Clear before descending: shared subgraphs need only one visit.
	o.fMarkA = 0
	Hop_ConeUnmark_rec(Hop_ObjFanin0(o))
	Hop_ConeUnmark_rec(Hop_ObjFanin1(o))
}
func Hop_DagSize(o *Hop_Obj_t) int32 {
	regular := Hop_Regular(o)
	n := Hop_ConeCountAndMark_rec(regular)
	Hop_ConeUnmark_rec(regular)
	return n
}
func Hop_ObjFanoutCount_rec(o, pivot *Hop_Obj_t) int32 {
	if o == nil || Hop_IsComplement(o) != 0 {
		panic("hop: fanout traversal requires a regular node")
	}
	if Hop_ObjIsNode(o) == 0 || o.fMarkA != 0 {
		return hopBool(o == pivot)
	}
	n := Hop_ObjFanoutCount_rec(Hop_ObjFanin0(o), pivot) + Hop_ObjFanoutCount_rec(Hop_ObjFanin1(o), pivot)
	o.fMarkA = 1
	return n
}
func Hop_ObjFanoutCount(o, pivot *Hop_Obj_t) int32 {
	if Hop_IsComplement(pivot) != 0 {
		panic("hop: pivot must be regular")
	}
	n := Hop_ObjFanoutCount_rec(Hop_Regular(o), pivot)
	Hop_ConeUnmark_rec(Hop_Regular(o))
	return n
}
func Hop_Transfer_rec(dest *Hop_Man_t, o *Hop_Obj_t) {
	hopDfsVisit(o, map[*Hop_Obj_t]bool{}, func(n *Hop_Obj_t) { n.pData = Hop_And(dest, Hop_ObjChild0Copy(n), Hop_ObjChild1Copy(n)) })
}
func Hop_Transfer(source, dest *Hop_Man_t, root *Hop_Obj_t, nVars int32) *Hop_Obj_t {
	if source == dest {
		return root
	}
	if Hop_ObjIsConst1(Hop_Regular(root)) != 0 {
		return Hop_NotCond(dest.pConst1, Hop_IsComplement(root))
	}
	for i := int32(0); i < Hop_ManPiNum(source) && i < nVars; i++ {
		Hop_ManPi(source, i).pData = Hop_IthVar(dest, i)
	}
	// Constants may appear as fanins of retained nodes as well as at the root.
	source.pConst1.pData = dest.pConst1
	Hop_Transfer_rec(dest, Hop_Regular(root))
	Hop_ConeUnmark_rec(Hop_Regular(root))
	copied, ok := Hop_Regular(root).pData.(*Hop_Obj_t)
	if !ok {
		panic("hop: missing variable mapping during transfer")
	}
	return Hop_NotCond(copied, Hop_IsComplement(root))
}
func hopComposeVisit(p *Hop_Man_t, o, function, variable *Hop_Obj_t, active map[*Hop_Obj_t]bool) {
	if o == nil || Hop_IsComplement(o) != 0 {
		panic("hop: compose requires a regular node")
	}
	if o.fMarkA != 0 {
		return
	}
	if Hop_ObjIsConst1(o) != 0 || Hop_ObjIsPi(o) != 0 {
		if o == variable {
			o.pData = function
		} else {
			o.pData = o
		}
		return
	}
	if active[o] {
		panic("hop: cycle in composition")
	}
	active[o] = true
	hopComposeVisit(p, Hop_ObjFanin0(o), function, variable, active)
	hopComposeVisit(p, Hop_ObjFanin1(o), function, variable, active)
	delete(active, o)
	o.pData = Hop_And(p, Hop_ObjChild0Copy(o), Hop_ObjChild1Copy(o))
	o.fMarkA = 1
}
func Hop_Compose_rec(p *Hop_Man_t, o, function, variable *Hop_Obj_t) {
	hopComposeVisit(p, o, function, variable, map[*Hop_Obj_t]bool{})
}
func Hop_Compose(p *Hop_Man_t, root, function *Hop_Obj_t, iVar int32) *Hop_Obj_t {
	if iVar < 0 || iVar >= Hop_ManPiNum(p) {
		fmt.Printf("Hop_Compose(): The PI variable %d is not defined.\n", iVar)
		return nil
	}
	Hop_Compose_rec(p, Hop_Regular(root), function, Hop_ManPi(p, iVar))
	Hop_ConeUnmark_rec(Hop_Regular(root))
	return Hop_NotCond(Hop_Regular(root).pData.(*Hop_Obj_t), Hop_IsComplement(root))
}
func Hop_Complement_rec(p *Hop_Man_t, o, variable *Hop_Obj_t) {
	Hop_Compose_rec(p, o, Hop_Not(variable), variable)
}
func Hop_Complement(p *Hop_Man_t, root *Hop_Obj_t, iVar int32) *Hop_Obj_t {
	if iVar < 0 || iVar >= Hop_ManPiNum(p) {
		fmt.Printf("Hop_Complement(): The PI variable %d is not defined.\n", iVar)
		return nil
	}
	Hop_Complement_rec(p, Hop_Regular(root), Hop_ManPi(p, iVar))
	Hop_ConeUnmark_rec(Hop_Regular(root))
	return Hop_NotCond(Hop_Regular(root).pData.(*Hop_Obj_t), Hop_IsComplement(root))
}
func Hop_Remap_rec(p *Hop_Man_t, o *Hop_Obj_t) { Hop_Transfer_rec(p, o) }
func Hop_Remap(p *Hop_Man_t, root *Hop_Obj_t, support uint32, nVars int32) *Hop_Obj_t {
	if nVars < 0 || nVars > Hop_ManPiNum(p) || nVars > 32 {
		fmt.Printf("Hop_Remap(): Invalid number of variables (%d).\n", nVars)
		return nil
	}
	if Hop_ObjIsConst1(Hop_Regular(root)) != 0 {
		return root
	}
	if support == 0 {
		return Hop_NotCond(Hop_ManConst0(p), Hop_ObjPhaseCompl(root))
	}
	var k int32
	for i := int32(0); i < nVars; i++ {
		o := Hop_ManPi(p, i)
		if support&(uint32(1)<<uint(i)) != 0 {
			o.pData = Hop_IthVar(p, k)
			k++
		} else {
			o.pData = Hop_ManConst0(p)
		}
	}
	if k <= 0 || k >= nVars {
		panic("hop: remap requires a nonempty proper support")
	}
	p.pConst1.pData = p.pConst1
	Hop_Remap_rec(p, Hop_Regular(root))
	Hop_ConeUnmark_rec(Hop_Regular(root))
	return Hop_NotCond(Hop_Regular(root).pData.(*Hop_Obj_t), Hop_IsComplement(root))
}
func Hop_Permute(p *Hop_Man_t, root *Hop_Obj_t, nVars int32, permutation []int32) *Hop_Obj_t {
	if Hop_ObjIsConst1(Hop_Regular(root)) != 0 {
		return root
	}
	if nVars < 0 || nVars > Hop_ManPiNum(p) || int(nVars) > len(permutation) {
		panic("hop: invalid permutation size")
	}
	for i := int32(0); i < nVars; i++ {
		if permutation[i] < 0 || permutation[i] >= Hop_ManPiNum(p) {
			panic("hop: invalid permutation index")
		}
		Hop_ManPi(p, i).pData = Hop_IthVar(p, permutation[i])
	}
	p.pConst1.pData = p.pConst1
	Hop_Remap_rec(p, Hop_Regular(root))
	Hop_ConeUnmark_rec(Hop_Regular(root))
	return Hop_NotCond(Hop_Regular(root).pData.(*Hop_Obj_t), Hop_IsComplement(root))
}
