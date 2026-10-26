package abc

// Port of src/sat/bsat/satStore.c and the store portion of satStore.h.

import (
	"fmt"
	"math"
	"os"
	"slices"
	"strconv"
	"strings"
	"unsafe"
)

const satProofStoreChunkBytes = 1 << 16

type SatProofStoreLimits struct {
	MaxVariables, MaxClauses, MaxLiterals, MaxBytes int // zero: 1M, 16M, 64M, 1 GiB
}

type SatProofStoreClause struct {
	ID                   int
	PartA, Root, Visited bool
	Literals             []SatLit
}

type SatProofStoreStats struct {
	Variables, Roots, Clauses, ClausesA, Literals int
	EmptyClause                                   int // -1 when absent
	Chunks, LastChunkBytes, MemoryBytes           int
	Stopped                                       bool
}

type SatProofStore struct {
	variables, roots, clausesA, literals int
	empty                                int
	clauses                              []SatProofStoreClause
	chunks, chunkUsed                    int
	limits                               SatProofStoreLimits
	stopped                              bool
}

func (s *SatProofStore) clone() *SatProofStore {
	if s == nil {
		return nil
	}
	result := *s
	result.clauses = make([]SatProofStoreClause, len(s.clauses))
	for index, clause := range s.clauses {
		result.clauses[index] = clause
		result.clauses[index].Literals = slices.Clone(clause.Literals)
	}
	return &result
}

func checkSatProofStoreLimits(limits SatProofStoreLimits) (SatProofStoreLimits, error) {
	if limits.MaxVariables < 0 || limits.MaxClauses < 0 || limits.MaxLiterals < 0 || limits.MaxBytes < 0 {
		return limits, ErrDimension
	}
	if limits.MaxVariables == 0 {
		limits.MaxVariables = 1 << 20
	}
	if limits.MaxClauses == 0 {
		limits.MaxClauses = 16 << 20
	}
	if limits.MaxLiterals == 0 {
		limits.MaxLiterals = 64 << 20
	}
	if limits.MaxBytes == 0 {
		limits.MaxBytes = 1 << 30
	}
	return limits, nil
}

func NewSatProofStore(limits SatProofStoreLimits) (*SatProofStore, error) {
	limits, err := checkSatProofStoreLimits(limits)
	if err != nil {
		return nil, err
	}
	return &SatProofStore{empty: -1, limits: limits}, nil
}

func satProofStoreRecordBytes(literals int) (int, error) {
	if literals < 0 || literals > (math.MaxInt-3*int(unsafe.Sizeof(uintptr(0)))-8)/4 {
		return 0, ErrLimit
	}
	pointerBytes := int(unsafe.Sizeof(uintptr(0)))
	bytes := 3*pointerBytes + 8 + 4*literals
	bytes = (bytes + pointerBytes - 1) / pointerBytes * pointerBytes
	if bytes > satProofStoreChunkBytes-pointerBytes {
		return 0, ErrLimit
	}
	return bytes, nil
}

func (s *SatProofStore) memoryUsage() int {
	if s == nil || s.chunks == 0 {
		return 0
	}
	return (s.chunks-1)*satProofStoreChunkBytes + s.chunkUsed
}

func (s *SatProofStore) reserve(bytes int) error {
	pointerBytes := int(unsafe.Sizeof(uintptr(0)))
	newChunks, newUsed := s.chunks, s.chunkUsed
	if newChunks == 0 || bytes > satProofStoreChunkBytes-newUsed {
		newChunks++
		newUsed = pointerBytes
	}
	newUsed += bytes
	memory := (newChunks-1)*satProofStoreChunkBytes + newUsed
	if memory > s.limits.MaxBytes {
		return ErrLimit
	}
	s.chunks, s.chunkUsed = newChunks, newUsed
	return nil
}

