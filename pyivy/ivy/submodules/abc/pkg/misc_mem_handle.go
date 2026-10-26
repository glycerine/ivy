package abc

// Checked page/handle memory managers from src/misc/mem/mem2.h.

import "math"

type MmrLimits struct {
	MaxWords, MaxEntries int
	Stop                 func() bool
}

func normalizeMmrLimits(limits MmrLimits) (MmrLimits, error) {
	if limits.MaxWords < 0 || limits.MaxEntries < 0 {
		return limits, ErrDimension
	}
	if limits.MaxWords == 0 {
		limits.MaxWords = 1 << 26
	}
	if limits.MaxEntries == 0 {
		limits.MaxEntries = 1 << 24
	}
	return limits, nil
}

func mmrStopped(limits MmrLimits) bool { return limits.Stop != nil && limits.Stop() }

type MmrFlexStats struct {
	PageBase, PageWords, Pages, LivePages int
	Entries, PeakEntries, NextOffset      int
}

type MmrFlexArena struct {
	pageBase, pageMask, next int
	pages                    [][]uint64
	lengths                  map[int]int
	entries, peak, words     int
	limits                   MmrLimits
	closed                   bool
}

func NewMmrFlexArena(pageBase int, limits MmrLimits) (*MmrFlexArena, error) {
	limits, err := normalizeMmrLimits(limits)
	if err != nil || pageBase < 2 || pageBase > 30 {
		if err != nil {
			return nil, err
		}
		return nil, ErrDimension
	}
	if mmrStopped(limits) {
		return nil, ErrCancelled
	}
	pageWords := 1 << uint(pageBase)
	if pageWords > limits.MaxWords {
		return nil, ErrLimit
	}
	return &MmrFlexArena{pageBase: pageBase, pageMask: pageWords - 1, next: pageWords, lengths: make(map[int]int), limits: limits}, nil
}

func (a *MmrFlexArena) check() error {
	if a == nil {
		return ErrDimension
	}
	if a.closed || a.pageBase < 2 || a.pageMask != 1<<uint(a.pageBase)-1 || a.lengths == nil {
		return ErrInvariant
	}
	return nil
}

func (a *MmrFlexArena) Fetch(words int) (int, error) {
	if err := a.check(); err != nil {
		return 0, err
	}
	if words <= 0 || words >= a.pageMask {
		return 0, ErrDimension
	}
	if mmrStopped(a.limits) {
		return 0, ErrCancelled
	}
	if a.entries == a.limits.MaxEntries {
		return 0, ErrLimit
	}
	if a.next+words >= a.pageMask {
		pageWords := a.pageMask + 1
		if pageWords > a.limits.MaxWords-a.words {
			return 0, ErrLimit
		}
		page := make([]uint64, pageWords)
		for index := range page {
			page[index] = ^uint64(0)
		}
		a.pages = append(a.pages, page)
		a.words += pageWords
		a.next = 1
	}
	handle := (len(a.pages)-1)<<uint(a.pageBase) | a.next
	a.next += words
	a.lengths[handle] = words
	a.entries++
	if a.peak < a.entries {
		a.peak = a.entries
	}
	return handle, nil
}

func (a *MmrFlexArena) block(handle int) ([]uint64, error) {
	if err := a.check(); err != nil {
		return nil, err
	}
	length, found := a.lengths[handle]
	if !found || handle <= 0 {
		return nil, ErrDimension
	}
	pageIndex, offset := handle>>uint(a.pageBase), handle&a.pageMask
	if pageIndex < 0 || pageIndex >= len(a.pages) || a.pages[pageIndex] == nil || offset+length > len(a.pages[pageIndex]) {
		return nil, ErrInvariant
	}
	return a.pages[pageIndex][offset : offset+length], nil
}

func (a *MmrFlexArena) Read(handle, index int) (uint64, error) {
	block, err := a.block(handle)
	if err != nil {
		return 0, err
	}
	if index < 0 || index >= len(block) {
		return 0, ErrDimension
	}
	return block[index], nil
}

