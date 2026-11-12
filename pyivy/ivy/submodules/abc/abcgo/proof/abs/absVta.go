package abs

import (
	"fmt"
	"os"
	"time"
)

/**CFile****************************************************************

  FileName    [absVta.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Abstraction package.]

  Synopsis    [Variable time-frame abstraction.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: absVta.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "sat/bsat/satSolver2.h"
//#include "base/main/main.h"
//#include "abs.h"

//ABC_NAMESPACE_IMPL_START

const VTA_LARGE =  0xFFFFFFF

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

    // object
type Vta_Obj_t struct {
    iObj           int32;
    iFrame           int32;
    iNext           int32;
    Prio uint32; /* C bitfield: 28 bits; review packing and truncation */  // related to VTA_LARGE
    Value uint32; /* C bitfield: 2 bits; review packing and truncation */
    fAdded uint32; /* C bitfield: 1 bits; review packing and truncation */
    fVisit uint32; /* C bitfield: 1 bits; review packing and truncation */
}

    // manager
type Vta_Man_t struct {
    // user data
    pGia    *Gia_Man_t;         // AIG manager
    pPars    *Abs_Par_t;        // parameters
    // internal data
    nObjs           int32;        // the number of objects
    nObjsAlloc           int32;   // the number of objects allocated
    nBins           int32;        // number of hash table entries
    pBins          *int32;        // hash table bins
    pObjs    *Vta_Obj_t;        // storage for objects
    vOrder    *Vec_Int_t;       // objects in DPS order
    // abstraction
    nObjBits           int32;     // the number of bits to represent objects
    nObjMask      uint32;     // object mask
    vFrames    *Vec_Ptr_t;      // start abstraction for each frame
    nWords           int32;       // the number of words in the record
    nCexes           int32;       // the number of CEXes
    nObjAdded           int32;    // objects added to the abstraction
    vSeens    *Vec_Int_t;       // seen objects
    vSeenGla    *Vec_Bit_t;     // seen objects in all frames
    nSeenGla           int32;     // seen objects in all frames
    nSeenAll           int32;     // seen objects in all frames
    // other data
    vCores    *Vec_Ptr_t;       // unsat core for each frame
    pSat  *sat_solver2;         // incremental SAT solver
    vAddedNew    *Vec_Int_t;    // the IDs of variables added to the solver
    // statistics 
    timeSat       abctime;
    timeUnsat       abctime;
    timeCex       abctime;
    timeOther       abctime;
}


// ternary simulation

const VTA_VAR0 =   1
const VTA_VAR1 =   2
const VTA_VARX =   3

func Vta_ValIs0(pThis *Vta_Obj_t, fCompl int32) int32 {
    if pThis.Value == VTA_VAR1 && (fCompl != 0) {
        return 1; }
    if pThis.Value == VTA_VAR0 && !(fCompl != 0) {
        return 1; }
    return 0;
}
func Vta_ValIs1(pThis *Vta_Obj_t, fCompl int32) int32 {
    if pThis.Value == VTA_VAR0 && (fCompl != 0) {
        return 1; }
    if pThis.Value == VTA_VAR1 && !(fCompl != 0) {
        return 1; }
    return 0;
}

func Vta_ManObj(p *Vta_Man_t, i int32) *Vta_Obj_t { assert(i >= 0 && i < p.nObjs, "i >= 0 && i < p.nObjs"); if i != 0 {
    return p.pObjs + i;
} else {
    return nil;
}                     }
func Vta_ObjId(p *Vta_Man_t, pObj *Vta_Obj_t) int32 { assert(pObj > p.pObjs && pObj < p.pObjs + p.nObjs, "pObj > p.pObjs && pObj < p.pObjs + p.nObjs"); return pObj - p.pObjs;      }

// Original macro:
// #define Vta_ManForEachObj( p, pObj, i )                                 \
//     for ( i = 1; (i < p->nObjs) && ((pObj) = Vta_ManObj(p, i)); i++ )

// Original macro:
// #define Vta_ManForEachObjObj( p, pObjVta, pObjGia, i )                  \
//     for ( i = 1; (i < p->nObjs) && ((pObjVta) = Vta_ManObj(p, i)) && ((pObjGia) = Gia_ManObj(p->pGia, pObjVta->iObj)); i++ )

// Original macro:
// #define Vta_ManForEachObjObjReverse( p, pObjVta, pObjGia, i )           \
//     for ( i = Vec_IntSize(vVec) - 1; (i >= 1) && ((pObjVta) = Vta_ManObj(p, i)) && ((pObjGia) = Gia_ManObj(p->pGia, pObjVta->iObj)); i++ )


// Original macro:
// #define Vta_ManForEachObjVec( vVec, p, pObj, i )                        \
//     for ( i = 0; (i < Vec_IntSize(vVec)) && ((pObj) = Vta_ManObj(p, Vec_IntEntry(vVec,i))); i++ )

// Original macro:
// #define Vta_ManForEachObjVecReverse( vVec, p, pObj, i )                 \
//     for ( i = Vec_IntSize(vVec) - 1; (i >= 0) && ((pObj) = Vta_ManObj(p, Vec_IntEntry(vVec,i))); i-- )


// Original macro:
// #define Vta_ManForEachObjObjVec( vVec, p, pObj, pObjG, i )              \
//     for ( i = 0; (i < Vec_IntSize(vVec)) && ((pObj) = Vta_ManObj(p, Vec_IntEntry(vVec,i))) && ((pObjG) = Gia_ManObj(p->pGia, pObj->iObj)); i++ )

// Original macro:
// #define Vta_ManForEachObjObjVecReverse( vVec, p, pObj, pObjG, i )       \
//     for ( i = Vec_IntSize(vVec) - 1; (i >= 0) && ((pObj) = Vta_ManObj(p, Vec_IntEntry(vVec,i))) && ((pObjG) = Gia_ManObj(p->pGia, pObj->iObj)); i-- )



// abstraction is given as an array of integers:
// - the first entry is the number of timeframes (F)
// - the next (F+1) entries give the beginning position of each timeframe
// - the following entries give the object IDs
// invariant:  assert( vec[vec[0]+1] == size(vec) );

// extern void Vga_ManAddClausesOne( Vta_Man_t * p, int iObj, int iFrame );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Converting from one array to per-frame arrays.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_VtaAbsToFrames(vAbs *Vec_Int_t) *Vec_Ptr_t {
    var vFrames *Vec_Ptr_t;
    var vFrame *Vec_Int_t;
    var i int32;
    var k int32;
    var Entry int32;
    var iStart int32;
    var iStop int32 = -1;
    var nFrames int32  = Vec_IntEntry( vAbs, 0 );
    assert(Vec_IntEntry(vAbs, nFrames+1) == Vec_IntSize(vAbs), "Vec_IntEntry(vAbs, nFrames+1) == Vec_IntSize(vAbs)");
    vFrames = Vec_PtrAlloc( nFrames );
    for  i = 0; i < nFrames; i++  {
        iStart = Vec_IntEntry( vAbs, i+1 );
        iStop  = Vec_IntEntry( vAbs, i+2 );
        vFrame = Vec_IntAlloc( iStop - iStart );
        // inlining of macro Vec_IntForEachEntryStartStop
        for  k = iStart; (k < iStop) && (((Entry) = Vec_IntEntry(vAbs, k)), 1); k++  {
            Vec_IntPush( vFrame, Entry );
        }
        Vec_PtrPush( vFrames, vFrame );
    }
    assert(iStop == Vec_IntSize(vAbs), "iStop == Vec_IntSize(vAbs)");
    return vFrames;
}

