
package cnf

import "fmt"

/**CFile****************************************************************

  FileName    [cnfPost.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [AIG-to-CNF conversion.]

  Synopsis    []

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - April 28, 2007.]

  Revision    [$Id: cnfPost.c,v 1.00 2007/04/28 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "cnf.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cnf_ManPostprocess_old(p *Cnf_Man_t) {
//    extern int Aig_ManLargeCutEval( Aig_Man_t * p, Aig_Obj_t * pRoot, Dar_Cut_t * pCutR, Dar_Cut_t * pCutL, int Leaf );
    var nNew, Gain, nGain, nVars int32;

    var pObj, pFan *Aig_Obj_t;
    var pCutBest, pCut *Dar_Cut_t;
    var i, k int32;//, a, b, Counter;
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pManAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pManAig.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        if !Aig_ObjIsNode(pObj) {
            continue; }
        if pObj.nRefs == 0 {
            continue; }
//        pCutBest = Aig_ObjBestCut(pObj);
        pCutBest = nil;

        // inlining of macro Dar_CutForEachLeaf
        for  k = 0; (k < int32((pCutBest).nLeaves)) && (((pFan) = Aig_ManObj(p.pManAig, (pCutBest).pLeaves[k])), 1); k++  {
            if !Aig_ObjIsNode(pFan) {
                continue; }
            assert(pFan.nRefs != 0, "pFan.nRefs != 0");
            if pFan.nRefs != 1 {
                continue; }
//            pCut = Aig_ObjBestCut(pFan);
            pCut = nil;
/*
            // find how many common variable they have
            Counter = 0;
            for ( a = 0; a < (int)pCut->nLeaves; a++ )
            {
                for ( b = 0; b < (int)pCutBest->nLeaves; b++ )
                    if ( pCut->pLeaves[a] == pCutBest->pLeaves[b] )
                        break;
                if ( b == (int)pCutBest->nLeaves )
                    continue;
                Counter++;
            }
            printf( "%d ", Counter );
*/
            // find the new truth table after collapsing these two cuts


//            nNew = Aig_ManLargeCutEval( p->pManAig, pObj, pCutBest, pCut, pFan->Id );
            nNew = 0;


//            printf( "%d+%d=%d:%d(%d) ", pCutBest->Cost, pCut->Cost, 
//                pCutBest->Cost+pCut->Cost, nNew, pCutBest->Cost+pCut->Cost-nNew );

            Gain = pCutBest.Value + pCut.Value - nNew;
            if Gain > 0 {
                nGain += Gain;
                nVars++;
            }
        }
    }
    }
    fmt.Printf("Total gain = %d.  Vars = %d.\n", nGain, nVars);
}

/**Function*************************************************************

  Synopsis    [Transfers cuts of the mapped nodes into internal representation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cnf_ManTransferCuts(p *Cnf_Man_t) {
    var pObj *Aig_Obj_t;
    var i int32;
    Aig_MmFlexRestart( p.pMemCuts );
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pManAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pManAig.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        if Aig_ObjIsNode(pObj) && pObj.nRefs > 0 {
            pObj.pData = Cnf_CutCreate( p, pObj );
        } else {
            pObj.pData = nil; }
    }
    }
}

/**Function*************************************************************

  Synopsis    [Transfers cuts of the mapped nodes into internal representation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cnf_ManFreeCuts(p *Cnf_Man_t) {
    var pObj *Aig_Obj_t;
    var i int32;
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pManAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pManAig.vObjs, i)), 1); i++  { if (pObj) == nil {} else
        if pObj.pData {
            Cnf_CutFree( (Cnf_Cut_t *)pObj.pData );
            pObj.pData = nil;
        }
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cnf_ManPostprocess(p *Cnf_Man_t) {
    var pCut, pCutFan, pCutRes *Cnf_Cut_t;
    var pObj, pFan *Aig_Obj_t;
    var Order, Costs [16]int32;
    var i, k, fChanges int32;
    // inlining of macro Aig_ManForEachNode
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pManAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pManAig.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
        if pObj.nRefs == 0 {
            continue; }
        pCut = Cnf_ObjBestCut(pObj);

        // sort fanins according to their size
        // inlining of macro Cnf_CutForEachLeaf
        for  k = 0; (k < int32((pCut).nFanins)) && ((pFan) = Aig_ManObj(p.pManAig, (pCut).pFanins[k])); k++  {
            Order[k] = k;
            Costs[k] = Aig_ObjIsNode(pFan)? Cnf_ObjBestCut(pFan).Cost : 0;
        }
        // sort the cuts by Weight
        do {
            var Temp int32;
            fChanges = 0;
            for  k = 0; k < pCut.nFanins - 1; k++  {
                if Costs[Order[k]] <= Costs[Order[k+1]] {
                    continue; }
                Temp = Order[k];
                Order[k] = Order[k+1];
                Order[k+1] = Temp;
                fChanges = 1;
            }
        } while ( fChanges );


//        Cnf_CutForEachLeaf( p->pManAig, pCut, pFan, k )
        for  k = 0; (k < int32((pCut).nFanins)) && ((pFan) = Aig_ManObj(p.pManAig, (pCut).pFanins[Order[k]])); k++  {
            if !Aig_ObjIsNode(pFan) {
                continue; }
            assert(pFan.nRefs != 0, "pFan.nRefs != 0");
            if pFan.nRefs != 1 {
                continue; }
            pCutFan = Cnf_ObjBestCut(pFan);
            // try composing these two cuts
//            Cnf_CutPrint( pCut );
            pCutRes = Cnf_CutCompose( p, pCut, pCutFan, pFan.Id );
//            Cnf_CutPrint( pCut );
//            printf( "\n" );
            // check if the cost if reduced
            if pCutRes == nil || pCutRes.Cost == 127 || pCutRes.Cost > pCut.Cost + pCutFan.Cost {
                if pCutRes != nil {
                    Cnf_CutFree( pCutRes ); }
                continue;
            }
            // update the cut
            Cnf_ObjSetBestCut( pObj, pCutRes );
            Cnf_ObjSetBestCut( pFan, nil );
            Cnf_CutUpdateRefs( p, pCut, pCutFan, pCutRes );
            assert(pFan.nRefs == 0, "pFan.nRefs == 0");
            Cnf_CutFree( pCut );
            Cnf_CutFree( pCutFan );
            break;
        }
    }
    }
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

