package abci

import (
	"fmt"
	"time"
)

import "os"

/**CFile****************************************************************

  FileName    [abcRpo.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Rpo package.]

  Synopsis    [Procedures for executing RPO.]

  Author      [Mayler G. A. Martins / Vinicius Callegaro]
  
  Affiliation [UFRGS]

  Date        [Ver. 1.0. Started - May 08, 2013.]

  Revision    [$Id: abcRpo.c,v 1.00 2013/05/08 00:00:00 mgamartins Exp $]

 ***********************************************************************/

//#include "misc/extra/extra.h"

//#include "bool/rpo/rpo.h"
//#include "bool/rpo/literal.h"

//ABC_NAMESPACE_IMPL_START


// data-structure to store a bunch of truth tables
   

type Rpo_TtStore_t struct {
    nVars int32;
    nWords int32;
    nFuncs int32;
    pFuncs  **word;
}


// read/write/flip i-th bit of a bit string table:

func Abc_TruthGetBit(p *word, i int32) int32 {
    return int32( (p[i >> 6] >> (i & 63))) & 1;
}

func Abc_TruthSetBit(p *word, i int32) {
    p[i >> 6] |= (((word) 1) << (i & 63));
}

func Abc_TruthXorBit(p *word, i int32) {
    p[i >> 6] ^= (((word) 1) << (i & 63));
}

// read/write k-th digit d of a hexadecimal number:

func Abc_TruthGetHex(p *word, k int32) int32 {
    return int32( (p[k >> 4] >> ((k << 2) & 63))) & 15;
}

func Abc_TruthSetHex(p *word, k int32, d int32) {
    p[k >> 4] |= (((word) d) << ((k << 2) & 63));
}

func Abc_TruthXorHex(p *word, k int32, d int32) {
    p[k >> 4] ^= (((word) d) << ((k << 2) & 63));
}

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
        fmt.Fprintf(pFile, "%c", 'A' + HexDigit - 10); }
}

// read one truth table in hexadecimal

