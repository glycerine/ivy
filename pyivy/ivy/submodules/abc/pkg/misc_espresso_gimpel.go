package abc

// Gimpel minimum-cover reduction from src/misc/espresso/gimpel.c.

import (
	"math/bits"
	"slices"
)

type MvcGimpelOptions struct {
	MaxRows, MaxColumns, MaxEntries int
	Bound                           int
	Minimum                         MvcMinimumCoverOptions
	Stop                            func() bool
}

type MvcGimpelReduction struct {
	Applied                   bool
	WithinBound               bool
	PrimaryRow, SecondaryRow  int
	FirstColumn, SecondColumn int
	Reduced                   *MvcCover
	Solution                  *MvcCoverSolution
}

func normalizeMvcGimpelOptions(options MvcGimpelOptions) (MvcGimpelOptions, error) {
	if options.MaxRows < 0 || options.MaxColumns < 0 || options.MaxEntries < 0 || options.Bound < 0 {
		return options, ErrDimension
	}
	if options.MaxRows == 0 {
		options.MaxRows = 1 << 20
	}
	if options.MaxColumns == 0 {
		options.MaxColumns = 1 << 20
	}
	if options.MaxEntries == 0 {
		options.MaxEntries = 16 << 20
	}
	if options.Minimum.Stop == nil {
		options.Minimum.Stop = options.Stop
	}
	return options, nil
}

func mvcGimpelRowColumns(cube *MvcCube) []int {
	result := make([]int, 0)
	for wordIndex, word := range cube.words {
		for word != 0 {
			bit := bits.TrailingZeros32(word)
			result = append(result, wordIndex*32+bit)
			word &^= uint32(1) << uint(bit)
		}
	}
	return result
}

func mvcGimpelColumnLengths(cover *MvcCover) []int {
	lengths := make([]int, cover.bits)
	for _, cube := range cover.cubes {
		for _, column := range mvcGimpelRowColumns(cube) {
			lengths[column]++
		}
	}
	return lengths
}

func mvcGimpelBuildReduced(cover *MvcCover, primary, secondary, first, second int, saved []uint32, options MvcGimpelOptions) (*MvcCover, error) {
	reduced, err := NewMvcCover(cover.manager, cover.bits)
	if err != nil {
		return nil, err
	}
	entries := 0
	for row, source := range cover.cubes {
		if options.Stop != nil && options.Stop() {
			_ = reduced.Close()
			return nil, ErrCancelled
		}
		if row == primary || row == secondary {
			continue
		}
		words := slices.Clone(source.words)
		if words[second>>5]&(uint32(1)<<uint(second&31)) != 0 {
			for index := range words {
				words[index] |= saved[index]
			}
		}
		words[first>>5] &^= uint32(1) << uint(first&31)
		words[second>>5] &^= uint32(1) << uint(second&31)
		count := 0
		for _, word := range words {
			count += bits.OnesCount32(word)
		}
		if count == 0 {
			continue
		}
		entries += count
		if entries > options.MaxEntries {
			_ = reduced.Close()
			return nil, ErrLimit
		}
		cube, cubeErr := reduced.NewCube()
		if cubeErr != nil {
			_ = reduced.Close()
			return nil, cubeErr
		}
		copy(cube.words, words)
		if cubeErr = reduced.AddTail(cube); cubeErr != nil {
			_ = cube.Close()
			_ = reduced.Close()
			return nil, cubeErr
		}
	}
	return reduced, nil
}

