package abci

import (
	"fmt"
	"os"

	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/misc/vec"
)

const SHARE_NUM = 2

type Abc_ShaMan_t struct {
	nMultiSize  int32
	fVerbose    int32
	pNtk        *Abc_Ntk_t
	vBuckets    *Vec_Ptr_t
	vObj2Lit    *Vec_Int_t
	nStartCols  int32
	nCountGates int32
	nFoundGates int32
}

func Abc_NtkSharePack(level, id int32) uint64 {
	return uint64(uint32(level))<<32 | uint64(uint32(id))
}
func Abc_NtkShareUnpackLev(num uint64) int32 { return int32(uint32(num >> 32)) }
func Abc_NtkShareUnpackId(num uint64) int32  { return int32(uint32(num)) }

func Abc_ShaManStart(network *Abc_Ntk_t) *Abc_ShaMan_t {
	return &Abc_ShaMan_t{pNtk: network, vObj2Lit: vec.Vec_IntAlloc(1000)}
}
func Abc_ShaManStop(manager *Abc_ShaMan_t) {
	if manager == nil {
		return
	}
	if manager.vBuckets != nil {
		for i := 0; i < vec.Vec_PtrSize(manager.vBuckets); i++ {
			bucket := vec.Vec_PtrEntry(manager.vBuckets, i).(*Vec_Ptr_t)
			for j := 0; j < vec.Vec_PtrSize(bucket); j++ {
				vec.Vec_IntFree(vec.Vec_PtrEntry(bucket, j).(*Vec_Int_t))
			}
			vec.Vec_PtrFree(bucket)
		}
		vec.Vec_PtrFree(manager.vBuckets)
	}
	vec.Vec_IntFree(manager.vObj2Lit)
}

func abcShareBool(value bool) int32 {
	if value {
		return 1
	}
	return 0
}
func abcShareObjLit(object *abc.Object, complement bool) int32 {
	return int32(object.ID<<1) | abcShareBool(complement)
}
func abcShareEdgeLit(edge abc.Edge) int32 { return abcShareObjLit(edge.Object, edge.Complement) }
func abcShareEdgeFromLit(network *Abc_Ntk_t, literal int32) abc.Edge {
	return abc.Edge{Object: network.Object(int(literal >> 1)), Complement: literal&1 != 0}
}
func abcSharePackObj(network *Abc_Ntk_t, object *abc.Object) uint64 {
	return Abc_NtkSharePack(abc.Abc_ObjLevel(object), abc.Abc_ObjId(object))
}

// Abc_NtkShareSuperXor expands a two-input XOR tree into its leaves and polarity.
func Abc_NtkShareSuperXor(object *Abc_Obj_t, polarity *int32, counter *int32) *vec.Vec_Wrd_t {
	network := abc.Abc_ObjNtk(object)
	if object == nil || abc.Abc_ObjIsComplement(object) != 0 || abc.Abc_NodeIsExorType(object) == 0 {
		panic("Abc_NtkShareSuperXor requires a regular XOR node")
	}
	super := vec.Vec_WrdAlloc(10)
	vec.Vec_WrdPush(super, abcSharePackObj(network, object))
	fCompl := int32(0)
	for vec.Vec_WrdSize(super) > 0 {
		if vec.Vec_WrdSize(super) > 1 {
			for i := 1; i < vec.Vec_WrdSize(super); i++ {
				if vec.Vec_WrdEntry(super, i-1) >= vec.Vec_WrdEntry(super, i) {
					panic("XOR supergate order invariant")
				}
			}
		}
		var root *abc.Object
		found := false
		for i := vec.Vec_WrdSize(super) - 1; i >= 0; i-- {
			num := vec.Vec_WrdEntry(super, i)
			candidate := network.Object(int(Abc_NtkShareUnpackId(num)))
			if abc.Abc_NodeIsExorType(candidate) != 0 {
				vec.Vec_WrdRemove(super, num)
				root, found = candidate, true
				break
			}
		}
		if !found {
			break
		}
		var thenEdge, elseEdge abc.Edge
		control := abc.Abc_NodeRecognizeMux(root, &thenEdge, &elseEdge)
		if thenEdge != elseEdge.Not() {
			panic("XOR mux arms are not complements")
		}
		fCompl ^= abcShareBool(abc.Abc_ObjIsComplement(control) != 0)
		fCompl ^= abcShareBool(elseEdge.Complement)
		vec.Vec_WrdPushOrder(super, abcSharePackObj(network, control))
		vec.Vec_WrdPushOrder(super, abcSharePackObj(network, elseEdge.Object))
		(*counter)++
		compactSortedWords(super, false)
	}
	*polarity = fCompl
	for i := 0; i < vec.Vec_WrdSize(super); i++ {
		vec.Vec_WrdWriteEntry(super, i, uint64(Abc_NtkShareUnpackId(vec.Vec_WrdEntry(super, i))))
	}
	return super
}

