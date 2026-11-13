package fra

// Simulation-guided implication generation from ABC's fraImp.c. The
// checking and refinement routines are in fraImpCheck.go and
// fraImpSimulation.go.
import (
	"fmt"
	"math/bits"
	"sort"
	"time"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
)

func Fra_SmlCountOnesOne(s *Fra_Sml_t, node int32) int32 {
	var result int32
	for _, word := range s.nodeWords(int(node))[s.prefixWords:] {
		result += int32(bits.OnesCount32(word))
	}
	return result
}
func Fra_SmlCountOnes(s *Fra_Sml_t) []int32 {
	counts := make([]int32, len(s.graph.Nodes()))
	for _, edge := range s.graph.LiveEdges() {
		counts[edge.NodeID()] = Fra_SmlCountOnesOne(s, int32(edge.NodeID()))
	}
	return counts
}
func Sml_NodeNotImpWeight(s *Fra_Sml_t, left, right int32) int32 {
	a, b := s.nodeWords(int(left)), s.nodeWords(int(right))
	var count int32
	for i := s.prefixWords; i < s.wordsTotal; i++ {
		count += int32(bits.OnesCount32(a[i] &^ b[i]))
	}
	return count
}
func Sml_NodeSaveNotImpPatterns(s *Fra_Sml_t, left, right int32, result []uint32) {
	a, b := s.nodeWords(int(left)), s.nodeWords(int(right))
	for i := s.prefixWords; i < s.wordsTotal; i++ {
		result[i] |= a[i] &^ b[i]
	}
}
func Fra_SmlSortUsingOnes(s *Fra_Sml_t, latchCorr int32) [][]int32 {
	buckets := make([][]int32, s.wordsTotal*32+1)
	for _, edge := range s.graph.LiveEdges() {
		if edge.NodeID() == 0 {
			continue
		}
		kind := s.graph.Node(edge).Kind
		if kind != aig.Input && (latchCorr != 0 || kind != aig.And && kind != aig.ExclusiveOr) {
			continue
		}
		count := Fra_SmlCountOnesOne(s, int32(edge.NodeID()))
		buckets[count] = append(buckets[count], int32(edge.NodeID()))
	}
	return buckets
}
func Fra_SmlSelectMaxCost(imps *vec.Vec_Int_t, costs []int32, maxCost, limit int32, costRange *int32) *vec.Vec_Int_t {
	if limit < 0 || int(limit) > vec.Vec_IntSize(imps) || len(costs) < vec.Vec_IntSize(imps) || maxCost < 0 {
		panic("fra: invalid implication costs")
	}
	counts := make([]int, int(maxCost)+1)
	for _, cost := range costs[:vec.Vec_IntSize(imps)] {
		if cost <= 0 || cost > maxCost {
			panic("fra: invalid implication cost")
		}
		counts[cost]++
	}
	pivot := maxCost
	seen := 0
	for pivot > 0 {
		seen += counts[pivot]
		if seen >= int(limit) {
			break
		}
		pivot--
	}
	out := vec.Vec_IntAlloc(int(limit))
	for i, imp := range vec.Vec_IntArray(imps) {
		if costs[i] >= pivot && vec.Vec_IntSize(out) < int(limit) {
			vec.Vec_IntPush(out, imp)
		}
	}
	if costRange != nil {
		*costRange = pivot
	}
	return out
}
func fraImpMaxID(imp int32) int32 {
	l, r := Fra_ImpLeft(imp), Fra_ImpRight(imp)
	if l > r {
		return l
	}
	return r
}

