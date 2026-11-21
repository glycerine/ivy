// Copyright Satoko contributors. Distributed under the BSD 2-Clause License.
package satoko

type shortOriginal struct {
	clause
	data [5]int
}

// These buffers belong to a single initial import. Independent clause and
// watch-list slices keep the occupied chunks alive after import ends.
type clauseImport struct {
	remaining     int
	watchCapacity int
	records       []shortOriginal
	watches       []watcher
}

func (b *clauseImport) nextRecord() *shortOriginal {
	if b == nil {
		return new(shortOriginal)
	}
	if len(b.records) == 0 {
		b.records = make([]shortOriginal, min(b.remaining, 128))
	}
	record := &b.records[0]
	b.records = b.records[1:]
	return record
}

func (b *clauseImport) startWatchList() []watcher {
	capacity := b.watchCapacity
	if len(b.watches) == 0 {
		b.watches = make([]watcher, 2*min(b.remaining, 64)*capacity)
	}
	entries := b.watches[:0:capacity]
	b.watches = b.watches[capacity:]
	return entries
}

// Clause IDs remain stable while Go's garbage collector owns the storage.
func (s *Solver) addStored(lits []int, learnt bool, lbd uint32) int {
	id := len(s.clauses)
	var c *clause
	if !learnt && len(lits) > 0 && len(lits) <= 5 {
		// The common short original clauses own their literals in the same allocation.
		record := s.importing.nextRecord()
		c = &record.clause
		c.lbd = lbd
		c.lits = record.data[:len(lits):len(lits)]
		copy(c.lits, lits)
	} else {
		c = &clause{lits: append([]int(nil), lits...), learnt: learnt, lbd: lbd}
	}
	s.clauses = append(s.clauses, c)
	if len(lits) > 1 {
		s.addWatch(lits[0]^1, id)
		s.addWatch(lits[1]^1, id)
	}
	if learnt && len(lits) > 1 {
		s.stats.LearntLiterals += int64(len(lits))
	} else if !learnt && len(lits) > 1 {
		s.stats.OriginalLiterals += int64(len(lits))
	}
	return id
}
func (s *Solver) rebuildWatches() {
	s.watches = make([]watchList, 2*len(s.values))
	for id, c := range s.clauses {
		if !c.deleted && len(c.lits) > 1 {
			s.addWatch(c.lits[0]^1, id)
			s.addWatch(c.lits[1]^1, id)
		}
	}
}

type watcher struct {
	clause, blocker int
}

type watchList struct {
	entries  []watcher
	binaries int
}

// Binary watches precede long watches, matching Satoko's propagation priority.
func (s *Solver) addWatch(lit, id int) {
	c := s.clauses[id]
	blocker := c.lits[0]
	if blocker == lit^1 {
		blocker = c.lits[1]
	}
	list := &s.watches[lit]
	if s.importing != nil && list.entries == nil {
		list.entries = s.importing.startWatchList()
	}
	last := len(list.entries)
	list.entries = append(list.entries, watcher{clause: id, blocker: blocker})
	if len(c.lits) == 2 {
		i := list.binaries
		list.entries[i], list.entries[last] = list.entries[last], list.entries[i]
		list.binaries++
	}
}
