package gia

import "fmt"

/**CFile****************************************************************

  FileName    [giaStr.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [AIG structuring.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaStr.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"
//#include "misc/util/utilNam.h"
//#include "misc/vec/vecWec.h"
//#include "misc/tim/tim.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// Original macro:
// #define Str_NtkManForEachObj( p, pObj )  \
//     for ( pObj = p->pObjs; Str_ObjId(p, pObj) < p->nObjs; pObj++ )

// Original macro:
// #define Str_NtkManForEachObjVec( vVec, p, pObj, i )  \
//     for ( i = 0; (i < Vec_IntSize(vVec)) && ((pObj) = Str_NtkObj(p, Vec_IntEntry(vVec,i))); i++ )


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Logic network manipulation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/









/**Function*************************************************************

  Synopsis    [Constructs a normalized AIG without structural hashing.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupMuxesNoHash(p *Gia_Man_t) *Gia_Man_t {
    var pNew *Gia_Man_t;
    var pObj, pFan0, pFan1, pFanC *Gia_Obj_t;
    var i, iLit0, iLit1, fCompl int32;
    assert(p.pMuxes == nil, "p.pMuxes == nil");
    // inlining of macro ABC_FREE
    p.Refs = nil;
    Gia_ManCreateRefs( p ); 
    // discount nodes with one fanout pointed to by MUX type
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        if !Gia_ObjIsMuxType(pObj) {
            continue; }
        Gia_ObjRefDec(p, Gia_ObjFanin0(pObj));
        Gia_ObjRefDec(p, Gia_ObjFanin1(pObj));
    }
    }
    // start the new manager
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName  = Abc_UtilStrsav( p.pName );
    pNew.pSpec  = Abc_UtilStrsav( p.pSpec );
    // inlining of macro ABC_CALLOC
    pNew.pMuxes = ((unsigned *) calloc((size_t)(pNew.nObjsAlloc), sizeof(unsigned)));
    Gia_ManFillValue(p);
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObj.Value = Gia_ManAppendCi( pNew );
    }
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        if !Gia_ObjRefNumId(p, i) {
            continue; }
        if !Gia_ObjIsMuxType(pObj) {
            pObj.Value = Gia_ManAppendAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
        } else if Gia_ObjRecognizeExor(pObj, &pFan0, &pFan1) {
            iLit0 = Gia_ObjLitCopy(p, Gia_ObjToLit(p, pFan0));
            iLit1 = Gia_ObjLitCopy(p, Gia_ObjToLit(p, pFan1));
            fCompl = Abc_LitIsCompl(iLit0) ^ Abc_LitIsCompl(iLit1);
            pObj.Value = fCompl ^ Gia_ManAppendXorReal( pNew, Abc_LitRegular(iLit0), Abc_LitRegular(iLit1) );
        
        } else {
            pFanC = Gia_ObjRecognizeMux( pObj, &pFan1, &pFan0 );
            iLit0 = Gia_ObjLitCopy(p, Gia_ObjToLit(p, pFan0));
            iLit1 = Gia_ObjLitCopy(p, Gia_ObjToLit(p, pFan1));
            if iLit0 == iLit1 {
                pObj.Value = iLit0;
            } else if Abc_Lit2Var(iLit0) == Abc_Lit2Var(iLit1) {
                iLit1 = Gia_ObjLitCopy(p, Gia_ObjToLit(p, pFanC));
                fCompl = Abc_LitIsCompl(iLit0) ^ Abc_LitIsCompl(iLit1);
                pObj.Value = fCompl ^ Gia_ManAppendXorReal( pNew, Abc_LitRegular(iLit0), Abc_LitRegular(iLit1) );
            
            } else {
                pObj.Value = Gia_ManAppendMuxReal( pNew, Gia_ObjLitCopy(p, Gia_ObjToLit(p, pFanC)), Gia_ObjLitCopy(p, Gia_ObjToLit(p, pFan1)), Gia_ObjLitCopy(p, Gia_ObjToLit(p, pFan0)) ); }
        }
    }
    }
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
    }
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(p) );
    assert(!Gia_ManHasDangling(pNew), "!Gia_ManHasDangling(pNew)");
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Constructs AIG ordered for balancing.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/













/**Function*************************************************************

  Synopsis    [Simplify multi-input AND/XOR.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/



/**Function*************************************************************

  Synopsis    [Collect multi-input AND/XOR.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/




/**Function*************************************************************

  Synopsis    [Constructs AIG ordered for balancing.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/



 
/**Function*************************************************************

  Synopsis    [Delay computation]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/






/**Function*************************************************************

  Synopsis    [Transposing 64-bit matrix.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Perform affinity computation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/





/**Function*************************************************************

  Synopsis    [Count 1s.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


func Str_PrintState(pCost *int32, pSuper *int32, pMatrix *word, nSize int32) {
    var i int32;
    for  i = 0; i < nSize; i++  {
        fmt.Printf("%6d : ", i);
        fmt.Printf("%6d : ", Abc_Lit2Var(pSuper[i]));
        fmt.Printf("%3d   ", pCost[i] >> 4);
        fmt.Printf("%3d   ", pCost[i] & 15);
//        Extra_PrintBinary( stdout, pMatrix+i, 64 ), printf( "\n" );
    }
    fmt.Printf("\n");
}


/**Function*************************************************************

  Synopsis    [Perform balancing.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/








/**Function*************************************************************

  Synopsis    [Test normalization procedure.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/






/**Function*************************************************************

  Synopsis    [Perform MUX restructuring.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
    











// return node whose both branches are non-trivial







/**Function*************************************************************

  Synopsis    [Perform MUX restructuring for area.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/




////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

