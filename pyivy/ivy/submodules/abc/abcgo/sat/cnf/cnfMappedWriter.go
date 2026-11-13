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
	nodes := graph.Nodes()
	outputs := graph.Outputs()
	if freeOutputs < 0 || int(freeOutputs) > len(outputs) {
		return nil, fmt.Errorf("cnf: invalid free output count")
	}
	if freeOutputs != 0 && ((graph.RegisterCount() == 0 && int(freeOutputs) != len(outputs)) || (graph.RegisterCount() != 0 && freeOutputs != graph.RegisterCount())) {
		return nil, fmt.Errorf("cnf: free output count must match outputs or registers")
	}
	result := &Cnf_Dat_t{pMan: graph, VarNums: make([]int32, len(nodes)), Obj2Clause: make([]int32, len(nodes)), Obj2Count: make([]int32, len(nodes))}
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
	for _, edge := range graph.Inputs() {
		result.VarNums[edge.NodeID()] = next
		next++
	}
	result.VarNums[0] = next
	next++
	result.Variables = int(next)
	result.nVars = next
	covers := cnfMsopCovers()
	appendClause := func(id int, clause []int32) {
		if result.Obj2Clause[id] < 0 {
			result.Obj2Clause[id] = int32(len(result.Clauses))
		}
		result.Obj2Count[id]++
		result.Clauses = append(result.Clauses, clause)
		result.nLiterals += int32(len(clause))
	}
	for _, item := range vec.Vec_PtrArray(mapped) {
		edge := item.(aig.Edge)
		id := edge.NodeID()
		if id >= len(manager.selectedCuts) || manager.selectedCuts[id] == nil {
			return nil, fmt.Errorf("cnf: missing mapped cut for %d", id)
		}
		cut := manager.selectedCuts[id]
		if len(cut.pFanins) > 16 || (len(cut.pFanins) < 5 && len(cut.truth) == 0) {
			return nil, fmt.Errorf("cnf: unsupported mapped cut for %d", id)
		}
		if len(cut.pFanins) >= 5 {
			vars := make([]int32, len(cut.pFanins))
			for i, leaf := range cut.pFanins {
				if leaf < 0 || int(leaf) >= len(result.VarNums) || result.VarNums[leaf] < 0 {
					return nil, fmt.Errorf("cnf: unmapped cut leaf %d", leaf)
				}
				vars[i] = result.VarNums[leaf]
			}
			for polarity, cover := range [2]*Vec_Int_t{cut.vIsop[1], cut.vIsop[0]} {
				if cover == nil {
					return nil, fmt.Errorf("cnf: missing ISOP for %d", id)
				}
				for _, cube := range vec.Vec_IntArray(cover) {
					clause := make([]int32, len(vars)+1)
					clause[0] = 2*result.VarNums[id] + int32(polarity)
					n := Cnf_IsopWriteCube(cube, int32(len(vars)), vars, clause[1:])
					appendClause(id, clause[:1+n])
				}
			}
			continue
		}
		truth := uint16(cut.truth[0])
		for polarity, cover := range [][]byte{covers[truth], covers[^truth]} {
			for _, encoded := range cover {
				cube := int(encoded)
				var states [4]int32
				for b := 3; b >= 0; b-- {
					digit := cube % 3
					cube /= 3
					if digit == 0 {
						states[b] = 1
					} else if digit == 1 {
						states[b] = 2
					}
				}
				clause := []int32{2*result.VarNums[id] + int32(polarity)}
				for i, leaf := range cut.pFanins {
					if leaf < 0 || int(leaf) >= len(result.VarNums) || result.VarNums[leaf] < 0 {
						return nil, fmt.Errorf("cnf: unmapped cut leaf %d", leaf)
					}
					if states[i] != 0 {
						clause = append(clause, 2*result.VarNums[leaf]+states[i]-1)
					}
				}
				appendClause(id, clause)
			}
		}
	}
	appendClause(0, []int32{2 * result.VarNums[0]})
	for i, edge := range outputs {
		driver := nodes[edge.NodeID()].Fanins[0]
		v := result.VarNums[driver.NodeID()]
		if v < 0 {
			return nil, fmt.Errorf("cnf: unmapped output driver")
		}
		sign := int32(0)
		if driver.Complemented() {
			sign = 1
		}
		if i < len(outputs)-int(freeOutputs) {
			appendClause(edge.NodeID(), []int32{2*v + sign})
		} else {
			out := result.VarNums[edge.NodeID()]
			appendClause(edge.NodeID(), []int32{2 * out, 2*v + 1 - sign})
			appendClause(edge.NodeID(), []int32{2*out + 1, 2*v + sign})
		}
	}
	result.nClauses = int32(len(result.Clauses))
	result.pVarNums = append([]int32(nil), result.VarNums...)
	result.pClauses = cloneClauses(result.Clauses)
	result.pObj2Clause = append([]int32(nil), result.Obj2Clause...)
	result.pObj2Count = append([]int32(nil), result.Obj2Count...)
	return result, nil
}
