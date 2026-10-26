package abc

// Go-only exhaustive prime/subset models for src/misc/espresso/exact.c.

import (
	"errors"
	"slices"
	"sort"
	"testing"
)

func espressoMintermCover(t testing.TB, variables int, truth uint64) *EspressoSetFamily {
	t.Helper()
	family, _ := NewEspressoSetFamily(1<<uint(variables), 2*variables)
	for assignment := 0; assignment < 1<<uint(variables); assignment++ {
		if truth>>uint(assignment)&1 == 0 {
			continue
		}
		cube, _ := NewEspressoSet(2 * variables)
		for variable := 0; variable < variables; variable++ {
			_ = cube.Set(2*variable+((assignment>>uint(variable))&1), true)
		}
		_ = family.Add(cube)
	}
	return family
}

func espressoBrutePrimes(layout *EspressoCubeLayout, onTruth, dcTruth uint64) []uint64 {
	allowed := onTruth | dcTruth
	limit := 1
	for range layout.VariableCount() {
		limit *= 3
	}
	result := make([]uint64, 0)
	for code := 0; code < limit; code++ {
		remaining, mask := code, uint64(0)
		for variable := 0; variable < layout.VariableCount(); variable++ {
			mask |= uint64(remaining%3+1) << uint(2*variable)
			remaining /= 3
		}
		truth := espressoCubeTruth(layout, mask)
		if truth == 0 || truth&^allowed != 0 || truth&onTruth == 0 {
			continue
		}
		prime := true
		for variable := 0; variable < layout.VariableCount(); variable++ {
			if mask>>(2*variable)&3 == 3 {
				continue
			}
			expanded := mask | uint64(3)<<uint(2*variable)
			if espressoCubeTruth(layout, expanded)&^allowed == 0 {
				prime = false
				break
			}
		}
		if prime {
			result = append(result, mask)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result
}

func espressoBruteExactObjective(layout *EspressoCubeLayout, primes []uint64, onTruth, dcTruth uint64, weighted bool) (primary, secondary int) {
	primary, secondary = int(^uint(0)>>1), int(^uint(0)>>1)
	for subset := uint64(0); subset < uint64(1)<<uint(len(primes)); subset++ {
		truth, count, literals := dcTruth, 0, 0
		for index, prime := range primes {
			if subset>>uint(index)&1 == 0 {
				continue
			}
			truth |= espressoCubeTruth(layout, prime)
			count++
			literals += 2*layout.VariableCount() - bitsCount64(prime)
		}
		if onTruth&^truth != 0 {
			continue
		}
		candidatePrimary, candidateSecondary := count, literals
		if weighted {
			candidatePrimary, candidateSecondary = literals, count
		}
		if espressoExactBetter(candidatePrimary, candidateSecondary, primary, secondary) {
			primary, secondary = candidatePrimary, candidateSecondary
		}
	}
	return primary, secondary
}

func TestEspressoExactMinimizePrimeSelectionAndOwnership(t *testing.T) {
	variables := 3
	layout, _ := NewEspressoCubeLayout(slices.Repeat([]int{2}, variables), variables)
	onTruth, dcTruth := uint64(0b11101000), uint64(0b00000100)
	universe := espressoTruthUniverse(layout)
	on := espressoMintermCover(t, variables, onTruth)
	dc := espressoMintermCover(t, variables, dcTruth)
	off := espressoMintermCover(t, variables, universe&^(onTruth|dcTruth))
	before := espressoFamilyMasks(t, on)
	for _, weighted := range []bool{false, true} {
		var result EspressoExactResult
		var err error
		if weighted {
			result, err = layout.MinimizeExactLiterals(on, dc, off, true, EspressoExactOptions{})
		} else {
			result, err = layout.MinimizeExact(on, dc, off, true, EspressoExactOptions{})
		}
		primes := espressoBrutePrimes(layout, onTruth, dcTruth)
		wantPrimary, _ := espressoBruteExactObjective(layout, primes, onTruth, dcTruth, weighted)
		if err != nil || result.Cost != wantPrimary || espressoFamilyTruth(t, layout, result.Cover)&^dcTruth != onTruth&^dcTruth {
			t.Fatal(weighted, result.Cost, wantPrimary, espressoFamilyMasks(t, result.Cover), err)
		}
		if !slices.Equal(espressoSortedMasks(t, result.Primes), primes) {
			t.Fatal(espressoSortedMasks(t, result.Primes), primes)
		}
		result.Cover.sets[0].words[0] = 0
		if !slices.Equal(espressoFamilyMasks(t, on), before) {
			t.Fatal("source cover mutated")
		}
	}
}

func TestEspressoExactHeuristicAndEmptyFunction(t *testing.T) {
	layout, _ := NewEspressoCubeLayout([]int{2, 2}, 2)
	on := espressoMintermCover(t, 2, 0b0111)
	dc := espressoMintermCover(t, 2, 0)
	off := espressoMintermCover(t, 2, 0b1000)
	result, err := layout.MinimizeExact(on, dc, off, false, EspressoExactOptions{})
	if err != nil || result.Exact || result.Weighted || espressoFamilyTruth(t, layout, result.Cover) != 0b0111 {
		t.Fatal(result, err)
	}
	empty := espressoMintermCover(t, 2, 0)
	fullOff := espressoMintermCover(t, 2, 0b1111)
	emptyResult, err := layout.MinimizeExact(empty, empty, fullOff, true, EspressoExactOptions{})
	if err != nil || len(espressoFamilyMasks(t, emptyResult.Cover)) != 0 || emptyResult.Cost != 0 {
		t.Fatal(emptyResult, err)
	}
}

func TestEspressoExactErrorsLimitsAndCancellation(t *testing.T) {
	layout, _ := NewEspressoCubeLayout([]int{2, 2}, 2)
	on := espressoMintermCover(t, 2, 0b0111)
	dc := espressoMintermCover(t, 2, 0)
	off := espressoMintermCover(t, 2, 0b1000)
	checks := []EspressoExactOptions{
		{MaxCubes: -1}, {MaxCubes: 1}, {MaxOperations: 1}, {MaxDepth: -1}, {MaxSearchNodes: 1},
		{Stop: func() bool { return true }},
	}
	for _, options := range checks {
		result, err := layout.MinimizeExact(on, dc, off, true, options)
		if err == nil || result.Cover != nil || !(errors.Is(err, ErrDimension) || errors.Is(err, ErrLimit) || errors.Is(err, ErrCancelled)) {
			t.Fatal(options, result, err)
		}
	}
	wrong := espressoMapFamily(t, 3)
	if result, err := layout.MinimizeExact(on, dc, wrong, true, EspressoExactOptions{}); !errors.Is(err, ErrDimension) || result.Cover != nil {
		t.Fatal(result, err)
	}
	overlap := espressoMintermCover(t, 2, 0b0001)
	if result, err := layout.MinimizeExact(on, overlap, off, true, EspressoExactOptions{}); !errors.Is(err, ErrInvariant) || result.Cover != nil {
		t.Fatal(result, err)
	}
}

func FuzzEspressoExactMinimize(f *testing.F) {
	f.Add([]byte{3, 0xe8, 0x04})
	f.Add([]byte{2, 0x07, 0})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) < 3 {
			return
		}
		variables := int(data[0]%3) + 1
		assignments := 1 << uint(variables)
		universe := uint64(1)<<uint(assignments) - 1
		onTruth := uint64(data[1]) & universe
		dcTruth := uint64(data[2]) & universe &^ onTruth
		offTruth := universe &^ (onTruth | dcTruth)
		layout, _ := NewEspressoCubeLayout(slices.Repeat([]int{2}, variables), variables)
		on := espressoMintermCover(t, variables, onTruth)
		dc := espressoMintermCover(t, variables, dcTruth)
		off := espressoMintermCover(t, variables, offTruth)
		primes := espressoBrutePrimes(layout, onTruth, dcTruth)
		for _, weighted := range []bool{false, true} {
			var result EspressoExactResult
			var err error
			if weighted {
				result, err = layout.MinimizeExactLiterals(on, dc, off, true, EspressoExactOptions{MaxSearchNodes: 1 << 16})
			} else {
				result, err = layout.MinimizeExact(on, dc, off, true, EspressoExactOptions{MaxSearchNodes: 1 << 16})
			}
			if err != nil {
				t.Fatal(err)
			}
			wantPrimary, _ := espressoBruteExactObjective(layout, primes, onTruth, dcTruth, weighted)
			gotTruth := espressoFamilyTruth(t, layout, result.Cover)
			if result.Cost != wantPrimary || onTruth&^(gotTruth|dcTruth) != 0 || gotTruth&offTruth != 0 || !slices.Equal(espressoSortedMasks(t, result.Primes), primes) {
				t.Fatal(variables, onTruth, dcTruth, weighted, result.Cost, wantPrimary, espressoFamilyMasks(t, result.Cover), primes)
			}
		}
	})
}
