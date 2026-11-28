package abci

import (
	"fmt"
	"os"
	"time"

	baseabc "github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
)

type ABC_INT64_T = int64

func Vec_IntAlloc(n int32) *Vec_Int_t                  { return vec.Vec_IntAlloc(int(n)) }
func Vec_IntFree(v *Vec_Int_t)                         { vec.Vec_IntFree(v) }
func Vec_IntSize(v *Vec_Int_t) int32                   { return int32(vec.Vec_IntSize(v)) }
func Vec_IntEntry(v *Vec_Int_t, i int32) int32         { return vec.Vec_IntEntry(v, int(i)) }
func Vec_IntEntryLast(v *Vec_Int_t) int32              { return vec.Vec_IntEntryLast(v) }
func Vec_IntFind(v *Vec_Int_t, x int32) int            { return vec.Vec_IntFind(v, x) }
func Vec_IntPush(v *Vec_Int_t, x int32)                { vec.Vec_IntPush(v, x) }
func Vec_IntPushOrder(v *Vec_Int_t, x int32)           { vec.Vec_IntPushOrder(v, x) }
func Vec_IntPushUnique(v *Vec_Int_t, x int32) int      { return vec.Vec_IntPushUnique(v, x) }
func Vec_IntPushUniqueOrder(v *Vec_Int_t, x int32) int { return vec.Vec_IntPushUniqueOrder(v, x) }
func Vec_IntWriteEntry(v *Vec_Int_t, i, x int32)       { vec.Vec_IntWriteEntry(v, int(i), x) }
func Vec_IntPop(v *Vec_Int_t) int32                    { return vec.Vec_IntPop(v) }
func Vec_PtrAlloc(n int32) *Vec_Ptr_t                  { return vec.Vec_PtrAlloc(int(n)) }
func Vec_PtrFree(v *Vec_Ptr_t)                         { vec.Vec_PtrFree(v) }
func Vec_PtrSize(v *Vec_Ptr_t) int32                   { return int32(vec.Vec_PtrSize(v)) }
func Vec_PtrEntry(v *Vec_Ptr_t, i int32) any           { return vec.Vec_PtrEntry(v, int(i)) }
func Vec_PtrPush(v *Vec_Ptr_t, x any)                  { vec.Vec_PtrPush(v, x) }
func Vec_PtrPop(v *Vec_Ptr_t) any                      { return vec.Vec_PtrPop(v) }
func Vec_PtrClear(v *Vec_Ptr_t)                        { vec.Vec_PtrClear(v) }
func Abc_NtkPiNum(n *Abc_Ntk_t) int32                  { return baseabc.Abc_NtkPiNum(n) }
func Abc_NtkPoNum(n *Abc_Ntk_t) int32                  { return baseabc.Abc_NtkPoNum(n) }
func Abc_NtkPi(n *Abc_Ntk_t, i int32) *Abc_Obj_t       { return baseabc.Abc_NtkPi(n, i) }
func Abc_NtkPo(n *Abc_Ntk_t, i int32) *Abc_Obj_t       { return baseabc.Abc_NtkPo(n, i) }
func Abc_NtkCiNum(n *Abc_Ntk_t) int32                  { return baseabc.Abc_NtkCiNum(n) }
func Abc_NtkCoNum(n *Abc_Ntk_t) int32                  { return baseabc.Abc_NtkCoNum(n) }
func Abc_NtkLatchNum(n *Abc_Ntk_t) int32               { return baseabc.Abc_NtkLatchNum(n) }
func Abc_NtkDelete(n *Abc_Ntk_t)                       { baseabc.Abc_NtkDelete(n) }
func Abc_AigLevel(n *Abc_Ntk_t) int32                  { return baseabc.Abc_AigLevel(n) }

/**CFile****************************************************************

  FileName    [bm.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Boolean Matching package.]

  Synopsis    [Check P-equivalence and PP-equivalence of two circuits.]

  Author      [Hadi Katebi]

  Affiliation [University of Michigan]

  Date        [Ver. 1.0. Started - January, 2009.]

  Revision    [No revisions so far]

  Comments    [This is the cleaned up version of the code I used for DATE 2010 publication.]
              [If you have any question or if you find a bug, contact me at hadik@umich.edu.]
              [I don't guarantee that I can fix all the bugs, but I can definitely point you to
               the right direction so you can fix the bugs yourself].

  Debugging   [There are some part of the code that are commented out. Those parts mostly print
               the contents of the data structures to the standard output. Un-comment them if you
               find them useful for debugging.]

***********************************************************************/

// This is a line-by-line port of src/base/abci/abcBm.c. Two C idioms
// need an explicit model in Go:
//
//   - Abc_Obj_t's pCopy/iTemp union holds simulation values as (sign
//     extended) ints. bmCopyRead/bmCopyWrite/bmTempWrite keep that value
//     in CopyData as an int64, so the C int arithmetic (including the
//     garbage values matchNonSingletonOutputs simulates, see there) is
//     reproduced exactly.
//   - computeCofactor tags pCopy with NULL and (void*)1 for constant
//     values; bmCofPtr models such a tagged pointer.

const FALSE = 0
const TRUE = 1

// bmCopyRead is (ABC_PTRINT_T)pObj->pCopy for a simulation value: the
// value stored by bmCopyWrite/bmTempWrite, or the bool the shared
// Abc_NtkVerifySimulatePattern stores (saucy simulates with it).
func bmCopyRead(o *Abc_Obj_t) int64 {
	switch v := o.CopyData.(type) {
	case int64:
		return v
	case bool:
		if v {
			return 1
		}
	}
	return 0
}

// bmCopyWrite is pObj->pCopy = (Abc_Obj_t *)(ABC_PTRINT_T)value.
func bmCopyWrite(o *Abc_Obj_t, value int64) { o.CopyData = value }

// bmTempWrite is pObj->iTemp = value: it replaces the low 32 bits of the
// pCopy/iTemp union and keeps the high ones.
func bmTempWrite(o *Abc_Obj_t, value int32) {
	o.CopyData = int64(uint64(bmCopyRead(o))&^0xffffffff | uint64(uint32(value)))
}

// bmNtkVerifySimulatePattern is Abc_NtkVerifySimulatePattern (abcVerify.c)
// with C's int values: pModel entries need not be 0 or 1 (see
// matchNonSingletonOutputs), and the AND of two values is their bitwise
// AND. pNtk is strashed.
func bmNtkVerifySimulatePattern(pNtk *Abc_Ntk_t, pModel []int32) []int32 {
	var pNode *Abc_Obj_t
	var Value0, Value1, i int32
	// increment the trav ID
	baseabc.Abc_NtkIncrementTravId(pNtk)
	// set the CI values
	bmCopyWrite(baseabc.Abc_AigConst1(pNtk).Object, 1)
	for i = 0; i < Abc_NtkCiNum(pNtk); i++ {
		bmCopyWrite(baseabc.Abc_NtkCi(pNtk, i), int64(pModel[i]))
	}
	// simulate in the topological order
	for i = 0; i < Abc_NtkObjNumMax(pNtk); i++ {
		pNode = Abc_NtkObj(pNtk, i)
		if pNode == nil || Abc_ObjIsNode(pNode) == 0 {
			continue
		}
		Value0 = int32(bmCopyRead(Abc_ObjFanin0(pNode))) ^ baseabc.Abc_ObjFaninC0(pNode)
		Value1 = int32(bmCopyRead(Abc_ObjFanin1(pNode))) ^ baseabc.Abc_ObjFaninC1(pNode)
		bmCopyWrite(pNode, int64(Value0&Value1))
	}
	// fill the output values
	pValues := make([]int32, Abc_NtkCoNum(pNtk))
	for i = 0; i < Abc_NtkCoNum(pNtk); i++ {
		pNode = baseabc.Abc_NtkCo(pNtk, i)
		pValues[i] = int32(bmCopyRead(Abc_ObjFanin0(pNode))) ^ baseabc.Abc_ObjFaninC0(pNode)
	}
	return pValues
}

// int match1by1(...);
// int Abc_NtkBmSat(...);

func getDependencies(pNtk *Abc_Ntk_t, iDep, oDep []*Vec_Int_t) {
	var vSuppFun *Vec_Ptr_t
	var i, j int32

	vSuppFun = Sim_ComputeFunSupp(pNtk, 0)
	for i = 0; i < Abc_NtkPoNum(pNtk); i++ {
		seg := Vec_PtrEntry(vSuppFun, i).([]uint32)

		// the C code reads the support bytes of the first PiNum inputs
		for j = 0; j < Abc_NtkPiNum(pNtk); j++ {
			if symHasBit(seg, j) {
				Vec_IntPushOrder(oDep[i], j)
			}
		}
	}

	for i = 0; i < Abc_NtkPoNum(pNtk); i++ {
		for j = 0; j < Vec_IntSize(oDep[i]); j++ {
			Vec_IntPush(iDep[Vec_IntEntry(oDep[i], j)], i)
		}
	}
}

func initMatchList(pNtk *Abc_Ntk_t, iDep, oDep []*Vec_Int_t, iMatch []*Vec_Int_t, iLastItem *int32, oMatch []*Vec_Int_t, oLastItem *int32, iGroup, oGroup []int32, p_equivalence int32) {
	var i, j, curr int32
	var temp []*Vec_Int_t

	if p_equivalence == 0 {
		temp = make([]*Vec_Int_t, Abc_NtkPiNum(pNtk)+1)

		for i = 0; i < Abc_NtkPiNum(pNtk)+1; i++ {
			temp[i] = Vec_IntAlloc(0)
		}

		for i = 0; i < Abc_NtkPoNum(pNtk); i++ {
			Vec_IntPush(temp[Vec_IntSize(oDep[i])], i)
		}

		curr = 0
		for i = 0; i < Abc_NtkPiNum(pNtk)+1; i++ {
			if Vec_IntSize(temp[i]) == 0 {
				Vec_IntFree(temp[i])
			} else {
				oMatch[curr] = temp[i]

				for j = 0; j < Vec_IntSize(temp[i]); j++ {
					oGroup[Vec_IntEntry(oMatch[curr], j)] = curr
				}

				curr++
			}
		}

		*oLastItem = curr
	} else {
		// the else part fixes the outputs for P-equivalence checking
		for i = 0; i < Abc_NtkPoNum(pNtk); i++ {
			Vec_IntPush(oMatch[i], i)
			oGroup[i] = i
			*oLastItem = Abc_NtkPoNum(pNtk)
		}
	}

	//////////////////////////////////////////////////////////////////////////////

	temp = make([]*Vec_Int_t, Abc_NtkPoNum(pNtk)+1)

	for i = 0; i < Abc_NtkPoNum(pNtk)+1; i++ {
		temp[i] = Vec_IntAlloc(0)
	}

	for i = 0; i < Abc_NtkPiNum(pNtk); i++ {
		Vec_IntPush(temp[Vec_IntSize(iDep[i])], i)
	}

	curr = 0
	for i = 0; i < Abc_NtkPoNum(pNtk)+1; i++ {
		if Vec_IntSize(temp[i]) == 0 {
			Vec_IntFree(temp[i])
		} else {
			iMatch[curr] = temp[i]
			for j = 0; j < Vec_IntSize(iMatch[curr]); j++ {
				iGroup[Vec_IntEntry(iMatch[curr], j)] = curr
			}
			curr++
		}
	}

	*iLastItem = curr
}

