package abc

// Byte vectors and binary codecs from src/misc/vec/vecStr.h.

import (
	"bytes"
	"encoding/binary"
	"math"
	"slices"
	"strconv"
	"strings"
	"unsafe"
)

type ByteVectorLimits struct {
	MaxBytes int
	Stop     func() bool
}

func normalizeByteVectorLimits(limits ByteVectorLimits) (ByteVectorLimits, error) {
	if limits.MaxBytes < 0 {
		return limits, ErrDimension
	}
	if limits.MaxBytes == 0 {
		limits.MaxBytes = 1 << 30
	}
	return limits, nil
}

// ByteVector keeps a capacity-sized backing slice and a separate logical size.
// Bytes hidden by Shrink or Clear therefore reappear after SetSize, as they do
// in Vec_Str. Public snapshots remain owned by the caller.
type ByteVector struct {
	data   []byte
	size   int
	limits ByteVectorLimits
	closed bool
}

func newByteVector(capacity int, exact bool, limits ByteVectorLimits) (*ByteVector, error) {
	limits, err := normalizeByteVectorLimits(limits)
	if err != nil || capacity < 0 {
		if err != nil {
			return nil, err
		}
		return nil, ErrDimension
	}
	if !exact && capacity > 0 && capacity < 16 {
		capacity = 16
	}
	if capacity > limits.MaxBytes {
		return nil, ErrLimit
	}
	if limits.Stop != nil && limits.Stop() {
		return nil, ErrCancelled
	}
	return &ByteVector{data: make([]byte, capacity), limits: limits}, nil
}

func NewByteVector(capacity int, limits ByteVectorLimits) (*ByteVector, error) {
	return newByteVector(capacity, false, limits)
}

func NewExactByteVector(capacity int, limits ByteVectorLimits) (*ByteVector, error) {
	return newByteVector(capacity, true, limits)
}

func NewZeroByteVector(size int, limits ByteVectorLimits) (*ByteVector, error) {
	vector, err := NewByteVector(size, limits)
	if err != nil {
		return nil, err
	}
	vector.size = size
	return vector, nil
}

// NewByteVectorFromBytes replaces both borrowing and copying array constructors
// with owned storage, preventing later caller mutation from changing the vector.
func NewByteVectorFromBytes(data []byte, limits ByteVectorLimits) (*ByteVector, error) {
	vector, err := NewExactByteVector(len(data), limits)
	if err != nil {
		return nil, err
	}
	copy(vector.data, data)
	vector.size = len(data)
	return vector, nil
}

func (vector *ByteVector) check() error {
	if vector == nil {
		return ErrDimension
	}
	if vector.closed || vector.size < 0 || vector.size > len(vector.data) || len(vector.data) > vector.limits.MaxBytes {
		return ErrInvariant
	}
	return nil
}

func (vector *ByteVector) stopped() bool {
	return vector.limits.Stop != nil && vector.limits.Stop()
}

func (vector *ByteVector) growUnchecked(capacity int) {
	if capacity <= len(vector.data) {
		return
	}
	data := make([]byte, capacity)
	copy(data, vector.data)
	vector.data = data
}

func (vector *ByteVector) Grow(capacity int) error {
	if err := vector.check(); err != nil {
		return err
	}
	if capacity < 0 {
		return ErrDimension
	}
	if capacity <= len(vector.data) {
		return nil
	}
	if capacity > vector.limits.MaxBytes {
		return ErrLimit
	}
	if vector.stopped() {
		return ErrCancelled
	}
	vector.growUnchecked(capacity)
	return nil
}

func (vector *ByteVector) Size() (int, error) {
	if err := vector.check(); err != nil {
		return 0, err
	}
	return vector.size, nil
}

func (vector *ByteVector) Capacity() (int, error) {
	if err := vector.check(); err != nil {
		return 0, err
	}
	return len(vector.data), nil
}

func (vector *ByteVector) SetSize(size int) error {
	if err := vector.check(); err != nil {
		return err
	}
	if size < 0 {
		return ErrDimension
	}
	if size > len(vector.data) {
		return ErrLimit
	}
	if vector.stopped() {
		return ErrCancelled
	}
	vector.size = size
	return nil
}

