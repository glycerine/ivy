package llb

import "fmt"

/**CFile****************************************************************

  FileName    [llb3Image.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [BDD based reachability.]

  Synopsis    [Computes image using partitioned structure.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: llb3Image.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "llbInt.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

   
type Llb_Var_t struct {
    iVar           int32;      // variable number
    nScore           int32;    // variable score
    vParts    *Vec_Int_t;    // partitions
}

   
type Llb_Prt_t struct {
    iPart           int32;     // partition number
    nSize           int32;     // the number of BDD nodes
    bFunc       *DdNode;     // the partition
    vVars    *Vec_Int_t;     // support
}

   
type Llb_Mgr_t struct {
    pAig    *Aig_Man_t;      // AIG manager
    vLeaves    *Vec_Ptr_t;   // leaves in the AIG manager
    vRoots    *Vec_Ptr_t;    // roots in the AIG manager
    dd    *DdManager;        // working BDD manager
    pVars2Q          *int32;   // variables to quantify
    // internal
    pParts   **Llb_Prt_t;    // partitions
    pVars   **Llb_Var_t;     // variables
    iPartFree           int32; // next free partition
    nVars           int32;     // the number of BDD variables
    nSuppMax           int32;  // maximum support size
    // temporary
    pSupp          *int32;     // temporary support storage
}

func Llb_MgrVar(p *Llb_Mgr_t, i int32) *Llb_Var_t { return p.pVars[i];  }
func Llb_MgrPart(p *Llb_Mgr_t, i int32) *Llb_Prt_t { return p.pParts[i]; }

// iterator over vars
// Original macro:
// #define Llb_MgrForEachVar( p, pVar, i )     \
//     for ( i = 0; (i < p->nVars) && (((pVar) = Llb_MgrVar(p, i)), 1); i++ ) if ( pVar == NULL ) {} else

// iterator over parts
// Original macro:
// #define Llb_MgrForEachPart( p, pPart, i )   \
//     for ( i = 0; (i < p->iPartFree) && (((pPart) = Llb_MgrPart(p, i)), 1); i++ ) if ( pPart == NULL ) {} else


// iterator over vars of one partition
// Original macro:
// #define Llb_PartForEachVar( p, pPart, pVar, i )   \
//     for ( i = 0; (i < Vec_IntSize(pPart->vVars)) && (((pVar) = Llb_MgrVar(p, Vec_IntEntry(pPart->vVars,i))), 1); i++ )

// iterator over parts of one variable
// Original macro:
// #define Llb_VarForEachPart( p, pVar, pPart, i )   \
//     for ( i = 0; (i < Vec_IntSize(pVar->vParts)) && (((pPart) = Llb_MgrPart(p, Vec_IntEntry(pVar->vParts,i))), 1); i++ )


// statistics
var timeBuild, timeAndEx, timeOther abctime;
var nSuppMax int32;

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Removes one variable.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_NonlinRemoveVar(p *Llb_Mgr_t, pVar *Llb_Var_t) {
    assert(p.pVars[pVar.iVar] == pVar, "p.pVars[pVar.iVar] == pVar");
    p.pVars[pVar.iVar] = nil;
    Vec_IntFree( pVar.vParts );
    ABC_FREE( pVar );
}

/**Function*************************************************************

  Synopsis    [Removes one partition.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_NonlinRemovePart(p *Llb_Mgr_t, pPart *Llb_Prt_t) {
    assert(p.pParts[pPart.iPart] == pPart, "p.pParts[pPart.iPart] == pPart");
    p.pParts[pPart.iPart] = nil;
    Vec_IntFree( pPart.vVars );
    Cudd_RecursiveDeref( p.dd, pPart.bFunc );
    ABC_FREE( pPart );
}

/**Function*************************************************************

  Synopsis    [Create cube with singleton variables.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_NonlinCreateCube1(p *Llb_Mgr_t, pPart *Llb_Prt_t) *DdNode {
    var bCube, bTemp *DdNode;
    var pVar *Llb_Var_t;
    var i int32;
    var TimeStop abctime;
    TimeStop = p.dd.TimeStop; p.dd.TimeStop = 0;
    bCube = Cudd_ReadOne(p.dd);   Cudd_Ref( bCube );
    // inlining of macro Llb_PartForEachVar
    for  i = 0; (i < Vec_IntSize(pPart.vVars)) && (((pVar) = Llb_MgrVar(p, Vec_IntEntry(pPart.vVars,i))), 1); i++  {
        assert(Vec_IntSize(pVar.vParts) > 0, "Vec_IntSize(pVar.vParts) > 0");
        if Vec_IntSize(pVar.vParts) != 1 {
            continue; }
        assert(Vec_IntEntry(pVar.vParts, 0) == pPart.iPart, "Vec_IntEntry(pVar.vParts, 0) == pPart.iPart");
        bCube = Cudd_bddAnd( p.dd, bTemp = bCube, Cudd_bddIthVar(p.dd, pVar.iVar) );    Cudd_Ref( bCube );
        Cudd_RecursiveDeref( p.dd, bTemp );
    }
    Cudd_Deref( bCube );
    p.dd.TimeStop = TimeStop;
    return bCube;
}

/**Function*************************************************************

  Synopsis    [Create cube of variables appearing only in two partitions.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_NonlinCreateCube2(p *Llb_Mgr_t, pPart1 *Llb_Prt_t, pPart2 *Llb_Prt_t) *DdNode {
    var bCube, bTemp *DdNode;
    var pVar *Llb_Var_t;
    var i int32;
    var TimeStop abctime;
    TimeStop = p.dd.TimeStop; p.dd.TimeStop = 0;
    bCube = Cudd_ReadOne(p.dd);   Cudd_Ref( bCube );
    // inlining of macro Llb_PartForEachVar
    for  i = 0; (i < Vec_IntSize(pPart1.vVars)) && (((pVar) = Llb_MgrVar(p, Vec_IntEntry(pPart1.vVars,i))), 1); i++  {
        assert(Vec_IntSize(pVar.vParts) > 0, "Vec_IntSize(pVar.vParts) > 0");
        if Vec_IntSize(pVar.vParts) != 2 {
            continue; }
        if (Vec_IntEntry(pVar.vParts, 0) == pPart1.iPart && Vec_IntEntry(pVar.vParts, 1) == pPart2.iPart) ||
             (Vec_IntEntry(pVar.vParts, 0) == pPart2.iPart && Vec_IntEntry(pVar.vParts, 1) == pPart1.iPart) {
            bCube = Cudd_bddAnd( p.dd, bTemp = bCube, Cudd_bddIthVar(p.dd, pVar.iVar) );   Cudd_Ref( bCube );
            Cudd_RecursiveDeref( p.dd, bTemp );
        }
    }
    Cudd_Deref( bCube );
    p.dd.TimeStop = TimeStop;
    return bCube;
}

/**Function*************************************************************

  Synopsis    [Returns 1 if partition has singleton variables.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_NonlinHasSingletonVars(p *Llb_Mgr_t, pPart *Llb_Prt_t) int32 {
    var pVar *Llb_Var_t;
    var i int32;
    // inlining of macro Llb_PartForEachVar
    for  i = 0; (i < Vec_IntSize(pPart.vVars)) && (((pVar) = Llb_MgrVar(p, Vec_IntEntry(pPart.vVars,i))), 1); i++  {
        if Vec_IntSize(pVar.vParts) == 1 {
            return 1; }
    }
    return 0;
}

/**Function*************************************************************

  Synopsis    [Returns 1 if partition has singleton variables.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_NonlinPrint(p *Llb_Mgr_t) {
    var pPart *Llb_Prt_t;
    var pVar *Llb_Var_t;
    var i, k int32;
    fmt.Printf("\n");
    // inlining of macro Llb_MgrForEachVar
    for  i = 0; (i < p.nVars) && (((pVar) = Llb_MgrVar(p, i)), 1); i++  { if pVar == nil {} else {
        fmt.Printf("Var %3d : ", i);
        // inlining of macro Llb_VarForEachPart
        for  k = 0; (k < Vec_IntSize(pVar.vParts)) && (((pPart) = Llb_MgrPart(p, Vec_IntEntry(pVar.vParts,k))), 1); k++  {
            fmt.Printf("%d ", pPart.iPart);
        }
        fmt.Printf("\n");
    }
    }
    // inlining of macro Llb_MgrForEachPart
    for  i = 0; (i < p.iPartFree) && (((pPart) = Llb_MgrPart(p, i)), 1); i++  { if pPart == nil {} else {
        fmt.Printf("Part %3d : ", i);
        // inlining of macro Llb_PartForEachVar
        for  k = 0; (k < Vec_IntSize(pPart.vVars)) && (((pVar) = Llb_MgrVar(p, Vec_IntEntry(pPart.vVars,k))), 1); k++  {
            fmt.Printf("%d ", pVar.iVar);
        }
        fmt.Printf("\n");
    }
    }
}

/**Function*************************************************************

  Synopsis    [Quantifies singles belonging to one partition.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_NonlinQuantify1(p *Llb_Mgr_t, pPart *Llb_Prt_t, fSubset int32) int32 {
    var pVar *Llb_Var_t;
    var pTemp *Llb_Prt_t;
    var vSingles *Vec_Ptr_t;
    var bCube, bTemp *DdNode;
    var i, RetValue, nSizeNew int32;
    if fSubset != 0 {        
        var Length int32;
//        int nSuppSize = Cudd_SupportSize( p->dd, pPart->bFunc );
//        pPart->bFunc = Cudd_SubsetHeavyBranch( p->dd, bTemp = pPart->bFunc, nSuppSize, 3*pPart->nSize/4 );  Cudd_Ref( pPart->bFunc );
        pPart.bFunc = Cudd_LargestCube( p.dd, bTemp = pPart.bFunc, &Length );  Cudd_Ref( pPart.bFunc );

        fmt.Printf("Subsetting %3d : ", pPart.iPart);
        fmt.Printf("(Supp =%3d  Node =%5d) -> ", Cudd_SupportSize(p.dd, bTemp), Cudd_DagSize(bTemp));
        fmt.Printf("(Supp =%3d  Node =%5d)\n", Cudd_SupportSize(p.dd, pPart.bFunc), Cudd_DagSize(pPart.bFunc));

        RetValue = (Cudd_DagSize(bTemp) == Cudd_DagSize(pPart.bFunc));

        Cudd_RecursiveDeref( p.dd, bTemp );

        if RetValue {
            return 1; }
    
    } else {
        // create cube to be quantified
        bCube = Llb_NonlinCreateCube1( p, pPart );   Cudd_Ref( bCube );
//        assert( !Cudd_IsConstant(bCube) );
        // derive new function
        pPart.bFunc = Cudd_bddExistAbstract( p.dd, bTemp = pPart.bFunc, bCube );  Cudd_Ref( pPart.bFunc );
        Cudd_RecursiveDeref( p.dd, bTemp );
        Cudd_RecursiveDeref( p.dd, bCube );
    }
    // get support
    vSingles = Vec_PtrAlloc( 0 );
    nSizeNew = Cudd_DagSize(pPart.bFunc);
    Extra_SupportArray( p.dd, pPart.bFunc, p.pSupp );
    // inlining of macro Llb_PartForEachVar
    for  i = 0; (i < Vec_IntSize(pPart.vVars)) && (((pVar) = Llb_MgrVar(p, Vec_IntEntry(pPart.vVars,i))), 1); i++  {
        if p.pSupp[pVar.iVar] {
            assert(Vec_IntSize(pVar.vParts) > 1, "Vec_IntSize(pVar.vParts) > 1");
            pVar.nScore -= pPart.nSize - nSizeNew;
        
        } else {
            RetValue = Vec_IntRemove( pVar.vParts, pPart.iPart );
            assert(RetValue, "RetValue");
            pVar.nScore -= pPart.nSize;
            if Vec_IntSize(pVar.vParts) == 0 {
                Llb_NonlinRemoveVar( p, pVar );
            } else if Vec_IntSize(pVar.vParts) == 1 {
                Vec_PtrPushUnique( vSingles, Llb_MgrPart(p, Vec_IntEntry(pVar.vParts,0)) ); }
        }
    }

    // update partition
    pPart.nSize = nSizeNew;
    Vec_IntClear( pPart.vVars );
    for  i = 0; i < p.nVars; i++  {
        if p.pSupp[i] && p.pVars2Q[i] {
            Vec_IntPush( pPart.vVars, i ); }
    }
    // remove other variables
    Vec_PtrForEachEntry( Llb_Prt_t *, vSingles, pTemp, i )
        Llb_NonlinQuantify1( p, pTemp, 0 );
    Vec_PtrFree( vSingles );
    return 0;
}

/**Function*************************************************************

  Synopsis    [Quantifies singles belonging to one partition.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_NonlinQuantify2(p *Llb_Mgr_t, pPart1 *Llb_Prt_t, pPart2 *Llb_Prt_t) int32 {
    var fVerbose int32;
    var pVar *Llb_Var_t;
    var pTemp *Llb_Prt_t;
    var vSingles *Vec_Ptr_t;
    var bCube, bFunc *DdNode;
    var i, RetValue, nSuppSize int32;
//    int iPart1 = pPart1->iPart;
//    int iPart2 = pPart2->iPart;

    // create cube to be quantified
    bCube = Llb_NonlinCreateCube2( p, pPart1, pPart2 );   Cudd_Ref( bCube );
if fVerbose != 0 {
fmt.Printf("\n");
fmt.Printf("\n");
Llb_NonlinPrint( p );
fmt.Printf("Conjoining partitions %d and %d.\n", pPart1.iPart, pPart2.iPart);
Extra_bddPrintSupport( p.dd, bCube );  fmt.Printf("\n");
}
 
    // derive new function
//    bFunc = Cudd_bddAndAbstract( p->dd, pPart1->bFunc, pPart2->bFunc, bCube );  Cudd_Ref( bFunc );
/*
    bFunc = Cudd_bddAndAbstractLimit( p->dd, pPart1->bFunc, pPart2->bFunc, bCube, Limit );  
    if ( bFunc == NULL )
    {
        int RetValue;
        Cudd_RecursiveDeref( p->dd, bCube );
        if ( pPart1->nSize < pPart2->nSize )
            RetValue = Llb_NonlinQuantify1( p, pPart1, 1 );
        else
            RetValue = Llb_NonlinQuantify1( p, pPart2, 1 );
        if ( RetValue )
            Limit = Limit + 1000;
        Llb_NonlinQuantify2( p, pPart1, pPart2 );
        return 0;
    }
    Cudd_Ref( bFunc );
*/

