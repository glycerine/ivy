package abci

import "fmt"

import "os"

import (
	"unsafe"

	dsc "github.com/glycerine/abcgo/opt/dsc"
	util "github.com/glycerine/abcgo/misc/util"
)

/**CFile****************************************************************

  FileName    [abcDec.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Network and node package.]

  Synopsis    [Procedures for testing and comparing decomposition algorithms.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: abcDec.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "misc/extra/extra.h"
//#include "misc/vec/vec.h"

//#include "bool/bdc/bdc.h"
//#include "bool/dec/dec.h"
//#include "bool/kit/kit.h"
//#include "opt/dau/dau.h"
//#include "misc/util/utilTruth.h"
//#include "opt/dsc/dsc.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////
 
// decomposition type
// 0 - none
// 1 - factoring
// 2 - bi-decomposition
// 3 - DSD

// data-structure to store a bunch of truth tables
    
type Abc_TtStore_t struct {
    nVars               int32;
    nWords               int32;
    nFuncs               int32;
    pFuncs            **word;
}

// read/write/flip i-th bit of a bit string table:
func Abc_TruthGetBit(p *word, i int32) int32 { return int32((p[i>>6] >> (i & 63))) & 1;        }
func Abc_TruthSetBit(p *word, i int32) { p[i>>6] |= (((word)1)<<(i & 63));             }
func Abc_TruthXorBit(p *word, i int32) { p[i>>6] ^= (((word)1)<<(i & 63));             }

// read/write k-th digit d of a hexadecimal number:
func Abc_TruthGetHex(p *word, k int32) int32 { return int32((p[k>>4] >> ((k<<2) & 63))) & 15;  }
func Abc_TruthSetHex(p *word, k int32, d int32) { p[k>>4] |= (((word)d)<<((k<<2) & 63));        }
func Abc_TruthXorHex(p *word, k int32, d int32) { p[k>>4] ^= (((word)d)<<((k<<2) & 63));        }

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

// read one hex character
func Abc_TruthReadHexDigit(HexChar byte) int32 {
    if HexChar >= '0' && HexChar <= '9' {
        return HexChar - '0'; }
    if HexChar >= 'A' && HexChar <= 'F' {
        return HexChar - 'A' + 10; }
    if HexChar >= 'a' && HexChar <= 'f' {
        return HexChar - 'a' + 10; }
    assert(false, "0"); // not a hexadecimal symbol
    return -1; // return value which makes no sense
}

// write one hex character
func Abc_TruthWriteHexDigit(pFile *os.File, HexDigit int32) {
    assert(HexDigit >= 0 && HexDigit < 16, "HexDigit >= 0 && HexDigit < 16");
    if HexDigit < 10 {
        fmt.Fprintf(pFile, "%d", HexDigit);
    } else {
        fmt.Fprintf(pFile, "%c", 'A' + HexDigit-10); }
}

// read one truth table in hexadecimal
func Abc_TruthReadHex(pTruth *word, pString string, nVars int32) {
    var nWords int32;
    if nVars < 7 {
        nWords = 1;
    } else {
        nWords = (1 << (nVars-6));
    }
    var k int32;
    var Digit int32;
    var nDigits int32;
    if nVars < 7 {
        nDigits = (1 << (nVars-2));
    } else {
        nDigits = (nWords << 4);
    }
    var EndSymbol byte;
    // skip the first 2 symbols if they are "0x"
    if pString[0] == '0' && pString[1] == 'x' {
        pString += 2; }
    // get the last symbol
    EndSymbol = pString[nDigits];
    // the end symbol of the TT (the one immediately following hex digits)
    // should be one of the following: space, a new-line, or a zero-terminator
    // (note that on Windows symbols '\r' can be inserted before each '\n')
    assert(EndSymbol == ' ' || EndSymbol == '\n' || EndSymbol == '\r' || EndSymbol == '\0', "EndSymbol == ' ' || EndSymbol == '\\n' || EndSymbol == '\\r' || EndSymbol == '\\0'");
    // read hexadecimal digits in the reverse order
    // (the last symbol in the string is the least significant digit)
    for  k = 0; k < nDigits; k++  {
        Digit = Abc_TruthReadHexDigit( pString[nDigits - 1 - k] );
        assert(Digit >= 0 && Digit < 16, "Digit >= 0 && Digit < 16");
        Abc_TruthSetHex( pTruth, k, Digit );
    }
}

// write one truth table in hexadecimal (do not add end-of-line!)
func Abc_TruthWriteHex(pFile *os.File, pTruth *word, nVars int32) {
    var nDigits, Digit, k int32;
    nDigits = (1 << (nVars-2));
    for  k = 0; k < nDigits; k++  {
        Digit = Abc_TruthGetHex( pTruth, k );
        assert(Digit >= 0 && Digit < 16, "Digit >= 0 && Digit < 16");
        Abc_TruthWriteHexDigit( pFile, Digit );
    }
}


/**Function*************************************************************

  Synopsis    [Allocate/Deallocate storage for truth tables..]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_TruthStoreAlloc(nVars int32, nFuncs int32) *Abc_TtStore_t {
    var p *Abc_TtStore_t;
    var i int32;
    p = (Abc_TtStore_t *)malloc( sizeof(Abc_TtStore_t) );
    p.nVars  =  nVars;
    if nVars < 7 {
        p.nWords = 1;
    } else {
        p.nWords = (1 << (nVars-6));
    }
    p.nFuncs =  nFuncs;
    // alloc storage for 'nFuncs' truth tables as one chunk of memory
    p.pFuncs = (word **)malloc( (sizeof(word *) + sizeof(word) * p.nWords) * p.nFuncs );
    // assign and clean the truth table storage
    p.pFuncs[0] = (word *)(p.pFuncs + p.nFuncs);
    memset( p.pFuncs[0], 0, sizeof(word) * p.nWords * p.nFuncs );
    // split it up into individual truth tables
    for  i = 1; i < p.nFuncs; i++  {
        p.pFuncs[i] = p.pFuncs[i-1] + p.nWords;
    }
    return p;
}
func Abc_TruthStoreAlloc2(nVars int32, nFuncs int32, pBuffer *word) *Abc_TtStore_t {
    var p *Abc_TtStore_t;
    var i int32;
    p = (Abc_TtStore_t *)malloc( sizeof(Abc_TtStore_t) );
    p.nVars  =  nVars;
    if nVars < 7 {
        p.nWords = 1;
    } else {
        p.nWords = (1 << (nVars-6));
    }
    p.nFuncs =  nFuncs;
    // alloc storage for 'nFuncs' truth tables as one chunk of memory
    p.pFuncs = (word **)malloc( sizeof(word *) * p.nFuncs );
    // assign and clean the truth table storage
    p.pFuncs[0] = pBuffer;
    // split it up into individual truth tables
    for  i = 1; i < p.nFuncs; i++  {
        p.pFuncs[i] = p.pFuncs[i-1] + p.nWords;
    }
    return p;
}
func Abc_TtStoreFree(p *Abc_TtStore_t, nVarNum int32) {
    if nVarNum >= 0 {
        ABC_FREE( p.pFuncs[0] ); }
    ABC_FREE( p.pFuncs );
    ABC_FREE( p );
}

/**Function*************************************************************

  Synopsis    [Read file contents.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_FileSize(pFileName string) int32 {
    var pFile *os.File;
    var nFileSize int32;
    pFile = os.Open(pFileName);
    if pFile == nil {
        fmt.Printf("Cannot open file \"%s\" for reading.\n", pFileName);
        return -1;
    }
    // get the file size, in bytes
    fseek( pFile, 0, SEEK_END );  
    nFileSize = ftell( pFile );  
    _ = (pFile).Close();
    return nFileSize;
}

/**Function*************************************************************

  Synopsis    [Read file contents.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_FileRead(pFileName string) string {
    var pFile *os.File;
    var pBuffer string;
    var nFileSize, RetValue int32;
    pFile = os.Open(pFileName);
    if pFile == nil {
        fmt.Printf("Cannot open file \"%s\" for reading.\n", pFileName);
        return nil;
    }
    // get the file size, in bytes
    fseek( pFile, 0, SEEK_END );  
    nFileSize = ftell( pFile );  
    // move the file current reading position to the beginning
    rewind( pFile ); 
    // load the contents of the file into memory
    pBuffer = (char *)malloc( nFileSize + 3 );
    RetValue = fread( pBuffer, nFileSize, 1, pFile );
    // add several empty lines at the end
    // (these will be used to signal the end of parsing)
    pBuffer[ nFileSize + 0] = '\n';
    pBuffer[ nFileSize + 1] = '\n';
    // terminate the string with '\0'
    pBuffer[ nFileSize + 2] = '\0';
    _ = (pFile).Close();
    return pBuffer;
}

/**Function*************************************************************

  Synopsis    [Determine the number of variables by reading the first line.]

  Description [Determine the number of functions by counting the lines.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_TruthGetParams(pFileName string, pnVars *int32, pnTruths *int32) {
    var pContents string;
    var i, nVars, nLines int32;
    // prepare the output 
    if pnVars != nil {
        *pnVars = 0; }
    if pnTruths != nil {
        *pnTruths = 0; }
    // read data from file
    pContents = Abc_FileRead( pFileName );
    if pContents == nil {
        return; }
    // count the number of symbols before the first space or new-line
    // (note that on Windows symbols '\r' can be inserted before each '\n')
    for  i = 0; pContents[i]; i++  {
        if pContents[i] == ' ' || pContents[i] == '\n' || pContents[i] == '\r' {
            break; }
    }
    if pContents[i] == 0 {
        fmt.Printf("Strange, the input file does not have spaces and new-lines...\n"); }

    // acount for the fact that truth tables may have "0x" at the beginning of each line
    if pContents[0] == '0' && pContents[1] == 'x' {
        i = i - 2; }

    // determine the number of variables
    for  nVars = 0; nVars < 32; nVars++  {
        if 4 * i == (1 << nVars) { // the number of bits equal to the size of truth table
            break; }
    }
    if nVars < 2 || nVars > 16 {
        fmt.Printf("Does not look like the input file contains truth tables...\n");
        return;
    }
    if pnVars != nil {
        *pnVars = nVars; }

    // determine the number of functions by counting the lines
    nLines = 0;
    for  i = 0; pContents[i]; i++  {
        nLines += (pContents[i] == '\n');
    }
    if pnTruths != nil {
        *pnTruths = nLines; }
    ABC_FREE( pContents );
}


/**Function*************************************************************

  Synopsis    [Read truth tables from file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_TruthStoreRead(pFileName string, p *Abc_TtStore_t) {
    var pContents string;
    var i, nLines int32;
    pContents = Abc_FileRead( pFileName );
    if pContents == nil {
        return; }
    // here it is assumed (without checking!) that each line of the file 
    // begins with a string of hexadecimal chars followed by space

    // the file will be read till the first empty line (pContents[i] == '\n')
    // (note that Abc_FileRead() added several empty lines at the end of the file contents)
    for  nLines = i = 0; pContents[i] != '\n';  {
        // read one line
        Abc_TruthReadHex( p.pFuncs[nLines++], &pContents[i], p.nVars );
        // skip till after the end-of-line symbol
        // (note that end-of-line symbol is also skipped)
        for  pContents[i++] != '\n'  {; }
    }
    // adjust the number of functions read 
    // (we may have allocated more storage because some lines in the file were empty)
    assert(p.nFuncs >= nLines, "p.nFuncs >= nLines");
    p.nFuncs = nLines;
    ABC_FREE( pContents );
}

/**Function*************************************************************

  Synopsis    [Write truth tables into file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_TtStoreWrite(pFileName string, p *Abc_TtStore_t, fBinary int32) {
    var pFile *os.File;
    var pBuffer [1000]byte;
    var i int32;
    var nBytes int32 = 8 * Abc_Truth6WordNum( p.nVars );
    pFile = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    if pFile == nil {
        fmt.Printf("Cannot open file \"%s\" for writing.\n", pFileName);
        return;
    }
    for  i = 0; i < p.nFuncs; i++  {
        if fBinary != 0 {
            fwrite( p.pFuncs[i], nBytes, 1, pFile );
        } else {
            Abc_TruthWriteHex( pFile, p.pFuncs[i], p.nVars );
fmt.Fprintf(pFile, "    ");
            Dau_DsdDecompose( p.pFuncs[i], p.nVars, 0, int32((p.nVars <= 10)), pBuffer );
            fmt.Fprintf(pFile, "%s\n", pBuffer);
        }
    }
    _ = (pFile).Close();
}


/**Function*************************************************************

  Synopsis    [Read truth tables from input file and write them into output file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_TtStoreLoad(pFileName string, nVarNum int32) *Abc_TtStore_t { 
    var p *Abc_TtStore_t;
    if nVarNum < 0 {
        var nVars, nTruths int32;
        // figure out how many truth table and how many variables
        Abc_TruthGetParams( pFileName, &nVars, &nTruths );
        if nVars < 2 || nVars > 16 || nTruths == 0 {
            return nil; }
        // allocate data-structure
        p = Abc_TruthStoreAlloc( nVars, nTruths );
        // read info from file
        Abc_TruthStoreRead( pFileName, p );
    
    } else { 
        var pBuffer string;
        var nFileSize int32  = Abc_FileSize( pFileName );
        var nBytes int32  = (1 << (nVarNum-3));
        var nTruths int32  = nFileSize / nBytes;
        if nFileSize == -1 {
            return nil; }
        assert(nVarNum >= 6, "nVarNum >= 6");
        if nFileSize % nBytes != 0 { 
            Abc_Print( 0, "The file size (%d) is divided by the truth table size (%d) with remainder (%d).\n", 
                nFileSize, nBytes, nFileSize % nBytes ); }
        // read file contents
        pBuffer = Abc_FileRead( pFileName );
        // allocate data-structure
        p = Abc_TruthStoreAlloc2( nVarNum, nTruths, (word *)pBuffer );
    }
    return p;
}

/**Function*************************************************************

  Synopsis    [Read truth tables from input file and write them into output file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_TtStoreLoadSave(pFileName string) { 
    var p *Abc_TtStore_t;
    var pFileInput string   = pFileName;
    var pFileOutput string  = Extra_FileNameGenericAppend(pFileName, "_binary.data");

    // read info from file
    p = Abc_TtStoreLoad( pFileInput, -1 );
    if p == nil {
        return; }

    // write into another file
    Abc_TtStoreWrite( pFileOutput, p, 1 );

    // delete data-structure
    Abc_TtStoreFree( p, -1 );
    fmt.Printf("Input file \"%s\" was copied into output file \"%s\".\n", pFileInput, pFileOutput);
}

/**Function*************************************************************

  Synopsis    [Read truth tables in binary text form and write them into file as binary data.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_TtStoreLoadSaveBin(pFileName string) { 
    var pTruth *uint32  = ABC_CALLOC( unsigned, (1 << 11) );
    var pBuffer string  = ABC_CALLOC( char, (1 << 16) );
    var pFileInput string   = pFileName;
    var pFileOutput string  = Extra_FileNameGenericAppend(pFileName, "_binary.data");
    var pFileI *os.File  = os.Open(pFileInput);
    var pFileO *os.File  = os.OpenFile(pFileOutput, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    var i int32;
    var Value int32;
    var nVarsAll int32 = -1;
    if pFileI == nil {
        return; }
    for  fgets(pBuffer, (1 << 16), pFileI)  {
        var Len int32    = len(pBuffer)-1; // subtract 1 for end-of-line
        var nVars int32  = Abc_Base2Log(Len);
        var nInts int32  = Abc_BitWordNum(Len);
        assert(Len == (1 << nVars), "Len == (1 << nVars)");
        if nVarsAll == -1 {
            nVarsAll = nVars;
        } else {
            assert(nVarsAll == nVars, "nVarsAll == nVars"); }
        memset( pTruth, 0, sizeof(int)*nInts );
        for  i = 0; i < Len; i++  {
            if pBuffer[i] == '1' {
                Abc_InfoSetBit( pTruth, i );
            } else {
                assert(pBuffer[i] == '0', "pBuffer[i] == '0'"); }
        }
        Value = fwrite( pTruth, 1, sizeof(int) * nInts, pFileO );
        assert(Value == int32(sizeof(int)) * nInts, "Value == int32(sizeof(int)) * nInts");
    }
    ABC_FREE( pTruth );
    ABC_FREE( pBuffer );
    _ = (pFileI).Close();
    _ = (pFileO).Close();
    fmt.Printf("Input file \"%s\" was copied into output file \"%s\".\n", pFileInput, pFileOutput);
}

/**Function*************************************************************

  Synopsis    [Read truth tables from input file and write them into output file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_TtStoreTest(pFileName string) { 
    var p *Abc_TtStore_t;
    var pFileInput string   = pFileName;
    var pFileOutput string  = "out.txt";

    // read info from file
    p = Abc_TtStoreLoad( pFileInput, -1 );
    if p == nil {
        return; }

    // write into another file
    Abc_TtStoreWrite( pFileOutput, p, 0 );

    // delete data-structure
    Abc_TtStoreFree( p, -1 );
    fmt.Printf("Input file \"%s\" was copied into output file \"%s\".\n", pFileInput, pFileOutput);
}

/**Function*************************************************************

  Synopsis    [Apply decomposition to the truth table.]

  Description [Returns the number of AIG nodes.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_TruthDecPerform(p *Abc_TtStore_t, DecType int32, fVerbose int32) {
    var clk abctime  = Abc_Clock();
    var i, nNodes int32;

    var pAlgoName string  = nil;
    if DecType == 1 {
        pAlgoName = "factoring";
    } else if DecType == 2 {
        pAlgoName = "bi-decomp";
    } else if DecType == 3 {
        pAlgoName = "DSD";
    } else if DecType == 4 {
        pAlgoName = "fast DSD";
    } else if DecType == 5 {
        pAlgoName = "analysis";
    } else if DecType == 6 {
            pAlgoName = "DSD ICCD'15"; }

    if pAlgoName {
        fmt.Printf("Applying %-10s to %8d func%s of %2d vars...  ", pAlgoName, p.nFuncs, (p.nFuncs == 1 ? "":"s"), p.nVars); }
    if fVerbose != 0 {
        fmt.Printf("\n"); }

    if DecType == 1 {
        // perform algebraic factoring and count AIG nodes
        var pFForm *Dec_Graph_t;
        var vCover *Vec_Int_t;
        var vStr *Vec_Str_t;
        var pSopStr string;
        vStr = Vec_StrAlloc( 10000 );
        vCover = Vec_IntAlloc( 1 << 16 );
        for  i = 0; i < p.nFuncs; i++  {
//            extern int Abc_IsopTest( word * pFunc, int nVars, Vec_Int_t * vCover );
//            if ( i == 0 ) printf( "\n" );
//            Abc_IsopTest( p->pFuncs[i], p->nVars, vCover );
//            continue;
            if fVerbose != 0 {
                fmt.Printf("%7d : ", i); }
            pSopStr = Kit_PlaFromTruthNew( (unsigned *)p.pFuncs[i], p.nVars, vCover, vStr );
            pFForm = Dec_Factor( pSopStr );
            nNodes += Dec_GraphNodeNum( pFForm );
            if fVerbose != 0 {
                Dec_GraphPrint( os.Stdout, pFForm, nil, nil ); }
            Dec_GraphFree( pFForm );
        }
        Vec_IntFree( vCover );
        Vec_StrFree( vStr );
    
    } else if DecType == 2 {
        // perform bi-decomposition and count AIG nodes
        var pManDec *Bdc_Man_t;
        var Pars Bdc_Par_t ;
        var pPars *Bdc_Par_t = &Pars;
        pPars.nVarsMax = p.nVars;
        pManDec = Bdc_ManAlloc( pPars );
        for  i = 0; i < p.nFuncs; i++  {
            if fVerbose != 0 {
                fmt.Printf("%7d :      ", i); }
            Bdc_ManDecompose( pManDec, (unsigned *)p.pFuncs[i], nil, p.nVars, nil, 1000 );
            nNodes += Bdc_ManAndNum( pManDec );
            if fVerbose != 0 {
                Bdc_ManDecPrint( pManDec ); }
        }
        Bdc_ManFree( pManDec );
    
    } else if DecType == 3 {
        // perform disjoint-support decomposition and count AIG nodes
        // (non-DSD blocks are decomposed into 2:1 MUXes, each counting as 3 AIG nodes)
        var pNtk *Kit_DsdNtk_t;
        for  i = 0; i < p.nFuncs; i++  {
            if fVerbose != 0 {
                fmt.Printf("%7d :      ", i); }
            pNtk = Kit_DsdDecomposeMux( (unsigned *)p.pFuncs[i], p.nVars, 3 );
            if fVerbose != 0 {
                Kit_DsdPrintExpanded( pNtk );
fmt.Printf("\n"); }
            nNodes += Kit_DsdCountAigNodes( pNtk );
            Kit_DsdNtkFree( pNtk );
        }
    
    } else if DecType == 4 {
        var pDsd [DAU_MAX_STR]byte;
        for  i = 0; i < p.nFuncs; i++  {
            if fVerbose != 0 {
                fmt.Printf("%7d :      ", i); }
            Dau_DsdDecompose( p.pFuncs[i], p.nVars, 0, 1, pDsd );
            if fVerbose != 0 {
                printf( "%s\n", pDsd ); }
            nNodes += Dau_DsdCountAnds( pDsd );
        }
    
    } else if DecType == 5 {
        for  i = 0; i < p.nFuncs; i++  {
            extern void Dau_DecTrySets( word * pInit, int nVars, int fVerbose );
            var nSuppSize int32  = Abc_TtSupportSize( p.pFuncs[i], p.nVars );
            if fVerbose != 0 {
                fmt.Printf("%7d :      ", i); }
            Dau_DecTrySets( p.pFuncs[i], nSuppSize, fVerbose );
            if fVerbose != 0 {
                fmt.Printf("\n"); }
        }
    } else if DecType == 6 {
        /* memory pool with a capacity of storing 3*nVars
        truth-tables for negative and positive cofactors and
        the boolean difference for each input variable */
        var mem_pool []uint64 = dsc.Dsc_AllocPool(p.nVars)
        var funcs = unsafe.Slice(p.pFuncs, int(p.nFuncs))
        for  i = 0; i < p.nFuncs; i++  {
            if fVerbose != 0 {
                fmt.Printf("%7d :      ", i); }
            var truth = unsafe.Slice(funcs[i], int(util.Abc_TtWordNum(p.nVars)))
            pDsd, status := dsc.Dsc_Decompose(truth, p.nVars, mem_pool)
            if fVerbose != 0 {
                if status == 0 {
                    fmt.Printf("%s\n", pDsd)
                } else {
                    fmt.Printf("NULL\n")
                }
            }
            nNodes += dsc.Dsc_CountAnds(pDsd)
        }
        dsc.Dsc_FreePool(mem_pool)
    
    } else { assert(false, "0"); }

    fmt.Printf("AIG nodes =%9d  ", nNodes);
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
}