/**Function*************************************************************

  Synopsis    [Converting from per-frame arrays to one integer array.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_VtaFramesToAbs(vFrames *Vec_Vec_t) *Vec_Int_t {
    var vOne, vAbs *Vec_Int_t;
    var i, k, Entry, nSize int32;
    vAbs = Vec_IntAlloc( 2 + Vec_VecSize(vFrames) + Vec_VecSizeSize(vFrames) );
    Vec_IntPush( vAbs, Vec_VecSize(vFrames) );
    nSize = Vec_VecSize(vFrames) + 2;
    // inlining of macro Vec_VecForEachLevelInt
    for  i = 0; (i < Vec_VecSize(vFrames)) && (((vOne) = Vec_VecEntryInt(vFrames, i)), 1); i++  {
        Vec_IntPush( vAbs, nSize );
        nSize += Vec_IntSize( vOne );
    }
    Vec_IntPush( vAbs, nSize );
    assert(Vec_IntSize(vAbs) == Vec_VecSize(vFrames) + 2, "Vec_IntSize(vAbs) == Vec_VecSize(vFrames) + 2");
    // inlining of macro Vec_VecForEachLevelInt
    for  i = 0; (i < Vec_VecSize(vFrames)) && (((vOne) = Vec_VecEntryInt(vFrames, i)), 1); i++  {
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vOne)) && (((Entry) = Vec_IntEntry(vOne, k)), 1); k++  {
            Vec_IntPush( vAbs, Entry );
        }
    }
    assert(Vec_IntEntry(vAbs, Vec_IntEntry(vAbs,0)+1) == Vec_IntSize(vAbs), "Vec_IntEntry(vAbs, Vec_IntEntry(vAbs,0)+1) == Vec_IntSize(vAbs)");
    return vAbs;
}

/**Function*************************************************************

  Synopsis    [Detects how many frames are completed.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vta_ManDeriveAbsAll(p *Vec_Int_t, nWords int32) *Vec_Int_t {
    var vRes *Vec_Int_t;
    var pThis *uint32;
    var i int32;
    var w int32;
    var nObjs int32 = Vec_IntSize(p) / nWords;
    assert(Vec_IntSize(p) % nWords == 0, "Vec_IntSize(p) % nWords == 0");
    vRes = Vec_IntAlloc( nObjs );
    for  i = 0; i < nObjs; i++  {
        pThis = (unsigned *)Vec_IntEntryP( p, nWords * i );
        for  w = 0; w < nWords; w++  {
            if pThis[w] {
                break; }
        }
        Vec_IntPush( vRes, int32((w < nWords)) );
    }
    return vRes;
}

/**Function*************************************************************

  Synopsis    [Collect nodes/flops involved in different timeframes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntDoubleWidth(p *Vec_Int_t, nWords int32) int32 {
    // inlining of macro ABC_CALLOC
    var pArray *int32  = ((int *) calloc((size_t)(Vec_IntSize(p) * 2), sizeof(int)));
    var i int32;
    var w int32;
    var nObjs int32 = Vec_IntSize(p) / nWords;
    assert(Vec_IntSize(p) % nWords == 0, "Vec_IntSize(p) % nWords == 0");
    for  i = 0; i < nObjs; i++  {
        for  w = 0; w < nWords; w++  {
            pArray[2 * nWords * i + w] = p.pArray[nWords * i + w];
        }
    }
    p.pArray = nil;
    p.pArray = pArray;
    p.nSize *= 2;
    p.nCap = p.nSize;
    return 2 * nWords;
}



/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vga_ManHash(iObj int32, iFrame int32, nBins int32) int32 {
    return (uint32(((iObj + iFrame)*(iObj + iFrame + 1)))) % nBins;
}
func Vga_ManLookup(p *Vta_Man_t, iObj int32, iFrame int32) *int32 {
    var pThis *Vta_Obj_t;
    var pPlace *int32  = p.pBins + Vga_ManHash( iObj, iFrame, p.nBins );
    for  pThis = Vta_ManObj(p, *pPlace); 
          pThis;  pPlace = &pThis.iNext, 
          pThis = Vta_ManObj(p, *pPlace)  {
        if pThis.iObj == iObj && pThis.iFrame == iFrame {
            break; }
    }
    return pPlace;
}
func Vga_ManFind(p *Vta_Man_t, iObj int32, iFrame int32) *Vta_Obj_t {
    var pPlace *int32  = Vga_ManLookup( p, iObj, iFrame );
    return Vta_ManObj(p, *pPlace);
}
func Vga_ManFindOrAdd(p *Vta_Man_t, iObj int32, iFrame int32) *Vta_Obj_t {
    var pThis *Vta_Obj_t;
    var i int32;
    var pPlace *int32;
    assert(iObj >= 0 && iFrame >= -1, "iObj >= 0 && iFrame >= -1");
    if p.nObjs == p.nObjsAlloc {
        // resize objects
        // inlining of macro ABC_REALLOC
        
        if p.pObjs {
            p.pObjs = ((Vta_Obj_t *) realloc((char *)(p.pObjs), sizeof(Vta_Obj_t) * (size_t)(2 * p.nObjsAlloc)));
        } else {
            p.pObjs = ((Vta_Obj_t *) malloc(sizeof(Vta_Obj_t) * (size_t)(2 * p.nObjsAlloc)));
        }
        memset( p.pObjs + p.nObjsAlloc, 0, p.nObjsAlloc * sizeof(Vta_Obj_t) );
        p.nObjsAlloc *= 2;
        // rehash entries in the table
        p.pBins = nil;
        p.nBins = Abc_PrimeCudd( 2 * p.nBins );
        // inlining of macro ABC_CALLOC
        p.pBins = ((int *) calloc((size_t)(p.nBins), sizeof(int)));
        // inlining of macro Vta_ManForEachObj
        for  i = 1; (i < p.nObjs) && ((pThis) = Vta_ManObj(p, i)); i++  {
            pThis.iNext = 0;
            pPlace = Vga_ManLookup( p, pThis.iObj, pThis.iFrame );
            assert(*pPlace == 0, "*pPlace == 0");
            *pPlace = i;
        }
    }
    pPlace = Vga_ManLookup( p, iObj, iFrame );
    if *pPlace {
        return Vta_ManObj(p, *pPlace); }
    *pPlace = p.nObjs++;
    pThis = Vta_ManObj(p, *pPlace);
    pThis.iObj   = iObj;
    pThis.iFrame = iFrame;
    return pThis;
}
func Vga_ManDelete(p *Vta_Man_t, iObj int32, iFrame int32) {
    var pPlace *int32  = Vga_ManLookup( p, iObj, iFrame );
    var pThis *Vta_Obj_t  = Vta_ManObj(p, *pPlace);
    assert(pThis != nil, "pThis != nil");
    *pPlace = pThis.iNext;
    pThis.iNext = -1;
}


/**Function*************************************************************

  Synopsis    [Derives counter-example using current assignments.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vga_ManDeriveCex(p *Vta_Man_t) *Abc_Cex_t {
    var pCex *Abc_Cex_t;
    var pThis *Vta_Obj_t;
    var pObj *Gia_Obj_t;
    var i int32;
    pCex = Abc_CexAlloc( Gia_ManRegNum(p.pGia), Gia_ManPiNum(p.pGia), p.pPars.iFrame+1 );
    pCex.iPo = 0;
    pCex.iFrame = p.pPars.iFrame;
    // inlining of macro Vta_ManForEachObjObj
    for  i = 1; (i < p.nObjs) && ((pThis) = Vta_ManObj(p, i)) && ((pObj) = Gia_ManObj(p.pGia, pThis.iObj)); i++  {
        if Gia_ObjIsPi(p.pGia, pObj) && sat_solver2_var_value(p.pSat, Vta_ObjId(p, pThis)) {
            Abc_InfoSetBit( pCex.pData, pCex.nRegs + pThis.iFrame * pCex.nPis + Gia_ObjCioId(pObj) ); }
    }
    return pCex;
}

/**Function*************************************************************

  Synopsis    [Remaps core into frame/node pairs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vta_ManUnsatCoreRemap(p *Vta_Man_t, vCore *Vec_Int_t) {
    var pThis *Vta_Obj_t;
    var i, Entry int32;
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vCore)) && (((Entry) = Vec_IntEntry(vCore, i)), 1); i++  {
        pThis = Vta_ManObj( p, Entry );
        Entry = (pThis.iFrame << p.nObjBits) | pThis.iObj;
        Vec_IntWriteEntry( vCore, i, Entry );
    }
}

/**Function*************************************************************

  Synopsis    [Compares two objects by their distance.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vta_ManComputeDepthIncrease(pp1 **Vta_Obj_t, pp2 **Vta_Obj_t) int32 {
    var Diff int32  = (*pp1).Prio - (*pp2).Prio;
    if Diff < 0 {
        return -1; }
    if Diff > 0 { 
        return 1; }
    Diff = (*pp1) - (*pp2);
    if Diff < 0 {
        return -1; }
    if Diff > 0 { 
        return 1; }
    return 0; 
}

/**Function*************************************************************

  Synopsis    [Returns 1 if the object is already used.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vta_ManObjIsUsed(p *Vta_Man_t, iObj int32) int32 {
    var i int32;
    var pInfo *uint32  = (unsigned *)Vec_IntEntryP( p.vSeens, p.nWords * iObj );
    for  i = 0; i < p.nWords; i++  {
        if pInfo[i] {
            return 1; }
    }
    return 0;
}

/**Function*************************************************************

  Synopsis    [Finds predecessors of the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vta_ObjPreds(p *Vta_Man_t, pThis *Vta_Obj_t, pObj *Gia_Obj_t, ppThis0 **Vta_Obj_t, ppThis1 **Vta_Obj_t) {
    *ppThis0 = nil;
    *ppThis1 = nil;
//    if ( !pThis->fAdded )
//        return;
    assert(!Gia_ObjIsPi(p.pGia, pObj), "!Gia_ObjIsPi(p.pGia, pObj)");
    if Gia_ObjIsConst0(pObj) || (Gia_ObjIsCi(pObj) && pThis.iFrame == 0) {
        return; }
    if Gia_ObjIsAnd(pObj) {
        *ppThis0 = Vga_ManFind( p, Gia_ObjFaninId0p(p.pGia, pObj), pThis.iFrame );
        *ppThis1 = Vga_ManFind( p, Gia_ObjFaninId1p(p.pGia, pObj), pThis.iFrame );
//        assert( *ppThis0 && *ppThis1 );
        return;
    }
    assert(Gia_ObjIsRo(p.pGia, pObj) && pThis.iFrame > 0, "Gia_ObjIsRo(p.pGia, pObj) && pThis.iFrame > 0");
    pObj = Gia_ObjRoToRi( p.pGia, pObj );
    *ppThis0 = Vga_ManFind( p, Gia_ObjFaninId0p(p.pGia, pObj), pThis.iFrame-1 );
//    assert( *ppThis0 );
}

/**Function*************************************************************

  Synopsis    [Collect const/PI/RO/AND in a topological order.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vta_ManCollectNodes_rec(p *Vta_Man_t, pThis *Vta_Obj_t, vOrder *Vec_Int_t) {
    var pObj *Gia_Obj_t;
    var pThis0, pThis1 *Vta_Obj_t;
    if pThis.fVisit {
        return; }
    pThis.fVisit = 1;
    pObj = Gia_ManObj( p.pGia, pThis.iObj );
    if pThis.fAdded {
        Vta_ObjPreds( p, pThis, pObj, &pThis0, &pThis1 );
        if pThis0 { Vta_ManCollectNodes_rec( p, pThis0, vOrder ); }
        if pThis1 != nil { Vta_ManCollectNodes_rec( p, pThis1, vOrder ); }
    }
    Vec_IntPush( vOrder, Vta_ObjId(p, pThis) );
}
func Vta_ManCollectNodes(p *Vta_Man_t, f int32) *Vec_Int_t {
    var pThis *Vta_Obj_t;
    var pObj *Gia_Obj_t;
    Vec_IntClear( p.vOrder );
    pObj = Gia_ManPo( p.pGia, 0 );
    pThis = Vga_ManFind( p, Gia_ObjFaninId0p(p.pGia, pObj), f );
    assert(pThis != nil, "pThis != nil");
    assert(!pThis.fVisit, "!pThis.fVisit");
    Vta_ManCollectNodes_rec( p, pThis, p.vOrder );
    assert(pThis.fVisit, "pThis.fVisit");
    return p.vOrder;
}

/**Function*************************************************************

  Synopsis    [Refines abstraction.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vta_ManSatVerify(p *Vta_Man_t) {
    var pThis, pThis0, pThis1 *Vta_Obj_t;
    var pObj *Gia_Obj_t;
    var i int32;
    // inlining of macro Vta_ManForEachObj
    for  i = 1; (i < p.nObjs) && ((pThis) = Vta_ManObj(p, i)); i++  {
        if sat_solver2_var_value(p.pSat, i) {
            pThis.Value = VTA_VAR1;
        } else {
            pThis.Value = VTA_VAR0;
        }
    }
    // inlining of macro Vta_ManForEachObjObj
    for  i = 1; (i < p.nObjs) && ((pThis) = Vta_ManObj(p, i)) && ((pObj) = Gia_ManObj(p.pGia, pThis.iObj)); i++  {
        if !pThis.fAdded {
            continue; }
        Vta_ObjPreds( p, pThis, pObj, &pThis0, &pThis1 );
        if Gia_ObjIsAnd(pObj) {
            if pThis.Value == VTA_VAR1 {
                assert((Vta_ValIs1(pThis0, Gia_ObjFaninC0(pObj)) != 0) && (Vta_ValIs1(pThis1, Gia_ObjFaninC1(pObj)) != 0), "Vta_ValIs1(pThis0, Gia_ObjFaninC0(pObj)) && Vta_ValIs1(pThis1, Gia_ObjFaninC1(pObj))");
            } else if pThis.Value == VTA_VAR0 {
                assert((Vta_ValIs0(pThis0, Gia_ObjFaninC0(pObj)) != 0) || (Vta_ValIs0(pThis1, Gia_ObjFaninC1(pObj)) != 0), "Vta_ValIs0(pThis0, Gia_ObjFaninC0(pObj)) || Vta_ValIs0(pThis1, Gia_ObjFaninC1(pObj))");
            } else { assert(false, "0"); }
        
        } else if Gia_ObjIsRo(p.pGia, pObj) {
            pObj = Gia_ObjRoToRi( p.pGia, pObj );
            if pThis.iFrame == 0 {
                assert(pThis.Value == VTA_VAR0, "pThis.Value == VTA_VAR0");
            } else if pThis.Value == VTA_VAR0 {
                assert((Vta_ValIs0(pThis0, Gia_ObjFaninC0(pObj)) != 0), "Vta_ValIs0(pThis0, Gia_ObjFaninC0(pObj))");
            } else if pThis.Value == VTA_VAR1 {
                assert((Vta_ValIs1(pThis0, Gia_ObjFaninC0(pObj)) != 0), "Vta_ValIs1(pThis0, Gia_ObjFaninC0(pObj))");
            } else { assert(false, "0"); }
        }
    }
}

/**Function*************************************************************

  Synopsis    [Refines abstraction.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vta_ManProfileAddition(p *Vta_Man_t, vTermsToAdd *Vec_Int_t) {
    var pThis *Vta_Obj_t;
    var pObj *Gia_Obj_t;
    // profile the added ones
    // inlining of macro ABC_CALLOC
    int i, * pCounters = ((int *) calloc((size_t)(p.pPars.iFrame+1), sizeof(int)));
    // inlining of macro Vta_ManForEachObjObjVec
    for  i = 0; (i < Vec_IntSize(vTermsToAdd)) && ((pThis) = Vta_ManObj(p, Vec_IntEntry(vTermsToAdd,i))) && ((pObj) = Gia_ManObj(p.pGia, pThis.iObj)); i++  {
        pCounters[pThis.iFrame]++;
    }
    for  i = 0; i <= p.pPars.iFrame; i++  {
        Abc_Print( 1, "%2d", pCounters[i] );
    }
    Abc_Print( 1, "***\n" );
}

/**Function*************************************************************

  Synopsis    [Refines abstraction.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vta_ManRefineAbstraction(p *Vta_Man_t, f int32) *Abc_Cex_t {
    var fVerify int32;
    var pCex *Abc_Cex_t  = nil;
    var vOrder, vTermsToAdd *Vec_Int_t;
    var vTermsUsed, vTermsUnused *Vec_Ptr_t;
    var pThis, pThis0, pThis1, pTop *Vta_Obj_t;
    var pObj *Gia_Obj_t;
    var i, Counter int32;

    if fVerify != 0 {
    Vta_ManSatVerify( p ); }

    // collect nodes in a topological order
    vOrder = Vta_ManCollectNodes( p, f );
    // inlining of macro Vta_ManForEachObjObjVec
    for  i = 0; (i < Vec_IntSize(vOrder)) && ((pThis) = Vta_ManObj(p, Vec_IntEntry(vOrder,i))) && ((pObj) = Gia_ManObj(p.pGia, pThis.iObj)); i++  {
        pThis.Prio = VTA_LARGE;
        if sat_solver2_var_value(p.pSat, Vta_ObjId(p, pThis)) {
            pThis.Value = VTA_VAR1;
        } else {
            pThis.Value = VTA_VAR0;
        }
        pThis.fVisit = 0;
    }

    // verify
    if fVerify != 0 {
    // inlining of macro Vta_ManForEachObjObjVec
    for  i = 0; (i < Vec_IntSize(vOrder)) && ((pThis) = Vta_ManObj(p, Vec_IntEntry(vOrder,i))) && ((pObj) = Gia_ManObj(p.pGia, pThis.iObj)); i++  {
        if !pThis.fAdded {
            continue; }
        Vta_ObjPreds( p, pThis, pObj, &pThis0, &pThis1 );
        if Gia_ObjIsAnd(pObj) {
            if pThis.Value == VTA_VAR1 {
                assert((Vta_ValIs1(pThis0, Gia_ObjFaninC0(pObj)) != 0) && (Vta_ValIs1(pThis1, Gia_ObjFaninC1(pObj)) != 0), "Vta_ValIs1(pThis0, Gia_ObjFaninC0(pObj)) && Vta_ValIs1(pThis1, Gia_ObjFaninC1(pObj))");
            } else if pThis.Value == VTA_VAR0 {
                assert((Vta_ValIs0(pThis0, Gia_ObjFaninC0(pObj)) != 0) || (Vta_ValIs0(pThis1, Gia_ObjFaninC1(pObj)) != 0), "Vta_ValIs0(pThis0, Gia_ObjFaninC0(pObj)) || Vta_ValIs0(pThis1, Gia_ObjFaninC1(pObj))");
            } else { assert(false, "0"); }
        
        } else if Gia_ObjIsRo(p.pGia, pObj) {
            pObj = Gia_ObjRoToRi( p.pGia, pObj );
            if pThis.iFrame == 0 {
                assert(pThis.Value == VTA_VAR0, "pThis.Value == VTA_VAR0");
            } else if pThis.Value == VTA_VAR0 {
                assert((Vta_ValIs0(pThis0, Gia_ObjFaninC0(pObj)) != 0), "Vta_ValIs0(pThis0, Gia_ObjFaninC0(pObj))");
            } else if pThis.Value == VTA_VAR1 {
                assert((Vta_ValIs1(pThis0, Gia_ObjFaninC0(pObj)) != 0), "Vta_ValIs1(pThis0, Gia_ObjFaninC0(pObj))");
            } else { assert(false, "0"); }
        }
    } }

    // compute distance in reverse order
    pThis = Vta_ManObj( p, Vec_IntEntryLast(vOrder) );
    pThis.Prio  = 1;
    // collect used and unused terms
    vTermsUsed   = Vec_PtrAlloc( 1015 );
    vTermsUnused = Vec_PtrAlloc( 1016 );
    // inlining of macro Vta_ManForEachObjObjVecReverse
    for  i = Vec_IntSize(vOrder) - 1; (i >= 0) && ((pThis) = Vta_ManObj(p, Vec_IntEntry(vOrder,i))) && ((pObj) = Gia_ManObj(p.pGia, pThis.iObj)); i--  {
        // there is no unreachable states
        assert(pThis.Prio < VTA_LARGE, "pThis.Prio < VTA_LARGE");
        // skip constants and PIs
        if Gia_ObjIsConst0(pObj) || Gia_ObjIsPi(p.pGia, pObj) {
            pThis.Prio = 0; // set highest priority
            continue;
        }
        // collect terminals
        assert(Gia_ObjIsAnd(pObj) || Gia_ObjIsRo(p.pGia, pObj), "Gia_ObjIsAnd(pObj) || Gia_ObjIsRo(p.pGia, pObj)");
        if !pThis.fAdded {
            assert(pThis.Prio > 0, "pThis.Prio > 0");
            if Vta_ManObjIsUsed(p, pThis.iObj) != 0 {
                Vec_PtrPush( vTermsUsed, pThis );
            } else {
                Vec_PtrPush( vTermsUnused, pThis ); }
            continue;
        }
        // propagate
        Vta_ObjPreds( p, pThis, pObj, &pThis0, &pThis1 );
        if pThis0 { 
            pThis0.Prio = Abc_MinInt( pThis0.Prio, pThis.Prio + 1 ); }
        if pThis1 != nil { 
            pThis1.Prio = Abc_MinInt( pThis1.Prio, pThis.Prio + 1 ); }
    }

/*
    Vta_ManForEachObjObjVecReverse( vOrder, p, pThis, pObj, i )
        if ( pThis->Prio > 0 )
            pThis->Prio = 10;
*/
/*
    // update priorities according to reconvergence counters
    Vec_PtrForEachEntry( Vta_Obj_t *, vTermsUsed, pThis, i )
    {
        Vta_Obj_t * pThis0, * pThis1;
        Gia_Obj_t * pObj = Gia_ManObj( p->pGia, pThis->iObj );
        Vta_ObjPreds( p, pThis, pObj, &pThis0, &pThis1 );
        pThis->Prio += 10000000;
        if ( pThis0 )
            pThis->Prio -= 1000000 * pThis0->fAdded;
        if ( pThis1 )
            pThis->Prio -= 1000000 * pThis1->fAdded;
    }
    Vec_PtrForEachEntry( Vta_Obj_t *, vTermsUnused, pThis, i )
    {
        Vta_Obj_t * pThis0, * pThis1;
        Gia_Obj_t * pObj = Gia_ManObj( p->pGia, pThis->iObj );
        Vta_ObjPreds( p, pThis, pObj, &pThis0, &pThis1 );
        pThis->Prio += 10000000;
        if ( pThis0 )
            pThis->Prio -= 1000000 * pThis0->fAdded;
        if ( pThis1 )
            pThis->Prio -= 1000000 * pThis1->fAdded;
    }
*/


    // update priorities according to reconvergence counters
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vTermsUsed)) && (((pThis) = (Vta_Obj_t *)Vec_PtrEntry(vTermsUsed, i)), 1); i++  {
        pThis.Prio = pThis.iObj;
    }
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vTermsUnused)) && (((pThis) = (Vta_Obj_t *)Vec_PtrEntry(vTermsUnused, i)), 1); i++  {
        pThis.Prio = pThis.iObj;
    }


    // objects with equal distance should receive priority based on number
    // those objects whose prototypes have been added in other timeframes
    // should have higher priority than the current object
    Vec_PtrSort( vTermsUsed,   (int (*)(void))Vta_ManComputeDepthIncrease );
    Vec_PtrSort( vTermsUnused, (int (*)(void))Vta_ManComputeDepthIncrease );
    if Vec_PtrSize(vTermsUsed) > 1 {
        pThis0 = (Vta_Obj_t *)Vec_PtrEntry(vTermsUsed, 0);
        pThis1 = (Vta_Obj_t *)Vec_PtrEntryLast(vTermsUsed);
        assert(pThis0.Prio <= pThis1.Prio, "pThis0.Prio <= pThis1.Prio");
    }
    // assign the priority based on these orders
    Counter = 1;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vTermsUsed)) && (((pThis) = (Vta_Obj_t *)Vec_PtrEntry(vTermsUsed, i)), 1); i++  {
        pThis.Prio = Counter++;
    }
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vTermsUnused)) && (((pThis) = (Vta_Obj_t *)Vec_PtrEntry(vTermsUnused, i)), 1); i++  {
        pThis.Prio = Counter++;
    }
