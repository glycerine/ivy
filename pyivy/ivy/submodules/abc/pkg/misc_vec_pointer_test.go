package abc

// Go-only models and fuzz tests for src/misc/vec/vecPtr.h.

import (
	"errors"
	"fmt"
	"slices"
	"testing"
	"unsafe"
)

func pointerValues(values ...int) []*int {
	result := make([]*int, len(values))
	for index := range values {
		result[index] = new(int)
		*result[index] = values[index]
	}
	return result
}

func TestPointerVectorConstructionCapacityAccessAndMemory(t *testing.T) {
	var absent *PointerVector[int]
	if memory, memoryErr := absent.MemoryBytes(); memoryErr != nil || memory != 0 {
		t.Fatal(memory, memoryErr)
	}
	vector, err := NewPointerVectorWithSize[int](3, PointerVectorLimits{})
	if err != nil {
		t.Fatal(err)
	}
	if capacity, _ := vector.Capacity(); capacity != 8 {
		t.Fatal(capacity)
	}
	if count, _ := vector.CountNil(); count != 3 {
		t.Fatal(count)
	}
	pointerBytes := int(unsafe.Sizeof(uintptr(0)))
	if memory, memoryErr := vector.MemoryBytes(); memoryErr != nil || memory != 8+pointerBytes+8*pointerBytes {
		t.Fatal(memory, memoryErr)
	}
	values := pointerValues(1, 2, 3)
	for index, value := range values {
		if err = vector.Write(index, value); err != nil {
			t.Fatal(err)
		}
	}
	if last, lastErr := vector.Last(); lastErr != nil || last != values[2] {
		t.Fatal(last, lastErr)
	}
	snapshot, err := vector.Snapshot()
	if err != nil || !slices.Equal(snapshot, values) {
		t.Fatal(snapshot, err)
	}
	snapshot[0] = nil
	if first, _ := vector.Entry(0); first != values[0] {
		t.Fatal("snapshot aliases slots")
	}
	input := append([]*int(nil), values...)
	from, err := NewPointerVectorFrom(input, PointerVectorLimits{})
	if err != nil {
		t.Fatal(err)
	}
	input[0] = nil
	if first, _ := from.Entry(0); first != values[0] {
		t.Fatal("constructor borrowed caller slice")
	}
	exact, _ := NewExactPointerVector[int](1, PointerVectorLimits{})
	if capacity, _ := exact.Capacity(); capacity != 1 {
		t.Fatal(capacity)
	}
}

func TestPointerVectorFillSparseCloneMoveAndRelease(t *testing.T) {
	values := pointerValues(1, 2, 3)
	vector, _ := NewExactPointerVector[int](0, PointerVectorLimits{})
	if err := vector.FillAlternating(5, values[0], values[1]); err != nil {
		t.Fatal(err)
	}
	if snapshot, _ := vector.Snapshot(); !slices.Equal(snapshot, []*int{values[0], values[1], values[0], values[1], values[0]}) {
		t.Fatal(snapshot)
	}
	if err := vector.Fill(2, values[2]); err != nil {
		t.Fatal(err)
	}
	if err := vector.FillExtra(6, values[1]); err != nil {
		t.Fatal(err)
	}
	if err := vector.Set(8, values[0]); err != nil {
		t.Fatal(err)
	}
	if value, err := vector.Get(7); err != nil || value != nil {
		t.Fatal(value, err)
	}
	clone, err := vector.Clone()
	if err != nil || !clone.Equal(vector) {
		t.Fatal(err)
	}
	if err = clone.Write(0, values[0]); err != nil {
		t.Fatal(err)
	}
	if first, _ := vector.Entry(0); first != values[2] {
		t.Fatal("clone aliases slot array")
	}
	moved, err := clone.Move()
	if err != nil {
		t.Fatal(err)
	}
	if size, _ := clone.Size(); size != 0 {
		t.Fatal(size)
	}
	if size, _ := moved.Size(); size != 9 {
		t.Fatal(size)
	}
	released, err := moved.Release()
	if err != nil || len(released) != 9 {
		t.Fatal(len(released), err)
	}
	released[0] = nil
	if err = moved.Push(values[0]); err != nil {
		t.Fatal("released vector should remain reusable", err)
	}
}

