package abci

import (
	"fmt"
	"os"
	"sort"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/bdd/cudd/core"
	"github.com/glycerine/abcgo/bdd/extrab"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bmc"
	"github.com/glycerine/abcgo/sat/bsat"
)

// Abc_NodeMinimumBase2 compacts a BDD node to its actual support.
func Abc_NodeMinimumBase2(node *Abc_Obj_t) int32 { return abc.Abc_NodeMinimumBase(node) }
func Abc_NtkMinimumBase2(n *Abc_Ntk_t) int32     { return abc.Abc_NtkMinimumBase(n) }

// Abc_NodeFromGlobalBdds transfers one global BDD to a local BDD manager and
// attaches the fanins in the current variable order.
func Abc_NodeFromGlobalBdds(dst *Abc_Ntk_t, src *core.DdManager, f *core.DdNode, reverse int32) *Abc_Obj_t {
	if dst == nil || src == nil || f == nil {
		return nil
	}
	manager, ok := dst.FunctionManager.(*core.DdManager)
	if !ok {
		return nil
	}
	node := abc.Abc_NtkCreateNode(dst)
	count := int32(abc.Abc_NtkCiNum(dst))
	for _, index := range src.InvPerm {
		ci := int32(index)
		if reverse != 0 {
			ci = count - 1 - ci
		}
		if ci < 0 || ci >= count {
			return nil
		}
		abc.Abc_ObjAddFanin(node, abc.Abc_NtkCi(dst, ci))
	}
	local := extrab.Extra_TransferLevelByLevel(src, manager, f)
	if local == nil {
		return nil
	}
	core.Cudd_Ref(local)
	node.Data = local
	return node
}

// Abc_NtkFromGlobalBdds builds the BDD logic network from output functions.
func Abc_NtkFromGlobalBdds(n *Abc_Ntk_t, reverse int32) *Abc_Ntk_t {
	if n == nil {
		return nil
	}
	dd := Abc_NtkGlobalBddMan(n)
	if dd == nil {
		return nil
	}
	if n.Exdc != nil {
		if abc.Abc_NtkIsStrash(n.Exdc) == 0 || abc.Abc_NtkCoNum(n.Exdc) != 1 {
			return nil
		}
		dcManager := Abc_NtkBuildGlobalBdds(n.Exdc, 10_000_000, 1, 1, 0, 0)
		if dcManager == nil {
			return nil
		}
		dc := Abc_ObjGlobalBdd(abc.Abc_NtkCo(n.Exdc, 0))
		mapped := core.Cudd_bddTransfer(dcManager, dd, dc)
		if mapped == nil {
			Abc_NtkFreeGlobalBdds(n.Exdc, 1)
			return nil
		}
		core.Cudd_Ref(mapped)
		Abc_NtkFreeGlobalBdds(n.Exdc, 1)
		for _, co := range n.ObjectsOfKind(abc.ObjectPO) {
			min := Abc_ObjGlobalBdd(co)
			if min == nil {
				continue
			}
			lo := core.Cudd_bddAnd(dd, min, core.Cudd_Not(mapped))
			if lo == nil {
				return nil
			}
			core.Cudd_Ref(lo)
			u := core.Cudd_bddAnd(dd, core.Cudd_Not(min), core.Cudd_Not(mapped))
			if u == nil {
				return nil
			}
			core.Cudd_Ref(u)
			core.Cudd_RecursiveDeref(dd, min)
			min = core.Cudd_bddIsop(dd, lo, core.Cudd_Not(u))
			if min == nil {
				return nil
			}
			core.Cudd_Ref(min)
			core.Cudd_RecursiveDeref(dd, lo)
			core.Cudd_RecursiveDeref(dd, u)
			Abc_ObjSetGlobalBdd(co, min)
		}
		core.Cudd_RecursiveDeref(dd, mapped)
	}
	out := abc.Abc_NtkStartFrom(n, abc.ABC_NTK_LOGIC, abc.ABC_FUNC_BDD)
	if out == nil {
		return nil
	}
	newDD, ok := out.FunctionManager.(*core.DdManager)
	if !ok {
		abc.Abc_NtkDelete(out)
		return nil
	}
	if len(dd.Vars) > 0 {
		core.Cudd_bddIthVar(newDD, len(dd.Vars)-1)
	}
	for i, co := range n.ObjectsOfKind(abc.ObjectPO) {
		if i >= int(abc.Abc_NtkCoNum(n)) {
			break
		}
		driver := abc.Abc_ObjFanin0(co)
		if driver.IsCI() && abc.Abc_ObjName(co) == abc.Abc_ObjName(driver) {
			abc.Abc_ObjAddFanin(co.Copy, abc.Edge{Object: driver.Copy, Complement: co.Fanins()[0].Complement})
			continue
		}
		f := Abc_ObjGlobalBdd(co)
		node := Abc_NodeFromGlobalBdds(out, dd, f, reverse)
		if node == nil {
			abc.Abc_NtkDelete(out)
			return nil
		}
		abc.Abc_ObjAddFanin(co.Copy, node)
	}
	return out
}

