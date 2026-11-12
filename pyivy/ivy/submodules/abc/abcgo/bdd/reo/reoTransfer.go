package reo

/**CFile****************************************************************

  FileName    [reoTransfer.c]

  PackageName [REO: A specialized DD reordering engine.]

  Synopsis    [Transfering a DD from the CUDD manager into REO"s internal data structures and back.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - October 15, 2002.]

  Revision    [$Id: reoTransfer.c,v 1.0 2002/15/10 03:00:00 alanmi Exp $]

***********************************************************************/

//#include "reo.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                    FUNCTION DEFINITIONS                          ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Transfers the DD into the internal reordering data structure.]

  Description [It is important that the hash table is lossless.]

  SideEffects []

  SeeAlso     []

***********************************************************************/
func reoTransferNodesToUnits_rec(p *reo_man, F *DdNode) *reo_unit {
    var dd *DdManager  = p.dd;
    var pUnit *reo_unit;
    var HKey int32  = -1; // Suppress "might be used uninitialized"
        var fComp int32;
    
    // inlining of macro Cudd_IsComplement
    fComp = (int32( ((ptrint) (F) & 01)));
    // inlining of macro Cudd_Regular
    F = ((DdNode *)((ptruint)(F) & ~01));

    // check the hash-table
    if F.ref != 1 {
        // search cache - use linear probing
        // inlining of macro hashKey2
        for  HKey = (((ABC_PTRUINT_T)(p.Signature) + (ABC_PTRUINT_T)(F) * DD_P1) % p.nTableSize); p.HTable[HKey].Sign == p.Signature; HKey = (HKey+1) % p.nTableSize  {
            if p.HTable[HKey].Arg1 == (reo_unit *)F {
                pUnit = p.HTable[HKey].Arg2;  
                assert((pUnit != nil), "pUnit");
                // increment the edge counter
                pUnit.n++;
                // inlining of macro Unit_NotCond
                return ((reo_unit *)((ABC_PTRUINT_T)(pUnit) ^ (fComp)));
            }
        }
    }
    // the entry in not found in the cache
    
    // create a new entry
    pUnit         = reoUnitsGetNextUnit( p );
    pUnit.n      = 1;
    // inlining of macro cuddIsConstant
    if  (F).index == CUDD_CONST_INDEX {
        pUnit.lev    = REO_CONST_LEVEL;
        // inlining of macro cuddV
        pUnit.pE     = (reo_unit*)(ABC_PTRUINT_T)(((F).type.value));
        pUnit.pT     = nil;
        // check if the diagram that is being reordering has complement edges
        if F != dd.one {
            p.fThisIsAdd = 1; }
        // insert the unit into the corresponding plane
        reoUnitsAddUnitToPlane( &(p.pPlanes[p.nSupp]), pUnit ); // increments the unit counter
    
    } else {
        pUnit.lev    = p.pMapToPlanes[F.index];
        // inlining of macro cuddE
        pUnit.pE     = reoTransferNodesToUnits_rec( p, ((F).type.kids.E) );
        // inlining of macro cuddT
        pUnit.pT     = reoTransferNodesToUnits_rec( p, ((F).type.kids.T) );
        // insert the unit into the corresponding plane
        reoUnitsAddUnitToPlane( &(p.pPlanes[pUnit.lev]), pUnit ); // increments the unit counter
    }

    // add to the hash table
    if F.ref != 1 {
        // the next free entry is already found - it is pointed to by HKey
        // while we traversed the diagram, the hash entry to which HKey points,
        // might have been used. Make sure that its signature is different.
        for  ; p.HTable[HKey].Sign == p.Signature; HKey = (HKey+1) % p.nTableSize  {}
        p.HTable[HKey].Sign = p.Signature;
        p.HTable[HKey].Arg1 = (reo_unit *)F;
        p.HTable[HKey].Arg2 = pUnit;
    }

    // increment the counter of nodes
    p.nNodesCur++;
    // inlining of macro Unit_NotCond
    return ((reo_unit *)((ABC_PTRUINT_T)(pUnit) ^ (fComp)));
}

