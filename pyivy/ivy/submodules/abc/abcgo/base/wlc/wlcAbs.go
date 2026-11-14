package wlc

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
)

/**CFile****************************************************************

  FileName    [wlcAbs.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Verilog parser.]

  Synopsis    [Abstraction for word-level networks.]

  Author      [Yen-Sheng Ho, Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - August 22, 2014.]

  Revision    [$Id: wlcAbs.c,v 1.00 2014/09/12 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "wlc.h"
//#include "proof/pdr/pdr.h"
//#include "proof/pdr/pdrInt.h"
//#include "proof/ssw/ssw.h"
//#include "aig/gia/giaAig.h"
//#include "sat/bmc/bmc.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// extern Vec_Vec_t *   IPdr_ManSaveClauses( Pdr_Man_t * p, int fDropLast );

// extern int           IPdr_ManRestoreClauses( Pdr_Man_t * p, Vec_Vec_t * vClauses, Vec_Int_t * vMap );

// extern int           IPdr_ManRebuildClauses( Pdr_Man_t * p, Vec_Vec_t * vClauses );

// extern int           IPdr_ManSolveInt( Pdr_Man_t * p, int fCheckClauses, int fPushClauses );

// extern int           IPdr_ManCheckCombUnsat( Pdr_Man_t * p );

// extern int           IPdr_ManReduceClauses( Pdr_Man_t * p, Vec_Vec_t * vClauses );

// extern void          IPdr_ManPrintClauses( Vec_Vec_t * vClauses, int kStart, int nRegs );

// extern void          Wla_ManJoinThread( Wla_Man_t * pWla, int RunId );

// extern void          Wla_ManConcurrentBmc3( Wla_Man_t * pWla, Aig_Man_t * pAig, Abc_Cex_t ** ppCex );

// extern int           Wla_CallBackToStop( int RunId );

// extern int           Wla_GetGlobalRunId();

type Int_Pair_t struct {
	first  int32
	second int32
}

func IntPairPtrCompare(a, b *any) int {
	left, right := (*a).(*Int_Pair_t).second, (*b).(*Int_Pair_t).second
	if left < right {
		return -1
	}
	if left > right {
		return 1
	}
	return 0
}

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

func Wlc_NtkPrintNtk(p *Wlc_Ntk_t) {
	var i int32
	var pObj *Wlc_Obj_t

	Abc_Print(1, "PIs:")
	for i = 0; i < Wlc_NtkPiNum(p); i++ {
		pObj = Wlc_NtkPi(p, i)
		Abc_Print(1, " %s", Wlc_ObjName(p, Wlc_ObjId(p, pObj)))
	}
	Abc_Print(1, "\n\n")

	Abc_Print(1, "POs:")
	for i = 0; i < Wlc_NtkPoNum(p); i++ {
		pObj = Wlc_NtkPo(p, i)
		Abc_Print(1, " %s", Wlc_ObjName(p, Wlc_ObjId(p, pObj)))
	}
	Abc_Print(1, "\n\n")

	Abc_Print(1, "FO(Fi)s:")
	for i = 0; i < Wlc_NtkCiNum(p); i++ {
		pObj = Wlc_NtkCi(p, i)
		if Wlc_ObjIsPi(pObj) == 0 {
			Abc_Print(1, " %s(%s)", Wlc_ObjName(p, Wlc_ObjId(p, pObj)), Wlc_ObjName(p, Wlc_ObjId(p, Wlc_ObjFo2Fi(p, pObj))))
		}
	}
	Abc_Print(1, "\n\n")

	Abc_Print(1, "Objs:\n")
	for i = 1; i < Wlc_NtkObjNumMax(p); i++ {
		pObj = Wlc_NtkObj(p, i)
		if Wlc_ObjIsCi(pObj) == 0 {
			Wlc_NtkPrintNode(p, pObj)
		}
	}
}

func Wlc_NtkAbsGetSupp_rec(p *Wlc_Ntk_t, pObj *Wlc_Obj_t, vCiMarks *Vec_Bit_t, vSuppRefs *Vec_Int_t, vSuppList *Vec_Int_t) {
	var i, iFanin, iObj int32
	if pObj.Mark != 0 { // visited
		return
	}
	pObj.Mark = 1
	iObj = Wlc_ObjId(p, pObj)
	if Vec_BitEntry(vCiMarks, iObj) {
		if vSuppRefs != nil {
			Vec_IntAddToEntry(vSuppRefs, iObj, 1)
		}
		if vSuppList != nil {
			Vec_IntPush(vSuppList, iObj)
		}
		return
	}
	for i = 0; i < Wlc_ObjFaninNum(pObj); i++ {
		iFanin = Wlc_ObjFaninId(pObj, i)
		Wlc_NtkAbsGetSupp_rec(p, Wlc_NtkObj(p, iFanin), vCiMarks, vSuppRefs, vSuppList)
	}
}

func Wlc_NtkAbsGetSupp(p *Wlc_Ntk_t, pObj *Wlc_Obj_t, vCiMarks *Vec_Bit_t, vSuppRefs *Vec_Int_t, vSuppList *Vec_Int_t) {
	assert((vSuppRefs != nil) || (vSuppList != nil), "vSuppRefs || vSuppList")
	Wlc_NtkCleanMarks(p)

	Wlc_NtkAbsGetSupp_rec(p, pObj, vCiMarks, vSuppRefs, vSuppList)
}

func Wlc_NtkNumPiBits(pNtk *Wlc_Ntk_t) int32 {
	var num int32
	var i int32
	var pObj *Wlc_Obj_t
	for i = 0; i < Wlc_NtkPiNum(pNtk); i++ {
		pObj = Wlc_NtkPi(pNtk, i)
		num += Wlc_ObjRange(pObj)
	}
	return num
}

func Wlc_NtkAbsAnalyzeRefine(p *Wlc_Ntk_t, vBlacks *Vec_Int_t, vUnmark *Vec_Bit_t, nDisj *int32, nNDisj *int32) {
	var vCurCis, vCandCis *Vec_Bit_t
	var vSuppList *Vec_Int_t
	var vDeltaB *Vec_Int_t
	var vSuppRefs *Vec_Int_t
	var i, Entry int32
	var pObj *Wlc_Obj_t

	vCurCis = Vec_BitStart(Wlc_NtkObjNumMax(p))
	vCandCis = Vec_BitStart(Wlc_NtkObjNumMax(p))
	vDeltaB = Vec_IntAlloc(Vec_IntSize(vBlacks))
	vSuppList = Vec_IntAlloc(Wlc_NtkCiNum(p) + Vec_IntSize(vBlacks))
	vSuppRefs = Vec_IntAlloc(Wlc_NtkObjNumMax(p))

	Vec_IntFill(vSuppRefs, Wlc_NtkObjNumMax(p), 0)

	for i = 0; i < Wlc_NtkCiNum(p); i++ {
		pObj = Wlc_NtkCi(p, i)
		Vec_BitWriteEntry(vCurCis, Wlc_ObjId(p, pObj), 1)
		Vec_BitWriteEntry(vCandCis, Wlc_ObjId(p, pObj), 1)
	}

	for i = 0; i < Vec_IntSize(vBlacks); i++ {
		Entry = Vec_IntEntry(vBlacks, i)
		Vec_BitWriteEntry(vCurCis, Entry, 1)
		if !Vec_BitEntry(vUnmark, Entry) {
			Vec_BitWriteEntry(vCandCis, Entry, 1)
		} else {
			Vec_IntPush(vDeltaB, Entry)
		}
	}
	assert(Vec_IntSize(vDeltaB) != 0, "Vec_IntSize( vDeltaB )")

	for i = 0; i < Wlc_NtkCoNum(p); i++ {
		pObj = Wlc_NtkCo(p, i)
		Wlc_NtkAbsGetSupp(p, pObj, vCurCis, vSuppRefs, nil)
	}

	/*
	   Abc_Print( 1, "SuppCurrentAbs =" );
	   Wlc_NtkForEachObj( p, pObj, i )
	       if ( Vec_IntEntry(vSuppRefs, i) > 0 )
	           Abc_Print( 1, " %d", i );
	   Abc_Print( 1, "\n" );
	*/

	for i = 0; i < Vec_IntSize(vDeltaB); i++ {
		Entry = Vec_IntEntry(vDeltaB, i)
		Wlc_NtkAbsGetSupp(p, Wlc_NtkObj(p, Entry), vCandCis, vSuppRefs, nil)
	}

	for i = 0; i < Vec_IntSize(vDeltaB); i++ {
		Entry = Vec_IntEntry(vDeltaB, i)
		var iSupp, ii int32
		var fDisjoint int32 = 1

		Vec_IntClear(vSuppList)
		Wlc_NtkAbsGetSupp(p, Wlc_NtkObj(p, Entry), vCandCis, nil, vSuppList)

		//Abc_Print( 1, "SuppCandAbs =" );
		for ii = 0; ii < Vec_IntSize(vSuppList); ii++ {
			iSupp = Vec_IntEntry(vSuppList, ii)
			//Abc_Print( 1, " %d(%d)", iSupp, Vec_IntEntry( vSuppRefs, iSupp ) );
			if Vec_IntEntry(vSuppRefs, iSupp) >= 2 {
				fDisjoint = 0
				break
			}
		}
		//Abc_Print( 1, "\n" );

		if fDisjoint != 0 {
			//Abc_Print( 1, "PPI[%d] is disjoint.\n", Entry );
			(*nDisj)++

		} else {
			//Abc_Print( 1, "PPI[%d] is non-disjoint.\n", Entry );
			(*nNDisj)++
		}
	}

	Vec_BitFree(vCurCis)
	Vec_BitFree(vCandCis)
	Vec_IntFree(vDeltaB)
	Vec_IntFree(vSuppList)
	Vec_IntFree(vSuppRefs)
}

