package abc

// Truth-table memory seeds from src/misc/vec/vecMem.h:
// Vec_MemAllocForTTSimple/AllocForTT and Vec_MemAddMuxTT.

// TruthTableWordCount returns the uint64 count used by ABC's packed tables.
// Up to six variables use one word; smaller tables repeat their bit pattern
// within that word. Counts outside C's positive signed-int shift range reject.
func TruthTableWordCount(variables int) (int, error) {
	if variables < 0 {
		return 0, ErrDimension
	}
	if variables <= 6 {
		return 1, nil
	}
	if variables > 36 {
		return 0, ErrLimit
	}
	return 1 << uint(variables-6), nil
}

// NewTruthTableMemory creates empty hash-enabled storage with 4096 records
// per page, matching AllocForTTSimple. Allocation limits check before storage
// use, including the minimum page size. Stable IDs follow Intern call order.
func NewTruthTableMemory(variables int, options WordMemoryOptions) (*WordMemory, error) {
	words, e := TruthTableWordCount(variables)
	if e != nil {
		return nil, e
	}
	m, e := NewWordMemory(words, 12, options)
	if e != nil {
		return nil, e
	}
	if e = m.EnableHash(); e != nil {
		return nil, e
	}
	return m, nil
}

// NewSeededTruthTableMemory interns false as ID 0 and variable zero (or its
// complement) as ID 1. Packed patterns stretch across every allocated word,
// including unused small-table bit positions, exactly as in C.
func NewSeededTruthTableMemory(variables int, complement bool, options WordMemoryOptions) (*WordMemory, error) {
	m, e := NewTruthTableMemory(variables, options)
	if e != nil {
		return nil, e
	}
	entry := make([]uint64, m.entryWords)
	if _, _, e = m.Intern(entry); e != nil {
		return nil, e
	}
	pattern := uint64(0xaaaaaaaaaaaaaaaa)
	if complement {
		pattern = 0x5555555555555555
	}
	for i := range entry {
		entry[i] = pattern
	}
	if _, _, e = m.Intern(entry); e != nil {
		return nil, e
	}
	return m, nil
}

// AddMuxTruthTable interns the stretched x2 ? x1 : x0 pattern at ID 2.
// Existing ID-2 mux entries are idempotent even after other insertions. Wrong
// widths/hash state or a conflicting ID layout reject without mutation, replacing
// C's post-insertion assert. variables controls width, not the source patterns.
func (m *WordMemory) AddMuxTruthTable(variables int) (int, error) {
	words, e := TruthTableWordCount(variables)
	if e != nil {
		return -1, e
	}
	if !m.configured() || words != m.entryWords {
		return -1, ErrDimension
	}
	if m.lookup == nil {
		return -1, ErrInvariant
	}
	entry := make([]uint64, words)
	for i := range entry {
		entry[i] = 0xcacacacacacacaca
	}
	if id, exists := m.lookup[wordMemoryKey(entry)]; exists {
		if id == 2 {
			return id, nil
		}
		return -1, ErrInvariant
	}
	if m.size != 2 {
		return -1, ErrInvariant
	}
	id, _, e := m.Intern(entry)
	return id, e
}
