package abc

import "fmt"

/**CFile****************************************************************

  FileName    [abcHieCec.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Network and node package.]

  Synopsis    [Hierarchical CEC manager.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: abcHieCec.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "abc.h"
//#include "base/io/ioAbc.h"
//#include "aig/gia/gia.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// Original macro:
// #define Abc_ObjForEachFaninReal( pObj, pFanin, i )          \
//     for ( i = 0; (i < Abc_ObjFaninNum(pObj)) && (((pFanin) = Abc_ObjFaninReal(pObj, i)), 1); i++ )


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Returns the real faniin of the object.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_ObjFaninReal(pObj *Abc_Obj_t, i int32) *Abc_Obj_t {
    var pRes *Abc_Obj_t;
    if Abc_ObjIsBox(pObj) {
        pRes = Abc_ObjFanin0( Abc_ObjFanin0( Abc_ObjFanin(pObj, i) ) );
    } else {
        assert(Abc_ObjIsPo(pObj) || Abc_ObjIsNode(pObj), "Abc_ObjIsPo(pObj) || Abc_ObjIsNode(pObj)");
        pRes = Abc_ObjFanin0( Abc_ObjFanin(pObj, i) );
    }
    if Abc_ObjIsBo(pRes) {
        return Abc_ObjFanin0(pRes); }
    return pRes;
}

/**Function*************************************************************

  Synopsis    [Performs DFS for one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDfsBoxes_rec(pNode *Abc_Obj_t, vNodes *Vec_Ptr_t) {
    var pFanin *Abc_Obj_t;
    var i int32;
    if Abc_ObjIsPi(pNode) {
        return; }
    assert(Abc_ObjIsNode(pNode) || Abc_ObjIsBox(pNode), "Abc_ObjIsNode(pNode) || Abc_ObjIsBox(pNode)");
    // if this node is already visited, skip
    if Abc_NodeIsTravIdCurrent( pNode ) {
        return; }
    Abc_NodeSetTravIdCurrent( pNode );
    // visit the transitive fanin of the node
    // inlining of macro Abc_ObjForEachFaninReal
    for  i = 0; (i < Abc_ObjFaninNum(pNode)) && (((pFanin) = Abc_ObjFaninReal(pNode, i)), 1); i++  {
        Abc_NtkDfsBoxes_rec( pFanin, vNodes );
    }
    // add the node after the fanins have been added
    Vec_PtrPush( vNodes, pNode );
}

/**Function*************************************************************

  Synopsis    [Returns the array of node and boxes reachable from POs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDfsBoxes(pNtk *Abc_Ntk_t) *Vec_Ptr_t {
    var vNodes *Vec_Ptr_t;
    var pObj *Abc_Obj_t;
    var i int32;
    assert(Abc_NtkIsNetlist(pNtk), "Abc_NtkIsNetlist(pNtk)");
    // set the traversal ID
    Abc_NtkIncrementTravId( pNtk );
    // start the array of nodes
    vNodes = Vec_PtrAlloc( 100 );
    // inlining of macro Abc_NtkForEachPo
    for  i = 0; (i < Abc_NtkPoNum(pNtk)) && (((pObj) = Abc_NtkPo(pNtk, i)), 1); i++  {
        Abc_NtkDfsBoxes_rec( Abc_ObjFaninReal(pObj, 0), vNodes );
    }
    return vNodes;
}


/**Function*************************************************************

  Synopsis    [Strashes one logic node using its SOP.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDeriveFlatGiaSop(pGia *Gia_Man_t, gFanins *int32, pSop string) int32 {
    var pCube string;
    var gAnd, gSum int32;
    var i, Value, nFanins int32;
    // get the number of variables
    nFanins = Abc_SopGetVarNum(pSop);
    if Abc_SopIsExorType(pSop) {
        gSum = 0; 
        for  i = 0; i < nFanins; i++  {
            gSum = Gia_ManHashXor( pGia, gSum, gFanins[i] );
        }
    
    } else {
        // go through the cubes of the node's SOP
        gSum = 0; 
        // inlining of macro Abc_SopForEachCube
        for  pCube = (pSop); *pCube; pCube += (nFanins) + 3  {
            // create the AND of literals
            gAnd = 1;
            // inlining of macro Abc_CubeForEachVar
            for  i = 0; (pCube[i] != ' ') && (Value = pCube[i]); i++  {
                if Value == '1' {
                    gAnd = Gia_ManHashAnd( pGia, gAnd, gFanins[i] );
                } else if Value == '0' {
                    gAnd = Gia_ManHashAnd( pGia, gAnd, Abc_LitNot(gFanins[i]) ); }
            }
            // add to the sum of cubes
            gSum = Gia_ManHashAnd( pGia, Abc_LitNot(gSum), Abc_LitNot(gAnd) );
            gSum = Abc_LitNot( gSum );
        }
    }
    // decide whether to complement the result
    if Abc_SopIsComplement(pSop) {
        gSum = Abc_LitNot(gSum); }
    return gSum;
}

/**Function*************************************************************

  Synopsis    [Flattens the logic hierarchy of the netlist.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDeriveFlatGia_rec(pGia *Gia_Man_t, pNtk *Abc_Ntk_t) {
    var gFanins [16]int32;
    var vOrder *Vec_Ptr_t  = (Vec_Ptr_t *)pNtk.pData;
    var pObj, pTerm *Abc_Obj_t;
    var pNtkModel *Abc_Ntk_t;
    var i, k int32;
    // inlining of macro Abc_NtkForEachPi
    for  i = 0; (i < Abc_NtkPiNum(pNtk)) && (((pTerm) = Abc_NtkPi(pNtk, i)), 1); i++  {
        assert(Abc_ObjFanout0(pTerm).iTemp >= 0, "Abc_ObjFanout0(pTerm).iTemp >= 0");
    }
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vOrder)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vOrder, i)), 1); i++  {
        if Abc_ObjIsNode(pObj) {
            var pSop string  = (char *)pObj.pData;
/*
            int nLength = strlen(pSop);
            if ( nLength == 4 ) // buf/inv
            {
                assert( pSop[2] == '1' );
                assert( pSop[0] == '0' || pSop[0] == '1' );
                assert( Abc_ObjFanin0(pObj)->iTemp >= 0 );
                Abc_ObjFanout0(pObj)->iTemp = Abc_LitNotCond( Abc_ObjFanin0(pObj)->iTemp, pSop[0]=='0' );
                continue;
            }
            if ( nLength == 5 ) // and2
            {
                assert( pSop[3] == '1' );
                assert( pSop[0] == '0' || pSop[0] == '1' );
                assert( pSop[1] == '0' || pSop[1] == '1' );
                assert( Abc_ObjFanin0(pObj)->iTemp >= 0 );
                assert( Abc_ObjFanin1(pObj)->iTemp >= 0 );
                Abc_ObjFanout0(pObj)->iTemp = Gia_ManHashAnd( pGia, 
                    Abc_LitNotCond( Abc_ObjFanin0(pObj)->iTemp, pSop[0]=='0' ),
                    Abc_LitNotCond( Abc_ObjFanin1(pObj)->iTemp, pSop[1]=='0' )
                    );
                continue;
            }
*/
            assert(Abc_ObjFaninNum(pObj) <= 16, "Abc_ObjFaninNum(pObj) <= 16");
            assert(Abc_ObjFaninNum(pObj) == Abc_SopGetVarNum(pSop), "Abc_ObjFaninNum(pObj) == Abc_SopGetVarNum(pSop)");
            // inlining of macro Abc_ObjForEachFanin
            for  k = 0; (k < Abc_ObjFaninNum(pObj)) && (((pTerm) = Abc_ObjFanin(pObj, k)), 1); k++  {
                gFanins[k] = pTerm.iTemp;
                assert(gFanins[k] >= 0, "gFanins[k] >= 0");
            }
            Abc_ObjFanout0(pObj).iTemp = Abc_NtkDeriveFlatGiaSop( pGia, gFanins, pSop );
            continue;
        }
        assert(Abc_ObjIsBox(pObj), "Abc_ObjIsBox(pObj)");
        pNtkModel = (Abc_Ntk_t *)pObj.pData;
        Abc_NtkFillTemp( pNtkModel );
        // check the match between the number of actual and formal parameters
        assert(Abc_ObjFaninNum(pObj) == Abc_NtkPiNum(pNtkModel), "Abc_ObjFaninNum(pObj) == Abc_NtkPiNum(pNtkModel)");
        assert(Abc_ObjFanoutNum(pObj) == Abc_NtkPoNum(pNtkModel), "Abc_ObjFanoutNum(pObj) == Abc_NtkPoNum(pNtkModel)");
        // assign PIs
        // inlining of macro Abc_ObjForEachFanin
        for  k = 0; (k < Abc_ObjFaninNum(pObj)) && (((pTerm) = Abc_ObjFanin(pObj, k)), 1); k++  {
            Abc_ObjFanout0( Abc_NtkPi(pNtkModel, k) ).iTemp = Abc_ObjFanin0(pTerm).iTemp;
        }
        // call recursively
        Abc_NtkDeriveFlatGia_rec( pGia, pNtkModel );
        // assign POs
        // inlining of macro Abc_ObjForEachFanout
        for  k = 0; (k < Abc_ObjFanoutNum(pObj)) && (((pTerm) = Abc_ObjFanout(pObj, k)), 1); k++  {
            Abc_ObjFanout0(pTerm).iTemp = Abc_ObjFanin0( Abc_NtkPo(pNtkModel, k) ).iTemp;
        }
    }
    // inlining of macro Abc_NtkForEachPo
    for  i = 0; (i < Abc_NtkPoNum(pNtk)) && (((pTerm) = Abc_NtkPo(pNtk, i)), 1); i++  {
        assert(Abc_ObjFanin0(pTerm).iTemp >= 0, "Abc_ObjFanin0(pTerm).iTemp >= 0");
    }
}

