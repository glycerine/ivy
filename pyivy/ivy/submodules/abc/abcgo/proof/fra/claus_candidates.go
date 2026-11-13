package fra

// Cut-derived inductive-clause candidates, ported from fraClaus.c.
import (
	"fmt"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/opt/dar"
	"github.com/glycerine/abcgo/sat/bsat"
	"github.com/glycerine/abcgo/sat/cnf"
)

func clausScores(p *Clu_Man_t, s *Fra_Sml_t, leaves []aig.Edge) []int32 {
	if len(leaves) < 2 || len(leaves) > 16 || (s.WordsTotal-int(p.nSimWordsPref))%8 != 0 {
		panic("fra: invalid simulation cut")
	}
	sim := make([][]uint32, len(leaves))
	for i, e := range leaves {
		sim[i] = s.NodeWords(e.NodeID())
	}
	return clausPatternScores(sim, int(p.nSimWordsPref))
}
func Fra_ClausProcessClausesCut(p *Clu_Man_t, s *Fra_Sml_t, cut *dar.Cut, scores []int32) int32 {
	if len(cut.Leaves) > 4 || len(scores) < 16 {
		panic("fra: invalid four-cut")
	}
	values := clausScores(p, s, cut.Leaves)
	clear(scores[:16])
	copy(scores, values)
	var present uint32
	for i, v := range values {
		if v != 0 {
			present |= 1 << uint(i)
		}
	}
	return int32(present)
}
func Fra_ClausProcessClausesCut2(p *Clu_Man_t, s *Fra_Sml_t, cut *dar.Cut, scores []int32) int32 {
	return Fra_ClausProcessClausesCut(p, s, cut, scores)
}
func Fra_ClausProcessClausesCut3(p *Clu_Man_t, s *Fra_Sml_t, cut *aig.Aig_Cut_t, scores []int32) {
	ids := cut.Leaves()
	leaves := make([]aig.Edge, len(ids))
	for i, id := range ids {
		leaves[i] = aig.Edge(id << 1)
	}
	values := clausScores(p, s, leaves)
	if len(scores) < len(values) {
		panic("fra: short cut scores")
	}
	copy(scores, values)
}
func Fra_ClausSelectClauses(p *Clu_Man_t) int32 {
	costs := vec.Vec_IntArray(p.vCosts)
	threshold, count := clausSelectCosts(costs, int(p.nClausesMax), int(p.nSimWords))
	p.nClauses = int32(count)
	if p.fVerbose != 0 {
		fmt.Printf("Selected %d clauses. Cost range: [1 < %d < %d]\n", count, threshold, p.nSimWords*32+1)
	}
	return threshold
}
func clausRecord(p *Clu_Man_t, leaves []aig.Edge, minterm, cost int32) {
	for i, e := range leaves {
		v := p.pCnf.VarNums[e.NodeID()]
		if v < 0 {
			panic("fra: cut leaf has no CNF variable")
		}
		vec.Vec_IntPush(p.vLits, 2*v+int32((uint32(minterm)>>uint(i))&1))
	}
	vec.Vec_IntPush(p.vClauses, int32(vec.Vec_IntSize(p.vLits)))
	vec.Vec_IntPush(p.vCosts, cost)
}
func Fra_ClausRecordClause(p *Clu_Man_t, cut *dar.Cut, minterm, cost int32) {
	clausRecord(p, cut.Leaves, minterm, cost)
}
func Fra_ClausRecordClause2(p *Clu_Man_t, cut *aig.Aig_Cut_t, minterm, cost int32) {
	ids := cut.Leaves()
	leaves := make([]aig.Edge, len(ids))
	for i, id := range ids {
		leaves[i] = aig.Edge(id << 1)
	}
	clausRecord(p, leaves, minterm, cost)
}
func Fra_ClausSmlNodeIsConst(s *Fra_Sml_t, obj aig.Edge) int32 {
	return fraBool(s.IsConstant(obj.NodeID()))
}
func Fra_ClausSmlNodesAreImp(s *Fra_Sml_t, left, right aig.Edge) int32 {
	a, b := s.NodeWords(left.NodeID()), s.NodeWords(right.NodeID())
	for i := s.PrefixWords; i < s.WordsTotal; i++ {
		if a[i]&^b[i] != 0 {
			return 0
		}
	}
	return 1
}
func Fra_ClausSmlNodesAreImpC(s *Fra_Sml_t, left, right aig.Edge) int32 {
	a, b := s.NodeWords(left.NodeID()), s.NodeWords(right.NodeID())
	for i := s.PrefixWords; i < s.WordsTotal; i++ {
		if a[i]&b[i] != 0 {
			return 0
		}
	}
	return 1
}
func Fra_ClausCollectLatchClauses(p *Clu_Man_t, s *Fra_Sml_t) int32 {
	oldPrefix := s.PrefixWords
	s.PrefixWords = int(p.nSimWordsPref)
	defer func() { s.PrefixWords = oldPrefix }()
	inputs := p.pAig.Inputs()
	regs := int(p.pAig.RegisterCount())
	latches := inputs[len(inputs)-regs:]
	cost := p.nSimWords * 32
	consts, imps := int32(0), int32(0)
	add := func(lits ...int32) {
		for _, lit := range lits {
			vec.Vec_IntPush(p.vLits, lit)
		}
		vec.Vec_IntPush(p.vClauses, int32(vec.Vec_IntSize(p.vLits)))
		vec.Vec_IntPush(p.vCosts, cost)
	}
	for _, a := range latches {
		av := p.pCnf.VarNums[a.NodeID()]
		if Fra_ClausSmlNodeIsConst(s, a) != 0 {
			add(2*av + 1)
			consts++
			continue
		}
		for _, b := range latches {
			bv := p.pCnf.VarNums[b.NodeID()]
			switch {
			case Fra_ClausSmlNodesAreImp(s, a, b) != 0:
				add(2*av+1, 2*bv)
			case Fra_ClausSmlNodesAreImp(s, b, a) != 0:
				add(2*bv+1, 2*av)
			case Fra_ClausSmlNodesAreImpC(s, a, b) != 0:
				add(2*av+1, 2*bv+1)
			default:
				continue
			}
			imps++
		}
		if consts+imps > p.nClausesMax/2 {
			break
		}
	}
	p.nOneHots = consts + imps
	p.nOneHotsProven = 0
	if p.fVerbose != 0 {
		fmt.Printf("Collected %d register constants and %d one-hotness implications.\n", consts, imps)
	}
	return 0
}

