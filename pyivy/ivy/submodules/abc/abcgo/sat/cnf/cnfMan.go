package cnf

import (
	"compress/gzip"
	"fmt"
	"os"

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
	return &Cnf_Man_t_{pSopSizes: make([]byte, 1<<16), pSops: make([]string, 1<<16), nMergeLimit: 10, vMemory: vec.Vec_IntAlloc(1024)}
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
	result := &Cnf_Dat_t_{Variables: int(variables), nVars: variables, Clauses: make([][]int32, 0, clauseCount), pClauses: make([][]int32, 0, clauseCount)}
	if graph != nil {
		result.VarNums = make([]int32, aigObjectCount(graph))
		for i := range result.VarNums {
			result.VarNums[i] = -1
		}
		result.pVarNums = append([]int32(nil), result.VarNums...)
	}
	_ = literalCount
	return result
}
func aigObjectCount(graph *Aig_Man_t) int {
	if graph == nil {
		return 0
	}
	return len(graph.Nodes())
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
	indices := make(map[int32]bool, vec.Vec_IntSize(flips))
	for _, i := range vec.Vec_IntArray(flips) {
		indices[i] = true
	}
	index := int32(0)
	for ci := range cnf.Clauses {
		for li, lit := range cnf.Clauses[ci] {
			if indices[index] {
				cnf.Clauses[ci][li] = (lit ^ 1) + variables<<1
			}
			index++
		}
	}
	cnf.pClauses = cloneClauses(cnf.Clauses)
}

func writeCNF(file *os.File, cnf *Cnf_Dat_t, zeroBased bool) error {
	if cnf == nil {
		return fmt.Errorf("cnf: nil data")
	}
	if _, err := fmt.Fprintf(file, "p cnf %d %d\n", cnf.Variables, len(cnf.Clauses)); err != nil {
		return err
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
	if err := writeCNF(os.Stdout, cnf, readable != 0); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
func Cnf_DataWriteIntoFile(cnf *Cnf_Dat_t, filename string, _ int32, _ *Vec_Int_t, _ *Vec_Int_t) {
	file, err := os.Create(filename)
	if err != nil {
		return
	}
	defer file.Close()
	_ = writeCNF(file, cnf, false)
}
func Cnf_DataWriteIntoFileGz(cnf *Cnf_Dat_t, filename string, readable int32, forall, exists *Vec_Int_t) {
	file, err := os.Create(filename)
	if err != nil {
		return
	}
	defer file.Close()
	writer := gzip.NewWriter(file)
	_ = writeCNFToWriter(writer, cnf, readable != 0)
	_ = writer.Close()
	_ = forall
	_ = exists
}
func writeCNFToWriter(writer *gzip.Writer, cnf *Cnf_Dat_t, zeroBased bool) error {
	if cnf == nil {
		return fmt.Errorf("cnf: nil data")
	}
	if _, err := fmt.Fprintf(writer, "p cnf %d %d\n", cnf.Variables, len(cnf.Clauses)); err != nil {
		return err
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
			if _, err := fmt.Fprintf(writer, "%d ", variable); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintln(writer, "0"); err != nil {
			return err
		}
	}
	return nil
}

func Cnf_DataWriteIntoSolverInt(sat *bsat.Solver, cnf *Cnf_Dat_t, frames int32, init int32) *bsat.Solver {
	if sat == nil || cnf == nil || frames < 1 {
		return nil
	}
	for sat.NumVars() < int32(cnf.Variables)*frames {
		sat.NewVariable()
	}
	for frame := int32(0); frame < frames; frame++ {
		for _, clause := range cnf.Clauses {
			lits := make([]bsat.Literal, len(clause))
			for i, lit := range clause {
				variable := int32(lit>>1) + frame*int32(cnf.Variables)
				if lit&1 != 0 {
					lits[i] = bsat.Negative(variable)
				} else {
					lits[i] = bsat.Positive(variable)
				}
			}
			if !sat.AddClause(lits...) {
				return nil
			}
		}
	}
	if frames > 1 {
		if cnf.pMan == nil {
			return nil
		}
		stride := int32(cnf.Variables)
		for frame := int32(1); frame < frames; frame++ {
			for i := int32(0); i < aig.Aig_ManRegNum(cnf.pMan); i++ {
				li, lo := aig.Aig_ManLi(cnf.pMan, i), aig.Aig_ManLo(cnf.pMan, i)
				liVar, loVar := cnf.VarNums[li.NodeID()], cnf.VarNums[lo.NodeID()]
				if liVar < 0 || loVar < 0 {
					return nil
				}
				left := liVar + (frame-1)*stride
				right := loVar + frame*stride
				if !sat.AddClause(bsat.Negative(left), bsat.Positive(right)) || !sat.AddClause(bsat.Positive(left), bsat.Negative(right)) {
					return nil
				}
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
	return sat
}
func Cnf_DataWriteIntoSolver(cnf *Cnf_Dat_t, frames int32, init int32) *bsat.Solver {
	return Cnf_DataWriteIntoSolverInt(bsat.NewSolver(), cnf, frames, init)
}
func Cnf_DataWriteIntoSolver2(cnf *Cnf_Dat_t, frames int32, init int32) *bsat.Solver {
	return Cnf_DataWriteIntoSolver(cnf, frames, init)
}
func Cnf_DataWriteOrClause(solver *bsat.Solver, cnf *Cnf_Dat_t) int32 {
	outputs, ok := cnfOutputLiterals(solver, cnf)
	if !ok || !solver.AddClause(outputs...) {
		return 0
	}
	return 1
}
func Cnf_DataWriteOrClause2(solver *bsat.Solver, cnf *Cnf_Dat_t) int32 {
	return Cnf_DataWriteOrClause(solver, cnf)
}
func cnfOutputLiterals(solver *bsat.Solver, cnf *Cnf_Dat_t) ([]bsat.Literal, bool) {
	if solver == nil || cnf == nil || cnf.pMan == nil {
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
	outputs, ok := cnfOutputLiterals(solver, cnf)
	if !ok {
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
