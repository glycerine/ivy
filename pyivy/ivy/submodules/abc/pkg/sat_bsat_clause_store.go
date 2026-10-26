package abc

// Port of src/sat/bsat/satClause.h using checked handles and owned clauses.

import (
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"
)

const (
	SatLearntMaxStartDefault = 10000
	SatLearntMaxIncreDefault = 1000
	SatLearntMaxRatioDefault = 50
)

type SatClauseHandle uint32

type SatClauseStoreLimits struct {
	MaxPages, MaxClauses, MaxLiterals int // zero: 65536, 16M, 64M
}

type SatStoredClause struct {
	Handle        SatClauseHandle
	Learned, Mark bool
	PartA         bool
	LBD           uint8
	Literals      []SatLit
	ID            int
	HasID         bool
	StorageWords  int
}

// String is the non-stdout replacement for clause_print_. It preserves the
// native signed zero-based variable rendering and trailing newline.
func (c SatStoredClause) String() string {
	var result strings.Builder
	result.WriteString("{ ")
	for _, literal := range c.Literals {
		variable := int64(literal >> 1)
		if literal&1 != 0 {
			variable = -variable
		}
		result.WriteString(strconv.FormatInt(variable, 10))
		result.WriteByte(' ')
	}
	result.WriteString("}\n")
	return result.String()
}

type SatClauseStoreStats struct {
	PageSize, PageWords, AllocatedPages        int
	ProblemEntries, LearnedEntries             int
	ProblemStored, LearnedStored, LiteralCount int
	ProblemCurrent, LearnedCurrent             SatClauseHandle
	MemoryUsed, MemoryAllocated                uint64
}

type satClauseRecord struct {
	handle        SatClauseHandle
	learned, mark bool
	partA         bool
	lbd           uint8
	literals      []SatLit
	id            int
	hasID         bool
	words         int
}

type satClauseBookmark struct {
	set     bool
	entries [2]int
	handles [2]SatClauseHandle
}

type SatClauseStore struct {
	pageSize, pageWords   int
	pageMask, learnedMask uint32
	limits                SatClauseStoreLimits
	currentPage           [2]int
	pageLimit             map[int]int // lookup-only; traversal order comes from records
	entries               [2]int
	records               [2][]*satClauseRecord
	byHandle              map[SatClauseHandle]*satClauseRecord // lookup-only
	bookmark              satClauseBookmark
	literalCount          int
	stopped               bool
}

func checkSatClauseStoreLimits(limits SatClauseStoreLimits) (SatClauseStoreLimits, error) {
	if limits.MaxPages < 0 || limits.MaxClauses < 0 || limits.MaxLiterals < 0 {
		return limits, ErrDimension
	}
	if limits.MaxPages == 0 {
		limits.MaxPages = 1 << 16
	}
	if limits.MaxClauses == 0 {
		limits.MaxClauses = 16 << 20
	}
	if limits.MaxLiterals == 0 {
		limits.MaxLiterals = 64 << 20
	}
	return limits, nil
}

func NewSatClauseStore(pageSize int, limits SatClauseStoreLimits) (*SatClauseStore, error) {
	if pageSize <= 8 || pageSize >= 32 {
		return nil, ErrDimension
	}
	limits, err := checkSatClauseStoreLimits(limits)
	if err != nil {
		return nil, err
	}
	s := &SatClauseStore{
		pageSize: pageSize, pageWords: 1 << uint(pageSize),
		pageMask: uint32((uint64(1) << uint(pageSize)) - 1), learnedMask: uint32(uint64(1) << uint(pageSize)),
		limits: limits, pageLimit: map[int]int{0: 2, 1: 2}, byHandle: make(map[SatClauseHandle]*satClauseRecord),
	}
	s.currentPage = [2]int{0, 1}
	return s, nil
}

func SatClauseHandleFromLiteral(literal SatLit) (SatClauseHandle, error) {
	if !literal.Valid() || uint64(literal) > uint64(math.MaxUint32>>1) {
		return 0, ErrInvalidLiteral
	}
	return SatClauseHandle(uint32(literal)<<1 | 1), nil
}

func (h SatClauseHandle) IsLiteral() bool { return h&1 != 0 }

func (h SatClauseHandle) Literal() (SatLit, error) {
	if !h.IsLiteral() {
		return 0, ErrInvalidLiteral
	}
	literal := SatLit(uint32(h) >> 1)
	if !literal.Valid() {
		return 0, ErrInvalidLiteral
	}
	return literal, nil
}

func (s *SatClauseStore) kind(learned bool) int {
	if learned {
		return 1
	}
	return 0
}

