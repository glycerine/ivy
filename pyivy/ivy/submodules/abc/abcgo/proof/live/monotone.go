package live

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/proof/pdr"
	"slices"
	"strings"
)

type aigPoIndices struct{ attrPendingSignalIndex, attrHintSingalBeginningMarker, attrHintSingalEndMarker, attrSafetyInvarIndex int32 }

func allocAigPoIndices() *aigPoIndices { return &aigPoIndices{-1, -1, -1, -1} }

type monotoneVectorsStruct struct{ attrKnownMonotone, attrCandMonotone, attrHintMonotone []int32 }

func allocPointersToMonotoneVectors() *monotoneVectorsStruct { return new(monotoneVectorsStruct) }
func findHintOutputs(n *Network) []int32                     { return findHintOutputsMode(n, true) }
func findHintOutputsMode(n *Network, source bool) []int32 {
	var result []int32
	for i := 0; i < outputCount(n.Graph, source); i++ {
		if strings.Contains(outputName(n, i), "hint_") {
			result = append(result, int32(i))
		}
	}
	return result
}
func findPendingSignal(n *Network) int32 { return findOutput(n, "pendingSignal") }

// The C helper only prints its inputs and always returns success.
func checkSanityOfKnownMonotone(known, candidates, hints []int32) int32 {
	for _, values := range [][]int32{known, candidates, hints} {
		for _, value := range values {
			fmt.Printf("%d ", value)
		}
		fmt.Println()
	}
	return 1
}

func createMonotoneTester(p *aig.Graph, indexes *aigPoIndices, vectors *monotoneVectorsStruct, start *int32) *aig.Graph {
	return createMonotoneTesterMode(p, indexes, vectors, start, true)
}
func createMonotoneTesterMode(p *aig.Graph, indexes *aigPoIndices, vectors *monotoneVectorsStruct, start *int32, source bool) *aig.Graph {
	if indexes.attrPendingSignalIndex < 0 {
		panic("live: pending signal missing")
	}
	if source && vectors.attrKnownMonotone != nil {
		checkSanityOfKnownMonotone(vectors.attrKnownMonotone, vectors.attrCandMonotone, vectors.attrHintMonotone)
	}
	c := newGraphCopy(p, "_monotone")
	if source {
		c = newSourceGraphCopy(p, "_monotone")
	}
	c.copyInputs()
	pendingOld := aig.Aig_ObjCreateCi(c.destination)
	delayed := make(map[int32]aig.Edge)
	delayedInOrder := make([]aig.Edge, 0, len(vectors.attrHintMonotone))
	for _, id := range vectors.attrHintMonotone {
		d := aig.Aig_ObjCreateCi(c.destination)
		delayed[id] = d
		delayedInOrder = append(delayedInOrder, d)
	}
	c.copySourceNodes()
	pending := c.signal(p.Outputs()[indexes.attrPendingSignalIndex])
	guard := aig.Aig_And(c.destination, pending, pendingOld)
	for _, id := range vectors.attrKnownMonotone {
		driverID, previous := id, delayed[id]
		if source {
			driverID = id - indexes.attrHintSingalBeginningMarker + 1 + indexes.attrPendingSignalIndex
			previous = delayedInOrder[int(driverID)-c.outputCount()]
		}
		guard = aig.Aig_And(c.destination, guard, aig.Aig_Exor(c.destination, c.signal(p.Outputs()[driverID]), previous).Not())
	}
	properties := []aig.Edge{}
	for _, id := range vectors.attrHintMonotone {
		if slices.Contains(vectors.attrKnownMonotone, id) {
			continue
		}
		property := aig.Aig_Or(c.destination, aig.Aig_And(c.destination, guard, delayed[id]).Not(), c.signal(p.Outputs()[id]))
		properties = append(properties, property)
	}
	for i := 0; i < c.outputCount(); i++ {
		aig.Aig_ObjCreateCo(c.destination, c.signal(p.Outputs()[i]))
	}
	if start != nil {
		*start = int32(c.outputCount())
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
	tester := negateOutputs(createMonotoneTesterMode(p, indexes, vectors, &start, false), propertyOutputs(p))
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
	var start int32
	tester := createMonotoneTester(p, indexes, vectors, &start)
	var proved []int32
	for i := int(start); i < trueOutputs(tester); i++ {
		output := tester.Outputs()[i]
		driver := outputDriver(tester, i)
		aig.Aig_ObjPatchFanin0(tester, output, driver.Not())
		result, err := pdr.Solve(tester, nil)
		aig.Aig_ObjPatchFanin0(tester, output, driver)
		if err != nil {
			panic(err)
		}
		if result.Status == pdr.Proved {
			proved = append(proved, int32(i)-(indexes.attrPendingSignalIndex+1)+indexes.attrHintSingalBeginningMarker)
		}
	}
	return proved
}
func findRemainingMonotoneCandidates(known, hints []int32) []int32 {
	if len(known) == 0 {
		return hints
	}
	return vectorDifference(hints, known)
}
func FindMonotone(n *Network, options SafetyOptions, checker SafetyChecker) []int32 {
	if checker == nil {
		checker = CheckSafety
	}
	pending, hints := findOutputMode(n, "pendingSignal", false), findHintOutputsMode(n, false)
	if pending < 0 || len(hints) == 0 {
		return nil
	}
	indexes := &aigPoIndices{pending, hints[0], hints[len(hints)-1], -1}
	vectors := &monotoneVectorsStruct{attrCandMonotone: hints, attrHintMonotone: hints}
	return findNewMonotoneWith(n.Graph, indexes, vectors, options, checker)
}

// findMonotoneSignals follows the C entry point, which discards its results.
func findMonotoneSignals(n *Network) []int32 {
	pending, hints := findPendingSignal(n), findHintOutputs(n)
	if pending < 0 || len(hints) == 0 {
		return nil
	}
	indexes := &aigPoIndices{pending, hints[0], hints[len(hints)-1], -1}
	vectors := &monotoneVectorsStruct{attrCandMonotone: hints, attrHintMonotone: hints}
	findNewMonotone(n.Graph, indexes, vectors)
	return nil
}
