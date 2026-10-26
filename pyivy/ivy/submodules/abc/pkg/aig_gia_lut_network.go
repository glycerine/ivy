package abc

// Alternate mapping and unit timing from aig/gia/giaEdge.c: Edg_ManToMapping,
// Edg_ObjEvalEdgeDelay[R], Edg_ManEvalEdgeDelay[R], Edg_ManCollectCritEdges.

import "slices"

type GiaLUTNetwork struct{ fanins, fanouts [][]GiaID }

// Network converts up to four fanin slots per LUT. MaxVisits bounds stored
// fanin slots; MaxObjects bounds the arena. Source storage remains unchanged.
func (mapping *GiaLUTMapping) Network(options GiaLUTAnalysisOptions) (*GiaLUTNetwork, error) {
	return mapping.alternateNetwork(options, 4)
}

func (mapping *GiaLUTMapping) alternateNetwork(options GiaLUTAnalysisOptions, maxWidth int) (*GiaLUTNetwork, error) {
	if mapping == nil || mapping.objects < 1 || options.MaxObjects < 0 || options.MaxVisits < 0 {
		return nil, ErrDimension
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if options.MaxVisits == 0 {
		options.MaxVisits = 1 << 20
	}
	if mapping.objects > options.MaxObjects {
		return nil, ErrLimit
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	if stopped() {
		return nil, ErrCancelled
	}
	n := &GiaLUTNetwork{fanins: make([][]GiaID, mapping.objects), fanouts: make([][]GiaID, mapping.objects)}
	visits := 0
	for _, root := range mapping.Roots() {
		if stopped() {
			return nil, ErrCancelled
		}
		offset := mapping.packed[root]
		count := mapping.packed[offset]
		if maxWidth > 0 && count > maxWidth {
			return nil, ErrUnsupported
		}
		if count > options.MaxVisits-visits {
			return nil, ErrLimit
		}
		visits += count
		n.fanins[root] = make([]GiaID, count)
		for i, id := range mapping.packed[offset+1 : offset+1+count] {
			if stopped() {
				return nil, ErrCancelled
			}
			n.fanins[root][i] = GiaID(id)
			n.fanouts[id] = append(n.fanouts[id], root)
		}
	}
	if stopped() {
		return nil, ErrCancelled
	}
	return n, nil
}
func (n *GiaLUTNetwork) Roots() []GiaID {
	if n == nil {
		return nil
	}
	r := make([]GiaID, 0)
	for id := 1; id < len(n.fanins); id++ {
		if len(n.fanins[id]) > 0 {
			r = append(r, GiaID(id))
		}
	}
	return r
}
func (n *GiaLUTNetwork) Fanins(id GiaID) ([]GiaID, error) {
	if n == nil || uint64(id) >= uint64(len(n.fanins)) {
		return nil, ErrDimension
	}
	return slices.Clone(n.fanins[id]), nil
}
func (n *GiaLUTNetwork) Fanouts(id GiaID) ([]GiaID, error) {
	if n == nil || uint64(id) >= uint64(len(n.fanouts)) {
		return nil, ErrDimension
	}
	return slices.Clone(n.fanouts[id]), nil
}

// GiaUnitEdgeTiming retains the global maximum over LUTs, including dangling
// LUTs. Reverse arrays evaluate LUT fanouts only; CI reverse entries stay zero.
// Marked edges cost zero, ordinary edges cost one, as in Edg_ManEvalEdgeDelay.
type GiaUnitEdgeTiming struct {
	Forward, Reverse        []int
	Maximum, ReverseMaximum int
}

// EvaluateEdges returns owned distances. MaxVisits bounds the combined
// forward fanin and reverse mapped-fanout slot visits across both passes.
func (n *GiaLUTNetwork) EvaluateEdges(e *GiaEdgeAssignment, options GiaLUTAnalysisOptions) (GiaUnitEdgeTiming, error) {
	fail := func(err error) (GiaUnitEdgeTiming, error) { return GiaUnitEdgeTiming{}, err }
	if n == nil || len(n.fanins) == 0 || e == nil || len(e.rows) != len(n.fanins) || options.MaxObjects < 0 || options.MaxVisits < 0 {
		return fail(ErrDimension)
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if options.MaxVisits == 0 {
		options.MaxVisits = 1 << 20
	}
	if len(n.fanins) > options.MaxObjects {
		return fail(ErrLimit)
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	if stopped() {
		return fail(ErrCancelled)
	}
	r := GiaUnitEdgeTiming{Forward: make([]int, len(n.fanins)), Reverse: make([]int, len(n.fanins))}
	visits := 0
	roots := n.Roots()
	evaluate := func(root GiaID, neighbors []GiaID, distances []int) (int, error) {
		result := 0
		for _, id := range neighbors {
			if stopped() {
				return 0, ErrCancelled
			}
			if visits == options.MaxVisits {
				return 0, ErrLimit
			}
			visits++
			step := 1
			if e.rows[root][0] == id || e.rows[root][1] == id {
				step = 0
			}
			value, err := checkedSize(distances[id], 1, step, int(^uint(0)>>1))
			if err != nil {
				return 0, err
			}
			result = max(result, value)
		}
		return result, nil
	}
	for _, root := range roots {
		if stopped() {
			return fail(ErrCancelled)
		}
		value, err := evaluate(root, n.fanins[root], r.Forward)
		if err != nil {
			return fail(err)
		}
		r.Forward[root] = value
		r.Maximum = max(r.Maximum, value)
	}
	for i := len(roots) - 1; i >= 0; i-- {
		if stopped() {
			return fail(ErrCancelled)
		}
		root := roots[i]
		value, err := evaluate(root, n.fanouts[root], r.Reverse)
		if err != nil {
			return fail(err)
		}
		r.Reverse[root] = value
		r.ReverseMaximum = max(r.ReverseMaximum, value)
	}
	if stopped() {
		return fail(ErrCancelled)
	}
	return r, nil
}

// CriticalLevels follows native LUT iteration and F+R==Maximum grouping.
// Inconsistent timing (including incompatible asymmetric edge weights) rejects
// the native F+R<=Maximum and nonempty-positive-level assertions safely.
func (n *GiaLUTNetwork) CriticalLevels(timing GiaUnitEdgeTiming) ([][]GiaID, error) {
	if n == nil || len(n.fanins) == 0 || len(timing.Forward) != len(n.fanins) || len(timing.Reverse) != len(n.fanins) || timing.Maximum < 0 || timing.Maximum > len(n.fanins) {
		return nil, ErrDimension
	}
	levels := make([][]GiaID, timing.Maximum+1)
	for _, root := range n.Roots() {
		f, r := timing.Forward[root], timing.Reverse[root]
		if f < 0 || r < 0 || f > timing.Maximum || r > timing.Maximum-f {
			return nil, ErrInvariant
		}
		if f+r == timing.Maximum {
			levels[f] = append(levels[f], root)
		}
	}
	for level := 1; level < len(levels); level++ {
		if len(levels[level]) == 0 {
			return nil, ErrInvariant
		}
	}
	return levels, nil
}
