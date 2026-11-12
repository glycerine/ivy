package gia

import (
	"fmt"
	"time"
)

/**CFile****************************************************************

  FileName    [gia.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    []

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: gia.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// combinational simulation manager
   
type Hcd_Man_t struct {
    // parameters
    pGia       *Gia_Man_t;           // the AIG to be used for simulation
    nBTLimit              int32;       // internal backtrack limit 
    fVerbose              int32;       // internal verbose flag 
    // internal variables
    pSimInfo        *uint32;       // simulation info for each object
    vSimInfo       *Vec_Ptr_t;       // pointers to the CI simulation info
    vSimPres       *Vec_Ptr_t;       // pointers to the presense of simulation info
    // temporaries
    vClassOld       *Vec_Int_t;      // old class numbers
    vClassNew       *Vec_Int_t;      // new class numbers
    vClassTemp       *Vec_Int_t;     // temporary storage
    vRefinedC       *Vec_Int_t;      // refined const reprs
}

func Hcd_ObjSim(p *Hcd_Man_t, Id int32) uint32 { return p.pSimInfo[Id];      }
func Hcd_ObjSimP(p *Hcd_Man_t, Id int32) *uint32 { return p.pSimInfo + Id;     }
func Hcd_ObjSetSim(p *Hcd_Man_t, Id int32, n uint32) uint32 { return p.pSimInfo[Id] = n;  }

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Starts the fraiging manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManEquivStart(pGia *Gia_Man_t, nBTLimit int32, fVerbose int32) *Hcd_Man_t {
    
    var pObj *Gia_Obj_t;
    var i int32;
    // inlining of macro ABC_CALLOC
    p := &Hcd_Man_t{};
    p.pGia       = pGia;
    p.nBTLimit   = nBTLimit;
    p.fVerbose   = fVerbose;
    // inlining of macro ABC_ALLOC
    p.pSimInfo   = ((unsigned *) malloc(sizeof(unsigned) * (size_t)(Gia_ManObjNum(pGia))));
    p.vClassOld  = Vec_IntAlloc( 100 );
    p.vClassNew  = Vec_IntAlloc( 100 );
    p.vClassTemp = Vec_IntAlloc( 100 );
    p.vRefinedC  = Vec_IntAlloc( 100 );
    // collect simulation info
    p.vSimInfo = Vec_PtrAlloc( 1000 );
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(pGia.vCis)) && ((pObj) = Gia_ManCi(pGia, i)); i++  {
        Vec_PtrPush( p.vSimInfo, Hcd_ObjSimP(p, Gia_ObjId(pGia,pObj)) );
    }
    p.vSimPres = Vec_PtrAllocSimInfo( Gia_ManCiNum(pGia), 1 );
    return p;
}

/**Function*************************************************************

  Synopsis    [Starts the fraiging manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManEquivStop(p *Hcd_Man_t) {
    Vec_PtrFree( p.vSimInfo );
    Vec_PtrFree( p.vSimPres );
    Vec_IntFree( p.vClassOld );
    Vec_IntFree( p.vClassNew );
    Vec_IntFree( p.vClassTemp );
    Vec_IntFree( p.vRefinedC );
    // inlining of macro ABC_FREE
    ((p.pSimInfo) ? (free((char *) (p.pSimInfo)), (p.pSimInfo) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}


/**Function*************************************************************

  Synopsis    [Compared two simulation infos.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hcd_ManCompareEqual(s0 uint32, s1 uint32) int32 {
    if (s0 & 1) == (s1 & 1) {
        return s0 == s1;
    } else {
        return s0 ==~s1; }
}

/**Function*************************************************************

  Synopsis    [Compares one simulation info.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hcd_ManCompareConst(s uint32) int32 {
    if s & 1 {
        return s ==~0;
    } else {
        return s == 0; }
}

/**Function*************************************************************

  Synopsis    [Creates one equivalence class.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hcd_ManClassCreate(pGia *Gia_Man_t, vClass *Vec_Int_t) {
    var Repr int32 = -1;
    var EntPrev int32 = -1;
    var Ent int32;
    var i int32;
    assert(Vec_IntSize(vClass) > 0, "Vec_IntSize(vClass) > 0");
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vClass)) && (((Ent) = Vec_IntEntry(vClass, i)), 1); i++  {
        if i == 0 {
            Repr = Ent;
            Gia_ObjSetRepr( pGia, Ent, -1 );
            EntPrev = Ent;
        
        } else {
            Gia_ObjSetRepr( pGia, Ent, Repr );
            Gia_ObjSetNext( pGia, EntPrev, Ent );
            EntPrev = Ent;
        }
    }
    Gia_ObjSetNext( pGia, EntPrev, 0 );
}

/**Function*************************************************************

  Synopsis    [Refines one equivalence class.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hcd_ManClassClassRemoveOne(p *Hcd_Man_t, i int32) int32 {
    var iRepr, Ent int32;
    if Gia_ObjIsConst(p.pGia, i) {
        Gia_ObjSetRepr( p.pGia, i, GIA_VOID );
        return 1;
    }
    if !Gia_ObjIsClass(p.pGia, i) {
        return 0; }
    assert(Gia_ObjIsClass(p.pGia, i), "Gia_ObjIsClass(p.pGia, i)");
    iRepr = Gia_ObjRepr( p.pGia, i );
    if iRepr == GIA_VOID {
        iRepr = i; }
    // collect nodes
    Vec_IntClear( p.vClassOld );
    Vec_IntClear( p.vClassNew );
    // inlining of macro Gia_ClassForEachObj
    for  assert(Gia_ObjIsHead(p.pGia, iRepr) && iRepr, "Gia_ObjIsHead(p.pGia, iRepr) && iRepr"), Ent = iRepr; Ent > 0; Ent = Gia_ObjNext(p.pGia, Ent)  {
        if Ent == i {
            Vec_IntPush( p.vClassNew, Ent );
        } else {
            Vec_IntPush( p.vClassOld, Ent ); }
    }
    assert(Vec_IntSize( p.vClassNew ) == 1, "Vec_IntSize( p.vClassNew ) == 1");
    Hcd_ManClassCreate( p.pGia, p.vClassOld );
    Hcd_ManClassCreate( p.pGia, p.vClassNew );
    assert(!Gia_ObjIsClass(p.pGia, i), "!Gia_ObjIsClass(p.pGia, i)");
    return 1;
}

/**Function*************************************************************

  Synopsis    [Refines one equivalence class.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hcd_ManClassRefineOne(p *Hcd_Man_t, i int32) int32 {
    var Sim0, Sim1 uint32;
    var Ent int32;
    Vec_IntClear( p.vClassOld );
    Vec_IntClear( p.vClassNew );
    Vec_IntPush( p.vClassOld, i );
    Sim0 = Hcd_ObjSim(p, i);
    // inlining of macro Gia_ClassForEachObj1
    for  assert(Gia_ObjIsHead(p.pGia, i), "Gia_ObjIsHead(p.pGia, i)"), Ent = Gia_ObjNext(p.pGia, i); Ent > 0; Ent = Gia_ObjNext(p.pGia, Ent)  {
        Sim1 = Hcd_ObjSim(p, Ent);
        if Hcd_ManCompareEqual( Sim0, Sim1 ) != 0 {
            Vec_IntPush( p.vClassOld, Ent );
        } else {
            Vec_IntPush( p.vClassNew, Ent ); }
    }
    if Vec_IntSize( p.vClassNew ) == 0 {
        return 0; }
    Hcd_ManClassCreate( p.pGia, p.vClassOld );
    Hcd_ManClassCreate( p.pGia, p.vClassNew );
    if Vec_IntSize(p.vClassNew) > 1 {
        return 1 + Hcd_ManClassRefineOne( p, Vec_IntEntry(p.vClassNew,0) ); }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Computes hash key of the simulation info.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Hcd_ManHashKey_s_Primes [16]int32 = [16]int32{ 
        1291, 1699, 1999, 2357, 2953, 3313, 3907, 4177, 
        4831, 5147, 5647, 6343, 6899, 7103, 7873, 8147 };
func Hcd_ManHashKey(pSim *uint32, nWords int32, nTableSize int32) int32 {
    
    var uHash uint32;
    var i int32;
    if pSim[0] & 1 {
        for  i = 0; i < nWords; i++  {
            uHash ^= ~pSim[i] * spStatic_Hcd_ManHashKey_s_Primes[i & 0xf];
        }
    } else {
        for  i = 0; i < nWords; i++  {
            uHash ^=  pSim[i] * spStatic_Hcd_ManHashKey_s_Primes[i & 0xf];
        } }
    return int32((uHash % nTableSize));

}

/**Function*************************************************************

  Synopsis    [Rehashes the refined classes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hcd_ManClassesRehash(p *Hcd_Man_t, vRefined *Vec_Int_t) {
    var pTable *int32;
    var nTableSize int32;
    var Key int32;
    var i int32;
    var k int32;
    nTableSize = Abc_PrimeCudd( 100 + Vec_IntSize(vRefined) / 5 );
    // inlining of macro ABC_CALLOC
    pTable = ((int *) calloc((size_t)(nTableSize), sizeof(int)));
    // inlining of macro Vec_IntForEachEntry
    for  k = 0; (k < Vec_IntSize(vRefined)) && (((i) = Vec_IntEntry(vRefined, k)), 1); k++  {
        assert(!(Hcd_ManCompareConst( Hcd_ObjSim(p, i) ) != 0), "!Hcd_ManCompareConst( Hcd_ObjSim(p, i) )");
        Key = Hcd_ManHashKey( Hcd_ObjSimP(p, i), 1, nTableSize );
        if pTable[Key] == 0 {
            Gia_ObjSetRepr( p.pGia, i, GIA_VOID );
        } else {
            Gia_ObjSetNext( p.pGia, pTable[Key], i );
            Gia_ObjSetRepr( p.pGia, i, Gia_ObjRepr(p.pGia, pTable[Key]) );
            if Gia_ObjRepr(p.pGia, i) == GIA_VOID {
                Gia_ObjSetRepr( p.pGia, i, pTable[Key] ); }
        }
        pTable[Key] = i;
    }
    // inlining of macro ABC_FREE
    ((pTable) ? (free((char *) (pTable)), (pTable) = 0) : 0);
//    Gia_ManEquivPrintClasses( p->pGia, 0, 0.0 );
    // refine classes in the table
    // inlining of macro Vec_IntForEachEntry
    for  k = 0; (k < Vec_IntSize(vRefined)) && (((i) = Vec_IntEntry(vRefined, k)), 1); k++  {
        if Gia_ObjIsHead( p.pGia, i ) {
            Hcd_ManClassRefineOne( p, i ); }
    }
    Gia_ManEquivPrintClasses( p.pGia, 0, 0.0 );
}

/**Function*************************************************************

  Synopsis    [Refines equivalence classes after simulation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hcd_ManClassesRefine(p *Hcd_Man_t) {
    var pObj *Gia_Obj_t;
    var i int32;
    Vec_IntClear( p.vRefinedC );
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.pGia.nObjs) && ((pObj) = Gia_ManObj(p.pGia, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        if Gia_ObjIsTail(p.pGia, i) { // add check for the class level
        
            Hcd_ManClassRefineOne( p, Gia_ObjRepr(p.pGia, i) );
        
        } else if Gia_ObjIsConst(p.pGia, i) {
            if !(Hcd_ManCompareConst( Hcd_ObjSim(p, i) ) != 0) {
                Vec_IntPush( p.vRefinedC, i ); }
        }
    }
    }
    Hcd_ManClassesRehash( p, p.vRefinedC );
}

/**Function*************************************************************

  Synopsis    [Creates equivalence classes for the first time.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hcd_ManClassesCreate(p *Hcd_Man_t) {
    var pObj *Gia_Obj_t;
    var i int32;
    assert(p.pGia.pReprs == nil, "p.pGia.pReprs == nil");
    // inlining of macro ABC_CALLOC
    p.pGia.pReprs = ((Gia_Rpr_t *) calloc((size_t)(Gia_ManObjNum(p.pGia)), sizeof(Gia_Rpr_t)));
    // inlining of macro ABC_CALLOC
    p.pGia.pNexts = ((int *) calloc((size_t)(Gia_ManObjNum(p.pGia)), sizeof(int)));
    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < p.pGia.nObjs) && ((pObj) = Gia_ManObj(p.pGia, i)); i++  {
        Gia_ObjSetRepr( p.pGia, i, Gia_ObjIsAnd(pObj) ? 0 : GIA_VOID );
    }
}

/**Function*************************************************************

  Synopsis    [Initializes simulation info.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hcd_ManSimulationInit(p *Hcd_Man_t) {
    var pObj *Gia_Obj_t;
    var i int32;
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.pGia.vCis)) && ((pObj) = Gia_ManCi(p.pGia, i)); i++  {
        Hcd_ObjSetSim( p, i, (Gia_ManRandom(0) << 1) );
    }
}

/**Function*************************************************************

  Synopsis    [Performs one round of simple combinational simulation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hcd_ManSimulateSimple(p *Hcd_Man_t) {
    var pObj *Gia_Obj_t;
    var Res0, Res1 uint32;
    var i int32;
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.pGia.nObjs) && ((pObj) = Gia_ManObj(p.pGia, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        Res0 = Hcd_ObjSim( p, Gia_ObjFaninId0(pObj, i) );
        Res1 = Hcd_ObjSim( p, Gia_ObjFaninId1(pObj, i) );
        Hcd_ObjSetSim( p, i, (Gia_ObjFaninC0(pObj)? ~Res0: Res0) &
                             (Gia_ObjFaninC1(pObj)? ~Res1: Res1) );
    }
    }
}


/**Function*************************************************************

  Synopsis    [Resimulate and refine one equivalence class.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_Resimulate_rec(p *Hcd_Man_t, iObj int32) uint32 {
    var pObj *Gia_Obj_t;
    var Res0, Res1 uint32;
    if Gia_ObjIsTravIdCurrentId( p.pGia, iObj ) {
        return Hcd_ObjSim( p, iObj ); }
    Gia_ObjSetTravIdCurrentId( p.pGia, iObj );
    pObj = Gia_ManObj(p.pGia, iObj);
    if Gia_ObjIsCi(pObj) {
        return Hcd_ObjSim( p, iObj ); }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    Res0 = Gia_Resimulate_rec( p, Gia_ObjFaninId0(pObj, iObj) );
    Res1 = Gia_Resimulate_rec( p, Gia_ObjFaninId1(pObj, iObj) );
    return Hcd_ObjSetSim( p, iObj, (Gia_ObjFaninC0(pObj)? ~Res0: Res0) &
                                   (Gia_ObjFaninC1(pObj)? ~Res1: Res1) );
}

/**Function*************************************************************

  Synopsis    [Resimulate and refine one equivalence class.]

  Description [Assumes that the counter-example is assigned at the PIs.
  The counter-example should have the first bit set to 0 at each PI.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ResimulateAndRefine(p *Hcd_Man_t, i int32) {
    var RetValue, iObj int32;
    Gia_ManIncrementTravId( p.pGia );
    // inlining of macro Gia_ClassForEachObj
    for  assert(Gia_ObjIsHead(p.pGia, i) && (i != 0), "Gia_ObjIsHead(p.pGia, i) && i"), iObj = i; iObj > 0; iObj = Gia_ObjNext(p.pGia, iObj)  {
        Gia_Resimulate_rec( p, iObj );
    }
    RetValue = Hcd_ManClassRefineOne( p, i );
    if RetValue == 0 {
        fmt.Printf("!!! no refinement !!!\n"); }
//    assert( RetValue );
}

/**Function*************************************************************

  Synopsis    [Returns temporary representative of the node.]

  Description [The temp repr is the first node among the nodes in the class that
  (a) precedes the given node, and (b) whose level is lower than the given node.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ObjTempRepr(p *Gia_Man_t, i int32, Level int32) *Gia_Obj_t {
    var iRepr, iMember int32;
    iRepr = Gia_ObjRepr( p, i ); 
    if !Gia_ObjProved(p, i) {
        return nil; }
    if Gia_ObjFailed(p, i) {
        return nil; }
    if iRepr == GIA_VOID {
        return nil; }
    if iRepr == 0 {
        return Gia_ManConst0( p ); }
//    if ( p->pLevels[iRepr] < Level )
//        return Gia_ManObj( p, iRepr );
    // inlining of macro Gia_ClassForEachObj
    for  assert(Gia_ObjIsHead(p, iRepr) && iRepr, "Gia_ObjIsHead(p, iRepr) && iRepr"), iMember = iRepr; iMember > 0; iMember = Gia_ObjNext(p, iMember)  {
        if Gia_ObjFailed(p, iMember) {
            continue; }
        if iMember >= i {
            return nil; }
        if Gia_ObjLevelId(p, iMember) < Level {
            return Gia_ManObj( p, iMember ); }
    }
    assert(false, "0");
    return nil;
}

/**Function*************************************************************

  Synopsis    [Generates reduced AIG for the given level.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_GenerateReducedLevel(p *Gia_Man_t, Level int32, pvRoots **Vec_Ptr_t) *Gia_Man_t {
    var pNew *Gia_Man_t;
    var pObj, pRepr *Gia_Obj_t;
    var vRoots *Vec_Ptr_t;
    var i int32;
    vRoots = Vec_PtrAlloc( 100 );
    // copy unmarked nodes
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObj.Value = Gia_ManAppendCi(pNew);
    }
    Gia_ManHashAlloc( pNew );
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        if Gia_ObjLevelId(p, i) > Level {
            continue; }
        if Gia_ObjLevelId(p, i) == Level {
            Vec_PtrPush( vRoots, pObj ); }
        if Gia_ObjLevelId(p, i) < Level && (pRepr = Gia_ObjTempRepr(p, i, Level)) {
//            printf( "Substituting %d <--- %d\n", Gia_ObjId(p, pRepr), Gia_ObjId(p, pObj) ); 
            assert(pRepr < pObj, "pRepr < pObj");
            pObj.Value  = Abc_LitNotCond( pRepr.Value, Gia_ObjPhase(pRepr) ^ Gia_ObjPhase(pObj) );
            continue;
        }
        pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
    }
    }
    *pvRoots = vRoots;
    // required by SAT solving
    Gia_ManCreateRefs( pNew );
    Gia_ManFillValue( pNew );
    Gia_ManIncrementTravId( pNew ); // needed for MiniSat to record cexes
//    Gia_ManSetPhase( pNew ); // needed if MiniSat is using polarity -- should not be enabled for TAS because fPhase is used to label
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Collects relevant classes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_CollectRelatedClasses(pGia *Gia_Man_t, vRoots *Vec_Ptr_t) *Vec_Ptr_t {
    var vClasses *Vec_Ptr_t;
    var pRoot, pRepr *Gia_Obj_t;
    var i int32;
    vClasses = Vec_PtrAlloc( 100 );
    Gia_ManConst0( pGia ).fMark0 = 1;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vRoots)) && (((pRoot) = (Gia_Obj_t *)Vec_PtrEntry(vRoots, i)), 1); i++  {
        pRepr = Gia_ObjReprObj( pGia, Gia_ObjId(pGia, pRoot) );
        if pRepr == nil || pRepr.fMark0 {
            continue; }
        pRepr.fMark0 = 1;
        Vec_PtrPush( vClasses, pRepr );
    }
    Gia_ManConst0( pGia ).fMark0 = 0;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vClasses)) && (((pRepr) = (Gia_Obj_t *)Vec_PtrEntry(vClasses, i)), 1); i++  {
        pRepr.fMark0 = 0;
    }
    return vClasses;
}

/**Function*************************************************************

  Synopsis    [Collects class members.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_CollectClassMembers(p *Gia_Man_t, pRepr *Gia_Obj_t, vMembers *Vec_Ptr_t, Level int32) *Gia_Obj_t {
    var pTempRepr *Gia_Obj_t  = nil;
    var iRepr, iMember int32;
    iRepr = Gia_ObjId( p, pRepr );
    Vec_PtrClear( vMembers );
    // inlining of macro Gia_ClassForEachObj
    for  assert(Gia_ObjIsHead(p, iRepr) && iRepr, "Gia_ObjIsHead(p, iRepr) && iRepr"), iMember = iRepr; iMember > 0; iMember = Gia_ObjNext(p, iMember)  {
        if Gia_ObjLevelId(p, iMember) == Level {
            Vec_PtrPush( vMembers, Gia_ManObj( p, iMember ) ); }
        if pTempRepr == nil && Gia_ObjLevelId(p, iMember) < Level {
            pTempRepr = Gia_ManObj( p, iMember ); }
    }
    return pTempRepr;
}


/**Function*************************************************************

  Synopsis    [Packs patterns into array of simulation info.]

  Description []
               
  SideEffects []

  SeeAlso     []

*************************************`**********************************/
func Gia_GiarfStorePatternTry(vInfo *Vec_Ptr_t, vPres *Vec_Ptr_t, iBit int32, pLits *int32, nLits int32) int32 {
    var pInfo, pPres *uint32;
    var i int32;
    for  i = 0; i < nLits; i++  {
        pInfo = (unsigned *)Vec_PtrEntry(vInfo, Abc_Lit2Var(pLits[i]));
        pPres = (unsigned *)Vec_PtrEntry(vPres, Abc_Lit2Var(pLits[i]));
        if Abc_InfoHasBit( pPres, iBit ) && 
             Abc_InfoHasBit( pInfo, iBit ) == Abc_LitIsCompl(pLits[i]) {
             return 0; }
    }
    for  i = 0; i < nLits; i++  {
        pInfo = (unsigned *)Vec_PtrEntry(vInfo, Abc_Lit2Var(pLits[i]));
        pPres = (unsigned *)Vec_PtrEntry(vPres, Abc_Lit2Var(pLits[i]));
        Abc_InfoSetBit( pPres, iBit );
        if Abc_InfoHasBit( pInfo, iBit ) == Abc_LitIsCompl(pLits[i]) {
            Abc_InfoXorBit( pInfo, iBit ); }
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Procedure to test the new SAT solver.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_GiarfStorePattern(vSimInfo *Vec_Ptr_t, vPres *Vec_Ptr_t, vCex *Vec_Int_t) int32 {
    var k int32;
    for  k = 1; k < 32; k++  {
        if Gia_GiarfStorePatternTry( vSimInfo, vPres, k, (int *)Vec_IntArray(vCex), Vec_IntSize(vCex) ) {
            break; }
    }
    return int32((k < 32));
}

/**Function*************************************************************

  Synopsis    [Inserts pattern into simulation info for the PIs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_GiarfInsertPattern(p *Hcd_Man_t, vCex *Vec_Int_t, k int32) {
    var pObj *Gia_Obj_t;
    var pInfo *uint32;
    var Lit, i int32;
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vCex)) && (((Lit) = Vec_IntEntry(vCex, i)), 1); i++  {
        pObj = Gia_ManCi( p.pGia, Abc_Lit2Var(Lit) );
        pInfo = Hcd_ObjSimP( p, Gia_ObjId( p.pGia, pObj ) );
        if Abc_InfoHasBit( pInfo, k ) == Abc_LitIsCompl(Lit) {
            Abc_InfoXorBit( pInfo, k ); }
    }
}

/**Function*************************************************************

  Synopsis    [Inserts pattern into simulation info for the PIs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_GiarfPrintClasses(pGia *Gia_Man_t) {
    var nFails int32;
    var nProves int32;
    var nTotal int32;
    var nBoth int32;
    var i int32;
    for  i = 0; i < Gia_ManObjNum(pGia); i++  {
        nFails  += Gia_ObjFailed(pGia, i);
        nProves += Gia_ObjProved(pGia, i);
        nTotal  += Gia_ObjReprObj(pGia, i) != nil;
        nBoth   += Gia_ObjFailed(pGia, i) && Gia_ObjProved(pGia, i);
    }
    fmt.Printf("nFails = %7d.  nProves = %7d.  nBoth = %7d.  nTotal = %7d.\n", nFails, nProves, nBoth, nTotal);
}

//ABC_NAMESPACE_IMPL_END

//#include "proof/cec/cecInt.h"

//ABC_NAMESPACE_IMPL_START


/**Function*************************************************************

  Synopsis    [Performs computation of AIGs with choices.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ComputeEquivalencesLevel(p *Hcd_Man_t, pGiaLev *Gia_Man_t, vOldRoots *Vec_Ptr_t, Level int32, fUseMiniSat int32) int32 {
    var fUse2Solver int32;
    var pSat *Cec_ManSat_t;
    var Pars Cec_ParSat_t;
    var pTas *Tas_Man_t;
    var vCex *Vec_Int_t;
    var vClasses, vMembers, vOldRootsNew *Vec_Ptr_t;
    var pRoot, pMember, pMemberPrev, pRepr, pTempRepr *Gia_Obj_t;
    var i, k, nIter, iRoot, iRootNew, iMember, iMemberPrev, status, fOneFailed int32;//, iRepr;//, fTwoMember;
    var nSaved, nRecords, nUndec, nClassRefs, nTsat, nMiniSat int32;
    var clk int64;
    var timeTsat int64;
    var timeMiniSat int64;
    var timeSim int64;
    var timeTotal int64 = time.Now().UnixNano();
    if Vec_PtrSize(vOldRoots) == 0 {
        return 0; }
    // start SAT solvers
    Cec_ManSatSetDefaultParams( &Pars );
    Pars.fPolarFlip = 0;
    Pars.nBTLimit   = p.nBTLimit;
    pSat = Cec_ManSatCreate( pGiaLev, &Pars );
    pTas = Tas_ManAlloc( pGiaLev, p.nBTLimit );
    if fUseMiniSat != 0 {
        vCex = Cec_ManSatReadCex( pSat );
    } else {
        vCex = Tas_ReadModel( pTas ); }
    vMembers = Vec_PtrAlloc( 100 );
    Vec_PtrCleanSimInfo( p.vSimPres, 0, 1 );
    // resolve constants
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vOldRoots)) && (((pRoot) = (Gia_Obj_t *)Vec_PtrEntry(vOldRoots, i)), 1); i++  {
        iRoot = Gia_ObjId( p.pGia, pRoot );
        if !Gia_ObjIsConst( p.pGia, iRoot ) {
            continue; }
        iRootNew = Abc_LitNotCond( pRoot.Value, pRoot.fPhase );
        assert(iRootNew != 1, "iRootNew != 1");
        if fUse2Solver != 0 {
            nTsat++;
           clk = time.Now().UnixNano();
            status = Tas_ManSolve( pTas, Gia_ObjFromLit(pGiaLev, iRootNew), nil );
           timeTsat += time.Now().UnixNano() - clk;
            if status == -1 {
                nMiniSat++;
               clk = time.Now().UnixNano();
                status = Cec_ManSatCheckNode( pSat, Gia_ObjFromLit(pGiaLev, iRootNew) );
               timeMiniSat += time.Now().UnixNano() - clk;
                if status == 0 {
                    Cec_ManSavePattern( pSat, Gia_ObjFromLit(pGiaLev, iRootNew), nil );
                    vCex = Cec_ManSatReadCex( pSat );
                }
            
            } else if status == 0 {
                vCex = Tas_ReadModel( pTas ); }
        
        } else if fUseMiniSat != 0 {
            nMiniSat++;
           clk = time.Now().UnixNano();
            status = Cec_ManSatCheckNode( pSat, Gia_ObjFromLit(pGiaLev, iRootNew) );
           timeMiniSat += time.Now().UnixNano() - clk;
            if status == 0 {
                Cec_ManSavePattern( pSat, Gia_ObjFromLit(pGiaLev, iRootNew), nil ); }
        
        } else {
            nTsat++;
           clk = time.Now().UnixNano();
            status = Tas_ManSolve( pTas, Gia_ObjFromLit(pGiaLev, iRootNew), nil );
           timeTsat += time.Now().UnixNano() - clk;
        }
        if status == -1 { // undec
        
//            Gia_ObjSetFailed( p->pGia, iRoot );
            nUndec++;
//            Hcd_ManClassClassRemoveOne( p, iRoot );
            Gia_ObjSetFailed( p.pGia, iRoot );
        
        } else if status == 1 { // unsat
        
            Gia_ObjSetProved( p.pGia, iRoot );
//            printf( "proved constant %d\n", iRoot );
        
        } else {  // sat
        
//            printf( "Disproved constant %d\n", iRoot );
            Gia_ObjUnsetRepr( p.pGia, iRoot ); // do we need this?
            nRecords++;
            nSaved += Gia_GiarfStorePattern( p.vSimInfo, p.vSimPres, vCex );
        }
    }

    vClasses = Vec_PtrAlloc( 100 );
    vOldRootsNew = Vec_PtrAlloc( 100 );
    for  nIter = 0; Vec_PtrSize(vOldRoots) > 0; nIter++  {
//        printf( "Iter = %d  (Size = %d)\n", nIter, Vec_PtrSize(vOldRoots) );
        // resolve equivalences 
        Vec_PtrClear( vClasses );
        Vec_PtrClear( vOldRootsNew );
        Gia_ManConst0( p.pGia ).fMark0 = 1;
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(vOldRoots)) && (((pRoot) = (Gia_Obj_t *)Vec_PtrEntry(vOldRoots, i)), 1); i++  {
            iRoot = Gia_ObjId( p.pGia, pRoot );
            if Gia_ObjIsHead( p.pGia, iRoot ) {
                pRepr = pRoot;
            } else if Gia_ObjIsClass( p.pGia, iRoot ) {
                pRepr = Gia_ObjReprObj( p.pGia, iRoot );
            } else {
                continue; }
            if pRepr.fMark0 {
                continue; }
            pRepr.fMark0 = 1;
            Vec_PtrPush( vClasses, pRepr );
//            iRepr = Gia_ObjId( p->pGia, pRepr );
//            fTwoMember = Gia_ClassIsPair(p->pGia, iRepr)
            // derive temp repr and members on this level
            pTempRepr = Gia_CollectClassMembers( p.pGia, pRepr, vMembers, Level );
            if pTempRepr != nil {
                Vec_PtrPush( vMembers, pTempRepr ); }
            if Vec_PtrSize(vMembers) < 2 {
                continue; }
            // try proving the members
            fOneFailed = 0;
            pMemberPrev = (Gia_Obj_t *)Vec_PtrEntryLast( vMembers );
            // inlining of macro Vec_PtrForEachEntry
            for  k = 0; (k < Vec_PtrSize(vMembers)) && (((pMember) = (Gia_Obj_t *)Vec_PtrEntry(vMembers, k)), 1); k++  {
                iMemberPrev = Abc_LitNotCond( pMemberPrev.Value,  pMemberPrev.fPhase ); 
                iMember     = Abc_LitNotCond( pMember.Value,     !pMember.fPhase ); 
                assert(iMemberPrev != iMember, "iMemberPrev != iMember");
                if fUse2Solver != 0 {
                    nTsat++;
                   clk = time.Now().UnixNano();
                    status = Tas_ManSolve( pTas, Gia_ObjFromLit(pGiaLev, iMemberPrev), Gia_ObjFromLit(pGiaLev, iMember) );
                   timeTsat += time.Now().UnixNano() - clk;
                    if status == -1 {
                        nMiniSat++;
                       clk = time.Now().UnixNano();
                        status = Cec_ManSatCheckNodeTwo( pSat, Gia_ObjFromLit(pGiaLev, iMemberPrev), Gia_ObjFromLit(pGiaLev, iMember) );
                       timeMiniSat += time.Now().UnixNano() - clk;
                        if status == 0 {
                            Cec_ManSavePattern( pSat, Gia_ObjFromLit(pGiaLev, iMemberPrev), Gia_ObjFromLit(pGiaLev, iMember) );
                            vCex = Cec_ManSatReadCex( pSat );
                        }
                    
                    } else if status == 0 {
                        vCex = Tas_ReadModel( pTas ); }
                
                } else if fUseMiniSat != 0 {
                    nMiniSat++;
                   clk = time.Now().UnixNano();
                    status = Cec_ManSatCheckNodeTwo( pSat, Gia_ObjFromLit(pGiaLev, iMemberPrev), Gia_ObjFromLit(pGiaLev, iMember) );
                   timeMiniSat += time.Now().UnixNano() - clk;
                    if status == 0 {
                        Cec_ManSavePattern( pSat, Gia_ObjFromLit(pGiaLev, iMemberPrev), Gia_ObjFromLit(pGiaLev, iMember) ); }
                
                } else {
                    nTsat++;
                   clk = time.Now().UnixNano();
                    status = Tas_ManSolve( pTas, Gia_ObjFromLit(pGiaLev, iMemberPrev), Gia_ObjFromLit(pGiaLev, iMember) );
                   timeTsat += time.Now().UnixNano() - clk;
                }
                if status == -1 { // undec
                
//                    Gia_ObjSetFailed( p->pGia, iRoot );
                    nUndec++;
                    if Gia_ObjLevel(p.pGia, pMemberPrev) > Gia_ObjLevel(p.pGia, pMember) {
//                        Hcd_ManClassClassRemoveOne( p, Gia_ObjId(p->pGia, pMemberPrev) );
                        Gia_ObjSetFailed( p.pGia, Gia_ObjId(p.pGia, pMemberPrev) );
                        Gia_ObjSetFailed( p.pGia, Gia_ObjId(p.pGia, pMember) );
                    
                    } else {
//                        Hcd_ManClassClassRemoveOne( p, Gia_ObjId(p->pGia, pMember) );
                        Gia_ObjSetFailed( p.pGia, Gia_ObjId(p.pGia, pMemberPrev) );
                        Gia_ObjSetFailed( p.pGia, Gia_ObjId(p.pGia, pMember) );
                    }
                
                } else if status == 1 { // unsat
                
//                    Gia_ObjSetProved( p->pGia, iRoot );
                
                } else {  // sat
                
//                    iRepr = Gia_ObjId( p->pGia, pRepr );
//                    if ( Gia_ClassIsPair(p->pGia, iRepr) )
//                        Gia_ClassUndoPair(p->pGia, iRepr);
//                    else
                    {
                        fOneFailed = 1;
                        nRecords++;
                        nSaved += Gia_GiarfStorePattern( p.vSimInfo, p.vSimPres, vCex );
                        Gia_GiarfInsertPattern( p, vCex, (k % 31) + 1 );
                    }
                }
                pMemberPrev = pMember;
//                if ( fOneFailed )
//                    k += Vec_PtrSize(vMembers) / 4;
            }
            // if fail, quit this class
            if fOneFailed != 0 {
                nClassRefs++;
                // inlining of macro Vec_PtrForEachEntry
                for  k = 0; (k < Vec_PtrSize(vMembers)) && (((pMember) = (Gia_Obj_t *)Vec_PtrEntry(vMembers, k)), 1); k++  {
                    if pMember != pTempRepr && !Gia_ObjFailed(p.pGia, Gia_ObjId(p.pGia, pMember)) {
                        Vec_PtrPush( vOldRootsNew, pMember ); }
                }
                clk = time.Now().UnixNano();
                Gia_ResimulateAndRefine( p, Gia_ObjId(p.pGia, pRepr) );
                timeSim += time.Now().UnixNano() - clk;
            
            } else {
                // inlining of macro Vec_PtrForEachEntry
                for  k = 0; (k < Vec_PtrSize(vMembers)) && (((pMember) = (Gia_Obj_t *)Vec_PtrEntry(vMembers, k)), 1); k++  {
                    Gia_ObjSetProved( p.pGia, Gia_ObjId(p.pGia, pMember) );
                }
/*
//            }
//            else
//            {
                printf( "Proved equivalent: " );
                Vec_PtrForEachEntry( Gia_Obj_t *, vMembers, pMember, k )
                    printf( "%d(L=%d)  ", Gia_ObjId(p->pGia, pMember), p->pGia->pLevels[Gia_ObjId(p->pGia, pMember)] );
                printf( "\n" );
*/
            }

        }
        Vec_PtrClear( vOldRoots );
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(vOldRootsNew)) && (((pMember) = (Gia_Obj_t *)Vec_PtrEntry(vOldRootsNew, i)), 1); i++  {
            Vec_PtrPush( vOldRoots, pMember );
        }
        // clean up
        Gia_ManConst0( p.pGia ).fMark0 = 0;
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(vClasses)) && (((pRepr) = (Gia_Obj_t *)Vec_PtrEntry(vClasses, i)), 1); i++  {
            pRepr.fMark0 = 0;
        }
    }
    Vec_PtrFree( vClasses );
    Vec_PtrFree( vOldRootsNew );
    fmt.Printf("nSaved = %d   nRecords = %d   nUndec = %d   nClassRefs = %d   nMiniSat = %d   nTas = %d\n", nSaved, nRecords, nUndec, nClassRefs, nMiniSat, nTsat);
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("Tas    "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((timeTsat)))/(float64(int64(time.Second))));;
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("MiniSat"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((timeMiniSat)))/(float64(int64(time.Second))));;
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("Sim    "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((timeSim)))/(float64(int64(time.Second))));;
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("Total  "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((time.Now().UnixNano() - timeTotal)))/(float64(int64(time.Second))));;

    // resimulate
