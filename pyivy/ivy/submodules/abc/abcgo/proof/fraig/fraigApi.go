package fraig

import "fmt"

/**CFile****************************************************************

  FileName    [fraigApi.c]

  PackageName [FRAIG: Functionally reduced AND-INV graphs.]

  Synopsis    [Access APIs for the FRAIG manager and node.]

  Author      [Alan Mishchenko <alanmi@eecs.berkeley.edu>]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 2.0. Started - October 1, 2004]

  Revision    [$Id: fraigApi.c,v 1.2 2005/07/08 01:01:30 alanmi Exp $]

***********************************************************************/

//#include "fraigInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Access functions to read the data members of the manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_ManReadVecInputs(p *Fraig_Man_t) *Fraig_NodeVec_t { return p.vInputs;            }
func Fraig_ManReadVecOutputs(p *Fraig_Man_t) *Fraig_NodeVec_t { return p.vOutputs;           }
func Fraig_ManReadVecNodes(p *Fraig_Man_t) *Fraig_NodeVec_t { return p.vNodes;             }
func Fraig_ManReadInputs(p *Fraig_Man_t) **Fraig_Node_t { return p.vInputs.pArray;    }
func Fraig_ManReadOutputs(p *Fraig_Man_t) **Fraig_Node_t { return p.vOutputs.pArray;   }
func Fraig_ManReadNodes(p *Fraig_Man_t) **Fraig_Node_t { return p.vNodes.pArray;     }
func Fraig_ManReadInputNum(p *Fraig_Man_t) int32 { return p.vInputs.nSize;     }
func Fraig_ManReadOutputNum(p *Fraig_Man_t) int32 { return p.vOutputs.nSize;    }
func Fraig_ManReadNodeNum(p *Fraig_Man_t) int32 { return p.vNodes.nSize;      }
func Fraig_ManReadConst1(p *Fraig_Man_t) *Fraig_Node_t { return p.pConst1;            }
func Fraig_ManReadIthNode(p *Fraig_Man_t, i int32) *Fraig_Node_t { assert(i < p.vNodes.nSize, "i < p.vNodes.nSize"); return p.vNodes.pArray[i];  }
func Fraig_ManReadInputNames(p *Fraig_Man_t) *string { return p.ppInputNames;       }
func Fraig_ManReadOutputNames(p *Fraig_Man_t) *string { return p.ppOutputNames;      }
func Fraig_ManReadVarsInt(p *Fraig_Man_t) string { return (char *)p.vVarsInt;   }
func Fraig_ManReadSat(p *Fraig_Man_t) string { return (char *)p.pSat;       }
func Fraig_ManReadFuncRed(p *Fraig_Man_t) int32 { return p.fFuncRed;   }
func Fraig_ManReadFeedBack(p *Fraig_Man_t) int32 { return p.fFeedBack;  }
func Fraig_ManReadDoSparse(p *Fraig_Man_t) int32 { return p.fDoSparse;  }
func Fraig_ManReadChoicing(p *Fraig_Man_t) int32 { return p.fChoicing;  }
func Fraig_ManReadVerbose(p *Fraig_Man_t) int32 { return p.fVerbose;   }
func Fraig_ManReadModel(p *Fraig_Man_t) *int32 { return p.pModel;     }
// returns the number of patterns used for random simulation (this number is fixed for the FRAIG run)
func Fraig_ManReadPatternNumRandom(p *Fraig_Man_t) int32 { return p.nWordsRand * 32;  }
// returns the number of dynamic patterns accumulated at runtime (include SAT solver counter-examples and distance-1 patterns derived from them)
func Fraig_ManReadPatternNumDynamic(p *Fraig_Man_t) int32 { return p.iWordStart * 32;  }
// returns the number of dynamic patterns proved useful to distinquish some FRAIG nodes (this number is more than 0 after the first garbage collection of patterns)
func Fraig_ManReadPatternNumDynamicFiltered(p *Fraig_Man_t) int32 { return p.iPatsPerm;        }
// returns the number of times FRAIG package timed out
func Fraig_ManReadSatFails(p *Fraig_Man_t) int32 { return p.nSatFailsReal;    }      
// returns the number of conflicts in the SAT solver
func Fraig_ManReadConflicts(p *Fraig_Man_t) int32 { if p.pSat {
    return Msat_SolverReadBackTracks(p.pSat);
} else {
    return 0;
} }      
// returns the number of inspections in the SAT solver
func Fraig_ManReadInspects(p *Fraig_Man_t) int32 { if p.pSat {
    return Msat_SolverReadInspects(p.pSat);
} else {
    return 0;
}   }            

