package io

import (
	"fmt"
	"strconv"
	"strings"
)

import "os"

/**CFile****************************************************************

  FileName    [ioReadBlifAig.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Command processing package.]

  Synopsis    [Procedures to read BLIF file into AIG.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - December 23, 2006.]

  Revision    [$Id: ioReadBlifAig.c,v 1.00 2006/12/23 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "base/abc/abc.h"
//#include "misc/vec/vecPtr.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// latch initial values
 type Io_BlifInit_t int32
const ( 
    IO_BLIF_INIT_NONE Io_BlifInit_t = 0
  // 0:  unknown
    IO_BLIF_INIT_ZERO Io_BlifInit_t = IO_BLIF_INIT_NONE + 1
      // 1:  zero
    IO_BLIF_INIT_ONE Io_BlifInit_t = IO_BLIF_INIT_ZERO + 1
       // 2:  one
    IO_BLIF_INIT_DC Io_BlifInit_t = IO_BLIF_INIT_ONE + 1         // 3:  don't-care

)

 

     // parsing object
type Io_BlifObj_t struct {
    fPi uint32; /* C bitfield: 1 bits; review packing and truncation */  // the object is a primary input
    fPo uint32; /* C bitfield: 1 bits; review packing and truncation */  // the object is a primary output
    fLi uint32; /* C bitfield: 1 bits; review packing and truncation */  // the object is a latch input
    fLo uint32; /* C bitfield: 1 bits; review packing and truncation */  // the object is a latch output
    fDef uint32; /* C bitfield: 1 bits; review packing and truncation */  // the object is defined as a table (node, PO, LI)
    fLoop uint32; /* C bitfield: 1 bits; review packing and truncation */  // flag for loop detection
    Init uint32; /* C bitfield: 2 bits; review packing and truncation */  // the latch initial state
    Offset uint32; /* C bitfield: 24 bits; review packing and truncation */  // temporary number
    pName                string;        // the name of this object
    pEquiv                any;       // the AIG node representing this line    
    pNext        *Io_BlifObj_t;        // the next obj in the hash table
}

     // parsing manager
type Io_BlifMan_t struct {
    // general info about file
    pFileName                string;    // the name of the file
    pBuffer                string;      // the begining of the file buffer
    vLines           *Vec_Ptr_t;       // the line beginnings
    // temporary objects
    pObjects        *Io_BlifObj_t;     // the storage for objects
    nObjects                  int32;     // the number of objects allocated
    iObjNext                  int32;     // the next free object
    // file lines
    pModel                string;       // .model line
    vInputs           *Vec_Ptr_t;      // .inputs lines
    vOutputs           *Vec_Ptr_t;     // .outputs lines
    vLatches           *Vec_Ptr_t;     // .latches lines
    vNames           *Vec_Ptr_t;       // .names lines
    // network objects
    vPis           *Vec_Ptr_t;         // the PI structures
    vPos           *Vec_Ptr_t;         // the PO structures
    vLis           *Vec_Ptr_t;         // the LI structures
    vLos           *Vec_Ptr_t;         // the LO structures
    // mapping of names into objects
    pTable       **Io_BlifObj_t;       // the hash table
    nTableSize                  int32;   // the hash table size
    // current processing info
    pAig           *Abc_Ntk_t;         // the network under construction
    vTokens           *Vec_Ptr_t;      // the current tokens
    sError [512]byte;  // the error string generated during parsing
    // statistics 
    nTablesRead                  int32;  // the number of processed tables
    nTablesLeft                  int32;  // the number of dangling tables
}

// static functions
// static Io_BlifMan_t *    Io_BlifAlloc();

// static void              Io_BlifFree( Io_BlifMan_t * p );

// static char *            Io_BlifLoadFile( char * pFileName );

// static void              Io_BlifReadPreparse( Io_BlifMan_t * p );

// static Abc_Ntk_t *       Io_BlifParse( Io_BlifMan_t * p );

// static int               Io_BlifParseModel( Io_BlifMan_t * p, char * pLine );

// static int               Io_BlifParseInputs( Io_BlifMan_t * p, char * pLine );

// static int               Io_BlifParseOutputs( Io_BlifMan_t * p, char * pLine );

