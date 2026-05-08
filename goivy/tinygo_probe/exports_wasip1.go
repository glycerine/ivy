//go:build !tinygo && wasip1

package main

//go:wasmexport ivy_probe_init
func ivyProbeInit() uint32 {
	return probeInit()
}

//go:wasmexport ivy_probe_z3_bool_sort_id
func ivyProbeZ3BoolSortID() uint32 {
	return probeZ3BoolSortID()
}

//go:wasmexport ivy_probe_round_trip_bool_sort_id
func ivyProbeRoundTripBoolSortID() uint32 {
	return probeRoundTripBoolSortID()
}

//go:wasmexport ivy_probe_smt_solver_sat_true
func ivyProbeSMTSolverSatTrue() int32 {
	return probeSMTSolverSatTrue()
}

//go:wasmexport ivy_probe_smt_solver_unsat_true_and_not_true
func ivyProbeSMTSolverUnsatTrueAndNotTrue() int32 {
	return probeSMTSolverUnsatTrueAndNotTrue()
}

//go:wasmexport ivy_probe_checksum
func ivyProbeChecksum() uint32 {
	return probeChecksumValue()
}