func (vector *ByteVector) Bytes() ([]byte, error) {
	if err := vector.check(); err != nil {
		return nil, err
	}
	return append([]byte(nil), vector.data[:vector.size]...), nil
}

func (vector *ByteVector) Entry(index int) (byte, error) {
	if err := vector.check(); err != nil {
		return 0, err
	}
	if index < 0 || index >= vector.size {
		return 0, ErrDimension
	}
	return vector.data[index], nil
}

func (vector *ByteVector) Last() (byte, error) {
	if err := vector.check(); err != nil {
		return 0, err
	}
	if vector.size == 0 {
		return 0, ErrDimension
	}
	return vector.data[vector.size-1], nil
}

func (vector *ByteVector) Write(index int, value byte) error {
	if err := vector.check(); err != nil {
		return err
	}
	if index < 0 || index >= vector.size {
		return ErrDimension
	}
	if vector.stopped() {
		return ErrCancelled
	}
	vector.data[index] = value
	return nil
}

func (vector *ByteVector) Fill(size int, value byte) error {
	if err := vector.check(); err != nil {
		return err
	}
	if size < 0 {
		return ErrDimension
	}
	if size > vector.limits.MaxBytes {
		return ErrLimit
	}
	if vector.stopped() {
		return ErrCancelled
	}
	vector.growUnchecked(size)
	for index := 0; index < size; index++ {
		vector.data[index] = value
	}
	vector.size = size
	return nil
}

func (vector *ByteVector) FillExtra(size int, value byte) error {
	if err := vector.check(); err != nil {
		return err
	}
	if size < 0 {
		return ErrDimension
	}
	if size <= vector.size {
		return nil
	}
	if size > vector.limits.MaxBytes {
		return ErrLimit
	}
	capacity := len(vector.data)
	if capacity > math.MaxInt/2 {
		if size > capacity {
			return ErrLimit
		}
	} else if size > 2*capacity {
		capacity = size
	} else if size > capacity {
		capacity *= 2
	}
	if capacity > vector.limits.MaxBytes {
		return ErrLimit
	}
	if vector.stopped() {
		return ErrCancelled
	}
	vector.growUnchecked(capacity)
	for index := vector.size; index < size; index++ {
		vector.data[index] = value
	}
	vector.size = size
	return nil
}

func (vector *ByteVector) Get(index int) (byte, error) {
	if index < 0 {
		return 0, ErrDimension
	}
	if err := vector.FillExtra(index+1, 0); err != nil {
		return 0, err
	}
	return vector.data[index], nil
}

func (vector *ByteVector) Set(index int, value byte) error {
	if index < 0 {
		return ErrDimension
	}
	if err := vector.FillExtra(index+1, 0); err != nil {
		return err
	}
	vector.data[index] = value
	return nil
}

func (vector *ByteVector) Shrink(size int) error {
	if err := vector.check(); err != nil {
		return err
	}
	if size < 0 || size > vector.size {
		return ErrDimension
	}
	if vector.stopped() {
		return ErrCancelled
	}
	vector.size = size
	return nil
}

func (vector *ByteVector) Clear() error {
	return vector.Shrink(0)
}

func (vector *ByteVector) preflightAppend(count int, bufferGrowth bool) (int, error) {
	if count < 0 {
		return 0, ErrDimension
	}
	if vector.size > vector.limits.MaxBytes-count {
		return 0, ErrLimit
	}
	needed := vector.size + count
	capacity := len(vector.data)
	if needed > capacity {
		if bufferGrowth {
			if needed > math.MaxInt/2 {
				return 0, ErrLimit
			}
			capacity = 2 * needed
		} else if capacity < 16 {
			capacity = 16
		} else {
			if capacity > math.MaxInt/2 {
				return 0, ErrLimit
			}
			capacity *= 2
		}
	}
	if capacity > vector.limits.MaxBytes {
		return 0, ErrLimit
	}
	return capacity, nil
}