//    Abc_Print( 1, "Used %d  Unused %d\n", Vec_PtrSize(vTermsUsed), Vec_PtrSize(vTermsUnused) );


    // propagate in the direct order
    // inlining of macro Vta_ManForEachObjObjVec
    for  i = 0; (i < Vec_IntSize(vOrder)) && ((pThis) = Vta_ManObj(p, Vec_IntEntry(vOrder,i))) && ((pObj) = Gia_ManObj(p.pGia, pThis.iObj)); i++  {
        assert(pThis.fVisit == 0, "pThis.fVisit == 0");
        assert(pThis.Prio < VTA_LARGE, "pThis.Prio < VTA_LARGE");
        // skip terminal objects
        if !pThis.fAdded {
            continue; }
        // assumes that values are assigned!!!
        assert(pThis.Value != 0, "pThis.Value != 0");
        // propagate
        if Gia_ObjIsAnd(pObj) {
            pThis0 = Vga_ManFind( p, Gia_ObjFaninId0p(p.pGia, pObj), pThis.iFrame );
            pThis1 = Vga_ManFind( p, Gia_ObjFaninId1p(p.pGia, pObj), pThis.iFrame );
            assert(pThis0 && (pThis1 != nil), "pThis0 && pThis1");
            if pThis.Value == VTA_VAR1 {
                assert((Vta_ValIs1(pThis0, Gia_ObjFaninC0(pObj)) != 0) && (Vta_ValIs1(pThis1, Gia_ObjFaninC1(pObj)) != 0), "Vta_ValIs1(pThis0, Gia_ObjFaninC0(pObj)) && Vta_ValIs1(pThis1, Gia_ObjFaninC1(pObj))");
                pThis.Prio = Abc_MaxInt( pThis0.Prio, pThis1.Prio );
            
            } else if pThis.Value == VTA_VAR0 {
                if (Vta_ValIs0(pThis0, Gia_ObjFaninC0(pObj)) != 0) && (Vta_ValIs0(pThis1, Gia_ObjFaninC1(pObj)) != 0) {
                    pThis.Prio = Abc_MinInt( pThis0.Prio, pThis1.Prio ); // choice!!!
                } else if Vta_ValIs0(pThis0, Gia_ObjFaninC0(pObj)) != 0 {
                    pThis.Prio = pThis0.Prio;
                } else if Vta_ValIs0(pThis1, Gia_ObjFaninC1(pObj)) != 0 {
                    pThis.Prio = pThis1.Prio;
                } else { assert(false, "0"); }
            
            } else { assert(false, "0"); }
        
        } else if Gia_ObjIsRo(p.pGia, pObj) {
            if pThis.iFrame > 0 {
                pObj = Gia_ObjRoToRi( p.pGia, pObj );
                pThis0 = Vga_ManFind( p, Gia_ObjFaninId0p(p.pGia, pObj), pThis.iFrame-1 );
                assert(pThis0, "pThis0");
                pThis.Prio = pThis0.Prio;
            
            } else {
                pThis.Prio = 0; }
        
        } else if Gia_ObjIsConst0(pObj) {
            pThis.Prio = 0;
        } else {
            assert(false, "0"); }
    }

    // select important values
    pTop = Vta_ManObj( p, Vec_IntEntryLast(vOrder) );
    pTop.fVisit = 1;
    vTermsToAdd = Vec_IntAlloc( 100 );
    // inlining of macro Vta_ManForEachObjObjVecReverse
    for  i = Vec_IntSize(vOrder) - 1; (i >= 0) && ((pThis) = Vta_ManObj(p, Vec_IntEntry(vOrder,i))) && ((pObj) = Gia_ManObj(p.pGia, pThis.iObj)); i--  {
        if !pThis.fVisit {
            continue; }
        pThis.fVisit = 0;
        assert(pThis.Prio >= 0 && pThis.Prio <= pTop.Prio, "pThis.Prio >= 0 && pThis.Prio <= pTop.Prio");
        // skip terminal objects
        if !pThis.fAdded {
            assert(Gia_ObjIsAnd(pObj) || Gia_ObjIsRo(p.pGia, pObj) || Gia_ObjIsConst0(pObj) || Gia_ObjIsPi(p.pGia, pObj), "Gia_ObjIsAnd(pObj) || Gia_ObjIsRo(p.pGia, pObj) || Gia_ObjIsConst0(pObj) || Gia_ObjIsPi(p.pGia, pObj)");
            Vec_IntPush( vTermsToAdd, Vta_ObjId(p, pThis) );
            continue;
        }
        // assumes that values are assigned!!!
        assert(pThis.Value != 0, "pThis.Value != 0");
        // propagate
        if Gia_ObjIsAnd(pObj) {
            pThis0 = Vga_ManFind( p, Gia_ObjFaninId0p(p.pGia, pObj), pThis.iFrame );
            pThis1 = Vga_ManFind( p, Gia_ObjFaninId1p(p.pGia, pObj), pThis.iFrame );
            assert(pThis0 && (pThis1 != nil), "pThis0 && pThis1");
            if pThis.Value == VTA_VAR1 {
                assert((Vta_ValIs1(pThis0, Gia_ObjFaninC0(pObj)) != 0) && (Vta_ValIs1(pThis1, Gia_ObjFaninC1(pObj)) != 0), "Vta_ValIs1(pThis0, Gia_ObjFaninC0(pObj)) && Vta_ValIs1(pThis1, Gia_ObjFaninC1(pObj))");
                assert(pThis0.Prio <= pThis.Prio, "pThis0.Prio <= pThis.Prio");
                assert(pThis1.Prio <= pThis.Prio, "pThis1.Prio <= pThis.Prio");
                pThis0.fVisit = 1;
                pThis1.fVisit = 1;
            
            } else if pThis.Value == VTA_VAR0 {
                if (Vta_ValIs0(pThis0, Gia_ObjFaninC0(pObj)) != 0) && (Vta_ValIs0(pThis1, Gia_ObjFaninC1(pObj)) != 0) {
                    if pThis0.fVisit {
                    
                    } else if pThis1.fVisit {
                    
                    } else if pThis0.Prio <= pThis1.Prio { // choice!!!
                    
                        pThis0.fVisit = 1;
                        assert(pThis0.Prio == pThis.Prio, "pThis0.Prio == pThis.Prio");
                    
                    } else {
                        pThis1.fVisit = 1;
                        assert(pThis1.Prio == pThis.Prio, "pThis1.Prio == pThis.Prio");
                    }
                
                } else if Vta_ValIs0(pThis0, Gia_ObjFaninC0(pObj)) != 0 {
                    pThis0.fVisit = 1;
                    assert(pThis0.Prio == pThis.Prio, "pThis0.Prio == pThis.Prio");
                
                } else if Vta_ValIs0(pThis1, Gia_ObjFaninC1(pObj)) != 0 {
                    pThis1.fVisit = 1;
                    assert(pThis1.Prio == pThis.Prio, "pThis1.Prio == pThis.Prio");
                
                } else { assert(false, "0"); }
            
            } else { assert(false, "0"); }
        
        } else if Gia_ObjIsRo(p.pGia, pObj) {
            if pThis.iFrame > 0 {
                pObj = Gia_ObjRoToRi( p.pGia, pObj );
                pThis0 = Vga_ManFind( p, Gia_ObjFaninId0p(p.pGia, pObj), pThis.iFrame-1 );
                assert(pThis0, "pThis0");
                pThis0.fVisit = 1;
                assert(pThis0.Prio == pThis.Prio, "pThis0.Prio == pThis.Prio");
            }
        
        } else if !Gia_ObjIsConst0(pObj) {
            assert(false, "0"); }
    }

    if p.pPars.fAddLayer {
        // mark those currently included
        // inlining of macro Vta_ManForEachObjVec
        for  i = 0; (i < Vec_IntSize(vTermsToAdd)) && ((pThis) = Vta_ManObj(p, Vec_IntEntry(vTermsToAdd,i))); i++  {
            assert(pThis.fVisit == 0, "pThis.fVisit == 0");
            pThis.fVisit = 1;
        }
        // add used terms, which have close relationship
        Counter = Vec_IntSize(vTermsToAdd);
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(vTermsUsed)) && (((pThis) = (Vta_Obj_t *)Vec_PtrEntry(vTermsUsed, i)), 1); i++  {
            if pThis.fVisit {
                continue; }
    //        Vta_ObjPreds( p, pThis, Gia_ManObj(p->pGia, pThis->iObj), &pThis0, &pThis1 );
    //        if ( (pThis0 && (pThis0->fAdded || pThis0->fVisit)) || (pThis1 && (pThis1->fAdded || pThis1->fVisit)) )
                Vec_IntPush( vTermsToAdd, Vta_ObjId(p, pThis) );
        }
        // remove those currenty included
        // inlining of macro Vta_ManForEachObjVec
        for  i = 0; (i < Vec_IntSize(vTermsToAdd)) && ((pThis) = Vta_ManObj(p, Vec_IntEntry(vTermsToAdd,i))); i++  {
            pThis.fVisit = 0;
        }
    }
