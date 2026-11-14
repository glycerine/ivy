package abci

import map_mio "github.com/glycerine/abcgo/map/mio"

import proof_pdr "github.com/glycerine/abcgo/proof/pdr"
import "github.com/glycerine/abcgo/proof/aiginterp"
import "github.com/glycerine/abcgo/proof/aigpartsat"

import aig_ioa "github.com/glycerine/abcgo/aig/ioa"

import aig_saig "github.com/glycerine/abcgo/aig/saig"

import base_main "github.com/glycerine/abcgo/base/main"

import bdd_bbr "github.com/glycerine/abcgo/bdd/bbr"

import misc_util "github.com/glycerine/abcgo/misc/util"

import opt_cgt "github.com/glycerine/abcgo/opt/cgt"

import opt_csw "github.com/glycerine/abcgo/opt/csw"

import opt_dar "github.com/glycerine/abcgo/opt/dar"

import proof_cec "github.com/glycerine/abcgo/proof/cec"

import proof_dch "github.com/glycerine/abcgo/proof/dch"

import proof_fra "github.com/glycerine/abcgo/proof/fra"

import proof_fraig "github.com/glycerine/abcgo/proof/fraig"

import proof_int "github.com/glycerine/abcgo/proof/int"

import proof_ssw "github.com/glycerine/abcgo/proof/ssw"

import sat_bmc "github.com/glycerine/abcgo/sat/bmc"

import sat_cnf "github.com/glycerine/abcgo/sat/cnf"

import (
	"fmt"
	"os"
	"time"

	aiggraph "github.com/glycerine/abcgo/aig/aig"
	giapkg "github.com/glycerine/abcgo/aig/gia"
	hoppkg "github.com/glycerine/abcgo/aig/hop"
	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/map/amap"
	"github.com/glycerine/abcgo/misc/vec"
)

/**CFile****************************************************************

  FileName    [abcDar.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Network and node package.]

  Synopsis    [DAG-aware rewriting.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: abcDar.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "base/abc/abc.h"
//#include "base/main/main.h"
//#include "aig/gia/giaAig.h"
//#include "opt/dar/dar.h"
//#include "sat/cnf/cnf.h"
//#include "proof/fra/fra.h"
//#include "proof/fraig/fraig.h"
//#include "proof/int/int.h"
//#include "proof/dch/dch.h"
//#include "proof/ssw/ssw.h"
//#include "opt/cgt/cgt.h"
//#include "bdd/bbr/bbr.h"
//#include "aig/gia/gia.h"
//#include "proof/cec/cec.h"
//#include "opt/csw/csw.h"
//#include "proof/pdr/pdr.h"
//#include "sat/bmc/bmc.h"
//#include "map/mio/mio.h"

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
func Abc_ObjCompareById(pp1 *any, pp2 *any) int {
	return int(abc.Abc_ObjId(abc.Abc_ObjRegular(*pp1)) - abc.Abc_ObjId(abc.Abc_ObjRegular(*pp2)))
}

/*
*Function*************************************************************

	Synopsis    [Collects the supergate.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_CollectTopOr_rec(pObj any, vSuper *Vec_Ptr_t) {
	regular := abc.Abc_ObjRegular(pObj)
	if abc.Abc_ObjIsComplement(pObj) != 0 || abc.Abc_ObjIsNode(regular) == 0 {
		Vec_PtrPush(vSuper, pObj)
		return
	}
	// go through the branches
	Abc_CollectTopOr_rec(abc.Abc_ObjChild0(regular), vSuper)
	Abc_CollectTopOr_rec(abc.Abc_ObjChild1(regular), vSuper)
}
func Abc_CollectTopOr(pObj any, vSuper *Vec_Ptr_t) {
	Vec_PtrClear(vSuper)
	if abc.Abc_ObjIsComplement(pObj) != 0 {
		Abc_CollectTopOr_rec(abc.Abc_ObjNot(pObj), vSuper)
		giapkg.Vec_PtrUniqify(vSuper, Abc_ObjCompareById)

	} else {
		Vec_PtrPush(vSuper, abc.Abc_ObjNot(pObj))
	}
}

/*
*Function*************************************************************

	Synopsis    [Converts the network from the AIG manager into ABC.]

	Description [The returned map maps new PO IDs into old ones.]

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkToDarBmc(pNtk *Abc_Ntk_t, pvMap **Vec_Int_t) *Aig_Man_t {
	if pNtk == nil || abc.Abc_NtkIsStrash(pNtk) == 0 {
		return nil
	}
	for _, latch := range pNtk.ObjectsOfKind(abc.ObjectLatch) {
		if abc.Abc_LatchIsInitDc(latch) != 0 {
			abc.Abc_LatchSetInit0(latch)
			fmt.Println("Warning: don't-care latch initialization is assumed to be 0.")
		}
	}

	// Build the ordinary AIG first. Its Temp map identifies the AIG literal
	// for each source object and its output tail contains latch inputs.
	source := abc.Abc_NtkToDar(pNtk, 0, 1)
	if source == nil {
		return nil
	}
	defer aiggraph.Aig_ManStop(source)
	result := aiggraph.Aig_ManDupWithoutPos(source)
	result.Name, result.Spec = source.Name, source.Spec
	result.BarrierBuffers = source.BarrierBuffers
	result.Assertions = source.Assertions
	aiggraph.Aig_ManSetConstrNum(result, pNtk.Constraints)

	var mapping *Vec_Int_t
	if pvMap != nil {
		mapping = Vec_IntAlloc(100)
		*pvMap = mapping
	}
	for i := int32(0); i < abc.Abc_NtkPoNum(pNtk); i++ {
		po := abc.Abc_NtkPo(pNtk, i)
		if po == nil || len(po.Fanins()) != 1 {
			return nil
		}
		driver := po.Fanins()[0]
		terms := make([]abc.Edge, 0, 2)
		if pNtk.Constraints != 0 && i >= pNtk.Constraints {
			terms = append(terms, driver.Not())
		} else {
			vSuper := Vec_PtrAlloc(4)
			Abc_CollectTopOr(driver, vSuper)
			for _, item := range giapkg.Vec_PtrArray(vSuper) {
				e, ok := item.(abc.Edge)
				if !ok {
					Vec_PtrFree(vSuper)
					return nil
				}
				terms = append(terms, e)
			}
			Vec_PtrFree(vSuper)
		}
		for _, term := range terms {
			lit, ok := Abc_NtkDarObjLit(term.Object)
			if !ok {
				return nil
			}
			if term.Complement {
				lit = lit.Not()
			}
			lit = aiggraph.Aig_ObjRealCopy(source, lit)
			result.AddOutput(lit)
			if mapping != nil {
				Vec_IntPush(mapping, i)
			}
		}
	}

	// Preserve the source's latch-input CO order after expanded POs.
	outputs := source.Outputs()
	poCount := int(abc.Abc_NtkPoNum(pNtk))
	regCount := int(abc.Abc_NtkLatchNum(pNtk))
	if len(outputs) < poCount+regCount {
		return nil
	}
	for _, output := range outputs[len(outputs)-regCount:] {
		driver := source.Node(output).Fanins[0]
		result.AddOutput(aiggraph.Aig_ObjRealCopy(source, driver))
	}
	aiggraph.Aig_ManSetRegNum(result, int32(regCount))
	aiggraph.Aig_ManCleanup(result)
	if err := result.Check(); err != nil {
		fmt.Printf("Abc_NtkToDarBmc: AIG check has failed: %v\n", err)
		aiggraph.Aig_ManStop(result)
		return nil
	}
	return result
}

/*
*Function*************************************************************

	Synopsis    [Collects information about what flops have unknown values.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkFindDcLatches(pNtk *Abc_Ntk_t) *Vec_Int_t {
	unknown := Vec_IntStart(abc.Abc_NtkLatchNum(pNtk))
	for i, latch := range pNtk.ObjectsOfKind(abc.ObjectLatch) {
		if abc.Abc_LatchIsInitDc(latch) != 0 {
			Vec_IntWriteEntry(unknown, int32(i), 1)
			abc.Abc_LatchSetInit0(latch)
		}
	}
	return unknown
}

/*
*Function*************************************************************

	Synopsis    [Converts the network from the AIG manager into ABC.]

	Description [Assumes that registers are ordered after PIs/POs.]

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkToDar(pNtk *Abc_Ntk_t, fExors int32, fRegisters int32) *Aig_Man_t {
	return abc.Abc_NtkToDar(pNtk, fExors, fRegisters)
}

/*
*Function*************************************************************

	Synopsis    [Converts the network from the AIG manager into ABC.]

	Description [Assumes that registers are ordered after PIs/POs.]

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkToDarChoices(pNtk *Abc_Ntk_t) *Aig_Man_t {
	if pNtk == nil || abc.Abc_NtkIsStrash(pNtk) == 0 {
		return nil
	}
	for _, obj := range pNtk.Objects() {
		obj.Temp = -1
	}
	pMan := aiggraph.NewGraph()
	pMan.Name, pMan.Spec = pNtk.Name, pNtk.Spec
	pMan.Assertions = pNtk.Constraints
	pMan.BarrierBuffers = int32(pNtk.Barriers)
	abc.Abc_AigConst1(pNtk).Object.Temp = int(pMan.Const1())
	for i := int32(0); i < abc.Abc_NtkCiNum(pNtk); i++ {
		ci := abc.Abc_NtkCi(pNtk, i)
		ci.Temp = int(pMan.NewInput())
	}
	for _, node := range abc.Abc_AigDfs(pNtk, 0, 0) {
		if node.Kind != abc.ObjectNode || len(node.Fanins()) != 2 {
			continue
		}
		fanins := node.Fanins()
		child := func(edge abc.Edge) aiggraph.Edge {
			if edge.Object.Temp < 0 {
				panic("abc: choice DFS did not visit a fanin")
			}
			lit := aiggraph.Edge(uint32(edge.Object.Temp))
			if edge.Complement {
				lit = lit.Not()
			}
			return lit
		}
		copy := pMan.AND(child(fanins[0]), child(fanins[1]))
		node.Temp = int(copy)
		previous := copy
		for choice := abc.Abc_ObjChoice(node); choice != nil; choice = abc.Abc_ObjChoice(choice) {
			if choice.Temp < 0 {
				panic("abc: choice DFS did not map an alternative")
			}
			alternative := aiggraph.Edge(uint32(choice.Temp))
			aiggraph.Aig_ObjSetEquiv(pMan, previous, alternative)
			previous = alternative
		}
	}
	for i := int32(0); i < abc.Abc_NtkCoNum(pNtk); i++ {
		co := abc.Abc_NtkCo(pNtk, i)
		fanins := co.Fanins()
		if len(fanins) != 1 {
			return nil
		}
		pMan.AddOutput(abcChoiceMappedEdge(fanins[0]))
	}
	if err := pMan.Check(); err != nil {
		return nil
	}
	return pMan
}

func abcChoiceMappedEdge(edge abc.Edge) aiggraph.Edge {
	if edge.Object == nil || edge.Object.Temp < 0 {
		panic("abc: choice output was not mapped")
	}
	lit := aiggraph.Edge(uint32(edge.Object.Temp))
	if edge.Complement {
		lit = lit.Not()
	}
	return lit
}

/*
*Function*************************************************************

	Synopsis    [Converts the network from the AIG manager into ABC.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkFromDar(pNtkOld *Abc_Ntk_t, pMan *Aig_Man_t) *Abc_Ntk_t {
	return abc.Abc_NtkFromDarSeqSweep(pNtkOld, pMan)
}

/*
*Function*************************************************************

	Synopsis    [Converts the network from the AIG manager into ABC.]

	Description [This procedure should be called after seq sweeping,
	which changes the number of registers.]

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkFromDarSeqSweep(pNtkOld *Abc_Ntk_t, pMan *Aig_Man_t) *Abc_Ntk_t {
	return abc.Abc_NtkFromDarSeqSweep(pNtkOld, pMan)
}

/*
*Function*************************************************************

	Synopsis    [Converts the network from the AIG manager into ABC.]

	Description [This procedure should be called after seq sweeping,
	which changes the number of registers.]

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkFromAigPhase(pMan *Aig_Man_t) *Abc_Ntk_t {
	if pMan == nil || pMan.Assertions != 0 {
		return nil
	}
	pNtkNew := abc.Abc_NtkAlloc(abc.ABC_NTK_STRASH, abc.ABC_FUNC_AIG, 1)
	pNtkNew.Constraints = pMan.Assertions
	pNtkNew.Barriers = int(pMan.BarrierBuffers)
	pNtkNew.Name, pNtkNew.Spec = pMan.Name, pMan.Spec
	pis := int(pMan.InputCount() - int(pMan.RegisterCount()))
	pos := int(pMan.OutputCount() - int(pMan.RegisterCount()))
	if pis < 0 || pos < 0 {
		return nil
	}
	for i := 0; i < pis; i++ {
		abc.Abc_NtkCreatePi(pNtkNew)
	}
	for i := 0; i < pos; i++ {
		abc.Abc_NtkCreatePo(pNtkNew)
	}
	copies := make([]abc.Edge, len(pMan.Nodes()))
	valid := make([]bool, len(copies))
	copies[0], valid[0] = pNtkNew.One(), true
	inputs := pMan.Inputs()
	for i := 0; i < int(pMan.RegisterCount()); i++ {
		latchOutput := abc.Abc_NtkAddLatch(pNtkNew, nil, abc.ABC_INIT_ZERO)
		id := inputs[pis+i].NodeID()
		copies[id], valid[id] = latchOutput.Signal(), true
	}
	for i := 0; i < pis; i++ {
		id := inputs[i].NodeID()
		copies[id], valid[id] = abc.Abc_NtkCi(pNtkNew, int32(i)).Signal(), true
	}
	abc.Abc_NtkOrderCisCos(pNtkNew)
	manager := abc.Abc_NtkReadFunctionManager(pNtkNew).(*abc.Abc_Aig_t)
	var build func(aiggraph.Edge) abc.Edge
	build = func(edge aiggraph.Edge) abc.Edge {
		id := edge.NodeID()
		if !valid[id] {
			node := pMan.Node(aiggraph.Edge(id << 1))
			switch node.Kind {
			case aiggraph.And:
				copies[id] = abc.Abc_AigAnd(manager, build(node.Fanins[0]), build(node.Fanins[1]))
			case aiggraph.ExclusiveOr:
				copies[id] = abc.Abc_AigXor(manager, build(node.Fanins[0]), build(node.Fanins[1]))
			case aiggraph.Buffer:
				copies[id] = build(node.Fanins[0])
			default:
				panic("abc: unsupported AIG object while rebuilding network")
			}
			valid[id] = true
		}
		result := copies[id]
		if edge.Complemented() {
			result = result.Not()
		}
		return result
	}
	for i, output := range pMan.Outputs() {
		co := abc.Abc_NtkCo(pNtkNew, int32(i))
		abc.Abc_ObjAddFanin(co, build(pMan.Node(output).Fanins[0]))
	}
	abc.Abc_NtkAddDummyPiNames(pNtkNew)
	abc.Abc_NtkAddDummyPoNames(pNtkNew)
	abc.Abc_NtkAddDummyBoxNames(pNtkNew)
	if abc.Abc_NtkCheck(pNtkNew) == 0 {
		fmt.Println("Abc_NtkFromAigPhase(): Network check has failed.")
	}
	return pNtkNew
}

/*
*Function*************************************************************

	Synopsis    [Creates local function of the node.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_ObjHopFromGia_rec(pHopMan *Hop_Man_t, p *Gia_Man_t, id int32, vCopies *Vec_Ptr_t) *Hop_Obj_t {
	if giapkg.Gia_ObjIsTravIdCurrentId(p, id) != 0 {
		obj, _ := Vec_PtrEntry(vCopies, int32(int(id))).(*Hop_Obj_t)
		return obj
	}
	giapkg.Gia_ObjSetTravIdCurrentId(p, id)
	obj := giapkg.Gia_ManObj(p, id)
	assert(giapkg.Gia_ObjIsAnd(obj) != 0, "Gia_ObjIsAnd(pObj)")
	left := Abc_ObjHopFromGia_rec(pHopMan, p, giapkg.Gia_ObjFaninId0(obj, id), vCopies)
	right := Abc_ObjHopFromGia_rec(pHopMan, p, giapkg.Gia_ObjFaninId1(obj, id), vCopies)
	result := hoppkg.Hop_And(pHopMan, hoppkg.Hop_NotCond(left, giapkg.Gia_ObjFaninC0(obj)), hoppkg.Hop_NotCond(right, giapkg.Gia_ObjFaninC1(obj)))
	Vec_PtrWriteEntry(vCopies, int(id), result)
	return result
}
func Abc_ObjHopFromGia(pHopMan *Hop_Man_t, p *Gia_Man_t, giaID int32, vCopies *Vec_Ptr_t) *Hop_Obj_t {
	assert(giapkg.Gia_ObjIsLut(p, giaID) != 0, "Gia_ObjIsLut(p, GiaId)")
	assert(giapkg.Gia_ObjLutSize(p, giaID) > 0, "Gia_ObjLutSize(p, GiaId) > 0")
	giapkg.Gia_ManIncrementTravId(p)
	for i, fanin := range giapkg.Gia_ObjLutFanins(p, giaID) {
		giapkg.Gia_ObjSetTravIdCurrentId(p, fanin)
		Vec_PtrWriteEntry(vCopies, int(fanin), hoppkg.Hop_IthVar(pHopMan, int32(i)))
	}
	return Abc_ObjHopFromGia_rec(pHopMan, p, giaID, vCopies)
}

/*
*Function*************************************************************

	Synopsis    [Converts the network from the mapped GIA manager.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkFromMappedGia_rec(pNtkNew *Abc_Ntk_t, p *Gia_Man_t, iObj int32, fAddInv int32) *Abc_Obj_t {
	pObj := giapkg.Gia_ManObj(p, iObj)
	value := giapkg.Gia_ObjValue(pObj)
	var pObjNew *abc.Object
	if value >= 0 {
		pObjNew = abc.Abc_NtkObj(pNtkNew, value)
	} else {
		left := Abc_NtkFromMappedGia_rec(pNtkNew, p, giapkg.Gia_ObjFaninId0(pObj, iObj), 0)
		right := Abc_NtkFromMappedGia_rec(pNtkNew, p, giapkg.Gia_ObjFaninId1(pObj, iObj), 0)
		pObjNew = abc.Abc_NtkCreateNode(pNtkNew)
		abc.Abc_ObjAddFanin(pObjNew, left)
		abc.Abc_ObjAddFanin(pObjNew, right)
		sop := abc.Abc_SopCreateAnd2(nil, giapkg.Gia_ObjFaninC0(pObj), giapkg.Gia_ObjFaninC1(pObj))
		pObjNew.Data = sop
		giapkg.Gia_ObjSetValue(pObj, abc.Abc_ObjId(pObjNew))
	}
	if fAddInv != 0 {
		pObjNew = abc.Abc_NtkCreateNodeInv(pNtkNew, pObjNew)
	}
	return pObjNew
}
func Abc_NtkFromMappedGia(p *Gia_Man_t, fFindEnables int32, fUseBuffs int32) *Abc_Ntk_t {
	if p == nil {
		return nil
	}
	mapped, muxes := giapkg.Gia_ManHasMapping(p) != 0, giapkg.Gia_ManHasMuxData(p)
	assert(mapped || muxes || fFindEnables != 0, "mapped GIA, mux data or enable extraction is required")
	assert(fFindEnables == 0 || !muxes, "enable extraction requires an ordinary AIG")
	function := abc.ABC_FUNC_SOP
	if mapped {
		function = abc.ABC_FUNC_AIG
	}
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_LOGIC, function, 1)
	n.Name, n.Spec = p.Identity()
	giapkg.Gia_ManFillValue(p)
	constant := abc.Abc_NtkCreateNodeConst0(n)
	giapkg.Gia_ObjSetValue(giapkg.Gia_ManConst0(p), abc.Abc_ObjId(constant))
	for i := int32(0); i < giapkg.Gia_ManPiNum(p); i++ {
		giapkg.Gia_ObjSetValue(giapkg.Gia_ManPi(p, i), abc.Abc_ObjId(abc.Abc_NtkCreatePi(n)))
	}
	for i := int32(0); i < giapkg.Gia_ManPoNum(p); i++ {
		giapkg.Gia_ObjSetValue(giapkg.Gia_ManPo(p, i), abc.Abc_ObjId(abc.Abc_NtkCreatePo(n)))
	}
	for i := int32(0); i < giapkg.Gia_ManRegNum(p); i++ {
		latch := abc.Abc_NtkCreateLatch(n)
		li, lo := abc.Abc_NtkCreateBi(n), abc.Abc_NtkCreateBo(n)
		abc.Abc_ObjAddFanin(latch, li)
		abc.Abc_ObjAddFanin(lo, latch)
		abc.Abc_LatchSetInit0(latch)
		giapkg.Gia_ObjSetValue(giapkg.Gia_ManRi(p, i), abc.Abc_ObjId(li))
		giapkg.Gia_ObjSetValue(giapkg.Gia_ManRo(p, i), abc.Abc_ObjId(lo))
	}
	object := func(id int32) *abc.Object { return abc.Abc_NtkObj(n, giapkg.Gia_ObjValue(giapkg.Gia_ManObj(p, id))) }
	switch {
	case fFindEnables != 0:
		for i := int32(0); i < giapkg.Gia_ManCoNum(p); i++ {
			co := giapkg.Gia_ManCo(p, i)
			coID := giapkg.Gia_ObjId(p, co)
			driverID := giapkg.Gia_ObjFaninId0(co, coID)
			var driver *abc.Object
			if giapkg.Gia_ObjIsRi(p, co) != 0 && giapkg.Gia_ObjIsMuxType(giapkg.Gia_ManObj(p, driverID)) != 0 {
				ro := giapkg.Gia_ObjRiToRoId(p, coID)
				var yes, no int32
				control := giapkg.Gia_ObjRecognizeMuxLits(p, giapkg.Gia_ManObj(p, driverID), &yes, &no)
				yes ^= giapkg.Gia_ObjFaninC0(co)
				no ^= giapkg.Gia_ObjFaninC0(co)
				if yes>>1 == ro {
					yes, no = no, yes
					control ^= 1
				}
				if no>>1 == ro {
					driver = abc.Abc_NtkCreateNode(n)
					for _, lit := range []int32{control, yes, no} {
						abc.Abc_ObjAddFanin(driver, Abc_NtkFromMappedGia_rec(n, p, lit>>1, lit&1))
					}
					driver.Data = abc.Abc_SopCreateMux(nil)
				}
			}
			if driver == nil {
				driver = Abc_NtkFromMappedGia_rec(n, p, driverID, giapkg.Gia_ObjFaninC0(co))
			}
			abc.Abc_ObjAddFanin(abc.Abc_NtkCo(n, i), driver)
		}
	case muxes:
		for id := int32(1); id < giapkg.Gia_ManObjNum(p); id++ {
			a, b, c, kind := p.Gate(id)
			if kind != 2 && kind != 3 && kind != 4 {
				continue
			}
			node := abc.Abc_NtkCreateNode(n)
			literals := []int32{a, b}
			sop := abc.Abc_SopCreateAnd(nil, 2, nil)
			if kind == 3 {
				sop = abc.Abc_SopCreateXor(nil, 2)
			}
			if kind == 4 {
				sop = abc.Abc_SopCreateMux(nil)
				literals = []int32{a, b, c}
			}
			for pin, lit := range literals {
				abc.Abc_ObjAddFanin(node, object(lit>>1))
				if lit&1 != 0 {
					sop = abc.Abc_SopComplementVar(sop, int32(pin))
				}
			}
			node.Data = sop
			giapkg.Gia_ObjSetValue(giapkg.Gia_ManObj(p, id), abc.Abc_ObjId(node))
		}
	default:
		copies := vec.Vec_PtrStart(int(giapkg.Gia_ManObjNum(p)))
		defer Vec_PtrFree(copies)
		for id := int32(1); id < giapkg.Gia_ManObjNum(p); id++ {
			if giapkg.Gia_ObjIsLut(p, id) == 0 {
				continue
			}
			if giapkg.Gia_ObjLutSize(p, id) == 0 {
				giapkg.Gia_ObjSetValue(giapkg.Gia_ManObj(p, id), abc.Abc_ObjId(constant))
				continue
			}
			node := abc.Abc_NtkCreateNode(n)
			for _, leaf := range giapkg.Gia_ObjLutFanins(p, id) {
				abc.Abc_ObjAddFanin(node, object(leaf))
			}
			node.Data = Abc_ObjHopFromGia(n.FunctionManager.(*Hop_Man_t), p, id, copies)
			giapkg.Gia_ObjSetValue(giapkg.Gia_ManObj(p, id), abc.Abc_ObjId(node))
		}
	}
	if fFindEnables == 0 {
		for i := int32(0); i < giapkg.Gia_ManCoNum(p); i++ {
			co := giapkg.Gia_ManCo(p, i)
			driver := object(giapkg.Gia_ObjFaninId0(co, giapkg.Gia_ObjId(p, co))).Signal()
			abc.Abc_ObjAddFanin(abc.Abc_NtkCo(n, i), abc.Abc_ObjNotCond(driver, giapkg.Gia_ObjFaninC0(co)))
		}
	}
	abc.Abc_NtkAddDummyPiNames(n)
	abc.Abc_NtkAddDummyPoNames(n)
	abc.Abc_NtkAddDummyBoxNames(n)
	abc.Abc_NtkLogicMakeSimpleCos(n, boolInt(fUseBuffs == 0))
	if constant.Kind != abc.ObjectNone && len(constant.Fanouts()) == 0 {
		abc.Abc_NtkDeleteObj(constant)
	}
	if abc.Abc_NtkCheck(n) == 0 {
		fmt.Println("Abc_NtkFromMappedGia(): Network check has failed.")
	}
	return n
}

/*
*Function*************************************************************

	Synopsis    [Converts the network from the mapped GIA manager.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkFromCellWrite(vCopyLits *Vec_Int_t, i int32, c int32, Id int32) {
	Vec_IntWriteEntry(vCopyLits, Abc_Var2Lit(i, c), Id)
}
func Abc_NtkFromCellRead(p *Abc_Ntk_t, vCopyLits *Vec_Int_t, i int32, c int32) *Abc_Obj_t {
	var pObjNew *Abc_Obj_t
	var iObjNew int32 = Vec_IntEntry(vCopyLits, Abc_Var2Lit(i, c))
	if iObjNew >= 0 {
		return Abc_NtkObj(p, iObjNew)
	}
	// opposite phase should be already constructed
	assert(false, "0")
	if i == 0 {
		if c != 0 {
			pObjNew = abc.Abc_NtkCreateNodeConst1(p)
		} else {
			pObjNew = abc.Abc_NtkCreateNodeConst0(p)
		}
	} else {
		iObjNew = Vec_IntEntry(vCopyLits, Abc_Var2Lit(i, boolInt((c == 0))))
		assert(iObjNew >= 0, "iObjNew >= 0")
		pObjNew = abc.Abc_NtkCreateNodeInv(p, Abc_NtkObj(p, iObjNew))
	}
	Abc_NtkFromCellWrite(vCopyLits, i, c, abc.Abc_ObjId(pObjNew))
	return pObjNew
}
func Abc_NtkFromCellMappedGia(p *Gia_Man_t, fUseBuffs int32) *Abc_Ntk_t {
	return abcNtkFromCellMappedGia(p, fUseBuffs)
}

/*
*Function*************************************************************

	Synopsis    [Converts the network from the AIG manager into ABC.]

	Description [This procedure should be called after seq sweeping,
	which changes the number of registers.]

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkAfterTrim(pMan *Aig_Man_t, pNtkOld *Abc_Ntk_t) *Abc_Ntk_t {
	return abcNtkAfterTrim(pMan, pNtkOld)
}

/*
*Function*************************************************************

	Synopsis    [Converts the network from the AIG manager into ABC.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkFromDarChoices(pNtkOld *Abc_Ntk_t, pMan *Aig_Man_t) *Abc_Ntk_t {
	return abcNtkFromDarChoices(pNtkOld, pMan)
}

/*
*Function*************************************************************

	Synopsis    [Converts the network from the AIG manager into ABC.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkFromDarSeq(pNtkOld *Abc_Ntk_t, pMan *Aig_Man_t) *Abc_Ntk_t {
	return abcNtkFromDarSeq(pNtkOld, pMan)
}

/*
*Function*************************************************************

	Synopsis    [Collects CI of the network.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkCollectCiNames(pNtk *Abc_Ntk_t) *Vec_Ptr_t {
	var vNames *Vec_Ptr_t
	vNames = Vec_PtrAlloc(100)
	for i := int32(0); i < Abc_NtkCiNum(pNtk); i++ {
		Vec_PtrPush(vNames, Abc_ObjName(abc.Abc_NtkCi(pNtk, i)))
	}
	return vNames
}

/*
*Function*************************************************************

	Synopsis    [Collects CO of the network.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkCollectCoNames(pNtk *Abc_Ntk_t) *Vec_Ptr_t {
	var vNames *Vec_Ptr_t
	vNames = Vec_PtrAlloc(100)
	for i := int32(0); i < Abc_NtkCoNum(pNtk); i++ {
		Vec_PtrPush(vNames, Abc_ObjName(abc.Abc_NtkCo(pNtk, i)))
	}
	return vNames
}

/*
*Function*************************************************************

	Synopsis    [Collect latch values.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkGetLatchValues(pNtk *Abc_Ntk_t) *Vec_Int_t {
	var vInits *Vec_Int_t
	vInits = Vec_IntAlloc(Abc_NtkLatchNum(pNtk))
	for _, pLatch := range pNtk.ObjectsOfKind(abc.ObjectLatch) {
		if abc.Abc_LatchIsInit0(pLatch) != 0 {
			Vec_IntPush(vInits, 0)
		} else if abc.Abc_LatchIsInit1(pLatch) != 0 {
			Vec_IntPush(vInits, 1)
		} else if abc.Abc_LatchIsInitDc(pLatch) != 0 {
			Vec_IntPush(vInits, 2)
		} else {
			assert(false, "0")
		}
	}
	return vInits
}

/*
*Function*************************************************************

	Synopsis    [Gives the current ABC network to AIG manager for processing.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkDar(pNtk *Abc_Ntk_t) *Abc_Ntk_t {
	var pNtkAig *Abc_Ntk_t = nil
	var pMan *Aig_Man_t
	// Fra_ManPartitionTest is declared in the proof package.

	assert((Abc_NtkIsStrash(pNtk) != 0), "Abc_NtkIsStrash(pNtk)")
	// convert to the AIG manager
	pMan = Abc_NtkToDar(pNtk, 0, 0)
	if pMan == nil {
		return nil
	}

	// perform computation
	//    Fra_ManPartitionTest( pMan, 4 );
	pNtkAig = Abc_NtkFromDar(pNtk, pMan)
	Aig_ManStop(pMan)

	// make sure everything is okay
	if pNtkAig != nil && abc.Abc_NtkCheck(pNtkAig) == 0 {
		Abc_Print(1, "Abc_NtkDar: The network check has failed.\n")
		Abc_NtkDelete(pNtkAig)
		return nil
	}
	return pNtkAig
}

/*
*Function*************************************************************

	Synopsis    [Gives the current ABC network to AIG manager for processing.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkDarFraig(pNtk *Abc_Ntk_t, nConfLimit int32, fDoSparse int32, fProve int32, fTransfer int32, fSpeculate int32, fChoicing int32, fVerbose int32) *Abc_Ntk_t {
	var Pars proof_fra.Fra_Par_t
	var pPars *proof_fra.Fra_Par_t = &Pars
	var pNtkAig *Abc_Ntk_t
	var pMan, pTemp *Aig_Man_t
	pMan = Abc_NtkToDar(pNtk, 0, 0)
	if pMan == nil {
		return nil
	}
	proof_fra.Fra_ParamsDefault(pPars)
	pPars.NBTLimitNode = nConfLimit
	pPars.FChoicing = fChoicing
	pPars.FDoSparse = fDoSparse
	pPars.FSpeculate = fSpeculate
	pPars.FProve = fProve
	pPars.FVerbose = fVerbose
	pTemp = pMan
	pMan = proof_fra.Fra_FraigPerform(pTemp, pPars)
	if fChoicing != 0 {
		pNtkAig = Abc_NtkFromDarChoices(pNtk, pMan)
	} else {
		pNtkAig = Abc_NtkFromDar(pNtk, pMan)
	}
	Aig_ManStop(pTemp)
	Aig_ManStop(pMan)
	return pNtkAig
}

/*
*Function*************************************************************

	Synopsis    [Gives the current ABC network to AIG manager for processing.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkDarFraigPart(pNtk *Abc_Ntk_t, nPartSize int32, nConfLimit int32, nLevelMax int32, fVerbose int32) *Abc_Ntk_t {
	var pNtkAig *Abc_Ntk_t
	var pMan, pTemp *Aig_Man_t
	pMan = Abc_NtkToDar(pNtk, 0, 0)
	if pMan == nil {
		return nil
	}
	pTemp = pMan
	pMan = proof_fra.Aig_ManFraigPartitioned(pTemp, nPartSize, nConfLimit, nLevelMax, fVerbose)
	Aig_ManStop(pTemp)
	pNtkAig = Abc_NtkFromDar(pNtk, pMan)
	Aig_ManStop(pMan)
	return pNtkAig
}

/*
*Function*************************************************************

	Synopsis    [Gives the current ABC network to AIG manager for processing.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkCSweep(pNtk *Abc_Ntk_t, nCutsMax int32, nLeafMax int32, fVerbose int32) *Abc_Ntk_t {
	//    extern Aig_Man_t * Csw_Sweep( Aig_Man_t * pAig, int nCutsMax, int nLeafMax, int fVerbose );
	var pNtkAig *Abc_Ntk_t
	var pMan, pTemp *Aig_Man_t
	pMan = Abc_NtkToDar(pNtk, 0, 0)
	if pMan == nil {
		return nil
	}
	pTemp = pMan
	pMan = opt_csw.Csw_Sweep(pTemp, nCutsMax, nLeafMax, fVerbose)
	pNtkAig = Abc_NtkFromDar(pNtk, pMan)
	Aig_ManStop(pTemp)
	Aig_ManStop(pMan)
	return pNtkAig
}

/*
*Function*************************************************************

	Synopsis    [Gives the current ABC network to AIG manager for processing.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkDRewrite(pNtk *Abc_Ntk_t, pPars *Dar_RwrPar_t) *Abc_Ntk_t {
	var pMan, pTemp *Aig_Man_t
	var pNtkAig *Abc_Ntk_t
	assert((Abc_NtkIsStrash(pNtk) != 0), "Abc_NtkIsStrash(pNtk)")
	pMan = Abc_NtkToDar(pNtk, 0, 0)
	if pMan == nil {
		return nil
	}
	//    Aig_ManPrintStats( pMan );
	/*
	   //    Aig_ManSupports( pMan );
	       {
	           Vec_Vec_t * vParts;
	           vParts = Aig_ManPartitionSmart( pMan, 50, 1, NULL );
	           Vec_VecFree( vParts );
	       }
	*/
	if err := opt_dar.Dar_ManRewrite(pMan, pPars); err != nil {
		Aig_ManStop(pMan)
		Abc_Print(1, "%s\n", err)
		return nil
	}
	//    pTemp = pMan
	// Balancing was commented out in the C source.
	//    Aig_ManStop( pTemp );

	pTemp = pMan
	pMan = aiggraph.Aig_ManDupDfs(pTemp)
	Aig_ManStop(pTemp)
	//ABC_PRT( "time", Abc_Clock() - clk );

	//    Aig_ManPrintStats( pMan );
	pNtkAig = Abc_NtkFromDar(pNtk, pMan)
	Aig_ManStop(pMan)
	return pNtkAig
}

