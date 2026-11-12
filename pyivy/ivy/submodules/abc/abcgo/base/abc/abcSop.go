package abc

import (
	"fmt"
	"os"
)

/**CFile****************************************************************

  FileName    [abcSop.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Network and node package.]

  Synopsis    [Implementation of a simple SOP representation of nodes.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: abcSop.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "abc.h"
//#include "bool/kit/kit.h"

//ABC_NAMESPACE_IMPL_START


/* 
    The SOPs in this package are represented using char * strings.
    For example, the SOP of the node: 

       .names c d0 d1 MUX
       01- 1
       1-1 1

    is the string: "01- 1\n1-1 1\n" where '\n' is a single char.
*/

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Registers the cube string with the network.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SopRegister(pMan *Mem_Flex_t, pName string) string {
    var pRegName string;
    if pName == nil { return nil; }
    pRegName = Mem_FlexEntryFetch( pMan, len(pName) + 1 );
    pRegName = pName;
    return pRegName;
}

/**Function*************************************************************

  Synopsis    [Creates the constant 1 cover with the given number of variables and cubes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SopStart(pMan *Mem_Flex_t, nCubes int32, nVars int32) string {
    var pSopCover, pCube string;
    var i, Length int32;

    Length = nCubes * (nVars + 3);
    pSopCover = Mem_FlexEntryFetch( pMan, Length + 1 );
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

  Synopsis    [Creates the constant 1 cover with 0 variables.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SopCreateConst1(pMan *Mem_Flex_t) string {
    return Abc_SopRegister( pMan, " 1\n" );
}

/**Function*************************************************************

  Synopsis    [Creates the constant 1 cover with 0 variables.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SopCreateConst0(pMan *Mem_Flex_t) string {
    return Abc_SopRegister( pMan, " 0\n" );
}

/**Function*************************************************************

  Synopsis    [Creates the AND2 cover.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SopCreateAnd2(pMan *Mem_Flex_t, fCompl0 int32, fCompl1 int32) string {
    var Buffer [6]byte;
    Buffer[0] = '1' - fCompl0;
    Buffer[1] = '1' - fCompl1;
    Buffer[2] = ' ';
    Buffer[3] = '1';
    Buffer[4] = '\n';
    Buffer[5] = 0;
    return Abc_SopRegister( pMan, Buffer );
}

/**Function*************************************************************

  Synopsis    [Creates the multi-input AND cover.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SopCreateAnd(pMan *Mem_Flex_t, nVars int32, pfCompl *int32) string {
    var pSop string;
    var i int32;
    pSop = Abc_SopStart( pMan, 1, nVars );
    for  i = 0; i < nVars; i++  {
        pSop[i] = '1' - (pfCompl? pfCompl[i] : 0);
    }
    pSop[nVars + 1] = '1';
    return pSop;
}

/**Function*************************************************************

  Synopsis    [Creates the multi-input NAND cover.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SopCreateNand(pMan *Mem_Flex_t, nVars int32) string {
    var pSop string;
    var i int32;
    pSop = Abc_SopStart( pMan, 1, nVars );
    for  i = 0; i < nVars; i++  {
        pSop[i] = '1';
    }
    pSop[nVars + 1] = '0';
    return pSop;
}

/**Function*************************************************************

  Synopsis    [Creates the multi-input OR cover.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SopCreateOr(pMan *Mem_Flex_t, nVars int32, pfCompl *int32) string {
    var pSop string;
    var i int32;
    pSop = Abc_SopStart( pMan, 1, nVars );
    for  i = 0; i < nVars; i++  {
        pSop[i] = '0' + (pfCompl? pfCompl[i] : 0);
    }
    pSop[nVars + 1] = '0';
    return pSop;
}

/**Function*************************************************************

  Synopsis    [Creates the multi-input OR cover.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SopCreateOrMultiCube(pMan *Mem_Flex_t, nVars int32, pfCompl *int32) string {
    var pSop, pCube string;
    var i int32;
    pSop = Abc_SopStart( pMan, nVars, nVars );
    i = 0;
    Abc_SopForEachCube( pSop, nVars, pCube )
    {
        pCube[i] = '1' - (pfCompl? pfCompl[i] : 0);
        i++;
    }
    return pSop;
}

/**Function*************************************************************

  Synopsis    [Creates the multi-input NOR cover.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SopCreateNor(pMan *Mem_Flex_t, nVars int32) string {
    var pSop string;
    var i int32;
    pSop = Abc_SopStart( pMan, 1, nVars );
    for  i = 0; i < nVars; i++  {
        pSop[i] = '0';
    }
    return pSop;
}

/**Function*************************************************************

  Synopsis    [Creates the multi-input XOR cover.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SopCreateXor(pMan *Mem_Flex_t, nVars int32) string {
    assert(nVars == 2, "nVars == 2");
    return Abc_SopRegister(pMan, "01 1\n10 1\n");
}

/**Function*************************************************************

  Synopsis    [Creates the multi-input XOR cover (special case).]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SopCreateXorSpecial(pMan *Mem_Flex_t, nVars int32) string {
    var pSop string;
    pSop = Abc_SopCreateAnd( pMan, nVars, nil );
    pSop[nVars+1] = 'x';
    assert(pSop[nVars+2] == '\n', "pSop[nVars+2] == '\\n'");
    return pSop;
}

/**Function*************************************************************

  Synopsis    [Creates the multi-input XNOR cover.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SopCreateNxor(pMan *Mem_Flex_t, nVars int32) string {
    assert(nVars == 2, "nVars == 2");
    return Abc_SopRegister(pMan, "11 1\n00 1\n");
}

/**Function*************************************************************

  Synopsis    [Creates the MUX cover.]

  Description [The first input of MUX is the control. The second input
  is DATA1. The third input is DATA0.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SopCreateMux(pMan *Mem_Flex_t) string {
    return Abc_SopRegister(pMan, "11- 1\n0-1 1\n");
}

/**Function*************************************************************

  Synopsis    [Creates the inv cover.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SopCreateInv(pMan *Mem_Flex_t) string {
    return Abc_SopRegister(pMan, "0 1\n");
}

/**Function*************************************************************

  Synopsis    [Creates the buf cover.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SopCreateBuf(pMan *Mem_Flex_t) string {
    return Abc_SopRegister(pMan, "1 1\n");
}

/**Function*************************************************************

  Synopsis    [Creates the arbitrary cover from the truth table.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SopCreateFromTruth(pMan *Mem_Flex_t, nVars int32, pTruth *uint32) string {
    var pSop, pCube string;
    var nMints, Counter, i, k int32;
    if nVars == 0 {
        if pTruth[0] {
            return Abc_SopCreateConst1(pMan);
        } else {
            return Abc_SopCreateConst0(pMan);
        } }
    // count the number of true minterms
    Counter = 0;
    nMints = (1 << nVars);
    for  i = 0; i < nMints; i++  {
        Counter += ((pTruth[i>>5] & (1 << (i&31))) > 0);
    }
    // SOP is not well-defined if the truth table is constant 0
    assert(Counter > 0, "Counter > 0");
    if Counter == 0 {
        return nil; }
    // start the cover
    pSop = Abc_SopStart( pMan, Counter, nVars );
    // create true minterms
    Counter = 0;
    for  i = 0; i < nMints; i++  {
        if (pTruth[i>>5] & (1 << (i&31))) > 0 {
            pCube = pSop + Counter * (nVars + 3);
            for  k = 0; k < nVars; k++  {
                pCube[k] = '0' + ((i & (1 << k)) > 0);
            }
            Counter++;
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
func Abc_SopCreateFromIsop(pMan *Mem_Flex_t, nVars int32, vCover *Vec_Int_t) string {
    var pSop, pCube string;
    var i, k, Entry, Literal int32;
    assert(Vec_IntSize(vCover) > 0, "Vec_IntSize(vCover) > 0");
    if Vec_IntSize(vCover) == 0 {
        return nil; }
    // start the cover
    pSop = Abc_SopStart( pMan, Vec_IntSize(vCover), nVars );
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
func Abc_SopCreateFromTruthIsop(pMan *Mem_Flex_t, nVars int32, pTruth *word, vCover *Vec_Int_t) string {
    var pSop string  = nil;
    var w int32;
    var nWords int32 = Abc_Truth6WordNum( nVars );
    assert(nVars < 16, "nVars < 16");

    for  w = 0; w < nWords; w++  {
        if pTruth[w] {
            break; }
    }
    if w == nWords {
        return Abc_SopRegister( pMan, " 0\n" ); }

    for  w = 0; w < nWords; w++  {
        if ~pTruth[w] {
            break; }
    }
    if w == nWords {
        return Abc_SopRegister( pMan, " 1\n" ); }

    {
        var RetValue int32  = Kit_TruthIsop( (unsigned *)pTruth, nVars, vCover, 1 );
        assert(nVars > 0, "nVars > 0");
        assert(RetValue == 0 || RetValue == 1, "RetValue == 0 || RetValue == 1");
        pSop = Abc_SopCreateFromIsop( pMan, nVars, vCover );
        if RetValue != 0 {
            Abc_SopComplement( pSop ); }
    }
    return pSop;
}

/**Function*************************************************************

  Synopsis    [Creates the cover from the ISOP computed from TT.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SopToIsop(pSop string, vCover *Vec_Int_t) {
    var pCube string;
    var k, nVars, Entry int32;
    nVars = Abc_SopGetVarNum( pSop );
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

  Synopsis    [Reads the number of cubes in the cover.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SopGetCubeNum(pSop string) int32 {
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

  Synopsis    [Reads the number of SOP literals in the cover.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SopGetLitNum(pSop string) int32 {
    var pCur string;
    var nLits int32;
    if pSop == nil {
        return 0; }
    for  pCur = pSop; *pCur; pCur++  {
        nLits  -= (*pCur == '\n');
        nLits  += (*pCur == '0' || *pCur == '1');
    }
    return nLits;
}

/**Function*************************************************************

  Synopsis    [Reads the number of variables in the cover.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SopGetVarNum(pSop string) int32 {
    var pCur string;
    for  pCur = pSop; *pCur != '\n'; pCur++  {
        if *pCur == 0 {
            return -1; }
    }
    return pCur - pSop - 2;
}

/**Function*************************************************************

  Synopsis    [Reads the phase of the cover.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SopGetPhase(pSop string) int32 {
    var nVars int32  = Abc_SopGetVarNum( pSop );
    if pSop[nVars+1] == '0' || pSop[nVars+1] == 'n' {
        return 0; }
    if pSop[nVars+1] == '1' || pSop[nVars+1] == 'x' {
        return 1; }
    assert(false, "0");
    return -1;
}

/**Function*************************************************************

  Synopsis    [Returns the i-th literal of the cover.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SopGetIthCareLit(pSop string, i int32) int32 {
    var pCube string;
    var nVars int32;
    nVars = Abc_SopGetVarNum( pSop );
    Abc_SopForEachCube( pSop, nVars, pCube )
        if ( pCube[i] != '-' )
            return pCube[i] - '0';
    return -1;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SopComplement(pSop string) {
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

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SopComplementVar(pSop string, iVar int32) {
    var pCube string;
    var nVars int32  = Abc_SopGetVarNum(pSop);
    assert(iVar < nVars, "iVar < nVars");
    Abc_SopForEachCube( pSop, nVars, pCube )
    {
        if pCube[iVar] == '0' {
            pCube[iVar] = '1';
        } else if pCube[iVar] == '1' {
            pCube[iVar] = '0'; }
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SopIsComplement(pSop string) int32 {
    var pCur string;
    for  pCur = pSop; *pCur; pCur++  {
        if *pCur == '\n' {
            return int32((*(pCur - 1) == '0' || *(pCur - 1) == 'n')); }
    }
    assert(false, "0");
    return 0;
}

/**Function*************************************************************

  Synopsis    [Checks if the cover is constant 0.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SopIsConst0(pSop string) int32 {
    return pSop[0] == ' ' && pSop[1] == '0';
}

/**Function*************************************************************

  Synopsis    [Checks if the cover is constant 1.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SopIsConst1(pSop string) int32 {
    return pSop[0] == ' ' && pSop[1] == '1';
}

/**Function*************************************************************

  Synopsis    [Checks if the cover is constant 1.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SopIsBuf(pSop string) int32 {
    if pSop[4] != 0 {
        return 0; }
    if (pSop[0] == '1' && pSop[2] == '1') || (pSop[0] == '0' && pSop[2] == '0') {
        return 1; }
    return 0;
}

/**Function*************************************************************

  Synopsis    [Checks if the cover is constant 1.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SopIsInv(pSop string) int32 {
    if pSop[4] != 0 {
        return 0; }
    if (pSop[0] == '0' && pSop[2] == '1') || (pSop[0] == '1' && pSop[2] == '0') {
        return 1; }
    return 0;
}

/**Function*************************************************************

  Synopsis    [Checks if the cover is AND with possibly complemented inputs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SopIsAndType(pSop string) int32 {
    var pCur string;
    if Abc_SopGetCubeNum(pSop) != 1 {
        return 0; }
    for  pCur = pSop; *pCur != ' '; pCur++  {
        if *pCur == '-' {
            return 0; }
    }
    if pCur[1] != '1' {
        return 0; }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Checks if the cover is OR with possibly complemented inputs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SopIsOrType(pSop string) int32 {
    var pCube, pCur string;
    var nVars, nLits int32;
    nVars = Abc_SopGetVarNum( pSop );
    if nVars != Abc_SopGetCubeNum(pSop) {
        return 0; }
    Abc_SopForEachCube( pSop, nVars, pCube )
    {
        // count the number of literals in the cube
        nLits = 0;
        for  pCur = pCube; *pCur != ' '; pCur++  {
            nLits += ( *pCur != '-' );
        }
        if nLits != 1 {
            return 0; }
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SopIsExorType(pSop string) int32 {
    var pCur string;
    for  pCur = pSop; *pCur; pCur++  {
        if *pCur == '\n' {
            return int32((*(pCur - 1) == 'x' || *(pCur - 1) == 'n')); }
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
func Abc_SopCheck(pSop string, nFanins int32) int32 {
    var pCubes, pCubesOld string;
    var fFound0, fFound1 int32;

    // check the logic function of the node
    for  pCubes = pSop; *pCubes; pCubes++  {
        // get the end of the next cube
        for  pCubesOld = pCubes; *pCubes != ' '; pCubes++  {}
        // compare the distance
        if pCubes - pCubesOld != nFanins {
            fmt.Fprintf(os.Stdout, "Abc_SopCheck: SOP has a mismatch between its cover size (%d) and its fanin number (%d).\n", int32((ABC_PTRDIFF_T)(pCubes - pCubesOld)), nFanins);
            return 0;
        }
        // check the output values for this cube
        pCubes++;
        if *pCubes == '0' {
            fFound0 = 1;
        } else if *pCubes == '1' {
            fFound1 = 1;
        } else if *pCubes != 'x' && *pCubes != 'n' {
            fmt.Fprintf(os.Stdout, "Abc_SopCheck: SOP has a strange character (%c) in the output part of its cube.\n", *pCubes);
            return 0;
        }
        // check the last symbol (new line)
        pCubes++;
        if *pCubes != '\n' {
            fmt.Fprintf(os.Stdout, "Abc_SopCheck: SOP has a cube without new line in the end.\n");
            return 0;
        }
    }
    if fFound0 && (fFound1 != 0) {
        fmt.Fprintf(os.Stdout, "Abc_SopCheck: SOP has cubes in both phases.\n");
        return 0;
    }
    return 1;
}


/**Function*************************************************************

  Synopsis    [Derives SOP from the truth table representation.]

  Description [Truth table is expected to be in the hexadecimal notation.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SopFromTruthBin(pTruth string) string {
    var pSopCover, pCube string;
    var nTruthSize, nVars, Digit, Length, Mint, i, b int32;
    var vMints *Vec_Int_t;

    // get the number of variables
    nTruthSize = len(pTruth);
    nVars = Abc_Base2Log( nTruthSize );
    if nTruthSize != (1 << (nVars)) {
        fmt.Printf("String %s does not look like a truth table of a %d-variable function.\n", pTruth, nVars);
        return nil;
    }

    // collect the on-set minterms
    vMints = Vec_IntAlloc( 100 );
    for  i = 0; i < nTruthSize; i++  {
        if pTruth[i] >= '0' && pTruth[i] <= '1' {
            Digit = pTruth[i] - '0';
        } else {
            Vec_IntFree( vMints );
            fmt.Printf("String %s does not look like a binary representation of the truth table.\n", pTruth);
            return nil;
        }
        if Digit == 1 {
            Vec_IntPush( vMints, nTruthSize - 1 - i ); }
    }
    if Vec_IntSize( vMints ) == 0 || Vec_IntSize( vMints ) == nTruthSize {
        Vec_IntFree( vMints );
        fmt.Printf("Cannot create constant function.\n");
        return nil;
    }

    // create the SOP representation of the minterms
    Length = Vec_IntSize(vMints) * (nVars + 3);
    pSopCover = ABC_ALLOC( char, Length + 1 );
    pSopCover[Length] = 0;
    Vec_IntForEachEntry( vMints, Mint, i )
    {
        pCube = pSopCover + i * (nVars + 3);
        for  b = 0; b < nVars; b++  {
            if Mint & (1 << (nVars-1-b)) {
//            if ( Mint & (1 << b) )
                pCube[b] = '1';
            } else {
                pCube[b] = '0'; }
        }
        pCube[nVars + 0] = ' ';
        pCube[nVars + 1] = '1';
        pCube[nVars + 2] = '\n';
    }
    Vec_IntFree( vMints );
    return pSopCover;
}

/**Function*************************************************************

  Synopsis    [Derives SOP from the truth table representation.]

  Description [Truth table is expected to be in the hexadecimal notation.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SopFromTruthHex(pTruth string) string {
    var pSopCover, pCube string;
    var nTruthSize, nVars, Digit, Length, Mint, i, b int32;
    var vMints *Vec_Int_t;

    // get the number of variables
    nTruthSize = len(pTruth);
    if nTruthSize < 2 {
        nVars = 2;
    } else {
        nVars = Abc_Base2Log(nTruthSize) + 2;
    }
    if nTruthSize != (1 << (nVars-2)) {
        fmt.Printf("String %s does not look like a truth table of a %d-variable function.\n", pTruth, nVars);
        return nil;
    }

    // collect the on-set minterms
    vMints = Vec_IntAlloc( 100 );
    for  i = 0; i < nTruthSize; i++  {
        if pTruth[i] >= '0' && pTruth[i] <= '9' {
            Digit = pTruth[i] - '0';
        } else if pTruth[i] >= 'a' && pTruth[i] <= 'f' {
            Digit = 10 + pTruth[i] - 'a';
        } else if pTruth[i] >= 'A' && pTruth[i] <= 'F' {
            Digit = 10 + pTruth[i] - 'A';
        } else {
            fmt.Printf("String %s does not look like a hexadecimal representation of the truth table.\n", pTruth);
            return nil;
        }
        for  b = 0; b < 4; b++  {
            if Digit & (1 << b) {
                Vec_IntPush( vMints, 4*(nTruthSize-1-i)+b ); }
        }
    }

    // create the SOP representation of the minterms
    Length = Vec_IntSize(vMints) * (nVars + 3);
    pSopCover = ABC_ALLOC( char, Length + 1 );
    pSopCover[Length] = 0;
    Vec_IntForEachEntry( vMints, Mint, i )
    {
        pCube = pSopCover + i * (nVars + 3);
        for  b = 0; b < nVars; b++  {
//            if ( Mint & (1 << (nVars-1-b)) )
            if Mint & (1 << b) {
                pCube[b] = '1';
            } else {
                pCube[b] = '0'; }
        }
        pCube[nVars + 0] = ' ';
        pCube[nVars + 1] = '1';
        pCube[nVars + 2] = '\n';
    }
/*
    // create TT representation
    {
        extern void Bdc_ManDecomposeTest( unsigned uTruth, int nVars );
        unsigned uTruth = 0;
        int nVarsAll = 4;
        assert( nVarsAll == 4 );
        assert( nVars <= nVarsAll );
        Vec_IntForEachEntry( vMints, Mint, i )
            uTruth |= (1 << Mint);
//        uTruth = uTruth | (uTruth << 8) | (uTruth << 16) | (uTruth << 24);
        uTruth = uTruth | (uTruth << 16);
        Bdc_ManDecomposeTest( uTruth, nVarsAll );
    }
*/
    Vec_IntFree( vMints );
    return pSopCover;
}