/**Function*************************************************************

  Synopsis    [Access functions to set the data members of the manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_ManSetFuncRed(p *Fraig_Man_t, fFuncRed int32) { p.fFuncRed  = fFuncRed;  }
func Fraig_ManSetFeedBack(p *Fraig_Man_t, fFeedBack int32) { p.fFeedBack = fFeedBack; }
func Fraig_ManSetDoSparse(p *Fraig_Man_t, fDoSparse int32) { p.fDoSparse = fDoSparse; }
func Fraig_ManSetChoicing(p *Fraig_Man_t, fChoicing int32) { p.fChoicing = fChoicing; }
func Fraig_ManSetTryProve(p *Fraig_Man_t, fTryProve int32) { p.fTryProve = fTryProve; }
func Fraig_ManSetVerbose(p *Fraig_Man_t, fVerbose int32) { p.fVerbose  = fVerbose;  }
func Fraig_ManSetOutputNames(p *Fraig_Man_t, ppNames *string) { p.ppOutputNames = ppNames; }
func Fraig_ManSetInputNames(p *Fraig_Man_t, ppNames *string) { p.ppInputNames  = ppNames; }

/**Function*************************************************************

  Synopsis    [Access functions to read the data members of the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_NodeReadData0(p *Fraig_Node_t) *Fraig_Node_t { return p.pData0;    }
func Fraig_NodeReadData1(p *Fraig_Node_t) *Fraig_Node_t { return p.pData1;    }
func Fraig_NodeReadNum(p *Fraig_Node_t) int32 { return p.Num;       }
// inlining of macro Fraig_IsComplement
func Fraig_NodeReadOne(p *Fraig_Node_t) *Fraig_Node_t { assert(!(((int32(((ABC_PTRUINT_T) (p) & 01)) != 0))), "!/*__starting_point_inline_macro:Fraig_IsComplement__*/((int32(((ABC_PTRUINT_T) (p) & 01))))"); return p.p1; }
// inlining of macro Fraig_IsComplement
func Fraig_NodeReadTwo(p *Fraig_Node_t) *Fraig_Node_t { assert(!(((int32(((ABC_PTRUINT_T) (p) & 01)) != 0))), "!/*__starting_point_inline_macro:Fraig_IsComplement__*/((int32(((ABC_PTRUINT_T) (p) & 01))))"); return p.p2; }
func Fraig_NodeReadNextE(p *Fraig_Node_t) *Fraig_Node_t { return p.pNextE;    }
func Fraig_NodeReadRepr(p *Fraig_Node_t) *Fraig_Node_t { return p.pRepr;     }
func Fraig_NodeReadNumRefs(p *Fraig_Node_t) int32 { return p.nRefs;     }
func Fraig_NodeReadNumFanouts(p *Fraig_Node_t) int32 { return p.nFanouts;  }
func Fraig_NodeReadSimInv(p *Fraig_Node_t) int32 { return p.fInv;      }
func Fraig_NodeReadNumOnes(p *Fraig_Node_t) int32 { return p.nOnes;     }
// returns the pointer to the random simulation patterns (their number is returned by Fraig_ManReadPatternNumRandom)
// memory pointed to by this and the following procedure is maintained by the FRAIG package and exists as long as the package runs
func Fraig_NodeReadPatternsRandom(p *Fraig_Node_t) *uint32 { return p.puSimR;    }
// returns the pointer to the dynamic simulation patterns (their number is returned by Fraig_ManReadPatternNumDynamic or Fraig_ManReadPatternNumDynamicFiltered)
// if the number of patterns is not evenly divisible by 32, the patterns beyond the given number contain garbage
func Fraig_NodeReadPatternsDynamic(p *Fraig_Node_t) *uint32 { return p.puSimD;    }

/**Function*************************************************************

  Synopsis    [Access functions to set the data members of the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_NodeSetData0(p *Fraig_Node_t, pData *Fraig_Node_t) { p.pData0  = pData;  }
func Fraig_NodeSetData1(p *Fraig_Node_t, pData *Fraig_Node_t) { p.pData1  = pData;  }

/**Function*************************************************************

  Synopsis    [Checks the type of the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
// inlining of macro Fraig_Regular
func Fraig_NodeIsConst(p *Fraig_Node_t) int32 {  return (((Fraig_Node_t *)((ABC_PTRUINT_T)(p) & ~01)) ).Num   == 0;  }
// inlining of macro Fraig_Regular
func Fraig_NodeIsVar(p *Fraig_Node_t) int32 {  return (((Fraig_Node_t *)((ABC_PTRUINT_T)(p) & ~01)) ).NumPi >= 0;  }
// inlining of macro Fraig_Regular
func Fraig_NodeIsAnd(p *Fraig_Node_t) int32 {  return (((Fraig_Node_t *)((ABC_PTRUINT_T)(p) & ~01)) ).NumPi <  0 && (((Fraig_Node_t *)((ABC_PTRUINT_T)(p) & ~01)) ).Num > 0;  }
// inlining of macro Fraig_IsComplement
func Fraig_NodeComparePhase(p1 *Fraig_Node_t, p2 *Fraig_Node_t) int32 { assert(!(((int32(((ABC_PTRUINT_T) (p1) & 01)) != 0))), "!/*__starting_point_inline_macro:Fraig_IsComplement__*/((int32(((ABC_PTRUINT_T) (p1) & 01))))"); assert(!(((int32(((ABC_PTRUINT_T) (p2) & 01)) != 0))), "!/*__starting_point_inline_macro:Fraig_IsComplement__*/((int32(((ABC_PTRUINT_T) (p2) & 01))))"); return p1.fInv ^ p2.fInv; }

