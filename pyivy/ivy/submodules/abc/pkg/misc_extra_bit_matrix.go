package abc

// Packed relation matrix from src/misc/extra/extraUtilBitMatrix.c.

import (
	"math"
	"math/bits"
	"strings"
)

type ExtraBitMatrixLimits struct {
	MaxWords, MaxOperations, MaxBytes int
	Stop                              func() bool
}

type ExtraBitMatrixStatistics struct {
	Lookups, Inserts, Deletes uint64
}

type ExtraBitMatrix struct {
	size, words int
	data        []uint32
	statistics  ExtraBitMatrixStatistics
	closed      bool
}

func normalizeExtraBitMatrixLimits(limits ExtraBitMatrixLimits) (ExtraBitMatrixLimits, error) {
	if limits.MaxWords < 0 || limits.MaxOperations < 0 || limits.MaxBytes < 0 {
		return limits, ErrDimension
	}
	if limits.MaxWords == 0 {
		limits.MaxWords = 1 << 26
	}
	if limits.MaxOperations == 0 {
		limits.MaxOperations = 1 << 28
	}
	if limits.MaxBytes == 0 {
		limits.MaxBytes = 16 << 20
	}
	return limits, nil
}

func NewExtraBitMatrix(size int, limits ExtraBitMatrixLimits) (*ExtraBitMatrix, error) {
	limits, err := normalizeExtraBitMatrixLimits(limits)
	if err != nil || size < 0 {
		if err != nil {
			return nil, err
		}
		return nil, ErrDimension
	}
	if limits.Stop != nil && limits.Stop() {
		return nil, ErrCancelled
	}
	words64 := (int64(size) + 31) / 32
	total64 := int64(size) * words64
	if words64 > int64(math.MaxInt) || total64 > int64(math.MaxInt) || total64 > int64(limits.MaxWords) {
		return nil, ErrLimit
	}
	return &ExtraBitMatrix{size: size, words: int(words64), data: make([]uint32, int(total64))}, nil
}

func (matrix *ExtraBitMatrix) check() error {
	if matrix == nil {
		return ErrDimension
	}
	if matrix.closed || matrix.size < 0 || matrix.words != (matrix.size+31)/32 || len(matrix.data) != matrix.size*matrix.words {
		return ErrInvariant
	}
	return nil
}

func (matrix *ExtraBitMatrix) checkIndex(index int) error {
	if err := matrix.check(); err != nil {
		return err
	}
	if index < 0 || index >= matrix.size {
		return ErrDimension
	}
	return nil
}

func (matrix *ExtraBitMatrix) Size() (int, error) {
	if err := matrix.check(); err != nil {
		return 0, err
	}
	return matrix.size, nil
}

func (matrix *ExtraBitMatrix) Statistics() (ExtraBitMatrixStatistics, error) {
	if err := matrix.check(); err != nil {
		return ExtraBitMatrixStatistics{}, err
	}
	return matrix.statistics, nil
}

func (matrix *ExtraBitMatrix) SnapshotRows() ([][]uint32, error) {
	if err := matrix.check(); err != nil {
		return nil, err
	}
	rows := make([][]uint32, matrix.size)
	for row := range rows {
		rows[row] = append([]uint32(nil), matrix.data[row*matrix.words:(row+1)*matrix.words]...)
	}
	return rows, nil
}

func (matrix *ExtraBitMatrix) Clear(limits ExtraBitMatrixLimits) error {
	if err := matrix.check(); err != nil {
		return err
	}
	limits, err := normalizeExtraBitMatrixLimits(limits)
	if err != nil {
		return err
	}
	if len(matrix.data) > limits.MaxOperations {
		return ErrLimit
	}
	if limits.Stop != nil && limits.Stop() {
		return ErrCancelled
	}
	clear(matrix.data)
	return nil
}

func (matrix *ExtraBitMatrix) Close() error {
	if err := matrix.check(); err != nil {
		return err
	}
	matrix.data = nil
	matrix.closed = true
	return nil
}

func (matrix *ExtraBitMatrix) bit(row, column int) bool {
	return matrix.data[row*matrix.words+column/32]&(uint32(1)<<uint(column&31)) != 0
}

