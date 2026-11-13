package ssw

import "time"

/**CFile****************************************************************

  FileName    [sswClass.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Inductive prover with constraints.]

  Synopsis    [Representation of candidate equivalence classes.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - September 1, 2008.]

  Revision    [$Id: sswClass.c,v 1.00 2008/09/01 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "sswInt.h"

//ABC_NAMESPACE_IMPL_START


/*
    The candidate equivalence classes are stored as a vector of pointers 
    to the array of pointers to the nodes in each class.
    The first node of the class is its representative node.
    The representative has the smallest topological order among the class nodes.
    The nodes inside each class are ordered according to their topological order.
    The classes are ordered according to the topo order of their representatives.
*/

// internal representation of candidate equivalence classes


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////




// iterator through the equivalence classes
// Original macro:
// #define Ssw_ManForEachClass( p, ppClass, i )                 \
//     for ( i = 0; i < Aig_ManObjNumMax(p->pAig); i++ )        \
//         if ( ((ppClass) = p->pId2Class[i]) == NULL ) {} else

// iterator through the nodes in one class
// Original macro:
// #define Ssw_ClassForEachNode( p, pRepr, pNode, i )           \
//     for ( i = 0; i < p->pClassSizes[pRepr->Id]; i++ )        \
//         if ( ((pNode) = p->pId2Class[pRepr->Id][i]) == NULL ) {} else


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Creates one equivalence class.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Removes one equivalence class.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Starts representation of equivalence classes.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Starts representation of equivalence classes.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Stop representation of equivalence classes.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Stop representation of equivalence classes.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Stop representation of equivalence classes.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Stop representation of equivalence classes.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Stop representation of equivalence classes.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Stop representation of equivalence classes.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Stop representation of equivalence classes.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Checks candidate equivalence classes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Prints simulation classes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_ClassesPrintOne(p *Ssw_Cla_t, pRepr *Aig_Obj_t) {
    var pObj *Aig_Obj_t;
    var i int32;
    Abc_Print( 1, "{ " );
    // inlining of macro Ssw_ClassForEachNode
    for  i = 0; i < p.pClassSizes[pRepr.Id]; i++  {                if ((pObj) = p.pId2Class[pRepr.Id][i]) == nil {} else {
        Abc_Print( 1, "%d(%d,%d,%d) ", pObj.Id, pObj.Level,
        Aig_SupportSize(p.pAig,pObj), Aig_NodeMffcSupp(p.pAig,pObj,0,nil) ); }
    }
    Abc_Print( 1, "}\n" );
}