func Fra_ImpDerive(p *Fra_Man_t, maxLimit, useLimit, latchCorr int32) *vec.Vec_Int_t {
	if len(p.pManAig.Nodes()) >= 1<<15 || maxLimit <= 0 || useLimit <= 0 || useLimit > maxLimit {
		panic("fra: invalid implication derivation limits")
	}
	started := time.Now()
	comb := Fra_SmlSimulateComb(p.pManAig, 64, 0)
	seq := Fra_SmlSimulateSeq(p.pManAig, p.pPars.NFramesP, 64, 1, 1)
	buckets := Fra_SmlSortUsingOnes(seq, latchCorr)
	imps := vec.Vec_IntAlloc(int(maxLimit))
	costs := make([]int32, 0, maxLimit)
	var total, tried, nonSeq, combTrue, collected int64
	for k := 64 * 32; k > 0; k-- {
		for i := k - 1; i > 0; i-- {
			total += int64(len(buckets[i]) * len(buckets[k]))
		}
	}
	minCost, maxCost := int32(^uint32(0)>>1), int32(0)
	for k := seq.wordsTotal * 32; k > 0 && vec.Vec_IntSize(imps) < int(maxLimit); k-- {
		for i := k - 1; i > 0 && vec.Vec_IntSize(imps) < int(maxLimit); i-- {
			for _, left := range buckets[i] {
				for _, right := range buckets[k] {
					tried++
					if Sml_NodeCheckImp(seq, left, right) == 0 {
						nonSeq++
						continue
					}
					if Sml_NodeCheckImp(comb, left, right) != 0 {
						combTrue++
						continue
					}
					cost := Sml_NodeNotImpWeight(comb, left, right)
					if cost < minCost {
						minCost = cost
					}
					if cost > maxCost {
						maxCost = cost
					}
					costs = append(costs, cost)
					vec.Vec_IntPush(imps, Fra_ImpCreate(left, right))
					collected++
					if vec.Vec_IntSize(imps) == int(maxLimit) {
						break
					}
				}
				if vec.Vec_IntSize(imps) == int(maxLimit) {
					break
				}
			}
		}
	}
	pivot := minCost
	if vec.Vec_IntSize(imps) > int(useLimit) {
		imps = Fra_SmlSelectMaxCost(imps, costs, 64*32, useLimit, &pivot)
	}
	values := vec.Vec_IntArray(imps)
	sort.SliceStable(values, func(i, j int) bool { return fraImpMaxID(values[i]) < fraImpMaxID(values[j]) })
	if p.pPars.FVerbose != 0 {
		fmt.Printf("Implications: All = %d. Try = %d. NonSeq = %d. Comb = %d. Res = %d.\n", total, tried, nonSeq, combTrue, collected)
		fmt.Printf("Implication weight: Min = %d. Pivot = %d. Max = %d. Time = %.2f sec\n", minCost, pivot, maxCost, time.Since(started).Seconds())
	}
	return imps
}

func Fra_ImpAddToSolver(p *Fra_Man_t, imps *vec.Vec_Int_t, satVarNums []int32) {
	for i, imp := range vec.Vec_IntArray(imps) {
		left, right := aig.Edge(Fra_ImpLeft(imp)<<1), aig.Edge(Fra_ImpRight(imp)<<1)
		missing := false
		for f := int32(0); f < p.pPars.NFramesK; f++ {
			if Fra_ObjFraig(p, left, f) == aig.InvalidEdge || Fra_ObjFraig(p, right, f) == aig.InvalidEdge {
				missing = true
				break
			}
		}
		if missing {
			vec.Vec_IntWriteEntry(imps, i, 0)
			continue
		}
		for f := int32(0); f < p.pPars.NFramesK; f++ {
			a, b := Fra_ObjFraig(p, left, f), Fra_ObjFraig(p, right, f)
			av, bv := satVarNums[a.NodeID()], satVarNums[b.NodeID()]
			// Go CNF numbering may assign zero to a real input variable.
			if av < 0 || av >= p.nSatVars || bv < 0 || bv >= p.nSatVars {
				panic("fra: invalid implication SAT variable")
			}
			negA := p.pManAig.Node(left).Phase != a.Complemented()
			negB := p.pManAig.Node(right).Phase != b.Complemented()
			if !p.pSat.AddClause(fraLit(av, !negA), fraLit(bv, negB)) {
				p.pSat = nil
				return
			}
		}
	}
	// Propagate and simplify without consuming a SAT search budget.
	if !p.pSat.Simplify() {
		p.pSat = nil
		return
	}
	Fra_ImpCompactArray(imps)
}

func Fra_ImpVerifyUsingSimulation(p *Fra_Man_t) int32 {
	if p.pCla.vImps == nil || vec.Vec_IntSize(p.pCla.vImps) == 0 {
		return 0
	}
	seq := Fra_SmlSimulateSeq(p.pManAig, p.pPars.NFramesP, 2000, 8, 1)
	var failed int32
	for _, imp := range vec.Vec_IntArray(p.pCla.vImps) {
		if Sml_NodeCheckImp(seq, Fra_ImpLeft(imp), Fra_ImpRight(imp)) == 0 {
			failed++
		}
	}
	return failed
}
func Fra_ImpRecordInManager(p *Fra_Man_t, graph *aig.Graph) {
	if p.pCla.vImps == nil || vec.Vec_IntSize(p.pCla.vImps) == 0 {
		return
	}
	initial := graph.OutputCount()
	for _, imp := range vec.Vec_IntArray(p.pCla.vImps) {
		left, right := aig.Edge(Fra_ImpLeft(imp)<<1), aig.Edge(Fra_ImpRight(imp)<<1)
		a, b := aig.Aig_ObjCopy(p.pManAig, left), aig.Aig_ObjCopy(p.pManAig, right)
		if a == aig.InvalidEdge || b == aig.InvalidEdge {
			panic("fra: missing implication output mapping")
		}
		if !p.pManAig.Node(left).Phase {
			a = a.Not()
		}
		if p.pManAig.Node(right).Phase {
			b = b.Not()
		}
		graph.AddOutput(graph.OR(a, b))
	}
	graph.Assertions = int32(graph.OutputCount() - initial)
}