func Wlc_NtkGetCoreSels(pFrames *Gia_Man_t, nFrames int32, first_sel_pi int32, num_sel_pis int32, vMark *Vec_Bit_t, nConfLimit int32, pPars *Wlc_Par_t, fSetPO int32, RunId int32) *Vec_Int_t {
	graph := Gia_ManToAigSimple(pFrames)
	defer Aig_ManStop(graph)
	data := Cnf_Derive(graph, Aig_ManCoNum(graph))
	defer Cnf_ManFree()
	solver := bsat.NewSolver()
	defer solver.Close()
	solver.SetNumVars(int32(data.Variables))
	if RunId >= 0 {
		solver.SetStopFunc(RunId, Wla_CallBackToStop)
	}
	for _, clause := range data.Clauses {
		lits := make([]bsat.Literal, len(clause))
		for j, lit := range clause {
			lits[j] = bsat.Literal(lit)
		}
		if !solver.AddClause(lits...) {
			assert(false, "inconsistent CNF while loading")
		}
	}
	outputs := graph.Outputs()
	poLits := make([]bsat.Literal, len(outputs))
	for i, output := range outputs {
		v := data.VarNums[output.NodeID()]
		assert(v >= 0, "PO has no CNF variable")
		poLits[i] = bsat.Positive(v)
	}
	if fSetPO == 0 {
		if !solver.AddClause(poLits...) {
			Abc_Print(1, "UNSAT after adding PO clauses.\n")
		}
	} else {
		for i, lit := range poLits {
			if i != len(poLits)-1 {
				lit = bsat.Literal(lit ^ 1)
			}
			if !solver.AddClause(lit) {
				Abc_Print(1, "UNSAT after adding PO clauses.\n")
			}
		}
	}
	assumptions := make([]bsat.Literal, 0, num_sel_pis)
	varToSel := make([]int32, data.Variables)
	for i := range varToSel {
		varToSel[i] = -1
	}
	for i := int32(0); i < num_sel_pis; i++ {
		cur := first_sel_pi + i
		v := data.VarNums[graph.Inputs()[cur].NodeID()]
		assert(v >= 0, "selector has no CNF variable")
		varToSel[v] = i
		lit := bsat.Positive(v)
		if Vec_BitEntry(vMark, i) {
			assumptions = append(assumptions, lit)
		} else if !solver.AddClause(lit) {
			Abc_Print(1, "UNSAT after adding selector clause.\n")
		}
	}
	status := solver.SolveWithResourceLimits(assumptions, int64(nConfLimit), 0, 0, 0)
	if status == bsat.Unsatisfiable {
		Abc_Print(1, "UNSAT.\n")
		core := solver.UnsatCore()
		result := Vec_IntAlloc(int32(len(core)))
		for _, lit := range core {
			v := lit >> 1
			if int(v) < len(varToSel) && varToSel[v] >= 0 {
				Vec_IntPush(result, varToSel[v])
			}
		}
		return result
	}
	if status == bsat.Satisfiable {
		Abc_Print(1, "SAT.\n")
	} else {
		Abc_Print(1, "UNKNOWN.\n")
	}
	return nil
}

func Wlc_NtkUnrollWoCex(pChoice *Wlc_Ntk_t, nFrames int32, first_sel_pi int32, num_sel_pis int32) *Gia_Man_t {
	choice := Wlc_NtkBitBlast(pChoice, nil)
	frames := gia.Gia_ManStart(10000)
	gia.Gia_ManSetNameSpec(frames, gia.Gia_ManName(choice), gia.Gia_ManSpec(choice))
	gia.Gia_ManHashAlloc(frames)
	gia.Gia_ObjSetValue(gia.Gia_ManConst0(choice), 0)
	for i := int32(0); i < gia.Gia_ManRegNum(choice); i++ {
		gia.Gia_ObjSetValue(gia.Gia_ManRi(choice, i), 0)
	}
	for f := int32(0); f < nFrames; f++ {
		for i := int32(0); i < gia.Gia_ManPiNum(choice); i++ {
			if i >= first_sel_pi && i < first_sel_pi+num_sel_pis {
				if f == 0 {
					gia.Gia_ObjSetValue(gia.Gia_ManPi(choice, i), gia.Gia_ManAppendCi(frames))
				}
			} else {
				gia.Gia_ObjSetValue(gia.Gia_ManPi(choice, i), gia.Gia_ManAppendCi(frames))
			}
		}
		for i := int32(0); i < gia.Gia_ManRegNum(choice); i++ {
			gia.Gia_ObjSetValue(gia.Gia_ManRo(choice, i), gia.Gia_ObjValue(gia.Gia_ManRi(choice, i)))
		}
		for i := int32(1); i < gia.Gia_ManObjNum(choice); i++ {
			obj := gia.Gia_ManObj(choice, i)
			if gia.Gia_ObjIsAnd(obj) != 0 {
				gia.Gia_ObjSetValue(obj, gia.Gia_ManHashAnd(frames, gia.Gia_ObjFanin0Copy(obj), gia.Gia_ObjFanin1Copy(obj)))
			}
		}
		for i := int32(0); i < gia.Gia_ManCoNum(choice); i++ {
			co := gia.Gia_ManCo(choice, i)
			gia.Gia_ObjSetValue(co, gia.Gia_ObjFanin0Copy(co))
		}
		for i := int32(0); i < gia.Gia_ManPoNum(choice); i++ {
			gia.Gia_ManAppendCo(frames, gia.Gia_ObjValue(gia.Gia_ManPo(choice, i)))
		}
	}
	gia.Gia_ManHashStop(frames)
	gia.Gia_ManSetRegNum(frames, 0)
	frames = gia.Gia_ManCleanup(frames)
	gia.Gia_ManStop(choice)
	return frames
}

func Wlc_NtkUnrollWithCex(pChoice *Wlc_Ntk_t, pCex *Abc_Cex_t, nbits_old_pis int32, num_sel_pis int32, p_num_ppis *int32, sel_pi_first int32, fUsePPI int32) *Gia_Man_t {
	choice := Wlc_NtkBitBlast(pChoice, nil)
	newPiBits := Wlc_NtkNumPiBits(pChoice)
	numPPIs := newPiBits - nbits_old_pis - num_sel_pis
	numUndc := gia.Gia_ManPiNum(choice) - newPiBits
	*p_num_ppis = numPPIs
	Abc_Print(1, "#orig_pis = %d, #ppis = %d, #sel_pis = %d, #undc_pis = %d\n", nbits_old_pis, numPPIs, num_sel_pis, numUndc)
	assert(gia.Gia_ManPiNum(choice) == nbits_old_pis+numPPIs+num_sel_pis+numUndc, "input bit count")
	assert(gia.Gia_ManPiNum(choice) == pCex.NPis+num_sel_pis, "counterexample input count")
	frames := gia.Gia_ManStart(10000)
	gia.Gia_ManSetNameSpec(frames, gia.Gia_ManName(choice), gia.Gia_ManSpec(choice))
	gia.Gia_ManHashAlloc(frames)
	gia.Gia_ObjSetValue(gia.Gia_ManConst0(choice), 0)
	for i := int32(0); i < gia.Gia_ManRegNum(choice); i++ {
		gia.Gia_ObjSetValue(gia.Gia_ManRi(choice, i), 0)
	}
	for f := int32(0); f <= pCex.IFrame; f++ {
		for i := int32(0); i < gia.Gia_ManPiNum(choice); i++ {
			obj := gia.Gia_ManPi(choice, i)
			if i >= nbits_old_pis && i < nbits_old_pis+numPPIs+num_sel_pis {
				isSel := i < nbits_old_pis+num_sel_pis
				if sel_pi_first == 0 {
					isSel = i >= nbits_old_pis+numPPIs
				}
				if f == 0 && isSel {
					gia.Gia_ObjSetValue(obj, gia.Gia_ManAppendCi(frames))
				}
				if !isSel {
					if fUsePPI == 0 {
						gia.Gia_ObjSetValue(obj, gia.Gia_ManAppendCi(frames))
					} else {
						gia.Gia_ObjSetValue(obj, boolToInt32(util.Abc_InfoHasBit(pCex.PData, pCex.NRegs+pCex.NPis*f+i+numUndc)))
					}
				}
			} else if i < nbits_old_pis {
				gia.Gia_ObjSetValue(obj, boolToInt32(util.Abc_InfoHasBit(pCex.PData, pCex.NRegs+pCex.NPis*f+i)))
			} else if i >= nbits_old_pis+numPPIs+num_sel_pis {
				gia.Gia_ObjSetValue(obj, boolToInt32(util.Abc_InfoHasBit(pCex.PData, pCex.NRegs+pCex.NPis*f+i-num_sel_pis-numPPIs)))
			}
		}
		for i := int32(0); i < gia.Gia_ManRegNum(choice); i++ {
			gia.Gia_ObjSetValue(gia.Gia_ManRo(choice, i), gia.Gia_ObjValue(gia.Gia_ManRi(choice, i)))
		}
		for i := int32(1); i < gia.Gia_ManObjNum(choice); i++ {
			obj := gia.Gia_ManObj(choice, i)
			if gia.Gia_ObjIsAnd(obj) != 0 {
				gia.Gia_ObjSetValue(obj, gia.Gia_ManHashAnd(frames, gia.Gia_ObjFanin0Copy(obj), gia.Gia_ObjFanin1Copy(obj)))
			}
		}
		for i := int32(0); i < gia.Gia_ManCoNum(choice); i++ {
			co := gia.Gia_ManCo(choice, i)
			gia.Gia_ObjSetValue(co, gia.Gia_ObjFanin0Copy(co))
		}
		for i := int32(0); i < gia.Gia_ManPoNum(choice); i++ {
			gia.Gia_ManAppendCo(frames, gia.Gia_ObjValue(gia.Gia_ManPo(choice, i)))
		}
	}
	gia.Gia_ManHashStop(frames)
	gia.Gia_ManSetRegNum(frames, 0)
	frames = gia.Gia_ManCleanup(frames)
	gia.Gia_ManStop(choice)
	return frames
}

