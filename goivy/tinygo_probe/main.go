//go:build tinygo

package main

import goivy "github.com/glycerine/ivy/goivy"

var (
	probeConfig    *goivy.Config
	probeModule    *goivy.Module
	probeZ3Context *goivy.Z3Context
	probeBoolSort  goivy.Z3Sort
	probeChecksum  uint32
)

func main() {
	probeChecksum = ivyProbeInit()
}

//export ivy_probe_init
func ivyProbeInit() uint32 {
	cfg := goivy.NewConfig()
	mod := goivy.NewModule()
	ctx := goivy.NewZ3Context()
	boolSort := ctx.BoolSort()

	probeConfig = cfg
	probeModule = mod
	probeZ3Context = ctx
	probeBoolSort = boolSort

	checksum := uint32(boolSort.GetId())
	if cfg.Coverage {
		checksum ^= 0x10000
	}
	if boolSort.Kind() == goivy.SortBool {
		checksum ^= 0x47100000
	}
	checksum += uint32(mod.Relations.Len())
	checksum += uint32(mod.Functions.Len()) << 4
	checksum += uint32(mod.Sig.Sorts.Len()) << 8
	return checksum
}

//export ivy_probe_z3_bool_sort_id
func ivyProbeZ3BoolSortID() uint32 {
	if probeZ3Context == nil {
		probeChecksum = ivyProbeInit()
	}
	return uint32(probeBoolSort.GetId())
}

//export ivy_probe_checksum
func ivyProbeChecksum() uint32 {
	if probeChecksum == 0 {
		probeChecksum = ivyProbeInit()
	}
	return probeChecksum
}