func TestPointerVectorMutationTraversalAndIdentity(t *testing.T) {
	values := pointerValues(0, 1, 2, 3, 4)
	vector, _ := NewPointerVectorFrom([]*int{values[0], values[1], values[2]}, PointerVectorLimits{})
	if err := vector.PushTwo(values[3], values[4]); err != nil {
		t.Fatal(err)
	}
	if err := vector.PushFirst(values[4]); err != nil {
		t.Fatal(err)
	}
	if duplicate, err := vector.PushUnique(values[2]); err != nil || !duplicate {
		t.Fatal(duplicate, err)
	}
	if duplicate, err := vector.PushUnique(values[3]); err != nil || !duplicate {
		t.Fatal(duplicate, err)
	}
	if err := vector.Insert(2, values[4]); err != nil {
		t.Fatal(err)
	}
	if err := vector.Remove(values[4]); err != nil {
		t.Fatal(err)
	}
	// Remove chooses the last matching pointer.
	if snapshot, _ := vector.Snapshot(); !slices.Equal(snapshot, []*int{values[4], values[0], values[4], values[1], values[2], values[3]}) {
		t.Fatal(snapshot)
	}
	if err := vector.Drop(1); err != nil {
		t.Fatal(err)
	}
	if err := vector.Reorder(2); err != nil {
		t.Fatal(err)
	}
	if err := vector.Reverse(); err != nil {
		t.Fatal(err)
	}
	if snapshot, _ := vector.Snapshot(); !slices.Equal(snapshot, []*int{values[4], values[4], values[3], values[2], values[1]}) {
		t.Fatal(snapshot)
	}
	if popped, err := vector.Pop(); err != nil || popped != values[1] {
		t.Fatal(popped, err)
	}
	entries, err := vector.Entries(1, 4, true)
	if err != nil || len(entries) != 3 || entries[0].Index != 3 || entries[0].Value != values[2] {
		t.Fatal(entries, err)
	}
	other, _ := NewPointerVectorFrom([]*string{new(string), new(string), new(string), new(string)}, PointerVectorLimits{})
	zipped, err := ZipPointerVectors(vector, other)
	if err != nil || len(zipped) != 4 || zipped[2].First != values[3] {
		t.Fatal(zipped, err)
	}
	pairs, err := PointerVectorPairs[int, int](vector, func(value *int) *int { return value })
	if err != nil || len(pairs) != 2 || pairs[1].Second != values[2] {
		t.Fatal(pairs, err)
	}
}

func TestPointerVectorSortUniqifyFreeAndCopy(t *testing.T) {
	values := pointerValues(3, 1, 2, 1)
	vector, _ := NewPointerVectorFrom([]*int{values[0], values[1], values[2], values[3], values[1]}, PointerVectorLimits{})
	compare := func(first, second *int) int { return *first - *second }
	freed := make([]*int, 0)
	counts, err := vector.UniqifyBy(compare, func(value *int) { freed = append(freed, value) })
	if err != nil || !slices.Equal(counts, []int{3, 1, 1}) || len(freed) != 2 {
		t.Fatal(counts, len(freed), err)
	}
	snapshot, _ := vector.Snapshot()
	if len(snapshot) != 3 || *snapshot[0] != 1 || *snapshot[1] != 2 || *snapshot[2] != 3 {
		t.Fatal(snapshot)
	}
	identity, _ := NewPointerVectorFrom([]*int{values[2], values[0], values[2], values[1], values[0]}, PointerVectorLimits{})
	if err = identity.Uniqify(nil); err != nil {
		t.Fatal(err)
	}
	if snapshot, _ = identity.Snapshot(); !slices.Equal(snapshot, []*int{values[2], values[0], values[1]}) {
		t.Fatal(snapshot)
	}
	destination, _ := NewPointerVector[int](0, PointerVectorLimits{})
	if err = destination.CopyFrom(vector); err != nil || !destination.Equal(vector) {
		t.Fatal(err)
	}
	if err = destination.Append(identity); err != nil {
		t.Fatal(err)
	}
	freed = freed[:0]
	if err = destination.FreeValues(func(value *int) { freed = append(freed, value) }); err != nil || len(freed) != 6 {
		t.Fatal(len(freed), err)
	}
}

