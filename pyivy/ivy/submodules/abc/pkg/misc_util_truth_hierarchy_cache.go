package abc

// Hierarchical representative storage from opt/dau/dauCanon.c:
// Abc_TtHieManStart/Stop and Abc_TtHieRetrieveOrInsert.

import "slices"

type TruthHierarchyStatus int

const (
	TruthHierarchyUnavailable TruthHierarchyStatus = -1
	TruthHierarchyInserted    TruthHierarchyStatus = 0
	TruthHierarchyHit         TruthHierarchyStatus = 1
)

// TruthHierarchyLookup owns a representative on a hit or final-level visit.
// Other visits return nil Representative and RepresentativeID -1. Like C, a
// final-level lookup returns Inserted even when that final truth already exists.
type TruthHierarchyLookup struct {
	Status           TruthHierarchyStatus
	Representative   *TruthTable
	RepresentativeID int
}
type truthHierarchyLevel struct {
	memory          *WordMemory
	representatives []int
}

// TruthHierarchyCache stores at most five normalization stages, with stable
// insertion IDs per level. Begin starts a traversal; a hit or final insertion
// resolves visited earlier stages and ends it. Abandoned pending stages remain
// reusable after Begin, unlike C's assumption that all paths complete. Use one
// traversal at a time; external synchronization is required for shared use.
// WordMemory lookup maps are never iterated to determine IDs or output order.
type TruthHierarchyCache struct {
	variables      int
	levels         []truthHierarchyLevel
	current        []int
	active, closed bool
}

// NewTruthHierarchyCache bounds each level using WordMemoryOptions. Zero levels
// disables caching. Pages adapt to entry width and configured bounds, up to C's
// 4096 entries per page, rather than requiring a huge fixed page for wide truth.
func NewTruthHierarchyCache(variables, levels int, options WordMemoryOptions) (*TruthHierarchyCache, error) {
	if levels < 0 || levels > 5 || options.MaxEntries < 0 || options.MaxWords < 0 {
		return nil, ErrDimension
	}
	words, e := TruthTableWordCount(variables)
	if e != nil {
		return nil, e
	}
	maxEntries, maxWords := options.MaxEntries, options.MaxWords
	if maxEntries == 0 {
		maxEntries = 1 << 20
	}
	if maxWords == 0 {
		maxWords = 1 << 20
	}
	if words > maxWords {
		return nil, ErrLimit
	}
	pageLog := 0
	for pageLog < 12 && (1<<uint(pageLog+1)) <= maxEntries && (1<<uint(pageLog+1)) <= maxWords/words {
		pageLog++
	}
	c := &TruthHierarchyCache{variables: variables, levels: make([]truthHierarchyLevel, levels), current: make([]int, levels)}
	for i := range c.levels {
		memory, e := NewWordMemory(words, pageLog, options)
		if e != nil {
			return nil, e
		}
		if e = memory.EnableHash(); e != nil {
			return nil, e
		}
		c.levels[i].memory = memory
		c.current[i] = -1
	}
	return c, nil
}
func (c *TruthHierarchyCache) Begin() error {
	if c == nil || c.closed {
		return ErrDimension
	}
	for i := range c.current {
		c.current[i] = -1
	}
	c.active = true
	return nil
}
func (c *TruthHierarchyCache) LevelSizes() []int {
	if c == nil || c.closed {
		return nil
	}
	sizes := make([]int, len(c.levels))
	for i, level := range c.levels {
		sizes[i] = level.memory.Len()
	}
	return sizes
}
func (c *TruthHierarchyCache) Close() {
	if c != nil {
		c.levels = nil
		c.current = nil
		c.active = false
		c.closed = true
	}
}
func (c *TruthHierarchyCache) Clone() *TruthHierarchyCache {
	if c == nil || c.closed {
		return nil
	}
	r := *c
	r.current = slices.Clone(c.current)
	r.levels = make([]truthHierarchyLevel, len(c.levels))
	for i, level := range c.levels {
		r.levels[i] = truthHierarchyLevel{level.memory.Clone(), slices.Clone(level.representatives)}
	}
	return &r
}

// RetrieveOrInsert accepts zero-based levels or negative indices counted from
// the end, matching C. Nil/closed caches and out-of-range levels report
// Unavailable. Valid visits require Begin and equal truth width. Intermediate
// misses leave a pending stage; hitting an already resolved stage or visiting
// the final level links only earlier stages visited by this traversal.
// Limits leave the failing insertion unchanged; Begin can abandon/retry a path.
func (c *TruthHierarchyCache) RetrieveOrInsert(level int, truth *TruthTable) (TruthHierarchyLookup, error) {
	result := TruthHierarchyLookup{Status: TruthHierarchyUnavailable, RepresentativeID: -1}
	if c == nil || c.closed {
		return result, nil
	}
	if level < 0 {
		level += len(c.levels)
	}
	if level < 0 || level >= len(c.levels) {
		return result, nil
	}
	if !c.active {
		return result, ErrInvariant
	}
	if e := truth.Check(); e != nil {
		return result, e
	}
	if truth.variables != c.variables {
		return result, ErrDimension
	}
	state := &c.levels[level]
	id, found, e := state.memory.Intern(truth.words)
	if e != nil {
		return result, e
	}
	if !found {
		state.representatives = append(state.representatives, -1)
	}
	c.current[level] = id
	last := len(c.levels) - 1
	representative := id
	if level < last {
		representative = state.representatives[id]
		if representative < 0 {
			return TruthHierarchyLookup{Status: TruthHierarchyInserted, RepresentativeID: -1}, nil
		}
	}
	words, e := c.levels[last].memory.Read(representative)
	if e != nil {
		return result, e
	}
	table, e := TruthTableFromWords(c.variables, words, truth.limit)
	if e != nil {
		return result, e
	}
	for i := 0; i < level; i++ {
		if c.current[i] >= 0 {
			c.levels[i].representatives[c.current[i]] = representative
		}
	}
	c.active = false
	status := TruthHierarchyInserted
	if level < last {
		status = TruthHierarchyHit
	}
	return TruthHierarchyLookup{status, table, representative}, nil
}