func Abc_NtkDumpVariableOrder(n *Abc_Ntk_t) {
	dd := Abc_NtkGlobalBddMan(n)
	if dd == nil {
		return
	}
	f, err := os.Create("order.txt")
	if err != nil {
		return
	}
	defer f.Close()
	for _, idx := range dd.InvPerm {
		fmt.Fprintf(f, "%d ", idx)
	}
	fmt.Fprintln(f)
}

func Abc_NtkCollapse(n *Abc_Ntk_t, bddSizeMax, dualRail, reorder, reverse, dumpOrder, verbose int32) *Abc_Ntk_t {
	if n == nil || abc.Abc_NtkIsStrash(n) == 0 {
		return nil
	}
	start := Abc_Clock()
	if Abc_NtkBuildGlobalBdds(n, bddSizeMax, 1, reorder, reverse, verbose) == nil {
		return nil
	}
	if verbose != 0 {
		fmt.Printf("Shared BDD size = %6d nodes.  ", Abc_NtkSizeOfGlobalBdds(n))
		Abc_PrintTime(1, "BDD construction time", Abc_Clock()-start)
	}
	if dumpOrder != 0 {
		Abc_NtkDumpVariableOrder(n)
	}
	out := Abc_NtkFromGlobalBdds(n, reverse)
	Abc_NtkFreeGlobalBdds(n, 1)
	if out == nil {
		return nil
	}
	Abc_NtkMinimumBase2(out)
	if n.Exdc != nil {
		out.Exdc = abc.Abc_NtkDup(n.Exdc)
	}
	if abc.Abc_NtkCheck(out) == 0 {
		fmt.Println("Abc_NtkCollapse: The network check has failed.")
		abc.Abc_NtkDelete(out)
		return nil
	}
	return out
}

// Abc_NtkCollapseReduce removes literals unused by any cube and updates each
// output's support list in the matching order.
func Abc_NtkCollapseReduce(sop *vec.Vec_Str_t, support, class *vec.Vec_Int_t, supports *vec.Vec_Wec_t) int32 {
	n := int(vec.Vec_IntSize(support))
	text := vec.Vec_StrArray(sop)
	if len(text) == 4 {
		for i := 0; i < vec.Vec_IntSize(class); i++ {
			vec.Vec_IntClear(vec.Vec_WecEntry(supports, int(vec.Vec_IntEntry(class, i))))
		}
		return 1
	}
	present := make([]bool, n)
	stride := n + 3
	for off := 0; off+stride-1 < len(text) && text[off] != '\x00'; off += stride {
		for k := 0; k < n; k++ {
			if text[off+k] != '-' {
				present[k] = true
			}
		}
	}
	kept := 0
	for _, v := range present {
		if v {
			kept++
		}
	}
	if kept == n {
		return 0
	}
	out := make([]byte, 0, len(text))
	for off := 0; off+stride <= len(text) && text[off] != '\x00'; off += stride {
		for k := 0; k < stride; k++ {
			if k >= n || present[k] {
				out = append(out, text[off+k])
			}
		}
	}
	out = append(out, 0)
	vec.Vec_StrClear(sop)
	for _, b := range out {
		vec.Vec_StrPush(sop, b)
	}
	for i := 0; i < vec.Vec_IntSize(class); i++ {
		co := int(vec.Vec_IntEntry(class, i))
		v := vec.Vec_WecEntry(supports, co)
		write := 0
		for k := 0; k < vec.Vec_IntSize(v); k++ {
			if present[k] {
				vec.Vec_IntWriteEntry(v, write, vec.Vec_IntEntry(v, k))
				write++
			}
		}
		vec.Vec_IntShrink(v, write)
	}
	return 1
}

