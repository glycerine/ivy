package fra

// Clause induction orchestration from fraClaus.c.
import (
	"fmt"
	"math/bits"
	"path/filepath"
	"strings"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/ioa"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"github.com/glycerine/abcgo/sat/cnf"
)

func Fra_ClausAlloc(graph *aig.Graph, frames, prefix, maxClauses, cutSize, levels, maxCuts, batches, step, target, verbose, veryVerbose int32) *Clu_Man_t {
	return &Clu_Man_t{pAig: graph, nFrames: frames, nPref: prefix, nClausesMax: maxClauses, nLutSize: cutSize, nLevels: levels, nCutsMax: maxCuts, nBatches: batches, fStepUp: step, fTarget: target, fVerbose: verbose, fVeryVerbose: veryVerbose, nSimWords: 512, nSimFrames: 32, nSimWordsPref: prefix * 512 / 32, vLits: vec.Vec_IntAlloc(1 << 14), vClauses: vec.Vec_IntAlloc(1 << 12), vCosts: vec.Vec_IntAlloc(1 << 12), vLitsProven: vec.Vec_IntAlloc(1 << 14), vClausesProven: vec.Vec_IntAlloc(1 << 12)}
}
func Fra_ClausFree(p *Clu_Man_t) {
	if p != nil {
		*p = Clu_Man_t{}
	}
}
func clausSlices(lits, ends *vec.Vec_Int_t) [][]int32 {
	values := vec.Vec_IntArray(lits)
	begin := 0
	out := make([][]int32, 0, vec.Vec_IntSize(ends))
	for _, end := range vec.Vec_IntArray(ends) {
		if end < int32(begin) || int(end) > len(values) {
			panic("invalid clause offsets")
		}
		out = append(out, values[begin:end])
		begin = int(end)
	}
	if begin != len(values) {
		panic("incomplete clause offsets")
	}
	return out
}
func Fra_ClausRunBmc(p *Clu_Man_t) int32 {
	p.pSatBmc.SetLimits(int64(p.nBTLimit), 0)
	out := p.pCnf.VarNums[p.pAig.Outputs()[0].NodeID()]
	for i := int32(0); i < p.nPref+p.nFrames; i++ {
		if p.pSatBmc.Solve(bsat.Positive(out+i*int32(p.pCnf.Variables))) != bsat.Unsatisfiable {
			return 0
		}
	}
	return 1
}
func Fra_ClausRunSat(p *Clu_Man_t) int32 {
	p.pSatMain.SetLimits(int64(p.nBTLimit), 0)
	out := p.pCnf.VarNums[p.pAig.Outputs()[0].NodeID()]
	lits := make([]int32, p.nFrames+1)
	for i := int32(0); i <= p.nFrames; i++ {
		v := out + i*int32(p.pCnf.Variables)
		lits[i] = bsat.Positive(v)
		if i != p.nFrames {
			lits[i] = bsat.Negative(v)
		}
	}
	if p.pSatMain.Solve(lits...) == bsat.Unsatisfiable {
		return 1
	}
	return 0
}
func Fra_ClausRunSat0(p *Clu_Man_t) int32 {
	p.pSatMain.SetLimits(int64(p.nBTLimit), 0)
	out := p.pCnf.VarNums[p.pAig.Outputs()[0].NodeID()]
	if p.pSatMain.Solve(bsat.Positive(out)) == bsat.Unsatisfiable {
		return 1
	}
	return 0
}
func transpose32a(a *[32]uint32) { clausTranspose32(a) }
func Fra_ClausBmcClauses(p *Clu_Man_t) int32 {
	p.pSatBmc.SetLimits(int64(p.nBTLimit), 0)
	return int32(clausFilterBMC(p.pSatBmc, clausSlices(p.vLits, p.vClauses), vec.Vec_IntArray(p.vCosts), p.nPref, p.nFrames, int32(p.pCnf.Variables)))
}
func Fra_ClausSimInfoClean(p *Clu_Man_t) {
	if p.models == nil {
		p.models = newClausModels(p.pCnf.Variables)
	} else {
		p.models.clear()
	}
}
func Fra_ClausSimInfoRealloc(_ *Clu_Man_t) { /* Go slices grow while recording. */ }
func Fra_ClausSimInfoRecord(p *Clu_Man_t, model []bool) {
	if p.models == nil {
		p.models = newClausModels(p.pCnf.Variables)
	}
	p.models.record(model)
}
func Fra_ClausSimInfoCheck(p *Clu_Man_t, lits []int32) int32 {
	if p.models != nil && p.models.disproves(lits, p.nFrames*int32(p.pCnf.Variables)) {
		return 1
	}
	return 0
}
func Fra_ClausInductiveClauses(p *Clu_Man_t) int32 {
	proven := clausSlices(p.vLitsProven, p.vClausesProven)
	candidates := clausSlices(p.vLits, p.vClauses)
	costs := vec.Vec_IntArray(p.vCosts)
	// Build the exact preceding-frame relation for the source's property check.
	p.pSatMain = cnf.Cnf_DataWriteIntoSolver(p.pCnf, p.nFrames+1, 0)
	if p.pSatMain == nil {
		return -1
	}
	p.pSatMain.SetLimits(int64(p.nBTLimit), 0)
	for f := int32(0); f < p.nFrames; f++ {
		offset := f * 2 * int32(p.pCnf.Variables)
		for _, clause := range proven {
			lifted := make([]int32, len(clause))
			for j, lit := range clause {
				lifted[j] = lit + offset
			}
			if !p.pSatMain.AddClause(lifted...) {
				return -1
			}
		}
		for i, clause := range candidates {
			if costs[i] == -1 {
				continue
			}
			lifted := make([]int32, len(clause))
			for j, lit := range clause {
				lifted[j] = lit + offset
			}
			if !p.pSatMain.AddClause(lifted...) {
				return -1
			}
		}
	}
	p.fFail = 0
	if p.fTarget != 0 && Fra_ClausRunSat0(p) == 0 {
		p.fFail = 1
	}
	removed, err := clausCheckInduction(p.pSatMain, p.pCnf.Variables, p.nFrames, candidates, costs)
	if err != nil {
		fmt.Println("Clause induction:", err)
		return -1
	}
	return int32(removed)
}

