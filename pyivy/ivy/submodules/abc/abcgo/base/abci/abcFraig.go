package abci

import (
	"fmt"
	"github.com/glycerine/abcgo/base/abc"
	abcmain "github.com/glycerine/abcgo/base/main"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/proof/fraig"
)

type Fraig_Man_t = fraig.Fraig_Man_t
type Fraig_Node_t = fraig.Fraig_Node_t
type Fraig_Params_t = fraig.Fraig_Params_t

// Abc_NtkFraig runs the CUDD-era ABC interface over a strashed network.
func Abc_NtkFraig(network *Abc_Ntk_t, params any, allNodes, useExdc int32) *Abc_Ntk_t {
	p, ok := params.(*Fraig_Params_t)
	if !ok || p == nil {
		panic("Abc_NtkFraig requires *Fraig_Params_t")
	}
	if useExdc != 0 && network.Exdc == nil {
		fmt.Println("Warning: Networks has no EXDC.")
		useExdc = 0
	}
	manager := Abc_NtkToFraig(network, p, allNodes, useExdc).(*Fraig_Man_t)
	if p.FTryProve != 0 {
		fraig.Fraig_ManProveMiter(manager)
	}
	var result *Abc_Ntk_t
	if useExdc != 0 {
		result = Abc_NtkFromFraig2(manager, network)
	} else {
		result = Abc_NtkFromFraig(manager, network)
	}
	fraig.Fraig_ManFree(manager)
	if network.Exdc != nil && result != nil {
		result.Exdc = abc.Abc_NtkDup(network.Exdc)
	}
	if result == nil || abc.Abc_NtkCheck(result) == 0 {
		fmt.Println("Abc_NtkFraig: The network check has failed.")
		if result != nil {
			abc.Abc_NtkDelete(result)
		}
		return nil
	}
	return result
}

// Abc_NtkToFraig translates each AIG edge without discarding fanin polarity.
func Abc_NtkToFraig(network *Abc_Ntk_t, params any, allNodes, useExdc int32) any {
	if abc.Abc_NtkIsStrash(network) == 0 {
		panic("Abc_NtkToFraig requires a strashed network")
	}
	p, ok := params.(*Fraig_Params_t)
	if !ok {
		panic("Abc_NtkToFraig requires *Fraig_Params_t")
	}
	manager := fraig.Fraig_ManCreate(p)
	mapping := make(map[*Abc_Obj_t]*Fraig_Node_t)
	for _, node := range abc.Abc_NtkNodes(network) {
		node.CopyData = nil
	}
	mapping[abc.Abc_AigConst1(network).Object] = fraig.Fraig_ManReadConst1(manager)
	for i, ci := range network.Inputs() {
		mapping[ci] = fraig.Fraig_ManReadIthVar(manager, int32(i))
	}
	for _, node := range abc.Abc_AigDfs(network, allNodes, 0) {
		if abc.Abc_AigNodeIsAnd(node) == 0 {
			continue
		}
		f := node.Fanins()
		a, b := mapping[f[0].Object], mapping[f[1].Object]
		if a == nil || b == nil {
			panic("AIG traversal encountered an unmapped fanin")
		}
		mapping[node] = fraig.Fraig_NodeAnd(manager, fraig.Fraig_NotCond(a, f[0].Complement), fraig.Fraig_NotCond(b, f[1].Complement))
		node.CopyData = mapping[node]
	}
	if useExdc != 0 {
		Abc_NtkFraigRemapUsingExdc(manager, network, mapping)
	}
	for _, co := range network.Outputs() {
		f := co.Fanins()[0]
		mapped := mapping[f.Object]
		if mapped == nil {
			panic("unmapped AIG output")
		}
		fraig.Fraig_ManSetPo(manager, fraig.Fraig_NotCond(mapped, f.Complement))
	}
	return manager
}

