package abc

// Window mapping updates/full evaluation from aig/gia/giaEdge.c:
// Gia_ManUpdateMapping and Gia_ManEvalWindow (mapping2-only context).

import "slices"

// WithWindow returns an owned proposal with replacement fanins in nodes order.
// Empty lists remove LUTs, as in mapping2. Nodes must be unique nonconstant IDs;
// fanins must precede their root and widths are limited to six. Unlike C's
// temporary swap, both fanin and derived fanout lists remain consistent.
// MaxVisits bounds total output fanin slots; MaxObjects bounds the arena.
func (n *GiaLUTNetwork) WithWindow(nodes []GiaID, window [][]GiaID, options GiaLUTAnalysisOptions) (*GiaLUTNetwork, error) {
	if n == nil || len(n.fanins) == 0 || len(nodes) != len(window) || options.MaxObjects < 0 || options.MaxVisits < 0 {
		return nil, ErrDimension
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if options.MaxVisits == 0 {
		options.MaxVisits = 1 << 20
	}
	if len(n.fanins) > options.MaxObjects {
		return nil, ErrLimit
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	if stopped() {
		return nil, ErrCancelled
	}
	replacements := make([]int, len(n.fanins))
	for i := range replacements {
		replacements[i] = -1
	}
	for i, id := range nodes {
		if stopped() {
			return nil, ErrCancelled
		}
		if id == 0 || uint64(id) >= uint64(len(n.fanins)) || replacements[id] >= 0 {
			return nil, ErrDimension
		}
		if len(window[i]) > 6 {
			return nil, ErrUnsupported
		}
		for _, fanin := range window[i] {
			if fanin >= id {
				return nil, ErrDimension
			}
		}
		replacements[id] = i
	}
	result := &GiaLUTNetwork{fanins: make([][]GiaID, len(n.fanins)), fanouts: make([][]GiaID, len(n.fanins))}
	visits := 0
	for id, original := range n.fanins {
		if stopped() {
			return nil, ErrCancelled
		}
		values := original
		if index := replacements[id]; index >= 0 {
			values = window[index]
		}
		if len(values) > options.MaxVisits-visits {
			return nil, ErrLimit
		}
		visits += len(values)
		result.fanins[id] = slices.Clone(values)
		for _, fanin := range values {
			result.fanouts[fanin] = append(result.fanouts[fanin], GiaID(id))
		}
	}
	if stopped() {
		return nil, ErrCancelled
	}
	return result, nil
}

// selectionMapping preserves mapping2's empty-list non-LUT semantics while
// sharing the checked native greedy selection engine with packed mappings.
func (n *GiaLUTNetwork) selectionMapping(options GiaLUTAnalysisOptions) (*GiaLUTMapping, error) {
	return n.PackedMapping(int(^uint32(0)>>1), options)
}

// SelectEdges implements the mapping2-only, no-box greedy timing path.
func (n *GiaLUTNetwork) SelectEdges(m *GiaManager, useTwo bool, options GiaLUTAnalysisOptions) (*GiaEdgeAssignment, GiaEdgeTiming, error) {
	if m == nil || n == nil || len(m.objects) != len(n.fanins) {
		return nil, GiaEdgeTiming{}, ErrDimension
	}
	mapping, err := n.selectionMapping(options)
	if err != nil {
		return nil, GiaEdgeTiming{}, err
	}
	return mapping.SelectEdges(m, useTwo, options)
}

// EvaluateWindow evaluates a temporary mapping2 proposal and returns its owned
// fresh edge assignment/timing. Source network, window and graph stay unchanged.
// As in the native full evaluator, this computes timing and does not establish
// Boolean equivalence of the proposed boundaries. MaxVisits applies separately
// to output storage and the selection pass. Incremental evaluation is separate.
func (n *GiaLUTNetwork) EvaluateWindow(m *GiaManager, nodes []GiaID, window [][]GiaID, useTwo bool, options GiaLUTAnalysisOptions) (*GiaEdgeAssignment, GiaEdgeTiming, error) {
	if m == nil || n == nil || len(m.objects) != len(n.fanins) {
		return nil, GiaEdgeTiming{}, ErrDimension
	}
	proposal, err := n.WithWindow(nodes, window, options)
	if err != nil {
		return nil, GiaEdgeTiming{}, err
	}
	return proposal.SelectEdges(m, useTwo, options)
}
