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
    dd    *DdManager;        // working BDD manager
    vVars2Q    *Vec_Int_t;   // variables to quantify
    nSizeMax           int32;  // maximum size of the cluster
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
//abctime timeBuild, timeAndEx, timeOther;
//int nSuppMax;

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Removes one variable.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_Nonlin4RemoveVar(p *Llb_Mgr_t, pVar *Llb_Var_t) {
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
func Llb_Nonlin4RemovePart(p *Llb_Mgr_t, pPart *Llb_Prt_t) {
//printf( "Removing %d\n", pPart->iPart );
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
func Llb_Nonlin4CreateCube1(p *Llb_Mgr_t, pPart *Llb_Prt_t) *DdNode {
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
func Llb_Nonlin4CreateCube2(p *Llb_Mgr_t, pPart1 *Llb_Prt_t, pPart2 *Llb_Prt_t) *DdNode {
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
func Llb_Nonlin4HasSingletonVars(p *Llb_Mgr_t, pPart *Llb_Prt_t) int32 {
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
func Llb_Nonlin4Print(p *Llb_Mgr_t) {
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
func Llb_Nonlin4Quantify1(p *Llb_Mgr_t, pPart *Llb_Prt_t) int32 {
    var pVar *Llb_Var_t;
    var pTemp *Llb_Prt_t;
    var vSingles *Vec_Ptr_t;
    var bCube, bTemp *DdNode;
    var i, RetValue, nSizeNew int32;
    // create cube to be quantified
    bCube = Llb_Nonlin4CreateCube1( p, pPart );   Cudd_Ref( bCube );
//    assert( !Cudd_IsConstant(bCube) );
    // derive new function
    pPart.bFunc = Cudd_bddExistAbstract( p.dd, bTemp = pPart.bFunc, bCube );  Cudd_Ref( pPart.bFunc );
    Cudd_RecursiveDeref( p.dd, bTemp );
    Cudd_RecursiveDeref( p.dd, bCube );
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
                Llb_Nonlin4RemoveVar( p, pVar );
            } else if Vec_IntSize(pVar.vParts) == 1 {
                Vec_PtrPushUnique( vSingles, Llb_MgrPart(p, Vec_IntEntry(pVar.vParts,0)) ); }
        }
    }

    // update partition
    pPart.nSize = nSizeNew;
    Vec_IntClear( pPart.vVars );
    for  i = 0; i < p.nVars; i++  {
        if p.pSupp[i] && Vec_IntEntry(p.vVars2Q, i) {
            Vec_IntPush( pPart.vVars, i ); }
    }
    // remove other variables
    Vec_PtrForEachEntry( Llb_Prt_t *, vSingles, pTemp, i )
        Llb_Nonlin4Quantify1( p, pTemp );
    Vec_PtrFree( vSingles );
    return 0;
}

/**Function*************************************************************

  Synopsis    [Quantifies singles belonging to one partition.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_Nonlin4Quantify2(p *Llb_Mgr_t, pPart1 *Llb_Prt_t, pPart2 *Llb_Prt_t) int32 {
    var fVerbose int32;
    var pVar *Llb_Var_t;
    var pTemp *Llb_Prt_t;
    var vSingles *Vec_Ptr_t;
    var bCube, bFunc *DdNode;
    var i, RetValue, nSuppSize int32;
//    int iPart1 = pPart1->iPart;
//    int iPart2 = pPart2->iPart;
    var liveBeg, liveEnd int32;

    // create cube to be quantified
    bCube = Llb_Nonlin4CreateCube2( p, pPart1, pPart2 );   Cudd_Ref( bCube );

//printf( "Quantifying  " ); Extra_bddPrintSupport( p->dd, bCube );  printf( "\n" );

if fVerbose != 0 {
fmt.Printf("\n");
fmt.Printf("\n");
Llb_Nonlin4Print( p );
fmt.Printf("Conjoining partitions %d and %d.\n", pPart1.iPart, pPart2.iPart);
Extra_bddPrintSupport( p.dd, bCube );  fmt.Printf("\n");
}
liveBeg = p.dd.keys - p.dd.dead;
    bFunc = Cudd_bddAndAbstract( p.dd, pPart1.bFunc, pPart2.bFunc, bCube );  
liveEnd = p.dd.keys - p.dd.dead;
//printf( "%d ", liveEnd-liveBeg );

    if bFunc == nil {
        Cudd_RecursiveDeref( p.dd, bCube );
        return 0;
    }
    Cudd_Ref( bFunc );
    Cudd_RecursiveDeref( p.dd, bCube );

//printf( "Creating part %d ", p->iPartFree ); Extra_bddPrintSupport( p->dd, bFunc );  printf( "\n" );

//printf( "Creating %d\n", p->iPartFree );

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
        if p.pSupp[i] && Vec_IntEntry(p.vVars2Q, i) {
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
            Llb_Nonlin4RemoveVar( p, pVar );
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
            Llb_Nonlin4RemoveVar( p, pVar );
        } else if Vec_IntSize(pVar.vParts) == 1 {
            if fVerbose != 0 {
                fmt.Printf("Adding partition %d because of var %d.\n", Llb_MgrPart(p, Vec_IntEntry(pVar.vParts,0)).iPart, pVar.iVar); }
            Vec_PtrPushUnique( vSingles, Llb_MgrPart(p, Vec_IntEntry(pVar.vParts,0)) );
        }
    }
    // remove partitions
    Llb_Nonlin4RemovePart( p, pPart1 );
    Llb_Nonlin4RemovePart( p, pPart2 );
    // remove other variables
if fVerbose != 0 {
Llb_Nonlin4Print( p ); }
    Vec_PtrForEachEntry( Llb_Prt_t *, vSingles, pTemp, i )
    {
if fVerbose != 0 {
fmt.Printf("Updating partitiong %d with singlton vars.\n", pTemp.iPart); }
        Llb_Nonlin4Quantify1( p, pTemp );
    }
if fVerbose != 0 {
Llb_Nonlin4Print( p ); }
    Vec_PtrFree( vSingles );
    return 1;
}

/**Function*************************************************************

  Synopsis    [Computes volume of the cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_Nonlin4CutNodes_rec(p *Aig_Man_t, pObj *Aig_Obj_t, vNodes *Vec_Ptr_t) {
    if Aig_ObjIsTravIdCurrent(p, pObj) {
        return; }
    Aig_ObjSetTravIdCurrent(p, pObj);
    if Saig_ObjIsLi(p, pObj) {
        Llb_Nonlin4CutNodes_rec(p, Aig_ObjFanin0(pObj), vNodes);
        return;
    }
    if Aig_ObjIsConst1(pObj) {
        return; }
    assert(Aig_ObjIsNode(pObj), "Aig_ObjIsNode(pObj)");
    Llb_Nonlin4CutNodes_rec(p, Aig_ObjFanin0(pObj), vNodes);
    Llb_Nonlin4CutNodes_rec(p, Aig_ObjFanin1(pObj), vNodes);
    Vec_PtrPush( vNodes, pObj );
}

/**Function*************************************************************

  Synopsis    [Computes volume of the cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_Nonlin4CutNodes(p *Aig_Man_t, vLower *Vec_Ptr_t, vUpper *Vec_Ptr_t) *Vec_Ptr_t {
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
        Llb_Nonlin4CutNodes_rec( p, pObj, vNodes );
    return vNodes;
}

/**Function*************************************************************

  Synopsis    [Starts non-linear quantification scheduling.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_Nonlin4AddPair(p *Llb_Mgr_t, iPart int32, iVar int32) {
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
func Llb_Nonlin4AddPartition(p *Llb_Mgr_t, i int32, bFunc *DdNode) {
    var k, nSuppSize int32;
    // inlining of macro Cudd_IsConstant
    // inlining of macro Cudd_Regular
    assert(!((((DdNode *)((ptruint)(bFunc) & ~01))).index == CUDD_CONST_INDEX), "!/*__starting_point_inline_macro:Cudd_IsConstant__*/((/*__starting_point_inline_macro:Cudd_Regular__*/((DdNode *)((ptruint)(bFunc) & ~01))).index == CUDD_CONST_INDEX)");
//printf( "Creating init %d\n", i );
    // create partition
    p.pParts[i] = new(Llb_Prt_t);
    p.pParts[i].iPart = i;
    p.pParts[i].bFunc = bFunc;  Cudd_Ref( bFunc );
    p.pParts[i].vVars = Vec_IntAlloc( 8 );
    // add support dependencies
    nSuppSize = 0;
    Extra_SupportArray( p.dd, bFunc, p.pSupp );
    for  k = 0; k < p.nVars; k++  {
        nSuppSize += p.pSupp[k];
        if p.pSupp[k] && Vec_IntEntry(p.vVars2Q, k) {
            Llb_Nonlin4AddPair( p, i, k ); }
    }
    p.nSuppMax = Abc_MaxInt( p.nSuppMax, nSuppSize ); 
}

/**Function*************************************************************

  Synopsis    [Checks that each var appears in at least one partition.]

  Description []
               
  SideEffects []

  SeeAlso     []
**********************************************************************/
func Llb_Nonlin4CheckVars(p *Llb_Mgr_t) {
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
func Llb_Nonlin4NextPartitions(p *Llb_Mgr_t, ppPart1 **Llb_Prt_t, ppPart2 **Llb_Prt_t) int32 {
    var pVar *Llb_Var_t;
    var pVarBest *Llb_Var_t = nil;
    var pPart *Llb_Prt_t;
    var pPart1Best *Llb_Prt_t = nil;
    var pPart2Best *Llb_Prt_t = nil;
    var i int32;
    Llb_Nonlin4CheckVars( p );
    // find variable with minimum score
    // inlining of macro Llb_MgrForEachVar
    for  i = 0; (i < p.nVars) && (((pVar) = Llb_MgrVar(p, i)), 1); i++  { if pVar == nil {} else {
        if p.nSizeMax && pVar.nScore > p.nSizeMax {
            continue; }
//        if ( pVarBest == NULL || Vec_IntSize(pVarBest->vParts) * pVarBest->nScore > Vec_IntSize(pVar->vParts) * pVar->nScore )
        if pVarBest == nil || pVarBest.nScore > pVar.nScore {
            pVarBest = pVar; }
//        printf( "%d ", pVar->nScore );
    }
    }
//printf( "\n" );
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
//printf( "Selecting %d and parts %d and %d\n", pVarBest->iVar, pPart1Best->nSize, pPart2Best->nSize );
//Extra_bddPrintSupport( p->dd, pPart1Best->bFunc ); printf( "\n" );
//Extra_bddPrintSupport( p->dd, pPart2Best->bFunc ); printf( "\n" );

    *ppPart1 = pPart1Best;
    *ppPart2 = pPart2Best;
    return 1; 
}

/**Function*************************************************************

  Synopsis    [Recomputes scores after variable reordering.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_Nonlin4RecomputeScores(p *Llb_Mgr_t) {
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
func Llb_Nonlin4VerifyScores(p *Llb_Mgr_t) {
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
func Llb_Nonlin4Alloc(dd *DdManager, vParts *Vec_Ptr_t, bCurrent *DdNode, vVars2Q *Vec_Int_t, nSizeMax int32) *Llb_Mgr_t {
    var p *Llb_Mgr_t;
    var bFunc *DdNode;
    var i int32;
    p = new(Llb_Mgr_t);
    p.dd        = dd;
    p.nSizeMax  = nSizeMax;
    p.vVars2Q   = vVars2Q;
    p.nVars     = Cudd_ReadSize(dd);
    p.iPartFree = Vec_PtrSize(vParts);
    p.pVars     = ABC_CALLOC( Llb_Var_t *, p.nVars );
    p.pParts    = ABC_CALLOC( Llb_Prt_t *, 2 * p.iPartFree + 2 );
    p.pSupp     = ABC_ALLOC( int, Cudd_ReadSize(dd) );
    // add pairs (refs are consumed inside)
    Vec_PtrForEachEntry( DdNode *, vParts, bFunc, i )
        Llb_Nonlin4AddPartition( p, i, bFunc );
    // add partition
    if bCurrent != nil {
        Llb_Nonlin4AddPartition( p, p.iPartFree++, bCurrent ); }
    return p;
}

/**Function*************************************************************

  Synopsis    [Stops non-linear quantification scheduling.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_Nonlin4Free(p *Llb_Mgr_t) {
    var pPart *Llb_Prt_t;
    var pVar *Llb_Var_t;
    var i int32;
    // inlining of macro Llb_MgrForEachVar
    for  i = 0; (i < p.nVars) && (((pVar) = Llb_MgrVar(p, i)), 1); i++  { if pVar == nil {} else {
        Llb_Nonlin4RemoveVar( p, pVar ); }
    }
    // inlining of macro Llb_MgrForEachPart
    for  i = 0; (i < p.iPartFree) && (((pPart) = Llb_MgrPart(p, i)), 1); i++  { if pPart == nil {} else {
        Llb_Nonlin4RemovePart( p, pPart ); }
    }
    ABC_FREE( p.pVars );
    ABC_FREE( p.pParts );
    ABC_FREE( p.pSupp );
    ABC_FREE( p );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_Nonlin4Image(dd *DdManager, vParts *Vec_Ptr_t, bCurrent *DdNode, vVars2Q *Vec_Int_t) *DdNode {
    var pPart, pPart1, pPart2 *Llb_Prt_t;
    var p *Llb_Mgr_t;
    var bFunc, bTemp *DdNode;
    var i, nReorders int32;
    // start the manager
    p = Llb_Nonlin4Alloc( dd, vParts, bCurrent, vVars2Q, 0 );
    // remove singles
    // inlining of macro Llb_MgrForEachPart
    for  i = 0; (i < p.iPartFree) && (((pPart) = Llb_MgrPart(p, i)), 1); i++  { if pPart == nil {} else
        if Llb_Nonlin4HasSingletonVars(p, pPart) != 0 {
            Llb_Nonlin4Quantify1( p, pPart ); }
    }
    // compute scores
    Llb_Nonlin4RecomputeScores( p );
    // iteratively quantify variables
    for  Llb_Nonlin4NextPartitions(p, &pPart1, &pPart2) != 0  {
        nReorders = Cudd_ReadReorderings(dd);
        if !(Llb_Nonlin4Quantify2( p, pPart1, pPart2 ) != 0) {
            Llb_Nonlin4Free( p );
            return nil;
        }
        if nReorders < Cudd_ReadReorderings(dd) {
            Llb_Nonlin4RecomputeScores( p ); }
//        else
//            Llb_Nonlin4VerifyScores( p );
    }
    // load partitions
    bFunc = Cudd_ReadOne(p.dd);   Cudd_Ref( bFunc );
    // inlining of macro Llb_MgrForEachPart
    for  i = 0; (i < p.iPartFree) && (((pPart) = Llb_MgrPart(p, i)), 1); i++  { if pPart == nil {} else {
        bFunc = Cudd_bddAnd( p.dd, bTemp = bFunc, pPart.bFunc );   Cudd_Ref( bFunc );
        Cudd_RecursiveDeref( p.dd, bTemp );
    }
    }
//    nSuppMax = p->nSuppMax;
    Llb_Nonlin4Free( p );
//printf( "\n" );
    // return
    Cudd_Deref( bFunc );
    return bFunc;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_Nonlin4Group(dd *DdManager, vParts *Vec_Ptr_t, vVars2Q *Vec_Int_t, nSizeMax int32) *Vec_Ptr_t {
    var vGroups *Vec_Ptr_t;
    var pPart, pPart1, pPart2 *Llb_Prt_t;
    var p *Llb_Mgr_t;
    var i, nReorders int32;//, clk = Abc_Clock();
    // start the manager
    p = Llb_Nonlin4Alloc( dd, vParts, nil, vVars2Q, nSizeMax );
    // remove singles
    // inlining of macro Llb_MgrForEachPart
    for  i = 0; (i < p.iPartFree) && (((pPart) = Llb_MgrPart(p, i)), 1); i++  { if pPart == nil {} else
        if Llb_Nonlin4HasSingletonVars(p, pPart) != 0 {
            Llb_Nonlin4Quantify1( p, pPart ); }
    }
    // compute scores
    Llb_Nonlin4RecomputeScores( p );
    // iteratively quantify variables
    for  Llb_Nonlin4NextPartitions(p, &pPart1, &pPart2) != 0  {
        nReorders = Cudd_ReadReorderings(dd);
        if !(Llb_Nonlin4Quantify2( p, pPart1, pPart2 ) != 0) {
            Llb_Nonlin4Free( p );
            return nil;
        }
        if nReorders < Cudd_ReadReorderings(dd) {
            Llb_Nonlin4RecomputeScores( p ); }
//        else
//            Llb_Nonlin4VerifyScores( p );
    }
    // load partitions
    vGroups = Vec_PtrAlloc( 1000 );
    // inlining of macro Llb_MgrForEachPart
    for  i = 0; (i < p.iPartFree) && (((pPart) = Llb_MgrPart(p, i)), 1); i++  { if pPart == nil {} else {
//printf( "Iteration %d ", pPart->iPart );
        // inlining of macro Cudd_IsConstant
        // inlining of macro Cudd_Regular
        if  (((DdNode *)((ptruint)(pPart.bFunc) & ~01))).index == CUDD_CONST_INDEX {
//printf( "Constant\n" );
            // inlining of macro Cudd_IsComplement
            assert(!((int32( ((ptrint) (pPart.bFunc) & 01)) != 0)), "!/*__starting_point_inline_macro:Cudd_IsComplement__*/(int32( ((ptrint) (pPart.bFunc) & 01)))");
            continue;
        }
//printf( "\n" );
        Vec_PtrPush( vGroups, pPart.bFunc );
        Cudd_Ref( pPart.bFunc );
//printf( "Part %d  ", pPart->iPart );
//Extra_bddPrintSupport( p->dd, pPart->bFunc ); printf( "\n" );
    }
    }
    Llb_Nonlin4Free( p );
//Abc_PrintTime( 1, "Reparametrization time", Abc_Clock() - clk );
    return vGroups;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

