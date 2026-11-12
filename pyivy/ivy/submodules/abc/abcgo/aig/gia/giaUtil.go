package gia

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [giaUtil.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [Various utilities.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaUtil.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"
//#include "base/main/mainInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////
 



////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Creates a sequence or random numbers.]

  Description []
               
  SideEffects []

  SeeAlso     [http://www.codeproject.com/KB/recipes/SimpleRNG.aspx]

***********************************************************************/





/**Function*************************************************************

  Synopsis    [Creates random info for the primary inputs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/



/**Function*************************************************************

  Synopsis    [Returns the time stamp.]

  Description [The file should be closed.]

  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Returns the composite name of the file.]

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

  Synopsis    [Sets phases of the internal nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Sets phases of the internal nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Sets phases of the internal nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Sets phases of the internal nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Sets phases of the internal nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Sets phases of the internal nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Sets phases of the internal nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Cleans the value.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Cleans the value.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Sets the phase of one object.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Sets phases of the internal nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/



/**Function*************************************************************

  Synopsis    [Sets phases of the internal nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Sets phases of the internal nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Returns the number of COs whose value is 1.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Prepares copies for the model.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Prepares copies for the model.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Assigns levels.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/




/**Function*************************************************************

  Synopsis    [Assigns levels using CI level information.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/



/**Function*************************************************************

  Synopsis    [Compute reverse levels.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Compute required levels.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Compute slacks measured using the number of AIG levels.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Assigns levels.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Assigns references.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Assigns references.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Computes the maximum frontier size.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Collects PO Ids into one array.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/



/**Function*************************************************************

  Synopsis    [Returns 1 if the node is the root of MUX or EXOR/NEXOR.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/



/**Function*************************************************************

  Synopsis    [Recognizes what nodes are inputs of the EXOR.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Recognizes what nodes are control and data inputs of a MUX.]

  Description [If the node is a MUX, returns the control variable C.
  Assigns nodes T and E to be the then and else variables of the MUX. 
  Node C is never complemented. Nodes T and E can be complemented.
  This function also recognizes EXOR/NEXOR gates as MUXes.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/




/**Function*************************************************************

  Synopsis    [Dereferences the node's MFFC.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [References the node's MFFC.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/



/**Function*************************************************************

  Synopsis    [References the node's MFFC.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Returns the number of internal nodes in the MFFC.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/



/**Function*************************************************************

  Synopsis    [Returns the number of internal nodes in the MFFC.]

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

  Synopsis    [Returns 1 if AIG has dangling nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Returns 1 if AIG has dangling nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Returns 1 if AIG has dangling nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Verbose printing of the AIG node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Complements the constraint outputs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/



/**Function*************************************************************

  Synopsis    [Testing the speedup due to grouping POs into batches.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCollectObjs_rec(p *Gia_Man_t, iObjId int32, vObjs *Vec_Int_t, Limit int32) {
    var pObj *Gia_Obj_t;
    if Vec_IntSize(vObjs) == Limit {
        return; }
    if Gia_ObjIsTravIdCurrentId(p, iObjId) {
        return; }
    Gia_ObjSetTravIdCurrentId(p, iObjId);
    pObj = Gia_ManObj( p, iObjId );
    if Gia_ObjIsAnd(pObj) {
        Gia_ManCollectObjs_rec( p, Gia_ObjFaninId0p(p, pObj), vObjs, Limit );
        if Vec_IntSize(vObjs) == Limit {
            return; }
        Gia_ManCollectObjs_rec( p, Gia_ObjFaninId1p(p, pObj), vObjs, Limit );
        if Vec_IntSize(vObjs) == Limit {
            return; }
    }
    Vec_IntPush( vObjs, iObjId );
}
func Gia_ManComputePoTruthTables(p *Gia_Man_t, nBytesMax int32) *uint32 {
    var nVars int32  = Gia_ManPiNum(p);
    var nTruthWords int32  = Abc_TruthWordNum( nVars );
    var nTruths int32  = nBytesMax / (sizeof(unsigned) * nTruthWords);
    var nTotalNodes, nRounds int32;
    var vObjs *Vec_Int_t;
    var pObj *Gia_Obj_t;
    var clk abctime  = Abc_Clock();
    var i int32;
    fmt.Printf("Var = %d. Words = %d. Truths = %d.\n", nVars, nTruthWords, nTruths);
    vObjs = Vec_IntAlloc( nTruths );
    Gia_ManIncrementTravId( p );
    Gia_ManForEachPo( p, pObj, i )
    {
        Gia_ManCollectObjs_rec( p, Gia_ObjFaninId0p(p, pObj), vObjs, nTruths );
        if Vec_IntSize(vObjs) == nTruths {
            nRounds++;
//            printf( "%d ", i );
            nTotalNodes += Vec_IntSize( vObjs );
            Vec_IntClear( vObjs );
            Gia_ManIncrementTravId( p );
        }
    }
//    printf( "\n" );
    nTotalNodes += Vec_IntSize( vObjs );
    Vec_IntFree( vObjs );

    fmt.Printf("Rounds = %d. Objects = %d. Total = %d.   ", nRounds, Gia_ManObjNum(p), nTotalNodes);
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );

    return nil;
}


/**Function*************************************************************

  Synopsis    [Returns 1 if the manager are structural identical.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCompare(p1 *Gia_Man_t, p2 *Gia_Man_t) int32 {
    var pObj1, pObj2 *Gia_Obj_t;
    var i int32;
    if Gia_ManObjNum(p1) != Gia_ManObjNum(p2) {
        fmt.Printf("AIGs have different number of objects.\n");
        return 0;
    }
    Gia_ManCleanValue( p1 );
    Gia_ManCleanValue( p2 );
    Gia_ManForEachObj( p1, pObj1, i )
    {
        pObj2 = Gia_ManObj( p2, i );
        if memcmp( pObj1, pObj2, sizeof(Gia_Obj_t) ) {
            fmt.Printf("Objects %d are different.\n", i);
            return 0;
        }
        if p1.pReprs && p2.pReprs {
            if memcmp( &p1.pReprs[i], &p2.pReprs[i], sizeof(Gia_Rpr_t) ) {
                fmt.Printf("Representatives of objects %d are different.\n", i);
                return 0;
            }
        }
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Marks nodes that appear as faninis of other nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/



/**Function*************************************************************

  Synopsis    [Swaps PO number 0 with PO number i.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Save/load value from file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/




/**Function*************************************************************

  Synopsis    [Returns the array containing the first fanout of each object.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Returns 1 if AIG has choices.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManHasChoices_very_old(p *Gia_Man_t) int32 {
    var pObj *Gia_Obj_t;
    var i, Counter1, Counter2 int32;
    var nFailNoRepr int32;
    var nFailHaveRepr int32;
    var nChoiceNodes int32;
    var nChoices int32;
    if p.pReprs == nil || p.pNexts == nil {
        return 0; }
    // check if there are any representatives
    Gia_ManForEachObj( p, pObj, i )
    {
        if Gia_ObjReprObj( p, Gia_ObjId(p, pObj) ) {
//            printf( "%d ", i );
            Counter1++;
        }
//        if ( Gia_ObjNext( p, Gia_ObjId(p, pObj) ) )
//            Counter2++;
    }
//    printf( "\n" );
    Gia_ManForEachObj( p, pObj, i )
    {
//        if ( Gia_ObjReprObj( p, Gia_ObjId(p, pObj) ) )
//            Counter1++;
        if Gia_ObjNext( p, Gia_ObjId(p, pObj) ) {
//            printf( "%d ", i );
            Counter2++;
        }
    }
//    printf( "\n" );
    if Counter1 == 0 {
        fmt.Printf("Warning: AIG has repr data-strucure but not reprs.\n");
        return 0;
    }
    fmt.Printf("%d nodes have reprs.\n", Counter1);
    fmt.Printf("%d nodes have nexts.\n", Counter2);
    // check if there are any internal nodes without fanout
    // make sure all nodes without fanout have representatives
    // make sure all nodes with fanout have no representatives
    p.Refs = nil;
    Gia_ManCreateRefs( p );
    Gia_ManForEachAnd( p, pObj, i )
    {
        if Gia_ObjRefNum(p, pObj) == 0 {
            if Gia_ObjReprObj( p, Gia_ObjId(p, pObj) ) == nil {
                nFailNoRepr++;
            } else {
                nChoices++; }
        
        } else {
            if Gia_ObjReprObj( p, Gia_ObjId(p, pObj) ) != nil {
                nFailHaveRepr++; }
            if Gia_ObjNextObj( p, Gia_ObjId(p, pObj) ) != nil {
                nChoiceNodes++; }
        }
        if Gia_ObjReprObj( p, i ) {
            assert(Gia_ObjRepr(p, i) < i, "Gia_ObjRepr(p, i) < i"); }
    }
    if nChoices == 0 {
        return 0; }
    if nFailNoRepr != 0 {
        fmt.Printf("Gia_ManHasChoices_very_old(): Error: %d internal nodes have no fanout and no repr.\n", nFailNoRepr);
//        return 0;
    }
    if nFailHaveRepr != 0 {
        fmt.Printf("Gia_ManHasChoices_very_old(): Error: %d internal nodes have both fanout and repr.\n", nFailHaveRepr);
//        return 0;
    }
//    printf( "Gia_ManHasChoices_very_old(): AIG has %d choice nodes with %d choices.\n", nChoiceNodes, nChoices );
    return 1;
}


/**Function*************************************************************

  Synopsis    [Proving multi-output properties.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManGroupProve(pInit *Gia_Man_t, pCommLine string, nGroupSize int32, fVerbose int32) *Vec_Int_t {
    var pAbc *Abc_Frame_t  = Abc_FrameReadGlobalFrame();
    var p *Gia_Man_t  = Gia_ManDup( pInit );
    var pGroup *Gia_Man_t;
    var vOuts *Vec_Int_t;
    var vOutMap *Vec_Int_t;
    var vCexes *Vec_Ptr_t;
    var i, k, nGroupCur, nGroups int32;
    var clk, timeComm abctime;
    var timeStart abctime  = Abc_Clock();
    // pre-conditions
    assert(nGroupSize > 0, "nGroupSize > 0");
    assert(pCommLine != nil, "pCommLine != nil");
    assert(p.nConstrs == 0, "p.nConstrs == 0");
    Abc_Print( 1, "RUNNING MultiProve: Group size = %d. Command line = \"%s\".\n", nGroupSize, pCommLine );
    // create output map
    vOuts   = Vec_IntStartNatural( Gia_ManPoNum(p) );
    vOutMap = Vec_IntAlloc( Gia_ManPoNum(p) );
    vCexes  = Vec_PtrAlloc( Gia_ManPoNum(p) );
    nGroups = Gia_ManPoNum(p) / nGroupSize + int32(((Gia_ManPoNum(p) % nGroupSize) > 0));
    for  i = 0; i < nGroups; i++  {
        // derive the group
        if (i + 1) * nGroupSize < Gia_ManPoNum(p) {
            nGroupCur = nGroupSize;
        } else {
            nGroupCur = Gia_ManPoNum(p) - i * nGroupSize;
        }
        pGroup = Gia_ManDupCones( p, Vec_IntArray(vOuts) + i * nGroupSize, nGroupCur, 0 );
        Abc_Print( 1, "GROUP %4d : %4d <= PoId < %4d : ", i, i * nGroupSize, i * nGroupSize + nGroupCur );
        // set the current GIA
        Abc_FrameUpdateGia( pAbc, pGroup );
        // solve the group
        clk = Abc_Clock();
        Cmd_CommandExecute( pAbc, pCommLine );
        timeComm += Abc_Clock() - clk;
        // get the solution status
        if nGroupSize == 1 {
            Vec_IntPush( vOutMap, Abc_FrameReadProbStatus(pAbc) );
            Vec_PtrPush( vCexes, Abc_FrameReadCex(pAbc) );
        
        } else { // if ( nGroupSize > 1 )  
        
            var vStatusCur *Vec_Int_t  = Abc_FrameReadPoStatuses( pAbc );
            var vCexesCur *Vec_Ptr_t  = Abc_FrameReadCexVec( pAbc );
            assert(vStatusCur != nil, "vStatusCur != nil"); // only works for "bmc3" and "pdr"
//            assert( vCexesCur != NULL );
            for  k = 0; k < nGroupCur; k++  {
                Vec_IntPush( vOutMap, Vec_IntEntry(vStatusCur, k) );
                Vec_PtrPush( vCexes, vCexesCur ? Vec_PtrEntry(vCexesCur, k) : nil );
            }
        }
    }
    assert(Vec_PtrSize(vCexes) == Gia_ManPoNum(p), "Vec_PtrSize(vCexes) == Gia_ManPoNum(p)");
    assert(Vec_IntSize(vOutMap) == Gia_ManPoNum(p), "Vec_IntSize(vOutMap) == Gia_ManPoNum(p)");
    // set CEXes
    if Vec_PtrCountZero(vCexes) < Vec_PtrSize(vCexes) {
        Abc_FrameReplaceCexVec( pAbc, &vCexes );
    } else {  // there is no CEXes
        Vec_PtrFree( vCexes ); }
    // report the result
    Abc_Print( 1, "SUMMARY:  " );
    Abc_Print( 1, "Properties = %6d. ", Gia_ManPoNum(p) );
    Abc_Print( 1, "UNSAT = %6d. ",      Vec_IntCountEntry(vOutMap, 1) );
    Abc_Print( 1, "SAT = %6d. ",        Vec_IntCountEntry(vOutMap, 0) );
    Abc_Print( 1, "UNDEC = %6d. ",      Vec_IntCountEntry(vOutMap, -1) );
    Abc_Print( 1, "\n" );
    Abc_PrintTime( 1, "Command time", timeComm );
    Abc_PrintTime( 1, "Total time  ", Abc_Clock() - timeStart );
    // cleanup
    Vec_IntFree( vOuts );
    Gia_ManStop( p );
    return vOutMap;
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
const MAX_LUT_SIZE = 8
typedef struct Gia_MapLut_t_
{
    int        Type;          // node type: PI=1, PO=2, LUT=3
    int        Out;           // ID
    int        StartId;       // -1
    int        nFans;         // fanin count
    float      Delay;         // 0.0
    int        pFans[MAX_LUT_SIZE];  // fanin IDs
    unsigned   pTruth[MAX_LUT_SIZE<6?1:(1<<(MAX_LUT_SIZE-5))]; // the truth table
}
 Gia_MapLut_t;

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_AigerWriteLut(p *Gia_Man_t, pFileName string) {
    var pObj *Gia_Obj_t;
    var i, k, iFan, iLut int32;
    var LutSizeMax int32  = Gia_ManLutSizeMax( p );
    var nUints int32  = Abc_TruthWordNum(LutSizeMax);
    var nLuts int32  = 1 + Gia_ManCiNum(p) + Gia_ManCoNum(p) + Gia_ManLutNum(p);
    var pLuts *Gia_MapLut_t  = ABC_CALLOC( Gia_MapLut_t, nLuts );
    var vTruths *Vec_Wrd_t  = Vec_WrdStart( Gia_ManObjNum(p) );
    assert(LutSizeMax <= 6, "LutSizeMax <= 6");
    // set obj numbers
    // constant
    pLuts.Type = 3;
    memset( pLuts.pTruth, 0xFF, sizeof(unsigned) * nUints );
    Gia_ManFillValue(p);
    Gia_ManConst0(p).Value = pLuts[iLut].Out = Abc_Var2Lit( iLut, 0 );
    iLut++;
    // inputs
    Gia_ManForEachCi( p, pObj, i )
    {
        pLuts[iLut].Type = 1;
        memset( pLuts[iLut].pTruth, 0xAA, sizeof(unsigned) * nUints );
        pObj.Value = pLuts[iLut].Out = Abc_Var2Lit( iLut, 0 );
        iLut++;
    }
    // nodes
    Gia_ManForEachObj( p, pObj, i )
        if i && Gia_ObjIsLut(p, i) {
            pLuts[iLut].Type = 3;
            Gia_LutForEachFanin( p, i, iFan, k )
                pLuts[iLut].pFans[k] = Gia_ManObj(p, iFan).Value;
            pLuts[iLut].nFans = k;
            *(word *)pLuts[iLut].pTruth = Gia_LutComputeTruth6(p, i, vTruths);
            pObj.Value = pLuts[iLut].Out = Abc_Var2Lit( iLut, 0 );
            iLut++;
        }
    // outputs
    Gia_ManForEachCo( p, pObj, i )
    {
        pLuts[iLut].Type = 2;
        pLuts[iLut].pFans[0] = Gia_ObjFanin0(pObj).Value;
        if Gia_ObjFaninC0(pObj) ^ Gia_ObjIsConst0(Gia_ObjFanin0(pObj)) {
            memset( pLuts[iLut].pTruth, 0x55, sizeof(unsigned) * nUints );
        } else {
            memset( pLuts[iLut].pTruth, 0xAA, sizeof(unsigned) * nUints ); }
        pLuts[iLut].nFans = 1;
        pObj.Value = pLuts[iLut].Out = Abc_Var2Lit( iLut, 0 );
        iLut++;
    }
    assert(iLut == nLuts, "iLut == nLuts");
    // dump into a file
    {
        var pFile *os.File  = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
        if pFile == nil {
            fmt.Printf("Cannot open file \"%s\" for writing.\n", pFileName);
        } else {
            var nSize1 int32  = nLuts * sizeof(Gia_MapLut_t);
            var nSize2 int32  = fwrite( pLuts, 1, nSize1, pFile );
            assert(nSize1 == nSize2, "nSize1 == nSize2");
            fmt.Printf("Successfully dumped %d bytes of binary data.\n", nSize1);
        }
        _ = (pFile).Close();
    }
    pLuts = nil;
    Vec_WrdFree( vTruths );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_DumpLutSizeDistrib(p *Gia_Man_t, pFileName string) {
    var pTable *os.File  = os.OpenFile(pFileName, os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666);
    var i int32;
    var Counts [10]int32 ;
    Gia_ManForEachLut( p, i )
        if ( Gia_ObjLutSize(p, i) > 0 && Gia_ObjLutSize(p, i) < 10 )
            Counts[ Gia_ObjLutSize(p, i) ]++;
    fmt.Fprintf(pTable, "%s", p.pName);
    for  i = 1; i < 10; i++  {
        fmt.Fprintf(pTable, " %d", Counts[i]);
    }
    fmt.Fprintf(pTable, "\n");
    _ = (pTable).Close();
}

/**Function*************************************************************

  Synopsis    [Check if two logic cones have overlap.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/






/**Function*************************************************************

  Synopsis    [Count PIs with fanout.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Count POs driven by non-zero driver.]

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

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupWithMuxPos(p *Gia_Man_t) *Gia_Man_t {
    var vPoints *Vec_Int_t  = Vec_IntAlloc( 1000 );
    var vQuads *Vec_Int_t   = Vec_IntAlloc( 1000 );
    var vHeads *Vec_Bit_t   = Vec_BitStart( Gia_ManObjNum(p) );
    var vDatas *Vec_Bit_t   = Vec_BitStart( Gia_ManObjNum(p) );
    var pObj, pCtrl, pData0, pData1 *Gia_Obj_t;
    var pNew *Gia_Man_t  = Gia_ManDup( p ); var i, iObj int32;
    assert(Gia_ManRegNum(pNew) == 0, "Gia_ManRegNum(pNew) == 0");
    Gia_ManForEachAnd( pNew, pObj, i )
    {
        if !(Gia_ObjIsMuxType(pObj) != 0) {
            continue; }
        pCtrl  = Gia_ObjRecognizeMux( pObj, &pData1, &pData0 );
        pCtrl  = Gia_Regular(pCtrl);
        pData0 = Gia_Regular(pData0);
        pData1 = Gia_Regular(pData1);
        Vec_IntPushTwo( vQuads, Gia_ObjId(pNew, pObj),   Gia_ObjId(pNew, pCtrl)  );
        Vec_IntPushTwo( vQuads, Gia_ObjId(pNew, pData0), Gia_ObjId(pNew, pData1) );
        Vec_BitWriteEntry( vHeads, Gia_ObjId(pNew, pObj),   1 );
        Vec_BitWriteEntry( vDatas, Gia_ObjId(pNew, pData0), 1 );
        Vec_BitWriteEntry( vDatas, Gia_ObjId(pNew, pData1), 1 );
    }
    Gia_ManForEachCo( pNew, pObj, i )
        Gia_ObjFanin0(pObj).fMark0 = 1;
    for  i = 0; i < Vec_IntSize(vQuads)/4; i++  {
        var iObj int32    = Vec_IntEntry( vQuads, 4*i+0 );
        var iCtrl int32   = Vec_IntEntry( vQuads, 4*i+1 );
        var iData0 int32  = Vec_IntEntry( vQuads, 4*i+2 );
        var iData1 int32  = Vec_IntEntry( vQuads, 4*i+3 );
        if (Vec_BitEntry(vHeads, iObj)   && Vec_BitEntry(vDatas, iObj))   || 
             (Vec_BitEntry(vHeads, iData0) && Vec_BitEntry(vDatas, iData0)) || 
             (Vec_BitEntry(vHeads, iData1) && Vec_BitEntry(vDatas, iData1)) {
            var pObj *Gia_Obj_t    = Gia_ManObj( p, iObj );
            var pCtrl *Gia_Obj_t   = Gia_ManObj( p, iCtrl );
            var pData0 *Gia_Obj_t  = Gia_ManObj( p, iData0 );
            var pData1 *Gia_Obj_t  = Gia_ManObj( p, iData1 );
            if Gia_ObjIsAnd(pObj)   && !pObj.fMark0 {  Vec_IntPush( vPoints, iObj   ); }
            if Gia_ObjIsAnd(pCtrl)  && !pCtrl.fMark0 {  Vec_IntPush( vPoints, iCtrl  ); }
            if Gia_ObjIsAnd(pData0) && !pData0.fMark0 {  Vec_IntPush( vPoints, iData0 ); }
            if Gia_ObjIsAnd(pData1) && !pData1.fMark0 {  Vec_IntPush( vPoints, iData1 ); }
        }
    }
    Gia_ManCleanMark0( pNew );
    Vec_IntUniqify( vPoints );
    Vec_IntForEachEntry( vPoints, iObj, i )
        Gia_ManAppendCo( pNew, Abc_Var2Lit(iObj, 0) );
    Vec_IntFree( vPoints );
    Vec_IntFree( vQuads );
    Vec_BitFree( vHeads );
    Vec_BitFree( vDatas );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Collect distance info.]

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
func Gia_ComputeTest() {
    var pStart string;
    var Line [1000]byte; var Total float32;
    var pFileName string  = "data.txt";
    var pFile *os.File  = os.Open(pFileName);
    if pFile == nil { fmt.Printf("Input file \"%s\" cannot be opened.\n", pFileName); return; }
    for  fgets( Line, 1000, pFile ) != nil  {
        if !strstr(Line, "xxx") {
            continue; }
        if !strstr(Line, "yyy") {
            continue; }
        //printf( "%s", Line );
        pStart = strstr(Line, "zzz");
        if pStart == nil {
            continue; }
        //printf( "%s", pStart + 4 );
        Total += -atof( pStart + 4 );
    }
    fmt.Printf("Total = %.2f\n", Total);
    _ = (pFile).Close();
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