func (vector *ByteVector) Push(value byte) error {
	if err := vector.check(); err != nil {
		return err
	}
	capacity, err := vector.preflightAppend(1, false)
	if err != nil {
		return err
	}
	if vector.stopped() {
		return ErrCancelled
	}
	vector.growUnchecked(capacity)
	vector.data[vector.size] = value
	vector.size++
	return nil
}

func (vector *ByteVector) PushBuffer(buffer []byte) error {
	if err := vector.check(); err != nil {
		return err
	}
	capacity, err := vector.preflightAppend(len(buffer), true)
	if err != nil {
		return err
	}
	if len(buffer) == 0 {
		return nil
	}
	if vector.stopped() {
		return ErrCancelled
	}
	vector.growUnchecked(capacity)
	copy(vector.data[vector.size:], buffer)
	vector.size += len(buffer)
	return nil
}

func (vector *ByteVector) Pop() (byte, error) {
	if err := vector.check(); err != nil {
		return 0, err
	}
	if vector.size == 0 {
		return 0, ErrDimension
	}
	if vector.stopped() {
		return 0, ErrCancelled
	}
	vector.size--
	return vector.data[vector.size], nil
}

func (vector *ByteVector) Clone() (*ByteVector, error) {
	if err := vector.check(); err != nil {
		return nil, err
	}
	if vector.stopped() {
		return nil, ErrCancelled
	}
	clone, err := NewExactByteVector(len(vector.data), vector.limits)
	if err != nil {
		return nil, err
	}
	copy(clone.data, vector.data[:vector.size])
	clone.size = vector.size
	return clone, nil
}

// Move transfers the backing array into a new vector and leaves the source as
// a reusable zero-capacity vector, matching Vec_StrDupArray.
func (vector *ByteVector) Move() (*ByteVector, error) {
	if err := vector.check(); err != nil {
		return nil, err
	}
	if vector.stopped() {
		return nil, ErrCancelled
	}
	moved := &ByteVector{data: vector.data, size: vector.size, limits: vector.limits}
	vector.data = nil
	vector.size = 0
	return moved, nil
}

// Release transfers the backing slice to the caller and resets the vector.
func (vector *ByteVector) Release() ([]byte, error) {
	if err := vector.check(); err != nil {
		return nil, err
	}
	if vector.stopped() {
		return nil, ErrCancelled
	}
	result := vector.data[:vector.size]
	vector.data = nil
	vector.size = 0
	return result, nil
}

// Erase releases storage but keeps the vector reusable.
func (vector *ByteVector) Erase() error {
	if err := vector.check(); err != nil {
		return err
	}
	if vector.stopped() {
		return ErrCancelled
	}
	clear(vector.data)
	vector.data = nil
	vector.size = 0
	return nil
}

func (vector *ByteVector) MemoryBytes() (int, error) {
	if err := vector.check(); err != nil {
		return 0, err
	}
	return 2*4 + int(unsafe.Sizeof(uintptr(0))) + len(vector.data), nil
}

func cStringPrefix(value string) string {
	if index := strings.IndexByte(value, 0); index >= 0 {
		return value[:index]
	}
	return value
}

func (vector *ByteVector) appendSequential(buffer []byte) error {
	if err := vector.check(); err != nil {
		return err
	}
	if len(buffer) > vector.limits.MaxBytes-vector.size {
		return ErrLimit
	}
	needed := vector.size + len(buffer)
	capacity := len(vector.data)
	for capacity < needed {
		if capacity < 16 {
			capacity = 16
		} else {
			if capacity > vector.limits.MaxBytes/2 {
				capacity = vector.limits.MaxBytes
			} else {
				capacity *= 2
			}
		}
		if capacity == vector.limits.MaxBytes && capacity < needed {
			return ErrLimit
		}
	}
	if capacity > vector.limits.MaxBytes {
		return ErrLimit
	}
	if len(buffer) == 0 {
		return nil
	}
	if vector.stopped() {
		return ErrCancelled
	}
	vector.growUnchecked(capacity)
	copy(vector.data[vector.size:], buffer)
	vector.size = needed
	return nil
}

