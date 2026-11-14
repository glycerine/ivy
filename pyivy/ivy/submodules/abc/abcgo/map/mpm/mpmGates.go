package mpm

import (
	"fmt"
	gia "github.com/glycerine/abcgo/aig/gia"
	. "github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/bool/kit"
	mio "github.com/glycerine/abcgo/map/mio"
	scl "github.com/glycerine/abcgo/map/scl"
	"github.com/glycerine/abcgo/misc/extra"
	"github.com/glycerine/abcgo/misc/util"
	. "github.com/glycerine/abcgo/misc/vec"
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

/*
*Function*************************************************************

	Synopsis    [Finds matches fore each DSD class.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Mpm_ManFindDsdMatches(p *Mpm_Man_t, pScl any) *Vec_Wec_t {
	var fVerbose int32 = p.pPars.fVeryVerbose
	pLib, ok := pScl.(*scl.SC_Lib)
	assert(ok, "pScl is *scl.SC_Lib")
	var vClasses *Vec_Wec_t
	var vClass *Vec_Int_t
	var pRepr *scl.SC_Cell
	var i, Config, iClass int32
	var Truth word
	vClasses = Vec_WecStart(600)
	// inlining of macro scl.SC_LibForEachCellClass
	// inlining of macro Vec_PtrForEachEntry
	for i = 0; int(i) < len(pLib.CellClasses); i++ {
		pRepr = pLib.CellClasses[int(i)]
		if pRepr.Inputs > 6 || pRepr.Outputs > 1 {
			if fVerbose != 0 {
				fmt.Printf("Skipping cell %s with %d inputs and %d outputs\n", pRepr.Name, pRepr.Inputs, pRepr.Outputs)
			}
			continue
		}
		Truth = scl.SC_CellFunc(pRepr)[0]
		Config = Mpm_CutCheckDsd6(p, Truth)
		if Config == -1 {
			if fVerbose != 0 {
				fmt.Printf("Skipping cell %s with non-DSD function\n", pRepr.Name)
			}
			continue
		}
		iClass = Config >> 17
		Config = (int32(pRepr.ID) << 17) | (Config & 0x1FFFF)
		// write gate and NPN config for this DSD class
		vClass = Vec_WecEntry(vClasses, int(iClass))
		Vec_IntPush(vClass, Config)
		if !(fVerbose != 0) {
			continue
		}

		fmt.Printf("Gate %5d  %-30s : ", pRepr.ID, pRepr.Name)
		fmt.Printf("Class %3d  ", iClass)
		fmt.Printf("Area %10.3f  ", pRepr.Area)
		extra.Extra_PrintBinary(os.Stdout, []uint32{uint32(Config)}, 17)
		fmt.Printf("   ")
		kit.Kit_DsdPrintFromTruth([]uint32{uint32(Truth), uint32(Truth >> 32)}, int32(pRepr.Inputs))
		fmt.Printf("\n")
	}
	return vClasses
}

/*
*Function*************************************************************

	Synopsis    [Find mapping of DSD classes into Genlib library cells.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Mpm_ManFindCells(pMio *mio.Mio_Library_t, pScl *scl.SC_Lib, vNpnConfigs *Vec_Wec_t) *Vec_Ptr_t {
	var vNpnGatesMio *Vec_Ptr_t
	var vClass *Vec_Int_t
	var pMioGate *mio.Mio_Gate_t
	var pCell *scl.SC_Cell
	var Config, iClass int32
	vNpnGatesMio = Vec_PtrStart(Vec_WecSize(vNpnConfigs))
	// inlining of macro Vec_WecForEachLevel
	for iClass = 0; iClass < int32(Vec_WecSize(vNpnConfigs)); iClass++ {
		vClass = Vec_WecEntry(vNpnConfigs, int(iClass))
		if Vec_IntSize(vClass) == 0 {
			continue
		}
		Config = Vec_IntEntry(vClass, 0)
		pCell = scl.SC_LibCell(pScl, int(Config>>17))
		pMioGate = mio.Mio_LibraryReadGateByName(pMio, pCell.Name, "")
		if pMioGate == nil {
			Vec_PtrFree(vNpnGatesMio)
			return nil
		}
		assert(pMioGate != nil, "pMioGate != nil")
		Vec_PtrWriteEntry(vNpnGatesMio, int(iClass), pMioGate)
	}
	return vNpnGatesMio
}

/*
*Function*************************************************************

	Synopsis    [Derive mapped network as an ABC network.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Mpm_ManFindMappedNodes(p *Mpm_Man_t) *Vec_Int_t {
	var vNodes *Vec_Int_t
	var pObj *Mig_Obj_t
	vNodes = Vec_IntAlloc(1000)
	// inlining of macro Mig_ManForEachObj
	for id := int32(0); id < Mig_ManObjNum(p.pMig); id++ {
		pObj = Mig_ManObj(p.pMig, id)
		if Mig_ObjIsNode(pObj) != 0 && Mpm_ObjMapRef(p, pObj) != 0 {
			Vec_IntPush(vNodes, Mig_ObjId(pObj))
		}
	}
	return vNodes
}
func Mpm_ManGetAbcNode(pNtk *Abc_Ntk_t, vCopy *Vec_Int_t, iMigLit int32) *Abc_Obj_t {
	var pObj *Abc_Obj_t
	var iObjId int32 = Vec_IntEntry(vCopy, int(iMigLit))
	if iObjId >= 0 {
		return Abc_NtkObj(pNtk, iObjId)
	}
	iObjId = Vec_IntEntry(vCopy, int(util.Abc_LitNot(iMigLit)))
	assert(iObjId >= 0, "iObjId >= 0")
	pObj = Abc_NtkCreateNodeInv(pNtk, Abc_NtkObj(pNtk, iObjId))
	Vec_IntWriteEntry(vCopy, int(iMigLit), Abc_ObjId(pObj))
	return pObj
}
func Mpm_ManDeriveMappedAbcNtk(p *Mpm_Man_t, pMio *mio.Mio_Library_t) *Abc_Ntk_t {
	var pNtk *Abc_Ntk_t
	var vNpnGatesMio *Vec_Ptr_t
	var vNodes, vCopy, vClass *Vec_Int_t
	var pObj, pFanin *Abc_Obj_t
	var pNode *Mig_Obj_t
	var pCutBest *Mpm_Cut_t
	var i, k, iNode, iMigLit, fCompl, Config int32

	// find mapping of SCL cells into MIO cells
	pScl, ok := p.pPars.pScl.(*scl.SC_Lib)
	assert(ok, "p.pPars.pScl is *scl.SC_Lib")
	vNpnGatesMio = Mpm_ManFindCells(pMio, pScl, p.vNpnConfigs)
	if vNpnGatesMio == nil {
		fmt.Printf("Genlib library does not match SCL library.\n")
		return nil
	}

	// create mapping for each phase of each node
	vCopy = Vec_IntStartFull(int(2 * Mig_ManObjNum(p.pMig)))

	// get internal nodes
	vNodes = Mpm_ManFindMappedNodes(p)

	// start the network
	pNtk = Abc_NtkAlloc(ABC_NTK_LOGIC, ABC_FUNC_MAP, 1)
	Abc_NtkSetName(pNtk, util.Abc_UtilStrsav(p.pMig.pName))
	pNtk.FunctionManager = pMio

	// create primary inputs
	// inlining of macro Mig_ManForEachCi
	for i = 0; int(i) < Vec_IntSize(&p.pMig.vCis); i++ {
		pNode = Mig_ManCi(p.pMig, i)
		pObj = Abc_NtkCreatePi(pNtk)
		Vec_IntWriteEntry(vCopy, int(util.Abc_Var2Lit(Mig_ObjId(pNode), 0)), Abc_ObjId(pObj))
	}
	Abc_NtkAddDummyPiNames(pNtk)

	// create constant nodes
	// inlining of macro Mig_ManForEachCo
	for i = 0; int(i) < Vec_IntSize(&p.pMig.vCos); i++ {
		pNode = Mig_ManCo(p.pMig, i)
		if Mig_ObjFaninLit(pNode, 0) == 0 {
			pObj = Abc_NtkCreateNodeConst0(pNtk)
			Vec_IntWriteEntry(vCopy, int(util.Abc_Var2Lit(0, 0)), Abc_ObjId(pObj))
			break
		}
	}
	// inlining of macro Mig_ManForEachCo
	for i = 0; int(i) < Vec_IntSize(&p.pMig.vCos); i++ {
		pNode = Mig_ManCo(p.pMig, i)
		if Mig_ObjFaninLit(pNode, 0) == 1 {
			pObj = Abc_NtkCreateNodeConst1(pNtk)
			Vec_IntWriteEntry(vCopy, int(util.Abc_Var2Lit(0, 1)), Abc_ObjId(pObj))
			break
		}
	}

	// create internal nodes
	// inlining of macro Vec_IntForEachEntry
	for i = 0; int(i) < Vec_IntSize(vNodes); i++ {
		iNode = Vec_IntEntry(vNodes, int(i))
		pCutBest = Mpm_ObjCutBestP(p, Mig_ManObj(p.pMig, iNode))
		vClass = Vec_WecEntry(p.vNpnConfigs, int(util.Abc_Lit2Var(int32(pCutBest.iFunc))))
		Config = Vec_IntEntry(vClass, 0)
		pObj = Abc_NtkCreateNode(pNtk)
		Abc_ObjSetData(pObj, Vec_PtrEntry(vNpnGatesMio, int(util.Abc_Lit2Var(int32(pCutBest.iFunc)))))
		assert(Abc_ObjData(pObj) != nil, "Abc_ObjData(pObj) != nil")
		fCompl = int32(pCutBest.fCompl) ^ util.Abc_LitIsCompl(int32(pCutBest.iFunc)) ^ ((Config >> 16) & 1)
		Config &= 0xFFFF
		for k = 0; k < int32(pCutBest.nLeaves); k++ {
			assert((Config>>6) < 720, "(Config >> 6) < 720")
			iMigLit = pCutBest.pLeaves[int32((p.Perm6[Config>>6][k]))]
			pFanin = Mpm_ManGetAbcNode(pNtk, vCopy, util.Abc_LitNotCond(iMigLit, (Config>>k)&1))
			Abc_ObjAddFanin(pObj, pFanin)
		}
		Vec_IntWriteEntry(vCopy, int(util.Abc_Var2Lit(iNode, fCompl)), Abc_ObjId(pObj))
	}

	// create primary outputs
	// inlining of macro Mig_ManForEachCo
	for i = 0; int(i) < Vec_IntSize(&p.pMig.vCos); i++ {
		pNode = Mig_ManCo(p.pMig, i)
		pObj = Abc_NtkCreatePo(pNtk)
		pFanin = Mpm_ManGetAbcNode(pNtk, vCopy, Mig_ObjFaninLit(pNode, 0))
		Abc_ObjAddFanin(pObj, pFanin)
	}
	Abc_NtkAddDummyPoNames(pNtk)

	// clean up
	Vec_PtrFree(vNpnGatesMio)
	Vec_IntFree(vNodes)
	Vec_IntFree(vCopy)
	return pNtk
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Mpm_ManPerformCellMapping(pMig *Mig_Man_t, pPars *Mpm_Par_t, pMio *mio.Mio_Library_t) *Abc_Ntk_t {
	var pNew *Abc_Ntk_t
	var p *Mpm_Man_t
	assert(pPars.fMap4Gates != 0, "pPars.fMap4Gates")
	p = Mpm_ManStart(pMig, pPars)
	if p.pPars.fVerbose != 0 {
		Mpm_ManPrintStatsInit(p)
	}
	p.vNpnConfigs = Mpm_ManFindDsdMatches(p, p.pPars.pScl)
	Mpm_ManPrepare(p)
	Mpm_ManPerform(p)
	if p.pPars.fVerbose != 0 {
		Mpm_ManPrintStats(p)
	}
	pNew = Mpm_ManDeriveMappedAbcNtk(p, pMio)
	Mpm_ManStop(p)
	return pNew
}
func Mpm_ManCellMapping(pGia *Gia_Man_t, pPars *Mpm_Par_t, pMio any) *Abc_Ntk_t {
	var p *Mig_Man_t
	var pNew *Abc_Ntk_t
	assert(pMio != nil, "pMio != nil")
	assert(pPars.pLib.LutMax <= MPM_VAR_MAX, "pPars.pLib.LutMax <= MPM_VAR_MAX")
	assert(pPars.nNumCuts <= MPM_CUT_MAX, "pPars.nNumCuts <= MPM_CUT_MAX")
	if pPars.fUseGates != 0 {
		pGia = gia.Gia_ManDupMuxes(pGia, 2)
		p = Mig_ManCreate(pGia)
		gia.Gia_ManStop(pGia)

	} else {
		p = Mig_ManCreate(pGia)
	}
	pMioLib, ok := pMio.(*mio.Mio_Library_t)
	assert(ok, "pMio is *mio.Mio_Library_t")
	pNew = Mpm_ManPerformCellMapping(p, pPars, pMioLib)
	Mig_ManStop(p)
	return pNew
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END
