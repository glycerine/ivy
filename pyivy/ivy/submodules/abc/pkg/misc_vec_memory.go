package abc

// Paged fixed-width words from src/misc/vec/vecMem.h.

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
)

// WordMemoryOptions bounds logical records and allocated uint64 page payload.
// Each zero bound defaults to one million. A page must fit the word budget.
type WordMemoryOptions struct{ MaxEntries, MaxWords int }

// WordMemory owns fixed-width records and pages. Use NewWordMemory; its zero
// value is unconfigured. Read snapshots own their storage. While hashing is
// enabled, records are immutable and added through Intern; IDs follow insertion
// order. The lookup map is never iterated to assign IDs or produce output.
type WordMemory struct {
	pages                                [][]uint64
	entryWords, pageLog, pageWords, size int
	options                              WordMemoryOptions
	lookup                               map[string]int
}

func NewWordMemory(entryWords, pageLog int, options WordMemoryOptions) (*WordMemory, error) {
	if entryWords <= 0 || pageLog < 0 || pageLog > 30 || options.MaxEntries < 0 || options.MaxWords < 0 {
		return nil, ErrDimension
	}
	if options.MaxEntries == 0 {
		options.MaxEntries = 1 << 20
	}
	if options.MaxWords == 0 {
		options.MaxWords = 1 << 20
	}
	pageWords, e := checkedSize(0, entryWords, 1<<uint(pageLog), options.MaxWords)
	if e != nil {
		return nil, e
	}
	if _, e = checkedSize(0, 8, pageWords, int(^uint(0)>>1)); e != nil {
		return nil, e
	}
	return &WordMemory{entryWords: entryWords, pageLog: pageLog, pageWords: pageWords, options: options}, nil
}
func (m *WordMemory) configured() bool { return m != nil && m.entryWords > 0 }
func (m *WordMemory) Len() int {
	if m == nil {
		return 0
	}
	return m.size
}
func (m *WordMemory) EntryWords() int {
	if m == nil {
		return 0
	}
	return m.entryWords
}
func (m *WordMemory) PageLog() int {
	if m == nil {
		return 0
	}
	return m.pageLog
}
func (m *WordMemory) PageCount() int {
	if m == nil {
		return 0
	}
	return len(m.pages)
}
func (m *WordMemory) StorageBytes() int {
	if m == nil {
		return 0
	}
	return len(m.pages) * m.pageWords * 8
}
func (m *WordMemory) entry(index int) []uint64 {
	page := index >> uint(m.pageLog)
	offset := (index & ((1 << uint(m.pageLog)) - 1)) * m.entryWords
	return m.pages[page][offset : offset+m.entryWords]
}
func (m *WordMemory) extend(size int) error {
	if size <= m.size {
		return nil
	}
	if size > m.options.MaxEntries {
		return ErrLimit
	}
	entriesPerPage := 1 << uint(m.pageLog)
	count := (size-1)/entriesPerPage + 1
	words, e := checkedSize(0, m.pageWords, count, m.options.MaxWords)
	if e != nil {
		return e
	}
	if _, e = checkedSize(0, 8, words, int(^uint(0)>>1)); e != nil {
		return e
	}
	if count > len(m.pages) {
		pages := make([][]uint64, count)
		copy(pages, m.pages)
		for i := len(m.pages); i < count; i++ {
			pages[i] = make([]uint64, m.pageWords)
		}
		m.pages = pages
	}
	for i := m.size; i < size; i++ {
		clear(m.entry(i))
	}
	m.size = size
	return nil
}
func (m *WordMemory) Read(index int) ([]uint64, error) {
	if !m.configured() || index < 0 || index >= m.size {
		return nil, ErrDimension
	}
	return append([]uint64(nil), m.entry(index)...), nil
}
func (m *WordMemory) Last() ([]uint64, error) {
	if !m.configured() {
		return nil, ErrDimension
	}
	return m.Read(m.size - 1)
}