func (a *MmrFlexArena) Write(handle, index int, value uint64) error {
	block, err := a.block(handle)
	if err != nil {
		return err
	}
	if index < 0 || index >= len(block) {
		return ErrDimension
	}
	block[index] = value
	return nil
}

func (a *MmrFlexArena) Snapshot(handle int) ([]uint64, error) {
	block, err := a.block(handle)
	if err != nil {
		return nil, err
	}
	return append([]uint64(nil), block...), nil
}

// Release preserves Mmr_FlexRelease's unusual behavior: an entry on page N
// releases page N-1, not the entry itself.
func (a *MmrFlexArena) Release(handle int) (bool, error) {
	if _, err := a.block(handle); err != nil {
		return false, err
	}
	pageIndex := handle >> uint(a.pageBase)
	if pageIndex == 0 || a.pages[pageIndex-1] == nil {
		return false, nil
	}
	for existing := range a.lengths {
		if existing>>uint(a.pageBase) == pageIndex-1 {
			delete(a.lengths, existing)
		}
	}
	a.words -= len(a.pages[pageIndex-1])
	a.pages[pageIndex-1] = nil
	return true, nil
}

func (a *MmrFlexArena) Stats() (MmrFlexStats, error) {
	if err := a.check(); err != nil {
		return MmrFlexStats{}, err
	}
	live := 0
	for _, page := range a.pages {
		if page != nil {
			live++
		}
	}
	return MmrFlexStats{a.pageBase, a.pageMask + 1, len(a.pages), live, a.entries, a.peak, a.next}, nil
}

func (a *MmrFlexArena) MemoryWords() (int, error) {
	if err := a.check(); err != nil {
		return 0, err
	}
	return a.words, nil
}
func (a *MmrFlexArena) Close() error {
	if err := a.check(); err != nil {
		return err
	}
	a.pages, a.lengths, a.words, a.closed = nil, nil, 0, true
	return nil
}

type MmrFixedStats struct {
	PageBase, PageWords, EntryWords, Pages int
	Entries, PeakEntries, FreeEntries      int
}

type MmrFixedArena struct {
	pageBase, pageMask, entryWords int
	pages                          [][]uint64
	free                           []int
	live                           map[int]bool
	entries, peak, words           int
	limits                         MmrLimits
	closed                         bool
}

func NewMmrFixedArena(pageBase, entryWords int, limits MmrLimits) (*MmrFixedArena, error) {
	limits, err := normalizeMmrLimits(limits)
	if err != nil || pageBase < 2 || pageBase > 30 || entryWords <= 0 || int64(entryWords) >= (int64(1)<<uint(pageBase))-1 {
		if err != nil {
			return nil, err
		}
		return nil, ErrDimension
	}
	if mmrStopped(limits) {
		return nil, ErrCancelled
	}
	pageWords := 1 << uint(pageBase)
	if pageWords > limits.MaxWords {
		return nil, ErrLimit
	}
	return &MmrFixedArena{pageBase: pageBase, pageMask: pageWords - 1, entryWords: entryWords, live: make(map[int]bool), limits: limits}, nil
}

func (a *MmrFixedArena) check() error {
	if a == nil {
		return ErrDimension
	}
	if a.closed || a.pageBase < 2 || a.entryWords <= 0 || a.pageMask != 1<<uint(a.pageBase)-1 || a.live == nil {
		return ErrInvariant
	}
	return nil
}

func (a *MmrFixedArena) grow() error {
	pageWords := a.pageMask + 1
	if pageWords > a.limits.MaxWords-a.words {
		return ErrLimit
	}
	pageIndex := len(a.pages)
	page := make([]uint64, pageWords)
	for index := range page {
		page[index] = ^uint64(0)
	}
	a.pages = append(a.pages, page)
	a.words += pageWords
	last := 1 + (a.pageMask-a.entryWords-1)/a.entryWords*a.entryWords
	for offset := last; offset >= 1; offset -= a.entryWords {
		a.free = append(a.free, pageIndex<<uint(a.pageBase)|offset)
	}
	if len(a.free) == 0 {
		return ErrInvariant
	}
	return nil
}