func clausSimulations(p *Clu_Man_t) (*Fra_Sml_t, *Fra_Sml_t) {
	aig.Aig_ManRandom(1)
	seq := Fra_SmlSimulateSeq(p.pAig, 0, p.nPref+p.nSimFrames, p.nSimWords/p.nSimFrames, 1)
	if p.fTarget != 0 && seq.NonconstantOutput {
		fmt.Println("Property failed after sequential simulation!")
		return nil, nil
	}
	aig.Aig_ManRandom(1)
	return seq, Fra_SmlSimulateComb(p.pAig, p.nSimWords+p.nSimWordsPref, 0)
}
func clausFinishSelection(p *Clu_Man_t, filterProven bool) int32 {
	p.nClauses = int32(vec.Vec_IntSize(p.vClauses))
	if filterProven && vec.Vec_IntSize(p.vClausesProven) > 0 {
		solver := cnf.Cnf_DataWriteIntoSolver(p.pCnf, 1, 0)
		if solver == nil {
			fmt.Println("Error: Main solver is unsat.")
			return -1
		}
		p.pSatMain = solver
		start := 0
		lits := vec.Vec_IntArray(p.vLitsProven)
		for _, end := range vec.Vec_IntArray(p.vClausesProven) {
			if !solver.AddClause(lits[start:int(end)]...) {
				fmt.Println("Error: Solver is UNSAT after adding assumption clauses.")
				return -1
			}
			start = int(end)
		}
		start = 0
		lits = vec.Vec_IntArray(p.vLits)
		solver.SetLimits(int64(p.nBTLimit), 0)
		for i, end := range vec.Vec_IntArray(p.vClauses) {
			assumptions := make([]int32, int(end)-start)
			for j, lit := range lits[start:int(end)] {
				assumptions[j] = lit ^ 1
			}
			if solver.Solve(assumptions...) == bsat.Unsatisfiable {
				vec.Vec_IntWriteEntry(p.vCosts, i, -1)
				p.nClauses--
			}
			start = int(end)
		}
	}
	p.fFiltering = 0
	if p.nClauses > p.nClausesMax {
		Fra_ClausSelectClauses(p)
		p.fFiltering = 1
	}
	return 1
}
func Fra_ClausProcessClauses2(p *Clu_Man_t, refs int32) int32 {
	p.nCuts = 0
	if p.nLutSize >= 13 {
		panic("fra: cut size exceeds source limit")
	}
	seq, comb := clausSimulations(p)
	if seq == nil {
		return 0
	}
	if refs != 0 {
		Fra_ClausCollectLatchClauses(p, seq)
	}
	cuts := aig.Aig_ComputeCuts(p.pAig, p.nCutsMax, p.nLutSize, 0, p.fVerbose)
	defer aig.Aig_ManCutStop(cuts)
	for _, obj := range p.pAig.LiveEdges() {
		node := p.pAig.Node(obj)
		if (node.Kind != aig.And && node.Kind != aig.ExclusiveOr) || int32(node.Level) > p.nLevels {
			continue
		}
		for _, cut := range cuts.Cuts(obj) {
			leaves := cut.Leaves()
			if len(leaves) < 2 {
				continue
			}
			p.nCuts++
			edges := make([]aig.Edge, len(leaves))
			for i, id := range leaves {
				edges[i] = aig.Edge(id << 1)
			}
			seqScores, combScores := clausScores(p, seq, edges), clausScores(p, comb, edges)
			for minterm, cost := range combScores {
				if cost != 0 && seqScores[minterm] == 0 {
					clausRecord(p, edges, int32(minterm), cost)
				}
			}
		}
	}
	return clausFinishSelection(p, true)
}
func Fra_ClausProcessClauses(p *Clu_Man_t, refs int32) int32 {
	p.nCuts = 0
	seq, comb := clausSimulations(p)
	if seq == nil {
		return 0
	}
	if refs != 0 {
		Fra_ClausCollectLatchClauses(p, seq)
	}
	cuts, err := dar.Dar_ManComputeCuts(p.pAig, 10, false)
	if err != nil {
		panic(err)
	}
	for _, obj := range p.pAig.LiveEdges() {
		node := p.pAig.Node(obj)
		if node.Kind != aig.And && node.Kind != aig.ExclusiveOr {
			continue
		}
		for _, cut := range cuts[obj] {
			if len(cut.Leaves) < 2 {
				continue
			}
			p.nCuts++
			seqScores, combScores := clausScores(p, seq, cut.Leaves), clausScores(p, comb, cut.Leaves)
			for minterm, cost := range combScores {
				if cost != 0 && seqScores[minterm] == 0 {
					clausRecord(p, cut.Leaves, int32(minterm), cost)
				}
			}
		}
	}
	return clausFinishSelection(p, false)
}
