package kit

// Covers use one bit per literal (two bits per variable). Scratch memory is a
// Go slice arena: callers may retain sub-slices because Go keeps old backing
// arrays alive when the arena grows.
func kitSopReset(s *Kit_Sop_t, capacity int, memory *[]uint32) {
	s.failed = false
	s.nLits = 0
	if memory == nil {
		s.pCubes = make([]uint32, 0, capacity)
		return
	}
	start := len(*memory)
	*memory = append(*memory, make([]uint32, capacity)...)
	s.pCubes = (*memory)[start : start : start+capacity]
}
func Kit_SopCreate(result *Kit_Sop_t, input []uint32, nVars int32, memory *[]uint32) {
	kitSopReset(result, len(input), memory)
	for _, c := range input {
		Kit_SopPushCube(result, c)
	}
}
func Kit_SopCreateInverse(result *Kit_Sop_t, input []uint32, nLits int32, memory *[]uint32) {
	kitSopReset(result, len(input), memory)
	for _, c := range input {
		m := (c | c>>1) & 0x55555555
		m |= m << 1
		Kit_SopPushCube(result, c^m)
	}
}
func Kit_SopDup(result, sop *Kit_Sop_t, memory *[]uint32) {
	input := sop.pCubes[:sop.cubeNum()]
	Kit_SopCreate(result, input, 0, memory)
}
func Kit_SopDivideByLiteralQuo(sop *Kit_Sop_t, lit int32) {
	k := 0
	mask := uint32(1) << uint(lit)
	for _, c := range sop.pCubes[:sop.cubeNum()] {
		if c&mask != 0 {
			sop.pCubes[k] = c &^ mask
			k++
		}
	}
	sop.pCubes = sop.pCubes[:k]
}
func Kit_SopDivideByCube(sop, div, quo, rem *Kit_Sop_t, memory *[]uint32) {
	if div.cubeNum() != 1 {
		panic("kit: divisor must have one cube")
	}
	d := div.pCubes[0]
	kitSopReset(quo, int(sop.cubeNum()), memory)
	kitSopReset(rem, int(sop.cubeNum()), memory)
	for _, c := range sop.pCubes[:sop.cubeNum()] {
		if c&d == d {
			Kit_SopPushCube(quo, c&^d)
		} else {
			Kit_SopPushCube(rem, c)
		}
	}
}
func Kit_SopDivideInternal(sop, div, quo, rem *Kit_Sop_t, memory *[]uint32) {
	if div.cubeNum() == 1 {
		Kit_SopDivideByCube(sop, div, quo, rem, memory)
		return
	}
	if div.cubeNum() <= 0 || sop.cubeNum() < div.cubeNum() {
		panic("kit: invalid algebraic divisor")
	}
	kitSopReset(quo, int(sop.cubeNum()/div.cubeNum()), memory)
	used := make([]bool, sop.cubeNum())
	for i, c := range sop.pCubes[:sop.cubeNum()] {
		if used[i] {
			continue
		}
		divisorIndex := -1
		for k, d := range div.pCubes[:div.cubeNum()] {
			if c&d == d {
				divisorIndex = k
				break
			}
		}
		if divisorIndex < 0 {
			continue
		}
		q := c &^ div.pCubes[divisorIndex]
		indices := make([]int, div.cubeNum())
		indices[divisorIndex] = i
		found := true
		for k, d := range div.pCubes[:div.cubeNum()] {
			if k == divisorIndex {
				continue
			}
			index := -1
			for j, candidate := range sop.pCubes[:sop.cubeNum()] {
				if !used[j] && candidate&d == d && candidate&^d == q {
					index = j
					break
				}
			}
			if index < 0 {
				found = false
				break
			}
			indices[k] = index
		}
		if !found {
			continue
		}
		Kit_SopPushCube(quo, q)
		for _, j := range indices {
			used[j] = true
		}
	}
	kitSopReset(rem, int(sop.cubeNum()), memory)
	for i, c := range sop.pCubes[:sop.cubeNum()] {
		if !used[i] {
			Kit_SopPushCube(rem, c)
		}
	}
}
func Kit_SopCommonCube(sop *Kit_Sop_t) uint32 {
	mask := ^uint32(0)
	for _, c := range sop.pCubes[:sop.cubeNum()] {
		mask &= c
	}
	return mask
}
func Kit_SopMakeCubeFree(sop *Kit_Sop_t) {
	mask := Kit_SopCommonCube(sop)
	for i := range sop.pCubes[:sop.cubeNum()] {
		sop.pCubes[i] &^= mask
	}
}
func Kit_SopIsCubeFree(sop *Kit_Sop_t) int32 {
	if Kit_SopCommonCube(sop) == 0 {
		return 1
	}
	return 0
}
func Kit_SopCommonCubeCover(result, sop *Kit_Sop_t, memory *[]uint32) {
	if sop.cubeNum() <= 0 {
		panic("kit: empty common cube cover")
	}
	kitSopReset(result, 1, memory)
	Kit_SopPushCube(result, Kit_SopCommonCube(sop))
}
func kitLiteralCount(sop *Kit_Sop_t, lit int32) int32 {
	var count int32
	for _, c := range sop.pCubes[:sop.cubeNum()] {
		if c&(1<<uint(lit)) != 0 {
			count++
		}
	}
	return count
}
func Kit_SopAnyLiteral(sop *Kit_Sop_t, nLits int32) int32 {
	for lit := int32(0); lit < nLits; lit++ {
		if kitLiteralCount(sop, lit) > 1 {
			return lit
		}
	}
	return -1
}
func Kit_SopWorstLiteral(sop *Kit_Sop_t, nLits int32) int32 {
	index, best := int32(-1), int32(1000000)
	for lit := int32(0); lit < nLits; lit++ {
		n := kitLiteralCount(sop, lit)
		if n >= 2 && n < best {
			index, best = lit, n
		}
	}
	return index
}
func Kit_SopBestLiteral(sop *Kit_Sop_t, nLits int32, mask uint32) int32 {
	index, best := int32(-1), int32(-1)
	for lit := int32(0); lit < nLits; lit++ {
		if mask&(1<<uint(lit)) == 0 {
			continue
		}
		n := kitLiteralCount(sop, lit)
		if n >= 2 && n > best {
			index, best = lit, n
		}
	}
	return index
}
func Kit_SopDivisorZeroKernel_rec(sop *Kit_Sop_t, nLits int32) {
	for {
		lit := Kit_SopWorstLiteral(sop, nLits)
		if lit < 0 {
			return
		}
		Kit_SopDivideByLiteralQuo(sop, lit)
		Kit_SopMakeCubeFree(sop)
	}
}
func Kit_SopDivisor(result, sop *Kit_Sop_t, nLits int32, memory *[]uint32) int32 {
	if sop.cubeNum() <= 1 || Kit_SopAnyLiteral(sop, nLits) < 0 {
		return 0
	}
	Kit_SopDup(result, sop, memory)
	Kit_SopDivisorZeroKernel_rec(result, nLits)
	return 1
}
func Kit_SopBestLiteralCover(result, sop *Kit_Sop_t, cube uint32, nLits int32, memory *[]uint32) {
	lit := Kit_SopBestLiteral(sop, nLits, cube)
	if lit < 0 {
		panic("kit: no repeated literal in divisor")
	}
	kitSopReset(result, 1, memory)
	Kit_SopPushCube(result, 1<<uint(lit))
}
