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

// Gia_ManRebuildIsop follows C's two-polarity ISOP selection and SOP factoring.
func Gia_ManRebuildIsop(dd *cudd.DdManager, local *cudd.DdNode, p *Gia_Man_t, fanins *Vec_Int_t, sop, cube *Vec_Str_t) int32 {
	if dd == nil || local == nil {
		return -1
	}
	// Constants are a Go extension; the source calls this only for prime nodes.
	if local == dd.LogicZero {
		return 0
	}
	if local == dd.One {
		return 1
	}
	var covers [2]*cudd.DdNode
	var counts [2]float64
	for phase := 0; phase < 2; phase++ {
		function := local
		if phase == 0 {
			function = cudd.Cudd_Not(local)
		}
		bdd := cudd.Cudd_zddIsop(dd, function, function, &covers[phase])
		if bdd == nil {
			if phase != 0 {
				cudd.Cudd_RecursiveDeref(dd, covers[0])
			}
			return -1
		}
		cudd.Cudd_Ref(covers[phase])
		cudd.Cudd_Ref(bdd)
		cudd.Cudd_RecursiveDeref(dd, bdd)
		counts[phase] = cudd.Cudd_zddCountDouble(dd, covers[phase])
	}
	phase := 1
	if counts[0] < counts[1] {
		phase = 0
	}
	cudd.Cudd_RecursiveDeref(dd, covers[phase^1])
	defer cudd.Cudd_RecursiveDeref(dd, covers[phase])
	// Counts at most 1000 are exact floats; larger counts are only rejected.
	if counts[phase] < 0 || counts[phase] > 1000 {
		return -1
	}
	n := Vec_IntSize(fanins)
	if sop == nil {
		sop = vec.Vec_StrAlloc(0)
	}
	if cube == nil {
		cube = vec.Vec_StrAlloc(0)
	}
	vec.Vec_StrFill(sop, (n+3)*int(counts[phase])+1, 0)
	vec.Vec_StrFill(cube, n, '-')
	vec.Vec_StrPush(cube, 0)
	offset := 0
	if !giaClpWriteCover(dd, covers[phase], vec.Vec_StrArray(sop), vec.Vec_StrArray(cube)[:n], byte('0'+phase), &offset) {
		return -1
	}
	return Gia_ManFactorNode(p, string(vec.Vec_StrArray(sop)[:offset]), fanins)
}

// Keep ABC's negative, positive, absent traversal order when serializing cubes.
// These small cover helpers live here because importing base/abc would cycle.
func giaClpWriteCover(dd *cudd.DdManager, cover *cudd.DdNode, sop, cube []byte, phase byte, offset *int) bool {
	if cover == dd.Zero {
		return true
	}
	if cover == dd.One {
		if *offset+len(cube)+3 >= len(sop) {
			return false
		}
		copy(sop[*offset:], cube)
		*offset += len(cube)
		copy(sop[*offset:], []byte{' ', phase, '\n'})
		*offset += 3
		return true
	}
	positive, negative, absent, ok := cudd.ZDDCoverCofactors(dd, cover, int(cover.Index&^1))
	index := int(cover.Index / 2)
	if !ok || index >= len(cube) {
		return false
	}
	cube[index] = '0'
	if !giaClpWriteCover(dd, negative, sop, cube, phase, offset) {
		return false
	}
	cube[index] = '1'
	if !giaClpWriteCover(dd, positive, sop, cube, phase, offset) {
		return false
	}
	cube[index] = '-'
	return giaClpWriteCover(dd, absent, sop, cube, phase, offset)
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
		cudd.Cudd_Ref(local)
		transferred := extrab.Extra_TransferLevelByLevel(src, dd, local)
		cudd.Cudd_Ref(transferred)
		cudd.Cudd_RecursiveDeref(src, local)
		lit = Gia_ManRebuildIsop(dd, transferred, p, fanins, sop, cube)
		cudd.Cudd_RecursiveDeref(dd, transferred)
	default:
		return -1
	}
	dsd.Dsd_NodeSetMark(node, int(lit))
	return lit
}