func (vector *ByteVector) AppendString(value string) error {
	return vector.appendSequential([]byte(cStringPrefix(value)))
}

// AppendFormatted stores text already formatted by the caller and returns its
// starting offset. It replaces Vec_StrPrintF without a runtime-interface API.
func (vector *ByteVector) AppendFormatted(value string) (int, error) {
	if err := vector.check(); err != nil {
		return 0, err
	}
	start := vector.size
	if len(value) > vector.limits.MaxBytes-vector.size {
		return 0, ErrLimit
	}
	capacity := vector.size + 1000
	if len(value) >= 1000 {
		if len(value) > vector.limits.MaxBytes-vector.size-1000 {
			return 0, ErrLimit
		}
		capacity = vector.size + len(value) + 1000
	}
	if capacity > vector.limits.MaxBytes {
		return 0, ErrLimit
	}
	if capacity < len(vector.data) {
		capacity = len(vector.data)
	}
	if vector.stopped() {
		return 0, ErrCancelled
	}
	vector.growUnchecked(capacity)
	copy(vector.data[vector.size:], value)
	vector.size += len(value)
	return start, nil
}

func (vector *ByteVector) CopyString(value string) error {
	value = cStringPrefix(value)
	if err := vector.check(); err != nil {
		return err
	}
	needed := len(value) + 1
	if needed > vector.limits.MaxBytes {
		return ErrLimit
	}
	capacity := len(vector.data)
	for capacity < needed {
		if capacity < 16 {
			capacity = 16
		} else if capacity <= vector.limits.MaxBytes/2 {
			capacity *= 2
		} else {
			capacity = vector.limits.MaxBytes
		}
		if capacity == vector.limits.MaxBytes && capacity < needed {
			return ErrLimit
		}
	}
	if capacity > vector.limits.MaxBytes {
		return ErrLimit
	}
	if vector.stopped() {
		return ErrCancelled
	}
	vector.growUnchecked(capacity)
	copy(vector.data, value)
	vector.data[len(value)] = 0
	vector.size = len(value) + 1
	return nil
}

func (vector *ByteVector) AppendInt(value int) error {
	if value < math.MinInt32 || value > math.MaxInt32 {
		return ErrLimit
	}
	return vector.appendSequential(strconv.AppendInt(nil, int64(value), 10))
}

func (vector *ByteVector) AppendPaddedInt(value, digits int) error {
	if value < math.MinInt32 || value > math.MaxInt32 || digits < 0 {
		if digits < 0 {
			return ErrDimension
		}
		return ErrLimit
	}
	if value == 0 {
		return vector.appendSequential(bytes.Repeat([]byte{'0'}, digits))
	}
	negative := value < 0
	magnitude := int64(value)
	if magnitude < 0 {
		magnitude = -magnitude
	}
	text := strconv.FormatInt(magnitude, 10)
	width := digits
	if negative && width > 0 {
		width--
	}
	zeros := width - len(text)
	if zeros < 0 {
		zeros = 0
	}
	needed := zeros + len(text)
	if negative {
		needed++
	}
	buffer := make([]byte, 0, needed)
	if negative {
		buffer = append(buffer, '-')
	}
	for range zeros {
		buffer = append(buffer, '0')
	}
	buffer = append(buffer, text...)
	return vector.appendSequential(buffer)
}

func (vector *ByteVector) IntText() (string, error) {
	if err := vector.check(); err != nil {
		return "", err
	}
	var output strings.Builder
	output.WriteString("Vector has ")
	output.WriteString(strconv.Itoa(vector.size))
	output.WriteString(" entries: {")
	for index := 0; index < vector.size; index++ {
		output.WriteByte(' ')
		output.WriteString(strconv.Itoa(int(int8(vector.data[index]))))
	}
	output.WriteString(" }\n")
	return output.String(), nil
}

func (vector *ByteVector) Reverse() error {
	if err := vector.check(); err != nil {
		return err
	}
	if vector.stopped() {
		return ErrCancelled
	}
	for left, right := 0, vector.size-1; left < right; left, right = left+1, right-1 {
		vector.data[left], vector.data[right] = vector.data[right], vector.data[left]
	}
	return nil
}

