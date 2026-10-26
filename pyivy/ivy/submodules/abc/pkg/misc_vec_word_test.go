package abc

// Go-only models and fuzz tests for src/misc/vec/vecWrd.h.

import (
	"errors"
	"math"
	"path/filepath"
	"slices"
	"testing"
	"unsafe"
)

func TestWordVectorConstructorsCapacityRandomTruthAndMemory(t *testing.T) {
	ordinary, err := NewWordVector(1, WordVectorLimits{})
	if err != nil {
		t.Fatal(err)
	}
	if capacity, _ := ordinary.Capacity(); capacity != 16 {
		t.Fatal(capacity)
	}
	exact, _ := NewExactWordVector(1, WordVectorLimits{})
	if capacity, _ := exact.Capacity(); capacity != 1 {
		t.Fatal(capacity)
	}
	zero, _ := NewFilledWordVector(3, 0, WordVectorLimits{})
	full, _ := NewFilledWordVector(3, math.MaxUint64, WordVectorLimits{})
	natural, _ := NewNaturalWordVector(4, WordVectorLimits{})
	if values, _ := zero.Snapshot(); !slices.Equal(values, []uint64{0, 0, 0}) {
		t.Fatal(values)
	}
	if values, _ := full.Snapshot(); !slices.Equal(values, []uint64{math.MaxUint64, math.MaxUint64, math.MaxUint64}) {
		t.Fatal(values)
	}
	if values, _ := natural.Snapshot(); !slices.Equal(values, []uint64{0, 1, 2, 3}) {
		t.Fatal(values)
	}

	random := &ABCRandom{}
	wantRandom := &ABCRandom{}
	generated, err := NewRandomWordVector(5, random, WordVectorLimits{})
	if err != nil {
		t.Fatal(err)
	}
	want := make([]uint64, 5)
	for index := range want {
		want[index] = wantRandom.Uint64(false)
	}
	if values, _ := generated.Snapshot(); !slices.Equal(values, want) {
		t.Fatal(values, want)
	}

	truth, err := NewElementaryTruthWordVector(7, WordVectorLimits{})
	if err != nil {
		t.Fatal(err)
	}
	values, _ := truth.Snapshot()
	if len(values) != 14 {
		t.Fatal(len(values))
	}
	for variable, mask := range []uint64{0xaaaaaaaaaaaaaaaa, 0xcccccccccccccccc, 0xf0f0f0f0f0f0f0f0, 0xff00ff00ff00ff00, 0xffff0000ffff0000} {
		if values[2*variable] != mask || values[2*variable+1] != mask {
			t.Fatal(variable, values[2*variable:2*variable+2])
		}
	}
	if values[10] != 0xffffffff00000000 || values[11] != 0xffffffff00000000 || values[12] != 0 || values[13] != math.MaxUint64 {
		t.Fatal(values[10:])
	}
	pointerBytes := int(unsafe.Sizeof(uintptr(0)))
	if memory, memoryErr := ordinary.NativeMemoryBytes(); memoryErr != nil || memory != 8+pointerBytes+16*8 {
		t.Fatal(memory, memoryErr)
	}
}

