package tim

/**CFile****************************************************************

  FileName    [timInt.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Hierarchy/timing manager.]

  Synopsis    [Internal declarations.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - April 28, 2007.]

  Revision    [$Id: timInt.h,v 1.00 2007/04/28 00:00:00 alanmi Exp $]

***********************************************************************/





////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

//#include <stdio.h>
//#include <stdlib.h>
//#include <string.h>
//#include <assert.h>

//#include "misc/vec/vec.h"
//#include "misc/mem/mem.h"
//#include "tim.h"

////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_HEADER_START

////////////////////////////////////////////////////////////////////////
///                         BASIC TYPES                              ///
////////////////////////////////////////////////////////////////////////

type Tim_Box_t =           Tim_Box_t_;
type Tim_Obj_t =           Tim_Obj_t_;

// timing manager
type Tim_Man_t_ struct {
    vBoxes       *Vec_Ptr_t;         // the timing boxes
    vDelayTables       *Vec_Ptr_t;   // pointers to the delay tables
    pMemObj      *Mem_Flex_t;        // memory manager for boxes
    nTravIds              int32;       // traversal ID of the manager
    fUseTravId              int32;     // enables the use of traversal ID
    nCis              int32;           // the number of PIs
    nCos              int32;           // the number of POs
    pCis       *Tim_Obj_t;           // timing info for the PIs
    pCos       *Tim_Obj_t;           // timing info for the POs
}

// timing box
type Tim_Box_t_ struct {
    iBox              int32;           // the unique ID of this box
    TravId              int32;         // traversal ID of this box
    nInputs              int32;        // the number of box inputs (POs)
    nOutputs              int32;       // the number of box outputs (PIs)
    iDelayTable              int32;    // index of the delay table
    iCopy              int32;          // copy of this box
    fBlack              int32;         // this is black box
    Inouts [0]int32;      // the int numbers of PIs and POs
}

// timing object
type Tim_Obj_t_ struct {
    Id              int32;             // the ID of this object
    TravId              int32;         // traversal ID of this object
    iObj2Box              int32;       // mapping of the object into its box
    iObj2Num              int32;       // mapping of the object into its number in the box
    timeArr            float32;        // arrival time of the object
    timeReq            float32;        // required time of the object
}

////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

func Tim_ManCi(p *Tim_Man_t, i int32) *Tim_Obj_t { assert(i < p.nCis, "i < p.nCis"); return p.pCis + i;      }
func Tim_ManCo(p *Tim_Man_t, i int32) *Tim_Obj_t { assert(i < p.nCos, "i < p.nCos"); return p.pCos + i;      }
func Tim_ManBox(p *Tim_Man_t, i int32) *Tim_Box_t { return (Tim_Box_t *)Vec_PtrEntry(p.vBoxes, i); }

func Tim_ManCiBox(p *Tim_Man_t, i int32) *Tim_Box_t { if Tim_ManCi(p,i).iObj2Box < 0 {
    return nil;
} else {
    return (Tim_Box_t *)Vec_PtrEntry( p.vBoxes, Tim_ManCi(p,i).iObj2Box );
} }
func Tim_ManCoBox(p *Tim_Man_t, i int32) *Tim_Box_t { if Tim_ManCo(p,i).iObj2Box < 0 {
    return nil;
} else {
    return (Tim_Box_t *)Vec_PtrEntry( p.vBoxes, Tim_ManCo(p,i).iObj2Box );
} }

func Tim_ManBoxInput(p *Tim_Man_t, pBox *Tim_Box_t, i int32) *Tim_Obj_t { assert(i < pBox.nInputs, "i < pBox.nInputs"); return p.pCos + pBox.Inouts[i];               }
func Tim_ManBoxOutput(p *Tim_Man_t, pBox *Tim_Box_t, i int32) *Tim_Obj_t { assert(i < pBox.nOutputs, "i < pBox.nOutputs"); return p.pCis + pBox.Inouts[pBox.nInputs+i]; }

////////////////////////////////////////////////////////////////////////
///                             ITERATORS                            ///
////////////////////////////////////////////////////////////////////////

// Original macro:
// #define Tim_ManForEachCi( p, pObj, i )                                  \
//     for ( i = 0; (i < (p)->nCis) && ((pObj) = (p)->pCis + i); i++ )

// Original macro:
// #define Tim_ManForEachCo( p, pObj, i )                                  \
//     for ( i = 0; (i < (p)->nCos) && ((pObj) = (p)->pCos + i); i++ )


// Original macro:
// #define Tim_ManForEachPi( p, pObj, i )                                  \
//     Tim_ManForEachCi( p, pObj, i ) if ( pObj->iObj2Box >= 0 ) {} else 

// Original macro:
// #define Tim_ManForEachPo( p, pObj, i )                                  \
//     Tim_ManForEachCo( p, pObj, i ) if ( pObj->iObj2Box >= 0 ) {} else 


// Original macro:
// #define Tim_ManForEachBox( p, pBox, i )                                 \
//     Vec_PtrForEachEntry( Tim_Box_t *, p->vBoxes, pBox, i )


// Original macro:
// #define Tim_ManBoxForEachInput( p, pBox, pObj, i )                      \
//     for ( i = 0; (i < (pBox)->nInputs) && ((pObj) = Tim_ManBoxInput(p, pBox, i)); i++ )

// Original macro:
// #define Tim_ManBoxForEachOutput( p, pBox, pObj, i )                     \
//     for ( i = 0; (i < (pBox)->nOutputs) && ((pObj) = Tim_ManBoxOutput(p, pBox, i)); i++ )


// Original macro:
// #define Tim_ManForEachTable( p, pTable, i )                             \
//     Vec_PtrForEachEntry( float *, p->vDelayTables, pTable, i )


////////////////////////////////////////////////////////////////////////
///                     SEQUENTIAL ITERATORS                         ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                    FUNCTION DECLARATIONS                         ///
////////////////////////////////////////////////////////////////////////

/*=== time.c ===========================================================*/


//ABC_NAMESPACE_HEADER_END





////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