/*
*Function*************************************************************

	Synopsis    [Gives the current ABC network to AIG manager for processing.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkDRefactor(pNtk *Abc_Ntk_t, pPars *Dar_RefPar_t) *Abc_Ntk_t {
	var pMan, pTemp *Aig_Man_t
	var pNtkAig *Abc_Ntk_t
	assert((Abc_NtkIsStrash(pNtk) != 0), "Abc_NtkIsStrash(pNtk)")
	pMan = Abc_NtkToDar(pNtk, 0, 0)
	if pMan == nil {
		return nil
	}
	//    Aig_ManPrintStats( pMan );

	if err := opt_dar.Dar_ManRefactor(pMan, pPars); err != nil {
		Aig_ManStop(pMan)
		Abc_Print(1, "%s\n", err)
		return nil
	}
	//    pTemp = pMan
	// Balancing was commented out in the C source.
	//    Aig_ManStop( pTemp );

	pTemp = pMan
	pMan = aiggraph.Aig_ManDupDfs(pTemp)
	Aig_ManStop(pTemp)
	//ABC_PRT( "time", Abc_Clock() - clk );

	//    Aig_ManPrintStats( pMan );
	pNtkAig = Abc_NtkFromDar(pNtk, pMan)
	Aig_ManStop(pMan)
	return pNtkAig
}

/*
*Function*************************************************************

	Synopsis    [Gives the current ABC network to AIG manager for processing.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkDC2(pNtk *Abc_Ntk_t, fBalance int32, fUpdateLevel int32, fFanout int32, fPower int32, fVerbose int32) *Abc_Ntk_t {
	var pMan, pTemp *Aig_Man_t
	var pNtkAig *Abc_Ntk_t
	assert((Abc_NtkIsStrash(pNtk) != 0), "Abc_NtkIsStrash(pNtk)")
	pMan = Abc_NtkToDar(pNtk, 0, 0)
	if pMan == nil {
		return nil
	}
	//    Aig_ManPrintStats( pMan );

	pTemp = pMan
	var err error
	pMan, err = opt_dar.Dar_ManCompress2(pTemp, (fBalance != 0), (fUpdateLevel != 0), (fFanout != 0), (fPower != 0), (fVerbose != 0))
	Aig_ManStop(pTemp)
	if err != nil {
		Abc_Print(1, "%s\n", err)
		return nil
	}
	//ABC_PRT( "time", Abc_Clock() - clk );

	//    Aig_ManPrintStats( pMan );
	pNtkAig = Abc_NtkFromDar(pNtk, pMan)
	Aig_ManStop(pMan)
	return pNtkAig
}

/*
*Function*************************************************************

	Synopsis    [Gives the current ABC network to AIG manager for processing.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkDChoice(pNtk *Abc_Ntk_t, fBalance int32, fUpdateLevel int32, fConstruct int32, nConfMax int32, nLevelMax int32, fVerbose int32) *Abc_Ntk_t {
	var pMan, pTemp *Aig_Man_t
	var pNtkAig *Abc_Ntk_t
	assert((Abc_NtkIsStrash(pNtk) != 0), "Abc_NtkIsStrash(pNtk)")
	pMan = Abc_NtkToDar(pNtk, 0, 0)
	if pMan == nil {
		return nil
	}
	pTemp = pMan
	var err error
	pMan, err = opt_dar.Dar_ManChoice(pTemp, (fBalance != 0), (fUpdateLevel != 0), (fConstruct != 0), nConfMax, nLevelMax, (fVerbose != 0))
	Aig_ManStop(pTemp)
	if err != nil {
		Abc_Print(1, "%s\n", err)
		return nil
	}
	pNtkAig = Abc_NtkFromDarChoices(pNtk, pMan)
	Aig_ManStop(pMan)
	return pNtkAig
}

/*
*Function*************************************************************

	Synopsis    [Gives the current ABC network to AIG manager for processing.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkDch(pNtk *Abc_Ntk_t, pPars *Dch_Pars_t) *Abc_Ntk_t {
	assert(Abc_NtkIsStrash(pNtk) != 0, "Abc_NtkIsStrash(pNtk)")
	g := Abc_NtkToDar(pNtk, 0, 0)
	if g == nil {
		return nil
	}
	params := pPars.native()
	start := Abc_Clock()
	var combined *Gia_Man_t
	if pPars.fSynthesis != 0 {
		options := opt_dar.ChoiceOptions{Balance: true, UpdateLevel: true, Power: params.Power, Light: params.LightSynth}
		if params.Verbose {
			options.Progress = os.Stdout
		}
		versions, err := opt_dar.Dar_NewChoiceSynthesis(g, &options)
		Aig_ManStop(g)
		if err != nil {
			Abc_Print(1, "%s\n", err)
			return nil
		}
		combined, err = giapkg.Gia_ManChoiceMiterFromAigs(versions)
		for _, version := range versions {
			Aig_ManStop(version)
		}
		if err != nil {
			Abc_Print(1, "%s\n", err)
			return nil
		}
	} else {
		combined = Gia_ManFromAig(g)
		Aig_ManStop(g)
	}
	defer Gia_ManStop(combined)
	pPars.timeSynth = Abc_Clock() - start
	params.SynthDuration = time.Duration(pPars.timeSynth)
	var result *Aig_Man_t
	if params.UseGia {
		result = proof_cec.Cec_ComputeChoices(combined, params)
	} else {
		input := giapkg.Gia_ManToAigSkip(combined, 3)
		var err error
		result, err = proof_dch.Dch_ComputeChoices(input, params)
		Aig_ManStop(input)
		if err != nil {
			Abc_Print(1, "%s\n", err)
			return nil
		}
	}
	if result == nil {
		return nil
	}
	defer Aig_ManStop(result)
	return Abc_NtkFromDarChoices(pNtk, result)
}

/*
*Function*************************************************************

	Synopsis    [Gives the current ABC network to AIG manager for processing.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkDrwsat(pNtk *Abc_Ntk_t, fBalance int32, fVerbose int32) *Abc_Ntk_t {
	var pMan, pTemp *Aig_Man_t
	var pNtkAig *Abc_Ntk_t
	assert((Abc_NtkIsStrash(pNtk) != 0), "Abc_NtkIsStrash(pNtk)")
	pMan = Abc_NtkToDar(pNtk, 0, 0)
	if pMan == nil {
		return nil
	}
	//    Aig_ManPrintStats( pMan );

	pTemp = pMan
	var err error
	pMan, err = opt_dar.Dar_ManRwsat(pTemp, (fBalance != 0), (fVerbose != 0))
	Aig_ManStop(pTemp)
	if err != nil {
		Abc_Print(1, "%s\n", err)
		return nil
	}
	//ABC_PRT( "time", Abc_Clock() - clk );

	//    Aig_ManPrintStats( pMan );
	pNtkAig = Abc_NtkFromDar(pNtk, pMan)
	Aig_ManStop(pMan)
	return pNtkAig
}

/*
*Function*************************************************************

	Synopsis    [Gives the current ABC network to AIG manager for processing.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkConstructFromCnf(pNtk *Abc_Ntk_t, p *Cnf_Man_t, vMapped *Vec_Ptr_t) *Abc_Ntk_t {
	n := abc.Abc_NtkStartFrom(pNtk, abc.ABC_NTK_LOGIC, abc.ABC_FUNC_SOP)
	g := p.MappingGraph()
	constant := abc.Abc_NtkCreateNodeConst1(n)
	copies := map[int]*Abc_Obj_t{g.Const1().NodeID(): constant}
	for i, edge := range g.Inputs() {
		copies[edge.NodeID()] = Abc_ObjReadCopyEdge(abc.Abc_NtkCi(pNtk, int32(i))).Object
	}
	cover := Vec_IntAlloc(1 << 16)
	for _, item := range giapkg.Vec_PtrArray(vMapped) {
		edge := item.(aiggraph.Edge)
		cut := sat_cnf.Cnf_ObjBestCut(p, &edge)
		node := abc.Abc_NtkCreateNode(n)
		for _, leaf := range sat_cnf.Cnf_CutLeaves(cut) {
			Abc_ObjAddFanin(node, copies[int(leaf)])
		}
		fanins := sat_cnf.Cnf_CutLeaveNum(cut)
		isop := cut.PositiveCover()
		if fanins < 5 {
			sop, cubes := p.TruthCover(uint16(sat_cnf.Cnf_CutTruth(cut)[0]))
			sat_cnf.Cnf_SopConvertToVector(sop, cubes, cover)
			isop = cover
		}
		node.Data = abc.Abc_SopCreateFromIsop(nil, fanins, isop)
		copies[edge.NodeID()] = node
	}
	Vec_IntFree(cover)
	for i, output := range g.Outputs() {
		fanin := g.Node(output).Fanins[0]
		driver := abc.Abc_ObjNotCond(copies[fanin.NodeID()], boolInt(fanin.Complemented()))
		Abc_ObjAddFanin(Abc_ObjReadCopyEdge(abc.Abc_NtkCo(pNtk, int32(i))).Object, driver)
	}
	if abc.Abc_ObjFanoutNum(constant) == 0 {
		abc.Abc_NtkDeleteObj(constant)
	}
	abc.Abc_NtkLogicMakeSimpleCos(n, 1)
	if abc.Abc_NtkCheck(n) == 0 {
		Abc_Print(1, "Abc_NtkConstructFromCnf(): Network check has failed.\n")
	}
	return n
}

/*
*Function*************************************************************

	Synopsis    [Gives the current ABC network to AIG manager for processing.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkDarToCnf(pNtk *Abc_Ntk_t, pFileName string, fFastAlgo int32, fChangePol int32, fVerbose int32) *Abc_Ntk_t {
	//    Vec_Ptr_t * vMapped = NULL;
	var pMan *Aig_Man_t
	//    Cnf_Man_t * pManCnf = NULL;
	var pCnf *sat_cnf.Cnf_Dat_t
	var pNtkNew *Abc_Ntk_t = nil
	var clk abctime = Abc_Clock()
	assert((Abc_NtkIsStrash(pNtk) != 0), "Abc_NtkIsStrash(pNtk)")

	// convert to the AIG manager
	pMan = Abc_NtkToDar(pNtk, 0, 0)
	if pMan == nil {
		return nil
	}
	if aiggraph.Aig_ManCheck(pMan) == 0 {
		Abc_Print(1, "Abc_NtkDarToCnf: AIG check has failed.\n")
		Aig_ManStop(pMan)
		return nil
	}
	// perform balance
	if fVerbose != 0 {
		aiggraph.Aig_ManPrintStats(pMan)
	}

	// derive CNF
	if fFastAlgo != 0 {
		pCnf = sat_cnf.Cnf_DeriveFast(pMan, 0)
	} else {
		pCnf = sat_cnf.Cnf_Derive(pMan, 0)
	}

	// adjust polarity
	if fChangePol != 0 {
		sat_cnf.Cnf_DataTranformPolarity(pCnf, 0)
	}

	// print stats
	//    if ( fVerbose )
	{
		Abc_Print(1, "CNF stats: Vars = %6d. Clauses = %7d. Literals = %8d.   ", sat_cnf.Cnf_DataVarNum(pCnf), sat_cnf.Cnf_DataClauseNum(pCnf), sat_cnf.Cnf_DataLiteralNum(pCnf))
		Abc_PrintTime(1, "Time", Abc_Clock()-clk)
	}

	/*
	   // write the network for verification
	   pManCnf = Cnf_ManRead();
	   vMapped = Cnf_ManScanMapping( pManCnf, 1, 0 );
	   pNtkNew = Abc_NtkConstructFromCnf( pNtk, pManCnf, vMapped );
	   Vec_PtrFree( vMapped );
	*/
	// write CNF into a file
	sat_cnf.Cnf_DataWriteIntoFile(pCnf, pFileName, 0, nil, nil)
	sat_cnf.Cnf_DataFree(pCnf)
	Cnf_ManFree()
	Aig_ManStop(pMan)
	return pNtkNew
}

