package kit

/**CFile****************************************************************

  FileName    [kitSop.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Computation kit.]

  Synopsis    [Procedures involving SOPs.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - Dec 6, 2006.]

  Revision    [$Id: kitSop.c,v 1.00 2006/12/06 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "kit.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Creates SOP from the cube array.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Kit_SopCreate(cResult *Kit_Sop_t, vInput *Vec_Int_t, nVars int32, vMemory *Vec_Int_t) {
    var uCube uint32;
    var i int32;
    // start the cover
    cResult.nCubes = 0;
    cResult.pCubes = Vec_IntFetch( vMemory, Vec_IntSize(vInput) );
    // add the cubes
    Vec_IntForEachEntry( vInput, uCube, i )
        Kit_SopPushCube( cResult, uCube );
}

/**Function*************************************************************

  Synopsis    [Creates SOP from the cube array.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Kit_SopCreateInverse(cResult *Kit_Sop_t, vInput *Vec_Int_t, nLits int32, vMemory *Vec_Int_t) {
    var uCube, uMask uint32;
    var i int32;
    var nCubes int32 = Vec_IntSize(vInput);
    // start the cover
    cResult.nCubes = 0;
    cResult.pCubes = Vec_IntFetch( vMemory, nCubes );
    // add the cubes
//    Vec_IntForEachEntry( vInput, uCube, i )
    for  i = 0; i < nCubes; i++  {
        uCube = Vec_IntEntry( vInput, i );
        uMask = ((uCube | (uCube >> 1)) & 0x55555555);
        uMask |= (uMask << 1);
        Kit_SopPushCube( cResult, uCube ^ uMask );
    }
}

/**Function*************************************************************

  Synopsis    [Duplicates SOP.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Kit_SopDup(cResult *Kit_Sop_t, cSop *Kit_Sop_t, vMemory *Vec_Int_t) {
    var uCube uint32;
    var i int32;
    // start the cover
    cResult.nCubes = 0;
    cResult.pCubes = Vec_IntFetch( vMemory, Kit_SopCubeNum(cSop) );
    // add the cubes
    // inlining of macro Kit_SopForEachCube
    for  i = 0; (i < Kit_SopCubeNum(cSop)) && ((uCube) = Kit_SopCube(cSop, i)); i++  {
        Kit_SopPushCube( cResult, uCube );
    }
}

/**Function*************************************************************

  Synopsis    [Derives the quotient of division by literal.]

  Description [Reduces the cover to be equal to the result of
  division of the given cover by the literal.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Kit_SopDivideByLiteralQuo(cSop *Kit_Sop_t, iLit int32) {
    var uCube uint32;
    var i, k int32;
    // inlining of macro Kit_SopForEachCube
    for  i = 0; (i < Kit_SopCubeNum(cSop)) && ((uCube) = Kit_SopCube(cSop, i)); i++  {
        if Kit_CubeHasLit(uCube, iLit) {
            Kit_SopWriteCube( cSop, Kit_CubeRemLit(uCube, iLit), k++ ); }
    }
    Kit_SopShrink( cSop, k );
}


/**Function*************************************************************

  Synopsis    [Divides cover by one cube.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Kit_SopDivideByCube(cSop *Kit_Sop_t, cDiv *Kit_Sop_t, vQuo *Kit_Sop_t, vRem *Kit_Sop_t, vMemory *Vec_Int_t) {
    var uCube, uDiv uint32;
    var i int32;
    // get the only cube
    assert(Kit_SopCubeNum(cDiv) == 1, "Kit_SopCubeNum(cDiv) == 1");
    uDiv = Kit_SopCube(cDiv, 0);
    // allocate covers
    vQuo.nCubes = 0;
    vQuo.pCubes = Vec_IntFetch( vMemory, Kit_SopCubeNum(cSop) );
    vRem.nCubes = 0;
    vRem.pCubes = Vec_IntFetch( vMemory, Kit_SopCubeNum(cSop) );
    // sort the cubes
    // inlining of macro Kit_SopForEachCube
    for  i = 0; (i < Kit_SopCubeNum(cSop)) && ((uCube) = Kit_SopCube(cSop, i)); i++  {
        if Kit_CubeContains( uCube, uDiv ) {
            Kit_SopPushCube( vQuo, Kit_CubeSharp(uCube, uDiv) );
        } else {
            Kit_SopPushCube( vRem, uCube ); }
    }
}

/**Function*************************************************************

  Synopsis    [Divides cover by one cube.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Kit_SopDivideInternal(cSop *Kit_Sop_t, cDiv *Kit_Sop_t, vQuo *Kit_Sop_t, vRem *Kit_Sop_t, vMemory *Vec_Int_t) {
    var uCube, uDiv uint32;
    var uCube2 uint32; // Suppress "might be used uninitialized"
    var uDiv2, uQuo uint32;
    var i, i2, k, k2, nCubesRem int32;
    assert(Kit_SopCubeNum(cSop) >= Kit_SopCubeNum(cDiv), "Kit_SopCubeNum(cSop) >= Kit_SopCubeNum(cDiv)");
    // consider special case
    if Kit_SopCubeNum(cDiv) == 1 {
        Kit_SopDivideByCube( cSop, cDiv, vQuo, vRem, vMemory );
        return;
    }
    // allocate quotient
    vQuo.nCubes = 0;
    vQuo.pCubes = Vec_IntFetch( vMemory, Kit_SopCubeNum(cSop) / Kit_SopCubeNum(cDiv) );
    // for each cube of the cover
    // it either belongs to the quotient or to the remainder
    // inlining of macro Kit_SopForEachCube
    for  i = 0; (i < Kit_SopCubeNum(cSop)) && ((uCube) = Kit_SopCube(cSop, i)); i++  {
        // skip taken cubes
        if Kit_CubeIsMarked(uCube) {
            continue; }
        // find a matching cube in the divisor
        uDiv = ~0;
        // inlining of macro Kit_SopForEachCube
        for  k = 0; (k < Kit_SopCubeNum(cDiv)) && ((uDiv) = Kit_SopCube(cDiv, k)); k++  {
            if Kit_CubeContains( uCube, uDiv ) {
                break; }
        }
        // the cube is not found 
        if k == Kit_SopCubeNum(cDiv) {
            continue; }
        // the quotient cube exists
        uQuo = Kit_CubeSharp( uCube, uDiv );
        // find corresponding cubes for other cubes of the divisor
        uDiv2 = ~0;
        // inlining of macro Kit_SopForEachCube
        for  k2 = 0; (k2 < Kit_SopCubeNum(cDiv)) && ((uDiv2) = Kit_SopCube(cDiv, k2)); k2++  {
            if k2 == k {
                continue; }
            // find a matching cube
            // inlining of macro Kit_SopForEachCube
            for  i2 = 0; (i2 < Kit_SopCubeNum(cSop)) && ((uCube2) = Kit_SopCube(cSop, i2)); i2++  {
                // skip taken cubes
                if Kit_CubeIsMarked(uCube2) {
                    continue; }
                // check if the cube can be used
                if Kit_CubeContains( uCube2, uDiv2 ) && uQuo == Kit_CubeSharp( uCube2, uDiv2 ) {
                    break; }
            }
            // the case when the cube is not found
            if i2 == Kit_SopCubeNum(cSop) {
                break; }
        }
        // we did not find some cubes - continue looking at other cubes
        if k2 != Kit_SopCubeNum(cDiv) {
            continue; }
        // we found all cubes - add the quotient cube
        Kit_SopPushCube( vQuo, uQuo );

        // mark the first cube
        Kit_SopWriteCube( cSop, Kit_CubeMark(uCube), i );
        // mark other cubes that have this quotient
        // inlining of macro Kit_SopForEachCube
        for  k2 = 0; (k2 < Kit_SopCubeNum(cDiv)) && ((uDiv2) = Kit_SopCube(cDiv, k2)); k2++  {
            if k2 == k {
                continue; }
            // find a matching cube
            // inlining of macro Kit_SopForEachCube
            for  i2 = 0; (i2 < Kit_SopCubeNum(cSop)) && ((uCube2) = Kit_SopCube(cSop, i2)); i2++  {
                // skip taken cubes
                if Kit_CubeIsMarked(uCube2) {
                    continue; }
                // check if the cube can be used
                if Kit_CubeContains( uCube2, uDiv2 ) && uQuo == Kit_CubeSharp( uCube2, uDiv2 ) {
                    break; }
            }
            assert(i2 < Kit_SopCubeNum(cSop), "i2 < Kit_SopCubeNum(cSop)");
            // the cube is found, mark it 
            // (later we will add all unmarked cubes to the remainder)
            Kit_SopWriteCube( cSop, Kit_CubeMark(uCube2), i2 );
        }
    }
    // determine the number of cubes in the remainder
    nCubesRem = Kit_SopCubeNum(cSop) - Kit_SopCubeNum(vQuo) * Kit_SopCubeNum(cDiv);
    // allocate remainder
    vRem.nCubes = 0;
    vRem.pCubes = Vec_IntFetch( vMemory, nCubesRem );
    // finally add the remaining unmarked cubes to the remainder 
    // and clean the marked cubes in the cover
    // inlining of macro Kit_SopForEachCube
    for  i = 0; (i < Kit_SopCubeNum(cSop)) && ((uCube) = Kit_SopCube(cSop, i)); i++  {
        if !Kit_CubeIsMarked(uCube) {
            Kit_SopPushCube( vRem, uCube );
            continue;
        }
        Kit_SopWriteCube( cSop, Kit_CubeUnmark(uCube), i );
    }
    assert(nCubesRem == Kit_SopCubeNum(vRem), "nCubesRem == Kit_SopCubeNum(vRem)");
}

/**Function*************************************************************

  Synopsis    [Returns the common cube.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Kit_SopCommonCube(cSop *Kit_Sop_t) uint32 {
    var uMask, uCube uint32;
    var i int32;
    uMask = ~uint32(0);
    // inlining of macro Kit_SopForEachCube
    for  i = 0; (i < Kit_SopCubeNum(cSop)) && ((uCube) = Kit_SopCube(cSop, i)); i++  {
        uMask &= uCube;
    }
    return uMask;
}

/**Function*************************************************************

  Synopsis    [Makes the cover cube-free.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Kit_SopMakeCubeFree(cSop *Kit_Sop_t) {
    var uMask, uCube uint32;
    var i int32;
    uMask = Kit_SopCommonCube( cSop );
    if uMask == 0 {
        return; }
    // remove the common cube
    // inlining of macro Kit_SopForEachCube
    for  i = 0; (i < Kit_SopCubeNum(cSop)) && ((uCube) = Kit_SopCube(cSop, i)); i++  {
        Kit_SopWriteCube( cSop, Kit_CubeSharp(uCube, uMask), i );
    }
}

/**Function*************************************************************

  Synopsis    [Checks if the cover is cube-free.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Kit_SopIsCubeFree(cSop *Kit_Sop_t) int32 {
    return Kit_SopCommonCube( cSop ) == 0;
}

/**Function*************************************************************

  Synopsis    [Creates SOP composes of the common cube of the given SOP.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Kit_SopCommonCubeCover(cResult *Kit_Sop_t, cSop *Kit_Sop_t, vMemory *Vec_Int_t) {
    assert(Kit_SopCubeNum(cSop) > 0, "Kit_SopCubeNum(cSop) > 0");
    cResult.nCubes = 0;
    cResult.pCubes = Vec_IntFetch( vMemory, 1 );
    Kit_SopPushCube( cResult, Kit_SopCommonCube(cSop) );
}


/**Function*************************************************************

  Synopsis    [Find any literal that occurs more than once.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Kit_SopAnyLiteral(cSop *Kit_Sop_t, nLits int32) int32 {
    var uCube uint32;
    var i, k, nLitsCur int32;
    // go through each literal
    for  i = 0; i < nLits; i++  {
        // go through all the cubes
        nLitsCur = 0;
        // inlining of macro Kit_SopForEachCube
        for  k = 0; (k < Kit_SopCubeNum(cSop)) && ((uCube) = Kit_SopCube(cSop, k)); k++  {
            if Kit_CubeHasLit(uCube, i) {
                nLitsCur++; }
        }
        if nLitsCur > 1 {
            return i; }
    }
    return -1;
}

/**Function*************************************************************

  Synopsis    [Find the least often occurring literal.]

  Description [Find the least often occurring literal among those
  that occur more than once.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Kit_SopWorstLiteral(cSop *Kit_Sop_t, nLits int32) int32 {
    var uCube uint32;
    var i, k, iMin, nLitsMin, nLitsCur int32;
    var fUseFirst int32  = 1;

    // go through each literal
    iMin = -1;
    nLitsMin = 1000000;
    for  i = 0; i < nLits; i++  {
        // go through all the cubes
        nLitsCur = 0;
        // inlining of macro Kit_SopForEachCube
        for  k = 0; (k < Kit_SopCubeNum(cSop)) && ((uCube) = Kit_SopCube(cSop, k)); k++  {
            if Kit_CubeHasLit(uCube, i) {
                nLitsCur++; }
        }
        // skip the literal that does not occur or occurs once
        if nLitsCur < 2 {
            continue; }
        // check if this is the best literal
        if fUseFirst != 0 {
            if nLitsMin > nLitsCur {
                nLitsMin = nLitsCur;
                iMin = i;
            }
        
        } else {
            if nLitsMin >= nLitsCur {
                nLitsMin = nLitsCur;
                iMin = i;
            }
        }
    }
    if nLitsMin < 1000000 {
        return iMin; }
    return -1;
}

/**Function*************************************************************

  Synopsis    [Find the least often occurring literal.]

  Description [Find the least often occurring literal among those
  that occur more than once.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Kit_SopBestLiteral(cSop *Kit_Sop_t, nLits int32, uMask uint32) int32 {
    var uCube uint32;
    var i, k, iMax, nLitsMax, nLitsCur int32;
    var fUseFirst int32  = 1;

    // go through each literal
    iMax = -1;
    nLitsMax = -1;
    for  i = 0; i < nLits; i++  {
        if !Kit_CubeHasLit(uMask, i) {
            continue; }
        // go through all the cubes
        nLitsCur = 0;
        // inlining of macro Kit_SopForEachCube
        for  k = 0; (k < Kit_SopCubeNum(cSop)) && ((uCube) = Kit_SopCube(cSop, k)); k++  {
            if Kit_CubeHasLit(uCube, i) {
                nLitsCur++; }
        }
        // skip the literal that does not occur or occurs once
        if nLitsCur < 2 {
            continue; }
        // check if this is the best literal
        if fUseFirst != 0 {
            if nLitsMax < nLitsCur {
                nLitsMax = nLitsCur;
                iMax = i;
            }
        
        } else {
            if nLitsMax <= nLitsCur {
                nLitsMax = nLitsCur;
                iMax = i;
            }
        }
    }
    if nLitsMax >= 0 {
        return iMax; }
    return -1;
}

/**Function*************************************************************

  Synopsis    [Computes a level-zero kernel.]

  Description [Modifies the cover to contain one level-zero kernel.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Kit_SopDivisorZeroKernel_rec(cSop *Kit_Sop_t, nLits int32) {
    var iLit int32;
    // find any literal that occurs at least two times
    iLit = Kit_SopWorstLiteral( cSop, nLits );
    if iLit == -1 {
        return; }
    // derive the cube-free quotient
    Kit_SopDivideByLiteralQuo( cSop, iLit ); // the same cover
    Kit_SopMakeCubeFree( cSop );             // the same cover
    // call recursively
    Kit_SopDivisorZeroKernel_rec( cSop, nLits );    // the same cover
}

/**Function*************************************************************

  Synopsis    [Computes the quick divisor of the cover.]

  Description [Returns 0, if there is no divisor other than trivial.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Kit_SopDivisor(cResult *Kit_Sop_t, cSop *Kit_Sop_t, nLits int32, vMemory *Vec_Int_t) int32 {
    if Kit_SopCubeNum(cSop) <= 1 {
        return 0; }
    if Kit_SopAnyLiteral( cSop, nLits ) == -1 {
        return 0; }
    // duplicate the cover
    Kit_SopDup( cResult, cSop, vMemory );
    // perform the kerneling
    Kit_SopDivisorZeroKernel_rec( cResult, nLits );
    assert(Kit_SopCubeNum(cResult) > 0, "Kit_SopCubeNum(cResult) > 0");
    return 1;
}


/**Function*************************************************************

  Synopsis    [Create the one-literal cover with the best literal from cSop.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Kit_SopBestLiteralCover(cResult *Kit_Sop_t, cSop *Kit_Sop_t, uCube uint32, nLits int32, vMemory *Vec_Int_t) {
    var iLitBest int32;
    // get the best literal
    iLitBest = Kit_SopBestLiteral( cSop, nLits, uCube );
    // start the cover
    cResult.nCubes = 0;
    cResult.pCubes = Vec_IntFetch( vMemory, 1 );
    // set the cube
    Kit_SopPushCube( cResult, Kit_CubeSetLit(0, iLitBest) );
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