// static int               Io_BlifParseLatch( Io_BlifMan_t * p, char * pLine );

// static int               Io_BlifParseNames( Io_BlifMan_t * p, char * pLine );

// static int               Io_BlifParseConstruct( Io_BlifMan_t * p );

func Io_BlifCharIsSpace(s byte) int32 { return s == ' ' || s == '\t' || s == '\r' || s == '\n';  }

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Reads the network from the BLIF file as an AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_ReadBlifAsAig(pFileName string, fCheck int32) *Abc_Ntk_t {
    var pFile *os.File;
    var p *Io_BlifMan_t;
    var pAig *Abc_Ntk_t;

    // check that the file is available
    pFile = os.Open(pFileName);
    if pFile == nil {
        fmt.Printf("Io_Blif(): The file is unavailable (absent or open).\n");
        return 0;
    }
    _ = (pFile).Close();

    // start the file reader
    p = Io_BlifAlloc();
    p.pFileName = pFileName;
    p.pBuffer   = Io_BlifLoadFile( pFileName );
    if p.pBuffer == nil {
        Io_BlifFree( p );
        return nil;
    }
    // prepare the file for parsing
    Io_BlifReadPreparse( p );
    // construct the network
    pAig = Io_BlifParse( p );
    if p.sError[0] {
        fmt.Fprintf(os.Stdout, "%s\n", p.sError); }
    if pAig == nil {
        return nil; }
    Io_BlifFree( p );

    // make sure that everything is okay with the network structure
    if (fCheck != 0) && !Abc_NtkCheckRead( pAig ) {
        fmt.Printf("Io_Blif: The network check has failed.\n");
        Abc_NtkDelete( pAig );
        return nil;
    }
    return pAig;
}

