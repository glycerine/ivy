package ver

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [verStream.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Verilog parser.]

  Synopsis    [Input file stream, which knows nothing about Verilog.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - August 19, 2006.]

  Revision    [$Id: verStream.c,v 1.00 2006/08/19 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "ver.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

const VER_BUFFER_SIZE =        1048576    // 1M  - size of the data chunk stored in memory
const VER_OFFSET_SIZE =          65536    // 64K - load new data when less than this is left
const VER_WORD_SIZE =            65536    // 64K - the largest token that can be returned

// Original macro:
// #define VER_MINIMUM(a,b)       (((a) < (b))? (a) : (b))


type Ver_Stream_t_ struct {
    // the input file
    pFileName            string;     // the input file name
    pFile            *os.File;         // the input file pointer
    nFileSize            iword;     // the total number of bytes in the file
    nFileRead            iword;     // the number of bytes currently read from file
    nLineCounter            iword;  // the counter of lines processed
    // temporary storage for data 
    nBufferSize            iword;   // the size of the buffer
    pBuffer            string;       // the buffer
    pBufferCur            string;    // the current reading position
    pBufferEnd            string;    // the first position not used by currently loaded data
    pBufferStop            string;   // the position where loading new data will be done
    // tokens given to the user
    pChars [VER_WORD_SIZE+5]byte; // temporary storage for a word (plus end-of-string and two parentheses)
    nChars              int32;        // the total number of characters in the word
    // status of the parser
    fStop              int32;         // this flag goes high when the end of file is reached
}

// static void Ver_StreamReload( Ver_Stream_t * p );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Starts the file reader for the given file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ver_StreamAlloc(pFileName string) *Ver_Stream_t {
    var p *Ver_Stream_t;
    var pFile *os.File;
    var nCharsToRead int32;
    var RetValue int32;
    // check if the file can be opened
    pFile = os.Open(pFileName);
    if pFile == nil {
        fmt.Printf("Ver_StreamAlloc(): Cannot open input file \"%s\".\n", pFileName);
        return nil;
    }
    // start the file reader    
    // inlining of macro ABC_ALLOC
    p = ((Ver_Stream_t *) malloc(sizeof(Ver_Stream_t) * (size_t)(1)));
    memset( p, 0, sizeof(Ver_Stream_t) );
    p.pFileName   = pFileName;
    p.pFile       = pFile;
    // get the file size, in bytes
    fseek( pFile, 0, SEEK_END );  
    p.nFileSize = ftell( pFile );  
    rewind( pFile ); 
    // allocate the buffer
    // inlining of macro ABC_ALLOC
    p.pBuffer = ((char *) malloc(sizeof(char) * (size_t)(VER_BUFFER_SIZE+1)));
    p.nBufferSize = VER_BUFFER_SIZE;
    p.pBufferCur  = p.pBuffer;
    // determine how many chars to read
    // inlining of macro VER_MINIMUM
    
    if (p.nFileSize) < (VER_BUFFER_SIZE) {
        nCharsToRead = (p.nFileSize);
    } else {
        nCharsToRead = (VER_BUFFER_SIZE);
    }
    // load the first part into the buffer
    RetValue = fread( p.pBuffer, nCharsToRead, 1, p.pFile );
    p.nFileRead = nCharsToRead;
    // set the ponters to the end and the stopping point
    p.pBufferEnd  = p.pBuffer + nCharsToRead;
    if p.nFileRead == p.nFileSize {
        p.pBufferStop = p.pBufferEnd;
    } else {
        p.pBufferStop = p.pBuffer + VER_BUFFER_SIZE - VER_OFFSET_SIZE;
    }
    // start the arrays
    p.nLineCounter = 1; // 1-based line counting
    return p;
}