func iSortDependencies(pNtk *Abc_Ntk_t, iDep []*Vec_Int_t, oGroup []int32) {
	var i, j, k int32
	var temp *Vec_Int_t
	var oGroupList *Vec_Int_t

	oGroupList = Vec_IntAlloc(10)

	for i = 0; i < Abc_NtkPiNum(pNtk); i++ {
		if Vec_IntSize(iDep[i]) == 1 {
			continue
		}

		temp = Vec_IntAlloc(Vec_IntSize(iDep[i]))

		for j = 0; j < Vec_IntSize(iDep[i]); j++ {
			Vec_IntPushUniqueOrder(oGroupList, oGroup[Vec_IntEntry(iDep[i], j)])
		}

		for j = 0; j < Vec_IntSize(oGroupList); j++ {
			for k = 0; k < Vec_IntSize(iDep[i]); k++ {
				if oGroup[Vec_IntEntry(iDep[i], k)] == Vec_IntEntry(oGroupList, j) {
					Vec_IntPush(temp, Vec_IntEntry(iDep[i], k))
					vec.Vec_IntRemove(iDep[i], Vec_IntEntry(iDep[i], k))
					k--
				}
			}
		}

		Vec_IntFree(iDep[i])
		iDep[i] = temp
		vec.Vec_IntClear(oGroupList)
	}

	Vec_IntFree(oGroupList)
}

func oSortDependencies(pNtk *Abc_Ntk_t, oDep []*Vec_Int_t, iGroup []int32) {
	var i, j, k int32
	var temp *Vec_Int_t
	var iGroupList *Vec_Int_t

	iGroupList = Vec_IntAlloc(10)

	for i = 0; i < Abc_NtkPoNum(pNtk); i++ {
		if Vec_IntSize(oDep[i]) == 1 {
			continue
		}

		temp = Vec_IntAlloc(Vec_IntSize(oDep[i]))

		for j = 0; j < Vec_IntSize(oDep[i]); j++ {
			Vec_IntPushUniqueOrder(iGroupList, iGroup[Vec_IntEntry(oDep[i], j)])
		}

		for j = 0; j < Vec_IntSize(iGroupList); j++ {
			for k = 0; k < Vec_IntSize(oDep[i]); k++ {
				if iGroup[Vec_IntEntry(oDep[i], k)] == Vec_IntEntry(iGroupList, j) {
					Vec_IntPush(temp, Vec_IntEntry(oDep[i], k))
					vec.Vec_IntRemove(oDep[i], Vec_IntEntry(oDep[i], k))
					k--
				}
			}
		}

		Vec_IntFree(oDep[i])
		oDep[i] = temp
		vec.Vec_IntClear(iGroupList)
	}

	Vec_IntFree(iGroupList)
}

func oSplitByDep(pNtk *Abc_Ntk_t, oDep []*Vec_Int_t, oMatch []*Vec_Int_t, oGroup []int32, oLastItem *int32, iGroup []int32) int32 {
	var i, j, k int32
	var numOfItemsAdded int32
	var array, sortedArray *Vec_Int_t

	numOfItemsAdded = 0

	for i = 0; i < *oLastItem; i++ {
		if Vec_IntSize(oMatch[i]) == 1 {
			continue
		}

		array = Vec_IntAlloc(Vec_IntSize(oMatch[i]))
		sortedArray = Vec_IntAlloc(Vec_IntSize(oMatch[i]))

		for j = 0; j < Vec_IntSize(oMatch[i]); j++ {
			var factor, encode int32

			encode = 0
			factor = 1

			for k = 0; k < Vec_IntSize(oDep[Vec_IntEntry(oMatch[i], j)]); k++ {
				encode += iGroup[Vec_IntEntry(oDep[Vec_IntEntry(oMatch[i], j)], k)] * factor
			}

			Vec_IntPush(array, encode)
			Vec_IntPushUniqueOrder(sortedArray, encode)

			if encode < 0 {
				fmt.Printf("WARNING! Integer overflow!")
			}
		}

		for Vec_IntSize(sortedArray) > 1 {
			for k = 0; k < Vec_IntSize(oMatch[i]); k++ {
				if Vec_IntEntry(array, k) == Vec_IntEntryLast(sortedArray) {
					Vec_IntPush(oMatch[*oLastItem+numOfItemsAdded], Vec_IntEntry(oMatch[i], k))
					oGroup[Vec_IntEntry(oMatch[i], k)] = *oLastItem + numOfItemsAdded
					vec.Vec_IntRemove(oMatch[i], Vec_IntEntry(oMatch[i], k))
					vec.Vec_IntRemove(array, Vec_IntEntry(array, k))
					k--
				}
			}
			numOfItemsAdded++
			Vec_IntPop(sortedArray)
		}

		Vec_IntFree(array)
		Vec_IntFree(sortedArray)
	}

	*oLastItem += numOfItemsAdded

	return numOfItemsAdded
}

func iSplitByDep(pNtk *Abc_Ntk_t, iDep []*Vec_Int_t, iMatch []*Vec_Int_t, iGroup []int32, iLastItem *int32, oGroup []int32) int32 {
	var i, j, k int32
	var numOfItemsAdded int32 = 0
	var array, sortedArray *Vec_Int_t

	for i = 0; i < *iLastItem; i++ {
		if Vec_IntSize(iMatch[i]) == 1 {
			continue
		}

		array = Vec_IntAlloc(Vec_IntSize(iMatch[i]))
		sortedArray = Vec_IntAlloc(Vec_IntSize(iMatch[i]))

		for j = 0; j < Vec_IntSize(iMatch[i]); j++ {
			var factor, encode int32

			encode = 0
			factor = 1

			for k = 0; k < Vec_IntSize(iDep[Vec_IntEntry(iMatch[i], j)]); k++ {
				encode += oGroup[Vec_IntEntry(iDep[Vec_IntEntry(iMatch[i], j)], k)] * factor
			}

			Vec_IntPush(array, encode)
			Vec_IntPushUniqueOrder(sortedArray, encode)
		}

		for Vec_IntSize(sortedArray) > 1 {
			for k = 0; k < Vec_IntSize(iMatch[i]); k++ {
				if Vec_IntEntry(array, k) == Vec_IntEntryLast(sortedArray) {
					Vec_IntPush(iMatch[*iLastItem+numOfItemsAdded], Vec_IntEntry(iMatch[i], k))
					iGroup[Vec_IntEntry(iMatch[i], k)] = *iLastItem + numOfItemsAdded
					vec.Vec_IntRemove(iMatch[i], Vec_IntEntry(iMatch[i], k))
					vec.Vec_IntRemove(array, Vec_IntEntry(array, k))
					k--
				}
			}
			numOfItemsAdded++
			Vec_IntPop(sortedArray)
		}

		Vec_IntFree(array)
		Vec_IntFree(sortedArray)
	}

	*iLastItem += numOfItemsAdded

	return numOfItemsAdded
}

func findTopologicalOrder(pNtk *Abc_Ntk_t) []*Vec_Ptr_t {
	var vNodes []*Vec_Ptr_t
	var pObj, pFanout *Abc_Obj_t
	var i, k int32

	// start the array of nodes
	vNodes = make([]*Vec_Ptr_t, Abc_NtkPiNum(pNtk))
	for i = 0; i < Abc_NtkPiNum(pNtk); i++ {
		vNodes[i] = Vec_PtrAlloc(50)
	}

	for i = 0; i < Abc_NtkCiNum(pNtk); i++ {
		pObj = baseabc.Abc_NtkCi(pNtk, i)
		// set the traversal ID
		baseabc.Abc_NtkIncrementTravId(pNtk)
		baseabc.Abc_NodeSetTravIdCurrent(pObj)
		if baseabc.Abc_NtkIsNetlist(pNtk) != 0 { // Abc_ObjFanout0Ntk
			pObj = baseabc.Abc_ObjFanout0(pObj)
		}
		var ordered []*Abc_Obj_t
		for k = 0; k < baseabc.Abc_ObjFanoutNum(pObj); k++ {
			pFanout = baseabc.Abc_ObjFanout(pObj, k)
			baseabc.Abc_NtkDfsReverse_rec(pFanout, &ordered)
		}
		for _, pNode := range ordered {
			Vec_PtrPush(vNodes[i], pNode)
		}
	}

	return vNodes
}

func Abc_NtkSimulateOneNode(pNtk *Abc_Ntk_t, pModel []int32, input int32, topOrder []*Vec_Ptr_t) []int32 {
	var pNode *Abc_Obj_t
	var vNodes []*Abc_Obj_t
	var Value0, Value1, i int32

	// increment the trav ID
	baseabc.Abc_NtkIncrementTravId(pNtk)
	// set the CI values
	bmCopyWrite(baseabc.Abc_AigConst1(pNtk).Object, 1)
	pNode = baseabc.Abc_NtkCi(pNtk, input)
	bmTempWrite(pNode, pModel[input])

	// simulate in the topological order
	for i = Vec_PtrSize(topOrder[input]) - 1; i >= 0; i-- {
		pNode = Vec_PtrEntry(topOrder[input], i).(*Abc_Obj_t)

		Value0 = int32(bmCopyRead(Abc_ObjFanin0(pNode))) ^ baseabc.Abc_ObjFaninC0(pNode)
		Value1 = int32(bmCopyRead(Abc_ObjFanin1(pNode))) ^ baseabc.Abc_ObjFaninC1(pNode)

		if int32(bmCopyRead(pNode)) != (Value0 & Value1) {
			bmTempWrite(pNode, Value0&Value1)
			vNodes = append(vNodes, pNode)
		}
	}
	// fill the output values
	pValues := make([]int32, Abc_NtkCoNum(pNtk))
	for i = 0; i < Abc_NtkCoNum(pNtk); i++ {
		pNode = baseabc.Abc_NtkCo(pNtk, i)
		pValues[i] = int32(bmCopyRead(Abc_ObjFanin0(pNode))) ^ baseabc.Abc_ObjFaninC0(pNode)
	}

	pNode = baseabc.Abc_NtkCi(pNtk, input)
	if bmCopyRead(pNode) == 1 {
		bmCopyWrite(pNode, 0)
	} else {
		bmCopyWrite(pNode, 1)
	}

	for _, pNode = range vNodes {
		if bmCopyRead(pNode) == 1 {
			bmCopyWrite(pNode, 0)
		} else {
			bmCopyWrite(pNode, 1)
		}
	}

	return pValues
}

