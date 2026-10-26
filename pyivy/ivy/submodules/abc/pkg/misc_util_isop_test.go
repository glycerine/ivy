package abc

// Standalone tests for the remaining src/misc/util/utilIsop.c operations.

import (
	"errors"
	"slices"
	"testing"
)

func exclusiveCubeCoverValue(cubes []uint32, assignment, variables int) bool {
	value := false
	for _, cube := range cubes {
		if cubeMatchesAssignment(cube, assignment, variables) {
			value = !value
		}
	}
	return value
}

func cubeLiteralCount(cubes []uint32, variables int) (int, error) {
	count := 0
	for _, cube := range cubes {
		for variable := 0; variable < variables; variable++ {
			switch cube >> uint(2*variable) & 3 {
			case 1, 2:
				count++
			case 3:
				return 0, ErrDimension
			}
		}
	}
	return count, nil
}

func TestTruthESOPExhaustiveThreeVariables(t *testing.T) {
	for function := 0; function < 256; function++ {
		table := kitSmallTruth(t, 3, uint64(function))
		before := table.Words()
		result, err := table.ESOP(TruthESOPOptions{})
		if err != nil || result.Table == nil {
			t.Fatal(function, result, err)
		}
		for assignment := 0; assignment < 8; assignment++ {
			want := function>>uint(assignment)&1 != 0
			got := exclusiveCubeCoverValue(result.Cubes, assignment, 3)
			tableValue, tableErr := result.Table.Get(uint64(assignment))
			if got != want || tableValue != want || tableErr != nil {
				t.Fatal(function, assignment, result, got, tableValue, tableErr)
			}
		}
		literals, err := cubeLiteralCount(result.Cubes, 3)
		if err != nil || literals != result.Literals || !slices.Equal(table.Words(), before) {
			t.Fatal(function, literals, result, err)
		}
		if _, err := result.Format(3); err != nil {
			t.Fatal(function, err)
		}
	}
}

func TestGreedyISOPIntervalsExhaustiveThreeVariables(t *testing.T) {
	for allowedValue := 0; allowedValue < 256; allowedValue++ {
		allowed := kitSmallTruth(t, 3, uint64(allowedValue))
		for onValue := allowedValue; ; onValue = (onValue - 1) & allowedValue {
			on := kitSmallTruth(t, 3, uint64(onValue))
			result, err := on.GreedyISOPInterval(allowed, TruthGreedyISOPOptions{})
			if err != nil || result.Table == nil {
				t.Fatal(onValue, allowedValue, result, err)
			}
			for assignment := 0; assignment < 8; assignment++ {
				value := kitCoverValue(result.Cubes, assignment, 3)
				required := onValue>>uint(assignment)&1 != 0
				permitted := allowedValue>>uint(assignment)&1 != 0
				tableValue, tableErr := result.Table.Get(uint64(assignment))
				if required && !value || value && !permitted || tableValue != value || tableErr != nil {
					t.Fatal(onValue, allowedValue, assignment, result, value, tableValue, tableErr)
				}
			}
			literals, err := cubeLiteralCount(result.Cubes, 3)
			if err != nil || literals != result.Literals {
				t.Fatal(onValue, allowedValue, result, literals, err)
			}
			if onValue == 0 {
				break
			}
		}
	}
}

func TestCubeCoverTruthModesAndISOPCNF(t *testing.T) {
	cubes := []uint32{1, 2, 1 | 2<<2}
	ordinary, err := BuildCubeCoverTruth(2, cubes, false, false)
	if err != nil {
		t.Fatal(err)
	}
	exclusive, err := BuildCubeCoverTruth(2, cubes, true, false)
	if err != nil {
		t.Fatal(err)
	}
	complement, err := BuildCubeCoverTruth(2, cubes, true, true)
	if err != nil {
		t.Fatal(err)
	}
	for assignment := 0; assignment < 4; assignment++ {
		ordinaryValue, _ := ordinary.Get(uint64(assignment))
		exclusiveValue, _ := exclusive.Get(uint64(assignment))
		complementValue, _ := complement.Get(uint64(assignment))
		if ordinaryValue != kitCoverValue(cubes, assignment, 2) || exclusiveValue != exclusiveCubeCoverValue(cubes, assignment, 2) || complementValue == exclusiveValue {
			t.Fatal(assignment, ordinaryValue, exclusiveValue, complementValue)
		}
	}
	if _, err := BuildCubeCoverTruth(1, []uint32{3}, false, false); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	for function := 0; function < 256; function++ {
		// Abc_IsopCnf requires the highest variable to be in the support.
		if function&15 == function>>4&15 {
			continue
		}
		table := kitSmallTruth(t, 3, uint64(function))
		cnf, err := table.ISOPCNF(TruthISOPCNFOptions{})
		if err != nil || cnf.OnCubes < 0 || cnf.OnCubes > len(cnf.Cubes) {
			t.Fatal(function, cnf, err)
		}
		for assignment := 0; assignment < 16; assignment++ {
			original := assignment & 7
			selector := assignment>>3&1 != 0
			want := (function>>uint(original)&1 != 0) != selector
			if got := kitCoverValue(cnf.Cubes, assignment, 4); got != want {
				t.Fatal(function, assignment, cnf, got, want)
			}
		}
	}
}