// Abc_NtkShareSuperAnd expands a positive AND tree into literal leaves.
func Abc_NtkShareSuperAnd(object *Abc_Obj_t, counter *int32) *vec.Vec_Wrd_t {
	network := abc.Abc_ObjNtk(object)
	if object == nil || abc.Abc_ObjIsComplement(object) != 0 {
		panic("Abc_NtkShareSuperAnd requires a regular root")
	}
	super := vec.Vec_WrdAlloc(10)
	vec.Vec_WrdPush(super, Abc_NtkSharePack(abc.Abc_ObjLevel(object), abcShareObjLit(object, false)))
	for vec.Vec_WrdSize(super) > 0 {
		for i := 1; i < vec.Vec_WrdSize(super); i++ {
			if vec.Vec_WrdEntry(super, i-1) >= vec.Vec_WrdEntry(super, i) {
				panic("AND supergate order invariant")
			}
		}
		var root *abc.Object
		found := false
		for i := vec.Vec_WrdSize(super) - 1; i >= 0; i-- {
			num := vec.Vec_WrdEntry(super, i)
			edge := abcShareEdgeFromLit(network, int32(Abc_NtkShareUnpackId(num)))
			if !edge.Complement && abc.Abc_ObjIsNode(edge.Object) != 0 {
				vec.Vec_WrdRemove(super, num)
				root, found = edge.Object, true
				break
			}
		}
		if !found {
			break
		}
		children := root.Fanins()
		if len(children) != 2 {
			panic("AND node must have two fanins")
		}
		for _, child := range children {
			regular := child.Object
			if abc.Abc_ObjIsNode(regular) == 0 && abc.Abc_ObjIsCi(regular) == 0 {
				panic("AND leaf is neither a node nor CI")
			}
			vec.Vec_WrdPushOrder(super, Abc_NtkSharePack(abc.Abc_ObjLevel(regular), abcShareEdgeLit(child)))
		}
		(*counter)++
		if compactSortedWords(super, true) {
			vec.Vec_WrdClear(super)
			return super
		}
	}
	for i := 0; i < vec.Vec_WrdSize(super); i++ {
		vec.Vec_WrdWriteEntry(super, i, uint64(Abc_NtkShareUnpackId(vec.Vec_WrdEntry(super, i))))
	}
	return super
}

// compactSortedWords removes duplicate terms; for AND, x and !x make false.
func compactSortedWords(words *vec.Vec_Wrd_t, andMode bool) bool {
	input := vec.Vec_WrdArray(words)
	out := make([]uint64, 0, len(input))
	for i := 0; i < len(input); {
		j := i + 1
		for j < len(input) && input[j] == input[i] {
			j++
		}
		if andMode {
			if len(out) != 0 && out[len(out)-1]+1 == input[i] && input[i]&1 != 0 {
				return true
			}
			out = append(out, input[i])
		} else if (j-i)&1 != 0 {
			out = append(out, input[i])
		}
		i = j
	}
	vec.Vec_WrdShrink(words, len(out))
	for i, value := range out {
		vec.Vec_WrdWriteEntry(words, i, value)
	}
	return false
}