func Gia_ManRebuild(p *Gia_Man_t, manager *dsd.Dsd_Manager_t, dd *cudd.DdManager) *Gia_Man_t {
	var count int
	nodes := dsd.Dsd_TreeCollectNodesDfs(manager, &count)
	if count == 0 {
		return Gia_ManDup(p)
	}
	out := Gia_ManStart(2 * Gia_ManObjNum(p))
	out.pName, out.PSpec = p.pName, p.PSpec
	Gia_ManHashAlloc(out)
	dsd.Dsd_NodeSetMark(dsd.Dsd_ManagerReadConst1(manager), 1)
	for i := int32(0); i < Gia_ManCiNum(p); i++ {
		dsd.Dsd_NodeSetMark(dsd.Dsd_ManagerReadInput(manager, int(i)), int(Gia_ManAppendCi(out)))
	}
	fanins := vec.Vec_IntAlloc(16)
	sop, cube := vec.Vec_StrAlloc(0), vec.Vec_StrAlloc(0)
	for _, node := range nodes {
		if Gia_ManRebuildNode(manager, node, out, dd, fanins, sop, cube) < 0 {
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
	giaCopyInterfaceMetadata(out, p)
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
	// The source algorithm accepts ordinary ANDs. Lower typed gates first so
	// every constituent AND obeys the same allocation limit as the source.
	if p.nXors != 0 || p.nMuxes != 0 {
		plain := Gia_ManDupNoMuxes(p, 0)
		defer Gia_ManStop(plain)
		return Gia_ManCollapse(plain, dd, limit, verbose)
	}
	values := make([]*cudd.DdNode, p.nObjs)
	remaining := make([]int32, p.nObjs)
	for id := int32(1); id < p.nObjs; id++ {
		a, b, _, kind := p.Gate(id)
		switch kind {
		case 2:
			remaining[b>>1]++
			fallthrough
		case 5, 6:
			remaining[a>>1]++
		}
	}
	// Own live roots until their final fanout is processed, as in the source.
	// The local counts also support repeated calls without changing p.Refs.
	defer func() {
		for _, node := range values {
			cudd.Cudd_RecursiveDeref(dd, node)
		}
	}()
	retain := func(id int32, node *cudd.DdNode) {
		values[id] = node
		cudd.Cudd_Ref(node)
	}
	release := func(lit int32) {
		id := lit >> 1
		remaining[id]--
		if remaining[id] == 0 {
			cudd.Cudd_RecursiveDeref(dd, values[id])
			values[id] = nil
		}
	}
	if remaining[0] != 0 {
		retain(0, dd.LogicZero)
	}
	for i := int32(0); i < Gia_ManCiNum(p); i++ {
		id := Gia_ObjId(p, Gia_ManCi(p, i))
		if remaining[id] != 0 {
			retain(id, cudd.Cudd_bddIthVar(dd, int(i)))
		}
	}
	read := func(lit int32) *cudd.DdNode { return cudd.Cudd_NotCond(values[lit>>1], lit&1 != 0) }
	for id := int32(1); id < p.nObjs; id++ {
		a, b, _, kind := p.Gate(id)
		switch kind {
		case 1:
			continue
		case 2:
			values[id] = cudd.Cudd_bddAndLimit(dd, read(a), read(b), uint32(limit))
		case 5, 6:
			values[id] = read(a)
		default:
			return nil
		}
		if values[id] == nil {
			return nil
		}
		cudd.Cudd_Ref(values[id])
		release(a)
		if kind == 2 {
			release(b)
		}
		// Dangling gates are accepted by Go; release their unused results too.
		if kind != 6 && remaining[id] == 0 {
			cudd.Cudd_RecursiveDeref(dd, values[id])
			values[id] = nil
		}
	}
	result := vec.Vec_PtrAlloc(int(Gia_ManCoNum(p)))
	for i := int32(0); i < Gia_ManCoNum(p); i++ {
		id := Gia_ObjId(p, Gia_ManCo(p, i))
		vec.Vec_PtrPush(result, values[id])
		values[id] = nil // Transfer the output reference to the returned vector.
	}
	return result
}

func Gia_ManCollapseTest(p *Gia_Man_t, verbose int32) *Gia_Man_t {
	dd := cudd.Cudd_Init(uint32(Gia_ManCiNum(p)), 0, 0, 0, 0)
	defer cudd.Cudd_Quit(dd)
	cudd.Cudd_AutodynEnable(dd, cudd.CUDD_REORDER_SYMM_SIFT)
	funcs := Gia_ManCollapse(p, dd, 10000, 0)
	cudd.Cudd_AutodynDisable(dd)
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
	if verbose != 0 {
		fmt.Printf("Ins = %d. Outs = %d.  Shared BDD nodes = %d.  Peak live nodes = %d. Peak nodes = %d.\n",
			Gia_ManCiNum(p), Gia_ManCoNum(p), giaClpSharingSize(roots), cudd.Cudd_ReadPeakLiveNodeCount(dd), cudd.Cudd_ReadNodeCount(dd))
	}
	dsd.Dsd_Decompose(manager, roots, len(roots))
	if verbose != 0 {
		inputs := make([]string, Gia_ManCiNum(p))
		outputs := make([]string, Gia_ManCoNum(p))
		inputNames, outputNames := Gia_GetFakeNames(Gia_ManCiNum(p)), Gia_GetFakeNames(Gia_ManCoNum(p))
		for i := range inputs {
			inputs[i] = Vec_PtrEntry(inputNames, i).(string)
		}
		for i := range outputs {
			outputs[i] = Vec_PtrEntry(outputNames, i).(string)
		}
		dsd.Dsd_TreePrint(os.Stdout, manager, inputs, outputs, 0, -1)
	}
	local := cudd.Cudd_Init(uint32(Gia_ManCiNum(p)), 0, 0, 0, 0)
	defer cudd.Cudd_Quit(local)
	cudd.Cudd_zddVarsFromBddVars(local, 2)
	return Gia_ManRebuild(p, manager, local)
}

// Cudd_SharingSize is not exposed by the Go core. Count each regular node
// once across all roots, including the shared constant, as the source does.
func giaClpSharingSize(roots []*cudd.DdNode) int {
	seen := make(map[*cudd.DdNode]bool)
	var visit func(*cudd.DdNode)
	visit = func(node *cudd.DdNode) {
		node = cudd.Cudd_Regular(node)
		if node == nil || seen[node] {
			return
		}
		seen[node] = true
		if !cudd.Cudd_IsConstant(node) {
			visit(node.T)
			visit(node.E)
		}
	}
	for _, node := range roots {
		visit(node)
	}
	return len(seen)
}

func Gia_ManCollapseTestTest(p *Gia_Man_t) {
	out := Gia_ManCollapseTest(p, 0)
	Gia_ManPrintStats(p, nil)
	Gia_ManPrintStats(out, nil)
	Gia_ManStop(out)
}