func (vector *ByteVector) SumSigned() (int, error) {
	if err := vector.check(); err != nil {
		return 0, err
	}
	sum := 0
	for index := 0; index < vector.size; index++ {
		sum += int(int8(vector.data[index]))
	}
	return sum, nil
}

func (vector *ByteVector) Count(value byte) (int, error) {
	if err := vector.check(); err != nil {
		return 0, err
	}
	return bytes.Count(vector.data[:vector.size], []byte{value}), nil
}

func (vector *ByteVector) CountSignedLarger(value int8) (int, error) {
	return vector.countSigned(value, 1)
}

func (vector *ByteVector) CountSignedSmaller(value int8) (int, error) {
	return vector.countSigned(value, -1)
}

func (vector *ByteVector) countSigned(value int8, direction int) (int, error) {
	if err := vector.check(); err != nil {
		return 0, err
	}
	count := 0
	for index := 0; index < vector.size; index++ {
		entry := int8(vector.data[index])
		if direction > 0 && entry > value || direction < 0 && entry < value {
			count++
		}
	}
	return count, nil
}

func (vector *ByteVector) countLiteralVariable(variable, relation int) (int, error) {
	if err := vector.check(); err != nil {
		return 0, err
	}
	if variable < math.MinInt8 || variable > math.MaxInt8 {
		return 0, ErrDimension
	}
	count := 0
	for index := 0; index < vector.size; index++ {
		entryVariable := int(int8(vector.data[index])) >> 1
		if relation == 0 && entryVariable == variable || relation > 0 && entryVariable > variable || relation < 0 && entryVariable < variable {
			count++
		}
	}
	return count, nil
}

func (vector *ByteVector) CountLiteralVariable(variable int) (int, error) {
	return vector.countLiteralVariable(variable, 0)
}

func (vector *ByteVector) CountLiteralVariableLarger(variable int) (int, error) {
	return vector.countLiteralVariable(variable, 1)
}

func (vector *ByteVector) CountLiteralVariableSmaller(variable int) (int, error) {
	return vector.countLiteralVariable(variable, -1)
}

func (vector *ByteVector) Equal(other *ByteVector) (bool, error) {
	if err := vector.check(); err != nil {
		return false, err
	}
	if err := other.check(); err != nil {
		return false, err
	}
	return bytes.Equal(vector.data[:vector.size], other.data[:other.size]), nil
}

func (vector *ByteVector) SortSigned(reverse bool) error {
	if err := vector.check(); err != nil {
		return err
	}
	if vector.stopped() {
		return ErrCancelled
	}
	slices.SortStableFunc(vector.data[:vector.size], func(left, right byte) int {
		comparison := int(int8(left)) - int(int8(right))
		if reverse {
			return -comparison
		}
		return comparison
	})
	return nil
}

func CompareByteVectors(first, second *ByteVector) (int, error) {
	if first == nil || second == nil {
		if first == nil && second == nil {
			return 0, nil
		}
		if first != nil {
			return 1, nil
		}
		return -1, nil
	}
	if err := first.check(); err != nil {
		return 0, err
	}
	if err := second.check(); err != nil {
		return 0, err
	}
	if first.size != second.size {
		return first.size - second.size, nil
	}
	for index := 0; index < first.size; index++ {
		if first.data[index] != second.data[index] {
			return int(first.data[index]) - int(second.data[index]), nil
		}
	}
	return 0, nil
}

func (vector *ByteVector) appendFixed(data []byte) error {
	return vector.appendSequential(data)
}

func (vector *ByteVector) PutIntNetwork(value int) error {
	if value < math.MinInt32 || value > math.MaxInt32 {
		return ErrLimit
	}
	var encoded [4]byte
	binary.BigEndian.PutUint32(encoded[:], uint32(int32(value)))
	return vector.appendFixed(encoded[:])
}

func (vector *ByteVector) PutVarInt(value int) error {
	if value < 0 || value > math.MaxInt32 {
		return ErrDimension
	}
	var encoded [5]byte
	count := 0
	remaining := uint32(value)
	for remaining >= 0x80 {
		encoded[count] = byte(remaining) | 0x80
		remaining >>= 7
		count++
	}
	encoded[count] = byte(remaining)
	return vector.appendFixed(encoded[:count+1])
}