func Abc_NtkCollapseSat(n *Abc_Ntk_t, cubeLimit, conflictLimit, costMax, canonical, reverse, cnfShared, verbose int32) *Abc_Ntk_t {
	if n == nil || abc.Abc_NtkIsStrash(n) == 0 {
		return nil
	}
	g := Abc_NtkClpGia(n)
	if g == nil {
		return nil
	}
	supports := gia.Gia_ManCreateCoSupps(g, verbose)
	if supports == nil {
		gia.Gia_ManStop(g)
		return nil
	}
	if cubeLimit > 0 && costMax > 0 {
		largest := gia.Gia_ManCoLargestSupp(g, supports)
		if largest < 0 {
			gia.Gia_ManStop(g)
			vec.Vec_WecFree(supports)
			return nil
		}
		root := gia.Gia_ManCo(g, largest)
		coneSize := gia.Gia_ManConeSize(g, []int32{gia.Gia_ObjId(g, root)}, 1)
		suppSize := vec.Vec_IntSize(vec.Vec_WecEntry(supports, int(largest)))
		if int64(coneSize)*int64(suppSize)*int64(cubeLimit) > int64(costMax) {
			gia.Gia_ManStop(g)
			vec.Vec_WecFree(supports)
			return nil
		}
	}
	out := abc.Abc_NtkStartFrom(n, abc.ABC_NTK_LOGIC, abc.ABC_FUNC_SOP)
	if out == nil {
		gia.Gia_ManStop(g)
		vec.Vec_WecFree(supports)
		return nil
	}
	sops := Abc_GiaDeriveSops(out, g, supports, cubeLimit, conflictLimit, costMax, canonical, reverse, cnfShared, verbose)
	gia.Gia_ManStop(g)
	if sops == nil {
		abc.Abc_NtkDelete(out)
		vec.Vec_WecFree(supports)
		return nil
	}
	for i, co := range n.ObjectsOfKind(abc.ObjectPO) {
		if i >= len(sops) {
			break
		}
		edge := co.Fanins()[0]
		driver := edge.Object
		if driver.IsCI() && abc.Abc_ObjName(co) == abc.Abc_ObjName(driver) {
			abc.Abc_ObjAddFanin(co.Copy, abc.Edge{Object: driver.Copy, Complement: edge.Complement})
			continue
		}
		node := abc.Abc_NtkCreateNode(out)
		if driver.IsCI() {
			abc.Abc_ObjAddFanin(node, driver.Copy)
			if edge.Complement {
				node.Data = "0 1\n"
			} else {
				node.Data = "1 1\n"
			}
		} else if driver.Kind == abc.ObjectConst1 {
			if edge.Complement {
				node.Data = " 0\n"
			} else {
				node.Data = " 1\n"
			}
		} else {
			supp := vec.Vec_WecEntry(supports, i)
			for k := 0; k < vec.Vec_IntSize(supp); k++ {
				abc.Abc_ObjAddFanin(node, abc.Abc_NtkCi(out, vec.Vec_IntEntry(supp, k)))
			}
			node.Data = abc.Abc_SopRegister(nil, sops[i])
		}
		abc.Abc_ObjAddFanin(co.Copy, node)
	}
	vec.Vec_WecFree(supports)
	abc.Abc_NtkSortSops(out)
	if n.Exdc != nil {
		out.Exdc = abc.Abc_NtkDup(n.Exdc)
	}
	if abc.Abc_NtkCheck(out) == 0 {
		abc.Abc_NtkDelete(out)
		return nil
	}
	return out
}

