package gia

// Gia_ManCecSimAlloc allocates the source Cec2 packed object signatures and
// the separate append-only CI-pattern history.
func Gia_ManCecSimAlloc(p *Gia_Man_t, words int32) {
	if words <= 0 {
		panic("gia: invalid CEC simulation word count")
	}
	p.vSims = Vec_WrdStart(int(p.nObjs) * int(words))
	p.VSimsPi = Vec_WrdAlloc(int(Gia_ManCiNum(p)) * int(words) * 4)
	p.NSimWords = words
	p.iPatsPi = 0
}
func Gia_ManCecSimWords(p *Gia_Man_t) []uint64 {
	if p.vSims == nil {
		panic("gia: CEC simulation not allocated")
	}
	return Vec_WrdArray(p.vSims)
}
func Gia_ManCecSimPattern(p *Gia_Man_t) int32 { return p.iPatsPi }
func Gia_ManCecSetSimPattern(p *Gia_Man_t, pattern int32) {
	if pattern < 0 || pattern >= 64*p.NSimWords {
		panic("gia: invalid CEC simulation pattern")
	}
	p.iPatsPi = pattern
}
func Gia_ManCecSaveCIWord(p *Gia_Man_t, word uint64) {
	if p.VSimsPi == nil {
		panic("gia: CEC simulation not allocated")
	}
	Vec_WrdPush(p.VSimsPi, word)
}
func Gia_ManCecSimFree(p *Gia_Man_t) { p.vSims = nil; p.NSimWords = 0; p.iPatsPi = 0 }
