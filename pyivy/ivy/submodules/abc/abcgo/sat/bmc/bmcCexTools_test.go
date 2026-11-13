package bmc

import (
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestCexUtilityOperations(t *testing.T) {
	v := vec.Vec_WrdStart(3)
	vec.Vec_WrdWriteEntry(v, 0, 10)
	vec.Vec_WrdWriteEntry(v, 1, 12)
	Bmc_CexAndSim(v, 2, 0, 1)
	if vec.Vec_WrdEntry(v, 2) != 8 {
		t.Fatal("AND")
	}
	Bmc_CexOrSim(v, 2, 0, 1)
	if vec.Vec_WrdEntry(v, 2) != 14 || Bmc_CexSim(v, 2, 1) != 1 {
		t.Fatal("OR or bit")
	}
	Bmc_CexCopySim(v, 2, 0)
	if vec.Vec_WrdEntry(v, 2) != 10 || Bmc_CexBitMask(1) != ^uint64(2) {
		t.Fatal("copy or mask")
	}
	c := util.Abc_CexAlloc(1, 2, 2)
	c.PData[0] = 0b01010
	if Bmc_CexBitCount(c, 1) != 2 {
		t.Fatalf("count %d", Bmc_CexBitCount(c, 1))
	}
}
