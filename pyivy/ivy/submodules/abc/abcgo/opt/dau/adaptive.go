package dau

// Adaptive tied-group canonicalization from the second half of dauCanon.c.
// Logical variable names, current positions and symmetry chains remain
// separate, as in the source; all scratch truth tables are call-local.
import (
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/opt/dau/canon"
	"math/bits"
)

type TiedGroup struct{ iStart, nGVars int }
type Abc_TgMan_t struct {
	pTruth                             []uint64
	original                           []uint64
	nVars, nGVars, nGroups             int
	uPhase                             uint32
	fPhased                            bool
	pPerm, pPermT, pPermTRev, pPermDir [16]int
	pGroup                             [16]TiedGroup
	symPhase                           [16]int
	symLink                            [17]int
	nAlgorithm                         int
	pFGrps                             [16]int
	vPhase                             *vec.Vec_Int_t
}

func Abc_TgInitMan(p *Abc_TgMan_t, t []uint64, nVars, nAlg int, phase *vec.Vec_Int_t) {
	*p = Abc_TgMan_t{pTruth: t[:dauCountWords(int32(nVars))], original: append([]uint64(nil), t[:dauCountWords(int32(nVars))]...), nVars: nVars, nGVars: nVars, nGroups: 1, nAlgorithm: nAlg, vPhase: phase}
	p.pGroup[0] = TiedGroup{0, nVars}
	for i := 0; i < nVars; i++ {
		p.pPerm[i] = i
		p.pPermT[i] = i
		p.pPermTRev[i] = i
		p.symPhase[i] = 1
		p.symLink[i] = -1
	}
	p.symLink[nVars] = -1
	if phase != nil {
		vec.Vec_IntClear(phase)
	}
}
func Abc_TgSplitGroup(p *Abc_TgMan_t, index int, coef []int32) int {
	g := p.pGroup[index]
	vars := p.pPerm[g.iStart : g.iStart+g.nGVars]
	for i := 1; i < g.nGVars; i++ {
		value, name := coef[i], vars[i]
		j := i
		for j > 0 && coef[j-1] > value {
			coef[j], vars[j] = coef[j-1], vars[j-1]
			j--
		}
		coef[j], vars[j] = value, name
	}
	splits := 0
	for i := 1; i < g.nGVars; i++ {
		if coef[i] != coef[i-1] {
			splits++
		}
	}
	if splits == 0 {
		return 0
	}
	copy(p.pGroup[index+splits+1:], p.pGroup[index+1:p.nGroups])
	j := 1
	for i := 1; i < g.nGVars; i++ {
		if coef[i] == coef[i-1] {
			continue
		}
		p.pGroup[index+j].iStart = g.iStart + i
		p.pGroup[index+j-1].nGVars = p.pGroup[index+j].iStart - p.pGroup[index+j-1].iStart
		j++
	}
	p.pGroup[index+splits].nGVars = g.iStart + g.nGVars - p.pGroup[index+splits].iStart
	p.nGroups += splits
	return splits
}
func Abc_TgManCopy(dst *Abc_TgMan_t, truth []uint64, src *Abc_TgMan_t) {
	*dst = *src
	copy(truth, src.pTruth)
	dst.pTruth = truth[:len(src.pTruth)]
}
func dauPermute(t []uint64, newToOld []int) {
	var current [16]int
	for i := range newToOld {
		current[i] = i
	}
	for i, want := range newToOld {
		if current[i] == want {
			continue
		}
		j := i + 1
		for current[j] != want {
			j++
		}
		canon.SwapVariables(t, i, j)
		current[i], current[j] = current[j], current[i]
	}
}
func Abc_TgCannonVerify(p *Abc_TgMan_t) bool {
	t := append([]uint64(nil), p.pTruth...)
	for i := 0; i < p.nVars; i++ {
		if p.uPhase&(1<<uint(i)) != 0 {
			canon.FlipVariable(t, i)
		}
	}
	dauPermute(t, p.pPermTRev[:p.nVars])
	for i, w := range t {
		if p.uPhase&(1<<uint(p.nVars)) != 0 {
			w = ^w
		}
		if w != p.original[i] {
			return false
		}
	}
	return true
}
func CheckConfig(p *Abc_TgMan_t) {
	var dest [16]byte
	Abc_TgExpendSymmetry(p, dest[:])
	for i := 0; i < p.nVars; i++ {
		if int(dest[i]) != p.pPermT[i] || p.pPermTRev[p.pPermT[i]] != i {
			panic("inconsistent DAU symmetry permutation")
		}
	}
}
func Abc_TgFlipVar(p *Abc_TgMan_t, v int) {
	pos := p.pPermTRev[v]
	canon.FlipVariable(p.pTruth, pos)
	p.uPhase ^= 1 << uint(pos)
}
func Abc_TgFlipSymGroupByVar(p *Abc_TgMan_t, v int) {
	for v >= 0 {
		if p.symPhase[v] != 0 {
			Abc_TgFlipVar(p, v)
		}
		v = p.symLink[v]
	}
}
func Abc_TgFlipSymGroup(p *Abc_TgMan_t, index int) { Abc_TgFlipSymGroupByVar(p, p.pPerm[index]) }
func Abc_TgClearSymGroupPhase(p *Abc_TgMan_t, v int) {
	for v >= 0 {
		p.symPhase[v] = 0
		v = p.symLink[v]
	}
}
func Abc_TgImplementPerm(p *Abc_TgMan_t, dest []byte) {
	var wanted [16]int
	phase := p.uPhase & (1 << uint(p.nVars))
	for i := 0; i < p.nVars; i++ {
		wanted[i] = p.pPermTRev[dest[i]]
		if p.uPhase&(1<<uint(wanted[i])) != 0 {
			phase |= 1 << uint(i)
		}
	}
	dauPermute(p.pTruth, wanted[:p.nVars])
	for i := 0; i < p.nVars; i++ {
		p.pPermT[i] = int(dest[i])
		p.pPermTRev[dest[i]] = i
	}
	p.uPhase = phase
}
func Abc_TgSwapAdjacentSymGroups(p *Abc_TgMan_t, index int) {
	a, b := p.pPerm[index], p.pPerm[index+1]
	p.pPerm[index], p.pPerm[index+1] = b, a
	p.pPermDir[index], p.pPermDir[index+1] = p.pPermDir[index+1], p.pPermDir[index]
	if p.symLink[a] >= 0 || p.symLink[b] >= 0 {
		var dest [16]byte
		Abc_TgExpendSymmetry(p, dest[:])
		Abc_TgImplementPerm(p, dest[:])
		return
	}
	pos := p.pPermTRev[a]
	canon.SwapAdjacent(p.pTruth, pos)
	p.pPermT[pos], p.pPermT[pos+1] = b, a
	p.pPermTRev[a], p.pPermTRev[b] = pos+1, pos
	if p.uPhase>>uint(pos)&1 != p.uPhase>>uint(pos+1)&1 {
		p.uPhase ^= 3 << uint(pos)
	}
}
func Abc_TtIsSymmetric(truth []uint64, nVars, i, j, fPhase int) int {
	t := append([]uint64(nil), truth[:dauCountWords(int32(nVars))]...)
	canon.SwapVariables(t, i, j)
	result := 0
	if canon.CompareReverse(t, truth) == 0 {
		result = 2
	}
	if fPhase == 0 {
		return result
	}
	canon.FlipVariable(t, i)
	canon.FlipVariable(t, j)
	if canon.CompareReverse(t, truth) == 0 {
		result++
	}
	return result
}
func Abc_TtIsSymmetricHigh(p *Abc_TgMan_t, a, b, fPhase int) int {
	t := append([]uint64(nil), p.pTruth...)
	i, j := a, b
	for i >= 0 && j >= 0 {
		canon.SwapVariables(t, i, j)
		i, j = p.symLink[i], p.symLink[j]
	}
	if i >= 0 || j >= 0 {
		panic("DAU symmetric groups have unequal sizes")
	}
	result := 0
	if canon.CompareReverse(t, p.pTruth) == 0 {
		result = 2
	}
	if fPhase == 0 {
		return result
	}
	for i, j = a, b; i >= 0 && j >= 0; i, j = p.symLink[i], p.symLink[j] {
		if p.symPhase[i] != 0 {
			canon.FlipVariable(t, i)
		}
		if p.symPhase[j] != 0 {
			canon.FlipVariable(t, j)
		}
	}
	if canon.CompareReverse(t, p.pTruth) == 0 {
		result++
	}
	return result
}
func Abc_TgCreateGroups(p *Abc_TgMan_t) {
	store := make([]int32, p.nVars+1)
	ones := Abc_TtCountOnesInTruth(p.pTruth, int32(p.nVars))
	if ones > int32(len(p.pTruth)*32) {
		for i := range p.pTruth {
			p.pTruth[i] = ^p.pTruth[i]
		}
		ones = int32(len(p.pTruth)*64) - ones
		p.uPhase |= 1 << uint(p.nVars)
	}
	Abc_TtCountOnesInCofs(p.pTruth, int32(p.nVars), store)
	store[p.nVars] = ones
	for v := 0; v < p.nVars; v++ {
		if store[v] < ones-store[v] {
			canon.FlipVariable(p.pTruth, v)
			p.uPhase |= 1 << uint(v)
			store[v] = ones - store[v]
		}
	}
	Abc_TgSplitGroup(p, 0, store)
	p.fPhased = store[0]*2 != ones
}
func Abc_TgGroupSymmetry(p *Abc_TgMan_t, index int, high bool) int {
	g := p.pGroup[index]
	vars := p.pPerm[g.iStart : g.iStart+g.nGVars]
	phase := 0
	if index == 0 && !p.fPhased {
		phase = 1
	}
	var done [16]bool
	var counts, types [16]int
	for i := range vars {
		counts[i] = 1
	}
	merged := 0
	for {
		modified := false
		for i := 0; i < len(vars)-1; i++ {
			a := vars[i]
			if a < 0 || done[i] {
				continue
			}
			for j := i + 1; j < len(vars); j++ {
				b := vars[j]
				if b < 0 || counts[j] != counts[i] {
					types[j] = 0
				} else if counts[j] == 1 {
					types[j] = Abc_TtIsSymmetric(p.pTruth, p.nVars, a, b, phase)
				} else {
					types[j] = Abc_TtIsSymmetricHigh(p, a, b, phase)
				}
			}
			done[i] = true
			for j := i + 1; j < len(vars); j++ {
				b := vars[j]
				switch types[j] {
				case 1:
					Abc_TgFlipSymGroupByVar(p, b)
					fallthrough
				case 2:
					p.symPhase[a] += p.symPhase[b]
				case 3:
					Abc_TgClearSymGroupPhase(p, b)
				default:
					continue
				}
				tail := a
				for p.symLink[tail] >= 0 {
					tail = p.symLink[tail]
				}
				p.symLink[tail] = b
				vars[j] = -1
				counts[i] += counts[j]
				modified = true
				done[i] = false
				merged++
			}
		}
		if !high || !modified {
			break
		}
	}
	return merged
}
func dauHasVariable(t []uint64, v int) bool {
	copy := append([]uint64(nil), t...)
	canon.FlipVariable(copy, v)
	return canon.CompareReverse(t, copy) != 0
}
func Abc_TgPurgeSymmetry(p *Abc_TgMan_t, high bool) {
	var removed [16]int
	if !p.fPhased {
		for j := 0; j < p.pGroup[0].nGVars; j++ {
			v := p.pPerm[j]
			if !dauHasVariable(p.pTruth, v) {
				p.symLink[v] = p.symLink[p.nVars]
				p.symLink[p.nVars] = v
				p.pPerm[j] = -1
				removed[0]++
			}
		}
	}
	for k := 0; k < p.nGroups; k++ {
		removed[k] += Abc_TgGroupSymmetry(p, k, high)
	}
	write := 0
	for i := 0; i < p.nVars; i++ {
		if p.pPerm[i] >= 0 {
			p.pPerm[write] = p.pPerm[i]
			write++
		}
	}
	sum := 0
	for k := 0; k < p.nGroups; k++ {
		p.pGroup[k].nGVars -= removed[k]
		p.pGroup[k].iStart -= sum
		sum += removed[k]
	}
	if p.pGroup[0].nGVars == 0 {
		p.nGroups--
		copy(p.pGroup[:], p.pGroup[1:p.nGroups+1])
	}
	p.nGVars -= sum
}
func Abc_TgExpendSymmetry(p *Abc_TgMan_t, dest []byte) int {
	n := 0
	for j := 0; j < p.nGVars; j++ {
		for v := p.pPerm[j]; v >= 0; v = p.symLink[v] {
			dest[n] = byte(v)
			n++
		}
	}
	supported := n
	for v := p.symLink[p.nVars]; v >= 0; v = p.symLink[v] {
		dest[n] = byte(v)
		n++
	}
	if n != p.nVars {
		panic("incomplete DAU symmetry expansion")
	}
	return supported
}
func Abc_TgResetGroup(p *Abc_TgMan_t) {
	for i := 1; i < p.nGVars; i++ {
		v := p.pPerm[i]
		j := i
		for j > 0 && p.pPerm[j-1] > v {
			p.pPerm[j] = p.pPerm[j-1]
			j--
		}
		p.pPerm[j] = v
	}
	var dest [16]byte
	Abc_TgExpendSymmetry(p, dest[:])
	Abc_TgImplementPerm(p, dest[:])
	p.fPhased = false
	p.nGroups = 1
	p.pGroup[0].nGVars = p.nGVars
	if p.vPhase != nil {
		vec.Vec_IntClear(p.vPhase)
	}
}
func Abc_TgResetGroup1(p *Abc_TgMan_t) {
	var dest [16]byte
	n := Abc_TgExpendSymmetry(p, dest[:])
	for i := 1; i < n; i++ {
		v := dest[i]
		j := i
		for j > 0 && dest[j-1] > v {
			dest[j] = dest[j-1]
			j--
		}
		dest[j] = v
	}
	Abc_TgImplementPerm(p, dest[:])
	p.fPhased = false
	p.nGroups = 1
	p.nGVars = n
	p.pGroup[0].nGVars = n
	for i := 0; i < p.nVars; i++ {
		p.pPerm[i] = int(dest[i])
		p.symPhase[i] = 1
		p.symLink[i] = -1
	}
	if p.vPhase != nil {
		vec.Vec_IntClear(p.vPhase)
	}
}
func Abc_TgSymGroupPerm(p *Abc_TgMan_t, index int, swapOnly bool) int {
	work, best := make([]uint64, len(p.pTruth)), make([]uint64, len(p.pTruth))
	var cur, minimal Abc_TgMan_t
	Abc_TgManCopy(&cur, work, p)
	if swapOnly {
		Abc_TgSwapAdjacentSymGroups(&cur, index)
		if canon.CompareReverse(p.pTruth, work) < 0 {
			Abc_TgManCopy(p, p.pTruth, &cur)
			return 4
		}
		return 0
	}
	Abc_TgManCopy(&minimal, best, p)
	config := 0
	for step, c := range []int{1, 3, 2, 6, 7, 5, 4} {
		switch step {
		case 0, 2, 5:
			Abc_TgFlipSymGroup(&cur, index)
		case 1, 4, 6:
			Abc_TgFlipSymGroup(&cur, index+1)
		case 3:
			Abc_TgSwapAdjacentSymGroups(&cur, index)
		}
		if canon.CompareReverse(best, work) == 1 {
			Abc_TgManCopy(&minimal, best, &cur)
			config = c
		}
	}
	if config != 0 {
		Abc_TgManCopy(p, p.pTruth, &minimal)
	}
	return config
}
func Abc_TgPermPhase(p *Abc_TgMan_t, v int) int {
	t := append([]uint64(nil), p.pTruth...)
	pos := p.pPermTRev[v]
	canon.FlipVariable(t, pos)
	if canon.CompareReverse(p.pTruth, t) == 1 {
		copy(p.pTruth, t)
		p.uPhase ^= 1 << uint(pos)
		return 16
	}
	return 0
}
func Abc_TgSimpleEnumeration(p *Abc_TgMan_t) {
	var groups [16]int
	k := 0
	for g := 0; g < p.nGroups; g++ {
		for i := 0; i < p.pGroup[g].nGVars; i++ {
			groups[k] = g
			k++
		}
	}
	for pass := 0; pass < 5; pass++ {
		changes := 0
		for i := p.nGVars - 2; i >= 0; i-- {
			if groups[i] == groups[i+1] {
				changes |= Abc_TgSymGroupPerm(p, i, groups[i] > 0 || p.fPhased)
			}
		}
		for i := 1; i < p.nGVars-1; i++ {
			if groups[i] == groups[i+1] {
				changes |= Abc_TgSymGroupPerm(p, i, groups[i] > 0 || p.fPhased)
			}
		}
		for v := p.nVars - 1; v >= 0; v-- {
			if p.symPhase[v] != 0 {
				changes |= Abc_TgPermPhase(p, v)
			}
		}
		for v := 1; v < p.nVars; v++ {
			if p.symPhase[v] != 0 {
				changes |= Abc_TgPermPhase(p, v)
			}
		}
		if changes == 0 {
			break
		}
	}
}

