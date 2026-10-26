package abc

// Deterministic table behavior from src/misc/st/st.c and stmm.c.

import (
	"math"
	"strings"

	"github.com/glycerine/insdict"
)

type stableTableEntry[K comparable, V any] struct {
	key          K
	value        V
	previousSame *stableTableEntry[K, V]
	live         bool
}

type StableTable[K comparable, V any] struct {
	index      *insdict.Dict[K, *stableTableEntry[K, V]]
	entries    []*stableTableEntry[K, V]
	live       int
	generation uint64
	closed     bool
}

type StableTableItem[K comparable, V any] struct {
	Key   K
	Value V
}

type StableTableAction uint8

const (
	StableTableContinue StableTableAction = iota
	StableTableStop
	StableTableDelete
)

type StableTableSlot[K comparable, V any] struct {
	table *StableTable[K, V]
	entry *stableTableEntry[K, V]
}

type StableTableCursor[K comparable, V any] struct {
	table      *StableTable[K, V]
	index      int
	generation uint64
	closed     bool
}

func NewStableTable[K comparable, V any](capacity int) (*StableTable[K, V], error) {
	if capacity < 0 {
		return nil, ErrDimension
	}
	return &StableTable[K, V]{index: insdict.NewDictSize[K, *stableTableEntry[K, V]](capacity)}, nil
}

func (table *StableTable[K, V]) check() error {
	if table == nil || table.closed || table.index == nil {
		return ErrInvariant
	}
	return nil
}

func (table *StableTable[K, V]) bump() error {
	if table.generation == math.MaxUint64 {
		return ErrLimit
	}
	table.generation++
	return nil
}

func (table *StableTable[K, V]) Len() (int, error) {
	if err := table.check(); err != nil {
		return 0, err
	}
	return table.live, nil
}

func (table *StableTable[K, V]) Get(key K) (V, bool, error) {
	var zero V
	if err := table.check(); err != nil {
		return zero, false, err
	}
	entry, found := table.index.Get2(key)
	if !found || entry == nil || !entry.live {
		return zero, false, nil
	}
	return entry.value, true, nil
}

func (table *StableTable[K, V]) addDirect(key K, value V) (*stableTableEntry[K, V], error) {
	if err := table.bump(); err != nil {
		return nil, err
	}
	previous, _ := table.index.Get2(key)
	entry := &stableTableEntry[K, V]{key: key, value: value, previousSame: previous, live: true}
	table.entries = append(table.entries, entry)
	table.index.Set(key, entry)
	table.live++
	return entry, nil
}

// Set inserts or replaces the newest entry for key. Replaced reports whether
// an existing entry was updated.
func (table *StableTable[K, V]) Set(key K, value V) (replaced bool, err error) {
	if err = table.check(); err != nil {
		return false, err
	}
	if entry, found := table.index.Get2(key); found {
		if entry == nil || !entry.live {
			return false, ErrInvariant
		}
		entry.value = value
		return true, nil
	}
	_, err = table.addDirect(key, value)
	return false, err
}

// AddDirect always adds an entry. A duplicate shadows older equal keys until
// it is deleted, matching the native direct-insertion operation.
func (table *StableTable[K, V]) AddDirect(key K, value V) error {
	if err := table.check(); err != nil {
		return err
	}
	_, err := table.addDirect(key, value)
	return err
}

func (table *StableTable[K, V]) FindSlot(key K) (*StableTableSlot[K, V], bool, error) {
	if err := table.check(); err != nil {
		return nil, false, err
	}
	entry, found := table.index.Get2(key)
	if !found {
		return nil, false, nil
	}
	if entry == nil || !entry.live {
		return nil, false, ErrInvariant
	}
	return &StableTableSlot[K, V]{table: table, entry: entry}, true, nil
}

