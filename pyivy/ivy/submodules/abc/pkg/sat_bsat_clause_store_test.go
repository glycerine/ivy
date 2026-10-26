package abc

// Stand-alone tests for the owned port of src/sat/bsat/satClause.h.

import (
	"errors"
	"math"
	"reflect"
	"testing"
)

func satClauseTestWords(size int, learned, plusOne bool) int {
	extra := 0
	if learned || plusOne {
		extra = 1
	}
	return 2 * ((size + 2 + extra) / 2)
}

func satClauseTestHandles(pageSize int, clauses []SatStoredClause) []SatClauseHandle {
	pageWords, page, offset := 1<<uint(pageSize), 1, 2
	result := make([]SatClauseHandle, 0, len(clauses))
	for _, clause := range clauses {
		if clause.Mark {
			continue
		}
		if offset+clause.StorageWords >= pageWords {
			page, offset = page+2, 2
		}
		result = append(result, SatClauseHandle(uint32(page)<<uint(pageSize)|uint32(offset)))
		offset += clause.StorageWords
	}
	return result
}

func TestSatClauseTaggedLiteralRoundTrip(t *testing.T) {
	for _, literal := range []SatLit{0, 1, 2, 3, SatLit(math.MaxInt32)} {
		handle, err := SatClauseHandleFromLiteral(literal)
		if err != nil || !handle.IsLiteral() {
			t.Fatal(literal, handle, err)
		}
		got, err := handle.Literal()
		if err != nil || got != literal {
			t.Fatal(literal, got, err)
		}
	}
	if _, err := SatClauseHandle(2).Literal(); !errors.Is(err, ErrInvalidLiteral) {
		t.Fatal(err)
	}
	if _, err := SatClauseHandleFromLiteral(SatLit(math.MaxUint32)); !errors.Is(err, ErrInvalidLiteral) {
		t.Fatal(err)
	}
}

func TestSatClauseStorePackingSnapshotsAndIteration(t *testing.T) {
	s, err := NewSatClauseStore(9, SatClauseStoreLimits{})
	if err != nil {
		t.Fatal(err)
	}
	scratch, err := s.AppendBlank(2, false, false)
	if err != nil || scratch != 2 {
		t.Fatal(scratch, err)
	}
	inputs := [][]SatLit{{0}, {2, 5}, {1, 3, 4}, make([]SatLit, 490), {6, 8, 10, 12}}
	var handles []SatClauseHandle
	offset, page := 2+satClauseTestWords(2, false, false), 0
	for i, literals := range inputs {
		plusOne := i&1 != 0
		words := satClauseTestWords(len(literals), false, plusOne)
		if offset+words+2 >= 512 {
			page, offset = page+2, 2
		}
		handle, appendErr := s.Append(literals, false, plusOne)
		want := SatClauseHandle(uint32(page)<<9 | uint32(offset))
		if appendErr != nil || handle != want {
			t.Fatal(i, handle, want, appendErr)
		}
		handles = append(handles, handle)
		offset += words
	}
	if got := s.ProblemClauses(); len(got) != len(inputs) {
		t.Fatal(len(got))
	} else {
		for i := range got {
			if got[i].Handle != handles[i] || !reflect.DeepEqual(got[i].Literals, inputs[i]) || got[i].StorageWords != satClauseTestWords(len(inputs[i]), false, i&1 != 0) || got[i].HasID != (i&1 != 0) {
				t.Fatal(i, got[i])
			}
		}
		got[0].Literals[0] = 99
		clause, _ := s.Clause(handles[0])
		if clause.Literals[0] != inputs[0][0] {
			t.Fatal("snapshot aliases store")
		}
	}
	if err = s.SetMetadata(handles[2], true, true, 255); err != nil {
		t.Fatal(err)
	}
	clause, err := s.Clause(handles[2])
	if err != nil || !clause.Mark || !clause.PartA || clause.LBD != 255 {
		t.Fatal(clause, err)
	}
	if got := clause.String(); got != "{ 0 -1 2 }\n" {
		t.Fatal(got)
	}
	page, pageErr := s.HandlePage(handles[2])
	offsetGot, offsetErr := s.HandleOffset(handles[2])
	words, wordsErr := s.StorageWords(3, false)
	memory, memoryErr := s.MemoryAt(handles[2])
	if pageErr != nil || offsetErr != nil || wordsErr != nil || memoryErr != nil || page != int(uint32(handles[2])>>9) || offsetGot != int(uint32(handles[2])&511) || words != 4 || memory == 0 {
		t.Fatal(page, offsetGot, words, memory, pageErr, offsetErr, wordsErr, memoryErr)
	}
	stats := s.Stats()
	if stats.PageSize != 9 || stats.PageWords != 512 || stats.ProblemEntries != 6 || stats.ProblemStored != 6 || stats.LearnedEntries != 0 || stats.LiteralCount != 502 || stats.ProblemCurrent != s.CurrentHandle(false) || stats.LearnedCurrent != SatClauseHandle(1<<9|2) {
		t.Fatal(stats)
	}
}

