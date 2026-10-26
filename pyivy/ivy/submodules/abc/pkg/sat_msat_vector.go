package abc

// Owned integer vector from src/sat/msat/msatVec.c.

import (
	"slices"
	"sort"
)

const defaultMsatVectorMaximum = 1 << 20

// MsatIntVector owns a bounded integer array. Use Values for copied access;
// TakeMsatIntVector and Release are the explicit ownership-transfer operations.
type MsatIntVector struct {
	values  []int
	size    int
	maximum int
	closed  bool
}

func msatVectorMaximum(maximum int) (int, error) {
	if maximum < 0 {
		return 0, ErrDimension
	}
	if maximum == 0 {
		maximum = defaultMsatVectorMaximum
	}
	return maximum, nil
}

func NewMsatIntVector(capacity, maximum int) (*MsatIntVector, error) {
	maximum, err := msatVectorMaximum(maximum)
	if err != nil || capacity < 0 {
		return nil, ErrDimension
	}
	if capacity > 0 && capacity < 16 {
		capacity = 16
	}
	if capacity > maximum {
		return nil, ErrLimit
	}
	return &MsatIntVector{values: make([]int, capacity), maximum: maximum}, nil
}

// TakeMsatIntVector transfers values into a vector and clears the caller's
// slice header. The caller must not retain another alias to the backing array.
func TakeMsatIntVector(values *[]int, maximum int) (*MsatIntVector, error) {
	if values == nil {
		return nil, ErrDimension
	}
	maximum, err := msatVectorMaximum(maximum)
	if err != nil {
		return nil, err
	}
	if len(*values) > maximum {
		return nil, ErrLimit
	}
	owned := (*values)[:len(*values):len(*values)]
	*values = nil
	return &MsatIntVector{values: owned, size: len(owned), maximum: maximum}, nil
}

func CopyMsatIntVector(values []int, maximum int) (*MsatIntVector, error) {
	copyValues := slices.Clone(values)
	return TakeMsatIntVector(&copyValues, maximum)
}

func (v *MsatIntVector) checkLive() error {
	if v == nil {
		return ErrDimension
	}
	if v.closed || v.maximum <= 0 || v.size < 0 || v.size > len(v.values) || len(v.values) > v.maximum {
		return ErrInvariant
	}
	return nil
}

func (v *MsatIntVector) Len() int {
	if v == nil || v.closed {
		return 0
	}
	return v.size
}

func (v *MsatIntVector) Capacity() int {
	if v == nil || v.closed {
		return 0
	}
	return len(v.values)
}

func (v *MsatIntVector) Values() []int {
	if v == nil || v.closed || v.size < 0 || v.size > len(v.values) {
		return nil
	}
	return slices.Clone(v.values[:v.size])
}

func (v *MsatIntVector) Clone() (*MsatIntVector, error) {
	if err := v.checkLive(); err != nil {
		return nil, err
	}
	result := &MsatIntVector{values: make([]int, len(v.values)), size: v.size, maximum: v.maximum}
	copy(result.values, v.values[:v.size])
	return result, nil
}

// Transfer moves the backing array into a new vector and leaves v empty but
// usable, matching Msat_IntVecDupArray.
func (v *MsatIntVector) Transfer() (*MsatIntVector, error) {
	if err := v.checkLive(); err != nil {
		return nil, err
	}
	result := &MsatIntVector{values: v.values, size: v.size, maximum: v.maximum}
	v.values, v.size = nil, 0
	return result, nil
}

// Release transfers the active prefix to the caller and leaves v empty.
func (v *MsatIntVector) Release() ([]int, error) {
	if err := v.checkLive(); err != nil {
		return nil, err
	}
	result := v.values[:v.size:v.size]
	v.values, v.size = nil, 0
	return result, nil
}

