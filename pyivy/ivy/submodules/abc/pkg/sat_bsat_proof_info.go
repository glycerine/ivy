package abc

// Port of src/sat/bsat/satProof2.h as an owned proof-information manager.

import (
	"slices"
	"sort"
	"unsafe"
)

type SatProofInfoLimits struct {
	MaxWidth, MaxRecords, MaxWords int // zero: 16M bits, 16M records, 64M words
}

type SatProofClauseRef struct {
	Learned bool
	ID      int
}

type SatProofInfoStats struct {
	FirstClause, WordCount, Records, Saved, WidthCapacity int
	Active                                                bool
	OwnedBytes                                            uint64
}

type SatProofInfoManager struct {
	first, firstSaved, words int
	info                     []uint64
	saved                    []int
	idToProof                []int
	hasMapping               bool
	active                   int
	limits                   SatProofInfoLimits
	stopped                  bool
}

func (m *SatProofInfoManager) clone() *SatProofInfoManager {
	if m == nil {
		return nil
	}
	clone := *m
	clone.info = slices.Clone(m.info)
	clone.saved = slices.Clone(m.saved)
	clone.idToProof = slices.Clone(m.idToProof)
	return &clone
}

func checkSatProofInfoLimits(limits SatProofInfoLimits) (SatProofInfoLimits, error) {
	if limits.MaxWidth < 0 || limits.MaxRecords < 0 || limits.MaxWords < 0 {
		return limits, ErrDimension
	}
	if limits.MaxWidth == 0 {
		limits.MaxWidth = 16 << 20
	}
	if limits.MaxRecords == 0 {
		limits.MaxRecords = 16 << 20
	}
	if limits.MaxWords == 0 {
		limits.MaxWords = 64 << 20
	}
	return limits, nil
}

func NewSatProofInfoManager(limits SatProofInfoLimits) (*SatProofInfoManager, error) {
	limits, err := checkSatProofInfoLimits(limits)
	if err != nil {
		return nil, err
	}
	return &SatProofInfoManager{first: -1, firstSaved: -1, active: -1, limits: limits}, nil
}

func SatProofBitWordCount(width int) (int, error) {
	if width < 0 {
		return 0, ErrDimension
	}
	words := width / 64
	if width&63 != 0 {
		words++
	}
	return words, nil
}

func (m *SatProofInfoManager) Size() int {
	if m == nil || m.stopped || m.words == 0 {
		return 0
	}
	return len(m.info) / m.words
}

func (m *SatProofInfoManager) Restart(idToProof []int, firstClause, width int) error {
	if m == nil || m.stopped || m.first != -1 || m.active != -1 {
		return ErrInvariant
	}
	if firstClause < 0 || width <= 0 || width > m.limits.MaxWidth {
		return ErrDimension
	}
	words, err := SatProofBitWordCount(width)
	if err != nil || words > m.limits.MaxWords {
		return ErrLimit
	}
	for _, entry := range idToProof {
		if entry < -1 || entry >= 64*words {
			return ErrInvalidLiteral
		}
	}
	m.first, m.firstSaved, m.words = firstClause, -1, words
	m.info, m.saved = m.info[:0], m.saved[:0]
	m.hasMapping = idToProof != nil
	m.idToProof = slices.Clone(idToProof)
	return nil
}

func (m *SatProofInfoManager) Grow(width int) error {
	if m == nil || m.stopped || m.first < 0 || m.active != -1 {
		return ErrInvariant
	}
	if width < 0 || width > m.limits.MaxWidth {
		return ErrDimension
	}
	if width < 64*m.words {
		return nil
	}
	requested, err := SatProofBitWordCount(width)
	if err != nil {
		return err
	}
	newWords := max(requested, 2*m.words)
	if newWords < m.words || newWords > m.limits.MaxWords || m.Size() > m.limits.MaxWords/newWords {
		return ErrLimit
	}
	newInfo := make([]uint64, m.Size()*newWords)
	for record := 0; record < m.Size(); record++ {
		copy(newInfo[record*newWords:], m.info[record*m.words:(record+1)*m.words])
	}
	m.info, m.words = newInfo, newWords
	return nil
}

func (m *SatProofInfoManager) SetProofID(rootID, bit int) error {
	if m == nil || m.stopped || m.first < 0 || !m.hasMapping || rootID < 0 || rootID >= len(m.idToProof) || bit < -1 || bit >= 64*m.words {
		return ErrInvalidLiteral
	}
	m.idToProof[rootID] = bit
	return nil
}

func (m *SatProofInfoManager) ProofIDs() []int {
	if m == nil || m.stopped || !m.hasMapping {
		return nil
	}
	return slices.Clone(m.idToProof)
}

func (m *SatProofInfoManager) Record(index int) ([]uint64, error) {
	if m == nil || m.stopped || index < 0 || index >= m.Size() {
		return nil, ErrInvalidLiteral
	}
	return slices.Clone(m.info[index*m.words : (index+1)*m.words]), nil
}

func (m *SatProofInfoManager) Shrink(clauseID int) error {
	if m == nil || m.stopped || m.first < 0 || m.active != -1 {
		return ErrInvariant
	}
	index := clauseID - m.first
	if index < 0 || index >= m.Size() {
		return ErrInvalidLiteral
	}
	m.info = m.info[:index*m.words]
	return nil
}

