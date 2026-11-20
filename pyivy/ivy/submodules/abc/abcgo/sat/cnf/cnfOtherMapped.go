package cnf

import (
	"fmt"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
)

// deriveOtherMapped follows Cnf_DeriveOther's DAR-cut, area-flow mapping and
// MSOP clause construction. Variables in this format are AIG object IDs.
func deriveOtherMapped(manager *Cnf_Man_t, graph *aig.Graph, skipTruthMin bool) (*Cnf_Dat_t, error) {
	manager.pManAig = graph
	if err := cnfApplyDarMapping(manager, skipTruthMin); err != nil {
		return nil, err
	}
	defer aig.Aig_ManResetRefs(graph)
	mapped := Cnf_ManScanMapping(manager, 1, 1)
	result, err := writeOtherMapped(manager, mapped)
	if err != nil {
		return nil, err
	}
	result.vMapping = Cnf_ManWriteCnfMapping(manager, mapped)
	result.Mapping = append([]int32(nil), vec.Vec_IntArray(result.vMapping)...)
	return result, nil
}

func writeOtherMapped(manager *Cnf_Man_t, mapped *vec.Vec_Ptr_t) (*Cnf_Dat_t, error) {
	if manager == nil || manager.pManAig == nil || mapped == nil {
		return nil, fmt.Errorf("cnf: missing mapped graph")
	}
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
	if err := writeMappedCutClauses(manager, mapped, result); err != nil {
		return nil, err
	}
	for _, edge := range graph.Outputs() {
		id := edge.NodeID()
		fan := nodes[id].Fanins[0]
		fid := fan.NodeID()
		compl := int32(0)
		if fan.Complemented() {
			compl = 1
		}
		result.appendMappedClause(id, []int32{int32(2 * id), int32(2*fid) + 1 - compl})
		result.appendMappedClause(id, []int32{int32(2*id + 1), int32(2*fid) + compl})
	}
	result.appendMappedClause(0, []int32{0})
	result.finishMappedClauses()
	return result, nil
}

// Cnf_DeriveOtherMapped derives the mapped-cut CNF used by ABC's old GLA.
// It returns LUT boundary records and uses object IDs directly as SAT variables.
func Cnf_DeriveOtherMapped(graph *aig.Graph, skipTTMin int32) (*Cnf_Dat_t, error) {
	if graph == nil {
		return nil, fmt.Errorf("cnf: nil AIG")
	}
	manager := Cnf_ManStart()
	defer Cnf_ManStop(manager)
	return deriveOtherMapped(manager, graph, skipTTMin != 0)
}
