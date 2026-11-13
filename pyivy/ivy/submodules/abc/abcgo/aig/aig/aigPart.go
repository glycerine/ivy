package aig

import (
	"fmt"
	"math/bits"
	"sort"
)

// Part_Block replaces the intrusive char* free-list records used for support
// propagation. Typed ownership keeps pooling separate from support contents.
type Part_Block struct {
	Data []byte
	next *Part_Block
}
type Part_One_t struct {
	nRefs      int32
	pOuts      []int32
	nOutsAlloc int32
}
type Part_Man_t struct {
	chunk, step int32
	free        map[int32][]*Part_Block
	live        map[*Part_Block]int32
	entries     map[int32][]*Part_One_t
	active      map[*Part_One_t]bool
}

func Part_SizeType(size, step int32) int32 {
	if size < 0 || step <= 0 {
		panic("aig: invalid partition allocation size")
	}
	count := size / step
	if size%step != 0 {
		count++
	}
	return count
}
func Part_OneNext(block *Part_Block) *Part_Block {
	if block == nil {
		return nil
	}
	return block.next
}
func Part_OneSetNext(block, next *Part_Block) {
	if block == nil {
		panic("aig: nil partition block")
	}
	block.next = next
}
func Part_ManStart(chunk, step int32) *Part_Man_t {
	if chunk <= 0 || step <= 0 {
		panic("aig: invalid partition pool size")
	}
	return &Part_Man_t{chunk: chunk, step: step, free: make(map[int32][]*Part_Block), live: make(map[*Part_Block]int32), entries: make(map[int32][]*Part_One_t), active: make(map[*Part_One_t]bool)}
}
func Part_ManStop(p *Part_Man_t) {
	if p != nil {
		*p = Part_Man_t{}
	}
}
func Part_ManFetch(p *Part_Man_t, size int32) *Part_Block {
	class := Part_SizeType(size, p.step)
	var block *Part_Block
	if list := p.free[class]; len(list) > 0 {
		block = list[len(list)-1]
		p.free[class] = list[:len(list)-1]
	} else {
		block = &Part_Block{Data: make([]byte, int(class)*int(p.step))}
	}
	block.next = nil
	p.live[block] = class
	return block
}
func Part_ManRecycle(p *Part_Man_t, block *Part_Block, size int32) {
	class := Part_SizeType(size, p.step)
	if actual, ok := p.live[block]; !ok || actual != class {
		panic("aig: foreign/recycled partition block")
	}
	delete(p.live, block)
	p.free[class] = append(p.free[class], block)
}
func Part_ManFetchEntry(p *Part_Man_t, words, refs int32) *Part_One_t {
	if words < 0 || refs < 0 {
		panic("aig: invalid support entry size")
	}
	var entry *Part_One_t
	if list := p.entries[words]; len(list) > 0 {
		entry = list[len(list)-1]
		p.entries[words] = list[:len(list)-1]
	} else {
		entry = &Part_One_t{pOuts: make([]int32, 0, words), nOutsAlloc: words}
	}
	entry.pOuts = entry.pOuts[:0]
	entry.nRefs = refs
	p.active[entry] = true
	return entry
}
func Part_ManRecycleEntry(p *Part_Man_t, entry *Part_One_t) {
	if !p.active[entry] || len(entry.pOuts) > int(entry.nOutsAlloc) {
		panic("aig: invalid/recycled support entry")
	}
	delete(p.active, entry)
	p.entries[entry.nOutsAlloc] = append(p.entries[entry.nOutsAlloc], entry)
}
func partUnion(a, b []int32) []int32 {
	result := make([]int32, 0, len(a)+len(b))
	i, j := 0, 0
	for i < len(a) || j < len(b) {
		var value int32
		if j == len(b) || (i < len(a) && a[i] < b[j]) {
			value = a[i]
			i++
		} else if i == len(a) || b[j] < a[i] {
			value = b[j]
			j++
		} else {
			value = a[i]
			i++
			j++
		}
		if len(result) == 0 || result[len(result)-1] != value {
			result = append(result, value)
		}
	}
	return result
}
func Part_ManMergeEntry(p *Part_Man_t, a, b *Part_One_t, refs int32) *Part_One_t {
	entry := Part_ManFetchEntry(p, int32(len(a.pOuts)+len(b.pOuts)), refs)
	entry.pOuts = append(entry.pOuts, partUnion(a.pOuts, b.pOuts)...)
	return entry
}
func Part_ManTransferEntry(entry *Part_One_t) []int32 { return append([]int32(nil), entry.pOuts...) }
func partOutputSupport(p *Graph, output Edge) []int32 {
	support := Aig_Support(p, p.outputDriver(output))
	result := make([]int32, len(support))
	for i, e := range support {
		result[i] = Aig_ObjCioId(p, e)
	}
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result
}

