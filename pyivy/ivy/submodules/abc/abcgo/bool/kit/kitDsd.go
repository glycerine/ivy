package kit

import (
	"fmt"
	"io"
	"math/bits"
	"os"
	"sort"
	"strings"
)

func Kit_DsdManAlloc(nVars, nNodes int32) *Kit_DsdMan_t {
	if nVars < 0 || nVars > 16 || nNodes < 0 {
		panic("kit: invalid DSD manager size")
	}
	p := &Kit_DsdMan_t{nVars: nVars, vTtElems: make([][]uint32, nVars), vTtNodes: make([][]uint32, nNodes), dd: Cloud_Init(16, 14)}
	for i := range p.vTtElems {
		p.vTtElems[i] = make([]uint32, Kit_TruthWordNum(p.nVars))
		Kit_TruthIthVar(p.vTtElems[i], nVars, int32(i))
	}
	for i := range p.vTtNodes {
		p.vTtNodes[i] = make([]uint32, Kit_TruthWordNum(p.nVars))
	}
	return p
}
func Kit_DsdManFree(p *Kit_DsdMan_t) {
	if p != nil {
		Cloud_Quit(p.dd)
		*p = Kit_DsdMan_t{}
	}
}

// Kit_DsdCloud returns the Cloud manager owned by the DSD manager.
func Kit_DsdCloud(p *Kit_DsdMan_t) *CloudManager { return p.dd }
func Kit_DsdObjAlloc(ntk *Kit_DsdNtk_t, kind Kit_Dsd_t, nFans int32) *Kit_DsdObj_t {
	if nFans < 0 || nFans > 16 || int(ntk.nVars)+len(ntk.pNodes) >= 32767 {
		panic("kit: DSD object dimensions exceed literal representation")
	}
	obj := &Kit_DsdObj_t{Id: uint32(ntk.nVars) + uint32(len(ntk.pNodes)), Type: kind, nFans: uint32(nFans), pFans: make([]uint16, nFans)}
	if kind == KIT_DSD_PRIME {
		obj.truth = make([]uint32, Kit_TruthWordNum(nFans))
		obj.Offset = Kit_DsdObjOffset(nFans)
	}
	ntk.pNodes = append(ntk.pNodes, obj)
	return obj
}
func Kit_DsdObjFree(ntk *Kit_DsdNtk_t, obj *Kit_DsdObj_t) {
	if obj == nil {
		return
	}
	index := int(obj.Id) - int(ntk.nVars)
	if index < 0 || index >= len(ntk.pNodes) || ntk.pNodes[index] != obj {
		panic("kit: foreign DSD node")
	}
	obj.Type = KIT_DSD_NONE
	obj.pFans = nil
	obj.truth = nil
	obj.nFans = 0
}
func Kit_DsdNtkAlloc(nVars int32) *Kit_DsdNtk_t {
	if nVars < 0 || nVars > 16 {
		panic("kit: DSD supports at most 16 variables")
	}
	return &Kit_DsdNtk_t{nVars: uint16(nVars), pNodes: make([]*Kit_DsdObj_t, 0, 3*nVars+1), pMem: make([]uint32, 6*Kit_TruthWordNum(nVars))}
}
func Kit_DsdNtkFree(ntk *Kit_DsdNtk_t) {
	if ntk != nil {
		*ntk = Kit_DsdNtk_t{}
	}
}
func Kit_DsdObjFanNum(obj *Kit_DsdObj_t) int32 { return int32(obj.nFans) }
func Kit_DsdObjFans(obj *Kit_DsdObj_t) []uint16 {
	return append([]uint16(nil), obj.pFans[:obj.nFans]...)
}
func Kit_DsdDecompose_rec(ntk *Kit_DsdNtk_t, obj *Kit_DsdObj_t, support uint32, parent *uint16, muxLimit int32) {
	n := int32(obj.nFans)
	truth := obj.truth
	if obj.Type != KIT_DSD_PRIME {
		panic("kit: only prime nodes can be decomposed")
	}
	if support != Kit_BitMask(n) {
		compact := int32(bits.OnesCount32(support))
		temp := make([]uint32, Kit_TruthWordNum(n))
		Kit_TruthShrink(temp, truth, compact, n, support, 1)
		k := 0
		for j := int32(0); j < n; j++ {
			if support&(1<<uint(j)) != 0 {
				obj.pFans[k] = obj.pFans[j]
				k++
			}
		}
		obj.nFans = uint32(k)
		n = compact
		support = Kit_BitMask(n)
	}
	if n == 0 {
		obj.Type = KIT_DSD_CONST1
		if truth[0] == 0 {
			*parent ^= 1
		}
		return
	}
	if n == 1 {
		obj.Type = KIT_DSD_NONE
		lit := obj.pFans[0]
		if truth[0]&1 != 0 {
			lit ^= 1
		}
		*parent = lit ^ (*parent & 1)
		return
	}
	words := Kit_TruthWordNum(n)
	low, high := make([]uint32, words), make([]uint32, words)
	// Output AND/XOR and disjoint-support MUX decomposition.
	if obj.fMark == 0 {
		for v := n - 1; v >= 0; v-- {
			Kit_TruthCofactor0New(low, truth, n, v)
			Kit_TruthCofactor1New(high, truth, n, v)
			z0, z1 := Kit_TruthIsConst0(low, n) != 0, Kit_TruthIsConst0(high, n) != 0
			o0, o1 := Kit_TruthIsConst1(low, n) != 0, Kit_TruthIsConst1(high, n) != 0
			opposite := Kit_TruthIsOpposite(low, high, n) != 0
			if !z0 && !z1 && !o0 && !o1 && !opposite {
				s0, s1 := Kit_TruthSupport(low, n), Kit_TruthSupport(high, n)
				if s0&s1 != 0 {
					continue
				}
				a, b := Kit_DsdObjAlloc(ntk, KIT_DSD_PRIME, n), Kit_DsdObjAlloc(ntk, KIT_DSD_PRIME, n)
				copy(a.pFans, obj.pFans)
				copy(b.pFans, obj.pFans)
				copy(a.truth, low)
				copy(b.truth, high)
				control := obj.pFans[v]
				obj.truth[0] = 0xcacacaca
				obj.nFans = 3
				obj.pFans[0] = uint16(2 * a.Id)
				obj.pFans[1] = uint16(2 * b.Id)
				obj.pFans[2] = control
				a.nRefs++
				b.nRefs++
				Kit_DsdDecompose_rec(ntk, a, s0, &obj.pFans[0], muxLimit)
				Kit_DsdDecompose_rec(ntk, b, s1, &obj.pFans[1], muxLimit)
				return
			}
			result := Kit_DsdObjAlloc(ntk, KIT_DSD_AND, 2)
			result.nRefs++
			result.pFans[0] = obj.pFans[v]
			result.pFans[1] = uint16(2 * obj.Id)
			obj.pFans[v] = 127
			support &^= 1 << uint(v)
			*parent = uint16(2*result.Id) ^ (*parent & 1)
			switch {
			case z0:
				copy(truth, high)
			case z1:
				result.pFans[0] ^= 1
				copy(truth, low)
			case o0:
				*parent ^= 1
				result.pFans[1] ^= 1
				copy(truth, high)
			case o1:
				*parent ^= 1
				result.pFans[0] ^= 1
				result.pFans[1] ^= 1
				copy(truth, low)
			case opposite:
				result.Type = KIT_DSD_XOR
				copy(truth, low)
			}
			Kit_DsdDecompose_rec(ntk, obj, support, &result.pFans[1], muxLimit)
			return
		}
	}
	obj.fMark = 1
	// Input disjoint MUX and two-variable AND/XOR decomposition.
	for v := n - 1; v >= 0; v-- {
		Kit_TruthCofactor0New(low, truth, n, v)
		Kit_TruthCofactor1New(high, truth, n, v)
		s0, s1 := Kit_TruthSupport(low, n), Kit_TruthSupport(high, n)
		if s0 == 0 || s1 == 0 {
			obj.fMark = 0
			Kit_DsdDecompose_rec(ntk, obj, support, parent, muxLimit)
			return
		}
		four := [4][]uint32{}
		for k := range four {
			four[k] = make([]uint32, words)
		}
		if bits.OnesCount32(s0&^s1) == 1 && bits.OnesCount32(s1&^s0) == 1 {
			v0, v1 := int32(bits.TrailingZeros32(s0&^s1)), int32(bits.TrailingZeros32(s1&^s0))
			Kit_TruthCofactor0New(four[0], low, n, v0)
			Kit_TruthCofactor1New(four[1], low, n, v0)
			Kit_TruthCofactor0New(four[2], high, n, v1)
			Kit_TruthCofactor1New(four[3], high, n, v1)
			same := Kit_TruthIsEqual(four[0], four[2], n) != 0 && Kit_TruthIsEqual(four[1], four[3], n) != 0
			cross := Kit_TruthIsEqual(four[0], four[3], n) != 0 && Kit_TruthIsEqual(four[1], four[2], n) != 0
			if same || cross {
				result := Kit_DsdObjAlloc(ntk, KIT_DSD_PRIME, 3)
				result.truth[0] = 0xcacacaca
				result.nRefs++
				result.pFans[0] = obj.pFans[v0]
				result.pFans[1] = obj.pFans[v1]
				result.pFans[2] = obj.pFans[v]
				if cross {
					result.pFans[0] ^= 1
				}
				obj.pFans[v0] = 127
				obj.pFans[v1] = 127
				obj.pFans[v] = uint16(2 * result.Id)
				support &^= 1<<uint(v0) | 1<<uint(v1)
				Kit_TruthMuxVar(truth, four[2], four[3], n, v)
				Kit_DsdDecompose_rec(ntk, obj, support, parent, muxLimit)
				return
			}
		}
		for k := v + 1; k < n; k++ {
			Kit_TruthCofactor0New(four[0], low, n, k)
			Kit_TruthCofactor1New(four[1], low, n, k)
			Kit_TruthCofactor0New(four[2], high, n, k)
			Kit_TruthCofactor1New(four[3], high, n, k)
			equal := [4][4]bool{}
			pairs := 0
			for a := 0; a < 4; a++ {
				for b := a + 1; b < 4; b++ {
					equal[a][b] = Kit_TruthIsEqual(four[a], four[b], n) != 0
					equal[b][a] = equal[a][b]
					if equal[a][b] {
						pairs++
					}
				}
			}
			if pairs != 2 && pairs != 3 {
				continue
			}
			result := Kit_DsdObjAlloc(ntk, KIT_DSD_AND, 2)
			result.nRefs++
			result.pFans[0] = obj.pFans[k]
			result.pFans[1] = obj.pFans[v]
			obj.pFans[k] = uint16(2 * result.Id)
			obj.pFans[v] = 127
			support &^= 1 << uint(v)
			unique := -1
			for a := 0; a < 4; a++ {
				count := 0
				for b := 0; b < 4; b++ {
					if equal[a][b] {
						count++
					}
				}
				if count == 0 {
					unique = a
					break
				}
			}
			switch unique {
			case 0:
				result.pFans[0] ^= 1
				result.pFans[1] ^= 1
				Kit_TruthMuxVar(truth, four[3], four[0], n, k)
			case 1:
				result.pFans[1] ^= 1
				Kit_TruthMuxVar(truth, four[0], four[1], n, k)
			case 2:
				result.pFans[0] ^= 1
				Kit_TruthMuxVar(truth, four[0], four[2], n, k)
			case 3:
				Kit_TruthMuxVar(truth, four[0], four[3], n, k)
			default:
				result.Type = KIT_DSD_XOR
				Kit_TruthMuxVar(truth, four[0], four[1], n, k)
			}
			Kit_DsdDecompose_rec(ntk, obj, support, parent, muxLimit)
			return
		}
	}
	if muxLimit > 0 && n > muxLimit {
		v := Kit_TruthBestCofVar(truth, n, low, high)
		s0, s1 := Kit_TruthSupport(low, n), Kit_TruthSupport(high, n)
		a, b := Kit_DsdObjAlloc(ntk, KIT_DSD_PRIME, n), Kit_DsdObjAlloc(ntk, KIT_DSD_PRIME, n)
		copy(a.pFans, obj.pFans)
		copy(b.pFans, obj.pFans)
		copy(a.truth, low)
		copy(b.truth, high)
		control := obj.pFans[v]
		obj.truth[0] = 0xcacacaca
		obj.nFans = 3
		obj.pFans[0] = uint16(2 * a.Id)
		obj.pFans[1] = uint16(2 * b.Id)
		obj.pFans[2] = control
		a.nRefs++
		b.nRefs++
		Kit_DsdDecompose_rec(ntk, a, s0, &obj.pFans[0], muxLimit)
		Kit_DsdDecompose_rec(ntk, b, s1, &obj.pFans[1], muxLimit)
	}
}
func Kit_DsdDecomposeInt(truth []uint32, nVars, muxLimit int32) *Kit_DsdNtk_t {
	ntk := Kit_DsdNtkAlloc(nVars)
	obj := Kit_DsdObjAlloc(ntk, KIT_DSD_PRIME, nVars)
	ntk.Root = uint16(2 * obj.Id)
	for i := range obj.pFans {
		obj.pFans[i] = uint16(2 * i)
	}
	copy(obj.truth, truth[:Kit_TruthWordNum(nVars)])
	support := Kit_TruthSupport(truth, nVars)
	if support == 0 {
		obj.Type = KIT_DSD_CONST1
		obj.nFans = 0
		if truth[0] == 0 {
			ntk.Root ^= 1
		}
		return ntk
	}
	if bits.OnesCount32(support) == 1 {
		obj.Type = KIT_DSD_VAR
		obj.nFans = 1
		obj.pFans[0] = uint16(2*bits.TrailingZeros32(support)) | uint16(truth[0]&1)
		return ntk
	}
	Kit_DsdDecompose_rec(ntk, obj, support, &ntk.Root, muxLimit)
	return ntk
}
func Kit_DsdDecompose(truth []uint32, nVars int32) *Kit_DsdNtk_t {
	return Kit_DsdDecomposeInt(truth, nVars, 0)
}
func Kit_DsdDecomposeMux(truth []uint32, nVars, limit int32) *Kit_DsdNtk_t {
	return Kit_DsdDecomposeInt(truth, nVars, limit)
}
func Kit_DsdNonDsdPrimeMax(ntk *Kit_DsdNtk_t) *Kit_DsdObj_t {
	var result *Kit_DsdObj_t
	for _, obj := range ntk.pNodes {
		if obj.Type == KIT_DSD_PRIME && (result == nil || obj.nFans > result.nFans) {
			result = obj
		}
	}
	return result
}
func Kit_DsdNonDsdSizeMax(ntk *Kit_DsdNtk_t) int32 {
	obj := Kit_DsdNonDsdPrimeMax(ntk)
	if obj == nil {
		return 0
	}
	return int32(obj.nFans)
}

