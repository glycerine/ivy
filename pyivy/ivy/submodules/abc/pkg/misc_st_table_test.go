package abc

// Tests for src/misc/st/st.c and src/misc/st/stmm.c.

import (
	"encoding/json"
	"errors"
	"math"
	"os"
	"reflect"
	"slices"
	"strings"
	"testing"
)

type stableTableOracle struct {
	Hashes        [][2]int `json:"hashes"`
	IntegerHashes []int    `json:"integer_hashes"`
	AddressHashes []int    `json:"address_hashes"`
	ST            struct {
		Insert      []int    `json:"insert"`
		Alpha       int      `json:"alpha"`
		FindMissing int      `json:"find_missing"`
		FindAdd     int      `json:"find_add"`
		FindAgain   []int    `json:"find_again"`
		Duplicate   []int    `json:"duplicate"`
		Count       int      `json:"count"`
		CopyAlpha   int      `json:"copy_alpha"`
		NativeOrder []string `json:"native_order"`
		ForEach     []int    `json:"foreach"`
		CopyCount   int      `json:"copy_count"`
	} `json:"st"`
	STMM struct {
		Insert      []int    `json:"insert"`
		FindAdd     int      `json:"find_add"`
		FindAgain   []int    `json:"find_again"`
		Duplicate   []int    `json:"duplicate"`
		Count       int      `json:"count"`
		NativeOrder []string `json:"native_order"`
		ForEach     []int    `json:"foreach"`
		CleanCount  int      `json:"clean_count"`
		Copy        []int    `json:"copy"`
	} `json:"stmm"`
}

func stableTableValue[K comparable, V any](t testing.TB, table *StableTable[K, V], key K) V {
	t.Helper()
	value, found, err := table.Get(key)
	if err != nil || !found {
		t.Fatal(key, value, found, err)
	}
	return value
}

