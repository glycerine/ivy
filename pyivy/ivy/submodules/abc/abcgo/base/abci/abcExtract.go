package abci

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [abcShare.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Network and node package.]

  Synopsis    [Shared logic extraction.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: abcShare.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "base/abc/abc.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

const SHARE_NUM = 2

   
type Abc_ShaMan_t struct {
    nMultiSize             int32;
    fVerbose             int32;
    pNtk      *Abc_Ntk_t;
    vBuckets      *Vec_Ptr_t;
    vObj2Lit      *Vec_Int_t;
    nStartCols             int32;
    nCountGates             int32;
    nFoundGates             int32;
}

func Abc_NtkSharePack(Lev int32, Id int32) word { return (((word)Lev) << 32) | Id; }
func Abc_NtkShareUnpackLev(Num word) int32 { return (Num >> 32);              }
func Abc_NtkShareUnpackId(Num word) int32 { return Num & 0xFFFFFFFF;         }


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Working with the manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_ShaManStart(pNtk *Abc_Ntk_t) *Abc_ShaMan_t {
    
    // inlining of macro ABC_CALLOC
    p := &Abc_ShaMan_t{};
    p.pNtk      = pNtk;
    p.vObj2Lit  = Vec_IntAlloc( 1000 );
    return p;
}
func Abc_ShaManStop(p *Abc_ShaMan_t) {
    var vBucket *Vec_Ptr_t;
    var i int32;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vBuckets)) && (((vBucket) = (Vec_Ptr_t *)Vec_PtrEntry(p.vBuckets, i)), 1); i++  {
        Vec_VecFree( (Vec_Vec_t *)vBucket );
    }
    Vec_PtrFreeP( &p.vBuckets );
    Vec_IntFreeP( &p.vObj2Lit );
    p = nil;
}


