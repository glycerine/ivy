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
func Ssw_ClassesCheck(p *Ssw_Cla_t) {
    var pObj *Aig_Obj_t;
    var pPrev *Aig_Obj_t;
    var ppClass **Aig_Obj_t;
    var i, k, nLits, nClasses, nCands1 int32;
    nClasses = nLits = 0;
    // inlining of macro Ssw_ManForEachClass
    for  k = 0; k < Aig_ManObjNumMax(p.pAig); k++  {                if ((ppClass) = p.pId2Class[k]) == nil {} else {
        pPrev = nil;
        assert(p.pClassSizes[ppClass[0].Id] >= 2, "p.pClassSizes[ppClass[0].Id] >= 2");
        // inlining of macro Ssw_ClassForEachNode
        for  i = 0; i < p.pClassSizes[ppClass[0].Id]; i++  {                if ((pObj) = p.pId2Class[ppClass[0].Id][i]) == nil {} else {
            if i == 0 {
                assert(Aig_ObjRepr(p.pAig, pObj) == nil, "Aig_ObjRepr(p.pAig, pObj) == nil");
            } else {
                assert(Aig_ObjRepr(p.pAig, pObj) == ppClass[0], "Aig_ObjRepr(p.pAig, pObj) == ppClass[0]");
                assert(pPrev.Id < pObj.Id, "pPrev.Id < pObj.Id");
                nLits++;
            }
            pPrev = pObj;
        }
        }
        nClasses++;
    }
    }
    nCands1 = 0;
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        nCands1 += Ssw_ObjIsConst1Cand( p.pAig, pObj ); }
    }
    assert(p.nLits == nLits, "p.nLits == nLits");
    assert(p.nCands1 == nCands1, "p.nCands1 == nCands1");
    assert(p.nClasses == nClasses, "p.nClasses == nClasses");
}

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
func Ssw_ClassesRemoveNode(p *Ssw_Cla_t, pObj *Aig_Obj_t) {
    var pRepr, pTemp *Aig_Obj_t;
    assert(p.pClassSizes[pObj.Id] == 0, "p.pClassSizes[pObj.Id] == 0");
    assert(p.pId2Class[pObj.Id] == nil, "p.pId2Class[pObj.Id] == nil");
    pRepr = Aig_ObjRepr( p.pAig, pObj );
    assert(pRepr != nil, "pRepr != nil");
//    Vec_PtrPush( p->vRefined, pObj );
    if Ssw_ObjIsConst1Cand( p.pAig, pObj ) {
        assert(p.pClassSizes[pRepr.Id] == 0, "p.pClassSizes[pRepr.Id] == 0");
        assert(p.pId2Class[pRepr.Id] == nil, "p.pId2Class[pRepr.Id] == nil");
        Aig_ObjSetRepr( p.pAig, pObj, nil );
        p.nCands1--;
        return;
    }
//    Vec_PtrPush( p->vRefined, pRepr );
    Aig_ObjSetRepr( p.pAig, pObj, nil );
    assert(p.pId2Class[pRepr.Id][0] == pRepr, "p.pId2Class[pRepr.Id][0] == pRepr");
    assert(p.pClassSizes[pRepr.Id] >= 2, "p.pClassSizes[pRepr.Id] >= 2");
    if p.pClassSizes[pRepr.Id] == 2 {
        p.pId2Class[pRepr.Id] = nil;
        p.nClasses--;
        p.pClassSizes[pRepr.Id] = 0;
        p.nLits--;
    
    } else {
        var i, k int32;
        // remove the entry from the class
        // inlining of macro Ssw_ClassForEachNode
        for  i = 0; i < p.pClassSizes[pRepr.Id]; i++  {                if ((pTemp) = p.pId2Class[pRepr.Id][i]) == nil {} else
            if pTemp != pObj {
                p.pId2Class[pRepr.Id][k++] = pTemp; }
        }
        assert(k + 1 == p.pClassSizes[pRepr.Id], "k + 1 == p.pClassSizes[pRepr.Id]");
        // reduce the class
        p.pClassSizes[pRepr.Id]--;
        p.nLits--;
    }
}

