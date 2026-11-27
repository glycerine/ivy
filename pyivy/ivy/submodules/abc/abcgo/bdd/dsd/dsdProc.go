package dsd

import (
	"fmt"
	"os"
	"time"

	bdd "github.com/glycerine/abcgo/bdd/cudd/core"
)

// dsdProc.c: the core procedures of the package.

// Dsd_Decompose performs DSD for the array of functions represented by
// BDDs. The resulting DSD tree is stored in the DSD manager (roots).
// The resulting tree is a shared DSD DAG for the functions given in the
// array. The root node pointers can be complemented. All the decomposition
// trees and intermediate nodes will be removed when the DSD manager is
// deallocated at the end by calling Dsd_ManagerStop().
func Dsd_Decompose(pDsdMan *Dsd_Manager_t, pbFuncs []*DdNode, nFuncs int) {
	dd := pDsdMan.dd
	var pTemp *Dsd_Node_t
	SumMaxGateSize := 0
	nDecOutputs := 0
	nCBFOutputs := 0
	stdout := os.Stdout

	if pDsdMan.verbose {
		fmt.Fprintf(stdout, "\nDecomposability statistics for individual outputs:\n")
	}

	// set the counter of decomposition nodes
	pDsdMan.sNDecBlocks = 0

	// perform decomposition for all outputs
	clk := time.Now()
	pDsdMan.roots = pDsdMan.roots[:0]
	pDsdMan.sNCascades = 0
	for i := 0; i < nFuncs; i++ {
		var nCascades, MaxBlock int

		clk := time.Now()
		nReusedBlocksPres := pDsdMan.sNReusedBlocks

		pDsdMan.roots = append(pDsdMan.roots, dsdKernelDecompose_rec(pDsdMan, pbFuncs[i]))

		Dsd_TreeNodeGetInfoOne(pDsdMan.roots[i], &nCascades, &MaxBlock)
		pDsdMan.sNCascades = max(pDsdMan.sNCascades, nCascades)
		pTemp = Dsd_Regular(pDsdMan.roots[i])
		if pTemp.Type != DSD_NODE_PRIME || len(pTemp.pDecs) != extraBddSuppSize(dd, pTemp.S) {
			nDecOutputs++
		}
		if MaxBlock < 3 {
			nCBFOutputs++
		}
		SumMaxGateSize += MaxBlock

		if pDsdMan.verbose {
			fmt.Fprintf(stdout, "#%02d: ", i)
			fmt.Fprintf(stdout, "Ins=%2d. ", cuddSupportSize(dd, pbFuncs[i]))
			fmt.Fprintf(stdout, "Gts=%3d. ", Dsd_TreeCountNonTerminalNodesOne(pDsdMan.roots[i]))
			fmt.Fprintf(stdout, "Pri=%3d. ", Dsd_TreeCountPrimeNodesOne(pDsdMan.roots[i]))
			fmt.Fprintf(stdout, "Max=%3d. ", MaxBlock)
			fmt.Fprintf(stdout, "Reuse=%2d. ", pDsdMan.sNReusedBlocks-nReusedBlocksPres)
			fmt.Fprintf(stdout, "Csc=%2d. ", nCascades)
			fmt.Fprintf(stdout, "T= %.2f s. ", time.Since(clk).Seconds())
			fmt.Fprintf(stdout, "Bdd=%2d. ", cuddSharingSize(pbFuncs[i:i+1]))
			fmt.Fprintf(stdout, "\n")
		}
	}
	if len(pDsdMan.roots) != nFuncs {
		panic("dsd: Dsd_Decompose: pDsdMan->nRoots == nFuncs")
	}

	if pDsdMan.verbose {
		fmt.Fprintf(stdout, "\n")
		fmt.Fprintf(stdout, "The cumulative decomposability statistics:\n")
		fmt.Fprintf(stdout, "  Total outputs                             = %5d\n", nFuncs)
		fmt.Fprintf(stdout, "  Decomposable outputs                      = %5d\n", nDecOutputs)
		fmt.Fprintf(stdout, "  Completely decomposable outputs           = %5d\n", nCBFOutputs)
		fmt.Fprintf(stdout, "  The sum of max gate sizes                 = %5d\n", SumMaxGateSize)
		fmt.Fprintf(stdout, "  Shared BDD size                           = %5d\n", cuddSharingSize(pbFuncs[:nFuncs]))
		fmt.Fprintf(stdout, "  Decomposition entries                     = %5d\n", len(pDsdMan.table))
		fmt.Fprintf(stdout, "  Pure decomposition time                   =  %.2f sec\n", time.Since(clk).Seconds())
	}
}

// Dsd_DecomposeOne performs decomposition for one function.
func Dsd_DecomposeOne(pDsdMan *Dsd_Manager_t, bFunc *DdNode) *Dsd_Node_t {
	return dsdKernelDecompose_rec(pDsdMan, bFunc)
}