/**Function*************************************************************

  Synopsis    [Allocates the BLIF parsing structure.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_BlifAlloc() *Io_BlifMan_t {
    var p *Io_BlifMan_t;
    // inlining of macro ABC_ALLOC
    p = ((Io_BlifMan_t *) malloc(sizeof(Io_BlifMan_t) * (size_t)(1)));
    *p = Io_BlifMan_t{};
    p.vLines   = Vec_PtrAlloc( 512 );
    p.vInputs  = Vec_PtrAlloc( 512 );
    p.vOutputs = Vec_PtrAlloc( 512 );
    p.vLatches = Vec_PtrAlloc( 512 );
    p.vNames   = Vec_PtrAlloc( 512 );
    p.vTokens  = Vec_PtrAlloc( 512 );
    p.vPis     = Vec_PtrAlloc( 512 );
    p.vPos     = Vec_PtrAlloc( 512 );
    p.vLis     = Vec_PtrAlloc( 512 );
    p.vLos     = Vec_PtrAlloc( 512 );
    return p;
}

/**Function*************************************************************

  Synopsis    [Frees the BLIF parsing structure.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_BlifFree(p *Io_BlifMan_t) {
    if p.pAig {
        Abc_NtkDelete( p.pAig ); }
    if p.pBuffer {  p.pBuffer = nil; }
    if p.pObjects { p.pObjects = nil; }
    if p.pTable {   p.pTable = nil; }
    Vec_PtrFree( p.vLines );
    Vec_PtrFree( p.vInputs );
    Vec_PtrFree( p.vOutputs );
    Vec_PtrFree( p.vLatches );
    Vec_PtrFree( p.vNames );
    Vec_PtrFree( p.vTokens );
    Vec_PtrFree( p.vPis );
    Vec_PtrFree( p.vPos );
    Vec_PtrFree( p.vLis );
    Vec_PtrFree( p.vLos );
    p = nil;
}


/**Function*************************************************************

  Synopsis    [Hashing for character strings.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Io_BlifHashString_s_Primes [10]int32 = [10]int32{ 
        1291, 1699, 2357, 4177, 5147, 
        5647, 6343, 7103, 7873, 8147,
    };
func Io_BlifHashString(pName string, TableSize int32) uint32 {
    
    var i, Key uint32;
    for  i = 0; pName[i] != '\0'; i++  {
        Key ^= spStatic_Io_BlifHashString_s_Primes[i%10]*pName[i]*pName[i];
    }
    return Key % TableSize;
}

/**Function*************************************************************

  Synopsis    [Checks if the given name exists in the table.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_BlifHashLookup(p *Io_BlifMan_t, pName string) **Io_BlifObj_t {
    var ppEntry **Io_BlifObj_t;
    for  ppEntry = p.pTable + Io_BlifHashString(pName, p.nTableSize); *ppEntry; ppEntry = &(*ppEntry).pNext  {
        if !(int32(strings.Compare((*ppEntry).pName, pName)) != 0) {
            return ppEntry; }
    }
    return ppEntry;
}

/**Function*************************************************************

  Synopsis    [Finds or add the given name to the table.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_BlifHashFindOrAdd(p *Io_BlifMan_t, pName string) *Io_BlifObj_t {
    var ppEntry **Io_BlifObj_t;
    ppEntry = Io_BlifHashLookup( p, pName );
    if *ppEntry == nil {
        assert(p.iObjNext < p.nObjects, "p.iObjNext < p.nObjects"); 
        *ppEntry = p.pObjects + p.iObjNext++;
        (*ppEntry).pName = pName;
    }
    return *ppEntry;
}


/**Function*************************************************************

  Synopsis    [Collects the already split tokens.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_BlifCollectTokens(vTokens *Vec_Ptr_t, pInput string, pOutput string) {
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
func Io_BlifSplitIntoTokens(vTokens *Vec_Ptr_t, pLine string, Stop byte) {
    var pCur string;
    // clear spaces
    for  pCur = pLine; *pCur != Stop; pCur++  {
        if Io_BlifCharIsSpace(*pCur) != 0 {
            *pCur = 0; }
    }
    // collect tokens
    Io_BlifCollectTokens( vTokens, pLine, pCur );
}

/**Function*************************************************************

  Synopsis    [Returns the 1-based number of the line in which the token occurs.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_BlifGetLine(p *Io_BlifMan_t, pToken string) int32 {
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

  Synopsis    [Conservatively estimates the number of primary inputs.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_BlifEstimatePiNum(p *Io_BlifMan_t) int32 {
    var pCur string;
    var i, fSpaces int32;
    var Counter int32;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vInputs)) && (((pCur) = (char *)Vec_PtrEntry(p.vInputs, i)), 1); i++  {
        for  fSpaces = 0; *pCur; pCur++  {
            if Io_BlifCharIsSpace(*pCur) != 0 {
                if !(fSpaces != 0) {
                    Counter++; }
                fSpaces = 1;
            
            } else {
                fSpaces = 0; }
        }
    }
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Conservatively estimates the number of AIG nodes.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_BlifEstimateAndNum(p *Io_BlifMan_t) int32 {
    var pObj *Io_BlifObj_t;
    var pCur string;
    var i, CounterOne, Counter int32;
    for  i = 0; i < p.iObjNext; i++  {
        pObj = p.pObjects + i;
        if !pObj.fDef {
            continue; }
        CounterOne = 0;
        for  pCur = pObj.pName + len(pObj.pName); *pCur != '.'; pCur++  {
            if *pCur == '0' || *pCur == '1' {
                CounterOne++; }
        }
        if CounterOne {
            Counter += CounterOne - 1; }
    }
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Reads the file into a character buffer.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_BlifLoadFile(pFileName string) string {
    var pFile *os.File;
    var nFileSize int32;
    var pContents string;
    var RetValue int32;
    pFile = os.Open(pFileName);
    if pFile == nil {
        fmt.Printf("Io_BlifLoadFile(): The file is unavailable (absent or open).\n");
        return nil;
    }
    fseek( pFile, 0, SEEK_END );  
    nFileSize = ftell( pFile ); 
    if nFileSize == 0 {
        _ = (pFile).Close();
        fmt.Printf("Io_BlifLoadFile(): The file is empty.\n");
        return nil;
    }
    // inlining of macro ABC_ALLOC
    pContents = ((char *) malloc(sizeof(char) * (size_t)(nFileSize + 10)));
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
func Io_BlifReadPreparse(p *Io_BlifMan_t) {
    var pCur, pPrev string;
    var i, fComment int32;
    // parse the buffer into lines and remove comments
    Vec_PtrPush( p.vLines, p.pBuffer );
    for  pCur = p.pBuffer; *pCur; pCur++  {
        if *pCur == '\n' {
            *pCur = 0;
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
            if !(Io_BlifCharIsSpace(*pPrev) != 0) {
                break; }
        }
        // if it is the line extender, overwrite it with spaces
        if *pPrev == '\\' {
            for  ; *pPrev; pPrev++  {
                *pPrev = ' ';
            }
            *pPrev = ' ';
            continue;
        }
        // skip spaces at the beginning of the line
        for  Io_BlifCharIsSpace(*pCur++)  {; }
        // parse directives
        if *(pCur-1) != '.' {
            continue; }
        if !strncmp(pCur, "names", 5) {
            Vec_PtrPush( p.vNames, pCur );
        } else if !strncmp(pCur, "latch", 5) {
            Vec_PtrPush( p.vLatches, pCur );
        } else if !strncmp(pCur, "inputs", 6) {
            Vec_PtrPush( p.vInputs, pCur );
        } else if !strncmp(pCur, "outputs", 7) {
            Vec_PtrPush( p.vOutputs, pCur );
        } else if !strncmp(pCur, "model", 5) { 
            p.pModel = pCur;
        } else if !strncmp(pCur, "end", 3) || !strncmp(pCur, "exdc", 4) {
            break;
        } else {
            pCur--;
            if pCur[len(pCur)-1] == '\r' {
                pCur[len(pCur)-1] = 0; }
            fmt.Fprintf(os.Stdout, "Line %d: Skipping line \"%s\".\n", Io_BlifGetLine(p, pCur), pCur);
        }
    }

    // count the number of objects
    p.nObjects = Io_BlifEstimatePiNum(p) + Vec_PtrSize(p.vLatches) + Vec_PtrSize(p.vNames) + 512;

    // allocate memory for objects
    // inlining of macro ABC_ALLOC
    p.pObjects = ((Io_BlifObj_t *) malloc(sizeof(Io_BlifObj_t) * (size_t)(p.nObjects)));
    memset( p.pObjects, 0, p.nObjects * sizeof(Io_BlifObj_t) );

    // allocate memory for the hash table
    p.nTableSize = p.nObjects/2 + 1;
    // inlining of macro ABC_ALLOC
    p.pTable = ((Io_BlifObj_t * *) malloc(sizeof(Io_BlifObj_t *) * (size_t)(p.nTableSize)));
    memset( p.pTable, 0, p.nTableSize * sizeof(Io_BlifObj_t *) );
}


/**Function*************************************************************

  Synopsis    [Reads the AIG in the binary AIGER format.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_BlifParse(p *Io_BlifMan_t) *Abc_Ntk_t {
    var pAig *Abc_Ntk_t;
    var pLine string;
    var i int32;
    // parse the model
    if !(Io_BlifParseModel( p, p.pModel ) != 0) {
        return nil; }
    // parse the inputs
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vInputs)) && (((pLine) = (char *)Vec_PtrEntry(p.vInputs, i)), 1); i++  {
        if !(Io_BlifParseInputs( p, pLine ) != 0) {
            return nil; }
    }
    // parse the outputs
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vOutputs)) && (((pLine) = (char *)Vec_PtrEntry(p.vOutputs, i)), 1); i++  {
        if !(Io_BlifParseOutputs( p, pLine ) != 0) {
            return nil; }
    }
    // parse the latches
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vLatches)) && (((pLine) = (char *)Vec_PtrEntry(p.vLatches, i)), 1); i++  {
        if !(Io_BlifParseLatch( p, pLine ) != 0) {
            return nil; }
    }
    // parse the nodes
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vNames)) && (((pLine) = (char *)Vec_PtrEntry(p.vNames, i)), 1); i++  {
        if !(Io_BlifParseNames( p, pLine ) != 0) {
            return nil; }
    }
    // reconstruct the network from the parsed data
    if !(Io_BlifParseConstruct( p ) != 0) {
        return nil; }
    // return the network
    pAig = p.pAig;
    p.pAig = nil;
    return pAig;
}

/**Function*************************************************************

  Synopsis    [Parses the model line.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_BlifParseModel(p *Io_BlifMan_t, pLine string) int32 {
    var pToken string;
    Io_BlifSplitIntoTokens( p.vTokens, pLine, '\0' );
    pToken = (char *)Vec_PtrEntry( p.vTokens, 0 );
    assert(!(int32(strings.Compare(pToken, "model")) != 0), "!strcmp(pToken, \"model\")");
    if Vec_PtrSize(p.vTokens) != 2 {
        sprintf( p.sError, "Line %d: Model line has %d entries while it should have 2.", Io_BlifGetLine(p, pToken), Vec_PtrSize(p.vTokens) );
        return 0;
    }
    p.pModel = (char *)Vec_PtrEntry( p.vTokens, 1 );
    return 1;
}

/**Function*************************************************************

  Synopsis    [Parses the inputs line.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_BlifParseInputs(p *Io_BlifMan_t, pLine string) int32 {
    var pObj *Io_BlifObj_t;
    var pToken string;
    var i int32;
    Io_BlifSplitIntoTokens( p.vTokens, pLine, '\0' );
    pToken = (char *)Vec_PtrEntry(p.vTokens, 0);
    assert(!(int32(strings.Compare(pToken, "inputs")) != 0), "!strcmp(pToken, \"inputs\")");
    // inlining of macro Vec_PtrForEachEntryStart
    for  i = 1; (i < Vec_PtrSize(p.vTokens)) && (((pToken) = (char *)Vec_PtrEntry(p.vTokens, i)), 1); i++  {
        pObj = Io_BlifHashFindOrAdd( p, pToken );
        if pObj.fPi {
            sprintf( p.sError, "Line %d: Primary input (%s) is defined more than once.", Io_BlifGetLine(p, pToken), pToken );
            return 0;
        }
        pObj.fPi = 1;
        Vec_PtrPush( p.vPis, pObj );
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Parses the outputs line.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_BlifParseOutputs(p *Io_BlifMan_t, pLine string) int32 {
    var pObj *Io_BlifObj_t;
    var pToken string;
    var i int32;
    Io_BlifSplitIntoTokens( p.vTokens, pLine, '\0' );
    pToken = (char *)Vec_PtrEntry(p.vTokens, 0);
    assert(!(int32(strings.Compare(pToken, "outputs")) != 0), "!strcmp(pToken, \"outputs\")");
    // inlining of macro Vec_PtrForEachEntryStart
    for  i = 1; (i < Vec_PtrSize(p.vTokens)) && (((pToken) = (char *)Vec_PtrEntry(p.vTokens, i)), 1); i++  {
        pObj = Io_BlifHashFindOrAdd( p, pToken );
        if pObj.fPo {
            fmt.Fprintf(os.Stdout, "Line %d: Primary output (%s) is defined more than once (warning only).\n", Io_BlifGetLine(p, pToken), pToken); }
        pObj.fPo = 1;
        Vec_PtrPush( p.vPos, pObj );
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Parses the latches line.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_BlifParseLatch(p *Io_BlifMan_t, pLine string) int32 {
    var pObj *Io_BlifObj_t;
    var pToken string;
    var Init int32;
    Io_BlifSplitIntoTokens( p.vTokens, pLine, '\0' );
    pToken = (char *)Vec_PtrEntry(p.vTokens,0);
    assert(!(int32(strings.Compare(pToken, "latch")) != 0), "!strcmp(pToken, \"latch\")");
    if Vec_PtrSize(p.vTokens) < 3 {
        sprintf( p.sError, "Line %d: Latch does not have input name and output name.", Io_BlifGetLine(p, pToken) );
        return 0;
    }
    // get initial value
    if Vec_PtrSize(p.vTokens) > 3 {
        Init = strconv.Atoi((char *)Vec_PtrEntry(p.vTokens,3));
    } else {
        Init = 2; }
    if Init < 0 || Init > 2 {
        sprintf( p.sError, "Line %d: Initial state of the latch is incorrect (%s).", Io_BlifGetLine(p, pToken), (char*)Vec_PtrEntry(p.vTokens,3) );
        return 0;
    }
    if Init == 0 {
        Init = IO_BLIF_INIT_ZERO;
    } else if Init == 1 {
        Init = IO_BLIF_INIT_ONE;
    } else { // if ( Init == 2 )
        Init = IO_BLIF_INIT_DC; }
    // get latch input
    pObj = Io_BlifHashFindOrAdd( p, (char *)Vec_PtrEntry(p.vTokens,1) );
    pObj.fLi = 1;
    Vec_PtrPush( p.vLis, pObj );
    pObj.Init = Init;
    // get latch output
    pObj = Io_BlifHashFindOrAdd( p, (char *)Vec_PtrEntry(p.vTokens,2) );
    if pObj.fPi {
        sprintf( p.sError, "Line %d: Primary input (%s) is also defined latch output.", Io_BlifGetLine(p, pToken), (char*)Vec_PtrEntry(p.vTokens,2) );
        return 0;
    }
    if pObj.fLo {
        sprintf( p.sError, "Line %d: Latch output (%s) is defined as the output of another latch.", Io_BlifGetLine(p, pToken), (char*)Vec_PtrEntry(p.vTokens,2) );
        return 0;
    }
    pObj.fLo = 1;
    Vec_PtrPush( p.vLos, pObj );
    pObj.Init = Init;
    return 1;
}

/**Function*************************************************************

  Synopsis    [Parses the nodes line.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_BlifParseNames(p *Io_BlifMan_t, pLine string) int32 {
    var pObj *Io_BlifObj_t;
    var pName string;
    Io_BlifSplitIntoTokens( p.vTokens, pLine, '\0' );
    assert(!int32(strings.Compare((char *)Vec_PtrEntry(p.vTokens,0), "names")), "!strcmp((char *)Vec_PtrEntry(p.vTokens,0), \"names\")");
    pName = (char *)Vec_PtrEntryLast( p.vTokens );
    pObj = Io_BlifHashFindOrAdd( p, pName );
    if pObj.fPi {
        sprintf( p.sError, "Line %d: Primary input (%s) has a table.", Io_BlifGetLine(p, pName), pName );
        return 0;
    }
    if pObj.fLo {
        sprintf( p.sError, "Line %d: Latch output (%s) has a table.", Io_BlifGetLine(p, pName), pName );
        return 0;
    }
    if pObj.fDef {
        sprintf( p.sError, "Line %d: Signal (%s) is defined more than once.", Io_BlifGetLine(p, pName), pName );
        return 0;
    }
    pObj.fDef = 1;
    // remember offset to the first fanin name
    pObj.pName = pName;
    pObj.Offset = pObj.pName - (char *)Vec_PtrEntry(p.vTokens,1);
    return 1;
}


/**Function*************************************************************

  Synopsis    [Constructs the AIG from the file parsing info.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_BlifParseTable(p *Io_BlifMan_t, pTable string, vFanins *Vec_Ptr_t) *Abc_Obj_t {
    var pProduct, pOutput string;
    var pRes, pCube *Abc_Obj_t;
    var i int32;
    var k int32;
    var Polarity int32 = -1;

    p.nTablesRead++;
    // get the tokens
    Io_BlifSplitIntoTokens( p.vTokens, pTable, '.' );
    if Vec_PtrSize(p.vTokens) == 0 {
        return Abc_ObjNot( Abc_AigConst1(p.pAig) ); }
    if Vec_PtrSize(p.vTokens) == 1 {
        pOutput = (char *)Vec_PtrEntry( p.vTokens, 0 );
        if ((pOutput[0] - '0') & 0x8E) || pOutput[1] {
            sprintf( p.sError, "Line %d: Constant table has wrong output value (%s).", Io_BlifGetLine(p, pOutput), pOutput );
            return nil;
        }
        return Abc_ObjNotCond( Abc_AigConst1(p.pAig), pOutput[0] == '0' );
    }
    pProduct = (char *)Vec_PtrEntry( p.vTokens, 0 );
    if Vec_PtrSize(p.vTokens) % 2 == 1 {
        sprintf( p.sError, "Line %d: Table has odd number of tokens (%d).", Io_BlifGetLine(p, pProduct), Vec_PtrSize(p.vTokens) );
        return nil;
    }
    // parse the table
    pRes = Abc_ObjNot( Abc_AigConst1(p.pAig) );
    for  i = 0; i < Vec_PtrSize(p.vTokens)/2; i++  {
        pProduct = (char *)Vec_PtrEntry( p.vTokens, 2*i + 0 );
        pOutput  = (char *)Vec_PtrEntry( p.vTokens, 2*i + 1 );
        if len(pProduct) != uint32(Vec_PtrSize(vFanins)) {
            sprintf( p.sError, "Line %d: Cube (%s) has size different from the fanin count (%d).", Io_BlifGetLine(p, pProduct), pProduct, Vec_PtrSize(vFanins) );
            return nil;
        }
        if ((pOutput[0] - '0') & 0x8E) || pOutput[1] {
            sprintf( p.sError, "Line %d: Output value (%s) is incorrect.", Io_BlifGetLine(p, pProduct), pOutput );
            return nil;
        }
        if Polarity == -1 {
            Polarity = pOutput[0] - '0';
        } else if Polarity != pOutput[0] - '0' {
            sprintf( p.sError, "Line %d: Output value (%s) differs from the value in the first line of the table (%d).", Io_BlifGetLine(p, pProduct), pOutput, Polarity );
            return nil;
        }
        // parse one product product
        pCube = Abc_AigConst1(p.pAig);
        for  k = 0; pProduct[k]; k++  {
            if pProduct[k] == '0' {
                pCube = Abc_AigAnd( (Abc_Aig_t *)p.pAig.pManFunc, pCube, Abc_ObjNot((Abc_Obj_t *)Vec_PtrEntry(vFanins,k)) );
            } else if pProduct[k] == '1' {
                pCube = Abc_AigAnd( (Abc_Aig_t *)p.pAig.pManFunc, pCube, (Abc_Obj_t *)Vec_PtrEntry(vFanins,k) );
            } else if pProduct[k] != '-' {
                sprintf( p.sError, "Line %d: Product term (%s) contains character (%c).", Io_BlifGetLine(p, pProduct), pProduct, pProduct[k] );
                return nil;
            }
        }
        pRes = Abc_AigOr( (Abc_Aig_t *)p.pAig.pManFunc, pRes, pCube );
    }
    pRes = Abc_ObjNotCond( pRes, Polarity == 0 );
    return pRes;
}

/**Function*************************************************************

  Synopsis    [Constructs the AIG from the file parsing info.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_BlifParseConstruct_rec(p *Io_BlifMan_t, pName string) *Abc_Obj_t {
    var vFanins *Vec_Ptr_t;
    var pFaninAbc *Abc_Obj_t;
    var pObjIo *Io_BlifObj_t;
    var pNameFanin string;
    var i int32;
    // get the IO object with this name
    pObjIo = *Io_BlifHashLookup( p, pName );
    if pObjIo == nil {
        sprintf( p.sError, "Line %d: Signal (%s) is not defined as a table.", Io_BlifGetLine(p, pName), pName );
        return nil;
    }
    // loop detection
    if pObjIo.fLoop {
        sprintf( p.sError, "Line %d: Signal (%s) appears twice on a combinational path.", Io_BlifGetLine(p, pName), pName );
        return nil;
    }
    // check if the AIG is already constructed
    if pObjIo.pEquiv {
        return (Abc_Obj_t *)pObjIo.pEquiv; }
    // mark this node on the path
    pObjIo.fLoop = 1;
    // construct the AIGs for the fanins
    vFanins = Vec_PtrAlloc( 8 );
    Io_BlifCollectTokens( vFanins, pObjIo.pName - pObjIo.Offset, pObjIo.pName );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vFanins)) && (((pNameFanin) = (char *)Vec_PtrEntry(vFanins, i)), 1); i++  {
        pFaninAbc = Io_BlifParseConstruct_rec( p, pNameFanin );
        if pFaninAbc == nil {
            Vec_PtrFree( vFanins );
            return nil;
        }
        Vec_PtrWriteEntry( vFanins, i, pFaninAbc );
    }
    // construct the node
    pObjIo.pEquiv = Io_BlifParseTable( p, pObjIo.pName + len(pObjIo.pName), vFanins );
    Vec_PtrFree( vFanins );
    // unmark this node on the path
    pObjIo.fLoop = 0;
    // remember the new node
    return (Abc_Obj_t *)pObjIo.pEquiv;
}

/**Function*************************************************************

  Synopsis    [Constructs the AIG from the file parsing info.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_BlifParseConstruct(p *Io_BlifMan_t) int32 {
    var pAig *Abc_Ntk_t;
    var pObjIo, pObjIoInput *Io_BlifObj_t;
    var pObj, pLatch *Abc_Obj_t;
    var i int32;
    // allocate the empty AIG
    pAig = p.pAig = Abc_NtkAlloc( ABC_NTK_STRASH, ABC_FUNC_AIG, 1 );
    pAig.pName = Extra_UtilStrsav( p.pModel );
    pAig.pSpec = Extra_UtilStrsav( p.pFileName );
    // create PIs
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vPis)) && (((pObjIo) = (Io_BlifObj_t *)Vec_PtrEntry(p.vPis, i)), 1); i++  {
        pObj = Abc_NtkCreatePi( pAig );
        Abc_ObjAssignName( pObj, pObjIo.pName, nil );
        pObjIo.pEquiv = pObj;
    }
    // create POs
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vPos)) && (((pObjIo) = (Io_BlifObj_t *)Vec_PtrEntry(p.vPos, i)), 1); i++  {
        pObj = Abc_NtkCreatePo( pAig );
        Abc_ObjAssignName( pObj, pObjIo.pName, nil );
    }
    // create latches
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vLos)) && (((pObjIo) = (Io_BlifObj_t *)Vec_PtrEntry(p.vLos, i)), 1); i++  {
        // add the latch input terminal
        pObj = Abc_NtkCreateBi( pAig );
        pObjIoInput = (Io_BlifObj_t *)Vec_PtrEntry( p.vLis, i );
        Abc_ObjAssignName( pObj, pObjIoInput.pName, nil );
        
        // add the latch box
        pLatch = Abc_NtkCreateLatch( pAig );
        pLatch.pData = (void *)(ABC_PTRUINT_T)pObjIo.Init;
        Abc_ObjAssignName( pLatch, pObjIo.pName, "L" );
        Abc_ObjAddFanin( pLatch, pObj  );

        // add the latch output terminal
        pObj = Abc_NtkCreateBo( pAig );
        Abc_ObjAssignName( pObj, pObjIo.pName, nil );
        Abc_ObjAddFanin( pObj, pLatch );
        // set the value of the latch output
//        pObjIo->pEquiv = Abc_ObjNotCond( pObj, pObjIo->Init );
        pObjIo.pEquiv = pObj;
    }
    // traverse the nodes from the POs
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vPos)) && (((pObjIo) = (Io_BlifObj_t *)Vec_PtrEntry(p.vPos, i)), 1); i++  {
        pObj = Io_BlifParseConstruct_rec( p, pObjIo.pName );
        if pObj == nil {
            return 0; }
        Abc_ObjAddFanin( Abc_NtkPo(p.pAig, i), pObj );
    }
    // traverse the nodes from the latch inputs
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vLis)) && (((pObjIo) = (Io_BlifObj_t *)Vec_PtrEntry(p.vLis, i)), 1); i++  {
        pObj = Io_BlifParseConstruct_rec( p, pObjIo.pName );
        if pObj == nil {
            return 0; }
//        pObj = Abc_ObjNotCond( pObj, pObjIo->Init );
        Abc_ObjAddFanin( Abc_ObjFanin0(Abc_NtkBox(p.pAig, i)), pObj );
    }
    p.nTablesLeft = Vec_PtrSize(p.vNames) - p.nTablesRead; 
    if p.nTablesLeft { 
        fmt.Printf("The number of dangling tables = %d.\n", p.nTablesLeft); }
    fmt.Printf("AND nodes = %6d.  Estimate = %6d.\n", Abc_NtkNodeNum(p.pAig), Io_BlifEstimateAndNum(p));
    return 1;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

