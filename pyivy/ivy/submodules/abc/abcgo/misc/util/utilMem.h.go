package util

/**CFile****************************************************************

  FileName    [utilInt.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Memory recycling utilities.]

  Synopsis    [Internal declarations.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: utilInt.h,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/
 




////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////



//ABC_NAMESPACE_HEADER_START


var s_vAllocs any;   // storage of allocated pointers
var s_vFrees any;    // storage of deallocated pointers
var s_fInterrupt int32;   // set to 1 when it is time to backout

////////////////////////////////////////////////////////////////////////
///                         BASIC TYPES                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                    FUNCTION DECLARATIONS                         ///
////////////////////////////////////////////////////////////////////////

/*=== utilMem.c ==========================================================*/
// extern void *       Util_MemRecAlloc( void * pMem );

// extern void *       Util_MemRecFree( void * pMem );

// extern void         Util_MemStart();

// extern void         Util_MemQuit();

// extern void         Util_MemRecycle();

// extern int          Util_MemRecIsSet();




//ABC_NAMESPACE_HEADER_END





////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