func Abc_NtkTraverseSupersXor_rec(manager *Abc_ShaMan_t, object *Abc_Obj_t, inputs *Vec_Ptr_t) {
	if object == nil || abc.Abc_NodeIsTravIdCurrent(object) != 0 {
		return
	}
	abc.Abc_NodeSetTravIdCurrent(object)
	if abc.Abc_ObjIsCi(object) != 0 {
		return
	}
	if abc.Abc_ObjIsNode(object) == 0 {
		panic("XOR traversal reached a non-node")
	}
	if abc.Abc_NodeIsExorType(object) != 0 {
		var polarity int32
		super := Abc_NtkShareSuperXor(object, &polarity, &manager.nFoundGates)
		if vec.Vec_WrdSize(super) <= 1 || int32(vec.Vec_WrdSize(super)) >= manager.nMultiSize {
			for i := 0; i < vec.Vec_WrdSize(super); i++ {
				id := int32(vec.Vec_WrdEntry(super, i))
				column := vec.Vec_PtrEntry(inputs, int(id))
				if column == nil {
					v := vec.Vec_IntAlloc(10)
					vec.Vec_IntPush(v, id<<1)
					vec.Vec_IntPush(v, abc.Abc_ObjLevel(manager.pNtk.Object(int(id))))
					vec.Vec_PtrWriteEntry(inputs, int(id), v)
					column = v
				}
				vec.Vec_IntPush(column.(*Vec_Int_t), int32(vec.Vec_IntSize(manager.vObj2Lit)))
			}
			vec.Vec_IntPush(manager.vObj2Lit, abcShareObjLit(object, polarity != 0))
		}
		for i := 0; i < vec.Vec_WrdSize(super); i++ {
			Abc_NtkTraverseSupersXor_rec(manager, manager.pNtk.Object(int(vec.Vec_WrdEntry(super, i))), inputs)
		}
		vec.Vec_WrdFree(super)
		return
	}
	for _, edge := range object.Fanins() {
		Abc_NtkTraverseSupersXor_rec(manager, edge.Object, inputs)
	}
}

func Abc_NtkTraverseSupersAnd_rec(manager *Abc_ShaMan_t, object *Abc_Obj_t, inputs *Vec_Ptr_t) {
	if object == nil || abc.Abc_NodeIsTravIdCurrent(object) != 0 {
		return
	}
	abc.Abc_NodeSetTravIdCurrent(object)
	if abc.Abc_ObjIsCi(object) != 0 {
		return
	}
	if abc.Abc_ObjIsNode(object) == 0 {
		panic("AND traversal reached a non-node")
	}
	super := Abc_NtkShareSuperAnd(object, &manager.nFoundGates)
	if vec.Vec_WrdSize(super) <= 1 || int32(vec.Vec_WrdSize(super)) >= manager.nMultiSize {
		for i := 0; i < vec.Vec_WrdSize(super); i++ {
			literal := int32(vec.Vec_WrdEntry(super, i))
			column := vec.Vec_PtrEntry(inputs, int(literal))
			if column == nil {
				v := vec.Vec_IntAlloc(10)
				vec.Vec_IntPush(v, literal)
				vec.Vec_IntPush(v, abc.Abc_ObjLevel(manager.pNtk.Object(int(literal>>1))))
				vec.Vec_PtrWriteEntry(inputs, int(literal), v)
				column = v
			}
			vec.Vec_IntPush(column.(*Vec_Int_t), int32(vec.Vec_IntSize(manager.vObj2Lit)))
		}
		vec.Vec_IntPush(manager.vObj2Lit, abcShareEdgeLit(object.Signal()))
	}
	for i := 0; i < vec.Vec_WrdSize(super); i++ {
		literal := int32(vec.Vec_WrdEntry(super, i))
		Abc_NtkTraverseSupersAnd_rec(manager, manager.pNtk.Object(int(literal>>1)), inputs)
	}
	vec.Vec_WrdFree(super)
}

func Abc_NtkTraverseSupers(manager *Abc_ShaMan_t, andMode int32) {
	inputCount := int(abc.Abc_NtkObjNumMax(manager.pNtk))
	if andMode != 0 {
		inputCount *= 2
	}
	inputs := vec.Vec_PtrStart(inputCount)
	abc.Abc_NtkIncrementTravId(manager.pNtk)
	for _, co := range manager.pNtk.Outputs() {
		if len(co.Fanins()) != 1 {
			continue
		}
		root := co.Fanins()[0].Object
		if abc.Abc_ObjIsNode(root) == 0 {
			continue
		}
		if andMode != 0 {
			Abc_NtkTraverseSupersAnd_rec(manager, root, inputs)
		} else {
			Abc_NtkTraverseSupersXor_rec(manager, root, inputs)
		}
	}
	manager.nStartCols = int32(vec.Vec_IntSize(manager.vObj2Lit))
	maxOnes := 0
	for i := 0; i < vec.Vec_PtrSize(inputs); i++ {
		if v, ok := vec.Vec_PtrEntry(inputs, i).(*Vec_Int_t); ok && v != nil && vec.Vec_IntSize(v)-SHARE_NUM > maxOnes {
			maxOnes = vec.Vec_IntSize(v) - SHARE_NUM
		}
	}
	manager.vBuckets = vec.Vec_PtrAlloc(maxOnes + 1)
	for i := 0; i <= maxOnes; i++ {
		vec.Vec_PtrPush(manager.vBuckets, vec.Vec_PtrAlloc(10))
	}
	for i := 0; i < vec.Vec_PtrSize(inputs); i++ {
		v, ok := vec.Vec_PtrEntry(inputs, i).(*Vec_Int_t)
		if !ok || v == nil {
			continue
		}
		bucket := vec.Vec_PtrEntry(manager.vBuckets, vec.Vec_IntSize(v)-SHARE_NUM).(*Vec_Ptr_t)
		vec.Vec_PtrPush(bucket, v)
	}
	vec.Vec_PtrFree(inputs)
}

