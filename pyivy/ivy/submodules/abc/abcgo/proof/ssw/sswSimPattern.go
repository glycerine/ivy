package ssw

func sswPatternBit(pattern []uint32, bit int) int32 {
	return int32((pattern[bit/32] >> (bit % 32)) & 1)
}
func sswTogglePatternBit(words []uint32, bit int) { words[bit/32] ^= 1 << (bit % 32) }

// Ssw_SmlAssignDist1 seeds one pattern and, for one frame, its distance-one neighbors.
func Ssw_SmlAssignDist1(p *Ssw_Sml_t, pattern []uint32) {
	if p.nFrames <= 0 {
		panic("ssw: simulation requires a frame")
	}
	inputs := p.pAig.Inputs()
	if p.nFrames == 1 {
		for i, obj := range inputs {
			Ssw_SmlObjAssignConst(p, obj, sswPatternBit(pattern, i), 0)
		}
		limit := min(len(inputs), int(p.nWordsTotal)*32-1)
		for i := 0; i < limit; i++ {
			sswTogglePatternBit(Ssw_ObjSim(p, int32(inputs[i].NodeID())), i+1)
		}
		return
	}
	nPis := len(inputs) - int(p.pAig.RegisterCount())
	for f := int32(0); f < p.nFrames; f++ {
		for i, obj := range inputs[:nPis] {
			Ssw_SmlObjAssignConst(p, obj, sswPatternBit(pattern, nPis*int(f)+i), f)
		}
	}
	for i, obj := range inputs[nPis:] {
		Ssw_SmlObjAssignConst(p, obj, sswPatternBit(pattern, nPis*int(p.nFrames)+i), 0)
	}
}

// Ssw_SmlAssignDist1Plus seeds frame zero, then randomizes later PI frames.
func Ssw_SmlAssignDist1Plus(p *Ssw_Sml_t, pattern []uint32) {
	if p.nFrames <= 0 {
		panic("ssw: simulation requires a frame")
	}
	inputs := p.pAig.Inputs()
	nPis := len(inputs) - int(p.pAig.RegisterCount())
	for i, obj := range inputs {
		Ssw_SmlObjAssignConst(p, obj, sswPatternBit(pattern, i), 0)
	}
	limit := min(nPis, int(p.nWordsFrame)*32-1)
	for i := 0; i < limit; i++ {
		sswTogglePatternBit(Ssw_ObjSim(p, int32(inputs[i].NodeID())), i+1)
	}
	for f := int32(1); f < p.nFrames; f++ {
		for _, obj := range inputs[:nPis] {
			Ssw_SmlAssignRandomFrame(p, obj, f)
		}
	}
}
