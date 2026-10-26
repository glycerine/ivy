package abc

// Non-DSD variable-set encoding and discovery from opt/dau/dauNonDsd.c.

import (
	"fmt"
	"math/bits"
	"strings"
)

type DauDecompositionRole uint8

const (
	DauDecompositionFree   DauDecompositionRole = 0
	DauDecompositionUnique DauDecompositionRole = 1
	DauDecompositionShared DauDecompositionRole = 3
)

type DauDecompositionSet uint32

type DauDecompositionSetStats struct {
	Unique, Shared, Free int
}

type DauDecompositionSetOptions struct {
	MaxSets   int
	MaxChecks uint64
	Stop      func() bool
}

func normalizeDauDecompositionSetOptions(options DauDecompositionSetOptions) (DauDecompositionSetOptions, error) {
	if options.MaxSets < 0 {
		return options, ErrDimension
	}
	if options.MaxSets == 0 {
		options.MaxSets = 1 << 20
	}
	if options.MaxChecks == 0 {
		options.MaxChecks = 1 << 32
	}
	return options, nil
}

func (set DauDecompositionSet) Role(variable int) (DauDecompositionRole, error) {
	if variable < 0 || variable >= 16 {
		return 0, ErrDimension
	}
	role := DauDecompositionRole((uint32(set) >> uint(2*variable)) & 3)
	if role == 2 {
		return 0, ErrInvariant
	}
	return role, nil
}

func (set DauDecompositionSet) Stats(variables int) (DauDecompositionSetStats, error) {
	if variables < 0 || variables > 16 {
		return DauDecompositionSetStats{}, ErrDimension
	}
	var result DauDecompositionSetStats
	for variable := 0; variable < variables; variable++ {
		role, err := set.Role(variable)
		if err != nil {
			return DauDecompositionSetStats{}, err
		}
		switch role {
		case DauDecompositionFree:
			result.Free++
		case DauDecompositionUnique:
			result.Unique++
		case DauDecompositionShared:
			result.Shared++
		}
	}
	if variables < 16 && uint32(set)>>uint(2*variables) != 0 {
		return DauDecompositionSetStats{}, ErrDimension
	}
	return result, nil
}

func ParseDauDecompositionSet(text string, variables int) (DauDecompositionSet, error) {
	if variables < 0 || variables > 16 {
		return 0, ErrDimension
	}
	var result uint32
	for _, symbol := range []byte(text) {
		variable, role := -1, DauDecompositionFree
		switch {
		case symbol >= 'a' && symbol <= 'z':
			variable, role = int(symbol-'a'), DauDecompositionUnique
		case symbol >= 'A' && symbol <= 'Z':
			variable, role = int(symbol-'A'), DauDecompositionShared
		default:
			return 0, ErrInvariant
		}
		if variable >= variables || variable >= 16 {
			return 0, ErrDimension
		}
		shift := uint(2 * variable)
		if result&(3<<shift) != 0 {
			return 0, ErrInvariant
		}
		result |= uint32(role) << shift
	}
	return DauDecompositionSet(result), nil
}

func FormatDauDecompositionSet(set DauDecompositionSet, variables int) (string, error) {
	stats, err := set.Stats(variables)
	if err != nil {
		return "", err
	}
	var result strings.Builder
	fmt.Fprintf(&result, "S =%2d  D =%2d  C =%2d   x=", stats.Shared, stats.Unique+stats.Shared, stats.Shared+stats.Free+1)
	printed := 0
	for variable := 0; variable < variables; variable++ {
		role, _ := set.Role(variable)
		if role == DauDecompositionUnique {
			result.WriteByte(byte('a' + variable))
			printed++
		} else if role == DauDecompositionShared {
			result.WriteByte(byte('A' + variable))
			printed++
		}
	}
	result.WriteString(" y=x")
	for variable := 0; variable < variables; variable++ {
		role, _ := set.Role(variable)
		if role == DauDecompositionFree {
			result.WriteByte(byte('a' + variable))
			printed++
		} else if role == DauDecompositionShared {
			result.WriteByte(byte('A' + variable))
			printed++
		}
	}
	for ; printed < 15; printed++ {
		result.WriteByte(' ')
	}
	return result.String(), nil
}

func FormatDauDecompositionSets(sets []DauDecompositionSet, variables int) (string, error) {
	if variables < 0 || variables > 16 {
		return "", ErrDimension
	}
	var result strings.Builder
	fmt.Fprintf(&result, "The %d-variable set family contains %d sets:\n", variables, len(sets))
	for _, set := range sets {
		line, err := FormatDauDecompositionSet(set, variables)
		if err != nil {
			return "", err
		}
		result.WriteString(line)
		result.WriteByte('\n')
	}
	result.WriteByte('\n')
	return result.String(), nil
}

func createDauDecompositionSet(bound []int, sharedMask int) DauDecompositionSet {
	var result uint32
	for position, variable := range bound {
		role := uint32(DauDecompositionUnique)
		if sharedMask&(1<<uint(position)) != 0 {
			role = uint32(DauDecompositionShared)
		}
		result |= role << uint(2*variable)
	}
	return DauDecompositionSet(result)
}

func dauDecompositionSetHas01(mask uint32) bool {
	return mask&((^mask)>>1)&0x55555555 != 0
}

func dauDecompositionSetContained(sets []DauDecompositionSet, candidate DauDecompositionSet) bool {
	newSet := uint32(candidate)
	for _, set := range sets {
		old := uint32(set)
		if old & ^newSet == 0 && !dauDecompositionSetHas01(^old&newSet) {
			return true
		}
	}
	return false
}

type dauDecompositionCheckBudget struct {
	used, maximum uint64
	stop          func() bool
}

