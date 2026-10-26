package abc

// Decomposition-manager lifecycle from src/bool/dec/decMan.c.

import (
	"slices"
	"sync"
)

type decNPN4Tables struct {
	canonical   []uint16
	phase       []uint8
	permutation []uint8
	class       []uint8
}

var decNPN4Cache struct {
	sync.Once
	tables decNPN4Tables
	err    error
}

// DecManager owns the factoring scratch and exact four-input NPN metadata used
// by the decomposition package. Construct with NewDecManager; do not copy it.
type DecManager struct {
	mvc    *MvcManager
	cubes  []int
	lits   []int
	tables decNPN4Tables
	closed bool
}

func decFactorial(value int) int {
	result := 1
	for number := 2; number <= value; number++ {
		result *= number
	}
	return result
}

func decPermutations(variables int) [][]int {
	count := decFactorial(variables)
	result := make([][]int, count)
	for index := range result {
		result[index] = make([]int, variables)
	}
	values := make([]int, variables)
	for index := range values {
		values[index] = index
	}
	var fill func([][]int, int, int)
	fill = func(destination [][]int, factorial, remaining int) {
		if remaining == 1 {
			destination[0][0] = values[0]
			return
		}
		nextFactorial := factorial / remaining
		last := remaining - 1
		for current := 0; current < remaining; current++ {
			values[current], values[last] = values[last], values[current]
			start := (remaining - 1 - current) * nextFactorial
			next := destination[start : start+nextFactorial]
			for index := range next {
				next[index][last] = values[last]
			}
			fill(next, nextFactorial, remaining-1)
			values[current], values[last] = values[last], values[current]
		}
	}
	fill(result, count, variables)
	return result
}

func decTransformTruth4(truth uint16, phase int, permutation []int) uint16 {
	var result uint16
	for assignment := 0; assignment < 16; assignment++ {
		source := assignment ^ phase
		if truth&(uint16(1)<<uint(source)) == 0 {
			continue
		}
		target := 0
		for variable := 0; variable < 4; variable++ {
			if assignment&(1<<uint(variable)) != 0 {
				target |= 1 << uint(permutation[variable])
			}
		}
		result |= uint16(1) << uint(target)
	}
	return result
}

func buildDecNPN4Tables() (decNPN4Tables, error) {
	const functions = 1 << 16
	tables := decNPN4Tables{
		canonical:   make([]uint16, functions),
		phase:       make([]uint8, functions),
		permutation: make([]uint8, functions),
		class:       make([]uint8, functions),
	}
	permutations := decPermutations(4)
	classes := 1
	for truth := 1; truth < 1<<15; truth++ {
		if tables.canonical[truth] != 0 {
			if truth <= int(tables.canonical[truth]) {
				return decNPN4Tables{}, ErrInvariant
			}
			class := tables.class[tables.canonical[truth]]
			tables.class[truth] = class
			tables.class[^uint16(truth)] = class
			continue
		}
		if classes > 255 {
			return decNPN4Tables{}, ErrLimit
		}
		tables.class[truth] = uint8(classes)
		classes++
		for phase := 0; phase < 16; phase++ {
			for _, base := range []uint16{uint16(truth), ^uint16(truth)} {
				for permutationIndex, permutation := range permutations {
					transformed := decTransformTruth4(base, phase, permutation)
					if tables.canonical[transformed] == 0 {
						tables.canonical[transformed] = uint16(truth)
						tables.phase[transformed] = uint8(phase)
						tables.permutation[transformed] = uint8(permutationIndex)
						complement := ^transformed
						tables.canonical[complement] = uint16(truth)
						tables.phase[complement] = uint8(phase | 16)
						tables.permutation[complement] = uint8(permutationIndex)
					} else if tables.canonical[transformed] != uint16(truth) {
						return decNPN4Tables{}, ErrInvariant
					}
				}
			}
		}
	}
	tables.phase[functions-1] = 16
	if classes != 222 {
		return decNPN4Tables{}, ErrInvariant
	}
	return tables, nil
}

// NewDecManager implements Dec_ManStart with Go-owned scratch and copied NPN
// tables. The package cache is immutable after initialization; every manager
// receives independent table storage.
func NewDecManager() (*DecManager, error) {
	decNPN4Cache.Do(func() {
		decNPN4Cache.tables, decNPN4Cache.err = buildDecNPN4Tables()
	})
	if decNPN4Cache.err != nil {
		return nil, decNPN4Cache.err
	}
	tables := decNPN4Cache.tables
	return &DecManager{
		mvc:   NewMvcManager(),
		cubes: make([]int, 0, 8),
		lits:  make([]int, 0, 8),
		tables: decNPN4Tables{
			canonical:   slices.Clone(tables.canonical),
			phase:       slices.Clone(tables.phase),
			permutation: slices.Clone(tables.permutation),
			class:       slices.Clone(tables.class),
		},
	}, nil
}

func (manager *DecManager) check() error {
	if manager == nil {
		return ErrDimension
	}
	if manager.closed || manager.mvc == nil {
		return ErrInvariant
	}
	return manager.mvc.check()
}

// NPN4 returns the representative, input/output phase, native permutation
// index and class number for one four-input truth table.
func (manager *DecManager) NPN4(truth uint16) (uint16, uint8, uint8, uint8, error) {
	if err := manager.check(); err != nil {
		return 0, 0, 0, 0, err
	}
	return manager.tables.canonical[truth], manager.tables.phase[truth], manager.tables.permutation[truth], manager.tables.class[truth], nil
}

// NPN4Tables returns independent snapshots in truth-table index order.
func (manager *DecManager) NPN4Tables() ([]uint16, []uint8, []uint8, []uint8, error) {
	if err := manager.check(); err != nil {
		return nil, nil, nil, nil, err
	}
	return slices.Clone(manager.tables.canonical), slices.Clone(manager.tables.phase), slices.Clone(manager.tables.permutation), slices.Clone(manager.tables.class), nil
}

// Close implements Dec_ManStop. Repeated or nil closes are errors instead of
// reproducing invalid native frees.
func (manager *DecManager) Close() error {
	if err := manager.check(); err != nil {
		return err
	}
	if err := manager.mvc.Close(); err != nil {
		return err
	}
	manager.closed = true
	manager.mvc = nil
	manager.cubes, manager.lits = nil, nil
	manager.tables = decNPN4Tables{}
	return nil
}
