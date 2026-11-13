package cnf

import (
	"sync"

	"github.com/glycerine/abcgo/misc/vec"
)

var s_pManCnf *Cnf_Man_t
var s_manCnfMu sync.Mutex

func Cnf_ManPrepare() {
	s_manCnfMu.Lock()
	defer s_manCnfMu.Unlock()
	if s_pManCnf == nil {
		s_pManCnf = Cnf_ManStart()
	}
}
func Cnf_ManRead() *Cnf_Man_t { s_manCnfMu.Lock(); defer s_manCnfMu.Unlock(); return s_pManCnf }
func Cnf_ManFree() {
	s_manCnfMu.Lock()
	defer s_manCnfMu.Unlock()
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
	return deriveLegacyAIG(graph, nOutputs)
}
func Cnf_Derive(graph *Aig_Man_t, nOutputs int32) *Cnf_Dat_t {
	return Cnf_DeriveWithMan(nil, graph, nOutputs)
}
func Cnf_DeriveOtherWithMan(manager *Cnf_Man_t, graph *Aig_Man_t, _ int32) *Cnf_Dat_t {
	if graph == nil {
		return nil
	}
	if manager != nil {
		manager.pManAig = graph
	}
	return deriveLegacyAIG(graph, int32(graph.OutputCount()))
}
func Cnf_DeriveOther(graph *Aig_Man_t, skipTTMin int32) *Cnf_Dat_t {
	return Cnf_DeriveOtherWithMan(nil, graph, skipTTMin)
}
