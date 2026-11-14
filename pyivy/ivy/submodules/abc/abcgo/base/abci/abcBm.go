package abci

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	aiggraph "github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/gia"
	baseabc "github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/opt/sim"
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

//#include "base/abc/abc.h"
//#include "opt/sim/sim.h"
//#include "sat/bsat/satSolver.h"
//#include "bdd/extrab/extraBdd.h"

//ABC_NAMESPACE_IMPL_START

const FALSE = 0
const TRUE = 1

// int match1by1(Abc_Ntk_t * pNtk1, Vec_Ptr_t ** nodesInLevel1, Vec_Int_t ** iMatch1, Vec_Int_t ** iDep1, Vec_Int_t * matchedInputs1, int * iGroup1, Vec_Int_t ** oMatch1, int * oGroup1,
//                Abc_Ntk_t * pNtk2, Vec_Ptr_t ** nodesInLevel2, Vec_Int_t ** iMatch2, Vec_Int_t ** iDep2, Vec_Int_t * matchedInputs2, int * iGroup2, Vec_Int_t ** oMatch2, int * oGroup2,
//                Vec_Int_t * matchedOutputs1, Vec_Int_t * matchedOutputs2, Vec_Int_t * oMatchedGroups, Vec_Int_t * iNonSingleton, int ii, int idx);

// int Abc_NtkBmSat( Abc_Ntk_t * pNtk1, Abc_Ntk_t * pNtk2, Vec_Ptr_t * iMatchPairs, Vec_Ptr_t * oMatchPairs, Vec_Int_t * mismatch, int mode);

func getDependencies(pNtk *Abc_Ntk_t, iDep, oDep []*Vec_Int_t) {
	graph := baseabc.Abc_NtkStrashToGia(pNtk)
	supports := sim.Sim_ComputeFunSupp(graph, false)
	inputs, outputs := int(Abc_NtkPiNum(pNtk)), graph.Outputs()
	if len(iDep) < inputs || len(oDep) < len(outputs) {
		panic("abcBm: dependency vectors do not match network dimensions")
	}
	for output := range outputs {
		for input := 0; input < inputs; input++ {
			if sim.Sim_HasBit(supports[output], input) {
				Vec_IntPushOrder(oDep[output], int32(input))
			}
		}
	}
	for output := range outputs {
		for j := int32(0); j < Vec_IntSize(oDep[output]); j++ {
			Vec_IntPush(iDep[Vec_IntEntry(oDep[output], j)], int32(output))
		}
	}
}

func initMatchList(pNtk *Abc_Ntk_t, iDep, oDep, iMatch, oMatch []*Vec_Int_t, iLastItem, oLastItem *int32, iGroup, oGroup []int32, pEquivalence int32) {
	nInputs, nOutputs := int(Abc_NtkPiNum(pNtk)), int(Abc_NtkPoNum(pNtk))
	if pEquivalence == 0 {
		for i := range oMatch {
			Vec_IntFree(oMatch[i])
			oMatch[i] = nil
		}
		buckets := make([]*Vec_Int_t, nInputs+1)
		for i := range buckets {
			buckets[i] = Vec_IntAlloc(0)
		}
		for i := 0; i < nOutputs; i++ {
			Vec_IntPush(buckets[Vec_IntSize(oDep[i])], int32(i))
		}
		curr := 0
		for _, bucket := range buckets {
			if Vec_IntSize(bucket) == 0 {
				Vec_IntFree(bucket)
				continue
			}
			oMatch[curr] = bucket
			for j := int32(0); j < Vec_IntSize(bucket); j++ {
				oGroup[Vec_IntEntry(bucket, j)] = int32(curr)
			}
			curr++
		}
		*oLastItem = int32(curr)
	} else {
		for i := 0; i < nOutputs; i++ {
			if oMatch[i] == nil {
				oMatch[i] = Vec_IntAlloc(1)
			}
			Vec_IntPush(oMatch[i], int32(i))
			oGroup[i] = int32(i)
		}
		*oLastItem = int32(nOutputs)
	}
	for i := range iMatch {
		Vec_IntFree(iMatch[i])
		iMatch[i] = nil
	}
	buckets := make([]*Vec_Int_t, nOutputs+1)
	for i := range buckets {
		buckets[i] = Vec_IntAlloc(0)
	}
	for i := 0; i < nInputs; i++ {
		Vec_IntPush(buckets[Vec_IntSize(iDep[i])], int32(i))
	}
	curr := 0
	for _, bucket := range buckets {
		if Vec_IntSize(bucket) == 0 {
			Vec_IntFree(bucket)
			continue
		}
		iMatch[curr] = bucket
		for j := int32(0); j < Vec_IntSize(bucket); j++ {
			iGroup[Vec_IntEntry(bucket, j)] = int32(curr)
		}
		curr++
	}
	*iLastItem = int32(curr)
}

func iSortDependencies(pNtk *Abc_Ntk_t, iDep []*Vec_Int_t, oGroup []int32) {
	for i := 0; i < int(Abc_NtkPiNum(pNtk)); i++ {
		if Vec_IntSize(iDep[i]) <= 1 {
			continue
		}
		order := Vec_IntAlloc(Vec_IntSize(iDep[i]))
		for j := int32(0); j < Vec_IntSize(iDep[i]); j++ {
			Vec_IntPushUniqueOrder(order, oGroup[Vec_IntEntry(iDep[i], j)])
		}
		tmp := Vec_IntAlloc(Vec_IntSize(iDep[i]))
		for g := int32(0); g < Vec_IntSize(order); g++ {
			group := Vec_IntEntry(order, g)
			for j := int32(0); j < Vec_IntSize(iDep[i]); j++ {
				item := Vec_IntEntry(iDep[i], j)
				if oGroup[item] == group {
					Vec_IntPush(tmp, item)
				}
			}
		}
		Vec_IntFree(order)
		Vec_IntFree(iDep[i])
		iDep[i] = tmp
	}
}

func oSortDependencies(pNtk *Abc_Ntk_t, oDep []*Vec_Int_t, iGroup []int32) {
	for i := 0; i < int(Abc_NtkPoNum(pNtk)); i++ {
		if Vec_IntSize(oDep[i]) <= 1 {
			continue
		}
		order := Vec_IntAlloc(Vec_IntSize(oDep[i]))
		for j := int32(0); j < Vec_IntSize(oDep[i]); j++ {
			Vec_IntPushUniqueOrder(order, iGroup[Vec_IntEntry(oDep[i], j)])
		}
		tmp := Vec_IntAlloc(Vec_IntSize(oDep[i]))
		for g := int32(0); g < Vec_IntSize(order); g++ {
			group := Vec_IntEntry(order, g)
			for j := int32(0); j < Vec_IntSize(oDep[i]); j++ {
				item := Vec_IntEntry(oDep[i], j)
				if iGroup[item] == group {
					Vec_IntPush(tmp, item)
				}
			}
		}
		Vec_IntFree(order)
		Vec_IntFree(oDep[i])
		oDep[i] = tmp
	}
}

