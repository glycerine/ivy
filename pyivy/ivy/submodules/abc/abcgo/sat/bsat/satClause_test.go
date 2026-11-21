package bsat

import "testing"

func TestClauseIterationAccountsForPlusOneStorage(t *testing.T) {
	m := Sat_MemAlloc(9)
	first := Sat_MemAppend(m, []int32{0, 2, 4}, 3, 0, 1)
	second := Sat_MemAppend(m, []int32{1, 3, 5}, 3, 0, 1)
	got := m.handles(0)
	if len(got) != 2 || got[0] != first || got[1] != second {
		t.Fatalf("handles=%v, want [%d %d]", got, first, second)
	}
}

func TestClauseIterationMixedLayoutsAcrossPagesAndRollback(t *testing.T) {
	m := Sat_MemAlloc(9)
	var before []cla
	for i := 0; i < 100; i++ {
		plus := int32(i % 2)
		before = append(before, Sat_MemAppend(m, []int32{0, 2, 4}, 3, 0, plus))
	}
	Sat_MemBookMark(m)
	for i := 0; i < 60; i++ {
		Sat_MemAppend(m, []int32{1, 3, 5}, 3, 0, int32(i%2))
	}
	if got := m.handles(0); len(got) != 160 {
		t.Fatalf("before rollback: %d handles", len(got))
	}
	Sat_MemRollBack(m)
	got := m.handles(0)
	if len(got) != len(before) {
		t.Fatalf("after rollback: %d handles", len(got))
	}
	for i := range before {
		if got[i] != before[i] {
			t.Fatalf("handle %d=%d, want %d", i, got[i], before[i])
		}
	}
}

func TestClauseStorageHandles(t *testing.T) {
	p := Sat_MemAlloc(9)
	var hs []cla
	for i := 0; i < 300; i++ {
		l := []int32{int32(i * 2), int32(i*2 + 1)}
		h := Sat_MemAppend(p, l, 2, int32(i%2), 0)
		hs = append(hs, h)
		if Sat_MemClauseHand(p, h).lits[0] != l[0] || clause_learnt_h(p, h) != int32(i%2) {
			t.Fatal(i, h)
		}
	}
	if p.iPage[0] < 2 || p.iPage[1] < 3 {
		t.Fatal("did not cross pages")
	}
	if Sat_MemCountL(p) != 150 {
		t.Fatal("learned count")
	}
	for i, h := range hs {
		if Sat_MemClauseHand(p, h).lits[0] != int32(i*2) {
			t.Fatal("handle unstable")
		}
	}
}
func TestClauseBookmarkRollback(t *testing.T) {
	p := Sat_MemAlloc(9)
	a := Sat_MemAppend(p, []int32{2}, 1, 0, 1)
	Sat_MemBookMark(p)
	for i := 0; i < 300; i++ {
		Sat_MemAppend(p, []int32{4, 6}, 2, int32(i%2), 0)
	}
	Sat_MemRollBack(p)
	if p.nEntries != [2]int32{1, 0} || Sat_MemClauseHand(p, a).lits[0] != 2 || Sat_MemCountL(p) != 0 {
		t.Fatal(p.nEntries)
	}
	h := Sat_MemAppend(p, []int32{8}, 1, 1, 0)
	if clause_id(Sat_MemClauseHand(p, h)) != 0 {
		t.Fatal("ID after rollback")
	}
	Sat_MemRestart(p)
	if len(p.handles(0))+len(p.handles(1)) != 0 {
		t.Fatal("restart")
	}
}
func TestClauseCompactionRemap(t *testing.T) {
	p := Sat_MemAlloc(9)
	var keep []*clause
	for i := 0; i < 250; i++ {
		h := Sat_MemAppend(p, []int32{int32(i * 2), 2, 4}, 3, 1, 0)
		c := Sat_MemClauseHand(p, h)
		if i%3 == 0 {
			c.mark = 1
		} else {
			keep = append(keep, c)
		}
		if i == 100 {
			Sat_MemBookMark(p)
		}
	}
	if n := Sat_MemCompactLearned(p, 0); int(n) != len(keep) {
		t.Fatal(n)
	}
	newHandles := make([]cla, len(keep))
	for i, c := range keep {
		newHandles[i] = clause_id(c)
	}
	Sat_MemCompactLearned(p, 1)
	for i, h := range newHandles {
		c := Sat_MemClauseHand(p, h)
		if c != keep[i] || clause_id(c) != int32(i) {
			t.Fatal(i, h)
		}
	}
	Sat_MemRollBack(p)
	for _, h := range p.handles(1) {
		if h >= p.BookMarkH[1] {
			t.Fatal("rollback retained entry")
		}
	}
}
func TestLiteralClauseHandles(t *testing.T) {
	for l := lit(0); l < 100; l++ {
		h := clause_from_lit(l)
		if clause_is_lit(h) != 1 || clause_read_lit(h) != l {
			t.Fatal(l, h)
		}
	}
}