//    printf( "\n%d -> %d\n", Counter, Vec_IntSize(vTermsToAdd) );
//Vec_IntReverseOrder( vTermsToAdd );
//Vec_IntSort( vTermsToAdd, 1 );


    // cleanup
    Vec_PtrFree( vTermsUsed );
    Vec_PtrFree( vTermsUnused );


    if fVerify != 0 {
    // verify
    // inlining of macro Vta_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vOrder)) && ((pThis) = Vta_ManObj(p, Vec_IntEntry(vOrder,i))); i++  {
        pThis.Value = VTA_VARX;
    }
    // inlining of macro Vta_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vTermsToAdd)) && ((pThis) = Vta_ManObj(p, Vec_IntEntry(vTermsToAdd,i))); i++  {
        assert(!pThis.fAdded, "!pThis.fAdded");
        if sat_solver2_var_value(p.pSat, Vta_ObjId(p, pThis)) {
            pThis.Value = VTA_VAR1;
        } else {
            pThis.Value = VTA_VAR0;
        }
    }
    // simulate 
    // inlining of macro Vta_ManForEachObjObjVec
    for  i = 0; (i < Vec_IntSize(vOrder)) && ((pThis) = Vta_ManObj(p, Vec_IntEntry(vOrder,i))) && ((pObj) = Gia_ManObj(p.pGia, pThis.iObj)); i++  {
        assert(pThis.fVisit == 0, "pThis.fVisit == 0");
        if !pThis.fAdded {
            continue; }
        if Gia_ObjIsAnd(pObj) {
            pThis0 = Vga_ManFind( p, Gia_ObjFaninId0p(p.pGia, pObj), pThis.iFrame );
            pThis1 = Vga_ManFind( p, Gia_ObjFaninId1p(p.pGia, pObj), pThis.iFrame );
            assert(pThis0 && (pThis1 != nil), "pThis0 && pThis1");
            if (Vta_ValIs1(pThis0, Gia_ObjFaninC0(pObj)) != 0) && (Vta_ValIs1(pThis1, Gia_ObjFaninC1(pObj)) != 0) {
                pThis.Value = VTA_VAR1;
            } else if (Vta_ValIs0(pThis0, Gia_ObjFaninC0(pObj)) != 0) || (Vta_ValIs0(pThis1, Gia_ObjFaninC1(pObj)) != 0) {
                pThis.Value = VTA_VAR0;
            } else {
                pThis.Value = VTA_VARX; }
        
        } else if Gia_ObjIsRo(p.pGia, pObj) {
            if pThis.iFrame > 0 {
                pObj = Gia_ObjRoToRi( p.pGia, pObj );
                pThis0 = Vga_ManFind( p, Gia_ObjFaninId0p(p.pGia, pObj), pThis.iFrame-1 );
                assert(pThis0, "pThis0");
                if Vta_ValIs0(pThis0, Gia_ObjFaninC0(pObj)) != 0 {
                    pThis.Value = VTA_VAR0;
                } else if Vta_ValIs1(pThis0, Gia_ObjFaninC0(pObj)) != 0 {
                    pThis.Value = VTA_VAR1;
                } else {
                    pThis.Value = VTA_VARX; }
            
            } else {
                pThis.Value = VTA_VAR0;
            }
        
        } else if Gia_ObjIsConst0(pObj) {
            pThis.Value = VTA_VAR0;
        
        } else { assert(false, "0"); }
        // double check the solver
        assert(pThis.Value == VTA_VARX || int32(pThis.Value) == (sat_solver2_var_value(p.pSat, Vta_ObjId(p, pThis)) ? VTA_VAR1 : VTA_VAR0), "pThis.Value == VTA_VARX || int32(pThis.Value) == (sat_solver2_var_value(p.pSat, Vta_ObjId(p, pThis)) ? VTA_VAR1 : VTA_VAR0)");
    }

    // check the output
    if !(Vta_ValIs1(pTop, Gia_ObjFaninC0(Gia_ManPo(p.pGia, 0))) != 0) {
        Abc_Print( 1, "Vta_ManRefineAbstraction(): Terminary simulation verification failed!\n" ); }
