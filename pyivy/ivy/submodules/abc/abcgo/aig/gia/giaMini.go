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



/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


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



/**Function*************************************************************

  Synopsis    [Transform.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/




/**Function*************************************************************

  Synopsis    [Derive AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/



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

