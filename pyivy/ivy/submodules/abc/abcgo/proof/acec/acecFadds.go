package acec

import "fmt"

/**CFile****************************************************************

  FileName    [acecFadds.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [CEC for arithmetic circuits.]

  Synopsis    [Detecting half-adders and full-adders.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: acecFadds.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "acecInt.h"
//#include "misc/vec/vecWec.h"
//#include "misc/tim/tim.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// Original macro:
// #define Dtc_ForEachCut( pList, pCut, i ) for ( i = 0, pCut = pList + 1; i < pList[0]; i++, pCut += pCut[0] + 1 )

// Original macro:
// #define Dtc_ForEachFadd( vFadds, i )     for ( i = 0; i < Vec_IntSize(vFadds)/5; i++ )


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Detecting HADDs in the AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManIllustrateBoxes(p *Gia_Man_t) {
    var pManTime *Tim_Man_t  = (Tim_Man_t *)p.pManTime;
    var nBoxes int32  = Tim_ManBoxNum( pManTime );
    var i, k, curCi, curCo, nBoxIns, nBoxOuts int32;
    var pObj *Gia_Obj_t;
    // walk through the boxes
    curCi = Tim_ManPiNum(pManTime);
    curCo = 0;
    for  i = 0; i < nBoxes; i++  {
        nBoxIns = Tim_ManBoxInputNum(pManTime, i);
        nBoxOuts = Tim_ManBoxOutputNum(pManTime, i);
        fmt.Printf("Box %4d  [%d x %d] :   ", i, nBoxIns, nBoxOuts);
        fmt.Printf("Input obj IDs = ");
        for  k = 0; k < nBoxIns; k++  {
            pObj = Gia_ManCo( p, curCo + k );
            fmt.Printf("%d ", Gia_ObjId(p, pObj));
        }
        fmt.Printf("  Output obj IDs = ");
        for  k = 0; k < nBoxOuts; k++  {
            pObj = Gia_ManCi( p, curCi + k );
            fmt.Printf("%d ", Gia_ObjId(p, pObj));
        }
        curCo += nBoxIns;
        curCi += nBoxOuts;
        fmt.Printf("\n");
    }
    curCo += Tim_ManPoNum(pManTime);
    // verify counts
    assert(curCi == Gia_ManCiNum(p), "curCi == Gia_ManCiNum(p)");
    assert(curCo == Gia_ManCoNum(p), "curCo == Gia_ManCoNum(p)");
}

/**Function*************************************************************

  Synopsis    [Detecting FADDs in the AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupFadd(pNew *Gia_Man_t, p *Gia_Man_t, vChain *Vec_Int_t, vFadds *Vec_Int_t, vMap *Vec_Int_t, vChains *Vec_Wec_t, vMap2Chain *Vec_Int_t, vTruths *Vec_Int_t) {
    extern void Gia_ManDupWithFaddBoxes_rec( Gia_Man_t * pNew, Gia_Man_t * p, Gia_Obj_t * pObj, Vec_Int_t * vFadds, Vec_Int_t * vMap, Vec_Wec_t * vChains, Vec_Int_t * vMap2Chain, Vec_Int_t * vTruths );
    var i int32;
    var k int32;
    var iFadd int32 = -1;
    var iCiLit int32;
    var pLits [3]int32;
    var pObj *Gia_Obj_t;
    // construct FADD inputs
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vChain)) && (((iFadd) = Vec_IntEntry(vChain, i)), 1); i++  {
        for  k = 0; k < 3; k++  {
            if (i != 0) && !(k != 0) { continue; }
            pObj = Gia_ManObj( p, Vec_IntEntry(vFadds, 5*iFadd+k) );
            Gia_ManDupWithFaddBoxes_rec( pNew, p, pObj, vFadds, vMap, vChains, vMap2Chain, vTruths );
        }
    }
    // construct boxes
    iCiLit = 0;
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vChain)) && (((iFadd) = Vec_IntEntry(vChain, i)), 1); i++  {
        var iXorTruth int32  = Vec_IntEntry( vTruths, 2*iFadd+0 );
        var iMajTruth int32  = Vec_IntEntry( vTruths, 2*iFadd+1 );
        for  k = 0; k < 3; k++  {
            pObj = Gia_ManObj( p, Vec_IntEntry(vFadds, 5*iFadd+k) );
            pLits[k] = (!k && iCiLit) ? iCiLit : pObj.Value;
            assert(pLits[k] >= 0, "pLits[k] >= 0");
        }
        // normalize truth table
        //    if ( Truth == 0xE8 || Truth == 0xD4 || Truth == 0xB2 || Truth == 0x71 ||
        //         Truth == 0x17 || Truth == 0x2B || Truth == 0x4D || Truth == 0x8E )
        if iMajTruth == 0x4D {
            pLits[0] = Abc_LitNot(pLits[0]), iMajTruth = 0x8E, iXorTruth = 0xFF & ~iXorTruth;
        } else if iMajTruth == 0xD4 {
            pLits[0] = Abc_LitNot(pLits[0]), iMajTruth = 0xE8, iXorTruth = 0xFF & ~iXorTruth;
        } else if iMajTruth == 0x2B {
            pLits[1] = Abc_LitNot(pLits[1]), iMajTruth = 0x8E, iXorTruth = 0xFF & ~iXorTruth;
        } else if iMajTruth == 0xB2 {
            pLits[1] = Abc_LitNot(pLits[1]), iMajTruth = 0xE8, iXorTruth = 0xFF & ~iXorTruth; }
        if iMajTruth == 0x8E {
            pLits[2] = Abc_LitNot(pLits[2]), iMajTruth = 0xE8, iXorTruth = 0xFF & ~iXorTruth;
        } else if iMajTruth == 0x71 {
            pLits[2] = Abc_LitNot(pLits[2]), iMajTruth = 0x17, iXorTruth = 0xFF & ~iXorTruth;
        } else { assert(iMajTruth == 0xE8 || iMajTruth == 0x17, "iMajTruth == 0xE8 || iMajTruth == 0x17"); }
        // normalize carry-in
        if Abc_LitIsCompl(pLits[0]) {
            for  k = 0; k < 3; k++  {
                pLits[k] = Abc_LitNot(pLits[k]);
            }
            iXorTruth = 0xFF & ~iXorTruth;
            iMajTruth = 0xFF & ~iMajTruth;
        }
        // add COs
        assert(!Abc_LitIsCompl(pLits[0]), "!Abc_LitIsCompl(pLits[0])");
        for  k = 0; k < 3; k++  {
            Gia_ManAppendCo( pNew, pLits[k] );
        }
        // create CI
        assert(iXorTruth == 0x96 || iXorTruth == 0x69, "iXorTruth == 0x96 || iXorTruth == 0x69");
        pObj = Gia_ManObj( p, Vec_IntEntry(vFadds, 5*iFadd+3) );
        pObj.Value = Abc_LitNotCond( Gia_ManAppendCi(pNew), int32((iXorTruth == 0x69)) );
        // create CI
        assert(iMajTruth == 0xE8 || iMajTruth == 0x17, "iMajTruth == 0xE8 || iMajTruth == 0x17");
        iCiLit = Abc_LitNotCond( Gia_ManAppendCi(pNew), int32((iMajTruth == 0x17)) );
    }   
    // assign carry out
    assert(iFadd == Vec_IntEntryLast(vChain), "iFadd == Vec_IntEntryLast(vChain)");
    pObj = Gia_ManObj( p, Vec_IntEntry(vFadds, 5*iFadd+4) );
    pObj.Value = iCiLit;
}
func Gia_ManDupWithFaddBoxes_rec(pNew *Gia_Man_t, p *Gia_Man_t, pObj *Gia_Obj_t, vFadds *Vec_Int_t, vMap *Vec_Int_t, vChains *Vec_Wec_t, vMap2Chain *Vec_Int_t, vTruths *Vec_Int_t) {
    var iChain int32;
    if ~pObj.Value {
        return; }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    iChain = Vec_IntEntry( vMap2Chain, Gia_ObjId(p, pObj) );
/*
    assert( iChain == -1 );
    if ( iChain >= 0 )
    {
        Gia_ManDupFadd( pNew, p, Vec_WecEntry(vChains, iChain), vFadds, vMap, vChains, vMap2Chain, vTruths );
        assert( ~pObj->Value );
        return;
    }
*/
    Gia_ManDupWithFaddBoxes_rec( pNew, p, Gia_ObjFanin0(pObj), vFadds, vMap, vChains, vMap2Chain, vTruths );
    Gia_ManDupWithFaddBoxes_rec( pNew, p, Gia_ObjFanin1(pObj), vFadds, vMap, vChains, vMap2Chain, vTruths );
    pObj.Value = Gia_ManAppendAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
}
func Gia_ManDupWithNaturalBoxes(p *Gia_Man_t, nFaddMin int32, fVerbose int32) *Gia_Man_t {
    var clk abctime  = Abc_Clock();
    var pNew *Gia_Man_t;//, * pTemp;
    var vFadds, vMap, vMap2Chain, vTruths, vChain *Vec_Int_t;
    var vChains *Vec_Wec_t;
    var pObj *Gia_Obj_t;  
    var i, nBoxes int32;
    if Gia_ManBoxNum(p) > 0 {
        fmt.Printf("Currently natural carry-chains cannot be detected when boxes are present.\n");
        return nil;
    }
    assert(Gia_ManBoxNum(p) == 0, "Gia_ManBoxNum(p) == 0");

    // detect FADDs
    vFadds = Gia_ManDetectFullAdders( p, fVerbose, nil );
    assert(Vec_IntSize(vFadds) % 5 == 0, "Vec_IntSize(vFadds) % 5 == 0");
    // map MAJ into its FADD
    vMap = Gia_ManCreateMap( p, vFadds );
    // for each FADD, find the longest chain and reorder its inputs
    Gia_ManFindChains( p, vFadds, vMap );
    // returns the set of topmost MAJ nodes
    vChains = Gia_ManCollectTopmost( p, vFadds, vMap, nFaddMin );
    if fVerbose != 0 {
        Gia_ManPrintChains( p, vFadds, vMap, vChains ); }
    if Vec_WecSize(vChains) == 0 {
        Vec_IntFree( vFadds );
        Vec_IntFree( vMap );
        Vec_WecFree( vChains );
        return Gia_ManDup( p );
    }
    // returns mapping of each MAJ into the topmost elements of its chain
    vMap2Chain = Gia_ManFindMapping( p, vFadds, vMap, vChains );
    // compute truth tables for FADDs
    vTruths = Gia_ManCollectTruthTables( p, vFadds );
    if fVerbose != 0 {
        Abc_PrintTime( 1, "Carry-chain detection time", Abc_Clock() - clk ); }

    // duplicate
    clk = Abc_Clock();
    Gia_ManFillValue( p );
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObj.Value = Gia_ManAppendCi( pNew );
    }
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(vChains)) && (((vChain) = Vec_WecEntry(vChains, i)), 1); i++  {
        Gia_ManDupFadd( pNew, p, vChain, vFadds, vMap, vChains, vMap2Chain, vTruths );
    }
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        Gia_ManDupWithFaddBoxes_rec( pNew, p, Gia_ObjFanin0(pObj), vFadds, vMap, vChains, vMap2Chain, vTruths );
    }
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
    }
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(p) );
    if Gia_ManRegNum(p) {
        if fVerbose != 0 {
            fmt.Printf("Warning: Sequential design is coverted into combinational one by adding white boxes.\n"); }
        pNew.nRegs = 0;
    }
    assert(!Gia_ManHasDangling(pNew), "!Gia_ManHasDangling(pNew)");

    // cleanup
    Vec_IntFree( vFadds );
    Vec_IntFree( vMap );
    Vec_WecFree( vChains );
    Vec_IntFree( vMap2Chain );
    Vec_IntFree( vTruths );
    
    // other information
    nBoxes = (Gia_ManCiNum(pNew) - Gia_ManCiNum(p)) / 2;
    assert(nBoxes == (Gia_ManCoNum(pNew) - Gia_ManCoNum(p)) / 3, "nBoxes == (Gia_ManCoNum(pNew) - Gia_ManCoNum(p)) / 3");
    pNew.pManTime  = Gia_ManGenerateTim( Gia_ManCiNum(p), Gia_ManCoNum(p), nBoxes, 3, 2 );
    pNew.pAigExtra = Gia_ManGenerateExtraAig( nBoxes, 3, 2 );
