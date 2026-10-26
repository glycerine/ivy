package abc

// Sparse row and column dominance from src/misc/espresso/dominate.c.

import (
	"math/bits"
	"slices"
)

type MvcDominanceOptions struct {
	MaxRows, MaxColumns, MaxChecks int
	Stop                           func() bool
}

type MvcDominanceResult struct {
	Cover          *MvcCover
	ActiveColumns  []bool
	RemovedRows    int
	RemovedColumns int
}

func normalizeMvcDominanceOptions(options MvcDominanceOptions) (MvcDominanceOptions, error) {
	if options.MaxRows < 0 || options.MaxColumns < 0 || options.MaxChecks < 0 {
		return options, ErrDimension
	}
	if options.MaxRows == 0 {
		options.MaxRows = 1 << 20
	}
	if options.MaxColumns == 0 {
		options.MaxColumns = 1 << 20
	}
	if options.MaxChecks == 0 {
		options.MaxChecks = 4 << 20
	}
	return options, nil
}

type mvcSparseDominance struct {
	cover           *MvcCover
	rows, columns   []bool
	activeWords     []uint32
	checks, maximum int
	stop            func() bool
}

func newMvcSparseDominance(cover *MvcCover, options MvcDominanceOptions) (*mvcSparseDominance, error) {
	if err := mvcCheckOwnedCubes(cover); err != nil {
		return nil, err
	}
	if len(cover.cubes) > options.MaxRows || cover.bits > options.MaxColumns {
		return nil, ErrLimit
	}
	matrix := &mvcSparseDominance{
		cover: cover, rows: make([]bool, len(cover.cubes)), columns: make([]bool, cover.bits),
		activeWords: make([]uint32, cover.words),
		maximum:     options.MaxChecks, stop: options.Stop,
	}
	for row, cube := range cover.cubes {
		if matrix.cancelled() {
			return nil, ErrCancelled
		}
		matrix.rows[row] = true
		present := false
		for wordIndex, word := range cube.words {
			matrix.activeWords[wordIndex] |= word
			for word != 0 {
				bit := bits.TrailingZeros32(word)
				column := wordIndex*32 + bit
				matrix.columns[column] = true
				present = true
				word &^= uint32(1) << uint(bit)
			}
		}
		if !present {
			return nil, ErrInvariant
		}
	}
	if matrix.cancelled() {
		return nil, ErrCancelled
	}
	return matrix, nil
}

func (matrix *mvcSparseDominance) cancelled() bool {
	return matrix.stop != nil && matrix.stop()
}

func (matrix *mvcSparseDominance) checkWork() error {
	matrix.checks++
	if matrix.checks > matrix.maximum {
		return ErrLimit
	}
	if matrix.cancelled() {
		return ErrCancelled
	}
	return nil
}

func (matrix *mvcSparseDominance) bit(row, column int) bool {
	return matrix.cover.cubes[row].words[column>>5]&(uint32(1)<<uint(column&31)) != 0
}

func (matrix *mvcSparseDominance) setColumnActive(column int, active bool) {
	matrix.columns[column] = active
	mask := uint32(1) << uint(column&31)
	if active {
		matrix.activeWords[column>>5] |= mask
	} else {
		matrix.activeWords[column>>5] &^= mask
	}
}

func (matrix *mvcSparseDominance) rowLength(row int) int {
	length := 0
	for index, word := range matrix.cover.cubes[row].words {
		length += bits.OnesCount32(word & matrix.activeWords[index])
	}
	return length
}

func (matrix *mvcSparseDominance) columnLength(column int) int {
	length := 0
	for row, active := range matrix.rows {
		if active && matrix.bit(row, column) {
			length++
		}
	}
	return length
}

func (matrix *mvcSparseDominance) rowSubset(first, second int) bool {
	for index, word := range matrix.cover.cubes[first].words {
		if word&matrix.activeWords[index]&^matrix.cover.cubes[second].words[index] != 0 {
			return false
		}
	}
	return true
}

func (matrix *mvcSparseDominance) columnSubset(first, second int) bool {
	for row, active := range matrix.rows {
		if active && matrix.bit(row, first) && !matrix.bit(row, second) {
			return false
		}
	}
	return true
}