//    else
//        Abc_Print( 1, "Verification OK.\n" );
    }


    // produce true counter-example
    if pTop.Prio == 0 {
        pCex = Vga_ManDeriveCex( p );
    } else {
//       Vta_ManProfileAddition( p, vTermsToAdd );

        // inlining of macro Vta_ManForEachObjObjVec
        for  i = 0; (i < Vec_IntSize(vTermsToAdd)) && ((pThis) = Vta_ManObj(p, Vec_IntEntry(vTermsToAdd,i))) && ((pObj) = Gia_ManObj(p.pGia, pThis.iObj)); i++  {
            if !Gia_ObjIsPi(p.pGia, pObj) {
                Vga_ManAddClausesOne( p, pThis.iObj, pThis.iFrame ); }
        }
        sat_solver2_simplify( p.pSat );
    }
    p.nObjAdded += Vec_IntSize(vTermsToAdd);
    Vec_IntFree( vTermsToAdd );
    return pCex;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vga_ManStart(pGia *Gia_Man_t, pPars *Abs_Par_t) *Vta_Man_t {
    
    // inlining of macro ABC_CALLOC
    p := &Vta_Man_t{};
    p.pGia        = pGia;
    p.pPars       = pPars;
    // internal data
    p.nObjsAlloc  = (1 << 18);
    // inlining of macro ABC_CALLOC
    p.pObjs       = ((Vta_Obj_t *) calloc((size_t)(p.nObjsAlloc), sizeof(Vta_Obj_t)));
    p.nObjs       = 1;
    p.nBins       = Abc_PrimeCudd( 2*p.nObjsAlloc );
    // inlining of macro ABC_CALLOC
    p.pBins       = ((int *) calloc((size_t)(p.nBins), sizeof(int)));
    p.vOrder      = Vec_IntAlloc( 1013 );
    // abstraction
    p.nObjBits    = Abc_Base2Log( Gia_ManObjNum(pGia) );
    p.nObjMask    = (1 << p.nObjBits) - 1;
    assert(Gia_ManObjNum(pGia) <= int32(p.nObjMask), "Gia_ManObjNum(pGia) <= int32(p.nObjMask)");
    p.nWords      = 1;
    p.vSeens      = Vec_IntStart( Gia_ManObjNum(pGia) * p.nWords );
    p.vSeenGla    = Vec_BitStart( Gia_ManObjNum(pGia) );
    p.nSeenGla    = 1;
    p.nSeenAll    = 1;
    // other data
    p.vCores      = Vec_PtrAlloc( 100 );
    p.pSat        = sat_solver2_new();
    p.pSat.pPrf1 = Vec_SetAlloc( 20 );
//    p->pSat->fVerbose = p->pPars->fVerbose;
//    sat_solver2_set_learntmax( p->pSat, pPars->nLearnedMax );
    p.pSat.nLearntStart = p.pPars.nLearnedStart;
    p.pSat.nLearntDelta = p.pPars.nLearnedDelta;
    p.pSat.nLearntRatio = p.pPars.nLearnedPerce;
    p.pSat.nLearntMax   = p.pSat.nLearntStart;
    // start the abstraction
    assert(pGia.vObjClasses != nil, "pGia.vObjClasses != nil");
    p.vFrames = Gia_VtaAbsToFrames( pGia.vObjClasses );
    p.vAddedNew   = Vec_IntAlloc( 1000 );
    return p;
}

