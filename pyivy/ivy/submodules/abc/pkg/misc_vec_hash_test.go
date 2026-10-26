package abc

// Standalone unit, model, and fuzz tests for src/misc/vec/vecHash.h.

import (
	"errors"
	"math"
	"slices"
	"testing"
)

func tuplePairBinModel(data0, data1, bins int) int {
	return int((uint32(4177)*uint32(int32(data0)) + uint32(7873)*uint32(int32(data1))) % uint32(bins))
}

func tupleTripleBinModel(data0, data1, data2, bins int) int {
	return int((uint32(4177)*uint32(int32(data0)) + uint32(7873)*uint32(int32(data1)) + uint32(1699)*uint32(int32(data2))) % uint32(bins))
}

func TestIntTupleHashFunctionsAndPairSemantics(t *testing.T) {
	values := []int{math.MinInt32, -1000001, -1, 0, 1, 1000001, math.MaxInt32}
	for bins := 1; bins < 100; bins++ {
		for _, first := range values {
			for _, second := range values {
				pair, err := IntPairHashBin(first, second, bins)
				triple, tripleErr := IntTripleHashBin(first, second, first^second, bins)
				if err != nil || tripleErr != nil || pair != tuplePairBinModel(first, second, bins) || triple != tupleTripleBinModel(first, second, first^second, bins) {
					t.Fatal(first, second, bins, pair, triple, err, tripleErr)
				}
			}
		}
	}
	pairs, err := NewIntPairHash(0, IntTupleHashLimits{})
	if err != nil {
		t.Fatal(err)
	}
	rows := []IntTupleHashEntry{{10, 11, 12}, {20, 21, 22}, {10, 11, 99}, {-1, math.MaxInt32, -7}}
	wantIDs := []int{1, 2, 1, 3}
	wantCreated := []bool{true, true, false, true}
	for index, row := range rows {
		id, created, internErr := pairs.InternPair(row.Data0, row.Data1, row.Data2)
		if internErr != nil || id != wantIDs[index] || created != wantCreated[index] {
			t.Fatal(index, id, created, internErr)
		}
	}
	entry, err := pairs.Entry(1)
	if err != nil || entry != rows[0] {
		t.Fatal(entry, err)
	}
	if old, err := pairs.PairPayloadIncrement(1); err != nil || old != 12 {
		t.Fatal(old, err)
	}
	if now, err := pairs.PairPayloadDecrement(1); err != nil || now != 12 {
		t.Fatal(now, err)
	}
	if err = pairs.SetPairPayload(1, 44); err != nil {
		t.Fatal(err)
	}
	if id, found, lookupErr := pairs.LookupPair(10, 11); lookupErr != nil || !found || id != 1 {
		t.Fatal(id, found, lookupErr)
	}
	if id, found, lookupErr := pairs.LookupPair(99, 11); lookupErr != nil || found || id != 0 {
		t.Fatal(id, found, lookupErr)
	}
	entries, err := pairs.Entries()
	if err != nil || !slices.Equal(entries, []IntTupleHashEntry{{10, 11, 44}, {20, 21, 22}, {-1, math.MaxInt32, -7}}) {
		t.Fatal(entries, err)
	}
	entries[0].Data0 = 999
	entry, _ = pairs.Entry(1)
	if entry.Data0 == 999 {
		t.Fatal("entries alias manager storage")
	}
	profile, err := pairs.Profile()
	if err != nil {
		t.Fatal(err)
	}
	total := 0
	for _, count := range profile {
		total += count
	}
	if total != 3 || pairs.Validate() != nil {
		t.Fatal(total, pairs.Validate())
	}
}