func oSplitByDep(pNtk *Abc_Ntk_t, oDep, oMatch []*Vec_Int_t, oGroup []int32, oLastItem *int32, iGroup []int32) int32 {
	added := int32(0)
	for i := int32(0); i < *oLastItem; i++ {
		if Vec_IntSize(oMatch[i]) == 1 {
			continue
		}
		encodings := make([]int32, Vec_IntSize(oMatch[i]))
		unique := Vec_IntAlloc(Vec_IntSize(oMatch[i]))
		for j := int32(0); j < Vec_IntSize(oMatch[i]); j++ {
			var encode int32
			for k := int32(0); k < Vec_IntSize(oDep[Vec_IntEntry(oMatch[i], j)]); k++ {
				encode += iGroup[Vec_IntEntry(oDep[Vec_IntEntry(oMatch[i], j)], k)]
			}
			encodings[j] = encode
			Vec_IntPushUniqueOrder(unique, encode)
		}
		for groupIndex := Vec_IntSize(unique) - 1; groupIndex > 0; groupIndex-- {
			value := Vec_IntEntry(unique, groupIndex)
			newGroup := *oLastItem + added
			if oMatch[newGroup] == nil {
				oMatch[newGroup] = Vec_IntAlloc(1)
			}
			remaining := Vec_IntAlloc(Vec_IntSize(oMatch[i]))
			for j := int32(0); j < Vec_IntSize(oMatch[i]); j++ {
				item := Vec_IntEntry(oMatch[i], j)
				if encodings[j] == value {
					Vec_IntPush(oMatch[newGroup], item)
					oGroup[item] = newGroup
				} else {
					Vec_IntPush(remaining, item)
				}
			}
			Vec_IntFree(oMatch[i])
			oMatch[i] = remaining
			added++
		}
		Vec_IntFree(unique)
	}
	*oLastItem += added
	return added
}

func iSplitByDep(pNtk *Abc_Ntk_t, iDep, iMatch []*Vec_Int_t, iGroup []int32, iLastItem *int32, oGroup []int32) int32 {
	added := int32(0)
	for i := int32(0); i < *iLastItem; i++ {
		if Vec_IntSize(iMatch[i]) == 1 {
			continue
		}
		encodings := make([]int32, Vec_IntSize(iMatch[i]))
		unique := Vec_IntAlloc(Vec_IntSize(iMatch[i]))
		for j := int32(0); j < Vec_IntSize(iMatch[i]); j++ {
			var encode int32
			for k := int32(0); k < Vec_IntSize(iDep[Vec_IntEntry(iMatch[i], j)]); k++ {
				encode += oGroup[Vec_IntEntry(iDep[Vec_IntEntry(iMatch[i], j)], k)]
			}
			encodings[j] = encode
			Vec_IntPushUniqueOrder(unique, encode)
		}
		for groupIndex := Vec_IntSize(unique) - 1; groupIndex > 0; groupIndex-- {
			value := Vec_IntEntry(unique, groupIndex)
			newGroup := *iLastItem + added
			if iMatch[newGroup] == nil {
				iMatch[newGroup] = Vec_IntAlloc(1)
			}
			remaining := Vec_IntAlloc(Vec_IntSize(iMatch[i]))
			for j := int32(0); j < Vec_IntSize(iMatch[i]); j++ {
				item := Vec_IntEntry(iMatch[i], j)
				if encodings[j] == value {
					Vec_IntPush(iMatch[newGroup], item)
					iGroup[item] = newGroup
				} else {
					Vec_IntPush(remaining, item)
				}
			}
			Vec_IntFree(iMatch[i])
			iMatch[i] = remaining
			added++
		}
		Vec_IntFree(unique)
	}
	*iLastItem += added
	return added
}

func findTopologicalOrder(pNtk *Abc_Ntk_t) []*Vec_Ptr_t {
	// The native routine builds reverse-fanout lists for one-input simulation.
	// The Go simulator evaluates the same strashed cone directly, so these
	// vectors remain empty compatibility slots.
	result := make([]*Vec_Ptr_t, Abc_NtkPiNum(pNtk))
	for i := range result {
		result[i] = Vec_PtrAlloc(0)
	}
	return result
}

func Abc_NtkSimulateOneNode(pNtk *Abc_Ntk_t, model []int32, input int32, topOrder []*Vec_Ptr_t) []int32 {
	_ = input
	_ = topOrder
	graph := baseabc.Abc_NtkStrashToGia(pNtk)
	inputs := make([]bool, len(graph.Inputs()))
	for i := range inputs {
		if i < len(model) {
			inputs[i] = model[i] != 0
		}
	}
	values, err := graph.Evaluate(inputs)
	if err != nil {
		panic(err)
	}
	outputs := make([]int32, len(graph.Outputs()))
	for i := range graph.Outputs() {
		if values[i] {
			outputs[i] = 1
		}
	}
	return outputs
}

func bmSplitByKey(match []*Vec_Int_t, group []int32, groupIndex int32, lastItem *int32, keys []int32) int32 {
	if Vec_IntSize(match[groupIndex]) < 2 {
		return 0
	}
	unique := Vec_IntAlloc(int32(len(keys)))
	for _, key := range keys {
		Vec_IntPushUniqueOrder(unique, key)
	}
	added := int32(0)
	for k := Vec_IntSize(unique) - 1; k > 0; k-- {
		key, newIndex := Vec_IntEntry(unique, k), *lastItem+added
		if match[newIndex] == nil {
			match[newIndex] = Vec_IntAlloc(1)
		}
		remaining := Vec_IntAlloc(Vec_IntSize(match[groupIndex]))
		for j := int32(0); j < Vec_IntSize(match[groupIndex]); j++ {
			item := Vec_IntEntry(match[groupIndex], j)
			if keys[j] == key {
				Vec_IntPush(match[newIndex], item)
				group[item] = newIndex
			} else {
				Vec_IntPush(remaining, item)
			}
		}
		Vec_IntFree(match[groupIndex])
		match[groupIndex] = remaining
		added++
	}
	Vec_IntFree(unique)
	*lastItem += added
	return added
}

