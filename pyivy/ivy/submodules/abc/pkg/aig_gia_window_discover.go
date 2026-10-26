package abc

// Pivot window growth from aig/gia/giaSplit.c:Spl_ManComputeOne and helpers.

import "slices"

type GiaDiscoveredWindow struct {
	Found                      bool
	Roots, Nodes, Leaves, Ands []GiaID // native sorted object-ID order
}

// DiscoverWindow grows an owned window around a mapped pivot. Nodes are LUT
// roots; Ands contains their unique structural cones. Growth stops at the first
// candidate whose additional cone exceeds limit. Reverse enables the native
// odd-window fanout preference. Source state is unchanged. Validation and
// discovery each use MaxVisits; all traversal is iterative. Native muxes are
// rejected because native cone collection omits their control input.
func (n *GiaLUTNetwork) DiscoverWindow(m *GiaManager, pivot GiaID, limit int, reverse bool, o GiaLUTAnalysisOptions) (GiaDiscoveredWindow, error) {
	return n.discoverWindow(m, pivot, limit, reverse, o, nil)
}

func (n *GiaLUTNetwork) discoverWindow(m *GiaManager, pivot GiaID, limit int, reverse bool, o GiaLUTAnalysisOptions, session *giaWindowDiscoverySession) (GiaDiscoveredWindow, error) {
	fail := func(e error) (GiaDiscoveredWindow, error) { return GiaDiscoveredWindow{}, e }
	if limit < 0 {
		return fail(ErrDimension)
	}
	if n == nil || m == nil || len(n.fanins) != len(m.objects) || o.MaxObjects < 0 || o.MaxVisits < 0 {
		return fail(ErrDimension)
	}
	if int(pivot) >= len(n.fanins) || len(n.fanins[pivot]) == 0 {
		return fail(ErrDimension)
	}
	if o.MaxObjects == 0 {
		o.MaxObjects = 1 << 20
	}
	if o.MaxVisits == 0 {
		o.MaxVisits = 1 << 20
	}
	work := 0
	charge := func() error {
		if o.Stop != nil && o.Stop() {
			return ErrCancelled
		}
		if work == o.MaxVisits {
			return ErrLimit
		}
		work++
		return nil
	}
	if session == nil {
		session = &giaWindowDiscoverySession{}
	}
	if e := session.prepare(n, m, o, charge); e != nil {
		return fail(e)
	}
	io, in, no, used := session.io, session.in, session.no, session.used
	refs, fanouts := session.refs, session.fanouts
	r := GiaDiscoveredWindow{}
	seen := session.seen
	epoch := 0
	cone := func(root GiaID) ([]GiaID, error) {
		epoch++
		seen[0] = epoch
		for _, f := range n.fanins[root] {
			seen[f] = epoch
		}
		stack := []GiaID{root}
		var result []GiaID
		for len(stack) > 0 {
			if e := charge(); e != nil {
				return nil, e
			}
			id := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if seen[id] == epoch {
				continue
			}
			seen[id] = epoch
			obj := m.objects[id]
			if obj.IsCI() {
				continue
			}
			if !obj.IsAnd() {
				return nil, ErrDimension
			}
			if !used[id] {
				result = append(result, id)
			}
			a, _ := obj.Fanin0(id)
			b, _ := obj.Fanin1(id)
			stack = append(stack, b.ID(), a.ID())
		}
		return result, nil
	}
	add := func(id GiaID, ands []GiaID) {
		r.Nodes = append(r.Nodes, id)
		no[id] = true
		for _, a := range ands {
			used[a] = true
			r.Ands = append(r.Ands, a)
		}
	}
	externalFanouts := func(id GiaID) ([]GiaID, error) {
		epoch++
		stack := make([]GiaID, 0, len(fanouts[id]))
		for i := len(fanouts[id]) - 1; i >= 0; i-- {
			stack = append(stack, fanouts[id][i])
		}
		var result []GiaID
		for len(stack) > 0 {
			if e := charge(); e != nil {
				return nil, e
			}
			id := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if no[id] || io[id] || seen[id] == epoch {
				continue
			}
			seen[id] = epoch
			if len(n.fanins[id]) > 0 {
				result = append(result, id)
				continue
			}
			for i := len(fanouts[id]) - 1; i >= 0; i-- {
				stack = append(stack, fanouts[id][i])
			}
		}
		return result, nil
	}
	best := func(cands, inputs []GiaID) (GiaID, error) {
		for _, id := range inputs {
			in[id] = true
		}
		defer func() {
			for _, id := range inputs {
				in[id] = false
			}
		}()
		bestCount := -1
		var res GiaID
		for _, id := range cands {
			count := 0
			for _, f := range n.fanins[id] {
				if e := charge(); e != nil {
					return 0, e
				}
				if in[f] {
					count++
				}
			}
			if count > bestCount {
				bestCount = count
				res = id
			}
		}
		return res, nil
	}
	find := func() (GiaID, error) {
		var decremented []GiaID
		defer func() {
			for _, id := range decremented {
				refs[id]++
			}
		}()
		for _, id := range r.Nodes {
			for _, f := range n.fanins[id] {
				if e := charge(); e != nil {
					return 0, e
				}
				if e := charge(); e != nil {
					return 0, e
				}
				refs[f]--
				decremented = append(decremented, f)
			}
		}
		if reverse && len(r.Nodes)&1 != 0 {
			for _, id := range r.Nodes {
				if refs[id] == 0 || refs[id] >= 5 {
					continue
				}
				fs, e := externalFanouts(id)
				if e != nil {
					return 0, e
				}
				if refs[id] == 1 && len(fs) == 1 {
					return fs[0], nil
				}
			}
		}
		var cands, inputs []GiaID
		for _, id := range r.Nodes {
			for _, f := range n.fanins[id] {
				if e := charge(); e != nil {
					return 0, e
				}
				if !no[f] && !io[f] && refs[f] == 0 {
					cands = append(cands, f)
					inputs = append(inputs, f)
				}
			}
		}
		res, e := best(cands, inputs)
		if e != nil || res != 0 {
			return res, e
		}
		cands = nil
		inputs = nil
		for _, id := range r.Nodes {
			for _, f := range n.fanins[id] {
				if e := charge(); e != nil {
					return 0, e
				}
				if !no[f] && !io[f] {
					cands = append(cands, f)
					inputs = append(inputs, f)
				}
			}
		}
		for _, id := range r.Nodes {
			if refs[id] == 0 || refs[id] >= 5 {
				continue
			}
			fs, e := externalFanouts(id)
			if e != nil {
				return 0, e
			}
			if refs[id] == 1 && len(fs) == 1 {
				return fs[0], nil
			}
			cands = append(cands, fs...)
		}
		res, e = best(cands, inputs)
		if e != nil {
			return 0, e
		}
		if res == 0 && len(cands) > 0 {
			res = cands[0]
		}
		return res, nil
	}
	initial, e := cone(pivot)
	if e != nil {
		return fail(e)
	}
	if len(initial) > limit {
		return GiaDiscoveredWindow{}, nil
	}
	add(pivot, initial)
	for {
		candidate, e := find()
		if e != nil {
			return fail(e)
		}
		if candidate == 0 {
			break
		}
		if no[candidate] || len(n.fanins[candidate]) == 0 {
			return fail(ErrInvariant)
		}
		more, e := cone(candidate)
		if e != nil {
			return fail(e)
		}
		if len(more) > limit-len(r.Ands) {
			break
		}
		add(candidate, more)
	}
	slices.Sort(r.Nodes)
	slices.Sort(r.Ands)
	leafSeen := session.leafSeen
	for _, id := range r.Ands {
		obj := m.objects[id]
		a, _ := obj.Fanin0(id)
		b, _ := obj.Fanin1(id)
		for _, f := range []GiaID{a.ID(), b.ID()} {
			if e := charge(); e != nil {
				return fail(e)
			}
			if !used[f] && !leafSeen[f] {
				if !io[f] && len(n.fanins[f]) == 0 {
					return fail(ErrInvariant)
				}
				leafSeen[f] = true
				r.Leaves = append(r.Leaves, f)
			}
		}
	}
	for _, id := range r.Nodes {
		for _, f := range n.fanins[id] {
			if e := charge(); e != nil {
				return fail(e)
			}
			refs[f]--
		}
	}
	for _, id := range r.Ands {
		if e := charge(); e != nil {
			return fail(e)
		}
		if refs[id] != 0 {
			r.Roots = append(r.Roots, id)
		}
	}
	slices.Sort(r.Leaves)
	slices.Sort(r.Roots)
	r.Found = true
	if o.Stop != nil && o.Stop() {
		return fail(ErrCancelled)
	}
	return r, nil
}