func (matrix *ExtraBitMatrix) setBit(row, column int, value bool) {
	word := &matrix.data[row*matrix.words+column/32]
	mask := uint32(1) << uint(column&31)
	if value {
		*word |= mask
	} else {
		*word &^= mask
	}
}

func (matrix *ExtraBitMatrix) relationIndices(first, second int, upper bool) (int, int, error) {
	if err := matrix.checkIndex(first); err != nil {
		return 0, 0, err
	}
	if err := matrix.checkIndex(second); err != nil {
		return 0, 0, err
	}
	if upper {
		if first > second {
			first, second = second, first
		}
		return first, second, nil
	}
	if first < second {
		first, second = second, first
	}
	return first, second, nil
}

func (matrix *ExtraBitMatrix) setRelation(first, second int, upper, value bool) error {
	row, column, err := matrix.relationIndices(first, second, upper)
	if err != nil {
		return err
	}
	matrix.setBit(row, column, value)
	if value {
		matrix.statistics.Inserts++
	} else {
		matrix.statistics.Deletes++
	}
	return nil
}

func (matrix *ExtraBitMatrix) lookupRelation(first, second int, upper bool) (bool, error) {
	row, column, err := matrix.relationIndices(first, second, upper)
	if err != nil {
		return false, err
	}
	matrix.statistics.Lookups++
	return matrix.bit(row, column), nil
}

func (matrix *ExtraBitMatrix) InsertUpper(first, second int) error {
	return matrix.setRelation(first, second, true, true)
}

func (matrix *ExtraBitMatrix) LookupUpper(first, second int) (bool, error) {
	return matrix.lookupRelation(first, second, true)
}

func (matrix *ExtraBitMatrix) DeleteUpper(first, second int) error {
	return matrix.setRelation(first, second, true, false)
}

func (matrix *ExtraBitMatrix) InsertLower(first, second int) error {
	return matrix.setRelation(first, second, false, true)
}

func (matrix *ExtraBitMatrix) LookupLower(first, second int) (bool, error) {
	return matrix.lookupRelation(first, second, false)
}

func (matrix *ExtraBitMatrix) DeleteLower(first, second int) error {
	return matrix.setRelation(first, second, false, false)
}

func (matrix *ExtraBitMatrix) OrRow(row int, words []uint32, limits ExtraBitMatrixLimits) error {
	if err := matrix.checkIndex(row); err != nil {
		return err
	}
	limits, err := normalizeExtraBitMatrixLimits(limits)
	if err != nil || len(words) != matrix.words {
		if err != nil {
			return err
		}
		return ErrDimension
	}
	if len(words) > limits.MaxOperations {
		return ErrLimit
	}
	if limits.Stop != nil && limits.Stop() {
		return ErrCancelled
	}
	start := row * matrix.words
	for index, word := range words {
		matrix.data[start+index] |= word
	}
	return nil
}

func (matrix *ExtraBitMatrix) OrRows(first, second int, limits ExtraBitMatrixLimits) error {
	if err := matrix.checkIndex(first); err != nil {
		return err
	}
	if err := matrix.checkIndex(second); err != nil {
		return err
	}
	limits, err := normalizeExtraBitMatrixLimits(limits)
	if err != nil {
		return err
	}
	if matrix.words > limits.MaxOperations {
		return ErrLimit
	}
	if limits.Stop != nil && limits.Stop() {
		return ErrCancelled
	}
	firstStart, secondStart := first*matrix.words, second*matrix.words
	for word := 0; word < matrix.words; word++ {
		union := matrix.data[firstStart+word] | matrix.data[secondStart+word]
		matrix.data[firstStart+word], matrix.data[secondStart+word] = union, union
	}
	return nil
}