// Abc_NtkToFraigExdc returns the first EXDC output, matching the source contract.
func Abc_NtkToFraigExdc(manager *Fraig_Man_t, main, exdc *Abc_Ntk_t) *Fraig_Node_t {
	strashed := abc.Abc_NtkStrash(exdc, 0, 0, 0)
	defer abc.Abc_NtkDelete(strashed)
	byName := make(map[string]int32)
	for i, ci := range main.Inputs() {
		byName[ci.Name()] = int32(i)
	}
	mapping := map[*Abc_Obj_t]*Fraig_Node_t{abc.Abc_AigConst1(strashed).Object: fraig.Fraig_ManReadConst1(manager)}
	for _, ci := range strashed.Inputs() {
		i, ok := byName[ci.Name()]
		if !ok {
			panic("EXDC input does not match a main-network CI: " + ci.Name())
		}
		mapping[ci] = fraig.Fraig_ManReadIthVar(manager, i)
	}
	for _, node := range abc.Abc_AigDfs(strashed, 1, 0) {
		if abc.Abc_AigNodeIsAnd(node) == 0 {
			continue
		}
		f := node.Fanins()
		mapping[node] = fraig.Fraig_NodeAnd(manager, fraig.Fraig_NotCond(mapping[f[0].Object], f[0].Complement), fraig.Fraig_NotCond(mapping[f[1].Object], f[1].Complement))
	}
	co := strashed.Outputs()[0].Fanins()[0]
	return fraig.Fraig_NotCond(mapping[co.Object], co.Complement)
}

// Abc_NtkFraigRemapUsingExdc groups nodes by their function under the global don't-care.
func Abc_NtkFraigRemapUsingExdc(manager *Fraig_Man_t, network *Abc_Ntk_t, mapping map[*Abc_Obj_t]*Fraig_Node_t) {
	if network.Exdc == nil {
		return
	}
	dc := Abc_NtkToFraigExdc(manager, network, network.Exdc)
	classes := make(map[*Fraig_Node_t][]*Abc_Obj_t)
	for _, node := range abc.Abc_NtkNodes(network) {
		if n := mapping[node]; n != nil {
			key := fraig.Fraig_Regular(fraig.Fraig_NodeAnd(manager, n, fraig.Fraig_Not(dc)))
			classes[key] = append(classes[key], node)
		}
	}
	abc.Abc_AigSetNodePhases(network)
	for _, members := range classes {
		if len(members) < 2 {
			continue
		}
		best := members[0]
		for _, node := range members[1:] {
			if node.Level < best.Level {
				best = node
			}
		}
		for _, node := range members {
			mapping[node] = abcFraigNotCond(mapping[best], node.Phase != best.Phase)
			node.CopyData = mapping[node]
		}
	}
}

func abcFraigNotCond(n *Fraig_Node_t, neg bool) *Fraig_Node_t { return fraig.Fraig_NotCond(n, neg) }

func Abc_NtkFromFraig(manager *Fraig_Man_t, network *Abc_Ntk_t) *Abc_Ntk_t {
	result := abc.Abc_NtkStartFrom(network, abc.ABC_NTK_STRASH, abc.ABC_FUNC_AIG)
	aig, ok := result.FunctionManager.(*abc.Abc_Aig_t)
	if !ok {
		panic("strashed network has no AIG manager")
	}
	memo := make(map[*Fraig_Node_t]abc.Edge)
	memo[fraig.Fraig_ManReadConst1(manager)] = abc.Abc_AigConst1(result)
	fraig.Fraig_NodeSetData1Copy(fraig.Fraig_ManReadConst1(manager), abc.Abc_AigConst1(result))
	for i, ci := range network.Inputs() {
		copy := abc.Abc_ObjCopyEdge(ci)
		memo[fraig.Fraig_ManReadIthVar(manager, int32(i))] = copy
		fraig.Fraig_NodeSetData1Copy(fraig.Fraig_ManReadIthVar(manager, int32(i)), copy)
	}
	for i := range network.Outputs() {
		edge := abcFraigBuildAIG(manager, fraig.Fraig_ManReadOutputs(manager)[i], aig, memo)
		if err := result.Outputs()[i].Network().AddFanin(result.Outputs()[i], edge); err != nil {
			panic(err)
		}
	}
	fraig.Fraig_NodeSetData1Copy(fraig.Fraig_ManReadConst1(manager), abc.Abc_AigConst1(result))
	abc.Abc_NtkReassignIds(result)
	abcFraigClearNodeScratch(network)
	return result
}