func Wlc_NtkIntroduceChoices(pNtk *Wlc_Ntk_t, vBlacks *Vec_Int_t, vPPIs *Vec_Int_t) *Wlc_Ntk_t {
	//if ( vBlacks== NULL ) return NULL;
	var vNodes *Vec_Int_t = Vec_IntDup(vBlacks)
	var pNew *Wlc_Ntk_t
	var pObj *Wlc_Obj_t
	var i, k, iObj, iFanin int32
	var vFanins *Vec_Int_t = Vec_IntAlloc(3)
	var vMapNode2Pi *Vec_Int_t = Vec_IntStart(Wlc_NtkObjNumMax(pNtk))
	var vMapNode2Sel *Vec_Int_t = Vec_IntStart(Wlc_NtkObjNumMax(pNtk))
	var nOrigObjNum int32 = Wlc_NtkObjNumMax(pNtk)
	var p *Wlc_Ntk_t = Wlc_NtkDupDfsSimple(pNtk)
	var vPPIMark *Vec_Bit_t = nil

	for i = 0; i < Vec_IntSize(vNodes); i++ {
		pObj = Wlc_NtkObj(pNtk, Vec_IntEntry(vNodes, i))
		// TODO : fix FOs here
		Vec_IntWriteEntry(vNodes, i, Wlc_ObjCopy(pNtk, Wlc_ObjId(pNtk, pObj)))
	}

	if vPPIs != nil {
		vPPIMark = Vec_BitStart(Wlc_NtkObjNumMax(pNtk))
		for i = 0; i < Vec_IntSize(vPPIs); i++ {
			pObj = Wlc_NtkObj(pNtk, Vec_IntEntry(vPPIs, i))
			Vec_IntWriteEntry(vPPIs, i, Wlc_ObjCopy(pNtk, Wlc_ObjId(pNtk, pObj)))
			Vec_BitWriteEntry(vPPIMark, Vec_IntEntry(vPPIs, i), 1)
		}
	}

	// Vec_IntPrint(vNodes);
	Wlc_NtkCleanCopy(p)

	// mark nodes
	for i = 0; i < Vec_IntSize(vNodes); i++ {
		pObj = Wlc_NtkObj(p, Vec_IntEntry(vNodes, i))
		iObj = Wlc_ObjId(p, pObj)
		pObj.Mark = 1
		// add fresh PI with the same number of bits
		Vec_IntWriteEntry(vMapNode2Pi, iObj, Wlc_ObjAlloc(p, int32(WLC_OBJ_PI), Wlc_ObjIsSigned(pObj), Wlc_ObjRange(pObj)-1, 0))
	}

	if vPPIs != nil {
		for i = 0; i < Vec_IntSize(vPPIs); i++ {
			pObj = Wlc_NtkObj(p, Vec_IntEntry(vPPIs, i))
			iObj = Wlc_ObjId(p, pObj)
			pObj.Mark = 1
			// add fresh PI with the same number of bits
			Vec_IntWriteEntry(vMapNode2Pi, iObj, Wlc_ObjAlloc(p, int32(WLC_OBJ_PI), Wlc_ObjIsSigned(pObj), Wlc_ObjRange(pObj)-1, 0))
		}
	}

	// add sel PI
	for i = 0; i < Vec_IntSize(vNodes); i++ {
		pObj = Wlc_NtkObj(p, Vec_IntEntry(vNodes, i))
		iObj = Wlc_ObjId(p, pObj)
		Vec_IntWriteEntry(vMapNode2Sel, iObj, Wlc_ObjAlloc(p, int32(WLC_OBJ_PI), 0, 0, 0))
	}

	// iterate through the nodes in the DFS order
	for i = 1; i < Wlc_NtkObjNumMax(p); i++ {
		pObj = Wlc_NtkObj(p, i)
		var isSigned, bitRange int32
		if i == nOrigObjNum {
			// cout << "break at " << i << endl;
			break
		}

		// update fanins
		for k = 0; k < Wlc_ObjFaninNum(pObj); k++ {
			iFanin = Wlc_ObjFaninId(pObj, k)
			Wlc_ObjFanins(pObj)[k] = Wlc_ObjCopy(p, iFanin)
		}
		// node to remain
		iObj = i

		if pObj.Mark != 0 {
			// clean
			pObj.Mark = 0

			if vPPIMark != nil && Vec_BitEntry(vPPIMark, i) {
				iObj = Vec_IntEntry(vMapNode2Pi, i)
			} else {
				isSigned = Wlc_ObjIsSigned(pObj)
				bitRange = Wlc_ObjRange(pObj)
				Vec_IntClear(vFanins)
				Vec_IntPush(vFanins, Vec_IntEntry(vMapNode2Sel, i))
				Vec_IntPush(vFanins, Vec_IntEntry(vMapNode2Pi, i))
				Vec_IntPush(vFanins, i)
				iObj = Wlc_ObjCreate(p, int32(WLC_OBJ_MUX), isSigned, bitRange-1, 0, vFanins)
			}
		}

		Wlc_ObjSetCopy(p, i, iObj)
	}

	for i = 0; i < Wlc_NtkCoNum(p); i++ {
		pObj = Wlc_NtkCo(p, i)
		iObj = Wlc_ObjId(p, pObj)
		if iObj != Wlc_ObjCopy(p, iObj) {
			if pObj.fIsFi != 0 {
				Wlc_NtkObj(p, Wlc_ObjCopy(p, iObj)).fIsFi = 1
			} else {
				Wlc_NtkObj(p, Wlc_ObjCopy(p, iObj)).fIsPo = 1
			}

			Vec_IntWriteEntry(&p.vCos, i, Wlc_ObjCopy(p, iObj))
		}
	}

	// DumpWlcNtk(p);
	pNew = Wlc_NtkDupDfsSimple(p)

	if vPPIMark != nil {
		Vec_BitFree(vPPIMark)
	}
	Vec_IntFree(vFanins)
	Vec_IntFree(vMapNode2Pi)
	Vec_IntFree(vMapNode2Sel)
	Vec_IntFree(vNodes)
	Wlc_NtkFree(p)

	return pNew
}

func Wlc_NtkCexIsReal(pOrig *Wlc_Ntk_t, pCex *Abc_Cex_t) *Abc_Cex_t {
	graph := Wlc_NtkBitBlast(pOrig, nil)
	cex := util.Abc_CexAlloc(gia.Gia_ManRegNum(graph), gia.Gia_ManPiNum(graph), pCex.IFrame+1)
	gia.Gia_ObjSetValue(gia.Gia_ManConst0(graph), 0)
	for i := int32(0); i < gia.Gia_ManRegNum(graph); i++ {
		gia.Gia_ObjSetValue(gia.Gia_ManRi(graph, i), 0)
	}
	for frame := int32(0); frame <= pCex.IFrame; frame++ {
		for i := int32(0); i < gia.Gia_ManPiNum(graph); i++ {
			bit := util.Abc_InfoHasBit(pCex.PData, pCex.NRegs+pCex.NPis*frame+i)
			gia.Gia_ObjSetValue(gia.Gia_ManPi(graph, i), boolToInt32(bit))
			if bit {
				util.Abc_InfoSetBit(cex.PData, cex.NRegs+cex.NPis*frame+i)
			}
		}
		for i := int32(0); i < gia.Gia_ManRegNum(graph); i++ {
			gia.Gia_ObjSetValue(gia.Gia_ManRo(graph, i), gia.Gia_ObjValue(gia.Gia_ManRi(graph, i)))
		}
		for i := int32(1); i < gia.Gia_ManObjNum(graph); i++ {
			obj := gia.Gia_ManObj(graph, i)
			if gia.Gia_ObjIsAnd(obj) != 0 {
				gia.Gia_ObjSetValue(obj, (gia.Gia_ObjFanin0Copy(obj)&1)&(gia.Gia_ObjFanin1Copy(obj)&1))
			}
		}
		for i := int32(0); i < gia.Gia_ManCoNum(graph); i++ {
			co := gia.Gia_ManCo(graph, i)
			gia.Gia_ObjSetValue(co, gia.Gia_ObjFanin0Copy(co))
		}
		for i := int32(0); i < gia.Gia_ManPoNum(graph); i++ {
			if gia.Gia_ObjValue(gia.Gia_ManPo(graph, i)) == 1 {
				Abc_Print(1, "CEX is real on the original model.\n")
				gia.Gia_ManStop(graph)
				cex.IFrame, cex.IPo = frame, i
				return cex
			}
		}
	}
	gia.Gia_ManStop(graph)
	util.Abc_CexFree(cex)
	return nil
}

func Wlc_NtkAbs2(pNtk *Wlc_Ntk_t, vBlacks *Vec_Int_t, pvFlops **Vec_Int_t) *Wlc_Ntk_t {
	var vFlops *Vec_Int_t = Vec_IntAlloc(100)
	var vNodes *Vec_Int_t = Vec_IntDup(vBlacks)
	var pNew *Wlc_Ntk_t
	var pObj *Wlc_Obj_t
	var i, k, iObj, iFanin int32
	var vMapNode2Pi *Vec_Int_t = Vec_IntStart(Wlc_NtkObjNumMax(pNtk))
	var nOrigObjNum int32 = Wlc_NtkObjNumMax(pNtk)
	var p *Wlc_Ntk_t = Wlc_NtkDupDfsSimple(pNtk)

	for i = 0; i < Wlc_NtkCiNum(pNtk); i++ {
		pObj = Wlc_NtkCi(pNtk, i)
		if Wlc_ObjIsPi(pObj) == 0 {
			Vec_IntPush(vFlops, Wlc_ObjId(pNtk, pObj))
		}
	}

	for i = 0; i < Vec_IntSize(vNodes); i++ {
		pObj = Wlc_NtkObj(pNtk, Vec_IntEntry(vNodes, i))
		Vec_IntWriteEntry(vNodes, i, Wlc_ObjCopy(pNtk, Wlc_ObjId(pNtk, pObj)))
	}

	// mark nodes
	for i = 0; i < Vec_IntSize(vNodes); i++ {
		pObj = Wlc_NtkObj(p, Vec_IntEntry(vNodes, i))
		iObj = Wlc_ObjId(p, pObj)
		pObj.Mark = 1
		// add fresh PI with the same number of bits
		Vec_IntWriteEntry(vMapNode2Pi, iObj, Wlc_ObjAlloc(p, int32(WLC_OBJ_PI), Wlc_ObjIsSigned(pObj), Wlc_ObjRange(pObj)-1, 0))
	}

	Wlc_NtkCleanCopy(p)

	for i = 1; i < Wlc_NtkObjNumMax(p); i++ {
		pObj = Wlc_NtkObj(p, i)
		if i == nOrigObjNum {
			break
		}

		if pObj.Mark != 0 {
			// clean
			pObj.Mark = 0
			iObj = Vec_IntEntry(vMapNode2Pi, i)

		} else {
			// update fanins
			for k = 0; k < Wlc_ObjFaninNum(pObj); k++ {
				iFanin = Wlc_ObjFaninId(pObj, k)
				Wlc_ObjFanins(pObj)[k] = Wlc_ObjCopy(p, iFanin)
			}
			// node to remain
			iObj = i
		}
		Wlc_ObjSetCopy(p, i, iObj)
	}

	for i = 0; i < Wlc_NtkCoNum(p); i++ {
		pObj = Wlc_NtkCo(p, i)
		iObj = Wlc_ObjId(p, pObj)
		if iObj != Wlc_ObjCopy(p, iObj) {
			if pObj.fIsFi != 0 {
				Wlc_NtkObj(p, Wlc_ObjCopy(p, iObj)).fIsFi = 1
			} else {
				Wlc_NtkObj(p, Wlc_ObjCopy(p, iObj)).fIsPo = 1
			}

			Vec_IntWriteEntry(&p.vCos, i, Wlc_ObjCopy(p, iObj))
		}
	}

	pNew = Wlc_NtkDupDfsSimple(p)
	Vec_IntFree(vMapNode2Pi)
	Vec_IntFree(vNodes)
	Wlc_NtkFree(p)

	if pvFlops != nil {
		*pvFlops = vFlops
	} else {
		Vec_IntFree(vFlops)
	}

	return pNew
}

func Wlc_NtkProofReduce(p *Wlc_Ntk_t, pPars *Wlc_Par_t, nFrames int32, vWhites *Vec_Int_t, vBlacks *Vec_Int_t, RunId int32) *Vec_Bit_t {
	var pGiaFrames *Gia_Man_t
	var pNtkWithChoices *Wlc_Ntk_t = nil
	var vCoreSels *Vec_Int_t
	var vChoiceMark *Vec_Bit_t
	var first_sel_pi int32
	var i, Entry int32
	var clk abctime = util.Abc_Clock()

	assert((vWhites != nil) && Vec_IntSize(vWhites) != 0, "vWhites && Vec_IntSize(vWhites)")
	pNtkWithChoices = Wlc_NtkIntroduceChoices(p, vWhites, vBlacks)

	first_sel_pi = Wlc_NtkNumPiBits(pNtkWithChoices) - Vec_IntSize(vWhites)
	pGiaFrames = Wlc_NtkUnrollWoCex(pNtkWithChoices, nFrames, first_sel_pi, Vec_IntSize(vWhites))

	vChoiceMark = Vec_BitStartFull(Vec_IntSize(vWhites))
	vCoreSels = Wlc_NtkGetCoreSels(pGiaFrames, nFrames, first_sel_pi, Vec_IntSize(vWhites), vChoiceMark, 0, pPars, 0, RunId)

	Wlc_NtkFree(pNtkWithChoices)
	Gia_ManStop(pGiaFrames)

	if vCoreSels == nil {
		return nil
	}

	Vec_BitReset(vChoiceMark)
	for i = 0; i < Vec_IntSize(vCoreSels); i++ {
		Entry = Vec_IntEntry(vCoreSels, i)
		Vec_BitWriteEntry(vChoiceMark, Entry, 1)
	}

	Abc_Print(1, "ProofReduce: remove %d out of %d white boxes.", Vec_IntSize(vWhites)-int32(vec.Vec_BitCount(vChoiceMark)), Vec_IntSize(vWhites))
	util.Abc_PrintTime(1, " Time", util.Abc_Clock()-clk)

	Vec_IntFree(vCoreSels)
	return vChoiceMark
}