/**Function*************************************************************

  Synopsis    [Apply decomposition to truth tables.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_TruthDecTest(pFileName string, DecType int32, nVarNum int32, fVerbose int32) {
    var p *Abc_TtStore_t;

    // allocate data-structure
    p = Abc_TtStoreLoad( pFileName, nVarNum );
    if p == nil { return; }

    // consider functions from the file
    Abc_TruthDecPerform( p, DecType, fVerbose );

    // delete data-structure
    Abc_TtStoreFree( p, nVarNum );
//    printf( "Finished decomposing truth tables from file \"%s\".\n", pFileName );
}


/**Function*************************************************************

  Synopsis    [Testbench for decomposition algorithms.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_DecTest(pFileName string, DecType int32, nVarNum int32, fVerbose int32) int32 {
    if fVerbose != 0 {
        fmt.Printf("Using truth tables from file \"%s\"...\n", pFileName); }
    if DecType == 0 { if nVarNum < 0 { Abc_TtStoreTest( pFileName ); } 
    } else if DecType >= 1 && DecType <= 6 {
        Abc_TruthDecTest( pFileName, DecType, nVarNum, fVerbose );
    } else {
        fmt.Printf("Unknown decomposition type value (%d).\n", DecType); }
    fflush( os.Stdout );
    return 0;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

