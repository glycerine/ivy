package core

import "slices"

// Partition by variable index so a permutation can change without invalidating
// other levels. Keep keys separate from the nodes that reordering rewires.
type uniqueTable struct {
	levels []uniqueLevel
	count  int
}

type uniqueEntry struct {
	t, e, node *DdNode
}

type uniqueLevel struct {
	// A level with only its projection needs no separate slot array.
	first       uniqueEntry
	slots       []uniqueEntry
	used, count int // used includes tombstones, count includes only live slots
}

// An erased slot keeps its probe chain intact and holds no diagram pointers.
var removedUniqueEntry = uniqueEntry{e: &DdNode{}}

func uniqueHash(t, e *DdNode) int {
	h := t.ID*0x9e3779b97f4a7c15 ^ e.ID*0xc2b2ae3d27d4eb4f
	return int(h ^ h>>32)
}

func (l *uniqueLevel) slot(t, e *DdNode) int {
	mask := len(l.slots) - 1
	firstRemoved := -1
	for pos := uniqueHash(t, e) & mask; ; pos = (pos + 1) & mask {
		record := &l.slots[pos]
		if record.t == t && record.e == e {
			return pos
		}
		if record.t == nil {
			if record.e == nil {
				if firstRemoved >= 0 {
					return firstRemoved
				}
				return pos
			}
			if firstRemoved < 0 {
				firstRemoved = pos
			}
		}
	}
}

func (table *uniqueTable) find(k bddUniqueKey) *uniqueEntry {
	if k.index >= uint32(len(table.levels)) {
		return nil
	}
	l := &table.levels[k.index]
	if len(l.slots) == 0 {
		if l.first.node == nil || l.first.t == k.t && l.first.e == k.e {
			return &l.first
		}
		return nil
	}
	return &l.slots[l.slot(k.t, k.e)]
}

func (l *uniqueLevel) resize(size int) {
	old := l.slots
	l.slots = make([]uniqueEntry, size)
	if l.first.node != nil {
		l.slots[l.slot(l.first.t, l.first.e)] = l.first
		l.first = uniqueEntry{}
	}
	for _, r := range old {
		if r.node != nil {
			l.slots[l.slot(r.t, r.e)] = r
		}
	}
	l.used = l.count
}

func (table *uniqueTable) get(k bddUniqueKey) *DdNode {
	if r := table.find(k); r != nil {
		return r.node
	}
	return nil
}

func (table *uniqueTable) set(k bddUniqueKey, node *DdNode) {
	record := table.find(k)
	if record != nil && record.node != nil {
		record.node = node
		return
	}
	table.insert(k, node, record)
}

// insert uses the probe performed by find. A resize invalidates its position.
func (table *uniqueTable) insert(k bddUniqueKey, node *DdNode, record *uniqueEntry) {
	if n := int(k.index) + 1; n > len(table.levels) {
		table.levels = slices.Grow(table.levels, n-len(table.levels))[:n]
	}
	l := &table.levels[k.index]
	r := uniqueEntry{k.t, k.e, node}
	if len(l.slots) == 0 {
		if l.count == 0 {
			l.first, l.count = r, 1
			table.count++
			return
		}
		l.resize(4)
		record = nil
	} else if 4*(l.used+1) > 3*len(l.slots) {
		next := len(l.slots)
		if 2*(l.count+1) > next {
			next *= 2
		}
		l.resize(next)
		record = nil
	}
	if record == nil {
		record = &l.slots[l.slot(k.t, k.e)]
	}
	table.count++
	l.count++
	if record.e == nil {
		l.used++
	}
	*record = r
}

func (table *uniqueTable) remove(k bddUniqueKey) {
	if k.index >= uint32(len(table.levels)) {
		return
	}
	l := &table.levels[k.index]
	if len(l.slots) == 0 {
		if l.first.node == nil || l.first.t != k.t || l.first.e != k.e {
			return
		}
		l.first = uniqueEntry{}
	} else {
		p := l.slot(k.t, k.e)
		if l.slots[p].node == nil {
			return
		}
		l.slots[p] = removedUniqueEntry
	}
	l.count--
	table.count--
	if l.count == 0 {
		*l = uniqueLevel{}
	}
}

func (table *uniqueTable) entries(yield func(bddUniqueKey, *DdNode) bool) {
	for i := range table.levels {
		l := &table.levels[i]
		if f := l.first; f.node != nil && !yield(bddUniqueKey{uint32(i), f.t, f.e}, f.node) {
			return
		}
		for _, r := range l.slots {
			if r.node != nil && !yield(bddUniqueKey{uint32(i), r.t, r.e}, r.node) {
				return
			}
		}
	}
}
