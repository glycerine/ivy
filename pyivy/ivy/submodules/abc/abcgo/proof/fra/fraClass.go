package fra

import "fmt"

/**CFile****************************************************************

  FileName    [fraClass.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [New FRAIG package.]

  Synopsis    []

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 30, 2007.]

  Revision    [$Id: fraClass.c,v 1.00 2007/06/30 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "fra.h"

//ABC_NAMESPACE_IMPL_START


/*
    The candidate equivalence classes are stored as a vector of pointers 
    to the array of pointers to the nodes in each class.
    The first node of the class is its representative node.
    The representative has the smallest topological order among the class nodes.
    The nodes inside each class are ordered according to their topological order.
    The classes are ordered according to the topological order of their representatives.
    The array of pointers to the class nodes is terminated with a NULL pointer.
    To enable dynamic addition of new classes (during class refinement),
    each array has at least as many NULLs in the end, as there are nodes in the class.
*/

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

func Fra_ObjNext(ppNexts **Aig_Obj_t, pObj *Aig_Obj_t) *Aig_Obj_t { return ppNexts[pObj.Id];  }
func Fra_ObjSetNext(ppNexts **Aig_Obj_t, pObj *Aig_Obj_t, pNext *Aig_Obj_t) { ppNexts[pObj.Id] = pNext; }

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Starts representation of equivalence classes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_ClassesStart(pAig *Aig_Man_t) *Fra_Cla_t {
    var p *Fra_Cla_t;
    // inlining of macro ABC_ALLOC
    p = ((Fra_Cla_t *) malloc(sizeof(Fra_Cla_t) * (size_t)(1)));
    memset( p, 0, sizeof(Fra_Cla_t) );
    p.pAig = pAig;
    // inlining of macro ABC_ALLOC
    p.pMemRepr  = ((Aig_Obj_t * *) malloc(sizeof(Aig_Obj_t *) * (size_t)(Aig_ManObjNumMax(pAig))));
    memset( p.pMemRepr, 0, sizeof(Aig_Obj_t *) * Aig_ManObjNumMax(pAig) );
    p.vClasses     = Vec_PtrAlloc( 100 );
    p.vClasses1    = Vec_PtrAlloc( 100 );
    p.vClassesTemp = Vec_PtrAlloc( 100 );
    p.vClassOld    = Vec_PtrAlloc( 100 );
    p.vClassNew    = Vec_PtrAlloc( 100 );
    p.pFuncNodeHash      = Fra_SmlNodeHash;
    p.pFuncNodeIsConst   = Fra_SmlNodeIsConst;
    p.pFuncNodesAreEqual = Fra_SmlNodesAreEqual;
    return p;
}