func (matrix *mvcSparseDominance) result(removedRows, removedColumns int) (MvcDominanceResult, error) {
	result, err := NewMvcCover(matrix.cover.manager, matrix.cover.bits)
	if err != nil {
		return MvcDominanceResult{}, err
	}
	for row, active := range matrix.rows {
		if !active {
			continue
		}
		cube, cubeErr := result.NewCube()
		if cubeErr != nil {
			_ = result.Close()
			return MvcDominanceResult{}, cubeErr
		}
		if matrix.cancelled() {
			_ = cube.Close()
			_ = result.Close()
			return MvcDominanceResult{}, ErrCancelled
		}
		for word := range cube.words {
			cube.words[word] = matrix.cover.cubes[row].words[word] & matrix.activeWords[word]
		}
		if cubeErr = result.AddTail(cube); cubeErr != nil {
			_ = cube.Close()
			_ = result.Close()
			return MvcDominanceResult{}, cubeErr
		}
	}
	return MvcDominanceResult{Cover: result, ActiveColumns: slices.Clone(matrix.columns), RemovedRows: removedRows, RemovedColumns: removedColumns}, nil
}

// ReduceDominatedRows removes a row when a lower-numbered or strictly smaller
// active row is its subset. It returns an owned cover rather than mutating the
// sparse matrix in place.
func (cover *MvcCover) ReduceDominatedRows(options MvcDominanceOptions) (MvcDominanceResult, error) {
	options, err := normalizeMvcDominanceOptions(options)
	if err != nil {
		return MvcDominanceResult{}, err
	}
	matrix, err := newMvcSparseDominance(cover, options)
	if err != nil {
		return MvcDominanceResult{}, err
	}
	original := len(matrix.rows)
	for row, active := range matrix.rows {
		if !active {
			continue
		}
		leastColumn, leastLength := -1, 0
		for column, present := range matrix.columns {
			if !present || !matrix.bit(row, column) {
				continue
			}
			length := matrix.columnLength(column)
			if leastColumn < 0 || length < leastLength {
				leastColumn, leastLength = column, length
			}
		}
		if leastColumn < 0 {
			return MvcDominanceResult{}, ErrInvariant
		}
		rowLength := matrix.rowLength(row)
		for candidate, candidateActive := range matrix.rows {
			if !candidateActive || !matrix.bit(candidate, leastColumn) {
				continue
			}
			if err = matrix.checkWork(); err != nil {
				return MvcDominanceResult{}, err
			}
			candidateLength := matrix.rowLength(candidate)
			if (candidateLength > rowLength || candidateLength == rowLength && candidate > row) && matrix.rowSubset(row, candidate) {
				matrix.rows[candidate] = false
			}
		}
	}
	remaining := 0
	for _, active := range matrix.rows {
		if active {
			remaining++
		}
	}
	for column, active := range matrix.columns {
		if active && matrix.columnLength(column) == 0 {
			matrix.setColumnActive(column, false)
		}
	}
	return matrix.result(original-remaining, 0)
}

// ReduceDominatedColumns removes a column when a no-more-expensive active
// column covers a superset of its rows. Nil weights assign unit cost.
func (cover *MvcCover) ReduceDominatedColumns(weights []int, options MvcDominanceOptions) (MvcDominanceResult, error) {
	options, err := normalizeMvcDominanceOptions(options)
	if err != nil {
		return MvcDominanceResult{}, err
	}
	if weights != nil && len(weights) != cover.BitCount() {
		return MvcDominanceResult{}, ErrDimension
	}
	matrix, err := newMvcSparseDominance(cover, options)
	if err != nil {
		return MvcDominanceResult{}, err
	}
	original := 0
	for _, active := range matrix.columns {
		if active {
			original++
		}
	}
	weight := func(column int) int {
		if weights == nil {
			return 1
		}
		return weights[column]
	}
	for column, active := range matrix.columns {
		if !active {
			continue
		}
		leastRow, leastLength := -1, 0
		for row, rowActive := range matrix.rows {
			if !rowActive || !matrix.bit(row, column) {
				continue
			}
			length := matrix.rowLength(row)
			if leastRow < 0 || length < leastLength {
				leastRow, leastLength = row, length
			}
		}
		if leastRow < 0 {
			return MvcDominanceResult{}, ErrInvariant
		}
		columnLength := matrix.columnLength(column)
		for candidate, candidateActive := range matrix.columns {
			if !candidateActive || !matrix.bit(leastRow, candidate) || weight(candidate) > weight(column) {
				continue
			}
			if err = matrix.checkWork(); err != nil {
				return MvcDominanceResult{}, err
			}
			candidateLength := matrix.columnLength(candidate)
			if (candidateLength > columnLength || candidateLength == columnLength && candidate > column) && matrix.columnSubset(column, candidate) {
				matrix.setColumnActive(column, false)
				break
			}
		}
	}
	remaining := 0
	for _, active := range matrix.columns {
		if active {
			remaining++
		}
	}
	return matrix.result(0, original-remaining)
}
