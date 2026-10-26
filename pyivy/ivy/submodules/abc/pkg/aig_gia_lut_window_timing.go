package abc

// Decoded-window timing evaluation from aig/gia/giaSatLut.c:
// Sbl_ManEvaluateMapping (unboxed/unit branch) and Sbl_ManCriticalFanin.

type GiaUnitWindowEvaluation struct {
	Accepted                bool
	ReferenceMaximum        int
	Arrivals, Required      []int    // proposed local arrivals; original required constraints
	ViolatingRoot, Boundary GiaID    // InvalidGiaID on acceptance
	Path                    []GiaID  // violating root toward the first external critical fanin
	Clause                  []SatLit // negated node-selection variables, indexed by nodes order
}

// EvaluateUnitWindow checks an already decoded proposal against original unit
// required times. nodes must be unique gates in proposal topological order; roots must refer
// to nodes. Empty replacement lists produce zero local arrival, as in C.
// First-root and first-critical-fanin order retain native conflict clauses.
// Boolean validity and candidate-cut closure are caller obligations, enforced
// separately by the SAT remapper. This method does not change the mapping.
// MaxVisits combines original timing, proposal slots and conflict-path steps.
func (n *GiaLUTNetwork) EvaluateUnitWindow(m *GiaManager, nodes []GiaID, window [][]GiaID, roots []GiaID, delayStart int, options GiaLUTAnalysisOptions) (GiaUnitWindowEvaluation, error) {
	return n.evaluateUnitWindow(m, nodes, window, roots, delayStart, options, nil)
}

func (n *GiaLUTNetwork) evaluateUnitWindow(m *GiaManager, nodes []GiaID, window [][]GiaID, roots []GiaID, delayStart int, options GiaLUTAnalysisOptions, prepared *GiaLUTUnitTiming) (GiaUnitWindowEvaluation, error) {
	fail := func(err error) (GiaUnitWindowEvaluation, error) { return GiaUnitWindowEvaluation{}, err }
	if n == nil || m == nil || len(n.fanins) != len(m.objects) || len(nodes) != len(window) || options.MaxObjects < 0 || options.MaxVisits < 0 {
		return fail(ErrDimension)
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if options.MaxVisits == 0 {
		options.MaxVisits = 1 << 20
	}
	if len(m.objects) > options.MaxObjects || len(nodes) > options.MaxObjects || len(roots) > options.MaxObjects {
		return fail(ErrLimit)
	}
	if err := m.Check(); err != nil {
		return fail(err)
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	if stopped() {
		return fail(ErrCancelled)
	}
	indices := make([]int, len(m.objects))
	for i := range indices {
		indices[i] = -1
	}
	for i, id := range nodes {
		if stopped() {
			return fail(ErrCancelled)
		}
		if uint64(id) >= uint64(len(m.objects)) || !m.objects[id].IsAnd() || indices[id] >= 0 {
			return fail(ErrDimension)
		}
		if len(window[i]) > 6 {
			return fail(ErrUnsupported)
		}
		indices[id] = i
		for _, fanin := range window[i] {
			if fanin >= id {
				return fail(ErrDimension)
			}
		}
	}
	for i, fanins := range window {
		for _, fanin := range fanins {
			if indices[fanin] >= i {
				return fail(ErrDimension)
			}
		}
	}
	for _, id := range roots {
		if uint64(id) >= uint64(len(indices)) || indices[id] < 0 {
			return fail(ErrDimension)
		}
	}
	work := 0
	var original GiaLUTUnitTiming
	var err error
	if prepared == nil {
		for _, id := range n.Roots() {
			work, err = checkedSize(work, 2, len(n.fanins[id]), options.MaxVisits)
			if err != nil {
				return fail(err)
			}
		}
		original, err = n.UnitTiming(m, delayStart, options)
		if err != nil {
			return fail(err)
		}
	} else {
		original = *prepared
		if len(original.Arrivals) != len(m.objects) || len(original.Required) != len(m.objects) {
			return fail(ErrDimension)
		}
	}
	charge := func() error {
		if stopped() {
			return ErrCancelled
		}
		if work == options.MaxVisits {
			return ErrLimit
		}
		work++
		return nil
	}
	r := GiaUnitWindowEvaluation{ReferenceMaximum: original.Maximum, Arrivals: original.Arrivals, Required: original.Required, ViolatingRoot: InvalidGiaID, Boundary: InvalidGiaID}
	for i, id := range nodes {
		if stopped() {
			return fail(ErrCancelled)
		}
		r.Arrivals[id] = 0
		for _, fanin := range window[i] {
			if err = charge(); err != nil {
				return fail(err)
			}
			value, err := giaEdgeDelayAdd(r.Arrivals[fanin], 1)
			if err != nil {
				return fail(err)
			}
			r.Arrivals[id] = max(r.Arrivals[id], value)
		}
	}
	for _, id := range roots {
		if stopped() {
			return fail(ErrCancelled)
		}
		if r.Arrivals[id] > r.Required[id] {
			r.ViolatingRoot = id
			break
		}
	}
	if r.ViolatingRoot == InvalidGiaID {
		r.Accepted = true
		return r, nil
	}
	current := r.ViolatingRoot
	for {
		if err = charge(); err != nil {
			return fail(err)
		}
		index := indices[current]
		if index < 0 {
			return fail(ErrInvariant)
		}
		literal, err := MakeSatLit(SatVar(index), true)
		if err != nil {
			return fail(err)
		}
		r.Path = append(r.Path, current)
		r.Clause = append(r.Clause, literal)
		critical := InvalidGiaID
		for _, fanin := range window[index] {
			if err = charge(); err != nil {
				return fail(err)
			}
			arrival, err := giaEdgeDelayAdd(r.Arrivals[fanin], 1)
			if err != nil {
				return fail(err)
			}
			if arrival == r.Arrivals[current] {
				critical = fanin
				break
			}
		}
		if critical == InvalidGiaID || critical == 0 {
			return fail(ErrInvariant)
		}
		if indices[critical] < 0 {
			r.Boundary = critical
			break
		}
		current = critical
	}
	if stopped() {
		return fail(ErrCancelled)
	}
	return r, nil
}
