package abc

// Go-only models and fuzz tests for src/misc/vec/vecWec.h.

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
	"sort"
	"testing"
	"unsafe"
)

func TestIntLevelVectorCapacityPushAccessAndMemory(t *testing.T) {
	vector, err := NewIntLevelVectorWithLevels(3, IntLevelVectorLimits{})
	if err != nil {
		t.Fatal(err)
	}
	if capacity, _ := vector.Capacity(); capacity != 8 {
		t.Fatal(capacity)
	}
	if err = vector.Push(4, 40); err != nil {
		t.Fatal(err)
	}
	if err = vector.Push(9, 90); err != nil {
		t.Fatal(err)
	}
	if size, _ := vector.LevelCount(); size != 10 {
		t.Fatal(size)
	}
	if capacity, _ := vector.Capacity(); capacity != 10 {
		t.Fatal(capacity)
	}
	for value := 0; value < 16; value++ {
		if err = vector.Push(4, value); err != nil {
			t.Fatal(err)
		}
	}
	if capacity, _ := vector.LevelCapacity(4); capacity != 32 {
		t.Fatal(capacity)
	}
	if level, _ := vector.LevelID(4); level != 4 {
		t.Fatal(level)
	}
	if entry, entryErr := vector.Entry(9, 0); entryErr != nil || entry != 90 {
		t.Fatal(entry, entryErr)
	}
	last, err := vector.LastLevel()
	if err != nil || !slices.Equal(last, []int{90}) {
		t.Fatal(last, err)
	}
	snapshot, err := vector.Snapshot()
	if err != nil || len(snapshot) != 10 || !slices.Equal(snapshot[4], append([]int{40}, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15)) {
		t.Fatal(snapshot, err)
	}
	snapshot[4][0] = -1
	if entry, _ := vector.Entry(4, 0); entry != 40 {
		t.Fatal("snapshot aliases storage", entry)
	}
	pointerBytes := int(unsafe.Sizeof(uintptr(0)))
	wantMemory := 10*(8+pointerBytes) + (32+16)*4
	if memory, memoryErr := vector.NativeMemoryBytes(); memoryErr != nil || memory != wantMemory {
		t.Fatal(memory, wantMemory, memoryErr)
	}
	if total, _ := vector.TotalEntries(); total != 18 {
		t.Fatal(total)
	}
	if used, _ := vector.UsedLevels(); used != 2 {
		t.Fatal(used)
	}
	if used, _ := vector.UsedLevelsRange(4, 9); used != 1 {
		t.Fatal(used)
	}
	if vector.Validate() != nil {
		t.Fatal(vector.Validate())
	}
}

func TestIntLevelVectorInitShrinkClearDormantStorage(t *testing.T) {
	vector, _ := NewIntLevelVectorWithLevels(3, IntLevelVectorLimits{})
	_ = vector.Push(0, 1)
	_ = vector.Push(1, 2)
	_ = vector.Push(2, 3)
	if err := vector.Init(1); err != nil {
		t.Fatal(err)
	}
	if snapshot, _ := vector.Snapshot(); !reflect.DeepEqual(snapshot, [][]int{{1}}) {
		t.Fatal(snapshot)
	}
	if err := vector.Init(3); err != nil {
		t.Fatal(err)
	}
	if snapshot, _ := vector.Snapshot(); !reflect.DeepEqual(snapshot, [][]int{{1}, {2}, {3}}) {
		t.Fatal("Init did not preserve dormant embedded headers", snapshot)
	}
	if err := vector.Shrink(1); err != nil {
		t.Fatal(err)
	}
	if err := vector.Init(3); err != nil {
		t.Fatal(err)
	}
	if snapshot, _ := vector.Snapshot(); !reflect.DeepEqual(snapshot, [][]int{{1}, nil, nil}) {
		t.Fatal("Shrink did not clear removed levels", snapshot)
	}
	if capacity, _ := vector.LevelCapacity(1); capacity != 16 {
		t.Fatal("Shrink should retain inner allocation", capacity)
	}
	if err := vector.Clear(); err != nil {
		t.Fatal(err)
	}
	if size, _ := vector.LevelCount(); size != 0 {
		t.Fatal(size)
	}
	if err := vector.Init(3); err != nil {
		t.Fatal(err)
	}
	if snapshot, _ := vector.Snapshot(); !reflect.DeepEqual(snapshot, [][]int{nil, nil, nil}) {
		t.Fatal(snapshot)
	}
}

