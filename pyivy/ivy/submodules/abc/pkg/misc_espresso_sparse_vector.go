package abc

// Sparse row and column vectors from src/misc/espresso/rows.c and cols.c.

import (
	"strconv"
	"strings"
)

type EspressoSparseRow struct {
	Number int
	Flag   int
	Tag    int
	values []int
}

type EspressoSparseColumn struct {
	Number int
	Flag   int
	Tag    int
	values []int
}

func sparseSortedFind(values []int, value int) (int, bool) {
	left, right := 0, len(values)
	for left < right {
		middle := left + (right-left)/2
		if values[middle] < value {
			left = middle + 1
		} else {
			right = middle
		}
	}
	return left, left < len(values) && values[left] == value
}

func sparseSortedInsert(values []int, value int) ([]int, bool, error) {
	if value < 0 {
		return values, false, ErrDimension
	}
	position, found := sparseSortedFind(values, value)
	if found {
		return values, false, nil
	}
	values = append(values, 0)
	copy(values[position+1:], values[position:])
	values[position] = value
	return values, true, nil
}

func sparseSortedRemove(values []int, value int) ([]int, bool, error) {
	if value < 0 {
		return values, false, ErrDimension
	}
	position, found := sparseSortedFind(values, value)
	if !found {
		return values, false, nil
	}
	copy(values[position:], values[position+1:])
	return values[:len(values)-1], true, nil
}

func sparseSortedContains(first, second []int) bool {
	left, right := 0, 0
	for left < len(first) {
		if right == len(second) || first[left] < second[right] {
			return false
		}
		if first[left] == second[right] {
			left++
		}
		right++
	}
	return true
}

func sparseSortedIntersects(first, second []int) bool {
	left, right := 0, 0
	for left < len(first) && right < len(second) {
		switch {
		case first[left] < second[right]:
			left++
		case first[left] > second[right]:
			right++
		default:
			return true
		}
	}
	return false
}

func sparseSortedCompare(first, second []int) int {
	for index := 0; index < min(len(first), len(second)); index++ {
		if first[index] < second[index] {
			return -1
		}
		if first[index] > second[index] {
			return 1
		}
	}
	if len(first) < len(second) {
		return -1
	}
	if len(first) > len(second) {
		return 1
	}
	return 0
}

func sparseSortedIntersection(first, second []int) []int {
	result := make([]int, 0, min(len(first), len(second)))
	left, right := 0, 0
	for left < len(first) && right < len(second) {
		switch {
		case first[left] < second[right]:
			left++
		case first[left] > second[right]:
			right++
		default:
			result = append(result, first[left])
			left++
			right++
		}
	}
	return result
}

func sparseSortedHash(values []int, modulus int) (int, error) {
	if modulus <= 0 {
		return 0, ErrDimension
	}
	sum := 0
	for _, value := range values {
		product := 0
		for range 17 {
			if product >= modulus-sum {
				product -= modulus - sum
			} else {
				product += sum
			}
		}
		value %= modulus
		if product >= modulus-value {
			sum = product - (modulus - value)
		} else {
			sum = product + value
		}
	}
	return sum, nil
}

func sparseSortedString(values []int) string {
	var text strings.Builder
	for _, value := range values {
		text.WriteByte(' ')
		text.WriteString(strconv.Itoa(value))
	}
	return text.String()
}

func NewEspressoSparseRow() *EspressoSparseRow { return &EspressoSparseRow{} }

func (row *EspressoSparseRow) Clone() (*EspressoSparseRow, error) {
	if row == nil {
		return nil, ErrInvariant
	}
	return &EspressoSparseRow{Number: row.Number, Flag: row.Flag, Tag: row.Tag, values: append([]int(nil), row.values...)}, nil
}

func (row *EspressoSparseRow) Values() ([]int, error) {
	if row == nil {
		return nil, ErrInvariant
	}
	return append([]int(nil), row.values...), nil
}

func (row *EspressoSparseRow) Len() int {
	if row == nil {
		return 0
	}
	return len(row.values)
}

func (row *EspressoSparseRow) Insert(column int) (bool, error) {
	if row == nil {
		return false, ErrInvariant
	}
	var inserted bool
	var err error
	row.values, inserted, err = sparseSortedInsert(row.values, column)
	return inserted, err
}

func (row *EspressoSparseRow) Remove(column int) (bool, error) {
	if row == nil {
		return false, ErrInvariant
	}
	var removed bool
	var err error
	row.values, removed, err = sparseSortedRemove(row.values, column)
	return removed, err
}