// collapseDeriveSops ports the representative-output and equivalence-class
// workflow in Abc_GiaDeriveSops. The optional shared-CNF path maps only the
// representative's cone and support into the SAT solver, as in C.
func Abc_GiaDeriveSops(dst *Abc_Ntk_t, g *gia.Gia_Man_t, supports *vec.Vec_Wec_t, cubeLimit, conflictLimit, costMax, canonical, reverse, cnfShared, verbose int32) []string {
	classes := gia.Gia_ManIsoStrashReduceInt(g, supports, 0)
	defer vec.Vec_WecFree(classes)
	reprs := vec.Vec_WecCollectFirsts(classes)
	defer vec.Vec_IntFree(reprs)
	type ranked struct {
		index   int
		support int
	}
	rank := make([]ranked, vec.Vec_IntSize(reprs))
	for i, output := range vec.Vec_IntArray(reprs) {
		rank[i] = ranked{i, vec.Vec_IntSize(vec.Vec_WecEntry(supports, int(output)))}
	}
	sort.SliceStable(rank, func(i, j int) bool { return rank[i].support < rank[j].support })
	var cnf *gia.Cnf_Dat_t
	if cnfShared != 0 && canonical == 0 {
		cnf = gia.Mf_ManGenerateCnf(g, 8, 1, 0, 0, 0)
		if cnf == nil {
			return nil
		}
	}
	reprSops := make([]string, vec.Vec_IntSize(reprs))
	for order := len(rank) - 1; order >= 0; order-- {
		entry := rank[order].index
		coIndex := vec.Vec_IntEntry(reprs, entry)
		supp := vec.Vec_WecEntry(supports, int(coIndex))
		if vec.Vec_IntSize(supp) < 2 {
			reprSops[entry] = "\x01"
			continue
		}
		var sop *vec.Vec_Str_t
		currentVerbose := int32(0)
		if order == len(rank)-1 {
			currentVerbose = verbose
		}
		if cnf != nil {
			sop = Abc_NtkClpGiaOne2(cnf, g, int(coIndex), supp, vec.Vec_WecEntry(classes, entry), supports, cubeLimit, conflictLimit, canonical, reverse, currentVerbose)
		} else {
			sop = Abc_NtkClpGiaOne(g, int(coIndex), cubeLimit, conflictLimit, canonical, reverse, supp, currentVerbose, vec.Vec_WecEntry(classes, entry), supports)
		}
		if sop == nil {
			return nil
		}
		reprSops[entry] = string(vec.Vec_StrArray(sop))
		vec.Vec_StrFree(sop)
	}
	result := make([]string, gia.Gia_ManCoNum(g))
	for class := 0; class < vec.Vec_WecSize(classes); class++ {
		for _, output := range vec.Vec_IntArray(vec.Vec_WecEntry(classes, class)) {
			result[output] = reprSops[class]
		}
	}
	_ = dst
	_ = costMax
	return result
}

func Abc_NtkClpGiaOne(g *gia.Gia_Man_t, coIndex int, cubeLimit, conflictLimit, canonical, reverse int32, support, class *vec.Vec_Int_t, verbose int32, supports *vec.Vec_Wec_t) *vec.Vec_Str_t {
	cone := gia.Gia_ManDupCones(g, []int32{int32(coIndex)}, 1, 1)
	if cone == nil {
		return nil
	}
	sop := bmc.Bmc_CollapseOneOld(cone, cubeLimit, conflictLimit, canonical, reverse, verbose)
	gia.Gia_ManStop(cone)
	if sop != nil {
		Abc_NtkCollapseReduce(sop, support, class, supports)
	}
	return sop
}