//    bFunc = Extra_bddAndAbstractTime( p->dd, pPart1->bFunc, pPart2->bFunc, bCube, TimeOut );  
    bFunc = Cudd_bddAndAbstract( p.dd, pPart1.bFunc, pPart2.bFunc, bCube );  
    if bFunc == nil {
        Cudd_RecursiveDeref( p.dd, bCube );
        return 0;
    }
    Cudd_Ref( bFunc );
    Cudd_RecursiveDeref( p.dd, bCube );

    // create new partition
    pTemp = p.pParts[p.iPartFree] = new(Llb_Prt_t);
    pTemp.iPart = p.iPartFree++;
    pTemp.nSize = Cudd_DagSize(bFunc);
    pTemp.bFunc = bFunc;
    pTemp.vVars = Vec_IntAlloc( 8 );
    // update variables
    // inlining of macro Llb_PartForEachVar
    for  i = 0; (i < Vec_IntSize(pPart1.vVars)) && (((pVar) = Llb_MgrVar(p, Vec_IntEntry(pPart1.vVars,i))), 1); i++  {
        RetValue = Vec_IntRemove( pVar.vParts, pPart1.iPart );
        assert(RetValue, "RetValue");
        pVar.nScore -= pPart1.nSize;
    }
    // update variables
    // inlining of macro Llb_PartForEachVar
    for  i = 0; (i < Vec_IntSize(pPart2.vVars)) && (((pVar) = Llb_MgrVar(p, Vec_IntEntry(pPart2.vVars,i))), 1); i++  {
        RetValue = Vec_IntRemove( pVar.vParts, pPart2.iPart );
        assert(RetValue, "RetValue");
        pVar.nScore -= pPart2.nSize;
    }
    // add variables to the new partition
    nSuppSize = 0;
    Extra_SupportArray( p.dd, bFunc, p.pSupp );
    for  i = 0; i < p.nVars; i++  {
        nSuppSize += p.pSupp[i];
        if p.pSupp[i] && p.pVars2Q[i] {
            pVar = Llb_MgrVar( p, i );
            pVar.nScore += pTemp.nSize;
            Vec_IntPush( pVar.vParts, pTemp.iPart );
            Vec_IntPush( pTemp.vVars, i );
        }
    }
    p.nSuppMax = Abc_MaxInt( p.nSuppMax, nSuppSize ); 
    // remove variables and collect partitions with singleton variables
    vSingles = Vec_PtrAlloc( 0 );
    // inlining of macro Llb_PartForEachVar
    for  i = 0; (i < Vec_IntSize(pPart1.vVars)) && (((pVar) = Llb_MgrVar(p, Vec_IntEntry(pPart1.vVars,i))), 1); i++  {
        if Vec_IntSize(pVar.vParts) == 0 {
            Llb_NonlinRemoveVar( p, pVar );
        } else if Vec_IntSize(pVar.vParts) == 1 {
            if fVerbose != 0 {
                fmt.Printf("Adding partition %d because of var %d.\n", Llb_MgrPart(p, Vec_IntEntry(pVar.vParts,0)).iPart, pVar.iVar); }
            Vec_PtrPushUnique( vSingles, Llb_MgrPart(p, Vec_IntEntry(pVar.vParts,0)) );
        }
    }
    // inlining of macro Llb_PartForEachVar
    for  i = 0; (i < Vec_IntSize(pPart2.vVars)) && (((pVar) = Llb_MgrVar(p, Vec_IntEntry(pPart2.vVars,i))), 1); i++  {
        if pVar == nil {
            continue; }
        if Vec_IntSize(pVar.vParts) == 0 {
            Llb_NonlinRemoveVar( p, pVar );
        } else if Vec_IntSize(pVar.vParts) == 1 {
            if fVerbose != 0 {
                fmt.Printf("Adding partition %d because of var %d.\n", Llb_MgrPart(p, Vec_IntEntry(pVar.vParts,0)).iPart, pVar.iVar); }
            Vec_PtrPushUnique( vSingles, Llb_MgrPart(p, Vec_IntEntry(pVar.vParts,0)) );
        }
    }
    // remove partitions
    Llb_NonlinRemovePart( p, pPart1 );
    Llb_NonlinRemovePart( p, pPart2 );
    // remove other variables