func TestIntLevelVectorPushInsertTraversalAndClone(t *testing.T) {
	vector, _ := NewExactIntLevelVector(0, IntLevelVectorLimits{})
	first, err := vector.PushLevel()
	if err != nil || first != 0 {
		t.Fatal(first, err)
	}
	if capacity, _ := vector.Capacity(); capacity != 16 {
		t.Fatal(capacity)
	}
	_ = vector.Push(0, 1)
	second, _ := vector.PushLevel()
	_ = vector.Push(second, 2)
	third, _ := vector.PushLevel()
	_ = vector.Push(third, 3)
	inserted, err := vector.InsertLevel(1)
	if err != nil || inserted != 1 {
		t.Fatal(inserted, err)
	}
	_ = vector.Push(1, 9)
	if snapshot, _ := vector.Snapshot(); !reflect.DeepEqual(snapshot, [][]int{{1}, {9}, {2}, {3}}) {
		t.Fatal(snapshot)
	}
	levels, err := vector.Levels(1, 4, true)
	wantLevels := []IntLevelRecord{{3, []int{3}}, {2, []int{2}}, {1, []int{9}}}
	if err != nil || !reflect.DeepEqual(levels, wantLevels) {
		t.Fatal(levels, err)
	}
	selected, err := vector.LevelsByIndex([]int{3, 0, 3})
	if err != nil || len(selected) != 3 || !slices.Equal(selected[0].Values, []int{3}) || !slices.Equal(selected[1].Values, []int{1}) {
		t.Fatal(selected, err)
	}
	pairs, err := vector.LevelPairs()
	if err != nil || len(pairs) != 2 || pairs[1].FirstIndex != 2 || !slices.Equal(pairs[1].Second, []int{3}) {
		t.Fatal(pairs, err)
	}
	other, _ := NewIntLevelVectorWithLevels(4, IntLevelVectorLimits{})
	_ = other.Push(2, 20)
	zipped, err := ZipIntLevelVectors(vector, other)
	if err != nil || len(zipped) != 4 || !slices.Equal(zipped[2].First, []int{2}) || !slices.Equal(zipped[2].Second, []int{20}) {
		t.Fatal(zipped, err)
	}
	clone, err := vector.Clone()
	if err != nil {
		t.Fatal(err)
	}
	_ = clone.Push(0, 100)
	if size, _ := vector.LevelSize(0); size != 1 {
		t.Fatal("clone aliases source", size)
	}
	odd, _ := NewIntLevelVectorWithLevels(1, IntLevelVectorLimits{})
	if _, err = odd.LevelPairs(); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
}

func TestIntLevelVectorUniqueSortFormatAndClasses(t *testing.T) {
	vector, _ := NewIntLevelVectorWithLevels(5, IntLevelVectorLimits{})
	for level, values := range [][]int{{8, 1}, {4}, {7, 2, 3}, {4, 5}, {6, 0}} {
		for _, value := range values {
			_ = vector.Push(level, value)
		}
	}
	if added, err := vector.PushUnique(3, 5); err != nil || added {
		t.Fatal(added, err)
	}
	if added, err := vector.PushUnique(3, 6); err != nil || !added {
		t.Fatal(added, err)
	}
	if err := vector.SortBySize(false); err != nil {
		t.Fatal(err)
	}
	if err := vector.SortByFirst(false); err != nil {
		t.Fatal(err)
	}
	if snapshot, _ := vector.Snapshot(); !reflect.DeepEqual(snapshot, [][]int{{4}, {4, 5, 6}, {6, 0}, {7, 2, 3}, {8, 1}}) {
		t.Fatal(snapshot)
	}
	if err := vector.SortByLast(true); err != nil {
		t.Fatal(err)
	}
	if entry, _ := vector.Entry(0, 0); entry != 4 {
		t.Fatal(entry)
	}
	text, err := vector.Format(true)
	wantText := "    0 : { 4 5 6 }\n    2 : { 7 2 3 }\n    3 : { 8 1 }\n    4 : { 6 0 }\n"
	if err != nil || text != wantText {
		t.Fatalf("%q %v", text, err)
	}
	literals, _ := NewIntLevelVectorWithLevels(2, IntLevelVectorLimits{})
	_ = literals.Push(0, 4)
	_ = literals.Push(0, 7)
	literalText, err := literals.FormatLiterals()
	wantLiterals := "    0 :  2  { +2 -3 }\n    1 :  0  { }\n"
	if err != nil || literalText != wantLiterals {
		t.Fatalf("%q %v", literalText, err)
	}
	classes, err := NewIntEquivalenceClasses([]int{2, 0, 2, 1, 2, 0}, IntLevelVectorLimits{})
	if err != nil {
		t.Fatal(err)
	}
	if snapshot, _ := classes.Snapshot(); !reflect.DeepEqual(snapshot, [][]int{{1, 5}, {3}, {0, 2, 4}}) {
		t.Fatal(snapshot)
	}
	if count, entries, countErr := classes.NonTrivialCounts(); countErr != nil || count != 2 || entries != 5 {
		t.Fatal(count, entries, countErr)
	}
	if firsts, firstErr := classes.FirstEntries(); firstErr != nil || !slices.Equal(firsts, []int{1, 3, 0}) {
		t.Fatal(firsts, firstErr)
	}
	emptyClasses, err := NewIntEquivalenceClasses(nil, IntLevelVectorLimits{})
	if err != nil {
		t.Fatal(err)
	}
	if size, _ := emptyClasses.LevelCount(); size != 1 {
		t.Fatal(size)
	}
}