func refineIOBySimulation(pNtk *Abc_Ntk_t, iMatch []*Vec_Int_t, iLastItem *int32, iGroup []int32, iDep []*Vec_Int_t, oMatch []*Vec_Int_t, oLastItem *int32, oGroup []int32, oDep []*Vec_Int_t, vPiValues []byte, observability []int32, topOrder []*Vec_Ptr_t) int32 {
	var pModel []int32
	var output, output2 []int32
	var lastItem int32
	var i, j, k int32
	var iComputedNum, iComputedNumSorted *Vec_Int_t
	var oComputedNum *Vec_Int_t // encoding the number of flips
	var factor int32
	var isRefined int32 = FALSE

	pModel = make([]int32, Abc_NtkCiNum(pNtk))

	for i = 0; i < Abc_NtkPiNum(pNtk); i++ {
		pModel[i] = int32(vPiValues[i]) - '0'
	}
	// (no latches: bm refuses sequential networks)

	output = bmNtkVerifySimulatePattern(pNtk, pModel)

	oComputedNum = Vec_IntAlloc(Abc_NtkPoNum(pNtk))
	for i = 0; i < Abc_NtkPoNum(pNtk); i++ {
		Vec_IntPush(oComputedNum, 0)
	}

	/****************************************************************************************/
	/********** group outputs that produce 1 and outputs that produce 0 together ************/

	lastItem = *oLastItem
	for i = 0; i < lastItem && (*oLastItem) != Abc_NtkPoNum(pNtk); i++ {
		flag := FALSE

		if Vec_IntSize(oMatch[i]) == 1 {
			continue
		}

		for j = 1; j < Vec_IntSize(oMatch[i]); j++ {
			if output[Vec_IntEntry(oMatch[i], 0)] != output[Vec_IntEntry(oMatch[i], j)] {
				flag = TRUE
				break
			}
		}

		if flag != 0 {
			for j = 0; j < Vec_IntSize(oMatch[i]); j++ {
				if output[Vec_IntEntry(oMatch[i], j)] != 0 {
					Vec_IntPush(oMatch[*oLastItem], Vec_IntEntry(oMatch[i], j))
					oGroup[Vec_IntEntry(oMatch[i], j)] = *oLastItem
					vec.Vec_IntRemove(oMatch[i], Vec_IntEntry(oMatch[i], j))
					j--
				}
			}

			(*oLastItem)++
		}
	}

	if (*oLastItem) > lastItem {
		isRefined = TRUE
		iSortDependencies(pNtk, iDep, oGroup)
	}

	/****************************************************************************************/
	/************* group inputs that make the same number of flips in outpus ****************/

	lastItem = *iLastItem
	for i = 0; i < lastItem && (*iLastItem) != Abc_NtkPiNum(pNtk); i++ {
		var num int32

		if Vec_IntSize(iMatch[i]) == 1 {
			continue
		}

		iComputedNum = Vec_IntAlloc(Vec_IntSize(iMatch[i]))
		iComputedNumSorted = Vec_IntAlloc(Vec_IntSize(iMatch[i]))

		for j = 0; j < Vec_IntSize(iMatch[i]); j++ {
			if vPiValues[Vec_IntEntry(iMatch[i], j)] == '0' {
				pModel[Vec_IntEntry(iMatch[i], j)] = 1
			} else {
				pModel[Vec_IntEntry(iMatch[i], j)] = 0
			}

			output2 = Abc_NtkSimulateOneNode(pNtk, pModel, Vec_IntEntry(iMatch[i], j), topOrder)

			num = 0
			factor = 1
			for k = 0; k < Vec_IntSize(iDep[Vec_IntEntry(iMatch[i], j)]); k++ {
				outputIndex := Vec_IntEntry(iDep[Vec_IntEntry(iMatch[i], j)], k)

				if output2[outputIndex] != 0 {
					num += (oGroup[outputIndex] + 1) * factor
				}

				if output[outputIndex] != output2[outputIndex] {
					temp := Vec_IntEntry(oComputedNum, outputIndex) + i + 1
					Vec_IntWriteEntry(oComputedNum, outputIndex, temp)
					observability[Vec_IntEntry(iMatch[i], j)]++
				}
			}

			Vec_IntPush(iComputedNum, num)
			Vec_IntPushUniqueOrder(iComputedNumSorted, num)

			pModel[Vec_IntEntry(iMatch[i], j)] = int32(vPiValues[Vec_IntEntry(iMatch[i], j)]) - '0'
		}

		for Vec_IntSize(iComputedNumSorted) > 1 {
			for k = 0; k < Vec_IntSize(iMatch[i]); k++ {
				if Vec_IntEntry(iComputedNum, k) == Vec_IntEntryLast(iComputedNumSorted) {
					Vec_IntPush(iMatch[*iLastItem], Vec_IntEntry(iMatch[i], k))
					iGroup[Vec_IntEntry(iMatch[i], k)] = *iLastItem
					vec.Vec_IntRemove(iMatch[i], Vec_IntEntry(iMatch[i], k))
					vec.Vec_IntRemove(iComputedNum, Vec_IntEntry(iComputedNum, k))
					k--
				}
			}
			(*iLastItem)++
			Vec_IntPop(iComputedNumSorted)
		}

		Vec_IntFree(iComputedNum)
		Vec_IntFree(iComputedNumSorted)
	}

	if (*iLastItem) > lastItem {
		isRefined = TRUE
		oSortDependencies(pNtk, oDep, iGroup)
	}

	/****************************************************************************************/
	/********** encode the number of flips in each output by flipping the outputs ***********/
	/********** and group all the outputs that have the same encoding         ***********/

	lastItem = *oLastItem
	for i = 0; i < lastItem && (*oLastItem) != Abc_NtkPoNum(pNtk); i++ {
		var encode, sortedEncode *Vec_Int_t // encoding the number of flips

		if Vec_IntSize(oMatch[i]) == 1 {
			continue
		}

		encode = Vec_IntAlloc(Vec_IntSize(oMatch[i]))
		sortedEncode = Vec_IntAlloc(Vec_IntSize(oMatch[i]))

		for j = 0; j < Vec_IntSize(oMatch[i]); j++ {
			Vec_IntPush(encode, Vec_IntEntry(oComputedNum, Vec_IntEntry(oMatch[i], j)))
			Vec_IntPushUniqueOrder(sortedEncode, Vec_IntEntry(encode, j))
		}

		for Vec_IntSize(sortedEncode) > 1 {
			for j = 0; j < Vec_IntSize(oMatch[i]); j++ {
				if Vec_IntEntry(encode, j) == Vec_IntEntryLast(sortedEncode) {
					Vec_IntPush(oMatch[*oLastItem], Vec_IntEntry(oMatch[i], j))
					oGroup[Vec_IntEntry(oMatch[i], j)] = *oLastItem
					vec.Vec_IntRemove(oMatch[i], Vec_IntEntry(oMatch[i], j))
					vec.Vec_IntRemove(encode, Vec_IntEntry(encode, j))
					j--
				}
			}

			(*oLastItem)++
			Vec_IntPop(sortedEncode)
		}

		Vec_IntFree(encode)
		Vec_IntFree(sortedEncode)
	}

	if (*oLastItem) > lastItem {
		isRefined = TRUE
	}

	Vec_IntFree(oComputedNum)

	return isRefined
}