/*
*Function*************************************************************

	Synopsis    [Solves combinational miter using a SAT solver.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkDSat(pNtk *Abc_Ntk_t, nConfLimit ABC_INT64_T, nInsLimit ABC_INT64_T, nLearnedStart int32, nLearnedDelta int32, nLearnedPerce int32, fAlignPol int32, fAndOuts int32, fNewSolver int32, fVerbose int32) int32 {
	var pMan *Aig_Man_t
	var RetValue int32 //, clk = Abc_Clock();
	assert((Abc_NtkIsStrash(pNtk) != 0), "Abc_NtkIsStrash(pNtk)")
	assert(Abc_NtkLatchNum(pNtk) == 0, "Abc_NtkLatchNum(pNtk) == 0")
	//    assert( Abc_NtkPoNum(pNtk) == 1 );
	pMan = Abc_NtkToDar(pNtk, 0, 0)
	RetValue = proof_fra.Fra_FraigSat(pMan, nConfLimit, nInsLimit, nLearnedStart, nLearnedDelta, nLearnedPerce, fAlignPol, fAndOuts, fNewSolver, fVerbose)
	Abc_NtkSetModel(pNtk, abcAigModel(pMan))
	pMan.Data = nil
	Aig_ManStop(pMan)
	return RetValue
}

/*
*Function*************************************************************

	Synopsis    [Solves combinational miter using a SAT solver.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkPartitionedSat(pNtk *Abc_Ntk_t, nAlgo int32, nPartSize int32, nConfPart int32, nConfTotal int32, fAlignPol int32, fSynthesize int32, fVerbose int32) int32 {
	var pMan *Aig_Man_t
	var RetValue int32 //, clk = Abc_Clock();
	assert((Abc_NtkIsStrash(pNtk) != 0), "Abc_NtkIsStrash(pNtk)")
	assert(Abc_NtkLatchNum(pNtk) == 0, "Abc_NtkLatchNum(pNtk) == 0")
	pMan = Abc_NtkToDar(pNtk, 0, 0)
	RetValue = aigpartsat.Aig_ManPartitionedSat(pMan, nAlgo, nPartSize, nConfPart, nConfTotal, fAlignPol, fSynthesize, fVerbose)
	Abc_NtkSetModel(pNtk, abcAigModel(pMan))
	pMan.Data = nil
	Aig_ManStop(pMan)
	return RetValue
}

/*
*Function*************************************************************

	Synopsis    [Gives the current ABC network to AIG manager for processing.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkDarCec(pNtk1 *Abc_Ntk_t, pNtk2 *Abc_Ntk_t, nConfLimit int32, fPartition int32, fVerbose int32) int32 {
	var pMan, pMan1, pMan2 *Aig_Man_t
	var pMiter *Abc_Ntk_t
	var RetValue int32
	var clkTotal abctime = Abc_Clock()
	/*
	   {
	   Aig_Man_t * pAig0 = Abc_NtkToDar( pNtk1, 0, 0 );
	   Aig_Man_t * pAig1 = Abc_NtkToDar( pNtk2, 0, 0 );
	   Cec_ManVerifyTwoAigs( pAig0, pAig1, 1 );
	   Aig_ManStop( pAig0 );
	   Aig_ManStop( pAig1 );
	   return 1;
	   }
	*/
	// cannot partition if it is already a miter
	if pNtk2 == nil && fPartition == 1 {
		Abc_Print(1, "Abc_NtkDarCec(): Switching to non-partitioned CEC for the miter.\n")
		fPartition = 0
	}

	// if partitioning is selected, call partitioned CEC
	if fPartition != 0 {
		pMan1 = Abc_NtkToDar(pNtk1, 0, 0)
		pMan2 = Abc_NtkToDar(pNtk2, 0, 0)
		RetValue = proof_fra.Fra_FraigCecPartitioned(pMan1, pMan2, nConfLimit, 100, 1, fVerbose)
		Aig_ManStop(pMan1)
		Aig_ManStop(pMan2)
		goto finish
	}

	if pNtk2 != nil {
		// get the miter of the two networks
		pMiter = Abc_NtkMiter(pNtk1, pNtk2, 0, 0, 0, 0)
		if pMiter == nil {
			Abc_Print(1, "Miter computation has failed.\n")
			return 0
		}

	} else {
		pMiter = Abc_NtkDup(pNtk1)
	}
	RetValue = Abc_NtkMiterIsConstant(pMiter)
	if RetValue == 0 {
		//        extern void Abc_NtkVerifyReportErrorSeq( Abc_Ntk_t * pNtk1, Abc_Ntk_t * pNtk2, int * pModel, int nFrames );
		Abc_Print(1, "Networks are NOT EQUIVALENT after structural hashing.\n")
		// report the error
		if pNtk2 == nil {
			Abc_NtkSetModel(pNtk1, Abc_NtkVerifyGetCleanModel(pNtk1, 1))
		}
		//        pMiter->pModel = Abc_NtkVerifyGetCleanModel( pMiter, nFrames );
		//        Abc_NtkVerifyReportErrorSeq( pNtk1, pNtk2, pMiter->pModel, nFrames );
		//        ABC_FREE( pMiter->pModel );
		Abc_NtkDelete(pMiter)
		return 0
	}
	if RetValue == 1 {
		Abc_NtkDelete(pMiter)
		Abc_Print(1, "Networks are equivalent after structural hashing.\n")
		return 1
	}

	// derive the AIG manager
	pMan = Abc_NtkToDar(pMiter, 0, 0)
	Abc_NtkDelete(pMiter)
	if pMan == nil {
		Abc_Print(1, "Converting miter into AIG has failed.\n")
		return -1
	}
	// perform verification
	RetValue = proof_fra.Fra_FraigCec(&pMan, 100000, fVerbose)
	// transfer model if given
	if pNtk2 == nil {
		Abc_NtkSetModel(pNtk1, abcAigModel(pMan))
		pMan.Data = nil
	}
	Aig_ManStop(pMan)

finish:
	// report the miter
	if RetValue == 1 {
		Abc_Print(1, "Networks are equivalent.  ")
		ABC_PRT("Time", Abc_Clock()-clkTotal)

	} else if RetValue == 0 {
		Abc_Print(1, "Networks are NOT EQUIVALENT.  ")
		ABC_PRT("Time", Abc_Clock()-clkTotal)

	} else {
		Abc_Print(1, "Networks are UNDECIDED.  ")
		ABC_PRT("Time", Abc_Clock()-clkTotal)
	}
	return RetValue
}

/*
*Function*************************************************************

	Synopsis    [Gives the current ABC network to AIG manager for processing.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkDarSeqSweep(pNtk *Abc_Ntk_t, pPars *Fra_Ssw_t) *Abc_Ntk_t {
	var Params Fraig_Params_t
	var pNtkAig *Abc_Ntk_t = nil
	var pNtkFraig *Abc_Ntk_t
	var pMan, pTemp *Aig_Man_t
	var clk abctime = Abc_Clock()

	// preprocess the miter by fraiging it
	// (note that for each functional class, fraiging leaves one representative;
	// so fraiging does not reduce the number of functions represented by nodes
	proof_fraig.Fraig_ParamsSetDefault(&Params)
	Params.NBTLimit = 100000
	if (pPars.FFraiging != 0) && pPars.NPartSize == 0 {
		pNtkFraig = Abc_NtkFraig(pNtk, &Params, 0, 0)
		if pPars.FVerbose != 0 {
			ABC_PRT("Initial fraiging time", Abc_Clock()-clk)
		}

	} else {
		pNtkFraig = Abc_NtkDup(pNtk)
	}

	pMan = Abc_NtkToDar(pNtkFraig, 0, 1)
	Abc_NtkDelete(pNtkFraig)
	if pMan == nil {
		return nil
	}

	//    pPars.TimeLimit = 5.0;
	pTemp = pMan
	pMan = proof_fra.Fra_FraigInduction(pTemp, pPars)
	Aig_ManStop(pTemp)
	if pMan != nil {
		if aiggraph.Aig_ManRegNum(pMan) < Abc_NtkLatchNum(pNtk) {
			pNtkAig = Abc_NtkFromDarSeqSweep(pNtk, pMan)
		} else {
			pNtkAig = Abc_NtkFromDar(pNtk, pMan)
			for _, pObj := range pNtkAig.ObjectsOfKind(abc.ObjectLatch) {
				abc.Abc_LatchSetInit0(pObj)
			}
		}
		Aig_ManStop(pMan)
	}
	return pNtkAig
}

/*
*Function*************************************************************

	Synopsis    [Print Latch Equivalence Classes.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkPrintLatchEquivClasses(pNtk *Abc_Ntk_t, pAig *Aig_Man_t) {
	if pNtk == nil || pAig == nil {
		return
	}
	latches := abc.Abc_NtkLatchNum(pNtk)
	objs := pNtk.ObjectsOfKind(abc.ObjectLatch)
	names := make([]string, latches)
	irrelevant := make([]bool, latches)
	for i := int32(0); i < latches && int(i) < len(objs); i++ {
		names[i] = abc.Abc_ObjName(abc.Abc_ObjFanout0(objs[i]))
	}
	aiggraph.Aig_ManSetCioIds(pAig)
	header := false
	pis := aig_saig.Saig_ManPiNum(pAig)
	for i := int32(0); i < aiggraph.Aig_ManRegNum(pAig); i++ {
		flop := aiggraph.Aig_ManLo(pAig, i)
		repr := aiggraph.Aig_ObjRepr(pAig, flop)
		if repr == flop {
			continue
		}
		if !header {
			fmt.Println("Here are the flop equivalences:")
			header = true
		}
		idx := aiggraph.Aig_ObjCioId(pAig, repr) - pis
		if idx < 0 || int(idx) >= len(names) {
			continue
		}
		fmt.Printf("Original flop %s is proved equivalent to flop %s.\n", names[i], names[idx])
	}
	header = false
	for i, unused := range irrelevant {
		if unused {
			if !header {
				fmt.Println("The following flops have been deemed irrelevant:")
				header = true
			}
			fmt.Printf("%s ", names[i])
		}
	}
	if header {
		fmt.Println()
	}
}

/*
*Function*************************************************************

	Synopsis    [Gives the current ABC network to AIG manager for processing.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkDarSeqSweep2(pNtk *Abc_Ntk_t, pPars *Ssw_Pars_t) *Abc_Ntk_t {
	var pNtkAig *Abc_Ntk_t
	var pMan, pTemp *Aig_Man_t

	pMan = Abc_NtkToDar(pNtk, 0, 1)
	if pMan == nil {
		return nil
	}

	pTemp = pMan
	pMan = proof_ssw.Ssw_SignalCorrespondence(pTemp, pPars)

	if pPars.FFlopVerbose != 0 {
		Abc_NtkPrintLatchEquivClasses(pNtk, pTemp)
	}

	Aig_ManStop(pTemp)
	if pMan == nil {
		return nil
	}

	if aiggraph.Aig_ManRegNum(pMan) < Abc_NtkLatchNum(pNtk) {
		pNtkAig = Abc_NtkFromDarSeqSweep(pNtk, pMan)
	} else {
		pNtkAig = Abc_NtkFromDar(pNtk, pMan)
		for _, pObj := range pNtkAig.ObjectsOfKind(abc.ObjectLatch) {
			abc.Abc_LatchSetInit0(pObj)
		}
	}
	Aig_ManStop(pMan)
	return pNtkAig
}

/*
*Function*************************************************************

	Synopsis    [Computes latch correspondence.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkDarLcorr(pNtk *Abc_Ntk_t, nFramesP int32, nConfMax int32, fVerbose int32) *Abc_Ntk_t {
	var pMan, pTemp *Aig_Man_t
	var pNtkAig *Abc_Ntk_t = nil
	pMan = Abc_NtkToDar(pNtk, 0, 1)
	if pMan == nil {
		return nil
	}
	pTemp = pMan
	pMan = proof_fra.Fra_FraigLatchCorrespondence(pTemp, nFramesP, nConfMax, 0, fVerbose, nil, 0.0)
	Aig_ManStop(pTemp)
	if pMan != nil {
		if aiggraph.Aig_ManRegNum(pMan) < Abc_NtkLatchNum(pNtk) {
			pNtkAig = Abc_NtkFromDarSeqSweep(pNtk, pMan)
		} else {
			pNtkAig = Abc_NtkFromDar(pNtk, pMan)
			for _, pObj := range pNtkAig.ObjectsOfKind(abc.ObjectLatch) {
				abc.Abc_LatchSetInit0(pObj)
			}
		}
		Aig_ManStop(pMan)
	}
	return pNtkAig
}

/*
*Function*************************************************************

	Synopsis    [Computes latch correspondence.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkDarLcorrNew(pNtk *Abc_Ntk_t, nVarsMax int32, nConfMax int32, fVerbose int32) *Abc_Ntk_t {
	var Pars Ssw_Pars_t
	var pPars *Ssw_Pars_t = &Pars
	var pMan, pTemp *Aig_Man_t
	var pNtkAig *Abc_Ntk_t = nil
	pMan = Abc_NtkToDar(pNtk, 0, 1)
	if pMan == nil {
		return nil
	}
	proof_ssw.Ssw_ManSetDefaultParams(pPars)
	pPars.FLatchCorrOpt = 1
	pPars.NBTLimit = nConfMax
	pPars.NSatVarMax = nVarsMax
	pPars.FVerbose = fVerbose
	pTemp = pMan
	pMan = proof_ssw.Ssw_SignalCorrespondence(pTemp, pPars)
	Aig_ManStop(pTemp)
	if pMan != nil {
		if aiggraph.Aig_ManRegNum(pMan) < Abc_NtkLatchNum(pNtk) {
			pNtkAig = Abc_NtkFromDarSeqSweep(pNtk, pMan)
		} else {
			pNtkAig = Abc_NtkFromDar(pNtk, pMan)
			for _, pObj := range pNtkAig.ObjectsOfKind(abc.ObjectLatch) {
				abc.Abc_LatchSetInit0(pObj)
			}
		}
		Aig_ManStop(pMan)
	}
	return pNtkAig
}

/*
#include <signal.h>
#include "misc/util/utilMem.h"
static void sigfunc( int signo )
{
    if (signo == SIGINT) {
        Abc_Print( 1, "SIGINT received!\n");
        s_fInterrupt = 1;
    }
}
*/