// Aig_ManSupports appends the output index to each sorted support row, then
// sorts rows by decreasing support size as in ABC's support propagation pass.
func Aig_ManSupports(p *Graph) [][]int32 {
	rows := make([][]int32, len(p.outputs))
	for i, e := range p.outputs {
		rows[i] = append(partOutputSupport(p, e), int32(i))
	}
	sort.SliceStable(rows, func(i, j int) bool { return len(rows[i]) > len(rows[j]) })
	return rows
}
func Aig_ManSupportsTest(p *Graph) {
	rows := Aig_ManSupports(p)
	if len(rows) != len(p.outputs) {
		panic("aig: support computation lost outputs")
	}
}
func Aig_ManSupportsInverse(p *Graph) [][]int32 {
	rows := make([][]int32, len(p.inputs))
	for _, row := range Aig_ManSupports(p) {
		out := row[len(row)-1]
		for _, input := range row[:len(row)-1] {
			rows[input] = append(rows[input], out)
		}
	}
	return rows
}
func Aig_ManSupportsRegisters(p *Graph) [][]int32 {
	rows := make([][]int32, p.nRegs)
	pi, po := len(p.inputs)-int(p.nRegs), len(p.outputs)-int(p.nRegs)
	for i, e := range p.outputs[po:] {
		for _, input := range partOutputSupport(p, e) {
			if int(input) >= pi {
				rows[i] = append(rows[i], input-int32(pi))
			}
		}
	}
	return rows
}
func Aig_ManSuppCharStart(support []int32, inputs int32) []uint32 {
	if inputs < 0 {
		panic("aig: negative support dimensions")
	}
	buffer := make([]uint32, (int(inputs)+31)/32)
	Aig_ManSuppCharAdd(buffer, support, inputs)
	return buffer
}
func Aig_ManSuppCharAdd(buffer []uint32, support []int32, inputs int32) {
	for _, id := range support {
		if id < 0 || id >= inputs || int(id/32) >= len(buffer) {
			panic("aig: support index outside bitmap")
		}
		buffer[id/32] |= 1 << uint(id&31)
	}
}
func Aig_ManSuppCharCommon(buffer []uint32, support []int32) int32 {
	var common int32
	for _, id := range support {
		if id < 0 || int(id/32) >= len(buffer) {
			panic("aig: support index outside bitmap")
		}
		if buffer[id/32]&(1<<uint(id&31)) != 0 {
			common++
		}
	}
	return common
}
func Aig_ManPartitionSmartFindPart(supports, parts [][]int32, bitmaps [][]uint32, limit int32, one []int32) int32 {
	if len(supports) != len(parts) || len(supports) != len(bitmaps) {
		panic("aig: inconsistent partition metadata")
	}
	best := int32(-1)
	score := int32(0)
	for i, support := range supports {
		common := Aig_ManSuppCharCommon(bitmaps[i], one)
		if common == 0 {
			continue
		}
		if int(common) == len(one) {
			return int32(i)
		}
		if limit > 0 && len(support) >= int(2*limit) {
			continue
		}
		attract := 1000 * common / int32(len(one))
		repulse := int32(1)
		if len(support) >= 100 {
			difference := len(support) - 100
			if difference > 1 {
				repulse += int32(bits.Len(uint(difference - 1)))
			}
		}
		value := attract / repulse
		if value > score {
			score = value
			best = int32(i)
		}
	}
	if score < 75 {
		return -1
	}
	return best
}
func Aig_ManPartitionPrint(p *Graph, parts, supports [][]int32) {
	if len(parts) != len(supports) {
		panic("aig: inconsistent partitions")
	}
	total := 0
	for i, part := range parts {
		fmt.Printf("%d=(%d,%d) ", i, len(supports[i]), len(part))
		total += len(part)
	}
	if total != len(p.outputs) {
		panic("aig: partition output coverage differs")
	}
}
func Aig_ManPartitionCompact(parts, supports *[][]int32, limit int32) {
	if len(*parts) != len(*supports) {
		panic("aig: inconsistent partitions")
	}
	if limit < 0 {
		return
	}
	if limit == 0 {
		limit = 200
	}
	var result, resultSupport [][]int32
	var pending, pendingSupport []int32
	flush := func() {
		if pending != nil {
			result = append(result, pending)
			resultSupport = append(resultSupport, pendingSupport)
			pending, pendingSupport = nil, nil
		}
	}
	for i, part := range *parts {
		support := (*supports)[i]
		if len(support) >= int(limit) {
			flush()
			result = append(result, append([]int32(nil), part...))
			resultSupport = append(resultSupport, append([]int32(nil), support...))
			continue
		}
		if pending == nil {
			pending = append([]int32{}, part...)
			pendingSupport = append([]int32(nil), support...)
		} else {
			pending = partUnion(pending, part)
			pendingSupport = partUnion(pendingSupport, support)
		}
		if len(pendingSupport) >= int(limit) {
			flush()
		}
	}
	flush()
	*parts, *supports = result, resultSupport
}
func partSmart(rows [][]int32, inputs, limit, verbose int32) ([][]int32, [][]int32) {
	var parts, supports [][]int32
	var bitmaps [][]uint32
	for _, row := range rows {
		if len(row) == 0 {
			panic("aig: support row has no output identifier")
		}
		id, one := row[len(row)-1], row[:len(row)-1]
		index := Aig_ManPartitionSmartFindPart(supports, parts, bitmaps, limit, one)
		if index < 0 {
			parts = append(parts, []int32{id})
			supports = append(supports, append([]int32(nil), one...))
			bitmaps = append(bitmaps, Aig_ManSuppCharStart(one, inputs))
		} else {
			parts[index] = append(parts[index], id)
			supports[index] = partUnion(supports[index], one)
			Aig_ManSuppCharAdd(bitmaps[index], one, inputs)
		}
	}
	order := make([]int, len(parts))
	for i := range order {
		order[i] = i
	}
	sort.SliceStable(order, func(i, j int) bool { return len(supports[order[i]]) > len(supports[order[j]]) })
	sortedParts, sortedSupports := make([][]int32, len(parts)), make([][]int32, len(parts))
	for i, j := range order {
		sort.Slice(parts[j], func(a, b int) bool { return parts[j][a] < parts[j][b] })
		sortedParts[i], sortedSupports[i] = parts[j], supports[j]
	}
	parts, supports = sortedParts, sortedSupports
	Aig_ManPartitionCompact(&parts, &supports, limit)
	if verbose != 0 {
		fmt.Printf("Created %d partitions.\n", len(parts))
	}
	return parts, supports
}

