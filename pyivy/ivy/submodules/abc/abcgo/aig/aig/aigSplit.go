package aig

import (
	"fmt"
	"math/rand/v2"
	"time"
)

// These source-named BDD helpers use an exact Shannon-reduced decision graph
// represented with Graph muxes, replacing CUDD pointers and reference counting.
func Aig_NodeBddToMuxes_rec(decisions *Graph, function Edge, destination *Graph, mapping map[int]Edge) Edge {
	id := function.NodeID()
	value, ok := mapping[id]
	if !ok {
		n := *decisions.node(function)
		if n.Kind != And && n.Kind != ExclusiveOr && n.Kind != Buffer && n.Kind != Output {
			panic("aig: unmapped decision variable")
		}
		a := Aig_NodeBddToMuxes_rec(decisions, n.Fanins[0], destination, mapping)
		if n.Kind == Buffer || n.Kind == Output {
			value = a
		} else {
			b := Aig_NodeBddToMuxes_rec(decisions, n.Fanins[1], destination, mapping)
			if n.Kind == And {
				value = destination.AND(a, b)
			} else {
				value = destination.NewXor(a, b)
			}
		}
		mapping[id] = value
	}
	if function.Complemented() {
		value = value.Not()
	}
	return value
}
func splitCofactor(p *Graph, root Edge, assignment map[int]bool) Edge {
	memo := map[int]Edge{0: p.Const1()}
	for id, value := range assignment {
		memo[id] = p.Const0()
		if value {
			memo[id] = p.Const1()
		}
	}
	var visit func(Edge) Edge
	visit = func(e Edge) Edge {
		id := e.NodeID()
		value, ok := memo[id]
		if !ok {
			n := p.nodes[id]
			switch n.Kind {
			case Input:
				value = Edge(id) << 1
			case Output, Buffer:
				value = visit(n.Fanins[0])
			case And:
				value = p.AND(visit(n.Fanins[0]), visit(n.Fanins[1]))
			case ExclusiveOr:
				value = p.NewXor(visit(n.Fanins[0]), visit(n.Fanins[1]))
			default:
				panic("aig: invalid cofactor node")
			}
			memo[id] = value
		}
		if e.Complemented() {
			value = value.Not()
		}
		return value
	}
	return visit(root)
}
func splitReduce(p *Graph, root Edge, memo map[Edge]Edge) Edge {
	regular := Aig_Regular(root)
	if result, ok := memo[regular]; ok {
		if root.Complemented() {
			return result.Not()
		}
		return result
	}
	support := Aig_Support(p, regular)
	value := p.Const1()
	if len(support) == 0 {
		if !Aig_ObjPhaseReal(p, regular) {
			value = p.Const0()
		}
	} else {
		positions := make(map[int]int, len(p.inputs))
		for i, e := range p.inputs {
			positions[e.NodeID()] = i
		}
		variable := support[0]
		for _, e := range support[1:] {
			if positions[e.NodeID()] < positions[variable.NodeID()] {
				variable = e
			}
		}
		low := splitReduce(p, splitCofactor(p, regular, map[int]bool{variable.NodeID(): false}), memo)
		high := splitReduce(p, splitCofactor(p, regular, map[int]bool{variable.NodeID(): true}), memo)
		if low == high {
			value = low
		} else if high.Complemented() {
			value = p.Mux(variable, high.Not(), low.Not()).Not()
		} else {
			value = p.Mux(variable, high, low)
		}
	}
	memo[regular] = value
	if root.Complemented() {
		value = value.Not()
	}
	return value
}
func Aig_ManBuildPoBdd_rec(p *Graph, e Edge, decisions *Graph) Edge {
	if p.copies == nil {
		panic("aig: decision variable mapping not initialized")
	}
	return Aig_ManDup_rec(decisions, p, e)
}
func Aig_ManBuildPoBdd(p *Graph, function *Edge) *Graph {
	if len(p.outputs)-int(p.nRegs) != 1 || function == nil {
		panic("aig: splitting requires one property output")
	}
	decisions := Aig_ManStartFrom(p)
	root := Aig_ManBuildPoBdd_rec(p, p.outputDriver(p.outputs[0]), decisions)
	*function = splitReduce(decisions, root, make(map[Edge]Edge))
	return decisions
}
func Aig_ManCofactorBdds(p *Graph, subset []Edge, decisions *Graph, function Edge) []Edge {
	if len(subset) > 16 {
		panic("aig: at most 16 cofactoring variables")
	}
	seen := make(map[int]bool)
	for _, e := range subset {
		if e.Complemented() || decisions.node(e).Kind != Input || seen[e.NodeID()] {
			panic("aig: invalid cofactoring variables")
		}
		seen[e.NodeID()] = true
	}
	results := make([]Edge, 0, 1<<uint(len(subset)))
	memo := make(map[Edge]Edge)
	for code := 0; code < 1<<uint(len(subset)); code++ {
		assignment := make(map[int]bool, len(subset))
		cube := decisions.Const1()
		for i, e := range subset {
			value := code&(1<<uint(len(subset)-1-i)) != 0
			assignment[e.NodeID()] = value
			literal := e
			if !value {
				literal = literal.Not()
			}
			cube = decisions.AND(cube, literal)
		}
		cofactor := splitReduce(decisions, splitCofactor(decisions, function, assignment), memo)
		results = append(results, decisions.AND(cofactor, cube))
	}
	return results
}
func Aig_ManConvertBddsToAigs(p, decisions *Graph, cofactors []Edge) *Graph {
	q := Aig_ManStartFrom(p)
	mapping := map[int]Edge{0: 0}
	if len(decisions.inputs) != len(q.inputs) {
		panic("aig: inconsistent decision variable count")
	}
	for i, e := range decisions.inputs {
		mapping[e.NodeID()] = q.inputs[i]
	}
	for _, function := range cofactors {
		if function == decisions.Const0() {
			continue
		}
		q.AddOutput(Aig_NodeBddToMuxes_rec(decisions, function, q, mapping))
	}
	for _, output := range p.outputs[1:] {
		q.AddOutput(Aig_ManDup_rec(q, p, p.outputDriver(output)))
	}
	Aig_ManCleanup(q)
	Aig_ManSetRegNum(q, p.nRegs)
	return q
}
func Aig_ManVecRandSubset(values []Edge, count int32) []Edge {
	if count < 0 {
		panic("aig: negative subset size")
	}
	result := append([]Edge(nil), values...)
	for len(result) > int(count) {
		i := rand.IntN(len(result))
		result = append(result[:i], result[i+1:]...)
	}
	return result
}
func Aig_ManSplit(p *Graph, variables, verbose int32) *Graph {
	start := time.Now()
	if len(p.outputs)-int(p.nRegs) != 1 {
		fmt.Println("Currently works only for one primary output.")
		return nil
	}
	if variables < 1 || variables > 16 {
		fmt.Println("The number of cofactoring variables must be between 1 and 16.")
		return nil
	}
	support := Aig_Support(p, Aig_Regular(p.outputDriver(p.outputs[0])))
	if len(support) == 0 {
		fmt.Println("Property output function is a constant.")
		return nil
	}
	var function Edge
	decisions := Aig_ManBuildPoBdd(p, &function)
	subset := Aig_ManVecRandSubset(support, variables)
	for i, e := range subset {
		index := Aig_ObjCioId(p, e)
		subset[i] = decisions.inputs[index]
	}
	cofactors := Aig_ManCofactorBdds(p, subset, decisions, function)
	result := Aig_ManConvertBddsToAigs(p, decisions, cofactors)
	if verbose != 0 {
		fmt.Printf("Support =%5d. Created %d cofactors (out of %d). Time=%s\n", len(support), len(result.outputs)-int(result.nRegs), len(cofactors), time.Since(start))
	}
	return result
}
