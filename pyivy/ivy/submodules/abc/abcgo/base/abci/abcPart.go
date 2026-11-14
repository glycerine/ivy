package abci

import (
	"fmt"
	"math"
	"os"
	"sort"
	"time"

	"github.com/glycerine/abcgo/aig/hop"
	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/base/cmd"
	abcmain "github.com/glycerine/abcgo/base/main"
	"github.com/glycerine/abcgo/misc/extra"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

type Supp_Man_t struct {
	nChunkSize, nStepSize int32
	pFreeBuf              []byte
	nFreeSize             int32
	vMemory               *Vec_Ptr_t
	vFree                 *Vec_Ptr_t
	freeEntries           map[int][]*Supp_One_t
}

type Supp_One_t struct {
	nRefs, nOuts, nOutsAlloc int32
	pOuts                    []int32
	next                     *Supp_One_t
}

func Supp_SizeType(nSize, nStepSize int32) int32 {
	if nSize <= 0 || nStepSize <= 0 {
		return 0
	}
	return nSize/nStepSize + boolInt32(nSize%nStepSize != 0)
}
func Supp_OneNext(part *Supp_One_t) *Supp_One_t { return part.next }
func Supp_OneSetNext(part, next *Supp_One_t)    { part.next = next }

func Supp_ManStart(chunkSize, stepSize int32) *Supp_Man_t {
	if chunkSize <= 0 || stepSize <= 0 {
		panic("support allocator requires positive chunk and step sizes")
	}
	return &Supp_Man_t{nChunkSize: chunkSize, nStepSize: stepSize, vMemory: vec.Vec_PtrAlloc(1000), vFree: vec.Vec_PtrAlloc(1000), freeEntries: make(map[int][]*Supp_One_t)}
}
func Supp_ManStop(p *Supp_Man_t) {
	if p == nil {
		return
	}
	vec.Vec_PtrFree(p.vMemory)
	vec.Vec_PtrFree(p.vFree)
	clear(p.freeEntries)
	p.pFreeBuf = nil
	p.nFreeSize = 0
}
func Supp_ManFetch(p *Supp_Man_t, size int32) []byte {
	if size <= 0 {
		panic("support allocator fetch requires positive size")
	}
	typ := Supp_SizeType(size, p.nStepSize)
	vec.Vec_PtrFillExtra(p.vFree, int(typ+1), nil)
	if pool, ok := vec.Vec_PtrEntry(p.vFree, int(typ)).([][]byte); ok && len(pool) != 0 {
		block := pool[len(pool)-1]
		vec.Vec_PtrWriteEntry(p.vFree, int(typ), pool[:len(pool)-1])
		return block[:size]
	}
	realSize := typ * p.nStepSize
	if p.nFreeSize < realSize {
		chunkSize := p.nChunkSize
		if chunkSize < realSize {
			chunkSize = realSize
		}
		chunk := make([]byte, chunkSize)
		vec.Vec_PtrPush(p.vMemory, chunk)
		p.pFreeBuf, p.nFreeSize = chunk, chunkSize
	}
	out := p.pFreeBuf[:realSize]
	p.pFreeBuf = p.pFreeBuf[realSize:]
	p.nFreeSize -= realSize
	return out[:size]
}
func Supp_ManRecycle(p *Supp_Man_t, memory []byte, size int32) {
	typ := Supp_SizeType(size, p.nStepSize)
	vec.Vec_PtrFillExtra(p.vFree, int(typ+1), nil)
	if int32(len(memory)) < size {
		panic("support recycle buffer is too short")
	}
	pool, _ := vec.Vec_PtrEntry(p.vFree, int(typ)).([][]byte)
	vec.Vec_PtrWriteEntry(p.vFree, int(typ), append(pool, memory[:typ*p.nStepSize]))
}
func Supp_ManFetchEntry(p *Supp_Man_t, words, refs int32) *Supp_One_t {
	if words < 0 || refs < 0 {
		panic("negative support-entry size or reference count")
	}
	capacity := int(words)
	pool := p.freeEntries[capacity]
	var entry *Supp_One_t
	if len(pool) != 0 {
		entry = pool[len(pool)-1]
		p.freeEntries[capacity] = pool[:len(pool)-1]
		entry.nOuts = 0
		entry.nRefs = refs
		entry.next = nil
	} else {
		entry = &Supp_One_t{nRefs: refs, nOutsAlloc: words, pOuts: make([]int32, 0, capacity)}
	}
	return entry
}
func Supp_ManRecycleEntry(p *Supp_Man_t, entry *Supp_One_t) {
	if entry == nil || entry.nOuts > entry.nOutsAlloc || entry.nOuts < entry.nOutsAlloc/2 {
		panic("invalid support entry recycle")
	}
	entry.pOuts = entry.pOuts[:0]
	p.freeEntries[int(entry.nOutsAlloc)] = append(p.freeEntries[int(entry.nOutsAlloc)], entry)
}
func Supp_ManMergeEntry(manager *Supp_Man_t, left, right *Supp_One_t, refs int32) *Supp_One_t {
	if left == nil || right == nil {
		panic("nil support entry")
	}
	merged := make([]int32, 0, int(left.nOuts+right.nOuts))
	i, j := 0, 0
	for i < int(left.nOuts) && j < int(right.nOuts) {
		a, b := left.pOuts[i], right.pOuts[j]
		switch {
		case a == b:
			merged = append(merged, a)
			i++
			j++
		case a < b:
			merged = append(merged, a)
			i++
		default:
			merged = append(merged, b)
			j++
		}
	}
	merged = append(merged, left.pOuts[i:int(left.nOuts)]...)
	merged = append(merged, right.pOuts[j:int(right.nOuts)]...)
	entry := Supp_ManFetchEntry(manager, int32(len(merged)), refs)
	entry.pOuts = append(entry.pOuts, merged...)
	entry.nOuts = int32(len(merged))
	if entry.nOuts < left.nOuts || entry.nOuts < right.nOuts {
		panic("support merge lost entries")
	}
	return entry
}
func Supp_ManTransferEntry(entry *Supp_One_t) *Vec_Int_t {
	out := vec.Vec_IntAlloc(int(entry.nOuts))
	for _, index := range entry.pOuts[:entry.nOuts] {
		vec.Vec_IntPush(out, index)
	}
	return out
}

func Abc_NtkDfsNatural(network *Abc_Ntk_t) *Vec_Ptr_t {
	if abc.Abc_NtkIsStrash(network) == 0 {
		panic("support DFS requires a strashed network")
	}
	out := vec.Vec_PtrAlloc(int(abc.Abc_NtkObjNum(network)))
	seen := make(map[*Abc_Obj_t]bool)
	constant := abc.Abc_AigConst1(network).Object
	seen[constant] = true
	vec.Vec_PtrPush(out, constant)
	for _, node := range network.Objects() {
		if abc.Abc_ObjIsNode(node) == 0 {
			continue
		}
		for _, fanin := range node.Fanins() {
			if abc.Abc_ObjIsCi(fanin.Object) != 0 && !seen[fanin.Object] {
				seen[fanin.Object] = true
				vec.Vec_PtrPush(out, fanin.Object)
			}
		}
		vec.Vec_PtrPush(out, node)
		for _, fanout := range node.Fanouts() {
			if abc.Abc_ObjIsCo(fanout) != 0 && !seen[fanout] {
				seen[fanout] = true
				vec.Vec_PtrPush(out, fanout)
			}
		}
	}
	return out
}

func Abc_NtkComputeSupportsSmart(network *Abc_Ntk_t) *Vec_Ptr_t {
	ciIndex := make(map[*Abc_Obj_t]int32)
	for i := int32(0); i < abc.Abc_NtkCiNum(network); i++ {
		ciIndex[abc.Abc_NtkCi(network, i)] = i
	}
	coIndex := make(map[*Abc_Obj_t]int32)
	for i := int32(0); i < abc.Abc_NtkCoNum(network); i++ {
		coIndex[abc.Abc_NtkCo(network, i)] = i
	}
	supports := vec.Vec_PtrAlloc(int(abc.Abc_NtkCoNum(network)))
	manager := Supp_ManStart(1<<20, 1<<6)
	defer Supp_ManStop(manager)
	entries := make(map[*Abc_Obj_t]*Supp_One_t)
	order := Abc_NtkDfsNatural(network)
	for _, object := range vec.Vec_PtrArray(order) {
		node := object.(*Abc_Obj_t)
		switch {
		case abc.Abc_ObjIsNode(node) != 0:
			fanins := node.Fanins()
			left, right := entries[fanins[0].Object], entries[fanins[1].Object]
			entries[node] = Supp_ManMergeEntry(manager, left, right, abc.Abc_ObjFanoutNum(node))
			for _, entry := range []*Supp_One_t{left, right} {
				if entry == nil || entry.nRefs <= 0 {
					panic("support entry reference underflow")
				}
				entry.nRefs--
				if entry.nRefs == 0 {
					Supp_ManRecycleEntry(manager, entry)
				}
			}
		case abc.Abc_ObjIsCo(node) != 0:
			fanins := node.Fanins()
			entry := entries[fanins[0].Object]
			if abc.Abc_ObjIsNode(fanins[0].Object) != 0 {
				support := Supp_ManTransferEntry(entry)
				vec.Vec_IntPush(support, coIndex[node])
				vec.Vec_PtrPush(supports, support)
			}
			if entry != nil {
				if entry.nRefs <= 0 {
					panic("support entry reference underflow")
				}
				entry.nRefs--
				if entry.nRefs == 0 {
					Supp_ManRecycleEntry(manager, entry)
				}
			}
		case abc.Abc_ObjIsCi(node) != 0:
			if fanouts := abc.Abc_ObjFanoutNum(node); fanouts != 0 {
				entry := Supp_ManFetchEntry(manager, 1, fanouts)
				entry.pOuts = append(entry.pOuts, ciIndex[node])
				entry.nOuts = 1
				entries[node] = entry
			}
		case node == abc.Abc_AigConst1(network).Object:
			if fanouts := abc.Abc_ObjFanoutNum(node); fanouts != 0 {
				entries[node] = Supp_ManFetchEntry(manager, 0, fanouts)
			}
		default:
			panic("unexpected object in natural support order")
		}
	}
	vec.Vec_PtrFree(order)
	sortSupportVectors(supports)
	return supports
}

func sortSupportVectors(vectors *Vec_Ptr_t) {
	sort.SliceStable(vec.Vec_PtrArray(vectors), func(i, j int) bool {
		a := vec.Vec_PtrEntry(vectors, i).(*Vec_Int_t)
		b := vec.Vec_PtrEntry(vectors, j).(*Vec_Int_t)
		return vec.Vec_IntSize(a) > vec.Vec_IntSize(b)
	})
}
func freeIntVectorList(vectors *Vec_Ptr_t) {
	for _, item := range vec.Vec_PtrArray(vectors) {
		if v, ok := item.(*Vec_Int_t); ok && v != nil {
			vec.Vec_IntFree(v)
		}
	}
	vec.Vec_PtrFree(vectors)
}
func Abc_NtkComputeSupportsNaive(network *Abc_Ntk_t) *Vec_Ptr_t {
	ciIndex := make(map[*Abc_Obj_t]int32)
	for i := int32(0); i < abc.Abc_NtkCiNum(network); i++ {
		ciIndex[abc.Abc_NtkCi(network, i)] = i
	}
	out := vec.Vec_PtrAlloc(int(abc.Abc_NtkCoNum(network)))
	for i := int32(0); i < abc.Abc_NtkCoNum(network); i++ {
		co := abc.Abc_NtkCo(network, i)
		root := abc.Abc_ObjFanin0(co)
		if abc.Abc_ObjIsNode(root) == 0 {
			continue
		}
		supportObjects := abc.Abc_NtkNodeSupport(network, []*Abc_Obj_t{co}, 1)
		support := vec.Vec_IntAlloc(len(supportObjects) + 1)
		for _, input := range supportObjects {
			if index, ok := ciIndex[input]; ok {
				vec.Vec_IntPush(support, index)
			}
		}
		vec.Vec_IntSort(support, 0)
		vec.Vec_IntPush(support, i)
		vec.Vec_PtrPush(out, support)
	}
	sortSupportVectors(out)
	return out
}

func Abc_NtkSuppCharStart(support *Vec_Int_t, nPis int32) []uint32 {
	buffer := make([]uint32, int(util.Abc_BitWordNum(nPis)))
	for i := 0; i < vec.Vec_IntSize(support); i++ {
		entry := vec.Vec_IntEntry(support, i)
		if entry < 0 || entry >= nPis {
			panic("support input index out of range")
		}
		util.Abc_InfoSetBit(buffer, entry)
	}
	return buffer
}
func Abc_NtkSuppCharAdd(buffer []uint32, support *Vec_Int_t, nPis int32) {
	for i := 0; i < vec.Vec_IntSize(support); i++ {
		entry := vec.Vec_IntEntry(support, i)
		if entry < 0 || entry >= nPis {
			panic("support input index out of range")
		}
		util.Abc_InfoSetBit(buffer, entry)
	}
}
func Abc_NtkSuppCharCommon(buffer []uint32, support *Vec_Int_t) int32 {
	var common int32
	for i := 0; i < vec.Vec_IntSize(support); i++ {
		if util.Abc_InfoHasBit(buffer, vec.Vec_IntEntry(support, i)) {
			common++
		}
	}
	return common
}
func Abc_NtkPartitionSmartFindPart(allSupports, allParts, charSupports *Vec_Ptr_t, sizeLimit int32, support *Vec_Int_t) int32 {
	var best int32 = -1
	bestValue := int32(0)
	for i := 0; i < vec.Vec_PtrSize(allSupports); i++ {
		partSupport := vec.Vec_PtrEntry(allSupports, i).(*Vec_Int_t)
		common := Abc_NtkSuppCharCommon(vec.Vec_PtrEntry(charSupports, i).([]uint32), support)
		if common == 0 {
			continue
		}
		if common == int32(vec.Vec_IntSize(support)) {
			return int32(i)
		}
		if sizeLimit > 0 && int32(vec.Vec_IntSize(partSupport)) >= 2*sizeLimit {
			continue
		}
		attract := 1000 * common / int32(vec.Vec_IntSize(support))
		repulse := int32(1)
		if vec.Vec_IntSize(partSupport) >= 100 {
			repulse += int32(math.Log2(float64(vec.Vec_IntSize(partSupport) - 100)))
		}
		value := attract / repulse
		if value > bestValue {
			best, bestValue = int32(i), value
		}
	}
	if bestValue < 75 {
		return -1
	}
	return best
}
func Abc_NtkPartitionPrint(network *Abc_Ntk_t, parts, allSupports *Vec_Ptr_t) {
	var total int32
	for i := 0; i < vec.Vec_PtrSize(allSupports); i++ {
		support := vec.Vec_PtrEntry(allSupports, i).(*Vec_Int_t)
		part := vec.Vec_PtrEntry(parts, i).(*Vec_Int_t)
		fmt.Printf("%d=(%d,%d) ", i, vec.Vec_IntSize(support), vec.Vec_IntSize(part))
		total += int32(vec.Vec_IntSize(part))
	}
	fmt.Printf("\nTotal = %d. Outputs = %d.\n", total, abc.Abc_NtkCoNum(network))
}
func Abc_NtkPartitionCompact(parts, allSupports *Vec_Ptr_t, sizeLimit int32) {
	if sizeLimit == 0 {
		sizeLimit = 200
	}
	write := 0
	var pendingPart, pendingSupport *Vec_Int_t
	for read := 0; read < vec.Vec_PtrSize(allSupports); read++ {
		support := vec.Vec_PtrEntry(allSupports, read).(*Vec_Int_t)
		part := vec.Vec_PtrEntry(parts, read).(*Vec_Int_t)
		if int32(vec.Vec_IntSize(support)) < sizeLimit {
			if pendingSupport == nil {
				pendingSupport = vec.Vec_IntDup(support)
				pendingPart = part
			} else {
				mergedSupport := vec.Vec_IntTwoMerge(pendingSupport, support)
				vec.Vec_IntFree(pendingSupport)
				pendingSupport = mergedSupport
				mergedPart := vec.Vec_IntTwoMerge(pendingPart, part)
				vec.Vec_IntFree(pendingPart)
				vec.Vec_IntFree(part)
				pendingPart = mergedPart
			}
			if int32(vec.Vec_IntSize(pendingSupport)) < sizeLimit {
				continue
			}
		} else {
			pendingPart = part
		}
		vec.Vec_PtrWriteEntry(parts, write, pendingPart)
		if pendingSupport != nil {
			vec.Vec_IntFree(vec.Vec_PtrEntry(allSupports, write).(*Vec_Int_t))
			vec.Vec_PtrWriteEntry(allSupports, write, pendingSupport)
			pendingSupport = nil
		}
		pendingPart = nil
		write++
	}
	if pendingPart != nil {
		vec.Vec_PtrWriteEntry(parts, write, pendingPart)
		if pendingSupport == nil {
			panic("final compacted support missing")
		}
		vec.Vec_IntFree(vec.Vec_PtrEntry(allSupports, write).(*Vec_Int_t))
		vec.Vec_PtrWriteEntry(allSupports, write, pendingSupport)
		write++
	}
	vec.Vec_PtrShrink(parts, write)
	vec.Vec_PtrShrink(allSupports, write)
}
func Abc_NtkPartitionSmart(network *Abc_Ntk_t, sizeLimit, verbose int32) *Vec_Ptr_t {
	started := time.Now()
	supports := Abc_NtkComputeSupportsSmart(network)
	if verbose != 0 {
		fmt.Printf("Supps = %9.2f sec\n", time.Since(started).Seconds())
	}
	parts := vec.Vec_PtrAlloc(256)
	partSupports := vec.Vec_PtrAlloc(256)
	charSupports := vec.Vec_PtrAlloc(256)
	progress := extra.Extra_ProgressBarStart(os.Stdout, int32(vec.Vec_PtrSize(supports)))
	for i := 0; i < vec.Vec_PtrSize(supports); i++ {
		support := vec.Vec_PtrEntry(supports, i).(*Vec_Int_t)
		extra.Extra_ProgressBarUpdate(progress, int32(i), "")
		output := vec.Vec_IntPop(support)
		partIndex := Abc_NtkPartitionSmartFindPart(partSupports, parts, charSupports, sizeLimit, support)
		if partIndex < 0 {
			part := vec.Vec_IntAlloc(32)
			vec.Vec_IntPush(part, output)
			union := vec.Vec_IntDup(support)
			vec.Vec_PtrPush(parts, part)
			vec.Vec_PtrPush(partSupports, union)
			vec.Vec_PtrPush(charSupports, Abc_NtkSuppCharStart(support, abc.Abc_NtkCiNum(network)))
			continue
		}
		part := vec.Vec_PtrEntry(parts, int(partIndex)).(*Vec_Int_t)
		vec.Vec_IntPush(part, output)
		oldSupport := vec.Vec_PtrEntry(partSupports, int(partIndex)).(*Vec_Int_t)
		union := vec.Vec_IntTwoMerge(oldSupport, support)
		vec.Vec_IntFree(oldSupport)
		vec.Vec_PtrWriteEntry(partSupports, int(partIndex), union)
		Abc_NtkSuppCharAdd(vec.Vec_PtrEntry(charSupports, int(partIndex)).([]uint32), support, abc.Abc_NtkCiNum(network))
	}
	extra.Extra_ProgressBarStop(progress)
	vec.Vec_PtrFree(charSupports)
	if verbose != 0 {
		fmt.Printf("Parts = %9.2f sec\n", time.Since(started).Seconds())
	}
	// Attach each old index to its support. Sorting supports then sorts the
	// corresponding output partitions without a second sort structure.
	for i := 0; i < vec.Vec_PtrSize(partSupports); i++ {
		vec.Vec_IntPush(vec.Vec_PtrEntry(partSupports, i).(*Vec_Int_t), int32(i))
	}
	sort.SliceStable(vec.Vec_PtrArray(partSupports), func(i, j int) bool {
		a := vec.Vec_PtrEntry(partSupports, i).(*Vec_Int_t)
		b := vec.Vec_PtrEntry(partSupports, j).(*Vec_Int_t)
		return vec.Vec_IntSize(a) > vec.Vec_IntSize(b)
	})
	sortedParts := vec.Vec_PtrAlloc(vec.Vec_PtrSize(parts))
	for i := 0; i < vec.Vec_PtrSize(partSupports); i++ {
		support := vec.Vec_PtrEntry(partSupports, i).(*Vec_Int_t)
		originalIndex := vec.Vec_IntPop(support)
		vec.Vec_PtrPush(sortedParts, vec.Vec_PtrEntry(parts, int(originalIndex)))
	}
	vec.Vec_PtrFree(parts)
	Abc_NtkPartitionCompact(sortedParts, partSupports, sizeLimit)
	if verbose != 0 {
		fmt.Printf("Comps = %9.2f sec\nCreated %d partitions.\n", time.Since(started).Seconds(), vec.Vec_PtrSize(sortedParts))
	}
	freeIntVectorList(supports)
	freeIntVectorList(partSupports)
	return sortedParts
}
func Abc_NtkPartitionNaive(network *Abc_Ntk_t, partSize int32) *Vec_Ptr_t {
	if partSize <= 0 {
		panic("partition size must be positive")
	}
	count := (abc.Abc_NtkCoNum(network) + partSize - 1) / partSize
	parts := vec.Vec_PtrAlloc(int(count))
	for i := int32(0); i < count; i++ {
		vec.Vec_PtrPush(parts, vec.Vec_IntAlloc(int(partSize)))
	}
	for i := int32(0); i < abc.Abc_NtkCoNum(network); i++ {
		part := vec.Vec_PtrEntry(parts, int(i/partSize)).(*Vec_Int_t)
		vec.Vec_IntPush(part, i)
	}
	return parts
}
func Abc_NtkConvertCos(network *Abc_Ntk_t, outputs *Vec_Int_t, outputObjects *Vec_Ptr_t) {
	vec.Vec_PtrClear(outputObjects)
	for i := 0; i < vec.Vec_IntSize(outputs); i++ {
		index := vec.Vec_IntEntry(outputs, i)
		vec.Vec_PtrPush(outputObjects, abc.Abc_NtkCo(network, index))
	}
}

func Abc_NtkPartStitchFindRepr_rec(equivalents *Vec_Ptr_t, object *Abc_Obj_t) *Abc_Obj_t {
	seen := map[*Abc_Obj_t]bool{}
	for {
		if seen[object] {
			panic("cycle in partition representative map")
		}
		seen[object] = true
		repr, _ := vec.Vec_PtrEntry(equivalents, int(object.ID)).(*Abc_Obj_t)
		if repr == nil || repr == object {
			return object
		}
		object = repr
	}
}
func Abc_NtkPartStitchCopy0(equivalents *Vec_Ptr_t, object *Abc_Obj_t) abc.Edge {
	fanin := object.Fanins()[0]
	repr := Abc_NtkPartStitchFindRepr_rec(equivalents, fanin.Object)
	return abc.Abc_ObjNotCond(repr.CopyEdge, boolInt32(repr.Phase != fanin.Object.Phase)^abc.Abc_ObjFaninC1(object))
}
func Abc_NtkPartStitchCopy1(equivalents *Vec_Ptr_t, object *Abc_Obj_t) abc.Edge {
	fanin := object.Fanins()[1]
	repr := Abc_NtkPartStitchFindRepr_rec(equivalents, fanin.Object)
	return abc.Abc_ObjNotCond(repr.CopyEdge, boolInt32(repr.Phase != fanin.Object.Phase)^abc.Abc_ObjFaninC1(object))
}
func Hop_ObjChild0Next(object *Abc_Obj_t, mapping map[*Abc_Obj_t]*Hop_Obj_t) *Hop_Obj_t {
	fanin := object.Fanins()[0]
	return hop.Hop_NotCond(mapping[fanin.Object], boolInt32(fanin.Complement))
}
func Hop_ObjChild1Next(object *Abc_Obj_t, mapping map[*Abc_Obj_t]*Hop_Obj_t) *Hop_Obj_t {
	fanin := object.Fanins()[1]
	return hop.Hop_NotCond(mapping[fanin.Object], boolInt32(fanin.Complement))
}
func Abc_NtkPartStartHop(network *Abc_Ntk_t) *Hop_Man_t {
	manager := hop.Hop_ManStartWithObjects(int(abc.Abc_NtkObjNum(network)) + 1)
	mapping := map[*Abc_Obj_t]*Hop_Obj_t{abc.Abc_AigConst1(network).Object: hop.Hop_ManConst1(manager)}
	for _, ci := range network.Inputs() {
		mapping[ci] = hop.Hop_ObjCreatePi(manager)
	}
	for _, node := range abc.Abc_NtkNodes(network) {
		if abc.Abc_AigNodeIsAnd(node) == 0 {
			continue
		}
		mapping[node] = hop.Hop_And(manager, Hop_ObjChild0Next(node, mapping), Hop_ObjChild1Next(node, mapping))
	}
	for _, node := range abc.Abc_NtkNodes(network) {
		if choice := abc.Abc_ObjChoice(node); choice != nil && mapping[node] != nil && mapping[choice] != nil {
			hop.Hop_ObjSetData(hop.Hop_Regular(mapping[node]), hop.Hop_Regular(mapping[choice]))
		}
	}
	for _, co := range network.Outputs() {
		hop.Hop_ObjCreatePo(manager, Hop_ObjChild0Next(co, mapping))
	}
	if hop.Hop_ManCheck(manager) == 0 {
		fmt.Println("Abc_NtkPartStartHop: HOP manager check has failed.")
	}
	return manager
}

func Abc_NtkPartStitchChoices(network *Abc_Ntk_t, parts *Vec_Ptr_t) *Abc_Ntk_t {
	if abc.Abc_NtkIsStrash(network) == 0 {
		panic("partition stitching requires a strashed network")
	}
	stitched := abc.Abc_NtkStartFrom(network, abc.ABC_NTK_STRASH, abc.ABC_FUNC_AIG)
	aig, ok := stitched.FunctionManager.(*abc.Abc_Aig_t)
	if !ok {
		panic("stitched strashed network has no AIG manager")
	}
	for partIndex := 0; partIndex < vec.Vec_PtrSize(parts); partIndex++ {
		partition := vec.Vec_PtrEntry(parts, partIndex).(*Abc_Ntk_t)
		if abc.Abc_NtkIsStrash(partition) == 0 {
			panic("partition is not strashed")
		}
		abc.Abc_NtkCleanCopy(partition)
		abc.Abc_AigConst1(partition).Object.CopyEdge = abc.Abc_AigConst1(stitched)
		for _, ci := range partition.Inputs() {
			var target *Abc_Obj_t
			for _, candidate := range stitched.Inputs() {
				if candidate.Name() == ci.Name() {
					target = candidate
					break
				}
			}
			if target == nil {
				fmt.Printf("Cannot find CI node %s in the original network.\n", ci.Name())
				abc.Abc_NtkDelete(stitched)
				return nil
			}
			ci.CopyEdge = target.Signal()
		}
		for _, node := range abc.Abc_AigDfs(partition, 1, 0) {
			fanins := node.Fanins()
			a := abc.Abc_ObjNotCond(fanins[0].Object.CopyEdge, boolInt32(fanins[0].Complement))
			b := abc.Abc_ObjNotCond(fanins[1].Object.CopyEdge, boolInt32(fanins[1].Complement))
			node.CopyEdge = abc.Abc_AigAnd(aig, a, b)
			if node.Choice != nil {
				for choice := node.Choice; choice != nil; choice = choice.Choice {
					choice.CopyEdge = node.CopyEdge
				}
			}
		}
		for _, co := range partition.Outputs() {
			var target *Abc_Obj_t
			for _, candidate := range stitched.Outputs() {
				if candidate.Name() == co.Name() {
					target = candidate
					break
				}
			}
			if target == nil {
				fmt.Printf("Cannot find CO node %s in the original network.\n", co.Name())
				abc.Abc_NtkDelete(stitched)
				return nil
			}
			co.CopyEdge = target.Signal()
			fanin := co.Fanins()[0]
			edge := abc.Abc_ObjNotCond(fanin.Object.CopyEdge, boolInt32(fanin.Complement))
			if err := target.Network().AddFanin(target, edge); err != nil {
				panic(err)
			}
		}
	}
	for _, co := range network.Outputs() {
		target := abc.Abc_ObjRegular(co.CopyEdge.Object)
		if target != nil && len(target.Fanins()) == 0 {
			fanin := co.Fanins()[0]
			edge := abc.Abc_ObjNotCond(fanin.Object.CopyEdge, boolInt32(fanin.Complement))
			if err := target.Network().AddFanin(target, edge); err != nil {
				panic(err)
			}
		}
	}
	history := Abc_NtkPartStartHop(stitched)
	result := Abc_NtkHopRemoveLoops(stitched, history)
	abc.Abc_NtkDelete(stitched)
	if result == nil || abc.Abc_NtkCheck(result) == 0 {
		fmt.Println("Abc_NtkPartStitchChoices: The network check has failed.")
		if result != nil {
			abc.Abc_NtkDelete(result)
		}
		return nil
	}
	return result
}

func Abc_NtkFraigPartitioned(store *Vec_Ptr_t, params any) *Abc_Ntk_t {
	if vec.Vec_PtrSize(store) == 0 {
		return nil
	}
	original, ok := vec.Vec_PtrEntry(store, 0).(*Abc_Ntk_t)
	if !ok || abc.Abc_NtkIsStrash(original) == 0 {
		panic("fraig partition store must begin with a strashed network")
	}
	parts := Abc_NtkPartitionSmart(original, 300, 0)
	cmd.Cmd_CommandExecute(abcmain.Abc_FrameGetGlobalFrame(), "unset progressbar")
	roots := vec.Vec_PtrAlloc(1000)
	fraigs := vec.Vec_PtrAlloc(vec.Vec_PtrSize(parts))
	for i := 0; i < vec.Vec_PtrSize(parts); i++ {
		outputs := vec.Vec_PtrEntry(parts, i).(*Vec_Int_t)
		Abc_NtkConvertCos(original, outputs, roots)
		partition := abc.Abc_NtkCreateConeArray(original, objectSlice(roots), 0)
		for j := 1; j < vec.Vec_PtrSize(store); j++ {
			network := vec.Vec_PtrEntry(store, j).(*Abc_Ntk_t)
			Abc_NtkConvertCos(network, outputs, roots)
			abc.Abc_NtkAppendToCone(partition, network, objectSlice(roots))
		}
		fmt.Printf("Fraiging part %4d  (out of %4d)  PI = %5d. PO = %5d. And = %6d. Lev = %4d.\r", i+1, vec.Vec_PtrSize(parts), abc.Abc_NtkPiNum(partition), abc.Abc_NtkPoNum(partition), abc.Abc_NtkNodeNum(partition), abc.Abc_AigLevel(partition))
		fraigged := Abc_NtkFraig(partition, params, 1, 0)
		vec.Vec_PtrPush(fraigs, fraigged)
		abc.Abc_NtkDelete(partition)
	}
	fmt.Print("                                                                                          \r")
	freeIntVectorList(parts)
	cmd.Cmd_CommandExecute(abcmain.Abc_FrameGetGlobalFrame(), "set progressbar")
	result := Abc_NtkPartStitchChoices(original, fraigs)
	for _, item := range vec.Vec_PtrArray(fraigs) {
		if network, ok := item.(*Abc_Ntk_t); ok && network != nil {
			abc.Abc_NtkDelete(network)
		}
	}
	vec.Vec_PtrFree(fraigs)
	vec.Vec_PtrFree(roots)
	return result
}
func Abc_NtkFraigPartitionedTime(network *Abc_Ntk_t, params any) {
	started := time.Now()
	if abc.Abc_NtkIsStrash(network) == 0 {
		panic("partitioned fraiging requires a strashed network")
	}
	parts := Abc_NtkPartitionSmart(network, 300, 0)
	cmd.Cmd_CommandExecute(abcmain.Abc_FrameGetGlobalFrame(), "unset progressbar")
	roots := vec.Vec_PtrAlloc(1000)
	fraigs := vec.Vec_PtrAlloc(vec.Vec_PtrSize(parts))
	for i := 0; i < vec.Vec_PtrSize(parts); i++ {
		outputs := vec.Vec_PtrEntry(parts, i).(*Vec_Int_t)
		Abc_NtkConvertCos(network, outputs, roots)
		partition := abc.Abc_NtkCreateConeArray(network, objectSlice(roots), 0)
		fraigged := Abc_NtkFraig(partition, params, 0, 0)
		vec.Vec_PtrPush(fraigs, fraigged)
		abc.Abc_NtkDelete(partition)
		fmt.Printf("Finished part %5d (out of %5d)\r", i+1, vec.Vec_PtrSize(parts))
	}
	freeIntVectorList(parts)
	cmd.Cmd_CommandExecute(abcmain.Abc_FrameGetGlobalFrame(), "set progressbar")
	for _, item := range vec.Vec_PtrArray(fraigs) {
		if n, ok := item.(*Abc_Ntk_t); ok && n != nil {
			abc.Abc_NtkDelete(n)
		}
	}
	vec.Vec_PtrFree(fraigs)
	vec.Vec_PtrFree(roots)
	fmt.Printf("Partitioned fraiging time = %9.2f sec\n", time.Since(started).Seconds())
}

func objectSlice(objects *Vec_Ptr_t) []*Abc_Obj_t {
	result := make([]*Abc_Obj_t, vec.Vec_PtrSize(objects))
	for i, item := range vec.Vec_PtrArray(objects) {
		result[i] = item.(*Abc_Obj_t)
	}
	return result
}
func boolInt32(value bool) int32 {
	if value {
		return 1
	}
	return 0
}