func abcFraigBuildAIG(manager *Fraig_Man_t, node *Fraig_Node_t, aig *abc.Abc_Aig_t, memo map[*Fraig_Node_t]abc.Edge) abc.Edge {
	regular := fraig.Fraig_Regular(node)
	if edge, ok := memo[regular]; ok {
		return abc.Abc_ObjNotCond(edge, boolInt32(fraig.Fraig_IsComplement(node)))
	}
	if fraig.Fraig_NodeIsVar(regular) != 0 {
		panic("FRAIG input is absent from the source network")
	}
	a := abcFraigBuildAIG(manager, fraig.Fraig_NodeReadOne(regular), aig, memo)
	b := abcFraigBuildAIG(manager, fraig.Fraig_NodeReadTwo(regular), aig, memo)
	edge := abc.Abc_AigAnd(aig, a, b)
	edge.Object.Phase = fraig.Fraig_NodeReadSimInv(regular) != 0
	memo[regular] = edge
	fraig.Fraig_NodeSetData1Copy(regular, edge)
	if fraig.Fraig_NodeReadRepr(regular) == nil {
		var alternatives []abc.Edge
		for next := fraig.Fraig_NodeReadNextE(regular); next != nil; next = fraig.Fraig_NodeReadNextE(next) {
			alternatives = append(alternatives, abcFraigBuildAIG(manager, next, aig, memo))
		}
		best := edge
		for _, alt := range alternatives {
			if alt.Object.Level < best.Object.Level {
				best = alt
			}
		}
		tail := best.Object
		choices := append([]abc.Edge{edge}, alternatives...)
		for _, alt := range choices {
			if alt.Object == best.Object {
				continue
			}
			if tail.Choice == nil {
				tail.Choice = alt.Object
				tail = alt.Object
			}
		}
		if best.Object != edge.Object {
			representative := abc.Abc_ObjNotCond(best, boolInt32(edge.Object.Phase != best.Object.Phase))
			memo[regular] = representative
			fraig.Fraig_NodeSetData1Copy(regular, representative)
		}
	}
	return abc.Abc_ObjNotCond(edge, boolInt32(fraig.Fraig_IsComplement(node)))
}

// The no-choice reconstruction preserves the source network's output functions
// while using the FRAIG representative graph as the shared AIG.
func Abc_NtkFromFraig2(manager *Fraig_Man_t, network *Abc_Ntk_t) *Abc_Ntk_t {
	best := make(map[*Fraig_Node_t]*Abc_Obj_t)
	for i, ci := range network.Inputs() {
		best[fraig.Fraig_Regular(fraig.Fraig_ManReadIthVar(manager, int32(i)))] = ci
	}
	best[fraig.Fraig_Regular(fraig.Fraig_ManReadConst1(manager))] = abc.Abc_AigConst1(network).Object
	for _, node := range abc.Abc_NtkNodes(network) {
		if f, ok := node.CopyData.(*Fraig_Node_t); ok && f != nil {
			k := fraig.Fraig_Regular(f)
			if old := best[k]; old == nil || node.Level < old.Level {
				best[k] = node
			}
		}
	}
	reprs := vec.Vec_PtrStart(int(abc.Abc_NtkObjNumMax(network)))
	for _, node := range abc.Abc_NtkNodes(network) {
		if f, ok := node.CopyData.(*Fraig_Node_t); ok && f != nil {
			if r := best[fraig.Fraig_Regular(f)]; r != nil && r != node {
				vec.Vec_PtrWriteEntry(reprs, node.ID, r)
			}
		}
	}
	result := abc.Abc_NtkStartFrom(network, abc.ABC_NTK_STRASH, abc.ABC_FUNC_AIG)
	aig, ok := result.FunctionManager.(*abc.Abc_Aig_t)
	if !ok {
		panic("strashed network has no AIG manager")
	}
	for i, ci := range network.Inputs() {
		ci.CopyEdge = abc.Abc_NtkCi(result, int32(i)).Signal()
	}
	for _, co := range network.Outputs() {
		fan := co.Fanins()[0]
		abcFraigFrom2Rec(network, fan.Object, reprs, aig, make(map[*Abc_Obj_t]bool))
	}
	vec.Vec_PtrFree(reprs)
	abc.Abc_NtkFinalize(network, result)
	abcFraigClearNodeScratch(network)
	return result
}