func Abc_NextPermSwapC(data, dir []int, size int) int {
	k := -1
	for i := 0; i < size; i++ {
		j := i + dir[i]
		if j >= 0 && j < size && data[i] > data[j] && (k < 0 || data[i] > data[k]) {
			k = i
		}
	}
	if k < 0 {
		k = 0
	}
	for i := 0; i < size; i++ {
		if data[i] > data[k] {
			dir[i] = -dir[i]
		}
	}
	j := k + dir[k]
	return min(j, k)
}
func Abc_TgIsInitPerm(data, dir []int, size int) bool {
	if size == 0 {
		return true
	}
	if dir[0] != -1 {
		return false
	}
	for i := 1; i < size; i++ {
		if dir[i] != -1 || data[i] < data[i-1] {
			return false
		}
	}
	return true
}
func Abc_TgFirstPermutation(p *Abc_TgMan_t) {
	for i := 0; i < p.nGVars; i++ {
		p.pPermDir[i] = -1
	}
}
func Abc_TgNextPermutation(p *Abc_TgMan_t) bool {
	for i := 0; i < p.nGroups; i++ {
		g := p.pGroup[i]
		if g.nGVars <= 1 {
			continue
		}
		j := Abc_NextPermSwapC(p.pPerm[g.iStart:], p.pPermDir[g.iStart:], g.nGVars)
		if j >= 0 {
			Abc_TgSwapAdjacentSymGroups(p, g.iStart+j)
			return true
		}
		Abc_TgSwapAdjacentSymGroups(p, g.iStart)
	}
	return false
}
func grayCode(a uint32) uint32 { return a ^ (a >> 1) }
func grayFlip(a uint32) int    { return bits.TrailingZeros32(a + 1) }
func Abc_TgSaveBest(p, best *Abc_TgMan_t) {
	if canon.CompareReverse(best.pTruth, p.pTruth) == 1 {
		Abc_TgManCopy(best, best.pTruth, p)
	}
}
func Abc_TgPhaseEnumeration(p, best *Abc_TgMan_t) {
	Abc_TgSaveBest(p, best)
	if p.fPhased {
		return
	}
	n := p.pGroup[0].nGVars
	var order [16]int
	for i := 0; i < n; i++ {
		v := p.pPerm[i]
		j := i
		for j > 0 && p.symPhase[order[j-1]] > p.symPhase[v] {
			order[j] = order[j-1]
			j--
		}
		order[j] = v
	}
	for i := uint32(0); i < (1<<uint(n))-1; i++ {
		Abc_TgFlipSymGroupByVar(p, order[grayFlip(i)])
		Abc_TgSaveBest(p, best)
	}
}
func Abc_TgCalcScc(p *Abc_TgMan_t, out []int32, sortCoefficients bool) {
	Abc_TtSccInCofs(p.pTruth, int32(p.nVars), out)
	i := 0
	for j := 0; j < p.nGVars; j++ {
		out[j] = out[i]
		for k := p.pPerm[j]; k >= 0; k = p.symLink[k] {
			i++
		}
	}
	if !sortCoefficients {
		return
	}
	for g := 0; g < p.nGroups; g++ {
		start := p.pGroup[g].iStart
		end := start + p.pGroup[g].nGVars
		for i := start + 1; i < end; i++ {
			v := out[i]
			j := i
			for j > start && out[j-1] > v {
				out[j] = out[j-1]
				j--
			}
			out[j] = v
		}
	}
}
func Abc_TgSplitGroupsByScc(p *Abc_TgMan_t) {
	var cofs [16]int32
	Abc_TgCalcScc(p, cofs[:], false)
	for g := 0; g < p.nGroups; g++ {
		g += Abc_TgSplitGroup(p, g, cofs[p.pGroup[g].iStart:])
	}
	var dest [16]byte
	Abc_TgExpendSymmetry(p, dest[:])
	Abc_TgImplementPerm(p, dest[:])
}
func Abc_TgCompareCoef(a, b []int32, n int) int {
	for i := 0; i < n; i++ {
		if a[i] < b[i] {
			return -1
		}
		if a[i] > b[i] {
			return 1
		}
	}
	return 0
}

