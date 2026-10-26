package abc

// Additional ISOP/ESOP operations from src/misc/util/utilIsop.c.

import "slices"

type TruthESOPOptions struct {
	MaxCubes, MaxElements, MaxVisits int
	Stop                             func() bool
}

type TruthESOP struct {
	Cubes    []uint32
	Literals int
	Table    *TruthTable
}

type utilCoverWork struct {
	maxCubes, maxElements, maxVisits int
	elements, visits                 int
	stop                             func() bool
}

func newUtilCoverWork(maxCubes, maxElements, maxVisits int, stop func() bool) (*utilCoverWork, error) {
	if maxCubes < 0 || maxElements < 0 || maxVisits < 0 {
		return nil, ErrDimension
	}
	if maxCubes == 0 {
		maxCubes = 0xffff
	}
	if maxElements == 0 {
		maxElements = 1 << 22
	}
	if maxVisits == 0 {
		maxVisits = 1 << 24
	}
	return &utilCoverWork{maxCubes: maxCubes, maxElements: maxElements, maxVisits: maxVisits, stop: stop}, nil
}

func (work *utilCoverWork) charge(amount int) error {
	if amount < 0 {
		return ErrDimension
	}
	if work.stop != nil && work.stop() {
		return ErrCancelled
	}
	if amount > work.maxVisits-work.visits {
		return ErrLimit
	}
	work.visits += amount
	return nil
}

func (work *utilCoverWork) reserve(amount int) error {
	if err := work.charge(1); err != nil {
		return err
	}
	if amount < 0 || amount > work.maxElements-work.elements {
		return ErrLimit
	}
	work.elements += amount
	return nil
}

func truthLogicalValues(table *TruthTable, work *utilCoverWork) ([]bool, error) {
	if table == nil || table.variables < 0 || table.variables > 16 {
		return nil, ErrDimension
	}
	if err := table.Check(); err != nil {
		return nil, err
	}
	size := 1 << uint(table.variables)
	if err := work.reserve(size); err != nil {
		return nil, err
	}
	values := make([]bool, size)
	for assignment := range values {
		if err := work.charge(1); err != nil {
			return nil, err
		}
		values[assignment] = table.words[assignment/64]&(uint64(1)<<uint(assignment%64)) != 0
	}
	return values, nil
}

func cubeMatchesAssignment(cube uint32, assignment, variables int) bool {
	for variable := 0; variable < variables; variable++ {
		switch cube >> uint(2*variable) & 3 {
		case 1:
			if assignment>>uint(variable)&1 != 0 {
				return false
			}
		case 2:
			if assignment>>uint(variable)&1 == 0 {
				return false
			}
		case 3:
			return false
		}
	}
	return true
}

// BuildCubeCoverTruth implements Abc_IsopBuildTruth for ordinary or exclusive
// covers and optional final complementation.
func BuildCubeCoverTruth(variables int, cubes []uint32, exclusive, complement bool) (*TruthTable, error) {
	if variables < 0 || variables > 16 {
		return nil, ErrDimension
	}
	for _, cube := range cubes {
		for variable := 0; variable < variables; variable++ {
			if cube>>uint(2*variable)&3 == 3 {
				return nil, ErrDimension
			}
		}
		if variables < 16 && cube>>uint(2*variables) != 0 {
			return nil, ErrDimension
		}
	}
	result, err := NewConstantTruthTable(variables, complement, 0)
	if err != nil {
		return nil, err
	}
	for assignment := 0; assignment < 1<<uint(variables); assignment++ {
		value := false
		for _, cube := range cubes {
			if !cubeMatchesAssignment(cube, assignment, variables) {
				continue
			}
			if exclusive {
				value = !value
			} else {
				value = true
				break
			}
		}
		if complement {
			value = !value
		}
		if err := result.Set(uint64(assignment), value); err != nil {
			return nil, err
		}
	}
	return result, nil
}

