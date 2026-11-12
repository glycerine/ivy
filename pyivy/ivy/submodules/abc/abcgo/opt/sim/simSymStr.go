package sim

import "fmt"

/**CFile****************************************************************

  FileName    [simSymStr.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Network and node package.]

  Synopsis    [Structural detection of symmetries.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: simSymStr.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "base/abc/abc.h"
//#include "sim.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// Original macro:
// #define SIM_READ_SYMMS(pNode)       ((Vec_Int_t *)pNode->pCopy) 

// Original macro:
// #define SIM_SET_SYMMS(pNode,vVect)  (pNode->pCopy = (Abc_Obj_t *)(vVect)) 


// static void  Sim_SymmsStructComputeOne( Abc_Ntk_t * pNtk, Abc_Obj_t * pNode, int * pMap );

// static void  Sim_SymmsBalanceCollect_rec( Abc_Obj_t * pNode, Vec_Ptr_t * vNodes );

// static void  Sim_SymmsPartitionNodes( Vec_Ptr_t * vNodes, Vec_Ptr_t * vNodesPis0, Vec_Ptr_t * vNodesPis1, Vec_Ptr_t * vNodesOther );

// static void  Sim_SymmsAppendFromGroup( Abc_Ntk_t * pNtk, Vec_Ptr_t * vNodesPi, Vec_Ptr_t * vNodesOther, Vec_Int_t * vSymms, int * pMap );

// static void  Sim_SymmsAppendFromNode( Abc_Ntk_t * pNtk, Vec_Int_t * vSymms0, Vec_Ptr_t * vNodesOther, Vec_Ptr_t * vNodesPi0, Vec_Ptr_t * vNodesPi1, Vec_Int_t * vSymms, int * pMap );

// static int   Sim_SymmsIsCompatibleWithNodes( Abc_Ntk_t * pNtk, unsigned uSymm, Vec_Ptr_t * vNodesOther, int * pMap );

// static int   Sim_SymmsIsCompatibleWithGroup( unsigned uSymm, Vec_Ptr_t * vNodesPi, int * pMap );

// static void  Sim_SymmsPrint( Vec_Int_t * vSymms );

// static void  Sim_SymmsTrans( Vec_Int_t * vSymms );

// static void  Sim_SymmsTransferToMatrix( Extra_BitMat_t * pMatSymm, Vec_Int_t * vSymms, unsigned * pSupport );

// static int * Sim_SymmsCreateMap( Abc_Ntk_t * pNtk );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Computes symmetries for a single output function.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sim_SymmsStructCompute(pNtk *Abc_Ntk_t, vMatrs *Vec_Ptr_t, vSuppFun *Vec_Ptr_t) {
    var vNodes *Vec_Ptr_t;
    var pTemp *Abc_Obj_t;
    var pMap *int32;
    var i int32;

    assert(Abc_NtkCiNum(pNtk) + 10 < (1<<16), "Abc_NtkCiNum(pNtk) + 10 < (1<<16)");

    // get the structural support
    pNtk.vSupps = Sim_ComputeStrSupp( pNtk );
    // set elementary info for the CIs
    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(pNtk)) && (((pTemp) = Abc_NtkCi(pNtk, i)), 1); i++  {
        // inlining of macro SIM_SET_SYMMS
        (pTemp.pCopy = (Abc_Obj_t *)(Vec_IntAlloc(0))) ;
    }
    // create the map of CI ids into their numbers
    pMap = Sim_SymmsCreateMap( pNtk );
    // collect the nodes in the TFI cone of this output
    vNodes = Abc_NtkDfs( pNtk, 0 );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pTemp) = (Abc_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
//        if ( Abc_NodeIsConst(pTemp) )
//            continue;
        Sim_SymmsStructComputeOne( pNtk, pTemp, pMap );
    }
    // collect the results for the COs;
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(pNtk)) && (((pTemp) = Abc_NtkCo(pNtk, i)), 1); i++  {
//printf( "Output %d:\n", i );
        pTemp = Abc_ObjFanin0(pTemp);
        if Abc_ObjIsCi(pTemp) || Abc_AigNodeIsConst(pTemp) {
            continue; }
        // inlining of macro SIM_READ_SYMMS
        Sim_SymmsTransferToMatrix( (Extra_BitMat_t *)Vec_PtrEntry(vMatrs, i), ((Vec_Int_t *)pTemp.pCopy) , (unsigned *)Vec_PtrEntry(vSuppFun, i) );
    }
    // clean the intermediate results
    Sim_UtilInfoFree( pNtk.vSupps );
    pNtk.vSupps = nil;
    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(pNtk)) && (((pTemp) = Abc_NtkCi(pNtk, i)), 1); i++  {
        // inlining of macro SIM_READ_SYMMS
        Vec_IntFree( ((Vec_Int_t *)pTemp.pCopy)  );
    }
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pTemp) = (Abc_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
//        if ( !Abc_NodeIsConst(pTemp) )
            // inlining of macro SIM_READ_SYMMS
            Vec_IntFree( ((Vec_Int_t *)pTemp.pCopy)  );
    }
    Vec_PtrFree( vNodes );
    // inlining of macro ABC_FREE
    ((pMap) ? (free((char *) (pMap)), (pMap) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    [Recursively computes symmetries. ]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sim_SymmsStructComputeOne(pNtk *Abc_Ntk_t, pNode *Abc_Obj_t, pMap *int32) {
    var vNodes, vNodesPi0, vNodesPi1, vNodesOther *Vec_Ptr_t;
    var vSymms *Vec_Int_t;
    var pTemp *Abc_Obj_t;
    var i int32;

    // allocate the temporary arrays
    vNodes      = Vec_PtrAlloc( 10 );
    vNodesPi0   = Vec_PtrAlloc( 10 );
    vNodesPi1   = Vec_PtrAlloc( 10 );
    vNodesOther = Vec_PtrAlloc( 10 );  

    // collect the fanins of the implication supergate
    Sim_SymmsBalanceCollect_rec( pNode, vNodes );

    // sort the nodes in the implication supergate
    Sim_SymmsPartitionNodes( vNodes, vNodesPi0, vNodesPi1, vNodesOther );

    // start the resulting set
    vSymms = Vec_IntAlloc( 10 );
    // generate symmetries from the groups
    Sim_SymmsAppendFromGroup( pNtk, vNodesPi0, vNodesOther, vSymms, pMap );
    Sim_SymmsAppendFromGroup( pNtk, vNodesPi1, vNodesOther, vSymms, pMap );
    // add symmetries from other inputs
    for  i = 0; i < vNodesOther.nSize; i++  {
        pTemp = Abc_ObjRegular((Abc_Obj_t *)vNodesOther.pArray[i]);
        // inlining of macro SIM_READ_SYMMS
        Sim_SymmsAppendFromNode( pNtk, ((Vec_Int_t *)pTemp.pCopy) , vNodesOther, vNodesPi0, vNodesPi1, vSymms, pMap );
    }
    Vec_PtrFree( vNodes );
    Vec_PtrFree( vNodesPi0 );
    Vec_PtrFree( vNodesPi1 );
    Vec_PtrFree( vNodesOther );

    // set the symmetry at the node
    // inlining of macro SIM_SET_SYMMS
    (pNode.pCopy = (Abc_Obj_t *)(vSymms)) ;
}


/**Function*************************************************************

  Synopsis    [Returns the array of nodes to be combined into one multi-input AND-gate.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sim_SymmsBalanceCollect_rec(pNode *Abc_Obj_t, vNodes *Vec_Ptr_t) {
    // if the new node is complemented, another gate begins
    if Abc_ObjIsComplement(pNode) {
        Vec_PtrPushUnique( vNodes, pNode );
        return;
    }
    // if pNew is the PI node, return
    if Abc_ObjIsCi(pNode) {
        Vec_PtrPushUnique( vNodes, pNode );
        return;    
    }
    // go through the branches
    Sim_SymmsBalanceCollect_rec( Abc_ObjChild0(pNode), vNodes );
    Sim_SymmsBalanceCollect_rec( Abc_ObjChild1(pNode), vNodes );
}

/**Function*************************************************************

  Synopsis    [Divides PI variables into groups.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sim_SymmsPartitionNodes(vNodes *Vec_Ptr_t, vNodesPis0 *Vec_Ptr_t, vNodesPis1 *Vec_Ptr_t, vNodesOther *Vec_Ptr_t) {
    var pNode *Abc_Obj_t;
    var i int32;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pNode) = (Abc_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        if !Abc_ObjIsCi(Abc_ObjRegular(pNode)) {
            Vec_PtrPush( vNodesOther, pNode );
        } else if Abc_ObjIsComplement(pNode) {
            Vec_PtrPush( vNodesPis0, pNode );
        } else {
            Vec_PtrPush( vNodesPis1, pNode ); }
    }
}

/**Function*************************************************************

  Synopsis    [Makes the product of two partitions.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sim_SymmsAppendFromGroup(pNtk *Abc_Ntk_t, vNodesPi *Vec_Ptr_t, vNodesOther *Vec_Ptr_t, vSymms *Vec_Int_t, pMap *int32) {
    var pNode1, pNode2 *Abc_Obj_t;
    var uSymm uint32;
    var i, k int32;

    if vNodesPi.nSize == 0 {
        return; }

    // go through the pairs
    for  i = 0; i < vNodesPi.nSize; i++  {
    for  k = i+1; k < vNodesPi.nSize; k++  {
        // get the two PI nodes
        pNode1 = Abc_ObjRegular((Abc_Obj_t *)vNodesPi.pArray[i]);
        pNode2 = Abc_ObjRegular((Abc_Obj_t *)vNodesPi.pArray[k]);
        assert(pMap[pNode1.Id] != pMap[pNode2.Id], "pMap[pNode1.Id] != pMap[pNode2.Id]");
        assert(pMap[pNode1.Id] >= 0, "pMap[pNode1.Id] >= 0");
        assert(pMap[pNode2.Id] >= 0, "pMap[pNode2.Id] >= 0");
        // generate symmetry
        if pMap[pNode1.Id] < pMap[pNode2.Id] {
            uSymm = ((pMap[pNode1.Id] << 16) | pMap[pNode2.Id]);
        } else {
            uSymm = ((pMap[pNode2.Id] << 16) | pMap[pNode1.Id]); }
        // check if symmetry belongs
        if Sim_SymmsIsCompatibleWithNodes( pNtk, uSymm, vNodesOther, pMap ) != 0 {
            Vec_IntPushUnique( vSymms, int32(uSymm) ); }
    }
    }
}

/**Function*************************************************************

  Synopsis    [Add the filters symmetries from the nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sim_SymmsAppendFromNode(pNtk *Abc_Ntk_t, vSymms0 *Vec_Int_t, vNodesOther *Vec_Ptr_t, vNodesPi0 *Vec_Ptr_t, vNodesPi1 *Vec_Ptr_t, vSymms *Vec_Int_t, pMap *int32) {
    var uSymm uint32;
    var i int32;

    if vSymms0.nSize == 0 {
        return; }

    // go through the pairs
    for  i = 0; i < vSymms0.nSize; i++  {
        uSymm = uint32(vSymms0.pArray[i]);
        // check if symmetry belongs
        if (Sim_SymmsIsCompatibleWithNodes( pNtk, uSymm, vNodesOther, pMap ) != 0) &&
             (Sim_SymmsIsCompatibleWithGroup( uSymm, vNodesPi0, pMap ) != 0) && 
             (Sim_SymmsIsCompatibleWithGroup( uSymm, vNodesPi1, pMap ) != 0) {
            Vec_IntPushUnique( vSymms, int32(uSymm) ); }
    }
}

/**Function*************************************************************

  Synopsis    [Returns 1 if symmetry is compatible with the group of nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sim_SymmsIsCompatibleWithNodes(pNtk *Abc_Ntk_t, uSymm uint32, vNodesOther *Vec_Ptr_t, pMap *int32) int32 {
    var vSymmsNode *Vec_Int_t;
    var pNode *Abc_Obj_t;
    var i, s, Ind1, Ind2, fIsVar1, fIsVar2 int32;

    if vNodesOther.nSize == 0 {
        return 1; }

    // get the indices of the PI variables
    Ind1 = (uSymm & 0xffff);
    Ind2 = (uSymm >> 16);

    // go through the nodes
    // if they do not belong to a support, it is okay
    // if one belongs, the other does not belong, quit
    // if they belong, but are not part of symmetry, quit
    for  i = 0; i < vNodesOther.nSize; i++  {
        pNode = Abc_ObjRegular((Abc_Obj_t *)vNodesOther.pArray[i]);
        // inlining of macro Sim_SuppStrHasVar
        // inlining of macro Sim_HasBit
        fIsVar1 = ((((unsigned*)(pNtk.vSupps).pArray[(pNode).Id])[((Ind1))>>5]  & (1<<(((Ind1)) & 31))) > 0);
        // inlining of macro Sim_SuppStrHasVar
        // inlining of macro Sim_HasBit
        fIsVar2 = ((((unsigned*)(pNtk.vSupps).pArray[(pNode).Id])[((Ind2))>>5]  & (1<<(((Ind2)) & 31))) > 0);

        if !fIsVar1 && !(fIsVar2 != 0) {
            continue; }
        if fIsVar1 ^ fIsVar2 {
            return 0; }
        // both belong
        // check if there is a symmetry
        // inlining of macro SIM_READ_SYMMS
        vSymmsNode = ((Vec_Int_t *)pNode.pCopy) ;
        for  s = 0; s < vSymmsNode.nSize; s++  {
            if uSymm == uint32(vSymmsNode.pArray[s]) {
                break; }
        }
        if s == vSymmsNode.nSize {
            return 0; }
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Returns 1 if symmetry is compatible with the group of PIs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sim_SymmsIsCompatibleWithGroup(uSymm uint32, vNodesPi *Vec_Ptr_t, pMap *int32) int32 {
    var pNode *Abc_Obj_t;
    var i, Ind1, Ind2, fHasVar1, fHasVar2 int32;

    if vNodesPi.nSize == 0 {
        return 1; }

    // get the indices of the PI variables
    Ind1 = (uSymm & 0xffff);
    Ind2 = (uSymm >> 16);

    // go through the PI nodes
    fHasVar1 = fHasVar2 = 0;
    for  i = 0; i < vNodesPi.nSize; i++  {
        pNode = Abc_ObjRegular((Abc_Obj_t *)vNodesPi.pArray[i]);
        if pMap[pNode.Id] == Ind1 {
            fHasVar1 = 1;
        } else if pMap[pNode.Id] == Ind2 {
            fHasVar2 = 1; }
    }
    return fHasVar1 == fHasVar2;
}



/**Function*************************************************************

  Synopsis    [Improvements due to transitivity.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sim_SymmsTrans(vSymms *Vec_Int_t) {
    var uSymm, uSymma, uSymmr uint32;
    var i, Ind1, Ind2 int32;
    var k, Ind1a, Ind2a int32;
    var j int32;
    var nTrans int32;

    for  i = 0; i < vSymms.nSize; i++  {
        uSymm = uint32(vSymms.pArray[i]);
        Ind1 = (uSymm & 0xffff);
        Ind2 = (uSymm >> 16);
        // find other symmetries that have Ind1
        for  k = i+1; k < vSymms.nSize; k++  {
            uSymma = uint32(vSymms.pArray[k]);
            if uSymma == uSymm {
                continue; }
            Ind1a = (uSymma & 0xffff);
            Ind2a = (uSymma >> 16);
            if Ind1a == Ind1 {
                // find the symmetry (Ind2,Ind2a)
                if Ind2 < Ind2a {
                    uSymmr = ((Ind2 << 16) | Ind2a);
                } else {
                    uSymmr = ((Ind2a << 16) | Ind2); }
                for  j = 0; j < vSymms.nSize; j++  {
                    if uSymmr == uint32(vSymms.pArray[j]) {
                        break; }
                }
                if j == vSymms.nSize {
                    nTrans++; }
            }
        }

    }
    fmt.Printf("Trans = %d.\n", nTrans);
}


/**Function*************************************************************

  Synopsis    [Transfers from the vector to the matrix.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sim_SymmsTransferToMatrix(pMatSymm *Extra_BitMat_t, vSymms *Vec_Int_t, pSupport *uint32) {
    var i, Ind1, Ind2, nInputs int32;
    var uSymm uint32;
    // add diagonal elements
    nInputs = Extra_BitMatrixReadSize( pMatSymm );
    for  i = 0; i < nInputs; i++  {
        Extra_BitMatrixInsert1( pMatSymm, i, i );
    }
    // add non-diagonal elements
    for  i = 0; i < vSymms.nSize; i++  {
        uSymm = uint32(vSymms.pArray[i]);
        Ind1 = (uSymm & 0xffff);
        Ind2 = (uSymm >> 16);
//printf( "%d,%d ", Ind1, Ind2 );
        // skip variables that are not in the true support
        // inlining of macro Sim_HasBit
        assert((((pSupport)[(Ind1)>>5]  & (1<<((Ind1) & 31))) > 0) == (((pSupport)[(Ind2)>>5]  & (1<<((Ind2) & 31))) > 0), "/*__starting_point_inline_macro:Sim_HasBit__*/(((pSupport)[(Ind1)>>5]  & (1<<((Ind1) & 31))) > 0) == /*__starting_point_inline_macro:Sim_HasBit__*/(((pSupport)[(Ind2)>>5]  & (1<<((Ind2) & 31))) > 0)");
        // inlining of macro Sim_HasBit
        if !(((pSupport)[(Ind1)>>5]  & (1<<((Ind1) & 31))) > 0) || !(((pSupport)[(Ind2)>>5]  & (1<<((Ind2) & 31))) > 0) {
            continue; }
        Extra_BitMatrixInsert1( pMatSymm, Ind1, Ind2 );
        Extra_BitMatrixInsert2( pMatSymm, Ind1, Ind2 );
    }
}

/**Function*************************************************************

  Synopsis    [Mapping of indices into numbers.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sim_SymmsCreateMap(pNtk *Abc_Ntk_t) *int32 {
    var pMap *int32;
    var pNode *Abc_Obj_t;
    var i int32;
    // inlining of macro ABC_ALLOC
    pMap = ((int *) malloc(sizeof(int) * (size_t)(Abc_NtkObjNumMax(pNtk))));
    for  i = 0; i < Abc_NtkObjNumMax(pNtk); i++  {
        pMap[i] = -1;
    }
    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(pNtk)) && (((pNode) = Abc_NtkCi(pNtk, i)), 1); i++  {
        pMap[pNode.Id] = i;
    }
    return pMap;
}



////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