func Abc_NtkSharePrint(manager *Abc_ShaMan_t) {
	ncols := vec.Vec_IntSize(manager.vObj2Lit)
	counters := make([]int, ncols+1)
	for b := 0; b < vec.Vec_PtrSize(manager.vBuckets); b++ {
		bucket := vec.Vec_PtrEntry(manager.vBuckets, b).(*Vec_Ptr_t)
		for j := 0; j < vec.Vec_PtrSize(bucket); j++ {
			input := vec.Vec_PtrEntry(bucket, j).(*Vec_Int_t)
			buffer := make([]byte, ncols)
			for i := range buffer {
				buffer[i] = '0'
			}
			for k := SHARE_NUM; k < vec.Vec_IntSize(input); k++ {
				id := vec.Vec_IntEntry(input, k)
				if id < 0 || int(id) >= ncols {
					panic("input column reference out of range")
				}
				buffer[id] = '1'
				counters[id]++
			}
			fmt.Printf("%4d%3d: %s\n", vec.Vec_IntEntry(input, 0), vec.Vec_IntEntry(input, 1), string(buffer))
		}
	}
	for i, count := range counters[:ncols] {
		if count > 0 {
			fmt.Printf("%d=%d ", i, count)
		}
	}
	fmt.Println()
	total := 0
	for i := 0; i < int(manager.nStartCols); i++ {
		total += counters[i] - 1
	}
	fmt.Printf("Total = %d.  Gates = %d.\n", total, ncols-int(manager.nStartCols)+total)
	fmt.Print("Bucket contents: ")
	for i := 0; i < vec.Vec_PtrSize(manager.vBuckets); i++ {
		fmt.Printf("%d ", vec.Vec_PtrSize(vec.Vec_PtrEntry(manager.vBuckets, i).(*Vec_Ptr_t)))
	}
	fmt.Println()
}

func Abc_NtkDumpBlif(network *Abc_Ntk_t) {
	file, err := os.OpenFile("multi_and.blif", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		fmt.Printf("Cannot open output file.\n")
		return
	}
	defer file.Close()
	fmt.Fprintf(file, ".model multi_and\n.inputs")
	for i := int32(0); i < abc.Abc_NtkCiNum(network); i++ {
		fmt.Fprintf(file, " i%d", i)
	}
	fmt.Fprintf(file, "\n.outputs")
	for i := int32(0); i < abc.Abc_NtkCoNum(network); i++ {
		fmt.Fprintf(file, " o%d", i)
	}
	fmt.Fprintln(file)
	for i, ci := range network.Inputs() {
		abc.Abc_ObjSetTemp(ci, int32(i))
	}
	for i, co := range network.Outputs() {
		support := abc.Abc_NtkNodeSupport(network, []*abc.Object{co}, 1)
		leaves := support[:0]
		for _, object := range support {
			if abc.Abc_ObjIsCi(object) != 0 {
				leaves = append(leaves, object)
			}
		}
		fmt.Fprint(file, ".names")
		for _, leaf := range leaves {
			fmt.Fprintf(file, " i%d", abc.Abc_ObjTemp(leaf))
		}
		fmt.Fprintf(file, " o%d\n", i)
		for range leaves {
			fmt.Fprint(file, "1")
		}
		fmt.Fprintln(file, " 1")
	}
	fmt.Fprint(file, ".end\n\n")
}

