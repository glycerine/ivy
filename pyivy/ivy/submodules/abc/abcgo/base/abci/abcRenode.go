package abci

import (
	"sync"
	"unsafe"

	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/bdd/cudd/core"
	"github.com/glycerine/abcgo/bdd/reo"
	"github.com/glycerine/abcgo/bool/kit"
	ifpkg "github.com/glycerine/abcgo/map/if"
	"github.com/glycerine/abcgo/misc/util"
)

// Renoding keeps the source module's shared work areas. Serialize calls so the
// callback state cannot be observed by another mapping invocation mid-flight.
var renodeState struct {
	sync.Mutex
	dd       *core.DdManager
	reorder  *reo.Manager
	memory   []uint32
	memory2  []uint32
	dsdCount int32
}

func Abc_NtkRenode(pNtk *Abc_Ntk_t, nFaninMax, nCubeMax, nFlowIters, nAreaIters, fArea, fUseBdds, fUseSops, fUseCnfs, fUseMv, fVerbose int32) *Abc_Ntk_t {
	renodeState.Lock()
	defer renodeState.Unlock()

	if pNtk == nil {
		return nil
	}
	if abc.Abc_NtkGetChoiceNum(pNtk) != 0 {
		util.Abc_Print(1, "Performing renoding with choices.\n")
	}
	renodeState.dsdCount = 0
	area := fArea
	if fUseCnfs != 0 {
		area = 1
	}

	var manager *core.DdManager
	var reorder *reo.Manager
	if fUseBdds != 0 {
		manager = core.NewBDDManager(int(nFaninMax))
		if manager == nil {
			return nil
		}
		reorder = reo.Extra_ReorderInit(int(nFaninMax), 100)
		renodeState.dd, renodeState.reorder = manager, reorder
	} else {
		renodeState.memory = make([]uint32, 0, 1<<16)
		renodeState.memory2 = make([]uint32, 0, 1<<16)
	}

	cost := renodeCostAig
	if fUseBdds != 0 {
		cost = renodeCostBdd
	} else if fUseSops != 0 {
		cost = renodeCostSop
	} else if fUseCnfs != 0 {
		cost = renodeCostCnf
	} else if fUseMv != 0 {
		cost = renodeCostMv
	}
	pars := &ifpkg.If_Par_t{}
	pars.SetRenodeConfig(nFaninMax, nCubeMax, nFlowIters, nAreaIters, area, fUseBdds, fUseSops, fUseCnfs, fUseMv, fVerbose, reorder, cost)
	result := Abc_NtkIf(pNtk, pars)

	if fUseBdds != 0 {
		if manager != nil {
			core.Cudd_Quit(manager)
		}
		if reorder != nil {
			reo.Extra_ReorderQuit(reorder)
		}
		renodeState.dd, renodeState.reorder = nil, nil
	} else {
		renodeState.memory, renodeState.memory2 = nil, nil
	}
	return result
}

func renodeTruth(p *ifpkg.If_Man_t, cut *ifpkg.If_Cut_t) []uint32 {
	n := ifpkg.If_CutLeaveNum(cut)
	if n < 0 {
		return nil
	}
	words := kit.Kit_TruthWordNum(n)
	if words == 0 {
		return []uint32{0}
	}
	return unsafe.Slice(ifpkg.If_CutTruth(p, cut), int(words))
}

func renodePerm(cut *ifpkg.If_Cut_t) []byte { return ifpkg.If_CutPerm(cut) }

func renodeCostAig(p *ifpkg.If_Man_t, cut *ifpkg.If_Cut_t) int32 {
	perm, truth := renodePerm(cut), renodeTruth(p, cut)
	n := ifpkg.If_CutLeaveNum(cut)
	graph := kit.Kit_TruthToGraph(truth, n, &renodeState.memory)
	if graph == nil {
		for i := int32(0); i < n; i++ {
			perm[i] = 100
		}
		return ifpkg.IF_COST_MAX
	}
	nodes := kit.Kit_GraphNodeNum(graph)
	var root *kit.Kit_Node_t
	if kit.Kit_GraphIsConst(graph) == 0 {
		root = kit.Kit_GraphNodeLast(graph)
	}
	var depths [ifpkg.IF_MAX_LUTSIZE]int32
	kit.Kit_GraphLeafDepths(graph, root, depths[:n])
	for i, d := range depths[:n] {
		perm[i] = byte(d)
	}
	kit.Kit_GraphFree(graph)
	return nodes
}

// renodeTruthToBDD is Kit_TruthToBdd( dd, pTruth, nVars, 0 ) (kitBdd.c):
// the least significant variable of the table is the topmost one.
func renodeTruthToBDD(dd *core.DdManager, truth []uint32, n int32) *core.DdNode {
	if dd == nil || n < 0 || n > int32(len(dd.Vars)) {
		return nil
	}
	if renodeIdentityOrder(dd, n) {
		return renodeTruthToBDDTopDown(dd, truth, n)
	}
	var build func(int32, int32) *core.DdNode
	build = func(bit, vars int32) *core.DdNode {
		if vars <= 5 {
			mask := ^uint32(0) >> (32 - (1 << vars))
			t := (truth[bit>>5] >> (bit & 31)) & mask
			if t == 0 {
				return dd.LogicZero
			}
			if t == mask {
				return dd.One
			}
		}
		lo := build(bit, vars-1)
		if lo == nil {
			return nil
		}
		core.Cudd_Ref(lo)
		hi := build(bit+(1<<uint(vars-1)), vars-1)
		if hi == nil {
			core.Cudd_RecursiveDeref(dd, lo)
			return nil
		}
		core.Cudd_Ref(hi)
		f := core.Cudd_bddIte(dd, dd.Vars[vars-1], hi, lo)
		if f != nil {
			core.Cudd_Ref(f)
		}
		core.Cudd_RecursiveDeref(dd, lo)
		core.Cudd_RecursiveDeref(dd, hi)
		if f != nil {
			core.Cudd_Deref(f)
		}
		return f
	}
	return build(0, n)
}