type utilESOPResult struct {
	cubes    []uint32
	literals int
}

func utilCoverCostCompare(first, second utilESOPResult) int {
	if len(first.cubes) < len(second.cubes) || len(first.cubes) == len(second.cubes) && first.literals < second.literals {
		return -1
	}
	if len(first.cubes) == len(second.cubes) && first.literals == second.literals {
		return 0
	}
	return 1
}

func (table *TruthTable) ESOP(options TruthESOPOptions) (TruthESOP, error) {
	work, err := newUtilCoverWork(options.MaxCubes, options.MaxElements, options.MaxVisits, options.Stop)
	if err != nil {
		return TruthESOP{}, err
	}
	values, err := truthLogicalValues(table, work)
	if err != nil {
		return TruthESOP{}, err
	}
	var recurse func([]bool, int) (utilESOPResult, error)
	recurse = func(function []bool, variables int) (utilESOPResult, error) {
		if err := work.charge(1); err != nil {
			return utilESOPResult{}, err
		}
		allZero, allOne := true, true
		for _, value := range function {
			if err := work.charge(1); err != nil {
				return utilESOPResult{}, err
			}
			allZero = allZero && !value
			allOne = allOne && value
		}
		if allZero {
			return utilESOPResult{}, nil
		}
		if allOne {
			if work.maxCubes <= 1 {
				return utilESOPResult{}, ErrLimit
			}
			if err := work.reserve(1); err != nil {
				return utilESOPResult{}, err
			}
			return utilESOPResult{cubes: []uint32{0}}, nil
		}
		variable := -1
		for candidate := variables - 1; candidate >= 0; candidate-- {
			half := 1 << uint(candidate)
			for assignment := 0; assignment < half; assignment++ {
				if err := work.charge(1); err != nil {
					return utilESOPResult{}, err
				}
				if function[assignment] != function[half+assignment] {
					variable = candidate
					break
				}
			}
			if variable >= 0 {
				break
			}
		}
		if variable < 0 {
			return utilESOPResult{}, ErrInvariant
		}
		half := 1 << uint(variable)
		if err := work.reserve(3 * half); err != nil {
			return utilESOPResult{}, err
		}
		zero := slices.Clone(function[:half])
		one := slices.Clone(function[half : 2*half])
		difference := make([]bool, half)
		for index := range difference {
			difference[index] = zero[index] != one[index]
		}
		r0, err := recurse(zero, variable)
		if err != nil {
			return utilESOPResult{}, err
		}
		r1, err := recurse(one, variable)
		if err != nil {
			return utilESOPResult{}, err
		}
		r2, err := recurse(difference, variable)
		if err != nil {
			return utilESOPResult{}, err
		}
		drop := 0
		if utilCoverCostCompare(r1, r0) > 0 {
			drop = 1
		}
		maximum := r0
		if drop == 1 {
			maximum = r1
		}
		if utilCoverCostCompare(r2, maximum) > 0 {
			drop = 2
		}
		var result utilESOPResult
		switch drop {
		case 0:
			result.cubes = append(result.cubes, r1.cubes...)
			for _, cube := range r2.cubes {
				result.cubes = append(result.cubes, cube|uint32(1)<<uint(2*variable))
			}
			result.literals = r1.literals + r2.literals + len(r2.cubes)
		case 1:
			result.cubes = append(result.cubes, r0.cubes...)
			for _, cube := range r2.cubes {
				result.cubes = append(result.cubes, cube|uint32(2)<<uint(2*variable))
			}
			result.literals = r0.literals + r2.literals + len(r2.cubes)
		case 2:
			for _, cube := range r0.cubes {
				result.cubes = append(result.cubes, cube|uint32(1)<<uint(2*variable))
			}
			for _, cube := range r1.cubes {
				result.cubes = append(result.cubes, cube|uint32(2)<<uint(2*variable))
			}
			result.literals = r0.literals + r1.literals + len(r0.cubes) + len(r1.cubes)
		}
		if len(result.cubes) >= work.maxCubes {
			return utilESOPResult{}, ErrLimit
		}
		if err := work.reserve(len(result.cubes)); err != nil {
			return utilESOPResult{}, err
		}
		return result, nil
	}
	result, err := recurse(values, table.variables)
	if err != nil {
		return TruthESOP{}, err
	}
	realized, err := BuildCubeCoverTruth(table.variables, result.cubes, true, false)
	if err != nil {
		return TruthESOP{}, err
	}
	return TruthESOP{Cubes: result.cubes, Literals: result.literals, Table: realized}, nil
}

