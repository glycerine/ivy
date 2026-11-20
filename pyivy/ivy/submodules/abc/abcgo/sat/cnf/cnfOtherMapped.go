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
	objects := aigObjectCount(graph)
	result := &Cnf_Dat_t{pMan: graph, Variables: objects, nVars: int32(objects), VarNums: make([]int32, objects), Obj2Clause: make([]int32, objects), Obj2Count: make([]int32, objects)}
	for i := range objects {
		result.VarNums[i] = int32(i)
		result.Obj2Clause[i] = -1
		result.Obj2Count[i] = -1
	}
	for i := range graph.InputCount() {
		edge := aig.Aig_ManCi(graph, int32(i))
		result.Obj2Count[edge.NodeID()] = 0
	}
	outputs := graph.Outputs()
	writer, err := writeMappedCutClauses(manager, mapped, result, 2*len(outputs)+1, 4*len(outputs)+1)
	if err != nil {
		return nil, err
	}
	for _, edge := range outputs {
		id := edge.NodeID()
		fan := graph.Node(edge).Fanins[0]
		fid := fan.NodeID()
		compl := int32(0)
		if fan.Complemented() {
			compl = 1
		}
		writer.write(id, int32(2*id), int32(2*fid)+1-compl)
		writer.write(id, int32(2*id+1), int32(2*fid)+compl)
	}
	writer.write(0, 0)
	writer.finish()
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
