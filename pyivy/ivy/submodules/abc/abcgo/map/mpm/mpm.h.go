package mpm

/**CFile****************************************************************

  FileName    [mpm.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Configurable technology mapper.]

  Synopsis    [External declarations.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 1, 2013.]

  Revision    [$Id: mpm.h,v 1.00 2013/06/01 00:00:00 alanmi Exp $]

***********************************************************************/
 




////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////
 
//#include <stdio.h>
//#include <stdlib.h>
//#include <string.h>
//#include <assert.h>

//ABC_NAMESPACE_HEADER_START

////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////

const MPM_VAR_MAX =  12  

////////////////////////////////////////////////////////////////////////
///                         BASIC TYPES                              ///
////////////////////////////////////////////////////////////////////////

   
type Mpm_LibLut_t struct {
    pName            string;                                     // the name of the LUT library
    LutMax              int32;                                    // the maximum LUT size 
    fVarPinDelays              int32;                             // set to 1 if variable pin delays are specified
    pLutAreas [MPM_VAR_MAX+1]int32;                  // the areas of LUTs
    pLutDelays [MPM_VAR_MAX+1][MPM_VAR_MAX+1]int32;  // the delays of LUTs
}

   
type Mpm_Par_t struct {
    pLib    *Mpm_LibLut_t;
    pScl            any;
    nNumCuts              int32;
    DelayTarget              int32;
    fUseGates              int32;
    fUseTruth              int32;
    fUseDsd              int32;
    fCutMin              int32;
    fOneRound              int32;
    fDeriveLuts              int32;
    fMap4Cnf              int32;
    fMap4Aig              int32;
    fMap4Gates              int32;
    fVerbose              int32;
    fVeryVerbose              int32;
}

////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////


////////////////////////////////////////////////////////////////////////
///                    FUNCTION DECLARATIONS                         ///
////////////////////////////////////////////////////////////////////////

/*=== mpmCore.c ===========================================================*/
// extern void           Mpm_ManSetParsDefault( Mpm_Par_t * p );

/*=== mpmLib.c ===========================================================*/
// extern Mpm_LibLut_t * Mpm_LibLutSetSimple( int nLutSize );

// extern void           Mpm_LibLutFree( Mpm_LibLut_t * pLib );


//ABC_NAMESPACE_HEADER_END



////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