func TestSatClauseStoreBookmarkRollbackShrinkAndRestart(t *testing.T) {
	s, _ := NewSatClauseStore(9, SatClauseStoreLimits{})
	_, _ = s.AppendBlank(2, false, false)
	prefixProblem, _ := s.Append([]SatLit{0, 2, 4}, false, false)
	prefixLearned, _ := s.Append([]SatLit{1, 3, 5}, true, false)
	if err := s.Bookmark(); err != nil {
		t.Fatal(err)
	}
	tailProblem, _ := s.Append([]SatLit{6, 8, 10}, false, false)
	tailLearned, _ := s.Append(make([]SatLit, 490), true, false)
	for handle, want := range map[SatClauseHandle]bool{prefixProblem: true, prefixLearned: true, tailProblem: false, tailLearned: false} {
		got, err := s.ClauseUsed(handle)
		if err != nil || got != want {
			t.Fatal(handle, got, want, err)
		}
	}
	if err := s.Rollback(); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Clause(tailProblem); !errors.Is(err, ErrInvalidLiteral) {
		t.Fatal(err)
	}
	if _, err := s.Clause(tailLearned); !errors.Is(err, ErrInvalidLiteral) {
		t.Fatal(err)
	}
	replayedTail, err := s.Append(make([]SatLit, 490), true, false)
	if err != nil || replayedTail != tailLearned {
		t.Fatal("rollback did not rewind an allocated page", replayedTail, tailLearned, err)
	}
	if err = s.Rollback(); err != nil {
		t.Fatal(err)
	}
	if s.EntryCount(false) != 2 || s.EntryCount(true) != 1 || len(s.ProblemClauses()) != 1 || len(s.LearnedClauses()) != 1 {
		t.Fatal(s.Stats())
	}
	firstTail, _ := s.Append([]SatLit{12, 14}, true, false)
	secondTail, _ := s.Append([]SatLit{16, 18}, true, false)
	entries := s.EntryCount(true)
	if err := s.Shrink(firstTail, true); err != nil {
		t.Fatal(err)
	}
	if s.EntryCount(true) != entries || len(s.LearnedClauses()) != 1 {
		t.Fatal("shrink changed native entry count", s.EntryCount(true), s.LearnedClauses())
	}
	reused, _ := s.Append([]SatLit{20, 22}, true, false)
	if reused != firstTail {
		t.Fatal("shrink did not rewind handle", reused, firstTail, secondTail)
	}
	if err := s.Restart(); err != nil {
		t.Fatal(err)
	}
	if stats := s.Stats(); stats.ProblemEntries != 0 || stats.LearnedEntries != 0 || stats.ProblemCurrent != 2 || stats.LearnedCurrent != SatClauseHandle(1<<9|2) {
		t.Fatal(stats)
	}
}

