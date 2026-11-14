package abci

import (
	"fmt"

	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/bdd/cudd/core"
	"github.com/glycerine/abcgo/misc/extra"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

const abcGlobalBddAttr = int(vec.VEC_ATTR_GLOBAL_BDD)

func abcGlobalBddAttrManager(n *abc.Network) *core.DdManager {
	if n == nil || abcGlobalBddAttr >= len(n.Attributes) || n.Attributes[abcGlobalBddAttr] == nil {
		return nil
	}
	manager, _ := vec.Vec_AttMan(n.Attributes[abcGlobalBddAttr]).(*core.DdManager)
	return manager
}

func Abc_NtkGlobalBddMan(n *Abc_Ntk_t) *core.DdManager { return abcGlobalBddAttrManager(n) }
func Abc_ObjGlobalBdd(o *Abc_Obj_t) *core.DdNode {
	if o == nil || o.Network() == nil {
		return nil
	}
	att := abcGlobalBddAttrManager(o.Network())
	if att == nil {
		return nil
	}
	value := vec.Vec_AttEntry(o.Network().Attributes[abcGlobalBddAttr], o.ID)
	bdd, _ := value.(*core.DdNode)
	return bdd
}
func Abc_ObjSetGlobalBdd(o *Abc_Obj_t, bdd *core.DdNode) {
	if o == nil || o.Network() == nil {
		return
	}
	att := abcGlobalBddAttrManager(o.Network())
	if att == nil {
		panic("Abc_ObjSetGlobalBdd: global BDD manager is not attached")
	}
	vec.Vec_AttWriteEntry(o.Network().Attributes[abcGlobalBddAttr], o.ID, bdd)
}

// Abc_NtkBuildGlobalBdds computes a shared BDD for each CO in an AIG network.
// As in ABC, internal BDD references can be dropped once all fanouts consumed them.
func Abc_NtkBuildGlobalBdds(n *Abc_Ntk_t, sizeMax, dropInternal, reorder, reverse, verbose int32) *core.DdManager {
	if n == nil || abc.Abc_NtkIsStrash(n) == 0 {
		return nil
	}
	if abcGlobalBddAttrManager(n) != nil {
		panic("Abc_NtkBuildGlobalBdds: global BDDs already exist")
	}
	if mgr, ok := n.FunctionManager.(*abc.Abc_Aig_t); ok {
		abc.Abc_AigCleanup(mgr)
	}
	d := core.Cudd_Init(uint32(abc.Abc_NtkCiNum(n)), 0, core.CUDD_UNIQUE_SLOTS, core.CUDD_CACHE_SLOTS, 0)
	if d == nil {
		return nil
	}
	att := vec.Vec_AttAlloc(int(abc.Abc_NtkObjNumMax(n))+1, d,
		func(v any) {
			if m, ok := v.(*core.DdManager); ok {
				core.Cudd_Quit(m)
			}
		}, nil,
		func(m, v any) {
			if dd, ok := m.(*core.DdManager); ok {
				if node, ok := v.(*core.DdNode); ok {
					core.Cudd_RecursiveDeref(dd, node)
				}
			}
		})
	for len(n.Attributes) <= abcGlobalBddAttr {
		n.Attributes = append(n.Attributes, nil)
	}
	n.Attributes[abcGlobalBddAttr] = att
	if reorder != 0 {
		core.Cudd_AutodynEnable(d, core.CUDD_REORDER_SYMM_SIFT)
	}
	remaining := make(map[*abc.Object]int32)
	for _, obj := range n.Objects() {
		remaining[obj] = abc.Abc_ObjFanoutNum(obj)
	}
	var constant *abc.Object
	for _, obj := range n.Objects() {
		if obj.Kind == abc.ObjectConst1 {
			constant = obj
			break
		}
	}
	if constant != nil && remaining[constant] > 0 {
		core.Cudd_Ref(d.One)
		Abc_ObjSetGlobalBdd(constant, d.One)
	}
	ciIndex := 0
	for _, ci := range n.Objects() {
		if !ci.IsCI() {
			continue
		}
		idx := ciIndex
		ciIndex++
		if remaining[ci] == 0 {
			continue
		}
		if reverse != 0 {
			idx = int(abc.Abc_NtkCiNum(n)) - 1 - idx
		}
		if idx < 0 || idx >= len(d.Vars) {
			idx = ci.TerminalIndex
		}
		if idx < 0 || idx >= len(d.Vars) {
			continue
		}
		v := d.Vars[idx]
		core.Cudd_Ref(v)
		Abc_ObjSetGlobalBdd(ci, v)
	}
	progress := extra.Extra_ProgressBarStart(nil, abc.Abc_NtkNodeNum(n))
	counter := int32(0)
	for _, co := range n.Objects() {
		if !co.IsCO() {
			continue
		}
		fanins := co.Fanins()
		if len(fanins) == 0 {
			continue
		}
		f := fanins[0]
		bdd := abcNodeGlobalBddRec(d, f.Object, sizeMax, dropInternal != 0, remaining, progress, &counter, verbose != 0)
		if bdd == nil {
			Abc_NtkFreeGlobalBdds(n, 1)
			return nil
		}
		bdd = core.Cudd_NotCond(bdd, f.Complement)
		core.Cudd_Ref(bdd)
		Abc_ObjSetGlobalBdd(co, bdd)
	}
	extra.Extra_ProgressBarStop(progress)
	if reorder != 0 {
		core.Cudd_ReduceHeap(d, core.CUDD_REORDER_SYMM_SIFT, 1)
		core.Cudd_AutodynDisable(d)
	}
	return d
}

func abcNodeGlobalBddRec(d *core.DdManager, node *abc.Object, limit int32, drop bool, remaining map[*abc.Object]int32, progress *extra.ProgressBar, counter *int32, verbose bool) *core.DdNode {
	if node == nil {
		return nil
	}
	if old := Abc_ObjGlobalBdd(node); old != nil {
		remaining[node]--
		if drop && remaining[node] == 0 {
			core.Cudd_Deref(old)
			Abc_ObjSetGlobalBdd(node, nil)
		}
		return old
	}
	if limit > 0 && int64(core.Cudd_ReadKeys(d)-core.Cudd_ReadDead(d)) > int64(limit) {
		extra.Extra_ProgressBarStop(progress)
		if verbose {
			fmt.Printf("The number of live nodes reached %d.\n", limit)
		}
		return nil
	}
	fanins := node.Fanins()
	if len(fanins) != 2 {
		return nil
	}
	f0 := abcNodeGlobalBddRec(d, fanins[0].Object, limit, drop, remaining, progress, counter, verbose)
	if f0 == nil {
		return nil
	}
	core.Cudd_Ref(f0)
	f1 := abcNodeGlobalBddRec(d, fanins[1].Object, limit, drop, remaining, progress, counter, verbose)
	if f1 == nil {
		core.Cudd_RecursiveDeref(d, f0)
		return nil
	}
	core.Cudd_Ref(f1)
	f0 = core.Cudd_NotCond(f0, fanins[0].Complement)
	f1 = core.Cudd_NotCond(f1, fanins[1].Complement)
	f := core.Cudd_bddAndLimit(d, f0, f1, uint32(max(int32(0), limit)))
	core.Cudd_RecursiveDeref(d, f0)
	core.Cudd_RecursiveDeref(d, f1)
	if f == nil {
		return nil
	}
	core.Cudd_Ref(f)
	Abc_ObjSetGlobalBdd(node, f)
	*counter++
	extra.Extra_ProgressBarUpdate(progress, *counter, "")
	remaining[node]--
	if drop && remaining[node] == 0 {
		core.Cudd_Deref(f)
		Abc_ObjSetGlobalBdd(node, nil)
	}
	return f
}

func Abc_NtkFreeGlobalBdds(n *Abc_Ntk_t, freeManager int32) *core.DdManager {
	if n == nil || abcGlobalBddAttrManager(n) == nil {
		return nil
	}
	value := abc.Abc_NtkAttrFree(n, int32(abcGlobalBddAttr), freeManager)
	manager, _ := value.(*core.DdManager)
	return manager
}
func Abc_NtkSizeOfGlobalBdds(n *Abc_Ntk_t) int32 {
	if n == nil {
		return 0
	}
	var roots []*core.DdNode
	for _, o := range n.Objects() {
		if o.IsCO() {
			if f := Abc_ObjGlobalBdd(o); f != nil {
				roots = append(roots, f)
			}
		}
	}
	seen := make(map[*core.DdNode]struct{})
	var visit func(*core.DdNode)
	visit = func(f *core.DdNode) {
		if f == nil {
			return
		}
		f = core.Cudd_Regular(f)
		if _, ok := seen[f]; ok {
			return
		}
		seen[f] = struct{}{}
		if !core.Cudd_IsConstant(f) {
			visit(f.T)
			visit(f.E)
		}
	}
	for _, f := range roots {
		visit(f)
	}
	return int32(len(seen))
}

// C source currently leaves this experimental metric disabled.
func Abc_NtkSpacePercentage(node *Abc_Obj_t) float64 { return 0 }

// BDD-to-MUX conversion is available for local BDD networks and shared AIG BDDs.
func Abc_NtkBddToMuxes(n *Abc_Ntk_t, global int32, limit int32) *Abc_Ntk_t {
	if n == nil {
		return nil
	}
	out := abc.Abc_NtkStartFrom(n, abc.ABC_NTK_LOGIC, abc.ABC_FUNC_SOP)
	finalize := false
	if global != 0 {
		dd := Abc_NtkBuildGlobalBdds(n, limit, 1, 1, 0, 0)
		if dd == nil {
			abc.Abc_NtkDelete(out)
			return nil
		}
		mapping := make(map[*core.DdNode]*abc.Object)
		ciIndex := 0
		for _, ci := range out.Objects() {
			if ci.IsCI() && ciIndex < len(dd.Vars) {
				mapping[dd.Vars[ciIndex]] = ci
				ciIndex++
			}
		}
		for _, co := range n.Objects() {
			if !co.IsCO() {
				continue
			}
			b := Abc_ObjGlobalBdd(co)
			if b == nil {
				continue
			}
			fanins := co.Fanins()
			if len(fanins) == 0 {
				continue
			}
			dst := fanins[0].Object.Copy
			x := abcBddToMuxRec(dd, core.Cudd_Regular(b), out, mapping)
			if core.Cudd_IsComplement(b) {
				x = abc.Abc_NtkCreateNodeInv(out, x)
			}
			abc.Abc_ObjAddFanin(dst, x)
		}
		Abc_NtkFreeGlobalBdds(n, 0)
		core.Cudd_Quit(dd)
		abc.Abc_NtkCleanCopy(n)
	} else {
		if abc.Abc_NtkIsBddLogic(n) == 0 {
			abc.Abc_NtkDelete(out)
			return nil
		}
		finalize = true
		for _, old := range abc.Abc_NtkDfs(n, 0) {
			if old.Kind != abc.ObjectNode {
				continue
			}
			b, ok := old.Data.(*core.DdNode)
			if !ok || b == nil {
				abc.Abc_NtkCleanCopy(n)
				abc.Abc_NtkDelete(out)
				return nil
			}
			cache := make(map[*core.DdNode]*abc.Object)
			for i, edge := range old.Fanins() {
				cache[core.Cudd_bddIthVar(n.FunctionManager.(*core.DdManager), i)] = edge.Object.Copy
			}
			x := abcBddToMuxRec(n.FunctionManager.(*core.DdManager), core.Cudd_Regular(b), out, cache)
			if core.Cudd_IsComplement(b) {
				x = abc.Abc_NtkCreateNodeInv(out, x)
			}
			old.Copy = x
		}
	}
	if finalize {
		abc.Abc_NtkFinalize(n, out)
		abc.Abc_NtkCleanCopy(n)
	}
	if abc.Abc_NtkCheck(out) == 0 {
		abc.Abc_NtkDelete(out)
		return nil
	}
	return out
}
func abcBddToMuxRec(dd *core.DdManager, f *core.DdNode, n *abc.Network, cache map[*core.DdNode]*abc.Object) *abc.Object {
	if f == nil {
		return nil
	}
	f = core.Cudd_Regular(f)
	if f == dd.One {
		return abc.Abc_NtkCreateNodeConst1(n)
	}
	if x := cache[f]; x != nil {
		return x
	}
	e := f.E
	t := f.T
	low := abcBddToMuxRec(dd, core.Cudd_Regular(e), n, cache)
	if core.Cudd_IsComplement(e) {
		low = abc.Abc_NtkCreateNodeInv(n, low)
	}
	high := abcBddToMuxRec(dd, core.Cudd_Regular(t), n, cache)
	control := cache[dd.Vars[f.Index]]
	x := abc.Abc_NtkCreateNodeMux(n, control, high, low)
	cache[f] = x
	return x
}

func Abc_NtkDeriveFromBdd(dd *core.DdManager, f *core.DdNode, poName string, piNames *vec.Vec_Ptr_t) *abc.Network {
	if dd == nil || f == nil {
		return nil
	}
	if poName == "" {
		poName = "F"
	}
	fakeNames := piNames == nil
	if fakeNames {
		piNames = abc.Abc_NodeGetFakeNames(int32(len(dd.Vars)))
		defer abc.Abc_NodeFreeNames(piNames)
	}
	supp := dd.Support(f)
	for _, i := range supp {
		if i >= vec.Vec_PtrSize(piNames) {
			return nil
		}
	}
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_LOGIC, abc.ABC_FUNC_BDD, 1)
	n.Name = poName
	dst, ok := n.FunctionManager.(*core.DdManager)
	if !ok || dst == nil {
		return nil
	}
	core.Cudd_bddIthVar(dst, vec.Vec_PtrSize(piNames))
	pis := make([]*abc.Object, vec.Vec_PtrSize(piNames))
	for i := range pis {
		pis[i] = abc.Abc_NtkCreatePi(n)
		name, _ := vec.Vec_PtrEntry(piNames, i).(string)
		pis[i].SetName(name)
	}
	node := abc.Abc_NtkCreateNode(n)
	node.Data = core.Cudd_bddTransfer(dd, dst, f)
	if node.Data == nil {
		return nil
	}
	core.Cudd_Ref(node.Data.(*core.DdNode))
	for _, pi := range pis {
		abc.Abc_ObjAddFanin(node, pi)
	}
	po := abc.Abc_NtkCreatePo(n)
	abc.Abc_ObjAddFanin(po, node)
	po.SetName(poName)
	abc.Abc_NtkMinimumBase(n)
	if abc.Abc_NtkCheck(n) == 0 {
		return nil
	}
	return n
}

