package gia

import "fmt"

type Gia_Rsb2Man_t struct {
	nDivsMax, nLevelIncrease, fUseXor, fUseZeroCost, fDebug, fVerbose int32
	nObjs, nPis, nNodes, nPos, iFirstPo, Level, nMffc                 int32
	objects                                                           []int32
	sims                                                              []word
	divisors                                                          [][]word
	divisorIDs, levels, refs, copies, tried                           []int32
	truth0, truth1, careSet                                           word
}

func Gia_Rsb2ManAlloc() *Gia_Rsb2Man_t { return &Gia_Rsb2Man_t{} }
func Gia_Rsb2ManFree(p *Gia_Rsb2Man_t) {
	if p != nil {
		*p = Gia_Rsb2Man_t{}
	}
}
func Gia_Rsb2ManStart(p *Gia_Rsb2Man_t, objects []int32, nObjs, divsMax, levelIncrease, useXor, useZeroCost, debug, verbose int32) {
	if p == nil || nObjs < 1 || len(objects) != int(2*nObjs) || objects[0] != 0 || objects[1] != 0 {
		panic("gia: invalid Rsb2 objects")
	}
	p.nDivsMax, p.nLevelIncrease, p.fUseXor, p.fUseZeroCost, p.fDebug, p.fVerbose = divsMax, levelIncrease, useXor, useZeroCost, debug, verbose
	p.objects = append(p.objects[:0], objects...)
	p.nObjs = nObjs
	p.nPis, p.nNodes, p.nPos = 0, 0, 0
	for i := int32(1); i < nObjs; i++ {
		a, b := objects[2*i], objects[2*i+1]
		if a == 0 && b == 0 && p.nNodes == 0 && p.nPos == 0 {
			p.nPis++
		} else if a == b {
			p.nPos++
		} else {
			if p.nPos != 0 {
				panic("gia: Rsb2 internal gate after output")
			}
			p.nNodes++
		}
	}
	if p.nPis > 6 {
		panic("gia: Rsb2 truth tables support at most six inputs")
	}
	p.iFirstPo = nObjs - p.nPos
	p.sims = make([]word, 2*nObjs)
	p.sims[1] = ^word(0)
	patterns := [6]word{0xaaaaaaaaaaaaaaaa, 0xcccccccccccccccc, 0xf0f0f0f0f0f0f0f0, 0xff00ff00ff00ff00, 0xffff0000ffff0000, 0xffffffff00000000}
	for i := int32(0); i < p.nPis; i++ {
		p.sims[2*(i+1)] = patterns[i]
		p.sims[2*(i+1)+1] = ^patterns[i]
	}
	p.divisors = nil
	p.divisorIDs = nil
	p.levels = nil
	p.refs = nil
	p.copies = nil
	p.tried = nil
}
func Gia_Rsb2ManPrint(p *Gia_Rsb2Man_t) {
	fmt.Printf("PI = %d.  PO = %d.  Obj = %d.\n", p.nPis, p.nPos, p.nObjs)
	for i := p.nPis + 1; i < p.iFirstPo; i++ {
		a, b := p.objects[2*i], p.objects[2*i+1]
		op := '&'
		if a > b {
			op = '^'
		}
		fmt.Printf("%2d = %c%2d %c %c%2d;\n", i, rsb2Sign(a), a>>1, op, rsb2Sign(b), b>>1)
	}
	for i := p.iFirstPo; i < p.nObjs; i++ {
		a := p.objects[2*i]
		fmt.Printf("%2d = %c%2d;\n", i, rsb2Sign(a), a>>1)
	}
}
func rsb2Sign(lit int32) rune {
	if lit&1 != 0 {
		return '!'
	}
	return ' '
}
func Gia_Rsb2ManLevel(p *Gia_Rsb2Man_t) int32 {
	p.levels = make([]int32, p.nObjs)
	level := int32(0)
	for i := p.nPis + 1; i < p.iFirstPo; i++ {
		a, b := p.objects[2*i]>>1, p.objects[2*i+1]>>1
		p.levels[i] = 1 + max(p.levels[a], p.levels[b])
	}
	for i := p.iFirstPo; i < p.nObjs; i++ {
		p.levels[i] = p.levels[p.objects[2*i]>>1]
		level = max(level, p.levels[i])
	}
	return level
}
func (p *Gia_Rsb2Man_t) simulate() {
	for i := p.nPis + 1; i < p.iFirstPo; i++ {
		a, b := p.objects[2*i], p.objects[2*i+1]
		if a < 0 || b < 0 || a>>1 >= i || b>>1 >= i || a == b {
			panic("gia: invalid Rsb2 gate")
		}
		if a < b {
			p.sims[2*i] = p.sims[a] & p.sims[b]
		} else {
			p.sims[2*i] = p.sims[a] ^ p.sims[b]
		}
		p.sims[2*i+1] = ^p.sims[2*i]
	}
	for i := p.iFirstPo; i < p.nObjs; i++ {
		p.sims[2*i] = p.sims[p.objects[2*i]]
		p.sims[2*i+1] = ^p.sims[2*i]
	}
}
func Gia_Rsb2ManOdcs(p *Gia_Rsb2Man_t, node int32) word {
	if node <= p.nPis || node >= p.iFirstPo {
		panic("gia: invalid Rsb2 care node")
	}
	p.simulate()
	before := make([]word, p.nPos)
	for i := int32(0); i < p.nPos; i++ {
		before[i] = p.sims[2*(p.iFirstPo+i)]
	}
	p.sims[2*node], p.sims[2*node+1] = p.sims[2*node+1], p.sims[2*node]
	for i := node + 1; i < p.iFirstPo; i++ {
		a, b := p.objects[2*i], p.objects[2*i+1]
		if a < b {
			p.sims[2*i] = p.sims[a] & p.sims[b]
		} else {
			p.sims[2*i] = p.sims[a] ^ p.sims[b]
		}
		p.sims[2*i+1] = ^p.sims[2*i]
	}
	var care word
	for i := int32(0); i < p.nPos; i++ {
		care |= before[i] ^ p.sims[p.objects[2*(p.iFirstPo+i)]]
	}
	p.simulate()
	return care
}
func Gia_Rsb2ManDeref_rec(p *Gia_Rsb2Man_t, objects, refs []int32, node int32) int32 {
	if node <= p.nPis {
		return 0
	}
	count := int32(1)
	for k := int32(0); k < 2; k++ {
		f := objects[2*node+k] >> 1
		refs[f]--
		if refs[f] == 0 {
			count += Gia_Rsb2ManDeref_rec(p, objects, refs, f)
		}
	}
	return count
}
func Gia_Rsb2ManMffc(p *Gia_Rsb2Man_t, node int32) int32 {
	if node <= p.nPis || node >= p.iFirstPo {
		panic("gia: invalid Rsb2 MFFC node")
	}
	p.refs = make([]int32, p.nObjs)
	for i := p.nPis + 1; i < p.iFirstPo; i++ {
		p.refs[p.objects[2*i]>>1]++
		p.refs[p.objects[2*i+1]>>1]++
	}
	for i := p.iFirstPo; i < p.nObjs; i++ {
		p.refs[p.objects[2*i]>>1]++
	}
	p.refs[node] = 0
	for i := node + 1; i < p.iFirstPo; i++ {
		if p.refs[p.objects[2*i]>>1] == 0 || p.refs[p.objects[2*i+1]>>1] == 0 {
			p.refs[i] = 0
		}
	}
	return Gia_Rsb2ManDeref_rec(p, p.objects, p.refs, node)
}
func Gia_Rsb2ManDivs(p *Gia_Rsb2Man_t, node int32) int32 {
	p.careSet = Gia_Rsb2ManOdcs(p, node)
	p.truth1 = p.careSet & p.sims[2*node]
	p.truth0 = p.careSet &^ p.truth1
	p.divisors = [][]word{{p.truth0}, {p.truth1}}
	p.divisorIDs = []int32{-1, -1}
	for i := int32(1); i <= p.nPis; i++ {
		p.divisors = append(p.divisors, p.sims[2*i:2*i+1])
		p.divisorIDs = append(p.divisorIDs, i)
	}
	p.nMffc = Gia_Rsb2ManMffc(p, node)
	var nodeLevel int32
	if p.nLevelIncrease >= 0 {
		p.Level = Gia_Rsb2ManLevel(p)
		nodeLevel = p.levels[node]
	}
	for i := p.nPis + 1; i < p.iFirstPo; i++ {
		if p.refs[i] == 0 || p.nLevelIncrease >= 0 && p.levels[i] > nodeLevel+p.nLevelIncrease {
			continue
		}
		p.divisors = append(p.divisors, p.sims[2*i:2*i+1])
		p.divisorIDs = append(p.divisorIDs, i)
	}
	return int32(len(p.divisors))
}
