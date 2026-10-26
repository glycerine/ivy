package abc

// Cover cost accounting and deterministic reports from src/misc/espresso/cvrmisc.c.

import (
	"slices"
	"strconv"
)

type EspressoCoverCost struct {
	Cubes, Input, Output, MultipleValued, Total, Primes int
}

type EspressoCoverCostOptions struct {
	SparseVariables     []bool
	MaxCubes, MaxChecks int
	Stop                func() bool
}

func normalizeEspressoCoverCostOptions(layout *EspressoCubeLayout, options EspressoCoverCostOptions) (EspressoCoverCostOptions, error) {
	if options.MaxCubes < 0 || options.MaxChecks < 0 {
		return options, ErrDimension
	}
	if len(options.SparseVariables) != 0 && len(options.SparseVariables) != len(layout.partSizes) {
		return options, ErrDimension
	}
	if options.MaxCubes == 0 {
		options.MaxCubes = 1 << 20
	}
	if options.MaxChecks == 0 {
		options.MaxChecks = 16 << 20
	}
	if len(options.SparseVariables) == 0 {
		options.SparseVariables = make([]bool, len(layout.partSizes))
		for variable := layout.binaryVariables; variable < len(layout.partSizes); variable++ {
			options.SparseVariables[variable] = true
		}
	} else {
		options.SparseVariables = slices.Clone(options.SparseVariables)
	}
	return options, nil
}

// CoverCost counts missing binary/dense parts and present sparse/output parts,
// matching cover_cost after massive_count with an empty cofactor.
func (layout *EspressoCubeLayout) CoverCost(family *EspressoSetFamily, options EspressoCoverCostOptions) (EspressoCoverCost, error) {
	if err := layout.check(); err != nil {
		return EspressoCoverCost{}, err
	}
	if err := family.check(); err != nil {
		return EspressoCoverCost{}, err
	}
	if family.bits != layout.bitCount {
		return EspressoCoverCost{}, ErrDimension
	}
	options, err := normalizeEspressoCoverCostOptions(layout, options)
	if err != nil {
		return EspressoCoverCost{}, err
	}
	if len(family.sets) > options.MaxCubes {
		return EspressoCoverCost{}, ErrLimit
	}
	zeros := make([]int, len(layout.partSizes))
	checks := 0
	for _, cube := range family.sets {
		for variable, size := range layout.partSizes {
			start := layout.firstParts[variable]
			for part := 0; part < size; part++ {
				checks++
				if checks > options.MaxChecks {
					return EspressoCoverCost{}, ErrLimit
				}
				if options.Stop != nil && options.Stop() {
					return EspressoCoverCost{}, ErrCancelled
				}
				present, _ := cube.Has(start + part)
				if !present {
					zeros[variable]++
				}
			}
		}
	}
	cost := EspressoCoverCost{Cubes: len(family.sets)}
	for variable := 0; variable < layout.binaryVariables; variable++ {
		cost.Input += zeros[variable]
	}
	for variable := layout.binaryVariables; variable < len(layout.partSizes)-1; variable++ {
		if options.SparseVariables[variable] {
			cost.MultipleValued += len(family.sets)*layout.partSizes[variable] - zeros[variable]
		} else {
			cost.MultipleValued += zeros[variable]
		}
	}
	if layout.binaryVariables != len(layout.partSizes) {
		output := len(layout.partSizes) - 1
		cost.Output = len(family.sets)*layout.partSizes[output] - zeros[output]
	}
	for _, cube := range family.sets {
		if cube.Flags&EspressoSetPrime != 0 {
			cost.Primes++
		}
	}
	cost.Total = cost.Input + cost.Output + cost.MultipleValued
	return cost, nil
}

func (cost EspressoCoverCost) validate() error {
	if cost.Cubes < 0 || cost.Input < 0 || cost.Output < 0 || cost.MultipleValued < 0 || cost.Total < 0 || cost.Primes < 0 || cost.Primes > cost.Cubes || cost.Total != cost.Input+cost.Output+cost.MultipleValued {
		return ErrInvariant
	}
	return nil
}

// Clone represents copy_cost through ordinary value ownership.
func (cost EspressoCoverCost) Clone() EspressoCoverCost { return cost }

