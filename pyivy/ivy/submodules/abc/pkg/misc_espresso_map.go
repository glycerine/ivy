package abc

// Minterm expansion and Karnaugh-map formatting from src/misc/espresso/map.c.

import (
	"strconv"
	"strings"
)

type EspressoMapOptions struct {
	MaxCubes, MaxMinterms, MaxVisits int
	Stop                             func() bool
}

func normalizeEspressoMapOptions(options EspressoMapOptions) (EspressoMapOptions, error) {
	if options.MaxCubes < 0 || options.MaxMinterms < 0 || options.MaxVisits < 0 {
		return options, ErrDimension
	}
	if options.MaxCubes == 0 {
		options.MaxCubes = 1 << 20
	}
	if options.MaxMinterms == 0 {
		options.MaxMinterms = espressoSetMaximumBits
	}
	if options.MaxVisits == 0 {
		options.MaxVisits = 16 << 20
	}
	return options, nil
}

func (layout *EspressoCubeLayout) mintermCount(limit int) (int, error) {
	if err := layout.check(); err != nil {
		return 0, err
	}
	if len(layout.partSizes) == 0 {
		return 0, ErrDimension
	}
	count := 1
	for _, size := range layout.partSizes {
		if count > limit/size || count > espressoSetMaximumBits/size {
			return 0, ErrLimit
		}
		count *= size
	}
	return count, nil
}

// Minterms expands every product cube into the source mixed-radix minterm
// numbering: variable zero is the least-significant coordinate.
func (layout *EspressoCubeLayout) Minterms(family *EspressoSetFamily, options EspressoMapOptions) (*EspressoSet, error) {
	options, err := normalizeEspressoMapOptions(options)
	if err != nil {
		return nil, err
	}
	if err = family.check(); err != nil {
		return nil, err
	}
	if family.bits != layout.bitCount {
		return nil, ErrDimension
	}
	if len(family.sets) > options.MaxCubes {
		return nil, ErrLimit
	}
	count, err := layout.mintermCount(options.MaxMinterms)
	if err != nil {
		return nil, err
	}
	result, _ := NewEspressoSet(count)
	visits := 0
	var expand func(*EspressoSet, int, int) error
	expand = func(cube *EspressoSet, variable, prefix int) error {
		start, size := layout.firstParts[variable], layout.partSizes[variable]
		for part := 0; part < size; part++ {
			visits++
			if visits > options.MaxVisits {
				return ErrLimit
			}
			if options.Stop != nil && options.Stop() {
				return ErrCancelled
			}
			present, _ := cube.Has(start + part)
			if !present {
				continue
			}
			index := prefix*size + part
			if variable == 0 {
				_ = result.Set(index, true)
			} else if err := expand(cube, variable-1, index); err != nil {
				return err
			}
		}
		return nil
	}
	for _, cube := range family.sets {
		if err = expand(cube, len(layout.partSizes)-1, 0); err != nil {
			return nil, err
		}
	}
	return result, nil
}

func espressoKarnaughIndex(row, column int) int {
	rowLow := row & 3
	rowLow ^= rowLow >> 1
	rowHigh := row >> 2
	rowHigh ^= rowHigh >> 1
	columnLow := column & 3
	columnLow ^= columnLow >> 1
	columnHigh := column >> 2
	columnHigh ^= columnHigh >> 1
	return (columnLow & 1) | (columnLow & 2) |
		(rowLow&1)<<2 | (rowLow&2)<<2 |
		(columnHigh&1)<<4 | (columnHigh&2)<<5 |
		(rowHigh&1)<<5 | (rowHigh&2)<<6
}

// FormatKarnaughMap returns map.c's deterministic stdout representation. The
// source's page count is retained exactly, including its linear count above
// eight binary variables.
func (layout *EspressoCubeLayout) FormatKarnaughMap(family *EspressoSetFamily, options EspressoMapOptions) (string, error) {
	if err := layout.check(); err != nil {
		return "", err
	}
	if layout.binaryVariables < 0 || len(layout.partSizes) != layout.binaryVariables+1 {
		return "", ErrDimension
	}
	minterms, err := layout.Minterms(family, options)
	if err != nil {
		return "", err
	}
	largestInputIndex := 1 << uint(layout.binaryVariables)
	outputs := layout.partSizes[len(layout.partSizes)-1]
	pages := max(layout.binaryVariables-8, 0) + 1
	var result strings.Builder
	for output := 0; output < outputs; output++ {
		result.WriteString("\n\nOutput space # ")
		result.WriteString(strconv.Itoa(output))
		result.WriteByte('\n')
		for page := 0; page < pages; page++ {
			offset := page * 256
			for row := 0; row < 16; row++ {
				someOutput := false
				for column := 0; column < 16; column++ {
					index := espressoKarnaughIndex(row, column) + offset
					if index < largestInputIndex {
						present, _ := minterms.Has(output*largestInputIndex + index)
						if present {
							result.WriteByte('1')
						} else {
							result.WriteByte('.')
						}
						someOutput = true
					}
					if (column+1)%4 == 0 {
						result.WriteByte(' ')
					}
					if (column+1)%8 == 0 {
						result.WriteString("  ")
					}
				}
				if someOutput {
					result.WriteByte('\n')
				}
				if (row+1)%4 == 0 {
					if row != 15 && espressoKarnaughIndex(row+1, 0) >= largestInputIndex {
						break
					}
					result.WriteByte('\n')
				}
				if (row+1)%8 == 0 {
					result.WriteByte('\n')
				}
			}
		}
	}
	return result.String(), nil
}
