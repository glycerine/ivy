package gia

type Gia_Rpr_t struct {
	iRepr   uint32 /* C bitfield: 28 bits; review packing and truncation */ // representative node
	fProved uint32 /* C bitfield: 1 bits; review packing and truncation */  // marks the proved equivalence
	fFailed uint32 /* C bitfield: 1 bits; review packing and truncation */  // marks the failed equivalence
	fColorA uint32 /* C bitfield: 1 bits; review packing and truncation */  // marks cone of A
	fColorB uint32 /* C bitfield: 1 bits; review packing and truncation */  // marks cone of B
}
type Gia_Plc_t struct {
	fFixed uint32 /* C bitfield: 1 bits; review packing and truncation */  // the placement of this object is fixed
	xCoord uint32 /* C bitfield: 15 bits; review packing and truncation */ // x-ooordinate of the placement
	fUndef uint32 /* C bitfield: 1 bits; review packing and truncation */  // the placement of this object is not assigned
	yCoord uint32 /* C bitfield: 15 bits; review packing and truncation */ // y-ooordinate of the placement
}