// FindOrAdd returns a checked mutable value slot. Existing reports whether the
// key was already present; a new entry starts with V's zero value.
func (table *StableTable[K, V]) FindOrAdd(key K) (slot *StableTableSlot[K, V], existing bool, err error) {
	if err = table.check(); err != nil {
		return nil, false, err
	}
	if entry, found := table.index.Get2(key); found {
		if entry == nil || !entry.live {
			return nil, false, ErrInvariant
		}
		return &StableTableSlot[K, V]{table: table, entry: entry}, true, nil
	}
	var zero V
	entry, err := table.addDirect(key, zero)
	if err != nil {
		return nil, false, err
	}
	return &StableTableSlot[K, V]{table: table, entry: entry}, false, nil
}

func (slot *StableTableSlot[K, V]) check() error {
	if slot == nil || slot.table == nil || slot.entry == nil {
		return ErrInvariant
	}
	if err := slot.table.check(); err != nil {
		return err
	}
	if !slot.entry.live {
		return ErrInvariant
	}
	return nil
}

func (slot *StableTableSlot[K, V]) Key() (K, error) {
	var zero K
	if err := slot.check(); err != nil {
		return zero, err
	}
	return slot.entry.key, nil
}

func (slot *StableTableSlot[K, V]) Value() (V, error) {
	var zero V
	if err := slot.check(); err != nil {
		return zero, err
	}
	return slot.entry.value, nil
}

func (slot *StableTableSlot[K, V]) Set(value V) error {
	if err := slot.check(); err != nil {
		return err
	}
	slot.entry.value = value
	return nil
}

func (table *StableTable[K, V]) removeEntry(entry *stableTableEntry[K, V]) error {
	if entry == nil || !entry.live {
		return ErrInvariant
	}
	head, found := table.index.Get2(entry.key)
	if !found || head == nil || !head.live {
		return ErrInvariant
	}
	if err := table.bump(); err != nil {
		return err
	}
	if head == entry {
		if entry.previousSame == nil {
			if !table.index.Del(entry.key) {
				return ErrInvariant
			}
		} else {
			table.index.Set(entry.key, entry.previousSame)
		}
	} else {
		previous := head
		for previous.previousSame != nil && previous.previousSame != entry {
			previous = previous.previousSame
		}
		if previous.previousSame != entry {
			return ErrInvariant
		}
		previous.previousSame = entry.previousSame
	}
	entry.live = false
	entry.previousSame = nil
	var zeroK K
	var zeroV V
	entry.key = zeroK
	entry.value = zeroV
	table.live--
	return nil
}

// Delete removes the newest equal key and returns the stored key and value.
func (table *StableTable[K, V]) Delete(key K) (storedKey K, value V, found bool, err error) {
	if err = table.check(); err != nil {
		return storedKey, value, false, err
	}
	entry, found := table.index.Get2(key)
	if !found {
		return storedKey, value, false, nil
	}
	if entry == nil || !entry.live {
		return storedKey, value, false, ErrInvariant
	}
	storedKey, value = entry.key, entry.value
	if err = table.removeEntry(entry); err != nil {
		return storedKey, value, false, err
	}
	return storedKey, value, true, nil
}

func (table *StableTable[K, V]) Entries() ([]StableTableItem[K, V], error) {
	if err := table.check(); err != nil {
		return nil, err
	}
	result := make([]StableTableItem[K, V], 0, table.live)
	for _, entry := range table.entries {
		if entry.live {
			result = append(result, StableTableItem[K, V]{Key: entry.key, Value: entry.value})
		}
	}
	return result, nil
}

func (table *StableTable[K, V]) Cursor() (*StableTableCursor[K, V], error) {
	if err := table.check(); err != nil {
		return nil, err
	}
	return &StableTableCursor[K, V]{table: table, generation: table.generation}, nil
}

func (cursor *StableTableCursor[K, V]) Next() (K, V, bool, error) {
	var zeroK K
	var zeroV V
	if cursor == nil || cursor.closed || cursor.table == nil {
		return zeroK, zeroV, false, ErrInvariant
	}
	if err := cursor.table.check(); err != nil {
		return zeroK, zeroV, false, err
	}
	if cursor.generation != cursor.table.generation {
		return zeroK, zeroV, false, ErrInvariant
	}
	for cursor.index < len(cursor.table.entries) {
		entry := cursor.table.entries[cursor.index]
		cursor.index++
		if entry.live {
			return entry.key, entry.value, true, nil
		}
	}
	return zeroK, zeroV, false, nil
}

