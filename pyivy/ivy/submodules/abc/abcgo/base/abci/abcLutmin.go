package abci

// LUT minimization routines translated from src/base/abci/abcLutmin.c.
import (
	"fmt"
	"math/bits"
	"sort"
	"unsafe"

	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/bdd/cudd/core"
	"github.com/glycerine/abcgo/bdd/extrab"
)

func lutManager(n *abc.Network) *core.DdManager {
	if n == nil {
		return nil
	}
	dd, _ := n.FunctionManager.(*core.DdManager)
	return dd
}

func lutAddNode(n *abc.Network, fanins ...*abc.Object) *abc.Object {
	obj := abc.Abc_NtkCreateNode(n)
	for _, fanin := range fanins {
		abc.Abc_ObjAddFanin(obj, fanin)
	}
	return obj
}

// Abc_NtkBddMux21 implements c ? d1 : d0 with the three-variable LUT BDD.
func Abc_NtkBddMux21(n *Abc_Ntk_t, fanins []*Abc_Obj_t) *Abc_Obj_t {
	if n == nil || len(fanins) < 3 {
		return nil
	}
	dd := lutManager(n)
	if dd == nil {
		return nil
	}
	obj := lutAddNode(n, fanins[0], fanins[1], fanins[2])
	obj.Data = core.Cudd_bddIte(dd, core.Cudd_bddIthVar(dd, 0), core.Cudd_bddIthVar(dd, 2), core.Cudd_bddIthVar(dd, 1))
	core.Cudd_Ref(obj.Data.(*core.DdNode))
	return obj
}

// Abc_NtkBddMux411 implements the six-input mux in one 6-LUT.
func Abc_NtkBddMux411(n *Abc_Ntk_t, fanins []*Abc_Obj_t) *Abc_Obj_t {
	if n == nil || len(fanins) < 6 {
		return nil
	}
	dd := lutManager(n)
	if dd == nil {
		return nil
	}
	obj := lutAddNode(n, fanins[:6]...)
	s := core.Cudd_bddIthVar(dd, 1)
	c0 := core.Cudd_bddIte(dd, s, core.Cudd_bddIthVar(dd, 3), core.Cudd_bddIthVar(dd, 2))
	core.Cudd_Ref(c0)
	c1 := core.Cudd_bddIte(dd, s, core.Cudd_bddIthVar(dd, 5), core.Cudd_bddIthVar(dd, 4))
	core.Cudd_Ref(c1)
	obj.Data = core.Cudd_bddIte(dd, core.Cudd_bddIthVar(dd, 0), c1, c0)
	core.Cudd_Ref(obj.Data.(*core.DdNode))
	core.Cudd_RecursiveDeref(dd, c0)
	core.Cudd_RecursiveDeref(dd, c1)
	return obj
}

// Abc_NtkBddMux412 implements the source's two-LUT decomposition.
func Abc_NtkBddMux412(n *Abc_Ntk_t, f []*Abc_Obj_t) *Abc_Obj_t {
	if n == nil || len(f) < 6 {
		return nil
	}
	dd := lutManager(n)
	if dd == nil {
		return nil
	}
	bot := lutAddNode(n, f[0], f[1], f[2], f[3])
	spin := core.Cudd_bddIthVar(dd, 0)
	c0 := core.Cudd_bddIte(dd, core.Cudd_bddIthVar(dd, 1), core.Cudd_bddIthVar(dd, 3), core.Cudd_bddIthVar(dd, 2))
	core.Cudd_Ref(c0)
	bot.Data = core.Cudd_bddIte(dd, spin, core.Cudd_bddIthVar(dd, 1), c0)
	core.Cudd_Ref(bot.Data.(*core.DdNode))
	core.Cudd_RecursiveDeref(dd, c0)
	top := lutAddNode(n, f[0], bot, f[4], f[5])
	c1 := core.Cudd_bddIte(dd, core.Cudd_bddIthVar(dd, 1), core.Cudd_bddIthVar(dd, 3), core.Cudd_bddIthVar(dd, 2))
	core.Cudd_Ref(c1)
	top.Data = core.Cudd_bddIte(dd, spin, c1, core.Cudd_bddIthVar(dd, 1))
	core.Cudd_Ref(top.Data.(*core.DdNode))
	core.Cudd_RecursiveDeref(dd, c1)
	return top
}