/*
    // normalize
    pNew = Gia_ManDupNormalize( pTemp = pNew, 0 );
    pNew->pManTime  = pTemp->pManTime;  pTemp->pManTime  = NULL;
    pNew->pAigExtra = pTemp->pAigExtra; pTemp->pAigExtra = NULL;
    Gia_ManStop( pTemp );
*/
    //pNew = Gia_ManDupCollapse( pTemp = pNew, pNew->pAigExtra, NULL );
    //Gia_ManStop( pTemp );

    //Gia_ManIllustrateBoxes( pNew );
    if fVerbose != 0 {
        Abc_PrintTime( 1, "AIG with boxes construction time", Abc_Clock() - clk ); }
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Converting AIG with annotated carry-chains into AIG with boxes.]

  Description [Assumes that annotations are pObj->fMark0 or pObj->fMark1.
  Only one of these can be set to 1.  If fMark0 (fMark1) is set to 1, 
  the first (second) input of an AND-gate is chained.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ObjFanin0CopyCarry(vCarries *Vec_Int_t, pObj *Gia_Obj_t, Id int32) int32 {
    if vCarries == nil || Vec_IntEntry(vCarries, Gia_ObjFaninId0(pObj, Id)) == -1 {
        return Gia_ObjFanin0Copy(pObj); }
    return Abc_LitNotCond( Vec_IntEntry(vCarries, Gia_ObjFaninId0(pObj, Id)), Gia_ObjFaninC0(pObj) );
}
func Gia_ObjFanin1CopyCarry(vCarries *Vec_Int_t, pObj *Gia_Obj_t, Id int32) int32 {
    if vCarries == nil || Vec_IntEntry(vCarries, Gia_ObjFaninId1(pObj, Id)) == -1 {
        return Gia_ObjFanin1Copy(pObj); }
    return Abc_LitNotCond( Vec_IntEntry(vCarries, Gia_ObjFaninId1(pObj, Id)), Gia_ObjFaninC1(pObj) );
}
func Gia_ManDupWithArtificalFaddBoxes(p *Gia_Man_t, fUseFanout int32, fXorTrick int32) *Gia_Man_t {
    var pNew *Gia_Man_t;
    var pObj *Gia_Obj_t;  
    var nBoxes int32  = Gia_ManBoxNum(p);
    var i, nRealPis, nRealPos int32;
    var vCarries *Vec_Int_t  = nil;
    // make sure two chains do not overlap
    Gia_ManCleanPhase( p );
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        assert(!pObj.fMark0 && !pObj.fMark1, "!pObj.fMark0 && !pObj.fMark1");
    }
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        assert(!pObj.fMark0 && !pObj.fMark1, "!pObj.fMark0 && !pObj.fMark1");
    }
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        assert(!pObj.fMark0 || !pObj.fMark1, "!pObj.fMark0 || !pObj.fMark1");
        if pObj.fMark0 {
            assert(Gia_ObjFanin0(pObj).fPhase == 0, "Gia_ObjFanin0(pObj).fPhase == 0");
            Gia_ObjFanin0(pObj).fPhase = 1;
        }
        if pObj.fMark1 {
            assert(Gia_ObjFanin1(pObj).fPhase == 0, "Gia_ObjFanin1(pObj).fPhase == 0");
            Gia_ObjFanin1(pObj).fPhase = 1;
        }
    }
    }
    // create mapping for carry-chains
    if !(fUseFanout != 0) {
        vCarries = Vec_IntStartFull( Gia_ManObjNum(p) ); }
    // create references and discount carries
    if vCarries != nil {
        Gia_ManCreateRefs( p );
        // inlining of macro Gia_ManForEachAnd
        for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else
            if pObj.fMark0 {
                Gia_ObjRefFanin0Dec( p, pObj );
            } else if pObj.fMark1 {
                Gia_ObjRefFanin1Dec( p, pObj ); }
        }
    }
    // if AIG already has (natural) FADD boxes, it should not un-normalized
    Gia_ManFillValue( p );
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachObj1
    for  i = 1; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        if Gia_ObjIsCi(pObj) {
            pObj.Value = Gia_ManAppendCi( pNew );
        } else if Gia_ObjIsCo(pObj) {
            pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
        } else if !pObj.fMark0 && !pObj.fMark1 { // AND-gate
            pObj.Value = Gia_ManAppendAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
        } else { // AND-gate with chain
        
            var iCiLit, iOtherLit, iLit0, iLit1, iLit2, iXorLit int32;
            assert(pObj.fMark0 != pObj.fMark1, "pObj.fMark0 != pObj.fMark1");
            if pObj.fMark0 {
                iCiLit = Gia_ObjFanin0CopyCarry(vCarries, pObj, i);
            } else {
                iCiLit = Gia_ObjFanin1CopyCarry(vCarries, pObj, i);
            }
            if pObj.fMark0 {
                iOtherLit = Gia_ObjFanin1Copy(pObj);
            } else {
                iOtherLit = Gia_ObjFanin0Copy(pObj);
            }
            assert(iCiLit >= 0 && iOtherLit >= 0, "iCiLit >= 0 && iOtherLit >= 0");
            iLit0 = Abc_LitNotCond( iCiLit,    Abc_LitIsCompl(iCiLit) );
            iLit1 = Abc_LitNotCond( iOtherLit, Abc_LitIsCompl(iCiLit) );
            iLit2 = Abc_LitNotCond( 0,         Abc_LitIsCompl(iCiLit) );
            // add COs
            assert(!Abc_LitIsCompl(iLit0), "!Abc_LitIsCompl(iLit0)");
            Gia_ManAppendCo( pNew, iLit0 );
            Gia_ManAppendCo( pNew, iLit1 );
            Gia_ManAppendCo( pNew, iLit2 );
            // add CI (unused sum bit)
            iXorLit = Gia_ManAppendCi(pNew);
            // add CI (carry bit)
            pObj.Value = Abc_LitNotCond( Gia_ManAppendCi(pNew), Abc_LitIsCompl(iCiLit) );
            if (vCarries != nil) && pObj.fPhase {
                Vec_IntWriteEntry( vCarries, i, pObj.Value );
                if Gia_ObjRefNum(p, pObj) > 0 {
                    if fXorTrick != 0 {
                        pObj.Value = Gia_ManAppendAnd( pNew, Abc_LitNotCond(iXorLit, !Abc_LitIsCompl(iCiLit)), iOtherLit );
                    } else {
                        pObj.Value = Gia_ManAppendAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) ); }
                }
            }
            nBoxes++;
        }
    }
    Gia_ManCleanPhase( p );
    Vec_IntFreeP( &vCarries );
    p.pRefs = nil;
    assert(!Gia_ManHasDangling(pNew), "!Gia_ManHasDangling(pNew)");
    // other information