func Wlc_NtkProofRefine(p *Wlc_Ntk_t, pPars *Wlc_Par_t, pCex *Abc_Cex_t, vBlacks *Vec_Int_t, pvRefine **Vec_Int_t) int32 {
	var pGiaFrames *Gia_Man_t
	var vRefine *Vec_Int_t = nil
	var vUnmark *Vec_Bit_t
	var vChoiceMark *Vec_Bit_t
	var pNtkWithChoices *Wlc_Ntk_t = nil
	var vCoreSels *Vec_Int_t
	var num_ppis int32 = -1
	var Entry, i int32

	if *pvRefine == nil {
		return 0
	}

	vUnmark = Vec_BitStart(Wlc_NtkObjNumMax(p))
	vChoiceMark = Vec_BitStart(Vec_IntSize(vBlacks))

	for i = 0; i < Vec_IntSize(*pvRefine); i++ {
		Entry = Vec_IntEntry(*pvRefine, i)
		Vec_BitWriteEntry(vUnmark, Entry, 1)
	}

	for i = 0; i < Vec_IntSize(vBlacks); i++ {
		Entry = Vec_IntEntry(vBlacks, i)
		if Vec_BitEntry(vUnmark, Entry) {
			Vec_BitWriteEntry(vChoiceMark, i, 1)
		}
	}

	if vBlacks != nil {
		pNtkWithChoices = Wlc_NtkIntroduceChoices(p, vBlacks, nil)
	} else {
		pNtkWithChoices = nil
	}
	pGiaFrames = Wlc_NtkUnrollWithCex(pNtkWithChoices, pCex, Wlc_NtkNumPiBits(p), Vec_IntSize(vBlacks), &num_ppis, 0, pPars.fProofUsePPI)

	if pPars.fProofUsePPI == 0 {
		vCoreSels = Wlc_NtkGetCoreSels(pGiaFrames, pCex.IFrame+1, num_ppis, Vec_IntSize(vBlacks), vChoiceMark, 0, pPars, 0, -1)
	} else {
		vCoreSels = Wlc_NtkGetCoreSels(pGiaFrames, pCex.IFrame+1, 0, Vec_IntSize(vBlacks), vChoiceMark, 0, pPars, 0, -1)
	}

	Wlc_NtkFree(pNtkWithChoices)
	Gia_ManStop(pGiaFrames)
	Vec_BitFree(vUnmark)
	Vec_BitFree(vChoiceMark)

	assert(vCoreSels != nil && Vec_IntSize(vCoreSels) != 0, "Vec_IntSize( vCoreSels )")

	vRefine = Vec_IntAlloc(100)
	for i = 0; i < Vec_IntSize(vCoreSels); i++ {
		Entry = Vec_IntEntry(vCoreSels, i)
		Vec_IntPush(vRefine, Vec_IntEntry(vBlacks, Entry))
	}

	Vec_IntFree(vCoreSels)

	if pPars.fVerbose != 0 {
		Abc_Print(1, "Proof-based refinement reduces %d (out of %d) white boxes\n", Vec_IntSize(*pvRefine)-Vec_IntSize(vRefine), Vec_IntSize(*pvRefine))
	}

	Vec_IntFree(*pvRefine)
	*pvRefine = vRefine

	return 0
}

func Wlc_NtkUpdateBlacks(p *Wlc_Ntk_t, pPars *Wlc_Par_t, pvBlacks **Vec_Int_t, vUnmark *Vec_Bit_t, vSignals *Vec_Int_t) int32 {
	var Entry, i int32
	var pObj *Wlc_Obj_t
	var Count [4]int32
	var vBlacks *Vec_Int_t = Vec_IntAlloc(100)
	var vVec *Vec_Int_t

	assert(*pvBlacks != nil, "*pvBlacks")

	if vSignals != nil {
		vVec = vSignals
	} else {
		vVec = *pvBlacks
	}

	for i = 0; i < Vec_IntSize(vVec); i++ {
		Entry = Vec_IntEntry(vVec, i)
		if Vec_BitEntry(vUnmark, Entry) {
			continue
		}
		Vec_IntPush(vBlacks, Entry)

		pObj = Wlc_NtkObj(p, Entry)
		if pObj.Type == uint32(WLC_OBJ_ARI_ADD) || pObj.Type == uint32(WLC_OBJ_ARI_SUB) || pObj.Type == uint32(WLC_OBJ_ARI_MINUS) {
			Count[0]++
		} else if pObj.Type == uint32(WLC_OBJ_ARI_MULTI) || pObj.Type == uint32(WLC_OBJ_ARI_DIVIDE) || pObj.Type == uint32(WLC_OBJ_ARI_REM) || pObj.Type == uint32(WLC_OBJ_ARI_MODULUS) {
			Count[1]++
		} else if pObj.Type == uint32(WLC_OBJ_MUX) {
			Count[2]++
		} else if Wlc_ObjIsCi(pObj) != 0 && Wlc_ObjIsPi(pObj) == 0 {
			Count[3]++
		}
	}

	Vec_IntFree(*pvBlacks)
	*pvBlacks = vBlacks

	if pPars.fVerbose != 0 {
		fmt.Printf("Abstraction engine marked %d adds/subs, %d muls/divs, %d muxes, and %d flops to be abstracted away.\n", Count[0], Count[1], Count[2], Vec_IntSize(vBlacks)-Count[0]-Count[1]-Count[2])
	}

	return 0
}

func Wlc_NtkMarkLimit(p *Wlc_Ntk_t, pPars *Wlc_Par_t) *Vec_Bit_t {
	var vMarks *Vec_Bit_t = nil
	var vAdds *Vec_Ptr_t = Vec_PtrAlloc(1000)
	var vMuxes *Vec_Ptr_t = Vec_PtrAlloc(1000)
	var vMults *Vec_Ptr_t = Vec_PtrAlloc(1000)
	var vFlops *Vec_Ptr_t = Vec_PtrAlloc(1000)
	var pObj *Wlc_Obj_t
	var i int32
	var pPair *Int_Pair_t = nil

	if pPars.nLimit == ABC_INFINITY {
		return nil
	}

	vMarks = Vec_BitStart(Wlc_NtkObjNumMax(p))

	for i = 1; i < Wlc_NtkObjNumMax(p); i++ {
		pObj = Wlc_NtkObj(p, i)
		if pObj.Type == uint32(WLC_OBJ_ARI_ADD) || pObj.Type == uint32(WLC_OBJ_ARI_SUB) || pObj.Type == uint32(WLC_OBJ_ARI_MINUS) {
			if Wlc_ObjRange(pObj) >= pPars.nBitsAdd {
				pPair = new(Int_Pair_t)
				pPair.first = i
				pPair.second = Wlc_ObjRange(pObj)
				Vec_PtrPush(vAdds, pPair)
			}

		} else if pObj.Type == uint32(WLC_OBJ_ARI_MULTI) || pObj.Type == uint32(WLC_OBJ_ARI_DIVIDE) || pObj.Type == uint32(WLC_OBJ_ARI_REM) || pObj.Type == uint32(WLC_OBJ_ARI_MODULUS) {
			if Wlc_ObjRange(pObj) >= pPars.nBitsMul {
				pPair = new(Int_Pair_t)
				pPair.first = i
				pPair.second = Wlc_ObjRange(pObj)
				Vec_PtrPush(vMults, pPair)
			}

		} else if pObj.Type == uint32(WLC_OBJ_MUX) {
			if Wlc_ObjRange(pObj) >= pPars.nBitsMux {
				pPair = new(Int_Pair_t)
				pPair.first = i
				pPair.second = Wlc_ObjRange(pObj)
				Vec_PtrPush(vMuxes, pPair)
			}

		} else if Wlc_ObjIsCi(pObj) != 0 && Wlc_ObjIsPi(pObj) == 0 {
			if Wlc_ObjRange(pObj) >= pPars.nBitsFlop {
				pPair = new(Int_Pair_t)
				pPair.first = i
				pPair.second = Wlc_ObjRange(pObj)
				Vec_PtrPush(vFlops, pPair)
			}
		}
	}

	Vec_PtrSort(vAdds, IntPairPtrCompare)
	Vec_PtrSort(vMults, IntPairPtrCompare)
	Vec_PtrSort(vMuxes, IntPairPtrCompare)
	Vec_PtrSort(vFlops, IntPairPtrCompare)

	for i = 0; i < Vec_PtrSize(vAdds); i++ {
		pPair = Vec_PtrEntry(vAdds, i).(*Int_Pair_t)
		if i >= pPars.nLimit {
			break
		}
		Vec_BitWriteEntry(vMarks, pPair.first, 1)
	}
	if i != 0 && pPars.fVerbose != 0 {
		Abc_Print(1, "%%PDRA: %d-th ADD has width = %d\n", i, pPair.second)
	}

	for i = 0; i < Vec_PtrSize(vMults); i++ {
		pPair = Vec_PtrEntry(vMults, i).(*Int_Pair_t)
		if i >= pPars.nLimit {
			break
		}
		Vec_BitWriteEntry(vMarks, pPair.first, 1)
	}
	if i != 0 && pPars.fVerbose != 0 {
		Abc_Print(1, "%%PDRA: %d-th MUL has width = %d\n", i, pPair.second)
	}

	for i = 0; i < Vec_PtrSize(vMuxes); i++ {
		pPair = Vec_PtrEntry(vMuxes, i).(*Int_Pair_t)
		if i >= pPars.nLimit {
			break
		}
		Vec_BitWriteEntry(vMarks, pPair.first, 1)
	}
	if i != 0 && pPars.fVerbose != 0 {
		Abc_Print(1, "%%PDRA: %d-th MUX has width = %d\n", i, pPair.second)
	}

	for i = 0; i < Vec_PtrSize(vFlops); i++ {
		pPair = Vec_PtrEntry(vFlops, i).(*Int_Pair_t)
		if i >= pPars.nLimit {
			break
		}
		Vec_BitWriteEntry(vMarks, pPair.first, 1)
	}
	if i != 0 && pPars.fVerbose != 0 {
		Abc_Print(1, "%%PDRA: %d-th FF has width = %d\n", i, pPair.second)
	}

	Vec_PtrFree(vAdds)
	Vec_PtrFree(vMults)
	Vec_PtrFree(vMuxes)
	Vec_PtrFree(vFlops)

	return vMarks
}

