package abc

// Go-only models and fuzz tests for src/misc/vec/vecVec.h.

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
	"sort"
	"testing"
	"unsafe"
)

func TestNestedVectorCapacityGrowthEntriesCloneAndMemory(t *testing.T) {
	vector, err := NewNestedVector[int](1, NestedVectorLimits{})
	if err != nil {
		t.Fatal(err)
	}
	if size, _ := vector.LevelCount(); size != 0 {
		t.Fatal(size)
	}
	if capacity, _ := vector.Capacity(); capacity != 8 {
		t.Fatal(capacity)
	}
	if err = vector.Expand(10); err != nil {
		t.Fatal(err)
	}
	if size, _ := vector.LevelCount(); size != 11 {
		t.Fatal(size)
	}
	if capacity, _ := vector.Capacity(); capacity != 11 {
		t.Fatal(capacity)
	}
	for value := 0; value < 17; value++ {
		if err = vector.Push(2, value); err != nil {
			t.Fatal(err)
		}
	}
	if capacity, _ := vector.LevelCapacity(2); capacity != 32 {
		t.Fatal(capacity)
	}
	if total, _ := vector.TotalEntries(); total != 17 {
		t.Fatal(total)
	}
	if entry, entryErr := vector.Entry(2, 16); entryErr != nil || entry != 16 {
		t.Fatal(entry, entryErr)
	}
	level, err := vector.Level(2)
	if err != nil || !slices.Equal(level, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}) {
		t.Fatal(level, err)
	}
	level[0] = 99
	if entry, _ := vector.Entry(2, 0); entry != 0 {
		t.Fatal("level snapshot aliases storage", entry)
	}
	pointerBytes := int(unsafe.Sizeof(uintptr(0)))
	wantMemory := (8 + pointerBytes) + 11*pointerBytes + 11*(8+pointerBytes) + 32*4
	if memory, memoryErr := vector.NativeMemoryBytes(4); memoryErr != nil || memory != wantMemory {
		t.Fatal(memory, wantMemory, memoryErr)
	}
	clone, err := vector.Clone()
	if err != nil {
		t.Fatal(err)
	}
	if err = clone.Push(2, 100); err != nil {
		t.Fatal(err)
	}
	if sourceSize, _ := vector.LevelSize(2); sourceSize != 17 {
		t.Fatal("clone aliases source", sourceSize)
	}
	if err = vector.Clear(); err != nil {
		t.Fatal(err)
	}
	if size, _ := vector.LevelCount(); size != 11 {
		t.Fatal(size)
	}
	if total, _ := vector.TotalEntries(); total != 0 {
		t.Fatal(total)
	}
	if capacity, _ := vector.LevelCapacity(2); capacity != 32 {
		t.Fatal(capacity)
	}
	if vector.Validate() != nil || clone.Validate() != nil {
		t.Fatal(vector.Validate(), clone.Validate())
	}
}

func TestNestedVectorUniqueTraversalZipAndPointerIdentity(t *testing.T) {
	vector, err := NewNestedVectorWithLevels[int](4, NestedVectorLimits{})
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range []struct {
		level int
		value int
	}{{0, 1}, {0, 2}, {2, 3}, {2, 4}, {3, 5}} {
		if err = vector.Push(row.level, row.value); err != nil {
			t.Fatal(err)
		}
	}
	if added, uniqueErr := vector.PushUnique(2, 3); uniqueErr != nil || added {
		t.Fatal(added, uniqueErr)
	}
	if added, uniqueErr := vector.PushUnique(1, 7); uniqueErr != nil || !added {
		t.Fatal(added, uniqueErr)
	}
	levels, err := vector.Levels(1, 4, true)
	wantLevels := []NestedVectorLevel[int]{{3, []int{5}}, {2, []int{3, 4}}, {1, []int{7}}}
	if err != nil || !reflect.DeepEqual(levels, wantLevels) {
		t.Fatal(levels, err)
	}
	entries, err := vector.Entries(0, 4, true, true)
	wantEntries := []NestedVectorEntry[int]{{3, 0, 5}, {2, 1, 4}, {2, 0, 3}, {1, 0, 7}, {0, 1, 2}, {0, 0, 1}}
	if err != nil || !reflect.DeepEqual(entries, wantEntries) {
		t.Fatal(entries, err)
	}
	other, _ := NewNestedVectorWithLevels[string](4, NestedVectorLimits{})
	_ = other.Push(0, "a")
	_ = other.Push(2, "b")
	pairs, err := ZipNestedVectorLevels(vector, other)
	if err != nil || len(pairs) != 4 || !slices.Equal(pairs[2].First, []int{3, 4}) || !slices.Equal(pairs[2].Second, []string{"b"}) {
		t.Fatal(pairs, err)
	}

	type item struct{ value int }
	first, second := &item{1}, &item{1}
	pointers, _ := NewNestedVector[*item](0, NestedVectorLimits{})
	if added, pushErr := pointers.PushUnique(0, first); pushErr != nil || !added {
		t.Fatal(added, pushErr)
	}
	if added, pushErr := pointers.PushUnique(0, first); pushErr != nil || added {
		t.Fatal(added, pushErr)
	}
	if added, pushErr := pointers.PushUnique(0, second); pushErr != nil || !added {
		t.Fatal(added, pushErr)
	}
	if total, _ := pointers.TotalEntries(); total != 2 {
		t.Fatal(total)
	}
}

