package gia

import (
	"fmt"
	"os"
	"time"
)

/**CFile****************************************************************

  FileName    [giaIso.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [Graph isomorphism.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaIso.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"

//ABC_NAMESPACE_IMPL_START
 

const ISO_MASK = 0xFF
static unsigned int s_256Primes[ISO_MASK+1] = 
{
    0x984b6ad9,0x18a6eed3,0x950353e2,0x6222f6eb,0xdfbedd47,0xef0f9023,0xac932a26,0x590eaf55,
    0x97d0a034,0xdc36cd2e,0x22736b37,0xdc9066b0,0x2eb2f98b,0x5d9c7baf,0x85747c9e,0x8aca1055,
    0x50d66b74,0x2f01ae9e,0xa1a80123,0x3e1ce2dc,0xebedbc57,0x4e68bc34,0x855ee0cf,0x17275120,
    0x2ae7f2df,0xf71039eb,0x7c283eec,0x70cd1137,0x7cf651f3,0xa87bfa7a,0x14d87f02,0xe82e197d,
    0x8d8a5ebe,0x1e6a15dc,0x197d49db,0x5bab9c89,0x4b55dea7,0x55dede49,0x9a6a8080,0xe5e51035,
    0xe148d658,0x8a17eb3b,0xe22e4b38,0xe5be2a9a,0xbe938cbb,0x3b981069,0x7f9c0c8e,0xf756df10,
    0x8fa783f7,0x252062ce,0x3dc46b4b,0xf70f6432,0x3f378276,0x44b137a1,0x2bf74b77,0x04892ed6,
    0xfd318de1,0xd58c235e,0x94c6d25b,0x7aa5f218,0x35c9e921,0x5732fbbb,0x06026481,0xf584a44f,
    0x946e1b5f,0x8463d5b2,0x4ebca7b2,0x54887b15,0x08d1e804,0x5b22067d,0x794580f6,0xb351ea43,
    0xbce555b9,0x19ae2194,0xd32f1396,0x6fc1a7f1,0x1fd8a867,0x3a89fdb0,0xea49c61c,0x25f8a879,
    0xde1e6437,0x7c74afca,0x8ba63e50,0xb1572074,0xe4655092,0xdb6f8b1c,0xc2955f3c,0x327f85ba,
    0x60a17021,0x95bd261d,0xdea94f28,0x04528b65,0xbe0109cc,0x26dd5688,0x6ab2729d,0xc4f029ce,
    0xacf7a0be,0x4c912f55,0x34c06e65,0x4fbb938e,0x1533fb5f,0x03da06bd,0x48262889,0xc2523d7d,
    0x28a71d57,0x89f9713a,0xf574c551,0x7a99deb5,0x52834d91,0x5a6f4484,0xc67ba946,0x13ae698f,
    0x3e390f34,0x34fc9593,0x894c7932,0x6cf414a3,0xdb7928ab,0x13a3b8a3,0x4b381c1d,0xa10b54cb,
    0x55359d9d,0x35a3422a,0x58d1b551,0x0fd4de20,0x199eb3f4,0x167e09e2,0x3ee6a956,0x5371a7fa,
    0xd424efda,0x74f521c5,0xcb899ff6,0x4a42e4f4,0x747917b6,0x4b08df0b,0x090c7a39,0x11e909e4,
    0x258e2e32,0xd9fad92d,0x48fe5f69,0x0545cde6,0x55937b37,0x9b4ae4e4,0x1332b40e,0xc3792351,
    0xaff982ef,0x4dba132a,0x38b81ef1,0x28e641bf,0x227208c1,0xec4bbe37,0xc4e1821c,0x512c9d09,
    0xdaef1257,0xb63e7784,0x043e04d7,0x9c2cea47,0x45a0e59a,0x281315ca,0x849f0aac,0xa4071ed3,
    0x0ef707b3,0xfe8dac02,0x12173864,0x471f6d46,0x24a53c0a,0x35ab9265,0xbbf77406,0xa2144e79,
    0xb39a884a,0x0baf5b6d,0xcccee3dd,0x12c77584,0x2907325b,0xfd1adcd2,0xd16ee972,0x345ad6c1,
    0x315ebe66,0xc7ad2b8d,0x99e82c8d,0xe52da8c8,0xba50f1d3,0x66689cd8,0x2e8e9138,0x43e15e74,
    0xf1ced14d,0x188ec52a,0xe0ef3cbb,0xa958aedc,0x4107a1bc,0x5a9e7a3e,0x3bde939f,0xb5b28d5a,
    0x596fe848,0xe85ad00c,0x0b6b3aae,0x44503086,0x25b5695c,0xc0c31dcd,0x5ee617f0,0x74d40c3a,
    0xd2cb2b9f,0x1e19f5fa,0x81e24faf,0xa01ed68f,0xcee172fc,0x7fdf2e4d,0x002f4774,0x664f82dd,
    0xc569c39a,0xa2d4dcbe,0xaadea306,0xa4c947bf,0xa413e4e3,0x81fb5486,0x8a404970,0x752c980c,
    0x98d1d881,0x5c932c1e,0xeee65dfb,0x37592cdd,0x0fd4e65b,0xad1d383f,0x62a1452f,0x8872f68d,
    0xb58c919b,0x345c8ee3,0xb583a6d6,0x43d72cb3,0x77aaa0aa,0xeb508242,0xf2db64f8,0x86294328,
    0x82211731,0x1239a9d5,0x673ba5de,0xaf4af007,0x44203b19,0x2399d955,0xa175cd12,0x595928a7,
    0x6918928b,0xde3126bb,0x6c99835c,0x63ba1fa2,0xdebbdff0,0x3d02e541,0xd6f7aac6,0xe80b4cd0,
    0xd0fa29f1,0x804cac5e,0x2c226798,0x462f624c,0xad05b377,0x22924fcd,0xfbea205c,0x1b47586d
};

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

         
type Gia_IsoMan_t struct {
    pGia       *Gia_Man_t;
    nObjs              int32;
    nUniques              int32;
    nSingles              int32;
    nEntries              int32;
    // internal data
    pLevels             *int32;
    pUniques             *int32;
    pStoreW            *word;
    pStoreU        *uint32;
    // equivalence classes
    vLevCounts       *Vec_Int_t;
    vClasses       *Vec_Int_t;
    vClasses2       *Vec_Int_t;
    // statistics 
    timeStart          abctime;
    timeSim          abctime;
    timeRefine          abctime;
    timeSort          abctime;
    timeOther          abctime;
    timeTotal          abctime;
}

func Gia_IsoGetValue(p *Gia_IsoMan_t, i int32) uint32 { return uint32((p.pStoreW[i]));       }
func Gia_IsoGetItem(p *Gia_IsoMan_t, i int32) uint32 { return uint32((p.pStoreW[i] >> 32)); }

func Gia_IsoSetValue(p *Gia_IsoMan_t, i int32, v uint32) { ((unsigned *)(p.pStoreW + i))[0] = v;  }
func Gia_IsoSetItem(p *Gia_IsoMan_t, i int32, v uint32) { ((unsigned *)(p.pStoreW + i))[1] = v;  }

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_IsoManStart(pGia *Gia_Man_t) *Gia_IsoMan_t {
    
    // inlining of macro ABC_CALLOC
    p := &Gia_IsoMan_t{};
    p.pGia      = pGia;
    p.nObjs     = Gia_ManObjNum( pGia );
    p.nUniques  = 1;
    p.nEntries  = p.nObjs;
    // internal data
    // inlining of macro ABC_CALLOC
    p.pLevels   = ((int *) calloc((size_t)(p.nObjs), sizeof(int)));
    // inlining of macro ABC_CALLOC
    p.pUniques  = ((int *) calloc((size_t)(p.nObjs), sizeof(int)));
    // inlining of macro ABC_CALLOC
    p.pStoreW   = ((word *) calloc((size_t)(p.nObjs), sizeof(word)));
    // class representation
    p.vClasses  = Vec_IntAlloc( p.nObjs/4 );
    p.vClasses2 = Vec_IntAlloc( p.nObjs/4 );
    return p;
}
func Gia_IsoManStop(p *Gia_IsoMan_t) {
    // class representation
    Vec_IntFree( p.vClasses );
    Vec_IntFree( p.vClasses2 );
    // internal data
    // inlining of macro ABC_FREE
    ((p.pLevels) ? (free((char *) (p.pLevels)), (p.pLevels) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pUniques) ? (free((char *) (p.pUniques)), (p.pUniques) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pStoreW) ? (free((char *) (p.pStoreW)), (p.pStoreW) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}
func Gia_IsoManTransferUnique(p *Gia_IsoMan_t) {
    var pObj *Gia_Obj_t;
    var i int32;
    // copy unique numbers into the nodes
    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < p.pGia.nObjs) && ((pObj) = Gia_ManObj(p.pGia, i)); i++  {
        pObj.Value = p.pUniques[i];
    }
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_IsoPrintClasses(p *Gia_IsoMan_t) {
    var fVerbose int32;
    var i, k, iBegin, nSize int32;
    fmt.Printf("The total of %d classes:\n", Vec_IntSize(p.vClasses)/2);
    // inlining of macro Vec_IntForEachEntryDouble
    for  i = 0; (i+1 < Vec_IntSize(p.vClasses)) && (((iBegin) = Vec_IntEntry(p.vClasses, i)), 1) && (((nSize) = Vec_IntEntry(p.vClasses, i+1)), 1); i += 2  {
        fmt.Printf("%5d : (%3d,%3d)  ", i/2, iBegin, nSize);
        if fVerbose != 0 {
            fmt.Printf("{");
            for  k = 0; k < nSize; k++  {
                fmt.Printf(" %3d,%08x", Gia_IsoGetItem(p, iBegin+k), Gia_IsoGetValue(p, iBegin+k));
            }
            fmt.Printf(" }");
        }
        fmt.Printf("\n");
    }
}
func Gia_IsoPrint(p *Gia_IsoMan_t, Iter int32, Time abctime) {
    fmt.Printf("Iter %4d :  ", Iter);
    fmt.Printf("Entries =%8d.  ", p.nEntries);
//    printf( "Classes =%8d.  ", Vec_IntSize(p->vClasses)/2 );
    fmt.Printf("Uniques =%8d.  ", p.nUniques);
    fmt.Printf("Singles =%8d.  ", p.nSingles);
    fmt.Printf("%9.2f sec", float32((Time))/float32((int64(time.Second))));
    fmt.Printf("\n");
    fflush( os.Stdout );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_IsoPrepare(p *Gia_IsoMan_t) {
    var pObj *Gia_Obj_t;
    var pLevBegins, pLevSizes *int32;
    var i, iObj, MaxLev int32;
    // assign levels
    p.pLevels[0] = 0;
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.pGia.vCis)) && ((pObj) = Gia_ManCi(p.pGia, i)); i++  {
        p.pLevels[Gia_ObjId(p.pGia, pObj)] = 0;
    }
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.pGia.nObjs) && ((pObj) = Gia_ManObj(p.pGia, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        p.pLevels[i] = 1 + Abc_MaxInt( p.pLevels[Gia_ObjFaninId0(pObj, i)], p.pLevels[Gia_ObjFaninId1(pObj, i)] ); }
    }
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.pGia.vCos)) && ((pObj) = Gia_ManCo(p.pGia, i)); i++  {
        iObj = Gia_ObjId(p.pGia, pObj);
        p.pLevels[iObj] = 1 + p.pLevels[Gia_ObjFaninId0(pObj, iObj)]; // "1 +" is different!
        MaxLev = Abc_MaxInt( MaxLev, p.pLevels[Gia_ObjId(p.pGia, pObj)] );
    }

    // count nodes on each level
    // inlining of macro ABC_CALLOC
    pLevSizes = ((int *) calloc((size_t)(MaxLev+1), sizeof(int)));
    for  i = 1; i < p.nObjs; i++  {
        pLevSizes[p.pLevels[i]]++;
    }
    // start classes
    Vec_IntClear( p.vClasses );
    Vec_IntPush( p.vClasses, 0 );
    Vec_IntPush( p.vClasses, 1 );
    // find beginning of each level
    // inlining of macro ABC_CALLOC
    pLevBegins = ((int *) calloc((size_t)(MaxLev+2), sizeof(int)));
    pLevBegins[0] = 1;
    for  i = 0; i <= MaxLev; i++  {
        assert(pLevSizes[i] > 0, "pLevSizes[i] > 0"); // we do not allow AIG with a const node and no PIs
        Vec_IntPush( p.vClasses, pLevBegins[i] );
        Vec_IntPush( p.vClasses, pLevSizes[i] );
        pLevBegins[i+1] = pLevBegins[i] + pLevSizes[i];
    }
    assert(pLevBegins[MaxLev+1] == p.nObjs, "pLevBegins[MaxLev+1] == p.nObjs");
    // put them into the structure
    for  i = 1; i < p.nObjs; i++  {
        Gia_IsoSetItem( p, pLevBegins[p.pLevels[i]]++, i );
    }
    // inlining of macro ABC_FREE
    ((pLevBegins) ? (free((char *) (pLevBegins)), (pLevBegins) = 0) : 0);
    // inlining of macro ABC_FREE
    ((pLevSizes) ? (free((char *) (pLevSizes)), (pLevSizes) = 0) : 0);
/*
    // print the results
    for ( i = 0; i < p->nObjs; i++ )
        printf( "%3d : (%d,%d)\n", i, Gia_IsoGetItem(p, i), Gia_IsoGetValue(p, i) );
    printf( "\n" );
*/
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_IsoAssignUnique(p *Gia_IsoMan_t) {
    var i, iBegin, nSize int32;
    p.nSingles = 0;
    Vec_IntClear( p.vClasses2 );
    // inlining of macro Vec_IntForEachEntryDouble
    for  i = 0; (i+1 < Vec_IntSize(p.vClasses)) && (((iBegin) = Vec_IntEntry(p.vClasses, i)), 1) && (((nSize) = Vec_IntEntry(p.vClasses, i+1)), 1); i += 2  {
        if nSize == 1 {
            assert(p.pUniques[Gia_IsoGetItem(p, iBegin)] == 0, "p.pUniques[Gia_IsoGetItem(p, iBegin)] == 0");
            p.pUniques[Gia_IsoGetItem(p, iBegin)] = p.nUniques++;
            p.nSingles++;
        
        } else {
            Vec_IntPush( p.vClasses2, iBegin );
            Vec_IntPush( p.vClasses2, nSize );
        }
    }
    // inlining of macro ABC_SWAP
    { var t *Vec_Int_t  = p.vClasses; p.vClasses = p.vClasses2; p.vClasses2 = t; };
    p.nEntries -= p.nSingles;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_IsoSort(p *Gia_IsoMan_t) int32 {
    var pObj, pObj0 *Gia_Obj_t;
    var i, k, fSameValue, iBegin, iBeginOld, nSize, nSizeNew int32;
    var fRefined int32;
    var clk abctime;

    // go through the equiv classes
    p.nSingles = 0;
    Vec_IntClear( p.vClasses2 );
    // inlining of macro Vec_IntForEachEntryDouble
    for  i = 0; (i+1 < Vec_IntSize(p.vClasses)) && (((iBegin) = Vec_IntEntry(p.vClasses, i)), 1) && (((nSize) = Vec_IntEntry(p.vClasses, i+1)), 1); i += 2  {
        assert(nSize > 1, "nSize > 1");
        fSameValue = 1;
        pObj0 = Gia_ManObj( p.pGia, Gia_IsoGetItem(p,iBegin) );
        for  k = 0; k < nSize; k++  {
            pObj = Gia_ManObj( p.pGia, Gia_IsoGetItem(p,iBegin+k) );
            Gia_IsoSetValue( p, iBegin+k, pObj.Value );
            if pObj.Value != pObj0.Value {
                fSameValue = 0; }
        }
        if fSameValue {
            Vec_IntPush( p.vClasses2, iBegin );
            Vec_IntPush( p.vClasses2, nSize );
            continue;
        }
        fRefined = 1;
        // sort objects
        clk = Abc_Clock();
        Abc_QuickSort3( p.pStoreW + iBegin, nSize, 0 );
        p.timeSort += Abc_Clock() - clk;
        // divide into new classes
        iBeginOld = iBegin;
        pObj0 = Gia_ManObj( p.pGia, Gia_IsoGetItem(p,iBegin) );
        for  k = 1; k < nSize; k++  {
            pObj = Gia_ManObj( p.pGia, Gia_IsoGetItem(p,iBegin+k) );
            if pObj0.Value == pObj.Value {
                continue; }
            nSizeNew = iBegin + k - iBeginOld;
            if nSizeNew == 1 {
                assert(p.pUniques[Gia_IsoGetItem(p, iBeginOld)] == 0, "p.pUniques[Gia_IsoGetItem(p, iBeginOld)] == 0");
                p.pUniques[Gia_IsoGetItem(p, iBeginOld)] = p.nUniques++;
                p.nSingles++;
            
            } else {
                Vec_IntPush( p.vClasses2, iBeginOld );
                Vec_IntPush( p.vClasses2, nSizeNew );
            }
            iBeginOld = iBegin + k;
            pObj0 = pObj;
        }
        // add the last one
        nSizeNew = iBegin + k - iBeginOld;
        if nSizeNew == 1 {
            assert(p.pUniques[Gia_IsoGetItem(p, iBeginOld)] == 0, "p.pUniques[Gia_IsoGetItem(p, iBeginOld)] == 0");
            p.pUniques[Gia_IsoGetItem(p, iBeginOld)] = p.nUniques++;
            p.nSingles++;
        
        } else {
            Vec_IntPush( p.vClasses2, iBeginOld );
            Vec_IntPush( p.vClasses2, nSizeNew );
        }
    }

    // inlining of macro ABC_SWAP
    { var t *Vec_Int_t  = p.vClasses; p.vClasses = p.vClasses2; p.vClasses2 = t; };
    p.nEntries -= p.nSingles;
    return fRefined;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_IsoCollectCosClasses(p *Gia_IsoMan_t, fVerbose int32) *Vec_Ptr_t {
    var vGroups *Vec_Ptr_t;
    var vLevel *Vec_Int_t;
    var pObj *Gia_Obj_t;
    var i, k, iBegin, nSize int32;
    // add singletons
    vGroups = Vec_PtrAlloc( 1000 );
    // inlining of macro Gia_ManForEachPo
    for  i = 0; (i < Gia_ManPoNum(p.pGia)) && ((pObj) = Gia_ManCo(p.pGia, i)); i++  {
        if p.pUniques[Gia_ObjId(p.pGia, pObj)] > 0 {
            vLevel = Vec_IntAlloc( 1 );
            Vec_IntPush( vLevel, i );
            Vec_PtrPush( vGroups, vLevel );
        }
    }

    // add groups
    // inlining of macro Vec_IntForEachEntryDouble
    for  i = 0; (i+1 < Vec_IntSize(p.vClasses)) && (((iBegin) = Vec_IntEntry(p.vClasses, i)), 1) && (((nSize) = Vec_IntEntry(p.vClasses, i+1)), 1); i += 2  {
        for  k = 0; k < nSize; k++  {
            pObj = Gia_ManObj( p.pGia, Gia_IsoGetItem(p,iBegin+k) );
            if Gia_ObjIsPo(p.pGia, pObj) {
                break; }
        }
        if k == nSize {
            continue; }
        vLevel = Vec_IntAlloc( 8 );
        for  k = 0; k < nSize; k++  {
            pObj = Gia_ManObj( p.pGia, Gia_IsoGetItem(p,iBegin+k) );
            if Gia_ObjIsPo(p.pGia, pObj) {
                Vec_IntPush( vLevel, Gia_ObjCioId(pObj) ); }
        }
        Vec_PtrPush( vGroups, vLevel );
    }
    // canonicize order
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vGroups)) && (((vLevel) = (Vec_Int_t *)Vec_PtrEntry(vGroups, i)), 1); i++  {
        Vec_IntSort( vLevel, 0 );
    }
    Vec_VecSortByFirstInt( (Vec_Vec_t *)vGroups, 0 );
//    Vec_VecFree( (Vec_Vec_t *)vGroups );
//    return NULL;
    return vGroups;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_IsoUpdateValue(Value int32, fCompl int32) uint32 {
    return (Value+1) * s_256Primes[Abc_Var2Lit(Value, fCompl) & ISO_MASK];
}
func Gia_IsoUpdate(p *Gia_IsoMan_t, Iter int32, iObj int32, fCompl int32) uint32 {
    if Iter == 0 {              return Gia_IsoUpdateValue( p.pLevels[iObj], fCompl ); }
    if p.pUniques[iObj] > 0 {  return Gia_IsoUpdateValue( p.pUniques[iObj], fCompl ); }
//    if ( p->pUniques[iObj] > 0 )  return Gia_IsoUpdateValue( 11, fCompl );
    return 0;
}
func Gia_IsoSimulate(p *Gia_IsoMan_t, Iter int32) {
    var pObj, pObjF *Gia_Obj_t;
    var i, iObj int32;
    // initialize constant, inputs, and flops in the first frame
    Gia_ManConst0(p.pGia).Value += s_256Primes[ISO_MASK];
    // inlining of macro Gia_ManForEachPi
    for  i = 0; (i < Gia_ManPiNum(p.pGia)) && ((pObj) = Gia_ManCi(p.pGia, i)); i++  {
        pObj.Value += s_256Primes[ISO_MASK-1];
    }
    if Iter == 0 {
        // inlining of macro Gia_ManForEachRo
        for  i = 0; (i < Gia_ManRegNum(p.pGia)) && ((pObj) = Gia_ManCi(p.pGia, Gia_ManPiNum(p.pGia)+i)); i++  {
            pObj.Value += s_256Primes[ISO_MASK-2];
        } }
    // simulate nodes
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.pGia.nObjs) && ((pObj) = Gia_ManObj(p.pGia, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        pObj.Value += Gia_ObjFanin0(pObj).Value + Gia_IsoUpdate(p, Iter, Gia_ObjFaninId0(pObj, i), Gia_ObjFaninC0(pObj));
        pObj.Value += Gia_ObjFanin1(pObj).Value + Gia_IsoUpdate(p, Iter, Gia_ObjFaninId1(pObj, i), Gia_ObjFaninC1(pObj));
    }
    }
    // simulate COs
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.pGia.vCos)) && ((pObj) = Gia_ManCo(p.pGia, i)); i++  {
        iObj = Gia_ObjId(p.pGia, pObj);
        pObj.Value += Gia_ObjFanin0(pObj).Value + Gia_IsoUpdate(p, Iter, Gia_ObjFaninId0(pObj, iObj), Gia_ObjFaninC0(pObj));
    }
    // transfer flop values
    // inlining of macro Gia_ManForEachRiRo
    for  i = 0; (i < Gia_ManRegNum(p.pGia)) && ((pObjF) = Gia_ManCo(p.pGia, Gia_ManPoNum(p.pGia)+i)) && ((pObj) = Gia_ManCi(p.pGia, Gia_ManPiNum(p.pGia)+i)); i++  {
        pObj.Value += pObjF.Value;
    }
}
func Gia_IsoSimulateBack(p *Gia_IsoMan_t, Iter int32) {
    var pObj, pObjF *Gia_Obj_t;
    var i, iObj int32;
    // simulate COs
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.pGia.vCos)) && ((pObj) = Gia_ManCo(p.pGia, i)); i++  {
        iObj = Gia_ObjId(p.pGia, pObj);
        Gia_ObjFanin0(pObj).Value += pObj.Value + Gia_IsoUpdate(p, Iter, iObj, Gia_ObjFaninC0(pObj));
    }
    // simulate objects
    // inlining of macro Gia_ManForEachAndReverse
    for  i = p.pGia.nObjs - 1; (i > 0) && ((pObj) = Gia_ManObj(p.pGia, i)); i--  {  if !Gia_ObjIsAnd(pObj) {} else {
        Gia_ObjFanin0(pObj).Value += pObj.Value + Gia_IsoUpdate(p, Iter, i, Gia_ObjFaninC0(pObj));
        Gia_ObjFanin1(pObj).Value += pObj.Value + Gia_IsoUpdate(p, Iter, i, Gia_ObjFaninC1(pObj));
    }
    }
    // transfer flop values
    // inlining of macro Gia_ManForEachRiRo
    for  i = 0; (i < Gia_ManRegNum(p.pGia)) && ((pObjF) = Gia_ManCo(p.pGia, Gia_ManPoNum(p.pGia)+i)) && ((pObj) = Gia_ManCi(p.pGia, Gia_ManPiNum(p.pGia)+i)); i++  {
        pObjF.Value += pObj.Value;
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_IsoAssignOneClass2(p *Gia_IsoMan_t) {
    var i int32;
    var iBegin int32 = -1;
    var nSize int32 = -1;
    // find two variable class
    assert(Vec_IntSize(p.vClasses) > 0, "Vec_IntSize(p.vClasses) > 0");
    // inlining of macro Vec_IntForEachEntryDouble
    for  i = 0; (i+1 < Vec_IntSize(p.vClasses)) && (((iBegin) = Vec_IntEntry(p.vClasses, i)), 1) && (((nSize) = Vec_IntEntry(p.vClasses, i+1)), 1); i += 2  {
        if nSize == 2 {
            break; }
    }
    assert(nSize > 1, "nSize > 1");

    if nSize == 2 {
        assert(p.pUniques[Gia_IsoGetItem(p, iBegin)] == 0, "p.pUniques[Gia_IsoGetItem(p, iBegin)] == 0");
        p.pUniques[Gia_IsoGetItem(p, iBegin)] = p.nUniques++;
        p.nSingles++;
        p.nEntries--;

        assert(p.pUniques[Gia_IsoGetItem(p, iBegin+1)] == 0, "p.pUniques[Gia_IsoGetItem(p, iBegin+1)] == 0");
        p.pUniques[Gia_IsoGetItem(p, iBegin+1)] = p.nUniques++;
        p.nSingles++;
        p.nEntries--;
    
    } else {
        assert(p.pUniques[Gia_IsoGetItem(p, iBegin)] == 0, "p.pUniques[Gia_IsoGetItem(p, iBegin)] == 0");
        p.pUniques[Gia_IsoGetItem(p, iBegin)] = p.nUniques++;
        p.nSingles++;
        p.nEntries--;
    }

    for  ; i < Vec_IntSize(p.vClasses) - 2; i += 2  {
        p.vClasses.pArray[i+0] = p.vClasses.pArray[i+2];
        p.vClasses.pArray[i+1] = p.vClasses.pArray[i+3];
    }
    Vec_IntShrink( p.vClasses, Vec_IntSize(p.vClasses) - 2 );

    fmt.Printf("Broke ties in class %d of size %d at level %d.\n", i/2, nSize, p.pLevels[Gia_IsoGetItem(p, iBegin)]);
}

func Gia_IsoAssignOneClass3(p *Gia_IsoMan_t) {
    var iBegin, nSize int32;
    // find the last class
    assert(Vec_IntSize(p.vClasses) > 0, "Vec_IntSize(p.vClasses) > 0");
    iBegin = Vec_IntEntry( p.vClasses, Vec_IntSize(p.vClasses) - 2 );
    nSize  = Vec_IntEntry( p.vClasses, Vec_IntSize(p.vClasses) - 1 );
    Vec_IntShrink( p.vClasses, Vec_IntSize(p.vClasses) - 2 );

    // assign the class
    assert(nSize > 1, "nSize > 1");
    if nSize == 2 {
        assert(p.pUniques[Gia_IsoGetItem(p, iBegin)] == 0, "p.pUniques[Gia_IsoGetItem(p, iBegin)] == 0");
        p.pUniques[Gia_IsoGetItem(p, iBegin)] = p.nUniques++;
        p.nSingles++;
        p.nEntries--;

        assert(p.pUniques[Gia_IsoGetItem(p, iBegin+1)] == 0, "p.pUniques[Gia_IsoGetItem(p, iBegin+1)] == 0");
        p.pUniques[Gia_IsoGetItem(p, iBegin+1)] = p.nUniques++;
        p.nSingles++;
        p.nEntries--;
    
    } else {
        assert(p.pUniques[Gia_IsoGetItem(p, iBegin)] == 0, "p.pUniques[Gia_IsoGetItem(p, iBegin)] == 0");
        p.pUniques[Gia_IsoGetItem(p, iBegin)] = p.nUniques++;
        p.nSingles++;
        p.nEntries--;
    }
    fmt.Printf("Broke ties in last class of size %d at level %d.\n", nSize, p.pLevels[Gia_IsoGetItem(p, iBegin)]);
}

func Gia_IsoAssignOneClass(p *Gia_IsoMan_t, fVerbose int32) {
    var i, k, iBegin0, iBegin, nSize, Shrink int32;
    // find the classes with the highest level
    assert(Vec_IntSize(p.vClasses) > 0, "Vec_IntSize(p.vClasses) > 0");
    iBegin0 = Vec_IntEntry( p.vClasses, Vec_IntSize(p.vClasses) - 2 );
    for  i = Vec_IntSize(p.vClasses) - 2; i >= 0; i -= 2  {
        iBegin = Vec_IntEntry( p.vClasses, i );
        if p.pLevels[Gia_IsoGetItem(p, iBegin)] != p.pLevels[Gia_IsoGetItem(p, iBegin0)] {
            break; }
    }
    i += 2;
    assert(i >= 0, "i >= 0");
    // assign all classes starting with this one
    for  Shrink = i; i < Vec_IntSize(p.vClasses); i += 2  {
        iBegin = Vec_IntEntry( p.vClasses, i );
        nSize  = Vec_IntEntry( p.vClasses, i + 1 );
        for  k = 0; k < nSize; k++  {
            assert(p.pUniques[Gia_IsoGetItem(p, iBegin+k)] == 0, "p.pUniques[Gia_IsoGetItem(p, iBegin+k)] == 0");
            p.pUniques[Gia_IsoGetItem(p, iBegin+k)] = p.nUniques++;
//            Gia_ManObj(p->pGia, Gia_IsoGetItem(p, iBegin+k))->Value += s_256Primes[0]; ///  new addition!!!
            p.nSingles++;
            p.nEntries--;
        }
        if fVerbose != 0 {
            fmt.Printf("Broke ties in class of size %d at level %d.\n", nSize, p.pLevels[Gia_IsoGetItem(p, iBegin)]); }
    }
    Vec_IntShrink( p.vClasses, Shrink );
}

/**Function*************************************************************

  Synopsis    [Report topmost equiv nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_IsoReportTopmost(p *Gia_IsoMan_t) {
    var pObj *Gia_Obj_t;
    var i, k, iBegin, nSize, Counter int32;
    // go through equivalence classes
    Gia_ManIncrementTravId( p.pGia );
    // inlining of macro Vec_IntForEachEntryDouble
    for  i = 0; (i+1 < Vec_IntSize(p.vClasses)) && (((iBegin) = Vec_IntEntry(p.vClasses, i)), 1) && (((nSize) = Vec_IntEntry(p.vClasses, i+1)), 1); i += 2  {
//        printf( "%d(%d) ", nSize, p->pLevels[Gia_IsoGetItem(p, iBegin)] );
        for  k = 0; k < nSize; k++  {
            pObj = Gia_ManObj( p.pGia, Gia_IsoGetItem(p, iBegin+k) );
            if Gia_ObjIsAnd(pObj) {
                Gia_ObjSetTravIdCurrent( p.pGia, Gia_ObjFanin0(pObj) );
                Gia_ObjSetTravIdCurrent( p.pGia, Gia_ObjFanin1(pObj) );
            
            } else if Gia_ObjIsRo(p.pGia, pObj) {
                Gia_ObjSetTravIdCurrent( p.pGia, Gia_ObjFanin0(Gia_ObjRoToRi(p.pGia, pObj)) ); }
        }
    }
//    printf( "\n" );

    // report non-labeled nodes
    // inlining of macro Vec_IntForEachEntryDouble
    for  i = 0; (i+1 < Vec_IntSize(p.vClasses)) && (((iBegin) = Vec_IntEntry(p.vClasses, i)), 1) && (((nSize) = Vec_IntEntry(p.vClasses, i+1)), 1); i += 2  {
        for  k = 0; k < nSize; k++  {
            pObj = Gia_ManObj( p.pGia, Gia_IsoGetItem(p, iBegin+k) );
            if !Gia_ObjIsTravIdCurrent(p.pGia, pObj) {
                fmt.Printf("%5d : ", ++Counter);
                fmt.Printf("Obj %6d : Level = %4d.  iBegin = %4d.  Size = %4d.\n", Gia_ObjId(p.pGia, pObj), p.pLevels[Gia_ObjId(p.pGia, pObj)], iBegin, nSize);
                break;
            }
        }
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_IsoRecognizeMuxes(pGia *Gia_Man_t) {
    var pObj, pObjC, pObj1, pObj0 *Gia_Obj_t;
    var i int32;
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < pGia.nObjs) && ((pObj) = Gia_ManObj(pGia, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        if !Gia_ObjIsMuxType(pObj) {
            continue; }
        pObjC = Gia_ObjRecognizeMux( pObj, &pObj1, &pObj0 );
        if Gia_Regular(pObj0) == Gia_Regular(pObj1) {
            // this is XOR
            Gia_Regular(pObj).Value += s_256Primes[233];
            Gia_Regular(pObjC).Value += s_256Primes[234];
            Gia_Regular(pObj0).Value += s_256Primes[234];
        
        } else {
            // this is MUX
            Gia_Regular(pObj).Value += s_256Primes[235];
            Gia_Regular(pObjC).Value += s_256Primes[236];
            Gia_Regular(pObj0).Value += s_256Primes[237];
            Gia_Regular(pObj1).Value += s_256Primes[237];
        }
    }
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_IsoDeriveEquivPos(pGia *Gia_Man_t, fForward int32, fVerbose int32) *Vec_Ptr_t {
    var nIterMax int32  = 10000;
    var nFixedPoint int32  = 1;
    var p *Gia_IsoMan_t;
    var vEquivs *Vec_Ptr_t  = nil;
    var fRefined, fRefinedAll int32;
    var i, c int32;
    var clk abctime = Abc_Clock();
    var clkTotal abctime = Abc_Clock();
    assert(Gia_ManCiNum(pGia) > 0, "Gia_ManCiNum(pGia) > 0");
    assert(Gia_ManPoNum(pGia) > 0, "Gia_ManPoNum(pGia) > 0");

    Gia_ManCleanValue( pGia );
    p = Gia_IsoManStart( pGia );
    Gia_IsoPrepare( p );
    Gia_IsoAssignUnique( p );
    p.timeStart = Abc_Clock() - clk;
    if fVerbose != 0 {
        Gia_IsoPrint( p, 0, Abc_Clock() - clkTotal ); }

//    Gia_IsoRecognizeMuxes( pGia );

    i = 0;
    if fForward != 0 {
        for  c = 0; i < nIterMax && c < nFixedPoint+1; i, c = i+1, fRefined ? 0 : c+1  {
            clk = Abc_Clock();   Gia_IsoSimulate( p, i );      p.timeSim += Abc_Clock() - clk;
            clk = Abc_Clock();   fRefined = Gia_IsoSort( p );  p.timeRefine += Abc_Clock() - clk;
            if fVerbose != 0 {
                Gia_IsoPrint( p, i+1, Abc_Clock() - clkTotal ); }
        }
    
    } else {
        for  Vec_IntSize(p.vClasses) > 0  {
            for  fRefinedAll = 1; i < nIterMax && (fRefinedAll != 0);  {
                fRefinedAll = 0;
                for  c = 0; i < nIterMax && c < nFixedPoint+1; i, c = i+1, fRefined ? 0 : c+1  {
                    clk = Abc_Clock();   Gia_IsoSimulate( p, i );      p.timeSim += Abc_Clock() - clk;
                    clk = Abc_Clock();   fRefined = Gia_IsoSort( p );  p.timeRefine += Abc_Clock() - clk;
                    if fVerbose != 0 {
                        Gia_IsoPrint( p, i+1, Abc_Clock() - clkTotal ); }
                    fRefinedAll |= fRefined;
                }
                for  c = 0; i < nIterMax && c < nFixedPoint+1; i, c = i+1, fRefined ? 0 : c+1  {
                    clk = Abc_Clock();   Gia_IsoSimulateBack( p, i );  p.timeSim += Abc_Clock() - clk;
                    clk = Abc_Clock();   fRefined = Gia_IsoSort( p );  p.timeRefine += Abc_Clock() - clk;
                    if fVerbose != 0 {
                        Gia_IsoPrint( p, i+1, Abc_Clock() - clkTotal ); }
                    fRefinedAll |= fRefined;
                }
            }
            if !(fRefinedAll != 0) {
                break; }
        }

//        Gia_IsoReportTopmost( p );

        for  Vec_IntSize(p.vClasses) > 0  {
            Gia_IsoAssignOneClass( p, fVerbose );
            for  fRefinedAll = 1; i < nIterMax && (fRefinedAll != 0);  {
                fRefinedAll = 0;
                for  c = 0; i < nIterMax && c < nFixedPoint; i, c = i+1, fRefined ? 0 : c+1  {
                    clk = Abc_Clock();   Gia_IsoSimulateBack( p, i );  p.timeSim += Abc_Clock() - clk;
                    clk = Abc_Clock();   fRefined = Gia_IsoSort( p );  p.timeRefine += Abc_Clock() - clk;
                    if fVerbose != 0 {
                        Gia_IsoPrint( p, i+1, Abc_Clock() - clkTotal ); }
                    fRefinedAll |= fRefined;
                }
                for  c = 0; i < nIterMax && c < nFixedPoint; i, c = i+1, fRefined ? 0 : c+1  {
                    clk = Abc_Clock();   Gia_IsoSimulate( p, i );      p.timeSim += Abc_Clock() - clk;
                    clk = Abc_Clock();   fRefined = Gia_IsoSort( p );  p.timeRefine += Abc_Clock() - clk;
                    if fVerbose != 0 {
                        Gia_IsoPrint( p, i+1, Abc_Clock() - clkTotal ); }
                    fRefinedAll |= fRefined;
//                    if ( fRefined )
//                        printf( "Refinedment happened.\n" );
                }
            }
        }

        if fVerbose != 0 {
            Gia_IsoPrint( p, i+2, Abc_Clock() - clkTotal ); }
    }
//    Gia_IsoPrintClasses( p );
    if fVerbose != 0 {
        p.timeTotal = Abc_Clock() - clkTotal;
        p.timeOther = p.timeTotal - p.timeStart - p.timeSim - p.timeRefine;
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("Start    "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeStart)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeStart)))/(float64((p.timeTotal))) : 0.0);;
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("Simulate "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeSim)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeSim)))/(float64((p.timeTotal))) : 0.0);;
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("Refine   "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeRefine-p.timeSort)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeRefine-p.timeSort)))/(float64((p.timeTotal))) : 0.0);;
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("Sort     "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeSort)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeSort)))/(float64((p.timeTotal))) : 0.0);;
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("Other    "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeOther)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeOther)))/(float64((p.timeTotal))) : 0.0);;
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("TOTAL    "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeTotal)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeTotal)))/(float64((p.timeTotal))) : 0.0);;
    }

    if Gia_ManPoNum(p.pGia) > 1 {
        vEquivs = Gia_IsoCollectCosClasses( p, fVerbose ); }
    Gia_IsoManTransferUnique( p );
    Gia_IsoManStop( p );

    return vEquivs;
}




/**Function*************************************************************

  Synopsis    [Finds canonical ordering of CIs/COs/nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ObjCompareByValue(pp1 **Gia_Obj_t, pp2 **Gia_Obj_t) int32 {
    var pObj1 *Gia_Obj_t  = *pp1;
    var pObj2 *Gia_Obj_t  = *pp2;
//    assert( pObj1->Value != pObj2->Value );
    return int32(pObj1.Value) - int32(pObj2.Value);
}
func Gia_ManFindCaninicalOrder_rec(p *Gia_Man_t, pObj *Gia_Obj_t, vAnds *Vec_Int_t) {
    if Gia_ObjIsTravIdCurrent(p, pObj) {
        return; }
    Gia_ObjSetTravIdCurrent(p, pObj);
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    if !Gia_ObjIsAnd(Gia_ObjFanin0(pObj)) || !Gia_ObjIsAnd(Gia_ObjFanin1(pObj)) {
        Gia_ManFindCaninicalOrder_rec( p, Gia_ObjFanin0(pObj), vAnds );
        Gia_ManFindCaninicalOrder_rec( p, Gia_ObjFanin1(pObj), vAnds );
    
    } else {
        assert(Gia_ObjFanin0(pObj).Value != Gia_ObjFanin1(pObj).Value, "Gia_ObjFanin0(pObj).Value != Gia_ObjFanin1(pObj).Value");
        if Gia_ObjFanin0(pObj).Value < Gia_ObjFanin1(pObj).Value {
            Gia_ManFindCaninicalOrder_rec( p, Gia_ObjFanin0(pObj), vAnds );
            Gia_ManFindCaninicalOrder_rec( p, Gia_ObjFanin1(pObj), vAnds );
        
        } else {
            Gia_ManFindCaninicalOrder_rec( p, Gia_ObjFanin1(pObj), vAnds );
            Gia_ManFindCaninicalOrder_rec( p, Gia_ObjFanin0(pObj), vAnds );
        }
    }
    Vec_IntPush( vAnds, Gia_ObjId(p, pObj) );
}
func Gia_ManFindCaninicalOrder(p *Gia_Man_t, vCis *Vec_Int_t, vAnds *Vec_Int_t, vCos *Vec_Int_t, pvPiPerm **Vec_Int_t) {
    var vTemp *Vec_Ptr_t;
    var pObj *Gia_Obj_t;
    var i int32;

    vTemp = Vec_PtrAlloc( 1000 );
    Vec_IntClear( vCis );
    Vec_IntClear( vAnds );
    Vec_IntClear( vCos );

    // assign unique IDs to PIs
    Vec_PtrClear( vTemp );
    // inlining of macro Gia_ManForEachPi
    for  i = 0; (i < Gia_ManPiNum(p)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        Vec_PtrPush( vTemp, pObj );
    }
    Vec_PtrSort( vTemp, (int (*)(void))Gia_ObjCompareByValue );
    // create the result
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vTemp)) && (((pObj) = (Gia_Obj_t *)Vec_PtrEntry(vTemp, i)), 1); i++  {
        Vec_IntPush( vCis, Gia_ObjId(p, pObj) );
    }
    // remember PI permutation
    if pvPiPerm {
        *pvPiPerm = Vec_IntAlloc( Gia_ManPiNum(p) );
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(vTemp)) && (((pObj) = (Gia_Obj_t *)Vec_PtrEntry(vTemp, i)), 1); i++  {
            Vec_IntPush( *pvPiPerm, Gia_ObjCioId(pObj) );
        }
    }

    // assign unique IDs to POs
    if Gia_ManPoNum(p) == 1 {
        Vec_IntPush( vCos, Gia_ObjId(p, Gia_ManPo(p, 0)) );
    } else {
        Vec_PtrClear( vTemp );
        // inlining of macro Gia_ManForEachPo
        for  i = 0; (i < Gia_ManPoNum(p)) && ((pObj) = Gia_ManCo(p, i)); i++  {
            pObj.Value = Abc_Var2Lit( Gia_ObjFanin0(pObj).Value, Gia_ObjFaninC0(pObj) );
            Vec_PtrPush( vTemp, pObj );
        }
        Vec_PtrSort( vTemp, (int (*)(void))Gia_ObjCompareByValue );
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(vTemp)) && (((pObj) = (Gia_Obj_t *)Vec_PtrEntry(vTemp, i)), 1); i++  {
            Vec_IntPush( vCos, Gia_ObjId(p, pObj) );
        }
    }

    // assign unique IDs to ROs
    Vec_PtrClear( vTemp );
    // inlining of macro Gia_ManForEachRo
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCi(p, Gia_ManPiNum(p)+i)); i++  {
        Vec_PtrPush( vTemp, pObj );
    }
    Vec_PtrSort( vTemp, (int (*)(void))Gia_ObjCompareByValue );
    // create the result
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vTemp)) && (((pObj) = (Gia_Obj_t *)Vec_PtrEntry(vTemp, i)), 1); i++  {
        Vec_IntPush( vCis, Gia_ObjId(p, pObj) );
        Vec_IntPush( vCos, Gia_ObjId(p, Gia_ObjRoToRi(p, pObj)) );
    }
    Vec_PtrFree( vTemp );

    // assign unique IDs to internal nodes
    Gia_ManIncrementTravId( p );
    Gia_ObjSetTravIdCurrent( p, Gia_ManConst0(p) );
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vCis)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vCis,i))); i++  {
        Gia_ObjSetTravIdCurrent( p, pObj );
    }
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vCos)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vCos,i))); i++  {
        Gia_ManFindCaninicalOrder_rec( p, Gia_ObjFanin0(pObj), vAnds );
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManIsoCanonicize(p *Gia_Man_t, fVerbose int32) *Gia_Man_t {
    var pRes *Gia_Man_t  = nil;
    var vCis, vAnds, vCos *Vec_Int_t;
    var vEquiv *Vec_Ptr_t;
    if Gia_ManCiNum(p) == 0 { // const AIG
    
        assert(Gia_ManPoNum(p) == 1, "Gia_ManPoNum(p) == 1");
        assert(Gia_ManObjNum(p) == 2, "Gia_ManObjNum(p) == 2");
        return Gia_ManDup(p);
    }
    // derive canonical values
    vEquiv = Gia_IsoDeriveEquivPos( p, 0, fVerbose );
    Vec_VecFreeP( (Vec_Vec_t **)&vEquiv );
    // find canonical order of CIs/COs/nodes
    // find canonical order
    vCis  = Vec_IntAlloc( Gia_ManCiNum(p) );
    vAnds = Vec_IntAlloc( Gia_ManAndNum(p) );
    vCos  = Vec_IntAlloc( Gia_ManCoNum(p) );
    Gia_ManFindCaninicalOrder( p, vCis, vAnds, vCos, nil );
    // derive the new AIG
    pRes = Gia_ManDupFromVecs( p, vCis, vAnds, vCos, Gia_ManRegNum(p) );
    // cleanup
    Vec_IntFree( vCis );
    Vec_IntFree( vAnds );
    Vec_IntFree( vCos );
    return pRes;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManIsoFindString(p *Gia_Man_t, iPo int32, fVerbose int32, pvPiPerm **Vec_Int_t) *Vec_Str_t {
    var pPart *Gia_Man_t;
    var vEquiv *Vec_Ptr_t;
    var vCis, vAnds, vCos *Vec_Int_t;
    var vStr *Vec_Str_t;
    // duplicate
    pPart = Gia_ManDupCones( p, &iPo, 1, 1 );
//Gia_ManPrint( pPart );
    assert(Gia_ManPoNum(pPart) == 1, "Gia_ManPoNum(pPart) == 1");
    if Gia_ManCiNum(pPart) == 0 { // const AIG
    
        assert(Gia_ManPoNum(pPart) == 1, "Gia_ManPoNum(pPart) == 1");
        assert(Gia_ManObjNum(pPart) == 2, "Gia_ManObjNum(pPart) == 2");
        vStr = Gia_AigerWriteIntoMemoryStr( pPart );
        Gia_ManStop( pPart );
        if pvPiPerm {
            *pvPiPerm = Vec_IntAlloc( 0 ); }
        return vStr;
    }
    // derive canonical values
    vEquiv = Gia_IsoDeriveEquivPos( pPart, 0, fVerbose );
    Vec_VecFreeP( (Vec_Vec_t **)&vEquiv );
    // find canonical order
    vCis  = Vec_IntAlloc( Gia_ManCiNum(pPart) );
    vAnds = Vec_IntAlloc( Gia_ManAndNum(pPart) );
    vCos  = Vec_IntAlloc( Gia_ManCoNum(pPart) );
    Gia_ManFindCaninicalOrder( pPart, vCis, vAnds, vCos, pvPiPerm );
//printf( "Internal: " );
//Vec_IntPrint( vCis );
    // derive the AIGER string
    vStr = Gia_AigerWriteIntoMemoryStrPart( pPart, vCis, vAnds, vCos, Gia_ManRegNum(pPart) );
    // cleanup
    Vec_IntFree( vCis );
    Vec_IntFree( vAnds );
    Vec_IntFree( vCos );
    Gia_ManStop( pPart );
    return vStr;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntCountNonTrivial(vEquivs *Vec_Ptr_t, pnUsed *int32) int32 {
    var vClass *Vec_Int_t;
    var i, nClasses int32;
    *pnUsed = 0;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vEquivs)) && (((vClass) = (Vec_Int_t *)Vec_PtrEntry(vEquivs, i)), 1); i++  {
        if Vec_IntSize(vClass) < 2 {
            continue; }
        nClasses++;
        (*pnUsed) += Vec_IntSize(vClass);
    }
    return nClasses;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManIsoReduce(pInit *Gia_Man_t, pvPosEquivs **Vec_Ptr_t, pvPiPerms **Vec_Ptr_t, fEstimate int32, fDualOut int32, fVerbose int32, fVeryVerbose int32) *Gia_Man_t { 
    var p, pPart *Gia_Man_t;
    var vEquivs, vEquivs2, vStrings *Vec_Ptr_t;
    var vRemain, vLevel, vLevel2 *Vec_Int_t;
    var vStr, vStr2 *Vec_Str_t;
    var i, k, s, sStart, iPo, Counter int32;
    var nClasses, nUsedPos int32;
    var clk abctime  = Abc_Clock();
    if pvPosEquivs {
        *pvPosEquivs = nil; }
    if pvPiPerms {
        *pvPiPerms = Vec_PtrStart( Gia_ManPoNum(pInit) ); }

    if fDualOut != 0 {
        assert((Gia_ManPoNum(pInit) & 1) == 0, "(Gia_ManPoNum(pInit) & 1) == 0");
        if Gia_ManPoNum(pInit) == 2 {
            return Gia_ManDup(pInit); }
        p = Gia_ManTransformMiter( pInit );
        p = Gia_ManSeqStructSweep( pPart = p, 1, 1, 0 );
        Gia_ManStop( pPart );
    
    } else {
        if Gia_ManPoNum(pInit) == 1 {
            return Gia_ManDup(pInit); }
        p = pInit;
    }

    // create preliminary equivalences
    vEquivs = Gia_IsoDeriveEquivPos( p, 1, fVeryVerbose );
    if vEquivs == nil {
        if fDualOut != 0 {
            Gia_ManStop( p ); }
        return nil;
    }
    nClasses = Vec_IntCountNonTrivial( vEquivs, &nUsedPos );
    fmt.Printf("Reduced %d outputs to %d candidate   classes (%d outputs are in %d non-trivial classes).  ", Gia_ManPoNum(p), Vec_PtrSize(vEquivs), nUsedPos, nClasses);
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
    if fEstimate != 0 {
        Vec_VecFree( (Vec_Vec_t *)vEquivs );
        return Gia_ManDup(pInit);
    }

    // perform refinement of equivalence classes
    Counter = 0;
    vEquivs2 = Vec_PtrAlloc( 100 );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vEquivs)) && (((vLevel) = (Vec_Int_t *)Vec_PtrEntry(vEquivs, i)), 1); i++  {
        if Vec_IntSize(vLevel) < 2 {
            Vec_PtrPush( vEquivs2, Vec_IntDup(vLevel) );
            for  k = 0; k < Vec_IntSize(vLevel); k++  {
                if Counter++; Counter % 100 == 0 {
                    fmt.Printf("%6d finished...\r", Counter); }
            }
            continue;
        }

        if fVerbose != 0 {
            iPo = Vec_IntEntry(vLevel, 0);
            fmt.Printf("%6d %6d %6d : ", i, Vec_IntSize(vLevel), iPo);
            pPart = Gia_ManDupCones( p, &iPo, 1, 1 );
            Gia_ManPrintStats(pPart, nil);
            Gia_ManStop( pPart );
        }

        sStart = Vec_PtrSize( vEquivs2 ); 
        vStrings = Vec_PtrAlloc( 100 );
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vLevel)) && (((iPo) = Vec_IntEntry(vLevel, k)), 1); k++  {
            if Counter++; Counter % 100 == 0 {
                fmt.Printf("%6d finished...\r", Counter); }
            assert(pvPiPerms == nil || Vec_PtrArray(*pvPiPerms)[iPo] == nil, "pvPiPerms == nil || Vec_PtrArray(*pvPiPerms)[iPo] == nil");
            vStr = Gia_ManIsoFindString( p, iPo, 0, pvPiPerms ? (Vec_Int_t **)Vec_PtrArray(*pvPiPerms) + iPo : nil );

//            printf( "Output %2d : ", iPo );
//            Vec_IntPrint( Vec_PtrArray(*pvPiPerms)[iPo] );

            // check if this string already exists
            // inlining of macro Vec_PtrForEachEntry
            for  s = 0; (s < Vec_PtrSize(vStrings)) && (((vStr2) = (Vec_Str_t *)Vec_PtrEntry(vStrings, s)), 1); s++  {
                if Vec_StrCompareVec(vStr, vStr2) == 0 {
                    break; }
            }
            if s == Vec_PtrSize(vStrings) {
                Vec_PtrPush( vStrings, vStr );
                Vec_PtrPush( vEquivs2, Vec_IntAlloc(8) );
            
            } else {
                Vec_StrFree( vStr ); }
            // add this entry to the corresponding level
            vLevel2 = (Vec_Int_t *)Vec_PtrEntry( vEquivs2, sStart + s );
            Vec_IntPush( vLevel2, iPo );
        }
//        if ( Vec_PtrSize(vEquivs2) - sStart > 1 )
//            printf( "Refined class %d into %d classes.\n", i, Vec_PtrSize(vEquivs2) - sStart );
        Vec_VecFree( (Vec_Vec_t *)vStrings );
    }
    assert(Counter == Gia_ManPoNum(p), "Counter == Gia_ManPoNum(p)");
    Vec_VecSortByFirstInt( (Vec_Vec_t *)vEquivs2, 0 );
    Vec_VecFree( (Vec_Vec_t *)vEquivs );
    vEquivs = vEquivs2;

    // collect the first ones
    vRemain = Vec_IntAlloc( 100 );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vEquivs)) && (((vLevel) = (Vec_Int_t *)Vec_PtrEntry(vEquivs, i)), 1); i++  {
        Vec_IntPush( vRemain, Vec_IntEntry(vLevel, 0) );
    }

    if fDualOut != 0 {
        var vTemp *Vec_Int_t  = Vec_IntAlloc( Vec_IntSize(vRemain) );
        var i, Entry int32;
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(vRemain)) && (((Entry) = Vec_IntEntry(vRemain, i)), 1); i++  {
//            printf( "%d ", Entry );
            Vec_IntPush( vTemp, 2*Entry );
            Vec_IntPush( vTemp, 2*Entry+1 );
        }
//        printf( "\n" );
        Vec_IntFree( vRemain );
        vRemain = vTemp;
        Gia_ManStop( p );
        p = pInit;
    }


    // derive the resulting AIG
    pPart = Gia_ManDupCones( p, Vec_IntArray(vRemain), Vec_IntSize(vRemain), 0 );
    Vec_IntFree( vRemain );
    // report the results
    nClasses = Vec_IntCountNonTrivial( vEquivs, &nUsedPos );
    if !(fDualOut != 0) {
        fmt.Printf("Reduced %d outputs to %d equivalence classes (%d outputs are in %d non-trivial classes).  ", Gia_ManPoNum(p), Vec_PtrSize(vEquivs), nUsedPos, nClasses);
    } else {
        fmt.Printf("Reduced %d dual outputs to %d dual outputs.  ", Gia_ManPoNum(p)/2, Gia_ManPoNum(pPart)/2); }
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
    if fVerbose != 0 {
        fmt.Printf("Nontrivial classes:\n");
        Vec_VecPrintInt( (Vec_Vec_t *)vEquivs, 1 );
    }
    if pvPosEquivs {
        *pvPosEquivs = vEquivs;
    } else {
        Vec_VecFree( (Vec_Vec_t *)vEquivs ); }
//    Gia_ManStopP( &pPart );
    return pPart;
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_IsoTestOld(p *Gia_Man_t, fVerbose int32) {
    var vEquivs *Vec_Ptr_t;
    var clk abctime  = Abc_Clock(); 
    vEquivs = Gia_IsoDeriveEquivPos( p, 0, fVerbose );
    fmt.Printf("Reduced %d outputs to %d.  ", Gia_ManPoNum(p), vEquivs ? Vec_PtrSize(vEquivs) : 1);
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
    if (fVerbose != 0) && (vEquivs != nil) && Gia_ManPoNum(p) != Vec_PtrSize(vEquivs) {
        fmt.Printf("Nontrivial classes:\n");
//        Vec_VecPrintInt( (Vec_Vec_t *)vEquivs, 1 );
    }
    Vec_VecFreeP( (Vec_Vec_t **)&vEquivs );
}

/**Function*************************************************************

  Synopsis    [Test remapping of CEXes for isomorphic POs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_IsoTestGenPerm(nPis int32) *Vec_Int_t {
    var vPerm *Vec_Int_t;
    var i int32;
    var pArray *int32;
    vPerm = Vec_IntStartNatural( nPis );
    pArray = Vec_IntArray( vPerm );
    for  i = 0; i < nPis; i++  {
        var iNew int32  = rand() % nPis;
        // inlining of macro ABC_SWAP
        { var t int32  = pArray[i]; pArray[i] = pArray[iNew]; pArray[iNew] = t; };
    }
    return vPerm;
}
func Gia_IsoTest(p *Gia_Man_t, pCex *Abc_Cex_t, fVerbose int32) {
    var pCexNew *Abc_Cex_t;
    var vPiPerm *Vec_Int_t;
    var vPosEquivs, vPisPerm *Vec_Ptr_t;
    var vPerm0, vPerm1 *Vec_Int_t;
    var pPerm, pDouble, pAig *Gia_Man_t;
    assert(Gia_ManPoNum(p) == 1, "Gia_ManPoNum(p) == 1");
    assert(Gia_ManRegNum(p) > 0, "Gia_ManRegNum(p) > 0");
    // generate random permutation of PIs
    vPiPerm = Gia_IsoTestGenPerm( Gia_ManPiNum(p) );
    fmt.Printf("Considering random permutation of the primary inputs of the AIG:\n");
    Vec_IntPrint( vPiPerm );
    // create AIG with two primary outputs (original and permuted)
    pPerm = Gia_ManDupPerm( p, vPiPerm );
    pDouble = Gia_ManDupAppendNew( p, pPerm );
//Gia_AigerWrite( pDouble, "test.aig", 0, 0, 0 );

    // analyze the two-output miter
    pAig = Gia_ManIsoReduce( pDouble, &vPosEquivs, &vPisPerm, 0, 0, 0, 0 );
    Vec_VecFree( (Vec_Vec_t *)vPosEquivs );

    // given CEX for output 0, derive CEX for output 1
    vPerm0 = (Vec_Int_t *)Vec_PtrEntry( vPisPerm, 0 );
    vPerm1 = (Vec_Int_t *)Vec_PtrEntry( vPisPerm, 1 );
    pCexNew = Abc_CexPermuteTwo( pCex, vPerm0, vPerm1 );
    Vec_VecFree( (Vec_Vec_t *)vPisPerm );

    // check that original CEX and the resulting CEX is valid
    if Gia_ManVerifyCex(p, pCex, 0) {
        fmt.Printf("CEX for the init AIG is valid.\n");
    } else {
        fmt.Printf("CEX for the init AIG is not valid.\n"); }
    if Gia_ManVerifyCex(pPerm, pCexNew, 0) {
        fmt.Printf("CEX for the perm AIG is valid.\n");
    } else {
        fmt.Printf("CEX for the perm AIG is not valid.\n"); }
    // delete
    Gia_ManStop( pAig );
    Gia_ManStop( pDouble );
    Gia_ManStop( pPerm );
    Vec_IntFree( vPiPerm );
    Abc_CexFree( pCexNew );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

