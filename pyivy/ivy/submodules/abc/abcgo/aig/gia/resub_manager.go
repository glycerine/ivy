package gia

import "github.com/glycerine/abcgo/misc/vec"

type Gia_ResbMan_t struct {
	nWords, nLimit, nDivsMax, iChoice, fUseXor, fDebug, fVerbose, fVeryVerbose int32
	vDivs                                                                      *Vec_Ptr_t
	vGates                                                                     *Vec_Int_t
	vUnateLits, vNotUnateVars, vUnatePairs, vUnateLitsW, vUnatePairsW          [2]*Vec_Int_t
	vBinateVars                                                                *Vec_Int_t
	vSorter                                                                    *Vec_Wec_t
	pSets                                                                      [2][]word
	pDivA, pDivB                                                               []word
}

func Gia_ResbAlloc(nWords int32) *Gia_ResbMan_t {
	if nWords <= 0 {
		panic("gia: invalid resubstitution truth size")
	}
	p := &Gia_ResbMan_t{nWords: nWords, vDivs: vec.Vec_PtrAlloc(100), vGates: vec.Vec_IntAlloc(100), vBinateVars: vec.Vec_IntAlloc(100), vSorter: vec.Vec_WecAlloc(int(nWords) * 64), pDivA: make([]word, nWords), pDivB: make([]word, nWords)}
	for i := 0; i < 2; i++ {
		p.vUnateLits[i] = vec.Vec_IntAlloc(100)
		p.vNotUnateVars[i] = vec.Vec_IntAlloc(100)
		p.vUnatePairs[i] = vec.Vec_IntAlloc(100)
		p.vUnateLitsW[i] = vec.Vec_IntAlloc(100)
		p.vUnatePairsW[i] = vec.Vec_IntAlloc(100)
		p.pSets[i] = make([]word, nWords)
	}
	return p
}
func Gia_ResbInit(p *Gia_ResbMan_t, divs *Vec_Ptr_t, nWords, limit, divsMax, choice, useXor, debug, verbose, veryVerbose int32) {
	if p == nil || p.nWords != nWords || vec.Vec_PtrSize(divs) < 2 || limit < 0 || divsMax < 0 {
		panic("gia: invalid resubstitution manager inputs")
	}
	p.nLimit, p.nDivsMax, p.iChoice, p.fUseXor, p.fDebug, p.fVerbose, p.fVeryVerbose = limit, divsMax, choice, useXor, debug, verbose, veryVerbose
	for i := 0; i < 2; i++ {
		copy(p.pSets[i], resubDiv(divs, int32(i), nWords))
		vec.Vec_IntClear(p.vUnateLits[i])
		vec.Vec_IntClear(p.vNotUnateVars[i])
		vec.Vec_IntClear(p.vUnatePairs[i])
		vec.Vec_IntClear(p.vUnateLitsW[i])
		vec.Vec_IntClear(p.vUnatePairsW[i])
	}
	vec.Vec_PtrClear(p.vDivs)
	vec.Vec_PtrAppend(p.vDivs, divs)
	vec.Vec_IntClear(p.vGates)
	vec.Vec_IntClear(p.vBinateVars)
}
func Gia_ResbFree(p *Gia_ResbMan_t) {
	if p != nil {
		*p = Gia_ResbMan_t{}
	}
}
func Gia_ManResubAddNode(p *Gia_ResbMan_t, a, b, kind int32) int32 {
	if a == b || kind < 0 || kind > 2 {
		panic("gia: invalid resubstitution gate")
	}
	node := int32(vec.Vec_PtrSize(p.vDivs) + vec.Vec_IntSize(p.vGates)/2)
	if (kind == 2) != (a > b) {
		a, b = b, a
	}
	if kind == 1 {
		a ^= 1
		b ^= 1
	}
	vec.Vec_IntPush(p.vGates, a)
	vec.Vec_IntPush(p.vGates, b)
	result := 2 * node
	if kind == 1 {
		result++
	}
	return result
}
func Gia_ManResubFindBestBinate(p *Gia_ResbMan_t) int32 {
	total := resubCount(p.pSets[0]) + resubCount(p.pSets[1])
	best, cost := int32(-1), int32(-1)
	for _, id := range vec.Vec_IntArray(p.vBinateVars) {
		d := resubDiv(p.vDivs, id, p.nWords)
		var n int32
		for i, v := range d {
			n += resubCount([]word{v & p.pSets[0][i], v & p.pSets[1][i]})
		}
		if n > cost {
			cost = n
			best = 2 * id
		}
		if total-n > cost {
			cost = total - n
			best = 2*id + 1
		}
	}
	return best
}
func Gia_ManResubVerify(p *Gia_ResbMan_t, result []word) int32 {
	gates := vec.Vec_IntArray(p.vGates)
	if len(gates) == 0 {
		return -1
	}
	if len(gates)%2 != 1 {
		panic("gia: invalid resubstitution gate list")
	}
	nVars := int32(vec.Vec_PtrSize(p.vDivs))
	sims := make([][]word, 0, len(gates)/2)
	resolve := func(lit int32) []word {
		id := lit >> 1
		if id < nVars {
			return resubDiv(p.vDivs, id, p.nWords)
		}
		if id < 0 || int(id-nVars) >= len(sims) {
			panic("gia: invalid resubstitution dependency")
		}
		return sims[id-nVars]
	}
	for i := 0; i < len(gates)-1; i += 2 {
		a, b := gates[i], gates[i+1]
		x, y := resolve(a), resolve(b)
		out := make([]word, p.nWords)
		if a>>1 < b>>1 {
			for k := range out {
				u, v := x[k], y[k]
				if a&1 != 0 {
					u = ^u
				}
				if b&1 != 0 {
					v = ^v
				}
				out[k] = u & v
			}
		} else if a>>1 > b>>1 && a&1 == 0 && b&1 == 0 {
			for k := range out {
				out[k] = x[k] ^ y[k]
			}
		} else {
			panic("gia: invalid resubstitution gate")
		}
		sims = append(sims, out)
	}
	top := gates[len(gates)-1]
	if top < 0 {
		panic("gia: invalid resubstitution top literal")
	}
	var function []word
	if top>>1 == 0 {
		function = make([]word, p.nWords)
	} else {
		function = resolve(top)
	}
	if result != nil && len(result) < int(p.nWords) {
		panic("gia: short resubstitution result")
	}
	valid := true
	for i, v := range function {
		if top&1 != 0 {
			v = ^v
		}
		if result != nil {
			result[i] = v
		}
		if p.pSets[0][i]&v != 0 || p.pSets[1][i]&^v != 0 {
			valid = false
		}
	}
	return giaBool(valid)
}
