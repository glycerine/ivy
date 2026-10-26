package abc

// Packed dynamic bits from src/misc/vec/vecBit.h. Word layout is LSB first.

import "math/bits"

// BitVector owns its packed storage. The zero value is empty and usable.
// Max capacity defaults to one million bits. Methods returning errors leave
// the vector unchanged on failure. No method exposes mutable borrowed storage.
// Start/fill sizes round up to 32 bits as in C; FillExtra and Push sizes do not.
type BitVector struct {
	words []uint32
	size  int
	limit int
}

// NewBitVector reserves rounded capacity without adding entries. maxBits bounds
// rounded allocated capacity; zero selects one million. Negative values reject.
func NewBitVector(capacity, maxBits int) (*BitVector, error) {
	if capacity < 0 || maxBits < 0 {
		return nil, ErrDimension
	}
	v := &BitVector{limit: maxBits}
	if e := v.Grow(capacity); e != nil {
		return nil, e
	}
	return v, nil
}

// NewFilledBitVector starts a word-rounded logical size, matching Start/StartFull.
func NewFilledBitVector(size int, value bool, maxBits int) (*BitVector, error) {
	v, e := NewBitVector(0, maxBits)
	if e != nil {
		return nil, e
	}
	if e = v.Fill(size, value); e != nil {
		return nil, e
	}
	return v, nil
}
func (v *BitVector) Len() int {
	if v == nil {
		return 0
	}
	return v.size
}
func (v *BitVector) Cap() int {
	if v == nil {
		return 0
	}
	return len(v.words) * 32
}
func (v *BitVector) maxBits() int {
	if v.limit == 0 {
		return 1 << 20
	}
	return v.limit
}
func (v *BitVector) Grow(capacity int) error {
	if v == nil || capacity < 0 {
		return ErrDimension
	}
	if capacity <= v.Cap() {
		return nil
	}
	words, e := BitWordCount(capacity, 32)
	if e != nil {
		return e
	}
	if _, e = checkedSize(0, 32, words, v.maxBits()); e != nil {
		return e
	}
	storage := make([]uint32, words)
	copy(storage, v.words)
	v.words = storage
	return nil
}
func (v *BitVector) Entry(index int) (bool, error) {
	if v == nil || index < 0 || index >= v.size {
		return false, ErrDimension
	}
	return v.words[index/32]&(uint32(1)<<uint(index%32)) != 0, nil
}
func (v *BitVector) write(index int, value bool) {
	mask := uint32(1) << uint(index%32)
	if value {
		v.words[index/32] |= mask
	} else {
		v.words[index/32] &^= mask
	}
}
func (v *BitVector) Write(index int, value bool) error {
	if v == nil || index < 0 || index >= v.size {
		return ErrDimension
	}
	v.write(index, value)
	return nil
}

// Add sets an existing entry and returns its previous value.
func (v *BitVector) Add(index int) (bool, error) {
	old, e := v.Entry(index)
	if e != nil {
		return false, e
	}
	v.write(index, true)
	return old, nil
}
func (v *BitVector) Last() (bool, error) {
	if v == nil {
		return false, ErrDimension
	}
	return v.Entry(v.size - 1)
}

// Fill replaces logical contents with a word-rounded number of equal bits.
func (v *BitVector) Fill(size int, value bool) error {
	if v == nil || size < 0 {
		return ErrDimension
	}
	words, e := BitWordCount(size, 32)
	if e != nil {
		return e
	}
	rounded, e := checkedSize(0, 32, words, v.maxBits())
	if e != nil {
		return e
	}
	if e = v.Grow(rounded); e != nil {
		return e
	}
	word := uint32(0)
	if value {
		word = ^uint32(0)
	}
	for i := 0; i < words; i++ {
		v.words[i] = word
	}
	v.size = rounded
	return nil
}

