package abc

// Go-only tests for src/bool/kit/kitTruth.c and kitPerm.c.

import (
	"errors"
	"math/bits"
	"slices"
	"testing"
)

func kitKernelTable(tb testing.TB, variables int, truth uint64) *TruthTable {
	tb.Helper()
	word, e := NormalizeSmallTruthWord(truth, variables)
	if e != nil {
		tb.Fatal(e)
	}
	table, e := TruthTableFromWords(variables, []uint64{word}, 0)
	if e != nil {
		tb.Fatal(e)
	}
	return table
}

func kitKernelLogical(t *TruthTable) uint64 {
	var result uint64
	for assignment := 0; assignment < 1<<uint(t.variables); assignment++ {
		value, _ := t.Get(uint64(assignment))
		if value {
			result |= uint64(1) << uint(assignment)
		}
	}
	return result
}

func kitKernelPermuteModel(truth uint64, variables int, permutation []int) uint64 {
	var result uint64
	for assignment := 0; assignment < 1<<uint(variables); assignment++ {
		source := 0
		for variable, target := range permutation {
			if assignment&(1<<uint(target)) != 0 {
				source |= 1 << uint(variable)
			}
		}
		if truth&(uint64(1)<<uint(source)) != 0 {
			result |= uint64(1) << uint(assignment)
		}
	}
	return result
}

func kitKernelCofactorModel(truth uint64, variables, variable int, value bool) uint64 {
	var result uint64
	for assignment := 0; assignment < 1<<uint(variables); assignment++ {
		source := assignment &^ (1 << uint(variable))
		if value {
			source |= 1 << uint(variable)
		}
		if truth&(uint64(1)<<uint(source)) != 0 {
			result |= uint64(1) << uint(assignment)
		}
	}
	return result
}

func TestKitTruthKernelExhaustiveFourInputs(t *testing.T) {
	for variables := 0; variables <= 4; variables++ {
		functionCount := 1 << uint(1<<uint(variables))
		for truth := 0; truth < functionCount; truth++ {
			table := kitKernelTable(t, variables, uint64(truth))
			mask, e := KitTruthSupportMask(table)
			if e != nil {
				t.Fatal(e)
			}
			var wantMask uint32
			for variable := 0; variable < variables; variable++ {
				zero := kitKernelCofactorModel(uint64(truth), variables, variable, false)
				one := kitKernelCofactorModel(uint64(truth), variables, variable, true)
				if zero != one {
					wantMask |= uint32(1) << uint(variable)
				}
				for _, value := range []bool{false, true} {
					cofactor, x := table.Cofactor(variable, value)
					if x != nil || kitKernelLogical(cofactor) != kitKernelCofactorModel(uint64(truth), variables, variable, value) {
						t.Fatalf("cofactor n=%d f=%x v=%d b=%v got=%x err=%v", variables, truth, variable, value, kitKernelLogical(cofactor), x)
					}
				}
				unique, x := KitTruthUnique(table, variable)
				if x != nil {
					t.Fatal(x)
				}
				var wantUnique uint64
				for assignment := 0; assignment < 1<<uint(variables); assignment++ {
					other := assignment ^ (1 << uint(variable))
					if (truth>>uint(assignment))&1 != (truth>>uint(other))&1 {
						wantUnique |= uint64(1) << uint(assignment)
					}
				}
				if kitKernelLogical(unique) != wantUnique {
					t.Fatalf("unique n=%d f=%x v=%d got=%x want=%x", variables, truth, variable, kitKernelLogical(unique), wantUnique)
				}
				difference, x := KitTruthBooleanDifferenceCount(table, variable)
				if x != nil || difference != bits.OnesCount64(wantUnique)/2*(32/(1<<uint(variables))) {
					t.Fatalf("difference n=%d f=%x v=%d got=%d err=%v", variables, truth, variable, difference, x)
				}
			}
			if mask != wantMask {
				t.Fatalf("support n=%d f=%x got=%x want=%x", variables, truth, mask, wantMask)
			}
		}
	}
}

