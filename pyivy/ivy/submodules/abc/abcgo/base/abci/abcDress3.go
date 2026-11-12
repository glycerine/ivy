package abci

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [abcDress3.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Network and node package.]

  Synopsis    [Transfers names from one netlist to the other.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: abcDress3.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "base/abc/abc.h"
//#include "base/io/ioAbc.h"
//#include "proof/cec/cec.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Converts AIG from HOP to GIA.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_ConvertHopToGia_rec1(p *Gia_Man_t, pObj *Hop_Obj_t) {
    assert(!Hop_IsComplement(pObj), "!Hop_IsComplement(pObj)");
    if !Hop_ObjIsNode(pObj) || Hop_ObjIsMarkA(pObj) {
        return; }
    Abc_ConvertHopToGia_rec1( p, Hop_ObjFanin0(pObj) ); 
    Abc_ConvertHopToGia_rec1( p, Hop_ObjFanin1(pObj) );
    pObj.iData = Gia_ManHashAnd( p, Hop_ObjChild0CopyI(pObj), Hop_ObjChild1CopyI(pObj) );
    assert(!Hop_ObjIsMarkA(pObj), "!Hop_ObjIsMarkA(pObj)"); // loop detection
    Hop_ObjSetMarkA( pObj );
}
func Abc_ConvertHopToGia_rec2(pObj *Hop_Obj_t) {
    assert(!Hop_IsComplement(pObj), "!Hop_IsComplement(pObj)");
    if !Hop_ObjIsNode(pObj) || !Hop_ObjIsMarkA(pObj) {
        return; }
    Abc_ConvertHopToGia_rec2( Hop_ObjFanin0(pObj) ); 
    Abc_ConvertHopToGia_rec2( Hop_ObjFanin1(pObj) );
    assert(Hop_ObjIsMarkA(pObj), "Hop_ObjIsMarkA(pObj)"); // loop detection
    Hop_ObjClearMarkA( pObj );
}
func Abc_ConvertHopToGia(p *Gia_Man_t, pRoot *Hop_Obj_t) int32 {
    assert(!Hop_IsComplement(pRoot), "!Hop_IsComplement(pRoot)");
    if Hop_ObjIsConst1( pRoot ) {
        return 1; }
    Abc_ConvertHopToGia_rec1( p, pRoot );
    Abc_ConvertHopToGia_rec2( pRoot );
    return pRoot.iData;
}

