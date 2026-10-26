package abc

// Multi-pivot reconstruction and signal transfer from src/aig/gia/giaCof.c:
// Gia_ManDupCofAllInt and Gia_ManTransfer.

import "slices"

type GiaMultiCofactorOptions struct {
	GiaCofactorOptions
	// MaxPasses zero selects 200. ABC also caps the initial signal list at 200.
	MaxPasses int
}

// CofactorRebuildAll repeatedly reconstructs the highest-ID remaining pivot,
// cleans the graph and transfers remaining signals through owned mappings.
// The result is equivalent to source, preserving CI/CO order and sequential
// metadata. Cancellation/resource failure returns no partial graph. An empty
// signal list returns an independent clone, matching C.
func (m *GiaManager) CofactorRebuildAll(signals []GiaID, options GiaMultiCofactorOptions) (*GiaManager, error) {
	if options.MaxPasses < 0 || options.MaxObjects < 0 || options.MaxMappingEntries < 0 {
		return nil, ErrDimension
	}
	if options.MaxPasses == 0 {
		options.MaxPasses = 200
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if options.MaxMappingEntries == 0 {
		options.MaxMappingEntries = 1 << 20
	}
	if err := m.Check(); err != nil {
		return nil, err
	}
	if len(signals) > 200 || len(m.objects) > options.MaxObjects || len(m.objects) > options.MaxMappingEntries {
		return nil, ErrLimit
	}
	for _, id := range signals {
		if id == 0 || uint64(id) >= uint64(len(m.objects)) || (!m.objects[id].IsCI() && !m.objects[id].IsAnd()) {
			return nil, ErrInvalidLiteral
		}
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	if stopped() {
		return nil, ErrCancelled
	}
	graph := m.Clone()
	remaining := slices.Clone(signals)
	for pass := 0; len(remaining) > 0; pass++ {
		if stopped() {
			return nil, ErrCancelled
		}
		if pass == options.MaxPasses {
			return nil, ErrLimit
		}
		slices.Sort(remaining)
		pivot := remaining[len(remaining)-1]
		remaining = remaining[:len(remaining)-1]
		rebuilt, copies, err := graph.CofactorRebuild(pivot, options.GiaCofactorOptions)
		if err != nil {
			return nil, err
		}
		if stopped() {
			return nil, ErrCancelled
		}
		if rebuilt.ObjectCount() > options.MaxMappingEntries {
			return nil, ErrLimit
		}
		cleaned, remap, err := rebuilt.Cleanup()
		if err != nil {
			return nil, err
		}
		if stopped() {
			return nil, ErrCancelled
		}
		next := make([]GiaID, 0, len(remaining))
		seen := make(map[GiaID]bool)
		for _, id := range remaining {
			copy := copies[id]
			if copy == InvalidGiaLit {
				continue
			}
			if uint64(copy.ID()) >= uint64(len(remap)) {
				return nil, ErrInvariant
			}
			mapped := remap[copy.ID()]
			if mapped == InvalidGiaLit || mapped == GiaConst0 {
				continue
			}
			if !mapped.Valid() || uint64(mapped.ID()) >= uint64(cleaned.ObjectCount()) {
				return nil, ErrInvariant
			}
			candidate := mapped.ID()
			// Constants no longer require cofactoring, regardless of phase. C's
			// ID-only transfer can otherwise enqueue object zero as an invalid pivot.
			if candidate == 0 || seen[candidate] {
				continue
			}
			if !cleaned.objects[candidate].IsAnd() && !cleaned.objects[candidate].IsCI() {
				return nil, ErrInvariant
			}
			seen[candidate] = true
			next = append(next, candidate)
		}
		graph, remaining = cleaned, next
	}
	return graph, nil
}