if fVerbose != 0 {
Llb_NonlinPrint( p ); }
    Vec_PtrForEachEntry( Llb_Prt_t *, vSingles, pTemp, i )
    {
if fVerbose != 0 {
fmt.Printf("Updating partitiong %d with singlton vars.\n", pTemp.iPart); }
        Llb_NonlinQuantify1( p, pTemp, 0 );
    }
if fVerbose != 0 {
Llb_NonlinPrint( p ); }
    Vec_PtrFree( vSingles );
    return 1;
}

/**Function*************************************************************

  Synopsis    [Computes volume of the cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_NonlinCutNodes_rec(p *Aig_Man_t, pObj *Aig_Obj_t, vNodes *Vec_Ptr_t) {
    if Aig_ObjIsTravIdCurrent(p, pObj) {
        return; }
    Aig_ObjSetTravIdCurrent(p, pObj);
    if Saig_ObjIsLi(p, pObj) {
        Llb_NonlinCutNodes_rec(p, Aig_ObjFanin0(pObj), vNodes);
        return;
    }
    if Aig_ObjIsConst1(pObj) {
        return; }
    assert(Aig_ObjIsNode(pObj), "Aig_ObjIsNode(pObj)");
    Llb_NonlinCutNodes_rec(p, Aig_ObjFanin0(pObj), vNodes);
    Llb_NonlinCutNodes_rec(p, Aig_ObjFanin1(pObj), vNodes);
    Vec_PtrPush( vNodes, pObj );
}

/**Function*************************************************************

  Synopsis    [Computes volume of the cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_NonlinCutNodes(p *Aig_Man_t, vLower *Vec_Ptr_t, vUpper *Vec_Ptr_t) *Vec_Ptr_t {
    var vNodes *Vec_Ptr_t;
    var pObj *Aig_Obj_t;
    var i int32;
    // mark the lower cut with the traversal ID
    Aig_ManIncrementTravId(p);
    Vec_PtrForEachEntry( Aig_Obj_t *, vLower, pObj, i )
        Aig_ObjSetTravIdCurrent( p, pObj );
    // count the upper cut
    vNodes = Vec_PtrAlloc( 100 );
    Vec_PtrForEachEntry( Aig_Obj_t *, vUpper, pObj, i )
        Llb_NonlinCutNodes_rec( p, pObj, vNodes );
    return vNodes;
}

/**Function*************************************************************

  Synopsis    [Returns array of BDDs for the roots in terms of the leaves.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_NonlinBuildBdds(p *Aig_Man_t, vLower *Vec_Ptr_t, vUpper *Vec_Ptr_t, dd *DdManager) *Vec_Ptr_t {
    var vNodes, vResult *Vec_Ptr_t;
    var pObj *Aig_Obj_t;
    var bBdd0, bBdd1, bProd *DdNode;
    var i, k int32;

    Aig_ManConst1(p).pData = Cudd_ReadOne( dd );
    Vec_PtrForEachEntry( Aig_Obj_t *, vLower, pObj, i )
        pObj.pData = Cudd_bddIthVar( dd, Aig_ObjId(pObj) );

    vNodes = Llb_NonlinCutNodes( p, vLower, vUpper );
    Vec_PtrForEachEntry( Aig_Obj_t *, vNodes, pObj, i )
    {
        // inlining of macro Cudd_NotCond
        bBdd0 = ((DdNode *)((ptrint)((DdNode *)Aig_ObjFanin0(pObj).pData) ^ (Aig_ObjFaninC0(pObj))));
        // inlining of macro Cudd_NotCond
        bBdd1 = ((DdNode *)((ptrint)((DdNode *)Aig_ObjFanin1(pObj).pData) ^ (Aig_ObjFaninC1(pObj))));
//        pObj->pData = Extra_bddAndTime( dd, bBdd0, bBdd1, TimeOut );
        pObj.pData = Cudd_bddAnd( dd, bBdd0, bBdd1 );
        if pObj.pData == nil {
            Vec_PtrForEachEntryStop( Aig_Obj_t *, vNodes, pObj, k, i )
                if ( pObj.pData )
                    Cudd_RecursiveDeref( dd, (DdNode *)pObj.pData );
            Vec_PtrFree( vNodes );
            return nil;
        }
        Cudd_Ref( (DdNode *)pObj.pData );
    }

    vResult = Vec_PtrAlloc( 100 );
    Vec_PtrForEachEntry( Aig_Obj_t *, vUpper, pObj, i )
    {
        if Aig_ObjIsNode(pObj) {
            bProd = Cudd_bddXnor( dd, Cudd_bddIthVar(dd, Aig_ObjId(pObj)), (DdNode *)pObj.pData );  Cudd_Ref( bProd );
        
        } else {
            assert(Saig_ObjIsLi(p, pObj), "Saig_ObjIsLi(p, pObj)");
            // inlining of macro Cudd_NotCond
            bBdd0 = ((DdNode *)((ptrint)((DdNode *)Aig_ObjFanin0(pObj).pData) ^ (Aig_ObjFaninC0(pObj))));
            bProd = Cudd_bddXnor( dd, Cudd_bddIthVar(dd, Aig_ObjId(pObj)), bBdd0 );                  Cudd_Ref( bProd );
        }
        Vec_PtrPush( vResult, bProd );
    }
    Vec_PtrForEachEntry( Aig_Obj_t *, vNodes, pObj, i )
        Cudd_RecursiveDeref( dd, (DdNode *)pObj.pData );

    Vec_PtrFree( vNodes );
    return vResult;
}

/**Function*************************************************************

  Synopsis    [Starts non-linear quantification scheduling.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_NonlinAddPair(p *Llb_Mgr_t, bFunc *DdNode, iPart int32, iVar int32) {
    if p.pVars[iVar] == nil {
        p.pVars[iVar] = new(Llb_Var_t);
        p.pVars[iVar].iVar   = iVar;
        p.pVars[iVar].nScore = 0;
        p.pVars[iVar].vParts = Vec_IntAlloc( 8 );
    }
    Vec_IntPush( p.pVars[iVar].vParts, iPart );
    Vec_IntPush( p.pParts[iPart].vVars, iVar );
}

/**Function*************************************************************

  Synopsis    [Starts non-linear quantification scheduling.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_NonlinAddPartition(p *Llb_Mgr_t, i int32, bFunc *DdNode) {
    var k, nSuppSize int32;
    // inlining of macro Cudd_IsConstant
    // inlining of macro Cudd_Regular
    assert(!((((DdNode *)((ptruint)(bFunc) & ~01))).index == CUDD_CONST_INDEX), "!/*__starting_point_inline_macro:Cudd_IsConstant__*/((/*__starting_point_inline_macro:Cudd_Regular__*/((DdNode *)((ptruint)(bFunc) & ~01))).index == CUDD_CONST_INDEX)");
    // create partition
    p.pParts[i] = new(Llb_Prt_t);
    p.pParts[i].iPart = i;
    p.pParts[i].bFunc = bFunc;
    p.pParts[i].vVars = Vec_IntAlloc( 8 );
    // add support dependencies
    nSuppSize = 0;
    Extra_SupportArray( p.dd, bFunc, p.pSupp );
    for  k = 0; k < p.nVars; k++  {
        nSuppSize += p.pSupp[k];
        if p.pSupp[k] && p.pVars2Q[k] {
            Llb_NonlinAddPair( p, bFunc, i, k ); }
    }
    p.nSuppMax = Abc_MaxInt( p.nSuppMax, nSuppSize ); 
}

