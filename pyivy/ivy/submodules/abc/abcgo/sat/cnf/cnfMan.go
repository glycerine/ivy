package cnf

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
)

func Cnf_Lit2Var(lit int32) int32 {
	if lit&1 != 0 {
		return -(lit >> 1) - 1
	}
	return (lit >> 1) + 1
}
func Cnf_Lit2Var2(lit int32) int32 {
	if lit&1 != 0 {
		return -(lit >> 1)
	}
	return lit >> 1
}

func Cnf_ManStart() *Cnf_Man_t {
	sizes, sops := cnfReadMsops()
	return &Cnf_Man_t_{pSopSizes: sizes, pSops: sops, nMergeLimit: 10, vMemory: vec.Vec_IntAlloc(1024)}
}
func Cnf_ManStop(p *Cnf_Man_t) {
	if p == nil {
		return
	}
	Cnf_ManFreeCuts(p)
	if p.vMemory != nil {
		vec.Vec_IntFree(p.vMemory)
	}
	p.pSopSizes, p.pSops, p.vMemory = nil, nil, nil
	p.pTruths = [4][]uint32{}
	p.pManAig = nil
}

func Cnf_DataCollectPiSatNums(cnf *Cnf_Dat_t, graph *Aig_Man_t) *Vec_Int_t {
	if cnf == nil || graph == nil {
		return vec.Vec_IntAllocExact(0)
	}
	result := vec.Vec_IntAllocExact(graph.InputCount())
	for _, input := range graph.Inputs() {
		id := input.NodeID()
		variable := int32(-1)
		if id >= 0 && id < len(cnf.VarNums) {
			variable = cnf.VarNums[id]
		}
		vec.Vec_IntPush(result, variable)
	}
	return result
}

func Cnf_DataAlloc(graph *Aig_Man_t, variables, clauseCount, literalCount int32) *Cnf_Dat_t {
	if variables < 0 || clauseCount < 0 || literalCount < 0 {
		return nil
	}
	clauses := make([][]int32, clauseCount)
	result := &Cnf_Dat_t_{pMan: graph, Variables: int(variables), nVars: variables, nClauses: clauseCount, nLiterals: literalCount, Clauses: clauses, pClauses: clauses}
	if graph != nil {
		result.VarNums = make([]int32, aigObjectCount(graph))
		for i := range result.VarNums {
			result.VarNums[i] = -1
		}
		result.pVarNums = append([]int32(nil), result.VarNums...)
	}
	return result
}
func aigObjectCount(graph *Aig_Man_t) int {
	if graph == nil {
		return 0
	}
	return int(aig.Aig_ManObjNumMax(graph))
}

func Cnf_DataDup(src *Cnf_Dat_t) *Cnf_Dat_t {
	if src == nil {
		return nil
	}
	result := *src
	result.Clauses = cloneClauses(src.Clauses)
	result.pClauses = cloneClauses(src.pClauses)
	result.VarNums = append([]int32(nil), src.VarNums...)
	result.pVarNums = append([]int32(nil), src.pVarNums...)
	result.Mapping = append([]int32(nil), src.Mapping...)
	result.Obj2Clause = append([]int32(nil), src.Obj2Clause...)
	result.Obj2Count = append([]int32(nil), src.Obj2Count...)
	result.pObj2Clause = append([]int32(nil), src.pObj2Clause...)
	result.pObj2Count = append([]int32(nil), src.pObj2Count...)
	result.pClaPols = append([]byte(nil), src.pClaPols...)
	if src.vMapping != nil {
		result.vMapping = vec.Vec_IntDup(src.vMapping)
	}
	return &result
}
func Cnf_DataFree(cnf *Cnf_Dat_t) {
	if cnf == nil {
		return
	}
	cnf.Clauses, cnf.pClauses = nil, nil
	cnf.VarNums, cnf.pVarNums = nil, nil
	cnf.Mapping, cnf.Obj2Clause, cnf.Obj2Count = nil, nil, nil
	cnf.pObj2Clause, cnf.pObj2Count, cnf.pClaPols = nil, nil, nil
	cnf.vMapping, cnf.pMan = nil, nil
}

func Cnf_DataLift(cnf *Cnf_Dat_t, variables int32) {
	if cnf == nil || variables == 0 {
		return
	}
	cnf.LiftVariableOffsets(variables)
}

