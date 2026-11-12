package io

import (
	"fmt"
	"strconv"
	"strings"
)

import "os"

/**CFile****************************************************************

  FileName    [ioReadBlifMv.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Command processing package.]

  Synopsis    [Procedures to read BLIF-MV file.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - January 8, 2007.]

  Revision    [$Id: ioReadBlifMv.c,v 1.00 2007/01/08 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "misc/zlib/zlib.h"
//#include "misc/bzlib/bzlib.h"
//#include "base/abc/abc.h"
//#include "misc/vec/vecPtr.h"
//#include "ioAbc.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

const IO_BLIFMV_MAXVALUES = 256
//#define IO_VERBOSE_OUTPUT

type Io_MvVar_t = Io_MvVar_t_; // parsing var
type Io_MvMod_t = Io_MvMod_t_; // parsing model
type Io_MvMan_t = Io_MvMan_t_; // parsing manager

var vGlobalLtlArray *Vec_Ptr_t;

type Io_MvVar_t_ struct {
    nValues                  int32;      // the number of values 
    pNames               *string;       // the value names
}

type Io_MvMod_t_ struct {
    // file lines
    pName                string;        // .model line
    vInputs           *Vec_Ptr_t;      // .inputs lines
    vOutputs           *Vec_Ptr_t;     // .outputs lines
    vLatches           *Vec_Ptr_t;     // .latch lines
    vFlops           *Vec_Ptr_t;       // .flop lines
    vResets           *Vec_Ptr_t;      // .reset lines
    vNames           *Vec_Ptr_t;       // .names lines
    vSubckts           *Vec_Ptr_t;     // .subckt lines
    vShorts           *Vec_Ptr_t;      // .short lines
    vOnehots           *Vec_Ptr_t;     // .onehot lines
    vMvs           *Vec_Ptr_t;         // .mv lines
    vConstrs           *Vec_Ptr_t;     // .constraint lines
    vLtlProperties              *Vec_Ptr_t;
    fBlackBox                  int32;    // indicates blackbox model
    // the resulting network
    pNtk           *Abc_Ntk_t;   
    pResetLatch           *Abc_Obj_t; 
    // the parent manager
    pMan          *Io_MvMan_t;         
}

type Io_MvMan_t_ struct {
    // general info about file
    fBlifMv                  int32;      // the file is BLIF-MV
    fUseReset                  int32;    // the reset circuitry is added
    pFileName                string;    // the name of the file
    pBuffer                string;      // the contents of the file
    vLines           *Vec_Ptr_t;       // the line beginnings
    // the results of reading
    pDesign           *Abc_Des_t;      // the design under construction
    nNDnodes                  int32;     // the counter of ND nodes
    // intermediate storage for models
    vModels           *Vec_Ptr_t;      // vector of models
    pLatest          *Io_MvMod_t;      // the current model
    // current processing info
    vTokens           *Vec_Ptr_t;      // the current tokens
    vTokens2           *Vec_Ptr_t;     // the current tokens
    vFunc           *Vec_Str_t;        // the local function
    // error reporting
    sError [512]byte;  // the error string generated during parsing
    // statistics 
    nTablesRead                  int32;  // the number of processed tables
    nTablesLeft                  int32;  // the number of dangling tables
}

// static functions
// static Io_MvMan_t *      Io_MvAlloc();

// static void              Io_MvFree( Io_MvMan_t * p );

// static Io_MvMod_t *      Io_MvModAlloc();

// static void              Io_MvModFree( Io_MvMod_t * p );

// static char *            Io_MvLoadFile( char * pFileName );

// static void              Io_MvReadPreparse( Io_MvMan_t * p );

// static int               Io_MvReadInterfaces( Io_MvMan_t * p );

// static Abc_Des_t *       Io_MvParse( Io_MvMan_t * p );

// static int               Io_MvParseLineModel( Io_MvMod_t * p, char * pLine );

// static int               Io_MvParseLineInputs( Io_MvMod_t * p, char * pLine );

// static int               Io_MvParseLineOutputs( Io_MvMod_t * p, char * pLine );

// static int               Io_MvParseLineConstrs( Io_MvMod_t * p, char * pLine );

// static int               Io_MvParseLineLatch( Io_MvMod_t * p, char * pLine );

// static int               Io_MvParseLineFlop( Io_MvMod_t * p, char * pLine );

// static int               Io_MvParseLineSubckt( Io_MvMod_t * p, char * pLine );

// static Vec_Int_t *       Io_MvParseLineOnehot( Io_MvMod_t * p, char * pLine );

// static int               Io_MvParseLineMv( Io_MvMod_t * p, char * pLine );

// static int               Io_MvParseLineNamesMv( Io_MvMod_t * p, char * pLine, int fReset );

// static int               Io_MvParseLineNamesBlif( Io_MvMod_t * p, char * pLine );

// static int               Io_MvParseLineShortBlif( Io_MvMod_t * p, char * pLine );

// static int                 Io_MvParseLineLtlProperty( Io_MvMod_t * p, char * pLine );

// static int               Io_MvParseLineGateBlif( Io_MvMod_t * p, Vec_Ptr_t * vTokens );

// static Io_MvVar_t *      Abc_NtkMvVarDup( Abc_Ntk_t * pNtk, Io_MvVar_t * pVar );


func Io_MvCharIsSpace(s byte) int32 { return s == ' ' || s == '\t' || s == '\r' || s == '\n';  }
func Io_MvCharIsMvSymb(s byte) int32 { return s == '(' || s == ')' || s == '{' || s == '}' || s == '-' || s == ',' || s == '!';  }

// extern void              Abc_NtkStartMvVars( Abc_Ntk_t * pNtk );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Reads the network from the BLIF or BLIF-MV file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_ReadBlifMv(pFileName string, fBlifMv int32, fCheck int32) *Abc_Ntk_t {
    var pFile *os.File;
    var p *Io_MvMan_t;
    var pNtk, pExdc *Abc_Ntk_t;
    var pDesign *Abc_Des_t  = nil; 
    var pDesignName string;
    var RetValue, i int32;
    var pLtlProp string;

    // check that the file is available
    pFile = os.Open(pFileName);
    if pFile == nil {
        fmt.Printf("Io_ReadBlifMv(): The file is unavailable (absent or open).\n");
        return 0;
    }
    _ = (pFile).Close();

    // start the file reader
    p = Io_MvAlloc();
    p.fBlifMv   = fBlifMv;
    p.fUseReset = 1;
    p.pFileName = pFileName;
    p.pBuffer   = Io_MvLoadFile( pFileName );
    if p.pBuffer == nil {
        Io_MvFree( p );
        return nil;
    }
    // set the design name
    pDesignName  = Extra_FileNameGeneric( pFileName );
    p.pDesign   = Abc_DesCreate( pDesignName );
    ABC_FREE( pDesignName );
    // free the HOP manager
    Hop_ManStop( (Hop_Man_t *)p.pDesign.pManFunc );
    p.pDesign.pManFunc = nil;
    // prepare the file for parsing
    Io_MvReadPreparse( p );
    // parse interfaces of each network and construct the network
    if Io_MvReadInterfaces( p ) != 0 {
        pDesign = Io_MvParse( p ); }
    if p.sError[0] {
        fmt.Fprintf(os.Stdout, "%s\n", p.sError); }
    Io_MvFree( p );
    if pDesign == nil {
        return nil; }
// pDesign should be linked to all models of the design

    // make sure that everything is okay with the network structure
    if fCheck != 0 {
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(pDesign.vModules)) && (((pNtk) = (Abc_Ntk_t *)Vec_PtrEntry(pDesign.vModules, i)), 1); i++  {
            if !Abc_NtkCheckRead( pNtk ) {
                fmt.Printf("Io_ReadBlifMv: The network check has failed for model %s.\n", pNtk.pName);
                Abc_DesFree( pDesign, nil );
                return nil;
            }
        }
    }

//Abc_DesPrint( pDesign );

    // check if there is an EXDC network
    if Vec_PtrSize(pDesign.vModules) > 1 {
        pNtk = (Abc_Ntk_t *)Vec_PtrEntry(pDesign.vModules, 0);
        // inlining of macro Vec_PtrForEachEntryStart
        for  i = 1; (i < Vec_PtrSize(pDesign.vModules)) && (((pExdc) = (Abc_Ntk_t *)Vec_PtrEntry(pDesign.vModules, i)), 1); i++  {
            if !(int32(strings.Compare(pExdc.pName, "EXDC")) != 0) {
                assert(pNtk.pExdc == nil, "pNtk.pExdc == nil");
                pNtk.pExdc = pExdc;
                Vec_PtrRemove(pDesign.vModules, pExdc);
                pExdc.pDesign = nil;
                i--;
            
            } else {
                pNtk = pExdc; }
        }
    }

    // detect top-level model
    RetValue = Abc_DesFindTopLevelModels( pDesign );
    pNtk = (Abc_Ntk_t *)Vec_PtrEntry( pDesign.vTops, 0 );
    if RetValue > 1 {
        fmt.Printf("Warning: The design has %d root-level modules. The first one (%s) will be used.\n", Vec_PtrSize(pDesign.vTops), pNtk.pName); }

    // extract the master network
    pNtk.pDesign = pDesign;
    pDesign.pManFunc = nil;

    // verify the design for cyclic dependence
    assert(Vec_PtrSize(pDesign.vModules) > 0, "Vec_PtrSize(pDesign.vModules) > 0");
    if Vec_PtrSize(pDesign.vModules) == 1 {
//        printf( "Warning: The design is not hierarchical.\n" );
        Abc_DesFree( pDesign, pNtk );
        pNtk.pDesign = nil;
        pNtk.pSpec = Extra_UtilStrsav( pFileName );
    
    } else {
        Abc_NtkIsAcyclicHierarchy( pNtk ); }

//Io_WriteBlifMv( pNtk, "_temp_.mv" );
    if pNtk.pSpec == nil {
        pNtk.pSpec = Extra_UtilStrsav( pFileName ); }

    vGlobalLtlArray = Vec_PtrAlloc( 100 );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vGlobalLtlArray)) && (((pLtlProp) = (char *)Vec_PtrEntry(vGlobalLtlArray, i)), 1); i++  {
        Vec_PtrPush( pNtk.vLtlProperties, pLtlProp );
    }
    Vec_PtrFreeP( &vGlobalLtlArray );
    return pNtk;
}

/**Function*************************************************************

  Synopsis    [Allocates the BLIF parsing structure.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_MvAlloc() *Io_MvMan_t {
    var p *Io_MvMan_t;
    p = new(Io_MvMan_t);
    memset( p, 0, sizeof(Io_MvMan_t) );
    p.vLines   = Vec_PtrAlloc( 512 );
    p.vModels  = Vec_PtrAlloc( 512 );
    p.vTokens  = Vec_PtrAlloc( 512 );
    p.vTokens2 = Vec_PtrAlloc( 512 );
    p.vFunc    = Vec_StrAlloc( 512 );
    return p;
}

/**Function*************************************************************

  Synopsis    [Frees the BLIF parsing structure.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_MvFree(p *Io_MvMan_t) {
    var pMod *Io_MvMod_t;
    var i int32;
    if p.pDesign {
        Abc_DesFree( p.pDesign, nil ); }
    if p.pBuffer {  
        ABC_FREE( p.pBuffer ); }
    if p.vLines {
        Vec_PtrFree( p.vLines  ); }
    if p.vModels {
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p.vModels)) && (((pMod) = (Io_MvMod_t *)Vec_PtrEntry(p.vModels, i)), 1); i++  {
            Io_MvModFree( pMod );
        }
        Vec_PtrFree( p.vModels );
    }
    Vec_PtrFree( p.vTokens );
    Vec_PtrFree( p.vTokens2 );
    Vec_StrFree( p.vFunc );
    ABC_FREE( p );
}

/**Function*************************************************************

  Synopsis    [Allocates the BLIF parsing structure for one model.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_MvModAlloc() *Io_MvMod_t {
    var p *Io_MvMod_t;
    p = new(Io_MvMod_t);
    memset( p, 0, sizeof(Io_MvMod_t) );
    p.vInputs  = Vec_PtrAlloc( 512 );
    p.vOutputs = Vec_PtrAlloc( 512 );
    p.vLatches = Vec_PtrAlloc( 512 );
    p.vFlops   = Vec_PtrAlloc( 512 );
    p.vResets  = Vec_PtrAlloc( 512 );
    p.vNames   = Vec_PtrAlloc( 512 );
    p.vSubckts = Vec_PtrAlloc( 512 );
    p.vShorts  = Vec_PtrAlloc( 512 );
    p.vOnehots = Vec_PtrAlloc( 512 );
    p.vMvs     = Vec_PtrAlloc( 512 );
    p.vConstrs = Vec_PtrAlloc( 512 );
    p.vLtlProperties = Vec_PtrAlloc( 512 );
    return p;
}

/**Function*************************************************************

  Synopsis    [Allocates the BLIF parsing structure for one model.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_MvModFree(p *Io_MvMod_t) {
//    if ( p->pNtk )
//        Abc_NtkDelete( p->pNtk );
    Vec_PtrFree( p.vLtlProperties );
    Vec_PtrFree( p.vInputs );
    Vec_PtrFree( p.vOutputs );
    Vec_PtrFree( p.vLatches );
    Vec_PtrFree( p.vFlops );
    Vec_PtrFree( p.vResets );
    Vec_PtrFree( p.vNames );
    Vec_PtrFree( p.vSubckts );
    Vec_PtrFree( p.vShorts );
    Vec_PtrFree( p.vOnehots );
    Vec_PtrFree( p.vMvs );
    Vec_PtrFree( p.vConstrs );
    ABC_FREE( p );
}



/**Function*************************************************************

  Synopsis    [Counts the number of given chars.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_MvCountChars(pLine string, Char byte) int32 {
    var pCur string;
    var Counter int32;
    for  pCur = pLine; *pCur; pCur++  {
        if *pCur == Char {
            Counter++; }
    }
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Returns the place where the arrow is hiding.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_MvFindArrow(pLine string) string {
    var pCur string;
    for  pCur = pLine; *(pCur+1); pCur++  {
        if *pCur == '-' && *(pCur+1) == '>' {
            *pCur = ' ';
            *(pCur+1) = ' ';
            return pCur;
        }
    }
    return nil;
}

/**Function*************************************************************

  Synopsis    [Collects the already split tokens.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_MvCollectTokens(vTokens *Vec_Ptr_t, pInput string, pOutput string) {
    var pCur string;
    Vec_PtrClear( vTokens );
    for  pCur = pInput; pCur < pOutput; pCur++  {
        if *pCur == 0 {
            continue; }
        Vec_PtrPush( vTokens, pCur );
        for  *++pCur  {; }
    }
}

/**Function*************************************************************

  Synopsis    [Splits the line into tokens.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_MvSplitIntoTokens(vTokens *Vec_Ptr_t, pLine string, Stop byte) {
    var pCur string;
    // clear spaces
    for  pCur = pLine; *pCur != Stop; pCur++  {
        if Io_MvCharIsSpace(*pCur) != 0 {
            *pCur = 0; }
    }
    // collect tokens
    Io_MvCollectTokens( vTokens, pLine, pCur );
}

/**Function*************************************************************

  Synopsis    [Splits the line into tokens when .default may be present.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_MvSplitIntoTokensMv(vTokens *Vec_Ptr_t, pLine string) {
    var pCur string;
    // clear spaces
    for  pCur = pLine; *pCur != '.' || *(pCur+1) == 'd'; pCur++  {
        if Io_MvCharIsSpace(*pCur) != 0 {
            *pCur = 0; }
    }
    // collect tokens
    Io_MvCollectTokens( vTokens, pLine, pCur );
}

/**Function*************************************************************

  Synopsis    [Splits the line into tokens.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_MvSplitIntoTokensAndClear(vTokens *Vec_Ptr_t, pLine string, Stop byte, Char byte) {
    var pCur string;
    // clear spaces
    for  pCur = pLine; *pCur != Stop; pCur++  {
        if (Io_MvCharIsSpace(*pCur) != 0) || *pCur == Char {
            *pCur = 0; }
    }
    // collect tokens
    Io_MvCollectTokens( vTokens, pLine, pCur );
}

/**Function*************************************************************

  Synopsis    [Returns the 1-based number of the line in which the token occurs.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_MvGetLine(p *Io_MvMan_t, pToken string) int32 {
    var pLine string;
    var i int32;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vLines)) && (((pLine) = (char *)Vec_PtrEntry(p.vLines, i)), 1); i++  {
        if pToken < pLine {
            return i; }
    }
    return -1;
}

/**Function*************************************************************

  Synopsis    [Reads the file into a character buffer.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
typedef struct buflist {
  char buf[1<<20];
  int nBuf;
  struct buflist * next;
}
 buflist;

func Io_MvLoadFileBz2(pFileName string, pnFileSize *int32) string {
    var pFile *os.File;
    var nFileSize int32;
    var pContents string;
    var b *BZFILE;
    var bzError, RetValue int32;
    struct buflist * pNext;
    var bufHead *buflist = nil;
    var buf *buflist = nil;

    pFile = os.Open(pFileName);
    if pFile == nil {
        Abc_Print( -1, "Io_MvLoadFileBz2(): The file is unavailable (absent or open).\n" );
        return nil;
    }
    b = BZ2_bzReadOpen(&bzError,pFile,0,0,nil,0);
    if bzError != BZ_OK {
        Abc_Print( -1, "Io_MvLoadFileBz2(): BZ2_bzReadOpen() failed with error %d.\n",bzError );
        return nil;
    }
    for {
        if !(bufHead != nil) {
            buf = bufHead = new(buflist);
        } else {
            buf = buf.next = new(buflist); }
        nFileSize += buf.nBuf = BZ2_bzRead(&bzError,b,buf.buf,1<<20);
        buf.next = nil;
    
if !(bzError == BZ_OK) { break }
} 
    if bzError == BZ_STREAM_END {
        // we're okay
        var p string;
        var nBytes int32;
        BZ2_bzReadClose(&bzError,b);
        p = pContents = ABC_ALLOC( char, nFileSize + 10 );
        buf = bufHead;
        for {
            memcpy(p+nBytes,buf.buf,(size_t)buf.nBuf);
            nBytes += buf.nBuf;
//        } while((buf = buf->next));
            pNext = buf.next;
            ABC_FREE( buf );
        
if !((buf = pNext)) { break }
} 
    } else if bzError == BZ_DATA_ERROR_MAGIC {
        // not a BZIP2 file
        BZ2_bzReadClose(&bzError,b);
        fseek( pFile, 0, SEEK_END );
        nFileSize = ftell( pFile );
        if nFileSize == 0 {
            Abc_Print( -1, "Io_MvLoadFileBz2(): The file is empty.\n" );
            return nil;
        }
        pContents = ABC_ALLOC( char, nFileSize + 10 );
        rewind( pFile );
        RetValue = fread( pContents, nFileSize, 1, pFile );
    } else { 
        // Some other error.
        Abc_Print( -1, "Io_MvLoadFileBz2(): Unable to read the compressed BLIF.\n" );
        return nil;
    }
    _ = (pFile).Close();
    // finish off the file with the spare .end line
    // some benchmarks suddenly break off without this line
    strcpy( pContents + nFileSize, "\n.end\n" );
    *pnFileSize = nFileSize;
    return pContents;
}

/**Function*************************************************************

  Synopsis    [Reads the file into a character buffer.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_MvLoadFileGz(pFileName string, pnFileSize *int32) string {
    int READ_BLOCK_SIZE = 100000;
    var pFile gzFile;
    var pContents string;
    var amtRead int32;
    var readBlock int32;
    var nFileSize int32 = READ_BLOCK_SIZE;
    pFile = gzopen( pFileName, "rb" ); // if pFileName doesn't end in ".gz" then this acts as a passthrough to fopen
    pContents = ABC_ALLOC( char, nFileSize );        
    readBlock = 0;
    for (amtRead = gzread(pFile, pContents + readBlock * READ_BLOCK_SIZE, READ_BLOCK_SIZE)) == READ_BLOCK_SIZE {
        //Abc_Print( 1,"%d: read %d bytes\n", readBlock, amtRead);
        nFileSize += READ_BLOCK_SIZE;
        pContents = ABC_REALLOC(char, pContents, nFileSize);
        ++readBlock;
    }
    //Abc_Print( 1,"%d: read %d bytes\n", readBlock, amtRead);
    assert(amtRead != -1, "amtRead != -1"); // indicates a zlib error
    nFileSize -= (READ_BLOCK_SIZE - amtRead);
    gzclose(pFile);
    *pnFileSize = nFileSize;
    return pContents;
}

/**Function*************************************************************

  Synopsis    [Reads the file into a character buffer.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_MvLoadFile(pFileName string) string {
    var pFile *os.File;
    var nFileSize int32;
    var pContents string;
    var RetValue int32;
    if !strncmp(pFileName+len(pFileName)-4,".bz2",4) {
        return Io_MvLoadFileBz2( pFileName, &nFileSize ); }
    if !strncmp(pFileName+len(pFileName)-3,".gz",3) {
        return Io_MvLoadFileGz( pFileName, &nFileSize ); }
    pFile = os.Open(pFileName);
    if pFile == nil {
        fmt.Printf("Io_MvLoadFile(): The file is unavailable (absent or open).\n");
        return nil;
    }
    fseek( pFile, 0, SEEK_END );  
    nFileSize = ftell( pFile ); 
    if nFileSize == 0 {
        _ = (pFile).Close();
        fmt.Printf("Io_MvLoadFile(): The file is empty.\n");
        return nil;
    }
    pContents = ABC_ALLOC( char, nFileSize + 10 );
    rewind( pFile );
    RetValue = fread( pContents, nFileSize, 1, pFile );
    _ = (pFile).Close();
    // finish off the file with the spare .end line
    // some benchmarks suddenly break off without this line
    strcpy( pContents + nFileSize, "\n.end\n" );
    return pContents;
}

/**Function*************************************************************

  Synopsis    [Prepares the parsing.]

  Description [Performs several preliminary operations:
  - Cuts the file buffer into separate lines.
  - Removes comments and line extenders.
  - Sorts lines by directives.
  - Estimates the number of objects.
  - Allocates room for the objects.
  - Allocates room for the hash table.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_MvReadPreparse(p *Io_MvMan_t) {
    var pCur, pPrev string;
    var i, fComment int32;
    // parse the buffer into lines and remove comments
    Vec_PtrPush( p.vLines, p.pBuffer );
    for  pCur = p.pBuffer; *pCur; pCur++  {
        if *pCur == '\n' {
            *pCur = 0;
//            if ( *(pCur-1) == '\r' )
//                *(pCur-1) = 0;
            fComment = 0;
            Vec_PtrPush( p.vLines, pCur + 1 );
        
        } else if *pCur == '#' {
            fComment = 1; }
        // remove comments
        if fComment != 0 {
            *pCur = 0; }
    }

    // unfold the line extensions and sort lines by directive
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vLines)) && (((pCur) = (char *)Vec_PtrEntry(p.vLines, i)), 1); i++  {
        if *pCur == 0 {
            continue; }
        // find previous non-space character
        for  pPrev = pCur - 2; pPrev >= p.pBuffer; pPrev--  {
            if !(Io_MvCharIsSpace(*pPrev) != 0) {
                break; }
        }
        // if it is the line extender, overwrite it with spaces
        if pPrev >= p.pBuffer && *pPrev == '\\' {
            for  ; *pPrev; pPrev++  {
                *pPrev = ' ';
            }
            *pPrev = ' ';
            continue;
        }
        // skip spaces at the beginning of the line
        for  Io_MvCharIsSpace(*pCur++)  {; }
        // parse directives
        if *(pCur-1) != '.' {
            continue; }
        if !strncmp(pCur, "names", 5) || !strncmp(pCur, "table", 5) || !strncmp(pCur, "gate", 4) {
            Vec_PtrPush( p.pLatest.vNames, pCur );
        } else if p.fBlifMv && (!strncmp(pCur, "def ", 4) || !strncmp(pCur, "default ", 8)) {
            continue;
        } else if !strncmp( pCur, "ltlformula", 10 ) {
            Vec_PtrPush( p.pLatest.vLtlProperties, pCur );
        } else if !strncmp(pCur, "latch", 5) {
            Vec_PtrPush( p.pLatest.vLatches, pCur );
        } else if !strncmp(pCur, "flop", 4) {
            Vec_PtrPush( p.pLatest.vFlops, pCur );
        } else if !strncmp(pCur, "r ", 2) || !strncmp(pCur, "reset ", 6) {
            Vec_PtrPush( p.pLatest.vResets, pCur );
        } else if !strncmp(pCur, "inputs", 6) {
            Vec_PtrPush( p.pLatest.vInputs, pCur );
        } else if !strncmp(pCur, "outputs", 7) {
            Vec_PtrPush( p.pLatest.vOutputs, pCur );
        } else if !strncmp(pCur, "subckt", 6) {
            Vec_PtrPush( p.pLatest.vSubckts, pCur );
        } else if !strncmp(pCur, "short", 5) {
            Vec_PtrPush( p.pLatest.vShorts, pCur );
        } else if !strncmp(pCur, "onehot", 6) {
            Vec_PtrPush( p.pLatest.vOnehots, pCur );
        } else if p.fBlifMv && !strncmp(pCur, "mv", 2) {
            Vec_PtrPush( p.pLatest.vMvs, pCur );
        } else if !strncmp(pCur, "constraint", 10) {
            Vec_PtrPush( p.pLatest.vConstrs, pCur );
        } else if !strncmp(pCur, "blackbox", 8) {
            p.pLatest.fBlackBox = 1;
        } else if !strncmp(pCur, "model", 5) {
            p.pLatest = Io_MvModAlloc();
            p.pLatest.pName = pCur;
            p.pLatest.pMan = p;
        
        } else if !strncmp(pCur, "end", 3) {
            if p.pLatest {
                Vec_PtrPush( p.vModels, p.pLatest ); }
            p.pLatest = nil;
        
        } else if !strncmp(pCur, "exdc", 4) {
//            fprintf( stdout, "Line %d: The design contains EXDC network (warning only).\n", Io_MvGetLine(p, pCur) );
            fmt.Fprintf(os.Stdout, "Warning: The design contains EXDC network.\n");
            if p.pLatest {
                Vec_PtrPush( p.vModels, p.pLatest ); }
            p.pLatest = Io_MvModAlloc();
            p.pLatest.pName = nil;
            p.pLatest.pMan = p;
        
        } else if !strncmp(pCur, "attrib", 6) {
        } else if !strncmp(pCur, "delay", 5) {
        } else if !strncmp(pCur, "input_", 6) {
        } else if !strncmp(pCur, "output_", 7) {
        } else if !strncmp(pCur, "no_merge", 8) {
        } else if !strncmp(pCur, "wd", 2) {
//        else if ( !strncmp(pCur, "inouts", 6) )
//        {}
        } else {
            pCur--;
            if pCur[len(pCur)-1] == '\r' {
                pCur[len(pCur)-1] = 0; }
            fmt.Fprintf(os.Stdout, "Line %d: Skipping line \"%s\".\n", Io_MvGetLine(p, pCur), pCur);
        }
    }
}

/**Function*************************************************************

  Synopsis    [Parses interfaces of the models.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_MvReadInterfaces(p *Io_MvMan_t) int32 {
    var pMod *Io_MvMod_t;
    var pLine string;
    var i, k, nOutsOld int32;
    // iterate through the models
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vModels)) && (((pMod) = (Io_MvMod_t *)Vec_PtrEntry(p.vModels, i)), 1); i++  {
        // parse the model
        if !(Io_MvParseLineModel( pMod, pMod.pName ) != 0) {
            return 0; }
        // add model to the design
        if !Abc_DesAddModel( p.pDesign, pMod.pNtk ) {
            sprintf( p.sError, "Line %d: Model %s is defined twice.", Io_MvGetLine(p, pMod.pName), pMod.pName );
            return 0;
        }
        // parse the inputs
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(pMod.vInputs)) && (((pLine) = (char *)Vec_PtrEntry(pMod.vInputs, k)), 1); k++  {
            if !(Io_MvParseLineInputs( pMod, pLine ) != 0) {
                return 0; }
        }
        // parse the outputs
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(pMod.vOutputs)) && (((pLine) = (char *)Vec_PtrEntry(pMod.vOutputs, k)), 1); k++  {
            if !(Io_MvParseLineOutputs( pMod, pLine ) != 0) {
                return 0; }
        }
        // parse the constraints
        nOutsOld = Abc_NtkPoNum(pMod.pNtk);
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(pMod.vConstrs)) && (((pLine) = (char *)Vec_PtrEntry(pMod.vConstrs, k)), 1); k++  {
            if !(Io_MvParseLineConstrs( pMod, pLine ) != 0) {
                return 0; }
        }
        pMod.pNtk.nConstrs = Abc_NtkPoNum(pMod.pNtk) - nOutsOld;
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(pMod.vLtlProperties)) && (((pLine) = (char *)Vec_PtrEntry(pMod.vLtlProperties, k)), 1); k++  {
            if !(Io_MvParseLineLtlProperty( pMod, pLine ) != 0) {
                return 0; }
        }
        // report the results
#ifdef IO_VERBOSE_OUTPUT
        if ( Vec_PtrSize(p->vModels) > 1 )
            printf( "Parsed %-32s: PI =%6d  PO =%6d  ND =%8d  FF =%6d  B =%6d\n", 
                pMod->pNtk->pName, Abc_NtkPiNum(pMod->pNtk), Abc_NtkPoNum(pMod->pNtk),
                Vec_PtrSize(pMod->vNames), Vec_PtrSize(pMod->vLatches), Vec_PtrSize(pMod->vSubckts) );
#endif
    }
    return 1;
}


/**Function*************************************************************

  Synopsis    []

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_MvParse(p *Io_MvMan_t) *Abc_Des_t {
    var pDesign *Abc_Des_t;
    var pMod *Io_MvMod_t;
    var pLine string;
    var i, k int32;
    // iterate through the models
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vModels)) && (((pMod) = (Io_MvMod_t *)Vec_PtrEntry(p.vModels, i)), 1); i++  { 
#ifdef IO_VERBOSE_OUTPUT
        if ( Vec_PtrSize(p->vModels) > 1 )
            printf( "Parsing model %s...\n", pMod->pNtk->pName );
#endif

        // check if there any MV lines
        if Vec_PtrSize(pMod.vMvs) > 0 {
            Abc_NtkStartMvVars( pMod.pNtk ); }
        // parse the mv lines
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(pMod.vMvs)) && (((pLine) = (char *)Vec_PtrEntry(pMod.vMvs, k)), 1); k++  {
            if !(Io_MvParseLineMv( pMod, pLine ) != 0) {
                return nil; }
        }
        // if reset lines are used there should be the same number of them as latches
        if Vec_PtrSize(pMod.vResets) > 0 {
            if Vec_PtrSize(pMod.vLatches) != Vec_PtrSize(pMod.vResets) {
                sprintf( p.sError, "Line %d: Model %s has different number of latches (%d) and reset nodes (%d).", 
                                    Io_MvGetLine(p, pMod.pName), Abc_NtkName(pMod.pNtk), Vec_PtrSize(pMod.vLatches), Vec_PtrSize(pMod.vResets) );
                return nil;
            }
            // create binary latch with 1-data and 0-init
            if p.fUseReset { 
                pMod.pResetLatch = Io_ReadCreateResetLatch( pMod.pNtk, p.fBlifMv ); }
        }
        // parse the flops
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(pMod.vFlops)) && (((pLine) = (char *)Vec_PtrEntry(pMod.vFlops, k)), 1); k++  {
            if !(Io_MvParseLineFlop( pMod, pLine ) != 0) {
                return nil; }
        }
        // parse the latches
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(pMod.vLatches)) && (((pLine) = (char *)Vec_PtrEntry(pMod.vLatches, k)), 1); k++  {
            if !(Io_MvParseLineLatch( pMod, pLine ) != 0) {
                return nil; }
        }
        // parse the reset lines
        if p.fUseReset {
            // inlining of macro Vec_PtrForEachEntry
            for  k = 0; (k < Vec_PtrSize(pMod.vResets)) && (((pLine) = (char *)Vec_PtrEntry(pMod.vResets, k)), 1); k++  {
                if !(Io_MvParseLineNamesMv( pMod, pLine, 1 ) != 0) {
                    return nil; }
            } }
        // parse the nodes
        if p.fBlifMv {
            // inlining of macro Vec_PtrForEachEntry
            for  k = 0; (k < Vec_PtrSize(pMod.vNames)) && (((pLine) = (char *)Vec_PtrEntry(pMod.vNames, k)), 1); k++  {
                if !(Io_MvParseLineNamesMv( pMod, pLine, 0 ) != 0) {
                    return nil; }
            }
        
        } else {
            // inlining of macro Vec_PtrForEachEntry
            for  k = 0; (k < Vec_PtrSize(pMod.vNames)) && (((pLine) = (char *)Vec_PtrEntry(pMod.vNames, k)), 1); k++  {
                if !(Io_MvParseLineNamesBlif( pMod, pLine ) != 0) {
                    return nil; }
            }
            // inlining of macro Vec_PtrForEachEntry
            for  k = 0; (k < Vec_PtrSize(pMod.vShorts)) && (((pLine) = (char *)Vec_PtrEntry(pMod.vShorts, k)), 1); k++  {
                if !(Io_MvParseLineShortBlif( pMod, pLine ) != 0) {
                    return nil; }
            }
        }
        // parse the subcircuits
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(pMod.vSubckts)) && (((pLine) = (char *)Vec_PtrEntry(pMod.vSubckts, k)), 1); k++  {
            if !(Io_MvParseLineSubckt( pMod, pLine ) != 0) {
                return nil; }
        }

        // allow for blackboxes without .blackbox line
        if Abc_NtkLatchNum(pMod.pNtk) == 0 && Abc_NtkNodeNum(pMod.pNtk) == 0 && Abc_NtkBoxNum(pMod.pNtk) == 0 {
            if pMod.pNtk.ntkFunc == ABC_FUNC_SOP {
                Mem_FlexStop( (Mem_Flex_t *)pMod.pNtk.pManFunc, 0 );
                pMod.pNtk.pManFunc = nil;
                pMod.pNtk.ntkFunc = ABC_FUNC_BLACKBOX;
            }
        }

        // finalize the network
        Abc_NtkFinalizeRead( pMod.pNtk );
        // read the one-hotness lines
        if Vec_PtrSize(pMod.vOnehots) > 0 {
            var vLine *Vec_Int_t; 
            var pObj *Abc_Obj_t;
            // set register numbers
            // inlining of macro Abc_NtkForEachLatch
            for  k = 0; (k < Vec_PtrSize((pMod.pNtk).vBoxes)) && (((pObj) = Abc_NtkBox(pMod.pNtk, k)), 1); k++  {           if !Abc_ObjIsLatch(pObj) {} else {
                pObj.pNext = (Abc_Obj_t *)(ABC_PTRINT_T)k; }
            }
            // derive register
            pMod.pNtk.vOnehots = Vec_PtrAlloc( Vec_PtrSize(pMod.vOnehots) );
            // inlining of macro Vec_PtrForEachEntry
            for  k = 0; (k < Vec_PtrSize(pMod.vOnehots)) && (((pLine) = (char *)Vec_PtrEntry(pMod.vOnehots, k)), 1); k++  {
                vLine = Io_MvParseLineOnehot( pMod, pLine );
                if vLine == nil {
                    return nil; }
                Vec_PtrPush( pMod.pNtk.vOnehots, vLine );
//                printf( "Parsed %d one-hot registers.\n", Vec_IntSize(vLine) );
            }
            // reset register numbers
            // inlining of macro Abc_NtkForEachLatch
            for  k = 0; (k < Vec_PtrSize((pMod.pNtk).vBoxes)) && (((pObj) = Abc_NtkBox(pMod.pNtk, k)), 1); k++  {           if !Abc_ObjIsLatch(pObj) {} else {
                pObj.pNext = nil; }
            }
            // print the result
            fmt.Printf("Parsed %d groups of 1-hot registers: { ", Vec_PtrSize(pMod.pNtk.vOnehots));
            // inlining of macro Vec_PtrForEachEntry
            for  k = 0; (k < Vec_PtrSize(pMod.pNtk.vOnehots)) && (((vLine) = (Vec_Int_t *)Vec_PtrEntry(pMod.pNtk.vOnehots, k)), 1); k++  {
                fmt.Printf("%d ", Vec_IntSize(vLine));
            }
            fmt.Printf("}\n");
            fmt.Printf("The total number of 1-hot registers = %d. (%.2f %%)\n", Vec_VecSizeSize( (Vec_Vec_t *)pMod.pNtk.vOnehots ), 100.0 * Vec_VecSizeSize( (Vec_Vec_t *)pMod.pNtk.vOnehots ) / Abc_NtkLatchNum(pMod.pNtk));
            {
                extern void Abc_GenOneHotIntervals( char * pFileName, int nPis, int nRegs, Vec_Ptr_t * vOnehots );
                var pFileName string  = Extra_FileNameGenericAppend( pMod.pMan.pFileName, "_1h.blif" );
                Abc_GenOneHotIntervals( pFileName, Abc_NtkPiNum(pMod.pNtk), Abc_NtkLatchNum(pMod.pNtk), pMod.pNtk.vOnehots );
                fmt.Printf("One-hotness condition is written into file \"%s\".\n", pFileName);
            }
        }
        if Vec_PtrSize(pMod.vFlops) {
            fmt.Printf("Warning: The parser converted %d .flop lines into .latch lines\n", Vec_PtrSize(pMod.vFlops));
            fmt.Printf("(information about set, reset, enable of the flops may be lost).\n");
        }

    }
    if p.nNDnodes {
//        printf( "Warning: The parser added %d PIs to replace non-deterministic nodes.\n", p->nNDnodes );
        fmt.Printf("Warning: The parser added %d constant 0 nodes to replace non-deterministic nodes.\n", p.nNDnodes); }
    // return the network
    pDesign = p.pDesign;
    p.pDesign = nil;
    return pDesign;
}

/**Function*************************************************************

  Synopsis    [Parses the model line.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_MvParseLineModel(p *Io_MvMod_t, pLine string) int32 {
    var vTokens *Vec_Ptr_t  = p.pMan.vTokens;
    var pToken, pPivot string;
    if pLine == nil {
        p.pNtk = Abc_NtkAlloc( ABC_NTK_NETLIST, ABC_FUNC_SOP, 1 );
        p.pNtk.pName = Extra_UtilStrsav( "EXDC" );
        return 1;
    }
    Io_MvSplitIntoTokens( vTokens, pLine, '\0' );
    pToken = (char *)Vec_PtrEntry( vTokens, 0 );
    assert(!(int32(strings.Compare(pToken, "model")) != 0), "!strcmp(pToken, \"model\")");
    if Vec_PtrSize(vTokens) != 2 {
        sprintf( p.pMan.sError, "Line %d: Model line has %d entries while it should have 2.", Io_MvGetLine(p.pMan, pToken), Vec_PtrSize(vTokens) );
        return 0;
    }
    if p.fBlackBox {
        p.pNtk = Abc_NtkAlloc( ABC_NTK_NETLIST, ABC_FUNC_BLACKBOX, 1 );
    } else if p.pMan.fBlifMv {
        p.pNtk = Abc_NtkAlloc( ABC_NTK_NETLIST, ABC_FUNC_BLIFMV, 1 );
    } else { 
        p.pNtk = Abc_NtkAlloc( ABC_NTK_NETLIST, ABC_FUNC_SOP, 1 ); }
//    for ( pPivot = pToken = Vec_PtrEntry(vTokens, 1); *pToken; pToken++ )
//        if ( *pToken == '/' || *pToken == '\\' )
//            pPivot = pToken+1;
    pPivot = pToken = (char *)Vec_PtrEntry(vTokens, 1);
    p.pNtk.pName = Extra_UtilStrsav( pPivot );
    return 1;
}

/**Function*************************************************************

  Synopsis    [Parses the inputs line.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_MvParseLineInputs(p *Io_MvMod_t, pLine string) int32 {
    var vTokens *Vec_Ptr_t  = p.pMan.vTokens;
    var pToken string;
    var i int32;
    Io_MvSplitIntoTokens( vTokens, pLine, '\0' );
    pToken = (char *)Vec_PtrEntry(vTokens, 0);
    assert(!(int32(strings.Compare(pToken, "inputs")) != 0), "!strcmp(pToken, \"inputs\")");
    // inlining of macro Vec_PtrForEachEntryStart
    for  i = 1; (i < Vec_PtrSize(vTokens)) && (((pToken) = (char *)Vec_PtrEntry(vTokens, i)), 1); i++  {
        Io_ReadCreatePi( p.pNtk, pToken );
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Parses the outputs line.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_MvParseLineOutputs(p *Io_MvMod_t, pLine string) int32 {
    var vTokens *Vec_Ptr_t  = p.pMan.vTokens;
    var pToken string;
    var i int32;
    Io_MvSplitIntoTokens( vTokens, pLine, '\0' );
    pToken = (char *)Vec_PtrEntry(vTokens, 0);
    assert(!(int32(strings.Compare(pToken, "outputs")) != 0), "!strcmp(pToken, \"outputs\")");
    // inlining of macro Vec_PtrForEachEntryStart
    for  i = 1; (i < Vec_PtrSize(vTokens)) && (((pToken) = (char *)Vec_PtrEntry(vTokens, i)), 1); i++  {
        Io_ReadCreatePo( p.pNtk, pToken );
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Parses the outputs line.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_MvParseLineConstrs(p *Io_MvMod_t, pLine string) int32 {
    var vTokens *Vec_Ptr_t  = p.pMan.vTokens;
    var pToken string;
    var i int32;
    Io_MvSplitIntoTokens( vTokens, pLine, '\0' );
    pToken = (char *)Vec_PtrEntry(vTokens, 0);
    assert(!(int32(strings.Compare(pToken, "constraint")) != 0), "!strcmp(pToken, \"constraint\")");
    // inlining of macro Vec_PtrForEachEntryStart
    for  i = 1; (i < Vec_PtrSize(vTokens)) && (((pToken) = (char *)Vec_PtrEntry(vTokens, i)), 1); i++  {
        Io_ReadCreatePo( p.pNtk, pToken );
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Parses the LTL property line.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_MvParseLineLtlProperty(p *Io_MvMod_t, pLine string) int32 {
    var i, j int32;
    var quoteBegin, quoteEnd int32;
    var keyWordLtlFormula [11]byte;
    var actualLtlFormula string;

    //checking if the line begins with the keyword "ltlformula" and
    //progressing the pointer forword
    for  i=0; i<10; i++  {
        keyWordLtlFormula[i] = pLine[i];
    }
    quoteBegin = i;
    keyWordLtlFormula[10] = '\0';
    assert(strcmp( "ltlformula", keyWordLtlFormula ) == 0, "strcmp( \"ltlformula\", keyWordLtlFormula ) == 0");
    for  pLine[i] != '"'  {
        i++;
    }
    quoteBegin = i;
    i = len(pLine);
    for  pLine[i] != '"'  {
        i--;
    }
    quoteEnd = i;
    actualLtlFormula = (char *)malloc( sizeof(char) * (quoteEnd - quoteBegin) );
    //printf("\nThe input ltl formula = ");
    for  i, j = quoteBegin + 1, 0; i<quoteEnd; i, j = i+1, j+1  {
        //printf("%c", pLine[i] );
        actualLtlFormula[j] = pLine[i];
    }
    actualLtlFormula[j] = '\0';
    Vec_PtrPush( vGlobalLtlArray, actualLtlFormula );
    return 1;
}


/**Function*************************************************************

  Synopsis    [Parses the latches line.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_MvParseLineLatch(p *Io_MvMod_t, pLine string) int32 {
    var vTokens *Vec_Ptr_t  = p.pMan.vTokens;
    var pObj, pNet *Abc_Obj_t;
    var pToken string;
    var Init int32;
    Io_MvSplitIntoTokens( vTokens, pLine, '\0' );
    pToken = (char *)Vec_PtrEntry(vTokens,0);
    assert(!(int32(strings.Compare(pToken, "latch")) != 0), "!strcmp(pToken, \"latch\")");
    if Vec_PtrSize(vTokens) < 3 {
        sprintf( p.pMan.sError, "Line %d: Latch does not have input name and output name.", Io_MvGetLine(p.pMan, pToken) );
        return 0;
    }
    // create latch
    if p.pResetLatch == nil {
        pObj = Io_ReadCreateLatch( p.pNtk, (char *)Vec_PtrEntry(vTokens,1), (char *)Vec_PtrEntry(vTokens,2) );
        // get initial value
        if p.pMan.fBlifMv {
            Abc_LatchSetInit0( pObj );
        } else {
            if Vec_PtrSize(vTokens) > 6 {
                fmt.Printf("Warning: Line %d has .latch directive with unrecognized entries (the total of %d entries).\n", Io_MvGetLine(p.pMan, pToken), Vec_PtrSize(vTokens)); } 
            if Vec_PtrSize(vTokens) > 3 {
                Init = strconv.Atoi((char *)Vec_PtrEntryLast(vTokens));
            } else {
                Init = 2; }
            if Init < 0 || Init > 3 {
                sprintf( p.pMan.sError, "Line %d: Initial state of the latch is incorrect \"%s\".", Io_MvGetLine(p.pMan, pToken), (char*)Vec_PtrEntry(vTokens,3) );
                return 0;
            }
            if Init == 0 {
                Abc_LatchSetInit0( pObj );
            } else if Init == 1 {
                Abc_LatchSetInit1( pObj );
            } else { // if ( Init == 2 )
                Abc_LatchSetInitDc( pObj ); }
        }
    
    } else {
        // get the net corresponding to the output of the latch
        pNet = Abc_NtkFindOrCreateNet( p.pNtk, (char *)Vec_PtrEntry(vTokens,2) );
        // get the net corresponding to the latch output (feeding into reset MUX)
        pNet = Abc_NtkFindOrCreateNet( p.pNtk, Abc_ObjNameSuffix(pNet, "_out") );
        // create latch
        pObj = Io_ReadCreateLatch( p.pNtk, (char *)Vec_PtrEntry(vTokens,1), Abc_ObjName(pNet) );
//        Abc_LatchSetInit0( pObj );
        Abc_LatchSetInit0( pObj );
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Parses the latches line.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_MvParseLineFlop(p *Io_MvMod_t, pLine string) int32 {
    var vTokens *Vec_Ptr_t  = p.pMan.vTokens;
    var pObj *Abc_Obj_t;
    var pToken, pOutput, pInput string;
    var i int32;
    var Init int32 = 2;
    assert(!p.pMan.fBlifMv, "!p.pMan.fBlifMv");
    Io_MvSplitIntoTokens( vTokens, pLine, '\0' );
    pToken = (char *)Vec_PtrEntry(vTokens,0);
    assert(!(int32(strings.Compare(pToken, "flop")) != 0), "!strcmp(pToken, \"flop\")");
    // get flop output
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vTokens)) && (((pToken) = (char *)Vec_PtrEntry(vTokens, i)), 1); i++  {
        if pToken[0] == 'Q' && pToken[1] == '=' {
            break; }
    }
    if i == Vec_PtrSize(vTokens) {
        sprintf( p.pMan.sError, "Line %d: Cannot find flop output.", Io_MvGetLine(p.pMan, (char *)Vec_PtrEntry(vTokens,0)) );
        return 0;
    }
    pOutput = pToken+2;
    // get flop input
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vTokens)) && (((pToken) = (char *)Vec_PtrEntry(vTokens, i)), 1); i++  {
        if pToken[0] == 'D' && pToken[1] == '=' {
            break; }
    }
    if i == Vec_PtrSize(vTokens) {
        sprintf( p.pMan.sError, "Line %d: Cannot find flop input.", Io_MvGetLine(p.pMan, (char *)Vec_PtrEntry(vTokens,0)) );
        return 0;
    }
    pInput = pToken+2;
    // create latch
    pObj = Io_ReadCreateLatch( p.pNtk, pInput, pOutput );
    // get the init value
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vTokens)) && (((pToken) = (char *)Vec_PtrEntry(vTokens, i)), 1); i++  {
        if !strncmp( pToken, "init=", 5 ) {
            Init = 0;
            if pToken[5] == '1' {
                Init = 1;
            } else if pToken[5] == '2' {
                Init = 2;
            } else if pToken[5] != '0' {
                sprintf( p.pMan.sError, "Line %d: Cannot read flop init value %s.", Io_MvGetLine(p.pMan, pToken), pToken );
                return 0;
            }
            break;
        }
    }
    if Init < 0 || Init > 2 {
        sprintf( p.pMan.sError, "Line %d: Initial state of the flop is incorrect \"%s\".", Io_MvGetLine(p.pMan, pToken), (char*)Vec_PtrEntry(vTokens,3) );
        return 0;
    }
    if Init == 0 {
        Abc_LatchSetInit0( pObj );
    } else if Init == 1 {
        Abc_LatchSetInit1( pObj );
    } else { // if ( Init == 2 )
        Abc_LatchSetInitDc( pObj ); }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Parses the subckt line.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_MvParseLineSubckt(p *Io_MvMod_t, pLine string) int32 {
    var vTokens *Vec_Ptr_t  = p.pMan.vTokens;
    var pModel *Abc_Ntk_t;
    var pBox, pNet, pTerm *Abc_Obj_t;
    var pToken string;
    var pName string;
    var pName2 string;
    var ppNames *string;
    var nEquals, Last, i, k int32;

    // split the line into tokens
    nEquals = Io_MvCountChars( pLine, '=' );
    Io_MvSplitIntoTokensAndClear( vTokens, pLine, '\0', '=' );
    pToken = (char *)Vec_PtrEntry(vTokens,0);
    assert(!(int32(strings.Compare(pToken, "subckt")) != 0), "!strcmp(pToken, \"subckt\")");
//printf( "%d ", nEquals );

    // get the model for this box
    pName = (char *)Vec_PtrEntry(vTokens,1);
    // skip instance name for now
    for  pToken = pName; *pToken; pToken++  {
        if *pToken == '|' {
            *pToken = 0;
            break;
        }
    }
    // find the model
    pModel = Abc_DesFindModelByName( p.pMan.pDesign, pName );
    if pModel == nil {
        sprintf( p.pMan.sError, "Line %d: Cannot find the model for subcircuit %s.", Io_MvGetLine(p.pMan, pToken), pName );
        return 0;
    }
/*
    // check if the number of tokens is correct
    if ( nEquals != Abc_NtkPiNum(pModel) + Abc_NtkPoNum(pModel) )
    {
        sprintf( p->pMan->sError, "Line %d: The number of ports (%d) in .subckt differs from the sum of PIs and POs of the model (%d).", 
            Io_MvGetLine(p->pMan, pToken), nEquals, Abc_NtkPiNum(pModel) + Abc_NtkPoNum(pModel) );
        return 0;
    }
*/
    // get the names
    ppNames = (char **)Vec_PtrArray(vTokens) + 2 + p.pMan.fBlifMv;

    // create the box with these terminals
    if Abc_NtkHasBlackbox(pModel) {
        pBox = Abc_NtkCreateBlackbox( p.pNtk );
    } else {
        pBox = Abc_NtkCreateWhitebox( p.pNtk ); }
    pBox.pData = pModel;
    if p.pMan.fBlifMv {
        Abc_ObjAssignName( pBox, (char *)Vec_PtrEntry(vTokens,2), nil ); }
    // go through formal inputs
    Last = 0;
    // inlining of macro Abc_NtkForEachPi
    for  i = 0; (i < Abc_NtkPiNum(pModel)) && (((pTerm) = Abc_NtkPi(pModel, i)), 1); i++  { 
        // find this terminal among the actual inputs of the subcircuit
        pName2 = nil;
        pName = Abc_ObjName(Abc_ObjFanout0(pTerm));
        for  k = 0; k < nEquals; k++  {
            if !(int32(strings.Compare(ppNames[2*((k+Last)%nEquals)], pName)) != 0) {
                pName2 = ppNames[2*((k+Last)%nEquals)+1];
                Last = k+Last+1;
                break;
            }
        }
/*
        if ( k == nEquals )
        {
            sprintf( p->pMan->sError, "Line %d: Cannot find PI \"%s\" of the model \"%s\" as a formal input of the subcircuit.", 
                Io_MvGetLine(p->pMan, pToken), pName, Abc_NtkName(pModel) );
            return 0;
        }
*/
        if pName2 == nil {
            var pNode *Abc_Obj_t  = Abc_NtkCreateNodeConst0( p.pNtk );
            //pNode->pData = Abc_SopRegister( (Mem_Flex_t *)p->pNtk->pManFunc, " 0\n" );
            pNet = Abc_NtkFindOrCreateNet( p.pNtk, Abc_ObjNameSuffix(pNode, "abc") );
            Abc_ObjAddFanin( pNet, pNode );
            pTerm = Abc_NtkCreateBi( p.pNtk );
            Abc_ObjAddFanin( pBox, pTerm );
            Abc_ObjAddFanin( pTerm, pNet );
            continue;
        }
        assert(pName2 != nil, "pName2 != nil");
  
        // create the BI with the actual name
        pNet = Abc_NtkFindOrCreateNet( p.pNtk, pName2 );
        pTerm = Abc_NtkCreateBi( p.pNtk );
        Abc_ObjAddFanin( pBox, pTerm );
        Abc_ObjAddFanin( pTerm, pNet );
    }
    // go through formal outputs
    Last = 0;
    // inlining of macro Abc_NtkForEachPo
    for  i = 0; (i < Abc_NtkPoNum(pModel)) && (((pTerm) = Abc_NtkPo(pModel, i)), 1); i++  {
        // find this terminal among the actual outputs of the subcircuit
        pName2 = nil;
        pName = Abc_ObjName(Abc_ObjFanin0(pTerm));
        for  k = 0; k < nEquals; k++  {
            if !(int32(strings.Compare(ppNames[2*((k+Last)%nEquals)], pName)) != 0) {
                pName2 = ppNames[2*((k+Last)%nEquals)+1];
                Last = k+Last+1;
                break;
            }
        }
/*
        if ( k == nEquals )
        {
            sprintf( p->pMan->sError, "Line %d: Cannot find PO \"%s\" of the modell \"%s\" as a formal output of the subcircuit.", 
                Io_MvGetLine(p->pMan, pToken), pName, Abc_NtkName(pModel) );
            return 0;
        }
*/

        // create the BI with the actual name
        pTerm = Abc_NtkCreateBo( p.pNtk );
        pNet = Abc_NtkFindOrCreateNet( p.pNtk, pName2 == nil  ? Abc_ObjNameSuffix(pTerm, "abc") : pName2 );
        Abc_ObjAddFanin( pNet, pTerm );
        Abc_ObjAddFanin( pTerm, pBox );
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Parses the subckt line.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_MvParseLineOnehot(p *Io_MvMod_t, pLine string) *Vec_Int_t {
    var vTokens *Vec_Ptr_t  = p.pMan.vTokens;
//    Vec_Ptr_t * vResult;
    var vResult *Vec_Int_t;
    var pNet, pTerm *Abc_Obj_t;
    var pToken string;
    var nEquals, i int32;

    // split the line into tokens
    nEquals = Io_MvCountChars( pLine, '=' );
    Io_MvSplitIntoTokensAndClear( vTokens, pLine, '\0', '=' );
    pToken = (char *)Vec_PtrEntry(vTokens,0);
    assert(!(int32(strings.Compare(pToken, "onehot")) != 0), "!strcmp(pToken, \"onehot\")");

    // iterate through the register names
//    vResult = Vec_PtrAlloc( Vec_PtrSize(vTokens) );
    vResult = Vec_IntAlloc( Vec_PtrSize(vTokens) );
    // inlining of macro Vec_PtrForEachEntryStart
    for  i = 1; (i < Vec_PtrSize(vTokens)) && (((pToken) = (char *)Vec_PtrEntry(vTokens, i)), 1); i++  {
        // check if this register exists
        pNet = Abc_NtkFindNet( p.pNtk, pToken );
        if pNet == nil {
            sprintf( p.pMan.sError, "Line %d: Signal with name \"%s\" does not exist in the model \"%s\".", 
                            Io_MvGetLine(p.pMan, pToken), pToken, Abc_NtkName(p.pNtk) );
            return nil;
        }
        // check if this is register output net
        pTerm = Abc_ObjFanin0( pNet );
        if pTerm == nil || Abc_ObjFanin0(pTerm) == nil || !Abc_ObjIsLatch(Abc_ObjFanin0(pTerm)) {
            sprintf( p.pMan.sError, "Line %d: Signal with name \"%s\" is not a register in the model \"%s\".", 
                            Io_MvGetLine(p.pMan, pToken), pToken, Abc_NtkName(p.pNtk) );
            return nil;
        }
        // save register name
//        Vec_PtrPush( vResult, Abc_ObjName(pNet) );
        Vec_IntPush( vResult, int32((ABC_PTRINT_T))Abc_ObjFanin0(pTerm).pNext );
//        printf( "%d(%d) ", (int)Abc_ObjFanin0(pTerm)->pNext, ((int)Abc_ObjFanin0(pTerm)->pData) -1 );
        fmt.Printf("%d", (int32((ABC_PTRINT_T))Abc_ObjFanin0(pTerm).pData)-1);
    }
    fmt.Printf("\n");
    return vResult;
}


/**Function*************************************************************

  Synopsis    [Parses the mv line.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_MvParseLineMv(p *Io_MvMod_t, pLine string) int32 {
    var vTokens *Vec_Ptr_t  = p.pMan.vTokens;
    var pObj *Abc_Obj_t;
    var pVar *Io_MvVar_t  = nil;
    var pFlex *Mem_Flex_t;
    var pName string;
    var nCommas, nValues, i, k int32;
    // count commas and get the tokens
    nCommas = Io_MvCountChars( pLine, ',' );
    Io_MvSplitIntoTokensAndClear( vTokens, pLine, '\0', ',' );
    pName = (char *)Vec_PtrEntry(vTokens,0);
    assert(!(int32(strings.Compare(pName, "mv")) != 0), "!strcmp(pName, \"mv\")");
    // get the number of values
    if Vec_PtrSize(vTokens) <= nCommas + 2 {
        sprintf( p.pMan.sError, "Line %d: The number of values in not specified in .mv line.", Io_MvGetLine(p.pMan, pName) );
        return 0;
    }
    nValues = strconv.Atoi((char *)Vec_PtrEntry(vTokens,nCommas+2));
    if nValues < 2 || nValues > IO_BLIFMV_MAXVALUES {
        sprintf( p.pMan.sError, "Line %d: The number of values (%d) is incorrect (should be >= 2 and <= %d).", 
                    Io_MvGetLine(p.pMan, pName), nValues, IO_BLIFMV_MAXVALUES );
        return 0;
    }
    // if there is no symbolic values, quit
    if nValues == 2 && Vec_PtrSize(vTokens) == nCommas + 3 {
        return 1; }
    if Vec_PtrSize(vTokens) > nCommas + 3 && Vec_PtrSize(vTokens) - (nCommas + 3) != nValues {
        sprintf( p.pMan.sError, "Line %d: Wrong number (%d) of symbolic value names (should be %d).", 
                    Io_MvGetLine(p.pMan, pName), Vec_PtrSize(vTokens) - (nCommas + 3), nValues );
        return 0;
    }
    // go through variables
    pFlex = (Mem_Flex_t *)Abc_NtkMvVarMan( p.pNtk );
    for  i = 0; i <= nCommas; i++  {
        pName = (char *)Vec_PtrEntry( vTokens, i+1 );
        pObj = Abc_NtkFindOrCreateNet( p.pNtk, pName );
        // allocate variable
        pVar = (Io_MvVar_t *)Mem_FlexEntryFetch( pFlex, sizeof(Io_MvVar_t) );
        pVar.nValues = nValues;
        pVar.pNames = nil;
        // create names
        if Vec_PtrSize(vTokens) > nCommas + 3 {
            pVar.pNames = (char **)Mem_FlexEntryFetch( pFlex, sizeof(char *) * nValues );
            // inlining of macro Vec_PtrForEachEntryStart
            for  k = nCommas + 3; (k < Vec_PtrSize(vTokens)) && (((pName) = (char *)Vec_PtrEntry(vTokens, k)), 1); k++  {
                pVar.pNames[k-(nCommas + 3)] = (char *)Mem_FlexEntryFetch( pFlex, len(pName) + 1 );
                strcpy( pVar.pNames[k-(nCommas + 3)], pName );
            }
        }
        // save the variable
        Abc_ObjSetMvVar( pObj, pVar );
    }
    // make sure the names are unique
    assert((pVar != nil), "pVar");
    if pVar.pNames {
        for  i = 0; i < nValues; i++  {
        for  k = i+1; k < nValues; k++  {
            if !(int32(strings.Compare(pVar.pNames[i], pVar.pNames[k])) != 0) {
                pName = (char *)Vec_PtrEntry(vTokens,0);
                sprintf( p.pMan.sError, "Line %d: Symbolic value name \"%s\" is repeated in .mv line.", 
                                    Io_MvGetLine(p.pMan, pName), pVar.pNames[i] );
                return 0;
            }
        }
        }
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Writes the values into the BLIF-MV representation for the node.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_MvWriteValues(pNode *Abc_Obj_t, vFunc *Vec_Str_t) int32 {
    var Buffer string // was [10]byte in C

    var pFanin *Abc_Obj_t;
    var i int32;
    // add the fanin number of values
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pNode)) && (((pFanin) = Abc_ObjFanin(pNode, i)), 1); i++  {
        Buffer = fmt.Sprintf("%d", Abc_ObjMvVarNum(pFanin));
        Vec_StrPrintStr( vFunc, Buffer );
        Vec_StrPush( vFunc, ' ' );
    }
    // add the node number of values
    Buffer = fmt.Sprintf("%d", Abc_ObjMvVarNum(Abc_ObjFanout0(pNode)));
    Vec_StrPrintStr( vFunc, Buffer );
    Vec_StrPush( vFunc, '\n' );
    return 1;
}

/**Function*************************************************************

  Synopsis    [Translated one literal.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_MvParseLiteralMv(p *Io_MvMod_t, pNode *Abc_Obj_t, pToken string, vFunc *Vec_Str_t, iLit int32) int32 {
    var Buffer string // was [10]byte in C

    var pVar *Io_MvVar_t;
    var pFanin, pNet *Abc_Obj_t;
    var pCur, pNext string;
    var i int32;
    // consider the equality literal
    if pToken[0] == '=' {
        // find the fanins
        // inlining of macro Abc_ObjForEachFanin
        for  i = 0; (i < Abc_ObjFaninNum(pNode)) && (((pFanin) = Abc_ObjFanin(pNode, i)), 1); i++  {
            if !(int32(strings.Compare(Abc_ObjName(pFanin), pToken + 1)) != 0) {
                break; }
        }
        if i == Abc_ObjFaninNum(pNode) {
            sprintf( p.pMan.sError, "Line %d: Node name in the table \"%s\" cannot be found on .names line.", 
                            Io_MvGetLine(p.pMan, pToken), pToken + 1 );
            return 0;
        }
        Vec_StrPush( vFunc, '=' );
        Buffer = fmt.Sprintf("%d", i);
        Vec_StrPrintStr( vFunc, Buffer );
        Vec_StrPush( vFunc, byte(((iLit == -1)? '\n' : ' ')) );
        return 1;
    }
    // consider regular literal
    assert(iLit < Abc_ObjFaninNum(pNode), "iLit < Abc_ObjFaninNum(pNode)");
    if iLit >= 0 {
        pNet = Abc_ObjFanin(pNode, iLit);
    } else {
        pNet = Abc_ObjFanout0(pNode);
    }
    pVar = (Io_MvVar_t *)Abc_ObjMvVar( pNet );
    // if the var is absent or has no symbolic values quit
    if pVar == nil || pVar.pNames == nil {
        Vec_StrPrintStr( vFunc, pToken );
        Vec_StrPush( vFunc, byte(((iLit == -1)? '\n' : ' ')) );
        return 1;
    }
    // parse the literal using symbolic values
    for  pCur = pToken; *pCur; pCur++  {
        if Io_MvCharIsMvSymb(*pCur) != 0 {
            Vec_StrPush( vFunc, *pCur );
            continue;
        }
        // find the next MvSymb char
        for  pNext = pCur+1; *pNext; pNext++  {
            if Io_MvCharIsMvSymb(*pNext) != 0 {
                break; }
        }
        // look for the value name
        for  i = 0; i < pVar.nValues; i++  {
            if !strncmp( pVar.pNames[i], pCur, pNext-pCur ) {
                break; }
        }
        if i == pVar.nValues {
            *pNext = 0;
            sprintf( p.pMan.sError, "Line %d: Cannot find value name \"%s\" among the value names of variable \"%s\".", 
                            Io_MvGetLine(p.pMan, pToken), pCur, Abc_ObjName(pNet) );
            return 0;
        }
        // value name is found
        Buffer = fmt.Sprintf("%d", i);
        Vec_StrPrintStr( vFunc, Buffer );
        // update the pointer
        pCur = pNext - 1;
    }
    Vec_StrPush( vFunc, byte(((iLit == -1)? '\n' : ' ')) );
    return 1;
}

/**Function*************************************************************

  Synopsis    [Constructs the MV-SOP cover from the file parsing info.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_MvParseTableMv(p *Io_MvMod_t, pNode *Abc_Obj_t, vTokens2 *Vec_Ptr_t, nInputs int32, nOutputs int32, iOut int32) string {
    var vFunc *Vec_Str_t  = p.pMan.vFunc;
    var pFirst, pToken string;
    var iStart, i int32;
    // prepare the place for the cover
    Vec_StrClear( vFunc );
    // write the number of values
//    Io_MvWriteValues( pNode, vFunc );
    // get the first token
    pFirst = (char *)Vec_PtrEntry( vTokens2, 0 );
    if pFirst[0] == '.' {
        // write the default literal
        Vec_StrPush( vFunc, 'd' );
        pToken = (char *)Vec_PtrEntry(vTokens2, 1 + iOut);
        if !(Io_MvParseLiteralMv( p, pNode, pToken, vFunc, -1 ) != 0) {
            return nil; }
        iStart = 1 + nOutputs;
    
    } else {
        iStart = 0; }
    // write the remaining literals
    for  iStart < Vec_PtrSize(vTokens2)  {
        // input literals
        for  i = 0; i < nInputs; i++  {
            pToken = (char *)Vec_PtrEntry( vTokens2, iStart + i );
            if !(Io_MvParseLiteralMv( p, pNode, pToken, vFunc, i ) != 0) {
                return nil; }
        }
        // output literal
        pToken = (char *)Vec_PtrEntry( vTokens2, iStart + nInputs + iOut );
        if !(Io_MvParseLiteralMv( p, pNode, pToken, vFunc, -1 ) != 0) {
            return nil; }
        // update the counter
        iStart += nInputs + nOutputs;
    }       
    Vec_StrPush( vFunc, '\0' );
    return Vec_StrArray( vFunc );
}

/**Function*************************************************************

  Synopsis    [Adds reset circuitry corresponding to latch with pName.]

  Description [Returns the reset node's net.]
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_MvParseAddResetCircuit(p *Io_MvMod_t, pName string) *Abc_Obj_t {
    var Buffer string // was [50]byte in C

    var pNode, pData0Net, pData1Net, pResetLONet, pOutNet *Abc_Obj_t;
    var pVar *Io_MvVar_t;
    // make sure the reset latch exists
    assert(p.pResetLatch != nil, "p.pResetLatch != nil");
    // get the reset net
    pResetLONet = Abc_ObjFanout0(Abc_ObjFanout0(p.pResetLatch));
    // get the output net
    pOutNet = Abc_NtkFindOrCreateNet( p.pNtk, pName );
    // get the data nets
    pData0Net = Abc_NtkFindOrCreateNet( p.pNtk, Abc_ObjNameSuffix(pOutNet, "_reset") );
    pData1Net = Abc_NtkFindOrCreateNet( p.pNtk, Abc_ObjNameSuffix(pOutNet, "_out") );
    // duplicate MV variables
    if Abc_NtkMvVar(p.pNtk) {
        pVar = (Io_MvVar_t *)Abc_ObjMvVar( pOutNet );
        Abc_ObjSetMvVar( pData0Net, Abc_NtkMvVarDup(p.pNtk, pVar) );
        Abc_ObjSetMvVar( pData1Net, Abc_NtkMvVarDup(p.pNtk, pVar) );
    }
    // create the node
    pNode = Abc_NtkCreateNode( p.pNtk );
    // create the output net
    Abc_ObjAddFanin( pOutNet, pNode );
    // create the function
    if p.pMan.fBlifMv {
//        Vec_Att_t * p = Abc_NtkMvVar( pNtk );
//        int nValues = Abc_ObjMvVarNum(pOutNet);
//        sprintf( Buffer, "2 %d %d %d\n1 - - =1\n0 - - =2\n", nValues, nValues, nValues );
        Buffer = fmt.Sprintf("1 - - =1\n0 - - =2\n");
        pNode.pData = Abc_SopRegister( (Mem_Flex_t *)p.pNtk.pManFunc, Buffer );
    
    } else {
        pNode.pData = Abc_SopCreateMux( (Mem_Flex_t *)p.pNtk.pManFunc ); }
    // add nets
    Abc_ObjAddFanin( pNode, pResetLONet );
    Abc_ObjAddFanin( pNode, pData1Net );
    Abc_ObjAddFanin( pNode, pData0Net );
    return pData0Net;
}

/**Function*************************************************************

  Synopsis    [Parses the nodes line.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_MvParseLineNamesMvOne(p *Io_MvMod_t, vTokens *Vec_Ptr_t, vTokens2 *Vec_Ptr_t, nInputs int32, nOutputs int32, iOut int32, fReset int32) int32 {
    var pNet, pNode *Abc_Obj_t;
    var pName string;
    // get the output name
    pName = (char *)Vec_PtrEntry( vTokens, Vec_PtrSize(vTokens) - nOutputs + iOut );
    // create the node
    if fReset != 0 {
        pNet = Abc_NtkFindNet( p.pNtk, pName );
        if pNet == nil {
            sprintf( p.pMan.sError, "Line %d: Latch with output signal \"%s\" does not exist.", Io_MvGetLine(p.pMan, pName), pName );
            return 0;
        }
/*
        if ( !Abc_ObjIsBo(Abc_ObjFanin0(pNet)) )
        {
            sprintf( p->pMan->sError, "Line %d: Reset line \"%s\" defines signal that is not a latch output.", Io_MvGetLine(p->pMan, pName), pName );
            return 0;
        }
*/
        // construct the reset circuit and get the reset net feeding into it
        pNet = Io_MvParseAddResetCircuit( p, pName );
        // create fanins
        pNode = Io_ReadCreateNode( p.pNtk, Abc_ObjName(pNet), (char **)(vTokens.pArray + 1), nInputs );
        assert(nInputs == Vec_PtrSize(vTokens) - 2, "nInputs == Vec_PtrSize(vTokens) - 2");
    
    } else {
        pNet = Abc_NtkFindOrCreateNet( p.pNtk, pName );
        if Abc_ObjFaninNum(pNet) > 0 {
            sprintf( p.pMan.sError, "Line %d: Signal \"%s\" is defined more than once.", Io_MvGetLine(p.pMan, pName), pName );
            return 0;
        }
        pNode = Io_ReadCreateNode( p.pNtk, pName, (char **)(vTokens.pArray + 1), nInputs );
    }
    // create the cover
    pNode.pData = Io_MvParseTableMv( p, pNode, vTokens2, nInputs, nOutputs, iOut );
    if pNode.pData == nil {
        return 0; }
    pNode.pData = Abc_SopRegister( (Mem_Flex_t *)p.pNtk.pManFunc, (char *)pNode.pData );