/**Function*************************************************************

  Synopsis    [Collects one multi-input gate.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkShareSuperXor(pObj *Abc_Obj_t, pfCompl *int32, pCounter *int32) *Vec_Wrd_t {
    var pNtk *Abc_Ntk_t  = Abc_ObjNtk(pObj);
    var pObjC *Abc_Obj_t;
    var pObj0 *Abc_Obj_t;
    var pObj1 *Abc_Obj_t;
    var pRoot *Abc_Obj_t = nil;
    var vSuper *Vec_Wrd_t;
    var Num, NumNext word;
    var i, k, fCompl int32;
    assert(!Abc_ObjIsComplement(pObj), "!Abc_ObjIsComplement(pObj)");
    assert(Abc_NodeIsExorType(pObj), "Abc_NodeIsExorType(pObj)");
    // start iteration
    vSuper = Vec_WrdAlloc( 10 );
    Vec_WrdPush( vSuper, Abc_NtkSharePack(Abc_ObjLevel(pObj), Abc_ObjId(pObj)) );
    for  Vec_WrdSize(vSuper) > 0  {
        // make sure there are no duplicates
        Num = Vec_WrdEntry( vSuper, 0 );
        // inlining of macro Vec_WrdForEachEntryStart
        for  i = 1; (i < Vec_WrdSize(vSuper)) && (((NumNext) = Vec_WrdEntry(vSuper, i)), 1); i++  {
            assert(Num < NumNext, "Num < NumNext");
            Num = NumNext;
        }
        // extract XOR gate decomposable on the topmost level
        // inlining of macro Vec_WrdForEachEntryReverse
        for  i = Vec_WrdSize(vSuper) - 1; (i >= 0) && (((Num) = Vec_WrdEntry(vSuper, i)), 1); i--  {
            pRoot = Abc_NtkObj( pNtk, Abc_NtkShareUnpackId(Num) );
            if Abc_NodeIsExorType(pRoot) {
                Vec_WrdRemove( vSuper, Num );
                break;
            }
        }
        if i == -1 {
            break; }
        // extract
        pObjC = Abc_NodeRecognizeMux( pRoot, &pObj1, &pObj0 );
        assert(pObj1 == Abc_ObjNot(pObj0), "pObj1 == Abc_ObjNot(pObj0)");
        fCompl ^= Abc_ObjIsComplement(pObjC);  pObjC = Abc_ObjRegular(pObjC);
        fCompl ^= Abc_ObjIsComplement(pObj0);  pObj0 = Abc_ObjRegular(pObj0);
        Vec_WrdPushOrder( vSuper, Abc_NtkSharePack(Abc_ObjLevel(pObjC), Abc_ObjId(pObjC)) );
        Vec_WrdPushOrder( vSuper, Abc_NtkSharePack(Abc_ObjLevel(pObj0), Abc_ObjId(pObj0)) );
        (*pCounter)++;
        // remove duplicates
        k = 0;
        // inlining of macro Vec_WrdForEachEntry
        for  i = 0; (i < Vec_WrdSize(vSuper)) && (((Num) = Vec_WrdEntry(vSuper, i)), 1); i++  {
            if i + 1 == Vec_WrdSize(vSuper) {
                Vec_WrdWriteEntry( vSuper, k++, Num );
                break;
            }
            NumNext = Vec_WrdEntry( vSuper, i+1 );
            assert(Num <= NumNext, "Num <= NumNext");
            if Num == NumNext {
                i++;
            } else {
                Vec_WrdWriteEntry( vSuper, k++, Num ); }
        }
        Vec_WrdShrink( vSuper, k );
    }
    *pfCompl = fCompl;
    // inlining of macro Vec_WrdForEachEntry
    for  i = 0; (i < Vec_WrdSize(vSuper)) && (((Num) = Vec_WrdEntry(vSuper, i)), 1); i++  {
        Vec_WrdWriteEntry( vSuper, i, Abc_NtkShareUnpackId(Num) );
    }
    return vSuper;
}
func Abc_NtkShareSuperAnd(pObj *Abc_Obj_t, pCounter *int32) *Vec_Wrd_t {
    var pNtk *Abc_Ntk_t  = Abc_ObjNtk(pObj);
    var pObj0 *Abc_Obj_t;
    var pObj1 *Abc_Obj_t;
    var pRoot *Abc_Obj_t = nil;
    var vSuper *Vec_Wrd_t;
    var Num, NumNext word;
    var i, k int32;
    assert(!Abc_ObjIsComplement(pObj), "!Abc_ObjIsComplement(pObj)");
    // start iteration
    vSuper = Vec_WrdAlloc( 10 );
    Vec_WrdPush( vSuper, Abc_NtkSharePack(Abc_ObjLevel(pObj), Abc_ObjToLit(pObj)) );
    for  Vec_WrdSize(vSuper) > 0  {
        // make sure there are no duplicates
        Num = Vec_WrdEntry( vSuper, 0 );
        // inlining of macro Vec_WrdForEachEntryStart
        for  i = 1; (i < Vec_WrdSize(vSuper)) && (((NumNext) = Vec_WrdEntry(vSuper, i)), 1); i++  {
            assert(Num < NumNext, "Num < NumNext");
            Num = NumNext;
        }
        // extract AND gate decomposable on the topmost level
        // inlining of macro Vec_WrdForEachEntryReverse
        for  i = Vec_WrdSize(vSuper) - 1; (i >= 0) && (((Num) = Vec_WrdEntry(vSuper, i)), 1); i--  {
            pRoot = Abc_ObjFromLit( pNtk, Abc_NtkShareUnpackId(Num) );
            if !Abc_ObjIsComplement(pRoot) && Abc_ObjIsNode(pRoot) {
                Vec_WrdRemove( vSuper, Num );
                break;
            }
        }
        if i == -1 {
            break; }
        assert(Abc_ObjIsNode(pRoot), "Abc_ObjIsNode(pRoot)");
        // extract
        pObj0 = Abc_ObjChild0(pRoot);
        pObj1 = Abc_ObjChild1(pRoot);
        assert(Abc_ObjIsNode(Abc_ObjRegular(pObj0)) || Abc_ObjIsCi(Abc_ObjRegular(pObj0)), "Abc_ObjIsNode(Abc_ObjRegular(pObj0)) || Abc_ObjIsCi(Abc_ObjRegular(pObj0))");
        assert(Abc_ObjIsNode(Abc_ObjRegular(pObj1)) || Abc_ObjIsCi(Abc_ObjRegular(pObj1)), "Abc_ObjIsNode(Abc_ObjRegular(pObj1)) || Abc_ObjIsCi(Abc_ObjRegular(pObj1))");
        Vec_WrdPushOrder( vSuper, Abc_NtkSharePack(Abc_ObjLevel(Abc_ObjRegular(pObj0)), Abc_ObjToLit(pObj0)) );
        Vec_WrdPushOrder( vSuper, Abc_NtkSharePack(Abc_ObjLevel(Abc_ObjRegular(pObj1)), Abc_ObjToLit(pObj1)) );
        (*pCounter)++;
        // remove duplicates
        k = 0;
        // inlining of macro Vec_WrdForEachEntry
        for  i = 0; (i < Vec_WrdSize(vSuper)) && (((Num) = Vec_WrdEntry(vSuper, i)), 1); i++  {
            if i + 1 == Vec_WrdSize(vSuper) {
                Vec_WrdWriteEntry( vSuper, k++, Num );
                break;
            }
            NumNext = Vec_WrdEntry( vSuper, i+1 );
            assert(Num <= NumNext, "Num <= NumNext");
            if Num + 1 == NumNext && (NumNext & 1) { // pos_lit & neg_lit = 0
            
                Vec_WrdClear( vSuper );
                return vSuper;
            }
            if Num < NumNext {
                Vec_WrdWriteEntry( vSuper, k++, Num ); }
        }
        Vec_WrdShrink( vSuper, k );
    }
    // inlining of macro Vec_WrdForEachEntry
    for  i = 0; (i < Vec_WrdSize(vSuper)) && (((Num) = Vec_WrdEntry(vSuper, i)), 1); i++  {
        Vec_WrdWriteEntry( vSuper, i, Abc_NtkShareUnpackId(Num) );
    }
    return vSuper;
}

/**Function*************************************************************

  Synopsis    [Creates multi-input XOR representation for the nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkTraverseSupersXor_rec(p *Abc_ShaMan_t, pObj *Abc_Obj_t, vInputs *Vec_Ptr_t) {
    if Abc_NodeIsTravIdCurrent( pObj ) {
        return; }
    Abc_NodeSetTravIdCurrent( pObj );
    if Abc_ObjIsCi(pObj) {
        return; }
    assert(Abc_ObjIsNode(pObj), "Abc_ObjIsNode(pObj)");
    if Abc_NodeIsExorType(pObj) {
        var vSuper *Vec_Wrd_t;
        var k, fCompl int32;
        var Num word;
        vSuper = Abc_NtkShareSuperXor( pObj, &fCompl, &p.nFoundGates );
        if Vec_WrdSize(vSuper) <= 1 || Vec_WrdSize(vSuper) >= p.nMultiSize {
            // inlining of macro Vec_WrdForEachEntry
            for  k = 0; (k < Vec_WrdSize(vSuper)) && (((Num) = Vec_WrdEntry(vSuper, k)), 1); k++  {
                var vInput *Vec_Int_t  = (Vec_Int_t *)Vec_PtrEntry( vInputs, int32(Num) );
                if vInput == nil {
                    vInput = Vec_IntAlloc( 10 );
                    Vec_IntPush( vInput, Abc_Var2Lit(int32(Num), 0) );
                    Vec_IntPush( vInput, Abc_ObjLevel(Abc_NtkObj(p.pNtk, int32(Num))) );
                    assert(SHARE_NUM == Vec_IntSize(vInput), "SHARE_NUM == Vec_IntSize(vInput)");
                    Vec_PtrWriteEntry( vInputs, int32(Num), vInput );
                }
                Vec_IntPush( vInput, Vec_IntSize(p.vObj2Lit) );
            }
            Vec_IntPush( p.vObj2Lit, Abc_Var2Lit(Abc_ObjId(pObj), fCompl) );
        }
        // call recursively
        // inlining of macro Vec_WrdForEachEntry
        for  k = 0; (k < Vec_WrdSize(vSuper)) && (((Num) = Vec_WrdEntry(vSuper, k)), 1); k++  {
            Abc_NtkTraverseSupersXor_rec( p, Abc_NtkObj(p.pNtk, int32(Num)), vInputs );
        }
        Vec_WrdFree( vSuper );
    
    } else {
        Abc_NtkTraverseSupersXor_rec( p, Abc_ObjFanin0(pObj), vInputs );
        Abc_NtkTraverseSupersXor_rec( p, Abc_ObjFanin1(pObj), vInputs );
    }
}
func Abc_NtkTraverseSupersAnd_rec(p *Abc_ShaMan_t, pObj *Abc_Obj_t, vInputs *Vec_Ptr_t) {
    var vSuper *Vec_Wrd_t;
    var Num word;
    var k int32;
    if Abc_NodeIsTravIdCurrent( pObj ) {
        return; }
    Abc_NodeSetTravIdCurrent( pObj );
    if Abc_ObjIsCi(pObj) {
        return; }
    assert(Abc_ObjIsNode(pObj), "Abc_ObjIsNode(pObj)");
    vSuper = Abc_NtkShareSuperAnd( pObj, &p.nFoundGates );
    if Vec_WrdSize(vSuper) <= 1 || Vec_WrdSize(vSuper) >= p.nMultiSize {
        // inlining of macro Vec_WrdForEachEntry
        for  k = 0; (k < Vec_WrdSize(vSuper)) && (((Num) = Vec_WrdEntry(vSuper, k)), 1); k++  {
            var vInput *Vec_Int_t  = (Vec_Int_t *)Vec_PtrEntry( vInputs, int32(Num) );
            if vInput == nil {
                vInput = Vec_IntAlloc( 10 );
                Vec_IntPush( vInput, int32(Num) );
                Vec_IntPush( vInput, Abc_ObjLevel(Abc_NtkObj(p.pNtk, Abc_Lit2Var(int32(Num)))) );
                assert(SHARE_NUM == Vec_IntSize(vInput), "SHARE_NUM == Vec_IntSize(vInput)");
                Vec_PtrWriteEntry( vInputs, int32(Num), vInput );
            }
            Vec_IntPush( vInput, Vec_IntSize(p.vObj2Lit) );
        }
        Vec_IntPush( p.vObj2Lit, Abc_ObjToLit(pObj) );
    }
    // call recursively
    // inlining of macro Vec_WrdForEachEntry
    for  k = 0; (k < Vec_WrdSize(vSuper)) && (((Num) = Vec_WrdEntry(vSuper, k)), 1); k++  {
        Abc_NtkTraverseSupersAnd_rec( p, Abc_NtkObj(p.pNtk, Abc_Lit2Var(int32(Num))), vInputs );
    }
    Vec_WrdFree( vSuper );
}
func Abc_NtkTraverseSupers(p *Abc_ShaMan_t, fAnd int32) {
    var vInputs *Vec_Ptr_t;
    var vInput *Vec_Int_t;
    var pObj *Abc_Obj_t;
    var i, nOnesMax int32;

    // create mapping of nodes into their column vectors
    vInputs = Vec_PtrStart( Abc_NtkObjNumMax(p.pNtk) * (1 + fAnd) );
    Abc_NtkIncrementTravId( p.pNtk );
    if fAnd != 0 {
        // inlining of macro Abc_NtkForEachCo
        for  i = 0; (i < Abc_NtkCoNum(p.pNtk)) && (((pObj) = Abc_NtkCo(p.pNtk, i)), 1); i++  {
            if Abc_ObjIsNode(Abc_ObjFanin0(pObj)) {
                Abc_NtkTraverseSupersAnd_rec( p, Abc_ObjFanin0(pObj), vInputs ); }
        }
    
    } else {
        // inlining of macro Abc_NtkForEachCo
        for  i = 0; (i < Abc_NtkCoNum(p.pNtk)) && (((pObj) = Abc_NtkCo(p.pNtk, i)), 1); i++  {
            if Abc_ObjIsNode(Abc_ObjFanin0(pObj)) {
                Abc_NtkTraverseSupersXor_rec( p, Abc_ObjFanin0(pObj), vInputs ); }
        }
    }
    p.nStartCols = Vec_IntSize(p.vObj2Lit);

    // find the largest number of 1s
    nOnesMax = 0;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vInputs)) && (((vInput) = (Vec_Int_t *)Vec_PtrEntry(vInputs, i)), 1); i++  {
        if vInput != nil {
            nOnesMax = Abc_MaxInt( nOnesMax, Vec_IntSize(vInput)-SHARE_NUM ); }
    }

    // create buckets
    assert(p.vBuckets == nil, "p.vBuckets == nil");
    p.vBuckets = Vec_PtrAlloc( nOnesMax + 1 );
    for  i = 0; i <= nOnesMax; i++  {
        Vec_PtrPush( p.vBuckets, Vec_PtrAlloc(10) );
    }

    // load vectors into buckets
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vInputs)) && (((vInput) = (Vec_Int_t *)Vec_PtrEntry(vInputs, i)), 1); i++  {
        if vInput != nil {
            Vec_PtrPush( (Vec_Ptr_t *)Vec_PtrEntry(p.vBuckets, Vec_IntSize(vInput)-SHARE_NUM), vInput ); }
    }
    Vec_PtrFree( vInputs );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkSharePrint(p *Abc_ShaMan_t) {
    var vBucket *Vec_Ptr_t;
    var vInput *Vec_Int_t;
    var i, k, j, ObjId int32;
    // inlining of macro ABC_ALLOC
    var pBuffer string  = ((char *) malloc(sizeof(char) * (size_t)(Vec_IntSize(p.vObj2Lit) + 1)));
    // inlining of macro ABC_CALLOC
    var pCounters *int32  = ((int *) calloc((size_t)(Vec_IntSize(p.vObj2Lit) + 1), sizeof(int)));
    var nTotal int32;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vBuckets)) && (((vBucket) = (Vec_Ptr_t *)Vec_PtrEntry(p.vBuckets, i)), 1); i++  {
    // inlining of macro Vec_PtrForEachEntry
    for  j = 0; (j < Vec_PtrSize(vBucket)) && (((vInput) = (Vec_Int_t *)Vec_PtrEntry(vBucket, j)), 1); j++  {
        for  k = 0; k < Vec_IntSize(p.vObj2Lit); k++  {
            pBuffer[k] = '0';
        }
        pBuffer[k] = 0;

        // inlining of macro Vec_IntForEachEntryStart
        for  k = SHARE_NUM; (k < Vec_IntSize(vInput)) && (((ObjId) = Vec_IntEntry(vInput, k)), 1); k++  {
            assert(ObjId < Vec_IntSize(p.vObj2Lit), "ObjId < Vec_IntSize(p.vObj2Lit)");
            pBuffer[ObjId] = '1';
            pCounters[ObjId]++;
        }
        fmt.Printf("%4d%3d: %s\n", Vec_IntEntry(vInput, 0), Vec_IntEntry(vInput, 1), pBuffer);
    }
    }

    for  i = 0; i < Vec_IntSize(p.vObj2Lit); i++  {
        if pCounters[i] > 0 {
            fmt.Printf("%d=%d ", i, pCounters[i]); }
    } 
    fmt.Printf("\n");

    nTotal = 0;
    for  i = 0; i < p.nStartCols; i++  {
        nTotal += pCounters[i] - 1;
    }
    fmt.Printf("Total = %d.  ", nTotal);
    fmt.Printf("Gates = %d.\n", Vec_IntSize(p.vObj2Lit) - p.nStartCols + nTotal);

    pCounters = nil;
    pBuffer = nil;

    fmt.Printf("Bucket contents: ");
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vBuckets)) && (((vBucket) = (Vec_Ptr_t *)Vec_PtrEntry(p.vBuckets, i)), 1); i++  {
        fmt.Printf("%d ", Vec_PtrSize(vBucket));
    }
    fmt.Printf("\n");
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDumpBlif(p *Abc_Ntk_t) {
    var pFile *os.File;
    var vSupp *Vec_Ptr_t;
    var pObj *Abc_Obj_t;
    var i, k int32;
    pFile = os.OpenFile("multi_and.blif", os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    if pFile == nil {
        fmt.Printf("Cannot open output file.\n");
        return;
    }
    fmt.Fprintf(pFile, ".model %s\n", "multi_and");
    fmt.Fprintf(pFile, ".inputs");
    for  i = 0; i < Abc_NtkCiNum(p); i++  {
        fmt.Fprintf(pFile, " i%d", i);
    }
    fmt.Fprintf(pFile, "\n");
    fmt.Fprintf(pFile, ".outputs");
    for  i = 0; i < Abc_NtkCoNum(p); i++  {
        fmt.Fprintf(pFile, " o%d", i);
    }
    fmt.Fprintf(pFile, "\n");
    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(p)) && (((pObj) = Abc_NtkCi(p, i)), 1); i++  {
        pObj.iTemp = i;
    }
    for  i = 0; i < Abc_NtkCoNum(p); i++  {
        pObj = Abc_NtkCo( p, i );
        vSupp = Abc_NtkNodeSupport( p, &pObj, 1 );
        fmt.Fprintf(pFile, ".names");
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(vSupp)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vSupp, k)), 1); k++  {
            fmt.Fprintf(pFile, " i%d", pObj.iTemp);
        }
        fmt.Fprintf(pFile, " o%d\n", i);
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(vSupp)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vSupp, k)), 1); k++  {
            fmt.Fprintf(pFile, "1");
        }
        fmt.Fprintf(pFile, " 1\n");
        Vec_PtrFree( vSupp );
    }
    fmt.Fprintf(pFile, ".end\n\n");
    _ = (pFile).Close();
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkShareFindBestMatch(vBuckets *Vec_Ptr_t, pvInput **Vec_Int_t, pvInput2 **Vec_Int_t) {
    var nPoolSize int32  = 40;
    var vPool *Vec_Ptr_t  = Vec_PtrAlloc( nPoolSize );
    var vBucket *Vec_Ptr_t;
    var vInput *Vec_Int_t;
    var vInput2 *Vec_Int_t;
    var vInputBest *Vec_Int_t = nil;
    var vInputBest2 *Vec_Int_t = nil;
    var i, k, Cost, CostBest, Delay, DelayBest int32;

    // inlining of macro Vec_PtrForEachEntryReverse
    for  i = Vec_PtrSize(vBuckets) - 1; (i >= 0) && (((vBucket) = (Vec_Ptr_t *)Vec_PtrEntry(vBuckets, i)), 1); i--  {
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(vBucket)) && (((vInput) = (Vec_Int_t *)Vec_PtrEntry(vBucket, k)), 1); k++  {
            Vec_PtrPush( vPool, vInput );
            if Vec_PtrSize(vPool) == nPoolSize {
                goto outside; }
        }
    }
outside:

    // inlining of macro Vec_PtrForEachEntryReverse
    for  i = Vec_PtrSize(vPool) - 1; (i >= 0) && (((vInput) = (Vec_Int_t *)Vec_PtrEntry(vPool, i)), 1); i--  {
    // inlining of macro Vec_PtrForEachEntryReverse
    for  k = Vec_PtrSize(vPool) - 1; (k >= 0) && (((vInput2) = (Vec_Int_t *)Vec_PtrEntry(vPool, k)), 1); k--  {
        if i == k {
            continue; }

        vInput.pArray += SHARE_NUM;
        vInput2.pArray += SHARE_NUM;
        vInput.nSize -= SHARE_NUM;
        vInput2.nSize -= SHARE_NUM;

        Cost = Vec_IntTwoCountCommon(vInput, vInput2);

        vInput.pArray -= SHARE_NUM;
        vInput2.pArray -= SHARE_NUM;
        vInput.nSize += SHARE_NUM;
        vInput2.nSize += SHARE_NUM;

        if Cost < 2 {
            continue; }

        Delay = Abc_MaxInt( Vec_IntEntry(vInput, 1), Vec_IntEntry(vInput2, 1) );

        if CostBest < Cost || (CostBest == Cost && (DelayBest > Delay)) {
            CostBest = Cost;
            DelayBest = Delay;
            vInputBest = vInput;
            vInputBest2 = vInput2;
        }
    }
    }
    Vec_PtrFree( vPool );

    *pvInput  = vInputBest;
    *pvInput2 = vInputBest2;

    if vInputBest == nil {
        return; }

    Vec_PtrRemove( (Vec_Ptr_t *)Vec_PtrEntry(vBuckets, Vec_IntSize(vInputBest)-SHARE_NUM),  (Vec_Int_t *)vInputBest ); 
    Vec_PtrRemove( (Vec_Ptr_t *)Vec_PtrEntry(vBuckets, Vec_IntSize(vInputBest2)-SHARE_NUM), (Vec_Int_t *)vInputBest2 ); 
}
func Abc_NtkShareOptimize(p *Abc_ShaMan_t, fAnd int32) {
    var pObj, pObj0, pObj1 *Abc_Obj_t;
    var vInput, vInput2 *Vec_Int_t;
    var vNew, vOld1, vOld2 *Vec_Int_t;
    var i int32;
    for  i = 0; ; i++  {
        Abc_NtkShareFindBestMatch( p.vBuckets, &vInput, &vInput2 );
        if vInput == nil {
            break; }

        // create new node
        pObj0 = Abc_ObjFromLit( p.pNtk, Vec_IntEntry(vInput,  0) );
        pObj1 = Abc_ObjFromLit( p.pNtk, Vec_IntEntry(vInput2, 0) );
        if fAnd != 0 {
            pObj  = Abc_AigAnd( (Abc_Aig_t *)p.pNtk.pManFunc, pObj0, pObj1 );
        } else {
            pObj  = Abc_AigXor( (Abc_Aig_t *)p.pNtk.pManFunc, pObj0, pObj1 ); }
        p.nCountGates++;

        // save new node
        vOld1 = Vec_IntAlloc( 16 );  Vec_IntPush( vOld1, Vec_IntEntry(vInput,  0) );  Vec_IntPush( vOld1, Vec_IntEntry(vInput,  1) );
        vOld2 = Vec_IntAlloc( 16 );  Vec_IntPush( vOld2, Vec_IntEntry(vInput2, 0) );  Vec_IntPush( vOld2, Vec_IntEntry(vInput2, 1) );
        vNew  = Vec_IntAlloc( 16 );  Vec_IntPush( vNew,  Abc_ObjToLit(pObj) );        Vec_IntPush( vNew, Abc_ObjLevel(Abc_ObjRegular(pObj)) );

        // compute new arrays
        vInput.pArray += SHARE_NUM;
        vInput2.pArray += SHARE_NUM;
        vInput.nSize -= SHARE_NUM;
        vInput2.nSize -= SHARE_NUM;

        Vec_IntTwoSplit( vInput, vInput2, vNew, vOld1, vOld2 );

        vInput.pArray -= SHARE_NUM;
        vInput2.pArray -= SHARE_NUM;
        vInput.nSize += SHARE_NUM;
        vInput2.nSize += SHARE_NUM;

        // add to the old ones
        Vec_IntPush( vOld1, Vec_IntSize(p.vObj2Lit) );
        Vec_IntPush( vOld2, Vec_IntSize(p.vObj2Lit) );
        Vec_IntPush( p.vObj2Lit, Abc_ObjToLit(pObj) );

        Vec_PtrPush( (Vec_Ptr_t *)Vec_PtrEntry(p.vBuckets, Vec_IntSize(vOld1)-SHARE_NUM), vOld1 );
        Vec_PtrPush( (Vec_Ptr_t *)Vec_PtrEntry(p.vBuckets, Vec_IntSize(vOld2)-SHARE_NUM), vOld2 );
        Vec_PtrPush( (Vec_Ptr_t *)Vec_PtrEntry(p.vBuckets, Vec_IntSize(vNew)-SHARE_NUM), vNew );

        Vec_IntFree( vInput );
        Vec_IntFree( vInput2 );
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkUpdateNetwork(p *Abc_ShaMan_t, fAnd int32) *Abc_Ntk_t {
    var pNtk *Abc_Ntk_t;
    var vInput, vMap2Repl *Vec_Int_t;
    var vOrig, vRepl, vBucket *Vec_Ptr_t;
    var pObj, pNew *Abc_Obj_t;
    var i, j, k, ObjId, iLit int32;
    var iLitConst1 int32  = Abc_ObjToLit( Abc_AigConst1(p.pNtk) );

    vOrig = Vec_PtrAlloc( p.nStartCols );
    vRepl = Vec_PtrAlloc( p.nStartCols );
    for  i = 0; i < p.nStartCols; i++  {
        iLit = Vec_IntEntry( p.vObj2Lit, i );
        assert(!(fAnd != 0) || !Abc_LitIsCompl(iLit), "!fAnd || !Abc_LitIsCompl(iLit)");

        pObj = Abc_NtkObj( p.pNtk, Abc_Lit2Var(iLit) );

        if fAnd != 0 {
            pNew = Abc_AigConst1(p.pNtk);
        } else {
            pNew = Abc_ObjNotCond( Abc_AigConst1(p.pNtk), !Abc_LitIsCompl(iLit) ); }

        Vec_PtrPush( vOrig, pObj );
        Vec_PtrPush( vRepl, pNew );

        p.nCountGates--;
    }

    // go through the columns
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vBuckets)) && (((vBucket) = (Vec_Ptr_t *)Vec_PtrEntry(p.vBuckets, i)), 1); i++  {
    // inlining of macro Vec_PtrForEachEntry
    for  j = 0; (j < Vec_PtrSize(vBucket)) && (((vInput) = (Vec_Int_t *)Vec_PtrEntry(vBucket, j)), 1); j++  {
        // inlining of macro Vec_IntForEachEntryStart
        for  k = SHARE_NUM; (k < Vec_IntSize(vInput)) && (((ObjId) = Vec_IntEntry(vInput, k)), 1); k++  {
            assert(ObjId < Vec_IntSize(p.vObj2Lit), "ObjId < Vec_IntSize(p.vObj2Lit)");
            if ObjId >= p.nStartCols {
                break; }
            assert(ObjId < p.nStartCols, "ObjId < p.nStartCols");
            iLit = Vec_IntEntry( vInput, 0 );

            pNew = (Abc_Obj_t *)Vec_PtrEntry( vRepl, ObjId );
            if fAnd != 0 {
                pNew = Abc_AigAnd( (Abc_Aig_t *)p.pNtk.pManFunc, pNew, Abc_ObjFromLit(p.pNtk, iLit) );
            } else {
                pNew = Abc_AigXor( (Abc_Aig_t *)p.pNtk.pManFunc, pNew, Abc_ObjFromLit(p.pNtk, iLit) ); }
            Vec_PtrWriteEntry( vRepl, ObjId, pNew );
            p.nCountGates++;
        }
    }
    }

    if p.fVerbose {
        fmt.Printf("Total gates collected = %d.  Total gates constructed = %d.\n", p.nFoundGates, p.nCountGates); }

    // create map of originals
    vMap2Repl = Vec_IntStartFull( Abc_NtkObjNumMax(p.pNtk) );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vOrig)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vOrig, i)), 1); i++  {
//        printf( "Replacing %d by %d.\n", Abc_ObjId(pObj), Abc_ObjToLit((Abc_Obj_t *)Vec_PtrEntry(vRepl, i)) );
        Vec_IntWriteEntry( vMap2Repl, Abc_ObjId(pObj), Abc_ObjToLit((Abc_Obj_t *)Vec_PtrEntry(vRepl, i)) );
    }
    Vec_PtrFree( vOrig );
    Vec_PtrFree( vRepl );

    // update fanin pointers
    // inlining of macro Abc_NtkForEachObj
    for  i = 0; (i < Vec_PtrSize((p.pNtk).vObjs)) && (((pObj) = Abc_NtkObj(p.pNtk, i)), 1); i++  {            if (pObj) == nil {} else {
        if Abc_ObjIsCo(pObj) || Abc_ObjIsNode(pObj) {
            iLit = Vec_IntEntry( vMap2Repl, Abc_ObjFaninId0(pObj) );
            if iLit >= 0 {
                if iLit == iLitConst1 && (fAnd != 0) {
                    pObj.fCompl0 ^= 1;
                    Vec_IntWriteEntry( &pObj.vFanins, 0, Abc_Lit2Var(iLitConst1) );
                
                } else {
                    pObj.fCompl0 ^= Abc_LitIsCompl(iLit);
                    Vec_IntWriteEntry( &pObj.vFanins, 0, Abc_Lit2Var(iLit) );
                }
            }
        }
        if Abc_ObjIsNode(pObj) {
            iLit = Vec_IntEntry( vMap2Repl, Abc_ObjFaninId1(pObj) );
            if iLit >= 0 {
                if iLit == iLitConst1 && (fAnd != 0) {
                    pObj.fCompl1 ^= 1;
                    Vec_IntWriteEntry( &pObj.vFanins, 1, Abc_Lit2Var(iLitConst1) );
                
                } else {
                    pObj.fCompl1 ^= Abc_LitIsCompl(iLit);
                    Vec_IntWriteEntry( &pObj.vFanins, 1, Abc_Lit2Var(iLit) );
                }
            }
        }
    }
    }
    Vec_IntFree( vMap2Repl );

//    pNtk = Abc_NtkRestrash( p->pNtk, 1 );
    if fAnd != 0 {
        pNtk = Abc_NtkBalance( p.pNtk, 0, 0, 1 );
    } else {
        pNtk = Abc_NtkBalanceExor( p.pNtk, 1, 0 ); }
    return pNtk;
}

/**Function*************************************************************

  Synopsis    [Extracts one multi-output XOR.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkShareXor(pNtk *Abc_Ntk_t, nMultiSize int32, fAnd int32, fVerbose int32) *Abc_Ntk_t {
    var pNtkNew *Abc_Ntk_t;
    var p *Abc_ShaMan_t;
    assert(Abc_NtkIsStrash(pNtk), "Abc_NtkIsStrash(pNtk)");
//    Abc_NtkDumpBlif( pNtk );
    p = Abc_ShaManStart( pNtk );
    p.nMultiSize = nMultiSize;
    p.fVerbose   = fVerbose;
    Abc_NtkTraverseSupers( p, fAnd );
    if p.nStartCols < 2 {
        Abc_ShaManStop( p );
        return Abc_NtkDup( pNtk );
    }
    if fVerbose != 0 {
        Abc_NtkSharePrint( p ); }
    Abc_NtkShareOptimize( p, fAnd );
    if fVerbose != 0 {
        Abc_NtkSharePrint( p ); }
    pNtkNew = Abc_NtkUpdateNetwork( p, fAnd );
    Abc_ShaManStop( p );
    return pNtkNew;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

