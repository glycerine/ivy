package fraig

// Fraig_NodeAndCanon performs structural hashing followed, when enabled, by
// simulation and SAT-based functional reduction, as in the original FRAIG.
func Fraig_NodeAndCanon(pMan *Fraig_Man_t, p1, p2 *Fraig_Node_t) *Fraig_Node_t {
	if p1 == p2 {
		return p1
	}
	if p1 == Fraig_Not(p2) {
		return Fraig_Not(pMan.pConst1)
	}
	if Fraig_NodeIsConst(p1) != 0 {
		if p1 == pMan.pConst1 {
			return p2
		}
		return Fraig_Not(pMan.pConst1)
	}
	if Fraig_NodeIsConst(p2) != 0 {
		if p2 == pMan.pConst1 {
			return p1
		}
		return Fraig_Not(pMan.pConst1)
	}

	var node *Fraig_Node_t
	if Fraig_HashTableLookupS(pMan, p1, p2, &node) != 0 {
		repr := Fraig_Regular(node).pRepr
		if pMan.fFuncRed != 0 && repr != nil {
			return Fraig_NotCond(repr, Fraig_IsComplement(node) != (Fraig_NodeComparePhase(Fraig_Regular(node), repr) != 0))
		}
		return node
	}
	if pMan.fFuncRed == 0 {
		return node
	}

	sparse := node.nOnes == 0 || node.nOnes == uint32(pMan.nWordsRand)*32
	var old *Fraig_Node_t
	if sparse {
		pMan.nSatZeros++
		if pMan.fDoSparse == 0 {
			return node
		}
		old = Fraig_HashTableLookupF0(pMan, node)
	} else {
		old = Fraig_HashTableLookupF(pMan, node)
	}
	if old == nil {
		return node
	}
	assert(old.pRepr == nil, "old.pRepr == nil")

	useSAT := pMan.nInspLimit == 0 || int64(Fraig_ManReadInspects(pMan)) < pMan.nInspLimit
	if useSAT && Fraig_NodeIsEquivalent(pMan, old, node, pMan.nBTLimit, 1000000) != 0 {
		if pMan.fChoicing != 0 && Fraig_CheckTfi(pMan, old, node) == 0 {
			node.pNextE = old.pNextE
			old.pNextE = node
		}
		node.pRepr = old
		return Fraig_NotCond(old, Fraig_NodeComparePhase(old, node) != 0)
	}
	if sparse {
		assert(pMan.fDoSparse != 0, "pMan.fDoSparse != 0")
		Fraig_HashTableLookupF0(pMan, node)
	} else {
		node.pNextD = old.pNextD
		old.pNextD = node
	}
	assert(node.pRepr == nil, "node.pRepr == nil")
	return node
}
