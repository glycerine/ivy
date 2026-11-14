package wlc

import (
	"fmt"
	"strconv"
	"strings"
)

import "os"

/**CFile****************************************************************

  FileName    [wlcAbc.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Verilog parser.]

  Synopsis    [Parses several flavors of word-level Verilog.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - August 22, 2014.]

  Revision    [$Id: wlcAbc.c,v 1.00 2014/09/12 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "wlc.h"
//#include "base/abc/abc.h"

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
func Wlc_NtkPrintInputInfo(pNtk *Wlc_Ntk_t) {
	var pObj *Wlc_Obj_t
	var i, k, nRange, nBeg, nEnd, nBits int32
	var output *os.File

	output, _ = os.OpenFile("abc_blast_input.info", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)

	// inlining of macro Wlc_NtkForEachCi
	for i = 0; i < Wlc_NtkCiNum(pNtk); i++ {
		pObj = Wlc_NtkCi(pNtk, i)
		nRange = Wlc_ObjRange(pObj)
		nBeg = pObj.Beg
		nEnd = pObj.End

		for k = 0; k < nRange; k++ {
			var index int32
			if nEnd > nBeg {
				index = nBeg + k
			} else {
				index = nEnd + k
			}
			var c byte
			if int32(pObj.Type) != int32(WLC_OBJ_FO) {
				c = 'i'
			} else {
				c = pNtk.pInits[nBits+k]
			}
			fmt.Fprintf(output, "%s[%d] : %c \n", Wlc_ObjName(pNtk, Wlc_ObjId(pNtk, pObj)), index, c)
		}
		if int32(pObj.Type) == int32(WLC_OBJ_FO) {
			nBits += nRange
		}
	}

	// inlining of macro Wlc_NtkForEachPo
	for i = 0; i < Wlc_NtkPoNum(pNtk); i++ {
		pObj = Wlc_NtkPo(pNtk, i)
		nRange = Wlc_ObjRange(pObj)
		nBeg = pObj.Beg
		nEnd = pObj.End

		for k = 0; k < nRange; k++ {
			var index int32
			if nEnd > nBeg {
				index = nBeg + k
			} else {
				index = nEnd + k
			}
			fmt.Fprintf(output, "%s[%d] : o \n", Wlc_ObjName(pNtk, Wlc_ObjId(pNtk, pObj)), index)
		}
	}

	_ = (output).Close()
	return
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Wlc_NtkPrintInvStats(pNtk *Wlc_Ntk_t, vCounts *Vec_Int_t, fVerbose int32) {
	var pObj *Wlc_Obj_t
	var i, k, nNum, nRange, nBits int32
	// inlining of macro Wlc_NtkForEachCi
	for i = 0; i < Wlc_NtkCiNum(pNtk); i++ {
		pObj = Wlc_NtkCi(pNtk, i)
		if int32(pObj.Type) != int32(WLC_OBJ_FO) {
			continue
		}
		nRange = Wlc_ObjRange(pObj)
		for k = 0; k < nRange; k++ {
			nNum = Vec_IntEntry(vCounts, nBits+k)
			if nNum != 0 {
				break
			}
		}
		if k == nRange {
			nBits += nRange
			continue
		}
		fmt.Printf("%s[%d:%d] : ", Wlc_ObjName(pNtk, Wlc_ObjId(pNtk, pObj)), pObj.End, pObj.Beg)
		for k = 0; k < nRange; k++ {
			nNum = Vec_IntEntry(vCounts, nBits+k)
			if nNum == 0 {
				continue
			}
			fmt.Printf("  [%d] -> %d", k, nNum)
		}
		fmt.Printf("\n")
		nBits += nRange
	}
	//printf( "%d %d\n", Vec_IntSize(vCounts), nBits );
	assert(Vec_IntSize(vCounts) == nBits, "Vec_IntSize(vCounts) == nBits")
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Wlc_NtkGetInv(pNtk *Wlc_Ntk_t, vInv *Vec_Int_t) *Abc_Ntk_t {
	// extern Vec_Int_t * Pdr_InvCounts( Vec_Int_t * vInv );
	// extern Vec_Str_t * Pdr_InvPrintStr( Vec_Int_t * vInv, Vec_Int_t * vCounts );

	var vCounts *Vec_Int_t = Pdr_InvCounts(vInv)
	var vSop *Vec_Str_t = Pdr_InvPrintStr(vInv, vCounts)

	var pObj *Wlc_Obj_t
	var i, k, nNum, nRange, nBits int32
	var pMainNtk *Abc_Ntk_t = nil
	var pMainObj, pMainTemp *Abc_Obj_t
	var Buffer string // was [5000]byte in C

	// start the network
	pMainNtk = Abc_NtkAlloc(ABC_NTK_LOGIC, ABC_FUNC_SOP, 1)
	// duplicate the name and the spec
	if pNtk != nil {
		pMainNtk.Name = strings.Clone(pNtk.pName)
	} else {
		pMainNtk.Name = "inv"
	}
	// create primary inputs
	if pNtk == nil {
		var Entry int32
		var nInputs int32 = Abc_SopGetVarNum(string(Vec_StrArray(vSop)))
		for i = 0; i < Vec_IntSize(vCounts); i++ {
			Entry = Vec_IntEntry(vCounts, i)
			if Entry == 0 {
				continue
			}
			pMainObj = Abc_NtkCreatePi(pMainNtk)
			Buffer = fmt.Sprintf("pi%d", i)
			Abc_ObjAssignName(pMainObj, Buffer, "")
		}
		if Abc_NtkPiNum(pMainNtk) != nInputs {
			fmt.Printf("Mismatch between number of inputs and the number of literals in the invariant.\n")
			Abc_NtkDelete(pMainNtk)
			return nil
		}

	} else {
		// inlining of macro Wlc_NtkForEachCi
		for i = 0; i < Wlc_NtkCiNum(pNtk); i++ {
			pObj = Wlc_NtkCi(pNtk, i)
			if int32(pObj.Type) != int32(WLC_OBJ_FO) {
				continue
			}
			nRange = Wlc_ObjRange(pObj)
			for k = 0; k < nRange; k++ {
				nNum = Vec_IntEntry(vCounts, nBits+k)
				if nNum != 0 {
					break
				}
			}
			if k == nRange {
				nBits += nRange
				continue
			}
			//printf( "%s[%d:%d] : ", Wlc_ObjName(pNtk, Wlc_ObjId(pNtk, pObj)), pObj->End, pObj->Beg );
			for k = 0; k < nRange; k++ {
				nNum = Vec_IntEntry(vCounts, nBits+k)
				if nNum == 0 {
					continue
				}
				//printf( "  [%d] -> %d", k, nNum );
				pMainObj = Abc_NtkCreatePi(pMainNtk)
				Buffer = fmt.Sprintf("%s[%d]", Wlc_ObjName(pNtk, Wlc_ObjId(pNtk, pObj)), k)
				Abc_ObjAssignName(pMainObj, Buffer, "")

			}
			//printf( "\n");
			nBits += nRange
		}
	}
	//printf( "%d %d\n", Vec_IntSize(vCounts), nBits );
	assert(pNtk == nil || Vec_IntSize(vCounts) == nBits, "pNtk == nil || Vec_IntSize(vCounts) == nBits")
	// create node
	pMainObj = Abc_NtkCreateNode(pMainNtk)
	for i = 0; i < Abc_NtkPiNum(pMainNtk); i++ {
		Abc_ObjAddFanin(pMainObj, Abc_NtkPi(pMainNtk, i))
	}
	pMainObj.Data = Abc_SopRegister(pMainNtk.FunctionManager.(*Mem_Flex_t), string(Vec_StrArray(vSop)))
	Vec_IntFree(vCounts)
	Vec_StrFree(vSop)
	// create PO
	pMainTemp = Abc_NtkCreatePo(pMainNtk)
	Abc_ObjAddFanin(pMainTemp, pMainObj)
	Abc_ObjAssignName(pMainTemp, "inv", "")
	return pMainNtk
}

/*
*Function*************************************************************

	Synopsis    [Translate current network into an invariant.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Wlc_NtkGetPut(pNtk *Abc_Ntk_t, pGia *Gia_Man_t) *Vec_Int_t {
	var nRegs int32 = Gia_ManRegNum(pGia)
	var vRes *Vec_Int_t = nil
	if Abc_NtkPoNum(pNtk) != 1 {
		fmt.Printf("The number of outputs is other than 1.\n")
	} else if Abc_NtkNodeNum(pNtk) != 1 {
		fmt.Printf("The number of internal nodes is other than 1.\n")
	} else {
		var pNames *Abc_Nam_t = nil
		var pFanin *Abc_Obj_t
		var pNode *Abc_Obj_t = Abc_ObjFanin0(Abc_NtkCo(pNtk, 0))
		var pName string
		var pCube string
		var pSop string = pNode.Data.(string)
		var vFanins *Vec_Int_t = Vec_IntAlloc(Abc_ObjFaninNum(pNode))
		var i, k, Value, nLits, Counter int32
		if pGia.vNamesIn != nil {
			// hash the names
			pNames = Abc_NamStart(100, 16)
			for i = 0; i < Vec_PtrSize(pGia.vNamesIn); i++ {
				pName = Vec_PtrEntry(pGia.vNamesIn, i).(string)
				Value = Abc_NamStrFindOrAdd(pNames, pName, nil)
				assert(Value == i+1, "Value == i+1")
				//printf( "%s(%d) ", pName, i );
			}
			//printf( "\n" );
		}
		for i = 0; i < Abc_ObjFaninNum(pNode); i++ {
			pFanin = Abc_ObjFanin(pNode, i)
			assert(Abc_ObjIsCi(pFanin) != 0, "Abc_ObjIsCi(pFanin)")
			pName = Abc_ObjName(pFanin)
			if pNames != nil {
				Value = Abc_NamStrFind(pNames, pName) - 1 - Gia_ManPiNum(pGia)
				if Value < 0 {
					if Counter == 0 {
						fmt.Printf("Cannot read input name \"%s\" of fanin %d.\n", pName, i)
					}
					Counter++
					Value = i
				}

			} else {
				for k = int32(len(pName)) - 1; k >= 0; k-- {
					if pName[k] < '0' || pName[k] > '9' {
						break
					}
				}
				if k == int32(len(pName))-1 {
					if Counter == 0 {
						fmt.Printf("Cannot read input name \"%s\" of fanin %d.\n", pName, i)
					}
					Counter++
					Value = i

				} else {
					Value, _ = strconv.Atoi(pName[k+1:])
				}
			}
			Vec_IntPush(vFanins, Value)
		}
		if Counter != 0 {
			fmt.Printf("Cannot read names for %d inputs of the invariant.\n", Counter)
		}
		if pNames != nil {
			Abc_NamStop(pNames)
		}
		assert(Vec_IntSize(vFanins) == Abc_ObjFaninNum(pNode), "Vec_IntSize(vFanins) == Abc_ObjFaninNum(pNode)")
		vRes = Vec_IntAlloc(1000)
		Vec_IntPush(vRes, Abc_SopGetCubeNum(pSop))
		for _, cube := range strings.Split(strings.TrimSpace(pSop), "\n") {
			if cube == "" {
				continue
			}
			nLits = 0
			for k = 0; k < int32(Abc_ObjFaninNum(pNode)); k++ {
				Value = int32(cube[k])
				if Value != '-' {
					nLits++
				}
			}
			Vec_IntPush(vRes, nLits)
			for k = 0; k < int32(Abc_ObjFaninNum(pNode)); k++ {
				Value = int32(cube[k])
				if Value != '-' {
					Vec_IntPush(vRes, Abc_Var2Lit(Vec_IntEntry(vFanins, k), Value == '0'))
				}
			}
		}
		Vec_IntPush(vRes, nRegs)
		Vec_IntFree(vFanins)
	}
	return vRes
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END