func Wlc_NtkSetUnmark(p *Wlc_Ntk_t, pPars *Wlc_Par_t, vUnmark *Vec_Bit_t) {
	var Entry, i int32
	var vMarks *Vec_Bit_t = Wlc_NtkMarkLimit(p, pPars)

	for i = 0; i < int32(vec.Vec_BitSize(vMarks)); i++ {
		Entry = int32(vec.Vec_BitEntry(vMarks, int(i)))
		Vec_BitWriteEntry(vUnmark, i, Entry^1)
	}

	Vec_BitFree(vMarks)
}

func Wlc_NtkGetBlacks(p *Wlc_Ntk_t, pPars *Wlc_Par_t) *Vec_Int_t {
	var vBlacks *Vec_Int_t = Vec_IntAlloc(100)
	var pObj *Wlc_Obj_t
	var i int32
	var Count [4]int32
	var vMarks *Vec_Bit_t = nil
	var nTotal int32

	vMarks = Wlc_NtkMarkLimit(p, pPars)

	for i = 1; i < Wlc_NtkObjNumMax(p); i++ {
		pObj = Wlc_NtkObj(p, i)
		if pObj.Type == uint32(WLC_OBJ_ARI_ADD) || pObj.Type == uint32(WLC_OBJ_ARI_SUB) || pObj.Type == uint32(WLC_OBJ_ARI_MINUS) {
			if Wlc_ObjRange(pObj) >= pPars.nBitsAdd {
				nTotal++
				if vMarks == nil {
					Vec_IntPushUniqueOrder(vBlacks, Wlc_ObjId(p, pObj))
					Count[0]++
				} else if Vec_BitEntry(vMarks, i) {
					Vec_IntPushUniqueOrder(vBlacks, Wlc_ObjId(p, pObj))
					Count[0]++
				}
			}
			continue
		}
		if pObj.Type == uint32(WLC_OBJ_ARI_MULTI) || pObj.Type == uint32(WLC_OBJ_ARI_DIVIDE) || pObj.Type == uint32(WLC_OBJ_ARI_REM) || pObj.Type == uint32(WLC_OBJ_ARI_MODULUS) {
			if Wlc_ObjRange(pObj) >= pPars.nBitsMul {
				nTotal++
				if vMarks == nil {
					Vec_IntPushUniqueOrder(vBlacks, Wlc_ObjId(p, pObj))
					Count[1]++
				} else if Vec_BitEntry(vMarks, i) {
					Vec_IntPushUniqueOrder(vBlacks, Wlc_ObjId(p, pObj))
					Count[1]++
				}
			}
			continue
		}
		if pObj.Type == uint32(WLC_OBJ_MUX) {
			if Wlc_ObjRange(pObj) >= pPars.nBitsMux {
				nTotal++
				if vMarks == nil {
					Vec_IntPushUniqueOrder(vBlacks, Wlc_ObjId(p, pObj))
					Count[2]++
				} else if Vec_BitEntry(vMarks, i) {
					Vec_IntPushUniqueOrder(vBlacks, Wlc_ObjId(p, pObj))
					Count[2]++
				}
			}
			continue
		}
		if Wlc_ObjIsCi(pObj) != 0 && Wlc_ObjIsPi(pObj) == 0 {
			if Wlc_ObjRange(pObj) >= pPars.nBitsFlop {
				nTotal++
				if vMarks == nil {
					Vec_IntPushUniqueOrder(vBlacks, Wlc_ObjId(p, Wlc_ObjFo2Fi(p, pObj)))
					Count[3]++
				} else if Vec_BitEntry(vMarks, i) {
					Vec_IntPushUniqueOrder(vBlacks, Wlc_ObjId(p, Wlc_ObjFo2Fi(p, pObj)))
					Count[3]++
				}
			}
			continue
		}
	}
	if vMarks != nil {
		Vec_BitFree(vMarks)
	}
	if pPars.fVerbose != 0 {
		fmt.Printf("Abstraction engine marked %d adds/subs, %d muls/divs, %d muxes, and %d flops to be abstracted away (out of %d signals).\n", Count[0], Count[1], Count[2], Count[3], nTotal)
	}
	return vBlacks
}

/**Function*************************************************************

  Synopsis    [Mark operators that meet the abstraction criteria.]

  Description [This procedure returns the array of objects (vLeaves) that
  should be abstracted because of their high bit-width. It uses input array (vUnmark)
  to not abstract those objects that have been refined in the previous rounds.]

  SideEffects []

  SeeAlso     []

***********************************************************************/

func Wlc_NtkAbsMarkOpers(p *Wlc_Ntk_t, pPars *Wlc_Par_t, vUnmark *Vec_Bit_t, fVerbose int32) *Vec_Bit_t {
	var vLeaves *Vec_Bit_t = Vec_BitStart(Wlc_NtkObjNumMax(p))
	var pObj *Wlc_Obj_t
	var i int32
	var Count [4]int32
	for i = 1; i < Wlc_NtkObjNumMax(p); i++ {
		pObj = Wlc_NtkObj(p, i)
		if (vUnmark != nil) && Vec_BitEntry(vUnmark, i) { // not allow this object to be abstracted away
			continue
		}
		if pObj.Type == uint32(WLC_OBJ_ARI_ADD) || pObj.Type == uint32(WLC_OBJ_ARI_SUB) || pObj.Type == uint32(WLC_OBJ_ARI_MINUS) {
			if Wlc_ObjRange(pObj) >= pPars.nBitsAdd {
				Vec_BitWriteEntry(vLeaves, Wlc_ObjId(p, pObj), 1)
				Count[0]++
			}
			continue
		}
		if pObj.Type == uint32(WLC_OBJ_ARI_MULTI) || pObj.Type == uint32(WLC_OBJ_ARI_DIVIDE) || pObj.Type == uint32(WLC_OBJ_ARI_REM) || pObj.Type == uint32(WLC_OBJ_ARI_MODULUS) {
			if Wlc_ObjRange(pObj) >= pPars.nBitsMul {
				Vec_BitWriteEntry(vLeaves, Wlc_ObjId(p, pObj), 1)
				Count[1]++
			}
			continue
		}
		if pObj.Type == uint32(WLC_OBJ_MUX) {
			if Wlc_ObjRange(pObj) >= pPars.nBitsMux {
				Vec_BitWriteEntry(vLeaves, Wlc_ObjId(p, pObj), 1)
				Count[2]++
			}
			continue
		}
		if Wlc_ObjIsCi(pObj) != 0 && Wlc_ObjIsPi(pObj) == 0 {
			if Wlc_ObjRange(pObj) >= pPars.nBitsFlop {
				Vec_BitWriteEntry(vLeaves, Wlc_ObjId(p, pObj), 1)
				Count[3]++
			}
			continue
		}
	}
	if fVerbose != 0 {
		fmt.Printf("Abstraction engine marked %d adds/subs, %d muls/divs, %d muxes, and %d flops to be abstracted away.\n", Count[0], Count[1], Count[2], Count[3])
	}
	return vLeaves
}

/*
*Function*************************************************************

	Synopsis    [Marks nodes to be included in the abstracted network.]

	Description [Marks all objects that will be included in the abstracted model.
	Stops at the objects (vLeaves) that are abstracted away. Returns three arrays:
	a subset of original PIs (vPisOld), a subset of pseudo-PIs (vPisNew) and the
	set of flops present as flops in the abstracted network.]

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Wlc_NtkAbsMarkNodes_rec(p *Wlc_Ntk_t, pObj *Wlc_Obj_t, vLeaves *Vec_Bit_t, vPisOld *Vec_Int_t, vPisNew *Vec_Int_t, vFlops *Vec_Int_t) {
	var i, iFanin int32
	if pObj.Mark != 0 {
		return
	}
	pObj.Mark = 1
	if Vec_BitEntry(vLeaves, Wlc_ObjId(p, pObj)) {
		assert(Wlc_ObjIsPi(pObj) == 0, "!Wlc_ObjIsPi(pObj)")
		Vec_IntPush(vPisNew, Wlc_ObjId(p, pObj))
		return
	}
	if Wlc_ObjIsCi(pObj) != 0 {
		if Wlc_ObjIsPi(pObj) != 0 {
			Vec_IntPush(vPisOld, Wlc_ObjId(p, pObj))
		} else {
			Vec_IntPush(vFlops, Wlc_ObjId(p, pObj))
		}
		return
	}
	for i = 0; i < Wlc_ObjFaninNum(pObj); i++ {
		iFanin = Wlc_ObjFaninId(pObj, i)
		Wlc_NtkAbsMarkNodes_rec(p, Wlc_NtkObj(p, iFanin), vLeaves, vPisOld, vPisNew, vFlops)
	}
}
func Wlc_NtkAbsMarkNodes(p *Wlc_Ntk_t, vLeaves *Vec_Bit_t, vPisOld *Vec_Int_t, vPisNew *Vec_Int_t, vFlops *Vec_Int_t) {
	var pObj *Wlc_Obj_t
	var i, Count int32
	Wlc_NtkCleanMarks(p)
	for i = 0; i < Wlc_NtkCoNum(p); i++ {
		pObj = Wlc_NtkCo(p, i)
		Wlc_NtkAbsMarkNodes_rec(p, pObj, vLeaves, vPisOld, vPisNew, vFlops)
	}

	/*
	   Vec_IntClear(vFlops);
	   Wlc_NtkForEachCi( p, pObj, i ) {
	       if ( !Wlc_ObjIsPi(pObj) ) {
	           Vec_IntPush( vFlops, Wlc_ObjId(p, pObj) );
	           pObj->Mark = 1;
	       }
	   }
	*/

	for i = 0; i < Vec_IntSize(vFlops); i++ {
		pObj = Wlc_NtkObj(p, Vec_IntEntry(vFlops, i))
		Wlc_NtkAbsMarkNodes_rec(p, Wlc_ObjFo2Fi(p, pObj), vLeaves, vPisOld, vPisNew, vFlops)
	}
	for i = 0; i < Wlc_NtkObjNumMax(p); i++ {
		pObj = Wlc_NtkObj(p, i)
		Count += int32(pObj.Mark)
	}
	//    printf( "Collected %d old PIs, %d new PIs, %d flops, and %d other objects.\n",
	//        Vec_IntSize(vPisOld), Vec_IntSize(vPisNew), Vec_IntSize(vFlops),
	//        Count - Vec_IntSize(vPisOld) - Vec_IntSize(vPisNew) - Vec_IntSize(vFlops) );
	Vec_IntSort(vPisOld, 0)
	Vec_IntSort(vPisNew, 0)
	Vec_IntSort(vFlops, 0)
	Wlc_NtkCleanMarks(p)
}

