package bac

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	abcapi "github.com/glycerine/abcgo/base/abc"
	abcmain "github.com/glycerine/abcgo/base/main"
	"github.com/glycerine/abcgo/bool/deco"
	"github.com/glycerine/abcgo/map/mio"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

/**CFile****************************************************************

  FileName    [bacBlast.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Hierarchical word-level netlist.]

  Synopsis    [Bit-blasting of the netlist.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - November 29, 2014.]

  Revision    [$Id: bacBlast.c,v 1.00 2014/11/29 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "bac.h"
//#include "base/abc/abc.h"
//#include "map/mio/mio.h"
//#include "bool/dec/dec.h"
//#include "base/main/mainInt.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Bac_ManPrepareGates(p *Bac_Man_t) {
	var ppGraphs []*deco.Dec_Graph_t
	var i int32
	if p.pMioLib == nil {
		return
	}
	ppGraphs = make([]*deco.Dec_Graph_t, int(util.Abc_NamObjNumMax(p.pMods)))
	lib := p.pMioLib.(*mio.Mio_Library_t)
	for i = 1; i < util.Abc_NamObjNumMax(p.pMods); i++ {
		var pGateName string = util.Abc_NamStr(p.pMods, i)
		var pGate *mio.Mio_Gate_t = mio.Mio_LibraryReadGateByName(lib, pGateName, "")
		if pGate != nil {
			ppGraphs[i] = deco.Dec_Factor(mio.Mio_GateReadSop(pGate))
		}
	}
	assert(p.ppGraphs == nil, "p.ppGraphs == nil")
	p.ppGraphs = ppGraphs
}
func Bac_ManUndoGates(p *Bac_Man_t) {
	var i int32
	if p.pMioLib == nil {
		return
	}
	for i = 1; i < util.Abc_NamObjNumMax(p.pMods); i++ {
		if p.ppGraphs[i] != nil {
			deco.Dec_GraphFree(p.ppGraphs[i])
		}
	}
	p.ppGraphs = nil
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Bac_ManAddBarbuf(pNew *Gia_Man_t, iRes int32, p *Bac_Man_t, iLNtk int32, iLObj int32, iRNtk int32, iRObj int32, vMap *vec.Vec_Int_t) int32 {
	var iBufLit, iIdLit int32
	if iRes == 0 || iRes == 1 {
		return iRes
	}
	assert(iRes > 0, "iRes > 0")
	iIdLit = -1
	if vMap != nil && int(Abc_Lit2Var(iRes)) < vec.Vec_IntSize(vMap) {
		iIdLit = vec.Vec_IntEntry(vMap, int(Abc_Lit2Var(iRes)))
	}
	if iIdLit >= 0 && vec.Vec_IntEntry(&p.vBuf2LeafNtk, int(Abc_Lit2Var(iIdLit))) == iLNtk && vec.Vec_IntEntry(&p.vBuf2RootNtk, int(Abc_Lit2Var(iIdLit))) == iRNtk {
		return util.Abc_LitNotCond(vec.Vec_IntEntry(gia.Gia_ManBarBufs(pNew), int(Abc_Lit2Var(iIdLit))), Abc_LitIsCompl(iRes)^Abc_LitIsCompl(iIdLit))
	}
	assert(Bac_ManNtkIsOk(p, iLNtk) != 0 && Bac_ManNtkIsOk(p, iRNtk) != 0, "Bac_ManNtkIsOk(p, iLNtk) && Bac_ManNtkIsOk(p, iRNtk)")
	vec.Vec_IntPush(&p.vBuf2LeafNtk, iLNtk)
	vec.Vec_IntPush(&p.vBuf2LeafObj, iLObj)
	vec.Vec_IntPush(&p.vBuf2RootNtk, iRNtk)
	vec.Vec_IntPush(&p.vBuf2RootObj, iRObj)
	iBufLit = gia.Gia_ManAppendBuf(pNew, iRes)
	if vMap != nil {
		bufs := gia.Gia_ManBarBufs(pNew)
		vec.Vec_IntSetEntryFull(vMap, int(Abc_Lit2Var(iRes)), Abc_Var2Lit(int32(vec.Vec_IntSize(bufs)), Abc_LitIsCompl(iRes)))
		vec.Vec_IntPush(bufs, iBufLit)
	}
	return iBufLit
}
func Bac_ManExtract_rec(pNew *Gia_Man_t, p *Bac_Ntk_t, i int32, fBuffers int32, vMap *vec.Vec_Int_t) int32 {
	var iRes int32 = Bac_ObjCopy(p, i)
	if iRes >= 0 {
		return iRes
	}
	if Bac_ObjIsCo(p, i) != 0 {
		iRes = Bac_ManExtract_rec(pNew, p, Bac_ObjFanin(p, i), fBuffers, vMap)
	} else if Bac_ObjIsPi(p, i) != 0 {
		var pHost *Bac_Ntk_t = Bac_NtkHostNtk(p)
		var iObj int32 = Bac_BoxBi(pHost, Bac_NtkHostObj(p), Bac_ObjIndex(p, i))
		iRes = Bac_ManExtract_rec(pNew, pHost, iObj, fBuffers, vMap)
		if fBuffers != 0 {
			iRes = Bac_ManAddBarbuf(pNew, iRes, p.pDesign, Bac_NtkId(p), i, Bac_NtkId(pHost), iObj, vMap)
		}

	} else if Bac_ObjIsBo(p, i) != 0 {
		var iBox int32 = Bac_BoxBoBox(p, i)
		if Bac_ObjIsBoxUser(p, iBox) != 0 { // user box

			var pBox *Bac_Ntk_t = Bac_BoxBoNtk(p, i)
			var iObj int32 = Bac_NtkPo(pBox, Bac_ObjIndex(p, i))
			iRes = Bac_ManExtract_rec(pNew, pBox, iObj, fBuffers, vMap)
			if fBuffers != 0 {
				iRes = Bac_ManAddBarbuf(pNew, iRes, p.pDesign, Bac_NtkId(p), i, Bac_NtkId(pBox), iObj, vMap)
			}

		} else { // primitive

			var iFanin int32
			var nLits int32
			var pLits [16]int32
			assert(Bac_ObjIsBoxPrim(p, iBox) != 0, "Bac_ObjIsBoxPrim(p, iBox)")
			for iFanin = 0; iBox-1-iFanin >= 0 && Bac_ObjIsBi(p, iBox-1-iFanin) != 0; iFanin++ {
				pLits[nLits] = Bac_ManExtract_rec(pNew, p, Bac_BoxFanin(p, iBox, nLits), fBuffers, vMap)
				nLits++
			}
			assert(nLits <= 16, "nLits <= 16")
			if p.pDesign.ppGraphs != nil { // mapped gate

				var pGraph *deco.Dec_Graph_t = p.pDesign.ppGraphs[int(Bac_BoxNtkId(p, iBox))]
				Leaves := *vec.Vec_IntAllocArray(pLits[:], int(nLits))
				assert(pGraph != nil, "pGraph != nil")
				return gia.Gia_ManFactorGraph(pNew, pGraph, &Leaves)

			} else {
				var Type Bac_ObjType_t = Bac_ObjType(p, iBox)
				if nLits == 0 {
					if Type == BAC_BOX_CF {
						iRes = 0
					} else if Type == BAC_BOX_CT {
						iRes = 1
					} else {
						assert(false, "0")
					}

				} else if nLits == 1 {
					if Type == BAC_BOX_BUF {
						iRes = pLits[0]
					} else if Type == BAC_BOX_INV {
						iRes = Abc_LitNot(pLits[0])
					} else {
						assert(false, "0")
					}

				} else if nLits == 2 {
					if Type == BAC_BOX_AND {
						iRes = gia.Gia_ManHashAnd(pNew, pLits[0], pLits[1])
					} else if Type == BAC_BOX_NAND {
						iRes = Abc_LitNot(gia.Gia_ManHashAnd(pNew, pLits[0], pLits[1]))
					} else if Type == BAC_BOX_OR {
						iRes = gia.Gia_ManHashOr(pNew, pLits[0], pLits[1])
					} else if Type == BAC_BOX_NOR {
						iRes = Abc_LitNot(gia.Gia_ManHashOr(pNew, pLits[0], pLits[1]))
					} else if Type == BAC_BOX_XOR {
						iRes = gia.Gia_ManHashXor(pNew, pLits[0], pLits[1])
					} else if Type == BAC_BOX_XNOR {
						iRes = Abc_LitNot(gia.Gia_ManHashXor(pNew, pLits[0], pLits[1]))
					} else if Type == BAC_BOX_SHARP {
						iRes = gia.Gia_ManHashAnd(pNew, pLits[0], Abc_LitNot(pLits[1]))
					} else if Type == BAC_BOX_SHARPL {
						iRes = gia.Gia_ManHashAnd(pNew, Abc_LitNot(pLits[0]), pLits[1])
					} else {
						assert(false, "0")
					}

				} else if nLits == 3 {
					if Type == BAC_BOX_MUX {
						iRes = gia.Gia_ManHashMux(pNew, pLits[0], pLits[1], pLits[2])
					} else if Type == BAC_BOX_MAJ {
						iRes = gia.Gia_ManHashMaj(pNew, pLits[0], pLits[1], pLits[2])
					} else if Type == BAC_BOX_ADD {
						var iRes0 int32 = gia.Gia_ManHashAnd(pNew, pLits[1], pLits[2])
						var iRes1 int32 = gia.Gia_ManHashOr(pNew, pLits[1], pLits[2])
						assert(Bac_BoxBoNum(p, iBox) == 2, "Bac_BoxBoNum(p, iBox) == 2")
						if Bac_BoxBo(p, iBox, 0) == i { // sum
							iRes = gia.Gia_ManHashXor(pNew, pLits[0], gia.Gia_ManHashAnd(pNew, Abc_LitNot(iRes0), iRes1))
						} else if Bac_BoxBo(p, iBox, 1) == i { // cout
							iRes = gia.Gia_ManHashOr(pNew, iRes0, gia.Gia_ManHashAnd(pNew, pLits[0], iRes1))
						} else {
							assert(false, "0")
						}

					} else {
						assert(false, "0")
					}

				} else {
					assert(false, "0")
				}
			}
		}

	} else {
		assert(false, "0")
	}
	Bac_ObjSetCopy(p, i, iRes)
	return iRes
}
func Bac_ManExtract(p *Bac_Man_t, fBuffers int32, fVerbose int32) *Gia_Man_t {
	var pNtk *Bac_Ntk_t
	var pRoot *Bac_Ntk_t = Bac_ManRoot(p)
	var pNew, pTemp *Gia_Man_t
	var vMap *vec.Vec_Int_t = nil
	var i, iObj int32

	vec.Vec_IntClear(&p.vBuf2LeafNtk)
	vec.Vec_IntClear(&p.vBuf2LeafObj)
	vec.Vec_IntClear(&p.vBuf2RootNtk)
	vec.Vec_IntClear(&p.vBuf2RootObj)

	for i = 1; i <= Bac_ManNtkNum(p); i++ {
		pNtk = Bac_ManNtk(p, i)
		Bac_NtkDeriveIndex(pNtk)
		Bac_NtkStartCopies(pNtk)
	}

	// start the manager
	pNew = gia.Gia_ManStart(Bac_ManNodeNum(p))
	gia.Gia_ManSetNameSpec(pNew, Abc_UtilStrsav(p.pName), Abc_UtilStrsav(p.pSpec))

	// primary inputs
	for i = 0; i < Bac_NtkPiNum(pRoot); i++ {
		iObj = Bac_NtkPi(pRoot, i)
		Bac_ObjSetCopy(pRoot, iObj, gia.Gia_ManAppendCi(pNew))
	}

	// internal nodes
	gia.Gia_ManHashAlloc(pNew)
	gia.Gia_ManBarBufs(pNew)
	vMap = vec.Vec_IntStartFull(10000)
	Bac_ManPrepareGates(p)
	for i = 0; i < Bac_NtkPoNum(pRoot); i++ {
		iObj = Bac_NtkPo(pRoot, i)
		Bac_ManExtract_rec(pNew, pRoot, iObj, fBuffers, vMap)
	}
	Bac_ManUndoGates(p)
	vec.Vec_IntFreeP(&vMap)
	gia.Gia_ManHashStop(pNew)

	// primary outputs
	for i = 0; i < Bac_NtkPoNum(pRoot); i++ {
		iObj = Bac_NtkPo(pRoot, i)
		gia.Gia_ManAppendCo(pNew, Bac_ObjCopy(pRoot, iObj))
	}
	assert(int32(vec.Vec_IntSize(&p.vBuf2LeafNtk)) == gia.Gia_ManBufNum(pNew), "vec.Vec_IntSize(&p.vBuf2LeafNtk) == Gia_ManBufNum(pNew)")

	// cleanup
	pTemp = pNew
	pNew = gia.Gia_ManCleanup(pTemp)
	gia.Gia_ManStop(pTemp)
	//gia.Gia_ManPrintStats( pNew, NULL );
	return pNew
}

/*
*Function*************************************************************

	Synopsis    [Mark each GIA node with the network it belongs to.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Bac_ManMarkNodesGia(p *Bac_Man_t, pGia *Gia_Man_t) {
	var pObj *Gia_Obj_t
	var i, Count int32
	assert(int32(vec.Vec_IntSize(&p.vBuf2LeafNtk)) == gia.Gia_ManBufNum(pGia), "vec.Vec_IntSize(&p.vBuf2LeafNtk) == gia.Gia_ManBufNum(pGia)")
	gia.Gia_ManConst0(pGia).Value = 1
	for i = 0; i < gia.Gia_ManPiNum(pGia); i++ {
		gia.Gia_ManPi(pGia, i).Value = 1
	}
	for i = 1; i < gia.Gia_ManObjNum(pGia); i++ {
		pObj = gia.Gia_ManObj(pGia, i)
		if gia.Gia_ObjIsAnd(pObj) == 0 {
			continue
		}
		if gia.Gia_ObjIsBuf(pObj) != 0 {
			pObj.Value = uint32(vec.Vec_IntEntry(&p.vBuf2LeafNtk, int(Count)))
			Count++
		} else {
			pObj.Value = gia.Gia_ObjFanin0(pObj).Value
			assert(pObj.Value == gia.Gia_ObjFanin1(pObj).Value, "pObj.Value == gia.Gia_ObjFanin1(pObj).Value")
		}
	}
	assert(Count == gia.Gia_ManBufNum(pGia), "Count == gia.Gia_ManBufNum(pGia)")
	for i = 0; i < gia.Gia_ManPoNum(pGia); i++ {
		pObj = gia.Gia_ManPo(pGia, i)
		assert(gia.Gia_ObjFanin0(pObj).Value == 1, "gia.Gia_ObjFanin0(pObj).Value == 1")
		pObj.Value = 1
	}
}
func Bac_ManRemapBarbufs(pNew *Bac_Man_t, p *Bac_Man_t) {
	var pNtk *Bac_Ntk_t
	var Entry, i int32
	//assert( vec.Vec_IntSize(&p->vBuf2RootNtk) );
	assert(vec.Vec_IntSize(&pNew.vBuf2RootNtk) == 0, "vec.Vec_IntSize(&pNew.vBuf2RootNtk) == 0")
	vec.Vec_IntAppend(&pNew.vBuf2RootNtk, &p.vBuf2RootNtk)
	vec.Vec_IntAppend(&pNew.vBuf2RootObj, &p.vBuf2RootObj)
	vec.Vec_IntAppend(&pNew.vBuf2LeafNtk, &p.vBuf2LeafNtk)
	vec.Vec_IntAppend(&pNew.vBuf2LeafObj, &p.vBuf2LeafObj)
	for i = 0; i < int32(vec.Vec_IntSize(&p.vBuf2LeafObj)); i++ {
		Entry = vec.Vec_IntEntry(&p.vBuf2LeafObj, int(i))
		pNtk = Bac_ManNtk(p, vec.Vec_IntEntry(&p.vBuf2LeafNtk, int(i)))
		vec.Vec_IntWriteEntry(&pNew.vBuf2LeafObj, int(i), Bac_ObjCopy(pNtk, Entry))
	}
	for i = 0; i < int32(vec.Vec_IntSize(&p.vBuf2RootObj)); i++ {
		Entry = vec.Vec_IntEntry(&p.vBuf2RootObj, int(i))
		pNtk = Bac_ManNtk(p, vec.Vec_IntEntry(&p.vBuf2RootNtk, int(i)))
		vec.Vec_IntWriteEntry(&pNew.vBuf2RootObj, int(i), Bac_ObjCopy(pNtk, Entry))
	}
}
func Bac_NtkCreateAndConnectBuffer(pGia *Gia_Man_t, pObj *Gia_Obj_t, p *Bac_Ntk_t, iTerm int32) {
	var iObj int32
	if pGia != nil && gia.Gia_ObjFaninId0p(pGia, pObj) > 0 {
		iObj = Bac_ObjAlloc(p, BAC_OBJ_BI, int32(gia.Gia_ObjFanin0(pObj).Value))
		typ := BAC_BOX_BUF
		if gia.Gia_ObjFaninC0(pObj) != 0 {
			typ = BAC_BOX_INV
		}
		Bac_ObjAlloc(p, typ, -1)

	} else {
		typ := BAC_BOX_CF
		if pGia != nil && gia.Gia_ObjFaninC0(pObj) != 0 {
			typ = BAC_BOX_CT
		}
		Bac_ObjAlloc(p, typ, -1)
	}
	iObj = Bac_ObjAlloc(p, BAC_OBJ_BO, -1)
	Bac_ObjSetFanin(p, iTerm, iObj)
}
func Bac_NtkInsertGia(p *Bac_Man_t, pGia *Gia_Man_t) {
	var pNtk *Bac_Ntk_t
	var pRoot *Bac_Ntk_t = Bac_ManRoot(p)
	var i, j, k, iBox, iTerm, Count int32
	var pObj *Gia_Obj_t

	gia.Gia_ManConst0(pGia).Value = ^uint32(0)
	for i = 0; i < gia.Gia_ManPiNum(pGia); i++ {
		gia.Gia_ManPi(pGia, i).Value = uint32(Bac_NtkPi(pRoot, i))
	}
	for i = 1; i < gia.Gia_ManObjNum(pGia); i++ {
		pObj = gia.Gia_ManObj(pGia, i)
		if gia.Gia_ObjIsAnd(pObj) == 0 {
			continue
		}
		if gia.Gia_ObjIsBuf(pObj) != 0 {
			pNtk = Bac_ManNtk(p, vec.Vec_IntEntry(&p.vBuf2RootNtk, int(Count)))
			iTerm = vec.Vec_IntEntry(&p.vBuf2RootObj, int(Count))
			assert(Bac_ObjIsCo(pNtk, iTerm) != 0, "Bac_ObjIsCo(pNtk, iTerm)")
			if Bac_ObjFanin(pNtk, iTerm) == -1 { // not a feedthrough
				Bac_NtkCreateAndConnectBuffer(pGia, pObj, pNtk, iTerm)
			}
			// prepare leaf
			pObj.Value = uint32(vec.Vec_IntEntry(&p.vBuf2LeafObj, int(Count)))
			Count++

		} else {
			var iLit0 int32 = int32(gia.Gia_ObjFanin0(pObj).Value)
			var iLit1 int32 = int32(gia.Gia_ObjFanin1(pObj).Value)
			var Type Bac_ObjType_t
			pNtk = Bac_ManNtk(p, int32(pObj.Value))
			if gia.Gia_ObjFaninC0(pObj) != 0 && gia.Gia_ObjFaninC1(pObj) != 0 {
				Type = BAC_BOX_NOR
			} else if gia.Gia_ObjFaninC1(pObj) != 0 {
				Type = BAC_BOX_SHARP
			} else if gia.Gia_ObjFaninC0(pObj) != 0 {
				Type = BAC_BOX_SHARP
				iLit0, iLit1 = iLit1, iLit0

			} else {
				Type = BAC_BOX_AND
			}
			// create box
			iTerm = Bac_ObjAlloc(pNtk, BAC_OBJ_BI, iLit1)
			iTerm = Bac_ObjAlloc(pNtk, BAC_OBJ_BI, iLit0)
			Bac_ObjAlloc(pNtk, Type, -1)
			pObj.Value = uint32(Bac_ObjAlloc(pNtk, BAC_OBJ_BO, -1))
		}
	}
	assert(Count == gia.Gia_ManBufNum(pGia), "Count == gia.Gia_ManBufNum(pGia)")

	// create constant 0 drivers for COs without barbufs
	for i = 1; i <= Bac_ManNtkNum(p); i++ {
		pNtk = Bac_ManNtk(p, i)
		for iBox = 0; iBox < Bac_NtkObjNum(pNtk); iBox++ {
			if Bac_ObjIsBox(pNtk, iBox) == 0 {
				continue
			}
			for j = 0; j < Bac_BoxBiNum(pNtk, iBox); j++ {
				iTerm = Bac_BoxBi(pNtk, iBox, j)
				if Bac_ObjFanin(pNtk, iTerm) == -1 {
					Bac_NtkCreateAndConnectBuffer(nil, nil, pNtk, iTerm)
				}
			}
		}
		for k = 0; k < Bac_NtkPoNum(pNtk); k++ {
			iTerm = Bac_NtkPo(pNtk, k)
			if pNtk != pRoot && Bac_ObjFanin(pNtk, iTerm) == -1 {
				Bac_NtkCreateAndConnectBuffer(nil, nil, pNtk, iTerm)
			}
		}
	}
	// create node and connect POs
	for i = 0; i < gia.Gia_ManPoNum(pGia); i++ {
		pObj = gia.Gia_ManPo(pGia, i)
		if Bac_ObjFanin(pRoot, Bac_NtkPo(pRoot, i)) == -1 { // not a feedthrough
			Bac_NtkCreateAndConnectBuffer(pGia, pObj, pRoot, Bac_NtkPo(pRoot, i))
		}
	}
}
func Bac_ManInsertGia(p *Bac_Man_t, pGia *Gia_Man_t) *Bac_Man_t {
	var pNew *Bac_Man_t = Bac_ManDupUserBoxes(p)
	Bac_ManMarkNodesGia(p, pGia)
	Bac_ManRemapBarbufs(pNew, p)
	Bac_NtkInsertGia(pNew, pGia)
	Bac_ManMoveNames(pNew, p)
	return pNew
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Bac_ManBlastTest(p *Bac_Man_t) *Bac_Man_t {
	var pGia *Gia_Man_t = Bac_ManExtract(p, 1, 0)
	var pNew *Bac_Man_t = Bac_ManInsertGia(p, pGia)
	gia.Gia_ManStop(pGia)
	return pNew
}

/*
*Function*************************************************************

	Synopsis    [Mark each GIA node with the network it belongs to.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NodeIsSeriousGate(p *Abc_Obj_t) int32 {
	if abcapi.Abc_ObjIsNode(p) != 0 && abcapi.Abc_ObjFaninNum(p) > 0 && abcapi.Abc_ObjIsBarBuf(p) == 0 {
		return 1
	}
	return 0
}
func Bac_ManMarkNodesAbc(p *Bac_Man_t, pNtk *Abc_Ntk_t) {
	var pObj *Abc_Obj_t
	var i, k, Count int32
	assert(int32(vec.Vec_IntSize(&p.vBuf2LeafNtk)) == abcapi.Abc_NtkBarBufNum(pNtk), "vBuf2LeafNtk size matches barrier count")
	for i = 0; i < abcapi.Abc_NtkPiNum(pNtk); i++ {
		abcapi.Abc_ObjSetTemp(abcapi.Abc_NtkPi(pNtk, i), 1)
	}
	for _, pObj = range abcapi.Abc_NtkNodes(pNtk) {
		if abcapi.Abc_ObjIsBarBuf(pObj) != 0 {
			abcapi.Abc_ObjSetTemp(pObj, vec.Vec_IntEntry(&p.vBuf2LeafNtk, int(Count)))
			Count++
		} else if Abc_NodeIsSeriousGate(pObj) != 0 {
			abcapi.Abc_ObjSetTemp(pObj, abcapi.Abc_ObjTemp(abcapi.Abc_ObjFanin0(pObj)))
			for k = 0; k < abcapi.Abc_ObjFaninNum(pObj); k++ {
				pFanin := abcapi.Abc_ObjFanin(pObj, k)
				assert(abcapi.Abc_ObjTemp(pObj) == abcapi.Abc_ObjTemp(pFanin), "node fanin network marks match")
			}
		}
	}
	for i = 0; i < abcapi.Abc_NtkPoNum(pNtk); i++ {
		pObj = abcapi.Abc_NtkPo(pNtk, i)
		if Abc_NodeIsSeriousGate(abcapi.Abc_ObjFanin0(pObj)) != 0 {
			assert(abcapi.Abc_ObjTemp(abcapi.Abc_ObjFanin0(pObj)) == 1, "PO gate belongs to root network")
			abcapi.Abc_ObjSetTemp(pObj, abcapi.Abc_ObjTemp(abcapi.Abc_ObjFanin0(pObj)))
		}
	}
	assert(Count == abcapi.Abc_NtkBarBufNum(pNtk), "barrier count matches")
}
func Bac_NtkCreateOrConnectFanin(pFanin *Abc_Obj_t, p *Bac_Ntk_t, iTerm int32) {
	var iObj int32
	if pFanin != nil && (Abc_NodeIsSeriousGate(pFanin) != 0) { //&& Bac_ObjName(p, pFanin->iTemp) == -1 ) // gate without name

		iObj = abcapi.Abc_ObjTemp(pFanin)

	} else if pFanin != nil && (abcapi.Abc_ObjIsPi(pFanin) != 0 || abcapi.Abc_ObjIsBarBuf(pFanin) != 0 || (Abc_NodeIsSeriousGate(pFanin) != 0)) { // PI/BO or gate with name

		iObj = Bac_ObjAlloc(p, BAC_OBJ_BI, abcapi.Abc_ObjTemp(pFanin))
		Bac_ObjAlloc(p, BAC_BOX_GATE, p.pDesign.ElemGates[2]) // buffer
		iObj = Bac_ObjAlloc(p, BAC_OBJ_BO, -1)

	} else {
		assert(pFanin == nil || abcapi.Abc_NodeIsConst0(pFanin) != 0 || abcapi.Abc_NodeIsConst1(pFanin) != 0, "!pFanin || abcapi.Abc_NodeIsConst0(pFanin) != 0 || abcapi.Abc_NodeIsConst1(pFanin) != 0")
		idx := 0
		if pFanin != nil && abcapi.Abc_NodeIsConst1(pFanin) != 0 {
			idx = 1
		}
		Bac_ObjAlloc(p, BAC_BOX_GATE, p.pDesign.ElemGates[idx]) // const 0/1
		iObj = Bac_ObjAlloc(p, BAC_OBJ_BO, -1)
	}
	Bac_ObjSetFanin(p, iTerm, iObj)
}
func Bac_NtkPrepareLibrary(p *Bac_Man_t, pLib *Mio_Library_t) {
	pGate0 := mio.Mio_LibraryReadConst0(pLib)
	pGate1 := mio.Mio_LibraryReadConst1(pLib)
	pGate2 := mio.Mio_LibraryReadBuf(pLib)
	if !(pGate0 != nil) || !(pGate1 != nil) || !(pGate2 != nil) {
		fmt.Printf("The library does not have one of the elementary gates.\n")
		return
	}
	p.ElemGates[0] = util.Abc_NamStrFindOrAdd(p.pMods, mio.Mio_GateReadName(pGate0), nil)
	p.ElemGates[1] = util.Abc_NamStrFindOrAdd(p.pMods, mio.Mio_GateReadName(pGate1), nil)
	p.ElemGates[2] = util.Abc_NamStrFindOrAdd(p.pMods, mio.Mio_GateReadName(pGate2), nil)
	for _, pGate := range mio.Mio_LibraryReadGateArray(pLib) {
		if pGate != nil && pGate != pGate0 && pGate != pGate1 && pGate != pGate2 {
			util.Abc_NamStrFindOrAdd(p.pMods, mio.Mio_GateReadName(pGate), nil)
		}
	}
	assert(util.Abc_NamObjNumMax(p.pMods) > 1, "Abc_NamObjNumMax(p.pMods) > 1")
}
func Bac_NtkBuildLibrary(p *Bac_Man_t) int32 {
	var RetValue int32 = 1
	var pLib *Mio_Library_t
	if lib, ok := abcmain.Abc_FrameReadLibGen().(*mio.Mio_Library_t); ok {
		pLib = lib
	}
	if pLib == nil {
		fmt.Printf("The standard cell library is not available.\n")
		RetValue = 0
	} else {
		Bac_NtkPrepareLibrary(p, pLib)
	}
	p.pMioLib = pLib
	return RetValue
}
func Bac_NtkInsertNtk(p *Bac_Man_t, pNtk *Abc_Ntk_t) {
	var pCbaNtk *Bac_Ntk_t
	var pRoot *Bac_Ntk_t = Bac_ManRoot(p)
	var i, j, k, iBox, iTerm, Count int32
	var pObj *Abc_Obj_t
	assert(abcapi.Abc_NtkHasMapping(pNtk) != 0, "abcapi.Abc_NtkHasMapping(pNtk)")
	lib := pNtk.FunctionManager.(*mio.Mio_Library_t)
	Bac_NtkPrepareLibrary(p, lib)
	p.pMioLib = lib

	for i = 0; i < abcapi.Abc_NtkPiNum(pNtk); i++ {
		abcapi.Abc_ObjSetTemp(abcapi.Abc_NtkPi(pNtk, i), Bac_NtkPi(pRoot, i))
	}
	for _, pObj = range abcapi.Abc_NtkNodes(pNtk) {
		if abcapi.Abc_ObjIsBarBuf(pObj) != 0 {
			pCbaNtk = Bac_ManNtk(p, vec.Vec_IntEntry(&p.vBuf2RootNtk, int(Count)))
			iTerm = vec.Vec_IntEntry(&p.vBuf2RootObj, int(Count))
			assert(Bac_ObjIsCo(pCbaNtk, iTerm) != 0, "Bac_ObjIsCo(pCbaNtk, iTerm)")
			if Bac_ObjFanin(pCbaNtk, iTerm) == -1 {
				Bac_NtkCreateOrConnectFanin(abcapi.Abc_ObjFanin0(pObj), pCbaNtk, iTerm)
			}
			abcapi.Abc_ObjSetTemp(pObj, vec.Vec_IntEntry(&p.vBuf2LeafObj, int(Count)))
			Count++
		} else if Abc_NodeIsSeriousGate(pObj) != 0 {
			pCbaNtk = Bac_ManNtk(p, abcapi.Abc_ObjTemp(pObj))
			for k = abcapi.Abc_ObjFaninNum(pObj) - 1; k >= 0; k-- {
				iTerm = Bac_ObjAlloc(pCbaNtk, BAC_OBJ_BI, abcapi.Abc_ObjTemp(abcapi.Abc_ObjFanin(pObj, k)))
			}
			gate := pObj.Data.(*mio.Mio_Gate_t)
			Bac_ObjAlloc(pCbaNtk, BAC_BOX_GATE, util.Abc_NamStrFind(p.pMods, mio.Mio_GateReadName(gate)))
			abcapi.Abc_ObjSetTemp(pObj, Bac_ObjAlloc(pCbaNtk, BAC_OBJ_BO, -1))
		}
	}
	assert(Count == abcapi.Abc_NtkBarBufNum(pNtk), "barrier count matches")

	// create constant 0 drivers for COs without barbufs
	for i = 1; i <= Bac_ManNtkNum(p); i++ {
		pCbaNtk = Bac_ManNtk(p, i)
		for iBox = 0; iBox < Bac_NtkObjNum(pCbaNtk); iBox++ {
			if Bac_ObjIsBox(pCbaNtk, iBox) == 0 {
				continue
			}
			for j = 0; j < Bac_BoxBiNum(pCbaNtk, iBox); j++ {
				iTerm = Bac_BoxBi(pCbaNtk, iBox, j)
				if Bac_ObjFanin(pCbaNtk, iTerm) == -1 {
					Bac_NtkCreateOrConnectFanin(nil, pCbaNtk, iTerm)
				}
			}
		}
		for k = 0; k < Bac_NtkPoNum(pCbaNtk); k++ {
			iTerm = Bac_NtkPo(pCbaNtk, k)
			if pCbaNtk != pRoot && Bac_ObjFanin(pCbaNtk, iTerm) == -1 {
				Bac_NtkCreateOrConnectFanin(nil, pCbaNtk, iTerm)
			}
		}
	}
	// create node and connect POs
	for i = 0; i < abcapi.Abc_NtkPoNum(pNtk); i++ {
		pObj = abcapi.Abc_NtkPo(pNtk, i)
		if Bac_ObjFanin(pRoot, Bac_NtkPo(pRoot, i)) == -1 { // not a feedthrough
			Bac_NtkCreateOrConnectFanin(abcapi.Abc_ObjFanin0(pObj), pRoot, Bac_NtkPo(pRoot, i))
		}
	}
}
func Bac_ManInsertAbc(p *Bac_Man_t, pAbc any) any {
	var pNtk *Abc_Ntk_t = pAbc.(*Abc_Ntk_t)
	var pNew *Bac_Man_t = Bac_ManDupUserBoxes(p)
	Bac_ManMarkNodesAbc(p, pNtk)
	Bac_ManRemapBarbufs(pNew, p)
	Bac_NtkInsertNtk(pNew, pNtk)
	Bac_ManMoveNames(pNew, p)
	return pNew
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END