func kitDsdComputeNode(ntk *Kit_DsdNtk_t, id int32, support uint32, cache map[int32][]uint32) []uint32 {
	if out := cache[id]; out != nil {
		return out
	}
	n := int32(ntk.nVars)
	words := Kit_TruthWordNum(n)
	out := make([]uint32, words)
	if id < int32(ntk.nVars) {
		Kit_TruthIthVar(out, n, id)
		cache[id] = out
		return out
	}
	obj := Kit_DsdNtkObj(ntk, id)
	if obj.Type == KIT_DSD_CONST1 {
		Kit_TruthFill(out, n)
		cache[id] = out
		return out
	}
	inputs := make([][]uint32, obj.nFans)
	selected := 0
	for i, lit := range obj.pFans[:obj.nFans] {
		if support != 0 && Kit_DsdLitSupport(ntk, int32(lit))&support == 0 {
			continue
		}
		inputs[i] = kitDsdComputeNode(ntk, int32(lit>>1), support, cache)
		selected++
	}
	if obj.Type == KIT_DSD_PRIME && selected < int(obj.nFans) {
		for _, input := range inputs {
			if input != nil {
				return input
			}
		}
		panic("kit: partial support misses prime node")
	}
	for x := int32(0); x < words*32; x++ {
		value := obj.Type == KIT_DSD_AND
		index := int32(0)
		for i, input := range inputs {
			if input == nil {
				continue
			}
			bit := Kit_TruthHasBit(input, x) != 0
			if obj.pFans[i]&1 != 0 {
				bit = !bit
			}
			switch obj.Type {
			case KIT_DSD_AND:
				value = value && bit
			case KIT_DSD_XOR:
				value = value != bit
			case KIT_DSD_VAR, KIT_DSD_NONE:
				value = bit
			case KIT_DSD_PRIME:
				if bit {
					index |= 1 << uint(i)
				}
			}
		}
		if obj.Type == KIT_DSD_PRIME {
			value = Kit_TruthHasBit(obj.truth, index) != 0
		}
		if value {
			Kit_TruthSetBit(out, x)
		}
	}
	cache[id] = out
	return out
}
func Kit_DsdTruthComputeNode_rec(p *Kit_DsdMan_t, ntk *Kit_DsdNtk_t, id int32) []uint32 {
	return kitDsdComputeNode(ntk, id, 0, map[int32][]uint32{})
}
func Kit_DsdTruthCompute(p *Kit_DsdMan_t, ntk *Kit_DsdNtk_t) []uint32 {
	return Kit_DsdTruthComputeOne(p, ntk, 0)
}
func Kit_DsdTruthComputeNodeOne_rec(p *Kit_DsdMan_t, ntk *Kit_DsdNtk_t, id int32, support uint32) []uint32 {
	if support != 0 {
		Kit_DsdGetSupports(ntk)
	}
	return kitDsdComputeNode(ntk, id, support, map[int32][]uint32{})
}
func Kit_DsdTruthComputeOne(p *Kit_DsdMan_t, ntk *Kit_DsdNtk_t, support uint32) []uint32 {
	if int32(ntk.nVars) > p.nVars {
		panic("kit: DSD manager too small")
	}
	out := Kit_DsdTruthComputeNodeOne_rec(p, ntk, int32(ntk.Root>>1), support)
	if ntk.Root&1 != 0 {
		Kit_TruthNot(out, out, int32(ntk.nVars))
	}
	return out
}
func Kit_DsdTruth(ntk *Kit_DsdNtk_t, out []uint32) {
	p := Kit_DsdManAlloc(int32(ntk.nVars), int32(len(ntk.pNodes)))
	copy(out, Kit_DsdTruthCompute(p, ntk))
}
func Kit_DsdTruthPartial(p *Kit_DsdMan_t, ntk *Kit_DsdNtk_t, out []uint32, support uint32) {
	copy(out, Kit_DsdTruthComputeOne(p, ntk, support))
}
func kitDsdSplitTruth(truth []uint32, nVars int32, support uint32, variable int32, dec []uint32) []uint32 {
	if variable < 0 || variable >= nVars || support&(1<<uint(variable)) == 0 {
		panic("kit: replacement variable must belong to bound support")
	}
	bound, free := []int32{}, []int32{}
	for v := int32(0); v < nVars; v++ {
		if support&(1<<uint(v)) != 0 {
			bound = append(bound, v)
		} else {
			free = append(free, v)
		}
	}
	scatter := func(index int, vars []int32) int32 {
		var x int32
		for b, v := range vars {
			x |= int32(index>>uint(b)&1) << uint(v)
		}
		return x
	}
	var patterns [][]bool
	labels := make([]int, 1<<uint(len(bound)))
	for b := range labels {
		values := make([]bool, 1<<uint(len(free)))
		for f := range values {
			values[f] = Kit_TruthHasBit(truth, scatter(b, bound)|scatter(f, free)) != 0
		}
		match := -1
		for k, pattern := range patterns {
			equal := true
			for i, value := range values {
				if value != pattern[i] {
					equal = false
					break
				}
			}
			if equal {
				match = k
				break
			}
		}
		if match < 0 {
			if len(patterns) >= 2 {
				panic("kit: bound support does not form a Boolean decomposition")
			}
			match = len(patterns)
			patterns = append(patterns, values)
		}
		labels[b] = match
	}
	out := make([]uint32, Kit_TruthWordNum(nVars))
	clear(dec[:Kit_TruthWordNum(nVars)])
	for x := int32(0); x < 32*Kit_TruthWordNum(nVars); x++ {
		b, f := 0, 0
		for i, v := range bound {
			b |= int(x>>uint(v)&1) << uint(i)
		}
		for i, v := range free {
			f |= int(x>>uint(v)&1) << uint(i)
		}
		if labels[b] != 0 {
			Kit_TruthSetBit(dec, x)
		}
		label := int(x >> uint(variable) & 1)
		if label >= len(patterns) {
			label = 0
		}
		if patterns[label][f] {
			Kit_TruthSetBit(out, x)
		}
	}
	return out
}
func Kit_DsdTruthComputeNodeTwo_rec(p *Kit_DsdMan_t, ntk *Kit_DsdNtk_t, id int32, support uint32, variable int32, dec []uint32) []uint32 {
	truth := Kit_DsdTruthComputeNode_rec(p, ntk, id)
	return kitDsdSplitTruth(truth, int32(ntk.nVars), support, variable, dec)
}
func Kit_DsdTruthComputeTwo(p *Kit_DsdMan_t, ntk *Kit_DsdNtk_t, support uint32, variable int32, dec []uint32) []uint32 {
	truth := Kit_DsdTruthCompute(p, ntk)
	return kitDsdSplitTruth(truth, int32(ntk.nVars), support, variable, dec)
}
func Kit_DsdTruthPartialTwo(p *Kit_DsdMan_t, ntk *Kit_DsdNtk_t, support uint32, variable int32, co, dec []uint32) {
	copy(co, Kit_DsdTruthComputeTwo(p, ntk, support, variable, dec))
}
func Kit_DsdGetSupports_rec(ntk *Kit_DsdNtk_t, lit int32) uint32 {
	id := lit >> 1
	if id < int32(ntk.nVars) {
		return 1 << uint(id)
	}
	obj := Kit_DsdNtkObj(ntk, id)
	var support uint32
	for _, fan := range obj.pFans[:obj.nFans] {
		support |= Kit_DsdGetSupports_rec(ntk, int32(fan))
	}
	ntk.pSupps[id-int32(ntk.nVars)] = support
	return support
}
func Kit_DsdGetSupports(ntk *Kit_DsdNtk_t) uint32 {
	ntk.pSupps = make([]uint32, len(ntk.pNodes))
	return Kit_DsdGetSupports_rec(ntk, int32(ntk.Root))
}
func Kit_DsdNonDsdSupports(ntk *Kit_DsdNtk_t) uint32 {
	Kit_DsdGetSupports(ntk)
	var result uint32
	for _, obj := range ntk.pNodes {
		if obj.Type == KIT_DSD_PRIME {
			result |= Kit_DsdLitSupport(ntk, int32(2*obj.Id))
		}
	}
	return result
}
func Kit_DsdFindLargeBox_rec(ntk *Kit_DsdNtk_t, id, size int32) int32 {
	obj := Kit_DsdNtkObj(ntk, id)
	if obj == nil {
		return 0
	}
	if obj.Type == KIT_DSD_PRIME && int32(obj.nFans) > size {
		return 1
	}
	for _, fan := range obj.pFans[:obj.nFans] {
		if Kit_DsdFindLargeBox_rec(ntk, int32(fan>>1), size) != 0 {
			return 1
		}
	}
	return 0
}
func Kit_DsdFindLargeBox(ntk *Kit_DsdNtk_t, size int32) int32 {
	return Kit_DsdFindLargeBox_rec(ntk, int32(ntk.Root>>1), size)
}
func Kit_DsdCountLuts_rec(ntk *Kit_DsdNtk_t, size, id int32, count *int32) int32 {
	obj := Kit_DsdNtkObj(ntk, id)
	if obj == nil || obj.Type == KIT_DSD_CONST1 || obj.Type == KIT_DSD_VAR {
		return 0
	}
	if obj.Type == KIT_DSD_AND || obj.Type == KIT_DSD_XOR {
		if obj.nFans != 2 {
			panic("kit: LUT counting requires a binary DSD network")
		}
		a := Kit_DsdCountLuts_rec(ntk, size, int32(obj.pFans[0]>>1), count)
		b := Kit_DsdCountLuts_rec(ntk, size, int32(obj.pFans[1]>>1), count)
		if a == 0 && b > 0 {
			return b - 1
		}
		if a > 0 && b == 0 {
			return a - 1
		}
		*count++
		return size - 2
	}
	if int32(obj.nFans) > size {
		*count = 1000
		return 0
	}
	for _, fan := range obj.pFans[:obj.nFans] {
		Kit_DsdCountLuts_rec(ntk, size, int32(fan>>1), count)
	}
	*count++
	return size - int32(obj.nFans)
}
func Kit_DsdCountLuts(ntk *Kit_DsdNtk_t, size int32) int32 {
	var count int32
	Kit_DsdCountLuts_rec(ntk, size, int32(ntk.Root>>1), &count)
	if count >= 1000 {
		return -1
	}
	return count
}
func kitDsdAigCost(obj *Kit_DsdObj_t) int32 {
	switch obj.Type {
	case KIT_DSD_AND:
		return max(0, int32(obj.nFans)-1)
	case KIT_DSD_XOR:
		return max(0, int32(obj.nFans)-1) * 3
	case KIT_DSD_PRIME:
		if obj.nFans == 3 {
			return 3
		}
		g := Kit_TruthToGraph(obj.truth, int32(obj.nFans), nil)
		if g == nil {
			return -1
		}
		return Kit_GraphNodeNum(g)
	}
	return 0
}
func Kit_DsdCountAigNodes_rec(ntk *Kit_DsdNtk_t, id int32) int32 {
	obj := Kit_DsdNtkObj(ntk, id)
	if obj == nil {
		return 0
	}
	cost := kitDsdAigCost(obj)
	for _, fan := range obj.pFans[:obj.nFans] {
		cost += Kit_DsdCountAigNodes_rec(ntk, int32(fan>>1))
	}
	return cost
}
func Kit_DsdCountAigNodes2(ntk *Kit_DsdNtk_t) int32 {
	return Kit_DsdCountAigNodes_rec(ntk, int32(ntk.Root>>1))
}
func Kit_DsdCountAigNodes(ntk *Kit_DsdNtk_t) int32 {
	var cost int32
	for _, obj := range ntk.pNodes {
		cost += kitDsdAigCost(obj)
	}
	return cost
}
func Kit_DsdRootNodeHasCommonVars(a, b *Kit_DsdObj_t) int32 {
	if a == nil || b == nil {
		return 0
	}
	for _, x := range a.pFans[:a.nFans] {
		if x>>1 >= 4 {
			continue
		}
		for _, y := range b.pFans[:b.nFans] {
			if x>>1 == y>>1 {
				return 1
			}
		}
	}
	return 0
}
func Kit_DsdCheckVar4Dec2(a, b *Kit_DsdNtk_t) int32 {
	if a.nVars != 4 || b.nVars != 4 {
		panic("kit: expected four-variable decompositions")
	}
	if Kit_DsdFindLargeBox(a, 2) != 0 || Kit_DsdFindLargeBox(b, 2) != 0 {
		return 0
	}
	return Kit_DsdRootNodeHasCommonVars(Kit_DsdNtkRoot(a), Kit_DsdNtkRoot(b))
}
func kitDsdPrint(w io.Writer, ntk *Kit_DsdNtk_t, id int32, hex bool) {
	obj := Kit_DsdNtkObj(ntk, id)
	if obj == nil {
		fmt.Fprintf(w, "%c", 'a'+id)
		return
	}
	if obj.Type == KIT_DSD_CONST1 {
		fmt.Fprint(w, "Const1")
		return
	}
	separator := " , "
	if obj.Type == KIT_DSD_AND {
		separator = "*"
	}
	if obj.Type == KIT_DSD_XOR {
		separator = "+"
	}
	if obj.Type == KIT_DSD_PRIME && hex {
		Kit_DsdPrintHex(w, obj.truth, int32(obj.nFans))
	}
	opening, closing := "(", ")"
	if obj.Type == KIT_DSD_PRIME && !hex {
		opening, closing = "[", "]"
	}
	fmt.Fprint(w, opening)
	for i, fan := range obj.pFans[:obj.nFans] {
		if i > 0 {
			fmt.Fprint(w, separator)
		}
		if fan&1 != 0 {
			fmt.Fprint(w, "!")
		}
		kitDsdPrint(w, ntk, int32(fan>>1), hex)
	}
	fmt.Fprint(w, closing)
}
func Kit_DsdPrintHex(w io.Writer, truth []uint32, nFans int32) { Kit_PrintHexadecimal(w, truth, nFans) }
func Kit_DsdWriteHex(prefix string, truth []uint32, nFans int32) string {
	var text strings.Builder
	text.WriteString(prefix)
	Kit_DsdPrintHex(&text, truth, nFans)
	return text.String()
}
func Kit_DsdPrint_rec(w io.Writer, ntk *Kit_DsdNtk_t, id int32)  { kitDsdPrint(w, ntk, id, true) }
func Kit_DsdPrint2_rec(w io.Writer, ntk *Kit_DsdNtk_t, id int32) { kitDsdPrint(w, ntk, id, false) }
func Kit_DsdPrint(w io.Writer, ntk *Kit_DsdNtk_t) {
	if ntk.Root&1 != 0 {
		fmt.Fprint(w, "!")
	}
	Kit_DsdPrint_rec(w, ntk, int32(ntk.Root>>1))
}
func Kit_DsdPrint2(w io.Writer, ntk *Kit_DsdNtk_t) {
	if ntk.Root&1 != 0 {
		fmt.Fprint(w, "!")
	}
	Kit_DsdPrint2_rec(w, ntk, int32(ntk.Root>>1))
}
func Kit_DsdWrite_rec(prefix string, ntk *Kit_DsdNtk_t, id int32) string {
	var text strings.Builder
	text.WriteString(prefix)
	Kit_DsdPrint_rec(&text, ntk, id)
	return text.String()
}
func Kit_DsdWrite(ntk *Kit_DsdNtk_t) string {
	var text strings.Builder
	Kit_DsdPrint(&text, ntk)
	return text.String()
}
func Kit_DsdPrintExpanded(ntk *Kit_DsdNtk_t) {
	Kit_DsdPrint(os.Stdout, Kit_DsdExpand(ntk))
	fmt.Println()
}
func Kit_DsdPrintFromTruth(truth []uint32, nVars int32) {
	Kit_DsdPrintFromTruth2(os.Stdout, truth, nVars)
}
func Kit_DsdPrintFromTruth2(w io.Writer, truth []uint32, nVars int32) {
	Kit_DsdPrint(w, Kit_DsdDecompose(truth, nVars))
	fmt.Fprintln(w)
}
func Kit_DsdWriteFromTruth(truth []uint32, nVars int32) string {
	return Kit_DsdWrite(Kit_DsdDecompose(truth, nVars))
}

