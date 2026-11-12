package mio

/**CFile****************************************************************

  FileName    [mioInt.h]

  PackageName [MVSIS 2.0: Multi-valued logic synthesis system.]

  Synopsis    [File reading/writing for technology mapping.]

  Author      [MVSIS Group]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - September 8, 2003.]

  Revision    [$Id: mioInt.h,v 1.4 2004/06/28 14:20:25 alanmi Exp $]

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
//#include "misc/st/st.h"
//#include "mio.h"
 
//ABC_NAMESPACE_HEADER_START


////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////

const MIO_STRING_GATE =       "GATE"
const MIO_STRING_LATCH =      "LATCH"
const MIO_STRING_PIN =        "PIN"
const MIO_STRING_NONINV =     "NONINV"
const MIO_STRING_INV =        "INV"
const MIO_STRING_UNKNOWN =    "UNKNOWN"

const MIO_STRING_CONST0 =     "CONST0"
const MIO_STRING_CONST1 =     "CONST1"
 
// the bit masks
// Original macro:
// #define    MIO_MASK(n)         ((~((unsigned)0)) >> (32-(n)))

#define    MIO_FULL             (~((unsigned)0))

////////////////////////////////////////////////////////////////////////
///                    STRUCTURE DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

type  Mio_LibraryStruct_t_ struct {
    pName              string;       // the name of the library
    nGates                int32;      // the number of the gates
    ppGates0       **Mio_Gate_t;    // the array of gates in the original order
    ppGatesName       **Mio_Gate_t; // the array of gates sorted by name
    pGates        *Mio_Gate_t;      // the linked list of all gates in no particular order
    pGate0        *Mio_Gate_t;      // the constant zero gate
    pGate1        *Mio_Gate_t;      // the constant one gate
    pGateBuf        *Mio_Gate_t;    // the buffer
    pGateInv        *Mio_Gate_t;    // the inverter
    pGateNand2        *Mio_Gate_t;  // the NAND2 gate
    pGateAnd2        *Mio_Gate_t;   // the AND2 gate
    pGateNor2        *Mio_Gate_t;   // the NOR2 gate
    pGateOr2        *Mio_Gate_t;    // the OR2 gate
    tName2Gate          *st__table;  // the mapping of gate names into their pointer
    pMmFlex        *Mem_Flex_t;     // the memory manaqer for SOPs
    vCube         *Vec_Str_t;       // temporary cube
    // matching
    fPinFilter                int32;  // pin filtering
    fPinPerm                int32;    // pin permutation
    fPinQuick                int32;   // pin permutation
    vTtMem         *Vec_Mem_t;      // truth tables
    vTt2Match         *Vec_Wec_t;   // matches for truth tables
    pCells       *Mio_Cell2_t;      // library gates
    nCells                int32;      // library gate count
    vNames         *Vec_Ptr_t;
    vTruths         *Vec_Wrd_t; 
    vTt2Match4         *Vec_Int_t;
    vConfigs         *Vec_Int_t;
    Vec_Mem_t *        vTtMem2[3]; 
    Vec_Int_t *        vTt2Match2[3];
} 

type  Mio_GateStruct_t_ struct {
    // information derived from the genlib file
    pName              string;       // the name of the gate
    dArea             float64;       // the area of the gate
    pForm              string;       // the formula describing functionality of the gate
    pPins         *Mio_Pin_t;       // the linked list of all pins (one pin if info is the same)
    pOutName              string;    // the name of the output pin 
    // the library to which this gate belongs
    pLib     *Mio_Library_t; 
    // the next gate in the list
    pNext        *Mio_Gate_t;    
    pTwin        *Mio_Gate_t;    

    // the derived information
    Cell                int32;        // cell id
    nInputs                int32;     // the number of inputs
    Profile                int32;     // the number of occurrences
    Profile2                int32;    // the number of occurrences
    dDelayMax             float64;   // the maximum delay
    pSop              string;        // sum-of-products
    vExpr         *Vec_Int_t;       // boolean expression
    union { word       uTruth;      // truth table
    word *             pTruth; };   // pointer to the truth table
    Value                int32;       // user's information
}

type  Mio_PinStruct_t_ struct {
    pName              string;
    Phase     Mio_PinPhase_t;
    dLoadInput             float64;
    dLoadMax             float64;
    dDelayBlockRise             float64;
    dDelayFanoutRise             float64;
    dDelayBlockFall             float64;
    dDelayFanoutFall             float64;
    dDelayBlockMax             float64;
    pNext         *Mio_Pin_t;     
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

/*=== mio.c =============================================================*/
/*=== mioRead.c =============================================================*/
/*=== mioUtils.c =============================================================*/


//ABC_NAMESPACE_HEADER_END



////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////