// LiftVariableOffsets shifts literals and object mappings into a larger
// solver namespace while preserving this CNF's local variable count. BMC
// callers that concatenate independently encoded frames need that local count
// unchanged when advancing the next frame.
func (cnf *Cnf_Dat_t_) LiftVariableOffsets(variables int32) {
	if cnf == nil || variables == 0 {
		return
	}
	for i, v := range cnf.VarNums {
		if v >= 0 {
			cnf.VarNums[i] += variables
		}
	}
	cnf.pVarNums = append(cnf.pVarNums[:0], cnf.VarNums...)
	for i := range cnf.Clauses {
		for j := range cnf.Clauses[i] {
			cnf.Clauses[i][j] += variables << 1
		}
	}
	cnf.pClauses = cloneClauses(cnf.Clauses)
}
func Cnf_DataCollectFlipLits(cnf *Cnf_Dat_t, variable int32, flips *Vec_Int_t) {
	if flips == nil {
		return
	}
	vec.Vec_IntClear(flips)
	if cnf == nil {
		return
	}
	index := int32(0)
	for _, clause := range cnf.Clauses {
		for _, lit := range clause {
			if lit>>1 == variable {
				vec.Vec_IntPush(flips, index)
			}
			index++
		}
	}
}
func Cnf_DataLiftAndFlipLits(cnf *Cnf_Dat_t, variables int32, flips *Vec_Int_t) {
	if cnf == nil || flips == nil {
		return
	}
	indices := vec.Vec_IntArray(flips)
	if !slices.IsSorted(indices) {
		indices = slices.Clone(indices)
		slices.Sort(indices)
	}
	// Collected offsets are already sorted. Walk the clause boundaries once
	// and apply each occurrence, including repeated offsets in supplied lists.
	clause, start := 0, 0
	for _, index := range indices {
		if index < 0 {
			continue
		}
		for clause < len(cnf.Clauses) && int(index)-start >= len(cnf.Clauses[clause]) {
			start += len(cnf.Clauses[clause])
			clause++
		}
		if clause == len(cnf.Clauses) {
			break
		}
		lit := &cnf.Clauses[clause][int(index)-start]
		*lit = (*lit ^ 1) + 2*variables
	}
	cnf.pClauses = cloneClauses(cnf.Clauses)
}

