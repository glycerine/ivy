package exor

/**CFile****************************************************************

  FileName    [exorCubes.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Exclusive sum-of-product minimization.]

  Synopsis    [Cube manipulation.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: exorCubes.c,v 1.0 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

////////////////////////////////////////////////////////////////////////
///                                                                  ///
///                  Implementation of EXORCISM - 4                  ///
///              An Exclusive Sum-of-Product Minimizer               ///
///                                                                  ///
///               Alan Mishchenko  <alanmi@ee.pdx.edu>               ///
///                                                                  ///
////////////////////////////////////////////////////////////////////////
///                                                                  ///
///              Cube Allocation and Free Cube Management            ///
///                                                                  ///
///  Ver. 1.0. Started - July 18, 2000. Last update - July 20, 2000  ///
///  Ver. 1.1. Started - July 24, 2000. Last update - July 29, 2000  ///
///  Ver. 1.2. Started - July 30, 2000. Last update - July 30, 2000  ///
///  Ver. 1.5. Started -  Aug 19, 2000. Last update -  Aug 19, 2000  ///
///                                                                  ///
////////////////////////////////////////////////////////////////////////
///   This software was tested with the BDD package "CUDD", v.2.3.0  ///
///                          by Fabio Somenzi                        ///
///                  http://vlsi.colorado.edu/~fabio/                ///
////////////////////////////////////////////////////////////////////////

//#include "exor.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                       EXTERNAL FUNCTIONS                         ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                       EXTERNAL VARIABLES                         ///
////////////////////////////////////////////////////////////////////////

// information about the cube cover before and after simplification
var g_CoverInfo cinfo;

////////////////////////////////////////////////////////////////////////
///                    FUNCTIONS OF THIS MODULE                      ///
////////////////////////////////////////////////////////////////////////

// cube cover memory allocation/delocation procedures
// (called from the ExorMain module)
// int AllocateCover( int nCubes, int nWordsIn, int nWordsOut );

// void DelocateCover();


// manipulation of the free cube list
// (called from Pseudo-Kronecker, ExorList, and ExorLink modules)
// void AddToFreeCubes( Cube * pC );

// Cube * GetFreeCube();


////////////////////////////////////////////////////////////////////////
///                      EXPORTED VARIABLES                          ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                      STATIC VARIABLES                            ///
////////////////////////////////////////////////////////////////////////

// the pointer to the allocated memory
var s_pCoverMemory **Cube;

// the list of free cubes
var s_CubesFree *Cube;

///////////////////////////////////////////////////////////////////
///                  CUBE COVER MEMORY MANAGEMENT                //
///////////////////////////////////////////////////////////////////

int AllocateCover( int nCubes, int nWordsIn, int nWordsOut )
// uses the cover parameters nCubes and nWords
// to allocate-and-clean the cover in one large piece
{
    var OneCubeSize int32;
    var OneInputSetSize int32;
    var pp **Cube;
    var TotalSize int32;
    var i, k int32;

    // determine the size of one cube WITH storage for bits
    OneCubeSize = sizeof(Cube) + (nWordsIn+nWordsOut)*sizeof(unsigned);
    // determine what is the amount of storage for the input part of the cube 
    OneInputSetSize = nWordsIn*sizeof(unsigned);

    // allocate memory for the array of pointers
    // inlining of macro ABC_ALLOC
    pp = (Cube **)((Cube * *) malloc(sizeof(Cube *) * (size_t)(nCubes)));
    if pp == nil {
        return 0; }

    // determine the size of the total cube cover
    TotalSize = nCubes*OneCubeSize;
    // allocate and clear memory for the cover in one large piece
    // inlining of macro ABC_ALLOC
    pp[0] = (Cube *)((char *) malloc(sizeof(char) * (size_t)(TotalSize)));
    if pp[0] == nil {
        return 0; }
    memset( pp[0], 0, (size_t)TotalSize );

    // assign pointers to cubes and bit strings inside this piece
    pp[0].pCubeDataIn  = (unsigned*)(pp[0] + 1);
    pp[0].pCubeDataOut = (unsigned*)((char*)pp[0].pCubeDataIn + OneInputSetSize);
    for  i = 1; i < nCubes; i++  {
        pp[i] = (Cube *)((char*)pp[i-1] + OneCubeSize);
        pp[i].pCubeDataIn = (unsigned*)(pp[i] + 1);
        pp[i].pCubeDataOut = (unsigned*)((char*)pp[i].pCubeDataIn + OneInputSetSize);
    }

    // connect the cubes into the list using Next pointers
    for  k = 0; k < nCubes-1; k++  {
        pp[k].Next = pp[k+1];
    }
    // the last pointer is already set to NULL

    // assign the head of the free list
    s_CubesFree = pp[0];
    // set the counters of the used and free cubes
    g_CoverInfo.nCubesInUse = 0;
    g_CoverInfo.nCubesFree = nCubes;

    // save the pointer to the allocated memory
    s_pCoverMemory = pp;

    assert(g_CoverInfo.nCubesInUse + g_CoverInfo.nCubesFree == g_CoverInfo.nCubesAlloc, "g_CoverInfo.nCubesInUse + g_CoverInfo.nCubesFree == g_CoverInfo.nCubesAlloc");

    return nCubes*sizeof(Cube *) + TotalSize;
}

func DelocateCover() {
    // inlining of macro ABC_FREE
    ((s_pCoverMemory[0]) ? (free((char *) (s_pCoverMemory[0])), (s_pCoverMemory[0]) = 0) : 0);
    // inlining of macro ABC_FREE
    ((s_pCoverMemory) ? (free((char *) (s_pCoverMemory)), (s_pCoverMemory) = 0) : 0);
}

///////////////////////////////////////////////////////////////////
///           FREE CUBE LIST MANIPULATION FUNCTIONS             ///
///////////////////////////////////////////////////////////////////

func AddToFreeCubes(p *Cube) {
    assert((p != nil), "p");
    assert(p.Prev == nil, "p.Prev == nil"); // the cube should not be in use
    assert(p.Next == nil, "p.Next == nil");
    assert(p.ID, "p.ID");

    p.Next = s_CubesFree;
    s_CubesFree = p;

    // set the ID of the cube to 0, 
    // so that cube pair garbage collection could recognize it as different
    p.ID = 0;

    g_CoverInfo.nCubesFree++;
}

func GetFreeCube() *Cube {
    var p *Cube;
    assert((s_CubesFree != nil), "s_CubesFree");
    p = s_CubesFree;
    s_CubesFree = s_CubesFree.Next;
    p.Next = nil;
    g_CoverInfo.nCubesFree--;
    return p;
}

///////////////////////////////////////////////////////////////////
////////////              End of File             /////////////////
///////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END
