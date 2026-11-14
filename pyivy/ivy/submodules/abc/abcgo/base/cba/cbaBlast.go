package cba

/**CFile****************************************************************

  FileName    [cbaBlast.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Hierarchical word-level netlist.]

  Synopsis    [Collapsing word-level design.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - July 21, 2015.]

  Revision    [$Id: cbaBlast.c,v 1.00 2014/11/29 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "cba.h"
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

	Synopsis    [Helper functions.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Cba_NtkPrepareBits(p *Cba_Ntk_t) int32 {
	var nBits int32
	Cba_NtkCleanFonCopies(p)
	for i := int32(1); i <= Cba_NtkFonNum(p); i++ {
		Cba_FonSetCopy(p, i, nBits)
		nBits += Cba_FonRangeSize(p, i)
	}
	return nBits
}
func Cba_VecCopy(vOut *Vec_Int_t, pArray []int32, nSize int32) []int32 {
	var i int32
	Vec_IntClear(vOut)
	for i = 0; i < nSize; i++ {
		Vec_IntPush(vOut, pArray[i])
	}
	return Vec_IntArray(vOut)
}

func Cba_BlastConst(p *Cba_Ntk_t, vOut *Vec_Int_t, iFon int32, nTotal int32, fSigned int32) {
	bits := cbaConstBits(Cba_NtkConst(p, Cba_FonConst(iFon)), nTotal, fSigned != 0)
	Vec_IntClear(vOut)
	for _, bit := range bits {
		Vec_IntPush(vOut, bit)
	}
}

func Cba_VecLoadFanins(p *Cba_Ntk_t, vOut *Vec_Int_t, iFon int32, pFanins []int32, nFanins int32, nTotal int32, fSigned int32) []int32 {
	assert(nFanins <= nTotal, "nFanins <= nTotal")
	if Cba_FonIsReal(iFon) != 0 {
		var i int32
		var Fill int32
		if fSigned != 0 {
			Fill = pFanins[nFanins-1]
		} else {
			Fill = 0
		}
		Vec_IntClear(vOut)
		for i = 0; i < nTotal; i++ {
			if i < nFanins {
				Vec_IntPush(vOut, pFanins[i])
			} else {
				Vec_IntPush(vOut, Fill)
			}
		}

	} else if Cba_FonIsConst(iFon) != 0 {
		Cba_BlastConst(p, vOut, iFon, nTotal, fSigned)
	} else if iFon == 0 { // undriven input
		Vec_IntFill(vOut, nTotal, 0)
	} else {
		assert(false, "0")
	}
	assert(Vec_IntSize(vOut) == nTotal, "Vec_IntSize(vOut) == nTotal")
	return Vec_IntArray(vOut)
}
func Cba_NtkMuxTree_rec(pNew *Gia_Man_t, pCtrl []int32, nCtrl int32, vData *Vec_Int_t, Shift int32) int32 {
	var iLit0, iLit1 int32
	if nCtrl == 0 {
		return Vec_IntEntry(vData, Shift)
	}
	iLit0 = Cba_NtkMuxTree_rec(pNew, pCtrl, nCtrl-1, vData, Shift)
	iLit1 = Cba_NtkMuxTree_rec(pNew, pCtrl, nCtrl-1, vData, Shift+(1<<(nCtrl-1)))
	return Gia_ManHashMux(pNew, pCtrl[nCtrl-1], iLit1, iLit0)
}

/*
*Function*************************************************************

	Synopsis    [Bit blasting for specific operations.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Cba_BlastShiftRight(pNew *Gia_Man_t, pNum []int32, nNum int32, pShift []int32, nShift int32, fSticky int32, vRes *Vec_Int_t) {
	pRes := Cba_VecCopy(vRes, pNum, nNum)
	var Fill int32
	if fSticky != 0 {
		Fill = pNum[nNum-1]
	} else {
		Fill = 0
	}
	var i, j, fShort int32
	if nShift > 32 {
		nShift = 32
	}
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
func Cba_BlastShiftLeft(pNew *Gia_Man_t, pNum []int32, nNum int32, pShift []int32, nShift int32, fSticky int32, vRes *Vec_Int_t) {
	pRes := Cba_VecCopy(vRes, pNum, nNum)
	var Fill int32
	if fSticky != 0 {
		Fill = pNum[0]
	} else {
		Fill = 0
	}
	var i, j, fShort int32
	if nShift > 32 {
		nShift = 32
	}
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
func Cba_BlastRotateRight(pNew *Gia_Man_t, pNum []int32, nNum int32, pShift []int32, nShift int32, vRes *Vec_Int_t) {
	pRes := Cba_VecCopy(vRes, pNum, nNum)
	var i int32
	var j int32
	pTemp := make([]int32, nNum)
	assert(nShift <= 32, "nShift <= 32")
	for i = 0; i < nShift; i++ {
		pRes = Cba_VecCopy(vRes, pTemp, nNum)
		for j = 0; j < nNum; j++ {
			pTemp[j] = Gia_ManHashMux(pNew, pShift[i], pRes[(j+(1<<i))%nNum], pRes[j])
		}
	}

}
func Cba_BlastRotateLeft(pNew *Gia_Man_t, pNum []int32, nNum int32, pShift []int32, nShift int32, vRes *Vec_Int_t) {
	pRes := Cba_VecCopy(vRes, pNum, nNum)
	var i int32
	var j int32
	pTemp := make([]int32, nNum)
	assert(nShift <= 32, "nShift <= 32")
	for i = 0; i < nShift; i++ {
		pRes = Cba_VecCopy(vRes, pTemp, nNum)
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
	}

}
func Cba_BlastReduction(pNew *Gia_Man_t, pFans []int32, nFans int32, Type Cba_ObjType_t) int32 {
	if Type == CBA_BOX_RAND {
		var k int32
		var iLit int32 = 1
		for k = 0; k < nFans; k++ {
			iLit = Gia_ManHashAnd(pNew, iLit, pFans[k])
		}
		return iLit
	}
	if Type == CBA_BOX_ROR {
		var k, iLit int32
		for k = 0; k < nFans; k++ {
			iLit = Gia_ManHashOr(pNew, iLit, pFans[k])
		}
		return iLit
	}
	if Type == CBA_BOX_RXOR {
		var k, iLit int32
		for k = 0; k < nFans; k++ {
			iLit = Gia_ManHashXor(pNew, iLit, pFans[k])
		}
		return iLit
	}
	assert(false, "0")
	return -1
}
func Cba_BlastLess2(pNew *Gia_Man_t, pArg0 []int32, pArg1 []int32, nBits int32) int32 {
	var k, iKnown, iRes int32
	for k = nBits - 1; k >= 0; k-- {
		iRes = Gia_ManHashMux(pNew, iKnown, iRes, Gia_ManHashAnd(pNew, Abc_LitNot(pArg0[k]), pArg1[k]))
		iKnown = Gia_ManHashOr(pNew, iKnown, Gia_ManHashXor(pNew, pArg0[k], pArg1[k]))
		if iKnown == 1 {
			break
		}
	}
	return iRes
}
func Cba_BlastLess_rec(pNew *Gia_Man_t, pArg0 []int32, pArg1 []int32, nBits int32, pYes *int32, pNo *int32) {
	if nBits > 1 {
		var Yes int32 = Gia_ManHashAnd(pNew, Abc_LitNot(pArg0[nBits-1]), pArg1[nBits-1])
		var YesR int32
		var No int32 = Gia_ManHashAnd(pNew, Abc_LitNot(pArg1[nBits-1]), pArg0[nBits-1])
		var NoR int32
		if Yes == 1 || No == 1 {
			*pYes = Yes
			*pNo = No
			return
		}
		Cba_BlastLess_rec(pNew, pArg0, pArg1, nBits-1, &YesR, &NoR)
		*pYes = Gia_ManHashOr(pNew, Yes, Gia_ManHashAnd(pNew, Abc_LitNot(No), YesR))
		*pNo = Gia_ManHashOr(pNew, No, Gia_ManHashAnd(pNew, Abc_LitNot(Yes), NoR))
		return
	}
	assert(nBits == 1, "nBits == 1")
	*pYes = Gia_ManHashAnd(pNew, Abc_LitNot(pArg0[0]), pArg1[0])
	*pNo = Gia_ManHashAnd(pNew, Abc_LitNot(pArg1[0]), pArg0[0])
}
func Cba_BlastLess(pNew *Gia_Man_t, pArg0 []int32, pArg1 []int32, nBits int32) int32 {
	var Yes, No int32
	if nBits == 0 {
		return 0
	}
	Cba_BlastLess_rec(pNew, pArg0, pArg1, nBits, &Yes, &No)
	return Yes
}
func Cba_BlastLessSigned(pNew *Gia_Man_t, pArg0 []int32, pArg1 []int32, nBits int32) int32 {
	var iDiffSign int32 = Gia_ManHashXor(pNew, pArg0[nBits-1], pArg1[nBits-1])
	return Gia_ManHashMux(pNew, iDiffSign, pArg0[nBits-1], Cba_BlastLess(pNew, pArg0, pArg1, nBits-1))
}
func Cba_BlastFullAdder(pNew *Gia_Man_t, a int32, b int32, c int32, pc *int32, ps *int32) {
	var fUseXor int32
	if fUseXor != 0 {
		var Xor int32 = Gia_ManHashXor(pNew, a, b)
		var And1 int32 = Gia_ManHashAnd(pNew, a, b)
		var And2 int32 = Gia_ManHashAnd(pNew, c, Xor)
		*ps = Gia_ManHashXor(pNew, c, Xor)
		*pc = Gia_ManHashOr(pNew, And1, And2)

	} else {
		var And1 int32 = Gia_ManHashAnd(pNew, a, b)
		var And1_ int32 = Gia_ManHashAnd(pNew, Abc_LitNot(a), Abc_LitNot(b))
		var Xor int32 = Abc_LitNot(Gia_ManHashOr(pNew, And1, And1_))
		var And2 int32 = Gia_ManHashAnd(pNew, c, Xor)
		var And2_ int32 = Gia_ManHashAnd(pNew, Abc_LitNot(c), Abc_LitNot(Xor))
		*ps = Abc_LitNot(Gia_ManHashOr(pNew, And2, And2_))
		*pc = Gia_ManHashOr(pNew, And1, And2)
	}
}
func Cba_BlastAdder(pNew *Gia_Man_t, Carry int32, pAdd0, pAdd1 []int32, nBits int32) int32 {
	var b int32
	for b = 0; b < nBits; b++ {
		Cba_BlastFullAdder(pNew, pAdd0[b], pAdd1[b], Carry, &Carry, &pAdd0[b])
	}
	return Carry
}
func Cba_BlastSubtract(pNew *Gia_Man_t, pAdd0, pAdd1 []int32, nBits int32) {
	var b int32
	var Carry int32 = 1
	for b = 0; b < nBits; b++ {
		Cba_BlastFullAdder(pNew, pAdd0[b], Abc_LitNot(pAdd1[b]), Carry, &Carry, &pAdd0[b])
	}
}
func Cba_BlastMinus(pNew *Gia_Man_t, pNum []int32, nNum int32, vRes *Vec_Int_t) {
	pRes := Cba_VecCopy(vRes, pNum, nNum)
	var i, invert int32
	for i = 0; i < nNum; i++ {
		pRes[i] = Gia_ManHashMux(pNew, invert, Abc_LitNot(pRes[i]), pRes[i])
		invert = Gia_ManHashOr(pNew, invert, pNum[i])
	}
}
func Cba_BlastMultiplier2(pNew *Gia_Man_t, pArg0 []int32, pArg1 []int32, nBits int32, vTemp *Vec_Int_t, vRes *Vec_Int_t) {
	var i, j int32
	Vec_IntFill(vRes, nBits, 0)
	for i = 0; i < nBits; i++ {
		Vec_IntFill(vTemp, i, 0)
		for j = 0; Vec_IntSize(vTemp) < nBits; j++ {
			Vec_IntPush(vTemp, Gia_ManHashAnd(pNew, pArg0[j], pArg1[i]))
		}
		assert(Vec_IntSize(vTemp) == nBits, "Vec_IntSize(vTemp) == nBits")
		Cba_BlastAdder(pNew, 0, Vec_IntArray(vRes), Vec_IntArray(vTemp), nBits)
	}
}
func Cba_BlastFullAdderCtrl(pNew *Gia_Man_t, a int32, ac int32, b int32, c int32, pc *int32, ps *int32, fNeg int32) {
	var And int32 = Abc_LitNotCond(Gia_ManHashAnd(pNew, a, ac), fNeg)
	Cba_BlastFullAdder(pNew, And, b, c, pc, ps)
}
func Cba_BlastFullAdderSubtr(pNew *Gia_Man_t, a int32, b int32, c int32, pc *int32, ps *int32, fSub int32) {
	Cba_BlastFullAdder(pNew, Gia_ManHashXor(pNew, a, fSub), b, c, pc, ps)
}
func Cba_BlastMultiplier(pNew *Gia_Man_t, pArgA, pArgB []int32, nArgA, nArgB int32, vTemp, vRes *Vec_Int_t, fSigned int32) {
	assert(nArgA > 0 && nArgB > 0, "nArgA > 0 && nArgB > 0")
	assert(fSigned == 0 || fSigned == 1, "fSigned == 0 || fSigned == 1")
	Vec_IntFill(vRes, nArgA+nArgB, 0)
	pRes := Vec_IntArray(vRes)
	Vec_IntFill(vTemp, 2*nArgA, 0)
	pArgC := Vec_IntArray(vTemp)
	pArgS := Vec_IntArray(vTemp)[nArgA:]
	for b := int32(0); b < nArgB; b++ {
		for a := int32(0); a < nArgA; a++ {
			ps := &pRes[b]
			if a > 0 {
				ps = &pArgS[a-1]
			}
			neg := boolToInt(fSigned != 0 && ((a+1 == nArgA) != (b+1 == nArgB)))
			Cba_BlastFullAdderCtrl(pNew, pArgA[a], pArgB[b], pArgS[a], pArgC[a], &pArgC[a], ps, neg)
		}
	}
	pArgS[nArgA-1] = fSigned
	carry := fSigned
	for a := int32(0); a < nArgA; a++ {
		Cba_BlastFullAdderCtrl(pNew, 1, pArgC[a], pArgS[a], carry, &carry, &pRes[nArgB+a], 0)
	}
}

func Cba_BlastDivider(pNew *Gia_Man_t, pNum []int32, nNum int32, pDiv []int32, nDiv int32, fQuo int32, vRes *Vec_Int_t) {
	pRes := Cba_VecCopy(vRes, pNum, nNum)
	pQuo := make([]int32, nNum)
	pTemp := make([]int32, nNum)
	assert(nNum == nDiv, "nNum == nDiv")
	for j := nNum - 1; j >= 0; j-- {
		known := int32(0)
		for i := nNum - 1; i > nNum-1-j; i-- {
			known = Gia_ManHashOr(pNew, known, pDiv[i])
			if known == 1 {
				break
			}
		}
		pQuo[j] = known
		for i := nNum - 1; i >= 0; i-- {
			if known == 1 {
				break
			}
			yBit := int32(0)
			if i >= j {
				yBit = pDiv[i-j]
			}
			pQuo[j] = Gia_ManHashMux(pNew, known, pQuo[j], Gia_ManHashAnd(pNew, yBit, Abc_LitNot(pRes[i])))
			known = Gia_ManHashOr(pNew, known, Gia_ManHashXor(pNew, yBit, pRes[i]))
		}
		pQuo[j] = Abc_LitNot(pQuo[j])
		if pQuo[j] == 0 {
			continue
		}
		borrow := int32(0)
		for i := int32(0); i < nNum; i++ {
			topBit := Gia_ManHashMux(pNew, borrow, Abc_LitNot(pRes[i]), pRes[i])
			yBit := int32(0)
			if i >= j {
				yBit = pDiv[i-j]
			}
			borrow = Gia_ManHashMux(pNew, pRes[i], Gia_ManHashAnd(pNew, borrow, yBit), Gia_ManHashOr(pNew, borrow, yBit))
			pTemp[i] = Gia_ManHashXor(pNew, topBit, yBit)
		}
		if pQuo[j] == 1 {
			Cba_VecCopy(vRes, pTemp, nNum)
		} else {
			for i := int32(0); i < nNum; i++ {
				pRes[i] = Gia_ManHashMux(pNew, pQuo[j], pTemp[i], pRes[i])
			}
		}
	}

	if fQuo != 0 {
		Cba_VecCopy(vRes, pQuo, nNum)
	}
}

func Cba_BlastDivider2(pNew *Gia_Man_t, pNum []int32, nNum int32, pDiv []int32, nDiv int32, fQuo int32, vRes *Vec_Int_t) {
	pQuo := make([]int32, nNum)
	assert(nNum > 0 && nDiv > 0, "nNum > 0 && nDiv > 0")
	assert(Vec_IntSize(vRes) < nNum+nDiv, "Vec_IntSize(vRes) < nNum+nDiv")
	Vec_IntGrowResize(vRes, nNum+nDiv)
	pRes := Vec_IntArray(vRes)
	for i := int32(0); i < nNum+nDiv; i++ {
		if i < nNum {
			pRes[i] = pNum[i]
		} else {
			pRes[i] = 0
		}
	}
	for i := nNum - 1; i >= 0; i-- {
		ctrl := int32(1)
		if i != nNum-1 {
			ctrl = pQuo[i+1]
		}
		carry := ctrl
		for k := int32(0); k <= nDiv; k++ {
			addend := int32(0)
			if k < nDiv {
				addend = pDiv[k]
			}
			Cba_BlastFullAdderSubtr(pNew, addend, pRes[i+k], carry, &carry, &pRes[i+k], ctrl)
		}
		pQuo[i] = Abc_LitNot(pRes[i+nDiv])
	}
	if fQuo != 0 {
		Cba_VecCopy(vRes, pQuo, nNum)
	} else {
		carry, temp := int32(0), int32(0)
		for k := int32(0); k < nDiv; k++ {
			Cba_BlastFullAdder(pNew, pDiv[k], pRes[k], carry, &carry, &temp)
			pRes[k] = Gia_ManHashMux(pNew, pQuo[0], pRes[k], temp)
		}
		Vec_IntShrink(vRes, nDiv)
	}
}

func Cba_BlastDividerSigned(pNew *Gia_Man_t, pNum []int32, nNum int32, pDiv []int32, nDiv int32, fQuo int32, vRes *Vec_Int_t) {
	vNum, vDiv := Vec_IntAlloc(nNum), Vec_IntAlloc(nDiv)
	vRes00, vRes01 := Vec_IntAlloc(nNum+nDiv), Vec_IntAlloc(nNum+nDiv)
	vRes10, vRes11 := Vec_IntAlloc(nNum+nDiv), Vec_IntAlloc(nNum+nDiv)
	vRes2 := Vec_IntAlloc(nNum)
	diffSign := Gia_ManHashXor(pNew, pNum[nNum-1], pDiv[nDiv-1])
	Cba_BlastMinus(pNew, pNum, nNum, vNum)
	Cba_BlastMinus(pNew, pDiv, nDiv, vDiv)
	Cba_BlastDivider(pNew, pNum, nNum, pDiv, nDiv, fQuo, vRes00)
	Cba_BlastDivider(pNew, pNum, nNum, Vec_IntArray(vDiv), nDiv, fQuo, vRes01)
	Cba_BlastDivider(pNew, Vec_IntArray(vNum), nNum, pDiv, nDiv, fQuo, vRes10)
	Cba_BlastDivider(pNew, Vec_IntArray(vNum), nNum, Vec_IntArray(vDiv), nDiv, fQuo, vRes11)
	Vec_IntClear(vRes)
	for k := int32(0); k < nNum; k++ {
		data0 := Gia_ManHashMux(pNew, pDiv[nDiv-1], Vec_IntEntry(vRes01, k), Vec_IntEntry(vRes00, k))
		data1 := Gia_ManHashMux(pNew, pDiv[nDiv-1], Vec_IntEntry(vRes11, k), Vec_IntEntry(vRes10, k))
		Vec_IntPush(vRes, Gia_ManHashMux(pNew, pNum[nNum-1], data1, data0))
	}
	Cba_BlastMinus(pNew, Vec_IntArray(vRes), nNum, vRes2)
	control := pNum[nNum-1]
	if fQuo != 0 {
		control = diffSign
	}
	for k := int32(0); k < nNum; k++ {
		Vec_IntWriteEntry(vRes, k, Gia_ManHashMux(pNew, control, Vec_IntEntry(vRes2, k), Vec_IntEntry(vRes, k)))
	}
	Vec_IntFree(vNum)
	Vec_IntFree(vDiv)
	Vec_IntFree(vRes00)
	Vec_IntFree(vRes01)
	Vec_IntFree(vRes10)
	Vec_IntFree(vRes11)
	Vec_IntFree(vRes2)
	assert(Vec_IntSize(vRes) == nNum, "Vec_IntSize(vRes) == nNum")
}

// cbaBlastTableTruth follows the retained C loop order: each minterm is set
// when any of its nFans table bits is set. The source currently does not use
// its output-loop index when selecting those bits, so every output gets this
// same truth table.
func cbaBlastTableTruth(pTable []uint64, nFans int32) []int32 {
	nMints := int32(1) << nFans
	truth := make([]int32, nMints)
	for m := int32(0); m < nMints; m++ {
		for i := int32(0); i < nFans; i++ {
			bit := m*nFans + i
			if (pTable[bit>>6]>>(bit&63))&1 != 0 {
				truth[m] = 1
			}
		}
	}
	return truth
}

func Cba_BlastTable(pNew *Gia_Man_t, pTable []uint64, pFans []int32, nFans int32, nOuts int32, vRes *Vec_Int_t) {
	truth := cbaBlastTableTruth(pTable, nFans)
	Vec_IntClear(vRes)
	for out := int32(0); out < nOuts; out++ {
		var build func(level, minterm int32) int32
		build = func(level, minterm int32) int32 {
			if level == nFans {
				return truth[minterm]
			}
			low := build(level+1, minterm)
			high := build(level+1, minterm|(1<<level))
			return Gia_ManHashMux(pNew, pFans[level], high, low)
		}
		Vec_IntPush(vRes, build(0, 0))
	}
}

func Cba_BlastZeroCondition(pNew *Gia_Man_t, pDiv []int32, nDiv int32, vRes *Vec_Int_t) {
	var i int32
	var Entry int32
	var iLit int32 = Cba_BlastReduction(pNew, pDiv, nDiv, CBA_BOX_ROR)
	for i = 0; i < Vec_IntSize(vRes); i++ {
		Entry = Vec_IntEntry(vRes, i)
		Vec_IntWriteEntry(vRes, i, Gia_ManHashAnd(pNew, iLit, Entry))
	}
}
func Cba_BlastPower(pNew *Gia_Man_t, pNum []int32, nNum int32, pExp []int32, nExp int32, vTemp *Vec_Int_t, vRes *Vec_Int_t) {
	var vDegrees *Vec_Int_t = Vec_IntAlloc(2 * nNum)
	var vResTemp *Vec_Int_t = Vec_IntAlloc(2 * nNum)
	var i int32
	var pDegrees []int32
	pRes := Vec_IntArray(vRes)
	var k int32
	pResTemp := Vec_IntArray(vResTemp)
	Vec_IntFill(vRes, nNum, 0)
	Vec_IntWriteEntry(vRes, 0, 1)
	for i = 0; i < nExp; i++ {
		if i == 0 {
			pDegrees = Cba_VecCopy(vDegrees, pNum, nNum)
		} else {
			Cba_BlastMultiplier2(pNew, pDegrees, pDegrees, nNum, vTemp, vResTemp)
			pDegrees = Cba_VecCopy(vDegrees, pResTemp, nNum)
		}
		Cba_BlastMultiplier2(pNew, pRes, pDegrees, nNum, vTemp, vResTemp)
		for k = 0; k < nNum; k++ {
			pRes[k] = Gia_ManHashMux(pNew, pExp[i], pResTemp[k], pRes[k])
		}
	}
	Vec_IntFree(vResTemp)
	Vec_IntFree(vDegrees)
}
func Cba_BlastSqrt(pNew *Gia_Man_t, pNum []int32, nNum int32, vTmp *Vec_Int_t, vRes *Vec_Int_t) {
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
	for i = 0; i < nNum/2; i++ {
		pSumP[0] = pNum[nNum-2*i-2]
		pSumP[1] = pNum[nNum-2*i-1]
		for k = 0; k < i+1; k++ {
			pSumP[k+2] = pSum[k]
		}
		for k = 0; k < i+3; k++ {
			if k >= 2 && k < i+2 { // middle ones
				Cba_BlastFullAdder(pNew, pSumP[k], Abc_LitNot(pRes[i-k+1]), Carry, &Carry, &pSum[k])
			} else {
				var subtrahend int32 = 1
				if k != 0 {
					subtrahend = Carry
				}
				Cba_BlastFullAdder(pNew, pSumP[k], Abc_LitNot(subtrahend), 1, &Carry, &pSum[k])
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

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Cba_NtkBlast(p *Cba_Ntk_t, fSeq int32) *Gia_Man_t {
	var fUseOldMultiplierBlasting int32
	var pTemp, pNew *Gia_Man_t
	var vTemp0, vTemp1, vTemp2, vRes *Vec_Int_t
	var vInit *Vec_Str_t
	if fSeq != 0 {
		vInit = Vec_StrAlloc(100)
	} else {
		vInit = nil
	}
	var vBits *Vec_Int_t = &p.vFonBits
	var nBits int32 = Cba_NtkPrepareBits(p)
	var pFans0, pFans1, pFans2 []int32
	var nRange, nRange0, nRange1, nRange2 int32
	var Type Cba_ObjType_t
	var iFon, iFon0, iFon1, iFon2, fSigned01 int32
	var i, k, b, iFin, iObj, iLit, nAndPrev int32
	Vec_IntClear(vBits)
	Vec_IntGrow(vBits, nBits)
	vTemp0 = Vec_IntAlloc(1000)
	vTemp1 = Vec_IntAlloc(1000)
	vTemp2 = Vec_IntAlloc(1000)
	vRes = Vec_IntAlloc(1000)
	// clean AND-gate counters
	for t := range p.pDesign.nAnds {
		p.pDesign.nAnds[t] = 0
	}
	// create AIG manager
	pNew = Gia_ManStart(5*Cba_NtkObjNum(p) + 1000)
	Gia_ManSetNameSpec(pNew, Cba_ManName(p.pDesign), Cba_ManName(p.pDesign))
	Gia_ManHashAlloc(pNew)
	// blast in the topological order
	for i = 1; i < Vec_StrSize(&p.vObjType); i++ {
		Type = Cba_ObjType(p, i)
		if Type == CBA_OBJ_PO {
			continue
		}
		assert(Vec_IntSize(vBits) == Cba_FonCopy(p, Cba_ObjFon0(p, i)), "Vec_IntSize(vBits) == Cba_FonCopy(p, Cba_ObjFon0(p, i))")
		nRange = Cba_ObjRangeSize(p, i)
		assert(nRange > 0, "nRange > 0")
		if Cba_ObjIsPi(p, i) != 0 || Cba_ObjIsSeq(p, i) != 0 {
			for k = 0; k < nRange; k++ {
				Vec_IntPush(vBits, Gia_ManAppendCi(pNew))
			}
			assert(Type == CBA_BOX_DFFCPL || Cba_ObjFonNum(p, i) == 1, "Type == CBA_BOX_DFFCPL || Cba_ObjFonNum(p, i) == 1")
			continue
		}
		assert(Cba_ObjFinNum(p, i) > 0, "Cba_ObjFinNum(p, i) > 0")
		if Cba_ObjFinNum(p, i) > 0 {
			iFon0 = Cba_ObjFinFon(p, i, 0)
		} else {
			iFon0 = -1
		}
		if Cba_ObjFinNum(p, i) > 1 {
			iFon1 = Cba_ObjFinFon(p, i, 1)
		} else {
			iFon1 = -1
		}
		if Cba_ObjFinNum(p, i) > 2 {
			iFon2 = Cba_ObjFinFon(p, i, 2)
		} else {
			iFon2 = -1
		}
		if Cba_ObjFinNum(p, i) > 0 {
			nRange0 = Cba_FonRangeSize(p, iFon0)
		} else {
			nRange0 = -1
		}
		if Cba_ObjFinNum(p, i) > 1 {
			nRange1 = Cba_FonRangeSize(p, iFon1)
		} else {
			nRange1 = -1
		}
		if Cba_ObjFinNum(p, i) > 2 {
			nRange2 = Cba_FonRangeSize(p, iFon2)
		} else {
			nRange2 = -1
		}
		if Cba_ObjFinNum(p, i) > 0 && Cba_FonIsReal(iFon0) != 0 {
			pFans0 = Vec_IntArray(vBits)[Cba_FonCopy(p, iFon0):]
		} else {
			pFans0 = nil
		}
		if Cba_ObjFinNum(p, i) > 1 && Cba_FonIsReal(iFon1) != 0 {
			pFans1 = Vec_IntArray(vBits)[Cba_FonCopy(p, iFon1):]
		} else {
			pFans1 = nil
		}
		if Cba_ObjFinNum(p, i) > 2 && Cba_FonIsReal(iFon2) != 0 {
			pFans2 = Vec_IntArray(vBits)[Cba_FonCopy(p, iFon2):]
		} else {
			pFans2 = nil
		}
		fSigned01 = boolToInt(Cba_ObjFinNum(p, i) > 1 && Cba_FonSigned(p, iFon0) != 0 && Cba_FonSigned(p, iFon1) != 0)
		nAndPrev = Gia_ManAndNum(pNew)
		Vec_IntClear(vRes)
		if Type == CBA_BOX_SLICE {
			var Left int32 = Cba_ObjLeft(p, i)
			var Right int32 = Cba_ObjRight(p, i)
			var Left0 int32 = Cba_FonLeft(p, iFon0)
			var Right0 int32 = Cba_FonRight(p, iFon0)
			if Left > Right {
				assert(Left0 > Right0, "Left0 > Right0")
				for k = Right; k <= Left; k++ {
					Vec_IntPush(vRes, pFans0[k-Right0])
				}

			} else {
				assert(Left < Right && Left0 < Right0, "Left < Right && Left0 < Right0")
				for k = Right; k >= Left; k-- {
					Vec_IntPush(vRes, pFans0[k-Right0])
				}
			}

		} else if Type == CBA_BOX_CONCAT {
			var iFinT, iFonT, nTotal int32
			for k = 0; k < Cba_ObjFinNum(p, i); k++ {
				iFinT = Cba_ObjFin0(p, i) + k
				iFonT = Cba_FinFon(p, iFinT)
				nTotal += Cba_FonRangeSize(p, iFonT)
			}
			assert(nRange == nTotal, "nRange == nTotal")
			for k = 0; k < Cba_ObjFinNum(p, i); k++ {
				iFinT = Cba_ObjFin0(p, i) + k
				iFonT = Cba_FinFon(p, iFinT)
				nRange0 = Cba_FonRangeSize(p, iFonT)
				if Cba_FonIsReal(iFonT) != 0 {
					pFans0 = Vec_IntArray(vBits)[Cba_FonCopy(p, iFonT):]
				} else {
					pFans0 = nil
				}
				pFans0 = Cba_VecLoadFanins(p, vTemp0, iFonT, pFans0, nRange0, nRange0, Cba_FonSigned(p, iFonT))
				for b = 0; b < nRange0; b++ {
					Vec_IntPush(vRes, pFans0[b])
				}
			}

		} else if Type == CBA_BOX_BUF {
			var nRangeMax int32 = Abc_MaxInt(nRange0, nRange)
			var pArg0 []int32 = Cba_VecLoadFanins(p, vTemp0, iFon0, pFans0, nRange0, nRangeMax, Cba_FonSigned(p, iFon0))
			for k = 0; k < nRange; k++ {
				Vec_IntPush(vRes, pArg0[k])
			}

		} else if Type >= CBA_BOX_CF && Type <= CBA_BOX_CZ {
			assert(false, "0")
			//word * pTruth = (word *)Cba_ObjFanins(p, i);
			//for ( k = 0; k < nRange; k++ )
			//    Vec_IntPush( vRes, Abc_TtGetBit(pTruth, k) );

		} else if Type == CBA_BOX_MUX || Type == CBA_BOX_NMUX {
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
			var iFinT int32
			var iFonT int32
			var fSigned int32 = 1
			assert(nRange0 >= 1 && Cba_ObjFinNum(p, i) >= 3, "nRange0 >= 1 && Cba_ObjFinNum(p, i) >= 3")
			assert(1+(1<<nRange0) == Cba_ObjFinNum(p, i), "1 + (1 << nRange0) == Cba_ObjFinNum(p, i)")
			for k = 0; k < Cba_ObjFinNum(p, i); k++ {
				iFinT = Cba_ObjFin0(p, i) + k
				iFonT = Cba_FinFon(p, iFinT)
				if k > 0 {
					fSigned &= Cba_FonSigned(p, iFonT)
				}
			}
			for b = 0; b < nRange; b++ {
				Vec_IntClear(vTemp0)
				for k = 0; k < Cba_ObjFinNum(p, i); k++ {
					iFinT = Cba_ObjFin0(p, i) + k
					iFonT = Cba_FinFon(p, iFinT)
					if k == 0 {
						continue
					}
					nRange1 = Cba_FonRangeSize(p, iFonT)
					if Cba_FonIsReal(iFonT) != 0 {
						pFans1 = Vec_IntArray(vBits)[Cba_FonCopy(p, iFonT):]
					} else {
						pFans1 = nil
					}
					var data int32
					if k < nRange1 {
						data = pFans1[k]
					} else if Cba_ObjFinNum(p, i) == 3 {
						if fSigned != 0 {
							data = pFans1[nRange1-1]
						}
					} else if Cba_FonSigned(p, iFonT) != 0 {
						data = pFans1[nRange1-1]
					}
					Vec_IntPush(vTemp0, data)
				}
				Vec_IntPush(vRes, Cba_NtkMuxTree_rec(pNew, pFans0, nRange0, vTemp0, 0))
			}

		} else if Type == CBA_BOX_SHIR || Type == CBA_BOX_SHIRA ||
			Type == CBA_BOX_SHIL || Type == CBA_BOX_SHILA {
			var nRangeMax int32 = Abc_MaxInt(nRange, nRange0)
			var pArg0 []int32 = Cba_VecLoadFanins(p, vTemp0, iFon0, pFans0, nRange0, nRangeMax, Cba_FonSigned(p, iFon0))
			if Type == CBA_BOX_SHIR || Type == CBA_BOX_SHIRA {
				Cba_BlastShiftRight(pNew, pArg0, nRangeMax, pFans1, nRange1, boolToInt(Cba_FonSigned(p, iFon0) != 0 && Type == CBA_BOX_SHIRA), vRes)
			} else {
				Cba_BlastShiftLeft(pNew, pArg0, nRangeMax, pFans1, nRange1, 0, vRes)
			}
			Vec_IntShrink(vRes, nRange)

		} else if Type == CBA_BOX_ROTR {
			assert(nRange0 == nRange, "nRange0 == nRange")
			Cba_BlastRotateRight(pNew, pFans0, nRange0, pFans1, nRange1, vRes)

		} else if Type == CBA_BOX_ROTL {
			assert(nRange0 == nRange, "nRange0 == nRange")
			Cba_BlastRotateLeft(pNew, pFans0, nRange0, pFans1, nRange1, vRes)

		} else if Type == CBA_BOX_INV {
			var nRangeMax int32 = Abc_MaxInt(nRange, nRange0)
			var pArg0 []int32 = Cba_VecLoadFanins(p, vTemp0, iFon0, pFans0, nRange0, nRangeMax, Cba_FonSigned(p, iFon0))
			for k = 0; k < nRange; k++ {
				Vec_IntPush(vRes, Abc_LitNot(pArg0[k]))
			}

		} else if Type == CBA_BOX_AND {
			var nRangeMax int32 = Abc_MaxInt(nRange, Abc_MaxInt(nRange0, nRange1))
			var pArg0 []int32 = Cba_VecLoadFanins(p, vTemp0, iFon0, pFans0, nRange0, nRangeMax, fSigned01)
			var pArg1 []int32 = Cba_VecLoadFanins(p, vTemp1, iFon1, pFans1, nRange1, nRangeMax, fSigned01)
			for k = 0; k < nRange; k++ {
				Vec_IntPush(vRes, Gia_ManHashAnd(pNew, pArg0[k], pArg1[k]))
			}

		} else if Type == CBA_BOX_OR {
			var nRangeMax int32 = Abc_MaxInt(nRange, Abc_MaxInt(nRange0, nRange1))
			var pArg0 []int32 = Cba_VecLoadFanins(p, vTemp0, iFon0, pFans0, nRange0, nRangeMax, fSigned01)
			var pArg1 []int32 = Cba_VecLoadFanins(p, vTemp1, iFon1, pFans1, nRange1, nRangeMax, fSigned01)
			for k = 0; k < nRange; k++ {
				Vec_IntPush(vRes, Gia_ManHashOr(pNew, pArg0[k], pArg1[k]))
			}

		} else if Type == CBA_BOX_XOR {
			var nRangeMax int32 = Abc_MaxInt(nRange, Abc_MaxInt(nRange0, nRange1))
			var pArg0 []int32 = Cba_VecLoadFanins(p, vTemp0, iFon0, pFans0, nRange0, nRangeMax, fSigned01)
			var pArg1 []int32 = Cba_VecLoadFanins(p, vTemp1, iFon1, pFans1, nRange1, nRangeMax, fSigned01)
			for k = 0; k < nRange; k++ {
				Vec_IntPush(vRes, Gia_ManHashXor(pNew, pArg0[k], pArg1[k]))
			}

		} else if Type == CBA_BOX_LNOT {
			iLit = Cba_BlastReduction(pNew, pFans0, nRange0, CBA_BOX_ROR)
			Vec_IntFill(vRes, 1, Abc_LitNot(iLit))
			for k = 1; k < nRange; k++ {
				Vec_IntPush(vRes, 0)
			}

		} else if Type == CBA_BOX_LAND {
			var iLit0 int32 = Cba_BlastReduction(pNew, pFans0, nRange0, CBA_BOX_ROR)
			var iLit1 int32 = Cba_BlastReduction(pNew, pFans1, nRange1, CBA_BOX_ROR)
			Vec_IntFill(vRes, 1, Gia_ManHashAnd(pNew, iLit0, iLit1))
			for k = 1; k < nRange; k++ {
				Vec_IntPush(vRes, 0)
			}

		} else if Type == CBA_BOX_LOR {
			var iLit0 int32 = Cba_BlastReduction(pNew, pFans0, nRange0, CBA_BOX_ROR)
			var iLit1 int32 = Cba_BlastReduction(pNew, pFans1, nRange1, CBA_BOX_ROR)
			Vec_IntFill(vRes, 1, Gia_ManHashOr(pNew, iLit0, iLit1))
			for k = 1; k < nRange; k++ {
				Vec_IntPush(vRes, 0)
			}

		} else if Type == CBA_BOX_LXOR {
			var iLit0 int32 = Cba_BlastReduction(pNew, pFans0, nRange0, CBA_BOX_ROR)
			var iLit1 int32 = Cba_BlastReduction(pNew, pFans1, nRange1, CBA_BOX_ROR)
			Vec_IntFill(vRes, 1, Gia_ManHashXor(pNew, iLit0, iLit1))
			for k = 1; k < nRange; k++ {
				Vec_IntPush(vRes, 0)
			}

		} else if Type == CBA_BOX_EQU || Type == CBA_BOX_NEQU {
			var iLit int32
			var nRangeMax int32 = Abc_MaxInt(nRange0, nRange1)
			var pArg0 []int32 = Cba_VecLoadFanins(p, vTemp0, iFon0, pFans0, nRange0, nRangeMax, fSigned01)
			var pArg1 []int32 = Cba_VecLoadFanins(p, vTemp1, iFon1, pFans1, nRange1, nRangeMax, fSigned01)
			for k = 0; k < nRangeMax; k++ {
				iLit = Gia_ManHashOr(pNew, iLit, Gia_ManHashXor(pNew, pArg0[k], pArg1[k]))
			}
			Vec_IntFill(vRes, 1, Abc_LitNotCond(iLit, boolToInt(Type == CBA_BOX_EQU)))
			for k = 1; k < nRange; k++ {
				Vec_IntPush(vRes, 0)
			}

		} else if Type == CBA_BOX_LTHAN || Type == CBA_BOX_METHAN ||
			Type == CBA_BOX_MTHAN || Type == CBA_BOX_LETHAN {
			var nRangeMax int32 = Abc_MaxInt(nRange0, nRange1)
			var pArg0 []int32 = Cba_VecLoadFanins(p, vTemp0, iFon0, pFans0, nRange0, nRangeMax, fSigned01)
			var pArg1 []int32 = Cba_VecLoadFanins(p, vTemp1, iFon1, pFans1, nRange1, nRangeMax, fSigned01)
			var fSwap int32 = boolToInt(Type == CBA_BOX_MTHAN || Type == CBA_BOX_LETHAN)
			var fCompl int32 = boolToInt(Type == CBA_BOX_METHAN || Type == CBA_BOX_LETHAN)
			if fSwap != 0 {
				pArg0, pArg1 = pArg1, pArg0
			}
			if fSigned01 != 0 {
				iLit = Cba_BlastLessSigned(pNew, pArg0, pArg1, nRangeMax)
			} else {
				iLit = Cba_BlastLess(pNew, pArg0, pArg1, nRangeMax)
			}
			iLit = Abc_LitNotCond(iLit, fCompl)
			Vec_IntFill(vRes, 1, iLit)
			for k = 1; k < nRange; k++ {
				Vec_IntPush(vRes, 0)
			}

		} else if Type == CBA_BOX_RAND || Type == CBA_BOX_ROR || Type == CBA_BOX_RXOR {
			Vec_IntPush(vRes, Cba_BlastReduction(pNew, pFans0, nRange0, Type))
			for k = 1; k < nRange; k++ {
				Vec_IntPush(vRes, 0)
			}

		} else if Type == CBA_BOX_ADD {
			var nRangeMax int32 = Abc_MaxInt(nRange, Abc_MaxInt(nRange1, nRange2))
			var pArg0 []int32 = Cba_VecLoadFanins(p, vTemp0, iFon0, pFans0, 1, 1, 0)
			var pArg1 []int32 = Cba_VecLoadFanins(p, vRes, iFon1, pFans1, nRange1, nRangeMax, fSigned01)
			var pArg2 []int32 = Cba_VecLoadFanins(p, vTemp1, iFon2, pFans2, nRange2, nRangeMax, fSigned01)
			var Carry int32 = Cba_BlastAdder(pNew, pArg0[0], pArg1, pArg2, nRange) // result is in pArg1 (vRes)
			assert(nRange0 == 1, "nRange0 == 1")
			Vec_IntShrink(vRes, nRange)
			Vec_IntPush(vRes, Carry)

		} else if Type == CBA_BOX_SUB {
			var nRangeMax int32 = Abc_MaxInt(nRange, Abc_MaxInt(nRange0, nRange1))
			var pArg0 []int32 = Cba_VecLoadFanins(p, vRes, iFon0, pFans0, nRange0, nRangeMax, fSigned01)
			var pArg1 []int32 = Cba_VecLoadFanins(p, vTemp1, iFon1, pFans1, nRange1, nRangeMax, fSigned01)
			Cba_BlastSubtract(pNew, pArg0, pArg1, nRange) // result is in pFan0 (vRes)
			Vec_IntShrink(vRes, nRange)

		} else if Type == CBA_BOX_MUL {
			if fUseOldMultiplierBlasting != 0 {
				var nRangeMax int32 = Abc_MaxInt(nRange, Abc_MaxInt(nRange0, nRange1))
				var pArg0 []int32 = Cba_VecLoadFanins(p, vTemp0, iFon0, pFans0, nRange0, nRangeMax, fSigned01)
				var pArg1 []int32 = Cba_VecLoadFanins(p, vTemp1, iFon1, pFans1, nRange1, nRangeMax, fSigned01)
				Cba_BlastMultiplier2(pNew, pArg0, pArg1, nRange, vTemp2, vRes)
				Vec_IntShrink(vRes, nRange)

			} else {
				var nRangeMax int32 = Abc_MaxInt(nRange0, nRange1)
				var pArg0 []int32 = Cba_VecLoadFanins(p, vTemp0, iFon0, pFans0, nRange0, nRangeMax, fSigned01)
				var pArg1 []int32 = Cba_VecLoadFanins(p, vTemp1, iFon1, pFans1, nRange1, nRangeMax, fSigned01)
				Cba_BlastMultiplier(pNew, pArg0, pArg1, nRangeMax, nRangeMax, vTemp2, vRes, fSigned01)
				if nRange > nRangeMax+nRangeMax {
					fill := int32(0)
					if fSigned01 != 0 {
						fill = Vec_IntEntryLast(vRes)
					}
					Vec_IntFillExtra(vRes, nRange, fill)
				} else {
					Vec_IntShrink(vRes, nRange)
				}
				assert(Vec_IntSize(vRes) == nRange, "Vec_IntSize(vRes) == nRange")
			}

		} else if Type == CBA_BOX_DIV || Type == CBA_BOX_MOD {
			var nRangeMax int32 = Abc_MaxInt(nRange, Abc_MaxInt(nRange0, nRange1))
			var pArg0 []int32 = Cba_VecLoadFanins(p, vTemp0, iFon0, pFans0, nRange0, nRangeMax, fSigned01)
			var pArg1 []int32 = Cba_VecLoadFanins(p, vTemp1, iFon1, pFans1, nRange1, nRangeMax, fSigned01)
			if fSigned01 != 0 {
				Cba_BlastDividerSigned(pNew, pArg0, nRangeMax, pArg1, nRangeMax, boolToInt(Type == CBA_BOX_DIV), vRes)
			} else {
				Cba_BlastDivider(pNew, pArg0, nRangeMax, pArg1, nRangeMax, boolToInt(Type == CBA_BOX_DIV), vRes)
			}
			Vec_IntShrink(vRes, nRange)
			if Type == CBA_BOX_DIV {
				Cba_BlastZeroCondition(pNew, pFans1, nRange1, vRes)
			}

		} else if Type == CBA_BOX_MIN {
			var nRangeMax int32 = Abc_MaxInt(nRange0, nRange)
			var pArg0 []int32 = Cba_VecLoadFanins(p, vTemp0, iFon0, pFans0, nRange0, nRangeMax, Cba_FonSigned(p, iFon0))
			Cba_BlastMinus(pNew, pArg0, nRangeMax, vRes)
			Vec_IntShrink(vRes, nRange)

		} else if Type == CBA_BOX_POW {
			var nRangeMax int32 = Abc_MaxInt(nRange0, nRange)
			var pArg0 []int32 = Cba_VecLoadFanins(p, vTemp0, iFon0, pFans0, nRange0, nRangeMax, Cba_FonSigned(p, iFon0))
			var pArg1 []int32 = Cba_VecLoadFanins(p, vTemp1, iFon1, pFans1, nRange1, nRange1, Cba_FonSigned(p, iFon1))
			Cba_BlastPower(pNew, pArg0, nRangeMax, pArg1, nRange1, vTemp2, vRes)
			Vec_IntShrink(vRes, nRange)

		} else if Type == CBA_BOX_SQRT {
			var pArg0 []int32 = Cba_VecLoadFanins(p, vTemp0, iFon0, pFans0, nRange0, nRange0+(nRange0&1), 0)
			nRange0 += (nRange0 & 1)
			Cba_BlastSqrt(pNew, pArg0, nRange0, vTemp2, vRes)
			if nRange > Vec_IntSize(vRes) {
				Vec_IntFillExtra(vRes, nRange, 0)
			} else {
				Vec_IntShrink(vRes, nRange)
			}

		} else if Type == CBA_BOX_TABLE {
			assert(false, "0")
			//Cba_BlastTable( pNew, Cba_ObjTable(p, p, i), pFans0, nRange0, nRange, vRes );

		} else {
			assert(false, "0")
		}
		Vec_IntAppend(vBits, vRes)
		p.pDesign.nAnds[Type] += Gia_ManAndNum(pNew) - nAndPrev
	}
	assert(nBits == Vec_IntSize(vBits), "nBits == Vec_IntSize(vBits)")
	p.pDesign.nAnds[0] = Gia_ManAndNum(pNew)
	// create COs
	for poIndex := int32(0); poIndex < Cba_NtkPoNum(p); poIndex++ {
		iObj = Cba_NtkPo(p, poIndex)
		for k = 0; k < Cba_ObjFinNum(p, iObj); k++ {
			iFin = Cba_ObjFin0(p, iObj) + k
			iFon = Cba_FinFon(p, iFin)
			nRange = Cba_FonRangeSize(p, iFon)
			if Cba_FonIsReal(iFon) != 0 {
				pFans0 = Vec_IntArray(vBits)[Cba_FonCopy(p, iFon):]
			} else {
				pFans0 = nil
			}
			pFans0 = Cba_VecLoadFanins(p, vTemp0, iFon, pFans0, nRange, nRange, Cba_FonSigned(p, iFon))
			for b = 0; b < nRange; b++ {
				Gia_ManAppendCo(pNew, pFans0[b])
			}
		}
	}
	for seqIndex := int32(0); seqIndex < Cba_NtkBoxSeqNum(p); seqIndex++ {
		iObj = Cba_NtkBoxSeq(p, seqIndex)
		if fSeq != 0 {
			assert(Cba_ObjType(p, iObj) == CBA_BOX_DFFCPL, "Cba_ObjType(p, iObj) == CBA_BOX_DFFCPL")
			iFon0 = Cba_ObjFinFon(p, iObj, 0)
			iFon1 = Cba_ObjFinFon(p, iObj, 1)
			nRange0 = Cba_FonRangeSize(p, iFon0)
			nRange1 = Cba_FonRangeSize(p, iFon1)
			assert(nRange0 == nRange1, "nRange0 == nRange1")
			for k = 0; k < Cba_ObjFinNum(p, iObj); k++ {
				iFin = Cba_ObjFin0(p, iObj) + k
				iFon = Cba_FinFon(p, iFin)
				nRange = Cba_FonRangeSize(p, iFon)
				if Cba_FonIsReal(iFon) != 0 {
					pFans0 = Vec_IntArray(vBits)[Cba_FonCopy(p, iFon):]
				} else {
					pFans0 = nil
				}
				pFans0 = Cba_VecLoadFanins(p, vTemp0, iFon, pFans0, nRange0, nRange0, Cba_FonSigned(p, iFon))
				if k == 0 {
					for b = 0; b < nRange; b++ {
						Gia_ManAppendCo(pNew, pFans0[b])
					}

				} else if k == 1 {
					for b = 0; b < nRange; b++ {
						if pFans0[b] == 0 {
							Vec_StrPush(vInit, '0')
						} else if pFans0[b] == 1 {
							Vec_StrPush(vInit, '1')
						} else {
							Vec_StrPush(vInit, 'x')
						}
					}

				} else {
					break
				}
			}

		} else { // combinational

			for k = 0; k < Cba_ObjFinNum(p, iObj); k++ {
				iFin = Cba_ObjFin0(p, iObj) + k
				iFon = Cba_FinFon(p, iFin)
				nRange = Cba_FonRangeSize(p, iFon)
				if Cba_FonIsReal(iFon) != 0 {
					pFans0 = Vec_IntArray(vBits)[Cba_FonCopy(p, iFon):]
				} else {
					pFans0 = nil
				}
				pFans0 = Cba_VecLoadFanins(p, vTemp0, iFon, pFans0, nRange, nRange, Cba_FonSigned(p, iFon))
				for b = 0; b < nRange; b++ {
					Gia_ManAppendCo(pNew, pFans0[b])
				}
			}
		}
	}
	Vec_IntFree(vTemp0)
	Vec_IntFree(vTemp1)
	Vec_IntFree(vTemp2)
	Vec_IntFree(vRes)
	// finalize AIG
	pTemp = pNew
	pNew = Gia_ManCleanup(pTemp)
	Gia_ManDupRemapLiterals(vBits, pTemp)
	Gia_ManStop(pTemp)
	// transform AIG with init state
	if fSeq != 0 {
		Gia_ManSetRegNum(pNew, Vec_StrSize(vInit))
		Vec_StrPush(vInit, byte(0))
		pTemp = pNew
		pNew = Gia_ManDupZeroUndc(pTemp, string(Vec_StrArray(vInit)), 0, 0, 1)
		Gia_ManDupRemapLiterals(vBits, pTemp)
		Gia_ManStop(pTemp)
		Vec_StrFreeP(&vInit)
	}
	//Vec_IntErase( vBits );
	//Vec_IntErase( &p->vCopies );
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
func Cba_ManBlast(p *Cba_Man_t, fBarBufs int32, fSeq int32, fVerbose int32) *Gia_Man_t {
	return Cba_NtkBlast(Cba_ManRoot(p), fSeq)
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Cba_ManInsertGia(p *Cba_Man_t, pGia *Gia_Man_t) *Cba_Man_t {
	return nil
}
func Cba_ManInsertAbc(p *Cba_Man_t, pAbc any) *Cba_Man_t {
	manager, _ := pAbc.(*Cba_Man_t)
	return manager
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END
