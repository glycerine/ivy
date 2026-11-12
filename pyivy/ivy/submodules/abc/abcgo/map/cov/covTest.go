package cov

import "fmt"

/**CFile****************************************************************

  FileName    [covTest.c]

  SystemName  [ABC: Logic synthesis and verification system.]
 
  PackageName [Mapping into network of SOPs/ESOPs.]

  Synopsis    [Testing procedures.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: covTest.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "cov.h"

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
func Abc_NodeDeriveCoverPro(p *Min_Man_t, pCover0 *Min_Cube_t, pCover1 *Min_Cube_t) *Min_Cube_t {
    var pCover *Min_Cube_t;
    var pCube0, pCube1, pCube *Min_Cube_t;
    if pCover0 == nil || pCover1 == nil {
        return nil; }
    // clean storage
    Min_ManClean( p, p.nVars );
    // go through the cube pairs
    // inlining of macro Min_CoverForEachCube
    for  pCube0 = pCover0;                                                 pCube0;                                                          pCube0 = pCube0.pNext  {
    // inlining of macro Min_CoverForEachCube
    for  pCube1 = pCover1;                                                 pCube1;                                                          pCube1 = pCube1.pNext  {
        if Min_CubesDisjoint( pCube0, pCube1 ) {
            continue; }
        pCube = Min_CubesProduct( p, pCube0, pCube1 );
        // add the cube to storage
        Min_SopAddCube( p, pCube );
    }
    }
    Min_SopMinimize( p );
    pCover = Min_CoverCollect( p, p.nVars );
    assert(p.nCubes == Min_CoverCountCubes(pCover), "p.nCubes == Min_CoverCountCubes(pCover)");
    return pCover;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NodeDeriveCoverSum(p *Min_Man_t, pCover0 *Min_Cube_t, pCover1 *Min_Cube_t) *Min_Cube_t {
    var pCover *Min_Cube_t;
    var pThis, pCube *Min_Cube_t;
    if pCover0 == nil || pCover1 == nil {
        return nil; }
    // clean storage
    Min_ManClean( p, p.nVars );
    // add the cubes to storage
    // inlining of macro Min_CoverForEachCube
    for  pThis = pCover0;                                                 pThis;                                                          pThis = pThis.pNext  {
        pCube = Min_CubeDup( p, pThis );
        Min_SopAddCube( p, pCube );
    }
    // inlining of macro Min_CoverForEachCube
    for  pThis = pCover1;                                                 pThis;                                                          pThis = pThis.pNext  {
        pCube = Min_CubeDup( p, pThis );
        Min_SopAddCube( p, pCube );
    }
    Min_SopMinimize( p );
    pCover = Min_CoverCollect( p, p.nVars );
    assert(p.nCubes == Min_CoverCountCubes(pCover), "p.nCubes == Min_CoverCountCubes(pCover)");
    return pCover;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NodeDeriveSops(p *Min_Man_t, pRoot *Abc_Obj_t, vSupp *Vec_Ptr_t, vNodes *Vec_Ptr_t) int32 {
    var pCov0, pCov1 [2]*Min_Cube_t;
    var pCoverP, pCoverN *Min_Cube_t;
    var pObj *Abc_Obj_t;
    var i, nCubes, fCompl0, fCompl1 int32;

    // set elementary vars
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vSupp)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vSupp, i)), 1); i++  {
        pObj.pCopy = (Abc_Obj_t *)Min_CubeAllocVar( p, i, 0 );
        pObj.pNext = (Abc_Obj_t *)Min_CubeAllocVar( p, i, 1 );
    }

    // get the cover for each node in the array
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        // get the complements
        fCompl0 = Abc_ObjFaninC0(pObj);
        fCompl1 = Abc_ObjFaninC1(pObj);
        // get the covers
        pCov0[0] = (Min_Cube_t *)Abc_ObjFanin0(pObj).pCopy;
        pCov0[1] = (Min_Cube_t *)Abc_ObjFanin0(pObj).pNext;
        pCov1[0] = (Min_Cube_t *)Abc_ObjFanin1(pObj).pCopy;
        pCov1[1] = (Min_Cube_t *)Abc_ObjFanin1(pObj).pNext;
        // compute the covers
        pCoverP = Abc_NodeDeriveCoverPro( p, pCov0[ fCompl0], pCov1[ fCompl1] );
        pCoverN = Abc_NodeDeriveCoverSum( p, pCov0[!fCompl0], pCov1[!fCompl1] );
        // set the covers
        pObj.pCopy = (Abc_Obj_t *)pCoverP;
        pObj.pNext = (Abc_Obj_t *)pCoverN;
    }

     nCubes = ABC_MIN( Min_CoverCountCubes(pCoverN), Min_CoverCountCubes(pCoverP) );

/*
printf( "\n\n" );
Min_CoverWrite( stdout, pCoverP );
printf( "\n\n" );
Min_CoverWrite( stdout, pCoverN );
*/