/**Function*************************************************************

  Synopsis    [Delete manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vga_ManStop(p *Vta_Man_t) {
    if p.pPars.fVerbose {
        Abc_Print( 1, "SAT solver:  Var = %d  Cla = %d  Conf = %d  Lrn = %d  Reduce = %d  Cex = %d  Objs+ = %d\n", 
            sat_solver2_nvars(p.pSat), sat_solver2_nclauses(p.pSat), sat_solver2_nconflicts(p.pSat), 
            sat_solver2_nlearnts(p.pSat), p.pSat.nDBreduces, p.nCexes, p.nObjAdded ); }
    Vec_VecFreeP( (Vec_Vec_t **)&p.vCores );
    Vec_VecFreeP( (Vec_Vec_t **)&p.vFrames );
    Vec_BitFreeP( &p.vSeenGla );
    Vec_IntFreeP( &p.vSeens );
    Vec_IntFreeP( &p.vOrder );
    Vec_IntFreeP( &p.vAddedNew );
    sat_solver2_delete( p.pSat );
    p.pBins = nil;
    p.pObjs = nil;
    p = nil;
}

/**Function*************************************************************

  Synopsis    [Returns the output literal.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vga_ManGetOutLit(p *Vta_Man_t, f int32) int32 {
    var pObj *Gia_Obj_t  = Gia_ManPo(p.pGia, 0);
    var pThis *Vta_Obj_t  = Vga_ManFind( p, Gia_ObjFaninId0p(p.pGia, pObj), f );
    assert(pThis != nil && pThis.fAdded, "pThis != nil && pThis.fAdded");
    if f == 0 && Gia_ObjIsRo(p.pGia, Gia_ObjFanin0(pObj)) && !Gia_ObjFaninC0(pObj) {
        return -Vta_ObjId(p, pThis); }
    return Abc_Var2Lit( Vta_ObjId(p, pThis), Gia_ObjFaninC0(pObj) );
} 

/**Function*************************************************************

  Synopsis    [Finds the set of clauses involved in the UNSAT core.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vta_ManUnsatCore(iLit int32, pSat *sat_solver2, nConfMax int32, fVerbose int32, piRetValue *int32, pnConfls *int32) *Vec_Int_t {
    var clk abctime  = Abc_Clock();
    var vCore *Vec_Int_t;
    var RetValue int32;
    var nConfPrev int32 = pSat.stats.conflicts;
    if piRetValue != nil {
        *piRetValue = 1; }
    // consider special case when PO points to the flop
    // this leads to immediate conflict in the first timeframe
    if iLit < 0 {
        vCore = Vec_IntAlloc( 1 );
        Vec_IntPush( vCore, -iLit );
        return vCore;
    }
    // solve the problem
    RetValue = sat_solver2_solve( pSat, &iLit, &iLit+1, (ABC_INT64_T)nConfMax, (ABC_INT64_T)0, (ABC_INT64_T)0, (ABC_INT64_T)0 );
    if pnConfls != nil {
        *pnConfls = int32(pSat.stats.conflicts) - nConfPrev; }
    if RetValue == l_Undef {
        if piRetValue != nil {
            *piRetValue = -1; }
        return nil;
    }
    if RetValue == l_True {
        if piRetValue != nil {
            *piRetValue = 0; }
        return nil;
    }
    if fVerbose != 0 {
//        Abc_Print( 1, "%6d", (int)pSat->stats.conflicts - nConfPrev );
//        Abc_Print( 1, "UNSAT after %7d conflicts.      ", pSat->stats.conflicts );
//        Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
    }
    assert(RetValue == l_False, "RetValue == l_False");
    // derive the UNSAT core
    clk = Abc_Clock();
    vCore = (Vec_Int_t *)Sat_ProofCore( pSat );
    if fVerbose != 0 {
//        Abc_Print( 1, "Core is %8d vars    (out of %8d).   ", Vec_IntSize(vCore), sat_solver2_nvars(pSat) );
//        Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
    }
    return vCore;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vta_ManAbsPrintFrame(p *Vta_Man_t, vCore *Vec_Int_t, nFrames int32, nConfls int32, nCexes int32, Time abctime, fVerbose int32) int32 {
    var pInfo *uint32;
    var pCountAll *int32 = nil;
    var pCountUni *int32 = nil;
    var i, iFrame, iObj, Entry, fChanges int32;
    // print info about frames
    if vCore != nil {
        // inlining of macro ABC_CALLOC
        pCountAll = ((int *) calloc((size_t)(nFrames + 1), sizeof(int)));
        // inlining of macro ABC_CALLOC
        pCountUni = ((int *) calloc((size_t)(nFrames + 1), sizeof(int)));
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(vCore)) && (((Entry) = Vec_IntEntry(vCore, i)), 1); i++  {
            iObj   = (Entry &  p.nObjMask);
            iFrame = (Entry >> p.nObjBits);
            assert(iFrame < nFrames, "iFrame < nFrames");
            pInfo  = (unsigned *)Vec_IntEntryP( p.vSeens, p.nWords * iObj );
            if !Abc_InfoHasBit(pInfo, iFrame) {
                Abc_InfoSetBit( pInfo, iFrame );
                pCountUni[iFrame+1]++;
                pCountUni[0]++;
                p.nSeenAll++;
            }
            pCountAll[iFrame+1]++;
            pCountAll[0]++;
            if !Vec_BitEntry(p.vSeenGla, iObj) {
                Vec_BitWriteEntry(p.vSeenGla, iObj, 1);
                p.nSeenGla++;
                fChanges = 1;
            }
        }
    }
    if !(fVerbose != 0) {
        pCountAll = nil;
        pCountUni = nil;
        return fChanges;
    }

    if Abc_FrameIsBatchMode() && !(vCore != nil) {
        return fChanges; }

//    Abc_Print( 1, "%5d%5d", pCountAll[0], pCountUni[0] ); 
    Abc_Print( 1, "%4d :", nFrames-1 );
    Abc_Print( 1, "%4d", Abc_MinInt(100, 100 * p.nSeenGla / (Gia_ManRegNum(p.pGia) + Gia_ManAndNum(p.pGia) + 1)) ); 
    Abc_Print( 1, "%6d", p.nSeenGla ); 
    Abc_Print( 1, "%4d", Abc_MinInt(100, 100 * p.nSeenAll / (p.nSeenGla * nFrames)) ); 
    Abc_Print( 1, "%8d", nConfls );
    if nCexes == 0 {
        Abc_Print( 1, "%5c", '-' ); 
    } else {
        Abc_Print( 1, "%5d", nCexes ); } 
//    Abc_Print( 1, " %9d", sat_solver2_nvars(p->pSat) ); 
    Abc_PrintInt( sat_solver2_nvars(p.pSat) );
    Abc_PrintInt( sat_solver2_nclauses(p.pSat) );
    Abc_PrintInt( sat_solver2_nlearnts(p.pSat) );
    if vCore == nil {
        Abc_Print( 1, "    ..." ); 
//        for ( k = 0; k < 7; k++ )
//            Abc_Print( 1, "     " );
        Abc_Print( 1, "%9.2f sec", 1.0*Time/int64(time.Second) );
        Abc_Print( 1, "%5.1f GB", (sat_solver2_memory_proof(p.pSat) + sat_solver2_memory(p.pSat, 0)) / (1<<30) );
        Abc_Print( 1, "\r" );
    
    } else {
        Abc_PrintInt( pCountAll[0] );
/*
        if ( nFrames > 7 )
        {
            for ( k = 0; k < 3; k++ )
                Abc_Print( 1, "%5d", pCountAll[k+1] ); 
            Abc_Print( 1, "  ..." );
            for ( k = nFrames-3; k < nFrames; k++ )
                Abc_Print( 1, "%5d", pCountAll[k+1] ); 
        }
        else
        {
            for ( k = 0; k < nFrames; k++ )
                Abc_Print( 1, "%5d", pCountAll[k+1] ); 
            for ( k = nFrames; k < 7; k++ )
                Abc_Print( 1, "     " );
        }
*/
        Abc_Print( 1, "%9.2f sec", 1.0*Time/int64(time.Second) );
        Abc_Print( 1, "%5.1f GB", (sat_solver2_memory_proof(p.pSat) + sat_solver2_memory(p.pSat, 0)) / (1<<30) );
        Abc_Print( 1, "\n" );
    }
    fflush( os.Stdout );

    if vCore != nil {
        pCountAll = nil;
        pCountUni = nil;
    }
    return fChanges;
}

