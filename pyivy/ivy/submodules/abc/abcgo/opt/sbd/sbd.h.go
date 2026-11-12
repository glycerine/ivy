package sbd

/**CFile****************************************************************

  FileName    [sbd.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT-based optimization using internal don't-cares.]

  Synopsis    [External declarations.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: sbd.h,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/
 



////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_HEADER_START

////////////////////////////////////////////////////////////////////////
///                         BASIC TYPES                              ///
////////////////////////////////////////////////////////////////////////

   
type Sbd_Par_t struct {
    nLutSize             int32;     // target LUT size
    nLutNum             int32;      // target LUT count
    nCutSize             int32;     // target cut size
    nCutNum             int32;      // target cut count
    nTfoLevels             int32;   // the number of TFO levels (windowing)
    nTfoFanMax             int32;   // the max number of fanouts (windowing)
    nWinSizeMax             int32;  // maximum window size (windowing)
    nBTLimit             int32;     // maximum number of SAT conflicts 
    nWords             int32;       // simulation word count
    fMapping             int32;     // generate mapping
    fMoreCuts             int32;    // use several cuts
    fFindDivs             int32;    // perform divisor search
    fUsePath             int32;     // optimize only critical path
    fArea             int32;        // area-oriented optimization
    fCover             int32;       // use complete cover procedure
    fVerbose             int32;     // verbose flag
    fVeryVerbose             int32; // verbose flag
}


////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                    FUNCTION DECLARATIONS                         ///
////////////////////////////////////////////////////////////////////////

/*=== sbdCnf.c ==========================================================*/
/*=== sbdCore.c ==========================================================*/
// extern void         Sbd_ParSetDefault( Sbd_Par_t * pPars );

// extern Gia_Man_t *  Sbd_NtkPerform( Gia_Man_t * p, Sbd_Par_t * pPars );



//ABC_NAMESPACE_HEADER_END



////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

