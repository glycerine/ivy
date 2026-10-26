package abc

// Sparse block partitioning from src/misc/espresso/part.c.

import "math/bits"

type MvcBlockPartitionOptions struct {
	MaxRows, MaxColumns, MaxEntries, MaxVisits int
	Stop                                       func() bool
}

type MvcBlockPartition struct {
	Partitioned         bool
	Left, Right         *MvcCover
	LeftRows, RightRows []int
	LeftColumns         []int
	RightColumns        []int
}

func normalizeMvcBlockPartitionOptions(options MvcBlockPartitionOptions) (MvcBlockPartitionOptions, error) {
	if options.MaxRows < 0 || options.MaxColumns < 0 || options.MaxEntries < 0 || options.MaxVisits < 0 {
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
	if options.MaxVisits == 0 {
		options.MaxVisits = 16 << 20
	}
	return options, nil
}

func mvcPartitionCovers(cover *MvcCover, leftRows []bool, stop func() bool) (*MvcCover, *MvcCover, error) {
	left, err := NewMvcCover(cover.manager, cover.bits)
	if err != nil {
		return nil, nil, err
	}
	right, err := NewMvcCover(cover.manager, cover.bits)
	if err != nil {
		_ = left.Close()
		return nil, nil, err
	}
	for row, cube := range cover.cubes {
		if stop != nil && stop() {
			_ = left.Close()
			_ = right.Close()
			return nil, nil, ErrCancelled
		}
		target := right
		if leftRows[row] {
			target = left
		}
		if _, err = target.AddCloneTail(cube); err != nil {
			_ = left.Close()
			_ = right.Close()
			return nil, nil, err
		}
	}
	return left, right, nil
}

// BlockPartition splits the row/column bipartite graph into the component of
// the first row and all remaining components. Connected and empty matrices
// return Partitioned false without allocating result covers.
func (cover *MvcCover) BlockPartition(options MvcBlockPartitionOptions) (MvcBlockPartition, error) {
	options, err := normalizeMvcBlockPartitionOptions(options)
	if err != nil {
		return MvcBlockPartition{}, err
	}
	if err = mvcCheckOwnedCubes(cover); err != nil {
		return MvcBlockPartition{}, err
	}
	if len(cover.cubes) > options.MaxRows || cover.bits > options.MaxColumns {
		return MvcBlockPartition{}, ErrLimit
	}
	if len(cover.cubes) == 0 {
		return MvcBlockPartition{}, nil
	}
	columnRows := make([][]int, cover.bits)
	presentColumns := make([]bool, cover.bits)
	entries := 0
	for row, cube := range cover.cubes {
		if options.Stop != nil && options.Stop() {
			return MvcBlockPartition{}, ErrCancelled
		}
		for wordIndex, word := range cube.words {
			for word != 0 {
				bit := bits.TrailingZeros32(word)
				column := wordIndex*32 + bit
				entries++
				if entries > options.MaxEntries {
					return MvcBlockPartition{}, ErrLimit
				}
				columnRows[column] = append(columnRows[column], row)
				presentColumns[column] = true
				word &^= uint32(1) << uint(bit)
			}
		}
	}
	type item struct {
		index int
		row   bool
	}
	rowsSeen := make([]bool, len(cover.cubes))
	columnsSeen := make([]bool, cover.bits)
	stack := []item{{index: 0, row: true}}
	visits := 0
	for len(stack) != 0 {
		visits++
		if visits > options.MaxVisits {
			return MvcBlockPartition{}, ErrLimit
		}
		if options.Stop != nil && options.Stop() {
			return MvcBlockPartition{}, ErrCancelled
		}
		current := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if current.row {
			if rowsSeen[current.index] {
				continue
			}
			rowsSeen[current.index] = true
			cube := cover.cubes[current.index]
			for wordIndex := len(cube.words) - 1; wordIndex >= 0; wordIndex-- {
				word := cube.words[wordIndex]
				for word != 0 {
					bit := 31 - bits.LeadingZeros32(word)
					column := wordIndex*32 + bit
					if !columnsSeen[column] {
						stack = append(stack, item{index: column})
					}
					word &^= uint32(1) << uint(bit)
				}
			}
		} else {
			if columnsSeen[current.index] {
				continue
			}
			columnsSeen[current.index] = true
			rows := columnRows[current.index]
			for index := len(rows) - 1; index >= 0; index-- {
				if !rowsSeen[rows[index]] {
					stack = append(stack, item{index: rows[index], row: true})
				}
			}
		}
	}
	seenRows, seenColumns, totalColumns := 0, 0, 0
	for _, seen := range rowsSeen {
		if seen {
			seenRows++
		}
	}
	for column, present := range presentColumns {
		if present {
			totalColumns++
			if columnsSeen[column] {
				seenColumns++
			}
		}
	}
	if seenRows == len(rowsSeen) || totalColumns > 0 && seenColumns == totalColumns {
		return MvcBlockPartition{}, nil
	}
	left, right, err := mvcPartitionCovers(cover, rowsSeen, options.Stop)
	if err != nil {
		return MvcBlockPartition{}, err
	}
	result := MvcBlockPartition{Partitioned: true, Left: left, Right: right}
	for row, seen := range rowsSeen {
		if options.Stop != nil && options.Stop() {
			_ = left.Close()
			_ = right.Close()
			return MvcBlockPartition{}, ErrCancelled
		}
		if seen {
			result.LeftRows = append(result.LeftRows, row)
		} else {
			result.RightRows = append(result.RightRows, row)
		}
	}
	for column, present := range presentColumns {
		if options.Stop != nil && options.Stop() {
			_ = left.Close()
			_ = right.Close()
			return MvcBlockPartition{}, ErrCancelled
		}
		if !present {
			continue
		}
		if columnsSeen[column] {
			result.LeftColumns = append(result.LeftColumns, column)
		} else {
			result.RightColumns = append(result.RightColumns, column)
		}
	}
	return result, nil
}