func TestKitTruthPermutationStretchQuantificationAndMux(t *testing.T) {
	table := kitKernelTable(t, 4, 0x96a5)
	permutation := []int{2, 0, 3, 1}
	originalPermutation := slices.Clone(permutation)
	permuted, e := KitTruthPermute(table, permutation, KitTruthOptions{})
	if e != nil || kitKernelLogical(permuted) != kitKernelPermuteModel(0x96a5, 4, permutation) {
		t.Fatalf("permutation got=%x err=%v", kitKernelLogical(permuted), e)
	}
	if !slices.Equal(permutation, originalPermutation) || kitKernelLogical(table) != 0x96a5 {
		t.Fatal("permutation mutated caller state")
	}
	swapped, e := KitTruthSwapAdjacent(table, 1)
	if e != nil || kitKernelLogical(swapped) != kitKernelPermuteModel(0x96a5, 4, []int{0, 2, 1, 3}) {
		t.Fatalf("adjacent got=%x err=%v", kitKernelLogical(swapped), e)
	}

	small := kitKernelTable(t, 2, 0x8)
	stretched, e := KitTruthStretch(small, 5, 1<<1|1<<4)
	if e != nil {
		t.Fatal(e)
	}
	for assignment := 0; assignment < 32; assignment++ {
		got, _ := stretched.Get(uint64(assignment))
		want := assignment&(1<<1) != 0 && assignment&(1<<4) != 0
		if got != want {
			t.Fatalf("stretch assignment=%d got=%v", assignment, got)
		}
	}
	shrunk, e := KitTruthShrink(stretched, 1<<1|1<<4)
	if e != nil || !shrunk.Equal(small) {
		t.Fatalf("shrink err=%v", e)
	}

	x0, _ := NewVariableTruthTable(4, 0, 0)
	x1, _ := NewVariableTruthTable(4, 1, 0)
	x2, _ := NewVariableTruthTable(4, 2, 0)
	and, _ := x0.And(x1, false, false)
	function, _ := and.Or(x2)
	exist, e := KitTruthQuantifySet(function, 1<<0|1<<1, false, KitTruthOptions{})
	if e != nil || !exist.IsConstant(true) {
		t.Fatalf("exist got=%x err=%v", kitKernelLogical(exist), e)
	}
	forall, e := KitTruthQuantifySet(function, 1<<0|1<<1, true, KitTruthOptions{})
	if e != nil || kitKernelLogical(forall) != kitKernelLogical(x2) {
		t.Fatalf("forall got=%x err=%v", kitKernelLogical(forall), e)
	}
	mux, e := KitTruthMuxVariable(x0, x1, 2, false)
	control, _ := NewVariableTruthTable(4, 2, 0)
	wantMux, _ := MuxTruthTables(control, x1, x0)
	if e != nil || !mux.Equal(wantMux) {
		t.Fatalf("mux err=%v", e)
	}
	phaseMux, e := KitTruthMuxVariable(x0, x1, 2, true)
	wantPhase, _ := MuxTruthTables(control, x1, x0.Not())
	if e != nil || !phaseMux.Equal(wantPhase) {
		t.Fatalf("phase mux err=%v", e)
	}
}