/**Function*************************************************************

  Synopsis    [Stop representation of equivalence classes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_ClassesStop(p *Fra_Cla_t) {
    p.pMemClasses = nil;
    p.pMemRepr = nil;
    if p.vClassesTemp { Vec_PtrFree( p.vClassesTemp ); }
    if p.vClassNew {    Vec_PtrFree( p.vClassNew ); }
    if p.vClassOld {    Vec_PtrFree( p.vClassOld ); }
    if p.vClasses1 {    Vec_PtrFree( p.vClasses1 ); }
    if p.vClasses {     Vec_PtrFree( p.vClasses ); }
    if p.vImps {        Vec_IntFree( p.vImps ); }
    p = nil;
}

/**Function*************************************************************

  Synopsis    [Starts representation of equivalence classes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_ClassesCopyReprs(p *Fra_Cla_t, vFailed *Vec_Ptr_t) {
    var pObj *Aig_Obj_t;
    var i int32;
    Aig_ManReprStart( p.pAig, Aig_ManObjNumMax(p.pAig) );
    memmove( p.pAig.pReprs, p.pMemRepr, sizeof(Aig_Obj_t *) * Aig_ManObjNumMax(p.pAig) );
    if Vec_PtrSize(p.vClasses1) == 0 && Vec_PtrSize(p.vClasses) == 0 {
        // inlining of macro Aig_ManForEachObj
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p.pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
            if p.pAig.pReprs[i] != nil {
                fmt.Printf("Classes are not cleared!\n"); }
            assert(p.pAig.pReprs[i] == nil, "p.pAig.pReprs[i] == nil");
        }
        }
    }
    if vFailed != nil {
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(vFailed)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(vFailed, i)), 1); i++  {
            p.pAig.pReprs[pObj.Id] = nil;
        } }
}

/**Function*************************************************************

  Synopsis    [Prints simulation classes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_ClassCount(pClass **Aig_Obj_t) int32 {
    var pTemp *Aig_Obj_t;
    var i int32;
    for  i = 0; (pTemp = pClass[i]); i++  {}
    return i;
}

/**Function*************************************************************

  Synopsis    [Count the number of literals.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_ClassesCountLits(p *Fra_Cla_t) int32 {
    var pClass **Aig_Obj_t;
    var i, nNodes, nLits int32;
    nLits = Vec_PtrSize( p.vClasses1 );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vClasses)) && (((pClass) = (Aig_Obj_t **)Vec_PtrEntry(p.vClasses, i)), 1); i++  {
        nNodes = Fra_ClassCount( pClass );
        assert(nNodes > 1, "nNodes > 1");
        nLits += nNodes - 1;
    }
    return nLits;
}

/**Function*************************************************************

  Synopsis    [Count the number of pairs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_ClassesCountPairs(p *Fra_Cla_t) int32 {
    var pClass **Aig_Obj_t;
    var i, nNodes, nPairs int32;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vClasses)) && (((pClass) = (Aig_Obj_t **)Vec_PtrEntry(p.vClasses, i)), 1); i++  {
        nNodes = Fra_ClassCount( pClass );
        assert(nNodes > 1, "nNodes > 1");
        nPairs += nNodes * (nNodes - 1) / 2;
    }
    return nPairs;
}

/**Function*************************************************************

  Synopsis    [Prints simulation classes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_PrintClass(p *Fra_Cla_t, pClass **Aig_Obj_t) {
    var pTemp *Aig_Obj_t;
    var i int32;
    for  i = 1; (pTemp = pClass[i]); i++  {
        assert(Fra_ClassObjRepr(pTemp) == pClass[0], "Fra_ClassObjRepr(pTemp) == pClass[0]");
    }
    fmt.Printf("{ ");
    for  i = 0; (pTemp = pClass[i]); i++  {
        fmt.Printf("%d(%d,%d) ", pTemp.Id, pTemp.Level, Aig_SupportSize(p.pAig,pTemp));
    }
    fmt.Printf("}\n");
}

/**Function*************************************************************

  Synopsis    [Prints simulation classes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_ClassesPrint(p *Fra_Cla_t, fVeryVerbose int32) {
    var pClass **Aig_Obj_t;
    var pObj *Aig_Obj_t;
    var i int32;

    fmt.Printf("Const = %5d. Class = %5d. Lit = %5d. ", Vec_PtrSize(p.vClasses1), Vec_PtrSize(p.vClasses), Fra_ClassesCountLits(p));
    if p.vImps && Vec_IntSize(p.vImps) > 0 {
        fmt.Printf("Imp = %5d. ", Vec_IntSize(p.vImps)); }
    fmt.Printf("\n");

    if fVeryVerbose != 0 {
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p.vClasses1)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vClasses1, i)), 1); i++  {
            assert(Fra_ClassObjRepr(pObj) == Aig_ManConst1(p.pAig), "Fra_ClassObjRepr(pObj) == Aig_ManConst1(p.pAig)");
        }
        fmt.Printf("Constants { ");
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p.vClasses1)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vClasses1, i)), 1); i++  {
            fmt.Printf("%d(%d,%d) ", pObj.Id, pObj.Level, Aig_SupportSize(p.pAig,pObj));
        }
        fmt.Printf("}\n");
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p.vClasses)) && (((pClass) = (Aig_Obj_t **)Vec_PtrEntry(p.vClasses, i)), 1); i++  {
            fmt.Printf("%3d (%3d) : ", i, Fra_ClassCount(pClass));
            Fra_PrintClass( p, pClass );
        }
        fmt.Printf("\n");
    }
}

/**Function*************************************************************

  Synopsis    [Creates initial simulation classes.]

  Description [Assumes that simulation info is assigned.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_ClassesPrepare(p *Fra_Cla_t, fLatchCorr int32, nMaxLevs int32) {
    var ppTable, ppNexts **Aig_Obj_t;
    var pObj, pTemp *Aig_Obj_t;
    var i, k, nTableSize, nEntries, nNodes, iEntry int32;

    // allocate the hash table hashing simulation info into nodes
    nTableSize = Abc_PrimeCudd( Aig_ManObjNumMax(p.pAig) );
    // inlining of macro ABC_FALLOC
    ppTable = ((Aig_Obj_t * *) memset(malloc(sizeof(Aig_Obj_t *) * (size_t)(nTableSize)), 0xff, sizeof(Aig_Obj_t *) * (size_t)(nTableSize))); 
    // inlining of macro ABC_FALLOC
    ppNexts = ((Aig_Obj_t * *) memset(malloc(sizeof(Aig_Obj_t *) * (size_t)(nTableSize)), 0xff, sizeof(Aig_Obj_t *) * (size_t)(nTableSize))); 
    memset( ppTable, 0, sizeof(Aig_Obj_t *) * nTableSize );

    // add all the nodes to the hash table
    Vec_PtrClear( p.vClasses1 );
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        if fLatchCorr != 0 {
            if !Aig_ObjIsCi(pObj) {
                continue; }
        
        } else {
            if !Aig_ObjIsNode(pObj) && !Aig_ObjIsCi(pObj) {
                continue; }
            // skip the node with more that the given number of levels
            if (nMaxLevs != 0) && int32(pObj.Level) > nMaxLevs {
                continue; }
        }
        // hash the node by its simulation info
        iEntry = p.pFuncNodeHash( pObj, nTableSize );
        // check if the node belongs to the class of constant 1
        if p.pFuncNodeIsConst( pObj ) {
            Vec_PtrPush( p.vClasses1, pObj );
            Fra_ClassObjSetRepr( pObj, Aig_ManConst1(p.pAig) );
            continue;
        }
        // add the node to the class
        if ppTable[iEntry] == nil {
            ppTable[iEntry] = pObj;
            Fra_ObjSetNext( ppNexts, pObj, pObj );
        
        } else {
            Fra_ObjSetNext( ppNexts, pObj, Fra_ObjNext(ppNexts,ppTable[iEntry]) );
            Fra_ObjSetNext( ppNexts, ppTable[iEntry], pObj );
        }
    }
    }

    // count the total number of nodes in the non-trivial classes
    // mark the representative nodes of each equivalence class
    nEntries = 0;
    for  i = 0; i < nTableSize; i++  {
        if ppTable[i] && ppTable[i] != Fra_ObjNext(ppNexts, ppTable[i]) {
            for  pTemp = Fra_ObjNext(ppNexts, ppTable[i]), k = 1; 
                  pTemp != ppTable[i]; 
                  pTemp = Fra_ObjNext(ppNexts, pTemp), k++  {}
            assert(k > 1, "k > 1");
            nEntries += k;
            // mark the node
            assert(ppTable[i].fMarkA == 0, "ppTable[i].fMarkA == 0");
            ppTable[i].fMarkA = 1;
        }
    }

    // allocate room for classes
    // inlining of macro ABC_ALLOC
    p.pMemClasses = ((Aig_Obj_t * *) malloc(sizeof(Aig_Obj_t *) * (size_t)(2*(nEntries + Vec_PtrSize(p.vClasses1)))));
    p.pMemClassesFree = p.pMemClasses + 2*nEntries;
 
    // copy the entries into storage in the topological order
    Vec_PtrClear( p.vClasses );
    nEntries = 0;
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        if !Aig_ObjIsNode(pObj) && !Aig_ObjIsCi(pObj) {
            continue; }
        // skip the nodes that are not representatives of non-trivial classes
        if pObj.fMarkA == 0 {
            continue; }
        pObj.fMarkA = 0;
        // add the class of nodes
        Vec_PtrPush( p.vClasses, p.pMemClasses + 2*nEntries );
        // count the number of entries in this class
        for  pTemp = Fra_ObjNext(ppNexts, pObj), k = 1; 
              pTemp != pObj; 
              pTemp = Fra_ObjNext(ppNexts, pTemp), k++  {}
        nNodes = k;
        assert(nNodes > 1, "nNodes > 1");
        // add the nodes to the class in the topological order
        p.pMemClasses[2*nEntries] = pObj;
        for  pTemp = Fra_ObjNext(ppNexts, pObj), k = 1; 
              pTemp != pObj; 
              pTemp = Fra_ObjNext(ppNexts, pTemp), k++  {
            p.pMemClasses[2*nEntries+nNodes-k] = pTemp;
            Fra_ClassObjSetRepr( pTemp, pObj );
        }
        // add as many empty entries
        p.pMemClasses[2*nEntries + nNodes] = nil;
        // increment the number of entries
        nEntries += k;
    }
    }
    ppTable = nil;
    ppNexts = nil;
    // now it is time to refine the classes
    Fra_ClassesRefine( p );
//    Fra_ClassesPrint( p, 0 );
}

/**Function*************************************************************

  Synopsis    [Refines one class using simulation info.]

  Description [Returns the new class if refinement happened.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_RefineClassOne(p *Fra_Cla_t, ppClass **Aig_Obj_t) **Aig_Obj_t {
    var pObj *Aig_Obj_t;
    var ppThis **Aig_Obj_t;
    var i int32;
    assert(ppClass[0] != nil && ppClass[1] != nil, "ppClass[0] != nil && ppClass[1] != nil");

    // check if the class is going to be refined
    for  ppThis = ppClass + 1; (pObj = *ppThis); ppThis++  {        
        if !p.pFuncNodesAreEqual(ppClass[0], pObj) {
            break; }
    }
    if pObj == nil {
        return nil; }
    // split the class
    Vec_PtrClear( p.vClassOld );
    Vec_PtrClear( p.vClassNew );
    Vec_PtrPush( p.vClassOld, ppClass[0] );
    for  ppThis = ppClass + 1; (pObj = *ppThis); ppThis++  {        
        if p.pFuncNodesAreEqual(ppClass[0], pObj) {
            Vec_PtrPush( p.vClassOld, pObj );
        } else {
            Vec_PtrPush( p.vClassNew, pObj ); }
    }
/*
    printf( "Refining class (" );
    Vec_PtrForEachEntry( Aig_Obj_t *, p->vClassOld, pObj, i )
        printf( "%d,", pObj->Id );
    printf( ") + (" );
    Vec_PtrForEachEntry( Aig_Obj_t *, p->vClassNew, pObj, i )
        printf( "%d,", pObj->Id );
    printf( ")\n" );
*/
    // put the nodes back into the class memory
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vClassOld)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vClassOld, i)), 1); i++  {
        ppClass[i] = pObj;
        ppClass[Vec_PtrSize(p.vClassOld)+i] = nil;
        Fra_ClassObjSetRepr( pObj, i? ppClass[0] : nil );
    }
    ppClass += 2*Vec_PtrSize(p.vClassOld);
    // put the new nodes into the class memory
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vClassNew)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vClassNew, i)), 1); i++  {
        ppClass[i] = pObj;
        ppClass[Vec_PtrSize(p.vClassNew)+i] = nil;
        Fra_ClassObjSetRepr( pObj, i? ppClass[0] : nil );
    }
    return ppClass;
}