//    nBoxes += (Gia_ManCiNum(pNew) - Gia_ManCiNum(p)) / 2;
//    assert( nBoxes == Gia_ManBoxNum(p) + (Gia_ManCoNum(pNew) - Gia_ManCoNum(p)) / 3 );
    if Gia_ManBoxNum(p) {
        nRealPis = Tim_ManPiNum((Tim_Man_t *)p.pManTime);
    } else {
        nRealPis = Gia_ManCiNum(p);
    }
    if Gia_ManBoxNum(p) {
        nRealPos = Tim_ManPoNum((Tim_Man_t *)p.pManTime);
    } else {
        nRealPos = Gia_ManCoNum(p);
    }
    pNew.pManTime  = Gia_ManGenerateTim( nRealPis, nRealPos, nBoxes, 3, 2 );
    pNew.pAigExtra = Gia_ManGenerateExtraAig( nBoxes, 3, 2 );
    // optionally normalize the AIG
    return pNew;
}
func Gia_ManDupWithArtificalFaddBoxesTest(p *Gia_Man_t) *Gia_Man_t {
    var pNew *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i int32;
    // label some and-gates
    Gia_ManCleanMark01( p );
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        pObj.fMark0 = i % 5;
        pObj.fMark1 = i % 7;
        if pObj.fMark0 && pObj.fMark1 {
            pObj.fMark0 = pObj.fMark1 = 0; }
    }
    }

    // output new AIG
    pNew = Gia_ManDupWithArtificalFaddBoxes( p, 0, 0 );
    Gia_ManCleanMark01( p );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Adds artificial carry chains to the manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
