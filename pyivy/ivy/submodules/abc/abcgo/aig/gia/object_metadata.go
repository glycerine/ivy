package gia

type Gia_Rpr_t struct {
	iRepr   uint32 // representative node, including the GIA_VOID sentinel
	fProved uint8  // source one-bit flag: proved equivalence
	fFailed uint8  // source one-bit flag: failed equivalence
	fColorA uint8  // source one-bit flag: cone A
	fColorB uint8  // source one-bit flag: cone B
}
type Gia_Plc_t struct {
	xCoord uint16 // source 15-bit x coordinate
	yCoord uint16 // source 15-bit y coordinate
	fFixed uint8  // source one-bit flag: placement is fixed
	fUndef uint8  // source one-bit flag: placement is unassigned
}