/**Function*************************************************************

  Synopsis    [Iteratively refines the classes after simulation.]

  Description [Returns the number of refinements performed.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_RefineClassLastIter(p *Fra_Cla_t, vClasses *Vec_Ptr_t) int32 {
    var pClass, pClass2 **Aig_Obj_t;
    var nRefis int32;
    pClass = (Aig_Obj_t **)Vec_PtrEntryLast( vClasses );
    for  nRefis = 0; (pClass2 = Fra_RefineClassOne( p, pClass )); nRefis++  {
        // if the original class is trivial, remove it
        if pClass[1] == nil {
            Vec_PtrPop( vClasses ); }
        // if the new class is trivial, stop
        if pClass2[1] == nil {
            nRefis++;
            break;
        }
        // othewise, add the class and continue
        assert(pClass2[0] != nil, "pClass2[0] != nil");
        Vec_PtrPush( vClasses, pClass2 );
        pClass = pClass2;
    }
    return nRefis;
}

/**Function*************************************************************

  Synopsis    [Refines the classes after simulation.]

  Description [Assumes that simulation info is assigned. Returns the
  number of classes refined.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_ClassesRefine(p *Fra_Cla_t) int32 {
    var vTemp *Vec_Ptr_t;
    var pClass **Aig_Obj_t;
    var i, nRefis int32;
    // refine the classes
    nRefis = 0;
    Vec_PtrClear( p.vClassesTemp );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vClasses)) && (((pClass) = (Aig_Obj_t **)Vec_PtrEntry(p.vClasses, i)), 1); i++  {
        // add the class to the new array
        assert(pClass[0] != nil, "pClass[0] != nil");
        Vec_PtrPush( p.vClassesTemp, pClass );
        // refine the class iteratively
        nRefis += Fra_RefineClassLastIter( p, p.vClassesTemp );
    }
    // exchange the class representation
    vTemp = p.vClassesTemp;
    p.vClassesTemp = p.vClasses;
    p.vClasses = vTemp;
    return nRefis;
}

/**Function*************************************************************

  Synopsis    [Refines constant 1 equivalence class.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_ClassesRefine1(p *Fra_Cla_t, fRefineNewClass int32, pSkipped *int32) int32 {
    var pObj *Aig_Obj_t;
    var ppClass **Aig_Obj_t;
    var i int32;
    var k int32;
    var nRefis int32 = 1;
    // check if there is anything to refine
    if Vec_PtrSize(p.vClasses1) == 0 {
        return 0; }
    // make sure constant 1 class contains only non-constant nodes
    assert(Vec_PtrEntry(p.vClasses1,0) != Aig_ManConst1(p.pAig), "Vec_PtrEntry(p.vClasses1,0) != Aig_ManConst1(p.pAig)");
    // collect all the nodes to be refined
    k = 0;
    Vec_PtrClear( p.vClassNew );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vClasses1)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vClasses1, i)), 1); i++  {
        if p.pFuncNodeIsConst( pObj ) {
            Vec_PtrWriteEntry( p.vClasses1, k++, pObj );
        } else { 
            Vec_PtrPush( p.vClassNew, pObj ); }
    }
    Vec_PtrShrink( p.vClasses1, k );
    if Vec_PtrSize(p.vClassNew) == 0 {
        return 0; }
/*
    printf( "Refined const-1 class: {" );
    Vec_PtrForEachEntry( Aig_Obj_t *, p->vClassNew, pObj, i )
        printf( " %d", pObj->Id );
    printf( " }\n" );
*/
    if Vec_PtrSize(p.vClassNew) == 1 {
        Fra_ClassObjSetRepr( (Aig_Obj_t *)Vec_PtrEntry(p.vClassNew,0), nil );
        return 1;
    }
    // create a new class composed of these nodes
    ppClass = p.pMemClassesFree;
    p.pMemClassesFree += 2 * Vec_PtrSize(p.vClassNew);
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vClassNew)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vClassNew, i)), 1); i++  {
        ppClass[i] = pObj;
        ppClass[Vec_PtrSize(p.vClassNew)+i] = nil;
        Fra_ClassObjSetRepr( pObj, i? ppClass[0] : nil );
    }
    assert(ppClass[0] != nil, "ppClass[0] != nil");
    Vec_PtrPush( p.vClasses, ppClass );
    // iteratively refine this class
    if fRefineNewClass != 0 {
        nRefis += Fra_RefineClassLastIter( p, p.vClasses );
    } else if pSkipped != nil {
        (*pSkipped)++; }
    return nRefis;
}

