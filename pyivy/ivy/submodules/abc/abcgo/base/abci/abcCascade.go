package abci

// Characteristic-function and cofactor partition routines translated from
// src/base/abci/abcCascade.c. The four pure partition primitives live in
// cascade_partition.go and are shared with their source-derived tests.
import (
	"fmt"
	"math/bits"
	"sync"

	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/bdd/cudd/core"
	"github.com/glycerine/abcgo/bdd/extrab"
	"github.com/glycerine/abcgo/bdd/extrab/cascade"
	"github.com/glycerine/abcgo/bdd/reo"
)

const BDD_FUNC_MAX = 256

var cascadeRandState = struct {
	sync.Mutex
	rng *cascadeCRand
}{rng: newCascadeCRand(1)}

// cascadeCRand matches the glibc TYPE_3 rand() stream used by ABC on the
// supported Linux build. C rand() starts with seed 1 unless srand() was called.
type cascadeCRand struct {
	state [31]uint32
	index uint64
}

func newCascadeCRand(seed uint32) *cascadeCRand {
	if seed == 0 {
		seed = 1
	}
	sequence := make([]uint32, 344)
	sequence[0] = seed
	for i := 1; i < 31; i++ {
		sequence[i] = uint32(uint64(16807) * uint64(sequence[i-1]) % 2147483647)
	}
	sequence[31], sequence[32], sequence[33] = sequence[0], sequence[1], sequence[2]
	for i := 34; i < len(sequence); i++ {
		sequence[i] = sequence[i-31] + sequence[i-3]
	}
	rng := &cascadeCRand{index: 344}
	for i := 313; i < 344; i++ {
		rng.state[i%31] = sequence[i]
	}
	return rng
}

func (rng *cascadeCRand) next() uint32 {
	i := rng.index
	value := rng.state[(i-31)%31] + rng.state[(i-3)%31]
	rng.state[i%31] = value
	rng.index++
	return value >> 1
}

func cascadeRandN(n int) int {
	cascadeRandState.Lock()
	defer cascadeRandState.Unlock()
	return int(cascadeRandState.rng.next() % uint32(n))
}

func Abc_ResBuildBdd(n *Abc_Ntk_t, dd *core.DdManager) *core.DdNode {
	if n == nil || dd == nil || abc.Abc_NtkIsSopLogic(n) == 0 || abc.Abc_NtkCoNum(n) > 3 {
		return nil
	}
	values := make(map[*abc.Object]*core.DdNode, len(n.Objects()))
	for i, ci := range n.Inputs() {
		values[ci] = core.Cudd_bddIthVar(dd, i)
	}
	for _, node := range abc.Abc_NtkDfs(n, 0) {
		if len(node.Fanins()) == 0 {
			sop, _ := node.Data.(string)
			values[node] = core.Cudd_NotCond(dd.One, abc.Abc_SopIsConst0(sop) != 0)
			core.Cudd_Ref(values[node])
			continue
		}
		locals := make([]*core.DdNode, 0, len(node.Fanins()))
		for _, edge := range node.Fanins() {
			f := values[edge.Object]
			if f == nil {
				return nil
			}
			locals = append(locals, f)
		}
		sop, ok := node.Data.(string)
		if !ok {
			return nil
		}
		values[node] = abc.Abc_ConvertSopToBdd(dd, sop, locals)
		if values[node] == nil {
			return nil
		}
		core.Cudd_Ref(values[node])
	}
	result := dd.One
	core.Cudd_Ref(result)
	for i, co := range n.Outputs() {
		fanins := co.Fanins()
		if len(fanins) != 1 {
			return nil
		}
		f := values[fanins[0].Object]
		if f == nil {
			return nil
		}
		out := core.Cudd_bddIthVar(dd, i+len(n.Inputs()))
		part := core.Cudd_bddXnor(dd, f, out)
		core.Cudd_Ref(part)
		old := result
		result = core.Cudd_bddAnd(dd, old, part)
		if result == nil {
			return nil
		}
		core.Cudd_Ref(result)
		core.Cudd_RecursiveDeref(dd, old)
		core.Cudd_RecursiveDeref(dd, part)
	}
	for _, node := range abc.Abc_NtkDfs(n, 0) {
		core.Cudd_RecursiveDeref(dd, values[node])
	}
	core.Cudd_ReduceHeap(dd, core.CUDD_REORDER_SYMM_SIFT, 1)
	core.Cudd_Deref(result)
	return result
}