//    printf( "\n" );
//    Min_CoverWrite( stdout, pCoverP );

//    Min_CoverExpand( p, pCoverP );
//    Min_SopMinimize( p );
//    pCoverP = Min_CoverCollect( p, p->nVars );

//    printf( "\n" );
//    Min_CoverWrite( stdout, pCoverP );

//    nCubes = Min_CoverCountCubes(pCoverP);

    // clean the copy fields
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        pObj.pCopy = pObj.pNext = nil;
    }
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vSupp)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vSupp, i)), 1); i++  {
        pObj.pCopy = pObj.pNext = nil;
    }

//    Min_CoverWriteFile( pCoverP, Abc_ObjName(pRoot), 0 );
//    printf( "\n" );
//    Min_CoverWrite( stdout, pCoverP );

//    printf( "\n" );
//    Min_CoverWrite( stdout, pCoverP );
//    printf( "\n" );
//    Min_CoverWrite( stdout, pCoverN );
    return nCubes;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkTestSop(pNtk *Abc_Ntk_t) {
    var p *Min_Man_t;
    var vSupp, vNodes *Vec_Ptr_t;
    var pObj *Abc_Obj_t;
    var i, nCubes int32;
    assert(Abc_NtkIsStrash(pNtk), "Abc_NtkIsStrash(pNtk)");

    Abc_NtkCleanCopy(pNtk);
    Abc_NtkCleanNext(pNtk);
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(pNtk)) && (((pObj) = Abc_NtkCo(pNtk, i)), 1); i++  {
        if !Abc_ObjIsNode(Abc_ObjFanin0(pObj)) {
            fmt.Printf("%-20s :  Trivial.\n", Abc_ObjName(pObj));
            continue;
        }

        vSupp  = Abc_NtkNodeSupport( pNtk, &pObj, 1 );
        vNodes = Abc_NtkDfsNodes( pNtk, &pObj, 1 );

        fmt.Printf("%20s :  Cone = %5d.  Supp = %5d.  ", Abc_ObjName(pObj), vNodes.nSize, vSupp.nSize);
//        if ( vSupp->nSize <= 128 )
        {
            p = Min_ManAlloc( vSupp.nSize );
            nCubes = Abc_NodeDeriveSops( p, pObj, vSupp, vNodes );
            fmt.Printf("Cubes = %5d.  ", nCubes);
            Min_ManFree( p );
        }
        fmt.Printf("\n");


        Vec_PtrFree( vNodes );
        Vec_PtrFree( vSupp );
    }
}