//    clk = clock();
    Hcd_ManSimulateSimple( p );
    Hcd_ManClassesRefine( p );
//    ABC_PRT( "Simulate/refine", clock() - clk );

    // verify the results
    Vec_PtrFree( vMembers );
    Tas_ManStop( pTas );
    Cec_ManSatStop( pSat );
    return nIter;
}

/**Function*************************************************************

  Synopsis    [Performs computation of AIGs with choices.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ComputeEquivalences(pGia *Gia_Man_t, nBTLimit int32, fUseMiniSat int32, fVerbose int32) {
    var p *Hcd_Man_t;
    var vRoots *Vec_Ptr_t;
    var pGiaLev *Gia_Man_t;
    var i, Lev, nLevels, nIters int32;
    var clk int64;
    Gia_ManRandom( 1 );
    Gia_ManSetPhase( pGia );
    nLevels = Gia_ManLevelNum( pGia );
    Gia_ManIncrementTravId( pGia );
    // start the manager
    p = Gia_ManEquivStart( pGia, nBTLimit, fVerbose );
    // create trivial classes
    Hcd_ManClassesCreate( p );
    // refine
    for  i = 0; i < 3; i++  {
        clk = time.Now().UnixNano();       
        Hcd_ManSimulationInit( p );
        Hcd_ManSimulateSimple( p );
        // inlining of macro ABC_PRT
        Abc_Print(1, "%s =", ("Sim"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((time.Now().UnixNano() - clk)))/(float64(int64(time.Second))));;
        clk = time.Now().UnixNano();
        Hcd_ManClassesRefine( p );
        // inlining of macro ABC_PRT
        Abc_Print(1, "%s =", ("Ref"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((time.Now().UnixNano() - clk)))/(float64(int64(time.Second))));;
    }
    // process in the levelized order
    for  Lev = 1; Lev < nLevels; Lev++  {
        clk = time.Now().UnixNano();
        fmt.Printf("LEVEL %3d  (out of %3d)   ", Lev, nLevels);
        pGiaLev = Gia_GenerateReducedLevel( pGia, Lev, &vRoots );
        nIters = Gia_ComputeEquivalencesLevel( p, pGiaLev, vRoots, Lev, fUseMiniSat );
        Gia_ManStop( pGiaLev );
        Vec_PtrFree( vRoots );
        fmt.Printf("Iters = %3d   ");
        // inlining of macro ABC_PRT
        Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((time.Now().UnixNano() - clk)))/(float64(int64(time.Second))));;
    }
    Gia_GiarfPrintClasses( pGia );
    // clean up
    Gia_ManEquivStop( p );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