func abcFraigClearNodeScratch(network *Abc_Ntk_t) {
	for _, node := range abc.Abc_NtkNodes(network) {
		node.CopyData = nil
	}
}
func Abc_NodeFromFraig_rec(network *Abc_Ntk_t, node *Fraig_Node_t) abc.Edge {
	return abcFraigNodeFromFraig(network, node)
}
func Abc_NtkFromFraig2_rec(network *Abc_Ntk_t, node *Abc_Obj_t, reprs *Vec_Ptr_t) {
	aig, ok := network.FunctionManager.(*abc.Abc_Aig_t)
	if !ok {
		panic("strashed network has no AIG manager")
	}
	abcFraigFrom2Rec(network, node, reprs, aig, make(map[*Abc_Obj_t]bool))
}

func abcFraigNodeFromFraig(network *Abc_Ntk_t, node *Fraig_Node_t) abc.Edge {
	r := fraig.Fraig_Regular(node)
	if value := fraig.Fraig_NodeReadData1Copy(r); value != nil {
		if edge, ok := value.(abc.Edge); ok {
			return abc.Abc_ObjNotCond(edge, boolInt32(fraig.Fraig_IsComplement(node)))
		}
		if edge, ok := value.(*Abc_Obj_t); ok {
			return abc.Abc_ObjNotCond(edge, boolInt32(fraig.Fraig_IsComplement(node)))
		}
	}
	aig, ok := network.FunctionManager.(*abc.Abc_Aig_t)
	if !ok {
		panic("strashed network has no AIG manager")
	}
	a := abcFraigNodeFromFraig(network, fraig.Fraig_NodeReadOne(r))
	b := abcFraigNodeFromFraig(network, fraig.Fraig_NodeReadTwo(r))
	edge := abc.Abc_AigAnd(aig, a, b)
	edge.Object.Phase = fraig.Fraig_NodeReadSimInv(r) != 0
	fraig.Fraig_NodeSetData1Copy(r, edge)
	return abc.Abc_ObjNotCond(edge, boolInt32(fraig.Fraig_IsComplement(node)))
}

func abcFraigFrom2Rec(_ *Abc_Ntk_t, node *Abc_Obj_t, reprs *Vec_Ptr_t, aig *abc.Abc_Aig_t, visited map[*Abc_Obj_t]bool) {
	if node == nil || len(node.Fanins()) < 2 || visited[node] {
		return
	}
	visited[node] = true
	if r, ok := vec.Vec_PtrEntry(reprs, node.ID).(*Abc_Obj_t); ok && r != nil {
		abcFraigFrom2Rec(nil, r, reprs, aig, visited)
		node.CopyEdge = abc.Abc_ObjNotCond(r.CopyEdge, boolInt32(node.Phase != r.Phase))
		return
	}
	f := node.Fanins()
	for _, e := range f {
		abcFraigFrom2Rec(nil, e.Object, reprs, aig, visited)
	}
	node.CopyEdge = abc.Abc_AigAnd(aig, abc.Abc_ObjNotCond(f[0].Object.CopyEdge, boolInt32(f[0].Complement)), abc.Abc_ObjNotCond(f[1].Object.CopyEdge, boolInt32(f[1].Complement)))
}