// Abc_NtkBddMux412a preserves the source's distinct five-input selector layout.
func Abc_NtkBddMux412a(n *Abc_Ntk_t, f []*Abc_Obj_t) *Abc_Obj_t {
	if n == nil || len(f) < 6 {
		return nil
	}
	dd := lutManager(n)
	if dd == nil {
		return nil
	}
	bot := lutAddNode(n, f[1], f[2], f[3])
	bot.Data = core.Cudd_bddIte(dd, core.Cudd_bddIthVar(dd, 0), core.Cudd_bddIthVar(dd, 2), core.Cudd_bddIthVar(dd, 1))
	core.Cudd_Ref(bot.Data.(*core.DdNode))
	top := lutAddNode(n, f[0], f[1], bot, f[4], f[5])
	c1 := core.Cudd_bddIte(dd, core.Cudd_bddIthVar(dd, 1), core.Cudd_bddIthVar(dd, 4), core.Cudd_bddIthVar(dd, 3))
	core.Cudd_Ref(c1)
	top.Data = core.Cudd_bddIte(dd, core.Cudd_bddIthVar(dd, 0), c1, core.Cudd_bddIthVar(dd, 2))
	core.Cudd_Ref(top.Data.(*core.DdNode))
	core.Cudd_RecursiveDeref(dd, c1)
	return top
}

// Abc_NtkBddMux413 implements a 4:1 mux with three 2:1 mux nodes.
func Abc_NtkBddMux413(n *Abc_Ntk_t, f []*Abc_Obj_t) *Abc_Obj_t {
	if len(f) < 6 {
		return nil
	}
	left := Abc_NtkBddMux21(n, []*Abc_Obj_t{f[1], f[2], f[3]})
	right := Abc_NtkBddMux21(n, []*Abc_Obj_t{f[1], f[4], f[5]})
	return Abc_NtkBddMux21(n, []*Abc_Obj_t{f[0], left, right})
}

// Abc_NtkBddCofactors_rec returns the selected cofactor over the first levels.
func Abc_NtkBddCofactors_rec(dd *core.DdManager, node *core.DdNode, code, level, levels int32) *core.DdNode {
	if dd == nil || node == nil || core.Cudd_IsConstant(node) || level == levels {
		return node
	}
	regular := core.Cudd_Regular(node)
	var low, high *core.DdNode
	if core.Cudd_ReadPerm(dd, int(regular.Index)) > int(level) {
		low, high = node, node
	} else {
		low, high = regular.E, regular.T
		if core.Cudd_IsComplement(node) {
			low, high = core.Cudd_Not(low), core.Cudd_Not(high)
		}
	}
	shift := uint(levels - 1 - level)
	if code&(int32(1)<<shift) != 0 {
		return Abc_NtkBddCofactors_rec(dd, high, code, level+1, levels)
	}
	return Abc_NtkBddCofactors_rec(dd, low, code, level+1, levels)
}

func Abc_NtkBddCofactors(dd *core.DdManager, node *core.DdNode, levels int32) []*core.DdNode {
	if dd == nil || node == nil || levels <= 0 || levels >= 10 {
		return nil
	}
	count := 1 << uint(levels)
	out := make([]*core.DdNode, count)
	for i := range out {
		out[i] = Abc_NtkBddCofactors_rec(dd, node, int32(i), 0, levels)
	}
	return out
}

// Abc_NtkCreateCofLut transfers one cofactor to a new node's local fanin order.
func Abc_NtkCreateCofLut(dst *Abc_Ntk_t, dd *core.DdManager, cof *core.DdNode, old *Abc_Obj_t, level int32) *Abc_Obj_t {
	if dst == nil || dd == nil || cof == nil || old == nil || int(level) >= len(old.Fanins()) {
		return nil
	}
	fanins := old.Fanins()
	obj := abc.Abc_NtkCreateNode(dst)
	for _, edge := range fanins[level:] {
		if edge.Object.Copy == nil {
			return nil
		}
		abc.Abc_ObjAddFanin(obj, edge.Object.Copy)
	}
	moved := extrab.Extra_bddMove(dd, cof, -int(level))
	if moved == nil {
		return nil
	}
	core.Cudd_Ref(moved)
	newDD := lutManager(dst)
	if newDD == nil {
		core.Cudd_RecursiveDeref(dd, moved)
		return nil
	}
	obj.Data = extrab.Extra_TransferLevelByLevel(dd, newDD, moved)
	core.Cudd_RecursiveDeref(dd, moved)
	if obj.Data == nil {
		return nil
	}
	core.Cudd_Ref(obj.Data.(*core.DdNode))
	return obj
}

