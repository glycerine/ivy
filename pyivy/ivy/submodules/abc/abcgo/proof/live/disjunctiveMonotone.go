package live

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/proof/pdr"
	"os"
	"slices"
	"strings"
)

type antecedentConsequentVectorsStruct struct{ attrAntecedents, attrConsequentCandidates []int32 }

func allocAntecedentConsequentVectorsStruct() *antecedentConsequentVectorsStruct {
	return new(antecedentConsequentVectorsStruct)
}
func createDisjunctiveMonotoneTester(p *aig.Graph, indexes *aigPoIndices, vectors *antecedentConsequentVectorsStruct, start *int32) *aig.Graph {
	return createDisjunctiveMonotoneTesterMode(p, indexes, vectors, start, true)
}
func createDisjunctiveMonotoneTesterMode(p *aig.Graph, indexes *aigPoIndices, vectors *antecedentConsequentVectorsStruct, start *int32, source bool) *aig.Graph {
	if indexes.attrPendingSignalIndex < 0 {
		panic("live: pending signal missing")
	}
	c := newGraphCopy(p, "_monotone")
	if source {
		c = newSourceGraphCopy(p, "_monotone")
	}
	c.copyInputs()
	previousPending := aig.Aig_ObjCreateCi(c.destination)
	delayed := make([]aig.Edge, len(vectors.attrConsequentCandidates))
	for i := range delayed {
		delayed[i] = aig.Aig_ObjCreateCi(c.destination)
	}
	c.copySourceNodes()
	pending := c.signal(p.Outputs()[indexes.attrPendingSignalIndex])
	guard := aig.Aig_And(c.destination, pending, previousPending)
	antecedents := []aig.Edge{}
	for _, id := range vectors.attrAntecedents {
		antecedents = append(antecedents, c.signal(p.Outputs()[id]))
	}
	antecedent := disjoin(c.destination, antecedents)
	invariant := aig.Edge(0)
	if indexes.attrSafetyInvarIndex >= 0 {
		invariant = c.signal(p.Outputs()[indexes.attrSafetyInvarIndex])
	}
	candidates, properties := make([]aig.Edge, len(delayed)), make([]aig.Edge, len(delayed))
	for i, id := range vectors.attrConsequentCandidates {
		candidate := aig.Aig_Or(c.destination, c.signal(p.Outputs()[id]), antecedent)
		candidates[i] = candidate
		property := aig.Aig_Or(c.destination, aig.Aig_And(c.destination, guard, delayed[i]).Not(), candidate)
		properties[i] = aig.Aig_And(c.destination, property, invariant)
	}
	for i := 0; i < propertyOutputs(p); i++ {
		aig.Aig_ObjCreateCo(c.destination, c.signal(p.Outputs()[i]))
	}
	if start != nil {
		*start = int32(propertyOutputs(p))
	}
	for _, e := range properties {
		aig.Aig_ObjCreateCo(c.destination, e)
	}
	c.appendConstraints()
	for r := 0; r < registers(p); r++ {
		aig.Aig_ObjCreateCo(c.destination, c.signal(p.Outputs()[trueOutputs(p)+r]))
	}
	aig.Aig_ObjCreateCo(c.destination, pending)
	for _, e := range candidates {
		aig.Aig_ObjCreateCo(c.destination, e)
	}
	aig.Aig_ManSetRegNum(c.destination, int32(registers(p)+1+len(delayed)))
	c.setConstraints()
	return c.destination
}
func findNewDisjunctiveWith(p *aig.Graph, indexes *aigPoIndices, vectors *antecedentConsequentVectorsStruct, options SafetyOptions, checker SafetyChecker) []int32 {
	if len(vectors.attrConsequentCandidates) == 0 {
		return nil
	}
	var start int32
	tester := negateOutputs(createDisjunctiveMonotoneTesterMode(p, indexes, vectors, &start, false), propertyOutputs(p))
	result := []int32{}
	for i, id := range vectors.attrConsequentCandidates {
		if checker(tester, int(start)+i, options).Status == 1 {
			result = append(result, id)
		}
	}
	return result
}
func findNewDisjunctiveMonotone(p *aig.Graph, indexes *aigPoIndices, vectors *antecedentConsequentVectorsStruct) []int32 {
	var start int32
	tester := createDisjunctiveMonotoneTester(p, indexes, vectors, &start)
	for i := 0; i < trueOutputs(tester); i++ {
		out := tester.Outputs()[i]
		aig.Aig_ObjPatchFanin0(tester, out, outputDriver(tester, i).Not())
	}
	opts := pdr.DefaultOptions()
	opts.SolveAll = true
	opts.Silent = true
	if _, err := pdr.Solve(tester, &opts); err != nil {
		panic(err)
	}
	models := tester.SequentialCounterexamples()
	var result []int32
	if models != nil {
		for i := 0; i < trueOutputs(tester); i++ {
			if vec.Vec_PtrEntry(models, i) == nil && i >= int(indexes.attrPendingSignalIndex)+1 {
				result = append(result, vectors.attrConsequentCandidates[i-int(indexes.attrPendingSignalIndex)-1])
			}
		}
	}
	return result
}
func vectorDifference(a, b []int32) []int32 {
	result := []int32{}
	for _, id := range a {
		if !slices.Contains(b, id) {
			result = append(result, id)
		}
	}
	return result
}
func createSingletonIntVector(id int32) []int32 { return []int32{id} }
func updateAnteConseVectors(vectors *antecedentConsequentVectorsStruct) []int32 {
	if len(vectors.attrAntecedents) == 0 {
		return vectors.attrConsequentCandidates
	}
	return vectorDifference(vectors.attrConsequentCandidates, vectors.attrAntecedents)
}
func Vec_IntPushUniqueLocal(values *[]int32, value int32) int32 {
	if slices.Contains(*values, value) {
		return 1
	}
	*values = append(*values, value)
	return 0
}
func nextDisjunctiveLevel(p *aig.Graph, indexes *aigPoIndices, vectors *antecedentConsequentVectorsStruct, previous [][]int32, options SafetyOptions, checker SafetyChecker) [][]int32 {
	union := []int32{}
	for _, group := range previous {
		for _, id := range group {
			Vec_IntPushUniqueLocal(&union, id)
		}
	}
	candidates := vectorDifference(vectors.attrConsequentCandidates, union)
	result := [][]int32{}
	for _, group := range previous {
		local := &antecedentConsequentVectorsStruct{attrAntecedents: group, attrConsequentCandidates: candidates}
		for _, id := range findNewDisjunctiveWith(p, indexes, local, options, checker) {
			result = append(result, append(append([]int32(nil), group...), id))
		}
	}
	return result
}
func findNextLevelDisjunctiveMonotone(p *aig.Graph, indexes *aigPoIndices, vectors *antecedentConsequentVectorsStruct, previous [][]int32) [][]int32 {
	union := []int32{}
	for _, group := range previous {
		for _, id := range group {
			Vec_IntPushUniqueLocal(&union, id)
		}
	}
	candidates := vectorDifference(vectors.attrConsequentCandidates, union)
	result := [][]int32{}
	for _, group := range previous {
		local := &antecedentConsequentVectorsStruct{attrAntecedents: group, attrConsequentCandidates: candidates}
		for _, id := range findNewDisjunctiveMonotone(p, indexes, local) {
			result = append(result, append(append([]int32(nil), group...), id))
		}
	}
	return result
}
func FindDisjunctiveMonotone(n *Network, options SafetyOptions, checker SafetyChecker) [][]int32 {
	if checker == nil {
		checker = CheckSafety
	}
	pending, hints := findPendingSignal(n), findHintOutputs(n)
	if pending < 0 || len(hints) == 0 {
		return nil
	}
	indexes := &aigPoIndices{pending, hints[0], hints[len(hints)-1], collectSafetyInvariantPOIndex(n)}
	vectors := &antecedentConsequentVectorsStruct{attrConsequentCandidates: hints}
	level1 := [][]int32{}
	for _, id := range findNewDisjunctiveWith(n.Graph, indexes, vectors, options, checker) {
		level1 = append(level1, []int32{id})
	}
	level2 := nextDisjunctiveLevel(n.Graph, indexes, vectors, level1, options, checker)
	return append(level1, level2...)
}
func findDisjunctiveMonotoneSignals(n *Network) [][]int32 {
	pending, hints := findPendingSignal(n), findHintOutputs(n)
	if pending < 0 || len(hints) == 0 {
		return nil
	}
	indexes := &aigPoIndices{pending, hints[0], hints[len(hints)-1], collectSafetyInvariantPOIndex(n)}
	vectors := &antecedentConsequentVectorsStruct{attrConsequentCandidates: hints}
	level1 := [][]int32{}
	for _, id := range findNewDisjunctiveMonotone(n.Graph, indexes, vectors) {
		level1 = append(level1, []int32{id})
	}
	return append(level1, findNextLevelDisjunctiveMonotone(n.Graph, indexes, vectors, level1)...)
}
func appendVecToMasterVecInt(master *[][]int32, candidates [][]int32) {
	for _, group := range candidates {
		*master = append(*master, append([]int32(nil), group...))
	}
}
func printGroups(groups [][]int32, n *Network, fileName, marker string) {
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	for i, group := range groups {
		if marker == "csLevel1Stabil" {
			fmt.Printf("INT[%d] : ( ", i)
		}
		fmt.Fprint(file, "( ")
		for j, id := range group {
			name := outputName(n, int(id))
			position := strings.Index(name, marker)
			if position < 0 {
				panic("live: group output has no expected name marker")
			}
			suffix := name[position:]
			fmt.Fprint(file, suffix)
			if marker == "csLevel1Stabil" {
				fmt.Print(suffix)
			}
			ending := " )\n"
			if j < len(group)-1 {
				ending = " || "
			}
			fmt.Fprint(file, ending)
			if marker == "csLevel1Stabil" {
				fmt.Print(ending)
			}
		}
	}
}
func printAllIntVectors(groups [][]int32, n *Network, fileName string) {
	printGroups(groups, n, fileName, "hint")
}
func printAllIntVectorsStabil(groups [][]int32, n *Network, fileName string) {
	printGroups(groups, n, fileName, "csLevel1Stabil")
}