/**Function*************************************************************

  Synopsis    [Creates one encoder node.]

  Description [Produces MV-SOP for BLIF-MV representation.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SopEncoderPos(pMan *Mem_Flex_t, iValue int32, nValues int32) string {
    var Buffer string // was [32]byte in C

    assert(iValue < nValues, "iValue < nValues");
    Buffer = fmt.Sprintf("d0\n%d 1\n", iValue);
    return Abc_SopRegister( pMan, Buffer );
}

/**Function*************************************************************

  Synopsis    [Creates one encoder node.]

  Description [Produces MV-SOP for BLIF-MV representation.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SopEncoderLog(pMan *Mem_Flex_t, iBit int32, nValues int32) string {
    var pResult string;
    var vSop *Vec_Str_t;
    var v int32;
    var Counter int32;
    var fFirst int32 = 1;
    var nBits int32 = Abc_Base2Log(nValues);
    assert(iBit < nBits, "iBit < nBits");
    // count the number of literals
    Counter = 0;
    for  v = 0; v < nValues; v++  {
        Counter += ( (v & (1 << iBit)) > 0 );
    }
    // create the cover
    vSop = Vec_StrAlloc( 100 );
    Vec_StrPrintStr( vSop, "d0\n" );
    if Counter > 1 {
        Vec_StrPrintStr( vSop, "(" ); }
    for  v = 0; v < nValues; v++  {
        if v & (1 << iBit) {
            if fFirst != 0 {
                fFirst = 0;
            } else {
                Vec_StrPush( vSop, ',' ); }
            Vec_StrPrintNum( vSop, v );
        }
    }
    if Counter > 1 {
        Vec_StrPrintStr( vSop, ")" ); }
    Vec_StrPrintStr( vSop, " 1\n" );
    Vec_StrPush( vSop, 0 );
    pResult = Abc_SopRegister( pMan, Vec_StrArray(vSop) );
    Vec_StrFree( vSop );
    return pResult;
}

/**Function*************************************************************

  Synopsis    [Creates the decoder node.]

  Description [Produces MV-SOP for BLIF-MV representation.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SopDecoderPos(pMan *Mem_Flex_t, nValues int32) string {
    var pResult string;
    var vSop *Vec_Str_t;
    var i, k int32;
    assert(nValues > 1, "nValues > 1");
    vSop = Vec_StrAlloc( 100 );
    for  i = 0; i < nValues; i++  {
        for  k = 0; k < nValues; k++  {
            if k == i {
                Vec_StrPrintStr( vSop, "1 " );
            } else {
                Vec_StrPrintStr( vSop, "- " ); }
        }
        Vec_StrPrintNum( vSop, i );
        Vec_StrPush( vSop, '\n' );
    }
    Vec_StrPush( vSop, 0 );
    pResult = Abc_SopRegister( pMan, Vec_StrArray(vSop) );
    Vec_StrFree( vSop );
    return pResult;
}

/**Function*************************************************************

  Synopsis    [Creates the decover node.]

  Description [Produces MV-SOP for BLIF-MV representation.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SopDecoderLog(pMan *Mem_Flex_t, nValues int32) string {
    var pResult string;
    var vSop *Vec_Str_t;
    var i int32;
    var b int32;
    var nBits int32 = Abc_Base2Log(nValues);
    assert(nValues > 1 && nValues <= (1<<nBits), "nValues > 1 && nValues <= (1<<nBits)");
    vSop = Vec_StrAlloc( 100 );
    for  i = 0; i < nValues; i++  {
        for  b = 0; b < nBits; b++  {
            Vec_StrPrintNum( vSop, int32(((i & (1 << b)) > 0)) );
            Vec_StrPush( vSop, ' ' );
        }
        Vec_StrPrintNum( vSop, i );
        Vec_StrPush( vSop, '\n' );
    }
    Vec_StrPush( vSop, 0 );
    pResult = Abc_SopRegister( pMan, Vec_StrArray(vSop) );
    Vec_StrFree( vSop );
    return pResult;
}

/**Function*************************************************************

  Synopsis    [Computes truth table of the node.]
 
  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SopToTruth(pSop string, nInputs int32) word {
    static word Truth[8] = {
        ABC_CONST(0xAAAAAAAAAAAAAAAA),
        ABC_CONST(0xCCCCCCCCCCCCCCCC),
        ABC_CONST(0xF0F0F0F0F0F0F0F0),
        ABC_CONST(0xFF00FF00FF00FF00),
        ABC_CONST(0xFFFF0000FFFF0000),
        ABC_CONST(0xFFFFFFFF00000000),
        ABC_CONST(0x0000000000000000),
        ABC_CONST(0xFFFFFFFFFFFFFFFF)
    };
    var Cube, Result word;
    var v, lit int32;
    var nVars int32  = Abc_SopGetVarNum(pSop);
    assert(nVars >= 0 && nVars <= 6, "nVars >= 0 && nVars <= 6");
    assert(nVars == nInputs, "nVars == nInputs");
    for {
        Cube = Truth[7];
        for  v = 0; v < nVars; v, lit = v+1, lit+1  {
            if pSop[lit] == '1' {
                Cube &=  Truth[v];
            } else if pSop[lit] == '0' {
                Cube &= ~Truth[v];
            } else if pSop[lit] != '-' {
                assert(false, "0"); }
        }
        Result |= Cube;
        assert(pSop[lit] == ' ', "pSop[lit] == ' '");
        lit++;
        lit++;
        assert(pSop[lit] == '\n', "pSop[lit] == '\\n'");
        lit++;
    
if !( pSop[lit] ) { break }
} 
    if Abc_SopIsComplement(pSop) != 0 {
        Result = ~Result; }
    return Result;
}

/**Function*************************************************************

  Synopsis    [Computes truth table of the node.]
 
  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SopToTruth7(pSop string, nInputs int32, r []word) {
    static word Truth[7][2] = {
        {ABC_CONST(0xAAAAAAAAAAAAAAAA),ABC_CONST(0xAAAAAAAAAAAAAAAA)},
        {ABC_CONST(0xCCCCCCCCCCCCCCCC),ABC_CONST(0xCCCCCCCCCCCCCCCC)},
        {ABC_CONST(0xF0F0F0F0F0F0F0F0),ABC_CONST(0xF0F0F0F0F0F0F0F0)},
        {ABC_CONST(0xFF00FF00FF00FF00),ABC_CONST(0xFF00FF00FF00FF00)},
        {ABC_CONST(0xFFFF0000FFFF0000),ABC_CONST(0xFFFF0000FFFF0000)},
        {ABC_CONST(0xFFFFFFFF00000000),ABC_CONST(0xFFFFFFFF00000000)},
        {ABC_CONST(0x0000000000000000),ABC_CONST(0xFFFFFFFFFFFFFFFF)},
    };
    var Cube [2]word;
    var v, lit int32;
    var nVars int32  = Abc_SopGetVarNum(pSop);
    assert(nVars >= 0 && nVars <= 7, "nVars >= 0 && nVars <= 7");
    assert(nVars == nInputs, "nVars == nInputs");
    r[0] = r[1] = 0;
    for {
        Cube[0] = Cube[1] = ~(word)0;
        for  v = 0; v < nVars; v, lit = v+1, lit+1  {
            if pSop[lit] == '1' {
                Cube[0] &=  Truth[v][0];
                Cube[1] &=  Truth[v][1];
            
            } else if pSop[lit] == '0' {
                Cube[0] &= ~Truth[v][0];
                Cube[1] &= ~Truth[v][1];
            
            } else if pSop[lit] != '-' {
                assert(false, "0"); }
        }
        r[0] |= Cube[0];
        r[1] |= Cube[1];
        assert(pSop[lit] == ' ', "pSop[lit] == ' '");
        lit++;
        lit++;
        assert(pSop[lit] == '\n', "pSop[lit] == '\\n'");
        lit++;
    
if !( pSop[lit] ) { break }
} 
    if Abc_SopIsComplement(pSop) != 0 {
        r[0] = ~r[0];
        r[1] = ~r[1];
    }
}

/**Function*************************************************************

  Synopsis    [Computes truth table of the node.]
 
  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SopToTruthBig(pSop string, nInputs int32, pVars **word, pCube *word, pRes *word) {
    var nVars int32  = Abc_SopGetVarNum(pSop);
    var nWords int32;
    if nVars <= 6 {
        nWords = 1;
    } else {
        nWords = 1 << (nVars-6);
    }
    var v, i, lit int32;
    assert(nVars >= 0 && nVars <= 16, "nVars >= 0 && nVars <= 16");
    assert(nVars == nInputs, "nVars == nInputs");
    for  i = 0; i < nWords; i++  {
        pRes[i] = 0;
    }
    for {
        for  i = 0; i < nWords; i++  {
            pCube[i] = ~(word)0;
        }
        for  v = 0; v < nVars; v, lit = v+1, lit+1  {
            if pSop[lit] == '1' {
                for  i = 0; i < nWords; i++  {
                    pCube[i] &= pVars[v][i];
                }
            
            } else if pSop[lit] == '0' {
                for  i = 0; i < nWords; i++  {
                    pCube[i] &= ~pVars[v][i];
                }
            
            } else if pSop[lit] != '-' {
                assert(false, "0"); }
        }
        for  i = 0; i < nWords; i++  {
            pRes[i] |= pCube[i];
        }
        assert(pSop[lit] == ' ', "pSop[lit] == ' '");
        lit++;
        lit++;
        assert(pSop[lit] == '\n', "pSop[lit] == '\\n'");
        lit++;
    
if !( pSop[lit] ) { break }
} 
    if Abc_SopIsComplement(pSop) != 0 {
        for  i = 0; i < nWords; i++  {
            pRes[i] = ~pRes[i];
        }
    }
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

