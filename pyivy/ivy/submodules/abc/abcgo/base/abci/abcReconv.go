package abci

import (
	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/bdd/cudd/core"
	"github.com/glycerine/abcgo/bdd/extrab"
	"github.com/glycerine/abcgo/misc/vec"
)

// Abc_ManCut_t owns the scratch vectors used by the reconvergence cut routines.
type Abc_ManCut_t struct {
	nNodeSizeMax, nConeSizeMax, nNodeFanStop, nConeFanStop int32
	vNodeLeaves, vConeLeaves, vVisited, vNodesTfo          *Vec_Ptr_t
	vLevels                                                *Vec_Vec_t
}

func Abc_NodesMark(v *Vec_Ptr_t) {
	for _, x := range vec.Vec_PtrArray(v) {
		x.(*Abc_Obj_t).MarkA = true
	}
}
func Abc_NodesUnmark(v *Vec_Ptr_t) {
	for _, x := range vec.Vec_PtrArray(v) {
		x.(*Abc_Obj_t).MarkA = false
	}
}
func Abc_NodesUnmarkB(v *Vec_Ptr_t) {
	for _, x := range vec.Vec_PtrArray(v) {
		x.(*Abc_Obj_t).MarkB = false
	}
}

func Abc_NodeGetLeafCostOne(n *Abc_Obj_t, faninLimit int32) int32 {
	assert(n != nil && n.MarkB, "leaf-cost candidate must be marked")
	if abc.Abc_ObjIsCi(n) != 0 {
		return 999
	}
	f0, f1 := abc.Abc_ObjFanin0(n), abc.Abc_ObjFanin1(n)
	cost := int32(0)
	if !f0.MarkB {
		cost++
	}
	if !f1.MarkB {
		cost++
	}
	if cost < 2 {
		return cost
	}
	if abc.Abc_ObjFanoutNum(n) > faninLimit {
		return 999
	}
	return cost
}

// Abc_NodeGetLeafCostTwo tests the C two-level expansion cases and returns
// the candidate leaf and the internal nodes that become marked.
func Abc_NodeGetLeafCostTwo(n *Abc_Obj_t, _ int32, leaf **Abc_Obj_t, mark1, mark2 **Abc_Obj_t) int32 {
	if n == nil {
		return 999
	}
	assert(n.MarkB, "two-level candidate must be marked")
	if abc.Abc_ObjIsCi(n) != 0 {
		return 999
	}
	a, b := abc.Abc_ObjFanin0(n), abc.Abc_ObjFanin1(n)
	// The candidate itself belongs to the cut, while both immediate fanins
	// are outside it. This is the precondition used by the C implementation.
	assert(!a.MarkB && !b.MarkB, "candidate fanins must be outside the cut")
	if abc.Abc_ObjIsCi(a) != 0 && abc.Abc_ObjIsCi(b) != 0 {
		return 999
	}
	if abc.Abc_ObjIsCi(a) == 0 && abc.Abc_ObjFanin0(a).MarkB && abc.Abc_ObjFanin1(a).MarkB {
		*leaf = b
		*mark1 = a
		*mark2 = nil
		return 1
	}
	if abc.Abc_ObjIsCi(b) == 0 && abc.Abc_ObjFanin0(b).MarkB && abc.Abc_ObjFanin1(b).MarkB {
		*leaf = a
		*mark1 = b
		*mark2 = nil
		return 1
	}
	if abc.Abc_ObjIsCi(b) != 0 {
		a, b = b, a
	}
	var toAdd *Abc_Obj_t
	m1 := (*Abc_Obj_t)(nil)
	if abc.Abc_ObjIsCi(a) != 0 {
		toAdd = a
	} else {
		m1 = a
		for _, g := range []*Abc_Obj_t{abc.Abc_ObjFanin0(a), abc.Abc_ObjFanin1(a)} {
			if !g.MarkB {
				if toAdd != nil && toAdd != g {
					return 999
				}
				toAdd = g
			}
		}
	}
	m2 := b
	for _, g := range []*Abc_Obj_t{abc.Abc_ObjFanin0(b), abc.Abc_ObjFanin1(b)} {
		if !g.MarkB {
			if toAdd != nil && toAdd != g {
				return 999
			}
			toAdd = g
		}
	}
	if toAdd == nil {
		return 999
	}
	*leaf, *mark1, *mark2 = toAdd, m1, m2
	return 1
}

