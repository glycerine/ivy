package abc

// Truth-table and permutation operations from src/bool/kit/kitTruth.c and
// src/bool/kit/kitPerm.c.

import (
	"math/bits"
	"strings"
)

// KitTruthOptions bounds operations which scan or allocate whole tables. A
// closed Stop channel cancels before a result is returned.
type KitTruthOptions struct {
	MaxVariables int
	MaxWords     int
	Stop         <-chan struct{}
}

func checkKitTruth(t *TruthTable, o KitTruthOptions) error {
	if e := t.Check(); e != nil {
		return e
	}
	if o.MaxVariables < 0 || o.MaxWords < 0 {
		return ErrDimension
	}
	if o.MaxVariables > 0 && t.variables > o.MaxVariables || o.MaxWords > 0 && len(t.words) > o.MaxWords {
		return ErrLimit
	}
	if o.Stop != nil {
		select {
		case <-o.Stop:
			return ErrCancelled
		default:
		}
	}
	return nil
}

// KitTruthSwapAdjacent swaps adjacent input variables without modifying t.
func KitTruthSwapAdjacent(t *TruthTable, variable int) (*TruthTable, error) {
	if e := t.Check(); e != nil {
		return nil, e
	}
	if variable < 0 || variable+1 >= t.variables {
		return nil, ErrDimension
	}
	return t.SwapVariables(variable, variable+1)
}

// KitTruthStretch embeds source inputs at set positions in phase.
func KitTruthStretch(t *TruthTable, variables int, phase uint64) (*TruthTable, error) {
	return t.ExtraStretchTruth(variables, phase)
}

// KitTruthShrink retains phase-selected inputs. Active discarded inputs are
// rejected instead of being silently truncated.
func KitTruthShrink(t *TruthTable, phase uint64) (*TruthTable, error) {
	return t.ExtraShrinkTruth(phase)
}

// KitTruthPermute maps source input i to target position permutation[i].
func KitTruthPermute(t *TruthTable, permutation []int, o KitTruthOptions) (*TruthTable, error) {
	if e := checkKitTruth(t, o); e != nil {
		return nil, e
	}
	return t.PermuteVariables(permutation)
}

// KitTruthReverseVariableOrder ports Kit_PermCompute and its naive reference.
// The C fast path is restricted to 6 through 16 inputs.
func KitTruthReverseVariableOrder(t *TruthTable, o KitTruthOptions) (*TruthTable, error) {
	if e := checkKitTruth(t, o); e != nil {
		return nil, e
	}
	if t.variables < 6 || t.variables > 16 {
		return nil, ErrDimension
	}
	permutation := make([]int, t.variables)
	for i := range permutation {
		permutation[i] = t.variables - 1 - i
	}
	return t.PermuteVariables(permutation)
}

// KitTruthSupportMask returns the source's 32-bit support representation.
func KitTruthSupportMask(t *TruthTable) (uint32, error) {
	if e := t.Check(); e != nil {
		return 0, e
	}
	if t.variables > 32 {
		return 0, ErrDimension
	}
	var mask uint32
	for variable := 0; variable < t.variables; variable++ {
		used, e := t.HasVariable(variable)
		if e != nil {
			return 0, e
		}
		if used {
			mask |= uint32(1) << uint(variable)
		}
	}
	return mask, nil
}

// KitTruthCofactorCount returns the packed population of one selected half.
func KitTruthCofactorCount(t *TruthTable, variable int, value bool) (int, error) {
	if e := t.Check(); e != nil {
		return 0, e
	}
	if variable < 0 || variable >= t.variables {
		return 0, ErrDimension
	}
	p, e := t.ExtraCofactorCounts()
	if e != nil {
		return 0, e
	}
	if value {
		return p.One[variable], nil
	}
	return p.Zero[variable], nil
}

