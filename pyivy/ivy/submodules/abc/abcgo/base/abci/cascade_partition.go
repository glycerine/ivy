package abci

// Partition and pattern helpers from src/base/abci/abcCascade.c.
import "math/bits"

func Abc_ResStartPart(inputs int32, parts []uint32, count int32) {
	group, left := inputs/count, inputs%count
	var shift, covered int32
	for i := int32(0); i < count; i++ {
		width := group
		if i < left {
			width++
		}
		parts[i] = (^uint32(0) >> (32 - uint(width))) << uint(shift)
		shift += width
		covered += int32(bits.OnesCount32(parts[i]))
	}
	if covered != inputs {
		panic("abci: partition does not cover inputs")
	}
}
func Abc_ResStartPart2(inputs int32, parts []uint32, count int32) {
	for i := int32(0); i < count; i++ {
		parts[i] = 0
	}
	for i := int32(0); i < inputs; i++ {
		parts[i%count] |= uint32(1) << uint(i)
	}
	var covered int32
	for i := int32(0); i < count; i++ {
		covered += int32(bits.OnesCount32(parts[i]))
	}
	if covered != inputs {
		panic("abci: partition does not cover inputs")
	}
}
func Abc_ResCheckUnique(patterns []int8, count, pattern int32) int32 {
	for i := int32(0); i < count; i++ {
		if int32(patterns[i]) == pattern {
			return 0
		}
	}
	return 1
}
func Abc_ResCheckNonStrict(pattern []int8, variables, bitCount int32) int32 {
	var left, right [256]int8
	limit := int32(1) << uint(bitCount-1)
	var result int32
	for v := int32(0); v < variables; v++ {
		var n0, n1 int32
		var m int32
		for m = 0; m < (int32(1) << uint(variables)); m++ {
			if m&(int32(1)<<uint(v)) == 0 {
				if Abc_ResCheckUnique(left[:], n0, int32(pattern[m])) != 0 {
					left[n0] = pattern[m]
					n0++
					if n0 > limit {
						break
					}
				}
			} else {
				if Abc_ResCheckUnique(right[:], n1, int32(pattern[m])) != 0 {
					right[n1] = pattern[m]
					n1++
					if n1 > limit {
						break
					}
				}
			}
		}
		if m == (int32(1) << uint(variables)) {
			result++
		}
	}
	return result
}