func Abc_NodeFindCut(m *Abc_ManCut_t, root *Abc_Obj_t, contain int32) *Vec_Ptr_t {
	if root == nil || abc.Abc_ObjIsComplement(root) != 0 || abc.Abc_ObjIsNode(root) == 0 {
		panic("Abc_NodeFindCut: root must be a regular node")
	}
	vec.Vec_PtrClear(m.vVisited)
	vec.Vec_PtrPush(m.vVisited, root)
	f0, f1 := abc.Abc_ObjFanin0(root), abc.Abc_ObjFanin1(root)
	vec.Vec_PtrPush(m.vVisited, f0)
	vec.Vec_PtrPush(m.vVisited, f1)
	root.MarkB = true
	f0.MarkB = true
	f1.MarkB = true
	vec.Vec_PtrClear(m.vNodeLeaves)
	vec.Vec_PtrPush(m.vNodeLeaves, f0)
	vec.Vec_PtrPush(m.vNodeLeaves, f1)
	for Abc_NodeBuildCutLevelOne_int(m.vVisited, m.vNodeLeaves, m.nNodeSizeMax, m.nNodeFanStop) != 0 {
	}
	assert(vec.Vec_PtrSize(m.vNodeLeaves) <= int(m.nNodeSizeMax), "node cut exceeds configured size")
	if contain == 0 {
		Abc_NodesUnmarkB(m.vVisited)
		return m.vNodeLeaves
	}
	assert(m.nNodeSizeMax < m.nConeSizeMax, "containing cut limit must exceed node cut limit")
	vec.Vec_PtrClear(m.vConeLeaves)
	for _, x := range vec.Vec_PtrArray(m.vNodeLeaves) {
		vec.Vec_PtrPush(m.vConeLeaves, x)
	}
	for Abc_NodeBuildCutLevelOne_int(m.vVisited, m.vConeLeaves, m.nConeSizeMax, m.nConeFanStop) != 0 {
	}
	assert(vec.Vec_PtrSize(m.vConeLeaves) <= int(m.nConeSizeMax), "containing cut exceeds configured size")
	Abc_NodesUnmarkB(m.vVisited)
	return m.vNodeLeaves
}

func Abc_NodeBuildCutLevelOne_int(visited, leaves *Vec_Ptr_t, sizeLimit, faninLimit int32) int32 {
	best, costBest := -1, int32(100)
	for i, x := range vec.Vec_PtrArray(leaves) {
		n := x.(*Abc_Obj_t)
		cost := Abc_NodeGetLeafCostOne(n, faninLimit)
		if cost < costBest || cost == costBest && best >= 0 && n.Level > vec.Vec_PtrEntry(leaves, best).(*Abc_Obj_t).Level {
			best, costBest = i, cost
		}
		if costBest == 0 {
			break
		}
	}
	if best < 0 || int32(vec.Vec_PtrSize(leaves))-1+costBest > sizeLimit {
		return 0
	}
	n := vec.Vec_PtrEntry(leaves, best).(*Abc_Obj_t)
	assert(abc.Abc_ObjIsNode(n) != 0, "selected cut leaf is not a node")
	vec.Vec_PtrRemove(leaves, n)
	for _, child := range []*Abc_Obj_t{abc.Abc_ObjFanin0(n), abc.Abc_ObjFanin1(n)} {
		if !child.MarkB {
			child.MarkB = true
			vec.Vec_PtrPush(leaves, child)
			vec.Vec_PtrPush(visited, child)
		}
	}
	assert(vec.Vec_PtrSize(leaves) <= int(sizeLimit), "cut exceeds configured size")
	return 1
}

func Abc_NodeBuildCutLevelTwo_int(visited, leaves *Vec_Ptr_t, faninLimit int32) int32 {
	var n, leaf, m1, m2 *Abc_Obj_t
	cost := int32(999)
	for _, x := range vec.Vec_PtrArray(leaves) {
		n = x.(*Abc_Obj_t)
		cost = Abc_NodeGetLeafCostTwo(n, faninLimit, &leaf, &m1, &m2)
		if cost < 2 {
			break
		}
	}
	if cost > 2 {
		return 0
	}
	vec.Vec_PtrRemove(leaves, n)
	if leaf != nil {
		assert(!leaf.MarkB, "candidate leaf already marked")
		leaf.MarkB = true
		vec.Vec_PtrPush(leaves, leaf)
		vec.Vec_PtrPush(visited, leaf)
	}
	for _, x := range []*Abc_Obj_t{m1, m2} {
		if x != nil {
			assert(!x.MarkB, "candidate node already marked")
			x.MarkB = true
			vec.Vec_PtrPush(visited, x)
		}
	}
	return 1
}

