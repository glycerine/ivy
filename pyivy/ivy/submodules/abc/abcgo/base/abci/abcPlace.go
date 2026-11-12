package abci

/**CFile****************************************************************

  FileName    [abcPlace.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Network and node package.]

  Synopsis    [Interface with a placer.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: abcPlace.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "base/abc/abc.h"

// placement includes
//#include "phys/place/place_base.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

var abstractCells *AbstractCell = nil;
var cells *ConcreteCell = nil;
var nets *ConcreteNet = nil;
var nAllocSize int32;

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Creates a new cell.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_PlaceCreateCell(pObj *Abc_Obj_t, fAnd int32) {
    assert(cells[pObj.Id].m_id == 0, "cells[pObj.Id].m_id == 0");

    cells[pObj.Id].m_id = pObj.Id;
    cells[pObj.Id].m_label = "";
    cells[pObj.Id].m_parent = &(abstractCells[fAnd]);
    cells[pObj.Id].m_fixed = 0;
    addConcreteCell(&(cells[pObj.Id]));
}

/**Function*************************************************************

  Synopsis    [Updates the net.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_PlaceUpdateNet(pObj *Abc_Obj_t) {
    var pFanout *Abc_Obj_t;
    var k int32;
    // free the old array of net terminals
    if nets[pObj.Id].m_terms {
        free( nets[pObj.Id].m_terms ); }
    // fill in the net with the new information
    nets[pObj.Id].m_id = pObj.Id;
    nets[pObj.Id].m_weight = 1.0;
    nets[pObj.Id].m_numTerms = Abc_ObjFanoutNum(pObj); //fanout
    nets[pObj.Id].m_terms = ALLOC(ConcreteCell*, Abc_ObjFanoutNum(pObj));
    // inlining of macro Abc_ObjForEachFanout
    for  k = 0; (k < Abc_ObjFanoutNum(pObj)) && (((pFanout) = Abc_ObjFanout(pObj, k)), 1); k++  {
        nets[pObj.Id].m_terms[k] = &(cells[pFanout.Id]);
    }
    addConcreteNet(&(nets[pObj.Id]));
}

/**Function*************************************************************

  Synopsis    [Returns the placement cost of the cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_PlaceEvaluateCut(pRoot *Abc_Obj_t, vFanins *Vec_Ptr_t) float32 {
    var pObj *Abc_Obj_t;
//    double x, y;
    var i int32;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vFanins)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vFanins, i)), 1); i++  {
//        pObj->Id
    }
    return 0.0;
}

/**Function*************************************************************

  Synopsis    [Updates placement after one step of rewriting.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_PlaceUpdate(vAddedCells *Vec_Ptr_t, vUpdatedNets *Vec_Ptr_t) {
    var pObj, pFanin *Abc_Obj_t;
    var i, k int32;
    var vCells, vNets *Vec_Ptr_t;

    // start the arrays of new cells and nets
    vCells = Vec_PtrAlloc( 16 );
    vNets = Vec_PtrAlloc( 32 );

    // go through the new nodes
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vAddedCells)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vAddedCells, i)), 1); i++  {
        assert(!Abc_ObjIsComplement(pObj), "!Abc_ObjIsComplement(pObj)");
        Abc_PlaceCreateCell( pObj, 1 );
        Abc_PlaceUpdateNet( pObj );

        // add the new cell and its fanin nets to temporary storage
        Vec_PtrPush( vCells, &(cells[pObj.Id]) );
        // inlining of macro Abc_ObjForEachFanin
        for  k = 0; (k < Abc_ObjFaninNum(pObj)) && (((pFanin) = Abc_ObjFanin(pObj, k)), 1); k++  {
            Vec_PtrPushUnique( vNets, &(nets[pFanin.Id]) );
        }
    }

    // go through the modified nets
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vUpdatedNets)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vUpdatedNets, i)), 1); i++  {
        assert(!Abc_ObjIsComplement(pObj), "!Abc_ObjIsComplement(pObj)");
        if Abc_ObjType(pObj) == ABC_OBJ_NONE { // dead node
            continue; }
        Abc_PlaceUpdateNet( pObj );
    }

    // update the placement
//    fastPlace( Vec_PtrSize(vCells), (ConcreteCell **)Vec_PtrArray(vCells), 
//               Vec_PtrSize(vNets), (ConcreteNet **)Vec_PtrArray(vNets) );

    // clean up
    Vec_PtrFree( vCells );
    Vec_PtrFree( vNets );
}

/**Function*************************************************************

  Synopsis    [This procedure is called before the writing start.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_PlaceBegin(pNtk *Abc_Ntk_t) {
    var pObj *Abc_Obj_t;
    var i int32;

    // allocate and clean internal storage
    nAllocSize = 5 * Abc_NtkObjNumMax(pNtk);
    cells = REALLOC(ConcreteCell, cells, nAllocSize);
    nets  = REALLOC(ConcreteNet, nets, nAllocSize);
    memset( cells, 0, sizeof(ConcreteCell) * nAllocSize );
    memset( nets, 0, sizeof(ConcreteNet) * nAllocSize );

    // create AbstractCells
    //   1: pad
    //   2: and
    if !(abstractCells != nil) {
        abstractCells = ALLOC(AbstractCell,2); }

    abstractCells[0].m_height = 1.0;
    abstractCells[0].m_width = 1.0;
    abstractCells[0].m_label = "pio";
    abstractCells[0].m_pad = 1;

    abstractCells[1].m_height = 1.0;
    abstractCells[1].m_width = 1.0;
    abstractCells[1].m_label = "and";
    abstractCells[1].m_pad = 0;

    // input pads
    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(pNtk)) && (((pObj) = Abc_NtkCi(pNtk, i)), 1); i++  {
        Abc_PlaceCreateCell( pObj, 0 );
    }

    // ouput pads
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(pNtk)) && (((pObj) = Abc_NtkCo(pNtk, i)), 1); i++  {
        Abc_PlaceCreateCell( pObj, 0 );
    }

    // AND nodes
    // inlining of macro Abc_AigForEachAnd
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pObj) == nil || !Abc_AigNodeIsAnd(pObj) {} else {
        Abc_PlaceCreateCell( pObj, 1 ); }
    }

    // all nets
    // inlining of macro Abc_NtkForEachObj
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {            if (pObj) == nil {} else {
        if !Abc_ObjIsCi(pObj) && !Abc_ObjIsNode(pObj) {
            continue; }
        Abc_PlaceUpdateNet( pObj );
    }
    }

    globalPreplace(float32(0.8));
    globalPlace();
}

/**Function*************************************************************

  Synopsis    [This procedure is called after the writing completes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_PlaceEnd(pNtk *Abc_Ntk_t) {
    var i int32;


    // clean up
    for  i = 0; i < nAllocSize; i++  {
        FREE( nets[i].m_terms );
    }
    FREE( abstractCells );
    FREE( cells );
    FREE( nets );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