func TestIntTupleHashResizeTriplesLimitsAndLifecycle(t *testing.T) {
	pairResize, err := NewIntPairHash(0, IntTupleHashLimits{MaxEntries: 200, MaxBins: 1000})
	if err != nil {
		t.Fatal(err)
	}
	for index := 0; index < 101; index++ {
		if _, _, err = pairResize.InternPair(index, -index, index); err != nil {
			t.Fatal(index, err)
		}
	}
	beforeProfile, err := pairResize.Profile()
	if err != nil || len(beforeProfile) != 101 {
		t.Fatal(len(beforeProfile), err)
	}
	if id, created, duplicateErr := pairResize.InternPair(0, 0, 999); duplicateErr != nil || created || id != 1 {
		t.Fatal(id, created, duplicateErr)
	}
	afterProfile, err := pairResize.Profile()
	if err != nil || len(afterProfile) != 211 {
		t.Fatal(len(afterProfile), err)
	}

	triples, err := NewIntTripleHash(0, IntTupleHashLimits{MaxEntries: 300, MaxBins: 1000})
	if err != nil {
		t.Fatal(err)
	}
	for index := 0; index < 220; index++ {
		id, created, internErr := triples.InternTriple(index%19-9, index/19, index*17-1000)
		if internErr != nil || !created || id != index+1 {
			t.Fatal(index, id, created, internErr)
		}
	}
	for index := 0; index < 220; index++ {
		id, found, lookupErr := triples.LookupTriple(index%19-9, index/19, index*17-1000)
		if lookupErr != nil || !found || id != index+1 {
			t.Fatal(index, id, found, lookupErr)
		}
	}
	if id, created, err := triples.InternTriple(1, 2, 3); err != nil || !created || id != 221 {
		t.Fatal(id, created, err)
	}
	if id, created, err := triples.InternTriple(1, 2, 4); err != nil || !created || id != 222 {
		t.Fatal(id, created, err)
	}
	if triples.Validate() != nil {
		t.Fatal(triples.Validate())
	}
	if _, _, err = triples.InternPair(1, 2, 3); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if err = triples.SetPairPayload(1, 0); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}

	limited, err := NewIntPairHash(0, IntTupleHashLimits{MaxEntries: 2})
	if err != nil {
		t.Fatal(err)
	}
	_, _, _ = limited.InternPair(1, 1, 0)
	_, _, _ = limited.InternPair(2, 2, 0)
	if _, _, err = limited.InternPair(3, 3, 0); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if count, countErr := limited.EntryCount(); countErr != nil || count != 2 {
		t.Fatal(count, countErr)
	}
	cancelled := false
	cancelHash, err := NewIntPairHash(0, IntTupleHashLimits{Stop: func() bool { return cancelled }})
	if err != nil {
		t.Fatal(err)
	}
	cancelled = true
	if _, _, err = cancelHash.InternPair(1, 2, 3); !errors.Is(err, ErrCancelled) {
		t.Fatal(err)
	}
	if err = cancelHash.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err = cancelHash.EntryCount(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	for _, create := range []func() error{
		func() error { _, e := NewIntPairHash(-1, IntTupleHashLimits{}); return e },
		func() error { _, e := NewIntTripleHash(0, IntTupleHashLimits{MaxBins: 99}); return e },
		func() error { _, e := NewIntPairHash(0, IntTupleHashLimits{MaxEntries: -1}); return e },
		func() error { _, e := IntPairHashBin(0, 0, 0); return e },
		func() error { _, e := IntTripleHashBin(0, 0, 0, -1); return e },
	} {
		if err = create(); err == nil {
			t.Fatal("expected error")
		}
	}
	if _, err = pairsForTupleBoundary(t, math.MaxInt32).PairPayloadIncrement(1); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, err = pairsForTupleBoundary(t, math.MinInt32).PairPayloadDecrement(1); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
}

func pairsForTupleBoundary(t *testing.T, payload int) *IntTupleHash {
	t.Helper()
	hash, err := NewIntPairHash(0, IntTupleHashLimits{})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err = hash.InternPair(0, 0, payload); err != nil {
		t.Fatal(err)
	}
	return hash
}

func FuzzIntTupleHashAgainstOrderedModel(f *testing.F) {
	f.Add([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8})
	f.Add([]byte{255, 0, 255, 1, 254, 2, 253})
	f.Fuzz(func(t *testing.T, operations []byte) {
		if len(operations) > 512 {
			operations = operations[:512]
		}
		pairs, err := NewIntPairHash(0, IntTupleHashLimits{MaxEntries: 1024})
		if err != nil {
			t.Fatal(err)
		}
		triples, err := NewIntTripleHash(0, IntTupleHashLimits{MaxEntries: 1024})
		if err != nil {
			t.Fatal(err)
		}
		type pairKey struct{ first, second int }
		ids := make(map[pairKey]int)
		entries := []IntTupleHashEntry{}
		tripleIDs := make(map[[3]int]int)
		tripleEntries := []IntTupleHashEntry{}
		for index, operation := range operations {
			first := int(int8(operation))
			second := index%31 - 15
			third := int(int8(operation ^ byte(index)))
			key := pairKey{first, second}
			switch operation % 4 {
			case 0, 1:
				id, created, internErr := pairs.InternPair(first, second, index)
				want, exists := ids[key]
				if !exists {
					want = len(entries) + 1
					ids[key] = want
					entries = append(entries, IntTupleHashEntry{first, second, index})
				}
				if internErr != nil || id != want || created == exists {
					t.Fatal(index, id, want, created, exists, internErr)
				}
			case 2:
				id, found, lookupErr := pairs.LookupPair(first, second)
				want, exists := ids[key]
				if lookupErr != nil || id != want || found != exists {
					t.Fatal(index, id, want, found, exists, lookupErr)
				}
			case 3:
				id, exists := ids[key]
				if exists {
					if setErr := pairs.SetPairPayload(id, -index); setErr != nil {
						t.Fatal(index, setErr)
					}
					entries[id-1].Data2 = -index
				}
			}
			if validateErr := pairs.Validate(); validateErr != nil {
				t.Fatal(index, validateErr)
			}
			tripleKey := [3]int{first, second, third}
			if operation&1 == 0 {
				id, created, internErr := triples.InternTriple(first, second, third)
				want, exists := tripleIDs[tripleKey]
				if !exists {
					want = len(tripleEntries) + 1
					tripleIDs[tripleKey] = want
					tripleEntries = append(tripleEntries, IntTupleHashEntry{first, second, third})
				}
				if internErr != nil || id != want || created == exists {
					t.Fatal("triple intern", index, id, want, created, exists, internErr)
				}
			} else {
				id, found, lookupErr := triples.LookupTriple(first, second, third)
				want, exists := tripleIDs[tripleKey]
				if lookupErr != nil || id != want || found != exists {
					t.Fatal("triple lookup", index, id, want, found, exists, lookupErr)
				}
			}
			if validateErr := triples.Validate(); validateErr != nil {
				t.Fatal("triple validate", index, validateErr)
			}
		}
		got, err := pairs.Entries()
		if err != nil || !slices.Equal(got, entries) {
			t.Fatal(got, entries, err)
		}
		gotTriples, err := triples.Entries()
		if err != nil || !slices.Equal(gotTriples, tripleEntries) {
			t.Fatal(gotTriples, tripleEntries, err)
		}
	})
}
