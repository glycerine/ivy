package aig

/**CFile****************************************************************

  FileName    [aigFrames.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [AIG package.]

  Synopsis    [Performs timeframe expansion of the AIG.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - April 28, 2007.]

  Revision    [$Id: aigFrames.c,v 1.00 2007/04/28 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "aig.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

func Aig_ObjFrames(pObjMap **Aig_Obj_t, nFs int32, pObj *Aig_Obj_t, i int32) *Aig_Obj_t { return pObjMap[nFs*pObj.Id + i];  }
func Aig_ObjSetFrames(pObjMap **Aig_Obj_t, nFs int32, pObj *Aig_Obj_t, i int32, pNode *Aig_Obj_t) { pObjMap[nFs*pObj.Id + i] = pNode; }

func Aig_ObjChild0Frames(pObjMap **Aig_Obj_t, nFs int32, pObj *Aig_Obj_t, i int32) *Aig_Obj_t { if Aig_ObjFanin0(pObj) {
    return Aig_NotCond(Aig_ObjFrames(pObjMap,nFs,Aig_ObjFanin0(pObj),i), Aig_ObjFaninC0(pObj));
} else {
    return nil;
}  }
func Aig_ObjChild1Frames(pObjMap **Aig_Obj_t, nFs int32, pObj *Aig_Obj_t, i int32) *Aig_Obj_t { if Aig_ObjFanin1(pObj) {
    return Aig_NotCond(Aig_ObjFrames(pObjMap,nFs,Aig_ObjFanin1(pObj),i), Aig_ObjFaninC1(pObj));
} else {
    return nil;
}  }

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Performs timeframe expansion of the AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManFrames(pAig *Aig_Man_t, nFs int32, fInit int32, fOuts int32, fRegs int32, fEnlarge int32, ppObjMap ***Aig_Obj_t) *Aig_Man_t {
    var pFrames *Aig_Man_t;
    var pObj, pObjLi, pObjLo, pObjNew *Aig_Obj_t;
    var pObjMap **Aig_Obj_t;
    var i, f int32;

    // create mapping for the frames nodes
    // inlining of macro ABC_ALLOC
    pObjMap  = ((Aig_Obj_t * *) malloc(sizeof(Aig_Obj_t *) * (size_t)(nFs * Aig_ManObjNumMax(pAig))));
    memset( pObjMap, 0, sizeof(Aig_Obj_t *) * nFs * Aig_ManObjNumMax(pAig) );

    // start the fraig package
    pFrames = Aig_ManStart( Aig_ManObjNumMax(pAig) * nFs );
    pFrames.pName = Abc_UtilStrsav( pAig.pName );
    pFrames.pSpec = Abc_UtilStrsav( pAig.pSpec );
    // map constant nodes
    for  f = 0; f < nFs; f++  {
        Aig_ObjSetFrames( pObjMap, nFs, Aig_ManConst1(pAig), f, Aig_ManConst1(pFrames) );
    }
    // create PI nodes for the frames
    for  f = 0; f < nFs; f++  {
        // inlining of macro Aig_ManForEachPiSeq
        // inlining of macro Vec_PtrForEachEntryStop
        for  i = 0; (i < Aig_ManCiNum(pAig)-Aig_ManRegNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCis, i)), 1); i++  {
            Aig_ObjSetFrames( pObjMap, nFs, pObj, f, Aig_ObjCreateCi(pFrames) );
        }
    }
    // set initial state for the latches
    if fInit != 0 {
        // inlining of macro Aig_ManForEachLoSeq
        // inlining of macro Vec_PtrForEachEntryStart
        for  i = Aig_ManCiNum(pAig)-Aig_ManRegNum(pAig); (i < Vec_PtrSize(pAig.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCis, i)), 1); i++  {
            Aig_ObjSetFrames( pObjMap, nFs, pObj, 0, Aig_ManConst0(pFrames) );
        }
    
    } else {
        // inlining of macro Aig_ManForEachLoSeq
        // inlining of macro Vec_PtrForEachEntryStart
        for  i = Aig_ManCiNum(pAig)-Aig_ManRegNum(pAig); (i < Vec_PtrSize(pAig.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCis, i)), 1); i++  {
            Aig_ObjSetFrames( pObjMap, nFs, pObj, 0, Aig_ObjCreateCi(pFrames) );
        }
    }

    // add timeframes
    for  f = 0; f < nFs; f++  {
//        printf( "Frame = %d.\n", f );
        // add internal nodes of this frame
        // inlining of macro Aig_ManForEachNode
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
//            Aig_Obj_t * pFanin0 = Aig_ObjChild0Frames(pObjMap,nFs,pObj,f);
//            Aig_Obj_t * pFanin1 = Aig_ObjChild1Frames(pObjMap,nFs,pObj,f);
//            printf( "Node = %3d.  Fanin0 = %3d. Fanin1 = %3d.\n", pObj->Id, Aig_Regular(pFanin0)->Id, Aig_Regular(pFanin1)->Id );
            pObjNew = Aig_And( pFrames, Aig_ObjChild0Frames(pObjMap,nFs,pObj,f), Aig_ObjChild1Frames(pObjMap,nFs,pObj,f) );
            Aig_ObjSetFrames( pObjMap, nFs, pObj, f, pObjNew );
        }
        }
        // set the latch inputs and copy them into the latch outputs of the next frame
        // inlining of macro Aig_ManForEachLiLoSeq
        for  i = 0; (i < Aig_ManRegNum(pAig)) && (((pObjLi) = Aig_ManLi(pAig, i)), 1)            && (((pObjLo)=Aig_ManLo(pAig, i)), 1); i++  {
            pObjNew = Aig_ObjChild0Frames(pObjMap,nFs,pObjLi,f);
            if f < nFs - 1 {
                Aig_ObjSetFrames( pObjMap, nFs, pObjLo, f+1, pObjNew ); }
        }
    }
    if fOuts != 0 {
        for  f = fEnlarge?nFs-1:0; f < nFs; f++  {
            // inlining of macro Aig_ManForEachPoSeq
            // inlining of macro Vec_PtrForEachEntryStop
            for  i = 0; (i < Aig_ManCoNum(pAig)-Aig_ManRegNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCos, i)), 1); i++  {
                pObjNew = Aig_ObjCreateCo( pFrames, Aig_ObjChild0Frames(pObjMap,nFs,pObj,f) );
                Aig_ObjSetFrames( pObjMap, nFs, pObj, f, pObjNew );
            }
        }
    }
    if fRegs != 0 {
        pFrames.nRegs = pAig.nRegs;
        // inlining of macro Aig_ManForEachLiSeq
        // inlining of macro Vec_PtrForEachEntryStart
        for  i = Aig_ManCoNum(pAig)-Aig_ManRegNum(pAig); (i < Vec_PtrSize(pAig.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCos, i)), 1); i++  {
            pObjNew = Aig_ObjCreateCo( pFrames, Aig_ObjChild0Frames(pObjMap,nFs,pObj,fEnlarge?0:nFs-1) );
            Aig_ObjSetFrames( pObjMap, nFs, pObj, nFs-1, pObjNew );
        }
        Aig_ManSetRegNum( pFrames, Aig_ManRegNum(pAig) );
    }
    Aig_ManCleanup( pFrames );
    // return the new manager
    if ppObjMap {
        *ppObjMap = pObjMap;
    } else {
        // inlining of macro ABC_FREE
        ((pObjMap) ? (free((char *) (pObjMap)), (pObjMap) = 0) : 0); }
    return pFrames;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

