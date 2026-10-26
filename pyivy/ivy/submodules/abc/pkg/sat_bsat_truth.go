package abc

// Truth interning from src/sat/bsat/satTruth.c and satTruth.h.

import (
	"encoding/binary"
	"slices"
)

type SatTruthStoreOptions struct {
	MaxVariables int // zero: 16
	MaxWords     int // zero: 1024 words per function
	MaxEntries   int // zero: one million nonconstant functions
}

type SatTruthStoreStats struct {
	Variables, Words, Lookups, Entries int
}

// SatTruthStore owns normalized truth functions. Handle zero is constant zero,
// handle one is constant one, and every stored function has an even handle;
// XOR one denotes its complement. The lookup map is never iterated, so handle
// assignment comes only from insertion-ordered storage.
type SatTruthStore struct {
	variables, words int
	maxEntries       int
	entries          [][]uint64
	index            map[string]int
	variableHandles  []int
	lookups          int
	closed           bool
}

func satTruthKey(words []uint64) string {
	bytes := make([]byte, 8*len(words))
	for index, word := range words {
		binary.LittleEndian.PutUint64(bytes[8*index:], word)
	}
	return string(bytes)
}

func satTruthAll(words []uint64, value uint64) bool {
	for _, word := range words {
		if word != value {
			return false
		}
	}
	return true
}

func NewSatTruthStore(variables int, options SatTruthStoreOptions) (*SatTruthStore, error) {
	if options.MaxVariables < 0 || options.MaxWords < 0 || options.MaxEntries < 0 {
		return nil, ErrDimension
	}
	maxVariables := options.MaxVariables
	if maxVariables == 0 {
		maxVariables = 16
	}
	if variables <= 0 || variables > maxVariables || variables > 16 {
		return nil, ErrDimension
	}
	words := 1
	if variables >= 6 {
		words = 1 << uint(variables-6)
	}
	maxWords := options.MaxWords
	if maxWords == 0 {
		maxWords = 1024
	}
	if words > maxWords {
		return nil, ErrLimit
	}
	maxEntries := options.MaxEntries
	if maxEntries == 0 {
		maxEntries = 1 << 20
	}
	store := &SatTruthStore{
		variables:       variables,
		words:           words,
		maxEntries:      maxEntries,
		entries:         make([][]uint64, 1),
		index:           make(map[string]int),
		variableHandles: make([]int, variables),
	}
	masks := [...]uint64{
		0xaaaaaaaaaaaaaaaa, 0xcccccccccccccccc, 0xf0f0f0f0f0f0f0f0,
		0xff00ff00ff00ff00, 0xffff0000ffff0000, 0xffffffff00000000,
	}
	work := make([]uint64, words)
	for variable := 0; variable < variables; variable++ {
		for word := range work {
			if variable < 6 {
				work[word] = masks[variable]
			} else if word&(1<<uint(variable-6)) != 0 {
				work[word] = ^uint64(0)
			} else {
				work[word] = 0
			}
		}
		handle, err := store.Insert(work)
		if err != nil {
			return nil, err
		}
		if variable > 0 && handle <= store.variableHandles[variable-1] {
			return nil, ErrInvariant
		}
		store.variableHandles[variable] = handle
	}
	return store, nil
}

func (store *SatTruthStore) check() error {
	if store == nil {
		return ErrDimension
	}
	if store.closed || store.words <= 0 || store.index == nil {
		return ErrInvariant
	}
	return nil
}

// Insert returns a stable signed handle and does not mutate words.
func (store *SatTruthStore) Insert(words []uint64) (int, error) {
	if err := store.check(); err != nil {
		return 0, err
	}
	if len(words) != store.words {
		return 0, ErrDimension
	}
	if satTruthAll(words, 0) {
		return 0, nil
	}
	if satTruthAll(words, ^uint64(0)) {
		return 1, nil
	}
	store.lookups++
	normalized := slices.Clone(words)
	complement := normalized[0]&1 != 0
	if complement {
		for index := range normalized {
			normalized[index] = ^normalized[index]
		}
	}
	key := satTruthKey(normalized)
	handle, found := store.index[key]
	if !found {
		if len(store.entries)-1 == store.maxEntries {
			return 0, ErrLimit
		}
		handle = len(store.entries) << 1
		store.entries = append(store.entries, normalized)
		store.index[key] = handle
	}
	if complement {
		handle ^= 1
	}
	return handle, nil
}

func (store *SatTruthStore) Variable(variable int) ([]uint64, error) {
	if err := store.check(); err != nil {
		return nil, err
	}
	if variable < 0 || variable >= store.variables {
		return nil, ErrDimension
	}
	return store.Function(store.variableHandles[variable])
}

// Function returns an owned regular function. Complemented handles are
// rejected, matching Tru_ManFunc's precondition; callers may complement the
// returned words explicitly.
func (store *SatTruthStore) Function(handle int) ([]uint64, error) {
	if err := store.check(); err != nil {
		return nil, err
	}
	if handle < 0 || handle&1 != 0 {
		return nil, ErrInvalidLiteral
	}
	if handle == 0 {
		return make([]uint64, store.words), nil
	}
	index := handle >> 1
	if index <= 0 || index >= len(store.entries) {
		return nil, ErrInvalidLiteral
	}
	return slices.Clone(store.entries[index]), nil
}

func (store *SatTruthStore) Stats() (SatTruthStoreStats, error) {
	if err := store.check(); err != nil {
		return SatTruthStoreStats{}, err
	}
	return SatTruthStoreStats{Variables: store.variables, Words: store.words, Lookups: store.lookups, Entries: len(store.entries) - 1}, nil
}

func (store *SatTruthStore) Close() error {
	if err := store.check(); err != nil {
		return err
	}
	store.closed = true
	store.entries, store.variableHandles = nil, nil
	store.index = nil
	return nil
}

func SatTruthEqual(left, right []uint64) bool { return slices.Equal(left, right) }
func SatTruthEqualZero(words []uint64) bool   { return satTruthAll(words, 0) }
func SatTruthEqualOne(words []uint64) bool    { return satTruthAll(words, ^uint64(0)) }

func SatTruthClear(words []uint64) []uint64 {
	clear(words)
	return words
}
func SatTruthFill(words []uint64) []uint64 {
	for index := range words {
		words[index] = ^uint64(0)
	}
	return words
}
func SatTruthNot(words []uint64) []uint64 {
	for index := range words {
		words[index] = ^words[index]
	}
	return words
}

func satTruthBinary(output, input []uint64, operation byte, complement bool) ([]uint64, error) {
	if len(output) != len(input) {
		return nil, ErrDimension
	}
	for index, word := range input {
		if complement {
			word = ^word
		}
		switch operation {
		case 'c':
			output[index] = word
		case 'a':
			output[index] &= word
		case 'o':
			output[index] |= word
		default:
			return nil, ErrInvariant
		}
	}
	return output, nil
}

func SatTruthCopy(output, input []uint64, complement bool) ([]uint64, error) {
	return satTruthBinary(output, input, 'c', complement)
}
func SatTruthAnd(output, input []uint64, complement bool) ([]uint64, error) {
	return satTruthBinary(output, input, 'a', complement)
}
func SatTruthOr(output, input []uint64, complement bool) ([]uint64, error) {
	return satTruthBinary(output, input, 'o', complement)
}
