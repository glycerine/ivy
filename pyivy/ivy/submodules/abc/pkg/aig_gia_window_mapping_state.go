package abc

// Existing-window mapping state from aig/gia/giaSatLut.c:Sbl_ManComputeCuts.

type GiaWindowMappingState struct {
	RootVariables []int // CI roots omitted; other roots retain order/duplicates
	InitialNodes  []int
	PreferredCuts []int // parallel to InitialNodes; -1 when the current cut is absent
}

// WindowMappingState finds current mapping cuts by their exact four masks.
// Native unset/outside-window fanins are skipped. This lookup does not certify
// mapping closure or Boolean equivalence. A mapped node remains in InitialNodes
// even when no generated cut matches it.
func (n *GiaLUTNetwork) WindowMappingState(m *GiaManager, leaves, nodes, roots []GiaID, cuts GiaWindowCuts, o GiaLUTAnalysisOptions) (GiaWindowMappingState, error) {
	fail := func(e error) (GiaWindowMappingState, error) { return GiaWindowMappingState{}, e }
	if n == nil || m == nil || len(n.fanins) != len(m.objects) || len(cuts.Ranges) != len(nodes) || len(leaves) > 128 || len(nodes) > 128 || o.MaxObjects < 0 || o.MaxVisits < 0 {
		return fail(ErrDimension)
	}
	if o.MaxObjects == 0 {
		o.MaxObjects = 1 << 20
	}
	if o.MaxVisits == 0 {
		o.MaxVisits = 1 << 20
	}
	total := 0
	for _, count := range []int{len(leaves), len(nodes), len(roots), len(cuts.Cuts)} {
		var e error
		total, e = checkedSize(total, 1, count, o.MaxObjects)
		if e != nil {
			return fail(e)
		}
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
	if e := charge(); e != nil {
		return fail(e)
	}
	pos := make(map[GiaID]int, len(leaves)+len(nodes))
	for category, list := range [][]GiaID{leaves, nodes} {
		for i, id := range list {
			if e := charge(); e != nil {
				return fail(e)
			}
			if int(id) >= len(m.objects) {
				return fail(ErrDimension)
			}
			if _, ok := pos[id]; ok {
				return fail(ErrDimension)
			}
			index := i
			if category == 1 {
				index += len(leaves)
				if !m.objects[id].IsAnd() {
					return fail(ErrDimension)
				}
			} else if m.objects[id].IsCO() {
				return fail(ErrDimension)
			}
			pos[id] = index
		}
	}
	r := GiaWindowMappingState{}
	for _, id := range roots {
		if e := charge(); e != nil {
			return fail(e)
		}
		if int(id) >= len(m.objects) {
			return fail(ErrDimension)
		}
		if m.objects[id].IsCI() {
			continue
		}
		p, ok := pos[id]
		if !ok || p < len(leaves) || len(n.fanins[id]) == 0 {
			return fail(ErrDimension)
		}
		r.RootVariables = append(r.RootVariables, p-len(leaves))
	}
	for i, id := range nodes {
		if e := charge(); e != nil {
			return fail(e)
		}
		span := cuts.Ranges[i]
		if span.Start < 0 || span.Count < 1 || span.Start > len(cuts.Cuts) || span.Count > len(cuts.Cuts)-span.Start {
			return fail(ErrDimension)
		}
		if len(n.fanins[id]) == 0 {
			continue
		}
		r.InitialNodes = append(r.InitialNodes, i)
		wanted := GiaWindowCut{Node: i}
		for _, fanin := range n.fanins[id] {
			if e := charge(); e != nil {
				return fail(e)
			}
			if int(fanin) >= len(m.objects) {
				return fail(ErrDimension)
			}
			p, ok := pos[fanin]
			if !ok {
				continue
			}
			if p < len(leaves) {
				wanted.Leaves[p/64] |= 1 << uint(p%64)
			} else {
				p -= len(leaves)
				if len(n.fanins[fanin]) == 0 {
					return fail(ErrDimension)
				}
				wanted.Nodes[p/64] |= 1 << uint(p%64)
			}
		}
		found := -1
		for k := span.Start; k < span.Start+span.Count; k++ {
			if e := charge(); e != nil {
				return fail(e)
			}
			c := cuts.Cuts[k]
			if c.Leaves == wanted.Leaves && c.Nodes == wanted.Nodes {
				found = k
				break
			}
		}
		r.PreferredCuts = append(r.PreferredCuts, found)
	}
	return r, nil
}

// SetMappingPreferences resets phase hints to false then prefers current
// mapping node/cut variables true, matching sat_solver_set_polarity. It changes
// decision preferences only; roots must be supplied when constructing the solver.
func (s *GiaWindowSolver) SetMappingPreferences(state GiaWindowMappingState) error {
	if s == nil || s.solver == nil || len(state.InitialNodes) != len(state.PreferredCuts) {
		return ErrDimension
	}
	vars := make([]SatVar, 0, 2*len(state.InitialNodes))
	for i, node := range state.InitialNodes {
		cut := state.PreferredCuts[i]
		if node < 0 || node >= s.nodes || cut < -1 || cut >= s.cuts {
			return ErrDimension
		}
		vars = append(vars, SatVar(node))
		if cut >= 0 {
			vars = append(vars, SatVar(s.first+cut))
		}
	}
	return s.solver.SetPreferredVariables(vars)
}
