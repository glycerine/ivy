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

type Amap_Lib_t = Amap_Lib_t_

type Amap_Par_t struct {
	NIterFlow int32   // iterations of area flow
	NIterArea int32   // iteratoins of exact area
	NCutsMax  int32   // the maximum number of cuts at a node
	FUseMuxes int32   // enables the use of MUXes
	FUseXors  int32   // enables the use of XORs
	FFreeInvs int32   // assume inverters are free (area = 0)
	FEpsilon  float32 // used to compare floating point numbers
	FADratio  float32 // ratio of area/delay improvement
	FVerbose  int32   // verbosity flag
}

type Amap_Out_t struct {
	pName string  // gate name
	Type  int16   // node type (-1=input; 0=internal; 1=output)
	nFans int16   // number of fanins
	pFans []int32 // fanin indices in the mapped-node array
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
