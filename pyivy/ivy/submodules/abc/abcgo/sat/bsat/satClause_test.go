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
	if len(p.clauses) != 0 {
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
