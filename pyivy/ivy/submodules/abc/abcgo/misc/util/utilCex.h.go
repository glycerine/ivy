package util

/**CFile****************************************************************

  FileName    [utilCex.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Handling sequential counter-examples.]

  Synopsis    [Handling sequential counter-examples.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - Feburary 13, 2011.]

  Revision    [$Id: utilCex.h,v 1.00 2011/02/11 00:00:00 alanmi Exp $]

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

// sequential counter-example
   
type Abc_Cex_t struct {
    iPo              int32;       // the zero-based number of PO, for which verification failed
    iFrame              int32;    // the zero-based number of the time-frame, for which verificaiton failed
    nRegs              int32;     // the number of registers in the miter 
    nPis              int32;      // the number of primary inputs in the miter
    nBits              int32;     // the number of words of bit data used
    pData [0]uint32;  // the cex bit data (the number of bits: nRegs + (iFrame+1) * nPis)
}

////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                    FUNCTION DECLARATIONS                         ///
////////////////////////////////////////////////////////////////////////

/*=== utilCex.c ===========================================================*/
// extern Abc_Cex_t *   Abc_CexAlloc( int nRegs, int nTruePis, int nFrames );

// extern Abc_Cex_t *   Abc_CexAllocFull( int nRegs, int nTruePis, int nFrames );

// extern Abc_Cex_t *   Abc_CexMakeTriv( int nRegs, int nTruePis, int nTruePos, int iFrameOut );

// extern Abc_Cex_t *   Abc_CexCreate( int nRegs, int nTruePis, int * pArray, int iFrame, int iPo, int fSkipRegs );

// extern Abc_Cex_t *   Abc_CexDup( Abc_Cex_t * p, int nRegsNew );

// extern Abc_Cex_t *   Abc_CexDeriveFromCombModel( int * pModel, int nPis, int nRegs, int iPo );

// extern Abc_Cex_t *   Abc_CexMerge( Abc_Cex_t * pCex, Abc_Cex_t * pPart, int iFrBeg, int iFrEnd );

// extern void          Abc_CexPrintStats( Abc_Cex_t * p );

// extern void          Abc_CexPrintStatsInputs( Abc_Cex_t * p, int nInputs );

// extern void          Abc_CexPrint( Abc_Cex_t * p );

// extern void          Abc_CexFreeP( Abc_Cex_t ** p );

// extern void          Abc_CexFree( Abc_Cex_t * p );

// extern Abc_Cex_t *   Abc_CexTransformPhase( Abc_Cex_t * p, int nPisOld, int nPosOld, int nRegsOld );

// extern Abc_Cex_t *   Abc_CexTransformTempor( Abc_Cex_t * p, int nPisOld, int nPosOld, int nRegsOld );

// extern Abc_Cex_t *   Abc_CexTransformUndc( Abc_Cex_t * p, char * pInit );

// extern Abc_Cex_t *   Abc_CexPermute( Abc_Cex_t * p, Vec_Int_t * vMapOld2New );

// extern Abc_Cex_t *   Abc_CexPermuteTwo( Abc_Cex_t * p, Vec_Int_t * vPermOld, Vec_Int_t * vPermNew );

// extern int           Abc_CexCountOnes( Abc_Cex_t * p );


//ABC_NAMESPACE_HEADER_END



////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////
