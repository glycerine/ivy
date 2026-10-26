package abc

// Standalone hash models and fuzz tests for src/misc/vec/vecHsh.h.

import (
	"encoding/binary"
	"errors"
	"slices"
	"testing"
)

func hshJenkinsModel(values []int32) uint32 {
	bytes := make([]byte, 4*len(values))
	for index, value := range values {
		binary.LittleEndian.PutUint32(bytes[4*index:], uint32(value))
	}
	var hash uint32
	for _, value := range bytes {
		hash += uint32(value)
		hash += hash << 10
		hash ^= hash >> 6
	}
	hash += hash << 3
	hash ^= hash >> 11
	hash += hash << 15
	return hash
}

func hshRecordKey(values []int32) string {
	bytes := make([]byte, 4*len(values))
	for index, value := range values {
		binary.LittleEndian.PutUint32(bytes[4*index:], uint32(value))
	}
	return string(bytes)
}

func hshWeightedModel(values []int32) uint32 {
	primes := [...]uint32{4177, 5147, 5647, 6343, 7103, 7873, 8147}
	bytes := make([]byte, 4*len(values))
	for index, value := range values {
		binary.LittleEndian.PutUint32(bytes[4*index:], uint32(value))
	}
	var hash uint32
	for index, value := range bytes {
		hash += uint32(value) * primes[index%len(primes)]
	}
	return hash
}

func hshVectorModel(values []int32) uint32 {
	primes := [...]uint32{4177, 5147, 5647, 6343, 7103, 7873, 8147}
	var hash uint32
	for index, value := range values {
		hash += uint32(value) * primes[index%len(primes)]
	}
	return hash
}

func TestVectorHashFunctionsAndPrimeSchedule(t *testing.T) {
	values := []int32{-1, 0, 1, 0x1020304, -2000000000, 2000000000}
	for bins := 1; bins < 100; bins++ {
		for start := range values {
			for stop := start + 1; stop <= len(values); stop++ {
				record := values[start:stop]
				got, err := FixedIntRecordHashBin(record, bins)
				if err != nil || got != int(hshJenkinsModel(record)%uint32(bins)) {
					t.Fatal(record, bins, got, err)
				}
				if len(record) == 1 {
					single, singleErr := Int32JenkinsHashBin(record[0], bins)
					if singleErr != nil || single != got {
						t.Fatal(record, bins, single, got, singleErr)
					}
				}
				weighted, weightedErr := FixedIntRecordWeightedHashBin(record, bins)
				if weightedErr != nil || weighted != int(hshWeightedModel(record)%uint32(bins)) {
					t.Fatal(record, bins, weighted, weightedErr)
				}
				vector, vectorErr := IntVectorHashBin(record, bins)
				if vectorErr != nil || vector != int(hshVectorModel(record)%uint32(bins)) {
					t.Fatal(record, bins, vector, vectorErr)
				}
			}
		}
	}
	for _, row := range []struct{ input, want int }{{1, 1}, {2, 3}, {3, 3}, {4, 5}, {10, 11}, {11, 11}, {12, 13}} {
		got, err := hshPrimeAtLeast(row.input, 100, nil)
		if err != nil || got != row.want {
			t.Fatal(row, got, err)
		}
	}
}

