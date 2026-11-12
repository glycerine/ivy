package amap

/**CFile****************************************************************

  FileName    [amap.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Technology mapper for standard cells.]

  Synopsis    [External declarations.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: amap.h,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

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

type Amap_Lib_t = Amap_Lib_t_;

   
type Amap_Par_t struct {
    nIterFlow    int32;   // iterations of area flow
    nIterArea    int32;   // iteratoins of exact area
    nCutsMax    int32;    // the maximum number of cuts at a node
    fUseMuxes    int32;   // enables the use of MUXes
    fUseXors    int32;    // enables the use of XORs
    fFreeInvs    int32;   // assume inverters are free (area = 0)
    fEpsilon  float32;    // used to compare floating point numbers
    fADratio  float32;    // ratio of area/delay improvement
    fVerbose    int32;    // verbosity flag
}

   
type Amap_Out_t struct {
    pName  string;     // gate name
    Type  int16;      // node type   (-1=input; 0=internal; 1=output)
    nFans  int16;     // number of fanins
    pFans [0]int32;  // fanin
}

////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                    FUNCTION DECLARATIONS                         ///
////////////////////////////////////////////////////////////////////////

/*=== amapCore.c ==========================================================*/
// extern void          Amap_ManSetDefaultParams( Amap_Par_t * pPars );

//extern Vec_Ptr_t * Amap_ManTest( Aig_Man_t * pAig, Amap_Par_t * pPars );
/*=== amapLib.c ==========================================================*/
// extern void          Amap_LibFree( Amap_Lib_t * p );

// extern void          Amap_LibPrintSelectedGates( Amap_Lib_t * p, int fAllGates );

// extern Amap_Lib_t *  Amap_LibReadAndPrepare( char * pFileName, char * pBuffer, int fVerbose, int fVeryVerbose );

/*=== amapLiberty.c ==========================================================*/
// extern int           Amap_LibertyParse( char * pFileName, int fVerbose );

// extern Vec_Str_t *   Amap_LibertyParseStr( char * pFileName, int fVerbose );




//ABC_NAMESPACE_HEADER_END





////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