func Abc_NtkMiterBm(pNtk1, pNtk2 *Abc_Ntk_t, iCurrMatch, oCurrMatch *Vec_Ptr_t) *Abc_Ntk_t {
	var Buffer string
	var pNtkMiter *Abc_Ntk_t

	pNtkMiter = baseabc.Abc_NtkAlloc(baseabc.ABC_NTK_STRASH, baseabc.ABC_FUNC_AIG, 1)
	Buffer = fmt.Sprintf("%s_%s_miter", Abc_NtkName(pNtk1), Abc_NtkName(pNtk2))
	Abc_NtkSetName(pNtkMiter, Extra_UtilStrsav(Buffer))
	pMan := pNtkMiter.FunctionManager.(*baseabc.Abc_Aig_t)

	//Abc_NtkMiterPrepare( pNtk1, pNtk2, pNtkMiter, fComb, nPartSize );
	{
		var pObj, pObjNew *Abc_Obj_t
		var i int32

		Abc_ObjSetCopyEdge(baseabc.Abc_AigConst1(pNtk1).Object, baseabc.Abc_AigConst1(pNtkMiter))
		Abc_ObjSetCopyEdge(baseabc.Abc_AigConst1(pNtk2).Object, baseabc.Abc_AigConst1(pNtkMiter))

		// create new PIs and remember them in the old PIs
		if iCurrMatch == nil {
			for i = 0; i < Abc_NtkCiNum(pNtk1); i++ {
				pObj = baseabc.Abc_NtkCi(pNtk1, i)
				pObjNew = baseabc.Abc_NtkCreatePi(pNtkMiter)
				// remember this PI in the old PIs
				Abc_ObjSetCopyEdge(pObj, pObjNew.Signal())
				pObj = baseabc.Abc_NtkCi(pNtk2, i)
				Abc_ObjSetCopyEdge(pObj, pObjNew.Signal())
				// add name
				baseabc.Abc_ObjAssignName(pObjNew, Abc_ObjName(pObj), "")
			}
		} else {
			for i = 0; i < Vec_PtrSize(iCurrMatch); i += 2 {
				pObjNew = baseabc.Abc_NtkCreatePi(pNtkMiter)
				pObj = Vec_PtrEntry(iCurrMatch, i).(*Abc_Obj_t)
				Abc_ObjSetCopyEdge(pObj, pObjNew.Signal())
				pObj = Vec_PtrEntry(iCurrMatch, i+1).(*Abc_Obj_t)
				Abc_ObjSetCopyEdge(pObj, pObjNew.Signal())
				// add name
				baseabc.Abc_ObjAssignName(pObjNew, Abc_ObjName(pObj), "")
			}
		}

		// create the only PO
		pObjNew = baseabc.Abc_NtkCreatePo(pNtkMiter)
		// add the PO name
		baseabc.Abc_ObjAssignName(pObjNew, "miter", "")
	}

	// Abc_NtkMiterAddOne( pNtk1, pNtkMiter );
	// Abc_NtkMiterAddOne( pNtk2, pNtkMiter );
	for _, pNtk := range []*Abc_Ntk_t{pNtk1, pNtk2} {
		var pNode *Abc_Obj_t
		var i int32
		for i = 0; i < Abc_NtkObjNumMax(pNtk); i++ {
			pNode = Abc_NtkObj(pNtk, i)
			if pNode == nil || baseabc.Abc_AigNodeIsAnd(pNode) == 0 {
				continue
			}
			Abc_ObjSetCopyEdge(pNode, baseabc.Abc_AigAndE(pMan, Abc_ObjChild0Copy(pNode), Abc_ObjChild1Copy(pNode)))
		}
	}

	// Abc_NtkMiterFinalize( pNtk1, pNtk2, pNtkMiter, fComb, nPartSize );
	{
		var vPairs []baseabc.Edge
		var pMiter baseabc.Edge
		var i int32

		vPairs = make([]baseabc.Edge, 0, 100)

		// collect the CO nodes for the miter
		if oCurrMatch != nil {
			for i = 0; i < Vec_PtrSize(oCurrMatch); i += 2 {
				vPairs = append(vPairs, Abc_ObjChild0Copy(Vec_PtrEntry(oCurrMatch, i).(*Abc_Obj_t)))
				vPairs = append(vPairs, Abc_ObjChild0Copy(Vec_PtrEntry(oCurrMatch, i+1).(*Abc_Obj_t)))
			}
		} else {
			var pNode *Abc_Obj_t

			for i = 0; i < Abc_NtkCoNum(pNtk1); i++ {
				pNode = baseabc.Abc_NtkCo(pNtk1, i)
				vPairs = append(vPairs, Abc_ObjChild0Copy(pNode))
				pNode = baseabc.Abc_NtkCo(pNtk2, i)
				vPairs = append(vPairs, Abc_ObjChild0Copy(pNode))
			}
		}

		pMiter = baseabc.Abc_AigMiter(pMan, &vPairs, 0)
		Abc_ObjAddFanin(Abc_NtkPo(pNtkMiter, 0), pMiter)
	}

	//Abc_AigCleanup(pNtkMiter->pManFunc);

	return pNtkMiter
}

// bmNtkVerifyReportError is abcBm.c's Abc_NtkVerifyReportError: it
// appends to mismatch the (PI index, value) pairs of the model restricted
// to the support of the first output where the networks differ.
func bmNtkVerifyReportError(pNtk1, pNtk2 *Abc_Ntk_t, pModel []int32, mismatch *Vec_Int_t) {
	var vNodes []*Abc_Obj_t
	var pNode *Abc_Obj_t
	var nErrors, nPrinted, i int32
	var iNode int32 = -1

	// get the CO values under this model
	pValues1__ := bmNtkVerifySimulatePattern(pNtk1, pModel)
	pValues2__ := bmNtkVerifySimulatePattern(pNtk2, pModel)
	// count the mismatches
	nErrors = 0
	for i = 0; i < Abc_NtkCoNum(pNtk1); i++ {
		if pValues1__[i] != pValues2__[i] {
			nErrors++
		}
	}
	_ = nErrors
	// print the first 3 outputs
	nPrinted = 0
	for i = 0; i < Abc_NtkCoNum(pNtk1); i++ {
		if pValues1__[i] != pValues2__[i] {
			if iNode == -1 {
				iNode = i
			}
			nPrinted++
			if nPrinted == 3 {
				break
			}
		}
	}
	// report mismatch for the first output
	if iNode >= 0 {
		// collect PIs in the cone
		pNode = baseabc.Abc_NtkCo(pNtk1, iNode)
		vNodes = baseabc.Abc_NtkNodeSupport(pNtk1, []*Abc_Obj_t{pNode}, 1)
		// set the PI numbers (pNode->iTemp = i; read back as (int)pNode->pCopy)
		ciIndex := make([]int32, Abc_NtkObjNumMax(pNtk1))
		for i = 0; i < Abc_NtkCiNum(pNtk1); i++ {
			ciIndex[baseabc.Abc_NtkCi(pNtk1, i).ID] = i
		}
		// print the model
		// C reads vNodes[0] even when the support is empty (an output
		// that is constant in pNtk1); no CI is reported then.
		if len(vNodes) > 0 && baseabc.Abc_ObjIsCi(vNodes[0]) != 0 {
			for _, pNode = range vNodes {
				Vec_IntPush(mismatch, baseabc.Abc_ObjId(pNode)-1)
				Vec_IntPush(mismatch, pModel[ciIndex[pNode.ID]])
			}
		}
	}
}

func Abc_NtkMiterSatBm(pNtk *Abc_Ntk_t, nConfLimit, nInsLimit ABC_INT64_T, fVerbose int32, pNumConfs, pNumInspects *ABC_INT64_T) int32 {
	var RetValue int32 = 0

	if pNumConfs != nil {
		*pNumConfs = 0
	}
	if pNumInspects != nil {
		*pNumInspects = 0
	}

	// load clauses into the sat_solver
	created := Abc_NtkMiterSatCreate(pNtk, 0)
	if created == nil {
		return 1
	}
	pSat := created.(*bsat.Solver)

	// simplify the problem
	status := pSat.SimplifyC()
	if status == 0 {
		pSat.Delete()
		return 1
	}

	// solve the miter
	if fVerbose != 0 {
		pSat.SetVerbosityC(1)
	}
	status = pSat.SolveC(nil, nConfLimit, nInsLimit, 0, 0)
	switch bsat.Status(status) {
	case bsat.Unknown:
		RetValue = -1
	case bsat.Satisfiable:
		RetValue = 0
	case bsat.Unsatisfiable:
		RetValue = 1
	default:
		panic("Abc_NtkMiterSatBm: unexpected solver status")
	}

	// if the problem is SAT, get the counterexample
	if bsat.Status(status) == bsat.Satisfiable {
		vCiIds := Abc_NtkGetCiSatVarNums(pNtk)
		nSize := int32(vec.Vec_IntSize(vCiIds))
		pNtk.Model = bsat.Sat_SolverGetModel(pSat, vec.Vec_IntArray(vCiIds), nSize)[:nSize]
		Vec_IntFree(vCiIds)
	}
	// free the sat_solver
	if fVerbose != 0 {
		bsat.Sat_SolverPrintStats(os.Stdout, pSat)
	}

	st := pSat.Statistics()
	if pNumConfs != nil {
		*pNumConfs = int64(int32(st.Conflicts))
	}
	if pNumInspects != nil {
		*pNumInspects = int64(int32(st.Inspections))
	}

	if pStore := pSat.StoreReleaseC(); pStore != nil { // sat_solver_store_free( pSat );
		bsat.Sto_ManFree(pStore)
	}
	pSat.Delete()
	return RetValue
}

func Abc_NtkBmSat(pNtk1, pNtk2 *Abc_Ntk_t, iMatchPairs, oMatchPairs *Vec_Ptr_t, mismatch *Vec_Int_t, mode int32) int32 {
	var pMiter *Abc_Ntk_t
	var pCnf *Abc_Ntk_t
	var RetValue int32 = 0

	// get the miter of the two networks
	if mode == 0 {
		pMiter = Abc_NtkMiterBm(pNtk1, pNtk2, iMatchPairs, oMatchPairs)
	}
	// modes 1 (add new outputs; it dereferences the NULL miter in C),
	// 2 (add some outputs) and 3 (remove all outputs) are never used.

	if pMiter == nil {
		fmt.Printf("Miter computation has failed.")
		return -1
	}
	RetValue = Abc_NtkMiterIsConstant(pMiter)
	if RetValue == 0 {
		// report the error
		if mismatch != nil {
			pMiter.Model = Abc_NtkVerifyGetCleanModel(pMiter, 1)
			bmNtkVerifyReportError(pNtk1, pNtk2, pMiter.Model, mismatch)
			pMiter.Model = nil
		}
		Abc_NtkDelete(pMiter)
		return RetValue
	}
	if RetValue == 1 {
		Abc_NtkDelete(pMiter)
		return RetValue
	}

	// convert the miter into a CNF
	pCnf = Abc_NtkMulti(pMiter, 0, 100, 1, 0, 0, 0)
	Abc_NtkDelete(pMiter)
	if pCnf == nil {
		fmt.Printf("Renoding for CNF has failed.")
		return -1
	}

	// solve the CNF using the SAT solver
	RetValue = Abc_NtkMiterSat(pCnf, 10000, 0, 0, nil, nil)
	if mismatch != nil && pCnf.Model != nil {
		bmNtkVerifyReportError(pNtk1, pNtk2, pCnf.Model, mismatch)
	}

	pCnf.Model = nil
	Abc_NtkDelete(pCnf)

	return RetValue
}

func checkEquivalence(pNtk1 *Abc_Ntk_t, matchedInputs1, matchedOutputs1 *Vec_Int_t, pNtk2 *Abc_Ntk_t, matchedInputs2, matchedOutputs2 *Vec_Int_t) int32 {
	var iMatchPairs, oMatchPairs *Vec_Ptr_t
	var i int32
	var result int32

	iMatchPairs = Vec_PtrAlloc(Abc_NtkPiNum(pNtk1) * 2)
	oMatchPairs = Vec_PtrAlloc(Abc_NtkPoNum(pNtk1) * 2)

	for i = 0; i < Abc_NtkPiNum(pNtk1); i++ {
		Vec_PtrPush(iMatchPairs, Abc_NtkPi(pNtk2, Vec_IntEntry(matchedInputs2, i)))
		Vec_PtrPush(iMatchPairs, Abc_NtkPi(pNtk1, Vec_IntEntry(matchedInputs1, i)))
	}

	for i = 0; i < Abc_NtkPoNum(pNtk1); i++ {
		Vec_PtrPush(oMatchPairs, Abc_NtkPo(pNtk2, Vec_IntEntry(matchedOutputs2, i)))
		Vec_PtrPush(oMatchPairs, Abc_NtkPo(pNtk1, Vec_IntEntry(matchedOutputs1, i)))
	}

	result = Abc_NtkBmSat(pNtk1, pNtk2, iMatchPairs, oMatchPairs, nil, 0)

	if result != 0 {
		fmt.Printf("*** Circuits are equivalent ***\n")
	} else {
		fmt.Printf("*** Circuits are NOT equivalent ***\n")
	}

	Vec_PtrFree(iMatchPairs)
	Vec_PtrFree(oMatchPairs)

	return result
}