func Abc_NodeConeCollect(roots []*Abc_Obj_t, leaves, visited *Vec_Ptr_t, includeFanins int32) {
	Abc_NodesMark(leaves)
	vec.Vec_PtrClear(visited)
	if includeFanins != 0 {
		for _, x := range vec.Vec_PtrArray(leaves) {
			vec.Vec_PtrPush(visited, x)
		}
	}
	for _, r := range roots {
		Abc_NodeConeMarkCollect_rec(r, visited)
	}
	Abc_NodesUnmark(leaves)
	Abc_NodesUnmark(visited)
}
func Abc_NodeConeMarkCollect_rec(n *Abc_Obj_t, visited *Vec_Ptr_t) {
	if n.MarkA {
		return
	}
	if abc.Abc_ObjIsNode(n) != 0 {
		Abc_NodeConeMarkCollect_rec(abc.Abc_ObjFanin0(n), visited)
		Abc_NodeConeMarkCollect_rec(abc.Abc_ObjFanin1(n), visited)
	}
	assert(!n.MarkA, "duplicate DFS visit")
	n.MarkA = true
	vec.Vec_PtrPush(visited, n)
}

func Abc_NtkManCutStart(nodeMax, coneMax, nodeFanStop, coneFanStop int32) *Abc_ManCut_t {
	return &Abc_ManCut_t{nodeMax, coneMax, nodeFanStop, coneFanStop, vec.Vec_PtrAlloc(100), vec.Vec_PtrAlloc(100), vec.Vec_PtrAlloc(100), vec.Vec_PtrAlloc(100), vec.Vec_VecAlloc(100)}
}
func Abc_NtkManCutStop(m *Abc_ManCut_t) {
	if m == nil {
		return
	}
	vec.Vec_PtrFree(m.vNodeLeaves)
	vec.Vec_PtrFree(m.vConeLeaves)
	vec.Vec_PtrFree(m.vVisited)
	vec.Vec_PtrFree(m.vNodesTfo)
	vec.Vec_VecFree(m.vLevels)
}
func Abc_NtkManCutReadCutLarge(m *Abc_ManCut_t) *Vec_Ptr_t { return m.vConeLeaves }
func Abc_NtkManCutReadCutSmall(m *Abc_ManCut_t) *Vec_Ptr_t { return m.vNodeLeaves }
func Abc_NtkManCutReadVisited(m *Abc_ManCut_t) *Vec_Ptr_t  { return m.vVisited }

func Abc_NodeCollectTfoCands(m *Abc_ManCut_t, root *Abc_Obj_t, leaves *Vec_Ptr_t, levelMax int32) *Vec_Ptr_t {
	n := root.Network()
	assert(abc.Abc_NtkIsStrash(n) != 0, "TFO requires strashed network")
	for i := 0; i < vec.Vec_VecSize(m.vLevels); i++ {
		assert(vec.Vec_PtrSize(vec.Vec_VecEntry(m.vLevels, i)) == 0, "TFO level scratch not empty")
	}
	abc.Abc_NtkIncrementTravId(n)
	levelMin := -1
	for _, x := range vec.Vec_PtrArray(leaves) {
		leaf := x.(*Abc_Obj_t)
		if leaf.Level > int(levelMax) {
			continue
		}
		abc.Abc_NodeSetTravIdCurrent(leaf)
		vec.Vec_VecPush(m.vLevels, leaf.Level, leaf)
		levelMin = max(levelMin, leaf.Level)
	}
	assert(levelMin >= 0, "no cut leaves under level bound")
	abc.Abc_NodeMffcLabelAig(root)
	vec.Vec_PtrClear(m.vNodesTfo)
	for level := levelMin; level <= int(levelMax); level++ {
		if level >= vec.Vec_VecSize(m.vLevels) {
			break
		}
		v := vec.Vec_VecEntry(m.vLevels, level)
		for _, x := range vec.Vec_PtrArray(v) {
			node := x.(*Abc_Obj_t)
			if abc.Abc_NodeIsTravIdCurrent(node) == 0 {
				if abc.Abc_NodeIsTravIdCurrent(abc.Abc_ObjFanin0(node)) == 0 || abc.Abc_NodeIsTravIdCurrent(abc.Abc_ObjFanin1(node)) == 0 {
					continue
				}
				vec.Vec_PtrPush(m.vNodesTfo, node)
				abc.Abc_NodeSetTravIdCurrent(node)
			}
			for _, fo := range node.Fanouts() {
				if abc.Abc_ObjIsCo(fo) != 0 || fo.Level > int(levelMax) || abc.Abc_NodeIsTravIdCurrent(fo) != 0 {
					continue
				}
				vec.Vec_VecPushUnique(m.vLevels, fo.Level, fo)
			}
		}
		vec.Vec_PtrClear(v)
	}
	return m.vNodesTfo
}