func Abc_ResCofCount(dd *core.DdManager, function *core.DdNode, mask uint32, check *int32) int32 {
	if dd == nil || function == nil {
		return 0
	}
	vars := make([]*core.DdNode, 0, 8)
	for i := 0; i < 32; i++ {
		if mask&(uint32(1)<<uint(i)) != 0 {
			if i >= len(dd.Vars) {
				return 0
			}
			vars = append(vars, core.Cudd_bddIthVar(dd, i))
		}
	}
	if len(vars) > 8 {
		return 0
	}
	pattern := make([]int8, 1<<uint(len(vars)))
	unique := make(map[*core.DdNode]int32)
	count := int32(0)
	for i := range pattern {
		cube := cascade.Extra_bddBitsToCube(dd, i, len(vars), vars, true)
		core.Cudd_Ref(cube)
		cofactor := core.Cudd_Cofactor(dd, function, cube)
		core.Cudd_Ref(cofactor)
		core.Cudd_RecursiveDeref(dd, cube)
		index, found := unique[cofactor]
		if !found {
			index = count
			unique[cofactor] = index
			count++
		} else {
			core.Cudd_RecursiveDeref(dd, cofactor)
		}
		pattern[i] = int8(index)
	}
	for cof := range unique {
		core.Cudd_RecursiveDeref(dd, cof)
	}
	if check != nil {
		*check = Abc_ResCheckNonStrict(pattern, int32(len(vars)), int32(abcBase2Log(int(count))))
	}
	return count
}

func abcBase2Log(n int) int { // exact Abc_Base2Log semantics (0->0, 1->1).
	if n < 2 {
		return n
	}
	return bits.Len(uint(n - 1))
}

func Abc_ResCost(dd *core.DdManager, function *core.DdNode, mask uint32, cofactors, check *int32) int32 {
	n := Abc_ResCofCount(dd, function, mask, check)
	log := abcBase2Log(int(n))
	if cofactors != nil {
		*cofactors = n
	}
	delta := int64(n) - (int64(1) << uint(log-1))
	return int32(10000*log + int(delta*delta))
}

func Abc_ResMigrate(dd *core.DdManager, function *core.DdNode, inputs int32, parts []uint32, part1, part2 int32) int32 {
	if dd == nil || function == nil || part1 < 0 || part2 < 0 || int(part1) >= len(parts) || int(part2) >= len(parts) {
		return 0
	}
	initial := [2]uint32{parts[part1], parts[part2]}
	bestParts := initial
	best := Abc_ResCost(dd, function, initial[0], nil, nil) + Abc_ResCost(dd, function, initial[1], nil, nil)
	changed := int32(0)
	for i := int32(0); i < inputs; i++ {
		if parts[part1]&(uint32(1)<<uint(i)) == 0 {
			continue
		}
		for k := int32(0); k < inputs; k++ {
			if parts[part2]&(uint32(1)<<uint(k)) == 0 || i == k {
				continue
			}
			bitsToSwap := (uint32(1) << uint(i)) | (uint32(1) << uint(k))
			parts[part1] ^= bitsToSwap
			parts[part2] ^= bitsToSwap
			cost := Abc_ResCost(dd, function, parts[part1], nil, nil) + Abc_ResCost(dd, function, parts[part2], nil, nil)
			// Preserve the original update order: each candidate is compared to
			// the starting cost, and the last improving candidate is retained.
			if cost < best {
				bestParts = [2]uint32{parts[part1], parts[part2]}
				changed = 1
			}
			parts[part1] ^= bitsToSwap
			parts[part2] ^= bitsToSwap
		}
	}
	parts[part1], parts[part2] = bestParts[0], bestParts[1]
	return changed
}

