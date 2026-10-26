package abc

// Stand-alone tests for the owned port of src/sat/bsat/satProof2.h.

import (
	"errors"
	"reflect"
	"slices"
	"testing"
)

func satProofTestRecord(t testing.TB, manager *SatProofInfoManager, index int, bits ...int) {
	t.Helper()
	record, err := manager.Record(index)
	if err != nil {
		t.Fatal(err)
	}
	want := make([]uint64, len(record))
	for _, bit := range bits {
		want[bit>>6] |= uint64(1) << uint(bit&63)
	}
	if !slices.Equal(record, want) {
		t.Fatal(index, record, want)
	}
}

func satProofTestChain(t testing.TB, manager *SatProofInfoManager, first SatProofClauseRef, rest ...SatProofClauseRef) {
	t.Helper()
	if err := manager.ChainStart(first); err != nil {
		t.Fatal(err)
	}
	for _, reference := range rest {
		if err := manager.ChainResolve(reference); err != nil {
			t.Fatal(err)
		}
	}
	if result, err := manager.ChainStop(); err != nil || result != 0 {
		t.Fatal(result, err)
	}
}

func TestSatProofInfoDirectChainsGrowAndCore(t *testing.T) {
	for _, test := range []struct{ width, words int }{{0, 0}, {1, 1}, {64, 1}, {65, 2}, {128, 2}, {129, 3}} {
		words, err := SatProofBitWordCount(test.width)
		if err != nil || words != test.words {
			t.Fatal(test, words, err)
		}
	}
	manager, err := NewSatProofInfoManager(SatProofInfoLimits{})
	if err != nil || manager.Stats().FirstClause != -1 {
		t.Fatal(manager, err)
	}
	if err = manager.Restart(nil, 10, 65); err != nil {
		t.Fatal(err)
	}
	satProofTestChain(t, manager, SatProofClauseRef{ID: 1}, SatProofClauseRef{ID: 64})
	satProofTestChain(t, manager, SatProofClauseRef{Learned: true, ID: 10}, SatProofClauseRef{ID: 5}, SatProofClauseRef{Learned: true, ID: 9})
	satProofTestRecord(t, manager, 0, 1, 64)
	satProofTestRecord(t, manager, 1, 1, 5, 64)
	core, err := manager.UnsatCore()
	if err != nil || !slices.Equal(core, []int{64, 5, 1}) {
		t.Fatal(core, err)
	}
	// Native grows even at the exact current bit capacity and doubles words.
	if err = manager.Grow(128); err != nil || manager.Stats().WordCount != 4 || manager.Stats().WidthCapacity != 256 {
		t.Fatal(manager.Stats(), err)
	}
	satProofTestRecord(t, manager, 1, 1, 5, 64)
	if stats := manager.Stats(); stats.Records != 2 || stats.OwnedBytes == 0 || stats.Active {
		t.Fatal(stats)
	}
}

func TestSatProofInfoOwnedMappingAndUpdates(t *testing.T) {
	mapping := []int{2, -1, 0, 65}
	manager, _ := NewSatProofInfoManager(SatProofInfoLimits{})
	if err := manager.Restart(mapping, 20, 66); err != nil {
		t.Fatal(err)
	}
	mapping[0] = 64
	if err := manager.SetProofID(1, 1); err != nil {
		t.Fatal(err)
	}
	satProofTestChain(t, manager, SatProofClauseRef{ID: 0}, SatProofClauseRef{ID: 1}, SatProofClauseRef{ID: 3})
	satProofTestRecord(t, manager, 0, 1, 2, 65)
	core, err := manager.UnsatCore()
	if err != nil || !slices.Equal(core, []int{3, 1, 0}) {
		t.Fatal(core, err)
	}
	ids := manager.ProofIDs()
	ids[0] = -1
	if manager.ProofIDs()[0] != 2 {
		t.Fatal("mapping snapshot aliases manager")
	}
	bad, _ := NewSatProofInfoManager(SatProofInfoLimits{})
	if err = bad.Restart([]int{0}, 0, 1); err != nil {
		t.Fatal(err)
	}
	if err = bad.ChainStart(SatProofClauseRef{ID: 1}); !errors.Is(err, ErrInvalidLiteral) || bad.Size() != 0 || bad.Stats().Active {
		t.Fatal(err, bad.Stats())
	}
}