/*
*Function*************************************************************

	Synopsis    [Derive word-level abstracted model based on the parameter values.]

	Description [Retuns the word-level abstracted network and the set of pseudo-PIs
	(vPisNew), which were created during abstraction. If the abstraction is
	satisfiable, some of the pseudo-PIs will be un-abstracted. These pseudo-PIs
	and their MFFC cones will be listed in the array (vUnmark), which will
	force the abstraction to not stop at these pseudo-PIs in the future.]

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Wlc_NtkAbs(p *Wlc_Ntk_t, pPars *Wlc_Par_t, vUnmark *Vec_Bit_t, pvPisNew **Vec_Int_t, pvFlops **Vec_Int_t, fVerbose int32) *Wlc_Ntk_t {
	var pNtkNew *Wlc_Ntk_t = nil
	var vPisOld *Vec_Int_t = Vec_IntAlloc(100)
	var vPisNew *Vec_Int_t = Vec_IntAlloc(100)
	var vFlops *Vec_Int_t = Vec_IntAlloc(100)
	var vLeaves *Vec_Bit_t = Wlc_NtkAbsMarkOpers(p, pPars, vUnmark, fVerbose)
	Wlc_NtkAbsMarkNodes(p, vLeaves, vPisOld, vPisNew, vFlops)
	Vec_BitFree(vLeaves)
	pNtkNew = Wlc_NtkDupDfsAbs(p, vPisOld, vPisNew, vFlops)
	Vec_IntFree(vPisOld)
	if pvFlops != nil {
		*pvFlops = vFlops
	} else {
		Vec_IntFree(vFlops)
	}
	if pvPisNew != nil {
		*pvPisNew = vPisNew
	} else {
		Vec_IntFree(vPisNew)
	}
	return pNtkNew
}

/*
*Function*************************************************************

	Synopsis    [Find what objects need to be un-abstracted.]

	Description [Returns a subset of pseudo-PIs (vPisNew), which will be
	prevented from being abstracted in the future rounds of abstraction.
	The AIG manager (pGia) is a bit-level view of the abstracted model.
	The counter-example (pCex) is used to find waht PPIs to refine.]

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Wlc_NtkAbsRefinement(p *Wlc_Ntk_t, pGia *Gia_Man_t, pCex *Abc_Cex_t, vPisNew *Vec_Int_t) *Vec_Int_t {
	var vRefine *Vec_Int_t = Vec_IntAlloc(100)
	var pCexCare *Abc_Cex_t
	var pObj *Wlc_Obj_t
	// count the number of bit-level PPIs and map them into word-level objects they were derived from
	var f, i, b, nRealPis, nPpiBits int32
	var vMap *Vec_Int_t = Vec_IntStartFull(pCex.NPis)
	for i = 0; i < Vec_IntSize(vPisNew); i++ {
		pObj = Wlc_NtkObj(p, Vec_IntEntry(vPisNew, i))
		for b = 0; b < Wlc_ObjRange(pObj); b++ {
			Vec_IntWriteEntry(vMap, nPpiBits, Wlc_ObjId(p, pObj))
			nPpiBits++
		}
	}
	// since PPIs are ordered last, the previous bits are real PIs
	nRealPis = pCex.NPis - nPpiBits
	// find the care-set
	pCexCare = Bmc_CexCareMinimizeAig(pGia, nRealPis, pCex, 1, 0, 0)
	assert(pCexCare.NPis == pCex.NPis, "pCexCare.NPis == pCex.NPis")
	// detect care PPIs
	for f = 0; f <= pCexCare.IFrame; f++ {
		for i = nRealPis; i < pCexCare.NPis; i++ {
			if Abc_InfoHasBit(pCexCare.PData, pCexCare.NRegs+pCexCare.NPis*f+i) {
				Vec_IntPushUniqueOrder(vRefine, Vec_IntEntry(vMap, i-nRealPis))
			}
		}
	}
	Abc_CexFree(pCexCare)
	Vec_IntFree(vMap)
	if Vec_IntSize(vRefine) == 0 { // real CEX
		Vec_IntFreeP(&vRefine)
	}
	return vRefine
}

/*
*Function*************************************************************

	Synopsis    [Mark MFFC cones of the un-abstracted objects.]

	Description [The MFFC cones of the objects in vRefine are traversed
	and all their nodes are marked in vUnmark.]

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Wlc_NtkNodeDeref_rec(p *Wlc_Ntk_t, pNode *Wlc_Obj_t, vUnmark *Vec_Bit_t) int32 {
	var i int32
	var Fanin int32
	var Counter int32 = 1
	if Wlc_ObjIsCi(pNode) != 0 {
		return 0
	}
	Vec_BitWriteEntry(vUnmark, Wlc_ObjId(p, pNode), 1)
	for i = 0; i < Wlc_ObjFaninNum(pNode); i++ {
		Fanin = Wlc_ObjFaninId(pNode, i)
		Vec_IntAddToEntry(&p.vRefs, Fanin, -1)
		if Vec_IntEntry(&p.vRefs, Fanin) == 0 {
			Counter += Wlc_NtkNodeDeref_rec(p, Wlc_NtkObj(p, Fanin), vUnmark)
		}
	}
	return Counter
}
func Wlc_NtkNodeRef_rec(p *Wlc_Ntk_t, pNode *Wlc_Obj_t) int32 {
	var i int32
	var Fanin int32
	var Counter int32 = 1
	if Wlc_ObjIsCi(pNode) != 0 {
		return 0
	}
	for i = 0; i < Wlc_ObjFaninNum(pNode); i++ {
		Fanin = Wlc_ObjFaninId(pNode, i)
		if Vec_IntEntry(&p.vRefs, Fanin) == 0 {
			Counter += Wlc_NtkNodeRef_rec(p, Wlc_NtkObj(p, Fanin))
		}
		Vec_IntAddToEntry(&p.vRefs, Fanin, 1)
	}
	return Counter
}
func Wlc_NtkMarkMffc(p *Wlc_Ntk_t, pNode *Wlc_Obj_t, vUnmark *Vec_Bit_t) int32 {
	var Count1, Count2 int32
	// if this is a flop output, compute MFFC of the corresponding flop input
	for Wlc_ObjIsCi(pNode) != 0 {
		Vec_BitWriteEntry(vUnmark, Wlc_ObjId(p, pNode), 1)
		pNode = Wlc_ObjFo2Fi(p, pNode)
	}
	assert(Wlc_ObjIsCi(pNode) == 0, "!Wlc_ObjIsCi(pNode)")
	// dereference the node (and set the bits in vUnmark)
	Count1 = Wlc_NtkNodeDeref_rec(p, pNode, vUnmark)
	// reference it back
	Count2 = Wlc_NtkNodeRef_rec(p, pNode)
	assert(Count1 == Count2, "Count1 == Count2")
	return Count1
}
func Wlc_NtkRemoveFromAbstraction(p *Wlc_Ntk_t, vRefine *Vec_Int_t, vUnmark *Vec_Bit_t) int32 {
	var pObj *Wlc_Obj_t
	var i, nNodes int32
	if Vec_IntSize(&p.vRefs) == 0 {
		Wlc_NtkSetRefs(p)
	}
	for i = 0; i < Vec_IntSize(vRefine); i++ {
		pObj = Wlc_NtkObj(p, Vec_IntEntry(vRefine, i))
		nNodes += Wlc_NtkMarkMffc(p, pObj, vUnmark)
	}
	return nNodes
}

func Wlc_NtkUnmarkRefinement(p *Wlc_Ntk_t, vRefine *Vec_Int_t, vUnmark *Vec_Bit_t) int32 {
	var pObj *Wlc_Obj_t
	var i, nNodes int32
	for i = 0; i < Vec_IntSize(vRefine); i++ {
		pObj = Wlc_NtkObj(p, Vec_IntEntry(vRefine, i))
		Vec_BitWriteEntry(vUnmark, Wlc_ObjId(p, pObj), 1)
		nNodes++
	}
	return nNodes
}

/*
*Function*************************************************************

	Synopsis    [Computes the map for remapping flop IDs used in the clauses.]

	Description [Takes the original network (Wlc_Ntk_t) and the array of word-level
	flops used in the old abstraction (vFfOld) and those used in the new abstraction
	(vFfNew). Returns the integer map, which remaps every binary flop found
	in the old abstraction into a binary flop found in the new abstraction.]

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Wlc_NtkFlopsRemap(p *Wlc_Ntk_t, vFfOld *Vec_Int_t, vFfNew *Vec_Int_t) *Vec_Int_t {
	var vMap *Vec_Int_t = Vec_IntAlloc(1000)           // the resulting map
	var vMapFfNew2Bit1 *Vec_Int_t = Vec_IntAlloc(1000) // first binary bit of each new word-level flop
	var i, b, iFfOld, iFfNew, iBit1New, nBits int32
	// map object IDs of old flops into their flop indexes
	var vMapFfObj2FfId *Vec_Int_t = Vec_IntStartFull(Wlc_NtkObjNumMax(p))
	for i = 0; i < Vec_IntSize(vFfNew); i++ {
		iFfNew = Vec_IntEntry(vFfNew, i)
		Vec_IntWriteEntry(vMapFfObj2FfId, iFfNew, i)
	}
	// map each new flop index into its first bit
	for i = 0; i < Vec_IntSize(vFfNew); i++ {
		iFfNew = Vec_IntEntry(vFfNew, i)
		var pObj *Wlc_Obj_t = Wlc_NtkObj(p, iFfNew)
		var nRange int32 = Wlc_ObjRange(pObj)
		Vec_IntPush(vMapFfNew2Bit1, nBits)
		nBits += nRange
	}
	assert(Vec_IntSize(vMapFfNew2Bit1) == Vec_IntSize(vFfNew), "Vec_IntSize(vMapFfNew2Bit1) == Vec_IntSize(vFfNew)")
	// remap old binary flops into new binary flops
	for i = 0; i < Vec_IntSize(vFfOld); i++ {
		iFfOld = Vec_IntEntry(vFfOld, i)
		var pObj *Wlc_Obj_t = Wlc_NtkObj(p, iFfOld)
		var nRange int32 = Wlc_ObjRange(pObj)
		iFfNew = Vec_IntEntry(vMapFfObj2FfId, iFfOld)
		assert(iFfNew >= 0, "iFfNew >= 0") // every old flop should be present in the new abstraction
		// find the first bit of this new flop
		iBit1New = Vec_IntEntry(vMapFfNew2Bit1, iFfNew)
		for b = 0; b < nRange; b++ {
			Vec_IntPush(vMap, iBit1New+b)
		}
	}
	Vec_IntFree(vMapFfNew2Bit1)
	Vec_IntFree(vMapFfObj2FfId)
	return vMap
}

/*
*Function*************************************************************

	Synopsis    [Performs PDR with word-level abstraction.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Wla_ManCollectNodes(pWla *Wla_Man_t, fBlack int32) *Vec_Int_t {
	var vNodes *Vec_Int_t = nil
	var i, Entry int32

	assert(pWla.vSignals != nil, "pWla.vSignals")

	vNodes = Vec_IntAlloc(100)
	for i = 0; i < Vec_IntSize(pWla.vSignals); i++ {
		Entry = Vec_IntEntry(pWla.vSignals, i)
		if !(fBlack != 0) && Vec_BitEntry(pWla.vUnmark, Entry) {
			Vec_IntPush(vNodes, Entry)
		}

		if (fBlack != 0) && !Vec_BitEntry(pWla.vUnmark, Entry) {
			Vec_IntPush(vNodes, Entry)
		}
	}

	return vNodes
}

func Wla_ManShrinkAbs(pWla *Wla_Man_t, nFrames int32, RunId int32) int32 {
	var i, Entry int32
	var RetValue int32

	var vWhites *Vec_Int_t = Wla_ManCollectNodes(pWla, 0)
	var vBlacks *Vec_Int_t = Wla_ManCollectNodes(pWla, 1)
	var vCoreMarks *Vec_Bit_t = Wlc_NtkProofReduce(pWla.p, pWla.pPars, nFrames, vWhites, vBlacks, RunId)

	if vCoreMarks == nil {
		Vec_IntFree(vWhites)
		Vec_IntFree(vBlacks)
		return -1
	}

	RetValue = boolToInt32(Vec_IntSize(vWhites) != Vec_BitCount(vCoreMarks))

	for i = 0; i < Vec_IntSize(vWhites); i++ {
		Entry = Vec_IntEntry(vWhites, i)
		if !Vec_BitEntry(vCoreMarks, i) {
			Vec_BitWriteEntry(pWla.vUnmark, Entry, 0)
		}
	}

	Vec_IntFree(vWhites)
	Vec_IntFree(vBlacks)
	Vec_BitFree(vCoreMarks)

	return RetValue
}

func Wla_ManCreateAbs(pWla *Wla_Man_t) *Wlc_Ntk_t {
	var pAbs *Wlc_Ntk_t

	// get abstracted GIA and the set of pseudo-PIs (vBlacks)
	if pWla.vBlacks == nil {
		pWla.vBlacks = Wlc_NtkGetBlacks(pWla.p, pWla.pPars)
		pWla.vSignals = Vec_IntDup(pWla.vBlacks)

	} else {
		Wlc_NtkUpdateBlacks(pWla.p, pWla.pPars, &pWla.vBlacks, pWla.vUnmark, pWla.vSignals)
	}
	pAbs = Wlc_NtkAbs2(pWla.p, pWla.vBlacks, nil)

	return pAbs
}

func Wla_ManBitBlast(pWla *Wla_Man_t, pAbs *Wlc_Ntk_t) *Aig_Man_t {
	var nDcFlops int32
	var pTemp *Gia_Man_t
	var pAig *Aig_Man_t

	pWla.pGia = Wlc_NtkBitBlast(pAbs, nil)

	// if the abstraction has flops with DC-init state,
	// new PIs were introduced by bit-blasting at the end of the PI list
	// here we move these variables to be *before* PPIs, because
	// PPIs are supposed to be at the end of the PI list for refinement
	nDcFlops = Wlc_NtkDcFlopNum(pAbs)
	if nDcFlops > 0 { // DC-init flops are present

		pTemp = pWla.pGia
		pWla.pGia = Gia_ManPermuteInputs(pTemp, Wlc_NtkCountObjBits(pWla.p, pWla.vBlacks), nDcFlops)
		Gia_ManStop(pTemp)
	}
	// if the word-level outputs have to be XORs, this is a place to do it
	if pWla.pPars.fXorOutput != 0 {
		pTemp = pWla.pGia
		pWla.pGia = Gia_ManTransformMiter2(pTemp)
		Gia_ManStop(pTemp)
	}
	if pWla.pPars.fVerbose != 0 {
		fmt.Printf("Derived abstraction with %d objects and %d PPIs. Bit-blasted AIG stats are:\n", Wlc_NtkObjNum(pAbs), Vec_IntSize(pWla.vBlacks))
		Gia_ManPrintStats(pWla.pGia, nil)
	}

	// try to prove abstracted GIA by converting it to AIG and calling PDR
	pAig = Gia_ManToAigSimple(pWla.pGia)

	return pAig
}

func Wla_ManCheckCombUnsat(pWla *Wla_Man_t, pAig *Aig_Man_t) int32 {
	var pPdr *Pdr_Man_t
	var pPdrPars *Pdr_Par_t = pWla.pPdrPars.(*Pdr_Par_t)
	var clk abctime
	var RetValue int32 = -1

	if Aig_ManAndNum(pAig) <= 20000 {
		var pAigScorr *Aig_Man_t
		var ScorrPars Ssw_Pars_t
		var pScorrPars *Ssw_Pars_t = &ScorrPars
		var nAnds int32

		clk = Abc_Clock()

		Ssw_ManSetDefaultParams(pScorrPars)
		pScorrPars.SetStopWhenGone(1, 1)
		pAigScorr = Ssw_SignalCorrespondence(pAig, pScorrPars)
		assert((pAigScorr != nil), "pAigScorr")
		nAnds = Aig_ManAndNum(pAigScorr)
		Aig_ManStop(pAigScorr)

		if nAnds == 0 {
			if pWla.pPars.fVerbose != 0 {
				Abc_PrintTime(1, "SCORR proved UNSAT. Time", Abc_Clock()-clk)
			}
			return 1

		} else if pWla.pPars.fVerbose != 0 {
			Abc_Print(1, "SCORR failed with %d ANDs. ", nAnds)
			Abc_PrintTime(1, "Time", Abc_Clock()-clk)
		}
	}

	clk = Abc_Clock()

	pPdrPars.SetVerbose(0)
	pPdr = Pdr_ManStart(pAig, pPdrPars, nil)
	RetValue = IPdr_ManCheckCombUnsat(pPdr)
	Pdr_ManStop(pPdr)
	pPdrPars.SetVerbose(pWla.pPars.fPdrVerbose)

	pWla.tPdr += Abc_Clock() - clk

	return RetValue
}

func Wla_ManSolveInt(pWla *Wla_Man_t, pAig *Aig_Man_t) int32 {
	var clk abctime
	var pPdr *Pdr_Man_t
	var pPdrPars *Pdr_Par_t = pWla.pPdrPars.(*Pdr_Par_t)
	var pBmcCex *Abc_Cex_t = nil
	var pCexReal *Abc_Cex_t = nil
	var RetValue int32 = -1
	var RunId int32 = Wla_GetGlobalRunId()

	if pWla.vClauses != nil && pWla.pPars.fCheckCombUnsat != 0 {
		clk = Abc_Clock()

		RetValue = Wla_ManCheckCombUnsat(pWla, pAig)
		if RetValue == 1 {
			if pWla.pPars.fVerbose != 0 {
				Abc_PrintTime(1, "ABS becomes combinationally UNSAT. Time", Abc_Clock()-clk)
			}
			return 1
		}

		if pWla.pPars.fVerbose != 0 {
			Abc_PrintTime(1, "Check comb. unsat failed. Time", Abc_Clock()-clk)
		}
	}

	if pWla.pPars.fUseBmc3 != 0 {
		pPdrPars.SetStopFunc(RunId, Wla_CallBackToStop)

		Wla_ManConcurrentBmc3(pWla, Aig_ManDupSimple(pAig), &pBmcCex)
	}

	clk = Abc_Clock()
	pPdr = Pdr_ManStart(pAig, pPdrPars, nil)
	if pWla.vClauses != nil {
		assert(Vec_VecSize(pWla.vClauses) >= 2, "Vec_VecSize( pWla.vClauses) >= 2")

		if pWla.fNewAbs != 0 {
			var clauses *Vec_Vec_t
			if pWla.pPars.fShrinkScratch != 0 {
				clauses = nil
			} else {
				clauses = pWla.vClauses
			}
			IPdr_ManRebuildClauses(pPdr, clauses)
		} else {
			IPdr_ManRestoreClauses(pPdr, pWla.vClauses, nil)
		}

		pWla.fNewAbs = 0
	}

	RetValue = IPdr_ManSolveInt(pPdr, pWla.pPars.fCheckClauses, pWla.pPars.fPushClauses)
	pPdr.AddTotalTime(Abc_Clock() - clk)
	pWla.tPdr += pPdr.TotalTime()
	if pWla.pPars.fLoadTrace != 0 {
		pWla.vClauses = IPdr_ManSaveClauses(pPdr, 0)
	}
	Pdr_ManStop(pPdr)

	if pWla.pPars.fUseBmc3 != 0 {
		Wla_ManJoinThread(pWla, RunId)
	}

	if pBmcCex != nil {
		pWla.pCex = pBmcCex

	} else {
		pWla.pCex = pAig.TakeSequentialCounterexample()
	}

	// consider outcomes
	if pWla.pCex == nil {
		assert((RetValue != 0), "RetValue") // proved or undecided
		return RetValue
	}

	// verify CEX
	pCexReal = Wlc_NtkCexIsReal(pWla.p, pWla.pCex)
	if pCexReal != nil {
		Abc_CexFree(pWla.pCex)
		pWla.pCex = pCexReal
		return 0
	}

	return -1
}

func Wla_ManRefine(pWla *Wla_Man_t) {
	var clk abctime
	var nNodes int32
	var vRefine *Vec_Int_t = nil

	if pWla.fNewAbs != 0 {
		if pWla.pCex != nil {
			Abc_CexFree(pWla.pCex)
		}
		pWla.pCex = nil
		Gia_ManStop(pWla.pGia)
		pWla.pGia = nil
		return
	}

	// perform refinement
	if pWla.pPars.fHybrid != 0 || pWla.pPars.fProofRefine == 0 {
		clk = Abc_Clock()
		vRefine = Wlc_NtkAbsRefinement(pWla.p, pWla.pGia, pWla.pCex, pWla.vBlacks)
		pWla.tCbr += Abc_Clock() - clk

	} else { // proof-based only

		vRefine = Vec_IntDup(pWla.vBlacks)
	}
	if pWla.pPars.fProofRefine != 0 {
		clk = Abc_Clock()
		Wlc_NtkProofRefine(pWla.p, pWla.pPars, pWla.pCex, pWla.vBlacks, &vRefine)
		pWla.tPbr += Abc_Clock() - clk
	}

	if pWla.vClauses != nil && pWla.pPars.fVerbose != 0 {
		var i int32
		var vVec *Vec_Int_t
		for i = 0; i < Vec_VecSize(pWla.vClauses); i++ {
			vVec = Vec_VecEntry(pWla.vClauses, i)
			pWla.nTotalCla += Vec_IntSize(vVec)
		}
	}

	// update the set of objects to be un-abstracted
	clk = Abc_Clock()
	if pWla.pPars.fMFFC != 0 {
		nNodes = Wlc_NtkRemoveFromAbstraction(pWla.p, vRefine, pWla.vUnmark)
		if pWla.pPars.fVerbose != 0 {
			fmt.Printf("Refinement of CEX in frame %d came up with %d un-abstacted PPIs, whose MFFCs include %d objects.\n", pWla.pCex.IFrame, Vec_IntSize(vRefine), nNodes)
		}

	} else {
		nNodes = Wlc_NtkUnmarkRefinement(pWla.p, vRefine, pWla.vUnmark)
		if pWla.pPars.fVerbose != 0 {
			fmt.Printf("Refinement of CEX in frame %d came up with %d un-abstacted PPIs.\n", pWla.pCex.IFrame, Vec_IntSize(vRefine))
		}

	}
	/*
	   if ( pWla->pPars->fVerbose )
	   {
	       Wlc_NtkAbsAnalyzeRefine( pWla->p, pWla->vBlacks, pWla->vUnmark, &pWla->nDisj, &pWla->nNDisj );
	       Abc_Print( 1, "Refine analysis (total): %d disjoint PPIs and %d non-disjoint PPIs\n", pWla->nDisj, pWla->nNDisj );
	   }
	*/
	pWla.tCbr += Abc_Clock() - clk

	// Experimental
	/*
	   if ( pWla->pCex->iFrame > 0 )
	   {
	       Vec_Int_t * vWhites = Wla_ManCollectWhites( pWla );
	       Vec_Bit_t * vCore = Wlc_NtkProofReduce( pWla->p, pWla->pPars, pWla->pCex->iFrame, vWhites );
	       Vec_IntFree( vWhites );
	       Vec_BitFree( vCore );
	   }
	*/

	pWla.iCexFrame = pWla.pCex.IFrame

	Vec_IntFree(vRefine)
	Gia_ManStop(pWla.pGia)
	pWla.pGia = nil
	Abc_CexFree(pWla.pCex)
	pWla.pCex = nil
}