// dsdKernelDecompose_rec is the main function of this module: the
// recursive implementation of DSD.
func dsdKernelDecompose_rec(pDsdMan *Dsd_Manager_t, bFunc0 *DdNode) *Dsd_Node_t {
	dd := pDsdMan.dd
	b0, b1 := dd.LogicZero, dd.One
	var bLow, bLowR, bHigh *DdNode

	var VarInt uint32
	var bVarCur *DdNode
	var pVarCurDE *Dsd_Node_t
	// works only if var indices start from 0!!!
	var bSuppNew, bTemp *DdNode

	var fContained bool
	var nSuppLH, nSuppL, nSuppH int

	// various decomposition nodes
	var pThis, pL, pH, pLR, pHR *Dsd_Node_t

	var pSmallR, pLargeR *Dsd_Node_t

	// treat the complemented case
	bF := bdd.Cudd_Regular(bFunc0)
	fCompF := bF != bFunc0

	// check cache
	if pTableEntry, ok := pDsdMan.table[bF]; ok { // the entry is present
		return Dsd_NotCond(pTableEntry, fCompF)
	}

	// proceed to consider "four cases"
	//////////////////////////////////////////////////////////////////////
	// TERMINAL CASES - CASES 1 and 2
	//////////////////////////////////////////////////////////////////////
	bLow = bF.E
	bLowR = bdd.Cudd_Regular(bLow)
	bHigh = bF.T
	VarInt = bF.Index
	bVarCur = dd.Vars[VarInt]
	pVarCurDE = pDsdMan.inputs[VarInt]
	// works only if var indices start from 0!!!
	bSuppNew = nil

	if bLowR.Index == bdd.CUDD_CONST_INDEX || bHigh.Index == bdd.CUDD_CONST_INDEX { // one of the cofactors in the constant
		if bHigh == b1 { // bHigh cannot be equal to b0, because then it will be complemented
			if bLow == b0 { // bLow cannot be equal to b1, because then the node will have bLow == bHigh
				/////////////////////////////////////////////////////////////////
				// bLow == 0, bHigh == 1, F = x'&0 + x&1 = x
				/////////////////////////////////////////////////////////////////
				// create the elementary variable node
				panic("dsd: dsdKernelDecompose_rec: the variable should be already in the hash table")
			} else { // if ( bLow != constant )
				/////////////////////////////////////////////////////////////////
				// bLow != const, bHigh == 1, F = x'&bLow + x&1 = bLow + x  --- DSD_NODE_OR(x,bLow)
				/////////////////////////////////////////////////////////////////
				pL = dsdKernelDecompose_rec(pDsdMan, bLow)
				pLR = Dsd_Regular(pL)
				bSuppNew = bdd.Cudd_bddAnd(dd, bVarCur, pLR.S)
				bdd.Cudd_Ref(bSuppNew)
				if pLR.Type == DSD_NODE_OR && pL == pLR { // OR and no complement
					// add to the components
					pThis = Dsd_TreeNodeCreate(int(DSD_NODE_OR), len(pL.pDecs)+1, pDsdMan.nextBlock())
					dsdKernelCopyListPlusOne(pThis, pVarCurDE, pL.pDecs, len(pL.pDecs))
				} else { // all other cases
					// create a new 2-input OR-gate
					pThis = Dsd_TreeNodeCreate(int(DSD_NODE_OR), 2, pDsdMan.nextBlock())
					dsdKernelCopyListPlusOne(pThis, pVarCurDE, []*Dsd_Node_t{pL}, 1)
				}
			}
		} else { // if ( bHigh != const ) // meaning that bLow should be a constant
			pH = dsdKernelDecompose_rec(pDsdMan, bHigh)
			pHR = Dsd_Regular(pH)
			bSuppNew = bdd.Cudd_bddAnd(dd, bVarCur, pHR.S)
			bdd.Cudd_Ref(bSuppNew)
			if bLow == b0 {
				/////////////////////////////////////////////////////////////////
				// Low == 0, High != 1, F = x'&0+x&High = (x'+High')'--- NOR(x',High')
				/////////////////////////////////////////////////////////////////
				if pHR.Type == DSD_NODE_OR && pH != pHR { // DSD_NODE_OR and complement
					// add to the components
					pThis = Dsd_TreeNodeCreate(int(DSD_NODE_OR), len(pHR.pDecs)+1, pDsdMan.nextBlock())
					dsdKernelCopyListPlusOne(pThis, Dsd_Not(pVarCurDE), pHR.pDecs, len(pHR.pDecs))
					pThis = Dsd_Not(pThis)
				} else { // all other cases
					// create a new 2-input NOR gate
					pThis = Dsd_TreeNodeCreate(int(DSD_NODE_OR), 2, pDsdMan.nextBlock())
					pH = Dsd_Not(pH)
					dsdKernelCopyListPlusOne(pThis, Dsd_Not(pVarCurDE), []*Dsd_Node_t{pH}, 1)
					pThis = Dsd_Not(pThis)
				}
			} else { // if ( bLow == b1 )
				/////////////////////////////////////////////////////////////////
				// Low == 1, High != 1, F = x'&1 + x&High = x' + High --- DSD_NODE_OR(x',High)
				/////////////////////////////////////////////////////////////////
				if pHR.Type == DSD_NODE_OR && pH == pHR { // OR and no complement
					// add to the components
					pThis = Dsd_TreeNodeCreate(int(DSD_NODE_OR), len(pH.pDecs)+1, pDsdMan.nextBlock())
					dsdKernelCopyListPlusOne(pThis, Dsd_Not(pVarCurDE), pH.pDecs, len(pH.pDecs))
				} else { // all other cases
					// create a new 2-input OR-gate
					pThis = Dsd_TreeNodeCreate(int(DSD_NODE_OR), 2, pDsdMan.nextBlock())
					dsdKernelCopyListPlusOne(pThis, Dsd_Not(pVarCurDE), []*Dsd_Node_t{pH}, 1)
				}
			}
		}
		return dsdKernelExit(pDsdMan, pThis, bF, bSuppNew, fCompF)
	}
	// else if ( bLow != const && bHigh != const )

	// the case of equal cofactors (up to complementation)
	if bLowR == bHigh {
		/////////////////////////////////////////////////////////////////
		// Low == G, High == G', F = x'&G + x&G' = (x(+)G) --- EXOR(x,Low)
		/////////////////////////////////////////////////////////////////
		pL = dsdKernelDecompose_rec(pDsdMan, bLow)
		pLR = Dsd_Regular(pL)
		bSuppNew = bdd.Cudd_bddAnd(dd, bVarCur, pLR.S)
		bdd.Cudd_Ref(bSuppNew)
		if pLR.Type == DSD_NODE_EXOR { // complemented or not - does not matter!
			// add to the components
			pThis = Dsd_TreeNodeCreate(int(DSD_NODE_EXOR), len(pLR.pDecs)+1, pDsdMan.nextBlock())
			dsdKernelCopyListPlusOne(pThis, pVarCurDE, pLR.pDecs, len(pLR.pDecs))
			if pL != pLR {
				pThis = Dsd_Not(pThis)
			}
		} else { // all other cases
			// create a new 2-input EXOR-gate
			pThis = Dsd_TreeNodeCreate(int(DSD_NODE_EXOR), 2, pDsdMan.nextBlock())
			if pL != pLR { // complemented
				dsdKernelCopyListPlusOne(pThis, pVarCurDE, []*Dsd_Node_t{pLR}, 1)
				pThis = Dsd_Not(pThis)
			} else { // non-complemented
				dsdKernelCopyListPlusOne(pThis, pVarCurDE, []*Dsd_Node_t{pL}, 1)
			}
		}
		return dsdKernelExit(pDsdMan, pThis, bF, bSuppNew, fCompF)
	}

	//////////////////////////////////////////////////////////////////////
	// solve subproblems
	//////////////////////////////////////////////////////////////////////
	pL = dsdKernelDecompose_rec(pDsdMan, bLow)
	pH = dsdKernelDecompose_rec(pDsdMan, bHigh)
	pLR = Dsd_Regular(pL)
	pHR = Dsd_Regular(pH)

	if !(pLR.Type == DSD_NODE_BUF || pLR.Type == DSD_NODE_OR || pLR.Type == DSD_NODE_EXOR || pLR.Type == DSD_NODE_PRIME) ||
		!(pHR.Type == DSD_NODE_BUF || pHR.Type == DSD_NODE_OR || pHR.Type == DSD_NODE_EXOR || pHR.Type == DSD_NODE_PRIME) {
		panic("dsd: dsdKernelDecompose_rec: bad cofactor component type")
	}

	// compute the new support
	bTemp = bdd.Cudd_bddAnd(dd, pLR.S, pHR.S)
	bdd.Cudd_Ref(bTemp)
	nSuppL = extraBddSuppSize(dd, pLR.S)
	nSuppH = extraBddSuppSize(dd, pHR.S)
	nSuppLH = extraBddSuppSize(dd, bTemp)
	bSuppNew = bdd.Cudd_bddAnd(dd, bTemp, bVarCur)
	bdd.Cudd_Ref(bSuppNew)
	bdd.Cudd_RecursiveDeref(dd, bTemp)

	// several possibilities are possible
	// (1) support of one component contains another
	// (2) none of the supports is contained in another
	fContained = dsdKernelCheckContainment(pDsdMan, pLR, pHR, &pLargeR, &pSmallR)

	//////////////////////////////////////////////////////////////////////
	// CASE 3.b One of the cofactors in a constant (OR and EXOR)
	//////////////////////////////////////////////////////////////////////
	// the support of the larger component should contain the support of the smaller
	// it is possible to have PRIME function in this role
	// for example: F = ITE( a+b, c(+)d, e+f ), F0 = ITE( b, c(+)d, e+f ), F1 = c(+)d
	if fContained {
		var pSmall, pLarge *Dsd_Node_t
		iCompLarge := -1 // the number of the component is Large is equal to the whole of Small
		var fLowIsLarge bool

		var bFTemp *DdNode // the changed input function
		var pDETemp, pDENew *Dsd_Node_t

		var pComp *Dsd_Node_t
		nComp := -1

		if pSmallR == pLR { // Low is Small => High is Large
			pSmall = pL
			pLarge = pH
			fLowIsLarge = false
		} else { // vice versa
			pSmall = pH
			pLarge = pL
			fLowIsLarge = true
		}

		// treat the situation when the larger is PRIME
		if pLargeR.Type == DSD_NODE_PRIME { //&& pLargeR->nDecs != pSmallR->nDecs )
			// find the component, which when substituted for 0 or 1, produces the desired result
			var g int
			fFoundComp := -1 // {0,1} depending on whether setting cofactor to 0 or 1 worked out

			var bLarge, bSmall *DdNode
			if fLowIsLarge {
				bLarge = bLow
				bSmall = bHigh
			} else {
				bLarge = bHigh
				bSmall = bLow
			}

			for g = 0; g < len(pLargeR.pDecs); g++ {
				pDETemp = pLargeR.pDecs[g] // cannot be complemented
				if pDsdMan.checkRootFunctionIdentity(bLarge, bSmall, pDETemp.G, b1) != 0 {
					fFoundComp = 1
					break
				}

				if pDsdMan.checkRootFunctionIdentity(bLarge, bSmall, bdd.Cudd_Not(pDETemp.G), b1) != 0 {
					fFoundComp = 0
					break
				}
			}

			if g != len(pLargeR.pDecs) { // decomposition is found
				if fFoundComp != 0 {
					if fLowIsLarge {
						bFTemp = bdd.Cudd_bddOr(dd, bVarCur, pLargeR.pDecs[g].G)
					} else {
						bFTemp = bdd.Cudd_bddOr(dd, bdd.Cudd_Not(bVarCur), pLargeR.pDecs[g].G)
					}
				} else {
					if fLowIsLarge {
						bFTemp = bdd.Cudd_bddAnd(dd, bdd.Cudd_Not(bVarCur), pLargeR.pDecs[g].G)
					} else {
						bFTemp = bdd.Cudd_bddAnd(dd, bVarCur, pLargeR.pDecs[g].G)
					}
				}
				bdd.Cudd_Ref(bFTemp)

				pDENew = dsdKernelDecompose_rec(pDsdMan, bFTemp)
				pDENew = Dsd_Regular(pDENew)
				bdd.Cudd_RecursiveDeref(dd, bFTemp)

				// get the new gate
				pThis = Dsd_TreeNodeCreate(int(DSD_NODE_PRIME), len(pLargeR.pDecs), pDsdMan.nextBlock())
				dsdKernelCopyListPlusOneMinusOne(pThis, pDENew, pLargeR.pDecs, len(pLargeR.pDecs), g)
				return dsdKernelExit(pDsdMan, pThis, bF, bSuppNew, fCompF)
			}
		}

		// try to find one component in the pLarger that is equal to the whole of pSmaller
		c := 0
		for c = 0; c < len(pLargeR.pDecs); c++ {
			if pLargeR.pDecs[c] == pSmall || pLargeR.pDecs[c] == Dsd_Not(pSmall) {
				iCompLarge = c
				break
			}
		}

		// assign the equal component
		if c != len(pLargeR.pDecs) { // the decomposition is possible!
			pComp = pLargeR.pDecs[iCompLarge]
			nComp = 1
		} else { // the decomposition is still possible
			// for example F = OR(ab,c,d), F(a=0) = OR(c,d), F(a=1) = OR(b,c,d)
			// supp(F0) is contained in supp(F1), Polarity(F(a=0)) == Polarity(F(a=1))

			// try to find a group of common components
			if pLargeR.Type == pSmallR.Type &&
				(pLargeR.Type == DSD_NODE_EXOR || (pSmallR.Type == DSD_NODE_OR && ((pLarge == pLargeR) == (pSmall == pSmallR)))) {
				var pLastDiffL, pLastDiffH *Dsd_Node_t
				pCommon, nCommon := dsdKernelFindCommonComponents(pDsdMan, pLargeR, pSmallR, &pLastDiffL, &pLastDiffH)
				_ = pCommon
				// if all the components of pSmall are contained in pLarge,
				// then the decomposition exists
				if nCommon == len(pSmallR.pDecs) {
					pComp = pSmallR
					nComp = len(pSmallR.pDecs)
				}
			}
		}

		if pComp != nil { // the decomposition is possible!
			pCompR := Dsd_Regular(pComp)
			fComp1 := pLarge != pLargeR
			fComp2 := pComp != pCompR
			fComp3 := pSmall != pSmallR

			var bFuncComp *DdNode // the function of the given component
			var bFuncNew *DdNode  // the function of the input component

			if pLargeR.Type == DSD_NODE_OR { // Figure 4 of Matsunaga's paper
				// the decomposition exists only if the polarity assignment
				// along the paths is the same
				if (fComp1 != fComp2) == fComp3 { // decomposition exists = consider 4 cases
					// consideration of cases leads to the following conclusion
					// fComp1 gives the polarity of the resulting DSD_NODE_OR gate
					// fComp2 gives the polarity of the common component feeding into the DSD_NODE_OR gate

					// at this point we have the function F (bFTemp) and the common component C (bFuncComp)
					// to get the remainder, R, in the relationship F = R + C, supp(R) & supp(C) = 0
					// we compute the following R = Exist( F - C, supp(C) )
					if fComp1 {
						bFTemp = bdd.Cudd_Not(bF)
					} else {
						bFTemp = bF
					}
					if fComp2 {
						bFuncComp = bdd.Cudd_Not(pCompR.G)
					} else {
						bFuncComp = pCompR.G
					}
					bFuncNew = bdd.Cudd_bddAndAbstract(dd, bFTemp, bdd.Cudd_Not(bFuncComp), pCompR.S)
					bdd.Cudd_Ref(bFuncNew)

					// there is no need to copy the dec entry list first, because pComp is a component
					// which will not be destroyed by the recursive call to decomposition
					pDENew = dsdKernelDecompose_rec(pDsdMan, bFuncNew)
					if !Dsd_IsComplement(pDENew) { // follows from the consideration of cases
						panic("dsd: dsdKernelDecompose_rec: Dsd_IsComplement(pDENew)")
					}
					bdd.Cudd_RecursiveDeref(dd, bFuncNew)

					// get the new gate
					if nComp == 1 {
						pThis = Dsd_TreeNodeCreate(int(DSD_NODE_OR), 2, pDsdMan.nextBlock())
						pThis.pDecs[0] = pDENew
						pThis.pDecs[1] = pComp // takes the complement
					} else { // pComp is not complemented
						pThis = Dsd_TreeNodeCreate(int(DSD_NODE_OR), nComp+1, pDsdMan.nextBlock())
						dsdKernelCopyListPlusOne(pThis, pDENew, pComp.pDecs, nComp)
					}

					if fComp1 {
						pThis = Dsd_Not(pThis)
					}
					return dsdKernelExit(pDsdMan, pThis, bF, bSuppNew, fCompF)
				}
			} else if pLargeR.Type == DSD_NODE_EXOR { // Figure 5 of Matsunaga's paper (with correction)
				// decomposition always exists = consider 4 cases

				// consideration of cases leads to the following conclusion
				// fComp3 gives the COMPLEMENT of the polarity of the resulting EXOR gate
				// (if fComp3 is 0, the EXOR gate is complemented, and vice versa)
				if fComp2 {
					panic("dsd: dsdKernelDecompose_rec: fComp2 == 0")
				}
				// find the functionality of the lower gates
				if fComp3 {
					bFTemp = bF
				} else {
					bFTemp = bdd.Cudd_Not(bF)
				}
				bFuncNew = bdd.Cudd_bddXor(dd, bFTemp, pComp.G)
				bdd.Cudd_Ref(bFuncNew)

				pDENew = dsdKernelDecompose_rec(pDsdMan, bFuncNew)
				if Dsd_IsComplement(pDENew) { // follows from the consideration of cases
					panic("dsd: dsdKernelDecompose_rec: !Dsd_IsComplement(pDENew)")
				}
				bdd.Cudd_RecursiveDeref(dd, bFuncNew)

				// get the new gate
				if nComp == 1 {
					pThis = Dsd_TreeNodeCreate(int(DSD_NODE_EXOR), 2, pDsdMan.nextBlock())
					pThis.pDecs[0] = pDENew
					pThis.pDecs[1] = pComp
				} else { // pComp is not complemented
					pThis = Dsd_TreeNodeCreate(int(DSD_NODE_EXOR), nComp+1, pDsdMan.nextBlock())
					dsdKernelCopyListPlusOne(pThis, pDENew, pComp.pDecs, nComp)
				}

				if !fComp3 {
					pThis = Dsd_Not(pThis)
				}
				return dsdKernelExit(pDsdMan, pThis, bF, bSuppNew, fCompF)
			}
		}
	}

	// this case was added to fix the trivial bug found November 4, 2002 in Japan
	// by running the example provided by T. Sasao
	if nSuppLH == nSuppL+nSuppH { // the supports of the components are disjoint
		// create a new component of the type ITE( a, pH, pL )
		pThis = Dsd_TreeNodeCreate(int(DSD_NODE_PRIME), 3, pDsdMan.nextBlock())
		if dd.Perm[pLR.S.Index] < dd.Perm[pHR.S.Index] { // pLR is higher in the varible order
			pThis.pDecs[1] = pLR
			pThis.pDecs[2] = pHR
		} else { // pHR is higher in the varible order
			pThis.pDecs[1] = pHR
			pThis.pDecs[2] = pLR
		}
		// add the first component
		pThis.pDecs[0] = pVarCurDE
		return dsdKernelExit(pDsdMan, pThis, bF, bSuppNew, fCompF)
	}

	//////////////////////////////////////////////////////////////////////
	// CASE 3.a Neither of the cofactors is a constant (OR, EXOR, PRIME)
	//////////////////////////////////////////////////////////////////////
	// the component types are identical
	// and if they are OR, they are either both complemented or both not complemented
	// and if they are PRIME, their dec numbers should be the same
	if pLR.Type == pHR.Type &&
		pLR.Type != DSD_NODE_BUF &&
		(pLR.Type != DSD_NODE_OR || ((pL == pLR && pH == pHR) || (pL != pLR && pH != pHR))) &&
		(pLR.Type != DSD_NODE_PRIME || len(pLR.pDecs) == len(pHR.pDecs)) {
		// array to store common comps in pL and pH
		var pLastDiffL, pLastDiffH *Dsd_Node_t
		pCommon, nCommon := dsdKernelFindCommonComponents(pDsdMan, pLR, pHR, &pLastDiffL, &pLastDiffH)
		if nCommon != 0 {
			if pLR.Type == DSD_NODE_OR { // Figure 2 of Matsunaga's paper
				// at this point we have the function F and the group of common components C
				// to get the remainder, R, in the relationship F = R + C, supp(R) & supp(C) = 0
				// we compute the following R = Exist( F - C, supp(C) )

				// compute the sum total of the common components and the union of their supports
				var bCommF, bCommS, bFTemp, bFuncNew *DdNode
				var pDENew *Dsd_Node_t

				dsdKernelComputeSumOfComponents(pDsdMan, pCommon, nCommon, &bCommF, &bCommS, false)
				bdd.Cudd_Ref(bCommF)
				bdd.Cudd_Ref(bCommS)
				if pL != pLR {
					bFTemp = bdd.Cudd_Not(bF)
				} else {
					bFTemp = bF
				}

				bFuncNew = bdd.Cudd_bddAndAbstract(dd, bFTemp, bdd.Cudd_Not(bCommF), bCommS)
				bdd.Cudd_Ref(bFuncNew)
				bdd.Cudd_RecursiveDeref(dd, bCommF)
				bdd.Cudd_RecursiveDeref(dd, bCommS)

				// get the new gate

				// copy the components first, then call the decomposition
				// because decomposition will distroy the list used for copying
				pThis = Dsd_TreeNodeCreate(int(DSD_NODE_OR), nCommon+1, pDsdMan.nextBlock())
				dsdKernelCopyListPlusOne(pThis, nil, pCommon, nCommon)

				// call the decomposition recursively
				pDENew = dsdKernelDecompose_rec(pDsdMan, bFuncNew)
				bdd.Cudd_RecursiveDeref(dd, bFuncNew)

				// add the first component
				pThis.pDecs[0] = pDENew

				if pL != pLR {
					pThis = Dsd_Not(pThis)
				}
				return dsdKernelExit(pDsdMan, pThis, bF, bSuppNew, fCompF)
			} else if pLR.Type == DSD_NODE_EXOR { // Figure 3 of Matsunaga's paper
				// compute the sum total of the common components and the union of their supports
				var bCommF, bFuncNew *DdNode
				var pDENew *Dsd_Node_t

				dsdKernelComputeSumOfComponents(pDsdMan, pCommon, nCommon, &bCommF, nil, true)
				bdd.Cudd_Ref(bCommF)

				bFuncNew = bdd.Cudd_bddXor(dd, bF, bCommF)
				bdd.Cudd_Ref(bFuncNew)
				bdd.Cudd_RecursiveDeref(dd, bCommF)

				// get the new gate

				// copy the components first, then call the decomposition
				// because decomposition will distroy the list used for copying
				pThis = Dsd_TreeNodeCreate(int(DSD_NODE_EXOR), nCommon+1, pDsdMan.nextBlock())
				dsdKernelCopyListPlusOne(pThis, nil, pCommon, nCommon)

				// call the decomposition recursively
				pDENew = dsdKernelDecompose_rec(pDsdMan, bFuncNew)
				bdd.Cudd_RecursiveDeref(dd, bFuncNew)

				// remember the fact that it was complemented
				fCompExor := Dsd_IsComplement(pDENew)
				pDENew = Dsd_Regular(pDENew)

				// add the first component
				pThis.pDecs[0] = pDENew

				if fCompExor {
					pThis = Dsd_Not(pThis)
				}
				return dsdKernelExit(pDsdMan, pThis, bF, bSuppNew, fCompF)
			} else if pLR.Type == DSD_NODE_PRIME && (nCommon == len(pLR.pDecs)-1 || nCommon == len(pLR.pDecs)) {
				// for example the function F(a,b,c,d) = ITE(b,c,a(+)d) produces
				// two cofactors F(a=0) = PRIME(b,c,d) and F(a=1) = PRIME(b,c,d)
				// with exactly the same list of common components

				var pDENew *Dsd_Node_t
				var bFuncNew *DdNode
				fCompComp := 0 // this flag can be {0,1,2}
				// if it is 0 there is no identity
				// if it is 1/2, the cofactored functions are equal in the direct/complemented polarity

				if nCommon == len(pLR.pDecs) { // all the components are the same
					// find the formal input, in which pLow and pHigh differ (if such input exists)
					var pTempL, pTempH *Dsd_Node_t

					for m := 0; m < len(pLR.pDecs); m++ {
						pTempL = pLR.pDecs[m] // cannot be complemented
						pTempH = pHR.pDecs[m] // cannot be complemented

						if pDsdMan.checkRootFunctionIdentity(bLow, bHigh, pTempL.G, bdd.Cudd_Not(pTempH.G)) != 0 &&
							pDsdMan.checkRootFunctionIdentity(bLow, bHigh, bdd.Cudd_Not(pTempL.G), pTempH.G) != 0 {
							pLastDiffL = pTempL
							pLastDiffH = pTempH
							if pLastDiffL != pLastDiffH {
								panic("dsd: dsdKernelDecompose_rec: pLastDiffL == pLastDiffH")
							}
							fCompComp = 2
							break
						}
					}

					if fCompComp != 0 { // put the equal components into pCommon, so that they could be copied into the new dec entry
						nCommon = 0
						for m := 0; m < len(pLR.pDecs); m++ {
							if pLR.pDecs[m] != pLastDiffL {
								pCommon[nCommon] = pLR.pDecs[m]
								nCommon++
							}
						}
						if nCommon != len(pLR.pDecs)-1 {
							panic("dsd: dsdKernelDecompose_rec: nCommon == pLR->nDecs-1")
						}
					}
				} else { // the differing components are known - check that they have compatible PRIME function
					// find the numbers of different components
					if pLastDiffL == nil || pLastDiffH == nil {
						panic("dsd: dsdKernelDecompose_rec: pLastDiffL && pLastDiffH")
					}
					// also, they cannot be complemented, because the decomposition type is PRIME

					if pDsdMan.checkRootFunctionIdentity(bLow, bHigh, bdd.Cudd_Not(pLastDiffL.G), bdd.Cudd_Not(pLastDiffH.G)) != 0 &&
						pDsdMan.checkRootFunctionIdentity(bLow, bHigh, pLastDiffL.G, pLastDiffH.G) != 0 {
						fCompComp = 1
					} else if pDsdMan.checkRootFunctionIdentity(bLow, bHigh, pLastDiffL.G, bdd.Cudd_Not(pLastDiffH.G)) != 0 &&
						pDsdMan.checkRootFunctionIdentity(bLow, bHigh, bdd.Cudd_Not(pLastDiffL.G), pLastDiffH.G) != 0 {
						fCompComp = 2
					}
				}

				if fCompComp != 0 {
					if fCompComp == 1 { // it is true that bLow(G=0) == bHigh(H=0) && bLow(G=1) == bHigh(H=1)
						bFuncNew = bdd.Cudd_bddIte(dd, bVarCur, pLastDiffH.G, pLastDiffL.G)
					} else { // it is true that bLow(G=0) == bHigh(H=1) && bLow(G=1) == bHigh(H=0)
						bFuncNew = bdd.Cudd_bddIte(dd, bVarCur, bdd.Cudd_Not(pLastDiffH.G), pLastDiffL.G)
					}
					bdd.Cudd_Ref(bFuncNew)

					// get the new gate

					// copy the components first, then call the decomposition
					// because decomposition will distroy the list used for copying
					pThis = Dsd_TreeNodeCreate(int(DSD_NODE_PRIME), len(pLR.pDecs), pDsdMan.nextBlock())
					dsdKernelCopyListPlusOne(pThis, nil, pCommon, nCommon)

					// create a new component
					pDENew = dsdKernelDecompose_rec(pDsdMan, bFuncNew)
					bdd.Cudd_RecursiveDeref(dd, bFuncNew)
					// the BDD of the argument function in PRIME decomposition, should be regular
					pDENew = Dsd_Regular(pDENew)

					// add the first component
					pThis.pDecs[0] = pDENew
					return dsdKernelExit(pDsdMan, pThis, bF, bSuppNew, fCompF)
				}
			} // end of PRIME type
		} // end of existing common components
	} // end of CASE 3.a

	//////////////////////////////////////////////////////////////////////
	// CASE 4
	//////////////////////////////////////////////////////////////////////
	{
		// estimate the number of entries in the list
		nEntriesMax := pDsdMan.nInputs - dd.Perm[VarInt]

		// create the new decomposition entry
		nEntries := 0

		var SuppL, SuppH, SuppL_init, SuppH_init *DdNode
		var pHigher, pLower, pTemp, pDENew *Dsd_Node_t

		var levTopSuppL, levTopSuppH, levTop int

		// (C allocates nEntriesMax entries; a declared input set lying below
		// unrelated variables could exceed it, so the list grows instead)
		pThis = Dsd_TreeNodeCreate(int(DSD_NODE_PRIME), 0, pDsdMan.nextBlock())
		pThis.pDecs = make([]*Dsd_Node_t, 0, max(nEntriesMax, 1))
		pThis.pDecs = append(pThis.pDecs, pVarCurDE)
		nEntries++
		// other entries will be added to this list one-by-one during analysis

		// consider the simplest case: when the supports are equal
		// and at least one of the components
		// is the PRIME without decompositions, or
		// when both of them are without decomposition
		if (((pLR.Type == DSD_NODE_PRIME && nSuppL == len(pLR.pDecs)) || (pHR.Type == DSD_NODE_PRIME && nSuppH == len(pHR.pDecs))) && pLR.S == pHR.S) ||
			((pLR.Type == DSD_NODE_PRIME && nSuppL == len(pLR.pDecs)) && (pHR.Type == DSD_NODE_PRIME && nSuppH == len(pHR.pDecs))) {
			// walk through both supports and create the decomposition list composed of simple entries
			SuppL = pLR.S
			SuppH = pHR.S
			for {
				// determine levels
				levTopSuppL = cuddI(dd, SuppL.Index)
				levTopSuppH = cuddI(dd, SuppH.Index)

				// skip the topmost variable in both supports
				if levTopSuppL <= levTopSuppH {
					levTop = levTopSuppL
					SuppL = SuppL.T
				} else {
					levTop = levTopSuppH
				}

				if levTopSuppH <= levTopSuppL {
					SuppH = SuppH.T
				}

				// set the new decomposition entry
				pThis.pDecs = append(pThis.pDecs, pDsdMan.inputs[dd.InvPerm[levTop]])
				nEntries++
				if !(SuppL != b1 || SuppH != b1) {
					break
				}
			}
		} else {
			// compare two different decomposition lists
			SuppL_init = pLR.S
			SuppH_init = pHR.S
			// start references (because these supports will change)
			SuppL = pLR.S
			bdd.Cudd_Ref(SuppL)
			SuppH = pHR.S
			bdd.Cudd_Ref(SuppH)
			for SuppL != b1 || SuppH != b1 {
				// determine the top level in cofactors and
				// whether they have the same top level
				TopLevL := cuddI(dd, SuppL.Index)
				TopLevH := cuddI(dd, SuppH.Index)
				TopLevel := TopLevH
				fEqualLevel := false

				var bVarTop *DdNode
				var bSuppSubract *DdNode

				if TopLevL < TopLevH {
					pHigher = pLR
					pLower = pHR
					TopLevel = TopLevL
				} else if TopLevL > TopLevH {
					pHigher = pHR
					pLower = pLR
				} else {
					fEqualLevel = true
				}
				_ = pLower
				if TopLevel == int(bdd.CUDD_CONST_INDEX) {
					panic("dsd: dsdKernelDecompose_rec: TopLevel != CUDD_CONST_INDEX")
				}

				// find the currently top variable in the decomposition lists
				bVarTop = dd.Vars[dd.InvPerm[TopLevel]]

				if !fEqualLevel {
					// find the lower support
					var bSuppLower *DdNode
					if TopLevL < TopLevH {
						bSuppLower = SuppH_init
					} else {
						bSuppLower = SuppL_init
					}

					// find the first component in pHigher
					// whose support does not overlap with supp(Lower)
					// and remember the previous component
					var fPolarity bool
					var pPrev *Dsd_Node_t                                 // the pointer to the component proceeding pCur
					pCur := pHigher                                       // the first component not contained in supp(Lower)
					for extraBddSuppOverlapping(dd, pCur.S, bSuppLower) { // get the next component
						pPrev = pCur
						pCur = dsdKernelFindContainingComponent(pDsdMan, pCur, bVarTop, &fPolarity)
					}

					// look for the possibility to subtract more than one component
					if pPrev == nil || pPrev.Type == DSD_NODE_PRIME {
						// if there is no previous component, or if the previous component is PRIME
						// there is no way to subtract more than one component

						// add the new decomposition entry (it is already regular)
						pThis.pDecs = append(pThis.pDecs, pCur)
						nEntries++
						// assign the support to be subtracted from both components
						bSuppSubract = pCur.S
					} else { // all other types
						// go through the decomposition list of pPrev and find components
						// whose support does not overlap with supp(Lower)
						pNonOverlap := make([]*Dsd_Node_t, 0, len(pPrev.pDecs))
						for i := 0; i < len(pPrev.pDecs); i++ {
							pTemp = Dsd_Regular(pPrev.pDecs[i])
							if !extraBddSuppOverlapping(dd, pTemp.S, bSuppLower) {
								pNonOverlap = append(pNonOverlap, pPrev.pDecs[i])
							}
						}
						nNonOverlap := len(pNonOverlap)
						if nNonOverlap == 0 {
							panic("dsd: dsdKernelDecompose_rec: nNonOverlap > 0")
						}

						if nNonOverlap == 1 { // one one component was found, which is the original one
							if Dsd_Regular(pNonOverlap[0]) != pCur {
								panic("dsd: dsdKernelDecompose_rec: Dsd_Regular(pNonOverlap[0]) == pCur")
							}
							// add the new decomposition entry
							pThis.pDecs = append(pThis.pDecs, pCur)
							nEntries++
							// assign the support to be subtracted from both components
							bSuppSubract = pCur.S
						} else { // more than one components was found
							// find the OR (EXOR) of the non-overlapping components
							var bCommF *DdNode
							dsdKernelComputeSumOfComponents(pDsdMan, pNonOverlap, nNonOverlap, &bCommF, nil, pPrev.Type == DSD_NODE_EXOR)
							bdd.Cudd_Ref(bCommF)

							// create a new gated
							pDENew = dsdKernelDecompose_rec(pDsdMan, bCommF)
							bdd.Cudd_RecursiveDeref(dd, bCommF)
							// make it regular... it must be regular already
							if Dsd_IsComplement(pDENew) {
								panic("dsd: dsdKernelDecompose_rec: !Dsd_IsComplement(pDENew)")
							}

							// add the new decomposition entry
							pThis.pDecs = append(pThis.pDecs, pDENew)
							nEntries++
							// assign the support to be subtracted from both components
							bSuppSubract = pDENew.S
						}
					}

					// subtract its support from the support of upper component
					if TopLevL < TopLevH {
						bTemp = SuppL
						SuppL = bdd.Cudd_bddExistAbstract(dd, bTemp, bSuppSubract)
						bdd.Cudd_Ref(SuppL)
						bdd.Cudd_RecursiveDeref(dd, bTemp)
					} else {
						bTemp = SuppH
						SuppH = bdd.Cudd_bddExistAbstract(dd, bTemp, bSuppSubract)
						bdd.Cudd_Ref(SuppH)
						bdd.Cudd_RecursiveDeref(dd, bTemp)
					}
				} else { // if ( fEqualLevel ) -- they have the same top level var
					pMarkedLeft := make([]*Dsd_Node_t, 0, 8) // the pointers to the marked blocks
					pMarkedPols := make([]bool, 0, 8)        // polarities of the marked blocks

					fPolarity := false
					pTempL := pLR

					fPolarityCurH := false
					var pPrevH *Dsd_Node_t
					pCurH := pHR

					fPolarityCurL := false
					var pPrevL *Dsd_Node_t
					pCurL := pLR // = pMarkedLeft[0];
					index := 1

					// set the new mark
					pDsdMan.sMark++
					s_Mark := pDsdMan.sMark

					// go over the dec list of pL, mark all components that contain the given variable
					if !extraBddSuppContainVar(dd, pLR.S, bVarTop) || !extraBddSuppContainVar(dd, pHR.S, bVarTop) {
						panic("dsd: dsdKernelDecompose_rec: Extra_bddSuppContainVar")
					}
					for {
						pTempL.kernelMark = s_Mark
						pMarkedLeft = append(pMarkedLeft, pTempL)
						pMarkedPols = append(pMarkedPols, fPolarity)
						pTempL = dsdKernelFindContainingComponent(pDsdMan, pTempL, bVarTop, &fPolarity)
						if pTempL == nil {
							break
						}
					}

					// go over the dec list of pH, and find the component that is marked and the previos one
					// (such component always exists, because they have common variables)
					for pCurH.kernelMark != s_Mark {
						pPrevH = pCurH
						pCurH = dsdKernelFindContainingComponent(pDsdMan, pCurH, bVarTop, &fPolarityCurH)
						if pCurH == nil {
							panic("dsd: dsdKernelDecompose_rec: pCurH")
						}
					}

					// go through the first list once again and find
					// the component proceeding the one marked found in the second list
					for pCurL != pCurH {
						pPrevL = pCurL
						pCurL = pMarkedLeft[index]
						fPolarityCurL = pMarkedPols[index]
						index++
					}

					// look for the possibility to subtract more than one component
					if pPrevL == nil || pPrevH == nil || pPrevL.Type != pPrevH.Type || pPrevL.Type == DSD_NODE_PRIME || fPolarityCurL != fPolarityCurH {
						// there is no way to extract more than one
						pThis.pDecs = append(pThis.pDecs, pCurH)
						nEntries++
						// assign the support to be subtracted from both components
						bSuppSubract = pCurH.S
					} else {
						// find the equal components in two decomposition lists
						var pLastDiffL, pLastDiffH *Dsd_Node_t
						pCommon, nCommon := dsdKernelFindCommonComponents(pDsdMan, pPrevL, pPrevH, &pLastDiffL, &pLastDiffH)

						if nCommon == 0 || nCommon == 1 {
							// one one component was found, which is the original one
							// add the new decomposition entry
							pThis.pDecs = append(pThis.pDecs, pCurL)
							nEntries++
							// assign the support to be subtracted from both components
							bSuppSubract = pCurL.S
						} else { // more than one components was found
							// find the OR (EXOR) of the non-overlapping components
							var bCommF *DdNode
							dsdKernelComputeSumOfComponents(pDsdMan, pCommon, nCommon, &bCommF, nil, pPrevL.Type == DSD_NODE_EXOR)
							bdd.Cudd_Ref(bCommF)

							pDENew = dsdKernelDecompose_rec(pDsdMan, bCommF)
							if Dsd_IsComplement(pDENew) { // cannot be complemented because of construction
								panic("dsd: dsdKernelDecompose_rec: !Dsd_IsComplement(pDENew)")
							}
							bdd.Cudd_RecursiveDeref(dd, bCommF)

							// add the new decomposition entry
							pThis.pDecs = append(pThis.pDecs, pDENew)
							nEntries++

							// assign the support to be subtracted from both components
							bSuppSubract = pDENew.S
						}
					}

					bTemp = SuppL
					SuppL = bdd.Cudd_bddExistAbstract(dd, bTemp, bSuppSubract)
					bdd.Cudd_Ref(SuppL)
					bdd.Cudd_RecursiveDeref(dd, bTemp)

					bTemp = SuppH
					SuppH = bdd.Cudd_bddExistAbstract(dd, bTemp, bSuppSubract)
					bdd.Cudd_Ref(SuppH)
					bdd.Cudd_RecursiveDeref(dd, bTemp)
				} // end of if ( fEqualLevel )
			} // end of decomposition list comparison
			bdd.Cudd_RecursiveDeref(dd, SuppL)
			bdd.Cudd_RecursiveDeref(dd, SuppH)
		}

		// update the number of entries in the new decomposition list
		pThis.pDecs = pThis.pDecs[:nEntries]
	}
	return dsdKernelExit(pDsdMan, pThis, bF, bSuppNew, fCompF)
}