func Kit_DsdExpandCollectAnd_rec(ntk *Kit_DsdNtk_t, lit uint32, out []uint32, count *int32) {
	obj := Kit_DsdNtkObj(ntk, int32(lit>>1))
	if lit&1 != 0 || obj == nil || obj.Type != KIT_DSD_AND {
		out[*count] = lit
		*count++
		return
	}
	for _, fan := range obj.pFans[:obj.nFans] {
		Kit_DsdExpandCollectAnd_rec(ntk, uint32(fan), out, count)
	}
}
func Kit_DsdExpandCollectXor_rec(ntk *Kit_DsdNtk_t, lit uint32, out []uint32, count *int32) {
	obj := Kit_DsdNtkObj(ntk, int32(lit>>1))
	if obj == nil || obj.Type != KIT_DSD_XOR {
		out[*count] = lit
		*count++
		return
	}
	start := *count
	for _, fan := range obj.pFans[:obj.nFans] {
		Kit_DsdExpandCollectXor_rec(ntk, uint32(fan), out, count)
	}
	if lit&1 != 0 && *count > start {
		out[start] ^= 1
	}
}
func Kit_DsdExpandNode_rec(next, old *Kit_DsdNtk_t, lit int32) int32 {
	obj := Kit_DsdNtkObj(old, lit>>1)
	if obj == nil {
		return lit
	}
	if obj.Type == KIT_DSD_VAR || obj.Type == KIT_DSD_NONE {
		if obj.nFans == 1 {
			return Kit_DsdExpandNode_rec(next, old, int32(obj.pFans[0])) ^ (lit & 1)
		}
	}
	if obj.Type == KIT_DSD_CONST1 {
		node := Kit_DsdObjAlloc(next, KIT_DSD_CONST1, 0)
		return int32(2*node.Id) ^ (lit & 1)
	}
	if obj.Type == KIT_DSD_AND || obj.Type == KIT_DSD_XOR {
		fans := make([]uint32, 16)
		var count int32
		if obj.Type == KIT_DSD_AND {
			Kit_DsdExpandCollectAnd_rec(old, uint32(lit)&^1, fans, &count)
		} else {
			Kit_DsdExpandCollectXor_rec(old, uint32(lit)&^1, fans, &count)
		}
		node := Kit_DsdObjAlloc(next, obj.Type, count)
		compl := lit & 1
		for i := int32(0); i < count; i++ {
			fan := int32(fans[i])
			if obj.Type == KIT_DSD_XOR {
				compl ^= fan & 1
				fan &^= 1
			}
			node.pFans[i] = uint16(Kit_DsdExpandNode_rec(next, old, fan))
		}
		return int32(2*node.Id) ^ compl
	}
	node := Kit_DsdObjAlloc(next, KIT_DSD_PRIME, int32(obj.nFans))
	copy(node.truth, obj.truth)
	for i, fan := range obj.pFans[:obj.nFans] {
		node.pFans[i] = uint16(Kit_DsdExpandNode_rec(next, old, int32(fan)))
		if node.pFans[i]&1 != 0 {
			node.pFans[i] &^= 1
			Kit_TruthChangePhase(node.truth, int32(node.nFans), int32(i))
		}
	}
	if lit&1 != 0 {
		Kit_TruthNot(node.truth, node.truth, int32(node.nFans))
	}
	return int32(2 * node.Id)
}

