package if_

import (
	"github.com/glycerine/abcgo/bool/kit"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/opt/dau/canon"
)

func ifTruthWords(vars int) int {
	if vars < 0 || vars > IF_MAX_FUNC_LUTSIZE {
		panic("if: truth support outside range")
	}
	return 1 << uint(max(0, vars-6))
}
func ifTruthBit(table []uint64, assignment int) bool {
	return table[assignment>>6]&(1<<uint(assignment&63)) != 0
}
func ifTruthSet(table []uint64, assignment int) { table[assignment>>6] |= 1 << uint(assignment&63) }
func ifTruthProject(table []uint64, source, target []int32, sourceMask, targetMask uint32, complement bool) []uint64 {
	positions := make([]int, len(source))
	for i, id := range source {
		positions[i] = -1
		for j, other := range target {
			if id == other {
				positions[i] = j
				break
			}
		}
		if positions[i] < 0 {
			panic("if: truth source leaf missing from merged cut")
		}
	}
	out := make([]uint64, ifTruthWords(len(target)))
	for assignment := 0; assignment < len(out)*64; assignment++ {
		input := 0
		for i, pos := range positions {
			bit := assignment >> uint(pos) & 1
			bit ^= int((sourceMask>>uint(i))&1) ^ int((targetMask>>uint(pos))&1)
			input |= bit << uint(i)
		}
		if ifTruthBit(table, input) != complement {
			ifTruthSet(out, assignment)
		}
	}
	return out
}
func ifTruthMinimize(table []uint64, leaves []int32) ([]uint64, []int32) {
	n := len(leaves)
	var support []int
	for v := 0; v < n; v++ {
		essential := false
		for assignment := 0; assignment < 1<<uint(n); assignment++ {
			if assignment&(1<<uint(v)) == 0 && ifTruthBit(table, assignment) != ifTruthBit(table, assignment|(1<<uint(v))) {
				essential = true
				break
			}
		}
		if essential {
			support = append(support, v)
		}
	}
	outLeaves := make([]int32, len(support))
	for i, v := range support {
		outLeaves[i] = leaves[v]
	}
	out := make([]uint64, ifTruthWords(len(support)))
	for assignment := 0; assignment < len(out)*64; assignment++ {
		input := 0
		for i, v := range support {
			input |= (assignment >> uint(i) & 1) << uint(v)
		}
		if ifTruthBit(table, input) {
			ifTruthSet(out, assignment)
		}
	}
	return out, outLeaves
}
func ifTruthRead(p *If_Man_t, c *If_Cut_t, function int32) []uint64 {
	if function < 0 || p.vTtMem[c.nLeaves] == nil {
		panic("if: missing cut truth table")
	}
	return vec.Vec_MemReadEntry(p.vTtMem[c.nLeaves], int(function>>1))
}
func ifTruthIntern(p *If_Man_t, c *If_Cut_t, table []uint64, phase uint32, makeISOP bool) {
	n := int(c.nLeaves)
	if p.vTtMem[n] == nil {
		p.vTtMem[n] = vec.Vec_MemAllocForTT(n, int(p.pPars.FUseTtPerm))
	}
	p.nTruth6Words[n] = int32(ifTruthWords(n))
	before := vec.Vec_MemEntryNum(p.vTtMem[n])
	id := vec.Vec_MemHashInsert(p.vTtMem[n], table)
	c.iCutFunc = int32(id)*2 + int32(phase&1)
	if makeISOP && p.vTtIsops[n] != nil && vec.Vec_MemEntryNum(p.vTtMem[n]) != before {
		if vec.Vec_WecSize(p.vTtIsops[n]) != id {
			panic("if: ISOP truth cache out of sync")
		}
		level := vec.Vec_WecPushLevel(p.vTtIsops[n])
		words := make([]uint32, max(1, len(table)*2))
		for i, w := range table {
			words[2*i], words[2*i+1] = uint32(w), uint32(w>>32)
		}
		var cover []uint32
		complement := kit.Kit_TruthIsop(words, int32(n), &cover, 1)
		if complement >= 0 {
			for _, cube := range cover {
				vec.Vec_IntPush(level, int32(cube))
			}
		}
		cacheIndex := n
		if n < 6 && p.vTtIsops[n] == p.vTtIsops[6] {
			cacheIndex = 6
		}
		if len(p.vTtIsopComplement[cacheIndex]) != id {
			panic("if: ISOP complement cache out of sync")
		}
		p.vTtIsopComplement[cacheIndex] = append(p.vTtIsopComplement[cacheIndex], complement > 0)
	}
}
func If_CutTruthPermute(table []uint64, nLeaves, nVars, nWords int32, delays []float32, vars []int32) {
	if nLeaves < 0 || nLeaves > nVars || int(nLeaves) > len(delays) || int(nLeaves) > len(vars) {
		panic("if: invalid truth permutation dimensions")
	}
	for changed := true; changed; {
		changed = false
		for i := 0; i < int(nLeaves)-1; i++ {
			if delays[i] >= delays[i+1] {
				continue
			}
			delays[i], delays[i+1] = delays[i+1], delays[i]
			vars[i], vars[i+1] = vars[i+1], vars[i]
			if table != nil {
				copyTable := append([]uint64(nil), table[:nWords]...)
				clear(table[:nWords])
				for assignment := 0; assignment < int(nWords)*64; assignment++ {
					input := assignment
					if assignment>>uint(i)&1 != assignment>>uint(i+1)&1 {
						input ^= (1 << uint(i)) | (1 << uint(i+1))
					}
					if ifTruthBit(copyTable, input) {
						ifTruthSet(table, assignment)
					}
				}
			}
			changed = true
		}
	}
}
func If_CutRotatePins(p *If_Man_t, c *If_Cut_t) {
	if p.pPars.FUseTtPerm != 0 {
		panic("if: pin rotation conflicts with canonical leaf phases")
	}
	delays := make([]float32, c.nLeaves)
	for i, id := range ifLeaves(c) {
		delays[i] = If_ManObj(p, id).CutBest.Delay
	}
	var table []uint64
	if p.vTtMem[c.nLeaves] != nil {
		table = append([]uint64(nil), ifTruthRead(p, c, c.iCutFunc)...)
	}
	If_CutTruthPermute(table, int32(c.nLeaves), int32(c.nLeaves), int32(ifTruthWords(int(c.nLeaves))), delays, ifLeaves(c))
	if table != nil {
		phase := uint32(c.iCutFunc & 1)
		ifTruthIntern(p, c, table, phase, false)
	}
}
func If_CutComputeTruth(p *If_Man_t, c, a, b *If_Cut_t, complementA, complementB int32) int32 {
	leaves := append([]int32(nil), ifLeaves(c)...)
	x := ifTruthProject(ifTruthRead(p, a, a.iCutFunc), ifLeaves(a), leaves, 0, 0, (complementA^int32(a.fCompl)^a.iCutFunc)&1 != 0)
	y := ifTruthProject(ifTruthRead(p, b, b.iCutFunc), ifLeaves(b), leaves, 0, 0, (complementB^int32(b.fCompl)^b.iCutFunc)&1 != 0)
	phase := uint32(x[0] & y[0] & 1)
	for i := range x {
		x[i] &= y[i]
		if phase != 0 {
			x[i] = ^x[i]
		}
	}
	reduced := int32(0)
	if p.pPars.FCutMin != 0 && (a.nLeaves+b.nLeaves > c.nLeaves || a.nLeaves == 0 || b.nLeaves == 0) {
		x, leaves = ifTruthMinimize(x, leaves)
		if len(leaves) < int(c.nLeaves) {
			reduced = 1
			c.nLeaves = uint32(len(leaves))
			c.pLeaves = append(c.pLeaves[:0], leaves...)
			c.uSign = If_ObjCutSignCompute(c)
		}
	}
	ifTruthIntern(p, c, x, phase, true)
	return reduced
}
func If_CutComputeTruthPerm_int(p *If_Man_t, c, a, b *If_Cut_t, functionA, functionB int32) int32 {
	if a.uMaskFunc < 0 || b.uMaskFunc < 0 {
		panic("if: missing canonical cut phases")
	}
	leaves := append([]int32(nil), ifLeaves(a)...)
	phases := uint32(a.uMaskFunc)
	for i, id := range ifLeaves(b) {
		found := false
		for _, other := range leaves {
			found = found || id == other
		}
		if !found {
			if uint32(b.uMaskFunc)&(1<<uint(i)) != 0 {
				phases |= 1 << uint(len(leaves))
			}
			leaves = append(leaves, id)
		}
	}
	if len(leaves) != int(c.nLeaves) {
		panic("if: canonical merge does not match cut leaves")
	}
	x := ifTruthProject(ifTruthRead(p, a, functionA), ifLeaves(a), leaves, uint32(a.uMaskFunc), phases, functionA&1 != 0)
	y := ifTruthProject(ifTruthRead(p, b, functionB), ifLeaves(b), leaves, uint32(b.uMaskFunc), phases, functionB&1 != 0)
	for i := range x {
		x[i] &= y[i]
	}
	encoded := make([]int32, len(leaves))
	for i, id := range leaves {
		encoded[i] = id*2 + int32(phases>>uint(i)&1)
	}
	reduced := int32(0)
	if p.pPars.FCutMin != 0 && (a.nLeaves+b.nLeaves > c.nLeaves || a.nLeaves == 0 || b.nLeaves == 0) {
		x, encoded = ifTruthMinimize(x, encoded)
		if len(encoded) < int(c.nLeaves) {
			reduced = 1
			c.nLeaves = uint32(len(encoded))
		}
	}
	p.uCanonPhase = canon.Canonicalize(x, int32(c.nLeaves), p.pCanonPerm[:])
	c.pLeaves = make([]int32, c.nLeaves)
	c.uMaskFunc = 0
	for i := range c.pLeaves {
		lit := encoded[p.pCanonPerm[i]] ^ int32(p.uCanonPhase>>uint(i)&1)
		c.pLeaves[i] = lit >> 1
		c.uMaskFunc |= (lit & 1) << uint(i)
	}
	c.uSign = If_ObjCutSignCompute(c)
	ifTruthIntern(p, c, x, p.uCanonPhase>>c.nLeaves, false)
	n := int(c.nLeaves)
	if p.vTtOccurs[n] == nil {
		p.vTtOccurs[n] = vec.Vec_IntAlloc(16)
	}
	for vec.Vec_IntSize(p.vTtOccurs[n]) < vec.Vec_MemEntryNum(p.vTtMem[n]) {
		vec.Vec_IntPush(p.vTtOccurs[n], 0)
	}
	vec.Vec_IntAddToEntry(p.vTtOccurs[n], int(c.iCutFunc>>1), 1)
	return reduced
}
func If_CutComputeTruthPerm(p *If_Man_t, c, a, b *If_Cut_t, functionA, functionB int32) int32 {
	if a.nLeaves+b.nLeaves > c.nLeaves || functionA < 2 || functionB < 2 {
		return If_CutComputeTruthPerm_int(p, c, a, b, functionA, functionB)
	}
	if p.vPairHash == nil {
		p.vPairHash = vec.Hash_IntManStart(100)
	}
	if p.vPairRes == nil {
		p.vPairRes = vec.Vec_IntStart(1)
	}
	stride := int(c.nLimit)
	if stride <= 0 {
		panic("if: canonical pair cache needs positive LUT limit")
	}
	if p.vPairPerms == nil {
		p.vPairPerms = vec.Vec_StrStart(stride)
	}
	before := vec.Hash_IntManEntryNum(p.vPairHash)
	index := vec.Hash_Int2ManInsert(p.vPairHash, functionA<<5|int32(a.nLeaves), functionB<<5|int32(b.nLeaves), -1)
	if vec.Hash_IntManEntryNum(p.vPairHash) == before {
		c.iCutFunc = vec.Vec_IntEntry(p.vPairRes, int(index))
		original := append([]int32(nil), ifLeaves(c)...)
		for i := range original {
			mask := uint32(b.uMaskFunc)
			bit := i - int(a.nLeaves)
			if i < int(a.nLeaves) {
				mask = uint32(a.uMaskFunc)
				bit = i
			}
			original[i] = original[i]*2 + int32(mask>>uint(bit)&1)
		}
		c.uMaskFunc = 0
		for i := range ifLeaves(c) {
			perm := vec.Vec_StrEntry(p.vPairPerms, int(index)*stride+i)
			lit := original[perm>>1] ^ int32(perm&1)
			c.pLeaves[i] = lit >> 1
			c.uMaskFunc |= (lit & 1) << uint(i)
		}
		p.nCacheHits++
		return 0
	}
	p.nCacheMisses++
	result := If_CutComputeTruthPerm_int(p, c, a, b, functionA, functionB)
	if result != 0 {
		panic("if: disjoint canonical pair unexpectedly reduced support")
	}
	if int(index) != vec.Vec_IntSize(p.vPairRes) {
		panic("if: canonical pair cache index mismatch")
	}
	vec.Vec_IntPush(p.vPairRes, c.iCutFunc)
	for i := 0; i < stride; i++ {
		perm := byte(255)
		if i < int(c.nLeaves) {
			perm = p.pCanonPerm[i]*2 + byte(p.uCanonPhase>>uint(i)&1)
		}
		vec.Vec_StrPush(p.vPairPerms, perm)
	}
	return 0
}
func If_DeriveHashTable6(nVars int32, truth uint64) *Vec_Mem_t {
	if nVars < 0 || nVars > 6 {
		panic("if: small truth orbit needs 0..6 inputs")
	}
	memory := vec.Vec_MemAllocForTTSimple(int(nVars))
	permutation := make([]int, nVars)
	for i := range permutation {
		permutation[i] = i
	}
	var enumerate func(int)
	enumerate = func(position int) {
		if position < len(permutation) {
			for i := position; i < len(permutation); i++ {
				permutation[i], permutation[position] = permutation[position], permutation[i]
				enumerate(position + 1)
				permutation[i], permutation[position] = permutation[position], permutation[i]
			}
			return
		}
		for phase := 0; phase < 1<<uint(nVars); phase++ {
			var table uint64
			for assignment := 0; assignment < 64; assignment++ {
				input := assignment & ^((1 << uint(nVars)) - 1)
				for i, source := range permutation {
					input |= ((assignment >> uint(i) & 1) ^ int(phase>>uint(i)&1)) << uint(source)
				}
				if truth>>uint(input)&1 != 0 {
					table |= 1 << uint(assignment)
				}
			}
			vec.Vec_MemHashInsert(memory, []uint64{table})
			vec.Vec_MemHashInsert(memory, []uint64{^table})
		}
	}
	enumerate(0)
	return memory
}
func If_CutCheckTruth6(p *If_Man_t, c *If_Cut_t) int32 {
	if c.nLeaves != 6 {
		return 0
	}
	if p.vTtMem6 == nil {
		p.vTtMem6 = If_DeriveHashTable6(6, 0xfedcba9876543210)
	}
	if *vec.Vec_MemHashLookup(p.vTtMem6, ifTruthRead(p, c, c.iCutFunc)) < 0 {
		return 0
	}
	return 1
}
