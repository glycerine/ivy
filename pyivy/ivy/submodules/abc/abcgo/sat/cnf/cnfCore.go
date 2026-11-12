package cnf

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
)

var s_pManCnf *Cnf_Man_t

func Cnf_ManPrepare() {
	if s_pManCnf == nil {
		s_pManCnf = Cnf_ManStart()
	}
}
func Cnf_ManRead() *Cnf_Man_t { return s_pManCnf }
func Cnf_ManFree() {
	if s_pManCnf != nil {
		Cnf_ManStop(s_pManCnf)
		s_pManCnf = nil
	}
}

// Cnf_DeriveMappingArray returns object-indexed SAT variable numbers for the
// direct Tseitin encoding. Unrepresented constant objects contain -1.
func Cnf_DeriveMappingArray(graph *Aig_Man_t) *Vec_Int_t {
	data, err := EncodeAIG(graph, 0)
	if err != nil {
		return nil
	}
	return vec.Vec_IntAllocArrayCopy(data.VarNums, len(data.VarNums))
}

func Cnf_DeriveWithMan(manager *Cnf_Man_t, graph *Aig_Man_t, nOutputs int32) *Cnf_Dat_t {
	if manager != nil {
		manager.pManAig = graph
	}
	data, err := EncodeAIG(graph, int(nOutputs))
	if err != nil {
		return nil
	}
	return data
}
func Cnf_Derive(graph *Aig_Man_t, nOutputs int32) *Cnf_Dat_t {
	return Cnf_DeriveWithMan(nil, graph, nOutputs)
}
func Cnf_DeriveOtherWithMan(manager *Cnf_Man_t, graph *Aig_Man_t, _ int32) *Cnf_Dat_t {
	return Cnf_DeriveWithMan(manager, graph, -1)
}
func Cnf_DeriveOther(graph *Aig_Man_t, skipTTMin int32) *Cnf_Dat_t {
	return Cnf_DeriveOtherWithMan(nil, graph, skipTTMin)
}

var _ *aig.Graph
