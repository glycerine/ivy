package abc

// Standalone layout, state-model, and fuzz tests for src/misc/vec/vecSet.h.

import (
	"errors"
	"slices"
	"testing"
)

func TestPagedWordSetLayoutPackingRolloverAndMemory(t *testing.T) {
	set, err := NewPagedWordSet(9, PagedWordSetLimits{MaxPages: 8, MaxEntries: 100, MaxBytes: 8 << 12})
	if err != nil {
		t.Fatal(err)
	}
	if words, err := WordCountForInts(0); err != nil || words != 0 {
		t.Fatal(words, err)
	}
	if words, err := WordCountForInts(5); err != nil || words != 3 {
		t.Fatal(words, err)
	}
	handle, err := set.AppendInts([]int32{-1, 2, -3, 4, 5})
	if err != nil || handle != 2 {
		t.Fatal(handle, err)
	}
	if values, err := set.SnapshotInts(handle, 5); err != nil || !slices.Equal(values, []int32{-1, 2, -3, 4, 5}) {
		t.Fatal(values, err)
	}
	page, _ := set.HandlePage(handle)
	shift, _ := set.HandleShift(handle)
	if page != 0 || shift != 2 {
		t.Fatal(page, shift)
	}
	large := make([]int32, 1010) // 505 words; current limit becomes 510.
	for index := range large {
		large[index] = int32(index*17 - 9000)
	}
	largeHandle, err := set.AppendInts(large)
	if err != nil || largeHandle != 5 {
		t.Fatal(largeHandle, err)
	}
	lastPageZero, err := set.AppendInts([]int32{7, 8})
	if err != nil || lastPageZero != 510 {
		t.Fatal(lastPageZero, err)
	}
	pageOne, err := set.AppendInts([]int32{9})
	if err != nil || pageOne != 514 {
		t.Fatal(pageOne, err)
	}
	if values, err := set.SnapshotInts(largeHandle, len(large)); err != nil || !slices.Equal(values, large) {
		t.Fatal(len(values), err)
	}
	if current, err := set.CurrentHandle(); err != nil || current != 515 {
		t.Fatal(current, err)
	}
	if memory, err := set.MemoryBytes(); err != nil || memory != 515*8 {
		t.Fatal(memory, err)
	}
	if allocated, err := set.AllocatedPageBytes(); err != nil || allocated != 2*512*8 {
		t.Fatal(allocated, err)
	}
	if reported, err := set.SourceReportedBytes(); err != nil || reported != 32+256*8+2*512*8 {
		t.Fatal(reported, err)
	}
	if count, err := set.EntryCount(); err != nil || count != 4 {
		t.Fatal(count, err)
	}
	if set.Validate() != nil {
		t.Fatal(set.Validate())
	}
}

