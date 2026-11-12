package io

import (
	"fmt"
	"os"
)

/**CFile****************************************************************

  FileName    [ioReadDsd.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Command processing package.]

  Synopsis    [Procedure to read network from file.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: ioReadDsd.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "ioAbc.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Finds the end of the part.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_ReadDsdFindEnd(pCur string) string {
    var pEnd string;
    var nParts int32;
    assert(*pCur == '(', "*pCur == '('");
    for  pEnd = pCur; *pEnd; pEnd++  {
        if *pEnd == '(' {
            nParts++;
        } else if *pEnd == ')' {
            nParts--; }
        if nParts == 0 {
            return pEnd; }
    }
    return nil;
}

/**Function*************************************************************

  Synopsis    [Splits the formula into parts.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_ReadDsdStrSplit(pCur string, pParts []string, pTypeXor *int32) int32 {
    var fAnd, fXor, fPri, nParts int32;
    assert(*pCur, "*pCur");
    // process the parts
    for {
        // save the current part
        pParts[nParts++] = pCur;
        // skip the complement
        if *pCur == '!' {
            pCur++; }
        // skip var
        if *pCur >= 'a' && *pCur <= 'z' {
            pCur++;
        } else {
            // skip hex truth table
            for  (*pCur >= '0' && *pCur <= '9') || (*pCur >= 'A' && *pCur <= 'F')  {
                pCur++;
            }
            // process parentheses
            if *pCur != '(' {
                fmt.Printf("Cannot find the opening parenthesis.\n");
                break;
            }
            // find the corresponding closing parenthesis
            pCur = Io_ReadDsdFindEnd( pCur );
            if pCur == nil {
                fmt.Printf("Cannot find the closing parenthesis.\n");
                break;
            }
            pCur++;
        }
        // check the end
        if *pCur == 0 {
            break; }
        // check symbol
        if *pCur != '*' && *pCur != '+' && *pCur != ',' {
            fmt.Printf("Wrong separating symbol.\n");
            break;
        }
        // remember the symbol
        fAnd |= (*pCur == '*');
        fXor |= (*pCur == '+');
        fPri |= (*pCur == ',');
        *pCur++ = 0;
    }
    // check separating symbols
    if fAnd + fXor + fPri > 1 {
        fmt.Printf("Different types of separating symbol ennPartsed.\n");
        return 0;
    }
    *pTypeXor = fXor;
    return nParts;
}

/**Function*************************************************************

  Synopsis    [Recursively parses the formula.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_ReadDsd_rec(pNtk *Abc_Ntk_t, pCur string, pSop string) *Abc_Obj_t {
    var pObj, pFanin *Abc_Obj_t;
    var pEnd string;
    var pParts [32]string;
    var i, nParts, TypeExor int32;

    // consider complemented formula
    if *pCur == '!' {
        pObj = Io_ReadDsd_rec( pNtk, pCur + 1, nil );
        return Abc_NtkCreateNodeInv( pNtk, pObj );
    }
    if *pCur == '(' {
        assert(pCur[len(pCur)-1] == ')', "pCur[strlen(pCur)-1] == ')'");
        pCur[len(pCur)-1] = 0;
        nParts = Io_ReadDsdStrSplit( pCur+1, pParts, &TypeExor );
        if nParts == 0 {
            Abc_NtkDelete( pNtk );
            return nil;
        }
        pObj = Abc_NtkCreateNode( pNtk );
        if pSop {
//            for ( i = nParts - 1; i >= 0; i-- )
            for  i = 0; i < nParts; i++  {
                pFanin = Io_ReadDsd_rec( pNtk, pParts[i], nil );
                if pFanin == nil {
                    return nil; }
                Abc_ObjAddFanin( pObj, pFanin );
            }
        
        } else {
            for  i = 0; i < nParts; i++  {
                pFanin = Io_ReadDsd_rec( pNtk, pParts[i], nil );
                if pFanin == nil {
                    return nil; }
                Abc_ObjAddFanin( pObj, pFanin );
            }
        }
        if pSop {
            pObj.pData = Abc_SopRegister( (Mem_Flex_t *)pNtk.pManFunc, pSop );
        } else if TypeExor != 0 {
            pObj.pData = Abc_SopCreateXorSpecial( (Mem_Flex_t *)pNtk.pManFunc, nParts );
        } else {
            pObj.pData = Abc_SopCreateAnd( (Mem_Flex_t *)pNtk.pManFunc, nParts, nil ); }
        return pObj;
    }
    if *pCur >= 'a' && *pCur <= 'z' {
        assert(*(pCur+1) == 0, "*(pCur+1) == 0");
        return Abc_NtkPi( pNtk, *pCur - 'a' );
    }

    // skip hex truth table
    pEnd = pCur;
    for  (*pEnd >= '0' && *pEnd <= '9') || (*pEnd >= 'A' && *pEnd <= 'F')  {
        pEnd++;
    }
    if *pEnd != '(' {
        fmt.Printf("Cannot find the end of hexidecimal truth table.\n");
        return nil;
    }

    // parse the truth table
    *pEnd = 0;
    pSop = Abc_SopFromTruthHex( pCur );
    *pEnd = '(';
    pObj = Io_ReadDsd_rec( pNtk, pEnd, pSop );
    // inlining of macro ABC_FREE
    ((pSop) ? (free((char *) (pSop)), (pSop) = 0) : 0);
    return pObj;
}

/**Function*************************************************************

  Synopsis    [Derives the DSD network of the formula.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_ReadDsd(pForm string) *Abc_Ntk_t {
    var pNtk *Abc_Ntk_t;
    var pObj, pTop *Abc_Obj_t;
    var vNames *Vec_Ptr_t;
    var pCur, pFormCopy string;
    var i, nInputs int32;

    // count the number of elementary variables
    nInputs = 0;
    for  pCur = pForm; *pCur; pCur++  {
        if *pCur >= 'a' && *pCur <= 'z' {
            nInputs = Abc_MaxInt( nInputs, *pCur - 'a' ); }
    }
    nInputs++;

    // create the network
    pNtk = Abc_NtkAlloc( ABC_NTK_LOGIC, ABC_FUNC_SOP, 1 );
    pNtk.pName = Extra_UtilStrsav( "dsd" );

    // create PIs
    vNames = Abc_NodeGetFakeNames( nInputs );
    for  i = 0; i < nInputs; i++  {
        Abc_ObjAssignName( Abc_NtkCreatePi(pNtk), (char *)Vec_PtrEntry(vNames, i), nil );
    }
    Abc_NodeFreeNames( vNames );

    // transform the formula by inserting parentheses
    // this transforms strings like PRIME(a,b,cd) into (PRIME((a),(b),(cd)))
    // inlining of macro ABC_ALLOC
    pCur = pFormCopy = ((char *) malloc(sizeof(char) * (size_t)(3 * len(pForm) + 10)));
    *pCur++ = '(';
    for  ; *pForm; pForm++  {
        if *pForm == '(' {
            *pCur++ = '(';
            *pCur++ = '(';
        
        } else if *pForm == ')' {
            *pCur++ = ')';
            *pCur++ = ')';
        
        } else if *pForm == ',' {
            *pCur++ = ')';
            *pCur++ = ',';
            *pCur++ = '(';
        
        } else {
            *pCur++ = *pForm; }
    }
    *pCur++ = ')';
    *pCur = 0;

    // parse the formula
    pObj = Io_ReadDsd_rec( pNtk, pFormCopy, nil );
    // inlining of macro ABC_FREE
    ((pFormCopy) ? (free((char *) (pFormCopy)), (pFormCopy) = 0) : 0);
    if pObj == nil {
        return nil; }

    // create output
    pTop = Abc_NtkCreatePo(pNtk);
    Abc_ObjAssignName( pTop, "F", nil );
    Abc_ObjAddFanin( pTop, pObj );

    // create the only PO
    if !Abc_NtkCheck( pNtk ) {
        fmt.Fprintf(os.Stdout, "Io_ReadDsd(): Network check has failed.\n");
        Abc_NtkDelete( pNtk );
        return nil;
    }
    return pNtk;
}



////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////



//ABC_NAMESPACE_IMPL_END

