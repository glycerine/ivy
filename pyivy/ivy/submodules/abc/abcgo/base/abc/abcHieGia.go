package abc

import "fmt"

/**CFile****************************************************************

  FileName    [abcHieGia.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Network and node package.]

  Synopsis    [Procedures to handle hierarchy.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: abcHieGia.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "abc.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Transfers the AIG from one manager into another.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NodeStrashToGia_rec(pNew *Gia_Man_t, pObj *Hop_Obj_t) {
    assert(!Hop_IsComplement(pObj), "!Hop_IsComplement(pObj)");
    if !Hop_ObjIsNode(pObj) || Hop_ObjIsMarkA(pObj) {
        return; }
    Abc_NodeStrashToGia_rec( pNew, Hop_ObjFanin0(pObj) ); 
    Abc_NodeStrashToGia_rec( pNew, Hop_ObjFanin1(pObj) );
    pObj.iData = Gia_ManHashAnd( pNew, Hop_ObjChild0CopyI(pObj), Hop_ObjChild1CopyI(pObj) );
    assert(!Hop_ObjIsMarkA(pObj), "!Hop_ObjIsMarkA(pObj)"); // loop detection
    Hop_ObjSetMarkA( pObj );
}
func Abc_NodeStrashToGia(pNew *Gia_Man_t, pNode *Abc_Obj_t) int32 {
    var pMan *Hop_Man_t  = (Hop_Man_t *)pNode.pNtk.pManFunc;
    var pRoot *Hop_Obj_t  = (Hop_Obj_t *)pNode.pData;
    var pFanin *Abc_Obj_t;  var i int32;
    assert(Abc_ObjIsNode(pNode), "Abc_ObjIsNode(pNode)");
    assert(Abc_NtkHasAig(pNode.pNtk) && !Abc_NtkIsStrash(pNode.pNtk), "Abc_NtkHasAig(pNode.pNtk) && !Abc_NtkIsStrash(pNode.pNtk)");
    // check the constant case
    if Abc_NodeIsConst(pNode) || Hop_Regular(pRoot) == Hop_ManConst1(pMan) {
        return Abc_LitNotCond( 1, Hop_IsComplement(pRoot) ); }
    // set elementary variables
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pNode)) && (((pFanin) = Abc_ObjFanin(pNode, i)), 1); i++  {
        assert(pFanin.iTemp != -1, "pFanin.iTemp != -1");
    }
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pNode)) && (((pFanin) = Abc_ObjFanin(pNode, i)), 1); i++  {
        Hop_IthVar(pMan, i).iData = pFanin.iTemp;
    }
    // strash the AIG of this node
    Abc_NodeStrashToGia_rec( pNew, Hop_Regular(pRoot) );
    Hop_ConeUnmark_rec( Hop_Regular(pRoot) );
    return Abc_LitNotCond( Hop_Regular(pRoot).iData, Hop_IsComplement(pRoot) );
}


/**Function*************************************************************

  Synopsis    [Flattens the logic hierarchy of the netlist.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkFlattenHierarchyGia2_rec(pNew *Gia_Man_t, pNtk *Abc_Ntk_t, pCounter *int32, vBufs *Vec_Int_t) {
    var vDfs *Vec_Ptr_t  = (Vec_Ptr_t *)pNtk.pData;
    var pObj, pTerm *Abc_Obj_t; 
    var i, k int32; (*pCounter)++;
    //printf( "[%d:%d] ", Abc_NtkPiNum(pNtk), Abc_NtkPoNum(pNtk) );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vDfs)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vDfs, i)), 1); i++  {
        if Abc_ObjIsNode(pObj) {
            Abc_ObjFanout0(pObj).iTemp = Abc_NodeStrashToGia( pNew, pObj );
        } else {
            var iBufStart int32  = Gia_ManBufNum(pNew);
            var pModel *Abc_Ntk_t  = (Abc_Ntk_t *)pObj.pData;
            assert(!Abc_ObjIsLatch(pObj), "!Abc_ObjIsLatch(pObj)");
            assert(Abc_NtkPiNum(pModel) == Abc_ObjFaninNum(pObj), "Abc_NtkPiNum(pModel) == Abc_ObjFaninNum(pObj)");
            assert(Abc_NtkPoNum(pModel) == Abc_ObjFanoutNum(pObj), "Abc_NtkPoNum(pModel) == Abc_ObjFanoutNum(pObj)");
            Abc_NtkFillTemp( pModel );
            // inlining of macro Abc_ObjForEachFanin
            for  k = 0; (k < Abc_ObjFaninNum(pObj)) && (((pTerm) = Abc_ObjFanin(pObj, k)), 1); k++  {
                assert(Abc_ObjIsNet(Abc_ObjFanin0(pTerm)), "Abc_ObjIsNet(Abc_ObjFanin0(pTerm))");
                Abc_ObjFanout0(Abc_NtkPi(pModel, k)).iTemp = Abc_ObjFanin0(pTerm).iTemp;
            }
            if vBufs != nil {
                // inlining of macro Abc_ObjForEachFanin
                for  k = 0; (k < Abc_ObjFaninNum(pObj)) && (((pTerm) = Abc_ObjFanin(pObj, k)), 1); k++  {
                    Abc_ObjFanout0(Abc_NtkPi(pModel, k)).iTemp = Gia_ManAppendBuf( pNew, Abc_ObjFanout0(Abc_NtkPi(pModel, k)).iTemp );
                } }
            Abc_NtkFlattenHierarchyGia2_rec( pNew, pModel, pCounter, vBufs );
            if vBufs != nil {
                // inlining of macro Abc_ObjForEachFanout
                for  k = 0; (k < Abc_ObjFanoutNum(pObj)) && (((pTerm) = Abc_ObjFanout(pObj, k)), 1); k++  {
                    Abc_ObjFanin0(Abc_NtkPo(pModel, k)).iTemp = Gia_ManAppendBuf( pNew, Abc_ObjFanin0(Abc_NtkPo(pModel, k)).iTemp );
                } }
            // inlining of macro Abc_ObjForEachFanout
            for  k = 0; (k < Abc_ObjFanoutNum(pObj)) && (((pTerm) = Abc_ObjFanout(pObj, k)), 1); k++  {
                assert(Abc_ObjIsNet(Abc_ObjFanout0(pTerm)), "Abc_ObjIsNet(Abc_ObjFanout0(pTerm))");
                Abc_ObjFanout0(pTerm).iTemp = Abc_ObjFanin0(Abc_NtkPo(pModel, k)).iTemp;
            }
            // save buffers
            if vBufs == nil {
                continue; }
            Vec_IntPush( vBufs, iBufStart );
            Vec_IntPush( vBufs, Abc_NtkPiNum(pModel) );
            Vec_IntPush( vBufs, Gia_ManBufNum(pNew) - Abc_NtkPoNum(pModel) );
            Vec_IntPush( vBufs, Abc_NtkPoNum(pModel) );
        }
    }
}
func Abc_NtkFlattenHierarchyGia2(pNtk *Abc_Ntk_t) *Gia_Man_t {
    var fUseBufs int32  = 1;
    var fUseInter int32;
    var pNew, pTemp *Gia_Man_t; 
    var pModel *Abc_Ntk_t;
    var pTerm *Abc_Obj_t;
    var i int32;
    var Counter int32 = -1;
    assert(Abc_NtkIsNetlist(pNtk), "Abc_NtkIsNetlist(pNtk)");
//    Abc_NtkPrintBoxInfo( pNtk );
    Abc_NtkFillTemp( pNtk );

    // start the manager
    pNew = Gia_ManStart( Abc_NtkObjNumMax(pNtk) );
    pNew.pName = Abc_UtilStrsav(pNtk.pName);
    pNew.pSpec = Abc_UtilStrsav(pNtk.pSpec);
    if fUseBufs != 0 {
        pNew.vBarBufs = Vec_IntAlloc( 1000 ); }

    // create PIs and buffers
    // inlining of macro Abc_NtkForEachPi
    for  i = 0; (i < Abc_NtkPiNum(pNtk)) && (((pTerm) = Abc_NtkPi(pNtk, i)), 1); i++  {
        pTerm.iTemp = Gia_ManAppendCi( pNew );
    }
    // inlining of macro Abc_NtkForEachPi
    for  i = 0; (i < Abc_NtkPiNum(pNtk)) && (((pTerm) = Abc_NtkPi(pNtk, i)), 1); i++  {
        Abc_ObjFanout0(pTerm).iTemp = fUseInter ? Gia_ManAppendBuf(pNew, pTerm.iTemp) : pTerm.iTemp;
    }

    // create DFS order of nets
    if !pNtk.pDesign {
        pNtk.pData = Abc_NtkDfsWithBoxes( pNtk );
    } else {
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(pNtk.pDesign.vModules)) && (((pModel) = (Abc_Ntk_t *)Vec_PtrEntry(pNtk.pDesign.vModules, i)), 1); i++  {
            pModel.pData = Abc_NtkDfsWithBoxes( pModel );
        } }

    // call recursively
    Gia_ManHashAlloc( pNew );
    Abc_NtkFlattenHierarchyGia2_rec( pNew, pNtk, &Counter, pNew.vBarBufs );
    Gia_ManHashStop( pNew );
    fmt.Printf("Hierarchy reader flattened %d instances of logic boxes.\n", Counter);

    // delete DFS order of nets
    if !pNtk.pDesign {
        Vec_PtrFreeP( (Vec_Ptr_t **)&pNtk.pData );
    } else {
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(pNtk.pDesign.vModules)) && (((pModel) = (Abc_Ntk_t *)Vec_PtrEntry(pNtk.pDesign.vModules, i)), 1); i++  {
            Vec_PtrFreeP( (Vec_Ptr_t **)&pModel.pData );
        } }

    // create buffers and POs
    // inlining of macro Abc_NtkForEachPo
    for  i = 0; (i < Abc_NtkPoNum(pNtk)) && (((pTerm) = Abc_NtkPo(pNtk, i)), 1); i++  {
        if fUseInter != 0 {
            pTerm.iTemp = Gia_ManAppendBuf(pNew, Abc_ObjFanin0(pTerm).iTemp);
        } else {
            pTerm.iTemp = Abc_ObjFanin0(pTerm).iTemp;
        }
    }
    // inlining of macro Abc_NtkForEachPo
    for  i = 0; (i < Abc_NtkPoNum(pNtk)) && (((pTerm) = Abc_NtkPo(pNtk, i)), 1); i++  {
        Gia_ManAppendCo( pNew, pTerm.iTemp );
    }

    // save buffers
    if fUseInter != 0 {
        Vec_IntPush( pNew.vBarBufs, 0 );
        Vec_IntPush( pNew.vBarBufs, Abc_NtkPiNum(pNtk) );
        Vec_IntPush( pNew.vBarBufs, Gia_ManBufNum(pNew) - Abc_NtkPoNum(pNtk) );
        Vec_IntPush( pNew.vBarBufs, Abc_NtkPoNum(pNtk) );
    }
    if fUseBufs != 0 {
        Vec_IntPrint( pNew.vBarBufs ); }

    // cleanup
    pNew = Gia_ManCleanup( pTemp = pNew );
    Gia_ManStop( pTemp );
    return pNew;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManPrintBarBufDrivers(p *Gia_Man_t) {
    var vMap, vFan, vCrits *Vec_Int_t;
    var pObj *Gia_Obj_t;
    var i int32;
    var iFanin int32;
    var CountCrit [2]int32 ;
    var CountFans [2]int32 ;
    // map barbuf drivers into barbuf literals of the first barbuf driven by them
    vMap = Vec_IntStart( Gia_ManObjNum(p) );
    vFan = Vec_IntStart( Gia_ManObjNum(p) );
    vCrits = Vec_IntAlloc( 100 );
    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        // count fanouts
        if Gia_ObjIsBuf(pObj) || Gia_ObjIsCo(pObj) {
            Vec_IntAddToEntry( vFan, Gia_ObjFaninId0(pObj, i), 1 );
        } else if Gia_ObjIsAnd(pObj) {
            Vec_IntAddToEntry( vFan, Gia_ObjFaninId0(pObj, i), 1 );
            Vec_IntAddToEntry( vFan, Gia_ObjFaninId1(pObj, i), 1 );
        }
        // count critical barbufs
        if Gia_ObjIsBuf(pObj) {
            iFanin = Gia_ObjFaninId0( pObj, i );
            if iFanin == 0 || Vec_IntEntry(vMap, iFanin) != 0 {
                CountCrit[int32((iFanin != 0))]++;
                Vec_IntPush( vCrits, i );
                continue;
            }
            Vec_IntWriteEntry( vMap, iFanin, Abc_Var2Lit(i, Gia_ObjFaninC0(pObj)) );
        }
    }
    // check fanouts of the critical barbufs
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vCrits)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vCrits,i))); i++  {
        assert(Gia_ObjIsBuf(pObj), "Gia_ObjIsBuf(pObj)");
        if Vec_IntEntry(vFan, i) == 0 {
            continue; }
        iFanin = Gia_ObjFaninId0p( p, pObj );
        CountFans[int32((iFanin != 0))]++;
    }
    fmt.Printf("Detected %d const (out of %d) and %d shared (out of %d) barbufs with fanout.\n", CountFans[0], CountCrit[0], CountFans[1], CountCrit[1]);
    Vec_IntFree( vMap );
    Vec_IntFree( vFan );
    Vec_IntFree( vCrits );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManPatchBufDriver(p *Gia_Man_t, iBuf int32, iLit0 int32) {
    var pObjBuf *Gia_Obj_t   = Gia_ManObj( p, iBuf );
    assert(Gia_ObjIsBuf(pObjBuf), "Gia_ObjIsBuf(pObjBuf)");
    assert(Gia_ObjId(p, pObjBuf) > Abc_Lit2Var(iLit0), "Gia_ObjId(p, pObjBuf) > Abc_Lit2Var(iLit0)");
    pObjBuf.iDiff1  = pObjBuf.iDiff0  = Gia_ObjId(p, pObjBuf) - Abc_Lit2Var(iLit0);
    pObjBuf.fCompl1 = pObjBuf.fCompl0 = Abc_LitIsCompl(iLit0);
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManSweepHierarchy(p *Gia_Man_t) *Gia_Man_t {
    var vMap *Vec_Int_t  = Vec_IntStart( Gia_ManObjNum(p) );
    var pNew, pTemp *Gia_Man_t;
    var pObj, pObjNew, pObjNewR *Gia_Obj_t;
    var i int32;
    var iFanin int32;
    var CountReals [2]int32 ;

    // duplicate AIG while propagating constants and equivalences 
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManConst0(p).Value = 0;
    Gia_ManHashAlloc( pNew );
    // inlining of macro Gia_ManForEachObj1
    for  i = 1; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        if Gia_ObjIsBuf(pObj) {
            pObj.Value = Gia_ManAppendBuf( pNew, Gia_ObjFanin0Copy(pObj) );
            pObjNew = Gia_ManObj( pNew, Abc_Lit2Var(pObj.Value) );
            iFanin = Gia_ObjFaninId0p( pNew, pObjNew );
            if iFanin == 0 {
                pObj.Value = Gia_ObjFaninC0(pObjNew);
                CountReals[0]++;
                Gia_ManPatchBufDriver( pNew, Gia_ObjId(pNew, pObjNew), 0 );
            
            } else if Vec_IntEntry(vMap, iFanin) {
                pObjNewR = Gia_ManObj( pNew, Vec_IntEntry(vMap, iFanin) );
                pObj.Value = Abc_Var2Lit( Vec_IntEntry(vMap, iFanin), Gia_ObjFaninC0(pObjNewR) ^ Gia_ObjFaninC0(pObjNew) );
                CountReals[1]++;
                Gia_ManPatchBufDriver( pNew, Gia_ObjId(pNew, pObjNew), 0 );
            
            } else {
                Vec_IntWriteEntry( vMap, iFanin, Gia_ObjId(pNew, pObjNew) ); }
        
        } else if Gia_ObjIsAnd(pObj) {
            pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
        } else if Gia_ObjIsCi(pObj) {
            pObj.Value = Gia_ManAppendCi( pNew );
        } else if Gia_ObjIsCo(pObj) {
            pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) ); }
    }
    Gia_ManHashStop( pNew );
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(p) );
    pNew = Gia_ManCleanup( pTemp = pNew );
    Gia_ManStop( pTemp );
//    printf( "Updated %d const and %d shared.\n", CountReals[0], CountReals[1] );
    Vec_IntFree( vMap );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Flattens the logic hierarchy of the netlist.]

  Description [This procedure requires that models are uniqified.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManFlattenLogicPrepare(pNtk *Abc_Ntk_t) int32 {
    var pTerm, pBox *Abc_Obj_t; 
    var i, k int32;
    Abc_NtkFillTemp( pNtk );
    // inlining of macro Abc_NtkForEachPi
    for  i = 0; (i < Abc_NtkPiNum(pNtk)) && (((pTerm) = Abc_NtkPi(pNtk, i)), 1); i++  {
        pTerm.iData = i;
    }
    // inlining of macro Abc_NtkForEachPo
    for  i = 0; (i < Abc_NtkPoNum(pNtk)) && (((pTerm) = Abc_NtkPo(pNtk, i)), 1); i++  {
        pTerm.iData = i;
    }
    // inlining of macro Abc_NtkForEachBox
    for  i = 0; (i < Vec_PtrSize((pNtk).vBoxes)) && (((pBox) = Abc_NtkBox(pNtk, i)), 1); i++  {
        assert(!Abc_ObjIsLatch(pBox), "!Abc_ObjIsLatch(pBox)");
        // inlining of macro Abc_ObjForEachFanin
        for  k = 0; (k < Abc_ObjFaninNum(pBox)) && (((pTerm) = Abc_ObjFanin(pBox, k)), 1); k++  {
            pTerm.iData = k;
        }
        // inlining of macro Abc_ObjForEachFanout
        for  k = 0; (k < Abc_ObjFanoutNum(pBox)) && (((pTerm) = Abc_ObjFanout(pBox, k)), 1); k++  {
            pTerm.iData = k;
        }
    }
    return Abc_NtkPiNum(pNtk) + Abc_NtkPoNum(pNtk);
}
func Abc_NtkFlattenHierarchyGia_rec(pNew *Gia_Man_t, vSupers *Vec_Ptr_t, pObj *Abc_Obj_t, vBuffers *Vec_Ptr_t) int32 {
    var pModel *Abc_Ntk_t;
    var pBox, pFanin *Abc_Obj_t;  
    var iLit, i int32;
    if pObj.iTemp != -1 {
        return pObj.iTemp; }
    if Abc_ObjIsNet(pObj) || Abc_ObjIsPo(pObj) || Abc_ObjIsBi(pObj) {
        return (pObj.iTemp = Abc_NtkFlattenHierarchyGia_rec(pNew, vSupers, Abc_ObjFanin0(pObj), vBuffers)); }
    if Abc_ObjIsPi(pObj) {
        pBox   = (Abc_Obj_t *)Vec_PtrPop( vSupers );
        pModel = Abc_ObjModel(pBox);
        //printf( "   Exiting %s\n", Abc_NtkName(pModel) );
        assert(Abc_ObjFaninNum(pBox) == Abc_NtkPiNum(pModel), "Abc_ObjFaninNum(pBox) == Abc_NtkPiNum(pModel)");
        assert(pObj.iData >= 0 && pObj.iData < Abc_NtkPiNum(pModel), "pObj.iData >= 0 && pObj.iData < Abc_NtkPiNum(pModel)");
        pFanin = Abc_ObjFanin( pBox, pObj.iData );
        iLit   = Abc_NtkFlattenHierarchyGia_rec( pNew, vSupers, pFanin, vBuffers );
        Vec_PtrPush( vSupers, pBox );
        //if ( vBuffers ) Vec_PtrPush( vBuffers, pFanin ); // save BI
        if vBuffers != nil { Vec_PtrPush( vBuffers, pObj ); }   // save PI
        return (pObj.iTemp = (vBuffers ? Gia_ManAppendBuf(pNew, iLit) : iLit));
    }
    if Abc_ObjIsBo(pObj) {
        pBox   = Abc_ObjFanin0(pObj);
        assert(Abc_ObjIsBox(pBox), "Abc_ObjIsBox(pBox)");
        Vec_PtrPush( vSupers, pBox );
        pModel = Abc_ObjModel(pBox);
        //printf( "Entering %s\n", Abc_NtkName(pModel) );
        assert(Abc_ObjFanoutNum(pBox) == Abc_NtkPoNum(pModel), "Abc_ObjFanoutNum(pBox) == Abc_NtkPoNum(pModel)");
        assert(pObj.iData >= 0 && pObj.iData < Abc_NtkPoNum(pModel), "pObj.iData >= 0 && pObj.iData < Abc_NtkPoNum(pModel)");
        pFanin = Abc_NtkPo( pModel, pObj.iData );
        iLit   = Abc_NtkFlattenHierarchyGia_rec( pNew, vSupers, pFanin, vBuffers );
        Vec_PtrPop( vSupers );
        //if ( vBuffers ) Vec_PtrPush( vBuffers, pObj );   // save BO
        if vBuffers != nil { Vec_PtrPush( vBuffers, pFanin ); } // save PO
        return (pObj.iTemp = (vBuffers ? Gia_ManAppendBuf(pNew, iLit) : iLit));
    }
    assert(Abc_ObjIsNode(pObj), "Abc_ObjIsNode(pObj)");
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pObj)) && (((pFanin) = Abc_ObjFanin(pObj, i)), 1); i++  {
        Abc_NtkFlattenHierarchyGia_rec( pNew, vSupers, pFanin, vBuffers );
    }
    return (pObj.iTemp = Abc_NodeStrashToGia( pNew, pObj ));
}
func Abc_NtkFlattenHierarchyGia(pNtk *Abc_Ntk_t, pvBuffers **Vec_Ptr_t, fVerbose int32) *Gia_Man_t {
    var fUseBufs int32  = 1;
    var pNew, pTemp *Gia_Man_t; 
    var pModel *Abc_Ntk_t;
    var pTerm *Abc_Obj_t;
    var vSupers *Vec_Ptr_t;
    var vBuffers *Vec_Ptr_t;
    if fUseBufs != 0 {
        vBuffers = Vec_PtrAlloc(1000);
    } else {
        vBuffers = nil;
    }
    var i, Counter int32; 
    assert(Abc_NtkIsNetlist(pNtk), "Abc_NtkIsNetlist(pNtk)");
//    Abc_NtkPrintBoxInfo( pNtk );

    // set the PI/PO numbers
    Counter -= Abc_NtkPiNum(pNtk) + Abc_NtkPoNum(pNtk);
    if !pNtk.pDesign {
        Counter += Gia_ManFlattenLogicPrepare( pNtk );
    } else {
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(pNtk.pDesign.vModules)) && (((pModel) = (Abc_Ntk_t *)Vec_PtrEntry(pNtk.pDesign.vModules, i)), 1); i++  {
            Counter += Gia_ManFlattenLogicPrepare( pModel );
        } }

    // start the manager
    pNew = Gia_ManStart( Abc_NtkObjNumMax(pNtk) );
    pNew.pName = Abc_UtilStrsav(pNtk.pName);
    pNew.pSpec = Abc_UtilStrsav(pNtk.pSpec);

    // create PIs and buffers
    // inlining of macro Abc_NtkForEachPi
    for  i = 0; (i < Abc_NtkPiNum(pNtk)) && (((pTerm) = Abc_NtkPi(pNtk, i)), 1); i++  {
        pTerm.iTemp = Gia_ManAppendCi( pNew );
    }

    // call recursively
    vSupers = Vec_PtrAlloc( 100 );
    Gia_ManHashAlloc( pNew );
    // inlining of macro Abc_NtkForEachPo
    for  i = 0; (i < Abc_NtkPoNum(pNtk)) && (((pTerm) = Abc_NtkPo(pNtk, i)), 1); i++  {
        Abc_NtkFlattenHierarchyGia_rec( pNew, vSupers, pTerm, vBuffers );
    }
    Gia_ManHashStop( pNew );
    Vec_PtrFree( vSupers );
    fmt.Printf("Hierarchy reader flattened %d instances of boxes and added %d barbufs (out of %d).\n", pNtk.pDesign ? Vec_PtrSize(pNtk.pDesign.vModules)-1 : 0, Vec_PtrSize(vBuffers), Counter);

    // create buffers and POs
    // inlining of macro Abc_NtkForEachPo
    for  i = 0; (i < Abc_NtkPoNum(pNtk)) && (((pTerm) = Abc_NtkPo(pNtk, i)), 1); i++  {
        Gia_ManAppendCo( pNew, pTerm.iTemp );
    }

    if pvBuffers {
        *pvBuffers = vBuffers;
    } else {
        Vec_PtrFreeP( &vBuffers ); }

    // cleanup
    pNew = Gia_ManCleanup( pTemp = pNew );
    Gia_ManStop( pTemp );
//    Gia_ManPrintStats( pNew, NULL );
    pNew = Gia_ManSweepHierarchy( pTemp = pNew );
    Gia_ManStop( pTemp );
//    Gia_ManPrintStats( pNew, NULL );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Inserts the result of mapping into logic hierarchy.]

  Description [When this procedure is called PIs/POs of pNtk
  point to the corresponding nodes in network with barbufs.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManInsertOne_rec(pNtk *Abc_Ntk_t, pNew *Abc_Ntk_t, pObj *Abc_Obj_t) *Abc_Obj_t {
    var pFanin *Abc_Obj_t;  var i int32;
    if pObj == nil {
        return Abc_NtkCreateNodeConst0( pNtk ); }
    assert(Abc_ObjNtk(pObj) == pNew, "Abc_ObjNtk(pObj) == pNew");
    if pObj.pCopy {
        return pObj.pCopy; }
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pObj)) && (((pFanin) = Abc_ObjFanin(pObj, i)), 1); i++  {
        Gia_ManInsertOne_rec( pNtk, pNew, pFanin );
    }
    pObj.pCopy = Abc_NtkDupObj( pNtk, pObj, 0 );
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pObj)) && (((pFanin) = Abc_ObjFanin(pObj, i)), 1); i++  {
        Abc_ObjAddFanin( pObj, pFanin );
    }
    return pObj.pCopy;
}
func Gia_ManInsertOne(pNtk *Abc_Ntk_t, pNew *Abc_Ntk_t) {
    var pObj, pBox *Abc_Obj_t;  var i, k int32;
    assert(!Abc_NtkHasMapping(pNtk), "!Abc_NtkHasMapping(pNtk)");
    assert(Abc_NtkHasMapping(pNew), "Abc_NtkHasMapping(pNew)");
    // check that PIs point to barbufs
    // inlining of macro Abc_NtkForEachPi
    for  i = 0; (i < Abc_NtkPiNum(pNtk)) && (((pObj) = Abc_NtkPi(pNtk, i)), 1); i++  {
        assert(!pObj.pCopy || Abc_ObjNtk(pObj.pCopy) == pNew, "!pObj.pCopy || Abc_ObjNtk(pObj.pCopy) == pNew");
    }
    // make barbufs point to box outputs
    // inlining of macro Abc_NtkForEachBox
    for  i = 0; (i < Vec_PtrSize((pNtk).vBoxes)) && (((pBox) = Abc_NtkBox(pNtk, i)), 1); i++  {
        // inlining of macro Abc_ObjForEachFanout
        for  k = 0; (k < Abc_ObjFanoutNum(pBox)) && (((pObj) = Abc_ObjFanout(pBox, k)), 1); k++  {
            pObj.pCopy = Abc_NtkPo(Abc_ObjModel(pBox), k).pCopy;
            assert(!pObj.pCopy || Abc_ObjNtk(pObj.pCopy) == pNew, "!pObj.pCopy || Abc_ObjNtk(pObj.pCopy) == pNew");
        }
    }
    // remove internal nodes
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pObj) == nil || !Abc_ObjIsNode(pObj) {} else {
        Abc_NtkDeleteObj( pObj ); }
    }
    // start traversal from box inputs
    // inlining of macro Abc_NtkForEachBox
    for  i = 0; (i < Vec_PtrSize((pNtk).vBoxes)) && (((pBox) = Abc_NtkBox(pNtk, i)), 1); i++  {
        // inlining of macro Abc_ObjForEachFanin
        for  k = 0; (k < Abc_ObjFaninNum(pBox)) && (((pObj) = Abc_ObjFanin(pBox, k)), 1); k++  {
            if Abc_ObjFaninNum(pObj) == 0 {
                Abc_ObjAddFanin( pObj, Gia_ManInsertOne_rec(pNtk, pNew, Abc_NtkPi(Abc_ObjModel(pBox), k).pCopy) ); }
        }
    }
    // start traversal from primary outputs
    // inlining of macro Abc_NtkForEachPo
    for  i = 0; (i < Abc_NtkPoNum(pNtk)) && (((pObj) = Abc_NtkPo(pNtk, i)), 1); i++  {
        if Abc_ObjFaninNum(pObj) == 0 {
            Abc_ObjAddFanin( pObj, Gia_ManInsertOne_rec(pNtk, pNew, pObj.pCopy) ); }
    }
    // update the functionality manager
    pNtk.pManFunc = pNew.pManFunc;
    pNtk.ntkFunc  = pNew.ntkFunc;
    assert(Abc_NtkHasMapping(pNtk), "Abc_NtkHasMapping(pNtk)");
}
func Abc_NtkInsertHierarchyGia(pNtk *Abc_Ntk_t, pNew *Abc_Ntk_t, fVerbose int32) {
    var vBuffers *Vec_Ptr_t;
    var pGia *Gia_Man_t  = Abc_NtkFlattenHierarchyGia( pNtk, &vBuffers, 0 );
    var pModel *Abc_Ntk_t;  
    var pObj *Abc_Obj_t; 
    var i, k int32;

    assert(Gia_ManPiNum(pGia) == Abc_NtkPiNum(pNtk), "Gia_ManPiNum(pGia) == Abc_NtkPiNum(pNtk)");
    assert(Gia_ManPiNum(pGia) == Abc_NtkPiNum(pNew), "Gia_ManPiNum(pGia) == Abc_NtkPiNum(pNew)");
    assert(Gia_ManPoNum(pGia) == Abc_NtkPoNum(pNtk), "Gia_ManPoNum(pGia) == Abc_NtkPoNum(pNtk)");
    assert(Gia_ManPoNum(pGia) == Abc_NtkPoNum(pNew), "Gia_ManPoNum(pGia) == Abc_NtkPoNum(pNew)");
    assert(Gia_ManBufNum(pGia) == Vec_PtrSize(vBuffers), "Gia_ManBufNum(pGia) == Vec_PtrSize(vBuffers)");
    assert(Gia_ManBufNum(pGia) == pNew.nBarBufs2, "Gia_ManBufNum(pGia) == pNew.nBarBufs2");
    Gia_ManStop( pGia );

    // clean the networks
    if !pNtk.pDesign {
        Abc_NtkCleanCopy( pNtk );
    } else {
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(pNtk.pDesign.vModules)) && (((pModel) = (Abc_Ntk_t *)Vec_PtrEntry(pNtk.pDesign.vModules, i)), 1); i++  {
            Abc_NtkCleanCopy( pModel );
        } }

    // annotate PIs and POs of each network with barbufs
    // inlining of macro Abc_NtkForEachPi
    for  i = 0; (i < Abc_NtkPiNum(pNew)) && (((pObj) = Abc_NtkPi(pNew, i)), 1); i++  {
        Abc_NtkPi(pNtk, i).pCopy = pObj;
    }
    // inlining of macro Abc_NtkForEachPo
    for  i = 0; (i < Abc_NtkPoNum(pNew)) && (((pObj) = Abc_NtkPo(pNew, i)), 1); i++  {
        Abc_NtkPo(pNtk, i).pCopy = pObj;
    }
    // inlining of macro Abc_NtkForEachBarBuf
    for  i = 0; (i < Vec_PtrSize((pNew).vObjs)) && (((pObj) = Abc_NtkObj(pNew, i)), 1); i++  {           if (pObj) == nil || !Abc_ObjIsBarBuf(pObj) {} else {
        ((Abc_Obj_t *)Vec_PtrEntry(vBuffers, k++)).pCopy = pObj; }
    }
    Vec_PtrFree( vBuffers );

    // connect each model
    Abc_NtkCleanCopy( pNew );
    Gia_ManInsertOne( pNtk, pNew );
    if pNtk.pDesign {
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(pNtk.pDesign.vModules)) && (((pModel) = (Abc_Ntk_t *)Vec_PtrEntry(pNtk.pDesign.vModules, i)), 1); i++  {
            if pModel != pNtk {
                Gia_ManInsertOne( pModel, pNew ); }
        } }
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

