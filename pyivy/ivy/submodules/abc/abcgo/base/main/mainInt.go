package main

import "os"

/**CFile****************************************************************

  FileName    [mainInt.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [The main package.]

  Synopsis    [Internal declarations of the main package.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: mainInt.h,v 1.1 2008/05/14 22:13:13 wudenni Exp $]

***********************************************************************/




 
////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

//#include "main.h"
//#include "misc/tim/tim.h"
//#include "map/if/if.h"
//#include "aig/aig/aig.h"
//#include "aig/gia/gia.h"
//#include "proof/ssw/ssw.h"
//#include "proof/fra/fra.h"

#ifdef ABC_USE_CUDD
#include "bdd/extrab/extraBdd.h"
#endif

//ABC_NAMESPACE_HEADER_START

////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////

// the current version
const ABC_VERSION = "UC Berkeley, ABC 1.01"

// the maximum length of an input line 
#define ABC_MAX_STR     (1<<15)

////////////////////////////////////////////////////////////////////////
///                    STRUCTURE DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

typedef void (*Abc_Frame_Callback_BmcFrameDone_Func)(int frame, int po, int status);

type Abc_Frame_t_ struct {
    // general info
    sVersion           string;      // the name of the current version
    sBinary           string;       // the name of the binary running
    // commands, aliases, etc
    tCommands       *st__table;     // the command table
    tAliases       *st__table;      // the alias table
    tFlags       *st__table;        // the flag table
    aHistory      *Vec_Ptr_t;      // the command history
    // the functionality
    pNtkCur      *Abc_Ntk_t;       // the current network
    pNtkBestDelay      *Abc_Ntk_t; // the current network
    pNtkBestArea      *Abc_Ntk_t;  // the current network
    pNtkBackup      *Abc_Ntk_t;    // the current network
    nSteps             int32;        // the counter of different network processed
    fSource             int32;       // marks the source mode
    fAutoexac             int32;     // marks the autoexec mode
    fBatchMode             int32;    // batch mode flag
    fBridgeMode             int32;   // bridge mode flag
    // save/load
    pNtkBest      *Abc_Ntk_t;        // the current network
    nBestNtkArea           float32;   // best area
    nBestNtkDelay           float32;  // best delay
    nBestNtkNodes             int32;  // best nodes
    nBestNtkLevels             int32; // best levels

    // output streams
    Out           *os.File;
    Err           *os.File;
    Hst           *os.File;
    // used for runtime measurement
    TimeCommand          float64;   // the runtime of the last command
    TimeTotal          float64;     // the total runtime of all commands
    // temporary storage for structural choices
    vStore      *Vec_Ptr_t;        // networks to be used by choice
    // decomposition package    
    pManDec           any;       // decomposition manager
    pManDsd           any;       // decomposition manager
    pManDsd2           any;      // decomposition manager
    // libraries for mapping
    pLibLut           any;       // the current LUT library
    pLibBox           any;       // the current box library
    pLibGen           any;       // the current genlib
    pLibGen2           any;      // the current genlib
    pLibSuper           any;     // the current supergate library
    pLibScl           any;       // the current Liberty library
    pAbcCon           any;       // constraint manager
    // timing constraints
    pDrivingCell           string;  // name of the driving cell
    MaxLoad           float32;       // maximum output load
    // inductive don't-cares
    vIndFlops      *Vec_Int_t;
    nIndFrames             int32;

    // new code
    pGia      *Gia_Man_t;          // alternative current network as a light-weight AIG
    pGia2      *Gia_Man_t;         // copy of the above
    pGiaBest      *Gia_Man_t;      // copy of the above
    pGiaBest2      *Gia_Man_t;     // copy of the above
    pGiaSaved      *Gia_Man_t;     // copy of the above
    nBestLuts             int32;     // best LUT count
    nBestEdges             int32;    // best edge count
    nBestLevels             int32;   // best level count
    nBestLuts2             int32;     // best LUT count
    nBestEdges2             int32;    // best edge count
    nBestLevels2             int32;   // best level count
    pCex      *Abc_Cex_t;          // a counter-example to fail the current network
    pCex2      *Abc_Cex_t;         // copy of the above
    vCexVec      *Vec_Ptr_t;       // a vector of counter-examples if more than one PO fails
    vPoEquivs      *Vec_Ptr_t;     // equivalence classes of isomorphic primary outputs
    vStatuses      *Vec_Int_t;     // problem status for each output
    vAbcObjIds      *Vec_Int_t;    // object IDs
    Status             int32;                // the status of verification problem (proved=1, disproved=0, undecided=-1)
    nFrames             int32;               // the number of time frames completed by BMC
    vPlugInComBinPairs      *Vec_Ptr_t;    // pairs of command and its binary name
    vLTLProperties_global      *Vec_Ptr_t; // related to LTL
    vSignalNames      *Vec_Ptr_t;  // temporary storage for signal names
    pSpecName           string;
    pSave1           any; 
    pSave2           any; 
    pSave3           any; 
    pSave4           any; 
    pAbc85Ntl           any;
    pAbc85Ntl2           any;
    pAbc85Best           any;
    pAbc85Delay           any;
    pAbcWlc           any;
    pAbcWlcInv      *Vec_Int_t;
    pAbcBac           any;
    pAbcCba           any;
    pAbcPla           any;
    pJsonStrs      *Abc_Nam_t;
    vJsonObjs      *Vec_Wec_t;
#ifdef ABC_USE_CUDD
    DdManager *     dd;            // temporary BDD package
#endif
    pGiaMiniAig      *Gia_Man_t; 
    pGiaMiniLut      *Gia_Man_t; 
    vCopyMiniAig      *Vec_Int_t;
    vCopyMiniLut      *Vec_Int_t;
    pArray            *int32;
    pBoxes            *int32;
    pNdr           any;
    pNdrArray            *int32;

    pFuncOnFrameDone Abc_Frame_Callback_BmcFrameDone_Func;
}