/*
*Function*************************************************************

	Synopsis    [Gives the current ABC network to AIG manager for processing.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkDarBmc(pNtk *Abc_Ntk_t, nStart int32, nFrames int32, nSizeMax int32, nNodeDelta int32, nTimeOut int32, nBTLimit int32, nBTLimitAll int32, fRewrite int32, fNewAlgo int32, fOrDecomp int32, nCofFanLit int32, fVerbose int32, piFrames *int32, fUseSatoko int32) int32 {
	var pMan *Aig_Man_t
	var vMap *Vec_Int_t = nil
	var status int32
	var RetValue int32 = -1
	var clk abctime = Abc_Clock()
	var nTimeLimit abctime
	if nTimeOut != 0 {
		nTimeLimit = int64(nTimeOut)*int64(time.Second) + Abc_Clock()
	} else {
		nTimeLimit = 0
	}
	// derive the AIG manager
	if fOrDecomp != 0 {
		pMan = Abc_NtkToDarBmc(pNtk, &vMap)
	} else {
		pMan = Abc_NtkToDar(pNtk, 0, 1)
	}
	if pMan == nil {
		Abc_Print(1, "Converting miter into AIG has failed.\n")
		return RetValue
	}
	assert(pMan.RegisterCount() > 0, "pMan.RegisterCount() > 0")
	assert(vMap == nil || Vec_IntSize(vMap) == Saig_ManPoNum(pMan), "vMap == nil || Vec_IntSize(vMap) == Saig_ManPoNum(pMan)")
	if (fVerbose != 0) && (vMap != nil) && Abc_NtkPoNum(pNtk) != Saig_ManPoNum(pMan) {
		Abc_Print(1, "Expanded %d outputs into %d outputs using OR decomposition.\n", Abc_NtkPoNum(pNtk), Saig_ManPoNum(pMan))
	}

	// perform verification
	if fNewAlgo != 0 { // command 'bmc'

		var iFrame int32
		RetValue = sat_bmc.Saig_ManBmcSimple(pMan, nFrames, nSizeMax, nBTLimit, fRewrite, fVerbose, &iFrame, nCofFanLit, fUseSatoko)
		if piFrames != nil {
			*piFrames = iFrame
		}
		Abc_NtkSetModel(pNtk, nil)
		Abc_NtkSetSeqModel(pNtk, nil)
		Abc_NtkSetSeqModel(pNtk, pMan.SequentialCounterexample())
		pMan.ClearSequentialCounterexample()
		if RetValue == 1 {
			Abc_Print(1, "Incorrect return value.  ")
		} else if RetValue == -1 {
			Abc_Print(1, "No output asserted in %d frames. Resource limit reached ", Abc_MaxInt(iFrame+1, 0))
			if (nTimeLimit != 0) && Abc_Clock() > nTimeLimit {
				Abc_Print(1, "(timeout %d sec). ", nTimeLimit)
			} else {
				Abc_Print(1, "(conf limit %d). ", nBTLimit)
			}

		} else { // if ( RetValue == 0 )

			var pCex *Abc_Cex_t = Abc_NtkReadSeqModel(pNtk)
			Abc_Print(1, "Output %d of miter \"%s\" was asserted in frame %d. ", pCex.IPo, Abc_NtkName(pNtk), pCex.IFrame)
		}
		ABC_PRT("Time", Abc_Clock()-clk)

	} else {
		RetValue = sat_bmc.Saig_BmcPerform(pMan, nStart, nFrames, nNodeDelta, nTimeOut, nBTLimit, nBTLimitAll, fVerbose, 0, piFrames, 0, fUseSatoko)
		Abc_NtkSetModel(pNtk, nil)
		Abc_NtkSetSeqModel(pNtk, nil)
		Abc_NtkSetSeqModel(pNtk, pMan.SequentialCounterexample())
		pMan.ClearSequentialCounterexample()
	}
	// verify counter-example
	if Abc_NtkReadSeqModel(pNtk) != nil {
		status = Saig_ManVerifyCex(pMan, Abc_NtkReadSeqModel(pNtk))
		if status == 0 {
			Abc_Print(1, "Abc_NtkDarBmc(): Counter-example verification has FAILED.\n")
		}
	}
	Aig_ManStop(pMan)
	// update the counter-example
	if (Abc_NtkReadSeqModel(pNtk) != nil) && (vMap != nil) {
		Abc_NtkReadSeqModel(pNtk).IPo = Vec_IntEntry(vMap, Abc_NtkReadSeqModel(pNtk).IPo)
	}
	Vec_IntFreeP(&vMap)
	return RetValue
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkDarBmc3(pNtk *Abc_Ntk_t, pPars *Saig_ParBmc_t, fOrDecomp int32) int32 {
	var pMan *Aig_Man_t
	var vMap *Vec_Int_t = nil
	var status int32
	var RetValue int32 = -1
	var clk abctime = Abc_Clock()
	var nTimeOut abctime
	if pPars.NTimeOut != 0 {
		nTimeOut = int64(pPars.NTimeOut)*int64(time.Second) + Abc_Clock()
	} else {
		nTimeOut = 0
	}
	if (fOrDecomp != 0) && (pPars.FSolveAll == 0) {
		pMan = Abc_NtkToDarBmc(pNtk, &vMap)
	} else {
		pMan = Abc_NtkToDar(pNtk, 0, 1)
	}
	if pMan == nil {
		Abc_Print(1, "Converting miter into AIG has failed.\n")
		return RetValue
	}
	assert(pMan.RegisterCount() > 0, "pMan.RegisterCount() > 0")
	if (pPars.FVerbose != 0) && (vMap != nil) && Abc_NtkPoNum(pNtk) != Saig_ManPoNum(pMan) {
		Abc_Print(1, "Expanded %d outputs into %d outputs using OR decomposition.\n", Abc_NtkPoNum(pNtk), Saig_ManPoNum(pMan))
	}

	RetValue = sat_bmc.Saig_ManBmcScalable(pMan, pPars)
	Abc_NtkSetModel(pNtk, nil)
	Abc_NtkSetSeqModel(pNtk, nil)
	Abc_NtkSetSeqModel(pNtk, pMan.SequentialCounterexample())
	pMan.ClearSequentialCounterexample()
	if pPars.FSilent == 0 {
		if RetValue == 1 {
			Abc_Print(1, "Explored all reachable states after completing %d frames.  ", 1<<aiggraph.Aig_ManRegNum(pMan))

		} else if RetValue == -1 {
			if pPars.NFailOuts == 0 {
				Abc_Print(1, "No output asserted in %d frames. Resource limit reached ", Abc_MaxInt(pPars.IFrame+1, 0))
				if (nTimeOut != 0) && Abc_Clock() > nTimeOut {
					Abc_Print(1, "(timeout %d sec). ", pPars.NTimeOut)
				} else {
					Abc_Print(1, "(conf limit %d). ", pPars.NConfLimit)
				}

			} else {
				Abc_Print(1, "The total of %d outputs asserted in %d frames. Resource limit reached ", pPars.NFailOuts, pPars.IFrame)
				if Abc_Clock() > nTimeOut {
					Abc_Print(1, "(timeout %d sec). ", pPars.NTimeOut)
				} else {
					Abc_Print(1, "(conf limit %d). ", pPars.NConfLimit)
				}
			}

		} else { // if ( RetValue == 0 )

			if pPars.FSolveAll == 0 {
				var pCex *Abc_Cex_t = Abc_NtkReadSeqModel(pNtk)
				Abc_Print(1, "Output %d of miter \"%s\" was asserted in frame %d. ", pCex.IPo, Abc_NtkName(pNtk), pCex.IFrame)

			} else {
				var nOutputs int32 = Saig_ManPoNum(pMan) - aig_saig.Saig_ManConstrNum(pMan)
				if pMan.SequentialCounterexamples() == nil || int32(giapkg.Vec_PtrCountZero(pMan.SequentialCounterexamples())) == nOutputs {
					Abc_Print(1, "None of the %d outputs is found to be SAT", nOutputs)
				} else if int32(giapkg.Vec_PtrCountZero(pMan.SequentialCounterexamples())) == 0 {
					Abc_Print(1, "All %d outputs are found to be SAT", nOutputs)
				} else {
					Abc_Print(1, "Some outputs are SAT (%d out of %d)", nOutputs-int32(giapkg.Vec_PtrCountZero(pMan.SequentialCounterexamples())), nOutputs)
					if pPars.NDropOuts != 0 {
						Abc_Print(1, " while others timed out (%d out of %d)", pPars.NDropOuts, nOutputs)
					}
				}
				Abc_Print(1, " after %d frames", pPars.IFrame+2)
				Abc_Print(1, ".   ")
			}
		}
		ABC_PRT("Time", Abc_Clock()-clk)
	}
	if RetValue == 0 && (pPars.FSolveAll != 0) {
		if abc.Abc_NtkReadSeqModelVec(pNtk) != nil {
			Vec_PtrFreeFree(abc.Abc_NtkReadSeqModelVec(pNtk))
		}
		abc.Abc_NtkSetSeqModelVec(pNtk, pMan.TakeSequentialCounterexamples())
	}
	if Abc_NtkReadSeqModel(pNtk) != nil {
		status = Saig_ManVerifyCex(pMan, Abc_NtkReadSeqModel(pNtk))
		if status == 0 {
			Abc_Print(1, "Abc_NtkDarBmc3(): Counter-example verification has FAILED.\n")
		}
	}
	Aig_ManStop(pMan)
	// update the counter-example
	if (Abc_NtkReadSeqModel(pNtk) != nil) && (vMap != nil) {
		Abc_NtkReadSeqModel(pNtk).IPo = Vec_IntEntry(vMap, Abc_NtkReadSeqModel(pNtk).IPo)
	}
	Vec_IntFreeP(&vMap)
	return RetValue
}

/*
*Function*************************************************************

	Synopsis    [Gives the current ABC network to AIG manager for processing.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkDarBmcInter_int(pMan *Aig_Man_t, pPars *Inter_ManParams_t, ppNtkRes **Aig_Man_t) int32 {
	var RetValue int32 = -1
	var iFrame int32
	var clk abctime = Abc_Clock()
	var nTotalProvedSat int32
	assert(pMan.RegisterCount() > 0, "pMan.RegisterCount() > 0")
	if ppNtkRes != nil {
		*ppNtkRes = nil
	}
	if pPars.FUseSeparate != 0 {
		var pTemp, pAux *Aig_Man_t
		var pObjPo aiggraph.Edge
		var i, Counter int32
		for i = 0; i < Saig_ManPoNum(pMan); i++ {
			pObjPo = pMan.Outputs()[i]
			if aiggraph.Aig_ObjFanin0(pMan, pObjPo) == aiggraph.Aig_ManConst1(pMan) {
				continue
			}
			if pPars.FVerbose != 0 {
				Abc_Print(1, "Solving output %2d (out of %2d):\n", i, Saig_ManPoNum(pMan))
			}
			pTemp = aiggraph.Aig_ManDupOneOutput(pMan, i, 1)
			pAux = pTemp
			pTemp = aiggraph.Aig_ManScl(pAux, 1, 1, 0, -1, -1, 0, 0)
			Aig_ManStop(pAux)
			if aiggraph.Aig_ManRegNum(pTemp) == 0 {
				pTemp.ClearSequentialCounterexample()
				RetValue = proof_fra.Fra_FraigSat(pTemp, int64(pPars.NBTLimit), 0, 0, 0, 0, 0, 0, 0, 0)
				if pTemp.Data != nil {
					pTemp.SetCecCounterexample(misc_util.Abc_CexCreate(aiggraph.Aig_ManRegNum(pMan), aig_saig.Saig_ManPiNum(pMan), abcAigModel(pTemp), 0, i, 1))
				}
				//                pNtk->pModel = pTemp->pData, pTemp->pData = NULL;

			} else {
				RetValue = proof_int.Inter_ManPerformInterpolation(pTemp, pPars, &iFrame)
			}
			if pTemp.SequentialCounterexample() != nil {
				if pPars.FDropSatOuts != 0 {
					Abc_Print(1, "Output %d proved SAT in frame %d (replacing by const 0 and continuing...)\n", i, pTemp.SequentialCounterexample().IFrame)
					aiggraph.Aig_ObjPatchFanin0(pMan, pObjPo, aiggraph.Aig_ManConst0(pMan))
					Aig_ManStop(pTemp)
					nTotalProvedSat++
					continue

				} else {
					var pCex *Abc_Cex_t
					pCex = pTemp.SequentialCounterexample()
					pMan.SetCecCounterexample(pCex)
					pTemp.ClearSequentialCounterexample()
					pCex.IPo = i
					Aig_ManStop(pTemp)
					break
				}
			}
			// if solved, remove the output
			if RetValue == 1 {
				aiggraph.Aig_ObjPatchFanin0(pMan, pObjPo, aiggraph.Aig_ManConst0(pMan))
				//                    Abc_Print( 1, "Output %3d : Solved ", i );

			} else {
				Counter++
				//                    Abc_Print( 1, "Output %3d : Undec  ", i );
			}
			//                Aig_ManPrintStats( pTemp );
			Aig_ManStop(pTemp)
			Abc_Print(1, "Solving output %3d (out of %3d) using interpolation.\r", i, Saig_ManPoNum(pMan))
		}
		aiggraph.Aig_ManCleanup(pMan)
		if pMan.SequentialCounterexample() == nil {
			Abc_Print(1, "Interpolation left %d (out of %d) outputs unsolved              \n", Counter, Saig_ManPoNum(pMan))
			if Counter != 0 {
				RetValue = -1
			}
		}
		if ppNtkRes != nil {
			pTemp = aiggraph.Aig_ManDupUnsolvedOutputs(pMan, 1)
			*ppNtkRes = aiggraph.Aig_ManScl(pTemp, 1, 1, 0, -1, -1, 0, 0)
			Aig_ManStop(pTemp)
		}

	} else {
		RetValue = proof_int.Inter_ManPerformInterpolation(pMan, pPars, &iFrame)
	}
	if nTotalProvedSat != 0 {
		Abc_Print(1, "The total of %d outputs proved SAT and replaced by const 0 in this run.\n", nTotalProvedSat)
	}
	if RetValue == 1 {
		Abc_Print(1, "Property proved.  ")
	} else if RetValue == 0 {
		out := int32(-1)
		if pMan.SequentialCounterexample() != nil {
			out = pMan.SequentialCounterexample().IPo
		}
		Abc_Print(1, "Output %d of miter \"%s\" was asserted in frame %d.  ", out, pMan.Name, iFrame)
	} else if RetValue == -1 {
		Abc_Print(1, "Property UNDECIDED.  ")
	} else {
		assert(false, "0")
	}
	ABC_PRT("Time", Abc_Clock()-clk)
	return RetValue
}

/*
*Function*************************************************************

	Synopsis    [Gives the current ABC network to AIG manager for processing.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkDarBmcInter(pNtk *Abc_Ntk_t, pPars *Inter_ManParams_t, ppNtkRes **Abc_Ntk_t) int32 {
	var pMan *Aig_Man_t
	var RetValue int32
	if ppNtkRes != nil {
		*ppNtkRes = nil
	}
	pMan = Abc_NtkToDar(pNtk, 0, 1)
	if pMan == nil {
		Abc_Print(1, "Converting miter into AIG has failed.\n")
		return -1
	}
	if (pPars.FUseSeparate != 0) && (ppNtkRes != nil) {
		var pManNew *Aig_Man_t
		RetValue = Abc_NtkDarBmcInter_int(pMan, pPars, &pManNew)
		*ppNtkRes = Abc_NtkFromAigPhase(pManNew)
		Aig_ManStop(pManNew)

	} else {
		RetValue = Abc_NtkDarBmcInter_int(pMan, pPars, nil)
	}
	Abc_NtkSetModel(pNtk, nil)
	Abc_NtkSetSeqModel(pNtk, nil)
	Abc_NtkSetSeqModel(pNtk, pMan.SequentialCounterexample())
	pMan.ClearSequentialCounterexample()
	Aig_ManStop(pMan)
	return RetValue
}

/*
*Function*************************************************************

	Synopsis    [Gives the current ABC network to AIG manager for processing.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkDarDemiter(pNtk *Abc_Ntk_t) int32 {
	var pFileName0 string // was [1000]byte in C

	var pFileName1 string // was [1000]byte in C

	var pMan, pPart0, pPart1 *Aig_Man_t //, * pMiter;
	// derive the AIG manager
	pMan = Abc_NtkToDar(pNtk, 0, 1)
	if pMan == nil {
		Abc_Print(1, "Converting network into AIG has failed.\n")
		return 0
	}
	//    if ( !Saig_ManDemiterSimple( pMan, &pPart0, &pPart1 ) )
	if aig_saig.Saig_ManDemiterSimpleDiff(pMan, &pPart0, &pPart1) == 0 {
		Aig_ManStop(pMan)
		Abc_Print(1, "Demitering has failed.\n")
		return 0
	}
	pFileName0, pFileName1 = "part0.aig", "part1.aig"
	// dump files
	aig_ioa.Ioa_WriteAiger(pPart0, pFileName0, 0, 0)
	aig_ioa.Ioa_WriteAiger(pPart1, pFileName1, 0, 0)
	Abc_Print(1, "Demitering produced two files \"%s\" and \"%s\".\n", pFileName0, pFileName1)
	// create two-level miter
	//    pMiter = Saig_ManCreateMiterTwo( pPart0, pPart1, 2 );
	//    Aig_ManDumpBlif( pMiter, "miter01.blif", NULL, NULL );
	//    Aig_ManStop( pMiter );
	//    Abc_Print( 1, "The new miter is written into file \"%s\".\n", "miter01.blif" );
	Aig_ManStop(pPart0)
	Aig_ManStop(pPart1)
	Aig_ManStop(pMan)
	return 1
}

/*
*Function*************************************************************

	Synopsis    [Gives the current ABC network to AIG manager for processing.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkDarDemiterNew(pNtk *Abc_Ntk_t) int32 {
	var pFileNameGeneric string
	var pFileName0 string // was [1000]byte in C

	var pFileName1 string // was [1000]byte in C

	var pMan, pPart0, pPart1 *Aig_Man_t //, * pMiter;
	// derive the AIG manager
	pMan = Abc_NtkToDar(pNtk, 0, 1)
	if pMan == nil {
		Abc_Print(1, "Converting network into AIG has failed.\n")
		return 0
	}

	aig_saig.Saig_ManDemiterNew(pMan)
	Aig_ManStop(pMan)
	return 1

	//    if ( !Saig_ManDemiterSimple( pMan, &pPart0, &pPart1 ) )
	if aig_saig.Saig_ManDemiterSimpleDiff(pMan, &pPart0, &pPart1) == 0 {
		Abc_Print(1, "Demitering has failed.\n")
		return 0
	}
	// create file names
	pFileNameGeneric = Extra_FileNameGeneric(abc.Abc_NtkSpec(pNtk))
	pFileName0 = fmt.Sprintf("%s%s", pFileNameGeneric, "_part0.aig")
	pFileName1 = fmt.Sprintf("%s%s", pFileNameGeneric, "_part1.aig")
	pFileNameGeneric = ""
	// dump files
	aig_ioa.Ioa_WriteAiger(pPart0, pFileName0, 0, 0)
	aig_ioa.Ioa_WriteAiger(pPart1, pFileName1, 0, 0)
	Abc_Print(1, "Demitering produced two files \"%s\" and \"%s\".\n", pFileName0, pFileName1)
	// create two-level miter
	//    pMiter = Saig_ManCreateMiterTwo( pPart0, pPart1, 2 );
	//    Aig_ManDumpBlif( pMiter, "miter01.blif", NULL, NULL );
	//    Aig_ManStop( pMiter );
	//    Abc_Print( 1, "The new miter is written into file \"%s\".\n", "miter01.blif" );
	Aig_ManStop(pPart0)
	Aig_ManStop(pPart1)
	Aig_ManStop(pMan)
	return 1
}

/*
*Function*************************************************************

	Synopsis    [Gives the current ABC network to AIG manager for processing.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkDarDemiterDual(pNtk *Abc_Ntk_t, fVerbose int32) int32 {
	var pFileName0 string // was [1000]byte in C

	var pFileName1 string // was [1000]byte in C

	var pMan, pPart0, pPart1 *Aig_Man_t //, * pMiter;
	if Abc_NtkPoNum(pNtk)&1 != 0 {
		Abc_Print(1, "The number of POs should be even.\n")
		return 0
	}
	// derive the AIG manager
	pMan = Abc_NtkToDar(pNtk, 0, 1)
	if pMan == nil {
		Abc_Print(1, "Converting network into AIG has failed.\n")
		return 0
	}
	//    if ( !Saig_ManDemiterSimple( pMan, &pPart0, &pPart1 ) )
	if aig_saig.Saig_ManDemiterDual(pMan, &pPart0, &pPart1) == 0 {
		Abc_Print(1, "Demitering has failed.\n")
		return 0
	}
	// create new AIG
	pPart0.Name = "part0"
	// create new AIGs
	pPart1.Name = "part1"
	// create file names
	//    sprintf( pFileName0,  "%s%s",  pFileNameGeneric, "_part0.aig" );
	//    sprintf( pFileName1,  "%s%s",  pFileNameGeneric, "_part1.aig" );
	pFileName0 = fmt.Sprintf("%s", "part0.aig")
	pFileName1 = fmt.Sprintf("%s", "part1.aig")
	aig_ioa.Ioa_WriteAiger(pPart0, pFileName0, 0, 0)
	aig_ioa.Ioa_WriteAiger(pPart1, pFileName1, 0, 0)
	Abc_Print(1, "Demitering produced two files \"%s\" and \"%s\".\n", pFileName0, pFileName1)
	// dump files
	if fVerbose != 0 {
		//        Abc_Print( 1, "Init:  " );
		aiggraph.Aig_ManPrintStats(pMan)
		//        Abc_Print( 1, "Part1: " );
		aiggraph.Aig_ManPrintStats(pPart0)
		//        Abc_Print( 1, "Part2: " );
		aiggraph.Aig_ManPrintStats(pPart1)
	}
	// create two-level miter
	//    pMiter = Saig_ManCreateMiterTwo( pPart0, pPart1, 2 );
	//    Aig_ManDumpBlif( pMiter, "miter01.blif", NULL, NULL );
	//    Aig_ManStop( pMiter );
	//    Abc_Print( 1, "The new miter is written into file \"%s\".\n", "miter01.blif" );
	Aig_ManStop(pPart0)
	Aig_ManStop(pPart1)
	Aig_ManStop(pMan)
	return 1
}

/*
*Function*************************************************************

	Synopsis    [Gives the current ABC network to AIG manager for processing.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkDarProve(pNtk *Abc_Ntk_t, pSecPar *Fra_Sec_t, nBmcFramesMax int32, nBmcConfMax int32) int32 {
	var pMan *Aig_Man_t
	var iFrame int32 = -1
	var RetValue int32 = -1
	var clkTotal abctime = Abc_Clock()
	if (pSecPar.FTryComb != 0) || Abc_NtkLatchNum(pNtk) == 0 {
		var Params Prove_Params_t
		var pParams *Prove_Params_t = &Params
		var pNtkComb *Abc_Ntk_t
		var RetValue int32
		if Abc_NtkLatchNum(pNtk) == 0 {
			Abc_Print(1, "The network has no latches. Running CEC.\n")
		}
		// create combinational network
		pNtkComb = Abc_NtkDup(pNtk)
		abc.Abc_NtkMakeComb(pNtkComb, 1)
		// solve it using combinational equivalence checking
		Prove_ParamsSetDefault(pParams)
		pParams.fVerbose = 1
		RetValue = Abc_NtkIvyProve(&pNtkComb, pParams)
		// transfer model if given
		//        pNtk->pModel = pNtkComb->pModel; pNtkComb->pModel = NULL;
		if RetValue == 0 && (Abc_NtkLatchNum(pNtk) == 0) {
			Abc_NtkSetModel(pNtk, Abc_NtkReadModel(pNtkComb))
			Abc_NtkSetModel(pNtkComb, nil)
			if pSecPar.FReportSolution != 0 {
				Abc_Print(1, "SOLUTION: FAIL       ")
			} else {
				Abc_Print(1, "SATISFIABLE    ")
			}
			ABC_PRT("Time", Abc_Clock()-clkTotal)
			return RetValue
		}
		Abc_NtkDelete(pNtkComb)
		// return the result, if solved
		if RetValue == 1 {
			if pSecPar.FReportSolution != 0 {
				Abc_Print(1, "SOLUTION: PASS       ")
			} else {
				Abc_Print(1, "UNSATISFIABLE  ")
			}
			ABC_PRT("Time", Abc_Clock()-clkTotal)
			return RetValue
		}
		// return undecided, if the miter is combinational
		if Abc_NtkLatchNum(pNtk) == 0 {
			Abc_Print(1, "UNDECIDED      ")
			ABC_PRT("Time", Abc_Clock()-clkTotal)
			return RetValue
		}
	}
	// derive the AIG manager
	pMan = Abc_NtkToDar(pNtk, 0, 1)
	if pMan == nil {
		Abc_Print(1, "Converting miter into AIG has failed.\n")
		return -1
	}
	assert(pMan.RegisterCount() > 0, "pMan.RegisterCount() > 0")

	if pSecPar.FTryBmc != 0 {
		RetValue = sat_bmc.Saig_BmcPerform(pMan, 0, nBmcFramesMax, 2000, 0, nBmcConfMax, 0, pSecPar.FVerbose, 0, &iFrame, 0, 0)
		if RetValue == 0 {
			Abc_Print(1, "Networks are not equivalent.\n")
			if pSecPar.FReportSolution != 0 {
				Abc_Print(1, "SOLUTION: FAIL       ")
				ABC_PRT("Time", Abc_Clock()-clkTotal)
			}
			// return the counter-example generated
			Abc_NtkSetModel(pNtk, nil)
			Abc_NtkSetSeqModel(pNtk, nil)
			Abc_NtkSetSeqModel(pNtk, pMan.SequentialCounterexample())
			pMan.ClearSequentialCounterexample()
			Aig_ManStop(pMan)
			return RetValue
		}
	}
	// perform verification
	if pSecPar.FUseNewProver != 0 {
		RetValue = proof_ssw.Ssw_SecGeneralMiter(pMan, nil)

	} else {
		RetValue = proof_fra.Fra_FraigSec(pMan, pSecPar, nil)
		Abc_NtkSetModel(pNtk, nil)
		Abc_NtkSetSeqModel(pNtk, nil)
		Abc_NtkSetSeqModel(pNtk, pMan.SequentialCounterexample())
		pMan.ClearSequentialCounterexample()
		if Abc_NtkReadSeqModel(pNtk) != nil {
			var pCex *Abc_Cex_t = Abc_NtkReadSeqModel(pNtk)
			Abc_Print(1, "Output %d of miter \"%s\" was asserted in frame %d.\n", pCex.IPo, Abc_NtkName(pNtk), pCex.IFrame)
			if Saig_ManVerifyCex(pMan, Abc_NtkReadSeqModel(pNtk)) == 0 {
				Abc_Print(1, "Abc_NtkDarProve(): Counter-example verification has FAILED.\n")
			}
		}
	}
	Aig_ManStop(pMan)
	return RetValue
}

/*
*Function*************************************************************

	Synopsis    [Gives the current ABC network to AIG manager for processing.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkDarSec(pNtk1 *Abc_Ntk_t, pNtk2 *Abc_Ntk_t, pSecPar *Fra_Sec_t) int32 {
	//    Fraig_Params_t Params;
	var pMan *Aig_Man_t
	var pMiter *Abc_Ntk_t //, * pTemp;
	var RetValue int32

	// get the miter of the two networks
	pMiter = Abc_NtkMiter(pNtk1, pNtk2, 0, 0, 0, 0)
	if pMiter == nil {
		Abc_Print(1, "Miter computation has failed.\n")
		return -1
	}
	RetValue = Abc_NtkMiterIsConstant(pMiter)
	if RetValue == 0 {
		Abc_Print(1, "Networks are NOT EQUIVALENT after structural hashing.\n")
		// report the error
		Abc_NtkSetModel(pMiter, Abc_NtkVerifyGetCleanModel(pMiter, pSecPar.NFramesMax))
		//        Abc_NtkVerifyReportErrorSeq( pNtk1, pNtk2, pMiter->pModel, pSecPar->nFramesMax );
		Abc_NtkSetModel(pMiter, nil)
		Abc_NtkDelete(pMiter)
		return 0
	}
	if RetValue == 1 {
		Abc_NtkDelete(pMiter)
		Abc_Print(1, "Networks are equivalent after structural hashing.\n")
		return 1
	}

	// commented out because sometimes the problem became non-inductive
	/*
	   // preprocess the miter by fraiging it
	   // (note that for each functional class, fraiging leaves one representative;
	   // so fraiging does not reduce the number of functions represented by nodes
	   Fraig_ParamsSetDefault( &Params );
	   Params.nBTLimit = 100000;
	   pMiter = Abc_NtkFraig( pTemp = pMiter, &Params, 0, 0 );
	   Abc_NtkDelete( pTemp );
	   RetValue = Abc_NtkMiterIsConstant( pMiter );
	   if ( RetValue == 0 )
	   {
	       Abc_Print( 1, "Networks are NOT EQUIVALENT after structural hashing.\n" );
	       // report the error
	       pMiter->pModel = Abc_NtkVerifyGetCleanModel( pMiter, nFrames );
	       Abc_NtkVerifyReportErrorSeq( pNtk1, pNtk2, pMiter->pModel, nFrames );
	       ABC_FREE( pMiter->pModel );
	       Abc_NtkDelete( pMiter );
	       return 0;
	   }
	   if ( RetValue == 1 )
	   {
	       Abc_NtkDelete( pMiter );
	       Abc_Print( 1, "Networks are equivalent after structural hashing.\n" );
	       return 1;
	   }
	*/
	// derive the AIG manager
	pMan = Abc_NtkToDar(pMiter, 0, 1)
	Abc_NtkDelete(pMiter)
	if pMan == nil {
		Abc_Print(1, "Converting miter into AIG has failed.\n")
		return -1
	}
	assert(pMan.RegisterCount() > 0, "pMan.RegisterCount() > 0")

	// perform verification
	RetValue = proof_fra.Fra_FraigSec(pMan, pSecPar, nil)
	Aig_ManStop(pMan)
	return RetValue
}

