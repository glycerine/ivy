package aig

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [aigUtil.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [AIG package.]

  Synopsis    [Various procedures.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - April 28, 2007.]

  Revision    [$Id: aigUtil.c,v 1.00 2007/04/28 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "aig.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Increments the current traversal ID of the network.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManIncrementTravId(p *Aig_Man_t) {
    if p.nTravIds >= (1<<30)-1 {
        Aig_ManCleanData( p ); }
    p.nTravIds++;
}

/**Function*************************************************************

  Synopsis    [Returns the time stamp.]

  Description [The file should be closed.]

  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Aig_TimeStamp_Buffer [100]byte;
func Aig_TimeStamp() string {
    
    var TimeStamp string;
    var ltime int64;
    // get the current time
    time( &ltime );
    TimeStamp = asctime( localtime( &ltime ) );
    TimeStamp[ len(TimeStamp) - 1 ] = 0;
    strcpy( spStatic_Aig_TimeStamp_Buffer, TimeStamp );
    return spStatic_Aig_TimeStamp_Buffer;
}

/**Function*************************************************************

  Synopsis    [Make sure AIG has not gaps in the numeric order.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManHasNoGaps(p *Aig_Man_t) int32 {
    return int32((Aig_ManObjNum(p) == Aig_ManCiNum(p) + Aig_ManCoNum(p) + Aig_ManNodeNum(p) + 1));
}

/**Function*************************************************************

  Synopsis    [Collect the latches.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManLevels(p *Aig_Man_t) int32 {
    var pObj *Aig_Obj_t;
    var i, LevelMax int32;
    // inlining of macro Aig_ManForEachCo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
        LevelMax = Abc_MaxInt( LevelMax, int32(Aig_ObjFanin0(pObj).Level) );
    }
    return LevelMax;
}

/**Function*************************************************************

  Synopsis    [Reset reference counters.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManResetRefs(p *Aig_Man_t) {
    var pObj *Aig_Obj_t;
    var i int32;
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        pObj.nRefs = 0; }
    }
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        if Aig_ObjFanin0(pObj) {
            Aig_ObjFanin0(pObj).nRefs++; }
        if Aig_ObjFanin1(pObj) {
            Aig_ObjFanin1(pObj).nRefs++; }
    }
    }
}

/**Function*************************************************************

  Synopsis    [Cleans fMarkA.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManCleanMarkA(p *Aig_Man_t) {
    var pObj *Aig_Obj_t;
    var i int32;
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        pObj.fMarkA = 0; }
    }
}

/**Function*************************************************************

  Synopsis    [Cleans fMarkB.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManCleanMarkB(p *Aig_Man_t) {
    var pObj *Aig_Obj_t;
    var i int32;
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        pObj.fMarkB = 0; }
    }
}

/**Function*************************************************************

  Synopsis    [Cleans fMarkB.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManCleanMarkAB(p *Aig_Man_t) {
    var pObj *Aig_Obj_t;
    var i int32;
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        pObj.fMarkA = pObj.fMarkB = 0; }
    }
}

/**Function*************************************************************

  Synopsis    [Cleans the data pointers for the nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManCleanData(p *Aig_Man_t) {
    var pObj *Aig_Obj_t;
    var i int32;
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        pObj.pData = nil; }
    }
}

/**Function*************************************************************

  Synopsis    [Cleans the data pointers for the nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManCleanNext(p *Aig_Man_t) {
    var pObj *Aig_Obj_t;
    var i int32;
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        pObj.pNext = nil; }
    }
}

/**Function*************************************************************

  Synopsis    [Recursively cleans the data pointers in the cone of the node.]

  Description [Applicable to small AIGs only because no caching is performed.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ObjCleanData_rec(pObj *Aig_Obj_t) {
    assert(!Aig_IsComplement(pObj), "!Aig_IsComplement(pObj)");
    assert(!Aig_ObjIsCo(pObj), "!Aig_ObjIsCo(pObj)");
    if Aig_ObjIsAnd(pObj) {
        Aig_ObjCleanData_rec( Aig_ObjFanin0(pObj) );
        Aig_ObjCleanData_rec( Aig_ObjFanin1(pObj) );
    }
    pObj.pData = nil;
}


/**Function*************************************************************

  Synopsis    [Detects multi-input gate rooted at this node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ObjCollectMulti_rec(pRoot *Aig_Obj_t, pObj *Aig_Obj_t, vSuper *Vec_Ptr_t) {
    if pRoot != pObj && (Aig_IsComplement(pObj) || Aig_ObjIsCi(pObj) || Aig_ObjType(pRoot) != Aig_ObjType(pObj)) {
        Vec_PtrPushUnique(vSuper, pObj);
        return;
    }
    Aig_ObjCollectMulti_rec( pRoot, Aig_ObjChild0(pObj), vSuper );
    Aig_ObjCollectMulti_rec( pRoot, Aig_ObjChild1(pObj), vSuper );
}

/**Function*************************************************************

  Synopsis    [Detects multi-input gate rooted at this node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ObjCollectMulti(pRoot *Aig_Obj_t, vSuper *Vec_Ptr_t) {
    assert(!Aig_IsComplement(pRoot), "!Aig_IsComplement(pRoot)");
    Vec_PtrClear( vSuper );
    Aig_ObjCollectMulti_rec( pRoot, pRoot, vSuper );
}

/**Function*************************************************************

  Synopsis    [Returns 1 if the node is the root of MUX or EXOR/NEXOR.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ObjIsMuxType(pNode *Aig_Obj_t) int32 {
    var pNode0, pNode1 *Aig_Obj_t;
    // check that the node is regular
    assert(!Aig_IsComplement(pNode), "!Aig_IsComplement(pNode)");
    // if the node is not AND, this is not MUX
    if !Aig_ObjIsAnd(pNode) {
        return 0; }
    // if the children are not complemented, this is not MUX
    if !Aig_ObjFaninC0(pNode) || !Aig_ObjFaninC1(pNode) {
        return 0; }
    // get children
    pNode0 = Aig_ObjFanin0(pNode);
    pNode1 = Aig_ObjFanin1(pNode);
    // if the children are not ANDs, this is not MUX
    if !Aig_ObjIsAnd(pNode0) || !Aig_ObjIsAnd(pNode1) {
        return 0; }
    // otherwise the node is MUX iff it has a pair of equal grandchildren
    return (Aig_ObjFanin0(pNode0) == Aig_ObjFanin0(pNode1) && (Aig_ObjFaninC0(pNode0) ^ Aig_ObjFaninC0(pNode1))) || 
           (Aig_ObjFanin0(pNode0) == Aig_ObjFanin1(pNode1) && (Aig_ObjFaninC0(pNode0) ^ Aig_ObjFaninC1(pNode1))) ||
           (Aig_ObjFanin1(pNode0) == Aig_ObjFanin0(pNode1) && (Aig_ObjFaninC1(pNode0) ^ Aig_ObjFaninC0(pNode1))) ||
           (Aig_ObjFanin1(pNode0) == Aig_ObjFanin1(pNode1) && (Aig_ObjFaninC1(pNode0) ^ Aig_ObjFaninC1(pNode1)));
}


/**Function*************************************************************

  Synopsis    [Recognizes what nodes are inputs of the EXOR.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ObjRecognizeExor(pObj *Aig_Obj_t, ppFan0 **Aig_Obj_t, ppFan1 **Aig_Obj_t) int32 {
    var p0, p1 *Aig_Obj_t;
    assert(!Aig_IsComplement(pObj), "!Aig_IsComplement(pObj)");
    if !Aig_ObjIsNode(pObj) {
        return 0; }
    if Aig_ObjIsExor(pObj) {
        *ppFan0 = Aig_ObjChild0(pObj);
        *ppFan1 = Aig_ObjChild1(pObj);
        return 1;
    }
    assert(Aig_ObjIsAnd(pObj), "Aig_ObjIsAnd(pObj)");
    p0 = Aig_ObjChild0(pObj);
    p1 = Aig_ObjChild1(pObj);
    if !Aig_IsComplement(p0) || !Aig_IsComplement(p1) {
        return 0; }
    p0 = Aig_Regular(p0);
    p1 = Aig_Regular(p1);
    if !Aig_ObjIsAnd(p0) || !Aig_ObjIsAnd(p1) {
        return 0; }
    if Aig_ObjFanin0(p0) != Aig_ObjFanin0(p1) || Aig_ObjFanin1(p0) != Aig_ObjFanin1(p1) {
        return 0; }
    if Aig_ObjFaninC0(p0) == Aig_ObjFaninC0(p1) || Aig_ObjFaninC1(p0) == Aig_ObjFaninC1(p1) {
        return 0; }
    *ppFan0 = Aig_ObjChild0(p0);
    *ppFan1 = Aig_ObjChild1(p0);
    return 1;
}

/**Function*************************************************************

  Synopsis    [Recognizes what nodes are control and data inputs of a MUX.]

  Description [If the node is a MUX, returns the control variable C.
  Assigns nodes T and E to be the then and else variables of the MUX. 
  Node C is never complemented. Nodes T and E can be complemented.
  This function also recognizes EXOR/NEXOR gates as MUXes.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ObjRecognizeMux(pNode *Aig_Obj_t, ppNodeT **Aig_Obj_t, ppNodeE **Aig_Obj_t) *Aig_Obj_t {
    var pNode0, pNode1 *Aig_Obj_t;
    assert(!Aig_IsComplement(pNode), "!Aig_IsComplement(pNode)");
    assert((Aig_ObjIsMuxType(pNode) != 0), "Aig_ObjIsMuxType(pNode)");
    // get children
    pNode0 = Aig_ObjFanin0(pNode);
    pNode1 = Aig_ObjFanin1(pNode);

    // find the control variable
    if Aig_ObjFanin1(pNode0) == Aig_ObjFanin1(pNode1) && (Aig_ObjFaninC1(pNode0) ^ Aig_ObjFaninC1(pNode1)) {
//        if ( Fraig_IsComplement(pNode1->p2) )
        if Aig_ObjFaninC1(pNode0) { // pNode2->p2 is positive phase of C
            *ppNodeT = Aig_Not(Aig_ObjChild0(pNode1));//pNode2->p1);
            *ppNodeE = Aig_Not(Aig_ObjChild0(pNode0));//pNode1->p1);
            return Aig_ObjChild1(pNode1);//pNode2->p2;
        
        } else { // pNode1->p2 is positive phase of C
            *ppNodeT = Aig_Not(Aig_ObjChild0(pNode0));//pNode1->p1);
            *ppNodeE = Aig_Not(Aig_ObjChild0(pNode1));//pNode2->p1);
            return Aig_ObjChild1(pNode0);//pNode1->p2;
        }
    
    } else if Aig_ObjFanin0(pNode0) == Aig_ObjFanin0(pNode1) && (Aig_ObjFaninC0(pNode0) ^ Aig_ObjFaninC0(pNode1)) {
//        if ( Fraig_IsComplement(pNode1->p1) )
        if Aig_ObjFaninC0(pNode0) { // pNode2->p1 is positive phase of C
            *ppNodeT = Aig_Not(Aig_ObjChild1(pNode1));//pNode2->p2);
            *ppNodeE = Aig_Not(Aig_ObjChild1(pNode0));//pNode1->p2);
            return Aig_ObjChild0(pNode1);//pNode2->p1;
        
        } else { // pNode1->p1 is positive phase of C
            *ppNodeT = Aig_Not(Aig_ObjChild1(pNode0));//pNode1->p2);
            *ppNodeE = Aig_Not(Aig_ObjChild1(pNode1));//pNode2->p2);
            return Aig_ObjChild0(pNode0);//pNode1->p1;
        }
    
    } else if Aig_ObjFanin0(pNode0) == Aig_ObjFanin1(pNode1) && (Aig_ObjFaninC0(pNode0) ^ Aig_ObjFaninC1(pNode1)) {
//        if ( Fraig_IsComplement(pNode1->p1) )
        if Aig_ObjFaninC0(pNode0) { // pNode2->p2 is positive phase of C
            *ppNodeT = Aig_Not(Aig_ObjChild0(pNode1));//pNode2->p1);
            *ppNodeE = Aig_Not(Aig_ObjChild1(pNode0));//pNode1->p2);
            return Aig_ObjChild1(pNode1);//pNode2->p2;
        
        } else { // pNode1->p1 is positive phase of C
            *ppNodeT = Aig_Not(Aig_ObjChild1(pNode0));//pNode1->p2);
            *ppNodeE = Aig_Not(Aig_ObjChild0(pNode1));//pNode2->p1);
            return Aig_ObjChild0(pNode0);//pNode1->p1;
        }
    
    } else if Aig_ObjFanin1(pNode0) == Aig_ObjFanin0(pNode1) && (Aig_ObjFaninC1(pNode0) ^ Aig_ObjFaninC0(pNode1)) {
//        if ( Fraig_IsComplement(pNode1->p2) )
        if Aig_ObjFaninC1(pNode0) { // pNode2->p1 is positive phase of C
            *ppNodeT = Aig_Not(Aig_ObjChild1(pNode1));//pNode2->p2);
            *ppNodeE = Aig_Not(Aig_ObjChild0(pNode0));//pNode1->p1);
            return Aig_ObjChild0(pNode1);//pNode2->p1;
        
        } else { // pNode1->p2 is positive phase of C
            *ppNodeT = Aig_Not(Aig_ObjChild0(pNode0));//pNode1->p1);
            *ppNodeE = Aig_Not(Aig_ObjChild1(pNode1));//pNode2->p2);
            return Aig_ObjChild1(pNode0);//pNode1->p2;
        }
    }
    assert(false, "0"); // this is not MUX
    return nil;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ObjReal_rec(pObj *Aig_Obj_t) *Aig_Obj_t {
    var pObjNew *Aig_Obj_t;
    var pObjR *Aig_Obj_t = Aig_Regular(pObj);
    if !Aig_ObjIsBuf(pObjR) {
        return pObj; }
    pObjNew = Aig_ObjReal_rec( Aig_ObjChild0(pObjR) );
    return Aig_NotCond( pObjNew, Aig_IsComplement(pObj) );
}

/**Function*************************************************************

  Synopsis    [Procedure used for sorting the nodes in increasing order of IDs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ObjCompareIdIncrease(pp1 **Aig_Obj_t, pp2 **Aig_Obj_t) int32 {
    var Diff int32  = Aig_ObjId(*pp1) - Aig_ObjId(*pp2);
    if Diff < 0 {
        return -1; }
    if Diff > 0 { 
        return 1; }
    return 0; 
}


/**Function*************************************************************

  Synopsis    [Prints Eqn formula for the AIG rooted at this node.]

  Description [The formula is in terms of PIs, which should have
  their names assigned in pObj->pData fields.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ObjPrintEqn(pFile *os.File, pObj *Aig_Obj_t, vLevels *Vec_Vec_t, Level int32) {
    var vSuper *Vec_Ptr_t;
    var pFanin *Aig_Obj_t;
    var fCompl, i int32;
    // store the complemented attribute
    fCompl = Aig_IsComplement(pObj);
    pObj = Aig_Regular(pObj);
    // constant case
    if Aig_ObjIsConst1(pObj) {
        fmt.Fprintf(pFile, "%d", !fCompl);
        return;
    }
    // PI case
    if Aig_ObjIsCi(pObj) {
        fmt.Fprintf(pFile, "%s%s", fCompl? "!" : "", (char*)pObj.pData);
        return;
    }
    // AND case
    Vec_VecExpand( vLevels, Level );
    vSuper = Vec_VecEntry(vLevels, Level);
    Aig_ObjCollectMulti( pObj, vSuper );
    fmt.Fprintf(pFile, "%s", (Level==0? "" : "("));
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vSuper)) && (((pFanin) = (Aig_Obj_t *)Vec_PtrEntry(vSuper, i)), 1); i++  {
        Aig_ObjPrintEqn( pFile, Aig_NotCond(pFanin, fCompl), vLevels, Level+1 );
        if i < Vec_PtrSize(vSuper) - 1 {
            fmt.Fprintf(pFile, " %s ", fCompl? "+" : "*"); }
    }
    fmt.Fprintf(pFile, "%s", (Level==0? "" : ")"));
    return;
}

/**Function*************************************************************

  Synopsis    [Prints Verilog formula for the AIG rooted at this node.]

  Description [The formula is in terms of PIs, which should have
  their names assigned in pObj->pData fields.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ObjPrintVerilog(pFile *os.File, pObj *Aig_Obj_t, vLevels *Vec_Vec_t, Level int32) {
    var vSuper *Vec_Ptr_t;
    var pFanin, pFanin0, pFanin1, pFaninC *Aig_Obj_t;
    var fCompl, i int32;
    // store the complemented attribute
    fCompl = Aig_IsComplement(pObj);
    pObj = Aig_Regular(pObj);
    // constant case
    if Aig_ObjIsConst1(pObj) {
        fmt.Fprintf(pFile, "1\'b%d", !fCompl);
        return;
    }
    // PI case
    if Aig_ObjIsCi(pObj) {
        fmt.Fprintf(pFile, "%s%s", fCompl? "~" : "", (char*)pObj.pData);
        return;
    }
    // EXOR case
    if Aig_ObjIsExor(pObj) {
        Vec_VecExpand( vLevels, Level );
        vSuper = Vec_VecEntry( vLevels, Level );
        Aig_ObjCollectMulti( pObj, vSuper );
        fmt.Fprintf(pFile, "%s", (Level==0? "" : "("));
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(vSuper)) && (((pFanin) = (Aig_Obj_t *)Vec_PtrEntry(vSuper, i)), 1); i++  {
            Aig_ObjPrintVerilog( pFile, Aig_NotCond(pFanin, (fCompl && i==0)), vLevels, Level+1 );
            if i < Vec_PtrSize(vSuper) - 1 {
                fmt.Fprintf(pFile, " ^ "); }
        }
        fmt.Fprintf(pFile, "%s", (Level==0? "" : ")"));
        return;
    }
    // MUX case
    if Aig_ObjIsMuxType(pObj) != 0 {
        if Aig_ObjRecognizeExor( pObj, &pFanin0, &pFanin1 ) != 0 {
            fmt.Fprintf(pFile, "%s", (Level==0? "" : "("));
            Aig_ObjPrintVerilog( pFile, Aig_NotCond(pFanin0, fCompl), vLevels, Level+1 );
            fmt.Fprintf(pFile, " ^ ");
            Aig_ObjPrintVerilog( pFile, pFanin1, vLevels, Level+1 );
            fmt.Fprintf(pFile, "%s", (Level==0? "" : ")"));
        
        } else {
            pFaninC = Aig_ObjRecognizeMux( pObj, &pFanin1, &pFanin0 );
            fmt.Fprintf(pFile, "%s", (Level==0? "" : "("));
            Aig_ObjPrintVerilog( pFile, pFaninC, vLevels, Level+1 );
            fmt.Fprintf(pFile, " ? ");
            Aig_ObjPrintVerilog( pFile, Aig_NotCond(pFanin1, fCompl), vLevels, Level+1 );
            fmt.Fprintf(pFile, " : ");
            Aig_ObjPrintVerilog( pFile, Aig_NotCond(pFanin0, fCompl), vLevels, Level+1 );
            fmt.Fprintf(pFile, "%s", (Level==0? "" : ")"));
        }
        return;
    }
    // AND case
    Vec_VecExpand( vLevels, Level );
    vSuper = Vec_VecEntry(vLevels, Level);
    Aig_ObjCollectMulti( pObj, vSuper );
    fmt.Fprintf(pFile, "%s", (Level==0? "" : "("));
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vSuper)) && (((pFanin) = (Aig_Obj_t *)Vec_PtrEntry(vSuper, i)), 1); i++  {
        Aig_ObjPrintVerilog( pFile, Aig_NotCond(pFanin, fCompl), vLevels, Level+1 );
        if i < Vec_PtrSize(vSuper) - 1 {
            fmt.Fprintf(pFile, " %s ", fCompl? "|" : "&"); }
    }
    fmt.Fprintf(pFile, "%s", (Level==0? "" : ")"));
    return;
}


/**Function*************************************************************

  Synopsis    [Prints node in HAIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ObjPrintVerbose(pObj *Aig_Obj_t, fHaig int32) {
    assert(!Aig_IsComplement(pObj), "!Aig_IsComplement(pObj)");
    fmt.Printf("Node %d : ", pObj.Id);
    if Aig_ObjIsConst1(pObj) {
        fmt.Printf("constant 1");
    } else if Aig_ObjIsCi(pObj) {
        fmt.Printf("CI");
    } else if Aig_ObjIsCo(pObj) {
        fmt.Printf("CO( ");
        fmt.Printf("%d%s )", Aig_ObjFanin0(pObj).Id, (Aig_ObjFaninC0(pObj)? "\'" : " "));
    
    } else {
        fmt.Printf("AND( %d%s, %d%s )", Aig_ObjFanin0(pObj).Id, (Aig_ObjFaninC0(pObj)? "\'" : " "), Aig_ObjFanin1(pObj).Id, (Aig_ObjFaninC1(pObj)? "\'" : " ")); }
    fmt.Printf(" (refs = %3d)", Aig_ObjRefs(pObj));
}
func Aig_ObjPrintVerboseCone(p *Aig_Man_t, pRoot *Aig_Obj_t, fHaig int32) {
    extern Vec_Ptr_t * Aig_ManDfsArray( Aig_Man_t * p, Aig_Obj_t ** pNodes, int nNodes );
    var vNodes *Vec_Ptr_t;
    var pObj *Aig_Obj_t;
    var i int32;
    vNodes = Aig_ManDfsArray( p, &pRoot, 1 );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        Aig_ObjPrintVerbose( pObj, fHaig );
fmt.Printf("\n");
    }
    fmt.Printf("\n");
    Vec_PtrFree( vNodes );

}

/**Function*************************************************************

  Synopsis    [Prints node in HAIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManPrintVerbose(p *Aig_Man_t, fHaig int32) {
    var vNodes *Vec_Ptr_t;
    var pObj *Aig_Obj_t;
    var i int32;
    fmt.Printf("PIs: ");
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i)), 1); i++  {
        fmt.Printf(" %p", pObj);
    }
    fmt.Printf("\n");
    vNodes = Aig_ManDfs( p, 0 );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        Aig_ObjPrintVerbose( pObj, fHaig );
fmt.Printf("\n");
    }
    fmt.Printf("\n");
    Vec_PtrFree( vNodes );
}

/**Function*************************************************************

  Synopsis    [Write speculative miter for one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Aig_ManDump_Counter int32;
func Aig_ManDump(p *Aig_Man_t) { 
    
    var FileName string // was [200]byte in C

    // dump the logic into a file
    FileName = fmt.Sprintf("aigbug\\%03d.blif", ++spStatic_Aig_ManDump_Counter);
    Aig_ManDumpBlif( p, FileName, nil, nil );
    printf( "Intermediate AIG with %d nodes was written into file \"%s\".\n", Aig_ManNodeNum(p), FileName );
}

/**Function*************************************************************

  Synopsis    [Writes the AIG into a BLIF file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManDumpBlif(p *Aig_Man_t, pFileName string, vPiNames *Vec_Ptr_t, vPoNames *Vec_Ptr_t) {
    var pFile *os.File;
    var vNodes *Vec_Ptr_t;
    var pObj *Aig_Obj_t;
    var pObjLi *Aig_Obj_t;
    var pObjLo *Aig_Obj_t;
    var pConst1 *Aig_Obj_t = nil;
    var i, nDigits, Counter int32;
    if Aig_ManCoNum(p) == 0 {
        fmt.Printf("Aig_ManDumpBlif(): AIG manager does not have POs.\n");
        return;
    }
    // check if constant is used
    // inlining of macro Aig_ManForEachCo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
        if Aig_ObjIsConst1(Aig_ObjFanin0(pObj)) {
            pConst1 = Aig_ManConst1(p); }
    }
    // collect nodes in the DFS order
    vNodes = Aig_ManDfs( p, 1 );
    // assign IDs to objects
    Aig_ManConst1(p).iData = Counter++;
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i)), 1); i++  {
        pObj.iData = Counter++;
    }
    // inlining of macro Aig_ManForEachCo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
        pObj.iData = Counter++;
    }
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        pObj.iData = Counter++;
    }
    nDigits = Abc_Base10Log( Counter );
    // write the file 
    pFile = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    fmt.Fprintf(pFile, "# BLIF file written by procedure Aig_ManDumpBlif()\n");
//    fprintf( pFile, "# http://www.eecs.berkeley.edu/~alanmi/abc/\n" );
    fmt.Fprintf(pFile, ".model %s\n", p.pName);
    // write PIs
    fmt.Fprintf(pFile, ".inputs");
    // inlining of macro Aig_ManForEachPiSeq
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Aig_ManCiNum(p)-Aig_ManRegNum(p)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i)), 1); i++  {
        if vPiNames != nil {
            fmt.Fprintf(pFile, " %s", (char*)Vec_PtrEntry(vPiNames, i));
        } else {
            fmt.Fprintf(pFile, " n%0*d", nDigits, pObj.iData); }
    }
    fmt.Fprintf(pFile, "\n");
    // write POs
    fmt.Fprintf(pFile, ".outputs");
    // inlining of macro Aig_ManForEachPoSeq
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Aig_ManCoNum(p)-Aig_ManRegNum(p)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
        if vPoNames != nil {
            fmt.Fprintf(pFile, " %s", (char*)Vec_PtrEntry(vPoNames, i));
        } else {
            fmt.Fprintf(pFile, " n%0*d", nDigits, pObj.iData); }
    }
    fmt.Fprintf(pFile, "\n");
    // write latches
    if Aig_ManRegNum(p) {
        fmt.Fprintf(pFile, "\n");
        // inlining of macro Aig_ManForEachLiLoSeq
        for  i = 0; (i < Aig_ManRegNum(p)) && (((pObjLi) = Aig_ManLi(p, i)), 1)            && (((pObjLo)=Aig_ManLo(p, i)), 1); i++  {
            fmt.Fprintf(pFile, ".latch");
            if vPoNames != nil {
                fmt.Fprintf(pFile, " %s", (char*)Vec_PtrEntry(vPoNames, Aig_ManCoNum(p)-Aig_ManRegNum(p)+i));
            } else {
                fmt.Fprintf(pFile, " n%0*d", nDigits, pObjLi.iData); }
            if vPiNames != nil {
                fmt.Fprintf(pFile, " %s", (char*)Vec_PtrEntry(vPiNames, Aig_ManCiNum(p)-Aig_ManRegNum(p)+i));
            } else {
                fmt.Fprintf(pFile, " n%0*d", nDigits, pObjLo.iData); }
            fmt.Fprintf(pFile, " 0\n");
        }
        fmt.Fprintf(pFile, "\n");
    } 
    // write nodes
    if pConst1 != nil {
        fmt.Fprintf(pFile, ".names n%0*d\n 1\n", nDigits, pConst1.iData); }
    Aig_ManSetCioIds( p );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        fmt.Fprintf(pFile, ".names");
        if (vPiNames != nil) && Aig_ObjIsCi(Aig_ObjFanin0(pObj)) {
            fmt.Fprintf(pFile, " %s", (char*)Vec_PtrEntry(vPiNames, Aig_ObjCioId(Aig_ObjFanin0(pObj))));
        } else {
            fmt.Fprintf(pFile, " n%0*d", nDigits, Aig_ObjFanin0(pObj).iData); }
        if (vPiNames != nil) && Aig_ObjIsCi(Aig_ObjFanin1(pObj)) {
            fmt.Fprintf(pFile, " %s", (char*)Vec_PtrEntry(vPiNames, Aig_ObjCioId(Aig_ObjFanin1(pObj))));
        } else {
            fmt.Fprintf(pFile, " n%0*d", nDigits, Aig_ObjFanin1(pObj).iData); }
        fmt.Fprintf(pFile, " n%0*d\n", nDigits, pObj.iData);
        fmt.Fprintf(pFile, "%d%d 1\n", !Aig_ObjFaninC0(pObj), !Aig_ObjFaninC1(pObj));
    }
    // write POs
    // inlining of macro Aig_ManForEachCo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
        fmt.Fprintf(pFile, ".names");
        if (vPiNames != nil) && Aig_ObjIsCi(Aig_ObjFanin0(pObj)) {
            fmt.Fprintf(pFile, " %s", (char*)Vec_PtrEntry(vPiNames, Aig_ObjCioId(Aig_ObjFanin0(pObj))));
        } else {
            fmt.Fprintf(pFile, " n%0*d", nDigits, Aig_ObjFanin0(pObj).iData); }
        if vPoNames != nil {
            fmt.Fprintf(pFile, " %s\n", (char*)Vec_PtrEntry(vPoNames, Aig_ObjCioId(pObj)));
        } else {
            fmt.Fprintf(pFile, " n%0*d\n", nDigits, pObj.iData); }
        fmt.Fprintf(pFile, "%d 1\n", !Aig_ObjFaninC0(pObj));
    }
    Aig_ManCleanCioIds( p );
    fmt.Fprintf(pFile, ".end\n\n");
    _ = (pFile).Close();
    Vec_PtrFree( vNodes );
}

/**Function*************************************************************

  Synopsis    [Writes the AIG into a Verilog file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManDumpVerilog(p *Aig_Man_t, pFileName string) {
    var pFile *os.File;
    var vNodes *Vec_Ptr_t;
    var pObj *Aig_Obj_t;
    var pObjLi *Aig_Obj_t;
    var pObjLo *Aig_Obj_t;
    var pConst1 *Aig_Obj_t = nil;
    var i, nDigits, Counter int32;
    if Aig_ManCoNum(p) == 0 {
        fmt.Printf("Aig_ManDumpBlif(): AIG manager does not have POs.\n");
        return;
    }
    // check if constant is used
    // inlining of macro Aig_ManForEachCo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
        if Aig_ObjIsConst1(Aig_ObjFanin0(pObj)) {
            pConst1 = Aig_ManConst1(p); }
    }
    // collect nodes in the DFS order
    vNodes = Aig_ManDfs( p, 1 );
    // assign IDs to objects
    Aig_ManConst1(p).iData = Counter++;
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i)), 1); i++  {
        pObj.iData = Counter++;
    }
    // inlining of macro Aig_ManForEachCo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
        pObj.iData = Counter++;
    }
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        pObj.iData = Counter++;
    }
    nDigits = Abc_Base10Log( Counter );
    // write the file
    pFile = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    fmt.Fprintf(pFile, "// Verilog file written by procedure Aig_ManDumpVerilog()\n");
//    fprintf( pFile, "// http://www.eecs.berkeley.edu/~alanmi/abc/\n" );
    if Aig_ManRegNum(p) {
        fmt.Fprintf(pFile, "module %s ( clock", p.pName? p.pName: "test");
    } else {
        fmt.Fprintf(pFile, "module %s (", p.pName? p.pName: "test"); }
    // inlining of macro Aig_ManForEachPiSeq
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Aig_ManCiNum(p)-Aig_ManRegNum(p)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i)), 1); i++  {
        fmt.Fprintf(pFile, "%s n%0*d", ((Aig_ManRegNum(p) || i)? ",":""), nDigits, pObj.iData);
    }
    // inlining of macro Aig_ManForEachPoSeq
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Aig_ManCoNum(p)-Aig_ManRegNum(p)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
        fmt.Fprintf(pFile, ", n%0*d", nDigits, pObj.iData);
    }
    fmt.Fprintf(pFile, " );\n");

    // write PIs
    if Aig_ManRegNum(p) {
        fmt.Fprintf(pFile, "input clock;\n"); }
    // inlining of macro Aig_ManForEachPiSeq
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Aig_ManCiNum(p)-Aig_ManRegNum(p)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i)), 1); i++  {
        fmt.Fprintf(pFile, "input n%0*d;\n", nDigits, pObj.iData);
    }
    // write POs
    // inlining of macro Aig_ManForEachPoSeq
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Aig_ManCoNum(p)-Aig_ManRegNum(p)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
        fmt.Fprintf(pFile, "output n%0*d;\n", nDigits, pObj.iData);
    }
    // write latches
    if Aig_ManRegNum(p) {
    // inlining of macro Aig_ManForEachLiLoSeq
    for  i = 0; (i < Aig_ManRegNum(p)) && (((pObjLi) = Aig_ManLi(p, i)), 1)            && (((pObjLo)=Aig_ManLo(p, i)), 1); i++  {
        fmt.Fprintf(pFile, "reg n%0*d;\n", nDigits, pObjLo.iData);
    }
    // inlining of macro Aig_ManForEachLiLoSeq
    for  i = 0; (i < Aig_ManRegNum(p)) && (((pObjLi) = Aig_ManLi(p, i)), 1)            && (((pObjLo)=Aig_ManLo(p, i)), 1); i++  {
        fmt.Fprintf(pFile, "wire n%0*d;\n", nDigits, pObjLi.iData);
    }
    }
    // write nodes
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        fmt.Fprintf(pFile, "wire n%0*d;\n", nDigits, pObj.iData);
    }
    if pConst1 != nil {
        fmt.Fprintf(pFile, "wire n%0*d;\n", nDigits, pConst1.iData); }
    // write nodes
    if pConst1 != nil {
        fmt.Fprintf(pFile, "assign n%0*d = 1\'b1;\n", nDigits, pConst1.iData); }
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        fmt.Fprintf(pFile, "assign n%0*d = %sn%0*d & %sn%0*d;\n", nDigits, pObj.iData, !Aig_ObjFaninC0(pObj) ? " " : "~", nDigits, Aig_ObjFanin0(pObj).iData, !Aig_ObjFaninC1(pObj) ? " " : "~", nDigits, Aig_ObjFanin1(pObj).iData);
    }
    // write POs
    // inlining of macro Aig_ManForEachPoSeq
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Aig_ManCoNum(p)-Aig_ManRegNum(p)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
        fmt.Fprintf(pFile, "assign n%0*d = %sn%0*d;\n", nDigits, pObj.iData, !Aig_ObjFaninC0(pObj) ? " " : "~", nDigits, Aig_ObjFanin0(pObj).iData);
    }
    if Aig_ManRegNum(p) {
        // inlining of macro Aig_ManForEachLiLoSeq
        for  i = 0; (i < Aig_ManRegNum(p)) && (((pObjLi) = Aig_ManLi(p, i)), 1)            && (((pObjLo)=Aig_ManLo(p, i)), 1); i++  {
            fmt.Fprintf(pFile, "assign n%0*d = %sn%0*d;\n", nDigits, pObjLi.iData, !Aig_ObjFaninC0(pObjLi) ? " " : "~", nDigits, Aig_ObjFanin0(pObjLi).iData);
        }
    }

    // write initial state
    if Aig_ManRegNum(p) {
        // inlining of macro Aig_ManForEachLiLoSeq
        for  i = 0; (i < Aig_ManRegNum(p)) && (((pObjLi) = Aig_ManLi(p, i)), 1)            && (((pObjLo)=Aig_ManLo(p, i)), 1); i++  {
            fmt.Fprintf(pFile, "always @ (posedge clock) begin n%0*d <= n%0*d; end\n", nDigits, pObjLo.iData, nDigits, pObjLi.iData);
        }
        // inlining of macro Aig_ManForEachLiLoSeq
        for  i = 0; (i < Aig_ManRegNum(p)) && (((pObjLi) = Aig_ManLi(p, i)), 1)            && (((pObjLo)=Aig_ManLo(p, i)), 1); i++  {
            fmt.Fprintf(pFile, "initial begin n%0*d <= 1\'b0; end\n", nDigits, pObjLo.iData);
        }
    }

    fmt.Fprintf(pFile, "endmodule\n\n");
    _ = (pFile).Close();
    Vec_PtrFree( vNodes );
}

/**Function*************************************************************

  Synopsis    [Sets the PI/PO numbers.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManSetCioIds(p *Aig_Man_t) {
    var pObj *Aig_Obj_t;
    var i int32;
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i)), 1); i++  {
        pObj.CioId = i;
    }
    // inlining of macro Aig_ManForEachCo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
        pObj.CioId = i;
    }
}

/**Function*************************************************************

  Synopsis    [Sets the PI/PO numbers.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManCleanCioIds(p *Aig_Man_t) {
    var pObj *Aig_Obj_t;
    var i int32;
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i)), 1); i++  {
        pObj.pNext = nil;
    }
    // inlining of macro Aig_ManForEachCo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
        pObj.pNext = nil;
    }
}

/**Function*************************************************************

  Synopsis    [Sets the PI/PO numbers.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManChoiceNum(p *Aig_Man_t) int32 {
    var pObj *Aig_Obj_t;
    var i, Counter int32;
    // inlining of macro Aig_ManForEachNode
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
        Counter += Aig_ObjIsChoice( p, pObj ); }
    }
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Prints the fanouts of the control register.]

  Description [Useful only for Intel MC benchmarks.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManPrintControlFanouts(p *Aig_Man_t) {
    var pObj, pFanin0, pFanin1, pCtrl *Aig_Obj_t;
    var i int32;

    pCtrl = Aig_ManCi( p, Aig_ManCiNum(p) - 1 );

    fmt.Printf("Control signal:\n");
    Aig_ObjPrint( p, pCtrl ); fmt.Printf("\n\n");

    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        if !Aig_ObjIsNode(pObj) {
            continue; }
        assert(pObj != pCtrl, "pObj != pCtrl");
        pFanin0 = Aig_ObjFanin0(pObj);
        pFanin1 = Aig_ObjFanin1(pObj);
        if pFanin0 == pCtrl && Aig_ObjIsCi(pFanin1) {
            Aig_ObjPrint( p, pObj ); fmt.Printf("\n");
            Aig_ObjPrint( p, pFanin1 ); fmt.Printf("\n");
            fmt.Printf("\n");
        }
        if pFanin1 == pCtrl && Aig_ObjIsCi(pFanin0) {
            Aig_ObjPrint( p, pObj ); fmt.Printf("\n");
            Aig_ObjPrint( p, pFanin0 ); fmt.Printf("\n");
            fmt.Printf("\n");
        }
    }
    }
}

/**Function*************************************************************

  Synopsis    [Returns the composite name of the file.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Aig_FileNameGenericAppend_Buffer [1000]byte;
func Aig_FileNameGenericAppend(pBase string, pSuffix string) string {
    
    var pDot string;
    strcpy( spStatic_Aig_FileNameGenericAppend_Buffer, pBase );
    if pDot = strrchr( spStatic_Aig_FileNameGenericAppend_Buffer, '.' ) {
        *pDot = 0; }
    strcat( spStatic_Aig_FileNameGenericAppend_Buffer, pSuffix );
    if (pDot = strrchr( spStatic_Aig_FileNameGenericAppend_Buffer, '\\' )) || (pDot = strrchr( spStatic_Aig_FileNameGenericAppend_Buffer, '/' )) {
        return pDot+1; }
    return spStatic_Aig_FileNameGenericAppend_Buffer;
}

/**Function*************************************************************

  Synopsis    [Creates a sequence of random numbers.]

  Description []
               
  SideEffects []

  SeeAlso     [http://en.wikipedia.org/wiki/LFSR]

***********************************************************************/
func Aig_ManRandomTest2() {
    var pFile *os.File;
    unsigned int lfsr = 1;
    unsigned int period = 0; 
    pFile = os.OpenFile("rand.txt", os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    for {
//        lfsr = (lfsr >> 1) ^ (-(lfsr & 1u) & 0xd0000001u); // taps 32 31 29 1 
        lfsr = 1; // to prevent the warning
        ++period;
        fmt.Fprintf(pFile, "%10d : %10d ", period, lfsr);
//        Extra_PrintBinary( pFile, &lfsr, 32 );
        fmt.Fprintf(pFile, "\n");
        if period == 20000 {
            break; }
    
if !(lfsr != uint32(1)) { break }
} 
    _ = (pFile).Close();
}

/**Function*************************************************************

  Synopsis    [Creates a sequence of random numbers.]

  Description []
               
  SideEffects []

  SeeAlso     [http://www.codeproject.com/KB/recipes/SimpleRNG.aspx]

***********************************************************************/
func Aig_ManRandomTest1() {
    var pFile *os.File;
    unsigned int lfsr;
    unsigned int period = 0; 
    pFile = os.OpenFile("rand.txt", os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    for {
        lfsr = Aig_ManRandom( 0 );
        ++period;
        fmt.Fprintf(pFile, "%10d : %10d ", period, lfsr);
//        Extra_PrintBinary( pFile, &lfsr, 32 );
        fmt.Fprintf(pFile, "\n");
        if period == 20000 {
            break; }
    
if !(lfsr != uint32(1)) { break }
} 
    _ = (pFile).Close();
}

 
const NUMBER1 uint32 =  3716960521
const NUMBER2 uint32 =  2174103536

/**Function*************************************************************

  Synopsis    [Creates a sequence of random numbers.]

  Description []
               
  SideEffects []

  SeeAlso     [http://www.codeproject.com/KB/recipes/SimpleRNG.aspx]

***********************************************************************/
func Aig_ManRandom(fReset int32) uint32 {
    static unsigned int m_z = NUMBER1;
    static unsigned int m_w = NUMBER2;
    if fReset != 0 {
        m_z = NUMBER1;
        m_w = NUMBER2;
    }
    m_z = 36969 * (m_z & 65535) + (m_z >> 16);
    m_w = 18000 * (m_w & 65535) + (m_w >> 16);
    return (m_z << 16) + m_w;
}

/**Function*************************************************************

  Synopsis    [Creates a sequence of random numbers.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManRandom64(fReset int32) word {
    var Res word  = (word)Aig_ManRandom(fReset);
    return Res | ((word)Aig_ManRandom(0) << 32);
}


/**Function*************************************************************

  Synopsis    [Creates random info for the primary inputs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManRandomInfo(vInfo *Vec_Ptr_t, iInputStart int32, iWordStart int32, iWordStop int32) {
    var pInfo *uint32;
    var i, w int32;
    // inlining of macro Vec_PtrForEachEntryStart
    for  i = iInputStart; (i < Vec_PtrSize(vInfo)) && (((pInfo) = (unsigned *)Vec_PtrEntry(vInfo, i)), 1); i++  {
        for  w = iWordStart; w < iWordStop; w++  {
            pInfo[w] = Aig_ManRandom(0);
        }
    }
}

/**Function*************************************************************

  Synopsis    [Returns the result of merging the two vectors.]

  Description [Assumes that the vectors are sorted in the increasing order.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_NodeUnionLists(vArr1 *Vec_Ptr_t, vArr2 *Vec_Ptr_t, vArr *Vec_Ptr_t) {
    var pBeg **Aig_Obj_t   = (Aig_Obj_t **)vArr.pArray;
    var pBeg1 **Aig_Obj_t  = (Aig_Obj_t **)vArr1.pArray;
    var pBeg2 **Aig_Obj_t  = (Aig_Obj_t **)vArr2.pArray;
    var pEnd1 **Aig_Obj_t  = (Aig_Obj_t **)vArr1.pArray + vArr1.nSize;
    var pEnd2 **Aig_Obj_t  = (Aig_Obj_t **)vArr2.pArray + vArr2.nSize;
    Vec_PtrGrow( vArr, Vec_PtrSize(vArr1) + Vec_PtrSize(vArr2) );
    pBeg  = (Aig_Obj_t **)vArr.pArray;
    for  pBeg1 < pEnd1 && pBeg2 < pEnd2  {
        if (*pBeg1).Id == (*pBeg2).Id {
            *pBeg++ = *pBeg1++, pBeg2++;
        } else if (*pBeg1).Id < (*pBeg2).Id {
            *pBeg++ = *pBeg1++;
        } else { 
            *pBeg++ = *pBeg2++; }
    }
    for  pBeg1 < pEnd1  {
        *pBeg++ = *pBeg1++;
    }
    for  pBeg2 < pEnd2  {
        *pBeg++ = *pBeg2++;
    }
    vArr.nSize = pBeg - (Aig_Obj_t **)vArr.pArray;
    assert(vArr.nSize <= vArr.nCap, "vArr.nSize <= vArr.nCap");
    assert(vArr.nSize >= vArr1.nSize, "vArr.nSize >= vArr1.nSize");
    assert(vArr.nSize >= vArr2.nSize, "vArr.nSize >= vArr2.nSize");
}

/**Function*************************************************************

  Synopsis    [Returns the result of intersecting the two vectors.]

  Description [Assumes that the vectors are sorted in the increasing order.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_NodeIntersectLists(vArr1 *Vec_Ptr_t, vArr2 *Vec_Ptr_t, vArr *Vec_Ptr_t) {
    var pBeg **Aig_Obj_t   = (Aig_Obj_t **)vArr.pArray;
    var pBeg1 **Aig_Obj_t  = (Aig_Obj_t **)vArr1.pArray;
    var pBeg2 **Aig_Obj_t  = (Aig_Obj_t **)vArr2.pArray;
    var pEnd1 **Aig_Obj_t  = (Aig_Obj_t **)vArr1.pArray + vArr1.nSize;
    var pEnd2 **Aig_Obj_t  = (Aig_Obj_t **)vArr2.pArray + vArr2.nSize;
    Vec_PtrGrow( vArr, Abc_MaxInt( Vec_PtrSize(vArr1), Vec_PtrSize(vArr2) ) );
    pBeg  = (Aig_Obj_t **)vArr.pArray;
    for  pBeg1 < pEnd1 && pBeg2 < pEnd2  {
        if (*pBeg1).Id == (*pBeg2).Id {
            *pBeg++ = *pBeg1++, pBeg2++;
        } else if (*pBeg1).Id < (*pBeg2).Id {
//            *pBeg++ = *pBeg1++;
            pBeg1++;
        } else { 
//            *pBeg++ = *pBeg2++;
            pBeg2++; }
    }
//    while ( pBeg1 < pEnd1 )
//        *pBeg++ = *pBeg1++;
//    while ( pBeg2 < pEnd2 )
//        *pBeg++ = *pBeg2++;
    vArr.nSize = pBeg - (Aig_Obj_t **)vArr.pArray;
    assert(vArr.nSize <= vArr.nCap, "vArr.nSize <= vArr.nCap");
    assert(vArr.nSize <= vArr1.nSize, "vArr.nSize <= vArr1.nSize");
    assert(vArr.nSize <= vArr2.nSize, "vArr.nSize <= vArr2.nSize");
}

//ABC_NAMESPACE_IMPL_END

//#include "proof/fra/fra.h"
//#include "aig/saig/saig.h"

//ABC_NAMESPACE_IMPL_START


// extern void Aig_ManCounterExampleValueStart( Aig_Man_t * pAig, Abc_Cex_t * pCex );

// extern void Aig_ManCounterExampleValueStop( Aig_Man_t * pAig );

// extern int Aig_ManCounterExampleValueLookup(  Aig_Man_t * pAig, int Id, int iFrame );


/**Function*************************************************************

  Synopsis    [Starts the process of retuning values for internal nodes.]

  Description [Should be called when pCex is available, before probing 
  any object for its value using Aig_ManCounterExampleValueLookup().]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManCounterExampleValueStart(pAig *Aig_Man_t, pCex *Abc_Cex_t) {
    var pObj, pObjRi, pObjRo *Aig_Obj_t;
    var Val0, Val1, nObjs, i, k, iBit int32;
    assert(Aig_ManRegNum(pAig) > 0, "Aig_ManRegNum(pAig) > 0"); // makes sense only for sequential AIGs
    assert(pAig.pData2 == nil, "pAig.pData2 == nil");    // if this fail, there may be a memory leak
    // allocate memory to store simulation bits for internal nodes
    // inlining of macro ABC_CALLOC
    pAig.pData2 = ((unsigned *) calloc((size_t)(Abc_BitWordNum( (pCex.iFrame + 1) * Aig_ManObjNumMax(pAig) )), sizeof(unsigned)));
    // the register values in the counter-example should be zero
    // inlining of macro Saig_ManForEachLo
    for  k = 0; (k < Saig_ManRegNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCis, k+Saig_ManPiNum(pAig))), 1); k++  {
        assert(Abc_InfoHasBit(pCex.pData, iBit++) == 0, "Abc_InfoHasBit(pCex.pData, iBit++) == 0");
    }
    // iterate through the timeframes
    nObjs = Aig_ManObjNumMax(pAig);
    for  i = 0; i <= pCex.iFrame; i++  {
        // set constant 1 node
        Abc_InfoSetBit( (unsigned *)pAig.pData2, nObjs * i + 0 );
        // set primary inputs according to the counter-example
        // inlining of macro Saig_ManForEachPi
        // inlining of macro Vec_PtrForEachEntryStop
        for  k = 0; (k < Saig_ManPiNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCis, k)), 1); k++  {
            if Abc_InfoHasBit(pCex.pData, iBit++) {
                Abc_InfoSetBit( (unsigned *)pAig.pData2, nObjs * i + Aig_ObjId(pObj) ); }
        }
        // compute values for each node
        // inlining of macro Aig_ManForEachNode
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vObjs, k)), 1); k++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
            Val0 = Abc_InfoHasBit( (unsigned *)pAig.pData2, nObjs * i + Aig_ObjFaninId0(pObj) );
            Val1 = Abc_InfoHasBit( (unsigned *)pAig.pData2, nObjs * i + Aig_ObjFaninId1(pObj) );
            if (Val0 ^ Aig_ObjFaninC0(pObj)) & (Val1 ^ Aig_ObjFaninC1(pObj)) {
                Abc_InfoSetBit( (unsigned *)pAig.pData2, nObjs * i + Aig_ObjId(pObj) ); }
        }
        }
        // derive values for combinational outputs
        // inlining of macro Aig_ManForEachCo
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(pAig.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCos, k)), 1); k++  {
            Val0 = Abc_InfoHasBit( (unsigned *)pAig.pData2, nObjs * i + Aig_ObjFaninId0(pObj) );
            if Val0 ^ Aig_ObjFaninC0(pObj) {
                Abc_InfoSetBit( (unsigned *)pAig.pData2, nObjs * i + Aig_ObjId(pObj) ); }
        }
        if i == pCex.iFrame {
            continue; }
        // transfer values to the register output of the next frame
        // inlining of macro Saig_ManForEachLiLo
        for  k = 0; (k < Saig_ManRegNum(pAig)) && (((pObjRi) = Saig_ManLi(pAig, k)), 1)            && (((pObjRo)=Saig_ManLo(pAig, k)), 1); k++  {
            if Abc_InfoHasBit( (unsigned *)pAig.pData2, nObjs * i + Aig_ObjId(pObjRi) ) {
                Abc_InfoSetBit( (unsigned *)pAig.pData2, nObjs * (i+1) + Aig_ObjId(pObjRo) ); }
        }
    }
    assert(iBit == pCex.nBits, "iBit == pCex.nBits");
    // check that the counter-example is correct, that is, the corresponding output is asserted
    assert(Abc_InfoHasBit( (unsigned *)pAig.pData2, nObjs * pCex.iFrame + Aig_ObjId(Aig_ManCo(pAig, pCex.iPo)) ), "Abc_InfoHasBit( (unsigned *)pAig.pData2, nObjs * pCex.iFrame + Aig_ObjId(Aig_ManCo(pAig, pCex.iPo)) )");
}

/**Function*************************************************************

  Synopsis    [Stops the process of retuning values for internal nodes.]

  Description [Should be called when probing is no longer needed]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManCounterExampleValueStop(pAig *Aig_Man_t) {
    assert(pAig.pData2 != nil, "pAig.pData2 != nil");    // if this fail, we try to call this procedure more than once
    free( pAig.pData2 );
    pAig.pData2 = nil;
}

/**Function*************************************************************

  Synopsis    [Returns the value of the given object in the given timeframe.]

  Description [Should be called to probe the value of an object with 
  the given ID (iFrame is a 0-based number of a time frame - should not 
  exceed the number of timeframes in the original counter-example).]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManCounterExampleValueLookup(pAig *Aig_Man_t, Id int32, iFrame int32) int32 {
    assert(Id >= 0 && Id < Aig_ManObjNumMax(pAig), "Id >= 0 && Id < Aig_ManObjNumMax(pAig)");
    return Abc_InfoHasBit( (unsigned *)pAig.pData2, Aig_ManObjNumMax(pAig) * iFrame + Id );
}

/**Function*************************************************************

  Synopsis    [Procedure to test the above code.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManCounterExampleValueTest(pAig *Aig_Man_t, pCex *Abc_Cex_t) {
    var pObj *Aig_Obj_t  = Aig_ManObj( pAig, Aig_ManObjNumMax(pAig)/2 );
    var iFrame int32  = Abc_MaxInt( 0, pCex.iFrame - 1 );
    fmt.Printf("\nUsing counter-example, which asserts output %d in frame %d.\n", pCex.iPo, pCex.iFrame);
    Aig_ManCounterExampleValueStart( pAig, pCex );
    fmt.Printf("Value of object %d in frame %d is %d.\n", Aig_ObjId(pObj), iFrame, Aig_ManCounterExampleValueLookup(pAig, Aig_ObjId(pObj), iFrame));
    Aig_ManCounterExampleValueStop( pAig );
}

/**Function*************************************************************

  Synopsis    [Handle the counter-example.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManSetPhase(pAig *Aig_Man_t) {
    var pObj *Aig_Obj_t;
    var i int32;
    // set the PI simulation information
    Aig_ManConst1( pAig ).fPhase = 1;
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pAig.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCis, i)), 1); i++  {
        pObj.fPhase = 0;
    }
    // simulate internal nodes
    // inlining of macro Aig_ManForEachNode
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
        pObj.fPhase = ( Aig_ObjFanin0(pObj).fPhase ^ Aig_ObjFaninC0(pObj) )
                     & ( Aig_ObjFanin1(pObj).fPhase ^ Aig_ObjFaninC1(pObj) ); }
    }
    // simulate PO nodes
    // inlining of macro Aig_ManForEachCo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pAig.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCos, i)), 1); i++  {
        pObj.fPhase = Aig_ObjFanin0(pObj).fPhase ^ Aig_ObjFaninC0(pObj);
    }
}


/**Function*************************************************************

  Synopsis    [Collects muxes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManMuxesCollect(pAig *Aig_Man_t) *Vec_Ptr_t {
    var vMuxes *Vec_Ptr_t;
    var pObj *Aig_Obj_t;
    var i int32;
    vMuxes = Vec_PtrAlloc( 100 );
    // inlining of macro Aig_ManForEachNode
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else
        if Aig_ObjIsMuxType(pObj) != 0 {
            Vec_PtrPush( vMuxes, pObj ); }
    }
    return vMuxes;
}

/**Function*************************************************************

  Synopsis    [Dereferences muxes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManMuxesDeref(pAig *Aig_Man_t, vMuxes *Vec_Ptr_t) {
    var pObj, pNodeT, pNodeE, pNodeC *Aig_Obj_t;
    var i int32;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vMuxes)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(vMuxes, i)), 1); i++  {
        if Aig_ObjRecognizeExor( pObj, &pNodeT, &pNodeE ) != 0 {
            pNodeT.nRefs--;
            pNodeE.nRefs--;
        
        } else {
            pNodeC = Aig_ObjRecognizeMux( pObj, &pNodeT, &pNodeE );
            pNodeC.nRefs--;
        }
    }
}

/**Function*************************************************************

  Synopsis    [References muxes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManMuxesRef(pAig *Aig_Man_t, vMuxes *Vec_Ptr_t) {
    var pObj, pNodeT, pNodeE, pNodeC *Aig_Obj_t;
    var i int32;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vMuxes)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(vMuxes, i)), 1); i++  {
        if Aig_ObjRecognizeExor( pObj, &pNodeT, &pNodeE ) != 0 {
            pNodeT.nRefs++;
            pNodeE.nRefs++;
        
        } else {
            pNodeC = Aig_ObjRecognizeMux( pObj, &pNodeT, &pNodeE );
            pNodeC.nRefs++;
        }
    }
}

/**Function*************************************************************

  Synopsis    [Complements the constraint outputs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManInvertConstraints(pAig *Aig_Man_t) {
    var pObj *Aig_Obj_t;
    var i int32;
    if Aig_ManConstrNum(pAig) == 0 {
        return; }
    // inlining of macro Saig_ManForEachPo
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Saig_ManPoNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCos, i)), 1); i++  {
        if i >= Saig_ManPoNum(pAig) - Aig_ManConstrNum(pAig) {
            Aig_ObjChild0Flip( pObj ); }
    }
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