func Abc_ResPrint(dd *core.DdManager, function *core.DdNode, inputs int32, parts []uint32, count int32) {
	costAll := int32(0)
	for i := int32(0); i < count; i++ {
		var cofactors, check int32
		cost := Abc_ResCost(dd, function, parts[i], &cofactors, &check)
		costAll += cost
		for k := int32(0); k < inputs; k++ {
			if parts[i]&(uint32(1)<<uint(k)) != 0 {
				fmt.Printf("%c", 'a'+k)
			} else {
				fmt.Print("-")
			}
		}
		fmt.Printf(" %2d %d-%d %6d   ", cofactors, abcBase2Log(int(cofactors)), check, cost)
	}
	fmt.Printf("%4d\n", costAll)
}

func Abc_ResPrintAllCofs(dd *core.DdManager, function *core.DdNode, inputs, maxCofs int32) {
	for mask := 0; mask < (1 << uint(inputs)); mask++ {
		weight := bits.OnesCount(uint(mask))
		if weight < 3 || weight > 6 {
			continue
		}
		var cofactors, check int32
		cost := Abc_ResCost(dd, function, uint32(mask), &cofactors, &check)
		if cofactors > maxCofs {
			continue
		}
		for k := int32(0); k < inputs; k++ {
			if uint32(mask)&(uint32(1)<<uint(k)) != 0 {
				fmt.Printf("%c", 'a'+k)
			} else {
				fmt.Print("-")
			}
		}
		fmt.Printf("  n=%2d  c=%2d  l=%d-%d   %6d\n", weight, cofactors, abcBase2Log(int(cofactors)), check, cost)
	}
}

func Abc_ResSwapRandom(dd *core.DdManager, function *core.DdNode, inputs int32, parts []uint32, count, times int32) {
	if inputs < 2 {
		return
	}
	for swaps := int32(0); swaps < times; {
		i, k := cascadeRandN(int(inputs)), cascadeRandN(int(inputs))
		if i == k {
			continue
		}
		p1, p2 := -1, -1
		for p := 0; p < int(count); p++ {
			if parts[p]&(uint32(1)<<uint(i)) != 0 {
				p1 = p
			}
			if parts[p]&(uint32(1)<<uint(k)) != 0 {
				p2 = p
			}
		}
		if p1 == p2 || p1 < 0 || p2 < 0 {
			continue
		}
		mask := uint32(1)<<uint(i) | uint32(1)<<uint(k)
		parts[p1] ^= mask
		parts[p2] ^= mask
		swaps++
	}
}

func Abc_ResPartition(dd *core.DdManager, function *core.DdNode, inputs int32) {
	if dd == nil || function == nil || inputs < 0 {
		return
	}
	fmt.Printf("Ins =%3d. Outs =%2d. Nodes =%3d. Supp =%2d.\n", inputs, len(dd.Vars)-int(inputs), core.Cudd_DagSize(function), core.Cudd_SupportSize(dd, function))
	if core.Cudd_SupportSize(dd, function) <= 6 {
		fmt.Println("Support is less or equal than 6")
		return
	}
	if inputs > 24 {
		return
	}
	nParts := int32(2)
	if inputs > 12 {
		nParts = 3
	}
	if inputs > 18 {
		nParts = 4
	}
	parts := make([]uint32, 10)
	Abc_ResStartPart(inputs, parts, nParts)
	Abc_ResPrint(dd, function, inputs, parts, nParts)
	for iter := 0; iter < 5; iter++ {
		if iter > 0 {
			fmt.Println("Randomizing... ")
			Abc_ResSwapRandom(dd, function, inputs, parts, nParts, 20)
			Abc_ResPrint(dd, function, inputs, parts, nParts)
		}
		changed := int32(1)
		for changed != 0 {
			changed = 0
			for i := int32(0); i < nParts; i++ {
				for j := i + 1; j < nParts; j++ {
					changed |= Abc_ResMigrate(dd, function, inputs, parts, i, j)
					Abc_ResPrint(dd, function, inputs, parts, nParts)
				}
			}
		}
	}
}