func (s *SatClauseStore) clauseWords(size int, learned, plusOne bool) (int, error) {
	if size < 0 {
		return 0, ErrDimension
	}
	extra := 0
	if learned || plusOne {
		extra = 1
	}
	if size > math.MaxInt-2-extra {
		return 0, ErrLimit
	}
	words := 2 * ((size + 2 + extra) / 2)
	if words+3 >= s.pageWords {
		return 0, ErrLimit
	}
	return words, nil
}

func (s *SatClauseStore) makeHandle(page, offset int) (SatClauseHandle, error) {
	if page < 0 || offset < 0 || offset >= s.pageWords || page >= s.limits.MaxPages {
		return 0, ErrLimit
	}
	value := uint64(page)<<uint(s.pageSize) | uint64(offset)
	if value > math.MaxUint32 || value&1 != 0 {
		return 0, ErrLimit
	}
	return SatClauseHandle(value), nil
}

func (s *SatClauseStore) appendClause(literals []SatLit, blankSize int, learned, plusOne bool) (SatClauseHandle, error) {
	if s == nil || s.stopped {
		return 0, ErrInvariant
	}
	size := len(literals)
	if blankSize >= 0 {
		size = blankSize
	}
	words, err := s.clauseWords(size, learned, plusOne)
	if err != nil {
		return 0, err
	}
	if s.entries[0]+s.entries[1] >= s.limits.MaxClauses || size > s.limits.MaxLiterals-s.literalCount {
		return 0, ErrLimit
	}
	if blankSize < 0 {
		for _, literal := range literals {
			if !literal.Valid() {
				return 0, ErrInvalidLiteral
			}
		}
	}
	kind := s.kind(learned)
	page := s.currentPage[kind]
	limit := s.pageLimit[page]
	if limit+words+2 >= s.pageWords {
		page += 2
		if page >= s.limits.MaxPages {
			return 0, ErrLimit
		}
		// Native rewinds an already allocated page before reusing it after a
		// rollback or shrink; stale contents never affect the new limit.
		s.pageLimit[page] = 2
		limit = 2
	}
	handle, err := s.makeHandle(page, limit)
	if err != nil {
		return 0, err
	}
	owned := make([]SatLit, size)
	copy(owned, literals)
	record := &satClauseRecord{handle: handle, learned: learned, literals: owned, words: words, id: -1}
	if learned || plusOne {
		record.id, record.hasID = s.entries[kind], true
	}
	s.currentPage[kind] = page
	s.pageLimit[page] = limit + words
	s.entries[kind]++
	s.records[kind] = append(s.records[kind], record)
	s.byHandle[handle] = record
	s.literalCount += size
	return handle, nil
}

func (s *SatClauseStore) Append(literals []SatLit, learned, plusOne bool) (SatClauseHandle, error) {
	return s.appendClause(literals, -1, learned, plusOne)
}

func (s *SatClauseStore) AppendBlank(size int, learned, plusOne bool) (SatClauseHandle, error) {
	return s.appendClause(nil, size, learned, plusOne)
}

func (s *SatClauseStore) pageAndOffset(handle SatClauseHandle) (int, int) {
	return int(uint32(handle) >> uint(s.pageSize)), int(uint32(handle) & s.pageMask)
}

func (s *SatClauseStore) HandlePage(handle SatClauseHandle) (int, error) {
	if s == nil || s.stopped || handle.IsLiteral() {
		return 0, ErrInvalidLiteral
	}
	page, _ := s.pageAndOffset(handle)
	return page, nil
}

func (s *SatClauseStore) HandleOffset(handle SatClauseHandle) (int, error) {
	if s == nil || s.stopped || handle.IsLiteral() {
		return 0, ErrInvalidLiteral
	}
	_, offset := s.pageAndOffset(handle)
	return offset, nil
}

func (s *SatClauseStore) StorageWords(size int, learned bool) (int, error) {
	if s == nil || s.stopped {
		return 0, ErrInvariant
	}
	return s.clauseWords(size, learned, false)
}

func (s *SatClauseStore) CurrentHandle(learned bool) SatClauseHandle {
	if s == nil || s.stopped {
		return 0
	}
	kind := s.kind(learned)
	handle, _ := s.makeHandle(s.currentPage[kind], s.pageLimit[s.currentPage[kind]])
	return handle
}

func (s *SatClauseStore) HandleLearned(handle SatClauseHandle) bool {
	return s != nil && uint32(handle)&s.learnedMask != 0
}

func (s *SatClauseStore) snapshot(record *satClauseRecord) SatStoredClause {
	return SatStoredClause{record.handle, record.learned, record.mark, record.partA, record.lbd, slices.Clone(record.literals), record.id, record.hasID, record.words}
}