/**Function*************************************************************

  Synopsis    [Starts non-linear quantification scheduling.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_NonlinStart(p *Llb_Mgr_t) int32 {
    var vRootBdds *Vec_Ptr_t;
    var bFunc *DdNode;
    var i int32;
    // create and collect BDDs
    vRootBdds = Llb_NonlinBuildBdds( p.pAig, p.vLeaves, p.vRoots, p.dd ); // come referenced
    if vRootBdds == nil {
        return 0; }
    // add pairs (refs are consumed inside)
    Vec_PtrForEachEntry( DdNode *, vRootBdds, bFunc, i )
        Llb_NonlinAddPartition( p, i, bFunc );
    Vec_PtrFree( vRootBdds );
    return 1;
}

/**Function*************************************************************

  Synopsis    [Checks that each var appears in at least one partition.]

  Description []
               
  SideEffects []

  SeeAlso     []
**********************************************************************/
func Llb_NonlinCheckVars(p *Llb_Mgr_t) {
    var pVar *Llb_Var_t;
    var i int32;
    // inlining of macro Llb_MgrForEachVar
    for  i = 0; (i < p.nVars) && (((pVar) = Llb_MgrVar(p, i)), 1); i++  { if pVar == nil {} else {
        assert(Vec_IntSize(pVar.vParts) > 1, "Vec_IntSize(pVar.vParts) > 1"); }
    }
}

