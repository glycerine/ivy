package abc

// Top-level conjunction reduction from src/aig/gia/giaDup.c:
// Gia_ManDupTopAnd_iter and Gia_ManDupTopAnd.

// GiaTopAndReduction records CI assignments removed from the conjunctive query.
// Fixed contains InvalidGiaLit (unfixed), GiaConst0 or GiaConst1, by original
// CI ordinal. Original conjunction equals fixed-assignment satisfaction AND
// the result's conjunction. Contradiction means the original conjunction is
// false. These are all-CO conjunctive-query semantics, not sequential safety.
type GiaTopAndReduction struct {
	Graph         *GiaManager
	Fixed         []GiaLit
	Contradiction bool
	Iterations    int
}

type GiaTopAndOptions struct {
	GiaCofactorOptions
	MaxIterations int
}

// ReduceTopAndOnce expands positive ordinary AND roots, fixes CI boundaries,
// and emits distinct non-CI boundaries in source-ID order. Negative edges and
// extended gates are opaque. All CIs remain in original order. With fixed CIs,
// remaining cones hash through substitution and source register count survives
// if enough outputs remain; otherwise copied signal cones become combinational,
// matching C. Actual reduction clears constraints; unchanged fallback retains
// roles. Contradictions return an owned normalized fallback. Unlike C, a false
// constant CO is never discarded when other roots expand. Source scratch stays
// unchanged and name/spec survive. Invalid positional state layouts reject.
func (m *GiaManager) ReduceTopAndOnce(options GiaCofactorOptions) (GiaTopAndReduction, error) {
	fail := func(e error) (GiaTopAndReduction, error) { return GiaTopAndReduction{}, e }
	if m == nil || options.MaxObjects < 0 || options.MaxMappingEntries < 0 {
		return fail(ErrDimension)
	}
	if e := m.Check(); e != nil {
		return fail(e)
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
	fixed := make([]GiaLit, len(m.cis))
	for i := range fixed {
		fixed[i] = InvalidGiaLit
	}
	phases := make([]uint8, len(m.objects))
	contradiction := false
	add := func(l GiaLit) {
		if l.ID() == 0 {
			if l == GiaConst0 {
				contradiction = true
			}
			return
		}
		phase := uint8(1)
		if l.Complemented() {
			phase = 2
		}
		if phases[l.ID()] != 0 && phases[l.ID()] != phase {
			contradiction = true
		}
		phases[l.ID()] = phase
	}
	front := []GiaID{}
	for _, co := range m.cos {
		if stopped() {
			return fail(ErrCancelled)
		}
		d, e := m.objects[co].Fanin0(co)
		if e != nil {
			return fail(e)
		}
		if d.ID() == 0 {
			add(d)
		} else if d.Complemented() {
			add(d)
		} else {
			front = append(front, d.ID())
		}
	}
	fallback := func() (GiaTopAndReduction, error) {
		g, _, e := m.NormalizeGraph(false, options)
		if e != nil {
			return fail(e)
		}
		return GiaTopAndReduction{g, fixed, contradiction, 1}, nil
	}
	if len(front) == 0 || contradiction {
		return fallback()
	}
	expanded := make([]bool, len(m.objects))
	for i := 0; i < len(front); i++ {
		if stopped() {
			return fail(ErrCancelled)
		}
		id := front[i]
		if expanded[id] {
			continue
		}
		expanded[id] = true
		o := m.objects[id]
		_, mux := m.MuxControl(id)
		if !o.IsAnd() || o.IsXor() || o.IsBuffer() || mux {
			add(GiaLit(id) << 1)
			continue
		}
		for k := 0; k < 2; k++ {
			var l GiaLit
			var e error
			if k == 0 {
				l, e = o.Fanin0(id)
			} else {
				l, e = o.Fanin1(id)
			}
			if e != nil {
				return fail(e)
			}
			if l.ID() == 0 || l.Complemented() {
				add(l)
			} else {
				front = append(front, l.ID())
			}
		}
	}
	if contradiction {
		return fallback()
	}
	leaves := []GiaLit{}
	fixedCount := 0
	for i, phase := range phases {
		if phase == 0 {
			continue
		}
		if stopped() {
			return fail(ErrCancelled)
		}
		id := GiaID(i)
		l := GiaLit(id << 1).NotIf(phase == 2)
		if m.objects[i].IsCI() {
			ordinal, e := m.objects[i].TerminalIndex()
			if e != nil {
				return fail(e)
			}
			fixed[ordinal] = GiaConst1.NotIf(l.Complemented())
			fixedCount++
		} else {
			leaves = append(leaves, l)
		}
	}
	var g *GiaManager
	var e error
	if fixedCount == 0 {
		g, e = m.ExtractSignalsRetainingInputs(leaves, options)
	} else {
		g, e = m.RemapInputs(fixed, leaves, options)
	}
	if e != nil {
		return fail(e)
	}
	return GiaTopAndReduction{g, fixed, false, 1}, nil
}

// ReduceTopAnd normalizes then repeats ReduceTopAndOnce until CO and gate
// counts stop changing, as in C. Fixed assignments accumulate by stable CI
// ordinal. MaxIterations zero selects one million; hitting it returns ErrLimit
// rather than an incomplete fixed-point result. No SAT/UNSAT proof is inferred
// beyond explicit contradictory conjunctive literals/constants.
func (m *GiaManager) ReduceTopAnd(options GiaTopAndOptions) (GiaTopAndReduction, error) {
	fail := func(e error) (GiaTopAndReduction, error) { return GiaTopAndReduction{}, e }
	if options.MaxIterations < 0 {
		return fail(ErrDimension)
	}
	capIterations := options.MaxIterations
	if capIterations == 0 {
		capIterations = 1 << 20
	}
	g, _, e := m.NormalizeGraph(false, options.GiaCofactorOptions)
	if e != nil {
		return fail(e)
	}
	fixed := make([]GiaLit, len(g.cis))
	for i := range fixed {
		fixed[i] = InvalidGiaLit
	}
	for iteration := 0; iteration < capIterations; iteration++ {
		r, err := g.ReduceTopAndOnce(options.GiaCofactorOptions)
		if err != nil {
			return fail(err)
		}
		contradiction := r.Contradiction
		for i, l := range r.Fixed {
			if l == InvalidGiaLit {
				continue
			}
			if fixed[i] != InvalidGiaLit && fixed[i] != l {
				contradiction = true
			}
			fixed[i] = l
		}
		done := r.Graph.OutputCount() == g.OutputCount() && r.Graph.AndCount() == g.AndCount()
		if contradiction || done {
			return GiaTopAndReduction{r.Graph, fixed, contradiction, iteration + 1}, nil
		}
		g = r.Graph
	}
	return fail(ErrLimit)
}
