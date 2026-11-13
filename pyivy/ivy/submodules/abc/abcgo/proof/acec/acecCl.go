package acec

import "fmt"

/**CFile****************************************************************

  FileName    [acecCl.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [CEC for arithmetic circuits.]

  Synopsis    [Core procedures.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: acecCl.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "acecInt.h"
//#include "misc/vec/vecWec.h"
//#include "misc/extra/extra.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////



////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Acec_CollectXorTops(p *Gia_Man_t) *Vec_Int_t {
    var vRootXorSet *Vec_Int_t  = Vec_IntAlloc( Gia_ManCoNum(p) );
    var pObj, pFan0, pFan1, pFan00, pFan01, pFan10, pFan11 *Gia_Obj_t;
    var i, fXor0, fXor1, fFirstXor int32;
    // inlining of macro Gia_ManForEachCoDriver
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ObjFanin0(Gia_ManCo(p, i))); i++  {
        if !Gia_ObjRecognizeExor(pObj, &pFan0, &pFan1) {
            if fFirstXor != 0 {
                fmt.Printf("XORs do not form a continuous sequence\n");
                Vec_IntFreeP( &vRootXorSet );
                break;
            }
            continue;
        }
        fFirstXor = 1;
        fXor0 = Gia_ObjRecognizeExor(Gia_Regular(pFan0), &pFan00, &pFan01);
        fXor1 = Gia_ObjRecognizeExor(Gia_Regular(pFan1), &pFan10, &pFan11);
        if fXor0 == fXor1 {
            fmt.Printf("Both inputs of top level XOR have XOR/non-XOR\n");
            Vec_IntFreeP( &vRootXorSet );
            break;
        }
        Vec_IntPush( vRootXorSet, Gia_ObjId(p, pObj) );
        Vec_IntPush( vRootXorSet, fXor1 ? Gia_ObjId(p, Gia_Regular(pFan0))  : Gia_ObjId(p, Gia_Regular(pFan1)) );
        Vec_IntPush( vRootXorSet, fXor1 ? Gia_ObjId(p, Gia_Regular(pFan10)) : Gia_ObjId(p, Gia_Regular(pFan00)) );
        Vec_IntPush( vRootXorSet, fXor1 ? Gia_ObjId(p, Gia_Regular(pFan11)) : Gia_ObjId(p, Gia_Regular(pFan01)) );
    }
    for  i = 0; 4*i < Vec_IntSize(vRootXorSet); i++  {
        fmt.Printf("%2d : ", i);
        fmt.Printf("%4d <- ", Vec_IntEntry(vRootXorSet, 4*i));
        fmt.Printf("%4d ", Vec_IntEntry(vRootXorSet, 4*i+1));
        fmt.Printf("%4d ", Vec_IntEntry(vRootXorSet, 4*i+2));
        fmt.Printf("%4d ", Vec_IntEntry(vRootXorSet, 4*i+3));
        fmt.Printf("\n");
    }
    return vRootXorSet;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Acec_DetectComputeSuppOne_rec(p *Gia_Man_t, pObj *Gia_Obj_t, vSupp *Vec_Int_t, vNods *Vec_Int_t) {
    if Gia_ObjIsTravIdCurrent(p, pObj) {
        return; }
    Gia_ObjSetTravIdCurrent(p, pObj);
    if pObj.fMark0 {
        Vec_IntPush( vSupp, Gia_ObjId(p, pObj) );
        return;
    }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    Acec_DetectComputeSuppOne_rec( p, Gia_ObjFanin0(pObj), vSupp, vNods );
    Acec_DetectComputeSuppOne_rec( p, Gia_ObjFanin1(pObj), vSupp, vNods );
    Vec_IntPush( vNods, Gia_ObjId(p, pObj) );
}
func Acec_DetectComputeSupports(p *Gia_Man_t, vRootXorSet *Vec_Int_t) {
    var vNods *Vec_Int_t  = Vec_IntAlloc( 100 ); 
    var vPols *Vec_Int_t  = Vec_IntAlloc( 100 ); 
    var vSupp *Vec_Int_t  = Vec_IntAlloc( 100 );  var i, k, Node, Pol int32;
    for  i = 0; 4*i < Vec_IntSize(vRootXorSet); i++  {
        Gia_ManObj( p, Vec_IntEntry(vRootXorSet, 4*i+1) ).fMark0 = 1;
        Gia_ManObj( p, Vec_IntEntry(vRootXorSet, 4*i+2) ).fMark0 = 1;
        Gia_ManObj( p, Vec_IntEntry(vRootXorSet, 4*i+3) ).fMark0 = 1;
    }
    for  i = 1; 4*i < Vec_IntSize(vRootXorSet); i++  {
        Vec_IntClear( vSupp );
        Gia_ManIncrementTravId( p );

        Gia_ManObj( p, Vec_IntEntry(vRootXorSet, 4*i+1) ).fMark0 = 0;
        Acec_DetectComputeSuppOne_rec( p, Gia_ManObj( p, Vec_IntEntry(vRootXorSet, 4*i+1) ), vSupp, vNods );
        Gia_ManObj( p, Vec_IntEntry(vRootXorSet, 4*i+1) ).fMark0 = 1;

        Vec_IntSort( vSupp, 0 );

        fmt.Printf("Out %4d : %4d  \n", i, Vec_IntEntry(vRootXorSet, 4*i+1));
        Vec_IntPrint( vSupp );

        fmt.Printf("Cone:\n");
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vNods)) && (((Node) = Vec_IntEntry(vNods, k)), 1); k++  {
            Gia_ObjPrint( p, Gia_ManObj(p, Node) );
        }


        Vec_IntClear( vPols );
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vSupp)) && (((Node) = Vec_IntEntry(vSupp, k)), 1); k++  {
            Vec_IntPush( vPols, Acec_DetectLitPolarity(p, Vec_IntEntry(vRootXorSet, 4*i+1), Node) );
        }

        // inlining of macro Vec_IntForEachEntryTwo
        for  k = 0; (k < Vec_IntSize(vSupp)) && (((Node) = Vec_IntEntry(vSupp, k)), 1) && (((Pol) = Vec_IntEntry(vPols, k)), 1); k++  {
            fmt.Printf("%d(%d)  ", Node, Abc_LitIsCompl(Pol));
        }

        fmt.Printf("\n");

        Vec_IntPrint( vSupp );
    }
    for  i = 0; 4*i < Vec_IntSize(vRootXorSet); i++  {
        Gia_ManObj( p, Vec_IntEntry(vRootXorSet, 4*i+1) ).fMark0 = 0;
        Gia_ManObj( p, Vec_IntEntry(vRootXorSet, 4*i+2) ).fMark0 = 0;
        Gia_ManObj( p, Vec_IntEntry(vRootXorSet, 4*i+3) ).fMark0 = 0;
    }
    Vec_IntFree( vSupp );
    Vec_IntFree( vPols );
    Vec_IntFree( vNods );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Acec_DetectXorBuildNew(p *Gia_Man_t, vRootXorSet *Vec_Int_t) *Gia_Man_t {
    var pNew *Gia_Man_t;
    var i int32;
    var k int32;
    var iOr1 int32;
    var iAnd1 int32;
    var iAnd2 int32;
    var pLits [3]int32; // carry, in1, in2
    var vMirrors *Vec_Int_t  = Vec_IntStart( Gia_ManObjNum(p) );
    for  i = 0; 4*i < Vec_IntSize(vRootXorSet); i++  {
        pLits[0] = Acec_DetectLitPolarity( p, Vec_IntEntry(vRootXorSet, 4*i), Vec_IntEntry(vRootXorSet, 4*i+1) );
        // get polarity of two new ones
        for  k = 1; k < 3; k++  {
            pLits[k] = Acec_DetectLitPolarity( p, Vec_IntEntry(vRootXorSet, 4*i), Vec_IntEntry(vRootXorSet, 4*i+k+1) );
        }
        // create the gate
        iOr1  = Gia_ManAppendOr( p, pLits[1], pLits[2] );
        iAnd1 = Gia_ManAppendAnd( p, pLits[0], iOr1 );
        iAnd2 = Gia_ManAppendAnd( p, pLits[1], pLits[2] );
        pLits[0] = Gia_ManAppendOr( p, iAnd1, iAnd2 );
        Vec_IntWriteEntry( vMirrors, Vec_IntEntry(vRootXorSet, 4*i+1), pLits[0] );
    }
    // remap the AIG using map
    pNew = Acec_ManDerive( p, vMirrors );
    Vec_IntFree( vMirrors );
    return pNew;
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Acec_DetectAdditional(p *Gia_Man_t, fVerbose int32) *Gia_Man_t {
    var clk abctime  = Abc_Clock();
    var pNew *Gia_Man_t;
    var vRootXorSet *Vec_Int_t;
//    Vec_Int_t * vXors, * vAdds = Ree_ManComputeCuts( p, &vXors, 0 );

    //Ree_ManPrintAdders( vAdds, 1 );
//    printf( "Detected %d full-adders and %d half-adders.  Found %d XOR-cuts.  ", Ree_ManCountFadds(vAdds), Vec_IntSize(vAdds)/6-Ree_ManCountFadds(vAdds), Vec_IntSize(vXors)/4 );
//    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );

    clk = Abc_Clock();
    vRootXorSet = Acec_CollectXorTops( p );
    if vRootXorSet != nil {
        Acec_DetectComputeSupports( p, vRootXorSet );

        pNew = Acec_DetectXorBuildNew( p, vRootXorSet );
        Vec_IntFree( vRootXorSet );
    
    } else {
        pNew = Gia_ManDup( p ); }

    fmt.Printf("Detected %d top XORs.  ", Vec_IntSize(vRootXorSet)/4);
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );

//    Vec_IntFree( vXors );
//    Vec_IntFree( vAdds );
    return pNew;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Acec_RewriteTop(p *Gia_Man_t, pBox *Acec_Box_t) *Vec_Int_t {
    var vRes *Vec_Int_t  = Vec_IntAlloc( Gia_ManCoNum(p) + 1 );
    var vLevel *Vec_Int_t;
    var i, k, iStart, iLit, Driver, Count int32;
    // determine how much to shift
    Driver = Gia_ObjFaninId0p( p, Gia_ManCo(p, 0) );
    // inlining of macro Vec_WecForEachLevel
    for  iStart = 0; (iStart < Vec_WecSize(pBox.vRootLits)) && (((vLevel) = Vec_WecEntry(pBox.vRootLits, iStart)), 1); iStart++  {
        if Abc_Lit2Var(Vec_IntEntry(vLevel,0)) == Driver {
            break; }
    }
    assert(iStart < Gia_ManCoNum(p), "iStart < Gia_ManCoNum(p)");
    //Vec_WecPrintLits( pBox->vRootLits );
    // inlining of macro Vec_WecForEachLevelStart
    for  i = iStart; (i < Vec_WecSize(pBox.vRootLits)) && (((vLevel) = Vec_WecEntry(pBox.vRootLits, i)), 1); i++  {
        var In [3]int32 ;
        var Out [2]int32;
        assert(Vec_IntSize(vLevel) > 0, "Vec_IntSize(vLevel) > 0");
        assert(Vec_IntSize(vLevel) <= 3, "Vec_IntSize(vLevel) <= 3");
        if Vec_IntSize(vLevel) == 1 {
            Vec_IntPush( vRes, Vec_IntEntry(vLevel, 0) );
            continue;
        }
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vLevel)) && (((iLit) = Vec_IntEntry(vLevel, k)), 1); k++  {
            In[k] = iLit;
        }
        Acec_InsertFadd( p, In, Out );
        Vec_IntPush( vRes, Out[0] );
        if i+1 < Vec_WecSize(pBox.vRootLits) {
            Vec_IntPush( Vec_WecEntry(pBox.vRootLits, i+1), Out[1] );
        } else {
            Vec_IntPush( Vec_WecPushLevel(pBox.vRootLits), Out[1] ); }
        Count++;
    }
    assert(Vec_IntSize(vRes) >= Gia_ManCoNum(p), "Vec_IntSize(vRes) >= Gia_ManCoNum(p)");
    Vec_IntShrink( vRes, Gia_ManCoNum(p) );
    fmt.Printf("Added %d adders for replace CLAs.  ", Count);
    return vRes;
}
func Acec_RewriteReplace(p *Gia_Man_t, vRes *Vec_Int_t) *Gia_Man_t {
    var pNew, pTemp *Gia_Man_t;
    var pObj *Gia_Obj_t;  var i int32;
    assert(Gia_ManCoNum(p) == Vec_IntSize(vRes), "Gia_ManCoNum(p) == Vec_IntSize(vRes)");
    // create new manager
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManFillValue( p );
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObj.Value = Gia_ManAppendCi(pNew);
    }
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        pObj.Value = Gia_ManAppendAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) ); }
    }
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        var iLit int32  = Vec_IntEntry( vRes, i );
        var pRepr *Gia_Obj_t  = Gia_ManObj( p, Abc_Lit2Var(iLit) );
        pObj.Value = Gia_ManAppendCo( pNew, pRepr.Value );
    }
    // set correct phase
    Gia_ManSetPhase( p );
    Gia_ManSetPhase( pNew );
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(pNew.vCos)) && ((pObj) = Gia_ManCo(pNew, i)); i++  {
        if Gia_ObjPhase(pObj) != Gia_ObjPhase(Gia_ManCo(p, i)) {
            Gia_ObjFlipFaninC0( pObj ); }
    }
    // remove dangling nodes
    pNew = Gia_ManCleanup( pTemp = pNew );
    Gia_ManStop( pTemp );
    return pNew;
}
func Acec_ManDecla(pGia *Gia_Man_t, fBooth int32, fVerbose int32) *Gia_Man_t {
    var clk abctime  = Abc_Clock();
    var pNew *Gia_Man_t  = nil;
    var vIgnore *Vec_Bit_t;
    if fBooth != 0 {
        vIgnore = Acec_BoothFindPPG(pGia);
    } else {
        vIgnore = nil;
    }
    var pBox *Acec_Box_t  = Acec_DeriveBox( pGia, vIgnore, 0, 0, fVerbose );
    var vResult *Vec_Int_t;
    Vec_BitFreeP( &vIgnore );
    if pBox == nil { // cannot match
    
        fmt.Printf("Cannot find arithmetic boxes.\n");
        return Gia_ManDup( pGia );
    }
    vResult = Acec_RewriteTop( pGia, pBox );
    Acec_BoxFreeP( &pBox );
    pNew = Acec_RewriteReplace( pGia, vResult );
    Vec_IntFree( vResult );
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
    return pNew;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

