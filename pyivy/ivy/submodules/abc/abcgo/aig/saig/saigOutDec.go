package saig

import (
	"fmt"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
)

// Saig_ManFindPrimes finds one- and, optionally, two-literal cubes that imply
// the property output. Each returned Vec_Int_t contains packed AIG literals
// (2*objectID + complement), in candidate traversal order.
func Saig_ManFindPrimes(graph *aig.Graph, nLits, verbose int32) *vec.Vec_Ptr_t {
	if graph == nil || (nLits != 1 && nLits != 2) || len(graph.Outputs()) == 0 {
		return nil
	}
	outputs := graph.Outputs()
	propertyOutput := outputs[0]
	root := aig.Aig_ObjChild0(graph, propertyOutput)
	candidates := outputConeNodes(graph, root)
	cnf, err := encodeAIG(graph)
	if err != nil {
		return nil
	}
	primes := vec.Vec_PtrAlloc(16)
	marked := make([]bool, len(candidates))
	propertyFalse := bsat.Negative(cnf.vars[propertyOutput.NodeID()])

	for i, id := range candidates {
		node := graph.Node(aig.Edge(id) << 1)
		lit := bsat.Positive(cnf.vars[id])
		if node.Phase {
			lit ^= 1
		}
		if cnf.solver.Solve(propertyFalse, lit) == bsat.Unsatisfiable {
			cube := vec.Vec_IntAlloc(1)
			vec.Vec_IntPush(cube, packedLiteral(id, node.Phase))
			vec.Vec_PtrPush(primes, cube)
			marked[i] = true
			if verbose != 0 {
				fmt.Printf("Adding prime %d%c\n", id, phaseChar(node.Phase))
			}
		}
	}
	oneLiteralCount := vec.Vec_PtrSize(primes)
	if nLits == 2 {
		for i, id0 := range candidates {
			if marked[i] {
				continue
			}
			phase0 := graph.Node(aig.Edge(id0) << 1).Phase
			for j := i + 1; j < len(candidates); j++ {
				if marked[j] {
					continue
				}
				id1 := candidates[j]
				phase1 := graph.Node(aig.Edge(id1) << 1).Phase
				found := false
				// The source checks the three combinations other than the
				// phase-aligned pair, in this exact order.
				for mask := int32(0); mask < 3; mask++ {
					compl0 := phase0 != (mask&1 != 0)
					compl1 := phase1 != (mask&2 != 0)
					lit0 := bsat.Positive(cnf.vars[id0])
					lit1 := bsat.Positive(cnf.vars[id1])
					if compl0 {
						lit0 ^= 1
					}
					if compl1 {
						lit1 ^= 1
					}
					if cnf.solver.Solve(propertyFalse, lit0, lit1) != bsat.Unsatisfiable {
						continue
					}
					cube := vec.Vec_IntAlloc(2)
					vec.Vec_IntPush(cube, packedLiteral(id0, compl0))
					vec.Vec_IntPush(cube, packedLiteral(id1, compl1))
					vec.Vec_PtrPush(primes, cube)
					if verbose != 0 {
						fmt.Printf("Adding prime %d%c %d%c\n", id0, phaseChar(compl0), id1, phaseChar(compl1))
					}
					found = true
					break
				}
				if found {
					break
				}
			}
		}
	}
	if verbose != 0 {
		fmt.Printf("Property cone size = %6d    1-lit primes = %5d    2-lit primes = %5d\n", len(candidates), oneLiteralCount, vec.Vec_PtrSize(primes)-oneLiteralCount)
	}
	return primes
}

// Saig_ManDecPropertyOutput copies the sequential graph, then adds one output
// for every prime cube discovered in the original property cone.
func Saig_ManDecPropertyOutput(graph *aig.Graph, nLits, verbose int32) *aig.Graph {
	if graph == nil {
		return nil
	}
	primes := Saig_ManFindPrimes(graph, nLits, verbose)
	dst := aig.NewGraph()
	dst.Name, dst.Spec = graph.Name, graph.Spec
	nodes := graph.Nodes()
	mapped := make([]aig.Edge, len(nodes))
	known := make([]bool, len(nodes))
	mapped[0] = dst.Const1()
	known[0] = true
	for _, input := range graph.Inputs() {
		mapped[input.NodeID()] = dst.NewInput()
		known[input.NodeID()] = true
	}
	var copyEdge func(aig.Edge) aig.Edge
	copyEdge = func(edge aig.Edge) aig.Edge {
		id := edge.NodeID()
		if id < 0 || id >= len(nodes) || !graph.ValidEdge(edge) {
			panic("saig: invalid edge while decomposing property output")
		}
		if !known[id] {
			node := nodes[id]
			switch node.Kind {
			case aig.And:
				mapped[id] = dst.AND(copyEdge(node.Fanins[0]), copyEdge(node.Fanins[1]))
			case aig.ExclusiveOr:
				mapped[id] = dst.NewXor(copyEdge(node.Fanins[0]), copyEdge(node.Fanins[1]))
			case aig.Buffer, aig.Output:
				mapped[id] = copyEdge(node.Fanins[0])
			case aig.Input:
				panic("saig: input missing from graph input list")
			default:
				panic("saig: unsupported node in property decomposition")
			}
			known[id] = true
		}
		result := mapped[id]
		if edge.Complemented() {
			result = result.Not()
		}
		return result
	}
	outputs := graph.Outputs()
	regCount := int(aig.Aig_ManRegNum(graph))
	poCount := len(outputs) - regCount
	for _, output := range outputs[:poCount] {
		dst.AddOutput(copyEdge(aig.Aig_ObjChild0(graph, output)))
	}
	for k := 0; k < vec.Vec_PtrSize(primes); k++ {
		cube := vec.Vec_PtrEntry(primes, k).(*vec.Vec_Int_t)
		term := dst.Const1()
		for i := 0; i < vec.Vec_IntSize(cube); i++ {
			literal := vec.Vec_IntEntry(cube, i)
			edge := copyEdge(aig.Edge(uint32(literal>>1)) << 1)
			if literal&1 != 0 {
				edge = edge.Not()
			}
			term = dst.AND(term, edge)
		}
		dst.AddOutput(term)
	}
	for _, output := range outputs[poCount:] {
		dst.AddOutput(copyEdge(aig.Aig_ObjChild0(graph, output)))
	}
	aig.Aig_ManSetRegNum(dst, int32(regCount))
	aig.Aig_ManSetConstrNum(dst, aig.Aig_ManConstrNum(graph))
	return dst
}

func outputConeNodes(graph *aig.Graph, root aig.Edge) []int {
	seen := make([]bool, len(graph.Nodes()))
	var order []int
	var visit func(aig.Edge)
	visit = func(edge aig.Edge) {
		id := edge.NodeID()
		if id < 0 || id >= len(seen) || seen[id] || !graph.ValidEdge(edge) {
			return
		}
		seen[id] = true
		node := graph.Node(edge)
		for _, fanin := range node.Fanins {
			if fanin != aig.InvalidEdge {
				visit(fanin)
			}
		}
		if node.Kind == aig.And || node.Kind == aig.ExclusiveOr {
			order = append(order, id)
		}
	}
	visit(root)
	if len(order) > 0 && order[len(order)-1] == root.NodeID() {
		order = order[:len(order)-1]
	}
	return order
}

func packedLiteral(object int, complemented bool) int32 {
	literal := int32(object << 1)
	if complemented {
		literal++
	}
	return literal
}

func phaseChar(complemented bool) byte {
	if complemented {
		return '-'
	}
	return '+'
}