typedef void (*Abc_Frame_Initialization_Func)( Abc_Frame_t * pAbc );

struct Abc_FrameInitializer_t_;
   

type Abc_FrameInitializer_t struct {
    init Abc_Frame_Initialization_Func;
    destroy Abc_Frame_Initialization_Func;

    next *Abc_FrameInitializer_t;
    prev *Abc_FrameInitializer_t;
}

////////////////////////////////////////////////////////////////////////
///                       GLOBAL VARIABLES                           ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                       MACRO DEFINITIONS                          ///
////////////////////////////////////////////////////////////////////////


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/*=== mvMain.c ===========================================================*/
extern ABC_DLL int             main( int argc, char * argv[] );
/*=== mvInit.c ===================================================*/
extern ABC_DLL void            Abc_FrameInit( Abc_Frame_t * pAbc );
extern ABC_DLL void            Abc_FrameEnd( Abc_Frame_t * pAbc );
extern ABC_DLL void            Abc_FrameAddInitializer( Abc_FrameInitializer_t* p );
/*=== mvFrame.c =====================================================*/
extern ABC_DLL Abc_Frame_t *   Abc_FrameAllocate();
extern ABC_DLL void            Abc_FrameDeallocate( Abc_Frame_t * p );
/*=== mvUtils.c =====================================================*/
extern ABC_DLL char *          Abc_UtilsGetVersion( Abc_Frame_t * pAbc );
extern ABC_DLL char *          Abc_UtilsGetUsersInput( Abc_Frame_t * pAbc );
extern ABC_DLL void            Abc_UtilsPrintHello( Abc_Frame_t * pAbc );
extern ABC_DLL void            Abc_UtilsPrintUsage( Abc_Frame_t * pAbc, char * ProgName );
extern ABC_DLL void            Abc_UtilsSource( Abc_Frame_t * pAbc );



//ABC_NAMESPACE_HEADER_END



////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////
