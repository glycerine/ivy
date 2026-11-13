package pdr

// Register equivalence utilities ported from pdrClass.c, Alan Mishchenko (2010).

import (
	"fmt"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
)

// Pdr_ManRehashWithMap duplicates a sequential AIG while replacing each
// register output with zero (-1) or the specified register output.
func Pdr_ManRehashWithMap(graph *aig.Graph, mapping *vec.Vec_Int_t) *aig.Graph {
	if graph == nil || mapping == nil || vec.Vec_IntSize(mapping) != int(graph.RegisterCount()) {
		panic("pdr: invalid register map")
	}
	result := aig.NewGraph()
	result.Name, result.Spec = graph.Name, graph.Spec
	inputs := graph.Inputs()
	copies := make([]aig.Edge, len(graph.Nodes()))
	done := make([]bool, len(copies))
	done[0] = true
	for _, input := range inputs {
		copies[input.NodeID()] = result.NewInput()
		done[input.NodeID()] = true
	}
	firstReg := len(inputs) - int(graph.RegisterCount())
	for i, representative := range vec.Vec_IntArray(mapping) {
		if representative < -1 || int(representative) >= int(graph.RegisterCount()) {
			panic("pdr: register representative out of range")
		}
		id := inputs[firstReg+i].NodeID()
		if representative == -1 {
			copies[id] = result.Const0()
		} else {
			copies[id] = copies[inputs[firstReg+int(representative)].NodeID()]
		}
	}
	var copyEdge func(aig.Edge) aig.Edge
	copyEdge = func(edge aig.Edge) aig.Edge {
		id := edge.NodeID()
		if !done[id] {
			node := graph.Node(edge)
			switch node.Kind {
			case aig.And:
				copies[id] = result.AND(copyEdge(node.Fanins[0]), copyEdge(node.Fanins[1]))
			case aig.ExclusiveOr:
				copies[id] = result.NewXor(copyEdge(node.Fanins[0]), copyEdge(node.Fanins[1]))
			case aig.Buffer, aig.Output:
				copies[id] = copyEdge(node.Fanins[0])
			default:
				panic("pdr: unexpected AIG object in register map")
			}
			done[id] = true
		}
		mapped := copies[id]
		if edge.Complemented() {
			mapped = mapped.Not()
		}
		return mapped
	}
	for _, output := range graph.Outputs() {
		result.AddOutput(copyEdge(graph.Node(output).Fanins[0]))
	}
	aig.Aig_ManCleanup(result)
	aig.Aig_ManSetRegNum(result, graph.RegisterCount())
	return result
}

// Pdr_ManCreateMap groups register inputs with identical drivers, including
// their polarities. A constant-zero driver is represented by -1.
func Pdr_ManCreateMap(graph *aig.Graph) *vec.Vec_Int_t {
	if graph == nil {
		panic("pdr: nil graph")
	}
	result := vec.Vec_IntAlloc(int(graph.RegisterCount()))
	first := make(map[aig.Edge]int32)
	for i := int32(0); i < graph.RegisterCount(); i++ {
		driver := graph.Node(aig.Aig_ManLi(graph, i)).Fanins[0]
		if driver == graph.Const0() {
			vec.Vec_IntPush(result, -1)
			continue
		}
		id, found := first[driver]
		if !found {
			id = i
			first[driver] = i
		}
		vec.Vec_IntPush(result, id)
	}
	return result
}

func Pdr_ManCountMap(mapping *vec.Vec_Int_t) int32 {
	var count int32
	for i, representative := range vec.Vec_IntArray(mapping) {
		if representative != int32(i) {
			count++
		}
	}
	return count
}

func Pdr_ManPrintMap(mapping *vec.Vec_Int_t) {
	entries := vec.Vec_IntArray(mapping)
	fmt.Print("    Consts: ")
	for i, representative := range entries {
		if representative == -1 {
			fmt.Printf("%d ", i)
		}
	}
	fmt.Println()
	printed := make(map[int32]bool)
	for i, representative := range entries {
		if representative < 0 || representative == int32(i) || printed[representative] {
			continue
		}
		printed[representative] = true
		fmt.Printf("    Class %d : ", representative)
		for j, candidate := range entries {
			if candidate == representative {
				fmt.Printf("%d ", j)
			}
		}
		fmt.Println()
	}
}

func Pdr_ManEquivClasses(graph *aig.Graph) {
	if graph == nil || graph.RegisterCount() == 0 {
		panic("pdr: expected registers")
	}
	mapping := vec.Vec_IntAlloc(int(graph.RegisterCount()))
	vec.Vec_IntFill(mapping, int(graph.RegisterCount()), -1)
	for frame := 0; frame < 100; frame++ {
		reduced := Pdr_ManRehashWithMap(graph, mapping)
		property := "unknown"
		if reduced.OutputCount() > 0 && reduced.Node(reduced.Outputs()[0]).Fanins[0] == reduced.Const0() {
			property = "proof"
		}
		fmt.Printf("F =%4d : Total = %6d. Nodes = %6d. RedRegs = %6d. Prop = %s\n", frame+1, aig.Aig_ManNodeNum(graph), aig.Aig_ManNodeNum(reduced), Pdr_ManCountMap(mapping), property)
		Pdr_ManPrintMap(mapping)
		mapping = Pdr_ManCreateMap(reduced)
		if Pdr_ManCountMap(mapping) == 0 {
			break
		}
	}
}