func Aig_ManPartitionSmart(p *Graph, limit, verbose int32, returnedSupports *[][]int32) [][]int32 {
	parts, supports := partSmart(Aig_ManSupports(p), int32(len(p.inputs)), limit, verbose)
	if returnedSupports != nil {
		*returnedSupports = supports
	}
	return parts
}
func Aig_ManPartitionSmartRegisters(p *Graph, limit, verbose int32) [][]int32 {
	rows := Aig_ManSupportsRegisters(p)
	for i := range rows {
		rows[i] = append(rows[i], int32(i))
	}
	parts, _ := partSmart(rows, p.nRegs, limit, verbose)
	return parts
}
func Aig_ManPartitionNaive(p *Graph, size int32) [][]int32 {
	if size <= 0 {
		panic("aig: partition size must be positive")
	}
	parts := make([][]int32, (len(p.outputs)+int(size)-1)/int(size))
	for i := range p.outputs {
		parts[i/int(size)] = append(parts[i/int(size)], int32(i))
	}
	return parts
}

func partCloneMode(dst, src *Graph, e Edge, mapping map[int]Edge, missingInput func(Edge) Edge, sourceAnd bool) Edge {
	id := e.NodeID()
	value, ok := mapping[id]
	if !ok {
		n := src.nodes[id]
		switch n.Kind {
		case Constant:
			value = dst.Const1()
		case Input:
			if missingInput == nil {
				panic("aig: partition omits an input")
			}
			value = missingInput(Aig_Regular(e))
		case And, ExclusiveOr:
			left := partCloneMode(dst, src, n.Fanins[0], mapping, missingInput, sourceAnd)
			right := partCloneMode(dst, src, n.Fanins[1], mapping, missingInput, sourceAnd)
			if n.Kind == ExclusiveOr && !sourceAnd {
				value = dst.NewXor(left, right)
			} else {
				value = dst.AND(left, right)
			}
		case Buffer, Output:
			value = partCloneMode(dst, src, n.Fanins[0], mapping, missingInput, sourceAnd)
		default:
			panic("aig: invalid partition node")
		}
		mapping[id] = value
	}
	if e.Complemented() {
		value = value.Not()
	}
	return value
}

