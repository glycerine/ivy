package abc

// Go-only models for src/map/if/ifCache.c.

import (
	"errors"
	"reflect"
	"slices"
	"testing"
)

func TestIfCutCacheRecordAndAnalyze(t *testing.T) {
	cache, err := NewIfCutCache(3, IfCutCacheOptions{})
	if err != nil {
		t.Fatal(err)
	}
	for _, record := range []IfCutCacheRecord{
		{DSD0: 5, DSD1: 2, Shared: 1, Result: 7},
		{DSD0: 2, DSD1: 5, Shared: 1, Result: 7},
		{DSD0: 2, DSD1: 5, Shared: 2, Result: 7},
		{DSD0: 9, DSD1: 1, Shared: 1, Result: 3},
	} {
		if err = cache.Record(record.DSD0, record.DSD1, record.Shared, record.Result); err != nil {
			t.Fatal(err)
		}
	}
	records, err := cache.Records()
	if err != nil || records[0].DSD0 != 2 || records[0].DSD1 != 5 || records[3].DSD0 != 1 || records[3].DSD1 != 9 {
		t.Fatal(records, err)
	}
	statistics, err := cache.Analyze()
	if err != nil || statistics.Entries != 4 || statistics.Unique != 3 || len(statistics.Buckets) != 4 {
		t.Fatal(statistics, err)
	}
	if got := statistics.Buckets[1]; got.Entries != 3 || got.Unique != 2 || got.EntryPercent != 75 || got.UniquePercent != 100*float64(2)/3 {
		t.Fatal(got)
	}
	if got := statistics.Buckets[2]; got.Entries != 1 || got.Unique != 1 || got.EntryPercent != 25 || got.UniquePercent != 100 {
		t.Fatal(got)
	}
	records[0].Result = 99
	again, _ := cache.Records()
	if again[0].Result == 99 {
		t.Fatal("record snapshot aliases cache")
	}
}

func TestIfCutCacheEmptyLimitsAndCancellation(t *testing.T) {
	empty, err := NewIfCutCache(2, IfCutCacheOptions{})
	statistics, analyzeErr := empty.Analyze()
	if err != nil || analyzeErr != nil || statistics.Entries != 0 || statistics.Unique != 0 || len(statistics.Buckets) != 3 {
		t.Fatal(statistics, err, analyzeErr)
	}
	for _, options := range []struct {
		lut int
		op  IfCutCacheOptions
	}{
		{-1, IfCutCacheOptions{}}, {2, IfCutCacheOptions{MaxRecords: -1}},
	} {
		if _, err := NewIfCutCache(options.lut, options.op); !errors.Is(err, ErrDimension) {
			t.Fatal(options, err)
		}
	}
	limited, _ := NewIfCutCache(2, IfCutCacheOptions{MaxRecords: 1})
	if err := limited.Record(0, 1, 0, 2); err != nil {
		t.Fatal(err)
	}
	before, _ := limited.Records()
	if err := limited.Record(1, 2, 1, 3); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	after, _ := limited.Records()
	if !reflect.DeepEqual(before, after) {
		t.Fatal("failed record mutated cache")
	}
	if err := limited.Record(0, 1, 3, 2); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	stop := false
	cancelled, _ := NewIfCutCache(2, IfCutCacheOptions{Stop: func() bool { return stop }})
	if err := cancelled.Record(0, 1, 1, 2); err != nil {
		t.Fatal(err)
	}
	stop = true
	if err := cancelled.Record(1, 2, 1, 3); !errors.Is(err, ErrCancelled) {
		t.Fatal(err)
	}
	if _, err := cancelled.Analyze(); !errors.Is(err, ErrCancelled) {
		t.Fatal(err)
	}
	var nilCache *IfCutCache
	if _, err := nilCache.Analyze(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
}

func FuzzIfCutCacheAnalyze(f *testing.F) {
	for _, seed := range [][]byte{nil, {1, 2, 0, 3}, {5, 2, 1, 7, 2, 5, 1, 7}} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 256 {
			return
		}
		cache, err := NewIfCutCache(4, IfCutCacheOptions{MaxRecords: 64})
		if err != nil {
			t.Fatal(err)
		}
		var model []IfCutCacheRecord
		for offset := 0; offset+3 < len(data) && len(model) < 64; offset += 4 {
			a, b := int(int8(data[offset])), int(int8(data[offset+1]))
			if a > b {
				a, b = b, a
			}
			record := IfCutCacheRecord{DSD0: a, DSD1: b, Shared: int(data[offset+2] % 5), Result: int(int8(data[offset+3]))}
			if err := cache.Record(record.DSD0, record.DSD1, record.Shared, record.Result); err != nil {
				t.Fatal(err)
			}
			model = append(model, record)
		}
		statistics, err := cache.Analyze()
		if err != nil || statistics.Entries != len(model) {
			t.Fatal(statistics, err)
		}
		unique := make([]IfCutCacheRecord, 0, len(model))
		bucketUnique := make([][]IfCutCacheRecord, 5)
		entries := make([]int, 5)
		for _, record := range model {
			entries[record.Shared]++
			if !slices.Contains(unique, record) {
				unique = append(unique, record)
			}
			if !slices.Contains(bucketUnique[record.Shared], record) {
				bucketUnique[record.Shared] = append(bucketUnique[record.Shared], record)
			}
		}
		if statistics.Unique != len(unique) {
			t.Fatal(statistics.Unique, len(unique))
		}
		for shared, bucket := range statistics.Buckets {
			if bucket.Entries != entries[shared] || bucket.Unique != len(bucketUnique[shared]) {
				t.Fatal(shared, bucket, entries[shared], len(bucketUnique[shared]))
			}
		}
	})
}
