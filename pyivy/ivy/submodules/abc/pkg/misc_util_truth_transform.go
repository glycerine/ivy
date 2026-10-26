package abc

// Truth variable transformations from src/misc/util/utilTruth.h:
// Abc_TtFlip/SwapVars/SwapAdjacent, Tt6Permute_rec, ImplementNpnConfig and Exist.
// SwapVariables also ports all storage cases in src/bool/lucky/luckySwapIJ.c.

// FlipVariable returns f with one input complemented, retaining variable count.
func (t *TruthTable) FlipVariable(variable int) (*TruthTable, error) {
	if e := t.Check(); e != nil {
		return nil, e
	}
	if variable < 0 || variable >= t.variables {
		return nil, ErrDimension
	}
	r := t.Clone()
	r.flipVariable(variable)
	return r, nil
}
func (t *TruthTable) flipVariable(variable int) {
	if variable < 6 {
		shift := uint(1) << uint(variable)
		mask := truthVariablePattern(variable)
		for i, w := range t.words {
			t.words[i] = ((w << shift) & mask) | ((w & mask) >> shift)
		}
	} else {
		block := 1 << uint(variable-6)
		for base := 0; base < len(t.words); base += 2 * block {
			for i := 0; i < block; i++ {
				t.words[base+i], t.words[base+block+i] = t.words[base+block+i], t.words[base+i]
			}
		}
	}
}
func (t *TruthTable) SwapVariables(first, second int) (*TruthTable, error) {
	if e := t.Check(); e != nil {
		return nil, e
	}
	if first < 0 || second < 0 || first >= t.variables || second >= t.variables {
		return nil, ErrDimension
	}
	r := t.Clone()
	r.swapVariables(first, second)
	return r, nil
}
func (t *TruthTable) swapVariables(first, second int) {
	if first == second {
		return
	}
	if second < first {
		first, second = second, first
	}
	if second < 6 {
		low, high := truthVariablePattern(first), truthVariablePattern(second)
		up, down := low&^high, high&^low
		unchanged := ^(up | down)
		shift := (uint(1) << uint(second)) - (uint(1) << uint(first))
		for i, w := range t.words {
			t.words[i] = (w & unchanged) | ((w & up) << shift) | ((w & down) >> shift)
		}
		return
	}
	highBlock := 1 << uint(second-6)
	if first < 6 {
		mask := truthVariablePattern(first)
		shift := uint(1) << uint(first)
		for base := 0; base < len(t.words); base += 2 * highBlock {
			for i := 0; i < highBlock; i++ {
				low, high := t.words[base+i], t.words[base+highBlock+i]
				t.words[base+i] = (low &^ mask) | ((high << shift) & mask)
				t.words[base+highBlock+i] = (high & mask) | ((low & mask) >> shift)
			}
		}
		return
	}
	lowBlock := 1 << uint(first-6)
	for base := 0; base < len(t.words); base += 2 * highBlock {
		for offset := 0; offset < highBlock; offset += 2 * lowBlock {
			for i := 0; i < lowBlock; i++ {
				a, b := base+lowBlock+offset+i, base+highBlock+offset+i
				t.words[a], t.words[b] = t.words[b], t.words[a]
			}
		}
	}
}

// PermuteVariables maps each source variable i to target variable order[i].
// Thus result(x) = source(y), y[i] = x[order[i]]. Nil is identity. Permutations
// and source storage are never mutated; arbitrary valid widths are supported.
func (t *TruthTable) PermuteVariables(order []int) (*TruthTable, error) {
	return t.TransformNPN(order, nil, false)
}

// TransformNPN applies source-input phases, output complementation and then
// the source-to-target permutation, matching ImplementNpnConfig. Result at x
// equals source(y) XOR outputPhase, with y[i] = x[order[i]] XOR inputPhases[i].
// Nil phase/permutation arrays mean identity; explicit arrays must match width.
// Unlike C, caller permutation arrays survive unchanged.
func (t *TruthTable) TransformNPN(order []int, inputPhases []bool, outputPhase bool) (*TruthTable, error) {
	if e := t.Check(); e != nil {
		return nil, e
	}
	if order != nil {
		if e := checkPermutation(order, t.variables); e != nil {
			return nil, e
		}
	}
	if inputPhases != nil && len(inputPhases) != t.variables {
		return nil, ErrDimension
	}
	r := t.Clone()
	for variable, flip := range inputPhases {
		if flip {
			r.flipVariable(variable)
		}
	}
	if outputPhase {
		for i, w := range r.words {
			r.words[i] = ^w
		}
	}
	if order != nil {
		labels := append([]int(nil), order...)
		for target := 0; target < len(labels); target++ {
			position := target
			for labels[position] != target {
				position++
			}
			if position != target {
				r.swapVariables(target, position)
				labels[target], labels[position] = labels[position], labels[target]
			}
		}
	}
	return r, nil
}

// Existential eliminates a variable by ORing its cofactors while retaining
// width. It avoids C's fixed 64-word temporary arrays, supporting larger tables.
func (t *TruthTable) Existential(variable int) (*TruthTable, error) {
	return t.quantify(variable, false)
}

// Universal ANDs a variable's cofactors, retaining width and variable slots.
func (t *TruthTable) Universal(variable int) (*TruthTable, error) { return t.quantify(variable, true) }
func (t *TruthTable) quantify(variable int, universal bool) (*TruthTable, error) {
	if e := t.Check(); e != nil {
		return nil, e
	}
	if variable < 0 || variable >= t.variables {
		return nil, ErrDimension
	}
	r := t.Clone()
	if variable < 6 {
		shift := uint(1) << uint(variable)
		mask := ^truthVariablePattern(variable)
		for i, w := range t.words {
			a, b := w&mask, (w>>shift)&mask
			combined := a | b
			if universal {
				combined = a & b
			}
			r.words[i] = combined | (combined << shift)
		}
	} else {
		block := 1 << uint(variable-6)
		for base := 0; base < len(t.words); base += 2 * block {
			for i := 0; i < block; i++ {
				a, b := t.words[base+i], t.words[base+block+i]
				combined := a | b
				if universal {
					combined = a & b
				}
				r.words[base+i], r.words[base+block+i] = combined, combined
			}
		}
	}
	return r, nil
}