// bmCofPtr is computeCofactor's pCopy: NULL (constant 0), (void*)1
// (constant 1), or a node of the cofactor network.
type bmCofPtr struct {
	kind uint8 // bmCofNull, bmCofOne or bmCofNode
	e    baseabc.Edge
}

const (
	bmCofNull = iota
	bmCofOne
	bmCofNode
)

// notCond is Abc_ObjNotCond on a tagged pointer: complementing toggles
// the pointer's low bit, so NULL and (void*)1 swap.
func (p bmCofPtr) notCond(c int32) bmCofPtr {
	if c == 0 {
		return p
	}
	switch p.kind {
	case bmCofNull:
		return bmCofPtr{kind: bmCofOne}
	case bmCofOne:
		return bmCofPtr{kind: bmCofNull}
	}
	return bmCofPtr{kind: bmCofNode, e: p.e.Not()}
}

func computeCofactor(pNtk *Abc_Ntk_t, nodesInLevel []*Vec_Ptr_t, bitVector []int32, currInputs *Vec_Int_t) *Abc_Ntk_t {
	var subNtk *Abc_Ntk_t
	var pObj, pObjNew *Abc_Obj_t
	var i, j, numOfLevels int32

	numOfLevels = Abc_AigLevel(pNtk) // number of levels excludes PI/POs

	// start a new network
	subNtk = baseabc.Abc_NtkAlloc(baseabc.ABC_NTK_STRASH, baseabc.ABC_FUNC_AIG, 1)
	Abc_NtkSetName(subNtk, Extra_UtilStrsav("subNtk"))
	pMan := subNtk.FunctionManager.(*baseabc.Abc_Aig_t)

	// pCopy of pNtk's objects, by ID
	copies := make([]bmCofPtr, Abc_NtkObjNumMax(pNtk))
	child0Copy := func(o *Abc_Obj_t) bmCofPtr {
		return copies[Abc_ObjFanin0(o).ID].notCond(baseabc.Abc_ObjFaninC0(o))
	}
	child1Copy := func(o *Abc_Obj_t) bmCofPtr {
		return copies[Abc_ObjFanin1(o).ID].notCond(baseabc.Abc_ObjFaninC1(o))
	}

	// Abc_AigConst1(pNtk)->pCopy = Abc_AigConst1(subNtk), which the
	// following Abc_NtkCleanCopy clears again: the constant reads as NULL.
	copies[baseabc.Abc_AigConst1(pNtk).Object.ID] = bmCofPtr{kind: bmCofNode, e: baseabc.Abc_AigConst1(subNtk)}

	// clean the node copy fields and mark the nodes that need to be copied to the new network
	baseabc.Abc_NtkCleanCopy(pNtk)
	for k := range copies {
		copies[k] = bmCofPtr{}
	}

	if bitVector != nil {
		for i = 0; i < Abc_NtkPiNum(pNtk); i++ {
			if bitVector[i] != 0 {
				pObj = Abc_NtkPi(pNtk, i)
				copies[pObj.ID] = bmCofPtr{kind: bmCofOne}
			}
		}
	}

	for i = 0; i < Vec_IntSize(currInputs); i++ {
		pObj = Abc_NtkPi(pNtk, Vec_IntEntry(currInputs, i))
		pObjNew = baseabc.Abc_NtkDupObj(subNtk, pObj, 1)
		copies[pObj.ID] = bmCofPtr{kind: bmCofNode, e: pObjNew.Signal()}
	}

	// i = 0 are the inputs and the inputs are not added to the 2d array ( nodesInLevel )
	for i = 0; i <= numOfLevels; i++ {
		for j = 0; j < Vec_PtrSize(nodesInLevel[i]); j++ {
			pObj = Vec_PtrEntry(nodesInLevel[i], j).(*Abc_Obj_t)
			c0, c1 := child0Copy(pObj), child1Copy(pObj)

			switch {
			case c0.kind == bmCofNull && c1.kind == bmCofNull:
				copies[pObj.ID] = bmCofPtr{}
			case c0.kind == bmCofNull && c1.kind == bmCofOne:
				copies[pObj.ID] = bmCofPtr{}
			case c0.kind == bmCofNull && c1.kind == bmCofNode:
				copies[pObj.ID] = bmCofPtr{}
			case c0.kind == bmCofOne && c1.kind == bmCofNull:
				copies[pObj.ID] = bmCofPtr{}
			case c0.kind == bmCofOne && c1.kind == bmCofOne:
				copies[pObj.ID] = bmCofPtr{kind: bmCofOne}
			case c0.kind == bmCofOne && c1.kind == bmCofNode:
				copies[pObj.ID] = c1
			case c0.kind == bmCofNode && c1.kind == bmCofNull:
				copies[pObj.ID] = bmCofPtr{}
			case c0.kind == bmCofNode && c1.kind == bmCofOne:
				copies[pObj.ID] = c0
			case c0.kind == bmCofNode && c1.kind == bmCofNode:
				copies[pObj.ID] = bmCofPtr{kind: bmCofNode, e: baseabc.Abc_AigAndE(pMan, c0.e, c1.e)}
			}
		}
	}

	for i = 0; i < Abc_NtkPoNum(pNtk); i++ {
		pObj = Abc_NtkPo(pNtk, i)
		pObjNew = baseabc.Abc_NtkDupObj(subNtk, pObj, 1)

		c0 := child0Copy(pObj)
		if c0.kind == bmCofNull {
			Abc_ObjAddFanin(pObjNew, baseabc.Abc_AigConst1(subNtk).Not()) // fCompl0 = 1
		} else if c0.kind == bmCofOne {
			Abc_ObjAddFanin(pObjNew, baseabc.Abc_AigConst1(subNtk)) // fCompl0 = 0
		} else {
			Abc_ObjAddFanin(pObjNew, c0.e)
		}
	}

	return subNtk
}

// static int MATCH_FOUND of matchNonSingletonOutputs and of match1by1:
// recursive calls of either function overwrite them, and the callers
// read them back after those calls.
var (
	bmMatchFoundOutputs int32
	bmCounterOutputs    int32
	bmMatchFound1by1    int32 = FALSE
	bmCounter1by1       int32
)

func matchNonSingletonOutputs(pNtk1 *Abc_Ntk_t, nodesInLevel1 []*Vec_Ptr_t, iMatch1, iDep1 []*Vec_Int_t, matchedInputs1 *Vec_Int_t, iGroup1 []int32, oMatch1 []*Vec_Int_t, oGroup1 []int32,
	pNtk2 *Abc_Ntk_t, nodesInLevel2 []*Vec_Ptr_t, iMatch2, iDep2 []*Vec_Int_t, matchedInputs2 *Vec_Int_t, iGroup2 []int32, oMatch2 []*Vec_Int_t, oGroup2 []int32,
	matchedOutputs1, matchedOutputs2, oMatchedGroups, iNonSingleton *Vec_Int_t,
	subNtk1, subNtk2 *Abc_Ntk_t, oMatchPairs *Vec_Ptr_t,
	oNonSingleton *Vec_Int_t, oI, idx, ii, iidx int32) int32 {
	var i int32
	var j, temp int32
	var mismatch *Vec_Int_t
	var skipList []int32

	bmMatchFoundOutputs = FALSE

	if oI == Vec_IntSize(oNonSingleton) {
		if iNonSingleton != nil {
			if match1by1(pNtk1, nodesInLevel1, iMatch1, iDep1, matchedInputs1, iGroup1, oMatch1, oGroup1,
				pNtk2, nodesInLevel2, iMatch2, iDep2, matchedInputs2, iGroup2, oMatch2, oGroup2,
				matchedOutputs1, matchedOutputs2, oMatchedGroups, iNonSingleton, ii, iidx) != 0 {
				bmMatchFoundOutputs = TRUE
			}
		}

		if iNonSingleton == nil {
			bmMatchFoundOutputs = TRUE
		}

		return bmMatchFoundOutputs
	}

	i = Vec_IntEntry(oNonSingleton, oI)

	mismatch = Vec_IntAlloc(10)

	skipList = make([]int32, Vec_IntSize(oMatch1[i]))

	for j = 0; j < Vec_IntSize(oMatch1[i]); j++ {
		skipList[j] = FALSE
	}

	Vec_PtrPush(oMatchPairs, Abc_NtkPo(subNtk1, Vec_IntEntry(oMatch1[i], idx)))
	Vec_IntPush(matchedOutputs1, Vec_IntEntry(oMatch1[i], idx))

	for j = 0; j < Vec_IntSize(oMatch2[i]) && bmMatchFoundOutputs == FALSE; j++ {
		if Vec_IntEntry(oMatch2[i], j) == -1 || skipList[j] == TRUE {
			continue
		}

		Vec_PtrPush(oMatchPairs, Abc_NtkPo(subNtk2, Vec_IntEntry(oMatch2[i], j)))
		Vec_IntPush(matchedOutputs2, Vec_IntEntry(oMatch2[i], j))

		bmCounterOutputs++
		if Abc_NtkBmSat(subNtk1, subNtk2, nil, oMatchPairs, mismatch, 0) != 0 {
			temp = Vec_IntEntry(oMatch2[i], j)
			Vec_IntWriteEntry(oMatch2[i], j, -1)

			if idx != Vec_IntSize(oMatch1[i])-1 {
				// call the same function with idx+1
				matchNonSingletonOutputs(pNtk1, nodesInLevel1, iMatch1, iDep1, matchedInputs1, iGroup1, oMatch1, oGroup1,
					pNtk2, nodesInLevel2, iMatch2, iDep2, matchedInputs2, iGroup2, oMatch2, oGroup2,
					matchedOutputs1, matchedOutputs2, oMatchedGroups, iNonSingleton,
					subNtk1, subNtk2, oMatchPairs,
					oNonSingleton, oI, idx+1, ii, iidx)
			} else {
				// call the same function with idx = 0 and oI++
				matchNonSingletonOutputs(pNtk1, nodesInLevel1, iMatch1, iDep1, matchedInputs1, iGroup1, oMatch1, oGroup1,
					pNtk2, nodesInLevel2, iMatch2, iDep2, matchedInputs2, iGroup2, oMatch2, oGroup2,
					matchedOutputs1, matchedOutputs2, oMatchedGroups, iNonSingleton,
					subNtk1, subNtk2, oMatchPairs,
					oNonSingleton, oI+1, 0, ii, iidx)
			}

			Vec_IntWriteEntry(oMatch2[i], j, temp)
		} else {
			var output1, output2 []int32
			var k int32
			var pModel []int32
			var vPiValues []byte

			vPiValues = make([]byte, Abc_NtkPiNum(subNtk1)+1)
			vPiValues[Abc_NtkPiNum(subNtk1)] = 0

			for k = 0; k < Abc_NtkPiNum(subNtk1); k++ {
				vPiValues[k] = '0'
			}

			// C stores the mismatch value itself (0 or 1), not '0' or '1',
			// so those inputs simulate as 0-'0' or 1-'0'
			for k = 0; k < Vec_IntSize(mismatch); k += 2 {
				vPiValues[Vec_IntEntry(mismatch, k)] = byte(Vec_IntEntry(mismatch, k+1))
			}

			pModel = make([]int32, Abc_NtkCiNum(subNtk1))

			for k = 0; k < Abc_NtkPiNum(subNtk1); k++ {
				pModel[k] = int32(vPiValues[k]) - '0'
			}

			output1 = bmNtkVerifySimulatePattern(subNtk1, pModel)

			output2 = bmNtkVerifySimulatePattern(subNtk2, pModel)

			for k = 0; k < Vec_IntSize(oMatch1[i]); k++ {
				// C reads output2[-1] for an entry matched (-1) at an outer
				// level: the high half of the malloc size word, 0. That k
				// is skipped by the -1 test above anyway.
				var value2 int32
				if o2 := Vec_IntEntry(oMatch2[i], k); o2 >= 0 {
					value2 = output2[o2]
				}
				if output1[Vec_IntEntry(oMatch1[i], idx)] != value2 {
					skipList[k] = TRUE
				}
			}
		}

		if bmMatchFoundOutputs == FALSE {
			Vec_PtrPop(oMatchPairs)
			Vec_IntPop(matchedOutputs2)
		}
	}

	if bmMatchFoundOutputs == FALSE {
		Vec_PtrPop(oMatchPairs)
		Vec_IntPop(matchedOutputs1)
	}

	if bmMatchFoundOutputs != 0 && bmCounterOutputs != 0 {
		bmCounterOutputs = 0
	}

	Vec_IntFree(mismatch)

	return bmMatchFoundOutputs
}