// Trivial roots retain an actual CONST1/VAR object, as the original C
// wrappers do. Collapsing VAR to a leaf literal breaks root-object clients.
func kitDsdCopyTrivialRoot(next, old *Kit_DsdNtk_t) bool {
	root := Kit_DsdNtkRoot(old)
	if root == nil || root.Type != KIT_DSD_CONST1 && root.Type != KIT_DSD_VAR {
		return false
	}
	node := Kit_DsdObjAlloc(next, root.Type, int32(root.nFans))
	copy(node.pFans, root.pFans[:root.nFans])
	next.Root = uint16(2*node.Id) | old.Root&1
	return true
}
func Kit_DsdExpand(old *Kit_DsdNtk_t) *Kit_DsdNtk_t {
	next := Kit_DsdNtkAlloc(int32(old.nVars))
	if kitDsdCopyTrivialRoot(next, old) {
		return next
	}
	next.Root = uint16(Kit_DsdExpandNode_rec(next, old, int32(old.Root)))
	return next
}
func Kit_DsdDecomposeExpand(truth []uint32, nVars int32) *Kit_DsdNtk_t {
	return Kit_DsdExpand(Kit_DsdDecompose(truth, nVars))
}
func Kit_DsdCompSort(priorities []int32, supports []uint32, lits []uint16, nVars int32, out []uint32) {
	type component struct {
		index          int
		priority, size int32
	}
	components := make([]component, nVars)
	best := 0
	for i := range components {
		priority := int32(KIT_INFINITY)
		for v := 0; v < len(priorities) && v < 32; v++ {
			if supports[i]&(1<<uint(v)) != 0 {
				priority = min(priority, priorities[v])
			}
		}
		components[i] = component{i, priority, int32(bits.OnesCount32(supports[i]))}
		if i > 0 && (components[i].size > components[best].size || components[i].size == components[best].size && components[i].priority > components[best].priority) {
			best = i
		}
	}
	if len(components) == 0 {
		return
	}
	out[0] = uint32(lits[best])
	sort.SliceStable(components, func(i, j int) bool { return components[i].priority > components[j].priority })
	pos := 1
	for _, c := range components {
		if c.index != best {
			out[pos] = uint32(lits[c.index])
			pos++
		}
	}
}
func Kit_DsdShrink_rec(next, old *Kit_DsdNtk_t, lit int32, priorities []int32) int32 {
	obj := Kit_DsdNtkObj(old, lit>>1)
	if obj == nil {
		return lit
	}
	if obj.Type == KIT_DSD_VAR || obj.Type == KIT_DSD_NONE {
		if obj.nFans == 1 {
			return Kit_DsdShrink_rec(next, old, int32(obj.pFans[0]), priorities) ^ (lit & 1)
		}
	}
	if obj.Type == KIT_DSD_CONST1 {
		node := Kit_DsdObjAlloc(next, KIT_DSD_CONST1, 0)
		return int32(2*node.Id) ^ (lit & 1)
	}
	if obj.Type == KIT_DSD_AND || obj.Type == KIT_DSD_XOR {
		supports := make([]uint32, obj.nFans)
		for i, fan := range obj.pFans[:obj.nFans] {
			supports[i] = Kit_DsdLitSupport(old, int32(fan))
		}
		order := make([]uint32, obj.nFans)
		Kit_DsdCompSort(priorities, supports, obj.pFans, int32(obj.nFans), order)
		result := Kit_DsdShrink_rec(next, old, int32(order[0]), priorities)
		for _, fan := range order[1:] {
			node := Kit_DsdObjAlloc(next, obj.Type, 2)
			node.pFans[0] = uint16(Kit_DsdShrink_rec(next, old, int32(fan), priorities))
			node.pFans[1] = uint16(result)
			result = int32(2 * node.Id)
		}
		return result ^ (lit & 1)
	}
	node := Kit_DsdObjAlloc(next, KIT_DSD_PRIME, int32(obj.nFans))
	copy(node.truth, obj.truth)
	for i, fan := range obj.pFans[:obj.nFans] {
		node.pFans[i] = uint16(Kit_DsdShrink_rec(next, old, int32(fan), priorities))
		if node.pFans[i]&1 != 0 {
			node.pFans[i] &^= 1
			Kit_TruthChangePhase(node.truth, int32(node.nFans), int32(i))
		}
	}
	if lit&1 != 0 {
		Kit_TruthNot(node.truth, node.truth, int32(node.nFans))
	}
	return int32(2 * node.Id)
}
func Kit_DsdShrink(old *Kit_DsdNtk_t, priorities []int32) *Kit_DsdNtk_t {
	next := Kit_DsdNtkAlloc(int32(old.nVars))
	if kitDsdCopyTrivialRoot(next, old) {
		return next
	}
	if len(priorities) < int(old.nVars) {
		panic("kit: short DSD priority array")
	}
	Kit_DsdGetSupports(old)
	next.Root = uint16(Kit_DsdShrink_rec(next, old, int32(old.Root), priorities))
	return next
}
func Kit_DsdRotate(ntk *Kit_DsdNtk_t, frequencies []int32) {
	if len(frequencies) < int(ntk.nVars) {
		panic("kit: short DSD frequency array")
	}
	Kit_DsdGetSupports(ntk)
	for _, obj := range ntk.pNodes {
		if obj.Type != KIT_DSD_PRIME {
			continue
		}
		best, weight := -1, int32(0)
		for i, fan := range obj.pFans[:obj.nFans] {
			support := Kit_DsdLitSupport(ntk, int32(fan))
			var current int32
			for v := 0; v < int(ntk.nVars); v++ {
				if support&(1<<uint(v)) != 0 {
					current += frequencies[v] - 1
				}
			}
			if current > weight {
				best, weight = i, current
			}
		}
		for i := best - 1; i >= 0; i-- {
			obj.pFans[i], obj.pFans[i+1] = obj.pFans[i+1], obj.pFans[i]
			Kit_TruthSwapAdjacentVars(obj.truth, obj.truth, int32(obj.nFans), int32(i))
		}
	}
}
