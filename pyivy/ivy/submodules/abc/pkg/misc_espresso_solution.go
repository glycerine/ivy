package abc

// Minimum-cover solution state from src/misc/espresso/solution.c.

import (
	"math"
	"math/bits"
	"slices"
)

type MvcCoverSolutionOptions struct {
	MaxRows, MaxEntries int
	Stop                func() bool
}

type MvcCoverSolution struct {
	columns  int
	selected []bool
	cost     int
	closed   bool
}

func NewMvcCoverSolution(columns int) (*MvcCoverSolution, error) {
	return NewMvcCoverSolutionWithLimit(columns, 1<<20)
}

func NewMvcCoverSolutionWithLimit(columns, maxColumns int) (*MvcCoverSolution, error) {
	if columns < 0 || maxColumns < 0 {
		return nil, ErrDimension
	}
	if columns > maxColumns {
		return nil, ErrLimit
	}
	return &MvcCoverSolution{columns: columns, selected: make([]bool, columns)}, nil
}

func (solution *MvcCoverSolution) check() error {
	if solution == nil {
		return ErrDimension
	}
	if solution.closed || solution.columns < 0 || len(solution.selected) != solution.columns || solution.cost < 0 {
		return ErrInvariant
	}
	return nil
}

func (solution *MvcCoverSolution) Cost() (int, error) {
	if err := solution.check(); err != nil {
		return 0, err
	}
	return solution.cost, nil
}

func (solution *MvcCoverSolution) Columns() ([]int, error) {
	if err := solution.check(); err != nil {
		return nil, err
	}
	result := make([]int, 0)
	for column, selected := range solution.selected {
		if selected {
			result = append(result, column)
		}
	}
	return result, nil
}

func (solution *MvcCoverSolution) Clone() (*MvcCoverSolution, error) {
	if err := solution.check(); err != nil {
		return nil, err
	}
	return &MvcCoverSolution{columns: solution.columns, selected: slices.Clone(solution.selected), cost: solution.cost}, nil
}

func (solution *MvcCoverSolution) Close() error {
	if err := solution.check(); err != nil {
		return err
	}
	solution.selected = nil
	solution.closed = true
	return nil
}

func mvcSolutionWeight(columns int, weights []int, column int) (int, error) {
	if column < 0 || column >= columns || weights != nil && len(weights) != columns {
		return 0, ErrDimension
	}
	weight := 1
	if weights != nil {
		weight = weights[column]
	}
	if weight < 0 {
		return 0, ErrDimension
	}
	return weight, nil
}

// Add selects one previously unselected column and adds its unit or supplied
// nonnegative weight.
func (solution *MvcCoverSolution) Add(column int, weights []int) error {
	if err := solution.check(); err != nil {
		return err
	}
	weight, err := mvcSolutionWeight(solution.columns, weights, column)
	if err != nil {
		return err
	}
	if solution.selected[column] {
		return ErrInvariant
	}
	if weight > math.MaxInt-solution.cost {
		return ErrLimit
	}
	solution.selected[column] = true
	solution.cost += weight
	return nil
}

func normalizeMvcCoverSolutionOptions(options MvcCoverSolutionOptions) (MvcCoverSolutionOptions, error) {
	if options.MaxRows < 0 || options.MaxEntries < 0 {
		return options, ErrDimension
	}
	if options.MaxRows == 0 {
		options.MaxRows = 1 << 20
	}
	if options.MaxEntries == 0 {
		options.MaxEntries = 16 << 20
	}
	return options, nil
}

func mvcSolutionTransform(cover *MvcCover, column int, accept bool, options MvcCoverSolutionOptions) (*MvcCover, bool, error) {
	if err := mvcCheckOwnedCubes(cover); err != nil {
		return nil, false, err
	}
	if column < 0 || column >= cover.bits {
		return nil, false, ErrDimension
	}
	if len(cover.cubes) > options.MaxRows {
		return nil, false, ErrLimit
	}
	result, err := NewMvcCover(cover.manager, cover.bits)
	if err != nil {
		return nil, false, err
	}
	present, entries := false, 0
	for _, source := range cover.cubes {
		if options.Stop != nil && options.Stop() {
			_ = result.Close()
			return nil, false, ErrCancelled
		}
		hasColumn := source.words[column>>5]&(uint32(1)<<uint(column&31)) != 0
		present = present || hasColumn
		if accept && hasColumn {
			continue
		}
		cube, cubeErr := result.cloneCube(source)
		if cubeErr != nil {
			_ = result.Close()
			return nil, false, cubeErr
		}
		if !accept && hasColumn {
			cube.words[column>>5] &^= uint32(1) << uint(column&31)
			if count, countErr := cube.OneCount(); countErr != nil {
				_ = cube.Close()
				_ = result.Close()
				return nil, false, countErr
			} else if count == 0 {
				_ = cube.Close()
				continue
			}
		}
		for _, word := range cube.words {
			entries += bits.OnesCount32(word)
			if entries > options.MaxEntries {
				_ = cube.Close()
				_ = result.Close()
				return nil, false, ErrLimit
			}
		}
		if cubeErr = result.AddTail(cube); cubeErr != nil {
			_ = cube.Close()
			_ = result.Close()
			return nil, false, cubeErr
		}
	}
	return result, present, nil
}

// Accept returns the matrix after deleting all rows covered by column and
// atomically adds that column to the solution.
func (solution *MvcCoverSolution) Accept(cover *MvcCover, column int, weights []int, options MvcCoverSolutionOptions) (*MvcCover, error) {
	if err := solution.check(); err != nil {
		return nil, err
	}
	if cover == nil || cover.bits != solution.columns {
		return nil, ErrDimension
	}
	weight, err := mvcSolutionWeight(solution.columns, weights, column)
	if err != nil {
		return nil, err
	}
	if solution.selected[column] {
		return nil, ErrInvariant
	}
	if weight > math.MaxInt-solution.cost {
		return nil, ErrLimit
	}
	options, err = normalizeMvcCoverSolutionOptions(options)
	if err != nil {
		return nil, err
	}
	result, present, err := mvcSolutionTransform(cover, column, true, options)
	if err != nil {
		return nil, err
	}
	if !present {
		_ = result.Close()
		return nil, ErrDimension
	}
	solution.selected[column] = true
	solution.cost += weight
	return result, nil
}

// Reject returns the matrix after deleting column and any rows made empty.
// The solution itself is unchanged.
func (solution *MvcCoverSolution) Reject(cover *MvcCover, column int, options MvcCoverSolutionOptions) (*MvcCover, error) {
	if err := solution.check(); err != nil {
		return nil, err
	}
	if cover == nil || cover.bits != solution.columns {
		return nil, ErrDimension
	}
	options, err := normalizeMvcCoverSolutionOptions(options)
	if err != nil {
		return nil, err
	}
	result, _, err := mvcSolutionTransform(cover, column, false, options)
	return result, err
}

// TakeBetterCoverSolution preserves the native ownership transfer: the lower-
// cost solution wins, ties select first, and the losing nonnil solution closes.
func TakeBetterCoverSolution(first, second *MvcCoverSolution) (*MvcCoverSolution, error) {
	if first != nil {
		if err := first.check(); err != nil {
			return nil, err
		}
	}
	if second != nil {
		if err := second.check(); err != nil {
			return nil, err
		}
	}
	if first == nil {
		return second, nil
	}
	if second == nil {
		return first, nil
	}
	if first.cost <= second.cost {
		if err := second.Close(); err != nil {
			return nil, err
		}
		return first, nil
	}
	if err := first.Close(); err != nil {
		return nil, err
	}
	return second, nil
}