func TestIntLevelVectorMarksAndRemoveEmpty(t *testing.T) {
	vector, _ := NewIntLevelVectorWithLevels(5, IntLevelVectorLimits{})
	_ = vector.Push(0, 10)
	_ = vector.Push(2, 20)
	_ = vector.Push(4, 40)
	if err := vector.MarkLevels([]int{0, 4}); err != nil {
		t.Fatal(err)
	}
	if marked, _ := vector.LevelMarked(4); !marked {
		t.Fatal(marked)
	}
	if err := vector.MarkLevels([]int{2, 4}); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if marked, _ := vector.LevelMarked(2); marked {
		t.Fatal("failed mark was not atomic")
	}
	if err := vector.UnmarkLevels([]int{0, 4}); err != nil {
		t.Fatal(err)
	}
	if err := vector.SetLevelMark(2); err != nil {
		t.Fatal(err)
	}
	if err := vector.ToggleLevelMark(2); err != nil {
		t.Fatal(err)
	}
	if err := vector.RemoveEmpty(); err != nil {
		t.Fatal(err)
	}
	if snapshot, _ := vector.Snapshot(); !reflect.DeepEqual(snapshot, [][]int{{10}, {20}, {40}}) {
		t.Fatal(snapshot)
	}
	if size, _ := vector.LevelCount(); size != 3 {
		t.Fatal(size)
	}
}

func TestIntLevelVectorErrorsLimitsCancellationAndClose(t *testing.T) {
	if vector, err := NewIntLevelVector(-1, IntLevelVectorLimits{}); vector != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(vector, err)
	}
	if vector, err := NewIntLevelVector(9, IntLevelVectorLimits{MaxLevels: 8}); vector != nil || !errors.Is(err, ErrLimit) {
		t.Fatal(vector, err)
	}
	limited, _ := NewExactIntLevelVector(0, IntLevelVectorLimits{MaxLevels: 2, MaxEntries: 2, MaxStorageEntries: 16})
	if err := limited.Push(1, 1); err != nil {
		t.Fatal(err)
	}
	if err := limited.Push(1, 2); err != nil {
		t.Fatal(err)
	}
	if err := limited.Push(1, 3); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if err := limited.Push(2, 1); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, err := limited.Entry(0, 0); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	short, _ := NewIntLevelVectorWithLevels(1, IntLevelVectorLimits{})
	long, _ := NewIntLevelVectorWithLevels(2, IntLevelVectorLimits{})
	if _, err := ZipIntLevelVectors(long, short); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	cancel := false
	cancelled, _ := NewExactIntLevelVector(0, IntLevelVectorLimits{Stop: func() bool { return cancel }})
	cancel = true
	if err := cancelled.Push(4, 1); !errors.Is(err, ErrCancelled) {
		t.Fatal(err)
	}
	if capacity, _ := cancelled.Capacity(); capacity != 0 {
		t.Fatal("cancelled push allocated", capacity)
	}
	if err := limited.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := limited.LevelCount(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
}

type intLevelModel struct {
	slots [][]int
	size  int
}

func (model *intLevelModel) grow(capacity int) {
	if capacity <= len(model.slots) {
		return
	}
	slots := make([][]int, capacity)
	copy(slots, model.slots)
	model.slots = slots
}

func (model *intLevelModel) push(level, value int) {
	if level >= len(model.slots) {
		capacity := level + 1
		if 2*model.size > capacity {
			capacity = 2 * model.size
		}
		model.grow(capacity)
	}
	if model.size <= level {
		model.size = level + 1
	}
	model.slots[level] = append(model.slots[level], value)
}

