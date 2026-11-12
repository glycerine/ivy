package cov

import (
	"fmt"
	"os"
)

/**CFile****************************************************************

  FileName    [covInt.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Mapping into network of SOPs/ESOPs.]

  Synopsis    [Internal declarations.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: covInt.h,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "base/abc/abc.h"

//ABC_NAMESPACE_HEADER_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

type Min_Man_t = Min_Man_t_
type Min_Cube_t = Min_Cube_t_

type Min_Man_t_ struct {
	nVars   int32 // the number of vars
	nWords  int32 // the number of words
	pMemMan any   // retained only as a compatibility field; Go slices own cube data
	// temporary cubes
	pOne0   *Min_Cube_t    // tautology cube
	pOne1   *Min_Cube_t    // tautology cube
	pTriv0  [2]*Min_Cube_t // trivial cube
	pTriv1  [2]*Min_Cube_t // trivial cube
	pTemp   *Min_Cube_t    // cube for computing the distance
	pBubble *Min_Cube_t    // cube used as a separator
	// temporary storage for the new cover
	nCubes  int32         // the number of cubes
	ppStore []*Min_Cube_t // storage for cubes by number of literals
}

type Min_Cube_t_ struct {
	pNext  *Min_Cube_t // next cube in a linked cover
	nVars  uint32      // number of variables
	nWords uint32      // number of machine words
	nLits  uint32      // number of literals in the cube
	uData  []uint32    // bit data; a slice replaces C's flexible trailing array
}

// iterators through the entries in the linked lists of cubes
// Original macro:
// #define Min_CoverForEachCube( pCover, pCube )                   \
//     for ( pCube = pCover;                                       \
//           pCube;                                                \
//           pCube = pCube->pNext )

// Original macro:
// #define Min_CoverForEachCubeSafe( pCover, pCube, pCube2 )       \
//     for ( pCube = pCover,                                       \
//           pCube2 = pCube? pCube->pNext: NULL;                   \
//           pCube;                                                \
//           pCube = pCube2,                                       \
//           pCube2 = pCube? pCube->pNext: NULL )

// Original macro:
// #define Min_CoverForEachCubePrev( pCover, pCube, ppPrev )       \
//     for ( pCube = pCover,                                       \
//           ppPrev = &(pCover);                                   \
//           pCube;                                                \
//           ppPrev = &pCube->pNext,                               \
//           pCube = pCube->pNext )

// macros to get hold of bits and values in the cubes
func Min_CubeHasBit(p *Min_Cube_t, i int32) bool { return (p.uData[(i)>>5] & (1 << ((i) & 31))) != 0 }
func Min_CubeSetBit(p *Min_Cube_t, i int32)      { p.uData[(i)>>5] |= (1 << ((i) & 31)) }
func Min_CubeXorBit(p *Min_Cube_t, i int32)      { p.uData[(i)>>5] ^= (1 << ((i) & 31)) }
func Min_CubeGetVar(p *Min_Cube_t, Var int32) int32 {
	return int32(3 & (p.uData[(2*Var)>>5] >> ((2 * Var) & 31)))
}
func Min_CubeXorVar(p *Min_Cube_t, Var int32, Value int32) {
	p.uData[(2*Var)>>5] ^= (uint32(Value) << ((2 * Var) & 31))
}

/*=== covMinEsop.c ==========================================================*/
// extern void         Min_EsopMinimize( Min_Man_t * p );

// extern void         Min_EsopAddCube( Min_Man_t * p, Min_Cube_t * pCube );

/*=== covMinSop.c ==========================================================*/
// extern void         Min_SopMinimize( Min_Man_t * p );

// extern void         Min_SopAddCube( Min_Man_t * p, Min_Cube_t * pCube );

/*=== covMinMan.c ==========================================================*/
// extern Min_Man_t *  Min_ManAlloc( int nVars );

// extern void         Min_ManClean( Min_Man_t * p, int nSupp );

// extern void         Min_ManFree( Min_Man_t * p );

/*=== covMinUtil.c ==========================================================*/
// extern void         Min_CoverCreate( Vec_Str_t * vCover, Min_Cube_t * pCover, char Type );

// extern void         Min_CubeWrite( FILE * pFile, Min_Cube_t * pCube );

// extern void         Min_CoverWrite( FILE * pFile, Min_Cube_t * pCover );

// extern void         Min_CoverWriteStore( FILE * pFile, Min_Man_t * p );

// extern void         Min_CoverWriteFile( Min_Cube_t * pCover, char * pName, int fEsop );

