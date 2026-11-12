package gia

import (
	"fmt"
	"os"
)

/**CFile****************************************************************

  FileName    [giaCone.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    []

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaCone.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"
//#include "misc/extra/extra.h"
//#include "misc/vec/vecHsh.h"
//#include "misc/vec/vecWec.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

   
type Opa_Man_t struct {
    pGia     *Gia_Man_t;
    vFront     *Vec_Int_t;
    pvParts     *Vec_Int_t;
    pId2Part           *int32; 
    nParts            int32;
}

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////
    
/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Opa_ManStart(pGia *Gia_Man_t) *Opa_Man_t {
    
    var pObj *Gia_Obj_t;
    var i int32;
    // inlining of macro ABC_CALLOC
    p := &Opa_Man_t{};
    p.pGia = pGia;
    // inlining of macro ABC_CALLOC
    p.pvParts  = ((Vec_Int_t *) calloc((size_t)(Gia_ManPoNum(pGia)), sizeof(Vec_Int_t)));
    // inlining of macro ABC_FALLOC
    p.pId2Part = ((int *) memset(malloc(sizeof(int) * (size_t)(Gia_ManObjNum(pGia))), 0xff, sizeof(int) * (size_t)(Gia_ManObjNum(pGia))));
    p.vFront   = Vec_IntAlloc( 100 );
    // inlining of macro Gia_ManForEachPo
    for  i = 0; (i < Gia_ManPoNum(pGia)) && ((pObj) = Gia_ManCo(pGia, i)); i++  {
        Vec_IntPush( p.pvParts + i, Gia_ObjId(pGia, pObj) );
        p.pId2Part[Gia_ObjId(pGia, pObj)] = i;
        Vec_IntPush( p.vFront, Gia_ObjId(pGia, pObj) );
    }
    p.nParts = Gia_ManPoNum(pGia);
    return p;
}
func Opa_ManStop(p *Opa_Man_t) {
    var i int32;
    Vec_IntFree( p.vFront );
    for  i = 0; i < Gia_ManPoNum(p.pGia); i++  {
        // inlining of macro ABC_FREE
        p.pvParts[i].pArray = nil;
    }
    // inlining of macro ABC_FREE
    p.pvParts = nil;
    // inlining of macro ABC_FREE
    p.pId2Part = nil;
    // inlining of macro ABC_FREE
    p = nil;
}
func Opa_ManPrint(p *Opa_Man_t) {
    var i, k int32;
    fmt.Printf("Groups:\n");
    for  i = 0; i < Gia_ManPoNum(p.pGia); i++  {
        if p.pvParts[i].nSize == 0 {
            continue; }
        fmt.Printf("%3d : ", i);
        for  k = 0; k < p.pvParts[i].nSize; k++  {
            fmt.Printf("%d ", p.pvParts[i].pArray[k]);
        }
        fmt.Printf("\n");
    }
}
func Opa_ManPrint2(p *Opa_Man_t) {
    var pObj *Gia_Obj_t;
    var i, k, Count int32;
    fmt.Printf("Groups %d: ", p.nParts);
    for  i = 0; i < Gia_ManPoNum(p.pGia); i++  {
        if p.pvParts[i].nSize == 0 {
            continue; }
        // count POs in this group
        Count = 0;
        // inlining of macro Gia_ManForEachObjVec
        for  k = 0; (k < Vec_IntSize(p.pvParts + i)) && ((pObj) = Gia_ManObj(p.pGia, Vec_IntEntry(p.pvParts + i,k))); k++  {
            Count += Gia_ObjIsPo(p.pGia, pObj);
        }
        fmt.Printf("%d ", Count);
    }
    fmt.Printf("\n");
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Opa_ManMoveOne(p *Opa_Man_t, pObj *Gia_Obj_t, pFanin *Gia_Obj_t) {
    var iObj int32    = Gia_ObjId(p.pGia, pObj);
    var iFanin int32  = Gia_ObjId(p.pGia, pFanin);
    if iFanin == 0 {
        return; }
    assert(p.pId2Part[ iObj ] >= 0, "p.pId2Part[ iObj ] >= 0");
    if p.pId2Part[ iFanin ] == -1 {
        p.pId2Part[ iFanin ] = p.pId2Part[ iObj ];
        Vec_IntPush( p.pvParts + p.pId2Part[ iObj ], iFanin );
        assert(Gia_ObjIsCi(pFanin) || Gia_ObjIsAnd(pFanin), "Gia_ObjIsCi(pFanin) || Gia_ObjIsAnd(pFanin)");
        if Gia_ObjIsAnd(pFanin) {
            Vec_IntPush( p.vFront, iFanin );
        } else if Gia_ObjIsRo(p.pGia, pFanin) {
            pFanin = Gia_ObjRoToRi(p.pGia, pFanin);
            iFanin = Gia_ObjId(p.pGia, pFanin);
            assert(p.pId2Part[ iFanin ] == -1, "p.pId2Part[ iFanin ] == -1");
            p.pId2Part[ iFanin ] = p.pId2Part[ iObj ];
            Vec_IntPush( p.pvParts + p.pId2Part[ iObj ], iFanin );
            Vec_IntPush( p.vFront, iFanin );
        }
    
    } else if p.pId2Part[ iObj ] != p.pId2Part[ iFanin ] {
        var vPartObj *Vec_Int_t  = p.pvParts + p.pId2Part[ iObj ];
        var vPartFan *Vec_Int_t  = p.pvParts + p.pId2Part[ iFanin ];
        var iTemp, i int32;
//        printf( "Moving %d to %d (%d -> %d)\n", iObj, iFanin, Vec_IntSize(vPartObj), Vec_IntSize(vPartFan) );
        // add group of iObj to group of iFanin
        assert(Vec_IntSize(vPartObj) > 0, "Vec_IntSize(vPartObj) > 0");
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(vPartObj)) && (((iTemp) = Vec_IntEntry(vPartObj, i)), 1); i++  {
            Vec_IntPush( vPartFan, iTemp );
            p.pId2Part[ iTemp ] = p.pId2Part[ iFanin ];
        }
        Vec_IntShrink( vPartObj, 0 );
        p.nParts--;
    }
}
func Opa_ManPerform(pGia *Gia_Man_t) {
    var p *Opa_Man_t;
    var pObj *Gia_Obj_t;
    var i, Limit, Count int32;
 
    p = Opa_ManStart( pGia );
    Limit = Vec_IntSize(p.vFront);
//Opa_ManPrint2( p );
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(p.vFront)) && ((pObj) = Gia_ManObj(pGia, Vec_IntEntry(p.vFront,i))); i++  {
        if i == Limit {
            fmt.Printf("%6d : %6d -> %6d\n", ++Count, i, p.nParts);
            Limit = Vec_IntSize(p.vFront);
            if Count > 1 {
                Opa_ManPrint2( p ); }
        }
//        printf( "*** Object %d  ", Gia_ObjId(pGia, pObj) );
        if Gia_ObjIsAnd(pObj) {
            Opa_ManMoveOne( p, pObj, Gia_ObjFanin0(pObj) );
            Opa_ManMoveOne( p, pObj, Gia_ObjFanin1(pObj) );
        
        } else if Gia_ObjIsCo(pObj) {
            Opa_ManMoveOne( p, pObj, Gia_ObjFanin0(pObj) );
        } else { assert(false, "0"); }
//        if ( i % 10 == 0 )
//            printf( "%d   ", p->nParts );
        if p.nParts == 1 {
            break; }
        if Count == 5 {
            break; }
    }
    fmt.Printf("\n");
    Opa_ManStop( p );
}



/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManConeMark_rec(p *Gia_Man_t, pObj *Gia_Obj_t, vRoots *Vec_Int_t, nLimit int32) int32 {
    if Gia_ObjIsTravIdCurrent(p, pObj) {
        return 0; }
    Gia_ObjSetTravIdCurrent(p, pObj);
    if Gia_ObjIsAnd(pObj) {
        if Gia_ManConeMark_rec( p, Gia_ObjFanin0(pObj), vRoots, nLimit ) != 0 {
            return 1; }
        if Gia_ManConeMark_rec( p, Gia_ObjFanin1(pObj), vRoots, nLimit ) != 0 {
            return 1; }
    
    } else if Gia_ObjIsCo(pObj) {
        if Gia_ManConeMark_rec( p, Gia_ObjFanin0(pObj), vRoots, nLimit ) != 0 {
            return 1; }
    
    } else if Gia_ObjIsRo(p, pObj) {
        Vec_IntPush( vRoots, Gia_ObjId(p, Gia_ObjRoToRi(p, pObj)) );
    } else if Gia_ObjIsPi(p, pObj) {
    } else { assert(false, "0"); }
    return int32((Vec_IntSize(vRoots) > nLimit));
}
func Gia_ManConeMark(p *Gia_Man_t, iOut int32, Limit int32) int32 {
    var vRoots *Vec_Int_t;
    var pObj *Gia_Obj_t;
    var i, RetValue int32;
    // start the outputs
    pObj = Gia_ManPo( p, iOut );
    vRoots = Vec_IntAlloc( 100 );
    Vec_IntPush( vRoots, Gia_ObjId(p, pObj) );
    // mark internal nodes
    Gia_ManIncrementTravId( p );
    Gia_ObjSetTravIdCurrent( p, Gia_ManConst0(p) );
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vRoots)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vRoots,i))); i++  {
        if Gia_ManConeMark_rec( p, pObj, vRoots, Limit ) != 0 {
            break; }
    }
    RetValue = Vec_IntSize( vRoots ) - 1;
    Vec_IntFree( vRoots );
    return RetValue;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCountFlops(p *Gia_Man_t, vOuts *Vec_Int_t) int32 {
    var Limit int32  = ABC_INFINITY;
    var vRoots *Vec_Int_t;
    var pObj *Gia_Obj_t;
    var i, RetValue, iOut int32;
    // start the outputs
    vRoots = Vec_IntAlloc( 100 );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vOuts)) && (((iOut) = Vec_IntEntry(vOuts, i)), 1); i++  {
        pObj = Gia_ManPo( p, iOut );
        Vec_IntPush( vRoots, Gia_ObjId(p, pObj) );
    }
    // mark internal nodes
    Gia_ManIncrementTravId( p );
    Gia_ObjSetTravIdCurrent( p, Gia_ManConst0(p) );
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vRoots)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vRoots,i))); i++  {
        if Gia_ManConeMark_rec( p, pObj, vRoots, Limit ) != 0 {
            break; }
    }
    RetValue = Vec_IntSize( vRoots ) - Vec_IntSize(vOuts);
    Vec_IntFree( vRoots );
    return RetValue;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManFindPoPartition3(p *Gia_Man_t, iOut int32, nDelta int32, nOutsMin int32, nOutsMax int32, fVerbose int32, pvPosEquivs **Vec_Ptr_t) *Gia_Man_t {
/*
    int i, Count = 0;
    // mark nodes belonging to output 'iOut'
    for ( i = 0; i < Gia_ManPoNum(p); i++ )
        Count += (Gia_ManConeMark(p, i, 10000) < 10000);
     //   printf( "%d ", Gia_ManConeMark(p, i, 1000) );
    printf( "%d out of %d\n", Count, Gia_ManPoNum(p) );

    // add other outputs as long as they are nDelta away
*/
//    Opa_ManPerform( p );

    return nil;
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManFindPivots(p *Gia_Man_t, SelectShift int32, fOnlyCis int32, fVerbose int32) *Vec_Int_t {
    var vPivots, vWeights *Vec_Int_t;
    var vCount, vResult *Vec_Int_t;
    var i int32;
    var j int32;
    var Count int32;
    var pPerm *int32;
    var Limit int32;
/*
    Gia_Obj_t * pObj;
    // count MUX controls
    vCount = Vec_IntStart( Gia_ManObjNum(p) );
    Gia_ManForEachAnd( p, pObj, i )
    {
        Gia_Obj_t * pNodeC, * pNodeT, * pNodeE;
        if ( !Gia_ObjIsMuxType(pObj) ) 
            continue;
        pNodeC = Gia_ObjRecognizeMux( pObj, &pNodeT, &pNodeE );
        Vec_IntAddToEntry( vCount, Gia_ObjId(p, Gia_Regular(pNodeC)), 1 );
    }
*/
    // count references
    Gia_ManCreateRefs( p );
    vCount = Vec_IntAllocArray( p.Refs, Gia_ManObjNum(p) ); p.Refs = nil;

    // collect nodes 
    vPivots  = Vec_IntAlloc( 100 );
    vWeights = Vec_IntAlloc( 100 );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vCount)) && (((Count) = Vec_IntEntry(vCount, i)), 1); i++  {
        if Count < 2 { continue; }
        if (fOnlyCis != 0) && !Gia_ObjIsCi(Gia_ManObj(p, i)) {
            continue; }
        Vec_IntPush( vPivots, i );
        Vec_IntPush( vWeights, Count );
    }
    Vec_IntFree( vCount );

    if fVerbose != 0 {
        fmt.Printf("Selected %d pivots with more than one fanout (out of %d CIs and ANDs).\n", Vec_IntSize(vWeights), Gia_ManCiNum(p) + Gia_ManAndNum(p)); }

    // permute
    Gia_ManRandom(1);
    Gia_ManRandom(0);
    for  i = 0; i < Vec_IntSize(vWeights); i++  {
        j = (Gia_ManRandom(0) >> 1) % Vec_IntSize(vWeights);
        // inlining of macro ABC_SWAP
        { var t int32  = vPivots.pArray[i]; vPivots.pArray[i] = vPivots.pArray[j]; vPivots.pArray[j] = t; };
        // inlining of macro ABC_SWAP
        { var t int32  = vWeights.pArray[i]; vWeights.pArray[i] = vWeights.pArray[j]; vWeights.pArray[j] = t; };
    }
    // sort
    if SelectShift == 0 {
        pPerm = Abc_QuickSortCost( Vec_IntArray(vWeights), Vec_IntSize(vWeights), 1 );
    } else {
        var vTemp *Vec_Int_t  = Vec_IntStartNatural( Vec_IntSize(vWeights) );
        pPerm = Vec_IntReleaseArray( vTemp );
        Vec_IntFree( vTemp );
    }

    // select    
    Limit = Abc_MinInt( 64, Vec_IntSize(vWeights) );
    vResult = Vec_IntAlloc( Limit );
    for  i = 0; i < Limit; i++  {
        j = (i + SelectShift) % Vec_IntSize(vWeights);
        if fVerbose != 0 {
            fmt.Printf("%2d : Pivot =%7d  Fanout =%7d\n", j, Vec_IntEntry(vPivots, pPerm[j]), Vec_IntEntry(vWeights, pPerm[j])); }
        Vec_IntPush( vResult, Vec_IntEntry(vPivots, pPerm[j]) );
    }

    Vec_IntFree( vPivots );
    Vec_IntFree( vWeights );
    // inlining of macro ABC_FREE
    pPerm = nil;

    return vResult;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDeriveSigns(p *Gia_Man_t, vPivots *Vec_Int_t, fVerbose int32) *Vec_Wrd_t {
    var vSigns *Vec_Wrd_t;
    var pObj, pObjRi *Gia_Obj_t;
    var i int32;
    var fChange int32 = 1;
    var Counter int32;

    Gia_ManFillValue( p );
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vPivots)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vPivots,i))); i++  {
        pObj.Value = i;
    }

    if fVerbose != 0 {
        fmt.Printf("Signature propagation: "); }
    vSigns = Vec_WrdStart( Gia_ManObjNum(p) );
    for  fChange != 0  {
        fChange = 0;
        // inlining of macro Gia_ManForEachObj
        for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
            if ~pObj.Value {
                assert(pObj.Value >= 0 && pObj.Value < 64, "pObj.Value >= 0 && pObj.Value < 64");
                *Vec_WrdEntryP( vSigns, i ) |= ( (word)1 << pObj.Value );
            }
            if Gia_ObjIsAnd(pObj) {
                *Vec_WrdEntryP( vSigns, i ) |= Vec_WrdEntry(vSigns, Gia_ObjFaninId0(pObj, i)) | Vec_WrdEntry(vSigns, Gia_ObjFaninId1(pObj, i));
            } else if Gia_ObjIsCo(pObj) {
                *Vec_WrdEntryP( vSigns, i ) |= Vec_WrdEntry(vSigns, Gia_ObjFaninId0(pObj, i)); }
        }
        Counter = 0;
        // inlining of macro Gia_ManForEachRiRo
        for  i = 0; (i < Gia_ManRegNum(p)) && ((pObjRi) = Gia_ManCo(p, Gia_ManPoNum(p)+i)) && ((pObj) = Gia_ManCi(p, Gia_ManPiNum(p)+i)); i++  {
            var Value word  = Vec_WrdEntry(vSigns, Gia_ObjId(p, pObj));
            *Vec_WrdEntryP( vSigns, Gia_ObjId(p, pObj) ) |= Vec_WrdEntry(vSigns, Gia_ObjId(p, pObjRi));
            if Value != Vec_WrdEntry(vSigns, Gia_ObjId(p, pObj)) {
                fChange = 1;
Counter++; }
        }
        if fVerbose != 0 {
            fmt.Printf("%d ", Counter); }
    }
    if fVerbose != 0 {
       fmt.Printf("\n"); }
    return vSigns;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManHashOutputs(p *Gia_Man_t, vSigns *Vec_Wrd_t, fVerbose int32) *Vec_Ptr_t {
    var vBins *Vec_Ptr_t;
    var vClasses *Vec_Wec_t;
    var vSignsPo *Vec_Wrd_t;
    var vPriority, vBin *Vec_Int_t;
    var pObj *Gia_Obj_t;
    var i int32;
    // collect PO signatures
    vSignsPo = Vec_WrdAlloc( Gia_ManPoNum(p) );
    // inlining of macro Gia_ManForEachPo
    for  i = 0; (i < Gia_ManPoNum(p)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        Vec_WrdPush( vSignsPo, Vec_WrdEntry(vSigns, Gia_ObjId(p, pObj)) );
    }
    // find equivalence classes
    vPriority = Hsh_WrdManHashArray( vSignsPo, 1 );
    Vec_WrdFree( vSignsPo );
    vClasses = Vec_WecCreateClasses( vPriority );
    Vec_IntFree( vPriority );
    vBins = (Vec_Ptr_t *)Vec_WecConvertToVecPtr( vClasses );
    Vec_WecFree( vClasses );
    Vec_VecSort( (Vec_Vec_t *)vBins, 1 );

    if fVerbose != 0 {
        fmt.Printf("Computed %d partitions:\n", Vec_PtrSize(vBins)); }
    if !(fVerbose != 0) {
        fmt.Printf("Listing partitions with more than 100 outputs:\n"); }
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vBins)) && (((vBin) = (Vec_Int_t *)Vec_PtrEntry(vBins, i)), 1); i++  {
        assert(Vec_IntSize(vBin) > 0, "Vec_IntSize(vBin) > 0");
        if (fVerbose != 0) || Vec_IntSize(vBin) > 100 {
            var PoNum int32  = Vec_IntEntry( vBin, 0 );
            var pObj *Gia_Obj_t  = Gia_ManPo( p, PoNum );
            var Sign word  = Vec_WrdEntry( vSigns, Gia_ObjId(p, pObj) );
            // print
            fmt.Printf("%3d ", i);
            Extra_PrintBinary( os.Stdout, (unsigned *)&Sign, 64 );
            fmt.Printf("  ");
            fmt.Printf("PO =%7d  ", Vec_IntSize(vBin));
            fmt.Printf("FF =%7d", Gia_ManCountFlops(p, vBin));
            fmt.Printf("\n");
        }
    }
    return vBins;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManFindPoPartition2(p *Gia_Man_t, iStartNum int32, nDelta int32, nOutsMin int32, nOutsMax int32, fSetLargest int32, fVerbose int32, pvPosEquivs **Vec_Ptr_t) *Gia_Man_t {
    return nil;
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManFindPoPartition(p *Gia_Man_t, SelectShift int32, fOnlyCis int32, fSetLargest int32, fVerbose int32, pvPosEquivs **Vec_Ptr_t) *Gia_Man_t {
    var pGia *Gia_Man_t  = nil;
    var vPivots *Vec_Int_t;
    var vSigns *Vec_Wrd_t;
    var vParts *Vec_Ptr_t;
    var vPart *Vec_Int_t;
    var clk abctime  = Abc_Clock();
    vPivots = Gia_ManFindPivots( p, SelectShift, fOnlyCis, fVerbose );
    vSigns = Gia_ManDeriveSigns( p, vPivots, fVerbose );
    Vec_IntFree( vPivots );
    vParts = Gia_ManHashOutputs( p, vSigns, fVerbose );
    Vec_WrdFree( vSigns );
    if fSetLargest != 0 {
        vPart = Vec_VecEntryInt( (Vec_Vec_t *)vParts, 0 );
        pGia = Gia_ManDupCones( p, Vec_IntArray(vPart), Vec_IntSize(vPart), 1 );
    }
    if pvPosEquivs {
        *pvPosEquivs = vParts;
        fmt.Printf("The algorithm divided %d POs into %d partitions.   ", Gia_ManPoNum(p), Vec_PtrSize(vParts));
        Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
    
    } else {
        Vec_VecFree( (Vec_Vec_t *)vParts ); }
    return pGia;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