// KitTruthVariableVacuous reports whether an incompletely specified function
// can be made independent of variable without crossing onset and offset.
func KitTruthVariableVacuous(onset, offset *TruthTable, variable int) (bool, error) {
	if e := onset.checkPair(offset); e != nil {
		return false, e
	}
	if variable < 0 || variable >= onset.variables {
		return false, ErrDimension
	}
	on0, e := onset.Cofactor(variable, false)
	if e != nil {
		return false, e
	}
	on1, _ := onset.Cofactor(variable, true)
	off0, _ := offset.Cofactor(variable, false)
	off1, _ := offset.Cofactor(variable, true)
	for i := range on0.words {
		if on0.words[i]&off1.words[i] != 0 || off0.words[i]&on1.words[i] != 0 {
			return false, nil
		}
	}
	return true, nil
}

// KitTruthQuantifySet quantifies every input selected by mask. Universal false
// means existential OR; true means universal AND.
func KitTruthQuantifySet(t *TruthTable, mask uint64, universal bool, o KitTruthOptions) (*TruthTable, error) {
	if e := checkKitTruth(t, o); e != nil {
		return nil, e
	}
	if t.variables > 64 || t.variables < 64 && mask>>uint(t.variables) != 0 {
		return nil, ErrDimension
	}
	r := t.Clone()
	for variable := 0; variable < t.variables; variable++ {
		if mask&(uint64(1)<<uint(variable)) == 0 {
			continue
		}
		if o.Stop != nil {
			select {
			case <-o.Stop:
				return nil, ErrCancelled
			default:
			}
		}
		var e error
		if universal {
			r, e = r.Universal(variable)
		} else {
			r, e = r.Existential(variable)
		}
		if e != nil {
			return nil, e
		}
	}
	return r, nil
}

// KitTruthUnique is the Boolean difference replicated across both cofactors.
func KitTruthUnique(t *TruthTable, variable int) (*TruthTable, error) {
	flipped, e := t.FlipVariable(variable)
	if e != nil {
		return nil, e
	}
	return t.Xor(flipped, false)
}

// KitTruthBooleanDifferenceCount counts differing cofactor assignments once.
func KitTruthBooleanDifferenceCount(t *TruthTable, variable int) (int, error) {
	u, e := KitTruthUnique(t, variable)
	if e != nil {
		return 0, e
	}
	packed, e := u.extraPacked32Words()
	if e != nil {
		return 0, e
	}
	count := 0
	for _, word := range packed {
		count += bits.OnesCount32(word)
	}
	return count / 2, nil
}

// KitTruthXorCount counts logical assignments where tables differ.
func KitTruthXorCount(first, second *TruthTable) (int, error) {
	x, e := first.Xor(second, false)
	if e != nil {
		return 0, e
	}
	packed, e := x.extraPacked32Words()
	if e != nil {
		return 0, e
	}
	count := 0
	for _, word := range packed {
		count += bits.OnesCount32(word)
	}
	return count, nil
}

// KitTruthMuxVariable recombines two same-width cofactors. complementZero
// complements only the variable-zero half before selection.
func KitTruthMuxVariable(zero, one *TruthTable, variable int, complementZero bool) (*TruthTable, error) {
	if e := zero.checkPair(one); e != nil {
		return nil, e
	}
	if variable < 0 || variable >= zero.variables {
		return nil, ErrDimension
	}
	if complementZero {
		zero = zero.Not()
	}
	return ExtraRecombineTruthCofactors(zero, one, variable)
}

// KitTruthVariablesSymmetric compares the 01 and 10 cofactors.
func KitTruthVariablesSymmetric(t *TruthTable, first, second int) (bool, error) {
	if first == second {
		return false, ErrDimension
	}
	s, e := t.InputSymmetry(first, second, false)
	return s&TruthSymmetryDirect != 0, e
}

// KitTruthVariablesAntiSymmetric compares the 00 and 11 cofactors.
func KitTruthVariablesAntiSymmetric(t *TruthTable, first, second int) (bool, error) {
	if e := t.Check(); e != nil {
		return false, e
	}
	if first < 0 || second < 0 || first >= t.variables || second >= t.variables || first == second {
		return false, ErrDimension
	}
	a, _ := t.Cofactor(first, false)
	a, _ = a.Cofactor(second, false)
	b, _ := t.Cofactor(first, true)
	b, _ = b.Cofactor(second, true)
	return a.Equal(b), nil
}

// KitTruthChangePhase complements one input variable.
func KitTruthChangePhase(t *TruthTable, variable int) (*TruthTable, error) {
	return t.FlipVariable(variable)
}

