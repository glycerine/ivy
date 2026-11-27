package dsd

import bdd "github.com/glycerine/abcgo/bdd/cudd/core"

// dsdLocal.c: deriving the local function of the DSD node.

// Dsd_TreeGetPrimeFunction returns the local function of the DSD node.
// The local function is computed using the global function of the node and
// the global functions of the formal inputs. The resulting local function is
// mapped using the topmost N variables of the manager. The number of
// variables N is equal to the number of formal inputs.
func Dsd_TreeGetPrimeFunction(dd *DdManager, pNode *Dsd_Node_t) *DdNode {
	var iVar, iLev int
	var bFunc, bRes, bTemp *DdNode

	pNode = Dsd_Regular(pNode)
	size := len(dd.Vars)
	pPermute := make([]int, size)  // C leaves the entries off the support unset
	pVar2Form := make([]int, size) // the mapping of each var into its formal inputs
	pForm2Var := make([]int, size) // the mapping of each formal input into its first var
	for i := range pPermute {
		pPermute[i] = i
	}

	nDecs := len(pNode.pDecs)
	pbCube0 := make([]*DdNode, nDecs)
	pbCube1 := make([]*DdNode, nDecs)

	// remap the global function in such a way that
	// the support variables of each formal input are adjacent
	iLev = 0
	for i := 0; i < nDecs; i++ {
		pForm2Var[i] = dd.InvPerm[i]
		for bTemp = pNode.pDecs[i].S; bTemp != dd.One; bTemp = bTemp.T {
			iVar = dd.InvPerm[iLev]
			pPermute[bTemp.Index] = iVar
			pVar2Form[iVar] = i
			iLev++
		}

		// collect the cubes representing each assignment
		pbCube0[i] = extraBddGetOneCube(dd, bdd.Cudd_Not(pNode.pDecs[i].G))
		bdd.Cudd_Ref(pbCube0[i])
		pbCube1[i] = extraBddGetOneCube(dd, pNode.pDecs[i].G)
		bdd.Cudd_Ref(pbCube1[i])
	}

	// remap the function
	bFunc = bdd.Cudd_bddPermute(dd, pNode.G, pPermute)
	bdd.Cudd_Ref(bFunc)
	// remap the cube
	for i := 0; i < nDecs; i++ {
		bTemp = pbCube0[i]
		pbCube0[i] = bdd.Cudd_bddPermute(dd, bTemp, pPermute)
		bdd.Cudd_Ref(pbCube0[i])
		bdd.Cudd_RecursiveDeref(dd, bTemp)
		bTemp = pbCube1[i]
		pbCube1[i] = bdd.Cudd_bddPermute(dd, bTemp, pPermute)
		bdd.Cudd_Ref(pbCube1[i])
		bdd.Cudd_RecursiveDeref(dd, bTemp)
	}

	// remap the function
	pCache := make(map[*DdNode]*DdNode)
	bRes = extraDsdRemap(dd, bFunc, pCache, pVar2Form, pForm2Var, pbCube0, pbCube1)
	bdd.Cudd_Ref(bRes)

	bdd.Cudd_RecursiveDeref(dd, bFunc)
	for i := 0; i < nDecs; i++ {
		bdd.Cudd_RecursiveDeref(dd, pbCube0[i])
		bdd.Cudd_RecursiveDeref(dd, pbCube1[i])
	}

	bdd.Cudd_Deref(bRes)
	return bRes
}

// extraDsdRemap is Extra_dsdRemap. C caches only nodes whose reference
// count differs from one; caching every node gives the same result.
func extraDsdRemap(dd *DdManager, bF *DdNode, pCache map[*DdNode]*DdNode, pVar2Form, pForm2Var []int, pbCube0, pbCube1 []*DdNode) *DdNode {
	var bF0, bF1, bRes0, bRes1, bRes *DdNode

	bFR := bdd.Cudd_Regular(bF)
	if bdd.Cudd_IsConstant(bFR) {
		return bF
	}

	// check the hash-table
	if r, ok := pCache[bF]; ok {
		return r
	}

	// get the formal input
	iForm := pVar2Form[bFR.Index]

	// get the nodes pointed to by the cube
	bF0 = extraBddNodePointedByCube(dd, bF, pbCube0[iForm])
	bF1 = extraBddNodePointedByCube(dd, bF, pbCube1[iForm])

	// call recursively for these nodes
	bRes0 = extraDsdRemap(dd, bF0, pCache, pVar2Form, pForm2Var, pbCube0, pbCube1)
	bdd.Cudd_Ref(bRes0)
	bRes1 = extraDsdRemap(dd, bF1, pCache, pVar2Form, pForm2Var, pbCube0, pbCube1)
	bdd.Cudd_Ref(bRes1)

	// derive the result using ITE
	bRes = bdd.Cudd_bddIte(dd, dd.Vars[pForm2Var[iForm]], bRes1, bRes0)
	bdd.Cudd_Ref(bRes)
	bdd.Cudd_RecursiveDeref(dd, bRes0)
	bdd.Cudd_RecursiveDeref(dd, bRes1)

	// add to the hash table
	pCache[bF] = bRes
	bdd.Cudd_Deref(bRes)
	return bRes
}

// extraBddNodePointedByCube is Extra_bddNodePointedByCube: the node of bF
// reached by walking along the cube bC.
func extraBddNodePointedByCube(dd *DdManager, bF, bC *DdNode) *DdNode {
	var bF0, bF1, bC0, bC1 *DdNode

	for {
		if bC == dd.LogicZero {
			panic("dsd: Extra_bddNodePointedByCube: bC != b0")
		}
		if bC == dd.One {
			return bF
		}

		bFR := bdd.Cudd_Regular(bF)
		bCR := bdd.Cudd_Regular(bC)
		if bdd.Cudd_IsConstant(bFR) {
			panic("dsd: Extra_bddNodePointedByCube: !cuddIsConstant( bFR )")
		}

		LevelF := dd.Perm[bFR.Index]
		LevelC := dd.Perm[bCR.Index]

		if LevelF <= LevelC {
			if bFR != bF {
				bF0 = bdd.Cudd_Not(bFR.E)
				bF1 = bdd.Cudd_Not(bFR.T)
			} else {
				bF0 = bFR.E
				bF1 = bFR.T
			}
		} else {
			bF0, bF1 = bF, bF
		}

		if LevelC <= LevelF {
			if bCR != bC {
				bC0 = bdd.Cudd_Not(bCR.E)
				bC1 = bdd.Cudd_Not(bCR.T)
			} else {
				bC0 = bCR.E
				bC1 = bCR.T
			}
		} else {
			bC0, bC1 = bC, bC
		}

		if bC0 != dd.LogicZero && bC1 != dd.LogicZero {
			panic("dsd: Extra_bddNodePointedByCube: bC0 == b0 || bC1 == b0")
		}
		if bC0 == dd.LogicZero {
			bF, bC = bF1, bC1
		} else {
			bF, bC = bF0, bC0
		}
	}
}
