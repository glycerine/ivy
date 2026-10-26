package abc

// Unique string-name management from src/misc/util/utilNam.c and utilNam.h.

import (
	"fmt"
	"slices"
	"strings"
)

type StringNameTableOptions struct {
	InitialNames     int
	AverageNameBytes int
	MaxNames         int
	MaxBytes         int
}

type StringNameTableStats struct {
	Names       int
	StoredBytes int
}

// StringNameTable is an append-only unique string interner. IDs begin at one
// and follow first insertion order. The zero value is invalid.
type StringNameTable struct {
	names            []string
	ids              map[string]int
	storedBytes      int
	maxNames         int
	maxBytes         int
	averageNameBytes int
}

func NewStringNameTable(options StringNameTableOptions) (*StringNameTable, error) {
	if options.InitialNames < 0 || options.AverageNameBytes < 0 || options.MaxNames < 0 || options.MaxBytes < 0 {
		return nil, ErrDimension
	}
	defaultInitialNames := options.InitialNames == 0
	if options.InitialNames == 0 {
		options.InitialNames = 16
	}
	if options.MaxNames == 0 {
		options.MaxNames = 1_000_000
	}
	if options.MaxBytes == 0 {
		options.MaxBytes = 256 << 20
	}
	if defaultInitialNames && options.InitialNames > options.MaxNames {
		options.InitialNames = options.MaxNames
	}
	if options.InitialNames > options.MaxNames || options.MaxBytes < 4 {
		return nil, ErrLimit
	}
	return &StringNameTable{
		names:            make([]string, 1, options.InitialNames+1),
		ids:              make(map[string]int, options.InitialNames),
		storedBytes:      4,
		maxNames:         options.MaxNames,
		maxBytes:         options.MaxBytes,
		averageNameBytes: options.AverageNameBytes,
	}, nil
}

func (table *StringNameTable) check() error {
	if table == nil || table.names == nil || table.ids == nil || table.maxNames <= 0 || table.maxBytes < 4 || len(table.names) != len(table.ids)+1 {
		return fmt.Errorf("%w: closed or uninitialized string-name table", ErrInvariant)
	}
	return nil
}

func (table *StringNameTable) validateName(name string) error {
	if err := table.check(); err != nil {
		return err
	}
	if name == "" || strings.IndexByte(name, 0) >= 0 {
		return ErrDimension
	}
	if len(name) >= table.maxBytes {
		return ErrLimit
	}
	return nil
}

func (table *StringNameTable) Close() error {
	if err := table.check(); err != nil {
		return err
	}
	table.names = nil
	table.ids = nil
	table.storedBytes = 0
	return nil
}

// ObjectCount returns the maximum ID plus one, retaining native ID-zero
// reservation semantics.
func (table *StringNameTable) ObjectCount() (int, error) {
	if err := table.check(); err != nil {
		return 0, err
	}
	return len(table.names), nil
}

func (table *StringNameTable) Stats() (StringNameTableStats, error) {
	if err := table.check(); err != nil {
		return StringNameTableStats{}, err
	}
	return StringNameTableStats{Names: len(table.names) - 1, StoredBytes: table.storedBytes}, nil
}

func (table *StringNameTable) Find(name string) (int, bool, error) {
	if err := table.validateName(name); err != nil {
		return 0, false, err
	}
	id, found := table.ids[name]
	return id, found, nil
}

func stringNamePrefix(text string, limit int) (string, error) {
	if limit <= 0 || limit > len(text) {
		return "", ErrDimension
	}
	return text[:limit], nil
}

func (table *StringNameTable) FindPrefix(text string, limit int) (int, bool, error) {
	name, err := stringNamePrefix(text, limit)
	if err != nil {
		return 0, false, err
	}
	return table.Find(name)
}

// Intern returns the stable ID and whether the name was already present.
func (table *StringNameTable) Intern(name string) (int, bool, error) {
	if err := table.validateName(name); err != nil {
		return 0, false, err
	}
	if id, found := table.ids[name]; found {
		return id, true, nil
	}
	if len(table.ids) >= table.maxNames || len(name)+1 > table.maxBytes-table.storedBytes {
		return 0, false, ErrLimit
	}
	id := len(table.names)
	owned := strings.Clone(name)
	table.names = append(table.names, owned)
	table.ids[owned] = id
	table.storedBytes += len(owned) + 1
	return id, false, nil
}

func (table *StringNameTable) InternPrefix(text string, limit int) (int, bool, error) {
	name, err := stringNamePrefix(text, limit)
	if err != nil {
		return 0, false, err
	}
	return table.Intern(name)
}

func (table *StringNameTable) Name(id int) (string, bool, error) {
	if err := table.check(); err != nil {
		return "", false, err
	}
	if id == 0 {
		return "", false, nil
	}
	if id < 0 || id >= len(table.names) {
		return "", false, ErrDimension
	}
	return table.names[id], true, nil
}

func (table *StringNameTable) Names() ([]string, error) {
	if err := table.check(); err != nil {
		return nil, err
	}
	return slices.Clone(table.names[1:]), nil
}