// nextBlock is s_nDecBlocks++.
func (pDsdMan *Dsd_Manager_t) nextBlock() int {
	n := pDsdMan.sNDecBlocks
	pDsdMan.sNDecBlocks++
	return n
}

// dsdKernelExit is the EXIT label of dsdKernelDecompose_rec.
func dsdKernelExit(pDsdMan *Dsd_Manager_t, pThis *Dsd_Node_t, bF, bSuppNew *DdNode, fCompF bool) *Dsd_Node_t {
	// if the component created is complemented, it represents a function without complement
	// therefore, as it is, without complement, it should recieve the complemented function
	pThisR := Dsd_Regular(pThis)
	if pThisR.G != nil || pThisR.S != nil {
		panic("dsd: dsdKernelDecompose_rec: pThisR->G == NULL && pThisR->S == NULL")
	}

	if pThisR == pThis { // set regular function
		pThisR.G = bF
	} else { // set complemented function
		pThisR.G = bdd.Cudd_Not(bF)
	}
	bdd.Cudd_Ref(bF) // reference the function in the component

	if bSuppNew == nil {
		panic("dsd: dsdKernelDecompose_rec: bSuppNew")
	}
	pThisR.S = bSuppNew // takes the reference from the new support
	if pDsdMan.tableInsert(bF, pThis) {
		panic("dsd: dsdKernelDecompose_rec: st__insert found the entry")
	}

	return Dsd_NotCond(pThis, fCompF)
}