func TestUtilISOPLimitsErrorsAndPinnedOrder(t *testing.T) {
	table := kitSmallTruth(t, 4, 0x6996)
	esop, err := table.ESOP(TruthESOPOptions{})
	if err != nil || !slices.Equal(esop.Cubes, []uint32{1, 4, 16, 64}) || esop.Literals != 4 {
		t.Fatal(esop, err)
	}
	greedy, err := table.GreedyISOPInterval(table, TruthGreedyISOPOptions{})
	if err != nil || len(greedy.Cubes) == 0 {
		t.Fatal(greedy, err)
	}
	for _, options := range []TruthESOPOptions{
		{MaxCubes: 1}, {MaxElements: 1}, {MaxVisits: 1}, {MaxCubes: -1}, {MaxElements: -1}, {MaxVisits: -1}, {Stop: func() bool { return true }},
	} {
		if result, err := table.ESOP(options); err == nil || result.Cubes != nil {
			t.Fatal(options, result, err)
		}
	}
	for _, options := range []TruthGreedyISOPOptions{
		{MaxCubes: 1}, {MaxElements: 1}, {MaxVisits: 1}, {MaxCubes: -1}, {Stop: func() bool { return true }},
	} {
		if result, err := table.GreedyISOPInterval(table, options); err == nil || result.Cubes != nil {
			t.Fatal(options, result, err)
		}
	}
	badAllowed := kitSmallTruth(t, 4, 0)
	if result, err := table.GreedyISOPInterval(badAllowed, TruthGreedyISOPOptions{}); !errors.Is(err, ErrDimension) || result.Cubes != nil {
		t.Fatal(result, err)
	}
	constant := kitSmallTruth(t, 3, 0xff)
	if _, err := constant.ISOPCNF(TruthISOPCNFOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err := table.ISOPCNF(TruthISOPCNFOptions{MaxCubes: 1}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	wide, _ := NewConstantTruthTable(17, false, 0)
	if _, err := wide.ESOP(TruthESOPOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
}

func FuzzTruthUtilCovers(f *testing.F) {
	f.Add(uint8(4), uint64(0x6996), uint64(0xffff))
	f.Add(uint8(3), uint64(0x81), uint64(0xaa))
	f.Fuzz(func(t *testing.T, rawVariables uint8, rawOn, rawExtra uint64) {
		variables := int(rawVariables % 7)
		bits := 1 << uint(variables)
		mask := ^uint64(0)
		if bits < 64 {
			mask = uint64(1)<<uint(bits) - 1
		}
		onValue := rawOn & mask
		allowedValue := onValue | rawExtra&mask
		on := kitSmallTruth(t, variables, onValue)
		allowed := kitSmallTruth(t, variables, allowedValue)
		esop, err := on.ESOP(TruthESOPOptions{})
		if err != nil {
			t.Fatal(variables, onValue, esop, err)
		}
		greedy, err := on.GreedyISOPInterval(allowed, TruthGreedyISOPOptions{})
		if err != nil {
			t.Fatal(variables, onValue, allowedValue, greedy, err)
		}
		for assignment := 0; assignment < bits; assignment++ {
			want := onValue>>uint(assignment)&1 != 0
			if exclusiveCubeCoverValue(esop.Cubes, assignment, variables) != want {
				t.Fatal("esop", variables, assignment, esop)
			}
			got := kitCoverValue(greedy.Cubes, assignment, variables)
			permitted := allowedValue>>uint(assignment)&1 != 0
			if want && !got || got && !permitted {
				t.Fatal("greedy", variables, assignment, greedy)
			}
		}
	})
}
