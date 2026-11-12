package fpga

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [fpgaLib.c]

  PackageName [MVSIS 1.3: Multi-valued logic synthesis system.]

  Synopsis    [Technology mapping for variable-size-LUT FPGAs.]

  Author      [MVSIS Group]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 2.0. Started - August 18, 2004.]

  Revision    [$Id: fpgaLib.c,v 1.4 2005/01/23 06:59:41 alanmi Exp $]

***********************************************************************/

//#include "fpgaInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [APIs to access LUT library.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_LutLibReadVarMax(p *Fpga_LutLib_t) int32 { return p.LutMax;     }
func Fpga_LutLibReadLutAreas(p *Fpga_LutLib_t) *float32 { return p.pLutAreas;  }
func Fpga_LutLibReadLutArea(p *Fpga_LutLib_t, Size int32) float32 { assert(Size <= p.LutMax, "Size <= p.LutMax"); return p.pLutAreas[Size];  }

/**Function*************************************************************

  Synopsis    [Reads the description of LUTs from the LUT library file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_LutLibRead(FileName string, fVerbose int32) *Fpga_LutLib_t {
    var pBuffer [1000]byte;
    var pToken string;
    var p *Fpga_LutLib_t;
    var pFile *os.File;
    var i, k int32;

    pFile = os.Open(FileName);
    if pFile == nil {
        fmt.Printf("Cannot open LUT library file \"%s\".\n", FileName);
        return nil;
    }

    // inlining of macro ABC_ALLOC
    p = ((Fpga_LutLib_t *) malloc(sizeof(Fpga_LutLib_t) * (size_t)(1)));
    memset( p, 0, sizeof(Fpga_LutLib_t) );
    p.pName = Extra_UtilStrsav( FileName );

    i = 1;
    for  fgets( pBuffer, 1000, pFile ) != nil  {
        pToken = strtok( pBuffer, " \t\n" );
        if pToken == nil {
            continue; }
        if pToken[0] == '#' {
            continue; }
        if i != atoi(pToken) {
            fmt.Printf("Error in the LUT library file \"%s\".\n", FileName);
            // inlining of macro ABC_FREE
            ((p) ? (free((char *) (p)), (p) = 0) : 0);
            return nil;
        }

        // read area
        pToken = strtok( nil, " \t\n" );
        p.pLutAreas[i] = float32(atof(pToken));

        // read delays
        k = 0;
        for  pToken = strtok( nil, " \t\n" )  {
            p.pLutDelays[i][k++] = float32(atof(pToken));
        }

        // check for out-of-bound
        if k > i {
            fmt.Printf("LUT %d has too many pins (%d). Max allowed is %d.\n", i, k, i);
            return nil;
        }

        // check if var delays are specifies
        if k > 1 {
            p.fVarPinDelays = 1; }

        if i == FPGA_MAX_LUTSIZE {
            fmt.Printf("Skipping LUTs of size more than %d.\n", i);
            return nil;
        }
        i++;
    }
    p.LutMax = i-1;
/*
    if ( p->LutMax > FPGA_MAX_LEAVES )
    {
        p->LutMax = FPGA_MAX_LEAVES;
        printf( "Warning: LUTs with more than %d inputs will not be used.\n", FPGA_MAX_LEAVES );
    }
*/
    // check the library
    if p.fVarPinDelays {
        for  i = 1; i <= p.LutMax; i++  {
            for  k = 0; k < i; k++  {
                if p.pLutDelays[i][k] <= 0.0 {
                    fmt.Printf("Warning: Pin %d of LUT %d has delay %f. Pin delays should be non-negative numbers. Technology mapping may not work correctly.\n", k, i, p.pLutDelays[i][k]); }
                if (k != 0) && p.pLutDelays[i][k-1] > p.pLutDelays[i][k] {
                    fmt.Printf("Warning: Pin %d of LUT %d has delay %f. Pin %d of LUT %d has delay %f. Pin delays should be in non-decreasing order. Technology mapping may not work correctly.\n", k-1, i, p.pLutDelays[i][k-1], k, i, p.pLutDelays[i][k]); }
            }
        }
    
    } else {
        for  i = 1; i <= p.LutMax; i++  {
            if p.pLutDelays[i][0] <= 0.0 {
                fmt.Printf("Warning: LUT %d has delay %f. Pin delays should be non-negative numbers. Technology mapping may not work correctly.\n", i, p.pLutDelays[i][0]); }
        }
    }

    return p;
}

/**Function*************************************************************

  Synopsis    [Duplicates the LUT library.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_LutLibDup(p *Fpga_LutLib_t) *Fpga_LutLib_t {
    var pNew *Fpga_LutLib_t;
    // inlining of macro ABC_ALLOC
    pNew = ((Fpga_LutLib_t *) malloc(sizeof(Fpga_LutLib_t) * (size_t)(1)));
    *pNew = *p;
    pNew.pName = Extra_UtilStrsav( pNew.pName );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Frees the LUT library.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_LutLibFree(pLutLib *Fpga_LutLib_t) {
    if pLutLib == nil {
        return; }
    // inlining of macro ABC_FREE
    ((pLutLib.pName) ? (free((char *) (pLutLib.pName)), (pLutLib.pName) = 0) : 0);
    // inlining of macro ABC_FREE
    ((pLutLib) ? (free((char *) (pLutLib)), (pLutLib) = 0) : 0);
}


/**Function*************************************************************

  Synopsis    [Prints the LUT library.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_LutLibPrint(pLutLib *Fpga_LutLib_t) {
    var i, k int32;
    fmt.Printf("# The area/delay of k-variable LUTs:\n");
    fmt.Printf("# k    area     delay\n");
    if pLutLib.fVarPinDelays {
        for  i = 1; i <= pLutLib.LutMax; i++  {
            fmt.Printf("%d   %7.2f  ", i, pLutLib.pLutAreas[i]);
            for  k = 0; k < i; k++  {
                fmt.Printf(" %7.2f", pLutLib.pLutDelays[i][k]);
            }
            fmt.Printf("\n");
        }
    
    } else {
        for  i = 1; i <= pLutLib.LutMax; i++  {
            fmt.Printf("%d   %7.2f   %7.2f\n", i, pLutLib.pLutAreas[i], pLutLib.pLutDelays[i][0]);
        } }
}

/**Function*************************************************************

  Synopsis    [Returns 1 if the delays are discrete.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_LutLibDelaysAreDiscrete(pLutLib *Fpga_LutLib_t) int32 {
    var Delay float32;
    var i int32;
    for  i = 1; i <= pLutLib.LutMax; i++  {
        Delay = pLutLib.pLutDelays[i][0];
        if (float32((int32(Delay)))) != Delay {
            return 0; }
    }
    return 1;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