func (cursor *StableTableCursor[K, V]) Close() error {
	if cursor == nil || cursor.closed {
		return ErrInvariant
	}
	cursor.closed = true
	cursor.table = nil
	return nil
}

// ForEach visits live entries in insertion order. Structural mutation through
// another table method during visit is rejected; StableTableDelete removes the
// current entry safely.
func (table *StableTable[K, V]) ForEach(visit func(K, V) StableTableAction) (completed bool, err error) {
	if err = table.check(); err != nil {
		return false, err
	}
	if visit == nil {
		return false, ErrDimension
	}
	generation := table.generation
	for _, entry := range table.entries {
		if !entry.live {
			continue
		}
		action := visit(entry.key, entry.value)
		if table.generation != generation {
			return false, ErrInvariant
		}
		switch action {
		case StableTableContinue:
		case StableTableStop:
			return false, nil
		case StableTableDelete:
			if err = table.removeEntry(entry); err != nil {
				return false, err
			}
			generation = table.generation
		default:
			return false, ErrDimension
		}
	}
	return true, nil
}

func (table *StableTable[K, V]) Clone() (*StableTable[K, V], error) {
	if err := table.check(); err != nil {
		return nil, err
	}
	clone, err := NewStableTable[K, V](table.live)
	if err != nil {
		return nil, err
	}
	for _, entry := range table.entries {
		if entry.live {
			if err = clone.AddDirect(entry.key, entry.value); err != nil {
				return nil, err
			}
		}
	}
	return clone, nil
}

func (table *StableTable[K, V]) Clear() error {
	if err := table.check(); err != nil {
		return err
	}
	if err := table.bump(); err != nil {
		return err
	}
	for _, entry := range table.entries {
		entry.live = false
		entry.previousSame = nil
		var zeroK K
		var zeroV V
		entry.key = zeroK
		entry.value = zeroV
	}
	table.entries = nil
	table.index.DeleteAll()
	table.live = 0
	return nil
}

func (table *StableTable[K, V]) Close() error {
	if err := table.Clear(); err != nil {
		return err
	}
	table.index = nil
	table.closed = true
	return nil
}

// StableStringHash reproduces st__strhash's uint32 polynomial hash.
func StableStringHash(value string, modulus int) (int, error) {
	if modulus <= 0 || strings.IndexByte(value, 0) >= 0 {
		return 0, ErrDimension
	}
	var hash uint32
	for index := 0; index < len(value); index++ {
		hash = hash*997 + uint32(value[index])
	}
	return int(hash % uint32(modulus)), nil
}

// StablePooledStringHash reproduces stmm_strhash's two's-complement int32
// polynomial and absolute-value result on the pinned native build.
func StablePooledStringHash(value string, modulus int) (int, error) {
	if modulus <= 0 || strings.IndexByte(value, 0) >= 0 {
		return 0, ErrDimension
	}
	var bits uint32
	for index := 0; index < len(value); index++ {
		bits = bits*997 + uint32(int32(int8(value[index])))
	}
	signed := int64(int32(bits))
	if signed < 0 {
		signed = -signed
	}
	return int(signed % int64(modulus)), nil
}

func StableIntegerHash(value int64, modulus int) (int, error) {
	if modulus <= 0 {
		return 0, ErrDimension
	}
	signed := int64(int32(uint32(value)))
	if signed < 0 {
		signed = -signed
	}
	return int(signed % int64(modulus)), nil
}

func StableAddressHash(address uintptr, modulus int) (int, error) {
	if modulus <= 0 {
		return 0, ErrDimension
	}
	return int((address >> 2) % uintptr(modulus)), nil
}