// computes AIG delay information when boxes are used
func Gia_ManFindAnnotatedDelay(p *Gia_Man_t, DelayC int32, pnBoxes *int32, fIgnoreBoxDelays int32) int32 {
    var pObj *Gia_Obj_t;
    var nRealPis int32;
    if Gia_ManBoxNum(p) {
        nRealPis = Tim_ManPiNum((Tim_Man_t *)p.pManTime);
    } else {
        nRealPis = Gia_ManCiNum(p);
    }
    var pDelays *int32  = Vec_IntArray(p.vLevels);
    var i, k, iBox, iBoxOutId, Delay, Delay0, Delay1, DelayMax, nBoxes int32;
    Vec_IntFill( p.vLevels, Gia_ManObjNum(p), 0 );
    // inlining of macro Gia_ManForEachObj1
    for  i = 1; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        if Gia_ObjIsCi(pObj) {
            if fIgnoreBoxDelays != 0 {
                continue; }
            // check if it is real PI
            iBoxOutId = Gia_ObjCioId(pObj) - nRealPis;
            if iBoxOutId < 0 {
                continue; }
            // if it is a box output, find box number
            iBox = iBoxOutId / 2;
            assert(iBox < Gia_ManBoxNum(p), "iBox < Gia_ManBoxNum(p)");
            // check find the maximum delay of the box inputs
            Delay = 0;
            for  k = 0; k < 3; k++  {
                var Id int32  = Gia_ObjId( p, Gia_ManCo(p, iBox*3+k) );
                assert(Id < i, "Id < i");
                Delay = Abc_MaxInt( Delay, pDelays[Id] );
            }
            // consider outputs
            if iBoxOutId & 1 { // carry output
                Delay += DelayC;
            } else { // sum output
                Delay += 100; }
            pDelays[i] = Delay;
            continue;
        }
        if Gia_ObjIsCo(pObj) {
            pDelays[i] = pDelays[Gia_ObjFaninId0(pObj, i)];
            DelayMax = Abc_MaxInt( DelayMax, pDelays[i] );
            continue;
        }
        assert(!pObj.fMark0 || !pObj.fMark1, "!pObj.fMark0 || !pObj.fMark1");
        Delay0 = pDelays[Gia_ObjFaninId0(pObj, i)];
        Delay1 = pDelays[Gia_ObjFaninId1(pObj, i)];
        if pObj.fMark0 {
            Delay = Abc_MaxInt( Delay0 + DelayC, Delay1 + 100 );
            nBoxes++;
        
        } else if pObj.fMark1 {
            Delay = Abc_MaxInt( Delay1 + DelayC, Delay0 + 100 );
            nBoxes++;
        
        } else {
            Delay = Abc_MaxInt( Delay0 + 100, Delay1 + 100 ); }
        pDelays[i] = Delay;
    }
    if pnBoxes != nil {
        *pnBoxes = nBoxes; }
    return DelayMax;
}
// check if the object is already used in some chain
func Gia_ObjIsUsed(pObj *Gia_Obj_t) int32 {
    return pObj.fMark0 || pObj.fMark1 || pObj.fPhase;
}
// finds internal node that can begin a new chain
func Gia_ManFindChainStart(p *Gia_Man_t) int32 {
    var pObj *Gia_Obj_t;
    var pDelays *int32  = Vec_IntArray(p.vLevels);
    var i int32;
    var iMax int32 = -1;
    var DelayMax int32;
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        if Gia_ObjIsUsed(pObj) != 0 {
            continue; }
        if DelayMax > pDelays[i] {
            continue; }
        DelayMax = pDelays[i];
        iMax = i;
    }
    }
    return iMax;
}
// finds a sequence of internal nodes that creates a new chain
func Gia_ManFindPath(p *Gia_Man_t, DelayC int32, nPathMin int32, nPathMax int32, vPath *Vec_Int_t) int32 {
    var pObj, pFanin0, pFanin1 *Gia_Obj_t;
    var pDelays *int32  = Vec_IntArray(p.vLevels);
    var i int32;
    var iLit int32;
    var iMax int32 = Gia_ManFindChainStart( p );
    if iMax == -1 {
        return -1; }
    Vec_IntClear( vPath );
    pObj = Gia_ManObj(p, iMax);
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    for  Gia_ObjIsAnd(pObj)  {
        assert(!(Gia_ObjIsUsed(pObj) != 0), "!Gia_ObjIsUsed(pObj)");
        pFanin0 = Gia_ObjFanin0(pObj);
        pFanin1 = Gia_ObjFanin1(pObj);
        if (Gia_ObjIsUsed(pFanin0) != 0) && (Gia_ObjIsUsed(pFanin1) != 0) {
            break; }
        if Gia_ObjIsUsed(pFanin0) != 0 {
            Vec_IntPush( vPath, Abc_Var2Lit(Gia_ObjId(p, pObj), 1) );
            pObj = pFanin1;
        
        } else if Gia_ObjIsUsed(pFanin1) != 0 {
            Vec_IntPush( vPath, Abc_Var2Lit(Gia_ObjId(p, pObj), 0) );
            pObj = pFanin0;
        
        } else {
            if pDelays[Gia_ObjId(p, pFanin1)] > pDelays[Gia_ObjId(p, pFanin0)] {
                Vec_IntPush( vPath, Abc_Var2Lit(Gia_ObjId(p, pObj), 1) );
                pObj = pFanin1;
            
            } else {
                Vec_IntPush( vPath, Abc_Var2Lit(Gia_ObjId(p, pObj), 0) );
                pObj = pFanin0;
            }
        }
    }
    if Vec_IntSize(vPath) < nPathMin {
        Gia_ManObj(p, iMax).fPhase = 1;
        return 0;
    }
    // label nodes
    if Vec_IntSize(vPath) > nPathMax {
        Vec_IntShrink( vPath, nPathMax ); }
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vPath)) && (((iLit) = Vec_IntEntry(vPath, i)), 1); i++  {
        pObj = Gia_ManObj( p, Abc_Lit2Var(iLit) );
        if Abc_LitIsCompl(iLit) {
            assert(pObj.fMark1 == 0, "pObj.fMark1 == 0");
            pObj.fMark1 = 1;
            assert(Gia_ObjFanin1(pObj).fPhase == 0, "Gia_ObjFanin1(pObj).fPhase == 0");
            Gia_ObjFanin1(pObj).fPhase = 1;
        
        } else {
            assert(pObj.fMark0 == 0, "pObj.fMark0 == 0");
            pObj.fMark0 = 1;
            assert(Gia_ObjFanin0(pObj).fPhase == 0, "Gia_ObjFanin0(pObj).fPhase == 0");
            Gia_ObjFanin0(pObj).fPhase = 1;
        }
    }
    return Vec_IntSize(vPath);
}
// iteratively create the given number of chains
func Gia_ManIteratePaths(p *Gia_Man_t, DelayC int32, nPathMin int32, nPathMax int32, nPathLimit int32, fIgnoreBoxDelays int32, fVerbose int32) int32 {
    var pObj *Gia_Obj_t;
    var vPath *Vec_Int_t  = Vec_IntAlloc( 100 );
    var i, RetValue, nBoxes, MaxDelay, nPaths int32;
    assert(p.vLevels == nil, "p.vLevels == nil");
    p.vLevels = Vec_IntStart( Gia_ManObjNum(p) );
    Gia_ManCleanMark01( p );
    Gia_ManCleanPhase( p );
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObj.fPhase = 1;
    }
    if fVerbose != 0 {
        fmt.Printf("Running path detection: BoxDelay = %d, PathMin = %d, PathMax = %d, PathLimit = %d.\n", DelayC, nPathMin, nPathMax, nPathLimit); }
    for  i = 0; i < nPathLimit; i++  {
        MaxDelay = Gia_ManFindAnnotatedDelay( p, DelayC, &nBoxes, fIgnoreBoxDelays );
        RetValue = Gia_ManFindPath( p, DelayC, nPathMin, nPathMax, vPath );
        if RetValue == -1 {
            break; }
        nPaths += (RetValue > 0);
        if fVerbose != 0 {
            fmt.Printf("Iter %5d : Paths = %2d. Boxes = %2d. Total boxes = %6d.  Max delay = %5d.\n", i, nPaths, RetValue, nBoxes, MaxDelay); }
    }
    Vec_IntFree( vPath );
    Vec_IntFreeP( &p.vLevels );
    Gia_ManCleanPhase( p );
    return 1;
}
// annotate artificial chains and then put them into boxes
func Gia_ManDupWithArtificialBoxes(p *Gia_Man_t, DelayC int32, nPathMin int32, nPathMax int32, nPathLimit int32, fUseFanout int32, fXorTrick int32, fIgnoreBoxDelays int32, fVerbose int32) *Gia_Man_t {
    var pNew *Gia_Man_t;
/*
    if ( Gia_ManBoxNum(p) > 0 )
    {
        printf( "Currently artifical carry-chains cannot be detected when natural ones are present.\n" );
        return NULL;
    }
*/
    Gia_ManIteratePaths( p, DelayC, nPathMin, nPathMax, nPathLimit, fIgnoreBoxDelays, fVerbose );
    pNew = Gia_ManDupWithArtificalFaddBoxes( p, fUseFanout, fXorTrick );
    Gia_ManCleanMark01( p );
    return pNew;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

