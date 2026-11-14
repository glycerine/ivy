package if_

import (
	"fmt"
	"math/bits"
	"slices"
	"sort"
)

const (
	CLU_VAR_MAX = 16
	CLU_WRD_MAX = 1 << (CLU_VAR_MAX - 6)
	CLU_MEM_MAX = 1000
	CLU_UNUSED  = 255
)

type If_Grp_t struct {
	nVars, nMyu int32
	pVars       [CLU_VAR_MAX]int32
}
type If_Hte_t struct {
	pNext          *If_Hte_t
	Group, Counter uint32
	pTruth         []uint64
}

func If_CluGrp2Uns(g *If_Grp_t) uint32 {
	result := uint32(g.nVars&15) | uint32(g.nMyu&15)<<4
	for i := 0; i < 6; i++ {
		result |= uint32(g.pVars[i]&15) << uint(8+4*i)
	}
	return result
}
func If_CluUns2Grp(code uint32, g *If_Grp_t) {
	g.nVars, g.nMyu = int32(code&15), int32(code>>4&15)
	for i := 0; i < 6; i++ {
		g.pVars[i] = int32(code >> uint(8+4*i) & 15)
	}
}
func If_CluPrimeCudd(p uint32) uint32 {
	if p <= 1 {
		return 1
	}
	if p == 2 {
		return 3
	}
	for {
		if p&1 != 0 {
			prime := true
			for divisor := uint32(3); uint64(divisor)*uint64(divisor) <= uint64(p); divisor += 2 {
				if p%divisor == 0 {
					prime = false
					break
				}
			}
			if prime {
				return p
			}
		}
		p++
	}
}
func If_CluWordNum(nVars int32) int32   { return int32(ifTruthWords(int(nVars))) }
func If_CluCountOnes(t uint64) int32    { return int32(bits.OnesCount64(t)) }
func If_CluClear(out []uint64, n int32) { clear(out[:If_CluWordNum(n)]) }
func If_CluFill(out []uint64, n int32) {
	for i := range out[:If_CluWordNum(n)] {
		out[i] = ^uint64(0)
	}
}
func If_CluCopy(out, in []uint64, n int32) { copy(out[:If_CluWordNum(n)], in[:If_CluWordNum(n)]) }
func If_CluEqual(a, b []uint64, n int32) int32 {
	if slices.Equal(a[:If_CluWordNum(n)], b[:If_CluWordNum(n)]) {
		return 1
	}
	return 0
}
func If_CluAnd(out, a, b []uint64, n int32) {
	for i := range out[:If_CluWordNum(n)] {
		out[i] = a[i] & b[i]
	}
}
func If_CluSharp(out, a, b []uint64, n int32) {
	for i := range out[:If_CluWordNum(n)] {
		out[i] = a[i] &^ b[i]
	}
}
func If_CluOr(out, a, b []uint64, n int32) {
	for i := range out[:If_CluWordNum(n)] {
		out[i] = a[i] | b[i]
	}
}
func If_CluAdjust(t uint64, n int32) uint64 {
	if n < 0 || n > 6 {
		panic("if: small truth support outside range")
	}
	if n == 6 {
		return t
	}
	period := 1 << uint(n)
	t &= (uint64(1) << uint(period)) - 1
	for period < 64 {
		t |= t << uint(period)
		period *= 2
	}
	return t
}
func If_CluAdjustBig(table []uint64, current, maximum int32) {
	if current < 0 || maximum < current || maximum > 16 {
		panic("if: invalid truth stretch")
	}
	for v := max(current, 6); v < maximum; v++ {
		words := If_CluWordNum(v)
		copy(table[words:2*words], table[:words])
	}
}
func cluTransform(out, in []uint64, n int32, assignment func(int) int) {
	source := append([]uint64(nil), in[:If_CluWordNum(n)]...)
	clear(out[:If_CluWordNum(n)])
	for bit := 0; bit < int(If_CluWordNum(n))*64; bit++ {
		if ifTruthBit(source, assignment(bit)) {
			ifTruthSet(out, bit)
		}
	}
}
func If_CluSwapAdjacent(out, in []uint64, v, n int32) {
	if v < 0 || v+1 >= n {
		panic("if: adjacent truth swap outside support")
	}
	cluTransform(out, in, n, func(a int) int {
		if a>>uint(v)&1 != a>>uint(v+1)&1 {
			a ^= (1 << uint(v)) | (1 << uint(v+1))
		}
		return a
	})
}
func If_CluChangePhase(table []uint64, n, v int32) {
	if v < 0 || v >= n {
		panic("if: truth phase outside support")
	}
	cluTransform(table, table, n, func(a int) int { return a ^ (1 << uint(v)) })
}
func If_CluSwapVars(table []uint64, n int32, v2p, p2v []int32, i, j int32) {
	if i == j {
		return
	}
	if i < 0 || j < 0 || i >= n || j >= n {
		panic("if: variable swap outside support")
	}
	cluTransform(table, table, n, func(a int) int {
		if a>>uint(i)&1 != a>>uint(j)&1 {
			a ^= (1 << uint(i)) | (1 << uint(j))
		}
		return a
	})
	if v2p != nil && p2v != nil {
		v2p[p2v[i]], v2p[p2v[j]] = j, i
		p2v[i], p2v[j] = p2v[j], p2v[i]
	}
}
func If_CluMoveVar2(table []uint64, n int32, v2p, p2v []int32, v, position int32) {
	If_CluSwapVars(table, n, v2p, p2v, v2p[v], position)
}
func If_CluMoveVar(table []uint64, n int32, v2p, p2v []int32, v, position int32) {
	if v < 0 || v >= n || position < 0 || position >= n {
		panic("if: moved variable outside support")
	}
	for v2p[v] < position {
		at := v2p[v]
		If_CluSwapVars(table, n, v2p, p2v, at, at+1)
	}
	for v2p[v] > position {
		at := v2p[v]
		If_CluSwapVars(table, n, v2p, p2v, at-1, at)
	}
}
func If_CluReverseOrder(table []uint64, n int32, v2p, p2v []int32, start int32) {
	for k := int32(0); k < (n-start)/2; k++ {
		If_CluSwapVars(table, n, v2p, p2v, start+k, n-1-k)
	}
}
func If_CluReverseOrder_old(table []uint64, n int32, v2p, p2v []int32, start int32) {
	If_CluReverseOrder(table, n, v2p, p2v, start)
}
func If_CluMoveGroupToMsb(table []uint64, n int32, v2p, p2v []int32, g *If_Grp_t) {
	for v := int32(0); v < g.nVars; v++ {
		If_CluMoveVar(table, n, v2p, p2v, g.pVars[g.nVars-1-v], n-1-v)
	}
}
func If_CluCofactors(table []uint64, n, v int32, zero, one []uint64) {
	if v < 0 || v >= n {
		panic("if: cofactor variable outside support")
	}
	cluTransform(zero, table, n, func(a int) int { return a &^ (1 << uint(v)) })
	cluTransform(one, table, n, func(a int) int { return a | (1 << uint(v)) })
}
func If_CluHasVar(table []uint64, n, v int32) int32 {
	for assignment := 0; assignment < 1<<uint(n); assignment++ {
		if assignment&(1<<uint(v)) == 0 && ifTruthBit(table, assignment) != ifTruthBit(table, assignment|(1<<uint(v))) {
			return 1
		}
	}
	return 0
}
func If_CluSupport(table []uint64, n int32) int32 {
	var support int32
	for v := int32(0); v < n; v++ {
		if If_CluHasVar(table, n, v) != 0 {
			support |= 1 << uint(v)
		}
	}
	return support
}
func If_CluSupportSize(table []uint64, n int32) int32 {
	return int32(bits.OnesCount32(uint32(If_CluSupport(table, n))))
}
func If_CluSuppIsMinBase(support int32) int32 {
	if support&(support+1) == 0 {
		return 1
	}
	return 0
}
func If_CluTruthShrink(table []uint64, n, nAll int32, support uint32) {
	positions := make([]int32, 0, nAll)
	for v := int32(0); v < nAll; v++ {
		if support&(1<<uint(v)) != 0 {
			positions = append(positions, v)
		}
	}
	if int(n) != len(positions) {
		panic("if: support count mismatch")
	}
	source := append([]uint64(nil), table[:If_CluWordNum(nAll)]...)
	cluTransform(table, source, nAll, func(a int) int {
		input := 0
		for i, v := range positions {
			input |= (a >> uint(i) & 1) << uint(v)
		}
		return input
	})
}
func If_CluMinimumBase(table []uint64, supp []int32, nAll int32, n *int32) int32 {
	mask := If_CluSupport(table, nAll)
	count := int32(0)
	for v := int32(0); v < nAll; v++ {
		if mask&(1<<uint(v)) != 0 {
			if supp != nil {
				supp[count] = supp[v]
			}
			count++
		}
	}
	if n != nil {
		*n = count
	}
	if If_CluSuppIsMinBase(mask) != 0 {
		return 0
	}
	If_CluTruthShrink(table, count, nAll, uint32(mask))
	return 1
}
func If_CluCountOnesInCofs(table []uint64, n int32, store []int32) {
	if int(2*n) > len(store) {
		panic("if: short cofactor counts")
	}
	clear(store[:2*n])
	for v := int32(0); v < n; v++ {
		for a := 0; a < 1<<uint(n); a++ {
			if ifTruthBit(table, a) {
				store[2*v+int32(a>>uint(v)&1)]++
			}
		}
		if n < 6 {
			store[2*v] *= 1 << uint(6-n)
			store[2*v+1] *= 1 << uint(6-n)
		}
	}
}
func If_CluSemiCanonicize(table []uint64, n int32, permutation []int32) uint32 {
	counts := make([]int32, 2*n)
	If_CluCountOnesInCofs(table, n, counts)
	var phase uint32
	for v := int32(0); v < n; v++ {
		permutation[v] = v
		if counts[2*v] > counts[2*v+1] {
			counts[2*v], counts[2*v+1] = counts[2*v+1], counts[2*v]
			phase |= 1 << uint(v)
			If_CluChangePhase(table, n, v)
		}
	}
	for changed := true; changed; {
		changed = false
		for v := int32(0); v < n-1; v++ {
			if counts[2*v] <= counts[2*(v+1)] {
				continue
			}
			counts[2*v], counts[2*(v+1)] = counts[2*(v+1)], counts[2*v]
			counts[2*v+1], counts[2*(v+1)+1] = counts[2*(v+1)+1], counts[2*v+1]
			permutation[v], permutation[v+1] = permutation[v+1], permutation[v]
			If_CluSwapAdjacent(table, table, v, n)
			changed = true
		}
	}
	return phase
}
func If_CluSemiCanonicizeVerify(table, original []uint64, n int32, permutation []int32, phase uint32) {
	for a := 0; a < 1<<uint(n); a++ {
		input := 0
		for i, v := range permutation[:n] {
			input |= ((a >> uint(i) & 1) ^ int(phase>>uint(v)&1)) << uint(v)
		}
		if ifTruthBit(table, a) != ifTruthBit(original, input) {
			panic("if: canonical transform verification failed")
		}
	}
}
func If_CluHashKey(table []uint64, nWords, size int32) int32 {
	if size <= 0 {
		panic("if: invalid decomposition hash size")
	}
	primes := [8]uint32{12582917, 25165843, 50331653, 100663319, 201326611, 402653189, 805306457, 1610612741}
	var value uint32
	if nWords < 4 {
		for i := 0; i < int(nWords)*8; i++ {
			value ^= primes[i%7] * uint32(table[i/8]>>uint(8*(i%8))&255)
		}
	} else {
		for i := 0; i < int(nWords)*2; i++ {
			value ^= primes[i%7] * uint32(table[i/2]>>uint(32*(i%2)))
		}
	}
	return int32(value % uint32(size))
}
func If_CluHashLookup(p *If_Man_t, table []uint64, t int32) *uint32 {
	if p == nil {
		return nil
	}
	if t < 0 || t > 1 {
		panic("if: invalid decomposition table")
	}
	words := If_CluWordNum(p.pPars.NLutSize)
	if p.pHashTable[t] == nil {
		objects := 0
		if p.vObjs != nil {
			objects = Vec_PtrSize(p.vObjs)
		}
		entries1 := int64(4) * int64(If_CluPrimeCudd(uint32(objects)*uint32(p.pPars.NCutsMax)))
		entries2 := int64(CLU_MEM_MAX) * (1 << 20) / int64(words) / 8
		p.nTableSize[t] = int32(If_CluPrimeCudd(uint32(min(entries1, entries2) / 2)))
		p.pHashTable[t] = make(map[uint32]any)
	}
	key := uint32(If_CluHashKey(table, words, p.nTableSize[t]))
	var head *If_Hte_t
	if raw := p.pHashTable[t][key]; raw != nil {
		head = raw.(*If_Hte_t)
	}
	for entry := head; entry != nil; entry = entry.pNext {
		if slices.Equal(entry.pTruth, table[:words]) {
			entry.Counter++
			return &entry.Group
		}
	}
	if p.nTableEntries[t] >= 2*p.nTableSize[t] {
		median := uint32(If_CluHashFindMedian(p, t))
		var kept []*If_Hte_t
		for bucket := int32(0); bucket < p.nTableSize[t]; bucket++ {
			if raw := p.pHashTable[t][uint32(bucket)]; raw != nil {
				for entry := raw.(*If_Hte_t); entry != nil; entry = entry.pNext {
					if entry.Counter > median {
						kept = append(kept, entry)
					}
				}
			}
		}
		p.pHashTable[t] = make(map[uint32]any)
		for _, entry := range kept {
			bucket := uint32(If_CluHashKey(entry.pTruth, words, p.nTableSize[t]))
			var first *If_Hte_t
			if raw := p.pHashTable[t][bucket]; raw != nil {
				first = raw.(*If_Hte_t)
			}
			if first == nil || entry.Counter >= first.Counter {
				entry.pNext = first
				p.pHashTable[t][bucket] = entry
			} else {
				previous := first
				for previous.pNext != nil && entry.Counter < previous.pNext.Counter {
					previous = previous.pNext
				}
				entry.pNext = previous.pNext
				previous.pNext = entry
			}
		}
		p.nTableEntries[t] = int32(len(kept))
	}
	entry := &If_Hte_t{Group: CLU_UNUSED, Counter: 1, pTruth: append([]uint64(nil), table[:words]...)}
	if raw := p.pHashTable[t][key]; raw == nil {
		p.pHashTable[t][key] = entry
	} else {
		tail := raw.(*If_Hte_t)
		for tail.pNext != nil {
			tail = tail.pNext
		}
		tail.pNext = entry
	}
	p.nTableEntries[t]++
	return &entry.Group
}
func If_CluHashPrintStats(p *If_Man_t, t int32) {
	for key, raw := range p.pHashTable[t] {
		count := 0
		for entry := raw.(*If_Hte_t); entry != nil; entry = entry.pNext {
			count++
		}
		if count >= 10 {
			fmt.Printf("%d=%d ", key, count)
		}
	}
}
func If_CluHashFindMedian(p *If_Man_t, t int32) int32 {
	var counts []uint32
	for _, raw := range p.pHashTable[t] {
		for entry := raw.(*If_Hte_t); entry != nil; entry = entry.pNext {
			counts = append(counts, entry.Counter)
		}
	}
	sort.Slice(counts, func(i, j int) bool { return counts[i] > counts[j] })
	if len(counts) == 0 {
		return 1
	}
	return int32(max(counts[len(counts)/2], 1))
}
func If_CluHashTableCheck(p *If_Man_t) {
	for t, table := range p.pHashTable {
		for _, raw := range table {
			for entry := raw.(*If_Hte_t); entry != nil; entry = entry.pNext {
				var group If_Grp_t
				If_CluUns2Grp(entry.Group, &group)
				if entry.Group == CLU_UNUSED {
					continue
				}
				if group.nVars > 6 {
					panic("if: invalid cached bound set")
				}
				if t == 1 && p.pPars.NLutSize == 13 {
					found := If_CluCheck3(nil, entry.pTruth, 13, 5, 5, 5, nil, nil, nil, nil, nil).nVars > 0
					if found != (group.nVars > 0) {
						panic("if: cached three-LUT result disagrees with decomposition")
					}
				}
			}
		}
	}
}
func If_CluPrintGroup(g *If_Grp_t) {
	fmt.Printf("Group n=%d cofs=%d vars=%v\n", g.nVars, g.nMyu, g.pVars[:g.nVars])
}
func If_CluPrintConfig(n int32, g, r *If_Grp_t, bound uint64, free []uint64) {
	If_CluPrintGroup(g)
	If_CluPrintGroup(r)
	fmt.Printf("leaf=%016x root=%x\n", bound, free)
}