/**Function*************************************************************

  Synopsis    [Find next partition to quantify]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_NonlinNextPartitions(p *Llb_Mgr_t, ppPart1 **Llb_Prt_t, ppPart2 **Llb_Prt_t) int32 {
    var pVar *Llb_Var_t;
    var pVarBest *Llb_Var_t = nil;
    var pPart *Llb_Prt_t;
    var pPart1Best *Llb_Prt_t = nil;
    var pPart2Best *Llb_Prt_t = nil;
    var i int32;
    Llb_NonlinCheckVars( p );
    // find variable with minimum score
    // inlining of macro Llb_MgrForEachVar
    for  i = 0; (i < p.nVars) && (((pVar) = Llb_MgrVar(p, i)), 1); i++  { if pVar == nil {} else
        if pVarBest == nil || pVarBest.nScore > pVar.nScore {
            pVarBest = pVar; }
    }
    if pVarBest == nil {
        return 0; }
    // find two partitions with minimum size
    // inlining of macro Llb_VarForEachPart
    for  i = 0; (i < Vec_IntSize(pVarBest.vParts)) && (((pPart) = Llb_MgrPart(p, Vec_IntEntry(pVarBest.vParts,i))), 1); i++  {
        if pPart1Best == nil {
            pPart1Best = pPart;
        } else if pPart2Best == nil {
            pPart2Best = pPart;
        } else if pPart1Best.nSize > pPart.nSize || pPart2Best.nSize > pPart.nSize {
            if pPart1Best.nSize > pPart2Best.nSize {
                pPart1Best = pPart;
            } else {
                pPart2Best = pPart; }
        }
    }
    *ppPart1 = pPart1Best;
    *ppPart2 = pPart2Best;
    return 1; 
}

/**Function*************************************************************

  Synopsis    [Reorders BDDs in the working manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_NonlinReorder(dd *DdManager, fTwice int32, fVerbose int32) {
    var clk abctime  = Abc_Clock();
    if fVerbose != 0 {
        Abc_Print( 1, "Reordering... Before =%5d. ", Cudd_ReadKeys(dd) - Cudd_ReadDead(dd) ); }
    Cudd_ReduceHeap( dd, CUDD_REORDER_SYMM_SIFT, 100 );
    if fVerbose != 0 {
        Abc_Print( 1, "After =%5d. ", Cudd_ReadKeys(dd) - Cudd_ReadDead(dd) ); }
    if fTwice != 0 {
        Cudd_ReduceHeap( dd, CUDD_REORDER_SYMM_SIFT, 100 );
        if fVerbose != 0 {
            Abc_Print( 1, "After =%5d. ", Cudd_ReadKeys(dd) - Cudd_ReadDead(dd) ); }
    }
    if fVerbose != 0 {
        Abc_PrintTime( 1, "Time", Abc_Clock() - clk ); }
}

/**Function*************************************************************

  Synopsis    [Recomputes scores after variable reordering.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_NonlinRecomputeScores(p *Llb_Mgr_t) {
    var pPart *Llb_Prt_t;
    var pVar *Llb_Var_t;
    var i, k int32;
    // inlining of macro Llb_MgrForEachPart
    for  i = 0; (i < p.iPartFree) && (((pPart) = Llb_MgrPart(p, i)), 1); i++  { if pPart == nil {} else {
        pPart.nSize = Cudd_DagSize(pPart.bFunc); }
    }
    // inlining of macro Llb_MgrForEachVar
    for  i = 0; (i < p.nVars) && (((pVar) = Llb_MgrVar(p, i)), 1); i++  { if pVar == nil {} else {
        pVar.nScore = 0;
        // inlining of macro Llb_VarForEachPart
        for  k = 0; (k < Vec_IntSize(pVar.vParts)) && (((pPart) = Llb_MgrPart(p, Vec_IntEntry(pVar.vParts,k))), 1); k++  {
            pVar.nScore += pPart.nSize;
        }
    }
    }
}

/**Function*************************************************************

  Synopsis    [Recomputes scores after variable reordering.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_NonlinVerifyScores(p *Llb_Mgr_t) {
    var pPart *Llb_Prt_t;
    var pVar *Llb_Var_t;
    var i, k, nScore int32;
    // inlining of macro Llb_MgrForEachPart
    for  i = 0; (i < p.iPartFree) && (((pPart) = Llb_MgrPart(p, i)), 1); i++  { if pPart == nil {} else {
        assert(pPart.nSize == Cudd_DagSize(pPart.bFunc), "pPart.nSize == Cudd_DagSize(pPart.bFunc)"); }
    }
    // inlining of macro Llb_MgrForEachVar
    for  i = 0; (i < p.nVars) && (((pVar) = Llb_MgrVar(p, i)), 1); i++  { if pVar == nil {} else {
        nScore = 0;
        // inlining of macro Llb_VarForEachPart
        for  k = 0; (k < Vec_IntSize(pVar.vParts)) && (((pPart) = Llb_MgrPart(p, Vec_IntEntry(pVar.vParts,k))), 1); k++  {
            nScore += pPart.nSize;
        }
        assert(nScore == pVar.nScore, "nScore == pVar.nScore");
    }
    }
}

/**Function*************************************************************

  Synopsis    [Starts non-linear quantification scheduling.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_NonlinAlloc(pAig *Aig_Man_t, vLeaves *Vec_Ptr_t, vRoots *Vec_Ptr_t, pVars2Q *int32, dd *DdManager) *Llb_Mgr_t {
    var p *Llb_Mgr_t;
    p = new(Llb_Mgr_t);
    p.pAig      = pAig;
    p.vLeaves   = vLeaves;
    p.vRoots    = vRoots;
    p.dd        = dd;
    p.pVars2Q   = pVars2Q;
    p.nVars     = Cudd_ReadSize(dd);
    p.iPartFree = Vec_PtrSize(vRoots);
    p.pVars     = ABC_CALLOC( Llb_Var_t *, p.nVars );
    p.pParts    = ABC_CALLOC( Llb_Prt_t *, 2 * p.iPartFree + 2 );
    p.pSupp     = ABC_ALLOC( int, Cudd_ReadSize(dd) );
    return p;
}

/**Function*************************************************************

  Synopsis    [Stops non-linear quantification scheduling.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_NonlinFree(p *Llb_Mgr_t) {
    var pPart *Llb_Prt_t;
    var pVar *Llb_Var_t;
    var i int32;
    // inlining of macro Llb_MgrForEachVar
    for  i = 0; (i < p.nVars) && (((pVar) = Llb_MgrVar(p, i)), 1); i++  { if pVar == nil {} else {
        Llb_NonlinRemoveVar( p, pVar ); }
    }
    // inlining of macro Llb_MgrForEachPart
    for  i = 0; (i < p.iPartFree) && (((pPart) = Llb_MgrPart(p, i)), 1); i++  { if pPart == nil {} else {
        Llb_NonlinRemovePart( p, pPart ); }
    }
    ABC_FREE( p.pVars );
    ABC_FREE( p.pParts );
    ABC_FREE( p.pSupp );
    ABC_FREE( p );
}

/**Function*************************************************************

  Synopsis    [Performs image computation.]

  Description [Computes image of BDDs (vFuncs).]
               
  SideEffects [BDDs in vFuncs are derefed inside. The result is refed.]

  SeeAlso     []

***********************************************************************/
func Llb_NonlinImage(pAig *Aig_Man_t, vLeaves *Vec_Ptr_t, vRoots *Vec_Ptr_t, pVars2Q *int32, dd *DdManager, bCurrent *DdNode, fReorder int32, fVerbose int32, pOrder *int32) *DdNode {
    var pPart, pPart1, pPart2 *Llb_Prt_t;
    var p *Llb_Mgr_t;
    var bFunc, bTemp *DdNode;
    var i, nReorders, timeInside int32;
    var clk abctime = Abc_Clock();
    var clk2 abctime;
    // start the manager
    clk2 = Abc_Clock();
    p = Llb_NonlinAlloc( pAig, vLeaves, vRoots, pVars2Q, dd );
    if !(Llb_NonlinStart( p ) != 0) {
        Llb_NonlinFree( p );
        return nil;
    }
    // add partition
    Llb_NonlinAddPartition( p, p.iPartFree++, bCurrent );
    // remove singles
    // inlining of macro Llb_MgrForEachPart
    for  i = 0; (i < p.iPartFree) && (((pPart) = Llb_MgrPart(p, i)), 1); i++  { if pPart == nil {} else
        if Llb_NonlinHasSingletonVars(p, pPart) != 0 {
            Llb_NonlinQuantify1( p, pPart, 0 ); }
    }
    timeBuild += Abc_Clock() - clk2;
    timeInside = Abc_Clock() - clk2;
    // compute scores
    Llb_NonlinRecomputeScores( p );
    // save permutation
    if pOrder != nil {
    memcpy( pOrder, dd.invperm, sizeof(int) * dd.size ); }
    // iteratively quantify variables
    for  Llb_NonlinNextPartitions(p, &pPart1, &pPart2) != 0  {
        clk2 = Abc_Clock();
        nReorders = Cudd_ReadReorderings(dd);
        if !(Llb_NonlinQuantify2( p, pPart1, pPart2 ) != 0) {
            Llb_NonlinFree( p );
            return nil;
        }
        timeAndEx  += Abc_Clock() - clk2;
        timeInside += Abc_Clock() - clk2;
        if nReorders < Cudd_ReadReorderings(dd) {
            Llb_NonlinRecomputeScores( p ); }
//        else
//            Llb_NonlinVerifyScores( p );
    }
    // load partitions
    bFunc = Cudd_ReadOne(p.dd);   Cudd_Ref( bFunc );
    // inlining of macro Llb_MgrForEachPart
    for  i = 0; (i < p.iPartFree) && (((pPart) = Llb_MgrPart(p, i)), 1); i++  { if pPart == nil {} else {
        bFunc = Cudd_bddAnd( p.dd, bTemp = bFunc, pPart.bFunc );   Cudd_Ref( bFunc );
        Cudd_RecursiveDeref( p.dd, bTemp );
    }
    }
    nSuppMax = p.nSuppMax;
    Llb_NonlinFree( p );
    // reorder variables
    if fReorder != 0 {
        Llb_NonlinReorder( dd, 0, fVerbose ); }
    timeOther += Abc_Clock() - clk - timeInside;
    // return
    Cudd_Deref( bFunc );
    return bFunc;
}