func match1by1(pNtk1 *Abc_Ntk_t, nodesInLevel1 []*Vec_Ptr_t, iMatch1, iDep1 []*Vec_Int_t, matchedInputs1 *Vec_Int_t, iGroup1 []int32, oMatch1 []*Vec_Int_t, oGroup1 []int32,
	pNtk2 *Abc_Ntk_t, nodesInLevel2 []*Vec_Ptr_t, iMatch2, iDep2 []*Vec_Int_t, matchedInputs2 *Vec_Int_t, iGroup2 []int32, oMatch2 []*Vec_Int_t, oGroup2 []int32,
	matchedOutputs1, matchedOutputs2, oMatchedGroups, iNonSingleton *Vec_Int_t, ii, idx int32) int32 {
	var subNtk1, subNtk2 *Abc_Ntk_t
	var oNonSingleton *Vec_Int_t
	var oMatchPairs *Vec_Ptr_t
	var skipList []int32
	var j, m int32
	var i int32

	bmMatchFound1by1 = FALSE

	if ii == Vec_IntSize(iNonSingleton) {
		bmMatchFound1by1 = TRUE
		return TRUE
	}

	i = Vec_IntEntry(iNonSingleton, ii)

	if idx == Vec_IntSize(iMatch1[i]) {
		// call again with the next element in iNonSingleton
		return match1by1(pNtk1, nodesInLevel1, iMatch1, iDep1, matchedInputs1, iGroup1, oMatch1, oGroup1,
			pNtk2, nodesInLevel2, iMatch2, iDep2, matchedInputs2, iGroup2, oMatch2, oGroup2,
			matchedOutputs1, matchedOutputs2, oMatchedGroups, iNonSingleton, ii+1, 0)
	}

	oNonSingleton = Vec_IntAlloc(10)
	oMatchPairs = Vec_PtrAlloc(100)
	skipList = make([]int32, Vec_IntSize(iMatch1[i]))

	for j = 0; j < Vec_IntSize(iMatch1[i]); j++ {
		skipList[j] = FALSE
	}

	Vec_IntPush(matchedInputs1, Vec_IntEntry(iMatch1[i], idx))
	idx++

	if idx == 1 {
		for j = 0; j < Vec_IntSize(iDep1[Vec_IntEntryLast(iMatch1[i])]); j++ {
			if Vec_IntSize(oMatch1[oGroup1[Vec_IntEntry(iDep1[Vec_IntEntryLast(iMatch1[i])], j)]]) == 1 {
				continue
			}
			if Vec_IntFind(oMatchedGroups, oGroup1[Vec_IntEntry(iDep1[Vec_IntEntryLast(iMatch1[i])], j)]) != -1 {
				continue
			}

			Vec_IntPushUnique(oNonSingleton, oGroup1[Vec_IntEntry(iDep1[Vec_IntEntryLast(iMatch1[i])], j)])
			Vec_IntPushUnique(oMatchedGroups, oGroup1[Vec_IntEntry(iDep1[Vec_IntEntryLast(iMatch1[i])], j)])
		}
	}

	subNtk1 = computeCofactor(pNtk1, nodesInLevel1, nil, matchedInputs1)

	for j = idx - 1; j < Vec_IntSize(iMatch2[i]) && bmMatchFound1by1 == FALSE; j++ {
		var tempJ int32
		var mismatch *Vec_Int_t

		if skipList[j] != 0 {
			continue
		}

		mismatch = Vec_IntAlloc(10)

		Vec_IntPush(matchedInputs2, Vec_IntEntry(iMatch2[i], j))

		subNtk2 = computeCofactor(pNtk2, nodesInLevel2, nil, matchedInputs2)

		for m = 0; m < Vec_IntSize(matchedOutputs1); m++ {
			Vec_PtrPush(oMatchPairs, Abc_NtkPo(subNtk1, Vec_IntEntry(matchedOutputs1, m)))
			Vec_PtrPush(oMatchPairs, Abc_NtkPo(subNtk2, Vec_IntEntry(matchedOutputs2, m)))
		}

		bmCounter1by1++

		if Abc_NtkBmSat(subNtk2, subNtk1, nil, oMatchPairs, mismatch, 0) != 0 {
			if idx-1 != j {
				tempJ = Vec_IntEntry(iMatch2[i], idx-1)
				Vec_IntWriteEntry(iMatch2[i], idx-1, Vec_IntEntry(iMatch2[i], j))
				Vec_IntWriteEntry(iMatch2[i], j, tempJ)
			}

			// we look for a match for outputs in oNonSingleton
			matchNonSingletonOutputs(pNtk1, nodesInLevel1, iMatch1, iDep1, matchedInputs1, iGroup1, oMatch1, oGroup1,
				pNtk2, nodesInLevel2, iMatch2, iDep2, matchedInputs2, iGroup2, oMatch2, oGroup2,
				matchedOutputs1, matchedOutputs2, oMatchedGroups, iNonSingleton,
				subNtk1, subNtk2, oMatchPairs, oNonSingleton, 0, 0, ii, idx)

			if idx-1 != j {
				tempJ = Vec_IntEntry(iMatch2[i], idx-1)
				Vec_IntWriteEntry(iMatch2[i], idx-1, Vec_IntEntry(iMatch2[i], j))
				Vec_IntWriteEntry(iMatch2[i], j, tempJ)
			}
		} else {
			var FpNtk1, FpNtk2 *Abc_Ntk_t
			var bitVector1, bitVector2 []int32
			var currInputs1, currInputs2 *Vec_Int_t
			var vSupp []*Abc_Obj_t
			var pObj *Abc_Obj_t
			var suppNum1 int32 = 0
			var suppNum2 []int32

			bitVector1 = make([]int32, Abc_NtkPiNum(pNtk1))
			bitVector2 = make([]int32, Abc_NtkPiNum(pNtk2))

			currInputs1 = Vec_IntAlloc(10)
			currInputs2 = Vec_IntAlloc(10)

			suppNum2 = make([]int32, Vec_IntSize(iMatch2[i])-idx+1)

			for m = 0; m < Abc_NtkPiNum(pNtk1); m++ {
				bitVector1[m] = 0
				bitVector2[m] = 0
			}

			for m = 0; m < Vec_IntSize(iMatch2[i])-idx+1; m++ {
				suppNum2[m] = 0
			}

			// First of all set the value of the inputs that are already matched and are in mismatch
			for m = 0; m < Vec_IntSize(mismatch); m += 2 {
				n := Vec_IntEntry(mismatch, m)

				bitVector1[Vec_IntEntry(matchedInputs1, n)] = Vec_IntEntry(mismatch, m+1)
				bitVector2[Vec_IntEntry(matchedInputs2, n)] = Vec_IntEntry(mismatch, m+1)
			}

			for m = idx - 1; m < Vec_IntSize(iMatch1[i]); m++ {
				Vec_IntPush(currInputs1, Vec_IntEntry(iMatch1[i], m))
				Vec_IntPush(currInputs2, Vec_IntEntry(iMatch2[i], m))
			}

			// Then add all the inputs that are not yet matched to the currInputs
			for m = 0; m < Abc_NtkPiNum(pNtk1); m++ {
				if Vec_IntFind(matchedInputs1, m) == -1 {
					Vec_IntPushUnique(currInputs1, m)
				}

				if Vec_IntFind(matchedInputs2, m) == -1 {
					Vec_IntPushUnique(currInputs2, m)
				}
			}

			FpNtk1 = computeCofactor(pNtk1, nodesInLevel1, bitVector1, currInputs1)
			FpNtk2 = computeCofactor(pNtk2, nodesInLevel2, bitVector2, currInputs2)

			for m = 0; m < Abc_NtkPoNum(FpNtk1); m++ {
				pObj = Abc_NtkPo(FpNtk1, m)
				vSupp = baseabc.Abc_NtkNodeSupport(FpNtk1, []*Abc_Obj_t{pObj}, 1)

				for _, pSupp := range vSupp {
					if baseabc.Abc_ObjId(pSupp) == 1 {
						suppNum1 += int32(Vec_IntFind(matchedOutputs1, m)) + 1
					}
				}
			}

			for m = 0; m < Abc_NtkPoNum(FpNtk2); m++ {
				pObj = Abc_NtkPo(FpNtk2, m)
				vSupp = baseabc.Abc_NtkNodeSupport(FpNtk2, []*Abc_Obj_t{pObj}, 1)

				for _, pSupp := range vSupp {
					if baseabc.Abc_ObjId(pSupp)-1 < Vec_IntSize(iMatch2[i])-idx+1 &&
						baseabc.Abc_ObjId(pSupp)-1 >= 0 {
						suppNum2[baseabc.Abc_ObjId(pSupp)-1] += int32(Vec_IntFind(matchedOutputs2, m)) + 1
					}
				}
			}

			for m = 0; m < Vec_IntSize(iMatch2[i])-idx+1; m++ {
				if suppNum2[m] != suppNum1 {
					skipList[m+idx-1] = TRUE
				}
			}

			Abc_NtkDelete(FpNtk1)
			Abc_NtkDelete(FpNtk2)
			Vec_IntFree(currInputs1)
			Vec_IntFree(currInputs2)
		}

		Vec_PtrClear(oMatchPairs)
		Abc_NtkDelete(subNtk2)
		Vec_IntFree(mismatch)

		if bmMatchFound1by1 == FALSE {
			Vec_IntPop(matchedInputs2)
		}
	}

	if bmMatchFound1by1 == FALSE {
		Vec_IntPop(matchedInputs1)

		if idx == 1 {
			for m = 0; m < Vec_IntSize(oNonSingleton); m++ {
				Vec_IntPop(oMatchedGroups)
			}
		}
	}

	Vec_IntFree(oNonSingleton)
	Vec_PtrFree(oMatchPairs)
	Abc_NtkDelete(subNtk1)

	if bmMatchFound1by1 != 0 && bmCounter1by1 != 0 {
		bmCounter1by1 = 0
	}

	return bmMatchFound1by1
}

