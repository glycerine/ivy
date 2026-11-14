package wlc

import (
	"fmt"
	"github.com/glycerine/abcgo/bool/kit"
	"github.com/glycerine/abcgo/misc/util"
	"os"
	"unsafe"
)

/**CFile****************************************************************

  FileName    [wlcBlast.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Verilog parser.]

  Synopsis    [Bit-blasting.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - August 22, 2014.]

  Revision    [$Id: wlcBlast.c,v 1.00 2014/09/12 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "wlc.h"
//#include "misc/tim/tim.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Counts constant bits.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/*
*Function*************************************************************

	Synopsis    [Helper functions.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Wlc_NtkPrepareBits(p *Wlc_Ntk_t) int32 {
	var pObj *Wlc_Obj_t
	var i, nBits int32
	Wlc_NtkCleanCopy(p)
	// inlining of macro Wlc_NtkForEachObj
	for i = 1; i < Wlc_NtkObjNumMax(p); i++ {
		pObj = Wlc_NtkObj(p, i)
		Wlc_ObjSetCopy(p, i, nBits)
		nBits += Wlc_ObjRange(pObj)
	}
	return nBits
}

func Wlc_NtkMuxTree_rec(pNew *Gia_Man_t, pCtrl []int32, nCtrl int32, vData *Vec_Int_t, Shift int32) int32 {
	var iLit0, iLit1 int32
	if nCtrl == 0 {
		return Vec_IntEntry(vData, Shift)
	}
	iLit0 = Wlc_NtkMuxTree_rec(pNew, pCtrl, nCtrl-1, vData, Shift)
	iLit1 = Wlc_NtkMuxTree_rec(pNew, pCtrl, nCtrl-1, vData, Shift+(1<<(nCtrl-1)))
	return Gia_ManHashMux(pNew, pCtrl[nCtrl-1], iLit1, iLit0)
}
func Wlc_NtkMuxTree2_nb(pNew *Gia_Man_t, pCtrl []int32, nCtrl int32, vData *Vec_Int_t, vAnds *Vec_Int_t) int32 {
	var iLitOr, iLitAnd, m int32
	assert(Vec_IntSize(vData) == (1<<nCtrl), "Vec_IntSize(vData) == (1 << nCtrl)")
	assert(Vec_IntSize(vAnds) == (1<<nCtrl), "Vec_IntSize(vAnds) == (1 << nCtrl)")
	for m = 0; m < (1 << nCtrl); m++ {
		iLitAnd = Gia_ManHashAnd(pNew, Vec_IntEntry(vAnds, m), Vec_IntEntry(vData, m))
		iLitOr = Gia_ManHashOr(pNew, iLitOr, iLitAnd)
	}
	return iLitOr
}
func Wlc_NtkMuxTree2(pNew *Gia_Man_t, pCtrl []int32, nCtrl int32, vData *Vec_Int_t, vAnds *Vec_Int_t, vTemp *Vec_Int_t) int32 {
	var m, iLit int32
	assert(!(nCtrl != 0) || Vec_IntSize(vData) == (1<<nCtrl), "!nCtrl || Vec_IntSize(vData) == (1 << nCtrl)")
	assert(!(nCtrl != 0) || Vec_IntSize(vAnds) == (1<<nCtrl), "!nCtrl || Vec_IntSize(vAnds) == (1 << nCtrl)")
	Vec_IntClear(vTemp)
	for m = 0; m < Vec_IntSize(vAnds); m++ {
		iLit = Vec_IntEntry(vAnds, m)
		Vec_IntPush(vTemp, Abc_LitNot(Gia_ManHashAnd(pNew, iLit, Vec_IntEntry(vData, m))))
	}
	return Abc_LitNot(Gia_ManHashAndMulti(pNew, vTemp))
}
func Wlc_NtkPrintNameArray(vNames *Vec_Ptr_t) {
	var i int32
	var pTemp string
	for i = 0; i < Vec_PtrSize(vNames); i++ {
		pTemp = Vec_PtrEntry(vNames, i).(string)
		fmt.Printf("%2d : %s\n", i, pTemp)
	}
}

/*
*Function*************************************************************

	Synopsis    [Bit blasting for specific operations.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Wlc_BlastShiftRightInt(pNew *Gia_Man_t, pNum []int32, nNum int32, pShift []int32, nShift int32, fSticky int32, vRes *Vec_Int_t) {
	var pRes []int32 = Wlc_VecCopy(vRes, pNum, nNum)
	var Fill int32
	if fSticky != 0 {
		Fill = pNum[nNum-1]
	} else {
		Fill = 0
	}
	var i, j, fShort int32
	assert(nShift <= 32, "nShift <= 32")
	for i = 0; i < nShift; i++ {
		for j = 0; j < nNum-fSticky; j++ {
			if (fShort != 0) || j+(1<<i) >= nNum {
				pRes[j] = Gia_ManHashMux(pNew, pShift[i], Fill, pRes[j])
				if (1 << i) > nNum {
					fShort = 1
				}

			} else {
				pRes[j] = Gia_ManHashMux(pNew, pShift[i], pRes[j+(1<<i)], pRes[j])
			}
		}
	}
}
func Wlc_BlastShiftRight(pNew *Gia_Man_t, pNum []int32, nNum int32, pShift []int32, nShift int32, fSticky int32, vRes *Vec_Int_t) {
	var nShiftMax int32 = Abc_Base2Log(nNum)
	pShiftNew := append([]int32(nil), pShift...)
	if nShiftMax < nShift && nShift > 30 {
		var i int32
		var iRes int32 = pShiftNew[nShiftMax]
		for i = nShiftMax + 1; i < nShift; i++ {
			iRes = Gia_ManHashOr(pNew, iRes, pShiftNew[i])
		}
		pShiftNew[nShiftMax] = iRes
		nShiftMax++

	} else {
		nShiftMax = nShift
	}
	Wlc_BlastShiftRightInt(pNew, pNum, nNum, pShiftNew, nShiftMax, fSticky, vRes)
}
func Wlc_BlastShiftLeftInt(pNew *Gia_Man_t, pNum []int32, nNum int32, pShift []int32, nShift int32, fSticky int32, vRes *Vec_Int_t) {
	var pRes []int32 = Wlc_VecCopy(vRes, pNum, nNum)
	var Fill int32
	if fSticky != 0 {
		Fill = pNum[0]
	} else {
		Fill = 0
	}
	var i, j, fShort int32
	assert(nShift <= 32, "nShift <= 32")
	for i = 0; i < nShift; i++ {
		for j = nNum - 1; j >= fSticky; j-- {
			if (fShort != 0) || (1<<i) > j {
				pRes[j] = Gia_ManHashMux(pNew, pShift[i], Fill, pRes[j])
				if (1 << i) > nNum {
					fShort = 1
				}

			} else {
				pRes[j] = Gia_ManHashMux(pNew, pShift[i], pRes[j-(1<<i)], pRes[j])
			}
		}
	}
}
func Wlc_BlastShiftLeft(pNew *Gia_Man_t, pNum []int32, nNum int32, pShift []int32, nShift int32, fSticky int32, vRes *Vec_Int_t) {
	var nShiftMax int32 = Abc_Base2Log(nNum)
	pShiftNew := append([]int32(nil), pShift...)
	if nShiftMax < nShift {
		var i int32
		var iRes int32 = pShiftNew[nShiftMax]
		for i = nShiftMax + 1; i < nShift; i++ {
			iRes = Gia_ManHashOr(pNew, iRes, pShiftNew[i])
		}
		pShiftNew[nShiftMax] = iRes
		nShiftMax++

	} else {
		nShiftMax = nShift
	}
	Wlc_BlastShiftLeftInt(pNew, pNum, nNum, pShiftNew, nShiftMax, fSticky, vRes)
}
func Wlc_BlastRotateRight(pNew *Gia_Man_t, pNum []int32, nNum int32, pShift []int32, nShift int32, vRes *Vec_Int_t) {
	var pRes []int32 = Wlc_VecCopy(vRes, pNum, nNum)
	var i int32
	var j int32
	pTemp := make([]int32, nNum)
	assert(nShift <= 32, "nShift <= 32")
	for i = 0; i < nShift; i++ {
		for j = 0; j < nNum; j++ {
			pTemp[j] = Gia_ManHashMux(pNew, pShift[i], pRes[(j+(1<<i))%nNum], pRes[j])
		}
		pRes = Wlc_VecCopy(vRes, pTemp, nNum)
	}
	pTemp = nil
}
func Wlc_BlastRotateLeft(pNew *Gia_Man_t, pNum []int32, nNum int32, pShift []int32, nShift int32, vRes *Vec_Int_t) {
	var pRes []int32 = Wlc_VecCopy(vRes, pNum, nNum)
	var i int32
	var j int32
	pTemp := make([]int32, nNum)
	assert(nShift <= 32, "nShift <= 32")
	for i = 0; i < nShift; i++ {
		for j = 0; j < nNum; j++ {
			var move int32
			if j >= (1 << i) {
				move = (j - (1 << i)) % nNum
			} else {
				move = (nNum - (((1 << i) - j) % nNum)) % nNum
			}
			pTemp[j] = Gia_ManHashMux(pNew, pShift[i], pRes[move], pRes[j])
			//            pTemp[j] = Gia_ManHashMux( pNew, pShift[i], pRes[((unsigned)(nNum-(1<<i)+j))%nNum], pRes[j] );
		}
		pRes = Wlc_VecCopy(vRes, pTemp, nNum)
	}
	pTemp = nil
}
func Wlc_BlastReduction(pNew *Gia_Man_t, pFans []int32, nFans int32, Type int32) int32 {
	if Type == int32(WLC_OBJ_REDUCT_AND) || Type == int32(WLC_OBJ_REDUCT_NAND) {
		var k int32
		var iLit int32 = 1
		for k = 0; k < nFans; k++ {
			iLit = Gia_ManHashAnd(pNew, iLit, pFans[k])
		}
		return Abc_LitNotCond(iLit, boolToInt32(Type == int32(WLC_OBJ_REDUCT_NAND)))
	}
	if Type == int32(WLC_OBJ_REDUCT_OR) || Type == int32(WLC_OBJ_REDUCT_NOR) {
		var k, iLit int32
		for k = 0; k < nFans; k++ {
			iLit = Gia_ManHashOr(pNew, iLit, pFans[k])
		}
		return Abc_LitNotCond(iLit, boolToInt32(Type == int32(WLC_OBJ_REDUCT_NOR)))
	}
	if Type == int32(WLC_OBJ_REDUCT_XOR) || Type == int32(WLC_OBJ_REDUCT_NXOR) {
		var k, iLit int32
		for k = 0; k < nFans; k++ {
			iLit = Gia_ManHashXor(pNew, iLit, pFans[k])
		}
		return Abc_LitNotCond(iLit, boolToInt32(Type == int32(WLC_OBJ_REDUCT_NXOR)))
	}
	assert(false, "0")
	return -1
}

func Wlc_BlastDivider2(pNew *Gia_Man_t, pNum []int32, nNum int32, pDiv []int32, nDiv int32, fQuo int32, vRes *Vec_Int_t) {
	var i int32
	pRes := Vec_IntArray(vRes)
	var k int32
	pQuo := make([]int32, nNum)
	assert(nNum > 0 && nDiv > 0, "nNum > 0 && nDiv > 0")
	assert(Vec_IntSize(vRes) < nNum+nDiv, "Vec_IntSize(vRes) < nNum + nDiv")
	for i = 0; i < nNum+nDiv; i++ {
		if i < nNum {
			pRes[i] = pNum[i]
		} else {
			pRes[i] = 0
		}
	}
	for i = nNum - 1; i >= 0; i-- {
		var Cntrl int32
		if i == nNum-1 {
			Cntrl = 1
		} else {
			Cntrl = pQuo[i+1]
		}
		var Carry int32 = Cntrl
		for k = 0; k <= nDiv; k++ {
			var divBit int32
			if k < nDiv {
				divBit = pDiv[k]
			}
			Wlc_BlastFullAdderSubtr(pNew, divBit, pRes[i+k], Carry, &Carry, &pRes[i+k], Cntrl)
		}
		pQuo[i] = Abc_LitNot(pRes[i+nDiv])
	}
	if fQuo != 0 {
		Wlc_VecCopy(vRes, pQuo, nNum)
	} else {
		var Carry, Temp int32
		for k = 0; k < nDiv; k++ {
			Wlc_BlastFullAdder(pNew, pDiv[k], pRes[k], Carry, &Carry, &Temp)
			pRes[k] = Gia_ManHashMux(pNew, pQuo[0], pRes[k], Temp)
		}
		Vec_IntShrink(vRes, nDiv)
	}
	pQuo = nil
}
func Wlc_BlastTable(pNew *Gia_Man_t, pTable *word, pFans []int32, nFans int32, nOuts int32, vRes *Vec_Int_t) {
	var vMemory []uint32
	var leaves = pFans
	table := unsafe.Slice(pTable, int(util.Abc_TtWordNum(nFans)))
	truth := make([]word, util.Abc_TtWordNum(nFans))
	var o int32
	var i int32
	var m int32
	var iLit int32
	var nMints int32 = (1 << nFans)
	Vec_IntClear(vRes)
	for o = 0; o < nOuts; o++ {
		// derive truth table
		clear(truth)
		for m = 0; m < nMints; m++ {
			for i = 0; i < nFans; i++ {
				if util.Abc_TtGetBit(table, m*nFans+i) != 0 {
					util.Abc_TtSetBit(truth, m)
				}
			}
		}
		// implement truth table
		if nFans < 6 {
			truth[0] = uint64(util.Abc_Tt6Stretch(truth[0], nFans))
		}
		iLit = kit.Kit_TruthToGia(pNew, truth32(truth), nFans, &vMemory, leaves, 1)
		Vec_IntPush(vRes, iLit)
	}
}
func Wlc_BlastLut(pNew *Gia_Man_t, Truth word, pFans []int32, nFans int32, nOuts int32, vRes *Vec_Int_t) {
	var vMemory []uint32
	leaves := pFans
	truth := []word{Truth}
	var iLit int32
	Vec_IntClear(vRes)
	assert(nOuts == 1, "nOuts == 1")
	if nFans < 6 {
		truth[0] = uint64(util.Abc_Tt6Stretch(Truth, nFans))
	}
	iLit = kit.Kit_TruthToGia(pNew, truth32(truth), nFans, &vMemory, leaves, 1)
	Vec_IntPush(vRes, iLit)
}

func truth32(words []uint64) []uint32 {
	out := make([]uint32, 2*len(words))
	for i, word := range words {
		out[2*i], out[2*i+1] = uint32(word), uint32(word>>32)
	}
	return out
}
func Wlc_BlastPower(pNew *Gia_Man_t, pNum []int32, nNum int32, pExp []int32, nExp int32, vTemp *Vec_Int_t, vRes *Vec_Int_t) {
	var vDegrees *Vec_Int_t = Vec_IntAlloc(2 * nNum)
	var vResTemp *Vec_Int_t = Vec_IntAlloc(2 * nNum)
	var i int32
	var pDegrees []int32 = nil
	pRes := Vec_IntArray(vRes)
	var k int32
	pResTemp := Vec_IntArray(vResTemp)
	Vec_IntFill(vRes, nNum, 0)
	Vec_IntWriteEntry(vRes, 0, 1)
	for i = 0; i < nExp; i++ {
		if i == 0 {
			pDegrees = Wlc_VecCopy(vDegrees, pNum, nNum)
		} else {
			Wlc_BlastMultiplier2(pNew, pDegrees, pDegrees, nNum, vTemp, vResTemp)
			pDegrees = Wlc_VecCopy(vDegrees, pResTemp, nNum)
		}
		Wlc_BlastMultiplier2(pNew, pRes, pDegrees, nNum, vTemp, vResTemp)
		for k = 0; k < nNum; k++ {
			pRes[k] = Gia_ManHashMux(pNew, pExp[i], pResTemp[k], pRes[k])
		}
	}
	Vec_IntFree(vResTemp)
	Vec_IntFree(vDegrees)
}
func Wlc_BlastSqrt(pNew *Gia_Man_t, pNum []int32, nNum int32, vTmp *Vec_Int_t, vRes *Vec_Int_t) {
	var pRes, pSum, pSumP []int32
	var i int32
	var k int32
	var Carry int32 = -1
	assert(nNum%2 == 0, "nNum % 2 == 0")
	Vec_IntFill(vRes, nNum/2, 0)
	Vec_IntFill(vTmp, 2*nNum, 0)
	pRes = Vec_IntArray(vRes)
	pSum = Vec_IntArray(vTmp)
	pSumP = pSum[nNum:]
	pNumBits := pNum
	for i = 0; i < nNum/2; i++ {
		pSumP[0] = pNumBits[nNum-2*i-2]
		pSumP[1] = pNumBits[nNum-2*i-1]
		for k = 0; k < i+1; k++ {
			pSumP[k+2] = pSum[k]
		}
		for k = 0; k < i+3; k++ {
			if k >= 2 && k < i+2 { // middle ones
				Wlc_BlastFullAdder(pNew, pSumP[k], Abc_LitNot(pRes[i-k+1]), Carry, &Carry, &pSum[k])
			} else {
				carryBit := Carry
				if k == 0 {
					carryBit = 1
				}
				Wlc_BlastFullAdder(pNew, pSumP[k], Abc_LitNot(carryBit), 1, &Carry, &pSum[k])
			}
			if k == 0 || k > i {
				Carry = Abc_LitNot(Carry)
			}
		}
		pRes[i] = Abc_LitNot(Carry)
		for k = 0; k < i+3; k++ {
			pSum[k] = Gia_ManHashMux(pNew, pRes[i], pSum[k], pSumP[k])
		}
	}
	Vec_IntReverseOrder(vRes)
}
func Wlc_BlastSqrtNR(pNew *Gia_Man_t, pNum []int32, nNum int32, vTmp *Vec_Int_t, vRes *Vec_Int_t) {
	var pSqr, pRem, pIn1 []int32
	var i int32
	var k int32
	var Carry int32 = 1
	assert(nNum%2 == 0, "nNum % 2 == 0")
	Vec_IntFill(vRes, nNum/2, 0)
	Vec_IntFill(vTmp, 2*nNum, 0)
	pSqr = Vec_IntArray(vRes)
	pRem = Vec_IntArray(vTmp)
	pIn1 = pRem[nNum:]
	pNumBits := pNum
	for i = 0; i < nNum/2; i++ {
		var SqrPrev int32 = Carry
		assert(pIn1[0] == 0, "pIn1[0] == 0")
		for k = 1; k < i; k++ {
			pIn1[k] = 0
		}
		for k = i; k < 2*i; k++ {
			pIn1[k] = pSqr[k-i]
		}
		pIn1[k] = Abc_LitNot(Carry)
		k++
		pIn1[k] = 1
		k++
		assert(k == 2*i+2, "k == 2*i+2")
		pRem[2*i+0] = pNumBits[nNum-2*i-1]
		pRem[2*i+1] = pNumBits[nNum-2*i-2]
		for k = 2*i + 1; k >= 0; k-- {
			Wlc_BlastFullAdder(pNew, Gia_ManHashXor(pNew, SqrPrev, pIn1[k]), pRem[k], Carry, &Carry, &pRem[k])
		}
		pSqr[i] = Carry
	}
	Vec_IntReverseOrder(vRes)
}
func Wlc_IntInsert(vProd *Vec_Int_t, vLevel *Vec_Int_t, Node int32, Level int32) {
	var i int32
	for i = Vec_IntSize(vLevel) - 1; i >= 0; i-- {
		if Vec_IntEntry(vLevel, i) >= Level {
			break
		}
	}
	Vec_IntInsert(vProd, i+1, Node)
	Vec_IntInsert(vLevel, i+1, Level)
}
func Wlc_BlastPrintMatrix(p *Gia_Man_t, vProds *Vec_Wec_t, fVerbose int32) {
	var vSupp *Vec_Int_t = Vec_IntAlloc(100)
	var vTemp *Vec_Wrd_t = Vec_WrdStart(Gia_ManObjNum(p))
	var vLevel *Vec_Int_t
	var Truth word
	var i, k, iLit int32
	for i = 0; i < Vec_WecSize(vProds); i++ {
		vLevel = Vec_WecEntry(vProds, i)
		for k = 0; k < Vec_IntSize(vLevel); k++ {
			iLit = Vec_IntEntry(vLevel, k)
			if Gia_ObjIsAnd(Gia_ManObj(p, Abc_Lit2Var(iLit))) != 0 {
				Vec_IntPushUnique(vSupp, Abc_Lit2Var(iLit))
			}
		}
	}
	fmt.Printf("Booth partial products: %d pps, %d unique, %d nodes.\n", Vec_WecSizeSize(vProds), Vec_IntSize(vSupp), Gia_ManAndNum(p))
	Vec_IntPrint(vSupp)

	if fVerbose != 0 {
		for i = 0; i < Vec_WecSize(vProds); i++ {
			vLevel = Vec_WecEntry(vProds, i)
			for k = 0; k < Vec_IntSize(vLevel); k++ {
				iLit = Vec_IntEntry(vLevel, k)
				fmt.Printf("Obj = %4d : ", Abc_Lit2Var(iLit))
				fmt.Printf("Compl = %d  ", Abc_LitIsCompl(iLit))
				fmt.Printf("Rank = %2d  ", i)
				Truth = Gia_ObjComputeTruth6Cis(p, iLit, vSupp, vTemp)
				Extra_PrintHex(os.Stdout, &Truth, Vec_IntSize(vSupp))
				if Vec_IntSize(vSupp) == 4 {
					fmt.Printf("    ")
				}
				if Vec_IntSize(vSupp) == 3 {
					fmt.Printf("      ")
				}
				if Vec_IntSize(vSupp) <= 2 {
					fmt.Printf("       ")
				}
				fmt.Printf("  ")
				Vec_IntPrint(vSupp)
				if k == Vec_IntSize(vLevel)-1 {
					fmt.Printf("\n")
				}
			}
		}
	}
	Vec_IntFree(vSupp)
	Vec_WrdFree(vTemp)
}
func Wlc_BlastReduceMatrix(pNew *Gia_Man_t, vProds *Vec_Wec_t, vLevels *Vec_Wec_t, vRes *Vec_Int_t, fSigned int32, fCla int32) {
	var vLevel, vProd *Vec_Int_t
	var i, NodeS, NodeC, LevelS, LevelC, Node1, Node2, Node3, Level1, Level2, Level3 int32
	var nSize int32 = Vec_WecSize(vProds)
	assert(nSize == Vec_WecSize(vLevels), "nSize == Vec_WecSize(vLevels)")
	for i = 0; i < nSize; i++ {
		for {
			vProd = Vec_WecEntry(vProds, i)
			if Vec_IntSize(vProd) < 3 {
				break
			}

			Node1 = Vec_IntPop(vProd)
			Node2 = Vec_IntPop(vProd)
			Node3 = Vec_IntPop(vProd)

			vLevel = Vec_WecEntry(vLevels, i)

			Level1 = Vec_IntPop(vLevel)
			Level2 = Vec_IntPop(vLevel)
			Level3 = Vec_IntPop(vLevel)

			Wlc_BlastFullAdder(pNew, Node1, Node2, Node3, &NodeC, &NodeS)
			LevelS = Abc_MaxInt(Abc_MaxInt(Level1, Level2), Level3) + 2
			LevelC = LevelS - 1

			Wlc_IntInsert(vProd, vLevel, NodeS, LevelS)

			vProd = Vec_WecEntry(vProds, i+1)
			vLevel = Vec_WecEntry(vLevels, i+1)

			Wlc_IntInsert(vProd, vLevel, NodeC, LevelC)
		}
	}

	// make all ranks have two products
	for i = 0; i < nSize; i++ {
		vProd = Vec_WecEntry(vProds, i)
		for Vec_IntSize(vProd) < 2 {
			Vec_IntPush(vProd, 0)
		}
		assert(Vec_IntSize(vProd) == 2, "Vec_IntSize(vProd) == 2")
	}
	//    Vec_WecPrint( vProds, 0 );

	vLevel = Vec_WecEntry(vLevels, 0)
	Vec_IntClear(vRes)
	Vec_IntClear(vLevel)
	for i = 0; i < nSize; i++ {
		vProd = Vec_WecEntry(vProds, i)
		Vec_IntPush(vRes, Vec_IntEntry(vProd, 0))
		Vec_IntPush(vLevel, Vec_IntEntry(vProd, 1))
	}
	Vec_IntPush(vRes, 0)
	Vec_IntPush(vLevel, 0)

	if fCla != 0 {
		Wlc_BlastAdderCLA(pNew, Vec_IntArray(vRes), Vec_IntArray(vLevel), Vec_IntSize(vRes), fSigned, 0)
	} else {
		Wlc_BlastAdder(pNew, Vec_IntArray(vRes), Vec_IntArray(vLevel), Vec_IntSize(vRes), 0)
	}
}

func Wlc_BlastAddLevel(pNew *Gia_Man_t, Start int32) int32 {
	var i int32
	if Start == 0 {
		Gia_ManCleanLevels(pNew, 5*Gia_ManObjNum(pNew))
	}
	for i = Start; i < Gia_ManObjNum(pNew); i++ {
		var pObj *Gia_Obj_t = Gia_ManObj(pNew, i)
		if Gia_ObjIsAnd(pObj) != 0 {
			Gia_ObjSetAndLevel(pNew, pObj)
		}
	}
	return Gia_ManObjNum(pNew)
}
func Wlc_IntInsert2(pNew *Gia_Man_t, vProd *Vec_Int_t, iLit int32) {
	var i int32
	var Entry int32
	var Level int32 = Gia_ObjLevelId(pNew, Abc_Lit2Var(iLit))
	for i = Vec_IntSize(vProd) - 1; i >= 0; i-- {
		Entry = Vec_IntEntry(vProd, i)
		if Gia_ObjLevelId(pNew, Abc_Lit2Var(Entry)) >= Level {
			break
		}
	}
	Vec_IntInsert(vProd, i+1, iLit)
	// Vec_IntForEachEntry( vProd, Entry, i )
	//
	//	printf( "Obj=%d(%d) ", Abc_Lit2Var(Entry), Gia_ObjLevelId(pNew, Abc_Lit2Var(Entry)) );
	//
	// printf( "\n" );
}
func Wlc_IntSortCostReverse(pNew *Gia_Man_t, pArray []int32, nSize int32) {
	var i, j, best_i int32
	for i = 0; i < nSize-1; i++ {
		best_i = i
		for j = i + 1; j < nSize; j++ {
			if Gia_ObjLevelId(pNew, Abc_Lit2Var(pArray[j])) > Gia_ObjLevelId(pNew, Abc_Lit2Var(pArray[best_i])) {
				best_i = j
			}
		}
		pArray[i], pArray[best_i] = pArray[best_i], pArray[i]
	}
}
func Wlc_BlastReduceMatrix2(pNew *Gia_Man_t, vProds *Vec_Wec_t, vRes *Vec_Int_t, fSigned int32, fCla int32) {
	var vProd, vTemp *Vec_Int_t
	var i, NodeS, NodeC, Node1, Node2, Node3 int32
	var Start int32 = Wlc_BlastAddLevel(pNew, 0)
	var nSize int32 = Vec_WecSize(vProds)
	for i = 0; i < nSize; i++ {
		vProd = Vec_WecEntry(vProds, i)
		Wlc_IntSortCostReverse(pNew, Vec_IntArray(vProd), Vec_IntSize(vProd))
	}
	for i = 0; i < nSize; i++ {
		for {
			vProd = Vec_WecEntry(vProds, i)
			if Vec_IntSize(vProd) < 3 {
				break
			}

			Node1 = Vec_IntPop(vProd)
			Node2 = Vec_IntPop(vProd)
			Node3 = Vec_IntPop(vProd)

			assert(Gia_ObjLevelId(pNew, Abc_Lit2Var(Node3)) >= Gia_ObjLevelId(pNew, Abc_Lit2Var(Node2)), "Gia_ObjLevelId(pNew, Abc_Lit2Var(Node3)) >= Gia_ObjLevelId(pNew, Abc_Lit2Var(Node2))")
			assert(Gia_ObjLevelId(pNew, Abc_Lit2Var(Node2)) >= Gia_ObjLevelId(pNew, Abc_Lit2Var(Node1)), "Gia_ObjLevelId(pNew, Abc_Lit2Var(Node2)) >= Gia_ObjLevelId(pNew, Abc_Lit2Var(Node1))")

			Wlc_BlastFullAdder(pNew, Node1, Node2, Node3, &NodeC, &NodeS)
			Start = Wlc_BlastAddLevel(pNew, Start)

			Wlc_IntInsert2(pNew, vProd, NodeS)

			vProd = Vec_WecEntry(vProds, i+1)
			Wlc_IntInsert2(pNew, vProd, NodeC)
		}
	}

	// make all ranks have two products
	for i = 0; i < nSize; i++ {
		vProd = Vec_WecEntry(vProds, i)
		for Vec_IntSize(vProd) < 2 {
			Vec_IntPush(vProd, 0)
		}
		assert(Vec_IntSize(vProd) == 2, "Vec_IntSize(vProd) == 2")
	}
	//    Vec_WecPrint( vProds, 0 );

	Vec_IntClear(vRes)
	vTemp = Vec_IntAlloc(100)
	for i = 0; i < nSize; i++ {
		vProd = Vec_WecEntry(vProds, i)
		Vec_IntPush(vRes, Vec_IntEntry(vProd, 0))
		Vec_IntPush(vTemp, Vec_IntEntry(vProd, 1))
	}
	Vec_IntPush(vRes, 0)
	Vec_IntPush(vTemp, 0)

	if fCla != 0 {
		Wlc_BlastAdderCLA(pNew, Vec_IntArray(vRes), Vec_IntArray(vTemp), Vec_IntSize(vRes), fSigned, 0)
	} else {
		Wlc_BlastAdder(pNew, Vec_IntArray(vRes), Vec_IntArray(vTemp), Vec_IntSize(vRes), 0)
	}
	Vec_IntFree(vTemp)
}

func Wlc_BlastMultiplier3(pNew *Gia_Man_t, pArgA []int32, pArgB []int32, nArgA int32, nArgB int32, vRes *Vec_Int_t, fSigned int32, fCla int32) {
	var vProds *Vec_Wec_t = Vec_WecStart(nArgA + nArgB)
	var vLevels *Vec_Wec_t = Vec_WecStart(nArgA + nArgB)
	var i, k int32
	for i = 0; i < nArgA; i++ {
		for k = 0; k < nArgB; k++ {
			fCompl := boolToInt32(fSigned != 0 && ((i == nArgA-1) != (k == nArgB-1)))
			Vec_WecPush(vProds, i+k, Abc_LitNotCond(Gia_ManHashAnd(pNew, pArgA[i], pArgB[k]), fCompl))
			Vec_WecPush(vLevels, i+k, 0)
		}
	}
	if fSigned != 0 {
		Vec_WecPush(vProds, nArgA, 1)
		Vec_WecPush(vLevels, nArgA, 0)

		Vec_WecPush(vProds, nArgA+nArgB-1, 1)
		Vec_WecPush(vLevels, nArgA+nArgB-1, 0)
	}

	Wlc_BlastReduceMatrix(pNew, vProds, vLevels, vRes, fSigned, fCla)
	//    Wlc_BlastReduceMatrix2( pNew, vProds, vRes, fSigned, fCla );

	Vec_WecFree(vProds)
	Vec_WecFree(vLevels)
}
func Wlc_BlastSquare(pNew *Gia_Man_t, pNum []int32, nNum int32, vTmp *Vec_Int_t, vRes *Vec_Int_t) {
	var vProds *Vec_Wec_t = Vec_WecStart(2 * nNum)
	var vLevels *Vec_Wec_t = Vec_WecStart(2 * nNum)
	var i, k int32
	for i = 0; i < nNum; i++ {
		for k = 0; k < nNum; k++ {
			if i == k {
				Vec_WecPush(vProds, i+k, pNum[i])
				Vec_WecPush(vLevels, i+k, 0)

			} else if i < k {
				Vec_WecPush(vProds, i+k+1, Gia_ManHashAnd(pNew, pNum[i], pNum[k]))
				Vec_WecPush(vLevels, i+k+1, 0)
			}
		}
	}

	Wlc_BlastReduceMatrix(pNew, vProds, vLevels, vRes, 0, 0)

	Vec_WecFree(vProds)
	Vec_WecFree(vLevels)
}
func Wlc_BlastDecoder(pNew *Gia_Man_t, pNum []int32, nNum int32, vTmp *Vec_Int_t, vRes *Vec_Int_t) {
	var i int32
	var k int32
	var nMints int32 = 1 << nNum
	Vec_IntClear(vRes)
	for i = 0; i < nMints; i++ {
		var iMint int32 = 1
		for k = 0; k < nNum; k++ {
			iMint = Gia_ManHashAnd(pNew, iMint, Abc_LitNotCond(pNum[k], boolToInt32(((i>>k)&1) == 0)))
		}
		Vec_IntPush(vRes, iMint)
	}
}
func Wlc_BlastBooth(pNew *Gia_Man_t, pArgA []int32, pArgB []int32, nArgA int32, nArgB int32, vRes *Vec_Int_t, fSigned int32, fCla int32) {
	var vProds *Vec_Wec_t = Vec_WecStart(nArgA + nArgB + 3)
	var vLevels *Vec_Wec_t = Vec_WecStart(nArgA + nArgB + 3)
	var FillA int32
	if fSigned != 0 {
		FillA = pArgA[nArgA-1]
	} else {
		FillA = 0
	}
	var FillB int32
	if fSigned != 0 {
		FillB = pArgB[nArgB-1]
	} else {
		FillB = 0
	}
	var i, k, Sign int32

	// extend argument B
	var vArgB *Vec_Int_t = Vec_IntAlloc(nArgB + 2)
	Vec_IntPush(vArgB, 0)
	for i = 0; i < nArgB; i++ {
		Vec_IntPush(vArgB, pArgB[i])
	}
	if !(fSigned != 0) {
		Vec_IntPushTwo(vArgB, FillB, FillB)
	}
	if Vec_IntSize(vArgB)%2 == 0 {
		Vec_IntPush(vArgB, FillB)
	}
	assert(Vec_IntSize(vArgB)%2 == 1, "Vec_IntSize(vArgB) % 2 == 1")

	// iterate through bit-pairs of B
	for k = 0; k+2 < Vec_IntSize(vArgB); k += 2 {
		var pp int32 = -1
		var Q2jM1 int32 = Vec_IntEntry(vArgB, k)   // q(2*j-1)
		var Q2j int32 = Vec_IntEntry(vArgB, k+1)   // q(2*j+0)
		var Q2jP1 int32 = Vec_IntEntry(vArgB, k+2) // q(2*j+1)
		var Neg int32 = Q2jP1
		var One int32 = Gia_ManHashXor(pNew, Q2j, Q2jM1)
		var Two int32 = Gia_ManHashMux(pNew, Neg, Gia_ManHashAnd(pNew, Abc_LitNot(Q2j), Abc_LitNot(Q2jM1)), Gia_ManHashAnd(pNew, Q2j, Q2jM1))
		for i = 0; i <= nArgA; i++ {
			var This int32
			if i == nArgA {
				This = FillA
			} else {
				This = pArgA[i]
			}
			var Prev int32
			if i != 0 {
				Prev = pArgA[i-1]
			} else {
				Prev = 0
			}
			var Part int32 = Gia_ManHashOr(pNew, Gia_ManHashAnd(pNew, One, This), Gia_ManHashAnd(pNew, Two, Prev))

			pp = Gia_ManHashXor(pNew, Part, Neg)
			if pp == 0 || ((fSigned != 0) && i == nArgA) {
				continue
			}

			Vec_WecPush(vProds, k+i, pp)
			Vec_WecPush(vLevels, k+i, 0)
		}
		if fSigned != 0 {
			i--
		}
		// perform sign extension
		if fSigned != 0 {
			Sign = pp
		} else {
			Sign = Neg
		}
		if k == 0 {
			Vec_WecPush(vProds, k+i, Sign)
			Vec_WecPush(vLevels, k+i, 0)

			Vec_WecPush(vProds, k+i+1, Sign)
			Vec_WecPush(vLevels, k+i+1, 0)

			Vec_WecPush(vProds, k+i+2, Abc_LitNot(Sign))
			Vec_WecPush(vLevels, k+i+2, 0)

		} else {
			Vec_WecPush(vProds, k+i, Abc_LitNot(Sign))
			Vec_WecPush(vLevels, k+i, 0)

			Vec_WecPush(vProds, k+i+1, 1)
			Vec_WecPush(vLevels, k+i+1, 0)
		}
		// add neg to the first column
		if Neg == 0 {
			continue
		}
		Vec_WecPush(vProds, k, Neg)
		Vec_WecPush(vLevels, k, 0)
	}
	//Vec_WecPrint( vProds, 0 );
	//Wlc_BlastPrintMatrix( pNew, vProds, 1 );
	//printf( "Cutoff ID for partial products = %d.\n", Gia_ManObjNum(pNew) );
	Wlc_BlastReduceMatrix(pNew, vProds, vLevels, vRes, fSigned, fCla)
	//    Wlc_BlastReduceMatrix2( pNew, vProds, vRes, fSigned, fCla );

	Vec_WecFree(vProds)
	Vec_WecFree(vLevels)
	Vec_IntFree(vArgB)
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Wlc_NtkBitBlast(p *Wlc_Ntk_t, pParIn *Wlc_BstPar_t) *Gia_Man_t {
	var fVerbose int32
	var fUseOldMultiplierBlasting int32
	var fSkipBitRange int32
	var pManTime *Tim_Man_t = nil
	var pBoxLib *If_LibBox_t = nil
	var vTables *Vec_Ptr_t = nil
	var vFf2Ci *Vec_Int_t = Vec_IntAlloc(100)
	var vRegClasses *Vec_Int_t = nil
	var pTemp *Gia_Man_t
	var pNew *Gia_Man_t
	var pExtra *Gia_Man_t = nil
	var pObj, pObj2 *Wlc_Obj_t
	var vBits *Vec_Int_t = &p.vBits
	var vTemp0 *Vec_Int_t
	var vTemp1 *Vec_Int_t
	var vTemp2 *Vec_Int_t
	var vRes *Vec_Int_t
	var vAddOutputs *Vec_Int_t = nil
	var vAddObjs *Vec_Int_t = nil
	var nBits int32 = Wlc_NtkPrepareBits(p)
	var nRange, nRange0, nRange1, nRange2, nRange3 int32
	var i int32
	var k int32
	var b int32
	var iFanin int32
	var iLit int32
	var nAndPrev int32
	var pFans0 []int32
	var pFans1 []int32
	var pFans2 []int32
	var pFans3 []int32
	var nFFins, nFFouts, curPi, curPo, nFf2Regs int32
	var nBitCis, nBitCos, fAdded int32
	var Par Wlc_BstPar_t
	var pPar *Wlc_BstPar_t = &Par
	Wlc_BstParDefault(pPar)
	if pParIn != nil {
		pPar = pParIn
	} else {
		pPar = pPar
	}
	Vec_IntClear(vBits)
	Vec_IntGrow(vBits, nBits)
	vTemp0 = Vec_IntAlloc(1000)
	vTemp1 = Vec_IntAlloc(1000)
	vTemp2 = Vec_IntAlloc(1000)
	vRes = Vec_IntAlloc(1000)
	// clean AND-gate counters
	clear(p.nAnds[:])
	// create AIG manager
	pNew = Gia_ManStart(5*Wlc_NtkObjNum(p) + 1000)
	Gia_ManSetNameSpec(pNew, Abc_UtilStrsav(p.pName), "")
	Gia_ManSetSimple(pNew, pPar.fGiaSimple)
	if pPar.fGiaSimple == 0 {
		Gia_ManHashAlloc(pNew)
	}
	if pPar.fAddOutputs != 0 {
		vAddOutputs = Vec_IntAlloc(100)
	}
	if pPar.fAddOutputs != 0 {
		vAddObjs = Vec_IntAlloc(100)
	}
	// prepare for AIG with boxes
	if pPar.vBoxIds != nil {
		var nNewCis, nNewCos int32
		assert(Vec_IntSize(&p.vFfs2) == 0, "Vec_IntSize(&p.vFfs2) == 0")
		// inlining of macro Wlc_NtkForEachObj
		for i = 1; i < Wlc_NtkObjNumMax(p); i++ {
			pObj = Wlc_NtkObj(p, i)
			pObj.Mark = 0
		}
		// count bit-width of regular CIs/COs
		// inlining of macro Wlc_NtkForEachCi
		for i = 0; i < Wlc_NtkCiNum(p); i++ {
			pObj = Wlc_NtkCi(p, i)
			nBitCis += Wlc_ObjRange(pObj)
		}
		// inlining of macro Wlc_NtkForEachCo
		for i = 0; i < Wlc_NtkCoNum(p); i++ {
			pObj = Wlc_NtkCo(p, i)
			nBitCos += Wlc_ObjRange(pObj)
		}
		// count bit-width of additional CIs/COs due to selected multipliers
		assert(Vec_IntSize(pPar.vBoxIds) > 0, "Vec_IntSize(pPar.vBoxIds) > 0")
		// inlining of macro Wlc_NtkForEachObjVec
		for i = 0; i < Vec_IntSize(pPar.vBoxIds); i++ {
			pObj = Wlc_NtkObj(p, Vec_IntEntry(pPar.vBoxIds, i))
			// currently works only for multipliers
			assert(pObj.Type == uint32(WLC_OBJ_ARI_MULTI) || pObj.Type == uint32(WLC_OBJ_ARI_ADD), "pObj.Type == uint32(WLC_OBJ_ARI_MULTI) || pObj.Type == uint32(WLC_OBJ_ARI_ADD)")
			nNewCis += Wlc_ObjRange(pObj)
			nNewCos += Wlc_ObjRange(Wlc_ObjFanin0(p, pObj))
			nNewCos += Wlc_ObjRange(Wlc_ObjFanin1(p, pObj))
			if Wlc_ObjFaninNum(pObj) > 2 {
				nNewCos += Wlc_ObjRange(Wlc_ObjFanin2(p, pObj))
			}
			pObj.Mark = 1
		}
		// create hierarchy manager
		pManTime = Tim_ManStart(nBitCis+nNewCis, nBitCos+nNewCos)
		curPi = nBitCis
		curPo = 0
		// create AIG manager for logic of the boxes
		pExtra = Gia_ManStart(Wlc_NtkObjNum(p))
		Gia_ManHashAlloc(pExtra)
		assert(pPar.fGiaSimple == 0, "pPar.fGiaSimple == 0")
		// create box library
		pBoxLib = If_LibBoxStart()
	}
	if Vec_IntSize(&p.vFfs2) > 0 {
		var vSignature *Vec_Int_t
		var nNewCis, nNewCos int32
		assert(pPar.vBoxIds == nil, "pPar.vBoxIds == nil")
		// count bit-width of regular CIs/COs
		// inlining of macro Wlc_NtkForEachCi
		for i = 0; i < Wlc_NtkCiNum(p); i++ {
			pObj = Wlc_NtkCi(p, i)
			nBitCis += Wlc_ObjRange(pObj)
		}
		// inlining of macro Wlc_NtkForEachCo
		for i = 0; i < Wlc_NtkCoNum(p); i++ {
			pObj = Wlc_NtkCo(p, i)
			nBitCos += Wlc_ObjRange(pObj)
		}
		// count bit-width of additional CIs/COs due to selected multipliers
		// inlining of macro Wlc_NtkForEachFf2
		for i = 0; i < Vec_IntSize(&p.vFfs2); i++ {
			pObj = Wlc_NtkFf2(p, i)
			// currently works only for multipliers
			assert(pObj.Type == uint32(WLC_OBJ_FF), "pObj.Type == uint32(WLC_OBJ_FF)")
			assert(Wlc_ObjRange(pObj) == Wlc_ObjRange(Wlc_ObjFanin0(p, pObj)), "Wlc_ObjRange(pObj) == Wlc_ObjRange(Wlc_ObjFanin0(p, pObj))")
			nNewCis += Wlc_ObjRange(pObj)
			nNewCos += 2*Wlc_ObjRange(pObj) + 3
			nFf2Regs += Wlc_ObjRange(pObj)
		}
		// create hierarchy manager
		pManTime = Tim_ManStart(nBitCis+nNewCis+nFf2Regs, nBitCos+nNewCos+nFf2Regs)
		curPi = nBitCis + nFf2Regs
		curPo = 0
		// create AIG manager for logic of the boxes
		pExtra = Gia_ManStart(Wlc_NtkObjNum(p))
		Gia_ManHashAlloc(pExtra)
		assert(pPar.fGiaSimple == 0, "pPar.fGiaSimple == 0")
		// create register classes
		vRegClasses = Vec_IntAlloc(0)
		vSignature = Vec_IntAlloc(100)
		Vec_IntPushTwo(vSignature, -1, -1)
		// inlining of macro Wlc_NtkForEachFf2
		for i = 0; i < Vec_IntSize(&p.vFfs2); i++ {
			pObj = Wlc_NtkFf2(p, i)
			var iClk0 int32
			var iClk int32 = Wlc_ObjFaninId(pObj, 1)
			var iAsyn0 int32
			var iAsyn int32 = Wlc_ObjFaninId(pObj, 5)
			nRange = Wlc_ObjRange(pObj)
			for k = 0; k < Vec_IntSize(vSignature); k += 2 {
				iClk0 = Vec_IntEntry(vSignature, k)
				iAsyn0 = Vec_IntEntry(vSignature, k+1)
				if iClk == iClk0 && iAsyn == iAsyn0 {
					for b = 0; b < nRange; b++ {
						Vec_IntPush(vRegClasses, k/2)
					}
					break
				}
				if k < Vec_IntSize(vSignature) {
					continue
				}
				for b = 0; b < nRange; b++ {
					Vec_IntPush(vRegClasses, k/2)
				}
				Vec_IntPushTwo(vSignature, iClk, iAsyn)
			}
		}
		assert(Vec_IntSize(vRegClasses) == nFf2Regs, "Vec_IntSize(vRegClasses) == nFf2Regs")
		Vec_IntFree(vSignature)
		// create box library
		pBoxLib = If_LibBoxStart()
	}
	//printf( "Init state: %s\n", p->pInits );

	// blast in the topological order
	// inlining of macro Wlc_NtkForEachObj
	for i = 1; i < Wlc_NtkObjNumMax(p); i++ {
		pObj = Wlc_NtkObj(p, i)
		//char * pName1 = Wlc_ObjName(p, i);
		//char * pName2 = Wlc_ObjFaninNum(pObj) ? Wlc_ObjName(p, Wlc_ObjFaninId0(pObj)) : NULL;

		nAndPrev = Gia_ManAndNum(pNew)
		nRange = Wlc_ObjRange(pObj)
		if Wlc_ObjFaninNum(pObj) > 0 {
			nRange0 = Wlc_ObjRange(Wlc_ObjFanin0(p, pObj))
		} else {
			nRange0 = -1
		}
		if Wlc_ObjFaninNum(pObj) > 1 {
			nRange1 = Wlc_ObjRange(Wlc_ObjFanin1(p, pObj))
		} else {
			nRange1 = -1
		}
		if Wlc_ObjFaninNum(pObj) > 2 {
			nRange2 = Wlc_ObjRange(Wlc_ObjFanin2(p, pObj))
		} else {
			nRange2 = -1
		}
		if Wlc_ObjFaninNum(pObj) > 3 {
			nRange3 = Wlc_ObjRange(Wlc_ObjFanin(p, pObj, 3))
		} else {
			nRange3 = -1
		}
		if pObj.Type != uint32(WLC_OBJ_FF) && Wlc_ObjFaninNum(pObj) > 0 {
			pFans0 = Vec_IntArray(vBits)[Wlc_ObjCopy(p, Wlc_ObjFaninId0(pObj)):]
		} else {
			pFans0 = nil
		}
		if pObj.Type != uint32(WLC_OBJ_FF) && Wlc_ObjFaninNum(pObj) > 1 {
			pFans1 = Vec_IntArray(vBits)[Wlc_ObjCopy(p, Wlc_ObjFaninId1(pObj)):]
		} else {
			pFans1 = nil
		}
		if pObj.Type != uint32(WLC_OBJ_FF) && Wlc_ObjFaninNum(pObj) > 2 {
			pFans2 = Vec_IntArray(vBits)[Wlc_ObjCopy(p, Wlc_ObjFaninId2(pObj)):]
		} else {
			pFans2 = nil
		}
		if pObj.Type != uint32(WLC_OBJ_FF) && Wlc_ObjFaninNum(pObj) > 3 {
			pFans3 = Vec_IntArray(vBits)[Wlc_ObjCopy(p, Wlc_ObjFaninId(pObj, 3)):]
		} else {
			pFans3 = nil
		}
		Vec_IntClear(vRes)
		assert(nRange > 0, "nRange > 0")
		if pPar.vBoxIds != nil && pObj.Mark != 0 {
			var pBox *If_Box_t
			var Buffer string // was [100]byte in C

			var pTable []float32
			var CarryIn int32

			pObj.Mark = 0
			assert(pObj.Type == uint32(WLC_OBJ_ARI_MULTI) || pObj.Type == uint32(WLC_OBJ_ARI_ADD) || pObj.Type == uint32(WLC_OBJ_ARI_SUB), "pObj.Type == uint32(WLC_OBJ_ARI_MULTI) || pObj.Type == uint32(WLC_OBJ_ARI_ADD) || pObj.Type == uint32(WLC_OBJ_ARI_SUB)")

			// account for carry-in
			if Wlc_ObjFaninNum(pObj) == 3 {
				assert(nRange2 == 1, "nRange2 == 1")
			} else {
				nRange2 = 0
			}

			// create new box
			if vTables == nil {
				vTables = Vec_PtrAlloc(100)
				Tim_ManSetDelayTables(pManTime, vTables)
			}
			Tim_ManCreateBox(pManTime, curPo, nRange0+nRange1+nRange2, curPi, nRange, Vec_PtrSize(vTables), 0)
			curPi += nRange
			curPo += nRange0 + nRange1 + nRange2

			// create delay table
			pTable = make([]float32, 3+int(nRange*(nRange0+nRange1+nRange2)))
			pTable[0] = float32(Vec_PtrSize(vTables))
			pTable[1] = float32(nRange0 + nRange1 + nRange2)
			pTable[2] = float32(nRange)
			for k = 0; k < nRange*(nRange0+nRange1+nRange2); k++ {
				pTable[3+k] = 1.0
			}
			Vec_PtrPush(vTables, pTable)

			// create combinational outputs in the normal manager
			for k = 0; k < nRange0; k++ {
				Gia_ManAppendCo(pNew, pFans0[k])
			}
			for k = 0; k < nRange1; k++ {
				Gia_ManAppendCo(pNew, pFans1[k])
			}
			for k = 0; k < nRange2; k++ {
				Gia_ManAppendCo(pNew, pFans2[0])
			}

			// make sure there is enough primary inputs in the manager
			for k = Gia_ManPiNum(pExtra); k < nRange0+nRange1+nRange2; k++ {
				Gia_ManAppendCi(pExtra)
			}
			// create combinational inputs
			Vec_IntClear(vTemp0)
			for k = 0; k < nRange0; k++ {
				Vec_IntPush(vTemp0, Gia_Obj2Lit(pExtra, Gia_ManPi(pExtra, k)))
			}
			Vec_IntClear(vTemp1)
			for k = 0; k < nRange1; k++ {
				Vec_IntPush(vTemp1, Gia_Obj2Lit(pExtra, Gia_ManPi(pExtra, nRange0+k)))
			}
			if nRange2 == 1 {
				CarryIn = Gia_Obj2Lit(pExtra, Gia_ManPi(pExtra, nRange0+nRange1))
			}

			// get new fanin arrays
			pFans0 = Vec_IntArray(vTemp0)
			pFans1 = Vec_IntArray(vTemp1)

			// bit-blast in the external manager
			if pObj.Type == uint32(WLC_OBJ_ARI_ADD) || pObj.Type == uint32(WLC_OBJ_ARI_SUB) {
				var nRangeMax int32 = Abc_MaxInt(nRange, Abc_MaxInt(nRange0, nRange1))
				var pArg0 []int32 = Wlc_VecLoadFanins(vRes, pFans0, nRange0, nRangeMax, Wlc_ObjIsSignedFanin01(p, pObj))
				var pArg1 []int32 = Wlc_VecLoadFanins(vTemp1, pFans1, nRange1, nRangeMax, Wlc_ObjIsSignedFanin01(p, pObj))
				if pObj.Type == uint32(WLC_OBJ_ARI_ADD) {
					Wlc_BlastAdder(pExtra, pArg0, pArg1, nRange, CarryIn) // result is in pFan0 (vRes)
				} else {
					Wlc_BlastSubtract(pExtra, pArg0, pArg1, nRange, 1)
				} // result is in pFan0 (vRes)
				Vec_IntShrink(vRes, nRange)

			} else if fUseOldMultiplierBlasting != 0 {
				var nRangeMax int32 = Abc_MaxInt(nRange, Abc_MaxInt(nRange0, nRange1))
				var pArg0 []int32 = Wlc_VecLoadFanins(vTemp0, pFans0, nRange0, nRangeMax, Wlc_ObjIsSignedFanin01(p, pObj))
				var pArg1 []int32 = Wlc_VecLoadFanins(vTemp1, pFans1, nRange1, nRangeMax, Wlc_ObjIsSignedFanin01(p, pObj))
				Wlc_BlastMultiplier2(pExtra, pArg0, pArg1, nRange, vTemp2, vRes)
				Vec_IntShrink(vRes, nRange)

			} else {
				var fSigned int32 = Wlc_ObjIsSignedFanin01(p, pObj)
				var nRangeMax int32 = Abc_MaxInt(nRange0, nRange1)
				var pArg0 []int32 = Wlc_VecLoadFanins(vTemp0, pFans0, nRange0, nRangeMax, fSigned)
				var pArg1 []int32 = Wlc_VecLoadFanins(vTemp1, pFans1, nRange1, nRangeMax, fSigned)
				Wlc_BlastMultiplier(pExtra, pArg0, pArg1, nRangeMax, nRangeMax, vTemp2, vRes, fSigned)
				if nRange > nRangeMax+nRangeMax {
					fill := int32(0)
					if fSigned != 0 {
						fill = Vec_IntEntryLast(vRes)
					}
					Vec_IntFillExtra(vRes, nRange, fill)
				} else {
					Vec_IntShrink(vRes, nRange)
				}
				assert(Vec_IntSize(vRes) == nRange, "Vec_IntSize(vRes) == nRange")
			}
			// create outputs in the external manager
			for k = 0; k < nRange; k++ {
				Gia_ManAppendCo(pExtra, Vec_IntEntry(vRes, k))
			}

			// create combinational inputs in the normal manager
			Vec_IntClear(vRes)
			for k = 0; k < nRange; k++ {
				Vec_IntPush(vRes, Gia_ManAppendCi(pNew))
			}

			// add box to the library
			opName := "mul"
			if pObj.Type == uint32(WLC_OBJ_ARI_ADD) {
				opName = "add"
			}
			Buffer = fmt.Sprintf("%s%03d", opName, 1+If_LibBoxNum(pBoxLib))
			pBox = If_BoxStart(Abc_UtilStrsav(Buffer), 1+If_LibBoxNum(pBoxLib), nRange0+nRange1+nRange2, nRange, 0, 0, 0)
			If_LibBoxAdd(pBoxLib, pBox)
			for output := int32(0); output < pBox.Outputs(); output++ {
				for input := int32(0); input < pBox.Inputs(); input++ {
					pBox.SetDelay(output, input, 1)
				}
			}

		} else if Wlc_ObjIsCi(pObj) != 0 || pObj.Type == uint32(WLC_OBJ_FF) { // assuming that FFs are ordered immediately after PIs

			if pObj.Type == uint32(WLC_OBJ_FF) {
				Vec_IntPush(vFf2Ci, Gia_ManCiNum(pNew))
			}
			if Wlc_ObjRangeIsReversed(pObj) != 0 {
				for k = 0; k < nRange; k++ {
					Vec_IntPush(vRes, -1)
				}
				for k = 0; k < nRange; k++ {
					Vec_IntWriteEntry(vRes, Vec_IntSize(vRes)-1-k, Gia_ManAppendCi(pNew))
				}

			} else {
				for k = 0; k < nRange; k++ {
					Vec_IntPush(vRes, Gia_ManAppendCi(pNew))
				}
			}
			if pObj.Type == uint32(WLC_OBJ_FO) {
				nFFouts += Vec_IntSize(vRes)
			}
			if pObj.Type == uint32(WLC_OBJ_FF) {
				// complement flop output whose init state is 1
			}

		} else if pObj.Type == uint32(WLC_OBJ_BUF) {
			var nRangeMax int32 = Abc_MaxInt(nRange0, nRange)
			var pArg0 []int32 = Wlc_VecLoadFanins(vTemp0, pFans0, nRange0, nRangeMax, Wlc_ObjIsSignedFanin0(p, pObj))
			for k = 0; k < nRange; k++ {
				Vec_IntPush(vRes, pArg0[k])
			}

		} else if pObj.Type == uint32(WLC_OBJ_CONST) {
			var pTruth *word = (*word)(unsafe.Pointer(&Wlc_ObjFanins(pObj)[0]))
			for k = 0; k < nRange; k++ {
				Vec_IntPush(vRes, Abc_TtGetBit(pTruth, k))
			}

		} else if pObj.Type == uint32(WLC_OBJ_MUX) {
			// It is strange and disturbing that Verilog standard treats these statements differently:
			// Statement 1:
			//     assign o = i ? b : a;
			// Statement 2:
			//     always @( i or a or b )
			//       begin
			//         case ( i )
			//           0 : o = a ;
			//           1 : o = b ;
			//         endcase
			//       end
			// If a is signed and b is unsigned,  Statement 1 does not sign-extend a, while Statement 2 does.
			// The signedness of o does not matter.
			//
			// Below we (somewhat arbitrarily) distinguish these two by assuming that
			// Statement 1 has three fanins, while Statement 2 has more than three fanins.
			//
			var fSigned int32 = 1
			assert(nRange0 >= 1 && Wlc_ObjFaninNum(pObj) >= 3, "nRange0 >= 1 && Wlc_ObjFaninNum(pObj) >= 3")
			assert(1+(1<<nRange0) == Wlc_ObjFaninNum(pObj), "1 + (1 << nRange0) == Wlc_ObjFaninNum(pObj)")
			// inlining of macro Wlc_ObjForEachFanin
			for k = 0; k < Wlc_ObjFaninNum(pObj); k++ {
				iFanin = Wlc_ObjFaninId(pObj, k)
				if k > 0 {
					fSigned &= int32(Wlc_NtkObj(p, iFanin).Signed)
				}
			}
			Vec_IntClear(vTemp1)
			if pPar.fDecMuxes != 0 {
				for k = 0; k < (1 << nRange0); k++ {
					var iLitAnd int32 = 1
					for b = 0; b < nRange0; b++ {
						iLitAnd = Gia_ManHashAnd(pNew, iLitAnd, Abc_LitNotCond(pFans0[b], boolToInt32(((k>>b)&1) == 0)))
					}
					Vec_IntPush(vTemp1, iLitAnd)
				}
			}
			for b = 0; b < nRange; b++ {
				Vec_IntClear(vTemp0)
				// inlining of macro Wlc_ObjForEachFanin
				for k = 0; k < Wlc_ObjFaninNum(pObj); k++ {
					iFanin = Wlc_ObjFaninId(pObj, k)
					if k > 0 {
						nRange1 = Wlc_ObjRange(Wlc_NtkObj(p, iFanin))
						pFans1 = Vec_IntArray(vBits)[Wlc_ObjCopy(p, iFanin):]
						if Wlc_ObjFaninNum(pObj) == 3 { // Statement 1
							value := int32(0)
							if b < nRange1 {
								value = pFans1[b]
							} else if fSigned != 0 {
								value = pFans1[nRange1-1]
							}
							Vec_IntPush(vTemp0, value)
						} else { // Statement 2
							value := int32(0)
							if b < nRange1 {
								value = pFans1[b]
							} else if Wlc_NtkObj(p, iFanin).Signed != 0 {
								value = pFans1[nRange1-1]
							}
							Vec_IntPush(vTemp0, value)
						}
					}
				}
				if pPar.fDecMuxes != 0 {
					Vec_IntPush(vRes, Wlc_NtkMuxTree2(pNew, pFans0, nRange0, vTemp0, vTemp1, vTemp2))
				} else {
					Vec_IntPush(vRes, Wlc_NtkMuxTree_rec(pNew, pFans0, nRange0, vTemp0, 0))
				}
			}

		} else if pObj.Type == uint32(WLC_OBJ_SEL) {
			assert(nRange0 == Wlc_ObjFaninNum(pObj)-1, "nRange0 == Wlc_ObjFaninNum(pObj)-1")
			Vec_IntClear(vTemp1)
			for k = 0; k < nRange0; k++ {
				Vec_IntPush(vTemp1, pFans0[k])
			}
			for b = 0; b < nRange; b++ {
				Vec_IntClear(vTemp0)
				// inlining of macro Wlc_ObjForEachFanin
				for k = 0; k < Wlc_ObjFaninNum(pObj); k++ {
					iFanin = Wlc_ObjFaninId(pObj, k)
					if k > 0 {
						var pFanin *Wlc_Obj_t = Wlc_NtkObj(p, iFanin)
						assert(nRange == Wlc_ObjRange(pFanin), "nRange == Wlc_ObjRange(pFanin)")
						pFans1 = Vec_IntArray(vBits)[Wlc_ObjCopy(p, iFanin):]
						Vec_IntPush(vTemp0, pFans1[b])
					}
				}
				Vec_IntPush(vRes, Wlc_NtkMuxTree2(pNew, nil, 0, vTemp0, vTemp1, vTemp2))
			}

		} else if pObj.Type == uint32(WLC_OBJ_SHIFT_R) || pObj.Type == uint32(WLC_OBJ_SHIFT_RA) ||
			pObj.Type == uint32(WLC_OBJ_SHIFT_L) || pObj.Type == uint32(WLC_OBJ_SHIFT_LA) {
			var nRangeMax int32 = Abc_MaxInt(nRange, nRange0)
			var pArg0 []int32 = Wlc_VecLoadFanins(vTemp0, pFans0, nRange0, nRangeMax, Wlc_ObjIsSignedFanin0(p, pObj))
			if pObj.Type == uint32(WLC_OBJ_SHIFT_R) || pObj.Type == uint32(WLC_OBJ_SHIFT_RA) {
				Wlc_BlastShiftRight(pNew, pArg0, nRangeMax, pFans1, nRange1, boolToInt32(Wlc_ObjIsSignedFanin0(p, pObj) != 0 && pObj.Type == uint32(WLC_OBJ_SHIFT_RA)), vRes)
			} else {
				Wlc_BlastShiftLeft(pNew, pArg0, nRangeMax, pFans1, nRange1, 0, vRes)
			}
			Vec_IntShrink(vRes, nRange)

		} else if pObj.Type == uint32(WLC_OBJ_ROTATE_R) {
			assert(nRange0 == nRange, "nRange0 == nRange")
			Wlc_BlastRotateRight(pNew, pFans0, nRange0, pFans1, nRange1, vRes)

		} else if pObj.Type == uint32(WLC_OBJ_ROTATE_L) {
			assert(nRange0 == nRange, "nRange0 == nRange")
			Wlc_BlastRotateLeft(pNew, pFans0, nRange0, pFans1, nRange1, vRes)

		} else if pObj.Type == uint32(WLC_OBJ_BIT_NOT) {
			var nRangeMax int32 = Abc_MaxInt(nRange, nRange0)
			var pArg0 []int32 = Wlc_VecLoadFanins(vTemp0, pFans0, nRange0, nRangeMax, Wlc_ObjIsSignedFanin0(p, pObj))
			for k = 0; k < nRange; k++ {
				Vec_IntPush(vRes, Abc_LitNot(pArg0[k]))
			}

		} else if pObj.Type == uint32(WLC_OBJ_BIT_AND) || pObj.Type == uint32(WLC_OBJ_BIT_NAND) {
			var nRangeMax int32 = Abc_MaxInt(nRange, Abc_MaxInt(nRange0, nRange1))
			var pArg0 []int32 = Wlc_VecLoadFanins(vTemp0, pFans0, nRange0, nRangeMax, Wlc_ObjIsSignedFanin01(p, pObj))
			var pArg1 []int32 = Wlc_VecLoadFanins(vTemp1, pFans1, nRange1, nRangeMax, Wlc_ObjIsSignedFanin01(p, pObj))
			for k = 0; k < nRange; k++ {
				Vec_IntPush(vRes, Abc_LitNotCond(Gia_ManHashAnd(pNew, pArg0[k], pArg1[k]), boolToInt32(pObj.Type == uint32(WLC_OBJ_BIT_NAND))))
			}

		} else if pObj.Type == uint32(WLC_OBJ_BIT_OR) || pObj.Type == uint32(WLC_OBJ_BIT_NOR) {
			var nRangeMax int32 = Abc_MaxInt(nRange, Abc_MaxInt(nRange0, nRange1))
			var pArg0 []int32 = Wlc_VecLoadFanins(vTemp0, pFans0, nRange0, nRangeMax, Wlc_ObjIsSignedFanin01(p, pObj))
			var pArg1 []int32 = Wlc_VecLoadFanins(vTemp1, pFans1, nRange1, nRangeMax, Wlc_ObjIsSignedFanin01(p, pObj))
			for k = 0; k < nRange; k++ {
				Vec_IntPush(vRes, Abc_LitNotCond(Gia_ManHashOr(pNew, pArg0[k], pArg1[k]), boolToInt32(pObj.Type == uint32(WLC_OBJ_BIT_NOR))))
			}

		} else if pObj.Type == uint32(WLC_OBJ_BIT_XOR) || pObj.Type == uint32(WLC_OBJ_BIT_NXOR) {
			var nRangeMax int32 = Abc_MaxInt(nRange, Abc_MaxInt(nRange0, nRange1))
			var pArg0 []int32 = Wlc_VecLoadFanins(vTemp0, pFans0, nRange0, nRangeMax, Wlc_ObjIsSignedFanin01(p, pObj))
			var pArg1 []int32 = Wlc_VecLoadFanins(vTemp1, pFans1, nRange1, nRangeMax, Wlc_ObjIsSignedFanin01(p, pObj))
			for k = 0; k < nRange; k++ {
				Vec_IntPush(vRes, Abc_LitNotCond(Gia_ManHashXor(pNew, pArg0[k], pArg1[k]), boolToInt32(pObj.Type == uint32(WLC_OBJ_BIT_NXOR))))
			}

		} else if pObj.Type == uint32(WLC_OBJ_BIT_SELECT) {
			var pFanin *Wlc_Obj_t = Wlc_ObjFanin0(p, pObj)
			var End int32 = Wlc_ObjRangeEnd(pObj)
			var Beg int32 = Wlc_ObjRangeBeg(pObj)
			if End >= Beg {
				assert(nRange == End-Beg+1, "nRange == End - Beg + 1")
				assert(pFanin.Beg <= Beg && End <= pFanin.End, "pFanin.Beg <= Beg && End <= pFanin.End")
				for k = Beg; k <= End; k++ {
					Vec_IntPush(vRes, pFans0[k-pFanin.Beg])
				}

			} else {
				assert(nRange == Beg-End+1, "nRange == Beg - End + 1")
				assert(pFanin.End <= End && Beg <= pFanin.Beg, "pFanin.End <= End && Beg <= pFanin.Beg")
				for k = End; k <= Beg; k++ {
					Vec_IntPush(vRes, pFans0[k-pFanin.End])
				}
			}

		} else if pObj.Type == uint32(WLC_OBJ_BIT_CONCAT) {
			var iFanin, nTotal int32
			// inlining of macro Wlc_ObjForEachFanin
			for k = 0; k < Wlc_ObjFaninNum(pObj); k++ {
				iFanin = Wlc_ObjFaninId(pObj, k)
				nTotal += Wlc_ObjRange(Wlc_NtkObj(p, iFanin))
			}
			assert(nRange == nTotal, "nRange == nTotal")
			// inlining of macro Wlc_ObjForEachFaninReverse
			for k = Wlc_ObjFaninNum(pObj) - 1; k >= 0; k-- {
				iFanin = Wlc_ObjFaninId(pObj, k)
				nRange0 = Wlc_ObjRange(Wlc_NtkObj(p, iFanin))
				pFans0 = Vec_IntArray(vBits)[Wlc_ObjCopy(p, iFanin):]
				for b = 0; b < nRange0; b++ {
					Vec_IntPush(vRes, pFans0[b])
				}
			}

		} else if pObj.Type == uint32(WLC_OBJ_BIT_ZEROPAD) || pObj.Type == uint32(WLC_OBJ_BIT_SIGNEXT) {
			var Pad int32
			if pObj.Type == uint32(WLC_OBJ_BIT_ZEROPAD) {
				Pad = 0
			} else {
				Pad = pFans0[nRange0-1]
			}
			assert(nRange0 <= nRange, "nRange0 <= nRange")
			for k = 0; k < nRange0; k++ {
				Vec_IntPush(vRes, pFans0[k])
			}
			for ; k < nRange; k++ {
				Vec_IntPush(vRes, Pad)
			}

		} else if pObj.Type == uint32(WLC_OBJ_LOGIC_NOT) {
			iLit = Wlc_BlastReduction(pNew, pFans0, nRange0, int32(WLC_OBJ_REDUCT_OR))
			Vec_IntFill(vRes, 1, Abc_LitNot(iLit))
			for k = 1; k < nRange; k++ {
				Vec_IntPush(vRes, 0)
			}

		} else if pObj.Type == uint32(WLC_OBJ_LOGIC_IMPL) {
			var iLit0 int32 = Wlc_BlastReduction(pNew, pFans0, nRange0, int32(WLC_OBJ_REDUCT_OR))
			var iLit1 int32 = Wlc_BlastReduction(pNew, pFans1, nRange1, int32(WLC_OBJ_REDUCT_OR))
			Vec_IntFill(vRes, 1, Gia_ManHashOr(pNew, Abc_LitNot(iLit0), iLit1))
			for k = 1; k < nRange; k++ {
				Vec_IntPush(vRes, 0)
			}

		} else if pObj.Type == uint32(WLC_OBJ_LOGIC_AND) {
			var iLit0 int32 = Wlc_BlastReduction(pNew, pFans0, nRange0, int32(WLC_OBJ_REDUCT_OR))
			var iLit1 int32 = Wlc_BlastReduction(pNew, pFans1, nRange1, int32(WLC_OBJ_REDUCT_OR))
			Vec_IntFill(vRes, 1, Gia_ManHashAnd(pNew, iLit0, iLit1))
			for k = 1; k < nRange; k++ {
				Vec_IntPush(vRes, 0)
			}

		} else if pObj.Type == uint32(WLC_OBJ_LOGIC_OR) {
			var iLit0 int32 = Wlc_BlastReduction(pNew, pFans0, nRange0, int32(WLC_OBJ_REDUCT_OR))
			var iLit1 int32 = Wlc_BlastReduction(pNew, pFans1, nRange1, int32(WLC_OBJ_REDUCT_OR))
			Vec_IntFill(vRes, 1, Gia_ManHashOr(pNew, iLit0, iLit1))
			for k = 1; k < nRange; k++ {
				Vec_IntPush(vRes, 0)
			}

		} else if pObj.Type == uint32(WLC_OBJ_LOGIC_XOR) {
			var iLit0 int32 = Wlc_BlastReduction(pNew, pFans0, nRange0, int32(WLC_OBJ_REDUCT_OR))
			var iLit1 int32 = Wlc_BlastReduction(pNew, pFans1, nRange1, int32(WLC_OBJ_REDUCT_OR))
			Vec_IntFill(vRes, 1, Gia_ManHashXor(pNew, iLit0, iLit1))
			for k = 1; k < nRange; k++ {
				Vec_IntPush(vRes, 0)
			}

		} else if pObj.Type == uint32(WLC_OBJ_COMP_NOTEQU) && Wlc_ObjFaninNum(pObj) > 2 {
			// find the max range
			var a int32
			var b int32
			var iRes int32 = 1
			var nRangeMax int32 = Abc_MaxInt(nRange0, nRange1)
			for k = 2; k < Wlc_ObjFaninNum(pObj); k++ {
				nRangeMax = Abc_MaxInt(nRangeMax, Wlc_ObjRange(Wlc_NtkObj(p, Wlc_ObjFaninId(pObj, k))))
			}
			// create pairwise distinct
			for a = 0; a < Wlc_ObjFaninNum(pObj); a++ {
				for b = a + 1; b < Wlc_ObjFaninNum(pObj); b++ {
					var nRange0 int32 = Wlc_ObjRange(Wlc_NtkObj(p, Wlc_ObjFaninId(pObj, a)))
					var nRange1 int32 = Wlc_ObjRange(Wlc_NtkObj(p, Wlc_ObjFaninId(pObj, b)))
					var pFans0 []int32 = Vec_IntArray(vBits)[Wlc_ObjCopy(p, Wlc_ObjFaninId(pObj, a)):]
					var pFans1 []int32 = Vec_IntArray(vBits)[Wlc_ObjCopy(p, Wlc_ObjFaninId(pObj, b)):]
					var pArg0 []int32 = Wlc_VecLoadFanins(vTemp0, pFans0, nRange0, nRangeMax, 0)
					var pArg1 []int32 = Wlc_VecLoadFanins(vTemp1, pFans1, nRange1, nRangeMax, 0)
					var iLit int32
					for k = 0; k < nRangeMax; k++ {
						iLit = Gia_ManHashOr(pNew, iLit, Gia_ManHashXor(pNew, pArg0[k], pArg1[k]))
					}
					iRes = Gia_ManHashAnd(pNew, iRes, iLit)
				}
			}
			Vec_IntFill(vRes, 1, iRes)
			for k = 1; k < nRange; k++ {
				Vec_IntPush(vRes, 0)
			}

		} else if pObj.Type == uint32(WLC_OBJ_COMP_EQU) || pObj.Type == uint32(WLC_OBJ_COMP_NOTEQU) {
			var iLit int32
			var nRangeMax int32 = Abc_MaxInt(nRange0, nRange1)
			var pArg0 []int32 = Wlc_VecLoadFanins(vTemp0, pFans0, nRange0, nRangeMax, Wlc_ObjIsSignedFanin01(p, pObj))
			var pArg1 []int32 = Wlc_VecLoadFanins(vTemp1, pFans1, nRange1, nRangeMax, Wlc_ObjIsSignedFanin01(p, pObj))
			for k = 0; k < nRangeMax; k++ {
				iLit = Gia_ManHashOr(pNew, iLit, Gia_ManHashXor(pNew, pArg0[k], pArg1[k]))
			}
			Vec_IntFill(vRes, 1, Abc_LitNotCond(iLit, boolToInt32(pObj.Type == uint32(WLC_OBJ_COMP_EQU))))
			for k = 1; k < nRange; k++ {
				Vec_IntPush(vRes, 0)
			}

		} else if pObj.Type == uint32(WLC_OBJ_COMP_LESS) || pObj.Type == uint32(WLC_OBJ_COMP_MOREEQU) ||
			pObj.Type == uint32(WLC_OBJ_COMP_MORE) || pObj.Type == uint32(WLC_OBJ_COMP_LESSEQU) {
			var nRangeMax int32 = Abc_MaxInt(nRange0, nRange1)
			var fSigned int32 = Wlc_ObjIsSignedFanin01(p, pObj)
			var pArg0 []int32 = Wlc_VecLoadFanins(vTemp0, pFans0, nRange0, nRangeMax, fSigned)
			var pArg1 []int32 = Wlc_VecLoadFanins(vTemp1, pFans1, nRange1, nRangeMax, fSigned)
			var fSwap int32 = boolToInt32(pObj.Type == uint32(WLC_OBJ_COMP_MORE) || pObj.Type == uint32(WLC_OBJ_COMP_LESSEQU))
			var fCompl int32 = boolToInt32(pObj.Type == uint32(WLC_OBJ_COMP_MOREEQU) || pObj.Type == uint32(WLC_OBJ_COMP_LESSEQU))
			if fSwap != 0 {
				pArg0, pArg1 = pArg1, pArg0
			}
			if fSigned != 0 {
				iLit = Wlc_BlastLessSigned(pNew, pArg0, pArg1, nRangeMax)
			} else {
				iLit = Wlc_BlastLess(pNew, pArg0, pArg1, nRangeMax)
			}
			iLit = Abc_LitNotCond(iLit, fCompl)
			Vec_IntFill(vRes, 1, iLit)
			for k = 1; k < nRange; k++ {
				Vec_IntPush(vRes, 0)
			}

		} else if pObj.Type == uint32(WLC_OBJ_REDUCT_AND) || pObj.Type == uint32(WLC_OBJ_REDUCT_OR) || pObj.Type == uint32(WLC_OBJ_REDUCT_XOR) ||
			pObj.Type == uint32(WLC_OBJ_REDUCT_NAND) || pObj.Type == uint32(WLC_OBJ_REDUCT_NOR) || pObj.Type == uint32(WLC_OBJ_REDUCT_NXOR) {
			Vec_IntPush(vRes, Wlc_BlastReduction(pNew, pFans0, nRange0, int32(pObj.Type)))
			for k = 1; k < nRange; k++ {
				Vec_IntPush(vRes, 0)
			}

		} else if pObj.Type == uint32(WLC_OBJ_ARI_ADD) || pObj.Type == uint32(WLC_OBJ_ARI_SUB) {
			var nRangeMax int32 = Abc_MaxInt(nRange, Abc_MaxInt(nRange0, nRange1))
			var pArg0 []int32 = Wlc_VecLoadFanins(vRes, pFans0, nRange0, nRangeMax, Wlc_ObjIsSignedFanin01(p, pObj))
			var pArg1 []int32 = Wlc_VecLoadFanins(vTemp1, pFans1, nRange1, nRangeMax, Wlc_ObjIsSignedFanin01(p, pObj))
			var CarryIn int32
			if Wlc_ObjFaninNum(pObj) == 3 {
				CarryIn = pFans2[0]
			} else {
				CarryIn = 0
			}
			if pObj.Type == uint32(WLC_OBJ_ARI_ADD) {
				if pPar.fCla != 0 {
					Wlc_BlastAdderCLA(pNew, pArg0, pArg1, nRangeMax, Wlc_ObjIsSignedFanin01(p, pObj), CarryIn) // result is in pFan0 (vRes)
					//Wlc_BlastAdderFast( pNew, pArg0, pArg1, nRangeMax, Wlc_ObjIsSignedFanin01(p, pObj), CarryIn ); // result is in pFan0 (vRes)
				} else {
					Wlc_BlastAdder(pNew, pArg0, pArg1, nRangeMax, CarryIn)
				} // result is in pFan0 (vRes)

			} else {
				Wlc_BlastSubtract(pNew, pArg0, pArg1, nRange, 1)
			} // result is in pFan0 (vRes)
			Vec_IntShrink(vRes, nRange)

		} else if pObj.Type == uint32(WLC_OBJ_ARI_ADDSUB) {
			var nRangeMax int32 = Abc_MaxInt(nRange, Abc_MaxInt(nRange2, nRange3))
			var pArg0 []int32 = Wlc_VecLoadFanins(vTemp0, pFans2, nRange2, nRangeMax, Wlc_ObjIsSignedFanin01(p, pObj))
			var pArg1 []int32 = Wlc_VecLoadFanins(vTemp1, pFans2, nRange2, nRangeMax, Wlc_ObjIsSignedFanin01(p, pObj))
			var pArg2 []int32 = Wlc_VecLoadFanins(vTemp2, pFans3, nRange3, nRangeMax, Wlc_ObjIsSignedFanin01(p, pObj))
			var ModeIn int32 = pFans0[0]
			var CarryIn int32 = pFans1[0]
			var j int32
			Wlc_BlastAdder(pNew, pArg0, pArg2, nRangeMax, CarryIn)                // result is in pArg0 (vTemp0)
			Wlc_BlastSubtract(pNew, pArg1, pArg2, nRangeMax, Abc_LitNot(CarryIn)) // result is in pArg1 (vTemp1)
			Vec_IntClear(vRes)
			for j = 0; j < nRange; j++ {
				Vec_IntPush(vRes, Gia_ManHashMux(pNew, ModeIn, pArg0[j], pArg1[j]))
			}

		} else if pObj.Type == uint32(WLC_OBJ_ARI_MULTI) {
			if fUseOldMultiplierBlasting != 0 {
				var nRangeMax int32 = Abc_MaxInt(nRange, Abc_MaxInt(nRange0, nRange1))
				var pArg0 []int32 = Wlc_VecLoadFanins(vTemp0, pFans0, nRange0, nRangeMax, Wlc_ObjIsSignedFanin01(p, pObj))
				var pArg1 []int32 = Wlc_VecLoadFanins(vTemp1, pFans1, nRange1, nRangeMax, Wlc_ObjIsSignedFanin01(p, pObj))
				Wlc_BlastMultiplier2(pNew, pArg0, pArg1, nRange, vTemp2, vRes)
				Vec_IntShrink(vRes, nRange)

			} else {
				var fSigned int32 = Wlc_ObjIsSignedFanin01(p, pObj)
				var nRangeMax int32 = Abc_MaxInt(nRange0, nRange1)
				var pArg0 []int32 = Wlc_VecLoadFanins(vTemp0, pFans0, nRange0, nRangeMax, fSigned)
				var pArg1 []int32 = Wlc_VecLoadFanins(vTemp1, pFans1, nRange1, nRangeMax, fSigned)
				if Wlc_NtkCountConstBits(pArg0, nRangeMax) < Wlc_NtkCountConstBits(pArg1, nRangeMax) {
					pArg0, pArg1 = pArg1, pArg0
				}
				if pPar.fBooth != 0 {
					Wlc_BlastBooth(pNew, pArg0, pArg1, nRange0, nRange1, vRes, fSigned, pPar.fCla)
				} else if pPar.fCla != 0 {
					Wlc_BlastMultiplier3(pNew, pArg0, pArg1, nRange0, nRange1, vRes, Wlc_ObjIsSignedFanin01(p, pObj), pPar.fCla)
				} else {
					Wlc_BlastMultiplier(pNew, pArg0, pArg1, nRangeMax, nRangeMax, vTemp2, vRes, fSigned)
				}
				if nRange > Vec_IntSize(vRes) {
					fill := int32(0)
					if fSigned != 0 {
						fill = Vec_IntEntryLast(vRes)
					}
					Vec_IntFillExtra(vRes, nRange, fill)
				} else {
					Vec_IntShrink(vRes, nRange)
				}
				assert(Vec_IntSize(vRes) == nRange, "Vec_IntSize(vRes) == nRange")
			}

		} else if pObj.Type == uint32(WLC_OBJ_ARI_DIVIDE) || pObj.Type == uint32(WLC_OBJ_ARI_REM) || pObj.Type == uint32(WLC_OBJ_ARI_MODULUS) {
			var nRangeMax int32 = Abc_MaxInt(nRange, Abc_MaxInt(nRange0, nRange1))
			var fSigned int32 = Wlc_ObjIsSignedFanin01(p, pObj)
			var pArg0 []int32 = Wlc_VecLoadFanins(vTemp0, pFans0, nRange0, nRangeMax, fSigned)
			var pArg1 []int32 = Wlc_VecLoadFanins(vTemp1, pFans1, nRange1, nRangeMax, fSigned)
			if fSigned != 0 {
				Wlc_BlastDividerSigned(pNew, pArg0, nRangeMax, pArg1, nRangeMax, boolToInt32(pObj.Type == uint32(WLC_OBJ_ARI_DIVIDE)), vRes)
			} else {
				Wlc_BlastDivider(pNew, pArg0, nRangeMax, pArg1, nRangeMax, boolToInt32(pObj.Type == uint32(WLC_OBJ_ARI_DIVIDE)), vRes)
			}
			Vec_IntShrink(vRes, nRange)
			if pPar.fDivBy0 == 0 {
				Wlc_BlastZeroCondition(pNew, pFans1, nRange1, vRes)
			}

		} else if pObj.Type == uint32(WLC_OBJ_ARI_MINUS) {
			var nRangeMax int32 = Abc_MaxInt(nRange0, nRange)
			var pArg0 []int32 = Wlc_VecLoadFanins(vTemp0, pFans0, nRange0, nRangeMax, Wlc_ObjIsSignedFanin0(p, pObj))
			Wlc_BlastMinus(pNew, pArg0, nRangeMax, vRes)
			Vec_IntShrink(vRes, nRange)

		} else if pObj.Type == uint32(WLC_OBJ_ARI_POWER) {
			var nRangeMax int32 = Abc_MaxInt(nRange0, nRange)
			var pArg0 []int32 = Wlc_VecLoadFanins(vTemp0, pFans0, nRange0, nRangeMax, Wlc_ObjIsSignedFanin0(p, pObj))
			var pArg1 []int32 = Wlc_VecLoadFanins(vTemp1, pFans1, nRange1, nRange1, Wlc_ObjIsSignedFanin1(p, pObj))
			Wlc_BlastPower(pNew, pArg0, nRangeMax, pArg1, nRange1, vTemp2, vRes)
			Vec_IntShrink(vRes, nRange)

		} else if pObj.Type == uint32(WLC_OBJ_ARI_SQRT) {
			var pArg0 []int32 = Wlc_VecLoadFanins(vTemp0, pFans0, nRange0, nRange0+(nRange0&1), 0)
			nRange0 += (nRange0 & 1)
			if pPar.fNonRest != 0 {
				Wlc_BlastSqrtNR(pNew, pArg0, nRange0, vTemp2, vRes)
			} else {
				Wlc_BlastSqrt(pNew, pArg0, nRange0, vTemp2, vRes)
			}
			if nRange > Vec_IntSize(vRes) {
				Vec_IntFillExtra(vRes, nRange, 0)
			} else {
				Vec_IntShrink(vRes, nRange)
			}

		} else if pObj.Type == uint32(WLC_OBJ_ARI_SQUARE) {
			var pArg0 []int32 = Wlc_VecLoadFanins(vTemp0, pFans0, nRange0, nRange0, 0)
			Wlc_BlastSquare(pNew, pArg0, nRange0, vTemp2, vRes)
			if nRange > Vec_IntSize(vRes) {
				Vec_IntFillExtra(vRes, nRange, 0)
			} else {
				Vec_IntShrink(vRes, nRange)
			}

		} else if pObj.Type == uint32(WLC_OBJ_DEC) {
			var pArg0 []int32 = Wlc_VecLoadFanins(vTemp0, pFans0, nRange0, nRange0, 0)
			Wlc_BlastDecoder(pNew, pArg0, nRange0, vTemp2, vRes)
			if nRange > Vec_IntSize(vRes) {
				Vec_IntFillExtra(vRes, nRange, 0)
			} else {
				Vec_IntShrink(vRes, nRange)
			}

		} else if pObj.Type == uint32(WLC_OBJ_TABLE) {
			Wlc_BlastTable(pNew, Wlc_ObjTable(p, pObj), pFans0, nRange0, nRange, vRes)
		} else if pObj.Type == uint32(WLC_OBJ_LUT) && p.vLutTruths != nil {
			Wlc_BlastLut(pNew, Vec_WrdEntry(p.vLutTruths, Wlc_ObjId(p, pObj)), pFans0, nRange0, nRange, vRes)
		} else {
			assert(false, "0")
		}
		assert(Vec_IntSize(vBits) == Wlc_ObjCopy(p, i), "Vec_IntSize(vBits) == Wlc_ObjCopy(p, i)")
		Vec_IntAppend(vBits, vRes)
		if vAddOutputs != nil && Wlc_ObjIsCo(pObj) == 0 && ((pObj.Type >= uint32(WLC_OBJ_MUX) && pObj.Type <= uint32(WLC_OBJ_ROTATE_L)) || (pObj.Type >= uint32(WLC_OBJ_COMP_EQU) && pObj.Type <= uint32(WLC_OBJ_COMP_MOREEQU)) || (pObj.Type >= uint32(WLC_OBJ_ARI_ADD) && pObj.Type <= uint32(WLC_OBJ_ARI_SQUARE))) {
			Vec_IntAppend(vAddOutputs, vRes)
			Vec_IntPush(vAddObjs, Wlc_ObjId(p, pObj))
		}
		p.nAnds[pObj.Type] += Gia_ManAndNum(pNew) - nAndPrev
	}
	p.nAnds[0] = Gia_ManAndNum(pNew)
	assert(nBits == Vec_IntSize(vBits), "nBits == Vec_IntSize(vBits)")
	Vec_IntFree(vTemp0)
	Vec_IntFree(vTemp1)
	Vec_IntFree(vTemp2)
	Vec_IntFree(vRes)
	// create flop boxes
	// inlining of macro Wlc_NtkForEachFf2
	for i = 0; i < Vec_IntSize(&p.vFfs2); i++ {
		pObj = Wlc_NtkFf2(p, i)
		var pBox *If_Box_t
		var Buffer string // was [100]byte in C

		var pTable []float32
		var vTemp0 *Vec_Int_t = Vec_IntAlloc(100)
		var vTemp1 *Vec_Int_t = Vec_IntAlloc(100)
		var iLit int32
		var nRange int32 = Wlc_ObjRange(pObj)
		var pFans0, pFans1, pFans2, pFans3 []int32
		var iReset, iSet, iEnable int32
		var nRangeIn int32 = 2*nRange + 3 // D, reset, set, enable, Q
		var iSre int32 = Wlc_ObjFaninId(pObj, 6)

		assert(pObj.Type == uint32(WLC_OBJ_FF), "pObj.Type == uint32(WLC_OBJ_FF)")

		// create new box
		if vTables == nil {
			vTables = Vec_PtrAlloc(100)
			Tim_ManSetDelayTables(pManTime, vTables)
		}
		Tim_ManCreateBox(pManTime, curPo, nRangeIn, curPi, nRange, Vec_PtrSize(vTables), 0)
		curPi += nRange
		curPo += nRangeIn

		// create delay table
		pTable = make([]float32, 3+int(nRange*nRangeIn))
		pTable[0] = float32(Vec_PtrSize(vTables))
		pTable[1] = float32(nRangeIn)
		pTable[2] = float32(nRange)
		for k = 0; k < nRange*nRangeIn; k++ {
			pTable[3+k] = 1.0
		}
		Vec_PtrPush(vTables, pTable)

		// create combinational outputs in the normal manager
		if Wlc_ObjFaninNum(pObj) > 0 {
			pFans0 = Vec_IntArray(vBits)[Wlc_ObjCopy(p, Wlc_ObjFaninId0(pObj)):]
		} else {
			pFans0 = nil
		}
		if Wlc_ObjFaninNum(pObj) > 1 {
			pFans1 = Vec_IntArray(vBits)[Wlc_ObjCopy(p, Wlc_ObjFaninId1(pObj)):]
		} else {
			pFans1 = nil
		}
		if Wlc_ObjFaninNum(pObj) > 2 {
			pFans2 = Vec_IntArray(vBits)[Wlc_ObjCopy(p, Wlc_ObjFaninId2(pObj)):]
		} else {
			pFans2 = nil
		}
		if Wlc_ObjFaninNum(pObj) > 3 {
			pFans3 = Vec_IntArray(vBits)[Wlc_ObjCopy(p, Wlc_ObjFaninId(pObj, 3)):]
		} else {
			pFans3 = nil
		}
		for k = 0; k < nRange; k++ {
			Gia_ManAppendCo(pNew, pFans0[k])
		}
		Gia_ManAppendCo(pNew, pFans1[0])
		Gia_ManAppendCo(pNew, pFans2[0])
		Gia_ManAppendCo(pNew, pFans3[0])
		for k = 0; k < nRange; k++ {
			Gia_ManAppendCo(pNew, Gia_Obj2Lit(pNew, Gia_ManCi(pNew, Vec_IntEntry(vFf2Ci, i)+k)))
		}

		// make sure there is enough primary inputs in the manager
		for k = Gia_ManPiNum(pExtra); k < nRangeIn; k++ {
			Gia_ManAppendCi(pExtra)
		}
		// create combinational inputs
		for k = 0; k < nRange; k++ {
			Vec_IntPush(vTemp0, Gia_Obj2Lit(pExtra, Gia_ManPi(pExtra, k)))
		}
		iReset = Gia_Obj2Lit(pExtra, Gia_ManPi(pExtra, nRange+0))
		iSet = Gia_Obj2Lit(pExtra, Gia_ManPi(pExtra, nRange+1))
		iEnable = Gia_Obj2Lit(pExtra, Gia_ManPi(pExtra, nRange+2))
		for k = 0; k < nRange; k++ {
			Vec_IntPush(vTemp1, Gia_Obj2Lit(pExtra, Gia_ManPi(pExtra, nRangeIn-nRange+k)))
		}

		// bit-blast in the external manager
		for k = 0; k < nRange; k++ {
			// enable
			//iLitFlop = Gia_LitNotCond( Gia_ObjCopy(pObj), Gia_FlopIsOne(pObj) );
			//iLit = Gia_ManHashMux( Gia_ObjGia(pObj), Gia_FlopEnableCopy(pObj), iLit, iLitFlop );
			iLit = Gia_ManHashMux(pExtra, iEnable, Vec_IntEntry(vTemp0, k), Vec_IntEntry(vTemp1, k))
			if iSre != 0 {
				// reset
				//iLit = Gia_ManHashAnd( Gia_ObjGia(pObj), iLit, Gia_LitNot(Gia_FlopResetCopy(pObj)) );
				iLit = Gia_ManHashAnd(pExtra, iLit, Abc_LitNot(iReset))
				// set
				//iLit = Gia_ManHashOr( Gia_ObjGia(pObj), iLit, Gia_FlopSetCopy(pObj) );
				iLit = Gia_ManHashOr(pExtra, iLit, iSet)

			} else {
				// set
				//iLit = Gia_ManHashOr( Gia_ObjGia(pObj), iLit, Gia_FlopSetCopy(pObj) );
				iLit = Gia_ManHashOr(pExtra, iLit, iSet)
				// reset
				//iLit = Gia_ManHashAnd( Gia_ObjGia(pObj), iLit, Gia_LitNot(Gia_FlopResetCopy(pObj)) );
				iLit = Gia_ManHashAnd(pExtra, iLit, Abc_LitNot(iReset))
			}
			// create outputs in the external manager
			Gia_ManAppendCo(pExtra, iLit)
		}

		// add box to the library
		Buffer = fmt.Sprintf("%s%03d", "ff_comb", 1+If_LibBoxNum(pBoxLib))
		pBox = If_BoxStart(Abc_UtilStrsav(Buffer), 1+If_LibBoxNum(pBoxLib), nRangeIn, nRange, 0, 0, 0)
		If_LibBoxAdd(pBoxLib, pBox)
		for output := int32(0); output < pBox.Outputs(); output++ {
			for input := int32(0); input < pBox.Inputs(); input++ {
				pBox.SetDelay(output, input, 1)
			}
		}
		Vec_IntFree(vTemp0)
		Vec_IntFree(vTemp1)
	}
	Vec_IntFree(vFf2Ci)
	// create COs
	if pPar.fCreateMiter != 0 || pPar.fCreateWordMiter != 0 {
		var nPairs, nBits int32
		var vOuts *Vec_Int_t = Vec_IntAlloc(100)
		assert(Wlc_NtkPoNum(p)%2 == 0, "Wlc_NtkPoNum(p) % 2 == 0")
		// inlining of macro Wlc_NtkForEachCo
		for i = 0; i < Wlc_NtkCoNum(p); i++ {
			pObj = Wlc_NtkCo(p, i)
			if pObj.fIsFi != 0 {
				nRange = Wlc_ObjRange(pObj)
				pFans0 = Vec_IntArray(vBits)[Wlc_ObjCopy(p, Wlc_ObjId(p, pObj)):]
				if Wlc_ObjRangeIsReversed(pObj) != 0 {
					for k = 0; k < nRange; k++ {
						Gia_ManAppendCo(pNew, pFans0[nRange-1-k])
					}

				} else {
					for k = 0; k < nRange; k++ {
						Gia_ManAppendCo(pNew, pFans0[k])
					}
				}
				nFFins += nRange
				continue
			}
			i++
			pObj2 = Wlc_NtkCo(p, i)
			nRange1 = Wlc_ObjRange(pObj)
			nRange2 = Wlc_ObjRange(pObj2)
			assert(nRange1 == nRange2, "nRange1 == nRange2")
			pFans1 = Vec_IntArray(vBits)[Wlc_ObjCopy(p, Wlc_ObjId(p, pObj)):]
			pFans2 = Vec_IntArray(vBits)[Wlc_ObjCopy(p, Wlc_ObjId(p, pObj2)):]
			if pPar.fCreateWordMiter != 0 {
				Vec_IntClear(vOuts)
				if Wlc_ObjRangeIsReversed(pObj) != 0 {
					for k = 0; k < nRange1; k++ {
						Vec_IntPushTwo(vOuts, pFans1[nRange1-1-k], pFans2[nRange2-1-k])
					}

				} else {
					for k = 0; k < nRange1; k++ {
						Vec_IntPushTwo(vOuts, pFans1[k], pFans2[k])
					}
				}
				Gia_ManAppendCo(pNew, Gia_ManHashDualMiter(pNew, vOuts))

			} else {
				if Wlc_ObjRangeIsReversed(pObj) != 0 {
					for k = 0; k < nRange1; k++ {
						Gia_ManAppendCo(pNew, pFans1[nRange1-1-k])
						Gia_ManAppendCo(pNew, pFans2[nRange2-1-k])
					}

				} else {
					for k = 0; k < nRange1; k++ {
						Gia_ManAppendCo(pNew, pFans1[k])
						Gia_ManAppendCo(pNew, pFans2[k])
					}
				}
			}
			nPairs++
			nBits += nRange1
		}
		Vec_IntFree(vOuts)
		if pPar.fCreateWordMiter != 0 {
			fmt.Printf("Derived an ordinary miter with %d bit-level outputs, one for each pair of word-level outputs.\n", nPairs)
		} else {
			fmt.Printf("Derived a dual-output miter with %d pairs of bits belonging to %d pairs of word-level outputs.\n", nBits, nPairs)
		}

	} else {
		// inlining of macro Wlc_NtkForEachCo
		for i = 0; i < Wlc_NtkCoNum(p); i++ {
			pObj = Wlc_NtkCo(p, i)
			// skip all outputs except the given ones
			if pPar.iOutput >= 0 && (i < pPar.iOutput || i >= pPar.iOutput+pPar.nOutputRange) {
				continue
			}
			// create additional PO literals
			if (vAddOutputs != nil) && pObj.fIsFi != 0 {
				for k = 0; k < Vec_IntSize(vAddOutputs); k++ {
					iLit = Vec_IntEntry(vAddOutputs, k)
					Gia_ManAppendCo(pNew, iLit)
				}
				fmt.Printf("Created %d additional POs for %d interesting internal word-level variables.\n", Vec_IntSize(vAddOutputs), Vec_IntSize(vAddObjs))
				Vec_IntFreeP(&vAddOutputs)
			}
			nRange = Wlc_ObjRange(pObj)
			pFans0 = Vec_IntArray(vBits)[Wlc_ObjCopy(p, Wlc_ObjId(p, pObj)):]
			if fVerbose != 0 {
				fmt.Printf("%s(%d) ", Wlc_ObjName(p, Wlc_ObjId(p, pObj)), Gia_ManCoNum(pNew))
			}
			if Wlc_ObjRangeIsReversed(pObj) != 0 {
				for k = 0; k < nRange; k++ {
					Gia_ManAppendCo(pNew, pFans0[nRange-1-k])
				}

			} else {
				for k = 0; k < nRange; k++ {
					Gia_ManAppendCo(pNew, pFans0[k])
				}
			}
			if pObj.fIsFi != 0 {
				nFFins += nRange
			}
		}
		if fVerbose != 0 {
			fmt.Printf("\n")
		}
	}
	//Vec_IntErase( vBits );
	//Vec_IntErase( &p->vCopies );
	// set the number of registers
	if Vec_IntSize(&p.vFfs2) > 0 {
		assert(nFFins == 0 && nFFouts == 0, "nFFins == 0 && nFFouts == 0")
		// complement flop inputs whose init state is 1
		for i = 0; i < nFf2Regs; i++ {
			Gia_ManAppendCo(pNew, Gia_ManAppendCi(pNew))
		}
		//Gia_ManSetRegNum( pNew, nFf2Regs );

	} else {
		assert(nFFins == nFFouts, "nFFins == nFFouts")
		Gia_ManSetRegNum(pNew, nFFins)
	}
	// finalize AIG
	if pPar.fGiaSimple == 0 && pPar.fNoCleanup == 0 {
		pTemp = pNew
		pNew = Gia_ManCleanup(pTemp)
		Gia_ManDupRemapLiterals(vBits, pTemp)
		//printf( "Cutoff ID %d became %d.\n", 75, Abc_Lit2Var(Gia_ManObj(pTemp, 73)->Value) );
		Gia_ManStop(pTemp)
	}
	// transform AIG with init state
	if p.pInits != "" {
		if int32(len(p.pInits)) != Gia_ManRegNum(pNew) {
			fmt.Printf("The number of init values (%d) does not match the number of flops (%d).\n", int32(len(p.pInits)), Gia_ManRegNum(pNew))
			fmt.Printf("It is assumed that the AIG has constant 0 initial state.\n")

		} else {
			saveNames := int32(0)
			if pPar.fSaveFfNames != 0 {
				saveNames = 1 + Gia_ManRegNum(pNew)
			}
			pTemp = pNew
			pNew = Gia_ManDupZeroUndc(pTemp, p.pInits, saveNames, pPar.fGiaSimple, pPar.fVerbose)
			Gia_ManDupRemapLiterals(vBits, pTemp)
			Gia_ManStop(pTemp)
		}
	}
	// finalize AIG with boxes
	if Vec_IntSize(&p.vFfs2) > 0 {
		curPo += nBitCos + nFf2Regs
		assert(curPi == Tim_ManCiNum(pManTime), "curPi == Tim_ManCiNum(pManTime)")
		assert(curPo == Tim_ManCoNum(pManTime), "curPo == Tim_ManCoNum(pManTime)")
		// finalize the extra AIG
		pTemp = pExtra
		pExtra = Gia_ManCleanup(pTemp)
		Gia_ManStop(pTemp)
		assert(Gia_ManPoNum(pExtra) == Gia_ManCiNum(pNew)-nBitCis-nFf2Regs, "Gia_ManPoNum(pExtra) == Gia_ManCiNum(pNew) - nBitCis - nFf2Regs")
		// attach
		Gia_ManSetBoxData(pNew, pExtra, pManTime)
		// normalize AIG
		pTemp = pNew
		pNew = Gia_ManDupNormalize(pTemp, 0)
		Gia_ManTransferTiming(pNew, pTemp)
		Gia_ManStop(pTemp)
		//Tim_ManPrint( pManTime );
	}
	if pPar.vBoxIds != nil {
		curPo += nBitCos
		assert(curPi == Tim_ManCiNum(pManTime), "curPi == Tim_ManCiNum(pManTime)")
		assert(curPo == Tim_ManCoNum(pManTime), "curPo == Tim_ManCoNum(pManTime)")
		// finalize the extra AIG
		pTemp = pExtra
		pExtra = Gia_ManCleanup(pTemp)
		Gia_ManStop(pTemp)
		assert(Gia_ManPoNum(pExtra) == Gia_ManCiNum(pNew)-nBitCis-nFf2Regs, "Gia_ManPoNum(pExtra) == Gia_ManCiNum(pNew) - nBitCis - nFf2Regs")
		// attach
		Gia_ManSetBoxData(pNew, pExtra, pManTime)
		// normalize AIG
		pTemp = pNew
		pNew = Gia_ManDupNormalize(pTemp, 0)
		Gia_ManTransferTiming(pNew, pTemp)
		Gia_ManStop(pTemp)
		//Tim_ManPrint( pManTime );
	}
	// create input names
	vNamesIn := Vec_PtrAlloc(Gia_ManCiNum(pNew))
	Gia_ManSetInputNames(pNew, vNamesIn)
	// inlining of macro Wlc_NtkForEachCi
	for i = 0; i < Wlc_NtkCiNum(p); i++ {
		pObj = Wlc_NtkCi(p, i)
		if Wlc_ObjIsPi(pObj) != 0 {
			var pName string = Wlc_ObjName(p, Wlc_ObjId(p, pObj))
			nRange = Wlc_ObjRange(pObj)
			if (fSkipBitRange != 0) && nRange == 1 {
				Vec_PtrPush(vNamesIn, Abc_UtilStrsav(pName))
			} else {
				for k = 0; k < nRange; k++ {
					var Buffer string // was [1000]byte in C

					bitIndex := pObj.Beg + k
					if pObj.Beg > pObj.End {
						bitIndex = pObj.Beg - k
					}
					Buffer = fmt.Sprintf("%s[%d]", pName, bitIndex)
					Vec_PtrPush(vNamesIn, Abc_UtilStrsav(Buffer))
					//printf( "Writing %s\n", Buffer );
				}
			}
		}
	}
	if p.pInits != "" {
		var Length int32 = int32(len(p.pInits))
		for i = 0; i < Length; i++ {
			if p.pInits[i] == 'x' || p.pInits[i] == 'X' {
				var Buffer string // was [100]byte in C

				Buffer = fmt.Sprintf("_%s_abc_%d_", "init", i)
				Vec_PtrPush(vNamesIn, Abc_UtilStrsav(Buffer))
				fAdded = 1
			}
		}
		if pPar.fSaveFfNames != 0 {
			for i = 0; i < 1+Length; i++ {
				Vec_PtrPush(vNamesIn, nil)
			}
		}
	}
	// inlining of macro Wlc_NtkForEachCi
	for i = 0; i < Wlc_NtkCiNum(p); i++ {
		pObj = Wlc_NtkCi(p, i)
		if Wlc_ObjIsPi(pObj) == 0 {
			var pName string = Wlc_ObjName(p, Wlc_ObjId(p, pObj))
			nRange = Wlc_ObjRange(pObj)
			if (fSkipBitRange != 0) && nRange == 1 {
				Vec_PtrPush(vNamesIn, Abc_UtilStrsav(pName))
			} else {
				for k = 0; k < nRange; k++ {
					var Buffer string // was [1000]byte in C

					bitIndex := pObj.Beg + k
					if pObj.Beg > pObj.End {
						bitIndex = pObj.Beg - k
					}
					Buffer = fmt.Sprintf("%s[%d]", pName, bitIndex)
					Vec_PtrPush(vNamesIn, Abc_UtilStrsav(Buffer))
				}
			}
		}
	}
	// inlining of macro Wlc_NtkForEachFf2
	for i = 0; i < Vec_IntSize(&p.vFfs2); i++ {
		pObj = Wlc_NtkFf2(p, i)
		var pName string = Wlc_ObjName(p, Wlc_ObjId(p, pObj))
		nRange = Wlc_ObjRange(pObj)
		if (fSkipBitRange != 0) && nRange == 1 {
			var Buffer string // was [1000]byte in C

			Buffer = fmt.Sprintf("%s_fo", pName)
			Vec_PtrPush(vNamesIn, Abc_UtilStrsav(Buffer))

		} else {
			for k = 0; k < nRange; k++ {
				var Buffer string // was [1000]byte in C

				bitIndex := pObj.Beg + k
				if pObj.Beg > pObj.End {
					bitIndex = pObj.Beg - k
				}
				Buffer = fmt.Sprintf("%s_fo[%d]", pName, bitIndex)
				Vec_PtrPush(vNamesIn, Abc_UtilStrsav(Buffer))
			}
		}
	}
	// inlining of macro Wlc_NtkForEachFf2
	for i = 0; i < Vec_IntSize(&p.vFfs2); i++ {
		pObj = Wlc_NtkFf2(p, i)
		var pName string = Wlc_ObjName(p, Wlc_ObjId(p, pObj))
		nRange = Wlc_ObjRange(pObj)
		if (fSkipBitRange != 0) && nRange == 1 {
			Vec_PtrPush(vNamesIn, Abc_UtilStrsav(pName))
		} else {
			for k = 0; k < nRange; k++ {
				var Buffer string // was [1000]byte in C

				bitIndex := pObj.Beg + k
				if pObj.Beg > pObj.End {
					bitIndex = pObj.Beg - k
				}
				Buffer = fmt.Sprintf("%s[%d]", pName, bitIndex)
				Vec_PtrPush(vNamesIn, Abc_UtilStrsav(Buffer))
			}
		}
	}
	if p.pInits != "" && pPar.fSaveFfNames != 0 {
		var pName string
		var Length int32 = int32(len(p.pInits))
		var NameStart int32 = Vec_PtrSize(vNamesIn) - Length
		var NullStart int32 = Vec_PtrSize(vNamesIn) - 2*Length
		var SepStart int32 = Vec_PtrSize(vNamesIn) - 2*Length - 1
		assert(Vec_PtrEntry(vNamesIn, SepStart) == nil, "Vec_PtrEntry(vNamesIn, SepStart) == nil")
		Vec_PtrWriteEntry(vNamesIn, SepStart, Abc_UtilStrsav("_abc_190121_abc_"))
		for i = 0; i < Length; i++ {
			var Buffer string // was [100]byte in C

			Buffer = fmt.Sprintf("%c%s", p.pInits[i], Vec_PtrEntry(vNamesIn, NameStart+i).(string))
			assert(Vec_PtrEntry(vNamesIn, NullStart+i) == nil, "Vec_PtrEntry(vNamesIn, NullStart+i) == nil")
			Vec_PtrWriteEntry(vNamesIn, NullStart+i, Abc_UtilStrsav(Buffer))
		}
		for i = 0; i < Vec_PtrSize(vNamesIn); i++ {
			pName = Vec_PtrEntry(vNamesIn, i).(string)
			assert(pName != "", "pName != nil")
		}
	}
	if p.pInits != "" && (fAdded != 0) {
		Vec_PtrPush(vNamesIn, Abc_UtilStrsav("abc_reset_flop"))
	}
	if pPar.vBoxIds != nil {
		// inlining of macro Wlc_NtkForEachObjVec
		for i = 0; i < Vec_IntSize(pPar.vBoxIds); i++ {
			pObj = Wlc_NtkObj(p, Vec_IntEntry(pPar.vBoxIds, i))
			var pName string = Wlc_ObjName(p, Wlc_ObjId(p, pObj))
			nRange = Wlc_ObjRange(pObj)
			assert(nRange > 1, "nRange > 1")
			for k = 0; k < nRange; k++ {
				var Buffer string // was [1000]byte in C

				bitIndex := pObj.Beg + k
				if pObj.Beg > pObj.End {
					bitIndex = pObj.Beg - k
				}
				Buffer = fmt.Sprintf("%s[%d]", pName, bitIndex)
				Vec_PtrPush(vNamesIn, Abc_UtilStrsav(Buffer))
			}
		}
	}
	assert(Vec_PtrSize(vNamesIn) == Gia_ManCiNum(pNew), "Vec_PtrSize(vNamesIn) == Gia_ManCiNum(pNew)")
	// create output names
	vNamesOut := Vec_PtrAlloc(Gia_ManCoNum(pNew))
	Gia_ManSetOutputNames(pNew, vNamesOut)
	// inlining of macro Wlc_NtkForEachFf2
	for i = 0; i < Vec_IntSize(&p.vFfs2); i++ {
		pObj = Wlc_NtkFf2(p, i)
		var iFanin int32
		// inlining of macro Wlc_ObjForEachFanin
		for b = 0; b < Wlc_ObjFaninNum(pObj); b++ {
			iFanin = Wlc_ObjFaninId(pObj, b)
			var pName string = Wlc_ObjName(p, iFanin)
			nRange = Wlc_ObjRange(Wlc_NtkObj(p, iFanin))
			if (fSkipBitRange != 0) && nRange == 1 {
				Vec_PtrPush(vNamesOut, Abc_UtilStrsav(pName))
			} else {
				for k = 0; k < nRange; k++ {
					var Buffer string // was [1000]byte in C

					var pFanin *Wlc_Obj_t = Wlc_NtkObj(p, iFanin)
					bitIndex := pFanin.Beg + k
					if pFanin.Beg > pFanin.End {
						bitIndex = pFanin.Beg - k
					}
					Buffer = fmt.Sprintf("%s[%d]", pName, bitIndex)
					Vec_PtrPush(vNamesOut, Abc_UtilStrsav(Buffer))
				}
			}
			if b == 3 {
				break
			}
		}
		{
			var pName string = Wlc_ObjName(p, Wlc_ObjId(p, pObj))
			nRange = Wlc_ObjRange(pObj)
			if (fSkipBitRange != 0) && nRange == 1 {
				var Buffer string // was [1000]byte in C

				Buffer = fmt.Sprintf("%s_in", pName)
				Vec_PtrPush(vNamesOut, Abc_UtilStrsav(Buffer))

			} else {
				for k = 0; k < nRange; k++ {
					var Buffer string // was [1000]byte in C

					bitIndex := pObj.Beg + k
					if pObj.Beg > pObj.End {
						bitIndex = pObj.Beg - k
					}
					Buffer = fmt.Sprintf("%s_in[%d]", pName, bitIndex)
					Vec_PtrPush(vNamesOut, Abc_UtilStrsav(Buffer))
				}
			}
		}
	}
	if pPar.vBoxIds != nil {
		// inlining of macro Wlc_NtkForEachObjVec
		for i = 0; i < Vec_IntSize(pPar.vBoxIds); i++ {
			pObj = Wlc_NtkObj(p, Vec_IntEntry(pPar.vBoxIds, i))
			var iFanin, f int32
			// inlining of macro Wlc_ObjForEachFanin
			for f = 0; f < Wlc_ObjFaninNum(pObj); f++ {
				iFanin = Wlc_ObjFaninId(pObj, f)
				var pName string = Wlc_ObjName(p, iFanin)
				nRange = Wlc_ObjRange(Wlc_NtkObj(p, iFanin))
				assert(nRange >= 1, "nRange >= 1")
				for k = 0; k < nRange; k++ {
					var Buffer string // was [1000]byte in C

					var pFanin *Wlc_Obj_t = Wlc_NtkObj(p, iFanin)
					bitIndex := pFanin.Beg + k
					if pFanin.Beg > pFanin.End {
						bitIndex = pFanin.Beg - k
					}
					Buffer = fmt.Sprintf("%s[%d]", pName, bitIndex)
					Vec_PtrPush(vNamesOut, Abc_UtilStrsav(Buffer))
				}
			}
		}
	}
	// add real primary outputs
	// inlining of macro Wlc_NtkForEachCo
	for i = 0; i < Wlc_NtkCoNum(p); i++ {
		pObj = Wlc_NtkCo(p, i)
		if Wlc_ObjIsPo(pObj) != 0 {
			var pName string = Wlc_ObjName(p, Wlc_ObjId(p, pObj))
			nRange = Wlc_ObjRange(pObj)
			if pPar.fCreateWordMiter != 0 {
				i++
				var pObj2 *Wlc_Obj_t = Wlc_NtkCo(p, i)
				var pName2 string = Wlc_ObjName(p, Wlc_ObjId(p, pObj2))
				var Buffer string // was [1000]byte in C

				Buffer = fmt.Sprintf("%s_xor_%s", pName, pName2)
				Vec_PtrPush(vNamesOut, Abc_UtilStrsav(Buffer))
				//printf( "Adding output %s\n", Buffer );

			} else if pPar.fCreateMiter != 0 && nRange > 1 {
				i++
				var pObj2 *Wlc_Obj_t = Wlc_NtkCo(p, i)
				var pName2 string = Wlc_ObjName(p, Wlc_ObjId(p, pObj2))
				var nRange1 int32 = Wlc_ObjRange(pObj)
				assert(nRange == nRange1, "nRange == nRange1")
				for k = 0; k < nRange; k++ {
					var Buffer string // was [1000]byte in C

					index := pObj.Beg + k
					if pObj.Beg >= pObj.End {
						index = pObj.Beg - k
					}
					Buffer = fmt.Sprintf("%s[%d]", pName, index)
					Vec_PtrPush(vNamesOut, Abc_UtilStrsav(Buffer))
					Buffer = fmt.Sprintf("%s[%d]", pName2, index)
					Vec_PtrPush(vNamesOut, Abc_UtilStrsav(Buffer))
				}

			} else {
				if (fSkipBitRange != 0) && nRange == 1 {
					Vec_PtrPush(vNamesOut, Abc_UtilStrsav(pName))
				} else {
					for k = 0; k < nRange; k++ {
						var Buffer string // was [1000]byte in C

						index := pObj.Beg + k
						if pObj.Beg >= pObj.End {
							index = pObj.Beg - k
						}
						Buffer = fmt.Sprintf("%s[%d]", pName, index)
						Vec_PtrPush(vNamesOut, Abc_UtilStrsav(Buffer))
					}
				}
			}
		}
	}
	if vAddObjs != nil {
		// add internal primary outputs
		// inlining of macro Wlc_NtkForEachObjVec
		for i = 0; i < Vec_IntSize(vAddObjs); i++ {
			pObj = Wlc_NtkObj(p, Vec_IntEntry(vAddObjs, i))
			var pName string = Wlc_ObjName(p, Wlc_ObjId(p, pObj))
			nRange = Wlc_ObjRange(pObj)
			if (fSkipBitRange != 0) && nRange == 1 {
				Vec_PtrPush(vNamesOut, Abc_UtilStrsav(pName))
			} else {
				for k = 0; k < nRange; k++ {
					var Buffer string // was [1000]byte in C

					index := pObj.Beg + k
					if pObj.Beg >= pObj.End {
						index = pObj.Beg - k
					}
					Buffer = fmt.Sprintf("%s[%d]", pName, index)
					Vec_PtrPush(vNamesOut, Abc_UtilStrsav(Buffer))
				}
			}
		}
		Vec_IntFreeP(&vAddObjs)
	}
	// add flop outputs
	if fAdded != 0 {
		Vec_PtrPush(vNamesOut, Abc_UtilStrsav("abc_reset_flop_in"))
	}
	// inlining of macro Wlc_NtkForEachCo
	for i = 0; i < Wlc_NtkCoNum(p); i++ {
		pObj = Wlc_NtkCo(p, i)
		if Wlc_ObjIsPo(pObj) == 0 {
			var pName string = Wlc_ObjName(p, Wlc_ObjId(p, pObj))
			nRange = Wlc_ObjRange(pObj)
			if (fSkipBitRange != 0) && nRange == 1 {
				var Buffer string // was [1000]byte in C

				Buffer = fmt.Sprintf("%s_fi", pName)
				Vec_PtrPush(vNamesOut, Abc_UtilStrsav(Buffer))

			} else {
				for k = 0; k < nRange; k++ {
					var Buffer string // was [1000]byte in C

					index := pObj.Beg + k
					if pObj.Beg >= pObj.End {
						index = pObj.Beg - k
					}
					Buffer = fmt.Sprintf("%s_fi[%d]", pName, index)
					Vec_PtrPush(vNamesOut, Abc_UtilStrsav(Buffer))
				}
			}
		}
	}
	// inlining of macro Wlc_NtkForEachFf2
	for i = 0; i < Vec_IntSize(&p.vFfs2); i++ {
		pObj = Wlc_NtkFf2(p, i)
		var pName string = Wlc_ObjName(p, Wlc_ObjId(p, pObj))
		nRange = Wlc_ObjRange(pObj)
		if (fSkipBitRange != 0) && nRange == 1 {
			var Buffer string // was [1000]byte in C

			Buffer = fmt.Sprintf("%s_fi", pName)
			Vec_PtrPush(vNamesOut, Abc_UtilStrsav(Buffer))

		} else {
			for k = 0; k < nRange; k++ {
				var Buffer string // was [1000]byte in C

				index := pObj.Beg + k
				if pObj.Beg >= pObj.End {
					index = pObj.Beg - k
				}
				Buffer = fmt.Sprintf("%s_fi[%d]", pName, index)
				Vec_PtrPush(vNamesOut, Abc_UtilStrsav(Buffer))
			}
		}
	}
	assert(Vec_PtrSize(vNamesOut) == Gia_ManCoNum(pNew), "Vec_PtrSize(vNamesOut) == Gia_ManCoNum(pNew)")

	// replace the current library
	if pBoxLib != nil {
		If_LibBoxFree(Abc_FrameReadLibBox())
		Abc_FrameSetLibBox(pBoxLib)
	}

	//pNew->pSpec = Abc_UtilStrsav( p->pSpec ? p->pSpec : p->pName );
	// dump the miter parts
	// The optional debug miter dump in the C source is disabled.

	//Wlc_NtkPrintNameArray( pNew->vNamesIn );
	//Wlc_NtkPrintNameArray( pNew->vNamesOut );
	if pPar.vBoxIds != nil {
		Vec_PtrFreeFree(vNamesIn)
		vNamesIn = nil
		Vec_PtrFreeFree(vNamesOut)
		vNamesOut = nil
		Gia_ManSetOutputNames(pNew, nil)
	}
	Gia_ManSetRegClasses(pNew, vRegClasses)
	return pNew
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END