var log2fn = [17]int{0, 0, 100, 258, 458, 691, 949, 1230, 1530, 1847, 2179, 2525, 2884, 3254, 3634, 4025, 4425}

func Abc_TgPermCostScc(p *Abc_TgMan_t, cofs []int32) int {
	cost, k := 0, 0
	for g := 0; g < p.nGroups; g++ {
		n := p.pGroup[g].nGVars
		same := 1
		k++
		for j := 1; j < n; j++ {
			if cofs[k] == cofs[k-1] {
				same++
			} else {
				cost += log2fn[same]
				same = 1
			}
			k++
		}
		cost += log2fn[same]
	}
	return cost
}
func Abc_TgPermEnumerationScc(p, best *Abc_TgMan_t) {
	var cur Abc_TgMan_t
	Abc_TgManCopy(&cur, make([]uint64, len(p.pTruth)), p)
	if p.nAlgorithm > 1 {
		Abc_TgSplitGroupsByScc(&cur)
	}
	Abc_TgFirstPermutation(&cur)
	for {
		Abc_TgSaveBest(&cur, best)
		if !Abc_TgNextPermutation(&cur) {
			break
		}
	}
}
func Abc_TgReorderFGrps(p *Abc_TgMan_t) {
	n := p.pGroup[0].nGVars
	if p.fPhased {
		n = 0
	}
	for i := 0; i < n; i++ {
		v := p.pPerm[i]
		j := i
		for j > 0 && p.symPhase[p.pFGrps[j-1]] > p.symPhase[v] {
			p.pFGrps[j] = p.pFGrps[j-1]
			j--
		}
		p.pFGrps[j] = v
	}
}
func ilog2(n int) int {
	if n <= 1 {
		return 0
	}
	return bits.Len(uint(n)) - 1
}