func (m *SatProofInfoManager) AddSaved(clauseID, newID int) error {
	if m == nil || m.stopped || m.first < 0 || m.active != -1 {
		return ErrInvariant
	}
	if newID < 0 {
		return ErrDimension
	}
	if clauseID < m.first {
		return nil
	}
	index := clauseID - m.first
	if index < 0 || index >= m.Size() {
		return ErrInvalidLiteral
	}
	if len(m.saved) == 0 {
		m.firstSaved = newID
	}
	m.saved = append(m.saved, clauseID)
	return nil
}

func (m *SatProofInfoManager) Compact(newFirst int) error {
	if m == nil || m.stopped || m.first < 0 || m.active != -1 {
		return ErrInvariant
	}
	if newFirst < 0 || len(m.saved) > m.limits.MaxRecords {
		return ErrDimension
	}
	for _, clauseID := range m.saved {
		if clauseID < m.first || clauseID-m.first >= m.Size() {
			return ErrInvalidLiteral
		}
	}
	newInfo := make([]uint64, len(m.saved)*m.words)
	for i, clauseID := range m.saved {
		index := clauseID - m.first
		copy(newInfo[i*m.words:], m.info[index*m.words:(index+1)*m.words])
	}
	m.info, m.saved = newInfo, m.saved[:0]
	if m.firstSaved == -1 {
		m.first = newFirst
	} else {
		m.first = m.firstSaved
	}
	m.firstSaved = -1
	return nil
}

func (m *SatProofInfoManager) resolve(reference SatProofClauseRef) error {
	if reference.ID < 0 {
		return nil
	}
	current := m.info[m.active*m.words : (m.active+1)*m.words]
	if reference.Learned {
		if reference.ID < m.first {
			return nil
		}
		index := reference.ID - m.first
		if index < 0 || index >= m.Size() {
			return ErrInvalidLiteral
		}
		other := m.info[index*m.words : (index+1)*m.words]
		for word := range current {
			current[word] |= other[word]
		}
		return nil
	}
	bit := reference.ID
	if m.hasMapping {
		if reference.ID >= len(m.idToProof) {
			return ErrInvalidLiteral
		}
		bit = m.idToProof[reference.ID]
	}
	if bit < 0 {
		return nil
	}
	if bit >= 64*m.words {
		return ErrInvalidLiteral
	}
	current[bit>>6] |= uint64(1) << uint(bit&63)
	return nil
}

func (m *SatProofInfoManager) ChainStart(reference SatProofClauseRef) error {
	if m == nil || m.stopped || m.first < 0 || m.active != -1 {
		return ErrInvariant
	}
	if m.Size() >= m.limits.MaxRecords || len(m.info) > m.limits.MaxWords-m.words {
		return ErrLimit
	}
	oldLength := len(m.info)
	m.info = append(m.info, make([]uint64, m.words)...)
	m.active = m.Size() - 1
	if err := m.resolve(reference); err != nil {
		m.info, m.active = m.info[:oldLength], -1
		return err
	}
	return nil
}

func (m *SatProofInfoManager) ChainResolve(reference SatProofClauseRef) error {
	if m == nil || m.stopped || m.first < 0 || m.active < 0 {
		return ErrInvariant
	}
	return m.resolve(reference)
}

func (m *SatProofInfoManager) ChainStop() (int, error) {
	if m == nil || m.stopped || m.active < 0 {
		return 0, ErrInvariant
	}
	m.active = -1
	return 0, nil
}

func (m *SatProofInfoManager) UnsatCore() ([]int, error) {
	if m == nil || m.stopped || m.first < 0 || m.active != -1 || m.Size() == 0 {
		return nil, ErrInvariant
	}
	record := m.info[(m.Size()-1)*m.words:]
	var core []int
	if !m.hasMapping {
		for bit := 0; bit < 64*m.words; bit++ {
			if record[bit>>6]>>uint(bit&63)&1 != 0 {
				core = append(core, bit)
			}
		}
	} else {
		for id, bit := range m.idToProof {
			if bit >= 0 && record[bit>>6]>>uint(bit&63)&1 != 0 {
				core = append(core, id)
			}
		}
	}
	sort.Sort(sort.Reverse(sort.IntSlice(core)))
	return core, nil
}

func (m *SatProofInfoManager) Stats() SatProofInfoStats {
	if m == nil || m.stopped {
		return SatProofInfoStats{}
	}
	bytes := uint64(unsafe.Sizeof(*m)) + uint64(cap(m.info))*8 + uint64(cap(m.saved))*uint64(unsafe.Sizeof(int(0))) + uint64(cap(m.idToProof))*uint64(unsafe.Sizeof(int(0)))
	return SatProofInfoStats{m.first, m.words, m.Size(), len(m.saved), 64 * m.words, m.active >= 0, bytes}
}

func (m *SatProofInfoManager) Stop() {
	if m == nil || m.stopped {
		return
	}
	m.info, m.saved, m.idToProof = nil, nil, nil
	m.first, m.firstSaved, m.words, m.active, m.stopped = -1, -1, 0, -1, true
}
