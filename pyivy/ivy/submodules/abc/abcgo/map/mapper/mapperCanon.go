package mapper

import "github.com/glycerine/abcgo/misc/extra"

/**CFile****************************************************************

  FileName    [mapperCanon.c]

  PackageName [MVSIS 1.3: Multi-valued logic synthesis system.]

  Synopsis    [Generic technology mapping engine.]

  Author      [MVSIS Group]

  Affiliation [UC Berkeley]

  Date        [Ver. 2.0. Started - June 1, 2004.]

  Revision    [$Id: mapperCanon.c,v 1.2 2005/01/23 06:59:42 alanmi Exp $]

***********************************************************************/

//#include "mapperInt.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static unsigned Map_CanonComputePhase( unsigned uTruths[][2], int nVars, unsigned uTruth, unsigned uPhase );

// static void     Map_CanonComputePhase6( unsigned uTruths[][2], int nVars, unsigned uTruth[], unsigned uPhase, unsigned uTruthRes[] );

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/*
*Function*************************************************************

	Synopsis    [Computes the N-canonical form of the Boolean function.]

	Description [The N-canonical form is defined as the truth table with
	the minimum integer value. This function exhaustively enumerates
	through the complete set of 2^N phase assignments.]

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_CanonComputeSlow(uTruths [6][2]uint32, nVarsMax, nVarsReal int32, uTruth []uint32, puPhases []uint8, uTruthRes []uint32) int32 {
	var uTruthPerm [2]uint32
	var nMints, nPhases, m int32

	nPhases = 0
	nMints = (1 << nVarsReal)
	if nVarsMax < 6 {
		// inlining of macro MAP_MASK
		uTruthRes[0] = ((^(uint32(0))) >> (32 - (32)))
		for m = 0; m < nMints; m++ {
			uTruthPerm[0] = Map_CanonComputePhase(uTruths, nVarsMax, uTruth[0], m)
			if uTruthRes[0] > uTruthPerm[0] {
				uTruthRes[0] = uTruthPerm[0]
				nPhases = 0
				puPhases[nPhases] = uint8(m)
				nPhases++

			} else if uTruthRes[0] == uTruthPerm[0] {
				if nPhases < 4 { // the max number of phases in Map_Super_t
					puPhases[nPhases] = uint8(m)
					nPhases++
				}
			}
		}
		uTruthRes[1] = uTruthRes[0]

	} else {
		// inlining of macro MAP_MASK
		uTruthRes[0] = ((^(uint32(0))) >> (32 - (32)))
		// inlining of macro MAP_MASK
		uTruthRes[1] = ((^(uint32(0))) >> (32 - (32)))
		for m = 0; m < nMints; m++ {
			Map_CanonComputePhase6(uTruths, nVarsMax, uTruth, m, uTruthPerm[:])
			if uTruthRes[1] > uTruthPerm[1] || (uTruthRes[1] == uTruthPerm[1] && uTruthRes[0] > uTruthPerm[0]) {
				uTruthRes[0] = uTruthPerm[0]
				uTruthRes[1] = uTruthPerm[1]
				nPhases = 0
				puPhases[nPhases] = uint8(m)
				nPhases++

			} else if uTruthRes[1] == uTruthPerm[1] && uTruthRes[0] == uTruthPerm[0] {
				if nPhases < 4 { // the max number of phases in Map_Super_t
					puPhases[nPhases] = uint8(m)
					nPhases++
				}
			}
		}
	}
	assert(nPhases > 0, "nPhases > 0")
	//    printf( "%d ", nPhases );
	return nPhases
}

/*
*Function*************************************************************

	Synopsis    [Performs phase transformation for one function of less than 6 variables.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_CanonComputePhase(uTruths [6][2]uint32, nVars int32, uTruth uint32, uPhase int32) uint32 {
	var v, Shift int32
	for v, Shift = int32(0), int32(1); v < nVars; v, Shift = v+1, Shift<<1 {
		if uPhase&Shift != 0 {
			uTruth = (((uTruth & ^uTruths[v][0]) << Shift) | ((uTruth & uTruths[v][0]) >> Shift))
		}
	}
	return uTruth
}

/*
*Function*************************************************************

	Synopsis    [Performs phase transformation for one function of 6 variables.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_CanonComputePhase6(uTruths [6][2]uint32, nVars int32, uTruth []uint32, uPhase int32, uTruthRes []uint32) {
	var uTemp uint32
	var v, Shift int32

	// initialize the result
	uTruthRes[0] = uTruth[0]
	uTruthRes[1] = uTruth[1]
	if uPhase == 0 {
		return
	}
	// compute the phase
	for v, Shift = int32(0), int32(1); v < nVars; v, Shift = v+1, Shift<<1 {
		if uPhase&Shift != 0 {
			if Shift < 32 {
				uTruthRes[0] = (((uTruthRes[0] & ^uTruths[v][0]) << Shift) | ((uTruthRes[0] & uTruths[v][0]) >> Shift))
				uTruthRes[1] = (((uTruthRes[1] & ^uTruths[v][1]) << Shift) | ((uTruthRes[1] & uTruths[v][1]) >> Shift))

			} else {
				uTemp = uTruthRes[0]
				uTruthRes[0] = uTruthRes[1]
				uTruthRes[1] = uTemp
			}
		}
	}
}

/*
*Function*************************************************************

	Synopsis    [Computes the N-canonical form of the Boolean function.]

	Description [The N-canonical form is defined as the truth table with
	the minimum integer value. This function exhaustively enumerates
	through the complete set of 2^N phase assignments.]

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_CanonComputeFast(p *Map_Man_t, nVarsMax, nVarsReal int32, uTruth []uint32, puPhases []uint8, uTruthRes []uint32) int32 {
	var uTruth0, uTruth1 uint32
	var uCanon0, uCanon1, uCanonBest uint32
	var uPhaseBest uint32 = 16 // Suppress "might be used uninitialized" (asserts require < 16)
	var i, Limit int32

	if nVarsMax == 6 {
		return Map_CanonComputeSlow(p.uTruths, nVarsMax, nVarsReal, uTruth, puPhases, uTruthRes)
	}

	if nVarsReal < 5 {
		//        return Map_CanonComputeSlow( p->uTruths, nVarsMax, nVarsReal, uTruth, puPhases, uTruthRes );

		uTruth0 = uTruth[0] & 0xFFFF
		assert(p.pCounters[uTruth0] > 0, "p.pCounters[uTruth0] > 0")
		uTruthRes[0] = (uint32(p.uCanons[uTruth0]) << 16) | uint32(p.uCanons[uTruth0])
		uTruthRes[1] = uTruthRes[0]
		puPhases[0] = p.uPhases[uTruth0][0]
		return 1
	}

	assert(nVarsMax == 5, "nVarsMax == 5")
	assert(nVarsReal == 5, "nVarsReal == 5")
	uTruth0 = uTruth[0] & 0xFFFF
	uTruth1 = (uTruth[0] >> 16)
	if uTruth1 == 0 {
		uTruthRes[0] = uint32(p.uCanons[uTruth0])
		uTruthRes[1] = uTruthRes[0]
		if p.pCounters[uTruth0] > 4 {
			Limit = 4
		} else {
			Limit = int32(p.pCounters[uTruth0])
		}
		for i = 0; i < Limit; i++ {
			puPhases[i] = p.uPhases[uTruth0][i]
		}
		return Limit

	} else if uTruth0 == 0 {
		uTruthRes[0] = uint32(p.uCanons[uTruth1])
		uTruthRes[1] = uTruthRes[0]
		if p.pCounters[uTruth1] > 4 {
			Limit = 4
		} else {
			Limit = int32(p.pCounters[uTruth1])
		}
		for i = 0; i < Limit; i++ {
			puPhases[i] = p.uPhases[uTruth1][i]
			puPhases[i] |= (1 << 4)
		}
		return Limit
	}
	uCanon0 = uint32(p.uCanons[uTruth0])
	uCanon1 = uint32(p.uCanons[uTruth1])
	if uCanon0 >= uCanon1 { // using nCanon1 as the main one

		assert(p.pCounters[uTruth1] > 0, "p.pCounters[uTruth1] > 0")
		uCanonBest = 0xFFFFFFFF
		for i = 0; i < int32(p.pCounters[uTruth1]); i++ {
			uCanon0 = extra.Extra_TruthPolarize(uTruth0, int32(p.uPhases[uTruth1][i]), 4)
			if uCanonBest > uCanon0 {
				uCanonBest = uCanon0
				uPhaseBest = uint32(p.uPhases[uTruth1][i])
				assert(uPhaseBest < 16, "uPhaseBest < 16")
			}
		}
		uTruthRes[0] = (uCanon1 << 16) | uCanonBest
		uTruthRes[1] = uTruthRes[0]
		puPhases[0] = uint8(uPhaseBest)
		return 1

	} else if uCanon0 < uCanon1 {
		assert(p.pCounters[uTruth0] > 0, "p.pCounters[uTruth0] > 0")
		uCanonBest = 0xFFFFFFFF
		for i = 0; i < int32(p.pCounters[uTruth0]); i++ {
			uCanon1 = extra.Extra_TruthPolarize(uTruth1, int32(p.uPhases[uTruth0][i]), 4)
			if uCanonBest > uCanon1 {
				uCanonBest = uCanon1
				uPhaseBest = uint32(p.uPhases[uTruth0][i])
				assert(uPhaseBest < 16, "uPhaseBest < 16")
			}
		}
		uTruthRes[0] = (uCanon0 << 16) | uCanonBest
		uTruthRes[1] = uTruthRes[0]
		puPhases[0] = uint8(uPhaseBest | (1 << 4))
		return 1

	} else {
		assert(false, "0")
		return Map_CanonComputeSlow(p.uTruths, nVarsMax, nVarsReal, uTruth, puPhases, uTruthRes)
	}
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END