func Abc_NodeConeBdd(dd *core.DdManager, vars []*core.DdNode, root *Abc_Obj_t, leaves, visited *Vec_Ptr_t) *core.DdNode {
	Abc_NodeConeCollect([]*Abc_Obj_t{root}, leaves, visited, 0)
	for i, x := range vec.Vec_PtrArray(leaves) {
		x.(*Abc_Obj_t).CopyData = vars[i]
	}
	var result *core.DdNode
	for _, x := range vec.Vec_PtrArray(visited) {
		n := x.(*Abc_Obj_t)
		f0 := abcBddNotCond(abc.Abc_ObjFanin0(n).CopyData.(*core.DdNode), abc.Abc_ObjFaninC0(n))
		f1 := abcBddNotCond(abc.Abc_ObjFanin1(n).CopyData.(*core.DdNode), abc.Abc_ObjFaninC1(n))
		result = core.Cudd_bddAnd(dd, f0, f1)
		if result == nil {
			return nil
		}
		core.Cudd_Ref(result)
		n.CopyData = result
	}
	if result == nil {
		return nil
	}
	core.Cudd_Ref(result)
	for _, x := range vec.Vec_PtrArray(visited) {
		core.Cudd_RecursiveDeref(dd, x.(*Abc_Obj_t).CopyData.(*core.DdNode))
	}
	core.Cudd_Deref(result)
	return result
}
func abcBddNotCond(f *core.DdNode, c int32) *core.DdNode { return core.Cudd_NotCond(f, c != 0) }

func Abc_NodeConeDcs(dd *core.DdManager, varsX, varsY []*core.DdNode, leaves, roots, visited *Vec_Ptr_t) *core.DdNode {
	rootList := make([]*Abc_Obj_t, vec.Vec_PtrSize(roots))
	for i, x := range vec.Vec_PtrArray(roots) {
		rootList[i] = x.(*Abc_Obj_t)
	}
	Abc_NodeConeCollect(rootList, leaves, visited, 0)
	for i, x := range vec.Vec_PtrArray(leaves) {
		x.(*Abc_Obj_t).CopyData = varsX[i]
	}
	for _, x := range vec.Vec_PtrArray(visited) {
		n := x.(*Abc_Obj_t)
		a := abcBddNotCond(abc.Abc_ObjFanin0(n).CopyData.(*core.DdNode), abc.Abc_ObjFaninC0(n))
		b := abcBddNotCond(abc.Abc_ObjFanin1(n).CopyData.(*core.DdNode), abc.Abc_ObjFaninC1(n))
		f := core.Cudd_bddAnd(dd, a, b)
		if f == nil {
			return nil
		}
		core.Cudd_Ref(f)
		n.CopyData = f
	}
	trans := dd.One
	core.Cudd_Ref(trans)
	for i, x := range vec.Vec_PtrArray(roots) {
		n := x.(*Abc_Obj_t)
		f := core.Cudd_bddXnor(dd, n.CopyData.(*core.DdNode), varsY[i])
		if f == nil {
			return nil
		}
		core.Cudd_Ref(f)
		old := trans
		trans = core.Cudd_bddAnd(dd, old, f)
		if trans == nil {
			return nil
		}
		core.Cudd_Ref(trans)
		core.Cudd_RecursiveDeref(dd, old)
		core.Cudd_RecursiveDeref(dd, f)
	}
	for _, x := range vec.Vec_PtrArray(visited) {
		core.Cudd_RecursiveDeref(dd, x.(*Abc_Obj_t).CopyData.(*core.DdNode))
	}
	cube := extrab.Extra_bddComputeRangeCube(dd, vec.Vec_PtrSize(roots), vec.Vec_PtrSize(roots)+vec.Vec_PtrSize(leaves))
	core.Cudd_Ref(cube)
	res := core.Cudd_Not(core.Cudd_bddExistAbstract(dd, trans, cube))
	core.Cudd_Ref(res)
	core.Cudd_RecursiveDeref(dd, cube)
	core.Cudd_RecursiveDeref(dd, trans)
	core.Cudd_Deref(res)
	return res
}