/**Function*************************************************************

  Synopsis    [Adds clauses to the solver.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vga_ManAddClausesOne(p *Vta_Man_t, iObj int32, iFrame int32) { 
    var pThis0, pThis1 *Vta_Obj_t;
    var pObj *Gia_Obj_t  = Gia_ManObj( p.pGia, iObj );
    var pThis *Vta_Obj_t  = Vga_ManFindOrAdd( p, iObj, iFrame );
    var iThis0 int32;
    var iMainVar int32 = Vta_ObjId(p, pThis);
    assert(pThis.iObj == iObj && pThis.iFrame == iFrame, "pThis.iObj == iObj && pThis.iFrame == iFrame");
    if pThis.fAdded {
        return; }
    pThis.fAdded = 1;
    Vec_IntPush( p.vAddedNew, iMainVar );
    if Gia_ObjIsAnd(pObj) {
        pThis0 = Vga_ManFindOrAdd( p, Gia_ObjFaninId0p(p.pGia, pObj), iFrame );
        iThis0 = Vta_ObjId(p, pThis0);
        pThis1 = Vga_ManFindOrAdd( p, Gia_ObjFaninId1p(p.pGia, pObj), iFrame );
        sat_solver2_add_and( p.pSat, iMainVar, iThis0, Vta_ObjId(p, pThis1), 
            Gia_ObjFaninC0(pObj), Gia_ObjFaninC1(pObj), 0, iMainVar );
    
    } else if Gia_ObjIsRo(p.pGia, pObj) {
        if iFrame == 0 {
            if p.pPars.fUseTermVars {
                pThis0 = Vga_ManFindOrAdd( p, iObj, -1 );
                sat_solver2_add_constraint( p.pSat, iMainVar, Vta_ObjId(p, pThis0), 1, 0, iMainVar );
            
            } else {
                sat_solver2_add_const( p.pSat, iMainVar, 1, 0, iMainVar );
            }
        
        } else {
            pObj = Gia_ObjRoToRi( p.pGia, pObj );
            pThis0 = Vga_ManFindOrAdd( p, Gia_ObjFaninId0p(p.pGia, pObj), iFrame-1 );
            sat_solver2_add_buffer( p.pSat, iMainVar, Vta_ObjId(p, pThis0), Gia_ObjFaninC0(pObj), 0, iMainVar );  
        }
    
    } else if Gia_ObjIsConst0(pObj) {
        sat_solver2_add_const( p.pSat, iMainVar, 1, 0, iMainVar );
    
    } else { //if ( !Gia_ObjIsPi(p->pGia, pObj) )
        assert(false, "0"); }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vga_ManLoadSlice(p *Vta_Man_t, vOne *Vec_Int_t, Lift int32) {
    var i, Entry int32;
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vOne)) && (((Entry) = Vec_IntEntry(vOne, i)), 1); i++  {
        Vga_ManAddClausesOne( p, Entry & p.nObjMask, (Entry >> p.nObjBits) + Lift );
    }
    sat_solver2_simplify( p.pSat );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vga_ManPrintCore(p *Vta_Man_t, vCore *Vec_Int_t, Lift int32) {
    var i, Entry, iObj, iFrame int32;
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vCore)) && (((Entry) = Vec_IntEntry(vCore, i)), 1); i++  {
        iObj   = (Entry &  p.nObjMask);
        iFrame = (Entry >> p.nObjBits);
        Abc_Print( 1, "%d*%d ", iObj, iFrame+Lift );
    }
    Abc_Print( 1, "\n" );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vga_ManRollBack(p *Vta_Man_t, nObjOld int32) {
    var pThis *Vta_Obj_t   = p.pObjs + nObjOld;
    var pLimit *Vta_Obj_t  = p.pObjs + p.nObjs;
    var i, Entry int32;
    for  ; pThis < pLimit; pThis++  {
        Vga_ManDelete( p, pThis.iObj, pThis.iFrame );
    }
    memset( p.pObjs + nObjOld, 0, sizeof(Vta_Obj_t) * (p.nObjs - nObjOld) );
    p.nObjs = nObjOld;
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(p.vAddedNew)) && (((Entry) = Vec_IntEntry(p.vAddedNew, i)), 1); i++  {
        if Entry < p.nObjs {
            pThis = Vta_ManObj(p, Entry);
            assert(pThis.fAdded == 1, "pThis.fAdded == 1");
            pThis.fAdded = 0;
        }
    }
}

/**Function*************************************************************

  Synopsis    [Send abstracted model or send cancel.]

  Description [Counter-example will be sent automatically when &vta terminates.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_VtaSendAbsracted(p *Vta_Man_t, fVerbose int32) {
    var pAbs *Gia_Man_t;
    assert(Abc_FrameIsBridgeMode(), "Abc_FrameIsBridgeMode()");
//    if ( fVerbose )
//        Abc_Print( 1, "Sending abstracted model...\n" );
    // create obj classes
    Vec_IntFreeP( &p.pGia.vObjClasses );
    p.pGia.vObjClasses = Gia_VtaFramesToAbs( (Vec_Vec_t *)p.vCores );
    // create gate classes
    Vec_IntFreeP( &p.pGia.vGateClasses );
    p.pGia.vGateClasses = Gia_VtaConvertToGla( p.pGia, p.pGia.vObjClasses );
    Vec_IntFreeP( &p.pGia.vObjClasses );
    // create abstrated model
    pAbs = Gia_ManDupAbsGates( p.pGia, p.pGia.vGateClasses );
    Vec_IntFreeP( &p.pGia.vGateClasses );
    // send it out
    Gia_ManToBridgeAbsNetlist( os.Stdout, pAbs, BRIDGE_ABS_NETLIST );
    Gia_ManStop( pAbs );
}
func Gia_VtaSendCancel(p *Vta_Man_t, fVerbose int32) {
    extern int Gia_ManToBridgeBadAbs( FILE * pFile );
    assert(Abc_FrameIsBridgeMode(), "Abc_FrameIsBridgeMode()");
//    if ( fVerbose )
//        Abc_Print( 1, "Cancelling previously sent model...\n" );
    Gia_ManToBridgeBadAbs( os.Stdout );
}

/**Function*************************************************************

  Synopsis    [Send abstracted model or send cancel.]

  Description [Counter-example will be sent automatically when &vta terminates.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_VtaDumpAbsracted(p *Vta_Man_t, fVerbose int32) {
    var pFileNameDef string  = "vabs.aig";
    var pFileName string;
    if p.pPars.pFileVabs {
        pFileName = p.pPars.pFileVabs;
    } else {
        pFileName = pFileNameDef;
    }
    var pAbs *Gia_Man_t;
    if fVerbose != 0 {
        Abc_Print( 1, "Dumping abstracted model into file \"%s\"...\n", pFileName ); }
    // create obj classes
    Vec_IntFreeP( &p.pGia.vObjClasses );
    p.pGia.vObjClasses = Gia_VtaFramesToAbs( (Vec_Vec_t *)p.vCores );
    // create gate classes
    Vec_IntFreeP( &p.pGia.vGateClasses );
    p.pGia.vGateClasses = Gia_VtaConvertToGla( p.pGia, p.pGia.vObjClasses );
    Vec_IntFreeP( &p.pGia.vObjClasses );
    // create abstrated model
    pAbs = Gia_ManDupAbsGates( p.pGia, p.pGia.vGateClasses );
    Vec_IntFreeP( &p.pGia.vGateClasses );
    // send it out
    Gia_AigerWrite( pAbs, pFileName, 0, 0, 0 );
    Gia_ManStop( pAbs );
}

    
/**Function*************************************************************

  Synopsis    [Print memory report.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_VtaPrintMemory(p *Vta_Man_t) {
    var memTot float64;
    var memAig float64  = Gia_ManObjNum(p.pGia) * sizeof(Gia_Obj_t);
    var memSat float64  = sat_solver2_memory( p.pSat, 1 );
    var memPro float64  = sat_solver2_memory_proof( p.pSat );
    var memMap float64  = p.nObjsAlloc * sizeof(Vta_Obj_t) + p.nBins * sizeof(int);
    var memOth float64  = sizeof(Vta_Man_t);
    memOth += Vec_IntCap(p.vOrder) * sizeof(int);
    memOth += Vec_VecMemoryInt( (Vec_Vec_t *)p.vFrames );
    memOth += Vec_BitCap(p.vSeenGla) * sizeof(int);
    memOth += Vec_VecMemoryInt( (Vec_Vec_t *)p.vCores );
    memOth += Vec_IntCap(p.vAddedNew) * sizeof(int);
    memTot = memAig + memSat + memPro + memMap + memOth;
    // inlining of macro ABC_PRMP
    Abc_Print(1, "%s =", ("Memory: AIG     "));
Abc_Print(1, "%10.3f MB (%6.2f %%)\n",  (1.0*(float64((memAig)))/(1<<20)), ((float64((memTot)))? 100.0*(float64((memAig)))/(float64((memTot))) : 0.0) );;
    // inlining of macro ABC_PRMP
    Abc_Print(1, "%s =", ("Memory: SAT     "));
Abc_Print(1, "%10.3f MB (%6.2f %%)\n",  (1.0*(float64((memSat)))/(1<<20)), ((float64((memTot)))? 100.0*(float64((memSat)))/(float64((memTot))) : 0.0) );;
    // inlining of macro ABC_PRMP
    Abc_Print(1, "%s =", ("Memory: Proof   "));
Abc_Print(1, "%10.3f MB (%6.2f %%)\n",  (1.0*(float64((memPro)))/(1<<20)), ((float64((memTot)))? 100.0*(float64((memPro)))/(float64((memTot))) : 0.0) );;
    // inlining of macro ABC_PRMP
    Abc_Print(1, "%s =", ("Memory: Map     "));
Abc_Print(1, "%10.3f MB (%6.2f %%)\n",  (1.0*(float64((memMap)))/(1<<20)), ((float64((memTot)))? 100.0*(float64((memMap)))/(float64((memTot))) : 0.0) );;
    // inlining of macro ABC_PRMP
    Abc_Print(1, "%s =", ("Memory: Other   "));
Abc_Print(1, "%10.3f MB (%6.2f %%)\n",  (1.0*(float64((memOth)))/(1<<20)), ((float64((memTot)))? 100.0*(float64((memOth)))/(float64((memTot))) : 0.0) );;
    // inlining of macro ABC_PRMP
    Abc_Print(1, "%s =", ("Memory: TOTAL   "));
Abc_Print(1, "%10.3f MB (%6.2f %%)\n",  (1.0*(float64((memTot)))/(1<<20)), ((float64((memTot)))? 100.0*(float64((memTot)))/(float64((memTot))) : 0.0) );;
}


/**Function*************************************************************

  Synopsis    [Collect nodes/flops involved in different timeframes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_VtaPerformInt(pAig *Gia_Man_t, pPars *Abs_Par_t) int32 {
    var p *Vta_Man_t;
    var vCore *Vec_Int_t;
    var pCex *Abc_Cex_t  = nil;
    var i int32;
    var f int32;
    var nConfls int32;
    var Status int32;
    var nObjOld int32;
    var RetValue int32 = -1;
    var nCountNoChange int32;
    var fOneIsSent int32;
    var clk abctime = Abc_Clock();
    var clk2 abctime;
    // preconditions
    assert(Gia_ManPoNum(pAig) == 1, "Gia_ManPoNum(pAig) == 1");
    assert(pPars.nFramesMax == 0 || pPars.nFramesStart <= pPars.nFramesMax, "pPars.nFramesMax == 0 || pPars.nFramesStart <= pPars.nFramesMax");
    if Gia_ObjIsConst0(Gia_ObjFanin0(Gia_ManPo(pAig,0))) {
        if !Gia_ObjFaninC0(Gia_ManPo(pAig,0)) {
            fmt.Printf("Sequential miter is trivially UNSAT.\n");
            return 1;
        } 
        pAig.pCexSeq = nil;
        pAig.pCexSeq = Abc_CexMakeTriv( Gia_ManRegNum(pAig), Gia_ManPiNum(pAig), 1, 0 );
        fmt.Printf("Sequential miter is trivially SAT.\n");
        return 0;
    }

    // compute intial abstraction
    if pAig.vObjClasses == nil {
        pAig.vObjClasses = Vec_IntAlloc( 5 );
        Vec_IntPush( pAig.vObjClasses, 1 );
        Vec_IntPush( pAig.vObjClasses, 3 );
        Vec_IntPush( pAig.vObjClasses, 4 );
        Vec_IntPush( pAig.vObjClasses, Gia_ObjFaninId0p(pAig, Gia_ManPo(pAig, 0)) );
    }
    // start the manager
    p = Vga_ManStart( pAig, pPars );
    // set runtime limit
    if p.pPars.nTimeOut {
        sat_solver2_set_runtime_limit( p.pSat, p.pPars.nTimeOut * int64(time.Second) + Abc_Clock() ); }
    // perform initial abstraction
    if p.pPars.fVerbose {
        Abc_Print( 1, "Running variable-timeframe abstraction (VTA) with the following parameters:\n" );
        Abc_Print( 1, "FramePast = %d  FrameMax = %d  ConfMax = %d  Timeout = %d  RatioMin = %d %%\n", 
            pPars.nFramesPast, pPars.nFramesMax, pPars.nConfLimit, pPars.nTimeOut, pPars.nRatioMin );
        Abc_Print( 1, "LearnStart = %d  LearnDelta = %d  LearnRatio = %d %%.\n", 
            pPars.nLearnedStart, pPars.nLearnedDelta, pPars.nLearnedPerce );
//        Abc_Print( 1, "Frame   %%   Abs   %%   Confl  Cex    SatVar   Core   F0   F1   F2  ...\n" );
        Abc_Print( 1, " Frame   %%   Abs   %%   Confl  Cex   Vars   Clas   Lrns   Core     Time      Mem\n" );
    }
    assert(Vec_PtrSize(p.vFrames) > 0, "Vec_PtrSize(p.vFrames) > 0");
    for  f = i = 0; !p.pPars.nFramesMax || f < p.pPars.nFramesMax; f++  {
        var nConflsBeg int32  = sat_solver2_nconflicts(p.pSat);
        p.pPars.iFrame = f;
        // realloc storage for abstraction marks
        if f == p.nWords * 32 {
            p.nWords = Vec_IntDoubleWidth( p.vSeens, p.nWords ); }

        // create bookmark to be used for rollback
        nObjOld = p.nObjs;
        sat_solver2_bookmark( p.pSat );
        Vec_IntClear( p.vAddedNew );

        // load new timeframe
        Vga_ManAddClausesOne( p, 0, f );
        if f < Vec_PtrSize(p.vFrames) {
            Vga_ManLoadSlice( p, (Vec_Int_t *)Vec_PtrEntry(p.vFrames, f), 0 );
        } else {
            for  i = 1; i <= Abc_MinInt(p.pPars.nFramesPast, f); i++  {
                Vga_ManLoadSlice( p, (Vec_Int_t *)Vec_PtrEntry(p.vCores, f-i), i );
            }
        }

        // iterate as long as there are counter-examples
        for  i = 0; ; i++  { 
            clk2 = Abc_Clock();
            vCore = Vta_ManUnsatCore( Vga_ManGetOutLit(p, f), p.pSat, pPars.nConfLimit, pPars.fVerbose, &Status, &nConfls );
            assert((vCore != nil) == (Status == 1), "(vCore != nil) == (Status == 1)");
            if Status == -1 { // resource limit is reached
            
                Vga_ManRollBack( p, nObjOld );
                goto finish;
            }
            // check timeout
            if p.pSat.nRuntimeLimit && Abc_Clock() > p.pSat.nRuntimeLimit {
                Vga_ManRollBack( p, nObjOld );
                goto finish;
            }
            if vCore != nil {
                p.timeUnsat += Abc_Clock() - clk2;
                break;
            }
            p.timeSat += Abc_Clock() - clk2;
            assert(Status == 0, "Status == 0");
            p.nCexes++;
            // perform the refinement
            clk2 = Abc_Clock();
            pCex = Vta_ManRefineAbstraction( p, f );
            p.timeCex += Abc_Clock() - clk2;
            if pCex != nil {
                goto finish; }
            // print the result (do not count it towards change)
            Vta_ManAbsPrintFrame( p, nil, f+1, sat_solver2_nconflicts(p.pSat)-nConflsBeg, i, Abc_Clock() - clk, p.pPars.fVerbose );
        }
        assert(Status == 1, "Status == 1");
        // valid core is obtained
        Vta_ManUnsatCoreRemap( p, vCore );
        Vec_IntSort( vCore, 1 );
        // update the SAT solver
        sat_solver2_rollback( p.pSat );
        // update storage
        Vga_ManRollBack( p, nObjOld );
        // load this timeframe
        Vga_ManLoadSlice( p, vCore, 0 );
        Vec_IntFree( vCore );

        // run SAT solver
        clk2 = Abc_Clock();
        vCore = Vta_ManUnsatCore( Vga_ManGetOutLit(p, f), p.pSat, pPars.nConfLimit, p.pPars.fVerbose, &Status, &nConfls );
        p.timeUnsat += Abc_Clock() - clk2;
        assert((vCore != nil) == (Status == 1), "(vCore != nil) == (Status == 1)");
        if Status == -1 { // resource limit is reached
            break; }
        if Status == 0 {
            Vta_ManSatVerify( p );
            // make sure, there was no initial abstraction (otherwise, it was invalid)
            assert(pAig.vObjClasses == nil && f < p.pPars.nFramesStart, "pAig.vObjClasses == nil && f < p.pPars.nFramesStart");
            pCex = Vga_ManDeriveCex( p );
            break;
        }
        // add the core
        Vta_ManUnsatCoreRemap( p, vCore );
        // add in direct topological order
        Vec_IntSort( vCore, 1 );
        Vec_PtrPush( p.vCores, vCore );
        // print the result
        if Vta_ManAbsPrintFrame( p, vCore, f+1, sat_solver2_nconflicts(p.pSat)-nConflsBeg, i, Abc_Clock() - clk, p.pPars.fVerbose ) != 0 {
            // reset the counter of frames without change
            nCountNoChange = 1;
            p.pPars.nFramesNoChange = 0;
        
        } else if nCountNoChange++; nCountNoChange == 2 { // time to send
        
            p.pPars.nFramesNoChange++;
            if Abc_FrameIsBridgeMode() {
                // cancel old one if it was sent
                if fOneIsSent != 0 {
                    Gia_VtaSendCancel( p, pPars.fVerbose ); }
                // send new one 
                Gia_VtaSendAbsracted( p, pPars.fVerbose );
                fOneIsSent = 1;
            }
        }
        // dump the model
        if p.pPars.fDumpVabs && (f & 1) {
            var Command string // was [1000]byte in C

            Abc_FrameSetStatus( -1 );
            Abc_FrameSetCex( nil );
            Abc_FrameSetNFrames( f+1 );
            Command = fmt.Sprintf("write_status %s", Extra_FileNameGenericAppend((char *)(p.pPars.pFileVabs ? p.pPars.pFileVabs : "vtabs.aig"), ".status"));
            Cmd_CommandExecute( Abc_FrameGetGlobalFrame(), Command );
            Gia_VtaDumpAbsracted( p, pPars.fVerbose );
        }
        // check if the number of objects is below limit
        if p.nSeenGla >= Gia_ManCandNum(pAig) * (100-pPars.nRatioMin) / 100 {
            Status = -1;
            break;
        }
    }
finish:
    // analize the results
    if pCex == nil {
        if p.pPars.fVerbose && Status == -1 {
            fmt.Printf("\n"); }
        if Vec_PtrSize(p.vCores) == 0 {
            Abc_Print( 1, "Abstraction is not produced because first frame is not solved.  " );
        } else {
            assert(Vec_PtrSize(p.vCores) > 0, "Vec_PtrSize(p.vCores) > 0");
//            if ( pAig->vObjClasses != NULL )
//                Abc_Print( 1, "Replacing the old abstraction by a new one.\n" );
            Vec_IntFreeP( &pAig.vObjClasses );
            pAig.vObjClasses = Gia_VtaFramesToAbs( (Vec_Vec_t *)p.vCores );
            if Status == -1 {
                if p.pPars.nTimeOut && Abc_Clock() >= p.pSat.nRuntimeLimit { 
                    Abc_Print( 1, "Timeout %d sec in frame %d with a %d-stable abstraction.    ", p.pPars.nTimeOut, f, p.pPars.nFramesNoChange );
                } else if pPars.nConfLimit && sat_solver2_nconflicts(p.pSat) >= pPars.nConfLimit {
                    Abc_Print( 1, "Exceeded %d conflicts in frame %d with a %d-stable abstraction.  ", pPars.nConfLimit, f, p.pPars.nFramesNoChange );
                } else if p.nSeenGla >= Gia_ManCandNum(pAig) * (100-pPars.nRatioMin) / 100 {
                    Abc_Print( 1, "The ratio of abstracted objects is less than %d %% in frame %d.  ", pPars.nRatioMin, f );
                } else {
                    Abc_Print( 1, "Abstraction stopped for unknown reason in frame %d.  ", f ); }
            
            } else {
                p.pPars.iFrame++;
                Abc_Print( 1, "VTA completed %d frames with a %d-stable abstraction.  ", f, p.pPars.nFramesNoChange );
            }
        }
    
    } else {
        if p.pPars.fVerbose {
            fmt.Printf("\n"); }
        p.pGia.pCexSeq = nil;
        p.pGia.pCexSeq = pCex;
        if !Gia_ManVerifyCex( p.pGia, pCex, 0 ) {
            Abc_Print( 1, "    Gia_VtaPerform(): CEX verification has failed!\n" ); }
        Abc_Print( 1, "Counter-example detected in frame %d.  ", f );
        p.pPars.iFrame = pCex.iFrame - 1;
        Vec_IntFreeP( &pAig.vObjClasses );
        RetValue = 0;
    }
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );

    if p.pPars.fVerbose {
        p.timeOther = (Abc_Clock() - clk) - p.timeUnsat - p.timeSat - p.timeCex;
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("Runtime: Solver UNSAT"));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeUnsat)))/(float64(int64(time.Second))), (float64((Abc_Clock() - clk)))? 100.0*(float64((p.timeUnsat)))/(float64((Abc_Clock() - clk))) : 0.0);;
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("Runtime: Solver SAT  "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeSat)))/(float64(int64(time.Second))), (float64((Abc_Clock() - clk)))? 100.0*(float64((p.timeSat)))/(float64((Abc_Clock() - clk))) : 0.0);;
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("Runtime: Refinement  "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeCex)))/(float64(int64(time.Second))), (float64((Abc_Clock() - clk)))? 100.0*(float64((p.timeCex)))/(float64((Abc_Clock() - clk))) : 0.0);;
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("Runtime: Other       "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeOther)))/(float64(int64(time.Second))), (float64((Abc_Clock() - clk)))? 100.0*(float64((p.timeOther)))/(float64((Abc_Clock() - clk))) : 0.0);;
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("Runtime: TOTAL       "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))), (float64((Abc_Clock() - clk)))? 100.0*(float64((Abc_Clock() - clk)))/(float64((Abc_Clock() - clk))) : 0.0);;
        Gia_VtaPrintMemory( p );
    }

    Vga_ManStop( p );
    fflush( os.Stdout );
    return RetValue;
}

/**Function*************************************************************

  Synopsis    [Collect nodes/flops involved in different timeframes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_VtaPerform(pAig *Gia_Man_t, pPars *Abs_Par_t) int32 {
    var RetValue int32  = -1;
    if pAig.vObjClasses == nil && pPars.fUseRollback {
        var nFramesMaxOld int32  = pPars.nFramesMax;
        pPars.nFramesMax = pPars.nFramesStart;
        RetValue = Gia_VtaPerformInt( pAig, pPars );
        pPars.nFramesMax = nFramesMaxOld;
    }
    if RetValue == 0 {
        return RetValue; }
    return Gia_VtaPerformInt( pAig, pPars );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