func (a *MmrFixedArena) Fetch() (int, error) {
	if err := a.check(); err != nil {
		return 0, err
	}
	if mmrStopped(a.limits) {
		return 0, ErrCancelled
	}
	if a.entries == a.limits.MaxEntries {
		return 0, ErrLimit
	}
	if len(a.free) == 0 {
		if err := a.grow(); err != nil {
			return 0, err
		}
	}
	handle := a.free[len(a.free)-1]
	a.free = a.free[:len(a.free)-1]
	if a.live[handle] {
		return 0, ErrInvariant
	}
	a.live[handle] = true
	a.entries++
	if a.peak < a.entries {
		a.peak = a.entries
	}
	return handle, nil
}

func (a *MmrFixedArena) block(handle int) ([]uint64, error) {
	if err := a.check(); err != nil {
		return nil, err
	}
	if !a.live[handle] || handle <= 0 {
		return nil, ErrDimension
	}
	pageIndex, offset := handle>>uint(a.pageBase), handle&a.pageMask
	if pageIndex < 0 || pageIndex >= len(a.pages) || offset+a.entryWords > len(a.pages[pageIndex]) {
		return nil, ErrInvariant
	}
	return a.pages[pageIndex][offset : offset+a.entryWords], nil
}

func (a *MmrFixedArena) Read(handle, index int) (uint64, error) {
	block, err := a.block(handle)
	if err != nil {
		return 0, err
	}
	if index < 0 || index >= len(block) {
		return 0, ErrDimension
	}
	return block[index], nil
}
func (a *MmrFixedArena) Write(handle, index int, value uint64) error {
	block, err := a.block(handle)
	if err != nil {
		return err
	}
	if index < 0 || index >= len(block) {
		return ErrDimension
	}
	block[index] = value
	return nil
}
func (a *MmrFixedArena) Snapshot(handle int) ([]uint64, error) {
	block, err := a.block(handle)
	if err != nil {
		return nil, err
	}
	return append([]uint64(nil), block...), nil
}

func (a *MmrFixedArena) Recycle(handle int) error {
	block, err := a.block(handle)
	if err != nil {
		return err
	}
	for index := range block {
		block[index] = ^uint64(0)
	}
	delete(a.live, handle)
	a.free = append(a.free, handle)
	a.entries--
	return nil
}

func (a *MmrFixedArena) Stats() (MmrFixedStats, error) {
	if err := a.check(); err != nil {
		return MmrFixedStats{}, err
	}
	return MmrFixedStats{a.pageBase, a.pageMask + 1, a.entryWords, len(a.pages), a.entries, a.peak, len(a.free)}, nil
}
func (a *MmrFixedArena) MemoryWords() (int, error) {
	if err := a.check(); err != nil {
		return 0, err
	}
	return a.words, nil
}
func (a *MmrFixedArena) Close() error {
	if err := a.check(); err != nil {
		return err
	}
	a.pages, a.free, a.live, a.words, a.closed = nil, nil, nil, 0, true
	return nil
}

type MmrStepStats struct {
	WordBits, Mask, Entries, PeakEntries, TotalEntries, MemoryWords int
}

type MmrStepArena struct {
	wordBits, mask                 int
	memories                       []*MmrFixedArena
	entries, peak, total, maxWords int
	limits                         MmrLimits
	closed                         bool
}

