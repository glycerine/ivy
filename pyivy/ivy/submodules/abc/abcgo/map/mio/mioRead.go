package mio

import (
	"fmt"
	"strings"
)

import "os"

/**CFile****************************************************************

  FileName    [mioRead.c]

  PackageName [MVSIS 1.3: Multi-valued logic synthesis system.]

  Synopsis    [File reading/writing for technology mapping.]

  Author      [MVSIS Group]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - September 8, 2003.]

  Revision    [$Id: mioRead.c,v 1.9 2004/10/19 06:40:16 satrajit Exp $]

***********************************************************************/

//#include <ctype.h>
//#include "mioInt.h"
//#include "base/io/ioAbc.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

// static Mio_Library_t * Mio_LibraryReadOne( char * FileName, int fExtendedFormat, st__table * tExcludeGate, int fVerbose );

       // Mio_Library_t * Mio_LibraryReadBuffer( char * pBuffer, int fExtendedFormat, st__table * tExcludeGate, int fVerbose );

// static int             Mio_LibraryReadInternal( Mio_Library_t * pLib, char * pBuffer, int fExtendedFormat, st__table * tExcludeGate, int fVerbose );

// static Mio_Gate_t *    Mio_LibraryReadGate( char ** ppToken, int fExtendedFormat );

// static Mio_Pin_t *     Mio_LibraryReadPin( char ** ppToken, int fExtendedFormat );

// static char *          chomp( char *s );

// static void            Mio_LibraryDetectSpecialGates( Mio_Library_t * pLib );

// static void            Io_ReadFileRemoveComments( char * pBuffer, int * pnDots, int * pnLines );


