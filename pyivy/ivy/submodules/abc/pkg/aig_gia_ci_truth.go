package abc

// CI support and packed truth from aig/gia/giaTruth.c:
// Gia_ObjComputeTruth6CisSupport_rec and Gia_ObjComputeTruth6Cis.

import "slices"

type GiaCITruthOptions struct {
	MaxObjects        int // reached nonconstant objects, including CIs; zero: one million
	MaxMappingEntries int // arena-indexed scratch capacity; zero: one million
	Stop              func() bool
}

// GiaCITruth6 owns sorted CI object IDs, including register outputs. Word uses
// that input order and repeats the function across all 64 bits. If support
// exceeds six inputs, Fits is false and Word is zero, as in C. Support remains
// complete; a false function with small structural support still has Fits true.
type GiaCITruth6 struct {
	Inputs []GiaID
	Word   uint64
	Fits   bool
}

// CITruth6 collects structural CI support and evaluates a phased root. Unlike
// C's recursive plain-AND implementation, this also handles constants within a
// cone, CO roots, buffers and native XOR/mux nodes. Independent scratch leaves
// source traversal IDs, marks and values unchanged. Limits and cancellation
// return no partial support or truth. Collection is iterative for deep cones.
func (m *GiaManager) CITruth6(root GiaLit, options GiaCITruthOptions) (GiaCITruth6, error) {
	fail := func(err error) (GiaCITruth6, error) { return GiaCITruth6{}, err }
	if m == nil || options.MaxObjects < 0 || options.MaxMappingEntries < 0 {
		return fail(ErrDimension)
	}
	if err := m.Check(); err != nil {
		return fail(err)
	}
	if !root.Valid() || uint64(root.ID()) >= uint64(len(m.objects)) {
		return fail(ErrInvalidLiteral)
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if options.MaxMappingEntries == 0 {
		options.MaxMappingEntries = 1 << 20
	}
	if len(m.objects) > options.MaxMappingEntries {
		return fail(ErrLimit)
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	if stopped() {
		return fail(ErrCancelled)
	}
	seen := make([]bool, len(m.objects))
	seen[0] = true
	stack := []GiaID{root.ID()}
	inputs := make([]GiaID, 0)
	reached := 0
	for len(stack) > 0 {
		if stopped() {
			return fail(ErrCancelled)
		}
		id := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if seen[id] {
			continue
		}
		seen[id] = true
		reached++
		if reached > options.MaxObjects {
			return fail(ErrLimit)
		}
		if m.objects[id].IsCI() {
			inputs = append(inputs, id)
			continue
		}
		children, count, err := m.crossCutFanins(id)
		if err != nil {
			return fail(err)
		}
		for i := count - 1; i >= 0; i-- {
			if !seen[children[i]] {
				stack = append(stack, children[i])
			}
		}
	}
	slices.Sort(inputs)
	if stopped() {
		return fail(ErrCancelled)
	}
	result := GiaCITruth6{Inputs: inputs, Fits: len(inputs) <= 6}
	if !result.Fits {
		return result, nil
	}
	truth, err := m.CutTruth(root, inputs, GiaCutTruthOptions{MaxObjects: options.MaxObjects, MaxMappingEntries: options.MaxMappingEntries, Stop: options.Stop})
	if err != nil {
		return fail(err)
	}
	result.Word = truth.Table.words[0]
	return result, nil
}
