package cov

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [covMinUtil.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Mapping into network of SOPs/ESOPs.]

  Synopsis    [Utilities.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: covMinUtil.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "covInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Min_CubeCreate(vCover *Vec_Str_t, pCube *Min_Cube_t, Type byte) {
    var i int32;
    assert(int32(pCube.nLits) == Min_CubeCountLits(pCube), "int32(pCube.nLits) == Min_CubeCountLits(pCube)");
    for  i = 0; i < int32(pCube.nVars); i++  {
        if Min_CubeHasBit(pCube, i*2) {
            if Min_CubeHasBit(pCube, i*2+1) {
//                fprintf( pFile, "-" );
                Vec_StrPush( vCover, '-' );
            } else {
//                fprintf( pFile, "0" );
                Vec_StrPush( vCover, '0' ); }
        
        } else {
            if Min_CubeHasBit(pCube, i*2+1) {
 //               fprintf( pFile, "1" );
                Vec_StrPush( vCover, '1' );
            } else {
//                fprintf( pFile, "?" );
                Vec_StrPush( vCover, '?' ); }
        }
    }
//    fprintf( pFile, " 1\n" );
    Vec_StrPush( vCover, ' ' );
    Vec_StrPush( vCover, Type );
    Vec_StrPush( vCover, '\n' );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Min_CoverCreate(vCover *Vec_Str_t, pCover *Min_Cube_t, Type byte) {
    var pCube *Min_Cube_t;
    assert(pCover != nil, "pCover != nil");
    Vec_StrClear( vCover );
    // inlining of macro Min_CoverForEachCube
    for  pCube = pCover;                                                 (pCube != nil);                                                          pCube = pCube.pNext  {
        Min_CubeCreate( vCover, pCube, Type );
    }
    Vec_StrPush( vCover, 0 );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Min_CubeWrite(pFile *os.File, pCube *Min_Cube_t) {
    var i int32;
    assert(int32(pCube.nLits) == Min_CubeCountLits(pCube), "int32(pCube.nLits) == Min_CubeCountLits(pCube)");
    for  i = 0; i < int32(pCube.nVars); i++  {
        if Min_CubeHasBit(pCube, i*2) {
            if Min_CubeHasBit(pCube, i*2+1) {
                fmt.Fprintf(pFile, "-");
            } else {
                fmt.Fprintf(pFile, "0"); }
        
        } else {
            if Min_CubeHasBit(pCube, i*2+1) {
                fmt.Fprintf(pFile, "1");
            } else {
                fmt.Fprintf(pFile, "?"); }
        }
    }
    fmt.Fprintf(pFile, " 1\n");
//    fprintf( pFile, " %d\n", pCube->nLits );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Min_CoverWrite(pFile *os.File, pCover *Min_Cube_t) {
    var pCube *Min_Cube_t;
    // inlining of macro Min_CoverForEachCube
    for  pCube = pCover;                                                 (pCube != nil);                                                          pCube = pCube.pNext  {
        Min_CubeWrite( pFile, pCube );
    }
    fmt.Printf("\n");
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Min_CoverWriteStore(pFile *os.File, p *Min_Man_t) {
    var pCube *Min_Cube_t;
    var i int32;
    for  i = 0; i <= p.nVars; i++  {
        // inlining of macro Min_CoverForEachCube
        for  pCube = p.ppStore[i];                                                 (pCube != nil);                                                          pCube = pCube.pNext  {
            fmt.Printf("%2d : ", i);
            if pCube == p.pBubble {
                fmt.Printf("Bubble\n");
                continue;
            }
            Min_CubeWrite( pFile, pCube );
        }
    }
    fmt.Printf("\n");
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Min_CoverWriteFile(pCover *Min_Cube_t, pName string, fEsop int32) {
    var Buffer string // was [1000]byte in C

    var pCube *Min_Cube_t;
    var pFile *os.File;
    var i int32;
    Buffer = fmt.Sprintf("%s.%s", pName, fEsop? "esop" : "pla");
    for  i = strlen(Buffer) - 1; i >= 0; i--  {
        if Buffer[i] == '<' || Buffer[i] == '>' {
            Buffer[i] = '_'; }
    }
    pFile = fopen( Buffer, "w" );
    fmt.Fprintf(pFile, "# %s cover for output %s generated by ABC on %s\n", fEsop? "ESOP":"SOP", pName, Extra_TimeStamp());
    fmt.Fprintf(pFile, ".i %d\n", pCover? pCover.nVars : 0);
    fmt.Fprintf(pFile, ".o %d\n", 1);
    fmt.Fprintf(pFile, ".p %d\n", Min_CoverCountCubes(pCover));
    if fEsop != 0 { fmt.Fprintf(pFile, ".type esop\n"); }
    // inlining of macro Min_CoverForEachCube
    for  pCube = pCover;                                                 (pCube != nil);                                                          pCube = pCube.pNext  {
        Min_CubeWrite( pFile, pCube );
    }
    fmt.Fprintf(pFile, ".e\n");
    _ = (pFile).Close();
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Min_CoverCheck(p *Min_Man_t) {
    var pCube *Min_Cube_t;
    var i int32;
    for  i = 0; i <= p.nVars; i++  {
        // inlining of macro Min_CoverForEachCube
        for  pCube = p.ppStore[i];                                                 (pCube != nil);                                                          pCube = pCube.pNext  {
            assert(i == int32(pCube.nLits), "i == int32(pCube.nLits)");
        }
    }
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Min_CubeCheck(pCube *Min_Cube_t) int32 {
    var i int32;
    for  i = 0; i < int32(pCube.nVars); i++  {
        if Min_CubeGetVar( pCube, i ) == 0 {
            return 0; }
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Converts the cover from the sorted structure.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Min_CoverCollect(p *Min_Man_t, nSuppSize int32) *Min_Cube_t {
    var pCov *Min_Cube_t = nil;
    var ppTail **Min_Cube_t = &pCov;
    var pCube, pCube2 *Min_Cube_t;
    var i int32;
    for  i = 0; i <= nSuppSize; i++  {
        // inlining of macro Min_CoverForEachCubeSafe
        for  pCube = p.ppStore[i],                                                 pCube2 = pCube? pCube.pNext: nil;                             (pCube != nil);                                                          pCube = pCube2,                                                 pCube2 = pCube? pCube.pNext: nil  {
            assert(i == int32(pCube.nLits), "i == int32(pCube.nLits)");
            *ppTail = pCube; 
            ppTail = &pCube.pNext;
            assert(pCube.uData[0], "pCube.uData[0]"); // not a bubble
        }
    }
    *ppTail = nil;
    return pCov;
}

/**Function*************************************************************

  Synopsis    [Sorts the cover in the increasing number of literals.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Min_CoverExpand(p *Min_Man_t, pCover *Min_Cube_t) {
    var pCube, pCube2 *Min_Cube_t;
    Min_ManClean( p, p.nVars );
    // inlining of macro Min_CoverForEachCubeSafe
    for  pCube = pCover,                                                 pCube2 = pCube? pCube.pNext: nil;                             (pCube != nil);                                                          pCube = pCube2,                                                 pCube2 = pCube? pCube.pNext: nil  {
        pCube.pNext = p.ppStore[pCube.nLits];
        p.ppStore[pCube.nLits] = pCube;
        p.nCubes++;
    }
}

/**Function*************************************************************

  Synopsis    [Sorts the cover in the increasing number of literals.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Min_CoverSuppVarNum(p *Min_Man_t, pCover *Min_Cube_t) int32 {
    var pCube *Min_Cube_t;
    var i, Counter int32;
    if pCover == nil {
        return 0; }
    // clean the cube
    for  i = 0; i < int32(pCover.nWords); i++  {
        p.pTemp.uData[i]  = ~(uint32(0));
    }
    // add the bit data
    // inlining of macro Min_CoverForEachCube
    for  pCube = pCover;                                                 (pCube != nil);                                                          pCube = pCube.pNext  {
        for  i = 0; i < int32(pCover.nWords); i++  {
            p.pTemp.uData[i] &= pCube.uData[i];
        }
    }
    // count the vars
    Counter = 0;
    for  i = 0; i < int32(pCover.nVars); i++  {
        Counter += ( Min_CubeGetVar(p.pTemp, i) != 3 );
    }
    return Counter;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