/**Function*************************************************************

  Synopsis    [Prints simulation classes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_ClassesPrint(p *Ssw_Cla_t, fVeryVerbose int32) {
    var ppClass **Aig_Obj_t;
    var pObj *Aig_Obj_t;
    var i int32;
    Abc_Print( 1, "Equiv classes: Const1 = %5d. Class = %5d. Lit = %5d.\n",
        p.nCands1, p.nClasses, p.nCands1+p.nLits );
    if !(fVeryVerbose != 0) {
        return; }
    Abc_Print( 1, "Constants { " );
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vObjs, i)), 1); i++  { if (pObj) == nil {} else
        if Ssw_ObjIsConst1Cand( p.pAig, pObj ) {
            Abc_Print( 1, "%d(%d,%d,%d) ", pObj.Id, pObj.Level,
            Aig_SupportSize(p.pAig,pObj), Aig_NodeMffcSupp(p.pAig,pObj,0,nil) ); }
    }
    Abc_Print( 1, "}\n" );
    // inlining of macro Ssw_ManForEachClass
    for  i = 0; i < Aig_ManObjNumMax(p.pAig); i++  {                if ((ppClass) = p.pId2Class[i]) == nil {} else {
        Abc_Print( 1, "%3d (%3d) : ", i, p.pClassSizes[i] );
        Ssw_ClassesPrintOne( p, ppClass[0] );
    }
    }
    Abc_Print( 1, "\n" );
}

/**Function*************************************************************

  Synopsis    [Prints simulation classes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Takes the set of const1 cands and rehashes them using sim info.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Creates initial simulation classes.]

  Description [Assumes that simulation info is assigned.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_ClassesPrepare(pAig *Aig_Man_t, nFramesK int32, fLatchCorr int32, fConstCorr int32, fOutputCorr int32, nMaxLevs int32, fVerbose int32) *Ssw_Cla_t {
//    int nFrames =  4;
//    int nWords  =  1;
//    int nIters  = 16;

//    int nFrames = 32;
//    int nWords  =  4;
//    int nIters  =  0;

    var nFrames int32  =  Abc_MaxInt( nFramesK, 4 );
    var nWords int32   =  2;
    var nIters int32   = 16;
    var p *Ssw_Cla_t;
    var pSml *Ssw_Sml_t;
    var vCands *Vec_Ptr_t;
    var pObj *Aig_Obj_t;
    var i, k, RetValue int32;
    var clk abctime;

    // start the classes
    p = Ssw_ClassesStart( pAig );
    p.fConstCorr = fConstCorr;

    // perform sequential simulation
clk = Abc_Clock();
    pSml = Ssw_SmlSimulateSeq( pAig, 0, nFrames, nWords );
if fVerbose != 0 {
    Abc_Print( 1, "Allocated %.2f MB to store simulation information.\n",
        1.0*(sizeof(unsigned) * Aig_ManObjNumMax(pAig) * nFrames * nWords)/(1<<20) );
    Abc_Print( 1, "Initial simulation of %d frames with %d words.     ", nFrames, nWords );
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
}

    // set comparison procedures
clk = Abc_Clock();
    Ssw_ClassesSetData( p, pSml, (unsigned(*)(void *,Aig_Obj_t *))Ssw_SmlObjHashWord, (int(*)(void *,Aig_Obj_t *))Ssw_SmlObjIsConstWord, (int(*)(void *,Aig_Obj_t *,Aig_Obj_t *))Ssw_SmlObjsAreEqualWord );

    // collect nodes to be considered as candidates
    vCands = Vec_PtrAlloc( 1000 );
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        if fLatchCorr != 0 {
            if !Saig_ObjIsLo(p.pAig, pObj) {
                continue; }
        
        } else {
            if !Aig_ObjIsNode(pObj) && !Aig_ObjIsCi(pObj) {
                continue; }
            // skip the node with more that the given number of levels
            if (nMaxLevs != 0) && int32(pObj.Level) > nMaxLevs {
                continue; }
        }
        Vec_PtrPush( vCands, pObj );
    }
    }

    // this change will consider all PO drivers
    if fOutputCorr != 0 {
        Vec_PtrClear( vCands );
        // inlining of macro Aig_ManForEachObj
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p.pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
            pObj.fMarkB = 0; }
        }
        // inlining of macro Saig_ManForEachPo
        // inlining of macro Vec_PtrForEachEntryStop
        for  i = 0; (i < Saig_ManPoNum(p.pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCos, i)), 1); i++  {
            if Aig_ObjIsCand(Aig_ObjFanin0(pObj)) {
                Aig_ObjFanin0(pObj).fMarkB = 1; }
        }
        // inlining of macro Aig_ManForEachObj
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p.pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vObjs, i)), 1); i++  { if (pObj) == nil {} else
            if pObj.fMarkB {
                Vec_PtrPush( vCands, pObj ); }
        }
        // inlining of macro Aig_ManForEachObj
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p.pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
            pObj.fMarkB = 0; }
        }
    }

    // allocate room for classes
    // inlining of macro ABC_ALLOC
    p.pMemClasses = ((Aig_Obj_t * *) malloc(sizeof(Aig_Obj_t *) * (size_t)(Vec_PtrSize(vCands))));
    p.pMemClassesFree = p.pMemClasses;

    // now it is time to refine the classes
    Ssw_ClassesPrepareRehash( p, vCands, fConstCorr );
if fVerbose != 0 {
    Abc_Print( 1, "Collecting candidate equivalence classes.        " );
// inlining of macro ABC_PRT
Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
}

clk = Abc_Clock();
    // perform iterative refinement using simulation
    for  i = 1; i < nIters; i++  {
        // collect const1 candidates
        Vec_PtrClear( vCands );
        // inlining of macro Aig_ManForEachObj
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(p.pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vObjs, k)), 1); k++  { if (pObj) == nil {} else
            if Ssw_ObjIsConst1Cand( p.pAig, pObj ) {
                Vec_PtrPush( vCands, pObj ); }
        }
        assert(Vec_PtrSize(vCands) == p.nCands1, "Vec_PtrSize(vCands) == p.nCands1");
        // perform new round of simulation
        Ssw_SmlResimulateSeq( pSml );
        // check equivalence classes
        RetValue = Ssw_ClassesPrepareRehash( p, vCands, fConstCorr );
        if RetValue == 0 {
            break; }
    }
    Ssw_SmlStop( pSml );
    Vec_PtrFree( vCands );
if fVerbose != 0 {
    Abc_Print( 1, "Simulation of %d frames with %d words (%2d rounds). ",
        nFrames, nWords, i-1 );
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
}
    Ssw_ClassesCheck( p );
//    Ssw_ClassesPrint( p, 0 );
    return p;
}

/**Function*************************************************************

  Synopsis    [Creates initial simulation classes.]

  Description [Assumes that simulation info is assigned.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Creates initial simulation classes.]

  Description [Assumes that simulation info is assigned.]

  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Creates initial simulation classes.]

  Description [Assumes that simulation info is assigned.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Creates classes from the temporary representation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Creates classes from the temporary representation.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Iteratively refines the classes after simulation.]

  Description [Returns the number of refinements performed.]

  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Refines the classes after simulation.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Refines the classes after simulation.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Refine the group of constant 1 nodes.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Refine the group of constant 1 nodes.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/



////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END
