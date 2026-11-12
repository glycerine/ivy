package mapper

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

/**CFile****************************************************************

  FileName    [mapperTree.c]

  PackageName [MVSIS 1.3: Multi-valued logic synthesis system.]

  Synopsis    [Generic technology mapping engine.]

  Author      [MVSIS Group]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 2.0. Started - June 1, 2004.]

  Revision    [$Id: mapperTree.c,v 1.9 2005/01/23 06:59:45 alanmi Exp $]

***********************************************************************/

#ifdef __linux__
#include <libgen.h>
#endif

//#include "mapperInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static void      Map_LibraryAddFaninDelays( Map_SuperLib_t * pLib, Map_Super_t * pGate, Map_Super_t * pFanin, Mio_Pin_t * pPin );

// static int       Map_LibraryGetMaxSuperPi_rec( Map_Super_t * pGate );

// static unsigned  Map_LibraryGetGateSupp_rec( Map_Super_t * pGate );


// fanout limits
static int s_MapFanoutLimits[10] = { 1/*0*/, 10/*1*/, 5/*2*/, 2/*3*/, 1/*4*/, 1/*5*/, 1/*6*/ };

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Reads one gate.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Map_LibraryReadGateTree(pLib *Map_SuperLib_t, pBuffer string, Number int32, nVarsMax int32) *Map_Super_t {
    var pGate *Map_Super_t;
    var pTemp string;
    var i, Num int32;

    // start and clean the gate
    pGate = (Map_Super_t *)Extra_MmFixedEntryFetch( pLib.mmSupers );
    memset( pGate, 0, sizeof(Map_Super_t) );

    // set the gate number
    pGate.Num = Number;

    // read the mark
    pTemp = strtok( pBuffer, " " );
    if pTemp[0] == '*' {
        pGate.fSuper = 1;
        pTemp = strtok( nil, " " );
    }

    // read the root gate
    pGate.pRoot = Mio_LibraryReadGateByName( pLib.pGenlib, pTemp, nil );
    if pGate.pRoot == nil {
        fmt.Printf("Cannot read the root gate names %s.\n", pTemp);
        return nil;
    }
    // set the max number of fanouts
    pGate.nFanLimit = s_MapFanoutLimits[ Mio_GateReadPinNum(pGate.pRoot) ];

    // read the pin-to-pin delay
    for  i = 0; ( pTemp = strtok( nil, " \n\0" ) ); i++  {
        if pTemp[0] == '#' {
            break; }
        if i == nVarsMax {
            fmt.Printf("There are too many entries on the line.\n");
            return nil;
        }
        Num = strconv.Atoi(pTemp);
        if Num < 0 {
            fmt.Printf("The number of a child supergate is negative.\n");
            return nil;
        }
        if Num > pLib.nLines {
            fmt.Printf("The number of a child supergate (%d) exceeded the number of lines (%d).\n", Num, pLib.nLines);
            return nil;
        }
        pGate.pFanins[i] = pLib.ppSupers[Num];
    }
    pGate.nFanins = i;
    if pGate.nFanins != uint32(Mio_GateReadPinNum(pGate.pRoot)) {
        fmt.Printf("The number of fanins of a root gate is wrong.\n");
        return nil;
    }

    // save the gate name, just in case
    if pTemp && pTemp[0] == '#' {
        if pTemp[1] == 0 {
            pTemp = strtok( nil, " \n\0" );
        } else { // skip spaces
            for  pTemp++; *pTemp == ' '; pTemp++  {} }
        // save the formula
        pGate.pFormula = Extra_MmFlexEntryFetch( pLib.mmForms, len(pTemp)+1 );
        strcpy( pGate.pFormula, pTemp );
    }
    // check the rest of the string
    pTemp = strtok( nil, " \n\0" );
    if pTemp != nil {
        fmt.Printf("The following trailing symbols found \"%s\".\n", pTemp); }
    return pGate;
}

