package fxu

import "fmt"

/**CFile****************************************************************

  FileName    [fxuReduce.c]

  PackageName [MVSIS 2.0: Multi-valued logic synthesis system.]

  Synopsis    [Procedures to reduce the number of considered cube pairs.]

  Author      [MVSIS Group]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - February 1, 2003.]

  Revision    [$Id: fxuReduce.c,v 1.0 2003/02/01 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "base/abc/abc.h"
//#include "fxuInt.h"
//#include "fxu.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static int Fxu_CountPairDiffs( char * pCover, unsigned char pDiffs[] );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Precomputes the pairs to use for creating two-cube divisors.]

  Description [This procedure takes the matrix with variables and cubes 
  allocated (p), the original covers of the nodes (i-sets) and their number 
  (ppCovers,nCovers). The maximum number of pairs to compute and the total 
  number of pairs in existence. This procedure adds to the storage of
  divisors exactly the given number of pairs (nPairsMax) while taking
  first those pairs that have the smallest number of literals in their
  cube-free form.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxu_PreprocessCubePairs(p *Fxu_Matrix, vCovers *Vec_Ptr_t, nPairsTotal int32, nPairsMax int32) int32 {
    unsigned char * pnLitsDiff;   // storage for the counters of diff literals
    var pnPairCounters *int32;         // the counters of pairs for each number of diff literals
    var pCubeFirst, pCubeLast *Fxu_Cube;
    var pCube1, pCube2 *Fxu_Cube;
    var pVar *Fxu_Var;
    var nCubes, nBitsMax, nSum int32;
    var CutOffNum int32 = -1;
    var CutOffQuant int32 = -1; // Suppress "might be used uninitialized"
    var iPair, iQuant, k, c int32;
//    abctime clk = Abc_Clock();
    var pSopCover string;
    var nFanins int32;

    assert(nPairsMax < nPairsTotal, "nPairsMax < nPairsTotal");

    // allocate storage for counter of diffs
    // inlining of macro ABC_FALLOC
    pnLitsDiff = ((unsigned char *) memset(malloc(sizeof(unsigned char) * (size_t)(nPairsTotal)), 0xff, sizeof(unsigned char) * (size_t)(nPairsTotal)));
    // go through the covers and precompute the distances between the pairs
    iPair    =  0;
    nBitsMax = -1;
    for  c = 0; c < vCovers.nSize; c++  {
        if pSopCover = (char *)vCovers.pArray[c] {
            nFanins = Abc_SopGetVarNum(pSopCover);
            // precompute the differences
            Fxu_CountPairDiffs( pSopCover, pnLitsDiff + iPair );
            // update the counter
            nCubes = Abc_SopGetCubeNum( pSopCover );
            iPair += nCubes * (nCubes - 1) / 2;
            if nBitsMax < nFanins {
                nBitsMax = nFanins; }
        }
    }
    assert(iPair == nPairsTotal, "iPair == nPairsTotal");

    // allocate storage for counters of cube pairs by difference
    // inlining of macro ABC_FALLOC
    pnPairCounters = ((int *) memset(malloc(sizeof(int) * (size_t)(2 * nBitsMax)), 0xff, sizeof(int) * (size_t)(2 * nBitsMax)));
    memset( pnPairCounters, 0, sizeof(int) * 2 * nBitsMax );
    // count the number of different pairs
    for  k = 0; k < nPairsTotal; k++  {
        pnPairCounters[ pnLitsDiff[k] ]++;
    }
    // determine what pairs to take starting from the lower
    // so that there would be exactly pPairsMax pairs
    if pnPairCounters[0] != 0 {
        pnLitsDiff = nil;
        pnPairCounters = nil;
        fmt.Printf("The SOPs of the nodes contain duplicated cubes. Run \"bdd; sop\" before \"fx\".\n");
        return 0;
    }
    if pnPairCounters[1] != 0 {
        pnLitsDiff = nil;
        pnPairCounters = nil;
        fmt.Printf("The SOPs of the nodes are not SCC-free. Run \"bdd; sop\" before \"fx\".\n");
        return 0;
    }
    assert(pnPairCounters[0] == 0, "pnPairCounters[0] == 0"); // otherwise, covers are not dup-free
    assert(pnPairCounters[1] == 0, "pnPairCounters[1] == 0"); // otherwise, covers are not SCC-free
    nSum = 0;
    for  k = 0; k < 2 * nBitsMax; k++  {
        nSum += pnPairCounters[k];
        if nSum >= nPairsMax {
            CutOffNum   = k;
            CutOffQuant = pnPairCounters[k] - (nSum - nPairsMax);
            break;
        }
    }
    pnPairCounters = nil;

    // set to 0 all the pairs above the cut-off number and quantity
    iQuant = 0;
    iPair  = 0;
    for  k = 0; k < nPairsTotal; k++  {
        if pnLitsDiff[k] > CutOffNum { 
            pnLitsDiff[k] = 0;
        } else if pnLitsDiff[k] == CutOffNum {
            if iQuant++ >= CutOffQuant {
                pnLitsDiff[k] = 0;
            } else {
                iPair++; }
        
        } else {
            iPair++; }
    }
    assert(iPair == nPairsMax, "iPair == nPairsMax");

    // collect the corresponding pairs and add the divisors
    iPair = 0;
    for  c = 0; c < vCovers.nSize; c++  {
        if pSopCover = (char *)vCovers.pArray[c] {
            // get the var
            pVar = p.ppVars[2*c+1];
            // get the first cube
            pCubeFirst = pVar.pFirst;
            // get the last cube
            pCubeLast = pCubeFirst;
            for  k = 0; k < pVar.nCubes; k++  {
                pCubeLast = pCubeLast.pNext;
            }
            assert(pCubeLast == nil || pCubeLast.pVar != pVar, "pCubeLast == nil || pCubeLast.pVar != pVar");

            // go through the cube pairs
            for  pCube1 = pCubeFirst; pCube1 != pCubeLast; pCube1 = pCube1.pNext  {
                for  pCube2 = pCube1.pNext; pCube2 != pCubeLast; pCube2 = pCube2.pNext  {
                    if pnLitsDiff[iPair++] {   // create the divisors for this pair
                        Fxu_MatrixAddDivisor( p, pCube1, pCube2 );
                    }
                }
            }
        }
    }
    assert(iPair == nPairsTotal, "iPair == nPairsTotal");
    pnLitsDiff = nil;
//ABC_PRT( "Preprocess", Abc_Clock() - clk );
    return 1;
}


/**Function*************************************************************

  Synopsis    [Counts the differences in each cube pair in the cover.]

  Description [Takes the cover (pCover) and the array where the
  diff counters go (pDiffs). The array pDiffs should have as many
  entries as there are different pairs of cubes in the cover: n(n-1)/2.
  Fills out the array pDiffs with the following info: For each cube
  pair, included in the array is the number of literals in both cubes
  after they are made cube free.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
int Fxu_CountPairDiffs( char * pCover, unsigned char pDiffs[] )
{
    var pCube1, pCube2 string;
    var nOnes, nCubePairs, nFanins, v int32;
    nCubePairs = 0;
    nFanins = Abc_SopGetVarNum(pCover);
    // inlining of macro Abc_SopForEachCube
    for  pCube1 = (pCover); *pCube1; pCube1 += (nFanins) + 3  {
    // inlining of macro Abc_SopForEachCube
    for  pCube2 = (pCube1); *pCube2; pCube2 += (nFanins) + 3  {
        if pCube1 == pCube2 {
            continue; }
        nOnes = 0;
        for  v = 0; v < nFanins; v++  {
            nOnes += (pCube1[v] != pCube2[v]);
        }
        pDiffs[nCubePairs++] = nOnes;
    }
    }
    return 1;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