/*
*Function*************************************************************

	Synopsis    [Gives the current ABC network to AIG manager for processing.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkDarPdr(pNtk *Abc_Ntk_t, pPars *Pdr_Par_t) int32 {
	var RetValue int32 = -1
	var clk abctime = Abc_Clock()
	var pMan *Aig_Man_t
	pMan = Abc_NtkToDar(pNtk, 0, 1)
	if pMan == nil {
		Abc_Print(1, "Converting network into AIG has failed.\n")
		return -1
	}
	RetValue = proof_pdr.Pdr_ManSolve(pMan, pPars)
	pPars.NDropOuts = Saig_ManPoNum(pMan) - pPars.NProveOuts - pPars.NFailOuts
	if pPars.FSilent == 0 {
		if pPars.FSolveAll != 0 {
			Abc_Print(1, "Properties:  All = %d. Proved = %d. Disproved = %d. Undecided = %d.   ",
				Saig_ManPoNum(pMan), pPars.NProveOuts, pPars.NFailOuts, pPars.NDropOuts)
		} else if RetValue == 1 {
			Abc_Print(1, "Property proved.  ")
		} else {
			if RetValue == 0 {
				if pMan.SequentialCounterexample() == nil {
					Abc_Print(1, "Abc_NtkDarPdr(): Counter-example is not available.\n")
				} else {
					Abc_Print(1, "Output %d of miter \"%s\" was asserted in frame %d.  ", pMan.SequentialCounterexample().IPo, Abc_NtkName(pNtk), pMan.SequentialCounterexample().IFrame)
					if Saig_ManVerifyCex(pMan, pMan.SequentialCounterexample()) == 0 {
						Abc_Print(1, "Abc_NtkDarPdr(): Counter-example verification has FAILED.\n")
					}
				}

			} else if RetValue == -1 {
				Abc_Print(1, "Property UNDECIDED.  ")
			} else {
				assert(false, "0")
			}
		}
		ABC_PRT("Time", Abc_Clock()-clk)
		/*
		   Abc_Print( 1, "Status: " );
		   if ( pPars.pOutMap )
		   {
		       int i;
		       for ( i = 0; i < Saig_ManPoNum(pMan); i++ )
		           if ( pPars.pOutMap[i] == 1 )
		               Abc_Print( 1, "%d=%s ", i, "unsat" );
		           else if ( pPars.pOutMap[i] == 0 )
		               Abc_Print( 1, "%d=%s ", i, "sat" );
		           else if ( pPars.pOutMap[i] < 0 )
		               Abc_Print( 1, "%d=%s ", i, "undec" );
		           else assert( 0 );
		   }
		   Abc_Print( 1, "\n" );
		*/
	}
	Abc_NtkSetSeqModel(pNtk, nil)
	Abc_NtkSetSeqModel(pNtk, pMan.SequentialCounterexample())
	pMan.ClearSequentialCounterexample()
	if abc.Abc_NtkReadSeqModelVec(pNtk) != nil {
		Vec_PtrFreeFree(abc.Abc_NtkReadSeqModelVec(pNtk))
	}
	abc.Abc_NtkSetSeqModelVec(pNtk, pMan.TakeSequentialCounterexamples())
	Aig_ManStop(pMan)
	return RetValue
}

