package abci

import "fmt"

/**CFile****************************************************************

  FileName    [abcTim.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Network and node package.]

  Synopsis    [Testing hierarchy/timing manager.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: abcTim.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "base/abc/abc.h"
//#include "aig/gia/giaAig.h"
//#include "misc/tim/tim.h"
//#include "opt/dar/dar.h"
//#include "proof/dch/dch.h"
//#include "base/main/main.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

const TIM_TEST_BOX_RATIO = 200

// assume that every TIM_TEST_BOX_RATIO'th object is a white box
func Abc_NodeIsWhiteBox(pObj *Abc_Obj_t) int32 { assert(Abc_ObjIsNode(pObj), "Abc_ObjIsNode(pObj)"); return Abc_ObjId(pObj) % TIM_TEST_BOX_RATIO == 0 && Abc_ObjFaninNum(pObj) > 0 && Abc_ObjFaninNum(pObj) < 10; }

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Derives GIA for the output of the local function of one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkTestTimNodeStrash_rec(pGia *Gia_Man_t, pObj *Hop_Obj_t) {
    assert(!Hop_IsComplement(pObj), "!Hop_IsComplement(pObj)");
    if !Hop_ObjIsNode(pObj) || Hop_ObjIsMarkA(pObj) {
        return; }
    Abc_NtkTestTimNodeStrash_rec( pGia, Hop_ObjFanin0(pObj) ); 
    Abc_NtkTestTimNodeStrash_rec( pGia, Hop_ObjFanin1(pObj) );
    pObj.iData = Gia_ManHashAnd( pGia, Hop_ObjChild0CopyI(pObj), Hop_ObjChild1CopyI(pObj) );
    assert(!Hop_ObjIsMarkA(pObj), "!Hop_ObjIsMarkA(pObj)"); // loop detection
    Hop_ObjSetMarkA( pObj );
}
func Abc_NtkTestTimNodeStrash(pGia *Gia_Man_t, pNode *Abc_Obj_t) int32 {
    var pMan *Hop_Man_t;
    var pRoot *Hop_Obj_t;
    var pFanin *Abc_Obj_t;
    var i int32;
    assert(Abc_ObjIsNode(pNode), "Abc_ObjIsNode(pNode)");
    assert(Abc_NtkIsAigLogic(pNode.pNtk), "Abc_NtkIsAigLogic(pNode.pNtk)");
    // get the local AIG manager and the local root node
    pMan = (Hop_Man_t *)pNode.pNtk.pManFunc;
    pRoot = (Hop_Obj_t *)pNode.pData;
    // check the constant case
    if Abc_NodeIsConst(pNode) || Hop_Regular(pRoot) == Hop_ManConst1(pMan) {
        return !Hop_IsComplement(pRoot); }
    // set elementary variables
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pNode)) && (((pFanin) = Abc_ObjFanin(pNode, i)), 1); i++  {
        Hop_IthVar(pMan, i).iData = pFanin.iTemp;
    }
    // strash the AIG of this node
    Abc_NtkTestTimNodeStrash_rec( pGia, Hop_Regular(pRoot) );
    Hop_ConeUnmark_rec( Hop_Regular(pRoot) );
    // return the final node with complement if needed
    return Abc_LitNotCond( Hop_Regular(pRoot).iData, Hop_IsComplement(pRoot) );
}


      

                                                                        

                                                                                

                
               
                

                

                                                                        
                                                                                        
 
                      
                            
                       
                               
                          
                            
                                   
                         
                    
                            
                                                  
                             
                            
                                     
                                            
                                                                  
                                   
                                                       
     
                     
                                                         
         
                                                   
                                                      
                                
                                                  
                                                                                 
         
                                   
                                                             
                      
                                                         
         
                                                   
                                                      
                                 
                                                                         
         
     
                          
                             
                                     
                                                                          
                   
                            
                                
                   
                                          
                         
                
 

                                                                        

                

                
               
                

                

                                                                        
                                                                           
 
                     
                                     
                                                                    
                                               
                        
                                                                                                
                                       
 

      

/**Function*************************************************************

  Synopsis    [Collect nodes reachable from this box.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkTestTimCollectCone_rec(pObj *Abc_Obj_t, vNodes *Vec_Ptr_t) {
    var pFanin *Abc_Obj_t;
    var i int32;
    if Abc_NodeIsTravIdCurrent( pObj ) {
        return; }
    Abc_NodeSetTravIdCurrent( pObj );
    if Abc_ObjIsCi(pObj) {
        return; }
    assert(Abc_ObjIsNode( pObj ), "Abc_ObjIsNode( pObj )");
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pObj)) && (((pFanin) = Abc_ObjFanin(pObj, i)), 1); i++  {
        Abc_NtkTestTimCollectCone_rec( pFanin, vNodes );
    }
    Vec_PtrPush( vNodes, pObj );
}
func Abc_NtkTestTimCollectCone(pNtk *Abc_Ntk_t, pObj *Abc_Obj_t) *Vec_Ptr_t {
    var vCone *Vec_Ptr_t  = Vec_PtrAlloc( 1000 );
    if pObj != nil {
        // collect for one node
        assert(Abc_ObjIsNode(pObj), "Abc_ObjIsNode(pObj)");
        assert(!Abc_NodeIsTravIdCurrent( pObj ), "!Abc_NodeIsTravIdCurrent( pObj )");
        Abc_NtkTestTimCollectCone_rec( pObj, vCone );
        // remove the node because it is a white box
        Vec_PtrPop( vCone );
    
    } else {
        // collect for all COs
        var pObj *Abc_Obj_t;
        var i int32;
        // inlining of macro Abc_NtkForEachCo
        for  i = 0; (i < Abc_NtkCoNum(pNtk)) && (((pObj) = Abc_NtkCo(pNtk, i)), 1); i++  {
            Abc_NtkTestTimCollectCone_rec( Abc_ObjFanin0(pObj), vCone );
        }
    }
    return vCone;

}

/**Function*************************************************************

  Synopsis    [Create arrival times]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkTestCreateArrivals(nInputs int32) *Vec_Flt_t {
    var p *Vec_Flt_t;
    var i int32;
    p = Vec_FltAlloc( nInputs );
    for  i = 0; i < nInputs; i++  {
        Vec_FltPush( p, 1.0*(i % 10) );
    }
    return p;
}
func Abc_NtkTestCreateRequired(nOutputs int32) *Vec_Flt_t {
    var p *Vec_Flt_t;
    var i int32;
    p = Vec_FltAlloc( nOutputs );
    for  i = 0; i < nOutputs; i++  {
        Vec_FltPush( p, 100.0 + 1.0*i );
    }
    return p;
}

/**Function*************************************************************

  Synopsis    [Derives GIA manager together with the hierachy manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkTestTimDeriveGia(pNtk *Abc_Ntk_t, fVerbose int32) *Gia_Man_t {
    var pTemp *Gia_Man_t;
    var pGia *Gia_Man_t  = nil;
    var pHoles *Gia_Man_t  = nil;
    var pTim *Tim_Man_t  = nil;
    var vGiaCoLits, vGiaCoLits2 *Vec_Int_t;
    var vArrTimes, vReqTimes *Vec_Flt_t;
    var pObj, pFanin *Abc_Obj_t;
    var i, k, Entry, curPi, curPo, BoxUniqueId int32;
    var nBoxFaninMax int32;
    assert(Abc_NtkIsTopo(pNtk), "Abc_NtkIsTopo(pNtk)");
    Abc_NtkFillTemp( pNtk );

    // create white boxes
    curPi = Abc_NtkCiNum(pNtk);
    curPo = Abc_NtkCoNum(pNtk);
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pObj) == nil || !Abc_ObjIsNode(pObj) {} else {
        pObj.fMarkA = Abc_NodeIsWhiteBox( pObj );
        if !pObj.fMarkA {
            continue; }
        nBoxFaninMax  = Abc_MaxInt( nBoxFaninMax, Abc_ObjFaninNum(pObj) );
        curPi++;
        curPo += Abc_ObjFaninNum(pObj);
        if fVerbose != 0 {
            fmt.Printf("Selecting node %6d as white boxes with %d inputs and %d output.\n", i, Abc_ObjFaninNum(pObj), 1); }
    }
    }

    // construct GIA
    pGia = Gia_ManStart( Abc_NtkObjNumMax(pNtk) );
    pHoles = Gia_ManStart( 1000 );
    for  i = 0; i < curPi; i++  {
        Gia_ManAppendCi(pGia);
    }
    for  i = 0; i < nBoxFaninMax; i++  {
        Gia_ManAppendCi(pHoles);
    }
    Gia_ManHashAlloc( pGia );
    Gia_ManHashAlloc( pHoles );

    // construct the timing manager
    pTim = Tim_ManStart( curPi, curPo );

    // assign primary inputs
    curPi = 0;
    curPo = 0;
    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(pNtk)) && (((pObj) = Abc_NtkCi(pNtk, i)), 1); i++  {
        pObj.iTemp = Abc_Var2Lit( Gia_ObjId(pGia, Gia_ManCi(pGia, curPi++)), 0 );
    }
    // create internal nodes in a topologic order from white boxes
    vGiaCoLits = Vec_IntAlloc( 1000 );
    vGiaCoLits2 = Vec_IntAlloc( 1000 );
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pObj) == nil || !Abc_ObjIsNode(pObj) {} else {
        if !pObj.fMarkA { // not a white box
        
            pObj.iTemp = Abc_NtkTestTimNodeStrash( pGia, pObj );
            continue;
        }
        // create box
        BoxUniqueId = Abc_ObjFaninNum(pObj); // in this case, the node size is the ID of its delay table
        Tim_ManCreateBox( pTim, curPo, Abc_ObjFaninNum(pObj), curPi, 1, BoxUniqueId, 0 );
        curPo += Abc_ObjFaninNum(pObj);

        // handle box inputs
        // inlining of macro Abc_ObjForEachFanin
        for  k = 0; (k < Abc_ObjFaninNum(pObj)) && (((pFanin) = Abc_ObjFanin(pObj, k)), 1); k++  {
            // save CO drivers for the AIG
            Vec_IntPush( vGiaCoLits, pFanin.iTemp );
            // load CI nodes for the Holes
            pFanin.iTemp = Abc_Var2Lit( Gia_ObjId(pHoles, Gia_ManCi(pHoles, k)), 0 );
        }

        // handle logic of the box
        pObj.iTemp = Abc_NtkTestTimNodeStrash( pHoles, pObj );

        // handle box outputs
        // save CO drivers for the Holes
        Vec_IntPush( vGiaCoLits2, pObj.iTemp );
//        Gia_ManAppendCo( pHoles, pObj->iTemp );
        // load CO drivers for the AIG
        pObj.iTemp = Abc_Var2Lit( Gia_ObjId(pGia, Gia_ManCi(pGia, curPi++)), 0 );
    }
    }
    Abc_NtkCleanMarkA( pNtk );
    // create COs of the AIG
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(pNtk)) && (((pObj) = Abc_NtkCo(pNtk, i)), 1); i++  {
        Gia_ManAppendCo( pGia, Abc_ObjFanin0(pObj).iTemp );
    }
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vGiaCoLits)) && (((Entry) = Vec_IntEntry(vGiaCoLits, i)), 1); i++  {
        Gia_ManAppendCo( pGia, Entry );
    }
    Vec_IntFree( vGiaCoLits );
    // second AIG
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vGiaCoLits2)) && (((Entry) = Vec_IntEntry(vGiaCoLits2, i)), 1); i++  {
        Gia_ManAppendCo( pHoles, Entry );
    }
    Vec_IntFree( vGiaCoLits2 );
    // check parameters
    curPo += Abc_NtkPoNum( pNtk );
    assert(curPi == Gia_ManPiNum(pGia), "curPi == Gia_ManPiNum(pGia)");
    assert(curPo == Gia_ManPoNum(pGia), "curPo == Gia_ManPoNum(pGia)");
    // finalize GIA
    Gia_ManHashStop( pGia );
    Gia_ManSetRegNum( pGia, 0 );
    Gia_ManHashStop( pHoles );
    Gia_ManSetRegNum( pHoles, 0 );

    // clean up GIA
    pGia = Gia_ManCleanup( pTemp = pGia );
    Gia_ManStop( pTemp );
    pHoles = Gia_ManCleanup( pTemp = pHoles );
    Gia_ManStop( pTemp );

    // attach the timing manager
    assert(pGia.pManTime == nil, "pGia.pManTime == nil");
    pGia.pManTime = pTim;

    // derive hierarchy manager from box info and input/output arrival/required info
    vArrTimes = Abc_NtkTestCreateArrivals( Abc_NtkPiNum(pNtk) );
    vReqTimes = Abc_NtkTestCreateRequired( Abc_NtkPoNum(pNtk) );

    Tim_ManPrint( (Tim_Man_t *)pGia.pManTime );
    Tim_ManCreate( (Tim_Man_t *)pGia.pManTime, Abc_FrameReadLibBox(), vArrTimes, vReqTimes );
    Tim_ManPrint( (Tim_Man_t *)pGia.pManTime );

    Vec_FltFree( vArrTimes );
    Vec_FltFree( vReqTimes );

Gia_AigerWrite( pHoles, "holes00.aig", 0, 0, 0 );

    // return 
    pGia.pAigExtra = pHoles;
    return pGia;
}


/**Function*************************************************************

  Synopsis    [Performs synthesis with or without structural choices.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkTestTimPerformSynthesis(p *Gia_Man_t, fChoices int32) *Gia_Man_t {
    var pGia *Gia_Man_t;
    var pNew, pTemp *Aig_Man_t;
    var Pars Dch_Pars_t;
    var pPars *Dch_Pars_t = &Pars;
    Dch_ManSetDefaultParams( pPars );
    pNew = Gia_ManToAig( p, 0 );
    if fChoices != 0 {
        pNew = Dar_ManChoiceNew( pNew, pPars );
    } else {
        pNew = Dar_ManCompress2( pTemp = pNew, 1, 1, 1, 0, 0 );
        Aig_ManStop( pTemp );
    }
    pGia = Gia_ManFromAig( pNew );
    Aig_ManStop( pNew );
    return pGia;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManVerifyChoices(p *Gia_Man_t) {
    var pObj *Gia_Obj_t;
    var i, iRepr, iNode, fProb int32;
    assert(p.pReprs, "p.pReprs");

    // mark nodes 
    Gia_ManCleanMark0(p);
    // inlining of macro Gia_ManForEachClass
    for  iRepr = 1; iRepr < Gia_ManObjNum(p); iRepr++  { if !Gia_ObjIsHead(p, iRepr) {} else {
        // inlining of macro Gia_ClassForEachObj1
        for  assert(Gia_ObjIsHead(p, iRepr), "Gia_ObjIsHead(p, iRepr)"), iNode = Gia_ObjNext(p, iRepr); iNode > 0; iNode = Gia_ObjNext(p, iNode)  {
            if Gia_ObjIsHead(p, iNode) {
                fmt.Printf("Member %d of choice class %d is a representative.\n", iNode, iRepr);
fProb = 1; }
            if Gia_ManObj( p, iNode ).fMark0 == 1 {
                fmt.Printf("Node %d participates in more than one choice node.\n", iNode);
fProb = 1; }
            Gia_ManObj( p, iNode ).fMark0 = 1;
        } }
    }
    Gia_ManCleanMark0(p);

    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        if Gia_ObjIsAnd(pObj) {
            if Gia_ObjHasRepr(p, Gia_ObjFaninId0(pObj, i)) {
                fmt.Printf("Fanin 0 of AND node %d has a repr.\n", i);
fProb = 1; }
            if Gia_ObjHasRepr(p, Gia_ObjFaninId1(pObj, i)) {
                fmt.Printf("Fanin 1 of AND node %d has a repr.\n", i);
fProb = 1; }
        
        } else if Gia_ObjIsCo(pObj) {
            if Gia_ObjHasRepr(p, Gia_ObjFaninId0(pObj, i)) {
                fmt.Printf("Fanin 0 of CO node %d has a repr.\n", i);
fProb = 1; }
        }
    }
//    if ( !fProb )
//        printf( "GIA with choices is correct.\n" );
}

/**Function*************************************************************

  Synopsis    [Reverse the order of nodes in equiv classes.]

  Description [If the flag is 1, assumed current increasing order ]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManReverseClasses(p *Gia_Man_t, fNowIncreasing int32) {
    var vCollected *Vec_Int_t;
    var vClass *Vec_Int_t;
    var i, k, iRepr, iNode, iPrev int32;
    // collect classes
    vCollected = Vec_IntAlloc( 100 );
    // inlining of macro Gia_ManForEachClass
    for  iRepr = 1; iRepr < Gia_ManObjNum(p); iRepr++  { if !Gia_ObjIsHead(p, iRepr) {} else {
        Vec_IntPush( vCollected, iRepr ); }
    }
    // correct each class
    vClass = Vec_IntAlloc( 100 );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vCollected)) && (((iRepr) = Vec_IntEntry(vCollected, i)), 1); i++  {
        Vec_IntClear( vClass );
        Vec_IntPush( vClass, iRepr );
        // inlining of macro Gia_ClassForEachObj1
        for  assert(Gia_ObjIsHead(p, iRepr), "Gia_ObjIsHead(p, iRepr)"), iNode = Gia_ObjNext(p, iRepr); iNode > 0; iNode = Gia_ObjNext(p, iNode)  {
            if fNowIncreasing != 0 {
                assert(iRepr < iNode, "iRepr < iNode");
            } else {
                assert(iRepr > iNode, "iRepr > iNode"); }
            Vec_IntPush( vClass, iNode );
        }
//        if ( !fNowIncreasing )
//            Vec_IntSort( vClass, 1 );
        // reverse the class
        iPrev = 0;
        iRepr = Vec_IntEntryLast( vClass );
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vClass)) && (((iNode) = Vec_IntEntry(vClass, k)), 1); k++  {
            if fNowIncreasing != 0 {
                Gia_ObjSetReprRev( p, iNode, iNode == iRepr ? GIA_VOID : iRepr );
            } else {
                Gia_ObjSetRepr( p, iNode, iNode == iRepr ? GIA_VOID : iRepr ); }
            Gia_ObjSetNext( p, iNode, iPrev );
            iPrev = iNode;
        }
    }
    Vec_IntFree( vCollected );
    Vec_IntFree( vClass );
    // verify
    // inlining of macro Gia_ManForEachClass
    for  iRepr = 1; iRepr < Gia_ManObjNum(p); iRepr++  { if !Gia_ObjIsHead(p, iRepr) {} else {
        // inlining of macro Gia_ClassForEachObj1
        for  assert(Gia_ObjIsHead(p, iRepr), "Gia_ObjIsHead(p, iRepr)"), iNode = Gia_ObjNext(p, iRepr); iNode > 0; iNode = Gia_ObjNext(p, iNode)  {
            if fNowIncreasing != 0 {
                assert(Gia_ObjRepr(p, iNode) == iRepr && iRepr > iNode, "Gia_ObjRepr(p, iNode) == iRepr && iRepr > iNode");
            } else {
                assert(Gia_ObjRepr(p, iNode) == iRepr && iRepr < iNode, "Gia_ObjRepr(p, iNode) == iRepr && iRepr < iNode"); }
        } }
    }
}

/**Function*************************************************************

  Synopsis    [Tests the hierarchy-timing manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkTestTimByWritingFile(pGia *Gia_Man_t, pFileName string) {
    var pGia2 *Gia_Man_t;

    // normalize choices
    if Gia_ManHasChoices(pGia) {
        Gia_ManVerifyChoices( pGia );
        Gia_ManReverseClasses( pGia, 0 );
    }
    // write file
    Gia_AigerWrite( pGia, pFileName, 0, 0, 0 );
    // unnormalize choices
    if Gia_ManHasChoices(pGia) {
        Gia_ManReverseClasses( pGia, 1 ); }

    // read file
    pGia2 = Gia_AigerRead( pFileName, 0, 1, 1 );

    // normalize choices
    if Gia_ManHasChoices(pGia2) {
        Gia_ManVerifyChoices( pGia2 );
        Gia_ManReverseClasses( pGia2, 1 );
    }

    // compare resulting managers
    if Gia_ManCompare( pGia, pGia2 ) {
        fmt.Printf("Verification suceessful.\n"); }

    Gia_ManStop( pGia2 );
}

/**Function*************************************************************

  Synopsis    [Tests construction and serialization.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkTestTim(pNtk *Abc_Ntk_t, fVerbose int32) {
    var fUseChoices int32;
    var pGia, pTemp *Gia_Man_t;

    // this test only works for a logic network (for example, network with LUT mapping)
    assert(Abc_NtkIsLogic(pNtk), "Abc_NtkIsLogic(pNtk)");
    // make sure local functions of the nodes are in the AIG form
    Abc_NtkToAig( pNtk );

    // create GIA manager (pGia) with hierarhy/timing manager attached (pGia->pManTime)
    // while assuming that some nodes are white boxes (see Abc_NodeIsWhiteBox)
    pGia = Abc_NtkTestTimDeriveGia( pNtk, fVerbose );
    fmt.Printf("Created GIA manager for network with %d white boxes.\n", Tim_ManBoxNum((Tim_Man_t *)pGia.pManTime));

    // print the timing manager
    if fVerbose != 0 {
        Tim_ManPrint( (Tim_Man_t *)pGia.pManTime ); }

    // test writing both managers into a file and reading them back
    Abc_NtkTestTimByWritingFile( pGia, "test1.aig" );

    // perform synthesis
    pGia = Abc_NtkTestTimPerformSynthesis( pTemp = pGia, fUseChoices );
    Gia_ManStop( pTemp );

    // test writing both managers into a file and reading them back
    Abc_NtkTestTimByWritingFile( pGia, "test2.aig" );

    Gia_ManStop( pGia );
}




////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