func Abc_NtkClpGiaOne2(cnf *gia.Cnf_Dat_t, g *gia.Gia_Man_t, coIndex int, support, class *vec.Vec_Int_t, supports *vec.Vec_Wec_t, cubeLimit, conflictLimit, canonical, reverse, verbose int32) *vec.Vec_Str_t {
	coObj := gia.Gia_ManCo(g, int32(coIndex))
	coID := gia.Gia_ObjId(g, coObj)
	ands := vec.Vec_IntAlloc(100)
	gia.Gia_ManIncrementTravId(g)
	gia.Gia_ManCollectAnds(g, []int32{coID}, 1, ands, nil)
	if int(coID) >= len(cnf.VarNums) || int(coID) >= len(cnf.Obj2Clause) {
		vec.Vec_IntFree(ands)
		return nil
	}
	supportObjects := make([]int32, vec.Vec_IntSize(support))
	for i, ci := range vec.Vec_IntArray(support) {
		supportObjects[i] = gia.Gia_ObjId(g, gia.Gia_ManCi(g, ci))
	}
	mapping := make([]int32, len(cnf.VarNums))
	for i := range mapping {
		mapping[i] = -1
	}
	solver := Abc_NtkClpDeriveSatSolver(cnf, coID, supportObjects, vec.Vec_IntArray(ands), &mapping)
	vec.Vec_IntFree(ands)
	if solver == nil {
		return nil
	}
	defer solver.Close()
	sop := bmc.Bmc_CollapseOne_int(solver, int32(vec.Vec_IntSize(support)), cubeLimit, conflictLimit, canonical, reverse, verbose)
	if sop == nil {
		return nil
	}
	Abc_NtkCollapseReduce(sop, support, class, supports)
	return sop
}

func Abc_NtkClpDeriveSatSolver(cnf *gia.Cnf_Dat_t, coID int32, supportObjects, ands []int32, mapping *[]int32) *bsat.Solver {
	if cnf == nil || mapping == nil || coID < 0 || int(coID) >= len(cnf.Obj2Clause) || int(coID) >= len(cnf.Obj2Count) {
		return nil
	}
	ids := *mapping
	if len(ids) < len(cnf.VarNums) || int(coID) >= len(ids) {
		return nil
	}
	nVars := int32(2)
	ids[coID] = nVars
	nVars++
	for _, objID := range supportObjects {
		if objID < 0 || int(objID) >= len(ids) {
			return nil
		}
		ids[objID] = nVars
		nVars++
	}
	for _, objID := range ands {
		if objID >= 0 && int(objID) < len(cnf.Obj2Clause) && cnf.Obj2Clause[objID] != -1 {
			ids[objID] = nVars
			nVars++
		}
	}
	solver := bsat.NewSolver()
	solver.SetNumVars(nVars)
	work := append(append([]int32(nil), ands...), coID)
	for _, objID := range work {
		if objID < 0 || int(objID) >= len(cnf.Obj2Clause) || cnf.Obj2Clause[objID] == -1 {
			continue
		}
		begin, count := cnf.Obj2Clause[objID], cnf.Obj2Count[objID]
		for ci := begin; ci < begin+count; ci++ {
			clause := cnf.Clauses[ci]
			lits := make([]bsat.Literal, len(clause))
			for i, lit := range clause {
				id := lit >> 1
				if id < 0 || int(id) >= len(ids) || ids[id] < 0 {
					solver.Close()
					return nil
				}
				if lit&1 != 0 {
					lits[i] = bsat.Negative(ids[id])
				} else {
					lits[i] = bsat.Positive(ids[id])
				}
			}
			if !solver.AddClause(lits...) {
				solver.Close()
				return nil
			}
		}
	}
	return solver
}

// Kept for C API compatibility; BMC derivation is centralized in bmc.
func Abc_NtkFromSopsInt(n *Abc_Ntk_t, cubeLimit, conflictLimit, costMax, canonical, reverse, cnfShared, verbose int32) *Abc_Ntk_t {
	return Abc_NtkCollapseSat(n, cubeLimit, conflictLimit, costMax, canonical, reverse, cnfShared, verbose)
}