//printf( "Finished parsing node \"%s\" with table:\n%s\n", pName, pNode->pData );
    return 1;
}

/**Function*************************************************************

  Synopsis    [Parses the nodes line.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_MvParseLineNamesMv(p *Io_MvMod_t, pLine string, fReset int32) int32 {
    var vTokens *Vec_Ptr_t  = p.pMan.vTokens;
    var vTokens2 *Vec_Ptr_t  = p.pMan.vTokens2;
    var pNet *Abc_Obj_t;
    var pName, pFirst, pArrow string;
    var nInputs, nOutputs, nLiterals, nLines, i int32;
    assert(p.pMan.fBlifMv, "p.pMan.fBlifMv");
    // get the arrow if it is present
    pArrow = Io_MvFindArrow( pLine );
    if !p.pMan.fBlifMv && pArrow {
        sprintf( p.pMan.sError, "Line %d: Multi-output node symbol (->) in binary BLIF file.", Io_MvGetLine(p.pMan, pLine) );
        return 0;
    }
    // split names line into tokens
    Io_MvSplitIntoTokens( vTokens, pLine, '\0' );
    if fReset != 0 {
        assert(!int32(strings.Compare((char *)Vec_PtrEntry(vTokens,0), "r")) || !int32(strings.Compare((char *)Vec_PtrEntry(vTokens,0), "reset")), "!strcmp((char *)Vec_PtrEntry(vTokens,0), \"r\") || !strcmp((char *)Vec_PtrEntry(vTokens,0), \"reset\")");
    } else {
        assert(!int32(strings.Compare((char *)Vec_PtrEntry(vTokens,0), "names")) || !int32(strings.Compare((char *)Vec_PtrEntry(vTokens,0), "table")), "!strcmp((char *)Vec_PtrEntry(vTokens,0), \"names\") || !strcmp((char *)Vec_PtrEntry(vTokens,0), \"table\")"); }
    // find the number of inputs and outputs
    nInputs  = Vec_PtrSize(vTokens) - 2;
    nOutputs = 1;
    if pArrow != nil {
        for  i = Vec_PtrSize(vTokens) - 2; i >= 1; i--  {
            if pArrow < (char*)Vec_PtrEntry(vTokens,i) {
                nInputs--;
                nOutputs++;
            }
        }
    }
    // split table into tokens
    pName = (char *)Vec_PtrEntryLast( vTokens );
    Io_MvSplitIntoTokensMv( vTokens2, pName + len(pName) );
    pFirst = (char *)Vec_PtrEntry( vTokens2, 0 );
    if pFirst[0] == '.' {
        assert(pFirst[1] == 'd', "pFirst[1] == 'd'");
        nLiterals = Vec_PtrSize(vTokens2) - 1 - nOutputs;
    
    } else {
        nLiterals = Vec_PtrSize(vTokens2); }
    // check the number of lines
    if nLiterals % (nInputs + nOutputs) != 0 {
        sprintf( p.pMan.sError, "Line %d: Wrong number of literals in the table of node \"%s\". (Spaces inside literals are not allowed.)", Io_MvGetLine(p.pMan, pFirst), pName );
        return 0;
    }
    // check for the ND table
    nLines = nLiterals / (nInputs + nOutputs);
    if nInputs == 0 && nLines > 1 {
        // add the outputs to the PIs
        for  i = 0; i < nOutputs; i++  {
            pName = (char *)Vec_PtrEntry( vTokens, Vec_PtrSize(vTokens) - nOutputs + i );
            // get the net corresponding to this node
            pNet = Abc_NtkFindOrCreateNet(p.pNtk, pName);
            if fReset != 0 {
                assert(p.pResetLatch != nil, "p.pResetLatch != nil");
                // construct the reset circuit and get the reset net feeding into it
                pNet = Io_MvParseAddResetCircuit( p, pName );
            }
            // add the new PI node
//            Abc_ObjAddFanin( pNet, Abc_NtkCreatePi(p->pNtk) );
//            fprintf( stdout, "Io_ReadBlifMv(): Adding PI for internal non-deterministic node \"%s\".\n", pName );
            p.pMan.nNDnodes++;
            Abc_ObjAddFanin( pNet, Abc_NtkCreateNodeConst0(p.pNtk) );
        }
        return 1;
    }
    // iterate through the outputs
    for  i = 0; i < nOutputs; i++  {
        if !(Io_MvParseLineNamesMvOne( p, vTokens, vTokens2, nInputs, nOutputs, i, fReset ) != 0) {
            return 0; }
    }
    return 1;
}


/**Function*************************************************************

  Synopsis    [Constructs the SOP cover from the file parsing info.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_MvParseTableBlif(p *Io_MvMod_t, pTable string, nFanins int32) string {
    var vTokens *Vec_Ptr_t  = p.pMan.vTokens;
    var vFunc *Vec_Str_t  = p.pMan.vFunc;
    var pProduct string;
    var pOutput string;
    var c byte;
    var i int32;
    var Polarity int32 = -1;

    p.pMan.nTablesRead++;
    // get the tokens
    Io_MvSplitIntoTokens( vTokens, pTable, '.' );
    if Vec_PtrSize(vTokens) == 0 {
        return Abc_SopCreateConst0( (Mem_Flex_t *)p.pNtk.pManFunc ); }
    if Vec_PtrSize(vTokens) == 1 {
        pOutput = (char *)Vec_PtrEntry( vTokens, 0 );
        c = pOutput[0];
        if (c!='0'&&c!='1'&&c!='x'&&c!='n') || pOutput[1] {
            sprintf( p.pMan.sError, "Line %d: Constant table has wrong output value \"%s\".", Io_MvGetLine(p.pMan, pOutput), pOutput );
            return nil;
        }
        if pOutput[0] == '0' {
            return Abc_SopCreateConst0((Mem_Flex_t *)p.pNtk.pManFunc);
        } else {
            return Abc_SopCreateConst1((Mem_Flex_t *)p.pNtk.pManFunc);
        }
    }
    pProduct = (char *)Vec_PtrEntry( vTokens, 0 );
    if Vec_PtrSize(vTokens) % 2 == 1 {
        sprintf( p.pMan.sError, "Line %d: Table has odd number of tokens (%d).", Io_MvGetLine(p.pMan, pProduct), Vec_PtrSize(vTokens) );
        return nil;
    }
    // parse the table
    Vec_StrClear( vFunc );
    for  i = 0; i < Vec_PtrSize(vTokens)/2; i++  {
        pProduct = (char *)Vec_PtrEntry( vTokens, 2*i + 0 );
        pOutput  = (char *)Vec_PtrEntry( vTokens, 2*i + 1 );
        if len(pProduct) != uint32(nFanins) {
            sprintf( p.pMan.sError, "Line %d: Cube \"%s\" has size different from the fanin count (%d).", Io_MvGetLine(p.pMan, pProduct), pProduct, nFanins );
            return nil;
        }
        c = pOutput[0];
        if (c!='0'&&c!='1'&&c!='x'&&c!='n') || pOutput[1] {
            sprintf( p.pMan.sError, "Line %d: Output value \"%s\" is incorrect.", Io_MvGetLine(p.pMan, pProduct), pOutput );
            return nil;
        }
        if Polarity == -1 {
            Polarity = (c=='1' || c=='x');
        } else if Polarity != (c=='1' || c=='x') {
            sprintf( p.pMan.sError, "Line %d: Output value \"%s\" differs from the value in the first line of the table (%d).", Io_MvGetLine(p.pMan, pProduct), pOutput, Polarity );
            return nil;
        }
        // parse one product 
        Vec_StrPrintStr( vFunc, pProduct );
        Vec_StrPush( vFunc, ' ' );
        Vec_StrPush( vFunc, pOutput[0] );
        Vec_StrPush( vFunc, '\n' );
    }
    Vec_StrPush( vFunc, '\0' );
    return Vec_StrArray( vFunc );
}

/**Function*************************************************************

  Synopsis    [Parses the nodes line.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_MvParseLineNamesBlif(p *Io_MvMod_t, pLine string) int32 {
    var vTokens *Vec_Ptr_t  = p.pMan.vTokens;
    var pNet, pNode *Abc_Obj_t;
    var pName string;
    assert(!p.pMan.fBlifMv, "!p.pMan.fBlifMv");
    Io_MvSplitIntoTokens( vTokens, pLine, '\0' );
    // parse the mapped node
    if !int32(strings.Compare((char *)Vec_PtrEntry(vTokens,0), "gate")) {
        return Io_MvParseLineGateBlif( p, vTokens ); }
    // parse the regular name line
    assert(!int32(strings.Compare((char *)Vec_PtrEntry(vTokens,0), "names")), "!strcmp((char *)Vec_PtrEntry(vTokens,0), \"names\")");
    pName = (char *)Vec_PtrEntryLast( vTokens );
    pNet = Abc_NtkFindOrCreateNet( p.pNtk, pName );
    if Abc_ObjFaninNum(pNet) > 0 {
        sprintf( p.pMan.sError, "Line %d: Signal \"%s\" is defined more than once.", Io_MvGetLine(p.pMan, pName), pName );
        return 0;
    }
    // create fanins
    pNode = Io_ReadCreateNode( p.pNtk, pName, (char **)(vTokens.pArray + 1), Vec_PtrSize(vTokens) - 2 );
    // parse the table of this node
    pNode.pData = Io_MvParseTableBlif( p, pName + len(pName), Abc_ObjFaninNum(pNode) );
    if pNode.pData == nil {
        return 0; }
    pNode.pData = Abc_SopRegister( (Mem_Flex_t *)p.pNtk.pManFunc, (char *)pNode.pData );
    return 1;
}

//ABC_NAMESPACE_IMPL_END

//#include "map/mio/mio.h"
//#include "base/main/main.h"

//ABC_NAMESPACE_IMPL_START


/**Function*************************************************************

  Synopsis    [Parses the nodes line.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_MvParseLineShortBlif(p *Io_MvMod_t, pLine string) int32 {
    var vTokens *Vec_Ptr_t  = p.pMan.vTokens;
    var pNet, pNode *Abc_Obj_t;
    var pName string;
    assert(!p.pMan.fBlifMv, "!p.pMan.fBlifMv");
    Io_MvSplitIntoTokens( vTokens, pLine, '\0' );
    if Vec_PtrSize(vTokens) != 3 {
        sprintf( p.pMan.sError, "Line %d: Expecting three entries in the .short line.", Io_MvGetLine(p.pMan, (char *)Vec_PtrEntry(vTokens,0)) );
        return 0;
    }
    // parse the regular name line
    assert(!int32(strings.Compare((char *)Vec_PtrEntry(vTokens,0), "short")), "!strcmp((char *)Vec_PtrEntry(vTokens,0), \"short\")");
    pName = (char *)Vec_PtrEntryLast( vTokens );
    pNet = Abc_NtkFindOrCreateNet( p.pNtk, pName );
    if Abc_ObjFaninNum(pNet) > 0 {
        sprintf( p.pMan.sError, "Line %d: Signal \"%s\" is defined more than once.", Io_MvGetLine(p.pMan, pName), pName );
        return 0;
    }
    // create fanins
    pNode = Io_ReadCreateNode( p.pNtk, pName, (char **)(vTokens.pArray + 1), 1 );
    // parse the table of this node
    if p.pNtk.ntkFunc == ABC_FUNC_MAP {
        var pGenlib *Mio_Library_t; 
        var pGate *Mio_Gate_t;
        // check that the library is available
        pGenlib = (Mio_Library_t *)Abc_FrameReadLibGen();
        if pGenlib == nil {
            sprintf( p.pMan.sError, "Line %d: The current library is not available.", Io_MvGetLine(p.pMan, pName) );
            return 0;
        }
        // get the gate
        pGate = Mio_LibraryReadBuf( pGenlib );
        if pGate == nil {
            sprintf( p.pMan.sError, "Line %d: Cannot find buffer gate in the library.", Io_MvGetLine(p.pMan, pName) );
            return 0;
        }
        Abc_ObjSetData( pNode, pGate );
    
    } else {
        pNode.pData = Abc_SopRegister( (Mem_Flex_t *)p.pNtk.pManFunc, "1 1\n" ); }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Duplicate the MV variable.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkMvVarDup(pNtk *Abc_Ntk_t, pVar *Io_MvVar_t) *Io_MvVar_t {
    var pFlex *Mem_Flex_t;
    var pVarDup *Io_MvVar_t;
    var i int32;
    if pVar == nil {
        return nil; }
    pFlex = (Mem_Flex_t *)Abc_NtkMvVarMan( pNtk );
    assert(pFlex != nil, "pFlex != nil");
    pVarDup = (Io_MvVar_t *)Mem_FlexEntryFetch( pFlex, sizeof(Io_MvVar_t) );
    pVarDup.nValues = pVar.nValues;
    pVarDup.pNames = nil;
    if pVar.pNames == nil {
        return pVarDup; }
    pVarDup.pNames = (char **)Mem_FlexEntryFetch( pFlex, sizeof(char *) * pVar.nValues );
    for  i = 0; i < pVar.nValues; i++  {
        pVarDup.pNames[i] = (char *)Mem_FlexEntryFetch( pFlex, len(pVar.pNames[i]) + 1 );
        strcpy( pVarDup.pNames[i], pVar.pNames[i] );
    }
    return pVarDup;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_ReadBlifCleanName(pName string) string {
    var i, Length int32;
    Length = len(pName);
    for  i = 0; i < Length; i++  {
        if pName[i] == '=' {
            return pName + i + 1; }
    }
    return nil;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_MvParseLineGateBlif(p *Io_MvMod_t, vTokens *Vec_Ptr_t) int32 {
    extern int Io_ReadBlifReorderFormalNames( Vec_Ptr_t * vTokens, Mio_Gate_t * pGate, Mio_Gate_t * pTwin );
    var pGenlib *Mio_Library_t; 
    var pGate *Mio_Gate_t;
    var pNode *Abc_Obj_t;
    var ppNames *string;
    var pName string;
    var i, nNames int32;

    pName = (char *)vTokens.pArray[0];

    // check that the library is available
    pGenlib = (Mio_Library_t *)Abc_FrameReadLibGen();
    if pGenlib == nil {
        sprintf( p.pMan.sError, "Line %d: The current library is not available.", Io_MvGetLine(p.pMan, pName) );
        return 0;
    }

    // create a new node and add it to the network
    if vTokens.nSize < 2 {
        sprintf( p.pMan.sError, "Line %d: The .gate line has less than two tokens.", Io_MvGetLine(p.pMan, pName) );
        return 0;
    }

    // get the gate
    pGate = Mio_LibraryReadGateByName( pGenlib, (char *)vTokens.pArray[1], nil );
    if pGate == nil {
        sprintf( p.pMan.sError, "Line %d: Cannot find gate \"%s\" in the library.", Io_MvGetLine(p.pMan, pName), (char*)vTokens.pArray[1] );
        return 0;
    }

    // if this is the first line with gate, update the network type
    if Abc_NtkNodeNum(p.pNtk) == 0 && p.pNtk.ntkFunc == ABC_FUNC_SOP {
        assert(p.pNtk.ntkFunc == ABC_FUNC_SOP, "p.pNtk.ntkFunc == ABC_FUNC_SOP");
        p.pNtk.ntkFunc = ABC_FUNC_MAP;
        Mem_FlexStop( (Mem_Flex_t *)p.pNtk.pManFunc, 0 );
        p.pNtk.pManFunc = pGenlib;
        if p.pMan && p.pMan.pDesign && Vec_PtrSize(p.pMan.pDesign.vModules) > 0 {
            var pModel *Abc_Ntk_t; var k int32;
            // inlining of macro Vec_PtrForEachEntry
            for  k = 0; (k < Vec_PtrSize(p.pMan.pDesign.vModules)) && (((pModel) = (Abc_Ntk_t *)Vec_PtrEntry(p.pMan.pDesign.vModules, k)), 1); k++  {
                if pModel == p.pNtk {
                    continue; }
                assert(pModel.ntkFunc == ABC_FUNC_SOP, "pModel.ntkFunc == ABC_FUNC_SOP");
                pModel.ntkFunc = ABC_FUNC_MAP;
                Mem_FlexStop( (Mem_Flex_t *)pModel.pManFunc, 0 );
                pModel.pManFunc = pGenlib;
            }
        }
    }

    // reorder the formal inputs to be in the same order as in the gate
    if !Io_ReadBlifReorderFormalNames( vTokens, pGate, Mio_GateReadTwin(pGate) ) {
        sprintf( p.pMan.sError, "Line %d: Mismatch in the fanins of gate \"%s\".", Io_MvGetLine(p.pMan, pName), (char*)vTokens.pArray[1] );
        return 0;
    }

    // remove the formal parameter names
    for  i = 2; i < vTokens.nSize; i++  {
        if vTokens.pArray[i] == nil {
            continue; }
        vTokens.pArray[i] = Io_ReadBlifCleanName( (char *)vTokens.pArray[i] );
        if vTokens.pArray[i] == nil {
            sprintf( p.pMan.sError, "Line %d: Invalid gate input assignment.", Io_MvGetLine(p.pMan, pName) );
            return 0;
        }
    }

    // create the node
    if Mio_GateReadTwin(pGate) == nil {
        nNames  = vTokens.nSize - 3;
        ppNames = (char **)vTokens.pArray + 2;
        pNode   = Io_ReadCreateNode( p.pNtk, ppNames[nNames], ppNames, nNames );
        Abc_ObjSetData( pNode, pGate );
    
    } else {
        nNames  = vTokens.nSize - 4;
        ppNames = (char **)vTokens.pArray + 2;
        assert(ppNames[nNames] != nil || ppNames[nNames+1] != nil, "ppNames[nNames] != nil || ppNames[nNames+1] != nil");
        if ppNames[nNames] {
            pNode   = Io_ReadCreateNode( p.pNtk, ppNames[nNames], ppNames, nNames );
            Abc_ObjSetData( pNode, pGate );
        }
        if ppNames[nNames+1] {
            pNode   = Io_ReadCreateNode( p.pNtk, ppNames[nNames+1], ppNames, nNames );
            Abc_ObjSetData( pNode, Mio_GateReadTwin(pGate) );
        }
    }

    return 1;
}

/**Function*************************************************************

  Synopsis    [Box mapping procedures.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_MapBoxSetPrevNext(vDrivers *Vec_Ptr_t, vMapIn *Vec_Int_t, vMapOut *Vec_Int_t, Id int32) {
    var pNode *Abc_Obj_t;
    pNode = (Abc_Obj_t *)Vec_PtrEntry(vDrivers, Id+2);
    Vec_IntWriteEntry( vMapIn, Abc_ObjId(Abc_ObjFanin0(Abc_ObjFanin0(pNode))), Id );
    pNode = (Abc_Obj_t *)Vec_PtrEntry(vDrivers, Id+4);
    Vec_IntWriteEntry( vMapOut, Abc_ObjId(Abc_ObjFanin0(Abc_ObjFanin0(pNode))), Id );
}
func Abc_MapBox2Next(vDrivers *Vec_Ptr_t, vMapIn *Vec_Int_t, vMapOut *Vec_Int_t, Id int32) int32 {
    var pNode *Abc_Obj_t  = (Abc_Obj_t *)Vec_PtrEntry(vDrivers, Id+4);
    return Vec_IntEntry( vMapIn, Abc_ObjId(Abc_ObjFanin0(Abc_ObjFanin0(pNode))) );
}
func Abc_MapBox2Prev(vDrivers *Vec_Ptr_t, vMapIn *Vec_Int_t, vMapOut *Vec_Int_t, Id int32) int32 {
    var pNode *Abc_Obj_t  = (Abc_Obj_t *)Vec_PtrEntry(vDrivers, Id+2);
    return Vec_IntEntry( vMapOut, Abc_ObjId(Abc_ObjFanin0(Abc_ObjFanin0(pNode))) );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

