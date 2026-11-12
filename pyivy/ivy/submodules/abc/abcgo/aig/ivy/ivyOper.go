package ivy

/**CFile****************************************************************

  FileName    [ivyOper.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [And-Inverter Graph package.]

  Synopsis    [AIG operations.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - May 11, 2006.]

  Revision    [$Id: ivyOper.c,v 1.00 2006/05/11 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "ivy.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// procedure to detect an EXOR gate
func Ivy_ObjIsExorType(p0 *Ivy_Obj_t, p1 *Ivy_Obj_t, ppFan0 **Ivy_Obj_t, ppFan1 **Ivy_Obj_t) int32 {
    if !Ivy_IsComplement(p0) || !Ivy_IsComplement(p1) {
        return 0; }
    p0 = Ivy_Regular(p0);
    p1 = Ivy_Regular(p1);
    if !Ivy_ObjIsAnd(p0) || !Ivy_ObjIsAnd(p1) {
        return 0; }
    if Ivy_ObjFanin0(p0) != Ivy_ObjFanin0(p1) || Ivy_ObjFanin1(p0) != Ivy_ObjFanin1(p1) {
        return 0; }
    if Ivy_ObjFaninC0(p0) == Ivy_ObjFaninC0(p1) || Ivy_ObjFaninC1(p0) == Ivy_ObjFaninC1(p1) {
        return 0; }
    *ppFan0 = Ivy_ObjChild0(p0);
    *ppFan1 = Ivy_ObjChild1(p0);
    return 1;
}

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Perform one operation.]

  Description [The argument nodes can be complemented.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_Oper(p *Ivy_Man_t, p0 *Ivy_Obj_t, p1 *Ivy_Obj_t, Type Ivy_Type_t) *Ivy_Obj_t {
    if Type == IVY_AND {
        return Ivy_And( p, p0, p1 ); }
    if Type == IVY_EXOR {
        return Ivy_Exor( p, p0, p1 ); }
    assert(false, "0");
    return nil;
}

/**Function*************************************************************

  Synopsis    [Performs canonicization step.]

  Description [The argument nodes can be complemented.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_And(p *Ivy_Man_t, p0 *Ivy_Obj_t, p1 *Ivy_Obj_t) *Ivy_Obj_t {
//    Ivy_Obj_t * pFan0, * pFan1;
    // check trivial cases
    if p0 == p1 {
        return p0; }
    if p0 == Ivy_Not(p1) {
        return Ivy_Not(p.pConst1); }
    if Ivy_Regular(p0) == p.pConst1 {
        if p0 == p.pConst1 {
            return p1;
        } else {
            return Ivy_Not(p.pConst1);
        } }
    if Ivy_Regular(p1) == p.pConst1 {
        if p1 == p.pConst1 {
            return p0;
        } else {
            return Ivy_Not(p.pConst1);
        } }
    // check if it can be an EXOR gate
//    if ( Ivy_ObjIsExorType( p0, p1, &pFan0, &pFan1 ) )
//        return Ivy_CanonExor( pFan0, pFan1 );
    return Ivy_CanonAnd( p, p0, p1 );
}

/**Function*************************************************************

  Synopsis    [Performs canonicization step.]

  Description [The argument nodes can be complemented.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_Exor(p *Ivy_Man_t, p0 *Ivy_Obj_t, p1 *Ivy_Obj_t) *Ivy_Obj_t {
/*
    // check trivial cases
    if ( p0 == p1 )
        return Ivy_Not(p->pConst1);
    if ( p0 == Ivy_Not(p1) )
        return p->pConst1;
    if ( Ivy_Regular(p0) == p->pConst1 )
        return Ivy_NotCond( p1, p0 == p->pConst1 );
    if ( Ivy_Regular(p1) == p->pConst1 )
        return Ivy_NotCond( p0, p1 == p->pConst1 );
    // check the table
    return Ivy_CanonExor( p, p0, p1 );
*/
    return Ivy_Or( p, Ivy_And(p, p0, Ivy_Not(p1)), Ivy_And(p, Ivy_Not(p0), p1) );
}