func TestWordVectorGrowthAccessMutationAndOwnership(t *testing.T) {
	vector, _ := NewExactWordVector(0, WordVectorLimits{})
	for value := uint64(1); value <= 17; value++ {
		if err := vector.Push(value); err != nil {
			t.Fatal(err)
		}
	}
	if capacity, _ := vector.Capacity(); capacity != 32 {
		t.Fatal(capacity)
	}
	if err := vector.PushFirst(0); err != nil {
		t.Fatal(err)
	}
	if err := vector.Insert(2, 99); err != nil {
		t.Fatal(err)
	}
	if value, _ := vector.Add(2, 1); value != 100 {
		t.Fatal(value)
	}
	if removed, err := vector.Remove(100); err != nil || !removed {
		t.Fatal(removed, err)
	}
	if found := vector.Find(10); found != 10 {
		t.Fatal(found)
	}
	if value, _ := vector.Pop(); value != 17 {
		t.Fatal(value)
	}
	if err := vector.Set(20, 200); err != nil {
		t.Fatal(err)
	}
	if value, _ := vector.Get(19); value != 0 {
		t.Fatal(value)
	}
	if capacity, _ := vector.Capacity(); capacity != 32 {
		t.Fatal(capacity)
	}
	if err := vector.Shrink(3); err != nil {
		t.Fatal(err)
	}
	start, err := vector.Fetch(2)
	if err != nil || start != 3 {
		t.Fatal(start, err)
	}
	if values, _ := vector.Snapshot(); !slices.Equal(values, []uint64{0, 1, 2, 0, 0}) {
		t.Fatal(values)
	}
	if _, err = vector.Fetch(100); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if err = vector.Fill(4, 7); err != nil {
		t.Fatal(err)
	}
	if err = vector.FillExtra(7, 8); err != nil {
		t.Fatal(err)
	}
	if values, _ := vector.Snapshot(); !slices.Equal(values, []uint64{7, 7, 7, 7, 8, 8, 8}) {
		t.Fatal(values)
	}
	if err = vector.Reverse(); err != nil {
		t.Fatal(err)
	}
	if entries, _ := vector.Entries(1, 6, true); !slices.Equal(entries, []uint64{7, 7, 7, 8, 8}) {
		t.Fatal(entries)
	}

	clone, _ := vector.Clone()
	_ = clone.Write(0, 100)
	if value, _ := vector.Entry(0); value == 100 {
		t.Fatal("clone aliases source")
	}
	moved, err := clone.Move()
	if err != nil {
		t.Fatal(err)
	}
	if length, _ := clone.Len(); length != 0 {
		t.Fatal(length)
	}
	if length, _ := moved.Len(); length != 7 {
		t.Fatal(length)
	}
	released, err := moved.Release()
	if err != nil || len(released) != 7 {
		t.Fatal(released, err)
	}
	released[0] = 123
	if length, _ := moved.Len(); length != 0 {
		t.Fatal(length)
	}
}

func TestWordVectorBulkAppendPreservesNativeGrowth(t *testing.T) {
	vector, err := NewExactWordVector(16, WordVectorLimits{})
	if err != nil {
		t.Fatal(err)
	}
	if err = vector.PushSlice(make([]uint64, 17)); err != nil {
		t.Fatal(err)
	}
	if capacity, capacityErr := vector.Capacity(); capacityErr != nil || capacity != 32 {
		t.Fatal(capacity, capacityErr)
	}
	other, err := NewWordVectorFrom([]uint64{1, 2, 3}, WordVectorLimits{})
	if err != nil {
		t.Fatal(err)
	}
	if err = vector.Append(other); err != nil {
		t.Fatal(err)
	}
	if length, lengthErr := vector.Len(); lengthErr != nil || length != 20 {
		t.Fatal(length, lengthErr)
	}
	limited, err := NewExactWordVector(16, WordVectorLimits{MaxWords: 20})
	if err != nil {
		t.Fatal(err)
	}
	if err = limited.PushSlice(make([]uint64, 17)); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if length, lengthErr := limited.Len(); lengthErr != nil || length != 0 {
		t.Fatal(length, lengthErr)
	}
}

