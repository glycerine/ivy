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