func (s *SatProofStore) AddClause(literals []SatLit) (int, error) {
	if s == nil || s.stopped {
		return -1, ErrInvariant
	}
	if len(s.clauses) == s.limits.MaxClauses || len(literals) > s.limits.MaxLiterals-s.literals {
		return -1, ErrLimit
	}
	if len(literals) == 0 && s.empty >= 0 {
		return -1, ErrInvariant
	}
	owned := slices.Clone(literals)
	for _, literal := range owned {
		if !literal.Valid() {
			return -1, ErrInvalidLiteral
		}
	}
	slices.Sort(owned)
	for i := 1; i < len(owned); i++ {
		if owned[i-1].Var() == owned[i].Var() {
			return -1, ErrInvariant
		}
	}
	variables := s.variables
	if len(owned) > 0 {
		variables = max(variables, int(owned[len(owned)-1].Var())+1)
	}
	if variables > s.limits.MaxVariables {
		return -1, ErrLimit
	}
	bytes, err := satProofStoreRecordBytes(len(owned))
	if err != nil {
		return -1, err
	}
	if err = s.reserve(bytes); err != nil {
		return -1, err
	}
	id := len(s.clauses)
	s.clauses = append(s.clauses, SatProofStoreClause{ID: id, Literals: owned})
	s.variables, s.literals = variables, s.literals+len(owned)
	if len(owned) == 0 {
		s.empty = id
	}
	return id, nil
}

func (s *SatProofStore) Clause(id int) (SatProofStoreClause, error) {
	if s == nil || s.stopped || id < 0 || id >= len(s.clauses) {
		return SatProofStoreClause{}, ErrInvalidLiteral
	}
	clause := s.clauses[id]
	clause.Literals = slices.Clone(clause.Literals)
	return clause, nil
}

func (s *SatProofStore) Clauses() []SatProofStoreClause {
	if s == nil || s.stopped {
		return nil
	}
	result := make([]SatProofStoreClause, len(s.clauses))
	for i := range s.clauses {
		result[i], _ = s.Clause(i)
	}
	return result
}

func (s *SatProofStore) RootClauses() []SatProofStoreClause {
	if s == nil || s.stopped {
		return nil
	}
	count := 0
	for count < len(s.clauses) && s.clauses[count].Root {
		count++
	}
	result := make([]SatProofStoreClause, count)
	for i := range result {
		result[i], _ = s.Clause(i)
	}
	return result
}

func (s *SatProofStore) MarkRoots() error {
	if s == nil || s.stopped {
		return ErrInvariant
	}
	for i := range s.clauses {
		s.clauses[i].Root = true
	}
	s.roots = len(s.clauses)
	return nil
}

func (s *SatProofStore) MarkClausesA() error {
	if s == nil || s.stopped {
		return ErrInvariant
	}
	for i := range s.clauses {
		s.clauses[i].PartA = true
	}
	s.clausesA = len(s.clauses)
	return nil
}

func (s *SatProofStore) SetVisited(id int, visited bool) error {
	if s == nil || s.stopped {
		return ErrInvariant
	}
	if id < 0 || id >= len(s.clauses) {
		return ErrInvalidLiteral
	}
	s.clauses[id].Visited = visited
	return nil
}

func (s *SatProofStore) ChangeLastClause() (SatVar, error) {
	if s == nil || s.stopped {
		return 0, ErrInvariant
	}
	if len(s.clauses) == 0 {
		return 0, ErrInvariant
	}
	last := &s.clauses[len(s.clauses)-1]
	if !last.PartA || len(last.Literals) != 1 || s.clausesA == 0 {
		return 0, ErrInvariant
	}
	last.PartA = false
	s.clausesA--
	return last.Literals[0].Var(), nil
}

func (s *SatProofStore) Stats() SatProofStoreStats {
	if s == nil {
		return SatProofStoreStats{EmptyClause: -1}
	}
	return SatProofStoreStats{s.variables, s.roots, len(s.clauses), s.clausesA, s.literals, s.empty, s.chunks, s.chunkUsed, s.memoryUsage(), s.stopped}
}

