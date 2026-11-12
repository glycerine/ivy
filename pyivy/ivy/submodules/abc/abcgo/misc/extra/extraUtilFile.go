package extra

import (
	"fmt"
	"strings"
)

import "os"

/**CFile****************************************************************

  FileName    [extraUtilFile.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [extra]

  Synopsis    [File management utilities.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: extraUtilFile.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "extra.h"

//ABC_NAMESPACE_IMPL_START


/*---------------------------------------------------------------------------*/
/* Constant declarations                                                     */
/*---------------------------------------------------------------------------*/

/*---------------------------------------------------------------------------*/
/* Stucture declarations                                                     */
/*---------------------------------------------------------------------------*/

/*---------------------------------------------------------------------------*/
/* Type declarations                                                         */
/*---------------------------------------------------------------------------*/

/*---------------------------------------------------------------------------*/
/* Variable declarations                                                     */
/*---------------------------------------------------------------------------*/

/*---------------------------------------------------------------------------*/
/* Macro declarations                                                        */
/*---------------------------------------------------------------------------*/


/**AutomaticStart*************************************************************/

/*---------------------------------------------------------------------------*/
/* Static function prototypes                                                */
/*---------------------------------------------------------------------------*/

/**AutomaticEnd***************************************************************/


/*---------------------------------------------------------------------------*/
/* Definition of exported functions                                          */
/*---------------------------------------------------------------------------*/

