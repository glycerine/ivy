package abci

import (
	"fmt"
	"time"
)

/**CFile****************************************************************

  FileName    [abcDressw.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Network and node package.]

  Synopsis    [Transfers names from one netlist to the other.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: abcDressw.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "base/abc/abc.h"
//#include "aig/aig/aig.h"
//#include "proof/dch/dch.h"
//#include "aig/gia/giaAig.h"

//ABC_NAMESPACE_IMPL_START


/*
  Procedure Abc_NtkDressComputeEquivs() implemented in this file computes 
  equivalence classes of objects of the two networks (pNtk1 and pNtk2).

  It is possible that pNtk1 is the network before synthesis and pNtk2 is the
  network after synthesis.  The equiv classes of nodes from these networks 
  can be used to transfer the names from pNtk1 to pNtk2, or vice versa. 

  The above procedure returns the array (Vec_Ptr_t) of integer arrays (Vec_Int_t).
  Each of the integer arrays contains entries of one equivalence class.
  Each entry (EquivId) contains the following information: 
  (1) object ID, which is a number 'num', such that 0 <= 'num' < MaxId 
      where MaxId is the largest ID of nodes in a network
  (2) the polarity of the node, which is a binary number, 0 or 1, giving
      the node's value when pattern (000...0) is applied to the inputs
  (3) the number of the network, 0 or 1, which stands for pNtk1 and pNtk2, respectively
  The first array in the array of arrays is empty, or contains nodes that 
  are equivalent to a constant (if such nodes appear in the network).

  Given EquivID defined above, use the APIs below to get its components.
*/

// declarations to be added to the application code
// extern int Abc_ObjEquivId2ObjId( int EquivId );

// extern int Abc_ObjEquivId2Polar( int EquivId );

// extern int Abc_ObjEquivId2NtkId( int EquivId );


// definition that may remain in this file
func Abc_ObjEquivId2ObjId(EquivId int32) int32 { return  EquivId >> 2;      }
func Abc_ObjEquivId2Polar(EquivId int32) int32 { return (EquivId >> 1) & 1; }
func Abc_ObjEquivId2NtkId(EquivId int32) int32 { return  EquivId       & 1; }

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// extern Aig_Man_t *  Abc_NtkToDar( Abc_Ntk_t * pNtk, int fExors, int fRegisters );

// extern void         Dch_ComputeEquivalences( Aig_Man_t * pAig, Dch_Pars_t * pPars );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Creates the dual-output miter.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManCreateDualOutputMiter(p1 *Aig_Man_t, p2 *Aig_Man_t) *Aig_Man_t {
    var pNew *Aig_Man_t;
    var pObj *Aig_Obj_t;
    var i int32;
    assert(Aig_ManCiNum(p1) == Aig_ManCiNum(p2), "Aig_ManCiNum(p1) == Aig_ManCiNum(p2)");
    assert(Aig_ManCoNum(p1) == Aig_ManCoNum(p2), "Aig_ManCoNum(p1) == Aig_ManCoNum(p2)");
    pNew = Aig_ManStart( Aig_ManObjNumMax(p1) + Aig_ManObjNumMax(p2) );
    // add first AIG
    Aig_ManConst1(p1).pData = Aig_ManConst1(pNew);
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p1.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p1.vCis, i)), 1); i++  {
        pObj.pData = Aig_ObjCreateCi( pNew );
    }
    // inlining of macro Aig_ManForEachNode
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p1.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p1.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
        pObj.pData = Aig_And( pNew, Aig_ObjChild0Copy(pObj), Aig_ObjChild1Copy(pObj) ); }
    }
    // add second AIG
    Aig_ManConst1(p2).pData = Aig_ManConst1(pNew);
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p2.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p2.vCis, i)), 1); i++  {
        pObj.pData = Aig_ManCi( pNew, i );
    }
    // inlining of macro Aig_ManForEachNode
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p2.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p2.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
        pObj.pData = Aig_And( pNew, Aig_ObjChild0Copy(pObj), Aig_ObjChild1Copy(pObj) ); }
    }
    // add the outputs
    for  i = 0; i < Aig_ManCoNum(p1); i++  {
        Aig_ObjCreateCo( pNew, Aig_ObjChild0Copy(Aig_ManCo(p1, i)) );
        Aig_ObjCreateCo( pNew, Aig_ObjChild0Copy(Aig_ManCo(p2, i)) );
    }
    Aig_ManCleanup( pNew );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Sets polarity attribute of each object in the network.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDressMapSetPolarity(pNtk *Abc_Ntk_t) {
    var pObj, pAnd *Abc_Obj_t;
    var i int32;
    // each node refers to the the strash copy whose polarity is set
    // inlining of macro Abc_NtkForEachObj
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {            if (pObj) == nil {} else {
        if (pAnd = Abc_ObjRegular(pObj.pCopy)) && Abc_ObjType(pAnd) != ABC_OBJ_NONE { // strashed object is present and legal
            pObj.fPhase = pAnd.fPhase ^ Abc_ObjIsComplement(pObj.pCopy); }
    }
    }
}