// ABC partition-copy paths use Aig_And for every internal node, including EXOR.
func partCloneABC(dst, src *Graph, e Edge, mapping map[int]Edge, missingInput func(Edge) Edge) Edge {
	return partCloneMode(dst, src, e, mapping, missingInput, true)
}

// Other algorithms need the original graph's native XOR semantics.
func partClone(dst, src *Graph, e Edge, mapping map[int]Edge, missingInput func(Edge) Edge) Edge {
	return partCloneMode(dst, src, e, mapping, missingInput, false)
}
func Aig_ManDupPart_rec(dst, src *Graph, e Edge, support *[]int32) Edge {
	if src.copies == nil {
		src.copies = map[int]Edge{0: 0}
	}
	missing := func(input Edge) Edge {
		if support == nil {
			panic("aig: inverse partition has an unmapped input")
		}
		if len(*support) != len(dst.inputs) {
			panic("aig: inconsistent partition input map")
		}
		*support = append(*support, Aig_ObjCioId(src, input))
		return dst.NewInput()
	}
	return partCloneABC(dst, src, e, src.copies, missing)
}
func Aig_ManDupPart(dst, src *Graph, part []int32, support *[]int32, inverse int32) []Edge {
	if dst == src {
		panic("aig: partition source and destination must differ")
	}
	src.copies = map[int]Edge{0: 0}
	var local []int32
	if support == nil {
		support = &local
	}
	for i, index := range *support {
		if inverse == 0 {
			if index < 0 || int(index) >= len(src.inputs) || i >= len(dst.inputs) {
				panic("aig: invalid partition input map")
			}
			src.copies[src.inputs[index].NodeID()] = dst.inputs[i]
		} else {
			if i >= len(src.inputs) || index < 0 || int(index) >= len(dst.inputs) {
				panic("aig: invalid inverse input map")
			}
			src.copies[src.inputs[i].NodeID()] = dst.inputs[index]
		}
	}
	var result []Edge
	if inverse == 0 {
		for _, index := range part {
			if index < 0 || int(index) >= len(src.outputs) {
				panic("aig: partition output index outside graph")
			}
			result = append(result, Aig_ManDupPart_rec(dst, src, src.outputDriver(src.outputs[index]), support))
		}
	} else {
		for _, output := range src.outputs {
			result = append(result, Aig_ManDupPart_rec(dst, src, src.outputDriver(output), nil))
		}
		for id, n := range src.nodes {
			if !src.deleted[id] && (n.Kind == And || n.Kind == ExclusiveOr || n.Kind == Buffer) && n.Refs == 0 {
				Aig_ManDupPart_rec(dst, src, Edge(id)<<1, nil)
			}
		}
	}
	return result
}
func Aig_ManDupPartAll_rec(dst, src *Graph, e Edge) {
	id := e.NodeID()
	if _, ok := src.copies[id]; ok {
		return
	}
	n := src.nodes[id]
	var value Edge
	if n.Kind == Output {
		Aig_ManDupPartAll_rec(dst, src, n.Fanins[0])
		value = dst.AddOutput(Aig_ObjRealCopy(src, n.Fanins[0]))
	} else {
		value = partCloneABC(dst, src, Aig_Regular(e), src.copies, func(input Edge) Edge { return dst.NewInput() })
	}
	src.copies[id] = value
	if dst.copies == nil {
		dst.copies = map[int]Edge{0: 0}
	}
	dst.copies[value.NodeID()] = Aig_Regular(e)
}
func Aig_ManDupPartAll(src *Graph, part []int32) *Graph {
	dst := NewGraph()
	dst.Name, dst.Spec = src.Name, src.Spec
	src.copies = map[int]Edge{0: 0}
	dst.copies = map[int]Edge{0: 0}
	for _, index := range part {
		if index < 0 || int(index) >= len(src.outputs) {
			panic("aig: partition output index outside graph")
		}
		Aig_ManDupPartAll_rec(dst, src, src.outputs[index])
	}
	// Record reverse mappings for every copied object, not just output roots.
	for sourceID, target := range src.copies {
		dst.copies[target.NodeID()] = Edge(sourceID) << 1
	}
	return dst
}
func Aig_ManSupportNodes_rec(p *Graph, e Edge, support *[]int32) {
	seen := make(map[int32]bool)
	for _, id := range *support {
		seen[id] = true
	}
	for _, input := range Aig_Support(p, e) {
		index := Aig_ObjCioId(p, input)
		if !seen[index] {
			*support = append(*support, index)
			seen[index] = true
		}
	}
}
func Aig_ManSupportNodes(p *Graph, parts [][]int32) [][]int32 {
	supports := make([][]int32, len(parts))
	for i, part := range parts {
		for _, index := range part {
			if index < 0 || int(index) >= len(p.outputs) {
				panic("aig: partition output index outside graph")
			}
			Aig_ManSupportNodes_rec(p, p.outputDriver(p.outputs[index]), &supports[i])
		}
	}
	return supports
}
func Aig_ManMiterPartitioned(a, b *Graph, size, smart int32) []*Graph {
	var parts, supports [][]int32
	if smart != 0 {
		parts = Aig_ManPartitionSmart(a, size, 0, &supports)
	} else {
		parts = Aig_ManPartitionNaive(a, size)
		supports = Aig_ManSupportNodes(a, parts)
	}
	result := make([]*Graph, len(parts))
	for i, part := range parts {
		q := NewGraph()
		inputMap := append([]int32(nil), supports[i]...)
		for range inputMap {
			q.NewInput()
		}
		left := Aig_ManDupPart(q, a, part, &inputMap, 0)
		right := Aig_ManDupPart(q, b, part, &inputMap, 0)
		q.AddOutput(Aig_MiterTwo(q, left, right))
		Aig_ManCleanup(q)
		q.PartitionInputs = append([]int32(nil), inputMap...)
		q.PartitionOutputs = append([]int32(nil), part...)
		result[i] = q
	}
	return result
}