/**Function*************************************************************

  Synopsis    [Read the genlib type of library.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mio_LibraryRead(FileName string, pBuffer string, ExcludeFile string, fVerbose int32) *Mio_Library_t {
    var pLib *Mio_Library_t;
    var num int32;
    var pBufferCopy string;

    var tExcludeGate *st__table;

    if ExcludeFile {
        tExcludeGate = st__init_table(st__strcmp, st__strhash);
        if (num = Mio_LibraryReadExclude( ExcludeFile, tExcludeGate )) == -1 {
            st__free_table( tExcludeGate );
            tExcludeGate = 0;
            return 0;
        }
        fmt.Fprintf(os.Stdout, "Read %d gates from exclude file\n", num);
    }

    pBufferCopy = Abc_UtilStrsav(pBuffer);
    if pBuffer == nil {
        pLib = Mio_LibraryReadOne( FileName, 0, tExcludeGate, fVerbose );       // try normal format first ..
    } else {
        pLib = Mio_LibraryReadBuffer( pBuffer, 0, tExcludeGate, fVerbose );       // try normal format first ..
        if pLib != nil {
            pLib.pName = Abc_UtilStrsav( Extra_FileNameGenericAppend(FileName, ".genlib") ); }
    }
    if pLib == nil {
        if pBuffer == nil {
            pLib = Mio_LibraryReadOne( FileName, 1, tExcludeGate, fVerbose );       // try normal format first ..
        } else {
            pLib = Mio_LibraryReadBuffer( pBufferCopy, 1, tExcludeGate, fVerbose );       // try normal format first ..
            if pLib != nil {
                pLib.pName = Abc_UtilStrsav( Extra_FileNameGenericAppend(FileName, ".genlib") ); }
        }
        if pLib != nil {
            fmt.Printf("Warning: Read extended genlib format but ignoring extensions\n"); }
    }
    pBufferCopy = nil;
    if tExcludeGate != nil {
        st__free_table( tExcludeGate ); }

    return pLib;
}

/**Function*************************************************************

  Synopsis    [Read contents of the file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mio_ReadFile(FileName string, fAddEnd int32) string {
    var pBuffer string;
    var pFile *os.File;
    var nFileSize int32;
    var RetValue int32;

    // open the BLIF file for binary reading
    pFile = Io_FileOpen( FileName, "open_path", "rb", 1 );
//    pFile = fopen( FileName, "rb" );
    // if we got this far, file should be okay otherwise would
    // have been detected by caller
    assert(pFile != nil, "pFile != nil");
    // get the file size, in bytes
    fseek( pFile, 0, SEEK_END );  
    nFileSize = ftell( pFile );  
    // move the file current reading position to the beginning
    rewind( pFile ); 
    // load the contents of the file into memory
    // inlining of macro ABC_ALLOC
    pBuffer   = ((char *) malloc(sizeof(char) * (size_t)(nFileSize + 10)));
    RetValue = fread( pBuffer, nFileSize, 1, pFile );
    // terminate the string with '\0'
    pBuffer[ nFileSize ] = '\0';
    if fAddEnd != 0 {
        pBuffer += "\n.end\n"; }
    // close file
    _ = (pFile).Close();
    return pBuffer;
}

/**Function*************************************************************

  Synopsis    [Read the genlib type of library.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mio_LibraryReadBuffer(pBuffer string, fExtendedFormat int32, tExcludeGate *st__table, fVerbose int32) *Mio_Library_t {
    

    // allocate the genlib structure
    // inlining of macro ABC_CALLOC
    pLib := &Mio_Library_t{};
    pLib.tName2Gate = st__init_table(st__strcmp, st__strhash);
    pLib.pMmFlex = Mem_FlexStart();
    pLib.vCube = Vec_StrAlloc( 100 );

    Io_ReadFileRemoveComments( pBuffer, nil, nil );

    // parse the contents of the file
    if Mio_LibraryReadInternal( pLib, pBuffer, fExtendedFormat, tExcludeGate, fVerbose ) != 0 {
        Mio_LibraryDelete( pLib );
        return nil;
    }

    // derive the functinality of gates
    if Mio_LibraryParseFormulas( pLib ) {
        fmt.Printf("Mio_LibraryRead: Had problems parsing formulas.\n");
        Mio_LibraryDelete( pLib );
        return nil;
    }

    // detect INV and NAND2
    Mio_LibraryDetectSpecialGates( pLib );
//Mio_WriteLibrary( stdout, pLib );
    return pLib;
}

/**Function*************************************************************

  Synopsis    [Read the genlib type of library.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mio_LibraryReadOne(FileName string, fExtendedFormat int32, tExcludeGate *st__table, fVerbose int32) *Mio_Library_t {
    var pLib *Mio_Library_t;
    var pBuffer string;
    // read the file and clean comments
    // pBuffer = Io_ReadFileFileContents( FileName, NULL );
    // we don't use above function but actually do the same thing explicitly
    // to handle open_path expansion correctly
    pBuffer = Mio_ReadFile( FileName, 1 );
    if pBuffer == nil {
        return nil; }
    pLib = Mio_LibraryReadBuffer( pBuffer, fExtendedFormat, tExcludeGate, fVerbose );
    pBuffer = nil;
    if pLib != nil {
        pLib.pName = Abc_UtilStrsav( FileName ); }
    return pLib;
}

/**Function*************************************************************

  Synopsis    [Read the genlib type of library.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mio_LibraryReadInternal(pLib *Mio_Library_t, pBuffer string, fExtendedFormat int32, tExcludeGate *st__table, fVerbose int32) int32 {
    var pGate *Mio_Gate_t;
    var ppGate **Mio_Gate_t;
    var pToken string;
    var nGates int32;
    var nDel int32;

    // start the linked list of gates
    pLib.pGates = nil;
    ppGate = &pLib.pGates;

    // read gates one by one
    pToken = strtok( pBuffer, " \t\r\n" );
    for  pToken && (int32(strings.Compare(pToken, MIO_STRING_GATE)) == 0 || int32(strings.Compare(pToken, MIO_STRING_LATCH)) == 0)  {
        // skip latches
        if int32(strings.Compare(pToken, MIO_STRING_LATCH)) == 0 {
            for  pToken && int32(strings.Compare(pToken, MIO_STRING_GATE)) != 0 && int32(strings.Compare(pToken, ".end")) != 0  {
                if int32(strings.Compare(pToken, MIO_STRING_LATCH)) == 0 {
                    pToken = strtok( nil, " \t\r\n" );
                    fmt.Printf("Skipping latch \"%s\"...\n", pToken);
                    continue;
                }
                pToken = strtok( nil, " \t\r\n" );
            }
            if !(pToken && int32(strings.Compare(pToken, MIO_STRING_GATE)) == 0) {
                break; }
        }

        // derive the next gate
        pGate = Mio_LibraryReadGate( &pToken, fExtendedFormat );
        if pGate == nil {
            return 1; }

        // skip the gate if its formula has problems
        if !Mio_ParseCheckFormula(pGate, pGate.pForm) {
            Mio_GateDelete( pGate );
            continue;
        }
       
        // set the library
        pGate.pLib = pLib;

        // printf ("Processing: '%s'\n", pGate->pName);

        // inlining of macro st__is_member
        if tExcludeGate && st__lookup(tExcludeGate,pGate.pName,(char **) 0) {
            //printf ("Excluding: '%s'\n", pGate->pName);
            Mio_GateDelete( pGate );
            nDel++;
         
        } else {
            // add this gate to the list
            *ppGate = pGate;
            ppGate  = &pGate.pNext;
            nGates++;

            // remember this gate by name
            // inlining of macro st__is_member
            if ! st__lookup(pLib.tName2Gate,pGate.pName,(char **) 0) {
                st__insert( pLib.tName2Gate, pGate.pName, (char *)pGate );
            } else {
                var pBase *Mio_Gate_t  = Mio_LibraryReadGateByName( pLib, pGate.pName, nil );
                if pBase.pTwin != nil {
                    fmt.Printf("Gates with more than 2 outputs are not supported.\n");
                    continue;
                }
                pBase.pTwin = pGate;
                pGate.pTwin = pBase;
//                printf( "Gate \"%s\" appears two times. Creating a 2-output gate.\n", pGate->pName );
            }
        }
    }

    if nGates == 0 {
        fmt.Printf("The library contains no gates.\n");
        return 1;
    }

    // check what is the last word read
    if pToken && int32(strings.Compare(pToken, ".end")) != 0 {
        return 1; }

    if nDel != 0 { 
        fmt.Printf("Actually excluded %d cells\n", nDel); }

    return 0;
}

/**Function*************************************************************

  Synopsis    [Read the genlib type of gate.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mio_LibraryCleanStr(p string) string {
    var i, k int32;
    var whitespace_state int32;
    var pRes string  = Abc_UtilStrsav( p );
    for  i = k = 0; pRes[i]; i++  {
        if pRes[i] != ' ' && pRes[i] != '\t' && pRes[i] != '\r' && pRes[i] != '\n' {
            if pRes[i] != '(' && pRes[i] != ')' && pRes[i] != '+' && pRes[i] != '*' && pRes[i] != '|' && pRes[i] != '&' && pRes[i] != '^' && pRes[i] != '\'' && pRes[i] != '!' {
                if whitespace_state == 2 {
                    pRes[k++] = ' '; }
                whitespace_state = 1;
             
            } else {
                whitespace_state = 0; }
            pRes[k++] = pRes[i];
         
        } else {
            if whitespace_state != 0 {
                whitespace_state = 2;
            } else {
                whitespace_state = 0;
            } }
    }
    pRes[k] = 0;
    return pRes;
}

func Mio_LibraryReadGate(ppToken *string, fExtendedFormat int32) *Mio_Gate_t {
    
    var pPin *Mio_Pin_t;
    var ppPin **Mio_Pin_t;
    var pToken string  = *ppToken;

    // allocate the gate structure
    // inlining of macro ABC_CALLOC
    pGate := &Mio_Gate_t{};
    pGate.Cell = -1;

    // read the name
    pToken = strtok( nil, " \t\r\n" );
    pGate.pName = Abc_UtilStrsav( pToken );

    // read the area
    pToken = strtok( nil, " \t\r\n" );
    pGate.dArea = atof( pToken );

    // read the formula

    // first the output name
    pToken = strtok( nil, "=" );
    pGate.pOutName = chomp( pToken );

    // then rest of the expression 
    pToken = strtok( nil, ";" );
//    pGate->pForm = Mio_LibraryCleanStr( pToken );
    pGate.pForm = Abc_UtilStrsav( pToken );

    // read the pin info
    // start the linked list of pins
    pGate.pPins = nil;
    ppPin = &pGate.pPins;

    // read gates one by one
    pToken = strtok( nil, " \t\r\n" );
    for  pToken && int32(strings.Compare(pToken, MIO_STRING_PIN)) == 0  {
        // derive the next gate
        pPin = Mio_LibraryReadPin( &pToken, fExtendedFormat );
        if pPin == nil {
            Mio_GateDelete( pGate );
            *ppToken = pToken;
            return nil;
        }
        // add this pin to the list
        *ppPin = pPin;
        ppPin  = &pPin.pNext;
        // get the next token
        pToken = strtok( nil, " \t\r\n" );
    }

    *ppToken = pToken;
    return pGate;
}



/**Function*************************************************************

  Synopsis    [Read the genlib type of pin.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mio_LibraryReadPin(ppToken *string, fExtendedFormat int32) *Mio_Pin_t {
    
    var pToken string  = *ppToken;

    // allocate the gate structure
    // inlining of macro ABC_CALLOC
    pPin := &Mio_Pin_t{};

    // read the name
    pToken = strtok( nil, " \t\r\n" );
    pPin.pName = Abc_UtilStrsav( pToken );

    // read the pin phase
    pToken = strtok( nil, " \t\r\n" );
    if int32(strings.Compare(pToken, MIO_STRING_UNKNOWN)) == 0 {
        pPin.Phase = MIO_PHASE_UNKNOWN;
    } else if int32(strings.Compare(pToken, MIO_STRING_INV)) == 0 {
        pPin.Phase = MIO_PHASE_INV;
    } else if int32(strings.Compare(pToken, MIO_STRING_NONINV)) == 0 {
        pPin.Phase = MIO_PHASE_NONINV;
    } else {
        fmt.Printf("Cannot read pin phase specification\n");
        Mio_PinDelete( pPin );
        *ppToken = pToken;
        return nil;
    }

    pToken = strtok( nil, " \t\r\n" );
    pPin.dLoadInput = atof( pToken );

    pToken = strtok( nil, " \t\r\n" );
    pPin.dLoadMax = atof( pToken );

    pToken = strtok( nil, " \t\r\n" );
    pPin.dDelayBlockRise = atof( pToken );

    pToken = strtok( nil, " \t\r\n" );
    pPin.dDelayFanoutRise = atof( pToken );

    pToken = strtok( nil, " \t\r\n" );
    pPin.dDelayBlockFall = atof( pToken );

    pToken = strtok( nil, " \t\r\n" );
    pPin.dDelayFanoutFall = atof( pToken );

    if fExtendedFormat != 0 {
        /* In extended format, the field after dDelayFanoutRise
         * is to be ignored
         **/

        pPin.dDelayBlockFall  = pPin.dDelayFanoutFall;

        pToken = strtok( nil, " \t" );
        pPin.dDelayFanoutFall = atof( pToken );

        /* last field is ignored */
        pToken = strtok( nil, " \t\r\n" );
    }

    if pPin.dDelayBlockRise > pPin.dDelayBlockFall {
        pPin.dDelayBlockMax = pPin.dDelayBlockRise;
    } else {
        pPin.dDelayBlockMax = pPin.dDelayBlockFall; }

    *ppToken = pToken;
    return pPin;
}


