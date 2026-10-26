package abc

// Dynamic float32 storage from src/misc/vec/vecFlt.h.

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

// FloatVector owns float32 storage. Zero values work; limits default to a
// million entries. Checked failures leave contents/size/capacity unchanged.
// NaN payloads survive storage/copy operations; IEEE comparisons match C.
type FloatVector struct {
	array       []float32
	size, limit int
}

// NewFloatVector uses C's minimum allocation of 16 for positive capacities.
func NewFloatVector(capacity, maxEntries int) (*FloatVector, error) {
	if capacity < 0 || maxEntries < 0 {
		return nil, ErrDimension
	}
	if capacity > 0 && capacity < 16 {
		capacity = 16
	}
	return NewFloatVectorExact(capacity, maxEntries)
}
func NewFloatVectorExact(capacity, maxEntries int) (*FloatVector, error) {
	if capacity < 0 || maxEntries < 0 {
		return nil, ErrDimension
	}
	v := &FloatVector{limit: maxEntries}
	if e := v.Grow(capacity); e != nil {
		return nil, e
	}
	return v, nil
}
func NewFilledFloatVector(size int, value float32, maxEntries int) (*FloatVector, error) {
	v, e := NewFloatVector(size, maxEntries)
	if e != nil {
		return nil, e
	}
	if e = v.Fill(size, value); e != nil {
		return nil, e
	}
	return v, nil
}

// NewFullFloatVector matches StartFull's all-one bytes (NaNs, not numeric ones).
func NewFullFloatVector(size, maxEntries int) (*FloatVector, error) {
	return NewFilledFloatVector(size, math.Float32frombits(0xffffffff), maxEntries)
}

// FloatVectorFrom copies input values into exact-sized owned storage.
func FloatVectorFrom(values []float32, maxEntries int) (*FloatVector, error) {
	v, e := NewFloatVectorExact(len(values), maxEntries)
	if e != nil {
		return nil, e
	}
	copy(v.array, values)
	v.size = len(values)
	return v, nil
}
func (v *FloatVector) maxEntries() int {
	if v.limit == 0 {
		return 1 << 20
	}
	return v.limit
}
func (v *FloatVector) Len() int {
	if v == nil {
		return 0
	}
	return v.size
}
func (v *FloatVector) Cap() int {
	if v == nil {
		return 0
	}
	return len(v.array)
}
func (v *FloatVector) StorageBytes() int {
	if v == nil {
		return 0
	}
	return len(v.array) * 4
}
func (v *FloatVector) Grow(capacity int) error {
	if v == nil || capacity < 0 {
		return ErrDimension
	}
	if capacity <= len(v.array) {
		return nil
	}
	if capacity > v.maxEntries() {
		return ErrLimit
	}
	if _, e := checkedSize(0, 4, capacity, int(^uint(0)>>1)); e != nil {
		return e
	}
	array := make([]float32, capacity)
	copy(array, v.array[:v.size])
	v.array = array
	return nil
}
func (v *FloatVector) Entry(index int) (float32, error) {
	if v == nil || index < 0 || index >= v.size {
		return 0, ErrDimension
	}
	return v.array[index], nil
}
func (v *FloatVector) Write(index int, value float32) error {
	if v == nil || index < 0 || index >= v.size {
		return ErrDimension
	}
	v.array[index] = value
	return nil
}
func (v *FloatVector) Add(index int, addition float32) error {
	value, e := v.Entry(index)
	if e != nil {
		return e
	}
	v.array[index] = float32(value + addition)
	return nil
}

// Update replaces an entry only when the new value is numerically larger.
func (v *FloatVector) Update(index int, value float32) error {
	old, e := v.Entry(index)
	if e != nil {
		return e
	}
	if old < value {
		v.array[index] = value
	}
	return nil
}
func (v *FloatVector) Last() (float32, error) {
	if v == nil {
		return 0, ErrDimension
	}
	return v.Entry(v.size - 1)
}
func (v *FloatVector) Fill(size int, value float32) error {
	if v == nil || size < 0 {
		return ErrDimension
	}
	if e := v.Grow(size); e != nil {
		return e
	}
	for i := 0; i < size; i++ {
		v.array[i] = value
	}
	v.size = size
	return nil
}
func (v *FloatVector) FillExtra(size int, value float32) error {
	if v == nil || size < 0 {
		return ErrDimension
	}
	if size <= v.size {
		return nil
	}
	capacity := size
	if size > v.Cap() && v.Cap() > 0 && v.Cap() <= v.maxEntries()/2 && size <= 2*v.Cap() {
		capacity = 2 * v.Cap()
	}
	if e := v.Grow(capacity); e != nil {
		return e
	}
	for i := v.size; i < size; i++ {
		v.array[i] = value
	}
	v.size = size
	return nil
}
func (v *FloatVector) Shrink(size int) error {
	if v == nil || size < 0 || size > v.size {
		return ErrDimension
	}
	v.size = size
	return nil
}
func (v *FloatVector) Clear() {
	if v != nil {
		v.size = 0
	}
}