/**Function*************************************************************

  Synopsis    [Create mapping of node IDs of pNtk into equiv classes of pMiter.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDressMapClasses(pMiter *Aig_Man_t, pNtk *Abc_Ntk_t) *Vec_Int_t {
    var vId2Lit *Vec_Int_t;
    var pObj, pAnd *Abc_Obj_t;
    var pObjMan, pObjMiter, pObjRepr *Aig_Obj_t;
    var i int32;
    vId2Lit = Vec_IntAlloc( 0 );
    Vec_IntFill( vId2Lit, Abc_NtkObjNumMax(pNtk), -1 );
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pObj) == nil || !Abc_ObjIsNode(pObj) {} else {
        // get the pointer to the miter node corresponding to pObj
        if (pAnd = Abc_ObjRegular(pObj.pCopy)) && Abc_ObjType(pAnd) != ABC_OBJ_NONE &&          // strashed node is present and legal
             (pObjMan = Aig_Regular((Aig_Obj_t *)pAnd.pCopy)) && Aig_ObjType(pObjMan) != AIG_OBJ_NONE &&       // AIG node is present and legal
             (pObjMiter = Aig_Regular((Aig_Obj_t *)pObjMan.pData)) && Aig_ObjType(pObjMiter) != AIG_OBJ_NONE { // miter node is present and legal
        
            // get the representative of the miter node
            pObjRepr = Aig_ObjRepr( pMiter, pObjMiter );
            if pObjRepr != nil {
                pObjRepr = pObjRepr;
            } else {
                pObjRepr = pObjMiter;
            }
            // map pObj (whose ID is i) into the repr node ID (i.e. equiv class)
            Vec_IntWriteEntry( vId2Lit, i, Aig_ObjId(pObjRepr) );
        }
    }
    }
    return vId2Lit;
}

/**Function*************************************************************

  Synopsis    [Returns the vector of given equivalence class of objects.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_ObjDressClass(vRes *Vec_Ptr_t, vClass2Num *Vec_Int_t, Class int32) *Vec_Int_t {
    var ClassNumber int32;
    assert(Class > 0, "Class > 0");
    ClassNumber = Vec_IntEntry( vClass2Num, Class );
    assert(ClassNumber != 0, "ClassNumber != 0");
    if ClassNumber > 0 {
        return (Vec_Int_t *)Vec_PtrEntry( vRes, ClassNumber ); } // previous class
    // create new class
    Vec_IntWriteEntry( vClass2Num, Class, Vec_PtrSize(vRes) );
    Vec_PtrPush( vRes, Vec_IntAlloc(4) );
    return (Vec_Int_t *)Vec_PtrEntryLast( vRes ); 
}

/**Function*************************************************************

  Synopsis    [Returns the ID of a node in an equivalence class.]

  Description [The ID is composed of three parts: object ID, followed
  by one bit telling the phase of this node, followed by one bit
  telling the network to which this node belongs.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_ObjDressMakeId(pNtk *Abc_Ntk_t, ObjId int32, iNtk int32) int32 {
    return (ObjId << 2) | (Abc_NtkObj(pNtk,ObjId).fPhase << 1) | iNtk;
}

/**Function*************************************************************

  Synopsis    [Computes equivalence classes of objects in pNtk1 and pNtk2.]

  Description [Internal procedure.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDressMapIds(pMiter *Aig_Man_t, pNtk1 *Abc_Ntk_t, pNtk2 *Abc_Ntk_t) *Vec_Ptr_t {
    var vRes *Vec_Ptr_t;
    var vId2Lit1, vId2Lit2, vCounts0, vCounts1, vClassC, vClass2Num *Vec_Int_t;
    var i, Class int32;
    // start the classes
    vRes = Vec_PtrAlloc( 1000 );
    // set polarity of the nodes
    Abc_NtkDressMapSetPolarity( pNtk1 );
    Abc_NtkDressMapSetPolarity( pNtk2 );
    // create mapping of node IDs of pNtk1/pNtk2 into the IDs of equiv classes of pMiter
    vId2Lit1 = Abc_NtkDressMapClasses( pMiter, pNtk1 );
    vId2Lit2 = Abc_NtkDressMapClasses( pMiter, pNtk2 );
    // count the number of nodes in each equivalence class
    vCounts0 = Vec_IntStart( Aig_ManObjNumMax(pMiter) );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vId2Lit1)) && (((Class) = Vec_IntEntry(vId2Lit1, i)), 1); i++  {
        if Class >= 0 {
            Vec_IntAddToEntry( vCounts0, Class, 1 ); }
    }
    vCounts1 = Vec_IntStart( Aig_ManObjNumMax(pMiter) );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vId2Lit2)) && (((Class) = Vec_IntEntry(vId2Lit2, i)), 1); i++  {
        if Class >= 0 {
            Vec_IntAddToEntry( vCounts1, Class, 1 ); }
    }
    // get the costant class
    vClassC = Vec_IntAlloc( 100 );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vId2Lit1)) && (((Class) = Vec_IntEntry(vId2Lit1, i)), 1); i++  {
        if Class == 0 {
            Vec_IntPush( vClassC, Abc_ObjDressMakeId(pNtk1, i, 0) ); }
    }
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vId2Lit2)) && (((Class) = Vec_IntEntry(vId2Lit2, i)), 1); i++  {
        if Class == 0 {
            Vec_IntPush( vClassC, Abc_ObjDressMakeId(pNtk2, i, 1) ); }
    }
    Vec_PtrPush( vRes, vClassC );
    // map repr node IDs into class numbers
    vClass2Num = Vec_IntAlloc( 0 );
    Vec_IntFill( vClass2Num, Aig_ManObjNumMax(pMiter), -1 );
    // keep classes having at least one element from pNtk1 and one from pNtk2
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vId2Lit1)) && (((Class) = Vec_IntEntry(vId2Lit1, i)), 1); i++  {
        if Class > 0 && Vec_IntEntry(vCounts0, Class) && Vec_IntEntry(vCounts1, Class) {
            Vec_IntPush( Abc_ObjDressClass(vRes, vClass2Num, Class), Abc_ObjDressMakeId(pNtk1, i, 0) ); }
    }
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vId2Lit2)) && (((Class) = Vec_IntEntry(vId2Lit2, i)), 1); i++  {
        if Class > 0 && Vec_IntEntry(vCounts0, Class) && Vec_IntEntry(vCounts1, Class) {
            Vec_IntPush( Abc_ObjDressClass(vRes, vClass2Num, Class), Abc_ObjDressMakeId(pNtk2, i, 1) ); }
    }
    // package them accordingly
    Vec_IntFree( vClass2Num );
    Vec_IntFree( vCounts0 );
    Vec_IntFree( vCounts1 );
    Vec_IntFree( vId2Lit1 );
    Vec_IntFree( vId2Lit2 );
    return vRes;
}

/**Function*************************************************************

  Synopsis    [Alternative way to compute equivalences.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dch_ComputeEquivalences2(pMiter *Aig_Man_t, pPars *Dch_Pars_t) {
    extern Gia_Man_t * Cec4_ManSimulateTest3( Gia_Man_t * p, int nBTLimit, int fVerbose );
    var pGia *Gia_Man_t  = Gia_ManFromAigSimple(pMiter);
    var pNew *Gia_Man_t  = Cec4_ManSimulateTest3( pGia, pPars.nBTLimit, pPars.fVerbose );
    var i, k int32;
    // inlining of macro ABC_FREE
    ((pMiter.pReprs) ? (free((char *) (pMiter.pReprs)), (pMiter.pReprs) = 0) : 0);
    // inlining of macro ABC_CALLOC
    pMiter.pReprs = ((Aig_Obj_t * *) calloc((size_t)(Aig_ManObjNumMax(pMiter)), sizeof(Aig_Obj_t *)));
    // inlining of macro Gia_ManForEachClass
    for  i = 1; i < Gia_ManObjNum(pGia); i++  { if !Gia_ObjIsHead(pGia, i) {} else {
        // inlining of macro Gia_ClassForEachObj1
        for  assert(Gia_ObjIsHead(pGia, i), "Gia_ObjIsHead(pGia, i)"), k = Gia_ObjNext(pGia, i); k > 0; k = Gia_ObjNext(pGia, k)  {
            pMiter.pReprs[k] = Aig_ManObj( pMiter, i );
        } }
    }
    Gia_ManStop( pGia );
    Gia_ManStop( pNew );
}

/**Function*************************************************************

  Synopsis    [Computes equivalence classes of objects in pNtk1 and pNtk2.]

  Description [Returns the array (Vec_Ptr_t) of integer arrays (Vec_Int_t).
  Each of the integer arrays contains entries of one equivalence class.
  Each entry contains the following information: the network number (0/1),
  the polarity (0/1) and the object ID in the the network (0 <= num < MaxId)
  where MaxId is the largest number of an ID of an object in that network.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDressComputeEquivs(pNtk1 *Abc_Ntk_t, pNtk2 *Abc_Ntk_t, nConflictLimit int32, fVerbose int32) *Vec_Ptr_t {
    var Pars Dch_Pars_t;
    var pPars *Dch_Pars_t = &Pars;
    var pAig1, pAig2 *Abc_Ntk_t;
    var pMan1, pMan2, pMiter *Aig_Man_t;
    var vRes *Vec_Ptr_t;
    assert(!Abc_NtkIsStrash(pNtk1), "!Abc_NtkIsStrash(pNtk1)");
    assert(!Abc_NtkIsStrash(pNtk2), "!Abc_NtkIsStrash(pNtk2)");
    // convert network into AIG
    pAig1 = Abc_NtkStrash( pNtk1, 1, 1, 0 );
    pAig2 = Abc_NtkStrash( pNtk2, 1, 1, 0 );
    pMan1 = Abc_NtkToDar( pAig1, 0, 0 );
    pMan2 = Abc_NtkToDar( pAig2, 0, 0 );
    // derive the miter
    pMiter = Aig_ManCreateDualOutputMiter( pMan1, pMan2 );
    // set up parameters for SAT sweeping
    Dch_ManSetDefaultParams( pPars );
    pPars.nBTLimit = nConflictLimit;
    pPars.fVerbose = fVerbose;
    // perform SAT sweeping
    Dch_ComputeEquivalences2( pMiter, pPars );
    // now, pMiter is annotated with the equivl class info
    // convert this info into the resulting array
    vRes = Abc_NtkDressMapIds( pMiter, pNtk1, pNtk2 );
    Aig_ManStop( pMiter );
    Aig_ManStop( pMan1 );
    Aig_ManStop( pMan2 );
    Abc_NtkDelete( pAig1 );
    Abc_NtkDelete( pAig2 );
    return vRes;
}

/**Function*************************************************************

  Synopsis    [Prints information about node equivalences.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDressPrintEquivs(vRes *Vec_Ptr_t) {
    var vClass *Vec_Int_t;
    var i, k, Entry int32;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vRes)) && (((vClass) = (Vec_Int_t *)Vec_PtrEntry(vRes, i)), 1); i++  {
        fmt.Printf("Class %5d : ", i);
        fmt.Printf("Num =%5d    ", Vec_IntSize(vClass));
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vClass)) && (((Entry) = Vec_IntEntry(vClass, k)), 1); k++  {
            fmt.Printf("%5d%c%d ", Abc_ObjEquivId2ObjId(Entry), Abc_ObjEquivId2Polar(Entry)? '-':'+', Abc_ObjEquivId2NtkId(Entry));
        }
        fmt.Printf("\n");
    }
}

/**Function*************************************************************

  Synopsis    [Prints information about node equivalences.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDressPrintStats(vRes *Vec_Ptr_t, nNodes0 int32, nNodes1 int32, Time abctime) {
    var vClass *Vec_Int_t;
    var i, k, Entry int32;
    var NegAll [2]int32 ;
    var PosAll [2]int32 ;
    var PairsAll int32;
    var PairsOne int32;
    var Pos, Neg [2]int32;
    // count the number of equivalences in each class
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vRes)) && (((vClass) = (Vec_Int_t *)Vec_PtrEntry(vRes, i)), 1); i++  {
        Pos[0] = Pos[1] = 0;
        Neg[0] = Neg[1] = 0;
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vClass)) && (((Entry) = Vec_IntEntry(vClass, k)), 1); k++  {
            if Abc_ObjEquivId2NtkId(Entry) != 0 {
                if Abc_ObjEquivId2Polar(Entry) != 0 {
                    Neg[1]++; // negative polarity in network 1
                } else {
                    Pos[1]++; } // positive polarity in network 1
            
            } else {
                if Abc_ObjEquivId2Polar(Entry) != 0 {
                    Neg[0]++; // negative polarity in network 0
                } else {
                    Pos[0]++; } // positive polarity in network 0
            }
        }
        PosAll[0] += Pos[0]; // total positive polarity in network 0
        PosAll[1] += Pos[1]; // total positive polarity in network 1
        NegAll[0] += Neg[0]; // total negative polarity in network 0
        NegAll[1] += Neg[1]; // total negative polarity in network 1

        // assuming that the name can be transferred to only one node
        PairsAll += Abc_MinInt(Neg[0] + Pos[0], Neg[1] + Pos[1]);
        PairsOne += Abc_MinInt(Neg[0], Neg[1]) + Abc_MinInt(Pos[0], Pos[1]);
    }
    fmt.Printf("Total number of equiv classes                = %7d.\n", Vec_PtrSize(vRes));
    fmt.Printf("Participating nodes from both networks       = %7d.\n", NegAll[0]+PosAll[0]+NegAll[1]+PosAll[1]);
    fmt.Printf("Participating nodes from the first network   = %7d. (%7.2f %% of nodes)\n", NegAll[0]+PosAll[0], 100.0*(NegAll[0]+PosAll[0])/(nNodes0+1));
    fmt.Printf("Participating nodes from the second network  = %7d. (%7.2f %% of nodes)\n", NegAll[1]+PosAll[1], 100.0*(NegAll[1]+PosAll[1])/(nNodes1+1));
    fmt.Printf("Node pairs (any polarity)                    = %7d. (%7.2f %% of names can be moved)\n", PairsAll, 100.0*PairsAll/(nNodes0+1));
    fmt.Printf("Node pairs (same polarity)                   = %7d. (%7.2f %% of names can be moved)\n", PairsOne, 100.0*PairsOne/(nNodes0+1));
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("Total runtime"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Time)))/(float64(int64(time.Second))));;
}

/**Function*************************************************************

  Synopsis    [Transfers IDs from pNtk1 to pNtk2 using equivalence classes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDress2Transfer(pNtk0 *Abc_Ntk_t, pNtk1 *Abc_Ntk_t, vRes *Vec_Ptr_t, fVerbose int32) {
    var vClass *Vec_Int_t;
    var pObj0, pObj1 *Abc_Obj_t;
    var i, k, fComp0, fComp1, Entry int32;
    var CounterInv, Counter int32;
    var pName string;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vRes)) && (((vClass) = (Vec_Int_t *)Vec_PtrEntry(vRes, i)), 1); i++  {
        pObj0 = pObj1 = nil;
        fComp0 = fComp1 = 0;
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vClass)) && (((Entry) = Vec_IntEntry(vClass, k)), 1); k++  {
            if Abc_ObjEquivId2NtkId(Entry) != 0 {
                pObj1 = Abc_NtkObj( pNtk1, Abc_ObjEquivId2ObjId(Entry) );
                fComp1 = Abc_ObjEquivId2Polar(Entry);
            
            } else {
                pObj0 = Abc_NtkObj( pNtk0, Abc_ObjEquivId2ObjId(Entry) );
                fComp0 = Abc_ObjEquivId2Polar(Entry);
            }
        }
        if pObj0 == nil || pObj1 == nil {
            continue; }
        // if the node already has a name, quit
        pName = Nm_ManFindNameById( pNtk0.pManName, pObj0.Id );
        if pName != nil {
            continue; }
        // if the other node has no name, quit
        pName = Nm_ManFindNameById( pNtk1.pManName, pObj1.Id );
        if pName == nil {
            continue; }
        // assign name
        if fComp0 ^ fComp1 {
            Abc_ObjAssignName( pObj0, pName, "_inv" );
            CounterInv++;
        
        } else {
            Abc_ObjAssignName( pObj0, pName, nil );
            Counter++;
        }
    }
    if fVerbose != 0 {
        fmt.Printf("Total number of names assigned  = %5d. (Dir = %5d. Compl = %5d.)\n", Counter + CounterInv, Counter, CounterInv);
    }
}

/**Function*************************************************************

  Synopsis    [Transfers names from pNtk1 to pNtk2.]

  Description [Internally calls new procedure for mapping node IDs of
  both networks into the shared equivalence classes.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDress2(pNtk1 *Abc_Ntk_t, pNtk2 *Abc_Ntk_t, nConflictLimit int32, fVerbose int32) {
    var vRes *Vec_Ptr_t;
    var clk abctime  = Abc_Clock();
    vRes = Abc_NtkDressComputeEquivs( pNtk1, pNtk2, nConflictLimit, fVerbose );
//    Abc_NtkDressPrintEquivs( vRes );
    Abc_NtkDressPrintStats( vRes, Abc_NtkNodeNum(pNtk1), Abc_NtkNodeNum(pNtk1), Abc_Clock() - clk );
    Abc_NtkDress2Transfer( pNtk1, pNtk2, vRes, fVerbose );
    Vec_VecFree( (Vec_Vec_t *)vRes );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