// dsdKernelFindContainingComponent returns the non-complemented pointer to
// the DecEntry of that component which contains the given variable in its
// support, or nil if no such component exists.
func dsdKernelFindContainingComponent(pDsdMan *Dsd_Manager_t, pWhere *Dsd_Node_t, Var *DdNode, fPolarity *bool) *Dsd_Node_t {
	if len(pWhere.pDecs) == 1 {
		return nil
	}

	for i := 0; i < len(pWhere.pDecs); i++ {
		pTemp := Dsd_Regular(pWhere.pDecs[i])
		if extraBddSuppContainVar(pDsdMan.dd, pTemp.S, Var) {
			*fPolarity = pTemp != pWhere.pDecs[i]
			return pTemp
		}
	}
	panic("dsd: dsdKernelFindContainingComponent: no component contains the variable")
}

// dsdKernelFindCommonComponents determines the common components. It
// counts the number of common components in the decomposition lists of pL
// and pH and returns their list and number. It assumes that pL and pH are
// regular pointers. It returns also the pointers to the last different
// components encountered in pL and pH.
func dsdKernelFindCommonComponents(pDsdMan *Dsd_Manager_t, pL, pH *Dsd_Node_t, pLastDiffL, pLastDiffH **Dsd_Node_t) ([]*Dsd_Node_t, int) {
	Common := make([]*Dsd_Node_t, min(len(pL.pDecs), len(pH.pDecs)))
	nCommon := 0

	// the indices running through the components
	iCurL := 0
	iCurH := 0
	for iCurL < len(pL.pDecs) && iCurH < len(pH.pDecs) { // both did not run out
		pLcur := Dsd_Regular(pL.pDecs[iCurL])
		pHcur := Dsd_Regular(pH.pDecs[iCurH])

		bSLcur := pLcur.S
		bSHcur := pHcur.S

		// find out what component is higher in the BDD
		var TopVar uint32
		if pDsdMan.dd.Perm[bSLcur.Index] < pDsdMan.dd.Perm[bSHcur.Index] {
			TopVar = bSLcur.Index
		} else {
			TopVar = bSHcur.Index
		}

		if TopVar == bSLcur.Index && TopVar == bSHcur.Index {
			// the components may be equal - should match exactly!
			if pL.pDecs[iCurL] == pH.pDecs[iCurH] {
				Common[nCommon] = pL.pDecs[iCurL]
				nCommon++
			} else {
				*pLastDiffL = pL.pDecs[iCurL]
				*pLastDiffH = pH.pDecs[iCurH]
			}

			// skip both
			iCurL++
			iCurH++
		} else if TopVar == bSLcur.Index {
			// the components cannot be equal
			// skip the top-most one
			*pLastDiffL = pL.pDecs[iCurL]
			iCurL++
		} else { // if ( TopVar == bSHcur->index )
			// the components cannot be equal
			// skip the top-most one
			*pLastDiffH = pH.pDecs[iCurH]
			iCurH++
		}
	}

	// if one of the lists still has components, write the first one down
	if iCurL < len(pL.pDecs) {
		*pLastDiffL = pL.pDecs[iCurL]
	}

	if iCurH < len(pH.pDecs) {
		*pLastDiffH = pH.pDecs[iCurH]
	}

	// return the number of common components
	return Common, nCommon
}

