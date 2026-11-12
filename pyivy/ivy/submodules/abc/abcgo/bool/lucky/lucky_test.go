package lucky

import (
	"math/rand"
	"slices"
	"testing"
)

func bit(table []uint64, index int) bool { return table[index>>6]&(1<<uint(index&63)) != 0 }
func setBit(table []uint64, index int)   { table[index>>6] |= 1 << uint(index&63) }
func randomTruth(rng *rand.Rand, vars int32) []uint64 {
	table := make([]uint64, Kit_TruthWordNum_64bit(vars))
	for i := range table {
		table[i] = rng.Uint64()
	}
	return table
}
func stretchedTruth(value uint64, vars int32) uint64 {
	width := uint(1) << uint(vars)
	for width < 64 {
		value |= value << width
		width *= 2
	}
	return value
}
func TestSwapsPhasesAndCofactorCounts(t *testing.T) {
	rng := rand.New(rand.NewSource(7))
	for vars := int32(1); vars <= 10; vars++ {
		original := randomTruth(rng, vars)
		for variable := int32(0); variable < vars; variable++ {
			table := slices.Clone(original)
			Kit_TruthChangePhase_64bit(table, vars, variable)
			for i := 0; i < len(table)*64; i++ {
				if bit(table, i) != bit(original, i^(1<<uint(variable))) {
					t.Fatal("phase", vars, variable, i)
				}
			}
			Kit_TruthChangePhase_64bit(table, vars, variable)
			if !slices.Equal(table, original) {
				t.Fatal("phase inverse")
			}
		}
		for i := int32(0); i < vars; i++ {
			for j := int32(0); j < vars; j++ {
				table := slices.Clone(original)
				swap_ij(table, vars, i, j)
				for a := 0; a < len(table)*64; a++ {
					index := a
					if (a>>uint(i)^a>>uint(j))&1 != 0 {
						index ^= 1<<uint(i) | 1<<uint(j)
					}
					if bit(table, a) != bit(original, index) {
						t.Fatal("swap", vars, i, j, a)
					}
				}
			}
		}
		counts := make([]int32, vars)
		Kit_TruthCountOnesInCofs_64bit(original, vars, counts)
		for variable := int32(0); variable < vars; variable++ {
			var want int32
			for i := 0; i < len(original)*64; i++ {
				if i&(1<<uint(variable)) == 0 && bit(original, i) {
					want++
				}
			}
			if counts[variable] != want {
				t.Fatal("counts", vars, variable, counts[variable], want)
			}
		}
	}
}
func TestPermutationAndGrayEnumerators(t *testing.T) {
	for n := int32(0); n <= 7; n++ {
		info := setPermInfoPtr(n)
		permutation := make([]int32, n)
		for i := range permutation {
			permutation[i] = int32(i)
		}
		seen := map[string]bool{string(runes(permutation)): true}
		for i := info.totalSwaps - 1; i >= 0; i-- {
			j := info.swapArray[i]
			permutation[j], permutation[j+1] = permutation[j+1], permutation[j]
			key := string(runes(permutation))
			if seen[key] {
				t.Fatal("duplicate permutation", n, key)
			}
			seen[key] = true
		}
		if len(seen) != int(factorial(n)) {
			t.Fatal(n, len(seen))
		}
		phase := int32(0)
		phases := map[int32]bool{0: true}
		for i := info.totalFlips - 1; i >= 0; i-- {
			phase ^= 1 << uint(info.flipArray[i])
			if phases[phase] {
				t.Fatal("duplicate Gray phase")
			}
			phases[phase] = true
		}
		if len(phases) != 1<<uint(n) {
			t.Fatal(n, len(phases))
		}
	}
}
func runes(values []int32) []rune {
	result := make([]rune, len(values))
	for i, v := range values {
		result[i] = rune('a' + v)
	}
	return result
}
func exhaustiveMinimum(original uint64, vars int32) uint64 {
	best := ^uint64(0)
	perm := make([]int32, vars)
	used := make([]bool, vars)
	var visit func(int32)
	visit = func(position int32) {
		if position < vars {
			for i := int32(0); i < vars; i++ {
				if !used[i] {
					used[i] = true
					perm[position] = i
					visit(position + 1)
					used[i] = false
				}
			}
			return
		}
		for phase := 0; phase < 1<<uint(vars); phase++ {
			var candidate uint64
			for assignment := 0; assignment < 64; assignment++ {
				source := 0
				for i := int32(0); i < vars; i++ {
					if (assignment>>uint(i)^phase>>uint(i))&1 != 0 {
						source |= 1 << uint(perm[i])
					}
				}
				if original>>uint(source)&1 != 0 {
					candidate |= 1 << uint(assignment)
				}
			}
			best = min(best, candidate, ^candidate)
		}
	}
	visit(0)
	return best
}
func TestSimpleMinimumAllThreeVariableTables(t *testing.T) {
	info := setPermInfoPtr(3)
	for value := uint64(0); value < 256; value++ {
		original := stretchedTruth(value, 3)
		table, aux, minimal := []uint64{original}, make([]uint64, 1), make([]uint64, 1)
		simpleMinimal(table, aux, minimal, info, 3)
		want := exhaustiveMinimum(original, 3)
		if table[0] != want {
			t.Fatalf("truth %x: got %x, want %x", original, table[0], want)
		}
	}
}
func TestCanonicalMetadataReconstructs(t *testing.T) {
	rng := rand.New(rand.NewSource(17))
	for _, vars := range []int32{6, 7, 8, 10, 16} {
		for _, second := range []bool{false, true} {
			for trial := 0; trial < 30; trial++ {
				original := randomTruth(rng, vars)
				table := slices.Clone(original)
				permutation := make([]byte, 16)
				resetPCanonPermArray(permutation, 16)
				phase := luckyFast(table, vars, permutation, second)

				if luckyCheck(slices.Clone(table), original, vars, slices.Clone(permutation), phase) != 0 {
					t.Fatal("metadata reconstruction", vars, second, trial, phase, permutation)
				}
			}
		}
	}
}

