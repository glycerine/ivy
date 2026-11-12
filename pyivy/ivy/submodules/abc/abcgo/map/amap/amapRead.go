package amap

import (
	"fmt"
	"strings"
)

import "os"

/**CFile****************************************************************

  FileName    [amapRead.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Technology mapper for standard cells.]

  Synopsis    []

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: amapRead.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "amapInt.h"
//#include "base/io/ioAbc.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

const AMAP_STRING_GATE =       "GATE"
const AMAP_STRING_PIN =        "PIN"
const AMAP_STRING_NONINV =     "NONINV"
const AMAP_STRING_INV =        "INV"
const AMAP_STRING_UNKNOWN =    "UNKNOWN"

// these symbols (and no other) can appear in the formulas
#define AMAP_SYMB_AND    '*'
#define AMAP_SYMB_AND2   '&'
#define AMAP_SYMB_OR1    '+'
#define AMAP_SYMB_OR2    '|'
#define AMAP_SYMB_XOR    '^'
#define AMAP_SYMB_NOT    '!'
#define AMAP_SYMB_AFTNOT '\''
#define AMAP_SYMB_OPEN   '('
#define AMAP_SYMB_CLOSE  ')'

 type Amap_PinPhase_t int32
const ( 
    AMAP_PHASE_UNKNOWN Amap_PinPhase_t = iota
 
    AMAP_PHASE_INV
 
    AMAP_PHASE_NONINV 

)

 

func Amap_ParseGateAlloc(p *Aig_MmFlex_t, nPins int32) *Amap_Gat_t { return (Amap_Gat_t *)Aig_MmFlexEntryFetch( p, sizeof(Amap_Gat_t)+sizeof(Amap_Pin_t)*nPins ); }
func Amap_ParseStrsav(p *Aig_MmFlex_t, pStr string) string { if pStr {
    return strcpy(Aig_MmFlexEntryFetch(p, len(pStr)+1), pStr);
} else {
    return nil;
} }

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Loads the file into temporary buffer.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Amap_LoadFile(pFileName string) string {
//    extern FILE * Io_FileOpen( const char * FileName, const char * PathVar, const char * Mode, int fVerbose );
    var pFile *os.File;
    var pBuffer string;
    var nFileSize int32;
    var RetValue int32;
    // open the BLIF file for binary reading
    pFile = Io_FileOpen( pFileName, "open_path", "rb", 1 );
//    pFile = fopen( FileName, "rb" );
    // if we got this far, file should be okay otherwise would
    // have been detected by caller
    if pFile == nil {
        fmt.Printf("Cannot open file \"%s\".\n", pFileName);
        return nil;
    }
    assert(pFile != nil, "pFile != nil");
    // get the file size, in bytes
    fseek( pFile, 0, SEEK_END );  
    nFileSize = ftell( pFile );  
    // move the file current reading position to the beginning
    rewind( pFile ); 
    // load the contents of the file into memory
    // inlining of macro ABC_ALLOC
    pBuffer = ((char *) malloc(sizeof(char) * (size_t)(nFileSize + 10)));
    RetValue = fread( pBuffer, nFileSize, 1, pFile );
    // terminate the string with '\0'
    pBuffer[ nFileSize ] = '\0';
    pBuffer += "\n.end\n";
    // close file
    _ = (pFile).Close();
    return pBuffer;
}

/**Function*************************************************************

  Synopsis    [Eliminates comments from the input file.]

  Description [As a byproduct, this procedure also counts the number
  lines and dot-statements in the input file. This also joins non-comment 
  lines that are joined with a backspace '\']
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Amap_RemoveComments(pBuffer string, pnDots *int32, pnLines *int32) {
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

/**Function*************************************************************

  Synopsis    [Splits the stream into tokens.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Amap_DeriveTokens(pBuffer string) *Vec_Ptr_t {
    var vTokens *Vec_Ptr_t;
    var pToken string;
    vTokens = Vec_PtrAlloc( 1000 );
    pToken = strtok( pBuffer, " =\t\r\n" );
    for  pToken  {
        Vec_PtrPush( vTokens, pToken );
        pToken = strtok( nil, " =\t\r\n" );
        // skip latches
        if pToken && int32(strings.Compare(pToken, "LATCH")) == 0 {
            for  pToken && int32(strings.Compare(pToken, "GATE")) != 0  {
                pToken = strtok( nil, " =\t\r\n" );
            } }
    }
    return vTokens;
}

/**Function*************************************************************

  Synopsis    [Finds the number of pins.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Amap_ParseCountPins(vTokens *Vec_Ptr_t, iPos int32) int32 {
    var pToken string;
    var i, Counter int32;
    // inlining of macro Vec_PtrForEachEntryStart
    for  i = iPos; (i < Vec_PtrSize(vTokens)) && (((pToken) = (char *)Vec_PtrEntry(vTokens, i)), 1); i++  {
        if !(int32(strings.Compare(pToken, AMAP_STRING_PIN)) != 0) {
            Counter++;
        } else if !(int32(strings.Compare(pToken, AMAP_STRING_GATE)) != 0) {
            return Counter; }
    }
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Collect the pin names used in the formula.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Amap_GateCollectNames(pMem *Aig_MmFlex_t, pForm string, pPinNames []string) int32 {
    var Buffer [1000]byte;
    var pTemp string;
    var nPins, i int32;
    // save the formula as it was
    strcpy( Buffer, pForm );
    // remove the non-name symbols
    for  pTemp = Buffer; *pTemp; pTemp++  {
        if *pTemp == AMAP_SYMB_AND || *pTemp == AMAP_SYMB_OR1 || *pTemp == AMAP_SYMB_OR2 
          || *pTemp == AMAP_SYMB_XOR || *pTemp == AMAP_SYMB_NOT || *pTemp == AMAP_SYMB_OPEN 
          || *pTemp == AMAP_SYMB_CLOSE || *pTemp == AMAP_SYMB_AFTNOT || *pTemp == AMAP_SYMB_AND2 {
            *pTemp = ' '; }
    }
    // save the names
    nPins = 0;
    pTemp = strtok( Buffer, " " );
    for  pTemp  {
        for  i = 0; i < nPins; i++  {
            if int32(strings.Compare(pTemp, pPinNames[i])) == 0 {
                break; }
        }
        if i == nPins { // cannot find this name; save it
            pPinNames[nPins++] = Amap_ParseStrsav( pMem, pTemp );
        }
        // get the next name
        pTemp = strtok( nil, " " );
    }
    return nPins;
}

/**Function*************************************************************

  Synopsis    [Creates a duplicate gate with pins specified.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Amap_ParseGateWithSamePins(p *Amap_Gat_t) *Amap_Gat_t {
    var pGate *Amap_Gat_t;
    var pPin *Amap_Pin_t;
    var pPinNames [128]string;
    var nPinNames int32;
    assert(p.nPins == 1 && !(int32(strings.Compare(p.Pins.pName, "*")) != 0), "p.nPins == 1 && !strcmp( p.Pins.pName, \"*\" )");
    nPinNames = Amap_GateCollectNames( p.pLib.pMemGates, p.pForm, pPinNames );
    pGate = Amap_ParseGateAlloc( p.pLib.pMemGates, nPinNames );
    *pGate = *p;
    pGate.nPins = nPinNames;
    // inlining of macro Amap_GateForEachPin
    for  pPin = pGate.Pins; pPin < pGate.Pins + pGate.nPins; pPin++  {
        *pPin = *p.Pins;
        pPin.pName = pPinNames[pPin - pGate.Pins];
    }
    return pGate;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Amap_CollectFormulaTokens(vTokens *Vec_Ptr_t, pToken string, iPos int32) int32 {
    var pNext, pPrev string;
    pPrev = pToken + len(pToken);
    for  *(pPrev-1) != ';'  {
        *pPrev++ = ' ';
        pNext = (char *)Vec_PtrEntry(vTokens, iPos++);
        for  *pNext  {
            *pPrev++ = *pNext++;
        }
    }
    *(pPrev-1) = 0;
    return iPos;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Amap_ParseTokens(vTokens *Vec_Ptr_t, fVerbose int32) *Amap_Lib_t {
    var p *Amap_Lib_t;
    var pGate, pPrev *Amap_Gat_t;
    var pPin *Amap_Pin_t;
    var pToken string;
    var pMoGate string = nil;
    var i, nPins, iPos, Count int32;
    p = Amap_LibAlloc();
    pToken = (char *)Vec_PtrEntry(vTokens, iPos++);
    for 
    {
        if int32(strings.Compare(pToken, AMAP_STRING_GATE)) != 0 {
            Amap_LibFree( p );
            fmt.Printf("The first line should begin with %s.\n", AMAP_STRING_GATE);
            return nil;
        }
        // start gate
        nPins = Amap_ParseCountPins( vTokens, iPos );
        pGate = Amap_ParseGateAlloc( p.pMemGates, nPins );
        memset( pGate, 0, sizeof(Amap_Gat_t) );
        pGate.Id = Vec_PtrSize( p.vGates );
        Vec_PtrPush( p.vGates, pGate );
        pGate.pLib = p;
        pGate.nPins = nPins;
        // read gate
        pToken = (char *)Vec_PtrEntry(vTokens, iPos++);
        pGate.pName = Amap_ParseStrsav( p.pMemGates, pToken );    
        pToken = (char *)Vec_PtrEntry(vTokens, iPos++);
        pGate.dArea = atof( pToken );
        pToken = (char *)Vec_PtrEntry(vTokens, iPos++);
        pGate.pOutName = Amap_ParseStrsav( p.pMemGates, pToken ); 
        pToken = (char *)Vec_PtrEntry(vTokens, iPos++);
        iPos = Amap_CollectFormulaTokens( vTokens, pToken, iPos );
        pGate.pForm = Amap_ParseStrsav( p.pMemGates, pToken ); 
        // read pins
        // inlining of macro Amap_GateForEachPin
        for  pPin = pGate.Pins; pPin < pGate.Pins + pGate.nPins; pPin++  {
            pToken = (char *)Vec_PtrEntry(vTokens, iPos++);
            if int32(strings.Compare(pToken, AMAP_STRING_PIN)) != 0 {
                Amap_LibFree( p );
                fmt.Printf("Cannot parse gate %s.\n", pGate.pName);
                return nil;
            }
            // read pin
            pToken = (char *)Vec_PtrEntry(vTokens, iPos++);
            pPin.pName = Amap_ParseStrsav( p.pMemGates, pToken );   
            pToken = (char *)Vec_PtrEntry(vTokens, iPos++);
            if int32(strings.Compare(pToken, AMAP_STRING_UNKNOWN)) == 0 {
                pPin.Phase = AMAP_PHASE_UNKNOWN;
            } else if int32(strings.Compare(pToken, AMAP_STRING_INV)) == 0 {
                pPin.Phase = AMAP_PHASE_INV;
            } else if int32(strings.Compare(pToken, AMAP_STRING_NONINV)) == 0 {
                pPin.Phase = AMAP_PHASE_NONINV;
            } else {
                Amap_LibFree( p );
                fmt.Printf("Cannot read phase of pin %s of gate %s\n", pPin.pName, pGate.pName);
                return nil;
            }
            pToken = (char *)Vec_PtrEntry(vTokens, iPos++);
            pPin.dLoadInput = atof( pToken );
            pToken = (char *)Vec_PtrEntry(vTokens, iPos++);
            pPin.dLoadMax = atof( pToken );
            pToken = (char *)Vec_PtrEntry(vTokens, iPos++);
            pPin.dDelayBlockRise = atof( pToken );
            pToken = (char *)Vec_PtrEntry(vTokens, iPos++);
            pPin.dDelayFanoutRise = atof( pToken );
            pToken = (char *)Vec_PtrEntry(vTokens, iPos++);
            pPin.dDelayBlockFall = atof( pToken );
            pToken = (char *)Vec_PtrEntry(vTokens, iPos++);
            pPin.dDelayFanoutFall = atof( pToken );
            if pPin.dDelayBlockRise > pPin.dDelayBlockFall {
                pPin.dDelayBlockMax = pPin.dDelayBlockRise;
            } else {
                pPin.dDelayBlockMax = pPin.dDelayBlockFall; }
        }
        // fix the situation when all pins are represented as one
        if pGate.nPins == 1 && !(int32(strings.Compare(pGate.Pins.pName, "*")) != 0) {
            pGate = Amap_ParseGateWithSamePins( pGate );
            Vec_PtrPop( p.vGates );
            Vec_PtrPush( p.vGates, pGate );
        }
        pToken = (char *)Vec_PtrEntry(vTokens, iPos++);
//printf( "Finished reading gate %s (%s)\n", pGate->pName, pGate->pOutName );
    
if !( (int32(strings.Compare(pToken, ".end")) != 0) ) { break }
}
    

    // check if there are gates with identical names
    pPrev = nil;
    // inlining of macro Amap_LibForEachGate
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vGates)) && (((pGate) = (Amap_Gat_t *)Vec_PtrEntry(p.vGates, i)), 1); i++  {
        if pPrev && !(int32(strings.Compare(pPrev.pName, pGate.pName)) != 0) {
            pPrev.pTwin = pGate, pGate.pTwin = pPrev;
//            printf( "Warning: Detected multi-output gate \"%s\".\n", pGate->pName );
            if pMoGate == nil {
                pMoGate = pGate.pName; }
            Count++;
        }
        pPrev = pGate;
    }
    if Count != 0 {
        fmt.Printf("Warning: Detected %d multi-output gates (for example, \"%s\").\n", Count, pMoGate); }
    return p;
}

/**Function*************************************************************

  Synopsis    [Reads the library from the input file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Amap_LibReadBuffer(pBuffer string, fVerbose int32) *Amap_Lib_t {
    var pLib *Amap_Lib_t;
    var vTokens *Vec_Ptr_t;
    Amap_RemoveComments( pBuffer, nil, nil );
    vTokens = Amap_DeriveTokens( pBuffer );
    pLib = Amap_ParseTokens( vTokens, fVerbose );
    if pLib == nil {
        Vec_PtrFree( vTokens );
        return nil;
    }
    Vec_PtrFree( vTokens );
    return pLib;
}

/**Function*************************************************************

  Synopsis    [Reads the library from the input file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Amap_LibReadFile(pFileName string, fVerbose int32) *Amap_Lib_t {
    var pLib *Amap_Lib_t;
    var pBuffer string;
    pBuffer = Amap_LoadFile( pFileName );
    if pBuffer == nil {
        return nil; }
    pLib = Amap_LibReadBuffer( pBuffer, fVerbose );
    if pLib != nil {
        pLib.pName = Abc_UtilStrsav( pFileName ); }
    // inlining of macro ABC_FREE
    ((pBuffer) ? (free((char *) (pBuffer)), (pBuffer) = 0) : 0);
    return pLib;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