// Get extends with zero records, correcting C's uninitialized gaps. Growth while
// hashing is active rejects; reads of existing interned records remain allowed.
func (m *WordMemory) Get(index int) ([]uint64, error) {
	if !m.configured() || index < 0 {
		return nil, ErrDimension
	}
	size, e := checkedSize(index, 1, 1, m.options.MaxEntries)
	if e != nil {
		return nil, e
	}
	if size > m.size {
		if m.lookup != nil {
			return nil, ErrInvariant
		}
		if e = m.extend(size); e != nil {
			return nil, e
		}
	}
	return m.Read(index)
}
func (m *WordMemory) Write(index int, words []uint64) error {
	if !m.configured() || index < 0 || index >= m.size || len(words) != m.entryWords {
		return ErrDimension
	}
	if m.lookup != nil {
		return ErrInvariant
	}
	copy(m.entry(index), words)
	return nil
}
func (m *WordMemory) Set(index int, words []uint64) error {
	if !m.configured() || index < 0 || len(words) != m.entryWords {
		return ErrDimension
	}
	if m.lookup != nil {
		return ErrInvariant
	}
	size, e := checkedSize(index, 1, 1, m.options.MaxEntries)
	if e != nil {
		return e
	}
	if e = m.extend(size); e != nil {
		return e
	}
	copy(m.entry(index), words)
	return nil
}
func (m *WordMemory) Push(words []uint64) error {
	if !m.configured() {
		return ErrDimension
	}
	return m.Set(m.size, words)
}

// Shrink frees unused pages, including every page at size zero, and removes
// truncated interned keys. C's exact-page/empty shrink bookkeeping is corrected.
func (m *WordMemory) Shrink(size int) error {
	if !m.configured() || size < 0 || size > m.size {
		return ErrDimension
	}
	if m.lookup != nil {
		for i := size; i < m.size; i++ {
			delete(m.lookup, wordMemoryKey(m.entry(i)))
		}
	}
	pages := 0
	if size > 0 {
		pages = (size-1)/(1<<uint(m.pageLog)) + 1
	}
	if pages > 0 {
		used := (size-1)%(1<<uint(m.pageLog)) + 1
		clear(m.pages[pages-1][used*m.entryWords:])
	}
	clear(m.pages[pages:])
	m.pages = m.pages[:pages]
	m.size = size
	return nil
}
func wordMemoryKey(words []uint64) string {
	data := make([]byte, len(words)*8)
	for i, w := range words {
		binary.LittleEndian.PutUint64(data[i*8:], w)
	}
	return string(data)
}

// WordMemoryNativeHashKey preserves Vec_MemHashKey's uint32 overflow and
// little-endian half-word order. tableSize is the already-selected bucket
// count; Go interning itself uses lookup-only maps and does not depend on it.
func WordMemoryNativeHashKey(words []uint64, tableSize int) (uint32, error) {
	if len(words) == 0 || tableSize <= 0 {
		return 0, ErrDimension
	}
	primes := [...]uint32{1699, 4177, 5147, 5647, 6343, 7103, 7873, 8147}
	var hash uint32
	index := 0
	for _, word := range words {
		for _, half := range [...]uint32{uint32(word), uint32(word >> 32)} {
			hash += half * primes[index&7]
			index++
		}
	}
	return hash % uint32(tableSize), nil
}

// EnableHash indexes existing unique records in ID order. Duplicate preexisting
// records reject without changing the vector; repeated enable also rejects.
func (m *WordMemory) EnableHash() error {
	if !m.configured() {
		return ErrDimension
	}
	if m.lookup != nil {
		return ErrInvariant
	}
	lookup := make(map[string]int, m.size)
	for i := 0; i < m.size; i++ {
		key := wordMemoryKey(m.entry(i))
		if _, found := lookup[key]; found {
			return ErrInvariant
		}
		lookup[key] = i
	}
	m.lookup = lookup
	return nil
}
func (m *WordMemory) DisableHash() {
	if m != nil {
		m.lookup = nil
	}
}
func (m *WordMemory) HashEnabled() bool { return m != nil && m.lookup != nil }