func TestPointerStringHelpers(t *testing.T) {
	alpha, beta := "alpha", "beta"
	vector, _ := NewPointerVectorFrom([]*string{&alpha, nil, &beta}, PointerVectorLimits{})
	clone, err := ClonePointerStrings(vector)
	if err != nil {
		t.Fatal(err)
	}
	cloned, _ := clone.Entry(0)
	if cloned == &alpha || *cloned != alpha {
		t.Fatal("string pointee was not duplicated")
	}
	if index, err := FindPointerString(vector, "beta"); err != nil || index != 2 {
		t.Fatal(index, err)
	}
	text, err := FormatPointerNames(vector)
	if err != nil || text != "Vector has 3 entries: {alpha (null) beta  }\n" {
		t.Fatalf("%q %v", text, err)
	}
}

func TestPointerSimulationInfoAndTruthTables(t *testing.T) {
	info, err := NewPointerSimulationInfo(3, 4, PointerSimulationLimits{})
	if err != nil {
		t.Fatal(err)
	}
	if err = info.Write(0, []uint32{1, 2, 3, 4}); err != nil {
		t.Fatal(err)
	}
	if err = info.FillFrom(2); err != nil {
		t.Fatal(err)
	}
	if row, _ := info.Row(0); !slices.Equal(row, []uint32{1, 2, ^uint32(0), ^uint32(0)}) {
		t.Fatal(row)
	}
	if err = info.CleanFrom(1); err != nil {
		t.Fatal(err)
	}
	if row, _ := info.Row(0); !slices.Equal(row, []uint32{1, 0, 0, 0}) {
		t.Fatal(row)
	}
	if err = info.DoubleEntries(); err != nil {
		t.Fatal(err)
	}
	if entries, _ := info.EntryCount(); entries != 6 {
		t.Fatal(entries)
	}
	if err = info.DoubleWords(); err != nil {
		t.Fatal(err)
	}
	if words, _ := info.WordCount(); words != 8 {
		t.Fatal(words)
	}
	if row, _ := info.Row(0); !slices.Equal(row, []uint32{1, 0, 0, 0, 0, 0, 0, 0}) {
		t.Fatal(row)
	}
	truth, err := NewPointerTruthTables(7, PointerSimulationLimits{})
	if err != nil {
		t.Fatal(err)
	}
	for variable := 0; variable < 7; variable++ {
		row, _ := truth.Row(variable)
		for word, got := range row {
			for bit := 0; bit < 32; bit++ {
				assignment := word*32 + bit
				want := uint32(0)
				if assignment&(1<<variable) != 0 {
					want = 1
				}
				if got>>bit&1 != want {
					t.Fatal(variable, word, bit, got)
				}
			}
		}
	}
}