/**Function*************************************************************

  Synopsis    [Returns a new primary input node.]

  Description [If the node with this number does not exist, 
  create a new PI node with this number.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_ManReadIthVar(p *Fraig_Man_t, i int32) *Fraig_Node_t {
    var k int32;
    if i < 0 {
        fmt.Printf("Requesting a PI with a negative number\n");
        return nil;
    }
    // create the PIs to fill in the interval
    if i >= p.vInputs.nSize {
        for  k = p.vInputs.nSize; k <= i; k++  {
            Fraig_NodeCreatePi( p );
        } } 
    return p.vInputs.pArray[i];
}

/**Function*************************************************************

  Synopsis    [Creates a new PO node.]

  Description [This procedure may take a complemented node.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_ManSetPo(p *Fraig_Man_t, pNode *Fraig_Node_t) {
    // internal node may be a PO two times
    // inlining of macro Fraig_Regular
    ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode) & ~01)) .fNodePo = 1;
    Fraig_NodeVecPush( p.vOutputs, pNode );
}

/**Function*************************************************************

  Synopsis    [Perfoms the AND operation with functional hashing.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_NodeAnd(p *Fraig_Man_t, p1 *Fraig_Node_t, p2 *Fraig_Node_t) *Fraig_Node_t {
    return Fraig_NodeAndCanon( p, p1, p2 );
}

/**Function*************************************************************

  Synopsis    [Perfoms the OR operation with functional hashing.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_NodeOr(p *Fraig_Man_t, p1 *Fraig_Node_t, p2 *Fraig_Node_t) *Fraig_Node_t {
    // inlining of macro Fraig_Not
    return ((Fraig_Node_t *)((ABC_PTRUINT_T)(Fraig_NodeAndCanon( p, ((Fraig_Node_t *)((ABC_PTRUINT_T)(p1) ^ 01)) , ((Fraig_Node_t *)((ABC_PTRUINT_T)(p2) ^ 01))  )) ^ 01)) ;
}

/**Function*************************************************************

  Synopsis    [Perfoms the EXOR operation with functional hashing.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_NodeExor(p *Fraig_Man_t, p1 *Fraig_Node_t, p2 *Fraig_Node_t) *Fraig_Node_t {
    // inlining of macro Fraig_Not
    return Fraig_NodeMux( p, p1, ((Fraig_Node_t *)((ABC_PTRUINT_T)(p2) ^ 01)) , p2 );
}

/**Function*************************************************************

  Synopsis    [Perfoms the MUX operation with functional hashing.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_NodeMux(p *Fraig_Man_t, pC *Fraig_Node_t, pT *Fraig_Node_t, pE *Fraig_Node_t) *Fraig_Node_t {
    var pAnd1, pAnd2, pRes *Fraig_Node_t;
    // inlining of macro Fraig_Ref
    pAnd1 = Fraig_NodeAndCanon( p, pC,          pT );     ;
    // inlining of macro Fraig_Not
    // inlining of macro Fraig_Ref
    pAnd2 = Fraig_NodeAndCanon( p, ((Fraig_Node_t *)((ABC_PTRUINT_T)(pC) ^ 01)) , pE );   ;
    pRes  = Fraig_NodeOr( p, pAnd1, pAnd2 ); 
    // inlining of macro Fraig_RecursiveDeref
    ;
    // inlining of macro Fraig_RecursiveDeref
    ;
    // inlining of macro Fraig_Deref
    ;
    return pRes;
}


/**Function*************************************************************

  Synopsis    [Sets the node to be equivalent to the given one.]

  Description [This procedure is a work-around for the equivalence check.
  Does not verify the equivalence. Use at the user's risk.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_NodeSetChoice(pMan *Fraig_Man_t, pNodeOld *Fraig_Node_t, pNodeNew *Fraig_Node_t) {
//    assert( pMan->fChoicing );
    pNodeNew.pNextE = pNodeOld.pNextE;
    pNodeOld.pNextE = pNodeNew;
    pNodeNew.pRepr  = pNodeOld;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