func renodeIdentityOrder(dd *core.DdManager, n int32) bool {
	for i := int32(0); i < n; i++ {
		if dd.Perm[i] != int(i) {
			return false
		}
	}
	return true
}

// renodeTruthToBDDTopDown builds the same (canonical) BDD as
// Kit_TruthToBdd when variable i is on level i: node by node from the top
// variable down, instead of by ITEs from the bottom variable up. The
// subfunction under the assignment a of the variables above level v is
// made of the table bits a + k*2^v.
func renodeTruthToBDDTopDown(dd *core.DdManager, truth []uint32, n int32) *core.DdNode {
	bitAt := func(i int32) bool { return truth[i>>5]>>(i&31)&1 != 0 }
	var build func(v, a int32) *core.DdNode
	build = func(v, a int32) *core.DdNode {
		// constant subfunction?
		first := bitAt(a)
		constant := true
		for i := a + (1 << v); i < 1<<n; i += 1 << v {
			if bitAt(i) != first {
				constant = false
				break
			}
		}
		if constant {
			if first {
				return dd.One
			}
			return dd.LogicZero
		}
		lo := build(v+1, a)
		core.Cudd_Ref(lo)
		hi := build(v+1, a|1<<v)
		core.Cudd_Ref(hi)
		f := core.Cudd_UniqueInter(dd, int(v), hi, lo)
		// f holds lo and hi: dropping these references frees nothing
		core.Cudd_Deref(lo)
		core.Cudd_Deref(hi)
		return f
	}
	return build(0, 0)
}

func renodeCostBdd(p *ifpkg.If_Man_t, cut *ifpkg.If_Cut_t) int32 {
	perm, n := renodePerm(cut), ifpkg.If_CutLeaveNum(cut)
	for i := int32(0); i < n; i++ {
		perm[i] = 0x9c // the byte representation of source char value -100
	}
	truth := renodeTruth(p, cut)
	f := renodeTruthToBDD(renodeState.dd, truth, n)
	if f == nil {
		return ifpkg.IF_COST_MAX
	}
	core.Cudd_Ref(f)
	order := make([]int, n)
	for i := range order {
		order[i] = -100
	}
	optimized := reo.Extra_Reorder(renodeState.reorder, renodeState.dd, f, order)
	if optimized == nil {
		core.Cudd_RecursiveDeref(renodeState.dd, f)
		return ifpkg.IF_COST_MAX
	}
	core.Cudd_Ref(optimized)
	k := 0
	for i := int32(0); i < n; i++ {
		if order[i] >= 0 && order[i] < int(n) {
			perm[order[i]] = byte(k + 1)
			k++
		}
	}
	cost := int32(core.Cudd_DagSize(optimized) - 1)
	core.Cudd_RecursiveDeref(renodeState.dd, optimized)
	core.Cudd_RecursiveDeref(renodeState.dd, f)
	return cost
}

func renodeCostSop(p *ifpkg.If_Man_t, cut *ifpkg.If_Cut_t) int32 {
	perm, n := renodePerm(cut), ifpkg.If_CutLeaveNum(cut)
	for i := int32(0); i < n; i++ {
		perm[i] = 1
	}
	cover := &renodeState.memory
	if kit.Kit_TruthIsop(renodeTruth(p, cut), n, cover, 1) == -1 {
		return ifpkg.IF_COST_MAX
	}
	return int32(len(*cover))
}

func renodeCostCnf(p *ifpkg.If_Man_t, cut *ifpkg.If_Cut_t) int32 {
	perm, n := renodePerm(cut), ifpkg.If_CutLeaveNum(cut)
	for i := int32(0); i < n; i++ {
		perm[i] = 1
	}
	truth := renodeTruth(p, cut)
	off := make([]uint32, len(truth))
	kit.Kit_TruthNot(off, truth, n)
	if kit.Kit_TruthIsop(truth, n, &renodeState.memory, 0) == -1 {
		return ifpkg.IF_COST_MAX
	}
	positive := len(renodeState.memory)
	if kit.Kit_TruthIsop(off, n, &renodeState.memory, 0) == -1 {
		return ifpkg.IF_COST_MAX
	}
	return int32(positive + len(renodeState.memory))
}

func renodeCostMv(p *ifpkg.If_Man_t, cut *ifpkg.If_Cut_t) int32 {
	perm, n := renodePerm(cut), ifpkg.If_CutLeaveNum(cut)
	for i := int32(0); i < n; i++ {
		perm[i] = 1
	}
	truth := renodeTruth(p, cut)
	off := make([]uint32, len(truth))
	kit.Kit_TruthNot(off, truth, n)
	if kit.Kit_TruthIsop(truth, n, &renodeState.memory, 0) == -1 {
		return ifpkg.IF_COST_MAX
	}
	if kit.Kit_TruthIsop(off, n, &renodeState.memory2, 0) == -1 {
		return ifpkg.IF_COST_MAX
	}
	return abc.Abc_NodeEvalMvCost(n, renodeState.memory, renodeState.memory2)
}
