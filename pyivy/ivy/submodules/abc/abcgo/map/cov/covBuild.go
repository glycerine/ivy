package cov

import "fmt"

/**CFile****************************************************************

  FileName    [covBuild.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Mapping into network of SOPs/ESOPs.]

  Synopsis    [Network construction procedures.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: covBuild.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

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

  Synopsis    [Derives the decomposed network.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkCovDeriveCube(pNtkNew *Abc_Ntk_t, pObj *Abc_Obj_t, pCube *Min_Cube_t, vSupp *Vec_Int_t, fCompl int32) *Abc_Obj_t {
    var vLits *Vec_Int_t;
    var pNodeNew, pFanin *Abc_Obj_t;
    var i, iFanin, Lit int32;
    // create empty cube
    if pCube.nLits == 0 {
        if fCompl != 0 {
            return Abc_NtkCreateNodeConst0(pNtkNew); }
        return Abc_NtkCreateNodeConst1(pNtkNew);
    }
    // get the literals of this cube
    vLits = Vec_IntAlloc( 10 );
    Min_CubeGetLits( pCube, vLits );
    assert(pCube.nLits == uint32(vLits.nSize), "pCube.nLits == uint32(vLits.nSize)");
    // create special case when there is only one literal
    if pCube.nLits == 1 {
        iFanin = Vec_IntEntry(vLits,0);
        pFanin = Abc_NtkObj( pObj.pNtk, Vec_IntEntry(vSupp, iFanin) );
        Lit = Min_CubeGetVar(pCube, iFanin);
        assert(Lit == 1 || Lit == 2, "Lit == 1 || Lit == 2");
        Vec_IntFree( vLits );
        if (Lit == 1) ^ fCompl {// negative
            return Abc_NtkCreateNodeInv( pNtkNew, pFanin.pCopy ); }
        return pFanin.pCopy;
    }
    assert(pCube.nLits > 1, "pCube.nLits > 1");
    // create the AND cube
    pNodeNew = Abc_NtkCreateNode( pNtkNew );
    for  i = 0; i < vLits.nSize; i++  {
        iFanin = Vec_IntEntry(vLits,i);
        pFanin = Abc_NtkObj( pObj.pNtk, Vec_IntEntry(vSupp, iFanin) );
        Lit = Min_CubeGetVar(pCube, iFanin);
        assert(Lit == 1 || Lit == 2, "Lit == 1 || Lit == 2");
        Vec_IntWriteEntry( vLits, i, Lit==1 );
        Abc_ObjAddFanin( pNodeNew, pFanin.pCopy );
    }
    pNodeNew.pData = Abc_SopCreateAnd( (Mem_Flex_t *)pNtkNew.pManFunc, vLits.nSize, vLits.pArray );
    if fCompl != 0 {
        Abc_SopComplement( (char *)pNodeNew.pData ); }
    Vec_IntFree( vLits );
    return pNodeNew;
}

/**Function*************************************************************

  Synopsis    [Derives the decomposed network.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkCovDeriveNode_rec(p *Cov_Man_t, pNtkNew *Abc_Ntk_t, pObj *Abc_Obj_t, Level int32) *Abc_Obj_t {
    var pCover *Min_Cube_t = nil;
    var pCube *Min_Cube_t;
    var pFaninNew, pNodeNew, pFanin *Abc_Obj_t;
    var vSupp *Vec_Int_t;
    var Entry, nCubes, i int32;

    if Abc_ObjIsCi(pObj) {
        return pObj.pCopy; }
    assert(Abc_ObjIsNode(pObj), "Abc_ObjIsNode(pObj)");
    // skip if already computed
    if pObj.pCopy {
        return pObj.pCopy; }

    // get the support and the cover
    vSupp  = Abc_ObjGetSupp( pObj ); 
    pCover = Abc_ObjGetCover2( pObj );
    assert((vSupp != nil), "vSupp");
/*
    if ( pCover &&  pCover->nVars - Min_CoverSuppVarNum(p->pManMin, pCover) > 0 )
    {
        printf( "%d\n ", pCover->nVars - Min_CoverSuppVarNum(p->pManMin, pCover) );
        Min_CoverWrite( stdout, pCover );
    }
*/
/*
    // print the support of this node
    printf( "{ " );
    Vec_IntForEachEntry( vSupp, Entry, i )
        printf( "%d ", Entry );
    printf( "}  cubes = %d\n", Min_CoverCountCubes( pCover ) );
*/
    // process the fanins
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vSupp)) && (((Entry) = Vec_IntEntry(vSupp, i)), 1); i++  {
        pFanin = Abc_NtkObj(pObj.pNtk, Entry);
        Abc_NtkCovDeriveNode_rec( p, pNtkNew, pFanin, Level+1 );
    }

    // for each cube, construct the node
    nCubes = Min_CoverCountCubes( pCover );
    if nCubes == 0 {
        pNodeNew = Abc_NtkCreateNodeConst0(pNtkNew);
    } else if nCubes == 1 {
        pNodeNew = Abc_NtkCovDeriveCube( pNtkNew, pObj, pCover, vSupp, 0 );
    } else {
        pNodeNew = Abc_NtkCreateNode( pNtkNew );
        // inlining of macro Min_CoverForEachCube
        for  pCube = pCover;                                                 (pCube != nil);                                                          pCube = pCube.pNext  {
            pFaninNew = Abc_NtkCovDeriveCube( pNtkNew, pObj, pCube, vSupp, 0 );
            Abc_ObjAddFanin( pNodeNew, pFaninNew );
        }
        pNodeNew.pData = Abc_SopCreateXorSpecial( (Mem_Flex_t *)pNtkNew.pManFunc, nCubes );
    }
