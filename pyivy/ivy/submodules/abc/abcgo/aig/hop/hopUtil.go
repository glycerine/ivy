package hop

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [hopUtil.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [And-Inverter Graph package.]

  Synopsis    [Various procedures.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - May 11, 2006.]

  Revision    [$Id: hopUtil.c,v 1.00 2006/05/11 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "hop.h"

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
func Hop_ManIncrementTravId(p *Hop_Man_t) {
    if p.nTravIds >= (1<<30)-1 {
        Hop_ManCleanData( p ); }
    p.nTravIds++;
}

/**Function*************************************************************

  Synopsis    [Cleans the data pointers for the nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hop_ManCleanData(p *Hop_Man_t) {
    var pObj *Hop_Obj_t;
    var i int32;
    p.nTravIds = 1;
    Hop_ManConst1(p).pData = nil;
    // inlining of macro Hop_ManForEachPi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vPis)) && (((pObj) = (Hop_Obj_t *)Vec_PtrEntry(p.vPis, i)), 1); i++  {
        pObj.pData = nil;
    }
    // inlining of macro Hop_ManForEachPo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vPos)) && (((pObj) = (Hop_Obj_t *)Vec_PtrEntry(p.vPos, i)), 1); i++  {
        pObj.pData = nil;
    }
    // inlining of macro Hop_ManForEachNode
    for  i = 0; i < p.nTableSize; i++  {                                               if ((pObj) = p.pTable[i]) == nil {} else {
        pObj.pData = nil; }
    }
}

/**Function*************************************************************

  Synopsis    [Recursively cleans the data pointers in the cone of the node.]

  Description [Applicable to small AIGs only because no caching is performed.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hop_ObjCleanData_rec(pObj *Hop_Obj_t) {
    assert(!Hop_IsComplement(pObj), "!Hop_IsComplement(pObj)");
    assert(!Hop_ObjIsPo(pObj), "!Hop_ObjIsPo(pObj)");
    if Hop_ObjIsAnd(pObj) {
        Hop_ObjCleanData_rec( Hop_ObjFanin0(pObj) );
        Hop_ObjCleanData_rec( Hop_ObjFanin1(pObj) );
    }
    pObj.pData = nil;
}

/**Function*************************************************************

  Synopsis    [Detects multi-input gate rooted at this node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hop_ObjCollectMulti_rec(pRoot *Hop_Obj_t, pObj *Hop_Obj_t, vSuper *Vec_Ptr_t) {
    if pRoot != pObj && (Hop_IsComplement(pObj) || Hop_ObjIsPi(pObj) || Hop_ObjType(pRoot) != Hop_ObjType(pObj)) {
        Vec_PtrPushUnique(vSuper, pObj);
        return;
    }
    Hop_ObjCollectMulti_rec( pRoot, Hop_ObjChild0(pObj), vSuper );
    Hop_ObjCollectMulti_rec( pRoot, Hop_ObjChild1(pObj), vSuper );
}

/**Function*************************************************************

  Synopsis    [Detects multi-input gate rooted at this node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hop_ObjCollectMulti(pRoot *Hop_Obj_t, vSuper *Vec_Ptr_t) {
    assert(!Hop_IsComplement(pRoot), "!Hop_IsComplement(pRoot)");
    Vec_PtrClear( vSuper );
    Hop_ObjCollectMulti_rec( pRoot, pRoot, vSuper );
}

/**Function*************************************************************

  Synopsis    [Returns 1 if the node is the root of MUX or EXOR/NEXOR.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hop_ObjIsMuxType(pNode *Hop_Obj_t) int32 {
    var pNode0, pNode1 *Hop_Obj_t;
    // check that the node is regular
    assert(!Hop_IsComplement(pNode), "!Hop_IsComplement(pNode)");
    // if the node is not AND, this is not MUX
    if !Hop_ObjIsAnd(pNode) {
        return 0; }
    // if the children are not complemented, this is not MUX
    if !Hop_ObjFaninC0(pNode) || !Hop_ObjFaninC1(pNode) {
        return 0; }
    // get children
    pNode0 = Hop_ObjFanin0(pNode);
    pNode1 = Hop_ObjFanin1(pNode);
    // if the children are not ANDs, this is not MUX
    if !Hop_ObjIsAnd(pNode0) || !Hop_ObjIsAnd(pNode1) {
        return 0; }
    // otherwise the node is MUX iff it has a pair of equal grandchildren
    return (Hop_ObjFanin0(pNode0) == Hop_ObjFanin0(pNode1) && (Hop_ObjFaninC0(pNode0) ^ Hop_ObjFaninC0(pNode1))) || 
           (Hop_ObjFanin0(pNode0) == Hop_ObjFanin1(pNode1) && (Hop_ObjFaninC0(pNode0) ^ Hop_ObjFaninC1(pNode1))) ||
           (Hop_ObjFanin1(pNode0) == Hop_ObjFanin0(pNode1) && (Hop_ObjFaninC1(pNode0) ^ Hop_ObjFaninC0(pNode1))) ||
           (Hop_ObjFanin1(pNode0) == Hop_ObjFanin1(pNode1) && (Hop_ObjFaninC1(pNode0) ^ Hop_ObjFaninC1(pNode1)));
}


/**Function*************************************************************

  Synopsis    [Recognizes what nodes are inputs of the EXOR.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hop_ObjRecognizeExor(pObj *Hop_Obj_t, ppFan0 **Hop_Obj_t, ppFan1 **Hop_Obj_t) int32 {
    var p0, p1 *Hop_Obj_t;
    assert(!Hop_IsComplement(pObj), "!Hop_IsComplement(pObj)");
    if !Hop_ObjIsNode(pObj) {
        return 0; }
    if Hop_ObjIsExor(pObj) {
        *ppFan0 = Hop_ObjChild0(pObj);
        *ppFan1 = Hop_ObjChild1(pObj);
        return 1;
    }
    assert(Hop_ObjIsAnd(pObj), "Hop_ObjIsAnd(pObj)");
    p0 = Hop_ObjChild0(pObj);
    p1 = Hop_ObjChild1(pObj);
    if !Hop_IsComplement(p0) || !Hop_IsComplement(p1) {
        return 0; }
    p0 = Hop_Regular(p0);
    p1 = Hop_Regular(p1);
    if !Hop_ObjIsAnd(p0) || !Hop_ObjIsAnd(p1) {
        return 0; }
    if Hop_ObjFanin0(p0) != Hop_ObjFanin0(p1) || Hop_ObjFanin1(p0) != Hop_ObjFanin1(p1) {
        return 0; }
    if Hop_ObjFaninC0(p0) == Hop_ObjFaninC0(p1) || Hop_ObjFaninC1(p0) == Hop_ObjFaninC1(p1) {
        return 0; }
    *ppFan0 = Hop_ObjChild0(p0);
    *ppFan1 = Hop_ObjChild1(p0);
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
func Hop_ObjRecognizeMux(pNode *Hop_Obj_t, ppNodeT **Hop_Obj_t, ppNodeE **Hop_Obj_t) *Hop_Obj_t {
    var pNode0, pNode1 *Hop_Obj_t;
    assert(!Hop_IsComplement(pNode), "!Hop_IsComplement(pNode)");
    assert((Hop_ObjIsMuxType(pNode) != 0), "Hop_ObjIsMuxType(pNode)");
    // get children
    pNode0 = Hop_ObjFanin0(pNode);
    pNode1 = Hop_ObjFanin1(pNode);

    // find the control variable
    if Hop_ObjFanin1(pNode0) == Hop_ObjFanin1(pNode1) && (Hop_ObjFaninC1(pNode0) ^ Hop_ObjFaninC1(pNode1)) {
//        if ( Fraig_IsComplement(pNode1->p2) )
        if Hop_ObjFaninC1(pNode0) { // pNode2->p2 is positive phase of C
            *ppNodeT = Hop_Not(Hop_ObjChild0(pNode1));//pNode2->p1);
            *ppNodeE = Hop_Not(Hop_ObjChild0(pNode0));//pNode1->p1);
            return Hop_ObjChild1(pNode1);//pNode2->p2;
        
        } else { // pNode1->p2 is positive phase of C
            *ppNodeT = Hop_Not(Hop_ObjChild0(pNode0));//pNode1->p1);
            *ppNodeE = Hop_Not(Hop_ObjChild0(pNode1));//pNode2->p1);
            return Hop_ObjChild1(pNode0);//pNode1->p2;
        }
    
    } else if Hop_ObjFanin0(pNode0) == Hop_ObjFanin0(pNode1) && (Hop_ObjFaninC0(pNode0) ^ Hop_ObjFaninC0(pNode1)) {
//        if ( Fraig_IsComplement(pNode1->p1) )
        if Hop_ObjFaninC0(pNode0) { // pNode2->p1 is positive phase of C
            *ppNodeT = Hop_Not(Hop_ObjChild1(pNode1));//pNode2->p2);
            *ppNodeE = Hop_Not(Hop_ObjChild1(pNode0));//pNode1->p2);
            return Hop_ObjChild0(pNode1);//pNode2->p1;
        
        } else { // pNode1->p1 is positive phase of C
            *ppNodeT = Hop_Not(Hop_ObjChild1(pNode0));//pNode1->p2);
            *ppNodeE = Hop_Not(Hop_ObjChild1(pNode1));//pNode2->p2);
            return Hop_ObjChild0(pNode0);//pNode1->p1;
        }
    
    } else if Hop_ObjFanin0(pNode0) == Hop_ObjFanin1(pNode1) && (Hop_ObjFaninC0(pNode0) ^ Hop_ObjFaninC1(pNode1)) {
//        if ( Fraig_IsComplement(pNode1->p1) )
        if Hop_ObjFaninC0(pNode0) { // pNode2->p2 is positive phase of C
            *ppNodeT = Hop_Not(Hop_ObjChild0(pNode1));//pNode2->p1);
            *ppNodeE = Hop_Not(Hop_ObjChild1(pNode0));//pNode1->p2);
            return Hop_ObjChild1(pNode1);//pNode2->p2;
        
        } else { // pNode1->p1 is positive phase of C
            *ppNodeT = Hop_Not(Hop_ObjChild1(pNode0));//pNode1->p2);
            *ppNodeE = Hop_Not(Hop_ObjChild0(pNode1));//pNode2->p1);
            return Hop_ObjChild0(pNode0);//pNode1->p1;
        }
    
    } else if Hop_ObjFanin1(pNode0) == Hop_ObjFanin0(pNode1) && (Hop_ObjFaninC1(pNode0) ^ Hop_ObjFaninC0(pNode1)) {
//        if ( Fraig_IsComplement(pNode1->p2) )
        if Hop_ObjFaninC1(pNode0) { // pNode2->p1 is positive phase of C
            *ppNodeT = Hop_Not(Hop_ObjChild1(pNode1));//pNode2->p2);
            *ppNodeE = Hop_Not(Hop_ObjChild0(pNode0));//pNode1->p1);
            return Hop_ObjChild0(pNode1);//pNode2->p1;
        
        } else { // pNode1->p2 is positive phase of C
            *ppNodeT = Hop_Not(Hop_ObjChild0(pNode0));//pNode1->p1);
            *ppNodeE = Hop_Not(Hop_ObjChild1(pNode1));//pNode2->p2);
            return Hop_ObjChild1(pNode0);//pNode1->p2;
        }
    }
    assert(false, "0"); // this is not MUX
    return nil;
}


/**Function*************************************************************

  Synopsis    [Prints Eqn formula for the AIG rooted at this node.]

  Description [The formula is in terms of PIs, which should have
  their names assigned in pObj->pData fields.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hop_ObjPrintEqn(pFile *os.File, pObj *Hop_Obj_t, vLevels *Vec_Vec_t, Level int32) {
    var vSuper *Vec_Ptr_t;
    var pFanin *Hop_Obj_t;
    var fCompl, i int32;
    // store the complemented attribute
    fCompl = Hop_IsComplement(pObj);
    pObj = Hop_Regular(pObj);
    // constant case
    if Hop_ObjIsConst1(pObj) {
        fmt.Fprintf(pFile, "%d", !fCompl);
        return;
    }
    // PI case
    if Hop_ObjIsPi(pObj) {
        fmt.Fprintf(pFile, "%s%s", fCompl? "!" : "", (char*)pObj.pData);
        return;
    }
    // AND case
    Vec_VecExpand( vLevels, Level );
    vSuper = Vec_VecEntry(vLevels, Level);
    Hop_ObjCollectMulti( pObj, vSuper );
    fmt.Fprintf(pFile, "%s", (Level==0? "" : "("));
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vSuper)) && (((pFanin) = (Hop_Obj_t *)Vec_PtrEntry(vSuper, i)), 1); i++  {
        Hop_ObjPrintEqn( pFile, Hop_NotCond(pFanin, fCompl), vLevels, Level+1 );
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
func Hop_ObjPrintVerilog(pFile *os.File, pObj *Hop_Obj_t, vLevels *Vec_Vec_t, Level int32, fOnlyAnds int32) {
    var vSuper *Vec_Ptr_t;
    var pFanin, pFanin0, pFanin1, pFaninC *Hop_Obj_t;
    var fCompl, i int32;
    // store the complemented attribute
    fCompl = Hop_IsComplement(pObj);
    pObj = Hop_Regular(pObj);
    // constant case
    if Hop_ObjIsConst1(pObj) {
        fmt.Fprintf(pFile, "1\'b%d", !fCompl);
        return;
    }
    // PI case
    if Hop_ObjIsPi(pObj) {
        fmt.Fprintf(pFile, "%s%s", fCompl? "~" : "", (char*)pObj.pData);
        return;
    }
    // EXOR case
    if !(fOnlyAnds != 0) && Hop_ObjIsExor(pObj) {
        Vec_VecExpand( vLevels, Level );
        vSuper = Vec_VecEntry( vLevels, Level );
        Hop_ObjCollectMulti( pObj, vSuper );
        fmt.Fprintf(pFile, "%s", (Level==0? "" : "("));
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(vSuper)) && (((pFanin) = (Hop_Obj_t *)Vec_PtrEntry(vSuper, i)), 1); i++  {
            Hop_ObjPrintVerilog( pFile, Hop_NotCond(pFanin, (fCompl && i==0)), vLevels, Level+1, fOnlyAnds );
            if i < Vec_PtrSize(vSuper) - 1 {
                fmt.Fprintf(pFile, " ^ "); }
        }
        fmt.Fprintf(pFile, "%s", (Level==0? "" : ")"));
        return;
    }
    // MUX case
    if !(fOnlyAnds != 0) && (Hop_ObjIsMuxType(pObj) != 0) {
        if Hop_ObjRecognizeExor( pObj, &pFanin0, &pFanin1 ) != 0 {
            fmt.Fprintf(pFile, "%s", (Level==0? "" : "("));
            Hop_ObjPrintVerilog( pFile, Hop_NotCond(pFanin0, fCompl), vLevels, Level+1, fOnlyAnds );
            fmt.Fprintf(pFile, " ^ ");
            Hop_ObjPrintVerilog( pFile, pFanin1, vLevels, Level+1, fOnlyAnds );
            fmt.Fprintf(pFile, "%s", (Level==0? "" : ")"));
        
        } else {
            pFaninC = Hop_ObjRecognizeMux( pObj, &pFanin1, &pFanin0 );
            fmt.Fprintf(pFile, "%s", (Level==0? "" : "("));
            Hop_ObjPrintVerilog( pFile, pFaninC, vLevels, Level+1, fOnlyAnds );
            fmt.Fprintf(pFile, " ? ");
            Hop_ObjPrintVerilog( pFile, Hop_NotCond(pFanin1, fCompl), vLevels, Level+1, fOnlyAnds );
            fmt.Fprintf(pFile, " : ");
            Hop_ObjPrintVerilog( pFile, Hop_NotCond(pFanin0, fCompl), vLevels, Level+1, fOnlyAnds );
            fmt.Fprintf(pFile, "%s", (Level==0? "" : ")"));
        }
        return;
    }
    // AND case
    Vec_VecExpand( vLevels, Level );
    vSuper = Vec_VecEntry(vLevels, Level);
    Hop_ObjCollectMulti( pObj, vSuper );
    fmt.Fprintf(pFile, "%s", (Level==0? "" : "("));
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vSuper)) && (((pFanin) = (Hop_Obj_t *)Vec_PtrEntry(vSuper, i)), 1); i++  {
        Hop_ObjPrintVerilog( pFile, Hop_NotCond(pFanin, fCompl), vLevels, Level+1, fOnlyAnds );
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
func Hop_ObjPrintVerbose(pObj *Hop_Obj_t, fHaig int32) {
    assert(!Hop_IsComplement(pObj), "!Hop_IsComplement(pObj)");
    fmt.Printf("Node %p : ", pObj);
    if Hop_ObjIsConst1(pObj) {
        fmt.Printf("constant 1");
    } else if Hop_ObjIsPi(pObj) {
        fmt.Printf("PI");
    } else {
        fmt.Printf("AND( %p%s, %p%s )", Hop_ObjFanin0(pObj), (Hop_ObjFaninC0(pObj)? "\'" : " "), Hop_ObjFanin1(pObj), (Hop_ObjFaninC1(pObj)? "\'" : " ")); }
    fmt.Printf(" (refs = %3d)", Hop_ObjRefs(pObj));
}

/**Function*************************************************************

  Synopsis    [Prints node in HAIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hop_ManPrintVerbose(p *Hop_Man_t, fHaig int32) {
    var vNodes *Vec_Ptr_t;
    var pObj *Hop_Obj_t;
    var i int32;
    fmt.Printf("PIs: ");
    // inlining of macro Hop_ManForEachPi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vPis)) && (((pObj) = (Hop_Obj_t *)Vec_PtrEntry(p.vPis, i)), 1); i++  {
        fmt.Printf(" %p", pObj);
    }
    fmt.Printf("\n");
    vNodes = Hop_ManDfs( p );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pObj) = (Hop_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        Hop_ObjPrintVerbose( pObj, fHaig );
fmt.Printf("\n");
    }
    fmt.Printf("\n");
    Vec_PtrFree( vNodes );
}

/**Function*************************************************************

  Synopsis    [Writes the AIG into the BLIF file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hop_ManDumpBlif(p *Hop_Man_t, pFileName string) {
    var pFile *os.File;
    var vNodes *Vec_Ptr_t;
    var pObj *Hop_Obj_t;
    var pConst1 *Hop_Obj_t = nil;
    var i, nDigits, Counter int32;
    if Hop_ManPoNum(p) == 0 {
        fmt.Printf("Hop_ManDumpBlif(): AIG manager does not have POs.\n");
        return;
    }
    // collect nodes in the DFS order
    vNodes = Hop_ManDfs( p );
    // assign IDs to objects
    Hop_ManConst1(p).pData = (void *)(ABC_PTRUINT_T)Counter++;
    // inlining of macro Hop_ManForEachPi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vPis)) && (((pObj) = (Hop_Obj_t *)Vec_PtrEntry(p.vPis, i)), 1); i++  {
        pObj.pData = (void *)(ABC_PTRUINT_T)Counter++;
    }
    // inlining of macro Hop_ManForEachPo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vPos)) && (((pObj) = (Hop_Obj_t *)Vec_PtrEntry(p.vPos, i)), 1); i++  {
        pObj.pData = (void *)(ABC_PTRUINT_T)Counter++;
    }
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pObj) = (Hop_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        pObj.pData = (void *)(ABC_PTRUINT_T)Counter++;
    }
    nDigits = Hop_Base10Log( Counter );
    // write the file
    pFile = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    fmt.Fprintf(pFile, "# BLIF file written by procedure Hop_ManDumpBlif() in ABC\n");
    fmt.Fprintf(pFile, "# http://www.eecs.berkeley.edu/~alanmi/abc/\n");
    fmt.Fprintf(pFile, ".model test\n");
    // write PIs
    fmt.Fprintf(pFile, ".inputs");
    // inlining of macro Hop_ManForEachPi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vPis)) && (((pObj) = (Hop_Obj_t *)Vec_PtrEntry(p.vPis, i)), 1); i++  {
        fmt.Fprintf(pFile, " n%0*d", nDigits, int32((ABC_PTRUINT_T))pObj.pData);
    }
    fmt.Fprintf(pFile, "\n");
    // write POs
    fmt.Fprintf(pFile, ".outputs");
    // inlining of macro Hop_ManForEachPo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vPos)) && (((pObj) = (Hop_Obj_t *)Vec_PtrEntry(p.vPos, i)), 1); i++  {
        fmt.Fprintf(pFile, " n%0*d", nDigits, int32((ABC_PTRUINT_T))pObj.pData);
    }
    fmt.Fprintf(pFile, "\n");
    // write nodes
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pObj) = (Hop_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        fmt.Fprintf(pFile, ".names n%0*d n%0*d n%0*d\n", nDigits, int32((ABC_PTRUINT_T))Hop_ObjFanin0(pObj).pData, nDigits, int32((ABC_PTRUINT_T))Hop_ObjFanin1(pObj).pData, nDigits, int32((ABC_PTRUINT_T))pObj.pData);
        fmt.Fprintf(pFile, "%d%d 1\n", !Hop_ObjFaninC0(pObj), !Hop_ObjFaninC1(pObj));
    }
    // write POs
    // inlining of macro Hop_ManForEachPo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vPos)) && (((pObj) = (Hop_Obj_t *)Vec_PtrEntry(p.vPos, i)), 1); i++  {
        fmt.Fprintf(pFile, ".names n%0*d n%0*d\n", nDigits, int32((ABC_PTRUINT_T))Hop_ObjFanin0(pObj).pData, nDigits, int32((ABC_PTRUINT_T))pObj.pData);
        fmt.Fprintf(pFile, "%d 1\n", !Hop_ObjFaninC0(pObj));
        if Hop_ObjIsConst1(Hop_ObjFanin0(pObj)) {
            pConst1 = Hop_ManConst1(p); }
    }
    if pConst1 != nil {
        fmt.Fprintf(pFile, ".names n%0*d\n 1\n", nDigits, int32((ABC_PTRUINT_T))pConst1.pData); }
    fmt.Fprintf(pFile, ".end\n\n");
    _ = (pFile).Close();
    Vec_PtrFree( vNodes );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

