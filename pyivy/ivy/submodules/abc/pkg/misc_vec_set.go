package abc

// Paged primary/shadow word storage from src/misc/vec/vecSet.h.

import "encoding/binary"

type PagedWordSetLimits struct {
	MaxPages   int
	MaxEntries int
	MaxBytes   int
	Stop       func() bool
}

type pagedWordSetPage struct {
	words       []uint64
	limit       int
	shadowLimit int
}

type PagedWordSet struct {
	pageBits, pageWords, pageMask int
	pages                         []pagedWordSetPage
	pageCapacity                  int
	entries                       int
	page, shadowPage              int
	limits                        PagedWordSetLimits
	closed                        bool
}

func normalizePagedWordSetLimits(limits PagedWordSetLimits) (PagedWordSetLimits, error) {
	if limits.MaxPages < 0 || limits.MaxEntries < 0 || limits.MaxBytes < 0 {
		return limits, ErrDimension
	}
	if limits.MaxPages == 0 {
		limits.MaxPages = 1 << 16
	}
	if limits.MaxEntries == 0 {
		limits.MaxEntries = 1 << 24
	}
	if limits.MaxBytes == 0 {
		limits.MaxBytes = 1 << 30
	}
	return limits, nil
}

func NewPagedWordSet(pageBits int, limits PagedWordSetLimits) (*PagedWordSet, error) {
	limits, err := normalizePagedWordSetLimits(limits)
	if err != nil || pageBits <= 8 || pageBits >= 31 {
		if err != nil {
			return nil, err
		}
		return nil, ErrDimension
	}
	if limits.Stop != nil && limits.Stop() {
		return nil, ErrCancelled
	}
	pageWords := 1 << uint(pageBits)
	if pageWords > limits.MaxBytes/8 || limits.MaxPages < 1 {
		return nil, ErrLimit
	}
	page := pagedWordSetPage{words: make([]uint64, pageWords), limit: 2, shadowLimit: -1}
	return &PagedWordSet{
		pageBits:     pageBits,
		pageWords:    pageWords,
		pageMask:     pageWords - 1,
		pages:        []pagedWordSetPage{page},
		pageCapacity: 256,
		limits:       limits,
	}, nil
}

func (set *PagedWordSet) check() error {
	if set == nil {
		return ErrDimension
	}
	if set.closed || set.pageBits <= 8 || set.pageWords != 1<<uint(set.pageBits) || set.pageMask != set.pageWords-1 || len(set.pages) == 0 || set.page < 0 || set.page >= len(set.pages) || set.shadowPage < 0 || set.shadowPage >= len(set.pages) || set.entries < 0 || set.entries > set.limits.MaxEntries {
		return ErrInvariant
	}
	return nil
}

func (set *PagedWordSet) HandlePage(handle int) (int, error) {
	if err := set.check(); err != nil {
		return 0, err
	}
	if handle < 0 {
		return 0, ErrDimension
	}
	return handle >> uint(set.pageBits), nil
}

func (set *PagedWordSet) HandleShift(handle int) (int, error) {
	if err := set.check(); err != nil {
		return 0, err
	}
	if handle < 0 {
		return 0, ErrDimension
	}
	return handle & set.pageMask, nil
}

func WordCountForInts(count int) (int, error) {
	if count < 0 {
		return 0, ErrDimension
	}
	return count/2 + count%2, nil
}

func (set *PagedWordSet) CurrentHandle() (int, error) {
	if err := set.check(); err != nil {
		return 0, err
	}
	return set.page<<uint(set.pageBits) + set.pages[set.page].limit, nil
}

func (set *PagedWordSet) ShadowHandle() (int, error) {
	if err := set.check(); err != nil {
		return 0, err
	}
	limit := set.pages[set.shadowPage].shadowLimit
	if limit < 2 {
		return 0, ErrInvariant
	}
	return set.shadowPage<<uint(set.pageBits) + limit, nil
}

func (set *PagedWordSet) EntryCount() (int, error) {
	if err := set.check(); err != nil {
		return 0, err
	}
	return set.entries, nil
}