/**Function*************************************************************

  Synopsis    [Starts representation of equivalence classes with one class.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_ClassesTest(p *Fra_Cla_t, Id1 int32, Id2 int32) {
    var pClass **Aig_Obj_t;
    // inlining of macro ABC_ALLOC
    p.pMemClasses = ((Aig_Obj_t * *) malloc(sizeof(Aig_Obj_t *) * (size_t)(4)));
    pClass = p.pMemClasses;
    assert(Id1 < Id2, "Id1 < Id2");
    pClass[0] = Aig_ManObj( p.pAig, Id1 );
    pClass[1] = Aig_ManObj( p.pAig, Id2 );
    pClass[2] = nil;
    pClass[3] = nil;
    Fra_ClassObjSetRepr( pClass[1], pClass[0] );
    Vec_PtrPush( p.vClasses, pClass );
}

/**Function*************************************************************

  Synopsis    [Creates latch correspondence classes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_ClassesLatchCorr(p *Fra_Man_t) {
    var pObj *Aig_Obj_t;
    var i, nEntries int32;
    Vec_PtrClear( p.pCla.vClasses1 );
    // inlining of macro Aig_ManForEachLoSeq
    // inlining of macro Vec_PtrForEachEntryStart
    for  i = Aig_ManCiNum(p.pManAig)-Aig_ManRegNum(p.pManAig); (i < Vec_PtrSize(p.pManAig.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pManAig.vCis, i)), 1); i++  {
        Vec_PtrPush( p.pCla.vClasses1, pObj );
        Fra_ClassObjSetRepr( pObj, Aig_ManConst1(p.pManAig) );
    }
    // allocate room for classes
    // inlining of macro ABC_ALLOC
    p.pCla.pMemClasses = ((Aig_Obj_t * *) malloc(sizeof(Aig_Obj_t *) * (size_t)(2*(nEntries + Vec_PtrSize(p.pCla.vClasses1)))));
    p.pCla.pMemClassesFree = p.pCla.pMemClasses + 2*nEntries;
}

/**Function*************************************************************

  Synopsis    [Postprocesses the classes by removing half of the less useful.]

  Description []
               
  SideEffects [] 

  SeeAlso     []

***********************************************************************/
func Fra_ClassesPostprocess(p *Fra_Cla_t) {
    var Ratio int32  = 2;
    var pComb *Fra_Sml_t;
    var pObj *Aig_Obj_t;
    var pRepr *Aig_Obj_t;
    var ppClass **Aig_Obj_t;
    var pWeights *int32;
    var WeightMax int32;
    var i int32;
    var k int32;
    var c int32;
    // perform combinational simulation
    pComb = Fra_SmlSimulateComb( p.pAig, 32, 0 );
    // compute the weight of each node in the classes
    // inlining of macro ABC_ALLOC
    pWeights = ((int *) malloc(sizeof(int) * (size_t)(Aig_ManObjNumMax(p.pAig))));
    memset( pWeights, 0, sizeof(int) * Aig_ManObjNumMax(p.pAig) );
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vObjs, i)), 1); i++  { if (pObj) == nil {} else { 
        pRepr = Fra_ClassObjRepr( pObj );
        if pRepr == nil {
            continue; }
        pWeights[i] = Fra_SmlNodeNotEquWeight( pComb, pRepr.Id, pObj.Id );
        WeightMax = Abc_MaxInt( WeightMax, pWeights[i] );
    }
    }
    Fra_SmlStop( pComb );
    fmt.Printf("Before: Const = %6d. Class = %6d.  ", Vec_PtrSize(p.vClasses1), Vec_PtrSize(p.vClasses));
    // remove nodes from classes whose weight is less than WeightMax/Ratio
    k = 0;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vClasses1)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vClasses1, i)), 1); i++  {
        if pWeights[pObj.Id] >= WeightMax/Ratio {
            Vec_PtrWriteEntry( p.vClasses1, k++, pObj );
        } else {
            Fra_ClassObjSetRepr( pObj, nil ); }
    }
    Vec_PtrShrink( p.vClasses1, k );
    // in each class, compact the nodes
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vClasses)) && (((ppClass) = (Aig_Obj_t **)Vec_PtrEntry(p.vClasses, i)), 1); i++  {
        k = 1;
        for  c = 1; ppClass[c]; c++  {
            if pWeights[ppClass[c].Id] >= WeightMax/Ratio {
                ppClass[k++] = ppClass[c];
            } else {
                Fra_ClassObjSetRepr( ppClass[c], nil ); }
        }
        ppClass[k] = nil;
    }
    // remove classes with only repr
    k = 0;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vClasses)) && (((ppClass) = (Aig_Obj_t **)Vec_PtrEntry(p.vClasses, i)), 1); i++  {
        if ppClass[1] != nil {
            Vec_PtrWriteEntry( p.vClasses, k++, ppClass ); }
    }
    Vec_PtrShrink( p.vClasses, k );
    fmt.Printf("After: Const = %6d. Class = %6d. \n", Vec_PtrSize(p.vClasses1), Vec_PtrSize(p.vClasses));
    pWeights = nil;
}