func (result TruthESOP) Format(variables int) (string, error) {
	return (TruthISOP{Cubes: result.Cubes}).Format(variables)
}

type TruthGreedyISOPOptions struct {
	MaxCubes, MaxElements, MaxVisits int
	Stop                             func() bool
}

// GreedyISOPInterval implements Abc_IsopNew's first-minterm cube expansion.
func (on *TruthTable) GreedyISOPInterval(allowed *TruthTable, options TruthGreedyISOPOptions) (TruthISOP, error) {
	if on == nil || allowed == nil || on.variables != allowed.variables || on.variables > 16 {
		return TruthISOP{}, ErrDimension
	}
	work, err := newUtilCoverWork(options.MaxCubes, options.MaxElements, options.MaxVisits, options.Stop)
	if err != nil {
		return TruthISOP{}, err
	}
	onset, err := truthLogicalValues(on, work)
	if err != nil {
		return TruthISOP{}, err
	}
	permitted, err := truthLogicalValues(allowed, work)
	if err != nil {
		return TruthISOP{}, err
	}
	for index := range onset {
		if onset[index] && !permitted[index] {
			return TruthISOP{}, ErrDimension
		}
	}
	variables := on.variables
	realized := make([]bool, len(onset))
	if err := work.reserve(len(realized)); err != nil {
		return TruthISOP{}, err
	}
	cubes := make([]uint32, 0)
	literals := 0
	for {
		minterm := -1
		for assignment, value := range onset {
			if err := work.charge(1); err != nil {
				return TruthISOP{}, err
			}
			if value {
				minterm = assignment
				break
			}
		}
		if minterm < 0 {
			break
		}
		if len(cubes)+1 >= work.maxCubes {
			return TruthISOP{}, ErrLimit
		}
		var cube uint32
		for variable := 0; variable < variables; variable++ {
			literal := uint32(1)
			if minterm>>uint(variable)&1 != 0 {
				literal = 2
			}
			cube |= literal << uint(2*variable)
		}
		blocked := make([]bool, variables)
		if err := work.reserve(variables); err != nil {
			return TruthISOP{}, err
		}
		blockedCount := 0
		for variable := 0; variable < variables; variable++ {
			blocked[variable] = !permitted[minterm^(1<<uint(variable))]
			if blocked[variable] {
				blockedCount++
			}
		}
		firstPair, secondPair := -1, -1
		if blockedCount < variables-1 {
			for first := 0; first < variables && firstPair < 0; first++ {
				if blocked[first] {
					continue
				}
				for second := first + 1; second < variables; second++ {
					if !blocked[second] && permitted[minterm^(1<<uint(first))^(1<<uint(second))] {
						firstPair, secondPair = first, second
						break
					}
				}
			}
		}
		if blockedCount == variables {
			// Keep the full minterm cube.
		} else if firstPair < 0 {
			for variable := 0; variable < variables; variable++ {
				if !blocked[variable] {
					cube &^= uint32(3) << uint(2*variable)
					break
				}
			}
		} else {
			cube &^= uint32(3)<<uint(2*firstPair) | uint32(3)<<uint(2*secondPair)
			if blockedCount < variables-2 {
				for variable := 0; variable < variables; variable++ {
					if variable == firstPair || variable == secondPair {
						continue
					}
					shift := uint(2 * variable)
					// Preserve the source's whole-tail comparison rather than
					// silently changing it to a masked literal comparison.
					literalTail := cube >> shift
					if literalTail != 1 && literalTail != 2 {
						continue
					}
					candidate := cube &^ (uint32(3) << shift)
					safe := true
					for assignment := range permitted {
						if err := work.charge(1); err != nil {
							return TruthISOP{}, err
						}
						if cubeMatchesAssignment(candidate, assignment, variables) && !permitted[assignment] {
							safe = false
							break
						}
					}
					if safe {
						cube = candidate
					}
				}
			}
		}
		cubeLiterals := 0
		for variable := 0; variable < variables; variable++ {
			if cube>>uint(2*variable)&3 != 0 {
				cubeLiterals++
			}
		}
		literals += cubeLiterals
		cubes = append(cubes, cube)
		for assignment := range onset {
			if err := work.charge(1); err != nil {
				return TruthISOP{}, err
			}
			if cubeMatchesAssignment(cube, assignment, variables) {
				realized[assignment] = true
				onset[assignment] = false
			}
		}
	}
	resultTable, err := BuildCubeCoverTruth(variables, cubes, false, false)
	if err != nil {
		return TruthISOP{}, err
	}
	for assignment, value := range realized {
		fromTable, _ := resultTable.Get(uint64(assignment))
		if value != fromTable {
			return TruthISOP{}, ErrInvariant
		}
	}
	return TruthISOP{Cubes: cubes, Literals: literals, Table: resultTable}, nil
}