func (layout *EspressoCubeLayout) FormatCoverCost(cost EspressoCoverCost) (string, error) {
	if err := layout.check(); err != nil {
		return "", err
	}
	if err := cost.validate(); err != nil {
		return "", err
	}
	result := "c=" + strconv.Itoa(cost.Cubes) + "(" + strconv.Itoa(cost.Cubes-cost.Primes) + ") in=" + strconv.Itoa(cost.Input)
	if layout.binaryVariables == len(layout.partSizes)-1 {
		result += " out=" + strconv.Itoa(cost.Output) + " tot=" + strconv.Itoa(cost.Total)
	} else {
		result += " mv=" + strconv.Itoa(cost.MultipleValued) + " out=" + strconv.Itoa(cost.Output)
	}
	return result, nil
}

func (layout *EspressoCubeLayout) FormatCover(family *EspressoSetFamily, options EspressoCoverCostOptions) (string, error) {
	cost, err := layout.CoverCost(family, options)
	if err != nil {
		return "", err
	}
	return layout.FormatCoverCost(cost)
}

func (layout *EspressoCubeLayout) FormatCoverSize(name string, family *EspressoSetFamily, options EspressoCoverCostOptions) (string, error) {
	formatted, err := layout.FormatCover(family, options)
	if err != nil {
		return "", err
	}
	return "# " + name + "\tCost is " + formatted + "\n", nil
}

func formatEspressoMilliseconds(milliseconds int64) (string, error) {
	if milliseconds < 0 {
		return "", ErrDimension
	}
	return strconv.FormatInt(milliseconds/1000, 10) + "." + twoDecimalDigits(milliseconds%1000/10) + " sec", nil
}

func twoDecimalDigits(value int64) string {
	if value < 10 {
		return "0" + strconv.FormatInt(value, 10)
	}
	return strconv.FormatInt(value, 10)
}

func (layout *EspressoCubeLayout) FormatCoverTrace(name string, elapsedMilliseconds int64, family *EspressoSetFamily, options EspressoCoverCostOptions) (string, error) {
	elapsed, err := formatEspressoMilliseconds(elapsedMilliseconds)
	if err != nil {
		return "", err
	}
	formatted, err := layout.FormatCover(family, options)
	if err != nil {
		return "", err
	}
	return "# " + name + "\tTime was " + elapsed + ", cost is " + formatted + "\n", nil
}

type EspressoCoverTotals struct {
	Names        []string
	Milliseconds []int64
	Calls        []int
}

func NewEspressoCoverTotals(names []string) *EspressoCoverTotals {
	return &EspressoCoverTotals{Names: slices.Clone(names), Milliseconds: make([]int64, len(names)), Calls: make([]int, len(names))}
}

// Record represents totals with caller-supplied elapsed time, avoiding hidden
// process clocks. It returns the current cost and the optional trace line.
func (totals *EspressoCoverTotals) Record(index int, elapsedMilliseconds int64, layout *EspressoCubeLayout, family *EspressoSetFamily, options EspressoCoverCostOptions, trace bool) (EspressoCoverCost, string, error) {
	if totals == nil || len(totals.Names) != len(totals.Milliseconds) || len(totals.Names) != len(totals.Calls) {
		return EspressoCoverCost{}, "", ErrInvariant
	}
	if index < 0 || index >= len(totals.Names) || elapsedMilliseconds < 0 {
		return EspressoCoverCost{}, "", ErrDimension
	}
	cost, err := layout.CoverCost(family, options)
	if err != nil {
		return EspressoCoverCost{}, "", err
	}
	if totals.Milliseconds[index] > int64(^uint64(0)>>1)-elapsedMilliseconds || totals.Calls[index] == int(^uint(0)>>1) {
		return EspressoCoverCost{}, "", ErrLimit
	}
	totals.Milliseconds[index] += elapsedMilliseconds
	totals.Calls[index]++
	if !trace {
		return cost, "", nil
	}
	formatted, _ := layout.FormatCoverCost(cost)
	elapsed, _ := formatEspressoMilliseconds(elapsedMilliseconds)
	return cost, "# " + totals.Names[index] + "\tTime was " + elapsed + ", cost is " + formatted + "\n", nil
}