// clausRefineAndStore commits candidates only after the induction pass has
// completed successfully. A failed solver setup must not promote unchecked
// candidates to proven clauses.
func clausRefineAndStore(p *Clu_Man_t) bool {
	for {
		removed := Fra_ClausInductiveClauses(p)
		if removed < 0 {
			return false
		}
		if removed == 0 {
			Fra_ClausAddToStorage(p)
			return true
		}
		p.nClauses -= removed
	}
}
func Fra_ClausAddToStorage(p *Clu_Man_t) {
	clauses := clausSlices(p.vLits, p.vClauses)
	count := 0
	for i, clause := range clauses {
		if vec.Vec_IntEntry(p.vCosts, i) == -1 {
			continue
		}
		vec.Vec_IntPushArray(p.vLitsProven, clause, len(clause))
		vec.Vec_IntPush(p.vClausesProven, int32(vec.Vec_IntSize(p.vLitsProven)))
		count++
		if int32(i) < p.nOneHots {
			p.nOneHotsProven++
		}
	}
	vec.Vec_IntClear(p.vClauses)
	vec.Vec_IntClear(p.vLits)
	vec.Vec_IntClear(p.vCosts)
	p.nClauses = 0
	p.fNothingNew = 0
	if count == 0 {
		p.fNothingNew = 1
	}
}
func Fra_ClausPrintIndClauses(p *Clu_Man_t) {
	var counts [9]int
	for _, clause := range clausSlices(p.vLitsProven, p.vClausesProven) {
		n := len(clause)
		if n > 8 {
			n = 8
		}
		counts[n]++
	}
	fmt.Printf("SUMMARY: Total proved clauses = %d. Clause per lit: ", vec.Vec_IntSize(p.vClausesProven))
	for i, n := range counts {
		if n != 0 {
			fmt.Printf("%d=%d ", i, n)
		}
	}
	fmt.Println()
}
func clausVarIDs(p *Clu_Man_t) []int {
	ids := make([]int, p.pCnf.Variables)
	for i := range ids {
		ids[i] = -1
	}
	for id, v := range p.pCnf.VarNums {
		if v >= 0 && int(v) < len(ids) {
			ids[v] = id
		}
	}
	return ids
}