// Intern returns (ID, alreadyPresent, error). Existing entries retain their IDs;
// failures leave storage, lookup and logical size unchanged.
func (m *WordMemory) Intern(words []uint64) (int, bool, error) {
	if !m.configured() || len(words) != m.entryWords {
		return -1, false, ErrDimension
	}
	if m.lookup == nil {
		return -1, false, ErrInvariant
	}
	key := wordMemoryKey(words)
	if id, ok := m.lookup[key]; ok {
		return id, true, nil
	}
	id := m.size
	size, e := checkedSize(id, 1, 1, m.options.MaxEntries)
	if e != nil {
		return -1, false, e
	}
	if e = m.extend(size); e != nil {
		return -1, false, e
	}
	copy(m.entry(id), words)
	m.lookup[key] = id
	return id, false, nil
}
func (m *WordMemory) Clone() *WordMemory {
	if !m.configured() {
		return nil
	}
	result := *m
	result.pages = make([][]uint64, len(m.pages))
	for i, page := range m.pages {
		result.pages[i] = append([]uint64(nil), page...)
	}
	result.lookup = nil
	if m.lookup != nil {
		result.lookup = make(map[string]int, m.size)
		for i := 0; i < m.size; i++ {
			result.lookup[wordMemoryKey(result.entry(i))] = i
		}
	}
	return &result
}

// HexDump matches Vec_MemDump's uppercase, high-word/high-digit-first lines.
// maxBytes zero selects sixteen million; output bounds check before allocation.
func (m *WordMemory) HexDump(maxBytes int) (string, error) {
	if !m.configured() || maxBytes < 0 {
		return "", ErrDimension
	}
	if maxBytes == 0 {
		maxBytes = 16 << 20
	}
	if m.size == 0 {
		return "", nil
	}
	width, e := checkedSize(1, 16, m.entryWords, maxBytes)
	if e != nil {
		return "", e
	}
	size, e := checkedSize(0, width, m.size, maxBytes)
	if e != nil {
		return "", e
	}
	output := make([]byte, size)
	var bytes [8]byte
	for i := 0; i < m.size; i++ {
		record := m.entry(i)
		offset := i * width
		for w := len(record) - 1; w >= 0; w-- {
			binary.BigEndian.PutUint64(bytes[:], record[w])
			hex.Encode(output[offset:offset+16], bytes[:])
			offset += 16
		}
		output[offset] = '\n'
	}
	for i, c := range output {
		if c >= 'a' && c <= 'f' {
			output[i] = c - 'a' + 'A'
		}
	}
	return string(output), nil
}

type WordMemoryTruthTableDump struct {
	FileName string
	Text     string
	Summary  string
}

// TruthTableDump owns Vec_MemDumpTruthTables' text and status instead of
// writing a caller-selected file or process-global stdout. A blank name uses
// the source's stdout label; a nonblank name retains its filename convention.
func (m *WordMemory) TruthTableDump(name string, lutSize, maxBytes int) (WordMemoryTruthTableDump, error) {
	var result WordMemoryTruthTableDump
	if !m.configured() || lutSize < 0 {
		return result, ErrDimension
	}
	text, err := m.HexDump(maxBytes)
	if err != nil {
		return result, err
	}
	result.FileName = "stdout"
	if name != "" {
		result.FileName = fmt.Sprintf("tt_%s_%02d.txt", name, lutSize)
	}
	result.Text = text
	megabytes := 8.0 * float64(m.size) * float64(m.entryWords) / float64(1<<20)
	result.Summary = fmt.Sprintf("Dumped %d %d-var truth tables into file %q (%.2f MB).\n", m.size, lutSize, result.FileName, megabytes)
	return result, nil
}
func (m *WordMemory) Check() error {
	if !m.configured() {
		return ErrDimension
	}
	entriesPerPage := 1 << uint(m.pageLog)
	count := 0
	if m.size > 0 {
		count = (m.size-1)/entriesPerPage + 1
	}
	if len(m.pages) != count || m.size > m.options.MaxEntries {
		return ErrInvariant
	}
	for _, page := range m.pages {
		if len(page) != m.pageWords {
			return ErrInvariant
		}
	}
	if m.lookup != nil {
		if len(m.lookup) != m.size {
			return ErrInvariant
		}
		for i := 0; i < m.size; i++ {
			id, ok := m.lookup[wordMemoryKey(m.entry(i))]
			if !ok || id != i {
				return ErrInvariant
			}
		}
	}
	return nil
}
