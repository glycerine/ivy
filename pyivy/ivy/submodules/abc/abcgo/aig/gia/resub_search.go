package gia

import "github.com/glycerine/abcgo/misc/vec"

func resubPushGate(p *Gia_ResbMan_t, a, b int32) {
	vec.Vec_IntPush(p.vGates, a)
	vec.Vec_IntPush(p.vGates, b)
}
func resubPairParts(lit int32) (int32, int32, int32) {
	packed := lit >> 1
	return packed & 0x7fff, packed >> 15, lit & 1
}
func resubTopWeight(v *Vec_Int_t) int32 {
	if vec.Vec_IntSize(v) == 0 {
		return 0
	}
	return vec.Vec_IntEntry(v, 0)
}

func Gia_ManResubPerform_rec(p *Gia_ResbMan_t, limit int32) int32 {
	nVars := int32(vec.Vec_PtrSize(p.vDivs))
	if resubCount(p.pSets[1]) == 0 {
		return 0
	}
	if resubCount(p.pSets[0]) == 0 {
		return 1
	}
	result := Gia_ManFindOneUnate(p.pSets, p.vDivs, p.nWords, p.vUnateLits, p.vNotUnateVars, p.fVerbose)
	if result >= 0 {
		return result
	}
	if limit == 0 {
		return -1
	}
	Gia_ManSortUnates(p.pSets, p.vDivs, p.nWords, p.vUnateLits, p.vUnateLitsW, p.vSorter)
	result = Gia_ManFindTwoUnate(p.pSets, p.vDivs, p.nWords, p.vUnateLits, p.vUnateLitsW, p.fVerbose)
	if result >= 0 {
		a, b, phase := resubPairParts(result)
		node := nVars + int32(vec.Vec_IntSize(p.vGates)/2)
		resubPushGate(p, a, b)
		return 2*node + phase
	}
	vec.Vec_IntTwoFindCommon(p.vNotUnateVars[0], p.vNotUnateVars[1], p.vBinateVars)
	if int32(vec.Vec_IntSize(p.vBinateVars)) > p.nDivsMax {
		vec.Vec_IntShrink(p.vBinateVars, int(p.nDivsMax))
	}
	if p.fUseXor != 0 {
		result = Gia_ManFindXor(p.pSets, p.vDivs, p.nWords, p.vBinateVars, p.vUnatePairs, p.fVerbose)
		if result >= 0 {
			a, b, phase := resubPairParts(result)
			node := nVars + int32(vec.Vec_IntSize(p.vGates)/2)
			resubPushGate(p, a, b)
			return 2*node + phase
		}
	}
	if limit == 1 {
		return -1
	}
	Gia_ManFindUnatePairs(p.pSets, p.vDivs, p.nWords, p.vBinateVars, p.vUnatePairs, p.fVerbose)
	Gia_ManSortPairs(p.pSets, p.vDivs, p.nWords, p.vUnatePairs, p.vUnatePairsW, p.vSorter)
	result = Gia_ManFindDivGate(p.pSets, p.vDivs, p.nWords, p.vUnateLits, p.vUnatePairs, p.vUnateLitsW, p.vUnatePairsW, p.pDivA)
	if result >= 0 {
		a, pairIndex, phase := resubPairParts(result)
		pair := vec.Vec_IntEntry(p.vUnatePairs[1-phase], int(pairIndex>>1))
		u, v, pairPhase := resubPairParts(pair)
		pairPhase ^= pairIndex & 1
		node := nVars + int32(vec.Vec_IntSize(p.vGates)/2)
		resubPushGate(p, u, v)
		resubPushGate(p, a, 2*node+pairPhase)
		return 2*(node+1) + phase
	}
	if limit >= 3 {
		result = Gia_ManFindGateGate(p.pSets, p.vDivs, p.nWords, p.vUnatePairs, p.vUnatePairsW, p.pDivA, p.pDivB)
		if result >= 0 {
			x, y, phase := resubPairParts(result)
			first := vec.Vec_IntEntry(p.vUnatePairs[1-phase], int(x>>1))
			second := vec.Vec_IntEntry(p.vUnatePairs[1-phase], int(y>>1))
			a, b, c := resubPairParts(first)
			d, e, f := resubPairParts(second)
			c ^= x & 1
			f ^= y & 1
			node := nVars + int32(vec.Vec_IntSize(p.vGates)/2)
			resubPushGate(p, a, b)
			resubPushGate(p, d, e)
			resubPushGate(p, 2*node+c, 2*(node+1)+f)
			return 2*(node+2) + phase
		}
	}
	ones := [2]int32{resubTopWeight(p.vUnateLitsW[0]), resubTopWeight(p.vUnateLitsW[1])}
	twos := [2]int32{resubTopWeight(p.vUnatePairsW[0]), resubTopWeight(p.vUnatePairsW[1])}
	maxOne, maxTwo := max(ones[0], ones[1]), max(twos[0], twos[1])
	if max(maxOne, maxTwo) == 0 {
		return -1
	}
	if maxOne > maxTwo/2 && limit >= 2 {
		phase := int32(0)
		if maxOne == ones[0] {
			phase = 1
		}
		lit := vec.Vec_IntEntry(p.vUnateLits[1-phase], 0)
		div := resubDiv(p.vDivs, lit>>1, p.nWords)
		for i := range div {
			d := div[i]
			if lit&1 == 0 {
				d = ^d
			}
			p.pSets[phase][i] &= d
		}
		result = Gia_ManResubPerform_rec(p, limit-1)
		if result >= 0 {
			node := nVars + int32(vec.Vec_IntSize(p.vGates)/2)
			resubPushGate(p, lit^1, result^phase)
			return 2*node + phase
		}
	} else if maxTwo > 0 && limit >= 3 {
		phase := int32(0)
		if maxTwo == twos[0] {
			phase = 1
		}
		pair := vec.Vec_IntEntry(p.vUnatePairs[1-phase], 0)
		a, b, comp := resubPairParts(pair)
		temp := make([]word, p.nWords)
		Gia_ManDeriveDivPair(pair, p.vDivs, p.nWords, temp)
		for i, d := range temp {
			if comp == 0 {
				d = ^d
			}
			p.pSets[phase][i] &= d
		}
		result = Gia_ManResubPerform_rec(p, limit-2)
		if result >= 0 {
			node := nVars + int32(vec.Vec_IntSize(p.vGates)/2)
			resubPushGate(p, a, b)
			resubPushGate(p, result^phase, 2*node+(comp^1))
			return 2*(node+1) + phase
		}
	}
	return -1
}
func Gia_ManResubPerform(p *Gia_ResbMan_t, divs *Vec_Ptr_t, nWords, limit, divsMax, choice, useXor, debug, verbose int32) {
	Gia_ResbInit(p, divs, nWords, limit, divsMax, choice, useXor, debug, verbose, verbose)
	saved := [2][]word{append([]word(nil), p.pSets[0]...), append([]word(nil), p.pSets[1]...)}
	result := Gia_ManResubPerform_rec(p, limit)
	copy(p.pSets[0], saved[0])
	copy(p.pSets[1], saved[1])
	if result >= 0 {
		vec.Vec_IntPush(p.vGates, result)
	} else {
		vec.Vec_IntClear(p.vGates)
	}
}