/**Function*************************************************************

  Synopsis    [Duplicates string and returns it with leading and 
               trailing spaces removed.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func chomp(s string) string {
    var a, b, c string;
    // remove leading spaces
    for  b = s; *b; b++  {
        if !isspace(*b) {
            break; }
    }
    // strsav the string
    // inlining of macro ABC_ALLOC
    a = strcpy( ((char *) malloc(sizeof(char) * (size_t)(len(b)+1))), b );
    // remove trailing spaces
    for  c = a+len(a); c > a; c--  {
        if *c == 0 || isspace(*c) {
            *c = 0;
        } else {
            break; }
    }
    return a;
}   
        
/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mio_LibraryCompareGatesByArea(pp1 **Mio_Gate_t, pp2 **Mio_Gate_t) int32 {
    var Diff float64  = (*pp1).dArea - (*pp2).dArea;
    if Diff < 0.0 {
        return -1; }
    if Diff > 0.0 { 
        return 1; }
    return 0; 
}
        
/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mio_LibraryCompareGatesByName(pp1 **Mio_Gate_t, pp2 **Mio_Gate_t) int32 {
    var Diff int32  = int32(strings.Compare((*pp1).pName, (*pp2).pName));
    if Diff < 0 {
        return -1; }
    if Diff > 0 { 
        return 1; }
    return 0; 
}
        
/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mio_LibrarySortGates(pLib *Mio_Library_t) {
    var ppGates **Mio_Gate_t;
    var pGate *Mio_Gate_t;
    var i int32;
    // inlining of macro ABC_ALLOC
    ppGates = ((Mio_Gate_t * *) malloc(sizeof(Mio_Gate_t *) * (size_t)(pLib.nGates)));
    // inlining of macro Mio_LibraryForEachGate
    for  pGate = Mio_LibraryReadGates(pLib);                             (pGate != nil);                                                         pGate = Mio_GateReadNext(pGate)  {
        pGate.Cell = i;
        ppGates[i++] = pGate;
    }
    assert(i == pLib.nGates, "i == pLib.nGates");
    // sort gates by name
    // inlining of macro ABC_ALLOC
    pLib.ppGates0 = ((Mio_Gate_t * *) malloc(sizeof(Mio_Gate_t *) * (size_t)(pLib.nGates)));
    for  i = 0; i < pLib.nGates; i++  {
        pLib.ppGates0[i] = ppGates[i];
    }
    qsort( (void *)ppGates, (size_t)pLib.nGates, sizeof(void *), 
                (int (*)(void *, void *)) Mio_LibraryCompareGatesByName );
    for  i = 0; i < pLib.nGates; i++  {
        ppGates[i].pNext = (i < pLib.nGates-1)? ppGates[i+1] : nil;
    }
    pLib.pGates = ppGates[0];
    pLib.ppGatesName = ppGates;
}
        
/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mio_GateCompare(pThis *Mio_Gate_t, pNew *Mio_Gate_t, uTruth word) *Mio_Gate_t {
    if pNew.uTruth != uTruth {
        return pThis; }
    if pThis == nil {
        return pNew; }
    if pThis.dArea > pNew.dArea || (pThis.dArea == pNew.dArea && int32(strings.Compare(pThis.pName, pNew.pName)) > 0) {
        return pNew; }
    return pThis;
}
func Mio_LibraryDetectSpecialGates(pLib *Mio_Library_t) {
    var pGate *Mio_Gate_t;
    var uFuncBuf, uFuncInv, uFuncNand2, uFuncAnd2, uFuncNor2, uFuncOr2 word;

    Mio_LibrarySortGates( pLib );

    uFuncBuf   = ABC_CONST(0xAAAAAAAAAAAAAAAA);
    uFuncAnd2  = ABC_CONST(0xAAAAAAAAAAAAAAAA) & ABC_CONST(0xCCCCCCCCCCCCCCCC);
    uFuncOr2   = ABC_CONST(0xAAAAAAAAAAAAAAAA) | ABC_CONST(0xCCCCCCCCCCCCCCCC);
    uFuncInv   = ~uFuncBuf;
    uFuncNand2 = ~uFuncAnd2;
    uFuncNor2  = ~uFuncOr2;

    // get smallest-area buffer
    // inlining of macro Mio_LibraryForEachGate
    for  pGate = Mio_LibraryReadGates(pLib);                             (pGate != nil);                                                         pGate = Mio_GateReadNext(pGate)  {
        pLib.pGateBuf = Mio_GateCompare( pLib.pGateBuf, pGate, uFuncBuf );
    }
    if pLib.pGateBuf == nil {
        fmt.Printf("Warnings: genlib library reader cannot detect the buffer gate.\n");
        fmt.Printf("Some parts of the supergate-based technology mapper may not work correctly.\n");
    }
 
    // get smallest-area inverter
    // inlining of macro Mio_LibraryForEachGate
    for  pGate = Mio_LibraryReadGates(pLib);                             (pGate != nil);                                                         pGate = Mio_GateReadNext(pGate)  {
        pLib.pGateInv = Mio_GateCompare( pLib.pGateInv, pGate, uFuncInv );
    }
    if pLib.pGateInv == nil {
        fmt.Printf("Warnings: genlib library reader cannot detect the invertor gate.\n");
        fmt.Printf("Some parts of the supergate-based technology mapper may not work correctly.\n");
    }

    // get smallest-area NAND2/AND2 gates
    // inlining of macro Mio_LibraryForEachGate
    for  pGate = Mio_LibraryReadGates(pLib);                             (pGate != nil);                                                         pGate = Mio_GateReadNext(pGate)  {
        pLib.pGateNand2 = Mio_GateCompare( pLib.pGateNand2, pGate, uFuncNand2 );
        pLib.pGateAnd2 = Mio_GateCompare( pLib.pGateAnd2, pGate, uFuncAnd2 );
        pLib.pGateNor2 = Mio_GateCompare( pLib.pGateNor2, pGate, uFuncNor2 );
        pLib.pGateOr2 = Mio_GateCompare( pLib.pGateOr2, pGate, uFuncOr2 );
    }
    if pLib.pGateAnd2 == nil && pLib.pGateNand2 == nil && pLib.pGateNor2 == nil && pLib.pGateOr2 == nil {
        fmt.Printf("Warnings: genlib library reader cannot detect the AND2, NAND2, OR2, and NOR2 gate.\n");
        fmt.Printf("Some parts of the supergate-based technology mapper may not work correctly.\n");
    }
}

/**Function*************************************************************

  Synopsis    [populate hash table of gates to be exlcuded from genlib]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mio_LibraryReadExclude(ExcludeFile string, tExcludeGate *st__table) int32 {
    var nDel int32;
    var pEx *os.File;
    var buffer [128]byte;

    assert((tExcludeGate != nil), "tExcludeGate");

    if ExcludeFile {
        pEx = os.Open(ExcludeFile);

        if pEx == nil {
            fmt.Fprintf(os.Stdout, "Error: Could not open exclude file %s. Stop.\n", ExcludeFile);
            return -1;
        }

        for 1 == fscanf( pEx, "%127s", buffer ) {
            //printf ("Read: '%s'\n", buffer );
            st__insert( tExcludeGate, Abc_UtilStrsav( buffer ), (char *)0 );
            nDel++;
        }

        _ = (pEx).Close();
    }

    return nDel;
}
    
/**Function*************************************************************

  Synopsis    [Eliminates comments from the input file.]

  Description [As a byproduct, this procedure also counts the number
  lines and dot-statements in the input file. This also joins non-comment 
  lines that are joined with a backspace '\']
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_ReadFileRemoveComments(pBuffer string, pnDots *int32, pnLines *int32) {
    var pCur string;
    var nDots, nLines int32;
    // scan through the buffer and eliminate comments
    // (in the BLIF file, comments are lines starting with "#")
    nDots = nLines = 0;
    for  pCur = pBuffer; *pCur; pCur++  {
        // if this is the beginning of comment
        // clean it with spaces until the new line statement
        if *pCur == '#' {
            for  *pCur != '\n'  {
                *pCur++ = ' ';
            } }
    
        // count the number of new lines and dots
        if *pCur == '\n' {
        if *(pCur-1)=='\r' {
        // DOS(R) file support
        if *(pCur-2)!='\\' { nLines++;
        } else {
            // rewind to backslash and overwrite with a space
            *(pCur-2) = ' ';
            *(pCur-1) = ' ';
            *pCur = ' ';
        }
        } else {
        // UNIX(TM) file support
        if *(pCur-1)!='\\' { nLines++;
        } else {
            // rewind to backslash and overwrite with a space
            *(pCur-1) = ' ';
            *pCur = ' ';
        }
        }
    
        } else if *pCur == '.' {
            nDots++; }
    }
    if pnDots != nil {
        *pnDots = nDots; } 
    if pnLines != nil {
        *pnLines = nLines; } 
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END
