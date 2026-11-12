package abci

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [abcOrder.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Network and node package.]

  Synopsis    [Exploring static BDD variable orders.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: abcOrder.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "base/abc/abc.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static void Abc_NtkChangeCiOrder( Abc_Ntk_t * pNtk, Vec_Ptr_t * vSupp, int fReverse );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////
 
/**Function*************************************************************

  Synopsis    [Changes the order of primary inputs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkFindCiOrder(pNtk *Abc_Ntk_t, fReverse int32, fVerbose int32) {
    var vSupp *Vec_Ptr_t;
    vSupp = Abc_NtkSupport( pNtk );
    Abc_NtkChangeCiOrder( pNtk, vSupp, fReverse );
    Vec_PtrFree( vSupp );
}

/**Function*************************************************************

  Synopsis    [Implements the given variable order.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkImplementCiOrder(pNtk *Abc_Ntk_t, pFileName string, fReverse int32, fVerbose int32) {
    var Buffer [1000]byte;
    var pFile *os.File;
    var vSupp *Vec_Ptr_t;
    var pObj *Abc_Obj_t;
    pFile = os.Open(pFileName);
    vSupp = Vec_PtrAlloc( Abc_NtkCiNum(pNtk) );
    for  fscanf( pFile, "%s", Buffer ) == 1  {
        pObj = Abc_NtkFindCi( pNtk, Buffer );
        if pObj == nil || !Abc_ObjIsCi(pObj) {
            printf( "Name \"%s\" is not a PI name. Cannot use this order.\n", Buffer );
            Vec_PtrFree( vSupp );
            _ = (pFile).Close();
            return;
        }
        Vec_PtrPush( vSupp, pObj );
    }
    _ = (pFile).Close();
    if Vec_PtrSize(vSupp) != Abc_NtkCiNum(pNtk) {
        fmt.Printf("The number of names in the order (%d) is not the same as the number of PIs (%d).\n", Vec_PtrSize(vSupp), Abc_NtkCiNum(pNtk));
        Vec_PtrFree( vSupp );
        return;
    }
    Abc_NtkChangeCiOrder( pNtk, vSupp, fReverse );
    Vec_PtrFree( vSupp );
}
 
/**Function*************************************************************

  Synopsis    [Changes the order of primary inputs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkChangeCiOrder(pNtk *Abc_Ntk_t, vSupp *Vec_Ptr_t, fReverse int32) {
    var pObj *Abc_Obj_t;
    var i int32;
    assert(Vec_PtrSize(vSupp) == Abc_NtkCiNum(pNtk), "Vec_PtrSize(vSupp) == Abc_NtkCiNum(pNtk)");
    // order CIs using the array
    if fReverse != 0 {
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(vSupp)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vSupp, i)), 1); i++  {
            Vec_PtrWriteEntry( pNtk.vCis, Vec_PtrSize(vSupp)-1-i, pObj );
        }
    } else {
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(vSupp)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vSupp, i)), 1); i++  {
            Vec_PtrWriteEntry( pNtk.vCis, i, pObj );
        } }
    // order PIs accordingly
    Vec_PtrClear( pNtk.vPis );
    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(pNtk)) && (((pObj) = Abc_NtkCi(pNtk, i)), 1); i++  {
        if Abc_ObjIsPi(pObj) {
            Vec_PtrPush( pNtk.vPis, pObj ); }
    }
//    Abc_NtkForEachCi( pNtk, pObj, i )
//        printf( "%s ", Abc_ObjName(pObj) );
//    printf( "\n" );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