func TestFastCanonicalizationCReferenceGoldens(t *testing.T) {
	for _, tc := range []struct {
		vars          int32
		second        bool
		input, output []uint64
		phase         uint32
		permutation   string
	}{
		{6, false, []uint64{0x0123456789abcdef}, []uint64{0x0123456789abcdef}, 384, "abcdefghijklmnop"},
		{6, true, []uint64{0x6996966996696996}, []uint64{0x6996966996696996}, 447, "abcdefghijklmnop"},
		{7, false, []uint64{0x0123456789abcdef, 0xfedcba9876543210}, []uint64{0xfff0ccaa00f0ccaa, 0x000f3355ff0f3355}, 807, "cdeabfghijklmnop"},
		{8, true, []uint64{0x0123456789abcdef, 0xfedcba9876543210, 0x8888888877777777, 0x0123012301230123}, []uint64{0xbbbb888803cf03cf, 0xcccc3333ffff0000, 0x44447777fc30fc30, 0x0000000003cf03cf}, 531, "eadcfbhgijklmnop"},
	} {
		table := slices.Clone(tc.input)
		perm := make([]byte, 16)
		resetPCanonPermArray(perm, 16)
		phase := luckyFast(table, tc.vars, perm, tc.second)
		if phase != tc.phase || string(perm) != tc.permutation || !slices.Equal(table, tc.output) {
			t.Fatal(tc.vars, table, phase, string(perm))
		}
	}
}
func TestSimpleMinimalSmallSupport(t *testing.T) {
	for vars := int32(0); vars <= 1; vars++ {
		table := []uint64{^uint64(0)}
		aux, minimal := make([]uint64, 1), make([]uint64, 1)
		simpleMinimal(table, aux, minimal, setPermInfoPtr(vars), vars)
		if table[0] != 0 {
			t.Fatal(vars, table)
		}
	}
}
func TestSemiCanonicalNewMetadata(t *testing.T) {
	rng := rand.New(rand.NewSource(33))
	for vars := int32(1); vars <= 9; vars++ {
		words := int(Kit_TruthWordNum_64bit(vars))
		input := make([]uint32, 2*words)
		for i := range input {
			input[i] = rng.Uint32()
		}
		if vars < 6 {
			input[0] = uint32(stretchedTruth(uint64(input[0])&((1<<uint(1<<uint(vars)))-1), vars))
		}
		before := make([]uint64, words)
		for i := range before {
			if vars < 6 {
				before[i] = uint64(input[0])<<32 | uint64(input[0])
			} else {
				before[i] = uint64(input[2*i]) | uint64(input[2*i+1])<<32
			}
		}
		perm := make([]byte, vars)
		resetPCanonPermArray(perm, vars)
		phase := Kit_TruthSemiCanonicize_new(input, nil, vars, perm)
		after := make([]uint64, words)
		for i := range after {
			if vars < 6 {
				after[i] = uint64(input[0])<<32 | uint64(input[0])
			} else {
				after[i] = uint64(input[2*i]) | uint64(input[2*i+1])<<32
			}
		}
		if luckyCheck(after, before, vars, slices.Clone(perm), phase) != 0 {
			t.Fatal(vars, phase, string(perm))
		}
	}
}

func TestGroupedMinimumAgainstEnumeration(t *testing.T) {
	infos := []*permInfo{setPermInfoPtr(0), setPermInfoPtr(1), setPermInfoPtr(2)}
	groups := []int32{2, 1}
	for _, flipOutput := range []int32{0, 1} {
		for _, flipInput := range []int32{0, 1} {
			for value := uint64(0); value < 256; value += 7 {
				original := stretchedTruth(value, 3)
				table, aux, minimal := []uint64{original}, make([]uint64, 1), make([]uint64, 1)
				simpleMinimalGroups(table, aux, minimal, groups, 2, infos, 3, flipOutput, flipInput)
				want := ^uint64(0)
				for swap := 0; swap < 2; swap++ {
					for phase := 0; phase < 8; phase++ {
						if flipInput == 0 && phase&4 != 0 {
							continue
						}
						candidate := []uint64{original}
						if swap != 0 {
							Kit_TruthSwapAdjacentVars_64bit(candidate, 3, 0)
						}
						for variable := int32(0); variable < 3; variable++ {
							if phase&(1<<uint(variable)) != 0 {
								Kit_TruthChangePhase_64bit(candidate, 3, variable)
							}
						}
						want = min(want, candidate[0])
						if flipOutput != 0 {
							want = min(want, ^candidate[0])
						}
					}
				}
				if table[0] != want {
					t.Fatalf("group result vars=3 table=%x output=%d input=%d: %x != %x", original, flipOutput, flipInput, table[0], want)
				}
			}
		}
	}
}
