package util

// Abc_Cex_t stores register initialization followed by frame-major input bits.
// NBits counts bits, despite the historical C header's misleading words comment.
type Abc_Cex_t struct {
	IPo, IFrame        int32
	NRegs, NPis, NBits int32
	PData              []uint32
}

// AbcCexSentinel replaces C's distinguished pointer value (Abc_Cex_t *)1.
var AbcCexSentinel = &Abc_Cex_t{IPo: -1, IFrame: -1}
