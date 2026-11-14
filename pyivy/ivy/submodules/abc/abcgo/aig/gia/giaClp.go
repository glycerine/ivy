package gia

// BDD collapsing and disjoint-support rebuilding, adapted from giaClp.c.
import (
	"fmt"
	"os"

	cudd "github.com/glycerine/abcgo/bdd/cudd/core"
	"github.com/glycerine/abcgo/bdd/dsd"
	"github.com/glycerine/abcgo/bdd/extrab"
	"github.com/glycerine/abcgo/misc/vec"
)

func Gia_GetFakeNames(n int32) *Vec_Ptr_t {
	names := vec.Vec_PtrAlloc(int(n))
	for i := int32(0); i < n; i++ {
		name := string(rune('a' + i%26))
		if n >= 26 {
			name += fmt.Sprint(i / 26)
		}
		vec.Vec_PtrPush(names, name)
	}
	return names
}

// Gia_ManRebuildIsop builds the local BDD using Shannon decomposition. Unlike
// the original SOP factoring, shared BDD nodes become shared AIG subgraphs.
func Gia_ManRebuildIsop(dd *cudd.DdManager, local *cudd.DdNode, p *Gia_Man_t, fanins *Vec_Int_t, sop, cube *Vec_Str_t) int32 {
	if dd == nil || local == nil {
		return -1
	}
	literals := vec.Vec_IntArray(fanins)
	memo := make(map[*cudd.DdNode]int32)
	var build func(*cudd.DdNode) int32
	build = func(node *cudd.DdNode) int32 {
		if node == dd.One {
			return 1
		}
		if node == dd.LogicZero {
			return 0
		}
		if lit, ok := memo[node]; ok {
			return lit
		}
		regular := cudd.Cudd_Regular(node)
		level := dd.Perm[regular.Index]
		if level < 0 || level >= len(literals) {
			return -1
		}
		inv := cudd.Cudd_IsComplement(node)
		lo := build(cudd.Cudd_NotCond(regular.E, inv))
		hi := build(cudd.Cudd_NotCond(regular.T, inv))
		if lo < 0 || hi < 0 {
			return -1
		}
		lit := Gia_ManHashMux(p, literals[level], hi, lo)
		memo[node] = lit
		return lit
	}
	return build(local)
}

func Gia_ManRebuildNode(manager *dsd.Dsd_Manager_t, node *dsd.Dsd_Node_t, p *Gia_Man_t, dd *cudd.DdManager, fanins *Vec_Int_t, sop, cube *Vec_Str_t) int32 {
	vec.Vec_IntClear(fanins)
	for i := 0; i < dsd.Dsd_NodeReadDecsNum(node); i++ {
		child := dsd.Dsd_NodeReadDec(node, i)
		lit := int32(dsd.Dsd_NodeReadMark(child))
		if dsd.Dsd_IsComplement(child) {
			lit ^= 1
		}
		vec.Vec_IntPush(fanins, lit)
	}
	var lit int32
	switch dsd.Dsd_NodeReadType(node) {
	case dsd.DSD_NODE_CONST1:
		lit = 1
	case dsd.DSD_NODE_OR:
		for _, v := range vec.Vec_IntArray(fanins) {
			lit = Gia_ManHashOr(p, lit, v)
		}
	case dsd.DSD_NODE_EXOR:
		for _, v := range vec.Vec_IntArray(fanins) {
			lit = Gia_ManHashXor(p, lit, v)
		}
	case dsd.DSD_NODE_PRIME:
		src := dsd.Dsd_ManagerReadDd(manager)
		local := dsd.Dsd_TreeGetPrimeFunction(src, node)
		local = extrab.Extra_TransferLevelByLevel(src, dd, local)
		lit = Gia_ManRebuildIsop(dd, local, p, fanins, sop, cube)
	default:
		return -1
	}
	dsd.Dsd_NodeSetMark(node, int(lit))
	return lit
}

func Gia_ManRebuild(p *Gia_Man_t, manager *dsd.Dsd_Manager_t, dd *cudd.DdManager) *Gia_Man_t {
	out := Gia_ManStart(2 * Gia_ManObjNum(p))
	out.pName, out.PSpec = p.pName, p.PSpec
	Gia_ManHashAlloc(out)
	dsd.Dsd_NodeSetMark(dsd.Dsd_ManagerReadConst1(manager), 1)
	for i := int32(0); i < Gia_ManCiNum(p); i++ {
		dsd.Dsd_NodeSetMark(dsd.Dsd_ManagerReadInput(manager, int(i)), int(Gia_ManAppendCi(out)))
	}
	fanins := vec.Vec_IntAlloc(16)
	var count int
	for _, node := range dsd.Dsd_TreeCollectNodesDfs(manager, &count) {
		if Gia_ManRebuildNode(manager, node, out, dd, fanins, nil, nil) < 0 {
			Gia_ManStop(out)
			return Gia_ManDup(p)
		}
	}
	for i := int32(0); i < Gia_ManCoNum(p); i++ {
		root := dsd.Dsd_ManagerReadRoot(manager, int(i))
		lit := int32(dsd.Dsd_NodeReadMark(root))
		if dsd.Dsd_IsComplement(root) {
			lit ^= 1
		}
		Gia_ManAppendCo(out, lit)
	}
	Gia_ManSetRegNum(out, Gia_ManRegNum(p))
	return out
}

