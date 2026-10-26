package abc

// Cofactor ordering from utilTruth.h:Abc_TtCompare1VarCofs/Rev, and phase-only
// normalization from opt/dau/dauCanon.c:Abc_TtCanonicizePhase and Var5/Var6.

// CompareCofactors compares zero/one variable cofactors, without allocating
// them. Words scan low-first or high-first; bits within a word compare unsigned.
func (t *TruthTable) CompareCofactors(variable int, reverseWords bool) (int, error) {
	if e := t.Check(); e != nil {
		return 0, e
	}
	if variable < 0 || variable >= t.variables {
		return 0, ErrDimension
	}
	compare := func(a, b uint64) int {
		if a < b {
			return -1
		}
		if a > b {
			return 1
		}
		return 0
	}
	if variable < 6 {
		shift := uint(1) << uint(variable)
		mask := ^truthVariablePattern(variable)
		for i := 0; i < len(t.words); i++ {
			index := i
			if reverseWords {
				index = len(t.words) - 1 - i
			}
			word := t.words[index]
			if cmp := compare(word&mask, (word>>shift)&mask); cmp != 0 {
				return cmp, nil
			}
		}
		return 0, nil
	}
	block := 1 << uint(variable-6)
	pairs := len(t.words) / (2 * block)
	for pair := 0; pair < pairs; pair++ {
		base := pair * 2 * block
		if reverseWords {
			base = (pairs - 1 - pair) * 2 * block
		}
		for i := 0; i < block; i++ {
			offset := i
			if reverseWords {
				offset = block - 1 - i
			}
			if cmp := compare(t.words[base+offset], t.words[base+block+offset]); cmp != 0 {
				return cmp, nil
			}
		}
	}
	return 0, nil
}

// TruthPhaseNormalization records a supplied NPN configuration with identity
// permutation. TransformNPN(nil, InputPhases, OutputPhase) reconstructs Table
// from the source. Each field/storage is independently owned.
type TruthPhaseNormalization struct {
	Table       *TruthTable
	InputPhases []bool
	OutputPhase bool
}

// NormalizePhases matches C's phase-only pass: complement output if its highest
// packed bit is one, then visit inputs from high to low, flipping only when
// the full high-word-first packed value decreases. This is ONE greedy pass,
// not an exhaustive NPN canonical form or an idempotence guarantee. It leaves
// source storage unchanged and uses explicit phase vectors beyond C's uint mask.
func (t *TruthTable) NormalizePhases() (TruthPhaseNormalization, error) {
	if e := t.Check(); e != nil {
		return TruthPhaseNormalization{}, e
	}
	result := t.Clone()
	output := result.words[len(result.words)-1]>>63 != 0
	if output {
		for i, w := range result.words {
			result.words[i] = ^w
		}
	}
	phases := make([]bool, t.variables)
	for variable := t.variables - 1; variable >= 0; variable-- {
		cmp, e := result.CompareCofactors(variable, true)
		if e != nil {
			return TruthPhaseNormalization{}, e
		}
		if cmp < 0 {
			result.flipVariable(variable)
			phases[variable] = true
		}
	}
	return TruthPhaseNormalization{result, phases, output}, nil
}