/**Function*************************************************************

  Synopsis    [Postprocesses the classes by selecting representative lowest in top order.]

  Description []
               
  SideEffects [] 

  SeeAlso     []

***********************************************************************/
func Fra_ClassesSelectRepr(p *Fra_Cla_t) {
    var pClass **Aig_Obj_t;
    var pNodeMin *Aig_Obj_t;
    var i, c, cMinSupp, nSuppSizeMin, nSuppSizeCur int32;
    // reassign representatives in each class
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vClasses)) && (((pClass) = (Aig_Obj_t **)Vec_PtrEntry(p.vClasses, i)), 1); i++  {
        // collect support sizes and find the min-support node
        cMinSupp = -1;
        pNodeMin = nil;
        nSuppSizeMin = ABC_INFINITY;
        for  c = 0; pClass[c]; c++  {
            nSuppSizeCur = Aig_SupportSize( p.pAig, pClass[c] );
//            nSuppSizeCur = 1;
            if nSuppSizeMin > nSuppSizeCur || 
                (nSuppSizeMin == nSuppSizeCur && pNodeMin.Level > pClass[c].Level) {
                nSuppSizeMin = nSuppSizeCur;
                pNodeMin = pClass[c];
                cMinSupp = c; 
            }
        }
        // skip the case when the repr did not change
        if cMinSupp == 0 {
            continue; }
        // make the new node the representative of the class
        pClass[cMinSupp] = pClass[0];
        pClass[0] = pNodeMin;
        // set the representative
        for  c = 0; pClass[c]; c++  {
            Fra_ClassObjSetRepr( pClass[c], c? pClass[0] : nil );
        }
    }
}