func Abc_NtkShareFindBestMatch(buckets *Vec_Ptr_t, inputBest, inputBest2 **Vec_Int_t) {
	const poolSize = 40
	pool := vec.Vec_PtrAlloc(poolSize)
	for i := vec.Vec_PtrSize(buckets) - 1; i >= 0 && vec.Vec_PtrSize(pool) < poolSize; i-- {
		bucket := vec.Vec_PtrEntry(buckets, i).(*Vec_Ptr_t)
		for k := 0; k < vec.Vec_PtrSize(bucket) && vec.Vec_PtrSize(pool) < poolSize; k++ {
			vec.Vec_PtrPush(pool, vec.Vec_PtrEntry(bucket, k))
		}
	}
	costBest, delayBest := 0, int32(0)
	var best, best2 *Vec_Int_t
	for i := vec.Vec_PtrSize(pool) - 1; i >= 0; i-- {
		v1 := vec.Vec_PtrEntry(pool, i).(*Vec_Int_t)
		for k := vec.Vec_PtrSize(pool) - 1; k >= 0; k-- {
			if i == k {
				continue
			}
			v2 := vec.Vec_PtrEntry(pool, k).(*Vec_Int_t)
			tail1 := vec.Vec_IntAllocArrayCopy(vec.Vec_IntArray(v1)[SHARE_NUM:], vec.Vec_IntSize(v1)-SHARE_NUM)
			tail2 := vec.Vec_IntAllocArrayCopy(vec.Vec_IntArray(v2)[SHARE_NUM:], vec.Vec_IntSize(v2)-SHARE_NUM)
			cost := vec.Vec_IntTwoCountCommon(tail1, tail2)
			vec.Vec_IntFree(tail1)
			vec.Vec_IntFree(tail2)
			if cost < 2 {
				continue
			}
			delay := max(vec.Vec_IntEntry(v1, 1), vec.Vec_IntEntry(v2, 1))
			if costBest < cost || costBest == cost && delayBest > delay {
				costBest, delayBest, best, best2 = cost, delay, v1, v2
			}
		}
	}
	vec.Vec_PtrFree(pool)
	*inputBest, *inputBest2 = best, best2
	if best == nil {
		return
	}
	firstBucket := vec.Vec_PtrEntry(buckets, vec.Vec_IntSize(best)-SHARE_NUM).(*Vec_Ptr_t)
	secondBucket := vec.Vec_PtrEntry(buckets, vec.Vec_IntSize(best2)-SHARE_NUM).(*Vec_Ptr_t)
	vec.Vec_PtrRemove(firstBucket, best)
	vec.Vec_PtrRemove(secondBucket, best2)
}

func Abc_NtkShareOptimize(manager *Abc_ShaMan_t, andMode int32) {
	aigManager, ok := manager.pNtk.FunctionManager.(*abc.Abc_Aig_t)
	if !ok {
		panic("shared extraction requires an AIG manager")
	}
	for {
		var input, input2 *Vec_Int_t
		Abc_NtkShareFindBestMatch(manager.vBuckets, &input, &input2)
		if input == nil {
			return
		}
		left := abcShareEdgeFromLit(manager.pNtk, vec.Vec_IntEntry(input, 0))
		right := abcShareEdgeFromLit(manager.pNtk, vec.Vec_IntEntry(input2, 0))
		var created abc.Edge
		if andMode != 0 {
			created = abc.Abc_AigAnd(aigManager, left, right)
		} else {
			created = abc.Abc_AigXor(aigManager, left, right)
		}
		manager.nCountGates++
		old1 := vec.Vec_IntAlloc(16)
		vec.Vec_IntPush(old1, vec.Vec_IntEntry(input, 0))
		vec.Vec_IntPush(old1, vec.Vec_IntEntry(input, 1))
		old2 := vec.Vec_IntAlloc(16)
		vec.Vec_IntPush(old2, vec.Vec_IntEntry(input2, 0))
		vec.Vec_IntPush(old2, vec.Vec_IntEntry(input2, 1))
		newCol := vec.Vec_IntAlloc(16)
		vec.Vec_IntPush(newCol, abcShareEdgeLit(created))
		vec.Vec_IntPush(newCol, abc.Abc_ObjLevel(created.Object))
		tail1 := vec.Vec_IntAllocArrayCopy(vec.Vec_IntArray(input)[SHARE_NUM:], vec.Vec_IntSize(input)-SHARE_NUM)
		tail2 := vec.Vec_IntAllocArrayCopy(vec.Vec_IntArray(input2)[SHARE_NUM:], vec.Vec_IntSize(input2)-SHARE_NUM)
		common := vec.Vec_IntAlloc(16)
		vec.Vec_IntTwoSplit(tail1, tail2, common, old1, old2)
		vec.Vec_IntFree(tail1)
		vec.Vec_IntFree(tail2)
		vec.Vec_IntFree(input)
		vec.Vec_IntFree(input2)
		for i := 0; i < vec.Vec_IntSize(common); i++ {
			vec.Vec_IntPush(newCol, vec.Vec_IntEntry(common, i))
		}
		vec.Vec_IntFree(common)
		vec.Vec_IntPush(old1, int32(vec.Vec_IntSize(manager.vObj2Lit)))
		vec.Vec_IntPush(old2, int32(vec.Vec_IntSize(manager.vObj2Lit)))
		vec.Vec_IntPush(manager.vObj2Lit, abcShareEdgeLit(created))
		for _, col := range []*Vec_Int_t{old1, old2, newCol} {
			bucket := vec.Vec_PtrEntry(manager.vBuckets, vec.Vec_IntSize(col)-SHARE_NUM).(*Vec_Ptr_t)
			vec.Vec_PtrPush(bucket, col)
		}
	}
}

