package bmc

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestCollapseIrredundant(t *testing.T) {
	for _, full := range []bool{false, true} {
		raw := []byte("00 1\n-0 1\n\x00")
		v := vec.Vec_StrAllocArrayCopy(raw, len(raw))
		var status int32
		if full {
			status = Bmc_CollapseIrredundantFull(v, 2, 2)
		} else {
			status = Bmc_CollapseIrredundant(v, 2, 2)
		}
		if status != 1 || string(vec.Vec_StrArray(v)) != "-0 1\n\x00" {
			t.Fatalf("full=%v status=%d sop=%q", full, status, vec.Vec_StrArray(v))
		}
	}
}
func TestCollapseIrredundantUniversalCube(t *testing.T) {
	raw := []byte("00 1\n-- 1\n\x00")
	v := vec.Vec_StrAllocArrayCopy(raw, len(raw))
	if Bmc_CollapseIrredundant(v, 2, 2) != 1 || string(vec.Vec_StrArray(v)) != "-- 1\n\x00" {
		t.Fatalf("sop=%q", vec.Vec_StrArray(v))
	}
}