// extern void         Min_CoverCheck( Min_Man_t * p );

// extern int          Min_CubeCheck( Min_Cube_t * pCube );

// extern Min_Cube_t * Min_CoverCollect( Min_Man_t * p, int nSuppSize );

// extern void         Min_CoverExpand( Min_Man_t * p, Min_Cube_t * pCover );

// extern int          Min_CoverSuppVarNum( Min_Man_t * p, Min_Cube_t * pCover );

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/*
*Function*************************************************************

	Synopsis    [Creates the cube.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Min_CubeAlloc(p *Min_Man_t) *Min_Cube_t {
	pCube := &Min_Cube_t_{nVars: uint32(p.nVars), nWords: uint32(p.nWords), uData: make([]uint32, p.nWords)}
	for i := range pCube.uData {
		pCube.uData[i] = ^uint32(0)
	}
	return pCube
}

/*
*Function*************************************************************

	Synopsis    [Creates the cube representing elementary var.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Min_CubeAllocVar(p *Min_Man_t, iVar int32, fCompl int32) *Min_Cube_t {
	var pCube *Min_Cube_t
	pCube = Min_CubeAlloc(p)
	Min_CubeXorBit(pCube, iVar*2+fCompl)
	pCube.nLits = 1
	return pCube
}

/*
*Function*************************************************************

	Synopsis    [Creates the cube.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Min_CubeDup(p *Min_Man_t, pCopy *Min_Cube_t) *Min_Cube_t {
	pCube := Min_CubeAlloc(p)
	copy(pCube.uData, pCopy.uData)
	pCube.nLits = pCopy.nLits
	return pCube
}

/*
*Function*************************************************************

	Synopsis    [Recycles the cube.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Min_CubeRecycle(p *Min_Man_t, pCube *Min_Cube_t) {
	// Go's collector owns cube storage; it becomes reclaimable with its cover.
	_ = p
	_ = pCube
}

/*
*Function*************************************************************

	Synopsis    [Recycles the cube cover.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Min_CoverRecycle(p *Min_Man_t, pCover *Min_Cube_t) {
	for cube := pCover; cube != nil; {
		next := cube.pNext
		Min_CubeRecycle(p, cube)
		cube = next
	}
}

/*
*Function*************************************************************

	Synopsis    [Counts the number of cubes in the cover.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Min_CubeCountLits(pCube *Min_Cube_t) int32 {
	var uData uint32
	var Count, i, w int32
	for w = 0; w < int32(pCube.nWords); w++ {
		uData = pCube.uData[w] ^ (pCube.uData[w] >> 1)
		for i = 0; i < 32; i += 2 {
			if uData&(1<<i) != 0 {
				Count++
			}
		}
	}
	return Count
}

/*
*Function*************************************************************

	Synopsis    [Counts the number of cubes in the cover.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Min_CubeGetLits(pCube *Min_Cube_t, vLits *Vec_Int_t) {
	var uData uint32
	var i, w int32
	Vec_IntClear(vLits)
	for w = 0; w < int32(pCube.nWords); w++ {
		uData = pCube.uData[w] ^ (pCube.uData[w] >> 1)
		for i = 0; i < 32; i += 2 {
			if uData&(1<<i) != 0 {
				Vec_IntPush(vLits, w*16+i/2)
			}
		}
	}
}

/*
*Function*************************************************************

	Synopsis    [Counts the number of cubes in the cover.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Min_CoverCountCubes(pCover *Min_Cube_t) int32 {
	var pCube *Min_Cube_t
	var Count int32
	// inlining of macro Min_CoverForEachCube
	for pCube = pCover; pCube != nil; pCube = pCube.pNext {
		Count++
	}
	return Count
}

/*
*Function*************************************************************

	Synopsis    [Checks if two cubes are disjoint.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Min_CubesDisjoint(pCube0 *Min_Cube_t, pCube1 *Min_Cube_t) int32 {
	var uData uint32
	var i int32
	assert(pCube0.nVars == pCube1.nVars, "pCube0.nVars == pCube1.nVars")
	for i = 0; i < int32(pCube0.nWords); i++ {
		uData = pCube0.uData[i] & pCube1.uData[i]
		uData = (uData | (uData >> 1)) & 0x55555555
		if uData != 0x55555555 {
			return 1
		}
	}
	return 0
}

/*
*Function*************************************************************

	Synopsis    [Collects the disjoint variables of the two cubes.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Min_CoverGetDisjVars(pThis *Min_Cube_t, pCube *Min_Cube_t, vVars *Vec_Int_t) {
	var uData uint32
	var i, w int32
	Vec_IntClear(vVars)
	for w = 0; w < int32(pCube.nWords); w++ {
		uData = pThis.uData[w] & (pThis.uData[w] >> 1) & 0x55555555
		uData &= (pCube.uData[w] ^ (pCube.uData[w] >> 1))
		if uData == 0 {
			continue
		}
		for i = 0; i < 32; i += 2 {
			if uData&(1<<i) != 0 {
				Vec_IntPush(vVars, w*16+i/2)
			}
		}
	}
}

/*
*Function*************************************************************

	Synopsis    [Checks if two cubes are disjoint.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Min_CubesDistOne(pCube0 *Min_Cube_t, pCube1 *Min_Cube_t, pTemp *Min_Cube_t) int32 {
	var uData uint32
	var i, fFound int32
	for i = 0; i < int32(pCube0.nWords); i++ {
		uData = pCube0.uData[i] ^ pCube1.uData[i]
		if uData == 0 {
			if pTemp != nil {
				pTemp.uData[i] = 0
			}
			continue
		}
		if fFound != 0 {
			return 0
		}
		uData = (uData | (uData >> 1)) & 0x55555555
		if (uData & (uData - 1)) > 0 { // more than one 1
			return 0
		}
		if pTemp != nil {
			pTemp.uData[i] = uData | (uData << 1)
		}
		fFound = 1
	}
	if fFound == 0 {
		fmt.Printf("\n")
		Min_CubeWrite(os.Stdout, pCube0)
		Min_CubeWrite(os.Stdout, pCube1)
		fmt.Printf("Error: Min_CubesDistOne() looks at two equal cubes!\n")
	}
	return 1
}

/*
*Function*************************************************************

	Synopsis    [Checks if two cubes are disjoint.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Min_CubesDistTwo(pCube0 *Min_Cube_t, pCube1 *Min_Cube_t, pVar0 *int32, pVar1 *int32) int32 {
	var uData uint32 //, uData2;
	var i int32
	var k int32
	var Var0 int32 = -1
	var Var1 int32 = -1
	for i = 0; i < int32(pCube0.nWords); i++ {
		uData = pCube0.uData[i] ^ pCube1.uData[i]
		if uData == 0 {
			continue
		}
		if Var0 >= 0 && Var1 >= 0 { // more than two 1s
			return 0
		}
		uData = (uData | (uData >> 1)) & 0x55555555
		if (Var0 >= 0 || Var1 >= 0) && (uData&(uData-1)) > 0 {
			return 0
		}
		for k = 0; k < 32; k += 2 {
			if uData&(1<<k) != 0 {
				if Var0 == -1 {
					Var0 = 16*i + k/2
				} else if Var1 == -1 {
					Var1 = 16*i + k/2
				} else {
					return 0
				}
			}
		}
		/*
		   if ( Var0 >= 0 )
		   {
		       uData &= 0xFFFF;
		       uData2 = (uData >> 16);
		       if ( uData && uData2 )
		           return 0;
		       if ( uData )
		       {
		       }
		       uData }= uData2;
		       uData &= 0x
		   }
		*/
	}
	if Var0 >= 0 && Var1 >= 0 {
		*pVar0 = Var0
		*pVar1 = Var1
		return 1
	}
	if Var0 == -1 || Var1 == -1 {
		fmt.Printf("\n")
		Min_CubeWrite(os.Stdout, pCube0)
		Min_CubeWrite(os.Stdout, pCube1)
		fmt.Printf("Error: Min_CubesDistTwo() looks at two equal cubes or dist1 cubes!\n")
	}
	return 0
}