func refineIOBySimulation(pNtk *Abc_Ntk_t, iMatch []*Vec_Int_t, iLastItem *int32, iGroup []int32, iDep []*Vec_Int_t, oMatch []*Vec_Int_t, oLastItem *int32, oGroup []int32, oDep []*Vec_Int_t, piValues string, observability []int32, topOrder []*Vec_Ptr_t) int32 {
	_ = oDep
	graph := baseabc.Abc_NtkStrashToGia(pNtk)
	inputCount, outputEdges := len(graph.Inputs()), graph.Outputs()
	model := make([]int32, inputCount)
	for i := 0; i < inputCount && i < len(piValues); i++ {
		model[i] = int32(piValues[i] - '0')
	}
	baseOutput := Abc_NtkSimulateOneNode(pNtk, model, 0, topOrder)
	refined := int32(0)
	lastOutput := *oLastItem
	for i := int32(0); i < lastOutput && *oLastItem != int32(len(outputEdges)); i++ {
		if Vec_IntSize(oMatch[i]) < 2 {
			continue
		}
		first := Vec_IntEntry(oMatch[i], 0)
		separates := false
		for j := int32(1); j < Vec_IntSize(oMatch[i]); j++ {
			if baseOutput[Vec_IntEntry(oMatch[i], j)] != baseOutput[first] {
				separates = true
				break
			}
		}
		if !separates {
			continue
		}
		newIndex := *oLastItem
		if oMatch[newIndex] == nil {
			oMatch[newIndex] = Vec_IntAlloc(1)
		}
		remaining := Vec_IntAlloc(Vec_IntSize(oMatch[i]))
		for j := int32(0); j < Vec_IntSize(oMatch[i]); j++ {
			item := Vec_IntEntry(oMatch[i], j)
			if baseOutput[item] != 0 {
				Vec_IntPush(oMatch[newIndex], item)
				oGroup[item] = newIndex
			} else {
				Vec_IntPush(remaining, item)
			}
		}
		Vec_IntFree(oMatch[i])
		oMatch[i] = remaining
		(*oLastItem)++
		refined = 1
	}
	if *oLastItem > lastOutput {
		iSortDependencies(pNtk, iDep, oGroup)
	}
	oComputed := make([]int32, len(outputEdges))
	lastInput := *iLastItem
	for groupIndex := int32(0); groupIndex < lastInput && *iLastItem != int32(inputCount); groupIndex++ {
		if Vec_IntSize(iMatch[groupIndex]) < 2 {
			continue
		}
		keys := make([]int32, Vec_IntSize(iMatch[groupIndex]))
		for j := int32(0); j < Vec_IntSize(iMatch[groupIndex]); j++ {
			input := Vec_IntEntry(iMatch[groupIndex], j)
			flipped := append([]int32(nil), model...)
			flipped[input] ^= 1
			output := Abc_NtkSimulateOneNode(pNtk, flipped, input, topOrder)
			var key int32
			for k := int32(0); k < Vec_IntSize(iDep[input]); k++ {
				o := Vec_IntEntry(iDep[input], k)
				if output[o] != 0 {
					key += oGroup[o] + 1
				}
				if baseOutput[o] != output[o] {
					oComputed[o] += groupIndex + 1
					observability[input]++
				}
			}
			keys[j] = key
		}
		if bmSplitByKey(iMatch, iGroup, groupIndex, iLastItem, keys) != 0 {
			refined = 1
		}
	}
	if *iLastItem > lastInput {
		oSortDependencies(pNtk, oDep, iGroup)
	}
	lastOutput = *oLastItem
	for groupIndex := int32(0); groupIndex < lastOutput && *oLastItem != int32(len(outputEdges)); groupIndex++ {
		if Vec_IntSize(oMatch[groupIndex]) < 2 {
			continue
		}
		keys := make([]int32, Vec_IntSize(oMatch[groupIndex]))
		for j := int32(0); j < Vec_IntSize(oMatch[groupIndex]); j++ {
			keys[j] = oComputed[Vec_IntEntry(oMatch[groupIndex], j)]
		}
		if bmSplitByKey(oMatch, oGroup, groupIndex, oLastItem, keys) != 0 {
			refined = 1
		}
	}
	return refined
}

func Abc_NtkMiterBm(pNtk1, pNtk2 *Abc_Ntk_t, iCurrMatch, oCurrMatch *Vec_Ptr_t) *Abc_Ntk_t {
	if pNtk1 == nil || pNtk2 == nil {
		return nil
	}
	graph1, graph2 := baseabc.Abc_NtkStrashToGia(pNtk1), baseabc.Abc_NtkStrashToGia(pNtk2)
	cis := func(n *Abc_Ntk_t) []*baseabc.Object {
		var result []*baseabc.Object
		for _, obj := range n.Objects() {
			if obj.IsCI() {
				result = append(result, obj)
			}
		}
		return result
	}
	cos := func(n *Abc_Ntk_t) []*baseabc.Object {
		var result []*baseabc.Object
		for _, obj := range n.Objects() {
			if obj.IsCO() {
				result = append(result, obj)
			}
		}
		return result
	}
	ci1, ci2, co1, co2 := cis(pNtk1), cis(pNtk2), cos(pNtk1), cos(pNtk2)
	inputs1, inputs2 := graph1.Inputs(), graph2.Inputs()
	outputs1, outputs2 := graph1.Outputs(), graph2.Outputs()
	if len(ci1) != len(inputs1) || len(ci2) != len(inputs2) || len(co1) != len(outputs1) || len(co2) != len(outputs2) {
		return nil
	}
	ciIndex1, ciIndex2 := make(map[*baseabc.Object]int, len(ci1)), make(map[*baseabc.Object]int, len(ci2))
	coIndex1, coIndex2 := make(map[*baseabc.Object]int, len(co1)), make(map[*baseabc.Object]int, len(co2))
	for i, o := range ci1 {
		ciIndex1[o] = i
	}
	for i, o := range ci2 {
		ciIndex2[o] = i
	}
	for i, o := range co1 {
		coIndex1[o] = i
	}
	for i, o := range co2 {
		coIndex2[o] = i
	}
	shared := aiggraph.NewGraph()
	aiggraph.Aig_ObjSetCopy(graph1, aiggraph.Edge(0), aiggraph.Edge(0))
	aiggraph.Aig_ObjSetCopy(graph2, aiggraph.Edge(0), aiggraph.Edge(0))
	mapInput := func(a, b *baseabc.Object) bool {
		i1, aIs1 := ciIndex1[a]
		i2, aIs2 := ciIndex2[a]
		j1, bIs1 := ciIndex1[b]
		j2, bIs2 := ciIndex2[b]
		var index1, index2 int
		switch {
		case aIs1 && bIs2:
			index1, index2 = i1, j2
		case bIs1 && aIs2:
			index1, index2 = j1, i2
		default:
			return false
		}
		input := shared.NewInput()
		aiggraph.Aig_ObjSetCopy(graph1, inputs1[index1], input)
		aiggraph.Aig_ObjSetCopy(graph2, inputs2[index2], input)
		return true
	}
	if iCurrMatch == nil {
		if len(ci1) != len(ci2) {
			return nil
		}
		for i := range ci1 {
			if !mapInput(ci1[i], ci2[i]) {
				return nil
			}
		}
	} else {
		if Vec_PtrSize(iCurrMatch)%2 != 0 {
			return nil
		}
		for i := int32(0); i < Vec_PtrSize(iCurrMatch); i += 2 {
			o1, ok1 := Vec_PtrEntry(iCurrMatch, i).(*baseabc.Object)
			o2, ok2 := Vec_PtrEntry(iCurrMatch, i+1).(*baseabc.Object)
			if !ok1 || !ok2 || !mapInput(o1, o2) {
				return nil
			}
		}
	}
	var pairs [][2]*baseabc.Object
	if oCurrMatch == nil {
		if len(co1) != len(co2) {
			return nil
		}
		for i := range co1 {
			pairs = append(pairs, [2]*baseabc.Object{co1[i], co2[i]})
		}
	} else {
		if Vec_PtrSize(oCurrMatch)%2 != 0 {
			return nil
		}
		for i := int32(0); i < Vec_PtrSize(oCurrMatch); i += 2 {
			o1, ok1 := Vec_PtrEntry(oCurrMatch, i).(*baseabc.Object)
			o2, ok2 := Vec_PtrEntry(oCurrMatch, i+1).(*baseabc.Object)
			if !ok1 || !ok2 {
				return nil
			}
			pairs = append(pairs, [2]*baseabc.Object{o1, o2})
		}
	}
	miter := aiggraph.Edge(1) // ABC_AigMiter returns constant zero for no pairs.
	for _, pair := range pairs {
		i1, aIs1 := coIndex1[pair[0]]
		i2, aIs2 := coIndex2[pair[0]]
		j1, bIs1 := coIndex1[pair[1]]
		j2, bIs2 := coIndex2[pair[1]]
		var index1, index2 int
		switch {
		case aIs1 && bIs2:
			index1, index2 = i1, j2
		case bIs1 && aIs2:
			index1, index2 = j1, i2
		default:
			return nil
		}
		a := aiggraph.Aig_ManDup_rec(shared, graph1, outputs1[index1])
		b := aiggraph.Aig_ManDup_rec(shared, graph2, outputs2[index2])
		miter = shared.OR(miter, shared.XOR(a, b))
	}
	shared.AddOutput(miter)
	manager := gia.Gia_ManFromAig(shared)
	defer gia.Gia_ManStop(manager)
	name := baseabc.Abc_NtkName(pNtk1) + "_" + baseabc.Abc_NtkName(pNtk2) + "_miter"
	return baseabc.Abc_NtkCreateFromGias(name, nil, manager)
}