// Abc_NtkBddCurtis performs the source's Ashenhurst-Curtis decomposition.
func Abc_NtkBddCurtis(dst *Abc_Ntk_t, old *Abc_Obj_t, cofactors, unique []*core.DdNode) *Abc_Obj_t {
	if dst == nil || old == nil || len(cofactors) == 0 || len(unique) == 0 {
		return nil
	}
	ddOld := lutManager(old.Network())
	ddNew := lutManager(dst)
	if ddOld == nil || ddNew == nil {
		return nil
	}
	k := int32(utilLog2(uint32(len(cofactors))))
	nBits := int32(utilLog2(uint32(len(unique))))
	if nBits+2 > k || k >= int32(len(old.Fanins())) || k > 10 || nBits > 10 {
		return nil
	}
	bound := old.Fanins()[:k]
	bits := make([]*core.DdNode, nBits)
	for i := range bits {
		bits[i] = ddNew.LogicZero
		core.Cudd_Ref(bits[i])
	}
	for c, cof := range cofactors {
		u := -1
		for i, item := range unique {
			if item == cof {
				u = i
				break
			}
		}
		if u < 0 {
			return nil
		}
		for b := range bits {
			if u&(1<<uint(b)) == 0 {
				continue
			}
			mint := extrab.Extra_bddBitsToCube(ddNew, c, int(k), ddNew.Vars, true)
			core.Cudd_Ref(mint)
			oldBit := bits[b]
			bits[b] = core.Cudd_bddOr(ddNew, oldBit, mint)
			core.Cudd_Ref(bits[b])
			core.Cudd_RecursiveDeref(ddNew, oldBit)
			core.Cudd_RecursiveDeref(ddNew, mint)
		}
	}
	bitNodes := make([]*abc.Object, nBits)
	for b := range bitNodes {
		bitNodes[b] = abc.Abc_NtkCreateNode(dst)
		for _, edge := range bound {
			abc.Abc_ObjAddFanin(bitNodes[b], edge.Object.Copy)
		}
		bitNodes[b].Data = bits[b]
	}
	composition := abc.Abc_NtkCreateNode(dst)
	for _, edge := range old.Fanins()[k:] {
		abc.Abc_ObjAddFanin(composition, edge.Object.Copy)
	}
	for _, node := range bitNodes {
		abc.Abc_ObjAddFanin(composition, node)
	}
	funcNode := ddNew.LogicZero
	core.Cudd_Ref(funcNode)
	codeVars := ddNew.Vars[len(old.Fanins())-int(k):]
	for u, item := range unique {
		moved := extrab.Extra_bddMove(ddOld, item, -int(k))
		if moved == nil {
			return nil
		}
		core.Cudd_Ref(moved)
		translated := extrab.Extra_TransferLevelByLevel(ddOld, ddNew, moved)
		if translated == nil {
			return nil
		}
		core.Cudd_Ref(translated)
		core.Cudd_RecursiveDeref(ddOld, moved)
		mint := extrab.Extra_bddBitsToCube(ddNew, u, int(nBits), codeVars, false)
		core.Cudd_Ref(mint)
		term := core.Cudd_bddAnd(ddNew, mint, translated)
		core.Cudd_Ref(term)
		core.Cudd_RecursiveDeref(ddNew, mint)
		core.Cudd_RecursiveDeref(ddNew, translated)
		oldFunc := funcNode
		funcNode = core.Cudd_bddOr(ddNew, oldFunc, term)
		core.Cudd_Ref(funcNode)
		core.Cudd_RecursiveDeref(ddNew, oldFunc)
		core.Cudd_RecursiveDeref(ddNew, term)
	}
	composition.Data = funcNode
	return composition
}

func utilLog2(n uint32) int { // ABC's Base2Log: ceil(log2(n)), with 0/1 preserved.
	if n < 2 {
		return int(n)
	}
	return bits.Len32(n - 1)
}

func lutUniqueCofactors(cofs []*core.DdNode) []*core.DdNode {
	seen := make(map[*core.DdNode]struct{}, len(cofs))
	out := make([]*core.DdNode, 0, len(cofs))
	for _, cof := range cofs {
		if _, ok := seen[cof]; !ok {
			seen[cof] = struct{}{}
			out = append(out, cof)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return uintptr(unsafe.Pointer(out[i])) < uintptr(unsafe.Pointer(out[j]))
	})
	return out
}

