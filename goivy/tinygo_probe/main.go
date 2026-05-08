//go:build wasip1

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
	probeChecksum = probeInit()
}

func probeInit() uint32 {
	cfg := goivy.NewConfig()
	mod := goivy.New()
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

func probeZ3BoolSortID() uint32 {
	if probeZ3Context == nil {
		probeChecksum = probeInit()
	}
	return uint32(probeBoolSort.GetId())
}

func probeRoundTripBoolSortID() uint32 {
	if probeZ3Context == nil {
		probeChecksum = probeInit()
	}
	return uint32(probeZ3Context.BoolSort().GetId())
}

func probeChecksumValue() uint32 {
	if probeChecksum == 0 {
		probeChecksum = probeInit()
	}
	return probeChecksum
}