func (table *StringNameTable) Render() (string, error) {
	if err := table.check(); err != nil {
		return "", err
	}
	var result strings.Builder
	for id := 1; id < len(table.names); id++ {
		fmt.Fprintf(&result, "%8d = %s\n", id, table.names[id])
	}
	return result.String(), nil
}

// IDMap returns a native-shaped vector indexed by this table's IDs. Missing
// names and the reserved zeroth entry map to zero.
func (table *StringNameTable) IDMap(other *StringNameTable) ([]int, error) {
	if err := table.check(); err != nil {
		return nil, err
	}
	if err := other.check(); err != nil {
		return nil, err
	}
	result := make([]int, len(table.names))
	if table == other {
		for id := range result {
			result[id] = id
		}
		return result, nil
	}
	for id := 1; id < len(table.names); id++ {
		result[id] = other.ids[table.names[id]]
	}
	return result, nil
}

func (table *StringNameTable) validateIDs(ids []int) error {
	if err := table.check(); err != nil {
		return err
	}
	for _, id := range ids {
		if id <= 0 || id >= len(table.names) {
			return ErrDimension
		}
	}
	return nil
}

func (table *StringNameTable) CommonCount(ids []int, other *StringNameTable) (int, error) {
	if err := table.validateIDs(ids); err != nil {
		return 0, err
	}
	if err := other.check(); err != nil {
		return 0, err
	}
	count := 0
	for _, id := range ids {
		if _, found := other.ids[table.names[id]]; found {
			count++
		}
	}
	return count, nil
}

func (table *StringNameTable) FirstUnique(ids []int, other *StringNameTable) (string, bool, error) {
	if err := table.validateIDs(ids); err != nil {
		return "", false, err
	}
	if err := other.check(); err != nil {
		return "", false, err
	}
	for _, id := range ids {
		if _, found := other.ids[table.names[id]]; !found {
			return table.names[id], true, nil
		}
	}
	return "", false, nil
}

var stringNameHashPrimes = [...]uint32{
	1009, 1049, 1093, 1151, 1201, 1249, 1297, 1361, 1427, 1459, 1499, 1559, 1607, 1657, 1709, 1759,
	1823, 1877, 1933, 1997, 2039, 2089, 2141, 2213, 2269, 2311, 2371, 2411, 2467, 2543, 2609, 2663,
	2699, 2741, 2797, 2851, 2909, 2969, 3037, 3089, 3169, 3221, 3299, 3331, 3389, 3461, 3517, 3557,
	3613, 3671, 3719, 3779, 3847, 3907, 3943, 4013, 4073, 4129, 4201, 4243, 4289, 4363, 4441, 4493,
	4549, 4621, 4663, 4729, 4793, 4871, 4933, 4973, 5021, 5087, 5153, 5227, 5281, 5351, 5417, 5471,
	5519, 5573, 5651, 5693, 5749, 5821, 5861, 5923, 6011, 6073, 6131, 6199, 6257, 6301, 6353, 6397,
	6481, 6563, 6619, 6689, 6737, 6803, 6863, 6917, 6977, 7027, 7109, 7187, 7237, 7309, 7393, 7477,
	7523, 7561, 7607, 7681, 7727, 7817, 7877, 7933, 8011, 8039, 8059, 8081, 8093, 8111, 8123, 8147,
}

func nativeNameByte(value byte) uint32 { return uint32(int32(int8(value))) }

// StringNameHash preserves Abc_NamStrHash's pinned signed-char arithmetic.
// A negative limit hashes through the first NUL; a nonnegative limit hashes
// exactly that prefix, including embedded NUL bytes as the bounded C form does.
func StringNameHash(text string, limit, tableSize int) (int, error) {
	if tableSize <= 0 || limit < -1 || limit > len(text) {
		return 0, ErrDimension
	}
	if limit < 0 {
		limit = len(text)
		if nul := strings.IndexByte(text, 0); nul >= 0 {
			limit = nul
		}
	}
	var hash uint32
	for index := 0; index < limit; index++ {
		term := nativeNameByte(text[index]) * stringNameHashPrimes[index&0x7f]
		if index&1 != 0 {
			hash *= term
		} else {
			hash ^= term
		}
	}
	return int(hash % uint32(tableSize)), nil
}

// StringNameHashJenkins preserves Abc_NamStrHash2, including stopping at a NUL
// even when a bounded prefix was supplied.
func StringNameHashJenkins(text string, limit, tableSize int) (int, error) {
	if tableSize <= 0 || limit < -1 || limit > len(text) {
		return 0, ErrDimension
	}
	if limit < 0 {
		limit = len(text)
	}
	var hash uint32
	for index := 0; index < limit && text[index] != 0; index++ {
		hash += nativeNameByte(text[index])
		hash += hash << 10
		hash ^= hash >> 6
	}
	hash += hash << 3
	hash ^= hash >> 11
	hash += hash << 15
	return int(hash % uint32(tableSize)), nil
}
