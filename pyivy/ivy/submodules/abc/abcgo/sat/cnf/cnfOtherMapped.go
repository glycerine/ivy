package cnf

import (
	"fmt"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	dar "github.com/glycerine/abcgo/opt/dar/core"
)

// deriveOtherMapped follows Cnf_DeriveOther's DAR-cut, area-flow mapping and
// MSOP clause construction. Variables in this format are AIG object IDs.
func deriveOtherMapped(manager *Cnf_Man_t, graph *aig.Graph, skipTruthMin bool) (*Cnf_Dat_t, error) {
	cuts, err := dar.Dar_ManComputeCuts(graph, 10, skipTruthMin)
	if err != nil {
		return nil, err
	}
	manager.pManAig = graph
	nodes := graph.Nodes()
	manager.selectedCuts = make([]*Cnf_Cut_t, len(nodes))
	areas := make([]int, len(nodes))
	covers := cnfMsopCovers()
	var visit func(aig.Edge) error
	visit = func(edge aig.Edge) error {
		id := edge.NodeID()
		node := nodes[id]
		if node.Kind != aig.And && node.Kind != aig.ExclusiveOr || manager.selectedCuts[id] != nil {
			return nil
		}
		for _, fanin := range node.Fanins {
			if err := visit(fanin); err != nil {
				return err
			}
		}
		var best *dar.Cut
		bestArea, bestRefs := 0, 0
		for _, candidate := range cuts[edge] {
			if len(candidate.Leaves) == 1 && candidate.Leaves[0].NodeID() == id {
				continue
			}
			score := 10 * (len(covers[candidate.Truth]) + len(covers[^candidate.Truth]))
			refs := 0
			for _, leaf := range candidate.Leaves {
				ln := graph.Node(leaf)
				refs += ln.Refs
				if ln.Kind == aig.And || ln.Kind == aig.ExclusiveOr {
					if ln.Refs <= 0 {
						return fmt.Errorf("cnf: unreferenced cut leaf %d", leaf.NodeID())
					}
					score += areas[leaf.NodeID()] / ln.Refs
				}
			}
			if best == nil || score < bestArea || score == bestArea && refs > bestRefs {
				c := candidate
				best = &c
				bestArea, bestRefs = score, refs
			}
		}
		if best == nil {
			return fmt.Errorf("cnf: no nontrivial cut for node %d", id)
		}
		areas[id] = bestArea
		leaves := make([]int32, len(best.Leaves))
		for i, e := range best.Leaves {
			leaves[i] = int32(e.NodeID())
		}
		manager.selectedCuts[id] = Cnf_CutFromTruth(leaves, []uint32{uint32(best.Truth) | uint32(best.Truth)<<16})
		return nil
	}
	for _, edge := range graph.LiveEdges() {
		if err := visit(edge); err != nil {
			return nil, err
		}
	}
	mapped := Cnf_ManScanMapping(manager, 1, 1)
	result, err := writeOtherMapped(manager, mapped, covers)
	if err != nil {
		return nil, err
	}
	result.vMapping = Cnf_ManWriteCnfMapping(manager, mapped)
	result.Mapping = append([]int32(nil), vec.Vec_IntArray(result.vMapping)...)
	return result, nil
}

func writeOtherMapped(manager *Cnf_Man_t, mapped *vec.Vec_Ptr_t, covers [][]byte) (*Cnf_Dat_t, error) {
	graph := manager.pManAig
	nodes := graph.Nodes()
	result := &Cnf_Dat_t{pMan: graph, Variables: len(nodes), nVars: int32(len(nodes)), VarNums: make([]int32, len(nodes)), Obj2Clause: make([]int32, len(nodes)), Obj2Count: make([]int32, len(nodes))}
	for i := range nodes {
		result.VarNums[i] = int32(i)
		result.Obj2Clause[i] = -1
		result.Obj2Count[i] = -1
	}
	for _, edge := range graph.Inputs() {
		result.Obj2Count[edge.NodeID()] = 0
	}
	for _, item := range vec.Vec_PtrArray(mapped) {
		edge, ok := item.(aig.Edge)
		if !ok {
			return nil, fmt.Errorf("cnf: invalid mapped node")
		}
		id := edge.NodeID()
		cut := manager.selectedCuts[id]
		if cut == nil || len(cut.pFanins) > 4 {
			return nil, fmt.Errorf("cnf: missing mapped cut for %d", id)
		}
		result.Obj2Clause[id] = int32(len(result.Clauses))
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
				clause := []int32{int32(2*id + polarity)}
				for i, leaf := range cut.pFanins {
					if states[i] != 0 {
						clause = append(clause, 2*leaf+states[i]-1)
					}
				}
				result.Clauses = append(result.Clauses, clause)
			}
		}
		result.Obj2Count[id] = int32(len(result.Clauses)) - result.Obj2Clause[id]
	}
	for _, edge := range graph.Outputs() {
		id := edge.NodeID()
		fan := nodes[id].Fanins[0]
		fid := fan.NodeID()
		compl := int32(0)
		if fan.Complemented() {
			compl = 1
		}
		result.Obj2Clause[id] = int32(len(result.Clauses))
		result.Obj2Count[id] = 2
		result.Clauses = append(result.Clauses, []int32{int32(2 * id), int32(2*fid) + 1 - compl}, []int32{int32(2*id + 1), int32(2*fid) + compl})
	}
	result.Obj2Clause[0] = int32(len(result.Clauses))
	result.Obj2Count[0] = 1
	result.Clauses = append(result.Clauses, []int32{0})
	result.pClauses = cloneClauses(result.Clauses)
	result.pObj2Clause = append([]int32(nil), result.Obj2Clause...)
	result.pObj2Count = append([]int32(nil), result.Obj2Count...)
	result.pVarNums = append([]int32(nil), result.VarNums...)
	result.nClauses = int32(len(result.Clauses))
	for _, c := range result.Clauses {
		result.nLiterals += int32(len(c))
	}
	return result, nil
}

// Cnf_DeriveOtherMapped derives the mapped-cut CNF used by ABC's old GLA.
// Unlike Cnf_DeriveOther's current Go compatibility implementation, this
// returns LUT boundary records and uses object IDs directly as SAT variables.
func Cnf_DeriveOtherMapped(graph *aig.Graph, skipTTMin int32) (*Cnf_Dat_t, error) {
	if graph == nil {
		return nil, fmt.Errorf("cnf: nil AIG")
	}
	manager := Cnf_ManStart()
	defer Cnf_ManStop(manager)
	return deriveOtherMapped(manager, graph, skipTTMin != 0)
}