/**Function*************************************************************

  Synopsis    [Loads new data into the file reader.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ver_StreamReload(p *Ver_Stream_t) {
    var nCharsUsed, nCharsToRead int32;
    var RetValue int32;
    assert(!p.fStop, "!p.fStop");
    assert(p.pBufferCur > p.pBufferStop, "p.pBufferCur > p.pBufferStop");
    assert(p.pBufferCur < p.pBufferEnd, "p.pBufferCur < p.pBufferEnd");
    // figure out how many chars are still not processed
    nCharsUsed = p.pBufferEnd - p.pBufferCur;
    // move the remaining data to the beginning of the buffer
    memmove( p.pBuffer, p.pBufferCur, (size_t)nCharsUsed );
    p.pBufferCur = p.pBuffer;
    // determine how many chars we will read
    // inlining of macro VER_MINIMUM
    
    if (p.nBufferSize - nCharsUsed) < (p.nFileSize - p.nFileRead) {
        nCharsToRead = (p.nBufferSize - nCharsUsed);
    } else {
        nCharsToRead = (p.nFileSize - p.nFileRead);
    }
    // read the chars
    RetValue = fread( p.pBuffer + nCharsUsed, nCharsToRead, 1, p.pFile );
    p.nFileRead += nCharsToRead;
    // set the ponters to the end and the stopping point
    p.pBufferEnd  = p.pBuffer + nCharsUsed + nCharsToRead;
    if p.nFileRead == p.nFileSize {
        p.pBufferStop = p.pBufferEnd;
    } else {
        p.pBufferStop = p.pBuffer + VER_BUFFER_SIZE - VER_OFFSET_SIZE;
    }
}

/**Function*************************************************************

  Synopsis    [Stops the file reader.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ver_StreamFree(p *Ver_Stream_t) {
    if p.pFile {
        _ = (p.pFile).Close(); }
    p.pBuffer = nil;
    p = nil;
}

/**Function*************************************************************

  Synopsis    [Returns the file size.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ver_StreamGetFileName(p *Ver_Stream_t) string {
    return p.pFileName;
}

/**Function*************************************************************

  Synopsis    [Returns the file size.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ver_StreamGetFileSize(p *Ver_Stream_t) int32 {
    return p.nFileSize;
}

/**Function*************************************************************

  Synopsis    [Returns the current reading position.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ver_StreamGetCurPosition(p *Ver_Stream_t) int32 {
    return p.nFileRead - (p.pBufferEnd - p.pBufferCur);
}

/**Function*************************************************************

  Synopsis    [Returns the line number for the given token.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ver_StreamGetLineNumber(p *Ver_Stream_t) int32 {
    return p.nLineCounter;
}



/**Function*************************************************************

  Synopsis    [Returns current symbol.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ver_StreamIsOkey(p *Ver_Stream_t) int32 {
    return !p.fStop;
}

/**Function*************************************************************

  Synopsis    [Returns current symbol.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ver_StreamScanChar(p *Ver_Stream_t) byte {
    assert(!p.fStop, "!p.fStop");
    return *p.pBufferCur;
}

/**Function*************************************************************

  Synopsis    [Returns current symbol and moves to the next.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ver_StreamPopChar(p *Ver_Stream_t) byte {
    assert(!p.fStop, "!p.fStop");
    // check if the new data should to be loaded
    if p.pBufferCur > p.pBufferStop {
        Ver_StreamReload( p ); }
    // check if there are symbols left
    if p.pBufferCur == p.pBufferEnd { // end of file
    
        p.fStop = 1;
        return -1;
    }
    // count the lines
    if *p.pBufferCur == '\n' {
        p.nLineCounter++; }
    return *p.pBufferCur++;
}

/**Function*************************************************************

  Synopsis    [Skips the current symbol and all symbols from the list.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ver_StreamSkipChars(p *Ver_Stream_t, pCharsToSkip string) {
    var pChar, pTemp string;
    assert(!p.fStop, "!p.fStop");
    assert(pCharsToSkip != nil, "pCharsToSkip != nil");
    // check if the new data should to be loaded
    if p.pBufferCur > p.pBufferStop {
        Ver_StreamReload( p ); }
    // skip the symbols
    for  pChar = p.pBufferCur; pChar < p.pBufferEnd; pChar++  {
        // skip symbols as long as they are in the list
        for  pTemp = pCharsToSkip; *pTemp; pTemp++  {
            if *pChar == *pTemp {
                break; }
        }
        if *pTemp == 0 { // pChar is not found in the list
        
            p.pBufferCur = pChar;
            return;
        }
        // count the lines
        if *pChar == '\n' {
            p.nLineCounter++; }
    }
    // the file is finished or the last part continued 
    // through VER_OFFSET_SIZE chars till the end of the buffer
    if p.pBufferStop == p.pBufferEnd { // end of file
    
        p.fStop = 1;
        return;
    }
    fmt.Printf("Ver_StreamSkipSymbol() failed to parse the file \"%s\".\n", p.pFileName);
}

/**Function*************************************************************

  Synopsis    [Skips all symbols until encountering one from the list.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ver_StreamSkipToChars(p *Ver_Stream_t, pCharsToStop string) {
    var pChar, pTemp string;
    assert(!p.fStop, "!p.fStop");
    assert(pCharsToStop != nil, "pCharsToStop != nil");
    // check if the new data should to be loaded
    if p.pBufferCur > p.pBufferStop {
        Ver_StreamReload( p ); }
    // skip the symbols
    for  pChar = p.pBufferCur; pChar < p.pBufferEnd; pChar++  {
        // skip symbols as long as they are NOT in the list
        for  pTemp = pCharsToStop; *pTemp; pTemp++  {
            if *pChar == *pTemp {
                break; }
        }
        if *pTemp == 0 { // pChar is not found in the list
        
            // count the lines
            if *pChar == '\n' {
                p.nLineCounter++; }
            continue;
        }
        // the symbol is found - move position and return
        p.pBufferCur = pChar;
        return;
    }
    // the file is finished or the last part continued 
    // through VER_OFFSET_SIZE chars till the end of the buffer
    if p.pBufferStop == p.pBufferEnd { // end of file
    
        p.fStop = 1;
        return;
    }
    fmt.Printf("Ver_StreamSkipToSymbol() failed to parse the file \"%s\".\n", p.pFileName);
}

/**Function*************************************************************

  Synopsis    [Returns current word delimited by the set of symbols.]

  Description [Modifies the stream by inserting 0 at the first encounter
  of one of the symbols in the list.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ver_StreamGetWord(p *Ver_Stream_t, pCharsToStop string) string {
    var pChar, pTemp string;
    if p.fStop {
        return nil; }
    assert(pCharsToStop != nil, "pCharsToStop != nil");
    // check if the new data should to be loaded
    if p.pBufferCur > p.pBufferStop {
        Ver_StreamReload( p ); }
    // skip the symbols
    p.nChars = 0;
    for  pChar = p.pBufferCur; pChar < p.pBufferEnd; pChar++  {
        // skip symbols as long as they are NOT in the list
        for  pTemp = pCharsToStop; *pTemp; pTemp++  {
            if *pChar == *pTemp {
                break; }
        }
        if *pTemp == 0 { // pChar is not found in the list
        
            p.pChars[p.nChars++] = *pChar;
            if p.nChars == VER_WORD_SIZE {
                fmt.Printf("Ver_StreamGetWord(): The buffer size is exceeded.\n");
                return nil;
            }
            // count the lines
            if *pChar == '\n' {
                p.nLineCounter++; }
            continue;
        }
        // the symbol is found - move the position, set the word end, return the word
        p.pBufferCur = pChar;
        p.pChars[p.nChars] = 0;
        return p.pChars;
    }
    // the file is finished or the last part continued 
    // through VER_OFFSET_SIZE chars till the end of the buffer
    if p.pBufferStop == p.pBufferEnd { // end of file
    
        p.fStop = 1;
        p.pChars[p.nChars] = 0;
        return p.pChars;
    }
    fmt.Printf("Ver_StreamGetWord() failed to parse the file \"%s\".\n", p.pFileName);
    return nil;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ver_StreamMove(p *Ver_Stream_t) {
    if !strncmp(p.pBufferCur+1, "z_g_", 4) || !strncmp(p.pBufferCur+1, "co_g", 3) {
        for  p.pBufferCur[0] != '('  {
            p.pBufferCur++;
        } }
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