func refineBySAT(pNtk1 *Abc_Ntk_t, iMatch1 []*Vec_Int_t, iGroup1 []int32, iDep1 []*Vec_Int_t, iLastItem1 *int32, oMatch1 []*Vec_Int_t, oGroup1 []int32, oDep1 []*Vec_Int_t, oLastItem1 *int32, observability1 []int32,
	pNtk2 *Abc_Ntk_t, iMatch2 []*Vec_Int_t, iGroup2 []int32, iDep2 []*Vec_Int_t, iLastItem2 *int32, oMatch2 []*Vec_Int_t, oGroup2 []int32, oDep2 []*Vec_Int_t, oLastItem2 *int32, observability2 []int32) float32 {
	var i, j int32
	var pObj *Abc_Obj_t
	var iNonSingleton *Vec_Int_t
	var matchedInputs1, matchedInputs2 *Vec_Int_t
	var matchedOutputs1, matchedOutputs2 *Vec_Int_t
	var nodesInLevel1, nodesInLevel2 []*Vec_Ptr_t
	var oMatchedGroups *Vec_Int_t
	var matchFound int32
	clk := time.Now()
	var satTime float32 = 0.0

	iNonSingleton = Vec_IntAlloc(10)

	matchedInputs1 = Vec_IntAlloc(Abc_NtkPiNum(pNtk1))
	matchedInputs2 = Vec_IntAlloc(Abc_NtkPiNum(pNtk2))

	matchedOutputs1 = Vec_IntAlloc(Abc_NtkPoNum(pNtk1))
	matchedOutputs2 = Vec_IntAlloc(Abc_NtkPoNum(pNtk2))

	nodesInLevel1 = make([]*Vec_Ptr_t, Abc_AigLevel(pNtk1)+1) // why numOfLevels+1? because the inputs are in level 0
	for i = 0; i <= Abc_AigLevel(pNtk1); i++ {
		nodesInLevel1[i] = Vec_PtrAlloc(20)
	}

	// bucket sort the objects based on their levels
	for i = 0; i < Abc_NtkObjNumMax(pNtk1); i++ {
		pObj = Abc_NtkObj(pNtk1, i)
		if pObj == nil || baseabc.Abc_AigNodeIsAnd(pObj) == 0 {
			continue
		}
		Vec_PtrPush(nodesInLevel1[baseabc.Abc_ObjLevel(pObj)], pObj)
	}

	nodesInLevel2 = make([]*Vec_Ptr_t, Abc_AigLevel(pNtk2)+1) // why numOfLevels+1? because the inputs are in level 0
	for i = 0; i <= Abc_AigLevel(pNtk2); i++ {
		nodesInLevel2[i] = Vec_PtrAlloc(20)
	}

	// bucket sort the objects based on their levels
	for i = 0; i < Abc_NtkObjNumMax(pNtk2); i++ {
		pObj = Abc_NtkObj(pNtk2, i)
		if pObj == nil || baseabc.Abc_AigNodeIsAnd(pObj) == 0 {
			continue
		}
		Vec_PtrPush(nodesInLevel2[baseabc.Abc_ObjLevel(pObj)], pObj)
	}

	oMatchedGroups = Vec_IntAlloc(10)

	for i = 0; i < *iLastItem1; i++ {
		if Vec_IntSize(iMatch1[i]) == 1 {
			Vec_IntPush(matchedInputs1, Vec_IntEntryLast(iMatch1[i]))
			Vec_IntPush(matchedInputs2, Vec_IntEntryLast(iMatch2[i]))
		} else {
			Vec_IntPush(iNonSingleton, i)
		}
	}

	for i = 0; i < *oLastItem1; i++ {
		if Vec_IntSize(oMatch1[i]) == 1 {
			Vec_IntPush(matchedOutputs1, Vec_IntEntryLast(oMatch1[i]))
			Vec_IntPush(matchedOutputs2, Vec_IntEntryLast(oMatch2[i]))
		}
	}

	for i = 0; i < Vec_IntSize(iNonSingleton)-1; i++ {
		for j = i + 1; j < Vec_IntSize(iNonSingleton); j++ {
			if observability2[Vec_IntEntry(iMatch2[Vec_IntEntry(iNonSingleton, j)], 0)] >
				observability2[Vec_IntEntry(iMatch2[Vec_IntEntry(iNonSingleton, i)], 0)] {
				temp := Vec_IntEntry(iNonSingleton, i)
				Vec_IntWriteEntry(iNonSingleton, i, Vec_IntEntry(iNonSingleton, j))
				Vec_IntWriteEntry(iNonSingleton, j, temp)
			} else if observability2[Vec_IntEntry(iMatch2[Vec_IntEntry(iNonSingleton, j)], 0)] ==
				observability2[Vec_IntEntry(iMatch2[Vec_IntEntry(iNonSingleton, i)], 0)] {
				if Vec_IntSize(iMatch2[Vec_IntEntry(iNonSingleton, j)]) < Vec_IntSize(iMatch2[Vec_IntEntry(iNonSingleton, i)]) {
					temp := Vec_IntEntry(iNonSingleton, i)
					Vec_IntWriteEntry(iNonSingleton, i, Vec_IntEntry(iNonSingleton, j))
					Vec_IntWriteEntry(iNonSingleton, j, temp)
				}
			}
		}
	}

	matchFound = match1by1(pNtk1, nodesInLevel1, iMatch1, iDep1, matchedInputs1, iGroup1, oMatch1, oGroup1,
		pNtk2, nodesInLevel2, iMatch2, iDep2, matchedInputs2, iGroup2, oMatch2, oGroup2,
		matchedOutputs1, matchedOutputs2, oMatchedGroups, iNonSingleton, 0, 0)

	if matchFound != 0 && Vec_IntSize(matchedOutputs1) != Abc_NtkPoNum(pNtk1) {
		var oNonSingleton *Vec_Int_t
		var oMatchPairs *Vec_Ptr_t
		var subNtk1, subNtk2 *Abc_Ntk_t

		oNonSingleton = Vec_IntAlloc(10)

		oMatchPairs = Vec_PtrAlloc(Abc_NtkPoNum(pNtk1) * 2)

		for i = 0; i < *oLastItem1; i++ {
			if Vec_IntSize(oMatch1[i]) > 1 && Vec_IntFind(oMatchedGroups, i) == -1 {
				Vec_IntPush(oNonSingleton, i)
			}
		}

		subNtk1 = computeCofactor(pNtk1, nodesInLevel1, nil, matchedInputs1)
		subNtk2 = computeCofactor(pNtk2, nodesInLevel2, nil, matchedInputs2)

		// (C passes iGroup1 for iGroup2 here)
		matchFound = matchNonSingletonOutputs(pNtk1, nodesInLevel1, iMatch1, iDep1, matchedInputs1, iGroup1, oMatch1, oGroup1,
			pNtk2, nodesInLevel2, iMatch2, iDep2, matchedInputs2, iGroup1, oMatch2, oGroup2,
			matchedOutputs1, matchedOutputs2, oMatchedGroups, nil,
			subNtk1, subNtk2, oMatchPairs, oNonSingleton, 0, 0, 0, 0)

		Vec_IntFree(oNonSingleton)
		Vec_PtrFree(oMatchPairs)

		Abc_NtkDelete(subNtk1)
		Abc_NtkDelete(subNtk2)
	}

	satTime = float32(time.Since(clk).Seconds())

	if matchFound != 0 {
		checkEquivalence(pNtk1, matchedInputs1, matchedOutputs1, pNtk2, matchedInputs2, matchedOutputs2)

		var result []byte
		result = fmt.Appendf(result, "I/O = %d / %d \n\n", Abc_NtkPiNum(pNtk1), Abc_NtkPoNum(pNtk1))

		for i = 0; i < Vec_IntSize(matchedInputs1); i++ {
			result = fmt.Appendf(result, "{%s}\t{%s}\n", Abc_ObjName(Abc_NtkPi(pNtk1, Vec_IntEntry(matchedInputs1, i))), Abc_ObjName(Abc_NtkPi(pNtk2, Vec_IntEntry(matchedInputs2, i))))
		}

		result = fmt.Appendf(result, "\n-----------------------------------------\n")

		for i = 0; i < Vec_IntSize(matchedOutputs1); i++ {
			result = fmt.Appendf(result, "{%s}\t{%s}\n", Abc_ObjName(Abc_NtkPo(pNtk1, Vec_IntEntry(matchedOutputs1, i))), Abc_ObjName(Abc_NtkPo(pNtk2, Vec_IntEntry(matchedOutputs2, i))))
		}

		// result = fopen("IOmatch.txt", "w"); ... fclose( result );
		_ = os.WriteFile("IOmatch.txt", result, 0o644)
	}

	Vec_IntFree(matchedInputs1)
	Vec_IntFree(matchedInputs2)
	Vec_IntFree(matchedOutputs1)
	Vec_IntFree(matchedOutputs2)
	Vec_IntFree(iNonSingleton)
	Vec_IntFree(oMatchedGroups)

	return satTime
}