var p *Llb_Mgr_t = nil;

/**Function*************************************************************

  Synopsis    [Starts image computation manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_NonlinImageStart(pAig *Aig_Man_t, vLeaves *Vec_Ptr_t, vRoots *Vec_Ptr_t, pVars2Q *int32, pOrder *int32, fFirst int32, TimeTarget abctime) *DdManager {
    var dd *DdManager;
    var clk abctime  = Abc_Clock();
    assert(p == nil, "p == nil");
    // start a new manager (disable reordering)
    dd = Cudd_Init( Aig_ManObjNumMax(pAig), 0, CUDD_UNIQUE_SLOTS, CUDD_CACHE_SLOTS, 0 );
    dd.TimeStop = TimeTarget;
    Cudd_ShuffleHeap( dd, pOrder );
//    if ( fFirst )
        Cudd_AutodynEnable( dd,  CUDD_REORDER_SYMM_SIFT );
    // start the manager
    p = Llb_NonlinAlloc( pAig, vLeaves, vRoots, pVars2Q, dd );
    if !(Llb_NonlinStart( p ) != 0) {
        Llb_NonlinFree( p );
        p = nil;
        return nil;
    }
    timeBuild += Abc_Clock() - clk;
//    if ( !fFirst )
//        Cudd_AutodynEnable( dd,  CUDD_REORDER_SYMM_SIFT );
    return dd;
}

/**Function*************************************************************

  Synopsis    [Performs image computation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_NonlinImageCompute(bCurrent *DdNode, fReorder int32, fDrop int32, fVerbose int32, pOrder *int32) *DdNode {
    var pPart, pPart1, pPart2 *Llb_Prt_t;
    var bFunc, bTemp *DdNode;
    var i, nReorders, timeInside int32;
    var clk abctime = Abc_Clock();
    var clk2 abctime;

    // add partition
    Llb_NonlinAddPartition( p, p.iPartFree++, bCurrent );
    // remove singles
    // inlining of macro Llb_MgrForEachPart
    for  i = 0; (i < p.iPartFree) && (((pPart) = Llb_MgrPart(p, i)), 1); i++  { if pPart == nil {} else
        if Llb_NonlinHasSingletonVars(p, pPart) != 0 {
            Llb_NonlinQuantify1( p, pPart, 0 ); }
    }
    // reorder
    if fReorder != 0 {
        Llb_NonlinReorder( p.dd, 0, 0 ); }
    // save permutation
    memcpy( pOrder, p.dd.invperm, sizeof(int) * p.dd.size );

    // compute scores
    Llb_NonlinRecomputeScores( p );
    // iteratively quantify variables
    for  Llb_NonlinNextPartitions(p, &pPart1, &pPart2) != 0  {
        clk2 = Abc_Clock();
        nReorders = Cudd_ReadReorderings(p.dd);
        if !(Llb_NonlinQuantify2( p, pPart1, pPart2 ) != 0) {
            Llb_NonlinFree( p );
            return nil;
        }
        timeAndEx  += Abc_Clock() - clk2;
        timeInside += Abc_Clock() - clk2;
        if nReorders < Cudd_ReadReorderings(p.dd) {
            Llb_NonlinRecomputeScores( p ); }
//        else
//            Llb_NonlinVerifyScores( p );
    }
    // load partitions
    bFunc = Cudd_ReadOne(p.dd);   Cudd_Ref( bFunc );
    // inlining of macro Llb_MgrForEachPart
    for  i = 0; (i < p.iPartFree) && (((pPart) = Llb_MgrPart(p, i)), 1); i++  { if pPart == nil {} else {
        bFunc = Cudd_bddAnd( p.dd, bTemp = bFunc, pPart.bFunc );
        if bFunc == nil {
            Cudd_RecursiveDeref( p.dd, bTemp );
            Llb_NonlinFree( p );
            return nil;
        }
        Cudd_Ref( bFunc );
        Cudd_RecursiveDeref( p.dd, bTemp );
    }
    }
    nSuppMax = p.nSuppMax;
    // reorder variables
//    if ( fReorder )
//        Llb_NonlinReorder( p->dd, 0, fVerbose );
    // save permutation
//    memcpy( pOrder, p->dd->invperm, sizeof(int) * Cudd_ReadSize(p->dd) );

    timeOther += Abc_Clock() - clk - timeInside;
    // return
    Cudd_Deref( bFunc );
    return bFunc;
}

/**Function*************************************************************

  Synopsis    [Quits image computation manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_NonlinImageQuit() {
    var dd *DdManager;
    if p == nil {
        return; }
    dd = p.dd;
    Llb_NonlinFree( p );
    if dd.bFunc {
        Cudd_RecursiveDeref( dd, dd.bFunc ); }
    Extra_StopManager( dd );
//    Cudd_Quit ( dd );
    p = nil;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