func writeCNF(file io.Writer, cnf *Cnf_Dat_t, zeroBased bool, forall, exists *Vec_Int_t) error {
	if cnf == nil {
		return fmt.Errorf("cnf: nil data")
	}
	if _, err := fmt.Fprintf(file, "p cnf %d %d\n", cnf.Variables, len(cnf.Clauses)); err != nil {
		return err
	}
	for _, quantifier := range []struct {
		kind      string
		variables *Vec_Int_t
	}{{"a", forall}, {"e", exists}} {
		if quantifier.variables == nil {
			continue
		}
		if _, err := fmt.Fprintf(file, "%s ", quantifier.kind); err != nil {
			return err
		}
		for _, variable := range vec.Vec_IntArray(quantifier.variables) {
			if !zeroBased {
				variable++
			}
			if _, err := fmt.Fprintf(file, "%d ", variable); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintln(file, "0"); err != nil {
			return err
		}
	}
	for _, clause := range cnf.Clauses {
		for _, lit := range clause {
			variable := int(lit >> 1)
			if !zeroBased {
				variable++
			}
			if lit&1 != 0 {
				variable = -variable
			}
			if _, err := fmt.Fprintf(file, "%d ", variable); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintln(file, "0"); err != nil {
			return err
		}
	}
	return nil
}
func Cnf_DataPrint(cnf *Cnf_Dat_t, readable int32) {
	if err := writeCNF(os.Stdout, cnf, readable != 0, nil, nil); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
func Cnf_DataWriteIntoFile(cnf *Cnf_Dat_t, filename string, readable int32, forall, exists *Vec_Int_t) {
	if strings.HasSuffix(filename, ".gz") {
		Cnf_DataWriteIntoFileGz(cnf, filename, readable, forall, exists)
		return
	}
	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("Cnf_WriteIntoFile(): Output file cannot be opened.\n")
		return
	}
	defer file.Close()
	writer := bufio.NewWriter(file)
	_ = writeCNFFile(writer, cnf, readable, forall, exists)
	_ = writer.Flush()
}

// writeCNFFile is the body of Cnf_DataWriteIntoFile(Gz): ABC's comment
// line, the clauses, then an empty line.
func writeCNFFile(w io.Writer, cnf *Cnf_Dat_t, readable int32, forall, exists *Vec_Int_t) error {
	if _, err := fmt.Fprintf(w, "c Result of efficient AIG-to-CNF conversion using package CNF\n"); err != nil {
		return err
	}
	if err := writeCNF(w, cnf, readable != 0, forall, exists); err != nil {
		return err
	}
	_, err := fmt.Fprintf(w, "\n")
	return err
}
func Cnf_DataWriteIntoFileGz(cnf *Cnf_Dat_t, filename string, readable int32, forall, exists *Vec_Int_t) {
	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("Cnf_WriteIntoFile(): Output file cannot be opened.\n")
		return
	}
	defer file.Close()
	writer := gzip.NewWriter(file)
	_ = writeCNFFile(writer, cnf, readable, forall, exists)
	_ = writer.Close()
}

type frameSolver interface {
	*bsat.Solver | *bsat.Solver2
	AddClause(...bsat.Literal) bool
	NumVars() int32
	SetNumVars(int32)
	Simplify() bool
}

func loadFrameClauses[S frameSolver](sat S, cnf *Cnf_Dat_t, frames int32, init int32) S {
	if sat == nil || cnf == nil || frames < 1 || cnf.Variables < 0 ||
		cnf.Variables > (1<<30)/int(frames) || frames > 1 && cnf.pMan == nil {
		return nil
	}
	stride := int32(cnf.Variables)
	if total := stride * frames; total > sat.NumVars() {
		sat.SetNumVars(total)
	}
	var lifted []bsat.Literal
	for f := int32(0); f < frames; f++ {
		if f > 0 {
			for i := int32(0); i < aig.Aig_ManRegNum(cnf.pMan); i++ {
				li, lo := aig.Aig_ManLi(cnf.pMan, i), aig.Aig_ManLo(cnf.pMan, i)
				left, right := cnf.VarNums[li.NodeID()], cnf.VarNums[lo.NodeID()]
				if left < 0 || right < 0 {
					return nil
				}
				left += (f - 1) * stride
				right += f * stride
				if !sat.AddClause(bsat.Positive(left), bsat.Negative(right)) ||
					!sat.AddClause(bsat.Negative(left), bsat.Positive(right)) {
					return nil
				}
			}
		}
		for _, clause := range cnf.Clauses {
			if f != 0 {
				lifted = slices.Grow(lifted[:0], len(clause))[:len(clause)]
				for i, lit := range clause {
					lifted[i] = lit + 2*f*stride
				}
				clause = lifted
			}
			if !sat.AddClause(clause...) {
				return nil
			}
		}
	}
	if init != 0 && cnf.pMan != nil {
		for i := int32(0); i < aig.Aig_ManRegNum(cnf.pMan); i++ {
			lo := aig.Aig_ManLo(cnf.pMan, i)
			variable := cnf.VarNums[lo.NodeID()]
			if variable < 0 || !sat.AddClause(bsat.Negative(variable)) {
				return nil
			}
		}
	}
	if !sat.Simplify() {
		return nil
	}
	return sat
}

func Cnf_DataWriteIntoSolverInt(sat *bsat.Solver, cnf *Cnf_Dat_t, frames int32, init int32) *bsat.Solver {
	return loadFrameClauses(sat, cnf, frames, init)
}
func Cnf_DataWriteIntoSolver(cnf *Cnf_Dat_t, frames int32, init int32) *bsat.Solver {
	return loadFrameClauses(bsat.NewSolver(), cnf, frames, init)
}
func Cnf_DataWriteIntoSolver2(cnf *Cnf_Dat_t, frames int32, init int32) *bsat.Solver2 {
	return loadFrameClauses(bsat.NewSolver2(), cnf, frames, init)
}
func Cnf_DataWriteOrClause(solver *bsat.Solver, cnf *Cnf_Dat_t) int32 {
	return Cnf_DataWriteOrClause2(solver, cnf)
}
func Cnf_DataWriteOrClause2[S frameSolver](solver S, cnf *Cnf_Dat_t) int32 {
	if solver == nil {
		return 0
	}
	outputs, ok := cnfOutputLiterals(cnf)
	if !ok || !solver.AddClause(outputs...) {
		return 0
	}
	return 1
}
func cnfOutputLiterals(cnf *Cnf_Dat_t) ([]bsat.Literal, bool) {
	if cnf == nil || cnf.pMan == nil {
		return nil, false
	}
	outputs := cnf.pMan.Outputs()
	literals := make([]bsat.Literal, 0, len(outputs))
	for _, output := range outputs {
		id := output.NodeID()
		if id >= len(cnf.VarNums) || cnf.VarNums[id] < 0 {
			return nil, false
		}
		literals = append(literals, bsat.Positive(cnf.VarNums[id]))
	}
	return literals, true
}
func Cnf_DataWriteAndClauses(solver *bsat.Solver, cnf *Cnf_Dat_t) int32 {
	outputs, ok := cnfOutputLiterals(cnf)
	if solver == nil || !ok {
		return 0
	}
	for _, output := range outputs {
		if !solver.AddClause(output) {
			return 0
		}
	}
	return 1
}
func Cnf_DataTranformPolarity(cnf *Cnf_Dat_t, transform int32) {
	if cnf == nil || cnf.pMan == nil {
		return
	}
	phases := make(map[int32]bool)
	for id, node := range cnf.pMan.Nodes() {
		if id >= len(cnf.VarNums) || cnf.VarNums[id] < 0 || (transform == 0 && node.Kind == aig.Output) {
			continue
		}
		phases[cnf.VarNums[id]] = node.Phase
	}
	for i := range cnf.Clauses {
		for j := range cnf.Clauses[i] {
			if phases[cnf.Clauses[i][j]>>1] {
				cnf.Clauses[i][j] ^= 1
			}
		}
	}
	cnf.pClauses = cloneClauses(cnf.Clauses)
}
func Cnf_DataAddXorClause(sat *bsat.Solver, a, b, c int32) int32 {
	if sat == nil {
		return 0
	}
	clauses := [][]bsat.Literal{
		{bsat.Negative(a), bsat.Negative(b), bsat.Negative(c)},
		{bsat.Negative(a), bsat.Positive(b), bsat.Positive(c)},
		{bsat.Positive(a), bsat.Negative(b), bsat.Positive(c)},
		{bsat.Positive(a), bsat.Positive(b), bsat.Negative(c)},
	}
	for _, clause := range clauses {
		if !sat.AddClause(clause...) {
			return 0
		}
	}
	return 4
}