/**Function*************************************************************

  Synopsis    [Flattens the logic hierarchy of the netlist.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDeriveFlatGia(pNtk *Abc_Ntk_t) *Gia_Man_t {
    var pTemp *Gia_Man_t;
    var pGia *Gia_Man_t = nil;
    var pTerm *Abc_Obj_t;
    var i int32;

    assert(Abc_NtkIsNetlist(pNtk), "Abc_NtkIsNetlist(pNtk)");
    assert(!Abc_NtkLatchNum(pNtk), "!Abc_NtkLatchNum(pNtk)");
    Abc_NtkFillTemp( pNtk );
    // start the network
    pGia = Gia_ManStart( (1<<16) );
    pGia.pName = Abc_UtilStrsav( Abc_NtkName(pNtk) );
    pGia.pSpec = Abc_UtilStrsav( Abc_NtkSpec(pNtk) );
    Gia_ManHashAlloc( pGia );
    // create PIs
    // inlining of macro Abc_NtkForEachPi
    for  i = 0; (i < Abc_NtkPiNum(pNtk)) && (((pTerm) = Abc_NtkPi(pNtk, i)), 1); i++  {
        Abc_ObjFanout0(pTerm).iTemp = Gia_ManAppendCi( pGia );
    }
    // recursively flatten hierarchy
    Abc_NtkDeriveFlatGia_rec( pGia, pNtk );
    // create POs
    // inlining of macro Abc_NtkForEachPo
    for  i = 0; (i < Abc_NtkPoNum(pNtk)) && (((pTerm) = Abc_NtkPo(pNtk, i)), 1); i++  {
        Gia_ManAppendCo( pGia, Abc_ObjFanin0(pTerm).iTemp );
    }
    // prepare return value
    Gia_ManHashStop( pGia );
    Gia_ManSetRegNum( pGia, 0 );
    pGia = Gia_ManCleanup( pTemp = pGia );
    Gia_ManStop( pTemp );
    return pGia;
}

/**Function*************************************************************

  Synopsis    [Counts the total number of AIG nodes before flattening.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkCountAndNodes(vOrder *Vec_Ptr_t) int32 {
    var pGiaBox *Gia_Man_t;
    var pNtkModel *Abc_Ntk_t;
    var pObj *Abc_Obj_t;
    var i, Counter int32;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vOrder)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vOrder, i)), 1); i++  {
        if Abc_ObjIsNode(pObj) {
            Counter++;
            continue;
        }
        assert(Abc_ObjIsBox(pObj), "Abc_ObjIsBox(pObj)");
        pNtkModel = (Abc_Ntk_t *)pObj.pData;
        pGiaBox   = (Gia_Man_t *)pNtkModel.pData;
        Counter  += Gia_ManAndNum(pGiaBox);
    }
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Flattens the logic hierarchy of the netlist.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDeriveFlatGia2Derive(pNtk *Abc_Ntk_t, vOrder *Vec_Ptr_t) *Gia_Man_t {
    var gFanins [16]int32;
    var pNtkModel *Abc_Ntk_t;
    var pGiaBox *Gia_Man_t;
    var pGia *Gia_Man_t = nil;
    var pGiaObj *Gia_Obj_t;
    var pTerm, pObj *Abc_Obj_t;
    var i, k int32;

    assert(Abc_NtkIsNetlist(pNtk), "Abc_NtkIsNetlist(pNtk)");
    assert(!Abc_NtkLatchNum(pNtk), "!Abc_NtkLatchNum(pNtk)");
    Abc_NtkFillTemp( pNtk );

    // start the network
    pGia = Gia_ManStart( (1<<15) );
    pGia.pName = Abc_UtilStrsav( Abc_NtkName(pNtk) );
    pGia.pSpec = Abc_UtilStrsav( Abc_NtkSpec(pNtk) );
    Gia_ManHashAlloc( pGia );
    // create PIs
    // inlining of macro Abc_NtkForEachPi
    for  i = 0; (i < Abc_NtkPiNum(pNtk)) && (((pTerm) = Abc_NtkPi(pNtk, i)), 1); i++  {
        Abc_ObjFanout0(pTerm).iTemp = Gia_ManAppendCi( pGia );
    }
    // recursively flatten hierarchy
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vOrder)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vOrder, i)), 1); i++  {
        if Abc_ObjIsNode(pObj) {
            var pSop string  = (char *)pObj.pData;
            assert(Abc_ObjFaninNum(pObj) <= 16, "Abc_ObjFaninNum(pObj) <= 16");
            assert(Abc_ObjFaninNum(pObj) == Abc_SopGetVarNum(pSop), "Abc_ObjFaninNum(pObj) == Abc_SopGetVarNum(pSop)");
            // inlining of macro Abc_ObjForEachFanin
            for  k = 0; (k < Abc_ObjFaninNum(pObj)) && (((pTerm) = Abc_ObjFanin(pObj, k)), 1); k++  {
                gFanins[k] = pTerm.iTemp;
                assert(gFanins[k] >= 0, "gFanins[k] >= 0");
            }
            Abc_ObjFanout0(pObj).iTemp = Abc_NtkDeriveFlatGiaSop( pGia, gFanins, pSop );
            continue;
        }
        assert(Abc_ObjIsBox(pObj), "Abc_ObjIsBox(pObj)");
        pNtkModel = (Abc_Ntk_t *)pObj.pData;
        // check the match between the number of actual and formal parameters
        assert(Abc_ObjFaninNum(pObj) == Abc_NtkPiNum(pNtkModel), "Abc_ObjFaninNum(pObj) == Abc_NtkPiNum(pNtkModel)");
        assert(Abc_ObjFanoutNum(pObj) == Abc_NtkPoNum(pNtkModel), "Abc_ObjFanoutNum(pObj) == Abc_NtkPoNum(pNtkModel)");
/*
        // assign PIs
        Abc_ObjForEachFanin( pObj, pTerm, k )
            Abc_ObjFanout0( Abc_NtkPi(pNtkModel, k) )->iTemp = Abc_ObjFanin0(pTerm)->iTemp;
        // call recursively
        Abc_NtkDeriveFlatGia_rec( pGia, pNtkModel );
        // assign POs
        Abc_ObjForEachFanout( pObj, pTerm, k )
            Abc_ObjFanout0(pTerm)->iTemp = Abc_ObjFanin0( Abc_NtkPo(pNtkModel, k) )->iTemp;
*/ 
        // duplicate the AIG
        pGiaBox = (Gia_Man_t *)pNtkModel.pData;
        assert(Abc_ObjFaninNum(pObj) == Gia_ManPiNum(pGiaBox), "Abc_ObjFaninNum(pObj) == Gia_ManPiNum(pGiaBox)");
        assert(Abc_ObjFanoutNum(pObj) == Gia_ManPoNum(pGiaBox), "Abc_ObjFanoutNum(pObj) == Gia_ManPoNum(pGiaBox)");
        Gia_ManFillValue( pGiaBox );
        Gia_ManConst0(pGiaBox).Value = 0;
        // inlining of macro Abc_ObjForEachFanin
        for  k = 0; (k < Abc_ObjFaninNum(pObj)) && (((pTerm) = Abc_ObjFanin(pObj, k)), 1); k++  {
            Gia_ManPi(pGiaBox, k).Value = Abc_ObjFanin0(pTerm).iTemp;
        }
        // inlining of macro Gia_ManForEachAnd
        for  k = 0; (k < pGiaBox.nObjs) && ((pGiaObj) = Gia_ManObj(pGiaBox, k)); k++  {      if !Gia_ObjIsAnd(pGiaObj) {} else {
            pGiaObj.Value = Gia_ManHashAnd( pGia, Gia_ObjFanin0Copy(pGiaObj), Gia_ObjFanin1Copy(pGiaObj) ); }
        }
        // inlining of macro Abc_ObjForEachFanout
        for  k = 0; (k < Abc_ObjFanoutNum(pObj)) && (((pTerm) = Abc_ObjFanout(pObj, k)), 1); k++  {
            Abc_ObjFanout0(pTerm).iTemp = Gia_ObjFanin0Copy(Gia_ManPo(pGiaBox, k));
        }
    }
    // create POs
    // inlining of macro Abc_NtkForEachPo
    for  i = 0; (i < Abc_NtkPoNum(pNtk)) && (((pTerm) = Abc_NtkPo(pNtk, i)), 1); i++  {
        Gia_ManAppendCo( pGia, Abc_ObjFanin0(pTerm).iTemp );
    }
    // prepare return value
    Gia_ManHashStop( pGia );
    Gia_ManSetRegNum( pGia, 0 );
    pGia = Gia_ManCleanup( pGiaBox = pGia );
    Gia_ManStop( pGiaBox );

    fmt.Printf("%8d -> ", Abc_NtkCountAndNodes(vOrder));
    Gia_ManPrintStats( pGia, nil );
    return pGia;
}
/*
void Abc_NtkDeriveFlatGia2_rec( Abc_Ntk_t * pNtk )
{
    Vec_Ptr_t * vOrder;
    Abc_Obj_t * pObj;
    int i;
    if ( pNtk->pData != NULL )
        return;
    vOrder = Abc_NtkDfsBoxes( pNtk );
    Vec_PtrForEachEntry( Abc_Obj_t *, vOrder, pObj, i )
        if ( Abc_ObjIsBox(pObj) )
            Abc_NtkDeriveFlatGia2_rec( (Abc_Ntk_t *)pObj->pData );
    pNtk->pData = Abc_NtkDeriveFlatGia2Derive( pNtk, vOrder );
    Vec_PtrFree( vOrder );
}

Gia_Man_t * Abc_NtkDeriveFlatGia2( Abc_Ntk_t * pNtk )
{
    Vec_Ptr_t * vMods;
    Abc_Ntk_t * pModel;
    Gia_Man_t * pGia = NULL;
    int i;

    assert( Abc_NtkIsNetlist(pNtk) );
    assert( !Abc_NtkLatchNum(pNtk) );

    vMods = pNtk->pDesign->vModules;
    Vec_PtrForEachEntry( Abc_Ntk_t *, vMods, pModel, i )
        pModel->pData = NULL;

    Abc_NtkDeriveFlatGia2_rec( pNtk );
    pGia = pNtk->pData;  pNtk->pData = NULL;
    
    Vec_PtrForEachEntry( Abc_Ntk_t *, vMods, pModel, i )
        Gia_ManStopP( (Gia_Man_t **)&pModel->pData );

    return pGia;
}
*/
func Abc_NtkDeriveFlatGia2(pNtk *Abc_Ntk_t, vModels *Vec_Ptr_t) *Gia_Man_t { 
    var vOrder *Vec_Ptr_t;
    var pModel *Abc_Ntk_t  = nil;
    var pGia *Gia_Man_t  = nil;
    var i int32;

    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vModels)) && (((pModel) = (Abc_Ntk_t *)Vec_PtrEntry(vModels, i)), 1); i++  {
        vOrder = Abc_NtkDfsBoxes( pModel );
        pModel.pData = Abc_NtkDeriveFlatGia2Derive( pModel, vOrder );
        Vec_PtrFree( vOrder );
    }

    pGia = (Gia_Man_t *)pModel.pData;  pModel.pData = nil;

    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vModels)) && (((pModel) = (Abc_Ntk_t *)Vec_PtrEntry(vModels, i)), 1); i++  {
        Gia_ManStopP( (Gia_Man_t **)&pModel.pData );
    }

    return pGia;
}