func TestPagedWordSetBytesStringsShadowCompactionAndIteration(t *testing.T) {
	set, err := NewPagedWordSet(9, PagedWordSetLimits{})
	if err != nil {
		t.Fatal(err)
	}
	first, _ := set.AppendInts([]int32{10, 11, 12, 13})
	second, _ := set.AppendInts([]int32{20, 21})
	third, _ := set.AppendInts([]int32{30, 31, 32, 33, 34, 35})
	stringHandle, err := set.SaveString("abc\x00def")
	if err != nil {
		t.Fatal(err)
	}
	bytes, err := set.SnapshotBytes(stringHandle, 8)
	if err != nil || !slices.Equal(bytes, []byte{'a', 'b', 'c', 0, 0, 0, 0, 0}) {
		t.Fatal(bytes, err)
	}
	if err = set.ShrinkShadow(2); err != nil {
		t.Fatal(err)
	}
	newSecond, err := set.ReserveShadowInts(2)
	if err != nil || newSecond != 2 {
		t.Fatal(newSecond, err)
	}
	if err = set.CopyWords(newSecond, second, 1); err != nil {
		t.Fatal(err)
	}
	newThird, err := set.ReserveShadowInts(6)
	if err != nil || newThird != 3 {
		t.Fatal(newThird, err)
	}
	if err = set.CopyWords(newThird, third, 3); err != nil {
		t.Fatal(err)
	}
	shadow, err := set.ShadowHandle()
	if err != nil || shadow != 6 {
		t.Fatal(shadow, err)
	}
	if memory, err := set.ShadowMemoryBytes(); err != nil || memory != 48 {
		t.Fatal(memory, err)
	}
	if err = set.Shrink(shadow); err != nil {
		t.Fatal(err)
	}
	if err = set.CommitShadowLimits(); err != nil {
		t.Fatal(err)
	}
	if err = set.SetEntryCount(2); err != nil {
		t.Fatal(err)
	}
	if values, err := set.SnapshotInts(newSecond, 2); err != nil || !slices.Equal(values, []int32{20, 21}) {
		t.Fatal(values, err)
	}
	if values, err := set.SnapshotInts(newThird, 6); err != nil || !slices.Equal(values, []int32{30, 31, 32, 33, 34, 35}) {
		t.Fatal(values, err)
	}
	sizes := map[int]int{newSecond: 1, newThird: 3}
	handles, err := set.PrimaryHandles(func(handle int) (int, error) { return sizes[handle], nil }, 10)
	if err != nil || !slices.Equal(handles, []int{newSecond, newThird}) {
		t.Fatal(handles, err)
	}
	if first == newSecond {
		// Compaction is intentionally overlap-safe and overwrote the old first.
		if values, _ := set.SnapshotInts(first, 2); !slices.Equal(values, []int32{20, 21}) {
			t.Fatal(values)
		}
	}
	if err = set.Restart(); err != nil {
		t.Fatal(err)
	}
	if current, _ := set.CurrentHandle(); current != 2 {
		t.Fatal(current)
	}
	if count, _ := set.EntryCount(); count != 0 {
		t.Fatal(count)
	}
	reused, err := set.AppendInts([]int32{99})
	if err != nil || reused != 2 {
		t.Fatal(reused, err)
	}
	if err = set.Restart(); err != nil {
		t.Fatal(err)
	}
	if _, err = set.AppendZeroInts(1018); err != nil {
		t.Fatal(err)
	}
	if rollover, appendErr := set.AppendInts([]int32{100}); appendErr != nil || rollover != 514 {
		t.Fatal(rollover, appendErr)
	}
	if _, err = set.ShadowHandle(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
}

func TestPagedWordSetErrorsLimitsCancellationAndLifecycle(t *testing.T) {
	empty, err := NewPagedWordSet(9, PagedWordSetLimits{})
	if err != nil {
		t.Fatal(err)
	}
	if handle, appendErr := empty.AppendInts(nil); appendErr != nil || handle != 2 {
		t.Fatal(handle, appendErr)
	}
	if current, currentErr := empty.CurrentHandle(); currentErr != nil || current != 2 {
		t.Fatal(current, currentErr)
	}
	if count, countErr := empty.EntryCount(); countErr != nil || count != 1 {
		t.Fatal(count, countErr)
	}

	for _, create := range []func() error{
		func() error { _, e := NewPagedWordSet(8, PagedWordSetLimits{}); return e },
		func() error { _, e := NewPagedWordSet(31, PagedWordSetLimits{}); return e },
		func() error { _, e := NewPagedWordSet(9, PagedWordSetLimits{MaxPages: -1}); return e },
		func() error { _, e := NewPagedWordSet(9, PagedWordSetLimits{MaxBytes: 100}); return e },
		func() error {
			_, e := NewPagedWordSet(9, PagedWordSetLimits{Stop: func() bool { return true }})
			return e
		},
	} {
		if err := create(); err == nil {
			t.Fatal("expected constructor error")
		}
	}
	set, err := NewPagedWordSet(9, PagedWordSetLimits{MaxPages: 1, MaxEntries: 2, MaxBytes: 4096})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = set.AppendZeroInts(1020); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, err = set.AppendInts([]int32{1}); err != nil {
		t.Fatal(err)
	}
	if _, err = set.AppendInts([]int32{2}); err != nil {
		t.Fatal(err)
	}
	if _, err = set.AppendInts([]int32{3}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if err = set.SetEntryCount(-1); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if err = set.SetEntryCount(3); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, err = set.SnapshotWords(1, 1); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if err = set.Shrink(1); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err = set.ReserveShadowInts(2); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if err = set.CommitShadowLimits(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	calls := 0
	cancelled, err := NewPagedWordSet(9, PagedWordSetLimits{Stop: func() bool { calls++; return calls > 1 }})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = cancelled.AppendInts([]int32{1}); !errors.Is(err, ErrCancelled) {
		t.Fatal(calls, err)
	}
	if err = set.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err = set.CurrentHandle(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
}

func FuzzPagedWordSetAgainstAppendModel(f *testing.F) {
	f.Add([]byte{1, 2, 3, 4, 5, 6, 7})
	f.Add([]byte{255, 0, 128, 64, 32})
	f.Fuzz(func(t *testing.T, input []byte) {
		if len(input) > 512 {
			input = input[:512]
		}
		set, err := NewPagedWordSet(9, PagedWordSetLimits{MaxPages: 16, MaxEntries: 1024, MaxBytes: 16 * 512 * 8})
		if err != nil {
			t.Fatal(err)
		}
		type modelRecord struct {
			handle int
			values []int32
		}
		records := []modelRecord{}
		for index := 0; index < len(input); {
			count := int(input[index]%31) + 1
			index++
			values := make([]int32, count)
			for valueIndex := range values {
				if index < len(input) {
					values[valueIndex] = int32(int8(input[index]))
					index++
				} else {
					values[valueIndex] = int32(valueIndex)
				}
			}
			handle, appendErr := set.AppendInts(values)
			if appendErr != nil {
				t.Fatal(appendErr)
			}
			records = append(records, modelRecord{handle, values})
			for _, record := range records {
				got, snapshotErr := set.SnapshotInts(record.handle, len(record.values))
				if snapshotErr != nil || !slices.Equal(got, record.values) {
					t.Fatal(record.handle, got, record.values, snapshotErr)
				}
			}
			if validateErr := set.Validate(); validateErr != nil {
				t.Fatal(validateErr)
			}
		}
		if count, countErr := set.EntryCount(); countErr != nil || count != len(records) {
			t.Fatal(count, len(records), countErr)
		}
		if err = set.ShrinkShadow(2); err != nil {
			t.Fatal(err)
		}
		retained := make([]modelRecord, 0, len(records))
		for index, record := range records {
			keep := len(input) == 0 || input[index%len(input)]>>uint(index%8)&1 != 0
			if !keep {
				continue
			}
			destination, reserveErr := set.ReserveShadowInts(len(record.values))
			if reserveErr != nil {
				t.Fatal(index, reserveErr)
			}
			words, _ := WordCountForInts(len(record.values))
			if copyErr := set.CopyWords(destination, record.handle, words); copyErr != nil {
				t.Fatal(index, copyErr)
			}
			retained = append(retained, modelRecord{destination, record.values})
		}
		shadow, err := set.ShadowHandle()
		if err != nil {
			t.Fatal(err)
		}
		if err = set.Shrink(shadow); err != nil {
			t.Fatal(err)
		}
		if err = set.CommitShadowLimits(); err != nil {
			t.Fatal(err)
		}
		if err = set.SetEntryCount(len(retained)); err != nil {
			t.Fatal(err)
		}
		for _, record := range retained {
			got, snapshotErr := set.SnapshotInts(record.handle, len(record.values))
			if snapshotErr != nil || !slices.Equal(got, record.values) {
				t.Fatal("compacted", record.handle, got, record.values, snapshotErr)
			}
		}
		if validateErr := set.Validate(); validateErr != nil {
			t.Fatal(validateErr)
		}
	})
}