func Abc_NtkVerifyReportErrorWithMismatch(pNtk1, pNtk2 *Abc_Ntk_t, model []int32, mismatch *Vec_Int_t) {
	graph1, graph2 := baseabc.Abc_NtkStrashToGia(pNtk1), baseabc.Abc_NtkStrashToGia(pNtk2)
	if len(graph1.Inputs()) != len(model) || len(graph2.Inputs()) != len(model) {
		return
	}
	inputs := make([]bool, len(model))
	for i, value := range model {
		inputs[i] = value != 0
	}
	values1, err1 := graph1.Evaluate(inputs)
	values2, err2 := graph2.Evaluate(inputs)
	if err1 != nil || err2 != nil {
		return
	}
	outputs1, outputs2 := graph1.Outputs(), graph2.Outputs()
	count := len(outputs1)
	if len(outputs2) < count {
		count = len(outputs2)
	}
	firstMismatch := -1
	for i := 0; i < count; i++ {
		if values1[i] != values2[i] {
			firstMismatch = i
			break
		}
	}
	if firstMismatch < 0 || mismatch == nil {
		return
	}
	supports := sim.Sim_ComputeFunSupp(graph1, false)
	support := supports[firstMismatch]
	for i, value := range model {
		if sim.Sim_HasBit(support, i) {
			Vec_IntPush(mismatch, int32(i))
			Vec_IntPush(mismatch, value)
		}
	}
}

func Abc_NtkMiterSatBm(pNtk *Abc_Ntk_t, nConfLimit, nInsLimit ABC_INT64_T, fVerbose int32, pNumConfs, pNumInspects *ABC_INT64_T) int32 {
	return Abc_NtkMiterSat(pNtk, nConfLimit, nInsLimit, fVerbose, pNumConfs, pNumInspects)
}

func Abc_NtkBmSat(pNtk1, pNtk2 *Abc_Ntk_t, iMatchPairs, oMatchPairs *Vec_Ptr_t, mismatch *Vec_Int_t, mode int32) int32 {
	if mode != 0 {
		fmt.Printf("Miter computation has failed.")
		return -1
	}
	miter := Abc_NtkMiterBm(pNtk1, pNtk2, iMatchPairs, oMatchPairs)
	if miter == nil {
		fmt.Printf("Miter computation has failed.")
		return -1
	}
	defer baseabc.Abc_NtkDelete(miter)
	constant := Abc_NtkMiterIsConstant(miter)
	if constant == 0 {
		return 0
	}
	if constant == 1 {
		return 1
	}
	status := Abc_NtkMiterSat(miter, 10000, 0, 0, nil, nil)
	if status == 0 && mismatch != nil && len(miter.Model) > 0 {
		Abc_NtkVerifyReportErrorWithMismatch(pNtk1, pNtk2, miter.Model, mismatch)
	}
	return status
}

func checkEquivalence(pNtk1 *Abc_Ntk_t, matchedInputs1, matchedOutputs1 *Vec_Int_t, pNtk2 *Abc_Ntk_t, matchedInputs2, matchedOutputs2 *Vec_Int_t) int32 {
	iMatchPairs := Vec_PtrAlloc(2 * Abc_NtkPiNum(pNtk1))
	oMatchPairs := Vec_PtrAlloc(2 * Abc_NtkPoNum(pNtk1))
	defer Vec_PtrFree(iMatchPairs)
	defer Vec_PtrFree(oMatchPairs)
	for i := int32(0); i < Abc_NtkPiNum(pNtk1); i++ {
		Vec_PtrPush(iMatchPairs, Abc_NtkPi(pNtk2, Vec_IntEntry(matchedInputs2, i)))
		Vec_PtrPush(iMatchPairs, Abc_NtkPi(pNtk1, Vec_IntEntry(matchedInputs1, i)))
	}
	for i := int32(0); i < Abc_NtkPoNum(pNtk1); i++ {
		Vec_PtrPush(oMatchPairs, Abc_NtkPo(pNtk2, Vec_IntEntry(matchedOutputs2, i)))
		Vec_PtrPush(oMatchPairs, Abc_NtkPo(pNtk1, Vec_IntEntry(matchedOutputs1, i)))
	}
	result := Abc_NtkBmSat(pNtk1, pNtk2, iMatchPairs, oMatchPairs, nil, 0)
	if result != 0 {
		fmt.Printf("*** Circuits are equivalent ***\n")
	} else {
		fmt.Printf("*** Circuits are NOT equivalent ***\n")
	}
	return result
}

func computeCofactor(pNtk *Abc_Ntk_t, nodesInLevel []*Vec_Ptr_t, bitVector []int32, currInputs *Vec_Int_t) *Abc_Ntk_t {
	_ = nodesInLevel
	source := baseabc.Abc_NtkStrashToGia(pNtk)
	destination := aiggraph.NewGraph()
	nodes := source.Nodes()
	mapped := make([]aiggraph.Edge, len(nodes))
	known := make([]bool, len(nodes))
	known[0], mapped[0] = true, aiggraph.Edge(0)
	for i, input := range source.Inputs() {
		if i < len(bitVector) && bitVector[i] != 0 {
			known[input.NodeID()], mapped[input.NodeID()] = true, aiggraph.Edge(0)
		}
	}
	for i := int32(0); i < Vec_IntSize(currInputs); i++ {
		index := int(Vec_IntEntry(currInputs, i))
		if index < 0 || index >= len(source.Inputs()) {
			continue
		}
		input := source.Inputs()[index]
		known[input.NodeID()], mapped[input.NodeID()] = true, destination.NewInput()
	}
	for id := 1; id < len(nodes); id++ {
		node := nodes[id]
		switch node.Kind {
		case aiggraph.And, aiggraph.ExclusiveOr:
			a, b := node.Fanins[0], node.Fanins[1]
			if !known[a.NodeID()] || !known[b.NodeID()] {
				continue
			}
			aMapped, bMapped := mapped[a.NodeID()], mapped[b.NodeID()]
			if a.Complemented() {
				aMapped = aMapped.Not()
			}
			if b.Complemented() {
				bMapped = bMapped.Not()
			}
			if node.Kind == aiggraph.ExclusiveOr {
				mapped[id] = destination.XOR(aMapped, bMapped)
			} else {
				mapped[id] = destination.AND(aMapped, bMapped)
			}
			known[id] = true
		case aiggraph.Buffer, aiggraph.Output:
			a := node.Fanins[0]
			if known[a.NodeID()] {
				mapped[id], known[id] = mapped[a.NodeID()], true
				if a.Complemented() {
					mapped[id] = mapped[id].Not()
				}
			}
		}
	}
	for _, output := range source.Outputs() {
		driver := nodes[output.NodeID()].Fanins[0]
		value := aiggraph.Edge(1) // An omitted/unmapped cofactor input is false.
		if known[driver.NodeID()] {
			value = mapped[driver.NodeID()]
			if driver.Complemented() {
				value = value.Not()
			}
		}
		destination.AddOutput(value)
	}
	manager := gia.Gia_ManFromAig(destination)
	defer gia.Gia_ManStop(manager)
	return baseabc.Abc_NtkCreateFromGias("subNtk", nil, manager)
}