/*
*Function*************************************************************

	Synopsis    [Makes the produce of two cubes.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Min_CubesProduct(p *Min_Man_t, pCube0 *Min_Cube_t, pCube1 *Min_Cube_t) *Min_Cube_t {
	var pCube *Min_Cube_t
	var i int32
	assert(pCube0.nVars == pCube1.nVars, "pCube0.nVars == pCube1.nVars")
	pCube = Min_CubeAlloc(p)
	for i = 0; i < p.nWords; i++ {
		pCube.uData[i] = pCube0.uData[i] & pCube1.uData[i]
	}
	pCube.nLits = Min_CubeCountLits(pCube)
	return pCube
}

/*
*Function*************************************************************

	Synopsis    [Makes the produce of two cubes.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Min_CubesXor(p *Min_Man_t, pCube0 *Min_Cube_t, pCube1 *Min_Cube_t) *Min_Cube_t {
	var pCube *Min_Cube_t
	var i int32
	assert(pCube0.nVars == pCube1.nVars, "pCube0.nVars == pCube1.nVars")
	pCube = Min_CubeAlloc(p)
	for i = 0; i < p.nWords; i++ {
		pCube.uData[i] = pCube0.uData[i] ^ pCube1.uData[i]
	}
	pCube.nLits = Min_CubeCountLits(pCube)
	return pCube
}

/*
*Function*************************************************************

	Synopsis    [Makes the produce of two cubes.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Min_CubesAreEqual(pCube0 *Min_Cube_t, pCube1 *Min_Cube_t) int32 {
	var i int32
	for i = 0; i < int32(pCube0.nWords); i++ {
		if pCube0.uData[i] != pCube1.uData[i] {
			return 0
		}
	}
	return 1
}

/*
*Function*************************************************************

	Synopsis    [Returns 1 if pCube1 is contained in pCube0, bitwise.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Min_CubeIsContained(pCube0 *Min_Cube_t, pCube1 *Min_Cube_t) int32 {
	var i int32
	for i = 0; i < int32(pCube0.nWords); i++ {
		if (pCube0.uData[i] & pCube1.uData[i]) != pCube1.uData[i] {
			return 0
		}
	}
	return 1
}

/*
*Function*************************************************************

	Synopsis    [Transforms the cube into the result of merging.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Min_CubesTransform(pCube *Min_Cube_t, pDist *Min_Cube_t, pMask *Min_Cube_t) {
	var w int32
	for w = 0; w < int32(pCube.nWords); w++ {
		pCube.uData[w] = pCube.uData[w] ^ pDist.uData[w]
		pCube.uData[w] |= (pDist.uData[w] & ~pMask.uData[w])
	}
}

/*
*Function*************************************************************

	Synopsis    [Transforms the cube into the result of distance-1 merging.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Min_CubesTransformOr(pCube *Min_Cube_t, pDist *Min_Cube_t) {
	var w int32
	for w = 0; w < int32(pCube.nWords); w++ {
		pCube.uData[w] |= pDist.uData[w]
	}
}

/*
*Function*************************************************************

	Synopsis    [Sorts the cover in the increasing number of literals.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Min_CoverExpandRemoveEqual(p *Min_Man_t, pCover *Min_Cube_t) {
	var pCube, pCube2, pThis *Min_Cube_t
	if pCover == nil {
		Min_ManClean(p, p.nVars)
		return
	}
	Min_ManClean(p, pCover.nVars)
	// inlining of macro Min_CoverForEachCubeSafe
	for pCube = pCover; pCube != nil; {
		pCube2 = pCube.pNext
		// go through the linked list
		// inlining of macro Min_CoverForEachCube
		for pThis = p.ppStore[pCube.nLits]; pThis != nil; pThis = pThis.pNext {
			if Min_CubesAreEqual(pCube, pThis) != 0 {
				Min_CubeRecycle(p, pCube)
				break
			}
		}
		if pThis != nil {
			pCube = pCube2
			continue
		}
		pCube.pNext = p.ppStore[pCube.nLits]
		p.ppStore[pCube.nLits] = pCube
		p.nCubes++
		pCube = pCube2
	}
}

/*
*Function*************************************************************

	Synopsis    [Returns 1 if the given cube is contained in one of the cubes of the cover.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Min_CoverContainsCube(p *Min_Man_t, pCube *Min_Cube_t) int32 {
	var pThis *Min_Cube_t
	var i int32
	/*
	   // this cube cannot be equal to any cube
	   Min_CoverForEachCube( p->ppStore[pCube->nLits], pThis )
	   {
	       if ( Min_CubesAreEqual( pCube, pThis ) )
	       {
	           Min_CubeWrite( stdout, pCube );
	           assert( 0 );
	       }
	   }
	*/
	// try to find a containing cube
	for i = 0; i <= int32(pCube.nLits); i++ {
		// inlining of macro Min_CoverForEachCube
		for pThis = p.ppStore[i]; pThis != nil; pThis = pThis.pNext {
			// skip the bubble
			if pThis != p.pBubble && (Min_CubeIsContained(pThis, pCube) != 0) {
				return 1
			}
		}
	}
	return 0
}

//ABC_NAMESPACE_HEADER_END

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////
