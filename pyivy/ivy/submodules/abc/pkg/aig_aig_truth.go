package abc

// Cut truth evaluation from aig/aig/aigTruth.c.

type AigCutTruthOptions struct {
	MaxObjects        int // reached non-leaf ANDs; zero: one million
	MaxMappingEntries int // object-indexed scratch entries; zero: one million
	MaxWords          int // internal storage plus returned words; zero: one million
	Stop              func() bool
}

type AigCutTruth struct {
	Table *TruthTable
	Nodes []AigID
}

// AigCutTruthOne implements Aig_ManCutTruthOne on owned truth tables. Native
// XOR objects have regular fanins, so complemented XOR inputs are rejected.
func AigCutTruthOne(first, second *TruthTable, operation AigType, complementFirst, complementSecond bool) (*TruthTable, error) {
	if operation != AigAND && operation != AigXOR {
		return nil, ErrDimension
	}
	if first == nil || second == nil {
		return nil, ErrDimension
	}
	if operation == AigXOR {
		if complementFirst || complementSecond {
			return nil, ErrInvariant
		}
		return first.Xor(second, false)
	}
	return first.And(second, complementFirst, complementSecond)
}

// CutTruth computes root in terms of the supplied ordered structural boundary.
// Repeated leaves retain their first slot. Constants are fixed and a
// complemented root is supported without changing the source graph. Every
// reached CI must be a leaf. Limits and cancellation return no partial result.
func (g *AigManager) CutTruth(root AigLit, leaves []AigID, o AigCutTruthOptions) (AigCutTruth, error) {
	fail := func(err error) (AigCutTruth, error) { return AigCutTruth{}, err }
	if o.MaxObjects < 0 || o.MaxMappingEntries < 0 || o.MaxWords < 0 {
		return fail(ErrDimension)
	}
	if g == nil {
		return fail(ErrDimension)
	}
	if err := g.Check(); err != nil {
		return fail(err)
	}
	if !g.validSignal(root) {
		return fail(ErrInvalidLiteral)
	}
	words, err := TruthTableWordCount(len(leaves))
	if err != nil {
		return fail(err)
	}
	maxObjects, maxEntries, maxWords := o.MaxObjects, o.MaxMappingEntries, o.MaxWords
	if maxObjects == 0 {
		maxObjects = 1 << 20
	}
	if maxEntries == 0 {
		maxEntries = 1 << 20
	}
	if maxWords == 0 {
		maxWords = 1 << 20
	}
	if len(g.objects) > maxEntries || words > maxWords {
		return fail(ErrLimit)
	}
	stopped := func() bool { return o.Stop != nil && o.Stop() }
	if stopped() {
		return fail(ErrCancelled)
	}
	slots := make([]int, len(g.objects))
	positions := make([]int, len(g.objects))
	states := make([]uint8, len(g.objects))
	for i := range slots {
		slots[i], positions[i] = -1, -1
	}
	states[0] = 2
	for slot, id := range leaves {
		if stopped() {
			return fail(ErrCancelled)
		}
		if id == 0 {
			return fail(ErrDimension)
		}
		if uint64(id) >= uint64(len(g.objects)) || (g.objects[id].Type != AigCI && g.objects[id].Type != AigAND) {
			return fail(ErrInvalidLiteral)
		}
		if slots[id] < 0 {
			slots[id] = slot
			states[id] = 2
		}
	}
	type frame struct {
		id    AigID
		ready bool
	}
	stack := []frame{{id: root.ID()}}
	nodes := make([]AigID, 0)
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
		if states[item.id] == 1 {
			return fail(ErrInvariant)
		}
		object := g.objects[item.id]
		if object.Type == AigCI {
			return fail(ErrDimension)
		}
		if object.Type != AigAND {
			return fail(ErrInvalidLiteral)
		}
		discovered++
		if discovered > maxObjects {
			return fail(ErrLimit)
		}
		states[item.id] = 1
		stack = append(stack, frame{id: item.id, ready: true})
		if states[object.Fanin1.ID()] != 2 {
			stack = append(stack, frame{id: object.Fanin1.ID()})
		}
		if states[object.Fanin0.ID()] != 2 {
			stack = append(stack, frame{id: object.Fanin0.ID()})
		}
	}
	total, err := checkedSize(words, words, len(nodes), maxWords)
	if err != nil {
		return fail(err)
	}
	if _, err = checkedSize(0, 8, total, int(^uint(0)>>1)); err != nil {
		return fail(err)
	}
	storage := make([]uint64, len(nodes)*words)
	read := func(literal AigLit, word int) uint64 {
		value := ^uint64(0)
		if literal.ID() != 0 {
			if slot := slots[literal.ID()]; slot >= 0 {
				if slot < 6 {
					value = truthVariablePattern(slot)
				} else if word&(1<<uint(slot-6)) != 0 {
					value = ^uint64(0)
				} else {
					value = 0
				}
			} else {
				value = storage[positions[literal.ID()]*words+word]
			}
		}
		if literal.Complemented() {
			value = ^value
		}
		return value
	}
	for position, id := range nodes {
		if stopped() {
			return fail(ErrCancelled)
		}
		object := g.objects[id]
		for word := 0; word < words; word++ {
			storage[position*words+word] = read(object.Fanin0, word) & read(object.Fanin1, word)
		}
	}
	table, err := NewConstantTruthTable(len(leaves), false, maxWords)
	if err != nil {
		return fail(err)
	}
	for word := range table.words {
		if stopped() {
			return fail(ErrCancelled)
		}
		table.words[word] = read(root, word)
	}
	return AigCutTruth{Table: table, Nodes: nodes}, nil
}
