package gia

import "fmt"

/**CFile****************************************************************

  FileName    [giaSplit.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [Structural AIG splitting.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaSplit.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"
//#include "misc/extra/extra.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

   
/**Function*************************************************************

  Synopsis    [Takes Nodes and Marks. Returns Leaves and Roots.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Computes LUTs that are fanouts of the node outside of the cone.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Returns the number of fanins beloning to the set.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Computing window for one pivot node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [External procedures.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Testing procedure.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Spl_ManComputeOneTest(pGia *Gia_Man_t) {
    var iLut, Count int32;
    Gia_ManComputeOneWinStart( pGia, 64, 0 );
    // inlining of macro Gia_ManForEachLut2
    for  iLut = 1; iLut < Gia_ManObjNum(pGia); iLut++  { if !Gia_ObjIsLut2(pGia, iLut) {} else {
        var vRoots, vNodes, vLeaves, vAnds *Vec_Int_t;
        Count = Gia_ManComputeOneWin( pGia, iLut, &vRoots, &vNodes, &vLeaves, &vAnds );
        fmt.Printf("Obj = %6d : Leaf = %2d.  Node = %2d.  Root = %2d.    AND = %3d.\n", iLut, Vec_IntSize(vLeaves), Vec_IntSize(vNodes), Vec_IntSize(vRoots), Count); 
    }
    }
    Gia_ManComputeOneWin( pGia, -1, nil, nil, nil, nil );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