func Wla_ManStart(pNtk *Wlc_Ntk_t, pPars *Wlc_Par_t) *Wla_Man_t {
	var p *Wla_Man_t = new(Wla_Man_t)
	var pPdrPars *Pdr_Par_t
	p.p = pNtk
	p.pPars = pPars
	p.vUnmark = Vec_BitStart(Wlc_NtkObjNumMax(pNtk))

	pPdrPars = new(Pdr_Par_t)
	Pdr_ManSetDefaultParams(pPdrPars)
	pPdrPars.ConfigureWlc(pPars.fPdrVerbose, pPars.RunId, pPars.pFuncStop, pPars.fPdra)
	p.pPdrPars = pPdrPars

	p.iCexFrame = 0
	p.fNewAbs = 0

	p.nIters = 1
	p.nTotalCla = 0
	p.nDisj = 0
	p.nNDisj = 0

	return p
}

func Wla_ManStop(p *Wla_Man_t) {
	if p.vBlacks != nil {
		Vec_IntFree(p.vBlacks)
	}
	if p.vSignals != nil {
		Vec_IntFree(p.vSignals)
	}
	if p.pGia != nil {
		Gia_ManStop(p.pGia)
	}
	if p.pCex != nil {
		Abc_CexFree(p.pCex)
	}
	Vec_BitFree(p.vUnmark)
	p.pPdrPars = nil
}