func TestPointerVectorErrorsLimitsCancellationAndLifecycle(t *testing.T) {
	if vector, err := NewPointerVector[int](-1, PointerVectorLimits{}); vector != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(vector, err)
	}
	if vector, err := NewPointerVector[int](9, PointerVectorLimits{MaxEntries: 8}); vector != nil || !errors.Is(err, ErrLimit) {
		t.Fatal(vector, err)
	}
	limited, _ := NewExactPointerVector[int](0, PointerVectorLimits{MaxEntries: 16})
	value := 1
	for index := 0; index < 16; index++ {
		if err := limited.Push(&value); err != nil {
			t.Fatal(index, err)
		}
	}
	if err := limited.Push(&value); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	empty, _ := NewExactPointerVector[int](0, PointerVectorLimits{})
	for _, operation := range []func() error{
		func() error { _, err := empty.Entry(-1); return err },
		func() error { return empty.Drop(2) },
		func() error { return empty.Insert(0, &value) },
		func() error { return empty.Reorder(0) },
		func() error { return empty.Remove(new(int)) },
	} {
		if err := operation(); !errors.Is(err, ErrDimension) {
			t.Fatal(err)
		}
	}
	cancel := false
	cancelled, _ := NewExactPointerVector[int](0, PointerVectorLimits{Stop: func() bool { return cancel }})
	cancel = true
	if err := cancelled.Push(&value); !errors.Is(err, ErrCancelled) {
		t.Fatal(err)
	}
	if capacity, _ := cancelled.Capacity(); capacity != 0 {
		t.Fatal("cancelled push allocated", capacity)
	}
	if _, err := NewPointerSimulationInfo(2, 3, PointerSimulationLimits{MaxCells: 5}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	simulationCancelled := false
	info, err := NewPointerSimulationInfo(2, 2, PointerSimulationLimits{Stop: func() bool { return simulationCancelled }})
	if err != nil {
		t.Fatal(err)
	}
	if err = info.Write(0, []uint32{7, 9}); err != nil {
		t.Fatal(err)
	}
	simulationCancelled = true
	if err = info.FillFrom(0); !errors.Is(err, ErrCancelled) {
		t.Fatal(err)
	}
	if row, rowErr := info.Row(0); rowErr != nil || !slices.Equal(row, []uint32{7, 9}) {
		t.Fatal("cancelled simulation fill mutated row", row, rowErr)
	}
	if _, err = NewPointerTruthTables(-1, PointerSimulationLimits{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err = NewPointerTruthTables(31, PointerSimulationLimits{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if err = info.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err = info.Row(0); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if err := limited.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := limited.Size(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
}

func TestPointerVectorCheckedIteratorsAndAtomicFailures(t *testing.T) {
	values := pointerValues(0, 1, 2, 3)
	vector, err := NewPointerVectorFrom(values, PointerVectorLimits{MaxEntries: 5})
	if err != nil {
		t.Fatal(err)
	}
	forward, err := vector.Entries(1, 3, false)
	if err != nil || len(forward) != 2 || forward[0].Index != 1 || forward[0].Value != values[1] || forward[1].Index != 2 || forward[1].Value != values[2] {
		t.Fatal(forward, err)
	}
	reverse, err := vector.Entries(1, 4, true)
	if err != nil || len(reverse) != 3 || reverse[0].Index != 3 || reverse[2].Index != 1 {
		t.Fatal(reverse, err)
	}
	if _, err = vector.Entries(0, 5, false); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	short, _ := NewPointerVectorFrom(values[:3], PointerVectorLimits{})
	if _, err = ZipPointerVectors(vector, short); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}

	before, _ := vector.Snapshot()
	if err = vector.PushTwo(values[0], values[1]); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if after, _ := vector.Snapshot(); !slices.Equal(after, before) {
		t.Fatal("failed PushTwo mutated vector", after, before)
	}
	if err = vector.Append(short); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if after, _ := vector.Snapshot(); !slices.Equal(after, before) {
		t.Fatal("failed Append mutated vector", after, before)
	}

	self, _ := NewPointerVectorFrom(values[:2], PointerVectorLimits{})
	if err = self.Append(self); err != nil {
		t.Fatal(err)
	}
	if doubled, _ := self.Snapshot(); !slices.Equal(doubled, []*int{values[0], values[1], values[0], values[1]}) {
		t.Fatal(doubled)
	}

	cancelled := false
	stopped, _ := NewPointerVectorFrom(values[:2], PointerVectorLimits{Stop: func() bool { return cancelled }})
	cancelled = true
	stoppedBefore, _ := stopped.Snapshot()
	if err = stopped.Fill(4, values[3]); !errors.Is(err, ErrCancelled) {
		t.Fatal(err)
	}
	if stoppedAfter, _ := stopped.Snapshot(); !slices.Equal(stoppedAfter, stoppedBefore) {
		t.Fatal("cancelled Fill mutated vector", stoppedAfter, stoppedBefore)
	}
}

func pointerVectorModel(data []byte) error {
	if len(data) > 3000 {
		data = data[:3000]
	}
	poolValues := [8]int{}
	pool := make([]*int, len(poolValues)+1)
	for index := range poolValues {
		poolValues[index] = index
		pool[index+1] = &poolValues[index]
	}
	vector, err := NewExactPointerVector[int](0, PointerVectorLimits{MaxEntries: 4096})
	if err != nil {
		return err
	}
	model := make([]*int, 0)
	for offset := 0; offset+2 < len(data); offset += 3 {
		op := data[offset] % 10
		value := pool[int(data[offset+1])%len(pool)]
		position := int(data[offset+2])
		switch op {
		case 0:
			if err = vector.Push(value); err != nil {
				return err
			}
			model = append(model, value)
		case 1:
			if err = vector.PushFirst(value); err != nil {
				return err
			}
			model = append([]*int{value}, model...)
		case 2:
			if len(model) != 0 {
				index := position % len(model)
				if err = vector.Drop(index); err != nil {
					return err
				}
				model = append(model[:index], model[index+1:]...)
			}
		case 3:
			if len(model) != 0 {
				index := position % len(model)
				if err = vector.Insert(index, value); err != nil {
					return err
				}
				model = append(model, nil)
				copy(model[index+1:], model[index:])
				model[index] = value
			}
		case 4:
			index := position % 32
			if err = vector.Set(index, value); err != nil {
				return err
			}
			for len(model) <= index {
				model = append(model, nil)
			}
			model[index] = value
		case 5:
			if len(model) != 0 {
				size := position % (len(model) + 1)
				if err = vector.Shrink(size); err != nil {
					return err
				}
				model = model[:size]
			}
		case 6:
			if err = vector.Clear(); err != nil {
				return err
			}
			model = nil
		case 7:
			if len(model) > 1 {
				count := position % len(model)
				if err = vector.Reorder(count); err != nil {
					return err
				}
				model = append(append([]*int(nil), model[count:]...), model[:count]...)
			}
		case 8:
			if err = vector.Reverse(); err != nil {
				return err
			}
			slices.Reverse(model)
		case 9:
			duplicate, pushErr := vector.PushUnique(value)
			if pushErr != nil {
				return pushErr
			}
			wantDuplicate := slices.Contains(model, value)
			if duplicate != wantDuplicate {
				return fmt.Errorf("duplicate %v want %v", duplicate, wantDuplicate)
			}
			if !wantDuplicate {
				model = append(model, value)
			}
		}
		snapshot, snapshotErr := vector.Snapshot()
		if snapshotErr != nil || !slices.Equal(snapshot, model) {
			return fmt.Errorf("snapshot %v want %v: %w", snapshot, model, snapshotErr)
		}
		if err = vector.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func TestPointerVectorModelCampaign(t *testing.T) {
	state := uint64(0x94d049bb133111eb)
	for run := 0; run < 5000; run++ {
		data := make([]byte, 96)
		for index := range data {
			state ^= state << 7
			state ^= state >> 9
			state ^= state << 8
			data[index] = byte(state)
		}
		if err := pointerVectorModel(data); err != nil {
			t.Fatal(run, err)
		}
	}
}

func FuzzPointerVectorModel(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte{0, 1, 0, 0, 2, 0, 3, 3, 0, 7, 0, 1})
	f.Add([]byte{4, 1, 31, 8, 0, 0, 9, 1, 0, 6, 0, 0})
	f.Fuzz(func(t *testing.T, data []byte) {
		if err := pointerVectorModel(data); err != nil {
			t.Fatal(err)
		}
	})
}

func pointerSimulationModel(data []byte) error {
	if len(data) > 1024 {
		data = data[:1024]
	}
	entries, words := 2, 1
	if len(data) > 0 {
		entries += int(data[0] % 4)
	}
	if len(data) > 1 {
		words += int(data[1] % 4)
	}
	info, err := NewPointerSimulationInfo(entries, words, PointerSimulationLimits{MaxEntries: 64, MaxWords: 64, MaxCells: 4096})
	if err != nil {
		return err
	}
	model := make([][]uint32, entries)
	for index := range model {
		model[index] = make([]uint32, words)
	}
	for offset := 2; offset+2 < len(data); offset += 3 {
		switch data[offset] % 5 {
		case 0, 1:
			first := int(data[offset+1]) % (words + 1)
			value := uint32(0)
			if data[offset]%5 == 1 {
				value = ^uint32(0)
			}
			if value == 0 {
				err = info.CleanFrom(first)
			} else {
				err = info.FillFrom(first)
			}
			if err != nil {
				return err
			}
			for row := range model {
				for word := first; word < words; word++ {
					model[row][word] = value
				}
			}
		case 2:
			row := int(data[offset+1]) % entries
			values := make([]uint32, words)
			for word := range values {
				values[word] = uint32(data[offset+2])<<24 | uint32(row)<<16 | uint32(word)
			}
			if err = info.Write(row, values); err != nil {
				return err
			}
			copy(model[row], values)
		case 3:
			if entries <= 32 {
				if err = info.DoubleEntries(); err != nil {
					return err
				}
				oldEntries := entries
				entries *= 2
				for len(model) < entries {
					model = append(model, make([]uint32, words))
				}
				if oldEntries == 0 {
					return ErrInvariant
				}
			}
		case 4:
			if words <= 32 {
				if err = info.DoubleWords(); err != nil {
					return err
				}
				words *= 2
				for row := range model {
					model[row] = append(model[row], make([]uint32, words-len(model[row]))...)
				}
			}
		}
		if gotEntries, entryErr := info.EntryCount(); entryErr != nil || gotEntries != entries {
			return fmt.Errorf("entries %d want %d: %w", gotEntries, entries, entryErr)
		}
		if gotWords, wordErr := info.WordCount(); wordErr != nil || gotWords != words {
			return fmt.Errorf("words %d want %d: %w", gotWords, words, wordErr)
		}
		for row := range model {
			got, rowErr := info.Row(row)
			if rowErr != nil || !slices.Equal(got, model[row]) {
				return fmt.Errorf("row %d = %v want %v: %w", row, got, model[row], rowErr)
			}
		}
		if err = info.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func TestPointerSimulationModelCampaign(t *testing.T) {
	state := uint64(0xd1b54a32d192ed03)
	for run := 0; run < 5000; run++ {
		data := make([]byte, 50)
		for index := range data {
			state ^= state << 7
			state ^= state >> 9
			state ^= state << 8
			data[index] = byte(state)
		}
		if err := pointerSimulationModel(data); err != nil {
			t.Fatal(run, err)
		}
	}
}

func FuzzPointerSimulationModel(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte{3, 3, 1, 0, 0, 2, 1, 7, 3, 0, 0, 4, 0, 0})
	f.Add([]byte{0, 0, 2, 1, 255, 1, 0, 0, 0, 0, 0})
	f.Fuzz(func(t *testing.T, data []byte) {
		if err := pointerSimulationModel(data); err != nil {
			t.Fatal(err)
		}
	})
}

func TestPointerVectorRecordsHaveNoRuntimeInterfaces(t *testing.T) {
	// This is a compile-time shape check: values remain concrete pointers.
	type object struct{ value int }
	vector, _ := NewPointerVectorFrom([]*object{new(object)}, PointerVectorLimits{})
	records, err := vector.Entries(0, 1, false)
	if err != nil || records[0].Value != vector.slots[0] {
		t.Fatal(err)
	}
}