func Gia_ManCollapseDeref(dd *cudd.DdManager, funcs *Vec_Ptr_t) {
	if funcs == nil {
		return
	}
	for _, v := range vec.Vec_PtrArray(funcs) {
		if node, ok := v.(*cudd.DdNode); ok && node != nil {
			cudd.Cudd_RecursiveDeref(dd, node)
		}
	}
	vec.Vec_PtrFree(funcs)
}

func Gia_ObjCollapseDeref(p *Gia_Man_t, dd *cudd.DdManager, funcs *Vec_Ptr_t, id int32) {
	if Gia_ObjRefDecId(p, id) != 0 {
		return
	}
	if node, ok := vec.Vec_PtrEntry(funcs, int(id)).(*cudd.DdNode); ok {
		cudd.Cudd_RecursiveDeref(dd, node)
	}
	vec.Vec_PtrWriteEntry(funcs, int(id), nil)
}

// Gia_ManCollapse returns referenced BDD roots in combinational-output order.
func Gia_ManCollapse(p *Gia_Man_t, dd *cudd.DdManager, limit, verbose int32) *Vec_Ptr_t {
	if p == nil || dd == nil || limit < 0 {
		return nil
	}
	values := make([]*cudd.DdNode, p.nObjs)
	values[0] = dd.LogicZero
	for i := int32(0); i < Gia_ManCiNum(p); i++ {
		values[Gia_ObjId(p, Gia_ManCi(p, i))] = cudd.Cudd_bddIthVar(dd, int(i))
	}
	read := func(lit int32) *cudd.DdNode { return cudd.Cudd_NotCond(values[lit>>1], lit&1 != 0) }
	for id := int32(1); id < p.nObjs; id++ {
		a, b, c, kind := p.Gate(id)
		switch kind {
		case 1:
			continue
		case 2:
			values[id] = cudd.Cudd_bddAndLimit(dd, read(a), read(b), uint32(limit))
		case 3:
			values[id] = cudd.Cudd_bddXor(dd, read(a), read(b))
		case 4:
			values[id] = cudd.Cudd_bddIte(dd, read(a), read(b), read(c))
		case 5, 6:
			values[id] = read(a)
		default:
			return nil
		}
		if values[id] == nil {
			return nil
		}
	}
	result := vec.Vec_PtrAlloc(int(Gia_ManCoNum(p)))
	for i := int32(0); i < Gia_ManCoNum(p); i++ {
		node := values[Gia_ObjId(p, Gia_ManCo(p, i))]
		cudd.Cudd_Ref(node)
		vec.Vec_PtrPush(result, node)
	}
	return result
}

func Gia_ManCollapseTest(p *Gia_Man_t, verbose int32) *Gia_Man_t {
	dd := cudd.Cudd_Init(uint32(Gia_ManCiNum(p)), 0, 0, 0, 0)
	defer cudd.Cudd_Quit(dd)
	funcs := Gia_ManCollapse(p, dd, 10000, 0)
	if funcs == nil {
		return Gia_ManDup(p)
	}
	defer Gia_ManCollapseDeref(dd, funcs)
	roots := make([]*cudd.DdNode, vec.Vec_PtrSize(funcs))
	for i, v := range vec.Vec_PtrArray(funcs) {
		roots[i] = v.(*cudd.DdNode)
	}
	manager := dsd.Dsd_ManagerStart(dd, int(Gia_ManCiNum(p)), 0)
	defer dsd.Dsd_ManagerStop(manager)
	dsd.Dsd_Decompose(manager, roots, len(roots))
	if verbose != 0 {
		inputs := make([]string, Gia_ManCiNum(p))
		outputs := make([]string, Gia_ManCoNum(p))
		for i := range inputs {
			inputs[i] = fmt.Sprintf("i%d", i)
		}
		for i := range outputs {
			outputs[i] = fmt.Sprintf("o%d", i)
		}
		fmt.Printf("Ins = %d. Outs = %d. BDD nodes = %d.\n", len(inputs), len(outputs), cudd.Cudd_ReadNodeCount(dd))
		dsd.Dsd_TreePrint(os.Stdout, manager, inputs, outputs, 0, -1)
	}
	local := cudd.Cudd_Init(uint32(Gia_ManCiNum(p)), 0, 0, 0, 0)
	defer cudd.Cudd_Quit(local)
	return Gia_ManRebuild(p, manager, local)
}

func Gia_ManCollapseTestTest(p *Gia_Man_t) {
	out := Gia_ManCollapseTest(p, 0)
	fmt.Printf("GIA collapse: %d nodes -> %d nodes\n", Gia_ManObjNum(p), Gia_ManObjNum(out))
	Gia_ManStop(out)
}
