package abc

// Go-only state models and fuzz tests for src/misc/hash.

import (
	"errors"
	"math"
	"slices"
	"testing"
)

func hashBinModel(key, bins int) int {
	value := int32((uint32(key)+11)*uint32(key)*7 + 3)
	remainder := value % int32(bins)
	if remainder < 0 {
		remainder = -remainder
	}
	return int(remainder)
}

func stringHashModel(key string, bins int) int {
	var value uint32
	for index := 0; index < len(key); index++ {
		value = value*6 + uint32(key[index])
	}
	return int(value % uint32(bins))
}

func TestHashFunctionsAndIntMapOrderCellsLifecycle(t *testing.T) {
	for bins := 1; bins < 100; bins++ {
		for _, key := range []int{math.MinInt32, -1000001, -1, 0, 1, 11, math.MaxInt32} {
			got, err := HashDefaultBin(key, bins)
			if err != nil || got != hashBinModel(key, bins) {
				t.Fatal(key, bins, got, err)
			}
		}
		for _, key := range []string{"", "a", "abc", "ABC", "a\x00b", "longer string"} {
			got, err := HashStringBin(key, bins)
			if err != nil || got != stringHashModel(key, bins) {
				t.Fatal(key, bins, got, err)
			}
		}
	}
	if _, err := HashDefaultBin(0, 0); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err := HashStringBin("", 0); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}

	m, err := NewHashIntMap(7, HashMapLimits{})
	if err != nil {
		t.Fatal(err)
	}
	keys := []int{9, -4, 77, 2, 16, -99, 101, 9}
	model := map[int]int{}
	insertion := []int{}
	for index, key := range keys {
		if _, exists := model[key]; !exists {
			insertion = append(insertion, key)
		}
		model[key] = index * 3
		if err = m.Write(key, index*3); err != nil {
			t.Fatal(err)
		}
	}
	size, err := m.Size()
	if err != nil || size != len(model) {
		t.Fatal(size, err)
	}
	for key, want := range model {
		exists, existsErr := m.Exists(key)
		got, created, entryErr := m.Entry(key, false)
		if !exists || existsErr != nil || entryErr != nil || created || got != want {
			t.Fatal(key, got, want, exists, created, existsErr, entryErr)
		}
	}
	entries, err := m.Entries()
	if err != nil {
		t.Fatal(err)
	}
	wantEntries := []HashIntEntry{}
	for bin := 0; bin < 7; bin++ {
		for _, key := range insertion {
			if hashBinModel(key, 7) == bin {
				wantEntries = append(wantEntries, HashIntEntry{key, model[key]})
			}
		}
	}
	if !slices.Equal(entries, wantEntries) {
		t.Fatal(entries, wantEntries)
	}
	cell, err := m.EntryCell(555)
	if err != nil {
		t.Fatal(err)
	}
	if value, valueErr := cell.Get(); valueErr != nil || value != 0 {
		t.Fatal(value, valueErr)
	}
	if err = cell.Set(42); err != nil {
		t.Fatal(err)
	}
	if value, _, entryErr := m.Entry(555, false); entryErr != nil || value != 42 {
		t.Fatal(value, entryErr)
	}
	entries[0].Data = -1
	again, _ := m.Entries()
	if again[0].Data == -1 {
		t.Fatal("entries borrowed map storage")
	}
	if err = m.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err = cell.Get(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if _, err = m.Size(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
}

func TestHashFloatAndPointerMapsRemovalHandlesLimits(t *testing.T) {
	floats, err := NewHashFloatMap(3, HashMapLimits{MaxEntries: 2})
	if err != nil {
		t.Fatal(err)
	}
	cell, err := floats.EntryCell(1)
	if err != nil || cell.Set(1.25) != nil {
		t.Fatal(err)
	}
	if err = floats.Write(4, 2.5); err != nil {
		t.Fatal(err)
	}
	if err = floats.Write(7, 3.5); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	entries, err := floats.Entries()
	if err != nil || !slices.Equal(entries, []HashFloatEntry{{1, 1.25}, {4, 2.5}}) {
		t.Fatal(entries, err)
	}
	removed, err := floats.Remove(1)
	if err != nil || !removed {
		t.Fatal(removed, err)
	}
	if _, err = cell.Get(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if removed, err = floats.Remove(1); err != nil || removed {
		t.Fatal(removed, err)
	}
	if err = floats.Close(); err != nil {
		t.Fatal(err)
	}

	pointers, err := NewHashPointerMap(5, HashMapLimits{})
	if err != nil {
		t.Fatal(err)
	}
	if err = pointers.Write(-1, uintptr(0xdeadbeef)); err != nil {
		t.Fatal(err)
	}
	pointerCell, err := pointers.EntryCell(6)
	if err != nil || pointerCell.Set(uintptr(0x1234)) != nil {
		t.Fatal(err)
	}
	data, found, err := pointers.Remove(-1)
	if err != nil || !found || data != uintptr(0xdeadbeef) {
		t.Fatal(data, found, err)
	}
	if data, created, entryErr := pointers.Entry(6, false); entryErr != nil || created || data != uintptr(0x1234) {
		t.Fatal(data, created, entryErr)
	}
	if _, found, err = pointers.Remove(99); err != nil || found {
		t.Fatal(found, err)
	}
	if err = pointers.Close(); err != nil {
		t.Fatal(err)
	}
	if err = pointerCell.Set(0); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}

	for _, create := range []func() error{
		func() error { _, e := NewHashIntMap(0, HashMapLimits{}); return e },
		func() error { _, e := NewHashFloatMap(1, HashMapLimits{MaxEntries: -1}); return e },
		func() error { _, e := NewHashPointerMap(-1, HashMapLimits{}); return e },
	} {
		if err = create(); !errors.Is(err, ErrDimension) {
			t.Fatal(err)
		}
	}
}

func TestHashStringAndLongConcreteGenericReplacements(t *testing.T) {
	stringsMap, err := NewHashStringMap(4, HashMapLimits{})
	if err != nil {
		t.Fatal(err)
	}
	keys := []string{"alpha", "beta", "gamma", "delta", "alpha"}
	values := []uintptr{1, 2, 3, 4, 5}
	model := map[string]uintptr{}
	insertion := []string{}
	for index, key := range keys {
		if _, exists := model[key]; !exists {
			insertion = append(insertion, key)
		}
		model[key] = values[index]
		if err = stringsMap.Write(key, values[index]); err != nil {
			t.Fatal(err)
		}
	}
	entries, err := stringsMap.Entries()
	if err != nil {
		t.Fatal(err)
	}
	want := []HashStringEntry{}
	for bin := 0; bin < 4; bin++ {
		for _, key := range insertion {
			if stringHashModel(key, 4) == bin {
				want = append(want, HashStringEntry{key, model[key]})
			}
		}
	}
	if !slices.Equal(entries, want) {
		t.Fatal(entries, want)
	}
	entry, found, created, err := stringsMap.Entry("beta", false)
	if err != nil || !found || created || entry != (HashStringEntry{"beta", 2}) {
		t.Fatal(entry, found, created, err)
	}
	entry, found, created, err = stringsMap.Entry("new", true)
	if err != nil || !found || !created || entry != (HashStringEntry{"new", 0}) {
		t.Fatal(entry, found, created, err)
	}
	data, found, err := stringsMap.Remove("alpha")
	if err != nil || !found || data != 5 {
		t.Fatal(data, found, err)
	}
	if err = stringsMap.Close(); err != nil {
		t.Fatal(err)
	}

	longs, err := NewHashLongMap(3, HashMapLimits{})
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range []HashLongEntry{{1, 10}, {4, 40}, {2, 20}, {7, 70}} {
		if err = longs.Write(row.Key, row.Data); err != nil {
			t.Fatal(err)
		}
	}
	longEntries, err := longs.Entries()
	if err != nil || !slices.Equal(longEntries, []HashLongEntry{{1, 10}, {4, 40}, {7, 70}, {2, 20}}) {
		t.Fatal(longEntries, err)
	}
	if data, found, err = longs.Remove(4); err != nil || !found || data != 40 {
		t.Fatal(data, found, err)
	}
	if err = longs.Close(); err != nil {
		t.Fatal(err)
	}

	generic, err := NewHashGenericMap(5,
		func(key uintptr, bins int) int { return int(key&15) % bins },
		func(first, second uintptr) int {
			if first&15 == second&15 {
				return 0
			}
			if first&15 < second&15 {
				return -1
			}
			return 1
		}, HashMapLimits{})
	if err != nil {
		t.Fatal(err)
	}
	if err = generic.Write(0x21, 10); err != nil {
		t.Fatal(err)
	}
	if err = generic.Write(0x31, 20); err != nil {
		t.Fatal(err)
	}
	cell, created, err := generic.EntryCell(0x41, false)
	if err != nil || created {
		t.Fatal(created, err)
	}
	if key, keyErr := cell.Key(); keyErr != nil || key != 0x21 {
		t.Fatal(key, keyErr)
	}
	if data, dataErr := cell.Get(); dataErr != nil || data != 20 {
		t.Fatal(data, dataErr)
	}
	if err = cell.Set(30); err != nil {
		t.Fatal(err)
	}
	if entries, entriesErr := generic.Entries(); entriesErr != nil || !slices.Equal(entries, []HashGenericEntry{{0x21, 30}}) {
		t.Fatal(entries, entriesErr)
	}
	if data, found, removeErr := generic.Remove(0x51); removeErr != nil || !found || data != 30 {
		t.Fatal(data, found, removeErr)
	}
	if _, err = cell.Get(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if err = generic.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err = NewHashGenericMap(1, nil, func(first, second uintptr) int { return 0 }, HashMapLimits{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	bad, _ := NewHashGenericMap(2, func(key uintptr, bins int) int { return bins }, func(first, second uintptr) int { return 0 }, HashMapLimits{})
	if _, err = bad.Exists(0); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
}

func FuzzHashConcreteMapsAgainstModels(f *testing.F) {
	f.Add([]byte{0, 1, 2, 3, 4, 5, 6, 7})
	f.Add([]byte{255, 0, 255, 17, 17})
	f.Fuzz(func(t *testing.T, operations []byte) {
		if len(operations) > 256 {
			operations = operations[:256]
		}
		integers, err := NewHashIntMap(11, HashMapLimits{MaxEntries: 64})
		if err != nil {
			t.Fatal(err)
		}
		floats, err := NewHashFloatMap(11, HashMapLimits{MaxEntries: 64})
		if err != nil {
			t.Fatal(err)
		}
		intModel := map[int]int{}
		floatModel := map[int]float32{}
		for index, operation := range operations {
			key := int(int8(operation))
			switch operation % 4 {
			case 0:
				value := index*31 - key
				if err = integers.Write(key, value); err != nil {
					t.Fatal(err)
				}
				intModel[key] = value
			case 1:
				value := float32(index) + float32(key)/17
				if err = floats.Write(key, value); err != nil {
					t.Fatal(err)
				}
				floatModel[key] = value
			case 2:
				removed, removeErr := floats.Remove(key)
				_, existed := floatModel[key]
				if removeErr != nil || removed != existed {
					t.Fatal(index, key, removed, existed, removeErr)
				}
				delete(floatModel, key)
			case 3:
				got, _, entryErr := integers.Entry(key, false)
				want := intModel[key]
				if entryErr != nil || got != want {
					t.Fatal(index, key, got, want, entryErr)
				}
			}
		}
		intEntries, err := integers.Entries()
		if err != nil || len(intEntries) != len(intModel) {
			t.Fatal(len(intEntries), len(intModel), err)
		}
		for _, entry := range intEntries {
			if intModel[entry.Key] != entry.Data {
				t.Fatal(entry, intModel[entry.Key])
			}
		}
		floatEntries, err := floats.Entries()
		if err != nil || len(floatEntries) != len(floatModel) {
			t.Fatal(len(floatEntries), len(floatModel), err)
		}
		for _, entry := range floatEntries {
			if floatModel[entry.Key] != entry.Data {
				t.Fatal(entry, floatModel[entry.Key])
			}
		}
	})
}