/*
*Function*************************************************************

	Synopsis    [Performs BDD-based reachability analysis.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkDarAbSec(pNtk1 *Abc_Ntk_t, pNtk2 *Abc_Ntk_t, nFrames int32, fVerbose int32) int32 {
	var pMan1 *Aig_Man_t
	var pMan2 *Aig_Man_t = nil
	var RetValue int32
	// derive AIG manager
	pMan1 = Abc_NtkToDar(pNtk1, 0, 1)
	if pMan1 == nil {
		Abc_Print(1, "Converting miter into AIG has failed.\n")
		return -1
	}
	assert(aiggraph.Aig_ManRegNum(pMan1) > 0, "Aig_ManRegNum(pMan1) > 0")
	// derive AIG manager
	if pNtk2 != nil {
		pMan2 = Abc_NtkToDar(pNtk2, 0, 1)
		if pMan2 == nil {
			Aig_ManStop(pMan1)
			Abc_Print(1, "Converting miter into AIG has failed.\n")
			return -1
		}
		assert(aiggraph.Aig_ManRegNum(pMan2) > 0, "Aig_ManRegNum(pMan2) > 0")
		if aig_saig.Saig_ManPiNum(pMan1) != aig_saig.Saig_ManPiNum(pMan2) {
			Aig_ManStop(pMan1)
			Aig_ManStop(pMan2)
			Abc_Print(1, "The networks have different number of PIs.\n")
			return -1
		}
		if Saig_ManPoNum(pMan1) != Saig_ManPoNum(pMan2) {
			Aig_ManStop(pMan1)
			Aig_ManStop(pMan2)
			Abc_Print(1, "The networks have different number of POs.\n")
			return -1
		}
		if aiggraph.Aig_ManRegNum(pMan1) != aiggraph.Aig_ManRegNum(pMan2) {
			Aig_ManStop(pMan1)
			Aig_ManStop(pMan2)
			Abc_Print(1, "The networks have different number of flops.\n")
			return -1
		}
	}
	// perform verification
	RetValue = aig_saig.Ssw_SecSpecialMiter(pMan1, pMan2, nFrames, fVerbose)
	Aig_ManStop(pMan1)
	if pMan2 != nil {
		Aig_ManStop(pMan2)
	}
	return RetValue
}

/*
*Function*************************************************************

	Synopsis    [Gives the current ABC network to AIG manager for processing.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkDarSimSec(pNtk1 *Abc_Ntk_t, pNtk2 *Abc_Ntk_t, pPars *Ssw_Pars_t) int32 {
	var pMan1 *Aig_Man_t
	var pMan2 *Aig_Man_t = nil
	var RetValue int32
	// derive AIG manager
	pMan1 = Abc_NtkToDar(pNtk1, 0, 1)
	if pMan1 == nil {
		Abc_Print(1, "Converting miter into AIG has failed.\n")
		return -1
	}
	assert(aiggraph.Aig_ManRegNum(pMan1) > 0, "Aig_ManRegNum(pMan1) > 0")
	// derive AIG manager
	if pNtk2 != nil {
		pMan2 = Abc_NtkToDar(pNtk2, 0, 1)
		if pMan2 == nil {
			Abc_Print(1, "Converting miter into AIG has failed.\n")
			return -1
		}
		assert(aiggraph.Aig_ManRegNum(pMan2) > 0, "Aig_ManRegNum(pMan2) > 0")
	}

	// perform verification
	RetValue = proof_ssw.Ssw_SecWithSimilarity(pMan1, pMan2, pPars)
	Aig_ManStop(pMan1)
	if pMan2 != nil {
		Aig_ManStop(pMan2)
	}
	return RetValue
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkDarMatch(pNtk1 *Abc_Ntk_t, pNtk2 *Abc_Ntk_t, nDist int32, fVerbose int32) *Abc_Ntk_t {
	var pNtkAig *Abc_Ntk_t
	var pMan1 *Aig_Man_t
	var pMan2 *Aig_Man_t = nil
	var pManRes *Aig_Man_t
	var vPairs *Vec_Int_t
	assert((Abc_NtkIsStrash(pNtk1) != 0), "Abc_NtkIsStrash(pNtk1)")
	// derive AIG manager
	pMan1 = Abc_NtkToDar(pNtk1, 0, 1)
	if pMan1 == nil {
		Abc_Print(1, "Converting miter into AIG has failed.\n")
		return nil
	}
	assert(aiggraph.Aig_ManRegNum(pMan1) > 0, "Aig_ManRegNum(pMan1) > 0")
	// derive AIG manager
	if pNtk2 != nil {
		pMan2 = Abc_NtkToDar(pNtk2, 0, 1)
		if pMan2 == nil {
			Abc_Print(1, "Converting miter into AIG has failed.\n")
			return nil
		}
		assert(aiggraph.Aig_ManRegNum(pMan2) > 0, "Aig_ManRegNum(pMan2) > 0")
	}

	// perform verification
	vPairs = aig_saig.Saig_StrSimPerformMatching(pMan1, pMan2, nDist, 1, &pManRes)
	pNtkAig = Abc_NtkFromAigPhase(pManRes)
	if vPairs != nil {
		Vec_IntFree(vPairs)
	}
	if pManRes != nil {
		Aig_ManStop(pManRes)
	}
	Aig_ManStop(pMan1)
	if pMan2 != nil {
		Aig_ManStop(pMan2)
	}
	return pNtkAig
}

/*
*Function*************************************************************

	Synopsis    [Gives the current ABC network to AIG manager for processing.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkDarLatchSweep(pNtk *Abc_Ntk_t, fLatchConst int32, fLatchEqual int32, fSaveNames int32, fUseMvSweep int32, nFramesSymb int32, nFramesSatur int32, fVerbose int32, fVeryVerbose int32) *Abc_Ntk_t {
	var pNtkAig *Abc_Ntk_t
	var pMan, pTemp *Aig_Man_t
	pMan = Abc_NtkToDar(pNtk, 0, 1)
	if pMan == nil {
		return nil
	}
	if fSaveNames != 0 {
		aiggraph.Aig_ManSeqCleanup(pMan)
		if (fLatchConst != 0) && (pMan.RegisterCount() != 0) {
			pMan = aiggraph.Aig_ManConstReduce(pMan, fUseMvSweep, nFramesSymb, nFramesSatur, fVerbose, fVeryVerbose)
		}
		if (fLatchEqual != 0) && (pMan.RegisterCount() != 0) {
			pMan = aiggraph.Aig_ManReduceLaches(pMan, fVerbose)
		}

	} else {
		if pMan.FlopNumbers != nil {
			pMan.FlopNumbers = nil
		}
		pMan.FlopNumbers = nil
		pTemp = pMan
		pMan = aiggraph.Aig_ManScl(pTemp, fLatchConst, fLatchEqual, fUseMvSweep, nFramesSymb, nFramesSatur, fVerbose, fVeryVerbose)
		Aig_ManStop(pTemp)
	}

	pNtkAig = Abc_NtkFromDarSeqSweep(pNtk, pMan)
	//Aig_ManPrintControlFanouts( pMan );
	Aig_ManStop(pMan)
	return pNtkAig
}

/*
*Function*************************************************************

	Synopsis    [Gives the current ABC network to AIG manager for processing.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkDarRetime(pNtk *Abc_Ntk_t, nStepsMax int32, fVerbose int32) *Abc_Ntk_t {
	var pNtkAig *Abc_Ntk_t
	var pMan, pTemp *Aig_Man_t
	pMan = Abc_NtkToDar(pNtk, 0, 1)
	if pMan == nil {
		return nil
	}
	//    Aig_ManReduceLachesCount( pMan );
	if pMan.FlopNumbers != nil {
		pMan.FlopNumbers = nil
	}
	pMan.FlopNumbers = nil

	pTemp = pMan
	pMan = aiggraph.Rtm_ManRetime(pTemp, 1, nStepsMax, fVerbose)
	Aig_ManStop(pTemp)

	//    pMan = Aig_ManReduceLaches( pMan, 1 );
	//    pMan = Aig_ManConstReduce( pMan, 1, 0 );

	pNtkAig = Abc_NtkFromDarSeqSweep(pNtk, pMan)
	Aig_ManStop(pMan)
	return pNtkAig
}

/*
*Function*************************************************************

	Synopsis    [Gives the current ABC network to AIG manager for processing.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkDarRetimeF(pNtk *Abc_Ntk_t, nStepsMax int32, fVerbose int32) *Abc_Ntk_t {
	var pNtkAig *Abc_Ntk_t
	var pMan, pTemp *Aig_Man_t
	pMan = Abc_NtkToDar(pNtk, 0, 1)
	if pMan == nil {
		return nil
	}
	//    Aig_ManReduceLachesCount( pMan );
	if pMan.FlopNumbers != nil {
		pMan.FlopNumbers = nil
	}
	pMan.FlopNumbers = nil

	pTemp = pMan
	pMan = aiggraph.Aig_ManRetimeFrontier(pTemp, nStepsMax)
	Aig_ManStop(pTemp)

	//    pMan = Aig_ManReduceLaches( pMan, 1 );
	//    pMan = Aig_ManConstReduce( pMan, 1, 0 );

	pNtkAig = Abc_NtkFromDarSeqSweep(pNtk, pMan)
	Aig_ManStop(pMan)
	return pNtkAig
}

/*
*Function*************************************************************

	Synopsis    [Gives the current ABC network to AIG manager for processing.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkDarRetimeMostFwd(pNtk *Abc_Ntk_t, nMaxIters int32, fVerbose int32) *Abc_Ntk_t {

	var pNtkAig *Abc_Ntk_t
	var pMan, pTemp *Aig_Man_t
	pMan = Abc_NtkToDar(pNtk, 0, 1)
	if pMan == nil {
		return nil
	}
	//    Aig_ManReduceLachesCount( pMan );
	if pMan.FlopNumbers != nil {
		pMan.FlopNumbers = nil
	}
	pMan.FlopNumbers = nil

	pTemp = pMan
	pMan = aig_saig.Saig_ManRetimeForward(pTemp, nMaxIters, fVerbose)
	Aig_ManStop(pTemp)

	//    pMan = Aig_ManReduceLaches( pMan, 1 );
	//    pMan = Aig_ManConstReduce( pMan, 1, 0 );

	pNtkAig = Abc_NtkFromDarSeqSweep(pNtk, pMan)
	Aig_ManStop(pMan)
	return pNtkAig
}

/*
*Function*************************************************************

	Synopsis    [Gives the current ABC network to AIG manager for processing.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkDarRetimeMinArea(pNtk *Abc_Ntk_t, nMaxIters int32, fForwardOnly int32, fBackwardOnly int32, fInitial int32, fVerbose int32) *Abc_Ntk_t {
	var pNtkAig *Abc_Ntk_t
	var pMan, pTemp *Aig_Man_t
	pMan = Abc_NtkToDar(pNtk, 0, 1)
	if pMan == nil {
		return nil
	}
	if pMan.FlopNumbers != nil {
		pMan.FlopNumbers = nil
	}
	pMan.FlopNumbers = nil

	pTemp = pMan
	pMan = aig_saig.Saig_ManRetimeMinArea(pTemp, nMaxIters, fForwardOnly, fBackwardOnly, fInitial, fVerbose)
	Aig_ManStop(pTemp)

	pNtkAig = Abc_NtkFromDarSeqSweep(pNtk, pMan)
	Aig_ManStop(pMan)
	return pNtkAig
}

/*
*Function*************************************************************

	Synopsis    [Gives the current ABC network to AIG manager for processing.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkDarRetimeStep(pNtk *Abc_Ntk_t, fVerbose int32) *Abc_Ntk_t {
	var pNtkAig *Abc_Ntk_t
	var pMan *Aig_Man_t
	assert((Abc_NtkIsStrash(pNtk) != 0), "Abc_NtkIsStrash(pNtk)")
	pMan = Abc_NtkToDar(pNtk, 0, 1)
	if pMan == nil {
		return nil
	}
	if pMan.FlopNumbers != nil {
		pMan.FlopNumbers = nil
	}
	pMan.FlopNumbers = nil

	aiggraph.Aig_ManPrintStats(pMan)
	aig_saig.Saig_ManRetimeSteps(pMan, 1000, 1, 0)
	aiggraph.Aig_ManPrintStats(pMan)

	pNtkAig = Abc_NtkFromDarSeqSweep(pNtk, pMan)
	Aig_ManStop(pMan)
	return pNtkAig
}

/**Function*************************************************************

  Synopsis    [Gives the current ABC network to AIG manager for processing.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
/*
Abc_Ntk_t * Abc_NtkDarHaigRecord( Abc_Ntk_t * pNtk, int nIters, int nSteps, int fRetimingOnly, int fAddBugs, int fUseCnf, int fVerbose )
{
    Abc_Ntk_t * pNtkAig;
    Aig_Man_t * pMan, * pTemp;
    pMan = Abc_NtkToDar( pNtk, 0, 1 );
    if ( pMan == NULL )
        return NULL;
    if ( pMan->vFlopNums )
        Vec_IntFree( pMan->vFlopNums );
    pMan->vFlopNums = NULL;

    pTemp = pMan
    pMan = Saig_ManHaigRecord(pTemp, nIters, nSteps, fRetimingOnly, fAddBugs, fUseCnf, fVerbose );
    Aig_ManStop( pTemp );

    pNtkAig = Abc_NtkFromDarSeqSweep( pNtk, pMan );
    Aig_ManStop( pMan );
    return pNtkAig;
}
*/