/**Function*************************************************************

  Synopsis    [Reads the supergate library from file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/*
int Map_LibraryReadFileTree( Map_SuperLib_t * pLib, FILE * pFile, char *pFileName )
{
    ProgressBar * pProgress;
    char pBuffer[5000];
    Map_Super_t * pGate;
    char * pTemp = 0, * pLibName;
    int nCounter, k, i;
    int RetValue;

    // skip empty and comment lines
    while ( fgets( pBuffer, 5000, pFile ) != NULL )
    {
        // skip leading spaces
        for ( pTemp = pBuffer; *pTemp == ' ' || *pTemp == '\r' || *pTemp == '\n'; pTemp++ );
        // skip comment lines and empty lines
        if ( *pTemp != 0 && *pTemp != '#' )
            break;
    }

    pLibName = strtok( pTemp, " \t\r\n" );
    pLib->pGenlib = (Mio_Library_t *)Abc_FrameReadLibGen();
    if ( pLib->pGenlib == NULL || strcmp( Mio_LibraryReadName(pLib->pGenlib), pLibName ) )
    {
        printf( "Supergate library \"%s\" requires the use of genlib library \"%s\".\n", pFileName, pLibName );
        return 0;
    }

    // read the number of variables
    RetValue = fscanf( pFile, "%d\n", &pLib->nVarsMax );
    if ( pLib->nVarsMax < 2 || pLib->nVarsMax > 10 )
    {
        printf( "Suspicious number of variables (%d).\n", pLib->nVarsMax );
        return 0;
    }

    // read the number of gates
    RetValue = fscanf( pFile, "%d\n", &pLib->nSupersReal );
    if ( pLib->nSupersReal < 1 || pLib->nSupersReal > 10000000 )
    {
        printf( "Suspicious number of gates (%d).\n", pLib->nSupersReal );
        return 0;
    }

    // read the number of lines
    RetValue = fscanf( pFile, "%d\n", &pLib->nLines );
    if ( pLib->nLines < 1 || pLib->nLines > 10000000 )
    {
        printf( "Suspicious number of lines (%d).\n", pLib->nLines );
        return 0;
    }

    // allocate room for supergate pointers
    pLib->ppSupers = ABC_ALLOC( Map_Super_t *, pLib->nLines + 10000 );

    // create the elementary supergates
    for ( i = 0; i < pLib->nVarsMax; i++ )
    {
        // get a new gate
        pGate = (Map_Super_t *)Extra_MmFixedEntryFetch( pLib->mmSupers );
        memset( pGate, 0, sizeof(Map_Super_t) );
        // assign the elementary variable, the truth table, and the delays
        pGate->Num = i;
        // set the truth table
        pGate->uTruth[0] = pLib->uTruths[i][0];
        pGate->uTruth[1] = pLib->uTruths[i][1];
        // set the arrival times of all input to non-existent delay
        for ( k = 0; k < pLib->nVarsMax; k++ )
        {
            pGate->tDelaysR[k].Rise = pGate->tDelaysR[k].Fall = MAP_NO_VAR;
            pGate->tDelaysF[k].Rise = pGate->tDelaysF[k].Fall = MAP_NO_VAR;
        }
        // set an existent arrival time for rise and fall
        pGate->tDelaysR[i].Rise = 0.0;
        pGate->tDelaysF[i].Fall = 0.0;
        // set the gate
        pLib->ppSupers[i] = pGate;
    }

    // read the lines
    nCounter = pLib->nVarsMax;
    pProgress = Extra_ProgressBarStart( stdout, pLib->nLines );
    while ( fgets( pBuffer, 5000, pFile ) != NULL )
    {
        for ( pTemp = pBuffer; *pTemp == ' ' || *pTemp == '\r' || *pTemp == '\n'; pTemp++ );
        if ( pTemp[0] == '\0' )
            continue;
//        if ( pTemp[0] == 'a' || pTemp[2] == 'a' )
//        {
//            pLib->nLines--;
//            continue;
//        }

        // get the gate
        pGate = Map_LibraryReadGateTree( pLib, pTemp, nCounter, pLib->nVarsMax );
        if ( pGate == NULL )
        {
            Extra_ProgressBarStop( pProgress );
            return 0;
        }
        pLib->ppSupers[nCounter++] = pGate;
        // later we will derive: truth table, delays, area, number of component gates, etc

        // update the progress bar
        Extra_ProgressBarUpdate( pProgress, nCounter, NULL );
    }
    Extra_ProgressBarStop( pProgress );
    if ( nCounter != pLib->nLines )
        printf( "The number of lines read (%d) is different what the file says (%d).\n", nCounter, pLib->nLines );
    pLib->nSupersAll = nCounter;
    // count the number of real supergates
    nCounter = 0;
    for ( k = 0; k < pLib->nLines; k++ )
        nCounter += pLib->ppSupers[k]->fSuper;
    if ( nCounter != pLib->nSupersReal )
        printf( "The number of gates read (%d) is different what the file says (%d).\n", nCounter, pLib->nSupersReal );
    pLib->nSupersReal = nCounter;
    return 1;
}
int Map_LibraryReadTree2( Map_SuperLib_t * pLib, char * pFileName, char * pExcludeFile )
{
    FILE * pFile;
    int Status, num;
    Abc_Frame_t * pAbc;
    st__table * tExcludeGate = 0;

    // read the beginning of the file
    assert( pLib->pGenlib == NULL );
    pFile = Io_FileOpen( pFileName, "open_path", "r", 1 );
//    pFile = fopen( pFileName, "r" ); 
    if ( pFile == NULL )
    {
        printf( "Cannot open input file \"%s\".\n", pFileName );
        return 0;
    }

    if ( pExcludeFile )
    {
        pAbc = Abc_FrameGetGlobalFrame();
        
        tExcludeGate = st__init_table(st__strcmp, st__strhash);
        if ( (num = Mio_LibraryReadExclude( pExcludeFile, tExcludeGate )) == -1 )
        {
            st__free_table( tExcludeGate );
            tExcludeGate = 0;
            return 0;
        }

        fprintf ( Abc_FrameReadOut( pAbc ), "Read %d gates from exclude file\n", num );
    }
    
    Status = Map_LibraryReadFileTree( pLib, pFile, pFileName );
    fclose( pFile );
    if ( Status == 0 )
        return 0;
    // prepare the info about the library
    return Map_LibraryDeriveGateInfo( pLib, tExcludeGate );
}
*/

