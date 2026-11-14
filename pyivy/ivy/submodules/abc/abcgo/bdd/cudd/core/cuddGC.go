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
	if !dd.runHooks(CUDD_PRE_GC_HOOK, "DD", nil) {
		return 0
	}
	marked := make(map[*DdNode]bool)
	incoming := make(map[*DdNode]uint64)
	stack := make([]*DdNode, 0, len(dd.unique)+len(dd.addUnique)+len(dd.zddUnique))
	tables := []map[bddUniqueKey]*DdNode{dd.unique, dd.addUnique, dd.zddUnique}
	for _, table := range tables {
		for _, n := range table {
			if !n.childrenReleased {
				if t := Cudd_Regular(n.T); t != nil {
					incoming[t]++
				}
				if e := Cudd_Regular(n.E); e != nil {
					incoming[e]++
				}
			}
		}
	}
	for _, n := range dd.constants {
		if !n.childrenReleased {
			if t := Cudd_Regular(n.T); t != nil {
				incoming[t]++
			}
			if e := Cudd_Regular(n.E); e != nil {
				incoming[e]++
			}
		}
	}
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
		for _, n := range table {
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
		for key, n := range table {
			if !marked[n] {
				if !n.childrenReleased {
					n.childrenReleased = true
					Cudd_RecursiveDeref(dd, n.T)
					Cudd_RecursiveDeref(dd, n.E)
				}
				delete(table, key)
				deleted++
			}
		}
	}
	for key, n := range dd.constants {
		if !marked[n] {
			if !n.childrenReleased {
				n.childrenReleased = true
				Cudd_RecursiveDeref(dd, n.T)
				Cudd_RecursiveDeref(dd, n.E)
			}
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
		for key, value := range dd.cache {
			if !live(key.f) || !live(key.g) || !live(key.h) || !live(value) {
				delete(dd.cache, key)
			}
		}
		for key, value := range dd.addCache {
			if !live(key.f) || !live(key.g) || !live(key.h) || !live(value) {
				delete(dd.addCache, key)
			}
		}
		for key, value := range dd.computed {
			if !live(key.f) || !live(key.g) || !live(key.h) || !live(value) {
				delete(dd.computed, key)
			}
		}
	}
	dd.gcTime += uint64(time.Since(started).Nanoseconds())
	if !dd.runHooks(CUDD_POST_GC_HOOK, "DD", nil) {
		return 0
	}
	return int32(deleted)
}
