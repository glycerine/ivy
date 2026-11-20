package kit

import (
	"fmt"
	"sync"
)

const KIT_ISOP_MEM_LIMIT = 1 << 20

// Callers own the returned cube or truth slice. Only the private recursion
// store is reused, including when the public input and output slices overlap.
var isopStores = sync.Pool{New: func() any { return new([]uint32) }}

func releaseIsopStore(store *[]uint32) {
	if cap(*store) > 64*1024 {
		*store = nil
	} else {
		*store = (*store)[:0]
	}
	isopStores.Put(store)
}

func kitIsopAlloc(memory *[]uint32, count int32) []uint32 {
	if memory == nil {
		return make([]uint32, count)
	}
	if count < 0 || len(*memory)+int(count) > KIT_ISOP_MEM_LIMIT {
		return nil
	}
	start := len(*memory)
	end := start + int(count)
	if end <= cap(*memory) {
		*memory = (*memory)[:end]
	} else {
		*memory = append(*memory, make([]uint32, count)...)
	}
	return (*memory)[start:end:end]
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
	if nVars <= 5 {
		out[0] = Kit_TruthIsop5_rec(on[0], onDc[0], nVars, result, memory)
		if result.failed {
			return nil
		}
		return out
	}
	if Kit_TruthIsConst0(on, nVars) != 0 {
		clear(out)
		*result = Kit_Sop_t{}
		return out
	}
	if Kit_TruthIsConst1(onDc, nVars) != 0 {
		cube := kitIsopAlloc(memory, 1)
		if cube == nil {
			result.failed = true
			return nil
		}
		cube[0] = 0
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
	if v < 5 {
		word := Kit_TruthIsop5_rec(on[0], onDc[0], v+1, result, memory)
		if result.failed {
			return nil
		}
		for i := range out {
			out[i] = word
		}
		return out
	}
	n := Kit_TruthWordNum(v)
	on0, on1 := on[:n], on[n:2*n]
	dc0, dc1 := onDc[:n], onDc[n:2*n]
	temp0, temp1 := out[:n], out[n:2*n]
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
	if !kitIsopJoin(v, &c0, &c1, &c2, result, memory) {
		return nil
	}
	for i := range n {
		temp0[i] = r0[i] | r2[i]
		temp1[i] = r1[i] | r2[i]
	}
	for start := 2 * n; start < words; start += 2 * n {
		copy(out[start:start+2*n], out[:2*n])
	}
	return out
}

// Both recursion paths append the negative, positive and shared cubes in
// the same order as kitIsop.c. Each branch contributes one literal per cube.
func kitIsopJoin(v int32, low, high, shared, result *Kit_Sop_t, memory *[]uint32) bool {
	count := len(low.pCubes) + len(high.pCubes) + len(shared.pCubes)
	result.nLits = low.nLits + high.nLits + shared.nLits + int32(len(low.pCubes)+len(high.pCubes))
	result.pCubes = kitIsopAlloc(memory, int32(count))
	if result.pCubes == nil {
		result.failed = true
		return false
	}
	pos := 0
	for _, c := range low.pCubes {
		result.pCubes[pos] = c | 1<<uint(2*v)
		pos++
	}
	for _, c := range high.pCubes {
		result.pCubes[pos] = c | 1<<uint(2*v+1)
		pos++
	}
	copy(result.pCubes[pos:], shared.pCubes)
	return true
}

func Kit_TruthIsop5_rec(on, onDc uint32, nVars int32, result *Kit_Sop_t, memory *[]uint32) uint32 {
	if nVars < 0 || nVars > 5 {
		panic("kit: single-word ISOP supports zero to five variables")
	}
	*result = Kit_Sop_t{}
	if on == 0 {
		return 0
	}
	if onDc == ^uint32(0) {
		result.pCubes = kitIsopAlloc(memory, 1)
		if result.pCubes == nil {
			result.failed = true
			return 0
		}
		result.pCubes[0] = 0
		return ^uint32(0)
	}
	v := nVars - 1
	for ; v >= 0; v-- {
		shift := uint(1) << uint(v)
		if ((on^(on>>shift))|(onDc^(onDc>>shift)))&^kitVarMasks[v] != 0 {
			break
		}
	}
	if v < 0 {
		panic("kit: inconsistent ISOP bounds")
	}
	mask, shift := kitVarMasks[v], uint(1)<<uint(v)
	on0, on1 := on&^mask, on&mask
	on0 |= on0 << shift
	on1 |= on1 >> shift
	dc0, dc1 := onDc&^mask, onDc&mask
	dc0 |= dc0 << shift
	dc1 |= dc1 >> shift

	var low, high, shared Kit_Sop_t
	r0 := Kit_TruthIsop5_rec(on0&^dc1, dc0, v, &low, memory)
	if low.failed {
		result.failed = true
		return 0
	}
	r1 := Kit_TruthIsop5_rec(on1&^dc0, dc1, v, &high, memory)
	if high.failed {
		result.failed = true
		return 0
	}
	r2 := Kit_TruthIsop5_rec((on0&^r0)|(on1&^r1), dc0&dc1, v, &shared, memory)
	if shared.failed || !kitIsopJoin(v, &low, &high, &shared, result, memory) {
		result.failed = true
		return 0
	}
	return r2 | (r0 &^ mask) | (r1 & mask)
}

func Kit_TruthIsop(truth []uint32, nVars int32, memory *[]uint32, tryBoth int32) int32 {
	if nVars < 0 || nVars > 16 {
		panic("kit: ISOP supports 0..16 variables")
	}
	if memory == nil {
		panic("kit: missing result cover")
	}
	arena := isopStores.Get().(*[]uint32)
	defer releaseIsopStore(arena)
	var direct Kit_Sop_t
	Kit_TruthIsop_rec(truth, truth, nVars, &direct, arena)
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
		Kit_TruthIsop_rec(inverse, inverse, nVars, &other, arena)
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
	arena := isopStores.Get().(*[]uint32)
	defer releaseIsopStore(arena)
	upper := make([]uint32, Kit_TruthWordNum(nVars))
	Kit_TruthNot(upper, off, nVars)
	var direct Kit_Sop_t
	directTruth := Kit_TruthIsop_rec(on, upper, nVars, &direct, arena)
	if direct.cubeNum() < 0 {
		*memory = nil
		return -1
	}
	chosen, chosenTruth, compl := direct, directTruth, int32(0)
	if tryBoth != 0 && direct.cubeNum() > 0 && !(direct.cubeNum() == 1 && direct.pCubes[0] == 0) {
		Kit_TruthNot(upper, on, nVars)
		var other Kit_Sop_t
		otherTruth := Kit_TruthIsop_rec(off, upper, nVars, &other, arena)
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