func (row *EspressoSparseRow) Has(column int) (bool, error) {
	if row == nil {
		return false, ErrInvariant
	}
	if column < 0 {
		return false, ErrDimension
	}
	_, found := sparseSortedFind(row.values, column)
	return found, nil
}

func (row *EspressoSparseRow) Contains(other *EspressoSparseRow) (bool, error) {
	if row == nil || other == nil {
		return false, ErrInvariant
	}
	return sparseSortedContains(other.values, row.values), nil
}

func (row *EspressoSparseRow) Intersects(other *EspressoSparseRow) (bool, error) {
	if row == nil || other == nil {
		return false, ErrInvariant
	}
	return sparseSortedIntersects(row.values, other.values), nil
}

func (row *EspressoSparseRow) Compare(other *EspressoSparseRow) (int, error) {
	if row == nil || other == nil {
		return 0, ErrInvariant
	}
	return sparseSortedCompare(row.values, other.values), nil
}

func (row *EspressoSparseRow) Intersection(other *EspressoSparseRow) (*EspressoSparseRow, error) {
	if row == nil || other == nil {
		return nil, ErrInvariant
	}
	return &EspressoSparseRow{values: sparseSortedIntersection(row.values, other.values)}, nil
}

func (row *EspressoSparseRow) Hash(modulus int) (int, error) {
	if row == nil {
		return 0, ErrInvariant
	}
	return sparseSortedHash(row.values, modulus)
}

func (row *EspressoSparseRow) String() string {
	if row == nil {
		return ""
	}
	return sparseSortedString(row.values)
}

func NewEspressoSparseColumn() *EspressoSparseColumn { return &EspressoSparseColumn{} }

func (column *EspressoSparseColumn) Clone() (*EspressoSparseColumn, error) {
	if column == nil {
		return nil, ErrInvariant
	}
	return &EspressoSparseColumn{Number: column.Number, Flag: column.Flag, Tag: column.Tag, values: append([]int(nil), column.values...)}, nil
}

func (column *EspressoSparseColumn) Values() ([]int, error) {
	if column == nil {
		return nil, ErrInvariant
	}
	return append([]int(nil), column.values...), nil
}

func (column *EspressoSparseColumn) Len() int {
	if column == nil {
		return 0
	}
	return len(column.values)
}

func (column *EspressoSparseColumn) Insert(row int) (bool, error) {
	if column == nil {
		return false, ErrInvariant
	}
	var inserted bool
	var err error
	column.values, inserted, err = sparseSortedInsert(column.values, row)
	return inserted, err
}

func (column *EspressoSparseColumn) Remove(row int) (bool, error) {
	if column == nil {
		return false, ErrInvariant
	}
	var removed bool
	var err error
	column.values, removed, err = sparseSortedRemove(column.values, row)
	return removed, err
}

func (column *EspressoSparseColumn) Has(row int) (bool, error) {
	if column == nil {
		return false, ErrInvariant
	}
	if row < 0 {
		return false, ErrDimension
	}
	_, found := sparseSortedFind(column.values, row)
	return found, nil
}

func (column *EspressoSparseColumn) Contains(other *EspressoSparseColumn) (bool, error) {
	if column == nil || other == nil {
		return false, ErrInvariant
	}
	return sparseSortedContains(other.values, column.values), nil
}

func (column *EspressoSparseColumn) Intersects(other *EspressoSparseColumn) (bool, error) {
	if column == nil || other == nil {
		return false, ErrInvariant
	}
	return sparseSortedIntersects(column.values, other.values), nil
}

func (column *EspressoSparseColumn) Compare(other *EspressoSparseColumn) (int, error) {
	if column == nil || other == nil {
		return 0, ErrInvariant
	}
	return sparseSortedCompare(column.values, other.values), nil
}

func (column *EspressoSparseColumn) Intersection(other *EspressoSparseColumn) (*EspressoSparseColumn, error) {
	if column == nil || other == nil {
		return nil, ErrInvariant
	}
	return &EspressoSparseColumn{values: sparseSortedIntersection(column.values, other.values)}, nil
}

func (column *EspressoSparseColumn) Hash(modulus int) (int, error) {
	if column == nil {
		return 0, ErrInvariant
	}
	return sparseSortedHash(column.values, modulus)
}

func (column *EspressoSparseColumn) String() string {
	if column == nil {
		return ""
	}
	return sparseSortedString(column.values)
}
