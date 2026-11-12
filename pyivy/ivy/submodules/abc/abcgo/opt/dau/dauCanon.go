package dau

// Adjacent-cofactor refinement helpers from dauCanon.c. The hierarchical and
// adaptive engines live in hierarchical.go and adaptive.go; shared minterm
// counters live in truth_counts.go. No C-only global scratch arrays are used.
import "github.com/glycerine/abcgo/opt/dau/canon"

func dauCompare2Cofs(truth []uint64, nWords, iVar, num1, num2 int32, reverse bool) int32 {
	if num1 < 0 || num1 >= num2 || num2 >= 4 || iVar < 0 {
		panic("invalid DAU adjacent cofactor pair")
	}
	cmp := func(a, b uint64) int32 {
		if a < b {
			return -1
		}
		if a > b {
			return 1
		}
		return 0
	}
	if iVar <= 4 {
		mask := ^dauTruthVariables[iVar] &^ dauTruthVariables[iVar+1]
		shift := uint(1) << uint(iVar)
		for k := int32(0); k < nWords; k++ {
			w := k
			if reverse {
				w = nWords - 1 - k
			}
			a := (truth[w] >> uint(num1*int32(shift))) & mask
			b := (truth[w] >> uint(num2*int32(shift))) & mask
			if r := cmp(a, b); r != 0 {
				return r
			}
		}
		return 0
	}
	if iVar == 5 {
		for k := int32(0); k < nWords; k += 2 {
			base := k
			if reverse {
				base = nWords - 2 - k
			}
			a := uint64(uint32(truth[base+num1/2] >> uint(32*(num1&1))))
			b := uint64(uint32(truth[base+num2/2] >> uint(32*(num2&1))))
			if r := cmp(a, b); r != 0 {
				return r
			}
		}
		return 0
	}
	step := int32(1) << uint(iVar-6)
	for k := int32(0); k < nWords; k += 4 * step {
		base := k
		if reverse {
			base = nWords - 4*step - k
		}
		for j := int32(0); j < step; j++ {
			off := j
			if reverse {
				off = step - 1 - j
			}
			if r := cmp(truth[base+num1*step+off], truth[base+num2*step+off]); r != 0 {
				return r
			}
		}
	}
	return 0
}
func Abc_TtCompare2VarCofs(truth []uint64, nWords, iVar, num1, num2 int32) int32 {
	return dauCompare2Cofs(truth, nWords, iVar, num1, num2, false)
}
func Abc_TtCompare2VarCofsRev(truth []uint64, nWords, iVar, num1, num2 int32) int32 {
	return dauCompare2Cofs(truth, nWords, iVar, num1, num2, true)
}
func Abc_TtCheckEqual2VarCofs(truth []uint64, nWords, iVar, num1, num2 int32) int32 {
	return dauBool(Abc_TtCompare2VarCofs(truth, nWords, iVar, num1, num2) == 0)
}
func Abc_TtSemiCanonicize(truth []uint64, nVars int32, perm []byte, store []int32, onlySwap int32) uint32 {
	if nVars < 0 || nVars > 16 || len(perm) < int(nVars) {
		panic("invalid DAU semi-canonical dimensions")
	}
	if store == nil {
		store = make([]int32, nVars+1)
	}
	for i := range perm[:nVars] {
		perm[i] = byte(i)
	}
	words := dauCountWords(nVars)
	table := truth[:words]
	ones := Abc_TtCountOnesInTruth(table, nVars)
	phase := uint32(0)
	if ones > int32(words*32) && onlySwap == 0 {
		for i := range table {
			table[i] = ^table[i]
		}
		ones = int32(words*64) - ones
		phase |= 1 << uint(nVars)
	}
	Abc_TtCountOnesInCofs(table, nVars, store)
	store[nVars] = ones
	for v := int32(0); v < nVars; v++ {
		if store[v] < ones-store[v] && onlySwap == 0 {
			canon.FlipVariable(table, int(v))
			phase |= 1 << uint(v)
			store[v] = ones - store[v]
		}
	}
	for i := int32(0); i < nVars-1; i++ {
		best := i + 1
		for j := i + 2; j < nVars; j++ {
			if store[j] < store[best] {
				best = j
			}
		}
		if store[i] <= store[best] {
			continue
		}
		perm[i], perm[best] = perm[best], perm[i]
		store[i], store[best] = store[best], store[i]
		if phase>>uint(i)&1 != phase>>uint(best)&1 {
			phase ^= 1<<uint(i) | 1<<uint(best)
		}
		canon.SwapVariables(table, int(i), int(best))
	}
	return phase
}
func Abc_TtCofactorTest10(truth []uint64, nVars, n int32) {
	words := dauCountWords(nVars)
	for v := int32(0); v < nVars-1; v++ {
		a, b := append([]uint64(nil), truth[:words]...), append([]uint64(nil), truth[:words]...)
		canon.SwapAdjacent(a, int(v))
		canon.SwapVariables(b, int(v), int(v+1))
		if canon.CompareReverse(a, b) != 0 {
			panic("inconsistent DAU adjacent swap")
		}
	}
}
func Abc_Tt6CofactorPermNaive(truth []uint64, v, swapOnly int32) int32 {
	return int32(canon.AdjacentConfiguration(truth[:1], int(v), swapOnly != 0))
}
func Abc_TtCofactorPermNaive(truth []uint64, v, nWords, swapOnly int32) int32 {
	return int32(canon.AdjacentConfiguration(truth[:nWords], int(v), swapOnly != 0))
}
func Abc_TtCofactorPermConfig(truth []uint64, v, nWords, swapOnly, naive int32) int32 {
	if nWords == 1 {
		return Abc_Tt6CofactorPermNaive(truth, v, swapOnly)
	}
	if naive != 0 {
		return Abc_TtCofactorPermNaive(truth, v, nWords, swapOnly)
	}
	table := truth[:nWords]
	compare := func(a, b int32) int32 { return Abc_TtCompare2VarCofsRev(table, nWords, v, a, b) }
	flip := func(variable int32) { canon.FlipVariable(table, int(variable)) }
	if swapOnly != 0 {
		if compare(1, 2) < 0 {
			canon.SwapAdjacent(table, int(v))
			return 4
		}
		return 0
	}
	config := int32(0)
	c01, c23 := compare(0, 1), compare(2, 3)
	if c23 >= 0 {
		if c01 >= 0 {
			c13 := compare(1, 3)
			if c13 < 0 || c13 == 0 && compare(0, 2) < 0 {
				flip(v + 1)
				config = 2
			}
		} else {
			if compare(0, 3) < 0 {
				flip(v)
				flip(v + 1)
				config = 3
			} else if c23 == 0 {
				flip(v)
				config = 1
			}
		}
	} else {
		if c01 >= 0 {
			c12 := compare(1, 2)
			if c12 > 0 {
				flip(v)
				config = 1
			} else if c12 == 0 {
				flip(v)
				flip(v + 1)
				config = 3
			} else {
				flip(v + 1)
				config = 2
				if c01 == 0 {
					flip(v)
					config ^= 1
				}
			}
		} else {
			c02 := compare(0, 2)
			if c02 < 0 {
				flip(v)
				flip(v + 1)
				config = 3
			} else if c02 == 0 {
				if compare(1, 3) >= 0 {
					flip(v)
					config = 1
				} else {
					flip(v)
					flip(v + 1)
					config = 3
				}
			} else {
				flip(v)
				config = 1
			}
		}
	}
	if compare(1, 2) < 0 {
		canon.SwapAdjacent(table, int(v))
		config ^= 4
	}
	return config
}
func Abc_TtCofactorPerm(truth []uint64, v, nWords, swapOnly int32, perm []byte, phase *uint32, naive int32) int32 {
	if swapOnly != 0 {
		config := Abc_TtCofactorPermConfig(truth, v, nWords, 1, 0)
		if config != 0 {
			if *phase>>uint(v)&1 != *phase>>uint(v+1)&1 {
				*phase ^= 3 << uint(v)
			}
			perm[v], perm[v+1] = perm[v+1], perm[v]
		}
		return config
	}
	old := append([]uint64(nil), truth[:nWords]...)
	config := Abc_TtCofactorPermConfig(truth, v, nWords, 0, naive)
	if config == 0 {
		return 0
	}
	if canon.CompareReverse(truth[:nWords], old) == 1 {
		copy(truth, old)
		return 0
	}
	if config&1 != 0 {
		*phase ^= 1 << uint(v)
	}
	if config&2 != 0 {
		*phase ^= 1 << uint(v+1)
	}
	if config&4 != 0 {
		if *phase>>uint(v)&1 != *phase>>uint(v+1)&1 {
			*phase ^= 3 << uint(v)
		}
		perm[v], perm[v+1] = perm[v+1], perm[v]
	}
	return config
}