func Abc_NtkBddImplicationTest() {
	const nVars, nImps = 200, 200
	start := util.Abc_Clock()
	dd := core.Cudd_Init(nVars, 0, core.CUDD_UNIQUE_SLOTS, core.CUDD_CACHE_SLOTS, 0)
	if dd == nil {
		return
	}
	core.Cudd_AutodynEnable(dd, core.CUDD_REORDER_SIFT)
	sum := dd.LogicZero
	core.Cudd_Ref(sum)
	rng := newGlibcRand(1) // C's initial libc RNG state is equivalent to srand(1).
	for i := 0; i < nImps; i++ {
		fmt.Print(".")
		i0, i1 := abcBddImplicationVars(rng, nVars)
		imp := core.Cudd_bddAnd(dd, dd.Vars[i0], dd.Vars[i1])
		if imp == nil {
			break
		}
		core.Cudd_Ref(imp)
		old := sum
		sum = core.Cudd_bddOr(dd, old, imp)
		if sum == nil {
			core.Cudd_RecursiveDeref(dd, old)
			core.Cudd_RecursiveDeref(dd, imp)
			break
		}
		core.Cudd_Ref(sum)
		core.Cudd_RecursiveDeref(dd, old)
		core.Cudd_RecursiveDeref(dd, imp)
	}
	fmt.Printf("The BDD before = %d.\n", core.Cudd_DagSize(sum))
	core.Cudd_ReduceHeap(dd, core.CUDD_REORDER_SIFT, 1)
	fmt.Printf("The BDD after  = %d.\n", core.Cudd_DagSize(sum))
	util.Abc_PrintTime(1, "Time", util.Abc_Clock()-start)
	core.Cudd_RecursiveDeref(dd, sum)
	core.Cudd_Quit(dd)
}

func abcBddImplicationVars(rng *glibcRand, nVars uint32) (int, int) {
	return int(rng.next() % nVars), int(rng.next() % nVars)
}
