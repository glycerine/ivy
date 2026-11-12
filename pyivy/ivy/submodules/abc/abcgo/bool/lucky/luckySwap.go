package lucky

import "math/bits"

var mask0 = [6]uint64{0x5555555555555555, 0x3333333333333333, 0x0f0f0f0f0f0f0f0f, 0x00ff00ff00ff00ff, 0x0000ffff0000ffff, 0x00000000ffffffff}
var adjacentMasks = [5][3]uint64{{0x9999999999999999, 0x2222222222222222, 0x4444444444444444}, {0xc3c3c3c3c3c3c3c3, 0x0c0c0c0c0c0c0c0c, 0x3030303030303030}, {0xf00ff00ff00ff00f, 0x00f000f000f000f0, 0x0f000f000f000f00}, {0xff0000ffff0000ff, 0x0000ff000000ff00, 0x00ff000000ff0000}, {0xffff00000000ffff, 0x00000000ffff0000, 0x0000ffff00000000}}

func Kit_TruthWordNum_64bit(vars int32) int32 {
	if vars < 0 || vars > 16 {
		panic("Lucky supports zero to sixteen variables")
	}
	if vars <= 6 {
		return 1
	}
	return 1 << uint(vars-6)
}
func Kit_WordCountOnes_64bit(x uint64) int32 { return int32(bits.OnesCount64(x)) }
func Kit_TruthCountOnes_64bit(table []uint64, vars int32) int32 {
	var count int32
	for _, word := range table[:Kit_TruthWordNum_64bit(vars)] {
		count += int32(bits.OnesCount64(word))
	}
	return count
}
func Kit_TruthCountOnesInCofs_64bit(table []uint64, vars int32, store []int32) {
	clear(store[:vars])
	for index, word := range table[:Kit_TruthWordNum_64bit(vars)] {
		for variable := int32(0); variable < vars; variable++ {
			if variable < 6 {
				store[variable] += int32(bits.OnesCount64(word & mask0[variable]))
			} else if index&(1<<uint(variable-6)) == 0 {
				store[variable] += int32(bits.OnesCount64(word))
			}
		}
	}
}
func Kit_TruthChangePhase_64bit(table []uint64, vars, variable int32) {
	if variable < 0 || variable >= vars {
		panic("phase variable outside truth support")
	}
	words := int(Kit_TruthWordNum_64bit(vars))
	if variable < 6 {
		shift := uint(1) << uint(variable)
		mask := mask0[variable]
		for i := 0; i < words; i++ {
			x := table[i]
			table[i] = (x&mask)<<shift | (x&^mask)>>shift
		}
		return
	}
	step := 1 << uint(variable-6)
	for block := 0; block < words; block += 2 * step {
		for i := 0; i < step; i++ {
			table[block+i], table[block+step+i] = table[block+step+i], table[block+i]
		}
	}
}
func Kit_TruthNot_64bit(table []uint64, vars int32) {
	for i := range table[:Kit_TruthWordNum_64bit(vars)] {
		table[i] = ^table[i]
	}
}
func Kit_TruthCopy_64bit(dst, src []uint64, vars int32) {
	copy(dst[:Kit_TruthWordNum_64bit(vars)], src[:Kit_TruthWordNum_64bit(vars)])
}
func Kit_TruthSwapAdjacentVars_64bit(table []uint64, vars, variable int32) {
	if variable < 0 || variable >= vars-1 {
		panic("adjacent swap outside truth support")
	}
	words := int(Kit_TruthWordNum_64bit(vars))
	switch {
	case variable < 5:
		for i := 0; i < words; i++ {
			table[i] = Extra_Truth6SwapAdjacent(table[i], variable)
		}
	case variable == 5:
		for i := 0; i < words; i += 2 {
			low, high := table[i], table[i+1]
			table[i] = low&0xffffffff | high<<32
			table[i+1] = low>>32 | high&0xffffffff00000000
		}
	default:
		step := 1 << uint(variable-6)
		for block := 0; block < words; block += 4 * step {
			for i := 0; i < step; i++ {
				table[block+step+i], table[block+2*step+i] = table[block+2*step+i], table[block+step+i]
			}
		}
	}
}
func swapPhaseBits(phase uint32, variable int32) uint32 {
	if ((phase>>uint(variable))^(phase>>uint(variable+1)))&1 != 0 {
		phase ^= 3 << uint(variable)
	}
	return phase
}
func semiCanonical(table []uint64, vars int32, permutation []byte, store []int32, extra bool) uint32 {
	words := Kit_TruthWordNum_64bit(vars)
	ones := Kit_TruthCountOnes_64bit(table, vars)
	var phase uint32
	if extra && ones == words*32 {
		phase |= 1 << uint(vars+2)
	} else if ones > words*32 {
		phase |= 1 << uint(vars)
		Kit_TruthNot_64bit(table, vars)
		ones = words*64 - ones
	}
	Kit_TruthCountOnesInCofs_64bit(table, vars, store)
	for i := int32(0); i < vars; i++ {
		if extra && 2*store[i] == ones {
			phase |= 1 << uint(vars+1)
			continue
		}
		if store[i] >= ones-store[i] {
			continue
		}
		phase |= 1 << uint(i)
		store[i] = ones - store[i]
		Kit_TruthChangePhase_64bit(table, vars, i)
	}
	for changed := true; changed; {
		changed = false
		for i := int32(0); i < vars-1; i++ {
			if store[i] <= store[i+1] {
				continue
			}
			changed = true
			store[i], store[i+1] = store[i+1], store[i]
			permutation[i], permutation[i+1] = permutation[i+1], permutation[i]
			phase = swapPhaseBits(phase, i)
			Kit_TruthSwapAdjacentVars_64bit(table, vars, i)
		}
	}
	return phase
}
func Kit_TruthSemiCanonicize_Yasha(table []uint64, vars int32, permutation []byte) uint32 {
	return semiCanonical(table, vars, permutation, make([]int32, vars), false)
}
func Kit_TruthSemiCanonicize_Yasha1(table []uint64, vars int32, permutation []byte, store []int32) uint32 {
	return semiCanonical(table, vars, permutation, store, true)
}
func Kit_TruthSemiCanonicize_Yasha_simple(table []uint64, vars int32, store []int32) {
	if store == nil {
		store = make([]int32, vars)
	}
	permutation := make([]byte, vars)
	resetPCanonPermArray(permutation, vars)
	semiCanonical(table, vars, permutation, store, false)
}