func (s *SatClauseStore) Clause(handle SatClauseHandle) (SatStoredClause, error) {
	if s == nil || s.stopped || handle == 0 || handle.IsLiteral() {
		return SatStoredClause{}, ErrInvalidLiteral
	}
	record := s.byHandle[handle]
	if record == nil {
		return SatStoredClause{}, ErrInvalidLiteral
	}
	return s.snapshot(record), nil
}

func (s *SatClauseStore) SetMetadata(handle SatClauseHandle, mark, partA bool, lbd uint8) error {
	if s == nil || s.stopped {
		return ErrInvariant
	}
	record := s.byHandle[handle]
	if record == nil {
		return ErrInvalidLiteral
	}
	record.mark, record.partA, record.lbd = mark, partA, lbd
	return nil
}

func (s *SatClauseStore) SetClauseID(handle SatClauseHandle, id int) error {
	if s == nil || s.stopped {
		return ErrInvariant
	}
	record := s.byHandle[handle]
	if record == nil {
		return ErrInvalidLiteral
	}
	if id < 0 {
		return ErrDimension
	}
	record.id, record.hasID = id, true
	return nil
}

func (s *SatClauseStore) EntryCount(learned bool) int {
	if s == nil {
		return 0
	}
	return s.entries[s.kind(learned)]
}

func (s *SatClauseStore) LearnedCount() int {
	if s == nil || s.stopped {
		return 0
	}
	return len(s.records[1])
}

func (s *SatClauseStore) ProblemClauses() []SatStoredClause {
	if s == nil || s.stopped || len(s.records[0]) <= 1 {
		return nil
	}
	result := make([]SatStoredClause, 0, len(s.records[0])-1)
	for _, record := range s.records[0][1:] {
		result = append(result, s.snapshot(record))
	}
	return result
}

func (s *SatClauseStore) LearnedClauses() []SatStoredClause {
	if s == nil || s.stopped {
		return nil
	}
	result := make([]SatStoredClause, len(s.records[1]))
	for i, record := range s.records[1] {
		result[i] = s.snapshot(record)
	}
	return result
}

func (s *SatClauseStore) Bookmark() error {
	if s == nil || s.stopped {
		return ErrInvariant
	}
	s.bookmark = satClauseBookmark{set: true, entries: s.entries, handles: [2]SatClauseHandle{s.CurrentHandle(false), s.CurrentHandle(true)}}
	return nil
}

func (s *SatClauseStore) ClauseUsed(handle SatClauseHandle) (bool, error) {
	if s == nil || s.stopped || !s.bookmark.set || s.byHandle[handle] == nil {
		return false, ErrInvariant
	}
	kind := 0
	if s.HandleLearned(handle) {
		kind = 1
	}
	return handle < s.bookmark.handles[kind], nil
}

func (s *SatClauseStore) rebuildLookup() {
	s.byHandle = make(map[SatClauseHandle]*satClauseRecord, len(s.records[0])+len(s.records[1]))
	s.literalCount = 0
	for kind := range s.records {
		for _, record := range s.records[kind] {
			s.byHandle[record.handle] = record
			s.literalCount += len(record.literals)
		}
	}
}

func (s *SatClauseStore) Rollback() error {
	if s == nil || s.stopped || !s.bookmark.set {
		return ErrInvariant
	}
	for kind := 0; kind < 2; kind++ {
		boundary := s.bookmark.handles[kind]
		kept := s.records[kind][:0]
		for _, record := range s.records[kind] {
			if record.handle < boundary {
				kept = append(kept, record)
			}
		}
		s.records[kind] = kept
		s.entries[kind] = s.bookmark.entries[kind]
		page, offset := s.pageAndOffset(boundary)
		s.currentPage[kind], s.pageLimit[page] = page, offset
	}
	s.rebuildLookup()
	return nil
}

func (s *SatClauseStore) Restart() error {
	if s == nil || s.stopped {
		return ErrInvariant
	}
	s.entries, s.currentPage = [2]int{}, [2]int{0, 1}
	s.records = [2][]*satClauseRecord{}
	s.pageLimit = map[int]int{0: 2, 1: 2}
	s.byHandle = make(map[SatClauseHandle]*satClauseRecord)
	s.bookmark, s.literalCount = satClauseBookmark{}, 0
	return nil
}

func (s *SatClauseStore) Shrink(handle SatClauseHandle, learned bool) error {
	if s == nil || s.stopped || handle == 0 || handle.IsLiteral() || s.HandleLearned(handle) != learned {
		return ErrInvariant
	}
	kind := s.kind(learned)
	if handle > s.CurrentHandle(learned) {
		return ErrInvalidLiteral
	}
	page, offset := s.pageAndOffset(handle)
	kept := s.records[kind][:0]
	for _, record := range s.records[kind] {
		if record.handle < handle {
			kept = append(kept, record)
		}
	}
	s.records[kind] = kept
	s.currentPage[kind], s.pageLimit[page] = page, offset
	s.rebuildLookup()
	return nil
}