func TestStableTableNative(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/st.json")
	if err != nil {
		t.Fatal(err)
	}
	var oracle stableTableOracle
	if err = json.Unmarshal(data, &oracle); err != nil {
		t.Fatal(err)
	}
	texts := []string{"", "a", "abc", "abcdefghijklmnopqrstuvwxyz", "ABC123!?", "012345678901234567890123456789", "\x80\xffA"}
	for index, value := range texts {
		ordinary, ordinaryErr := StableStringHash(value, 97)
		pooled, pooledErr := StablePooledStringHash(value, 97)
		if ordinaryErr != nil || pooledErr != nil || [2]int{ordinary, pooled} != oracle.Hashes[index] {
			t.Fatal(index, ordinary, pooled, ordinaryErr, pooledErr, oracle.Hashes[index])
		}
	}
	for index, value := range []int64{-123, 0, 123, 2147483647} {
		got, hashErr := StableIntegerHash(value, 97)
		if hashErr != nil || got != oracle.IntegerHashes[index] {
			t.Fatal(index, got, hashErr, oracle.IntegerHashes[index])
		}
	}
	for index, value := range []uintptr{0, 4, 4096, 0x12345} {
		got, hashErr := StableAddressHash(value, 97)
		if hashErr != nil || got != oracle.AddressHashes[index] {
			t.Fatal(index, got, hashErr, oracle.AddressHashes[index])
		}
	}

	table, _ := NewStableTable[string, int](1)
	insert := make([]int, 0, 3)
	for _, row := range []struct {
		key   string
		value int
	}{{"alpha", 10}, {"beta", 20}, {"alpha", 11}} {
		replaced, setErr := table.Set(row.key, row.value)
		if setErr != nil {
			t.Fatal(setErr)
		}
		if replaced {
			insert = append(insert, 1)
		} else {
			insert = append(insert, 0)
		}
	}
	if !slices.Equal(insert, oracle.ST.Insert) || stableTableValue(t, table, "alpha") != oracle.ST.Alpha {
		t.Fatal(insert, oracle.ST.Insert)
	}
	if slot, found, findErr := table.FindSlot("gamma"); slot != nil || found != (oracle.ST.FindMissing != 0) || findErr != nil {
		t.Fatal(slot, found, findErr)
	}
	slot, existing, err := table.FindOrAdd("gamma")
	if err != nil || existing != (oracle.ST.FindAdd != 0) || slot.Set(30) != nil {
		t.Fatal(existing, err)
	}
	slot, existing, err = table.FindOrAdd("gamma")
	value, valueErr := slot.Value()
	if err != nil || valueErr != nil || existing != (oracle.ST.FindAgain[0] != 0) || value != oracle.ST.FindAgain[1] {
		t.Fatal(existing, value, err, valueErr)
	}
	_ = table.AddDirect("dup", 40)
	_ = table.AddDirect("dup", 50)
	lookup := stableTableValue(t, table, "dup")
	_, deleted, found, err := table.Delete("dup")
	after := stableTableValue(t, table, "dup")
	if err != nil || !found || !slices.Equal([]int{lookup, deleted, after}, oracle.ST.Duplicate) {
		t.Fatal(lookup, deleted, after, found, err)
	}
	count, _ := table.Len()
	if count != oracle.ST.Count {
		t.Fatal(count, oracle.ST.Count)
	}
	clone, _ := table.Clone()
	_, _ = table.Set("alpha", 99)
	if stableTableValue(t, clone, "alpha") != oracle.ST.CopyAlpha {
		t.Fatal(stableTableValue(t, clone, "alpha"))
	}
	entries, _ := table.Entries()
	wantEntries := []StableTableItem[string, int]{{"alpha", 99}, {"beta", 20}, {"gamma", 30}, {"dup", 40}}
	if !reflect.DeepEqual(entries, wantEntries) {
		t.Fatal(entries, wantEntries)
	}
	cursor, _ := table.Cursor()
	cursorEntries := make([]StableTableItem[string, int], 0, len(wantEntries))
	for {
		key, cursorValue, found, cursorErr := cursor.Next()
		if cursorErr != nil {
			t.Fatal(cursorErr)
		}
		if !found {
			break
		}
		cursorEntries = append(cursorEntries, StableTableItem[string, int]{Key: key, Value: cursorValue})
	}
	if !reflect.DeepEqual(cursorEntries, wantEntries) || cursor.Close() != nil {
		t.Fatal(cursorEntries, wantEntries)
	}
	if !slices.Equal(oracle.ST.NativeOrder, []string{"gamma:30", "alpha:99", "dup:40", "beta:20"}) {
		t.Fatal("native bucket-order evidence changed", oracle.ST.NativeOrder)
	}
	completed, err := table.ForEach(func(_ string, value int) StableTableAction {
		if value&1 == 0 {
			return StableTableDelete
		}
		return StableTableContinue
	})
	count, _ = table.Len()
	cloneCount, _ := clone.Len()
	if err != nil || completed != (oracle.ST.ForEach[0] != 0) || count != oracle.ST.ForEach[1] || cloneCount != oracle.ST.CopyCount {
		t.Fatal(completed, count, cloneCount, err)
	}

	memory, _ := NewStableTable[int, int](1)
	insert = insert[:0]
	for _, row := range [][2]int{{1, 10}, {2, 20}, {1, 11}} {
		replaced, setErr := memory.Set(row[0], row[1])
		if setErr != nil {
			t.Fatal(setErr)
		}
		insert = append(insert, map[bool]int{false: 0, true: 1}[replaced])
	}
	if !slices.Equal(insert, oracle.STMM.Insert) {
		t.Fatal(insert, oracle.STMM.Insert)
	}
	memorySlot, memoryExisting, _ := memory.FindOrAdd(3)
	_ = memorySlot.Set(30)
	if memoryExisting != (oracle.STMM.FindAdd != 0) {
		t.Fatal(memoryExisting)
	}
	memorySlot, memoryExisting, _ = memory.FindOrAdd(3)
	memoryValue, _ := memorySlot.Value()
	if memoryExisting != (oracle.STMM.FindAgain[0] != 0) || memoryValue != oracle.STMM.FindAgain[1] {
		t.Fatal(memoryExisting, memoryValue)
	}
	_ = memory.AddDirect(2, 22)
	lookup = stableTableValue(t, memory, 2)
	_, deleted, found, err = memory.Delete(2)
	after = stableTableValue(t, memory, 2)
	if err != nil || !found || !slices.Equal([]int{lookup, deleted, after}, oracle.STMM.Duplicate) {
		t.Fatal(lookup, deleted, after, found, err)
	}
	memoryClone, _ := memory.Clone()
	memoryEntries, _ := memory.Entries()
	if !reflect.DeepEqual(memoryEntries, []StableTableItem[int, int]{{1, 11}, {2, 20}, {3, 30}}) {
		t.Fatal(memoryEntries)
	}
	if !slices.Equal(oracle.STMM.NativeOrder, []string{"1:11", "2:20", "3:30"}) {
		t.Fatal("native bucket-order evidence changed", oracle.STMM.NativeOrder)
	}
	completed, err = memory.ForEach(func(_ int, value int) StableTableAction {
		if value >= 20 {
			return StableTableDelete
		}
		return StableTableContinue
	})
	count, _ = memory.Len()
	if err != nil || completed != (oracle.STMM.ForEach[0] != 0) || count != oracle.STMM.ForEach[1] {
		t.Fatal(completed, count, err)
	}
	if err = memory.Clear(); err != nil {
		t.Fatal(err)
	}
	count, _ = memory.Len()
	cloneCount, _ = memoryClone.Len()
	if count != oracle.STMM.CleanCount || cloneCount != oracle.STMM.Copy[0] || stableTableValue(t, memoryClone, 1) != oracle.STMM.Copy[1] {
		t.Fatal(count, cloneCount)
	}
}

