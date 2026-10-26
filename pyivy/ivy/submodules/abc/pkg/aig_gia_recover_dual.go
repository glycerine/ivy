package abc

// Structural dual recovery from src/aig/gia/giaDup.c:Gia_ManCollectTopXors,
// Gia_ManDupDemiterOrderXors, Gia_ManDecideWhereToAdd and Gia_ManDemiterToDual.

// GiaRecoveredDual maps target COs to phased source signals. Candidates records
// the support-ordered roots before partitioning (small/new supports first).
// Signals are normalized to false under zero CIs except C's constant-first
// special case. This is structural recovery, not a proof of equivalence.
type GiaRecoveredDual struct {
	Graph      *GiaManager
	Signals    []GiaLit
	Candidates []GiaID
}

// RecoverDualOutputs recovers adjacent signal pairs using C's greedy support
// order and maximum same-side cone overlap. Source output polarity and XOR
// operand phases are intentionally replaced by zero-input phase normalization,
// matching C. All CIs survive as combinational inputs, including source ROs;
// source RIs participate as candidate outputs. Constraint/state roles clear.
// C's first-CO-constant special case emits exactly (false, that constant),
// even when other COs exist. Name/spec survive and source scratch is untouched.
// Extended gates are opaque top-level boundaries except recognized native XORs;
// their full fanins participate in support/overlap and raw copying.
func (m *GiaManager) RecoverDualOutputs(options GiaCofactorOptions) (GiaRecoveredDual, error) {
	fail := func(e error) (GiaRecoveredDual, error) { return GiaRecoveredDual{}, e }
	if m == nil || options.MaxObjects < 0 || options.MaxMappingEntries < 0 {
		return fail(ErrDimension)
	}
	if e := m.Check(); e != nil {
		return fail(e)
	}
	if len(m.cos) == 0 {
		return fail(ErrDimension)
	}
	limit := options.MaxMappingEntries
	if limit == 0 {
		limit = 1 << 20
	}
	if len(m.objects) > limit {
		return fail(ErrLimit)
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	if stopped() {
		return fail(ErrCancelled)
	}
	first, e := m.objects[m.cos[0]].Fanin0(m.cos[0])
	if e != nil {
		return fail(e)
	}
	candidates := []GiaID{}
	seen := make([]bool, len(m.objects))
	add := func(id GiaID) {
		if !seen[id] {
			seen[id] = true
			candidates = append(candidates, id)
		}
	}
	if len(m.cos) > 1 {
		for _, co := range m.cos {
			d, err := m.objects[co].Fanin0(co)
			if err != nil {
				return fail(err)
			}
			if d.ID() != 0 {
				candidates = append(candidates, d.ID())
			}
		}
	} else if !first.Complemented() {
		candidates = append(candidates, first.ID())
	} else {
		type task struct {
			id       GiaID
			boundary bool
		}
		stack := []task{{id: first.ID()}}
		expanded := make([]bool, len(m.objects))
		for len(stack) > 0 {
			if stopped() {
				return fail(ErrCancelled)
			}
			item := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			id := item.id
			if item.boundary {
				add(id)
				continue
			}
			_, _, xor, err := m.RecognizeXor(id)
			if err != nil {
				return fail(err)
			}
			obj := m.objects[id]
			_, mux := m.MuxControl(id)
			if xor || !obj.IsAnd() || obj.IsBuffer() || mux {
				add(id)
				continue
			}
			if expanded[id] {
				continue
			}
			expanded[id] = true
			a, err := obj.Fanin0(id)
			if err != nil {
				return fail(err)
			}
			b, err := obj.Fanin1(id)
			if err != nil {
				return fail(err)
			}
			stack = append(stack, task{b.ID(), b.Complemented()}, task{a.ID(), a.Complemented()})
		}
	}

	supports, e := m.NodeSupports(candidates, limit, limit)
	if e != nil {
		return fail(e)
	}
	usedInputs := make([]bool, len(m.cis))
	taken := make([]bool, len(candidates))
	ordered := make([]GiaID, 0, len(candidates))
	for range candidates {
		if stopped() {
			return fail(ErrCancelled)
		}
		best, bestCost := -1, int(^uint(0)>>1)
		for i, set := range supports {
			if taken[i] {
				continue
			}
			cost := 0
			for _, ci := range set {
				if !usedInputs[ci] {
					cost++
				}
			}
			if cost < bestCost {
				best, bestCost = i, cost
			}
		}
		ordered = append(ordered, candidates[best])
		taken[best] = true
		for _, ci := range supports[best] {
			usedInputs[ci] = true
		}
	}
	marks := [2][]bool{make([]bool, len(m.objects)), make([]bool, len(m.objects))}
	cone := func(root GiaID) ([]GiaID, error) {
		found := make([]bool, len(m.objects))
		stack := []GiaID{root}
		ids := []GiaID{}
		for len(stack) > 0 {
			if stopped() {
				return nil, ErrCancelled
			}
			id := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if found[id] {
				continue
			}
			found[id] = true
			ids = append(ids, id)
			children, n, err := m.crossCutFanins(id)
			if err != nil {
				return nil, err
			}
			stack = append(stack, children[:n]...)
		}
		return ids, nil
	}
	pairs := make([][2]GiaID, 0, len(ordered))
	for i := len(ordered) - 1; i >= 0; i-- {
		if stopped() {
			return fail(ErrCancelled)
		}
		a, b, ok, err := m.RecognizeXor(ordered[i])
		if err != nil {
			return fail(err)
		}
		ids := [2]GiaID{ordered[i], 0}
		if ok {
			ids = [2]GiaID{a.ID(), b.ID()}
		}
		cones := [2][]GiaID{}
		for side, id := range ids {
			cones[side], err = cone(id)
			if err != nil {
				return fail(err)
			}
		}
		flip := false
		if len(pairs) > 0 {
			scores := [2]int{}
			for side, nodes := range cones {
				for _, id := range nodes {
					if marks[side][id] {
						scores[0]++
					}
					if marks[1-side][id] {
						scores[1]++
					}
				}
			}
			flip = scores[0] < scores[1]
		}
		if flip {
			ids[0], ids[1] = ids[1], ids[0]
			cones[0], cones[1] = cones[1], cones[0]
		}
		pairs = append(pairs, ids)
		for side, nodes := range cones {
			for _, id := range nodes {
				marks[side][id] = true
			}
		}
	}
	values, e := m.evaluateValues(make([]bool, len(m.cis)))
	if e != nil {
		return fail(e)
	}
	signals := make([]GiaLit, 0, 2*len(pairs))
	for i := len(pairs) - 1; i >= 0; i-- {
		for side := 1; side >= 0; side-- {
			id := pairs[i][side]
			signals = append(signals, GiaLit(id<<1).NotIf(values[id]))
		}
	}
	// C collects/copies candidates before handling a constant first output.
	roots := make([]GiaID, len(signals))
	for i, l := range signals {
		roots[i] = l.ID()
	}
	nodes, e := m.CollectTransitiveFanins(roots, limit)
	if e != nil {
		return fail(e)
	}
	g, mapping, e := m.DuplicateFromLists(m.cis, nodes, nil, 0, options)
	if e != nil {
		return fail(e)
	}
	if first.ID() == 0 {
		signals = []GiaLit{GiaConst0, first}
	}
	for _, l := range signals {
		if stopped() {
			return fail(ErrCancelled)
		}
		v, err := mappedGiaSignal(mapping, l)
		if err != nil {
			return fail(err)
		}
		if _, e = g.AppendCO(v); e != nil {
			return fail(e)
		}
	}
	if e = g.Check(); e != nil {
		return fail(e)
	}
	return GiaRecoveredDual{g, signals, ordered}, nil
}