type TruthISOPCNFOptions struct {
	MaxCubes int
	ISOP     TruthISOPOptions
}

type TruthISOPCNF struct {
	Cubes   []uint32
	OnCubes int
}

// ISOPCNF implements Abc_IsopCnf by tagging the function and complement
// covers with opposite literals of a new selector variable.
func (table *TruthTable) ISOPCNF(options TruthISOPCNFOptions) (TruthISOPCNF, error) {
	if table == nil || table.variables < 1 || table.variables > 15 || options.MaxCubes < 0 {
		return TruthISOPCNF{}, ErrDimension
	}
	if err := table.Check(); err != nil {
		return TruthISOPCNF{}, err
	}
	if options.MaxCubes == 0 {
		options.MaxCubes = 0xffff
	}
	// The source requires exact support of its highest variable.
	dependent := false
	step := 1 << uint(table.variables-1)
	for assignment := 0; assignment < step; assignment++ {
		zero := table.words[assignment/64] >> (uint(assignment) % 64) & 1
		oneIndex := assignment + step
		one := table.words[oneIndex/64] >> (uint(oneIndex) % 64) & 1
		if zero != one {
			dependent = true
			break
		}
	}
	if !dependent {
		return TruthISOPCNF{}, ErrDimension
	}
	directOptions := options.ISOP
	directOptions.TryBoth = false
	direct, err := table.ISOP(directOptions)
	if err != nil {
		return TruthISOPCNF{}, err
	}
	complement, err := table.Not().ISOP(directOptions)
	if err != nil {
		return TruthISOPCNF{}, err
	}
	if len(direct.Cubes)+len(complement.Cubes) >= options.MaxCubes {
		return TruthISOPCNF{}, ErrLimit
	}
	result := TruthISOPCNF{Cubes: make([]uint32, 0, len(direct.Cubes)+len(complement.Cubes)), OnCubes: len(direct.Cubes)}
	negative := uint32(1) << uint(2*table.variables)
	positive := uint32(2) << uint(2*table.variables)
	for _, cube := range direct.Cubes {
		result.Cubes = append(result.Cubes, cube|negative)
	}
	for _, cube := range complement.Cubes {
		result.Cubes = append(result.Cubes, cube|positive)
	}
	return result, nil
}
