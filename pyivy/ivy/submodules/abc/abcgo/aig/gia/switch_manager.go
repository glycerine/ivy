package gia

// Gia_ParSwi_t configures word-parallel switching simulation.
type Gia_ParSwi_t struct {
	nWords, nIters, nPref, nRandPiFactor int32
	fProbOne, fProbTrans, fVerbose       int32
}
type Gia_ManSwi_t struct {
	pAig                               *Gia_Man_t
	pPars                              *Gia_ParSwi_t
	nWords                             int32
	pDataSim, pDataSimCis, pDataSimCos []uint32
	pData1                             []int32
}

func Gia_SwiData(p *Gia_ManSwi_t, i int32) []uint32 {
	start := i * p.nWords
	return p.pDataSim[start : start+p.nWords]
}
func Gia_SwiDataCi(p *Gia_ManSwi_t, i int32) []uint32 {
	start := i * p.nWords
	return p.pDataSimCis[start : start+p.nWords]
}
func Gia_SwiDataCo(p *Gia_ManSwi_t, i int32) []uint32 {
	start := i * p.nWords
	return p.pDataSimCos[start : start+p.nWords]
}
func Gia_ManSetDefaultParamsSwi(p *Gia_ParSwi_t) {
	*p = Gia_ParSwi_t{nWords: 10, nIters: 48, nPref: 16, fProbTrans: 1}
}
func Gia_ManSwiCreate(graph *Gia_Man_t, pars *Gia_ParSwi_t) *Gia_ManSwi_t {
	if graph == nil || pars == nil || pars.nWords <= 0 {
		panic("gia: invalid switching simulation parameters")
	}
	front := Gia_ManFront(graph)
	return &Gia_ManSwi_t{pAig: front, pPars: pars, nWords: pars.nWords, pDataSim: make([]uint32, pars.nWords*front.nFront), pDataSimCis: make([]uint32, pars.nWords*Gia_ManCiNum(front)), pDataSimCos: make([]uint32, pars.nWords*Gia_ManCoNum(front)), pData1: make([]int32, Gia_ManObjNum(graph))}
}
func Gia_ManSwiDelete(p *Gia_ManSwi_t) { Gia_ManStop(p.pAig); *p = Gia_ManSwi_t{} }
func giaSwiMask(factor int32) uint32 {
	if factor == -1 {
		return Gia_ManRandom(0)&Gia_ManRandom(0) | Gia_ManRandom(0)&Gia_ManRandom(0)&Gia_ManRandom(0)
	}
	if factor < 0 {
		panic("gia: invalid input toggle factor")
	}
	mask := Gia_ManRandom(0)
	for i := int32(0); i < factor; i++ {
		mask &= Gia_ManRandom(0)
	}
	return mask
}
func Gia_ManSwiSimInfoRandom(p *Gia_ManSwi_t, info []uint32, factor int32) {
	if factor == 0 {
		for w := int(p.nWords) - 1; w >= 0; w-- {
			info[w] = Gia_ManRandom(0)
		}
		return
	}
	mask := giaSwiMask(factor)
	for w := int(p.nWords) - 1; w >= 0; w-- {
		info[w] ^= mask
	}
}
func Gia_ManSwiSimInfoRandomShift(p *Gia_ManSwi_t, info []uint32, factor int32) {
	mask := giaSwiMask(factor)
	for w := int(p.nWords) - 1; w >= 0; w-- {
		info[w] = info[w]<<16 | (info[w]^mask)&0xffff
	}
}
func Gia_ManSwiSimInfoZero(p *Gia_ManSwi_t, info []uint32) { clear(info[:p.nWords]) }
func Gia_ManSwiSimInfoOne(p *Gia_ManSwi_t, info []uint32) {
	for w := int32(0); w < p.nWords; w++ {
		info[w] = ^uint32(0)
	}
}
func Gia_ManSwiSimInfoCopy(p *Gia_ManSwi_t, dst, src []uint32) { copy(dst[:p.nWords], src[:p.nWords]) }
func Gia_ManSwiSimInfoCopyShift(p *Gia_ManSwi_t, dst, src []uint32) {
	for w := int32(0); w < p.nWords; w++ {
		dst[w] = dst[w]<<16 | src[w]&0xffff
	}
}
