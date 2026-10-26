package abc

// Deterministic cut-composition cache statistics from src/map/if/ifCache.c.

import "github.com/glycerine/insdict"

type IfCutCacheRecord struct {
	DSD0, DSD1 int
	Shared     int
	Result     int
}

type IfCutCacheOptions struct {
	MaxRecords int
	Stop       func() bool
}

type IfCutCacheBucketStatistics struct {
	Shared                      int
	Entries, Unique             int
	EntryPercent, UniquePercent float64
}

type IfCutCacheStatistics struct {
	Entries, Unique int
	Buckets         []IfCutCacheBucketStatistics
}

type IfCutCache struct {
	lutSize, maxRecords int
	stop                func() bool
	records             []IfCutCacheRecord
}

func NewIfCutCache(lutSize int, options IfCutCacheOptions) (*IfCutCache, error) {
	if lutSize < 0 || options.MaxRecords < 0 {
		return nil, ErrDimension
	}
	if options.MaxRecords == 0 {
		options.MaxRecords = 1 << 20
	}
	return &IfCutCache{lutSize: lutSize, maxRecords: options.MaxRecords, stop: options.Stop}, nil
}

func (cache *IfCutCache) check() error {
	if cache == nil || cache.lutSize < 0 || cache.maxRecords < 1 {
		return ErrInvariant
	}
	return nil
}

// Record preserves If_ManCacheRecord's canonical operand order.
func (cache *IfCutCache) Record(dsd0, dsd1, shared, result int) error {
	if err := cache.check(); err != nil {
		return err
	}
	if shared < 0 || shared > cache.lutSize {
		return ErrDimension
	}
	if cache.stop != nil && cache.stop() {
		return ErrCancelled
	}
	if len(cache.records) >= cache.maxRecords {
		return ErrLimit
	}
	if dsd0 > dsd1 {
		dsd0, dsd1 = dsd1, dsd0
	}
	cache.records = append(cache.records, IfCutCacheRecord{DSD0: dsd0, DSD1: dsd1, Shared: shared, Result: result})
	return nil
}

func (cache *IfCutCache) Records() ([]IfCutCacheRecord, error) {
	if err := cache.check(); err != nil {
		return nil, err
	}
	result := make([]IfCutCacheRecord, len(cache.records))
	copy(result, cache.records)
	return result, nil
}

// Analyze preserves If_ManCacheAnalize's first-occurrence tuple IDs and
// per-shared-input uniqueness accounting while returning data instead of
// writing process-global stdout.
func (cache *IfCutCache) Analyze() (IfCutCacheStatistics, error) {
	result := IfCutCacheStatistics{}
	if err := cache.check(); err != nil {
		return result, err
	}
	hashRecord := func(record IfCutCacheRecord) uint64 {
		hash := insdict.Mix64(uint64(int64(record.DSD0)))
		hash = insdict.Mix64(hash ^ uint64(int64(record.DSD1)))
		hash = insdict.Mix64(hash ^ uint64(int64(record.Shared)))
		return insdict.Mix64(hash ^ uint64(int64(record.Result)))
	}
	index := insdict.NewDictFuncSize[IfCutCacheRecord, int](hashRecord, min(len(cache.records), 1024))
	bucketSeen := make([]*insdict.Dict[int, bool], cache.lutSize+1)
	for shared := range bucketSeen {
		bucketSeen[shared] = insdict.NewDict[int, bool]()
	}
	entries := make([]int, cache.lutSize+1)
	for _, record := range cache.records {
		if cache.stop != nil && cache.stop() {
			return IfCutCacheStatistics{}, ErrCancelled
		}
		id, found := index.Get2(record)
		if !found {
			id = index.Len()
			index.Set(record, id)
		}
		entries[record.Shared]++
		bucketSeen[record.Shared].Set(id, true)
	}
	result.Entries, result.Unique = len(cache.records), index.Len()
	result.Buckets = make([]IfCutCacheBucketStatistics, cache.lutSize+1)
	for shared := range result.Buckets {
		bucket := IfCutCacheBucketStatistics{Shared: shared, Entries: entries[shared], Unique: bucketSeen[shared].Len()}
		bucket.EntryPercent = 100 * float64(bucket.Entries) / float64(max(1, result.Entries))
		bucket.UniquePercent = 100 * float64(bucket.Unique) / float64(max(1, bucket.Entries))
		result.Buckets[shared] = bucket
	}
	return result, nil
}