// Abc_NtkBddFindCofactor tries the source's two-LUT decomposition for K+1 fanins.
func Abc_NtkBddFindCofactor(dst *Abc_Ntk_t, node *Abc_Obj_t, lutSize int32) *Abc_Obj_t {
	if dst == nil || node == nil || int32(len(node.Fanins())) != lutSize+1 {
		return nil
	}
	oldDD, newDD := lutManager(node.Network()), lutManager(dst)
	if oldDD == nil || newDD == nil {
		return nil
	}
	n := len(node.Fanins())
	f, ok := node.Data.(*core.DdNode)
	if !ok || f == nil {
		return nil
	}
	iCof, smaller := -1, -1
	var cof0, cof1 *core.DdNode
	for i := 0; i < n; i++ {
		v := core.Cudd_bddIthVar(oldDD, i)
		cof0 = core.Cudd_Cofactor(oldDD, f, core.Cudd_Not(v))
		core.Cudd_Ref(cof0)
		cof1 = core.Cudd_Cofactor(oldDD, f, v)
		core.Cudd_Ref(cof1)
		if core.Cudd_SupportSize(oldDD, cof0) <= int(lutSize-2) {
			iCof, smaller = i, 0
			break
		}
		if core.Cudd_SupportSize(oldDD, cof1) <= int(lutSize-2) {
			iCof, smaller = i, 1
			break
		}
		core.Cudd_RecursiveDeref(oldDD, cof0)
		core.Cudd_RecursiveDeref(oldDD, cof1)
	}
	if iCof < 0 {
		return nil
	}
	small := cof0
	if smaller == 1 {
		small = cof1
	}
	support := make(map[int]struct{})
	for _, index := range oldDD.Support(small) {
		support[index] = struct{}{}
	}
	free := -1
	for i := 0; i < n; i++ {
		if i == iCof {
			continue
		}
		if _, found := support[i]; !found {
			free = i
			break
		}
	}
	if free < 0 {
		return nil
	}
	cof0n := extrab.Extra_TransferLevelByLevel(oldDD, newDD, cof0)
	core.Cudd_Ref(cof0n)
	cof1n := extrab.Extra_TransferLevelByLevel(oldDD, newDD, cof1)
	core.Cudd_Ref(cof1n)
	core.Cudd_RecursiveDeref(oldDD, cof0)
	core.Cudd_RecursiveDeref(oldDD, cof1)
	fanins := node.Fanins()
	bot := abc.Abc_NtkCreateNode(dst)
	for _, edge := range fanins {
		abc.Abc_ObjAddFanin(bot, edge.Object.Copy)
	}
	if smaller == 1 {
		bot.Data = cof0n
	} else {
		bot.Data = cof1n
	}
	top := abc.Abc_NtkCreateNode(dst)
	for i, edge := range fanins {
		if i == free {
			abc.Abc_ObjAddFanin(top, bot)
		} else {
			abc.Abc_ObjAddFanin(top, edge.Object.Copy)
		}
	}
	selector := core.Cudd_bddIthVar(newDD, iCof)
	freeVar := core.Cudd_bddIthVar(newDD, free)
	if smaller == 1 {
		top.Data = core.Cudd_bddIte(newDD, selector, cof1n, freeVar)
		core.Cudd_RecursiveDeref(newDD, cof0n)
	} else {
		top.Data = core.Cudd_bddIte(newDD, selector, freeVar, cof0n)
		core.Cudd_RecursiveDeref(newDD, cof1n)
	}
	core.Cudd_Ref(top.Data.(*core.DdNode))
	return top
}