func TestKitTruthSymmetryVacuityBestProfileHashAndHex(t *testing.T) {
	x0, _ := NewVariableTruthTable(4, 0, 0)
	x1, _ := NewVariableTruthTable(4, 1, 0)
	x2, _ := NewVariableTruthTable(4, 2, 0)
	xor, _ := x0.Xor(x1, false)
	symmetric, e := KitTruthVariablesSymmetric(xor, 0, 1)
	if e != nil || !symmetric {
		t.Fatalf("symmetry=%v err=%v", symmetric, e)
	}
	antiInput, _ := x0.Xor(x1, true)
	anti, e := KitTruthVariablesAntiSymmetric(antiInput, 0, 1)
	if e != nil || !anti {
		t.Fatalf("antisymmetry=%v err=%v", anti, e)
	}

	onset, _ := x0.And(x2, false, false)
	offset, _ := x0.Not().And(x2, false, false)
	vacuous, e := KitTruthVariableVacuous(onset, offset, 1)
	if e != nil || !vacuous {
		t.Fatalf("vacuous=%v err=%v", vacuous, e)
	}
	vacuous, e = KitTruthVariableVacuous(onset, offset, 0)
	if e != nil || vacuous {
		t.Fatalf("active vacuous=%v err=%v", vacuous, e)
	}

	function, _ := xor.Or(x2)
	changed, e := KitTruthChangePhase(function, 2)
	changedAgain, x := KitTruthChangePhase(changed, 2)
	if e != nil || x != nil || !changedAgain.Equal(function) {
		t.Fatalf("phase involution err=%v/%v", e, x)
	}
	overlap, overlapVariable, e := KitTruthMinimumCofactorSupportOverlap(function)
	if e != nil || overlap < 0 || overlapVariable < 0 {
		t.Fatalf("overlap=%d variable=%d err=%v", overlap, overlapVariable, e)
	}
	variable, zero, one, e := KitTruthBestCofactorVariable(function)
	if e != nil || variable < 0 || zero == nil || one == nil {
		t.Fatalf("best=%d err=%v", variable, e)
	}
	profile, e := KitTruthProfile(function, KitTruthOptions{MaxVariables: 4, MaxWords: 1})
	if e != nil || profile.Total != function.CountOnes()*2 || len(profile.ZeroCofactorCount) != 4 || len(profile.BooleanDifference) != 4 {
		t.Fatalf("profile=%+v err=%v", profile, e)
	}
	fastCounts, e := KitTruthCofactorCounts(function, false)
	slowCounts, x := KitTruthCofactorCounts(function, true)
	if e != nil || x != nil || !slices.Equal(fastCounts.Zero, slowCounts.Zero) || !slices.Equal(fastCounts.One, slowCounts.One) {
		t.Fatalf("cofactor counts fast=%+v slow=%+v err=%v/%v", fastCounts, slowCounts, e, x)
	}
	detailed, e := KitTruthPrintProfileData(function, KitTruthOptions{MaxVariables: 6})
	if e != nil || len(detailed.UniquePair) != 6 || detailed.UniquePair[0][0] != -1 {
		t.Fatalf("detailed profile dimensions err=%v", e)
	}
	for first := 0; first < 6; first++ {
		for second := 0; second < 6; second++ {
			if detailed.UniquePair[first][second] != detailed.UniquePair[second][first] || detailed.CommonPair[first][second] != detailed.CommonPair[second][first] || detailed.TotalPair[first][second] != detailed.TotalPair[second][first] {
				t.Fatalf("asymmetric detailed profile %d/%d", first, second)
			}
		}
	}
	xorDetailed, e := KitTruthPrintProfileData(xor, KitTruthOptions{})
	if e != nil || xorDetailed.UniquePair[0][1] != 128 || xorDetailed.CommonPair[0][1] != 0 || xorDetailed.TotalPair[0][1] != 1024 {
		t.Fatalf("xor detailed unique/common/total=%d/%d/%d err=%v", xorDetailed.UniquePair[0][1], xorDetailed.CommonPair[0][1], xorDetailed.TotalPair[0][1], e)
	}
	hex, e := KitTruthHex(kitKernelTable(t, 4, 0x96a5))
	if e != nil || hex != "96a5" {
		t.Fatalf("hex=%q err=%v", hex, e)
	}
	script, e := KitTruthABCScript(kitKernelTable(t, 4, 0x96a5))
	if e != nil || script != "rt 96a5; bdd; sop; ps\n" {
		t.Fatalf("script=%q err=%v", script, e)
	}
	firstHash, _ := KitTruthHash(function)
	secondHash, _ := KitTruthHash(function)
	if firstHash != secondHash {
		t.Fatal("hash is not deterministic")
	}
	semi, e := KitTruthSemiCanonicalize(function)
	if e != nil || semi.Table == nil {
		t.Fatalf("semi err=%v", e)
	}
}

func TestKitTruthReverseVariableOrderSixThroughTen(t *testing.T) {
	for variables := 6; variables <= 10; variables++ {
		count, _ := TruthTableWordCount(variables)
		words := make([]uint64, count)
		for i := range words {
			words[i] = uint64(0x9e3779b97f4a7c15) * uint64(i+1+variables)
		}
		table, e := TruthTableFromWords(variables, words, 0)
		if e != nil {
			t.Fatal(e)
		}
		reversed, e := KitTruthReverseVariableOrder(table, KitTruthOptions{})
		if e != nil {
			t.Fatal(e)
		}
		order := make([]int, variables)
		for i := range order {
			order[i] = variables - 1 - i
		}
		want, _ := table.PermuteVariables(order)
		if !reversed.Equal(want) {
			t.Fatalf("reverse mismatch n=%d", variables)
		}
		twice, _ := KitTruthReverseVariableOrder(reversed, KitTruthOptions{})
		if !twice.Equal(table) {
			t.Fatalf("reverse involution n=%d", variables)
		}
	}
}