func (matrix *ExtraBitMatrix) CountOnesUpper(limits ExtraBitMatrixLimits) (int, error) {
	if err := matrix.check(); err != nil {
		return 0, err
	}
	limits, err := normalizeExtraBitMatrixLimits(limits)
	if err != nil {
		return 0, err
	}
	operations := int64(matrix.size) * int64(matrix.size-1) / 2
	if operations > int64(limits.MaxOperations) {
		return 0, ErrLimit
	}
	total := 0
	for row := 0; row < matrix.size; row++ {
		if limits.Stop != nil && limits.Stop() {
			return 0, ErrCancelled
		}
		for column := row + 1; column < matrix.size; column++ {
			if matrix.bit(row, column) {
				total++
			}
		}
	}
	return total, nil
}

func (matrix *ExtraBitMatrix) Disjoint(other *ExtraBitMatrix, limits ExtraBitMatrixLimits) (bool, error) {
	if err := matrix.check(); err != nil {
		return false, err
	}
	if err := other.check(); err != nil {
		return false, err
	}
	if matrix.size != other.size {
		return false, ErrDimension
	}
	limits, err := normalizeExtraBitMatrixLimits(limits)
	if err != nil {
		return false, err
	}
	if len(matrix.data) > limits.MaxOperations {
		return false, ErrLimit
	}
	for index, word := range matrix.data {
		if limits.Stop != nil && limits.Stop() {
			return false, ErrCancelled
		}
		if word&other.data[index] != 0 {
			return false, nil
		}
	}
	return true, nil
}

func (matrix *ExtraBitMatrix) IsClique(limits ExtraBitMatrixLimits) (bool, error) {
	if err := matrix.check(); err != nil {
		return false, err
	}
	limits, err := normalizeExtraBitMatrixLimits(limits)
	if err != nil {
		return false, err
	}
	operations, lookups := 0, uint64(0)
	lookup := func(first, second int) (bool, error) {
		operations++
		if operations > limits.MaxOperations {
			return false, ErrLimit
		}
		if limits.Stop != nil && limits.Stop() {
			return false, ErrCancelled
		}
		if first > second {
			first, second = second, first
		}
		lookups++
		return matrix.bit(first, second), nil
	}
	for first := 0; first < matrix.size; first++ {
		for second := first + 1; second < matrix.size; second++ {
			connected, err := lookup(first, second)
			if err != nil {
				return false, err
			}
			if !connected {
				continue
			}
			for candidate := 0; candidate < matrix.size; candidate++ {
				if candidate == first || candidate == second {
					continue
				}
				left, leftErr := lookup(candidate, first)
				if leftErr != nil {
					return false, leftErr
				}
				right, rightErr := lookup(candidate, second)
				if rightErr != nil {
					return false, rightErr
				}
				if left != right {
					matrix.statistics.Lookups += lookups
					return false, nil
				}
			}
		}
	}
	matrix.statistics.Lookups += lookups
	return true, nil
}

func (matrix *ExtraBitMatrix) FormatUpper(limits ExtraBitMatrixLimits) (string, error) {
	if err := matrix.check(); err != nil {
		return "", err
	}
	limits, err := normalizeExtraBitMatrixLimits(limits)
	if err != nil {
		return "", err
	}
	operations := int64(matrix.size) * int64(matrix.size-1) / 2
	bytes64 := int64(1) + int64(matrix.size)*int64(matrix.size+1)
	if operations > int64(limits.MaxOperations) || bytes64 > int64(limits.MaxBytes) || bytes64 > int64(math.MaxInt) {
		return "", ErrLimit
	}
	var builder strings.Builder
	builder.Grow(int(bytes64))
	builder.WriteByte('\n')
	for row := 0; row < matrix.size; row++ {
		if limits.Stop != nil && limits.Stop() {
			return "", ErrCancelled
		}
		builder.WriteString(strings.Repeat(" ", row+1))
		for column := row + 1; column < matrix.size; column++ {
			if matrix.bit(row, column) {
				builder.WriteByte('1')
			} else {
				builder.WriteByte('.')
			}
		}
		builder.WriteByte('\n')
	}
	matrix.statistics.Lookups += uint64(operations)
	return builder.String(), nil
}

func ExtraBitMatrixCountWords(words []uint32) int {
	total := 0
	for _, word := range words {
		total += bits.OnesCount32(word)
	}
	return total
}