func Fra_ObjEqu(ppEquivs **Aig_Obj_t, pObj *Aig_Obj_t) *Aig_Obj_t { return ppEquivs[pObj.Id];  }
func Fra_ObjSetEqu(ppEquivs **Aig_Obj_t, pObj *Aig_Obj_t, pNode *Aig_Obj_t) { ppEquivs[pObj.Id] = pNode; }

func Fra_ObjChild0Equ(ppEquivs **Aig_Obj_t, pObj *Aig_Obj_t) *Aig_Obj_t { return Aig_NotCond(Fra_ObjEqu(ppEquivs,Aig_ObjFanin0(pObj)), Aig_ObjFaninC0(pObj));  }
func Fra_ObjChild1Equ(ppEquivs **Aig_Obj_t, pObj *Aig_Obj_t) *Aig_Obj_t { return Aig_NotCond(Fra_ObjEqu(ppEquivs,Aig_ObjFanin1(pObj)), Aig_ObjFaninC1(pObj));  }

/**Function*************************************************************

  Synopsis    [Add the node and its constraints to the new AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_ClassesDeriveNode(pManFraig *Aig_Man_t, pObj *Aig_Obj_t, ppEquivs **Aig_Obj_t) {
    var pObjNew, pObjRepr, pObjReprNew, pMiter *Aig_Obj_t;//, * pObjNew2;
    // skip nodes without representative
    if (pObjRepr = Fra_ClassObjRepr(pObj)) == nil {
        return; }
    assert(pObjRepr.Id < pObj.Id, "pObjRepr.Id < pObj.Id");
    // get the new node 
    pObjNew = Fra_ObjEqu( ppEquivs, pObj );
    // get the new node of the representative
    pObjReprNew = Fra_ObjEqu( ppEquivs, pObjRepr );
    // if this is the same node, no need to add constraints
    if Aig_Regular(pObjNew) == Aig_Regular(pObjReprNew) {
        return; }
    // these are different nodes - perform speculative reduction
//    pObjNew2 = Aig_NotCond( pObjReprNew, pObj->fPhase ^ pObjRepr->fPhase );
    // set the new node
//    Fra_ObjSetEqu( ppEquivs, pObj, pObjNew2 );
    // add the constraint
    pMiter = Aig_Exor( pManFraig, Aig_Regular(pObjNew), Aig_Regular(pObjReprNew) );
    pMiter = Aig_NotCond( pMiter, Aig_Regular(pMiter).fPhase ^ Aig_IsComplement(pMiter) );
    pMiter = Aig_Not( pMiter );
    Aig_ObjCreateCo( pManFraig, pMiter );
}

/**Function*************************************************************

  Synopsis    [Derives AIG for the partitioned problem.]

  Description []
               
  SideEffects [] 

  SeeAlso     []

***********************************************************************/
func Fra_ClassesDeriveAig(p *Fra_Cla_t, nFramesK int32) *Aig_Man_t {
    var pManFraig *Aig_Man_t;
    var pObj, pObjNew *Aig_Obj_t;
    var pLatches, ppEquivs **Aig_Obj_t;
    var i int32;
    var k int32;
    var f int32;
    var nFramesAll int32 = nFramesK + 1;
    assert(Aig_ManRegNum(p.pAig) > 0, "Aig_ManRegNum(p.pAig) > 0");
    assert(Aig_ManRegNum(p.pAig) < Aig_ManCiNum(p.pAig), "Aig_ManRegNum(p.pAig) < Aig_ManCiNum(p.pAig)");
    assert(nFramesK > 0, "nFramesK > 0");
    // start the fraig package
    pManFraig = Aig_ManStart( Aig_ManObjNumMax(p.pAig) * nFramesAll );
    pManFraig.pName = Abc_UtilStrsav( p.pAig.pName );
    pManFraig.pSpec = Abc_UtilStrsav( p.pAig.pSpec );
    // allocate place for the node mapping
    // inlining of macro ABC_ALLOC
    ppEquivs = ((Aig_Obj_t * *) malloc(sizeof(Aig_Obj_t *) * (size_t)(Aig_ManObjNumMax(p.pAig))));
    Fra_ObjSetEqu( ppEquivs, Aig_ManConst1(p.pAig), Aig_ManConst1(pManFraig) );
    // create latches for the first frame
    // inlining of macro Aig_ManForEachLoSeq
    // inlining of macro Vec_PtrForEachEntryStart
    for  i = Aig_ManCiNum(p.pAig)-Aig_ManRegNum(p.pAig); (i < Vec_PtrSize(p.pAig.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCis, i)), 1); i++  {
        Fra_ObjSetEqu( ppEquivs, pObj, Aig_ObjCreateCi(pManFraig) );
    }
    // add timeframes
    // inlining of macro ABC_ALLOC
    pLatches = ((Aig_Obj_t * *) malloc(sizeof(Aig_Obj_t *) * (size_t)(Aig_ManRegNum(p.pAig))));
    for  f = 0; f < nFramesAll; f++  {
        // create PIs for this frame
        // inlining of macro Aig_ManForEachPiSeq
        // inlining of macro Vec_PtrForEachEntryStop
        for  i = 0; (i < Aig_ManCiNum(p.pAig)-Aig_ManRegNum(p.pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCis, i)), 1); i++  {
            Fra_ObjSetEqu( ppEquivs, pObj, Aig_ObjCreateCi(pManFraig) );
        }
        // set the constraints on the latch outputs
        // inlining of macro Aig_ManForEachLoSeq
        // inlining of macro Vec_PtrForEachEntryStart
        for  i = Aig_ManCiNum(p.pAig)-Aig_ManRegNum(p.pAig); (i < Vec_PtrSize(p.pAig.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCis, i)), 1); i++  {
            Fra_ClassesDeriveNode( pManFraig, pObj, ppEquivs );
        }
        // add internal nodes of this frame
        // inlining of macro Aig_ManForEachNode
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p.pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
            pObjNew = Aig_And( pManFraig, Fra_ObjChild0Equ(ppEquivs, pObj), Fra_ObjChild1Equ(ppEquivs, pObj) );
            Fra_ObjSetEqu( ppEquivs, pObj, pObjNew );
            Fra_ClassesDeriveNode( pManFraig, pObj, ppEquivs );
        }
        }
        if f == nFramesAll - 1 {
            break; }
        if f == nFramesAll - 2 {
            pManFraig.nAsserts = Aig_ManCoNum(pManFraig); }
        // save the latch input values
        k = 0;
        // inlining of macro Aig_ManForEachLiSeq
        // inlining of macro Vec_PtrForEachEntryStart
        for  i = Aig_ManCoNum(p.pAig)-Aig_ManRegNum(p.pAig); (i < Vec_PtrSize(p.pAig.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCos, i)), 1); i++  {
            pLatches[k++] = Fra_ObjChild0Equ( ppEquivs, pObj );
        }
        // insert them to the latch output values
        k = 0;
        // inlining of macro Aig_ManForEachLoSeq
        // inlining of macro Vec_PtrForEachEntryStart
        for  i = Aig_ManCiNum(p.pAig)-Aig_ManRegNum(p.pAig); (i < Vec_PtrSize(p.pAig.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCis, i)), 1); i++  {
            Fra_ObjSetEqu( ppEquivs, pObj, pLatches[k++] );
        }
    }
    pLatches = nil;
    ppEquivs = nil;
    // mark the asserts
    assert(Aig_ManCoNum(pManFraig) % nFramesAll == 0, "Aig_ManCoNum(pManFraig) % nFramesAll == 0");
fmt.Printf("Assert miters = %6d. Output miters = %6d.\n", pManFraig.nAsserts, Aig_ManCoNum(pManFraig) - pManFraig.nAsserts);
    // remove dangling nodes
    Aig_ManCleanup( pManFraig );
    return pManFraig;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

