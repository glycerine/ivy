package extra

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [extraUtilReader.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [extra]

  Synopsis    [File reading utilities.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: extraUtilReader.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include <stdio.h>
//#include "extra.h"
//#include "misc/vec/vec.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

#define EXTRA_BUFFER_SIZE        4*1048576    // 1M   - size of the data chunk stored in memory
const EXTRA_OFFSET_SIZE =           4096    // 4K   - load new data when less than this is left

// Original macro:
// #define EXTRA_MINIMUM(a,b)       (((a) < (b))? (a) : (b))


type Extra_FileReader_t_ struct {
    // the input file
    pFileName            string;     // the input file name
    pFile            *os.File;         // the input file pointer
    nFileSize              int32;     // the total number of bytes in the file
    nFileRead              int32;     // the number of bytes currently read from file
    // info about processing different types of input chars
    pCharMap [256]byte; // the character map
    // temporary storage for data 
    pBuffer            string;       // the buffer
    nBufferSize              int32;   // the size of the buffer
    pBufferCur            string;    // the current reading position
    pBufferEnd            string;    // the first position not used by currently loaded data
    pBufferStop            string;   // the position where loading new data will be done
    // tokens given to the user
    vTokens       *Vec_Ptr_t;       // the vector of tokens returned to the user
    vLines       *Vec_Int_t;        // the vector of line numbers for each token
    nLineCounter              int32;  // the counter of lines processed
    // status of the parser
    fStop              int32;         // this flag goes high when the end of file is reached
}

// character types
 type Extra_CharType_t int32
const ( 
    EXTRA_CHAR_COMMENT Extra_CharType_t = iota
  // a character that begins the comment
    EXTRA_CHAR_NORMAL
   // a regular character
    EXTRA_CHAR_STOP
     // a character that delimits a series of tokens
    EXTRA_CHAR_CLEAN     // a character that should be cleaned

)

 

// the static functions
// static void * Extra_FileReaderGetTokens_int( Extra_FileReader_t * p );

// static void Extra_FileReaderReload( Extra_FileReader_t * p );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Starts the file reader.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Extra_FileReaderAlloc(pFileName string, pCharsComment string, pCharsStop string, pCharsClean string) *Extra_FileReader_t {
    var p *Extra_FileReader_t;
    var pFile *os.File;
    var pChar string;
    var nCharsToRead int32;
    var RetValue int32;
    // check if the file can be opened
    pFile = os.Open(pFileName);
    if pFile == nil {
        fmt.Printf("Extra_FileReaderAlloc(): Cannot open input file \"%s\".\n", pFileName);
        return nil;
    }
    // start the file reader    
    // inlining of macro ABC_ALLOC
    p = ((Extra_FileReader_t *) malloc(sizeof(Extra_FileReader_t) * (size_t)(1)));
    memset( p, 0, sizeof(Extra_FileReader_t) );
    p.pFileName   = pFileName;
    p.pFile       = pFile;
    // set the character map
    memset( p.pCharMap, EXTRA_CHAR_NORMAL, 256 );
    for  pChar = pCharsComment; *pChar; pChar++  {
        p.pCharMap[(unsigned char)*pChar] = EXTRA_CHAR_COMMENT;
    }
    for  pChar = pCharsStop; *pChar; pChar++  {
        p.pCharMap[(unsigned char)*pChar] = EXTRA_CHAR_STOP;
    }
    for  pChar = pCharsClean; *pChar; pChar++  {
        p.pCharMap[(unsigned char)*pChar] = EXTRA_CHAR_CLEAN;
    }
    // get the file size, in bytes
    fseek( pFile, 0, SEEK_END );  
    p.nFileSize = ftell( pFile );  
    rewind( pFile ); 
    // allocate the buffer
    // inlining of macro ABC_ALLOC
    p.pBuffer = ((char *) malloc(sizeof(char) * (size_t)(EXTRA_BUFFER_SIZE+1)));
    p.nBufferSize = EXTRA_BUFFER_SIZE;
    p.pBufferCur  = p.pBuffer;
    // determine how many chars to read
    // inlining of macro EXTRA_MINIMUM
    
    if (p.nFileSize) < (EXTRA_BUFFER_SIZE) {
        nCharsToRead = (p.nFileSize);
    } else {
        nCharsToRead = (EXTRA_BUFFER_SIZE);
    }
    // load the first part into the buffer
    RetValue = fread( p.pBuffer, nCharsToRead, 1, p.pFile );
    p.nFileRead = nCharsToRead;
    // set the ponters to the end and the stopping point
    p.pBufferEnd  = p.pBuffer + nCharsToRead;
    if p.nFileRead == p.nFileSize {
        p.pBufferStop = p.pBufferEnd;
    } else {
        p.pBufferStop = p.pBuffer + EXTRA_BUFFER_SIZE - EXTRA_OFFSET_SIZE;
    }
    // start the arrays
    p.vTokens = Vec_PtrAlloc( 100 );
    p.vLines = Vec_IntAlloc( 100 );
    p.nLineCounter = 1; // 1-based line counting
    return p;
}

/**Function*************************************************************

  Synopsis    [Stops the file reader.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Extra_FileReaderFree(p *Extra_FileReader_t) {
    if p.pFile {
        _ = (p.pFile).Close(); }
    // inlining of macro ABC_FREE
    ((p.pBuffer) ? (free((char *) (p.pBuffer)), (p.pBuffer) = 0) : 0);
    Vec_PtrFree( p.vTokens );
    Vec_IntFree( p.vLines );
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    [Returns the file size.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Extra_FileReaderGetFileName(p *Extra_FileReader_t) string {
    return p.pFileName;
}

/**Function*************************************************************

  Synopsis    [Returns the file size.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Extra_FileReaderGetFileSize(p *Extra_FileReader_t) int32 {
    return p.nFileSize;
}

/**Function*************************************************************

  Synopsis    [Returns the current reading position.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Extra_FileReaderGetCurPosition(p *Extra_FileReader_t) int32 {
    return p.nFileRead - (p.pBufferEnd - p.pBufferCur);
}

/**Function*************************************************************

  Synopsis    [Returns the line number for the given token.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Extra_FileReaderGetLineNumber(p *Extra_FileReader_t, iToken int32) int32 {
    assert(iToken >= 0 && iToken < p.vTokens.nSize, "iToken >= 0 && iToken < p.vTokens.nSize");
    return p.vLines.pArray[iToken];
}


/**Function*************************************************************

  Synopsis    [Returns the next set of tokens.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Extra_FileReaderGetTokens(p *Extra_FileReader_t) any {
    var vTokens *Vec_Ptr_t;
    for  vTokens = (Vec_Ptr_t *)Extra_FileReaderGetTokens_int( p )  {
        if vTokens.nSize > 0 {
            break; }
    }
    return vTokens;
}

/**Function*************************************************************

  Synopsis    [Returns the next set of tokens.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Extra_FileReaderGetTokens_int(p *Extra_FileReader_t) any {
    var pChar string;
    var fTokenStarted, MapValue int32;
    if p.fStop {
        return nil; }
    // reset the token info
    p.vTokens.nSize = 0;
    p.vLines.nSize = 0;
    fTokenStarted = 0;
    // check if the new data should to be loaded
    if p.pBufferCur > p.pBufferStop {
        Extra_FileReaderReload( p ); }

//    printf( "%d\n", p->pBufferEnd - p->pBufferCur );

    // process the string starting from the current position
    for  pChar = p.pBufferCur; pChar < p.pBufferEnd; pChar++  {
        // count the lines
        if *pChar == '\n' {
            p.nLineCounter++; }
        // switch depending on the character
        MapValue = p.pCharMap[int32(*pChar)];

//        printf( "Char value = %d. Map value = %d.\n", *pChar, MapValue );


        switch  MapValue  {
            case EXTRA_CHAR_COMMENT:
                if *pChar != '/' || *(pChar+1) == '/' { // dealing with the need to have // as a comment
                    // if the token was being written, stop it
                    if fTokenStarted {
                        fTokenStarted = 0; }
                    // eraze the comment till the end of line
                    for  *pChar != '\n'  {
                        *pChar++ = 0;
                        if pChar == p.pBufferEnd {   // this failure is due to the fact the comment continued 
                            // through EXTRA_OFFSET_SIZE chars till the end of the buffer
                            fmt.Printf("Extra_FileReader failed to parse the file \"%s\".\n", p.pFileName);
                            return nil;
                        }
                    }
                    pChar--;
                    break;
                }
                // otherwise it is a normal character
            fallthrough
case EXTRA_CHAR_NORMAL:
                if !fTokenStarted {
                    Vec_PtrPush( p.vTokens, pChar );
                    Vec_IntPush( p.vLines, p.nLineCounter );
                    fTokenStarted = 1;
                }
                break;
            case EXTRA_CHAR_STOP:
                if ( fTokenStarted )
                    fTokenStarted = 0;
                *pChar = 0;
                // prepare before leaving
                p.pBufferCur = pChar + 1;
                return p.vTokens;
            case EXTRA_CHAR_CLEAN:
                if ( fTokenStarted )
                    fTokenStarted = 0;
                *pChar = 0;
                break;
            default:
                assert(false, "0");
        }
    }
    // the file is finished or the last part continued 
    // through EXTRA_OFFSET_SIZE chars till the end of the buffer
    if p.pBufferStop == p.pBufferEnd { // end of file
    
        *pChar = 0;
        p.fStop = 1;
        return p.vTokens;
    }
    fmt.Printf("Extra_FileReader failed to parse the file \"%s\".\n", p.pFileName);
/*
    {
        int i;
        for ( i = 0; i < p->vTokens->nSize; i++ )
            printf( "%s ", p->vTokens->pArray[i] );
        printf( "\n" );
    }
*/
    return nil;
}

/**Function*************************************************************

  Synopsis    [Loads new data into the file reader.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Extra_FileReaderReload(p *Extra_FileReader_t) {
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
    // inlining of macro EXTRA_MINIMUM
    
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
        p.pBufferStop = p.pBuffer + EXTRA_BUFFER_SIZE - EXTRA_OFFSET_SIZE;
    }
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

