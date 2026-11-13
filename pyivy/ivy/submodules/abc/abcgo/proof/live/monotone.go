package live

import (
	"github.com/glycerine/abcgo/aig/aig"
	"slices"
	"strings"
)

type aigPoIndices struct{ attrPendingSignalIndex, attrHintSingalBeginningMarker, attrHintSingalEndMarker, attrSafetyInvarIndex int32 }

func allocAigPoIndices() *aigPoIndices { return &aigPoIndices{-1, -1, -1, -1} }

type monotoneVectorsStruct struct{ attrKnownMonotone, attrCandMonotone, attrHintMonotone []int32 }

func allocPointersToMonotoneVectors() *monotoneVectorsStruct { return new(monotoneVectorsStruct) }
func findHintOutputs(n *Network) []int32 {
	result := []int32{}
	for i := 0; i < propertyOutputs(n.Graph); i++ {
		if strings.Contains(outputName(n, i), "hint_") {
			result = append(result, int32(i))
		}
	}
	return result
}
func findPendingSignal(n *Network) int32 { return findOutput(n, "pendingSignal") }
func checkSanityOfKnownMonotone(known, candidates, hints []int32) int32 {
	for _, id := range known {
		if !slices.Contains(hints, id) {
			return 0
		}
	}
	for _, id := range candidates {
		if !slices.Contains(hints, id) {
			return 0
		}
	}
	return 1
}
func createMonotoneTester(p *aig.Graph, indexes *aigPoIndices, vectors *monotoneVectorsStruct, start *int32) *aig.Graph {
	if indexes.attrPendingSignalIndex < 0 {
		panic("live: pending signal missing")
	}
	if checkSanityOfKnownMonotone(vectors.attrKnownMonotone, vectors.attrCandMonotone, vectors.attrHintMonotone) == 0 {
		panic("live: inconsistent monotone candidate vectors")
	}
	c := newGraphCopy(p, "_monotone")
	c.copyInputs()
	pendingOld := aig.Aig_ObjCreateCi(c.destination)
	delayed := make(map[int32]aig.Edge)
	for _, id := range vectors.attrHintMonotone {
		delayed[id] = aig.Aig_ObjCreateCi(c.destination)
	}
	pending := c.signal(p.Outputs()[indexes.attrPendingSignalIndex])
	guard := aig.Aig_And(c.destination, pending, pendingOld)
	for _, id := range vectors.attrKnownMonotone {
		guard = aig.Aig_And(c.destination, guard, aig.Aig_Exor(c.destination, c.signal(p.Outputs()[id]), delayed[id]).Not())
	}
	properties := []aig.Edge{}
	for _, id := range vectors.attrHintMonotone {
		if slices.Contains(vectors.attrKnownMonotone, id) {
			continue
		}
		property := aig.Aig_Or(c.destination, aig.Aig_And(c.destination, guard, delayed[id]).Not(), c.signal(p.Outputs()[id]))
		properties = append(properties, property)
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
	for _, id := range vectors.attrHintMonotone {
		aig.Aig_ObjCreateCo(c.destination, c.signal(p.Outputs()[id]))
	}
	aig.Aig_ManSetRegNum(c.destination, int32(registers(p)+1+len(vectors.attrHintMonotone)))
	c.setConstraints()
	return c.destination
}
func negateOutputs(p *aig.Graph, first int) *aig.Graph {
	c := newGraphCopy(p, "_check")
	c.copyInputs()
	for i := 0; i < propertyOutputs(p); i++ {
		driver := c.signal(p.Outputs()[i])
		if i >= first {
			driver = driver.Not()
		}
		aig.Aig_ObjCreateCo(c.destination, driver)
	}
	c.appendConstraints()
	for r := 0; r < registers(p); r++ {
		aig.Aig_ObjCreateCo(c.destination, c.signal(p.Outputs()[trueOutputs(p)+r]))
	}
	aig.Aig_ManSetRegNum(c.destination, int32(registers(p)))
	c.setConstraints()
	return c.destination
}
func findNewMonotoneWith(p *aig.Graph, indexes *aigPoIndices, vectors *monotoneVectorsStruct, options SafetyOptions, checker SafetyChecker) []int32 {
	var start int32
	tester := negateOutputs(createMonotoneTester(p, indexes, vectors, &start), propertyOutputs(p))
	result := []int32{}
	serial := int(start)
	for _, id := range vectors.attrHintMonotone {
		if slices.Contains(vectors.attrKnownMonotone, id) {
			continue
		}
		if checker(tester, serial, options).Status == 1 {
			result = append(result, id)
		}
		serial++
	}
	return result
}
func findNewMonotone(p *aig.Graph, indexes *aigPoIndices, vectors *monotoneVectorsStruct) []int32 {
	return findNewMonotoneWith(p, indexes, vectors, SafetyOptions{}, CheckSafety)
}
func findRemainingMonotoneCandidates(known, hints []int32) []int32 {
	return vectorDifference(hints, known)
}
func FindMonotone(n *Network, options SafetyOptions, checker SafetyChecker) []int32 {
	if checker == nil {
		checker = CheckSafety
	}
	pending, hints := findPendingSignal(n), findHintOutputs(n)
	if pending < 0 || len(hints) == 0 {
		return nil
	}
	indexes := &aigPoIndices{pending, hints[0], hints[len(hints)-1], -1}
	vectors := &monotoneVectorsStruct{attrCandMonotone: hints, attrHintMonotone: hints}
	return findNewMonotoneWith(n.Graph, indexes, vectors, options, checker)
}
func findMonotoneSignals(n *Network) []int32 { return FindMonotone(n, SafetyOptions{}, CheckSafety) }