func matchNonSingletonOutputs(pNtk1 *Abc_Ntk_t, nodesInLevel1 []*Vec_Ptr_t, iMatch1, iDep1 []*Vec_Int_t, matchedInputs1 *Vec_Int_t, iGroup1 []int32, oMatch1 []*Vec_Int_t, oGroup1 []int32, pNtk2 *Abc_Ntk_t, nodesInLevel2 []*Vec_Ptr_t, iMatch2, iDep2 []*Vec_Int_t, matchedInputs2 *Vec_Int_t, iGroup2 []int32, oMatch2 []*Vec_Int_t, oGroup2 []int32, matchedOutputs1, matchedOutputs2, oMatchedGroups, iNonSingleton *Vec_Int_t, subNtk1, subNtk2 *Abc_Ntk_t, oMatchPairs *Vec_Ptr_t, oNonSingleton *Vec_Int_t, oI, idx, ii, iidx int32) int32 {
	if oI == Vec_IntSize(oNonSingleton) {
		if iNonSingleton == nil {
			return 1
		}
		return match1by1(pNtk1, nodesInLevel1, iMatch1, iDep1, matchedInputs1, iGroup1, oMatch1, oGroup1, pNtk2, nodesInLevel2, iMatch2, iDep2, matchedInputs2, iGroup2, oMatch2, oGroup2, matchedOutputs1, matchedOutputs2, oMatchedGroups, iNonSingleton, ii, iidx)
	}
	group := Vec_IntEntry(oNonSingleton, oI)
	if idx < 0 || idx >= Vec_IntSize(oMatch1[group]) {
		return 0
	}
	output1 := Vec_IntEntry(oMatch1[group], idx)
	Vec_PtrPush(oMatchPairs, Abc_NtkPo(subNtk1, output1))
	Vec_IntPush(matchedOutputs1, output1)
	skip := make([]bool, Vec_IntSize(oMatch2[group]))
	found := int32(0)
	for j := int32(0); j < Vec_IntSize(oMatch2[group]) && found == 0; j++ {
		output2 := Vec_IntEntry(oMatch2[group], j)
		if output2 < 0 || skip[j] {
			continue
		}
		Vec_PtrPush(oMatchPairs, Abc_NtkPo(subNtk2, output2))
		Vec_IntPush(matchedOutputs2, output2)
		mismatch := Vec_IntAlloc(10)
		if Abc_NtkBmSat(subNtk1, subNtk2, nil, oMatchPairs, mismatch, 0) != 0 {
			Vec_IntWriteEntry(oMatch2[group], j, -1)
			nextO, nextIdx := oI, idx+1
			if nextIdx == Vec_IntSize(oMatch1[group]) {
				nextO++
				nextIdx = 0
			}
			found = matchNonSingletonOutputs(pNtk1, nodesInLevel1, iMatch1, iDep1, matchedInputs1, iGroup1, oMatch1, oGroup1, pNtk2, nodesInLevel2, iMatch2, iDep2, matchedInputs2, iGroup2, oMatch2, oGroup2, matchedOutputs1, matchedOutputs2, oMatchedGroups, iNonSingleton, subNtk1, subNtk2, oMatchPairs, oNonSingleton, nextO, nextIdx, ii, iidx)
			Vec_IntWriteEntry(oMatch2[group], j, output2)
		} else {
			model := make([]int32, Abc_NtkPiNum(subNtk1))
			for k := int32(0); k+1 < Vec_IntSize(mismatch); k += 2 {
				input := Vec_IntEntry(mismatch, k)
				if input >= 0 && int(input) < len(model) {
					model[input] = Vec_IntEntry(mismatch, k+1)
				}
			}
			values1 := Abc_NtkSimulateOneNode(subNtk1, model, 0, nil)
			values2 := Abc_NtkSimulateOneNode(subNtk2, model, 0, nil)
			for k := int32(0); k < Vec_IntSize(oMatch2[group]); k++ {
				candidate := Vec_IntEntry(oMatch2[group], k)
				if candidate >= 0 && output1 < int32(len(values1)) && candidate < int32(len(values2)) && values1[output1] != values2[candidate] {
					skip[k] = true
				}
			}
		}
		Vec_IntFree(mismatch)
		if found == 0 {
			Vec_PtrPop(oMatchPairs)
			Vec_IntPop(matchedOutputs2)
		}
	}
	if found == 0 {
		Vec_PtrPop(oMatchPairs)
		Vec_IntPop(matchedOutputs1)
	}
	return found
}

