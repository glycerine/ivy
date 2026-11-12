package bmc

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/sat/cnf"
)

// Cnf_DeriveGiaRemapped builds a typed CNF for all GIA outputs while treating
// the sequential terminals as ordinary SAT variables, matching the historic
// remapped-CNF callers' use of every CO as a constraint.
func Cnf_DeriveGiaRemapped(graph *gia.Gia_Man_t) *cnf.Cnf_Dat_t {
	if graph == nil {
		return nil
	}
	native := gia.Gia_ManToAigSimple(graph)
	data, err := cnf.EncodeAIG(native, native.OutputCount())
	if err != nil {
		return nil
	}
	return data
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
