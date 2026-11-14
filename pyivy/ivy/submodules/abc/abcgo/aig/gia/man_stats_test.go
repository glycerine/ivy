package gia

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/glycerine/abcgo/misc/vec"
)

func TestGiaPrintStatsAndSwitching(t *testing.T) {
	p := Gia_ManStart(8)
	p.pName = "example"
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	Gia_ManAppendCo(p, x)
	previous := LfSwitchingProvider
	defer func() { LfSwitchingProvider = previous }()
	LfSwitchingProvider = func(g *Gia_Man_t, frames, prefix, prob int32) *Vec_Flt_t {
		if frames != 48 || prefix != 16 || prob != 0 {
			t.Fatal("wrong switching options")
		}
		v := vec.Vec_FltStart(int(g.nObjs))
		vec.Vec_FltWriteEntry(v, int(a>>1), 0.25)
		vec.Vec_FltWriteEntry(v, int(b>>1), 0.5)
		return v
	}
	if got := Gia_ManComputeSwitching(p, 48, 16, 0); got != 0.75 {
		t.Fatalf("switching: %v", got)
	}
	path := filepath.Join(t.TempDir(), "stats.json")
	Gia_ManPrintStats(p, &Gps_Par_t{FNoColor: 1, FCut: 1, FSwitch: 1, FMuxXor: 1})
	if p.vLevels != nil {
		t.Fatal("temporary levels retained")
	}
	Gia_ManPrintStats(p, &Gps_Par_t{FNoColor: 1, PDumpFile: path})
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}
}