// Erase releases storage while retaining the allocation limit (Zero/Erase in C).
func (v *FloatVector) Erase() {
	if v != nil {
		v.array = nil
		v.size = 0
	}
}
func (v *FloatVector) Push(value float32) error {
	if v == nil {
		return ErrDimension
	}
	size, e := checkedSize(v.size, 1, 1, v.maxEntries())
	if e != nil {
		return e
	}
	if size > v.Cap() {
		capacity := min(16, v.maxEntries())
		if v.Cap() >= 16 {
			capacity = v.maxEntries()
			if v.Cap() <= v.maxEntries()/2 {
				capacity = 2 * v.Cap()
			}
		}
		capacity = max(capacity, size)
		if e = v.Grow(capacity); e != nil {
			return e
		}
	}
	v.array[v.size] = value
	v.size = size
	return nil
}

// PushOrdered inserts after equal values in an ascending vector. NaNs in the
// input or vector reject atomically; callers supply an already sorted vector.
func (v *FloatVector) PushOrdered(value float32) error {
	if v == nil || math.IsNaN(float64(value)) {
		return ErrDimension
	}
	for _, x := range v.array[:v.size] {
		if math.IsNaN(float64(x)) {
			return ErrDimension
		}
	}
	if e := v.Push(value); e != nil {
		return e
	}
	i := v.size - 1
	for i > 0 && v.array[i-1] > value {
		v.array[i] = v.array[i-1]
		i--
	}
	v.array[i] = value
	return nil
}

// PushUnique returns true if an equal entry already existed, matching C.
// IEEE NaNs compare unequal, so each NaN request appends a new entry.
func (v *FloatVector) PushUnique(value float32) (bool, error) {
	if v == nil {
		return false, ErrDimension
	}
	if v.Find(value) >= 0 {
		return true, nil
	}
	return false, v.Push(value)
}
func (v *FloatVector) Pop() (float32, error) {
	value, e := v.Last()
	if e != nil {
		return 0, e
	}
	v.size--
	return value, nil
}
func (v *FloatVector) Find(value float32) int {
	if v == nil {
		return -1
	}
	for i, x := range v.array[:v.size] {
		if x == value {
			return i
		}
	}
	return -1
}
func (v *FloatVector) Remove(value float32) bool {
	index := v.Find(value)
	if index < 0 {
		return false
	}
	copy(v.array[index:], v.array[index+1:v.size])
	v.size--
	return true
}

// Min/Max follow C's first-entry scan: an initial NaN survives; later NaNs do
// not replace a numeric best. Empty vectors return (0,false).
func (v *FloatVector) Min() (float32, bool) {
	if v.Len() == 0 {
		return 0, false
	}
	best := v.array[0]
	for _, x := range v.array[1:v.size] {
		if best > x {
			best = x
		}
	}
	return best, true
}
func (v *FloatVector) Max() (float32, bool) {
	if v.Len() == 0 {
		return 0, false
	}
	best := v.array[0]
	for _, x := range v.array[1:v.size] {
		if best < x {
			best = x
		}
	}
	return best, true
}

// Equal uses numeric equality (signed zeros equal, NaNs unequal), as in C.
func (v *FloatVector) Equal(other *FloatVector) bool {
	if v.Len() != other.Len() {
		return false
	}
	for i := 0; i < v.Len(); i++ {
		if v.array[i] != other.array[i] {
			return false
		}
	}
	return true
}

// Sort orders numbers stably, retaining signed-zero encounter order. C leaves
// equal-key order unspecified. NaNs reject before mutation, avoiding C's
// non-transitive comparator. Infinities sort normally.
func (v *FloatVector) Sort(reverse bool) error {
	if v == nil {
		return ErrDimension
	}
	for _, x := range v.array[:v.size] {
		if math.IsNaN(float64(x)) {
			return ErrDimension
		}
	}
	sort.SliceStable(v.array[:v.size], func(i, j int) bool {
		if reverse {
			return v.array[i] > v.array[j]
		}
		return v.array[i] < v.array[j]
	})
	return nil
}
func (v *FloatVector) Values() []float32 {
	if v == nil {
		return nil
	}
	return append([]float32(nil), v.array[:v.size]...)
}
func (v *FloatVector) Clone() *FloatVector {
	if v == nil {
		return nil
	}
	array := make([]float32, len(v.array))
	copy(array, v.array[:v.size])
	return &FloatVector{array, v.size, v.limit}
}

// Move transfers allocation/contents into a new vector and empties the source.
func (v *FloatVector) Move() *FloatVector {
	if v == nil {
		return nil
	}
	result := &FloatVector{v.array, v.size, v.limit}
	v.array = nil
	v.size = 0
	return result
}
func (v *FloatVector) ReleaseValues() []float32 {
	if v == nil {
		return nil
	}
	clear(v.array[v.size:])
	result := v.array[:v.size]
	v.array = nil
	v.size = 0
	return result
}

// Format returns Vec_FltPrint's deterministic finite-value layout instead of
// writing stdout. Nonfinite spellings follow Go rather than the C library's
// platform-dependent printf spellings.
func (v *FloatVector) Format() string {
	var output strings.Builder
	fmt.Fprintf(&output, "Vector has %d entries: {", v.Len())
	if v != nil {
		for _, value := range v.array[:v.size] {
			fmt.Fprintf(&output, " %f", value)
		}
	}
	output.WriteString(" }\n")
	return output.String()
}