// Fra_ClausWriteIndClauses recreates the C routine's care AIGER output.
func Fra_ClausWriteIndClauses(p *Clu_Man_t) {
	src := p.pAig
	dst := aig.NewGraph()
	mapping := map[int]aig.Edge{0: dst.Const1()}
	for _, in := range src.Inputs() {
		mapping[in.NodeID()] = dst.NewInput()
	}
	var copyNode func(aig.Edge) aig.Edge
	copyNode = func(e aig.Edge) aig.Edge {
		value, ok := mapping[e.NodeID()]
		if !ok {
			n := src.Node(e)
			a := copyNode(n.Fanins[0])
			switch n.Kind {
			case aig.And:
				value = dst.AND(a, copyNode(n.Fanins[1]))
			case aig.ExclusiveOr:
				value = dst.XOR(a, copyNode(n.Fanins[1]))
			case aig.Buffer, aig.Output:
				value = a
			default:
				panic("invalid care graph node")
			}
			mapping[e.NodeID()] = value
		}
		if e.Complemented() {
			return value.Not()
		}
		return value
	}
	ids := clausVarIDs(p)
	for _, clause := range clausSlices(p.vLitsProven, p.vClausesProven) {
		out := dst.Const0()
		for _, lit := range clause {
			v := int(lit >> 1)
			if v >= len(ids) || ids[v] < 0 {
				panic("unmapped care literal")
			}
			e := copyNode(aig.Edge(ids[v]) << 1)
			if lit&1 != 0 {
				e = e.Not()
			}
			out = dst.OR(out, e)
		}
		dst.AddOutput(out)
	}
	name := strings.TrimSuffix(src.Name, filepath.Ext(src.Name)) + "_care.aig"
	if err := ioa.Ioa_WriteAiger(dst, name, 0, 1); err != nil {
		fmt.Println("Could not write care clauses:", err)
	}
}
func Fra_ClausEstimateCoverageOne(sim *Fra_Sml_t, lits []int32, ids []int, result []uint32) {
	for i := range result {
		result[i] = ^uint32(0)
	}
	for _, lit := range lits {
		v := int(lit >> 1)
		words := Fra_ObjSim(sim, int32(ids[v]))
		for i := range result {
			value := words[i]
			if lit&1 == 0 {
				value = ^value
			}
			result[i] &= value
		}
	}
}
func Fra_ClausEstimateCoverage(p *Clu_Man_t) {
	sim := Fra_SmlSimulateComb(p.pAig, 1<<11, 0)
	ids := clausVarIDs(p)
	total := make([]uint32, 1<<11)
	one := make([]uint32, len(total))
	for _, clause := range clausSlices(p.vLitsProven, p.vClausesProven) {
		Fra_ClausEstimateCoverageOne(sim, clause, ids, one)
		for i := range total {
			total[i] |= one[i]
		}
	}
	count := 0
	for _, v := range total {
		count += bits.OnesCount32(v)
	}
	fmt.Printf("Care states ratio = %f. (%d out of %d patterns)\n", float64(len(total)*32-count)/float64(len(total)*32), len(total)*32-count, len(total)*32)
}

// Fra_Claus runs ABC's clause discovery driver. As in the original C, its
// return value is 1 on normal completion and several failure paths; it does
// not communicate whether the property was proved. The driver reports the
// outcome separately and verifies its stored invariant when applicable.
func Fra_Claus(graph *aig.Graph, frames, prefix, maxClauses, cutSize, levels, maxCuts, batches, step, bmc, refs, target, verbose, veryVerbose int32) int32 {
	p := Fra_ClausAlloc(graph, frames, prefix, maxClauses, cutSize, levels, maxCuts, batches, step, target, verbose, veryVerbose)
	defer Fra_ClausFree(p)
	if graph == nil || frames <= 0 || prefix < 0 || (target != 0 && graph.OutputCount()-int(graph.RegisterCount()) != 1) {
		panic("invalid clause induction input")
	}
	p.pCnf = cnf.Cnf_DeriveSimple(graph, int32(graph.OutputCount()))
	p.pSatBmc = cnf.Cnf_DataWriteIntoSolver(p.pCnf, p.nPref+p.nFrames, 1)
	if p.pSatBmc == nil {
		fmt.Println("Error: BMC solver is unsat.")
		return 1
	}
	if target != 0 && Fra_ClausRunBmc(p) == 0 {
		fmt.Printf("Problem fails the base case after %d frame expansion.\n", p.nPref+p.nFrames)
		return 1
	}
	p.pSatMain = cnf.Cnf_DataWriteIntoSolver(p.pCnf, p.nFrames+1, 0)
	if p.pSatMain == nil {
		fmt.Println("Error: Main solver is unsat.")
		return 1
	}
	for batch := int32(0); batch < p.nBatches; batch++ {
		if batch != 0 && p.nLutSize < 12 && (p.fFiltering == 0 || p.fNothingNew != 0 || p.fStepUp != 0) {
			p.nLutSize++
		}
		if target != 0 && Fra_ClausRunSat(p) != 0 {
			fmt.Println("Problem is inductive without strengthening.")
			return 1
		}
		old := p.nPref
		p.nPref = 0
		p.nSimWordsPref = 0
		Fra_ClausProcessClauses2(p, refs)
		p.nPref = old
		p.nSimWordsPref = p.nPref * p.nSimWords / p.nSimFrames
		if bmc != 0 {
			p.nClauses -= Fra_ClausBmcClauses(p)
		}
		if !clausRefineAndStore(p) {
			fmt.Println("Fra_Claus: internal error; candidates were not stored.")
			return 1
		}
		if target != 0 {
			if p.fFail != 0 {
				fmt.Println("Property fails during refinement.")
			} else {
				fmt.Println("Property holds inductively after strengthening.")
				break
			}
		}
	}
	Fra_InvariantVerify(graph, frames, p.vClausesProven, p.vLitsProven)
	if verbose != 0 {
		Fra_ClausPrintIndClauses(p)
		Fra_ClausEstimateCoverage(p)
	}
	if target == 0 {
		Fra_ClausWriteIndClauses(p)
	}
	return 1
}