/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NodeDeriveCover(p *Min_Man_t, pCov0 *Min_Cube_t, pCov1 *Min_Cube_t, fComp0 int32, fComp1 int32) *Min_Cube_t {
    var pCover0, pCover1, pCover *Min_Cube_t;
    var pCube0, pCube1, pCube *Min_Cube_t;

    // complement the first if needed
    if !(fComp0 != 0) {
        pCover0 = pCov0;
    } else if (pCov0 != nil) && pCov0.nLits == 0 { // topmost one is the tautology cube
        pCover0 = pCov0.pNext;
    } else {
        pCover0 = p.pOne0;
p.pOne0.pNext = pCov0; }

    // complement the second if needed
    if !(fComp1 != 0) {
        pCover1 = pCov1;
    } else if (pCov1 != nil) && pCov1.nLits == 0 { // topmost one is the tautology cube
        pCover1 = pCov1.pNext;
    } else {
        pCover1 = p.pOne1;
p.pOne1.pNext = pCov1; }

    if pCover0 == nil || pCover1 == nil {
        return nil; }

    // clean storage
    Min_ManClean( p, p.nVars );
    // go through the cube pairs
    // inlining of macro Min_CoverForEachCube
    for  pCube0 = pCover0;                                                 pCube0;                                                          pCube0 = pCube0.pNext  {
    // inlining of macro Min_CoverForEachCube
    for  pCube1 = pCover1;                                                 pCube1;                                                          pCube1 = pCube1.pNext  {
        if Min_CubesDisjoint( pCube0, pCube1 ) {
            continue; }
        pCube = Min_CubesProduct( p, pCube0, pCube1 );
        // add the cube to storage
        Min_EsopAddCube( p, pCube );
    }
    }
 
    if p.nCubes > 10 {
//        printf( "(%d,", p->nCubes );
        Min_EsopMinimize( p );
//        printf( "%d) ", p->nCubes );
    }

    pCover = Min_CoverCollect( p, p.nVars );
    assert(p.nCubes == Min_CoverCountCubes(pCover), "p.nCubes == Min_CoverCountCubes(pCover)");

//    if ( p->nCubes > 1000 )
//        printf( "%d ", p->nCubes );
    return pCover;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NodeDeriveEsops(p *Min_Man_t, pRoot *Abc_Obj_t, vSupp *Vec_Ptr_t, vNodes *Vec_Ptr_t) int32 {
    var pCover, pCube *Min_Cube_t;
    var pObj *Abc_Obj_t;
    var i int32;

    // set elementary vars
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vSupp)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vSupp, i)), 1); i++  {
        pObj.pCopy = (Abc_Obj_t *)Min_CubeAllocVar( p, i, 0 );
    }

    // get the cover for each node in the array
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        pCover = Abc_NodeDeriveCover( p,  
            (Min_Cube_t *)Abc_ObjFanin0(pObj).pCopy,  
            (Min_Cube_t *)Abc_ObjFanin1(pObj).pCopy,
            Abc_ObjFaninC0(pObj), Abc_ObjFaninC1(pObj) );
        pObj.pCopy = (Abc_Obj_t *)pCover;
        if p.nCubes > 3000 {
            return -1; }
    }

    // add complement if needed
    if Abc_ObjFaninC0(pRoot) {
        if (pCover != nil) && pCover.nLits == 0 { // topmost one is the tautology cube
        
            pCube = pCover;
            pCover = pCover.pNext;
            Min_CubeRecycle( p, pCube );
            p.nCubes--;
        
        } else {
            pCube = Min_CubeAlloc( p );
            pCube.pNext = pCover;
            p.nCubes++;
        }
    }
/*
    Min_CoverExpand( p, pCover );
    Min_EsopMinimize( p );
    pCover = Min_CoverCollect( p, p->nVars );
*/
    // clean the copy fields
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        pObj.pCopy = nil;
    }
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vSupp)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vSupp, i)), 1); i++  {
        pObj.pCopy = nil;
    }

//    Min_CoverWriteFile( pCover, Abc_ObjName(pRoot), 1 );
//    Min_CoverWrite( stdout, pCover );
    return p.nCubes;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkTestEsop(pNtk *Abc_Ntk_t) {
    var p *Min_Man_t;
    var vSupp, vNodes *Vec_Ptr_t;
    var pObj *Abc_Obj_t;
    var i, nCubes int32;
    assert(Abc_NtkIsStrash(pNtk), "Abc_NtkIsStrash(pNtk)");

    Abc_NtkCleanCopy(pNtk);
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(pNtk)) && (((pObj) = Abc_NtkCo(pNtk, i)), 1); i++  {
        if !Abc_ObjIsNode(Abc_ObjFanin0(pObj)) {
            fmt.Printf("%-20s :  Trivial.\n", Abc_ObjName(pObj));
            continue;
        }

        vSupp  = Abc_NtkNodeSupport( pNtk, &pObj, 1 );
        vNodes = Abc_NtkDfsNodes( pNtk, &pObj, 1 );

        fmt.Printf("%20s :  Cone = %5d.  Supp = %5d.  ", Abc_ObjName(pObj), vNodes.nSize, vSupp.nSize);
//        if ( vSupp->nSize <= 128 )
        {
            p = Min_ManAlloc( vSupp.nSize );
            nCubes = Abc_NodeDeriveEsops( p, pObj, vSupp, vNodes );
            fmt.Printf("Cubes = %5d.  ", nCubes);
            Min_ManFree( p );
        }
        fmt.Printf("\n");


        Vec_PtrFree( vNodes );
        Vec_PtrFree( vSupp );
    }
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