func checkListConsistency(iMatch1, oMatch1, iMatch2, oMatch2 []*Vec_Int_t, iLastItem1, oLastItem1, iLastItem2, oLastItem2 int32) int32 {
	if iLastItem1 != iLastItem2 || oLastItem1 != oLastItem2 {
		return FALSE
	}

	return TRUE
}

func bmGateWay(pNtk1, pNtk2 *Abc_Ntk_t, p_equivalence int32) {
	var iDep1, oDep1 []*Vec_Int_t
	var iDep2, oDep2 []*Vec_Int_t
	var iMatch1, oMatch1 []*Vec_Int_t
	var iMatch2, oMatch2 []*Vec_Int_t
	var iGroup1, oGroup1 []int32
	var iGroup2, oGroup2 []int32
	var iLastItem1, oLastItem1 int32
	var iLastItem2, oLastItem2 int32
	var i, j int32

	var vPiValues1, vPiValues2 []byte
	var observability1, observability2 []int32
	clk := time.Now()
	var initTime float32
	var simulTime float32
	var satTime float32
	var topOrder1, topOrder2 []*Vec_Ptr_t

	iDep1 = make([]*Vec_Int_t, Abc_NtkPiNum(pNtk1))
	oDep1 = make([]*Vec_Int_t, Abc_NtkPoNum(pNtk1))

	iDep2 = make([]*Vec_Int_t, Abc_NtkPiNum(pNtk2))
	oDep2 = make([]*Vec_Int_t, Abc_NtkPoNum(pNtk2))

	iMatch1 = make([]*Vec_Int_t, Abc_NtkPiNum(pNtk1))
	oMatch1 = make([]*Vec_Int_t, Abc_NtkPoNum(pNtk1))

	iMatch2 = make([]*Vec_Int_t, Abc_NtkPiNum(pNtk2))
	oMatch2 = make([]*Vec_Int_t, Abc_NtkPoNum(pNtk2))

	iGroup1 = make([]int32, Abc_NtkPiNum(pNtk1))
	oGroup1 = make([]int32, Abc_NtkPoNum(pNtk1))

	iGroup2 = make([]int32, Abc_NtkPiNum(pNtk2))
	oGroup2 = make([]int32, Abc_NtkPoNum(pNtk2))

	vPiValues1 = make([]byte, Abc_NtkPiNum(pNtk1)+1)
	vPiValues1[Abc_NtkPiNum(pNtk1)] = 0

	vPiValues2 = make([]byte, Abc_NtkPiNum(pNtk2)+1)
	vPiValues2[Abc_NtkPiNum(pNtk2)] = 0

	observability1 = make([]int32, Abc_NtkPiNum(pNtk1))
	observability2 = make([]int32, Abc_NtkPiNum(pNtk2))

	for i = 0; i < Abc_NtkPiNum(pNtk1); i++ {
		iDep1[i] = Vec_IntAlloc(1)
		iMatch1[i] = Vec_IntAlloc(1)

		iDep2[i] = Vec_IntAlloc(1)
		iMatch2[i] = Vec_IntAlloc(1)

		vPiValues1[i] = '0'
		vPiValues2[i] = '0'

		observability1[i] = 0
		observability2[i] = 0
	}

	for i = 0; i < Abc_NtkPoNum(pNtk1); i++ {
		oDep1[i] = Vec_IntAlloc(1)
		oMatch1[i] = Vec_IntAlloc(1)

		oDep2[i] = Vec_IntAlloc(1)
		oMatch2[i] = Vec_IntAlloc(1)
	}

	/************* Strashing ************/
	pNtk1 = Abc_NtkStrash(pNtk1, 0, 0, 0)
	pNtk2 = Abc_NtkStrash(pNtk2, 0, 0, 0)
	fmt.Printf("Network  strashing is done!\n")
	/************************************/

	/******* Getting Dependencies *******/
	getDependencies(pNtk1, iDep1, oDep1)
	getDependencies(pNtk2, iDep2, oDep2)
	fmt.Printf("Getting dependencies is done!\n")
	/************************************/

	/***** Intializing match lists ******/
	initMatchList(pNtk1, iDep1, oDep1, iMatch1, &iLastItem1, oMatch1, &oLastItem1, iGroup1, oGroup1, p_equivalence)
	initMatchList(pNtk2, iDep2, oDep2, iMatch2, &iLastItem2, oMatch2, &oLastItem2, iGroup2, oGroup2, p_equivalence)
	fmt.Printf("Initializing match lists is done!\n")
	/************************************/

	if checkListConsistency(iMatch1, oMatch1, iMatch2, oMatch2, iLastItem1, oLastItem1, iLastItem2, oLastItem2) == 0 {
		fmt.Fprintf(os.Stdout, "I/O dependencies of two circuits are different.\n")
		goto freeAndExit
	}

	fmt.Printf("Refining IOs by dependencies ...")
	// split match lists further by checking dependencies
	for {
		var iNumOfItemsAdded, oNumOfItemsAdded int32 = 1, 1

		for {
			if oNumOfItemsAdded != 0 {
				iSortDependencies(pNtk1, iDep1, oGroup1)
				iSortDependencies(pNtk2, iDep2, oGroup2)
			}

			if iNumOfItemsAdded != 0 {
				oSortDependencies(pNtk1, oDep1, iGroup1)
				oSortDependencies(pNtk2, oDep2, iGroup2)
			}

			if iLastItem1 < Abc_NtkPiNum(pNtk1) {
				iSplitByDep(pNtk1, iDep1, iMatch1, iGroup1, &iLastItem1, oGroup1)
				if oLastItem1 < Abc_NtkPoNum(pNtk1) {
					oSplitByDep(pNtk1, oDep1, oMatch1, oGroup1, &oLastItem1, iGroup1)
				}
			}

			if iLastItem2 < Abc_NtkPiNum(pNtk2) {
				iNumOfItemsAdded = iSplitByDep(pNtk2, iDep2, iMatch2, iGroup2, &iLastItem2, oGroup2)
			} else {
				iNumOfItemsAdded = 0
			}

			if oLastItem2 < Abc_NtkPoNum(pNtk2) {
				oNumOfItemsAdded = oSplitByDep(pNtk2, oDep2, oMatch2, oGroup2, &oLastItem2, iGroup2)
			} else {
				oNumOfItemsAdded = 0
			}

			if checkListConsistency(iMatch1, oMatch1, iMatch2, oMatch2, iLastItem1, oLastItem1, iLastItem2, oLastItem2) == 0 {
				fmt.Fprintf(os.Stdout, "I/O dependencies of two circuits are different.\n")
				goto freeAndExit
			}
			if !(iNumOfItemsAdded != 0 || oNumOfItemsAdded != 0) {
				break
			}
		}
		break
	}

	fmt.Printf(" done!\n")

	initTime = float32(time.Since(clk).Seconds())
	clk = time.Now()

	topOrder1 = findTopologicalOrder(pNtk1)
	topOrder2 = findTopologicalOrder(pNtk2)

	fmt.Printf("Refining IOs by simulation ...")

	for {
		var counter int32 = 0
		var ioSuccess1, ioSuccess2 int32

		for {
			for i = 0; i < iLastItem1; i++ {
				temp := int32(util.Sim_RandomUnsigned() % 2) // SIM_RANDOM_UNSIGNED

				if Vec_IntSize(iMatch1[i]) != Vec_IntSize(iMatch2[i]) {
					fmt.Fprintf(os.Stdout, "Input refinement by simulation finds two circuits different.\n")
					goto freeAndExit
				}

				for j = 0; j < Vec_IntSize(iMatch1[i]); j++ {
					vPiValues1[Vec_IntEntry(iMatch1[i], j)] = byte(temp + '0')
					vPiValues2[Vec_IntEntry(iMatch2[i], j)] = byte(temp + '0')
				}
			}

			ioSuccess1 = refineIOBySimulation(pNtk1, iMatch1, &iLastItem1, iGroup1, iDep1, oMatch1, &oLastItem1, oGroup1, oDep1, vPiValues1, observability1, topOrder1)
			ioSuccess2 = refineIOBySimulation(pNtk2, iMatch2, &iLastItem2, iGroup2, iDep2, oMatch2, &oLastItem2, oGroup2, oDep2, vPiValues2, observability2, topOrder2)

			if ioSuccess1 != 0 && ioSuccess2 != 0 {
				counter = 0
			} else {
				counter++
			}

			if ioSuccess1 != ioSuccess2 ||
				checkListConsistency(iMatch1, oMatch1, iMatch2, oMatch2, iLastItem1, oLastItem1, iLastItem2, oLastItem2) == 0 {
				fmt.Fprintf(os.Stdout, "Input refinement by simulation finds two circuits different.\n")
				goto freeAndExit
			}
			if !(counter <= 200) {
				break
			}
		}
		break
	}

	fmt.Printf(" done!\n")

	simulTime = float32(time.Since(clk).Seconds())
	fmt.Printf("SAT-based search started ...\n")

	satTime = refineBySAT(pNtk1, iMatch1, iGroup1, iDep1, &iLastItem1, oMatch1, oGroup1, oDep1, &oLastItem1, observability1,
		pNtk2, iMatch2, iGroup2, iDep2, &iLastItem2, oMatch2, oGroup2, oDep2, &oLastItem2, observability2)

	fmt.Printf("Init Time = %4.2f\n", initTime)
	fmt.Printf("Simulation Time = %4.2f\n", simulTime)
	fmt.Printf("SAT Time = %4.2f\n", satTime)
	fmt.Printf("Overall Time = %4.2f\n", initTime+simulTime+satTime)

freeAndExit:
	// (the C frees the match lists, dependencies and orders here; the
	// strashed copies of the networks are never deleted)
}