func TestWordVectorOrderingSetsInvertAndRecords(t *testing.T) {
	ordered, _ := NewWordVector(0, WordVectorLimits{})
	for _, value := range []uint64{5, 1, 3, 3, 2} {
		if err := ordered.PushOrdered(value); err != nil {
			t.Fatal(err)
		}
	}
	if values, _ := ordered.Snapshot(); !slices.Equal(values, []uint64{1, 2, 3, 3, 5}) {
		t.Fatal(values)
	}
	if added, err := ordered.PushUniqueOrdered(3); err != nil || added {
		t.Fatal(added, err)
	}
	if added, err := ordered.PushUniqueOrdered(4); err != nil || !added {
		t.Fatal(added, err)
	}
	if err := ordered.Uniqify(); err != nil {
		t.Fatal(err)
	}
	if values, _ := ordered.Snapshot(); !slices.Equal(values, []uint64{1, 2, 3, 4, 5}) {
		t.Fatal(values)
	}
	if max, _ := ordered.Maximum(); max != 5 {
		t.Fatal(max)
	}
	if min, _ := ordered.Minimum(); min != 1 {
		t.Fatal(min)
	}
	if sum, _ := ordered.Sum(); sum != 15 {
		t.Fatal(sum)
	}
	if zeros, _ := ordered.CountZero(); zeros != 0 {
		t.Fatal(zeros)
	}
	inverted, err := ordered.Invert(math.MaxUint64)
	if err != nil {
		t.Fatal(err)
	}
	if values, _ := inverted.Snapshot(); !slices.Equal(values, []uint64{math.MaxUint64, 0, 1, 2, 3, 4}) {
		t.Fatal(values)
	}
	other, _ := NewWordVectorFrom([]uint64{5, 1, 9}, WordVectorLimits{})
	if common, err := ordered.CountCommon(other); err != nil || common != 2 {
		t.Fatal(common, err)
	}
	if equal, _ := ordered.Equal(other); equal {
		t.Fatal(equal)
	}
	if err := ordered.Sort(true); err != nil {
		t.Fatal(err)
	}
	if values, _ := ordered.Snapshot(); !slices.Equal(values, []uint64{5, 4, 3, 2, 1}) {
		t.Fatal(values)
	}

	records, _ := NewWordVectorFrom([]uint64{1, 2, 3, 4, 1, 2, 5, 6, 3, 4}, WordVectorLimits{})
	count, mapping, err := records.UniqueRecords(2)
	if err != nil || count != 3 || !slices.Equal(mapping, []int{0, 1, 0, 3, 1}) {
		t.Fatal(count, mapping, err)
	}
	unique, err := records.UniqifyRecords(2)
	if err != nil {
		t.Fatal(err)
	}
	if values, _ := unique.Snapshot(); !slices.Equal(values, []uint64{1, 2, 3, 4, 5, 6}) {
		t.Fatal(values)
	}
	if capacity, capacityErr := unique.Capacity(); capacityErr != nil || capacity != 8 {
		t.Fatal(capacity, capacityErr)
	}
	if _, _, err = records.UniqueRecords(3); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
}