func (s *SatClauseStore) compactPosition(index int) (SatClauseHandle, error) {
	page, offset := 1, 2
	for i, record := range s.records[1] {
		if record.mark {
			continue
		}
		if i == index {
			return s.makeHandle(page, offset)
		}
		if offset+record.words >= s.pageWords {
			page, offset = page+2, 2
		}
		offset += record.words
	}
	return s.makeHandle(page, offset)
}

// CompactLearned(false) performs the native dry run and writes each retained
// clause's future handle into its ID. CompactLearned(true) requires that plan,
// moves clauses, rewrites IDs densely and repairs the learned bookmark.
func (s *SatClauseStore) CompactLearned(move bool) (int, error) {
	if s == nil || s.stopped {
		return 0, ErrInvariant
	}
	page, offset, count := 1, 2, 0
	oldBoundary := s.bookmark.handles[1]
	boundarySet := !s.bookmark.set || oldBoundary == 0
	newBoundary := SatClauseHandle(0)
	newBoundaryEntries := 0
	for _, record := range s.records[1] {
		if record.mark {
			continue
		}
		if offset+record.words >= s.pageWords {
			page, offset = page+2, 2
		}
		newHandle, err := s.makeHandle(page, offset)
		if err != nil {
			return 0, err
		}
		if s.bookmark.set && !boundarySet && record.handle >= oldBoundary {
			newBoundary, newBoundaryEntries, boundarySet = newHandle, count, true
		}
		if move {
			if !record.hasID || record.id != int(newHandle) {
				return 0, fmt.Errorf("%w: learned compaction was not planned", ErrInvariant)
			}
		} else {
			record.id, record.hasID = int(newHandle), true
		}
		offset += record.words
		count++
	}
	if !move {
		return count, nil
	}
	kept := make([]*satClauseRecord, 0, count)
	page, offset, count = 1, 2, 0
	for _, record := range s.records[1] {
		if record.mark {
			continue
		}
		if offset+record.words >= s.pageWords {
			page, offset = page+2, 2
		}
		record.handle, record.id = SatClauseHandle(uint32(page)<<uint(s.pageSize)|uint32(offset)), count
		record.hasID = true
		kept = append(kept, record)
		offset += record.words
		count++
	}
	s.records[1], s.entries[1], s.currentPage[1], s.pageLimit[page] = kept, count, page, offset
	if s.bookmark.set {
		if !boundarySet {
			newBoundary, newBoundaryEntries = s.CurrentHandle(true), count
		}
		s.bookmark.handles[1], s.bookmark.entries[1] = newBoundary, newBoundaryEntries
	}
	s.rebuildLookup()
	return count, nil
}

func (s *SatClauseStore) memoryAt(handle SatClauseHandle) uint64 {
	page, offset := s.pageAndOffset(handle)
	return uint64((page+2)/2*s.pageWords+offset) * 4
}

func (s *SatClauseStore) MemoryAt(handle SatClauseHandle) (uint64, error) {
	if s == nil || s.stopped || handle.IsLiteral() {
		return 0, ErrInvalidLiteral
	}
	page, offset := s.pageAndOffset(handle)
	limit, exists := s.pageLimit[page]
	if !exists || page > s.currentPage[page&1] || offset > limit {
		return 0, ErrInvalidLiteral
	}
	return s.memoryAt(handle), nil
}

func (s *SatClauseStore) Stats() SatClauseStoreStats {
	if s == nil || s.stopped {
		return SatClauseStoreStats{}
	}
	problem, learned := s.CurrentHandle(false), s.CurrentHandle(true)
	allocatedPages := s.currentPage[0]/2 + s.currentPage[1]/2 + 2
	return SatClauseStoreStats{
		PageSize: s.pageSize, PageWords: s.pageWords, AllocatedPages: allocatedPages,
		ProblemEntries: s.entries[0], LearnedEntries: s.entries[1], ProblemStored: len(s.records[0]), LearnedStored: len(s.records[1]), LiteralCount: s.literalCount,
		ProblemCurrent: problem, LearnedCurrent: learned,
		MemoryUsed: s.memoryAt(problem) + s.memoryAt(learned), MemoryAllocated: uint64(s.currentPage[0]+s.currentPage[1]+2) * uint64(s.pageWords) * 4,
	}
}

func (s *SatClauseStore) Stop() {
	if s == nil || s.stopped {
		return
	}
	s.records, s.pageLimit, s.byHandle = [2][]*satClauseRecord{}, nil, nil
	s.literalCount, s.stopped = 0, true
}