func TestSatClauseStoreTwoPassCompactionAndBookmark(t *testing.T) {
	s, _ := NewSatClauseStore(9, SatClauseStoreLimits{})
	var old []SatClauseHandle
	for i := 0; i < 70; i++ {
		literals := make([]SatLit, 1+(i*17)%31)
		for j := range literals {
			literals[j] = SatLit((i+j)*2) & SatLit(math.MaxInt32)
		}
		handle, err := s.Append(literals, true, false)
		if err != nil {
			t.Fatal(i, err)
		}
		old = append(old, handle)
		if i == 27 {
			if err = s.Bookmark(); err != nil {
				t.Fatal(err)
			}
		}
	}
	for i, handle := range old {
		mark := i%4 == 0 || i == 27 || i == 28
		if err := s.SetMetadata(handle, mark, i%3 == 0, uint8(i)); err != nil {
			t.Fatal(err)
		}
	}
	before := s.LearnedClauses()
	wantHandles := satClauseTestHandles(9, before)
	wantCount := len(wantHandles)
	count, err := s.CompactLearned(false)
	if err != nil || count != wantCount {
		t.Fatal(count, wantCount, err)
	}
	planned := s.LearnedClauses()
	index := 0
	for _, clause := range planned {
		if clause.Mark {
			continue
		}
		if !clause.HasID || clause.ID != int(wantHandles[index]) || clause.Handle != old[clause.LBD] {
			t.Fatal(index, clause, wantHandles[index])
		}
		index++
	}
	count, err = s.CompactLearned(true)
	if err != nil || count != wantCount || s.EntryCount(true) != wantCount || s.LearnedCount() != wantCount {
		t.Fatal(count, s.EntryCount(true), s.LearnedCount(), err)
	}
	after := s.LearnedClauses()
	for i, clause := range after {
		if clause.Handle != wantHandles[i] || clause.ID != i || clause.Mark {
			t.Fatal(i, clause, wantHandles[i])
		}
		used, usedErr := s.ClauseUsed(clause.Handle)
		oldIndex := int(clause.LBD)
		if usedErr != nil || used != (oldIndex < 28) {
			t.Fatal(i, oldIndex, used, usedErr)
		}
	}
	if err = s.Rollback(); err != nil {
		t.Fatal(err)
	}
	for _, clause := range s.LearnedClauses() {
		if int(clause.LBD) >= 28 {
			t.Fatal("rollback retained post-bookmark clause", clause)
		}
	}
}

func TestSatClauseStoreCompactionRequiresPlanAndErrors(t *testing.T) {
	for _, pageSize := range []int{-1, 8, 32, 64} {
		if s, err := NewSatClauseStore(pageSize, SatClauseStoreLimits{}); s != nil || !errors.Is(err, ErrDimension) {
			t.Fatal(pageSize, s, err)
		}
	}
	s, _ := NewSatClauseStore(9, SatClauseStoreLimits{MaxPages: 4, MaxClauses: 2, MaxLiterals: 4})
	handle, err := s.Append([]SatLit{0, 2}, true, false)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.CompactLearned(true); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if _, err = s.Append([]SatLit{SatLit(math.MaxUint32)}, false, false); !errors.Is(err, ErrInvalidLiteral) {
		t.Fatal(err)
	}
	if _, err = s.Append([]SatLit{0, 2, 4}, false, false); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if err = s.SetClauseID(handle, -1); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if err = s.Rollback(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	s.Stop()
	s.Stop()
	if _, err = s.Append(nil, false, false); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
}

func FuzzSatClauseStoreModel(f *testing.F) {
	f.Add([]byte{3, 1, 4, 1, 5, 9, 2, 6, 5, 3, 5})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 128 {
			data = data[:128]
		}
		s, err := NewSatClauseStore(9, SatClauseStoreLimits{MaxPages: 64, MaxClauses: 256, MaxLiterals: 4096})
		if err != nil {
			t.Fatal(err)
		}
		var handles []SatClauseHandle
		for i, value := range data {
			size := int(value%23) + 1
			literals := make([]SatLit, size)
			for j := range literals {
				literals[j] = SatLit((i*31+j*7)&255) * 2
			}
			handle, appendErr := s.Append(literals, true, false)
			if appendErr != nil {
				t.Fatal(appendErr)
			}
			handles = append(handles, handle)
			if value&3 == 0 {
				if err := s.SetMetadata(handle, true, value&4 != 0, value); err != nil {
					t.Fatal(err)
				}
			}
		}
		before := s.LearnedClauses()
		want := satClauseTestHandles(9, before)
		var wantLiterals [][]SatLit
		for _, clause := range before {
			if !clause.Mark {
				wantLiterals = append(wantLiterals, clause.Literals)
			}
		}
		count, err := s.CompactLearned(false)
		if err != nil || count != len(want) {
			t.Fatal(count, len(want), err)
		}
		count, err = s.CompactLearned(true)
		if err != nil || count != len(want) {
			t.Fatal(count, len(want), err)
		}
		for i, clause := range s.LearnedClauses() {
			if clause.Handle != want[i] || clause.ID != i || clause.Mark || !reflect.DeepEqual(clause.Literals, wantLiterals[i]) {
				t.Fatal(i, clause, want[i])
			}
		}
		stats := s.Stats()
		if stats.LearnedEntries != count || stats.LearnedStored != count || stats.LiteralCount < count || stats.MemoryUsed > stats.MemoryAllocated*2 {
			t.Fatal(stats)
		}
	})
}
