package aig

import (
	"fmt"
	"time"
)

func interValidate(on, off *Graph) {
	if on == nil || off == nil || len(on.inputs) != len(off.inputs) {
		panic("aig: interpolation requires equal input counts")
	}
	if err := on.Check(); err != nil {
		panic(err)
	}
	if err := off.Check(); err != nil {
		panic(err)
	}
}
func interCopy(source, destination *Graph, e Edge, inputs []Edge) Edge {
	if len(source.inputs) != len(inputs) {
		panic("aig: inconsistent shared-variable map")
	}
	mapping := map[int]Edge{0: destination.Const1()}
	for i, input := range source.inputs {
		mapping[input.NodeID()] = inputs[i]
	}
	return Aig_NodeBddToMuxes_rec(source, e, destination, mapping)
}
func interConjunction(source, destination *Graph, outputs, inputs []Edge) Edge {
	result := destination.Const1()
	mapping := map[int]Edge{0: destination.Const1()}
	for i, input := range source.inputs {
		mapping[input.NodeID()] = inputs[i]
	}
	for _, output := range outputs {
		result = destination.AND(result, Aig_NodeBddToMuxes_rec(source, source.outputDriver(output), destination, mapping))
	}
	return result
}

// Aig_ManInterFast checks each corresponding output pair for disjointness under
// shared CIs. Results replace the C routine's diagnostic-only status; source
// managers and register metadata are never borrowed or reset for CNF generation.
func Aig_ManInterFast(on, off *Graph, verbose int32) []bool {
	interValidate(on, off)
	if len(on.outputs) != len(off.outputs) {
		panic("aig: interpolation output counts differ")
	}
	work := NewGraph()
	for range on.inputs {
		work.NewInput()
	}
	results := make([]bool, len(on.outputs))
	memo := make(map[Edge]Edge)
	for i := range results {
		a := interCopy(on, work, on.outputDriver(on.outputs[i]), work.inputs)
		b := interCopy(off, work, off.outputDriver(off.outputs[i]), work.inputs)
		results[i] = splitReduce(work, work.AND(a, b), memo) == work.Const0()
		if !results[i] {
			fmt.Println("The incremental SAT problem is not UNSAT.")
		}
	}
	return results
}

// Aig_ManInter constructs an exact Craig interpolant over the common inputs.
// It uses Shannon reduction rather than a SAT-resolution proof: since all CIs
// are shared, existential projection of A gives the strongest interpolant.
//
// In relation mode the source moves A's LAST output unit clause to B and makes
// that unit's regular node variable the FIRST common input. The projection is
// (extra == regular(lastOutput)) AND the other asserted A outputs. B still
// asserts the original last output (including its complement) and all B outputs.
func Aig_ManInter(on, off *Graph, relation, verbose int32) *Graph {
	start := time.Now()
	interValidate(on, off)
	if relation != 0 && len(on.outputs) == 0 {
		panic("aig: relation interpolation requires an A output")
	}
	work := NewGraph()
	for range on.inputs {
		work.NewInput()
	}
	a := interConjunction(on, work, on.outputs, work.inputs)
	b := interConjunction(off, work, off.outputs, work.inputs)
	memo := make(map[Edge]Edge)
	if splitReduce(work, work.AND(a, b), memo) != work.Const0() {
		fmt.Println("The SAT problem is not unsat.")
		return nil
	}
	result := NewGraph()
	result.Name = on.Name
	extra := InvalidEdge
	if relation != 0 {
		extra = result.NewInput()
	}
	shared := make([]Edge, len(on.inputs))
	for i := range shared {
		shared[i] = result.NewInput()
	}
	var interpolant Edge
	if relation == 0 {
		canonical := splitReduce(work, a, memo)
		interpolant = interCopy(work, result, canonical, shared)
	} else {
		prefix := interConjunction(on, result, on.outputs[:len(on.outputs)-1], shared)
		regularLast := Aig_Regular(on.outputDriver(on.outputs[len(on.outputs)-1]))
		function := interCopy(on, result, regularLast, shared)
		interpolant = result.AND(prefix, result.XOR(extra, function).Not())
	}
	result.AddOutput(interpolant)
	Aig_ManCleanup(result)
	if verbose != 0 {
		fmt.Printf("Exact graph interpolation: inputs=%d nodes=%d time=%s\n", len(result.inputs), Aig_ManNodeNum(result), time.Since(start))
	}
	return result
}
