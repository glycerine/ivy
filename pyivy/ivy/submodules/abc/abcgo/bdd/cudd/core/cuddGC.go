package core

import "time"

// Cudd_GarbageCollect performs mark-and-sweep over the unique tables. Nodes
// with references and manager-owned variable/terminal roots are retained,
// along with every node reachable through their children.
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
		for key, n := range table.entries {
			if !marked[n] {
				dd.sweepNode(n)
				table.remove(key)
				deleted++
			}
		}
	}
	for key, n := range dd.constants {
		if !marked[n] {
			dd.sweepNode(n)
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

func (dd *DdManager) sweepNode(n *DdNode) {
	if n.inBDDTable && n.childrenReleased {
		dd.dead--
	}
	n.inBDDTable = false
	if !n.childrenReleased {
		n.childrenReleased = true
		dd.recursiveDeref(n.T)
		dd.recursiveDeref(n.E)
	}
}
