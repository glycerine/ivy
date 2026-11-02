package abc

// Hand-written skeleton of the type in src/proof/pdr/pdrTsim3.c.

// Txs3Man is Txs3_Man_t.
type Txs3Man struct {
	pGia    *GiaManager // user's AIG
	vPrio   *IntVector  // priority of each flop
	vCiObjs *IntVector  // cone leaves (CI obj IDs)
	vFosPre *IntVector  // cone leaves (CI obj IDs)
	vFosAbs *IntVector  // cone leaves (CI obj IDs)
	vCoObjs *IntVector  // cone roots (CO obj IDs)
	vCiVals *IntVector  // cone leaf values (0/1 CI values)
	vCoVals *IntVector  // cone root values (0/1 CO values)
	vNodes  *IntVector  // cone nodes (node obj IDs)
	vTemp   *IntVector  // cone nodes (node obj IDs)
	vPiLits *IntVector  // resulting array of PI literals
	vFfLits *IntVector  // resulting array of flop literals
	pMan    *PdrMan     // calling manager
	nPiLits int         // the number of PI literals
}