func (model *intLevelModel) pushLevel() {
	if model.size == len(model.slots) {
		capacity := 16
		if len(model.slots) >= 16 {
			capacity = 2 * len(model.slots)
		}
		model.grow(capacity)
	}
	model.size++
}

func (model *intLevelModel) insert(index int) {
	if model.size == len(model.slots) {
		capacity := 16
		if len(model.slots) >= 16 {
			capacity = 2 * len(model.slots)
		}
		model.grow(capacity)
	}
	copy(model.slots[index+1:model.size+1], model.slots[index:model.size])
	model.slots[index] = nil
	model.size++
}

func intLevelVectorModel(data []byte) error {
	if len(data) > 3000 {
		data = data[:3000]
	}
	vector, err := NewExactIntLevelVector(0, IntLevelVectorLimits{MaxLevels: 32, MaxEntries: 4096, MaxStorageEntries: 8192})
	if err != nil {
		return err
	}
	model := intLevelModel{}
	for offset := 0; offset+2 < len(data); offset += 3 {
		op := data[offset] % 9
		level := int(data[offset+1] % 16)
		value := int(int8(data[offset+2]))
		switch op {
		case 0:
			if err = vector.Push(level, value); err != nil {
				return err
			}
			model.push(level, value)
		case 1:
			added, pushErr := vector.PushUnique(level, value)
			if pushErr != nil {
				return pushErr
			}
			if level >= len(model.slots) || level >= model.size || !slices.Contains(model.slots[level], value) {
				if !added {
					return errors.New("unique insert was skipped")
				}
				model.push(level, value)
			} else if added {
				return errors.New("duplicate was inserted")
			}
		case 2:
			size := level
			if err = vector.Init(size); err != nil {
				return err
			}
			model.grow(size)
			model.size = size
		case 3:
			if model.size != 0 {
				size := level % (model.size + 1)
				if err = vector.Shrink(size); err != nil {
					return err
				}
				for index := size; index < model.size; index++ {
					model.slots[index] = nil
				}
				model.size = size
			}
		case 4:
			if err = vector.Clear(); err != nil {
				return err
			}
			for index := 0; index < model.size; index++ {
				model.slots[index] = nil
			}
			model.size = 0
		case 5:
			if model.size < 16 {
				if _, err = vector.PushLevel(); err != nil {
					return err
				}
				model.pushLevel()
			}
		case 6:
			if model.size < 16 {
				index := 0
				if model.size != 0 {
					index = level % (model.size + 1)
				}
				if _, err = vector.InsertLevel(index); err != nil {
					return err
				}
				model.insert(index)
			}
		case 7:
			if err = vector.RemoveEmpty(); err != nil {
				return err
			}
			write := 0
			for read := 0; read < model.size; read++ {
				if len(model.slots[read]) != 0 {
					model.slots[write] = model.slots[read]
					write++
				}
			}
			for index := write; index < model.size; index++ {
				model.slots[index] = nil
			}
			model.size = write
		case 8:
			reverse := value&1 != 0
			if err = vector.SortBySize(reverse); err != nil {
				return err
			}
			sort.SliceStable(model.slots[:model.size], func(first, second int) bool {
				if reverse {
					return len(model.slots[first]) > len(model.slots[second])
				}
				return len(model.slots[first]) < len(model.slots[second])
			})
		}
		snapshot, snapshotErr := vector.Snapshot()
		if snapshotErr != nil || len(snapshot) != model.size {
			return fmt.Errorf("snapshot size %d/%d: %w", len(snapshot), model.size, snapshotErr)
		}
		for index := 0; index < model.size; index++ {
			if !slices.Equal(snapshot[index], model.slots[index]) {
				return fmt.Errorf("level %d got %v want %v", index, snapshot[index], model.slots[index])
			}
		}
		if err = vector.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func TestIntLevelVectorModelCampaign(t *testing.T) {
	state := uint64(0xd1b54a32d192ed03)
	for run := 0; run < 5000; run++ {
		data := make([]byte, 96)
		for index := range data {
			state ^= state << 7
			state ^= state >> 9
			state ^= state << 8
			data[index] = byte(state)
		}
		if err := intLevelVectorModel(data); err != nil {
			t.Fatal(run, err)
		}
	}
}

func FuzzIntLevelVectorModel(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte{0, 0, 1, 0, 7, 2, 2, 3, 0, 5, 0, 0, 6, 0, 0})
	f.Add([]byte{0, 4, 1, 2, 1, 0, 7, 0, 0, 8, 0, 1})
	f.Fuzz(func(t *testing.T, data []byte) {
		if err := intLevelVectorModel(data); err != nil {
			t.Fatal(err)
		}
	})
}