func match1by1(pNtk1 *Abc_Ntk_t, nodesInLevel1 []*Vec_Ptr_t, iMatch1, iDep1 []*Vec_Int_t, matchedInputs1 *Vec_Int_t, iGroup1 []int32, oMatch1 []*Vec_Int_t, oGroup1 []int32, pNtk2 *Abc_Ntk_t, nodesInLevel2 []*Vec_Ptr_t, iMatch2, iDep2 []*Vec_Int_t, matchedInputs2 *Vec_Int_t, iGroup2 []int32, oMatch2 []*Vec_Int_t, oGroup2 []int32, matchedOutputs1, matchedOutputs2, oMatchedGroups, iNonSingleton *Vec_Int_t, ii, idx int32) int32 {
	if ii == Vec_IntSize(iNonSingleton) {
		return 1
	}
	group := Vec_IntEntry(iNonSingleton, ii)
	if idx == Vec_IntSize(iMatch1[group]) {
		return match1by1(pNtk1, nodesInLevel1, iMatch1, iDep1, matchedInputs1, iGroup1, oMatch1, oGroup1, pNtk2, nodesInLevel2, iMatch2, iDep2, matchedInputs2, iGroup2, oMatch2, oGroup2, matchedOutputs1, matchedOutputs2, oMatchedGroups, iNonSingleton, ii+1, 0)
	}
	oNonSingleton := Vec_IntAlloc(10)
	oMatchPairs := Vec_PtrAlloc(100)
	defer Vec_IntFree(oNonSingleton)
	defer Vec_PtrFree(oMatchPairs)
	input1 := Vec_IntEntry(iMatch1[group], idx)
	Vec_IntPush(matchedInputs1, input1)
	nextIdx := idx + 1
	if nextIdx == 1 {
		for j := int32(0); j < Vec_IntSize(iDep1[input1]); j++ {
			outputGroup := oGroup1[Vec_IntEntry(iDep1[input1], j)]
			if Vec_IntSize(oMatch1[outputGroup]) == 1 || Vec_IntFind(oMatchedGroups, outputGroup) >= 0 {
				continue
			}
			Vec_IntPushUnique(oNonSingleton, outputGroup)
			Vec_IntPushUnique(oMatchedGroups, outputGroup)
		}
	}
	subNtk1 := computeCofactor(pNtk1, nodesInLevel1, nil, matchedInputs1)
	defer Abc_NtkDelete(subNtk1)
	found := int32(0)
	skip := make([]bool, Vec_IntSize(iMatch1[group]))
	for j := idx; j < Vec_IntSize(iMatch2[group]) && found == 0; j++ {
		if j < int32(len(skip)) && skip[j] {
			continue
		}
		input2 := Vec_IntEntry(iMatch2[group], j)
		Vec_IntPush(matchedInputs2, input2)
		subNtk2 := computeCofactor(pNtk2, nodesInLevel2, nil, matchedInputs2)
		for m := int32(0); m < Vec_IntSize(matchedOutputs1); m++ {
			Vec_PtrPush(oMatchPairs, Abc_NtkPo(subNtk1, Vec_IntEntry(matchedOutputs1, m)))
			Vec_PtrPush(oMatchPairs, Abc_NtkPo(subNtk2, Vec_IntEntry(matchedOutputs2, m)))
		}
		mismatch := Vec_IntAlloc(10)
		if Abc_NtkBmSat(subNtk2, subNtk1, nil, oMatchPairs, mismatch, 0) != 0 {
			if idx != j {
				temp := Vec_IntEntry(iMatch2[group], idx)
				Vec_IntWriteEntry(iMatch2[group], idx, input2)
				Vec_IntWriteEntry(iMatch2[group], j, temp)
			}
			found = matchNonSingletonOutputs(pNtk1, nodesInLevel1, iMatch1, iDep1, matchedInputs1, iGroup1, oMatch1, oGroup1, pNtk2, nodesInLevel2, iMatch2, iDep2, matchedInputs2, iGroup2, oMatch2, oGroup2, matchedOutputs1, matchedOutputs2, oMatchedGroups, iNonSingleton, subNtk1, subNtk2, oMatchPairs, oNonSingleton, 0, 0, ii, nextIdx)
			if idx != j {
				Vec_IntWriteEntry(iMatch2[group], j, Vec_IntEntry(iMatch2[group], idx))
				Vec_IntWriteEntry(iMatch2[group], idx, input2)
			}
		} else {
			// A failed SAT match supplies a counterexample.  Cofactor both
			// networks on the already matched inputs and compare the remaining
			// support counts, as the C implementation does to prune candidates.
			bitVector1 := make([]int32, Abc_NtkPiNum(pNtk1))
			bitVector2 := make([]int32, Abc_NtkPiNum(pNtk2))
			for k := int32(0); k+1 < Vec_IntSize(mismatch); k += 2 {
				matchedIndex := Vec_IntEntry(mismatch, k)
				if matchedIndex < 0 || matchedIndex >= Vec_IntSize(matchedInputs1) || matchedIndex >= Vec_IntSize(matchedInputs2) {
					continue
				}
				value := Vec_IntEntry(mismatch, k+1)
				input1, input2 := Vec_IntEntry(matchedInputs1, matchedIndex), Vec_IntEntry(matchedInputs2, matchedIndex)
				if input1 >= 0 && input1 < int32(len(bitVector1)) {
					bitVector1[input1] = value
				}
				if input2 >= 0 && input2 < int32(len(bitVector2)) {
					bitVector2[input2] = value
				}
			}
			currInputs1, currInputs2 := Vec_IntAlloc(10), Vec_IntAlloc(10)
			for k := idx; k < Vec_IntSize(iMatch1[group]); k++ {
				Vec_IntPush(currInputs1, Vec_IntEntry(iMatch1[group], k))
				Vec_IntPush(currInputs2, Vec_IntEntry(iMatch2[group], k))
			}
			for k := int32(0); k < Abc_NtkPiNum(pNtk1); k++ {
				if Vec_IntFind(matchedInputs1, k) < 0 {
					Vec_IntPushUnique(currInputs1, k)
				}
				if k < Abc_NtkPiNum(pNtk2) && Vec_IntFind(matchedInputs2, k) < 0 {
					Vec_IntPushUnique(currInputs2, k)
				}
			}
			fpNtk1 := computeCofactor(pNtk1, nodesInLevel1, bitVector1, currInputs1)
			fpNtk2 := computeCofactor(pNtk2, nodesInLevel2, bitVector2, currInputs2)
			graph1, graph2 := baseabc.Abc_NtkStrashToGia(fpNtk1), baseabc.Abc_NtkStrashToGia(fpNtk2)
			support1 := sim.Sim_ComputeFunSupp(graph1, false)
			support2 := sim.Sim_ComputeFunSupp(graph2, false)
			var count1 int32
			for output, bits := range support1 {
				if sim.Sim_HasBit(bits, 0) {
					count1 += int32(Vec_IntFind(matchedOutputs1, int32(output)) + 1)
				}
			}
			count2 := make([]int32, Vec_IntSize(iMatch2[group])-idx)
			for output, bits := range support2 {
				weight := int32(Vec_IntFind(matchedOutputs2, int32(output)) + 1)
				for input := range count2 {
					if sim.Sim_HasBit(bits, input) {
						count2[input] += weight
					}
				}
			}
			for offset := int32(0); offset < int32(len(count2)); offset++ {
				candidate := idx + offset
				if candidate >= 0 && candidate < int32(len(skip)) && count2[offset] != count1 {
					skip[candidate] = true
				}
			}
			Abc_NtkDelete(fpNtk1)
			Abc_NtkDelete(fpNtk2)
			Vec_IntFree(currInputs1)
			Vec_IntFree(currInputs2)
		}
		Vec_IntFree(mismatch)
		Vec_PtrClear(oMatchPairs)
		Abc_NtkDelete(subNtk2)
		if found == 0 {
			Vec_IntPop(matchedInputs2)
		}
	}
	if found == 0 {
		Vec_IntPop(matchedInputs1)
		if nextIdx == 1 {
			for i := int32(0); i < Vec_IntSize(oNonSingleton); i++ {
				Vec_IntPop(oMatchedGroups)
			}
		}
	}
	return found
}