func (vector *ByteVector) PutWord(value uint64) error {
	var encoded [8]byte
	binary.LittleEndian.PutUint64(encoded[:], value)
	return vector.appendFixed(encoded[:])
}

func (vector *ByteVector) PutFloat(value float32) error {
	var encoded [4]byte
	binary.LittleEndian.PutUint32(encoded[:], math.Float32bits(value))
	return vector.appendFixed(encoded[:])
}

func (vector *ByteVector) PutDouble(value float64) error {
	var encoded [8]byte
	binary.LittleEndian.PutUint64(encoded[:], math.Float64bits(value))
	return vector.appendFixed(encoded[:])
}

func (vector *ByteVector) PutString(value string) error {
	value = cStringPrefix(value)
	buffer := make([]byte, len(value)+1)
	copy(buffer, value)
	return vector.appendSequential(buffer)
}

func (vector *ByteVector) PutByte(value byte) error {
	return vector.Push(value)
}

func (vector *ByteVector) read(position, count int) ([]byte, int, error) {
	if err := vector.check(); err != nil {
		return nil, position, err
	}
	if position < 0 || count < 0 || position > vector.size || count > vector.size-position {
		return nil, position, ErrDimension
	}
	return vector.data[position : position+count], position + count, nil
}

func (vector *ByteVector) GetIntNetwork(position int) (int, int, error) {
	data, next, err := vector.read(position, 4)
	if err != nil {
		return 0, position, err
	}
	return int(int32(binary.BigEndian.Uint32(data))), next, nil
}

func (vector *ByteVector) GetVarInt(position int) (int, int, error) {
	if err := vector.check(); err != nil {
		return 0, position, err
	}
	if position < 0 || position >= vector.size {
		return 0, position, ErrDimension
	}
	var value uint32
	for index := 0; index < 5; index++ {
		if position+index >= vector.size {
			return 0, position, ErrDimension
		}
		entry := vector.data[position+index]
		if index == 4 && entry > 7 {
			return 0, position, ErrLimit
		}
		value |= uint32(entry&0x7f) << (7 * index)
		if entry&0x80 == 0 {
			if value > math.MaxInt32 {
				return 0, position, ErrLimit
			}
			return int(value), position + index + 1, nil
		}
	}
	return 0, position, ErrLimit
}

func (vector *ByteVector) GetWord(position int) (uint64, int, error) {
	data, next, err := vector.read(position, 8)
	if err != nil {
		return 0, position, err
	}
	return binary.LittleEndian.Uint64(data), next, nil
}

func (vector *ByteVector) GetFloat(position int) (float32, int, error) {
	data, next, err := vector.read(position, 4)
	if err != nil {
		return 0, position, err
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(data)), next, nil
}

func (vector *ByteVector) GetDouble(position int) (float64, int, error) {
	data, next, err := vector.read(position, 8)
	if err != nil {
		return 0, position, err
	}
	return math.Float64frombits(binary.LittleEndian.Uint64(data)), next, nil
}

func (vector *ByteVector) GetString(position int) (string, int, error) {
	if err := vector.check(); err != nil {
		return "", position, err
	}
	if position < 0 || position >= vector.size {
		return "", position, ErrDimension
	}
	end := bytes.IndexByte(vector.data[position:vector.size], 0)
	if end < 0 {
		return "", position, ErrDimension
	}
	return string(vector.data[position : position+end]), position + end + 1, nil
}

func (vector *ByteVector) GetByte(position int) (byte, int, error) {
	data, next, err := vector.read(position, 1)
	if err != nil {
		return 0, position, err
	}
	return data[0], next, nil
}

func (vector *ByteVector) Validate() error {
	return vector.check()
}

func (vector *ByteVector) Close() error {
	if err := vector.check(); err != nil {
		return err
	}
	clear(vector.data)
	vector.data = nil
	vector.size = 0
	vector.closed = true
	return nil
}
