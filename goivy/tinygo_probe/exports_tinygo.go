//go:build tinygo && wasip1

package main

//export ivy_probe_init
func ivyProbeInit() uint32 {
	return probeInit()
}

//export ivy_probe_z3_bool_sort_id
func ivyProbeZ3BoolSortID() uint32 {
	return probeZ3BoolSortID()
}

//export ivy_probe_round_trip_bool_sort_id
func ivyProbeRoundTripBoolSortID() uint32 {
	return probeRoundTripBoolSortID()
}

//export ivy_probe_checksum
func ivyProbeChecksum() uint32 {
	return probeChecksumValue()
}
