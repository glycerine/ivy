package mio

import (
	"fmt"
	"os"
	"strings"
)

/**CFile****************************************************************

  FileName    [mioForm.c]

  PackageName [MVSIS 1.3: Multi-valued logic synthesis system.]

  Synopsis    [Parsing equestion formula.]

  Author      [MVSIS Group]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - September 8, 2003.]

  Revision    [$Id: mioForm.c,v 1.4 2004/06/28 14:20:25 alanmi Exp $]

***********************************************************************/

//#include "mioInt.h"
//#include "bdd/parse/parse.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// these symbols (and no other) can appear in the formulas
#define MIO_SYMB_AND    '*'
#define MIO_SYMB_AND2   '&'
#define MIO_SYMB_OR     '+'
#define MIO_SYMB_OR2    '|'
#define MIO_SYMB_XOR    '^'
#define MIO_SYMB_NOT    '!'
#define MIO_SYMB_AFTNOT '\''
#define MIO_SYMB_OPEN   '('
#define MIO_SYMB_CLOSE  ')'

// static int Mio_GateParseFormula( Mio_Gate_t * pGate );

// static int Mio_GateCollectNames( char * pFormula, char * pPinNames[] );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Deriving the functionality of the gates.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mio_LibraryParseFormulas(pLib *Mio_Library_t) int32 {
    var pGate *Mio_Gate_t;

    // count the gates
    pLib.nGates = 0;
    // inlining of macro Mio_LibraryForEachGate
    for  pGate = Mio_LibraryReadGates(pLib);                             (pGate != nil);                                                         pGate = Mio_GateReadNext(pGate)  {
        pLib.nGates++;
    }        

    // start a temporary BDD manager
    pLib.dd = Cudd_Init( 20, 0, CUDD_UNIQUE_SLOTS, CUDD_CACHE_SLOTS, 0 );
    // introduce ZDD variables
    Cudd_zddVarsFromBddVars( pLib.dd, 2 );

    // for each gate, derive its function
    // inlining of macro Mio_LibraryForEachGate
    for  pGate = Mio_LibraryReadGates(pLib);                             (pGate != nil);                                                         pGate = Mio_GateReadNext(pGate)  {
        if Mio_GateParseFormula( pGate ) != 0 {
            return 1; }
    }
    return 0;
}


