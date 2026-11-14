package acb

import (
	"fmt"
	"strings"
	"unsafe"
)

import "os"

/**CFile****************************************************************

  FileName    [acbTest.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Hierarchical word-level netlist.]

  Synopsis    []

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - July 21, 2015.]

  Revision    [$Id: acbTest.c,v 1.00 2014/11/29 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "acb.h"
//#include "aig/saig/saig.h"
//#include "aig/gia/giaAig.h"
//#include "base/abc/abc.h"
//#include "proof/fraig/fraig.h"
//#include "misc/util/utilTruth.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

var fForceZero int32

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
func Gia_ManSimTry(pF *Gia_Man_t, pG *Gia_Man_t) {
	var i int32
	var j int32
	var n int32
	var nWords int32 = 500
	var vSimsF, vSimsG *Vec_Wrd_t
	Abc_Random(1)
	inputs := Vec_WrdStartRandom(Gia_ManCiNum(pF) * nWords)
	Gia_ManSetSimInputs(pF, inputs)
	Gia_ManSetSimInputs(pG, Vec_WrdDup(inputs))
	vSimsF = Gia_ManSimPatSim(pF)
	vSimsG = Gia_ManSimPatSim(pG)
	assert(Gia_ManObjNum(pF)*nWords == Vec_WrdSize(vSimsF), "Gia_ManObjNum(pF) * nWords == Vec_WrdSize(vSimsF)")
	for i = 0; i < Gia_ManCoNum(pF)/2; i++ {
		var pObjFb *Gia_Obj_t = Gia_ManCo(pF, 2*i+0)
		var pObjFx *Gia_Obj_t = Gia_ManCo(pF, 2*i+1)
		var pObjGb *Gia_Obj_t = Gia_ManCo(pG, 2*i+0)
		var pObjGx *Gia_Obj_t = Gia_ManCo(pG, 2*i+1)
		var pSimFb *word = Vec_WrdEntryP(vSimsF, Gia_ObjId(pF, pObjFb)*nWords)
		var pSimFx *word = Vec_WrdEntryP(vSimsF, Gia_ObjId(pF, pObjFx)*nWords)
		var pSimGb *word = Vec_WrdEntryP(vSimsG, Gia_ObjId(pG, pObjGb)*nWords)
		var pSimGx *word = Vec_WrdEntryP(vSimsG, Gia_ObjId(pG, pObjGx)*nWords)

		var nBitsFx int32 = Abc_TtCountOnesVec(pSimFx, nWords)
		var nBitsF1 int32 = Abc_TtCountOnesVecMask(pSimFx, pSimFb, nWords, 1)
		var nBitsF0 int32 = nWords*64 - nBitsFx - nBitsF1

		var nBitsGx int32 = Abc_TtCountOnesVec(pSimGx, nWords)
		var nBitsG1 int32 = Abc_TtCountOnesVecMask(pSimGx, pSimGb, nWords, 1)
		var nBitsG0 int32 = nWords*64 - nBitsGx - nBitsG1

		fmt.Printf("Output %4d : ", i)

		fmt.Printf("    RF :  ")
		fmt.Printf("0 =%7.3f %%  ", 100.0*float64(nBitsF0)/64/float64(nWords))
		fmt.Printf("1 =%7.3f %%  ", 100.0*float64(nBitsF1)/64/float64(nWords))
		fmt.Printf("X =%7.3f %%  ", 100.0*float64(nBitsFx)/64/float64(nWords))

		fmt.Printf("   GF :  ")
		fmt.Printf("0 =%7.3f %%  ", 100.0*float64(nBitsG0)/64/float64(nWords))
		fmt.Printf("1 =%7.3f %%  ", 100.0*float64(nBitsG1)/64/float64(nWords))
		fmt.Printf("X =%7.3f %%  ", 100.0*float64(nBitsGx)/64/float64(nWords))

		fmt.Printf("\n")
		if i == 20 {
			break
		}
	}

	fmt.Printf("\n")
	for j = 0; j < 20; j++ {
		for n = 0; n < 2; n++ {
			for i = 0; i < Gia_ManCoNum(pF)/2; i++ {
				var pObjFb *Gia_Obj_t = Gia_ManCo(pF, 2*i+0)
				var pObjFx *Gia_Obj_t = Gia_ManCo(pF, 2*i+1)
				var pObjGb *Gia_Obj_t = Gia_ManCo(pG, 2*i+0)
				var pObjGx *Gia_Obj_t = Gia_ManCo(pG, 2*i+1)
				var pSimFb *word = Vec_WrdEntryP(vSimsF, Gia_ObjId(pF, pObjFb)*nWords)
				var pSimFx *word = Vec_WrdEntryP(vSimsF, Gia_ObjId(pF, pObjFx)*nWords)
				var pSimGb *word = Vec_WrdEntryP(vSimsG, Gia_ObjId(pG, pObjGb)*nWords)
				var pSimGx *word = Vec_WrdEntryP(vSimsG, Gia_ObjId(pG, pObjGx)*nWords)
				var pSimb *word
				if n != 0 {
					pSimb = pSimGb
				} else {
					pSimb = pSimFb
				}
				var pSimx *word
				if n != 0 {
					pSimx = pSimGx
				} else {
					pSimx = pSimFx
				}
				if Abc_TtGetBit(pSimx, j) != 0 {
					fmt.Printf("x")
				} else if Abc_TtGetBit(pSimb, j) != 0 {
					fmt.Printf("1")
				} else {
					fmt.Printf("0")
				}
			}
			fmt.Printf("\n")
		}
		fmt.Printf("\n")
	}

	Vec_WrdFree(vSimsF)
	Vec_WrdFree(vSimsG)
	fmt.Printf("\n")
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Gia_ManDualNot(p *Gia_Man_t, LitA []int32, LitZ []int32) {
	LitZ[0] = Abc_LitNot(LitA[0])
	LitZ[1] = LitA[1]

	if fForceZero != 0 {
		LitZ[0] = Gia_ManHashAnd(p, LitZ[0], Abc_LitNot(LitZ[1]))
	}
}

// computes Z = XOR(A, B) where A, B, Z belong to {0,1,x} encoded as 0=00, 1=01, x=1-
func Gia_ManDualXor2(p *Gia_Man_t, LitA []int32, LitB []int32, LitZ []int32) {
	LitZ[0] = Gia_ManHashXor(p, LitA[0], LitB[0])
	LitZ[1] = Gia_ManHashOr(p, LitA[1], LitB[1])

	if fForceZero != 0 {
		LitZ[0] = Gia_ManHashAnd(p, LitZ[0], Abc_LitNot(LitZ[1]))
	}
}
func Gia_ManDualXorN(p *Gia_Man_t, pLits []int32, n int32, LitZ []int32) {
	var i int32
	LitZ[0] = 0
	LitZ[1] = 0
	for i = 0; i < n; i++ {
		LitZ[0] = Gia_ManHashXor(p, LitZ[0], pLits[2*i])
		LitZ[1] = Gia_ManHashOr(p, LitZ[1], pLits[2*i+1])
	}
}

// computes Z = AND(A, B) where A, B, Z belong to {0,1,x} encoded as 0=00, 1=01, z=1-
func Gia_ManDualAnd2(p *Gia_Man_t, LitA []int32, LitB []int32, LitZ []int32) {
	var ZeroA int32 = Gia_ManHashAnd(p, Abc_LitNot(LitA[0]), Abc_LitNot(LitA[1]))
	var ZeroB int32 = Gia_ManHashAnd(p, Abc_LitNot(LitB[0]), Abc_LitNot(LitB[1]))
	var ZeroZ int32 = Gia_ManHashOr(p, ZeroA, ZeroB)
	LitZ[0] = Gia_ManHashAnd(p, LitA[0], LitB[0])
	LitZ[1] = Gia_ManHashAnd(p, Gia_ManHashOr(p, LitA[1], LitB[1]), Abc_LitNot(ZeroZ))

	//LitZ[0]   = Gia_ManHashAnd( p, Gia_ManHashAnd(p, LitA[0], Abc_LitNot(LitA[1])), Gia_ManHashAnd(p, LitB[0], Abc_LitNot(LitB[1])) );
	//LitZ[1]   = Gia_ManHashAnd( p, Gia_ManHashOr(p, LitA[0], LitA[1]), Gia_ManHashOr(p, LitB[0], LitB[1]) );
	//LitZ[1]   = Gia_ManHashAnd( p, LitZ[1], Abc_LitNot(LitZ[0]) );
}
func Gia_ManDualAndN(p *Gia_Man_t, pLits []int32, n int32, LitZ []int32) {
	var i, LitZero, LitOne int32
	LitZ[0] = 1
	for i = 0; i < n; i++ {
		var Lit int32 = Gia_ManHashAnd(p, Abc_LitNot(pLits[2*i]), Abc_LitNot(pLits[2*i+1]))
		LitZero = Gia_ManHashOr(p, LitZero, Lit)
		LitOne = Gia_ManHashOr(p, LitOne, pLits[2*i+1])
		LitZ[0] = Gia_ManHashAnd(p, LitZ[0], pLits[2*i])
	}
	LitZ[1] = Gia_ManHashAnd(p, LitOne, Abc_LitNot(LitZero))

	if fForceZero != 0 {
		LitZ[0] = Gia_ManHashAnd(p, LitZ[0], Abc_LitNot(LitZ[1]))
	}
}

/*
module _DC(O, C, D);

	output O;
	input C, D;
	assign O = D ? 1'bx : C;

endmodule
*/
func Gia_ManDualDc(p *Gia_Man_t, LitC []int32, LitD []int32, LitZ []int32) {
	LitZ[0] = LitC[0]
	//    LitZ[0]   = Gia_ManHashMux( p, LitD[0], 0, LitC[0] );
	LitZ[1] = Gia_ManHashOr(p, Gia_ManHashOr(p, LitD[0], LitD[1]), LitC[1])

	if fForceZero != 0 {
		LitZ[0] = Gia_ManHashAnd(p, LitZ[0], Abc_LitNot(LitZ[1]))
	}
}
func Gia_ManDualMux(p *Gia_Man_t, LitC []int32, LitT []int32, LitE []int32, LitZ []int32) {
	/*
	   // total logic size: 18 nodes
	   int Xnor = Gia_ManHashXor( p, Abc_LitNot(LitT[0]), LitE[0] );
	   int Cond = Gia_ManHashAnd( p, Abc_LitNot(LitT[1]), Abc_LitNot(LitE[1]) );
	   int pTempE[2], pTempT[2];
	   pTempE[0] = Gia_ManHashMux( p, LitC[0], LitT[0], LitE[0] );
	   pTempE[1] = Gia_ManHashMux( p, LitC[0], LitT[1], LitE[1] );
	   //pTempT[0] = LitT[0];
	   pTempT[0] = Gia_ManHashAnd( p, LitT[0], LitE[0] );
	   pTempT[1] = Gia_ManHashAnd( p, Cond, Xnor );
	   LitZ[0] = Gia_ManHashMux( p, LitC[1], pTempT[0], pTempE[0] );
	   LitZ[1] = Gia_ManHashMux( p, LitC[1], pTempT[1], pTempE[1] );
	*/
	// total logic size: 14 nodes
	var Xnor int32 = Gia_ManHashXor(p, Abc_LitNot(LitT[0]), LitE[0])
	var Cond int32 = Gia_ManHashAnd(p, Abc_LitNot(LitT[1]), Abc_LitNot(LitE[1]))
	var XVal1 int32 = Abc_LitNot(Gia_ManHashAnd(p, Cond, Xnor))
	var XVal0 int32 = Gia_ManHashMux(p, LitC[0], LitT[1], LitE[1])
	LitZ[0] = Gia_ManHashMux(p, LitC[0], LitT[0], LitE[0])
	LitZ[1] = Gia_ManHashMux(p, LitC[1], XVal1, XVal0)

	if fForceZero != 0 {
		LitZ[0] = Gia_ManHashAnd(p, LitZ[0], Abc_LitNot(LitZ[1]))
	}
}
func Gia_ManDualCompare(p *Gia_Man_t, LitF []int32, LitS []int32) int32 {
	var iMiter int32 = Gia_ManHashXor(p, LitF[0], LitS[0])
	iMiter = Gia_ManHashOr(p, LitF[1], iMiter)
	iMiter = Gia_ManHashAnd(p, Abc_LitNot(LitS[1]), iMiter)
	return iMiter
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Acb_ObjToGiaDual(pNew *Gia_Man_t, p *Acb_Ntk_t, iObj int32, vTemp *Vec_Int_t, vCopies *Vec_Int_t, pRes []int32) {
	//char * pName = Abc_NamStr( p->pDesign->pStrs, Acb_ObjName(p, iObj) );
	var pFanin []int32
	var iFanin int32
	var k int32
	var Type int32
	assert(Acb_ObjIsCio(p, iObj) == 0, "Acb_ObjIsCio(p, iObj) == 0")
	Vec_IntClear(vTemp)
	// inlining of macro Acb_ObjForEachFaninFast
	for k = 0; k < Acb_ObjFaninNum(p, iObj); k++ {
		pFanin = Acb_ObjFanins(p, iObj)
		iFanin = pFanin[k+1]
		pLits := Vec_IntArray(vCopies)[2*iFanin : 2*iFanin+2]
		assert(pLits[0] >= 0 && pLits[1] >= 0, "pLits[0] >= 0 && pLits[1] >= 0")
		Vec_IntPushTwo(vTemp, pLits[0], pLits[1])
	}
	Type = int32(Acb_ObjType(p, iObj))
	if Type == ABC_OPER_CONST_F {
		pRes[0] = 0
		pRes[1] = 0
		return
	}
	if Type == ABC_OPER_CONST_T {
		pRes[0] = 1
		pRes[1] = 0
		return
	}
	if Type == ABC_OPER_CONST_X {
		pRes[0] = 0
		pRes[1] = 1
		return
	}
	if Type == ABC_OPER_BIT_BUF {
		pRes[0] = Vec_IntEntry(vTemp, 0)
		pRes[1] = Vec_IntEntry(vTemp, 1)
		return
	}
	if Type == ABC_OPER_BIT_INV {
		Gia_ManDualNot(pNew, Vec_IntArray(vTemp), pRes)
		return
	}
	if Type == ABC_OPER_TRI {
		// in the file inputs are ordered as follows:  _DC \n6_5[9] ( .O(\108 ), .C(\96 ), .D(\107 ));
		// in this code, we expect them as follows: void Gia_ManDualDc( Gia_Man_t * p, int LitC[2], int LitD[2], int LitZ[2] )
		assert(Vec_IntSize(vTemp) == 4, "Vec_IntSize(vTemp) == 4")
		Gia_ManDualDc(pNew, Vec_IntArray(vTemp), Vec_IntArray(vTemp)[2:], pRes)
		return
	}
	if Type == ABC_OPER_BIT_MUX {
		// in the file inputs are ordered as follows:  _HMUX \U$1 ( .O(\282 ), .I0(1'b1), .I1(\277 ), .S(\281 ));
		// in this code, we expect them as follows: void Gia_ManDualMux( Gia_Man_t * p, int LitC[2], int LitT[2], int LitE[2], int LitZ[2] )
		assert(Vec_IntSize(vTemp) == 6, "Vec_IntSize(vTemp) == 6")
		// inlining of macro ABC_SWAP
		{
			var t int32 = Vec_IntArray(vTemp)[0]
			Vec_IntArray(vTemp)[0] = Vec_IntArray(vTemp)[4]
			Vec_IntArray(vTemp)[4] = t
		}
		// inlining of macro ABC_SWAP
		{
			var t int32 = Vec_IntArray(vTemp)[1]
			Vec_IntArray(vTemp)[1] = Vec_IntArray(vTemp)[5]
			Vec_IntArray(vTemp)[5] = t
		}
		Gia_ManDualMux(pNew, Vec_IntArray(vTemp), Vec_IntArray(vTemp)[2:], Vec_IntArray(vTemp)[4:], pRes)
		return
	}
	if Type == ABC_OPER_BIT_AND || Type == ABC_OPER_BIT_NAND {
		Gia_ManDualAndN(pNew, Vec_IntArray(vTemp), Vec_IntSize(vTemp)/2, pRes)
		if Type == ABC_OPER_BIT_NAND {
			pRes[0] = Abc_LitNot(pRes[0])
		}
		return
	}
	if Type == ABC_OPER_BIT_OR || Type == ABC_OPER_BIT_NOR {
		pArray := Vec_IntArray(vTemp)
		for k = 0; k < Vec_IntSize(vTemp)/2; k++ {
			pArray[2*k] = Abc_LitNot(pArray[2*k])
		}
		Gia_ManDualAndN(pNew, pArray, Vec_IntSize(vTemp)/2, pRes)
		if Type == ABC_OPER_BIT_OR {
			pRes[0] = Abc_LitNot(pRes[0])
		}
		return
	}
	if Type == ABC_OPER_BIT_XOR || Type == ABC_OPER_BIT_NXOR {
		assert(Vec_IntSize(vTemp) == 4, "Vec_IntSize(vTemp) == 4")
		Gia_ManDualXor2(pNew, Vec_IntArray(vTemp), Vec_IntArray(vTemp)[2:], pRes)
		if Type == ABC_OPER_BIT_NXOR {
			pRes[0] = Abc_LitNot(pRes[0])
		}
		return
	}
	assert(false, "0")
}
func Acb_NtkGiaDeriveDual(p *Acb_Ntk_t) *Gia_Man_t {

	var pNew, pOne *Gia_Man_t
	var vFanins, vNodes *Vec_Int_t
	var vCopies *Vec_Int_t = Vec_IntStartFull(2 * Acb_NtkObjNum(p))
	var i int32
	var iObj int32
	var pLits []int32
	pNew = Gia_ManStart(5 * Acb_NtkObjNum(p))
	Gia_ManSetNameSpec(pNew, Abc_UtilStrsav(Acb_NtkName(p)), "")
	Gia_ManHashAlloc(pNew)
	pLits = Vec_IntArray(vCopies)[:2]
	pLits[0] = 0
	pLits[1] = 0
	for i = 0; i < Acb_NtkCiNum(p); i++ {
		iObj = Acb_NtkCi(p, i)
		pLits = Vec_IntArray(vCopies)[2*iObj : 2*iObj+2]
		pLits[0] = Gia_ManAppendCi(pNew)
		pLits[1] = 0
	}
	vFanins = Vec_IntAlloc(4)
	vNodes = Acb_NtkFindNodes2(p)
	for i = 0; i < Vec_IntSize(vNodes); i++ {
		iObj = Vec_IntEntry(vNodes, i)
		pLits = Vec_IntArray(vCopies)[2*iObj : 2*iObj+2]
		Acb_ObjToGiaDual(pNew, p, iObj, vFanins, vCopies, pLits)
	}
	Vec_IntFree(vNodes)
	Vec_IntFree(vFanins)
	for i = 0; i < Acb_NtkCoNum(p); i++ {
		iObj = Acb_NtkCo(p, i)
		pLits = Vec_IntArray(vCopies)[2*Acb_ObjFanin(p, iObj, 0):]
		Gia_ManAppendCo(pNew, pLits[0])
		Gia_ManAppendCo(pNew, pLits[1])
	}
	Vec_IntFree(vCopies)
	pOne = pNew
	pNew = Gia_ManCleanup(pOne)
	Gia_ManStop(pOne)
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
func Acb_NtkGiaDeriveMiter(pOne *Gia_Man_t, pTwo *Gia_Man_t, Type int32) *Gia_Man_t {
	var pNew, pTemp *Gia_Man_t
	var pObj *Gia_Obj_t
	var i int32
	assert(Gia_ManCiNum(pOne) == Gia_ManCiNum(pTwo), "Gia_ManCiNum(pOne) == Gia_ManCiNum(pTwo)")
	assert(Gia_ManCoNum(pOne) == Gia_ManCoNum(pTwo), "Gia_ManCoNum(pOne) == Gia_ManCoNum(pTwo)")
	pNew = Gia_ManStart(Gia_ManObjNum(pOne) + Gia_ManObjNum(pTwo) + 5*Gia_ManCoNum(pOne)/2)
	Gia_ManSetNameSpec(pNew, Abc_UtilStrsav("miter"), "")
	Gia_ManHashAlloc(pNew)
	Gia_ManConst0(pOne).Value = 0
	Gia_ManConst0(pTwo).Value = 0
	for i = 0; i < Gia_ManCiNum(pOne); i++ {
		pObj = Gia_ManCi(pOne, i)
		pObj.Value = uint32(Gia_ManAppendCi(pNew))
	}
	for i = 0; i < Gia_ManCiNum(pTwo); i++ {
		pObj = Gia_ManCi(pTwo, i)
		pObj.Value = Gia_ManCi(pOne, i).Value
	}
	for i = 0; i < Gia_ManObjNum(pOne); i++ {
		pObj = Gia_ManObj(pOne, i)
		if Gia_ObjIsAnd(pObj) != 0 {
			pObj.Value = uint32(Gia_ManHashAnd(pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj)))
		}
	}
	for i = 0; i < Gia_ManObjNum(pTwo); i++ {
		pObj = Gia_ManObj(pTwo, i)
		if Gia_ObjIsAnd(pObj) != 0 {
			pObj.Value = uint32(Gia_ManHashAnd(pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj)))
		}
	}
	for i = 0; i < Gia_ManCoNum(pOne); i++ {
		pObj = Gia_ManCo(pOne, i)
		pObj.Value = uint32(Gia_ObjFanin0Copy(pObj))
	}
	for i = 0; i < Gia_ManCoNum(pTwo); i++ {
		pObj = Gia_ManCo(pTwo, i)
		pObj.Value = uint32(Gia_ObjFanin0Copy(pObj))
	}
	if Type == 0 { // only main circuit

		for i = 0; i < Gia_ManCoNum(pOne); i += 2 {
			var pLitsF [2]int32 = [2]int32{int32(Gia_ManCo(pOne, i).Value), int32(Gia_ManCo(pOne, i+1).Value)}
			var pLitsS [2]int32 = [2]int32{int32(Gia_ManCo(pTwo, i).Value), int32(Gia_ManCo(pTwo, i+1).Value)}
			Gia_ManAppendCo(pNew, pLitsF[0])
			Gia_ManAppendCo(pNew, pLitsS[0])
		}

	} else if Type == 1 { // only shadow circuit

		for i = 0; i < Gia_ManCoNum(pOne); i += 2 {
			var pLitsF [2]int32 = [2]int32{int32(Gia_ManCo(pOne, i).Value), int32(Gia_ManCo(pOne, i+1).Value)}
			var pLitsS [2]int32 = [2]int32{int32(Gia_ManCo(pTwo, i).Value), int32(Gia_ManCo(pTwo, i+1).Value)}
			Gia_ManAppendCo(pNew, pLitsF[1])
			Gia_ManAppendCo(pNew, pLitsS[1])
		}

	} else { // comparator of the two

		for i = 0; i < Gia_ManCoNum(pOne); i += 2 {
			var pLitsF [2]int32 = [2]int32{int32(Gia_ManCo(pOne, i).Value), int32(Gia_ManCo(pOne, i+1).Value)}
			var pLitsS [2]int32 = [2]int32{int32(Gia_ManCo(pTwo, i).Value), int32(Gia_ManCo(pTwo, i+1).Value)}
			Gia_ManAppendCo(pNew, Gia_ManDualCompare(pNew, pLitsF[:], pLitsS[:]))
		}
	}
	Gia_ManHashStop(pNew)
	pTemp = pNew
	pNew = Gia_ManCleanup(pTemp)
	Gia_ManStop(pTemp)
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
func Acb_OutputFile(pFileName string, pNtkF *Acb_Ntk_t, pModel *int32) {
	var pFileName0 string
	if pFileName != "" {
		pFileName0 = pFileName
	} else {
		pFileName0 = "output"
	}
	pFile, err := os.OpenFile(pFileName0, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		fmt.Printf("Cannot open results file \"%s\".\n", pFileName0)
		return
	}
	if pModel == nil {
		fmt.Fprintf(pFile, "EQ\n")
	} else {
		/*
		   NEQ
		   in 1
		   a 1
		   b 0
		*/
		var i, iObj int32
		fmt.Fprintf(pFile, "NEQ\n")
		for i = 0; i < Acb_NtkPiNum(pNtkF); i++ {
			iObj = Acb_NtkCi(pNtkF, i)
			fmt.Fprintf(pFile, "%s %d\n", Acb_ObjNameStr(pNtkF, iObj), unsafe.Slice(pModel, int(Acb_NtkPiNum(pNtkF)))[i])
		}
	}
	_ = (pFile).Close()
	fmt.Printf("Produced output file \"%s\".\n\n", pFileName0)
}

// Acb_ProveMiter is installed by the ABCI integration package. The C source
// calls ABCI's AIG conversion and Ivy proving driver; a callback preserves
// that call chain without the forbidden Go import cycle ACB -> ABCI -> ACB.
// Applications using Acb_NtkSolve must install this dependency at startup.
var Acb_ProveMiter func(*Gia_Man_t) *int32

func Acb_NtkSolve(p *Gia_Man_t) *int32 {
	if Acb_ProveMiter == nil {
		panic("acb: ABCI miter prover is not installed")
	}
	return Acb_ProveMiter(p)
}

/*
*Function*************************************************************

	Synopsis    [Various statistics.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Acb_NtkPrintCecStats(pNtk *Acb_Ntk_t) {
	var iObj, nDcs, nMuxes int32
	// inlining of macro Acb_NtkForEachNode
	for iObj = 1; iObj < Vec_StrSize(&pNtk.vObjType); iObj++ {
		if Acb_ObjType(pNtk, iObj) == 0 || Acb_ObjIsCio(pNtk, iObj) != 0 {
		} else if Acb_ObjType(pNtk, iObj) == ABC_OPER_TRI {
			nDcs++
		} else if Acb_ObjType(pNtk, iObj) == ABC_OPER_BIT_MUX {
			nMuxes++
		}
	}

	fmt.Printf("PI = %6d  ", Acb_NtkCiNum(pNtk))
	fmt.Printf("PO = %6d  ", Acb_NtkCoNum(pNtk))
	fmt.Printf("Obj = %6d  ", Acb_NtkObjNum(pNtk))
	fmt.Printf("DC = %4d  ", nDcs)
	fmt.Printf("Mux = %4d  ", nMuxes)
	fmt.Printf("\n")
}

/*
*Function*************************************************************

	Synopsis    [Changing the PI order.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Acb_NtkUpdateCiOrder(pNtkF *Acb_Ntk_t, pNtkG *Acb_Ntk_t) {
	var i, iObj int32
	var vMap *Vec_Int_t = Vec_IntStartFull(Acb_ManNameIdMax(pNtkG.pDesign))
	var vOrder *Vec_Int_t = Vec_IntStartFull(Acb_NtkCiNum(pNtkG))
	for i = 0; i < Acb_NtkCiNum(pNtkG); i++ {
		iObj = Acb_NtkCi(pNtkG, i)
		Vec_IntWriteEntry(vMap, Acb_ObjName(pNtkG, iObj), i)
	}
	for i = 0; i < Acb_NtkCiNum(pNtkF); i++ {
		iObj = Acb_NtkCi(pNtkF, i)
		var NameIdG int32 = Acb_ManStrId(pNtkG.pDesign, Acb_ObjNameStr(pNtkF, iObj))
		var iPerm int32
		if NameIdG < Vec_IntSize(vMap) {
			iPerm = Vec_IntEntry(vMap, NameIdG)
		} else {
			iPerm = -1
		}
		if iPerm == -1 {
			fmt.Printf("Cannot find name \"%s\" of PI %d of F among PIs of G.\n", Acb_ObjNameStr(pNtkF, iObj), i)
		} else {
			Vec_IntWriteEntry(vOrder, iPerm, iObj)
		}
	}
	Vec_IntClear(&pNtkF.vCis)
	Vec_IntAppend(&pNtkF.vCis, vOrder)
	Vec_IntFree(vOrder)
	Vec_IntFree(vMap)
}
func Acb_NtkCheckPiOrder(pNtkF *Acb_Ntk_t, pNtkG *Acb_Ntk_t) int32 {
	var i int32
	var nPis int32 = Acb_NtkCiNum(pNtkF)
	for i = 0; i < nPis; i++ {
		var pNameF string = Acb_ObjNameStr(pNtkF, Acb_NtkCi(pNtkF, i))
		var pNameG string = Acb_ObjNameStr(pNtkG, Acb_NtkCi(pNtkG, i))
		if int32(strings.Compare(pNameF, pNameG)) != 0 {
			//            printf( "PI %d has different names (%s and %s) in these networks.\n", i, pNameF, pNameG );
			fmt.Printf("Networks have different PI names. Reordering PIs of the implementation network.\n")
			Acb_NtkUpdateCiOrder(pNtkF, pNtkG)
			break
		}
	}
	if i == nPis {
		fmt.Printf("Networks have the same PI names.\n")
	}
	return int32(boolInt(i == nPis, 1, 0))
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Acb_NtkRunTest(pFileNames []string, fFancy int32, fVerbose int32) {

	var fSolve int32 = 1
	var pModel *int32 = nil
	var pGiaF *Gia_Man_t = nil
	var pGiaG *Gia_Man_t = nil
	var pGia *Gia_Man_t = nil
	var pNtkF *Acb_Ntk_t = Acb_VerilogSimpleRead(pFileNames[0], "")
	var pNtkG *Acb_Ntk_t = Acb_VerilogSimpleRead(pFileNames[1], "")
	if !(pNtkF != nil) || !(pNtkG != nil) {
		return
	}

	assert(Acb_NtkCiNum(pNtkF) == Acb_NtkCiNum(pNtkG), "Acb_NtkCiNum(pNtkF) == Acb_NtkCiNum(pNtkG)")
	assert(Acb_NtkCoNum(pNtkF) == Acb_NtkCoNum(pNtkG), "Acb_NtkCoNum(pNtkF) == Acb_NtkCoNum(pNtkG)")

	Acb_NtkCheckPiOrder(pNtkF, pNtkG)
	//Acb_NtkCheckPiOrder( pNtkG, pNtkF );
	Acb_NtkPrintCecStats(pNtkF)
	Acb_NtkPrintCecStats(pNtkG)

	pGiaF = Acb_NtkGiaDeriveDual(pNtkF)
	pGiaG = Acb_NtkGiaDeriveDual(pNtkG)
	pGia = Acb_NtkGiaDeriveMiter(pGiaF, pGiaG, 2)
	//Gia_AigerWrite( pGiaF, Extra_FileNameGenericAppend(pFileNames[1], "_f2.aig"), 0, 0, 0 );
	//Gia_AigerWrite( pGiaG, Extra_FileNameGenericAppend(pFileNames[1], "_g2.aig"), 0, 0, 0 );
	//Gia_AigerWrite( pGia,  Extra_FileNameGenericAppend(pFileNames[1], "_miter_0.aig"), 0, 0, 0 );
	//printf( "Written the miter info file \"%s\".\n", Extra_FileNameGenericAppend(pFileNames[1], "_miter_0.aig") );

	//Gia_ManPrintStats( pGia, NULL );
	//Gia_ManSimTry( pGiaF, pGiaG );

	if fSolve != 0 {
		pModel = Acb_NtkSolve(pGia)
		Acb_OutputFile(pFileNames[2], pNtkF, pModel)
		pModel = nil
	}

	Gia_ManStop(pGia)
	Gia_ManStop(pGiaF)
	Gia_ManStop(pGiaG)

	Acb_ManFree(pNtkF.pDesign)
	Acb_ManFree(pNtkG.pDesign)
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END