/**Function*************************************************************

  Synopsis    [Takes the set of const1 cands and rehashes them using sim info.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_ClassesPrepareRehash(p *Ssw_Cla_t, vCands *Vec_Ptr_t, fConstCorr int32) int32 {
//    Aig_Man_t * pAig = p->pAig;
    var ppTable, ppNexts, ppClassNew **Aig_Obj_t;
    var pObj, pTemp, pRepr *Aig_Obj_t;
    var i, k, nTableSize, nNodes, iEntry, nEntries, nEntries2 int32;

    // allocate the hash table hashing simulation info into nodes
    nTableSize = Abc_PrimeCudd( Vec_PtrSize(vCands)/2 );
    // inlining of macro ABC_CALLOC
    ppTable = ((Aig_Obj_t * *) calloc((size_t)(nTableSize), sizeof(Aig_Obj_t *)));
    // inlining of macro ABC_CALLOC
    ppNexts = ((Aig_Obj_t * *) calloc((size_t)(Aig_ManObjNumMax(p.pAig)), sizeof(Aig_Obj_t *)));

    // sort through the candidates
    nEntries = 0;
    p.nCands1 = 0;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vCands)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(vCands, i)), 1); i++  {
        assert(p.pClassSizes[pObj.Id] == 0, "p.pClassSizes[pObj.Id] == 0");
        Aig_ObjSetRepr( p.pAig, pObj, nil );
        // check if the node belongs to the class of constant 1
        if p.pFuncNodeIsConst( p.pManData, pObj ) {
            Ssw_ObjSetConst1Cand( p.pAig, pObj );
            p.nCands1++;
            continue;
        }
        if fConstCorr != 0 {
            continue; }
        // hash the node by its simulation info
        iEntry = p.pFuncNodeHash( p.pManData, pObj ) % nTableSize;
        // add the node to the class
        if ppTable[iEntry] == nil {
            ppTable[iEntry] = pObj;
        
        } else {
            // set the representative of this node
            pRepr = ppTable[iEntry];
            Aig_ObjSetRepr( p.pAig, pObj, pRepr );
            // add node to the table
            if Ssw_ObjNext( ppNexts, pRepr ) == nil { // this will be the second entry
                p.pClassSizes[pRepr.Id]++;
                nEntries++;
            }
            // add the entry to the list
            Ssw_ObjSetNext( ppNexts, pObj, Ssw_ObjNext( ppNexts, pRepr ) );
            Ssw_ObjSetNext( ppNexts, pRepr, pObj );
            p.pClassSizes[pRepr.Id]++;
            nEntries++;
        }
    }

    // copy the entries into storage in the topological order
    nEntries2 = 0;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vCands)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(vCands, i)), 1); i++  {
        nNodes = p.pClassSizes[pObj.Id];
        // skip the nodes that are not representatives of non-trivial classes
        if nNodes == 0 {
            continue; }
        assert(nNodes > 1, "nNodes > 1");
        // add the nodes to the class in the topological order
        ppClassNew = p.pMemClassesFree + nEntries2;
        ppClassNew[0] = pObj;
        for  pTemp = Ssw_ObjNext(ppNexts, pObj), k = 1; (pTemp != nil);
              pTemp = Ssw_ObjNext(ppNexts, pTemp), k++  {
            ppClassNew[nNodes-k] = pTemp;
        }
        // add the class of nodes
        p.pClassSizes[pObj.Id] = 0;
        Ssw_ObjAddClass( p, pObj, ppClassNew, nNodes );
        // increment the number of entries
        nEntries2 += nNodes;
    }
    p.pMemClassesFree += nEntries2;
    assert(nEntries == nEntries2, "nEntries == nEntries2");
    ppTable = nil;
    ppNexts = nil;
    // now it is time to refine the classes
    return Ssw_ClassesRefine( p, 1 );
}

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
func Ssw_ClassesPrepareSimple(pAig *Aig_Man_t, fLatchCorr int32, nMaxLevs int32) *Ssw_Cla_t {
    var p *Ssw_Cla_t;
    var pObj *Aig_Obj_t;
    var i int32;
    // start the classes
    p = Ssw_ClassesStart( pAig );
    // go through the nodes
    p.nCands1 = 0;
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        if fLatchCorr != 0 {
            if !Saig_ObjIsLo(pAig, pObj) {
                continue; }
        
        } else {
            if !Aig_ObjIsNode(pObj) && !Saig_ObjIsLo(pAig, pObj) {
                continue; }
            // skip the node with more that the given number of levels
            if (nMaxLevs != 0) && int32(pObj.Level) > nMaxLevs {
                continue; }
        }
        Ssw_ObjSetConst1Cand( pAig, pObj );
        p.nCands1++;
    }
    }
    // allocate room for classes
    // inlining of macro ABC_ALLOC
    p.pMemClassesFree = p.pMemClasses = ((Aig_Obj_t * *) malloc(sizeof(Aig_Obj_t *) * (size_t)(p.nCands1)));
//    Ssw_ClassesPrint( p, 0 );
    return p;
}

/**Function*************************************************************

  Synopsis    [Creates initial simulation classes.]

  Description [Assumes that simulation info is assigned.]

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_ClassesPrepareFromReprs(pAig *Aig_Man_t) *Ssw_Cla_t {
    var p *Ssw_Cla_t;
    var pObj, pRepr *Aig_Obj_t;
    var pClassSizes *int32;
    var nEntries int32;
    var i int32;
    // start the classes
    p = Ssw_ClassesStart( pAig );
    // allocate memory for classes
    // inlining of macro ABC_CALLOC
    p.pMemClasses = ((Aig_Obj_t * *) calloc((size_t)(Aig_ManObjNumMax(pAig)), sizeof(Aig_Obj_t *)));
    // count classes
    p.nCands1 = 0;
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        if Ssw_ObjIsConst1Cand(pAig, pObj) {
            p.nCands1++;
            continue;
        }
        if pRepr = Aig_ObjRepr(pAig, pObj) {
            if p.pClassSizes[pRepr.Id]++ == 0 {
                p.pClassSizes[pRepr.Id]++; }
        }
    }
    }
    // add nodes
    nEntries = 0;
    p.nClasses = 0;
    // inlining of macro ABC_CALLOC
    pClassSizes = ((int *) calloc((size_t)(Aig_ManObjNumMax(pAig)), sizeof(int)));
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        if p.pClassSizes[i] {
            p.pId2Class[i] = p.pMemClasses + nEntries;
            nEntries += p.pClassSizes[i];
            p.pId2Class[i][pClassSizes[i]++] = pObj;
            p.nClasses++;
            continue;
        }
        if Ssw_ObjIsConst1Cand(pAig, pObj) {
            continue; }
        if pRepr = Aig_ObjRepr(pAig, pObj) {
            p.pId2Class[pRepr.Id][pClassSizes[pRepr.Id]++] = pObj; }
    }
    }
    p.pMemClassesFree = p.pMemClasses + nEntries;
    p.nLits = nEntries - p.nClasses;
    assert(memcmp(pClassSizes, p.pClassSizes, sizeof(int)*Aig_ManObjNumMax(pAig)) == 0, "memcmp(pClassSizes, p.pClassSizes, sizeof(int)*Aig_ManObjNumMax(pAig)) == 0");
    pClassSizes = nil;
//    Abc_Print( 1, "After converting:\n" );
//    Ssw_ClassesPrint( p, 0 );
    return p;
}

/**Function*************************************************************

  Synopsis    [Creates initial simulation classes.]

  Description [Assumes that simulation info is assigned.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_ClassesPrepareTargets(pAig *Aig_Man_t) *Ssw_Cla_t {
    var p *Ssw_Cla_t;
    var pObj *Aig_Obj_t;
    var i int32;
    // start the classes
    p = Ssw_ClassesStart( pAig );
    // go through the nodes
    p.nCands1 = 0;
    // inlining of macro Saig_ManForEachPo
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Saig_ManPoNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCos, i)), 1); i++  {
        Ssw_ObjSetConst1Cand( pAig, Aig_ObjFanin0(pObj) );
        p.nCands1++;
    }
    // allocate room for classes
    // inlining of macro ABC_ALLOC
    p.pMemClassesFree = p.pMemClasses = ((Aig_Obj_t * *) malloc(sizeof(Aig_Obj_t *) * (size_t)(p.nCands1)));
//    Ssw_ClassesPrint( p, 0 );
    return p;
}

/**Function*************************************************************

  Synopsis    [Creates classes from the temporary representation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_ClassesPreparePairs(pAig *Aig_Man_t, pvClasses **Vec_Int_t) *Ssw_Cla_t {
    var p *Ssw_Cla_t;
    var ppClassNew **Aig_Obj_t;
    var pObj, pRepr, pPrev *Aig_Obj_t;
    var i, k, nTotalObjs, nEntries, Entry int32;
    // start the classes
    p = Ssw_ClassesStart( pAig );
    // count the number of entries in the classes
    nTotalObjs = 0;
    for  i = 0; i < Aig_ManObjNumMax(pAig); i++  {
        nTotalObjs += pvClasses[i] ? Vec_IntSize(pvClasses[i]) : 0;
    }
    // allocate memory for classes
    // inlining of macro ABC_ALLOC
    p.pMemClasses = ((Aig_Obj_t * *) malloc(sizeof(Aig_Obj_t *) * (size_t)(nTotalObjs)));
    // create constant-1 class
    if pvClasses[0] {
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(pvClasses[0])) && (((Entry) = Vec_IntEntry(pvClasses[0], i)), 1); i++  {
        assert((i == 0) == (Entry == 0), "(i == 0) == (Entry == 0)");
        if i == 0 {
            continue; }
        pObj = Aig_ManObj( pAig, Entry );
        Ssw_ObjSetConst1Cand( pAig, pObj );
        p.nCands1++;
    } }
    // create classes
    nEntries = 0;
    for  i = 1; i < Aig_ManObjNumMax(pAig); i++  {
        if pvClasses[i] == nil {
            continue; }
        // get room for storing the class
        ppClassNew = p.pMemClasses + nEntries;
        nEntries  += Vec_IntSize( pvClasses[i] );
        // store the nodes of the class
        pPrev = pRepr = Aig_ManObj( pAig, Vec_IntEntry(pvClasses[i],0) );
        ppClassNew[0] = pRepr;
        // inlining of macro Vec_IntForEachEntryStart
        for  k = 1; (k < Vec_IntSize(pvClasses[i])) && (((Entry) = Vec_IntEntry(pvClasses[i], k)), 1); k++  {
            pObj = Aig_ManObj( pAig, Entry );
            assert(pPrev.Id < pObj.Id, "pPrev.Id < pObj.Id");
            pPrev = pObj;
            ppClassNew[k] = pObj;
            Aig_ObjSetRepr( pAig, pObj, pRepr );
        }
        // create new class
        Ssw_ObjAddClass( p, pRepr, ppClassNew, Vec_IntSize(pvClasses[i]) );
    }
    // prepare room for new classes
    p.pMemClassesFree = p.pMemClasses + nEntries;
    Ssw_ClassesCheck( p );
//    Ssw_ClassesPrint( p, 0 );
    return p;
}

/**Function*************************************************************

  Synopsis    [Creates classes from the temporary representation.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_ClassesPreparePairsSimple(pMiter *Aig_Man_t, vPairs *Vec_Int_t) *Ssw_Cla_t {
    var p *Ssw_Cla_t;
    var ppClassNew **Aig_Obj_t;
    var pObj, pRepr *Aig_Obj_t;
    var i int32;
    // start the classes
    p = Ssw_ClassesStart( pMiter );
    // allocate memory for classes
    // inlining of macro ABC_ALLOC
    p.pMemClasses = ((Aig_Obj_t * *) malloc(sizeof(Aig_Obj_t *) * (size_t)(Vec_IntSize(vPairs))));
    // create classes
    for  i = 0; i < Vec_IntSize(vPairs); i += 2  {
        pRepr = Aig_ManObj( pMiter, Vec_IntEntry(vPairs, i) );
        pObj  = Aig_ManObj( pMiter, Vec_IntEntry(vPairs, i+1) );
        assert(Aig_ObjId(pRepr) < Aig_ObjId(pObj), "Aig_ObjId(pRepr) < Aig_ObjId(pObj)");
        Aig_ObjSetRepr( pMiter, pObj, pRepr );
        // get room for storing the class
        ppClassNew = p.pMemClasses + i;
        ppClassNew[0] = pRepr;
        ppClassNew[1] = pObj;
        // create new class
        Ssw_ObjAddClass( p, pRepr, ppClassNew, 2 );
    }
    // prepare room for new classes
    p.pMemClassesFree = nil;
    Ssw_ClassesCheck( p );
//    Ssw_ClassesPrint( p, 0 );
    return p;
}

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
func Ssw_ClassesRefineConst1Group(p *Ssw_Cla_t, vRoots *Vec_Ptr_t, fRecursive int32) int32 {
    var pObj *Aig_Obj_t;
    var pReprNew *Aig_Obj_t;
    var ppClassNew **Aig_Obj_t;
    var i int32;
    if Vec_PtrSize(vRoots) == 0 {
        return 0; }
    // collect the nodes to be refined
    Vec_PtrClear( p.vClassNew );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vRoots)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(vRoots, i)), 1); i++  {
        if !p.pFuncNodeIsConst( p.pManData, pObj ) {
            Vec_PtrPush( p.vClassNew, pObj ); }
    }
    // check if there is a new class
    if Vec_PtrSize(p.vClassNew) == 0 {
        return 0; }
    p.nCands1 -= Vec_PtrSize(p.vClassNew);
    pReprNew = (Aig_Obj_t *)Vec_PtrEntry( p.vClassNew, 0 );
    Aig_ObjSetRepr( p.pAig, pReprNew, nil );
    if Vec_PtrSize(p.vClassNew) == 1 {
        return 1; }
    // create a new class composed of these nodes
    ppClassNew = p.pMemClassesFree;
    p.pMemClassesFree += Vec_PtrSize(p.vClassNew);
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vClassNew)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vClassNew, i)), 1); i++  {
        ppClassNew[i] = pObj;
        Aig_ObjSetRepr( p.pAig, pObj, i? pReprNew : nil );
    }
    Ssw_ObjAddClass( p, pReprNew, ppClassNew, Vec_PtrSize(p.vClassNew) );
    // refine them recursively
    if fRecursive != 0 {
        return 1 + Ssw_ClassesRefineOneClass( p, pReprNew, 1 ); }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Refine the group of constant 1 nodes.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_ClassesRefineConst1(p *Ssw_Cla_t, fRecursive int32) int32 {
    var pObj *Aig_Obj_t;
    var pReprNew *Aig_Obj_t;
    var ppClassNew **Aig_Obj_t;
    var i int32;
    // collect the nodes to be refined
    Vec_PtrClear( p.vClassNew );
    for  i = 0; i < Vec_PtrSize(p.pAig.vObjs); i++  {
        if p.pAig.pReprs[i] == Aig_ManConst1(p.pAig) {
            pObj = Aig_ManObj( p.pAig, i );
            if !p.pFuncNodeIsConst( p.pManData, pObj ) {
                Vec_PtrPush( p.vClassNew, pObj );
//                Vec_PtrPush( p->vRefined, pObj );
            }
        }
    }
    // check if there is a new class
    if Vec_PtrSize(p.vClassNew) == 0 {
        return 0; }
    if p.fConstCorr {
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p.vClassNew)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vClassNew, i)), 1); i++  {
            Aig_ObjSetRepr( p.pAig, pObj, nil );
        }
        return 1;
    }
    p.nCands1 -= Vec_PtrSize(p.vClassNew);
    pReprNew = (Aig_Obj_t *)Vec_PtrEntry( p.vClassNew, 0 );
    Aig_ObjSetRepr( p.pAig, pReprNew, nil );
    if Vec_PtrSize(p.vClassNew) == 1 {
        return 1; }
    // create a new class composed of these nodes
    ppClassNew = p.pMemClassesFree;
    p.pMemClassesFree += Vec_PtrSize(p.vClassNew);
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vClassNew)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vClassNew, i)), 1); i++  {
        ppClassNew[i] = pObj;
        Aig_ObjSetRepr( p.pAig, pObj, i? pReprNew : nil );
    }
    Ssw_ObjAddClass( p, pReprNew, ppClassNew, Vec_PtrSize(p.vClassNew) );
    // refine them recursively
    if fRecursive != 0 {
        return 1 + Ssw_ClassesRefineOneClass( p, pReprNew, 1 ); }
    return 1;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END
