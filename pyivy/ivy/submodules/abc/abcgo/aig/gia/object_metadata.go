package gia

type Gia_Rpr_t struct {
	iRepr   uint32 // representative node, including the GIA_VOID sentinel
	fProved uint8  // source one-bit flag: proved equivalence
	fFailed uint8  // source one-bit flag: failed equivalence
	fColorA uint8  // source one-bit flag: cone A
	fColorB uint8  // source one-bit flag: cone B
}
type Gia_Plc_t struct {
	fFixed uint32 /* C bitfield: 1 bits; review packing and truncation */  // the placement of this object is fixed
	xCoord uint32 /* C bitfield: 15 bits; review packing and truncation */ // x-ooordinate of the placement
	fUndef uint32 /* C bitfield: 1 bits; review packing and truncation */  // the placement of this object is not assigned
	yCoord uint32 /* C bitfield: 15 bits; review packing and truncation */ // y-ooordinate of the placement
}
