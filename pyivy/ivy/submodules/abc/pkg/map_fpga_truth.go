package abc

// Cut truth and volume algorithms from src/map/fpga/fpgaTruth.c.

// FPGACutEdge is one possibly complemented input of an FPGA cut node.
// Complement affects truth evaluation but not structural volume.
type FPGACutEdge struct {
	Cut        *FPGACut
	Complement bool
}

// FPGACut is a concrete immutable cut DAG node. Nodes supplied separately as
// leaves need no inputs. Every reached non-leaf node must have both inputs.
// Phase complements the conjunction after its input complements are applied.
type FPGACut struct {
	One   FPGACutEdge
	Two   FPGACutEdge
	Phase bool
}

type FPGACutTruthOptions struct {
	MaxCuts  int // reached non-leaf cuts; zero: one million
	MaxWords int // combined intermediate and result words; zero: one million
	Stop     func() bool
}

// FPGACutVolume returns the number of distinct non-leaf cuts reached from
// root, including root. The ordered leaves form the traversal boundary. This
// is Fpga_CutVolume without temporary fMark mutation or stdout diagnostics.
func FPGACutVolume(root *FPGACut, leaves []*FPGACut, options FPGACutTruthOptions) (int, error) {
	order, _, e := collectFPGACutPostorder(root, leaves, options)
	if e != nil {
		return 0, e
	}
	return len(order), nil
}

// FPGACutTruth derives root as a packed function of the ordered leaves. It
// covers Fpga_TruthsCutBdd_rec/Fpga_TruthsCutBdd while returning an owned native
// truth table rather than a manager-owned BDD. Input cut nodes are never marked
// or otherwise mutated, so frozen DAGs may be evaluated concurrently.
func FPGACutTruth(root *FPGACut, leaves []*FPGACut, options FPGACutTruthOptions) (*TruthTable, error) {
	order, leafSlots, e := collectFPGACutPostorder(root, leaves, options)
	if e != nil {
		return nil, e
	}
	words, e := TruthTableWordCount(len(leaves))
	if e != nil {
		return nil, e
	}
	maxWords := options.MaxWords
	if maxWords == 0 {
		maxWords = 1 << 20
	}
	total, e := checkedSize(words, words, len(order), maxWords)
	if e != nil {
		return nil, e
	}
	if _, e = checkedSize(0, 8, total, int(^uint(0)>>1)); e != nil {
		return nil, e
	}
	positions := make(map[*FPGACut]int, len(order))
	storage := make([]uint64, len(order)*words)
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	read := func(edge FPGACutEdge, word int) uint64 {
		var value uint64
		if slot, found := leafSlots[edge.Cut]; found {
			if slot < 6 {
				value = truthVariablePattern(slot)
			} else if word&(1<<uint(slot-6)) != 0 {
				value = ^uint64(0)
			}
		} else {
			value = storage[positions[edge.Cut]*words+word]
		}
		if edge.Complement {
			value = ^value
		}
		return value
	}
	for position, cut := range order {
		if stopped() {
			return nil, ErrCancelled
		}
		positions[cut] = position
		for word := 0; word < words; word++ {
			value := read(cut.One, word) & read(cut.Two, word)
			if cut.Phase {
				value = ^value
			}
			storage[position*words+word] = value
		}
	}
	result := make([]uint64, words)
	copy(result, storage[(len(order)-1)*words:len(order)*words])
	return TruthTableFromWords(len(leaves), result, maxWords)
}

func collectFPGACutPostorder(root *FPGACut, leaves []*FPGACut, options FPGACutTruthOptions) ([]*FPGACut, map[*FPGACut]int, error) {
	if root == nil || len(leaves) < 2 || options.MaxCuts < 0 || options.MaxWords < 0 {
		return nil, nil, ErrDimension
	}
	maxCuts := options.MaxCuts
	if maxCuts == 0 {
		maxCuts = 1 << 20
	}
	leafSlots := make(map[*FPGACut]int, len(leaves))
	for slot, leaf := range leaves {
		if leaf == nil {
			return nil, nil, ErrDimension
		}
		if _, duplicate := leafSlots[leaf]; duplicate {
			return nil, nil, ErrDimension
		}
		leafSlots[leaf] = slot
	}
	if _, rootIsLeaf := leafSlots[root]; rootIsLeaf {
		return nil, nil, ErrDimension
	}
	type frame struct {
		cut   *FPGACut
		ready bool
	}
	states := make(map[*FPGACut]uint8)
	stack := []frame{{cut: root}}
	order := make([]*FPGACut, 0)
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	for len(stack) > 0 {
		if stopped() {
			return nil, nil, ErrCancelled
		}
		item := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if _, leaf := leafSlots[item.cut]; leaf {
			continue
		}
		if item.cut == nil {
			return nil, nil, ErrDimension
		}
		state := states[item.cut]
		if item.ready {
			if state != 1 {
				continue
			}
			states[item.cut] = 2
			order = append(order, item.cut)
			continue
		}
		if state == 2 {
			continue
		}
		if state == 1 {
			return nil, nil, ErrInvariant
		}
		if item.cut.One.Cut == nil || item.cut.Two.Cut == nil {
			return nil, nil, ErrDimension
		}
		states[item.cut] = 1
		if len(states) > maxCuts {
			return nil, nil, ErrLimit
		}
		stack = append(stack, frame{cut: item.cut, ready: true})
		stack = append(stack, frame{cut: item.cut.Two.Cut})
		stack = append(stack, frame{cut: item.cut.One.Cut})
	}
	if len(order) == 0 || len(order) > maxCuts {
		return nil, nil, ErrLimit
	}
	return order, leafSlots, nil
}
