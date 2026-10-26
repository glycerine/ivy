package abc

// Pair cofactor comparisons and optimized selection from opt/dau/dauCanon.c:
// Abc_TtCheckEqual2VarCofs, Compare2VarCofs/Rev, CofactorPermConfig/Perm.

// CompareAdjacentCofactors compares pair cofactors numbered by bit 0 for first
// and bit 1 for first+1. Words scan low/high first according to reverseWords.
// Equal or reversed cofactor numbers are accepted, extending C's a<b assertion.
func (t *TruthTable) CompareAdjacentCofactors(first, a, b int, reverseWords bool) (int, error) {
	if e := t.Check(); e != nil {
		return 0, e
	}
	if first < 0 || first >= t.variables-1 || a < 0 || a > 3 || b < 0 || b > 3 {
		return 0, ErrDimension
	}
	return t.compareAdjacentCofactors(first, a, b, reverseWords), nil
}
func (t *TruthTable) compareAdjacentCofactors(first, a, b int, reverse bool) int {
	compare := func(x, y uint64) int {
		if x < y {
			return -1
		}
		if x > y {
			return 1
		}
		return 0
	}
	if first < 5 {
		shift := uint(1) << uint(first)
		mask := ^(truthVariablePattern(first) | truthVariablePattern(first+1))
		for i := range t.words {
			index := i
			if reverse {
				index = len(t.words) - 1 - i
			}
			w := t.words[index]
			if cmp := compare((w>>(uint(a)*shift))&mask, (w>>(uint(b)*shift))&mask); cmp != 0 {
				return cmp
			}
		}
		return 0
	}
	if first == 5 {
		for i := 0; i < len(t.words)/2; i++ {
			base := i * 2
			if reverse {
				base = len(t.words) - 2 - i*2
			}
			x := (t.words[base+a/2] >> uint((a%2)*32)) & 0xffffffff
			y := (t.words[base+b/2] >> uint((b%2)*32)) & 0xffffffff
			if cmp := compare(x, y); cmp != 0 {
				return cmp
			}
		}
		return 0
	}
	block := 1 << uint(first-6)
	groups := len(t.words) / (4 * block)
	for group := 0; group < groups; group++ {
		base := group * 4 * block
		if reverse {
			base = (groups - 1 - group) * 4 * block
		}
		for i := 0; i < block; i++ {
			offset := i
			if reverse {
				offset = block - 1 - i
			}
			if cmp := compare(t.words[base+a*block+offset], t.words[base+b*block+offset]); cmp != 0 {
				return cmp
			}
		}
	}
	return 0
}
func (t *TruthTable) refineAdjacentSmartCandidate(first int, swapOnly bool) TruthPairRefinement {
	// C always uses the naive path for one word; swap-only is equivalent to
	// comparing the adjacent swap's complete packed value.
	if len(t.words) == 1 || swapOnly {
		return t.refineAdjacent(first, swapOnly)
	}
	r := t.Clone()
	config := uint8(0)
	cmp := func(a, b int) int { return r.compareAdjacentCofactors(first, a, b, true) }
	c01, c23 := cmp(0, 1), cmp(2, 3)
	if c23 >= 0 {
		if c01 >= 0 {
			c13 := cmp(1, 3)
			if c13 < 0 || (c13 == 0 && cmp(0, 2) < 0) {
				r.flipVariable(first + 1)
				config = 2
			}
		} else {
			if cmp(0, 3) < 0 {
				r.flipVariable(first)
				r.flipVariable(first + 1)
				config = 3
			} else if c23 == 0 {
				r.flipVariable(first)
				config = 1
			}
		}
	} else {
		if c01 >= 0 {
			c12 := cmp(1, 2)
			if c12 > 0 {
				r.flipVariable(first)
				config = 1
			} else if c12 == 0 {
				r.flipVariable(first)
				r.flipVariable(first + 1)
				config = 3
			} else {
				r.flipVariable(first + 1)
				config = 2
				if c01 == 0 {
					r.flipVariable(first)
					config ^= 1
				}
			}
		} else {
			c02 := cmp(0, 2)
			if c02 < 0 {
				r.flipVariable(first)
				r.flipVariable(first + 1)
				config = 3
			} else if c02 == 0 {
				if cmp(1, 3) >= 0 {
					r.flipVariable(first)
					config = 1
				} else {
					r.flipVariable(first + 1)
					config = 2
				}
			} else {
				r.flipVariable(first)
				config = 1
			}
		}
	}
	if cmp(1, 2) < 0 {
		r.swapVariables(first, first+1)
		config ^= 4
	}
	return TruthPairRefinement{r, config}
}

// RefineAdjacentCofactorsSmart ports C's optimized candidate path and rollback:
// a candidate that increases the packed value is discarded. Unlike naive Gray
// enumeration, smart ties may yield nonidentity configurations at equal value.
// Configuration retains the source-phase-before-swap contract. Source is owned
// separately and arbitrary valid widths replace C's static scratch buffers.
func (t *TruthTable) RefineAdjacentCofactorsSmart(first int, swapOnly bool) (TruthPairRefinement, error) {
	if e := t.Check(); e != nil {
		return TruthPairRefinement{}, e
	}
	if first < 0 || first >= t.variables-1 {
		return TruthPairRefinement{}, ErrDimension
	}
	result := t.refineAdjacentSmartCandidate(first, swapOnly)
	if truthWordsLessReverse(t.words, result.Table.words) {
		copy(result.Table.words, t.words)
		result.Configuration = 0
	}
	return result, nil
}
