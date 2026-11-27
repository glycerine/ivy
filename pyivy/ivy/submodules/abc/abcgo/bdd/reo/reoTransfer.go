package reo

// Port of src/bdd/reo/reoTransfer.c: transferring the DDs between CUDD and REO.

import (
	cudd "github.com/glycerine/abcgo/bdd/cudd/core"
)

// reoTransferNodesToUnits_rec transfers the DD into the internal reordering
// data structure. It is important that the hash table is lossless.
//
// C skips the hash table for nodes with one reference (F->ref == 1), which
// are visited only once; Go hashes every node, which gives the same units
// without depending on the reference counts of the Go CUDD port.
func reoTransferNodesToUnits_rec(p *reo_man, F *DdNode) uref {
	dd := p.dd
	var pUnit uref
	var HKey int

	fComp := cudd.Cudd_IsComplement(F)
	F = cudd.Cudd_Regular(F)

	// check the hash-table
	// search cache - use linear probing
	for HKey = hashKey2(uint64(p.Signature), F.ID, p.tableShift); p.HTable[HKey].Sign == p.Signature; HKey = (HKey + 1) & p.tableMask {
		if p.HTable[HKey].Node == F {
			pUnit = p.HTable[HKey].Arg2
			assert(pUnit != 0, "pUnit")
			// increment the edge counter
			p.U(pUnit).n++
			return Unit_NotCond(pUnit, fComp)
		}
	}
	// the entry in not found in the cache

	// create a new entry
	pUnit = reoUnitsGetNextUnit(p)
	u := p.U(pUnit)
	u.n = 1
	if cudd.Cudd_IsConstant(F) {
		u.lev = REO_CONST_LEVEL
		u.value = int32(int64(F.Value))
		u.pE = 0
		u.pT = 0
		// check if the diagram that is being reordering has complement edges
		if F != dd.One {
			p.fThisIsAdd = 1
		}
		// insert the unit into the corresponding plane
		reoUnitsAddUnitToPlane(p, &p.pPlanes[p.nSupp], pUnit) // increments the unit counter
	} else {
		u.lev = int16(p.pMapToPlanes[F.Index])
		pE := reoTransferNodesToUnits_rec(p, F.E)
		pT := reoTransferNodesToUnits_rec(p, F.T)
		u = p.U(pUnit)
		u.pE = pE
		u.pT = pT
		// insert the unit into the corresponding plane
		reoUnitsAddUnitToPlane(p, &p.pPlanes[u.lev], pUnit) // increments the unit counter
	}

	// add to the hash table
	// the next free entry is already found - it is pointed to by HKey
	// while we traversed the diagram, the hash entry to which HKey points,
	// might have been used. Make sure that its signature is different.
	for ; p.HTable[HKey].Sign == p.Signature; HKey = (HKey + 1) & p.tableMask {
	}
	p.HTable[HKey].Sign = p.Signature
	p.HTable[HKey].Node = F
	p.HTable[HKey].Arg2 = pUnit

	// increment the counter of nodes
	p.nNodesCur++
	return Unit_NotCond(pUnit, fComp)
}

// reoTransferUnitsToNodes_rec creates the DD from the internal reordering
// data structure. It is important that the hash table is lossless.
func reoTransferUnitsToNodes_rec(p *reo_man, pUnitC uref) *DdNode {
	dd := p.dd
	var bRes, E, T *DdNode
	HKey := -1

	fComp := Unit_IsComplement(pUnitC)
	pUnitR := Unit_Regular(pUnitC)
	pUnit := p.U(pUnitR)

	// check the hash-table
	if pUnit.n != 1 {
		for HKey = hashKey2(uint64(p.Signature), uint64(pUnitR), p.tableShift); p.HTable[HKey].Sign == p.Signature; HKey = (HKey + 1) & p.tableMask {
			if p.HTable[HKey].Arg1 == pUnitR {
				bRes = p.HTable[HKey].Node
				assert(bRes != nil, "bRes")
				return cudd.Cudd_NotCond(bRes, fComp)
			}
		}
	}
	// treat the case of constants
	if Unit_IsConstant(pUnit) {
		bRes = cudd.Cudd_addConst(dd, float64(pUnit.value))
		cudd.Cudd_Ref(bRes)
	} else {
		// split and recur on children of this node
		E = reoTransferUnitsToNodes_rec(p, pUnit.pE)
		if E == nil {
			return nil
		}
		cudd.Cudd_Ref(E)
		T = reoTransferUnitsToNodes_rec(p, pUnit.pT)
		if T == nil {
			cudd.Cudd_RecursiveDeref(dd, E)
			return nil
		}
		cudd.Cudd_Ref(T)

		// consider the case when Res0 and Res1 are the same node
		assert(E != T, "E != T")
		assert(!cudd.Cudd_IsComplement(T), "!Cudd_IsComplement(T)")

		bRes = cudd.Cudd_UniqueInter(dd, p.pMapToDdVarsFinal[pUnit.lev], T, E)
		if bRes == nil {
			cudd.Cudd_RecursiveDeref(dd, E)
			cudd.Cudd_RecursiveDeref(dd, T)
			return nil
		}
		cudd.Cudd_Ref(bRes)
		cudd.Cudd_Deref(E)
		cudd.Cudd_Deref(T)
	}

	// do not keep the result if the ref count is only 1, since it will not be visited again
	if pUnit.n != 1 {
		// while we traversed the diagram, the hash entry to which HKey points,
		// might have been used. Make sure that its signature is different.
		for ; p.HTable[HKey].Sign == p.Signature; HKey = (HKey + 1) & p.tableMask {
		}
		p.HTable[HKey].Sign = p.Signature
		p.HTable[HKey].Arg1 = pUnitR
		p.HTable[HKey].Node = bRes

		// add the DD to the referenced DD list in order to be able to store it in cache
		p.pRefNodes[p.nRefNodes] = bRes
		p.nRefNodes++
		cudd.Cudd_Ref(bRes)
		// no need to do this, because the garbage collection will not take bRes away
		// it is held by the diagram in the making
	}
	// increment the counter of nodes
	p.nNodesCur++
	cudd.Cudd_Deref(bRes)
	return cudd.Cudd_NotCond(bRes, fComp)
}