func TestKitTruthLimitsCancellationAndErrors(t *testing.T) {
	table := kitKernelTable(t, 4, 0x1234)
	closed := make(chan struct{})
	close(closed)
	if r, e := KitTruthPermute(table, []int{0, 1, 2, 3}, KitTruthOptions{Stop: closed}); r != nil || !errors.Is(e, ErrCancelled) {
		t.Fatalf("cancel result=%v err=%v", r, e)
	}
	if _, e := KitTruthProfile(table, KitTruthOptions{MaxVariables: 3}); !errors.Is(e, ErrLimit) {
		t.Fatalf("variable limit=%v", e)
	}
	if _, e := KitTruthPrintProfileData(table, KitTruthOptions{Stop: closed}); !errors.Is(e, ErrCancelled) {
		t.Fatalf("detailed cancel=%v", e)
	}
	if _, e := KitTruthQuantifySet(table, 1<<4, false, KitTruthOptions{}); !errors.Is(e, ErrDimension) {
		t.Fatalf("mask dimension=%v", e)
	}
	if _, e := KitTruthSwapAdjacent(table, 3); !errors.Is(e, ErrDimension) {
		t.Fatalf("adjacent dimension=%v", e)
	}
	if _, e := KitTruthReverseVariableOrder(table, KitTruthOptions{}); !errors.Is(e, ErrDimension) {
		t.Fatalf("reverse dimension=%v", e)
	}
	if _, e := KitTruthSupportMask(nil); !errors.Is(e, ErrDimension) {
		t.Fatalf("nil dimension=%v", e)
	}
}

func runKitTruthKernelModel(data []byte) error {
	if len(data) < 3 {
		return nil
	}
	variables := 1 + int(data[0]%10)
	count, _ := TruthTableWordCount(variables)
	words := make([]uint64, count)
	for i := range words {
		for j := 0; j < 8; j++ {
			words[i] |= uint64(data[(1+i*8+j)%len(data)]) << uint(8*j)
		}
	}
	if variables < 6 {
		words[0], _ = NormalizeSmallTruthWord(words[0], variables)
	}
	table, e := TruthTableFromWords(variables, words, 0)
	if e != nil {
		return e
	}
	first := int(data[1] % byte(variables))
	second := int(data[2] % byte(variables))
	swapped, e := table.SwapVariables(first, second)
	if e != nil {
		return e
	}
	twice, e := swapped.SwapVariables(first, second)
	if e != nil || !twice.Equal(table) {
		return ErrInvariant
	}
	flipped, e := table.FlipVariable(first)
	if e != nil {
		return e
	}
	flipped, e = flipped.FlipVariable(first)
	if e != nil || !flipped.Equal(table) {
		return ErrInvariant
	}
	unique, e := KitTruthUnique(table, first)
	if e != nil {
		return e
	}
	if used, _ := unique.HasVariable(first); used {
		return ErrInvariant
	}
	fast, e := table.CofactorPopulations()
	if e != nil {
		return e
	}
	slow, e := table.CofactorPopulationsSlow()
	if e != nil || !slices.Equal(fast.Zero, slow.Zero) || !slices.Equal(fast.One, slow.One) {
		return ErrInvariant
	}
	return nil
}

func TestKitTruthIndependentModelCampaign(t *testing.T) {
	state := uint64(0x6a09e667f3bcc909)
	for execution := 0; execution < 5000; execution++ {
		data := make([]byte, 33)
		for i := range data {
			state ^= state << 13
			state ^= state >> 7
			state ^= state << 17
			data[i] = byte(state)
		}
		if e := runKitTruthKernelModel(data); e != nil {
			t.Fatalf("execution %d: %v", execution, e)
		}
	}
}

func FuzzKitTruthKernel(f *testing.F) {
	f.Add([]byte{4, 0, 1, 0x96, 0xa5})
	f.Add([]byte{10, 9, 0, 1, 2, 3, 4, 5, 6, 7})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 512 {
			data = data[:512]
		}
		if e := runKitTruthKernelModel(data); e != nil {
			t.Fatal(e)
		}
	})
}
