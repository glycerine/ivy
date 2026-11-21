package core

import "time"

// Cudd_GarbageCollect performs mark-and-sweep over the unique tables. Nodes
// with references and manager-owned variable/terminal roots are retained,
// along with every node reachable through their children. Pass clearCache=false
// only when the computed cache has already been flushed.
func Cudd_GarbageCollect(dd *DdManager, clearCache bool) int32 {
	if dd == nil || dd.closed {
		return 0
	}
	started := time.Now()
	if !dd.runHooks(CUDD_PRE_GC_HOOK, "DD", nil) {
		return 0
	}
	dd.peakLive = max(dd.peakLive, dd.liveBDDKeys())
	marked := make(map[*DdNode]bool)
	incoming := make(map[*DdNode]uint64)
	dd.countIncomingReferences(incoming)
	stack := make([]*DdNode, 0, dd.unique.count+dd.addUnique.count+dd.zddUnique.count)
	tables := []*uniqueTable{&dd.unique, &dd.addUnique, &dd.zddUnique}
	mark := func(n *DdNode) {
		if n != nil {
			n = Cudd_Regular(n)
			if !marked[n] {
				marked[n] = true
				stack = append(stack, n)
			}
		}
	}
	mark(dd.One)
	mark(dd.LogicZero)
	mark(dd.Zero)
	mark(dd.Background)
	mark(dd.PlusInfinity)
	mark(dd.MinusInfinity)
	for _, n := range dd.Vars {
		mark(n)
	}
	for _, n := range dd.Univ {
		mark(n)
	}
	for _, table := range tables {
		for _, n := range table.entries {
			if n.Ref == ^uint32(0) || uint64(n.Ref) > incoming[n] {
				mark(n)
			}
		}
	}
	for _, n := range dd.constants {
		if n.Ref == ^uint32(0) || uint64(n.Ref) > incoming[n] {
			mark(n)
		}
	}
	for len(stack) != 0 {
		n := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		mark(n.T)
		mark(n.E)
	}
	deleted := 0
	for _, table := range tables {
		deleted += table.sweep(func(n *DdNode) bool {
			if marked[n] {
				return false
			}
			dd.retireNode(n)
			return true
		})
	}
	for key, n := range dd.constants {
		if !marked[n] {
			dd.retireNode(n)
			delete(dd.constants, key)
			deleted++
		}
	}
	if deleted == 0 {
		if !dd.runHooks(CUDD_POST_GC_HOOK, "DD", nil) {
			return 0
		}
		return 0
	}
	dd.garbageCollections++
	if clearCache {
		live := func(n *DdNode) bool { return n == nil || marked[Cudd_Regular(n)] }
		dd.computed.valid -= dd.computed.invalidate(func(entry computedEntry) bool {
			key := entry.key
			return !live(key.f) || !live(key.g) || !live(key.h) || !live(entry.result)
		})
	}
	dd.gcTime += uint64(time.Since(started).Nanoseconds())
	if !dd.runHooks(CUDD_POST_GC_HOOK, "DD", nil) {
		return 0
	}
	return int32(deleted)
}

func (dd *DdManager) retireNode(n *DdNode) {
	if n.inBDDTable && n.childrenReleased {
		dd.dead--
	}
	n.inBDDTable = false
	if !n.childrenReleased {
		n.childrenReleased = true
		dd.recursiveDeref(n.T)
		dd.recursiveDeref(n.E)
	}
	n.detachHandle()
}

// purgeReleasedKeys collects only nodes whose owned child references were
// explicitly released. An unreferenced result still in use by a Go caller
// retains its children and remains canonical.
func (dd *DdManager) purgeReleasedKeys() int {
	reclaimNode(Cudd_Regular(dd.Background))
	deleted := 0
	for _, table := range []*uniqueTable{&dd.unique, &dd.addUnique, &dd.zddUnique} {
		if dd.dead == 0 && table != &dd.zddUnique {
			continue
		}
		deleted += table.sweep(func(n *DdNode) bool {
			if !n.childrenReleased {
				return false
			}
			dd.retireNode(n)
			return true
		})
	}
	if dd.dead != 0 {
		for key, n := range dd.constants {
			if n.childrenReleased {
				delete(dd.constants, key)
				dd.retireNode(n)
				deleted++
			}
		}
	}
	return deleted
}

// Detach a retired handle after accounting for its former child references.
// Keep its ID and complementary handle available for identity comparisons.
func (n *DdNode) detachHandle() {
	n.manager, n.T, n.E = nil, nil, nil
	n.inBDDTable = false
	n.childrenReleased = true
	if n.inverse != nil {
		n.inverse.manager = nil
	}
}

func (dd *DdManager) collectUnderPressure(t, e *DdNode) {
	dd.pressureGC = true
	defer func() { dd.pressureGC = false }()
	started := time.Now()
	if !dd.runHooks(CUDD_PRE_GC_HOOK, "DD", nil) {
		return
	}
	reclaimNode(Cudd_Regular(t))
	reclaimNode(Cudd_Regular(e))
	dd.peakLive = max(dd.peakLive, dd.liveBDDKeys())
	if dd.purgeReleasedKeys() != 0 {
		retired := func(n *DdNode) bool {
			return n != nil && n != DD_NON_CONSTANT && n.manager == nil
		}
		dd.computed.valid -= dd.computed.invalidate(func(e computedEntry) bool {
			return retired(e.key.f) || retired(e.key.g) || retired(e.key.h) || retired(e.result)
		})
		dd.garbageCollections++
		dd.gcTime += uint64(time.Since(started).Nanoseconds())
	}
	dd.runHooks(CUDD_POST_GC_HOOK, "DD", nil)
}