/**Function*************************************************************

  Synopsis    [Collect models in the DFS order.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkCollectHie_rec(pNtk *Abc_Ntk_t, vModels *Vec_Ptr_t) {
    var vOrder *Vec_Ptr_t;
    var pObj *Abc_Obj_t;
    var i int32;
    if pNtk.iStep >= 0 {
        return; }
    vOrder = Abc_NtkDfsBoxes( pNtk );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vOrder)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vOrder, i)), 1); i++  {
        if Abc_ObjIsBox(pObj) && (Abc_Ntk_t *)pObj.pData != pNtk {
            Abc_NtkCollectHie_rec( (Abc_Ntk_t *)pObj.pData, vModels ); }
    }
    Vec_PtrFree( vOrder );
    pNtk.iStep = Vec_PtrSize(vModels);
    Vec_PtrPush( vModels, pNtk );
}

func Abc_NtkCollectHie(pNtk *Abc_Ntk_t) *Vec_Ptr_t {
    var vMods, vResult *Vec_Ptr_t;
    var pModel *Abc_Ntk_t;
    var i int32;

    assert(Abc_NtkIsNetlist(pNtk), "Abc_NtkIsNetlist(pNtk)");
    assert(!Abc_NtkLatchNum(pNtk), "!Abc_NtkLatchNum(pNtk)");

    vResult = Vec_PtrAlloc( 1000 );
    if pNtk.pDesign == nil {
        Vec_PtrPush( vResult, pNtk );
        return vResult;
    }

    vMods = pNtk.pDesign.vModules;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vMods)) && (((pModel) = (Abc_Ntk_t *)Vec_PtrEntry(vMods, i)), 1); i++  {
        pModel.iStep = -1;
    }

    Abc_NtkCollectHie_rec( pNtk, vResult );
    return vResult;    
}

/**Function*************************************************************

  Synopsis    [Counts the number of intstances.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkCountInst_rec(pNtk *Abc_Ntk_t) int32 {
    var vOrder *Vec_Ptr_t;
    var pObj *Abc_Obj_t;
    var i, Counter int32;
    if pNtk.iStep >= 0 {
        return pNtk.iStep; }
    vOrder = Abc_NtkDfsBoxes( pNtk );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vOrder)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vOrder, i)), 1); i++  {
        if Abc_ObjIsBox(pObj) && (Abc_Ntk_t *)pObj.pData != pNtk {
            Counter += Abc_NtkCountInst_rec( (Abc_Ntk_t *)pObj.pData ); }
    }
    Vec_PtrFree( vOrder );
    return pNtk.iStep = 1 + Counter;
}

func Abc_NtkCountInst(pNtk *Abc_Ntk_t) {
    var vMods *Vec_Ptr_t;
    var pModel *Abc_Ntk_t;
    var i, Counter int32;

    if pNtk.pDesign == nil {
        Counter = Abc_NtkNodeNum(pNtk);
    } else {
        vMods = pNtk.pDesign.vModules;
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(vMods)) && (((pModel) = (Abc_Ntk_t *)Vec_PtrEntry(vMods, i)), 1); i++  {
            pModel.iStep = -1;
        }
        Counter = Abc_NtkCountInst_rec( pNtk );
    }
    fmt.Printf("Instances = %10d.\n", Counter);
}

/**Function*************************************************************

  Synopsis    [Counts the number of nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkCountNodes_rec(pNtk *Abc_Ntk_t) float64 {
    var vOrder *Vec_Ptr_t;
    var pObj *Abc_Obj_t;
    var Counter float64;
    var i int32;
    if pNtk.dTemp >= 0 {
        return pNtk.dTemp; }
    vOrder = Abc_NtkDfsBoxes( pNtk );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vOrder)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vOrder, i)), 1); i++  {
        if Abc_ObjIsNode(pObj) {
            Counter++;
        } else if Abc_ObjIsBox(pObj) && (Abc_Ntk_t *)pObj.pData != pNtk {
            Counter += Abc_NtkCountNodes_rec( (Abc_Ntk_t *)pObj.pData ); }
    }
    Vec_PtrFree( vOrder );
    return pNtk.dTemp = Counter;
}

func Abc_NtkCountNodes(pNtk *Abc_Ntk_t) {
    var vMods *Vec_Ptr_t;
    var pModel *Abc_Ntk_t;
    var Counter float64;
    var i int32;

    if pNtk.pDesign == nil {
        Counter = Abc_NtkNodeNum(pNtk);
    } else {
        vMods = pNtk.pDesign.vModules;
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(vMods)) && (((pModel) = (Abc_Ntk_t *)Vec_PtrEntry(vMods, i)), 1); i++  {
            pModel.dTemp = -1;
        }
        Counter = Abc_NtkCountNodes_rec( pNtk );
    }
    fmt.Printf("Nodes = %.0f\n", Counter);
}

/**Function*************************************************************

  Synopsis    [Checks if there is a recursive definition.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkCheckRecursive(pNtk *Abc_Ntk_t) int32 {
    var vMods *Vec_Ptr_t;
    var pModel *Abc_Ntk_t;
    var pObj *Abc_Obj_t;
    var i, k, RetValue int32;

    assert(Abc_NtkIsNetlist(pNtk), "Abc_NtkIsNetlist(pNtk)");
    assert(!Abc_NtkLatchNum(pNtk), "!Abc_NtkLatchNum(pNtk)");

    if pNtk.pDesign == nil {
        return RetValue; }

    vMods = pNtk.pDesign.vModules;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vMods)) && (((pModel) = (Abc_Ntk_t *)Vec_PtrEntry(vMods, i)), 1); i++  {
        // inlining of macro Abc_NtkForEachObj
        for  k = 0; (k < Vec_PtrSize((pModel).vObjs)) && (((pObj) = Abc_NtkObj(pModel, k)), 1); k++  {            if (pObj) == nil {} else
            if Abc_ObjIsBox(pObj) && pObj.pData == (void *)pModel {
                fmt.Printf("WARNING: Model \"%s\" contains a recursive definition.\n", Abc_NtkName(pModel));
                RetValue = 1;
                break;
            }
        }
    }
    return RetValue;
}

/**Function*************************************************************

  Synopsis    [Performs hierarchical equivalence checking.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkHieCecTest(pFileName string, fVerbose int32) *Gia_Man_t {
    var fUseTest int32  = 1;
    var fUseNew int32;
    var fCheck int32  = 1;
    var vMods, vOrder *Vec_Ptr_t;
    var pNtk, pModel *Abc_Ntk_t;
    var pGia *Gia_Man_t;
    var i int32;
    var clk abctime  = Abc_Clock();

    // read hierarchical netlist
    pNtk = Io_ReadBlifMv( pFileName, 0, fCheck );
    if pNtk == nil {
        fmt.Printf("Reading BLIF file has failed.\n");
        return nil;
    }
    if pNtk.pDesign == nil || pNtk.pDesign.vModules == nil {
        fmt.Printf("There is no hierarchy information.\n");
//        Abc_NtkDelete( pNtk );
//        return NULL;
    }
    Abc_PrintTime( 1, "Reading file", Abc_Clock() - clk );

    assert(Abc_NtkIsNetlist(pNtk), "Abc_NtkIsNetlist(pNtk)");
    assert(!Abc_NtkLatchNum(pNtk), "!Abc_NtkLatchNum(pNtk)");
/*
    if ( pNtk->pDesign != NULL )
    {
        clk = Abc_Clock();
        Abc_NtkCountNodes( pNtk );
        Abc_PrintTime( 1, "Count nodes", Abc_Clock() - clk );
    }
*/
    // print stats
    if fVerbose != 0 {
        Abc_NtkPrintBoxInfo( pNtk ); }

    // test the new data-structure
    if fUseTest != 0 {
        extern Gia_Man_t * Au_ManDeriveTest( Abc_Ntk_t * pRoot );
        var pGia *Gia_Man_t;
        pGia = Au_ManDeriveTest( pNtk );
        Abc_NtkDelete( pNtk );
        return pGia;
    }

    if Abc_NtkCheckRecursive(pNtk) != 0 {
        return nil; }

    if fUseNew != 0 {
        clk = Abc_Clock();
        vOrder = Abc_NtkCollectHie( pNtk );
        Abc_PrintTime( 1, "Collect DFS ", Abc_Clock() - clk );

        // derive GIA
        clk = Abc_Clock();
        pGia = Abc_NtkDeriveFlatGia2( pNtk, vOrder );
        Abc_PrintTime( 1, "Deriving GIA", Abc_Clock() - clk );
        Gia_ManPrintStats( pGia, nil );
    //    Gia_ManStop( pGia );
 
        Vec_PtrFree( vOrder );
    
    } else {
        // order nodes/boxes of all models
        vMods = pNtk.pDesign.vModules;
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(vMods)) && (((pModel) = (Abc_Ntk_t *)Vec_PtrEntry(vMods, i)), 1); i++  {
            pModel.pData = Abc_NtkDfsBoxes( pModel );
        }

        // derive GIA
        clk = Abc_Clock();
        pGia = Abc_NtkDeriveFlatGia( pNtk );
        Abc_PrintTime( 1, "Deriving GIA", Abc_Clock() - clk );
        Gia_ManPrintStats( pGia, nil );

        // clean nodes/boxes of all nodes
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(vMods)) && (((pModel) = (Abc_Ntk_t *)Vec_PtrEntry(vMods, i)), 1); i++  {
            Vec_PtrFree( (Vec_Ptr_t *)pModel.pData );
        }
    }

    clk = Abc_Clock();
    Abc_NtkCountInst( pNtk );
    Abc_PrintTime( 1, "Gather stats", Abc_Clock() - clk );

    Abc_NtkDelete( pNtk );
    return pGia;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