/**Function*************************************************************

  Synopsis    [Creates the DD from the internal reordering data structure.]

  Description [It is important that the hash table is lossless.]

  SideEffects []

  SeeAlso     []

***********************************************************************/
func reoTransferUnitsToNodes_rec(p *reo_man, pUnit *reo_unit) *DdNode {
    var dd *DdManager  = p.dd;
    var bRes, E, T *DdNode;
    var HKey int32  = -1; // Suppress "might be used uninitialized"
        var fComp int32;

    // inlining of macro Cudd_IsComplement
    fComp = (int32( ((ptrint) (pUnit) & 01)));
    // inlining of macro Unit_Regular
    pUnit = ((reo_unit *)((ABC_PTRUINT_T)(pUnit) & ~01));

    // check the hash-table
    if pUnit.n != 1 {
        // inlining of macro hashKey2
        for  HKey = (((ABC_PTRUINT_T)(p.Signature) + (ABC_PTRUINT_T)(pUnit) * DD_P1) % p.nTableSize); p.HTable[HKey].Sign == p.Signature; HKey = (HKey+1) % p.nTableSize  {
            if p.HTable[HKey].Arg1 == pUnit {
                bRes = (DdNode*) p.HTable[HKey].Arg2;  
                assert(bRes, "bRes");
                // inlining of macro Cudd_NotCond
                return ((DdNode *)((ptrint)(bRes) ^ (fComp)));
            }
        }
    }

    // treat the case of constants
    // inlining of macro Unit_IsConstant
    if  int32(((pUnit).lev == REO_CONST_LEVEL)) != 0 {
        bRes = cuddUniqueConst( dd, (float64((int32((ABC_PTRUINT_T)(pUnit.pE))))) );
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(bRes).ref);
    
    } else {
        // split and recur on children of this node
        E = reoTransferUnitsToNodes_rec( p, pUnit.pE );
        if E == nil {
            return nil; }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(E).ref);

        T = reoTransferUnitsToNodes_rec( p, pUnit.pT );
        if T == nil {
            Cudd_RecursiveDeref(dd, E);
            return nil;
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(T).ref);
        
        // consider the case when Res0 and Res1 are the same node
        assert(E != T, "E != T");
        // inlining of macro Cudd_IsComplement
        assert(!((int32( ((ptrint) (T) & 01)) != 0)), "!/*__starting_point_inline_macro:Cudd_IsComplement__*/(int32( ((ptrint) (T) & 01)))");

        bRes = cuddUniqueInter( dd, p.pMapToDdVarsFinal[pUnit.lev], T, E );
        if bRes == nil {
            Cudd_RecursiveDeref(dd,E);
            Cudd_RecursiveDeref(dd,T);
            return nil;
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(bRes).ref);
        // inlining of macro cuddDeref
        cuddSatDec(Cudd_Regular(E).ref);
        // inlining of macro cuddDeref
        cuddSatDec(Cudd_Regular(T).ref);
    }

    // do not keep the result if the ref count is only 1, since it will not be visited again
    if pUnit.n != 1 {
         // while we traversed the diagram, the hash entry to which HKey points,
         // might have been used. Make sure that its signature is different.
         for  ; p.HTable[HKey].Sign == p.Signature; HKey = (HKey+1) % p.nTableSize  {}
         p.HTable[HKey].Sign = p.Signature;
         p.HTable[HKey].Arg1 = pUnit;
         p.HTable[HKey].Arg2 = (reo_unit *)bRes;  

         // add the DD to the referenced DD list in order to be able to store it in cache
         p.pRefNodes[p.nRefNodes++] = bRes;  Cudd_Ref( bRes ); 
         // no need to do this, because the garbage collection will not take bRes away
         // it is held by the diagram in the making
    }
    // increment the counter of nodes
    p.nNodesCur++;
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(bRes).ref);
    // inlining of macro Cudd_NotCond
    return ((DdNode *)((ptrint)(bRes) ^ (fComp)));
}

////////////////////////////////////////////////////////////////////////
///                         END OF FILE                              ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END

