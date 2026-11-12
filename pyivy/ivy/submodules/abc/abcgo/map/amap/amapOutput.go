package amap

/**CFile****************************************************************

  FileName    [amapOutput.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Technology mapper for standard cells.]

  Synopsis    [Core mapping procedures.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: amapOutput.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "amapInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

func Amap_OuputStrsav(p *Aig_MmFlex_t, pStr string) string { if pStr {
    return strcpy(Aig_MmFlexEntryFetch(p, len(pStr)+1), pStr);
} else {
    return nil;
} }

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Allocates structure for storing one gate.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Amap_OutputStructAlloc(pMem *Aig_MmFlex_t, pGate *Amap_Gat_t) *Amap_Out_t {
    var pRes *Amap_Out_t;
    var nFans int32;
    if pGate != nil {
        nFans = pGate.nPins;
    } else {
        nFans = 1;
    }
    pRes = (Amap_Out_t *)Aig_MmFlexEntryFetch( pMem, sizeof(Amap_Out_t)+sizeof(int)*nFans );
    memset( pRes, 0, sizeof(Amap_Out_t) );
    memset( pRes.pFans, 0xff, sizeof(int)*nFans );
    if pGate != nil {
        pRes.pName = Amap_OuputStrsav( pMem, pGate.pName );
    } else {
        pRes.pName = nil;
    }
    pRes.nFans = nFans;
    return pRes;
}

/**Function*************************************************************

  Synopsis    [Returns mapped network as an array of structures.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Amap_ManProduceMapped(p *Amap_Man_t) *Vec_Ptr_t {
    var vNodes *Vec_Ptr_t;
    var pMem *Aig_MmFlex_t;
    var pObj, pFanin *Amap_Obj_t;
    var pGate *Amap_Gat_t;
    var pRes *Amap_Out_t;
    var i, k, iFanin, fCompl int32;
    var TotalArea float32  = 0.0;
    pMem = Aig_MmFlexStart();
    // create mapping object for each node used in the mapping
    vNodes = Vec_PtrAlloc( 10 );
    // inlining of macro Amap_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Amap_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        if Amap_ObjIsPi(pObj) {
            assert(pObj.fPolar == 0, "pObj.fPolar == 0");
            pRes = Amap_OutputStructAlloc( pMem, nil );
            pRes.Type  = -1;
            pRes.nFans = 0;
            // save this structure
            pObj.iData = Vec_PtrSize( vNodes );
            Vec_PtrPush( vNodes, pRes );
            // create invertor if needed
            if pObj.nFouts[1] { // this PI is used in the neg polarity
            
                pRes = Amap_OutputStructAlloc( pMem, p.pLib.pGateInv );
                pRes.pFans[0] = pObj.iData;
                // save this structure
                Vec_PtrPush( vNodes, pRes );
                TotalArea += p.pLib.pGateInv.dArea;
            }
            continue;
        }
        if Amap_ObjIsNode(pObj) {
            // skip the node that is not used in the mapping
            if Amap_ObjRefsTotal(pObj) == 0 {
                continue; }
            // get the gate
            pGate = Amap_LibGate( p.pLib, pObj.Best.pSet.iGate );
            assert(pGate.nPins == pObj.Best.pCut.nFans, "pGate.nPins == pObj.Best.pCut.nFans");
            // allocate structure
            pRes = Amap_OutputStructAlloc( pMem, pGate );
            // inlining of macro Amap_MatchForEachFaninCompl
            for  k = 0; k < int32((&pObj.Best).pCut.nFans) &&                                     ((pFanin = Amap_ManObj((p), Abc_Lit2Var((&pObj.Best).pCut.Fans[Abc_Lit2Var((&pObj.Best).pSet.Ins[k])]))), 1) &&         ((fCompl = Abc_LitIsCompl((&pObj.Best).pSet.Ins[k]) ^ Abc_LitIsCompl((&pObj.Best).pCut.Fans[Abc_Lit2Var((&pObj.Best).pSet.Ins[k])])), 1);         k++  {
                assert(Amap_ObjRefsTotal(pFanin), "Amap_ObjRefsTotal(pFanin)");
                if int32(pFanin.fPolar) == fCompl {
                    pRes.pFans[k] = pFanin.iData;
                } else {
                    pRes.pFans[k] = pFanin.iData + 1; }
            }
            // save this structure
            pObj.iData = Vec_PtrSize( vNodes );
            Vec_PtrPush( vNodes, pRes );
            TotalArea += pGate.dArea;
            // create invertor if needed
            if pObj.nFouts[!pObj.fPolar] { // needed in the opposite polarity
            
                pRes = Amap_OutputStructAlloc( pMem, p.pLib.pGateInv );
                pRes.pFans[0] = pObj.iData;
                // save this structure
                Vec_PtrPush( vNodes, pRes );
                TotalArea += p.pLib.pGateInv.dArea;
            }
            continue;
        }
        if Amap_ObjIsPo(pObj) {
            assert(pObj.fPolar == 0, "pObj.fPolar == 0");
            pFanin = Amap_ObjFanin0(p, pObj);
            assert(Amap_ObjRefsTotal(pFanin), "Amap_ObjRefsTotal(pFanin)");
            if Amap_ObjIsConst1(pFanin) { // create constant node
                if Amap_ObjFaninC0(pObj) {
                    pRes = Amap_OutputStructAlloc( pMem, p.pLib.pGate0 );
                    TotalArea += p.pLib.pGate0.dArea;
                
                } else {
                    pRes = Amap_OutputStructAlloc( pMem, p.pLib.pGate1 );
                    TotalArea += p.pLib.pGate1.dArea;
                }
                // save this structure
                iFanin = Vec_PtrSize( vNodes );
                Vec_PtrPush( vNodes, pRes );
            
            } else {
                if int32(pFanin.fPolar) == Amap_ObjFaninC0(pObj) {
                    iFanin = pFanin.iData;
                } else {
                    iFanin = pFanin.iData + 1; }
            }
            // create PO node
            pRes = Amap_OutputStructAlloc( pMem, nil );
            pRes.Type     = 1;
            pRes.pFans[0] = iFanin;
            // save this structure
            Vec_PtrPush( vNodes, pRes );
        }
    }
    }
    // return memory manager in the last entry of the array
    Vec_PtrPush( vNodes, pMem );
    return vNodes;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