func TestClauseCompactionAllDeleted(t *testing.T) {
	p := Sat_MemAlloc(9)
	for i := 0; i < 100; i++ {
		h := Sat_MemAppend(p, []int32{2, 4}, 2, 1, 0)
		Sat_MemClauseHand(p, h).mark = 1
	}
	Sat_MemBookMark(p)
	if Sat_MemCompactLearned(p, 0) != 0 || Sat_MemCompactLearned(p, 1) != 0 || p.BookMarkE[1] != 0 || p.BookMarkH[1] != Sat_MemHand(p, 1, 2) {
		t.Fatal("all-deleted bookmark")
	}
	Sat_MemRollBack(p)
	h := Sat_MemAppend(p, []int32{8}, 1, 1, 0)
	if clause_id(Sat_MemClauseHand(p, h)) != 0 {
		t.Fatal("ID after empty compaction")
	}
}

func TestClausePageExtractionAtHighestSize(t *testing.T) {
	p := Sat_MemAlloc(31)
	problem := Sat_MemAppend(p, []int32{2, 4}, 2, 0, 0)
	learned := Sat_MemAppend(p, []int32{3, 5}, 2, 1, 0)
	if page := Sat_MemHandPage(p, learned); page != 1 {
		t.Fatalf("learned handle %#x decodes to page %d, want 1", uint32(learned), page)
	}
	Sat_MemBookMark(p)
	Sat_MemAppend(p, []int32{0, 2, 4}, 3, 1, 0)
	Sat_MemRollBack(p)
	if clause_size(Sat_MemClauseHand(p, problem)) != 2 || clause_size(Sat_MemClauseHand(p, learned)) != 2 || Sat_MemCountL(p) != 1 {
		t.Fatal("high-bit learned handle did not survive rollback")
	}
}

func TestReclaimedClauseHandlesRejectReads(t *testing.T) {
	p := Sat_MemAlloc(9)
	first := Sat_MemAppend(p, []int32{2, 4, 6}, 3, 1, 0)
	other := Sat_MemAppend(p, []int32{1, 3, 5}, 3, 0, 0)
	Sat_MemBookMark(p)
	last := first
	for range 120 {
		last = Sat_MemAppend(p, []int32{0, 2, 4}, 3, 1, 0)
	}
	reject := func(h int32) {
		t.Helper()
		defer func() {
			if recover() == nil {
				t.Errorf("accepted invalid clause handle %d", h)
			}
		}()
		Sat_MemClauseHand(p, h)
	}
	Sat_MemRollBack(p)
	reject(last)
	reject(first + 1)
	reject(first + 2)
	if Sat_MemClauseHand(p, other).lits[2] != 5 || Sat_MemClauseHand(p, first).lits[2] != 6 {
		t.Fatal("rollback removed a live clause")
	}
	Sat_MemRestart(p)
	reject(first)
	reject(other)
}

func TestRejectedCompactionPreservesLearnedClauses(t *testing.T) {
	p := Sat_MemAlloc(9)
	first := Sat_MemAppend(p, []int32{0, 2, 4}, 3, 1, 0)
	second := Sat_MemAppend(p, []int32{1, 3, 5}, 3, 1, 0)
	one, two := Sat_MemClauseHand(p, first), Sat_MemClauseHand(p, second)
	Sat_MemCompactLearned(p, 0)
	clause_set_id(two, -1) // A stale second relocation must not commit the first.
	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("invalid relocation accepted")
			}
		}()
		Sat_MemCompactLearned(p, 1)
	}()
	if clause_id(one) != first || Sat_MemClauseHand(p, first) != one || Sat_MemClauseHand(p, second) != two {
		t.Fatal("failed compaction changed a live clause or its relocation ID")
	}
	Sat_MemCompactLearned(p, 0)
	if Sat_MemCompactLearned(p, 1) != 2 || clause_id(one) != 0 || clause_id(two) != 1 {
		t.Fatal("retry failed to restore packed clause IDs")
	}
}

