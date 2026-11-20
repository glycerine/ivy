package cnf

import (
	"fmt"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
)

// cnfWriteMapped follows Cnf_ManWriteCnf: free COs, mapped nodes in preorder,
// CIs, and constant one receive variables in that order, starting at one.
func cnfWriteMapped(manager *Cnf_Man_t, mapped *vec.Vec_Ptr_t, freeOutputs int32) (*Cnf_Dat_t, error) {
	if manager == nil || manager.pManAig == nil || mapped == nil {
		return nil, fmt.Errorf("cnf: missing mapped graph")
	}
	graph := manager.pManAig
	objects := aigObjectCount(graph)
	outputs := graph.Outputs()
	if freeOutputs < 0 || int(freeOutputs) > len(outputs) {
		return nil, fmt.Errorf("cnf: invalid free output count")
	}
	// ABC's PDR caller temporarily sets nRegs to the CO count to leave all
	// outputs free. Accept that form without mutating the caller's graph.
	if freeOutputs != 0 && int(freeOutputs) != len(outputs) && (graph.RegisterCount() == 0 || freeOutputs != graph.RegisterCount()) {
		return nil, fmt.Errorf("cnf: free output count must match outputs or registers")
	}
	result := &Cnf_Dat_t{pMan: graph, VarNums: make([]int32, objects), Obj2Clause: make([]int32, objects), Obj2Count: make([]int32, objects)}
	for i := range result.VarNums {
		result.VarNums[i] = -1
		result.Obj2Clause[i] = -1
	}
	next := int32(1)
	for _, edge := range outputs[len(outputs)-int(freeOutputs):] {
		result.VarNums[edge.NodeID()] = next
		next++
	}
	for _, item := range vec.Vec_PtrArray(mapped) {
		edge, ok := item.(aig.Edge)
		if !ok || !graph.ValidEdge(edge) {
			return nil, fmt.Errorf("cnf: invalid mapped edge")
		}
		result.VarNums[edge.NodeID()] = next
		next++
	}
	for i := range graph.InputCount() {
		edge := aig.Aig_ManCi(graph, int32(i))
		result.VarNums[edge.NodeID()] = next
		next++
	}
	result.VarNums[0] = next
	next++
	result.Variables = int(next)
	result.nVars = next
	writer, err := writeMappedCutClauses(manager, mapped, result, 1+len(outputs)+int(freeOutputs), 1+len(outputs)+3*int(freeOutputs))
	if err != nil {
		return nil, err
	}
	writer.write(0, 2*result.VarNums[0])
	for i, edge := range outputs {
		driver := graph.Node(edge).Fanins[0]
		v := result.VarNums[driver.NodeID()]
		if v < 0 {
			return nil, fmt.Errorf("cnf: unmapped output driver")
		}
		sign := int32(0)
		if driver.Complemented() {
			sign = 1
		}
		if i < len(outputs)-int(freeOutputs) {
			writer.write(edge.NodeID(), 2*v+sign)
		} else {
			out := result.VarNums[edge.NodeID()]
			writer.write(edge.NodeID(), 2*out, 2*v+1-sign)
			writer.write(edge.NodeID(), 2*out+1, 2*v+sign)
		}
	}
	writer.finish()
	return result, nil
}

// Each public and private clause is a capacity-bounded window on its
// respective literal array. Copy the private literals in one pass at the end.
type mappedClauseWriter struct {
	result   *Cnf_Dat_t
	literals []int32
	private  []int32
}

func (w *mappedClauseWriter) finish() {
	result := w.result
	copy(w.private, w.literals)
	result.pVarNums = append([]int32(nil), result.VarNums...)
	result.pObj2Clause = append([]int32(nil), result.Obj2Clause...)
	result.pObj2Count = append([]int32(nil), result.Obj2Count...)
}