func (budget *dauDecompositionCheckBudget) take() error {
	if budget.stop != nil && budget.stop() {
		return ErrCancelled
	}
	if budget.used >= budget.maximum {
		return ErrLimit
	}
	budget.used++
	return nil
}

// dauDecompositionSetWorks tests whether each assignment to the shared
// variables admits at most two distinct free-variable cofactors as the unique
// bound variables vary.
func dauDecompositionSetWorks(table *TruthTable, bound []int, sharedMask int, budget *dauDecompositionCheckBudget) (bool, error) {
	variables := table.variables
	boundFlag := make([]bool, variables)
	shared, unique := make([]int, 0, len(bound)), make([]int, 0, len(bound))
	for position, variable := range bound {
		boundFlag[variable] = true
		if sharedMask&(1<<uint(position)) != 0 {
			shared = append(shared, variable)
		} else {
			unique = append(unique, variable)
		}
	}
	free := make([]int, 0, variables-len(bound))
	for variable := 0; variable < variables; variable++ {
		if !boundFlag[variable] {
			free = append(free, variable)
		}
	}
	cofactorWords := (1<<uint(len(free)) + 63) / 64
	for sharedAssignment := 0; sharedAssignment < 1<<uint(len(shared)); sharedAssignment++ {
		var first, second []uint64
		for uniqueAssignment := 0; uniqueAssignment < 1<<uint(len(unique)); uniqueAssignment++ {
			cofactor := make([]uint64, cofactorWords)
			for freeAssignment := 0; freeAssignment < 1<<uint(len(free)); freeAssignment++ {
				if err := budget.take(); err != nil {
					return false, err
				}
				assignment := uint64(0)
				for i, variable := range shared {
					if sharedAssignment&(1<<uint(i)) != 0 {
						assignment |= 1 << uint(variable)
					}
				}
				for i, variable := range unique {
					if uniqueAssignment&(1<<uint(i)) != 0 {
						assignment |= 1 << uint(variable)
					}
				}
				for i, variable := range free {
					if freeAssignment&(1<<uint(i)) != 0 {
						assignment |= 1 << uint(variable)
					}
				}
				value, err := table.Get(assignment)
				if err != nil {
					return false, err
				}
				if value {
					cofactor[freeAssignment/64] |= 1 << uint(freeAssignment&63)
				}
			}
			if first == nil {
				first = cofactor
				continue
			}
			if equalUint64Words(first, cofactor) {
				continue
			}
			if second == nil {
				second = cofactor
				continue
			}
			if !equalUint64Words(second, cofactor) {
				return false, nil
			}
		}
	}
	return true, nil
}

func equalUint64Words(first, second []uint64) bool {
	if len(first) != len(second) {
		return false
	}
	for i := range first {
		if first[i] != second[i] {
			return false
		}
	}
	return true
}

// FindDauDecompositionSets returns the source-order family of disjoint and
// shared bound sets. Every input must be in the support, matching the source
// precondition. Returned storage is independent of the truth table.
func FindDauDecompositionSets(table *TruthTable, options DauDecompositionSetOptions) ([]DauDecompositionSet, error) {
	if err := table.Check(); err != nil {
		return nil, err
	}
	if table.variables < 3 || table.variables > 16 {
		return nil, ErrDimension
	}
	o, err := normalizeDauDecompositionSetOptions(options)
	if err != nil {
		return nil, err
	}
	support, err := table.Support()
	if err != nil {
		return nil, err
	}
	if len(support) != table.variables {
		return nil, ErrUnsupported
	}
	variables := table.variables
	v2p, p2v := make([]int, variables), make([]int, variables)
	for variable := 0; variable < variables; variable++ {
		v2p[variable], p2v[variable] = variable, variable
	}
	move := func(variable, position int) {
		old := v2p[variable]
		if old == position {
			return
		}
		other := p2v[position]
		p2v[old], p2v[position] = other, variable
		v2p[variable], v2p[other] = position, old
	}
	result := make([]DauDecompositionSet, 0, 32)
	budget := &dauDecompositionCheckBudget{maximum: o.MaxChecks, stop: o.Stop}
	limit := 1 << uint(variables)
	for boundSize := 2; boundSize < variables; boundSize++ {
		for boundMask := 0; boundMask < limit; boundMask++ {
			if bits.OnesCount(uint(boundMask)) != boundSize {
				continue
			}
			position := 0
			for variable := 0; variable < variables; variable++ {
				if boundMask&(1<<uint(variable)) == 0 {
					move(variable, position)
					position++
				}
			}
			bound := append([]int(nil), p2v[variables-boundSize:]...)
			works, err := dauDecompositionSetWorks(table, bound, 0, budget)
			if err != nil {
				return nil, err
			}
			if works {
				if len(result) >= o.MaxSets {
					return nil, ErrLimit
				}
				result = append(result, createDauDecompositionSet(bound, 0))
				continue
			}
			if boundSize == 2 {
				continue
			}
			for sharedSize := 1; sharedSize <= boundSize-2 && sharedSize <= 3; sharedSize++ {
				for sharedMask := 0; sharedMask < 1<<uint(boundSize); sharedMask++ {
					if bits.OnesCount(uint(sharedMask)) != sharedSize {
						continue
					}
					candidate := createDauDecompositionSet(bound, sharedMask)
					if dauDecompositionSetContained(result, candidate) {
						continue
					}
					works, err = dauDecompositionSetWorks(table, bound, sharedMask, budget)
					if err != nil {
						return nil, err
					}
					if works {
						if len(result) >= o.MaxSets {
							return nil, ErrLimit
						}
						result = append(result, candidate)
					}
				}
			}
		}
	}
	return result, nil
}