func TestWordVectorHexRoundTripAndFiles(t *testing.T) {
	vector, _ := NewWordVectorFrom([]uint64{0x0123456789abcdef, 0xfedcba9876543210, 7, 8}, WordVectorLimits{})
	text, err := vector.FormatHex(2)
	want := "FEDCBA98765432100123456789ABCDEF\n70000000000000008000000000000000\n"
	if err != nil || text != want {
		t.Fatalf("%q %v", text, err)
	}
	parsed, width, err := ParseWordVectorHex([]byte(text), WordVectorLimits{})
	if err != nil || width != 2 {
		t.Fatal(width, err)
	}
	if equal, equalErr := vector.Equal(parsed); equalErr != nil || !equal {
		t.Fatal(equal, equalErr)
	}
	withoutNewline, width, err := ParseWordVectorHex([]byte("1a"), WordVectorLimits{})
	if err != nil || width != 1 {
		t.Fatal(width, err)
	}
	if value, _ := withoutNewline.Entry(0); value != 0xa1 {
		t.Fatal(value)
	}
	if _, _, err = ParseWordVectorHex([]byte("X\n"), WordVectorLimits{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, _, err = ParseWordVectorHex([]byte("1\n12\n12345678901234567\n"), WordVectorLimits{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "vectors.hex")
	if records, dumpErr := vector.DumpHexFile(path, 2); dumpErr != nil || records != 2 {
		t.Fatal(records, dumpErr)
	}
	loaded, loadedWidth, err := ReadWordVectorHexFile(path, WordVectorLimits{})
	if err != nil || loadedWidth != 2 {
		t.Fatal(loadedWidth, err)
	}
	if equal, _ := vector.Equal(loaded); !equal {
		t.Fatal("file round trip")
	}
}

func TestWordVectorLimitsCancellationAndLifecycle(t *testing.T) {
	if vector, err := NewWordVector(-1, WordVectorLimits{}); vector != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(vector, err)
	}
	if vector, err := NewWordVector(17, WordVectorLimits{MaxWords: 16}); vector != nil || !errors.Is(err, ErrLimit) {
		t.Fatal(vector, err)
	}
	limited, _ := NewExactWordVector(2, WordVectorLimits{MaxWords: 2})
	_ = limited.Push(1)
	_ = limited.Push(2)
	if err := limited.Push(3); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if err := limited.Set(3, 4); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, err := limited.Invert(0); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	cancel := false
	cancelled, _ := NewExactWordVector(0, WordVectorLimits{Stop: func() bool { return cancel }})
	cancel = true
	if err := cancelled.Push(1); !errors.Is(err, ErrCancelled) {
		t.Fatal(err)
	}
	if length, _ := cancelled.Len(); length != 0 {
		t.Fatal(length)
	}
	if err := limited.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := limited.Len(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
}

func wordVectorModel(data []byte) error {
	if len(data) > 3000 {
		data = data[:3000]
	}
	vector, err := NewExactWordVector(0, WordVectorLimits{MaxWords: 4096})
	if err != nil {
		return err
	}
	model := make([]uint64, 0)
	for offset := 0; offset+2 < len(data); offset += 3 {
		op := data[offset] % 8
		value := uint64(data[offset+2]) | uint64(data[offset+1])<<8
		switch op {
		case 0:
			if err = vector.Push(value); err != nil {
				return err
			}
			model = append(model, value)
		case 1:
			added, pushErr := vector.PushUnique(value)
			if pushErr != nil {
				return pushErr
			}
			want := !slices.Contains(model, value)
			if added != want {
				return errors.New("unique result")
			}
			if want {
				model = append(model, value)
			}
		case 2:
			index := int(data[offset+1]) % 32
			if err = vector.Set(index, value); err != nil {
				return err
			}
			for len(model) <= index {
				model = append(model, 0)
			}
			model[index] = value
		case 3:
			if len(model) > 0 {
				size := int(data[offset+1]) % (len(model) + 1)
				if err = vector.Shrink(size); err != nil {
					return err
				}
				model = model[:size]
			}
		case 4:
			if err = vector.Reverse(); err != nil {
				return err
			}
			slices.Reverse(model)
		case 5:
			removed, removeErr := vector.Remove(value)
			if removeErr != nil {
				return removeErr
			}
			index := slices.Index(model, value)
			if removed != (index >= 0) {
				return errors.New("remove result")
			}
			if index >= 0 {
				model = append(model[:index], model[index+1:]...)
			}
		case 6:
			reverse := value&1 != 0
			if err = vector.Sort(reverse); err != nil {
				return err
			}
			slices.Sort(model)
			if reverse {
				slices.Reverse(model)
			}
		case 7:
			if err = vector.Uniqify(); err != nil {
				return err
			}
			slices.Sort(model)
			model = slices.Compact(model)
		}
		got, snapshotErr := vector.Snapshot()
		if snapshotErr != nil || !slices.Equal(got, model) {
			return errors.New("model mismatch")
		}
		if err = vector.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func TestWordVectorModelCampaign(t *testing.T) {
	state := uint64(0xa0761d6478bd642f)
	for run := 0; run < 5000; run++ {
		data := make([]byte, 96)
		for index := range data {
			state ^= state << 7
			state ^= state >> 9
			state ^= state << 8
			data[index] = byte(state)
		}
		if err := wordVectorModel(data); err != nil {
			t.Fatal(run, err)
		}
	}
}

func FuzzWordVectorModel(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte{0, 0, 1, 0, 0, 2, 1, 0, 1, 6, 0, 1})
	f.Add([]byte{2, 7, 9, 3, 2, 0, 4, 0, 0, 7, 0, 0})
	f.Fuzz(func(t *testing.T, data []byte) {
		if err := wordVectorModel(data); err != nil {
			t.Fatal(err)
		}
	})
}