// dsdKernelComputeSumOfComponents computes the sum (OR or EXOR) of the
// functions of the components.
func dsdKernelComputeSumOfComponents(pDsdMan *Dsd_Manager_t, pCommon []*Dsd_Node_t, nCommon int, pCompF, pCompS **DdNode, fExor bool) {
	dd := pDsdMan.dd
	var bF, bFadd, bTemp, bS *DdNode

	// start the function
	bF = dd.LogicZero
	bdd.Cudd_Ref(bF)
	// start the support
	if pCompS != nil {
		bS = dd.One
		bdd.Cudd_Ref(bS)
	}

	if nCommon <= 0 {
		panic("dsd: dsdKernelComputeSumOfComponents: nCommon > 0")
	}
	for i := 0; i < nCommon; i++ {
		pDE := pCommon[i]
		pDER := Dsd_Regular(pDE)
		if pDE != pDER {
			bFadd = bdd.Cudd_Not(pDER.G)
		} else {
			bFadd = pDER.G
		}
		// add to the function
		bTemp = bF
		if fExor {
			bF = bdd.Cudd_bddXor(dd, bTemp, bFadd)
		} else {
			bF = bdd.Cudd_bddOr(dd, bTemp, bFadd)
		}
		bdd.Cudd_Ref(bF)
		bdd.Cudd_RecursiveDeref(dd, bTemp)
		if pCompS != nil {
			// add to the support
			bTemp = bS
			bS = bdd.Cudd_bddAnd(dd, bTemp, pDER.S)
			bdd.Cudd_Ref(bS)
			bdd.Cudd_RecursiveDeref(dd, bTemp)
		}
	}
	// return the function
	bdd.Cudd_Deref(bF)
	*pCompF = bF

	// return the support
	if pCompS != nil {
		bdd.Cudd_Deref(bS)
		*pCompS = bS
	}
}

