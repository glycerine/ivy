package mvc

/**CFile****************************************************************

  FileName    [mvcDivisor.c]

  PackageName [MVSIS 2.0: Multi-valued logic synthesis system.]

  Synopsis    [Procedures for compute the quick divisor.]

  Author      [MVSIS Group]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - February 1, 2003.]

  Revision    [$Id: mvcDivisor.c,v 1.1 2003/04/03 15:34:08 alanmi Exp $]

***********************************************************************/

//#include "mvc.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static void  Mvc_CoverDivisorZeroKernel( Mvc_Cover_t * pCover );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Returns the quick divisor of the cover.]

  Description [Returns NULL, if there is not divisor other than
  trivial.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mvc_CoverDivisor(pCover *Mvc_Cover_t) *Mvc_Cover_t {
    var pKernel *Mvc_Cover_t;
    if Mvc_CoverReadCubeNum(pCover) <= 1 {
        return nil; }
    // allocate the literal array and count literals
    if Mvc_CoverAnyLiteral( pCover, nil ) == -1 {
        return nil; }
    // duplicate the cover
    pKernel = Mvc_CoverDup(pCover);
    // perform the kerneling
    Mvc_CoverDivisorZeroKernel( pKernel );
    assert(Mvc_CoverReadCubeNum(pKernel), "Mvc_CoverReadCubeNum(pKernel)");
    return pKernel;
}

/**Function*************************************************************

  Synopsis    [Computes a level-zero kernel.]

  Description [Modifies the cover to contain one level-zero kernel.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mvc_CoverDivisorZeroKernel(pCover *Mvc_Cover_t) {
    var iLit int32;
    // find any literal that occurs at least two times
//    iLit = Mvc_CoverAnyLiteral( pCover, NULL );
    iLit = Mvc_CoverWorstLiteral( pCover, nil );
//    iLit = Mvc_CoverBestLiteral( pCover, NULL );
    if iLit == -1 {
        return; }
    // derive the cube-free quotient
    Mvc_CoverDivideByLiteralQuo( pCover, iLit ); // the same cover
    Mvc_CoverMakeCubeFree( pCover );             // the same cover
    // call recursively
    Mvc_CoverDivisorZeroKernel( pCover );              // the same cover
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

