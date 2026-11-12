package gia

import (
	"fmt"
	"strconv"
)

import "os"

/**CFile****************************************************************

  FileName    [giaMini.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [Reader/writer for MiniAIG.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaMini.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"
//#include "opt/dau/dau.h"
//#include "base/main/mainInt.h"
//#include "misc/util/utilTruth.h"
//#include "aig/miniaig/miniaig.h"
//#include "aig/miniaig/minilut.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Converts MiniAIG into GIA.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/



/**Function*************************************************************

  Synopsis    [Procedures to read/write GIA to/from MiniAIG file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManReadMiniAig(pFileName string) *Gia_Man_t {
    var p *Mini_Aig_t  = Mini_AigLoad( pFileName );
    var pGia *Gia_Man_t  = Gia_ManFromMiniAig( p, nil );
    pGia.pName = "";
    pGia.pName = Extra_FileNameGeneric( pFileName ); 
    Mini_AigStop( p );
    return pGia;
}
func Gia_ManWriteMiniAig(pGia *Gia_Man_t, pFileName string) {
    var p *Mini_Aig_t  = Gia_ManToMiniAig( pGia );
    Mini_AigDump( p, pFileName );
    //Mini_AigDumpVerilog( "test_miniaig.v", "top", p );
    Mini_AigStop( p );
}




/**Function*************************************************************

  Synopsis    [Converts MiniLUT into GIA.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Marks LUTs that should be complemented.]

  Description [These are LUTs whose all PO fanouts require them
  in negative polarity.  Other fanouts may require them in 
  positive polarity.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/




/**Function*************************************************************

  Synopsis    [Procedures to read/write GIA to/from MiniAIG file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManReadMiniLut(pFileName string) *Gia_Man_t {
    var p *Mini_Lut_t  = Mini_LutLoad( pFileName );
    var pGia *Gia_Man_t  = Gia_ManFromMiniLut( p, nil );
    pGia.pName = "";
    pGia.pName = Extra_FileNameGeneric( pFileName ); 
    Mini_LutStop( p );
    return pGia;
}
func Gia_ManWriteMiniLut(pGia *Gia_Man_t, pFileName string) {
    var p *Mini_Lut_t  = Gia_ManToMiniLut( pGia );
    Mini_LutDump( p, pFileName );
    Mini_LutStop( p );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManMapMiniLut2MiniAig(p *Gia_Man_t, p1 *Gia_Man_t, p2 *Gia_Man_t, vMap1 *Vec_Int_t, vMap2 *Vec_Int_t) *int32 {
    var pRes *int32  = ABC_FALLOC( int, Vec_IntSize(vMap2) );
    var vMap *Vec_Int_t  = Vec_IntStartFull( Gia_ManObjNum(p) );
    var i, Entry, iRepr, fCompl, iLit int32;
    var pObj *Gia_Obj_t;
    Gia_ManSetPhase( p1 );
    Gia_ManSetPhase( p2 );
    Vec_IntForEachEntry( vMap1, Entry, i )
    {
        if Entry == -1 {
            continue; }
        pObj = Gia_ManObj( p1, Abc_Lit2Var(Entry) );
        if ~pObj.Value == 0 {
            continue; }
        fCompl = Abc_LitIsCompl(Entry) ^ pObj.fPhase;
        iRepr = Gia_ObjReprSelf(p, Abc_Lit2Var(pObj.Value));
        Vec_IntWriteEntry( vMap, iRepr, Abc_Var2Lit(i, fCompl) );
    }
    Vec_IntForEachEntry( vMap2, Entry, i )
    {
        if Entry == -1 {
            continue; }
        pObj = Gia_ManObj( p2, Abc_Lit2Var(Entry) );
        if ~pObj.Value == 0 {
            continue; }
        fCompl = Abc_LitIsCompl(Entry) ^ pObj.fPhase;
        iRepr = Gia_ObjReprSelf(p, Abc_Lit2Var(pObj.Value));
        if (iLit = Vec_IntEntry(vMap, iRepr)) == -1 {
            continue; }
        pRes[i] = Abc_LitNotCond( iLit, fCompl );
    }
    Vec_IntFill( vMap, Gia_ManCoNum(p1), -1 );
    Vec_IntForEachEntry( vMap1, Entry, i )
    {
        if Entry == -1 {
            continue; }
        pObj = Gia_ManObj( p1, Abc_Lit2Var(Entry) );
        if !Gia_ObjIsCo(pObj) {
            continue; }
        Vec_IntWriteEntry( vMap, Gia_ObjCioId(pObj), i );
    }
    Vec_IntForEachEntry( vMap2, Entry, i )
    {
        if Entry == -1 {
            continue; }
        pObj = Gia_ManObj( p2, Abc_Lit2Var(Entry) );
        if !Gia_ObjIsCo(pObj) {
            continue; }
        assert(pRes[i] == -1, "pRes[i] == -1");
        pRes[i] = Abc_Var2Lit( Vec_IntEntry(vMap, Gia_ObjCioId(pObj)), 0 );
        assert(pRes[i] != -1, "pRes[i] != -1");
    }
    Vec_IntFree( vMap );
    return pRes;
}
func Gia_ManNameMapVerify(p *Gia_Man_t, p1 *Gia_Man_t, p2 *Gia_Man_t, vMap1 *Vec_Int_t, vMap2 *Vec_Int_t, pMap *int32) {
    var iLut int32;
    var iObj1 int32;
    var iObj2 int32;
    var nSize int32 = Vec_IntSize(vMap2);
    var pObjAig, pObjLut *Gia_Obj_t;
    Gia_ManSetPhase( p1 );
    Gia_ManSetPhase( p2 );
    for  iLut = 0; iLut < nSize; iLut++  {
        if pMap[iLut] >= 0 {
            var iObj int32    = Abc_Lit2Var( pMap[iLut] );
            var fCompl int32  = Abc_LitIsCompl( pMap[iLut] );
            var iLitAig int32  = Vec_IntEntry( vMap1, iObj );
            var iLitLut int32  = Vec_IntEntry( vMap2, iLut );
            pObjAig = Gia_ManObj( p1, Abc_Lit2Var(iLitAig) );
            if Gia_ObjIsCo(pObjAig) {
                continue; }
            if ~pObjAig.Value == 0 {
                continue; }
            pObjLut = Gia_ManObj( p2, Abc_Lit2Var(iLitLut) );
            if ~pObjLut.Value == 0 {
                continue; }
            iObj1 = Gia_ObjReprSelf(p, Abc_Lit2Var(pObjAig.Value));
            iObj2 = Gia_ObjReprSelf(p, Abc_Lit2Var(pObjLut.Value));
            if iObj1 != iObj2 {
                fmt.Printf("Found functional mismatch for LutId %d and AigId %d.\n", iLut, iObj); }
            if (pObjLut.fPhase ^ Abc_LitIsCompl(iLitLut)) != (pObjAig.fPhase ^ Abc_LitIsCompl(iLitAig) ^ fCompl) {
                fmt.Printf("Found phase mismatch for LutId %d and AigId %d.\n", iLut, iObj); }
        }
    }
}
func Abc_FrameReadMiniLutNameMapping(pAbc *Abc_Frame_t) *int32 {
    var fVerbose int32;
    var nConfs int32  = 1000;
    var pGia, pTemp *Gia_Man_t;
    var pRes *int32  = nil;
    if pAbc.pGiaMiniAig == nil {
        fmt.Printf("GIA derived from MiniAig is not available.\n"); }
    if pAbc.pGiaMiniLut == nil {
        fmt.Printf("GIA derived from MiniLut is not available.\n"); }
    if pAbc.pGiaMiniAig == nil || pAbc.pGiaMiniLut == nil {
        return nil; }
    pGia = Gia_ManDup2( pAbc.pGiaMiniAig, pAbc.pGiaMiniLut );
    //Gia_AigerWrite( pGia, "aig_m_lut.aig", 0, 0, 0 );
    // compute equivalences in this AIG
    pTemp = Gia_ManComputeGiaEquivs( pGia, nConfs, fVerbose );
    Gia_ManStop( pTemp );
    //if ( fVerbose )
    //    Abc_PrintTime( 1, "Equivalence computation time", Abc_Clock() - clk );
    //if ( fVerbose )
    //    Gia_ManPrintStats( pGia, NULL );
    //Vec_IntPrint( pAbc->vCopyMiniAig );
    //Vec_IntPrint( pAbc->vCopyMiniLut );
    pRes = Gia_ManMapMiniLut2MiniAig( pGia, pAbc.pGiaMiniAig, pAbc.pGiaMiniLut, pAbc.vCopyMiniAig, pAbc.vCopyMiniLut );
    //Gia_ManNameMapVerify( pGia, pAbc->pGiaMiniAig, pAbc->pGiaMiniLut, pAbc->vCopyMiniAig, pAbc->vCopyMiniLut, pRes );
    Gia_ManStop( pGia );
    return pRes;
}

/**Function*************************************************************

  Synopsis    [Returns equivalences of MiniAig nodes.]

  Description [The resulting array contains as many entries as there are objects
  in the initial MiniAIG. If the i-th entry of the array is equal to -1, it means
  that the i-th MiniAIG object is not equivalent to any other object. Otherwise, 
  the i-th entry contains the literal of the representative of the equivalence 
  class of objects, to which the i-th object belongs. The representative is defined 
  as the first object belonging to the equivalence class in the current topological 
  order. It can be the constant 0 node, a flop output or an internal node. It is 
  the user's responsibility to free the resulting array when it is not needed.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_FrameReadMiniAigEquivClasses(pAbc *Abc_Frame_t) *int32 {
    var vRes *Vec_Int_t;
    var pRes *int32;
    if pAbc.pGiaMiniAig == nil {
        fmt.Printf("GIA derived from MiniAig is not available.\n"); }
    if pAbc.vCopyMiniAig == nil {
        fmt.Printf("Mapping of MiniAig nodes is not available.\n"); }
    if pAbc.pGia2 == nil {
        fmt.Printf("Internal GIA with equivalence classes is not available.\n"); }
    if pAbc.pGia2.pReprs == nil {
        fmt.Printf("Equivalence classes of internal GIA are not available.\n"); }
    if Gia_ManObjNum(pAbc.pGia2) != Gia_ManObjNum(pAbc.pGiaMiniAig) {
        fmt.Printf("Internal GIA with equivalence classes is not directly derived from MiniAig.\n"); }
    // derive the set of equivalent node pairs
    vRes = Gia_ManMapEquivAfterScorr( pAbc.pGia2, pAbc.vCopyMiniAig );
    pRes = Vec_IntReleaseArray( vRes );
    Vec_IntFree( vRes );
    return pRes;
}

/**Function*************************************************************

  Synopsis    [Verifies equivalences of MiniAig nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_MiniAigVerify(pAbc *Abc_Frame_t, pFileName string) {
    var pEquivs *int32;
    var pGia *Gia_Man_t;
    var pFileMiter string    = "mini_aig_miter.aig";
    var pFileReduced string  = "mini_aig_reduced.aig";
    var p *Mini_Aig_t  = Mini_AigLoad( pFileName );
    Abc_FrameGiaInputMiniAig( pAbc, p );
    Cmd_CommandExecute( pAbc, "&ps; &scorr; &ps" );
    pEquivs = Abc_FrameReadMiniAigEquivClasses( pAbc );
    // dump miter for verification
    pGia = Gia_MiniAigMiter( p, pEquivs );
    Gia_AigerWrite( pGia, pFileMiter, 0, 0, 0 );
    fmt.Printf("Dumped miter AIG in file \"%s\".\n", pFileMiter);
    Gia_ManStop( pGia );
    // dump reduced AIG
    pGia = Gia_MiniAigReduce( p, pEquivs );
    Gia_AigerWrite( pGia, pFileReduced, 0, 0, 0 );
    fmt.Printf("Dumped reduced AIG in file \"%s\".\n", pFileReduced);
    Gia_ManStop( pGia );
    // cleanup
    pEquivs = nil;
    Mini_AigStop( p );
}

/**Function*************************************************************

  Synopsis    [Collects supergate for the outputs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_MiniAigSuperGates_rec(p *Mini_Aig_t, iObj int32, vRes *Vec_Int_t, vMap *Vec_Int_t) {
    var iFan0, iFan1 int32;
    if Mini_AigNodeIsPi(p, iObj) {
        assert(Vec_IntEntry(vMap, iObj) >= 0, "Vec_IntEntry(vMap, iObj) >= 0");
        Vec_IntPush( vRes, Vec_IntEntry(vMap, iObj) );
        return;
    }
    iFan0 = Mini_AigNodeFanin0( p, iObj );
    iFan1 = Mini_AigNodeFanin1( p, iObj );
    assert(!Abc_LitIsCompl(iFan0), "!Abc_LitIsCompl(iFan0)");
    assert(!Abc_LitIsCompl(iFan1), "!Abc_LitIsCompl(iFan1)");
    Gia_MiniAigSuperGates_rec( p, Abc_Lit2Var(iFan0), vRes, vMap );
    Gia_MiniAigSuperGates_rec( p, Abc_Lit2Var(iFan1), vRes, vMap );
}
func Gia_MiniAigSuperGates(p *Mini_Aig_t) *Vec_Wec_t {
    var vRes *Vec_Wec_t  = Vec_WecStart( Mini_AigPoNum(p) );
    var vMap *Vec_Int_t  = Vec_IntStartFull( Mini_AigNodeNum(p) );
    var i, Index int32;
    // inlining of macro Mini_AigForEachPi
    for i = 1; i < Mini_AigNodeNum(p); i++ { if !Mini_AigNodeIsPi(p, i) {} else { 
        Vec_IntWriteEntry( vMap, i, Index++ ); }
    }
    assert(Index == Mini_AigPiNum(p), "Index == Mini_AigPiNum(p)");
    Index = 0;
    // inlining of macro Mini_AigForEachPo
    for i = 1; i < Mini_AigNodeNum(p); i++ { if !Mini_AigNodeIsPo(p, i) {} else {
        var iFan0 int32  = Mini_AigNodeFanin0( p, i );
        assert(!Abc_LitIsCompl(iFan0), "!Abc_LitIsCompl(iFan0)");
        Gia_MiniAigSuperGates_rec( p, Abc_Lit2Var(iFan0), Vec_WecEntry(vRes, Index++), vMap );
    }
    }
    assert(Index == Mini_AigPoNum(p), "Index == Mini_AigPoNum(p)");
    Vec_IntFree( vMap );
    return vRes;
}

/**Function*************************************************************

  Synopsis    [Transform.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_MiniAigSuperPrintDouble(p *Vec_Int_t, nPis int32) {
    var i, Entry int32;
    fmt.Printf("\n");
    Vec_IntForEachEntry( p, Entry, i )
        fmt.Printf("%d(%d) ", Entry%nPis, Entry/nPis);
    fmt.Printf("  Total = %d\n", Vec_IntSize(p));
}
func Gia_MiniAigSuperMerge(p *Vec_Int_t, nPis int32) int32 {
    var i int32;
    var k int32;
    var This int32;
    var Prev int32 = -1;
    var fChange int32;
    Vec_IntForEachEntry( p, This, i )
    {
        if Prev == This {
            Vec_IntWriteEntry( p, k++, (This/nPis+1)*nPis + This%nPis );
            Prev = -1;
            fChange = 1;
        
        } else {
            if Prev != -1 { 
                Vec_IntWriteEntry( p, k++, Prev ); }
            Prev = This;
        }
    }
    if Prev != -1 {
        Vec_IntWriteEntry( p, k++, Prev ); }
    Vec_IntShrink( p, k );
    return fChange;
}
func Gia_MiniAigSuperPreprocess(p *Mini_Aig_t, vSuper *Vec_Wec_t, nPis int32, fVerbose int32) int32 {
    var vRes *Vec_Int_t;
    var i int32;
    var nIters int32;
    var Multi int32 = 1;
    Vec_WecForEachLevel( vSuper, vRes, i )
    {
        Vec_IntSort( vRes, 0 );
        if fVerbose != 0 { 
            fmt.Printf("\nOutput %d\n", i); }
        if fVerbose != 0 {
            Gia_MiniAigSuperPrintDouble( vRes, nPis ); }
        for  nIters = 1; (Gia_MiniAigSuperMerge(vRes, nPis) != 0); nIters++  {
            if fVerbose != 0 {
                Gia_MiniAigSuperPrintDouble( vRes, nPis ); }
        }
        Multi = Abc_MaxInt( Multi, nIters );
    }
    if fVerbose != 0 {
        fmt.Printf("Multi = %d.\n", Multi); }
    return Multi;
}

/**Function*************************************************************

  Synopsis    [Derive AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_MiniAigSuperDeriveGia(p *Vec_Wec_t, nPis int32, Multi int32) *Gia_Man_t {
    var pNew *Gia_Man_t;
    var vTemp *Vec_Int_t;
    var vLits *Vec_Int_t = Vec_IntAlloc( 100 );
    var vDrivers *Vec_Int_t  = Vec_IntAlloc(100);
    var i int32;
    var k int32;
    var iObj int32;
    var iLit int32;
    var nInputs int32 = nPis*Multi;
    pNew = Gia_ManStart( 1000 );
    pNew.pName = Abc_UtilStrsav( "tree" );
    for  i = 0; i < nInputs; i++  {
        Gia_ManAppendCi( pNew );
    }
    Gia_ManHashAlloc( pNew );
    Vec_WecForEachLevel( p, vTemp, i )
    {
        Vec_IntClear( vLits );
        Vec_IntForEachEntry( vTemp, iObj, k )
        {
            assert(iObj < nInputs, "iObj < nInputs");
            Vec_IntPush( vLits, 2+2*((iObj%nPis)*Multi+iObj/nPis) );
        }
        Vec_IntPush( vDrivers, Gia_ManHashAndMulti2(pNew, vLits) );
    }
    Gia_ManHashStop( pNew );
    Vec_IntFree( vLits );
    Vec_IntForEachEntry( vDrivers, iLit, i )
        Gia_ManAppendCo( pNew, iLit );
    Vec_IntFree( vDrivers );
    return pNew;
}
func Gia_MiniAigSuperDerive(pFileName string, fVerbose int32) *Gia_Man_t {
    var p *Mini_Aig_t      = Mini_AigLoad( pFileName );
    var vSuper *Vec_Wec_t  = Gia_MiniAigSuperGates( p );
    var Multi int32           = Gia_MiniAigSuperPreprocess( p, vSuper, Mini_AigPiNum(p), fVerbose );
    var pNew *Gia_Man_t    = Gia_MiniAigSuperDeriveGia( vSuper, Mini_AigPiNum(p), Multi );
    Vec_WecFree( vSuper );
    Mini_AigStop( p );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Process file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_MiniAigProcessFile() *Vec_Int_t {
    var vTriples *Vec_Int_t  = Vec_IntAlloc( 100 );
    var pFileName string  = "test.txt";
    var pFile *os.File  = os.Open(pFileName);
    if pFile == nil {
        fmt.Printf("Cannot open the file.\n");
    } else {
        var nLines, nLinesAll int32;
        var pToken string;
        var Buffer [1000]byte;
        for  fgets( Buffer, 1000, pFile ) != nil  {
            nLinesAll++;
            if Buffer[0] != '#' {
                continue; }
            //printf( "%s", Buffer );
            nLines++;
            pToken = strtok( Buffer+3, " \r\n\r+=" );
            for  pToken  {
                Vec_IntPush( vTriples, strconv.Atoi(pToken) );
                pToken = strtok( nil, " \r\n\r+=" );
            }
        }
        _ = (pFile).Close();
        fmt.Printf("Collected %d (out of %d) lines.\n", nLines, nLinesAll);
        fmt.Printf("Entries = %d\n", Vec_IntSize(vTriples));
    }
    return vTriples;
}
func Gia_MiniAigGenerate_rec(p *Mini_Aig_t, vTriples *Vec_Int_t, iObj int32, vDefs *Vec_Int_t, vMap *Vec_Int_t) {
    var Index, Entry0, Entry1, Entry2, Value int32;
    if Vec_IntEntry(vMap, iObj) >= 0 {
        return; }
    Index  = Vec_IntEntry( vDefs, iObj );
    Entry0 = Vec_IntEntry( vTriples, 3*Index+0 );
    Entry1 = Vec_IntEntry( vTriples, 3*Index+1 );
    Entry2 = Vec_IntEntry( vTriples, 3*Index+2 );
    Gia_MiniAigGenerate_rec( p, vTriples, Entry1, vDefs, vMap );
    Gia_MiniAigGenerate_rec( p, vTriples, Entry2, vDefs, vMap );
    assert(Vec_IntEntry(vMap, Entry1) >= 0, "Vec_IntEntry(vMap, Entry1) >= 0");
    assert(Vec_IntEntry(vMap, Entry2) >= 0, "Vec_IntEntry(vMap, Entry2) >= 0");
    Value  = Mini_AigAnd( p, Vec_IntEntry(vMap, Entry1), Vec_IntEntry(vMap, Entry2) );
    Vec_IntWriteEntry( vMap, Entry0, Value );
}
func Gia_MiniAigGenerateFromFile() {
    var p *Mini_Aig_t  = Mini_AigStart();
    var vTriples *Vec_Int_t  = Gia_MiniAigProcessFile();
    var vDefs *Vec_Int_t     = Vec_IntStartFull( Vec_IntSize(vTriples) ); 
    var vMap *Vec_Int_t      = Vec_IntStartFull( Vec_IntSize(vTriples) ); 
    var vMapIn *Vec_Int_t    = Vec_IntStart( Vec_IntSize(vTriples) );
    var vMapOut *Vec_Int_t   = Vec_IntStart( Vec_IntSize(vTriples) );
    var vPis *Vec_Int_t  = Vec_IntAlloc( 100 );
    var vPos *Vec_Int_t  = Vec_IntAlloc( 100 );
    var i, ObjOut, ObjIn int32;
    assert(Vec_IntSize(vTriples) % 3 == 0, "Vec_IntSize(vTriples) % 3 == 0");
    for  i = 0; i < Vec_IntSize(vTriples)/3; i++  {
        var Entry0 int32  = Vec_IntEntry(vTriples, 3*i+0);
        var Entry1 int32  = Vec_IntEntry(vTriples, 3*i+1);
        var Entry2 int32  = Vec_IntEntry(vTriples, 3*i+2);
        Vec_IntWriteEntry( vDefs,   Entry0, i );
        Vec_IntAddToEntry( vMapOut, Entry0, 1 );
        Vec_IntAddToEntry( vMapIn,  Entry1, 1 );
        Vec_IntAddToEntry( vMapIn,  Entry2, 1 );
    }
    Vec_IntForEachEntryTwo( vMapOut, vMapIn, ObjOut, ObjIn, i )
        if ( !ObjOut && ObjIn )
            Vec_IntPush( vPis, i );
        else if ( ObjOut && !ObjIn )
            Vec_IntPush( vPos, i );
    Vec_IntForEachEntry( vPis, ObjIn, i )
        Vec_IntWriteEntry( vMap, ObjIn, Mini_AigCreatePi(p) );
    Vec_IntForEachEntry( vPos, ObjOut, i )
        Gia_MiniAigGenerate_rec( p, vTriples, ObjOut, vDefs, vMap );
    Vec_IntForEachEntry( vPos, ObjOut, i )
    {
        assert(Vec_IntEntry(vMap, ObjOut) >= 0, "Vec_IntEntry(vMap, ObjOut) >= 0");
        Mini_AigCreatePo( p, Vec_IntEntry(vMap, ObjOut) );
    }
    Vec_IntFree( vTriples );
    Vec_IntFree( vDefs );
    Vec_IntFree( vMap );
    Vec_IntFree( vMapIn );
    Vec_IntFree( vMapOut );
    Vec_IntFree( vPis );
    Vec_IntFree( vPos );
    Mini_AigDump( p, "test.miniaig" );
    Mini_AigStop( p );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