/**Function*************************************************************

  Synopsis    [Implements Boolean OR.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_Or(p *Ivy_Man_t, p0 *Ivy_Obj_t, p1 *Ivy_Obj_t) *Ivy_Obj_t {
    return Ivy_Not( Ivy_And( p, Ivy_Not(p0), Ivy_Not(p1) ) );
}

/**Function*************************************************************

  Synopsis    [Implements ITE operation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_Mux(p *Ivy_Man_t, pC *Ivy_Obj_t, p1 *Ivy_Obj_t, p0 *Ivy_Obj_t) *Ivy_Obj_t {
    var pTempA1, pTempA2, pTempB1, pTempB2, pTemp *Ivy_Obj_t;
    var Count0, Count1 int32;
    // consider trivial cases
    if p0 == Ivy_Not(p1) {
        return Ivy_Exor( p, pC, p0 ); }
    // other cases can be added
    // implement the first MUX (F = C * x1 + C' * x0)
    pTempA1 = Ivy_TableLookup( p, Ivy_ObjCreateGhost(p, pC,          p1, IVY_AND, IVY_INIT_NONE) );
    pTempA2 = Ivy_TableLookup( p, Ivy_ObjCreateGhost(p, Ivy_Not(pC), p0, IVY_AND, IVY_INIT_NONE) );
    if pTempA1 && pTempA2 {
        pTemp = Ivy_TableLookup( p, Ivy_ObjCreateGhost(p, Ivy_Not(pTempA1), Ivy_Not(pTempA2), IVY_AND, IVY_INIT_NONE) );
        if pTemp != nil { return Ivy_Not(pTemp); }
    }
    Count0 = (pTempA1 != nil) + (pTempA2 != nil);
    // implement the second MUX (F' = C * x1' + C' * x0')
    pTempB1 = Ivy_TableLookup( p, Ivy_ObjCreateGhost(p, pC,          Ivy_Not(p1), IVY_AND, IVY_INIT_NONE) );
    pTempB2 = Ivy_TableLookup( p, Ivy_ObjCreateGhost(p, Ivy_Not(pC), Ivy_Not(p0), IVY_AND, IVY_INIT_NONE) );
    if pTempB1 && pTempB2 {
        pTemp = Ivy_TableLookup( p, Ivy_ObjCreateGhost(p, Ivy_Not(pTempB1), Ivy_Not(pTempB2), IVY_AND, IVY_INIT_NONE) );
        if pTemp != nil { return pTemp; }
    }
    Count1 = (pTempB1 != nil) + (pTempB2 != nil);
    // compare and decide which one to implement
    if Count0 >= Count1 {
        if pTempA1 {
            pTempA1 = pTempA1;
        } else {
            pTempA1 = Ivy_And(p, pC,          p1);
        }
        if pTempA2 {
            pTempA2 = pTempA2;
        } else {
            pTempA2 = Ivy_And(p, Ivy_Not(pC), p0);
        }
        return Ivy_Or( p, pTempA1, pTempA2 );
    }
    if pTempB1 {
        pTempB1 = pTempB1;
    } else {
        pTempB1 = Ivy_And(p, pC,          Ivy_Not(p1));
    }
    if pTempB2 {
        pTempB2 = pTempB2;
    } else {
        pTempB2 = Ivy_And(p, Ivy_Not(pC), Ivy_Not(p0));
    }
    return Ivy_Not( Ivy_Or( p, pTempB1, pTempB2 ) );

//    return Ivy_Or( Ivy_And(pC, p1), Ivy_And(Ivy_Not(pC), p0) );
}

/**Function*************************************************************

  Synopsis    [Implements ITE operation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_Maj(p *Ivy_Man_t, pA *Ivy_Obj_t, pB *Ivy_Obj_t, pC *Ivy_Obj_t) *Ivy_Obj_t {
    return Ivy_Or( p, Ivy_Or(p, Ivy_And(p, pA, pB), Ivy_And(p, pA, pC)), Ivy_And(p, pB, pC) );
}

/**Function*************************************************************

  Synopsis    [Constructs the well-balanced tree of gates.]

  Description [Disregards levels and possible logic sharing.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_Multi_rec(p *Ivy_Man_t, ppObjs **Ivy_Obj_t, nObjs int32, Type Ivy_Type_t) *Ivy_Obj_t {
    var pObj1, pObj2 *Ivy_Obj_t;
    if nObjs == 1 {
        return ppObjs[0]; }
    pObj1 = Ivy_Multi_rec( p, ppObjs,           nObjs/2,         Type );
    pObj2 = Ivy_Multi_rec( p, ppObjs + nObjs/2, nObjs - nObjs/2, Type );
    return Ivy_Oper( p, pObj1, pObj2, Type );
}

/**Function*************************************************************

  Synopsis    [Old code.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_Multi(p *Ivy_Man_t, pArgs **Ivy_Obj_t, nArgs int32, Type Ivy_Type_t) *Ivy_Obj_t {
    assert(Type == IVY_AND || Type == IVY_EXOR, "Type == IVY_AND || Type == IVY_EXOR");
    assert(nArgs > 0, "nArgs > 0");
    return Ivy_Multi_rec( p, pArgs, nArgs, Type );
}

/**Function*************************************************************

  Synopsis    [Implements the miter.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_Miter(p *Ivy_Man_t, vPairs *Vec_Ptr_t) *Ivy_Obj_t {
    var i int32;
    assert(vPairs.nSize > 0, "vPairs.nSize > 0");
    assert(vPairs.nSize % 2 == 0, "vPairs.nSize % 2 == 0");
    // go through the cubes of the node's SOP
    for  i = 0; i < vPairs.nSize; i += 2  {
        vPairs.pArray[i/2] = Ivy_Not( Ivy_Exor( p, (Ivy_Obj_t *)vPairs.pArray[i], (Ivy_Obj_t *)vPairs.pArray[i+1] ) );
    }
    vPairs.nSize = vPairs.nSize/2;
    return Ivy_Not( Ivy_Multi_rec( p, (Ivy_Obj_t **)vPairs.pArray, vPairs.nSize, IVY_AND ) );
}
 
/**Function*************************************************************

  Synopsis    [Performs canonicization step.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_Latch(p *Ivy_Man_t, pObj *Ivy_Obj_t, Init Ivy_Init_t) *Ivy_Obj_t {
    return Ivy_CanonLatch( p, pObj, Init );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