func Abc_NtkFraigTrust(network *Abc_Ntk_t) *Abc_Ntk_t {
	if abc.Abc_NtkIsSopLogic(network) == 0 || Abc_NtkFraigTrustCheck(network) == 0 {
		return nil
	}
	result := abc.Abc_NtkStartFrom(network, abc.ABC_NTK_STRASH, abc.ABC_FUNC_AIG)
	Abc_NtkFraigTrustOne(network, result)
	abc.Abc_NtkFinalize(network, result)
	abc.Abc_NtkReassignIds(result)
	if abc.Abc_NtkCheck(result) == 0 {
		abc.Abc_NtkDelete(result)
		return nil
	}
	return result
}
func Abc_NtkFraigTrustCheck(network *Abc_Ntk_t) int32 {
	for _, node := range abc.Abc_NtkNodes(network) {
		n := len(node.Fanins())
		if n < 2 {
			continue
		}
		sop, _ := node.Data.(string)
		if n == 2 && abc.Abc_SopIsAndType(sop) != 0 {
			continue
		}
		if abc.Abc_SopIsOrType(sop) == 0 {
			return 0
		}
	}
	return 1
}
func Abc_NtkFraigTrustOne(network, result *Abc_Ntk_t) {
	if _, ok := result.FunctionManager.(*abc.Abc_Aig_t); !ok {
		panic("strashed network has no AIG manager")
	}
	for _, node := range abc.Abc_NtkDfs(network, 0) {
		if abc.Abc_ObjIsNode(node) == 0 {
			continue
		}
		edge := Abc_NodeFraigTrust(result, node)
		old := node
		if abc.Abc_NtkIsNetlist(network) != 0 && len(node.Fanouts()) != 0 {
			old = node.Fanouts()[0]
		}
		old.CopyEdge = edge
	}
}
func Abc_NodeFraigTrust(network *Abc_Ntk_t, node *Abc_Obj_t) abc.Edge {
	aig, ok := network.FunctionManager.(*abc.Abc_Aig_t)
	if !ok {
		panic("strashed network has no AIG manager")
	}
	f := node.Fanins()
	sop, _ := node.Data.(string)
	n := len(f)
	copyFanin := func(i int) abc.Edge {
		return abc.Abc_ObjNotCond(abc.Abc_ObjCopyEdge(f[i].Object), boolInt32(f[i].Complement))
	}
	switch n {
	case 0:
		return abc.Abc_ObjNotCond(abc.Abc_AigConst1(network), abc.Abc_SopIsConst0(sop))
	case 1:
		return abc.Abc_ObjNotCond(copyFanin(0), abc.Abc_SopIsInv(sop))
	case 2:
		if abc.Abc_SopIsAndType(sop) != 0 {
			return abc.Abc_AigAnd(aig, abc.Abc_ObjNotCond(copyFanin(0), 1-abc.Abc_SopGetIthCareLit(sop, 0)), abc.Abc_ObjNotCond(copyFanin(1), 1-abc.Abc_SopGetIthCareLit(sop, 1)))
		}
	}
	if abc.Abc_SopIsOrType(sop) == 0 {
		panic("unsupported SOP in trust reconstruction")
	}
	result := copyFanin(0)
	phase := abc.Abc_SopGetIthCareLit(sop, 0)
	tail := result.Object
	for i := 1; i < n; i++ {
		lit := copyFanin(i)
		if tail.Choice != nil {
			panic("choice tail already linked")
		}
		tail.Choice = lit.Object
		tail = lit.Object
		if phase != abc.Abc_SopGetIthCareLit(sop, int32(i)) {
			lit.Object.Phase = true
		}
	}
	return result
}