func Abc_NtkBddDecompose(dst *Abc_Ntk_t, node *Abc_Obj_t, lutSize, verbose int32) *Abc_Obj_t {
	if node == nil || dst == nil || int32(len(node.Fanins())) <= lutSize {
		return nil
	}
	if int32(len(node.Fanins())) == lutSize+1 {
		if result := Abc_NtkBddFindCofactor(dst, node, lutSize); result != nil {
			if verbose != 0 {
				fmt.Printf("Decomposing %d-input node %d using MUX.\n", len(node.Fanins()), node.ID)
			}
			return result
		}
	}
	dd := lutManager(node.Network())
	f, _ := node.Data.(*core.DdNode)
	cofs := Abc_NtkBddCofactors(dd, f, lutSize)
	uniq := lutUniqueCofactors(cofs)
	if len(uniq) > 1<<uint(lutSize-2) {
		cofs = Abc_NtkBddCofactors(dd, f, 2)
		if verbose != 0 {
			fmt.Printf("Decomposing %d-input node %d using cofactoring with %d cofactors.\n", len(node.Fanins()), node.ID, len(cofs))
		}
		fanins := make([]*abc.Object, 6)
		fanins[0], fanins[1] = node.Fanins()[0].Object.Copy, node.Fanins()[1].Object.Copy
		for i, cof := range cofs {
			fanins[i+2] = Abc_NtkCreateCofLut(dst, dd, cof, node, 2)
		}
		switch lutSize {
		case 4:
			return Abc_NtkBddMux412(dst, fanins)
		case 5:
			return Abc_NtkBddMux412a(dst, fanins)
		case 6:
			return Abc_NtkBddMux411(dst, fanins)
		}
		return nil
	}
	if verbose != 0 {
		fmt.Printf("Decomposing %d-input node %d using Curtis with %d unique columns.\n", len(node.Fanins()), node.ID, len(uniq))
	}
	return Abc_NtkBddCurtis(dst, node, cofs, uniq)
}

func Abc_NtkLutminConstruct(collapsed, decomposed *Abc_Ntk_t, lutSize, verbose int32) {
	if collapsed == nil || decomposed == nil {
		return
	}
	for _, node := range abc.Abc_NtkDfs(collapsed, 0) {
		if int32(len(node.Fanins())) <= lutSize {
			node.Copy = abc.Abc_NtkDupObj(decomposed, node, 0)
			for _, edge := range node.Fanins() {
				abc.Abc_ObjAddFanin(node.Copy, edge.Object.Copy)
			}
		} else {
			node.Copy = Abc_NtkBddDecompose(decomposed, node, lutSize, verbose)
		}
	}
}

func Abc_NtkLutminInt(n *Abc_Ntk_t, lutSize, verbose int32) *Abc_Ntk_t {
	if n == nil {
		return nil
	}
	Abc_NtkBddReorder(n, 0)
	out := abc.Abc_NtkStartFrom(n, abc.ABC_NTK_LOGIC, abc.ABC_FUNC_BDD)
	if out == nil {
		return nil
	}
	dd := lutManager(out)
	if dd == nil {
		return nil
	}
	core.Cudd_bddIthVar(dd, int(abc.Abc_NtkGetFaninMax(n)))
	Abc_NtkLutminConstruct(n, out, lutSize, verbose)
	abc.Abc_NtkFinalize(n, out)
	abc.Abc_NtkMinimumBase(out)
	return out
}

func Abc_NtkLutmin(initial *Abc_Ntk_t, lutSize, verbose int32) *Abc_Ntk_t {
	if lutSize < 4 {
		fmt.Printf("The LUT count (%d) should be at least 4.\n", lutSize)
		return nil
	}
	if lutSize > 6 {
		fmt.Printf("The LUT count (%d) should not exceed 6.\n", lutSize)
		return nil
	}
	if initial == nil {
		return nil
	}
	var work *abc.Network
	if abc.Abc_NtkIsStrash(initial) != 0 {
		work = abc.Abc_NtkDup(initial)
	} else {
		work = Abc_NtkStrash(initial, 0, 1, 0)
	}
	if work == nil {
		return nil
	}
	collapsed := Abc_NtkCollapse(work, 10000, 0, 1, 0, 0, 0)
	abc.Abc_NtkDelete(work)
	if collapsed == nil {
		return nil
	}
	if abc.Abc_NtkIsBddLogic(collapsed) == 0 {
		abc.Abc_NtkToBdd(collapsed)
	}
	for i := int32(0); abc.Abc_NtkGetFaninMax(collapsed) > lutSize; i++ {
		if verbose != 0 {
			fmt.Printf("*** Iteration %d:\nDecomposing network with %d nodes and %d max fanin count for K = %d.\n", i+1, abc.Abc_NtkNodeNum(collapsed), abc.Abc_NtkGetFaninMax(collapsed), lutSize)
		}
		next := Abc_NtkLutminInt(collapsed, lutSize, verbose)
		abc.Abc_NtkDelete(collapsed)
		if next == nil {
			return nil
		}
		collapsed = next
	}
	abc.Abc_NtkLogicMakeSimpleCos(collapsed, 0)
	Abc_NtkFraigSweep(collapsed, 1, 0, 0, 0)
	if abc.Abc_NtkCheck(collapsed) == 0 {
		fmt.Println("Abc_NtkLutmin: The network check has failed.")
		return nil
	}
	return collapsed
}