func Abc_ResPartitionTest(n *Abc_Ntk_t) {
	if n == nil {
		return
	}
	dd := core.Cudd_Init(uint32(abc.Abc_NtkCiNum(n)+abc.Abc_NtkCoNum(n)), 0, core.CUDD_UNIQUE_SLOTS, core.CUDD_CACHE_SLOTS, 0)
	if dd == nil {
		return
	}
	function := Abc_ResBuildBdd(n, dd)
	if function != nil {
		core.Cudd_Ref(function)
		Abc_ResPartition(dd, function, abc.Abc_NtkCiNum(n))
		core.Cudd_RecursiveDeref(dd, function)
	}
	extrab.Extra_StopManager(dd)
}

func Abc_NtkBddCofCount(dd *core.DdManager, function *core.DdNode, vars []*core.DdNode, count int32) int32 {
	if dd == nil || function == nil || count < 0 || int(count) > len(vars) || count > 20 {
		return 0
	}
	seen := make(map[*core.DdNode]struct{})
	for code := 0; code < 1<<uint(count); code++ {
		cube := cascade.Extra_bddBitsToCube(dd, code, int(count), vars, true)
		core.Cudd_Ref(cube)
		cofactor := core.Cudd_Cofactor(dd, function, cube)
		core.Cudd_Ref(cofactor)
		core.Cudd_RecursiveDeref(dd, cube)
		if _, exists := seen[cofactor]; exists {
			core.Cudd_RecursiveDeref(dd, cofactor)
		} else {
			seen[cofactor] = struct{}{}
		}
	}
	for cofactor := range seen {
		core.Cudd_RecursiveDeref(dd, cofactor)
	}
	return int32(len(seen))
}

func Abc_NtkExploreCofs2(dd *core.DdManager, function *core.DdNode, vars []*core.DdNode, inputs, lutSize int32) {
	fmt.Printf("Inputs = %2d.  Nodes = %2d.  LutSize = %2d.\n", inputs, core.Cudd_DagSize(function), lutSize)
	for i := int32(0); i <= inputs-lutSize; i++ {
		fmt.Printf("[%2d %2d] : %3d\n", i, i+lutSize-1, Abc_NtkBddCofCount(dd, function, vars[i:i+lutSize], lutSize))
	}
}

func Abc_NtkExploreCofs(dd *core.DdManager, function *core.DdNode, vars []*core.DdNode, inputs, lutSize int32) {
	if dd == nil || function == nil {
		return
	}
	newDD := core.Cudd_Init(uint32(len(dd.Vars)), 0, core.CUDD_UNIQUE_SLOTS, core.CUDD_CACHE_SLOTS, 0)
	if newDD == nil {
		return
	}
	defer core.Cudd_Quit(newDD)
	core.Cudd_ShuffleHeap(newDD, dd.InvPerm)
	copyFunc := core.Cudd_bddTransfer(dd, newDD, function)
	if copyFunc == nil {
		return
	}
	for mask := 0; mask < 1<<uint(inputs); mask++ {
		weight := bits.OnesCount(uint(mask))
		if weight != int(lutSize) && weight != int(lutSize-1) && weight != int(lutSize-2) {
			continue
		}
		selected := make([]*core.DdNode, 0, weight)
		for k := int32(0); k < inputs; k++ {
			if mask&(1<<uint(k)) != 0 {
				selected = append(selected, core.Cudd_bddIthVar(newDD, int(k)))
			}
		}
		cofs := Abc_NtkBddCofCount(newDD, copyFunc, selected, int32(len(selected)))
		if cofs > 8 {
			continue
		}
		for k := int32(0); k < inputs; k++ {
			if mask&(1<<uint(k)) == 0 {
				fmt.Print("-")
			} else {
				fmt.Printf("%c", 'a'+k)
			}
		}
		fmt.Printf(" : %2d\n", cofs)
	}
}

func Abc_NtkBddFindAddConst(dd *core.DdManager, function *core.DdNode, outputs int32) *core.DdNode {
	mask := 0
	for i := int32(0); i < outputs; i++ {
		if function == nil || core.Cudd_IsConstant(function) {
			return nil
		}
		low, high := core.Cudd_E(function), core.Cudd_T(function)
		if core.Cudd_IsComplement(function) {
			low, high = core.Cudd_Not(low), core.Cudd_Not(high)
		}
		if low == dd.LogicZero {
			mask ^= 1 << uint(i)
			function = high
		} else if high == dd.LogicZero {
			function = low
		} else {
			return nil
		}
	}
	if function != dd.One {
		return nil
	}
	return core.Cudd_addConst(dd, float64(mask))
}