func refineBySAT(pNtk1 *Abc_Ntk_t, iMatch1 []*Vec_Int_t, iGroup1 []int32, iDep1 []*Vec_Int_t, iLastItem1 *int32, oMatch1 []*Vec_Int_t, oGroup1 []int32, oDep1 []*Vec_Int_t, oLastItem1 *int32, observability1 []int32, pNtk2 *Abc_Ntk_t, iMatch2 []*Vec_Int_t, iGroup2 []int32, iDep2 []*Vec_Int_t, iLastItem2 *int32, oMatch2 []*Vec_Int_t, oGroup2 []int32, oDep2 []*Vec_Int_t, oLastItem2 *int32, observability2 []int32) float32 {
	start := time.Now()
	iNonSingleton := Vec_IntAlloc(10)
	matchedInputs1, matchedInputs2 := Vec_IntAlloc(Abc_NtkPiNum(pNtk1)), Vec_IntAlloc(Abc_NtkPiNum(pNtk2))
	matchedOutputs1, matchedOutputs2 := Vec_IntAlloc(Abc_NtkPoNum(pNtk1)), Vec_IntAlloc(Abc_NtkPoNum(pNtk2))
	oMatchedGroups := Vec_IntAlloc(10)
	defer Vec_IntFree(iNonSingleton)
	defer Vec_IntFree(matchedInputs1)
	defer Vec_IntFree(matchedInputs2)
	defer Vec_IntFree(matchedOutputs1)
	defer Vec_IntFree(matchedOutputs2)
	defer Vec_IntFree(oMatchedGroups)
	nodesInLevel1 := make([]*Vec_Ptr_t, int(baseabc.Abc_AigLevel(pNtk1))+1)
	nodesInLevel2 := make([]*Vec_Ptr_t, int(baseabc.Abc_AigLevel(pNtk2))+1)
	for i := range nodesInLevel1 {
		nodesInLevel1[i] = Vec_PtrAlloc(0)
	}
	for i := range nodesInLevel2 {
		nodesInLevel2[i] = Vec_PtrAlloc(0)
	}
	defer func() {
		for _, v := range nodesInLevel1 {
			Vec_PtrFree(v)
		}
		for _, v := range nodesInLevel2 {
			Vec_PtrFree(v)
		}
	}()
	for i := int32(0); i < *iLastItem1; i++ {
		if Vec_IntSize(iMatch1[i]) == 1 {
			Vec_IntPush(matchedInputs1, Vec_IntEntryLast(iMatch1[i]))
			Vec_IntPush(matchedInputs2, Vec_IntEntryLast(iMatch2[i]))
		} else {
			Vec_IntPush(iNonSingleton, i)
		}
	}
	for i := int32(0); i < *oLastItem1; i++ {
		if Vec_IntSize(oMatch1[i]) == 1 {
			Vec_IntPush(matchedOutputs1, Vec_IntEntryLast(oMatch1[i]))
			Vec_IntPush(matchedOutputs2, Vec_IntEntryLast(oMatch2[i]))
		}
	}
	for i := int32(0); i < Vec_IntSize(iNonSingleton)-1; i++ {
		for j := i + 1; j < Vec_IntSize(iNonSingleton); j++ {
			a, b := Vec_IntEntry(iNonSingleton, i), Vec_IntEntry(iNonSingleton, j)
			inputA, inputB := Vec_IntEntry(iMatch2[a], 0), Vec_IntEntry(iMatch2[b], 0)
			swap := observability2[inputB] > observability2[inputA]
			if observability2[inputB] == observability2[inputA] {
				swap = Vec_IntSize(iMatch2[b]) < Vec_IntSize(iMatch2[a])
			}
			if swap {
				Vec_IntWriteEntry(iNonSingleton, i, b)
				Vec_IntWriteEntry(iNonSingleton, j, a)
			}
		}
	}
	matched := match1by1(pNtk1, nodesInLevel1, iMatch1, iDep1, matchedInputs1, iGroup1, oMatch1, oGroup1, pNtk2, nodesInLevel2, iMatch2, iDep2, matchedInputs2, iGroup2, oMatch2, oGroup2, matchedOutputs1, matchedOutputs2, oMatchedGroups, iNonSingleton, 0, 0) != 0
	if matched && Vec_IntSize(matchedOutputs1) != Abc_NtkPoNum(pNtk1) {
		oNonSingleton := Vec_IntAlloc(10)
		oMatchPairs := Vec_PtrAlloc(Abc_NtkPoNum(pNtk1) * 2)
		for i := int32(0); i < *oLastItem1; i++ {
			if Vec_IntSize(oMatch1[i]) > 1 && Vec_IntFind(oMatchedGroups, i) < 0 {
				Vec_IntPush(oNonSingleton, i)
			}
		}
		subNtk1 := computeCofactor(pNtk1, nodesInLevel1, nil, matchedInputs1)
		subNtk2 := computeCofactor(pNtk2, nodesInLevel2, nil, matchedInputs2)
		matched = matchNonSingletonOutputs(pNtk1, nodesInLevel1, iMatch1, iDep1, matchedInputs1, iGroup1, oMatch1, oGroup1, pNtk2, nodesInLevel2, iMatch2, iDep2, matchedInputs2, iGroup2, oMatch2, oGroup2, matchedOutputs1, matchedOutputs2, oMatchedGroups, nil, subNtk1, subNtk2, oMatchPairs, oNonSingleton, 0, 0, 0, 0) != 0
		Vec_IntFree(oNonSingleton)
		Vec_PtrFree(oMatchPairs)
		Abc_NtkDelete(subNtk1)
		Abc_NtkDelete(subNtk2)
	}
	if matched {
		checkEquivalence(pNtk1, matchedInputs1, matchedOutputs1, pNtk2, matchedInputs2, matchedOutputs2)
		result, err := os.Create("IOmatch.txt")
		if err == nil {
			fmt.Fprintf(result, "I/O = %d / %d \n\n", Abc_NtkPiNum(pNtk1), Abc_NtkPoNum(pNtk1))
			for i := int32(0); i < Vec_IntSize(matchedInputs1); i++ {
				fmt.Fprintf(result, "{%s}\t{%s}\n", baseabc.Abc_ObjName(Abc_NtkPi(pNtk1, Vec_IntEntry(matchedInputs1, i))), baseabc.Abc_ObjName(Abc_NtkPi(pNtk2, Vec_IntEntry(matchedInputs2, i))))
			}
			fmt.Fprintln(result, "\n-----------------------------------------")
			for i := int32(0); i < Vec_IntSize(matchedOutputs1); i++ {
				fmt.Fprintf(result, "{%s}\t{%s}\n", baseabc.Abc_ObjName(Abc_NtkPo(pNtk1, Vec_IntEntry(matchedOutputs1, i))), baseabc.Abc_ObjName(Abc_NtkPo(pNtk2, Vec_IntEntry(matchedOutputs2, i))))
			}
			_ = result.Close()
		}
	}
	return float32(time.Since(start).Seconds())
}

func checkListConsistency(iMatch1, oMatch1, iMatch2, oMatch2 []*Vec_Int_t, iLastItem1, oLastItem1, iLastItem2, oLastItem2 int32) int32 {
	if iLastItem1 != iLastItem2 || oLastItem1 != oLastItem2 {
		return FALSE
	}
	return TRUE
}

