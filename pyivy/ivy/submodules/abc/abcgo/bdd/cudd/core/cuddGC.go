package core

import "time"

// Cudd_GarbageCollect performs mark-and-sweep over the Go unique maps. Nodes
// with references and manager-owned variable/terminal roots are retained,
// along with every node reachable through their children.
func Cudd_GarbageCollect(dd *DdManager, clearCache bool) int32 {
	if dd == nil || dd.closed {
		return 0
	}
	started := time.Now()
	defer func() {
		dd.garbageCollections++
		dd.gcTime += uint64(time.Since(started).Nanoseconds())
	}()
	if !dd.runHooks(CUDD_PRE_GC_HOOK, "DD", nil) {
		return 0
	}
	marked := make(map[*DdNode]bool)
	stack := make([]*DdNode, 0, len(dd.unique)+len(dd.addUnique)+len(dd.zddUnique))
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
	mark(dd.Zero)
	mark(dd.Background)
	for _, n := range dd.Vars {
		mark(n)
	}
	for _, table := range []map[bddUniqueKey]*DdNode{dd.unique, dd.addUnique, dd.zddUnique} {
		for _, n := range table {
			if n.Ref != 0 {
				mark(n)
			}
		}
	}
	for _, n := range dd.constants {
		if n.Ref != 0 {
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
	for _, table := range []map[bddUniqueKey]*DdNode{dd.unique, dd.addUnique, dd.zddUnique} {
		for key, n := range table {
			if !marked[n] {
				delete(table, key)
				deleted++
			}
		}
	}
	for key, n := range dd.constants {
		if !marked[n] {
			delete(dd.constants, key)
			deleted++
		}
	}
	if clearCache {
		clear(dd.cache)
		clear(dd.computed)
		clear(dd.addCache)
	}
	dd.runHooks(CUDD_POST_GC_HOOK, "DD", nil)
	return int32(deleted)
}
