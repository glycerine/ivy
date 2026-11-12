package lucky

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [luckyRead.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Semi-canonical form computation package.]

  Synopsis    [Reading truth tables from file.]

  Author      [Jake]

  Date        [Started - August 2012]

***********************************************************************/

//#include "luckyInt.h"

//ABC_NAMESPACE_IMPL_START


// read/write/flip i-th bit of a bit string table:
func Abc_TruthGetBit(p *word, i int32) int32 { return int32((p[i>>6] >> (i & 63))) & 1;        }
func Abc_TruthSetBit(p *word, i int32) { p[i>>6] |= (((word)1)<<(i & 63));             }
func Abc_TruthXorBit(p *word, i int32) { p[i>>6] ^= (((word)1)<<(i & 63));             }

// read/write k-th digit d of a hexadecimal number:
func Abc_TruthGetHex(p *word, k int32) int32 { return int32((p[k>>4] >> ((k<<2) & 63))) & 15;  }
func Abc_TruthSetHex(p *word, k int32, d int32) { p[k>>4] |= (((word)d)<<((k<<2) & 63));        }
func Abc_TruthXorHex(p *word, k int32, d int32) { p[k>>4] ^= (((word)d)<<((k<<2) & 63));        }

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
    // write hexadecimal digits in the reverse order
    // (the last symbol in the string is the least significant digit)
    nDigits = (1 << (nVars-2));
    for  k = 0; k < nDigits; k++  {
        Digit = Abc_TruthGetHex( pTruth, nDigits - 1 - k );
        assert(Digit >= 0 && Digit < 16, "Digit >= 0 && Digit < 16");
        Abc_TruthWriteHexDigit( pFile, Digit );
    }
}

// allocate and clear memory to store 'nTruths' truth tables of 'nVars' variables
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
    // alloc array of 'nFuncs' pointers to truth tables
    p.pFuncs = (word **)malloc( sizeof(word *) * p.nFuncs );
    // alloc storage for 'nFuncs' truth tables as one chunk of memory
    p.pFuncs[0] = (word *)calloc( sizeof(word), p.nFuncs * p.nWords );
    // split it up into individual truth tables
    for  i = 1; i < p.nFuncs; i++  {
        p.pFuncs[i] = p.pFuncs[i-1] + p.nWords;
    }
    return p;
}

// free memory previously allocated for storing truth tables
func Abc_TruthStoreFree(p *Abc_TtStore_t) {
    free( p.pFuncs[0] );
    free( p.pFuncs );
    free( p );
}

// read file contents
func Abc_FileRead(pFileName string) string {
    var pFile *os.File;
    var pBuffer string;
    var nFileSize int32;
    var RetValue int32;
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

// determine the number of variables by reading the first line
// determine the number of functions by counting the lines
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

    // account for the fact that truth tables may have "0x" at the beginning of each line
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
}

func Abc_Create_TtSpore(pFileInput string) *Abc_TtStore_t {
    var nVars, nTruths int32;
    var p *Abc_TtStore_t;
    Abc_TruthGetParams( pFileInput, &nVars, &nTruths );
    p = Abc_TruthStoreAlloc( nVars, nTruths );
    return p;

}
// read truth tables from file
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
}

// write truth tables into file
// (here we write one symbol at a time - it can be optimized by writing 
// each truth table into a string and then writing the string into a file)
func Abc_TruthStoreWrite(pFileName string, p *Abc_TtStore_t) {
    var pFile *os.File;
    var i int32;
    pFile = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    if pFile == nil {
        fmt.Printf("Cannot open file \"%s\" for writing.\n", pFileName);
        return;
    }
    for  i = 0; i < p.nFuncs; i++  {
        Abc_TruthWriteHex( pFile, p.pFuncs[i], p.nVars );
        fmt.Fprintf(pFile, "\n");
    }
    _ = (pFile).Close();
}

func WriteToFile(pFileName string, p *Abc_TtStore_t, a *word) {
    var pFile *os.File;
    var i int32;
    pFile = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    if pFile == nil {
        fmt.Printf("Cannot open file \"%s\" for writing.\n", pFileName);
        return;
    }
    for  i = 0; i < p.nFuncs; i++  {
        Abc_TruthWriteHex( pFile, &a[i], p.nVars );
        fmt.Fprintf(pFile, "\n");
    }
    _ = (pFile).Close();
}

func WriteToFile1(pFileName string, p *Abc_TtStore_t, a **word) {
    var pFile *os.File;
    var i, j int32;
    pFile = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    if pFile == nil {
        fmt.Printf("Cannot open file \"%s\" for writing.\n", pFileName);
        return;
    }
    for  i = 0; i < p.nFuncs; i++  {
        fmt.Fprintf(pFile, "0");
        fmt.Fprintf(pFile, "x");
        for  j=p.nWords-1; j >= 0; j--  {
            Abc_TruthWriteHex( pFile, &a[i][j], p.nVars );
        }
        fmt.Fprintf(pFile, "\n");
    }
    fmt.Fprintf(pFile, "\n");
    _ = (pFile).Close();
}
func WriteToFile2(pFileName string, p *Abc_TtStore_t, a *word) {
    var pFile *os.File;
    var i, j int32;
    pFile = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    if pFile == nil {
        fmt.Printf("Cannot open file \"%s\" for writing.\n", pFileName);
        return;
    }
    for  i = 0; i < p.nFuncs; i++  {
        fmt.Fprintf(pFile, "0");
        fmt.Fprintf(pFile, "x");
        for  j=p.nWords-1; j >= 0; j--  {
            Abc_TruthWriteHex( pFile, a+i, p.nVars );
        }
        fmt.Fprintf(pFile, "\n");
    }
    fmt.Fprintf(pFile, "\n");
    _ = (pFile).Close();
}


func setTtStore(pFileInput string) *Abc_TtStore_t {
    var nVars, nTruths int32;
    var p *Abc_TtStore_t;
    // figure out how many truth table and how many variables
    Abc_TruthGetParams( pFileInput, &nVars, &nTruths );
    // allocate data-structure
    p = Abc_TruthStoreAlloc( nVars, nTruths );

    Abc_TruthStoreRead( pFileInput, p );
    return p;
}


//ABC_NAMESPACE_IMPL_END
