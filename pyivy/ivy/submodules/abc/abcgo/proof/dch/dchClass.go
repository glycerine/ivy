package dch

/**CFile****************************************************************

  FileName    [dchClass.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Choice computation for tech-mapping.]

  Synopsis    [Representation of candidate equivalence classes.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 29, 2008.]

  Revision    [$Id: dchClass.c,v 1.00 2008/07/29 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "dchInt.h"

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
type Dch_Cla_t_ struct {
    // class information
    pAig       *Aig_Man_t;             // original AIG manager
    pId2Class     ***Aig_Obj_t;        // non-const classes by ID of repr node
    pClassSizes             *int32;      // sizes of each equivalence class
    // statistics
    nClasses              int32;         // the total number of non-const classes
    nCands1              int32;          // the total number of const candidates
    nLits              int32;            // the number of literals in all classes
    // memory
    pMemClasses      **Aig_Obj_t;      // memory allocated for equivalence classes
    pMemClassesFree      **Aig_Obj_t;  // memory allocated for equivalence classes to be used
    // temporary data
    vClassOld       *Vec_Ptr_t;        // old equivalence class after splitting
    vClassNew       *Vec_Ptr_t;        // new equivalence class(es) after splitting
    // procedures used for class refinement
    pManData            any;
    unsigned (*pFuncNodeHash) (void *,Aig_Obj_t *);              // returns hash key of the node
    int (*pFuncNodeIsConst)   (void *,Aig_Obj_t *);              // returns 1 if the node is a constant
    int (*pFuncNodesAreEqual) (void *,Aig_Obj_t *, Aig_Obj_t *); // returns 1 if nodes are equal up to a complement
}

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

func Dch_ObjNext(ppNexts **Aig_Obj_t, pObj *Aig_Obj_t) *Aig_Obj_t { return ppNexts[pObj.Id];  }
func Dch_ObjSetNext(ppNexts **Aig_Obj_t, pObj *Aig_Obj_t, pNext *Aig_Obj_t) { ppNexts[pObj.Id] = pNext; }

// iterator through the equivalence classes
// Original macro:
// #define Dch_ManForEachClass( p, ppClass, i )                 \
//     for ( i = 0; i < Aig_ManObjNumMax(p->pAig); i++ )        \
//         if ( ((ppClass) = p->pId2Class[i]) == NULL ) {} else

// iterator through the nodes in one class
// Original macro:
// #define Dch_ClassForEachNode( p, pRepr, pNode, i )           \
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
func Dch_ObjAddClass(p *Dch_Cla_t, pRepr *Aig_Obj_t, pClass **Aig_Obj_t, nSize int32) {
    assert(p.pId2Class[pRepr.Id] == nil, "p.pId2Class[pRepr.Id] == nil");
    p.pId2Class[pRepr.Id] = pClass; 
    assert(p.pClassSizes[pRepr.Id] == 0, "p.pClassSizes[pRepr.Id] == 0");
    assert(nSize > 1, "nSize > 1");
    p.pClassSizes[pRepr.Id] = nSize;
    p.nClasses++;
    p.nLits += nSize - 1;
}

/**Function*************************************************************

  Synopsis    [Removes one equivalence class.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dch_ObjRemoveClass(p *Dch_Cla_t, pRepr *Aig_Obj_t) **Aig_Obj_t {
    var pClass **Aig_Obj_t  = p.pId2Class[pRepr.Id];
    var nSize int32;
    assert(pClass != nil, "pClass != nil");
    p.pId2Class[pRepr.Id] = nil; 
    nSize = p.pClassSizes[pRepr.Id];
    assert(nSize > 1, "nSize > 1");
    p.nClasses--;
    p.nLits -= nSize - 1;
    p.pClassSizes[pRepr.Id] = 0;
    return pClass;
}

/**Function*************************************************************

  Synopsis    [Starts representation of equivalence classes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dch_ClassesStart(pAig *Aig_Man_t) *Dch_Cla_t {
    var p *Dch_Cla_t;
    // inlining of macro ABC_ALLOC
    p = ((Dch_Cla_t *) malloc(sizeof(Dch_Cla_t) * (size_t)(1)));
    memset( p, 0, sizeof(Dch_Cla_t) );
    p.pAig         = pAig;
    // inlining of macro ABC_CALLOC
    p.pId2Class    = ((Aig_Obj_t ** *) calloc((size_t)(Aig_ManObjNumMax(pAig)), sizeof(Aig_Obj_t **)));
    // inlining of macro ABC_CALLOC
    p.pClassSizes  = ((int *) calloc((size_t)(Aig_ManObjNumMax(pAig)), sizeof(int)));
    p.vClassOld    = Vec_PtrAlloc( 100 );
    p.vClassNew    = Vec_PtrAlloc( 100 );
    assert(pAig.pReprs == nil, "pAig.pReprs == nil");
    Aig_ManReprStart( pAig, Aig_ManObjNumMax(pAig) );
    return p;
}

/**Function*************************************************************

  Synopsis    [Starts representation of equivalence classes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
void Dch_ClassesSetData( Dch_Cla_t * p, void * pManData, 
    unsigned (*pFuncNodeHash)(void *,Aig_Obj_t *),               // returns hash key of the node
    int (*pFuncNodeIsConst)(void *,Aig_Obj_t *),                 // returns 1 if the node is a constant
    int (*pFuncNodesAreEqual)(void *,Aig_Obj_t *, Aig_Obj_t *) ) // returns 1 if nodes are equal up to a complement
{
    p.pManData           = pManData;
    p.pFuncNodeHash      = pFuncNodeHash;
    p.pFuncNodeIsConst   = pFuncNodeIsConst;
    p.pFuncNodesAreEqual = pFuncNodesAreEqual;
}

/**Function*************************************************************

  Synopsis    [Stop representation of equivalence classes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dch_ClassesStop(p *Dch_Cla_t) {
    if p.vClassNew {    Vec_PtrFree( p.vClassNew ); }
    if p.vClassOld {    Vec_PtrFree( p.vClassOld ); }
    p.pId2Class = nil;
    p.pClassSizes = nil;
    p.pMemClasses = nil;
    p = nil;
}

/**Function*************************************************************

  Synopsis    [Stop representation of equivalence classes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dch_ClassesLitNum(p *Dch_Cla_t) int32 {
    return p.nLits;
}

/**Function*************************************************************

  Synopsis    [Stop representation of equivalence classes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dch_ClassesReadClass(p *Dch_Cla_t, pRepr *Aig_Obj_t, pnSize *int32) **Aig_Obj_t {
    assert(p.pId2Class[pRepr.Id] != nil, "p.pId2Class[pRepr.Id] != nil");
    assert(p.pClassSizes[pRepr.Id] > 1, "p.pClassSizes[pRepr.Id] > 1");
    *pnSize = p.pClassSizes[pRepr.Id];
    return p.pId2Class[pRepr.Id];
}

/**Function*************************************************************

  Synopsis    [Checks candidate equivalence classes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dch_ClassesCheck(p *Dch_Cla_t) {
    var pObj *Aig_Obj_t;
    var pPrev *Aig_Obj_t;
    var ppClass **Aig_Obj_t;
    var i, k, nLits, nClasses, nCands1 int32;
    nClasses = nLits = 0;
    // inlining of macro Dch_ManForEachClass
    for  k = 0; k < Aig_ManObjNumMax(p.pAig); k++  {                if ((ppClass) = p.pId2Class[k]) == nil {} else {
        pPrev = nil;
        // inlining of macro Dch_ClassForEachNode
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
        nCands1 += Dch_ObjIsConst1Cand( p.pAig, pObj ); }
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
func Dch_ClassesPrintOne(p *Dch_Cla_t, pRepr *Aig_Obj_t) {
    var pObj *Aig_Obj_t;
    var i int32;
    Abc_Print( 1, "{ " );
    // inlining of macro Dch_ClassForEachNode
    for  i = 0; i < p.pClassSizes[pRepr.Id]; i++  {                if ((pObj) = p.pId2Class[pRepr.Id][i]) == nil {} else {
        Abc_Print( 1, "%d(%d,%d) ", pObj.Id, pObj.Level, Aig_SupportSize(p.pAig,pObj) ); }
    }
    Abc_Print( 1, "}\n" );
}

/**Function*************************************************************

  Synopsis    [Prints simulation classes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dch_ClassesPrint(p *Dch_Cla_t, fVeryVerbose int32) {
    var ppClass **Aig_Obj_t;
    var pObj *Aig_Obj_t;
    var i int32;
    Abc_Print( 1, "Equivalence classes: Const1 = %5d. Class = %5d. Lit = %5d.\n", 
        p.nCands1, p.nClasses, p.nLits );
    if !(fVeryVerbose != 0) {
        return; }
    Abc_Print( 1, "Constants { " );
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vObjs, i)), 1); i++  { if (pObj) == nil {} else
        if Dch_ObjIsConst1Cand( p.pAig, pObj ) {
            Abc_Print( 1, "%d(%d,%d) ", pObj.Id, pObj.Level, Aig_SupportSize(p.pAig,pObj) ); }
    }
    Abc_Print( 1, "}\n" );
    // inlining of macro Dch_ManForEachClass
    for  i = 0; i < Aig_ManObjNumMax(p.pAig); i++  {                if ((ppClass) = p.pId2Class[i]) == nil {} else {
        Abc_Print( 1, "%3d (%3d) : ", i, p.pClassSizes[i] );
        Dch_ClassesPrintOne( p, ppClass[0] );
    }
    }
    Abc_Print( 1, "\n" );
}

/**Function*************************************************************

  Synopsis    [Creates initial simulation classes.]

  Description [Assumes that simulation info is assigned.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dch_ClassesPrepare(p *Dch_Cla_t, fLatchCorr int32, nMaxLevs int32) {
    var ppTable, ppNexts, ppClassNew **Aig_Obj_t;
    var pObj, pTemp, pRepr *Aig_Obj_t;
    var i, k, nTableSize, nNodes, iEntry, nEntries, nEntries2 int32;

    // allocate the hash table hashing simulation info into nodes
    nTableSize = Abc_PrimeCudd( Aig_ManObjNumMax(p.pAig)/4 );
    // inlining of macro ABC_CALLOC
    ppTable = ((Aig_Obj_t * *) calloc((size_t)(nTableSize), sizeof(Aig_Obj_t *))); 
    // inlining of macro ABC_CALLOC
    ppNexts = ((Aig_Obj_t * *) calloc((size_t)(Aig_ManObjNumMax(p.pAig)), sizeof(Aig_Obj_t *))); 

    // add all the nodes to the hash table
    nEntries = 0;
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
            if (nMaxLevs != 0) && int32(pObj.Level) >= nMaxLevs {
                continue; }
        }
        // check if the node belongs to the class of constant 1
        if p.pFuncNodeIsConst( p.pManData, pObj ) {
            Dch_ObjSetConst1Cand( p.pAig, pObj );
            p.nCands1++;
            continue;
        }
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
            if Dch_ObjNext( ppNexts, pRepr ) == nil { // this will be the second entry
                p.pClassSizes[pRepr.Id]++;
                nEntries++;
            }
            // add the entry to the list
            Dch_ObjSetNext( ppNexts, pObj, Dch_ObjNext( ppNexts, pRepr ) );
            Dch_ObjSetNext( ppNexts, pRepr, pObj );
            p.pClassSizes[pRepr.Id]++;
            nEntries++;
        }
    }
    }

    // allocate room for classes
    // inlining of macro ABC_ALLOC
    p.pMemClasses = ((Aig_Obj_t * *) malloc(sizeof(Aig_Obj_t *) * (size_t)(nEntries + p.nCands1)));
    p.pMemClassesFree = p.pMemClasses + nEntries;
 
    // copy the entries into storage in the topological order
    nEntries2 = 0;
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        if !Aig_ObjIsNode(pObj) && !Aig_ObjIsCi(pObj) {
            continue; }
        nNodes = p.pClassSizes[pObj.Id];
        // skip the nodes that are not representatives of non-trivial classes
        if nNodes == 0 {
            continue; }
        assert(nNodes > 1, "nNodes > 1");
        // add the nodes to the class in the topological order
        ppClassNew = p.pMemClasses + nEntries2;
        ppClassNew[0] = pObj;
        for  pTemp = Dch_ObjNext(ppNexts, pObj), k = 1; pTemp; 
              pTemp = Dch_ObjNext(ppNexts, pTemp), k++  {
            ppClassNew[nNodes-k] = pTemp;
        }
        // add the class of nodes
        p.pClassSizes[pObj.Id] = 0;
        Dch_ObjAddClass( p, pObj, ppClassNew, nNodes );
        // increment the number of entries
        nEntries2 += nNodes;
    }
    }
    assert(nEntries == nEntries2, "nEntries == nEntries2");
    ppTable = nil;
    ppNexts = nil;
    // now it is time to refine the classes
    Dch_ClassesRefine( p );
    Dch_ClassesCheck( p );
}

/**Function*************************************************************

  Synopsis    [Iteratively refines the classes after simulation.]

  Description [Returns the number of refinements performed.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dch_ClassesRefineOneClass(p *Dch_Cla_t, pReprOld *Aig_Obj_t, fRecursive int32) int32 {
    var pClassOld, pClassNew **Aig_Obj_t;
    var pObj, pReprNew *Aig_Obj_t;
    var i int32;

    // split the class
    Vec_PtrClear( p.vClassOld );
    Vec_PtrClear( p.vClassNew );
    // inlining of macro Dch_ClassForEachNode
    for  i = 0; i < p.pClassSizes[pReprOld.Id]; i++  {                if ((pObj) = p.pId2Class[pReprOld.Id][i]) == nil {} else
        if p.pFuncNodesAreEqual(p.pManData, pReprOld, pObj) {
            Vec_PtrPush( p.vClassOld, pObj );
        } else {
            Vec_PtrPush( p.vClassNew, pObj ); }
    }
    // check if splitting happened
    if Vec_PtrSize(p.vClassNew) == 0 {
        return 0; }

    // get the new representative
    pReprNew = (Aig_Obj_t *)Vec_PtrEntry( p.vClassNew, 0 );
    assert(Vec_PtrSize(p.vClassOld) > 0, "Vec_PtrSize(p.vClassOld) > 0");
    assert(Vec_PtrSize(p.vClassNew) > 0, "Vec_PtrSize(p.vClassNew) > 0");

    // create old class
    pClassOld = Dch_ObjRemoveClass( p, pReprOld );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vClassOld)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vClassOld, i)), 1); i++  {
        pClassOld[i] = pObj;
        Aig_ObjSetRepr( p.pAig, pObj, i? pReprOld : nil );
    }
    // create new class
    pClassNew = pClassOld + i;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vClassNew)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vClassNew, i)), 1); i++  {
        pClassNew[i] = pObj;
        Aig_ObjSetRepr( p.pAig, pObj, i? pReprNew : nil );
    }

    // put classes back
    if Vec_PtrSize(p.vClassOld) > 1 {
        Dch_ObjAddClass( p, pReprOld, pClassOld, Vec_PtrSize(p.vClassOld) ); }
    if Vec_PtrSize(p.vClassNew) > 1 {
        Dch_ObjAddClass( p, pReprNew, pClassNew, Vec_PtrSize(p.vClassNew) ); }

    // check if the class should be recursively refined
    if (fRecursive != 0) && Vec_PtrSize(p.vClassNew) > 1 {
        return 1 + Dch_ClassesRefineOneClass( p, pReprNew, 1 ); }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Refines the classes after simulation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dch_ClassesRefine(p *Dch_Cla_t) int32 {
    var ppClass **Aig_Obj_t;
    var i, nRefis int32;
    // inlining of macro Dch_ManForEachClass
    for  i = 0; i < Aig_ManObjNumMax(p.pAig); i++  {                if ((ppClass) = p.pId2Class[i]) == nil {} else {
        nRefis += Dch_ClassesRefineOneClass( p, ppClass[0], 0 ); }
    }
    return nRefis;
}


/**Function*************************************************************

  Synopsis    [Returns equivalence class of the given node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dch_ClassesCollectOneClass(p *Dch_Cla_t, pRepr *Aig_Obj_t, vRoots *Vec_Ptr_t) {
    var pObj *Aig_Obj_t;
    var i int32;
    Vec_PtrClear( vRoots );
    // inlining of macro Dch_ClassForEachNode
    for  i = 0; i < p.pClassSizes[pRepr.Id]; i++  {                if ((pObj) = p.pId2Class[pRepr.Id][i]) == nil {} else {
        Vec_PtrPush( vRoots, pObj ); }
    }
    assert(Vec_PtrSize(vRoots) > 1, "Vec_PtrSize(vRoots) > 1"); 
}

/**Function*************************************************************

  Synopsis    [Returns equivalence class of the given node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dch_ClassesCollectConst1Group(p *Dch_Cla_t, pObj *Aig_Obj_t, nNodes int32, vRoots *Vec_Ptr_t) {
    var i, Limit int32;
    Vec_PtrClear( vRoots );
    Limit = Abc_MinInt( pObj.Id + nNodes, Aig_ManObjNumMax(p.pAig) );
    for  i = pObj.Id; i < Limit; i++  {
        pObj = Aig_ManObj( p.pAig, i );
        if (pObj != nil) && Dch_ObjIsConst1Cand( p.pAig, pObj ) {
            Vec_PtrPush( vRoots, pObj ); }
    }
}

/**Function*************************************************************

  Synopsis    [Refine the group of constant 1 nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dch_ClassesRefineConst1Group(p *Dch_Cla_t, vRoots *Vec_Ptr_t, fRecursive int32) int32 {
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
    Dch_ObjAddClass( p, pReprNew, ppClassNew, Vec_PtrSize(p.vClassNew) );
    // refine them recursively
    if fRecursive != 0 {
        return 1 + Dch_ClassesRefineOneClass( p, pReprNew, 1 ); }
    return 1;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