func Gia_ManResubPerformMux_rec(p *Gia_ResbMan_t, limit int32) int32 {
	if limit < 3 {
		return -1
	}
	selector := Gia_ManResubFindBestBinate(p)
	if selector < 0 {
		return -1
	}
	saved := [2][]word{append([]word(nil), p.pSets[0]...), append([]word(nil), p.pSets[1]...)}
	start := vec.Vec_IntSize(p.vGates)
	success := false
	defer func() {
		copy(p.pSets[0], saved[0])
		copy(p.pSets[1], saved[1])
		if !success {
			vec.Vec_IntShrink(p.vGates, start)
		}
	}()
	div := resubDiv(p.vDivs, selector>>1, p.nWords)
	for phase := 0; phase < 2; phase++ {
		for i, d := range div {
			if selector&1 == 0 {
				d = ^d
			}
			p.pSets[phase][i] = saved[phase][i] & d
		}
	}
	lo := Gia_ManResubPerform_rec(p, limit-3)
	if lo < 0 {
		return -1
	}
	used := int32((vec.Vec_IntSize(p.vGates) - start) / 2)
	if limit-used < 3 {
		return -1
	}
	for phase := 0; phase < 2; phase++ {
		for i, d := range div {
			if selector&1 != 0 {
				d = ^d
			}
			p.pSets[phase][i] = saved[phase][i] & d
		}
	}
	hi := Gia_ManResubPerform_rec(p, limit-3-used)
	if hi < 0 {
		return -1
	}
	a := Gia_ManResubAddNode(p, selector^1, lo, 0)
	b := Gia_ManResubAddNode(p, selector, hi, 0)
	result := Gia_ManResubAddNode(p, a, b, 1)
	success = true
	return result
}