func TestBulkShortClauseStorage(t *testing.T) {
	const count = 512
	data := []lit{0, 2, 4}
	allocs := testing.AllocsPerRun(10, func() {
		m := Sat_MemAlloc(12)
		for i := 0; i < count; i++ {
			h := Sat_MemAppend(m, data, 3, int32(i&1), 1)
			if clause_id(Sat_MemClauseHand(m, h)) != int32(i/2) {
				t.Fatal("short clause lost its stream ID")
			}
		}
	})
	if allocs > 100 {
		t.Errorf("storing %d short clauses allocated %.0f objects", count, allocs)
	}

	m := Sat_MemAlloc(9)
	first := Sat_MemClauseHand(m, Sat_MemAppend(m, data, 3, 0, 1))
	second := Sat_MemClauseHand(m, Sat_MemAppend(m, []lit{1, 3, 5}, 3, 0, 1))
	changed := append(first.lits, 123)
	changed[0] = 101
	if second.lits[0] != 1 || first.lits[0] != 0 || data[0] != 0 {
		t.Fatal("clause ID or literal storage aliases its neighbor or input")
	}
	for i := 0; i < 400; i++ {
		Sat_MemAppend(m, data, 3, 1, 0)
	}
	Sat_MemRestart(m)
	Sat_MemAppend(m, nil, 3, 0, 0)
	if first.lits[2] != 4 || second.lits[2] != 5 {
		t.Fatal("reusing a page overwrote a retained clause")
	}
}

func TestWatchTagPreservesHighestEncodableVariables(t *testing.T) {
	for _, v := range []int32{1 << 28, 1<<29 - 1, 1 << 29, 1<<30 - 1} {
		for _, l := range []Literal{Positive(v), Negative(v)} {
			h := clause_from_lit(l)
			if got := clause_read_lit(h); clause_is_lit(h) == 0 || got != l {
				t.Errorf("literal %d encoded as %#x and decoded as %d", l, uint32(h), got)
			}
		}
	}
}

func TestLiteralSliceAppendPreservesClauseID(t *testing.T) {
	m := Sat_MemAlloc(9)
	for _, size := range []int32{0, 1, 2, 3, 4, 17} {
		c := Sat_MemClauseHand(m, Sat_MemAppend(m, nil, size, 1, 0))
		want := clause_id(c)
		for _, view := range [][]lit{clause_begin(c), clause_end(c)} {
			grown := append(view, 99)
			if grown[len(grown)-1] != 99 || clause_id(c) != want {
				t.Errorf("appending to %d-literal clause view overwrote ID %d with %d", size, want, clause_id(c))
				clause_set_id(c, want)
			}
		}
	}
}

func TestProofPartitionSettersKeepOneBit(t *testing.T) {
	s := sat_solver2_new()
	defer sat_solver2_delete(s)
	sat_solver2_setnvars(s, 2)
	h := Sat_MemAppend(&s.Mem, []lit{0, 2}, 2, 0, 1)
	for _, value := range []int32{1, 2, 3, 256, -1, -2, 0} {
		clause2_set_partA(s, h, value)
		var_set_partA(s, 1, value)
		if c, v := clause2_is_partA(s, h), var_is_partA(s, 1); c != value&1 || v != value&1 {
			t.Errorf("partition %d: clause %d and variable %d, want %d", value, c, v, value&1)
		}
	}
}

func BenchmarkStoreMixedSATClauses(b *testing.B) {
	data := []lit{0, 2, 4, 6, 8, 10, 12, 14}
	b.ReportAllocs()
	for b.Loop() {
		m := Sat_MemAlloc(12)
		var last cla
		for i := range 1024 {
			size := int32(2 + i%7)
			last = Sat_MemAppend(m, data, size, int32(i%2), 1)
		}
		if clause_id(Sat_MemClauseHand(m, last)) != 511 {
			b.Fatal("wrong clause ID")
		}
		Sat_MemFree(m)
	}
}

func TestClauseWordCountRejectsOverflow(t *testing.T) {
	if got := Sat_MemIntSize(1<<31-5, 1); got != 1<<31-2 {
		t.Fatalf("largest even word count %d", got)
	}
	for _, learned := range []int32{0, 1} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("word count wrapped for maximum clause size, learned=%d", learned)
				}
			}()
			Sat_MemIntSize(1<<31-1, learned)
		}()
	}
}