func (set *PagedWordSet) SetEntryCount(entries int) error {
	if err := set.check(); err != nil {
		return err
	}
	if entries < 0 {
		return ErrDimension
	}
	if entries > set.limits.MaxEntries {
		return ErrLimit
	}
	set.entries = entries
	return nil
}

func (set *PagedWordSet) allocatePage() error {
	if len(set.pages) == set.limits.MaxPages || (len(set.pages)+1)*set.pageWords > set.limits.MaxBytes/8 {
		return ErrLimit
	}
	if set.limits.Stop != nil && set.limits.Stop() {
		return ErrCancelled
	}
	page := pagedWordSetPage{words: make([]uint64, set.pageWords), limit: 2, shadowLimit: -1}
	set.pages = append(set.pages, page)
	if len(set.pages) > set.pageCapacity {
		if set.pageCapacity > set.limits.MaxPages/2 {
			set.pageCapacity = set.limits.MaxPages
		} else {
			set.pageCapacity *= 2
		}
	}
	return nil
}

func (set *PagedWordSet) reservePrimaryWords(words int, countEntry bool) (int, error) {
	if err := set.check(); err != nil {
		return 0, err
	}
	if words < 0 {
		return 0, ErrDimension
	}
	// The source assertion checks only words < pageWords, but the two metadata
	// words and strict boundary comparison make pageWords-3 the real maximum.
	if words > set.pageWords-3 {
		return 0, ErrLimit
	}
	if countEntry && set.entries == set.limits.MaxEntries {
		return 0, ErrLimit
	}
	targetPage := set.page
	targetLimit := set.pages[targetPage].limit
	resetTarget := false
	if targetLimit < 2 || targetLimit >= set.pageWords {
		return 0, ErrInvariant
	}
	if targetLimit+words >= set.pageWords {
		targetPage++
		targetLimit = 2
		resetTarget = true
		if targetPage == len(set.pages) {
			if err := set.allocatePage(); err != nil {
				return 0, err
			}
		}
	}
	if set.limits.Stop != nil && set.limits.Stop() {
		return 0, ErrCancelled
	}
	handle := targetPage<<uint(set.pageBits) + targetLimit
	clear(set.pages[targetPage].words[targetLimit : targetLimit+words])
	if resetTarget {
		set.pages[targetPage].shadowLimit = -1
	}
	set.pages[targetPage].limit = targetLimit + words
	set.page = targetPage
	if countEntry {
		set.entries++
	}
	return handle, nil
}

func (set *PagedWordSet) AppendInts(values []int32) (int, error) {
	words, _ := WordCountForInts(len(values))
	handle, err := set.reservePrimaryWords(words, true)
	if err != nil {
		return 0, err
	}
	page, shift := handle>>uint(set.pageBits), handle&set.pageMask
	for index, value := range values {
		word := &set.pages[page].words[shift+index/2]
		if index&1 == 0 {
			*word = uint64(uint32(value))
		} else {
			*word |= uint64(uint32(value)) << 32
		}
	}
	return handle, nil
}

func (set *PagedWordSet) AppendZeroInts(count int) (int, error) {
	words, err := WordCountForInts(count)
	if err != nil {
		return 0, err
	}
	return set.reservePrimaryWords(words, true)
}

func (set *PagedWordSet) FetchBytes(count int) (int, error) {
	if count < 0 || count > int(^uint(0)>>1)-3 {
		return 0, ErrDimension
	}
	return set.AppendZeroInts((count + 3) >> 2)
}

func (set *PagedWordSet) SaveString(value string) (int, error) {
	length := len(value)
	for index := 0; index < len(value); index++ {
		if value[index] == 0 {
			length = index
			break
		}
	}
	if length == int(^uint(0)>>1) {
		return 0, ErrLimit
	}
	handle, err := set.FetchBytes(length + 1)
	if err != nil {
		return 0, err
	}
	if err = set.WriteBytes(handle, append([]byte(value[:length]), 0)); err != nil {
		return 0, err
	}
	return handle, nil
}