func (v *MsatIntVector) Grow(capacity int) error {
	if err := v.checkLive(); err != nil {
		return err
	}
	if capacity < 0 {
		return ErrDimension
	}
	if capacity <= len(v.values) {
		return nil
	}
	if capacity > v.maximum {
		return ErrLimit
	}
	grown := make([]int, capacity)
	copy(grown, v.values[:v.size])
	v.values = grown
	return nil
}

func (v *MsatIntVector) Fill(size, entry int) error {
	if size < 0 {
		return ErrDimension
	}
	if err := v.Grow(size); err != nil {
		return err
	}
	for index := 0; index < size; index++ {
		v.values[index] = entry
	}
	if size < v.size {
		clear(v.values[size:v.size])
	}
	v.size = size
	return nil
}

func (v *MsatIntVector) At(index int) (int, error) {
	if err := v.checkLive(); err != nil {
		return 0, err
	}
	if index < 0 || index >= v.size {
		return 0, ErrDimension
	}
	return v.values[index], nil
}

func (v *MsatIntVector) Last() (int, error) {
	if err := v.checkLive(); err != nil {
		return 0, err
	}
	if v.size == 0 {
		return 0, ErrDimension
	}
	return v.values[v.size-1], nil
}

func (v *MsatIntVector) Set(index, entry int) error {
	if err := v.checkLive(); err != nil {
		return err
	}
	if index < 0 || index >= v.size {
		return ErrDimension
	}
	v.values[index] = entry
	return nil
}

func (v *MsatIntVector) Shrink(size int) error {
	if err := v.checkLive(); err != nil {
		return err
	}
	if size < 0 || size > v.size {
		return ErrDimension
	}
	clear(v.values[size:v.size])
	v.size = size
	return nil
}

func (v *MsatIntVector) Clear() error { return v.Shrink(0) }

func (v *MsatIntVector) Push(entry int) error {
	if err := v.checkLive(); err != nil {
		return err
	}
	if v.size == len(v.values) {
		capacity := min(16, v.maximum)
		if len(v.values) >= 16 {
			if len(v.values) > v.maximum/2 {
				capacity = v.maximum
			} else {
				capacity = 2 * len(v.values)
			}
		}
		if capacity <= len(v.values) {
			return ErrLimit
		}
		if err := v.Grow(capacity); err != nil {
			return err
		}
	}
	v.values[v.size] = entry
	v.size++
	return nil
}

// PushUnique returns true when entry already existed.
func (v *MsatIntVector) PushUnique(entry int) (bool, error) {
	if err := v.checkLive(); err != nil {
		return false, err
	}
	for _, value := range v.values[:v.size] {
		if value == entry {
			return true, nil
		}
	}
	return false, v.Push(entry)
}

func (v *MsatIntVector) PushUniqueOrdered(entry int, increasing bool) (bool, error) {
	found, err := v.PushUnique(entry)
	if err != nil {
		return false, err
	}
	for index := v.size - 1; index > 0; index-- {
		current, previous := v.values[index], v.values[index-1]
		if increasing && current >= previous || !increasing && current <= previous {
			break
		}
		v.values[index], v.values[index-1] = previous, current
	}
	return found, nil
}

func (v *MsatIntVector) Pop() (int, error) {
	if err := v.checkLive(); err != nil {
		return 0, err
	}
	if v.size == 0 {
		return 0, ErrDimension
	}
	v.size--
	result := v.values[v.size]
	v.values[v.size] = 0
	return result, nil
}

func (v *MsatIntVector) Sort(reverse bool) error {
	if err := v.checkLive(); err != nil {
		return err
	}
	if reverse {
		sort.Sort(sort.Reverse(sort.IntSlice(v.values[:v.size])))
	} else {
		slices.Sort(v.values[:v.size])
	}
	return nil
}

func (v *MsatIntVector) Close() {
	if v == nil || v.closed {
		return
	}
	clear(v.values)
	v.values, v.size = nil, 0
	v.closed = true
}
