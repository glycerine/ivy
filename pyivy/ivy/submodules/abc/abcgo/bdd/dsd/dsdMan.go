package dsd

import bdd "github.com/glycerine/abcgo/bdd/cudd/core"

// Dsd_ManagerStart starts the DSD manager. Takes the started BDD manager
// and the maximum support size of the function to be DSD-decomposed. The
// manager should have at least as many variables as there are variables
// in the support. The functions should be expressed using the first
// nSuppSizeMax variables in the manager (these may be ordered not
// necessarily on top of the manager).
func Dsd_ManagerStart(dd *DdManager, nSuppMax, fVerbose int) *Dsd_Manager_t {
	var dMan *Dsd_Manager_t
	var pNode *Dsd_Node_t
	if nSuppMax > len(dd.Vars) {
		panic("dsd: nSuppMax <= dd->size")
	}
	dMan = &Dsd_Manager_t{}
	dMan.dd = dd
	dMan.nInputs = nSuppMax
	dMan.verbose = fVerbose != 0
	dMan.roots = make([]*Dsd_Node_t, 0, 50)
	dMan.inputs = make([]*Dsd_Node_t, dMan.nInputs)
	// create the primary inputs and insert them into the table
	dMan.table = make(map[*DdNode]*Dsd_Node_t)
	for i := 0; i < dMan.nInputs; i++ {
		pNode = Dsd_TreeNodeCreate(int(DSD_NODE_BUF), 1, 0)
		pNode.G = dd.Vars[i]
		bdd.Cudd_Ref(pNode.G)
		pNode.S = dd.Vars[i]
		bdd.Cudd_Ref(pNode.S)
		dMan.tableInsert(dd.Vars[i], pNode)
		dMan.inputs[i] = pNode
	}
	pNode = Dsd_TreeNodeCreate(int(DSD_NODE_CONST1), 0, 0)
	pNode.G = dd.One
	bdd.Cudd_Ref(pNode.G)
	pNode.S = dd.One
	bdd.Cudd_Ref(pNode.S)
	dMan.tableInsert(dd.One, pNode)
	dMan.pConst1 = pNode

	dMan.cache = dsdCheckCacheAllocate(5000)
	return dMan
}

// tableInsert is st__insert into dMan->Table; it returns true if the key
// was already present.
func (dMan *Dsd_Manager_t) tableInsert(bF *DdNode, p *Dsd_Node_t) bool {
	if _, ok := dMan.table[bF]; ok {
		dMan.table[bF] = p
		return true
	}
	dMan.table[bF] = p
	dMan.order = append(dMan.order, p)
	return false
}

// Dsd_ManagerStop stops the DSD manager. Stopping the DSD manager
// automatically dereferences and deallocates all the DSD nodes that were
// created during the life time of the DSD manager.
func Dsd_ManagerStop(dMan *Dsd_Manager_t) {
	if dMan == nil {
		return
	}
	// delete the nodes
	for _, pNode := range dMan.order {
		Dsd_TreeNodeDelete(dMan.dd, Dsd_Regular(pNode))
	}
	dMan.table = nil
	dMan.order = nil
	dMan.inputs = nil
	dMan.roots = nil
	dMan.cache = nil
}