// GimpelReduce applies the first native two-column reduction, solves the
// reduced unit-cost matrix exactly, and reconstructs the selected original
// column. Base is cloned and never modified. Bound zero means unbounded;
// otherwise solutions whose total cost is not strictly below Bound are absent.
func (cover *MvcCover) GimpelReduce(base *MvcCoverSolution, options MvcGimpelOptions) (MvcGimpelReduction, error) {
	options, err := normalizeMvcGimpelOptions(options)
	if err != nil {
		return MvcGimpelReduction{}, err
	}
	if err = mvcCheckOwnedCubes(cover); err != nil {
		return MvcGimpelReduction{}, err
	}
	if len(cover.cubes) > options.MaxRows || cover.bits > options.MaxColumns {
		return MvcGimpelReduction{}, ErrLimit
	}
	if base != nil {
		if err = base.check(); err != nil {
			return MvcGimpelReduction{}, err
		}
		if base.columns != cover.bits {
			return MvcGimpelReduction{}, ErrDimension
		}
	}
	inputEntries := 0
	for _, cube := range cover.cubes {
		if options.Stop != nil && options.Stop() {
			return MvcGimpelReduction{}, ErrCancelled
		}
		columns := mvcGimpelRowColumns(cube)
		if len(columns) == 0 {
			return MvcGimpelReduction{}, ErrInvariant
		}
		inputEntries += len(columns)
		if inputEntries > options.MaxEntries {
			return MvcGimpelReduction{}, ErrLimit
		}
	}
	lengths := mvcGimpelColumnLengths(cover)
	primary, secondary, first, second := -1, -1, -1, -1
	findSecondary := func(row, column int) int {
		for candidate, cube := range cover.cubes {
			if candidate != row && cube.words[column>>5]&(uint32(1)<<uint(column&31)) != 0 {
				return candidate
			}
		}
		return -1
	}
search:
	for row, cube := range cover.cubes {
		columns := mvcGimpelRowColumns(cube)
		if len(columns) != 2 {
			continue
		}
		for _, orientation := range [][2]int{{columns[0], columns[1]}, {columns[1], columns[0]}} {
			if lengths[orientation[0]] != 2 {
				continue
			}
			other := findSecondary(row, orientation[0])
			if other < 0 || cover.cubes[other].words[orientation[1]>>5]&(uint32(1)<<uint(orientation[1]&31)) != 0 {
				continue
			}
			if len(mvcGimpelRowColumns(cover.cubes[other])) < 2 {
				continue
			}
			primary, secondary, first, second = row, other, orientation[0], orientation[1]
			break search
		}
	}
	if primary < 0 {
		return MvcGimpelReduction{}, nil
	}
	saved := slices.Clone(cover.cubes[secondary].words)
	saved[first>>5] &^= uint32(1) << uint(first&31)
	reduced, err := mvcGimpelBuildReduced(cover, primary, secondary, first, second, saved, options)
	if err != nil {
		return MvcGimpelReduction{}, err
	}
	mask, err := reduced.MinimumColumnCover(options.Minimum)
	if err != nil {
		_ = reduced.Close()
		return MvcGimpelReduction{}, err
	}
	solution, err := NewMvcCoverSolutionWithLimit(cover.bits, options.MaxColumns)
	if err != nil {
		_ = mask.Close()
		_ = reduced.Close()
		return MvcGimpelReduction{}, err
	}
	if base != nil {
		_ = solution.Close()
		solution, err = base.Clone()
		if err != nil {
			_ = mask.Close()
			_ = reduced.Close()
			return MvcGimpelReduction{}, err
		}
	}
	selectedWords, _ := mask.Words()
	for wordIndex, word := range selectedWords {
		for word != 0 {
			bit := bits.TrailingZeros32(word)
			column := wordIndex*32 + bit
			if err = solution.Add(column, nil); err != nil {
				_ = solution.Close()
				_ = mask.Close()
				_ = reduced.Close()
				return MvcGimpelReduction{}, err
			}
			word &^= uint32(1) << uint(bit)
		}
	}
	intersectsSaved := false
	for index, word := range selectedWords {
		if word&saved[index] != 0 {
			intersectsSaved = true
			break
		}
	}
	reconstructed := first
	if intersectsSaved {
		reconstructed = second
	}
	if err = solution.Add(reconstructed, nil); err != nil {
		_ = solution.Close()
		_ = mask.Close()
		_ = reduced.Close()
		return MvcGimpelReduction{}, err
	}
	_ = mask.Close()
	cost, _ := solution.Cost()
	result := MvcGimpelReduction{
		Applied: true, WithinBound: options.Bound == 0 || cost < options.Bound,
		PrimaryRow: primary, SecondaryRow: secondary, FirstColumn: first, SecondColumn: second,
		Reduced: reduced, Solution: solution,
	}
	if !result.WithinBound {
		_ = solution.Close()
		result.Solution = nil
	}
	return result, nil
}
