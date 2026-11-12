package gia

import (
	"fmt"
	"strconv"
)

import "os"

/**CFile****************************************************************

  FileName    [giaGig.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [Parser for Gate-Inverter Graph by Niklas Een.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaGig.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"
//#include "misc/extra/extra.h"
//#include "misc/util/utilTruth.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

const MAX_LINE = 1000000

// network types
const ( 
    GLS_NONE = -1
  // not used
    GLS_ZERO =  0
  // zero
    GLS_ONE  =  1
  // one
    GLS_PI   =  2
  // primary input
    GLS_PO   =  3
  // primary output
    GLS_BAR  =  4
  // barrier
    GLS_SEQ  =  5
  // sequential
    GLS_SEL  =  6
  // fan
    GLS_LUT4 =  7
  // LUT4
    GLS_LUT6 =  8
  // LUT6
    GLS_BOX  =  9
  // sequential box
    GLS_DEL  = 10
  // delay box
    GLS_FINAL = GLS_DEL + 1

)


static char * s_Strs[GLS_FINAL] =
{
    "0",     // GLS_ZERO =  0,  // zero
    "1",     // GLS_ONE  =  1,  // one
    "PI",    // GLS_PI   =  2,  // primary input
    "PO",    // GLS_PO   =  3,  // primary output
    "Bar",   // GLS_BAR  =  4,  // barrier
    "Seq",   // GLS_SEQ  =  5,  // sequential
    "Sel",   // GLS_SEL  =  6,  // fan
    "Lut4",  // GLS_LUT4 =  7,  // LUT4
    "Lut6",  // GLS_LUT6 =  8,  // LUT6
    "Box",   // GLS_BOX  =  9,  // sequential box
    "Del"    // GLS_DEL  = 10,  // delay box
};

   
type Gls_Man_t struct {
    // general
    vLines     *Vec_Str_t;       // line types
    vTypes     *Vec_Str_t;       // gate types
    vIndexes     *Vec_Int_t;     // gate indexes
    // specific types
    vLut4s     *Vec_Int_t;       // 4-LUTs (4-tuples)
    vLut4TTs     *Vec_Int_t;     // truth tables
    vLut6s     *Vec_Int_t;       // 6-LUTs (6-tuples)
    vLut6TTs     *Vec_Wrd_t;     // truth tables
    vBoxes     *Vec_Int_t;       // boxes (5-tuples)
    vDelayIns     *Vec_Wec_t;    // delay fanins
    vDelayOuts     *Vec_Wec_t;   // delay fanouts
    vDelays     *Vec_Int_t;      // delay values
    // ordering
    vOrderPis     *Vec_Int_t;  
    vOrderPos     *Vec_Int_t;  
    vOrderBoxes     *Vec_Int_t;  
    vOrderDelays     *Vec_Int_t;  
    vOrderLuts     *Vec_Int_t;  
    vOrderSeqs     *Vec_Int_t;  
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
func Gls_ManAlloc(vLines *Vec_Str_t, pCounts *int32) *Gls_Man_t {
    // inlining of macro ABC_CALLOC
    var p *Gls_Man_t  = &Gls_Man_t{};
    p.vLines     = vLines;
    p.vTypes     = Vec_StrStart( Vec_StrSize(vLines)+100 ); 
    p.vIndexes   = Vec_IntStart( Vec_StrSize(vLines)+100 ); 
    p.vLut4s     = Vec_IntAlloc( 4 * pCounts[GLS_LUT4] );
    p.vLut4TTs   = Vec_IntAlloc( pCounts[GLS_LUT4] );
    p.vLut6s     = Vec_IntAlloc( 6 * pCounts[GLS_LUT6] );
    p.vLut6TTs   = Vec_WrdAlloc( pCounts[GLS_LUT6] );
    p.vBoxes     = Vec_IntAlloc( 5 * pCounts[GLS_BOX] );
    p.vDelays    = Vec_IntAlloc( pCounts[GLS_DEL] );
    p.vDelayIns  = Vec_WecAlloc( pCounts[GLS_DEL] );
    p.vDelayOuts = Vec_WecAlloc( pCounts[GLS_DEL] );
    // ordering
    p.vOrderPis    = Vec_IntAlloc( pCounts[GLS_PI] );
    p.vOrderPos    = Vec_IntAlloc( pCounts[GLS_PO] );
    p.vOrderBoxes  = Vec_IntAlloc( pCounts[GLS_BOX] );
    p.vOrderDelays = Vec_IntAlloc( pCounts[GLS_DEL] );
    p.vOrderLuts   = Vec_IntAlloc( pCounts[GLS_LUT4] + pCounts[GLS_LUT6] + 2*pCounts[GLS_BAR] );
    p.vOrderSeqs   = Vec_IntAlloc( pCounts[GLS_SEQ] );
    return p;
}
func Gls_ManStop(p *Gls_Man_t) {
    Vec_StrFree( p.vLines );
    Vec_StrFree( p.vTypes );
    Vec_IntFree( p.vIndexes );
    Vec_IntFree( p.vLut4s );
    Vec_IntFree( p.vLut4TTs );
    Vec_IntFree( p.vLut6s );
    Vec_WrdFree( p.vLut6TTs );
    Vec_IntFree( p.vBoxes );
    Vec_IntFree( p.vDelays );
    Vec_WecFree( p.vDelayIns );
    Vec_WecFree( p.vDelayOuts );
    // ordering
    Vec_IntFree( p.vOrderPis );
    Vec_IntFree( p.vOrderPos );
    Vec_IntFree( p.vOrderBoxes );
    Vec_IntFree( p.vOrderDelays );
    Vec_IntFree( p.vOrderLuts );
    Vec_IntFree( p.vOrderSeqs );
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gls_ManCount(pFile *os.File, pCounts []int32) *Vec_Str_t {
    // inlining of macro ABC_ALLOC
    char * pLine, * pBuffer = ((char *) malloc(sizeof(char) * (size_t)(MAX_LINE))); var Type int32;
    var vLines *Vec_Str_t  = Vec_StrAlloc( 10000 );
    memset( pCounts, 0, sizeof(int)*GLS_FINAL );
    for  fgets( pBuffer, MAX_LINE, pFile ) != nil  {
        pLine = pBuffer;
        for  *pLine  {
            if *pLine++ == '=' {
                break; }
        }
        for  *pLine == ' '  {
            pLine++;
        }
        if *pLine == 'L' {
            if pLine[3] == '4' {
                Type = GLS_LUT4;
            } else if pLine[3] == '6' {
                Type = GLS_LUT6;
            } else { assert(false, "0"); }
        
        } else if *pLine == 'P' {
            if pLine[1] == 'I' {
                Type = GLS_PI;
            } else if pLine[1] == 'O' {
                Type = GLS_PO;
            } else { assert(false, "0"); }
        
        } else if *pLine == 'B' {
            if pLine[1] == 'o' {
                Type = GLS_BOX;
            } else if pLine[1] == 'a' {
                Type = GLS_BAR;
            } else { assert(false, "0"); }
        
        } else if *pLine == 'S' {
            if pLine[2] == 'l' {
                Type = GLS_SEL;
            } else if pLine[2] == 'q' {
                Type = GLS_SEQ;
            } else { assert(false, "0"); }
        
        } else if *pLine == 'D' {
            Type = GLS_DEL;
        } else { assert(false, "0"); }
        Vec_StrPush( vLines, byte(Type) );
        pCounts[Type]++;
    }
    // inlining of macro ABC_FREE
    ((pBuffer) ? (free((char *) (pBuffer)), (pBuffer) = 0) : 0);
    return vLines;
}
func Gls_ManParseOne(ppLine *string) int32 {
    var Entry int32;
    var pLine string  = *ppLine;
    for  *pLine == ' '  {   pLine++; }
    if *pLine == '-' {
        Entry = GLS_NONE;
    } else if *pLine == '0' {
        Entry = 0;
    } else if *pLine == '1' {
        Entry = 1;
    } else if *pLine == 'w' {
        Entry = strconv.Atoi(++pLine);
    } else { assert(false, "0"); }
    for  *pLine == '-' || (*pLine >= '0' && *pLine <= '9')  {   pLine++; }
    for  *pLine == ' '  {   pLine++; }
    *ppLine = pLine;
    return Entry;
}
func Gls_ManParse(pFile *os.File, p *Gls_Man_t) int32 {
    // inlining of macro ABC_ALLOC
    char * pLine, * pBuffer = ((char *) malloc(sizeof(char) * (size_t)(MAX_LINE))); 
    var i, k, Type, iObj, Entry, iItem int32;  var Truth word;
    for  i = 0; fgets( pBuffer, MAX_LINE, pFile ) != nil; i++  {
        pLine = pBuffer;
        Type = Vec_StrEntry( p.vLines, i );
        iObj = Gls_ManParseOne( &pLine );
        Vec_StrWriteEntry( p.vTypes, iObj, byte(Type) );
        if Type == GLS_PI {
            Vec_IntPush( p.vOrderPis, iObj );
            Vec_IntWriteEntry( p.vIndexes, iObj, -1 );
            continue;
        }
        for  *pLine  {
            if *pLine++ == '(' {
                break; }
        }
        Entry = Gls_ManParseOne( &pLine );
        if Type == GLS_PO || Type == GLS_BAR || Type == GLS_SEQ || Type == GLS_SEL {
            if Type == GLS_PO {
                Vec_IntPush( p.vOrderPos, iObj );
            } else if Type == GLS_BAR {
                Vec_IntPush( p.vOrderLuts, iObj );
            } else if Type == GLS_SEQ {
                Vec_IntPush( p.vOrderSeqs, iObj );
            } else if Type == GLS_SEL {
                if int32(Vec_StrEntry(p.vTypes, Entry)) == GLS_DEL {
                    var vOuts *Vec_Int_t  = Vec_WecEntry( p.vDelayOuts, Vec_IntEntry(p.vIndexes, Entry) );
                    Vec_IntPush( vOuts, iObj );
                
                } else if int32(Vec_StrEntry(p.vTypes, Entry)) == GLS_BAR {
                    Vec_IntPush( p.vOrderLuts, iObj );
                } else { assert(false, "0"); }
            }
            Vec_IntWriteEntry( p.vIndexes, iObj, Entry );
            continue;
        }
        if Type == GLS_LUT4 {
            Vec_IntWriteEntry( p.vIndexes, iObj, Vec_IntSize(p.vLut4TTs) );
            Vec_IntPush( p.vLut4s, Entry );
            for  k = 1; ; k++  {
                if *pLine != ',' {     break; }
                pLine++;
                Entry = Gls_ManParseOne( &pLine );
                Vec_IntPush( p.vLut4s, Entry );
            }
            assert(*pLine == ')', "*pLine == ')'");
            assert(k == 4, "k == 4");
            pLine++;
            for  *pLine  {
                if *pLine++ == '[' {
                    break; }
            }
            Abc_TtReadHex( &Truth, pLine );
            Vec_IntPush( p.vLut4TTs, uint32(Truth) );
            Vec_IntPush( p.vOrderLuts, iObj );
        
        } else if Type == GLS_LUT6 {
            Vec_IntWriteEntry( p.vIndexes, iObj, Vec_WrdSize(p.vLut6TTs) );
            Vec_IntPush( p.vLut6s, Entry );
            for  k = 1; ; k++  {
                if *pLine != ',' {     break; }
                pLine++;
                Entry = Gls_ManParseOne( &pLine );
                Vec_IntPush( p.vLut6s, Entry );
            }
            assert(*pLine == ')', "*pLine == ')'");
            assert(k == 4, "k == 4");
            pLine++;
            for  *pLine  {
                if *pLine++ == '[' {
                    break; }
            }
            Abc_TtReadHex( &Truth, pLine );
            Vec_WrdPush( p.vLut6TTs, Truth );
            Vec_IntPush( p.vOrderLuts, iObj );
        
        } else if Type == GLS_BOX {
            Vec_IntWriteEntry( p.vIndexes, iObj, Vec_IntSize(p.vBoxes)/5 );
            Vec_IntPush( p.vBoxes, Entry );
            for  k = 1; ; k++  {
                if *pLine != ',' {     break; }
                pLine++;
                Entry = Gls_ManParseOne( &pLine );
                Vec_IntPush( p.vBoxes, Entry );
            }
            assert(*pLine == ')', "*pLine == ')'");
            assert(k == 4 || k == 5, "k == 4 || k == 5");
            if k == 4 {
                Vec_IntPush( p.vBoxes, GLS_NONE ); }
            Vec_IntPush( p.vOrderBoxes, iObj );
        
        } else if Type == GLS_DEL {
            var vIns *Vec_Int_t   = Vec_WecPushLevel( p.vDelayIns );
            var vOuts *Vec_Int_t  = Vec_WecPushLevel( p.vDelayOuts );
            Vec_IntWriteEntry( p.vIndexes, iObj, Vec_IntSize(p.vDelays) );
            Vec_IntPush( vIns, Entry );
            if *pLine != ')' {
                for  k = 1; ; k++  {
                    if *pLine != ',' {     break; }
                    pLine++;
                    Entry = Gls_ManParseOne( &pLine );
                    Vec_IntPush( vIns, Entry );
                }
            }
            assert(*pLine == ')', "*pLine == ')'");
            pLine++;
            for  *pLine  {
                if *pLine++ == '[' {
                    break; }
            }
            iItem = strconv.Atoi(pLine);
            Vec_IntPush( p.vDelays, iItem );
            Vec_IntPush( p.vOrderDelays, iObj );
            vOuts = vIns; // harmless use to prevent a compiler warning
        
        } else { assert(false, "0"); }
    }
    // inlining of macro ABC_FREE
    ((pBuffer) ? (free((char *) (pBuffer)), (pBuffer) = 0) : 0);
    return 1;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gls_ManConstruct(p *Gls_Man_t, pFileName string) *Gia_Man_t {
    extern int Kit_TruthToGia( Gia_Man_t * pMan, unsigned * pTruth, int nVars, Vec_Int_t * vMemory, Vec_Int_t * vLeaves, int fHash );
    var pGia *Gia_Man_t  = nil;  
    var vMap, vArray *Vec_Int_t;
    var vCover *Vec_Int_t  = Vec_IntAlloc(0);
    var vLeaves *Vec_Int_t  = Vec_IntAlloc(6);
    var k, iObj, iLit, Index int32;  var Type byte;
    // create new manager
    pGia = Gia_ManStart( Vec_StrSize(p.vTypes) );
    pGia.pName = Abc_UtilStrsav( pFileName );
    pGia.pSpec = Abc_UtilStrsav( pFileName );
    // create constants
    vMap = Vec_IntStartFull( Vec_StrSize(p.vTypes) );
    Vec_IntWriteEntry( vMap, 0, 0 );
    Vec_IntWriteEntry( vMap, 1, 1 );
    // create primary inputs
    // inlining of macro Vec_IntForEachEntry
    for  k = 0; (k < Vec_IntSize(p.vOrderPis)) && (((iObj) = Vec_IntEntry(p.vOrderPis, k)), 1); k++  {
        Vec_IntWriteEntry( vMap, iObj, Gia_ManAppendCi(pGia) );
    }
    // create box outputs
    // inlining of macro Vec_IntForEachEntry
    for  k = 0; (k < Vec_IntSize(p.vOrderBoxes)) && (((iObj) = Vec_IntEntry(p.vOrderBoxes, k)), 1); k++  {
        Vec_IntWriteEntry( vMap, iObj, Gia_ManAppendCi(pGia) );
    }
    // create delay outputs
    // inlining of macro Vec_IntForEachEntry
    for  Index = 0; (Index < Vec_IntSize(p.vOrderDelays)) && (((iObj) = Vec_IntEntry(p.vOrderDelays, Index)), 1); Index++  {
        assert(Index == Vec_IntEntry(p.vIndexes, iObj), "Index == Vec_IntEntry(p.vIndexes, iObj)");
        vArray = Vec_WecEntry(p.vDelayOuts, Index);
        if Vec_IntSize(vArray) == 0 {
            Vec_IntWriteEntry( vMap, iObj, Gia_ManAppendCi(pGia) );
        } else {
            // inlining of macro Vec_IntForEachEntry
            for  k = 0; (k < Vec_IntSize(vArray)) && (((iObj) = Vec_IntEntry(vArray, k)), 1); k++  {
                Vec_IntWriteEntry( vMap, iObj, Gia_ManAppendCi(pGia) );
            } }
    }
    // construct LUTs
    // inlining of macro Vec_IntForEachEntry
    for  Index = 0; (Index < Vec_IntSize(p.vOrderLuts)) && (((iObj) = Vec_IntEntry(p.vOrderLuts, Index)), 1); Index++  {
        Type = Vec_StrEntry( p.vTypes, iObj );
        if Type == GLS_LUT4 || Type == GLS_LUT6 {
            var Limit int32;
            if Type == GLS_LUT4 {
                Limit = 4;
            } else {
                Limit = 6;
            }
            var Index int32  = Vec_IntEntry(p.vIndexes, iObj);
            var pFanins *int32;
            if Type == GLS_LUT4 {
                pFanins = Vec_IntEntryP(p.vLut4s, 4*Index);
            } else {
                pFanins = Vec_IntEntryP(p.vLut6s, 6*Index);
            }
            var Truth word;
            if Type == GLS_LUT4 {
                Truth = (word)Vec_IntEntry(p.vLut4TTs, Index);
            } else {
                Truth = Vec_WrdEntry(p.vLut6TTs, Index);
            }
            Vec_IntClear( vLeaves );
            for  k = 0; k < Limit; k++  {
                Vec_IntPush( vLeaves, pFanins[k] == GLS_NONE ? 0 : Vec_IntEntry(vMap, pFanins[k]) );
            }
            iLit = Kit_TruthToGia( pGia, (unsigned *)&Truth, Vec_IntSize(vLeaves), vCover, vLeaves, 0 );
            Vec_IntWriteEntry( vMap, iObj, iLit );
        
        } else if Type == GLS_BAR || Type == GLS_SEL {
            iLit = Vec_IntEntry( vMap, Vec_IntEntry(p.vIndexes, iObj) );
            Vec_IntWriteEntry( vMap, iObj, iLit );
        }
    }
    // delay inputs
    // inlining of macro Vec_IntForEachEntry
    for  Index = 0; (Index < Vec_IntSize(p.vOrderDelays)) && (((iObj) = Vec_IntEntry(p.vOrderDelays, Index)), 1); Index++  {
        vArray = Vec_WecEntry(p.vDelayIns, Index);
        assert(Vec_IntSize(vArray) > 0, "Vec_IntSize(vArray) > 0");
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vArray)) && (((iObj) = Vec_IntEntry(vArray, k)), 1); k++  {
            Gia_ManAppendCo( pGia, Vec_IntEntry(vMap, iObj) );
        }
    }
    // create primary outputs
    // inlining of macro Vec_IntForEachEntry
    for  k = 0; (k < Vec_IntSize(p.vOrderPos)) && (((iObj) = Vec_IntEntry(p.vOrderPos, k)), 1); k++  {
        Gia_ManAppendCo( pGia, Vec_IntEntry(vMap, Vec_IntEntry(p.vIndexes, iObj)) );
    }
    // create sequential nodes
    // inlining of macro Vec_IntForEachEntry
    for  k = 0; (k < Vec_IntSize(p.vOrderSeqs)) && (((iObj) = Vec_IntEntry(p.vOrderSeqs, k)), 1); k++  {
        Gia_ManAppendCo( pGia, Vec_IntEntry(vMap, Vec_IntEntry(p.vIndexes, iObj)) );
    }
    Vec_IntFree( vMap );
    Vec_IntFree( vCover );
    Vec_IntFree( vLeaves );
    // print delay boxes
//    for ( k = 0; k < Vec_IntSize(p->vDelays); k++ )
//        printf( "%d:%d  ", Vec_IntSize(Vec_WecEntry(p->vDelayIns, k)), Vec_IntSize(Vec_WecEntry(p->vDelayOuts, k)) );
//    printf( "\n" );
    return pGia;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManReadGig(pFileName string) *Gia_Man_t {
    var clk abctime  = Abc_Clock();
    var p *Gls_Man_t  = nil;
    var pGia *Gia_Man_t  = nil;
    var vLines *Vec_Str_t;
    var i int32;
    var pCounts [GLS_FINAL]int32;
    var pFile *os.File  = os.Open(pFileName);
    if pFile == nil {
        fmt.Printf("Cannot read file \"%s\".\n", pFileName);
        return nil;
    }
    vLines = Gls_ManCount( pFile, pCounts );
    rewind( pFile );
    // statistics
    for  i = 0; i < GLS_FINAL; i++  {
        if pCounts[i] {
            fmt.Printf("%s=%d  ", s_Strs[i], pCounts[i]); }
    }
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
    // collect data and derive AIG
    p = Gls_ManAlloc( vLines, pCounts );
    if Gls_ManParse( pFile, p ) != 0 {
        pGia = Gls_ManConstruct( p, pFileName ); }
    Gls_ManStop( p );
    _ = (pFile).Close();
    //printf( "\n" );
    return pGia;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