/**Function*************************************************************

  Synopsis    [Registers the cube string with the network.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mio_SopRegister(pMan *Mem_Flex_t, pName string) string {
    var pRegName string;
    if pName == nil { return nil; }
    pRegName = Mem_FlexEntryFetch( pMan, len(pName) + 1 );
    pRegName = pName;
    return pRegName;
}

/**Function*************************************************************

  Synopsis    [Deriving the functionality of the gates.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mio_GateParseFormula(pGate *Mio_Gate_t) int32 {
    extern char * Abc_ConvertBddToSop( Mem_Flex_t * pMan, DdManager * dd, DdNode * bFuncOn, DdNode * bFuncOnDc, int nFanins, int fAllPrimes, Vec_Str_t * vCube, int fMode );
    var dd *DdManager  = pGate.pLib.dd;
    var pPinNames [100]string;
    var pPinNamesCopy [100]string;
    var pPin *Mio_Pin_t;
    var ppPin **Mio_Pin_t;
    var nPins, iPin, i int32;

    // set the maximum delay of the gate; count pins
    pGate.dDelayMax = 0.0;
    nPins = 0;
    // inlining of macro Mio_GateForEachPin
    for  pPin = Mio_GateReadPins(pGate);                                 (pPin != nil);                                                          pPin = Mio_PinReadNext(pPin)  {
        // set the maximum delay of the gate
        if pGate.dDelayMax < pPin.dDelayBlockMax {
            pGate.dDelayMax = pPin.dDelayBlockMax; }
        // count the pin
        nPins++;
    }

    // check for the gate with const function
    if nPins == 0 {
        if int32(strings.Compare(pGate.pForm, MIO_STRING_CONST0)) == 0 {
            pGate.bFunc = b0;
            pGate.pSop = Mio_SopRegister( (Mem_Flex_t *)pGate.pLib.pMmFlex, " 0\n" );
            pGate.pLib.pGate0 = pGate;
        
        } else if int32(strings.Compare(pGate.pForm, MIO_STRING_CONST1)) == 0 {
            pGate.bFunc = b1;
            pGate.pSop = Mio_SopRegister( (Mem_Flex_t *)pGate.pLib.pMmFlex, " 1\n" );
            pGate.pLib.pGate1 = pGate;
        
        } else {
            fmt.Printf("Cannot parse formula \"%s\" of gate \"%s\".\n", pGate.pForm, pGate.pName);
            return 1;
        }
        Cudd_Ref( pGate.bFunc );
        return 0;
    }

    // collect the names as they appear in the formula
    nPins = Mio_GateCollectNames( pGate.pForm, pPinNames );
    if nPins == 0 {
        fmt.Printf("Cannot read formula \"%s\" of gate \"%s\".\n", pGate.pForm, pGate.pName);
        return 1;
    }

    // set the number of inputs
    pGate.nInputs = nPins;

    // consider the case when all the pins have identical pin info
    if int32(strings.Compare(pGate.pPins.pName, "*")) == 0 {
        // get the topmost (generic) pin
        pPin = pGate.pPins;
        // inlining of macro ABC_FREE
        ((pPin.pName) ? (free((char *) (pPin.pName)), (pPin.pName) = 0) : 0);

        // create individual pins from the generic pin
        ppPin = &pPin.pNext;
        for  i = 1; i < nPins; i++  {
            // get the new pin
            *ppPin = Mio_PinDup( pPin );
            // set its name
            (*ppPin).pName = pPinNames[i];
            // prepare the next place in the list
            ppPin = &((*ppPin).pNext);
        }
        *ppPin = nil;

        // set the name of the topmost pin
        pPin.pName = pPinNames[0];
    
    } else {
        // reorder the variable names to appear the save way as the pins
        iPin = 0;
        // inlining of macro Mio_GateForEachPin
        for  pPin = Mio_GateReadPins(pGate);                                 (pPin != nil);                                                          pPin = Mio_PinReadNext(pPin)  {
            // find the pin with the name pPin->pName
            for  i = 0; i < nPins; i++  {
                if pPinNames[i] && int32(strings.Compare(pPinNames[i], pPin.pName)) == 0 {
                    // free pPinNames[i] because it is already available as pPin->pName
                    // setting pPinNames[i] to NULL is useful to make sure that
                    // this name is not assigned to two pins in the list
                    // inlining of macro ABC_FREE
                    ((pPinNames[i]) ? (free((char *) (pPinNames[i])), (pPinNames[i]) = 0) : 0);
                    pPinNamesCopy[iPin++] = pPin.pName;
                    break;
                }
                if i == nPins {
                    fmt.Printf("Cannot find pin name \"%s\" in the formula \"%s\" of gate \"%s\".\n", pPin.pName, pGate.pForm, pGate.pName);
                    return 1;
                }
            }
        }

        // check for the remaining names
        for  i = 0; i < nPins; i++  {
            if pPinNames[i] {
                fmt.Printf("Name \"%s\" appears in the formula \"%s\" of gate \"%s\" but there is no such pin.\n", pPinNames[i], pGate.pForm, pGate.pName);
                return 1;
            }
        }

        // copy the names back
        memcpy( pPinNames, pPinNamesCopy, nPins * sizeof(char *) );
    }

    // expand the manager if necessary
    if dd.size < nPins {
        Cudd_Quit( dd );
        dd = Cudd_Init( nPins + 10, 0, CUDD_UNIQUE_SLOTS, CUDD_CACHE_SLOTS, 0 );
        Cudd_zddVarsFromBddVars( dd, 2 );
    }

    // derive the formula as the BDD
    pGate.bFunc = Parse_FormulaParser( os.Stdout, pGate.pForm, nPins, 0, pPinNames, dd, dd.vars );
    if pGate.bFunc == nil {
        return 1; }
    Cudd_Ref( pGate.bFunc );

    // derive the cover (SOP)
    pGate.pSop = Abc_ConvertBddToSop( pGate.pLib.pMmFlex, dd, pGate.bFunc, pGate.bFunc, nPins, 0, pGate.pLib.vCube, -1 );

    // derive the truth table
    if pGate.nInputs <= 6 {
        pGate.uTruth = Mio_DeriveTruthTable6( pGate ); }

    return 0;
}

/**Function*************************************************************

  Synopsis    [Collect the pin names in the formula.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mio_GateCollectNames(pFormula string, pPinNames []string) int32 {
    var Buffer [1000]byte;
    var pTemp string;
    var nPins, i int32;

    // save the formula as it was
    strcpy( Buffer, pFormula );

    // remove the non-name symbols
    for  pTemp = Buffer; *pTemp; pTemp++  {
        if *pTemp == MIO_SYMB_AND  || *pTemp == MIO_SYMB_AND2   || 
             *pTemp == MIO_SYMB_OR   || *pTemp == MIO_SYMB_OR2    || 
             *pTemp == MIO_SYMB_XOR  || 
             *pTemp == MIO_SYMB_NOT  || *pTemp == MIO_SYMB_AFTNOT ||
             *pTemp == MIO_SYMB_OPEN || *pTemp == MIO_SYMB_CLOSE {
             *pTemp = ' '; }
    }

    // save the names
    nPins = 0;
    pTemp = strtok( Buffer, " " );
    for  pTemp  {
        for  i = 0; i < nPins; i++  {
            if int32(strings.Compare(pTemp, pPinNames[i])) == 0 {
                break; }
        }
        if i == nPins { // cannot find this name; save it
            pPinNames[nPins++] = Abc_UtilStrsav(pTemp);
        }
        // get the next name
        pTemp = strtok( nil, " " );
    }
    return nPins;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