func Abc_TruthReadHex(pTruth *word, pString string, nVars int32) {
    var nWords int32;
    if nVars < 7 {
        nWords = 1;
    } else {
        nWords = (1 << (nVars - 6));
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
    for k = 0; k < nDigits; k++ {
        Digit = Abc_TruthReadHexDigit(pString[nDigits - 1 - k]);
        assert(Digit >= 0 && Digit < 16, "Digit >= 0 && Digit < 16");
        Abc_TruthSetHex(pTruth, k, Digit);
    }
}

// write one truth table in hexadecimal (do not add end-of-line!)

func Abc_TruthWriteHex(pFile *os.File, pTruth *word, nVars int32) {
    var nDigits, Digit, k int32;
    nDigits = (1 << (nVars - 2));
    for k = 0; k < nDigits; k++ {
        Digit = Abc_TruthGetHex(pTruth, k);
        assert(Digit >= 0 && Digit < 16, "Digit >= 0 && Digit < 16");
        Abc_TruthWriteHexDigit(pFile, Digit);
    }
}



/**Function*************************************************************

  Synopsis    [Allocate/Deallocate storage for truth tables..]

  Description []
               
  SideEffects []

  SeeAlso     []

 ***********************************************************************/
func Abc_TruthStoreAlloc(nVars int32, nFuncs int32) *Rpo_TtStore_t {
    var p *Rpo_TtStore_t;
    var i int32;
    p = (Rpo_TtStore_t *) malloc(sizeof (Rpo_TtStore_t));
    p.nVars = nVars;
    if nVars < 7 {
        p.nWords = 1;
    } else {
        p.nWords = (1 << (nVars - 6));
    }
    p.nFuncs = nFuncs;
    // alloc storage for 'nFuncs' truth tables as one chunk of memory
    p.pFuncs = (word **) malloc((sizeof (word *) + sizeof (word) * p.nWords) * p.nFuncs);
    // assign and clean the truth table storage
    p.pFuncs[0] = (word *) (p.pFuncs + p.nFuncs);
    memset(p.pFuncs[0], 0, sizeof (word) * p.nWords * p.nFuncs);
    // split it up into individual truth tables
    for i = 1; i < p.nFuncs; i++ {
        p.pFuncs[i] = p.pFuncs[i - 1] + p.nWords;
    }
    return p;
}

func Abc_TruthStoreAlloc2(nVars int32, nFuncs int32, pBuffer *word) *Rpo_TtStore_t {
    var p *Rpo_TtStore_t;
    var i int32;
    p = (Rpo_TtStore_t *) malloc(sizeof (Rpo_TtStore_t));
    p.nVars = nVars;
    if nVars < 7 {
        p.nWords = 1;
    } else {
        p.nWords = (1 << (nVars - 6));
    }
    p.nFuncs = nFuncs;
    // alloc storage for 'nFuncs' truth tables as one chunk of memory
    p.pFuncs = (word **) malloc(sizeof (word *) * p.nFuncs);
    // assign and clean the truth table storage
    p.pFuncs[0] = pBuffer;
    // split it up into individual truth tables
    for i = 1; i < p.nFuncs; i++ {
        p.pFuncs[i] = p.pFuncs[i - 1] + p.nWords;
    }
    return p;
}

func Abc_TtStoreFree(p *Rpo_TtStore_t, nVarNum int32) {
    if nVarNum >= 0 {
        ABC_FREE(p.pFuncs[0]); }
    ABC_FREE(p.pFuncs);
    ABC_FREE(p);
}

/**Function*************************************************************

  Synopsis    [Read file contents.]

  Description []
               
  SideEffects []

  SeeAlso     []

 ***********************************************************************/
// extern int Abc_FileSize(char * pFileName);


/**Function*************************************************************

  Synopsis    [Read file contents.]

  Description []
               
  SideEffects []

  SeeAlso     []

 ***********************************************************************/
// extern char * Abc_FileRead(char * pFileName);


/**Function*************************************************************

  Synopsis    [Determine the number of variables by reading the first line.]

  Description [Determine the number of functions by counting the lines.]
               
  SideEffects []

  SeeAlso     []

 ***********************************************************************/
// extern void Abc_TruthGetParams(char * pFileName, int * pnVars, int * pnTruths);



/**Function*************************************************************

  Synopsis    [Read truth tables from file.]

  Description []
               
  SideEffects []

  SeeAlso     []

 ***********************************************************************/
func Abc_TruthStoreRead(pFileName string, p *Rpo_TtStore_t) {
    var pContents string;
    var i, nLines int32;
    pContents = Abc_FileRead(pFileName);
    if pContents == nil {
        return; }
    // here it is assumed (without checking!) that each line of the file 
    // begins with a string of hexadecimal chars followed by space

    // the file will be read till the first empty line (pContents[i] == '\n')
    // (note that Abc_FileRead() added several empty lines at the end of the file contents)
    for nLines = i = 0; pContents[i] != '\n'; {
        // read one line
        Abc_TruthReadHex(p.pFuncs[nLines++], &pContents[i], p.nVars);
        // skip till after the end-of-line symbol
        // (note that end-of-line symbol is also skipped)
        for pContents[i++] != '\n' {; }
    }
    // adjust the number of functions read 
    // (we may have allocated more storage because some lines in the file were empty)
    assert(p.nFuncs >= nLines, "p.nFuncs >= nLines");
    p.nFuncs = nLines;
    ABC_FREE(pContents);
}

/**Function*************************************************************

  Synopsis    [Write truth tables into file.]

  Description []
               
  SideEffects []

  SeeAlso     []

 ***********************************************************************/
func Abc_TtStoreWrite(pFileName string, p *Rpo_TtStore_t, fBinary int32) {
    var pFile *os.File;
    var i int32;
    var nBytes int32 = 8 * Abc_Truth6WordNum(p.nVars);
    pFile = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    if pFile == nil {
        fmt.Printf("Cannot open file \"%s\" for writing.\n", pFileName);
        return;
    }
    for i = 0; i < p.nFuncs; i++ {
        if fBinary != 0 {
            fwrite(p.pFuncs[i], nBytes, 1, pFile);
        } else {
            Abc_TruthWriteHex(pFile, p.pFuncs[i], p.nVars);
fmt.Fprintf(pFile, "\n"); }
    }
    _ = (pFile).Close();
}

/**Function*************************************************************

  Synopsis    [Read truth tables from input file and write them into output file.]

  Description []
               
  SideEffects []

  SeeAlso     []

 ***********************************************************************/
func Abc_TtStoreLoad(pFileName string, nVarNum int32) *Rpo_TtStore_t {
    var p *Rpo_TtStore_t;
    if nVarNum < 0 {
        var nVars, nTruths int32;
        // figure out how many truth table and how many variables
        Abc_TruthGetParams(pFileName, &nVars, &nTruths);
        if nVars < 2 || nVars > 16 || nTruths == 0 {
            return nil; }
        // allocate data-structure
        p = Abc_TruthStoreAlloc(nVars, nTruths);
        // read info from file
        Abc_TruthStoreRead(pFileName, p);
    } else {
        var pBuffer string;
        var nFileSize int32  = Abc_FileSize(pFileName);
        var nBytes int32  = (1 << (nVarNum - 3)); // why mishchencko put -3? ###
        var nTruths int32  = nFileSize / nBytes;
        //Abc_Print(-2,"nFileSize=%d,nTruths=%d\n",nFileSize, nTruths);
        if nFileSize == -1 {
            return nil; }
        assert(nVarNum >= 6, "nVarNum >= 6");
        if nFileSize % nBytes != 0 {
            Abc_Print(0, "The file size (%d) is divided by the truth table size (%d) with remainder (%d).\n",
                nFileSize, nBytes, nFileSize % nBytes); }
        // read file contents
        pBuffer = Abc_FileRead(pFileName);
        // allocate data-structure
        p = Abc_TruthStoreAlloc2(nVarNum, nTruths, (word *) pBuffer);
    }
    return p;
}


/**Function*************************************************************

  Synopsis    [Apply decomposition to the truth table.]

  Description [Returns the number of AIG nodes.]
               
  SideEffects []

  SeeAlso     []

 ***********************************************************************/
func Abc_TruthRpoPerform(p *Rpo_TtStore_t, nThreshold int32, fVerbose int32) {
    var clk int64  = time.Now().UnixNano();
    var i int32;
    var rpoCount int32;
    var lit *Literal_t;
    var percent float32;
    for i = 0; i < p.nFuncs; i++ {
//        if(i>1000) {
//            continue;
//        }
////        
//        if(i!= 2196 ) { //5886
//            continue;
//        }
        if fVerbose != 0 {
            Abc_Print(-2,"%d: ", i+1);
        }
            
        lit = Rpo_Factorize((unsigned *) p.pFuncs[i], p.nVars, nThreshold, fVerbose);
        if lit != nil {
            if fVerbose != 0 {
                Abc_Print(-2, "Solution : %s\n", lit.expression.pArray);
                Abc_Print(-2, "\n\n");
            }
            Lit_Free(lit);
            rpoCount++;
        } else {
            if fVerbose != 0 {
                Abc_Print(-2, "null\n");
                Abc_Print(-2, "\n\n");
            }
        }
    }
    percent = (rpoCount * 100.0) / p.nFuncs;
    Abc_Print(-2,"%d of %d (%.2f %%) functions are RPO.\n", rpoCount,p.nFuncs,percent);
    Abc_PrintTime(1, "Time", time.Now().UnixNano() - clk);
}

/**Function*************************************************************

  Synopsis    [Apply decomposition to truth tables.]

  Description []
               
  SideEffects []

  SeeAlso     []

 ***********************************************************************/
func Abc_TruthRpoTest(pFileName string, nVarNum int32, nThreshold int32, fVerbose int32) {
    var p *Rpo_TtStore_t;

    // allocate data-structure
//    if (fVerbose) {
//        Abc_Print(-2, "Number of variables = %d\n", nVarNum);
//    }
    p = Abc_TtStoreLoad(pFileName, nVarNum);

    if fVerbose != 0 {
        Abc_Print(-2, "Number of variables = %d\n", p.nVars);
    }
    // consider functions from the file
    Abc_TruthRpoPerform(p, nThreshold, fVerbose);

    // delete data-structure
    Abc_TtStoreFree(p, nVarNum);
    //    printf( "Finished decomposing truth tables from file \"%s\".\n", pFileName );
}

/**Function*************************************************************

  Synopsis    [Testbench for decomposition algorithms.]

  Description []
               
  SideEffects []

  SeeAlso     []

 ***********************************************************************/
func Abc_RpoTest(pFileName string, nVarNum int32, nThreshold int32, fVerbose int32) int32 {
    if fVerbose != 0 {
        fmt.Printf("Using truth tables from file \"%s\"...\n", pFileName);
    }
    Abc_TruthRpoTest(pFileName, nVarNum, nThreshold, fVerbose);
    fflush(os.Stdout);
    return 0;
}


/////////////////////ert truth table to ///////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