// partExactClasses proves equivalence by exact Shannon reduction in separate
// scratch storage. It does not pollute the manager with auxiliary decision gates.
// The conflict bound is retained for the source API; this backend produces no
// SAT conflicts. The source's level bound still filters candidate nodes.
func partExactClasses(p *Graph, conflicts, levelLimit int32) {
	if conflicts < 0 || levelLimit < 0 {
		panic("aig: negative equivalence-analysis budget")
	}
	if p.reprs == nil {
		Aig_ManReprStart(p, int32(len(p.nodes)))
	}
	decisions := NewGraph()
	mapping := map[int]Edge{0: 0}
	for _, input := range p.inputs {
		mapping[input.NodeID()] = decisions.NewInput()
	}
	memo := make(map[Edge]Edge)
	classes := map[Edge]Edge{decisions.Const0(): p.Const1()}
	for id, n := range p.nodes {
		if p.deleted[id] || n.Kind == Constant || n.Kind == Output {
			continue
		}
		if n.Kind != Input && n.Kind != And && n.Kind != ExclusiveOr {
			panic("aig: equivalence analysis requires propagated buffers")
		}
		e := Edge(id) << 1
		function := partClone(decisions, p, e, mapping, nil)
		if n.Phase {
			function = function.Not()
		}
		key := splitReduce(decisions, function, memo)
		representative, exists := classes[key]
		if !exists {
			classes[key] = e
			continue
		}
		if representative.NodeID() == id {
			continue
		}
		if levelLimit > 0 && n.Level > int(levelLimit) {
			continue
		}
		if representative.NodeID() < id && Aig_ObjCheckTfi(p, representative, e) == 0 {
			Aig_ObjSetRepr_(p, representative, e)
		}
	}
}
func Aig_ManFraigPartitioned(p *Graph, size, conflicts, levelLimit, verbose int32) *Graph {
	parts := Aig_ManPartitionNaive(p, size)
	p.reprs = nil
	Aig_ManReprStart(p, int32(len(p.nodes)))
	for i, part := range parts {
		fragment := Aig_ManDupPartAll(p, part)
		partExactClasses(fragment, conflicts, levelLimit)
		for id, representative := range fragment.reprs {
			a, okA := fragment.copies[id]
			b, okB := fragment.copies[representative.NodeID()]
			if okA && okB && a != b {
				Aig_ObjSetRepr_(p, Aig_Regular(a), Aig_Regular(b))
			}
		}
		if verbose != 0 {
			fmt.Printf("Part %d/%d: inputs=%d outputs=%d nodes=%d\n", i+1, len(parts), len(fragment.inputs), len(fragment.outputs), Aig_ManNodeNum(fragment))
		}
	}
	return Aig_ManDupRepr(p, 0)
}
func Aig_ManChoicePartitioned(versions []*Graph, size, conflicts, levelLimit, verbose int32) *Graph {
	if len(versions) < 2 {
		panic("aig: choice partitioning needs at least two graphs")
	}
	base := versions[0]
	for _, p := range versions {
		if len(p.inputs) != len(base.inputs) || len(p.outputs) != len(base.outputs) || p.nRegs != base.nRegs {
			panic("aig: choice graph terminal counts differ")
		}
	}
	parts := Aig_ManPartitionSmart(base, size, 0, nil)
	total := Aig_ManStartFrom(base)
	Aig_ManReprStart(total, int32(len(base.nodes)))
	outputs := make([]Edge, len(base.outputs))
	for i := range outputs {
		outputs[i] = InvalidEdge
	}
	for i, part := range parts {
		fragment := NewGraph()
		var inputMap []int32
		for v, p := range versions {
			roots := Aig_ManDupPart(fragment, p, part, &inputMap, 0)
			if v == 0 {
				for _, root := range roots {
					fragment.AddOutput(root)
				}
			}
		}
		restored := Aig_ManDupPart(total, fragment, part, &inputMap, 1)
		for j, e := range restored {
			outputs[part[j]] = e
		}
		// The inverse transfer establishes fragment.copies before classification;
		// classification's scratch graph does not overwrite those source mappings.
		partExactClasses(fragment, conflicts, levelLimit)
		for id, representative := range fragment.reprs {
			a, okA := fragment.copies[id]
			b, okB := fragment.copies[representative.NodeID()]
			if okA && okB && a != b {
				Aig_ObjSetRepr_(total, Aig_Regular(a), Aig_Regular(b))
			}
		}
		if verbose != 0 {
			fmt.Printf("Part %d/%d: inputs=%d outputs=%d nodes=%d\n", i+1, len(parts), len(fragment.inputs), len(fragment.outputs), Aig_ManNodeNum(fragment))
		}
	}
	for _, e := range outputs {
		if e == InvalidEdge {
			panic("aig: output omitted by choice partitions")
		}
		total.AddOutput(e)
	}
	dupMetadata(total, base)
	total = Aig_ManRehash(total)
	Aig_ManMarkValidChoices(total)
	order := Aig_ManOrderPios(total, base)
	result := Aig_ManDupDfsGuided(total, order)
	Aig_ManChoiceLevel(result)
	return result
}

