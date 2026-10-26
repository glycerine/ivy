package abc

// Mapped references and overlap from aig/gia/giaIf.c:
// Gia_ManSetRefsMapped/SetLutRefs and Gia_ManComputeOverlap[2][One][_rec].

type GiaLUTAnalysisOptions struct {
	MaxObjects int // full graph arena; zero: one million
	MaxVisits  int // total unique non-leaf objects across all LUT cones; zero: one million
	Stop       func() bool
}

// GiaLUTAnalysis owns mapped references and per-root structural cone sizes.
// References include each CO edge and each mapped fanin slot, including
// duplicates and unused LUTs. Overlap is sum(ConeSizes)-source AND count; it can
// be negative for partial mappings. Constant mapping records are skipped.
type GiaLUTAnalysis struct {
	References []int
	ConeSizes  []int
	Overlap    int
}

// Analyze counts structural nodes inside each LUT boundary. It preserves C's
// shared-node-once-per-cone counting, without recursion or source traversal
// marks. Missing CI boundaries reject instead of asserting. Constants remain
// fixed; buffers and native XOR/mux are safe extensions. Limits/cancellation
// return no partial arrays. References have the same values in C's pRefs and
// pLutRefs representations, so one owned array represents both.
func (mapping *GiaLUTMapping) Analyze(m *GiaManager, options GiaLUTAnalysisOptions) (GiaLUTAnalysis, error) {
	fail := func(err error) (GiaLUTAnalysis, error) { return GiaLUTAnalysis{}, err }
	if m == nil || mapping == nil || options.MaxObjects < 0 || options.MaxVisits < 0 {
		return fail(ErrDimension)
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if options.MaxVisits == 0 {
		options.MaxVisits = 1 << 20
	}
	if len(m.objects) > options.MaxObjects {
		return fail(ErrLimit)
	}
	if err := mapping.checkGraph(m); err != nil {
		return fail(err)
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	if stopped() {
		return fail(ErrCancelled)
	}
	result := GiaLUTAnalysis{References: make([]int, len(m.objects)), ConeSizes: make([]int, len(m.objects))}
	for _, id := range m.cos {
		if stopped() {
			return fail(ErrCancelled)
		}
		lit, err := m.objects[id].Fanin0(id)
		if err != nil {
			return fail(err)
		}
		result.References[lit.ID()]++
	}
	marks := make([]uint32, len(m.objects))
	stack := make([]GiaID, 0)
	visits := 0
	for _, obj := range m.objects {
		if obj.IsAnd() {
			result.Overlap--
		}
	}
	for _, root := range mapping.Roots() {
		if stopped() {
			return fail(ErrCancelled)
		}
		epoch := uint32(root) + 1
		marks[0] = epoch
		offset := mapping.packed[root]
		count := mapping.packed[offset]
		for _, fanin := range mapping.packed[offset+1 : offset+1+count] {
			result.References[fanin]++
			marks[fanin] = epoch
		}
		stack = append(stack[:0], root)
		for len(stack) > 0 {
			if stopped() {
				return fail(ErrCancelled)
			}
			id := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if marks[id] == epoch {
				continue
			}
			marks[id] = epoch
			if m.objects[id].IsCI() {
				return fail(ErrDimension)
			}
			if visits == options.MaxVisits {
				return fail(ErrLimit)
			}
			visits++
			result.ConeSizes[root]++
			result.Overlap++
			children, n, err := m.crossCutFanins(id)
			if err != nil {
				return fail(err)
			}
			for i := n - 1; i >= 0; i-- {
				if marks[children[i]] != epoch {
					stack = append(stack, children[i])
				}
			}
		}
	}
	if stopped() {
		return fail(ErrCancelled)
	}
	return result, nil
}