func Wla_ManSolve(pWla *Wla_Man_t, pPars *Wlc_Par_t) int32 {
	var clk abctime = Abc_Clock()
	var tTotal abctime
	var pAbs *Wlc_Ntk_t = nil
	var pAig *Aig_Man_t = nil

	var RetValue int32 = -1

	// perform refinement iterations
	for pWla.nIters = 1; pWla.nIters < pPars.nIterMax; pWla.nIters++ {
		if pPars.fVerbose != 0 {
			fmt.Printf("\nIteration %d:\n", pWla.nIters)
		}

		pAbs = Wla_ManCreateAbs(pWla)

		pAig = Wla_ManBitBlast(pWla, pAbs)
		Wlc_NtkFree(pAbs)

		RetValue = Wla_ManSolveInt(pWla, pAig)
		Aig_ManStop(pAig)

		if RetValue != -1 || (pPars.pFuncStop != nil && pPars.pFuncStop(pPars.RunId) != 0) {
			break
		}

		Wla_ManRefine(pWla)
	}

	// report the result
	if pPars.fVerbose != 0 {
		fmt.Printf("\n")
	}
	fmt.Printf("Abstraction ")
	if RetValue == 0 {
		fmt.Printf("resulted in a real CEX")
	} else if RetValue == 1 {
		fmt.Printf("is successfully proved")
	} else {
		fmt.Printf("timed out")
	}
	fmt.Printf(" after %d iterations. ", pWla.nIters)
	tTotal = Abc_Clock() - clk
	Abc_PrintTime(1, "Time", tTotal)

	if pPars.fVerbose != 0 {
		Abc_Print(1, "PDRA reused %d clauses.\n", pWla.nTotalCla)
	}
	if pPars.fVerbose != 0 {
		ABC_PRTP("PDR          ", pWla.tPdr, tTotal)
		ABC_PRTP("CEX Refine   ", pWla.tCbr, tTotal)
		ABC_PRTP("Proof Refine ", pWla.tPbr, tTotal)
		ABC_PRTP("Misc.        ", tTotal-pWla.tPdr-pWla.tCbr-pWla.tPbr, tTotal)
		ABC_PRTP("Total        ", tTotal, tTotal)
	}

	return RetValue
}

func Wlc_NtkPdrAbs(p *Wlc_Ntk_t, pPars *Wlc_Par_t) int32 {
	var pWla *Wla_Man_t = nil

	var RetValue int32 = -1

	pWla = Wla_ManStart(p, pPars)

	RetValue = Wla_ManSolve(pWla, pPars)

	Wla_ManStop(pWla)

	return RetValue
}

/*
*Function*************************************************************

	Synopsis    [Performs abstraction.]

	Description [Derives initial abstraction based on user-specified
	parameter values, which tell what is the smallest bit-width of a
	primitive that is being abstracted away.  Currently only add/sub,
	mul/div, mux, and flop are supported with individual parameters.
	The second step is to refine the initial abstraction until the
	point when the property is proved.]

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Wlc_NtkAbsCore(p *Wlc_Ntk_t, pPars *Wlc_Par_t) int32 {
	var clk abctime = Abc_Clock()
	var vBlacks *Vec_Int_t = nil
	var nIters int32
	var nNodes int32
	var nDcFlops int32
	var RetValue int32 = -1
	// start the bitmap to mark objects that cannot be abstracted because of refinement
	// currently, this bitmap is empty because abstraction begins without refinement
	var vUnmark *Vec_Bit_t = Vec_BitStart(Wlc_NtkObjNumMax(p))
	// set up parameters to run PDR
	var PdrPars Pdr_Par_t
	var pPdrPars *Pdr_Par_t = &PdrPars
	Pdr_ManSetDefaultParams(pPdrPars)
	//pPdrPars->fUseAbs    = 1;   // use 'pdr -t'  (on-the-fly abstraction)
	//pPdrPars->fCtgs      = 1;   // use 'pdr -nc' (improved generalization)
	//pPdrPars->fSkipDown  = 0;   // use 'pdr -nc' (improved generalization)
	//pPdrPars->nRestLimit = 500; // reset queue or proof-obligations when it gets larger than this
	pPdrPars.SetVerbose(pPars.fPdrVerbose)
	// perform refinement iterations
	for nIters = 1; nIters < pPars.nIterMax; nIters++ {
		var pAig *Aig_Man_t
		var pCex *Abc_Cex_t
		var vPisNew, vRefine *Vec_Int_t
		var pGia, pTemp *Gia_Man_t
		var pAbs *Wlc_Ntk_t

		if pPars.fVerbose != 0 {
			fmt.Printf("\nIteration %d:\n", nIters)
		}

		// get abstracted GIA and the set of pseudo-PIs (vPisNew)
		if pPars.fAbs2 != 0 {
			if vBlacks == nil {
				vBlacks = Wlc_NtkGetBlacks(p, pPars)
			} else {
				Wlc_NtkUpdateBlacks(p, pPars, &vBlacks, vUnmark, nil)
			}
			pAbs = Wlc_NtkAbs2(p, vBlacks, nil)
			vPisNew = Vec_IntDup(vBlacks)

		} else {
			if nIters == 1 && pPars.nLimit < ABC_INFINITY {
				Wlc_NtkSetUnmark(p, pPars, vUnmark)
			}

			pAbs = Wlc_NtkAbs(p, pPars, vUnmark, &vPisNew, nil, pPars.fVerbose)
		}
		pGia = Wlc_NtkBitBlast(pAbs, nil)

		// if the abstraction has flops with DC-init state,
		// new PIs were introduced by bit-blasting at the end of the PI list
		// here we move these variables to be *before* PPIs, because
		// PPIs are supposed to be at the end of the PI list for refinement
		nDcFlops = Wlc_NtkDcFlopNum(pAbs)
		if nDcFlops > 0 { // DC-init flops are present

			pTemp = pGia
			pGia = Gia_ManPermuteInputs(pTemp, Wlc_NtkCountObjBits(p, vPisNew), nDcFlops)
			Gia_ManStop(pTemp)
		}
		// if the word-level outputs have to be XORs, this is a place to do it
		if pPars.fXorOutput != 0 {
			pTemp = pGia
			pGia = Gia_ManTransformMiter2(pTemp)
			Gia_ManStop(pTemp)
		}
		if pPars.fVerbose != 0 {
			fmt.Printf("Derived abstraction with %d objects and %d PPIs. Bit-blasted AIG stats are:\n", Wlc_NtkObjNum(pAbs), Vec_IntSize(vPisNew))
			Gia_ManPrintStats(pGia, nil)
		}
		Wlc_NtkFree(pAbs)

		// try to prove abstracted GIA by converting it to AIG and calling PDR
		pAig = Gia_ManToAigSimple(pGia)
		RetValue = Pdr_ManSolve(pAig, pPdrPars)
		pCex = pAig.TakeSequentialCounterexample()
		Aig_ManStop(pAig)

		// consider outcomes
		if pCex == nil {
			assert((RetValue != 0), "RetValue") // proved or undecided
			Gia_ManStop(pGia)
			Vec_IntFree(vPisNew)
			break
		}

		// perform refinement
		vRefine = Wlc_NtkAbsRefinement(p, pGia, pCex, vPisNew)
		Gia_ManStop(pGia)
		Vec_IntFree(vPisNew)
		if vRefine == nil { // real CEX

			Abc_CexFree(pCex) // return CEX in the future
			break
		}

		// update the set of objects to be un-abstracted
		nNodes = Wlc_NtkRemoveFromAbstraction(p, vRefine, vUnmark)
		if pPars.fVerbose != 0 {
			fmt.Printf("Refinement of CEX in frame %d came up with %d un-abstacted PPIs, whose MFFCs include %d objects.\n", pCex.IFrame, Vec_IntSize(vRefine), nNodes)
		}
		Vec_IntFree(vRefine)
		Abc_CexFree(pCex)
	}
	Vec_IntFree(vBlacks)
	Vec_BitFree(vUnmark)
	// report the result
	if pPars.fVerbose != 0 {
		fmt.Printf("\n")
	}
	fmt.Printf("Abstraction ")
	if RetValue == 0 {
		fmt.Printf("resulted in a real CEX")
	} else if RetValue == 1 {
		fmt.Printf("is successfully proved")
	} else {
		fmt.Printf("timed out")
	}
	fmt.Printf(" after %d iterations. ", nIters)
	Abc_PrintTime(1, "Time", Abc_Clock()-clk)
	return RetValue
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END