/**Function*************************************************************

  Synopsis    [Similar to fgets.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_StrGets(pBuffer string, nBufferSize int32, vStr *Vec_Str_t, pPos *int32) int32 {
    var pCur string;
    var pBeg string  = Vec_StrArray(vStr) + *pPos;
    var pEnd string  = Vec_StrArray(vStr) + Vec_StrSize(vStr);
    assert(nBufferSize > 1, "nBufferSize > 1");
    if pBeg == pEnd {
        *pBuffer = 0;
        return 0;
    }
    assert(pBeg < pEnd, "pBeg < pEnd");
    for  pCur = pBeg; pCur < pEnd; pCur++  {
        *pBuffer++ = *pCur;
        if *pCur == 0 {
            *pPos += pCur - pBeg;
            return 0;
        }
        if *pCur == '\n' {
            *pPos += pCur - pBeg + 1;
            *pBuffer = 0;
            return 1;
        }
        if pCur - pBeg == nBufferSize-1 {
            *pPos += pCur - pBeg + 1;
            *pBuffer = 0;
            return 1;
        }
    }
    return 0;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Map_LibraryCompareLibNames(pName1 string, pName2 string) int32 {
    var p1 string  = Abc_UtilStrsav( pName1 );
    var p2 string  = Abc_UtilStrsav( pName2 );
    var i, RetValue int32;
    for  i = 0; p1[i]; i++  {
        if p1[i] == '>' || p1[i] == '\\' || p1[i] == '/' {
            p1[i] = '/'; }
    }
    for  i = 0; p2[i]; i++  {
        if p2[i] == '>' || p2[i] == '\\' || p2[i] == '/' {
            p2[i] = '/'; }
    }
    RetValue = int32(strings.Compare(p1, p2));
    p1 = nil;
    p2 = nil;
    return RetValue; 
}

/**Function*************************************************************

  Synopsis    [Reads the supergate library from file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Map_LibraryReadFileTreeStr(pLib *Map_SuperLib_t, pGenlib *Mio_Library_t, vStr *Vec_Str_t, pFileName string) int32 {
    var pProgress *ProgressBar;
    var pBuffer [5000]byte;
    var pGate *Map_Super_t;
    var pTemp, pLibName string;
    var nCounter, k, i int32;
    var RetValue, nPos int32;

    // skip empty and comment lines
//    while ( fgets( pBuffer, 5000, pFile ) != NULL )
    for {
        RetValue = Vec_StrGets( pBuffer, 5000, vStr, &nPos );
        if RetValue == 0 {
            return 0; }
        // skip leading spaces
        for  pTemp = pBuffer; *pTemp == ' ' || *pTemp == '\r' || *pTemp == '\n'; pTemp++  {}
        // skip comment lines and empty lines
        if *pTemp != 0 && *pTemp != '#' {
            break; }
    }

    pLibName = strtok( pTemp, " \t\r\n" );
//    pLib->pGenlib = (Mio_Library_t *)Abc_FrameReadLibGen();
    pLib.pGenlib = pGenlib;
//    if ( pLib->pGenlib == NULL || strcmp( , pLibName ) )
    if pLib.pGenlib == nil || (Map_LibraryCompareLibNames(Mio_LibraryReadName(pLib.pGenlib), pLibName) != 0) {
        fmt.Printf("Supergate library \"%s\" requires the use of genlib library \"%s\".\n", pFileName, pLibName);
        return 0;
    }

    // read the number of variables
    RetValue = Vec_StrGets( pBuffer, 5000, vStr, &nPos );
    if RetValue == 0 {
        return 0; }
    RetValue = sscanf( pBuffer, "%d\n", &pLib.nVarsMax );
    if pLib.nVarsMax < 2 || pLib.nVarsMax > 10 {
        fmt.Printf("Suspicious number of variables (%d).\n", pLib.nVarsMax);
        return 0;
    }

    // read the number of gates
    RetValue = Vec_StrGets( pBuffer, 5000, vStr, &nPos );
    if RetValue == 0 {
        return 0; }
    RetValue = sscanf( pBuffer, "%d\n", &pLib.nSupersReal );
    if pLib.nSupersReal < 1 || pLib.nSupersReal > 10000000 {
        fmt.Printf("Suspicious number of gates (%d).\n", pLib.nSupersReal);
        return 0;
    }

    // read the number of lines
    RetValue = Vec_StrGets( pBuffer, 5000, vStr, &nPos );
    if RetValue == 0 {
        return 0; }
    RetValue = sscanf( pBuffer, "%d\n", &pLib.nLines );
    if pLib.nLines < 1 || pLib.nLines > 10000000 {
        fmt.Printf("Suspicious number of lines (%d).\n", pLib.nLines);
        return 0;
    }

    // allocate room for supergate pointers
    // inlining of macro ABC_ALLOC
    pLib.ppSupers = ((Map_Super_t * *) malloc(sizeof(Map_Super_t *) * (size_t)(pLib.nLines + 10000)));

    // create the elementary supergates
    for  i = 0; i < pLib.nVarsMax; i++  {
        // get a new gate
        pGate = (Map_Super_t *)Extra_MmFixedEntryFetch( pLib.mmSupers );
        memset( pGate, 0, sizeof(Map_Super_t) );
        // assign the elementary variable, the truth table, and the delays
        pGate.Num = i;
        // set the truth table
        pGate.uTruth[0] = pLib.uTruths[i][0];
        pGate.uTruth[1] = pLib.uTruths[i][1];
        // set the arrival times of all input to non-existent delay
        for  k = 0; k < pLib.nVarsMax; k++  {
            pGate.tDelaysR[k].Rise = pGate.tDelaysR[k].Fall = MAP_NO_VAR;
            pGate.tDelaysF[k].Rise = pGate.tDelaysF[k].Fall = MAP_NO_VAR;
        }
        // set an existent arrival time for rise and fall
        pGate.tDelaysR[i].Rise = 0.0;
        pGate.tDelaysF[i].Fall = 0.0;
        // set the gate
        pLib.ppSupers[i] = pGate;
    }

    // read the lines
    nCounter = pLib.nVarsMax;
    pProgress = Extra_ProgressBarStart( os.Stdout, pLib.nLines );
//    while ( fgets( pBuffer, 5000, pFile ) != NULL )
    for  Vec_StrGets( pBuffer, 5000, vStr, &nPos ) != 0  {
        for  pTemp = pBuffer; *pTemp == ' ' || *pTemp == '\r' || *pTemp == '\n'; pTemp++  {}
        if pTemp[0] == '\0' {
            continue; }
//        if ( pTemp[0] == 'a' || pTemp[2] == 'a' )
//        {
//            pLib->nLines--;
//            continue;
//        }

        // get the gate
        pGate = Map_LibraryReadGateTree( pLib, pTemp, nCounter, pLib.nVarsMax );
        if pGate == nil {
            Extra_ProgressBarStop( pProgress );
            return 0;
        }
        pLib.ppSupers[nCounter++] = pGate;
        // later we will derive: truth table, delays, area, number of component gates, etc

        // update the progress bar
        Extra_ProgressBarUpdate( pProgress, nCounter, nil );
    }
    Extra_ProgressBarStop( pProgress );
    if nCounter != pLib.nLines {
        fmt.Printf("The number of lines read (%d) is different from what the file says (%d).\n", nCounter, pLib.nLines); }
    pLib.nSupersAll = nCounter;
    // count the number of real supergates
    nCounter = 0;
    for  k = 0; k < pLib.nLines; k++  {
        nCounter += pLib.ppSupers[k].fSuper;
    }
    if nCounter != pLib.nSupersReal {
        fmt.Printf("The number of gates read (%d) is different what the file says (%d).\n", nCounter, pLib.nSupersReal); }
    pLib.nSupersReal = nCounter;
    return 1;
}
func Map_LibraryReadTree(pLib *Map_SuperLib_t, pGenlib *Mio_Library_t, pFileName string, pExcludeFile string) int32 {
    var pBuffer string;
    var vStr *Vec_Str_t;
    var Status, num int32;
    var pAbc *Abc_Frame_t;
    var tExcludeGate *st__table;

    // read the beginning of the file
    assert(pLib.pGenlib == nil, "pLib.pGenlib == nil");
//    pFile = Io_FileOpen( pFileName, "open_path", "r", 1 );
    pBuffer = Mio_ReadFile( pFileName, 0 );
    if pBuffer == nil {
        fmt.Printf("Cannot open input file \"%s\".\n", pFileName);
        return 0;
    }
    vStr = Vec_StrAllocArray( pBuffer, len(pBuffer) );

    if pExcludeFile {
        pAbc = Abc_FrameGetGlobalFrame();
        
        tExcludeGate = st__init_table(st__strcmp, st__strhash);
        if (num = Mio_LibraryReadExclude( pExcludeFile, tExcludeGate )) == -1 {
            st__free_table( tExcludeGate );
            tExcludeGate = 0;
            Vec_StrFree( vStr );
            return 0;
        }

        fmt.Fprintf(Abc_FrameReadOut( pAbc ), "Read %d gates from exclude file\n", num);
    }
    
    Status = Map_LibraryReadFileTreeStr( pLib, pGenlib, vStr, pFileName );
    Vec_StrFree( vStr );
    if Status == 0 {
        return 0; }
    // prepare the info about the library
    return Map_LibraryDeriveGateInfo( pLib, tExcludeGate );
}








/**Function*************************************************************

  Synopsis    [Derives information about the library.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Map_LibraryDeriveGateInfo(pLib *Map_SuperLib_t, tExcludeGate *st__table) int32 {
    var pGate, pFanin *Map_Super_t;
    var pPin *Mio_Pin_t;
    var uCanon [2]uint32;
    var uTruths [6][2]uint32;
    var i, k, nRealVars int32;

    // set all the derivable info related to the supergates
    for  i = pLib.nVarsMax; i < int32(pLib.nLines); i++  {
        pGate = pLib.ppSupers[i];

        if tExcludeGate != nil {
            // inlining of macro st__is_member
            if  st__lookup(tExcludeGate,Mio_GateReadName( pGate.pRoot ),(char **) 0) {
                pGate.fExclude = 1; }
            for  k = 0; k < int32(pGate.nFanins); k++  {
                pFanin = pGate.pFanins[k];
                if pFanin.fExclude {
                    pGate.fExclude = 1;
                    continue;
                }
            }
        }
        
        // collect the truth tables of the fanins
        for  k = 0; k < int32(pGate.nFanins); k++  {
            pFanin = pGate.pFanins[k];
            uTruths[k][0] = pFanin.uTruth[0];
            uTruths[k][1] = pFanin.uTruth[1];
        }
        // derive the new truth table
        Mio_DeriveTruthTable( pGate.pRoot, uTruths, pGate.nFanins, 6, pGate.uTruth );

        // set the initial delays of the supergate
        for  k = 0; k < pLib.nVarsMax; k++  {
            pGate.tDelaysR[k].Rise = pGate.tDelaysR[k].Fall = MAP_NO_VAR;
            pGate.tDelaysF[k].Rise = pGate.tDelaysF[k].Fall = MAP_NO_VAR;
        }
        // get the linked list of pins for the given root gate
        pPin = Mio_GateReadPins( pGate.pRoot );
        // update the initial delay of the supergate using info from the corresponding pin
        for  k = 0; k < int32(pGate.nFanins); k++, pPin = Mio_PinReadNext(pPin)  {
            // if there is no corresponding pin, this is a bug, return fail
            if pPin == nil {
                fmt.Printf("There are less pins than gate inputs.\n");
                return 0;
            }
            // update the delay information of k-th fanins info from the corresponding pin
            Map_LibraryAddFaninDelays( pLib, pGate, pGate.pFanins[k], pPin );
        }
        // if there are some pins left, this is a bug, return fail
        if pPin != nil {
            fmt.Printf("There are more pins than gate inputs.\n");
            return 0;
        }
        // find the max delay
        pGate.tDelayMax.Rise = pGate.tDelayMax.Fall = MAP_NO_VAR;
        for  k = 0; k < pLib.nVarsMax; k++  {
            // the rise of the output depends on the rise and fall of the output
            if pGate.tDelayMax.Rise < pGate.tDelaysR[k].Rise {
                pGate.tDelayMax.Rise = pGate.tDelaysR[k].Rise; }
            if pGate.tDelayMax.Rise < pGate.tDelaysR[k].Fall {
                pGate.tDelayMax.Rise = pGate.tDelaysR[k].Fall; }
            // the fall of the output depends on the rise and fall of the output
            if pGate.tDelayMax.Fall < pGate.tDelaysF[k].Rise {
                pGate.tDelayMax.Fall = pGate.tDelaysF[k].Rise; }
            if pGate.tDelayMax.Fall < pGate.tDelaysF[k].Fall {
                pGate.tDelayMax.Fall = pGate.tDelaysF[k].Fall; }

            // inlining of macro MAP_MAX
            pGate.tDelaysF[k].Worst = (((pGate.tDelaysF[k].Fall) > (pGate.tDelaysF[k].Rise))? (pGate.tDelaysF[k].Fall) : (pGate.tDelaysF[k].Rise));
            // inlining of macro MAP_MAX
            pGate.tDelaysR[k].Worst = (((pGate.tDelaysR[k].Fall) > (pGate.tDelaysR[k].Rise))? (pGate.tDelaysR[k].Fall) : (pGate.tDelaysR[k].Rise));
        }

        // count gates and area of the supergate
        pGate.nGates = 1;
        pGate.Area   = float32(Mio_GateReadArea(pGate.pRoot));
        for  k = 0; k < int32(pGate.nFanins); k++  {
            pGate.nGates += pGate.pFanins[k].nGates;
            pGate.Area   += pGate.pFanins[k].Area;
        }
        // do not add the gate to the table, if this gate is an internal gate
        // of some supegate and does not correspond to a supergate output
        if ( !pGate.fSuper ) || pGate.fExclude {
            continue; }

        // find the maximum index of a variable in the support of the supergates
        // this is important for two reasons:
        // (1) to limit the number of permutations considered for canonicization
        // (2) to get rid of equivalence phases to speed-up matching
        nRealVars = Map_LibraryGetMaxSuperPi_rec( pGate ) + 1;
        assert(nRealVars > 0 && nRealVars <= pLib.nVarsMax, "nRealVars > 0 && nRealVars <= pLib.nVarsMax");
        // if there are some problems with this code, try this instead
//        nRealVars = pLib->nVarsMax;

        // find the N-canonical form of this supergate
        pGate.nPhases = Map_CanonComputeSlow( pLib.uTruths, pLib.nVarsMax, nRealVars, pGate.uTruth, pGate.uPhases, uCanon );
        // add the supergate into the table by its N-canonical table
        Map_SuperTableInsertC( pLib.tTableC, uCanon, pGate );
/*
        {
            int uCanon1, uCanon2;
            uCanon1 = uCanon[0];
            pGate->uTruth[0] = ~pGate->uTruth[0];
            pGate->uTruth[1] = ~pGate->uTruth[1];
            Map_CanonComputeSlow( pLib->uTruths, pLib->nVarsMax, nRealVars, pGate->uTruth, pGate->uPhases, uCanon );
            uCanon2 = uCanon[0];
Rwt_Man5ExploreCount( uCanon1 < uCanon2 ? uCanon1 : uCanon2 );
        }
*/
    }
    // sort the gates in each line
    Map_SuperTableSortSupergatesByDelay( pLib.tTableC, pLib.nSupersAll );

    // let the glory be manifest