func TestStableTableSlotsCursorsAndErrors(t *testing.T) {
	if table, err := NewStableTable[int, int](-1); table != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(table, err)
	}
	table, _ := NewStableTable[int, int](0)
	slot, existing, err := table.FindOrAdd(1)
	if err != nil || existing {
		t.Fatal(existing, err)
	}
	if value, valueErr := slot.Value(); value != 0 || valueErr != nil {
		t.Fatal(value, valueErr)
	}
	_ = slot.Set(10)
	cursor, _ := table.Cursor()
	if _, _, found, nextErr := cursor.Next(); !found || nextErr != nil {
		t.Fatal(found, nextErr)
	}
	_, _ = table.Set(1, 11)
	if _, _, found, nextErr := cursor.Next(); found || nextErr != nil {
		t.Fatal(found, nextErr)
	}
	_ = table.AddDirect(2, 20)
	if _, _, _, nextErr := cursor.Next(); !errors.Is(nextErr, ErrInvariant) {
		t.Fatal(nextErr)
	}
	if err = cursor.Close(); err != nil {
		t.Fatal(err)
	}
	if err = cursor.Close(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if _, _, found, err := table.Delete(1); !found || err != nil {
		t.Fatal(found, err)
	}
	if err = slot.Set(12); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if completed, eachErr := table.ForEach(nil); completed || !errors.Is(eachErr, ErrDimension) {
		t.Fatal(completed, eachErr)
	}
	if completed, eachErr := table.ForEach(func(int, int) StableTableAction { return StableTableAction(99) }); completed || !errors.Is(eachErr, ErrDimension) {
		t.Fatal(completed, eachErr)
	}
	if completed, eachErr := table.ForEach(func(int, int) StableTableAction { return StableTableStop }); completed || eachErr != nil {
		t.Fatal(completed, eachErr)
	}
	if completed, eachErr := table.ForEach(func(int, int) StableTableAction {
		_, _ = table.Set(3, 30)
		return StableTableContinue
	}); completed || !errors.Is(eachErr, ErrInvariant) {
		t.Fatal(completed, eachErr)
	}
	entries, _ := table.Entries()
	entries[0].Value = 999
	if stableTableValue(t, table, entries[0].Key) == 999 {
		t.Fatal("entry snapshot aliases table")
	}

	duplicates, _ := NewStableTable[int, int](0)
	_ = duplicates.AddDirect(7, 1)
	_ = duplicates.AddDirect(7, 2)
	completed, err := duplicates.ForEach(func(_ int, value int) StableTableAction {
		if value == 1 {
			return StableTableDelete
		}
		return StableTableContinue
	})
	if err != nil || !completed || stableTableValue(t, duplicates, 7) != 2 {
		t.Fatal(completed, err)
	}
	_, _, found, _ := duplicates.Delete(7)
	if _, stillFound, getErr := duplicates.Get(7); !found || stillFound || getErr != nil {
		t.Fatal(found, stillFound, getErr)
	}

	limited, _ := NewStableTable[int, int](0)
	limited.generation = math.MaxUint64
	if err = limited.AddDirect(1, 1); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if count, _ := limited.Len(); count != 0 {
		t.Fatal(count)
	}
	if err = table.Close(); err != nil {
		t.Fatal(err)
	}
	if _, _, err = table.Get(2); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if err = table.Close(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if _, err = (*StableTable[int, int])(nil).Len(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if _, err = StableStringHash("bad\x00value", 7); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err = StablePooledStringHash("x", 0); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err = StableIntegerHash(1, -1); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err = StableAddressHash(1, 0); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
}

type stableTableModelEntry struct {
	key, value int
	live       bool
}

type stableTableModel struct {
	entries []stableTableModelEntry
}

func (model *stableTableModel) find(key int) int {
	for index := len(model.entries) - 1; index >= 0; index-- {
		if model.entries[index].live && model.entries[index].key == key {
			return index
		}
	}
	return -1
}

func (model *stableTableModel) set(key, value int) bool {
	if index := model.find(key); index >= 0 {
		model.entries[index].value = value
		return true
	}
	model.entries = append(model.entries, stableTableModelEntry{key: key, value: value, live: true})
	return false
}

func (model *stableTableModel) add(key, value int) {
	model.entries = append(model.entries, stableTableModelEntry{key: key, value: value, live: true})
}

func (model *stableTableModel) delete(key int) (int, bool) {
	index := model.find(key)
	if index < 0 {
		return 0, false
	}
	model.entries[index].live = false
	return model.entries[index].value, true
}

func (model *stableTableModel) items() []StableTableItem[int, int] {
	result := make([]StableTableItem[int, int], 0)
	for _, entry := range model.entries {
		if entry.live {
			result = append(result, StableTableItem[int, int]{Key: entry.key, Value: entry.value})
		}
	}
	return result
}

func FuzzStableTable(f *testing.F) {
	f.Add([]byte{0, 1, 10, 1, 1, 20, 2, 1, 0, 3, 1, 0})
	f.Add([]byte{1, 7, 1, 1, 7, 2, 2, 7, 0, 4, 0, 0})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 768 {
			data = data[:768]
		}
		modulus := 97
		if len(data) > 0 {
			modulus = int(data[0]%127) + 1
		}
		text := string(data)
		ordinary, ordinaryErr := StableStringHash(text, modulus)
		pooled, pooledErr := StablePooledStringHash(text, modulus)
		if strings.IndexByte(text, 0) >= 0 {
			if !errors.Is(ordinaryErr, ErrDimension) || !errors.Is(pooledErr, ErrDimension) {
				t.Fatal("NUL hash", ordinaryErr, pooledErr)
			}
		} else {
			var ordinaryBits, pooledBits uint32
			for _, value := range data {
				ordinaryBits = ordinaryBits*997 + uint32(value)
				pooledBits = pooledBits*997 + uint32(int32(int8(value)))
			}
			pooledSigned := int64(int32(pooledBits))
			if pooledSigned < 0 {
				pooledSigned = -pooledSigned
			}
			if ordinaryErr != nil || pooledErr != nil || ordinary != int(ordinaryBits%uint32(modulus)) || pooled != int(pooledSigned%int64(modulus)) {
				t.Fatal("hash", ordinary, pooled, ordinaryErr, pooledErr)
			}
		}
		table, err := NewStableTable[int, int](1)
		if err != nil {
			t.Fatal(err)
		}
		model := &stableTableModel{}
		for offset := 0; offset+2 < len(data); offset += 3 {
			op := data[offset] % 6
			key := int(data[offset+1] % 12)
			value := int(data[offset+2])
			switch op {
			case 0:
				replaced, setErr := table.Set(key, value)
				if setErr != nil || replaced != model.set(key, value) {
					t.Fatal("set", offset, replaced, setErr)
				}
			case 1:
				if err = table.AddDirect(key, value); err != nil {
					t.Fatal("add", offset, err)
				}
				model.add(key, value)
			case 2:
				_, got, found, deleteErr := table.Delete(key)
				want, wantFound := model.delete(key)
				if deleteErr != nil || found != wantFound || got != want {
					t.Fatal("delete", offset, got, found, deleteErr, want, wantFound)
				}
			case 3:
				slot, existing, slotErr := table.FindOrAdd(key)
				index := model.find(key)
				if slotErr != nil || existing != (index >= 0) {
					t.Fatal("find/add", offset, existing, slotErr, index)
				}
				if index < 0 {
					model.add(key, 0)
				}
				if data[offset]&0x80 != 0 {
					_ = slot.Set(value)
					model.entries[model.find(key)].value = value
				}
			case 4:
				if err = table.Clear(); err != nil {
					t.Fatal("clear", offset, err)
				}
				model.entries = nil
			case 5:
				clone, cloneErr := table.Clone()
				if cloneErr != nil {
					t.Fatal("clone", offset, cloneErr)
				}
				cloneItems, _ := clone.Entries()
				if !reflect.DeepEqual(cloneItems, model.items()) {
					t.Fatal("clone state", offset, cloneItems, model.items())
				}
			}
			items, itemsErr := table.Entries()
			count, countErr := table.Len()
			wantItems := model.items()
			if itemsErr != nil || countErr != nil || count != len(wantItems) || !reflect.DeepEqual(items, wantItems) {
				t.Fatal("state", offset, count, countErr, items, itemsErr, wantItems)
			}
			for candidate := 0; candidate < 12; candidate++ {
				got, found, getErr := table.Get(candidate)
				index := model.find(candidate)
				if getErr != nil || found != (index >= 0) || index >= 0 && got != model.entries[index].value {
					t.Fatal("lookup", offset, candidate, got, found, getErr, index)
				}
			}
		}
		if err = table.Close(); err != nil {
			t.Fatal(err)
		}
	})
}
