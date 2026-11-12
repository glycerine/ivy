package kit

import "fmt"

/**CFile****************************************************************

  FileName    [kitPla.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Computation kit.]

  Synopsis    [Manipulating SOP in the form of a C-string.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - Dec 6, 2006.]

  Revision    [$Id: kitPla.c,v 1.00 2006/12/06 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "kit.h"
//#include "aig/aig/aig.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Checks if the cover is constant 0.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Kit_PlaIsConst0(pSop string) int32 {
    return pSop[0] == ' ' && pSop[1] == '0';
}

/**Function*************************************************************

  Synopsis    [Checks if the cover is constant 1.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Kit_PlaIsConst1(pSop string) int32 {
    return pSop[0] == ' ' && pSop[1] == '1';
}

/**Function*************************************************************

  Synopsis    [Checks if the cover is a buffer.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Kit_PlaIsBuf(pSop string) int32 {
    if pSop[4] != 0 {
        return 0; }
    if (pSop[0] == '1' && pSop[2] == '1') || (pSop[0] == '0' && pSop[2] == '0') {
        return 1; }
    return 0;
}

/**Function*************************************************************

  Synopsis    [Checks if the cover is an inverter.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Kit_PlaIsInv(pSop string) int32 {
    if pSop[4] != 0 {
        return 0; }
    if (pSop[0] == '0' && pSop[2] == '1') || (pSop[0] == '1' && pSop[2] == '0') {
        return 1; }
    return 0;
}

/**Function*************************************************************

  Synopsis    [Reads the number of variables in the cover.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Kit_PlaGetVarNum(pSop string) int32 {
    var pCur string;
    for  pCur = pSop; *pCur != '\n'; pCur++  {
        if *pCur == 0 {
            return -1; }
    }
    return pCur - pSop - 2;
}

/**Function*************************************************************

  Synopsis    [Reads the number of cubes in the cover.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Kit_PlaGetCubeNum(pSop string) int32 {
    var pCur string;
    var nCubes int32;
    if pSop == nil {
        return 0; }
    for  pCur = pSop; *pCur; pCur++  {
        nCubes += (*pCur == '\n');
    }
    return nCubes;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Kit_PlaIsComplement(pSop string) int32 {
    var pCur string;
    for  pCur = pSop; *pCur; pCur++  {
        if *pCur == '\n' {
            return int32((*(pCur - 1) == '0' || *(pCur - 1) == 'n')); }
    }
    assert(false, "0");
    return 0;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Kit_PlaComplement(pSop string) {
    var pCur string;
    for  pCur = pSop; *pCur; pCur++  {
        if *pCur == '\n' {
            if *(pCur - 1) == '0' {
                *(pCur - 1) = '1';
            } else if *(pCur - 1) == '1' {
                *(pCur - 1) = '0';
            } else if *(pCur - 1) == 'x' {
                *(pCur - 1) = 'n';
            } else if *(pCur - 1) == 'n' {
                *(pCur - 1) = 'x';
            } else {
                assert(false, "0"); }
        }
    }
}

/**Function*************************************************************

  Synopsis    [Creates the constant 1 cover with the given number of variables and cubes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Kit_PlaStart(p any, nCubes int32, nVars int32) string {
    var pMan *Aig_MmFlex_t  = (Aig_MmFlex_t *)p;
    var pSopCover, pCube string;
    var i, Length int32;

    Length = nCubes * (nVars + 3);
    pSopCover = Aig_MmFlexEntryFetch( pMan, Length + 1 );
    memset( pSopCover, '-', (size_t)Length );
    pSopCover[Length] = 0;

    for  i = 0; i < nCubes; i++  {
        pCube = pSopCover + i * (nVars + 3);
        pCube[nVars + 0] = ' ';
        pCube[nVars + 1] = '1';
        pCube[nVars + 2] = '\n';
    }
    return pSopCover;
}

/**Function*************************************************************

  Synopsis    [Creates the cover from the ISOP computed from TT.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Kit_PlaCreateFromIsop(p any, nVars int32, vCover *Vec_Int_t) string {
    var pMan *Aig_MmFlex_t  = (Aig_MmFlex_t *)p;
    var pSop, pCube string;
    var i, k, Entry, Literal int32;
    assert(Vec_IntSize(vCover) > 0, "Vec_IntSize(vCover) > 0");
    if Vec_IntSize(vCover) == 0 {
        return nil; }
    // start the cover
    pSop = Kit_PlaStart( pMan, Vec_IntSize(vCover), nVars );
    // create cubes
    Vec_IntForEachEntry( vCover, Entry, i )
    {
        pCube = pSop + i * (nVars + 3);
        for  k = 0; k < nVars; k++  {
            Literal = 3 & (Entry >> (k << 1));
            if Literal == 1 {
                pCube[k] = '0';
            } else if Literal == 2 {
                pCube[k] = '1';
            } else if Literal != 0 {
                assert(false, "0"); }
        }
    }
    return pSop;
}

/**Function*************************************************************

  Synopsis    [Creates the cover from the ISOP computed from TT.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Kit_PlaToIsop(pSop string, vCover *Vec_Int_t) {
    var pCube string;
    var k, nVars, Entry int32;
    nVars = Kit_PlaGetVarNum( pSop );
    assert(nVars > 0, "nVars > 0");
    // create cubes
    Vec_IntClear( vCover );
    for  pCube = pSop; *pCube; pCube += nVars + 3  {
        Entry = 0;
        for  k = nVars - 1; k >= 0; k--  {
            if pCube[k] == '0' {
                Entry = (Entry << 2) | 1;
            } else if pCube[k] == '1' {
                Entry = (Entry << 2) | 2;
            } else if pCube[k] == '-' {
                Entry = (Entry << 2);
            } else { 
                assert(false, "0"); }
        }
        Vec_IntPush( vCover, Entry );
    }
}

/**Function*************************************************************

  Synopsis    [Allocates memory and copies the SOP into it.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Kit_PlaStoreSop(p any, pSop string) string {
    var pMan *Aig_MmFlex_t  = (Aig_MmFlex_t *)p;
    var pStore string;
    pStore = Aig_MmFlexEntryFetch( pMan, len(pSop) + 1 );
    pStore = pSop;
    return pStore;
}

/**Function*************************************************************

  Synopsis    [Transforms truth table into the SOP.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Kit_PlaFromTruth(p any, pTruth *uint32, nVars int32, vCover *Vec_Int_t) string {
    var pMan *Aig_MmFlex_t  = (Aig_MmFlex_t *)p;
    var pSop string;
    var RetValue int32;
    if Kit_TruthIsConst0(pTruth, nVars) {
        return Kit_PlaStoreSop( pMan, " 0\n" ); }
    if Kit_TruthIsConst1(pTruth, nVars) {
        return Kit_PlaStoreSop( pMan, " 1\n" ); }
    RetValue = Kit_TruthIsop( pTruth, nVars, vCover, 0 ); // 1 );
    assert(RetValue == 0 || RetValue == 1, "RetValue == 0 || RetValue == 1");
    pSop = Kit_PlaCreateFromIsop( pMan, nVars, vCover );
    if RetValue != 0 {
        Kit_PlaComplement( pSop ); }
    return pSop;
}


/**Function*************************************************************

  Synopsis    [Creates the cover from the ISOP computed from TT.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Kit_PlaFromIsop(vStr *Vec_Str_t, nVars int32, vCover *Vec_Int_t) string {
    var i, k, Entry, Literal int32;
    assert(Vec_IntSize(vCover) > 0, "Vec_IntSize(vCover) > 0");
    if Vec_IntSize(vCover) == 0 {
        return nil; }
    Vec_StrClear( vStr );
    Vec_IntForEachEntry( vCover, Entry, i )
    {
        for  k = 0; k < nVars; k++  {
            Literal = 3 & (Entry >> (k << 1));
            if Literal == 1 {
                Vec_StrPush( vStr, '0' );
            } else if Literal == 2 {
                Vec_StrPush( vStr, '1' );
            } else if Literal == 0 {
                Vec_StrPush( vStr, '-' );
            } else {
                assert(false, "0"); }
        }
        Vec_StrPush( vStr, ' ' );
        Vec_StrPush( vStr, '1' );
        Vec_StrPush( vStr, '\n' );
    }
    Vec_StrPush( vStr, '\0' );
    return Vec_StrArray( vStr );
}

/**Function*************************************************************

  Synopsis    [Creates the SOP from TT.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Kit_PlaFromTruthNew(pTruth *uint32, nVars int32, vCover *Vec_Int_t, vStr *Vec_Str_t) string {
    var pResult string;
    // transform truth table into the SOP
    var RetValue int32  = Kit_TruthIsop( pTruth, nVars, vCover, 1 );
    assert(RetValue == 0 || RetValue == 1, "RetValue == 0 || RetValue == 1");
    // check the case of constant cover
    if Vec_IntSize(vCover) == 0 || (Vec_IntSize(vCover) == 1 && Vec_IntEntry(vCover,0) == 0) {
        assert(RetValue == 0, "RetValue == 0");
        Vec_StrClear( vStr );
        Vec_StrAppend( vStr, (Vec_IntSize(vCover) == 0) ? " 0\n" : " 1\n" );
        Vec_StrPush( vStr, '\0' );
        return Vec_StrArray( vStr );
    }
    pResult = Kit_PlaFromIsop( vStr, nVars, vCover );
    if RetValue != 0 {
        Kit_PlaComplement( pResult ); }
    if nVars < 6 {
        assert(pTruth[0] == uint32(Kit_PlaToTruth6(pResult, nVars)), "pTruth[0] == uint32(Kit_PlaToTruth6(pResult, nVars))");
    } else if nVars == 6 {
        assert(*((ABC_UINT64_T*)pTruth) == Kit_PlaToTruth6(pResult, nVars), "*((ABC_UINT64_T*)pTruth) == Kit_PlaToTruth6(pResult, nVars)"); }
    return pResult;
}

/**Function*************************************************************

  Synopsis    [Converts SOP into a truth table.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Kit_PlaToTruth6(pSop string, nVars int32) ABC_UINT64_T {
    static ABC_UINT64_T Truth[8] = {
        ABC_CONST(0xAAAAAAAAAAAAAAAA),
        ABC_CONST(0xCCCCCCCCCCCCCCCC),
        ABC_CONST(0xF0F0F0F0F0F0F0F0),
        ABC_CONST(0xFF00FF00FF00FF00),
        ABC_CONST(0xFFFF0000FFFF0000),
        ABC_CONST(0xFFFFFFFF00000000),
        ABC_CONST(0x0000000000000000),
        ABC_CONST(0xFFFFFFFFFFFFFFFF)
    };
    var valueAnd ABC_UINT64_T;
    var valueOr ABC_UINT64_T = Truth[6];
    var v, lit int32;
    assert(nVars < 7, "nVars < 7");
    for {
        valueAnd = Truth[7];
        for  v = 0; v < nVars; v, lit = v+1, lit+1  {
            if pSop[lit] == '1' {
                valueAnd &=  Truth[v];
            } else if pSop[lit] == '0' {
                valueAnd &= ~Truth[v];
            } else if pSop[lit] != '-' {
                assert(false, "0"); }
        }
        valueOr |= valueAnd;
        assert(pSop[lit] == ' ', "pSop[lit] == ' '");
        lit++;
        lit++;
        assert(pSop[lit] == '\n', "pSop[lit] == '\\n'");
        lit++;
    
if !( pSop[lit] ) { break }
} 
    if Kit_PlaIsComplement(pSop) != 0 {
        valueOr = ~valueOr; }
    return valueOr;
}

/**Fnction*************************************************************

  Synopsis    [Converting SOP into a truth table.]

  Description [The SOP is represented as a C-string, as documented in 
  file "bblif.h". The truth table is returned as a bit-string composed 
  of 2^nVars bits. For functions of less than 6 variables, the full
  machine word is returned. (The truth table looks as if the function
  had 5 variables.) The use of this procedure should be limited to 
  Boolean functions with no more than 16 inputs.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Kit_PlaToTruth(pSop string, nVars int32, vVars *Vec_Ptr_t, pTemp *uint32, pTruth *uint32) {
    var v, c, nCubes, fCompl int32;
    assert(pSop != nil, "pSop != nil");
    assert(nVars >= 0, "nVars >= 0");
    if len(pSop) % (nVars + 3) != 0 {
        fmt.Printf("Kit_PlaToTruth(): SOP is represented incorrectly.\n");
        return;
    }
    // iterate through the cubes
    Kit_TruthClear( pTruth, nVars );
    nCubes =  len(pSop) / (nVars + 3);
    for  c = 0; c < nCubes; c++  {
        fCompl = (pSop[nVars+1] == '0');
        Kit_TruthFill( pTemp, nVars );
        // iterate through the literals of the cube
        for  v = 0; v < nVars; v++  {
            if pSop[v] == '1' {
                Kit_TruthAnd( pTemp, pTemp, (unsigned *)Vec_PtrEntry(vVars, v), nVars );
            } else if pSop[v] == '0' {
                Kit_TruthSharp( pTemp, pTemp, (unsigned *)Vec_PtrEntry(vVars, v), nVars ); }
        }
        // add cube to storage
        Kit_TruthOr( pTruth, pTruth, pTemp, nVars );
        // go to the next cube
        pSop += (nVars + 3);
    }
    if fCompl != 0 {
        Kit_TruthNot( pTruth, pTruth, nVars ); }
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