//    Map_LibraryPrintTree( pLib );
    return 1;
}

/**Function*************************************************************

  Synopsis    [Finds the largest PI number in the support of the supergate.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Map_LibraryGetMaxSuperPi_rec(pGate *Map_Super_t) int32 {
    var i, VarCur, VarMax int32;
    if pGate.pRoot == nil {
        return pGate.Num; }
    for  i = 0; i < int32(pGate.nFanins); i++  {
        VarCur = Map_LibraryGetMaxSuperPi_rec( pGate.pFanins[i] );
        if VarMax < VarCur {
            VarMax = VarCur; }
    }
    return VarMax;
}

/**Function*************************************************************

  Synopsis    [Finds the largest PI number in the support of the supergate.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Map_LibraryGetGateSupp_rec(pGate *Map_Super_t) uint32 {
    var uSupport uint32;
    var i int32;
    if pGate.pRoot == nil {
        return uint32((1 << (pGate.Num))); }
    uSupport = 0;
    for  i = 0; i < int32(pGate.nFanins); i++  {
        uSupport |= Map_LibraryGetGateSupp_rec( pGate.pFanins[i] );
    }
    return uSupport;
}

/**Function*************************************************************

  Synopsis    [Derives the pin-to-pin delay constraints for the supergate.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Map_LibraryAddFaninDelays(pLib *Map_SuperLib_t, pGate *Map_Super_t, pFanin *Map_Super_t, pPin *Mio_Pin_t) {
    var PinPhase Mio_PinPhase_t;
    var tDelayBlockRise, tDelayBlockFall, tDelayPin float32;
    var fMaxDelay int32;
    var i int32;

    // use this node to enable max-delay model
    if fMaxDelay != 0 {
        var tDelayBlockMax float32;
        // get the maximum delay
        tDelayBlockMax = float32(Mio_PinReadDelayBlockMax(pPin));
        // go through the supergate inputs
        for  i = 0; i < pLib.nVarsMax; i++  {
            if pFanin.tDelaysR[i].Rise < 0 {
                continue; }
            tDelayPin = pFanin.tDelaysR[i].Rise + tDelayBlockMax;
            if pGate.tDelaysR[i].Rise < tDelayPin {
                pGate.tDelaysR[i].Rise = tDelayPin; }
        }
        // go through the supergate inputs
        for  i = 0; i < pLib.nVarsMax; i++  {
            if pFanin.tDelaysF[i].Fall < 0 {
                continue; }
            tDelayPin = pFanin.tDelaysF[i].Fall + tDelayBlockMax;
            if pGate.tDelaysF[i].Fall < tDelayPin {
                pGate.tDelaysF[i].Fall = tDelayPin; }
        }
        return;
    }

    // get the interesting parameters of this pin
    PinPhase = Mio_PinReadPhase(pPin);
    tDelayBlockRise = float32(Mio_PinReadDelayBlockRise( pPin ));  
    tDelayBlockFall = float32(Mio_PinReadDelayBlockFall( pPin ));  

    // update the rise and fall of the output depending on the phase of the pin 
    if PinPhase != MIO_PHASE_INV {  // NONINV phase is present
    
        // the rise of the gate is determined by the rise of the fanin
        // the fall of the gate is determined by the fall of the fanin
        for  i = 0; i < pLib.nVarsMax; i++  {
            ////////////////////////////////////////////////////////
            // consider the rise of the gate
            ////////////////////////////////////////////////////////
            // check two types of constraints on the rise of the fanin:
            // (1) the constraints related to the rise of the PIs
            // (2) the constraints related to the fall of the PIs
            if pFanin.tDelaysR[i].Rise >= 0 { // case (1)
             // fanin's rise depends on the rise of i-th PI
                // update the rise of the gate's output
                if pGate.tDelaysR[i].Rise < pFanin.tDelaysR[i].Rise + tDelayBlockRise {
                    pGate.tDelaysR[i].Rise = pFanin.tDelaysR[i].Rise + tDelayBlockRise; }
            }
            if pFanin.tDelaysR[i].Fall >= 0 { // case (2)
             // fanin's rise depends on the fall of i-th PI
                // update the rise of the gate's output
                if pGate.tDelaysR[i].Fall < pFanin.tDelaysR[i].Fall + tDelayBlockRise {
                    pGate.tDelaysR[i].Fall = pFanin.tDelaysR[i].Fall + tDelayBlockRise; }
            }
            ////////////////////////////////////////////////////////

            ////////////////////////////////////////////////////////
            // consider the fall of the gate (similar)
            ////////////////////////////////////////////////////////
            // check two types of constraints on the fall of the fanin:
            // (1) the constraints related to the rise of the PIs
            // (2) the constraints related to the fall of the PIs
            if pFanin.tDelaysF[i].Rise >= 0 { // case (1) 
             
                if pGate.tDelaysF[i].Rise < pFanin.tDelaysF[i].Rise + tDelayBlockFall {
                    pGate.tDelaysF[i].Rise = pFanin.tDelaysF[i].Rise + tDelayBlockFall; }
            }
            if pFanin.tDelaysF[i].Fall >= 0 { // case (2) 
             
                if pGate.tDelaysF[i].Fall < pFanin.tDelaysF[i].Fall + tDelayBlockFall {
                    pGate.tDelaysF[i].Fall = pFanin.tDelaysF[i].Fall + tDelayBlockFall; }
            }
            ////////////////////////////////////////////////////////
        }
    }
    if PinPhase != MIO_PHASE_NONINV {  // INV phase is present
    
        // the rise of the gate is determined by the fall of the fanin
        // the fall of the gate is determined by the rise of the fanin
        for  i = 0; i < pLib.nVarsMax; i++  {
            ////////////////////////////////////////////////////////
            // consider the rise of the gate's output
            ////////////////////////////////////////////////////////
            // check two types of constraints on the fall of the fanin:
            // (1) the constraints related to the rise of the PIs
            // (2) the constraints related to the fall of the PIs
            if pFanin.tDelaysF[i].Rise >= 0 { // case (1)
             // fanin's rise depends on the rise of i-th PI
                // update the rise of the gate
                if pGate.tDelaysR[i].Rise < pFanin.tDelaysF[i].Rise + tDelayBlockRise {
                    pGate.tDelaysR[i].Rise = pFanin.tDelaysF[i].Rise + tDelayBlockRise; }
            }
            if pFanin.tDelaysF[i].Fall >= 0 { // case (2)
             // fanin's rise depends on the fall of i-th PI
                // update the rise of the gate
                if pGate.tDelaysR[i].Fall < pFanin.tDelaysF[i].Fall + tDelayBlockRise {
                    pGate.tDelaysR[i].Fall = pFanin.tDelaysF[i].Fall + tDelayBlockRise; }
            }
            ////////////////////////////////////////////////////////

            ////////////////////////////////////////////////////////
            // consider the fall of the gate (similar)
            ////////////////////////////////////////////////////////
            // check two types of constraints on the rise of the fanin:
            // (1) the constraints related to the rise of the PIs
            // (2) the constraints related to the fall of the PIs
            if pFanin.tDelaysR[i].Rise >= 0 { // case (1) 
             
                if pGate.tDelaysF[i].Rise < pFanin.tDelaysR[i].Rise + tDelayBlockFall {
                    pGate.tDelaysF[i].Rise = pFanin.tDelaysR[i].Rise + tDelayBlockFall; }
            }
            if pFanin.tDelaysR[i].Fall >= 0 { // case (2) 
             
                if pGate.tDelaysF[i].Fall < pFanin.tDelaysR[i].Fall + tDelayBlockFall {
                    pGate.tDelaysF[i].Fall = pFanin.tDelaysR[i].Fall + tDelayBlockFall; }
            }
            ////////////////////////////////////////////////////////
        }
    }
}


/**Function*************************************************************

  Synopsis    [Performs phase transformation for one function.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Map_CalculatePhase(uTruths [][2]uint32, nVars int32, uTruth uint32, uPhase uint32) uint32 {
    var v, Shift int32;
    for  v, Shift = 0, 1; v < nVars; v++, Shift <<= 1  {
        if uPhase & Shift {
            uTruth = (((uTruth & ~uTruths[v][0]) << Shift) | ((uTruth & uTruths[v][0]) >> Shift)); }
    }
    return uTruth;
}

/**Function*************************************************************

  Synopsis    [Performs phase transformation for one function.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Map_CalculatePhase6(uTruths [][2]uint32, nVars int32, uTruth []uint32, uPhase uint32, uTruthRes []uint32) {
    var uTemp uint32;
    var v, Shift int32;

    // initialize the result
    uTruthRes[0] = uTruth[0];
    uTruthRes[1] = uTruth[1];
    if uPhase == 0 {
        return; }
    // compute the phase 
    for  v, Shift = 0, 1; v < nVars; v++, Shift <<= 1  {
        if uPhase & Shift {
            if Shift < 32 {
                uTruthRes[0] = (((uTruthRes[0] & ~uTruths[v][0]) << Shift) | ((uTruthRes[0] & uTruths[v][0]) >> Shift));
                uTruthRes[1] = (((uTruthRes[1] & ~uTruths[v][1]) << Shift) | ((uTruthRes[1] & uTruths[v][1]) >> Shift));
            
            } else {
                uTemp        = uTruthRes[0];
                uTruthRes[0] = uTruthRes[1];
                uTruthRes[1] = uTemp;
            }
        }
    }
}

/**Function*************************************************************

  Synopsis    [Prints the supergate library after deriving parameters.]

  Description [This procedure is very useful to see the library after
  it has been read into the mapper by "read_super" and all the information
  about the supergates derived.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Map_LibraryPrintTree(pLib *Map_SuperLib_t) {
    var pGate *Map_Super_t;
    var i, k int32;

    // print all the info related to the supergates
//    for ( i = pLib->nVarsMax; i < (int)pLib->nLines; i++ )
    for  i = pLib.nVarsMax; i < 20; i++  {
        pGate = pLib.ppSupers[i];

        // write the gate's fanin info and formula
        fmt.Printf("%6d  ", pGate.Num);
        fmt.Printf("%c ", pGate.fSuper? '*' : ' ');
        fmt.Printf("%6s", Mio_GateReadName(pGate.pRoot));
        for  k = 0; k < int32(pGate.nFanins); k++  {
            fmt.Printf(" %6d", pGate.pFanins[k].Num);
        }
        fmt.Printf("  %s", pGate.pFormula);
        fmt.Printf("\n");

        // write the gate's derived info
        Extra_PrintBinary( os.Stdout, pGate.uTruth, 64 );
        fmt.Printf("  %3d", pGate.nGates);
        fmt.Printf("  %6.2f", pGate.Area);
        fmt.Printf("  (%4.2f, %4.2f)", pGate.tDelayMax.Rise, pGate.tDelayMax.Fall);
        fmt.Printf("\n");
        for  k = 0; k < pLib.nVarsMax; k++  {
            // print the constraint on the rise of the gate in the form (D1, D2), 
            // where D1 is the constraint related to the rise of the k-th PI
            // where D2 is the constraint related to the fall of the k-th PI
            if pGate.tDelaysR[k].Rise < 0 && pGate.tDelaysR[k].Fall < 0 {
                fmt.Printf(" (----, ----)");
            } else if pGate.tDelaysR[k].Fall < 0 {
                fmt.Printf(" (%4.2f, ----)", pGate.tDelaysR[k].Rise);
            } else if pGate.tDelaysR[k].Rise < 0 {
                fmt.Printf(" (----, %4.2f)", pGate.tDelaysR[k].Fall);
            } else {
                fmt.Printf(" (%4.2f, %4.2f)", pGate.tDelaysR[k].Rise, pGate.tDelaysR[k].Fall); }

            // print the constraint on the fall of the gate in the form (D1, D2), 
            // where D1 is the constraint related to the rise of the k-th PI
            // where D2 is the constraint related to the fall of the k-th PI
            if pGate.tDelaysF[k].Rise < 0 && pGate.tDelaysF[k].Fall < 0 {
                fmt.Printf(" (----, ----)");
            } else if pGate.tDelaysF[k].Fall < 0 {
                fmt.Printf(" (%4.2f, ----)", pGate.tDelaysF[k].Rise);
            } else if pGate.tDelaysF[k].Rise < 0 {
                fmt.Printf(" (----, %4.2f)", pGate.tDelaysF[k].Fall);
            } else {
                fmt.Printf(" (%4.2f, %4.2f)", pGate.tDelaysF[k].Rise, pGate.tDelaysF[k].Fall); }
            fmt.Printf("\n");
        }
        fmt.Printf("\n");
    }
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END