/*
    printf( "Created node %d(%d) at level %d: ", pNodeNew->Id, pObj->Id, Level );
    Vec_IntForEachEntry( vSupp, Entry, i )
    {
        pFanin = Abc_NtkObj(pObj->pNtk, Entry);
        printf( "%d(%d) ", pFanin->pCopy->Id, pFanin->Id );
    }
    printf( "\n" );
    Min_CoverWrite( stdout, pCover );
*/
    pObj.pCopy = pNodeNew;
    return pNodeNew;
}

/**Function*************************************************************

  Synopsis    [Derives the decomposed network.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkCovDerive(p *Cov_Man_t, pNtk *Abc_Ntk_t) *Abc_Ntk_t {
    var pNtkNew *Abc_Ntk_t;
    var pObj *Abc_Obj_t;
    var i int32;
    assert(Abc_NtkIsStrash(pNtk), "Abc_NtkIsStrash(pNtk)");
    // perform strashing
    pNtkNew = Abc_NtkStartFrom( pNtk, ABC_NTK_LOGIC, ABC_FUNC_SOP );
    // reconstruct the network
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(pNtk)) && (((pObj) = Abc_NtkCo(pNtk, i)), 1); i++  {
        Abc_NtkCovDeriveNode_rec( p, pNtkNew, Abc_ObjFanin0(pObj), 0 );
//        printf( "*** CO %s : %d -> %d \n", Abc_ObjName(pObj), pObj->pCopy->Id, Abc_ObjFanin0(pObj)->pCopy->Id );
    }
    // add the COs
    Abc_NtkFinalize( pNtk, pNtkNew );
    Abc_NtkLogicMakeSimpleCos( pNtkNew, 1 );
    // make sure everything is okay
    if !Abc_NtkCheck( pNtkNew ) {
        fmt.Printf("Abc_NtkCovDerive: The network check has failed.\n");
        Abc_NtkDelete( pNtkNew );
        return nil;
    }
    return pNtkNew;
}




/**Function*************************************************************

  Synopsis    [Derives the decomposed network.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkCovDeriveInv(pNtkNew *Abc_Ntk_t, pObj *Abc_Obj_t, fCompl int32) *Abc_Obj_t {
    assert(pObj.pCopy, "pObj.pCopy");
    if !(fCompl != 0) {
        return pObj.pCopy; }
    if pObj.pCopy.pCopy == nil {
        pObj.pCopy.pCopy = Abc_NtkCreateNodeInv( pNtkNew, pObj.pCopy ); }
    return pObj.pCopy.pCopy;
 }

/**Function*************************************************************

  Synopsis    [Derives the decomposed network.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkCovDeriveCubeInv(pNtkNew *Abc_Ntk_t, pObj *Abc_Obj_t, pCube *Min_Cube_t, vSupp *Vec_Int_t) *Abc_Obj_t {
    var vLits *Vec_Int_t;
    var pNodeNew, pFanin *Abc_Obj_t;
    var i, iFanin, Lit int32;
    // create empty cube
    if pCube.nLits == 0 {
        return Abc_NtkCreateNodeConst1(pNtkNew); }
    // get the literals of this cube
    vLits = Vec_IntAlloc( 10 );
    Min_CubeGetLits( pCube, vLits );
    assert(pCube.nLits == uint32(vLits.nSize), "pCube.nLits == uint32(vLits.nSize)");
    // create special case when there is only one literal
    if pCube.nLits == 1 {
        iFanin = Vec_IntEntry(vLits,0);
        pFanin = Abc_NtkObj( pObj.pNtk, Vec_IntEntry(vSupp, iFanin) );
        Lit = Min_CubeGetVar(pCube, iFanin);
        assert(Lit == 1 || Lit == 2, "Lit == 1 || Lit == 2");
        Vec_IntFree( vLits );
//        if ( Lit == 1 )// negative
//            return Abc_NtkCreateNodeInv( pNtkNew, pFanin->pCopy );
//        return pFanin->pCopy;
        return Abc_NtkCovDeriveInv( pNtkNew, pFanin, Lit==1 );
    }
    assert(pCube.nLits > 1, "pCube.nLits > 1");
    // create the AND cube
    pNodeNew = Abc_NtkCreateNode( pNtkNew );
    for  i = 0; i < vLits.nSize; i++  {
        iFanin = Vec_IntEntry(vLits,i);
        pFanin = Abc_NtkObj( pObj.pNtk, Vec_IntEntry(vSupp, iFanin) );
        Lit = Min_CubeGetVar(pCube, iFanin);
        assert(Lit == 1 || Lit == 2, "Lit == 1 || Lit == 2");
        Vec_IntWriteEntry( vLits, i, Lit==1 );
//        Abc_ObjAddFanin( pNodeNew, pFanin->pCopy );
        Abc_ObjAddFanin( pNodeNew, Abc_NtkCovDeriveInv( pNtkNew, pFanin, Lit==1 ) );
    }
//    pNodeNew->pData = Abc_SopCreateAnd( (Mem_Flex_t *)pNtkNew->pManFunc, vLits->nSize, vLits->pArray );
    pNodeNew.pData = Abc_SopCreateAnd( (Mem_Flex_t *)pNtkNew.pManFunc, vLits.nSize, nil );
    Vec_IntFree( vLits );
    return pNodeNew;
}

/**Function*************************************************************

  Synopsis    [Derives the decomposed network.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkCovDeriveNodeInv_rec(p *Cov_Man_t, pNtkNew *Abc_Ntk_t, pObj *Abc_Obj_t, fCompl int32) *Abc_Obj_t {
    var pCover, pCube *Min_Cube_t;
    var pFaninNew, pNodeNew, pFanin *Abc_Obj_t;
    var vSupp *Vec_Int_t;
    var Entry, nCubes, i int32;

    // skip if already computed
    if pObj.pCopy {
        return Abc_NtkCovDeriveInv( pNtkNew, pObj, fCompl ); }
    assert(Abc_ObjIsNode(pObj), "Abc_ObjIsNode(pObj)");

    // get the support and the cover
    vSupp  = Abc_ObjGetSupp( pObj ); 
    pCover = Abc_ObjGetCover2( pObj );
    assert((vSupp != nil), "vSupp");

    // process the fanins
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vSupp)) && (((Entry) = Vec_IntEntry(vSupp, i)), 1); i++  {
        pFanin = Abc_NtkObj(pObj.pNtk, Entry);
        Abc_NtkCovDeriveNodeInv_rec( p, pNtkNew, pFanin, 0 );
    }

    // for each cube, construct the node
    nCubes = Min_CoverCountCubes( pCover );
    if nCubes == 0 {
        pNodeNew = Abc_NtkCreateNodeConst0(pNtkNew);
    } else if nCubes == 1 {
        pNodeNew = Abc_NtkCovDeriveCubeInv( pNtkNew, pObj, pCover, vSupp );
    } else {
        pNodeNew = Abc_NtkCreateNode( pNtkNew );
        // inlining of macro Min_CoverForEachCube
        for  pCube = pCover;                                                 (pCube != nil);                                                          pCube = pCube.pNext  {
            pFaninNew = Abc_NtkCovDeriveCubeInv( pNtkNew, pObj, pCube, vSupp );
            Abc_ObjAddFanin( pNodeNew, pFaninNew );
        }
        pNodeNew.pData = Abc_SopCreateXorSpecial( (Mem_Flex_t *)pNtkNew.pManFunc, nCubes );
    }

    pObj.pCopy = pNodeNew;
    return Abc_NtkCovDeriveInv( pNtkNew, pObj, fCompl );
}

/**Function*************************************************************

  Synopsis    [Derives the decomposed network.]

  Description [The resulting network contains only pure AND/OR/EXOR gates
  and inverters. This procedure is usedful to generate Verilog.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkCovDeriveClean(p *Cov_Man_t, pNtk *Abc_Ntk_t) *Abc_Ntk_t {
    var pNtkNew *Abc_Ntk_t;
    var pObj, pNodeNew *Abc_Obj_t;
    var i int32;
    assert(Abc_NtkIsStrash(pNtk), "Abc_NtkIsStrash(pNtk)");
    // perform strashing
    pNtkNew = Abc_NtkStartFrom( pNtk, ABC_NTK_LOGIC, ABC_FUNC_SOP );
    // reconstruct the network
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(pNtk)) && (((pObj) = Abc_NtkCo(pNtk, i)), 1); i++  {
        pNodeNew = Abc_NtkCovDeriveNodeInv_rec( p, pNtkNew, Abc_ObjFanin0(pObj), Abc_ObjFaninC0(pObj) );
        Abc_ObjAddFanin( pObj.pCopy, pNodeNew );
    }
    // add the COs
    Abc_NtkLogicMakeSimpleCos( pNtkNew, 0 );
    // make sure everything is okay
    if !Abc_NtkCheck( pNtkNew ) {
        fmt.Printf("Abc_NtkCovDeriveInv: The network check has failed.\n");
        Abc_NtkDelete( pNtkNew );
        return nil;
    }
    return pNtkNew;
}



/**Function*************************************************************

  Synopsis    [Derives the decomposed network.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkCovDerive_rec(p *Cov_Man_t, pNtkNew *Abc_Ntk_t, pObj *Abc_Obj_t) *Abc_Obj_t {
    var fVerbose int32;
    var pCover *Min_Cube_t;
    var pCovers [3]*Min_Cube_t;
    var pNodeNew, pFanin *Abc_Obj_t;
    var vSupp *Vec_Int_t;
    var vCover *Vec_Str_t;
    var i, Entry, nCubes, Type int32;
    // skip if already computed
    if pObj.pCopy {
        return pObj.pCopy; }
    assert(Abc_ObjIsNode(pObj), "Abc_ObjIsNode(pObj)");
    // get the support and the cover
    vSupp  = Abc_ObjGetSupp( pObj ); 
    assert((vSupp != nil), "vSupp");
    // choose the cover to implement
    pCovers[0] = Abc_ObjGetCover( pObj, 0 );
    pCovers[1] = Abc_ObjGetCover( pObj, 1 );
    pCovers[2] = Abc_ObjGetCover2( pObj );
    // use positive polarity
    if pCovers[0] 
        && (!pCovers[1] || Min_CoverCountCubes(pCovers[0]) <= Min_CoverCountCubes(pCovers[1]))
        && (!pCovers[2] || Min_CoverCountCubes(pCovers[0]) <= Min_CoverCountCubes(pCovers[2])) {
        pCover = pCovers[0];
        Type = '1';
    
    } else 
    // use negative polarity
    if pCovers[1] 
        && (!pCovers[0] || Min_CoverCountCubes(pCovers[1]) <= Min_CoverCountCubes(pCovers[0]))
        && (!pCovers[2] || Min_CoverCountCubes(pCovers[1]) <= Min_CoverCountCubes(pCovers[2])) {
        pCover = pCovers[1];
        Type = '0';
    
    } else 
    // use XOR polarity
    if pCovers[2] 
        && (!pCovers[0] || Min_CoverCountCubes(pCovers[2]) < Min_CoverCountCubes(pCovers[0]))
        && (!pCovers[1] || Min_CoverCountCubes(pCovers[2]) < Min_CoverCountCubes(pCovers[1])) {
        pCover = pCovers[2];
        Type = 'x';
    
    } else {
        assert(false, "0"); }
    // print the support of this node
    if fVerbose != 0 {
        fmt.Printf("{ ");
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(vSupp)) && (((Entry) = Vec_IntEntry(vSupp, i)), 1); i++  {
            fmt.Printf("%d ", Entry);
        }
        fmt.Printf("}  cubes = %d\n", Min_CoverCountCubes( pCover ));
    }
    // process the fanins
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vSupp)) && (((Entry) = Vec_IntEntry(vSupp, i)), 1); i++  {
        pFanin = Abc_NtkObj(pObj.pNtk, Entry);
        Abc_NtkCovDerive_rec( p, pNtkNew, pFanin );
    }
    // for each cube, construct the node
    nCubes = Min_CoverCountCubes( pCover );
    if nCubes == 0 {
        pNodeNew = Abc_NtkCreateNodeConst0(pNtkNew);
    } else if nCubes == 1 {
        pNodeNew = Abc_NtkCovDeriveCube( pNtkNew, pObj, pCover, vSupp, Type == '0' );
    } else {
        // create the node
        pNodeNew = Abc_NtkCreateNode( pNtkNew );
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(vSupp)) && (((Entry) = Vec_IntEntry(vSupp, i)), 1); i++  {
            pFanin = Abc_NtkObj(pObj.pNtk, Entry);
            Abc_ObjAddFanin( pNodeNew, pFanin.pCopy );
        }
        // derive the function
        vCover = Vec_StrAlloc( 100 );
        Min_CoverCreate( vCover, pCover, byte(Type) );
        pNodeNew.pData = Abc_SopRegister((Mem_Flex_t *)pNtkNew.pManFunc, Vec_StrArray(vCover) );
        Vec_StrFree( vCover );
    }

/*
    printf( "Created node %d(%d) at level %d: ", pNodeNew->Id, pObj->Id, Level );
    Vec_IntForEachEntry( vSupp, Entry, i )
    {
        pFanin = Abc_NtkObj(pObj->pNtk, Entry);
        printf( "%d(%d) ", pFanin->pCopy->Id, pFanin->Id );
    }
    printf( "\n" );
    Min_CoverWrite( stdout, pCover );
*/
    return pObj.pCopy = pNodeNew;
}

