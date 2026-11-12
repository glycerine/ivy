package abci

import "fmt"

/**CFile****************************************************************

  FileName    [abcAttach.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Network and node package.]

  Synopsis    [Attaches the library gates to the current network.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: abcAttach.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "base/abc/abc.h"
//#include "base/main/main.h"
//#include "map/mio/mio.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////
 
#define    ATTACH_FULL             (~((unsigned)0))
// Original macro:
// #define    ATTACH_MASK(n)         ((~((unsigned)0)) >> (32-(n)))


// static void Abc_AttachSetupTruthTables( unsigned uTruths[][2] );

// static void Abc_AttachComputeTruth( char * pSop, unsigned uTruthsIn[][2], unsigned * uTruthNode );

// static Mio_Gate_t * Abc_AttachFind( Mio_Gate_t ** ppGates, unsigned ** puTruthGates, int nGates, unsigned * uTruthNode, int * Perm );

// static int Abc_AttachCompare( unsigned ** puTruthGates, int nGates, unsigned * uTruthNode );

// static int Abc_NodeAttach( Abc_Obj_t * pNode, Mio_Gate_t ** ppGates, unsigned ** puTruthGates, int nGates, unsigned uTruths[][2] );

// static void Abc_TruthPermute( char * pPerm, int nVars, unsigned * uTruthNode, unsigned * uTruthPerm );


var s_pPerms *string = nil;
var s_nPerms int32;

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Attaches gates from the current library to the internal nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkAttach(pNtk *Abc_Ntk_t) int32 {
    var pGenlib *Mio_Library_t;
    var puTruthGates **uint32;
    var uTruths [6][2]uint32;
    var pNode *Abc_Obj_t;
    var ppGates **Mio_Gate_t;
    var nGates, nFanins, i int32;

    assert(Abc_NtkIsSopLogic(pNtk), "Abc_NtkIsSopLogic(pNtk)");

    // check that the library is available
    pGenlib = (Mio_Library_t *)Abc_FrameReadLibGen();
    if pGenlib == nil {
        fmt.Printf("The current library is not available.\n");
        return 0;
    }

    // start the truth tables
    Abc_AttachSetupTruthTables( uTruths );
    
    // collect all the gates
    ppGates = Mio_CollectRoots( pGenlib, 6, float32(1.0e+20), 1, &nGates, 0 );

    // derive the gate truth tables
    // inlining of macro ABC_ALLOC
    puTruthGates    = ((unsigned * *) malloc(sizeof(unsigned *) * (size_t)(nGates)));
    // inlining of macro ABC_ALLOC
    puTruthGates[0] = ((unsigned *) malloc(sizeof(unsigned) * (size_t)(2 * nGates)));
    for  i = 1; i < nGates; i++  {
        puTruthGates[i] = puTruthGates[i-1] + 2;
    }
    for  i = 0; i < nGates; i++  {
        Mio_DeriveTruthTable( ppGates[i], uTruths, Mio_GateReadPinNum(ppGates[i]), 6, puTruthGates[i] );
    }

    // assign the gates to pNode->pCopy
    Abc_NtkCleanCopy( pNtk );
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pNode) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pNode) == nil || !Abc_ObjIsNode(pNode) {} else {
        nFanins = Abc_ObjFaninNum(pNode);
        if nFanins == 0 {
            if Abc_SopIsConst1((char *)pNode.pData) {
                pNode.pCopy = (Abc_Obj_t *)Mio_LibraryReadConst1(pGenlib);
            } else {
                pNode.pCopy = (Abc_Obj_t *)Mio_LibraryReadConst0(pGenlib); }
        
        } else if nFanins == 1 {
            if Abc_SopIsBuf((char *)pNode.pData) {
                pNode.pCopy = (Abc_Obj_t *)Mio_LibraryReadBuf(pGenlib);
            } else {
                pNode.pCopy = (Abc_Obj_t *)Mio_LibraryReadInv(pGenlib); }
        
        } else if nFanins > 6 {
            fmt.Printf("Cannot attach gate with more than 6 inputs to node %s.\n", Abc_ObjName(pNode));
            // inlining of macro ABC_FREE
            ((puTruthGates[0]) ? (free((char *) (puTruthGates[0])), (puTruthGates[0]) = 0) : 0);
            // inlining of macro ABC_FREE
            ((puTruthGates) ? (free((char *) (puTruthGates)), (puTruthGates) = 0) : 0);
            // inlining of macro ABC_FREE
            ((ppGates) ? (free((char *) (ppGates)), (ppGates) = 0) : 0);
            return 0;
        
        } else if !(Abc_NodeAttach( pNode, ppGates, puTruthGates, nGates, uTruths ) != 0) {
            fmt.Printf("Could not attach the library gate to node %s.\n", Abc_ObjName(pNode));
            // inlining of macro ABC_FREE
            ((puTruthGates[0]) ? (free((char *) (puTruthGates[0])), (puTruthGates[0]) = 0) : 0);
            // inlining of macro ABC_FREE
            ((puTruthGates) ? (free((char *) (puTruthGates)), (puTruthGates) = 0) : 0);
            // inlining of macro ABC_FREE
            ((ppGates) ? (free((char *) (ppGates)), (ppGates) = 0) : 0);
            return 0;
        }
    }
    }
    // inlining of macro ABC_FREE
    ((puTruthGates[0]) ? (free((char *) (puTruthGates[0])), (puTruthGates[0]) = 0) : 0);
    // inlining of macro ABC_FREE
    ((puTruthGates) ? (free((char *) (puTruthGates)), (puTruthGates) = 0) : 0);
    // inlining of macro ABC_FREE
    ((ppGates) ? (free((char *) (ppGates)), (ppGates) = 0) : 0);
    // inlining of macro ABC_FREE
    ((s_pPerms) ? (free((char *) (s_pPerms)), (s_pPerms) = 0) : 0);

    // perform the final transformation
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pNode) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pNode) == nil || !Abc_ObjIsNode(pNode) {} else {
        if pNode.pCopy == nil {
            fmt.Printf("Some elementary gates (constant, buffer, or inverter) are missing in the library.\n");
            return 0;
        }
    }
    }

    // replace SOP representation by the gate representation
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pNode) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pNode) == nil || !Abc_ObjIsNode(pNode) {} else {
        pNode.pData = pNode.pCopy, pNode.pCopy = nil; }
    }
    pNtk.ntkFunc = ABC_FUNC_MAP;
    Extra_MmFlexStop( (Extra_MmFlex_t *)pNtk.pManFunc );
    pNtk.pManFunc = pGenlib;

    fmt.Printf("Library gates are successfully attached to the nodes.\n");

    // make sure that everything is okay
    if !Abc_NtkCheck( pNtk ) {
        fmt.Printf("Abc_NtkAttach: The network check has failed.\n");
        return 0;
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NodeAttach(pNode *Abc_Obj_t, ppGates **Mio_Gate_t, puTruthGates **uint32, nGates int32, uTruths [][2]uint32) int32 {
    var Perm [10]int32;
    var pTempInts [10]int32;
    var uTruthNode [2]uint32;
    var pFanin *Abc_Obj_t;
    var pGate *Mio_Gate_t;
    var nFanins, i int32;

    // compute the node's truth table
    Abc_AttachComputeTruth( (char *)pNode.pData, uTruths, uTruthNode );
    // find the matching gate and permutation
    pGate = Abc_AttachFind( ppGates, puTruthGates, nGates, uTruthNode, Perm );
    if pGate == nil {
        return 0; }
    // permute the fanins
    nFanins = Abc_ObjFaninNum(pNode);
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pNode)) && (((pFanin) = Abc_ObjFanin(pNode, i)), 1); i++  {
        pTempInts[i] = pFanin.Id;
    }
    for  i = 0; i < nFanins; i++  {
        pNode.vFanins.pArray[Perm[i]] = pTempInts[i];
    }
    // set the gate
    pNode.pCopy = (Abc_Obj_t *)pGate;
    return 1;
}

/**Function*************************************************************

  Synopsis    [Sets up the truth tables.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_AttachSetupTruthTables(uTruths [][2]uint32) {
    var m, v int32;
    for  v = 0; v < 5; v++  {
        uTruths[v][0] = 0;
    }
    // set up the truth tables
    for  m = 0; m < 32; m++  {
        for  v = 0; v < 5; v++  {
            if m & (1 << v) {
                uTruths[v][0] |= (1 << m); }
        }
    }
    // make adjustments for the case of 6 variables
    for  v = 0; v < 5; v++  {
        uTruths[v][1] = uTruths[v][0];
    }
    uTruths[5][0] = 0;
    uTruths[5][1] = ATTACH_FULL;
}

/**Function*************************************************************

  Synopsis    [Compute the truth table of the node's cover.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_AttachComputeTruth(pSop string, uTruthsIn [][2]uint32, uTruthRes *uint32) {
//    Mvc_Cube_t * pCube;
    var uSignCube [2]uint32;
    var Value int32;
//    int nInputs = pCover->nBits/2;
    var nInputs int32  = 6;
    var nFanins int32  = Abc_SopGetVarNum(pSop);
    var pCube string;
    var k int32;

    // make sure that the number of input truth tables in equal to the number of gate inputs
    assert(nInputs < 7, "nInputs < 7");

    // clean the resulting truth table
    uTruthRes[0] = 0;
    uTruthRes[1] = 0;
    if nInputs < 6 {
        // consider the case when only one unsigned can be used
//        Mvc_CoverForEachCube( pCover, pCube )
        // inlining of macro Abc_SopForEachCube
        for  pCube = (pSop); *pCube; pCube += (nFanins) + 3  {
            uSignCube[0] = ATTACH_FULL;
//            Mvc_CubeForEachVarValue( pCover, pCube, Var, Value )
            // inlining of macro Abc_CubeForEachVar
            for  k = 0; (pCube[k] != ' ') && (Value = pCube[k]); k++  {
                if Value == '0' {
                    uSignCube[0] &= ~uTruthsIn[k][0];
                } else if Value == '1' {
                    uSignCube[0] &=  uTruthsIn[k][0]; }
            }
            uTruthRes[0] |= uSignCube[0];
        }
        if Abc_SopGetPhase(pSop) == 0 {
            uTruthRes[0] = ~uTruthRes[0]; }
        if nInputs < 5 {
            // inlining of macro ATTACH_MASK
            uTruthRes[0] &= ((~(uint32(0))) >> (32-(1<<nInputs))); }
    
    } else {
        // consider the case when two unsigneds should be used
//        Mvc_CoverForEachCube( pCover, pCube )
        // inlining of macro Abc_SopForEachCube
        for  pCube = (pSop); *pCube; pCube += (nFanins) + 3  {
            uSignCube[0] = ATTACH_FULL;
            uSignCube[1] = ATTACH_FULL;
//            Mvc_CubeForEachVarValue( pCover, pCube, Var, Value )
            // inlining of macro Abc_CubeForEachVar
            for  k = 0; (pCube[k] != ' ') && (Value = pCube[k]); k++  {
                if Value == '0' {
                    uSignCube[0] &= ~uTruthsIn[k][0];
                    uSignCube[1] &= ~uTruthsIn[k][1];
                
                } else if Value == '1' {
                    uSignCube[0] &=  uTruthsIn[k][0];
                    uSignCube[1] &=  uTruthsIn[k][1];
                }
            }
            uTruthRes[0] |= uSignCube[0];
            uTruthRes[1] |= uSignCube[1];
        }

        // complement if the SOP is complemented
        if Abc_SopGetPhase(pSop) == 0 {
            uTruthRes[0] = ~uTruthRes[0];
            uTruthRes[1] = ~uTruthRes[1];
        }
    }
}

/**Function*************************************************************

  Synopsis    [Find the gate by truth table.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_AttachFind(ppGates **Mio_Gate_t, puTruthGates **uint32, nGates int32, uTruthNode *uint32, Perm *int32) *Mio_Gate_t {
    var uTruthPerm [2]uint32;
    var i, v, iNum int32;

    // try the gates without permutation
    if (iNum = Abc_AttachCompare( puTruthGates, nGates, uTruthNode )) >= 0 {
        for  v = 0; v < 6; v++  {
            Perm[v] = v;
        }
        return ppGates[iNum];
    }
    // get permutations
    if s_pPerms == nil {
        s_pPerms = Extra_Permutations( 6 );
        s_nPerms = Extra_Factorial( 6 );
    }
    // try permutations
    for  i = 0; i < s_nPerms; i++  {
        Abc_TruthPermute( s_pPerms[i], 6, uTruthNode, uTruthPerm );
        if (iNum = Abc_AttachCompare( puTruthGates, nGates, uTruthPerm )) >= 0 {
            for  v = 0; v < 6; v++  {
                Perm[v] = int32(s_pPerms[i][v]);
            }
            return ppGates[iNum];
        }
    }
    return nil;
}

/**Function*************************************************************

  Synopsis    [Find the gate by truth table.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_AttachCompare(puTruthGates **uint32, nGates int32, uTruthNode *uint32) int32 {
    var i int32;
    for  i = 0; i < nGates; i++  {
        if puTruthGates[i][0] == uTruthNode[0] && puTruthGates[i][1] == uTruthNode[1] {
            return i; }
    }
    return -1;
}

/**Function*************************************************************

  Synopsis    [Permutes the 6-input truth table.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_TruthPermute(pPerm string, nVars int32, uTruthNode *uint32, uTruthPerm *uint32) {
    var nMints, iMintPerm, iMint, v int32;
    uTruthPerm[0] = uTruthPerm[1] = 0;
    nMints = (1 << nVars);
    for  iMint = 0; iMint < nMints; iMint++  {
        if (uTruthNode[iMint>>5] & (1 << (iMint&31))) == 0 {
            continue; }
        iMintPerm = 0;
        for  v = 0; v < nVars; v++  {
            if iMint & (1 << v) {
                iMintPerm |= (1 << pPerm[v]); }
        }
        uTruthPerm[iMintPerm>>5] |= (1 << (iMintPerm&31));     
    }
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