func Abc_NtkFraigStore(network *Abc_Ntk_t) int32 {
	strashed := abc.Abc_NtkStrash(network, 0, 0, 0)
	if strashed == nil {
		return 0
	}
	stored := abcmain.Abc_FrameReadStore()
	if len(stored) != 0 {
		first, ok := stored[0].(*Abc_Ntk_t)
		if !ok {
			Abc_NtkFraigStoreClean()
		} else if abc.Abc_NodeCompareCiCo(strashed, first) == 0 && abc.Abc_NtkCompareSignals(strashed, first, 1, 1) == 0 {
			fmt.Println("Trying to store the network with different primary inputs.")
			fmt.Println("The previously stored networks are deleted and this one is added.")
			Abc_NtkFraigStoreClean()
		}
	}
	abcmain.Abc_FrameSetNtkStore(strashed)
	return 1
}
func Abc_NtkFraigRestore(randomPatterns, dynamicPatterns, backtrackLimit int32) *Abc_Ntk_t {
	stored := abcmain.Abc_FrameReadStore()
	if len(stored) == 0 {
		fmt.Println("There are no network currently in storage.")
		return nil
	}
	if len(stored) > 1 {
		first := stored[0]
		stored[0] = stored[len(stored)-1]
		stored[len(stored)-1] = first
		abcmain.Abc_FrameSetNtkStoreSize(0)
		for _, n := range stored {
			abcmain.Abc_FrameSetNtkStore(n)
		}
	}
	network := stored[0].(*Abc_Ntk_t)
	words := int32(32)
	denominator := abc.Abc_NtkNodeNum(network) + abc.Abc_NtkCiNum(network)
	if denominator > 0 {
		words = min(words, int32((1<<27)/denominator))
	}
	p := &Fraig_Params_t{}
	fraig.Fraig_ParamsSetDefault(p)
	p.NPatsRand = choosePatternCount(randomPatterns, words*32)
	p.NPatsDyna = choosePatternCount(dynamicPatterns, words*32)
	p.NBTLimit = backtrackLimit
	p.FFuncRed, p.FFeedBack, p.FDist1Pats, p.FDoSparse, p.FChoicing = 1, 1, 1, 1, 1
	p.FTryProve, p.FInternal, p.FVerbose = 0, 1, 0
	store := abcFraigStoreVec(stored)
	result := Abc_NtkFraigPartitioned(store, p)
	vec.Vec_PtrFree(store)
	Abc_NtkFraigStoreClean()
	return result
}
func choosePatternCount(value, fallback int32) int32 {
	if value != 0 {
		return value
	}
	return fallback
}
func abcFraigStoreVec(values []any) *Vec_Ptr_t {
	v := vec.Vec_PtrAlloc(len(values))
	for _, x := range values {
		vec.Vec_PtrPush(v, x)
	}
	return v
}
func Abc_NtkFraigStoreClean() {
	for _, value := range abcmain.Abc_FrameReadStore() {
		if n, ok := value.(*Abc_Ntk_t); ok && n != nil {
			abc.Abc_NtkDelete(n)
		}
	}
	abcmain.Abc_FrameSetNtkStoreSize(0)
}
func Abc_NtkFraigStoreCheck(network *Abc_Ntk_t) {
	stored := len(abcmain.Abc_FrameReadStore())
	poCount := int(abc.Abc_NtkPoNum(network))
	if stored == 0 || poCount%stored != 0 {
		return
	}
	per := poCount / stored
	for i := 0; i < per; i++ {
		first := abc.Abc_NtkPo(network, int32(i)).Fanins()[0].Object
		for k := 1; k < stored; k++ {
			if abc.Abc_NtkPo(network, int32(k*per+i)).Fanins()[0].Object != first {
				fmt.Printf("Verification for PO #%d of network #%d has failed. The PO function is not used.\n", i+1, k+1)
			}
		}
	}
}
