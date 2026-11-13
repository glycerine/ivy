package cnf

import "sync"

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

// Cnf_DeriveMappingArray returns object-indexed offsets into packed cut records.
func Cnf_DeriveMappingArray(graph *Aig_Man_t) *Vec_Int_t {
	if graph == nil {
		return nil
	}
	manager := Cnf_ManStart()
	manager.pManAig = graph
	Cnf_DeriveMapping(manager)
	mapped := Cnf_ManScanMapping(manager, 1, 0)
	result := Cnf_ManWriteCnfMapping(manager, mapped)
	Cnf_ManStop(manager)
	return result
}

func Cnf_DeriveWithMan(manager *Cnf_Man_t, graph *Aig_Man_t, nOutputs int32) *Cnf_Dat_t {
	if graph == nil {
		return nil
	}
	if manager == nil {
		manager = Cnf_ManStart()
		defer Cnf_ManStop(manager)
	}
	manager.pManAig = graph
	Cnf_DeriveMapping(manager)
	mapped := Cnf_ManScanMapping(manager, 1, 1)
	return Cnf_ManWriteCnf(manager, mapped, nOutputs)
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
