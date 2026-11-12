package mpm

import (
	"fmt"
	"os"
)

/**CFile****************************************************************

  FileName    [mpmGates.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Configurable technology mapper.]

  Synopsis    [Standard-cell mapping.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 1, 2013.]

  Revision    [$Id: mpmGates.c,v 1.00 2013/06/01 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "mpmInt.h"
//#include "misc/st/st.h"
//#include "map/mio/mio.h"
//#include "map/scl/sclSize.h"
//#include "map/scl/sclTime.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Finds matches fore each DSD class.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mpm_ManFindDsdMatches(p *Mpm_Man_t, pScl any) *Vec_Wec_t {
    var fVerbose int32  = p.pPars.fVeryVerbose;
    var pLib *SC_Lib  = (SC_Lib *)pScl;
    var vClasses *Vec_Wec_t;
    var vClass *Vec_Int_t;
    var pRepr *SC_Cell;
    var i, Config, iClass int32;
    var Truth word;
    vClasses = Vec_WecStart( 600 );
    // inlining of macro SC_LibForEachCellClass
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(&pLib.vCellClasses)) && (((pRepr) = (SC_Cell *)Vec_PtrEntry(&pLib.vCellClasses, i)), 1); i++  {
        if pRepr.n_inputs > 6 || pRepr.n_outputs > 1 {
            if fVerbose != 0 {
            fmt.Printf("Skipping cell %s with %d inputs and %d outputs\n", pRepr.pName, pRepr.n_inputs, pRepr.n_outputs); }
            continue;
        }
        Truth = *Vec_WrdArray( &SC_CellPin(pRepr, pRepr.n_inputs).vFunc );
        Config = Mpm_CutCheckDsd6( p, Truth );
        if Config == -1 {
            if fVerbose != 0 {
            fmt.Printf("Skipping cell %s with non-DSD function\n", pRepr.pName); }
            continue;
        }
        iClass = Config >> 17;
        Config = (pRepr.Id << 17) | (Config & 0x1FFFF);
        // write gate and NPN config for this DSD class
        vClass = Vec_WecEntry( vClasses, iClass );
        Vec_IntPush( vClass, Config );
        if !(fVerbose != 0) {
            continue; }

        fmt.Printf("Gate %5d  %-30s : ", pRepr.Id, pRepr.pName);
        fmt.Printf("Class %3d  ", iClass);
        fmt.Printf("Area %10.3f  ", pRepr.area);
        Extra_PrintBinary( os.Stdout, (unsigned *)&Config, 17 );
        fmt.Printf("   ");
        Kit_DsdPrintFromTruth( (unsigned *)&Truth, pRepr.n_inputs ); fmt.Printf("\n");
    }
    return vClasses;
}

/**Function*************************************************************

  Synopsis    [Find mapping of DSD classes into Genlib library cells.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mpm_ManFindCells(pMio *Mio_Library_t, pScl *SC_Lib, vNpnConfigs *Vec_Wec_t) *Vec_Ptr_t {
    var vNpnGatesMio *Vec_Ptr_t;
    var vClass *Vec_Int_t;
    var pMioGate *Mio_Gate_t;
    var pCell *SC_Cell;
    var Config, iClass int32;
    vNpnGatesMio = Vec_PtrStart( Vec_WecSize(vNpnConfigs) );
    // inlining of macro Vec_WecForEachLevel
    for  iClass = 0; (iClass < Vec_WecSize(vNpnConfigs)) && (((vClass) = Vec_WecEntry(vNpnConfigs, iClass)), 1); iClass++  {
        if Vec_IntSize(vClass) == 0 {
            continue; }
        Config = Vec_IntEntry(vClass, 0);
        pCell = SC_LibCell( pScl, (Config >> 17) );
        pMioGate = Mio_LibraryReadGateByName( pMio, pCell.pName, nil );
        if pMioGate == nil {
            Vec_PtrFree( vNpnGatesMio );
            return nil;
        }
        assert(pMioGate != nil, "pMioGate != nil");
        Vec_PtrWriteEntry( vNpnGatesMio, iClass, pMioGate );
    }
    return vNpnGatesMio;
}

/**Function*************************************************************

  Synopsis    [Derive mapped network as an ABC network.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mpm_ManFindMappedNodes(p *Mpm_Man_t) *Vec_Int_t {
    var vNodes *Vec_Int_t;
    var pObj *Mig_Obj_t;
    vNodes = Vec_IntAlloc( 1000 );
    // inlining of macro Mig_ManForEachObj
    for  p.pMig.iPage = 0; p.pMig.iPage < Vec_PtrSize(&p.pMig.vPages) &&                   ((p.pMig.pPage) = (Mig_Obj_t *)Vec_PtrEntry(&p.pMig.vPages, p.pMig.iPage)); p.pMig.iPage++  {         for  pObj = p.pMig.pPage; !Mig_ObjIsNone(pObj); pObj++  {
        if Mig_ObjIsNode(pObj) && Mpm_ObjMapRef(p, pObj) {
            Vec_IntPush( vNodes, Mig_ObjId(pObj) ); }
    }
    }
    return vNodes;
}
func Mpm_ManGetAbcNode(pNtk *Abc_Ntk_t, vCopy *Vec_Int_t, iMigLit int32) *Abc_Obj_t {
    var pObj *Abc_Obj_t;
    var iObjId int32  = Vec_IntEntry( vCopy, iMigLit );
    if iObjId >= 0 {
        return Abc_NtkObj( pNtk, iObjId ); }
    iObjId = Vec_IntEntry( vCopy, Abc_LitNot(iMigLit) );
    assert(iObjId >= 0, "iObjId >= 0");
    pObj = Abc_NtkCreateNodeInv( pNtk, Abc_NtkObj(pNtk, iObjId) );
    Vec_IntWriteEntry( vCopy, iMigLit, Abc_ObjId(pObj) );
    return pObj;
}
func Mpm_ManDeriveMappedAbcNtk(p *Mpm_Man_t, pMio *Mio_Library_t) *Abc_Ntk_t {
    var pNtk *Abc_Ntk_t;
    var vNpnGatesMio *Vec_Ptr_t;
    var vNodes, vCopy, vClass *Vec_Int_t;
    var pObj, pFanin *Abc_Obj_t;
    var pNode *Mig_Obj_t;
    var pCutBest *Mpm_Cut_t;
    var i, k, iNode, iMigLit, fCompl, Config int32;

    // find mapping of SCL cells into MIO cells
    vNpnGatesMio = Mpm_ManFindCells( pMio, (SC_Lib *)p.pPars.pScl, p.vNpnConfigs );
    if vNpnGatesMio == nil {
        fmt.Printf("Genlib library does not match SCL library.\n");
        return nil;
    }

    // create mapping for each phase of each node
    vCopy = Vec_IntStartFull( 2 * Mig_ManObjNum(p.pMig) );

    // get internal nodes
    vNodes = Mpm_ManFindMappedNodes( p );

    // start the network
    pNtk = Abc_NtkAlloc( ABC_NTK_LOGIC, ABC_FUNC_MAP, 1 );
    pNtk.pName = Extra_UtilStrsav( p.pMig.pName );
    pNtk.pManFunc = pMio;

    // create primary inputs
    // inlining of macro Mig_ManForEachCi
    for  i = 0; (i < Vec_IntSize(&p.pMig.vCis)) && ((pNode) = Mig_ManCi(p.pMig, i)); i++  {
        pObj = Abc_NtkCreatePi(pNtk);
        Vec_IntWriteEntry( vCopy, Abc_Var2Lit( Mig_ObjId(pNode), 0 ), Abc_ObjId(pObj) );
    }
    Abc_NtkAddDummyPiNames( pNtk );

    // create constant nodes
    // inlining of macro Mig_ManForEachCo
    for  i = 0; (i < Vec_IntSize(&p.pMig.vCos)) && ((pNode) = Mig_ManCo(p.pMig, i)); i++  {
        if Mig_ObjFaninLit(pNode, 0) == 0 {
            pObj = Abc_NtkCreateNodeConst0(pNtk);
            Vec_IntWriteEntry( vCopy, Abc_Var2Lit( 0, 0 ), Abc_ObjId(pObj) );
            break;
        }
    }
    // inlining of macro Mig_ManForEachCo
    for  i = 0; (i < Vec_IntSize(&p.pMig.vCos)) && ((pNode) = Mig_ManCo(p.pMig, i)); i++  {
        if Mig_ObjFaninLit(pNode, 0) == 1 {
            pObj = Abc_NtkCreateNodeConst1(pNtk);
            Vec_IntWriteEntry( vCopy, Abc_Var2Lit( 0, 1 ), Abc_ObjId(pObj) );
            break;
        }
    }

    // create internal nodes
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vNodes)) && (((iNode) = Vec_IntEntry(vNodes, i)), 1); i++  {
        pCutBest = Mpm_ObjCutBestP( p, Mig_ManObj(p.pMig, iNode) );
        vClass = Vec_WecEntry( p.vNpnConfigs, Abc_Lit2Var(pCutBest.iFunc) );
        Config = Vec_IntEntry( vClass, 0 );
        pObj = Abc_NtkCreateNode( pNtk );
        pObj.pData = Vec_PtrEntry( vNpnGatesMio, Abc_Lit2Var(pCutBest.iFunc) );
        assert(pObj.pData != nil, "pObj.pData != nil");
        fCompl = pCutBest.fCompl ^ Abc_LitIsCompl(pCutBest.iFunc) ^ ((Config >> 16) & 1);
        Config &= 0xFFFF;
        for  k = 0; k < int32(pCutBest.nLeaves); k++  {
            assert((Config >> 6) < 720, "(Config >> 6) < 720");
            iMigLit = pCutBest.pLeaves[ int32((p.Perm6[Config >> 6][k])) ];
            pFanin = Mpm_ManGetAbcNode( pNtk, vCopy, Abc_LitNotCond(iMigLit, (Config >> k) & 1) );
            Abc_ObjAddFanin( pObj, pFanin );
        }
        Vec_IntWriteEntry( vCopy, Abc_Var2Lit(iNode, fCompl), Abc_ObjId(pObj) );
    }

    // create primary outputs
    // inlining of macro Mig_ManForEachCo
    for  i = 0; (i < Vec_IntSize(&p.pMig.vCos)) && ((pNode) = Mig_ManCo(p.pMig, i)); i++  {
        pObj = Abc_NtkCreatePo(pNtk);
        pFanin = Mpm_ManGetAbcNode( pNtk, vCopy, Mig_ObjFaninLit(pNode, 0) );
        Abc_ObjAddFanin( pObj, pFanin );
    }
    Abc_NtkAddDummyPoNames( pNtk );

    // clean up
    Vec_PtrFree( vNpnGatesMio );
    Vec_IntFree( vNodes );
    Vec_IntFree( vCopy );
    return pNtk;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mpm_ManPerformCellMapping(pMig *Mig_Man_t, pPars *Mpm_Par_t, pMio *Mio_Library_t) *Abc_Ntk_t {
    var pNew *Abc_Ntk_t;
    var p *Mpm_Man_t;
    assert(pPars.fMap4Gates, "pPars.fMap4Gates");
    p = Mpm_ManStart( pMig, pPars );
    if p.pPars.fVerbose { 
        Mpm_ManPrintStatsInit( p ); }
    p.vNpnConfigs = Mpm_ManFindDsdMatches( p, p.pPars.pScl );
    Mpm_ManPrepare( p );
    Mpm_ManPerform( p );
    if p.pPars.fVerbose { 
        Mpm_ManPrintStats( p ); }
    pNew = Mpm_ManDeriveMappedAbcNtk( p, pMio );
    Mpm_ManStop( p );
    return pNew;
}
func Mpm_ManCellMapping(pGia *Gia_Man_t, pPars *Mpm_Par_t, pMio any) *Abc_Ntk_t {
    var p *Mig_Man_t;
    var pNew *Abc_Ntk_t;
    assert(pMio != nil, "pMio != nil");
    assert(pPars.pLib.LutMax <= MPM_VAR_MAX, "pPars.pLib.LutMax <= MPM_VAR_MAX");
    assert(pPars.nNumCuts <= MPM_CUT_MAX, "pPars.nNumCuts <= MPM_CUT_MAX");
    if pPars.fUseGates {
        pGia = Gia_ManDupMuxes( pGia, 2 );
        p = Mig_ManCreate( pGia );
        Gia_ManStop( pGia );
    
    } else {
        p = Mig_ManCreate( pGia ); }
    pNew = Mpm_ManPerformCellMapping( p, pPars, (Mio_Library_t *)pMio );
    Mig_ManStop( p );
    return pNew;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