func TestNestedVectorSortingAndFormatting(t *testing.T) {
	vector, _ := NewNestedVectorWithLevels[int](5, NestedVectorLimits{})
	for level, values := range [][]int{{8, 9}, {4}, {7, 1, 2}, {4, 5}, {6, 7}} {
		for _, value := range values {
			if err := vector.Push(level, value); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := vector.SortLevelsBySize(false); err != nil {
		t.Fatal(err)
	}
	levels, _ := vector.Levels(0, 5, false)
	got := make([][]int, len(levels))
	for index := range levels {
		got[index] = levels[index].Values
	}
	if !reflect.DeepEqual(got, [][]int{{4}, {8, 9}, {4, 5}, {6, 7}, {7, 1, 2}}) {
		t.Fatal(got)
	}
	if err := vector.SortLevelsBySize(true); err != nil {
		t.Fatal(err)
	}
	if err := SortNestedIntsByFirst(vector, false); err != nil {
		t.Fatal(err)
	}
	levels, _ = vector.Levels(0, 5, false)
	got = got[:0]
	for _, level := range levels {
		got = append(got, level.Values)
	}
	if !reflect.DeepEqual(got, [][]int{{4, 5}, {4}, {6, 7}, {7, 1, 2}, {8, 9}}) {
		t.Fatal(got)
	}
	formatted, err := FormatNestedInts(vector, true)
	want := "    0 : { 4 5 }\n    2 : { 6 7 }\n    3 : { 7 1 2 }\n    4 : { 8 9 }\n"
	if err != nil || formatted != want {
		t.Fatalf("%q %v", formatted, err)
	}
	if err = SortNestedIntsByFirst(vector, true); err != nil {
		t.Fatal(err)
	}
	first, _ := vector.Entry(0, 0)
	if first != 8 {
		t.Fatal(first)
	}
	empty, _ := NewNestedVectorWithLevels[int](1, NestedVectorLimits{})
	if err = SortNestedIntsByFirst(empty, false); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
}

func TestNestedVectorErrorsLimitsCancellationAndClose(t *testing.T) {
	if vector, err := NewNestedVector[int](-1, NestedVectorLimits{}); vector != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(vector, err)
	}
	if vector, err := NewNestedVector[int](9, NestedVectorLimits{MaxLevels: 8}); vector != nil || !errors.Is(err, ErrLimit) {
		t.Fatal(vector, err)
	}
	if vector, err := NewNestedVector[int](0, NestedVectorLimits{MaxEntries: -1}); vector != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(vector, err)
	}
	limited, _ := NewNestedVector[int](0, NestedVectorLimits{MaxLevels: 2, MaxEntries: 20, MaxStorageEntries: 16})
	for value := 0; value < 16; value++ {
		if err := limited.Push(1, value); err != nil {
			t.Fatal(value, err)
		}
	}
	if err := limited.Push(1, 16); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if err := limited.Push(2, 1); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	atomic, _ := NewNestedVector[int](0, NestedVectorLimits{MaxLevels: 8, MaxEntries: 8, MaxStorageEntries: 1})
	if err := atomic.Push(5, 1); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if levels, _ := atomic.LevelCount(); levels != 0 {
		t.Fatal("failed push expanded the vector", levels)
	}
	if _, err := limited.Entry(-1, 0); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err := limited.Levels(1, 0, false); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	short, _ := NewNestedVectorWithLevels[int](1, NestedVectorLimits{})
	long, _ := NewNestedVectorWithLevels[int](2, NestedVectorLimits{})
	if _, err := ZipNestedVectorLevels(long, short); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	cancel := false
	cancelled, _ := NewNestedVector[int](0, NestedVectorLimits{Stop: func() bool { return cancel }})
	cancel = true
	if err := cancelled.Push(0, 1); !errors.Is(err, ErrCancelled) {
		t.Fatal(err)
	}
	if size, _ := cancelled.LevelCount(); size != 0 {
		t.Fatal("cancelled mutation was not atomic", size)
	}
	if err := limited.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := limited.LevelCount(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
}

func nestedVectorModel(data []byte) error {
	if len(data) > 3000 {
		data = data[:3000]
	}
	vector, err := NewNestedVectorWithLevels[int](1, NestedVectorLimits{MaxLevels: 8, MaxEntries: 4096, MaxStorageEntries: 8192})
	if err != nil {
		return err
	}
	model := make([][]int, 1)
	for offset := 0; offset+2 < len(data); offset += 3 {
		op := data[offset] % 5
		level := int(data[offset+1] % 8)
		value := int(int8(data[offset+2]))
		switch op {
		case 0:
			if err = vector.Push(level, value); err != nil {
				return err
			}
			for len(model) <= level {
				model = append(model, nil)
			}
			model[level] = append(model[level], value)
		case 1:
			added, pushErr := vector.PushUnique(level, value)
			if pushErr != nil {
				return pushErr
			}
			for len(model) <= level {
				model = append(model, nil)
			}
			wantAdded := !slices.Contains(model[level], value)
			if added != wantAdded {
				return fmt.Errorf("unique result %v want %v", added, wantAdded)
			}
			if wantAdded {
				model[level] = append(model[level], value)
			}
		case 2:
			if err = vector.Clear(); err != nil {
				return err
			}
			for index := range model {
				model[index] = nil
			}
		case 3:
			if err = vector.Expand(level); err != nil {
				return err
			}
			for len(model) <= level {
				model = append(model, nil)
			}
		case 4:
			reverse := value&1 != 0
			if err = vector.SortLevelsBySize(reverse); err != nil {
				return err
			}
			sort.SliceStable(model, func(first, second int) bool {
				if reverse {
					return len(model[first]) > len(model[second])
				}
				return len(model[first]) < len(model[second])
			})
		}
		levels, snapshotErr := vector.Levels(0, len(model), false)
		if snapshotErr != nil || len(levels) != len(model) {
			return fmt.Errorf("snapshot: %w", snapshotErr)
		}
		total := 0
		for index := range model {
			if !slices.Equal(levels[index].Values, model[index]) {
				return fmt.Errorf("level %d got %v want %v", index, levels[index].Values, model[index])
			}
			total += len(model[index])
		}
		if got, totalErr := vector.TotalEntries(); totalErr != nil || got != total {
			return fmt.Errorf("total %d want %d: %w", got, total, totalErr)
		}
		if err = vector.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func TestNestedVectorModelCampaign(t *testing.T) {
	state := uint64(0x9e3779b97f4a7c15)
	for run := 0; run < 5000; run++ {
		data := make([]byte, 96)
		for index := range data {
			state ^= state << 7
			state ^= state >> 9
			state ^= state << 8
			data[index] = byte(state)
		}
		if err := nestedVectorModel(data); err != nil {
			t.Fatal(run, err)
		}
	}
}

func FuzzNestedVectorModel(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte{0, 0, 1, 1, 0, 1, 0, 7, 255, 4, 0, 1})
	f.Add([]byte{3, 7, 0, 0, 7, 2, 1, 7, 2, 2, 0, 0})
	f.Fuzz(func(t *testing.T, data []byte) {
		if err := nestedVectorModel(data); err != nil {
			t.Fatal(err)
		}
	})
}