func TestLastClausePageCannotWrapOntoLiveHandle(t *testing.T) {
	for stream := range int32(2) {
		m := Sat_MemAlloc(31)
		h := Sat_MemAppend(m, []lit{0, 2}, 2, stream, 1)
		original := Sat_MemClauseHand(m, h)
		// Exercise the page transition without allocating a two-billion-word page.
		m.pPages[stream].limit = 1<<31 - 4
		before := m.nEntries
		func() {
			defer func() {
				if recover() == nil {
					t.Error("new page reused a live clause handle")
				}
			}()
			Sat_MemAppend(m, []lit{4, 6}, 2, stream, 1)
		}()
		if m.iPage[stream] != stream || m.nEntries != before || len(m.pPages) != 2 || Sat_MemClauseHand(m, h) != original {
			t.Fatal("page exhaustion changed the live clause store")
		}
	}
}

func TestSparseClauseOffsetsOnlyIndexLiveHeaders(t *testing.T) {
	s := newSatSolver(19, 91648253)
	defer sat_solver_delete(s)
	m := &s.Mem
	huge := Sat_MemAppend(m, nil, 1<<16, 0, 0)
	large := Sat_MemClauseHand(m, huge)
	before := sat_solver_memory(s)
	tail := Sat_MemAppend(m, []lit{2, 4, 6}, 3, 0, 1)
	if extra := sat_solver_memory(s) - before; extra > 8192 {
		t.Fatalf("indexing one clause beyond 65536 literals uses %.0f more bytes", extra)
	}
	if c := Sat_MemClauseHand(m, tail); clause_size(c) != 3 || c.lits[2] != 6 || Sat_MemClauseHand(m, huge) != large {
		t.Fatal("sparse header lookup lost a live clause")
	}
}

func TestMixedClauseBatchUsesBlocks(t *testing.T) {
	allocs := testing.AllocsPerRun(10, func() {
		m := Sat_MemAlloc(12)
		for i := range 1024 {
			Sat_MemAppend(m, nil, int32(2+i%7), int32(i%2), 1)
		}
		Sat_MemFree(m)
	})
	if allocs > 200 {
		t.Fatalf("1024 mixed clauses used %.0f allocations", allocs)
	}
}

func TestRetainedMediumClausesSurviveRollbackAndRestart(t *testing.T) {
	m := Sat_MemAlloc(12)
	Sat_MemBookMark(m)
	var saved []*clause
	for i := range 300 {
		size := int32(5 + i%40)
		c := Sat_MemClauseHand(m, Sat_MemAppend(m, nil, size, int32(i%2), 1))
		c.lits[len(c.lits)-1] = int32(2 * i)
		saved = append(saved, c)
	}
	Sat_MemRollBack(m)
	for i := range 600 {
		Sat_MemAppend(m, nil, int32(5+i%40), int32(i%2), 1)
	}
	Sat_MemRestart(m)
	for i, c := range saved {
		grew := append(c.lits, 123)
		grew[0] = 456
		if c.lits[0] != 0 || c.lits[len(c.lits)-1] != int32(2*i) || clause_id(c) != int32(i/2) {
			t.Fatalf("retained clause %d changed", i)
		}
	}
}

func TestClauseShrinkRejectsOffsetsInsideALiveClause(t *testing.T) {
	for _, offset := range []int32{1, 2, 256, 512, 768} {
		m := Sat_MemAlloc(12)
		h := Sat_MemAppend(m, nil, 1024, 0, 0)
		end := Sat_MemHandCurrent(m, 0)
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("accepted interior offset %d", offset)
				}
			}()
			Sat_MemShrink(m, h+offset, 0)
		}()
		if Sat_MemHandCurrent(m, 0) != end {
			t.Errorf("bad offset %d moved the allocation cursor", offset)
		}
	}
}

func TestStaleLearnedBookmarkPreservesProblemStream(t *testing.T) {
	m := Sat_MemAlloc(12)
	Sat_MemAppend(m, nil, 3, 0, 0)
	learned := Sat_MemAppend(m, nil, 3, 1, 0)
	Sat_MemBookMark(m)
	later := Sat_MemAppend(m, nil, 3, 0, 0)
	end := Sat_MemHandCurrent(m, 0)
	Sat_MemShrink(m, learned, 1)
	func() {
		defer func() {
			if recover() == nil {
				t.Error("stale learned bookmark was accepted")
			}
		}()
		Sat_MemRollBack(m)
	}()
	if Sat_MemHandCurrent(m, 0) != end || Sat_MemClauseHand(m, later) == nil {
		t.Fatal("failed rollback removed a problem clause")
	}
}
