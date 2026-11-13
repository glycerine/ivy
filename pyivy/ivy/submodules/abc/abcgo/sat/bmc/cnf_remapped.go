package bmc

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/sat/cnf"
)

// Cnf_DeriveGiaRemapped encodes GIA logic without asserting outputs.
// VarNums is indexed by original GIA object ID; output variables are
// equivalent to their drivers, including complemented and constant drivers.
func Cnf_DeriveGiaRemapped(graph *gia.Gia_Man_t) *cnf.Cnf_Dat_t {
	if graph == nil {
		return nil
	}
	data, err := gia.Mf_ManGenerateCnfResult(graph, 6, true, false, false)
	if err != nil {
		return nil
	}
	result := cnf.NewMappedCNF(data.Variables, data.Clauses, data.VarNums, data.Obj2Clause, data.Obj2Count)
	result.Mapping = append([]int32(nil), data.Mapping...)
	return result
}

// Cnf_DataLiftGia relocates the CNF's variables and clauses into a larger
// solver namespace. The GIA argument remains for source compatibility and
// validates that the corresponding manager exists.
func Cnf_DataLiftGia(data *cnf.Cnf_Dat_t, graph *gia.Gia_Man_t, variables int32) {
	if data == nil || graph == nil {
		return
	}
	data.LiftVariableOffsets(variables)
}