func (set *PagedWordSet) validWordRange(handle, words int) (int, int, error) {
	if err := set.check(); err != nil {
		return 0, 0, err
	}
	if handle < 0 || words < 0 {
		return 0, 0, ErrDimension
	}
	page := handle >> uint(set.pageBits)
	shift := handle & set.pageMask
	if page < 0 || page >= len(set.pages) || shift < 2 || words > set.pageWords-shift {
		return 0, 0, ErrDimension
	}
	pageLimit := set.pages[page].limit
	if set.pages[page].shadowLimit > pageLimit {
		pageLimit = set.pages[page].shadowLimit
	}
	if shift+words > pageLimit {
		return 0, 0, ErrDimension
	}
	return page, shift, nil
}

func (set *PagedWordSet) SnapshotWords(handle, words int) ([]uint64, error) {
	page, shift, err := set.validWordRange(handle, words)
	if err != nil {
		return nil, err
	}
	return append([]uint64(nil), set.pages[page].words[shift:shift+words]...), nil
}

func (set *PagedWordSet) WriteWords(handle int, values []uint64) error {
	page, shift, err := set.validWordRange(handle, len(values))
	if err != nil {
		return err
	}
	copy(set.pages[page].words[shift:], values)
	return nil
}

func (set *PagedWordSet) CopyWords(destination, source, words int) error {
	values, err := set.SnapshotWords(source, words)
	if err != nil {
		return err
	}
	return set.WriteWords(destination, values)
}

func (set *PagedWordSet) SnapshotInts(handle, count int) ([]int32, error) {
	words, err := WordCountForInts(count)
	if err != nil {
		return nil, err
	}
	values, err := set.SnapshotWords(handle, words)
	if err != nil {
		return nil, err
	}
	result := make([]int32, count)
	for index := range result {
		value := values[index/2]
		if index&1 != 0 {
			value >>= 32
		}
		result[index] = int32(uint32(value))
	}
	return result, nil
}

func (set *PagedWordSet) WriteBytes(handle int, values []byte) error {
	words := len(values) / 8
	if len(values)%8 != 0 {
		words++
	}
	page, shift, err := set.validWordRange(handle, words)
	if err != nil {
		return err
	}
	clear(set.pages[page].words[shift : shift+words])
	for index := 0; index < words; index++ {
		start := 8 * index
		end := min(start+8, len(values))
		var bytes [8]byte
		copy(bytes[:], values[start:end])
		set.pages[page].words[shift+index] = binary.LittleEndian.Uint64(bytes[:])
	}
	return nil
}

func (set *PagedWordSet) SnapshotBytes(handle, count int) ([]byte, error) {
	if count < 0 {
		return nil, ErrDimension
	}
	words := count / 8
	if count%8 != 0 {
		words++
	}
	values, err := set.SnapshotWords(handle, words)
	if err != nil {
		return nil, err
	}
	result := make([]byte, 8*words)
	for index, value := range values {
		binary.LittleEndian.PutUint64(result[8*index:], value)
	}
	return result[:count], nil
}

func (set *PagedWordSet) ReserveShadowInts(count int) (int, error) {
	if err := set.check(); err != nil {
		return 0, err
	}
	words, err := WordCountForInts(count)
	if err != nil {
		return 0, err
	}
	if words > set.pageWords-3 {
		return 0, ErrLimit
	}
	targetPage := set.shadowPage
	limit := set.pages[targetPage].shadowLimit
	if limit < 2 {
		return 0, ErrInvariant
	}
	if limit+words >= set.pageWords {
		targetPage++
		if targetPage >= len(set.pages) {
			return 0, ErrLimit
		}
		limit = 2
	}
	if set.limits.Stop != nil && set.limits.Stop() {
		return 0, ErrCancelled
	}
	handle := targetPage<<uint(set.pageBits) + limit
	set.pages[targetPage].shadowLimit = limit + words
	set.shadowPage = targetPage
	return handle, nil
}

func (set *PagedWordSet) Shrink(handle int) error {
	if err := set.check(); err != nil {
		return err
	}
	current, _ := set.CurrentHandle()
	if handle < 0 || handle > current {
		return ErrDimension
	}
	page, shift := handle>>uint(set.pageBits), handle&set.pageMask
	if page >= len(set.pages) || shift < 2 || shift >= set.pageWords {
		return ErrDimension
	}
	set.page = page
	set.pages[page].limit = shift
	return nil
}