type Abc_SccCost_t struct{ cNeg, cPhase, cPerm int }

func Abc_TgRecordPhase(p *Abc_TgMan_t, mode int) Abc_SccCost_t {
	var cofs, minCofs [18]int32
	ret := Abc_SccCost_t{}
	if p.fPhased {
		Abc_TgCalcScc(p, cofs[2:], true)
		ret.cPerm = Abc_TgPermCostScc(p, cofs[2:])
		return ret
	}
	n := p.pGroup[0].nGVars
	if n == 0 {
		return ret
	}
	if p.vPhase == nil {
		panic("DAU SCC enumeration needs phase storage")
	}
	start := 0
	if mode == 0 {
		start = 1
	}
	nCoefs := p.nGVars + 2 - start
	Abc_TgReorderFGrps(p)
	minCofs[1] = Abc_TtScc(p.pTruth, int32(p.nVars))
	Abc_TgCalcScc(p, minCofs[2:], true)
	if mode != 0 {
		minCofs[0] = int32(Abc_TgPermCostScc(p, minCofs[2:]))
	}
	vec.Vec_IntPush(p.vPhase, 0)
	for i := uint32(0); ; i++ {
		j := grayFlip(i)
		if j >= n {
			break
		}
		Abc_TgFlipSymGroupByVar(p, p.pFGrps[j])
		cofs[1] = Abc_TtScc(p.pTruth, int32(p.nVars))
		if mode == 0 && cofs[1] > minCofs[1] {
			continue
		}
		Abc_TgCalcScc(p, cofs[2:], true)
		if mode > 0 {
			cofs[0] = int32(Abc_TgPermCostScc(p, cofs[2:]))
		}
		if Abc_TgCompareCoef(cofs[start:], minCofs[start:], nCoefs) < 0 {
			copy(minCofs[start:start+nCoefs], cofs[start:start+nCoefs])
			vec.Vec_IntClear(p.vPhase)
		}
		if Abc_TgCompareCoef(cofs[start:], minCofs[start:], nCoefs) == 0 {
			vec.Vec_IntPush(p.vPhase, int32(grayCode(i+1)))
		}
	}
	Abc_TgFlipSymGroupByVar(p, p.pFGrps[n-1])
	ret.cNeg = n
	ret.cPhase = ilog2(vec.Vec_IntSize(p.vPhase))
	ret.cPerm = Abc_TgPermCostScc(p, minCofs[2:])
	return ret
}
func Abc_TgRecordPhase1(p *Abc_TgMan_t) int {
	if p.fPhased {
		return 0
	}
	n := p.pGroup[0].nGVars
	if n == 0 {
		return 0
	}
	if p.vPhase == nil {
		panic("DAU SCC enumeration needs phase storage")
	}
	Abc_TgReorderFGrps(p)
	minimum := Abc_TtScc(p.pTruth, int32(p.nVars))
	vec.Vec_IntPush(p.vPhase, 0)
	for i := uint32(0); ; i++ {
		j := grayFlip(i)
		if j >= n {
			break
		}
		Abc_TgFlipSymGroupByVar(p, p.pFGrps[j])
		value := Abc_TtScc(p.pTruth, int32(p.nVars))
		if value > minimum {
			continue
		}
		if value < minimum {
			minimum = value
			vec.Vec_IntClear(p.vPhase)
		}
		vec.Vec_IntPush(p.vPhase, int32(grayCode(i+1)))
	}
	Abc_TgFlipSymGroupByVar(p, p.pFGrps[n-1])
	return ilog2(vec.Vec_IntSize(p.vPhase))
}
func Abc_TgPhaseEnumerationScc(p, best *Abc_TgMan_t) {
	if p.fPhased {
		Abc_TgPermEnumerationScc(p, best)
		return
	}
	if p.pGroup[0].nGVars == 0 {
		Abc_TgSaveBest(p, best)
		return
	}
	var cur Abc_TgMan_t
	Abc_TgManCopy(&cur, make([]uint64, len(p.pTruth)), p)
	previous := int32(0)
	for _, phase := range vec.Vec_IntArray(p.vPhase) {
		flip := phase ^ previous
		for j := 0; j < p.pGroup[0].nGVars; j++ {
			if flip&(1<<uint(j)) != 0 {
				Abc_TgFlipSymGroupByVar(&cur, p.pFGrps[j])
			}
		}
		previous = phase
		Abc_TgPermEnumerationScc(&cur, best)
	}
}
func Abc_TgFullEnumeration(work, best *Abc_TgMan_t) {
	if work.nAlgorithm > 0 {
		for i := range best.pTruth {
			best.pTruth[i] = ^uint64(0)
		}
		Abc_TgPhaseEnumerationScc(work, best)
	} else {
		Abc_TgFirstPermutation(work)
		for {
			Abc_TgPhaseEnumeration(work, best)
			if !Abc_TgNextPermutation(work) {
				break
			}
		}
	}
	best.uPhase |= 1 << 30
}
func Abc_SccPhaseCost(p *Abc_TgMan_t) float64 {
	return float64(p.nVars)*0.997 + float64(p.nGVars)*1.043 - 15.9
}
func Abc_SccEnumCost(p *Abc_TgMan_t, c Abc_SccCost_t) float64 {
	switch p.nAlgorithm {
	case 0:
		return float64(p.nVars) + float64(c.cPhase)*1.09 + float64(c.cPerm)*0.01144
	case 1:
		return float64(p.nVars) + float64(c.cPhase)*0.855 + float64(c.cPerm)*0.00797
	case 2:
		return float64(p.nVars)*0.94 + float64(c.cPhase)*0.885 + float64(c.cPerm)*0.00855 - 20.59
	}
	return 0
}
func Abc_TgEnumerationCost(p *Abc_TgMan_t) int {
	if p.nGroups == 0 {
		return 0
	}
	cost := Abc_SccCost_t{}
	for i := 0; i < p.nGroups; i++ {
		cost.cPerm += log2fn[p.pGroup[i].nGVars]
	}
	if !p.fPhased {
		if p.nAlgorithm == 0 {
			cost.cPhase = p.pGroup[0].nGVars
		} else {
			cost.cPhase = Abc_TgRecordPhase1(p)
		}
	}
	return int(Abc_SccEnumCost(p, cost) + 0.5)
}
func Abc_TtCanonicizeAda(p *Abc_TtHieMan_t, truth []uint64, nVars int32, perm []byte, threshold int32) uint32 {
	if nVars == 0 {
		truth[0] = 0
		return 0
	}
	if nVars < 0 || nVars > 16 || len(perm) < int(nVars) {
		panic("invalid DAU adaptive dimensions")
	}
	Abc_TtVerifySmallTruth(truth, nVars)
	algorithm := 0
	if threshold >= 1000 {
		algorithm = 1
	}
	if algorithm != 0 && p == nil {
		panic("DAU SCC algorithm requires a cache manager")
	}
	const hash uint32 = 1 << 29
	if p != nil && Abc_TtHieRetrieveOrInsert(p, -5, truth, truth) > 0 {
		return hash
	}
	var cur, saved Abc_TgMan_t
	var phase *vec.Vec_Int_t
	if p != nil {
		phase = p.vPhase
	}
	Abc_TgInitMan(&cur, truth, int(nVars), algorithm, phase)
	Abc_TgCreateGroups(&cur)
	if p != nil && Abc_TtHieRetrieveOrInsert(p, -4, truth, truth) > 0 {
		return hash
	}
	Abc_TgPurgeSymmetry(&cur, threshold%1000 >= 100)
	Abc_TgExpendSymmetry(&cur, perm)
	Abc_TgImplementPerm(&cur, perm)
	cost := Abc_TgEnumerationCost(&cur)
	exact := threshold%100 > 84 || cost < int(threshold%100)
	if p == nil || len(p.levels) == 1 {
		if exact {
			Abc_TgManCopy(&saved, make([]uint64, len(cur.pTruth)), &cur)
			Abc_TgFullEnumeration(&saved, &cur)
		} else {
			Abc_TgSimpleEnumeration(&cur)
		}
	} else {
		flag := uint32(0)
		if exact {
			flag = 1 << 30
		}
		if Abc_TtHieRetrieveOrInsert(p, -3, truth, truth) > 0 {
			return hash | flag
		}
		Abc_TgManCopy(&saved, make([]uint64, len(cur.pTruth)), &cur)
		Abc_TgSimpleEnumeration(&cur)
		if Abc_TtHieRetrieveOrInsert(p, -2, truth, truth) > 0 {
			return hash | flag
		}
		if exact {
			Abc_TgManCopy(&cur, truth, &saved)
			Abc_TgFullEnumeration(&saved, &cur)
		}
		Abc_TtHieRetrieveOrInsert(p, -1, truth, truth)
	}
	for i := 0; i < int(nVars); i++ {
		perm[i] = byte(cur.pPermT[i])
	}
	return cur.uPhase
}
func Abc_TtCanonicizeCA(p *Abc_TtHieMan_t, truth []uint64, nVars int32, perm []byte, costAware int32) uint32 {
	if nVars == 0 {
		truth[0] = 0
		return 0
	}
	if p == nil || nVars < 0 || nVars > 16 || len(perm) < int(nVars) {
		panic("invalid DAU cost-aware dimensions")
	}
	Abc_TtVerifySmallTruth(truth, nVars)
	const hash uint32 = 1 << 29
	if Abc_TtHieRetrieveOrInsert(p, -5, truth, truth) > 0 {
		return hash
	}
	var cur, saved Abc_TgMan_t
	Abc_TgInitMan(&cur, truth, int(nVars), 2, p.vPhase)
	Abc_TgCreateGroups(&cur)
	if Abc_TtHieRetrieveOrInsert(p, -4, truth, truth) > 0 {
		return hash
	}
	Abc_TgPurgeSymmetry(&cur, true)
	Abc_TgExpendSymmetry(&cur, perm)
	Abc_TgImplementPerm(&cur, perm)
	if Abc_TtHieRetrieveOrInsert(p, -3, truth, truth) > 0 {
		return hash
	}
	Abc_TgManCopy(&saved, make([]uint64, len(cur.pTruth)), &cur)
	Abc_TgSimpleEnumeration(&cur)
	if Abc_TtHieRetrieveOrInsert(p, -2, truth, truth) > 0 {
		return hash
	}
	Abc_TgManCopy(&cur, truth, &saved)
	for i := range truth[:len(cur.pTruth)] {
		truth[i] = ^uint64(0)
	}
	cost := Abc_TgRecordPhase(&saved, 0)
	hard := uint32(0)
	if costAware != 0 && Abc_SccEnumCost(&saved, cost) > Abc_SccPhaseCost(&saved) {
		Abc_TgResetGroup(&saved)
		Abc_TgRecordPhase(&saved, 1)
		hard = 1 << 28
	}
	Abc_TgFullEnumeration(&saved, &cur)
	Abc_TtHieRetrieveOrInsert(p, -1, truth, truth)
	for i := 0; i < int(nVars); i++ {
		perm[i] = byte(cur.pPermT[i])
	}
	return cur.uPhase | hard
}