// KitTruthMinimumCofactorSupportOverlap returns the source score and first
// minimizing input, including its (32,-1) no-candidate sentinel.
func KitTruthMinimumCofactorSupportOverlap(t *TruthTable) (int, int, error) {
	return t.ExtraMinimumCofactorSupportOverlap()
}

// KitTruthBestCofactorVariable minimizes the sum of the two support sizes;
// equal scores retain the lower input index.
func KitTruthBestCofactorVariable(t *TruthTable) (variable int, zero, one *TruthTable, err error) {
	if e := t.Check(); e != nil {
		return -1, nil, nil, e
	}
	if t.IsConstant(false) || t.IsConstant(true) {
		return -1, nil, nil, nil
	}
	bestScore := int(^uint(0) >> 1)
	best := -1
	for candidate := 0; candidate < t.variables; candidate++ {
		z, _ := t.Cofactor(candidate, false)
		o, _ := t.Cofactor(candidate, true)
		zs, _ := z.Support()
		os, _ := o.Support()
		if len(zs)+len(os) < bestScore {
			bestScore, best, zero, one = len(zs)+len(os), candidate, z, o
		}
	}
	return best, zero, one, nil
}

type KitTruthMintermProfile struct {
	Total             int
	ZeroCofactorCount []int
	BooleanDifference []int
}

// KitTruthDetailedProfile is the data formerly emitted by
// Kit_TruthPrintProfile. Inputs below six are stretched to six, exactly as the
// source wrapper does; matrices therefore have at least six rows.
type KitTruthDetailedProfile struct {
	KitTruthMintermProfile
	UniquePair [][]int
	CommonPair [][]int
	TotalPair  [][]int
}

// KitTruthCofactorCounts returns the source's packed 32-bit cofactor counts.
// slow selects the independently implemented direct cofactor model.
func KitTruthCofactorCounts(t *TruthTable, slow bool) (ExtraTruthCofactorCounts, error) {
	if !slow {
		return t.ExtraCofactorCounts()
	}
	if e := t.Check(); e != nil {
		return ExtraTruthCofactorCounts{}, e
	}
	zero := make([]int, t.variables)
	one := make([]int, t.variables)
	for variable := 0; variable < t.variables; variable++ {
		for value := 0; value < 2; value++ {
			cofactor, e := t.Cofactor(variable, value != 0)
			if e != nil {
				return ExtraTruthCofactorCounts{}, e
			}
			packed, e := cofactor.extraPacked32Words()
			if e != nil {
				return ExtraTruthCofactorCounts{}, e
			}
			count := 0
			for _, word := range packed {
				count += bits.OnesCount32(word)
			}
			if value == 0 {
				zero[variable] = count / 2
			} else {
				one[variable] = count / 2
			}
		}
	}
	return ExtraTruthCofactorCounts{Zero: zero, One: one}, nil
}

// KitTruthHash is the exact wrapping prime-weighted 32-bit source hash.
func KitTruthHash(t *TruthTable) (uint32, error) { return t.ExtraTruthHash() }

// KitTruthSemiCanonicalize ports the source's signature pass using owned
// permutation and phase vectors instead of mutable caller buffers.
func KitTruthSemiCanonicalize(t *TruthTable) (ExtraTruthSemiCanonicalization, error) {
	return t.ExtraSemiCanonicalize()
}

// KitTruthProfile returns the non-printing portion of the source profile. It
// replaces process output and fixed twenty-input scratch arrays with owned data.
func KitTruthProfile(t *TruthTable, o KitTruthOptions) (KitTruthMintermProfile, error) {
	if e := checkKitTruth(t, o); e != nil {
		return KitTruthMintermProfile{}, e
	}
	p, e := t.ExtraCofactorCounts()
	if e != nil {
		return KitTruthMintermProfile{}, e
	}
	diff := make([]int, t.variables)
	for variable := range diff {
		if o.Stop != nil {
			select {
			case <-o.Stop:
				return KitTruthMintermProfile{}, ErrCancelled
			default:
			}
		}
		diff[variable], e = KitTruthBooleanDifferenceCount(t, variable)
		if e != nil {
			return KitTruthMintermProfile{}, e
		}
	}
	packed, e := t.extraPacked32Words()
	if e != nil {
		return KitTruthMintermProfile{}, e
	}
	total := 0
	for _, word := range packed {
		total += bits.OnesCount32(word)
	}
	return KitTruthMintermProfile{total, p.Zero, diff}, nil
}

