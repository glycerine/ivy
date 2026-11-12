package reo

// Internal node and level representations from src/bdd/reo/reo.h.
type reo_unit = _reo_unit
type reo_plane = _reo_plane

type _reo_unit struct {
	regular, inverse *reo_unit // explicit complemented edge handles
	origin           *DdNode
	value            float64
	add              bool

	lev       int16 // the level of this node at the beginning
	TopRef    int16 // the top level from which this node is refed (used to update BDD width)
	TopRefNew int16 // the new top level from which this node is refed (used to update BDD width)
	n         int16 // the number of incoming edges (similar to ref count in the BDD)
	Sign      int32 // the signature

	pE     *reo_unit // the pointer to the "else" branch
	pT     *reo_unit // the pointer to the "then" branch
	Next   *reo_unit // the link to the next one in the list
	Weight float64   // the probability of traversing this node
}

type _reo_plane struct {
	fSifted        int32   // to mark the sifted variables
	statsNodes     int32   // the number of nodes in the current level
	statsWidth     int32   // the width on the current level
	statsApl       float64 // the sum of node probabilities on this level
	statsCost      float64 // the current cost is stored here
	statsCostAbove float64 // the current cost is stored here
	statsCostBelow float64 // the current cost is stored here

	pHead *reo_unit // the pointer to the beginning of the unit list
}