// No global elementary-table initialization is needed: composition computes
// elementary variables from assignment bits and has no mutable shared state.
func cluIdentity(n int32) ([]int32, []int32) {
	a, b := make([]int32, n+2), make([]int32, n+2)
	for i := range a {
		a[i], b[i] = int32(i), int32(i)
	}
	return a, b
}
func cluCofactorRows(table []uint64, n, bound, shift int32) ([][]uint64, []int) {
	free := n - bound
	if bound < 1 || bound >= n || bound > 7 {
		panic("if: invalid bound cofactor dimensions")
	}
	var unique [][]uint64
	classes := make([]int, 1<<uint(bound))
	for row := range classes {
		cof := make([]uint64, If_CluWordNum(free))
		for assignment := 0; assignment < 1<<uint(free); assignment++ {
			bit := int(shift) + (row << uint(free)) + assignment
			if ifTruthBit(table, bit) {
				ifTruthSet(cof, assignment)
			}
		}
		class := 0
		for class < len(unique) && !slices.Equal(unique[class], cof) {
			class++
		}
		if class == len(unique) {
			unique = append(unique, cof)
		}
		classes[row] = class
		if len(unique) == 5 {
			return unique, classes
		}
	}
	return unique, classes
}
func If_CluCountCofs(table []uint64, n, bound, shift int32, cofs [][]uint64) int32 {
	unique, classes := cluCofactorRows(table, n, bound, shift)
	if len(unique) <= 2 && cofs != nil {
		if bound > 6 || len(cofs) < 3 {
			panic("if: invalid cofactor output storage")
		}
		copy(cofs[0], unique[0])
		copy(cofs[1], unique[min(1, len(unique)-1)])
		var selection uint64
		for i, class := range classes {
			if class != 0 {
				selection |= 1 << uint(i)
			}
		}
		cofs[2][0] = selection
	}
	return int32(len(unique))
}
func If_CluCountCofs4(table []uint64, n, bound int32, cofs [][]uint64) int32 {
	unique, classes := cluCofactorRows(table, n, bound, 0)
	if len(unique) < 3 || len(unique) > 4 || bound > 6 || len(cofs) < 6 {
		panic("if: two-output decomposition needs three or four cofactors")
	}
	for i := 0; i < 4; i++ {
		copy(cofs[i], unique[min(i, len(unique)-1)])
	}
	for i, class := range classes {
		if class&1 != 0 {
			cofs[4][0] |= 1 << uint(i)
		}
		if class&2 != 0 {
			cofs[5][0] |= 1 << uint(i)
		}
	}
	return int32(len(unique))
}
func If_CluDetectSpecialCaseCofs(table []uint64, n, v int32) int32 {
	size := 1 << uint(n)
	states := [6]bool{true, true, true, true, true, true}
	for assignment := 0; assignment < size; assignment++ {
		if assignment&(1<<uint(v)) != 0 {
			continue
		}
		a, b := ifTruthBit(table, assignment), ifTruthBit(table, assignment|(1<<uint(v)))
		states[0] = states[0] && !a
		states[1] = states[1] && a
		states[2] = states[2] && !b
		states[3] = states[3] && b
		states[4] = states[4] && (a != b)
		states[5] = states[5] && (a == b)
	}
	for i := 0; i < 5; i++ {
		if states[i] {
			return int32(i)
		}
	}
	return -1
}
func If_CluDecUsingCofs(table []uint64, n, leaf int32) If_Grp_t {
	var group If_Grp_t
	if n <= leaf {
		return group
	}
	work := append([]uint64(nil), table[:If_CluWordNum(n)]...)
	v2p, p2v := cluIdentity(n)
	needed := n - leaf
	for removed := int32(0); removed < needed; removed++ {
		found := false
		for v := n - 1; v >= 0; v-- {
			state := If_CluDetectSpecialCaseCofs(work, n, v)
			if state < 0 {
				continue
			}
			original := p2v[v]
			for v2p[original] < n-1 {
				at := v2p[original]
				v2p[p2v[at]], v2p[p2v[at+1]] = at+1, at
				p2v[at], p2v[at+1] = p2v[at+1], p2v[at]
			}
			a, b := cluIdentity(n)
			If_CluMoveVar(work, n, a, b, v, n-1)
			bit := 0
			if state == 0 || state == 1 {
				bit = 1
			}
			source := append([]uint64(nil), work...)
			cluTransform(work, source, n, func(assignment int) int { return assignment&^(1<<uint(n-1)) | bit<<uint(n-1) })
			n--
			found = true
			break
		}
		if !found {
			return group
		}
	}
	group.nVars, group.nMyu = leaf, 2
	copy(group.pVars[:], p2v[:leaf])
	return group
}
func If_CluDeriveDisjoint(table []uint64, n int32, v2p, p2v []int32, g, r *If_Grp_t) uint64 {
	unique, classes := cluCofactorRows(table, n, g.nVars, 0)
	if len(unique) > 2 {
		panic("if: group is not disjointly decomposable")
	}
	free := n - g.nVars
	var selector uint64
	for i, class := range classes {
		if class != 0 {
			selector |= 1 << uint(i)
		}
	}
	out := make([]uint64, If_CluWordNum(free+1))
	for assignment := 0; assignment < 1<<uint(free+1); assignment++ {
		cof := unique[min(assignment>>uint(free), len(unique)-1)]
		if ifTruthBit(cof, assignment&((1<<uint(free))-1)) {
			ifTruthSet(out, assignment)
		}
	}
	copy(table, out)
	if r != nil {
		*r = If_Grp_t{nVars: free + 1}
		copy(r.pVars[:], p2v[:free])
		r.pVars[free] = n
	}
	return selector
}
func If_CluDeriveDisjoint4(table []uint64, n int32, v2p, p2v []int32, g, r *If_Grp_t, truth0, truth1 *uint64) {
	unique, classes := cluCofactorRows(table, n, g.nVars, 0)
	if len(unique) < 3 || len(unique) > 4 {
		panic("if: invalid four-cofactor decomposition")
	}
	free := n - g.nVars
	out := make([]uint64, If_CluWordNum(free+2))
	for assignment := 0; assignment < 1<<uint(free+2); assignment++ {
		cof := unique[min(assignment>>uint(free), len(unique)-1)]
		if ifTruthBit(cof, assignment&((1<<uint(free))-1)) {
			ifTruthSet(out, assignment)
		}
	}
	if free+2 < 6 {
		out[0] = If_CluAdjust(out[0], free+2)
	}
	copy(table, out)
	*r = If_Grp_t{nVars: free + 2}
	copy(r.pVars[:], p2v[:free])
	r.pVars[free], r.pVars[free+1] = n, n+1
	var first, second uint64
	for i, class := range classes {
		if class&1 != 0 {
			first |= 1 << uint(i)
		}
		if class&2 != 0 {
			second |= 1 << uint(i)
		}
	}
	*truth0, *truth1 = If_CluAdjust(first, g.nVars), If_CluAdjust(second, g.nVars)
}
func If_CluDeriveNonDisjoint(table []uint64, n int32, v2p, p2v []int32, g, r *If_Grp_t) uint64 {
	free := n - g.nVars
	size := If_CluWordNum(n)
	a, b := make([]uint64, size), make([]uint64, size)
	If_CluCofactors(table, n, n-1, a, b)
	g.nVars--
	truth0 := If_CluDeriveDisjoint(a, n-1, v2p, p2v, g, nil)
	truth1 := If_CluDeriveDisjoint(b, n-1, v2p, p2v, g, nil)
	g.nVars++
	boundTruth := truth0 | truth1<<uint(1<<uint(g.nVars-1))
	out := make([]uint64, If_CluWordNum(free+2))
	for assignment := 0; assignment < 1<<uint(free+2); assignment++ {
		cof := a
		if assignment&(1<<uint(free+1)) != 0 {
			cof = b
		}
		if ifTruthBit(cof, assignment&((1<<uint(free+1))-1)) {
			ifTruthSet(out, assignment)
		}
	}
	copy(table, out)
	*r = If_Grp_t{nVars: free + 2}
	copy(r.pVars[:], p2v[:free])
	r.pVars[free], r.pVars[free+1] = n, g.pVars[g.nVars-1]
	return boundTruth
}
func If_CluCheckNonDisjointGroup(table []uint64, n int32, v2p, p2v []int32, g *If_Grp_t) int32 {
	if g.nMyu != 3 && g.nMyu != 4 {
		return 0
	}
	for v := int32(0); v < g.nVars; v++ {
		a, b := make([]uint64, If_CluWordNum(n)), make([]uint64, If_CluWordNum(n))
		If_CluCofactors(table, n, v2p[g.pVars[v]], a, b)
		if If_CluCountCofs(a, n, g.nVars, 0, nil) > 2 || If_CluCountCofs(b, n, g.nVars, 0, nil) > 2 {
			continue
		}
		If_CluMoveVar(table, n, v2p, p2v, g.pVars[v], n-1)
		copy(g.pVars[:g.nVars], p2v[n-g.nVars:n])
		return 1
	}
	return 0
}
func If_CluFindGroup(table []uint64, n, start, stop int32, v2p, p2v []int32, bound, disjoint int32) If_Grp_t {
	g := If_Grp_t{nVars: bound, nMyu: If_CluCountCofs(table, n, bound, 0, nil)}
	copy(g.pVars[:], p2v[n-bound:n])
	if g.nMyu == 2 || disjoint == 0 && If_CluCheckNonDisjointGroup(table, n, v2p, p2v, &g) != 0 {
		return g
	}
	if n == bound+start {
		g.nVars = 0
		return g
	}
	for round := 0; round < 2; round++ {
		if bound < n-1 {
			best := p2v[n-1-bound]
			bestCount := If_CluCountCofs(table, n, bound+1, 0, nil)
			for v := n - 2 - bound; v >= start; v-- {
				If_CluMoveVar2(table, n, v2p, p2v, p2v[v], n-1-bound)
				count := If_CluCountCofs(table, n, bound+1, 0, nil)
				if bestCount >= count {
					bestCount, best = count, p2v[n-1-bound]
				}
			}
			If_CluMoveVar2(table, n, v2p, p2v, best, n-1-bound)
		}
		best := p2v[n-1-bound]
		bestCount := If_CluCountCofs(table, n, bound, 0, nil)
		for v := n - bound; v < n-stop; v++ {
			If_CluMoveVar2(table, n, v2p, p2v, p2v[v], n-1-bound)
			count := If_CluCountCofs(table, n, bound, 0, nil)
			if bestCount >= count {
				bestCount, best = count, p2v[n-1-bound]
			}
		}
		If_CluMoveVar2(table, n, v2p, p2v, best, n-1-bound)
		count := If_CluCountCofs(table, n, bound, 0, nil)
		if g.nMyu >= count {
			g.nMyu = count
			g.nVars = bound
			copy(g.pVars[:], p2v[n-bound:n])
		}
		if g.nMyu == 2 || disjoint == 0 && If_CluCheckNonDisjointGroup(table, n, v2p, p2v, &g) != 0 {
			return g
		}
	}
	g.nVars = 0
	return g
}
func If_CluCheck(p *If_Man_t, truth []uint64, n, start, stop, leaf, root int32, result *If_Grp_t, rootTruth, leafTruth *uint64, leftover []uint64, hashing int32) If_Grp_t {
	var group If_Grp_t
	if n < 0 || n > 16 || leaf < 2 || leaf > 6 || n <= leaf || n > leaf+root-1 {
		return group
	}
	if result != nil {
		*result = If_Grp_t{}
		if rootTruth != nil {
			*rootTruth = 0
		}
		if leafTruth != nil {
			*leafTruth = 0
		}
	}
	work := make([]uint64, CLU_WRD_MAX)
	copy(work, truth[:If_CluWordNum(n)])
	support := If_CluSupport(work, n)
	if support == 0 || If_CluSuppIsMinBase(support) == 0 {
		return group
	}
	v2p, p2v := cluIdentity(n)
	var cached *uint32
	if p != nil && hashing != 0 {
		cached = If_CluHashLookup(p, truth, 0)
		if *cached != CLU_UNUSED {
			If_CluUns2Grp(*cached, &group)
		}
	}
	if stop != 0 {
		If_CluMoveVar(work, n, v2p, p2v, 0, n-1)
	}
	if group.nVars == 0 {
		if start == 0 {
			group = If_CluDecUsingCofs(truth, n, leaf)
		}
		if group.nVars == 0 {
			group = If_CluFindGroup(work, n, start, stop, v2p, p2v, leaf, boolInt(leaf+root == n+1))
			if group.nVars == 0 {
				if n < leaf+root-2 {
					group = If_CluFindGroup(work, n, start, stop, v2p, p2v, leaf-1, boolInt(leaf+root == n+1))
				}
				if leaf > 4 && n < leaf+root-3 {
					group = If_CluFindGroup(work, n, start, stop, v2p, p2v, leaf-2, boolInt(leaf+root == n+1))
				}
				if group.nVars == 0 {
					If_CluReverseOrder(work, n, v2p, p2v, start)
					group = If_CluFindGroup(work, n, start, stop, v2p, p2v, leaf, boolInt(leaf+root == n+1))
				}
			}
		}
	}
	if cached != nil {
		*cached = If_CluGrp2Uns(&group)
	}
	if group.nVars == 0 || result == nil {
		return group
	}
	If_CluMoveGroupToMsb(work, n, v2p, p2v, &group)
	var bound uint64
	var remaining If_Grp_t
	if group.nMyu == 2 {
		bound = If_CluDeriveDisjoint(work, n, v2p, p2v, &group, &remaining)
	} else {
		bound = If_CluDeriveNonDisjoint(work, n, v2p, p2v, &group, &remaining)
	}
	bound = If_CluAdjust(bound, group.nVars)
	if remaining.nVars < 6 {
		work[0] = If_CluAdjust(work[0], remaining.nVars)
	}
	newPos := remaining.nVars - 1
	if group.nMyu > 2 {
		newPos--
	}
	a, b := cluIdentity(remaining.nVars)
	If_CluMoveVar(work, remaining.nVars, a, b, newPos, start)
	for i := newPos; i > start; i-- {
		remaining.pVars[i] = remaining.pVars[i-1]
	}
	remaining.pVars[start] = n
	if leftover != nil {
		copy(leftover, work[:If_CluWordNum(remaining.nVars)])
		maximum := n
		if p != nil {
			maximum = p.pPars.NLutSize
		}
		If_CluAdjustBig(leftover, remaining.nVars, maximum)
	}
	*result = remaining
	if rootTruth != nil {
		*rootTruth = work[0]
	}
	if leafTruth != nil {
		*leafTruth = bound
	}
	return group
}
func boolInt(v bool) int32 {
	if v {
		return 1
	}
	return 0
}
func If_CluCheck3(p *If_Man_t, truth []uint64, n, leaf, leaf2, root int32, result, second *If_Grp_t, rootTruth, firstTruth, secondTruth *uint64) If_Grp_t {
	leftover := make([]uint64, CLU_WRD_MAX)
	var remaining If_Grp_t
	var f0, f1, f2 uint64
	first := If_CluCheck(p, truth, n, 0, 0, leaf, root+leaf2-1, &remaining, &f0, &f1, leftover, 0)
	if first.nVars == 0 {
		if (first.nMyu == 3 || first.nMyu == 4) && leaf == leaf2 && n-leaf+2 <= root {
			first.nVars = leaf
			copy(leftover, truth[:If_CluWordNum(n)])
			v2p, p2v := cluIdentity(n)
			If_CluMoveGroupToMsb(leftover, n, v2p, p2v, &first)
			If_CluDeriveDisjoint4(leftover, n, v2p, p2v, &first, &remaining, &f1, &f2)
			a, b := cluIdentity(remaining.nVars)
			If_CluMoveVar(leftover, remaining.nVars, a, b, remaining.nVars-2, 0)
			If_CluMoveVar(leftover, remaining.nVars, a, b, remaining.nVars-1, 1)
			old0, old1 := remaining.pVars[remaining.nVars-2], remaining.pVars[remaining.nVars-1]
			for i := remaining.nVars - 1; i > 1; i-- {
				remaining.pVars[i] = remaining.pVars[i-2]
			}
			remaining.pVars[0], remaining.pVars[1] = old0, old1
			f0 = leftover[0]
			if result != nil {
				*result = remaining
			}
			if second != nil {
				*second = first
			}
			if rootTruth != nil {
				*rootTruth = f0
			}
			if firstTruth != nil {
				*firstTruth = f1
			}
			if secondTruth != nil {
				*secondTruth = f2
			}
			return first
		}
		return first
	}
	if remaining.nVars <= root {
		if result != nil {
			*result = remaining
		}
		if second != nil {
			*second = If_Grp_t{}
		}
		if rootTruth != nil {
			*rootTruth = f0
		}
		if firstTruth != nil {
			*firstTruth = f1
		}
		if secondTruth != nil {
			*secondTruth = 0
		}
		return first
	}
	start, stop := int32(0), int32(0)
	if p != nil {
		switch p.pPars.NStructType {
		case 0:
		case 1:
			start = 1
		case 2:
			stop = 1
		default:
			panic("if: invalid three-LUT structure restriction")
		}
	}
	var final If_Grp_t
	next := If_CluCheck(p, leftover, remaining.nVars, start, stop, leaf2, root, &final, &f0, &f2, nil, 0)
	if next.nVars == 0 {
		return next
	}
	for i := int32(0); i < next.nVars; i++ {
		next.pVars[i] = remaining.pVars[next.pVars[i]]
	}
	for i := int32(0); i < final.nVars; i++ {
		if final.pVars[i] == remaining.nVars {
			final.pVars[i] = n + 1
		} else {
			final.pVars[i] = remaining.pVars[final.pVars[i]]
		}
	}
	if result != nil {
		*result = final
	}
	if second != nil {
		*second = next
	}
	if rootTruth != nil {
		*rootTruth = f0
	}
	if firstTruth != nil {
		*firstTruth = f1
	}
	if secondTruth != nil {
		*secondTruth = f2
	}
	return first
}
func If_CluInitTruthTables(n int32) [][]uint64 {
	tables := make([][]uint64, n+2)
	for v := range tables {
		tables[v] = make([]uint64, If_CluWordNum(n))
		if v >= int(n) {
			continue
		}
		for a := 0; a < len(tables[v])*64; a++ {
			if a>>uint(v)&1 != 0 {
				ifTruthSet(tables[v], a)
			}
		}
	}
	return tables
}
func If_CluComposeLut(n int32, g *If_Grp_t, truth []uint64, functions [][]uint64, out []uint64) {
	clear(out[:If_CluWordNum(n)])
	for a := 0; a < int(If_CluWordNum(n))*64; a++ {
		index := 0
		for v := int32(0); v < g.nVars; v++ {
			if ifTruthBit(functions[g.pVars[v]], a) {
				index |= 1 << uint(v)
			}
		}
		if ifTruthBit(truth, index) {
			ifTruthSet(out, a)
		}
	}
}
func If_CluVerify(truth []uint64, n int32, g, r *If_Grp_t, bound uint64, free []uint64) {
	functions := If_CluInitTruthTables(n)
	If_CluComposeLut(n, g, []uint64{bound}, functions, functions[n])
	out := make([]uint64, If_CluWordNum(n))
	If_CluComposeLut(n, r, free, functions, out)
	if If_CluEqual(truth, out, n) == 0 {
		panic("if: two-LUT decomposition verification failed")
	}
}
func If_CluVerify3(truth []uint64, n int32, g, g2, r *If_Grp_t, bound, bound2, free uint64) {
	functions := If_CluInitTruthTables(n)
	If_CluComposeLut(n, g, []uint64{bound}, functions, functions[n])
	if g2.nVars != 0 {
		If_CluComposeLut(n, g2, []uint64{bound2}, functions, functions[n+1])
	}
	out := make([]uint64, If_CluWordNum(n))
	If_CluComposeLut(n, r, []uint64{free}, functions, out)
	if If_CluEqual(truth, out, n) == 0 {
		panic("if: three-LUT decomposition verification failed")
	}
}
func If_CluCheckGroup(truth []uint64, n int32, g *If_Grp_t) {
	work := append([]uint64(nil), truth[:If_CluWordNum(n)]...)
	a, b := cluIdentity(n)
	If_CluMoveGroupToMsb(work, n, a, b, g)
	if If_CluCountCofs(work, n, g.nVars, 0, nil) != g.nMyu {
		panic("if: decomposition cofactor count mismatch")
	}
	if g.nMyu > 2 {
		half := 1 << uint(n-1)
		if If_CluCountCofs(work, n-1, g.nVars-1, 0, nil) > 2 || If_CluCountCofs(work, n-1, g.nVars-1, int32(half), nil) > 2 {
			panic("if: incompatible shared-variable group")
		}
	}
}
func If_CluCheckPerm(truth, work []uint64, n int32, v2p, p2v []int32) {
	for i := int32(0); i < n; i++ {
		If_CluMoveVar(work, n, v2p, p2v, i, i)
	}
	if If_CluEqual(truth, work, n) == 0 {
		panic("if: decomposition permutation mismatch")
	}
}
func cluCofactor6(truth uint64, v int32, one bool) uint64 {
	var out uint64
	for a := 0; a < 64; a++ {
		index := a &^ (1 << uint(v))
		if one {
			index |= 1 << uint(v)
		}
		if truth>>uint(index)&1 != 0 {
			out |= 1 << uint(a)
		}
	}
	return out
}
func cluCheckDecIn(truth uint64, n int32, firstOnly, unate bool) int32 {
	for v := int32(0); v < n; v++ {
		if firstOnly && v > 0 {
			break
		}
		a, b := cluCofactor6(truth, v, false), cluCofactor6(truth, v, true)
		for u := v + 1; u < n; u++ {
			cofs := [4]uint64{cluCofactor6(a, u, false), cluCofactor6(a, u, true), cluCofactor6(b, u, false), cluCofactor6(b, u, true)}
			if cofs[0] == cofs[1] && cofs[0] == cofs[2] || cofs[0] == cofs[2] && cofs[0] == cofs[3] || !unate && (cofs[0] == cofs[1] && cofs[0] == cofs[3] || cofs[1] == cofs[2] && cofs[1] == cofs[3]) {
				return 1
			}
		}
	}
	return 0
}
func If_CluCheckDecInAny(t uint64, n int32) int32 { return cluCheckDecIn(t, n, false, false) }
func If_CluCheckDecIn(t uint64, n int32) int32    { return cluCheckDecIn(t, n, true, false) }
func If_CluCheckDecInU(t uint64, n int32) int32   { return cluCheckDecIn(t, n, true, true) }
func If_CluCheckDecOut(t uint64, n int32) int32 {
	for v := int32(0); v < n; v++ {
		a, b := cluCofactor6(t, v, false), cluCofactor6(t, v, true)
		if a == 0 || a == ^uint64(0) || b == 0 || b == ^uint64(0) {
			return 1
		}
	}
	return 0
}
func If_CluCheckDecOutU(t uint64, n int32) int32 {
	for v := int32(0); v < n; v++ {
		a := cluCofactor6(t, v, false)
		if a == 0 || a == ^uint64(0) {
			return 1
		}
	}
	return 0
}
func cluPackGroup(g If_Grp_t, out []byte) {
	if len(out) < 18 {
		panic("if: short LUT group descriptor")
	}
	out[0], out[1] = byte(g.nVars), byte(g.nMyu)
	for i, v := range g.pVars {
		out[2+i] = byte(v)
	}
}
func If_CluCheckExt(p *If_Man_t, truth []uint64, n, leaf, root int32, lut0, lut1 []byte, rootTruth, leafTruth *uint64) int32 {
	var r If_Grp_t
	g := If_CluCheck(p, truth, n, 0, 0, leaf, root, &r, rootTruth, leafTruth, nil, 0)
	cluPackGroup(r, lut0)
	cluPackGroup(g, lut1)
	return boolInt(g.nVars > 0)
}
func If_CluCheckExt3(p *If_Man_t, truth []uint64, n, leaf, leaf2, root int32, lut0, lut1, lut2 []byte, rootTruth, leafTruth, leafTruth2 *uint64) int32 {
	var r, g2 If_Grp_t
	g := If_CluCheck3(p, truth, n, leaf, leaf2, root, &r, &g2, rootTruth, leafTruth, leafTruth2)
	cluPackGroup(r, lut0)
	cluPackGroup(g, lut1)
	cluPackGroup(g2, lut2)
	return boolInt(g.nVars > 0)
}
func If_CluDelayMax(g *If_Grp_t, delays []float32) float32 {
	var delay float32
	for _, v := range g.pVars[:g.nVars] {
		delay = max(delay, delays[v])
	}
	return delay
}
func cluWords32(truth []uint32, n int32) []uint64 {
	words := If_CluWordNum(n)
	out := make([]uint64, words)
	if len(truth) < int(2*words) {
		if n <= 5 && len(truth) >= 1 {
			out[0] = uint64(truth[0]) | uint64(truth[0])<<32
			return out
		}
		panic("if: short 32-bit truth table")
	}
	for i := range out {
		out[i] = uint64(truth[2*i]) | uint64(truth[2*i+1])<<32
	}
	return out
}
func If_CutPerformCheck45(p *If_Man_t, truth []uint32, n, leaves int32, structure string) int32 {
	var r If_Grp_t
	var f0, f1 uint64
	g := If_CluCheck(p, cluWords32(truth, n), leaves, 0, 0, 5, 4, &r, &f0, &f1, nil, 0)
	if g.nVars == 0 {
		return 0
	}
	f1 = If_CluAdjust(f1, g.nVars)
	return boolInt(g.nVars < 5 || p.pPars.FEnableCheck75 != 0 && If_CluCheckDecOut(f1, 5) != 0 || p.pPars.FEnableCheck75u != 0 && If_CluCheckDecOutU(f1, 5) != 0)
}
func If_CutPerformCheck54(p *If_Man_t, truth []uint32, n, leaves int32, structure string) int32 {
	var r If_Grp_t
	var f0, f1 uint64
	g := If_CluCheck(p, cluWords32(truth, n), leaves, 0, 0, 4, 5, &r, &f0, &f1, nil, 0)
	if g.nVars == 0 {
		return 0
	}
	f0 = If_CluAdjust(f0, r.nVars)
	return boolInt(r.nVars < 5 || p.pPars.FEnableCheck75 != 0 && If_CluCheckDecIn(f0, 5) != 0 || p.pPars.FEnableCheck75u != 0 && If_CluCheckDecInU(f0, 5) != 0)
}
func cluStructure(s string) ([]int32, bool) {
	if len(s) != 2 && len(s) != 3 {
		return nil, false
	}
	sizes := make([]int32, len(s))
	for i, ch := range []byte(s) {
		if ch < '3' || ch > '6' {
			return nil, false
		}
		sizes[i] = int32(ch - '0')
	}
	return sizes, true
}
func If_CutPerformCheck16(p *If_Man_t, truth []uint32, n, leaves int32, structure string) int32 {
	sizes, ok := cluStructure(structure)
	if !ok || leaves < 0 || leaves > 16 || n < leaves || n > 16 {
		return 0
	}
	work := make([]uint64, CLU_WRD_MAX)
	copy(work, cluWords32(truth, n))
	maximum := n
	if p != nil {
		maximum = p.pPars.NLutSize
	}
	If_CluAdjustBig(work, leaves, maximum)
	if p == nil || p.pPars.FCutMin == 0 {
		If_CluMinimumBase(work, nil, leaves, &leaves)
	}
	capacity := sizes[0] + sizes[len(sizes)-1] - 1
	largest := max(sizes[0], sizes[len(sizes)-1])
	if len(sizes) == 3 {
		capacity += sizes[1] - 1
		largest = max(largest, sizes[1])
	}
	if leaves > capacity {
		return 0
	}
	if leaves <= largest {
		return 1
	}
	if len(sizes) == 2 {
		return boolInt(If_CluCheck(p, work, leaves, 0, 0, sizes[0], sizes[1], nil, nil, nil, nil, 1).nVars > 0)
	}
	return boolInt(If_CluCheck3(p, work, leaves, sizes[0], sizes[1], sizes[2], nil, nil, nil, nil, nil).nVars > 0)
}
func If_CutDelayLutStruct(p *If_Man_t, c *If_Cut_t, structure string, wireDelay float32) float32 {
	sizes, ok := cluStructure(structure)
	if !ok || len(sizes) != 2 || int32(c.nLeaves) > sizes[0]+sizes[1]-1 {
		return float32(IF_INFINITY)
	}
	n := int32(c.nLeaves)
	delays := make([]float32, n+2)
	for i, id := range ifLeaves(c) {
		delays[i] = If_ManObj(p, id).CutBest.Delay
	}
	if n <= max(sizes[0], sizes[1]) {
		c.permutation = make([]byte, n)
		g := If_Grp_t{nVars: n}
		for i := int32(0); i < n; i++ {
			g.pVars[i] = i
			c.permutation[i] = 1
		}
		return 1 + If_CluDelayMax(&g, delays)
	}
	If_CutTruthW(p, c)
	g := If_CluCheck(p, p.puTempW, n, 0, 0, sizes[0], sizes[1], nil, nil, nil, nil, 1)
	if g.nVars == 0 {
		return float32(IF_INFINITY)
	}
	increment := wireDelay
	if increment == 0 {
		increment = 1
	}
	delays[n] = If_CluDelayMax(&g, delays) + increment
	used := make([]bool, n)
	for _, id := range g.pVars[:g.nVars] {
		used[id] = true
	}
	if g.nMyu > 2 {
		used[g.pVars[g.nVars-1]] = false
	}
	r := If_Grp_t{}
	for i, v := range used {
		if !v {
			r.pVars[r.nVars] = int32(i)
			r.nVars++
		}
	}
	r.pVars[r.nVars] = n
	r.nVars++
	c.Cost = 2
	return 1 + If_CluDelayMax(&r, delays)
}
func If_CluTest() {
	truth := []uint64{0x6dd9926d962d6996}
	var r, g2 If_Grp_t
	var f0, f1, f2 uint64
	g := If_CluCheck3(nil, truth, 6, 4, 4, 4, &r, &g2, &f0, &f1, &f2)
	if g.nVars != 0 {
		If_CluVerify3(truth, 6, &g, &g2, &r, f1, f2, f0)
	}
}
