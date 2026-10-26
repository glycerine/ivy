package abc

// Checked LUT functions and application complement giaSatLut.c:Sbl_ManUpdateMapping.

type GiaLUTFunction struct {
	Root  GiaID
	Truth uint64 // six-variable packed word; duplicate fanins use their last slot
}

// ValidateFunctions checks complete mapped dependencies, output coverage and
// structural cut coverage for every LUT, including dangling LUTs. Functions
// are derived from the source graph, so composing them preserves its outputs.
// Constants are fixed; buffers, XORs and mux controls are included. MaxVisits
// bounds arena scans, fanin slots and aggregate structural traversal work.
func (n *GiaLUTNetwork) ValidateFunctions(m *GiaManager, o GiaLUTAnalysisOptions) ([]GiaLUTFunction, error) {
	if n == nil || m == nil || len(n.fanins) != len(m.objects) || o.MaxObjects < 0 || o.MaxVisits < 0 {
		return nil, ErrDimension
	}
	if o.MaxObjects == 0 {
		o.MaxObjects = 1 << 20
	}
	if o.MaxVisits == 0 {
		o.MaxVisits = 1 << 20
	}
	if len(m.objects) > o.MaxObjects {
		return nil, ErrLimit
	}
	if e := m.Check(); e != nil {
		return nil, e
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
	covered := func(id GiaID) bool {
		return id == 0 || m.objects[id].IsCI() || m.objects[id].IsAnd() && len(n.fanins[id]) > 0
	}
	for id, fanins := range n.fanins {
		if e := charge(); e != nil {
			return nil, e
		}
		if len(fanins) == 0 {
			continue
		}
		if !m.objects[id].IsAnd() || len(fanins) > 6 {
			return nil, ErrDimension
		}
		for _, fanin := range fanins {
			if e := charge(); e != nil {
				return nil, e
			}
			if int(fanin) >= id || !covered(fanin) {
				return nil, ErrDimension
			}
		}
	}
	for _, id := range m.cos {
		if e := charge(); e != nil {
			return nil, e
		}
		l, e := m.objects[id].Fanin0(id)
		if e != nil {
			return nil, e
		}
		if !covered(l.ID()) {
			return nil, ErrDimension
		}
	}
	done, active := make([]int, len(m.objects)), make([]int, len(m.objects))
	values := make([]uint64, len(m.objects))
	type frame struct {
		id    GiaID
		ready bool
	}
	var stack []frame
	var result []GiaLUTFunction
	elementary := [6]uint64{0xaaaaaaaaaaaaaaaa, 0xcccccccccccccccc, 0xf0f0f0f0f0f0f0f0, 0xff00ff00ff00ff00, 0xffff0000ffff0000, 0xffffffff00000000}
	read := func(l GiaLit) uint64 {
		v := values[l.ID()]
		if l.Complemented() {
			v = ^v
		}
		return v
	}
	for root, fanins := range n.fanins {
		if len(fanins) == 0 {
			continue
		}
		epoch := root + 1
		done[0] = epoch
		values[0] = 0
		for slot, id := range fanins {
			if id != 0 {
				done[id] = epoch
				values[id] = elementary[slot]
			}
		}
		stack = append(stack[:0], frame{id: GiaID(root)})
		for len(stack) > 0 {
			if e := charge(); e != nil {
				return nil, e
			}
			f := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if done[f.id] == epoch {
				continue
			}
			obj := m.objects[f.id]
			if !obj.IsAnd() {
				return nil, ErrDimension
			}
			a, e := obj.Fanin0(f.id)
			if e != nil {
				return nil, e
			}
			b := GiaConst0
			if !obj.IsBuffer() {
				b, e = obj.Fanin1(f.id)
				if e != nil {
					return nil, e
				}
			}
			control, mux := m.MuxControl(f.id)
			if f.ready {
				v := read(a)
				if !obj.IsBuffer() {
					second := read(b)
					if mux {
						c := read(control)
						v = c&second | ^c&v
					} else if obj.IsXor() {
						v ^= second
					} else {
						v &= second
					}
				}
				values[f.id] = v
				done[f.id] = epoch
				continue
			}
			if active[f.id] == epoch {
				continue
			}
			active[f.id] = epoch
			stack = append(stack, frame{id: f.id, ready: true})
			if mux {
				stack = append(stack, frame{id: control.ID()})
			}
			if !obj.IsBuffer() {
				stack = append(stack, frame{id: b.ID()})
			}
			stack = append(stack, frame{id: a.ID()})
		}
		result = append(result, GiaLUTFunction{Root: GiaID(root), Truth: values[root]})
	}
	if o.Stop != nil && o.Stop() {
		return nil, ErrCancelled
	}
	return result, nil
}

type GiaValidatedUnitWindow struct {
	Network   *GiaLUTNetwork
	Functions []GiaLUTFunction
	Timing    GiaLUTUnitTiming
}

// WithValidatedUnitWindow returns a complete owned replacement after checking
// source/proposal functions, global dependency closure and global unit delay.
// It rejects proposals whose maximum exceeds the original reference maximum,
// even if a caller omitted an escaping window root. Limits apply per phase;
// errors/cancellation expose no replacement and never mutate the source.
func (n *GiaLUTNetwork) WithValidatedUnitWindow(m *GiaManager, nodes []GiaID, window [][]GiaID, delayStart int, o GiaLUTAnalysisOptions) (GiaValidatedUnitWindow, error) {
	return n.withValidatedUnitWindow(m, nodes, window, delayStart, o, true)
}

func (n *GiaLUTNetwork) withValidatedUnitWindow(m *GiaManager, nodes []GiaID, window [][]GiaID, delayStart int, o GiaLUTAnalysisOptions, enforceTiming bool) (GiaValidatedUnitWindow, error) {
	fail := func(e error) (GiaValidatedUnitWindow, error) { return GiaValidatedUnitWindow{}, e }
	if _, e := n.ValidateFunctions(m, o); e != nil {
		return fail(e)
	}
	original, e := n.UnitTiming(m, delayStart, o)
	if e != nil {
		return fail(e)
	}
	proposal, e := n.WithWindow(nodes, window, o)
	if e != nil {
		return fail(e)
	}
	functions, e := proposal.ValidateFunctions(m, o)
	if e != nil {
		return fail(e)
	}
	timing, e := proposal.UnitTiming(m, delayStart, o)
	if e != nil {
		return fail(e)
	}
	if enforceTiming && timing.Maximum > original.Maximum {
		return fail(ErrDimension)
	}
	return GiaValidatedUnitWindow{Network: proposal, Functions: functions, Timing: timing}, nil
}

type GiaUnitWindowRemapResult struct {
	Search      GiaUnitWindowSearchResult
	Replacement GiaValidatedUnitWindow // Network is nil when no improvement was found
}

// RemapUnitWindow searches and validates an explicit unboxed window. A
// successful improvement includes an owned complete replacement network and
// its LUT functions; the caller selects that replacement without mutating the
// previous network. Automatic window discovery remains a separate layer.
func (n *GiaLUTNetwork) RemapUnitWindow(m *GiaManager, leaves, nodes, roots []GiaID, delayStart int, o GiaUnitWindowSearchOptions) (GiaUnitWindowRemapResult, error) {
	return n.remapUnitWindow(m, leaves, nodes, roots, delayStart, o, nil)
}

func (n *GiaLUTNetwork) remapUnitWindow(m *GiaManager, leaves, nodes, roots []GiaID, delayStart int, o GiaUnitWindowSearchOptions, session *giaUnitWindowSolverSession) (GiaUnitWindowRemapResult, error) {
	search, e := n.searchUnitWindow(m, leaves, nodes, roots, delayStart, o, session)
	if e != nil {
		return GiaUnitWindowRemapResult{}, e
	}
	r := GiaUnitWindowRemapResult{Search: search}
	if !search.Improved {
		return r, nil
	}
	analysis := o.Analysis
	prior := analysis.Stop
	analysis.Stop = func() bool { return o.Stop != nil && o.Stop() || prior != nil && prior() }
	r.Replacement, e = n.withValidatedUnitWindow(m, nodes, search.BestWindow, delayStart, analysis, !o.AreaOnly)
	if e != nil {
		return GiaUnitWindowRemapResult{}, e
	}
	return r, nil
}