func Abc_NtkUpdateNetwork(manager *Abc_ShaMan_t, andMode int32) *Abc_Ntk_t {
	network := manager.pNtk
	aigManager, ok := network.FunctionManager.(*abc.Abc_Aig_t)
	if !ok {
		panic("shared extraction requires an AIG manager")
	}
	original := make([]*abc.Object, manager.nStartCols)
	replacements := make([]abc.Edge, manager.nStartCols)
	for i := int32(0); i < manager.nStartCols; i++ {
		literal := vec.Vec_IntEntry(manager.vObj2Lit, int(i))
		if andMode != 0 && literal&1 != 0 {
			panic("AND supergate root is complemented")
		}
		original[i] = network.Object(int(literal >> 1))
		if andMode != 0 {
			replacements[i] = abc.Abc_AigConst1(network)
		} else {
			replacements[i] = abc.Abc_ObjNotCond(abc.Abc_AigConst1(network), 1-(literal&1))
		}
		manager.nCountGates--
	}
	for i := 0; i < vec.Vec_PtrSize(manager.vBuckets); i++ {
		bucket := vec.Vec_PtrEntry(manager.vBuckets, i).(*Vec_Ptr_t)
		for j := 0; j < vec.Vec_PtrSize(bucket); j++ {
			input := vec.Vec_PtrEntry(bucket, j).(*Vec_Int_t)
			for k := SHARE_NUM; k < vec.Vec_IntSize(input); k++ {
				column := vec.Vec_IntEntry(input, k)
				if column >= manager.nStartCols {
					break
				}
				if column < 0 {
					panic("negative input column")
				}
				leaf := abcShareEdgeFromLit(network, vec.Vec_IntEntry(input, 0))
				var combined abc.Edge
				if andMode != 0 {
					combined = abc.Abc_AigAnd(aigManager, replacements[column], leaf)
				} else {
					combined = abc.Abc_AigXor(aigManager, replacements[column], leaf)
				}
				replacements[column] = combined
				manager.nCountGates++
			}
		}
	}
	if manager.fVerbose != 0 {
		fmt.Printf("Total gates collected = %d.  Total gates constructed = %d.\n", manager.nFoundGates, manager.nCountGates)
	}
	replacementLit := make(map[int]int32, manager.nStartCols)
	for i, object := range original {
		replacementLit[object.ID] = abcShareEdgeLit(replacements[i])
	}
	for _, object := range network.Objects() {
		fanins := object.Fanins()
		for i, fanin := range fanins {
			if literal, found := replacementLit[fanin.Object.ID]; found {
				newEdge := abcShareEdgeFromLit(network, literal)
				newEdge.Complement = newEdge.Complement != fanin.Complement
				if err := network.PatchFanin(object, i, newEdge); err != nil {
					panic(err)
				}
			}
		}
	}
	if andMode != 0 {
		return Abc_NtkBalance(network, 0, 0, 1)
	}
	return Abc_NtkBalanceExor(network, 1, 0)
}

func Abc_NtkShareXor(network *Abc_Ntk_t, nMultiSize, andMode, verbose int32) *Abc_Ntk_t {
	if abc.Abc_NtkIsStrash(network) == 0 {
		panic("Abc_NtkShareXor requires a strashed network")
	}
	manager := Abc_ShaManStart(network)
	defer Abc_ShaManStop(manager)
	manager.nMultiSize, manager.fVerbose = nMultiSize, verbose
	Abc_NtkTraverseSupers(manager, andMode)
	if manager.nStartCols < 2 {
		return abc.Abc_NtkDup(network)
	}
	if verbose != 0 {
		Abc_NtkSharePrint(manager)
	}
	Abc_NtkShareOptimize(manager, andMode)
	if verbose != 0 {
		Abc_NtkSharePrint(manager)
	}
	return Abc_NtkUpdateNetwork(manager, andMode)
}
