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
	// All parents precede their children in the current variable order. Once
	// unowned parents have released their edges, a remaining reference comes
	// from a survivor or an external owner. Explicitly mark survivor children
	// as well, including children of roots whose own reference count is zero.
	mark := func(n *DdNode) {
		if n != nil {
			Cudd_Regular(n).gcMarked = true
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
	drop := func(n *DdNode) bool {
		if n.Ref != 0 || n.gcMarked {
			n.gcMarked = false
			mark(n.T)
			mark(n.E)
			return false
		}
		dd.retireNode(n)
		return true
	}
	deleted := dd.unique.sweep(dd.InvPerm, drop)
	deleted += dd.addUnique.sweep(dd.InvPerm, drop)
	deleted += dd.zddUnique.sweep(dd.InvPermZ, drop)
	for key, n := range dd.constants {
		if drop(n) {
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
		live := func(n *DdNode) bool {
			if n == nil {
				return true
			}
			r := Cudd_Regular(n)
			return r.manager == dd && (r.inBDDTable || r.family == 1)
		}
		dd.computed.valid -= dd.computed.invalidate(func(entry computedEntry) bool {
			key := entry.key
			return !live(key.f) || !live(key.g) || !live(key.h) || (entry.result != DD_NON_CONSTANT && !live(entry.result))
		})
	}
	dd.gcTime += uint64(time.Since(started).Nanoseconds())
	if !dd.runHooks(CUDD_POST_GC_HOOK, "DD", nil) {
		return 0
	}
	return int32(deleted)
}

func (dd *DdManager) retireNode(n *DdNode) {
	if n.childrenReleased {
		if n.inBDDTable {
			dd.dead--
		} else if n.family == 1 {
			dd.deadZ--
		}
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
		if table == &dd.zddUnique && dd.deadZ == 0 || table != &dd.zddUnique && dd.dead == 0 {
			continue
		}
		deleted += table.sweep(nil, func(n *DdNode) bool {
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

// collectReleasedKeys preserves active Go temporaries. During a table insertion
// t and e also protect the children of the key being constructed.
func (dd *DdManager) collectReleasedKeys(t, e *DdNode) {
	previous := dd.pressureGC
	dd.pressureGC = true
	defer func() { dd.pressureGC = previous }()
	started := time.Now()
	if !dd.runHooks(CUDD_PRE_GC_HOOK, "DD", nil) {
		return
	}
	reclaimNode(Cudd_Regular(t))
	reclaimNode(Cudd_Regular(e))
	dd.peakLive = max(dd.peakLive, dd.liveBDDKeys())
	if dd.purgeReleasedKeys() != 0 {
		if dd.computed.valid != 0 {
			retired := func(n *DdNode) bool {
				return n != nil && n != DD_NON_CONSTANT && n.manager == nil
			}
			dd.computed.valid -= dd.computed.invalidate(func(e computedEntry) bool {
				return retired(e.key.f) || retired(e.key.g) || retired(e.key.h) || retired(e.result)
			})
		}
		dd.garbageCollections++
		dd.gcTime += uint64(time.Since(started).Nanoseconds())
	}
	dd.runHooks(CUDD_POST_GC_HOOK, "DD", nil)
}