func (set *PagedWordSet) ShrinkShadow(handle int) error {
	if err := set.check(); err != nil {
		return err
	}
	current, _ := set.CurrentHandle()
	if handle < 0 || handle > current {
		return ErrDimension
	}
	page, shift := handle>>uint(set.pageBits), handle&set.pageMask
	if page >= len(set.pages) || shift < 2 || shift >= set.pageWords {
		return ErrDimension
	}
	set.shadowPage = page
	set.pages[page].shadowLimit = shift
	return nil
}

func (set *PagedWordSet) CommitShadowLimits() error {
	if err := set.check(); err != nil {
		return err
	}
	for page := 0; page <= set.page; page++ {
		limit := set.pages[page].shadowLimit
		if limit < 2 || limit >= set.pageWords {
			return ErrInvariant
		}
	}
	for page := 0; page <= set.page; page++ {
		set.pages[page].limit = set.pages[page].shadowLimit
	}
	return nil
}

func (set *PagedWordSet) Restart() error {
	if err := set.check(); err != nil {
		return err
	}
	set.entries, set.page, set.shadowPage = 0, 0, 0
	set.pages[0].limit, set.pages[0].shadowLimit = 2, -1
	return nil
}

func (set *PagedWordSet) HandleMemoryBytes(handle int) (int, error) {
	page, err := set.HandlePage(handle)
	if err != nil {
		return 0, err
	}
	shift, _ := set.HandleShift(handle)
	if page > (int(^uint(0)>>1)/8-shift)/set.pageWords {
		return 0, ErrLimit
	}
	return (page*set.pageWords + shift) * 8, nil
}

func (set *PagedWordSet) MemoryBytes() (int, error) {
	handle, err := set.CurrentHandle()
	if err != nil {
		return 0, err
	}
	return set.HandleMemoryBytes(handle)
}

func (set *PagedWordSet) ShadowMemoryBytes() (int, error) {
	handle, err := set.ShadowHandle()
	if err != nil {
		return 0, err
	}
	return set.HandleMemoryBytes(handle)
}

func (set *PagedWordSet) AllocatedPageBytes() (int, error) {
	if err := set.check(); err != nil {
		return 0, err
	}
	return (set.page + 1) * set.pageWords * 8, nil
}

// SourceReportedBytes reproduces Vec_ReportMemory for ABC's 64-bit ABI.
func (set *PagedWordSet) SourceReportedBytes() (int, error) {
	allocated, err := set.AllocatedPageBytes()
	if err != nil {
		return 0, err
	}
	return 32 + set.pageCapacity*8 + allocated, nil
}

func (set *PagedWordSet) PrimaryHandles(size func(handle int) (int, error), maxRecords int) ([]int, error) {
	if err := set.check(); err != nil {
		return nil, err
	}
	if size == nil || maxRecords < 0 {
		return nil, ErrDimension
	}
	if maxRecords == 0 {
		maxRecords = set.limits.MaxEntries
	}
	result := make([]int, 0, min(set.entries, maxRecords))
	for page := 0; page <= set.page; page++ {
		for shift := 2; shift < set.pages[page].limit; {
			if len(result) == maxRecords {
				return nil, ErrLimit
			}
			handle := page<<uint(set.pageBits) + shift
			words, err := size(handle)
			if err != nil {
				return nil, err
			}
			if words <= 0 || words > set.pages[page].limit-shift {
				return nil, ErrInvariant
			}
			result = append(result, handle)
			shift += words
		}
	}
	return result, nil
}

func (set *PagedWordSet) Validate() error {
	if err := set.check(); err != nil {
		return err
	}
	if len(set.pages) > set.limits.MaxPages || len(set.pages)*set.pageWords > set.limits.MaxBytes/8 || set.pageCapacity < len(set.pages) {
		return ErrInvariant
	}
	for page, value := range set.pages {
		if len(value.words) != set.pageWords || value.limit < 2 || value.limit >= set.pageWords {
			return ErrInvariant
		}
		if page <= set.shadowPage && value.shadowLimit != -1 && (value.shadowLimit < 2 || value.shadowLimit >= set.pageWords) {
			return ErrInvariant
		}
	}
	return nil
}

func (set *PagedWordSet) Close() error {
	if err := set.check(); err != nil {
		return err
	}
	set.pages = nil
	set.closed = true
	return nil
}
