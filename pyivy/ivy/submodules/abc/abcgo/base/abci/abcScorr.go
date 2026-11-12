package abci

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [abcScorr.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Network and node package.]

  Synopsis    [Signal correspondence testing procedures.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: abcScorr.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/
 
//#include "base/abc/abc.h"
//#include "base/io/ioAbc.h"
//#include "aig/saig/saig.h"
//#include "proof/ssw/ssw.h"
//#include "aig/gia/gia.h"
//#include "proof/cec/cec.h"
//#include "aig/gia/giaAig.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

   
type Tst_Dat_t struct {
    pNetlist  *Abc_Ntk_t;
    pAig  *Aig_Man_t;
    pGia  *Gia_Man_t;
    vId2Name  *Vec_Int_t;
    pFileNameOut       string;
    fFlopOnly         int32;
    fFfNdOnly         int32;
    fDumpBmc         int32;      
}

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkMapGiaIntoNameId(pNetlist *Abc_Ntk_t, pAig *Aig_Man_t, pGia *Gia_Man_t) *Vec_Int_t {
    var vId2Name *Vec_Int_t;
    var pNet, pNode, pAnd *Abc_Obj_t;
    var pObjAig *Aig_Obj_t;
    var i int32;
    vId2Name = Vec_IntAlloc( 0 );
    Vec_IntFill( vId2Name, pGia ? Gia_ManObjNum(pGia) : Aig_ManObjNumMax(pAig), ~0 );
    // copy all names
    // inlining of macro Abc_NtkForEachNet
    for  i = 0; (i < Vec_PtrSize((pNetlist).vObjs)) && (((pNet) = Abc_NtkObj(pNetlist, i)), 1); i++  {            if (pNet) == nil || !Abc_ObjIsNet(pNet) {} else {
        pNode = Abc_ObjFanin0(pNet).pCopy; 
        if pNode && (pAnd = Abc_ObjRegular(pNode.pCopy)) && 
            (pObjAig = (Aig_Obj_t *)Abc_ObjRegular(pAnd.pCopy)) && 
             Aig_ObjType(pObjAig) != AIG_OBJ_NONE {
            if pGia == nil {
                Vec_IntWriteEntry( vId2Name, Aig_ObjId(pObjAig), Abc_ObjId(pNet) );
            } else {
                Vec_IntWriteEntry( vId2Name, Abc_Lit2Var(pObjAig.iData), Abc_ObjId(pNet) ); }
        }
    }
    }
    // overwrite CO names
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(pNetlist)) && (((pNode) = Abc_NtkCo(pNetlist, i)), 1); i++  {
        pNet = Abc_ObjFanin0(pNode);
        pNode = pNode.pCopy;
        if pNode && (pAnd = Abc_ObjRegular(pNode.pCopy)) && 
            (pObjAig = (Aig_Obj_t *)Abc_ObjRegular(pAnd.pCopy)) && 
             Aig_ObjType(pObjAig) != AIG_OBJ_NONE {
            if pGia == nil {
                Vec_IntWriteEntry( vId2Name, Aig_ObjId(pObjAig), Abc_ObjId(pNet) );
            } else {
                Vec_IntWriteEntry( vId2Name, Abc_Lit2Var(pObjAig.iData), Abc_ObjId(pNet) ); }
        }
    }
    // overwrite CI names
    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(pNetlist)) && (((pNode) = Abc_NtkCi(pNetlist, i)), 1); i++  {
        pNet = Abc_ObjFanout0(pNode);
        pNode = pNode.pCopy;
        if pNode && (pAnd = Abc_ObjRegular(pNode.pCopy)) && 
            (pObjAig = (Aig_Obj_t *)Abc_ObjRegular(pAnd.pCopy)) && 
             Aig_ObjType(pObjAig) != AIG_OBJ_NONE {
            if pGia == nil {
                Vec_IntWriteEntry( vId2Name, Aig_ObjId(pObjAig), Abc_ObjId(pNet) );
            } else {
                Vec_IntWriteEntry( vId2Name, Abc_Lit2Var(pObjAig.iData), Abc_ObjId(pNet) ); }
        }
    }
    return vId2Name;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkTestScorrGetName(pNetlist *Abc_Ntk_t, vId2Name *Vec_Int_t, Id int32) string {
//    Abc_Obj_t * pObj;
//printf( "trying to get name for %d\n", Id );
    if Vec_IntEntry(vId2Name, Id) == ~0 {
        return nil; }
//    pObj = Abc_NtkObj( pNetlist, Vec_IntEntry(vId2Name, Id) );
//    pObj = Abc_ObjFanin0(pObj);
//    assert( Abc_ObjIsCi(pObj) );
    return Nm_ManFindNameById( pNetlist.pManName, Vec_IntEntry(vId2Name, Id) );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkTestScorrWriteEquivPair(pNetlist *Abc_Ntk_t, vId2Name *Vec_Int_t, Id1 int32, Id2 int32, pFile *os.File, fPol int32) int32 {
    var pName1 string  = Abc_NtkTestScorrGetName( pNetlist, vId2Name, Id1 );
    var pName2 string  = Abc_NtkTestScorrGetName( pNetlist, vId2Name, Id2 );
    if pName1 == nil || pName2 == nil {
        return 0; }
    fmt.Fprintf(pFile, "%s=%s%s\n", pName1, fPol? "~": "", pName2);
    return 1;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkTestScorrWriteEquivConst(pNetlist *Abc_Ntk_t, vId2Name *Vec_Int_t, Id1 int32, pFile *os.File, fPol int32) int32 {
    var pName1 string  = Abc_NtkTestScorrGetName( pNetlist, vId2Name, Id1 );
    if pName1 == nil {
        return 0; }
    fmt.Fprintf(pFile, "%s=%s%s\n", pName1, fPol? "~": "", "const0");
    return 1;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Abc_NtkBmcFileName_Buffer string // was [1000]byte in C

func Abc_NtkBmcFileName(pName string) string {
    
    var pNameGeneric string  = Extra_FileNameGeneric( pName );
    spStatic_Abc_NtkBmcFileName_Buffer = fmt.Sprintf("%s_bmc%s", pNameGeneric, pName + len(pNameGeneric));
    pNameGeneric = nil;
    return spStatic_Abc_NtkBmcFileName_Buffer;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkTestScorrWriteEquivGia(pData *Tst_Dat_t) int32 {
    var pNetlist *Abc_Ntk_t  = pData.pNetlist;
    var vId2Name *Vec_Int_t  = pData.vId2Name;
    var pGia *Gia_Man_t      = pData.pGia;
    var pFileNameOut string   = pData.pFileNameOut;
    var pFile *os.File;
    var pObj, pRepr *Gia_Obj_t;
    var i, Counter int32;
    if pData.fDumpBmc {
        pData.fDumpBmc = 0;
        pFileNameOut = Abc_NtkBmcFileName( pFileNameOut );
    }
    pFile = os.OpenFile(pFileNameOut, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    Gia_ManSetPhase( pGia );
    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < pGia.nObjs) && ((pObj) = Gia_ManObj(pGia, i)); i++  {
        if !Gia_ObjHasRepr(pGia, i) {
            continue; }
        pRepr = Gia_ManObj( pGia,Gia_ObjRepr(pGia, i) );
        if pData.fFlopOnly {
            if !Gia_ObjIsRo(pGia, pObj) || !(Gia_ObjIsRo(pGia, pRepr)||Gia_ObjIsConst0(pRepr)) {
                continue; }
        
        } else if pData.fFfNdOnly {
            if !Gia_ObjIsRo(pGia, pObj) && !(Gia_ObjIsRo(pGia, pRepr)||Gia_ObjIsConst0(pRepr)) {
                continue; }
        }
        if Gia_ObjRepr(pGia, i) == 0 {
            Counter += Abc_NtkTestScorrWriteEquivConst( pNetlist, vId2Name, i, pFile, Gia_ObjPhase(pObj) );
        } else {
            Counter += Abc_NtkTestScorrWriteEquivPair( pNetlist, vId2Name, Gia_ObjRepr(pGia, i), i, pFile, 
                Gia_ObjPhase(pRepr) ^ Gia_ObjPhase(pObj) ); }
    }
    _ = (pFile).Close();
    fmt.Printf("%d pairs of sequentially equivalent nodes are written into file \"%s\".\n", Counter, pFileNameOut);
    return Counter;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkTestScorrWriteEquivAig(pData *Tst_Dat_t) int32 {
    var pNetlist *Abc_Ntk_t  = pData.pNetlist;
    var vId2Name *Vec_Int_t  = pData.vId2Name;
    var pAig *Aig_Man_t      = pData.pAig;
    var pFileNameOut string   = pData.pFileNameOut;
    var pFile *os.File;
    var pObj, pRepr *Aig_Obj_t;
    var i, Counter int32;
    if pData.fDumpBmc {
        pData.fDumpBmc = 0;
        pFileNameOut = Abc_NtkBmcFileName( pFileNameOut );
    }
    pFile = os.OpenFile(pFileNameOut, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        if (pRepr = Aig_ObjRepr(pAig, pObj)) == nil {
            continue; }
        if pData.fFlopOnly {
            if !Saig_ObjIsLo(pAig, pObj) || !(Saig_ObjIsLo(pAig, pRepr)||pRepr==Aig_ManConst1(pAig)) {
                continue; }
        
        } else if pData.fFfNdOnly {
            if !Saig_ObjIsLo(pAig, pObj) && !(Saig_ObjIsLo(pAig, pRepr)||pRepr==Aig_ManConst1(pAig)) {
                continue; }
        }
        if pRepr == Aig_ManConst1(pAig) {
            Counter += Abc_NtkTestScorrWriteEquivConst( pNetlist, vId2Name, Aig_ObjId(pObj), pFile, Aig_ObjPhase(pObj) );
        } else {
            Counter += Abc_NtkTestScorrWriteEquivPair( pNetlist, vId2Name, Aig_ObjId(pRepr), Aig_ObjId(pObj), pFile, 
                Aig_ObjPhase(pRepr) ^ Aig_ObjPhase(pObj) ); }
    }
    }
    _ = (pFile).Close();
    fmt.Printf("%d pairs of sequentially equivalent nodes are written into file \"%s\".\n", Counter, pFileNameOut);
    return Counter;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkTestScorr(pFileNameIn string, pFileNameOut string, nStepsMax int32, nBTLimit int32, fNewAlgo int32, fFlopOnly int32, fFfNdOnly int32, fVerbose int32) *Abc_Ntk_t {
    extern Aig_Man_t * Abc_NtkToDar( Abc_Ntk_t * pNtk, int fExors, int fRegisters );
    extern Abc_Ntk_t * Abc_NtkFromDarSeqSweep( Abc_Ntk_t * pNtkOld, Aig_Man_t * pMan );

    var pFile *os.File;
    var Data Tst_Dat_t;
    var pData *Tst_Dat_t = &Data;
    var vId2Name *Vec_Int_t;
    var pNetlist, pLogic, pStrash, pResult *Abc_Ntk_t;
    var pAig, pTempAig *Aig_Man_t;
    var pGia, pTempGia *Gia_Man_t;
//    int Counter = 0;
    // check the files
    pFile = os.Open(pFileNameIn);
    if pFile == nil {
        fmt.Printf("Input file \"%s\" cannot be opened.\n", pFileNameIn);
        return nil;
    }
    _ = (pFile).Close();
    // check the files
    pFile = os.OpenFile(pFileNameOut, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    if pFile == nil {
        fmt.Printf("Output file \"%s\" cannot be opened.\n", pFileNameOut);
        return nil;
    }
    _ = (pFile).Close();
    // derive AIG for signal correspondence
    pNetlist = Io_ReadNetlist( pFileNameIn, Io_ReadFileType(pFileNameIn), 1 );
    if pNetlist == nil {
        fmt.Printf("Reading input file \"%s\" has failed.\n", pFileNameIn);
        return nil;
    }
    pLogic = Abc_NtkToLogic( pNetlist );
    if pLogic == nil {
        Abc_NtkDelete( pNetlist );
        fmt.Printf("Deriving logic network from input file %s has failed.\n", pFileNameIn);
        return nil;
    }
    if Extra_FileIsType( pFileNameIn, ".bench", ".BENCH", nil ) {
        // get the init file name
        var pFileNameInit string  = Extra_FileNameGenericAppend( pLogic.pSpec, ".init" );
        pFile = os.Open(pFileNameInit);
        if pFile == nil {
            fmt.Printf("Init file \"%s\" cannot be opened.\n", pFileNameInit);
            return nil;
        }
        Io_ReadBenchInit( pLogic, pFileNameInit );
        Abc_NtkConvertDcLatches( pLogic );
        if fVerbose != 0 {
            fmt.Printf("Initial state was derived from file \"%s\".\n", pFileNameInit); }
    }
    pStrash = Abc_NtkStrash( pLogic, 0, 1, 0 );
    if pStrash == nil {
        Abc_NtkDelete( pLogic );
        Abc_NtkDelete( pNetlist );
        fmt.Printf("Deriving strashed network from input file %s has failed.\n", pFileNameIn);
        return nil;
    }
    pAig = Abc_NtkToDar( pStrash, 0, 1 ); // performs "zero" internally
    // the newer computation (&scorr)
    if fNewAlgo != 0 {
        var CorPars Cec_ParCor_t;
        var pCorPars *Cec_ParCor_t = &CorPars;
        Cec_ManCorSetDefaultParams( pCorPars );
        pCorPars.nBTLimit  = nBTLimit;
        pCorPars.nStepsMax = nStepsMax;
        pCorPars.fVerbose  = fVerbose;
        pCorPars.fUseCSat  = 1;
        pGia = Gia_ManFromAig( pAig );
        // prepare the data-structure
        *pData = Tst_Dat_t{};
        pData.pNetlist     = pNetlist;
        pData.pAig         = nil;
        pData.pGia         = pGia;
        pData.vId2Name     = vId2Name = Abc_NtkMapGiaIntoNameId( pNetlist, pAig, pGia );
        pData.pFileNameOut = pFileNameOut;
        pData.fFlopOnly    = fFlopOnly;
        pData.fFfNdOnly    = fFfNdOnly;
        pData.fDumpBmc     = 1;
        pCorPars.pData     = pData;
        pCorPars.pFunc     = (void *)Abc_NtkTestScorrWriteEquivGia;
        // call signal correspondence
        pTempGia = Cec_ManLSCorrespondence( pGia, pCorPars );
        pTempAig = Gia_ManToAigSimple( pTempGia );
        Gia_ManStop( pTempGia );
        Gia_ManStop( pGia );
    
    // the older computation (scorr)
    } else {
        var SswPars Ssw_Pars_t;
        var pSswPars *Ssw_Pars_t = &SswPars;
        Ssw_ManSetDefaultParams( pSswPars );
        pSswPars.nBTLimit  = nBTLimit;
        pSswPars.nStepsMax = nStepsMax;
        pSswPars.fVerbose  = fVerbose;
        // preSswPare the data-structure
        *pData = Tst_Dat_t{};
        pData.pNetlist     = pNetlist;
        pData.pAig         = pAig;
        pData.pGia         = nil;
        pData.vId2Name     = vId2Name = Abc_NtkMapGiaIntoNameId( pNetlist, pAig, nil );
        pData.pFileNameOut = pFileNameOut;
        pData.fFlopOnly    = fFlopOnly;
        pData.fFfNdOnly    = fFfNdOnly;
        pData.fDumpBmc     = 1;
        pSswPars.pData     = pData;
        pSswPars.pFunc     = (void *)Abc_NtkTestScorrWriteEquivAig;
        // call signal correspondence
        pTempAig = Ssw_SignalCorrespondence( pAig, pSswPars );
    }
    // create the resulting AIG
    pResult = Abc_NtkFromDarSeqSweep( pStrash, pTempAig );
    // cleanup
    Vec_IntFree( vId2Name );
    Aig_ManStop( pAig );
    Aig_ManStop( pTempAig );
    Abc_NtkDelete( pStrash );
    Abc_NtkDelete( pLogic );
    Abc_NtkDelete( pNetlist );
    return pResult;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