func NewMmrStepArena(pageBase, wordBits int, limits MmrLimits) (*MmrStepArena, error) {
	limits, err := normalizeMmrLimits(limits)
	if err != nil || pageBase < 2 || pageBase > 30 || wordBits < 1 || wordBits > 12 {
		if err != nil {
			return nil, err
		}
		return nil, ErrDimension
	}
	if mmrStopped(limits) {
		return nil, ErrCancelled
	}
	mask := 1<<uint(wordBits) - 1
	a := &MmrStepArena{wordBits: wordBits, mask: mask, memories: make([]*MmrFixedArena, mask+1), limits: limits, maxWords: limits.MaxWords}
	for words := 1; words <= mask; words++ {
		memory, memoryErr := NewMmrFixedArena(pageBase, words, limits)
		if memoryErr != nil {
			a.Close()
			return nil, memoryErr
		}
		a.memories[words] = memory
	}
	return a, nil
}

func (a *MmrStepArena) check() error {
	if a == nil {
		return ErrDimension
	}
	if a.closed || a.wordBits < 1 || a.mask != 1<<uint(a.wordBits)-1 || len(a.memories) != a.mask+1 {
		return ErrInvariant
	}
	return nil
}

func (a *MmrStepArena) memoryWords() int {
	total := 0
	for words := 1; words <= a.mask; words++ {
		if a.memories[words] != nil {
			total += a.memories[words].words
		}
	}
	return total
}

func (a *MmrStepArena) Fetch(words int) (int, error) {
	if err := a.check(); err != nil {
		return 0, err
	}
	if words <= 0 || words > a.mask {
		return 0, ErrDimension
	}
	if mmrStopped(a.limits) {
		return 0, ErrCancelled
	}
	if a.entries == a.limits.MaxEntries {
		return 0, ErrLimit
	}
	memory := a.memories[words]
	if len(memory.free) == 0 && memory.pageMask+1 > a.maxWords-a.memoryWords() {
		return 0, ErrLimit
	}
	handle, err := memory.Fetch()
	if err != nil {
		return 0, err
	}
	if handle > math.MaxInt>>uint(a.wordBits) {
		_ = memory.Recycle(handle)
		return 0, ErrLimit
	}
	a.entries++
	a.total++
	if a.peak < a.entries {
		a.peak = a.entries
	}
	return handle<<uint(a.wordBits) | words, nil
}

func (a *MmrStepArena) decode(handle int) (*MmrFixedArena, int, error) {
	if err := a.check(); err != nil {
		return nil, 0, err
	}
	words := handle & a.mask
	inner := handle >> uint(a.wordBits)
	if handle <= 0 || words <= 0 || words > a.mask || inner <= 0 {
		return nil, 0, ErrDimension
	}
	return a.memories[words], inner, nil
}

func (a *MmrStepArena) Read(handle, index int) (uint64, error) {
	memory, inner, err := a.decode(handle)
	if err != nil {
		return 0, err
	}
	return memory.Read(inner, index)
}
func (a *MmrStepArena) Write(handle, index int, value uint64) error {
	memory, inner, err := a.decode(handle)
	if err != nil {
		return err
	}
	return memory.Write(inner, index, value)
}
func (a *MmrStepArena) Snapshot(handle int) ([]uint64, error) {
	memory, inner, err := a.decode(handle)
	if err != nil {
		return nil, err
	}
	return memory.Snapshot(inner)
}
func (a *MmrStepArena) Recycle(handle int) error {
	memory, inner, err := a.decode(handle)
	if err != nil {
		return err
	}
	if err = memory.Recycle(inner); err != nil {
		return err
	}
	a.entries--
	return nil
}
func (a *MmrStepArena) Stats() (MmrStepStats, error) {
	if err := a.check(); err != nil {
		return MmrStepStats{}, err
	}
	return MmrStepStats{a.wordBits, a.mask, a.entries, a.peak, a.total, a.memoryWords()}, nil
}
func (a *MmrStepArena) Close() error {
	if a == nil {
		return ErrDimension
	}
	if a.closed {
		return ErrInvariant
	}
	for _, memory := range a.memories {
		if memory != nil && !memory.closed {
			_ = memory.Close()
		}
	}
	a.memories, a.closed = nil, true
	return nil
}