func TestSingleIntInternerIdentityResizeOwnershipAndLifecycle(t *testing.T) {
	interner, err := NewSingleIntInterner(1, VectorHashLimits{MaxEntries: 100, MaxBins: 1000})
	if err != nil {
		t.Fatal(err)
	}
	input := []int32{5, 7, 7, 7, 8, 5, 3, 4}
	want := []int{0, 1, 1, 1, 2, 0, 3, 4}
	for index, value := range input {
		id, created, internErr := interner.Intern(value)
		if internErr != nil || id != want[index] || created != (index == 0 || index == 1 || index == 4 || index >= 6) {
			t.Fatal(index, id, created, internErr)
		}
		if interner.Validate() != nil {
			t.Fatal(interner.Validate())
		}
	}
	if id, found, err := interner.Lookup(8); err != nil || !found || id != 2 {
		t.Fatal(id, found, err)
	}
	if id, found, err := interner.Lookup(99); err != nil || found || id != 0 {
		t.Fatal(id, found, err)
	}
	values, err := interner.Values()
	if err != nil || !slices.Equal(values, []int32{5, 7, 8, 3, 4}) {
		t.Fatal(values, err)
	}
	values[0] = 99
	again, _ := interner.Values()
	if again[0] == 99 {
		t.Fatal("values alias manager storage")
	}
	if err = interner.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err = interner.EntryCount(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
}

func TestFixedRecordAndWordHashing(t *testing.T) {
	data := []int32{12, 17, 13, 12, 15, 3, 16, 16, 12, 17, 12, 12}
	ids, err := HashFixedIntRecords(data, 2, VectorHashLimits{})
	if err != nil || !slices.Equal(ids, []int{0, 1, 2, 3, 0, 4}) {
		t.Fatal(ids, err)
	}
	interner, err := NewFixedIntRecordInterner(data, 2, 2, VectorHashLimits{})
	if err != nil {
		t.Fatal(err)
	}
	for index, want := range ids {
		id, _, internErr := interner.InternIndex(index)
		if internErr != nil || id != want {
			t.Fatal(index, id, want, internErr)
		}
	}
	index, id, created, err := interner.AppendAndIntern([]int32{-1, 99})
	if err != nil || index != 6 || id != 5 || !created {
		t.Fatal(index, id, created, err)
	}
	duplicateIndex, duplicateID, duplicateCreated, err := interner.AppendAndIntern([]int32{12, 17})
	if err != nil || duplicateIndex != 7 || duplicateID != 0 || duplicateCreated {
		t.Fatal(duplicateIndex, duplicateID, duplicateCreated, err)
	}
	record, source, err := interner.Entry(0)
	if err != nil || source != 0 || !slices.Equal(record, []int32{12, 17}) {
		t.Fatal(record, source, err)
	}
	record[0] = 100
	again, _, _ := interner.Entry(0)
	if again[0] == 100 || interner.Validate() != nil {
		t.Fatal(again, interner.Validate())
	}

	words := []uint64{0x1122334455667788, 9, 0x1122334455667788, 10}
	wordIDs, err := HashFixedWordRecords(words, 1, VectorHashLimits{})
	if err != nil || !slices.Equal(wordIDs, []int{0, 1, 0, 2}) {
		t.Fatal(wordIDs, err)
	}
	unique, err := NewUniqueWordRecordInterner([]uint64{1, 2, 3}, 1, VectorHashLimits{})
	if err != nil {
		t.Fatal(err)
	}
	if count, err := unique.EntryCount(); err != nil || count != 3 {
		t.Fatal(count, err)
	}
	if duplicate, err := NewUniqueWordRecordInterner([]uint64{1, 2, 1}, 1, VectorHashLimits{}); duplicate != nil || !errors.Is(err, ErrInvariant) {
		t.Fatal(duplicate, err)
	}
}

func TestIntVectorInternerNaturalSequenceEmptyStorageAndLimits(t *testing.T) {
	interner, err := NewIntVectorInterner(5, VectorHashLimits{})
	if err != nil {
		t.Fatal(err)
	}
	for size := 0; size < 20; size++ {
		values := make([]int32, size)
		for index := range values {
			values[index] = int32(index)
		}
		id, created, internErr := interner.Intern(values)
		if internErr != nil || !created || id != size {
			t.Fatal(size, id, created, internErr)
		}
	}
	for size := 20; size > 0; size-- {
		values := make([]int32, size)
		for index := range values {
			values[index] = int32(index)
		}
		id, created, internErr := interner.Intern(values)
		want := size
		if size == 20 {
			want = 20
		}
		if internErr != nil || id != want || created != (size == 20) {
			t.Fatal(size, id, want, created, internErr)
		}
	}
	if count, err := interner.EntryCount(); err != nil || count != 21 {
		t.Fatal(count, err)
	}
	storage := 0
	for size := 0; size <= 20; size++ {
		storage += 2 + size + size%2
	}
	if got, err := interner.StoredInts(); err != nil || got != storage {
		t.Fatal(got, storage, err)
	}
	entry, err := interner.Entry(10)
	if err != nil || len(entry) != 10 {
		t.Fatal(entry, err)
	}
	entry[0] = 99
	again, _ := interner.Entry(10)
	if again[0] == 99 || interner.Validate() != nil {
		t.Fatal(again, interner.Validate())
	}

	limited, err := NewIntVectorInterner(1, VectorHashLimits{MaxEntries: 1, MaxDataInts: 2})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err = limited.Intern(nil); err != nil {
		t.Fatal(err)
	}
	if _, _, err = limited.Intern([]int32{1}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
}

func TestVectorHashErrorsCancellationAndBounds(t *testing.T) {
	for _, action := range []func() error{
		func() error { _, e := NewSingleIntInterner(0, VectorHashLimits{}); return e },
		func() error { _, e := NewIntVectorInterner(1, VectorHashLimits{MaxBins: -1}); return e },
		func() error { _, e := NewFixedIntRecordInterner([]int32{1}, 2, 1, VectorHashLimits{}); return e },
		func() error { _, e := FixedIntRecordHashBin(nil, 0); return e },
		func() error { _, e := IntVectorHashBin(nil, -1); return e },
		func() error {
			_, e := NewSingleIntInterner(1, VectorHashLimits{Stop: func() bool { return true }})
			return e
		},
	} {
		if err := action(); err == nil {
			t.Fatal("expected error")
		}
	}
	cancelled := false
	interner, err := NewIntVectorInterner(1, VectorHashLimits{Stop: func() bool { return cancelled }})
	if err != nil {
		t.Fatal(err)
	}
	cancelled = true
	if _, _, err = interner.Intern([]int32{1}); !errors.Is(err, ErrCancelled) {
		t.Fatal(err)
	}
}

func FuzzVectorHashInternersAgainstOrderedModels(f *testing.F) {
	f.Add([]byte{1, 2, 3, 1, 2, 255, 0})
	f.Add([]byte{0, 0, 0, 128, 127})
	f.Fuzz(func(t *testing.T, input []byte) {
		if len(input) > 512 {
			input = input[:512]
		}
		single, err := NewSingleIntInterner(3, VectorHashLimits{MaxEntries: 1024})
		vectors, vectorErr := NewIntVectorInterner(3, VectorHashLimits{MaxEntries: 1024, MaxDataInts: 1 << 16})
		fixed, fixedErr := NewFixedIntRecordInterner(nil, 2, 3, VectorHashLimits{MaxEntries: 1024, MaxDataInts: 1 << 16})
		if err != nil || vectorErr != nil || fixedErr != nil {
			t.Fatal(err, vectorErr, fixedErr)
		}
		singleIDs := map[int32]int{}
		vectorIDs := map[string]int{}
		fixedIDs := map[string]int{}
		for index, raw := range input {
			value := int32(int8(raw))
			id, created, internErr := single.Intern(value)
			want, exists := singleIDs[value]
			if !exists {
				want = len(singleIDs)
				singleIDs[value] = want
			}
			if internErr != nil || id != want || created == exists {
				t.Fatal("single", index, id, want, created, exists, internErr)
			}

			length := int(raw % 8)
			record := make([]int32, length)
			for offset := range record {
				record[offset] = int32(int8(input[(index+offset)%len(input)]))
			}
			key := hshRecordKey(record)
			id, created, internErr = vectors.Intern(record)
			want, exists = vectorIDs[key]
			if !exists {
				want = len(vectorIDs)
				vectorIDs[key] = want
			}
			if internErr != nil || id != want || created == exists {
				t.Fatal("vector", index, id, want, created, exists, internErr)
			}

			fixedRecord := []int32{value, int32(index % 17)}
			key = hshRecordKey(fixedRecord)
			_, id, created, internErr = fixed.AppendAndIntern(fixedRecord)
			want, exists = fixedIDs[key]
			if !exists {
				want = len(fixedIDs)
				fixedIDs[key] = want
			}
			if internErr != nil || id != want || created == exists {
				t.Fatal("fixed", index, id, want, created, exists, internErr)
			}
			if single.Validate() != nil || vectors.Validate() != nil || fixed.Validate() != nil {
				t.Fatal(single.Validate(), vectors.Validate(), fixed.Validate())
			}
		}
	})
}