func bmGateWay(pNtk1, pNtk2 *Abc_Ntk_t, pEquivalence int32) {
	totalStart := time.Now()
	if pNtk1 == nil || pNtk2 == nil {
		return
	}
	if Abc_NtkPiNum(pNtk1) != Abc_NtkPiNum(pNtk2) || Abc_NtkPoNum(pNtk1) != Abc_NtkPoNum(pNtk2) {
		fmt.Printf("I/O dependencies of two circuits are different.\n")
		return
	}
	pNtk1, pNtk2 = Abc_NtkStrash(pNtk1, 0, 0, 0), Abc_NtkStrash(pNtk2, 0, 0, 0)
	if pNtk1 == nil || pNtk2 == nil {
		return
	}
	nInputs, nOutputs := int(Abc_NtkPiNum(pNtk1)), int(Abc_NtkPoNum(pNtk1))
	iDep1, iDep2 := make([]*Vec_Int_t, nInputs), make([]*Vec_Int_t, nInputs)
	oDep1, oDep2 := make([]*Vec_Int_t, nOutputs), make([]*Vec_Int_t, nOutputs)
	iMatch1, iMatch2 := make([]*Vec_Int_t, nInputs), make([]*Vec_Int_t, nInputs)
	oMatch1, oMatch2 := make([]*Vec_Int_t, nOutputs), make([]*Vec_Int_t, nOutputs)
	iGroup1, iGroup2 := make([]int32, nInputs), make([]int32, nInputs)
	oGroup1, oGroup2 := make([]int32, nOutputs), make([]int32, nOutputs)
	observability1, observability2 := make([]int32, nInputs), make([]int32, nInputs)
	for i := range iDep1 {
		iDep1[i] = Vec_IntAlloc(1)
		iDep2[i] = Vec_IntAlloc(1)
		iMatch1[i] = Vec_IntAlloc(1)
		iMatch2[i] = Vec_IntAlloc(1)
	}
	for i := range oDep1 {
		oDep1[i] = Vec_IntAlloc(1)
		oDep2[i] = Vec_IntAlloc(1)
		oMatch1[i] = Vec_IntAlloc(1)
		oMatch2[i] = Vec_IntAlloc(1)
	}
	var iLastItem1, oLastItem1, iLastItem2, oLastItem2 int32
	defer func() {
		for i := range iMatch1 {
			Vec_IntFree(iMatch1[i])
			Vec_IntFree(iMatch2[i])
		}
		for i := range oMatch1 {
			Vec_IntFree(oMatch1[i])
			Vec_IntFree(oMatch2[i])
		}
		for i := range iDep1 {
			Vec_IntFree(iDep1[i])
			Vec_IntFree(iDep2[i])
		}
		for i := range oDep1 {
			Vec_IntFree(oDep1[i])
			Vec_IntFree(oDep2[i])
		}
	}()
	fmt.Printf("Network  strashing is done!\n")
	getDependencies(pNtk1, iDep1, oDep1)
	getDependencies(pNtk2, iDep2, oDep2)
	fmt.Printf("Getting dependencies is done!\n")
	initMatchList(pNtk1, iDep1, oDep1, iMatch1, &iLastItem1, oMatch1, &oLastItem1, iGroup1, oGroup1, pEquivalence)
	initMatchList(pNtk2, iDep2, oDep2, iMatch2, &iLastItem2, oMatch2, &oLastItem2, iGroup2, oGroup2, pEquivalence)
	fmt.Printf("Initializing match lists is done!\n")
	if checkListConsistency(iMatch1, oMatch1, iMatch2, oMatch2, iLastItem1, oLastItem1, iLastItem2, oLastItem2) == 0 {
		fmt.Printf("I/O dependencies of two circuits are different.\n")
		return
	}
	fmt.Printf("Refining IOs by dependencies ...")
	for {
		iAdded, oAdded := int32(1), int32(1)
		for iAdded != 0 || oAdded != 0 {
			if oAdded != 0 {
				iSortDependencies(pNtk1, iDep1, oGroup1)
				iSortDependencies(pNtk2, iDep2, oGroup2)
			}
			if iAdded != 0 {
				oSortDependencies(pNtk1, oDep1, iGroup1)
				oSortDependencies(pNtk2, oDep2, iGroup2)
			}
			if iLastItem1 < int32(nInputs) {
				_ = iSplitByDep(pNtk1, iDep1, iMatch1, iGroup1, &iLastItem1, oGroup1)
				if oLastItem1 < int32(nOutputs) {
					_ = oSplitByDep(pNtk1, oDep1, oMatch1, oGroup1, &oLastItem1, iGroup1)
				}
			}
			if iLastItem2 < int32(nInputs) {
				iAdded = iSplitByDep(pNtk2, iDep2, iMatch2, iGroup2, &iLastItem2, oGroup2)
			} else {
				iAdded = 0
			}
			if oLastItem2 < int32(nOutputs) {
				oAdded = oSplitByDep(pNtk2, oDep2, oMatch2, oGroup2, &oLastItem2, iGroup2)
			} else {
				oAdded = 0
			}
			if checkListConsistency(iMatch1, oMatch1, iMatch2, oMatch2, iLastItem1, oLastItem1, iLastItem2, oLastItem2) == 0 {
				fmt.Printf("I/O dependencies of two circuits are different.\n")
				return
			}
		}
		break
	}
	fmt.Printf(" done!\n")
	topOrder1, topOrder2 := findTopologicalOrder(pNtk1), findTopologicalOrder(pNtk2)
	defer func() {
		for _, v := range topOrder1 {
			Vec_PtrFree(v)
		}
		for _, v := range topOrder2 {
			Vec_PtrFree(v)
		}
	}()
	piValues1, piValues2 := make([]byte, nInputs), make([]byte, nInputs)
	for i := range piValues1 {
		piValues1[i] = '0'
		piValues2[i] = '0'
	}
	fmt.Printf("Refining IOs by simulation ...")
	initTime := float32(time.Since(totalStart).Seconds())
	simulationStart := time.Now()
	for tries := 0; tries <= 200; {
		for i := int32(0); i < iLastItem1; i++ {
			value := byte('0' + rand.Uint32()%2)
			if Vec_IntSize(iMatch1[i]) != Vec_IntSize(iMatch2[i]) {
				fmt.Printf("Input refinement by simulation finds two circuits different.\n")
				return
			}
			for j := int32(0); j < Vec_IntSize(iMatch1[i]); j++ {
				piValues1[Vec_IntEntry(iMatch1[i], j)] = value
				piValues2[Vec_IntEntry(iMatch2[i], j)] = value
			}
		}
		refined1 := refineIOBySimulation(pNtk1, iMatch1, &iLastItem1, iGroup1, iDep1, oMatch1, &oLastItem1, oGroup1, oDep1, string(piValues1), observability1, topOrder1)
		refined2 := refineIOBySimulation(pNtk2, iMatch2, &iLastItem2, iGroup2, iDep2, oMatch2, &oLastItem2, oGroup2, oDep2, string(piValues2), observability2, topOrder2)
		if (refined1 != 0) != (refined2 != 0) || checkListConsistency(iMatch1, oMatch1, iMatch2, oMatch2, iLastItem1, oLastItem1, iLastItem2, oLastItem2) == 0 {
			fmt.Printf("Input refinement by simulation finds two circuits different.\n")
			return
		}
		if refined1 != 0 && refined2 != 0 {
			tries = 0
		} else {
			tries++
		}
	}
	fmt.Printf(" done!\nSAT-based search started ...\n")
	simulTime := float32(time.Since(simulationStart).Seconds())
	satTime := refineBySAT(pNtk1, iMatch1, iGroup1, iDep1, &iLastItem1, oMatch1, oGroup1, oDep1, &oLastItem1, observability1, pNtk2, iMatch2, iGroup2, iDep2, &iLastItem2, oMatch2, oGroup2, oDep2, &oLastItem2, observability2)
	fmt.Printf("Init Time = %4.2f\n", initTime)
	fmt.Printf("Simulation Time = %4.2f\n", simulTime)
	fmt.Printf("SAT Time = %4.2f\n", satTime)
	fmt.Printf("Overall Time = %4.2f\n", initTime+simulTime+satTime)
}