func Abc_NtkBddToAdd_rec(dd *core.DdManager, function *core.DdNode, outputs int32, table map[*core.DdNode]*core.DdNode, complement bool) *core.DdNode {
	if function == nil {
		return nil
	}
	if cached, ok := table[function]; ok {
		return cached
	}
	if int(function.Index) >= len(dd.Vars)-int(outputs) {
		result := Abc_NtkBddFindAddConst(dd, core.Cudd_NotCond(function, complement), outputs)
		core.Cudd_Ref(result)
		table[function] = result
		return result
	}
	low, high := function.E, function.T
	lowF := core.Cudd_IsComplement(low)
	if core.Cudd_IsComplement(function) {
		low, high = core.Cudd_Not(low), core.Cudd_Not(high)
	}
	low = core.Cudd_Regular(low)
	high = core.Cudd_Regular(high)
	a0 := Abc_NtkBddToAdd_rec(dd, low, outputs, table, complement != lowF)
	a1 := Abc_NtkBddToAdd_rec(dd, high, outputs, table, complement)
	result := core.Cudd_addIte(dd, core.Cudd_addIthVar(dd, int(function.Index)), a1, a0)
	core.Cudd_Ref(result)
	table[function] = result
	return result
}

func Abc_NtkBddToAdd(dd *core.DdManager, function *core.DdNode, outputs int32) *core.DdNode {
	if dd == nil || function == nil {
		return nil
	}
	table := make(map[*core.DdNode]*core.DdNode)
	result := Abc_NtkBddToAdd_rec(dd, core.Cudd_Regular(function), outputs, table, core.Cudd_IsComplement(function))
	for _, cached := range table {
		core.Cudd_RecursiveDeref(dd, cached)
	}
	core.Cudd_Deref(result)
	return result
}

func Abc_NtkAddToBdd_rec(dd *core.DdManager, function *core.DdNode, inputs, outputs int32, table map[*core.DdNode]*core.DdNode) *core.DdNode {
	if function == nil {
		return nil
	}
	if result, ok := table[function]; ok {
		return result
	}
	if core.Cudd_IsConstant(function) {
		result := cascade.Extra_bddBitsToCube(dd, int(function.Value), int(outputs), dd.Vars[inputs:], true)
		core.Cudd_Ref(result)
		table[function] = result
		return result
	}
	if int32(function.Index) >= inputs {
		return nil
	}
	low := Abc_NtkAddToBdd_rec(dd, function.E, inputs, outputs, table)
	high := Abc_NtkAddToBdd_rec(dd, function.T, inputs, outputs, table)
	result := core.Cudd_bddIte(dd, core.Cudd_bddIthVar(dd, int(function.Index)), high, low)
	core.Cudd_Ref(result)
	table[function] = result
	return result
}

func Abc_NtkAddToBdd(dd *core.DdManager, function *core.DdNode, inputs, outputs int32) *core.DdNode {
	if dd == nil || function == nil {
		return nil
	}
	table := make(map[*core.DdNode]*core.DdNode)
	result := Abc_NtkAddToBdd_rec(dd, function, inputs, outputs, table)
	for _, cached := range table {
		core.Cudd_RecursiveDeref(dd, cached)
	}
	core.Cudd_Deref(result)
	return result
}

func Abc_NtkBddDecCharFunc(dd *core.DdManager, functions []*core.DdNode, outputs int32, mask uint32, nBits int32) *core.DdNode {
	if dd == nil || outputs < 0 || int(outputs) > len(functions) || int(outputs) > len(dd.Vars) {
		return nil
	}
	result := dd.One
	count := int32(0)
	for i := int32(0); i < outputs; i++ {
		if mask&(uint32(1)<<uint(i)) == 0 {
			continue
		}
		if functions[i] == nil {
			return nil
		}
		count++
		out := core.Cudd_bddIthVar(dd, len(dd.Vars)-int(outputs)+int(i))
		result = core.Cudd_bddAnd(dd, result, core.Cudd_Not(core.Cudd_bddXor(dd, functions[i], out)))
		if result == nil {
			return nil
		}
	}
	if count != nBits {
		return nil
	}
	return result
}