/**Function*************************************************************

  Synopsis    [Tries to find a file name with a different extension.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Extra_FileGetSimilarName(pFileNameWrong string, pS1 string, pS2 string, pS3 string, pS4 string, pS5 string) string {
    var pFile *os.File;
    var pFileNameOther string;
    var pFileGen string;

    if pS1 == nil {
        return nil; }

    // get the generic file name
    pFileGen = Extra_FileNameGeneric( pFileNameWrong );
    pFileNameOther = Extra_FileNameAppend( pFileGen, pS1 );
    pFile = os.Open(pFileNameOther);
    if pFile == nil && pS2 { // try one more
        pFileNameOther = Extra_FileNameAppend( pFileGen, pS2 );
        pFile = os.Open(pFileNameOther);
        if pFile == nil && pS3 { // try one more
            pFileNameOther = Extra_FileNameAppend( pFileGen, pS3 );
            pFile = os.Open(pFileNameOther);
            if pFile == nil && pS4 { // try one more
                pFileNameOther = Extra_FileNameAppend( pFileGen, pS4 );
                pFile = os.Open(pFileNameOther);
                if pFile == nil && pS5 { // try one more
                    pFileNameOther = Extra_FileNameAppend( pFileGen, pS5 );
                    pFile = os.Open(pFileNameOther);
                }
            }
        }
    }
    // inlining of macro ABC_FREE
    ((pFileGen) ? (free((char *) (pFileGen)), (pFileGen) = 0) : 0);
    if pFile != nil {
        _ = (pFile).Close();
        return pFileNameOther;
    }
    // did not find :(
    return nil;            
}

/**Function*************************************************************

  Synopsis    [Returns the pointer to the file extension.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Extra_FileNameExtension(FileName string) string {
    var pDot string;
    // find the last "dot" in the file name, if it is present
    for  pDot = FileName + len(FileName)-1; pDot >= FileName; pDot--  {
        if *pDot == '.' {
            return pDot + 1; }
    }
   return FileName;
}

/**Function*************************************************************

  Synopsis    [Returns the composite name of the file.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Extra_FileNameAppend_Buffer string // was [500]byte in C

func Extra_FileNameAppend(pBase string, pSuffix string) string {
    
    assert(len(pBase) + len(pSuffix) < 500, "strlen(pBase) + strlen(pSuffix) < 500");
    spStatic_Extra_FileNameAppend_Buffer = fmt.Sprintf("%s%s", pBase, pSuffix);
    return spStatic_Extra_FileNameAppend_Buffer;
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Extra_FileNameGeneric(FileName string) string {
    var pDot, pRes string;
    pRes = Extra_UtilStrsav( FileName );
    if pDot = strrchr( pRes, '.' ) {
        *pDot = 0; }
    return pRes;
}

/**Function*************************************************************

  Synopsis    [Returns the composite name of the file.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Extra_FileNameGenericAppend_Buffer [1000]byte;
func Extra_FileNameGenericAppend(pBase string, pSuffix string) string {
    
    var pDot string;
    assert(len(pBase) + len(pSuffix) < 1000, "strlen(pBase) + strlen(pSuffix) < 1000");
    strcpy( spStatic_Extra_FileNameGenericAppend_Buffer, pBase );
    if pDot = strrchr( spStatic_Extra_FileNameGenericAppend_Buffer, '.' ) {
        *pDot = 0; }
    strcat( spStatic_Extra_FileNameGenericAppend_Buffer, pSuffix );
    return spStatic_Extra_FileNameGenericAppend_Buffer;
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Extra_FileNameCorrectPath(FileName string) {
    var pStart string;
    if FileName {
        for  pStart = FileName; *pStart; pStart++  {
            if *pStart == '>' || *pStart == '\\' {
                *pStart = '/'; }
        } }
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Extra_FileNameWithoutPath(FileName string) string {
    var pRes string;
    for  pRes = FileName + len(FileName) - 1; pRes >= FileName; pRes--  {
        if *pRes == '\\' || *pRes == '/' {
            return pRes + 1; }
    }
    return FileName;
}
func Extra_FilePathWithoutName(FileName string) string {
    var pRes string;
    FileName = Abc_UtilStrsav( FileName );
    for  pRes = FileName + len(FileName) - 1; pRes >= FileName; pRes--  {
        if *pRes == '\\' || *pRes == '/' {
           *pRes = 0;
           Extra_FileNameCorrectPath( FileName );
           return FileName;
        }
    }
    // inlining of macro ABC_FREE
    ((FileName) ? (free((char *) (FileName)), (FileName) = 0) : 0);
    return nil;
}
func Extra_FileDesignName(pFileName string) string {
    var pBeg, pEnd, pStore, pCur string;
    // find the first dot
    for  pEnd = pFileName; *pEnd; pEnd++  {
        if *pEnd == '.' {
            break; }
    }
    // find the first char
    for  pBeg = pEnd - 1; pBeg >= pFileName; pBeg--  {
        if !((*pBeg >= 'a' && *pBeg <= 'z') || (*pBeg >= 'A' && *pBeg <= 'Z') || (*pBeg >= '0' && *pBeg <= '9') || *pBeg == '_') {
            break; }
    }
    pBeg++;
    // fill up storage
    // inlining of macro ABC_ALLOC
    pStore = ((char *) malloc(sizeof(char) * (size_t)(pEnd - pBeg + 1)));
    for  pCur = pStore; pBeg < pEnd; pBeg, pCur = pBeg+1, pCur+1  {
        *pCur = *pBeg;
    }
    *pCur = 0;
    return pStore;
}

/**Function*************************************************************

  Synopsis    [Returns the file size.]

  Description [The file should be closed.]

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Extra_FileCheck(pFileName string) int32 {
    var pFile *os.File;
    pFile = os.Open(pFileName);
    if pFile == nil {
        fmt.Printf("Extra_FileCheck():  File \"%s\" does not exist.\n", pFileName);
        return 0;
    }
    fseek( pFile, 0, SEEK_END );
    if ftell( pFile ) == 0 {
        fmt.Printf("Extra_FileCheck():  File \"%s\" is empty.\n", pFileName); }
    _ = (pFile).Close();
    return 1;
}

/**Function*************************************************************

  Synopsis    [Returns the file size.]

  Description [The file should be closed.]

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Extra_FileSize(pFileName string) int32 {
    var pFile *os.File;
    var nFileSize int32;
    pFile = os.Open(pFileName);
    if pFile == nil {
        fmt.Printf("Extra_FileSize(): The file is unavailable (absent or open).\n");
        return 0;
    }
    fseek( pFile, 0, SEEK_END );  
    nFileSize = ftell( pFile ); 
    _ = (pFile).Close();
    return nFileSize;
}


/**Function*************************************************************

  Synopsis    [Read the file into the internal buffer.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Extra_FileRead(pFile *os.File) string {
    var nFileSize int32;
    var pBuffer string;
    var RetValue int32;
    // get the file size, in bytes
    fseek( pFile, 0, SEEK_END );  
    nFileSize = ftell( pFile );  
    // move the file current reading position to the beginning
    rewind( pFile ); 
    // load the contents of the file into memory
    // inlining of macro ABC_ALLOC
    pBuffer = ((char *) malloc(sizeof(char) * (size_t)(nFileSize + 3)));
    RetValue = fread( pBuffer, nFileSize, 1, pFile );
    // terminate the string with '\0'
    pBuffer[ nFileSize + 0] = '\n';
    pBuffer[ nFileSize + 1] = '\0';
    return pBuffer;
}
func Extra_FileRead2(pFile *os.File, pFile2 *os.File) string {
    var pBuffer string;
    var nSize, nSize2 int32;
    var RetValue int32;
    // get the file size, in bytes
    fseek( pFile, 0, SEEK_END );  
    nSize = ftell( pFile );  
    rewind( pFile ); 
    // get the file size, in bytes
    fseek( pFile2, 0, SEEK_END );  
    nSize2 = ftell( pFile2 );  
    rewind( pFile2 ); 
    // load the contents of the file into memory
    // inlining of macro ABC_ALLOC
    pBuffer = ((char *) malloc(sizeof(char) * (size_t)(nSize + nSize2 + 3)));
    RetValue = fread( pBuffer,         nSize,  1, pFile );
    RetValue = fread( pBuffer + nSize, nSize2, 1, pFile2 );
    // terminate the string with '\0'
    pBuffer[ nSize + nSize2 + 0] = '\n';
    pBuffer[ nSize + nSize2 + 1] = '\0';
    return pBuffer;
}

/**Function*************************************************************

  Synopsis    [Read the file into the internal buffer.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Extra_FileReadContents(pFileName string) string {
    var pFile *os.File;
    var pBuffer string;
    pFile = os.Open(pFileName);
    if pFile != nil {
        pBuffer = Extra_FileRead( pFile );
    } else {
        pBuffer = nil;
    }
    if pFile != nil {  _ = (pFile).Close(); }
    return pBuffer;
}
func Extra_FileReadContents2(pFileName string, pFileName2 string) string {
    var pFile, pFile2 *os.File;
    var pBuffer string;
    pFile  = os.Open(pFileName);
    pFile2 = os.Open(pFileName2);
    if (pFile != nil) && (pFile2 != nil) {
        pBuffer = Extra_FileRead2( pFile, pFile2 );
    } else {
        pBuffer = nil;
    }
    if pFile != nil {  _ = (pFile).Close(); }
    if pFile2 != nil { _ = (pFile2).Close(); }
    return pBuffer;
}

/**Function*************************************************************

  Synopsis    [Returns one if the file has a given extension.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Extra_FileIsType(pFileName string, pS1 string, pS2 string, pS3 string) int32 {
    var lenS int32;
    var lenF int32 = len(pFileName);
    if pS1 {
        lenS = len(pS1);
    } else {
        lenS = 0;
    }
    if (lenS != 0) && lenF > lenS && !strncmp( pFileName+lenF-lenS, pS1, lenS ) {
        return 1; }
    if pS2 {
        lenS = len(pS2);
    } else {
        lenS = 0;
    }
    if (lenS != 0) && lenF > lenS && !strncmp( pFileName+lenF-lenS, pS2, lenS ) {
        return 1; }
    if pS3 {
        lenS = len(pS3);
    } else {
        lenS = 0;
    }
    if (lenS != 0) && lenF > lenS && !strncmp( pFileName+lenF-lenS, pS3, lenS ) {
        return 1; }
    return 0;
}

/**Function*************************************************************

  Synopsis    [Returns the time stamp.]

  Description [The file should be closed.]

  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Extra_TimeStamp_Buffer [100]byte;
func Extra_TimeStamp() string {
    
    var TimeStamp string;
    var ltime int64;
    // get the current time
    time( &ltime );
    TimeStamp = asctime( localtime( &ltime ) );
    TimeStamp[ len(TimeStamp) - 1 ] = 0;
    strcpy( spStatic_Extra_TimeStamp_Buffer, TimeStamp );
    return spStatic_Extra_TimeStamp_Buffer;
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Extra_ReadBinary(Buffer string) uint32 {
    var Result uint32;
    var i int32;

    Result = 0;
    for  i = 0; Buffer[i]; i++  {
        if Buffer[i] == '0' || Buffer[i] == '1' {
            Result = Result * 2 + Buffer[i] - '0';
        } else {
            assert(false, "0");
        }
    }
    return Result;
}

/**Function*************************************************************

  Synopsis    [Prints the bit string.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Extra_PrintBinary(pFile *os.File, Sign []uint32, nBits int32) {
    var i int32;
    for  i = nBits-1; i >= 0; i--  {
        fmt.Fprintf(pFile, "%c", '0' + Abc_InfoHasBit(Sign, i));
    }
//    fprintf( pFile, "\n" );
}
func Extra_PrintBinary2(pFile *os.File, Sign []uint32, nBits int32) {
    var i int32;
    for  i = 0; i < nBits; i++  {
        fmt.Fprintf(pFile, "%c", '0' + Abc_InfoHasBit(Sign, i));
    }
//    fprintf( pFile, "\n" );
}

/**Function*************************************************************

  Synopsis    [Reads the hex unsigned into the bit-string.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Extra_ReadHex(Sign []uint32, pString string, nDigits int32) int32 {
    var Digit, k, c int32;
    for  k = 0; k < nDigits; k++  {
        c = nDigits-1-k;
        if pString[c] >= '0' && pString[c] <= '9' {
            Digit = pString[c] - '0';
        } else if pString[c] >= 'A' && pString[c] <= 'F' {
            Digit = pString[c] - 'A' + 10;
        } else if pString[c] >= 'a' && pString[c] <= 'f' {
            Digit = pString[c] - 'a' + 10;
        } else { assert(false, "0"); return 0; }
        Sign[k/8] |= ( (Digit & 15) << ((k%8) * 4) );
    }
    return 1;
}
func Extra_ReadHexadecimal(Sign []uint32, pString string, nVars int32) int32 {
    var nWords, nDigits, k int32;
    nWords = Extra_TruthWordNum( nVars );
    for  k = 0; k < nWords; k++  {
        Sign[k] = 0;
    }
    // read the number from the string
    nDigits = (1 << nVars) / 4;
    if nDigits == 0 {
        nDigits = 1; }
    Extra_ReadHex( Sign, pString, nDigits );
    return 1;
}

/**Function*************************************************************

  Synopsis    [Prints the hex unsigned into a file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Extra_PrintHexadecimal(pFile *os.File, Sign []uint32, nVars int32) {
    var nDigits, Digit, k int32;
    // write the number into the file
    nDigits = (1 << nVars) / 4;
    for  k = nDigits - 1; k >= 0; k--  {
        Digit = ((Sign[k/8] >> ((k%8) * 4)) & 15);
        if Digit < 10 {
            fmt.Fprintf(pFile, "%d", Digit);
        } else {
            fmt.Fprintf(pFile, "%c", 'a' + Digit-10); }
    }
//    fprintf( pFile, "\n" );
}

/**Function*************************************************************

  Synopsis    [Prints the hex unsigned into a file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Extra_PrintHexadecimalString(pString string, Sign []uint32, nVars int32) {
    var nDigits, Digit, k int32;
    if nVars == 0 && !(Sign[0] & 1) { pString = fmt.Sprintf("0"); return; } // const0
    if nVars == 0 &&  (Sign[0] & 1) { pString = fmt.Sprintf("1"); return; } // const1
    if nVars == 1 &&  (Sign[0] & 1) { pString = fmt.Sprintf("1"); return; } // inverter
    if nVars == 1 && !(Sign[0] & 1) { pString = fmt.Sprintf("2"); return; } // buffer
    // write the number into the file
    nDigits = (1 << nVars) / 4;
    for  k = nDigits - 1; k >= 0; k--  {
        Digit = ((Sign[k/8] >> ((k%8) * 4)) & 15);
        if Digit < 10 {
            *pString++ = '0' + Digit;
        } else {
            *pString++ = 'a' + Digit-10; }
    }
//    fprintf( pFile, "\n" );
    *pString = 0;
}

/**Function*************************************************************

  Synopsis    [Prints the hex unsigned into a file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Extra_PrintHex(pFile *os.File, pTruth *uint32, nVars int32) {
    var nMints, nDigits, Digit, k int32;

    // write the number into the file
    fmt.Fprintf(pFile, "0x");
    nMints  = (1 << nVars);
    nDigits = nMints / 4 + ((nMints % 4) > 0);
    for  k = nDigits - 1; k >= 0; k--  {
        Digit = ((pTruth[k/8] >> (k * 4)) & 15);
        if Digit < 10 {
            fmt.Fprintf(pFile, "%d", Digit);
        } else {
            fmt.Fprintf(pFile, "%c", 'A' + Digit-10); }
    }
//    fprintf( pFile, "\n" );
}
func Extra_PrintHex2(pFile *os.File, pTruth *uint32, nVars int32) {
    var nMints, nDigits, Digit, k int32;

    // write the number into the file
    //fprintf( pFile, "0x" );
    nMints  = (1 << nVars);
    nDigits = nMints / 4 + ((nMints % 4) > 0);
    for  k = nDigits - 1; k >= 0; k--  {
        Digit = ((pTruth[k/8] >> (k * 4)) & 15);
        if Digit < 10 {
            fmt.Fprintf(pFile, "%d", Digit);
        } else {
            fmt.Fprintf(pFile, "%c", 'A' + Digit-10); }
    }
//    fprintf( pFile, "\n" );
}
func Extra_PrintHexReverse(pFile *os.File, pTruth *uint32, nVars int32) {
    var nMints, nDigits, Digit, k int32;

    // write the number into the file
    fmt.Fprintf(pFile, "0x");
    nMints  = (1 << nVars);
    nDigits = nMints / 4 + ((nMints % 4) > 0);
    for  k = 0; k < nDigits; k++  {
        Digit = ((pTruth[k/8] >> (k * 4)) & 15);
        if Digit < 10 {
            fmt.Fprintf(pFile, "%d", Digit);
        } else {
            fmt.Fprintf(pFile, "%c", 'A' + Digit-10); }
    }
//    fprintf( pFile, "\n" );
}

/**Function*************************************************************

  Synopsis    [Returns the composite name of the file.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Extra_PrintSymbols(pFile *os.File, Char byte, nTimes int32, fPrintNewLine int32) {
    var i int32;
    for  i = 0; i < nTimes; i++  {
        fmt.Printf("%c", Char);
    }
    if fPrintNewLine != 0 {
        fmt.Printf("\n"); }
}

/**Function*************************************************************

  Synopsis    [Appends the string.]

  Description [Assumes that the given string (pStrGiven) has been allocated
  before using malloc(). The additional string has not been allocated.
  Allocs more root, appends the additional part, frees the old given string.]

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Extra_StringAppend(pStrGiven string, pStrAdd string) string {
    var pTemp string;
    if pStrGiven {
        // inlining of macro ABC_ALLOC
        pTemp = ((char *) malloc(sizeof(char) * (size_t)(len(pStrGiven) + len(pStrAdd) + 2)));
        pTemp = fmt.Sprintf("%s%s", pStrGiven, pStrAdd);
        // inlining of macro ABC_FREE
        ((pStrGiven) ? (free((char *) (pStrGiven)), (pStrGiven) = 0) : 0);
    
    } else {
        pTemp = Extra_UtilStrsav( pStrAdd ); }
    return pTemp;
}

/**Function*************************************************************

  Synopsis    [Only keep characters belonging to the second string.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Extra_StringClean(pStrGiven string, pCharKeep string) {
    var pTemp string;
    var pChar string;
    var pSave string = pStrGiven;
    for  pTemp = pStrGiven; *pTemp; pTemp++  {
        for  pChar = pCharKeep; *pChar; pChar++  {
            if *pTemp == *pChar {
                break; }
        }
        if *pChar == 0 {
            continue; }
        *pSave++ = *pTemp;
    }
    *pSave = 0;
}

/**Function*************************************************************

  Synopsis    [String comparison procedure.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Extra_StringCompare(pp1 string, pp2 string) int32 {
    return int32(strings.Compare(*(char **)pp1, *(char **)pp2));
}

/**Function*************************************************************

  Synopsis    [Sorts lines in the file alphabetically.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Extra_FileSort(pFileName string, pFileNameOut string) {
    var pFile *os.File;
    var pContents string;
    var pLines *string;
    var i, nLines, Begin int32;
    pFile = os.Open(pFileName);
    if pFile == nil {
        fmt.Printf("Extra_FileSort(): Cannot open file \"%s\".\n", pFileName);
        return;
    }
    pContents = Extra_FileRead( pFile );
    _ = (pFile).Close();
    if pContents == nil {
        fmt.Printf("Extra_FileSort(): Cannot read contents of file \"%s\".\n", pFileName);
        return;
    }
    // count end of lines
    for  nLines, i = 0, 0; pContents[i]; i++  {
        nLines += (pContents[i] == '\n');
    }
    // break the file into lines
    pLines = (char **)malloc( sizeof(char *) * nLines );
    Begin = 0;
    for  nLines, i = 0, 0; pContents[i]; i++  {
        if pContents[i] == '\n' {
            pContents[i] = 0;
            pLines[nLines++] = pContents + Begin;
            Begin = i + 1;
        }
    }
    // sort the lines
    qsort( pLines, (size_t)nLines, sizeof(char *), (int(*)(void *,void *))Extra_StringCompare );
    // write a new file
    pFile = os.OpenFile(pFileNameOut, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    for  i = 0; i < nLines; i++  {
        if pLines[i][0] {
            fmt.Fprintf(pFile, "%s\n", pLines[i]); }
    }
    _ = (pFile).Close();
    // cleanup
    free( pLines );
    free( pContents );
    // report the result
    fmt.Printf("The file after sorting is \"%s\".\n", pFileNameOut);
}


/**Function*************************************************************

  Synopsis    [Appends line number in the end.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Extra_FileLineNumAdd(pFileName string, pFileNameOut string) {
    var Buffer [1000]byte;
    var pFile *os.File;
    var pFile2 *os.File;
    var iLine int32;
    pFile = os.Open(pFileName);
    if pFile == nil {
        fmt.Printf("Extra_FileLineNumAdd(): Cannot open file \"%s\".\n", pFileName);
        return;
    }
    pFile2 = os.OpenFile(pFileNameOut, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    if pFile2 == nil {
        _ = (pFile).Close();
        fmt.Printf("Extra_FileLineNumAdd(): Cannot open file \"%s\".\n", pFileNameOut);
        return;
    }
    for  iLine = 0; fgets( Buffer, 1000, pFile ); iLine++  {
        sprintf( Buffer + len(Buffer) - 2, "%03d\n%c", iLine, 0 );
        fputs( Buffer, pFile2 );
    }
    _ = (pFile).Close();
    _ = (pFile2).Close();
    // report the result
    fmt.Printf("The resulting file is \"%s\".\n", pFileNameOut);
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
/*
int main( int argc, char ** argv )
{
    if ( argc == 2 )
        Extra_FileSort( argv[1], Extra_FileNameAppend(argv[1], "_sorted") );
    else
        printf( "%s: Wrong number of command line arguments.\n", argv[0] );
    return 1;
}
*/

/*---------------------------------------------------------------------------*/
/* Definition of internal functions                                          */
/*---------------------------------------------------------------------------*/

/*---------------------------------------------------------------------------*/
/* Definition of static Functions                                            */
/*---------------------------------------------------------------------------*/


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