func (w *mappedClauseWriter) finishClause(id, n int) {
	result := w.result
	if result.Obj2Clause[id] < 0 {
		result.Obj2Clause[id] = result.nClauses
		result.Obj2Count[id] = 0
	}
	result.Obj2Count[id]++
	start := int(result.nLiterals)
	end := start + n
	i := result.nClauses
	result.Clauses[i] = w.literals[start:end:end]
	result.pClauses[i] = w.private[start:end:end]
	result.nLiterals = int32(end)
	result.nClauses++
}

func (w *mappedClauseWriter) write(id int, literals ...int32) {
	copy(w.literals[w.result.nLiterals:], literals)
	w.finishClause(id, len(literals))
}

func (w *mappedClauseWriter) writeCube(id int, out, cube int32, vars []int32) {
	dest := w.literals[w.result.nLiterals:]
	dest[0] = out
	n := Cnf_IsopWriteCube(cube, int32(len(vars)), vars, dest[1:])
	w.finishClause(id, 1+int(n))
}

// Both native writers count once, allocate the complete CNF, and emit the
// same cut clauses using their own SAT variable numbering.
func writeMappedCutClauses(manager *Cnf_Man_t, mapped *vec.Vec_Ptr_t, result *Cnf_Dat_t, clauses, literals int) (*mappedClauseWriter, error) {
	covers := cnfMsopCovers()
	objects := vec.Vec_PtrArray(mapped)
	for _, item := range objects {
		edge, ok := item.(aig.Edge)
		if !ok || !manager.pManAig.ValidEdge(edge) {
			return nil, fmt.Errorf("cnf: invalid mapped edge")
		}
		id := edge.NodeID()
		if id >= len(manager.selectedCuts) || manager.selectedCuts[id] == nil {
			return nil, fmt.Errorf("cnf: missing mapped cut for %d", id)
		}
		cut := manager.selectedCuts[id]
		n := len(cut.pFanins)
		if n > 16 || n < 5 && len(cut.truth) == 0 {
			return nil, fmt.Errorf("cnf: unsupported mapped cut for %d", id)
		}
		if n < 5 {
			truth := smallCutTruth(cut.truth[0], len(cut.pFanins))
			for _, cover := range [2][]byte{covers[truth], covers[^truth]} {
				clauses += len(cover)
				literals += len(cover)
				for _, cube := range cover {
					literals += int(sopCubes[cube].literals[n])
				}
			}
		} else {
			for _, cover := range cut.vIsop {
				if cover == nil {
					return nil, fmt.Errorf("cnf: missing ISOP for %d", id)
				}
				count := vec.Vec_IntSize(cover)
				clauses += count
				literals += count + int(Cnf_IsopCountLiterals(cover, int32(n)))
			}
		}
	}
	result.Clauses = make([][]int32, clauses)
	result.pClauses = make([][]int32, clauses)
	w := &mappedClauseWriter{result: result, literals: make([]int32, literals), private: make([]int32, literals)}
	var vars [16]int32
	for _, item := range objects {
		id := item.(aig.Edge).NodeID()
		cut := manager.selectedCuts[id]
		for i, leaf := range cut.pFanins {
			if leaf < 0 || int(leaf) >= len(result.VarNums) || result.VarNums[leaf] < 0 {
				return nil, fmt.Errorf("cnf: unmapped cut leaf %d", leaf)
			}
			vars[i] = result.VarNums[leaf]
		}
		inputs := vars[:len(cut.pFanins)]
		out := 2 * result.VarNums[id]
		if len(inputs) < 5 {
			truth := smallCutTruth(cut.truth[0], len(cut.pFanins))
			for polarity, cover := range [2][]byte{covers[truth], covers[^truth]} {
				for _, encoded := range cover {
					w.writeCube(id, out+int32(polarity), int32(sopCubes[encoded].packed), inputs)
				}
			}
		} else {
			for polarity, cover := range [2]*Vec_Int_t{cut.vIsop[1], cut.vIsop[0]} {
				for _, cube := range vec.Vec_IntArray(cover) {
					w.writeCube(id, out+int32(polarity), cube, inputs)
				}
			}
		}
	}
	return w, nil
}
