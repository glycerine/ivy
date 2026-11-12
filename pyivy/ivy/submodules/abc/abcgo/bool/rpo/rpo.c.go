package rpo

// Recursive partial read-once factorization from ABC rpo.c. Candidate groups
// are traversed by minimum endpoint degree, then minimum endpoint index. Each
// recursive branch restores the cluster slice before trying the next edge.
import "fmt"

func Rpo_CheckANDGroup(a, b *Literal_t, vars int32) bool {
	for i := 0; i < rpoWordNum(vars); i++ {
		if a.Transition[i]&^b.Function[i] != 0 || b.Transition[i]&^a.Function[i] != 0 {
			return false
		}
	}
	return true
}
func Rpo_CheckORGroup(a, b *Literal_t, vars int32) bool {
	for i := 0; i < rpoWordNum(vars); i++ {
		if a.Transition[i]&b.Function[i] != 0 || b.Transition[i]&a.Function[i] != 0 {
			return false
		}
	}
	return true
}
func Rpo_CreateEdge(op Operator_t, i, j int32, degree []int32) *Rpo_LCI_Edge_t {
	degree[i]++
	degree[j]++
	return &Rpo_LCI_Edge_t{ConnectionType: op, Idx1: i, Idx2: j}
}
func Rpo_computeMinEdgeCost(edges []*Rpo_LCI_Edge_t, count int32, degree []int32) int32 {
	best := int32(-1)
	var bestCost, bestVertex int32
	for i, edge := range edges[:count] {
		if edge.Visited {
			continue
		}
		cost := degree[edge.Idx1] + degree[edge.Idx2]
		vertex := min(edge.Idx1, edge.Idx2)
		if best < 0 || cost < bestCost || cost == bestCost && vertex < bestVertex {
			best = int32(i)
			bestCost = cost
			bestVertex = vertex
		}
	}
	return best
}
func Rpo_PrintEdge(edge *Rpo_LCI_Edge_t) {
	fmt.Printf("Edge (%d,%d)/%d\n", edge.Idx1, edge.Idx2, edge.ConnectionType)
}
func Rpo_Factorize(target []uint32, vars, threshold, verbose int32) *Literal_t {
	if rpoTruthConstant(target, vars, 0) {
		return Lit_CreateLiteralConst(target, vars, 0)
	}
	if rpoTruthConstant(target, vars, ^uint32(0)) {
		return Lit_CreateLiteralConst(target, vars, 1)
	}
	if verbose != 0 {
		fmt.Print("Target: ")
		Lit_PrintTT(target, vars)
		fmt.Println()
	}
	literals := make([]*Literal_t, 0, 2*vars)
	for index := vars - 1; index >= 0; index-- {
		for _, polarity := range []byte{'+', '-'} {
			if literal := Lit_Alloc(target, vars, index, polarity); literal != nil {
				literals = append(literals, literal)
			}
		}
	}
	if verbose != 0 {
		fmt.Printf("Allocated %d literal clusters\n", len(literals))
	}
	count := int32(0)
	return Rpo_Recursion(target, literals, int32(len(literals)), int32(len(literals)), vars, &count, threshold, verbose)
}
func Rpo_Recursion(target []uint32, literals []*Literal_t, nLiterals, nActive, vars int32, count *int32, maximum, verbose int32) *Literal_t {
	(*count)++
	if maximum > 0 && *count >= maximum {
		return nil
	}
	if verbose != 0 {
		fmt.Printf("Entering recursion %d\n", *count)
	}
	if nActive == 1 {
		for _, literal := range literals[:nLiterals] {
			if literal != nil && rpoTruthEqual(target, literal.Function, vars) {
				return Lit_Copy(literal, vars)
			}
		}
		return nil
	}
	if nActive < 1 {
		return nil
	}
	degree := make([]int32, nLiterals)
	edges := make([]*Rpo_LCI_Edge_t, 0, nLiterals*(nLiterals-1)/2)
	for i := int32(0); i < nLiterals; i++ {
		if literals[i] == nil {
			continue
		}
		for j := i + 1; j < nLiterals; j++ {
			if literals[j] == nil {
				continue
			}
			if Rpo_CheckANDGroup(literals[i], literals[j], vars) {
				edges = append(edges, Rpo_CreateEdge(LIT_AND, i, j, degree))
			}
			if Rpo_CheckORGroup(literals[i], literals[j], vars) {
				edges = append(edges, Rpo_CreateEdge(LIT_OR, i, j, degree))
			}
		}
	}
	for {
		index := Rpo_computeMinEdgeCost(edges, int32(len(edges)), degree)
		if index < 0 {
			return nil
		}
		edge := edges[index]
		edge.Visited = true
		lo, hi := min(edge.Idx1, edge.Idx2), max(edge.Idx1, edge.Idx2)
		a, b := literals[lo], literals[hi]
		group := Lit_GroupLiterals(a, b, edge.ConnectionType, vars)
		literals[lo], literals[hi] = group, nil
		if verbose != 0 {
			fmt.Printf("New Literal Cluster found: %s\n", group.Expression)
		}
		result := Rpo_Recursion(target, literals, nLiterals, nActive-1, vars, count, maximum, verbose)
		literals[lo], literals[hi] = a, b
		if result != nil {
			return result
		}
		if maximum > 0 && *count >= maximum {
			return nil
		}
	}
}