// FillExtra grows to exactly size, filling only newly visible bits. Smaller
// nonnegative sizes are no-ops. Stale bits after shrink/clear never reappear.
// This corrects C's out-of-range word writes and shifts at word boundaries.
func (v *BitVector) FillExtra(size int, value bool) error {
	if v == nil || size < 0 {
		return ErrDimension
	}
	if size <= v.size {
		return nil
	}
	if size > v.Cap() {
		capacity := size
		if v.Cap() > 0 && v.Cap() <= v.maxBits()/2 && size <= 2*v.Cap() {
			capacity = 2 * v.Cap()
		}
		if e := v.Grow(capacity); e != nil {
			return e
		}
	}
	for start := v.size; start < size; {
		end := min(size, (start/32+1)*32)
		width := end - start
		mask := ^uint32(0)
		if width < 32 {
			mask = (uint32(1) << uint(width)) - 1
		}
		mask <<= uint(start % 32)
		if value {
			v.words[start/32] |= mask
		} else {
			v.words[start/32] &^= mask
		}
		start = end
	}
	v.size = size
	return nil
}

// Get extends with zero bits if needed, unlike Entry's checked read.
func (v *BitVector) Get(index int) (bool, error) {
	if v == nil || index < 0 {
		return false, ErrDimension
	}
	size, e := checkedSize(index, 1, 1, v.maxBits())
	if e != nil {
		return false, e
	}
	if e = v.FillExtra(size, false); e != nil {
		return false, e
	}
	return v.Entry(index)
}

// Set extends with zero bits before writing the selected entry.
func (v *BitVector) Set(index int, value bool) error {
	if v == nil || index < 0 {
		return ErrDimension
	}
	size, e := checkedSize(index, 1, 1, v.maxBits())
	if e != nil {
		return e
	}
	if e = v.FillExtra(size, false); e != nil {
		return e
	}
	v.write(index, value)
	return nil
}
func (v *BitVector) Shrink(size int) error {
	if v == nil || size < 0 || size > v.size {
		return ErrDimension
	}
	v.size = size
	return nil
}
func (v *BitVector) Clear() {
	if v != nil {
		v.size = 0
	}
}
func (v *BitVector) Push(value bool) error {
	if v == nil {
		return ErrDimension
	}
	size, e := checkedSize(v.size, 1, 1, v.maxBits())
	if e != nil {
		return e
	}
	if e = v.FillExtra(size, value); e != nil {
		return e
	}
	return nil
}
func (v *BitVector) Pop() (bool, error) {
	old, e := v.Last()
	if e != nil {
		return false, e
	}
	v.size--
	return old, nil
}
func (v *BitVector) Count() int {
	if v == nil {
		return 0
	}
	total := 0
	whole := v.size / 32
	for _, word := range v.words[:whole] {
		total += bits.OnesCount32(word)
	}
	if tail := v.size % 32; tail != 0 {
		total += bits.OnesCount32(v.words[whole] & ((uint32(1) << uint(tail)) - 1))
	}
	return total
}

// Reset clears logical bits while retaining size and capacity.
func (v *BitVector) Reset() {
	if v != nil {
		words, _ := BitWordCount(v.size, 32)
		clear(v.words[:words])
	}
}

// Words returns an independent logical packed snapshot with zero tail padding.
func (v *BitVector) Words() []uint32 {
	if v == nil {
		return nil
	}
	n, _ := BitWordCount(v.size, 32)
	result := append([]uint32(nil), v.words[:n]...)
	if tail := v.size % 32; tail != 0 {
		result[n-1] &= (uint32(1) << uint(tail)) - 1
	}
	return result
}

// Clone supports arbitrary bit sizes, extending C's word-aligned-only Dup.
func (v *BitVector) Clone() *BitVector {
	if v == nil {
		return nil
	}
	return &BitVector{words: v.Words(), size: v.size, limit: v.limit}
}

// ReleaseWords transfers logical packed storage and empties the vector. Tail
// padding and unused backing capacity clear; the configured limit survives.
func (v *BitVector) ReleaseWords() []uint32 {
	if v == nil {
		return nil
	}
	n, _ := BitWordCount(v.size, 32)
	if tail := v.size % 32; tail != 0 {
		v.words[n-1] &= (uint32(1) << uint(tail)) - 1
	}
	clear(v.words[n:])
	result := v.words[:n]
	v.words = nil
	v.size = 0
	return result
}

// StorageBytes reports packed backing payload, excluding Go object/slice headers.
// C's Vec_BitMemory overcounts the payload by treating bit capacity as integers.
func (v *BitVector) StorageBytes() int {
	if v == nil {
		return 0
	}
	return len(v.words) * 4
}
