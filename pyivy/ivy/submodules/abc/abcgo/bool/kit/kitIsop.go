package kit

import "fmt"

const KIT_ISOP_MEM_LIMIT = 1 << 20

func kitIsopAlloc(memory *[]uint32, count int32) []uint32 {
	if memory == nil {
		return make([]uint32, count)
	}
	if count < 0 || len(*memory)+int(count) > KIT_ISOP_MEM_LIMIT {
		return nil
	}
	start := len(*memory)
	*memory = append(*memory, make([]uint32, count)...)
	return (*memory)[start:]
}

var kitVarMasks = [5]uint32{0xaaaaaaaa, 0xcccccccc, 0xf0f0f0f0, 0xff00ff00, 0xffff0000}

func kitTruthHasVar(truth []uint32, nVars, v int32) bool {
	words := Kit_TruthWordNum(nVars)
	if v < 5 {
		shift := uint(1) << uint(v)
		mask := ^kitVarMasks[v]
		for _, x := range truth[:words] {
			if (x^(x>>shift))&mask != 0 {
				return true
			}
		}
		return false
	}
	block := int32(1) << uint(v-5)
	for start := int32(0); start < words; start += 2 * block {
		for i := int32(0); i < block; i++ {
			if truth[start+i] != truth[start+block+i] {
				return true
			}
		}
	}
	return false
}
func kitIsopCofs(truth []uint32, v int32) ([]uint32, []uint32) {
	if v >= 5 {
		n := Kit_TruthWordNum(v)
		return truth[:n], truth[n : 2*n]
	}
	shift := uint(1) << uint(v)
	mask := kitVarMasks[v]
	low := truth[0] &^ mask
	low |= low << shift
	high := truth[0] & mask
	high |= high >> shift
	return []uint32{low}, []uint32{high}
}
func Kit_TruthIsop_rec(on, onDc []uint32, nVars int32, result *Kit_Sop_t, memory *[]uint32) []uint32 {
	*result = Kit_Sop_t{}
	words := Kit_TruthWordNum(nVars)
	out := kitIsopAlloc(memory, words)
	if out == nil {
		result.failed = true
		return nil
	}
	if Kit_TruthIsConst0(on, nVars) != 0 {
		*result = Kit_Sop_t{}
		return out
	}
	if Kit_TruthIsConst1(onDc, nVars) != 0 {
		cube := kitIsopAlloc(memory, 1)
		if cube == nil {
			result.failed = true
			return nil
		}
		*result = Kit_Sop_t{pCubes: cube}
		Kit_TruthFill(out, nVars)
		return out
	}
	v := nVars - 1
	for v >= 0 && !kitTruthHasVar(on, nVars, v) && !kitTruthHasVar(onDc, nVars, v) {
		v--
	}
	if v < 0 {
		panic("kit: inconsistent ISOP bounds")
	}
	on0, on1 := kitIsopCofs(on, v)
	dc0, dc1 := kitIsopCofs(onDc, v)
	n := Kit_TruthWordNum(v)
	temp0, temp1 := make([]uint32, n), make([]uint32, n)
	var c0, c1, c2 Kit_Sop_t
	Kit_TruthSharp(temp0, on0, dc1, v)
	r0 := Kit_TruthIsop_rec(temp0, dc0, v, &c0, memory)
	if c0.cubeNum() < 0 {
		result.failed = true
		return nil
	}
	Kit_TruthSharp(temp1, on1, dc0, v)
	r1 := Kit_TruthIsop_rec(temp1, dc1, v, &c1, memory)
	if c1.cubeNum() < 0 {
		result.failed = true
		return nil
	}
	for i := range temp0 {
		temp0[i] = (on0[i] &^ r0[i]) | (on1[i] &^ r1[i])
		temp1[i] = dc0[i] & dc1[i]
	}
	r2 := Kit_TruthIsop_rec(temp0, temp1, v, &c2, memory)
	if c2.cubeNum() < 0 {
		result.failed = true
		return nil
	}
	count := c0.cubeNum() + c1.cubeNum() + c2.cubeNum()
	result.nLits = c0.nLits + c1.nLits + c2.nLits + c0.cubeNum() + c1.cubeNum()
	result.pCubes = kitIsopAlloc(memory, count)
	if result.pCubes == nil {
		result.failed = true
		return nil
	}
	pos := 0
	for _, c := range c0.pCubes[:c0.cubeNum()] {
		result.pCubes[pos] = c | 1<<uint(2*v)
		pos++
	}
	for _, c := range c1.pCubes[:c1.cubeNum()] {
		result.pCubes[pos] = c | 1<<uint(2*v+1)
		pos++
	}
	for _, c := range c2.pCubes[:c2.cubeNum()] {
		result.pCubes[pos] = c
		pos++
	}
	if v < 5 {
		mask := kitVarMasks[v]
		value := r2[0] | (r0[0] &^ mask) | (r1[0] & mask)
		for i := range out {
			out[i] = value
		}
	} else {
		for start := int32(0); start < words; start += 2 * n {
			for i := int32(0); i < n; i++ {
				out[start+i] = r0[i] | r2[i]
				out[start+n+i] = r1[i] | r2[i]
			}
		}
	}
	return out
}
func Kit_TruthIsop5_rec(on, onDc uint32, nVars int32, result *Kit_Sop_t, memory *[]uint32) uint32 {
	if nVars > 5 {
		panic("kit: single-word ISOP supports at most five variables")
	}
	out := Kit_TruthIsop_rec([]uint32{on}, []uint32{onDc}, nVars, result, memory)
	if out == nil {
		return 0
	}
	return out[0]
}
func Kit_TruthIsop(truth []uint32, nVars int32, memory *[]uint32, tryBoth int32) int32 {
	if nVars < 0 || nVars > 16 {
		panic("kit: ISOP supports 0..16 variables")
	}
	if memory == nil {
		panic("kit: missing result cover")
	}
	var arena []uint32
	var direct Kit_Sop_t
	Kit_TruthIsop_rec(truth, truth, nVars, &direct, &arena)
	if direct.cubeNum() < 0 {
		*memory = nil
		return -1
	}
	chosen := direct
	compl := int32(0)
	if tryBoth != 0 && direct.cubeNum() > 0 && !(direct.cubeNum() == 1 && direct.pCubes[0] == 0) {
		inverse := make([]uint32, Kit_TruthWordNum(nVars))
		Kit_TruthNot(inverse, truth, nVars)
		var other Kit_Sop_t
		Kit_TruthIsop_rec(inverse, inverse, nVars, &other, &arena)
		if other.cubeNum() >= 0 && (other.cubeNum() < direct.cubeNum() || (other.cubeNum() == direct.cubeNum() && other.nLits < direct.nLits)) {
			chosen = other
			compl = 1
		}
	}
	*memory = append((*memory)[:0], chosen.pCubes[:chosen.cubeNum()]...)
	return compl
}
func Kit_TruthIsop2(off, on []uint32, nVars int32, memory *[]uint32, tryBoth, returnTruth int32) int32 {
	if nVars < 0 || nVars > 16 || memory == nil {
		panic("kit: invalid incompletely specified ISOP")
	}
	var arena []uint32
	upper := make([]uint32, Kit_TruthWordNum(nVars))
	Kit_TruthNot(upper, off, nVars)
	var direct Kit_Sop_t
	directTruth := Kit_TruthIsop_rec(on, upper, nVars, &direct, &arena)
	if direct.cubeNum() < 0 {
		*memory = nil
		return -1
	}
	chosen, chosenTruth, compl := direct, directTruth, int32(0)
	if tryBoth != 0 && direct.cubeNum() > 0 && !(direct.cubeNum() == 1 && direct.pCubes[0] == 0) {
		Kit_TruthNot(upper, on, nVars)
		var other Kit_Sop_t
		otherTruth := Kit_TruthIsop_rec(off, upper, nVars, &other, &arena)
		if other.cubeNum() >= 0 && (other.cubeNum() < direct.cubeNum() || (other.cubeNum() == direct.cubeNum() && other.nLits < direct.nLits)) {
			chosen, chosenTruth, compl = other, otherTruth, 1
		}
	}
	if returnTruth != 0 {
		*memory = append((*memory)[:0], chosenTruth...)
	} else {
		*memory = append((*memory)[:0], chosen.pCubes[:chosen.cubeNum()]...)
	}
	return compl
}
func Kit_TruthIsopPrintCover(cover []uint32, nVars, compl int32) {
	for _, cube := range cover {
		for v := int32(0); v < nVars; v++ {
			switch cube >> uint(2*v) & 3 {
			case 0:
				fmt.Print("-")
			case 1:
				fmt.Print("0")
			case 2:
				fmt.Print("1")
			default:
				panic("kit: contradictory cube")
			}
		}
		fmt.Printf(" %d\n", 1-compl)
	}
}
func Kit_TruthIsopPrint(truth []uint32, nVars int32, memory *[]uint32, tryBoth int32) {
	phase := Kit_TruthIsop(truth, nVars, memory, tryBoth)
	if phase < 0 {
		fmt.Println("ISOP memory limit exceeded")
		return
	}
	Kit_TruthIsopPrintCover(*memory, nVars, phase)
}