func TestSatProofInfoShrinkSaveCompactAndLifecycle(t *testing.T) {
	manager, _ := NewSatProofInfoManager(SatProofInfoLimits{})
	if err := manager.Restart(nil, 100, 16); err != nil {
		t.Fatal(err)
	}
	for bit := 0; bit < 3; bit++ {
		satProofTestChain(t, manager, SatProofClauseRef{ID: bit})
	}
	if err := manager.AddSaved(99, 0); err != nil || manager.Stats().Saved != 0 {
		t.Fatal(err, manager.Stats())
	}
	if err := manager.AddSaved(101, 7); err != nil {
		t.Fatal(err)
	}
	if err := manager.AddSaved(102, 8); err != nil {
		t.Fatal(err)
	}
	if err := manager.Compact(9); err != nil {
		t.Fatal(err)
	}
	if stats := manager.Stats(); stats.FirstClause != 7 || stats.Records != 2 || stats.Saved != 0 {
		t.Fatal(stats)
	}
	satProofTestRecord(t, manager, 0, 1)
	satProofTestRecord(t, manager, 1, 2)
	satProofTestChain(t, manager, SatProofClauseRef{Learned: true, ID: 7}, SatProofClauseRef{ID: 3})
	satProofTestRecord(t, manager, 2, 1, 3)
	if err := manager.Shrink(8); err != nil || manager.Size() != 1 {
		t.Fatal(err, manager.Size())
	}
	if err := manager.Compact(20); err != nil || manager.Size() != 0 || manager.Stats().FirstClause != 20 {
		t.Fatal(err, manager.Stats())
	}
	manager.Stop()
	manager.Stop()
	if err := manager.Grow(128); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
}

func TestSatProofInfoErrorsAndLimits(t *testing.T) {
	if _, err := SatProofBitWordCount(-1); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if manager, err := NewSatProofInfoManager(SatProofInfoLimits{MaxWidth: -1}); manager != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(manager, err)
	}
	manager, _ := NewSatProofInfoManager(SatProofInfoLimits{MaxWidth: 256, MaxRecords: 1, MaxWords: 2})
	for _, restart := range []struct {
		mapping []int
		first   int
		width   int
	}{{nil, -1, 1}, {nil, 0, 0}, {[]int{64}, 0, 1}} {
		fresh, _ := NewSatProofInfoManager(SatProofInfoLimits{})
		if err := fresh.Restart(restart.mapping, restart.first, restart.width); err == nil {
			t.Fatal(restart)
		}
	}
	if err := manager.Restart(nil, 0, 64); err != nil {
		t.Fatal(err)
	}
	if err := manager.Restart(nil, 0, 64); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if _, err := manager.UnsatCore(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if err := manager.ChainResolve(SatProofClauseRef{}); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	satProofTestChain(t, manager, SatProofClauseRef{ID: 1})
	if err := manager.ChainStart(SatProofClauseRef{ID: 2}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if err := manager.Grow(129); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if err := manager.AddSaved(1, 0); !errors.Is(err, ErrInvalidLiteral) {
		t.Fatal(err)
	}
}

func FuzzSatProofInfoChainsAndCompaction(f *testing.F) {
	f.Add([]byte{3, 1, 4, 1, 5, 9, 2, 6})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 128 {
			data = data[:128]
		}
		manager, err := NewSatProofInfoManager(SatProofInfoLimits{MaxWidth: 256, MaxRecords: 256, MaxWords: 1024})
		if err != nil || manager.Restart(nil, 50, 128) != nil {
			t.Fatal(err)
		}
		var model [][]uint64
		for i, value := range data {
			record := make([]uint64, 2)
			firstBit := int(value) & 127
			record[firstBit>>6] |= uint64(1) << uint(firstBit&63)
			if err := manager.ChainStart(SatProofClauseRef{ID: firstBit}); err != nil {
				t.Fatal(err)
			}
			if len(model) > 0 && value&1 != 0 {
				index := int(value) % len(model)
				if err := manager.ChainResolve(SatProofClauseRef{Learned: true, ID: 50 + index}); err != nil {
					t.Fatal(err)
				}
				for word := range record {
					record[word] |= model[index][word]
				}
			}
			secondBit := (i*37 + int(value)) & 127
			if err := manager.ChainResolve(SatProofClauseRef{ID: secondBit}); err != nil {
				t.Fatal(err)
			}
			record[secondBit>>6] |= uint64(1) << uint(secondBit&63)
			if _, err := manager.ChainStop(); err != nil {
				t.Fatal(err)
			}
			model = append(model, record)
			got, _ := manager.Record(i)
			if !slices.Equal(got, record) {
				t.Fatal(i, got, record)
			}
		}
		var saved [][]uint64
		firstNew := -1
		for i, value := range data {
			if value&2 == 0 {
				continue
			}
			newID := 200 + len(saved)
			if firstNew == -1 {
				firstNew = newID
			}
			if err := manager.AddSaved(50+i, newID); err != nil {
				t.Fatal(err)
			}
			saved = append(saved, model[i])
		}
		if err := manager.Compact(300); err != nil {
			t.Fatal(err)
		}
		wantFirst := 300
		if firstNew >= 0 {
			wantFirst = firstNew
		}
		if manager.Size() != len(saved) || manager.Stats().FirstClause != wantFirst {
			t.Fatal(manager.Stats(), len(saved), wantFirst)
		}
		for i, want := range saved {
			got, err := manager.Record(i)
			if err != nil || !reflect.DeepEqual(got, want) {
				t.Fatal(i, got, want, err)
			}
		}
	})
}