// dsdKernelCheckContainment returns true if support of one component is
// contained in that of another. In this case, pLarge (pSmall) is assigned
// to point to the larger (smaller) support. If the supports are identical
// returns false, and does not assign the components.
func dsdKernelCheckContainment(pDsdMan *Dsd_Manager_t, pL, pH *Dsd_Node_t, pLarge, pSmall **Dsd_Node_t) bool {
	dd := pDsdMan.dd
	var bSuppLarge, bSuppSmall *DdNode

	RetValue := extraBddSuppCheckContainment(dd, pL.S, pH.S, &bSuppLarge, &bSuppSmall)

	if !RetValue {
		return false
	}

	if pH.S == bSuppLarge {
		*pLarge = pH
		*pSmall = pL
	} else { // if ( pL->S == bSuppLarge )
		*pLarge = pL
		*pSmall = pH
	}
	return true
}

// dsdKernelCopyListPlusOne copies the list of components plus one.
func dsdKernelCopyListPlusOne(p, First *Dsd_Node_t, ppList []*Dsd_Node_t, nListSize int) {
	if nListSize+1 != len(p.pDecs) {
		panic("dsd: dsdKernelCopyListPlusOne: nListSize+1 == p->nDecs")
	}
	p.pDecs[0] = First
	for i := 0; i < nListSize; i++ {
		p.pDecs[i+1] = ppList[i]
	}
}

// dsdKernelCopyListPlusOneMinusOne copies the list of components plus
// one, and skips one.
func dsdKernelCopyListPlusOneMinusOne(p, First *Dsd_Node_t, ppList []*Dsd_Node_t, nListSize, iSkipped int) {
	if nListSize != len(p.pDecs) {
		panic("dsd: dsdKernelCopyListPlusOneMinusOne: nListSize == p->nDecs")
	}
	p.pDecs[0] = First
	Counter := 1
	for i := 0; i < nListSize; i++ {
		if i != iSkipped {
			p.pDecs[Counter] = ppList[i]
			Counter++
		}
	}
}