func Abc_NtkBddDecTry(reorder *reo.Manager, dd *core.DdManager, functions []*core.DdNode, inputs, outputs int32, mask uint32, nBits int32) *core.DdNode {
	return Abc_NtkBddDecCharFunc(dd, functions, outputs, mask, nBits)
}

func Abc_NtkBddDecInt(reorder *reo.Manager, dd *core.DdManager, functions []*core.DdNode, inputs, outputs int32) *core.DdNode {
	if outputs <= 0 || outputs >= 32 {
		return nil
	}
	return Abc_NtkBddDecTry(reorder, dd, functions, inputs, outputs, ^(uint32(1) << uint(32-outputs)), outputs)
}

func Abc_NtkCreateFromCharFunc(n *Abc_Ntk_t, dd *core.DdManager, function *core.DdNode) *Abc_Ntk_t {
	if n == nil || dd == nil || function == nil {
		return nil
	}
	out := abc.Abc_NtkAlloc(abc.ABC_NTK_LOGIC, abc.ABC_FUNC_BDD, 1)
	out.Name = n.Name
	root := abc.Abc_NtkCreateNode(out)
	for _, old := range n.Inputs() {
		pi := abc.Abc_NtkCreatePi(out)
		abc.Abc_ObjAddFanin(root, pi)
		abc.Abc_ObjAssignName(pi, old.Name(), "")
	}
	for _, old := range n.Outputs() {
		pi := abc.Abc_NtkCreatePi(out)
		abc.Abc_ObjAddFanin(root, pi)
		abc.Abc_ObjAssignName(pi, old.Name(), "")
	}
	newDD := out.FunctionManager.(*core.DdManager)
	root.Data = extrab.Extra_TransferLevelByLevel(dd, newDD, function)
	if root.Data == nil {
		return nil
	}
	core.Cudd_Ref(root.Data.(*core.DdNode))
	po := abc.Abc_NtkCreatePo(out)
	abc.Abc_ObjAddFanin(po, root)
	abc.Abc_ObjAssignName(po, "out", "")
	if abc.Abc_NtkCheck(out) == 0 {
		fmt.Println("Abc_NtkCreateFromCharFunc(): Network check has failed.")
	}
	return out
}

func Abc_NtkBddDec(n *Abc_Ntk_t, verbose int32) *Abc_Ntk_t {
	if n == nil || abc.Abc_NtkIsStrash(n) == 0 || abc.Abc_NtkCoNum(n) > BDD_FUNC_MAX {
		return nil
	}
	dd := Abc_NtkBuildGlobalBdds(n, 1000000, 0, 1, 0, verbose)
	if dd == nil {
		Abc_Print(-1, "Construction of global BDDs has failed.\n")
		return nil
	}
	defer Abc_NtkFreeGlobalBdds(n, 1)
	functions := make([]*core.DdNode, abc.Abc_NtkCoNum(n))
	for i, co := range n.Outputs() {
		functions[i] = Abc_ObjGlobalBdd(co)
	}
	if len(dd.Vars) != len(n.Inputs()) {
		return nil
	}
	for i := 0; i < len(functions); i++ {
		core.Cudd_addNewVarAtLevel(dd, len(dd.Vars))
	}
	reorder := reo.Extra_ReorderInit(len(n.Inputs()), 1000)
	defer reo.Extra_ReorderQuit(reorder)
	reo.Extra_ReorderSetMinimizationType(reorder, reo.REO_MINIMIZE_WIDTH)
	reo.Extra_ReorderSetVerification(reorder, 1)
	reo.Extra_ReorderSetVerbosity(reorder, 1)
	characteristic := Abc_NtkBddDecInt(reorder, dd, functions, int32(len(n.Inputs())), int32(len(functions)))
	if characteristic == nil {
		return nil
	}
	core.Cudd_Ref(characteristic)
	defer core.Cudd_RecursiveDeref(dd, characteristic)
	Abc_NtkExploreCofs(dd, characteristic, dd.Vars, int32(len(n.Inputs())), 6)
	return abc.Abc_NtkDup(n)
}