/**Function*************************************************************

  Synopsis    [Derives the decomposed network.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkCovDeriveRegular(p *Cov_Man_t, pNtk *Abc_Ntk_t) *Abc_Ntk_t {
    var pNtkNew *Abc_Ntk_t;
    var pObj, pNodeNew *Abc_Obj_t;
    var i int32;
    assert(Abc_NtkIsStrash(pNtk), "Abc_NtkIsStrash(pNtk)");
    // perform strashing
    pNtkNew = Abc_NtkStartFrom( pNtk, ABC_NTK_LOGIC, ABC_FUNC_SOP );
    // reconstruct the network
    if Abc_ObjFanoutNum(Abc_AigConst1(pNtk)) > 0 {
        Abc_AigConst1(pNtk).pCopy = Abc_NtkCreateNodeConst1(pNtkNew); }
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(pNtk)) && (((pObj) = Abc_NtkCo(pNtk, i)), 1); i++  {
        pNodeNew = Abc_NtkCovDerive_rec( p, pNtkNew, Abc_ObjFanin0(pObj) );
        if Abc_ObjFaninC0(pObj) {
            if pNodeNew.pData && Abc_ObjFanoutNum(Abc_ObjFanin0(pObj)) == 1 {
                Abc_SopComplement( (char *)pNodeNew.pData );
            } else {
                pNodeNew = Abc_NtkCreateNodeInv( pNtkNew, pNodeNew ); }
        }
        Abc_ObjAddFanin( pObj.pCopy, pNodeNew );
    }
    // add the COs
    Abc_NtkLogicMakeSimpleCos( pNtkNew, 0 );
    // make sure everything is okay
    if !Abc_NtkCheck( pNtkNew ) {
        fmt.Printf("Abc_NtkCovDerive: The network check has failed.\n");
        Abc_NtkDelete( pNtkNew );
        return nil;
    }
    return pNtkNew;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