/**Function*************************************************************

  Synopsis    [Add logic from pNtk to the AIG manager p.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkAigToGiaOne(p *Gia_Man_t, pNtk *Abc_Ntk_t, vMap *Vec_Int_t) {
    var pHopMan *Hop_Man_t;
    var pHopObj *Hop_Obj_t;
    var vNodes *Vec_Ptr_t;
    var pNode, pFanin *Abc_Obj_t;
    var i, k int32;
    assert(Abc_NtkIsAigLogic(pNtk), "Abc_NtkIsAigLogic(pNtk)");
    pHopMan = (Hop_Man_t *)pNtk.pManFunc;
    Hop_ManConst1(pHopMan).iData = 1;
    // image primary inputs
    Abc_NtkCleanCopy( pNtk );
    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(pNtk)) && (((pNode) = Abc_NtkCi(pNtk, i)), 1); i++  {
        pNode.iTemp = Gia_ManCiLit(p, Vec_IntEntry(vMap, i));
    }
    // iterate through nodes used in the mapping
    vNodes = Abc_NtkDfs( pNtk, 1 );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pNode) = (Abc_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        // inlining of macro Abc_ObjForEachFanin
        for  k = 0; (k < Abc_ObjFaninNum(pNode)) && (((pFanin) = Abc_ObjFanin(pNode, k)), 1); k++  {
            Hop_ManPi(pHopMan, k).iData = pFanin.iTemp;
        }
        pHopObj = Hop_Regular( (Hop_Obj_t *)pNode.pData );
        assert(Abc_ObjFaninNum(pNode) <= Hop_ManPiNum(pHopMan), "Abc_ObjFaninNum(pNode) <= Hop_ManPiNum(pHopMan)");
        if Hop_DagSize(pHopObj) > 0 {
            Abc_ConvertHopToGia( p, pHopObj ); }
        pNode.iTemp = Abc_LitNotCond( pHopObj.iData, Hop_IsComplement( (Hop_Obj_t *)pNode.pData ) );
    }
    Vec_PtrFree( vNodes );
    // create primary outputs
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(pNtk)) && (((pNode) = Abc_NtkCo(pNtk, i)), 1); i++  {
        Gia_ManAppendCo( p, Abc_ObjFanin0(pNode).iTemp );
    }
}
func Abc_NtkAigToGiaTwo(pNtk1 *Abc_Ntk_t, pNtk2 *Abc_Ntk_t, fByName int32) *Gia_Man_t {
    var p *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var pNode *Abc_Obj_t;
    var vMap1, vMap2 *Vec_Int_t;
    var i, Index int32;
    assert(Abc_NtkIsAigLogic(pNtk1), "Abc_NtkIsAigLogic(pNtk1)");
    assert(Abc_NtkIsAigLogic(pNtk2), "Abc_NtkIsAigLogic(pNtk2)");
    // find common variables
    if fByName != 0 {
        var nCommon int32;
        vMap1 = Vec_IntStartNatural( Abc_NtkCiNum(pNtk1) );
        vMap2 = Vec_IntAlloc( Abc_NtkCiNum(pNtk2) );
        // inlining of macro Abc_NtkForEachCi
        for  i = 0; (i < Abc_NtkCiNum(pNtk1)) && (((pNode) = Abc_NtkCi(pNtk1, i)), 1); i++  {
            pNode.iTemp = Index++;
        }
        assert(Index == Abc_NtkCiNum(pNtk1), "Index == Abc_NtkCiNum(pNtk1)");
        // inlining of macro Abc_NtkForEachCi
        for  i = 0; (i < Abc_NtkCiNum(pNtk2)) && (((pNode) = Abc_NtkCi(pNtk2, i)), 1); i++  {
            var Num int32  = Nm_ManFindIdByName( pNtk1.pManName, Abc_ObjName(pNode), ABC_OBJ_PI );
            if Num < 0 {
                Num = Nm_ManFindIdByName( pNtk1.pManName, Abc_ObjName(pNode), ABC_OBJ_BO ); }
            assert(Num < 0 || Abc_ObjIsCi(Abc_NtkObj(pNtk1, Num)), "Num < 0 || Abc_ObjIsCi(Abc_NtkObj(pNtk1, Num))");
            if Num >= 0 {
                Vec_IntPush( vMap2, Abc_NtkObj(pNtk1, Num).iTemp );
nCommon++;
            } else {
                Vec_IntPush( vMap2, Index++ ); }
        }
        // report
        fmt.Printf("Matched %d vars by name.", nCommon);
        if nCommon != Abc_NtkCiNum(pNtk1) {
            fmt.Printf(" Netlist1 has %d unmatched vars.", Abc_NtkCiNum(pNtk1) - nCommon); }
        if nCommon != Abc_NtkCiNum(pNtk2) {
            fmt.Printf(" Netlist2 has %d unmatched vars.", Abc_NtkCiNum(pNtk2) - nCommon); }
        fmt.Printf("\n");
    
    } else {
        vMap1 = Vec_IntStartNatural( Abc_NtkCiNum(pNtk1) );
        vMap2 = Vec_IntStartNatural( Abc_NtkCiNum(pNtk2) );
        Index = Abc_MaxInt( Vec_IntSize(vMap1), Vec_IntSize(vMap2) );
        // report
        fmt.Printf("Matched %d vars by order.", Abc_MinInt(Abc_NtkCiNum(pNtk1), Abc_NtkCiNum(pNtk2)));
        if Abc_NtkCiNum(pNtk1) < Abc_NtkCiNum(pNtk2) {
            fmt.Printf(" The last %d vars of Netlist2 are unmatched vars.", Abc_NtkCiNum(pNtk2) - Abc_NtkCiNum(pNtk1)); }
        if Abc_NtkCiNum(pNtk1) > Abc_NtkCiNum(pNtk2) {
            fmt.Printf(" The last %d vars of Netlist1 are unmatched vars.", Abc_NtkCiNum(pNtk1) - Abc_NtkCiNum(pNtk2)); }
        fmt.Printf("\n");
    }
    // create new manager
    p = Gia_ManStart( 10000 );
    p.pName = Abc_UtilStrsav( Abc_NtkName(pNtk1) );
    p.pSpec = Abc_UtilStrsav( Abc_NtkSpec(pNtk1) );
    for  i = 0; i < Index; i++  {
        Gia_ManAppendCi(p);
    }
    // add logic
    Gia_ManHashAlloc( p );
    Abc_NtkAigToGiaOne( p, pNtk1, vMap1 );
    Abc_NtkAigToGiaOne( p, pNtk2, vMap2 );
    Gia_ManHashStop( p );
    Vec_IntFree( vMap1 );
    Vec_IntFree( vMap2 );
    // add extra POs to dangling nodes
    Gia_ManCreateValueRefs( p );
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else
        if pObj.Value == 0 {
            Gia_ManAppendCo( p, Abc_Var2Lit(i, 0) ); }
    }
    return p;
}

/**Function*************************************************************

  Synopsis    [Collect equivalence class information.]

  Description [Each class is represented as follows: 
      <num_entries><entry1><entry2>...<entryN> 
  where <entry> is nodeId with 1-bit for complement and 1-bit for network.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkCollectAddOne(iNtk int32, iObj int32, iGiaLit int32, pGia *Gia_Man_t, vGia2Cla *Vec_Int_t, vNexts []*Vec_Int_t) {
    var iRepr int32  = Gia_ObjReprSelf( pGia, Abc_Lit2Var(iGiaLit) );
    var Compl int32  = Abc_LitIsCompl(iGiaLit) ^ Gia_ObjPhase(Gia_ManObj(pGia, iRepr)) ^ Gia_ObjPhase(Gia_ManObj(pGia, Abc_Lit2Var(iGiaLit)));
    var Added int32  = Abc_Var2Lit( Abc_Var2Lit(iObj, Compl), iNtk );
    var Entry int32  = Vec_IntEntry( vGia2Cla, iRepr );
    Vec_IntWriteEntry( vNexts[iNtk], iObj, Entry );
    Vec_IntWriteEntry( vGia2Cla, iRepr, Added );
}
func Abc_NtkCollectEquivClasses(pNtks []*Abc_Ntk_t, pGia *Gia_Man_t) *Vec_Int_t {
    var vClass *Vec_Int_t  = Vec_IntAlloc( 100 );
    var vClasses *Vec_Int_t  = Vec_IntAlloc( 1000 );
    var vGia2Cla *Vec_Int_t  = Vec_IntStartFull( Gia_ManObjNum(pGia) ); // mapping objId into classId
    var vNexts [2]*Vec_Int_t = [2]*Vec_Int_t{ Vec_IntStartFull(Abc_NtkObjNumMax(pNtks[0])), Vec_IntStartFull(Abc_NtkObjNumMax(pNtks[1])) };
    var pObj *Abc_Obj_t;
    var n, i, k, Entry, fCompl int32;
    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(pNtks[0])) && (((pObj) = Abc_NtkCi(pNtks[0], i)), 1); i++  {
        Abc_NtkCollectAddOne( 0, Abc_ObjId(pObj), pObj.iTemp, pGia, vGia2Cla, vNexts );
    }
    for  n = 0; n < 2; n++  {
        // inlining of macro Abc_NtkForEachNode
        for  i = 0; (i < Vec_PtrSize((pNtks[n]).vObjs)) && (((pObj) = Abc_NtkObj(pNtks[n], i)), 1); i++  {           if (pObj) == nil || !Abc_ObjIsNode(pObj) {} else {
            Abc_NtkCollectAddOne( n, Abc_ObjId(pObj), pObj.iTemp, pGia, vGia2Cla, vNexts ); }
        }
    }
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vGia2Cla)) && (((Entry) = Vec_IntEntry(vGia2Cla, i)), 1); i++  {
        Vec_IntClear( vClass );
        for  ; Entry >= 0; Entry = Vec_IntEntry(vNexts[Entry&1], Entry>>2)  {
            Vec_IntPush( vClass, Entry );
        }
        if Vec_IntSize(vClass) < 2 {
            continue; }
        Vec_IntReverseOrder( vClass );
        fCompl = 2 & Vec_IntEntry( vClass, 0 );
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vClass)) && (((Entry) = Vec_IntEntry(vClass, k)), 1); k++  {
            Vec_IntWriteEntry( vClass, k, Entry ^ fCompl );
        }        
        Vec_IntPush( vClasses, Vec_IntSize(vClass) );
        Vec_IntAppend( vClasses, vClass );
    }
    Vec_IntFree( vGia2Cla );
    Vec_IntFree( vNexts[0] );
    Vec_IntFree( vNexts[1] );
    Vec_IntFree( vClass );
    return vClasses;
}

/**Function*************************************************************

  Synopsis    [Write the output file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDumpEquivFile(pFileName string, vClasses *Vec_Int_t, pNtks []*Abc_Ntk_t) {
    var i, c, k, Entry int32;
    var pFile *os.File  = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    if pFile == nil { fmt.Printf("Cannot open file %s for writing.\n", pFileName); return; }
    fmt.Fprintf(pFile, "# Node equivalences computed by ABC for networks \"%s\" and \"%s\" on %s\n\n", Abc_NtkName(pNtks[0]), Abc_NtkName(pNtks[1]), Extra_TimeStamp());
    for  i = c = 0; i < Vec_IntSize(vClasses); c++, i += 1 + Vec_IntEntry(vClasses, i)  {
        // inlining of macro Vec_IntForEachEntryStartStop
        for  k = i + 1; (k < i + 1 + Vec_IntEntry(vClasses, i)) && (((Entry) = Vec_IntEntry(vClasses, k)), 1); k++  {
            var pNtk *Abc_Ntk_t  = pNtks[Entry & 1];
            var pObjName string  = Abc_ObjName( Abc_NtkObj(pNtk, Entry>>2) );
            fmt.Fprintf(pFile, "%d:%s:%s%s\n", c+1, Abc_NtkName(pNtk), (Entry&2) ? "NOT:":"", pObjName);
        }
        fmt.Fprintf(pFile, "\n");
    }
    _ = (pFile).Close();
}


/**Function*************************************************************

  Synopsis    [Compute and dump equivalent name classes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDumpEquiv(pNtks []*Abc_Ntk_t, pFileName string, nConfs int32, fByName int32, fVerbose int32) {
    //abctime clk = Abc_Clock();
    var pTemp *Gia_Man_t;
    var vClasses *Vec_Int_t;
    // derive shared AIG for the two networks
    var pGia *Gia_Man_t  = Abc_NtkAigToGiaTwo( pNtks[0], pNtks[1], fByName );
    if fVerbose != 0 {
        fmt.Printf("Computing equivalences for networks \"%s\" and \"%s\" with conflict limit %d.\n", Abc_NtkName(pNtks[0]), Abc_NtkName(pNtks[1]), nConfs); }
    // compute equivalences in this AIG
    pTemp = Gia_ManComputeGiaEquivs( pGia, nConfs, fVerbose );
    Gia_ManStop( pTemp );
    //if ( fVerbose )
    //    Abc_PrintTime( 1, "Equivalence computation time", Abc_Clock() - clk );
    if fVerbose != 0 {
        Gia_ManPrintStats( pGia, nil ); }
    // collect equivalence class information
    vClasses = Abc_NtkCollectEquivClasses( pNtks, pGia );
    Gia_ManStop( pGia );
    // dump information into the output file
    Abc_NtkDumpEquivFile( pFileName, vClasses, pNtks );
    Vec_IntFree( vClasses );
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