func (s *SatProofStore) MarshalText() ([]byte, error) {
	if s == nil || s.stopped {
		return nil, ErrInvariant
	}
	var result strings.Builder
	fmt.Fprintf(&result, "p %d %d %d %d\n", s.variables, len(s.clauses), s.roots, s.clausesA)
	for _, clause := range s.clauses {
		for _, literal := range clause.Literals {
			fmt.Fprintf(&result, " %d", literal.DIMACS())
		}
		result.WriteString(" 0\n")
	}
	return []byte(result.String()), nil
}

func (s *SatProofStore) WriteFile(name string) error {
	if name == "" {
		return ErrDimension
	}
	data, err := s.MarshalText()
	if err != nil {
		return err
	}
	return os.WriteFile(name, data, 0o666)
}

func parseSatProofStoreHeader(line string) ([4]int, error) {
	fields := strings.Fields(line)
	if len(fields) != 5 || fields[0] != "p" {
		return [4]int{}, ErrDimension
	}
	var result [4]int
	for i := range result {
		value, err := strconv.ParseInt(fields[i+1], 10, 0)
		if err != nil || value < 0 || value > int64(math.MaxInt) {
			return [4]int{}, ErrDimension
		}
		result[i] = int(value)
	}
	if result[2] > result[1] || result[3] > result[1] {
		return [4]int{}, ErrInvariant
	}
	return result, nil
}

func ParseSatProofStore(data []byte, limits SatProofStoreLimits) (*SatProofStore, error) {
	limits, err := checkSatProofStoreLimits(limits)
	if err != nil {
		return nil, err
	}
	if len(data) > limits.MaxBytes {
		return nil, ErrLimit
	}
	lines := strings.Split(string(data), "\n")
	headerIndex := -1
	var header [4]int
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "c") {
			continue
		}
		header, err = parseSatProofStoreHeader(trimmed)
		if err != nil {
			return nil, fmt.Errorf("proof store header: %w", err)
		}
		headerIndex = i
		break
	}
	if headerIndex < 0 {
		return nil, ErrDimension
	}
	if header[0] > limits.MaxVariables || header[1] > limits.MaxClauses {
		return nil, ErrLimit
	}
	store, _ := NewSatProofStore(limits)
	var clause []SatLit
	for _, token := range strings.Fields(strings.Join(lines[headerIndex+1:], "\n")) {
		value, parseErr := strconv.ParseInt(token, 10, 0)
		if parseErr != nil || value < int64(math.MinInt) || value > int64(math.MaxInt) {
			return nil, ErrDimension
		}
		if value == 0 {
			if _, err = store.AddClause(clause); err != nil {
				return nil, err
			}
			clause = clause[:0]
			continue
		}
		literal, literalErr := ParseSatDIMACSLiteral(int(value))
		if literalErr != nil {
			return nil, literalErr
		}
		clause = append(clause, literal)
		if len(clause) > limits.MaxLiterals-store.literals {
			return nil, ErrLimit
		}
	}
	if len(clause) != 0 || len(store.clauses) != header[1] || store.variables > header[0] {
		return nil, ErrInvariant
	}
	store.variables, store.roots, store.clausesA = header[0], header[2], header[3]
	for i := 0; i < store.roots; i++ {
		store.clauses[i].Root = true
	}
	for i := 0; i < store.clausesA; i++ {
		store.clauses[i].PartA = true
	}
	return store, nil
}

func LoadSatProofStore(name string, limits SatProofStoreLimits) (*SatProofStore, error) {
	if name == "" {
		return nil, ErrDimension
	}
	data, err := os.ReadFile(name)
	if err != nil {
		return nil, err
	}
	return ParseSatProofStore(data, limits)
}

func (s *SatProofStore) Stop() {
	if s == nil || s.stopped {
		return
	}
	s.clauses, s.chunks, s.chunkUsed, s.stopped = nil, 0, 0, true
}
