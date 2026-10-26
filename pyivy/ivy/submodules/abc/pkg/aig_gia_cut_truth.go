package abc

// Ordered cut collection/evaluation from aig/gia/giaTruth.c:
// Gia_ObjCollectInternalCut_rec/CollectInternalCut and Gia_ObjComputeTruthTableCut.

type GiaCutTruthOptions struct {
	MaxObjects        int // reached non-leaf gates/COs; zero: one million
	MaxMappingEntries int // arena-indexed scratch entries; zero: one million
	MaxWords          int // combined internal truth and owned result words; zero: one million
	Stop              func() bool
}

// GiaCutTruth owns a function over the supplied ordered leaves and the reached
// non-leaf gate/CO postorder (C's leading sentinel is omitted). Repeated leaves
// retain their first slot, leaving later slots unused. No source marks, object
// values, traversal IDs or shared truth buffers are modified.
type GiaCutTruth struct {
	Table *TruthTable
	Nodes []GiaID
}

// CutTruth computes a root literal in terms of an ordered structural boundary.
// Leaves must be nonconstant, non-CO objects; every reached CI must be covered.
// Constants inside the cone remain fixed. Root-as-leaf, constants, CO roots,
// buffers and native XOR/mux are safe extensions of C's plain-AND assumptions.
// Leaf order includes unused/duplicate slots. Limits/cancellation return no
// partial table or node order. Truth word storage is bounded before allocation.
func (m *GiaManager) CutTruth(root GiaLit, leaves []GiaID, options GiaCutTruthOptions) (GiaCutTruth, error) {
	fail := func(e error) (GiaCutTruth, error) { return GiaCutTruth{}, e }
	if options.MaxObjects < 0 || options.MaxMappingEntries < 0 || options.MaxWords < 0 {
		return fail(ErrDimension)
	}
	if m == nil {
		return fail(ErrDimension)
	}
	if e := m.Check(); e != nil {
		return fail(e)
	}
	if !root.Valid() || uint64(root.ID()) >= uint64(len(m.objects)) {
		return fail(ErrInvalidLiteral)
	}
	words, e := TruthTableWordCount(len(leaves))
	if e != nil {
		return fail(e)
	}
	maxObjects, maxEntries, maxWords := options.MaxObjects, options.MaxMappingEntries, options.MaxWords
	if maxObjects == 0 {
		maxObjects = 1 << 20
	}
	if maxEntries == 0 {
		maxEntries = 1 << 20
	}
	if maxWords == 0 {
		maxWords = 1 << 20
	}
	if len(m.objects) > maxEntries || words > maxWords {
		return fail(ErrLimit)
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	if stopped() {
		return fail(ErrCancelled)
	}
	slots := make([]int, len(m.objects))
	positions := make([]int, len(m.objects))
	states := make([]uint8, len(m.objects))
	for i := range slots {
		slots[i] = -1
		positions[i] = -1
	}
	states[0] = 2
	for slot, id := range leaves {
		if id == 0 {
			return fail(ErrDimension)
		}
		if uint64(id) >= uint64(len(m.objects)) || m.objects[id].IsCO() {
			return fail(ErrInvalidLiteral)
		}
		if slots[id] < 0 {
			slots[id] = slot
			states[id] = 2
		}
	}
	type frame struct {
		id    GiaID
		ready bool
	}
	stack := []frame{{id: root.ID()}}
	nodes := make([]GiaID, 0)
	discovered := 0
	for len(stack) > 0 {
		if stopped() {
			return fail(ErrCancelled)
		}
		item := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if states[item.id] == 2 {
			continue
		}
		if item.ready {
			positions[item.id] = len(nodes)
			nodes = append(nodes, item.id)
			states[item.id] = 2
			continue
		}
		if states[item.id] != 0 {
			continue
		}
		if m.objects[item.id].IsCI() {
			return fail(ErrDimension)
		}
		discovered++
		if discovered > maxObjects {
			return fail(ErrLimit)
		}
		states[item.id] = 1
		children, n, e := m.crossCutFanins(item.id)
		if e != nil {
			return fail(e)
		}
		stack = append(stack, frame{item.id, true})
		for i := n - 1; i >= 0; i-- {
			if states[children[i]] != 2 {
				stack = append(stack, frame{id: children[i]})
			}
		}
	}
	if len(nodes) > maxObjects {
		return fail(ErrLimit)
	}
	total, e := checkedSize(0, words, len(nodes)+1, maxWords)
	if e != nil {
		return fail(e)
	}
	if _, e = checkedSize(0, 8, total, int(^uint(0)>>1)); e != nil {
		return fail(e)
	}
	storage := make([]uint64, len(nodes)*words)
	read := func(lit GiaLit, index int) uint64 {
		value := uint64(0)
		if lit.ID() != 0 {
			if slot := slots[lit.ID()]; slot >= 0 {
				if slot < 6 {
					value = truthVariablePattern(slot)
				} else if index&(1<<uint(slot-6)) != 0 {
					value = ^uint64(0)
				}
			} else {
				value = storage[positions[lit.ID()]*words+index]
			}
		}
		if lit.Complemented() {
			value = ^value
		}
		return value
	}
	for position, id := range nodes {
		if stopped() {
			return fail(ErrCancelled)
		}
		object := m.objects[id]
		a, e := object.Fanin0(id)
		if e != nil {
			return fail(e)
		}
		b := GiaConst0
		if object.IsAnd() && !object.IsBuffer() {
			b, e = object.Fanin1(id)
			if e != nil {
				return fail(e)
			}
		}
		control, mux := m.MuxControl(id)
		for w := 0; w < words; w++ {
			first := read(a, w)
			value := first
			if !object.IsCO() && !object.IsBuffer() {
				second := read(b, w)
				if mux {
					c := read(control, w)
					value = (c & second) | (^c & first)
				} else if object.IsXor() {
					value = first ^ second
				} else {
					value = first & second
				}
			}
			storage[position*words+w] = value
		}
	}
	table, e := NewConstantTruthTable(len(leaves), false, maxWords)
	if e != nil {
		return fail(e)
	}
	for w := range table.words {
		if stopped() {
			return fail(ErrCancelled)
		}
		table.words[w] = read(root, w)
	}
	return GiaCutTruth{table, nodes}, nil
}