// KitTruthPrintProfileData returns the complete source profile without stdout
// or fixed scratch arrays.
func KitTruthPrintProfileData(t *TruthTable, o KitTruthOptions) (KitTruthDetailedProfile, error) {
	if e := checkKitTruth(t, o); e != nil {
		return KitTruthDetailedProfile{}, e
	}
	work := t
	if t.variables > 20 {
		return KitTruthDetailedProfile{}, ErrLimit
	}
	if t.variables < 6 {
		var e error
		work, e = t.Stretch(6)
		if e != nil {
			return KitTruthDetailedProfile{}, e
		}
	}
	profile, e := KitTruthProfile(work, o)
	if e != nil {
		return KitTruthDetailedProfile{}, e
	}
	n := work.variables
	unique := make([][]int, n)
	common := make([][]int, n)
	total := make([][]int, n)
	for i := 0; i < n; i++ {
		unique[i], common[i], total[i] = make([]int, n), make([]int, n), make([]int, n)
		for j := 0; j < n; j++ {
			unique[i][j], common[i][j], total[i][j] = -1, -1, -1
		}
	}
	remaining := 1 << uint(n-2)
	for first := 0; first < n; first++ {
		for second := first + 1; second < n; second++ {
			if o.Stop != nil {
				select {
				case <-o.Stop:
					return KitTruthDetailedProfile{}, ErrCancelled
				default:
				}
			}
			cofactors := make([]*TruthTable, 4)
			counts := make([]int, 4)
			for assignment := 0; assignment < 4; assignment++ {
				cofactor, _ := work.Cofactor(first, assignment&1 != 0)
				cofactor, _ = cofactor.Cofactor(second, assignment&2 != 0)
				cofactors[assignment] = cofactor
				counts[assignment] = cofactor.CountOnes() / 4
			}
			zeros := []int{remaining - counts[0], remaining - counts[1], remaining - counts[2], remaining - counts[3]}
			commonValue := counts[0]*zeros[3] + zeros[0]*counts[3] + counts[1]*zeros[2] + zeros[1]*counts[2]
			totalValue, uniqueValue := 0, 0
			for a := 0; a < 4; a++ {
				for b := a + 1; b < 4; b++ {
					totalValue += counts[a]*zeros[b] + zeros[a]*counts[b]
					x, _ := cofactors[a].Xor(cofactors[b], false)
					// The source compares (n-1)-variable cofactors with the
					// second selected input still replicated, so each remaining
					// assignment contributes twice.
					uniqueValue += x.CountOnes() / 2
				}
			}
			unique[first][second], unique[second][first] = uniqueValue, uniqueValue
			common[first][second], common[second][first] = commonValue, commonValue
			total[first][second], total[second][first] = totalValue, totalValue
		}
	}
	return KitTruthDetailedProfile{profile, unique, common, total}, nil
}

// KitTruthHex returns the source's most-significant-nibble-first lowercase
// representation, with exactly 2^variables logical bits.
func KitTruthHex(t *TruthTable) (string, error) {
	if e := t.Check(); e != nil {
		return "", e
	}
	if t.variables >= bits.UintSize-1 {
		return "", ErrLimit
	}
	digits := (1 << uint(t.variables)) / 4
	const hex = "0123456789abcdef"
	var b strings.Builder
	b.Grow(digits)
	for digit := digits - 1; digit >= 0; digit-- {
		word := t.words[digit/16]
		b.WriteByte(hex[(word>>uint(4*(digit%16)))&15])
	}
	return b.String(), nil
}

// KitTruthABCScript returns the bytes formerly written by Kit_TruthDumpToFile.
// It has no fixed path or process-global filename state.
func KitTruthABCScript(t *TruthTable) (string, error) {
	hex, e := KitTruthHex(t)
	if e != nil {
		return "", e
	}
	return "rt " + hex + "; bdd; sop; ps\n", nil
}