// Aig_ObjSetRepr_ is shared with aigRepr.go; links always decrease object IDs.
func Aig_ManChoiceConstructiveOne(destination, previous, current *Graph) {
	if len(destination.inputs) != len(previous.inputs) || len(destination.inputs) != len(current.inputs) || len(destination.outputs) != len(current.outputs) || len(destination.outputs) != len(previous.outputs) {
		panic("aig: constructive choice terminal counts differ")
	}
	current.copies = map[int]Edge{0: 0}
	for i, e := range current.inputs {
		current.copies[e.NodeID()] = destination.inputs[i]
	}
	for id, n := range current.nodes {
		if !current.deleted[id] && (n.Kind == And || n.Kind == ExclusiveOr || n.Kind == Buffer) {
			partClone(destination, current, Edge(id)<<1, current.copies, nil)
		}
	}
	proof := NewGraph()
	for range destination.inputs {
		proof.NewInput()
	}
	memo := make(map[Edge]Edge)
	if destination.reprs == nil {
		Aig_ManReprStart(destination, int32(len(destination.nodes)))
	}
	for i, output := range current.outputs {
		copied := partClone(destination, current, current.outputDriver(output), current.copies, nil)
		original := destination.outputDriver(destination.outputs[i])
		a := interCopy(destination, proof, original, proof.inputs)
		b := interCopy(destination, proof, copied, proof.inputs)
		if splitReduce(proof, proof.XOR(a, b), memo) != proof.Const0() {
			panic("aig: constructive choices are not equivalent")
		}
		Aig_ObjSetRepr_(destination, Aig_Regular(original), Aig_Regular(copied))
		current.copies[output.NodeID()] = destination.outputs[i]
	}
}
func Aig_ManChoiceEval(p *Graph) {
	for id, n := range p.nodes {
		if p.deleted[id] || n.Refs == 0 {
			continue
		}
		if _, ok := p.equivs[id]; !ok {
			continue
		}
		count := 0
		seen := make(map[int]bool)
		for e := Edge(id) << 1; e != InvalidEdge; e = Aig_ObjEquiv(p, e) {
			if seen[e.NodeID()] {
				panic("aig: cyclic choice list")
			}
			seen[e.NodeID()] = true
			count++
		}
		fmt.Printf("Choice node = %5d. Level = %2d. Choices = %d. { ", id, n.Level, count)
		for e := Edge(id) << 1; e != InvalidEdge; e = Aig_ObjEquiv(p, e) {
			var support []Edge
			size := Aig_NodeMffcSupp(p, e, 0, &support)
			fmt.Printf("S=%d N=%d L=%d  ", len(support), size, p.node(e).Level)
		}
		fmt.Println("}")
	}
}
func Aig_ManChoiceConstructive(versions []*Graph, verbose int32) *Graph {
	if len(versions) == 0 {
		panic("aig: empty choice graph list")
	}
	result := Aig_ManDupOrdered(versions[0])
	result.reprs = nil
	result.equivs = nil
	Aig_ManReprStart(result, int32(len(result.nodes)))
	for i := 1; i < len(versions); i++ {
		Aig_ManChoiceConstructiveOne(result, versions[i-1], versions[i])
	}
	result = Aig_ManRehash(result)
	Aig_ManMarkValidChoices(result)
	result = Aig_ManDupDfsGuided(result, Aig_ManOrderPios(result, versions[0]))
	Aig_ManChoiceLevel(result)
	if verbose != 0 {
		Aig_ManChoiceEval(result)
	}
	return result
}