/*
*Function*************************************************************

	Synopsis    [Performs random simulation.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkDarSeqSim(pNtk *Abc_Ntk_t, nFrames int32, nWords int32, TimeOut int32, fNew int32, fMiter int32, fVerbose int32, pFileSim string) int32 {
	var pMan *Aig_Man_t
	var pCex *Abc_Cex_t
	var status int32
	var RetValue int32 = -1
	var clk abctime = Abc_Clock()
	if Abc_NtkGetChoiceNum(pNtk) != 0 {
		Abc_Print(1, "Removing %d choices from the AIG.\n", Abc_NtkGetChoiceNum(pNtk))
		abc.Abc_AigCleanup(abc.Abc_NtkReadFunctionManager(pNtk).(*abc.Abc_Aig_t))
	}
	pMan = Abc_NtkToDar(pNtk, 0, 1)
	if fNew != 0 {
		var pGia *Gia_Man_t
		var Pars giapkg.Gia_ParSim_t
		var pPars *giapkg.Gia_ParSim_t = &Pars
		giapkg.Gia_ManSimSetDefaultParams(pPars)
		pPars.NWords = nWords
		pPars.NIters = nFrames
		pPars.TimeLimit = TimeOut
		pPars.FCheckMiter = fMiter
		pPars.FVerbose = fVerbose
		pGia = Gia_ManFromAig(pMan)
		if giapkg.Gia_ManSimSimulate(pGia, pPars) != 0 {
			if pGia.SequentialCounterexample() != nil {
				Abc_Print(1, "Simulation of %d frames with %d words asserted output %d in frame %d. ",
					nFrames, nWords, pGia.SequentialCounterexample().IPo, pGia.SequentialCounterexample().IFrame)
				status = Saig_ManVerifyCex(pMan, pGia.SequentialCounterexample())
				if status == 0 {
					Abc_Print(1, "Abc_NtkDarSeqSim(): Counter-example verification has FAILED.\n")
				}
			}
			Abc_NtkSetModel(pNtk, nil)
			Abc_NtkSetSeqModel(pNtk, nil)
			Abc_NtkSetSeqModel(pNtk, pGia.SequentialCounterexample())
			pGia.ClearSequentialCounterexample()
			RetValue = 0

		} else {
			Abc_Print(1, "Simulation of %d frames with %d words did not assert the outputs.    ",
				nFrames, nWords)
		}
		Gia_ManStop(pGia)

	} else { // comb/seq simulator

		var pSml *proof_fra.Fra_Sml_t
		if pFileSim != "" {
			assert(Abc_NtkLatchNum(pNtk) == 0, "Abc_NtkLatchNum(pNtk) == 0")
			pSml = proof_fra.Fra_SmlSimulateCombGiven(pMan, pFileSim, fMiter, fVerbose)

		} else if Abc_NtkLatchNum(pNtk) == 0 {
			pSml = proof_fra.Fra_SmlSimulateComb(pMan, nWords, fMiter)
		} else {
			pSml = proof_fra.Fra_SmlSimulateSeq(pMan, 0, nFrames, nWords, fMiter)
		}
		if pSml.NonconstantOutput {
			pCex = proof_fra.Fra_SmlGetCounterExample(pSml)
			if pCex != nil {
				Abc_Print(1, "Simulation of %d frame%s with %d word%s asserted output %d in frame %d. ",
					int32(pSml.Frames), pluralFrames(int32(pSml.Frames)),
					int32(pSml.WordsPerFrame), pluralFrames(int32(pSml.WordsPerFrame)),
					pCex.IPo, pCex.IFrame)
				status = Saig_ManVerifyCex(pMan, pCex)
				if status == 0 {
					Abc_Print(1, "Abc_NtkDarSeqSim(): Counter-example verification has FAILED.\n")
				}
			}
			Abc_NtkSetModel(pNtk, nil)
			Abc_NtkSetSeqModel(pNtk, nil)
			Abc_NtkSetSeqModel(pNtk, pCex)
			RetValue = 0

		} else {
			Abc_Print(1, "Simulation of %d frames with %d words did not assert the outputs.    ",
				nFrames, nWords)
		}
		proof_fra.Fra_SmlStop(pSml)
	}
	ABC_PRT("Time", Abc_Clock()-clk)
	Aig_ManStop(pMan)
	return RetValue
}

/*
*Function*************************************************************

	Synopsis    [Performs random simulation.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkDarSeqSim3(pNtk *Abc_Ntk_t, pPars *Ssw_RarPars_t) int32 {
	var pMan *Aig_Man_t
	var status int32
	var RetValue int32 = -1
	//    abctime clk = Abc_Clock();
	if Abc_NtkGetChoiceNum(pNtk) != 0 {
		Abc_Print(1, "Removing %d choices from the AIG.\n", Abc_NtkGetChoiceNum(pNtk))
		abc.Abc_AigCleanup(abc.Abc_NtkReadFunctionManager(pNtk).(*abc.Abc_Aig_t))
	}
	pMan = Abc_NtkToDar(pNtk, 0, 1)
	if proof_ssw.Ssw_RarSimulate(pMan, pPars) == 0 {
		if pMan.SequentialCounterexample() != nil {
			//            Abc_Print( 1, "Simulation of %d frames with %d words asserted output %d in frame %d. ",
			//                nFrames, nWords, pMan->pSeqModel->iPo, pMan->pSeqModel->iFrame );
			status = Saig_ManVerifyCex(pMan, pMan.SequentialCounterexample())
			if status == 0 {
				Abc_Print(1, "Abc_NtkDarSeqSim(): Counter-example verification has FAILED.\n")
			}
		}
		Abc_NtkSetModel(pNtk, nil)
		Abc_NtkSetSeqModel(pNtk, nil)
		Abc_NtkSetSeqModel(pNtk, pMan.SequentialCounterexample())
		pMan.ClearSequentialCounterexample()
		RetValue = 0

	} else {
		//        Abc_Print( 1, "Simulation of %d frames with %d words did not assert the outputs.    ",
		//            nFrames, nWords );
	}
	if abc.Abc_NtkReadSeqModelVec(pNtk) != nil {
		Vec_PtrFreeFree(abc.Abc_NtkReadSeqModelVec(pNtk))
	}
	abc.Abc_NtkSetSeqModelVec(pNtk, pMan.TakeSequentialCounterexamples())
	//    ABC_PRT( "Time", Abc_Clock() - clk );
	pNtk.Data = pMan.Data
	pMan.Data = nil
	Aig_ManStop(pMan)
	return RetValue
}

/*
*Function*************************************************************

	Synopsis    [Gives the current ABC network to AIG manager for processing.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkDarClau(pNtk *Abc_Ntk_t, nFrames int32, nPref int32, nClauses int32, nLutSize int32, nLevels int32, nCutsMax int32, nBatches int32, fStepUp int32, fBmc int32, fRefs int32, fTarget int32, fVerbose int32, fVeryVerbose int32) int32 {
	var pMan *Aig_Man_t
	if (fTarget != 0) && Abc_NtkPoNum(pNtk) != 1 {
		Abc_Print(1, "The number of outputs should be 1.\n")
		return 1
	}
	pMan = Abc_NtkToDar(pNtk, 0, 1)
	if pMan == nil {
		return 1
	}
	//    Aig_ManReduceLachesCount( pMan );
	if pMan.FlopNumbers != nil {
		pMan.FlopNumbers = nil
	}
	pMan.FlopNumbers = nil

	//    Fra_Clau( pMan, nStepsMax, fVerbose, fVeryVerbose );
	proof_fra.Fra_Claus(pMan, nFrames, nPref, nClauses, nLutSize, nLevels, nCutsMax, nBatches, fStepUp, fBmc, fRefs, fTarget, fVerbose, fVeryVerbose)
	Aig_ManStop(pMan)
	return 1
}

/*
*Function*************************************************************

	Synopsis    [Performs targe enlargement.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkDarEnlarge(pNtk *Abc_Ntk_t, nFrames int32, fVerbose int32) *Abc_Ntk_t {
	var pNtkAig *Abc_Ntk_t
	var pMan, pTemp *Aig_Man_t
	pMan = Abc_NtkToDar(pNtk, 0, 1)
	if pMan == nil {
		return nil
	}
	pTemp = pMan
	pMan = aiggraph.Aig_ManFrames(pTemp, nFrames, 0, 1, 1, 1, nil)
	Aig_ManStop(pTemp)
	pNtkAig = Abc_NtkFromDarSeqSweep(pNtk, pMan)
	Aig_ManStop(pMan)
	return pNtkAig
}

/*
*Function*************************************************************

	Synopsis    [Performs targe enlargement.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkDarTempor(pNtk *Abc_Ntk_t, nFrames int32, TimeOut int32, nConfLimit int32, fUseBmc int32, fUseTransSigs int32, fVerbose int32, fVeryVerbose int32) *Abc_Ntk_t {
	var pNtkAig *Abc_Ntk_t
	var pMan, pTemp *Aig_Man_t
	pMan = Abc_NtkToDar(pNtk, 0, 1)
	if pMan == nil {
		return nil
	}
	pTemp = aig_saig.Saig_ManTempor(pMan, nFrames, TimeOut, nConfLimit, fUseBmc, fUseTransSigs, fVerbose, fVeryVerbose)
	Aig_ManStop(pMan)
	if pTemp == nil {
		return Abc_NtkDup(pNtk)
	}
	pNtkAig = Abc_NtkFromDarSeqSweep(pNtk, pTemp)
	Aig_ManStop(pTemp)
	return pNtkAig
}

/*
*Function*************************************************************

	Synopsis    [Performs induction for property only.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkDarInduction(pNtk *Abc_Ntk_t, nTimeOut int32, nFramesMax int32, nConfMax int32, fUnique int32, fUniqueAll int32, fGetCex int32, fVerbose int32, fVeryVerbose int32) int32 {
	var pMan *Aig_Man_t
	var clkTotal abctime = Abc_Clock()
	var RetValue int32
	pMan = Abc_NtkToDar(pNtk, 0, 1)
	if pMan == nil {
		return -1
	}
	RetValue = aig_saig.Saig_ManInduction(pMan, nTimeOut, nFramesMax, nConfMax, fUnique, fUniqueAll, fGetCex, fVerbose, fVeryVerbose)
	if RetValue == 1 {
		Abc_Print(1, "Networks are equivalent.  ")
		ABC_PRT("Time", Abc_Clock()-clkTotal)

	} else if RetValue == 0 {
		Abc_Print(1, "Networks are NOT EQUIVALENT.  ")
		ABC_PRT("Time", Abc_Clock()-clkTotal)

	} else {
		Abc_Print(1, "Networks are UNDECIDED.  ")
		ABC_PRT("Time", Abc_Clock()-clkTotal)
	}
	if fGetCex != 0 {
		pNtk.TakeInductionCounterexample(pMan)
	}
	Aig_ManStop(pMan)
	return RetValue
}

/*
*Function*************************************************************

	Synopsis    [Interplates two networks.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkInterOne(pNtkOn *Abc_Ntk_t, pNtkOff *Abc_Ntk_t, fRelation int32, fVerbose int32) *Abc_Ntk_t {
	var pNtkAig *Abc_Ntk_t
	var pManOn, pManOff, pManAig *Aig_Man_t
	if Abc_NtkCoNum(pNtkOn) != 1 || Abc_NtkCoNum(pNtkOff) != 1 {
		Abc_Print(1, "Currently works only for single-output networks.\n")
		return nil
	}
	if Abc_NtkCiNum(pNtkOn) != Abc_NtkCiNum(pNtkOff) {
		Abc_Print(1, "The number of PIs should be the same.\n")
		return nil
	}
	// create internal AIGs
	pManOn = Abc_NtkToDar(pNtkOn, 0, 0)
	if pManOn == nil {
		return nil
	}
	pManOff = Abc_NtkToDar(pNtkOff, 0, 0)
	if pManOff == nil {
		return nil
	}
	// derive the interpolant
	pManAig = aiginterp.Aig_ManInter(pManOn, pManOff, fRelation, fVerbose)
	if pManAig == nil {
		Abc_Print(1, "Interpolant computation failed.\n")
		return nil
	}
	Aig_ManStop(pManOn)
	Aig_ManStop(pManOff)
	// for the relation, add an extra input
	if fRelation != 0 {
		var pObj *Abc_Obj_t
		pObj = abc.Abc_NtkCreatePi(pNtkOff)
		abc.Abc_ObjAssignName(pObj, "New", "")
	}
	// create logic network
	pNtkAig = Abc_NtkFromDar(pNtkOff, pManAig)
	Aig_ManStop(pManAig)
	return pNtkAig
}
func Gia_ManInterOne(pNtkOn *Gia_Man_t, pNtkOff *Gia_Man_t, fVerbose int32) *Gia_Man_t {
	var pNtkAig *Gia_Man_t
	var pManOn, pManOff, pManAig *Aig_Man_t
	assert(Gia_ManCiNum(pNtkOn) == Gia_ManCiNum(pNtkOff), "Gia_ManCiNum(pNtkOn)  == Gia_ManCiNum(pNtkOff)")
	assert(Gia_ManCoNum(pNtkOn) == 1, "Gia_ManCoNum(pNtkOn)  == 1")
	assert(Gia_ManCoNum(pNtkOff) == 1, "Gia_ManCoNum(pNtkOff) == 1")
	// create internal AIGs
	pManOn = Gia_ManToAigSimple(pNtkOn)
	if pManOn == nil {
		return nil
	}
	pManOff = Gia_ManToAigSimple(pNtkOff)
	if pManOff == nil {
		return nil
	}
	// derive the interpolant
	pManAig = aiginterp.Aig_ManInter(pManOn, pManOff, 0, fVerbose)
	if pManAig == nil {
		Abc_Print(1, "Interpolant computation failed.\n")
		return nil
	}
	Aig_ManStop(pManOn)
	Aig_ManStop(pManOff)
	// create logic network
	pNtkAig = Gia_ManFromAigSimple(pManAig)
	Aig_ManStop(pManAig)
	return pNtkAig
}

/*
*Function*************************************************************

	Synopsis    [Fast interpolation.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkInterFast(pNtkOn *Abc_Ntk_t, pNtkOff *Abc_Ntk_t, fVerbose int32) {
	var pManOn, pManOff *Aig_Man_t
	// create internal AIGs
	pManOn = Abc_NtkToDar(pNtkOn, 0, 0)
	if pManOn == nil {
		return
	}
	pManOff = Abc_NtkToDar(pNtkOff, 0, 0)
	if pManOff == nil {
		return
	}
	aiginterp.Aig_ManInterFast(pManOn, pManOff, fVerbose)
	Aig_ManStop(pManOn)
	Aig_ManStop(pManOff)
}

var timeCnf abctime
var timeSat abctime
var timeInt abctime

/*
*Function*************************************************************

	Synopsis    [Interplates two networks.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkInter(pNtkOn *Abc_Ntk_t, pNtkOff *Abc_Ntk_t, fRelation int32, fVerbose int32) *Abc_Ntk_t {
	var pNtkOn1, pNtkOff1, pNtkInter1, pNtkInter *Abc_Ntk_t
	var pObj *Abc_Obj_t
	var i int32 //, clk = Abc_Clock();
	if Abc_NtkCoNum(pNtkOn) != Abc_NtkCoNum(pNtkOff) {
		Abc_Print(1, "Currently works only for networks with equal number of POs.\n")
		return nil
	}
	// compute the fast interpolation time
	//    Abc_NtkInterFast( pNtkOn, pNtkOff, fVerbose );
	// consider the case of one output
	if Abc_NtkCoNum(pNtkOn) == 1 {
		return Abc_NtkInterOne(pNtkOn, pNtkOff, fRelation, fVerbose)
	}
	// start the new network
	pNtkInter = abc.Abc_NtkAlloc(abc.ABC_NTK_STRASH, abc.ABC_FUNC_AIG, 1)
	Abc_NtkSetName(pNtkInter, Extra_UtilStrsav(Abc_NtkName(pNtkOn)))
	for _, pObj := range pNtkOn.Inputs() {
		abc.Abc_NtkDupObj(pNtkInter, pObj, 1)
	}
	// process each POs separately
	timeCnf = 0
	timeSat = 0
	timeInt = 0
	for i = 0; i < Abc_NtkCoNum(pNtkOn); i++ {
		pObj = abc.Abc_NtkCo(pNtkOn, i)
		pNtkOn1 = abc.Abc_NtkCreateCone(pNtkOn, Abc_ObjFanin0(pObj), Abc_ObjName(pObj), 1)
		if abc.Abc_ObjFaninC0(pObj) != 0 {
			abc.Abc_ObjXorFaninC(Abc_NtkPo(pNtkOn1, 0), 0)
		}

		pObj = abc.Abc_NtkCo(pNtkOff, i)
		pNtkOff1 = abc.Abc_NtkCreateCone(pNtkOff, Abc_ObjFanin0(pObj), Abc_ObjName(pObj), 1)
		if abc.Abc_ObjFaninC0(pObj) != 0 {
			abc.Abc_ObjXorFaninC(Abc_NtkPo(pNtkOff1, 0), 0)
		}

		if Abc_NtkNodeNum(pNtkOn1) == 0 {
			pNtkInter1 = Abc_NtkDup(pNtkOn1)
		} else if Abc_NtkNodeNum(pNtkOff1) == 0 {
			pNtkInter1 = Abc_NtkDup(pNtkOff1)
			abc.Abc_ObjXorFaninC(Abc_NtkPo(pNtkInter1, 0), 0)

		} else {
			pNtkInter1 = Abc_NtkInterOne(pNtkOn1, pNtkOff1, 0, fVerbose)
		}
		if pNtkInter1 != nil {
			Abc_NtkAppend(pNtkInter, pNtkInter1, 1)
			Abc_NtkDelete(pNtkInter1)
		}

		Abc_NtkDelete(pNtkOn1)
		Abc_NtkDelete(pNtkOff1)
	}
	//    ABC_PRT( "CNF", timeCnf );
	//    ABC_PRT( "SAT", timeSat );
	//    ABC_PRT( "Int", timeInt );
	//    ABC_PRT( "Slow interpolation time", Abc_Clock() - clk );

	// return the network
	if abc.Abc_NtkCheck(pNtkInter) == 0 {
		Abc_Print(1, "Abc_NtkAttachBottom(): Network check has failed.\n")
	}
	return pNtkInter
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkPrintSccs(pNtk *Abc_Ntk_t, fVerbose int32) {
	//    extern Vec_Ptr_t * Aig_ManRegPartitionLinear( Aig_Man_t * pAig, int nPartSize );
	var pMan *Aig_Man_t
	pMan = Abc_NtkToDar(pNtk, 0, 1)
	if pMan == nil {
		return
	}
	aiggraph.Aig_ManComputeSccs(pMan)
	//    Aig_ManRegPartitionLinear( pMan, 1000 );
	Aig_ManStop(pMan)
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkDarPrintCone(pNtk *Abc_Ntk_t) int32 {
	var pMan *Aig_Man_t
	pMan = Abc_NtkToDar(pNtk, 0, 1)
	if pMan == nil {
		return 0
	}
	assert(aiggraph.Aig_ManRegNum(pMan) > 0, "Aig_ManRegNum(pMan) > 0")
	aig_saig.Saig_ManPrintCones(pMan)
	Aig_ManStop(pMan)
	return 1
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkBalanceExor(pNtk *Abc_Ntk_t, fUpdateLevel int32, fVerbose int32) *Abc_Ntk_t {
	var pNtkAig *Abc_Ntk_t
	var pMan, pTemp *Aig_Man_t //, * pTemp2;
	assert((Abc_NtkIsStrash(pNtk) != 0), "Abc_NtkIsStrash(pNtk)")
	// derive AIG with EXORs
	pMan = Abc_NtkToDar(pNtk, 1, 0)
	if pMan == nil {
		return nil
	}
	//    Aig_ManPrintStats( pMan );
	if fVerbose != 0 {
		opt_dar.Dar_BalancePrintStats(os.Stdout, pMan)
	}
	// perform balancing
	pTemp = opt_dar.Dar_ManBalance(pMan, (fUpdateLevel != 0))
	//    Aig_ManPrintStats( pTemp );
	// create logic network
	pNtkAig = Abc_NtkFromDar(pNtk, pTemp)
	Aig_ManStop(pTemp)
	Aig_ManStop(pMan)
	return pNtkAig
}

/*
*Function*************************************************************

	Synopsis    [Performs phase abstraction.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkPhaseAbstract(pNtk *Abc_Ntk_t, nFrames int32, nPref int32, fIgnore int32, fPrint int32, fVerbose int32) *Abc_Ntk_t {
	var vInits *Vec_Int_t
	var pNtkAig *Abc_Ntk_t
	var pMan, pTemp *Aig_Man_t
	pMan = Abc_NtkToDar(pNtk, 0, 1)
	if pMan == nil {
		return nil
	}
	vInits = Abc_NtkGetLatchValues(pNtk)
	pTemp = pMan
	pMan = aig_saig.Saig_ManPhaseAbstract(pTemp, vInits, nFrames, nPref, fIgnore, fPrint, fVerbose)
	Vec_IntFree(vInits)
	Aig_ManStop(pTemp)
	if pMan == nil {
		return nil
	}
	pNtkAig = Abc_NtkFromAigPhase(pMan)
	//    pNtkAig->pName = Extra_UtilStrsav(pNtk->pName);
	//    pNtkAig->pSpec = Extra_UtilStrsav(pNtk->pSpec);
	Aig_ManStop(pMan)
	return pNtkAig
}

/*
*Function*************************************************************

	Synopsis    [Performs phase abstraction.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkPhaseFrameNum(pNtk *Abc_Ntk_t) int32 {
	var vInits *Vec_Int_t
	var pMan *Aig_Man_t
	var nFrames int32
	pMan = Abc_NtkToDar(pNtk, 0, 1)
	if pMan == nil {
		return 1
	}
	vInits = Abc_NtkGetLatchValues(pNtk)
	nFrames = aig_saig.Saig_ManPhaseFrameNum(pMan, vInits)
	Vec_IntFree(vInits)
	Aig_ManStop(pMan)
	return nFrames
}

/*
*Function*************************************************************

	Synopsis    [Performs phase abstraction.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkDarSynchOne(pNtk *Abc_Ntk_t, nWords int32, fVerbose int32) *Abc_Ntk_t {
	var pNtkAig *Abc_Ntk_t
	var pMan, pTemp *Aig_Man_t
	pMan = Abc_NtkToDar(pNtk, 0, 1)
	if pMan == nil {
		return nil
	}
	pTemp = pMan
	pMan = aig_saig.Saig_SynchSequenceApply(pTemp, nWords, fVerbose)
	Aig_ManStop(pTemp)
	if pMan == nil {
		return nil
	}
	pNtkAig = Abc_NtkFromDar(pNtk, pMan)
	Aig_ManStop(pMan)
	return pNtkAig
}

/*
*Function*************************************************************

	Synopsis    [Performs phase abstraction.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkDarSynch(pNtk1 *Abc_Ntk_t, pNtk2 *Abc_Ntk_t, nWords int32, fVerbose int32) *Abc_Ntk_t {
	var pNtkAig *Abc_Ntk_t
	var pMan1, pMan2, pMan *Aig_Man_t
	pMan1 = Abc_NtkToDar(pNtk1, 0, 1)
	if pMan1 == nil {
		return nil
	}
	pMan2 = Abc_NtkToDar(pNtk2, 0, 1)
	if pMan2 == nil {
		Aig_ManStop(pMan1)
		return nil
	}
	pMan = aig_saig.Saig_Synchronize(pMan1, pMan2, nWords, fVerbose)
	Aig_ManStop(pMan1)
	Aig_ManStop(pMan2)
	if pMan == nil {
		return nil
	}
	pNtkAig = Abc_NtkFromAigPhase(pMan)
	//    pNtkAig->pName = Extra_UtilStrsav("miter");
	//    pNtkAig->pSpec = NULL;
	Aig_ManStop(pMan)
	return pNtkAig
}

/*
*Function*************************************************************

	Synopsis    [Performs phase abstraction.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkDarClockGate(pNtk *Abc_Ntk_t, pCare *Abc_Ntk_t, pPars *Cgt_Par_t) *Abc_Ntk_t {
	var pNtkAig *Abc_Ntk_t
	var pMan1 *Aig_Man_t
	var pMan2 *Aig_Man_t = nil
	var pMan *Aig_Man_t
	pMan1 = Abc_NtkToDar(pNtk, 0, 1)
	if pMan1 == nil {
		return nil
	}
	if pCare != nil {
		pMan2 = Abc_NtkToDar(pCare, 0, 0)
		if pMan2 == nil {
			Aig_ManStop(pMan1)
			return nil
		}
	}
	var err error
	pMan, err = opt_cgt.Cgt_ClockGating(pMan1, pMan2, pPars)
	Aig_ManStop(pMan1)
	if pMan2 != nil {
		Aig_ManStop(pMan2)
	}
	if err != nil {
		Abc_Print(1, "%s\n", err)
		if pMan != nil {
			Aig_ManStop(pMan)
		}
		return nil
	}
	if pMan == nil {
		return nil
	}
	pNtkAig = Abc_NtkFromDar(pNtk, pMan)
	Aig_ManStop(pMan)
	return pNtkAig
}

/*
*Function*************************************************************

	Synopsis    [Performs phase abstraction.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkDarExtWin(pNtk *Abc_Ntk_t, nObjId int32, nDist int32, fVerbose int32) *Abc_Ntk_t {
	var pNtkAig *Abc_Ntk_t
	var pMan1, pMan *Aig_Man_t
	var pObj aiggraph.Edge
	pMan1 = Abc_NtkToDar(pNtk, 0, 1)
	if pMan1 == nil {
		return nil
	}
	if nObjId == -1 {
		pObj = aig_saig.Saig_ManFindPivot(pMan1)
		Abc_Print(1, "Selected object %d as a window pivot.\n", pObj.NodeID())

	} else {
		if nObjId >= aiggraph.Aig_ManObjNumMax(pMan1) {
			Aig_ManStop(pMan1)
			Abc_Print(1, "The ID is too large.\n")
			return nil
		}
		pObj = aiggraph.Aig_ManObj(pMan1, nObjId)
		if pObj == aiggraph.InvalidEdge {
			Aig_ManStop(pMan1)
			Abc_Print(1, "Object with ID %d does not exist.\n", nObjId)
			return nil
		}
		if !aig_saig.Saig_ObjIsLo(pMan1, pObj) && !aiggraph.Aig_ObjIsNode(pMan1, pObj) {
			Aig_ManStop(pMan1)
			Abc_Print(1, "Object with ID %d is not a node or reg output.\n", nObjId)
			return nil
		}
	}
	pMan = aig_saig.Saig_ManWindowExtract(pMan1, pObj, nDist)
	Aig_ManStop(pMan1)
	if pMan == nil {
		return nil
	}
	pNtkAig = Abc_NtkFromAigPhase(pMan)
	Abc_NtkSetName(pNtkAig, Extra_UtilStrsav(Abc_NtkName(pNtk)))
	abc.Abc_NtkSetSpec(pNtkAig, Extra_UtilStrsav(abc.Abc_NtkSpec(pNtk)))
	Aig_ManStop(pMan)
	return pNtkAig
}

/*
*Function*************************************************************

	Synopsis    [Performs phase abstraction.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkDarInsWin(pNtk *Abc_Ntk_t, pCare *Abc_Ntk_t, nObjId int32, nDist int32, fVerbose int32) *Abc_Ntk_t {
	var pNtkAig *Abc_Ntk_t
	var pMan1 *Aig_Man_t
	var pMan2 *Aig_Man_t = nil
	var pMan *Aig_Man_t
	var pObj aiggraph.Edge
	pMan1 = Abc_NtkToDar(pNtk, 0, 1)
	if pMan1 == nil {
		return nil
	}
	if nObjId == -1 {
		pObj = aig_saig.Saig_ManFindPivot(pMan1)
		Abc_Print(1, "Selected object %d as a window pivot.\n", pObj.NodeID())

	} else {
		if nObjId >= aiggraph.Aig_ManObjNumMax(pMan1) {
			Aig_ManStop(pMan1)
			Abc_Print(1, "The ID is too large.\n")
			return nil
		}
		pObj = aiggraph.Aig_ManObj(pMan1, nObjId)
		if pObj == aiggraph.InvalidEdge {
			Aig_ManStop(pMan1)
			Abc_Print(1, "Object with ID %d does not exist.\n", nObjId)
			return nil
		}
		if !aig_saig.Saig_ObjIsLo(pMan1, pObj) && !aiggraph.Aig_ObjIsNode(pMan1, pObj) {
			Aig_ManStop(pMan1)
			Abc_Print(1, "Object with ID %d is not a node or reg output.\n", nObjId)
			return nil
		}
	}
	if pCare != nil {
		pMan2 = Abc_NtkToDar(pCare, 0, 0)
		if pMan2 == nil {
			Aig_ManStop(pMan1)
			return nil
		}
	}
	pMan = aig_saig.Saig_ManWindowInsert(pMan1, pObj, nDist, pMan2)
	Aig_ManStop(pMan1)
	if pMan2 != nil {
		Aig_ManStop(pMan2)
	}
	if pMan == nil {
		return nil
	}
	pNtkAig = Abc_NtkFromDarSeqSweep(pNtk, pMan)
	Aig_ManStop(pMan)
	return pNtkAig
}

/*
*Function*************************************************************

	Synopsis    [Performs phase abstraction.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkDarFrames(pNtk *Abc_Ntk_t, nPrefix int32, nFrames int32, fInit int32, fVerbose int32) *Abc_Ntk_t {
	var pNtkAig *Abc_Ntk_t
	var pMan, pTemp *Aig_Man_t
	pMan = Abc_NtkToDar(pNtk, 0, 1)
	if pMan == nil {
		return nil
	}
	pTemp = pMan
	pMan = proof_fra.Saig_ManTimeframeSimplify(pTemp, nPrefix, nFrames, fInit, fVerbose)
	Aig_ManStop(pTemp)
	if pMan == nil {
		return nil
	}
	pNtkAig = Abc_NtkFromAigPhase(pMan)
	Abc_NtkSetName(pNtkAig, Extra_UtilStrsav(Abc_NtkName(pNtk)))
	abc.Abc_NtkSetSpec(pNtkAig, Extra_UtilStrsav(abc.Abc_NtkSpec(pNtk)))
	Aig_ManStop(pMan)
	return pNtkAig
}

/*
*Function*************************************************************

	Synopsis    [Performs phase abstraction.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkDarCleanupAig(pNtk *Abc_Ntk_t, fCleanupPis int32, fCleanupPos int32, fVerbose int32) *Abc_Ntk_t {
	var pNtkAig *Abc_Ntk_t
	var pMan *Aig_Man_t
	pMan = Abc_NtkToDar(pNtk, 0, 1)
	if pMan == nil {
		return nil
	}
	if fCleanupPis != 0 {
		var Temp int32 = aiggraph.Aig_ManCiCleanup(pMan)
		if fVerbose != 0 {
			Abc_Print(1, "Cleanup removed %d primary inputs without fanout.\n", Temp)
		}
	}
	if fCleanupPos != 0 {
		var Temp int32 = aiggraph.Aig_ManCoCleanup(pMan)
		if fVerbose != 0 {
			Abc_Print(1, "Cleanup removed %d primary outputs driven by const-0.\n", Temp)
		}
	}
	pNtkAig = Abc_NtkFromAigPhase(pMan)
	Abc_NtkSetName(pNtkAig, Extra_UtilStrsav(Abc_NtkName(pNtk)))
	abc.Abc_NtkSetSpec(pNtkAig, Extra_UtilStrsav(abc.Abc_NtkSpec(pNtk)))
	Aig_ManStop(pMan)
	return pNtkAig
}

/*
*Function*************************************************************

	Synopsis    [Performs BDD-based reachability analysis.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkDarReach(pNtk *Abc_Ntk_t, pPars *Saig_ParBbr_t) int32 {
	var pMan *Aig_Man_t
	var RetValue int32
	pMan = Abc_NtkToDar(pNtk, 0, 1)
	if pMan == nil {
		return -1
	}
	RetValue = bdd_bbr.Aig_ManVerifyUsingBdds(pMan, pPars)
	Abc_NtkSetModel(pNtk, nil)
	Abc_NtkSetSeqModel(pNtk, nil)
	Abc_NtkSetSeqModel(pNtk, pMan.SequentialCounterexample())
	pMan.ClearSequentialCounterexample()
	Aig_ManStop(pMan)
	return RetValue
}

//ABC_NAMESPACE_IMPL_END

//#include "map/amap/amap.h"
//#include "map/mio/mio.h"

//ABC_NAMESPACE_IMPL_START

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func pluralFrames(n int32) string {
	if n == 1 {
		return ""
	}
	return "s"
}

func Amap_ManProduceNetwork(pNtk *Abc_Ntk_t, vMapping *Vec_Ptr_t) *Abc_Ntk_t {
	//    extern void * Abc_FrameReadLibGen();
	var pLib *Mio_Library_t = base_main.Abc_FrameReadLibGen().(*Mio_Library_t)
	var pRes *Amap_Out_t
	var vNodesNew *Vec_Ptr_t
	var pNtkNew *Abc_Ntk_t
	var pNodeNew, pFaninNew *Abc_Obj_t
	var k, iPis, iPos int32
	// make sure gates exist in the current library
	for _, item := range giapkg.Vec_PtrArray(vMapping) {
		pRes = item.(*Amap_Out_t)
		if amap.Amap_OutName(pRes) != "" && map_mio.Mio_LibraryReadGateByName(pLib, amap.Amap_OutName(pRes), "") == nil {
			Abc_Print(1, "Current library does not contain gate \"%s\".\n", amap.Amap_OutName(pRes))
			return nil
		}
	}
	// create the network
	pNtkNew = abc.Abc_NtkStartFrom(pNtk, abc.ABC_NTK_LOGIC, abc.ABC_FUNC_MAP)
	abc.Abc_NtkSetFunctionManager(pNtkNew, pLib)
	iPis, iPos = 0, 0
	vNodesNew = Vec_PtrAlloc(Vec_PtrSize(vMapping))
	for _, item := range giapkg.Vec_PtrArray(vMapping) {
		pRes = item.(*Amap_Out_t)
		if amap.Amap_OutType(pRes) == -1 {
			pNodeNew = abc.Abc_NtkCi(pNtkNew, iPis)
			iPis++
		} else if amap.Amap_OutType(pRes) == 1 {
			pNodeNew = abc.Abc_NtkCo(pNtkNew, iPos)
			iPos++
		} else {
			pNodeNew = abc.Abc_NtkCreateNode(pNtkNew)
			abc.Abc_ObjSetData(pNodeNew, map_mio.Mio_LibraryReadGateByName(pLib, amap.Amap_OutName(pRes), ""))
		}
		for k = 0; k < int32(amap.Amap_OutFanCount(pRes)); k++ {
			pFaninNew = Vec_PtrEntry(vNodesNew, amap.Amap_OutFan(pRes, int16(k))).(*Abc_Obj_t)
			Abc_ObjAddFanin(pNodeNew, pFaninNew)
		}
		Vec_PtrPush(vNodesNew, pNodeNew)
	}
	Vec_PtrFree(vNodesNew)
	assert(iPis == Abc_NtkCiNum(pNtkNew), "iPis == Abc_NtkCiNum(pNtkNew)")
	assert(iPos == Abc_NtkCoNum(pNtkNew), "iPos == Abc_NtkCoNum(pNtkNew)")
	// decouple the PO driver nodes to reduce the number of levels
	abc.Abc_NtkLogicMakeSimpleCos(pNtkNew, 0)
	//    if ( nDupGates && Map_ManReadVerbose(pMan) )
	//        Abc_Print( 1, "Duplicated %d gates to decouple the CO drivers.\n", nDupGates );
	return pNtkNew
}

/*
*Function*************************************************************

	Synopsis    [Gives the current ABC network to AIG manager for processing.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkDarAmap(pNtk *Abc_Ntk_t, pPars *Amap_Par_t) *Abc_Ntk_t {
	var vMapping *Vec_Ptr_t
	var pNtkAig *Abc_Ntk_t = nil
	var pMan *Aig_Man_t
	var pMem *aiggraph.Aig_MmFlex_t

	assert((Abc_NtkIsStrash(pNtk) != 0), "Abc_NtkIsStrash(pNtk)")
	// convert to the AIG manager
	pMan = Abc_NtkToDarChoices(pNtk)
	if pMan == nil {
		return nil
	}

	// perform computation
	vMapping = amap.Amap_ManTest(pMan, pPars)
	Aig_ManStop(pMan)
	if vMapping == nil {
		return nil
	}
	pMem = Vec_PtrPop(vMapping).(*aiggraph.Aig_MmFlex_t)
	pNtkAig = Amap_ManProduceNetwork(pNtk, vMapping)
	aiggraph.Aig_MmFlexStop(pMem, 0)
	Vec_PtrFree(vMapping)

	// make sure everything is okay
	if pNtkAig != nil && abc.Abc_NtkCheck(pNtkAig) == 0 {
		Abc_Print(1, "Abc_NtkDar: The network check has failed.\n")
		Abc_NtkDelete(pNtkAig)
		return nil
	}
	return pNtkAig
}

/*
*Function*************************************************************

	Synopsis    [Performs BDD-based reachability analysis.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkDarConstr(pNtk *Abc_Ntk_t, nFrames int32, nConfs int32, nProps int32, fStruct int32, fOldAlgo int32, fVerbose int32) {
	var pMan *Aig_Man_t //, * pMan2;//, * pTemp;
	assert((Abc_NtkIsStrash(pNtk) != 0), "Abc_NtkIsStrash(pNtk)")
	pMan = Abc_NtkToDar(pNtk, 0, 1)
	if pMan == nil {
		return
	}
	if fStruct != 0 {
		Saig_ManDetectConstrTest(pMan)
	} else {
		aig_saig.Saig_ManDetectConstrFuncTest(pMan, nFrames, nConfs, nProps, fOldAlgo, fVerbose)
	}
	Aig_ManStop(pMan)
}

/*
*Function*************************************************************

	Synopsis    [Performs BDD-based reachability analysis.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkDarOutdec(pNtk *Abc_Ntk_t, nLits int32, fVerbose int32) *Abc_Ntk_t {
	var pNtkAig *Abc_Ntk_t
	var pMan, pTemp *Aig_Man_t
	assert((Abc_NtkIsStrash(pNtk) != 0), "Abc_NtkIsStrash(pNtk)")
	pMan = Abc_NtkToDar(pNtk, 0, 1)
	if pMan == nil {
		return nil
	}
	pTemp = pMan
	pMan = aig_saig.Saig_ManDecPropertyOutput(pTemp, nLits, fVerbose)
	Aig_ManStop(pTemp)
	if pMan == nil {
		return nil
	}
	pNtkAig = Abc_NtkFromAigPhase(pMan)
	Abc_NtkSetName(pNtkAig, Extra_UtilStrsav(pMan.Name))
	abc.Abc_NtkSetSpec(pNtkAig, Extra_UtilStrsav(pMan.Spec))
	Aig_ManStop(pMan)
	return pNtkAig
}

/*
*Function*************************************************************

	Synopsis    [Performs BDD-based reachability analysis.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkDarUnfold(pNtk *Abc_Ntk_t, nFrames int32, nConfs int32, nProps int32, fStruct int32, fOldAlgo int32, fVerbose int32) *Abc_Ntk_t {
	var pNtkAig *Abc_Ntk_t
	var pMan, pTemp *Aig_Man_t
	assert((Abc_NtkIsStrash(pNtk) != 0), "Abc_NtkIsStrash(pNtk)")
	pMan = Abc_NtkToDar(pNtk, 0, 1)
	if pMan == nil {
		return nil
	}
	if fStruct != 0 {
		pTemp = pMan
		pMan = aig_saig.Saig_ManDupUnfoldConstrs(pTemp)
	} else {
		pTemp = pMan
		pMan = aig_saig.Saig_ManDupUnfoldConstrsFunc(pTemp, nFrames, nConfs, nProps, fOldAlgo, fVerbose)
	}
	Aig_ManStop(pTemp)
	if pMan == nil {
		return nil
	}
	pNtkAig = Abc_NtkFromAigPhase(pMan)
	Abc_NtkSetName(pNtkAig, Extra_UtilStrsav(pMan.Name))
	abc.Abc_NtkSetSpec(pNtkAig, Extra_UtilStrsav(pMan.Spec))
	Aig_ManStop(pMan)
	return pNtkAig
}

/*
*Function*************************************************************

	Synopsis    [Performs BDD-based reachability analysis.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkDarFold(pNtk *Abc_Ntk_t, fCompl int32, fVerbose int32) *Abc_Ntk_t {
	var pNtkAig *Abc_Ntk_t
	var pMan, pTemp *Aig_Man_t
	assert((Abc_NtkIsStrash(pNtk) != 0), "Abc_NtkIsStrash(pNtk)")
	pMan = Abc_NtkToDar(pNtk, 0, 1)
	if pMan == nil {
		return nil
	}
	pTemp = pMan
	pMan = aig_saig.Saig_ManDupFoldConstrsFunc(pTemp, fCompl, fVerbose)
	Aig_ManStop(pTemp)
	pNtkAig = Abc_NtkFromAigPhase(pMan)
	Abc_NtkSetName(pNtkAig, Extra_UtilStrsav(pMan.Name))
	abc.Abc_NtkSetSpec(pNtkAig, Extra_UtilStrsav(pMan.Spec))
	Aig_ManStop(pMan)
	return pNtkAig
}

/*
*Function*************************************************************

	Synopsis    [Performs BDD-based reachability analysis.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkDarConstrProfile(pNtk *Abc_Ntk_t, fVerbose int32) {
	var pMan *Aig_Man_t
	//    Vec_Int_t * vProbOne;
	//    Aig_Obj_t * pObj;
	//    int i, Entry;
	assert((Abc_NtkIsStrash(pNtk) != 0), "Abc_NtkIsStrash(pNtk)")
	assert((Abc_NtkConstrNum(pNtk) != 0), "Abc_NtkConstrNum(pNtk)")
	pMan = Abc_NtkToDar(pNtk, 0, 1)
	if pMan == nil {
		return
	}
	// value in the init state
	//    Abc_AigSetNodePhases( pNtk );
	/*
	   // derive probabilities
	   vProbOne = Saig_ManComputeSwitchProbs( pMan, 48, 16, 1 );
	   // iterate over the constraint outputs
	   Saig_ManForEachPo( pMan, pObj, i )
	   {
	       Entry = Vec_IntEntry( vProbOne, Aig_ObjId(pObj) );
	       if ( i < Saig_ManPoNum(pMan) - Saig_ManConstrNum(pMan) )
	           Abc_Print( 1, "Primary output :  ", i );
	       else
	           Abc_Print( 1, "Constraint %3d :  ", i-(Saig_ManPoNum(pMan) - Saig_ManConstrNum(pMan)) );
	       Abc_Print( 1, "ProbOne = %f  ", Abc_Int2Float(Entry) );
	       Abc_Print( 1, "AllZeroValue = %d ", Aig_ObjPhase(pObj) );
	       Abc_Print( 1, "\n" );
	   }
	*/
	// double-check
	aig_saig.Ssw_ManProfileConstraints(pMan, 16, 64, 1)
	Abc_Print(1, "TwoFrameSatValue = %d.\n", proof_ssw.Ssw_ManSetConstrPhases(pMan, 2, nil))
	// clean up
	//    Vec_IntFree( vProbOne );
	Aig_ManStop(pMan)
}

