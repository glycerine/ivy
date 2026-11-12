package nm

/**CFile****************************************************************

  FileName    [nmInt.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Name manager.]

  Synopsis    [Internal declarations.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: nmInt.h,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/
 




////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

//#include "misc/extra/extra.h"
//#include "misc/vec/vec.h"
//#include "nm.h"

////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////



//ABC_NAMESPACE_HEADER_START


////////////////////////////////////////////////////////////////////////
///                         BASIC TYPES                              ///
////////////////////////////////////////////////////////////////////////

   
type Nm_Entry_t struct {
    Type         uint32;          // object type
    ObjId         uint32;         // object ID
    pNextI2N      *Nm_Entry_t;      // the next entry in the ID hash table
    pNextN2I      *Nm_Entry_t;      // the next entry in the name hash table
    pNameSake      *Nm_Entry_t;     // the next entry with the same name
    Name [0]byte;       // name of the object
}

type Nm_Man_t_ struct {
    pBinsI2N     **Nm_Entry_t;      // mapping IDs into names
    pBinsN2I     **Nm_Entry_t;      // mapping names into IDs 
    nBins              int32;         // the number of bins in tables
    nEntries              int32;      // the number of entries
    nSizeFactor              int32;   // determined how much larger the table should be
    nGrowthFactor              int32; // determined how much the table grows after resizing
    pMem  *Extra_MmFlex_t;          // memory manager for entries (and names)
}

////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                    FUNCTION DECLARATIONS                         ///
////////////////////////////////////////////////////////////////////////

/*=== nmTable.c ==========================================================*/
// extern int              Nm_ManTableAdd( Nm_Man_t * p, Nm_Entry_t * pEntry );

// extern int              Nm_ManTableDelete( Nm_Man_t * p, int ObjId );

// extern Nm_Entry_t *     Nm_ManTableLookupId( Nm_Man_t * p, int ObjId );

// extern Nm_Entry_t *     Nm_ManTableLookupName( Nm_Man_t * p, char * pName, int Type );




//ABC_NAMESPACE_HEADER_END





////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


