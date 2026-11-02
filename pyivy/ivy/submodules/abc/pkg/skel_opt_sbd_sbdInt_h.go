package abc

// Hand-written skeleton of the types in src/opt/sbd/sbdInt.h. The constants
// are SBD_SAT_UNDEC 0x1234567812345678, SBD_SAT_SAT 0x8765432187654321,
// SBD_LUTS_MAX 2, SBD_SIZE_MAX 4, SBD_DIV_MAX 10 and SBD_FVAR_MAX 100.
// Sbd_Sto_t and Sbd_Srv_t are defined in sbdCut.c and sbdCut2.c.

// SbdStr is Sbd_Str_t.
type SbdStr struct {
	fLut    int     // LUT or SEL
	nVarIns int     // input count
	VarIns  [10]int // input vars
	Res     uint64  // word : result of solving
}