/*
*Function*************************************************************

	Synopsis    [Performs BDD-based reachability analysis.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkDarTest(pNtk *Abc_Ntk_t, Num int32) {
	//    extern void Saig_ManDetectConstr( Aig_Man_t * p );
	//    extern void Saig_ManDetectConstrFuncTest( Aig_Man_t * p );
	//    extern void Saig_ManFoldConstrTest( Aig_Man_t * pAig );

	//    extern void Fsim_ManTest( Aig_Man_t * pAig );
	//    Vec_Int_t * vPairs;
	var pMan *Aig_Man_t //, * pMan2;//, * pTemp;
	assert((Abc_NtkIsStrash(pNtk) != 0), "Abc_NtkIsStrash(pNtk)")
	pMan = Abc_NtkToDar(pNtk, 0, 1)
	if pMan == nil {
		return
	}
	/*
	   Aig_ManSetRegNum( pMan, pMan->nRegs );
	   Aig_ManPrintStats( pMan );
	   Saig_ManDumpBlif( pMan, "_temp_.blif" );
	   Aig_ManStop( pMan );
	   pMan = Saig_ManReadBlif( "_temp_.blif" );
	   Aig_ManPrintStats( pMan );
	*/
	/*
	   Aig_ManSetRegNum( pMan, pMan->nRegs );
	   pTemp = Ssw_SignalCorrespondeceTestPairs( pMan );
	   Aig_ManStop( pTemp );
	*/

	/*
	   //    Ssw_SecSpecialMiter( pMan, NULL, 2, 1 );
	       pMan2 = Aig_ManDupSimple(pMan);
	       vPairs = Saig_StrSimPerformMatching( pMan, pMan2, 0, 1, NULL );
	       Vec_IntFree( vPairs );
	       Aig_ManStop( pMan );
	       Aig_ManStop( pMan2 );
	*/
	//    Ioa_WriteAigerBufferTest( pMan, "test.aig", 0, 0 );
	//    Saig_ManFoldConstrTest( pMan );
	{
		//    Saig_ManBmcSectionsTest( pMan );
		//    Saig_ManBmcTerSimTest( pMan );
		//    Saig_ManBmcSupergateTest( pMan );
		//    Saig_ManBmcMappingTest( pMan );
	}

	{
		//        void Pdr_ManEquivClasses( Aig_Man_t * pMan );
		//        Pdr_ManEquivClasses( pMan );
	}

	//    Llb_ManComputeDomsTest( pMan, Num );
	{
		//        Llb_BddStructAnalysis( pMan );
		//        Llb_ManMinCutTest( pMan, Num );
		//        Llb_NonlinExperiment( pMan, Num );
	}

	//    Saig_MvManSimulate( pMan, 1 );
	//    Saig_ManDetectConstr( pMan );
	//    Saig_ManDetectConstrFuncTest( pMan );

	//    Fsim_ManTest( pMan );
	Aig_ManStop(pMan)

}

/*
*Function*************************************************************

	Synopsis    [Performs BDD-based reachability analysis.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkDarTestNtk(pNtk *Abc_Ntk_t) *Abc_Ntk_t {
	//    extern Aig_Man_t * Saig_ManDualRail( Aig_Man_t * p, int fMiter );

	/*

	   Abc_Ntk_t * pNtkAig;
	   Aig_Man_t * pMan, * pTemp;
	   assert( Abc_NtkIsStrash(pNtk) );
	   pMan = Abc_NtkToDar( pNtk, 0, 1 );
	   if ( pMan == NULL )
	       return NULL;

	   Aig_ManSetRegNum( pMan, pMan->nRegs );
	   pMan = Ssw_SignalCorrespondeceTestPairs( pTemp = pMan );
	   Aig_ManStop( pTemp );
	   if ( pMan == NULL )
	       return NULL;

	   pNtkAig = Abc_NtkFromAigPhase( pMan );
	   pNtkAig->pName = Extra_UtilStrsav(pNtk->pName);
	   pNtkAig->pSpec = Extra_UtilStrsav(pNtk->pSpec);
	   Aig_ManStop( pMan );
	   return pNtkAig;
	*/
	var pNtkAig *Abc_Ntk_t
	var pMan *Aig_Man_t //, * pTemp;
	assert((Abc_NtkIsStrash(pNtk) != 0), "Abc_NtkIsStrash(pNtk)")
	pMan = Abc_NtkToDar(pNtk, 0, 1)
	if pMan == nil {
		return nil
	}

	/*
	   Aig_ManSetRegNum( pMan, pMan->nRegs );
	   pTemp = pMan
	   pMan = Saig_ManDualRail(pTemp, 1 );
	   Aig_ManStop( pTemp );
	   if ( pMan == NULL )
	       return NULL;

	   pNtkAig = Abc_NtkFromAigPhase( pMan );
	   pNtkAig->pName = Extra_UtilStrsav(pNtk->pName);
	   pNtkAig->pSpec = Extra_UtilStrsav(pNtk->pSpec);
	   Aig_ManStop( pMan );
	*/

	pNtkAig = Abc_NtkFromDar(pNtk, pMan)
	Aig_ManStop(pMan)

	return pNtkAig

}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//#include "abcDarUnfold2.c"
//ABC_NAMESPACE_IMPL_END))))))))))))))))))))))))))))))

func abcAigModel(g *Aig_Man_t) []int32 {
	data := aiggraph.Aig_ManReleaseData(g)
	if data == nil {
		return nil
	}
	return data.([]int32)
}

func Saig_ManDetectConstrTest(p *Aig_Man_t) int32 {
	_, _, status := aig_saig.Saig_ManDetectConstr(p, 0)
	return status
}
